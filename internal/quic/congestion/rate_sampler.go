package congestion

import (
	"time"
)

// RateSample represents a bandwidth measurement sample
// Used by BBRv2 to estimate available bandwidth
type RateSample struct {
	Delivered    int64         // total bytes delivered at time of packet send
	DeliveredAt  time.Time     // time when delivered measurement was captured
	FirstSentAt  time.Time     // time when the packet measuring the interval was sent
	Interval     time.Duration // DeliveredAt - FirstSentAt (with protections)
	BytesAcked   int64         // number of bytes acked in this ACK
	IsAppLimited bool          // whether packet was sent in app-limited state
}

// Sampler measures delivery rate for BBRv2
// Tracks bandwidth measurements across QUIC connections
type Sampler struct {
	delivered   int64     // total bytes delivered
	deliveredAt time.Time // time of last delivered update
	firstSentAt time.Time // time of first packet sent in measurement interval
	appLimited  bool      // flag indicating app-limited state
}

// NewSampler creates a new rate sampler
func NewSampler() *Sampler {
	return &Sampler{}
}

// OnPacketSent is called when a packet is sent
func (s *Sampler) OnPacketSent(now time.Time, size int, isAppLimited bool) {
	if s.firstSentAt.IsZero() {
		s.firstSentAt = now
	}

	// Update app-limited flag
	if isAppLimited {
		s.appLimited = true
	}
}

// OnAck is called when an ACK is received
// Returns a RateSample for bandwidth estimation
func (s *Sampler) OnAck(now time.Time, ackedBytes int) RateSample {
	s.delivered += int64(ackedBytes)

	rs := RateSample{
		Delivered:    s.delivered,
		DeliveredAt:  now,
		FirstSentAt:  s.firstSentAt,
		Interval:     now.Sub(s.firstSentAt),
		BytesAcked:   int64(ackedBytes),
		IsAppLimited: s.appLimited,
	}

	// Protection against zero/too short intervals
	if rs.Interval < 1*time.Millisecond {
		rs.Interval = 1 * time.Millisecond
	}

	// On first ACK, shift interval start
	s.firstSentAt = now
	s.appLimited = false

	return rs
}

// BandwidthBps returns the bandwidth in bytes per second
func (rs *RateSample) BandwidthBps() float64 {
	if rs.Interval <= 0 {
		return 0
	}
	return float64(rs.BytesAcked) / rs.Interval.Seconds()
}

// BandwidthMbps returns the bandwidth in megabits per second
func (rs *RateSample) BandwidthMbps() float64 {
	return rs.BandwidthBps() * 8 / (1024 * 1024)
}

// IsValid checks if the sample is valid
func (rs *RateSample) IsValid() bool {
	return rs.Interval > 0 && rs.BytesAcked > 0
}

// Reset resets the sampler state
func (s *Sampler) Reset() {
	s.delivered = 0
	s.deliveredAt = time.Time{}
	s.firstSentAt = time.Time{}
	s.appLimited = false
}

// GetDelivered returns the current number of delivered bytes
func (s *Sampler) GetDelivered() int64 {
	return s.delivered
}

// IsAppLimited returns whether we're in app-limited state
func (s *Sampler) IsAppLimited() bool {
	return s.appLimited
}
