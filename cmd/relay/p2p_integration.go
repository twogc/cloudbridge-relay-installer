package main

import (
	"context"
	"fmt"

"github.com/twogc/cloudbridge-relay-installer/config"
	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
	"github.com/twogc/cloudbridge-relay-installer/internal/p2p"
	"go.uber.org/zap"
)

// P2PIntegration представляет интеграцию P2P компонентов
type P2PIntegration struct {
	logger             *logger.Logger
	config             *config.Config
	relayManager       *p2p.RelayManager
	relayManagerHybrid *p2p.RelayManagerHybrid
	heartbeatManager   *p2p.HeartbeatManager
	peerManager        *p2p.SimplePeerManager
	tokenValidator     *p2p.TokenValidator
	ctx                context.Context
	cancel             context.CancelFunc
}

// NewP2PIntegration создает новую P2P интеграцию
func NewP2PIntegration(logger *logger.Logger, config *config.Config, jwtSecret []byte) (*P2PIntegration, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Create components
	relayManager := p2p.NewRelayManager(logger.GetZapLogger(), nil) // TODO: Initialize repository in production
	heartbeatManager := p2p.NewHeartbeatManager(relayManager, logger.GetZapLogger())
	peerManager := p2p.NewPeerManager(logger.GetZapLogger())
	tokenValidator := p2p.NewTokenValidator(jwtSecret, logger.GetZapLogger())

	// Optional SCORE hybrid path: SCORE_BASE_URL (+ SCORE_API_KEY).
	// Wires production SCORETenantNetworkAllocatorProduction into RelayManagerHybrid.
	var hybrid *p2p.RelayManagerHybrid
	scoreCfg := p2p.SCOREWiringFromEnv()
	if scoreCfg.Enabled() {
		hm, err := p2p.NewRelayManagerHybridFromSCORE(logger.GetZapLogger(), scoreCfg, nil)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("SCORE hybrid wiring: %w", err)
		}
		hybrid = hm
		logger.Info("SCORE hybrid relay manager enabled (SCORE_BASE_URL set)")
	} else {
		logger.Debug("SCORE hybrid wiring skipped (SCORE_BASE_URL empty)")
	}

	return &P2PIntegration{
		logger:             logger,
		config:             config,
		relayManager:       relayManager,
		relayManagerHybrid: hybrid,
		heartbeatManager:   heartbeatManager,
		peerManager:        peerManager,
		tokenValidator:     tokenValidator,
		ctx:                ctx,
		cancel:             cancel,
	}, nil
}

// Start запускает P2P интеграцию
func (pi *P2PIntegration) Start() error {
	pi.logger.Info("Starting P2P Integration")

	// Start heartbeat manager
	if err := pi.heartbeatManager.Start(); err != nil {
		return fmt.Errorf("failed to start heartbeat manager: %v", err)
	}

	// Peer callbacks are handled internally by SimplePeerManager
	pi.logger.Debug("P2P peer callbacks configured")

	pi.logger.Info("P2P Integration started successfully")
	return nil
}

// Stop останавливает P2P интеграцию
func (pi *P2PIntegration) Stop() error {
	pi.logger.Info("Stopping P2P Integration")

	pi.cancel()

	// Stop components
	if pi.heartbeatManager != nil {
		pi.heartbeatManager.Stop()
	}

	if pi.peerManager != nil {
		pi.peerManager.Shutdown()
	}

	if pi.relayManagerHybrid != nil {
		if err := pi.relayManagerHybrid.Shutdown(); err != nil {
			pi.logger.Warn("Failed to shutdown hybrid relay manager", zap.Error(err))
		}
	}

	if pi.relayManager != nil {
		pi.relayManager.Shutdown()
	}

	pi.logger.Info("P2P Integration stopped")
	return nil
}

// GetRelayManager возвращает relay manager
func (pi *P2PIntegration) GetRelayManager() *p2p.RelayManager {
	return pi.relayManager
}

// GetRelayManagerHybrid returns SCORE-wired hybrid manager (nil if SCORE_BASE_URL unset).
func (pi *P2PIntegration) GetRelayManagerHybrid() *p2p.RelayManagerHybrid {
	return pi.relayManagerHybrid
}

// GetHeartbeatManager возвращает heartbeat manager
func (pi *P2PIntegration) GetHeartbeatManager() *p2p.HeartbeatManager {
	return pi.heartbeatManager
}

// GetPeerManager возвращает peer manager
func (pi *P2PIntegration) GetPeerManager() *p2p.SimplePeerManager {
	return pi.peerManager
}

// GetTokenValidator возвращает token validator
func (pi *P2PIntegration) GetTokenValidator() *p2p.TokenValidator {
	return pi.tokenValidator
}
