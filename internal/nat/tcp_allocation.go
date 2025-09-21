package nat

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
)

// TCPAllocation представляет TCP allocation для TURN
type TCPAllocation struct {
	AllocationID     string
	Server           *TURNServerPool
	ClientID         string
	RelayAddress     string
	RelayPort        int
	ClientAddress    string
	ClientPort       int
	CreatedAt        time.Time
	LastActivity     time.Time
	BandwidthIn      uint64
	BandwidthOut     uint64
	IsActive         bool
	ConnectionCount  int
	MaxConnections   int
}

// TCPAllocationManager управляет TCP allocations
type TCPAllocationManager struct {
	logger *logger.Logger

	// Активные allocations
	allocations      map[string]*TCPAllocation
	allocationsMutex sync.RWMutex

	// TCP соединения
	connections      map[string]*TCPConnection
	connectionsMutex sync.RWMutex

	// Мониторинг
	cleanupInterval time.Duration
	maxIdleTime     time.Duration

	ctx    context.Context
	cancel context.CancelFunc
}

// TCPConnection представляет TCP соединение в allocation
type TCPConnection struct {
	ConnectionID   string
	AllocationID   string
	ClientConn     net.Conn
	RelayConn      net.Conn
	CreatedAt      time.Time
	LastActivity   time.Time
	BytesIn        uint64
	BytesOut       uint64
	IsActive       bool
}

// NewTCPAllocationManager создает новый TCP allocation manager
func NewTCPAllocationManager(logger *logger.Logger) *TCPAllocationManager {
	ctx, cancel := context.WithCancel(context.Background())
	tam := &TCPAllocationManager{
		logger:          logger,
		allocations:     make(map[string]*TCPAllocation),
		connections:     make(map[string]*TCPConnection),
		cleanupInterval: 1 * time.Minute,
		maxIdleTime:     5 * time.Minute,
		ctx:             ctx,
		cancel:          cancel,
	}
	go tam.startCleanup()
	return tam
}

// CreateTCPAllocation создает новое TCP allocation
func (tam *TCPAllocationManager) CreateTCPAllocation(clientID string, server *TURNServerPool, clientAddr string) (*TCPAllocation, error) {
	allocationID := fmt.Sprintf("tcp-alloc-%d", time.Now().UnixNano())
	
	// Генерируем relay адрес
	relayAddr, relayPort, err := tam.generateRelayAddress(server)
	if err != nil {
		return nil, fmt.Errorf("failed to generate relay address: %w", err)
	}

	allocation := &TCPAllocation{
		AllocationID:    allocationID,
		Server:          server,
		ClientID:        clientID,
		RelayAddress:    relayAddr,
		RelayPort:       relayPort,
		ClientAddress:   clientAddr,
		CreatedAt:       time.Now(),
		LastActivity:    time.Now(),
		IsActive:        true,
		MaxConnections:  10, // Максимум 10 TCP соединений на allocation
	}

	tam.allocationsMutex.Lock()
	tam.allocations[allocationID] = allocation
	tam.allocationsMutex.Unlock()

	tam.logger.Info("Created TCP allocation",
		zap.String("allocation_id", allocationID),
		zap.String("client_id", clientID),
		zap.String("server", server.Address),
		zap.String("relay_address", fmt.Sprintf("%s:%d", relayAddr, relayPort)),
		zap.String("client_address", clientAddr),
	)

	return allocation, nil
}

// CreateTCPConnection создает новое TCP соединение в allocation
func (tam *TCPAllocationManager) CreateTCPConnection(allocationID string, clientConn net.Conn, targetAddr string) (*TCPConnection, error) {
	tam.allocationsMutex.RLock()
	allocation, exists := tam.allocations[allocationID]
	tam.allocationsMutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("allocation not found: %s", allocationID)
	}

	if !allocation.IsActive {
		return nil, fmt.Errorf("allocation is not active: %s", allocationID)
	}

	if allocation.ConnectionCount >= allocation.MaxConnections {
		return nil, fmt.Errorf("allocation connection limit reached: %s", allocationID)
	}

	// Создаем соединение к целевой машине
	relayConn, err := net.DialTimeout("tcp", targetAddr, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to target: %w", err)
	}

	connectionID := fmt.Sprintf("tcp-conn-%d", time.Now().UnixNano())
	connection := &TCPConnection{
		ConnectionID: connectionID,
		AllocationID: allocationID,
		ClientConn:   clientConn,
		RelayConn:    relayConn,
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
		IsActive:     true,
	}

	tam.connectionsMutex.Lock()
	tam.connections[connectionID] = connection
	tam.connectionsMutex.Unlock()

	// Увеличиваем счетчик соединений в allocation
	tam.allocationsMutex.Lock()
	allocation.ConnectionCount++
	allocation.LastActivity = time.Now()
	tam.allocationsMutex.Unlock()

	tam.logger.Info("Created TCP connection",
		zap.String("connection_id", connectionID),
		zap.String("allocation_id", allocationID),
		zap.String("target_address", targetAddr),
	)

	// Запускаем проксирование данных
	go tam.proxyTCPConnection(connection)

	return connection, nil
}

