package auth

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/zitadel/oidc/v3/pkg/client"
	"github.com/zitadel/oidc/v3/pkg/oidc"
	"go.uber.org/zap"
)

// ZitadelClient предоставляет клиент для работы с Zitadel API
type ZitadelClient struct {
	domain          string
	httpClient      *http.Client
	discoveryConfig *oidc.DiscoveryConfiguration
	logger          *zap.Logger

	// JWKS caching
	jwksCache    *JWKSCache
	jwksCacheTTL time.Duration
}

// JWKSCache кеш для JWKS с thread-safe доступом
type JWKSCache struct {
	mu        sync.RWMutex
	jwks      *oidc.KeySet
	expiresAt time.Time
}

// ZitadelClientConfig конфигурация для ZitadelClient
type ZitadelClientConfig struct {
	Domain     string
	HTTPClient *http.Client
	Logger     *zap.Logger
}

// NewZitadelClient создает новый клиент для работы с Zitadel
func NewZitadelClient(config ZitadelClientConfig) (*ZitadelClient, error) {
	if config.HTTPClient == nil {
		config.HTTPClient = &http.Client{Timeout: 30 * time.Second}
	}

	// Получаем конфигурацию discovery
	discoveryConfig, err := client.Discover(context.Background(), config.Domain, config.HTTPClient)
	if err != nil {
		return nil, fmt.Errorf("failed to discover Zitadel configuration: %w", err)
	}

	// Проверяем, что получили правильный тип
	oidcConfig, ok := discoveryConfig.(*oidc.DiscoveryConfiguration)
	if !ok {
		return nil, fmt.Errorf("unexpected discovery config type")
	}

	return &ZitadelClient{
		domain:          config.Domain,
		httpClient:      config.HTTPClient,
		discoveryConfig: oidcConfig,
		logger:          config.Logger,
		jwksCache: &JWKSCache{
			expiresAt: time.Time{}, // Инициализируем с нулевым временем
		},
		jwksCacheTTL: 1 * time.Hour, // JWKS кешируем на 1 час
	}, nil
}

// GetJWKS возвращает JSON Web Key Set для валидации JWT токенов
func (c *ZitadelClient) GetJWKS(ctx context.Context) (*oidc.KeySet, error) {
	// Проверяем кеш
	c.jwksCache.mu.RLock()
	if c.jwksCache.jwks != nil && time.Now().Before(c.jwksCache.expiresAt) {
		jwks := c.jwksCache.jwks
		c.jwksCache.mu.RUnlock()
		c.logger.Debug("JWKS returned from cache",
			zap.Time("expires_at", c.jwksCache.expiresAt))
		return jwks, nil
	}
	c.jwksCache.mu.RUnlock()

	// Получаем JWKS URI из discovery config
	jwksURI := c.discoveryConfig.JwksURI
	if jwksURI == "" {
		return nil, fmt.Errorf("jwks_uri not found in discovery configuration")
	}

	c.logger.Info("Fetching JWKS from Zitadel", zap.String("jwks_uri", jwksURI))

	// Создаем HTTP запрос
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, jwksURI, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWKS request: %w", err)
	}

	// Выполняем запрос
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("JWKS endpoint returned status %d", resp.StatusCode)
	}

	// Парсим JWKS
	var jwks oidc.KeySet
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, fmt.Errorf("failed to decode JWKS: %w", err)
	}

	// Проверяем что получили хотя бы один ключ
	if len(jwks.Keys) == 0 {
		return nil, fmt.Errorf("JWKS contains no keys")
	}

	c.logger.Info("JWKS fetched successfully",
		zap.Int("keys_count", len(jwks.Keys)))

	// Обновляем кеш
	c.jwksCache.mu.Lock()
	c.jwksCache.jwks = &jwks
	c.jwksCache.expiresAt = time.Now().Add(c.jwksCacheTTL)
	c.jwksCache.mu.Unlock()

	return &jwks, nil
}

// GetPublicKey возвращает публичный ключ для валидации JWT
func (c *ZitadelClient) GetPublicKey(ctx context.Context, keyID string) (*rsa.PublicKey, error) {
	// Получаем JWKS
	jwks, err := c.GetJWKS(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get JWKS: %w", err)
	}

	// Ищем ключ по keyID
	for _, key := range jwks.Keys {
		if key.KeyID == keyID {
			// Проверяем что это RSA ключ
			if key.Algorithm != "RS256" {
				c.logger.Warn("Unsupported key algorithm",
					zap.String("key_id", keyID),
					zap.String("algorithm", key.Algorithm))
				continue
			}

			// Конвертируем в RSA public key
			pubKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(key.PublicKeyPEM()))
			if err != nil {
				c.logger.Error("Failed to parse RSA public key",
					zap.String("key_id", keyID),
					zap.Error(err))
				continue
			}

			c.logger.Debug("Public key found",
				zap.String("key_id", keyID),
				zap.String("algorithm", key.Algorithm))

			return pubKey, nil
		}
	}

	return nil, fmt.Errorf("public key with keyID %s not found in JWKS", keyID)
}

