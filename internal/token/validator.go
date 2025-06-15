package token

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/maxlanies/cloudbridge-relay-installer/internal/django"
)

// Config represents the token validator configuration
type Config struct {
	DjangoClient *django.Client
	JWTSecret    string
	CacheTTL     time.Duration
}

// Validator implements token validation
type Validator struct {
	config     Config
	cache      map[string]time.Time
	mu         sync.RWMutex
	jwtKey     []byte
}

// NewValidator creates a new token validator
func NewValidator(config Config) *Validator {
	return &Validator{
		config: config,
		cache:  make(map[string]time.Time),
		jwtKey: []byte(config.JWTSecret),
	}
}

// Validate validates a token
func (v *Validator) Validate(token string) bool {
	// Check cache first
	v.mu.RLock()
	expiry, ok := v.cache[token]
	v.mu.RUnlock()

	if ok && time.Now().Before(expiry) {
		return true
	}

	// Validate token
	var valid bool
	var err error

	if v.config.DjangoClient != nil {
		valid, err = v.config.DjangoClient.ValidateToken(context.Background(), token)
		if err != nil {
			fmt.Printf("Django validation error: %v\n", err)
			valid = false
		}
	} else if v.config.JWTSecret != "" {
		valid = v.validateJWT(token)
	} else {
		// No validation configured
		valid = true
	}

	if valid {
		// Cache valid token
		v.mu.Lock()
		v.cache[token] = time.Now().Add(v.config.CacheTTL)
		v.mu.Unlock()
	}

	return valid
}

// validateJWT validates a token using JWT
func (v *Validator) validateJWT(token string) bool {
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return v.jwtKey, nil
	})

	if err != nil {
		return false
	}

	// Check expiration
	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return false
		}
	}

	return true
}

// Cleanup removes expired tokens from the cache
func (v *Validator) Cleanup() {
	v.mu.Lock()
	defer v.mu.Unlock()

	now := time.Now()
	for token, expiry := range v.cache {
		if now.After(expiry) {
			delete(v.cache, token)
		}
	}
} 