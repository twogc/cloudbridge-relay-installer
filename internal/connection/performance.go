package connection

import (
	"context"
	"net"
	"runtime"
	"sync"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
	"go.uber.org/zap"
)

// ConnectionMultiplexer manages connection multiplexing for better performance
type ConnectionMultiplexer interface {
	GetConnection(key string) (net.Conn, error)
	ReturnConnection(key string, conn net.Conn)
	CloseIdleConnections()
	GetStats() map[string]interface{}
}

// PerformanceMonitor monitors connection performance metrics
type PerformanceMonitor interface {
	GetMemoryUsage() map[string]uint64
	GetCPUUsage() float64
	GetConnectionStats() map[string]interface{}
	StartMonitoring(ctx context.Context)
	StopMonitoring()
}

// ConnectionMultiplexerImpl implements connection multiplexing
type ConnectionMultiplexerImpl struct {
	connections map[string]*MultiplexedConnection
	mu          sync.RWMutex
	logger      *logger.Logger
	config      *MultiplexerConfig
}

// MultiplexerConfig holds configuration for connection multiplexing
type MultiplexerConfig struct {
	MaxConnections       int           `json:"max_connections"`
	MaxIdleTime          time.Duration `json:"max_idle_time"`
	CleanupInterval      time.Duration `json:"cleanup_interval"`
	EnableCompression    bool          `json:"enable_compression"`
	MaxConcurrentStreams int           `json:"max_concurrent_streams"`
}

// DefaultMultiplexerConfig returns default multiplexer configuration
func DefaultMultiplexerConfig() *MultiplexerConfig {
	return &MultiplexerConfig{
		MaxConnections:       100,
		MaxIdleTime:          30 * time.Second,
		CleanupInterval:      60 * time.Second,
		EnableCompression:    true,
		MaxConcurrentStreams: 100,
	}
}

// MultiplexedConnection represents a multiplexed connection
type MultiplexedConnection struct {
	conn          net.Conn
	key           string
	created       time.Time
	lastUsed      time.Time
	useCount      int64
	activeStreams int64
	compressed    bool
	mu            sync.RWMutex
}

// NewConnectionMultiplexer creates a new connection multiplexer
func NewConnectionMultiplexer(config *MultiplexerConfig, logger *logger.Logger) *ConnectionMultiplexerImpl {
	if config == nil {
		config = DefaultMultiplexerConfig()
	}

	multiplexer := &ConnectionMultiplexerImpl{
		connections: make(map[string]*MultiplexedConnection),
		logger:      logger,
		config:      config,
	}

	// Start cleanup goroutine
	go multiplexer.cleanupLoop()

	return multiplexer
}

// GetConnection gets a multiplexed connection
func (cm *ConnectionMultiplexerImpl) GetConnection(key string) (net.Conn, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Check if connection exists and is healthy
	if conn, exists := cm.connections[key]; exists && conn.isHealthy() {
		conn.updateLastUsed()
		conn.incrementUseCount()
		conn.incrementActiveStreams()
		return conn.getConnection(), nil
	}

	// Create new connection
	conn, err := cm.createConnection(key)
	if err != nil {
		return nil, err
	}

	multiplexedConn := &MultiplexedConnection{
		conn:          conn,
		key:           key,
		created:       time.Now(),
		lastUsed:      time.Now(),
		useCount:      1,
		activeStreams: 1,
		compressed:    cm.config.EnableCompression,
	}

	cm.connections[key] = multiplexedConn

	cm.logger.Debug("Created new multiplexed connection",
		zap.String("key", key),
		zap.String("remote_addr", conn.RemoteAddr().String()))

	return conn, nil
}

// ReturnConnection returns a connection to the multiplexer
func (cm *ConnectionMultiplexerImpl) ReturnConnection(key string, conn net.Conn) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if multiplexedConn, exists := cm.connections[key]; exists {
		multiplexedConn.decrementActiveStreams()
		multiplexedConn.updateLastUsed()
	}
}

// CloseIdleConnections closes idle connections
func (cm *ConnectionMultiplexerImpl) CloseIdleConnections() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cutoff := time.Now().Add(-cm.config.MaxIdleTime)
	closed := 0

	for key, conn := range cm.connections {
		if conn.lastUsed.Before(cutoff) && conn.getActiveStreams() == 0 {
			conn.close()
			delete(cm.connections, key)
			closed++
		}
	}

	if closed > 0 {
		cm.logger.Info("Closed idle multiplexed connections",
			zap.Int("closed", closed),
			zap.Int("remaining", len(cm.connections)))
	}
}

// GetStats returns multiplexer statistics
func (cm *ConnectionMultiplexerImpl) GetStats() map[string]interface{} {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	totalConnections := len(cm.connections)
	totalStreams := int64(0)
	totalUseCount := int64(0)

	for _, conn := range cm.connections {
		totalStreams += conn.getActiveStreams()
		totalUseCount += conn.getUseCount()
	}

	return map[string]interface{}{
		"total_connections":      totalConnections,
		"total_active_streams":   totalStreams,
		"total_use_count":        totalUseCount,
		"max_connections":        cm.config.MaxConnections,
		"max_concurrent_streams": cm.config.MaxConcurrentStreams,
		"enable_compression":     cm.config.EnableCompression,
	}
}

