package tunnel

import (
	"context"
	"encoding/binary"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// WebSocketTunnel представляет WebSocket туннель для fallback механизмов
type WebSocketTunnel struct {
	logger      *zap.Logger
	upgrader    websocket.Upgrader
	connections map[string]*WebSocketConnection
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc

	// Metrics
	totalConnections      uint64
	activeConnections     int64
	totalBytesTransferred uint64
}

// WebSocketConnection представляет WebSocket соединение
type WebSocketConnection struct {
	ID           string
	Conn         *websocket.Conn
	RemoteAddr   string
	ConnectedAt  time.Time
	LastActivity time.Time
	BytesSent    uint64
	BytesRecv    uint64
	IsActive     bool
}

// WebSocketTunnelConfig конфигурация WebSocket туннеля (на основе gorilla/websocket)
type WebSocketTunnelConfig struct {
	ReadBufferSize    int
	WriteBufferSize   int
	CheckOrigin       func(r *http.Request) bool
	HandshakeTimeout  time.Duration
	PingPeriod        time.Duration
	PongWait          time.Duration
	WriteWait         time.Duration
	EnableCompression bool
	Subprotocols      []string
}

// DefaultWebSocketTunnelConfig возвращает конфигурацию по умолчанию
func DefaultWebSocketTunnelConfig() *WebSocketTunnelConfig {
	return &WebSocketTunnelConfig{
		ReadBufferSize:   4096,
		WriteBufferSize:  4096,
		CheckOrigin:      func(r *http.Request) bool { return true }, // Allow all origins for now
		HandshakeTimeout: 10 * time.Second,
		PingPeriod:       54 * time.Second,
		PongWait:         60 * time.Second,
		WriteWait:        10 * time.Second,
	}
}

// NewWebSocketTunnel создает новый WebSocket туннель
func NewWebSocketTunnel(logger *zap.Logger, config *WebSocketTunnelConfig) *WebSocketTunnel {
	if config == nil {
		config = DefaultWebSocketTunnelConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	upgrader := websocket.Upgrader{
		ReadBufferSize:    config.ReadBufferSize,
		WriteBufferSize:   config.WriteBufferSize,
		CheckOrigin:       config.CheckOrigin,
		HandshakeTimeout:  config.HandshakeTimeout,
		EnableCompression: config.EnableCompression,
		Subprotocols:      config.Subprotocols,
	}

	return &WebSocketTunnel{
		logger:      logger,
		upgrader:    upgrader,
		connections: make(map[string]*WebSocketConnection),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// HandleWebSocket обрабатывает WebSocket соединения
func (wst *WebSocketTunnel) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	wst.logger.Info("WebSocket connection attempt",
		zap.String("remote_addr", r.RemoteAddr),
		zap.String("user_agent", r.UserAgent()))

	// Upgrade HTTP connection to WebSocket
	conn, err := wst.upgrader.Upgrade(w, r, nil)
	if err != nil {
		wst.logger.Error("Failed to upgrade to WebSocket", zap.Error(err))
		return
	}

	// Create connection ID
	connID := wst.generateConnectionID()

	// Create WebSocket connection
	wsConn := &WebSocketConnection{
		ID:           connID,
		Conn:         conn,
		RemoteAddr:   r.RemoteAddr,
		ConnectedAt:  time.Now(),
		LastActivity: time.Now(),
		IsActive:     true,
	}

	// Store connection
	wst.mu.Lock()
	wst.connections[connID] = wsConn
	wst.totalConnections++
	wst.activeConnections++
	wst.mu.Unlock()

	wst.logger.Info("WebSocket connection established",
		zap.String("conn_id", connID),
		zap.String("remote_addr", r.RemoteAddr))

	// Handle connection
	go wst.handleConnection(wsConn)

	// Cleanup on disconnect
	defer func() {
		wst.mu.Lock()
		delete(wst.connections, connID)
		wst.activeConnections--
		wst.mu.Unlock()

		conn.Close()
		wst.logger.Info("WebSocket connection closed", zap.String("conn_id", connID))
	}()
}

// handleConnection обрабатывает WebSocket соединение
func (wst *WebSocketTunnel) handleConnection(conn *WebSocketConnection) {
	// Set up ping/pong handlers
	conn.Conn.SetPongHandler(func(string) error {
		conn.LastActivity = time.Now()
		return nil
	})

	// Set up close handler
	conn.Conn.SetCloseHandler(func(code int, text string) error {
		wst.logger.Info("WebSocket connection closing",
			zap.String("conn_id", conn.ID),
			zap.Int("code", code),
			zap.String("text", text))
		return nil
	})

	// Start ping ticker
	pingTicker := time.NewTicker(54 * time.Second)
	defer pingTicker.Stop()

	// Start message handling
	go wst.handleMessages(conn)

	// Handle ping/pong
	for {
		select {
		case <-wst.ctx.Done():
			return
		case <-pingTicker.C:
			conn.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				wst.logger.Error("Failed to send ping",
					zap.String("conn_id", conn.ID),
					zap.Error(err))
				return
			}
		}
	}
}

// handleMessages обрабатывает сообщения WebSocket
func (wst *WebSocketTunnel) handleMessages(conn *WebSocketConnection) {
	for {
		select {
		case <-wst.ctx.Done():
			return
		default:
			// Set read deadline
			conn.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))

			messageType, data, err := conn.Conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					wst.logger.Error("WebSocket read error",
						zap.String("conn_id", conn.ID),
						zap.Error(err))
				}
				return
			}

			conn.LastActivity = time.Now()
			conn.BytesRecv += uint64(len(data))

			// Handle different message types
			switch messageType {
			case websocket.BinaryMessage:
				wst.handleBinaryMessage(conn, data)
			case websocket.TextMessage:
				wst.handleTextMessage(conn, data)
			case websocket.PingMessage:
				wst.handlePingMessage(conn, data)
			case websocket.PongMessage:
				wst.handlePongMessage(conn, data)
			default:
				wst.logger.Debug("Unknown message type",
					zap.String("conn_id", conn.ID),
					zap.Int("type", messageType))
			}
		}
	}
}

