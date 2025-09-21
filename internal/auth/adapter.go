package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/config"
	"github.com/twogc/cloudbridge-relay-installer/internal/keycloak"
	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
)

// KeycloakClientInterface defines the interface for Keycloak client
type KeycloakClientInterface interface {
	ValidateToken(ctx context.Context, token string) (*keycloak.TokenInfo, error)
	CheckPermission(tokenInfo *keycloak.TokenInfo, resource, action string) bool
}

// Provider represents authentication provider interface
type Provider interface {
	ValidateToken(ctx context.Context, token string) (*TokenInfo, error)
	GetUserInfo(ctx context.Context, token string) (*UserInfo, error)
	CheckPermission(ctx context.Context, token string, resource, action string) (bool, error)
}

// TokenInfo represents unified token information
type TokenInfo struct {
	Subject   string                 `json:"sub"`
	Issuer    string                 `json:"iss"`
	Audience  []string               `json:"aud"`
	ExpiresAt time.Time              `json:"exp"`
	IssuedAt  time.Time              `json:"iat"`
	NotBefore time.Time              `json:"nbf"`
	Username  string                 `json:"username"`
	Email     string                 `json:"email"`
	FullName  string                 `json:"full_name"`
	Groups    []string               `json:"groups"`
	Roles     []string               `json:"roles"`
	Claims    map[string]interface{} `json:"claims"`
	Provider  string                 `json:"provider"`
}

// UserInfo represents unified user information
type UserInfo struct {
	ID       string                 `json:"id"`
	Username string                 `json:"username"`
	Email    string                 `json:"email"`
	FullName string                 `json:"full_name"`
	Groups   []string               `json:"groups"`
	Roles    []string               `json:"roles"`
	Active   bool                   `json:"active"`
	Claims   map[string]interface{} `json:"claims"`
	Provider string                 `json:"provider"`
}

// Adapter represents authentication adapter
type Adapter struct {
	keycloakProvider KeycloakClientInterface
	config           *config.Config
	logger           *logger.Logger
	provider         Provider
}

// NewAdapter creates a new authentication adapter
func NewAdapter(cfg *config.Config, logger *logger.Logger) (*Adapter, error) {
	adapter := &Adapter{
		config: cfg,
		logger: logger,
	}

	// Проверяем, включена ли аутентификация
	if !cfg.ZeroTrust.Authentication.Enabled {
		// Аутентификация отключена - создаем no-op provider
		adapter.provider = &NoOpProvider{}
		logger.Info("Authentication disabled - using no-op provider")
	} else if cfg.Keycloak.Enabled {
		// Initialize Keycloak provider
		keycloakConfig := &keycloak.Config{
			ServerURL:    cfg.Keycloak.ServerURL,
			Realm:        cfg.Keycloak.Realm,
			ClientID:     cfg.Keycloak.ClientID,
			ClientSecret: cfg.Keycloak.ClientSecret,
			PublicKey:    cfg.Keycloak.PublicKey,
			TokenValidation: struct {
				Issuer    string        `json:"issuer"`
				Audience  string        `json:"audience"`
				ClockSkew time.Duration `json:"clock_skew"`
			}{
				Issuer:    cfg.Keycloak.TokenValidation.Issuer,
				Audience:  cfg.Keycloak.TokenValidation.Audience,
				ClockSkew: cfg.Keycloak.TokenValidation.ClockSkew,
			},
			HTTPClient: nil, // Will be set by keycloak client
			Logger:     logger,
		}

		keycloakClient, err := keycloak.NewClient(keycloakConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create Keycloak client: %v", err)
		}

		adapter.keycloakProvider = keycloakClient
		adapter.provider = &KeycloakProvider{client: keycloakClient}
	} else {
		return nil, fmt.Errorf("authentication enabled but no provider configured - Keycloak must be enabled")
	}

	return adapter, nil
}

// ValidateToken validates token using the configured provider
func (a *Adapter) ValidateToken(ctx context.Context, token string) (*TokenInfo, error) {
	return a.provider.ValidateToken(ctx, token)
}

// GetUserInfo gets user information using the configured provider
func (a *Adapter) GetUserInfo(ctx context.Context, token string) (*UserInfo, error) {
	return a.provider.GetUserInfo(ctx, token)
}

// CheckPermission checks permission using the configured provider
func (a *Adapter) CheckPermission(ctx context.Context, token string, resource, action string) (bool, error) {
	return a.provider.CheckPermission(ctx, token, resource, action)
}

