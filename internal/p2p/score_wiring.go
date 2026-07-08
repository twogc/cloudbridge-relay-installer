package p2p

import (
	"fmt"
	"os"
	"strings"

	"github.com/twogc/cloudbridge-relay-installer/internal/database"
	"github.com/twogc/cloudbridge-relay-installer/internal/network"
	"go.uber.org/zap"
)

// SCOREWiringConfig holds optional SCORE (tenant network control-plane) settings.
// Env: SCORE_BASE_URL, SCORE_API_KEY. Empty base URL disables wiring.
type SCOREWiringConfig struct {
	BaseURL string
	APIKey  string
}

// SCOREWiringFromEnv reads SCORE_BASE_URL and SCORE_API_KEY.
func SCOREWiringFromEnv() SCOREWiringConfig {
	return SCOREWiringConfig{
		BaseURL: strings.TrimSpace(os.Getenv("SCORE_BASE_URL")),
		APIKey:  strings.TrimSpace(os.Getenv("SCORE_API_KEY")),
	}
}

// Enabled reports whether SCORE hybrid wiring should run.
func (c SCOREWiringConfig) Enabled() bool {
	return c.BaseURL != ""
}

// NewRelayManagerHybridFromSCORE constructs RelayManagerHybrid with production
// SCORE network store + SCORETenantNetworkAllocatorProduction.
// repository may be nil (peer IPAM / WG persistence limited).
func NewRelayManagerHybridFromSCORE(
	logger *zap.Logger,
	cfg SCOREWiringConfig,
	repository *database.Repository,
) (*RelayManagerHybrid, error) {
	if logger == nil {
		logger = zap.NewNop()
	}
	if !cfg.Enabled() {
		return nil, fmt.Errorf("SCORE wiring disabled: SCORE_BASE_URL is empty")
	}

	store := network.NewSCORETenantNetworkStoreFixed(cfg.BaseURL, cfg.APIKey, logger)
	allocator := network.NewSCORETenantNetworkAllocatorProduction(cfg.BaseURL, cfg.APIKey, logger)

	rm := NewRelayManagerHybrid(logger, repository, store, allocator)
	logger.Info("SCORE hybrid relay manager wired",
		zap.String("base_url", cfg.BaseURL),
		zap.Bool("api_key_set", cfg.APIKey != ""),
		zap.Bool("repository_set", repository != nil),
	)
	return rm, nil
}
