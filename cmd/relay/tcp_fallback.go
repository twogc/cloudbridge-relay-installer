package main

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/twogc/cloudbridge-relay-installer/config"
	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
)

// TCPFallback представляет TCP fallback для блокирующих сетей
type TCPFallback struct {
	logger      *logger.Logger
	config      *config.Config
	listener    net.Listener
	connections map[string]*TCPConnection
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
}

// TCPConnection представляет TCP соединение
type TCPConnection struct {
	ID           string
	Conn         net.Conn
	RemoteAddr   string
	CreatedAt    time.Time
	LastActivity time.Time
	IsActive     bool
}

// NewTCPFallback создает новый TCP fallback
func NewTCPFallback(logger *logger.Logger, config *config.Config) *TCPFallback {
	ctx, cancel := context.WithCancel(context.Background())

	return &TCPFallback{
		logger:      logger,
		config:      config,
		connections: make(map[string]*TCPConnection),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start запускает TCP fallback сервер
func (tcpf *TCPFallback) Start() error {
	if !tcpf.config.TCPFallback.Enabled {
		tcpf.logger.Info("TCP fallback is disabled")
		return nil
	}

	tcpf.logger.Info("Starting TCP fallback server",
		zap.String("listen_addr", tcpf.config.TCPFallback.ListenAddr))

	// Создаем TCP listener
	listener, err := net.Listen("tcp", tcpf.config.TCPFallback.ListenAddr)
	if err != nil {
		return fmt.Errorf("failed to start TCP listener: %v", err)
	}

	tcpf.listener = listener

	// Запускаем обработку соединений
	go tcpf.handleConnections()

	tcpf.logger.Info("TCP fallback server started successfully")
	return nil
}

// handleConnections обрабатывает входящие TCP соединения
func (tcpf *TCPFallback) handleConnections() {
	for {
		select {
		case <-tcpf.ctx.Done():
			return
		default:
			conn, err := tcpf.listener.Accept()
			if err != nil {
				if tcpf.ctx.Err() != nil {
					return
				}
				tcpf.logger.Error("Failed to accept TCP connection", zap.Error(err))
				continue
			}

			go tcpf.handleTCPConnection(conn)
		}
	}
}

// handleTCPConnection обрабатывает TCP соединение
func (tcpf *TCPFallback) handleTCPConnection(conn net.Conn) {
	defer conn.Close()

	// Создаем TCP соединение
	tcpConn := &TCPConnection{
		ID:           fmt.Sprintf("tcp_%d", time.Now().UnixNano()),
		Conn:         conn,
		RemoteAddr:   conn.RemoteAddr().String(),
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
		IsActive:     true,
	}

	// Сохраняем соединение
	tcpf.mu.Lock()
	tcpf.connections[tcpConn.ID] = tcpConn
	tcpf.mu.Unlock()

	tcpf.logger.Info("TCP connection established",
		zap.String("conn_id", tcpConn.ID),
		zap.String("remote_addr", tcpConn.RemoteAddr))

	// Обрабатываем соединение
	tcpf.processTCPConnection(tcpConn)

	// Удаляем соединение
	tcpf.removeConnection(tcpConn.ID)
}

// processTCPConnection обрабатывает TCP соединение
func (tcpf *TCPFallback) processTCPConnection(conn *TCPConnection) {
	defer func() {
		conn.IsActive = false
		conn.Conn.Close()
	}()

	// Устанавливаем таймауты
	conn.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

	buffer := make([]byte, 4096)
	for {
		select {
		case <-tcpf.ctx.Done():
			return
		default:
			// Читаем данные
			n, err := conn.Conn.Read(buffer)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					// Таймаут - продолжаем
					continue
				}
				tcpf.logger.Debug("TCP read error",
					zap.String("conn_id", conn.ID),
					zap.Error(err))
				return
			}

			// Обновляем активность
			conn.LastActivity = time.Now()

			// Обрабатываем данные
			if err := tcpf.processTCPData(conn, buffer[:n]); err != nil {
				tcpf.logger.Error("Failed to process TCP data",
					zap.String("conn_id", conn.ID),
					zap.Error(err))
				return
			}
		}
	}
}

// processTCPData обрабатывает TCP данные
func (tcpf *TCPFallback) processTCPData(conn *TCPConnection, data []byte) error {
	tcpf.logger.Debug("Processing TCP data",
		zap.String("conn_id", conn.ID),
		zap.Int("data_len", len(data)))

	// Здесь можно добавить логику обработки различных протоколов
	// Например, HTTP, MASQUE, или кастомные протоколы

	// Проверяем, является ли это HTTP запросом
	if tcpf.isHTTPRequest(data) {
		return tcpf.handleHTTPOverTCP(conn, data)
	}

	// Проверяем, является ли это MASQUE запросом
	if tcpf.isMASQUERequest(data) {
		return tcpf.handleMASQUEOverTCP(conn, data)
	}

	// Для остальных данных просто эхо-ответ
	response := fmt.Sprintf("TCP Echo: %d bytes received", len(data))
	_, err := conn.Conn.Write([]byte(response))
	return err
}

