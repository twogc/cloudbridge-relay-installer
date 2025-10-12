package congestion

import (
	"time"
)

// BBRv2 congestion control algorithm implementation
// Based on RFC 9002 and BBRv2 paper
// Integrated from quic-test experimental features

type BBRState int

const (
	BBRStartup BBRState = iota
	BBRDrain
	BBRProbeBW
	BBRProbeRTT
)

// BBRv2 implements the BBRv2 congestion control algorithm
type BBRv2 struct {
	state       BBRState
	mtu         int
	cwnd        int
	pacingBps   int64
	minRTT      time.Duration
	minRTTSince time.Time
	bwBps       float64 // bandwidth estimate
	cycleIdx    int
	lastStateTs time.Time
	pacer       *Pacer

	// Qlog callback for CloudBridge integration
	qlogCallback func(eventType string, data map[string]interface{})
}

// NewBBRv2 creates a new BBRv2 congestion controller
func NewBBRv2(mtu int, initialCWND int) *BBRv2 {
	if mtu <= 0 {
		mtu = 1460 // Standard Ethernet MTU
	}
	if initialCWND <= 0 {
		initialCWND = 32 * mtu // Conservative initial CWND
	}

	b := &BBRv2{
		state:       BBRStartup,
		mtu:         mtu,
		cwnd:        initialCWND,
		lastStateTs: time.Now(),
		pacer:       NewPacer(mtu),
	}
	return b
}

// Sample contains the input data for BBRv2
type Sample struct {
	RS   RateSample // from rate sampler
	RTT  time.Duration
	Loss bool
}

// OnAck handles an ACK event
func (b *BBRv2) OnAck(s Sample) (cwnd int, pacing int64) {
	now := time.Now()
	oldState := b.state
	oldCWND := b.cwnd
	oldPacing := b.pacingBps

	// Update min RTT
	if s.RTT > 0 && (b.minRTT == 0 || s.RTT < b.minRTT) {
		oldMinRTT := b.minRTT
		b.minRTT = s.RTT
		b.minRTTSince = now

		// Log RTT update for CloudBridge monitoring
		b.logQlogEvent("rtt_update", map[string]interface{}{
			"old_rtt":      float64(oldMinRTT.Nanoseconds()) / 1e6,
			"new_rtt":      float64(s.RTT.Nanoseconds()) / 1e6,
			"min_rtt":      float64(b.minRTT.Nanoseconds()) / 1e6,
			"rtt_variance": 0.0, // TODO: implement RTT variance tracking
			"smoothed_rtt": float64(s.RTT.Nanoseconds()) / 1e6,
			"sample_count": 1,
		})
	}

	// Update bandwidth estimate
	if br := s.RS.BandwidthBps(); br > 0 && br > b.bwBps {
		b.bwBps = br

		// Log bandwidth sample for CloudBridge monitoring
		b.logQlogEvent("bandwidth_sample", map[string]interface{}{
			"sample_bandwidth":   br,
			"smoothed_bandwidth": b.bwBps,
			"interval":           s.RS.Interval.Seconds() * 1000,
			"bytes_acked":        s.RS.BytesAcked,
			"is_app_limited":     s.RS.IsAppLimited,
			"rtt":                float64(s.RTT.Nanoseconds()) / 1e6,
		})
	}

	// BBRv2 state machine
	switch b.state {
	case BBRStartup:
		// Aggressive growth phase - discover available bandwidth
		b.cwnd += max(1, int(s.RS.BytesAcked))
		b.pacingBps = int64(2.0 * b.bwBps) // 2x bandwidth during startup
		if now.Sub(b.lastStateTs) > 2*time.Second {
			b.state = BBRDrain
			b.lastStateTs = now
		}

	case BBRDrain:
		// Drain excess packets from network
		b.cwnd = int(b.bdp() * 1.0)
		b.pacingBps = int64(0.5 * b.bwBps) // Slow down to drain
		if now.Sub(b.lastStateTs) > 500*time.Millisecond {
			b.state = BBRProbeBW
			b.cycleIdx = 0
			b.lastStateTs = now
		}

	case BBRProbeBW:
		// Probe bandwidth with gain cycling
		gains := []float64{1.25, 1.0, 0.75, 1.0}
		g := gains[b.cycleIdx%len(gains)]
		b.cwnd = int(g * b.bdp())
		b.pacingBps = int64(g * b.bwBps)
		if now.Sub(b.lastStateTs) > 300*time.Millisecond {
			b.cycleIdx++
			b.lastStateTs = now
		}
		// Periodically probe RTT to detect path changes
		if b.minRTT > 0 && now.Sub(b.minRTTSince) > 5*time.Second {
			b.state = BBRProbeRTT
			b.lastStateTs = now
		}

	case BBRProbeRTT:
		// Reduce to 0.5*BDP to measure new RTT
		b.cwnd = int(0.5 * b.bdp())
		b.pacingBps = int64(0.5 * b.bwBps)
		if now.Sub(b.lastStateTs) > 200*time.Millisecond {
			b.minRTTSince = now
			b.state = BBRProbeBW
			b.lastStateTs = now
		}
	}

	// Log state change if it occurred
	if oldState != b.state {
		b.logQlogEvent("state_change", map[string]interface{}{
			"old_state":   b.GetStateStringFromState(oldState),
			"new_state":   b.GetStateString(),
			"reason":      "timeout_or_condition",
			"bandwidth":   b.bwBps,
			"min_rtt":     float64(b.minRTT.Nanoseconds()) / 1e6,
			"cwnd":        b.cwnd,
			"pacing_rate": b.pacingBps,
		})
	}

	// Ensure minimum cwnd (at least 2 MTUs)
	if b.cwnd < 2*b.mtu {
		b.cwnd = 2 * b.mtu
	}

	// Set pacing rate
	if b.pacingBps <= 0 && b.minRTT > 0 {
		b.pacingBps = int64(float64(b.cwnd) / b.minRTT.Seconds())
	}
	b.pacer.SetRate(b.pacingBps)

	// Log CWND update if it changed
	if oldCWND != b.cwnd {
		b.logQlogEvent("cwnd_update", map[string]interface{}{
			"old_cwnd":         oldCWND,
			"new_cwnd":         b.cwnd,
			"change":           b.cwnd - oldCWND,
			"reason":           "ack_processing",
			"bandwidth":        b.bwBps,
			"min_rtt":          float64(b.minRTT.Nanoseconds()) / 1e6,
			"packets_in_flight": 0, // TODO: integrate with CloudBridge packet tracking
		})
	}

	// Log pacing update if it changed
	if oldPacing != b.pacingBps {
		b.logQlogEvent("pacing_update", map[string]interface{}{
			"old_rate":   oldPacing,
			"new_rate":   b.pacingBps,
			"tokens":     b.pacer.GetTokens(),
			"burst_size": 10 * b.mtu,
			"bandwidth":  b.bwBps,
			"min_rtt":    float64(b.minRTT.Nanoseconds()) / 1e6,
		})
	}

	return b.cwnd, b.pacingBps
}

