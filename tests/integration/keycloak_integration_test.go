package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twogc/cloudbridge-relay-installer/config"
	"github.com/twogc/cloudbridge-relay-installer/internal/auth"
	"github.com/twogc/cloudbridge-relay-installer/internal/keycloak"
	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
)

// TestKeycloakIntegration tests the integration with Keycloak
func TestKeycloakIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup test configuration
	cfg := &config.Config{}
	cfg.Keycloak.Enabled = true
	cfg.Keycloak.ServerURL = "http://localhost:8080"
	cfg.Keycloak.Realm = "cloudbridge"
	cfg.Keycloak.ClientID = "relay-client"
	cfg.Keycloak.ClientSecret = "relay-client-secret"
	cfg.Keycloak.TokenValidation.Issuer = "http://localhost:8080/realms/cloudbridge"
	cfg.Keycloak.TokenValidation.Audience = "relay-client"
	cfg.Keycloak.TokenValidation.ClockSkew = 30 * time.Second

	// Create logger
	logger := logger.NewLogger(logger.Config{
		Level:   logger.DEBUG,
		Console: true,
		Format:  "json",
	})

	// Test Keycloak client creation
	t.Run("CreateKeycloakClient", func(t *testing.T) {
		keycloakConfig := &keycloak.Config{
			ServerURL:    cfg.Keycloak.ServerURL,
			Realm:        cfg.Keycloak.Realm,
			ClientID:     cfg.Keycloak.ClientID,
			ClientSecret: cfg.Keycloak.ClientSecret,
			TokenValidation: struct {
				Issuer    string        `json:"issuer"`
				Audience  string        `json:"audience"`
				ClockSkew time.Duration `json:"clock_skew"`
			}{
				Issuer:    cfg.Keycloak.TokenValidation.Issuer,
				Audience:  cfg.Keycloak.TokenValidation.Audience,
				ClockSkew: cfg.Keycloak.TokenValidation.ClockSkew,
			},
			Logger: logger,
		}

		client, err := keycloak.NewClient(keycloakConfig)
		require.NoError(t, err)
		assert.NotNil(t, client)
	})

	// Test authentication adapter
	t.Run("CreateAuthAdapter", func(t *testing.T) {
		adapter, err := auth.NewAdapter(cfg, logger)
		require.NoError(t, err)
		assert.NotNil(t, adapter)
		assert.Equal(t, "keycloak", adapter.GetProvider())
	})

	// Test token validation (requires running Keycloak)
	t.Run("ValidateToken", func(t *testing.T) {
		// Skip if Keycloak is not running
		if !isKeycloakRunning() {
			t.Skip("Keycloak is not running, skipping token validation test")
		}

		adapter, err := auth.NewAdapter(cfg, logger)
		require.NoError(t, err)

		// Get a valid token from Keycloak
		token, err := getKeycloakToken(cfg)
		require.NoError(t, err)

		// Validate token
		tokenInfo, err := adapter.ValidateToken(context.Background(), token)
		require.NoError(t, err)
		assert.NotNil(t, tokenInfo)
		assert.Equal(t, "keycloak", tokenInfo.Provider)
	})

	// Test user info extraction
	t.Run("GetUserInfo", func(t *testing.T) {
		if !isKeycloakRunning() {
			t.Skip("Keycloak is not running, skipping user info test")
		}

		adapter, err := auth.NewAdapter(cfg, logger)
		require.NoError(t, err)

		token, err := getKeycloakToken(cfg)
		require.NoError(t, err)

		userInfo, err := adapter.GetUserInfo(context.Background(), token)
		require.NoError(t, err)
		assert.NotNil(t, userInfo)
		assert.Equal(t, "keycloak", userInfo.Provider)
		assert.True(t, userInfo.Active)
	})

	// Test permission checking
	t.Run("CheckPermission", func(t *testing.T) {
		if !isKeycloakRunning() {
			t.Skip("Keycloak is not running, skipping permission test")
		}

		adapter, err := auth.NewAdapter(cfg, logger)
		require.NoError(t, err)

		token, err := getKeycloakToken(cfg)
		require.NoError(t, err)

		// Test admin permission
		hasPermission, err := adapter.CheckPermission(context.Background(), token, "tunnel", "create")
		require.NoError(t, err)
		assert.True(t, hasPermission)
	})

	// Test invalid token
	t.Run("InvalidToken", func(t *testing.T) {
		adapter, err := auth.NewAdapter(cfg, logger)
		require.NoError(t, err)

		_, err = adapter.ValidateToken(context.Background(), "invalid-token")
		assert.Error(t, err)
	})

	// Test expired token
	t.Run("ExpiredToken", func(t *testing.T) {
		if !isKeycloakRunning() {
			t.Skip("Keycloak is not running, skipping expired token test")
		}

		adapter, err := auth.NewAdapter(cfg, logger)
		require.NoError(t, err)

		// Create an expired token (this would require a custom implementation)
		// For now, we'll test with an invalid token
		_, err = adapter.ValidateToken(context.Background(), "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJleHAiOjE1MTYyMzkwMjJ9.invalid")
		assert.Error(t, err)
	})
}

