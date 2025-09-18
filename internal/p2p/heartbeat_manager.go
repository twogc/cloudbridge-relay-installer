package p2p

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// HeartbeatManager управляет heartbeat системой
type HeartbeatManager struct {
	peers           map[string]*PeerHeartbeat
	relayManager    *RelayManager
	mutex           sync.RWMutex
	logger          *zap.Logger
	ctx             context.Context
	cancel          context.CancelFunc
	heartbeatTicker *time.Ticker
	cleanupTicker   *time.Ticker
}

// PeerHeartbeat представляет heartbeat информацию о пире
type PeerHeartbeat struct {
	PeerID          string
	TenantID        string
	LastHeartbeat   time.Time
	HeartbeatCount  int64
	IsOnline        bool
	LastSeen        time.Time
	HeartbeatTicker *time.Ticker
	Mutex           sync.RWMutex
}

// HeartbeatRequest представляет запрос heartbeat
type HeartbeatRequest struct {
	PeerID    string `json:"peer_id"`
	TenantID  string `json:"tenant_id"`
	Timestamp int64  `json:"timestamp"`
	Status    string `json:"status"`
}

// HeartbeatResponse представляет ответ heartbeat
type HeartbeatResponse struct {
	Status    string `json:"status"`
	Timestamp int64  `json:"timestamp"`
	Online    bool   `json:"online"`
	Message   string `json:"message,omitempty"`
}

// NewHeartbeatManager создает новый Heartbeat Manager
func NewHeartbeatManager(relayManager *RelayManager, logger *zap.Logger) *HeartbeatManager {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &HeartbeatManager{
		peers:        make(map[string]*PeerHeartbeat),
		relayManager: relayManager,
		logger:       logger,
		ctx:          ctx,
		cancel:       cancel,
	}
}

// Start запускает Heartbeat Manager
func (hm *HeartbeatManager) Start() error {
	hm.logger.Info("Starting Heartbeat Manager")
	
	// Запускаем heartbeat ticker (каждые 30 секунд)
	hm.heartbeatTicker = time.NewTicker(30 * time.Second)
	go hm.processHeartbeats()
	
	// Запускаем cleanup ticker (каждые 60 секунд)
	hm.cleanupTicker = time.NewTicker(60 * time.Second)
	go hm.cleanupOfflinePeers()
	
	return nil
}

// RegisterPeer регистрирует пира для heartbeat
func (hm *HeartbeatManager) RegisterPeer(peerID, tenantID string) error {
	hm.mutex.Lock()
	defer hm.mutex.Unlock()
	
	// Создаем heartbeat для пира
	peerHeartbeat := &PeerHeartbeat{
		PeerID:        peerID,
		TenantID:      tenantID,
		LastHeartbeat: time.Now(),
		HeartbeatCount: 0,
		IsOnline:      true,
		LastSeen:      time.Now(),
	}
	
	hm.peers[peerID] = peerHeartbeat
	
	hm.logger.Info("Peer registered for heartbeat",
		zap.String("peer_id", peerID),
		zap.String("tenant_id", tenantID))
	
	return nil
}

// ProcessHeartbeat обрабатывает heartbeat от пира
func (hm *HeartbeatManager) ProcessHeartbeat(peerID, tenantID string) (*HeartbeatResponse, error) {
	hm.mutex.Lock()
	defer hm.mutex.Unlock()
	
	peerHeartbeat, exists := hm.peers[peerID]
	if !exists {
		// Регистрируем нового пира
		peerHeartbeat = &PeerHeartbeat{
			PeerID:        peerID,
			TenantID:      tenantID,
			LastHeartbeat: time.Now(),
			HeartbeatCount: 1,
			IsOnline:      true,
			LastSeen:      time.Now(),
		}
		hm.peers[peerID] = peerHeartbeat
		
		hm.logger.Info("New peer registered via heartbeat",
			zap.String("peer_id", peerID),
			zap.String("tenant_id", tenantID))
	} else {
		// Обновляем существующего пира
		peerHeartbeat.Mutex.Lock()
		peerHeartbeat.LastHeartbeat = time.Now()
		peerHeartbeat.HeartbeatCount++
		peerHeartbeat.IsOnline = true
		peerHeartbeat.LastSeen = time.Now()
		peerHeartbeat.Mutex.Unlock()
		
		hm.logger.Debug("Heartbeat received",
			zap.String("peer_id", peerID),
			zap.String("tenant_id", tenantID),
			zap.Int64("count", peerHeartbeat.HeartbeatCount))
	}
	
	// Обновляем статус в Relay Manager
	hm.updatePeerStatus(peerID, tenantID, true)
	
	return &HeartbeatResponse{
		Status:    "ok",
		Timestamp: time.Now().Unix(),
		Online:    true,
		Message:   "Heartbeat processed successfully",
	}, nil
}

// processHeartbeats обрабатывает heartbeat события
func (hm *HeartbeatManager) processHeartbeats() {
	for {
		select {
		case <-hm.ctx.Done():
			return
		case <-hm.heartbeatTicker.C:
			hm.checkHeartbeatTimeouts()
		}
	}
}

