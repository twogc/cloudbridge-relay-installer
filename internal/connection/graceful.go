package connection

import (
	"context"
	"sync"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
	"go.uber.org/zap"
)

// GracefulShutdown interface for managing graceful shutdown
type GracefulShutdown interface {
	Shutdown(ctx context.Context, gracePeriod time.Duration) error
	IsShuttingDown() bool
	AddActiveRequest()
	RemoveActiveRequest()
	GetActiveRequests() int
}

// GracefulShutdownManager manages graceful shutdown for connections
// Adapted from cloudflared connection/control.go and connection/http2.go
type GracefulShutdownManager struct {
	activeRequests sync.WaitGroup
	shutdownChan   chan struct{}
	shuttingDown   bool
	mu             sync.RWMutex
	logger         *logger.Logger
}

// GracefulShutdownConfig holds configuration for graceful shutdown
type GracefulShutdownConfig struct {
	GracePeriod time.Duration  `json:"grace_period"`
	MaxWaitTime time.Duration  `json:"max_wait_time"`
	Logger      *logger.Logger `json:"-"`
}

// DefaultGracefulShutdownConfig returns default graceful shutdown configuration
func DefaultGracefulShutdownConfig() *GracefulShutdownConfig {
	return &GracefulShutdownConfig{
		GracePeriod: 30 * time.Second,
		MaxWaitTime: 60 * time.Second,
	}
}

// NewGracefulShutdownManager creates a new graceful shutdown manager
func NewGracefulShutdownManager(config *GracefulShutdownConfig) *GracefulShutdownManager {
	if config == nil {
		config = DefaultGracefulShutdownConfig()
	}

	return &GracefulShutdownManager{
		shutdownChan: make(chan struct{}),
		logger:       config.Logger,
	}
}

// Shutdown initiates graceful shutdown
func (g *GracefulShutdownManager) Shutdown(ctx context.Context, gracePeriod time.Duration) error {
	g.mu.Lock()
	if g.shuttingDown {
		g.mu.Unlock()
		return nil // Already shutting down
	}
	g.shuttingDown = true
	close(g.shutdownChan)
	g.mu.Unlock()

	g.logger.Info("Graceful shutdown initiated",
		zap.Duration("grace_period", gracePeriod))

	// Create a context with timeout for the grace period
	graceCtx, cancel := context.WithTimeout(ctx, gracePeriod)
	defer cancel()

	// Wait for active requests to complete or timeout
	done := make(chan struct{})
	go func() {
		g.activeRequests.Wait()
		close(done)
	}()

	select {
	case <-done:
		g.logger.Info("All active requests completed during graceful shutdown")
		return nil
	case <-graceCtx.Done():
		g.logger.Warn("Grace period expired, forcing shutdown",
			zap.Duration("grace_period", gracePeriod),
			zap.Int("remaining_requests", g.GetActiveRequests()))
		return graceCtx.Err()
	}
}

// IsShuttingDown returns true if shutdown is in progress
func (g *GracefulShutdownManager) IsShuttingDown() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.shuttingDown
}

// AddActiveRequest increments the active request counter
func (g *GracefulShutdownManager) AddActiveRequest() {
	g.activeRequests.Add(1)
}

// RemoveActiveRequest decrements the active request counter
func (g *GracefulShutdownManager) RemoveActiveRequest() {
	g.activeRequests.Done()
}

// GetActiveRequests returns the number of active requests
func (g *GracefulShutdownManager) GetActiveRequests() int {
	// Note: This is an approximation since sync.WaitGroup doesn't expose count
	// In a real implementation, you might want to use atomic operations
	return 0 // Placeholder
}

// WaitForShutdown waits for shutdown signal
func (g *GracefulShutdownManager) WaitForShutdown() <-chan struct{} {
	return g.shutdownChan
}

// ConnectionLifecycleManager manages the lifecycle of connections
type ConnectionLifecycleManager struct {
	gracefulShutdown *GracefulShutdownManager
	connections      map[string]*PooledConnection
	mu               sync.RWMutex
	logger           *logger.Logger
}

// NewConnectionLifecycleManager creates a new connection lifecycle manager
func NewConnectionLifecycleManager(logger *logger.Logger) *ConnectionLifecycleManager {
	return &ConnectionLifecycleManager{
		gracefulShutdown: NewGracefulShutdownManager(nil),
		connections:      make(map[string]*PooledConnection),
		logger:           logger,
	}
}

// AddConnection adds a connection to the lifecycle manager
func (clm *ConnectionLifecycleManager) AddConnection(conn *PooledConnection) {
	clm.mu.Lock()
	defer clm.mu.Unlock()
	clm.connections[conn.GetID()] = conn
}

// RemoveConnection removes a connection from the lifecycle manager
func (clm *ConnectionLifecycleManager) RemoveConnection(connID string) {
	clm.mu.Lock()
	defer clm.mu.Unlock()
	delete(clm.connections, connID)
}

// CloseAllConnections closes all managed connections
func (clm *ConnectionLifecycleManager) CloseAllConnections() error {
	clm.mu.RLock()
	connections := make([]*PooledConnection, 0, len(clm.connections))
	for _, conn := range clm.connections {
		connections = append(connections, conn)
	}
	clm.mu.RUnlock()

	var lastErr error
	for _, conn := range connections {
		if err := conn.Close(); err != nil {
			lastErr = err
			clm.logger.Error("Failed to close connection",
				zap.String("connection_id", conn.GetID()),
				zap.Error(err))
		}
	}

	return lastErr
}

// GetConnectionCount returns the number of managed connections
func (clm *ConnectionLifecycleManager) GetConnectionCount() int {
	clm.mu.RLock()
	defer clm.mu.RUnlock()
	return len(clm.connections)
}

// Shutdown gracefully shuts down all connections
func (clm *ConnectionLifecycleManager) Shutdown(ctx context.Context, gracePeriod time.Duration) error {
	clm.logger.Info("Starting connection lifecycle shutdown",
		zap.Duration("grace_period", gracePeriod),
		zap.Int("connection_count", clm.GetConnectionCount()))

	// Initiate graceful shutdown
	if err := clm.gracefulShutdown.Shutdown(ctx, gracePeriod); err != nil {
		return err
	}

	// Close all connections
	return clm.CloseAllConnections()
}

// IsShuttingDown returns true if shutdown is in progress
func (clm *ConnectionLifecycleManager) IsShuttingDown() bool {
	return clm.gracefulShutdown.IsShuttingDown()
}

// WaitGroupManager manages WaitGroups for graceful shutdown
type WaitGroupManager struct {
	wg     sync.WaitGroup
	logger *logger.Logger
}

// NewWaitGroupManager creates a new WaitGroup manager
func NewWaitGroupManager(logger *logger.Logger) *WaitGroupManager {
	return &WaitGroupManager{
		logger: logger,
	}
}

// Add adds a goroutine to the WaitGroup
func (wgm *WaitGroupManager) Add(delta int) {
	wgm.wg.Add(delta)
}

// Done marks a goroutine as completed
func (wgm *WaitGroupManager) Done() {
	wgm.wg.Done()
}

// Wait waits for all goroutines to complete
func (wgm *WaitGroupManager) Wait() {
	wgm.wg.Wait()
}

// WaitWithTimeout waits for all goroutines to complete with a timeout
func (wgm *WaitGroupManager) WaitWithTimeout(timeout time.Duration) error {
	done := make(chan struct{})
	go func() {
		wgm.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-time.After(timeout):
		return &TimeoutError{Message: "WaitGroup timeout exceeded"}
	}
}
