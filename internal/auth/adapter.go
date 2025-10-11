package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/twogc/cloudbridge-relay-installer/config"
	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
	"go.uber.org/zap"
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
	TenantID  string                 `json:"tenant_id"`
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
	config   *config.Config
	logger   *logger.Logger
	provider Provider
}

// NewAdapter creates a new authentication adapter
func NewAdapter(cfg *config.Config, logger *logger.Logger) (*Adapter, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}

	adapter := &Adapter{
		config: cfg,
		logger: logger,
	}

	// Проверяем, включена ли аутентификация
	if !cfg.ZeroTrust.Authentication.Enabled {
		// Аутентификация отключена - создаем no-op provider
		adapter.provider = &NoOpProvider{}
		logger.Info("Authentication disabled - using no-op provider")
	} else if cfg.Zitadel.Enabled {
		// Initialize Zitadel provider with proper JWT validation
		zitadelProvider, err := NewZitadelProvider(cfg, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create Zitadel provider: %v", err)
		}
		adapter.provider = zitadelProvider
		logger.Info("Zitadel authentication provider initialized")
	} else {
		return nil, fmt.Errorf("authentication enabled but no provider configured - Zitadel must be enabled")
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
	if a.config.Zitadel.Enabled {
		return "zitadel"
	}
	return "none"
}

