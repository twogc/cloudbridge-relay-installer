package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"github.com/twogc/cloudbridge-relay-installer/config"
	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
)

// WebSocketFallback представляет WebSocket fallback для блокирующих сетей
type WebSocketFallback struct {
	logger      *logger.Logger
	config      *config.Config
	upgrader    websocket.Upgrader
	connections map[string]*WebSocketConnection
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
}

// WebSocketConnection представляет WebSocket соединение
type WebSocketConnection struct {
	ID           string
	Conn         *websocket.Conn
	RemoteAddr   string
	CreatedAt    time.Time
	LastActivity time.Time
	IsActive     bool
}

// NewWebSocketFallback создает новый WebSocket fallback
func NewWebSocketFallback(logger *logger.Logger, config *config.Config) *WebSocketFallback {
	ctx, cancel := context.WithCancel(context.Background())

	return &WebSocketFallback{
		logger:      logger,
		config:      config,
		connections: make(map[string]*WebSocketConnection),
		ctx:         ctx,
		cancel:      cancel,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// В продакшене следует добавить проверку origin
				return true
			},
			ReadBufferSize:  4096,
			WriteBufferSize: 4096,
		},
	}
}

// Start запускает WebSocket fallback сервер
func (wsf *WebSocketFallback) Start() error {
	if !wsf.config.WebSocketFallback.Enabled {
		wsf.logger.Info("WebSocket fallback is disabled")
		return nil
	}

	wsf.logger.Info("Starting WebSocket fallback server",
		zap.String("listen_addr", wsf.config.WebSocketFallback.ListenAddr))

	// Создаем HTTP сервер для WebSocket
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", wsf.handleWebSocket)
	mux.HandleFunc("/ws/masque", wsf.handleMASQUEWebSocket)
	mux.HandleFunc("/health", wsf.handleHealth)

	server := &http.Server{
		Addr:    wsf.config.WebSocketFallback.ListenAddr,
		Handler: mux,
	}

	// Запускаем сервер в горутине
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			wsf.logger.Error("WebSocket fallback server error", zap.Error(err))
		}
	}()

	wsf.logger.Info("WebSocket fallback server started successfully")
	return nil
}

// handleWebSocket обрабатывает обычные WebSocket соединения
func (wsf *WebSocketFallback) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	wsf.logger.Info("WebSocket connection request",
		zap.String("remote_addr", r.RemoteAddr),
		zap.String("user_agent", r.UserAgent()))

	// Обновляем соединение до WebSocket
	conn, err := wsf.upgrader.Upgrade(w, r, nil)
	if err != nil {
		wsf.logger.Error("Failed to upgrade to WebSocket", zap.Error(err))
		return
	}

	// Создаем WebSocket соединение
	wsConn := &WebSocketConnection{
		ID:           fmt.Sprintf("ws_%d", time.Now().UnixNano()),
		Conn:         conn,
		RemoteAddr:   r.RemoteAddr,
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
		IsActive:     true,
	}

	// Сохраняем соединение
	wsf.mu.Lock()
	wsf.connections[wsConn.ID] = wsConn
	wsf.mu.Unlock()

	wsf.logger.Info("WebSocket connection established",
		zap.String("conn_id", wsConn.ID),
		zap.String("remote_addr", wsConn.RemoteAddr))

	// Обрабатываем соединение
	wsf.handleWebSocketConnection(wsConn)
}

// handleMASQUEWebSocket обрабатывает MASQUE через WebSocket
func (wsf *WebSocketFallback) handleMASQUEWebSocket(w http.ResponseWriter, r *http.Request) {
	wsf.logger.Info("MASQUE WebSocket connection request",
		zap.String("remote_addr", r.RemoteAddr),
		zap.String("user_agent", r.UserAgent()))

	// Проверяем Capsule-Protocol заголовок
	if r.Header.Get("Capsule-Protocol") != "?1" {
		http.Error(w, "Capsule-Protocol header required", http.StatusBadRequest)
		return
	}

	// Обновляем соединение до WebSocket
	conn, err := wsf.upgrader.Upgrade(w, r, nil)
	if err != nil {
		wsf.logger.Error("Failed to upgrade to MASQUE WebSocket", zap.Error(err))
		return
	}

	// Создаем MASQUE WebSocket соединение
	wsConn := &WebSocketConnection{
		ID:           fmt.Sprintf("masque_ws_%d", time.Now().UnixNano()),
		Conn:         conn,
		RemoteAddr:   r.RemoteAddr,
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
		IsActive:     true,
	}

	// Сохраняем соединение
	wsf.mu.Lock()
	wsf.connections[wsConn.ID] = wsConn
	wsf.mu.Unlock()

	wsf.logger.Info("MASQUE WebSocket connection established",
		zap.String("conn_id", wsConn.ID),
		zap.String("remote_addr", wsConn.RemoteAddr))

	// Обрабатываем MASQUE соединение
	wsf.handleMASQUEWebSocketConnection(wsConn)
}

// handleWebSocketConnection обрабатывает WebSocket соединение
func (wsf *WebSocketFallback) handleWebSocketConnection(conn *WebSocketConnection) {
	defer func() {
		conn.IsActive = false
		conn.Conn.Close()
		wsf.removeConnection(conn.ID)
	}()

	// Устанавливаем таймауты
	conn.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

	for {
		select {
		case <-wsf.ctx.Done():
			return
		default:
			// Читаем сообщение
			messageType, data, err := conn.Conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					wsf.logger.Error("WebSocket read error",
						zap.String("conn_id", conn.ID),
						zap.Error(err))
				}
				return
			}

			// Обновляем активность
			conn.LastActivity = time.Now()

			// Обрабатываем сообщение
			if err := wsf.processWebSocketMessage(conn, messageType, data); err != nil {
				wsf.logger.Error("Failed to process WebSocket message",
					zap.String("conn_id", conn.ID),
					zap.Error(err))
				return
			}
		}
	}
}

