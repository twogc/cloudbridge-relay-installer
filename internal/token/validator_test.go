package token

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestValidator_Validate_EmptyToken(t *testing.T) {
	v := NewValidator(Config{})
	if _, err := v.ValidateToken(""); err == nil {
		t.Error("empty token should be invalid")
	}
}

func TestValidator_Validate_NoValidation(t *testing.T) {
	v := NewValidator(Config{})
	if _, err := v.ValidateToken("any-token"); err != nil {
		t.Error("token should be valid when no validation is configured")
	}
}

func TestValidator_JWT(t *testing.T) {
	secret := []byte("test-secret")
	validator := NewValidator(Config{
		JWTSecret: string(secret),
	})

	claims := jwt.MapClaims{
		"exp":       time.Now().Add(time.Hour).Unix(),
		"client_id": "test-client",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secret)
	if err != nil {
		t.Fatalf("Failed to create test token: %v", err)
	}

	if _, err := validator.ValidateToken(tokenString); err != nil {
		t.Error("Expected valid token to be accepted")
	}

	claims["exp"] = time.Now().Add(-time.Hour).Unix()
	claims["client_id"] = "test-client"
	token = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err = token.SignedString(secret)
	if err != nil {
		t.Fatalf("Failed to create test token: %v", err)
	}
	if _, err := validator.ValidateToken(tokenString); err == nil {
		t.Error("Expected expired token to be rejected")
	}

	if _, err := validator.ValidateToken(tokenString + "invalid"); err == nil {
		t.Error("Expected token with invalid signature to be rejected")
	}

	claims["exp"] = time.Now().Add(time.Hour).Unix()
	claims["client_id"] = "test-client"
	token = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err = token.SignedString(secret)
	if err != nil {
		t.Fatalf("Failed to create test token: %v", err)
	}
	if _, err := validator.ValidateToken(tokenString); err != nil {
		t.Error("Expected cached token to be accepted")
	}
}

func TestValidator_NoValidation(t *testing.T) {
	validator := NewValidator(Config{})
	if _, err := validator.ValidateToken("any-token"); err != nil {
		t.Error("Expected any token to be accepted when no validation is configured")
	}
}

func TestValidator_Cleanup(t *testing.T) {
	validator := NewValidator(Config{})
	validator.mu.Lock()
	validator.cache["token1"] = TokenInfo{ClientID: "client1", Expiry: time.Now().Add(time.Hour)}
	validator.cache["token2"] = TokenInfo{ClientID: "client2", Expiry: time.Now().Add(-time.Hour)}
	validator.mu.Unlock()

	validator.Cleanup()

	validator.mu.RLock()
	if _, ok := validator.cache["token1"]; !ok {
		t.Error("Expected valid token to remain in cache")
	}
	if _, ok := validator.cache["token2"]; ok {
		t.Error("Expected expired token to be removed from cache")
	}
	validator.mu.RUnlock()
}

func TestValidator_P2PToken(t *testing.T) {
	secret := []byte("test-secret")
	validator := NewValidator(Config{
		JWTSecret: string(secret),
	})

	// Test valid P2P token
	claims := &P2PClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
		ProtocolType:   "p2p-mesh",
		Scope:          "p2p",
		TenantID:       "tenant-123",
		ServerID:       "server-456",
		ConnectionType: "quic",
		Permissions:    []string{"p2p_connect", "p2p_relay"},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secret)
	if err != nil {
		t.Fatalf("Failed to create test token: %v", err)
	}

	validatedClaims, err := validator.ValidateP2PToken(tokenString)
	if err != nil {
		t.Errorf("Expected valid P2P token to be accepted, got error: %v", err)
	}
	if validatedClaims == nil {
		t.Error("Expected validated claims to be returned")
	}
	if validatedClaims.TenantID != "tenant-123" {
		t.Errorf("Expected tenant ID to be 'tenant-123', got '%s'", validatedClaims.TenantID)
	}

	// Test invalid protocol type
	claims.ProtocolType = "invalid-protocol"
	token = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err = token.SignedString(secret)
	if err != nil {
		t.Fatalf("Failed to create test token: %v", err)
	}

	_, err = validator.ValidateP2PToken(tokenString)
	if err == nil {
		t.Error("Expected invalid protocol type to be rejected")
	}

	// Test missing permission
	claims.ProtocolType = "p2p-mesh"
	claims.Permissions = []string{"other_permission"}
	token = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err = token.SignedString(secret)
	if err != nil {
		t.Fatalf("Failed to create test token: %v", err)
	}

	_, err = validator.ValidateP2PToken(tokenString)
	if err == nil {
		t.Error("Expected missing p2p_connect permission to be rejected")
	}
}
