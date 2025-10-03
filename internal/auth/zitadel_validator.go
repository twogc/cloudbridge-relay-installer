package auth

import (
	"context"
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// ZitadelValidator предоставляет валидацию JWT токенов от Zitadel
type ZitadelValidator struct {
	publicKey *rsa.PublicKey
	issuer    string
	audience  string
	logger    *zap.Logger
	keyID     string
	clockSkew time.Duration
}

// ZitadelValidatorConfig конфигурация для ZitadelValidator
type ZitadelValidatorConfig struct {
	PublicKey *rsa.PublicKey
	Issuer    string
	Audience  string
	KeyID     string
	ClockSkew time.Duration
	Logger    *zap.Logger
}

// NewZitadelValidator создает новый валидатор для Zitadel токенов
func NewZitadelValidator(config ZitadelValidatorConfig) *ZitadelValidator {
	if config.ClockSkew == 0 {
		config.ClockSkew = 5 * time.Minute
	}

	return &ZitadelValidator{
		publicKey: config.PublicKey,
		issuer:    config.Issuer,
		audience:  config.Audience,
		keyID:     config.KeyID,
		clockSkew: config.ClockSkew,
		logger:    config.Logger,
	}
}

// ValidateRelayToken валидирует relay токен от Zitadel
func (v *ZitadelValidator) ValidateRelayToken(ctx context.Context, tokenString string) (*RelayClaims, error) {
	// Парсим токен
	token, err := jwt.ParseWithClaims(tokenString, &RelayClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Проверяем алгоритм подписи
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Проверяем key ID если указан
		if v.keyID != "" {
			if kid, ok := token.Header["kid"].(string); ok && kid != v.keyID {
				return nil, fmt.Errorf("unexpected key ID: %v", kid)
			}
		}

		return v.publicKey, nil
	})

	if err != nil {
		v.logger.Error("Failed to parse JWT token", zap.Error(err))
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*RelayClaims)
	if !ok || !token.Valid {
		v.logger.Error("Invalid token claims or token not valid")
		return nil, fmt.Errorf("invalid token claims")
	}

	// Проверяем issuer
	if claims.Issuer != v.issuer {
		v.logger.Error("Invalid issuer",
			zap.String("expected", v.issuer),
			zap.String("actual", claims.Issuer))
		return nil, fmt.Errorf("invalid issuer: expected %s, got %s", v.issuer, claims.Issuer)
	}

	// Проверяем audience
	if !v.validateAudience(claims.Audience) {
		v.logger.Error("Invalid audience",
			zap.String("expected", v.audience),
			zap.Strings("actual", claims.Audience))
		return nil, fmt.Errorf("invalid audience: expected %s, got %v", v.audience, claims.Audience)
	}

	// Проверяем тип токена
	if !claims.IsRelayToken() {
		v.logger.Error("Invalid token type", zap.String("token_type", claims.TokenType))
		return nil, fmt.Errorf("invalid token type: %s", claims.TokenType)
	}

	// Проверяем время истечения только для сессионных токенов
	if !claims.IsPermanentToken() && claims.ExpiresAt != nil {
		now := time.Now()
		if claims.ExpiresAt.Time.Before(now.Add(-v.clockSkew)) {
			v.logger.Error("Token expired",
				zap.Time("expires_at", claims.ExpiresAt.Time),
				zap.Time("now", now))
			return nil, fmt.Errorf("token expired")
		}
	}

	// Проверяем время "не раньше" (nbf)
	if claims.NotBefore != nil {
		now := time.Now()
		if claims.NotBefore.Time.After(now.Add(v.clockSkew)) {
			v.logger.Error("Token not yet valid",
				zap.Time("not_before", claims.NotBefore.Time),
				zap.Time("now", now))
			return nil, fmt.Errorf("token not yet valid")
		}
	}

	v.logger.Info("Successfully validated relay token",
		zap.String("relay_server_id", claims.GetRelayServerID()),
		zap.String("tenant_id", claims.GetTenantID()),
		zap.String("org_id", claims.GetOrgID()),
		zap.String("token_type", claims.TokenType))

	return claims, nil
}

// ValidatePermissions проверяет наличие необходимых разрешений
func (v *ZitadelValidator) ValidatePermissions(claims *RelayClaims, requiredPermissions []string) error {
	if claims == nil {
		return fmt.Errorf("claims is nil")
	}

	if !claims.ValidatePermissions(requiredPermissions) {
		v.logger.Error("Insufficient permissions",
			zap.Strings("required", requiredPermissions),
			zap.Strings("actual", claims.Permissions))
		return fmt.Errorf("insufficient permissions: required %v, got %v", requiredPermissions, claims.Permissions)
	}

	return nil
}

// ValidateRoles проверяет наличие необходимых ролей
func (v *ZitadelValidator) ValidateRoles(claims *RelayClaims, requiredRoles []string) error {
	if claims == nil {
		return fmt.Errorf("claims is nil")
	}

	if !claims.ValidateRoles(requiredRoles) {
		v.logger.Error("Insufficient roles",
			zap.Strings("required", requiredRoles),
			zap.Strings("actual", claims.Roles))
		return fmt.Errorf("insufficient roles: required %v, got %v", requiredRoles, claims.Roles)
	}

	return nil
}

// validateAudience проверяет аудиторию токена
func (v *ZitadelValidator) validateAudience(audience []string) bool {
	if len(audience) == 0 {
		return false
	}

	for _, aud := range audience {
		if aud == v.audience {
			return true
		}
	}

	return false
}

// GetTokenInfo возвращает информацию о токене для логирования
func (v *ZitadelValidator) GetTokenInfo(claims *RelayClaims) map[string]interface{} {
	if claims == nil {
		return nil
	}

	info := map[string]interface{}{
		"subject":         claims.Subject,
		"issuer":          claims.Issuer,
		"audience":        claims.Audience,
		"token_type":      claims.TokenType,
		"relay_server_id": claims.GetRelayServerID(),
		"tenant_id":       claims.GetTenantID(),
		"org_id":          claims.GetOrgID(),
		"permissions":     claims.Permissions,
		"roles":           claims.Roles,
		"is_permanent":    claims.IsPermanentToken(),
	}

	if claims.ExpiresAt != nil {
		info["expires_at"] = claims.ExpiresAt.Time
	}

	if claims.NotBefore != nil {
		info["not_before"] = claims.NotBefore.Time
	}

	return info
}
