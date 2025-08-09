package protocol

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

// Protocol represents the supported protocols
type Protocol int

const (
	HTTP2 Protocol = iota
	QUIC
	WebSocket
)

// String returns the string representation of the protocol
func (p Protocol) String() string {
	switch p {
	case HTTP2:
		return "http2"
	case QUIC:
		return "quic"
	case WebSocket:
		return "websocket"
	default:
		return "unknown"
	}
}

// ProtocolMetrics represents metrics for protocol performance
type ProtocolMetrics struct {
	Latency     time.Duration
	Throughput  float64 // bytes per second
	ErrorRate   float64 // errors per second
	SuccessRate float64 // success rate percentage
	LastUpdate  time.Time
}

// ProtocolSelector interface for protocol selection
type ProtocolSelector interface {
	Current() Protocol
	Fallback() (Protocol, bool)
	ShouldSwitch() bool
	GetMetrics() *ProtocolMetrics
	UpdateMetrics(protocol Protocol, latency time.Duration, success bool)
	SelectBest() Protocol
}

// ProtocolEngine implements the protocol selection logic
type ProtocolEngine struct {
	currentProtocol   Protocol
	preferredOrder    []Protocol
	metrics           map[Protocol]*ProtocolMetrics
	fallbackEnabled   bool
	switchThreshold   float64
	mu                sync.RWMutex
	logger            *zap.Logger
	prometheusMetrics *PrometheusMetrics
}

// PrometheusMetrics holds Prometheus metrics
type PrometheusMetrics struct {
	protocolLatency  *prometheus.HistogramVec
	protocolErrors   *prometheus.CounterVec
	protocolSwitches *prometheus.CounterVec
	protocolSuccess  *prometheus.CounterVec
}

// NewProtocolEngine creates a new protocol engine
func NewProtocolEngine(preferredOrder []Protocol, switchThreshold float64, logger *zap.Logger) *ProtocolEngine {
	if len(preferredOrder) == 0 {
		preferredOrder = []Protocol{QUIC, HTTP2, WebSocket}
	}

	engine := &ProtocolEngine{
		currentProtocol: preferredOrder[0],
		preferredOrder:  preferredOrder,
		metrics:         make(map[Protocol]*ProtocolMetrics),
		fallbackEnabled: true,
		switchThreshold: switchThreshold,
		logger:          logger,
		prometheusMetrics: &PrometheusMetrics{
			protocolLatency: prometheus.NewHistogramVec(
				prometheus.HistogramOpts{
					Name:    "protocol_latency_seconds",
					Help:    "Protocol latency in seconds",
					Buckets: prometheus.DefBuckets,
				},
				[]string{"protocol"},
			),
			protocolErrors: prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Name: "protocol_errors_total",
					Help: "Total number of protocol errors",
				},
				[]string{"protocol"},
			),
			protocolSwitches: prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Name: "protocol_switches_total",
					Help: "Total number of protocol switches",
				},
				[]string{"from", "to"},
			),
			protocolSuccess: prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Name: "protocol_success_total",
					Help: "Total number of successful protocol operations",
				},
				[]string{"protocol"},
			),
		},
	}

	// Initialize metrics for all protocols
	for _, protocol := range preferredOrder {
		engine.metrics[protocol] = &ProtocolMetrics{
			LastUpdate: time.Now(),
		}
	}

	// Register Prometheus metrics (ignore duplicates)
	collectors := []prometheus.Collector{
		engine.prometheusMetrics.protocolLatency,
		engine.prometheusMetrics.protocolErrors,
		engine.prometheusMetrics.protocolSwitches,
		engine.prometheusMetrics.protocolSuccess,
	}

	for _, collector := range collectors {
		if err := prometheus.Register(collector); err != nil {
			// Ignore already registered errors
			if !strings.Contains(err.Error(), "duplicate metrics collector") {
				logger.Error("Failed to register prometheus metric", zap.Error(err))
			}
		}
	}

	return engine
}

// Current returns the current protocol
func (pe *ProtocolEngine) Current() Protocol {
	pe.mu.RLock()
	defer pe.mu.RUnlock()
	return pe.currentProtocol
}

// Fallback returns the next available protocol in the preferred order
func (pe *ProtocolEngine) Fallback() (Protocol, bool) {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	currentIndex := -1
	for i, protocol := range pe.preferredOrder {
		if protocol == pe.currentProtocol {
			currentIndex = i
			break
		}
	}

	if currentIndex == -1 || currentIndex >= len(pe.preferredOrder)-1 {
		return pe.currentProtocol, false
	}

	return pe.preferredOrder[currentIndex+1], true
}

