package auth

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twogc/cloudbridge-relay-installer/config"
	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
)

func TestNewAdapter(t *testing.T) {
	t.Run("valid keycloak config", func(t *testing.T) {
		cfg := &config.Config{
			Keycloak: struct {
				Enabled         bool   `json:"enabled"`
				ServerURL       string `json:"server_url"`
				Realm           string `json:"realm"`
				ClientID        string `json:"client_id"`
				ClientSecret    string `json:"client_secret"`
				PublicKey       string `json:"public_key"`
				FallbackSecret  string `json:"fallback_secret"`
				JWKSURL         string `json:"jwks_url"`
				TokenValidation struct {
					Issuer    string        `json:"issuer"`
					Audience  string        `json:"audience"`
					ClockSkew time.Duration `json:"clock_skew"`
				} `json:"token_validation"`
				JWKSCacheTTL time.Duration `json:"jwks_cache_ttl"`
				HTTPTimeout  time.Duration `json:"http_timeout"`
			}{
				Enabled:   true,
				ServerURL: "https://keycloak.example.com",
				Realm:     "test-realm",
				ClientID:  "test-client",
			},
		}
		logger, err := logger.New(logger.Config{
			Level:  logger.INFO,
			Output: "stdout",
			JSON:   false,
		})
		require.NoError(t, err)
		adapter, err := NewAdapter(cfg, logger)
		require.NoError(t, err)
		assert.NotNil(t, adapter)
		assert.Equal(t, "keycloak", adapter.GetProvider())
	})

	t.Run("no providers enabled", func(t *testing.T) {
		cfg := &config.Config{}
		logger, err := logger.New(logger.Config{
			Level:  logger.INFO,
			Output: "stdout",
			JSON:   false,
		})
		require.NoError(t, err)
		adapter, err := NewAdapter(cfg, logger)
		assert.Error(t, err)
		assert.Nil(t, adapter)
		assert.Contains(t, err.Error(), "no authentication provider enabled")
	})


	t.Run("nil logger", func(t *testing.T) {
		cfg := &config.Config{
			Keycloak: struct {
				Enabled         bool   `json:"enabled"`
				ServerURL       string `json:"server_url"`
				Realm           string `json:"realm"`
				ClientID        string `json:"client_id"`
				ClientSecret    string `json:"client_secret"`
				PublicKey       string `json:"public_key"`
				FallbackSecret  string `json:"fallback_secret"`
				JWKSURL         string `json:"jwks_url"`
				TokenValidation struct {
					Issuer    string        `json:"issuer"`
					Audience  string        `json:"audience"`
					ClockSkew time.Duration `json:"clock_skew"`
				} `json:"token_validation"`
				JWKSCacheTTL time.Duration `json:"jwks_cache_ttl"`
				HTTPTimeout  time.Duration `json:"http_timeout"`
			}{
				Enabled: true,
			},
		}
		// NewAdapter doesn't validate nil logger, so it succeeds
		adapter, err := NewAdapter(cfg, nil)
		require.NoError(t, err)
		assert.NotNil(t, adapter)
		assert.Equal(t, "keycloak", adapter.GetProvider())
	})
}

func TestAdapter_ValidateToken(t *testing.T) {
	t.Run("keycloak provider", func(t *testing.T) {
		cfg := &config.Config{
			Keycloak: struct {
				Enabled         bool   `json:"enabled"`
				ServerURL       string `json:"server_url"`
				Realm           string `json:"realm"`
				ClientID        string `json:"client_id"`
				ClientSecret    string `json:"client_secret"`
				PublicKey       string `json:"public_key"`
				FallbackSecret  string `json:"fallback_secret"`
				JWKSURL         string `json:"jwks_url"`
				TokenValidation struct {
					Issuer    string        `json:"issuer"`
					Audience  string        `json:"audience"`
					ClockSkew time.Duration `json:"clock_skew"`
				} `json:"token_validation"`
				JWKSCacheTTL time.Duration `json:"jwks_cache_ttl"`
				HTTPTimeout  time.Duration `json:"http_timeout"`
			}{
				Enabled:   true,
				ServerURL: "https://keycloak.example.com",
				Realm:     "test-realm",
				ClientID:  "test-client",
			},
		}
		logger, err := logger.New(logger.Config{
			Level:  logger.INFO,
			Output: "stdout",
			JSON:   false,
		})
		require.NoError(t, err)
		adapter, err := NewAdapter(cfg, logger)
		require.NoError(t, err)

		ctx := context.Background()
		tokenInfo, err := adapter.ValidateToken(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, tokenInfo)

		tokenInfo, err = adapter.ValidateToken(ctx, "invalid-token")
		assert.Error(t, err)
		assert.Nil(t, tokenInfo)
	})

}