// TestKeycloakJWKS tests JWKS (JSON Web Key Set) functionality
func TestKeycloakJWKS(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if !isKeycloakRunning() {
		t.Skip("Keycloak is not running, skipping JWKS test")
	}

	cfg := &config.Config{}
	cfg.Keycloak.Enabled = true
	cfg.Keycloak.ServerURL = "http://localhost:8080"
	cfg.Keycloak.Realm = "cloudbridge"

	logger := logger.NewLogger(logger.Config{
		Level:   logger.DEBUG,
		Console: true,
		Format:  "json",
	})

	keycloakConfig := &keycloak.Config{
		ServerURL: cfg.Keycloak.ServerURL,
		Realm:     cfg.Keycloak.Realm,
		Logger:    logger,
	}

	client, err := keycloak.NewClient(keycloakConfig)
	require.NoError(t, err)

	// Test JWKS fetching
	t.Run("FetchJWKS", func(t *testing.T) {
		ctx := context.Background()
		jwks, err := client.(*keycloak.Client).FetchJWKS(ctx)
		require.NoError(t, err)
		assert.NotNil(t, jwks)
		assert.NotEmpty(t, jwks.Keys)
	})

	// Test public key retrieval
	t.Run("GetPublicKey", func(t *testing.T) {
		ctx := context.Background()
		
		// First, get JWKS to find a key ID
		jwks, err := client.(*keycloak.Client).FetchJWKS(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, jwks.Keys)

		kid := jwks.Keys[0].Kid
		assert.NotEmpty(t, kid)

		// Get public key for this key ID
		publicKey, err := client.(*keycloak.Client).GetPublicKey(ctx, kid)
		require.NoError(t, err)
		assert.NotNil(t, publicKey)
	})
}

// TestKeycloakClaims tests JWT claims validation
func TestKeycloakClaims(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := &config.Config{}
	cfg.Keycloak.Enabled = true
	cfg.Keycloak.TokenValidation.Issuer = "http://localhost:8080/realms/cloudbridge"
	cfg.Keycloak.TokenValidation.Audience = "relay-client"
	cfg.Keycloak.TokenValidation.ClockSkew = 30 * time.Second

	logger := logger.NewLogger(logger.Config{
		Level:   logger.DEBUG,
		Console: true,
		Format:  "json",
	})

	keycloakConfig := &keycloak.Config{
		TokenValidation: struct {
			Issuer    string        `json:"issuer"`
			Audience  string        `json:"audience"`
			ClockSkew time.Duration `json:"clock_skew"`
		}{
			Issuer:    cfg.Keycloak.TokenValidation.Issuer,
			Audience:  cfg.Keycloak.TokenValidation.Audience,
			ClockSkew: cfg.Keycloak.TokenValidation.ClockSkew,
		},
		Logger: logger,
	}

	client, err := keycloak.NewClient(keycloakConfig)
	require.NoError(t, err)

	// Test valid claims
	t.Run("ValidClaims", func(t *testing.T) {
		claims := map[string]interface{}{
			"iss": "http://localhost:8080/realms/cloudbridge",
			"aud": "relay-client",
			"exp": float64(time.Now().Add(time.Hour).Unix()),
			"iat": float64(time.Now().Unix()),
			"nbf": float64(time.Now().Unix()),
		}

		err := client.(*keycloak.Client).ValidateClaims(claims)
		assert.NoError(t, err)
	})

	// Test invalid issuer
	t.Run("InvalidIssuer", func(t *testing.T) {
		claims := map[string]interface{}{
			"iss": "http://invalid-issuer.com",
			"aud": "relay-client",
			"exp": float64(time.Now().Add(time.Hour).Unix()),
		}

		err := client.(*keycloak.Client).ValidateClaims(claims)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid issuer")
	})

	// Test invalid audience
	t.Run("InvalidAudience", func(t *testing.T) {
		claims := map[string]interface{}{
			"iss": "http://localhost:8080/realms/cloudbridge",
			"aud": "invalid-audience",
			"exp": float64(time.Now().Add(time.Hour).Unix()),
		}

		err := client.(*keycloak.Client).ValidateClaims(claims)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid audience")
	})

	// Test expired token
	t.Run("ExpiredToken", func(t *testing.T) {
		claims := map[string]interface{}{
			"iss": "http://localhost:8080/realms/cloudbridge",
			"aud": "relay-client",
			"exp": float64(time.Now().Add(-time.Hour).Unix()), // Expired 1 hour ago
		}

		err := client.(*keycloak.Client).ValidateClaims(claims)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token expired")
	})

	// Test token not yet valid
	t.Run("TokenNotYetValid", func(t *testing.T) {
		claims := map[string]interface{}{
			"iss": "http://localhost:8080/realms/cloudbridge",
			"aud": "relay-client",
			"nbf": float64(time.Now().Add(time.Hour).Unix()), // Valid in 1 hour
		}

		err := client.(*keycloak.Client).ValidateClaims(claims)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token not yet valid")
	})
}

// Helper functions

// isKeycloakRunning checks if Keycloak is running
func isKeycloakRunning() bool {
	resp, err := http.Get("http://localhost:8080/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// getKeycloakToken gets a valid token from Keycloak
func getKeycloakToken(cfg *config.Config) (string, error) {
	// This is a simplified version - in a real test, you'd use the Keycloak API
	// to get a proper token. For now, we'll return a placeholder.
	
	// In a real implementation, you would:
	// 1. Make a POST request to Keycloak token endpoint
	// 2. Use client credentials or resource owner password flow
	// 3. Return the access token
	
	return "placeholder-token", fmt.Errorf("not implemented - requires running Keycloak instance")
}

// Mock functions for testing (these would be implemented in the actual keycloak package)

func (c *keycloak.Client) FetchJWKS(ctx context.Context) (*keycloak.JWKS, error) {
	// This would be implemented in the actual keycloak package
	return nil, fmt.Errorf("not implemented")
}

func (c *keycloak.Client) GetPublicKey(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	// This would be implemented in the actual keycloak package
	return nil, fmt.Errorf("not implemented")
}

func (c *keycloak.Client) ValidateClaims(claims map[string]interface{}) error {
	// This would be implemented in the actual keycloak package
	return fmt.Errorf("not implemented")
} 