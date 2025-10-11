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
	t.Run("valid zitadel config", func(t *testing.T) {
		cfg := &config.Config{
			ZeroTrust: struct {
				Authentication struct {
					Enabled   bool   `json:"enabled"`
					Type      string `json:"type"`
					JWTSecret string `json:"jwt_secret,omitempty"`
				} `json:"authentication"`
			}{
				Authentication: struct {
					Enabled   bool   `json:"enabled"`
					Type      string `json:"type"`
					JWTSecret string `json:"jwt_secret,omitempty"`
				}{
					Enabled: true,
					Type:    "zitadel",
				},
			},
			Zitadel: struct {
				Enabled         bool   `json:"enabled"`
				Issuer          string `json:"issuer"`
				Audience        string `json:"audience"`
				TokenURL        string `json:"token_url"`
				ClientID        string `json:"client_id"`
				ClientSecret    string `json:"client_secret"`
				GRPCHost        string `json:"grpc_host"`
				FallbackSecret  string `json:"fallback_secret"`
				StrictMode      bool   `json:"strict_mode"`
				TokenValidation struct {
					Audience  string        `json:"audience"`
					ClockSkew time.Duration `json:"clock_skew"`
				} `json:"token_validation"`
				CacheTTL         time.Duration `json:"cache_ttl"`
				HTTPTimeout      time.Duration `json:"http_timeout"`
				TenantManagement struct {
					Enabled       bool `json:"enabled"`
					AutoCreate    bool `json:"auto_create"`
					DefaultQuotas struct {
						MaxPeers       int    `json:"max_peers"`
						MaxConnections int    `json:"max_connections"`
						MaxTunnels     int    `json:"max_tunnels"`
						BandwidthLimit string `json:"bandwidth_limit"`
						StorageLimit   string `json:"storage_limit"`
					} `json:"default_quotas"`
				} `json:"tenant_management"`
			}{
				Enabled:        true,
				Issuer:         "https://zitadel.example.com",
				Audience:       "test-project",
				TokenURL:       "https://zitadel.example.com/oauth/v2/token",
				ClientID:       "test-client",
				ClientSecret:   "test-secret",
				GRPCHost:       "zitadel.example.com:443",
				FallbackSecret: "test-fallback-secret",
				StrictMode:     false,
				TokenValidation: struct {
					Audience  string        `json:"audience"`
					ClockSkew time.Duration `json:"clock_skew"`
				}{
					Audience:  "test-audience",
					ClockSkew: 30 * time.Second,
				},
				CacheTTL:    5 * time.Minute,
				HTTPTimeout: 30 * time.Second,
				TenantManagement: struct {
					Enabled       bool `json:"enabled"`
					AutoCreate    bool `json:"auto_create"`
					DefaultQuotas struct {
						MaxPeers       int    `json:"max_peers"`
						MaxConnections int    `json:"max_connections"`
						MaxTunnels     int    `json:"max_tunnels"`
						BandwidthLimit string `json:"bandwidth_limit"`
						StorageLimit   string `json:"storage_limit"`
					} `json:"default_quotas"`
				}{
					Enabled:    true,
					AutoCreate: true,
					DefaultQuotas: struct {
						MaxPeers       int    `json:"max_peers"`
						MaxConnections int    `json:"max_connections"`
						MaxTunnels     int    `json:"max_tunnels"`
						BandwidthLimit string `json:"bandwidth_limit"`
						StorageLimit   string `json:"storage_limit"`
					}{
						MaxPeers:       10,
						MaxConnections: 100,
						MaxTunnels:     5,
						BandwidthLimit: "1Gbps",
						StorageLimit:   "10GB",
					},
				},
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
		assert.Equal(t, "zitadel", adapter.GetProvider())
	})

	t.Run("no providers enabled", func(t *testing.T) {
		cfg := &config.Config{
			ZeroTrust: struct {
				Authentication struct {
					Enabled   bool   `json:"enabled"`
					Type      string `json:"type"`
					JWTSecret string `json:"jwt_secret,omitempty"`
				} `json:"authentication"`
			}{
				Authentication: struct {
					Enabled   bool   `json:"enabled"`
					Type      string `json:"type"`
					JWTSecret string `json:"jwt_secret,omitempty"`
				}{
					Enabled: false, // Authentication disabled
				},
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
		assert.Equal(t, "noop", adapter.GetProvider())
	})

	t.Run("nil logger", func(t *testing.T) {
		cfg := &config.Config{
			ZeroTrust: struct {
				Authentication struct {
					Enabled   bool   `json:"enabled"`
					Type      string `json:"type"`
					JWTSecret string `json:"jwt_secret,omitempty"`
				} `json:"authentication"`
			}{
				Authentication: struct {
					Enabled   bool   `json:"enabled"`
					Type      string `json:"type"`
					JWTSecret string `json:"jwt_secret,omitempty"`
				}{
					Enabled: true,
					Type:    "zitadel",
				},
			},
			Zitadel: struct {
				Enabled         bool   `json:"enabled"`
				Issuer          string `json:"issuer"`
				Audience        string `json:"audience"`
				TokenURL        string `json:"token_url"`
				ClientID        string `json:"client_id"`
				ClientSecret    string `json:"client_secret"`
				GRPCHost        string `json:"grpc_host"`
				FallbackSecret  string `json:"fallback_secret"`
				StrictMode      bool   `json:"strict_mode"`
				TokenValidation struct {
					Audience  string        `json:"audience"`
					ClockSkew time.Duration `json:"clock_skew"`
				} `json:"token_validation"`
				CacheTTL         time.Duration `json:"cache_ttl"`
				HTTPTimeout      time.Duration `json:"http_timeout"`
				TenantManagement struct {
					Enabled       bool `json:"enabled"`
					AutoCreate    bool `json:"auto_create"`
					DefaultQuotas struct {
						MaxPeers       int    `json:"max_peers"`
						MaxConnections int    `json:"max_connections"`
						MaxTunnels     int    `json:"max_tunnels"`
						BandwidthLimit string `json:"bandwidth_limit"`
						StorageLimit   string `json:"storage_limit"`
					} `json:"default_quotas"`
				} `json:"tenant_management"`
			}{
				Enabled: true,
			},
		}
		adapter, err := NewAdapter(cfg, nil)
		assert.Error(t, err)
		assert.Nil(t, adapter)
		assert.Contains(t, err.Error(), "logger is required")
	})
}

func TestAdapter_ValidateToken(t *testing.T) {
	t.Run("zitadel provider", func(t *testing.T) {
		cfg := &config.Config{
			ZeroTrust: struct {
				Authentication struct {
					Enabled   bool   `json:"enabled"`
					Type      string `json:"type"`
					JWTSecret string `json:"jwt_secret,omitempty"`
				} `json:"authentication"`
			}{
				Authentication: struct {
					Enabled   bool   `json:"enabled"`
					Type      string `json:"type"`
					JWTSecret string `json:"jwt_secret,omitempty"`
				}{
					Enabled: true,
					Type:    "zitadel",
				},
			},
			Zitadel: struct {
				Enabled         bool   `json:"enabled"`
				Issuer          string `json:"issuer"`
				Audience        string `json:"audience"`
				TokenURL        string `json:"token_url"`
				ClientID        string `json:"client_id"`
				ClientSecret    string `json:"client_secret"`
				GRPCHost        string `json:"grpc_host"`
				FallbackSecret  string `json:"fallback_secret"`
				StrictMode      bool   `json:"strict_mode"`
				TokenValidation struct {
					Audience  string        `json:"audience"`
					ClockSkew time.Duration `json:"clock_skew"`
				} `json:"token_validation"`
				CacheTTL         time.Duration `json:"cache_ttl"`
				HTTPTimeout      time.Duration `json:"http_timeout"`
				TenantManagement struct {
					Enabled       bool `json:"enabled"`
					AutoCreate    bool `json:"auto_create"`
					DefaultQuotas struct {
						MaxPeers       int    `json:"max_peers"`
						MaxConnections int    `json:"max_connections"`
						MaxTunnels     int    `json:"max_tunnels"`
						BandwidthLimit string `json:"bandwidth_limit"`
						StorageLimit   string `json:"storage_limit"`
					} `json:"default_quotas"`
				} `json:"tenant_management"`
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

		// Test with invalid token
		ctx := context.Background()
		claims, err := adapter.ValidateToken(ctx, "invalid-token")
		assert.Error(t, err)
		assert.Nil(t, claims)
	})
}

func TestAdapter_ContextCancellation(t *testing.T) {
	cfg := &config.Config{
		ZeroTrust: struct {
			Authentication struct {
				Enabled   bool   `json:"enabled"`
				Type      string `json:"type"`
				JWTSecret string `json:"jwt_secret,omitempty"`
			} `json:"authentication"`
		}{
			Authentication: struct {
				Enabled   bool   `json:"enabled"`
				Type      string `json:"type"`
				JWTSecret string `json:"jwt_secret,omitempty"`
			}{
				Enabled: true,
				Type:    "zitadel",
			},
		},
		Zitadel: struct {
			Enabled         bool   `json:"enabled"`
			Issuer          string `json:"issuer"`
			Audience        string `json:"audience"`
			TokenURL        string `json:"token_url"`
			ClientID        string `json:"client_id"`
			ClientSecret    string `json:"client_secret"`
			GRPCHost        string `json:"grpc_host"`
			FallbackSecret  string `json:"fallback_secret"`
			StrictMode      bool   `json:"strict_mode"`
			TokenValidation struct {
				Audience  string        `json:"audience"`
				ClockSkew time.Duration `json:"clock_skew"`
			} `json:"token_validation"`
			CacheTTL         time.Duration `json:"cache_ttl"`
			HTTPTimeout      time.Duration `json:"http_timeout"`
			TenantManagement struct {
				Enabled       bool `json:"enabled"`
				AutoCreate    bool `json:"auto_create"`
				DefaultQuotas struct {
					MaxPeers       int    `json:"max_peers"`
					MaxConnections int    `json:"max_connections"`
					MaxTunnels     int    `json:"max_tunnels"`
					BandwidthLimit string `json:"bandwidth_limit"`
					StorageLimit   string `json:"storage_limit"`
				} `json:"default_quotas"`
			} `json:"tenant_management"`
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

	// Test context cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	claims, err := adapter.ValidateToken(ctx, "test-token")
	assert.Error(t, err)
	assert.Nil(t, claims)
}


