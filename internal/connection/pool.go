package connection

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
	"go.uber.org/zap"
)

// PooledConnection represents a connection in the pool
type PooledConnection struct {
	conn       net.Conn
	id         string
	remoteAddr string
	created    time.Time
	lastUsed   time.Time
	useCount   int64
	healthy    bool
	mu         sync.RWMutex
	closed     bool
}

// ConnectionPool manages a pool of connections
type ConnectionPool struct {
	connections map[string]*PooledConnection
	balancer    LoadBalancer
	healthCheck HealthChecker
	metrics     *PoolMetrics
	logger      *logger.Logger
	mu          sync.RWMutex

	// Configuration
	maxConnections  int
	maxIdleTime     time.Duration
	cleanupInterval time.Duration

	// Context for cleanup
	ctx    context.Context
	cancel context.CancelFunc

	// New components for enhanced connection management
	retryStrategy    RetryStrategy
	gracefulShutdown *GracefulShutdownManager
	lifecycleManager *ConnectionLifecycleManager
	waitGroupManager *WaitGroupManager

	// HTTP/2 connection reuse
	http2Reuse *HTTP2ConnectionReuse
}

// LoadBalancer interface for connection load balancing
type LoadBalancer interface {
	Select(connections []*PooledConnection) *PooledConnection
	UpdateMetrics(conn *PooledConnection, success bool)
}

// HealthChecker interface for connection health checking
type HealthChecker interface {
	Check(conn *PooledConnection) bool
	GetHealthScore(conn *PooledConnection) float64
}

// PoolMetrics holds pool-specific metrics
type PoolMetrics struct {
	TotalConnections    int64
	ActiveConnections   int64
	IdleConnections     int64
	FailedConnections   int64
	ConnectionErrors    int64
	HealthCheckFailures int64
}

// RoundRobinBalancer implements round-robin load balancing
type RoundRobinBalancer struct {
	current int
	mu      sync.Mutex
}

// NewRoundRobinBalancer creates a new round-robin balancer
func NewRoundRobinBalancer() *RoundRobinBalancer {
	return &RoundRobinBalancer{}
}

// Select selects the next connection in round-robin fashion
func (rrb *RoundRobinBalancer) Select(connections []*PooledConnection) *PooledConnection {
	if len(connections) == 0 {
		return nil
	}

	rrb.mu.Lock()
	defer rrb.mu.Unlock()

	// Filter healthy connections
	healthy := make([]*PooledConnection, 0)
	for _, conn := range connections {
		if conn.IsHealthy() {
			healthy = append(healthy, conn)
		}
	}

	if len(healthy) == 0 {
		return nil
	}

	selected := healthy[rrb.current%len(healthy)]
	rrb.current++

	return selected
}

// UpdateMetrics updates metrics for the balancer
func (rrb *RoundRobinBalancer) UpdateMetrics(conn *PooledConnection, success bool) {
	// Round-robin doesn't use metrics for selection
}

// BasicHealthChecker implements basic health checking
type BasicHealthChecker struct {
	timeout time.Duration
}

// NewBasicHealthChecker creates a new basic health checker
func NewBasicHealthChecker(timeout time.Duration) *BasicHealthChecker {
	return &BasicHealthChecker{
		timeout: timeout,
	}
}

// Check performs a basic health check on the connection
func (bhc *BasicHealthChecker) Check(conn *PooledConnection) bool {
	if conn == nil || conn.IsClosed() {
		return false
	}

	// Try to set a deadline and read a byte
	if err := conn.conn.SetReadDeadline(time.Now().Add(bhc.timeout)); err != nil {
		return false
	}

	// For TCP connections, we can't easily check health without sending data
	// So we'll just check if the connection is not closed
	return !conn.IsClosed()
}

// GetHealthScore returns a health score for the connection
func (bhc *BasicHealthChecker) GetHealthScore(conn *PooledConnection) float64 {
	if conn == nil || conn.IsClosed() {
		return 0.0
	}

	// Simple health score based on last usage and use count
	age := time.Since(conn.created)
	if age > 24*time.Hour {
		return 0.5 // Older connections get lower score
	}

	// Higher use count indicates more reliable connection
	score := 1.0
	if conn.useCount > 100 {
		score = 1.0
	} else if conn.useCount > 10 {
		score = 0.8
	} else {
		score = 0.6
	}

	return score
}

