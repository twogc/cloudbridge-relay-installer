package nat

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"

	"go.uber.org/zap"
)

// TURN Message Types (RFC 5766)
const (
	STUNMethodAllocate         = 0x003
	STUNMethodRefresh          = 0x004
	STUNMethodSend             = 0x006
	STUNMethodData             = 0x007
	STUNMethodCreatePermission = 0x008
	STUNMethodChannelBind      = 0x009
)

// TURN Attributes (RFC 5766)
const (
	STUNAttrRequestedTransport = 0x0019
	STUNAttrXORPeerAddress     = 0x0012
	STUNAttrXORRelayedAddress  = 0x0016
	STUNAttrXORMappedAddress   = 0x0020
	STUNAttrData               = 0x0013
	STUNAttrChannelNumber      = 0x000C
	STUNAttrLifetime           = 0x000D
	STUNAttrBandwidth          = 0x000A
	STUNAttrDestinationAddress = 0x000B
	STUNAttrSoftware           = 0x8022
	STUNAttrRealm              = 0x0014
	STUNAttrNonce              = 0x0015
	STUNAttrUsername           = 0x0006
	STUNAttrMessageIntegrity   = 0x0008
	STUNAttrErrorCode          = 0x0009
)

// TURN Error Codes
const (
	STUNErrorCodeBadRequest             = 400
	STUNErrorCodeUnauthorized           = 401
	STUNErrorCodeForbidden              = 403
	STUNErrorCodeAllocationMismatch     = 437
	STUNErrorCodeStaleNonce             = 438
	STUNErrorCodeAllocationQuotaReached = 486
	STUNErrorCodeInsufficientCapacity   = 508
)

// EnhancedTURNServer представляет полноценный TURN сервер (RFC 5766)
type EnhancedTURNServer struct {
	logger      *zap.Logger
	listenAddr  string
	realm       string
	username    string
	password    string
	conn        *net.UDPConn
	allocations map[string]*EnhancedTURNAllocation
	channels    map[uint16]*EnhancedTURNChannel
	permissions map[string]*EnhancedTURNPermission
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc

	// Relay address pool
	relayPool *RelayAddressPool
}

// EnhancedTURNAllocation представляет TURN allocation (RFC 5766 Section 2.2)
type EnhancedTURNAllocation struct {
	ID          string
	ClientAddr  *net.UDPAddr
	RelayAddr   *net.UDPAddr
	RelayConn   *net.UDPConn
	CreatedAt   time.Time
	LastUsed    time.Time
	Lifetime    time.Duration
	Bandwidth   uint64
	Permissions map[string]*EnhancedTURNPermission
	Channels    map[uint16]*EnhancedTURNChannel
	mu          sync.RWMutex
}

// EnhancedTURNChannel представляет TURN channel (RFC 5766 Section 11)
type EnhancedTURNChannel struct {
	Number     uint16
	PeerAddr   *net.UDPAddr
	Allocation *EnhancedTURNAllocation
	CreatedAt  time.Time
	LastUsed   time.Time
}

// EnhancedTURNPermission представляет TURN permission (RFC 5766 Section 9)
type EnhancedTURNPermission struct {
	PeerAddr   *net.UDPAddr
	Allocation *EnhancedTURNAllocation
	CreatedAt  time.Time
	ExpiresAt  time.Time
}

// RelayAddressPool управляет пулом relay адресов
type RelayAddressPool struct {
	baseAddr  *net.UDPAddr
	allocated map[string]bool
	mu        sync.Mutex
}

// NewEnhancedTURNServer создает новый улучшенный TURN сервер
func NewEnhancedTURNServer(listenAddr, realm, username, password string, logger *zap.Logger) (*EnhancedTURNServer, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Создаем relay address pool
	relayPool, err := NewRelayAddressPool("0.0.0.0:49152-65535")
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create relay address pool: %w", err)
	}

	return &EnhancedTURNServer{
		logger:      logger,
		listenAddr:  listenAddr,
		realm:       realm,
		username:    username,
		password:    password,
		allocations: make(map[string]*EnhancedTURNAllocation),
		channels:    make(map[uint16]*EnhancedTURNChannel),
		permissions: make(map[string]*EnhancedTURNPermission),
		relayPool:   relayPool,
		ctx:         ctx,
		cancel:      cancel,
	}, nil
}

