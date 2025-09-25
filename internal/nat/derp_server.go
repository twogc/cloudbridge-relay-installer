package nat

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"

	"go.uber.org/zap"
)

// DERPServer представляет DERP сервер для fallback механизмов
// Основано на архитектуре Tailscale DERP
type DERPServer struct {
	logger     *zap.Logger
	listenAddr string
	conn       *net.UDPConn
	clients    map[string]*DERPClient
	relays     map[string]*DERPRelay
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc

	// Metrics
	totalPackets   uint64
	relayedPackets uint64
	activeClients  int64
}

// DERPClient представляет клиента DERP сервера
type DERPClient struct {
	ID          string
	Addr        *net.UDPAddr
	PublicKey   []byte
	ConnectedAt time.Time
	LastSeen    time.Time
	PacketsSent uint64
	PacketsRecv uint64
}

// DERPRelay представляет relay соединение между клиентами
type DERPRelay struct {
	ID        string
	ClientA   *DERPClient
	ClientB   *DERPClient
	CreatedAt time.Time
	LastUsed  time.Time
	Packets   uint64
}

// DERP Packet Types
const (
	DERPPacketTypeConnect    = 0x01
	DERPPacketTypeData       = 0x02
	DERPPacketTypeAck        = 0x03
	DERPPacketTypeDisconnect = 0x04
	DERPPacketTypePing       = 0x05
	DERPPacketTypePong       = 0x06
)

// DERP Packet Header
type DERPPacketHeader struct {
	Type      uint8
	ClientID  [16]byte
	TargetID  [16]byte
	Length    uint16
	Sequence  uint32
	Timestamp uint64
}

// NewDERPServer создает новый DERP сервер
func NewDERPServer(listenAddr string, logger *zap.Logger) (*DERPServer, error) {
	ctx, cancel := context.WithCancel(context.Background())

	return &DERPServer{
		logger:     logger,
		listenAddr: listenAddr,
		clients:    make(map[string]*DERPClient),
		relays:     make(map[string]*DERPRelay),
		ctx:        ctx,
		cancel:     cancel,
	}, nil
}

// Start запускает DERP сервер
func (ds *DERPServer) Start() error {
	// Resolve listen address
	udpAddr, err := net.ResolveUDPAddr("udp", ds.listenAddr)
	if err != nil {
		return fmt.Errorf("failed to resolve listen address: %w", err)
	}

	// Create UDP listener
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on UDP: %w", err)
	}

	ds.conn = conn
	ds.logger.Info("DERP server started", zap.String("addr", ds.listenAddr))

	// Start packet handling
	go ds.handlePackets()

	// Start cleanup routine
	go ds.cleanupClients()

	return nil
}

// Stop останавливает DERP сервер
func (ds *DERPServer) Stop() error {
	ds.cancel()

	if ds.conn != nil {
		ds.conn.Close()
	}

	// Close all relays
	ds.mu.Lock()
	ds.clients = make(map[string]*DERPClient)
	ds.relays = make(map[string]*DERPRelay)
	ds.mu.Unlock()

	ds.logger.Info("DERP server stopped")
	return nil
}

// handlePackets обрабатывает входящие пакеты
func (ds *DERPServer) handlePackets() {
	buffer := make([]byte, 4096)

	for {
		select {
		case <-ds.ctx.Done():
			return
		default:
			n, clientAddr, err := ds.conn.ReadFromUDP(buffer)
			if err != nil {
				if ds.ctx.Err() != nil {
					return
				}
				ds.logger.Error("Failed to read UDP packet", zap.Error(err))
				continue
			}

			// Handle packet
			go ds.handlePacket(buffer[:n], clientAddr)
		}
	}
}

