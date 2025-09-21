package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockKeycloakAuth is a mock implementation of KeycloakAuthInterface
type MockKeycloakAuth struct {
	mock.Mock
}

func (m *MockKeycloakAuth) AuthenticateUser(token string) (bool, error) {
	args := m.Called(token)
	return args.Bool(0), args.Error(1)
}

func (m *MockKeycloakAuth) ValidateToken(token string) (bool, error) {
	args := m.Called(token)
	return args.Bool(0), args.Error(1)
}

func TestNewEnhancedKeycloakAuth(t *testing.T) {
	t.Run("creates enhanced keycloak auth with config", func(t *testing.T) {
		config := &KeycloakConfig{
			ServerURL:     "https://keycloak.example.com",
			Realm:         "test-realm",
			ClientID:      "test-client",
			ClientSecret:  "test-secret",
			AdminUsername: "admin",
			AdminPassword: "password",
		}
		logger := zap.NewNop()

		auth := NewEnhancedKeycloakAuth(config, logger)

		assert.NotNil(t, auth)
		assert.Equal(t, config, auth.config)
		assert.Equal(t, logger, auth.logger)
		assert.True(t, auth.enhancedEnabled)
		assert.NotNil(t, auth.enhancedKeycloakAuth)
	})

	t.Run("creates enhanced keycloak auth with nil config", func(t *testing.T) {
		logger := zap.NewNop()

		auth := NewEnhancedKeycloakAuth(nil, logger)

		assert.NotNil(t, auth)
		assert.NotNil(t, auth.config)
		assert.Equal(t, logger, auth.logger)
		assert.True(t, auth.enhancedEnabled)
		assert.NotNil(t, auth.enhancedKeycloakAuth)
	})

	t.Run("creates enhanced keycloak auth with nil logger", func(t *testing.T) {
		config := &KeycloakConfig{
			ServerURL: "https://keycloak.example.com",
			Realm:     "test-realm",
		}

		auth := NewEnhancedKeycloakAuth(config, nil)

		assert.NotNil(t, auth)
		assert.Equal(t, config, auth.config)
		assert.Nil(t, auth.logger)
		assert.True(t, auth.enhancedEnabled)
		assert.NotNil(t, auth.enhancedKeycloakAuth)
	})
}

func TestNewKeycloakAuthCompatible(t *testing.T) {
	t.Run("creates compatible keycloak auth", func(t *testing.T) {
		config := &KeycloakConfig{
			ServerURL: "https://keycloak.example.com",
			Realm:     "test-realm",
		}
		logger := zap.NewNop()

		auth := NewKeycloakAuthCompatible(config, logger)

		assert.NotNil(t, auth)
		assert.Equal(t, config, auth.config)
		assert.Equal(t, logger, auth.logger)
		assert.False(t, auth.enhancedEnabled)
		assert.Nil(t, auth.enhancedKeycloakAuth)
	})

	t.Run("creates compatible keycloak auth with nil config", func(t *testing.T) {
		logger := zap.NewNop()

		auth := NewKeycloakAuthCompatible(nil, logger)

		assert.NotNil(t, auth)
		assert.Nil(t, auth.config)
		assert.Equal(t, logger, auth.logger)
		assert.False(t, auth.enhancedEnabled)
		assert.Nil(t, auth.enhancedKeycloakAuth)
	})
}