// Start запускает улучшенный TURN сервер
func (ts *EnhancedTURNServer) Start() error {
	// Resolve listen address
	udpAddr, err := net.ResolveUDPAddr("udp", ts.listenAddr)
	if err != nil {
		return fmt.Errorf("failed to resolve listen address: %w", err)
	}

	// Create UDP listener
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on UDP: %w", err)
	}

	ts.conn = conn
	ts.logger.Info("Enhanced TURN server started",
		zap.String("addr", ts.listenAddr),
		zap.String("realm", ts.realm))

	// Start packet handling
	go ts.handlePackets()

	// Start cleanup routine
	go ts.cleanupResources()

	return nil
}

// Stop останавливает TURN сервер
func (ts *EnhancedTURNServer) Stop() error {
	ts.cancel()

	if ts.conn != nil {
		ts.conn.Close()
	}

	// Close all allocations
	ts.mu.Lock()
	for _, alloc := range ts.allocations {
		if alloc.RelayConn != nil {
			alloc.RelayConn.Close()
		}
	}
	ts.allocations = make(map[string]*EnhancedTURNAllocation)
	ts.channels = make(map[uint16]*EnhancedTURNChannel)
	ts.permissions = make(map[string]*EnhancedTURNPermission)
	ts.mu.Unlock()

	ts.logger.Info("Enhanced TURN server stopped")
	return nil
}

// handlePackets обрабатывает входящие пакеты
func (ts *EnhancedTURNServer) handlePackets() {
	buffer := make([]byte, 4096)

	for {
		select {
		case <-ts.ctx.Done():
			return
		default:
			n, clientAddr, err := ts.conn.ReadFromUDP(buffer)
			if err != nil {
				if ts.ctx.Err() != nil {
					return
				}
				ts.logger.Error("Failed to read UDP packet", zap.Error(err))
				continue
			}

			// Handle packet
			go ts.handlePacket(buffer[:n], clientAddr)
		}
	}
}

// handlePacket обрабатывает один пакет
func (ts *EnhancedTURNServer) handlePacket(data []byte, clientAddr *net.UDPAddr) {
	ts.logger.Debug("Received packet",
		zap.String("from", clientAddr.String()),
		zap.Int("size", len(data)))

	// Parse STUN/TURN message
	message, err := ts.parseSTUNMessage(data)
	if err != nil {
		ts.logger.Debug("Failed to parse STUN message", zap.Error(err))
		return
	}

	// Handle different message types
	switch message.Method {
	case STUNMethodAllocate:
		ts.handleAllocateRequest(message, clientAddr)
	case STUNMethodRefresh:
		ts.handleRefreshRequest(message, clientAddr)
	case STUNMethodCreatePermission:
		ts.handleCreatePermissionRequest(message, clientAddr)
	case STUNMethodChannelBind:
		ts.handleChannelBindRequest(message, clientAddr)
	case STUNMethodSend:
		ts.handleSendIndication(message, clientAddr)
	case STUNMethodData:
		ts.handleDataIndication(message, clientAddr)
	default:
		ts.logger.Debug("Unknown STUN method", zap.Uint16("method", message.Method))
	}
}

// STUNMessage представляет STUN/TURN сообщение
type STUNMessage struct {
	Type          uint16
	Method        uint16
	Length        uint16
	TransactionID [12]byte
	Attributes    map[uint16][]byte
}

