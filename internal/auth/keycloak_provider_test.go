package auth

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/twogc/cloudbridge-relay-installer/internal/keycloak"
)

// MockKeycloakClient is a mock implementation of KeycloakClientInterface
type MockKeycloakClient struct {
	mock.Mock
}

func (m *MockKeycloakClient) ValidateToken(ctx context.Context, token string) (*keycloak.TokenInfo, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*keycloak.TokenInfo), args.Error(1)
}

func (m *MockKeycloakClient) CheckPermission(tokenInfo *keycloak.TokenInfo, resource, action string) bool {
	args := m.Called(tokenInfo, resource, action)
	return args.Bool(0)
}

func TestKeycloakProvider_ValidateToken(t *testing.T) {
	t.Run("successful validation", func(t *testing.T) {
		mockClient := new(MockKeycloakClient)
		provider := &KeycloakProvider{client: mockClient}

		expectedTokenInfo := &keycloak.TokenInfo{
			Subject:   "user123",
			Issuer:    "https://keycloak.example.com",
			Audience:  []string{"client1"},
			ExpiresAt: time.Now().Add(time.Hour),
			IssuedAt:  time.Now(),
			NotBefore: time.Now(),
			Username:  "testuser",
			Email:     "test@example.com",
			FullName:  "Test User",
			Groups:    []string{"group1"},
			Roles:     []string{"user"},
			Claims:    map[string]interface{}{"custom": "value"},
		}

		mockClient.On("ValidateToken", mock.Anything, "valid-token").Return(expectedTokenInfo, nil)

		ctx := context.Background()
		tokenInfo, err := provider.ValidateToken(ctx, "valid-token")

		assert.NoError(t, err)
		assert.NotNil(t, tokenInfo)
		assert.Equal(t, "user123", tokenInfo.Subject)
		assert.Equal(t, "https://keycloak.example.com", tokenInfo.Issuer)
		assert.Equal(t, "testuser", tokenInfo.Username)
		assert.Equal(t, "test@example.com", tokenInfo.Email)
		assert.Equal(t, "Test User", tokenInfo.FullName)
		assert.Equal(t, "keycloak", tokenInfo.Provider)
		assert.Len(t, tokenInfo.Audience, 1)
		assert.Equal(t, "client1", tokenInfo.Audience[0])
		assert.Len(t, tokenInfo.Groups, 1)
		assert.Equal(t, "group1", tokenInfo.Groups[0])
		assert.Len(t, tokenInfo.Roles, 1)
		assert.Equal(t, "user", tokenInfo.Roles[0])
		assert.Equal(t, "value", tokenInfo.Claims["custom"])

		mockClient.AssertExpectations(t)
	})

	t.Run("validation error", func(t *testing.T) {
		mockClient := new(MockKeycloakClient)
		provider := &KeycloakProvider{client: mockClient}

		mockClient.On("ValidateToken", mock.Anything, "invalid-token").Return(nil, assert.AnError)

		ctx := context.Background()
		tokenInfo, err := provider.ValidateToken(ctx, "invalid-token")

		assert.Error(t, err)
		assert.Nil(t, tokenInfo)

		mockClient.AssertExpectations(t)
	})

	t.Run("empty token", func(t *testing.T) {
		mockClient := new(MockKeycloakClient)
		provider := &KeycloakProvider{client: mockClient}

		mockClient.On("ValidateToken", mock.Anything, "").Return(nil, assert.AnError)

		ctx := context.Background()
		tokenInfo, err := provider.ValidateToken(ctx, "")

		assert.Error(t, err)
		assert.Nil(t, tokenInfo)

		mockClient.AssertExpectations(t)
	})

	t.Run("context cancellation", func(t *testing.T) {
		mockClient := new(MockKeycloakClient)
		provider := &KeycloakProvider{client: mockClient}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		mockClient.On("ValidateToken", mock.Anything, "token").Return(nil, context.Canceled)

		tokenInfo, err := provider.ValidateToken(ctx, "token")

		assert.Error(t, err)
		assert.Nil(t, tokenInfo)

		mockClient.AssertExpectations(t)
	})
}