// NewConnectionPool creates a new connection pool
func NewConnectionPool(
	maxConnections int,
	maxIdleTime time.Duration,
	cleanupInterval time.Duration,
	logger *logger.Logger,
) *ConnectionPool {
	ctx, cancel := context.WithCancel(context.Background())

	// Create retry strategy
	retryConfig := DefaultRetryConfig()
	retryConfig.Logger = logger
	retryStrategy := NewBackoffHandler(retryConfig)

	// Create graceful shutdown manager
	gracefulConfig := DefaultGracefulShutdownConfig()
	gracefulConfig.Logger = logger
	gracefulShutdown := NewGracefulShutdownManager(gracefulConfig)

	// Create lifecycle manager
	lifecycleManager := NewConnectionLifecycleManager(logger)

	// Create wait group manager
	waitGroupManager := NewWaitGroupManager(logger)

	// Create HTTP/2 connection reuse
	http2Config := DefaultHTTP2ReuseConfig()
	http2Config.Logger = logger
	http2Reuse := NewHTTP2ConnectionReuse(http2Config)

	pool := &ConnectionPool{
		connections:      make(map[string]*PooledConnection),
		balancer:         NewRoundRobinBalancer(),
		healthCheck:      NewBasicHealthChecker(5 * time.Second),
		metrics:          &PoolMetrics{},
		logger:           logger,
		maxConnections:   maxConnections,
		maxIdleTime:      maxIdleTime,
		cleanupInterval:  cleanupInterval,
		ctx:              ctx,
		cancel:           cancel,
		retryStrategy:    retryStrategy,
		gracefulShutdown: gracefulShutdown,
		lifecycleManager: lifecycleManager,
		waitGroupManager: waitGroupManager,
		http2Reuse:       http2Reuse,
	}

	// Start cleanup goroutine
	go pool.cleanupLoop()

	return pool
}

// GetConnection gets a connection from the pool with retry logic
func (cp *ConnectionPool) GetConnection(remoteAddr string) (*PooledConnection, error) {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	// Check if shutdown is in progress
	if cp.gracefulShutdown.IsShuttingDown() {
		return nil, &ConnectionError{Message: "Connection pool is shutting down", Retryable: false}
	}

	// Check if we have an existing connection
	if conn, exists := cp.connections[remoteAddr]; exists && conn.IsHealthy() {
		conn.UpdateLastUsed()
		conn.IncrementUseCount()
		cp.metrics.ActiveConnections++
		return conn, nil
	}

	// Create new connection if we haven't reached the limit
	if len(cp.connections) >= cp.maxConnections {
		// Try to find a connection using load balancer
		connections := cp.getHealthyConnections()
		if selected := cp.balancer.Select(connections); selected != nil {
			selected.UpdateLastUsed()
			selected.IncrementUseCount()
			cp.metrics.ActiveConnections++
			return selected, nil
		}

		return nil, fmt.Errorf("connection pool is full")
	}

	// Create new connection
	conn, err := cp.createConnection(remoteAddr)
	if err != nil {
		cp.metrics.ConnectionErrors++
		return nil, err
	}

	cp.connections[remoteAddr] = conn
	cp.metrics.TotalConnections++
	cp.metrics.ActiveConnections++

	cp.logger.Info("Created new connection",
		zap.String("remote_addr", remoteAddr),
		zap.String("connection_id", conn.id),
	)

	return conn, nil
}

// createConnection creates a new connection
func (cp *ConnectionPool) createConnection(remoteAddr string) (*PooledConnection, error) {
	conn, err := net.DialTimeout("tcp", remoteAddr, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection: %v", err)
	}

	pooledConn := &PooledConnection{
		conn:       conn,
		id:         fmt.Sprintf("%s-%d", remoteAddr, time.Now().UnixNano()),
		remoteAddr: remoteAddr,
		created:    time.Now(),
		lastUsed:   time.Now(),
		useCount:   0,
		healthy:    true,
	}

	return pooledConn, nil
}

