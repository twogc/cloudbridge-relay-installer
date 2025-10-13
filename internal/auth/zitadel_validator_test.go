package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestZitadelValidator_ValidateRelayToken(t *testing.T) {
	// Создаем тестовые ключи
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	publicKey := &privateKey.PublicKey

	// Создаем валидатор
	validator := NewZitadelValidator(ZitadelValidatorConfig{
		PublicKey: publicKey,
		Issuer:    "https://edge.2gc.ru/oauth/v2",
		Audience:  "cloudbridge-relay",
		KeyID:     "test-key-123",
		ClockSkew: 5 * time.Minute,
		Logger:    zaptest.NewLogger(t),
	})

	t.Run("ValidPermanentToken", func(t *testing.T) {
		// Создаем валидный токен
		now := time.Now()
		claims := &RelayClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    "https://edge.2gc.ru/oauth/v2",
				Subject:   "relay-server-12345",
				Audience:  []string{"cloudbridge-relay"},
				IssuedAt:  jwt.NewNumericDate(now),
				NotBefore: jwt.NewNumericDate(now),
				// Нет ExpiresAt для постоянного токена
			},
			OrgID:         "org-123",
			TenantID:      "tenant-456",
			ServiceName:   "cloudbridge-service",
			Permissions:   []string{"p2p_connect", "mesh_join", "relay_access"},
			Roles:         []string{"service", "relay-service"},
			RelayServerID: "relay-server-12345",
			ProtocolType:  "quick",
			TokenType:     "relay_permanent",
		}

		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		token.Header["kid"] = "test-key-123"

		tokenString, err := token.SignedString(privateKey)
		require.NoError(t, err)

		// Валидируем токен
		validatedClaims, err := validator.ValidateRelayToken(context.Background(), tokenString)
		require.NoError(t, err)
		require.NotNil(t, validatedClaims)

		assert.Equal(t, "relay-server-12345", validatedClaims.GetRelayServerID())
		assert.Equal(t, "tenant-456", validatedClaims.GetTenantID())
		assert.Equal(t, "org-123", validatedClaims.GetOrgID())
		assert.True(t, validatedClaims.IsRelayToken())
		assert.True(t, validatedClaims.IsPermanentToken())
	})

	t.Run("ValidSessionToken", func(t *testing.T) {
		// Создаем валидный сессионный токен
		now := time.Now()
		expTime := now.Add(time.Hour)

		claims := &RelayClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    "https://edge.2gc.ru/oauth/v2",
				Subject:   "relay-server-67890",
				Audience:  []string{"cloudbridge-relay"},
				IssuedAt:  jwt.NewNumericDate(now),
				NotBefore: jwt.NewNumericDate(now),
				ExpiresAt: jwt.NewNumericDate(expTime),
			},
			OrgID:         "org-789",
			TenantID:      "tenant-012",
			ServiceName:   "cloudbridge-service",
			Permissions:   []string{"p2p_connect", "mesh_join", "relay_access"},
			Roles:         []string{"service", "relay-service"},
			RelayServerID: "relay-server-67890",
			ProtocolType:  "quick",
			TokenType:     "relay_session",
		}

		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		token.Header["kid"] = "test-key-123"

		tokenString, err := token.SignedString(privateKey)
		require.NoError(t, err)

		// Валидируем токен
		validatedClaims, err := validator.ValidateRelayToken(context.Background(), tokenString)
		require.NoError(t, err)
		require.NotNil(t, validatedClaims)

		assert.Equal(t, "relay-server-67890", validatedClaims.GetRelayServerID())
		assert.Equal(t, "tenant-012", validatedClaims.GetTenantID())
		assert.Equal(t, "org-789", validatedClaims.GetOrgID())
		assert.True(t, validatedClaims.IsRelayToken())
		assert.False(t, validatedClaims.IsPermanentToken())
	})

	t.Run("InvalidIssuer", func(t *testing.T) {
		claims := &RelayClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:   "https://wrong-issuer.com",
				Subject:  "relay-server-12345",
				Audience: []string{"cloudbridge-relay"},
			},
			TokenType: "relay_permanent",
		}

		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		token.Header["kid"] = "test-key-123"

		tokenString, err := token.SignedString(privateKey)
		require.NoError(t, err)

		_, err = validator.ValidateRelayToken(context.Background(), tokenString)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid issuer")
	})

	t.Run("InvalidAudience", func(t *testing.T) {
		claims := &RelayClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:   "https://edge.2gc.ru/oauth/v2",
				Subject:  "relay-server-12345",
				Audience: []string{"wrong-audience"},
			},
			TokenType: "relay_permanent",
		}

		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		token.Header["kid"] = "test-key-123"

		tokenString, err := token.SignedString(privateKey)
		require.NoError(t, err)

		_, err = validator.ValidateRelayToken(context.Background(), tokenString)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid audience")
	})

	t.Run("InvalidTokenType", func(t *testing.T) {
		claims := &RelayClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:   "https://edge.2gc.ru/oauth/v2",
				Subject:  "relay-server-12345",
				Audience: []string{"cloudbridge-relay"},
			},
			TokenType: "user_session", // Неправильный тип токена
		}

		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		token.Header["kid"] = "test-key-123"

		tokenString, err := token.SignedString(privateKey)
		require.NoError(t, err)

		_, err = validator.ValidateRelayToken(context.Background(), tokenString)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid token type")
	})

	t.Run("ExpiredToken", func(t *testing.T) {
		now := time.Now()
		expTime := now.Add(-time.Hour) // Токен истек час назад

		claims := &RelayClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    "https://edge.2gc.ru/oauth/v2",
				Subject:   "relay-server-12345",
				Audience:  []string{"cloudbridge-relay"},
				IssuedAt:  jwt.NewNumericDate(now.Add(-2 * time.Hour)),
				NotBefore: jwt.NewNumericDate(now.Add(-2 * time.Hour)),
				ExpiresAt: jwt.NewNumericDate(expTime),
			},
			TokenType: "relay_session",
		}

		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		token.Header["kid"] = "test-key-123"

		tokenString, err := token.SignedString(privateKey)
		require.NoError(t, err)

		_, err = validator.ValidateRelayToken(context.Background(), tokenString)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token is expired")
	})
}