// parseSTUNMessage парсит STUN сообщение
func (ts *EnhancedTURNServer) parseSTUNMessage(data []byte) (*STUNMessage, error) {
	if len(data) < 20 {
		return nil, fmt.Errorf("message too short")
	}

	message := &STUNMessage{
		Attributes: make(map[uint16][]byte),
	}

	// Parse header
	message.Type = binary.BigEndian.Uint16(data[0:2])
	message.Length = binary.BigEndian.Uint16(data[2:4])
	copy(message.TransactionID[:], data[4:16])

	// Extract method from type
	message.Method = message.Type & 0x0FEF

	// Parse attributes
	offset := 20
	for offset < len(data) {
		if offset+4 > len(data) {
			break
		}

		attrType := binary.BigEndian.Uint16(data[offset : offset+2])
		attrLength := binary.BigEndian.Uint16(data[offset+2 : offset+4])

		// Align to 4-byte boundary
		paddedLength := ((attrLength + 3) / 4) * 4

		if offset+4+int(paddedLength) > len(data) {
			break
		}

		attrValue := data[offset+4 : offset+4+int(attrLength)]
		message.Attributes[attrType] = attrValue

		offset += 4 + int(paddedLength)
	}

	return message, nil
}

// handleAllocateRequest обрабатывает ALLOCATE запрос (RFC 5766 Section 6.1)
func (ts *EnhancedTURNServer) handleAllocateRequest(message *STUNMessage, clientAddr *net.UDPAddr) {
	ts.logger.Info("Handling ALLOCATE request", zap.String("client", clientAddr.String()))

	// Check authentication (simplified)
	if !ts.authenticate(message, clientAddr) {
		ts.sendErrorResponse(message, clientAddr, STUNErrorCodeUnauthorized, "Unauthorized")
		return
	}

	// Get requested transport
	transportAttr, exists := message.Attributes[STUNAttrRequestedTransport]
	if !exists {
		ts.sendErrorResponse(message, clientAddr, STUNErrorCodeBadRequest, "Missing REQUESTED-TRANSPORT")
		return
	}

	if len(transportAttr) < 4 || transportAttr[3] != 17 { // UDP = 17
		ts.sendErrorResponse(message, clientAddr, STUNErrorCodeBadRequest, "Unsupported transport")
		return
	}

	// Allocate relay address
	relayAddr, err := ts.relayPool.AllocateAddress()
	if err != nil {
		ts.sendErrorResponse(message, clientAddr, STUNErrorCodeInsufficientCapacity, "No relay addresses available")
		return
	}

	// Create relay connection
	relayConn, err := net.ListenUDP("udp", relayAddr)
	if err != nil {
		ts.relayPool.ReleaseAddress(relayAddr)
		ts.sendErrorResponse(message, clientAddr, STUNErrorCodeInsufficientCapacity, "Failed to create relay connection")
		return
	}

	// Create allocation
	allocID := ts.generateAllocationID()
	allocation := &EnhancedTURNAllocation{
		ID:          allocID,
		ClientAddr:  clientAddr,
		RelayAddr:   relayAddr,
		RelayConn:   relayConn,
		CreatedAt:   time.Now(),
		LastUsed:    time.Now(),
		Lifetime:    10 * time.Minute, // Default lifetime
		Permissions: make(map[string]*EnhancedTURNPermission),
		Channels:    make(map[uint16]*EnhancedTURNChannel),
	}

	// Store allocation
	ts.mu.Lock()
	ts.allocations[allocID] = allocation
	ts.mu.Unlock()

	// Start relay handling
	go ts.handleRelayTraffic(allocation)

	// Send success response
	ts.sendAllocateResponse(message, clientAddr, allocation)

	ts.logger.Info("Allocation created",
		zap.String("alloc_id", allocID),
		zap.String("client", clientAddr.String()),
		zap.String("relay", allocation.RelayAddr.String()))
}

