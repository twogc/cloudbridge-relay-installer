package quic

import (
	"testing"
)

func TestNewGreasingManager(t *testing.T) {
	tests := []struct {
		name          string
		config        *GreasingConfig
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
			config: &GreasingConfig{
				EnableVersionGreasing:        false,
				EnableTransportParamGreasing: true,
				EnableFrameTypeGreasing:      false,
				GreaseProbability:            0.8,
				GreaseValues:                 []uint64{0x0a0a0a0a},
			},
			wantNil:       false,
			checkDefaults: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gm := NewGreasingManager(newTestLogger(), tt.config)

			if (gm == nil) != tt.wantNil {
				t.Errorf("NewGreasingManager() = %v, wantNil %v", gm, tt.wantNil)
			}

			if tt.checkDefaults && gm != nil {
				if !gm.config.EnableVersionGreasing {
					t.Error("Default EnableVersionGreasing should be true")
				}
				if !gm.config.EnableTransportParamGreasing {
					t.Error("Default EnableTransportParamGreasing should be true")
				}
				if !gm.config.EnableFrameTypeGreasing {
					t.Error("Default EnableFrameTypeGreasing should be true")
				}
				if gm.config.GreaseProbability != 0.5 {
					t.Errorf("Default GreaseProbability = %v, want 0.5", gm.config.GreaseProbability)
				}
				if len(gm.config.GreaseValues) != 16 {
					t.Errorf("Default GreaseValues count = %d, want 16", len(gm.config.GreaseValues))
				}
			}
		})
	}
}

func TestGreasingManager_DefaultGreaseValues(t *testing.T) {
	config := DefaultGreasingConfig()

	// RFC 9287 specifies 16 reserved grease values
	expectedValues := []uint64{
		0x0a0a0a0a, 0x1a1a1a1a, 0x2a2a2a2a, 0x3a3a3a3a,
		0x4a4a4a4a, 0x5a5a5a5a, 0x6a6a6a6a, 0x7a7a7a7a,
		0x8a8a8a8a, 0x9a9a9a9a, 0xaaaaaaaa, 0xbabababa,
		0xcacacaca, 0xdadadada, 0xeaeaeaea, 0xfafafafa,
	}

	if len(config.GreaseValues) != len(expectedValues) {
		t.Fatalf("GreaseValues count = %d, want %d", len(config.GreaseValues), len(expectedValues))
	}

	for i, expected := range expectedValues {
		if config.GreaseValues[i] != expected {
			t.Errorf("GreaseValues[%d] = 0x%x, want 0x%x", i, config.GreaseValues[i], expected)
		}
	}
}

func TestGreasingManager_ShouldGrease(t *testing.T) {
	tests := []struct {
		name        string
		probability float64
		iterations  int
	}{
		{
			name:        "always grease",
			probability: 1.0,
			iterations:  100,
		},
		{
			name:        "never grease",
			probability: 0.0,
			iterations:  100,
		},
		{
			name:        "50% probability",
			probability: 0.5,
			iterations:  1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gm := NewGreasingManager(newTestLogger(), &GreasingConfig{
				GreaseProbability: tt.probability,
				GreaseValues:      []uint64{0x0a0a0a0a},
			})

			greaseCount := 0
			for i := 0; i < tt.iterations; i++ {
				if gm.ShouldGrease() {
					greaseCount++
				}
			}

			rate := float64(greaseCount) / float64(tt.iterations)

			switch tt.probability {
			case 1.0:
				if greaseCount != tt.iterations {
					t.Errorf("With probability 1.0, all calls should grease, got %d/%d", greaseCount, tt.iterations)
				}
			case 0.0:
				if greaseCount != 0 {
					t.Errorf("With probability 0.0, no calls should grease, got %d/%d", greaseCount, tt.iterations)
				}
			case 0.5:
				// With 1000 iterations, expect roughly 40-60% due to randomness
				if rate < 0.4 || rate > 0.6 {
					t.Logf("Warning: grease rate = %.2f%%, expected ~50%%", rate*100)
				}
			}
		})
	}
}

func TestGreasingManager_GetGreaseValue(t *testing.T) {
	gm := NewGreasingManager(newTestLogger(), DefaultGreasingConfig())

	// Get multiple grease values and verify they're from the allowed list
	valuesSeen := make(map[uint64]bool)

	for i := 0; i < 100; i++ {
		value, err := gm.GetGreaseValue()
		if err != nil {
			t.Fatalf("GetGreaseValue() error = %v", err)
		}

		if !gm.ValidateGreaseValue(value) {
			t.Errorf("GetGreaseValue() returned invalid grease value: 0x%x", value)
		}

		valuesSeen[value] = true
	}

	// With 100 iterations and 16 possible values, we should see some variety
	if len(valuesSeen) < 5 {
		t.Logf("Only saw %d unique grease values in 100 iterations", len(valuesSeen))
	}
}

