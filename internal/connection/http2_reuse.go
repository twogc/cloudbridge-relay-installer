package connection

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
	"golang.org/x/net/http2"
	"go.uber.org/zap"
)

// HTTP2ConnectionReuse управляет переиспользованием HTTP/2 соединений
// Адаптировано из cloudflared connection/http2.go
type HTTP2ConnectionReuse struct {
	connections map[string]*HTTP2ReusedConnection
	mu          sync.RWMutex
	logger      *logger.Logger
	
	// Конфигурация
	maxConnections int
	maxIdleTime    time.Duration
	cleanupInterval time.Duration
	
	// Контекст для cleanup
	ctx    context.Context
	cancel context.CancelFunc
}

// HTTP2ReusedConnection представляет переиспользуемое HTTP/2 соединение
type HTTP2ReusedConnection struct {
	conn        net.Conn
	client      *http.Client
	lastUsed    time.Time
	useCount    int64
	healthy     bool
	mu          sync.RWMutex
	closed      bool
}

// HTTP2ReuseConfig конфигурация для HTTP/2 connection reuse
type HTTP2ReuseConfig struct {
	MaxConnections  int           `json:"max_connections"`
	MaxIdleTime     time.Duration `json:"max_idle_time"`
	CleanupInterval time.Duration `json:"cleanup_interval"`
	Logger          *logger.Logger `json:"-"`
}

// DefaultHTTP2ReuseConfig возвращает конфигурацию по умолчанию
func DefaultHTTP2ReuseConfig() *HTTP2ReuseConfig {
	return &HTTP2ReuseConfig{
		MaxConnections:  100,
		MaxIdleTime:     90 * time.Second,
		CleanupInterval: 60 * time.Second,
	}
}

// NewHTTP2ConnectionReuse создает новый менеджер HTTP/2 connection reuse
func NewHTTP2ConnectionReuse(config *HTTP2ReuseConfig) *HTTP2ConnectionReuse {
	if config == nil {
		config = DefaultHTTP2ReuseConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())
	
	reuse := &HTTP2ConnectionReuse{
		connections:     make(map[string]*HTTP2ReusedConnection),
		logger:          config.Logger,
		maxConnections:  config.MaxConnections,
		maxIdleTime:     config.MaxIdleTime,
		cleanupInterval: config.CleanupInterval,
		ctx:             ctx,
		cancel:          cancel,
	}

	// Запускаем cleanup loop
	go reuse.cleanupLoop()

	return reuse
}

// GetConnection возвращает HTTP/2 соединение для указанного адреса
func (h *HTTP2ConnectionReuse) GetConnection(addr string) (*http.Client, error) {
	h.mu.RLock()
	conn, exists := h.connections[addr]
	h.mu.RUnlock()

	if exists && conn.IsHealthy() {
		conn.UpdateLastUsed()
		conn.IncrementUseCount()
		return conn.GetClient(), nil
	}

	// Создаем новое соединение
	return h.createConnection(addr)
}

// createConnection создает новое HTTP/2 соединение
func (h *HTTP2ConnectionReuse) createConnection(addr string) (*http.Client, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Проверяем лимит соединений
	if len(h.connections) >= h.maxConnections {
		if h.logger != nil {
			h.logger.Warn("Maximum HTTP/2 connections reached, removing oldest connection")
		}
		h.removeOldestConnection()
	}

	// Создаем HTTP/2 транспорт
	transport := &http2.Transport{
		AllowHTTP: true,
		DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
			return net.Dial(network, addr)
		},
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	// Создаем соединение
	conn := &HTTP2ReusedConnection{
		conn:     nil, // Будет установлено при первом использовании
		client:   client,
		lastUsed: time.Now(),
		useCount: 0,
		healthy:  true,
	}

	h.connections[addr] = conn
	if h.logger != nil {
		h.logger.Debug("Created new HTTP/2 connection",
			zap.String("addr", addr),
			zap.Int("total_connections", len(h.connections)))
	}

	return client, nil
}

// ReturnConnection возвращает соединение в пул
func (h *HTTP2ConnectionReuse) ReturnConnection(addr string, client *http.Client) {
	h.mu.RLock()
	conn, exists := h.connections[addr]
	h.mu.RUnlock()

	if exists {
		conn.UpdateLastUsed()
	}
}

// RemoveConnection удаляет соединение из пула
func (h *HTTP2ConnectionReuse) RemoveConnection(addr string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if conn, exists := h.connections[addr]; exists {
		conn.Close()
		delete(h.connections, addr)
		if h.logger != nil {
			h.logger.Debug("Removed HTTP/2 connection",
				zap.String("addr", addr),
				zap.Int("total_connections", len(h.connections)))
		}
	}
}