func TestZitadelValidator_ValidatePermissions(t *testing.T) {
	validator := NewZitadelValidator(ZitadelValidatorConfig{
		Logger: zaptest.NewLogger(t),
	})

	claims := &RelayClaims{
		Permissions: []string{"p2p_connect", "mesh_join", "relay_access"},
	}

	t.Run("ValidPermissions", func(t *testing.T) {
		requiredPermissions := []string{"p2p_connect", "mesh_join"}
		err := validator.ValidatePermissions(claims, requiredPermissions)
		assert.NoError(t, err)
	})

	t.Run("InvalidPermissions", func(t *testing.T) {
		requiredPermissions := []string{"p2p_connect", "admin_access"}
		err := validator.ValidatePermissions(claims, requiredPermissions)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient permissions")
	})

	t.Run("NilClaims", func(t *testing.T) {
		err := validator.ValidatePermissions(nil, []string{"p2p_connect"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "claims is nil")
	})
}

func TestZitadelValidator_ValidateRoles(t *testing.T) {
	validator := NewZitadelValidator(ZitadelValidatorConfig{
		Logger: zaptest.NewLogger(t),
	})

	claims := &RelayClaims{
		Roles: []string{"service", "relay-service"},
	}

	t.Run("ValidRoles", func(t *testing.T) {
		requiredRoles := []string{"service", "relay-service"}
		err := validator.ValidateRoles(claims, requiredRoles)
		assert.NoError(t, err)
	})

	t.Run("InvalidRoles", func(t *testing.T) {
		requiredRoles := []string{"admin", "system-admin"}
		err := validator.ValidateRoles(claims, requiredRoles)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient roles")
	})

	t.Run("NilClaims", func(t *testing.T) {
		err := validator.ValidateRoles(nil, []string{"service"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "claims is nil")
	})
}

func TestZitadelValidator_GetTokenInfo(t *testing.T) {
	validator := NewZitadelValidator(ZitadelValidatorConfig{
		Logger: zaptest.NewLogger(t),
	})

	now := time.Now()
	claims := &RelayClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "relay-server-12345",
			Issuer:    "https://edge.2gc.ru/oauth/v2",
			Audience:  []string{"cloudbridge-relay"},
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
		},
		OrgID:       "org-123",
		TenantID:    "tenant-456",
		Permissions: []string{"p2p_connect", "mesh_join"},
		Roles:       []string{"service", "relay-service"},
		TokenType:   "relay_permanent",
	}

	info := validator.GetTokenInfo(claims)
	require.NotNil(t, info)

	assert.Equal(t, "relay-server-12345", info["subject"])
	assert.Equal(t, "https://edge.2gc.ru/oauth/v2", info["issuer"])
	assert.Equal(t, []string{"cloudbridge-relay"}, info["audience"])
	assert.Equal(t, "relay_permanent", info["token_type"])
	assert.Equal(t, "org-123", info["org_id"])
	assert.Equal(t, "tenant-456", info["tenant_id"])
	assert.Equal(t, []string{"p2p_connect", "mesh_join"}, info["permissions"])
	assert.Equal(t, []string{"service", "relay-service"}, info["roles"])
	assert.True(t, info["is_permanent"].(bool))

	// Тестируем nil claims
	infoNil := validator.GetTokenInfo(nil)
	assert.Nil(t, infoNil)
}
