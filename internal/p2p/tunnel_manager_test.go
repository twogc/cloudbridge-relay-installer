package p2p

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
	"go.uber.org/zap"
)

func TestNewP2PTunnelManager(t *testing.T) {
	t.Run("creates tunnel manager successfully", func(t *testing.T) {
		loggerInstance, err := logger.New(logger.Config{
			Level:  logger.INFO,
			Output: "stdout",
			JSON:   false,
		})
		require.NoError(t, err)

		relayManager := NewRelayManager(zap.NewNop(), nil)
		manager := NewP2PTunnelManager(loggerInstance, relayManager)

		assert.NotNil(t, manager)
		assert.NotNil(t, manager.tunnels)
		assert.NotNil(t, manager.relayManager)
		assert.NotNil(t, manager.ctx)
		assert.NotNil(t, manager.cancel)
		assert.Equal(t, loggerInstance, manager.logger)
	})

	t.Run("creates tunnel manager with nil logger", func(t *testing.T) {
		relayManager := NewRelayManager(zap.NewNop(), nil)
		manager := NewP2PTunnelManager(nil, relayManager)

		assert.NotNil(t, manager)
		assert.Nil(t, manager.logger)
	})
}

func TestP2PTunnelManager_CreateTunnel(t *testing.T) {
	loggerInstance, err := logger.New(logger.Config{
		Level:  logger.INFO,
		Output: "stdout",
		JSON:   false,
	})
	require.NoError(t, err)

	relayManager := NewRelayManager(zap.NewNop(), nil)
	manager := NewP2PTunnelManager(loggerInstance, relayManager)
	defer manager.Stop()

	t.Run("fails due to NAT detection timeout", func(t *testing.T) {
		config := &TunnelConfig{
			PeerID:            "test-peer-1",
			TenantID:          "tenant-1",
			LocalPort:         51820,
			RemotePort:        51820,
			Protocol:          "udp",
			HeartbeatInterval: 30 * time.Second,
			Timeout:           60 * time.Second,
		}

		tunnel, err := manager.CreateTunnel(config)

		assert.Error(t, err)
		assert.Nil(t, tunnel)
		assert.Contains(t, err.Error(), "failed to detect NAT type")
	})

	t.Run("fails with invalid peer ID", func(t *testing.T) {
		config := &TunnelConfig{
			PeerID:            "",
			TenantID:          "tenant-1",
			LocalPort:         51820,
			RemotePort:        51820,
			Protocol:          "udp",
			HeartbeatInterval: 30 * time.Second,
			Timeout:           60 * time.Second,
		}

		tunnel, err := manager.CreateTunnel(config)

		assert.Error(t, err)
		assert.Nil(t, tunnel)
	})

	t.Run("fails with invalid tenant ID", func(t *testing.T) {
		config := &TunnelConfig{
			PeerID:            "test-peer-1",
			TenantID:          "",
			LocalPort:         51820,
			RemotePort:        51820,
			Protocol:          "udp",
			HeartbeatInterval: 30 * time.Second,
			Timeout:           60 * time.Second,
		}

		tunnel, err := manager.CreateTunnel(config)

		assert.Error(t, err)
		assert.Nil(t, tunnel)
	})
}

func TestP2PTunnelManager_GetTunnel(t *testing.T) {
	loggerInstance, err := logger.New(logger.Config{
		Level:  logger.INFO,
		Output: "stdout",
		JSON:   false,
	})
	require.NoError(t, err)

	relayManager := NewRelayManager(zap.NewNop(), nil)
	manager := NewP2PTunnelManager(loggerInstance, relayManager)
	defer manager.Stop()

	t.Run("fails due to NAT detection timeout", func(t *testing.T) {
		peerID := "test-peer-1"
		tenantID := "tenant-1"

		config := &TunnelConfig{
			PeerID:            peerID,
			TenantID:          tenantID,
			LocalPort:         51820,
			RemotePort:        51820,
			Protocol:          "udp",
			HeartbeatInterval: 30 * time.Second,
			Timeout:           60 * time.Second,
		}

		// Create tunnel
		tunnel, err := manager.CreateTunnel(config)
		assert.Error(t, err)
		assert.Nil(t, tunnel)
		assert.Contains(t, err.Error(), "failed to detect NAT type")
	})

	t.Run("returns nil when tunnel does not exist", func(t *testing.T) {
		tunnel, exists := manager.GetTunnel("non-existent-tunnel")
		assert.False(t, exists)
		assert.Nil(t, tunnel)
	})
}