// GetProvider returns the current authentication provider name
func (a *Adapter) GetProvider() string {
	if a.config.Keycloak.Enabled {
		return "keycloak"
	}
	return "none"
}

// KeycloakProvider implements Provider interface for Keycloak
type KeycloakProvider struct {
	client KeycloakClientInterface
}

// ValidateToken validates token using Keycloak
func (p *KeycloakProvider) ValidateToken(ctx context.Context, token string) (*TokenInfo, error) {
	tokenInfo, err := p.client.ValidateToken(ctx, token)
	if err != nil {
		return nil, err
	}

	return &TokenInfo{
		Subject:   tokenInfo.Subject,
		Issuer:    tokenInfo.Issuer,
		Audience:  tokenInfo.Audience,
		ExpiresAt: tokenInfo.ExpiresAt,
		IssuedAt:  tokenInfo.IssuedAt,
		NotBefore: tokenInfo.NotBefore,
		Username:  tokenInfo.Username,
		Email:     tokenInfo.Email,
		FullName:  tokenInfo.FullName,
		Groups:    tokenInfo.Groups,
		Roles:     tokenInfo.Roles,
		Claims:    tokenInfo.Claims,
		Provider:  "keycloak",
	}, nil
}

// GetUserInfo gets user information from Keycloak token
func (p *KeycloakProvider) GetUserInfo(ctx context.Context, token string) (*UserInfo, error) {
	tokenInfo, err := p.ValidateToken(ctx, token)
	if err != nil {
		return nil, err
	}

	return &UserInfo{
		ID:       tokenInfo.Subject,
		Username: tokenInfo.Username,
		Email:    tokenInfo.Email,
		FullName: tokenInfo.FullName,
		Groups:   tokenInfo.Groups,
		Roles:    tokenInfo.Roles,
		Active:   true, // Keycloak tokens are only issued for active users
		Claims:   tokenInfo.Claims,
		Provider: "keycloak",
	}, nil
}

// CheckPermission checks permission using Keycloak roles and groups
func (p *KeycloakProvider) CheckPermission(ctx context.Context, token string, resource, action string) (bool, error) {
	tokenInfo, err := p.ValidateToken(ctx, token)
	if err != nil {
		return false, err
	}

	kt := &keycloak.TokenInfo{
		Subject:   tokenInfo.Subject,
		Issuer:    tokenInfo.Issuer,
		Audience:  tokenInfo.Audience,
		ExpiresAt: tokenInfo.ExpiresAt,
		IssuedAt:  tokenInfo.IssuedAt,
		NotBefore: tokenInfo.NotBefore,
		Username:  tokenInfo.Username,
		Email:     tokenInfo.Email,
		FullName:  tokenInfo.FullName,
		Groups:    tokenInfo.Groups,
		Roles:     tokenInfo.Roles,
		Claims:    tokenInfo.Claims,
	}
	return p.client.CheckPermission(kt, resource, action), nil
}

// NoOpProvider представляет провайдер аутентификации без проверки
type NoOpProvider struct{}

// ValidateToken всегда возвращает успешную валидацию
func (p *NoOpProvider) ValidateToken(ctx context.Context, token string) (*TokenInfo, error) {
	return &TokenInfo{
		Subject:   "anonymous",
		Issuer:    "no-op",
		Audience:  []string{"cloudbridge-relay"},
		ExpiresAt: time.Now().Add(24 * time.Hour),
		IssuedAt:  time.Now(),
		NotBefore: time.Now(),
		Username:  "anonymous",
		Email:     "anonymous@cloudbridge.local",
		FullName:  "Anonymous User",
		Groups:    []string{"users"},
		Roles:     []string{"user"},
		Claims:    make(map[string]interface{}),
		Provider:  "no-op",
	}, nil
}

// GetUserInfo возвращает информацию о пользователе
func (p *NoOpProvider) GetUserInfo(ctx context.Context, token string) (*UserInfo, error) {
	return &UserInfo{
		ID:       "anonymous",
		Username: "anonymous",
		Email:    "anonymous@cloudbridge.local",
		FullName: "Anonymous User",
		Groups:   []string{"users"},
		Roles:    []string{"user"},
		Active:   true,
		Claims:   make(map[string]interface{}),
		Provider: "no-op",
	}, nil
}

// CheckPermission всегда возвращает true
func (p *NoOpProvider) CheckPermission(ctx context.Context, token string, resource, action string) (bool, error) {
	return true, nil
}
