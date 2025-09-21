package api

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// WebSocketConnection представляет WebSocket соединение
type WebSocketConnection struct {
	PeerID   string
	TenantID string
	Conn     *websocket.Conn
	Send     chan []byte
	Hub      *Hub
	mu       sync.RWMutex
}

// Hub управляет WebSocket соединениями
type Hub struct {
	Connections    map[string]*WebSocketConnection
	Register       chan *WebSocketConnection
	Unregister     chan *WebSocketConnection
	Broadcast      chan []byte
	upgrader       websocket.Upgrader
	mu             sync.RWMutex
	logger         *zap.Logger
	peers          map[string]*PeerConnection // Добавляем поле peers
	peerRegister   chan *PeerConnection       // Канал для регистрации пиров
	peerUnregister chan *PeerConnection       // Канал для отмены регистрации пиров
	peerBroadcast  chan []byte                // Канал для broadcast сообщений
}

// NewHub создает новый Hub
func NewHub(logger *zap.Logger) *Hub {
	return &Hub{
		Connections:    make(map[string]*WebSocketConnection),
		Register:       make(chan *WebSocketConnection),
		Unregister:     make(chan *WebSocketConnection),
		Broadcast:      make(chan []byte),
		peers:          make(map[string]*PeerConnection),
		peerRegister:   make(chan *PeerConnection),
		peerUnregister: make(chan *PeerConnection),
		peerBroadcast:  make(chan []byte),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // В production здесь должна быть проверка origin
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		logger: logger,
	}
}

// Run запускает Hub
func (h *Hub) Run() {
	for {
		select {
		case conn := <-h.Register:
			h.mu.Lock()
			h.Connections[conn.PeerID] = conn
			h.mu.Unlock()
			h.logger.Info("Peer connected to WebSocket",
				zap.String("peer_id", conn.PeerID),
				zap.String("tenant_id", conn.TenantID),
			)

		case conn := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.Connections[conn.PeerID]; ok {
				delete(h.Connections, conn.PeerID)
				close(conn.Send)
				h.logger.Info("Peer disconnected from WebSocket",
					zap.String("peer_id", conn.PeerID),
					zap.String("tenant_id", conn.TenantID),
				)
			}
			h.mu.Unlock()

		case message := <-h.Broadcast:
			h.mu.RLock()
			for _, conn := range h.Connections {
				select {
				case conn.Send <- message:
				default:
					close(conn.Send)
					delete(h.Connections, conn.PeerID)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// HandleWebSocket обрабатывает WebSocket соединения
func (h *Hub) HandleWebSocket(c *gin.Context) {
	// Валидация JWT токена
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Missing authorization header",
		})
		return
	}

	// TODO: Валидация токена и извлечение peer_id и tenant_id
	// Пока используем mock данные
	peerID := "mock-peer-id"
	tenantID := "mock-tenant-id"

	// Обновление HTTP соединения до WebSocket
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade connection to WebSocket", zap.Error(err))
		return
	}

	// Создание WebSocket соединения
	wsConn := &WebSocketConnection{
		PeerID:   peerID,
		TenantID: tenantID,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		Hub:      h,
	}

	// Регистрация соединения
	h.Register <- wsConn

	// Запуск горутин для чтения и записи
	go wsConn.writePump()
	go wsConn.readPump()
}

// BroadcastToTenant отправляет сообщение всем пирам тенанта
func (h *Hub) BroadcastToTenant(tenantID string, message WebSocketMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, conn := range h.Connections {
		if conn.TenantID == tenantID {
			select {
			case conn.Send <- []byte(message.String()):
			default:
				close(conn.Send)
				delete(h.Connections, conn.PeerID)
			}
		}
	}
}

// BroadcastToPeer отправляет сообщение конкретному пиру
func (h *Hub) BroadcastToPeer(peerID string, message WebSocketMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if conn, exists := h.Connections[peerID]; exists {
		select {
		case conn.Send <- []byte(message.String()):
		default:
			close(conn.Send)
			delete(h.Connections, peerID)
		}
	}
}

// NotifyPeerJoined уведомляет о подключении нового пира
func (h *Hub) NotifyPeerJoined(tenantID string, peer *Peer) {
	message := WebSocketMessage{
		Type: "peer_joined",
		Data: PeerJoinedData{
			PeerID: peer.PeerID,
			// P2P QUIC configuration integrated into simplified Peer structure
			Timestamp: time.Now(),
		},
		Timestamp: time.Now(),
	}

	h.BroadcastToTenant(tenantID, message)
}

// NotifyPeerLeft уведомляет об отключении пира
func (h *Hub) NotifyPeerLeft(tenantID string, peerID string) {
	message := WebSocketMessage{
		Type: "peer_left",
		Data: PeerLeftData{
			PeerID:    peerID,
			Timestamp: time.Now(),
		},
		Timestamp: time.Now(),
	}

	h.BroadcastToTenant(tenantID, message)
}

// NotifyPeerStatusUpdate уведомляет об обновлении статуса пира
func (h *Hub) NotifyPeerStatusUpdate(tenantID string, peerID string, isOnline bool, latencyMS int, meshStatus string) {
	message := WebSocketMessage{
		Type: "peer_status_update",
		Data: PeerStatusUpdateData{
			PeerID:     peerID,
			IsOnline:   isOnline,
			LatencyMS:  latencyMS,
			MeshStatus: meshStatus,
			Timestamp:  time.Now(),
		},
		Timestamp: time.Now(),
	}

	h.BroadcastToTenant(tenantID, message)
}

// SendRelayMessage отправляет сообщение через relay
func (h *Hub) SendRelayMessage(peerID string, messageType string, data string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	conn, exists := h.Connections[peerID]
	if !exists {
		return false
	}

	message := WebSocketMessage{
		Type: "relay_message",
		Data: RelayMessageData{
			Type:        "relay_message",
			MessageType: messageType,
			Data:        data,
			Timestamp:   time.Now(),
		},
		Timestamp: time.Now(),
	}

	select {
	case conn.Send <- []byte(message.String()):
		return true
	default:
		return false
	}
}

// GetConnectionCount возвращает количество активных соединений
func (h *Hub) GetConnectionCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.Connections)
}

