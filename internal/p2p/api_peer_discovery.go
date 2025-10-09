package p2p

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// APIPeerDiscovery реализует PeerDiscoveryInterface через API
type APIPeerDiscovery struct {
	baseURL    string
	httpClient *http.Client
	logger     *zap.Logger
	tenantID   string
	token      string
}

// NewAPIPeerDiscovery создает новый API peer discovery
func NewAPIPeerDiscovery(baseURL, tenantID, token string, logger *zap.Logger) *APIPeerDiscovery {
	return &APIPeerDiscovery{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger:   logger,
		tenantID: tenantID,
		token:    token,
	}
}

// Start запускает peer discovery
func (apd *APIPeerDiscovery) Start() error {
	apd.logger.Info("API Peer Discovery started",
		zap.String("base_url", apd.baseURL),
		zap.String("tenant_id", apd.tenantID))
	return nil
}

// Stop останавливает peer discovery
func (apd *APIPeerDiscovery) Stop() error {
	apd.logger.Info("API Peer Discovery stopped")
	return nil
}

// DiscoverPeers обнаруживает пиров через API
func (apd *APIPeerDiscovery) DiscoverPeers() ([]*DiscoveredPeer, error) {
	url := fmt.Sprintf("%s/api/v1/tenants/%s/peers/discover", apd.baseURL, apd.tenantID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apd.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := apd.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var discoveryResponse struct {
		Success  bool    `json:"success"`
		TenantID string  `json:"tenant_id"`
		Peers    []*Peer `json:"peers"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&discoveryResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !discoveryResponse.Success {
		return nil, fmt.Errorf("API returned success=false")
	}

	// Конвертируем Peer в DiscoveredPeer
	var discoveredPeers []*DiscoveredPeer
	for _, peer := range discoveryResponse.Peers {
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

	apd.logger.Debug("Discovered peers via API",
		zap.String("tenant_id", apd.tenantID),
		zap.Int("peers_count", len(discoveredPeers)))

	return discoveredPeers, nil
}