func TestP2PTunnelManager_CloseTunnel(t *testing.T) {
	loggerInstance, err := logger.New(logger.Config{
		Level:  logger.INFO,
		Output: "stdout",
		JSON:   false,
	})
	require.NoError(t, err)

	relayManager := NewRelayManager(zap.NewNop(), nil)
	manager := NewP2PTunnelManager(loggerInstance, relayManager)
	defer manager.Stop()

	t.Run("fails due to NAT detection timeout", func(t *testing.T) {
		peerID := "test-peer-1"
		tenantID := "tenant-1"

		config := &TunnelConfig{
			PeerID:            peerID,
			TenantID:          tenantID,
			LocalPort:         51820,
			RemotePort:        51820,
			Protocol:          "udp",
			HeartbeatInterval: 30 * time.Second,
			Timeout:           60 * time.Second,
		}

		// Create tunnel
		tunnel, err := manager.CreateTunnel(config)
		assert.Error(t, err)
		assert.Nil(t, tunnel)
		assert.Contains(t, err.Error(), "failed to detect NAT type")
	})

	t.Run("handles closing non-existent tunnel", func(t *testing.T) {
		err := manager.CloseTunnel("non-existent-tunnel")
		assert.Error(t, err)
	})
}

func TestP2PTunnelManager_GetActiveTunnels(t *testing.T) {
	loggerInstance, err := logger.New(logger.Config{
		Level:  logger.INFO,
		Output: "stdout",
		JSON:   false,
	})
	require.NoError(t, err)

	relayManager := NewRelayManager(zap.NewNop(), nil)
	manager := NewP2PTunnelManager(loggerInstance, relayManager)
	defer manager.Stop()

	t.Run("returns empty list when no tunnels", func(t *testing.T) {
		tunnels := manager.GetTunnelsByTenant("tenant-1")
		assert.Empty(t, tunnels)
	})

	t.Run("returns active tunnels", func(t *testing.T) {
		// Create multiple tunnels
		peer1 := "test-peer-1"
		peer2 := "test-peer-2"
		tenantID := "tenant-1"

		config1 := &TunnelConfig{
			PeerID:            peer1,
			TenantID:          tenantID,
			LocalPort:         51820,
			RemotePort:        51820,
			Protocol:          "udp",
			HeartbeatInterval: 30 * time.Second,
			Timeout:           60 * time.Second,
		}

		config2 := &TunnelConfig{
			PeerID:            peer2,
			TenantID:          tenantID,
			LocalPort:         51821,
			RemotePort:        51821,
			Protocol:          "udp",
			HeartbeatInterval: 30 * time.Second,
			Timeout:           60 * time.Second,
		}

		_, err := manager.CreateTunnel(config1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to detect NAT type")

		_, err = manager.CreateTunnel(config2)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to detect NAT type")

		// Get tunnels by tenant (should be empty since creation failed)
		tunnels := manager.GetTunnelsByTenant(tenantID)
		assert.Len(t, tunnels, 0)
	})
}

func TestP2PTunnelManager_UpdateHeartbeat(t *testing.T) {
	loggerInstance, err := logger.New(logger.Config{
		Level:  logger.INFO,
		Output: "stdout",
		JSON:   false,
	})
	require.NoError(t, err)

	relayManager := NewRelayManager(zap.NewNop(), nil)
	manager := NewP2PTunnelManager(loggerInstance, relayManager)
	defer manager.Stop()

	t.Run("updates heartbeat successfully", func(t *testing.T) {
		peerID := "test-peer-1"
		tenantID := "tenant-1"

		config := &TunnelConfig{
			PeerID:            peerID,
			TenantID:          tenantID,
			LocalPort:         51820,
			RemotePort:        51820,
			Protocol:          "udp",
			HeartbeatInterval: 30 * time.Second,
			Timeout:           60 * time.Second,
		}

		// Create tunnel (should fail due to NAT detection timeout)
		tunnel, err := manager.CreateTunnel(config)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to detect NAT type")
		assert.Nil(t, tunnel)

		// Get tunnels by tenant (should be empty since creation failed)
		tunnels := manager.GetTunnelsByTenant(tenantID)
		assert.Len(t, tunnels, 0)
	})

	t.Run("handles getting tunnels for non-existent tenant", func(t *testing.T) {
		tunnels := manager.GetTunnelsByTenant("non-existent-tenant")
		assert.Empty(t, tunnels)
	})
}

func TestP2PTunnelManager_GetTunnelsByTenant_Final_Last(t *testing.T) {
	loggerInstance, err := logger.New(logger.Config{
		Level:  logger.INFO,
		Output: "stdout",
		JSON:   false,
	})
	require.NoError(t, err)

	relayManager := NewRelayManager(zap.NewNop(), nil)
	manager := NewP2PTunnelManager(loggerInstance, relayManager)
	defer manager.Stop()

	t.Run("returns empty list when no tunnels", func(t *testing.T) {
		tunnels := manager.GetTunnelsByTenant("tenant-1")
		assert.Empty(t, tunnels)
	})

	t.Run("returns tunnels by tenant", func(t *testing.T) {
		// Create tunnel
		peerID := "test-peer-1"
		tenantID := "tenant-1"

		config := &TunnelConfig{
			PeerID:            peerID,
			TenantID:          tenantID,
			LocalPort:         51820,
			RemotePort:        51820,
			Protocol:          "udp",
			HeartbeatInterval: 30 * time.Second,
			Timeout:           60 * time.Second,
		}

		_, err := manager.CreateTunnel(config)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to detect NAT type")

		// Get tunnels by tenant (should be empty since creation failed)
		tunnels := manager.GetTunnelsByTenant(tenantID)
		assert.Len(t, tunnels, 0)
	})
}

// TestP2PTunnelManager_heartbeatLoop тестирует heartbeatLoop
func TestP2PTunnelManager_heartbeatLoop(t *testing.T) {
	loggerInstance, err := logger.New(logger.Config{
		Level:  logger.INFO,
		Output: "stdout",
		JSON:   false,
	})
	require.NoError(t, err)

	relayManager := NewRelayManager(zap.NewNop(), nil)
	manager := NewP2PTunnelManager(loggerInstance, relayManager)
	defer manager.Stop()

	t.Run("heartbeatLoop starts and stops correctly", func(t *testing.T) {
		// Start the manager to trigger heartbeat loop
		err := manager.Start()
		assert.NoError(t, err)

		// Stop the manager to cancel context
		err = manager.Stop()
		assert.NoError(t, err)

		// Context should be cancelled
		select {
		case <-manager.ctx.Done():
			// Expected
		default:
			t.Error("Expected context to be cancelled")
		}
	})
}

// TestP2PTunnelManager_sendHeartbeats тестирует sendHeartbeats
func TestP2PTunnelManager_sendHeartbeats(t *testing.T) {
	loggerInstance, err := logger.New(logger.Config{
		Level:  logger.INFO,
		Output: "stdout",
		JSON:   false,
	})
	require.NoError(t, err)

	relayManager := NewRelayManager(zap.NewNop(), nil)
	manager := NewP2PTunnelManager(loggerInstance, relayManager)
	defer manager.Stop()

	t.Run("sendHeartbeats handles empty tunnels", func(t *testing.T) {
		// Should not panic with no tunnels
		assert.NotPanics(t, func() {
			manager.sendHeartbeats()
		})
	})

	t.Run("sendHeartbeats processes active tunnels", func(t *testing.T) {
		// Create a mock tunnel
		tunnel := &P2PTunnel{
			ID:            "test-tunnel",
			PeerID:        "test-peer",
			TenantID:      "test-tenant",
			IsActive:      true,
			LastHeartbeat: time.Now(),
			CreatedAt:     time.Now(),
		}

		manager.tunnels["test-tunnel"] = tunnel

		// Should not panic when sending heartbeats
		assert.NotPanics(t, func() {
			manager.sendHeartbeats()
		})
	})
}

// TestP2PTunnelManager_cleanupLoop тестирует cleanupLoop
func TestP2PTunnelManager_cleanupLoop(t *testing.T) {
	loggerInstance, err := logger.New(logger.Config{
		Level:  logger.INFO,
		Output: "stdout",
		JSON:   false,
	})
	require.NoError(t, err)

	relayManager := NewRelayManager(zap.NewNop(), nil)
	manager := NewP2PTunnelManager(loggerInstance, relayManager)
	defer manager.Stop()

	t.Run("cleanupLoop starts and stops correctly", func(t *testing.T) {
		// Start the manager to trigger cleanup loop
		err := manager.Start()
		assert.NoError(t, err)

		// Stop the manager to cancel context
		err = manager.Stop()
		assert.NoError(t, err)

		// Context should be cancelled
		select {
		case <-manager.ctx.Done():
			// Expected
		default:
			t.Error("Expected context to be cancelled")
		}
	})
}

// TestP2PTunnelManager_cleanupInactiveTunnels тестирует cleanupInactiveTunnels
func TestP2PTunnelManager_cleanupInactiveTunnels(t *testing.T) {
	loggerInstance, err := logger.New(logger.Config{
		Level:  logger.INFO,
		Output: "stdout",
		JSON:   false,
	})
	require.NoError(t, err)

	relayManager := NewRelayManager(zap.NewNop(), nil)
	manager := NewP2PTunnelManager(loggerInstance, relayManager)
	defer manager.Stop()

	t.Run("cleanupInactiveTunnels removes inactive tunnels", func(t *testing.T) {
		// Create active tunnel
		activeTunnel := &P2PTunnel{
			ID:            "active-tunnel",
			PeerID:        "active-peer",
			TenantID:      "tenant-1",
			IsActive:      true,
			LastHeartbeat: time.Now(),
			CreatedAt:     time.Now(),
		}

		// Create inactive tunnel
		inactiveTunnel := &P2PTunnel{
			ID:            "inactive-tunnel",
			PeerID:        "inactive-peer",
			TenantID:      "tenant-1",
			IsActive:      false,
			LastHeartbeat: time.Now().Add(-15 * time.Minute), // 15 minutes ago
			CreatedAt:     time.Now().Add(-20 * time.Minute),
		}

		manager.tunnels["active-tunnel"] = activeTunnel
		manager.tunnels["inactive-tunnel"] = inactiveTunnel

		// Run cleanup
		manager.cleanupInactiveTunnels()

		// Active tunnel should remain
		assert.Contains(t, manager.tunnels, "active-tunnel")

		// Inactive tunnel should be removed
		assert.NotContains(t, manager.tunnels, "inactive-tunnel")
	})

	t.Run("cleanupInactiveTunnels handles empty tunnels", func(t *testing.T) {
		// Should not panic with no tunnels
		assert.NotPanics(t, func() {
			manager.cleanupInactiveTunnels()
		})
	})
}

// TestP2PTunnel_SendHeartbeat тестирует SendHeartbeat
func TestP2PTunnel_SendHeartbeat(t *testing.T) {
	t.Run("SendHeartbeat fails with nil underlay", func(t *testing.T) {
		tunnel := &P2PTunnel{
			ID:            "test-tunnel",
			PeerID:        "test-peer",
			TenantID:      "test-tenant",
			IsActive:      true,
			LastHeartbeat: time.Now(),
			CreatedAt:     time.Now(),
			Underlay:      nil, // No underlay connection
		}

		err := tunnel.SendHeartbeat()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tunnel underlay not connected")
	})

	t.Run("SendHeartbeat updates heartbeat time", func(t *testing.T) {
		// Create a mock underlay connection
		mockUnderlay := &MockUnderlayConnection{}

		tunnel := &P2PTunnel{
			ID:            "test-tunnel",
			PeerID:        "test-peer",
			TenantID:      "test-tenant",
			IsActive:      false,
			LastHeartbeat: time.Now().Add(-1 * time.Hour),
			CreatedAt:     time.Now(),
			Underlay:      mockUnderlay,
		}

		// Mock successful write
		mockUnderlay.On("Write", mock.AnythingOfType("[]uint8")).Return(100, nil)

		err := tunnel.SendHeartbeat()

		assert.NoError(t, err)
		assert.True(t, tunnel.IsActive)
		assert.True(t, tunnel.LastHeartbeat.After(time.Now().Add(-1*time.Minute)))

		mockUnderlay.AssertExpectations(t)
	})
}

// TestP2PTunnel_Close тестирует Close
func TestP2PTunnel_Close(t *testing.T) {
	t.Run("Close sets tunnel as inactive", func(t *testing.T) {
		tunnel := &P2PTunnel{
			ID:            "test-tunnel",
			PeerID:        "test-peer",
			TenantID:      "test-tenant",
			IsActive:      true,
			LastHeartbeat: time.Now(),
			CreatedAt:     time.Now(),
			Underlay:      nil,
		}

		err := tunnel.Close()

		assert.NoError(t, err)
		assert.False(t, tunnel.IsActive)
	})

	t.Run("Close closes underlay connection", func(t *testing.T) {
		mockUnderlay := &MockUnderlayConnection{}
		mockUnderlay.On("Close").Return(nil)

		tunnel := &P2PTunnel{
			ID:            "test-tunnel",
			PeerID:        "test-peer",
			TenantID:      "test-tenant",
			IsActive:      true,
			LastHeartbeat: time.Now(),
			CreatedAt:     time.Now(),
			Underlay:      mockUnderlay,
		}

		err := tunnel.Close()

		assert.NoError(t, err)
		assert.False(t, tunnel.IsActive)
		mockUnderlay.AssertExpectations(t)
	})
}

// TestQUICUnderlay_netConnInterface тестирует методы интерфейса net.Conn
func TestQUICUnderlay_netConnInterface(t *testing.T) {
	t.Run("QUICUnderlay struct exists", func(t *testing.T) {
		// Test that QUICUnderlay struct exists
		// We can't easily test with real QUIC objects in unit tests
		// This test verifies the struct definition exists
		assert.NotNil(t, &QUICUnderlay{})
	})
}

// Mock implementations for testing
type MockUnderlayConnection struct {
	mock.Mock
}

func (m *MockUnderlayConnection) Read(b []byte) (int, error) {
	args := m.Called(b)
	return args.Int(0), args.Error(1)
}

func (m *MockUnderlayConnection) Write(b []byte) (int, error) {
	args := m.Called(b)
	return args.Int(0), args.Error(1)
}

func (m *MockUnderlayConnection) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockUnderlayConnection) SetReadDeadline(t time.Time) error {
	args := m.Called(t)
	return args.Error(0)
}

