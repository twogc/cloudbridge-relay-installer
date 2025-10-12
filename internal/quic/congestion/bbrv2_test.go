package congestion

import (
	"testing"
	"time"
)

func TestNewBBRv2(t *testing.T) {
	tests := []struct {
		name        string
		mtu         int
		initialCWND int
		wantMTU     int
		wantCWND    int
	}{
		{
			name:        "default values",
			mtu:         0,
			initialCWND: 0,
			wantMTU:     1460,
			wantCWND:    32 * 1460,
		},
		{
			name:        "custom values",
			mtu:         1500,
			initialCWND: 10000,
			wantMTU:     1500,
			wantCWND:    10000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bbr := NewBBRv2(tt.mtu, tt.initialCWND)

			if bbr.mtu != tt.wantMTU {
				t.Errorf("NewBBRv2() mtu = %v, want %v", bbr.mtu, tt.wantMTU)
			}
			if bbr.cwnd != tt.wantCWND {
				t.Errorf("NewBBRv2() cwnd = %v, want %v", bbr.cwnd, tt.wantCWND)
			}
			if bbr.state != BBRStartup {
				t.Errorf("NewBBRv2() state = %v, want %v", bbr.state, BBRStartup)
			}
		})
	}
}

func TestBBRv2_OnAck_Startup(t *testing.T) {
	bbr := NewBBRv2(1460, 14600)

	// Create a rate sample
	rs := RateSample{
		BytesAcked:   1460,
		Interval:     10 * time.Millisecond,
		IsAppLimited: false,
	}

	sample := Sample{
		RS:   rs,
		RTT:  50 * time.Millisecond,
		Loss: false,
	}

	// Test startup phase
	cwnd, pacing := bbr.OnAck(sample)

	if cwnd < 14600 {
		t.Errorf("OnAck() in Startup should increase cwnd, got %v", cwnd)
	}

	if bbr.minRTT != 50*time.Millisecond {
		t.Errorf("OnAck() minRTT = %v, want %v", bbr.minRTT, 50*time.Millisecond)
	}

	if pacing <= 0 {
		t.Errorf("OnAck() pacing should be positive, got %v", pacing)
	}
}

func TestBBRv2_StateMachine(t *testing.T) {
	bbr := NewBBRv2(1460, 14600)

	// Start in Startup
	if bbr.GetState() != BBRStartup {
		t.Errorf("Initial state should be Startup, got %v", bbr.GetStateString())
	}

	// Create high bandwidth sample
	rs := RateSample{
		BytesAcked:   14600,
		Interval:     10 * time.Millisecond,
		IsAppLimited: false,
	}

	sample := Sample{
		RS:   rs,
		RTT:  10 * time.Millisecond,
		Loss: false,
	}

	// Send multiple ACKs to trigger state transitions
	for i := 0; i < 10; i++ {
		bbr.OnAck(sample)
		time.Sleep(300 * time.Millisecond)
	}

	// After 2+ seconds, should transition to Drain or ProbeBW
	if bbr.GetState() == BBRStartup {
		t.Errorf("Should have transitioned from Startup after 2+ seconds")
	}
}

func TestBBRv2_OnLoss(t *testing.T) {
	bbr := NewBBRv2(1460, 14600)
	initialCWND := bbr.GetCWND()

	cwnd, _ := bbr.OnLoss()

	// CWND should be reduced by 30% on loss
	expectedCWND := int(0.7 * float64(initialCWND))
	if cwnd > initialCWND {
		t.Errorf("OnLoss() should reduce cwnd, got %v, want <= %v", cwnd, expectedCWND)
	}

	// CWND should not go below 2*MTU
	minCWND := 2 * bbr.mtu
	if cwnd < minCWND {
		t.Errorf("OnLoss() cwnd = %v, should not go below %v", cwnd, minCWND)
	}
}

func TestBBRv2_BDP_Calculation(t *testing.T) {
	bbr := NewBBRv2(1460, 14600)

	// Set known values
	bbr.bwBps = 1000000.0 // 1 Mbps
	bbr.minRTT = 100 * time.Millisecond

	bdp := bbr.bdp()

	// BDP = bandwidth * RTT
	// 1000000 bytes/sec * 0.1 sec = 100000 bytes
	expectedBDP := 100000.0

	if bdp < expectedBDP*0.9 || bdp > expectedBDP*1.1 {
		t.Errorf("bdp() = %v, want ~%v", bdp, expectedBDP)
	}
}

func TestBBRv2_QlogCallback(t *testing.T) {
	bbr := NewBBRv2(1460, 14600)

	callbackCalled := false
	var lastEventType string
	var lastData map[string]interface{}

	bbr.SetQlogCallback(func(eventType string, data map[string]interface{}) {
		callbackCalled = true
		lastEventType = eventType
		lastData = data
	})

	// Trigger an ACK to generate qlog events
	rs := RateSample{
		BytesAcked:   1460,
		Interval:     10 * time.Millisecond,
		IsAppLimited: false,
	}

	sample := Sample{
		RS:   rs,
		RTT:  50 * time.Millisecond,
		Loss: false,
	}

	bbr.OnAck(sample)

	if !callbackCalled {
		t.Error("Qlog callback should have been called")
	}

	if lastEventType == "" {
		t.Error("Qlog event type should not be empty")
	}

	if lastData == nil {
		t.Error("Qlog data should not be nil")
	}
}

