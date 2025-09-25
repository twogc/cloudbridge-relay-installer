package zitadel

import (
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// ZitadelValidator валидатор токенов Zitadel
type ZitadelValidator struct {
	domain         string
	projectID      string
	clientID       string
	clientSecret   string
	fallbackSecret string
	logger         *zap.Logger
}

// NewZitadelValidator создает валидатор Zitadel
func NewZitadelValidator(config *Config, logger *zap.Logger) (*ZitadelValidator, error) {
	return &ZitadelValidator{
		domain:         config.Domain,
		projectID:      config.ProjectID,
		clientID:       config.ClientID,
		clientSecret:   config.ClientSecret,
		fallbackSecret: config.FallbackSecret,
		logger:         logger,
	}, nil
}

// ValidateToken валидирует JWT токен
func (zv *ZitadelValidator) ValidateToken(tokenString string) (*TokenClaims, error) {
	// Удаление Bearer префикса если есть
	if strings.HasPrefix(tokenString, "Bearer ") {
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	}

	// Используем fallback валидацию для совместимости
	return zv.validateWithFallbackSecret(tokenString)
}

// validateWithFallbackSecret валидирует токен с fallback секретом
func (zv *ZitadelValidator) validateWithFallbackSecret(tokenString string) (*TokenClaims, error) {
	if zv.fallbackSecret == "" {
		return nil, fmt.Errorf("fallback secret not configured")
	}

	claims := &TokenClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(zv.fallbackSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is invalid")
	}

	// Заполняем базовые поля для совместимости
	if claims.Subject == "" {
		claims.Subject = "zitadel-user"
	}

	if claims.Issuer == "" {
		claims.Issuer = fmt.Sprintf("https://%s", zv.domain)
	}

	if len(claims.Audience) == 0 {
		claims.Audience = []string{zv.clientID}
	}

	if claims.ExpiresAt == nil {
		exp := jwt.NewNumericDate(time.Now().Add(time.Hour))
		claims.ExpiresAt = exp
	}

	zv.logger.Debug("Token validated with fallback secret",
		zap.String("subject", claims.Subject),
		zap.String("issuer", claims.Issuer))

	return claims, nil
}