// proxyTCPConnection проксирует данные между клиентом и целевой машиной
func (tam *TCPAllocationManager) proxyTCPConnection(connection *TCPConnection) {
	defer tam.closeTCPConnection(connection)

	// Каналы для передачи данных
	clientToRelay := make(chan []byte, 100)
	relayToClient := make(chan []byte, 100)

	// Горутины для чтения данных
	go tam.readFromConnection(connection.ClientConn, clientToRelay, "client")
	go tam.readFromConnection(connection.RelayConn, relayToClient, "relay")

	// Основной цикл проксирования
	for {
		select {
		case data, ok := <-clientToRelay:
			if !ok {
				tam.logger.Debug("Client connection closed", zap.String("connection_id", connection.ConnectionID))
				return
			}
			if _, err := connection.RelayConn.Write(data); err != nil {
				tam.logger.Error("Failed to write to relay", zap.Error(err))
				return
			}
			connection.BytesOut += uint64(len(data))
			connection.LastActivity = time.Now()

		case data, ok := <-relayToClient:
			if !ok {
				tam.logger.Debug("Relay connection closed", zap.String("connection_id", connection.ConnectionID))
				return
			}
			if _, err := connection.ClientConn.Write(data); err != nil {
				tam.logger.Error("Failed to write to client", zap.Error(err))
				return
			}
			connection.BytesIn += uint64(len(data))
			connection.LastActivity = time.Now()

		case <-tam.ctx.Done():
			return
		}
	}
}

// readFromConnection читает данные из соединения
func (tam *TCPAllocationManager) readFromConnection(conn net.Conn, dataChan chan<- []byte, name string) {
	defer close(dataChan)
	
	buffer := make([]byte, 4096)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			tam.logger.Debug("Connection read error", zap.String("name", name), zap.Error(err))
			return
		}
		
		if n > 0 {
			data := make([]byte, n)
			copy(data, buffer[:n])
			select {
			case dataChan <- data:
			case <-tam.ctx.Done():
				return
			}
		}
	}
}

// closeTCPConnection закрывает TCP соединение
func (tam *TCPAllocationManager) closeTCPConnection(connection *TCPConnection) {
	connection.IsActive = false

	// Закрываем соединения
	if connection.ClientConn != nil {
		connection.ClientConn.Close()
	}
	if connection.RelayConn != nil {
		connection.RelayConn.Close()
	}

	// Уменьшаем счетчик соединений в allocation
	tam.allocationsMutex.Lock()
	if allocation, exists := tam.allocations[connection.AllocationID]; exists {
		if allocation.ConnectionCount > 0 {
			allocation.ConnectionCount--
		}
		allocation.LastActivity = time.Now()
	}
	tam.allocationsMutex.Unlock()

	// Удаляем соединение
	tam.connectionsMutex.Lock()
	delete(tam.connections, connection.ConnectionID)
	tam.connectionsMutex.Unlock()

	tam.logger.Info("Closed TCP connection",
		zap.String("connection_id", connection.ConnectionID),
		zap.String("allocation_id", connection.AllocationID),
		zap.Uint64("bytes_in", connection.BytesIn),
		zap.Uint64("bytes_out", connection.BytesOut),
	)
}

