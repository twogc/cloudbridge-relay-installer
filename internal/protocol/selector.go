package protocol

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

// SelectorStrategy defines the strategy for protocol selection
type SelectorStrategy int

const (
	// PerformanceBased selects protocol based on performance metrics
	PerformanceBased SelectorStrategy = iota
	// RoundRobin cycles through protocols
	RoundRobin
	// Fixed uses a fixed protocol
	Fixed
)

// ProtocolSelectorConfig holds configuration for protocol selection
type ProtocolSelectorConfig struct {
	Strategy            SelectorStrategy
	PreferredOrder      []Protocol
	SwitchThreshold     float64
	HealthCheckInterval time.Duration
	MetricsWindow       time.Duration
	Logger              *zap.Logger
}

// IntelligentProtocolSelector implements intelligent protocol selection
type IntelligentProtocolSelector struct {
	engine         *ProtocolEngine
	strategy       SelectorStrategy
	preferredOrder []Protocol
	currentIndex   int
	lastSwitch     time.Time
	switchCooldown time.Duration
	healthTicker   *time.Ticker
	ctx            context.Context
	cancel         context.CancelFunc
	mu             sync.RWMutex
	logger         *zap.Logger
}

// NewIntelligentProtocolSelector creates a new protocol selector
func NewIntelligentProtocolSelector(config ProtocolSelectorConfig) *IntelligentProtocolSelector {
	if config.Logger == nil {
		config.Logger = zap.NewNop()
	}

	if len(config.PreferredOrder) == 0 {
		config.PreferredOrder = []Protocol{QUIC, HTTP2, WebSocket}
	}

	ctx, cancel := context.WithCancel(context.Background())

	selector := &IntelligentProtocolSelector{
		engine:         NewProtocolEngine(config.PreferredOrder, config.SwitchThreshold, config.Logger),
		strategy:       config.Strategy,
		preferredOrder: config.PreferredOrder,
		switchCooldown: 30 * time.Second, // Prevent rapid switching
		ctx:            ctx,
		cancel:         cancel,
		logger:         config.Logger,
	}

	// Start health check loop if interval is specified
	if config.HealthCheckInterval > 0 {
		selector.healthTicker = time.NewTicker(config.HealthCheckInterval)
		go selector.healthCheckLoop()
	}

	return selector
}

// Current returns the current protocol
func (ps *IntelligentProtocolSelector) Current() Protocol {
	return ps.engine.Current()
}

// Fallback returns the next available protocol
func (ps *IntelligentProtocolSelector) Fallback() (Protocol, bool) {
	return ps.engine.Fallback()
}

// ShouldSwitch determines if we should switch protocols
func (ps *IntelligentProtocolSelector) ShouldSwitch() bool {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	// Check cooldown period
	if time.Since(ps.lastSwitch) < ps.switchCooldown {
		return false
	}

	switch ps.strategy {
	case PerformanceBased:
		return ps.engine.ShouldSwitch()
	case RoundRobin:
		return ps.shouldSwitchRoundRobin()
	case Fixed:
		return false
	default:
		return false
	}
}

// shouldSwitchRoundRobin determines if we should switch in round-robin mode
func (ps *IntelligentProtocolSelector) shouldSwitchRoundRobin() bool {
	// Switch every 5 minutes in round-robin mode
	return time.Since(ps.lastSwitch) > 5*time.Minute
}

// SelectBest selects the best protocol based on the current strategy
func (ps *IntelligentProtocolSelector) SelectBest() Protocol {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	switch ps.strategy {
	case PerformanceBased:
		return ps.engine.SelectBest()
	case RoundRobin:
		return ps.selectRoundRobin()
	case Fixed:
		return ps.engine.Current()
	default:
		return ps.engine.Current()
	}
}

// selectRoundRobin selects the next protocol in round-robin fashion
func (ps *IntelligentProtocolSelector) selectRoundRobin() Protocol {
	ps.currentIndex = (ps.currentIndex + 1) % len(ps.preferredOrder)
	return ps.preferredOrder[ps.currentIndex]
}