// GetTenantConnectionCount возвращает количество соединений для тенанта
func (h *Hub) GetTenantConnectionCount(tenantID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	count := 0
	for _, conn := range h.Connections {
		if conn.TenantID == tenantID {
			count++
		}
	}
	return count
}

// GetPeerConnections возвращает соединения пиров
func (h *Hub) GetPeerConnections() map[string]*PeerConnection {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.peers
}

// register регистрирует пира
func (h *Hub) register(peer *PeerConnection) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.peers[peer.ID] = peer
}

// unregister отменяет регистрацию пира
func (h *Hub) unregister(peerID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.peers, peerID)
}

// broadcast отправляет сообщение всем пирам
func (h *Hub) broadcast(message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, conn := range h.Connections {
		select {
		case conn.Send <- message:
		default:
			close(conn.Send)
			delete(h.Connections, conn.PeerID)
		}
	}
}

// readPump читает сообщения из WebSocket соединения
func (c *WebSocketConnection) readPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512)
	if err := c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second)); err != nil {
		return
	}
	c.Conn.SetPongHandler(func(string) error {
		if err := c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second)); err != nil {
			return err
		}
		return nil
	})

	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.Hub.logger.Error("WebSocket error", zap.Error(err))
			}
			break
		}
	}
}

// writePump записывает сообщения в WebSocket соединение
func (c *WebSocketConnection) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if err := c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
				return
			}
			if !ok {
				if err := c.Conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					// Log error but continue
				}
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			if _, err := w.Write(message); err != nil {
				w.Close()
				return
			}

			// Добавляем дополнительные сообщения из очереди
			n := len(c.Send)
			for i := 0; i < n; i++ {
				if _, err := w.Write([]byte{'\n'}); err != nil {
					w.Close()
					return
				}
				if _, err := w.Write(<-c.Send); err != nil {
					w.Close()
					return
				}
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			if err := c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
				return
			}
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// String возвращает JSON представление сообщения
func (m WebSocketMessage) String() string {
	data, err := json.Marshal(m)
	if err != nil {
		return `{"type":"` + m.Type + `","timestamp":"` + m.Timestamp.Format(time.RFC3339) + `"}`
	}
	return string(data)
}