// handlePacket обрабатывает один пакет
func (ds *DERPServer) handlePacket(data []byte, clientAddr *net.UDPAddr) {
	ds.logger.Debug("Received packet",
		zap.String("from", clientAddr.String()),
		zap.Int("size", len(data)))

	if len(data) < 32 { // Minimum packet size
		ds.logger.Debug("Packet too small", zap.Int("size", len(data)))
		return
	}

	// Parse packet header
	header, err := ds.parsePacketHeader(data)
	if err != nil {
		ds.logger.Debug("Failed to parse packet header", zap.Error(err))
		return
	}

	// Update metrics
	ds.totalPackets++

	// Handle different packet types
	switch header.Type {
	case DERPPacketTypeConnect:
		ds.handleConnectPacket(header, data[32:], clientAddr)
	case DERPPacketTypeData:
		ds.handleDataPacket(header, data[32:], clientAddr)
	case DERPPacketTypeAck:
		ds.handleAckPacket(header, data[32:], clientAddr)
	case DERPPacketTypeDisconnect:
		ds.handleDisconnectPacket(header, data[32:], clientAddr)
	case DERPPacketTypePing:
		ds.handlePingPacket(header, data[32:], clientAddr)
	case DERPPacketTypePong:
		ds.handlePongPacket(header, data[32:], clientAddr)
	default:
		ds.logger.Debug("Unknown packet type", zap.Uint8("type", header.Type))
	}
}

// parsePacketHeader парсит заголовок DERP пакета
func (ds *DERPServer) parsePacketHeader(data []byte) (*DERPPacketHeader, error) {
	if len(data) < 32 {
		return nil, fmt.Errorf("packet too small")
	}

	header := &DERPPacketHeader{
		Type:      data[0],
		Length:    binary.BigEndian.Uint16(data[1:3]),
		Sequence:  binary.BigEndian.Uint32(data[3:7]),
		Timestamp: binary.BigEndian.Uint64(data[7:15]),
	}

	copy(header.ClientID[:], data[15:31])
	copy(header.TargetID[:], data[31:47])

	return header, nil
}

// handleConnectPacket обрабатывает CONNECT пакет
func (ds *DERPServer) handleConnectPacket(header *DERPPacketHeader, payload []byte, clientAddr *net.UDPAddr) {
	clientID := fmt.Sprintf("%x", header.ClientID)

	ds.logger.Info("Client connecting",
		zap.String("client_id", clientID),
		zap.String("addr", clientAddr.String()))

	// Create or update client
	client := &DERPClient{
		ID:          clientID,
		Addr:        clientAddr,
		PublicKey:   payload, // In real implementation, this would be a proper public key
		ConnectedAt: time.Now(),
		LastSeen:    time.Now(),
	}

	ds.mu.Lock()
	ds.clients[clientID] = client
	ds.activeClients++
	ds.mu.Unlock()

	// Send connection acknowledgment
	ds.sendAckPacket(clientID, header.Sequence, clientAddr)

	ds.logger.Info("Client connected",
		zap.String("client_id", clientID),
		zap.String("addr", clientAddr.String()))
}

// handleDataPacket обрабатывает DATA пакет
func (ds *DERPServer) handleDataPacket(header *DERPPacketHeader, payload []byte, clientAddr *net.UDPAddr) {
	clientID := fmt.Sprintf("%x", header.ClientID)
	targetID := fmt.Sprintf("%x", header.TargetID)

	ds.logger.Debug("Data packet received",
		zap.String("from", clientID),
		zap.String("to", targetID),
		zap.Int("size", len(payload)))

	// Update client last seen
	ds.mu.Lock()
	client, exists := ds.clients[clientID]
	if exists {
		client.LastSeen = time.Now()
		client.PacketsSent++
	}
	ds.mu.Unlock()

	if !exists {
		ds.logger.Debug("Unknown client", zap.String("client_id", clientID))
		return
	}

	// Find target client
	ds.mu.RLock()
	targetClient, exists := ds.clients[targetID]
	ds.mu.RUnlock()

	if !exists {
		ds.logger.Debug("Target client not found", zap.String("target_id", targetID))
		// Send error response
		ds.sendErrorPacket(clientID, "target_not_found", clientAddr)
		return
	}

	// Create or update relay
	relayID := ds.generateRelayID(clientID, targetID)
	ds.mu.Lock()
	relay, exists := ds.relays[relayID]
	if !exists {
		relay = &DERPRelay{
			ID:        relayID,
			ClientA:   client,
			ClientB:   targetClient,
			CreatedAt: time.Now(),
			LastUsed:  time.Now(),
		}
		ds.relays[relayID] = relay
	}
	relay.LastUsed = time.Now()
	relay.Packets++
	ds.mu.Unlock()

	// Relay data to target
	ds.relayDataToTarget(targetClient, header, payload)

	// Update metrics
	ds.relayedPackets++

	ds.logger.Debug("Data relayed",
		zap.String("from", clientID),
		zap.String("to", targetID),
		zap.Int("size", len(payload)))
}

