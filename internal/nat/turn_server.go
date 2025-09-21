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

	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
)

// TURNRelayServer представляет TURN сервер для relay кандидатов
type TURNRelayServer struct {
	logger      *logger.Logger
	config      *TURNConfig
	allocations map[string]*Allocation
	permissions map[string]*Permission
	channels    map[uint16]*Channel

	// Сетевые интерфейсы
	udpListener *net.UDPConn
	tcpListener *net.TCPListener

	// Управление
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
	isRunning bool
}

// TURNConfig конфигурация TURN сервера
type TURNConfig struct {
	ListenAddr     string
	Realm          string
	MinPort        int
	MaxPort        int
	MaxAllocations int
	AuthTimeout    time.Duration
	AllocTimeout   time.Duration
}

// Allocation представляет TURN allocation
type Allocation struct {
	ID           string
	ClientAddr   *net.UDPAddr
	RelayAddr    *net.UDPAddr
	RelayConn    *net.UDPConn
	Permissions  map[string]*Permission
	Channels     map[uint16]*Channel
	CreatedAt    time.Time
	LastActivity time.Time
	Username     string
	Realm        string
	Nonce        string
	Lifetime     time.Duration
}

// Permission представляет TURN permission
type Permission struct {
	ID        string
	PeerAddr  *net.UDPAddr
	CreatedAt time.Time
	Lifetime  time.Duration
}

// Channel представляет TURN channel
type Channel struct {
	ID        uint16
	PeerAddr  *net.UDPAddr
	CreatedAt time.Time
	Lifetime  time.Duration
}

// TURN константы
const (
	TURNDefaultPort     = 3478
	TURNDefaultTlsPort  = 5349
	TURNDefaultMinPort  = 49152
	TURNDefaultMaxPort  = 65535
	TURNDefaultRealm    = "cloudbridge-relay"
	TURNDefaultLifetime = 10 * time.Minute
)

// NewTURNRelayServer создает новый TURN сервер
func NewTURNRelayServer(logger *logger.Logger, config *TURNConfig) *TURNRelayServer {
	if config == nil {
		config = &TURNConfig{
			ListenAddr:     fmt.Sprintf(":%d", TURNDefaultPort),
			Realm:          TURNDefaultRealm,
			MinPort:        TURNDefaultMinPort,
			MaxPort:        TURNDefaultMaxPort,
			MaxAllocations: 100,
			AuthTimeout:    30 * time.Second,
			AllocTimeout:   TURNDefaultLifetime,
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &TURNRelayServer{
		logger:      logger,
		config:      config,
		allocations: make(map[string]*Allocation),
		permissions: make(map[string]*Permission),
		channels:    make(map[uint16]*Channel),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start запускает TURN сервер
func (ts *TURNRelayServer) Start() error {
	ts.logger.Info("Starting TURN server",
		zap.String("listen_addr", ts.config.ListenAddr),
		zap.String("realm", ts.config.Realm))

	// Запускаем UDP listener
	if err := ts.startUDPListener(); err != nil {
		return fmt.Errorf("failed to start UDP listener: %w", err)
	}

	// Запускаем TCP listener
	if err := ts.startTCPListener(); err != nil {
		return fmt.Errorf("failed to start TCP listener: %w", err)
	}

	// Запускаем cleanup goroutine
	go ts.cleanupRoutine()

	ts.isRunning = true
	ts.logger.Info("TURN server started successfully")

	return nil
}

// startUDPListener запускает UDP listener
func (ts *TURNRelayServer) startUDPListener() error {
	addr, err := net.ResolveUDPAddr("udp", ts.config.ListenAddr)
	if err != nil {
		return fmt.Errorf("failed to resolve UDP address: %w", err)
	}

	ts.udpListener, err = net.ListenUDP("udp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on UDP: %w", err)
	}

	ts.logger.Info("UDP listener started", zap.String("addr", ts.udpListener.LocalAddr().String()))

	// Запускаем обработчик UDP сообщений
	go ts.handleUDPMessages()

	return nil
}

// startTCPListener запускает TCP listener
func (ts *TURNRelayServer) startTCPListener() error {
	addr, err := net.ResolveTCPAddr("tcp", ts.config.ListenAddr)
	if err != nil {
		return fmt.Errorf("failed to resolve TCP address: %w", err)
	}

	ts.tcpListener, err = net.ListenTCP("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on TCP: %w", err)
	}

	ts.logger.Info("TCP listener started", zap.String("addr", ts.tcpListener.Addr().String()))

	// Запускаем обработчик TCP соединений
	go ts.handleTCPConnections()

	return nil
}

// handleUDPMessages обрабатывает UDP сообщения
func (ts *TURNRelayServer) handleUDPMessages() {
	buffer := make([]byte, 1500)

	for {
		select {
		case <-ts.ctx.Done():
			return
		default:
			ts.udpListener.SetReadDeadline(time.Now().Add(1 * time.Second))
			n, clientAddr, err := ts.udpListener.ReadFromUDP(buffer)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				ts.logger.Error("Failed to read UDP message", zap.Error(err))
				continue
			}

			// Обрабатываем сообщение
			go ts.processMessage(buffer[:n], clientAddr, ts.udpListener)
		}
	}
}

// handleTCPConnections обрабатывает TCP соединения
func (ts *TURNRelayServer) handleTCPConnections() {
	for {
		select {
		case <-ts.ctx.Done():
			return
		default:
			ts.tcpListener.SetDeadline(time.Now().Add(1 * time.Second))
			conn, err := ts.tcpListener.AcceptTCP()
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				ts.logger.Error("Failed to accept TCP connection", zap.Error(err))
				continue
			}

			// Обрабатываем TCP соединение
			go ts.handleTCPConnection(conn)
		}
	}
}