// isHTTPRequest проверяет, является ли данные HTTP запросом
func (tcpf *TCPFallback) isHTTPRequest(data []byte) bool {
	if len(data) < 4 {
		return false
	}
	
	// Проверяем HTTP методы
	httpMethods := []string{"GET ", "POST", "PUT ", "HEAD", "DELE", "OPTI", "PATC"}
	for _, method := range httpMethods {
		if len(data) >= len(method) && string(data[:len(method)]) == method {
			return true
		}
	}
	
	return false
}

// isMASQUERequest проверяет, является ли данные MASQUE запросом
func (tcpf *TCPFallback) isMASQUERequest(data []byte) bool {
	// Проверяем на CONNECT запросы к MASQUE endpoints
	dataStr := string(data)
	return contains(dataStr, "CONNECT") && 
		   (contains(dataStr, "/.well-known/masque/") || 
		    contains(dataStr, "masque"))
}

// contains проверяет, содержит ли строка подстроку
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}

// handleHTTPOverTCP обрабатывает HTTP запросы через TCP
func (tcpf *TCPFallback) handleHTTPOverTCP(conn *TCPConnection, data []byte) error {
	tcpf.logger.Info("HTTP request over TCP",
		zap.String("conn_id", conn.ID),
		zap.String("remote_addr", conn.RemoteAddr))

	// Создаем простой HTTP ответ
	response := `HTTP/1.1 200 OK
Content-Type: text/plain
Content-Length: 25
Connection: close

TCP Fallback Active
`

	_, err := conn.Conn.Write([]byte(response))
	return err
}

// handleMASQUEOverTCP обрабатывает MASQUE запросы через TCP
func (tcpf *TCPFallback) handleMASQUEOverTCP(conn *TCPConnection, data []byte) error {
	tcpf.logger.Info("MASQUE request over TCP",
		zap.String("conn_id", conn.ID),
		zap.String("remote_addr", conn.RemoteAddr))

	// Создаем HTTP ответ для MASQUE
	response := `HTTP/1.1 200 OK
Capsule-Protocol: ?1
Content-Type: application/octet-stream
Connection: close

`

	_, err := conn.Conn.Write([]byte(response))
	if err != nil {
		return err
	}

	// Здесь можно добавить логику обработки MASQUE протокола через TCP
	// Например, туннелирование UDP/IP трафика

	return nil
}

// removeConnection удаляет соединение из списка
func (tcpf *TCPFallback) removeConnection(connID string) {
	tcpf.mu.Lock()
	defer tcpf.mu.Unlock()
	
	delete(tcpf.connections, connID)
	tcpf.logger.Debug("TCP connection removed", zap.String("conn_id", connID))
}

// GetConnectionCount возвращает количество активных соединений
func (tcpf *TCPFallback) GetConnectionCount() int {
	tcpf.mu.RLock()
	defer tcpf.mu.RUnlock()
	
	return len(tcpf.connections)
}

// GetConnectionInfo возвращает информацию о соединениях
func (tcpf *TCPFallback) GetConnectionInfo() map[string]interface{} {
	tcpf.mu.RLock()
	defer tcpf.mu.RUnlock()

	connections := make([]map[string]interface{}, 0, len(tcpf.connections))
	for connID, conn := range tcpf.connections {
		connections = append(connections, map[string]interface{}{
			"id":            connID,
			"remote_addr":   conn.RemoteAddr,
			"created_at":    conn.CreatedAt,
			"last_activity": conn.LastActivity,
			"is_active":     conn.IsActive,
		})
	}

	return map[string]interface{}{
		"total_connections": len(tcpf.connections),
		"connections":       connections,
	}
}

// Close закрывает TCP fallback
func (tcpf *TCPFallback) Close() error {
	tcpf.logger.Info("Closing TCP fallback")

	tcpf.cancel()

	// Закрываем listener
	if tcpf.listener != nil {
		if err := tcpf.listener.Close(); err != nil {
			tcpf.logger.Warn("Failed to close TCP listener", zap.Error(err))
		}
	}

	// Закрываем все соединения
	tcpf.mu.Lock()
	for connID, conn := range tcpf.connections {
		conn.Conn.Close()
		tcpf.logger.Debug("Closed TCP connection", zap.String("conn_id", connID))
	}
	tcpf.connections = make(map[string]*TCPConnection)
	tcpf.mu.Unlock()

	tcpf.logger.Info("TCP fallback closed successfully")
	return nil
}