func TestEnhancedKeycloakAuth_AuthenticateEnhanced(t *testing.T) {
	t.Run("authenticates with enhanced mode enabled", func(t *testing.T) {
		config := &KeycloakConfig{
			ServerURL: "https://keycloak.example.com",
			Realm:     "test-realm",
		}
		logger := zap.NewNop()
		auth := NewEnhancedKeycloakAuth(config, logger)

		// Mock the enhanced keycloak auth
		mockEnhanced := new(MockKeycloakAuth)
		auth.enhancedKeycloakAuth = mockEnhanced

		mockEnhanced.On("AuthenticateUser", "valid-token").Return(true, nil)

		ctx := context.Background()
		isValid, err := auth.AuthenticateEnhanced(ctx, "valid-token")

		assert.NoError(t, err)
		assert.True(t, isValid)
		mockEnhanced.AssertExpectations(t)
	})

	t.Run("handles enhanced authentication error", func(t *testing.T) {
		config := &KeycloakConfig{
			ServerURL: "https://keycloak.example.com",
			Realm:     "test-realm",
		}
		logger := zap.NewNop()
		auth := NewEnhancedKeycloakAuth(config, logger)

		// Mock the enhanced keycloak auth
		mockEnhanced := new(MockKeycloakAuth)
		auth.enhancedKeycloakAuth = mockEnhanced

		mockEnhanced.On("AuthenticateUser", "invalid-token").Return(false, assert.AnError)

		ctx := context.Background()
		isValid, err := auth.AuthenticateEnhanced(ctx, "invalid-token")

		assert.Error(t, err)
		assert.False(t, isValid)
		mockEnhanced.AssertExpectations(t)
	})

	t.Run("falls back to basic authentication when enhanced disabled", func(t *testing.T) {
		config := &KeycloakConfig{
			ServerURL: "https://keycloak.example.com",
			Realm:     "test-realm",
		}
		logger := zap.NewNop()
		auth := NewKeycloakAuthCompatible(config, logger)

		ctx := context.Background()
		isValid, err := auth.AuthenticateEnhanced(ctx, "valid-token")

		assert.NoError(t, err)
		assert.True(t, isValid) // Should pass basic validation
	})

	t.Run("falls back to basic authentication when enhanced auth is nil", func(t *testing.T) {
		config := &KeycloakConfig{
			ServerURL: "https://keycloak.example.com",
			Realm:     "test-realm",
		}
		logger := zap.NewNop()
		auth := NewEnhancedKeycloakAuth(config, logger)
		auth.enhancedKeycloakAuth = nil // Simulate nil enhanced auth

		ctx := context.Background()
		isValid, err := auth.AuthenticateEnhanced(ctx, "valid-token")

		assert.NoError(t, err)
		assert.True(t, isValid) // Should pass basic validation
	})
}

func TestEnhancedKeycloakAuth_AuthenticateBasic(t *testing.T) {
	t.Run("authenticates valid token", func(t *testing.T) {
		config := &KeycloakConfig{
			ServerURL: "https://keycloak.example.com",
			Realm:     "test-realm",
		}
		logger := zap.NewNop()
		auth := NewKeycloakAuthCompatible(config, logger)

		ctx := context.Background()
		isValid, err := auth.AuthenticateBasic(ctx, "valid-long-token")

		assert.NoError(t, err)
		assert.True(t, isValid)
	})

	t.Run("rejects empty token", func(t *testing.T) {
		config := &KeycloakConfig{
			ServerURL: "https://keycloak.example.com",
			Realm:     "test-realm",
		}
		logger := zap.NewNop()
		auth := NewKeycloakAuthCompatible(config, logger)

		ctx := context.Background()
		isValid, err := auth.AuthenticateBasic(ctx, "")

		assert.Error(t, err)
		assert.False(t, isValid)
		assert.Contains(t, err.Error(), "empty token")
	})

	t.Run("rejects short token", func(t *testing.T) {
		config := &KeycloakConfig{
			ServerURL: "https://keycloak.example.com",
			Realm:     "test-realm",
		}
		logger := zap.NewNop()
		auth := NewKeycloakAuthCompatible(config, logger)

		ctx := context.Background()
		isValid, err := auth.AuthenticateBasic(ctx, "short")

		assert.NoError(t, err)
		assert.False(t, isValid) // Token too short
	})

	t.Run("handles nil logger", func(t *testing.T) {
		config := &KeycloakConfig{
			ServerURL: "https://keycloak.example.com",
			Realm:     "test-realm",
		}
		auth := NewKeycloakAuthCompatible(config, nil)

		ctx := context.Background()
		isValid, err := auth.AuthenticateBasic(ctx, "valid-long-token")

		assert.NoError(t, err)
		assert.True(t, isValid)
	})
}

