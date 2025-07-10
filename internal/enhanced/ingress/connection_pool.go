package ingress

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

// ConnectionPool представляет пул соединений
type ConnectionPool struct {
	config  ConnectionPoolConfig
	logger  *logger.Logger
	metrics *metrics.Metrics
	
	// Соединения по адресу
	connections map[string]*PooledConnection
	mu          sync.RWMutex
	
	// Контекст для graceful shutdown
	ctx    context.Context
	cancel context.CancelFunc
	
	// Статистика
	stats *PoolStats
}

// PooledConnection представляет соединение в пуле
type PooledConnection struct {
	conn       net.Conn
	addr       string
	created    time.Time
	lastUsed   time.Time
	useCount   int64
	healthy    bool
	mu         sync.RWMutex
	closed     bool
}

// PoolStats представляет статистику пула
type PoolStats struct {
	TotalConnections   int64
	ActiveConnections  int64
	IdleConnections    int64
	FailedConnections  int64
	ConnectionErrors   int64
	mu                 sync.RWMutex
}

// NewConnectionPool создает новый пул соединений
func NewConnectionPool(config ConnectionPoolConfig, logger *logger.Logger, metrics *metrics.Metrics) *ConnectionPool {
	ctx, cancel := context.WithCancel(context.Background())
	
	pool := &ConnectionPool{
		config:     config,
		logger:     logger,
		metrics:    metrics,
		connections: make(map[string]*PooledConnection),
		ctx:        ctx,
		cancel:     cancel,
		stats:      &PoolStats{},
	}
	
	// Запускаем cleanup goroutine
	go pool.cleanupLoop()
	
	logger.Info("Connection pool created",
		zap.Int("max_size", config.MaxSize),
		zap.Duration("idle_timeout", config.IdleTimeout),
		zap.Int("max_idle", config.MaxIdle),
	)
	
	return pool
}

// GetConnection получает соединение из пула
func (p *ConnectionPool) GetConnection(addr string) (net.Conn, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	// Проверяем существующее соединение
	if conn, exists := p.connections[addr]; exists && conn.IsHealthy() {
		conn.UpdateLastUsed()
		conn.IncrementUseCount()
		p.stats.IncActiveConnections()
		return conn.GetConn(), nil
	}
	
	// Проверяем лимит соединений
	if len(p.connections) >= p.config.MaxSize {
		// Удаляем самое старое неиспользуемое соединение
		p.removeOldestConnection()
	}
	
	// Создаем новое соединение
	conn, err := p.createConnection(addr)
	if err != nil {
		p.stats.IncConnectionErrors()
		return nil, fmt.Errorf("failed to create connection: %v", err)
	}
	
	p.connections[addr] = conn
	p.stats.IncTotalConnections()
	p.stats.IncActiveConnections()
	
	p.logger.Debug("Created new connection",
		zap.String("addr", addr),
		zap.String("connection_id", conn.GetID()),
	)
	
	return conn.GetConn(), nil
}

// ReturnConnection возвращает соединение в пул
func (p *ConnectionPool) ReturnConnection(conn net.Conn) {
	// Ищем соединение в пуле по адресу
	p.mu.Lock()
	defer p.mu.Unlock()
	
	for addr, pooledConn := range p.connections {
		if pooledConn.GetConn() == conn {
			pooledConn.UpdateLastUsed()
			p.stats.DecActiveConnections()
			p.stats.IncIdleConnections()
			
			p.logger.Debug("Connection returned to pool",
				zap.String("addr", addr),
				zap.String("connection_id", pooledConn.GetID()),
			)
			break
		}
	}
}

// createConnection создает новое соединение
func (p *ConnectionPool) createConnection(addr string) (*PooledConnection, error) {
	// Устанавливаем таймаут для создания соединения
	dialer := &net.Dialer{
		Timeout: 30 * time.Second,
	}
	
	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	
	pooledConn := &PooledConnection{
		conn:     conn,
		addr:     addr,
		created:  time.Now(),
		lastUsed: time.Now(),
		useCount: 0,
		healthy:  true,
	}
	
	return pooledConn, nil
}

// removeOldestConnection удаляет самое старое неиспользуемое соединение
func (p *ConnectionPool) removeOldestConnection() {
	var oldestConn *PooledConnection
	var oldestAddr string
	var oldestTime time.Time
	
	for addr, conn := range p.connections {
		if !conn.IsHealthy() {
			// Удаляем нездоровое соединение
			delete(p.connections, addr)
			conn.Close()
			p.stats.DecTotalConnections()
			return
		}
		
		lastUsed := conn.GetLastUsed()
		if oldestConn == nil || lastUsed.Before(oldestTime) {
			oldestConn = conn
			oldestAddr = addr
			oldestTime = lastUsed
		}
	}
	
	if oldestConn != nil {
		delete(p.connections, oldestAddr)
		oldestConn.Close()
		p.stats.DecTotalConnections()
		
		p.logger.Debug("Removed oldest connection",
			zap.String("addr", oldestAddr),
			zap.String("connection_id", oldestConn.GetID()),
		)
	}
}