func TestKeycloakProvider_GetUserInfo(t *testing.T) {
	t.Run("successful user info retrieval", func(t *testing.T) {
		mockClient := new(MockKeycloakClient)
		provider := &KeycloakProvider{client: mockClient}

		expectedTokenInfo := &keycloak.TokenInfo{
			Subject:   "user123",
			Issuer:    "https://keycloak.example.com",
			Audience:  []string{"client1"},
			ExpiresAt: time.Now().Add(time.Hour),
			IssuedAt:  time.Now(),
			NotBefore: time.Now(),
			Username:  "testuser",
			Email:     "test@example.com",
			FullName:  "Test User",
			Groups:    []string{"group1", "group2"},
			Roles:     []string{"user", "admin"},
			Claims:    map[string]interface{}{"department": "IT"},
		}

		mockClient.On("ValidateToken", mock.Anything, "valid-token").Return(expectedTokenInfo, nil)

		ctx := context.Background()
		userInfo, err := provider.GetUserInfo(ctx, "valid-token")

		assert.NoError(t, err)
		assert.NotNil(t, userInfo)
		assert.Equal(t, "user123", userInfo.ID)
		assert.Equal(t, "testuser", userInfo.Username)
		assert.Equal(t, "test@example.com", userInfo.Email)
		assert.Equal(t, "Test User", userInfo.FullName)
		assert.True(t, userInfo.Active) // Keycloak tokens are only issued for active users
		assert.Equal(t, "keycloak", userInfo.Provider)
		assert.Len(t, userInfo.Groups, 2)
		assert.Equal(t, "group1", userInfo.Groups[0])
		assert.Equal(t, "group2", userInfo.Groups[1])
		assert.Len(t, userInfo.Roles, 2)
		assert.Equal(t, "user", userInfo.Roles[0])
		assert.Equal(t, "admin", userInfo.Roles[1])
		assert.Equal(t, "IT", userInfo.Claims["department"])

		mockClient.AssertExpectations(t)
	})

	t.Run("token validation error", func(t *testing.T) {
		mockClient := new(MockKeycloakClient)
		provider := &KeycloakProvider{client: mockClient}

		mockClient.On("ValidateToken", mock.Anything, "invalid-token").Return(nil, assert.AnError)

		ctx := context.Background()
		userInfo, err := provider.GetUserInfo(ctx, "invalid-token")

		assert.Error(t, err)
		assert.Nil(t, userInfo)

		mockClient.AssertExpectations(t)
	})

	t.Run("empty token", func(t *testing.T) {
		mockClient := new(MockKeycloakClient)
		provider := &KeycloakProvider{client: mockClient}

		mockClient.On("ValidateToken", mock.Anything, "").Return(nil, assert.AnError)

		ctx := context.Background()
		userInfo, err := provider.GetUserInfo(ctx, "")

		assert.Error(t, err)
		assert.Nil(t, userInfo)

		mockClient.AssertExpectations(t)
	})
}