// handleAckPacket обрабатывает ACK пакет
func (ds *DERPServer) handleAckPacket(header *DERPPacketHeader, payload []byte, clientAddr *net.UDPAddr) {
	clientID := fmt.Sprintf("%x", header.ClientID)

	ds.logger.Debug("ACK received",
		zap.String("client_id", clientID),
		zap.Uint32("sequence", header.Sequence))

	// Update client last seen
	ds.mu.Lock()
	if client, exists := ds.clients[clientID]; exists {
		client.LastSeen = time.Now()
		client.PacketsRecv++
	}
	ds.mu.Unlock()
}

// handleDisconnectPacket обрабатывает DISCONNECT пакет
func (ds *DERPServer) handleDisconnectPacket(header *DERPPacketHeader, payload []byte, clientAddr *net.UDPAddr) {
	clientID := fmt.Sprintf("%x", header.ClientID)

	ds.logger.Info("Client disconnecting", zap.String("client_id", clientID))

	// Remove client and associated relays
	ds.mu.Lock()
	delete(ds.clients, clientID)
	ds.activeClients--

	// Remove relays involving this client
	for relayID, relay := range ds.relays {
		if relay.ClientA.ID == clientID || relay.ClientB.ID == clientID {
			delete(ds.relays, relayID)
		}
	}
	ds.mu.Unlock()

	ds.logger.Info("Client disconnected", zap.String("client_id", clientID))
}

// handlePingPacket обрабатывает PING пакет
func (ds *DERPServer) handlePingPacket(header *DERPPacketHeader, payload []byte, clientAddr *net.UDPAddr) {
	clientID := fmt.Sprintf("%x", header.ClientID)

	ds.logger.Debug("PING received", zap.String("client_id", clientID))

	// Update client last seen
	ds.mu.Lock()
	if client, exists := ds.clients[clientID]; exists {
		client.LastSeen = time.Now()
	}
	ds.mu.Unlock()

	// Send PONG response
	ds.sendPongPacket(clientID, header.Sequence, clientAddr)
}

// handlePongPacket обрабатывает PONG пакет
func (ds *DERPServer) handlePongPacket(header *DERPPacketHeader, payload []byte, clientAddr *net.UDPAddr) {
	clientID := fmt.Sprintf("%x", header.ClientID)

	ds.logger.Debug("PONG received", zap.String("client_id", clientID))

	// Update client last seen
	ds.mu.Lock()
	if client, exists := ds.clients[clientID]; exists {
		client.LastSeen = time.Now()
	}
	ds.mu.Unlock()
}

// relayDataToTarget релеирует данные целевому клиенту
func (ds *DERPServer) relayDataToTarget(targetClient *DERPClient, header *DERPPacketHeader, payload []byte) {
	// Create relay packet
	relayData := make([]byte, 32+len(payload))
	relayData[0] = DERPPacketTypeData
	binary.BigEndian.PutUint16(relayData[1:3], uint16(len(payload)))
	binary.BigEndian.PutUint32(relayData[3:7], header.Sequence)
	binary.BigEndian.PutUint64(relayData[7:15], uint64(time.Now().UnixNano()))

	// Copy client IDs (swap for relay)
	copy(relayData[15:31], header.ClientID[:]) // Original sender
	copy(relayData[31:47], header.TargetID[:]) // Target

	// Copy payload
	copy(relayData[32:], payload)

	// Send to target
	_, err := ds.conn.WriteToUDP(relayData, targetClient.Addr)
	if err != nil {
		ds.logger.Error("Failed to relay data to target",
			zap.String("target", targetClient.ID),
			zap.Error(err))
		return
	}

	// Update target client stats
	ds.mu.Lock()
	targetClient.PacketsRecv++
	ds.mu.Unlock()
}