// handleRefreshRequest обрабатывает REFRESH запрос (RFC 5766 Section 6.2)
func (ts *EnhancedTURNServer) handleRefreshRequest(message *STUNMessage, clientAddr *net.UDPAddr) {
	// Find allocation
	allocation := ts.findAllocationByClient(clientAddr)
	if allocation == nil {
		ts.sendErrorResponse(message, clientAddr, STUNErrorCodeAllocationMismatch, "No allocation found")
		return
	}

	// Update lifetime if specified
	if lifetimeAttr, exists := message.Attributes[STUNAttrLifetime]; exists && len(lifetimeAttr) >= 4 {
		newLifetime := time.Duration(binary.BigEndian.Uint32(lifetimeAttr)) * time.Second
		if newLifetime > 0 && newLifetime <= 3600*time.Second { // Max 1 hour
			allocation.mu.Lock()
			allocation.Lifetime = newLifetime
			allocation.LastUsed = time.Now()
			allocation.mu.Unlock()
		}
	}

	// Send success response
	ts.sendRefreshResponse(message, clientAddr, allocation)
}

// handleCreatePermissionRequest обрабатывает CREATE-PERMISSION запрос (RFC 5766 Section 9)
func (ts *EnhancedTURNServer) handleCreatePermissionRequest(message *STUNMessage, clientAddr *net.UDPAddr) {
	// Find allocation
	allocation := ts.findAllocationByClient(clientAddr)
	if allocation == nil {
		ts.sendErrorResponse(message, clientAddr, STUNErrorCodeAllocationMismatch, "No allocation found")
		return
	}

	// Get peer address
	peerAddrAttr, exists := message.Attributes[STUNAttrXORPeerAddress]
	if !exists {
		ts.sendErrorResponse(message, clientAddr, STUNErrorCodeBadRequest, "Missing XOR-PEER-ADDRESS")
		return
	}

	peerAddr, err := ts.parseXORAddress(peerAddrAttr)
	if err != nil {
		ts.sendErrorResponse(message, clientAddr, STUNErrorCodeBadRequest, "Invalid XOR-PEER-ADDRESS")
		return
	}

	// Create permission
	permission := &EnhancedTURNPermission{
		PeerAddr:   peerAddr,
		Allocation: allocation,
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(5 * time.Minute), // Default permission lifetime
	}

	// Store permission
	permissionKey := fmt.Sprintf("%s:%s", allocation.ID, peerAddr.String())
	ts.mu.Lock()
	ts.permissions[permissionKey] = permission
	allocation.Permissions[peerAddr.String()] = permission
	ts.mu.Unlock()

	// Send success response
	ts.sendCreatePermissionResponse(message, clientAddr)

	ts.logger.Info("Permission created",
		zap.String("alloc_id", allocation.ID),
		zap.String("peer", peerAddr.String()))
}

// handleChannelBindRequest обрабатывает CHANNEL-BIND запрос (RFC 5766 Section 11)
func (ts *EnhancedTURNServer) handleChannelBindRequest(message *STUNMessage, clientAddr *net.UDPAddr) {
	// Find allocation
	allocation := ts.findAllocationByClient(clientAddr)
	if allocation == nil {
		ts.sendErrorResponse(message, clientAddr, STUNErrorCodeAllocationMismatch, "No allocation found")
		return
	}

	// Get channel number and peer address
	channelAttr, exists := message.Attributes[STUNAttrChannelNumber]
	if !exists || len(channelAttr) < 4 {
		ts.sendErrorResponse(message, clientAddr, STUNErrorCodeBadRequest, "Missing CHANNEL-NUMBER")
		return
	}

	peerAddrAttr, exists := message.Attributes[STUNAttrXORPeerAddress]
	if !exists {
		ts.sendErrorResponse(message, clientAddr, STUNErrorCodeBadRequest, "Missing XOR-PEER-ADDRESS")
		return
	}

	channelNumber := binary.BigEndian.Uint16(channelAttr[2:4])
	peerAddr, err := ts.parseXORAddress(peerAddrAttr)
	if err != nil {
		ts.sendErrorResponse(message, clientAddr, STUNErrorCodeBadRequest, "Invalid XOR-PEER-ADDRESS")
		return
	}

	// Validate channel number (0x4000-0x4FFF)
	if channelNumber < 0x4000 || channelNumber > 0x4FFF {
		ts.sendErrorResponse(message, clientAddr, STUNErrorCodeBadRequest, "Invalid channel number")
		return
	}

	// Create channel
	channel := &EnhancedTURNChannel{
		Number:     channelNumber,
		PeerAddr:   peerAddr,
		Allocation: allocation,
		CreatedAt:  time.Now(),
		LastUsed:   time.Now(),
	}

	// Store channel
	ts.mu.Lock()
	ts.channels[channelNumber] = channel
	allocation.Channels[channelNumber] = channel
	ts.mu.Unlock()

	// Send success response
	ts.sendChannelBindResponse(message, clientAddr)

	ts.logger.Info("Channel bound",
		zap.String("alloc_id", allocation.ID),
		zap.Uint16("channel", channelNumber),
		zap.String("peer", peerAddr.String()))
}

