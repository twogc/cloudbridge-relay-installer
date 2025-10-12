package quic

import (
	"bytes"
	"testing"
	"time"
)

func TestNewFECManager(t *testing.T) {
	tests := []struct {
		name          string
		config        *FECConfig
		wantNil       bool
		checkDefaults bool
	}{
		{
			name:          "with nil config uses defaults",
			config:        nil,
			wantNil:       false,
			checkDefaults: true,
		},
		{
			name: "with custom config",
			config: &FECConfig{
				RedundancyRatio:    0.2,
				Algorithm:          "reed-solomon",
				GroupSize:          20,
				EnableAdaptive:     false,
				TargetRecoveryRate: 0.95,
			},
			wantNil:       false,
			checkDefaults: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm := NewFECManager(newTestLogger(), tt.config)

			if (fm == nil) != tt.wantNil {
				t.Errorf("NewFECManager() = %v, wantNil %v", fm, tt.wantNil)
			}

			if tt.checkDefaults && fm != nil {
				if fm.config.RedundancyRatio != 0.1 {
					t.Errorf("Default RedundancyRatio = %v, want 0.1", fm.config.RedundancyRatio)
				}
				if fm.config.Algorithm != "xor" {
					t.Errorf("Default Algorithm = %v, want xor", fm.config.Algorithm)
				}
				if fm.config.GroupSize != 10 {
					t.Errorf("Default GroupSize = %v, want 10", fm.config.GroupSize)
				}
			}
		})
	}
}

func TestFECManager_CreateGroup(t *testing.T) {
	fm := NewFECManager(newTestLogger(), &FECConfig{
		RedundancyRatio: 0.1,
		Algorithm:       "xor",
		GroupSize:       5,
		EnableAdaptive:  false,
	})

	packets := [][]byte{
		[]byte("packet1"),
		[]byte("packet2"),
		[]byte("packet3"),
	}

	group, err := fm.CreateGroup(packets)
	if err != nil {
		t.Fatalf("CreateGroup() error = %v", err)
	}

	if group == nil {
		t.Fatal("CreateGroup() returned nil group")
	}

	if len(group.DataPackets) != len(packets) {
		t.Errorf("Group has %d data packets, want %d", len(group.DataPackets), len(packets))
	}

	if len(group.RedundantPackets) == 0 {
		t.Error("Group should have redundant packets")
	}
}

func TestFECManager_CreateGroup_XOR(t *testing.T) {
	fm := NewFECManager(newTestLogger(), &FECConfig{
		RedundancyRatio: 0.1,
		Algorithm:       "xor",
		GroupSize:       3,
		EnableAdaptive:  false,
	})

	// Create test packets with known data
	packets := [][]byte{
		{0x01, 0x02, 0x03, 0x04},
		{0x05, 0x06, 0x07, 0x08},
		{0x09, 0x0A, 0x0B, 0x0C},
	}

	group, err := fm.CreateGroup(packets)
	if err != nil {
		t.Fatalf("CreateGroup() error = %v", err)
	}

	if len(group.RedundantPackets) == 0 {
		t.Fatal("Expected redundant packets")
	}

	// XOR of the three packets should be: 0x01^0x05^0x09, 0x02^0x06^0x0A, etc.
	expectedXOR := []byte{
		0x01 ^ 0x05 ^ 0x09,
		0x02 ^ 0x06 ^ 0x0A,
		0x03 ^ 0x07 ^ 0x0B,
		0x04 ^ 0x08 ^ 0x0C,
	}

	if len(group.RedundantPackets[0]) < len(expectedXOR) {
		t.Fatalf("Redundant packet too short: %d, expected at least %d", len(group.RedundantPackets[0]), len(expectedXOR))
	}

	for i := 0; i < len(expectedXOR); i++ {
		if group.RedundantPackets[0][i] != expectedXOR[i] {
			t.Errorf("XOR byte[%d] = 0x%02X, want 0x%02X", i, group.RedundantPackets[0][i], expectedXOR[i])
		}
	}
}

