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

// resolvePeerIPAMRepository picks Postgres (prod) then SQLite (dev/fallback).
// Env:
//
//	Postgres: SCORE_DATABASE_URL | DATABASE_URL | SCORE_PG_HOST+…
//	SQLite:   SCORE_SQLITE_PATH (default shared memory)
func resolvePeerIPAMRepository(logger *zap.Logger, repository *database.Repository) *database.Repository {
	if repository != nil {
		return repository
	}
	if logger == nil {
		logger = zap.NewNop()
	}

	pgCfg := database.PostgresPeerIPAMFromEnv()
	if pgCfg.Enabled() {
		r, err := database.OpenPostgresPeerIPAM(pgCfg, logger)
		if err != nil {
			logger.Warn("Postgres peer IPAM unavailable; trying SQLite fallback",
				zap.Error(err),
			)
		} else {
			logger.Info("SCORE hybrid using Postgres peer IPAM")
			return r
		}
	}

	sqlitePath := strings.TrimSpace(os.Getenv("SCORE_SQLITE_PATH"))
	r, err := database.OpenSQLitePeerIPAM(sqlitePath, logger)
	if err != nil {
		logger.Warn("SQLite peer IPAM unavailable; hybrid continues without peer IP persistence",
			zap.Error(err),
		)
		return nil
	}
	logger.Info("SCORE hybrid using SQLite peer IPAM",
		zap.String("path", database.NormalizeSQLiteDSN(sqlitePath)),
	)
	return r
}

// NewRelayManagerHybridFromSCORE constructs RelayManagerHybrid with production
// SCORE network store + SCORETenantNetworkAllocatorProduction.
// Peer IPAM: injected repository, else Postgres env, else SQLite.
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

	repo := resolvePeerIPAMRepository(logger, repository)

	store := network.NewSCORETenantNetworkStoreFixed(cfg.BaseURL, cfg.APIKey, logger)
	allocator := network.NewSCORETenantNetworkAllocatorProduction(cfg.BaseURL, cfg.APIKey, logger)

	rm := NewRelayManagerHybrid(logger, repo, store, allocator)
	logger.Info("SCORE hybrid relay manager wired",
		zap.String("base_url", cfg.BaseURL),
		zap.Bool("api_key_set", cfg.APIKey != ""),
		zap.Bool("repository_set", repo != nil),
	)
	return rm, nil
}