// cleanupLoop периодически очищает неиспользуемые соединения
func (p *ConnectionPool) cleanupLoop() {
	ticker := time.NewTicker(p.config.IdleTimeout / 2)
	defer ticker.Stop()
	
	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			p.cleanup()
		}
	}
}

// cleanup очищает неиспользуемые соединения
func (p *ConnectionPool) cleanup() {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	now := time.Now()
	removed := 0
	
	for addr, conn := range p.connections {
		// Удаляем соединения, которые не использовались дольше idle_timeout
		if now.Sub(conn.GetLastUsed()) > p.config.IdleTimeout {
			delete(p.connections, addr)
			conn.Close()
			p.stats.DecTotalConnections()
			removed++
		}
	}
	
	if removed > 0 {
		p.logger.Info("Cleaned up idle connections",
			zap.Int("removed", removed),
			zap.Int("remaining", len(p.connections)),
		)
	}
}

// Shutdown корректно завершает работу пула
func (p *ConnectionPool) Shutdown(ctx context.Context) error {
	p.logger.Info("Shutting down connection pool")
	
	// Отменяем контекст
	p.cancel()
	
	// Закрываем все соединения
	p.mu.Lock()
	defer p.mu.Unlock()
	
	for addr, conn := range p.connections {
		conn.Close()
		delete(p.connections, addr)
	}
	
	p.logger.Info("Connection pool shutdown completed",
		zap.Int("closed_connections", len(p.connections)),
	)
	
	return nil
}

// GetStats возвращает статистику пула
func (p *ConnectionPool) GetStats() map[string]interface{} {
	p.stats.mu.RLock()
	defer p.stats.mu.RUnlock()
	
	return map[string]interface{}{
		"total_connections":   p.stats.TotalConnections,
		"active_connections":  p.stats.ActiveConnections,
		"idle_connections":    p.stats.IdleConnections,
		"failed_connections":  p.stats.FailedConnections,
		"connection_errors":   p.stats.ConnectionErrors,
		"current_connections": len(p.connections),
		"max_size":           p.config.MaxSize,
	}
}

// Методы PooledConnection

// GetConn возвращает underlying соединение
func (pc *PooledConnection) GetConn() net.Conn {
	return pc.conn
}

// GetAddr возвращает адрес соединения
func (pc *PooledConnection) GetAddr() string {
	return pc.addr
}

// GetID возвращает ID соединения
func (pc *PooledConnection) GetID() string {
	return fmt.Sprintf("%s-%d", pc.addr, pc.created.UnixNano())
}

// GetLastUsed возвращает время последнего использования
func (pc *PooledConnection) GetLastUsed() time.Time {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	return pc.lastUsed
}

// UpdateLastUsed обновляет время последнего использования
func (pc *PooledConnection) UpdateLastUsed() {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.lastUsed = time.Now()
}

// IncrementUseCount увеличивает счетчик использования
func (pc *PooledConnection) IncrementUseCount() {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.useCount++
}

// IsHealthy проверяет, здорово ли соединение
func (pc *PooledConnection) IsHealthy() bool {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	
	if pc.closed {
		return false
	}
	
	// Простая проверка здоровья - пытаемся записать пустой байт
	if pc.conn != nil {
		if err := pc.conn.SetWriteDeadline(time.Now().Add(100 * time.Millisecond)); err != nil {
			return false
		}
		
		if _, err := pc.conn.Write([]byte{}); err != nil {
			return false
		}
	}
	
	return true
}

// Close закрывает соединение
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

// Методы PoolStats

// IncTotalConnections увеличивает счетчик общих соединений
func (ps *PoolStats) IncTotalConnections() {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.TotalConnections++
}

// DecTotalConnections уменьшает счетчик общих соединений
func (ps *PoolStats) DecTotalConnections() {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	if ps.TotalConnections > 0 {
		ps.TotalConnections--
	}
}

// IncActiveConnections увеличивает счетчик активных соединений
func (ps *PoolStats) IncActiveConnections() {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.ActiveConnections++
}

// DecActiveConnections уменьшает счетчик активных соединений
func (ps *PoolStats) DecActiveConnections() {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	if ps.ActiveConnections > 0 {
		ps.ActiveConnections--
	}
}

// IncIdleConnections увеличивает счетчик неактивных соединений
func (ps *PoolStats) IncIdleConnections() {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.IdleConnections++
}

// IncFailedConnections увеличивает счетчик неудачных соединений
func (ps *PoolStats) IncFailedConnections() {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.FailedConnections++
}

// IncConnectionErrors увеличивает счетчик ошибок соединений
func (ps *PoolStats) IncConnectionErrors() {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.ConnectionErrors++
} 