// GetPublicKeyByIndex возвращает публичный ключ по индексу (для случаев без keyID)
func (c *ZitadelClient) GetPublicKeyByIndex(ctx context.Context, index int) (*rsa.PublicKey, error) {
	// Получаем JWKS
	jwks, err := c.GetJWKS(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get JWKS: %w", err)
	}

	if index >= len(jwks.Keys) {
		return nil, fmt.Errorf("key index %d out of range (have %d keys)", index, len(jwks.Keys))
	}

	key := jwks.Keys[index]

	// Проверяем что это RSA ключ
	if key.Algorithm != "RS256" {
		return nil, fmt.Errorf("unsupported key algorithm: %s", key.Algorithm)
	}

	// Конвертируем в RSA public key
	pubKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(key.PublicKeyPEM()))
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSA public key: %w", err)
	}

	c.logger.Debug("Public key found by index",
		zap.Int("index", index),
		zap.String("key_id", key.KeyID),
		zap.String("algorithm", key.Algorithm))

	return pubKey, nil
}

// ValidateTokenWithIntrospection валидирует токен через introspection endpoint
func (c *ZitadelClient) ValidateTokenWithIntrospection(ctx context.Context, token string) (*IntrospectionResponse, error) {
	// Получаем introspection URI из discovery config
	introspectionURI := c.discoveryConfig.IntrospectionEndpoint
	if introspectionURI == "" {
		return nil, fmt.Errorf("introspection_endpoint not found in discovery configuration")
	}

	c.logger.Debug("Validating token via introspection", zap.String("endpoint", introspectionURI))

	// Создаем форму для introspection
	form := map[string][]string{
		"token": {token},
	}

	// Создаем HTTP запрос
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, introspectionURI, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create introspection request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.PostForm = form

	// Выполняем запрос
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to introspect token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("introspection endpoint returned status %d", resp.StatusCode)
	}

	// Парсим ответ
	var introspectionResp IntrospectionResponse
	if err := json.NewDecoder(resp.Body).Decode(&introspectionResp); err != nil {
		return nil, fmt.Errorf("failed to decode introspection response: %w", err)
	}

	c.logger.Debug("Token introspection completed",
		zap.Bool("active", introspectionResp.Active),
		zap.String("sub", introspectionResp.Sub))

	return &introspectionResp, nil
}

// RefreshJWKSCache принудительно обновляет кеш JWKS
func (c *ZitadelClient) RefreshJWKSCache(ctx context.Context) error {
	c.jwksCache.mu.Lock()
	// Сбрасываем время истечения чтобы вызвать обновление
	c.jwksCache.expiresAt = time.Time{}
	c.jwksCache.mu.Unlock()

	// Получаем JWKS - это вызовет обновление кеша
	_, err := c.GetJWKS(ctx)
	return err
}

// IntrospectionResponse представляет ответ от introspection endpoint
type IntrospectionResponse struct {
	Active    bool                   `json:"active"`
	Scope     string                 `json:"scope,omitempty"`
	ClientID  string                 `json:"client_id,omitempty"`
	Username  string                 `json:"username,omitempty"`
	TokenType string                 `json:"token_type,omitempty"`
	Exp       *int64                 `json:"exp,omitempty"`
	Iat       *int64                 `json:"iat,omitempty"`
	Nbf       *int64                 `json:"nbf,omitempty"`
	Sub       string                 `json:"sub,omitempty"`
	Aud       []string               `json:"aud,omitempty"`
	Iss       string                 `json:"iss,omitempty"`
	Jti       string                 `json:"jti,omitempty"`
	Claims    map[string]interface{} `json:"claims,omitempty"`
}

// IsActive проверяет, активен ли токен
func (r *IntrospectionResponse) IsActive() bool {
	return r.Active
}

// GetClaims возвращает claims токена
func (r *IntrospectionResponse) GetClaims() map[string]interface{} {
	return r.Claims
}

// GetExpirationTime возвращает время истечения токена
func (r *IntrospectionResponse) GetExpirationTime() *time.Time {
	if r.Exp == nil {
		return nil
	}
	exp := time.Unix(*r.Exp, 0)
	return &exp
}

// GetIssuedAt возвращает время выдачи токена
func (r *IntrospectionResponse) GetIssuedAt() *time.Time {
	if r.Iat == nil {
		return nil
	}
	iat := time.Unix(*r.Iat, 0)
	return &iat
}

// GetNotBefore возвращает время "не раньше"
func (r *IntrospectionResponse) GetNotBefore() *time.Time {
	if r.Nbf == nil {
		return nil
	}
	nbf := time.Unix(*r.Nbf, 0)
	return &nbf
}

// GetDomain возвращает домен Zitadel
func (c *ZitadelClient) GetDomain() string {
	return c.domain
}

// GetDiscoveryConfig возвращает конфигурацию discovery
func (c *ZitadelClient) GetDiscoveryConfig() *oidc.DiscoveryConfiguration {
	return c.discoveryConfig
}

// Close закрывает клиент
func (c *ZitadelClient) Close() error {
	if c.httpClient != nil {
		// Закрываем все соединения
		c.httpClient.CloseIdleConnections()
	}
	return nil
}
