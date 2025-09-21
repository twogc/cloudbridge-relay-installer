package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewAuthMiddleware(t *testing.T) {
	validator := &MockTokenValidator{}
	logger := zap.NewNop()

	middleware := NewAuthMiddleware(validator, logger)

	assert.NotNil(t, middleware)
}

func TestAuthMiddleware_RequireAuth(t *testing.T) {
	validator := &MockTokenValidator{}
	logger := zap.NewNop()

	middleware := NewAuthMiddleware(validator, logger)
	authFunc := middleware.RequireAuth()

	assert.NotNil(t, authFunc)

	// Test with valid token
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer valid-token")

	validator.On("ValidateP2PToken", "valid-token").Return(&P2PClaims{
		Subject:  "test-user",
		TenantID: "test-tenant",
	}, nil)

	authFunc(c)

	// Should not abort if token is valid
	assert.False(t, c.IsAborted())
}

func TestAuthMiddleware_RequireAuth_MissingToken(t *testing.T) {
	validator := &MockTokenValidator{}
	logger := zap.NewNop()

	middleware := NewAuthMiddleware(validator, logger)
	authFunc := middleware.RequireAuth()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	// No Authorization header

	authFunc(c)

	// Should abort with 401 if no token
	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_RequireAdmin(t *testing.T) {
	validator := &MockTokenValidator{}
	logger := zap.NewNop()

	middleware := NewAuthMiddleware(validator, logger)
	adminFunc := middleware.RequireAdmin()

	assert.NotNil(t, adminFunc)

	// Test with admin token
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/admin", nil)
	c.Request.Header.Set("Authorization", "Bearer admin-token")

	validator.On("ValidateP2PToken", "admin-token").Return(&P2PClaims{
		Subject:     "admin-user",
		TenantID:    "test-tenant",
		Permissions: []string{"admin"},
	}, nil)

	adminFunc(c)

	// Should not abort if user is admin
	assert.False(t, c.IsAborted())
}
