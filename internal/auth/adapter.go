package auth

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/config"
	"github.com/twogc/cloudbridge-relay-installer/internal/django"
	"github.com/twogc/cloudbridge-relay-installer/internal/keycloak"
	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
)

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
	keycloakProvider *keycloak.Client
	djangoProvider   *django.Client
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

	if cfg.Keycloak.Enabled {
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
		// Initialize Django provider
		// Get JWT secret from environment variable instead of config
		jwtSecret := os.Getenv("JWT_SECRET")
		if jwtSecret == "" {
			jwtSecret = cfg.Django.JWTSecret // Fallback to config if env var not set
		}
		
		djangoConfig := django.Config{
			APIEndpoint: cfg.Django.APIEndpoint,
			JWTSecret:   jwtSecret,
			Timeout:     cfg.Django.RESTTimeout,
		}

		djangoClient := django.NewClient(djangoConfig)
		adapter.djangoProvider = djangoClient
		adapter.provider = &DjangoProvider{client: djangoClient}
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
	return "django"
}

// KeycloakProvider implements Provider interface for Keycloak
type KeycloakProvider struct {
	client *keycloak.Client
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

// DjangoProvider implements Provider interface for Django
type DjangoProvider struct {
	client *django.Client
}

// ValidateToken validates token using Django
func (p *DjangoProvider) ValidateToken(ctx context.Context, token string) (*TokenInfo, error) {
	// This would need to be implemented based on your Django token validation
	// For now, returning a placeholder implementation
	return nil, fmt.Errorf("Django token validation not implemented")
}

// GetUserInfo gets user information from Django
func (p *DjangoProvider) GetUserInfo(ctx context.Context, token string) (*UserInfo, error) {
	// This would need to be implemented based on your Django API
	// For now, returning a placeholder implementation
	return nil, fmt.Errorf("Django user info not implemented")
}

// CheckPermission checks permission using Django
func (p *DjangoProvider) CheckPermission(ctx context.Context, token string, resource, action string) (bool, error) {
	// This would need to be implemented based on your Django permission system
	// For now, returning a placeholder implementation
	return false, fmt.Errorf("Django permission check not implemented")
}