func TestEnhancedKeycloakAuth_ValidateTokenEnhanced(t *testing.T) {
	t.Run("validates token with enhanced mode enabled", func(t *testing.T) {
		config := &KeycloakConfig{
			ServerURL: "https://keycloak.example.com",
			Realm:     "test-realm",
		}
		logger := zap.NewNop()
		auth := NewEnhancedKeycloakAuth(config, logger)

		// Mock the enhanced keycloak auth
		mockEnhanced := new(MockKeycloakAuth)
		auth.enhancedKeycloakAuth = mockEnhanced

		mockEnhanced.On("ValidateToken", "valid-token").Return(true, nil)

		isValid, err := auth.ValidateTokenEnhanced("valid-token")

		assert.NoError(t, err)
		assert.True(t, isValid)
		mockEnhanced.AssertExpectations(t)
	})

	t.Run("handles enhanced validation error", func(t *testing.T) {
		config := &KeycloakConfig{
			ServerURL: "https://keycloak.example.com",
			Realm:     "test-realm",
		}
		logger := zap.NewNop()
		auth := NewEnhancedKeycloakAuth(config, logger)

		// Mock the enhanced keycloak auth
		mockEnhanced := new(MockKeycloakAuth)
		auth.enhancedKeycloakAuth = mockEnhanced

		mockEnhanced.On("ValidateToken", "invalid-token").Return(false, assert.AnError)

		isValid, err := auth.ValidateTokenEnhanced("invalid-token")

		assert.Error(t, err)
		assert.False(t, isValid)
		mockEnhanced.AssertExpectations(t)
	})

	t.Run("falls back to basic validation when enhanced disabled", func(t *testing.T) {
		config := &KeycloakConfig{
			ServerURL: "https://keycloak.example.com",
			Realm:     "test-realm",
		}
		logger := zap.NewNop()
		auth := NewKeycloakAuthCompatible(config, logger)

		isValid, err := auth.ValidateTokenEnhanced("header.payload.signature")

		assert.NoError(t, err)
		assert.True(t, isValid)
	})

	t.Run("falls back to basic validation when enhanced auth is nil", func(t *testing.T) {
		config := &KeycloakConfig{
			ServerURL: "https://keycloak.example.com",
			Realm:     "test-realm",
		}
		logger := zap.NewNop()
		auth := NewEnhancedKeycloakAuth(config, logger)
		auth.enhancedKeycloakAuth = nil // Simulate nil enhanced auth

		isValid, err := auth.ValidateTokenEnhanced("header.payload.signature")

		assert.NoError(t, err)
		assert.True(t, isValid)
	})
}

func TestEnhancedKeycloakAuth_validateTokenBasic(t *testing.T) {
	t.Run("validates valid JWT token", func(t *testing.T) {
		config := &KeycloakConfig{
			ServerURL: "https://keycloak.example.com",
			Realm:     "test-realm",
		}
		logger := zap.NewNop()
		auth := NewKeycloakAuthCompatible(config, logger)

		isValid, err := auth.validateTokenBasic("header.payload.signature")

		assert.NoError(t, err)
		assert.True(t, isValid)
	})

	t.Run("rejects empty token", func(t *testing.T) {
		config := &KeycloakConfig{
			ServerURL: "https://keycloak.example.com",
			Realm:     "test-realm",
		}
		logger := zap.NewNop()
		auth := NewKeycloakAuthCompatible(config, logger)

		isValid, err := auth.validateTokenBasic("")

		assert.Error(t, err)
		assert.False(t, isValid)
		assert.Contains(t, err.Error(), "empty token")
	})

	t.Run("rejects short token", func(t *testing.T) {
		config := &KeycloakConfig{
			ServerURL: "https://keycloak.example.com",
			Realm:     "test-realm",
		}
		logger := zap.NewNop()
		auth := NewKeycloakAuthCompatible(config, logger)

		isValid, err := auth.validateTokenBasic("short")

		assert.Error(t, err)
		assert.False(t, isValid)
		assert.Contains(t, err.Error(), "token too short")
	})

	t.Run("rejects token with insufficient dots", func(t *testing.T) {
		config := &KeycloakConfig{
			ServerURL: "https://keycloak.example.com",
			Realm:     "test-realm",
		}
		logger := zap.NewNop()
		auth := NewKeycloakAuthCompatible(config, logger)

		isValid, err := auth.validateTokenBasic("thisisalongtokenwithnodots")

		assert.Error(t, err)
		assert.False(t, isValid)
		assert.Contains(t, err.Error(), "invalid token format")
	})

	t.Run("rejects token with no dots", func(t *testing.T) {
		config := &KeycloakConfig{
			ServerURL: "https://keycloak.example.com",
			Realm:     "test-realm",
		}
		logger := zap.NewNop()
		auth := NewKeycloakAuthCompatible(config, logger)

		isValid, err := auth.validateTokenBasic("thisisalongtokenwithnodots")

		assert.Error(t, err)
		assert.False(t, isValid)
		assert.Contains(t, err.Error(), "invalid token format")
	})

	t.Run("validates token with multiple dots", func(t *testing.T) {
		config := &KeycloakConfig{
			ServerURL: "https://keycloak.example.com",
			Realm:     "test-realm",
		}
		logger := zap.NewNop()
		auth := NewKeycloakAuthCompatible(config, logger)

		isValid, err := auth.validateTokenBasic("header.payload.signature.extra")

		assert.NoError(t, err)
		assert.True(t, isValid)
	})
}