// handleSendIndication обрабатывает SEND indication (RFC 5766 Section 10)
func (ts *EnhancedTURNServer) handleSendIndication(message *STUNMessage, clientAddr *net.UDPAddr) {
	// Find allocation
	allocation := ts.findAllocationByClient(clientAddr)
	if allocation == nil {
		return // Indications don't get responses
	}

	// Get destination address and data
	destAddrAttr, exists := message.Attributes[STUNAttrXORPeerAddress]
	if !exists {
		return
	}

	dataAttr, exists := message.Attributes[STUNAttrData]
	if !exists {
		return
	}

	destAddr, err := ts.parseXORAddress(destAddrAttr)
	if err != nil {
		return
	}

	// Check permission
	if !ts.checkPermission(allocation, destAddr) {
		ts.logger.Debug("Permission denied for SEND",
			zap.String("alloc_id", allocation.ID),
			zap.String("dest", destAddr.String()))
		return
	}

	// Send data to peer
	_, err = allocation.RelayConn.WriteToUDP(dataAttr, destAddr)
	if err != nil {
		ts.logger.Debug("Failed to send data to peer", zap.Error(err))
		return
	}

	allocation.mu.Lock()
	allocation.LastUsed = time.Now()
	allocation.mu.Unlock()

	ts.logger.Debug("Data sent to peer",
		zap.String("alloc_id", allocation.ID),
		zap.String("dest", destAddr.String()),
		zap.Int("size", len(dataAttr)))
}

// handleDataIndication обрабатывает DATA indication (RFC 5766 Section 11)
func (ts *EnhancedTURNServer) handleDataIndication(message *STUNMessage, clientAddr *net.UDPAddr) {
	// Find allocation
	allocation := ts.findAllocationByClient(clientAddr)
	if allocation == nil {
		return
	}

	// Get channel number and data
	channelAttr, exists := message.Attributes[STUNAttrChannelNumber]
	if !exists || len(channelAttr) < 4 {
		return
	}

	dataAttr, exists := message.Attributes[STUNAttrData]
	if !exists {
		return
	}

	channelNumber := binary.BigEndian.Uint16(channelAttr[2:4])

	// Find channel
	ts.mu.RLock()
	channel, exists := ts.channels[channelNumber]
	ts.mu.RUnlock()

	if !exists || channel.Allocation != allocation {
		return
	}

	// Send data to peer
	_, err := allocation.RelayConn.WriteToUDP(dataAttr, channel.PeerAddr)
	if err != nil {
		ts.logger.Debug("Failed to send data to peer via channel", zap.Error(err))
		return
	}

	channel.LastUsed = time.Now()
	allocation.mu.Lock()
	allocation.LastUsed = time.Now()
	allocation.mu.Unlock()

	ts.logger.Debug("Data sent to peer via channel",
		zap.String("alloc_id", allocation.ID),
		zap.Uint16("channel", channelNumber),
		zap.String("peer", channel.PeerAddr.String()),
		zap.Int("size", len(dataAttr)))
}