// ReturnConnection returns a connection to the pool
func (cp *ConnectionPool) ReturnConnection(conn *PooledConnection) {
	if conn == nil {
		return
	}

	cp.mu.Lock()
	defer cp.mu.Unlock()

	// Update metrics
	cp.metrics.ActiveConnections--
	cp.metrics.IdleConnections++

	// Update last used time
	conn.UpdateLastUsed()

	cp.logger.Debug("Returned connection to pool",
		zap.String("connection_id", conn.id),
		zap.String("remote_addr", conn.remoteAddr),
	)
}

// RemoveConnection removes a connection from the pool
func (cp *ConnectionPool) RemoveConnection(conn *PooledConnection) {
	if conn == nil {
		return
	}

	cp.mu.Lock()
	defer cp.mu.Unlock()

	// Close the connection
	if err := conn.Close(); err != nil {
		cp.logger.Error("Error closing connection during removal", zap.Error(err))
	}

	// Remove from pool
	delete(cp.connections, conn.remoteAddr)

	// Update metrics
	cp.metrics.TotalConnections--
	if conn.IsHealthy() {
		cp.metrics.IdleConnections--
	} else {
		cp.metrics.FailedConnections++
	}

	cp.logger.Info("Removed connection from pool",
		zap.String("connection_id", conn.id),
		zap.String("remote_addr", conn.remoteAddr),
	)
}

// getHealthyConnections returns all healthy connections
func (cp *ConnectionPool) getHealthyConnections() []*PooledConnection {
	healthy := make([]*PooledConnection, 0)
	for _, conn := range cp.connections {
		if conn.IsHealthy() {
			healthy = append(healthy, conn)
		}
	}
	return healthy
}

// cleanupLoop periodically cleans up old connections
func (cp *ConnectionPool) cleanupLoop() {
	ticker := time.NewTicker(cp.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-cp.ctx.Done():
			return
		case <-ticker.C:
			cp.cleanup()
		}
	}
}

// cleanup removes old and unhealthy connections
func (cp *ConnectionPool) cleanup() {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	cutoff := time.Now().Add(-cp.maxIdleTime)
	removed := 0

	for _, conn := range cp.connections {
		// Remove connections that are too old or unhealthy
		if conn.lastUsed.Before(cutoff) || !cp.healthCheck.Check(conn) {
			cp.RemoveConnection(conn)
			removed++
		}
	}

	if removed > 0 {
		cp.logger.Info("Cleaned up connections",
			zap.Int("removed", removed),
			zap.Int("remaining", len(cp.connections)),
		)
	}
}

// GetStats returns pool statistics
func (cp *ConnectionPool) GetStats() map[string]interface{} {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	return map[string]interface{}{
		"total_connections":     cp.metrics.TotalConnections,
		"active_connections":    cp.metrics.ActiveConnections,
		"idle_connections":      cp.metrics.IdleConnections,
		"failed_connections":    cp.metrics.FailedConnections,
		"connection_errors":     cp.metrics.ConnectionErrors,
		"health_check_failures": cp.metrics.HealthCheckFailures,
		"pool_size":             len(cp.connections),
		"max_connections":       cp.maxConnections,
	}
}

// Close closes the connection pool
func (cp *ConnectionPool) Close() error {
	cp.cancel()

	cp.mu.Lock()
	defer cp.mu.Unlock()

	// Close all connections
	for _, conn := range cp.connections {
		if err := conn.Close(); err != nil {
			cp.logger.Error("Error closing connection in pool", zap.Error(err))
		}
	}

	cp.connections = make(map[string]*PooledConnection)

	// Close HTTP/2 connection reuse
	if cp.http2Reuse != nil {
		if err := cp.http2Reuse.Close(); err != nil {
			cp.logger.Error("Error closing HTTP/2 connection reuse", zap.Error(err))
		}
	}

	cp.logger.Info("Connection pool closed")
	return nil
}

// PooledConnection methods

// IsHealthy returns true if the connection is healthy
func (pc *PooledConnection) IsHealthy() bool {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	return pc.healthy && !pc.closed
}

// IsClosed returns true if the connection is closed
func (pc *PooledConnection) IsClosed() bool {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	return pc.closed
}

// UpdateLastUsed updates the last used time
func (pc *PooledConnection) UpdateLastUsed() {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.lastUsed = time.Now()
}

