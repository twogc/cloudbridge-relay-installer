package congestion

import (
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// CongestionControlAlgorithm represents the type of congestion control algorithm
type CongestionControlAlgorithm string

const (
	// AlgorithmCubic is the default CUBIC algorithm
	AlgorithmCubic CongestionControlAlgorithm = "cubic"
	// AlgorithmBBR is the BBR algorithm
	AlgorithmBBR CongestionControlAlgorithm = "bbr"
	// AlgorithmBBRv2 is the BBRv2 algorithm (experimental)
	AlgorithmBBRv2 CongestionControlAlgorithm = "bbrv2"
	// AlgorithmReno is the classic TCP Reno algorithm
	AlgorithmReno CongestionControlAlgorithm = "reno"
)

// CongestionControlManager manages congestion control algorithms
// Provides a unified interface for CloudBridge Relay
type CongestionControlManager struct {
	logger    *zap.Logger
	algorithm CongestionControlAlgorithm
	mu        sync.RWMutex
	metrics   *CCMetrics
	isActive  bool

	// BBRv2 specific
	bbrv2       *BBRv2
	sampler     *Sampler
	mtu         int
	initialCWND int
}

// CCMetrics holds congestion control metrics
type CCMetrics struct {
	Algorithm          string        `json:"algorithm"`
	CongestionWindow   int64         `json:"congestion_window"`
	SlowStartThreshold int64         `json:"slow_start_threshold"`
	RTT                time.Duration `json:"rtt"`
	RTTVariance        time.Duration `json:"rtt_variance"`
	PacketsInFlight    int64         `json:"packets_in_flight"`
	LossRate           float64       `json:"loss_rate"`
	Throughput         float64       `json:"throughput_mbps"`
	PacingRate         int64         `json:"pacing_rate_bps"`
	Bandwidth          float64       `json:"bandwidth_bps"`
	LastUpdate         time.Time     `json:"last_update"`

	// BBRv2 specific
	BBRState           string        `json:"bbr_state,omitempty"`
	MinRTT             time.Duration `json:"min_rtt,omitempty"`
	BandwidthEstimate  float64       `json:"bandwidth_estimate,omitempty"`
}

// NewCongestionControlManager creates a new congestion control manager
func NewCongestionControlManager(logger *zap.Logger, algorithm CongestionControlAlgorithm, mtu int) *CongestionControlManager {
	if mtu <= 0 {
		mtu = 1460 // Default MTU
	}

	ccm := &CongestionControlManager{
		logger:      logger,
		algorithm:   algorithm,
		mtu:         mtu,
		initialCWND: 32 * mtu, // Conservative initial CWND
		metrics: &CCMetrics{
			Algorithm:  string(algorithm),
			LastUpdate: time.Now(),
		},
		isActive: true,
	}

	// Initialize algorithm-specific components
	if algorithm == AlgorithmBBRv2 {
		ccm.bbrv2 = NewBBRv2(mtu, ccm.initialCWND)
		ccm.sampler = NewSampler()

		// Set qlog callback for CloudBridge integration
		ccm.bbrv2.SetQlogCallback(func(eventType string, data map[string]interface{}) {
			logger.Debug("BBRv2 event",
				zap.String("event_type", eventType),
				zap.Any("data", data))
		})

		logger.Info("BBRv2 congestion control initialized",
			zap.Int("mtu", mtu),
			zap.Int("initial_cwnd", ccm.initialCWND))
	}

	return ccm
}

// GetAlgorithm returns the current algorithm
func (ccm *CongestionControlManager) GetAlgorithm() CongestionControlAlgorithm {
	ccm.mu.RLock()
	defer ccm.mu.RUnlock()
	return ccm.algorithm
}

// SetAlgorithm changes the congestion control algorithm
func (ccm *CongestionControlManager) SetAlgorithm(algorithm CongestionControlAlgorithm) error {
	ccm.mu.Lock()
	defer ccm.mu.Unlock()

	validAlgorithms := map[CongestionControlAlgorithm]bool{
		AlgorithmCubic:  true,
		AlgorithmBBR:    true,
		AlgorithmBBRv2:  true,
		AlgorithmReno:   true,
	}

	if !validAlgorithms[algorithm] {
		return fmt.Errorf("invalid congestion control algorithm: %s", algorithm)
	}

	oldAlgorithm := ccm.algorithm
	ccm.algorithm = algorithm
	ccm.metrics.Algorithm = string(algorithm)
	ccm.metrics.LastUpdate = time.Now()

	// Reinitialize BBRv2 if switching to it
	if algorithm == AlgorithmBBRv2 && ccm.bbrv2 == nil {
		ccm.bbrv2 = NewBBRv2(ccm.mtu, ccm.initialCWND)
		ccm.sampler = NewSampler()
	}

	ccm.logger.Info("Congestion control algorithm changed",
		zap.String("old_algorithm", string(oldAlgorithm)),
		zap.String("new_algorithm", string(algorithm)))

	return nil
}

// OnPacketSent is called when a packet is sent
func (ccm *CongestionControlManager) OnPacketSent(now time.Time, size int, isAppLimited bool) {
	ccm.mu.Lock()
	defer ccm.mu.Unlock()

	if ccm.algorithm == AlgorithmBBRv2 && ccm.sampler != nil {
		ccm.sampler.OnPacketSent(now, size, isAppLimited)
	}
}

// OnAck is called when an ACK is received
// Returns the new congestion window and pacing rate
func (ccm *CongestionControlManager) OnAck(now time.Time, ackedBytes int, rtt time.Duration) (cwnd int, pacingRate int64) {
	ccm.mu.Lock()
	defer ccm.mu.Unlock()

	if ccm.algorithm == AlgorithmBBRv2 && ccm.bbrv2 != nil && ccm.sampler != nil {
		// Get rate sample
		rs := ccm.sampler.OnAck(now, ackedBytes)

		// Feed to BBRv2
		sample := Sample{
			RS:   rs,
			RTT:  rtt,
			Loss: false,
		}
		cwnd, pacingRate = ccm.bbrv2.OnAck(sample)

		// Update metrics
		ccm.updateMetricsBBRv2(cwnd, pacingRate, rtt)

		return cwnd, pacingRate
	}

	// For other algorithms, return default values
	// TODO: Implement CUBIC, BBR, Reno
	return ccm.initialCWND, 0
}

// OnLoss is called when packet loss is detected
func (ccm *CongestionControlManager) OnLoss() (cwnd int, pacingRate int64) {
	ccm.mu.Lock()
	defer ccm.mu.Unlock()

	if ccm.algorithm == AlgorithmBBRv2 && ccm.bbrv2 != nil {
		cwnd, pacingRate = ccm.bbrv2.OnLoss()

		// Update metrics
		ccm.metrics.CongestionWindow = int64(cwnd)
		ccm.metrics.PacingRate = pacingRate
		ccm.metrics.LastUpdate = time.Now()

		ccm.logger.Debug("BBRv2 loss detected",
			zap.Int("new_cwnd", cwnd),
			zap.Int64("pacing_rate", pacingRate))

		return cwnd, pacingRate
	}

	// For other algorithms, return default values
	return ccm.initialCWND, 0
}

// updateMetricsBBRv2 updates metrics from BBRv2 state
func (ccm *CongestionControlManager) updateMetricsBBRv2(cwnd int, pacingRate int64, rtt time.Duration) {
	ccm.metrics.CongestionWindow = int64(cwnd)
	ccm.metrics.PacingRate = pacingRate
	ccm.metrics.RTT = rtt
	ccm.metrics.LastUpdate = time.Now()

	if ccm.bbrv2 != nil {
		ccm.metrics.BBRState = ccm.bbrv2.GetStateString()
		ccm.metrics.MinRTT = ccm.bbrv2.GetMinRTT()
		ccm.metrics.BandwidthEstimate = ccm.bbrv2.GetBandwidth()
		ccm.metrics.Bandwidth = ccm.bbrv2.GetBandwidth()

		// Calculate throughput in Mbps
		if ccm.metrics.BandwidthEstimate > 0 {
			ccm.metrics.Throughput = ccm.metrics.BandwidthEstimate * 8 / (1024 * 1024)
		}
	}
}

// UpdateMetrics manually updates congestion control metrics
// Used for non-BBRv2 algorithms
func (ccm *CongestionControlManager) UpdateMetrics(cwnd, ssthresh int64, rtt, rttVar time.Duration, packetsInFlight int64, lossRate, throughput float64) {
	ccm.mu.Lock()
	defer ccm.mu.Unlock()

	ccm.metrics.CongestionWindow = cwnd
	ccm.metrics.SlowStartThreshold = ssthresh
	ccm.metrics.RTT = rtt
	ccm.metrics.RTTVariance = rttVar
	ccm.metrics.PacketsInFlight = packetsInFlight
	ccm.metrics.LossRate = lossRate
	ccm.metrics.Throughput = throughput
	ccm.metrics.LastUpdate = time.Now()
}

// GetMetrics returns the current congestion control metrics
func (ccm *CongestionControlManager) GetMetrics() *CCMetrics {
	ccm.mu.RLock()
	defer ccm.mu.RUnlock()

	// Return a copy to prevent data races
	metricsCopy := *ccm.metrics
	return &metricsCopy
}

// GetBBRv2 returns the BBRv2 instance (if active)
func (ccm *CongestionControlManager) GetBBRv2() *BBRv2 {
	ccm.mu.RLock()
	defer ccm.mu.RUnlock()
	return ccm.bbrv2
}

// Stop stops the congestion control manager
func (ccm *CongestionControlManager) Stop() {
	ccm.mu.Lock()
	defer ccm.mu.Unlock()

	ccm.isActive = false
	ccm.logger.Info("Congestion control manager stopped")
}

// IsActive returns whether the manager is active
func (ccm *CongestionControlManager) IsActive() bool {
	ccm.mu.RLock()
	defer ccm.mu.RUnlock()
	return ccm.isActive
}

// Reset resets the congestion control state
func (ccm *CongestionControlManager) Reset() {
	ccm.mu.Lock()
	defer ccm.mu.Unlock()

	if ccm.sampler != nil {
		ccm.sampler.Reset()
	}

	if ccm.bbrv2 != nil {
		// Reinitialize BBRv2
		ccm.bbrv2 = NewBBRv2(ccm.mtu, ccm.initialCWND)
	}

	ccm.logger.Debug("Congestion control state reset")
}