func TestAdapter_GetUserInfo(t *testing.T) {
	t.Run("keycloak provider", func(t *testing.T) {
		cfg := &config.Config{
			Keycloak: struct {
				Enabled         bool   `json:"enabled"`
				ServerURL       string `json:"server_url"`
				Realm           string `json:"realm"`
				ClientID        string `json:"client_id"`
				ClientSecret    string `json:"client_secret"`
				PublicKey       string `json:"public_key"`
				FallbackSecret  string `json:"fallback_secret"`
				JWKSURL         string `json:"jwks_url"`
				TokenValidation struct {
					Issuer    string        `json:"issuer"`
					Audience  string        `json:"audience"`
					ClockSkew time.Duration `json:"clock_skew"`
				} `json:"token_validation"`
				JWKSCacheTTL time.Duration `json:"jwks_cache_ttl"`
				HTTPTimeout  time.Duration `json:"http_timeout"`
			}{
				Enabled:   true,
				ServerURL: "https://keycloak.example.com",
				Realm:     "test-realm",
				ClientID:  "test-client",
			},
		}
		logger, err := logger.New(logger.Config{
			Level:  logger.INFO,
			Output: "stdout",
			JSON:   false,
		})
		require.NoError(t, err)
		adapter, err := NewAdapter(cfg, logger)
		require.NoError(t, err)

		ctx := context.Background()
		userInfo, err := adapter.GetUserInfo(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, userInfo)

		userInfo, err = adapter.GetUserInfo(ctx, "invalid-token")
		assert.Error(t, err)
		assert.Nil(t, userInfo)
	})

}

func TestAdapter_CheckPermission(t *testing.T) {
	t.Run("keycloak provider", func(t *testing.T) {
		cfg := &config.Config{
			Keycloak: struct {
				Enabled         bool   `json:"enabled"`
				ServerURL       string `json:"server_url"`
				Realm           string `json:"realm"`
				ClientID        string `json:"client_id"`
				ClientSecret    string `json:"client_secret"`
				PublicKey       string `json:"public_key"`
				FallbackSecret  string `json:"fallback_secret"`
				JWKSURL         string `json:"jwks_url"`
				TokenValidation struct {
					Issuer    string        `json:"issuer"`
					Audience  string        `json:"audience"`
					ClockSkew time.Duration `json:"clock_skew"`
				} `json:"token_validation"`
				JWKSCacheTTL time.Duration `json:"jwks_cache_ttl"`
				HTTPTimeout  time.Duration `json:"http_timeout"`
			}{
				Enabled:   true,
				ServerURL: "https://keycloak.example.com",
				Realm:     "test-realm",
				ClientID:  "test-client",
			},
		}
		logger, err := logger.New(logger.Config{
			Level:  logger.INFO,
			Output: "stdout",
			JSON:   false,
		})
		require.NoError(t, err)
		adapter, err := NewAdapter(cfg, logger)
		require.NoError(t, err)

		ctx := context.Background()
		hasPermission, err := adapter.CheckPermission(ctx, "", "resource", "action")
		assert.Error(t, err)
		assert.False(t, hasPermission)

		hasPermission, err = adapter.CheckPermission(ctx, "invalid-token", "resource", "action")
		assert.Error(t, err)
		assert.False(t, hasPermission)
	})

}

func TestAdapter_GetProvider(t *testing.T) {
	t.Run("keycloak provider", func(t *testing.T) {
		cfg := &config.Config{
			Keycloak: struct {
				Enabled         bool   `json:"enabled"`
				ServerURL       string `json:"server_url"`
				Realm           string `json:"realm"`
				ClientID        string `json:"client_id"`
				ClientSecret    string `json:"client_secret"`
				PublicKey       string `json:"public_key"`
				FallbackSecret  string `json:"fallback_secret"`
				JWKSURL         string `json:"jwks_url"`
				TokenValidation struct {
					Issuer    string        `json:"issuer"`
					Audience  string        `json:"audience"`
					ClockSkew time.Duration `json:"clock_skew"`
				} `json:"token_validation"`
				JWKSCacheTTL time.Duration `json:"jwks_cache_ttl"`
				HTTPTimeout  time.Duration `json:"http_timeout"`
			}{
				Enabled: true,
			},
		}
		logger, err := logger.New(logger.Config{
			Level:  logger.INFO,
			Output: "stdout",
			JSON:   false,
		})
		require.NoError(t, err)
		adapter, err := NewAdapter(cfg, logger)
		require.NoError(t, err)
		assert.Equal(t, "keycloak", adapter.GetProvider())
	})

}

