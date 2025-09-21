package p2p

import (
	"context"

	"github.com/twogc/cloudbridge-relay-installer/internal/nat"
	"github.com/twogc/cloudbridge-relay-installer/internal/wireguard"
)

// WireGuardMeshTopologyInterface defines the interface for WireGuard mesh topology operations
type WireGuardMeshTopologyInterface interface {
	AddPeer(peerID string, peer *wireguard.WireGuardPeer) error
	GetPeer(peerID string) (*wireguard.WireGuardPeer, bool)
	GetOptimalRoute(clientIP string) ([]string, error)
}

// PeerDiscoveryInterface defines the interface for peer discovery operations
type PeerDiscoveryInterface interface {
	Start() error
	Stop() error
	DiscoverPeers() ([]*wireguard.DiscoveredPeer, error)
}

// HybridRouterInterface defines the interface for hybrid routing operations
type HybridRouterInterface interface {
	GetRoutingStats() map[string]interface{}
}

// SiteToSiteManagerInterface defines the interface for site-to-site operations
type SiteToSiteManagerInterface interface {
	// Add methods as needed for site-to-site operations
}

// CadenceP2PIntegrationInterface defines the interface for Cadence P2P integration
type CadenceP2PIntegrationInterface interface {
	OrchestrateTunnelCreation(ctx context.Context, tenantID, peerID string, config map[string]interface{}) error
	OrchestrateConnectionEstablishment(ctx context.Context, tenantID, sourcePeerID, targetPeerID string, config map[string]interface{}) error
	OrchestrateAIOptimizedRouting(ctx context.Context, tenantID, sourcePeerID, targetPeerID string, dataSize int64, priority string, networkConditions map[string]interface{}) error
}

// ICEAgentInterface defines the interface for ICE agent operations
type ICEAgentInterface interface {
	StartGathering() error
	GetLocalCandidates() []*nat.Candidate
	AddRemoteCandidate(candidate *nat.Candidate) error
	StartConnectivityChecks() error
	GetState() nat.ICEState
	GetSelectedPair() *nat.CandidatePair
	Close() error
}