// processMessage обрабатывает входящее сообщение
func (ts *TURNRelayServer) processMessage(data []byte, clientAddr *net.UDPAddr, conn *net.UDPConn) {
	ts.logger.Debug("Processing message",
		zap.String("client_addr", clientAddr.String()),
		zap.Int("data_len", len(data)))

	// Парсим STUN/TURN сообщение
	message, err := ts.parseMessage(data)
	if err != nil {
		ts.logger.Warn("Failed to parse message", zap.Error(err))
		return
	}

	// Обрабатываем в зависимости от типа сообщения
	switch message.Type {
	case STUNBindingRequest:
		ts.handleBindingRequest(message, clientAddr, conn)
	case STUNAllocateRequest:
		ts.handleAllocateRequest(message, clientAddr, conn)
	case STUNRefreshRequest:
		ts.handleRefreshRequest(message, clientAddr, conn)
	case STUNCreatePermissionRequest:
		ts.handleCreatePermissionRequest(message, clientAddr, conn)
	case STUNChannelBindRequest:
		ts.handleChannelBindRequest(message, clientAddr, conn)
	case STUNSendIndication:
		ts.handleSendIndication(message, clientAddr, conn)
	case STUNDataIndication:
		ts.handleDataIndication(message, clientAddr, conn)
	default:
		ts.logger.Warn("Unknown message type", zap.Uint16("type", message.Type))
	}
}

// handleAllocateRequest обрабатывает TURN Allocate запрос
func (ts *TURNRelayServer) handleAllocateRequest(message *STUNMessage, clientAddr *net.UDPAddr, conn *net.UDPConn) {
	ts.logger.Info("Handling Allocate request",
		zap.String("client_addr", clientAddr.String()))

	// Проверяем аутентификацию
	if !ts.authenticate(message, clientAddr) {
		ts.sendErrorResponse(message, clientAddr, conn, STUNUnauthorized)
		return
	}

	// Создаем allocation
	allocation, err := ts.createAllocation(clientAddr, message.Username, message.Realm)
	if err != nil {
		ts.logger.Error("Failed to create allocation", zap.Error(err))
		ts.sendErrorResponse(message, clientAddr, conn, STUNServerError)
		return
	}

	// Отправляем успешный ответ
	ts.sendAllocateResponse(message, clientAddr, conn, allocation)
}