// OnLoss handles a loss event
func (b *BBRv2) OnLoss() (cwnd int, pacing int64) {
	// Conservative loss response - reduce cwnd by 30%
	b.cwnd = int(0.7 * float64(b.cwnd))
	if b.cwnd < 2*b.mtu {
		b.cwnd = 2 * b.mtu
	}

	b.logQlogEvent("loss_detected", map[string]interface{}{
		"new_cwnd":  b.cwnd,
		"reduction": "30%",
		"state":     b.GetStateString(),
	})

	return b.cwnd, b.pacingBps
}

// bdp calculates the bandwidth-delay product
func (b *BBRv2) bdp() float64 {
	if b.minRTT <= 0 {
		return float64(b.cwnd)
	}
	return b.bwBps * b.minRTT.Seconds()
}

// GetState returns the current BBRv2 state
func (b *BBRv2) GetState() BBRState {
	return b.state
}

// GetCWND returns the current congestion window
func (b *BBRv2) GetCWND() int {
	return b.cwnd
}

// GetPacingRate returns the current pacing rate
func (b *BBRv2) GetPacingRate() int64 {
	return b.pacingBps
}

// GetBandwidth returns the current bandwidth estimate
func (b *BBRv2) GetBandwidth() float64 {
	return b.bwBps
}

// GetMinRTT returns the minimum RTT
func (b *BBRv2) GetMinRTT() time.Duration {
	return b.minRTT
}

// SetQlogCallback sets the callback for qlog events (CloudBridge integration)
func (b *BBRv2) SetQlogCallback(callback func(eventType string, data map[string]interface{})) {
	b.qlogCallback = callback
}

// logQlogEvent logs an event to qlog (CloudBridge integration)
func (b *BBRv2) logQlogEvent(eventType string, data map[string]interface{}) {
	if b.qlogCallback != nil {
		b.qlogCallback(eventType, data)
	}
}

// GetStateString returns the string representation of the current state
func (b *BBRv2) GetStateString() string {
	return b.GetStateStringFromState(b.state)
}

// GetStateStringFromState returns the string representation of a state
func (b *BBRv2) GetStateStringFromState(state BBRState) string {
	switch state {
	case BBRStartup:
		return "Startup"
	case BBRDrain:
		return "Drain"
	case BBRProbeBW:
		return "ProbeBW"
	case BBRProbeRTT:
		return "ProbeRTT"
	default:
		return "Unknown"
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
