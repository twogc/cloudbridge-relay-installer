package token

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/twogc/cloudbridge-relay-installer/internal/django"
)

// Config represents the token validator configuration
type Config struct {
	DjangoClient *django.Client
	JWTSecret    string
	CacheTTL     time.Duration
}

// TokenInfo represents cached token information
type TokenInfo struct {
	ClientID string
	Expiry   time.Time
}

// Validator implements token validation
type Validator struct {
	config     Config
	cache      map[string]TokenInfo
	mu         sync.RWMutex
	jwtKey     []byte
}

// NewValidator creates a new token validator
func NewValidator(config Config) *Validator {
	// Set default cache TTL if not specified
	if config.CacheTTL == 0 {
		config.CacheTTL = 5 * time.Minute
	}
	
	return &Validator{
		config: config,
		cache:  make(map[string]TokenInfo),
		jwtKey: []byte(config.JWTSecret),
	}
}

// ValidateToken validates a token and returns the client ID
func (v *Validator) ValidateToken(token string) (string, error) {
	// Reject empty tokens
	if token == "" {
		return "", fmt.Errorf("empty token")
	}

	// Check cache first
	v.mu.RLock()
	info, ok := v.cache[token]
	v.mu.RUnlock()

	if ok && time.Now().Before(info.Expiry) {
		return info.ClientID, nil
	}

	// Validate token
	var clientID string
	var err error

	if v.config.DjangoClient != nil {
		clientID, err = v.config.DjangoClient.ValidateToken(context.Background(), token)
		if err != nil {
			return "", fmt.Errorf("Django validation error: %v", err)
		}
	} else if v.config.JWTSecret != "" {
		clientID, err = v.validateJWT(token)
		if err != nil {
			return "", fmt.Errorf("JWT validation error: %v", err)
		}
	} else {
		// No validation configured - accept any token
		return "default-client", nil
	}

	// Cache valid token
	v.mu.Lock()
	v.cache[token] = TokenInfo{
		ClientID: clientID,
		Expiry:   time.Now().Add(v.config.CacheTTL),
	}
	v.mu.Unlock()

	return clientID, nil
}

// validateJWT validates a token using JWT and returns the client ID
func (v *Validator) validateJWT(token string) (string, error) {
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return v.jwtKey, nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to parse JWT: %v", err)
	}

	// Check expiration
	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return "", fmt.Errorf("token expired")
		}
	}

	// Get client ID
	clientID, ok := claims["client_id"].(string)
	if !ok {
		return "", fmt.Errorf("missing client_id in token")
	}

	return clientID, nil
}

// Cleanup removes expired tokens from the cache
func (v *Validator) Cleanup() {
	v.mu.Lock()
	defer v.mu.Unlock()

	now := time.Now()
	for token, info := range v.cache {
		if now.After(info.Expiry) {
			delete(v.cache, token)
		}
	}
} 