func TestKeycloakProvider_CheckPermission(t *testing.T) {
	t.Run("permission granted", func(t *testing.T) {
		mockClient := new(MockKeycloakClient)
		provider := &KeycloakProvider{client: mockClient}

		expectedTokenInfo := &keycloak.TokenInfo{
			Subject:  "user123",
			Username: "testuser",
			Roles:    []string{"admin"},
		}

		mockClient.On("ValidateToken", mock.Anything, "valid-token").Return(expectedTokenInfo, nil)
		mockClient.On("CheckPermission", expectedTokenInfo, "resource", "action").Return(true)

		ctx := context.Background()
		hasPermission, err := provider.CheckPermission(ctx, "valid-token", "resource", "action")

		assert.NoError(t, err)
		assert.True(t, hasPermission)

		mockClient.AssertExpectations(t)
	})

	t.Run("permission denied", func(t *testing.T) {
		mockClient := new(MockKeycloakClient)
		provider := &KeycloakProvider{client: mockClient}

		expectedTokenInfo := &keycloak.TokenInfo{
			Subject:  "user123",
			Username: "testuser",
			Roles:    []string{"user"},
		}

		mockClient.On("ValidateToken", mock.Anything, "valid-token").Return(expectedTokenInfo, nil)
		mockClient.On("CheckPermission", expectedTokenInfo, "resource", "action").Return(false)

		ctx := context.Background()
		hasPermission, err := provider.CheckPermission(ctx, "valid-token", "resource", "action")

		assert.NoError(t, err)
		assert.False(t, hasPermission)

		mockClient.AssertExpectations(t)
	})

	t.Run("token validation error", func(t *testing.T) {
		mockClient := new(MockKeycloakClient)
		provider := &KeycloakProvider{client: mockClient}

		mockClient.On("ValidateToken", mock.Anything, "invalid-token").Return(nil, assert.AnError)

		ctx := context.Background()
		hasPermission, err := provider.CheckPermission(ctx, "invalid-token", "resource", "action")

		assert.Error(t, err)
		assert.False(t, hasPermission)

		mockClient.AssertExpectations(t)
	})

	t.Run("empty token", func(t *testing.T) {
		mockClient := new(MockKeycloakClient)
		provider := &KeycloakProvider{client: mockClient}

		mockClient.On("ValidateToken", mock.Anything, "").Return(nil, assert.AnError)

		ctx := context.Background()
		hasPermission, err := provider.CheckPermission(ctx, "", "resource", "action")

		assert.Error(t, err)
		assert.False(t, hasPermission)

		mockClient.AssertExpectations(t)
	})

	t.Run("empty resource and action", func(t *testing.T) {
		mockClient := new(MockKeycloakClient)
		provider := &KeycloakProvider{client: mockClient}

		expectedTokenInfo := &keycloak.TokenInfo{
			Subject:  "user123",
			Username: "testuser",
			Roles:    []string{"admin"},
		}

		mockClient.On("ValidateToken", mock.Anything, "valid-token").Return(expectedTokenInfo, nil)
		mockClient.On("CheckPermission", expectedTokenInfo, "", "").Return(false)

		ctx := context.Background()
		hasPermission, err := provider.CheckPermission(ctx, "valid-token", "", "")

		assert.NoError(t, err)
		assert.False(t, hasPermission)

		mockClient.AssertExpectations(t)
	})
}

func TestKeycloakProvider_EdgeCases(t *testing.T) {
	t.Run("nil client", func(t *testing.T) {
		provider := &KeycloakProvider{client: nil}

		ctx := context.Background()
		
		// This should panic or handle gracefully
		assert.Panics(t, func() {
			provider.ValidateToken(ctx, "token")
		})
	})

	t.Run("token with special characters", func(t *testing.T) {
		mockClient := new(MockKeycloakClient)
		provider := &KeycloakProvider{client: mockClient}

		specialToken := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VyMTIzIn0.signature"
		
		mockClient.On("ValidateToken", mock.Anything, specialToken).Return(nil, assert.AnError)

		ctx := context.Background()
		tokenInfo, err := provider.ValidateToken(ctx, specialToken)

		assert.Error(t, err)
		assert.Nil(t, tokenInfo)

		mockClient.AssertExpectations(t)
	})

	t.Run("very long token", func(t *testing.T) {
		mockClient := new(MockKeycloakClient)
		provider := &KeycloakProvider{client: mockClient}

		longToken := "very-long-token-" + string(make([]byte, 1000))
		
		mockClient.On("ValidateToken", mock.Anything, longToken).Return(nil, assert.AnError)

		ctx := context.Background()
		tokenInfo, err := provider.ValidateToken(ctx, longToken)

		assert.Error(t, err)
		assert.Nil(t, tokenInfo)

		mockClient.AssertExpectations(t)
	})
}