// removeOldestConnection удаляет самое старое соединение
func (h *HTTP2ConnectionReuse) removeOldestConnection() {
	var oldestAddr string
	var oldestTime time.Time

	for addr, conn := range h.connections {
		if oldestAddr == "" || conn.GetLastUsed().Before(oldestTime) {
			oldestAddr = addr
			oldestTime = conn.GetLastUsed()
		}
	}

	if oldestAddr != "" {
		h.RemoveConnection(oldestAddr)
	}
}

// cleanupLoop периодически очищает неиспользуемые соединения
func (h *HTTP2ConnectionReuse) cleanupLoop() {
	ticker := time.NewTicker(h.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-h.ctx.Done():
			return
		case <-ticker.C:
			h.cleanup()
		}
	}
}

// cleanup удаляет неиспользуемые соединения
func (h *HTTP2ConnectionReuse) cleanup() {
	h.mu.Lock()
	defer h.mu.Unlock()

	now := time.Now()
	removed := 0

	for addr, conn := range h.connections {
		if now.Sub(conn.GetLastUsed()) > h.maxIdleTime {
			conn.Close()
			delete(h.connections, addr)
			removed++
		}
	}

	if removed > 0 {
		if h.logger != nil {
			h.logger.Info("Cleaned up idle HTTP/2 connections",
				zap.Int("removed", removed),
				zap.Int("remaining", len(h.connections)))
		}
	}
}

// Close закрывает все соединения
func (h *HTTP2ConnectionReuse) Close() error {
	h.cancel()
	
	h.mu.Lock()
	defer h.mu.Unlock()

	for addr, conn := range h.connections {
		conn.Close()
		delete(h.connections, addr)
	}

	if h.logger != nil {
		h.logger.Info("Closed all HTTP/2 connections")
	}
	return nil
}

// GetStats возвращает статистику connection reuse
func (h *HTTP2ConnectionReuse) GetStats() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	totalUseCount := int64(0)
	healthyConnections := 0

	for _, conn := range h.connections {
		totalUseCount += conn.GetUseCount()
		if conn.IsHealthy() {
			healthyConnections++
		}
	}

	return map[string]interface{}{
		"total_connections":     len(h.connections),
		"healthy_connections":   healthyConnections,
		"total_use_count":       totalUseCount,
		"max_connections":       h.maxConnections,
		"max_idle_time":         h.maxIdleTime.String(),
		"cleanup_interval":      h.cleanupInterval.String(),
	}
}

// IsHealthy возвращает true если соединение здорово
func (hcr *HTTP2ReusedConnection) IsHealthy() bool {
	hcr.mu.RLock()
	defer hcr.mu.RUnlock()
	return hcr.healthy && !hcr.closed
}

// IsClosed возвращает true если соединение закрыто
func (hcr *HTTP2ReusedConnection) IsClosed() bool {
	hcr.mu.RLock()
	defer hcr.mu.RUnlock()
	return hcr.closed
}

// UpdateLastUsed обновляет время последнего использования
func (hcr *HTTP2ReusedConnection) UpdateLastUsed() {
	hcr.mu.Lock()
	defer hcr.mu.Unlock()
	hcr.lastUsed = time.Now()
}

// IncrementUseCount увеличивает счетчик использования
func (hcr *HTTP2ReusedConnection) IncrementUseCount() {
	hcr.mu.Lock()
	defer hcr.mu.Unlock()
	hcr.useCount++
}

// GetClient возвращает HTTP клиент
func (hcr *HTTP2ReusedConnection) GetClient() *http.Client {
	return hcr.client
}

// GetLastUsed возвращает время последнего использования
func (hcr *HTTP2ReusedConnection) GetLastUsed() time.Time {
	hcr.mu.RLock()
	defer hcr.mu.RUnlock()
	return hcr.lastUsed
}

// GetUseCount возвращает количество использований
func (hcr *HTTP2ReusedConnection) GetUseCount() int64 {
	hcr.mu.RLock()
	defer hcr.mu.RUnlock()
	return hcr.useCount
}

// Close закрывает соединение
func (hcr *HTTP2ReusedConnection) Close() error {
	hcr.mu.Lock()
	defer hcr.mu.Unlock()

	if hcr.closed {
		return nil
	}

	hcr.closed = true
	hcr.healthy = false

	if hcr.conn != nil {
		return hcr.conn.Close()
	}

	return nil
} 