package auth

import (
	"context"
	"fmt"

	"github.com/twogc/cloudbridge-relay-installer/internal/quic"
	"go.uber.org/zap"
)

// KeycloakConfig конфигурация Keycloak для совместимости
type KeycloakConfig struct {
	ServerURL     string `json:"server_url"`
	Realm         string `json:"realm"`
	ClientID      string `json:"client_id"`
	ClientSecret  string `json:"client_secret"`
	AdminUsername string `json:"admin_username"`
	AdminPassword string `json:"admin_password"`
}

// EnhancedKeycloakAuth расширяет существующую аутентификацию Enhanced QUIC функциями
type EnhancedKeycloakAuth struct {
	// Существующие поля
	config *KeycloakConfig
	logger *zap.Logger

	// Новое поле для Enhanced QUIC
	enhancedKeycloakAuth KeycloakAuthInterface
	enhancedEnabled      bool
}

// NewEnhancedKeycloakAuth создает новый Enhanced Keycloak Auth
func NewEnhancedKeycloakAuth(config *KeycloakConfig, logger *zap.Logger) *EnhancedKeycloakAuth {
	if config == nil {
		config = &KeycloakConfig{}
	}

	enhancedConfig := &quic.KeycloakConfig{
		ServerURL:     config.ServerURL,
		Realm:         config.Realm,
		ClientID:      config.ClientID,
		ClientSecret:  config.ClientSecret,
		AdminUsername: config.AdminUsername,
		AdminPassword: config.AdminPassword,
	}

	return &EnhancedKeycloakAuth{
		config:               config,
		logger:               logger,
		enhancedKeycloakAuth: quic.NewKeycloakAuth(enhancedConfig),
		enhancedEnabled:      true,
	}
}

// NewKeycloakAuthCompatible создает совместимую версию без Enhanced функций
func NewKeycloakAuthCompatible(config *KeycloakConfig, logger *zap.Logger) *EnhancedKeycloakAuth {
	return &EnhancedKeycloakAuth{
		config:          config,
		logger:          logger,
		enhancedEnabled: false,
	}
}

// AuthenticateEnhanced аутентификация с Enhanced QUIC
func (eka *EnhancedKeycloakAuth) AuthenticateEnhanced(ctx context.Context, token string) (bool, error) {
	if eka.enhancedEnabled && eka.enhancedKeycloakAuth != nil {
		isValid, err := eka.enhancedKeycloakAuth.AuthenticateUser(token)
		if err != nil {
			if eka.logger != nil {
				eka.logger.Error("Enhanced Keycloak authentication failed",
					zap.Error(err),
					zap.String("token_prefix", token[:min(len(token), 10)]+"..."),
				)
			}
			return false, err
		}

		if eka.logger != nil {
			eka.logger.Info("Enhanced Keycloak authentication successful",
				zap.Bool("is_valid", isValid),
			)
		}

		return isValid, nil
	}

	// Fallback к обычной аутентификации
	return eka.AuthenticateBasic(ctx, token)
}

// AuthenticateBasic базовая аутентификация (заглушка)
func (eka *EnhancedKeycloakAuth) AuthenticateBasic(ctx context.Context, token string) (bool, error) {
	// Простая проверка токена для fallback
	if token == "" {
		return false, fmt.Errorf("empty token")
	}

	// В реальной реализации здесь будет полноценная аутентификация
	if eka.logger != nil {
		eka.logger.Debug("Basic authentication (fallback)",
			zap.String("token_prefix", token[:min(len(token), 10)]+"..."),
		)
	}

	return len(token) > 10, nil // Простая проверка для примера
}

// ValidateTokenEnhanced валидирует токен через Enhanced QUIC
func (eka *EnhancedKeycloakAuth) ValidateTokenEnhanced(token string) (bool, error) {
	if eka.enhancedEnabled && eka.enhancedKeycloakAuth != nil {
		return eka.enhancedKeycloakAuth.ValidateToken(token)
	}

	// Fallback к базовой валидации
	return eka.validateTokenBasic(token)
}

// validateTokenBasic базовая валидация токена
func (eka *EnhancedKeycloakAuth) validateTokenBasic(token string) (bool, error) {
	if token == "" {
		return false, fmt.Errorf("empty token")
	}

	// Простая проверка формата JWT токена
	if len(token) < 20 {
		return false, fmt.Errorf("token too short")
	}

	// Проверяем наличие точек (характерно для JWT)
	dotCount := 0
	for _, char := range token {
		if char == '.' {
			dotCount++
		}
	}

	if dotCount < 2 {
		return false, fmt.Errorf("invalid token format")
	}

	return true, nil
}

// EnableEnhanced включает Enhanced Keycloak Auth
func (eka *EnhancedKeycloakAuth) EnableEnhanced(config *KeycloakConfig) error {
	if eka.enhancedEnabled {
		return nil // Уже включен
	}

	if config == nil {
		config = eka.config
	}

	enhancedConfig := &quic.KeycloakConfig{
		ServerURL:     config.ServerURL,
		Realm:         config.Realm,
		ClientID:      config.ClientID,
		ClientSecret:  config.ClientSecret,
		AdminUsername: config.AdminUsername,
		AdminPassword: config.AdminPassword,
	}

	eka.enhancedKeycloakAuth = quic.NewKeycloakAuth(enhancedConfig)
	eka.enhancedEnabled = true

	if eka.logger != nil {
		eka.logger.Info("Enhanced Keycloak Auth enabled")
	}
	return nil
}

// DisableEnhanced отключает Enhanced Keycloak Auth
func (eka *EnhancedKeycloakAuth) DisableEnhanced() {
	eka.enhancedEnabled = false
	eka.enhancedKeycloakAuth = nil

	if eka.logger != nil {
		eka.logger.Info("Enhanced Keycloak Auth disabled")
	}
}

// IsEnhancedEnabled проверяет, включен ли Enhanced режим
func (eka *EnhancedKeycloakAuth) IsEnhancedEnabled() bool {
	return eka.enhancedEnabled && eka.enhancedKeycloakAuth != nil
}

// GetEnhancedKeycloakAuth возвращает Enhanced Keycloak Auth
func (eka *EnhancedKeycloakAuth) GetEnhancedKeycloakAuth() KeycloakAuthInterface {
	if eka.enhancedEnabled {
		return eka.enhancedKeycloakAuth
	}
	return nil
}

// GetConfig возвращает конфигурацию Keycloak
func (eka *EnhancedKeycloakAuth) GetConfig() *KeycloakConfig {
	return eka.config
}

// UpdateConfig обновляет конфигурацию Keycloak
func (eka *EnhancedKeycloakAuth) UpdateConfig(config *KeycloakConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	eka.config = config

	// Обновляем Enhanced конфигурацию, если включена
	if eka.enhancedEnabled {
		return eka.EnableEnhanced(config)
	}

	return nil
}

// Вспомогательная функция min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
