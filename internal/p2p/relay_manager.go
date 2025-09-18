package p2p

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/quic-go/quic-go"
	"go.uber.org/zap"
)

// RelayManager управляет P2P relay соединениями
type RelayManager struct {
	tenants        map[string]*Tenant
	connections    map[string]*RelayP2PConnection
	quicConnections map[string]*QUICConnection
	mutex          sync.RWMutex
	logger         *zap.Logger
	ctx            context.Context
	cancel         context.CancelFunc
}

// Tenant представляет изолированного пользователя
type Tenant struct {
	TenantID        string                    `json:"tenant_id"`
	NetworkCIDR     string                    `json:"network_cidr"`
	Peers           map[string]*Peer          `json:"peers"`
	Connections     map[string]*RelayP2PConnection `json:"connections"`
	MaxPeers        int                       `json:"max_peers"`
	MaxConnections  int                       `json:"max_connections"`
	CreatedAt       time.Time                 `json:"created_at"`
	LastActivity    time.Time                 `json:"last_activity"`
	Mutex           sync.RWMutex              `json:"-"`
}

// RelayP2PConnection представляет P2P соединение между пирами для relay
type RelayP2PConnection struct {
	ConnectionID    string            `json:"connection_id"`
	TenantID        string            `json:"tenant_id"`
	SourcePeer      string            `json:"source_peer"`
	TargetPeer      string            `json:"target_peer"`
	SourceIP        string            `json:"source_ip"`
	TargetIP        string            `json:"target_ip"`
	ConnectionType  string            `json:"connection_type"` // "quic", "quic-ice"
	Status          string            `json:"status"`          // "connecting", "established", "disconnected"
	CreatedAt       time.Time         `json:"created_at"`
	LastActivity    time.Time         `json:"last_activity"`
	BytesTransferred int64            `json:"bytes_transferred"`
	QUICConn        *quic.Conn        `json:"-"`
	ICEConn         *net.Conn         `json:"-"`
	Mutex           sync.RWMutex      `json:"-"`
}

// QUICConnection представляет QUIC соединение с пиром
type QUICConnection struct {
	PeerID      string
	TenantID    string
	Conn        *quic.Conn
	Streams     map[quic.StreamID]quic.Stream
	LastSeen    time.Time
	IsActive    bool
	Mutex       sync.RWMutex
}

// NewRelayManager создает новый Relay Manager
func NewRelayManager(logger *zap.Logger) *RelayManager {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &RelayManager{
		tenants:         make(map[string]*Tenant),
		connections:     make(map[string]*RelayP2PConnection),
		quicConnections: make(map[string]*QUICConnection),
		logger:          logger,
		ctx:             ctx,
		cancel:          cancel,
	}
}

// RegisterPeer регистрирует пира в tenant
func (rm *RelayManager) RegisterPeer(tenantID, peerID, publicKey string, allowedIPs []string) (*Peer, error) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	
	// Создаем tenant если не существует
	if _, exists := rm.tenants[tenantID]; !exists {
		rm.tenants[tenantID] = &Tenant{
			TenantID:       tenantID,
			NetworkCIDR:    rm.getTenantNetwork(tenantID),
			Peers:          make(map[string]*Peer),
			Connections:    make(map[string]*RelayP2PConnection),
			MaxPeers:       100,
			MaxConnections: 1000,
			CreatedAt:      time.Now(),
		}
	}
	
	tenant := rm.tenants[tenantID]
	
	// Проверяем лимиты
	if len(tenant.Peers) >= tenant.MaxPeers {
		return nil, fmt.Errorf("tenant %s has reached max peers limit: %d", tenantID, tenant.MaxPeers)
	}
	
	// Выделяем IP адрес
	allocatedIP := rm.allocateIPForTenant(tenantID)
	
	// Создаем peer
	peer := &Peer{
		PeerID:      peerID,
		TenantID:    tenantID,
		IsOnline:    true,
		LastSeen:    time.Now(),
		RegisteredAt: time.Now(),
	}
	
	// Добавляем в tenant
	tenant.Peers[peerID] = peer
	
	rm.logger.Info("Peer registered",
		zap.String("peer_id", peerID),
		zap.String("tenant_id", tenantID),
		zap.String("allocated_ip", allocatedIP))
	
	return peer, nil
}

// RegisterQUICConnection регистрирует QUIC соединение с пиром
func (rm *RelayManager) RegisterQUICConnection(peerID, tenantID string, conn *quic.Conn) error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	
	quicConn := &QUICConnection{
		PeerID:   peerID,
		TenantID: tenantID,
		Conn:     conn,
		Streams:  make(map[quic.StreamID]quic.Stream),
		LastSeen: time.Now(),
		IsActive: true,
	}
	
	rm.quicConnections[peerID] = quicConn
	
	rm.logger.Info("QUIC connection registered",
		zap.String("peer_id", peerID),
		zap.String("tenant_id", tenantID))
	
	return nil
}