// checkHeartbeatTimeouts проверяет timeout heartbeat
func (hm *HeartbeatManager) checkHeartbeatTimeouts() {
	hm.mutex.RLock()
	peers := make(map[string]*PeerHeartbeat)
	for peerID, peer := range hm.peers {
		peers[peerID] = peer
	}
	hm.mutex.RUnlock()
	
	timeout := 2 * time.Minute // 2 минуты без heartbeat
	
	for peerID, peer := range peers {
		peer.Mutex.RLock()
		lastHeartbeat := peer.LastHeartbeat
		isOnline := peer.IsOnline
		peer.Mutex.RUnlock()
		
		if time.Since(lastHeartbeat) > timeout && isOnline {
			// Помечаем пира как офлайн
			peer.Mutex.Lock()
			peer.IsOnline = false
			peer.Mutex.Unlock()
			
			hm.updatePeerStatus(peerID, peer.TenantID, false)
			
			hm.logger.Info("Peer marked as offline due to heartbeat timeout",
				zap.String("peer_id", peerID),
				zap.String("tenant_id", peer.TenantID),
				zap.Duration("timeout", timeout))
		}
	}
}

// cleanupOfflinePeers очищает офлайн пиров
func (hm *HeartbeatManager) cleanupOfflinePeers() {
	for {
		select {
		case <-hm.ctx.Done():
			return
		case <-hm.cleanupTicker.C:
			hm.performCleanup()
		}
	}
}

// performCleanup выполняет очистку офлайн пиров
func (hm *HeartbeatManager) performCleanup() {
	hm.mutex.Lock()
	defer hm.mutex.Unlock()
	
	cutoff := time.Now().Add(-10 * time.Minute) // 10 минут без активности
	var peersToRemove []string
	
	for peerID, peer := range hm.peers {
		peer.Mutex.RLock()
		lastSeen := peer.LastSeen
		isOnline := peer.IsOnline
		peer.Mutex.RUnlock()
		
		if !isOnline && lastSeen.Before(cutoff) {
			peersToRemove = append(peersToRemove, peerID)
		}
	}
	
	// Удаляем офлайн пиров
	for _, peerID := range peersToRemove {
		delete(hm.peers, peerID)
		
		hm.logger.Info("Peer removed due to inactivity",
			zap.String("peer_id", peerID))
	}
	
	if len(peersToRemove) > 0 {
		hm.logger.Info("Cleanup completed",
			zap.Int("removed_peers", len(peersToRemove)))
	}
}

// updatePeerStatus обновляет статус пира в Relay Manager
func (hm *HeartbeatManager) updatePeerStatus(peerID, tenantID string, isOnline bool) {
	// Здесь можно добавить логику для обновления статуса в Relay Manager
	// Например, закрытие соединений для офлайн пиров
	
	if !isOnline {
		hm.logger.Info("Peer went offline",
			zap.String("peer_id", peerID),
			zap.String("tenant_id", tenantID))
	} else {
		hm.logger.Debug("Peer is online",
			zap.String("peer_id", peerID),
			zap.String("tenant_id", tenantID))
	}
}

// GetPeerStatus возвращает статус пира
func (hm *HeartbeatManager) GetPeerStatus(peerID string) (*PeerHeartbeat, error) {
	hm.mutex.RLock()
	defer hm.mutex.RUnlock()
	
	peer, exists := hm.peers[peerID]
	if !exists {
		return nil, fmt.Errorf("peer not found: %s", peerID)
	}
	
	return peer, nil
}

// GetOnlinePeers возвращает список онлайн пиров
func (hm *HeartbeatManager) GetOnlinePeers(tenantID string) []string {
	hm.mutex.RLock()
	defer hm.mutex.RUnlock()
	
	var onlinePeers []string
	for peerID, peer := range hm.peers {
		peer.Mutex.RLock()
		isOnline := peer.IsOnline
		peerTenantID := peer.TenantID
		peer.Mutex.RUnlock()
		
		if isOnline && peerTenantID == tenantID {
			onlinePeers = append(onlinePeers, peerID)
		}
	}
	
	return onlinePeers
}

// GetTenantStats возвращает статистику по tenant
func (hm *HeartbeatManager) GetTenantStats(tenantID string) map[string]interface{} {
	hm.mutex.RLock()
	defer hm.mutex.RUnlock()
	
	stats := map[string]interface{}{
		"total_peers":   0,
		"online_peers":  0,
		"offline_peers": 0,
	}
	
	for _, peer := range hm.peers {
		peer.Mutex.RLock()
		isOnline := peer.IsOnline
		peerTenantID := peer.TenantID
		peer.Mutex.RUnlock()
		
		if peerTenantID == tenantID {
			stats["total_peers"] = stats["total_peers"].(int) + 1
			
			if isOnline {
				stats["online_peers"] = stats["online_peers"].(int) + 1
			} else {
				stats["offline_peers"] = stats["offline_peers"].(int) + 1
			}
		}
	}
	
	return stats
}

// Stop останавливает Heartbeat Manager
func (hm *HeartbeatManager) Stop() {
	hm.logger.Info("Stopping Heartbeat Manager")
	
	if hm.heartbeatTicker != nil {
		hm.heartbeatTicker.Stop()
	}
	
	if hm.cleanupTicker != nil {
		hm.cleanupTicker.Stop()
	}
	
	hm.cancel()
	
	hm.logger.Info("Heartbeat Manager stopped")
}