func TestEnhancedKeycloakAuth_EnableEnhanced(t *testing.T) {
	t.Run("enables enhanced mode with new config", func(t *testing.T) {
		config := &KeycloakConfig{
			ServerURL: "https://keycloak.example.com",
			Realm:     "test-realm",
		}
		logger := zap.NewNop()
		auth := NewKeycloakAuthCompatible(config, logger)

		newConfig := &KeycloakConfig{
			ServerURL:     "https://new-keycloak.example.com",
			Realm:         "new-realm",
			ClientID:      "new-client",
			ClientSecret:  "new-secret",
			AdminUsername: "new-admin",
			AdminPassword: "new-password",
		}

		err := auth.EnableEnhanced(newConfig)

		assert.NoError(t, err)
		assert.True(t, auth.enhancedEnabled)
		assert.NotNil(t, auth.enhancedKeycloakAuth)
	})

	t.Run("enables enhanced mode with nil config", func(t *testing.T) {
		config := &KeycloakConfig{
			ServerURL: "https://keycloak.example.com",
			Realm:     "test-realm",
		}
		logger := zap.NewNop()
		auth := NewKeycloakAuthCompatible(config, logger)

		err := auth.EnableEnhanced(nil)

		assert.NoError(t, err)
		assert.True(t, auth.enhancedEnabled)
		assert.NotNil(t, auth.enhancedKeycloakAuth)
	})

	t.Run("does nothing when already enabled", func(t *testing.T) {
		config := &KeycloakConfig{
			ServerURL: "https://keycloak.example.com",
			Realm:     "test-realm",
		}
		logger := zap.NewNop()
		auth := NewEnhancedKeycloakAuth(config, logger)

		// Should already be enabled
		assert.True(t, auth.enhancedEnabled)

		err := auth.EnableEnhanced(config)

		assert.NoError(t, err)
		assert.True(t, auth.enhancedEnabled)
	})
}

func TestEnhancedKeycloakAuth_DisableEnhanced(t *testing.T) {
	t.Run("disables enhanced mode", func(t *testing.T) {
		config := &KeycloakConfig{
			ServerURL: "https://keycloak.example.com",
			Realm:     "test-realm",
		}
		logger := zap.NewNop()
		auth := NewEnhancedKeycloakAuth(config, logger)

		// Should be enabled initially
		assert.True(t, auth.enhancedEnabled)
		assert.NotNil(t, auth.enhancedKeycloakAuth)

		auth.DisableEnhanced()

		assert.False(t, auth.enhancedEnabled)
		assert.Nil(t, auth.enhancedKeycloakAuth)
	})

	t.Run("handles nil logger", func(t *testing.T) {
		config := &KeycloakConfig{
			ServerURL: "https://keycloak.example.com",
			Realm:     "test-realm",
		}
		auth := NewEnhancedKeycloakAuth(config, nil)

		auth.DisableEnhanced()

		assert.False(t, auth.enhancedEnabled)
		assert.Nil(t, auth.enhancedKeycloakAuth)
	})
}

