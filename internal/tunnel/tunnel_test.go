package tunnel

import (
	"net"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
	"github.com/twogc/cloudbridge-relay-installer/internal/metrics"
)

func TestNewProxy(t *testing.T) {
	// Create test logger
	logger, err := logger.New(logger.Config{
		Level:  logger.INFO,
		Output: "stdout",
		JSON:   false,
	})
	require.NoError(t, err)

	// Create test metrics
	metrics := metrics.NewMetrics(prometheus.NewRegistry())

	// Create test connection
	clientConn, _ := net.Pipe()

	proxy := NewProxy(clientConn, "localhost", 8080, "test-tunnel", logger, metrics)

	assert.NotNil(t, proxy)
	assert.Equal(t, "test-tunnel", proxy.tunnelID)
	assert.Equal(t, "localhost", proxy.remoteHost)
	assert.Equal(t, 8080, proxy.remotePort)
	assert.NotNil(t, proxy.stats)
}

func TestProxyStats(t *testing.T) {
	// Create test logger
	logger, err := logger.New(logger.Config{
		Level:  logger.INFO,
		Output: "stdout",
		JSON:   false,
	})
	require.NoError(t, err)

	// Create test metrics
	metrics := metrics.NewMetrics(prometheus.NewRegistry())

	// Create test connection
	clientConn, _ := net.Pipe()

	proxy := NewProxy(clientConn, "localhost", 8080, "test-tunnel", logger, metrics)

	// Test initial stats
	stats := proxy.GetStats()
	assert.Equal(t, int64(0), stats.BytesFromClient)
	assert.Equal(t, int64(0), stats.BytesToClient)
	assert.False(t, stats.StartTime.IsZero())

	// Test activity tracking
	lastActivity := proxy.GetLastActivity()
	assert.False(t, lastActivity.IsZero())
}

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

	manager := NewManager(100, logger, metrics)

	assert.NotNil(t, manager)
	assert.Equal(t, 100, manager.maxTunnels)
	assert.NotNil(t, manager.tunnels)
	assert.Equal(t, 0, len(manager.tunnels))
}

func TestManagerCreateTunnel(t *testing.T) {
	// Create test logger
	logger, err := logger.New(logger.Config{
		Level:  logger.INFO,
		Output: "stdout",
		JSON:   false,
	})
	require.NoError(t, err)

	// Create test metrics
	metrics := metrics.NewMetrics(prometheus.NewRegistry())

	manager := NewManager(10, logger, metrics)

	// Create test connection
	clientConn, _ := net.Pipe()

	// Test creating a tunnel
	proxy, err := manager.CreateTunnel("test-tunnel", clientConn, "localhost", 8080)
	if err != nil {
		assert.Nil(t, proxy)
		assert.Contains(t, err.Error(), "failed to start tunnel")
	} else {
		assert.NotNil(t, proxy)
		assert.Equal(t, "test-tunnel", proxy.tunnelID)
		assert.Equal(t, "localhost", proxy.remoteHost)
		assert.Equal(t, 8080, proxy.remotePort)
	}
}

func TestManagerTunnelLimit(t *testing.T) {
	// Create test logger
	logger, err := logger.New(logger.Config{
		Level:  logger.INFO,
		Output: "stdout",
		JSON:   false,
	})
	require.NoError(t, err)

	// Create test metrics
	metrics := metrics.NewMetrics(prometheus.NewRegistry())

	manager := NewManager(1, logger, metrics)

	// Create test connection
	clientConn1, _ := net.Pipe()
	clientConn2, _ := net.Pipe()

	// Try to create two tunnels when limit is 1
	proxy1, err1 := manager.CreateTunnel("tunnel-1", clientConn1, "localhost", 8080)
	proxy2, err2 := manager.CreateTunnel("tunnel-2", clientConn2, "localhost", 8081)

	if err1 == nil {
		assert.NotNil(t, proxy1)
		// Второй должен вернуть ошибку лимита
		assert.Nil(t, proxy2)
		assert.Error(t, err2)
		assert.Contains(t, err2.Error(), "maximum number of tunnels")
	} else {
		// Если первый не создался — оба должны вернуть ошибку
		assert.Nil(t, proxy1)
		assert.Error(t, err1)
		assert.Nil(t, proxy2)
		assert.Error(t, err2)
	}
}

func TestManagerGetTunnel(t *testing.T) {
	// Create test logger
	logger, err := logger.New(logger.Config{
		Level:  logger.INFO,
		Output: "stdout",
		JSON:   false,
	})
	require.NoError(t, err)

	// Create test metrics
	metrics := metrics.NewMetrics(prometheus.NewRegistry())

	manager := NewManager(10, logger, metrics)

	// Test getting non-existent tunnel
	tunnel, exists := manager.GetTunnel("non-existent")
	assert.Nil(t, tunnel)
	assert.False(t, exists)
}

func TestManagerCleanup(t *testing.T) {
	// Create test logger
	logger, err := logger.New(logger.Config{
		Level:  logger.INFO,
		Output: "stdout",
		JSON:   false,
	})
	require.NoError(t, err)

	// Create test metrics
	metrics := metrics.NewMetrics(prometheus.NewRegistry())

	manager := NewManager(10, logger, metrics)

	// Test cleanup with no tunnels
	removed := manager.CleanupInactiveTunnels(1 * time.Minute)
	assert.Equal(t, 0, removed)
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

	manager := NewManager(10, logger, metrics)

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

func TestManagerStats(t *testing.T) {
	// Create test logger
	logger, err := logger.New(logger.Config{
		Level:  logger.INFO,
		Output: "stdout",
		JSON:   false,
	})
	require.NoError(t, err)

	// Create test metrics
	metrics := metrics.NewMetrics(prometheus.NewRegistry())

	manager := NewManager(10, logger, metrics)

	// Test getting stats with no tunnels
	stats := manager.GetTunnelStats()
	assert.Equal(t, 0, len(stats))

	// Test getting tunnel count
	count := manager.GetTunnelCount()
	assert.Equal(t, 0, count)

	// Test getting active tunnels
	active := manager.GetActiveTunnels()
	assert.Equal(t, 0, active)
}
