package wireguard

import (
	"fmt"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

// PeerWhitelist manages peer whitelist validation
type PeerWhitelist struct {
	AllowedPeers map[string]bool
	mu           sync.RWMutex
	logger       *zap.Logger
	metrics      *PeerWhitelistMetrics
}

// PeerWhitelistMetrics represents metrics for peer whitelist
type PeerWhitelistMetrics struct {
	checksTotal     prometheus.Counter
	allowedTotal    prometheus.Counter
	deniedTotal     prometheus.Counter
	violationsTotal *prometheus.CounterVec
}

// NewPeerWhitelist creates a new peer whitelist
func NewPeerWhitelist(logger *zap.Logger) *PeerWhitelist {
	return &PeerWhitelist{
		AllowedPeers: make(map[string]bool),
		logger:       logger,
		metrics:      NewPeerWhitelistMetrics(),
	}
}

// NewPeerWhitelistMetrics creates new peer whitelist metrics
func NewPeerWhitelistMetrics() *PeerWhitelistMetrics {
	return &PeerWhitelistMetrics{
		checksTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "cloudbridge_peer_whitelist_checks_total",
			Help: "Total number of peer whitelist checks",
		}),
		allowedTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "cloudbridge_peer_whitelist_allowed_total",
			Help: "Total number of allowed peers",
		}),
		deniedTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "cloudbridge_peer_whitelist_denied_total",
			Help: "Total number of denied peers",
		}),
		violationsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "cloudbridge_peer_whitelist_violations_total",
			Help: "Total number of peer whitelist violations",
		}, []string{"violation_type"}),
	}
}

// RegisterMetrics registers peer whitelist metrics
func (pw *PeerWhitelist) RegisterMetrics(reg prometheus.Registerer) {
	reg.MustRegister(
		pw.metrics.checksTotal,
		pw.metrics.allowedTotal,
		pw.metrics.deniedTotal,
		pw.metrics.violationsTotal,
	)
}

// IsPeerAllowed checks if a peer is allowed
func (pw *PeerWhitelist) IsPeerAllowed(peerID string) bool {
	pw.mu.RLock()
	defer pw.mu.RUnlock()

	pw.metrics.checksTotal.Inc()

	allowed, exists := pw.AllowedPeers[peerID]
	if !exists || !allowed {
		pw.metrics.deniedTotal.Inc()
		pw.metrics.violationsTotal.WithLabelValues("peer_not_whitelisted").Inc()
		pw.logger.Warn("Peer not in whitelist",
			zap.String("peer_id", peerID),
		)
		return false
	}

	pw.metrics.allowedTotal.Inc()
	return true
}

// AddPeer adds a peer to the whitelist
func (pw *PeerWhitelist) AddPeer(peerID string) error {
	if peerID == "" {
		return fmt.Errorf("peer ID cannot be empty")
	}

	pw.mu.Lock()
	defer pw.mu.Unlock()

	pw.AllowedPeers[peerID] = true
	pw.logger.Info("Peer added to whitelist",
		zap.String("peer_id", peerID),
	)
	return nil
}

// RemovePeer removes a peer from the whitelist
func (pw *PeerWhitelist) RemovePeer(peerID string) error {
	pw.mu.Lock()
	defer pw.mu.Unlock()

	delete(pw.AllowedPeers, peerID)
	pw.logger.Info("Peer removed from whitelist",
		zap.String("peer_id", peerID),
	)
	return nil
}

// LoadFromJWT loads peer whitelist from JWT claims
func (pw *PeerWhitelist) LoadFromJWT(peerWhitelist []string) error {
	pw.mu.Lock()
	defer pw.mu.Unlock()

	pw.AllowedPeers = make(map[string]bool)
	for _, peerID := range peerWhitelist {
		if peerID != "" {
			pw.AllowedPeers[peerID] = true
		}
	}

	pw.logger.Info("Peer whitelist loaded from JWT",
		zap.Int("count", len(peerWhitelist)),
	)
	return nil
}

// GetAllowedPeers returns a copy of allowed peers
func (pw *PeerWhitelist) GetAllowedPeers() []string {
	pw.mu.RLock()
	defer pw.mu.RUnlock()

	peers := make([]string, 0, len(pw.AllowedPeers))
	for peerID := range pw.AllowedPeers {
		peers = append(peers, peerID)
	}
	return peers
}

// GetCount returns the number of allowed peers
func (pw *PeerWhitelist) GetCount() int {
	pw.mu.RLock()
	defer pw.mu.RUnlock()

	return len(pw.AllowedPeers)
}

// Clear clears the whitelist
func (pw *PeerWhitelist) Clear() {
	pw.mu.Lock()
	defer pw.mu.Unlock()

	pw.AllowedPeers = make(map[string]bool)
	pw.logger.Info("Peer whitelist cleared")
}