// createAllocation создает новое TURN allocation
func (ts *TURNRelayServer) createAllocation(clientAddr *net.UDPAddr, username, realm string) (*Allocation, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	// Проверяем лимит allocations
	if len(ts.allocations) >= ts.config.MaxAllocations {
		return nil, fmt.Errorf("maximum allocations reached")
	}

	// Генерируем уникальный ID
	allocationID := ts.generateAllocationID()

	// Создаем relay адрес
	relayAddr, err := ts.allocateRelayAddress()
	if err != nil {
		return nil, fmt.Errorf("failed to allocate relay address: %w", err)
	}

	// Создаем UDP соединение для relay
	relayConn, err := net.ListenUDP("udp", relayAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create relay connection: %w", err)
	}

	// Создаем allocation
	allocation := &Allocation{
		ID:           allocationID,
		ClientAddr:   clientAddr,
		RelayAddr:    relayAddr,
		RelayConn:    relayConn,
		Permissions:  make(map[string]*Permission),
		Channels:     make(map[uint16]*Channel),
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
		Username:     username,
		Realm:        realm,
		Nonce:        ts.generateNonce(),
		Lifetime:     ts.config.AllocTimeout,
	}

	// Сохраняем allocation
	ts.allocations[allocationID] = allocation

	ts.logger.Info("Created TURN allocation",
		zap.String("allocation_id", allocationID),
		zap.String("client_addr", clientAddr.String()),
		zap.String("relay_addr", relayAddr.String()))

	// Запускаем обработчик relay соединения
	go ts.handleRelayConnection(allocation)

	return allocation, nil
}

// allocateRelayAddress выделяет relay адрес
func (ts *TURNRelayServer) allocateRelayAddress() (*net.UDPAddr, error) {
	// Находим свободный порт в диапазоне
	for port := ts.config.MinPort; port <= ts.config.MaxPort; port++ {
		addr := &net.UDPAddr{
			IP:   net.IPv4zero, // Будет заменен на реальный IP
			Port: port,
		}

		// Проверяем, свободен ли порт
		conn, err := net.ListenUDP("udp", addr)
		if err != nil {
			continue
		}
		conn.Close()

		return addr, nil
	}

	return nil, fmt.Errorf("no free ports available")
}

// handleRelayConnection обрабатывает relay соединение
func (ts *TURNRelayServer) handleRelayConnection(allocation *Allocation) {
	buffer := make([]byte, 1500)

	for {
		select {
		case <-ts.ctx.Done():
			return
		default:
			allocation.RelayConn.SetReadDeadline(time.Now().Add(1 * time.Second))
			n, peerAddr, err := allocation.RelayConn.ReadFromUDP(buffer)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				ts.logger.Error("Failed to read from relay connection",
					zap.String("allocation_id", allocation.ID),
					zap.Error(err))
				continue
			}

			// Проверяем permission
			if !ts.hasPermission(allocation, peerAddr) {
				ts.logger.Warn("No permission for peer",
					zap.String("allocation_id", allocation.ID),
					zap.String("peer_addr", peerAddr.String()))
				continue
			}

			// Отправляем данные клиенту
			ts.sendDataToClient(allocation, buffer[:n], peerAddr)
		}
	}
}

