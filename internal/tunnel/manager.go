package tunnel

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
	"github.com/twogc/cloudbridge-relay-installer/internal/metrics"
	"go.uber.org/zap"
)

// Manager manages all active tunnels
type Manager struct {
	tunnels    map[string]*Proxy
	mu         sync.RWMutex
	logger     *logger.Logger
	metrics    *metrics.Metrics
	ctx        context.Context
	cancel     context.CancelFunc
	maxTunnels int
}

// NewManager creates a new tunnel manager
func NewManager(maxTunnels int, logger *logger.Logger, metrics *metrics.Metrics) *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &Manager{
		tunnels:    make(map[string]*Proxy),
		logger:     logger,
		metrics:    metrics,
		ctx:        ctx,
		cancel:     cancel,
		maxTunnels: maxTunnels,
	}
}

// CreateTunnel creates a new tunnel
func (m *Manager) CreateTunnel(
	tunnelID string,
	clientConn net.Conn,
	remoteHost string,
	remotePort int,
) (*Proxy, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if tunnel already exists
	if _, exists := m.tunnels[tunnelID]; exists {
		return nil, fmt.Errorf("tunnel %s already exists", tunnelID)
	}

	// Check tunnel limit
	if len(m.tunnels) >= m.maxTunnels {
		return nil, fmt.Errorf("maximum number of tunnels (%d) reached", m.maxTunnels)
	}

	// Create new proxy
	proxy := NewProxy(clientConn, remoteHost, remotePort, tunnelID, m.logger, m.metrics)

	// Start the proxy
	if err := proxy.Start(); err != nil {
		return nil, fmt.Errorf("failed to start tunnel: %v", err)
	}

	// Add to manager
	m.tunnels[tunnelID] = proxy

	// Update metrics
	if m.metrics != nil {
		m.metrics.SetActiveTunnels(len(m.tunnels))
	}

	m.logger.Info("Tunnel created",
		zap.String("tunnel_id", tunnelID),
		zap.String("remote_host", remoteHost),
		zap.Int("remote_port", remotePort),
		zap.Int("total_tunnels", len(m.tunnels)),
	)

	return proxy, nil
}

// GetTunnel returns a tunnel by ID
func (m *Manager) GetTunnel(tunnelID string) (*Proxy, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tunnel, exists := m.tunnels[tunnelID]
	return tunnel, exists
}

// CloseTunnel closes a tunnel by ID
func (m *Manager) CloseTunnel(tunnelID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	tunnel, exists := m.tunnels[tunnelID]
	if !exists {
		return fmt.Errorf("tunnel %s not found", tunnelID)
	}

	// Close the tunnel
	tunnel.Close()

	// Remove from manager
	delete(m.tunnels, tunnelID)

	// Update metrics
	if m.metrics != nil {
		m.metrics.SetActiveTunnels(len(m.tunnels))
	}

	m.logger.Info("Tunnel closed",
		zap.String("tunnel_id", tunnelID),
		zap.Int("total_tunnels", len(m.tunnels)),
	)

	return nil
}

// CloseAllTunnels closes all tunnels
func (m *Manager) CloseAllTunnels() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.logger.Info("Closing all tunnels", zap.Int("count", len(m.tunnels)))

	for tunnelID, tunnel := range m.tunnels {
		tunnel.Close()
		m.logger.Debug("Closed tunnel", zap.String("tunnel_id", tunnelID))
	}

	m.tunnels = make(map[string]*Proxy)

	// Update metrics
	if m.metrics != nil {
		m.metrics.SetActiveTunnels(0)
	}
}

// GetTunnelStats returns statistics for all tunnels
func (m *Manager) GetTunnelStats() map[string]*ProxyStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make(map[string]*ProxyStats)
	for tunnelID, tunnel := range m.tunnels {
		stats[tunnelID] = tunnel.GetStats()
	}

	return stats
}

// GetActiveTunnels returns the number of active tunnels
func (m *Manager) GetActiveTunnels() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	active := 0
	for _, tunnel := range m.tunnels {
		if tunnel.IsActive() {
			active++
		}
	}

	return active
}

// CleanupInactiveTunnels removes inactive tunnels
func (m *Manager) CleanupInactiveTunnels(timeout time.Duration) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	removed := 0

	for tunnelID, tunnel := range m.tunnels {
		lastActivity := tunnel.GetLastActivity()
		if now.Sub(lastActivity) > timeout {
			m.logger.Info("Removing inactive tunnel",
				zap.String("tunnel_id", tunnelID),
				zap.Duration("inactive_duration", now.Sub(lastActivity)),
			)

			tunnel.Close()
			delete(m.tunnels, tunnelID)
			removed++
		}
	}

	// Update metrics
	if m.metrics != nil {
		m.metrics.SetActiveTunnels(len(m.tunnels))
	}

	if removed > 0 {
		m.logger.Info("Cleanup completed",
			zap.Int("removed_tunnels", removed),
			zap.Int("remaining_tunnels", len(m.tunnels)),
		)
	}

	return removed
}

// StartCleanupLoop starts the cleanup loop
func (m *Manager) StartCleanupLoop(interval, timeout time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-m.ctx.Done():
				return
			case <-ticker.C:
				m.CleanupInactiveTunnels(timeout)
			}
		}
	}()
}

// Shutdown gracefully shuts down the manager
func (m *Manager) Shutdown() {
	m.logger.Info("Shutting down tunnel manager")
	m.cancel()
	m.CloseAllTunnels()
}

// GetTunnelCount returns the total number of tunnels
func (m *Manager) GetTunnelCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.tunnels)
}

// IsTunnelActive checks if a tunnel is active
func (m *Manager) IsTunnelActive(tunnelID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tunnel, exists := m.tunnels[tunnelID]
	if !exists {
		return false
	}

	return tunnel.IsActive()
} 