func TestEnhancedKeycloakAuth_IsEnhancedEnabled(t *testing.T) {
	t.Run("returns true when enhanced enabled", func(t *testing.T) {
		config := &KeycloakConfig{
			ServerURL: "https://keycloak.example.com",
			Realm:     "test-realm",
		}
		logger := zap.NewNop()
		auth := NewEnhancedKeycloakAuth(config, logger)

		assert.True(t, auth.IsEnhancedEnabled())
	})

	t.Run("returns false when enhanced disabled", func(t *testing.T) {
		config := &KeycloakConfig{
			ServerURL: "https://keycloak.example.com",
			Realm:     "test-realm",
		}
		logger := zap.NewNop()
		auth := NewKeycloakAuthCompatible(config, logger)

		assert.False(t, auth.IsEnhancedEnabled())
	})

	t.Run("returns false when enhanced auth is nil", func(t *testing.T) {
		config := &KeycloakConfig{
			ServerURL: "https://keycloak.example.com",
			Realm:     "test-realm",
		}
		logger := zap.NewNop()
		auth := NewEnhancedKeycloakAuth(config, logger)
		auth.enhancedKeycloakAuth = nil

		assert.False(t, auth.IsEnhancedEnabled())
	})
}

func TestEnhancedKeycloakAuth_GetEnhancedKeycloakAuth(t *testing.T) {
	t.Run("returns enhanced auth when enabled", func(t *testing.T) {
		config := &KeycloakConfig{
			ServerURL: "https://keycloak.example.com",
			Realm:     "test-realm",
		}
		logger := zap.NewNop()
		auth := NewEnhancedKeycloakAuth(config, logger)

		enhancedAuth := auth.GetEnhancedKeycloakAuth()

		assert.NotNil(t, enhancedAuth)
	})

	t.Run("returns nil when enhanced disabled", func(t *testing.T) {
		config := &KeycloakConfig{
			ServerURL: "https://keycloak.example.com",
			Realm:     "test-realm",
		}
		logger := zap.NewNop()
		auth := NewKeycloakAuthCompatible(config, logger)

		enhancedAuth := auth.GetEnhancedKeycloakAuth()

		assert.Nil(t, enhancedAuth)
	})
}

func TestEnhancedKeycloakAuth_GetConfig(t *testing.T) {
	t.Run("returns config", func(t *testing.T) {
		config := &KeycloakConfig{
			ServerURL: "https://keycloak.example.com",
			Realm:     "test-realm",
		}
		logger := zap.NewNop()
		auth := NewEnhancedKeycloakAuth(config, logger)

		returnedConfig := auth.GetConfig()

		assert.Equal(t, config, returnedConfig)
	})

	t.Run("returns nil config", func(t *testing.T) {
		logger := zap.NewNop()
		auth := NewEnhancedKeycloakAuth(nil, logger)

		returnedConfig := auth.GetConfig()

		assert.NotNil(t, returnedConfig) // Should return default config
	})
}

func TestEnhancedKeycloakAuth_UpdateConfig(t *testing.T) {
	t.Run("updates config successfully", func(t *testing.T) {
		config := &KeycloakConfig{
			ServerURL: "https://keycloak.example.com",
			Realm:     "test-realm",
		}
		logger := zap.NewNop()
		auth := NewKeycloakAuthCompatible(config, logger)

		newConfig := &KeycloakConfig{
			ServerURL: "https://new-keycloak.example.com",
			Realm:     "new-realm",
		}

		err := auth.UpdateConfig(newConfig)

		assert.NoError(t, err)
		assert.Equal(t, newConfig, auth.config)
	})

	t.Run("updates config and enables enhanced", func(t *testing.T) {
		config := &KeycloakConfig{
			ServerURL: "https://keycloak.example.com",
			Realm:     "test-realm",
		}
		logger := zap.NewNop()
		auth := NewKeycloakAuthCompatible(config, logger)

		// Enable enhanced first
		err := auth.EnableEnhanced(config)
		assert.NoError(t, err)

		newConfig := &KeycloakConfig{
			ServerURL: "https://new-keycloak.example.com",
			Realm:     "new-realm",
		}

		err = auth.UpdateConfig(newConfig)

		assert.NoError(t, err)
		assert.Equal(t, newConfig, auth.config)
		assert.True(t, auth.enhancedEnabled)
	})

	t.Run("returns error for nil config", func(t *testing.T) {
		config := &KeycloakConfig{
			ServerURL: "https://keycloak.example.com",
			Realm:     "test-realm",
		}
		logger := zap.NewNop()
		auth := NewEnhancedKeycloakAuth(config, logger)

		err := auth.UpdateConfig(nil)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "config cannot be nil")
	})
}