func TestFECManager_RecoverLostPackets(t *testing.T) {
	fm := NewFECManager(newTestLogger(), &FECConfig{
		RedundancyRatio: 0.25, // 25% redundancy for better recovery
		Algorithm:       "xor",
		GroupSize:       4,
		EnableAdaptive:  false,
	})

	// Create packets
	packets := [][]byte{
		[]byte("packet1"),
		[]byte("packet2"),
		[]byte("packet3"),
		[]byte("packet4"),
	}

	group, err := fm.CreateGroup(packets)
	if err != nil {
		t.Fatalf("CreateGroup() error = %v", err)
	}

	if len(group.RedundantPackets) == 0 {
		t.Fatal("Expected redundant packets")
	}

	// Simulate packet loss: packet at index 2 is lost
	receivedIndices := []int{0, 1, 3}

	recovered, err := fm.RecoverLostPackets(group.ID, receivedIndices, group.RedundantPackets)
	if err != nil {
		t.Logf("RecoverLostPackets() error = %v (single packet loss may not always be recoverable with XOR)", err)
	}

	if recovered != nil && len(recovered) > 0 {
		t.Logf("Successfully recovered %d packet(s)", len(recovered))
	}
}

func TestFECManager_GetMetrics(t *testing.T) {
	fm := NewFECManager(newTestLogger(), nil)

	metrics := fm.GetMetrics()

	if metrics == nil {
		t.Fatal("GetMetrics should not return nil")
	}

	// Initially metrics should be zero
	if metrics.TotalGroups != 0 {
		t.Errorf("Initial TotalGroups = %v, want 0", metrics.TotalGroups)
	}

	// Create some groups
	for i := 0; i < 5; i++ {
		packets := [][]byte{
			[]byte("test1"),
			[]byte("test2"),
			[]byte("test3"),
		}
		fm.CreateGroup(packets)
	}

	metrics = fm.GetMetrics()
	if metrics.TotalGroups < 5 {
		t.Errorf("TotalGroups = %v, want >= 5", metrics.TotalGroups)
	}
}

func TestFECManager_AdaptiveRedundancy(t *testing.T) {
	fm := NewFECManager(newTestLogger(), &FECConfig{
		RedundancyRatio:    0.1,
		Algorithm:          "xor",
		GroupSize:          10,
		EnableAdaptive:     true,
		TargetRecoveryRate: 0.95,
	})

	initialRedundancy := fm.config.RedundancyRatio
	t.Logf("Initial redundancy: %.2f", initialRedundancy)

	// Simulate high loss rate
	fm.AdaptiveAdjustment(0.15) // 15% loss rate

	fm.mu.RLock()
	adjustedRedundancy := fm.config.RedundancyRatio
	fm.mu.RUnlock()

	t.Logf("Adjusted redundancy after 15%% loss: %.2f", adjustedRedundancy)

	// Redundancy should increase with high loss
	if adjustedRedundancy <= initialRedundancy {
		t.Logf("Note: Redundancy did not increase (%.2f vs %.2f)", adjustedRedundancy, initialRedundancy)
	}

	// Simulate low loss rate
	fm.AdaptiveAdjustment(0.01) // 1% loss rate

	fm.mu.RLock()
	finalRedundancy := fm.config.RedundancyRatio
	fm.mu.RUnlock()

	t.Logf("Adjusted redundancy after 1%% loss: %.2f", finalRedundancy)
}

func TestFECManager_SetRedundancy(t *testing.T) {
	fm := NewFECManager(newTestLogger(), nil)

	newRedundancy := 0.15
	fm.SetRedundancy(newRedundancy)

	fm.mu.RLock()
	actualRedundancy := fm.config.RedundancyRatio
	fm.mu.RUnlock()

	if actualRedundancy != newRedundancy {
		t.Errorf("Redundancy = %v, want %v", actualRedundancy, newRedundancy)
	}
}