// sendDataToClient отправляет данные клиенту
func (ts *TURNRelayServer) sendDataToClient(allocation *Allocation, data []byte, peerAddr *net.UDPAddr) {
	// Создаем Data Indication
	indication := &STUNMessage{
		Type:          STUNDataIndication,
		Length:        0,
		MagicCookie:   STUNMagicCookie,
		TransactionID: [12]byte{}, // Для Data Indication не нужен transaction ID
		Attributes: []STUNAttribute{
			{
				Type:   STUNAttrXORPeerAddress,
				Length: uint16(len(peerAddr.IP) + 2), // IP + Port
				Value:  ts.encodeAddress(peerAddr),
			},
			{
				Type:   STUNAttrData,
				Length: uint16(len(data)),
				Value:  data,
			},
		},
	}

	// Отправляем через UDP listener
	response := ts.encodeMessage(indication)
	ts.udpListener.WriteToUDP(response, allocation.ClientAddr)

	allocation.LastActivity = time.Now()
}

// hasPermission проверяет, есть ли permission для peer адреса
func (ts *TURNRelayServer) hasPermission(allocation *Allocation, peerAddr *net.UDPAddr) bool {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	peerKey := peerAddr.String()
	permission, exists := allocation.Permissions[peerKey]
	if !exists {
		return false
	}

	// Проверяем lifetime
	if time.Since(permission.CreatedAt) > permission.Lifetime {
		delete(allocation.Permissions, peerKey)
		return false
	}

	return true
}

// generateAllocationID генерирует уникальный ID allocation
func (ts *TURNRelayServer) generateAllocationID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// generateNonce генерирует nonce для аутентификации
func (ts *TURNRelayServer) generateNonce() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// authenticate проверяет аутентификацию
func (ts *TURNRelayServer) authenticate(message *STUNMessage, clientAddr *net.UDPAddr) bool {
	// Упрощенная аутентификация для демонстрации
	// В реальной реализации здесь должна быть проверка username/password
	return message.Username != "" && message.Realm == ts.config.Realm
}

// sendAllocateResponse отправляет успешный Allocate ответ
func (ts *TURNRelayServer) sendAllocateResponse(message *STUNMessage, clientAddr *net.UDPAddr, conn *net.UDPConn, allocation *Allocation) {
	response := &STUNMessage{
		Type:          STUNAllocateResponse,
		Length:        0,
		MagicCookie:   STUNMagicCookie,
		TransactionID: message.TransactionID,
		Attributes: []STUNAttribute{
			{
				Type:   STUNAttrXORMappedAddress,
				Length: uint16(len(allocation.RelayAddr.IP) + 2),
				Value:  ts.encodeAddress(allocation.RelayAddr),
			},
			{
				Type:   STUNAttrLifetime,
				Length: 4,
				Value:  make([]byte, 4),
			},
		},
	}

	// Устанавливаем lifetime
	binary.BigEndian.PutUint32(response.Attributes[1].Value, uint32(allocation.Lifetime.Seconds()))

	// Отправляем ответ
	responseData := ts.encodeMessage(response)
	conn.WriteToUDP(responseData, clientAddr)

	ts.logger.Info("Sent Allocate response",
		zap.String("client_addr", clientAddr.String()),
		zap.String("relay_addr", allocation.RelayAddr.String()))
}

// sendErrorResponse отправляет ошибку
func (ts *TURNRelayServer) sendErrorResponse(message *STUNMessage, clientAddr *net.UDPAddr, conn *net.UDPConn, errorCode uint16) {
	response := &STUNMessage{
		Type:          STUNBindingError,
		Length:        0,
		MagicCookie:   STUNMagicCookie,
		TransactionID: message.TransactionID,
		Attributes: []STUNAttribute{
			{
				Type:   STUNAttrErrorCode,
				Length: 4,
				Value:  make([]byte, 4),
			},
		},
	}

	// Устанавливаем error code
	binary.BigEndian.PutUint32(response.Attributes[0].Value, uint32(errorCode))

	// Отправляем ответ
	responseData := ts.encodeMessage(response)
	conn.WriteToUDP(responseData, clientAddr)
}

