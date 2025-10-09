package p2p

import (
	"fmt"
	"sync"

	"go.uber.org/zap"
)

// SimpleMeshTopology простая реализация P2PMeshTopologyInterface
type SimpleMeshTopology struct {
	peers  map[string]*Peer
	mutex  sync.RWMutex
	logger *zap.Logger
}

// NewSimpleMeshTopology создает новую простую mesh топологию
func NewSimpleMeshTopology(logger *zap.Logger) *SimpleMeshTopology {
	return &SimpleMeshTopology{
		peers:  make(map[string]*Peer),
		logger: logger,
	}
}

// AddPeer добавляет пира в mesh
func (smt *SimpleMeshTopology) AddPeer(peerID string, peer *Peer) error {
	smt.mutex.Lock()
	defer smt.mutex.Unlock()

	smt.peers[peerID] = peer

	smt.logger.Info("Peer added to mesh topology",
		zap.String("peer_id", peerID),
		zap.String("tenant_id", peer.TenantID),
		zap.String("endpoint", peer.Endpoint))

	return nil
}

// GetPeer получает пира по ID
func (smt *SimpleMeshTopology) GetPeer(peerID string) (*Peer, bool) {
	smt.mutex.RLock()
	defer smt.mutex.RUnlock()

	peer, exists := smt.peers[peerID]
	return peer, exists
}

// GetOptimalRoute получает оптимальный маршрут
func (smt *SimpleMeshTopology) GetOptimalRoute(clientIP string) ([]string, error) {
	smt.mutex.RLock()
	defer smt.mutex.RUnlock()

	var routes []string
	for peerID, peer := range smt.peers {
		if peer.IsOnline && peer.Endpoint != "" {
			routes = append(routes, peerID)
		}
	}

	if len(routes) == 0 {
		return nil, fmt.Errorf("no online peers available")
	}

	smt.logger.Debug("Optimal route calculated",
		zap.String("client_ip", clientIP),
		zap.Strings("routes", routes))

	return routes, nil
}

// GetAllPeers возвращает всех пиров
func (smt *SimpleMeshTopology) GetAllPeers() map[string]*Peer {
	smt.mutex.RLock()
	defer smt.mutex.RUnlock()

	// Создаем копию для безопасности
	peers := make(map[string]*Peer)
	for k, v := range smt.peers {
		peers[k] = v
	}

	return peers
}

// RemovePeer удаляет пира из mesh
func (smt *SimpleMeshTopology) RemovePeer(peerID string) {
	smt.mutex.Lock()
	defer smt.mutex.Unlock()

	delete(smt.peers, peerID)

	smt.logger.Info("Peer removed from mesh topology",
		zap.String("peer_id", peerID))
}
