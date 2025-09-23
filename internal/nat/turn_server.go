package nat

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"go.uber.org/zap"
)

// TURNServer представляет простой TURN сервер
type TURNServer struct {
	logger      *zap.Logger
	listenAddr  string
	conn        *net.UDPConn
	allocations map[string]*TURNServerAllocation
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
}

// TURNServerAllocation представляет TURN allocation для сервера
type TURNServerAllocation struct {
	ID          string
	ClientAddr  *net.UDPAddr
	RelayAddr   *net.UDPAddr
	RelayConn   *net.UDPConn
	CreatedAt   time.Time
	LastUsed    time.Time
	Permissions map[string]*TURNPermission
	mu          sync.RWMutex
}

// TURNPermission представляет TURN permission
type TURNPermission struct {
	PeerAddr  *net.UDPAddr
	CreatedAt time.Time
	ExpiresAt time.Time
}

// NewTURNServer создает новый TURN сервер
func NewTURNServer(listenAddr string, logger *zap.Logger) (*TURNServer, error) {
	ctx, cancel := context.WithCancel(context.Background())

	return &TURNServer{
		logger:      logger,
		listenAddr:  listenAddr,
		allocations: make(map[string]*TURNServerAllocation),
		ctx:         ctx,
		cancel:      cancel,
	}, nil
}

// Start запускает TURN сервер
func (ts *TURNServer) Start() error {
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
	ts.logger.Info("TURN server started", zap.String("addr", ts.listenAddr))

	// Start packet handling
	go ts.handlePackets()

	// Start cleanup routine
	go ts.cleanupAllocations()

	return nil
}

// Stop останавливает TURN сервер
func (ts *TURNServer) Stop() error {
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
	ts.allocations = make(map[string]*TURNServerAllocation)
	ts.mu.Unlock()

	ts.logger.Info("TURN server stopped")
	return nil
}

// handlePackets обрабатывает входящие пакеты
func (ts *TURNServer) handlePackets() {
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
func (ts *TURNServer) handlePacket(data []byte, clientAddr *net.UDPAddr) {
	ts.logger.Debug("Received packet",
		zap.String("from", clientAddr.String()),
		zap.Int("size", len(data)))

	// Simplified TURN protocol handling
	// In a real implementation, this would parse STUN/TURN messages

	// For now, we'll implement a simple echo-based allocation system
	if len(data) < 4 {
		return
	}

	// Check if this is an allocation request (simplified)
	if data[0] == 0x00 && data[1] == 0x03 { // Simplified TURN Allocate Request
		ts.handleAllocateRequest(clientAddr)
	} else {
		// Handle data relay
		ts.handleDataRelay(data, clientAddr)
	}
}

// handleAllocateRequest обрабатывает запрос на allocation
func (ts *TURNServer) handleAllocateRequest(clientAddr *net.UDPAddr) {
	ts.logger.Info("Handling allocation request", zap.String("client", clientAddr.String()))

	// Create relay connection
	relayConn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(0, 0, 0, 0), Port: 0})
	if err != nil {
		ts.logger.Error("Failed to create relay connection", zap.Error(err))
		return
	}

	// Create allocation
	allocID := fmt.Sprintf("alloc_%d", time.Now().UnixNano())
	allocation := &TURNServerAllocation{
		ID:          allocID,
		ClientAddr:  clientAddr,
		RelayAddr:   relayConn.LocalAddr().(*net.UDPAddr),
		RelayConn:   relayConn,
		CreatedAt:   time.Now(),
		LastUsed:    time.Now(),
		Permissions: make(map[string]*TURNPermission),
	}

	// Store allocation
	ts.mu.Lock()
	ts.allocations[allocID] = allocation
	ts.mu.Unlock()

	// Start relay handling for this allocation
	go ts.handleRelayTraffic(allocation)

	// Send allocation response (simplified)
	response := []byte{0x01, 0x03, 0x00, 0x00} // Simplified success response
	_, err = ts.conn.WriteToUDP(response, clientAddr)
	if err != nil {
		ts.logger.Error("Failed to send allocation response", zap.Error(err))
	}

	ts.logger.Info("Allocation created",
		zap.String("alloc_id", allocID),
		zap.String("client", clientAddr.String()),
		zap.String("relay", allocation.RelayAddr.String()))
}

// handleDataRelay обрабатывает relay данных
func (ts *TURNServer) handleDataRelay(data []byte, clientAddr *net.UDPAddr) {
	// Find allocation for this client
	ts.mu.RLock()
	var allocation *TURNServerAllocation
	for _, alloc := range ts.allocations {
		if alloc.ClientAddr.String() == clientAddr.String() {
			allocation = alloc
			break
		}
	}
	ts.mu.RUnlock()

	if allocation == nil {
		ts.logger.Debug("No allocation found for client", zap.String("client", clientAddr.String()))
		return
	}

	// Update last used time
	allocation.mu.Lock()
	allocation.LastUsed = time.Now()
	allocation.mu.Unlock()

	// Relay data (simplified - in real implementation would handle TURN data messages)
	ts.logger.Debug("Relaying data",
		zap.String("alloc_id", allocation.ID),
		zap.Int("size", len(data)))
}

// handleRelayTraffic обрабатывает трафик для конкретного allocation
func (ts *TURNServer) handleRelayTraffic(allocation *TURNServerAllocation) {
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

			// Check permissions (simplified)
			if !ts.checkPermission(allocation, peerAddr) {
				ts.logger.Debug("Permission denied",
					zap.String("peer", peerAddr.String()),
					zap.String("alloc_id", allocation.ID))
				continue
			}

			// Forward to client (simplified - should wrap in TURN data message)
			_, err = ts.conn.WriteToUDP(buffer[:n], allocation.ClientAddr)
			if err != nil {
				ts.logger.Error("Failed to forward to client", zap.Error(err))
				continue
			}

			ts.logger.Debug("Forwarded packet to client",
				zap.String("from", peerAddr.String()),
				zap.String("to", allocation.ClientAddr.String()),
				zap.Int("size", n))
		}
	}
}

// checkPermission проверяет разрешение для peer
func (ts *TURNServer) checkPermission(allocation *TURNServerAllocation, peerAddr *net.UDPAddr) bool {
	allocation.mu.RLock()
	defer allocation.mu.RUnlock()

	// For simplicity, allow all peers (in real implementation would check permissions)
	return true
}

// cleanupAllocations очищает устаревшие allocations
func (ts *TURNServer) cleanupAllocations() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ts.ctx.Done():
			return
		case <-ticker.C:
			ts.mu.Lock()
			now := time.Now()
			for allocID, allocation := range ts.allocations {
				// Remove allocations older than 10 minutes
				if now.Sub(allocation.LastUsed) > 10*time.Minute {
					if allocation.RelayConn != nil {
						allocation.RelayConn.Close()
					}
					delete(ts.allocations, allocID)
					ts.logger.Info("Cleaned up allocation", zap.String("alloc_id", allocID))
				}
			}
			ts.mu.Unlock()
		}
	}
}

// GetAllocations возвращает текущие allocations (для мониторинга)
func (ts *TURNServer) GetAllocations() map[string]*TURNServerAllocation {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	allocations := make(map[string]*TURNServerAllocation)
	for k, v := range ts.allocations {
		allocations[k] = v
	}
	return allocations
}