// RelayPacket пересылает пакет между пирами
func (rm *RelayManager) RelayPacket(sourcePeerID string, data []byte) error {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()
	
	// Получаем информацию об отправителе
	sourceConn, exists := rm.quicConnections[sourcePeerID]
	if !exists {
		return fmt.Errorf("source peer connection not found: %s", sourcePeerID)
	}
	
	// Парсим пакет для определения получателя
	targetPeerID, err := rm.parsePacketTarget(data)
	if err != nil {
		return fmt.Errorf("failed to parse packet target: %v", err)
	}
	
	// Получаем информацию о получателе
	targetConn, exists := rm.quicConnections[targetPeerID]
	if !exists {
		return fmt.Errorf("target peer connection not found: %s", targetPeerID)
	}
	
	// Проверяем изоляцию tenant
	if sourceConn.TenantID != targetConn.TenantID {
		return fmt.Errorf("cross-tenant relay not allowed: %s -> %s", 
			sourceConn.TenantID, targetConn.TenantID)
	}
	
	// Находим или создаем P2P соединение
	connectionID := rm.generateConnectionID(sourcePeerID, targetPeerID)
	connection, exists := rm.connections[connectionID]
	if !exists {
		connection, err = rm.establishP2PConnection(sourcePeerID, targetPeerID)
		if err != nil {
			return fmt.Errorf("failed to establish P2P connection: %v", err)
		}
		rm.connections[connectionID] = connection
	}
	
	// Пересылаем пакет
	err = rm.forwardPacket(targetConn, data)
	if err != nil {
		return fmt.Errorf("failed to forward packet: %v", err)
	}
	
	// Обновляем статистику
	connection.Mutex.Lock()
	connection.BytesTransferred += int64(len(data))
	connection.LastActivity = time.Now()
	connection.Mutex.Unlock()
	
	rm.logger.Debug("Packet relayed",
		zap.String("source_peer", sourcePeerID),
		zap.String("target_peer", targetPeerID),
		zap.Int("bytes", len(data)))
	
	return nil
}

// parsePacketTarget парсит пакет для определения получателя
func (rm *RelayManager) parsePacketTarget(data []byte) (string, error) {
	// Простая реализация - в реальности нужно парсить QUIC заголовки
	// Для демонстрации используем первые 32 байта как peer ID
	if len(data) >= 32 {
		// Ищем peer ID в заголовке пакета
		// В реальной реализации здесь должен быть парсинг QUIC заголовков
		peerID := string(data[:32])
		
		// Проверяем, что это валидный peer ID
		if _, exists := rm.quicConnections[peerID]; exists {
			return peerID, nil
		}
	}
	
	return "", fmt.Errorf("cannot determine target peer from packet")
}

// establishP2PConnection создает P2P соединение
func (rm *RelayManager) establishP2PConnection(sourcePeerID, targetPeerID string) (*RelayP2PConnection, error) {
	// Получаем информацию о пирах
	sourceConn := rm.quicConnections[sourcePeerID]
	targetConn := rm.quicConnections[targetPeerID]
	
	if sourceConn == nil || targetConn == nil {
		return nil, fmt.Errorf("peer connections not found")
	}
	
	// Получаем tenant
	tenant := rm.tenants[sourceConn.TenantID]
	if tenant == nil {
		return nil, fmt.Errorf("tenant not found: %s", sourceConn.TenantID)
	}
	
	// Проверяем лимиты соединений
	if len(tenant.Connections) >= tenant.MaxConnections {
		return nil, fmt.Errorf("tenant %s has reached max connections limit: %d", 
			tenant.TenantID, tenant.MaxConnections)
	}
	
	// Создаем соединение
	connection := &RelayP2PConnection{
		ConnectionID:   rm.generateConnectionID(sourcePeerID, targetPeerID),
		TenantID:       sourceConn.TenantID,
		SourcePeer:     sourcePeerID,
		TargetPeer:     targetPeerID,
		SourceIP:       rm.getPeerIP(sourcePeerID),
		TargetIP:       rm.getPeerIP(targetPeerID),
		ConnectionType: "quic",
		Status:         "established",
		CreatedAt:      time.Now(),
		LastActivity:   time.Now(),
		QUICConn:       targetConn.Conn,
	}
	
	// Добавляем в tenant
	tenant.Connections[connection.ConnectionID] = connection
	
	rm.logger.Info("P2P connection established",
		zap.String("connection_id", connection.ConnectionID),
		zap.String("source_peer", sourcePeerID),
		zap.String("target_peer", targetPeerID),
		zap.String("tenant_id", sourceConn.TenantID))
	
	return connection, nil
}