// createConnection creates a new connection
func (cm *ConnectionMultiplexerImpl) createConnection(key string) (net.Conn, error) {
	// This is a placeholder - in a real implementation, you would create
	// the actual connection based on the key
	return &net.TCPConn{}, nil
}

// cleanupLoop periodically cleans up idle connections
func (cm *ConnectionMultiplexerImpl) cleanupLoop() {
	ticker := time.NewTicker(cm.config.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		cm.CloseIdleConnections()
	}
}

// MultiplexedConnection methods

func (mc *MultiplexedConnection) isHealthy() bool {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.conn != nil
}

func (mc *MultiplexedConnection) updateLastUsed() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.lastUsed = time.Now()
}

func (mc *MultiplexedConnection) incrementUseCount() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.useCount++
}

func (mc *MultiplexedConnection) incrementActiveStreams() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.activeStreams++
}

func (mc *MultiplexedConnection) decrementActiveStreams() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	if mc.activeStreams > 0 {
		mc.activeStreams--
	}
}

func (mc *MultiplexedConnection) getConnection() net.Conn {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.conn
}

func (mc *MultiplexedConnection) getActiveStreams() int64 {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.activeStreams
}

func (mc *MultiplexedConnection) getUseCount() int64 {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.useCount
}

func (mc *MultiplexedConnection) close() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	if mc.conn != nil {
		mc.conn.Close()
	}
}

// PerformanceMonitorImpl implements performance monitoring
type PerformanceMonitorImpl struct {
	logger    *logger.Logger
	startTime time.Time
	stopChan  chan struct{}
	stats     *PerformanceStats
	mu        sync.RWMutex
}

// PerformanceStats holds performance statistics
type PerformanceStats struct {
	MemoryUsage     map[string]uint64
	CPUUsage        float64
	ConnectionCount int
	ActiveRequests  int
	LastUpdate      time.Time
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor(logger *logger.Logger) *PerformanceMonitorImpl {
	return &PerformanceMonitorImpl{
		logger:    logger,
		startTime: time.Now(),
		stopChan:  make(chan struct{}),
		stats:     &PerformanceStats{},
	}
}

// GetMemoryUsage returns memory usage statistics
func (pm *PerformanceMonitorImpl) GetMemoryUsage() map[string]uint64 {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return map[string]uint64{
		"alloc":       m.Alloc,
		"total_alloc": m.TotalAlloc,
		"sys":         m.Sys,
		"num_gc":      uint64(m.NumGC),
		"heap_alloc":  m.HeapAlloc,
		"heap_sys":    m.HeapSys,
		"heap_idle":   m.HeapIdle,
		"heap_inuse":  m.HeapInuse,
	}
}

// GetCPUUsage returns CPU usage (simplified implementation)
func (pm *PerformanceMonitorImpl) GetCPUUsage() float64 {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.stats.CPUUsage
}

// GetConnectionStats returns connection statistics
func (pm *PerformanceMonitorImpl) GetConnectionStats() map[string]interface{} {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	return map[string]interface{}{
		"connection_count": pm.stats.ConnectionCount,
		"active_requests":  pm.stats.ActiveRequests,
		"uptime":           time.Since(pm.startTime).String(),
		"last_update":      pm.stats.LastUpdate,
	}
}

// StartMonitoring starts performance monitoring
func (pm *PerformanceMonitorImpl) StartMonitoring(ctx context.Context) {
	pm.logger.Info("Starting performance monitoring")

	go func() {
		ticker := time.NewTicker(30 * time.Second) // Update every 30 seconds
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				pm.logger.Info("Performance monitoring stopped")
				return
			case <-pm.stopChan:
				pm.logger.Info("Performance monitoring stopped")
				return
			case <-ticker.C:
				pm.updateStats()
			}
		}
	}()
}

// StopMonitoring stops performance monitoring
func (pm *PerformanceMonitorImpl) StopMonitoring() {
	close(pm.stopChan)
}

// updateStats updates performance statistics
func (pm *PerformanceMonitorImpl) updateStats() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Update memory usage
	pm.stats.MemoryUsage = pm.GetMemoryUsage()

	// Update CPU usage (simplified - in real implementation, you'd use more sophisticated CPU monitoring)
	pm.stats.CPUUsage = pm.calculateCPUUsage()

	// Update connection stats
	pm.stats.ConnectionCount = pm.getConnectionCount()
	pm.stats.ActiveRequests = pm.getActiveRequests()
	pm.stats.LastUpdate = time.Now()

	pm.logger.Debug("Updated performance stats",
		zap.Float64("cpu_usage", pm.stats.CPUUsage),
		zap.Int("connection_count", pm.stats.ConnectionCount),
		zap.Int("active_requests", pm.stats.ActiveRequests))
}

// calculateCPUUsage calculates CPU usage (simplified implementation)
func (pm *PerformanceMonitorImpl) calculateCPUUsage() float64 {
	// This is a simplified implementation
	// In a real implementation, you would use more sophisticated CPU monitoring
	return 0.0
}

// getConnectionCount returns the current connection count
func (pm *PerformanceMonitorImpl) getConnectionCount() int {
	// This is a placeholder - in a real implementation, you would get this from your connection pool
	return 0
}

// getActiveRequests returns the current active request count
func (pm *PerformanceMonitorImpl) getActiveRequests() int {
	// This is a placeholder - in a real implementation, you would get this from your request tracking
	return 0
}