// IncrementUseCount increments the use count
func (pc *PooledConnection) IncrementUseCount() {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.useCount++
}

// Close closes the connection
func (pc *PooledConnection) Close() error {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	if pc.closed {
		return nil
	}

	pc.closed = true
	pc.healthy = false

	if pc.conn != nil {
		return pc.conn.Close()
	}

	return nil
}

// GetConn returns the underlying connection
func (pc *PooledConnection) GetConn() net.Conn {
	return pc.conn
}

// GetID returns the connection ID
func (pc *PooledConnection) GetID() string {
	return pc.id
}

// GetRemoteAddr returns the remote address
func (pc *PooledConnection) GetRemoteAddr() string {
	return pc.remoteAddr
}

// GracefulShutdown performs graceful shutdown of the connection pool
func (cp *ConnectionPool) GracefulShutdown(ctx context.Context, gracePeriod time.Duration) error {
	cp.logger.Info("Starting graceful shutdown of connection pool",
		zap.Duration("grace_period", gracePeriod),
		zap.Int("connection_count", len(cp.connections)))

	return cp.lifecycleManager.Shutdown(ctx, gracePeriod)
}

// IsShuttingDown returns true if the pool is shutting down
func (cp *ConnectionPool) IsShuttingDown() bool {
	return cp.gracefulShutdown.IsShuttingDown()
}

// GetConnectionWithRetry gets a connection with retry logic
func (cp *ConnectionPool) GetConnectionWithRetry(remoteAddr string, maxAttempts int) (*PooledConnection, error) {
	var lastErr error

	for attempt := 0; attempt < maxAttempts; attempt++ {
		conn, err := cp.GetConnection(remoteAddr)
		if err == nil {
			return conn, nil
		}

		lastErr = err

		// Check if we should retry
		if !cp.retryStrategy.ShouldRetry(err, attempt) {
			break
		}

		// Get backoff duration
		backoffDuration := cp.retryStrategy.GetBackoffDuration(attempt)
		if backoffDuration == 0 {
			break
		}

		cp.logger.Debug("Retrying connection",
			zap.String("remote_addr", remoteAddr),
			zap.Int("attempt", attempt+1),
			zap.Duration("backoff", backoffDuration),
			zap.Error(err))

		// Wait before retry
		select {
		case <-time.After(backoffDuration):
			continue
		case <-cp.ctx.Done():
			return nil, cp.ctx.Err()
		}
	}

	return nil, lastErr
}

// AddActiveRequest increments the active request counter
func (cp *ConnectionPool) AddActiveRequest() {
	cp.gracefulShutdown.AddActiveRequest()
	cp.waitGroupManager.Add(1)
}

// RemoveActiveRequest decrements the active request counter
func (cp *ConnectionPool) RemoveActiveRequest() {
	cp.gracefulShutdown.RemoveActiveRequest()
	cp.waitGroupManager.Done()
}

// WaitForActiveRequests waits for all active requests to complete
func (cp *ConnectionPool) WaitForActiveRequests(timeout time.Duration) error {
	return cp.waitGroupManager.WaitWithTimeout(timeout)
}

// GetRetryStats returns retry statistics
func (cp *ConnectionPool) GetRetryStats() map[string]interface{} {
	if backoffHandler, ok := cp.retryStrategy.(*BackoffHandler); ok {
		return map[string]interface{}{
			"retries":     backoffHandler.Retries(),
			"max_retries": backoffHandler.GetMaxAttempts(),
			"reached_max": backoffHandler.ReachedMaxRetries(),
		}
	}
	return map[string]interface{}{}
}

// GetHTTP2Connection returns an HTTP/2 connection for the specified address
func (cp *ConnectionPool) GetHTTP2Connection(addr string) (*http.Client, error) {
	return cp.http2Reuse.GetConnection(addr)
}

// ReturnHTTP2Connection returns an HTTP/2 connection to the pool
func (cp *ConnectionPool) ReturnHTTP2Connection(addr string, client *http.Client) {
	cp.http2Reuse.ReturnConnection(addr, client)
}

// GetHTTP2Stats returns HTTP/2 connection reuse statistics
func (cp *ConnectionPool) GetHTTP2Stats() map[string]interface{} {
	return cp.http2Reuse.GetStats()
}
