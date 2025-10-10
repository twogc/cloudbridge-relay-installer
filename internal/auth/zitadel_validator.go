package auth

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// Ошибки валидации для удобства обработки
var (
	ErrExpired          = errors.New("token expired")
	ErrNotYetValid      = errors.New("token not yet valid")
	ErrInvalidIssuer    = errors.New("invalid issuer")
	ErrInvalidAudience  = errors.New("invalid audience")
	ErrInvalidTokenType = errors.New("invalid token type")
	ErrPermissionDenied = errors.New("insufficient permissions")
	ErrRoleDenied       = errors.New("insufficient roles")
	ErrNotInitialized   = errors.New("validator is not initialized")
)

// ZitadelValidator предоставляет валидацию JWT токенов от Zitadel
type ZitadelValidator struct {
	publicKey         *rsa.PublicKey
	issuer            string
	audience          string
	allowedAudiences  []string
	allowedAlgorithms map[string]bool
	logger            *zap.Logger
	keyID             string
	clockSkew         time.Duration
}

// ZitadelValidatorConfig конфигурация для ZitadelValidator
type ZitadelValidatorConfig struct {
	PublicKey         *rsa.PublicKey
	Issuer            string
	Audience          string
	AllowedAudiences  []string
	AllowedAlgorithms []string
	KeyID             string
	ClockSkew         time.Duration
	Logger            *zap.Logger
}

// NewZitadelValidator создает новый валидатор для Zitadel токенов
func NewZitadelValidator(config ZitadelValidatorConfig) *ZitadelValidator {
	if config.ClockSkew == 0 {
		config.ClockSkew = 5 * time.Minute
	}

	// Инициализируем разрешенные алгоритмы
	allowedAlgorithms := make(map[string]bool)
	if len(config.AllowedAlgorithms) > 0 {
		for _, alg := range config.AllowedAlgorithms {
			allowedAlgorithms[alg] = true
		}
	} else {
		// По умолчанию разрешаем RS256
		allowedAlgorithms[jwt.SigningMethodRS256.Alg()] = true
	}

	// Инициализируем разрешенные аудитории
	allowedAudiences := config.AllowedAudiences
	if len(allowedAudiences) == 0 && config.Audience != "" {
		allowedAudiences = []string{config.Audience}
	}

	return &ZitadelValidator{
		publicKey:         config.PublicKey,
		issuer:            config.Issuer,
		audience:          config.Audience,
		allowedAudiences:  allowedAudiences,
		allowedAlgorithms: allowedAlgorithms,
		keyID:             config.KeyID,
		clockSkew:         config.ClockSkew,
		logger:            config.Logger,
	}
}

// ValidateRelayToken валидирует relay токен от Zitadel
func (v *ZitadelValidator) ValidateRelayToken(ctx context.Context, tokenString string) (*RelayClaims, error) {
	if v.publicKey == nil {
		v.logger.Warn("Zitadel public key is nil")
		return nil, ErrNotInitialized
	}

	// Создаем список разрешенных алгоритмов для парсера
	validMethods := make([]string, 0, len(v.allowedAlgorithms))
	for alg := range v.allowedAlgorithms {
		validMethods = append(validMethods, alg)
	}

	// Создаем парсер с встроенными проверками
	parser := jwt.NewParser(
		jwt.WithValidMethods(validMethods),
		jwt.WithIssuer(v.issuer),
		jwt.WithAudience(v.allowedAudiences...),
		jwt.WithLeeway(v.clockSkew),
	)

	token, err := parser.ParseWithClaims(tokenString, &RelayClaims{}, func(t *jwt.Token) (interface{}, error) {
		alg := t.Method.Alg()
		if !v.allowedAlgorithms[alg] {
			return nil, fmt.Errorf("unexpected signing alg: %s", alg)
		}

		// Проверяем key ID (мягко - только логируем)
		if v.keyID != "" {
			if kid, _ := t.Header["kid"].(string); kid != v.keyID {
				v.logger.Warn("JWT kid mismatch",
					zap.String("expected", v.keyID),
					zap.String("actual", kid))
			}
		}

		return v.publicKey, nil
	})

	if err != nil {
		v.logger.Warn("JWT parse/validate failed", zap.Error(err))
		return nil, fmt.Errorf("token parse/validate failed: %w", err)
	}

	claims, ok := token.Claims.(*RelayClaims)
	if !ok || !token.Valid {
		v.logger.Warn("JWT invalid claims or token invalid")
		return nil, fmt.Errorf("invalid token claims")
	}

	// Дополнительные доменные проверки
	if !claims.IsRelayToken() {
		v.logger.Warn("Invalid token type", zap.String("token_type", claims.TokenType))
		return nil, ErrInvalidTokenType
	}

	// Для permanent токенов можно смягчить проверки времени
	if claims.IsPermanentToken() {
		// Parser уже проверил exp/nbf с учетом leeway
		// Если нужно игнорировать exp для permanent токенов, можно добавить логику здесь
	}

	v.logger.Info("Relay token validated",
		zap.String("relay_server_id", claims.GetRelayServerID()),
		zap.String("tenant_id", claims.GetTenantID()),
		zap.String("org_id", claims.GetOrgID()),
		zap.String("token_type", claims.TokenType),
	)

	return claims, nil
}

// ValidatePermissions проверяет наличие необходимых разрешений
func (v *ZitadelValidator) ValidatePermissions(claims *RelayClaims, requiredPermissions []string) error {
	if claims == nil {
		return fmt.Errorf("claims is nil")
	}

	if !claims.ValidatePermissions(requiredPermissions) {
		v.logger.Warn("Insufficient permissions",
			zap.Strings("required", requiredPermissions),
			zap.Strings("actual", claims.Permissions))
		return ErrPermissionDenied
	}

	return nil
}

// ValidateRoles проверяет наличие необходимых ролей
func (v *ZitadelValidator) ValidateRoles(claims *RelayClaims, requiredRoles []string) error {
	if claims == nil {
		return fmt.Errorf("claims is nil")
	}

	if !claims.ValidateRoles(requiredRoles) {
		v.logger.Warn("Insufficient roles",
			zap.Strings("required", requiredRoles),
			zap.Strings("actual", claims.Roles))
		return ErrRoleDenied
	}

	return nil
}

// validateAudienceList проверяет аудиторию токена против списка разрешенных
func (v *ZitadelValidator) validateAudienceList(aud []string, allowed []string) bool {
	if len(aud) == 0 || len(allowed) == 0 {
		return false
	}

	allowedSet := make(map[string]struct{}, len(allowed))
	for _, a := range allowed {
		allowedSet[a] = struct{}{}
	}

	for _, a := range aud {
		if _, ok := allowedSet[a]; ok {
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
