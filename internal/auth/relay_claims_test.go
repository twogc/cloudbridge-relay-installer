package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRelayClaims_HasPermission(t *testing.T) {
	claims := &RelayClaims{
		Permissions: []string{"p2p_connect", "mesh_join", "relay_access"},
	}

	// Тестируем существующие разрешения
	assert.True(t, claims.HasPermission("p2p_connect"))
	assert.True(t, claims.HasPermission("mesh_join"))
	assert.True(t, claims.HasPermission("relay_access"))

	// Тестируем несуществующие разрешения
	assert.False(t, claims.HasPermission("admin_access"))
	assert.False(t, claims.HasPermission("network:allocate"))

	// Тестируем nil claims
	var nilClaims *RelayClaims
	assert.False(t, nilClaims.HasPermission("p2p_connect"))
}

func TestRelayClaims_HasRole(t *testing.T) {
	claims := &RelayClaims{
		Roles: []string{"service", "relay-service", "network-manager"},
	}

	// Тестируем существующие роли
	assert.True(t, claims.HasRole("service"))
	assert.True(t, claims.HasRole("relay-service"))
	assert.True(t, claims.HasRole("network-manager"))

	// Тестируем несуществующие роли
	assert.False(t, claims.HasRole("admin"))
	assert.False(t, claims.HasRole("user"))

	// Тестируем nil claims
	var nilClaims *RelayClaims
	assert.False(t, nilClaims.HasRole("service"))
}

func TestRelayClaims_IsRelayToken(t *testing.T) {
	tests := []struct {
		name      string
		tokenType string
		expected  bool
	}{
		{"Permanent token", "relay_permanent", true},
		{"Session token", "relay_session", true},
		{"User token", "user_session", false},
		{"Service token", "service_session", false},
		{"Empty token type", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := &RelayClaims{TokenType: tt.tokenType}
			assert.Equal(t, tt.expected, claims.IsRelayToken())
		})
	}

	// Тестируем nil claims
	var nilClaims *RelayClaims
	assert.False(t, nilClaims.IsRelayToken())
}

func TestRelayClaims_IsPermanentToken(t *testing.T) {
	tests := []struct {
		name      string
		tokenType string
		expected  bool
	}{
		{"Permanent token", "relay_permanent", true},
		{"Session token", "relay_session", false},
		{"User token", "user_session", false},
		{"Empty token type", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := &RelayClaims{TokenType: tt.tokenType}
			assert.Equal(t, tt.expected, claims.IsPermanentToken())
		})
	}

	// Тестируем nil claims
	var nilClaims *RelayClaims
	assert.False(t, nilClaims.IsPermanentToken())
}

func TestRelayClaims_GetExpirationTime(t *testing.T) {
	now := time.Now()
	expTime := now.Add(time.Hour)

	claims := &RelayClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expTime),
		},
	}

	exp, err := claims.GetExpirationTime()
	require.NoError(t, err)
	require.NotNil(t, exp)
	assert.Equal(t, expTime.Unix(), exp.Unix())

	// Тестируем claims без времени истечения
	claimsNoExp := &RelayClaims{}
	expNoExp, err := claimsNoExp.GetExpirationTime()
	require.NoError(t, err)
	assert.Nil(t, expNoExp)

	// Тестируем nil claims
	var nilClaims *RelayClaims
	expNil, err := nilClaims.GetExpirationTime()
	require.NoError(t, err)
	assert.Nil(t, expNil)
}

func TestRelayClaims_ValidatePermissions(t *testing.T) {
	claims := &RelayClaims{
		Permissions: []string{"p2p_connect", "mesh_join", "relay_access", "network:allocate"},
	}

	// Тестируем валидные разрешения
	validPermissions := []string{"p2p_connect", "mesh_join"}
	assert.True(t, claims.ValidatePermissions(validPermissions))

	// Тестируем невалидные разрешения
	invalidPermissions := []string{"p2p_connect", "admin_access"}
	assert.False(t, claims.ValidatePermissions(invalidPermissions))

	// Тестируем пустой список разрешений
	assert.True(t, claims.ValidatePermissions([]string{}))

	// Тестируем nil claims
	var nilClaims *RelayClaims
	assert.False(t, nilClaims.ValidatePermissions([]string{"p2p_connect"}))
}

func TestRelayClaims_ValidateRoles(t *testing.T) {
	claims := &RelayClaims{
		Roles: []string{"service", "relay-service", "network-manager"},
	}

	// Тестируем валидные роли
	validRoles := []string{"service", "relay-service"}
	assert.True(t, claims.ValidateRoles(validRoles))

	// Тестируем невалидные роли
	invalidRoles := []string{"service", "admin"}
	assert.False(t, claims.ValidateRoles(invalidRoles))

	// Тестируем пустой список ролей
	assert.False(t, claims.ValidateRoles([]string{}))

	// Тестируем nil claims
	var nilClaims *RelayClaims
	assert.False(t, nilClaims.ValidateRoles([]string{"service"}))
}

func TestRelayClaims_GetRelayServerID(t *testing.T) {
	claims := &RelayClaims{
		RelayServerID: "relay-server-12345",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "relay-server-67890",
		},
	}

	// Тестируем с RelayServerID
	assert.Equal(t, "relay-server-12345", claims.GetRelayServerID())

	// Тестируем без RelayServerID (должен вернуть Subject)
	claimsNoID := &RelayClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "relay-server-67890",
		},
	}
	assert.Equal(t, "relay-server-67890", claimsNoID.GetRelayServerID())

	// Тестируем nil claims
	var nilClaims *RelayClaims
	assert.Equal(t, "", nilClaims.GetRelayServerID())
}

func TestRelayClaims_GetTenantID(t *testing.T) {
	claims := &RelayClaims{
		TenantID: "tenant-12345",
	}

	assert.Equal(t, "tenant-12345", claims.GetTenantID())

	// Тестируем nil claims
	var nilClaims *RelayClaims
	assert.Equal(t, "", nilClaims.GetTenantID())
}

func TestRelayClaims_GetOrgID(t *testing.T) {
	claims := &RelayClaims{
		OrgID: "org-12345",
	}

	assert.Equal(t, "org-12345", claims.GetOrgID())

	// Тестируем nil claims
	var nilClaims *RelayClaims
	assert.Equal(t, "", nilClaims.GetOrgID())
}

func TestRelayClaims_NetworkConfig(t *testing.T) {
	networkConfig := &NetworkConfig{
		AllowedSubnets: []string{"10.0.0.0/8", "192.168.0.0/16"},
		MaxConnections: 1000,
		BandwidthLimit: 1024 * 1024 * 1024, // 1GB
	}

	claims := &RelayClaims{
		NetworkConfig: networkConfig,
	}

	assert.NotNil(t, claims.NetworkConfig)
	assert.Equal(t, []string{"10.0.0.0/8", "192.168.0.0/16"}, claims.NetworkConfig.AllowedSubnets)
	assert.Equal(t, 1000, claims.NetworkConfig.MaxConnections)
	assert.Equal(t, int64(1024*1024*1024), claims.NetworkConfig.BandwidthLimit)
}