func TestMin(t *testing.T) {
	t.Run("returns smaller value", func(t *testing.T) {
		result := min(5, 10)
		assert.Equal(t, 5, result)

		result = min(10, 5)
		assert.Equal(t, 5, result)
	})

	t.Run("returns same value when equal", func(t *testing.T) {
		result := min(5, 5)
		assert.Equal(t, 5, result)
	})

	t.Run("handles negative values", func(t *testing.T) {
		result := min(-5, 10)
		assert.Equal(t, -5, result)

		result = min(10, -5)
		assert.Equal(t, -5, result)
	})

	t.Run("handles zero", func(t *testing.T) {
		result := min(0, 10)
		assert.Equal(t, 0, result)

		result = min(10, 0)
		assert.Equal(t, 0, result)
	})
}

func TestKeycloakConfig_Struct(t *testing.T) {
	t.Run("creates keycloak config", func(t *testing.T) {
		config := &KeycloakConfig{
			ServerURL:     "https://keycloak.example.com",
			Realm:         "test-realm",
			ClientID:      "test-client",
			ClientSecret:  "test-secret",
			AdminUsername: "admin",
			AdminPassword: "password",
		}

		assert.Equal(t, "https://keycloak.example.com", config.ServerURL)
		assert.Equal(t, "test-realm", config.Realm)
		assert.Equal(t, "test-client", config.ClientID)
		assert.Equal(t, "test-secret", config.ClientSecret)
		assert.Equal(t, "admin", config.AdminUsername)
		assert.Equal(t, "password", config.AdminPassword)
	})

	t.Run("creates empty keycloak config", func(t *testing.T) {
		config := &KeycloakConfig{}

		assert.Empty(t, config.ServerURL)
		assert.Empty(t, config.Realm)
		assert.Empty(t, config.ClientID)
		assert.Empty(t, config.ClientSecret)
		assert.Empty(t, config.AdminUsername)
		assert.Empty(t, config.AdminPassword)
	})
}

func TestEnhancedKeycloakAuth_EdgeCases(t *testing.T) {
	t.Run("handles very long token", func(t *testing.T) {
		config := &KeycloakConfig{
			ServerURL: "https://keycloak.example.com",
			Realm:     "test-realm",
		}
		logger := zap.NewNop()
		auth := NewKeycloakAuthCompatible(config, logger)

		// Create a very long token
		longToken := "header.payload.signature"
		for i := 0; i < 100; i++ {
			longToken += ".extra"
		}

		ctx := context.Background()
		isValid, err := auth.AuthenticateBasic(ctx, longToken)

		assert.NoError(t, err)
		assert.True(t, isValid)
	})

	t.Run("handles token with special characters", func(t *testing.T) {
		config := &KeycloakConfig{
			ServerURL: "https://keycloak.example.com",
			Realm:     "test-realm",
		}
		logger := zap.NewNop()
		auth := NewKeycloakAuthCompatible(config, logger)

		ctx := context.Background()
		isValid, err := auth.AuthenticateBasic(ctx, "header!@#$%^&*().payload.signature")

		assert.NoError(t, err)
		assert.True(t, isValid)
	})

	t.Run("handles concurrent access", func(t *testing.T) {
		config := &KeycloakConfig{
			ServerURL: "https://keycloak.example.com",
			Realm:     "test-realm",
		}
		logger := zap.NewNop()
		auth := NewEnhancedKeycloakAuth(config, logger)

		// Test concurrent access to IsEnhancedEnabled
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func() {
				enabled := auth.IsEnhancedEnabled()
				assert.True(t, enabled)
				done <- true
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}
	})

	t.Run("handles context cancellation", func(t *testing.T) {
		config := &KeycloakConfig{
			ServerURL: "https://keycloak.example.com",
			Realm:     "test-realm",
		}
		logger := zap.NewNop()
		auth := NewKeycloakAuthCompatible(config, logger)

		// Create cancelled context
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		isValid, err := auth.AuthenticateBasic(ctx, "valid-token")

		// Should still work as basic auth doesn't use context
		assert.NoError(t, err)
		assert.True(t, isValid)
	})
}
