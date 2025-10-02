package zitadel

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// JWKSProvider предоставляет JWKS (JSON Web Key Set) для валидации JWT токенов
type JWKSProvider struct {
	logger     *zap.Logger
	httpClient *http.Client
	domain     string

	// Cache
	keys       map[string]*rsa.PublicKey
	keysMutex  sync.RWMutex
	lastUpdate time.Time
	cacheTTL   time.Duration

	// JWKS endpoint
	jwksURL string
}

// JWKS представляет JSON Web Key Set
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// JWK представляет JSON Web Key
type JWK struct {
	Kty string `json:"kty"` // Key Type
	Use string `json:"use"` // Public Key Use
	Kid string `json:"kid"` // Key ID
	N   string `json:"n"`   // Modulus
	E   string `json:"e"`   // Exponent
	Alg string `json:"alg"` // Algorithm
}

// NewJWKSProvider создает новый JWKS провайдер
func NewJWKSProvider(domain string, httpClient *http.Client, logger *zap.Logger) (*JWKSProvider, error) {
	// Проверяем, содержит ли domain уже протокол
	var jwksURL string
	if strings.HasPrefix(domain, "http://") || strings.HasPrefix(domain, "https://") {
		jwksURL = fmt.Sprintf("%s/oauth/v2/keys", domain)
	} else {
		jwksURL = fmt.Sprintf("https://%s/oauth/v2/keys", domain)
	}

	provider := &JWKSProvider{
		logger:     logger,
		httpClient: httpClient,
		domain:     domain,
		keys:       make(map[string]*rsa.PublicKey),
		cacheTTL:   15 * time.Minute, // JWKS keys are cached for 15 minutes
		jwksURL:    jwksURL,
	}

	logger.Debug("JWKS provider created", zap.String("jwks_url", jwksURL))
	return provider, nil
}

// GetPublicKey получает публичный ключ по Key ID
func (j *JWKSProvider) GetPublicKey(ctx context.Context, keyID string) (*rsa.PublicKey, error) {
	// Check cache first
	j.keysMutex.RLock()
	if key, exists := j.keys[keyID]; exists && time.Since(j.lastUpdate) < j.cacheTTL {
		j.keysMutex.RUnlock()
		j.logger.Debug("Using cached JWKS key", zap.String("kid", keyID))
		return key, nil
	}
	j.keysMutex.RUnlock()

	// Refresh keys if not in cache or cache expired
	if err := j.RefreshKeys(ctx); err != nil {
		return nil, fmt.Errorf("failed to refresh JWKS keys: %w", err)
	}

	// Try again after refresh
	j.keysMutex.RLock()
	defer j.keysMutex.RUnlock()

	key, exists := j.keys[keyID]
	if !exists {
		return nil, fmt.Errorf("key with ID %s not found in JWKS", keyID)
	}

	return key, nil
}

// RefreshKeys обновляет ключи из JWKS endpoint
func (j *JWKSProvider) RefreshKeys(ctx context.Context) error {
	j.logger.Debug("Refreshing JWKS keys", zap.String("jwks_url", j.jwksURL))

	req, err := http.NewRequestWithContext(ctx, "GET", j.jwksURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create JWKS request: %w", err)
	}

	resp, err := j.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("JWKS request failed with status: %d", resp.StatusCode)
	}

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return fmt.Errorf("failed to decode JWKS response: %w", err)
	}

	// Convert JWKs to RSA public keys
	newKeys := make(map[string]*rsa.PublicKey)
	for _, jwk := range jwks.Keys {
		if jwk.Kty != "RSA" {
			j.logger.Debug("Skipping non-RSA key", zap.String("kid", jwk.Kid), zap.String("kty", jwk.Kty))
			continue
		}

		publicKey, err := j.jwkToRSAPublicKey(jwk)
		if err != nil {
			j.logger.Warn("Failed to convert JWK to RSA public key",
				zap.String("kid", jwk.Kid), zap.Error(err))
			continue
		}

		newKeys[jwk.Kid] = publicKey
		j.logger.Debug("Loaded JWKS key", zap.String("kid", jwk.Kid), zap.String("alg", jwk.Alg))
	}

	if len(newKeys) == 0 {
		return fmt.Errorf("no valid RSA keys found in JWKS")
	}

	// Update cache
	j.keysMutex.Lock()
	j.keys = newKeys
	j.lastUpdate = time.Now()
	j.keysMutex.Unlock()

	j.logger.Info("JWKS keys refreshed successfully", zap.Int("key_count", len(newKeys)))
	return nil
}

// jwkToRSAPublicKey конвертирует JWK в RSA публичный ключ
func (j *JWKSProvider) jwkToRSAPublicKey(jwk JWK) (*rsa.PublicKey, error) {
	// Decode base64url encoded modulus and exponent
	nBytes, err := jwt.NewParser().DecodeSegment(jwk.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode modulus: %w", err)
	}

	eBytes, err := jwt.NewParser().DecodeSegment(jwk.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exponent: %w", err)
	}

	// Convert bytes to big integers
	n := new(big.Int).SetBytes(nBytes)

	// Convert exponent bytes to int
	var e int
	for _, b := range eBytes {
		e = e*256 + int(b)
	}

	// Create RSA public key
	publicKey := &rsa.PublicKey{
		N: n,
		E: e,
	}

	return publicKey, nil
}

// GetAllKeys возвращает все кэшированные ключи
func (j *JWKSProvider) GetAllKeys() map[string]*rsa.PublicKey {
	j.keysMutex.RLock()
	defer j.keysMutex.RUnlock()

	keys := make(map[string]*rsa.PublicKey)
	for kid, key := range j.keys {
		keys[kid] = key
	}

	return keys
}

// GetKeyIDs возвращает список всех Key IDs
func (j *JWKSProvider) GetKeyIDs() []string {
	j.keysMutex.RLock()
	defer j.keysMutex.RUnlock()

	var keyIDs []string
	for kid := range j.keys {
		keyIDs = append(keyIDs, kid)
	}

	return keyIDs
}

// IsKeyExpired проверяет, истек ли кэш ключей
func (j *JWKSProvider) IsKeyExpired() bool {
	j.keysMutex.RLock()
	defer j.keysMutex.RUnlock()

	return time.Since(j.lastUpdate) >= j.cacheTTL
}

// SetCacheTTL устанавливает время жизни кэша
func (j *JWKSProvider) SetCacheTTL(ttl time.Duration) {
	j.keysMutex.Lock()
	defer j.keysMutex.Unlock()

	j.cacheTTL = ttl
	j.logger.Debug("JWKS cache TTL updated", zap.Duration("ttl", ttl))
}

// GetCacheInfo возвращает информацию о кэше
func (j *JWKSProvider) GetCacheInfo() map[string]interface{} {
	j.keysMutex.RLock()
	defer j.keysMutex.RUnlock()

	return map[string]interface{}{
		"key_count":   len(j.keys),
		"last_update": j.lastUpdate,
		"cache_ttl":   j.cacheTTL,
		"expired":     time.Since(j.lastUpdate) >= j.cacheTTL,
		"jwks_url":    j.jwksURL,
	}
}
