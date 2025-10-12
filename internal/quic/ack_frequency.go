package quic

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ACKFrequencyManager manages ACK frequency optimization
// Implements draft-ietf-quic-ack-frequency for reduced overhead
// Integrated from quic-test experimental features
type ACKFrequencyManager struct {
	logger      *zap.Logger
	config      *ACKFrequencyConfig
	mu          sync.RWMutex
	connections map[string]*ACKFrequencyConnection
	lastUpdate  time.Time
	metrics     *ACKFrequencyMetrics
}

// ACKFrequencyConfig holds ACK frequency configuration
type ACKFrequencyConfig struct {
	// Basic settings
	MaxACKDelay     time.Duration // Maximum ACK delay
	MinACKDelay     time.Duration // Minimum ACK delay
	DefaultACKDelay time.Duration // Default ACK delay

	// Adaptive settings
	EnableAdaptive    bool // Enable adaptive ACK frequency
	HighLoadThreshold int  // High load threshold (packets/sec)
	LowLoadThreshold  int  // Low load threshold (packets/sec)

	// Traffic type specific settings
	BulkTransferDelay time.Duration // Delay for bulk transfer
	InteractiveDelay  time.Duration // Delay for interactive traffic
	RealTimeDelay     time.Duration // Delay for real-time traffic
}

// ACKFrequencyConnection manages ACK frequency for a single connection
type ACKFrequencyConnection struct {
	connectionID  string
	lastACKTime   time.Time
	packetCount   int64
	bytesReceived int64
	ackDelay      time.Duration
	adaptiveMode  bool
	lastUpdate    time.Time
	metrics       *ConnectionACKMetrics
}

// ACKFrequencyMetrics holds ACK frequency metrics
type ACKFrequencyMetrics struct {
	TotalConnections  int64         `json:"total_connections"`
	ActiveConnections int64         `json:"active_connections"`
	TotalACKs         int64         `json:"total_acks"`
	DelayedACKs       int64         `json:"delayed_acks"`
	AdaptiveACKs      int64         `json:"adaptive_acks"`
	AverageACKDelay   time.Duration `json:"average_ack_delay"`
	BandwidthSaved    int64         `json:"bandwidth_saved_bytes"`
}

// ConnectionACKMetrics holds per-connection ACK metrics
type ConnectionACKMetrics struct {
	ACKsSent        int64         `json:"acks_sent"`
	PacketsReceived int64         `json:"packets_received"`
	BytesReceived   int64         `json:"bytes_received"`
	AverageDelay    time.Duration `json:"average_delay"`
	LastACKTime     time.Time     `json:"last_ack_time"`
}

// NewACKFrequencyManager creates a new ACK frequency manager
func NewACKFrequencyManager(logger *zap.Logger, config *ACKFrequencyConfig) *ACKFrequencyManager {
	if config == nil {
		config = DefaultACKFrequencyConfig()
	}

	return &ACKFrequencyManager{
		logger:      logger,
		config:      config,
		connections: make(map[string]*ACKFrequencyConnection),
		metrics:     &ACKFrequencyMetrics{},
		lastUpdate:  time.Now(),
	}
}

// DefaultACKFrequencyConfig returns default configuration
func DefaultACKFrequencyConfig() *ACKFrequencyConfig {
	return &ACKFrequencyConfig{
		MaxACKDelay:       25 * time.Millisecond,
		MinACKDelay:       1 * time.Millisecond,
		DefaultACKDelay:   5 * time.Millisecond,
		EnableAdaptive:    true,
		HighLoadThreshold: 1000, // 1000 packets/sec
		LowLoadThreshold:  100,  // 100 packets/sec
		BulkTransferDelay: 10 * time.Millisecond,
		InteractiveDelay:  2 * time.Millisecond,
		RealTimeDelay:     1 * time.Millisecond,
	}
}

// RegisterConnection registers a new connection
func (afm *ACKFrequencyManager) RegisterConnection(conn interface{}) string {
	afm.mu.Lock()
	defer afm.mu.Unlock()

	connID := fmt.Sprintf("%p", conn)
	afm.connections[connID] = &ACKFrequencyConnection{
		connectionID: connID,
		lastACKTime:  time.Now(),
		ackDelay:     afm.config.DefaultACKDelay,
		adaptiveMode: afm.config.EnableAdaptive,
		lastUpdate:   time.Now(),
		metrics:      &ConnectionACKMetrics{},
	}

	afm.metrics.TotalConnections++
	afm.metrics.ActiveConnections++

	afm.logger.Info("ACK Frequency connection registered",
		zap.String("connection_id", connID),
		zap.Duration("initial_delay", afm.config.DefaultACKDelay))

	return connID
}

// UnregisterConnection removes a connection
func (afm *ACKFrequencyManager) UnregisterConnection(connID string) {
	afm.mu.Lock()
	defer afm.mu.Unlock()

	if conn, exists := afm.connections[connID]; exists {
		afm.logger.Info("ACK Frequency connection unregistered",
			zap.String("connection_id", connID),
			zap.Int64("acks_sent", conn.metrics.ACKsSent),
			zap.Int64("packets_received", conn.metrics.PacketsReceived))

		delete(afm.connections, connID)
		afm.metrics.ActiveConnections--
	}
}