// ShouldSwitch determines if we should switch protocols based on performance
func (pe *ProtocolEngine) ShouldSwitch() bool {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	currentMetrics := pe.metrics[pe.currentProtocol]
	if currentMetrics == nil {
		return false
	}

	// Check if current protocol has high error rate
	if currentMetrics.ErrorRate > pe.switchThreshold {
		pe.logger.Info("Protocol switch recommended due to high error rate",
			zap.String("protocol", pe.currentProtocol.String()),
			zap.Float64("error_rate", currentMetrics.ErrorRate),
			zap.Float64("threshold", pe.switchThreshold),
		)
		return true
	}

	// Check if there's a better protocol available
	bestProtocol := pe.selectBestInternal()
	if bestProtocol != pe.currentProtocol {
		pe.logger.Info("Protocol switch recommended to better performing protocol",
			zap.String("current", pe.currentProtocol.String()),
			zap.String("recommended", bestProtocol.String()),
		)
		return true
	}

	return false
}

// GetMetrics returns metrics for the current protocol
func (pe *ProtocolEngine) GetMetrics() *ProtocolMetrics {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	if metrics, exists := pe.metrics[pe.currentProtocol]; exists {
		return metrics
	}
	return &ProtocolMetrics{}
}

// UpdateMetrics updates metrics for a specific protocol
func (pe *ProtocolEngine) UpdateMetrics(protocol Protocol, latency time.Duration, success bool) {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	metrics, exists := pe.metrics[protocol]
	if !exists {
		metrics = &ProtocolMetrics{
			LastUpdate: time.Now(),
		}
		pe.metrics[protocol] = metrics
	}

	// Update Prometheus metrics
	pe.prometheusMetrics.protocolLatency.WithLabelValues(protocol.String()).Observe(latency.Seconds())

	if success {
		pe.prometheusMetrics.protocolSuccess.WithLabelValues(protocol.String()).Inc()
		metrics.SuccessRate = 1.0
	} else {
		pe.prometheusMetrics.protocolErrors.WithLabelValues(protocol.String()).Inc()
		metrics.SuccessRate = 0.0
	}

	// Update latency (exponential moving average)
	if metrics.Latency == 0 {
		metrics.Latency = latency
	} else {
		// EMA with alpha = 0.1
		metrics.Latency = time.Duration(0.9*float64(metrics.Latency) + 0.1*float64(latency))
	}

	metrics.LastUpdate = time.Now()

	pe.logger.Debug("Updated protocol metrics",
		zap.String("protocol", protocol.String()),
		zap.Duration("latency", latency),
		zap.Bool("success", success),
		zap.Float64("success_rate", metrics.SuccessRate),
	)
}

// SelectBest returns the best performing protocol
func (pe *ProtocolEngine) SelectBest() Protocol {
	pe.mu.RLock()
	defer pe.mu.RUnlock()
	return pe.selectBestInternal()
}

// selectBestInternal selects the best protocol based on metrics
func (pe *ProtocolEngine) selectBestInternal() Protocol {
	var bestProtocol Protocol
	var bestScore float64

	for _, protocol := range pe.preferredOrder {
		metrics := pe.metrics[protocol]
		if metrics == nil {
			continue
		}

		// Calculate score based on latency and success rate
		latencyScore := 1.0 / (1.0 + float64(metrics.Latency.Milliseconds()))
		successScore := metrics.SuccessRate

		// Combined score (weighted average)
		score := 0.7*successScore + 0.3*latencyScore

		if score > bestScore {
			bestScore = score
			bestProtocol = protocol
		}
	}

	// If no metrics available, return the first preferred protocol
	if bestScore == 0 {
		return pe.preferredOrder[0]
	}

	return bestProtocol
}

// SwitchTo switches to the specified protocol
func (pe *ProtocolEngine) SwitchTo(protocol Protocol) error {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	oldProtocol := pe.currentProtocol
	pe.currentProtocol = protocol

	// Record the switch in Prometheus metrics
	pe.prometheusMetrics.protocolSwitches.WithLabelValues(
		oldProtocol.String(),
		protocol.String(),
	).Inc()

	pe.logger.Info("Switched protocol",
		zap.String("from", oldProtocol.String()),
		zap.String("to", protocol.String()),
	)

	return nil
}

// GetProtocolInfo returns information about all protocols
func (pe *ProtocolEngine) GetProtocolInfo() map[string]interface{} {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	info := make(map[string]interface{})
	for protocol, metrics := range pe.metrics {
		info[protocol.String()] = map[string]interface{}{
			"latency":      metrics.Latency.String(),
			"throughput":   metrics.Throughput,
			"error_rate":   metrics.ErrorRate,
			"success_rate": metrics.SuccessRate,
			"last_update":  metrics.LastUpdate,
			"is_current":   protocol == pe.currentProtocol,
		}
	}

	return info
}

// HealthCheck performs a health check on the current protocol
func (pe *ProtocolEngine) HealthCheck(ctx context.Context) error {
	protocol := pe.Current()

	pe.logger.Debug("Performing protocol health check",
		zap.String("protocol", protocol.String()),
	)

	// Simulate a health check (this would be implemented by the actual protocol)
	start := time.Now()

	// TODO: Implement actual health check for each protocol
	// For now, we'll just simulate a successful check
	time.Sleep(10 * time.Millisecond)

	latency := time.Since(start)
	pe.UpdateMetrics(protocol, latency, true)

	return nil
}