func TestTokenInfo_Struct(t *testing.T) {
	t.Run("token info creation", func(t *testing.T) {
		now := time.Now()
		tokenInfo := &TokenInfo{
			Subject:   "user123",
			Issuer:    "https://keycloak.example.com",
			Audience:  []string{"client1", "client2"},
			ExpiresAt: now.Add(time.Hour),
			IssuedAt:  now,
			NotBefore: now,
			Username:  "testuser",
			Email:     "test@example.com",
			Roles:     []string{"user", "admin"},
			Claims:    map[string]interface{}{"custom": "value"},
		}

		assert.Equal(t, "user123", tokenInfo.Subject)
		assert.Equal(t, "https://keycloak.example.com", tokenInfo.Issuer)
		assert.Len(t, tokenInfo.Audience, 2)
		assert.Equal(t, "client1", tokenInfo.Audience[0])
		assert.Equal(t, "client2", tokenInfo.Audience[1])
		assert.True(t, tokenInfo.ExpiresAt.After(now))
		assert.Equal(t, now.Unix(), tokenInfo.IssuedAt.Unix())
		assert.Equal(t, now.Unix(), tokenInfo.NotBefore.Unix())
		assert.Equal(t, "testuser", tokenInfo.Username)
		assert.Equal(t, "test@example.com", tokenInfo.Email)
		assert.Len(t, tokenInfo.Roles, 2)
		assert.Equal(t, "user", tokenInfo.Roles[0])
		assert.Equal(t, "admin", tokenInfo.Roles[1])
		assert.Equal(t, "value", tokenInfo.Claims["custom"])
	})

	t.Run("empty token info", func(t *testing.T) {
		tokenInfo := &TokenInfo{}

		assert.Empty(t, tokenInfo.Subject)
		assert.Empty(t, tokenInfo.Issuer)
		assert.Nil(t, tokenInfo.Audience)
		assert.True(t, tokenInfo.ExpiresAt.IsZero())
		assert.True(t, tokenInfo.IssuedAt.IsZero())
		assert.True(t, tokenInfo.NotBefore.IsZero())
		assert.Empty(t, tokenInfo.Username)
		assert.Empty(t, tokenInfo.Email)
		assert.Nil(t, tokenInfo.Roles)
		assert.Nil(t, tokenInfo.Claims)
	})
}

func TestUserInfo_Struct(t *testing.T) {
	t.Run("user info creation", func(t *testing.T) {
		userInfo := &UserInfo{
			ID:       "user123",
			Username: "testuser",
			Email:    "test@example.com",
			FullName: "Test User",
			Roles:    []string{"user", "admin"},
			Groups:   []string{"group1", "group2"},
			Active:   true,
			Claims: map[string]interface{}{
				"department": "IT",
				"location":   "NYC",
			},
			Provider: "keycloak",
		}

		assert.Equal(t, "user123", userInfo.ID)
		assert.Equal(t, "testuser", userInfo.Username)
		assert.Equal(t, "test@example.com", userInfo.Email)
		assert.Equal(t, "Test User", userInfo.FullName)
		assert.Len(t, userInfo.Roles, 2)
		assert.Equal(t, "user", userInfo.Roles[0])
		assert.Equal(t, "admin", userInfo.Roles[1])
		assert.Len(t, userInfo.Groups, 2)
		assert.Equal(t, "group1", userInfo.Groups[0])
		assert.Equal(t, "group2", userInfo.Groups[1])
		assert.True(t, userInfo.Active)
		assert.Equal(t, "IT", userInfo.Claims["department"])
		assert.Equal(t, "NYC", userInfo.Claims["location"])
		assert.Equal(t, "keycloak", userInfo.Provider)
	})

	t.Run("empty user info", func(t *testing.T) {
		userInfo := &UserInfo{}

		assert.Empty(t, userInfo.ID)
		assert.Empty(t, userInfo.Username)
		assert.Empty(t, userInfo.Email)
		assert.Empty(t, userInfo.FullName)
		assert.Nil(t, userInfo.Roles)
		assert.Nil(t, userInfo.Groups)
		assert.False(t, userInfo.Active)
		assert.Nil(t, userInfo.Claims)
		assert.Empty(t, userInfo.Provider)
	})
}

func TestAdapter_ContextCancellation(t *testing.T) {
	cfg := &config.Config{
		Keycloak: struct {
			Enabled         bool   `json:"enabled"`
			ServerURL       string `json:"server_url"`
			Realm           string `json:"realm"`
			ClientID        string `json:"client_id"`
			ClientSecret    string `json:"client_secret"`
			PublicKey       string `json:"public_key"`
			FallbackSecret  string `json:"fallback_secret"`
			JWKSURL         string `json:"jwks_url"`
			TokenValidation struct {
				Issuer    string        `json:"issuer"`
				Audience  string        `json:"audience"`
				ClockSkew time.Duration `json:"clock_skew"`
			} `json:"token_validation"`
			JWKSCacheTTL time.Duration `json:"jwks_cache_ttl"`
			HTTPTimeout  time.Duration `json:"http_timeout"`
		}{
			Enabled: true,
		},
	}
	logger, err := logger.New(logger.Config{
		Level:  logger.INFO,
		Output: "stdout",
		JSON:   false,
	})
	require.NoError(t, err)
	adapter, err := NewAdapter(cfg, logger)
	require.NoError(t, err)

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Test with cancelled context
	tokenInfo, err := adapter.ValidateToken(ctx, "token")
	assert.Error(t, err)
	assert.Nil(t, tokenInfo)

	userInfo, err := adapter.GetUserInfo(ctx, "token")
	assert.Error(t, err)
	assert.Nil(t, userInfo)

	hasPermission, err := adapter.CheckPermission(ctx, "token", "resource", "action")
	assert.Error(t, err)
	assert.False(t, hasPermission)
}