// SwitchTo switches to the specified protocol
func (ps *IntelligentProtocolSelector) SwitchTo(protocol Protocol) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	oldProtocol := ps.engine.Current()
	err := ps.engine.SwitchTo(protocol)
	if err == nil {
		ps.lastSwitch = time.Now()
		ps.logger.Info("Protocol selector switched protocol",
			zap.String("from", oldProtocol.String()),
			zap.String("to", protocol.String()),
			zap.String("strategy", ps.strategy.String()),
		)
	}

	return err
}

// GetMetrics returns metrics for the current protocol
func (ps *IntelligentProtocolSelector) GetMetrics() *ProtocolMetrics {
	return ps.engine.GetMetrics()
}

// UpdateMetrics updates metrics for a specific protocol
func (ps *IntelligentProtocolSelector) UpdateMetrics(protocol Protocol, latency time.Duration, success bool) {
	ps.engine.UpdateMetrics(protocol, latency, success)
}

// healthCheckLoop performs periodic health checks
func (ps *IntelligentProtocolSelector) healthCheckLoop() {
	for {
		select {
		case <-ps.ctx.Done():
			return
		case <-ps.healthTicker.C:
			ps.performHealthCheck()
		}
	}
}

// performHealthCheck performs a health check and switches if necessary
func (ps *IntelligentProtocolSelector) performHealthCheck() {
	// Perform health check on current protocol
	err := ps.engine.HealthCheck(ps.ctx)
	if err != nil {
		ps.logger.Warn("Health check failed for current protocol",
			zap.String("protocol", ps.engine.Current().String()),
			zap.Error(err),
		)

		// Try to switch to a fallback protocol
		if fallback, available := ps.Fallback(); available {
			ps.logger.Info("Switching to fallback protocol due to health check failure",
				zap.String("from", ps.engine.Current().String()),
				zap.String("to", fallback.String()),
			)
			if err := ps.SwitchTo(fallback); err != nil {
				ps.logger.Error("Failed to switch to fallback protocol",
					zap.String("fallback", fallback.String()),
					zap.Error(err))
			}
		}
	}
}

// AutoSwitch enables automatic protocol switching
func (ps *IntelligentProtocolSelector) AutoSwitch() {
	go ps.autoSwitchLoop()
}

// autoSwitchLoop continuously monitors and switches protocols
func (ps *IntelligentProtocolSelector) autoSwitchLoop() {
	ticker := time.NewTicker(10 * time.Second) // Check every 10 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ps.ctx.Done():
			return
		case <-ticker.C:
			if ps.ShouldSwitch() {
				bestProtocol := ps.SelectBest()
				if bestProtocol != ps.Current() {
					if err := ps.SwitchTo(bestProtocol); err != nil {
						ps.logger.Error("Failed to switch to best protocol",
							zap.String("protocol", bestProtocol.String()),
							zap.Error(err))
					}
				}
			}
		}
	}
}

// GetStatus returns the current status of the protocol selector
func (ps *IntelligentProtocolSelector) GetStatus() map[string]interface{} {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	return map[string]interface{}{
		"current_protocol": ps.engine.Current().String(),
		"strategy":         ps.strategy.String(),
		"last_switch":      ps.lastSwitch,
		"switch_cooldown":  ps.switchCooldown.String(),
		"protocols":        ps.engine.GetProtocolInfo(),
		"metrics": map[string]interface{}{
			"latency":      ps.engine.GetMetrics().Latency.String(),
			"success_rate": ps.engine.GetMetrics().SuccessRate,
		},
	}
}

// Close closes the protocol selector
func (ps *IntelligentProtocolSelector) Close() error {
	ps.cancel()
	if ps.healthTicker != nil {
		ps.healthTicker.Stop()
	}
	return nil
}

// String returns the string representation of the selector strategy
func (s SelectorStrategy) String() string {
	switch s {
	case PerformanceBased:
		return "performance_based"
	case RoundRobin:
		return "round_robin"
	case Fixed:
		return "fixed"
	default:
		return "unknown"
	}
}