// sendAckPacket отправляет ACK пакет
func (ds *DERPServer) sendAckPacket(clientID string, sequence uint32, clientAddr *net.UDPAddr) {
	packet := make([]byte, 32)
	packet[0] = DERPPacketTypeAck
	binary.BigEndian.PutUint32(packet[3:7], sequence)
	binary.BigEndian.PutUint64(packet[7:15], uint64(time.Now().UnixNano()))

	// Copy client ID
	clientIDBytes := []byte(clientID)
	copy(packet[15:31], clientIDBytes)

	_, err := ds.conn.WriteToUDP(packet, clientAddr)
	if err != nil {
		ds.logger.Error("Failed to send ACK", zap.Error(err))
	}
}

// sendErrorPacket отправляет ERROR пакет
func (ds *DERPServer) sendErrorPacket(clientID string, errorMsg string, clientAddr *net.UDPAddr) {
	packet := make([]byte, 32+len(errorMsg))
	packet[0] = 0xFF // Error packet type
	binary.BigEndian.PutUint16(packet[1:3], uint16(len(errorMsg)))
	binary.BigEndian.PutUint64(packet[7:15], uint64(time.Now().UnixNano()))

	// Copy client ID
	clientIDBytes := []byte(clientID)
	copy(packet[15:31], clientIDBytes)

	// Copy error message
	copy(packet[32:], []byte(errorMsg))

	_, err := ds.conn.WriteToUDP(packet, clientAddr)
	if err != nil {
		ds.logger.Error("Failed to send error packet", zap.Error(err))
	}
}

// sendPongPacket отправляет PONG пакет
func (ds *DERPServer) sendPongPacket(clientID string, sequence uint32, clientAddr *net.UDPAddr) {
	packet := make([]byte, 32)
	packet[0] = DERPPacketTypePong
	binary.BigEndian.PutUint32(packet[3:7], sequence)
	binary.BigEndian.PutUint64(packet[7:15], uint64(time.Now().UnixNano()))

	// Copy client ID
	clientIDBytes := []byte(clientID)
	copy(packet[15:31], clientIDBytes)

	_, err := ds.conn.WriteToUDP(packet, clientAddr)
	if err != nil {
		ds.logger.Error("Failed to send PONG", zap.Error(err))
	}
}

// generateRelayID генерирует ID для relay соединения
func (ds *DERPServer) generateRelayID(clientA, clientB string) string {
	if clientA < clientB {
		return fmt.Sprintf("%s-%s", clientA, clientB)
	}
	return fmt.Sprintf("%s-%s", clientB, clientA)
}

// cleanupClients очищает неактивных клиентов
func (ds *DERPServer) cleanupClients() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ds.ctx.Done():
			return
		case <-ticker.C:
			ds.mu.Lock()
			now := time.Now()

			// Remove inactive clients (no activity for 5 minutes)
			for clientID, client := range ds.clients {
				if now.Sub(client.LastSeen) > 5*time.Minute {
					delete(ds.clients, clientID)
					ds.activeClients--
					ds.logger.Info("Cleaned up inactive client", zap.String("client_id", clientID))
				}
			}

			// Remove inactive relays
			for relayID, relay := range ds.relays {
				if now.Sub(relay.LastUsed) > 10*time.Minute {
					delete(ds.relays, relayID)
					ds.logger.Info("Cleaned up inactive relay", zap.String("relay_id", relayID))
				}
			}

			ds.mu.Unlock()
		}
	}
}

// GetStats возвращает статистику DERP сервера
func (ds *DERPServer) GetStats() map[string]interface{} {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	return map[string]interface{}{
		"total_packets":   ds.totalPackets,
		"relayed_packets": ds.relayedPackets,
		"active_clients":  ds.activeClients,
		"total_relays":    len(ds.relays),
		"clients":         len(ds.clients),
	}
}

// GetClientList возвращает список активных клиентов
func (ds *DERPServer) GetClientList() []*DERPClient {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	clients := make([]*DERPClient, 0, len(ds.clients))
	for _, client := range ds.clients {
		clients = append(clients, client)
	}

	return clients
}

// GetRelayList возвращает список активных relay соединений
func (ds *DERPServer) GetRelayList() []*DERPRelay {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	relays := make([]*DERPRelay, 0, len(ds.relays))
	for _, relay := range ds.relays {
		relays = append(relays, relay)
	}

	return relays
}
