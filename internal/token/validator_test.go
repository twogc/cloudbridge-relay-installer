package token

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/twogc/cloudbridge-relay-installer/internal/django"
)

// mockHTTPClient is a mock implementation of HTTPDoer
//nolint:unused // Used in tests
type mockHTTPClient struct {
	statusCode int
	err        error
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &http.Response{
		StatusCode: m.statusCode,
		Body:       http.NoBody,
	}, nil
}

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

func TestValidator_REST(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse JSON request body
		var req struct {
			Token string `json:"token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Check token
		if req.Token == "valid-token" {
			response := map[string]interface{}{
				"valid":     true,
				"client_id": "test-client",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		} else {
			response := map[string]interface{}{
				"valid": false,
				"error": "invalid token",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()

	validator := NewValidator(Config{
		DjangoClient: django.NewClient(django.Config{
			APIEndpoint: server.URL,
			Timeout:     time.Second,
		}),
	})

	if _, err := validator.ValidateToken("valid-token"); err != nil {
		t.Error("Expected valid token to be accepted")
	}
	if _, err := validator.ValidateToken("invalid-token"); err == nil {
		t.Error("Expected invalid token to be rejected")
	}
	if _, err := validator.ValidateToken("valid-token"); err != nil {
		t.Error("Expected cached token to be accepted")
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
	validator := NewValidator(Config{
		DjangoClient: django.NewClient(django.Config{
			APIEndpoint: "http://example.com",
			Timeout:     time.Second,
		}),
	})
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