// handleMASQUEWebSocketConnection обрабатывает MASQUE WebSocket соединение
func (wsf *WebSocketFallback) handleMASQUEWebSocketConnection(conn *WebSocketConnection) {
	defer func() {
		conn.IsActive = false
		conn.Conn.Close()
		wsf.removeConnection(conn.ID)
	}()

	// Устанавливаем таймауты
	conn.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

	for {
		select {
		case <-wsf.ctx.Done():
			return
		default:
			// Читаем сообщение
			messageType, data, err := conn.Conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					wsf.logger.Error("MASQUE WebSocket read error",
						zap.String("conn_id", conn.ID),
						zap.Error(err))
				}
				return
			}

			// Обновляем активность
			conn.LastActivity = time.Now()

			// Обрабатываем MASQUE сообщение
			if err := wsf.processMASQUEMessage(conn, messageType, data); err != nil {
				wsf.logger.Error("Failed to process MASQUE message",
					zap.String("conn_id", conn.ID),
					zap.Error(err))
				return
			}
		}
	}
}

// processWebSocketMessage обрабатывает WebSocket сообщение
func (wsf *WebSocketFallback) processWebSocketMessage(conn *WebSocketConnection, messageType int, data []byte) error {
	wsf.logger.Debug("Processing WebSocket message",
		zap.String("conn_id", conn.ID),
		zap.Int("message_type", messageType),
		zap.Int("data_len", len(data)))

	// Здесь можно добавить логику обработки различных типов сообщений
	// Например, маршрутизация, проксирование и т.д.

	// Для демонстрации просто эхо-ответ
	if messageType == websocket.TextMessage {
		response := fmt.Sprintf("Echo: %s", string(data))
		return conn.Conn.WriteMessage(websocket.TextMessage, []byte(response))
	} else if messageType == websocket.BinaryMessage {
		// Для бинарных сообщений просто пересылаем обратно
		return conn.Conn.WriteMessage(websocket.BinaryMessage, data)
	}

	return nil
}

// processMASQUEMessage обрабатывает MASQUE сообщение через WebSocket
func (wsf *WebSocketFallback) processMASQUEMessage(conn *WebSocketConnection, messageType int, data []byte) error {
	wsf.logger.Debug("Processing MASQUE WebSocket message",
		zap.String("conn_id", conn.ID),
		zap.Int("message_type", messageType),
		zap.Int("data_len", len(data)))

	// Здесь можно добавить логику обработки MASQUE протокола через WebSocket
	// Например, CONNECT-UDP/CONNECT-IP запросы, HTTP Datagrams и т.д.

	// Для демонстрации просто логируем datagram
	if messageType == websocket.BinaryMessage {
		wsf.logger.Debug("MASQUE datagram received",
			zap.String("conn_id", conn.ID),
			zap.Int("datagram_len", len(data)))
		
		// Здесь можно добавить логику маршрутизации datagrams
		// Например, пересылка к другим MASQUE соединениям
	}

	return nil
}

// removeConnection удаляет соединение из списка
func (wsf *WebSocketFallback) removeConnection(connID string) {
	wsf.mu.Lock()
	defer wsf.mu.Unlock()
	
	delete(wsf.connections, connID)
	wsf.logger.Debug("WebSocket connection removed", zap.String("conn_id", connID))
}

// handleHealth обрабатывает health check запросы
func (wsf *WebSocketFallback) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	wsf.mu.RLock()
	connectionCount := len(wsf.connections)
	wsf.mu.RUnlock()

	response := fmt.Sprintf(`{
		"status": "healthy",
		"websocket_fallback_enabled": %t,
		"connections": %d,
		"timestamp": "%s"
	}`, wsf.config.WebSocketFallback.Enabled, connectionCount, time.Now().UTC().Format(time.RFC3339))

	w.Write([]byte(response))
}

// GetConnectionCount возвращает количество активных соединений
func (wsf *WebSocketFallback) GetConnectionCount() int {
	wsf.mu.RLock()
	defer wsf.mu.RUnlock()
	
	return len(wsf.connections)
}

// GetConnectionInfo возвращает информацию о соединениях
func (wsf *WebSocketFallback) GetConnectionInfo() map[string]interface{} {
	wsf.mu.RLock()
	defer wsf.mu.RUnlock()

	connections := make([]map[string]interface{}, 0, len(wsf.connections))
	for connID, conn := range wsf.connections {
		connections = append(connections, map[string]interface{}{
			"id":            connID,
			"remote_addr":   conn.RemoteAddr,
			"created_at":    conn.CreatedAt,
			"last_activity": conn.LastActivity,
			"is_active":     conn.IsActive,
		})
	}

	return map[string]interface{}{
		"total_connections": len(wsf.connections),
		"connections":       connections,
	}
}

// Close закрывает WebSocket fallback
func (wsf *WebSocketFallback) Close() error {
	wsf.logger.Info("Closing WebSocket fallback")

	wsf.cancel()

	// Закрываем все соединения
	wsf.mu.Lock()
	for connID, conn := range wsf.connections {
		conn.Conn.Close()
		wsf.logger.Debug("Closed WebSocket connection", zap.String("conn_id", connID))
	}
	wsf.connections = make(map[string]*WebSocketConnection)
	wsf.mu.Unlock()

	wsf.logger.Info("WebSocket fallback closed successfully")
	return nil
}