func TestGreasingManager_GetGreaseValue_EmptyList(t *testing.T) {
	gm := NewGreasingManager(newTestLogger(), &GreasingConfig{
		GreaseValues: []uint64{}, // Empty list
	})

	_, err := gm.GetGreaseValue()
	if err == nil {
		t.Error("GetGreaseValue() with empty GreaseValues should return error")
	}
}

func TestGreasingManager_GreaseVersionNegotiation(t *testing.T) {
	gm := NewGreasingManager(newTestLogger(), &GreasingConfig{
		EnableVersionGreasing: true,
		GreaseProbability:     1.0, // Always grease for testing
		GreaseValues:          []uint64{0x0a0a0a0a},
	})

	originalVersions := []uint32{1, 2, 3}

	greasedVersions, err := gm.GreaseVersionNegotiation(originalVersions)
	if err != nil {
		t.Fatalf("GreaseVersionNegotiation() error = %v", err)
	}

	// Should have one more version than original
	if len(greasedVersions) != len(originalVersions)+1 {
		t.Errorf("Greased versions length = %d, want %d", len(greasedVersions), len(originalVersions)+1)
	}

	// Should contain the grease value
	foundGrease := false
	for _, v := range greasedVersions {
		if v == uint32(0x0a0a0a0a) {
			foundGrease = true
			break
		}
	}

	if !foundGrease {
		t.Error("Greased versions should contain grease value")
	}

	// Should contain all original versions
	for _, orig := range originalVersions {
		found := false
		for _, greased := range greasedVersions {
			if greased == orig {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Greased versions missing original version: %d", orig)
		}
	}
}

func TestGreasingManager_GreaseVersionNegotiation_Disabled(t *testing.T) {
	gm := NewGreasingManager(newTestLogger(), &GreasingConfig{
		EnableVersionGreasing: false,
		GreaseProbability:     1.0,
		GreaseValues:          []uint64{0x0a0a0a0a},
	})

	originalVersions := []uint32{1, 2, 3}

	greasedVersions, err := gm.GreaseVersionNegotiation(originalVersions)
	if err != nil {
		t.Fatalf("GreaseVersionNegotiation() error = %v", err)
	}

	// Should return unchanged versions
	if len(greasedVersions) != len(originalVersions) {
		t.Errorf("With greasing disabled, versions should be unchanged")
	}
}

func TestGreasingManager_GreaseTransportParameters(t *testing.T) {
	gm := NewGreasingManager(newTestLogger(), &GreasingConfig{
		EnableTransportParamGreasing: true,
		GreaseProbability:            1.0,
		GreaseValues:                 []uint64{0x0a0a0a0a},
	})

	originalParams := map[uint64][]byte{
		0x01: []byte("param1"),
		0x02: []byte("param2"),
	}

	greasedParams, err := gm.GreaseTransportParameters(originalParams)
	if err != nil {
		t.Fatalf("GreaseTransportParameters() error = %v", err)
	}

	// Should have one more parameter
	if len(greasedParams) != len(originalParams)+1 {
		t.Errorf("Greased params length = %d, want %d", len(greasedParams), len(originalParams)+1)
	}

	// Should contain grease parameter
	_, hasGrease := greasedParams[0x0a0a0a0a]
	if !hasGrease {
		t.Error("Greased params should contain grease parameter")
	}

	// Should contain all original parameters
	for k, v := range originalParams {
		greasedV, exists := greasedParams[k]
		if !exists {
			t.Errorf("Greased params missing original key: %d", k)
		}
		if string(greasedV) != string(v) {
			t.Errorf("Greased params value changed for key %d", k)
		}
	}
}

func TestGreasingManager_GreaseTransportParameters_Disabled(t *testing.T) {
	gm := NewGreasingManager(newTestLogger(), &GreasingConfig{
		EnableTransportParamGreasing: false,
		GreaseProbability:            1.0,
		GreaseValues:                 []uint64{0x0a0a0a0a},
	})

	originalParams := map[uint64][]byte{
		0x01: []byte("param1"),
	}

	greasedParams, err := gm.GreaseTransportParameters(originalParams)
	if err != nil {
		t.Fatalf("GreaseTransportParameters() error = %v", err)
	}

	// Should return unchanged params
	if len(greasedParams) != len(originalParams) {
		t.Error("With greasing disabled, params should be unchanged")
	}
}

func TestGreasingManager_GreaseFrameType(t *testing.T) {
	gm := NewGreasingManager(newTestLogger(), &GreasingConfig{
		EnableFrameTypeGreasing: true,
		GreaseProbability:       1.0,
		GreaseValues:            []uint64{0x0a0a0a0a},
	})

	frameType, shouldGrease, err := gm.GreaseFrameType()
	if err != nil {
		t.Fatalf("GreaseFrameType() error = %v", err)
	}

	if !shouldGrease {
		t.Error("GreaseFrameType() should return true when greasing")
	}

	if frameType != 0x0a0a0a0a {
		t.Errorf("GreaseFrameType() = 0x%x, want 0x%x", frameType, 0x0a0a0a0a)
	}
}

func TestGreasingManager_GreaseFrameType_Disabled(t *testing.T) {
	gm := NewGreasingManager(newTestLogger(), &GreasingConfig{
		EnableFrameTypeGreasing: false,
		GreaseProbability:       1.0,
		GreaseValues:            []uint64{0x0a0a0a0a},
	})

	_, shouldGrease, err := gm.GreaseFrameType()
	if err != nil {
		t.Fatalf("GreaseFrameType() error = %v", err)
	}

	if shouldGrease {
		t.Error("GreaseFrameType() should return false when disabled")
	}
}

func TestGreasingManager_ValidateGreaseValue(t *testing.T) {
	gm := NewGreasingManager(newTestLogger(), DefaultGreasingConfig())

	tests := []struct {
		name  string
		value uint64
		want  bool
	}{
		{
			name:  "valid grease value",
			value: 0x0a0a0a0a,
			want:  true,
		},
		{
			name:  "another valid grease value",
			value: 0xfafafafa,
			want:  true,
		},
		{
			name:  "invalid value",
			value: 0x12345678,
			want:  false,
		},
		{
			name:  "zero",
			value: 0,
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := gm.ValidateGreaseValue(tt.value)
			if got != tt.want {
				t.Errorf("ValidateGreaseValue(0x%x) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestGreasingManager_IsGreaseValue(t *testing.T) {
	gm := NewGreasingManager(newTestLogger(), DefaultGreasingConfig())

	// IsGreaseValue should be same as ValidateGreaseValue
	testValues := []uint64{0x0a0a0a0a, 0xfafafafa, 0x12345678, 0}

	for _, value := range testValues {
		isGrease := gm.IsGreaseValue(value)
		validate := gm.ValidateGreaseValue(value)

		if isGrease != validate {
			t.Errorf("IsGreaseValue(0x%x) = %v, ValidateGreaseValue = %v, should be equal",
				value, isGrease, validate)
		}
	}
}

func TestGreasingManager_HandleCompatibilityIssue(t *testing.T) {
	gm := NewGreasingManager(newTestLogger(), nil)

	initialIssues := gm.GetMetrics().CompatibilityIssues

	gm.HandleCompatibilityIssue("Test issue 1")
	gm.HandleCompatibilityIssue("Test issue 2")

	metrics := gm.GetMetrics()
	if metrics.CompatibilityIssues != initialIssues+2 {
		t.Errorf("CompatibilityIssues = %d, want %d", metrics.CompatibilityIssues, initialIssues+2)
	}
}

func TestGreasingManager_GetMetrics(t *testing.T) {
	gm := NewGreasingManager(newTestLogger(), &GreasingConfig{
		EnableVersionGreasing:        true,
		EnableTransportParamGreasing: true,
		EnableFrameTypeGreasing:      true,
		GreaseProbability:            1.0,
		GreaseValues:                 []uint64{0x0a0a0a0a},
	})

	// Trigger various greasing operations
	gm.GreaseVersionNegotiation([]uint32{1, 2})
	gm.GreaseTransportParameters(map[uint64][]byte{1: []byte("test")})
	gm.GreaseFrameType()

	metrics := gm.GetMetrics()

	if metrics.VersionGreaseCount != 1 {
		t.Errorf("VersionGreaseCount = %d, want 1", metrics.VersionGreaseCount)
	}
	if metrics.TransportParamGreaseCount != 1 {
		t.Errorf("TransportParamGreaseCount = %d, want 1", metrics.TransportParamGreaseCount)
	}
	if metrics.FrameTypeGreaseCount != 1 {
		t.Errorf("FrameTypeGreaseCount = %d, want 1", metrics.FrameTypeGreaseCount)
	}
	if metrics.TotalGreaseApplications != 3 {
		t.Errorf("TotalGreaseApplications = %d, want 3", metrics.TotalGreaseApplications)
	}
}

func TestGreasingManager_SetGreaseProbability(t *testing.T) {
	gm := NewGreasingManager(newTestLogger(), nil)

	tests := []struct {
		name  string
		input float64
		want  float64
	}{
		{
			name:  "valid probability",
			input: 0.75,
			want:  0.75,
		},
		{
			name:  "negative clamped to 0",
			input: -0.5,
			want:  0.0,
		},
		{
			name:  "above 1.0 clamped",
			input: 1.5,
			want:  1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gm.SetGreaseProbability(tt.input)

			gm.mu.RLock()
			actual := gm.config.GreaseProbability
			gm.mu.RUnlock()

			if actual != tt.want {
				t.Errorf("GreaseProbability = %v, want %v", actual, tt.want)
			}
		})
	}
}

func TestGreasingManager_EnableDisableFeature(t *testing.T) {
	gm := NewGreasingManager(newTestLogger(), &GreasingConfig{
		EnableVersionGreasing:        false,
		EnableTransportParamGreasing: false,
		EnableFrameTypeGreasing:      false,
	})

	// Enable features
	gm.EnableFeature("version")
	gm.EnableFeature("transport_param")
	gm.EnableFeature("frame_type")

	gm.mu.RLock()
	if !gm.config.EnableVersionGreasing {
		t.Error("Version greasing should be enabled")
	}
	if !gm.config.EnableTransportParamGreasing {
		t.Error("Transport param greasing should be enabled")
	}
	if !gm.config.EnableFrameTypeGreasing {
		t.Error("Frame type greasing should be enabled")
	}
	gm.mu.RUnlock()

	// Disable features
	gm.DisableFeature("version")
	gm.DisableFeature("transport_param")
	gm.DisableFeature("frame_type")

	gm.mu.RLock()
	if gm.config.EnableVersionGreasing {
		t.Error("Version greasing should be disabled")
	}
	if gm.config.EnableTransportParamGreasing {
		t.Error("Transport param greasing should be disabled")
	}
	if gm.config.EnableFrameTypeGreasing {
		t.Error("Frame type greasing should be disabled")
	}
	gm.mu.RUnlock()
}

func TestGreasingManager_Stop(t *testing.T) {
	gm := NewGreasingManager(newTestLogger(), nil)

	if !gm.IsActive() {
		t.Error("Manager should be active initially")
	}

	gm.Stop()

	if gm.IsActive() {
		t.Error("Manager should not be active after Stop()")
	}

	// ShouldGrease should return false after stop
	if gm.ShouldGrease() {
		t.Error("ShouldGrease() should return false after stop")
	}
}

func TestGreasingManager_GetGreasingStatus(t *testing.T) {
	gm := NewGreasingManager(newTestLogger(), DefaultGreasingConfig())

	status := gm.GetGreasingStatus()

	if status == nil {
		t.Fatal("GetGreasingStatus() should not return nil")
	}

	expectedKeys := []string{
		"active",
		"version_greasing",
		"transport_param_greasing",
		"frame_type_greasing",
		"grease_probability",
		"total_grease_applications",
		"compatibility_issues",
	}

	for _, key := range expectedKeys {
		if _, exists := status[key]; !exists {
			t.Errorf("Status missing key: %s", key)
		}
	}

	if status["active"] != true {
		t.Error("Status active should be true")
	}
}

func TestGreasingManager_ConcurrentAccess(t *testing.T) {
	gm := NewGreasingManager(newTestLogger(), &GreasingConfig{
		EnableVersionGreasing:        true,
		EnableTransportParamGreasing: true,
		EnableFrameTypeGreasing:      true,
		GreaseProbability:            0.5,
		GreaseValues:                 []uint64{0x0a0a0a0a, 0x1a1a1a1a},
	})

	done := make(chan bool)

	// Concurrent version greasing
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				gm.GreaseVersionNegotiation([]uint32{1, 2, 3})
			}
			done <- true
		}()
	}

	// Concurrent transport param greasing
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				gm.GreaseTransportParameters(map[uint64][]byte{1: []byte("test")})
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}

	// Should complete without race conditions
	metrics := gm.GetMetrics()
	t.Logf("Concurrent access test completed: %d total grease applications", metrics.TotalGreaseApplications)
}

// Benchmark tests
func BenchmarkGreasingManager_ShouldGrease(b *testing.B) {
	gm := NewGreasingManager(newTestLogger(), nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gm.ShouldGrease()
	}
}

func BenchmarkGreasingManager_GetGreaseValue(b *testing.B) {
	gm := NewGreasingManager(newTestLogger(), nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gm.GetGreaseValue()
	}
}

func BenchmarkGreasingManager_GreaseVersionNegotiation(b *testing.B) {
	gm := NewGreasingManager(newTestLogger(), nil)
	versions := []uint32{1, 2, 3, 4, 5}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gm.GreaseVersionNegotiation(versions)
	}
}

func BenchmarkGreasingManager_ValidateGreaseValue(b *testing.B) {
	gm := NewGreasingManager(newTestLogger(), nil)
	value := uint64(0x0a0a0a0a)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gm.ValidateGreaseValue(value)
	}
}