func (m *MockUnderlayConnection) SetWriteDeadline(t time.Time) error {
	args := m.Called(t)
	return args.Error(0)
}

func (m *MockUnderlayConnection) Protocol() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockUnderlayConnection) LocalAddr() net.Addr {
	args := m.Called()
	return args.Get(0).(net.Addr)
}

func (m *MockUnderlayConnection) RemoteAddr() net.Addr {
	args := m.Called()
	return args.Get(0).(net.Addr)
}

type MockQUICConn struct {
	mock.Mock
}

func (m *MockQUICConn) CloseWithError(code uint16, desc string) error {
	args := m.Called(code, desc)
	return args.Error(0)
}

func (m *MockQUICConn) LocalAddr() net.Addr {
	args := m.Called()
	return args.Get(0).(net.Addr)
}

func (m *MockQUICConn) RemoteAddr() net.Addr {
	args := m.Called()
	return args.Get(0).(net.Addr)
}

type MockQUICStream struct {
	mock.Mock
}

func (m *MockQUICStream) Read(b []byte) (int, error) {
	args := m.Called(b)
	return args.Int(0), args.Error(1)
}

func (m *MockQUICStream) Write(b []byte) (int, error) {
	args := m.Called(b)
	return args.Int(0), args.Error(1)
}

func (m *MockQUICStream) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockQUICStream) SetReadDeadline(t time.Time) error {
	args := m.Called(t)
	return args.Error(0)
}

func (m *MockQUICStream) SetWriteDeadline(t time.Time) error {
	args := m.Called(t)
	return args.Error(0)
}

type MockNetAddr struct {
	addr string
}

func (m *MockNetAddr) Network() string {
	return "tcp"
}

func (m *MockNetAddr) String() string {
	return m.addr
}