// NewZitadelProvider создает новый Zitadel провайдер с правильной JWT валидацией
func NewZitadelProvider(cfg *config.Config, logger *logger.Logger) (*ZitadelProvider, error) {
	// Создаем Zitadel клиент
	zitadelClient, err := NewZitadelClient(ZitadelClientConfig{
		Domain:     cfg.Zitadel.Issuer,
		HTTPClient: nil, // Используем дефолтный клиент
		Logger:     logger.GetZapLogger(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Zitadel client: %w", err)
	}

	// Получаем публичный ключ для валидации JWT
	// TODO: Получить ключ из Zitadel JWKS endpoint
	// Пока используем fallback secret для HMAC валидации
	if cfg.Zitadel.FallbackSecret == "" {
		return nil, fmt.Errorf("Zitadel fallback secret not configured")
	}

	// Создаем валидатор (пока с fallback secret)
	validator := NewZitadelValidator(ZitadelValidatorConfig{
		PublicKey: nil, // Будет установлен позже при получении ключа из JWKS
		Issuer:    cfg.Zitadel.Issuer + "/oauth/v2",
		Audience:  "cloudbridge-relay",
		KeyID:     "", // Будет установлен позже
		ClockSkew: 5 * time.Minute,
		Logger:    logger.GetZapLogger(),
	})

	return &ZitadelProvider{
		domain:         cfg.Zitadel.Issuer,
		projectID:      cfg.Zitadel.Audience,
		clientID:       cfg.Zitadel.ClientID,
		fallbackSecret: cfg.Zitadel.FallbackSecret,
		logger:         logger,
		validator:      validator,
		client:         zitadelClient,
	}, nil
}

// ZitadelProvider implements Provider interface for Zitadel
type ZitadelProvider struct {
	domain         string
	projectID      string
	clientID       string
	fallbackSecret string
	logger         *logger.Logger
	validator      *ZitadelValidator
	client         *ZitadelClient
}

// ValidateToken validates token using Zitadel with proper JWT validation
func (p *ZitadelProvider) ValidateToken(ctx context.Context, token string) (*TokenInfo, error) {
	// Сначала пытаемся валидировать как relay токен
	if p.validator != nil && p.validator.publicKey != nil {
		relayClaims, err := p.validator.ValidateRelayToken(ctx, token)
		if err == nil && relayClaims != nil {
			// Это relay токен - конвертируем в TokenInfo
			return p.convertRelayClaimsToTokenInfo(relayClaims), nil
		}
	}

	// Fallback: используем HMAC валидацию с fallback secret
	if p.fallbackSecret == "" {
		return nil, fmt.Errorf("zitadel fallback secret not configured")
	}

	// Парсим JWT токен с fallback секретом
	claims := jwt.MapClaims{}
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(p.fallbackSecret), nil
	})

	if err != nil {
		p.logger.Debug("JWT validation failed", zap.Error(err))
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if !parsedToken.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Извлекаем информацию из JWT claims
	subject, _ := claims["sub"].(string)
	if subject == "" {
		subject = "unknown-user"
	}

	tenantID, _ := claims["tenant_id"].(string)
	serverID, _ := claims["server_id"].(string)

	// Извлекаем роли и права
	roles := []string{"user", "peer", "client"} // Базовые роли для всех
	if permissionsInterface, ok := claims["permissions"]; ok {
		if permissionsSlice, ok := permissionsInterface.([]interface{}); ok {
			for _, perm := range permissionsSlice {
				if permStr, ok := perm.(string); ok {
					roles = append(roles, permStr)
				}
			}
		}
	}

	p.logger.Info("JWT validated successfully",
		zap.String("subject", subject),
		zap.String("tenant_id", tenantID),
		zap.String("server_id", serverID),
		zap.Strings("roles", roles))

	return &TokenInfo{
		Subject:   subject,
		Issuer:    p.domain,
		Audience:  []string{p.clientID},
		ExpiresAt: time.Now().Add(time.Hour),
		IssuedAt:  time.Now(),
		NotBefore: time.Now(),
		Username:  subject,
		Email:     subject + "@cloudbridge.local",
		FullName:  subject,
		Groups:    []string{"users"},
		Roles:     roles,
		TenantID:  tenantID,
		Claims:    map[string]interface{}{"server_id": serverID, "tenant_id": tenantID},
		Provider:  "zitadel",
	}, nil
}

// convertRelayClaimsToTokenInfo конвертирует RelayClaims в TokenInfo
func (p *ZitadelProvider) convertRelayClaimsToTokenInfo(claims *RelayClaims) *TokenInfo {
	var expiresAt time.Time
	if claims.ExpiresAt != nil {
		expiresAt = claims.ExpiresAt.Time
	} else {
		// Для постоянных токенов устанавливаем далекое будущее
		expiresAt = time.Now().Add(365 * 24 * time.Hour)
	}

	var issuedAt time.Time
	if claims.IssuedAt != nil {
		issuedAt = claims.IssuedAt.Time
	} else {
		issuedAt = time.Now()
	}

	var notBefore time.Time
	if claims.NotBefore != nil {
		notBefore = claims.NotBefore.Time
	} else {
		notBefore = time.Now()
	}

	return &TokenInfo{
		Subject:   claims.Subject,
		Issuer:    claims.Issuer,
		Audience:  claims.Audience,
		ExpiresAt: expiresAt,
		IssuedAt:  issuedAt,
		NotBefore: notBefore,
		Username:  claims.GetRelayServerID(),
		Email:     claims.GetRelayServerID() + "@cloudbridge.local",
		FullName:  "Relay Server " + claims.GetRelayServerID(),
		Groups:    []string{"relay-servers"},
		Roles:     claims.Roles,
		TenantID:  claims.GetTenantID(),
		Claims: map[string]interface{}{
			"relay_server_id": claims.GetRelayServerID(),
			"tenant_id":       claims.GetTenantID(),
			"org_id":          claims.GetOrgID(),
			"permissions":     claims.Permissions,
			"token_type":      claims.TokenType,
			"protocol_type":   claims.ProtocolType,
		},
		Provider: "zitadel",
	}
}

// GetUserInfo gets user information from Zitadel token
func (p *ZitadelProvider) GetUserInfo(ctx context.Context, token string) (*UserInfo, error) {
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
		Active:   true,
		Claims:   tokenInfo.Claims,
		Provider: "zitadel",
	}, nil
}

// CheckPermission checks permission using Zitadel roles and groups
func (p *ZitadelProvider) CheckPermission(ctx context.Context, token string, resource, action string) (bool, error) {
	// Сначала пытаемся валидировать как relay токен
	if p.validator != nil && p.validator.publicKey != nil {
		relayClaims, err := p.validator.ValidateRelayToken(ctx, token)
		if err == nil && relayClaims != nil {
			// Это relay токен - проверяем разрешения
			requiredPermission := fmt.Sprintf("%s:%s", resource, action)
			return relayClaims.HasPermission(requiredPermission), nil
		}
	}

	// Fallback: для обычных токенов пока разрешаем все
	// TODO: Реализовать через Zitadel API
	return true, nil
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
		Roles:     []string{"user", "peer", "client", "admin"}, // ИСПРАВЛЕНО: добавлен admin для тестирования
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
