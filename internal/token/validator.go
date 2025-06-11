package token

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// HTTPDoer is an interface for making HTTP requests
// Used for easier mocking in tests
type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}

// Config holds the token validator configuration
type Config struct {
	// REST validation settings
	RESTEnabled     bool
	RESTEndpoint    string
	RESTTimeout     time.Duration
	RESTCacheTTL    time.Duration
	RESTMaxRetries  int
	RESTRetryDelay  time.Duration

	// JWT validation settings
	JWTEnabled      bool
	JWTSigningKey   []byte
	JWTIssuer       string
	JWTMaxAge       time.Duration

	// Debug mode enables detailed logging
	Debug bool
}

// Validator handles token validation
type Validator struct {
	config     Config
	httpClient HTTPDoer
	cache      *tokenCache
	logger     *log.Logger
}

// tokenCache stores validated tokens with TTL
type tokenCache struct {
	tokens map[string]time.Time
	mu     sync.RWMutex
}

// NewValidator creates a new token validator
func NewValidator(config Config) *Validator {
	if config.RESTTimeout == 0 {
		config.RESTTimeout = 5 * time.Second
	}
	if config.RESTCacheTTL == 0 {
		config.RESTCacheTTL = 1 * time.Minute
	}
	if config.RESTRetryDelay == 0 {
		config.RESTRetryDelay = 100 * time.Millisecond
	}
	if config.JWTMaxAge == 0 {
		config.JWTMaxAge = 24 * time.Hour
	}

	return &Validator{
		config: config,
		httpClient: &http.Client{
			Timeout: config.RESTTimeout,
		},
		cache: &tokenCache{
			tokens: make(map[string]time.Time),
		},
		logger: log.New(log.Writer(), "[token-validator] ", log.LstdFlags),
	}
}

// Validate checks if a token is valid using the configured method
func (v *Validator) Validate(token string) bool {
	if token == "" {
		v.debugLog("empty token provided")
		return false
	}

	// Check cache first
	if v.checkCache(token) {
		v.debugLog("token found in cache: %s", token)
		return true
	}

	// Validate using configured method
	var valid bool
	var method string
	if v.config.JWTEnabled {
		method = "JWT"
		valid = v.validateJWT(token)
	} else if v.config.RESTEnabled {
		method = "REST"
		valid = v.validateREST(token)
	} else {
		method = "none"
		valid = true
	}

	v.debugLog("token validation result: %s (method: %s)", token, method)

	// Cache valid tokens
	if valid {
		v.cacheToken(token)
	}

	return valid
}

// checkCache checks if a token is in the cache and not expired
func (v *Validator) checkCache(token string) bool {
	v.cache.mu.RLock()
	defer v.cache.mu.RUnlock()

	if expiry, ok := v.cache.tokens[token]; ok {
		valid := time.Now().Before(expiry)
		v.debugLog("cache check for token %s: %v (expires: %v)", token, valid, expiry)
		return valid
	}
	return false
}

// cacheToken adds a token to the cache with TTL
func (v *Validator) cacheToken(token string) {
	v.cache.mu.Lock()
	defer v.cache.mu.Unlock()

	expiry := time.Now().Add(v.config.RESTCacheTTL)
	v.cache.tokens[token] = expiry
	v.debugLog("cached token %s (expires: %v)", token, expiry)
}

// validateREST validates a token by making a request to the REST API
func (v *Validator) validateREST(token string) bool {
	url := fmt.Sprintf("%s?token=%s", v.config.RESTEndpoint, token)
	
	for i := 0; i < v.config.RESTMaxRetries; i++ {
		v.debugLog("REST validation attempt %d/%d for token %s", i+1, v.config.RESTMaxRetries, token)

		req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
		if err != nil {
			v.debugLog("failed to create request: %v", err)
			continue
		}

		resp, err := v.httpClient.Do(req)
		if err != nil {
			v.debugLog("request failed: %v", err)
			time.Sleep(v.config.RESTRetryDelay)
			continue
		}
		resp.Body.Close()

		switch resp.StatusCode {
		case http.StatusOK:
			v.debugLog("REST validation successful for token %s", token)
			return true
		case http.StatusForbidden, http.StatusUnauthorized:
			v.debugLog("REST validation failed for token %s (status: %d)", token, resp.StatusCode)
			return false
		default:
			v.debugLog("unexpected status code %d for token %s", resp.StatusCode, token)
		}
	}

	v.debugLog("REST validation failed after %d attempts for token %s", v.config.RESTMaxRetries, token)
	return false
}

// validateJWT validates a JWT token
func (v *Validator) validateJWT(tokenString string) bool {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return v.config.JWTSigningKey, nil
	})

	if err != nil {
		v.debugLog("JWT parsing failed: %v", err)
		return false
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Validate issuer
		if v.config.JWTIssuer != "" {
			if iss, ok := claims["iss"].(string); !ok || iss != v.config.JWTIssuer {
				v.debugLog("JWT issuer validation failed: expected %s, got %v", v.config.JWTIssuer, iss)
				return false
			}
		}

		// Validate expiration
		if exp, ok := claims["exp"].(float64); ok {
			expTime := time.Unix(int64(exp), 0)
			if expTime.Before(time.Now()) {
				v.debugLog("JWT token expired at %v", expTime)
				return false
			}
		}

		// Validate not before
		if nbf, ok := claims["nbf"].(float64); ok {
			nbfTime := time.Unix(int64(nbf), 0)
			if nbfTime.After(time.Now()) {
				v.debugLog("JWT token not valid before %v", nbfTime)
				return false
			}
		}

		// Validate max age if specified
		if v.config.JWTMaxAge > 0 {
			if iat, ok := claims["iat"].(float64); ok {
				issuedAt := time.Unix(int64(iat), 0)
				if time.Since(issuedAt) > v.config.JWTMaxAge {
					v.debugLog("JWT token exceeds max age: %v", time.Since(issuedAt))
					return false
				}
			}
		}

		v.debugLog("JWT validation successful for token %s", tokenString)
		return true
	}

	v.debugLog("JWT validation failed: invalid claims or token")
	return false
}

// Cleanup removes expired tokens from the cache
func (v *Validator) Cleanup() {
	v.cache.mu.Lock()
	defer v.cache.mu.Unlock()

	now := time.Now()
	count := 0
	for token, expiry := range v.cache.tokens {
		if now.After(expiry) {
			delete(v.cache.tokens, token)
			count++
		}
	}
	v.debugLog("cleaned up %d expired tokens", count)
}

// debugLog logs a message if debug mode is enabled
func (v *Validator) debugLog(format string, args ...interface{}) {
	if v.config.Debug {
		v.logger.Printf(format, args...)
	}
} 