// ShouldSendACK determines if an ACK should be sent
func (afm *ACKFrequencyManager) ShouldSendACK(connID string, packetSize int) bool {
	afm.mu.RLock()
	defer afm.mu.RUnlock()

	conn, exists := afm.connections[connID]
	if !exists {
		return true // If connection not registered, send ACK
	}

	now := time.Now()

	// Update counters
	conn.packetCount++
	conn.bytesReceived += int64(packetSize)
	conn.metrics.PacketsReceived++
	conn.metrics.BytesReceived += int64(packetSize)

	// Check if enough time has passed
	timeSinceLastACK := now.Sub(conn.lastACKTime)

	// Adaptive delay adjustment
	if conn.adaptiveMode {
		afm.adjustACKDelay(conn, now)
	}

	shouldSend := timeSinceLastACK >= conn.ackDelay

	if shouldSend {
		conn.lastACKTime = now
		conn.metrics.ACKsSent++
		conn.metrics.LastACKTime = now
		afm.metrics.TotalACKs++

		// Update average delay
		if conn.metrics.ACKsSent > 0 {
			conn.metrics.AverageDelay = time.Duration(
				int64(conn.metrics.AverageDelay)*int64(conn.metrics.ACKsSent-1)+
					int64(timeSinceLastACK)) / time.Duration(conn.metrics.ACKsSent)
		}
	} else {
		afm.metrics.DelayedACKs++
	}

	return shouldSend
}

// adjustACKDelay adaptively adjusts ACK delay based on load
func (afm *ACKFrequencyManager) adjustACKDelay(conn *ACKFrequencyConnection, now time.Time) {
	// Calculate packet rate
	timeSinceUpdate := now.Sub(conn.lastUpdate)
	if timeSinceUpdate < time.Second {
		return // Don't update too frequently
	}

	packetsPerSecond := float64(conn.packetCount) / timeSinceUpdate.Seconds()

	// Adaptive adjustment based on load
	if packetsPerSecond > float64(afm.config.HighLoadThreshold) {
		// High load - increase delay
		conn.ackDelay = time.Duration(float64(conn.ackDelay) * 1.2)
		if conn.ackDelay > afm.config.MaxACKDelay {
			conn.ackDelay = afm.config.MaxACKDelay
		}
		afm.metrics.AdaptiveACKs++
	} else if packetsPerSecond < float64(afm.config.LowLoadThreshold) {
		// Low load - decrease delay
		conn.ackDelay = time.Duration(float64(conn.ackDelay) * 0.8)
		if conn.ackDelay < afm.config.MinACKDelay {
			conn.ackDelay = afm.config.MinACKDelay
		}
		afm.metrics.AdaptiveACKs++
	}

	// Reset counters
	conn.packetCount = 0
	conn.lastUpdate = now
}

// SetTrafficType sets traffic type for ACK optimization
func (afm *ACKFrequencyManager) SetTrafficType(connID string, trafficType string) {
	afm.mu.Lock()
	defer afm.mu.Unlock()

	conn, exists := afm.connections[connID]
	if !exists {
		return
	}

	switch trafficType {
	case "bulk":
		conn.ackDelay = afm.config.BulkTransferDelay
	case "interactive":
		conn.ackDelay = afm.config.InteractiveDelay
	case "realtime":
		conn.ackDelay = afm.config.RealTimeDelay
	default:
		conn.ackDelay = afm.config.DefaultACKDelay
	}

	afm.logger.Debug("ACK delay adjusted for traffic type",
		zap.String("connection_id", connID),
		zap.String("traffic_type", trafficType),
		zap.Duration("ack_delay", conn.ackDelay))
}

// GetMetrics returns ACK frequency metrics
func (afm *ACKFrequencyManager) GetMetrics() *ACKFrequencyMetrics {
	afm.mu.RLock()
	defer afm.mu.RUnlock()

	// Update average delay
	totalDelay := time.Duration(0)
	connectionCount := 0

	for _, conn := range afm.connections {
		if conn.metrics.ACKsSent > 0 {
			totalDelay += conn.metrics.AverageDelay
			connectionCount++
		}
	}

	if connectionCount > 0 {
		afm.metrics.AverageACKDelay = totalDelay / time.Duration(connectionCount)
	}

	// Calculate bandwidth saved
	// (approximate - each delayed ACK saves ~64 bytes)
	afm.metrics.BandwidthSaved = afm.metrics.DelayedACKs * 64

	// Return a copy
	metricsCopy := *afm.metrics
	return &metricsCopy
}

// GetConnectionMetrics returns metrics for a specific connection
func (afm *ACKFrequencyManager) GetConnectionMetrics(connID string) *ConnectionACKMetrics {
	afm.mu.RLock()
	defer afm.mu.RUnlock()

	conn, exists := afm.connections[connID]
	if !exists {
		return nil
	}

	// Return a copy
	metricsCopy := *conn.metrics
	return &metricsCopy
}

// StartMonitoring starts ACK frequency monitoring
func (afm *ACKFrequencyManager) StartMonitoring(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			afm.logMonitoringStats()
		}
	}
}

// logMonitoringStats logs monitoring statistics
func (afm *ACKFrequencyManager) logMonitoringStats() {
	afm.mu.RLock()
	defer afm.mu.RUnlock()

	afm.logger.Info("ACK Frequency monitoring stats",
		zap.Int64("active_connections", afm.metrics.ActiveConnections),
		zap.Int64("total_acks", afm.metrics.TotalACKs),
		zap.Int64("delayed_acks", afm.metrics.DelayedACKs),
		zap.Int64("adaptive_acks", afm.metrics.AdaptiveACKs),
		zap.Duration("average_delay", afm.metrics.AverageACKDelay),
		zap.Int64("bandwidth_saved_bytes", afm.metrics.BandwidthSaved))
}