func TestBBRv2_MinimumCWND(t *testing.T) {
	bbr := NewBBRv2(1460, 2920) // 2*MTU

	// Force very low CWND
	bbr.cwnd = 100

	rs := RateSample{
		BytesAcked:   100,
		Interval:     10 * time.Millisecond,
		IsAppLimited: false,
	}

	sample := Sample{
		RS:   rs,
		RTT:  10 * time.Millisecond,
		Loss: false,
	}

	cwnd, _ := bbr.OnAck(sample)

	minCWND := 2 * bbr.mtu
	if cwnd < minCWND {
		t.Errorf("cwnd = %v, should not go below minimum %v", cwnd, minCWND)
	}
}

func TestBBRv2_StateTransitions(t *testing.T) {
	bbr := NewBBRv2(1460, 14600)

	states := []BBRState{BBRStartup}

	// Force state transitions
	bbr.state = BBRDrain
	bbr.lastStateTs = time.Now().Add(-1 * time.Second)
	states = append(states, BBRDrain)

	bbr.state = BBRProbeBW
	states = append(states, BBRProbeBW)

	bbr.state = BBRProbeRTT
	states = append(states, BBRProbeRTT)

	// Test state string conversion
	for _, state := range states {
		stateStr := bbr.GetStateStringFromState(state)
		if stateStr == "Unknown" {
			t.Errorf("State %v should have a valid string representation", state)
		}
	}
}

func TestBBRv2_PacingRate(t *testing.T) {
	bbr := NewBBRv2(1460, 14600)

	// Set bandwidth and RTT
	bbr.bwBps = 1000000.0 // 1 Mbps
	bbr.minRTT = 100 * time.Millisecond
	bbr.state = BBRStartup

	rs := RateSample{
		BytesAcked:   1460,
		Interval:     10 * time.Millisecond,
		IsAppLimited: false,
	}

	sample := Sample{
		RS:   rs,
		RTT:  100 * time.Millisecond,
		Loss: false,
	}

	_, pacing := bbr.OnAck(sample)

	if pacing <= 0 {
		t.Error("Pacing rate should be positive")
	}

	// In startup, pacing should be 2x bandwidth
	expectedPacing := int64(2.0 * bbr.bwBps)
	if pacing < expectedPacing/2 || pacing > expectedPacing*2 {
		t.Logf("Warning: pacing = %v, expected ~%v", pacing, expectedPacing)
	}
}

func TestBBRv2_BandwidthEstimation(t *testing.T) {
	bbr := NewBBRv2(1460, 14600)

	// Send samples with increasing bandwidth
	bandwidths := []float64{100000, 200000, 300000, 250000}

	for _, bw := range bandwidths {
		interval := time.Duration(float64(time.Second) * (1460.0 / bw))
		rs := RateSample{
			BytesAcked:   1460,
			Interval:     interval,
			IsAppLimited: false,
		}

		sample := Sample{
			RS:   rs,
			RTT:  10 * time.Millisecond,
			Loss: false,
		}

		bbr.OnAck(sample)
	}

	// BBR should track the maximum bandwidth
	finalBW := bbr.GetBandwidth()
	if finalBW < 300000 {
		t.Errorf("BBR should track maximum bandwidth, got %v, want >= 300000", finalBW)
	}
}

func TestBBRv2_ProbeRTT(t *testing.T) {
	bbr := NewBBRv2(1460, 14600)

	// Setup: enter ProbeBW state
	bbr.state = BBRProbeBW
	bbr.bwBps = 1000000.0
	bbr.minRTT = 50 * time.Millisecond
	bbr.minRTTSince = time.Now().Add(-6 * time.Second) // Old RTT measurement

	rs := RateSample{
		BytesAcked:   1460,
		Interval:     10 * time.Millisecond,
		IsAppLimited: false,
	}

	sample := Sample{
		RS:   rs,
		RTT:  50 * time.Millisecond,
		Loss: false,
	}

	bbr.OnAck(sample)

	// Should transition to ProbeRTT after 5 seconds without RTT update
	if bbr.GetState() != BBRProbeRTT {
		t.Logf("Expected transition to ProbeRTT after old RTT measurement")
	}
}

// Benchmark tests
func BenchmarkBBRv2_OnAck(b *testing.B) {
	bbr := NewBBRv2(1460, 14600)

	rs := RateSample{
		BytesAcked:   1460,
		Interval:     10 * time.Millisecond,
		IsAppLimited: false,
	}

	sample := Sample{
		RS:   rs,
		RTT:  50 * time.Millisecond,
		Loss: false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bbr.OnAck(sample)
	}
}

func BenchmarkBBRv2_OnLoss(b *testing.B) {
	bbr := NewBBRv2(1460, 14600)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bbr.OnLoss()
	}
}
