package p2p

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// SimplePeerManager упрощенный менеджер пиров
type SimplePeerManager struct {
	peers  map[string]*Peer
	mutex  sync.RWMutex
	logger *zap.Logger
	ctx    context.Context
	cancel context.CancelFunc
}

// NewPeerManager создает новый Peer Manager
func NewPeerManager(logger *zap.Logger) *SimplePeerManager {
	ctx, cancel := context.WithCancel(context.Background())

	pm := &SimplePeerManager{
		peers:  make(map[string]*Peer),
		logger: logger,
		ctx:    ctx,
		cancel: cancel,
	}

	// Запускаем периодическую очистку
	go pm.startCleanupRoutine()

	return pm
}

// RegisterPeer регистрирует нового пира
func (pm *SimplePeerManager) RegisterPeer(peer *Peer) error {
	if peer.ID == "" {
		return fmt.Errorf("peer ID cannot be empty")
	}

	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	now := time.Now()
	if existingPeer, exists := pm.peers[peer.ID]; exists {
		// Обновляем существующего пира
		existingPeer.LastSeen = now
		existingPeer.IsOnline = true

		pm.logger.Info("Peer updated",
			zap.String("peer_id", peer.ID),
			zap.String("tenant_id", peer.TenantID))
	} else {
		// Регистрируем нового пира
		peer.LastSeen = now
		peer.RegisteredAt = now
		peer.IsOnline = true
		pm.peers[peer.ID] = peer

		pm.logger.Info("New peer registered",
			zap.String("peer_id", peer.ID),
			zap.String("tenant_id", peer.TenantID))
	}

	return nil
}

// UnregisterPeer удаляет пира
func (pm *SimplePeerManager) UnregisterPeer(peerID string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	peer, exists := pm.peers[peerID]
	if !exists {
		return fmt.Errorf("peer not found: %s", peerID)
	}

	peer.IsOnline = false
	peer.Status = "disconnected"

	delete(pm.peers, peerID)

	pm.logger.Info("Peer unregistered",
		zap.String("peer_id", peerID),
		zap.String("tenant_id", peer.TenantID))

	return nil
}

// GetPeer возвращает пира по ID
func (pm *SimplePeerManager) GetPeer(peerID string) (*Peer, bool) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	peer, exists := pm.peers[peerID]
	return peer, exists
}

// GetPeersByTenant возвращает всех пиров тенанта
func (pm *SimplePeerManager) GetPeersByTenant(tenantID string) []*Peer {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	var peers []*Peer
	for _, peer := range pm.peers {
		if peer.TenantID == tenantID {
			peers = append(peers, peer)
		}
	}

	return peers
}

// GetOnlinePeers возвращает всех онлайн пиров
func (pm *SimplePeerManager) GetOnlinePeers() []*Peer {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	var peers []*Peer
	for _, peer := range pm.peers {
		if peer.IsOnline {
			peers = append(peers, peer)
		}
	}

	return peers
}

// UpdatePeerLastSeen обновляет время последней активности пира
func (pm *SimplePeerManager) UpdatePeerLastSeen(peerID string) error {
	pm.mutex.RLock()
	peer, exists := pm.peers[peerID]
	pm.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("peer not found: %s", peerID)
	}

	peer.LastSeen = time.Now()
	if !peer.IsOnline {
		peer.IsOnline = true
		peer.Status = "online"
	}

	return nil
}

// GetPeerCount возвращает количество пиров
func (pm *SimplePeerManager) GetPeerCount() int {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	return len(pm.peers)
}

// GetOnlinePeerCount возвращает количество онлайн пиров
func (pm *SimplePeerManager) GetOnlinePeerCount() int {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	count := 0
	for _, peer := range pm.peers {
		if peer.IsOnline {
			count++
		}
	}

	return count
}

// startCleanupRoutine запускает периодическую очистку неактивных пиров
func (pm *SimplePeerManager) startCleanupRoutine() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-pm.ctx.Done():
			return
		case <-ticker.C:
			pm.cleanupInactivePeers()
		}
	}
}

// cleanupInactivePeers удаляет неактивных пиров
func (pm *SimplePeerManager) cleanupInactivePeers() {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	now := time.Now()
	inactiveThreshold := 5 * time.Minute

	var toRemove []string
	for peerID, peer := range pm.peers {
		if now.Sub(peer.LastSeen) > inactiveThreshold && peer.IsOnline {
			toRemove = append(toRemove, peerID)
		}
	}

	for _, peerID := range toRemove {
		peer := pm.peers[peerID]
		peer.IsOnline = false
		peer.Status = "offline"

		pm.logger.Info("Peer marked as offline due to inactivity",
			zap.String("peer_id", peerID))
	}
}

// Shutdown останавливает Peer Manager
func (pm *SimplePeerManager) Shutdown() {
	pm.cancel()

	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	// Уведомляем о disconnection всех пиров
	for _, peer := range pm.peers {
		peer.IsOnline = false
		peer.Status = "disconnected"
	}

	pm.peers = make(map[string]*Peer)
	pm.logger.Info("Peer Manager shutdown completed")
}
