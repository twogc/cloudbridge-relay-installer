package p2p

import (
	"time"
	"go.uber.org/zap"
)

// PeerStoreInterface интерфейс для работы с peer store
type PeerStoreInterface interface {
	Store(tenantID, peerID string, peer *Peer)
	Get(tenantID, peerID string) (*Peer, bool)
	Update(tenantID, peerID string, peer *Peer)
	GetAll(tenantID string) []*Peer
	IsPeerOnline(peerID string) bool
	GetLastSeen(peerID string) time.Time
}

// DirectPeerDiscovery реализует PeerDiscoveryInterface через прямое обращение к peerStore
type DirectPeerDiscovery struct {
	peerStore PeerStoreInterface
	logger    *zap.Logger
	tenantID  string
}

// NewDirectPeerDiscovery создает новый прямой peer discovery
func NewDirectPeerDiscovery(peerStore PeerStoreInterface, tenantID string, logger *zap.Logger) *DirectPeerDiscovery {
	return &DirectPeerDiscovery{
		peerStore: peerStore,
		logger:    logger,
		tenantID:  tenantID,
	}
}

// Start запускает peer discovery
func (dpd *DirectPeerDiscovery) Start() error {
	dpd.logger.Info("Direct Peer Discovery started",
		zap.String("tenant_id", dpd.tenantID))
	return nil
}

// Stop останавливает peer discovery
func (dpd *DirectPeerDiscovery) Stop() error {
	dpd.logger.Info("Direct Peer Discovery stopped")
	return nil
}

// DiscoverPeers обнаруживает пиров напрямую из peerStore
func (dpd *DirectPeerDiscovery) DiscoverPeers() ([]*DiscoveredPeer, error) {
	// Получаем пиров напрямую из peerStore
	peers := dpd.peerStore.GetAll(dpd.tenantID)

	dpd.logger.Debug("Direct peer discovery",
		zap.String("tenant_id", dpd.tenantID),
		zap.Int("peers_count", len(peers)))

	// Конвертируем Peer в DiscoveredPeer
	var discoveredPeers []*DiscoveredPeer
	for _, peer := range peers {
		discoveredPeer := &DiscoveredPeer{
			ID:           peer.ID,
			PeerID:       peer.PeerID,
			TenantID:     peer.TenantID,
			Endpoint:     peer.Endpoint,
			Protocol:     peer.ConnectionType,
			LastSeen:     peer.LastSeen,
			Latency:      peer.LatencyMS,
			IsOnline:     peer.IsOnline,
			Capabilities: []string{"quic", "ice", "stun"},
			Region:       "default",
			TrustLevel:   "basic",
		}
		discoveredPeers = append(discoveredPeers, discoveredPeer)
	}

	dpd.logger.Debug("Discovered peers via direct access",
		zap.String("tenant_id", dpd.tenantID),
		zap.Int("peers_count", len(discoveredPeers)))

	return discoveredPeers, nil
}