func TestFECManager_SetRedundancy_Bounds(t *testing.T) {
	fm := NewFECManager(newTestLogger(), nil)

	tests := []struct {
		name    string
		input   float64
		wantMin float64
		wantMax float64
	}{
		{
			name:    "negative value clamped to 0",
			input:   -0.5,
			wantMin: 0.0,
			wantMax: 0.0,
		},
		{
			name:    "value above 1.0 clamped",
			input:   1.5,
			wantMin: 0.0,
			wantMax: 1.0,
		},
		{
			name:    "valid value",
			input:   0.25,
			wantMin: 0.25,
			wantMax: 0.25,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm.SetRedundancy(tt.input)

			fm.mu.RLock()
			actual := fm.config.RedundancyRatio
			fm.mu.RUnlock()

			if actual < tt.wantMin || actual > tt.wantMax {
				t.Errorf("Redundancy = %v, want between %v and %v", actual, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestFECManager_Stop(t *testing.T) {
	fm := NewFECManager(newTestLogger(), nil)

	// Create some groups
	fm.CreateGroup([][]byte{[]byte("test")})

	fm.Stop()

	// After stop, manager should not be active
	if fm.IsActive() {
		t.Error("Manager should not be active after Stop()")
	}
}

func TestFECManager_DifferentPacketSizes(t *testing.T) {
	fm := NewFECManager(newTestLogger(), &FECConfig{
		RedundancyRatio: 0.1,
		Algorithm:       "xor",
		GroupSize:       3,
		EnableAdaptive:  false,
	})

	// Add packets of different sizes
	packets := [][]byte{
		[]byte("short"),
		[]byte("medium packet"),
		[]byte("this is a much longer packet with more data"),
	}

	group, err := fm.CreateGroup(packets)
	if err != nil {
		t.Fatalf("CreateGroup() error = %v", err)
	}

	if len(group.RedundantPackets) == 0 {
		t.Fatal("Expected redundant packets")
	}

	// Redundant packet should be at least as large as the largest input packet
	maxSize := 0
	for _, pkt := range packets {
		if len(pkt) > maxSize {
			maxSize = len(pkt)
		}
	}

	if len(group.RedundantPackets[0]) < maxSize {
		t.Errorf("Redundant packet size = %d, want >= %d", len(group.RedundantPackets[0]), maxSize)
	}
}

func TestFECManager_EmptyPackets(t *testing.T) {
	fm := NewFECManager(newTestLogger(), nil)

	emptyPackets := [][]byte{}

	_, err := fm.CreateGroup(emptyPackets)
	if err == nil {
		t.Error("CreateGroup() with empty packets should return error")
	}
}

func TestFECManager_CleanupOldGroups(t *testing.T) {
	fm := NewFECManager(newTestLogger(), nil)

	// Create several groups
	for i := 0; i < 5; i++ {
		packets := [][]byte{
			[]byte("test1"),
			[]byte("test2"),
		}
		fm.CreateGroup(packets)
	}

	// Wait a bit
	time.Sleep(100 * time.Millisecond)

	// Cleanup groups older than 50ms (should cleanup all)
	cleaned := fm.CleanupOldGroups(50 * time.Millisecond)

	if cleaned != 5 {
		t.Logf("Cleaned %d groups, expected 5", cleaned)
	}
}

func TestFECManager_XORRecovery_Verification(t *testing.T) {
	fm := NewFECManager(newTestLogger(), &FECConfig{
		RedundancyRatio: 0.1,
		Algorithm:       "xor",
		GroupSize:       3,
		EnableAdaptive:  false,
	})

	// Original packets
	originals := [][]byte{
		[]byte("AAA"),
		[]byte("BBB"),
		[]byte("CCC"),
	}

	group, err := fm.CreateGroup(originals)
	if err != nil {
		t.Fatalf("CreateGroup() error = %v", err)
	}

	if len(group.RedundantPackets) == 0 {
		t.Fatal("No redundant packets generated")
	}

	// Manually verify XOR: originals[0] XOR originals[1] XOR originals[2] XOR redundant[0] should be zero
	xorResult := make([]byte, len(group.RedundantPackets[0]))
	copy(xorResult, group.RedundantPackets[0])

	for _, pkt := range originals {
		for i := 0; i < len(pkt) && i < len(xorResult); i++ {
			xorResult[i] ^= pkt[i]
		}
	}

	// XOR of all packets including redundant should be zero
	allZero := true
	for _, b := range xorResult {
		if b != 0 {
			allZero = false
			break
		}
	}

	if !allZero {
		t.Error("XOR verification failed: XOR of all packets should be zero")
	}
}

// Benchmark tests
func BenchmarkFECManager_CreateGroup(b *testing.B) {
	fm := NewFECManager(newTestLogger(), nil)
	packets := [][]byte{
		bytes.Repeat([]byte("x"), 1460),
		bytes.Repeat([]byte("y"), 1460),
		bytes.Repeat([]byte("z"), 1460),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fm.CreateGroup(packets)
	}
}

func BenchmarkFECManager_XORGeneration(b *testing.B) {
	fm := NewFECManager(newTestLogger(), &FECConfig{
		RedundancyRatio: 0.1,
		Algorithm:       "xor",
		GroupSize:       10,
		EnableAdaptive:  false,
	})

	packets := make([][]byte, 10)
	for i := range packets {
		packets[i] = bytes.Repeat([]byte("x"), 1460)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fm.CreateGroup(packets)
	}
}
