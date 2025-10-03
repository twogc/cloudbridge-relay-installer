package auth

import (
	"context"
	"crypto/rsa"
	"fmt"
	"net/http"
	"time"

	"github.com/zitadel/oidc/v3/pkg/client"
	"go.uber.org/zap"
)

// ZitadelClient предоставляет клиент для работы с Zitadel API
type ZitadelClient struct {
	domain          string
	httpClient      *http.Client
	discoveryConfig interface{}
	logger          *zap.Logger
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

	return &ZitadelClient{
		domain:          config.Domain,
		httpClient:      config.HTTPClient,
		discoveryConfig: discoveryConfig,
		logger:          config.Logger,
	}, nil
}

// GetJWKS возвращает JSON Web Key Set для валидации JWT токенов
func (c *ZitadelClient) GetJWKS(ctx context.Context) (interface{}, error) {
	// TODO: Реализовать получение JWKS из discovery config
	return nil, fmt.Errorf("JWKS not implemented yet")
}

// GetPublicKey возвращает публичный ключ для валидации JWT
func (c *ZitadelClient) GetPublicKey(ctx context.Context, keyID string) (*rsa.PublicKey, error) {
	// TODO: Реализовать получение публичного ключа из JWKS
	return nil, fmt.Errorf("public key retrieval not implemented yet")
}

// ValidateTokenWithIntrospection валидирует токен через introspection endpoint
func (c *ZitadelClient) ValidateTokenWithIntrospection(ctx context.Context, token string) (*IntrospectionResponse, error) {
	// TODO: Реализовать introspection endpoint
	return nil, fmt.Errorf("introspection not implemented yet")
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
func (c *ZitadelClient) GetDiscoveryConfig() interface{} {
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