// handleBinaryMessage обрабатывает бинарные сообщения (туннелированные данные)
func (wst *WebSocketTunnel) handleBinaryMessage(conn *WebSocketConnection, data []byte) {
	if len(data) < 8 {
		wst.logger.Debug("Binary message too short",
			zap.String("conn_id", conn.ID),
			zap.Int("size", len(data)))
		return
	}

	// Parse tunnel packet header
	// Format: [4 bytes length][4 bytes type][payload]
	length := binary.BigEndian.Uint32(data[0:4])
	packetType := binary.BigEndian.Uint32(data[4:8])
	payload := data[8:]

	if len(payload) != int(length) {
		wst.logger.Debug("Payload length mismatch",
			zap.String("conn_id", conn.ID),
			zap.Uint32("expected", length),
			zap.Int("actual", len(payload)))
		return
	}

	wst.logger.Debug("Binary message received",
		zap.String("conn_id", conn.ID),
		zap.Uint32("type", packetType),
		zap.Int("size", len(payload)))

	// Handle different packet types
	switch packetType {
	case 1: // Data packet
		wst.handleDataPacket(conn, payload)
	case 2: // Control packet
		wst.handleControlPacket(conn, payload)
	case 3: // Heartbeat packet
		wst.handleHeartbeatPacket(conn, payload)
	default:
		wst.logger.Debug("Unknown packet type",
			zap.String("conn_id", conn.ID),
			zap.Uint32("type", packetType))
	}
}

// handleTextMessage обрабатывает текстовые сообщения
func (wst *WebSocketTunnel) handleTextMessage(conn *WebSocketConnection, data []byte) {
	wst.logger.Debug("Text message received",
		zap.String("conn_id", conn.ID),
		zap.String("message", string(data)))

	// Echo back the message
	response := fmt.Sprintf("Echo: %s", string(data))
	conn.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if err := conn.Conn.WriteMessage(websocket.TextMessage, []byte(response)); err != nil {
		wst.logger.Error("Failed to send text response",
			zap.String("conn_id", conn.ID),
			zap.Error(err))
	}
}

// handlePingMessage обрабатывает ping сообщения
func (wst *WebSocketTunnel) handlePingMessage(conn *WebSocketConnection, data []byte) {
	wst.logger.Debug("Ping message received", zap.String("conn_id", conn.ID))

	// Send pong response
	conn.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if err := conn.Conn.WriteMessage(websocket.PongMessage, data); err != nil {
		wst.logger.Error("Failed to send pong",
			zap.String("conn_id", conn.ID),
			zap.Error(err))
	}
}

// handlePongMessage обрабатывает pong сообщения
func (wst *WebSocketTunnel) handlePongMessage(conn *WebSocketConnection, data []byte) {
	wst.logger.Debug("Pong message received", zap.String("conn_id", conn.ID))
	conn.LastActivity = time.Now()
}

// handleDataPacket обрабатывает туннелированные данные
func (wst *WebSocketTunnel) handleDataPacket(conn *WebSocketConnection, data []byte) {
	wst.logger.Debug("Data packet received",
		zap.String("conn_id", conn.ID),
		zap.Int("size", len(data)))

	// In a real implementation, this would forward the data to the appropriate destination
	// For now, we'll just log it and echo it back

	// Create response packet
	responseData := make([]byte, 8+len(data))
	binary.BigEndian.PutUint32(responseData[0:4], uint32(len(data)))
	binary.BigEndian.PutUint32(responseData[4:8], 1) // Data packet type
	copy(responseData[8:], data)

	// Send response
	conn.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if err := conn.Conn.WriteMessage(websocket.BinaryMessage, responseData); err != nil {
		wst.logger.Error("Failed to send data response",
			zap.String("conn_id", conn.ID),
			zap.Error(err))
		return
	}

	conn.BytesSent += uint64(len(responseData))
	wst.totalBytesTransferred += uint64(len(responseData))
}

