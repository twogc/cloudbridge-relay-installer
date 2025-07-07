package tenant

import (
	"testing"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
	"github.com/twogc/cloudbridge-relay-installer/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	// Create test logger
	logger, err := logger.New(logger.Config{
		Level:  logger.INFO,
		Output: "stdout",
		JSON:   false,
	})
	require.NoError(t, err)

	// Create test metrics
	metrics := metrics.NewMetrics(prometheus.NewRegistry())

	manager := NewManager(logger, metrics)

	assert.NotNil(t, manager)
	assert.NotNil(t, manager.tenants)
	assert.NotNil(t, manager.tenantLimits)
	assert.Equal(t, 0, len(manager.tenants))
}

func TestCreateTenant(t *testing.T) {
	// Create test logger
	logger, err := logger.New(logger.Config{
		Level:  logger.INFO,
		Output: "stdout",
		JSON:   false,
	})
	require.NoError(t, err)

	// Create test metrics
	metrics := metrics.NewMetrics(prometheus.NewRegistry())

	manager := NewManager(logger, metrics)

	limits := &TenantLimits{
		MaxTunnels:     10,
		MaxConnections: 100,
		RateLimit:      100.0,
		Burst:          200,
		AllowedIPs:     []string{"192.168.1.0/24"},
		AllowedPorts:   []int{80, 443, 22},
	}

	// Test creating a tenant
	err = manager.CreateTenant("test-tenant", "Test Company", limits)
	assert.NoError(t, err)

	// Test creating duplicate tenant
	err = manager.CreateTenant("test-tenant", "Test Company 2", limits)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestGetTenant(t *testing.T) {
	// Create test logger
	logger, err := logger.New(logger.Config{
		Level:  logger.INFO,
		Output: "stdout",
		JSON:   false,
	})
	require.NoError(t, err)

	// Create test metrics
	metrics := metrics.NewMetrics(prometheus.NewRegistry())

	manager := NewManager(logger, metrics)

	limits := &TenantLimits{
		MaxTunnels:     10,
		MaxConnections: 100,
		RateLimit:      100.0,
		Burst:          200,
	}

	// Create a tenant
	err = manager.CreateTenant("test-tenant", "Test Company", limits)
	require.NoError(t, err)

	// Test getting existing tenant
	tenant, exists := manager.GetTenant("test-tenant")
	assert.True(t, exists)
	assert.NotNil(t, tenant)
	assert.Equal(t, "test-tenant", tenant.ID)
	assert.Equal(t, "Test Company", tenant.Name)
	assert.Equal(t, 10, tenant.MaxTunnels)
	assert.Equal(t, 100, tenant.MaxConnections)

	// Test getting non-existent tenant
	tenant, exists = manager.GetTenant("non-existent")
	assert.False(t, exists)
	assert.Nil(t, tenant)
}

func TestGetTenantLimits(t *testing.T) {
	// Create test logger
	logger, err := logger.New(logger.Config{
		Level:  logger.INFO,
		Output: "stdout",
		JSON:   false,
	})
	require.NoError(t, err)

	// Create test metrics
	metrics := metrics.NewMetrics(prometheus.NewRegistry())

	manager := NewManager(logger, metrics)

	limits := &TenantLimits{
		MaxTunnels:     10,
		MaxConnections: 100,
		RateLimit:      100.0,
		Burst:          200,
		AllowedIPs:     []string{"192.168.1.0/24"},
		AllowedPorts:   []int{80, 443, 22},
	}

	// Create a tenant
	err = manager.CreateTenant("test-tenant", "Test Company", limits)
	require.NoError(t, err)

	// Test getting existing tenant limits
	retrievedLimits, exists := manager.GetTenantLimits("test-tenant")
	assert.True(t, exists)
	assert.NotNil(t, retrievedLimits)
	assert.Equal(t, 10, retrievedLimits.MaxTunnels)
	assert.Equal(t, 100, retrievedLimits.MaxConnections)
	assert.Equal(t, 100.0, retrievedLimits.RateLimit)
	assert.Equal(t, 200, retrievedLimits.Burst)
	assert.Equal(t, []string{"192.168.1.0/24"}, retrievedLimits.AllowedIPs)
	assert.Equal(t, []int{80, 443, 22}, retrievedLimits.AllowedPorts)

	// Test getting non-existent tenant limits
	retrievedLimits, exists = manager.GetTenantLimits("non-existent")
	assert.False(t, exists)
	assert.Nil(t, retrievedLimits)
}

func TestCheckTenantLimits(t *testing.T) {
	// Create test logger
	logger, err := logger.New(logger.Config{
		Level:  logger.INFO,
		Output: "stdout",
		JSON:   false,
	})
	require.NoError(t, err)

	// Create test metrics
	metrics := metrics.NewMetrics(prometheus.NewRegistry())

	manager := NewManager(logger, metrics)

	limits := &TenantLimits{
		MaxTunnels:     5,
		MaxConnections: 50,
		RateLimit:      100.0,
		Burst:          200,
	}

	// Create a tenant
	err = manager.CreateTenant("test-tenant", "Test Company", limits)
	require.NoError(t, err)

	// Test within limits
	err = manager.CheckTenantLimits("test-tenant", 3, 25)
	assert.NoError(t, err)

	// Test tunnel limit exceeded
	err = manager.CheckTenantLimits("test-tenant", 6, 25)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tunnel limit exceeded")

	// Test connection limit exceeded
	err = manager.CheckTenantLimits("test-tenant", 3, 51)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection limit exceeded")

	// Test non-existent tenant
	err = manager.CheckTenantLimits("non-existent", 1, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestIsPortAllowed(t *testing.T) {
	// Create test logger
	logger, err := logger.New(logger.Config{
		Level:  logger.INFO,
		Output: "stdout",
		JSON:   false,
	})
	require.NoError(t, err)

	// Create test metrics
	metrics := metrics.NewMetrics(prometheus.NewRegistry())

	manager := NewManager(logger, metrics)

	limits := &TenantLimits{
		MaxTunnels:     10,
		MaxConnections: 100,
		RateLimit:      100.0,
		Burst:          200,
		AllowedPorts:   []int{80, 443, 22},
	}

	// Create a tenant
	err = manager.CreateTenant("test-tenant", "Test Company", limits)
	require.NoError(t, err)

	// Test allowed ports
	assert.True(t, manager.IsPortAllowed("test-tenant", 80))
	assert.True(t, manager.IsPortAllowed("test-tenant", 443))
	assert.True(t, manager.IsPortAllowed("test-tenant", 22))

	// Test disallowed ports
	assert.False(t, manager.IsPortAllowed("test-tenant", 8080))
	assert.False(t, manager.IsPortAllowed("test-tenant", 3306))

	// Test non-existent tenant
	assert.False(t, manager.IsPortAllowed("non-existent", 80))
}

func TestIsIPAllowed(t *testing.T) {
	// Create test logger
	logger, err := logger.New(logger.Config{
		Level:  logger.INFO,
		Output: "stdout",
		JSON:   false,
	})
	require.NoError(t, err)

	// Create test metrics
	metrics := metrics.NewMetrics(prometheus.NewRegistry())

	manager := NewManager(logger, metrics)

	limits := &TenantLimits{
		MaxTunnels:     10,
		MaxConnections: 100,
		RateLimit:      100.0,
		Burst:          200,
		AllowedIPs:     []string{"192.168.1.100", "10.0.0.0/8"},
	}

	// Create a tenant
	err = manager.CreateTenant("test-tenant", "Test Company", limits)
	require.NoError(t, err)

	// Test allowed IPs
	assert.True(t, manager.IsIPAllowed("test-tenant", "192.168.1.100"))
	assert.True(t, manager.IsIPAllowed("test-tenant", "10.0.0.1"))
	assert.True(t, manager.IsIPAllowed("test-tenant", "10.255.255.255"))

	// Test disallowed IPs
	assert.False(t, manager.IsIPAllowed("test-tenant", "192.168.1.101"))
	assert.False(t, manager.IsIPAllowed("test-tenant", "172.16.0.1"))

	// Test non-existent tenant
	assert.False(t, manager.IsIPAllowed("non-existent", "192.168.1.100"))
}

func TestUpdateTenantActivity(t *testing.T) {
	// Create test logger
	logger, err := logger.New(logger.Config{
		Level:  logger.INFO,
		Output: "stdout",
		JSON:   false,
	})
	require.NoError(t, err)

	// Create test metrics
	metrics := metrics.NewMetrics(prometheus.NewRegistry())

	manager := NewManager(logger, metrics)

	limits := &TenantLimits{
		MaxTunnels:     10,
		MaxConnections: 100,
		RateLimit:      100.0,
		Burst:          200,
	}

	// Create a tenant
	err = manager.CreateTenant("test-tenant", "Test Company", limits)
	require.NoError(t, err)

	// Get initial activity time
	tenant, exists := manager.GetTenant("test-tenant")
	require.True(t, exists)
	initialActivity := tenant.LastActivity

	// Wait a bit
	time.Sleep(10 * time.Millisecond)

	// Update activity
	manager.UpdateTenantActivity("test-tenant")

	// Check that activity was updated
	tenant, exists = manager.GetTenant("test-tenant")
	require.True(t, exists)
	assert.True(t, tenant.LastActivity.After(initialActivity))
}

func TestGetTenantStats(t *testing.T) {
	// Create test logger
	logger, err := logger.New(logger.Config{
		Level:  logger.INFO,
		Output: "stdout",
		JSON:   false,
	})
	require.NoError(t, err)

	// Create test metrics
	metrics := metrics.NewMetrics(prometheus.NewRegistry())

	manager := NewManager(logger, metrics)

	limits := &TenantLimits{
		MaxTunnels:     10,
		MaxConnections: 100,
		RateLimit:      100.0,
		Burst:          200,
		AllowedIPs:     []string{"192.168.1.0/24"},
		AllowedPorts:   []int{80, 443},
	}

	// Create a tenant
	err = manager.CreateTenant("test-tenant", "Test Company", limits)
	require.NoError(t, err)

	// Get stats
	stats := manager.GetTenantStats()

	// Verify stats
	assert.Equal(t, 1, len(stats))
	tenantStats, exists := stats["test-tenant"]
	assert.True(t, exists)
	assert.Equal(t, "Test Company", tenantStats["name"])
	assert.Equal(t, 10, tenantStats["max_tunnels"])
	assert.Equal(t, 100, tenantStats["max_connections"])
	assert.Equal(t, 100.0, tenantStats["rate_limit"])
	assert.Equal(t, 200, tenantStats["burst"])
	assert.NotNil(t, tenantStats["created_at"])
	assert.NotNil(t, tenantStats["last_activity"])
}

func TestCleanupInactiveTenants(t *testing.T) {
	// Create test logger
	logger, err := logger.New(logger.Config{
		Level:  logger.INFO,
		Output: "stdout",
		JSON:   false,
	})
	require.NoError(t, err)

	// Create test metrics
	metrics := metrics.NewMetrics(prometheus.NewRegistry())

	manager := NewManager(logger, metrics)

	limits := &TenantLimits{
		MaxTunnels:     10,
		MaxConnections: 100,
		RateLimit:      100.0,
		Burst:          200,
	}

	// Create tenants
	err = manager.CreateTenant("active-tenant", "Active Company", limits)
	require.NoError(t, err)
	err = manager.CreateTenant("inactive-tenant", "Inactive Company", limits)
	require.NoError(t, err)

	// Update activity for active tenant
	manager.UpdateTenantActivity("active-tenant")

	// Cleanup with short timeout (should remove inactive tenant)
	removed := manager.CleanupInactiveTenants(1 * time.Millisecond)
	assert.Equal(t, 1, removed)

	// Verify active tenant still exists
	_, exists := manager.GetTenant("active-tenant")
	assert.True(t, exists)

	// Verify inactive tenant was removed
	_, exists = manager.GetTenant("inactive-tenant")
	assert.False(t, exists)
}

func TestManagerShutdown(t *testing.T) {
	// Create test logger
	logger, err := logger.New(logger.Config{
		Level:  logger.INFO,
		Output: "stdout",
		JSON:   false,
	})
	require.NoError(t, err)

	// Create test metrics
	metrics := metrics.NewMetrics(prometheus.NewRegistry())

	manager := NewManager(logger, metrics)

	// Test shutdown
	manager.Shutdown()
	
	// Verify context is cancelled
	select {
	case <-manager.ctx.Done():
		// Expected
	default:
		t.Error("Context should be cancelled after shutdown")
	}
}

func TestGetTenantCount(t *testing.T) {
	// Create test logger
	logger, err := logger.New(logger.Config{
		Level:  logger.INFO,
		Output: "stdout",
		JSON:   false,
	})
	require.NoError(t, err)

	// Create test metrics
	metrics := metrics.NewMetrics(prometheus.NewRegistry())

	manager := NewManager(logger, metrics)

	// Test initial count
	assert.Equal(t, 0, manager.GetTenantCount())

	limits := &TenantLimits{
		MaxTunnels:     10,
		MaxConnections: 100,
		RateLimit:      100.0,
		Burst:          200,
	}

	// Create tenants
	err = manager.CreateTenant("tenant-1", "Company 1", limits)
	require.NoError(t, err)
	err = manager.CreateTenant("tenant-2", "Company 2", limits)
	require.NoError(t, err)

	// Test count after creating tenants
	assert.Equal(t, 2, manager.GetTenantCount())
} 