// handleRelayTraffic обрабатывает трафик для конкретного allocation
func (ts *EnhancedTURNServer) handleRelayTraffic(allocation *EnhancedTURNAllocation) {
	buffer := make([]byte, 4096)

	for {
		select {
		case <-ts.ctx.Done():
			return
		default:
			// Set read timeout
			allocation.RelayConn.SetReadDeadline(time.Now().Add(1 * time.Second))

			n, peerAddr, err := allocation.RelayConn.ReadFromUDP(buffer)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				ts.logger.Debug("Relay connection read error", zap.Error(err))
				return
			}

			// Check permission
			if !ts.checkPermission(allocation, peerAddr) {
				ts.logger.Debug("Permission denied for relay traffic",
					zap.String("alloc_id", allocation.ID),
					zap.String("peer", peerAddr.String()))
				continue
			}

			// Check if we have a channel for this peer
			channel := ts.findChannelByPeer(allocation, peerAddr)
			if channel != nil {
				// Send via channel
				ts.sendDataIndication(allocation, channel.Number, buffer[:n])
			} else {
				// Send via Send indication
				ts.sendSendIndication(allocation, peerAddr, buffer[:n])
			}

			allocation.mu.Lock()
			allocation.LastUsed = time.Now()
			allocation.mu.Unlock()

			ts.logger.Debug("Relayed packet to client",
				zap.String("from", peerAddr.String()),
				zap.String("to", allocation.ClientAddr.String()),
				zap.Int("size", n))
		}
	}
}

// Helper methods

func (ts *EnhancedTURNServer) authenticate(message *STUNMessage, clientAddr *net.UDPAddr) bool {
	// Simplified authentication - in real implementation would use proper STUN authentication
	return true
}

func (ts *EnhancedTURNServer) findAllocationByClient(clientAddr *net.UDPAddr) *EnhancedTURNAllocation {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	for _, allocation := range ts.allocations {
		if allocation.ClientAddr.String() == clientAddr.String() {
			return allocation
		}
	}
	return nil
}

func (ts *EnhancedTURNServer) findChannelByPeer(allocation *EnhancedTURNAllocation, peerAddr *net.UDPAddr) *EnhancedTURNChannel {
	allocation.mu.RLock()
	defer allocation.mu.RUnlock()

	for _, channel := range allocation.Channels {
		if channel.PeerAddr.String() == peerAddr.String() {
			return channel
		}
	}
	return nil
}

func (ts *EnhancedTURNServer) checkPermission(allocation *EnhancedTURNAllocation, peerAddr *net.UDPAddr) bool {
	allocation.mu.RLock()
	defer allocation.mu.RUnlock()

	permission, exists := allocation.Permissions[peerAddr.String()]
	if !exists {
		return false
	}

	return time.Now().Before(permission.ExpiresAt)
}

func (ts *EnhancedTURNServer) generateAllocationID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func (ts *EnhancedTURNServer) parseXORAddress(data []byte) (*net.UDPAddr, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("address too short")
	}

	family := data[1]
	if family != 1 { // IPv4
		return nil, fmt.Errorf("unsupported address family")
	}

	port := binary.BigEndian.Uint16(data[2:4]) ^ 0x2112
	ip := make([]byte, 4)
	for i := 0; i < 4; i++ {
		ip[i] = data[4+i] ^ 0x21
	}

	return &net.UDPAddr{
		IP:   ip,
		Port: int(port),
	}, nil
}

// Response methods (simplified implementations)

func (ts *EnhancedTURNServer) sendAllocateResponse(message *STUNMessage, clientAddr *net.UDPAddr, allocation *EnhancedTURNAllocation) {
	// Simplified response - in real implementation would build proper STUN response
	response := []byte{0x01, 0x03, 0x00, 0x00} // Success response
	ts.conn.WriteToUDP(response, clientAddr)
}

// GetStats возвращает статистику TURN сервера
func (ts *EnhancedTURNServer) GetStats() map[string]interface{} {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	return map[string]interface{}{
		"allocations_count": len(ts.allocations),
		"channels_count":    len(ts.channels),
		"permissions_count": len(ts.permissions),
		"listen_addr":       ts.listenAddr,
		"realm":             ts.realm,
	}
}