// encodeAddress кодирует адрес в STUN формат
func (ts *TURNRelayServer) encodeAddress(addr *net.UDPAddr) []byte {
	data := make([]byte, 8)
	data[1] = 1 // IPv4
	binary.BigEndian.PutUint16(data[2:4], uint16(addr.Port))
	copy(data[4:8], addr.IP.To4())
	return data
}

// cleanupRoutine очищает устаревшие allocations
func (ts *TURNRelayServer) cleanupRoutine() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ts.ctx.Done():
			return
		case <-ticker.C:
			ts.cleanupExpiredAllocations()
		}
	}
}

// cleanupExpiredAllocations очищает устаревшие allocations
func (ts *TURNRelayServer) cleanupExpiredAllocations() {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	now := time.Now()
	for id, allocation := range ts.allocations {
		if now.Sub(allocation.LastActivity) > allocation.Lifetime {
			ts.logger.Info("Cleaning up expired allocation",
				zap.String("allocation_id", id),
				zap.Duration("age", now.Sub(allocation.CreatedAt)))

			// Закрываем relay соединение
			if allocation.RelayConn != nil {
				allocation.RelayConn.Close()
			}

			// Удаляем allocation
			delete(ts.allocations, id)
		}
	}
}

// Close закрывает TURN сервер
func (ts *TURNRelayServer) Close() error {
	ts.logger.Info("Closing TURN server")

	ts.cancel()

	// Закрываем listeners
	if ts.udpListener != nil {
		ts.udpListener.Close()
	}
	if ts.tcpListener != nil {
		ts.tcpListener.Close()
	}

	// Закрываем все allocations
	ts.mu.Lock()
	for _, allocation := range ts.allocations {
		if allocation.RelayConn != nil {
			allocation.RelayConn.Close()
		}
	}
	ts.allocations = make(map[string]*Allocation)
	ts.mu.Unlock()

	ts.isRunning = false
	ts.logger.Info("TURN server closed")

	return nil
}

// GetStats возвращает статистику TURN сервера
func (ts *TURNRelayServer) GetStats() map[string]interface{} {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	return map[string]interface{}{
		"allocations_count": len(ts.allocations),
		"is_running":        ts.isRunning,
		"listen_addr":       ts.config.ListenAddr,
		"realm":             ts.config.Realm,
	}
}

// Заглушки для методов, которые будут реализованы позже
func (ts *TURNRelayServer) handleBindingRequest(message *STUNMessage, clientAddr *net.UDPAddr, conn *net.UDPConn) {
	// Реализация STUN Binding Request
}

func (ts *TURNRelayServer) handleRefreshRequest(message *STUNMessage, clientAddr *net.UDPAddr, conn *net.UDPConn) {
	// Реализация TURN Refresh Request
}

func (ts *TURNRelayServer) handleCreatePermissionRequest(message *STUNMessage, clientAddr *net.UDPAddr, conn *net.UDPConn) {
	// Реализация TURN Create Permission Request
}

func (ts *TURNRelayServer) handleChannelBindRequest(message *STUNMessage, clientAddr *net.UDPAddr, conn *net.UDPConn) {
	// Реализация TURN Channel Bind Request
}

func (ts *TURNRelayServer) handleSendIndication(message *STUNMessage, clientAddr *net.UDPAddr, conn *net.UDPConn) {
	// Реализация TURN Send Indication
}

func (ts *TURNRelayServer) handleDataIndication(message *STUNMessage, clientAddr *net.UDPAddr, conn *net.UDPConn) {
	// Реализация TURN Data Indication
}

func (ts *TURNRelayServer) handleTCPConnection(conn *net.TCPConn) {
	// Реализация TCP соединения
	defer conn.Close()
}

func (ts *TURNRelayServer) parseMessage(data []byte) (*STUNMessage, error) {
	// Используем существующий парсер из stun.go
	return nil, fmt.Errorf("not implemented")
}

func (ts *TURNRelayServer) encodeMessage(message *STUNMessage) []byte {
	// Используем существующий кодировщик из stun.go
	return nil
}
