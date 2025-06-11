package token

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// mockHTTPClient is a mock implementation of HTTPDoer
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
	if v.Validate("") {
		t.Error("empty token should be invalid")
	}
}

func TestValidator_Validate_NoValidation(t *testing.T) {
	v := NewValidator(Config{})
	if !v.Validate("any-token") {
		t.Error("token should be valid when no validation is configured")
	}
}

func TestValidator_REST(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "Bearer valid-token" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	}))
	defer server.Close()

	validator := NewValidator(Config{
		RESTEndpoint: server.URL,
	})

	if !validator.Validate("valid-token") {
		t.Error("Expected valid token to be accepted")
	}
	if validator.Validate("invalid-token") {
		t.Error("Expected invalid token to be rejected")
	}
	if !validator.Validate("valid-token") {
		t.Error("Expected cached token to be accepted")
	}
}

func TestValidator_JWT(t *testing.T) {
	secret := []byte("test-secret")
	validator := NewValidator(Config{
		JWTSecret: string(secret),
	})

	claims := jwt.MapClaims{
		"exp": time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secret)
	if err != nil {
		t.Fatalf("Failed to create test token: %v", err)
	}

	if !validator.Validate(tokenString) {
		t.Error("Expected valid token to be accepted")
	}

	claims["exp"] = time.Now().Add(-time.Hour).Unix()
	token = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err = token.SignedString(secret)
	if err != nil {
		t.Fatalf("Failed to create test token: %v", err)
	}
	if validator.Validate(tokenString) {
		t.Error("Expected expired token to be rejected")
	}

	if validator.Validate(tokenString+"invalid") {
		t.Error("Expected token with invalid signature to be rejected")
	}

	claims["exp"] = time.Now().Add(time.Hour).Unix()
	token = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err = token.SignedString(secret)
	if err != nil {
		t.Fatalf("Failed to create test token: %v", err)
	}
	if !validator.Validate(tokenString) {
		t.Error("Expected cached token to be accepted")
	}
}

func TestValidator_NoValidation(t *testing.T) {
	validator := NewValidator(Config{})
	if !validator.Validate("any-token") {
		t.Error("Expected any token to be accepted when no validation is configured")
	}
}

func TestValidator_Cleanup(t *testing.T) {
	validator := NewValidator(Config{
		RESTEndpoint: "http://example.com",
	})
	validator.mu.Lock()
	validator.cache["token1"] = time.Now().Add(time.Hour)
	validator.cache["token2"] = time.Now().Add(-time.Hour)
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