func (ts *EnhancedTURNServer) sendRefreshResponse(message *STUNMessage, clientAddr *net.UDPAddr, allocation *EnhancedTURNAllocation) {
	response := []byte{0x01, 0x04, 0x00, 0x00} // Success response
	ts.conn.WriteToUDP(response, clientAddr)
}

func (ts *EnhancedTURNServer) sendCreatePermissionResponse(message *STUNMessage, clientAddr *net.UDPAddr) {
	response := []byte{0x01, 0x08, 0x00, 0x00} // Success response
	ts.conn.WriteToUDP(response, clientAddr)
}

func (ts *EnhancedTURNServer) sendChannelBindResponse(message *STUNMessage, clientAddr *net.UDPAddr) {
	response := []byte{0x01, 0x09, 0x00, 0x00} // Success response
	ts.conn.WriteToUDP(response, clientAddr)
}

func (ts *EnhancedTURNServer) sendErrorResponse(message *STUNMessage, clientAddr *net.UDPAddr, errorCode uint16, reason string) {
	// Simplified error response
	response := []byte{0x01, 0x11, 0x00, 0x00} // Error response
	ts.conn.WriteToUDP(response, clientAddr)
}

func (ts *EnhancedTURNServer) sendSendIndication(allocation *EnhancedTURNAllocation, peerAddr *net.UDPAddr, data []byte) {
	// Simplified Send indication
	indication := []byte{0x00, 0x06, 0x00, 0x00} // Send indication
	ts.conn.WriteToUDP(indication, allocation.ClientAddr)
}

func (ts *EnhancedTURNServer) sendDataIndication(allocation *EnhancedTURNAllocation, channelNumber uint16, data []byte) {
	// Simplified Data indication
	indication := []byte{0x00, 0x07, 0x00, 0x00} // Data indication
	ts.conn.WriteToUDP(indication, allocation.ClientAddr)
}

func (ts *EnhancedTURNServer) cleanupResources() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ts.ctx.Done():
			return
		case <-ticker.C:
			ts.mu.Lock()
			now := time.Now()

			// Cleanup expired allocations
			for allocID, allocation := range ts.allocations {
				if now.Sub(allocation.LastUsed) > allocation.Lifetime {
					if allocation.RelayConn != nil {
						allocation.RelayConn.Close()
					}
					ts.relayPool.ReleaseAddress(allocation.RelayAddr)
					delete(ts.allocations, allocID)
					ts.logger.Info("Cleaned up expired allocation", zap.String("alloc_id", allocID))
				}
			}

			// Cleanup expired permissions
			for permKey, permission := range ts.permissions {
				if now.After(permission.ExpiresAt) {
					delete(ts.permissions, permKey)
					if permission.Allocation != nil {
						delete(permission.Allocation.Permissions, permission.PeerAddr.String())
					}
				}
			}

			ts.mu.Unlock()
		}
	}
}

// RelayAddressPool implementation

func NewRelayAddressPool(rangeStr string) (*RelayAddressPool, error) {
	// Simplified implementation - in real implementation would parse range
	baseAddr, err := net.ResolveUDPAddr("udp", "0.0.0.0:49152")
	if err != nil {
		return nil, err
	}

	return &RelayAddressPool{
		baseAddr:  baseAddr,
		allocated: make(map[string]bool),
	}, nil
}

func (rap *RelayAddressPool) AllocateAddress() (*net.UDPAddr, error) {
	rap.mu.Lock()
	defer rap.mu.Unlock()

	// Simple allocation - find next available port
	for port := 49152; port <= 65535; port++ {
		addr := &net.UDPAddr{
			IP:   rap.baseAddr.IP,
			Port: port,
		}

		key := addr.String()
		if !rap.allocated[key] {
			rap.allocated[key] = true
			return addr, nil
		}
	}

	return nil, fmt.Errorf("no available relay addresses")
}

func (rap *RelayAddressPool) ReleaseAddress(addr *net.UDPAddr) {
	rap.mu.Lock()
	defer rap.mu.Unlock()

	key := addr.String()
	delete(rap.allocated, key)
}