// ReleaseTCPAllocation освобождает TCP allocation
func (tam *TCPAllocationManager) ReleaseTCPAllocation(allocationID string) error {
	tam.allocationsMutex.Lock()
	allocation, exists := tam.allocations[allocationID]
	if exists {
		allocation.IsActive = false
	}
	tam.allocationsMutex.Unlock()

	if !exists {
		return fmt.Errorf("allocation not found: %s", allocationID)
	}

	// Закрываем все соединения в allocation
	tam.connectionsMutex.Lock()
	connectionsToClose := make([]*TCPConnection, 0)
	for _, connection := range tam.connections {
		if connection.AllocationID == allocationID {
			connectionsToClose = append(connectionsToClose, connection)
		}
	}
	tam.connectionsMutex.Unlock()

	for _, connection := range connectionsToClose {
		tam.closeTCPConnection(connection)
	}

	// Удаляем allocation
	tam.allocationsMutex.Lock()
	delete(tam.allocations, allocationID)
	tam.allocationsMutex.Unlock()

	tam.logger.Info("Released TCP allocation",
		zap.String("allocation_id", allocationID),
		zap.String("client_id", allocation.ClientID),
	)

	return nil
}

// generateRelayAddress генерирует relay адрес для TCP allocation
func (tam *TCPAllocationManager) generateRelayAddress(server *TURNServerPool) (string, int, error) {
	// В реальной реализации здесь должна быть логика генерации relay адреса
	// Пока возвращаем фиктивный адрес
	relayAddr := "127.0.0.1"
	relayPort := 49152 + (time.Now().UnixNano() % 16384) // Порт в диапазоне 49152-65535
	return relayAddr, int(relayPort), nil
}

// startCleanup запускает периодическую очистку
func (tam *TCPAllocationManager) startCleanup() {
	ticker := time.NewTicker(tam.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-tam.ctx.Done():
			return
		case <-ticker.C:
			tam.cleanupExpiredAllocations()
			tam.cleanupExpiredConnections()
		}
	}
}

// cleanupExpiredAllocations очищает просроченные allocations
func (tam *TCPAllocationManager) cleanupExpiredAllocations() {
	tam.allocationsMutex.Lock()
	defer tam.allocationsMutex.Unlock()

	expiredTime := time.Now().Add(-tam.maxIdleTime)
	for allocationID, allocation := range tam.allocations {
		if allocation.LastActivity.Before(expiredTime) {
			allocation.IsActive = false
			delete(tam.allocations, allocationID)
			tam.logger.Debug("Cleaned up expired TCP allocation", zap.String("allocation_id", allocationID))
		}
	}
}

// cleanupExpiredConnections очищает просроченные соединения
func (tam *TCPAllocationManager) cleanupExpiredConnections() {
	tam.connectionsMutex.Lock()
	defer tam.connectionsMutex.Unlock()

	expiredTime := time.Now().Add(-tam.maxIdleTime)
	for connectionID, connection := range tam.connections {
		if connection.LastActivity.Before(expiredTime) {
			connection.IsActive = false
			delete(tam.connections, connectionID)
			tam.logger.Debug("Cleaned up expired TCP connection", zap.String("connection_id", connectionID))
		}
	}
}

// GetTCPAllocationStats возвращает статистику TCP allocations
func (tam *TCPAllocationManager) GetTCPAllocationStats() map[string]interface{} {
	tam.allocationsMutex.RLock()
	defer tam.allocationsMutex.RUnlock()

	tam.connectionsMutex.RLock()
	defer tam.connectionsMutex.RUnlock()

	activeAllocations := 0
	totalConnections := 0
	totalBandwidthIn := uint64(0)
	totalBandwidthOut := uint64(0)

	for _, allocation := range tam.allocations {
		if allocation.IsActive {
			activeAllocations++
		}
		totalBandwidthIn += allocation.BandwidthIn
		totalBandwidthOut += allocation.BandwidthOut
	}

	for _, connection := range tam.connections {
		if connection.IsActive {
			totalConnections++
		}
	}

	return map[string]interface{}{
		"total_allocations":    len(tam.allocations),
		"active_allocations":   activeAllocations,
		"total_connections":    len(tam.connections),
		"active_connections":   totalConnections,
		"total_bandwidth_in":   totalBandwidthIn,
		"total_bandwidth_out":  totalBandwidthOut,
	}
}

// Stop останавливает TCP allocation manager
func (tam *TCPAllocationManager) Stop() {
	tam.cancel()
	tam.logger.Info("TCP allocation manager stopped")
}