// forwardPacket пересылает пакет получателю
func (rm *RelayManager) forwardPacket(targetConn *QUICConnection, data []byte) error {
	targetConn.Mutex.Lock()
	defer targetConn.Mutex.Unlock()
	
	if !targetConn.IsActive {
		return fmt.Errorf("target connection is not active")
	}
	
	// Создаем новый поток для пересылки данных
	stream, err := targetConn.Conn.OpenStream()
	if err != nil {
		return fmt.Errorf("failed to open stream: %v", err)
	}
	defer stream.Close()
	
	// Отправляем данные
	_, err = stream.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write to stream: %v", err)
	}
	
	targetConn.LastSeen = time.Now()
	
	return nil
}

// getTenantNetwork возвращает сеть для tenant
func (rm *RelayManager) getTenantNetwork(tenantID string) string {
	// Каждый tenant получает свою подсеть
	tenantNetworks := map[string]string{
		"tenant-216420165": "10.100.77.0/24",
		"tenant-123456789": "10.100.78.0/24",
		"tenant-987654321": "10.100.79.0/24",
	}
	
	network, exists := tenantNetworks[tenantID]
	if !exists {
		// Генерируем новую подсеть для нового tenant
		network = rm.generateNewTenantNetwork(tenantID)
	}
	
	return network
}

// generateNewTenantNetwork генерирует новую сеть для tenant
func (rm *RelayManager) generateNewTenantNetwork(tenantID string) string {
	// Простая реализация - в реальности нужно более сложная логика
	return "10.100.80.0/24"
}

// allocateIPForTenant выделяет IP адрес для tenant
func (rm *RelayManager) allocateIPForTenant(tenantID string) string {
	// Простая реализация - в реальности нужно проверять занятость IP
	return "10.100.77.10"
}

// getPeerIP возвращает IP адрес пира
func (rm *RelayManager) getPeerIP(peerID string) string {
	// Простая реализация - в реальности нужно получать из конфигурации
	return "10.100.77.11"
}

// generateConnectionID генерирует уникальный ID соединения
func (rm *RelayManager) generateConnectionID(sourcePeer, targetPeer string) string {
	if sourcePeer < targetPeer {
		return fmt.Sprintf("conn-%s-%s", sourcePeer, targetPeer)
	}
	return fmt.Sprintf("conn-%s-%s", targetPeer, sourcePeer)
}

// GetTenantInfo возвращает информацию о tenant
func (rm *RelayManager) GetTenantInfo(tenantID string) (*Tenant, error) {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()
	
	tenant, exists := rm.tenants[tenantID]
	if !exists {
		return nil, fmt.Errorf("tenant not found: %s", tenantID)
	}
	
	return tenant, nil
}

// GetConnectionInfo возвращает информацию о соединении
func (rm *RelayManager) GetConnectionInfo(connectionID string) (*RelayP2PConnection, error) {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()
	
	connection, exists := rm.connections[connectionID]
	if !exists {
		return nil, fmt.Errorf("connection not found: %s", connectionID)
	}
	
	return connection, nil
}

// CleanupOfflinePeers очищает офлайн пиров
func (rm *RelayManager) CleanupOfflinePeers() {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	
	cutoff := time.Now().Add(-5 * time.Minute) // 5 минут без активности
	
	for peerID, quicConn := range rm.quicConnections {
		if quicConn.LastSeen.Before(cutoff) {
			quicConn.IsActive = false
			quicConn.Conn.CloseWithError(0, "peer offline")
			delete(rm.quicConnections, peerID)
			
			rm.logger.Info("Peer marked as offline",
				zap.String("peer_id", peerID),
				zap.String("tenant_id", quicConn.TenantID))
		}
	}
}

// StartCleanupRoutine запускает routine для очистки офлайн пиров
func (rm *RelayManager) StartCleanupRoutine() {
	ticker := time.NewTicker(60 * time.Second)
	go func() {
		for {
			select {
			case <-rm.ctx.Done():
				return
			case <-ticker.C:
				rm.CleanupOfflinePeers()
			}
		}
	}()
}

// Shutdown останавливает Relay Manager
func (rm *RelayManager) Shutdown() {
	rm.cancel()
	
	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	
	// Закрываем все QUIC соединения
	for _, quicConn := range rm.quicConnections {
		quicConn.Conn.CloseWithError(0, "server shutdown")
	}
	
	rm.logger.Info("Relay Manager shutdown completed")
}