// handleControlPacket обрабатывает управляющие пакеты
func (wst *WebSocketTunnel) handleControlPacket(conn *WebSocketConnection, data []byte) {
	wst.logger.Debug("Control packet received",
		zap.String("conn_id", conn.ID),
		zap.Int("size", len(data)))

	// Parse control packet
	if len(data) < 4 {
		return
	}

	controlType := binary.BigEndian.Uint32(data[0:4])
	controlData := data[4:]

	switch controlType {
	case 1: // Connection request
		wst.handleConnectionRequest(conn, controlData)
	case 2: // Connection response
		wst.handleConnectionResponse(conn, controlData)
	case 3: // Disconnect request
		wst.handleDisconnectRequest(conn, controlData)
	default:
		wst.logger.Debug("Unknown control type",
			zap.String("conn_id", conn.ID),
			zap.Uint32("type", controlType))
	}
}

// handleHeartbeatPacket обрабатывает heartbeat пакеты
func (wst *WebSocketTunnel) handleHeartbeatPacket(conn *WebSocketConnection, data []byte) {
	wst.logger.Debug("Heartbeat packet received", zap.String("conn_id", conn.ID))
	conn.LastActivity = time.Now()

	// Send heartbeat response
	responseData := make([]byte, 8+len(data))
	binary.BigEndian.PutUint32(responseData[0:4], uint32(len(data)))
	binary.BigEndian.PutUint32(responseData[4:8], 3) // Heartbeat packet type
	copy(responseData[8:], data)

	conn.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if err := conn.Conn.WriteMessage(websocket.BinaryMessage, responseData); err != nil {
		wst.logger.Error("Failed to send heartbeat response",
			zap.String("conn_id", conn.ID),
			zap.Error(err))
	}
}

// handleConnectionRequest обрабатывает запросы на соединение
func (wst *WebSocketTunnel) handleConnectionRequest(conn *WebSocketConnection, data []byte) {
	wst.logger.Info("Connection request received", zap.String("conn_id", conn.ID))

	// In a real implementation, this would establish a connection to the target
	// For now, we'll just send a success response

	responseData := make([]byte, 4)
	binary.BigEndian.PutUint32(responseData, 0) // Success

	controlResponse := make([]byte, 8+len(responseData))
	binary.BigEndian.PutUint32(controlResponse[0:4], uint32(len(responseData)))
	binary.BigEndian.PutUint32(controlResponse[4:8], 2) // Connection response
	copy(controlResponse[8:], responseData)

	conn.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if err := conn.Conn.WriteMessage(websocket.BinaryMessage, controlResponse); err != nil {
		wst.logger.Error("Failed to send connection response",
			zap.String("conn_id", conn.ID),
			zap.Error(err))
	}
}

// handleConnectionResponse обрабатывает ответы на соединение
func (wst *WebSocketTunnel) handleConnectionResponse(conn *WebSocketConnection, data []byte) {
	wst.logger.Info("Connection response received", zap.String("conn_id", conn.ID))
}

// handleDisconnectRequest обрабатывает запросы на отключение
func (wst *WebSocketTunnel) handleDisconnectRequest(conn *WebSocketConnection, data []byte) {
	wst.logger.Info("Disconnect request received", zap.String("conn_id", conn.ID))

	// Close the connection
	conn.Conn.Close()
}

// generateConnectionID генерирует уникальный ID соединения
func (wst *WebSocketTunnel) generateConnectionID() string {
	return fmt.Sprintf("ws_%d", time.Now().UnixNano())
}

// GetStats возвращает статистику WebSocket туннеля
func (wst *WebSocketTunnel) GetStats() map[string]interface{} {
	wst.mu.RLock()
	defer wst.mu.RUnlock()

	return map[string]interface{}{
		"total_connections":       wst.totalConnections,
		"active_connections":      wst.activeConnections,
		"total_bytes_transferred": wst.totalBytesTransferred,
		"connections":             len(wst.connections),
	}
}

// GetConnectionList возвращает список активных соединений
func (wst *WebSocketTunnel) GetConnectionList() []*WebSocketConnection {
	wst.mu.RLock()
	defer wst.mu.RUnlock()

	connections := make([]*WebSocketConnection, 0, len(wst.connections))
	for _, conn := range wst.connections {
		connections = append(connections, conn)
	}

	return connections
}

// Stop останавливает WebSocket туннель
func (wst *WebSocketTunnel) Stop() {
	wst.cancel()

	// Close all connections
	wst.mu.Lock()
	for _, conn := range wst.connections {
		conn.Conn.Close()
	}
	wst.connections = make(map[string]*WebSocketConnection)
	wst.mu.Unlock()

	wst.logger.Info("WebSocket tunnel stopped")
}

// CreateHTTPHandler создает HTTP обработчик для WebSocket туннеля
func (wst *WebSocketTunnel) CreateHTTPHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Add CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Only allow GET requests for WebSocket upgrade
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Handle WebSocket upgrade
		wst.HandleWebSocket(w, r)
	}
}
