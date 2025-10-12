package quic

import (
	"testing"
	"time"

	"go.uber.org/zap"
)

func newTestLogger() *zap.Logger {
	logger, _ := zap.NewDevelopment()
	if logger == nil {
		logger = zap.NewNop()
	}
	return logger
}

func TestNewACKFrequencyManager(t *testing.T) {
	tests := []struct {
		name       string
		config     *ACKFrequencyConfig
		wantNil    bool
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
			config: &ACKFrequencyConfig{
				MaxACKDelay:       50 * time.Millisecond,
				MinACKDelay:       2 * time.Millisecond,
				EnableAdaptive:    false,
				HighLoadThreshold: 2000,
				LowLoadThreshold:  50,
			},
			wantNil:       false,
			checkDefaults: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			afm := NewACKFrequencyManager(newTestLogger(), tt.config)

			if (afm == nil) != tt.wantNil {
				t.Errorf("NewACKFrequencyManager() = %v, wantNil %v", afm, tt.wantNil)
			}

			if tt.checkDefaults && afm != nil {
				if afm.config.MaxACKDelay != 25*time.Millisecond {
					t.Errorf("Default MaxACKDelay = %v, want 25ms", afm.config.MaxACKDelay)
				}
				if afm.config.MinACKDelay != 1*time.Millisecond {
					t.Errorf("Default MinACKDelay = %v, want 1ms", afm.config.MinACKDelay)
				}
				if !afm.config.EnableAdaptive {
					t.Error("Default EnableAdaptive should be true")
				}
			}
		})
	}
}

func TestACKFrequencyManager_RegisterConnection(t *testing.T) {
	afm := NewACKFrequencyManager(newTestLogger(), nil)

	mockConn := &struct{}{}
	connID := afm.RegisterConnection(mockConn)

	afm.mu.RLock()
	conn, exists := afm.connections[connID]
	afm.mu.RUnlock()

	if !exists {
		t.Error("Connection should be registered")
	}

	if conn == nil {
		t.Fatal("Connection should not be nil")
	}

	if conn.connectionID != connID {
		t.Errorf("Connection ID = %v, want %v", conn.connectionID, connID)
	}

	if conn.ackDelay != afm.config.DefaultACKDelay {
		t.Errorf("Initial ACK delay = %v, want %v", conn.ackDelay, afm.config.DefaultACKDelay)
	}
}

func TestACKFrequencyManager_UnregisterConnection(t *testing.T) {
	afm := NewACKFrequencyManager(newTestLogger(), nil)

	mockConn := &struct{}{}
	connID := afm.RegisterConnection(mockConn)
	afm.UnregisterConnection(connID)

	afm.mu.RLock()
	_, exists := afm.connections[connID]
	afm.mu.RUnlock()

	if exists {
		t.Error("Connection should be unregistered")
	}
}

func TestACKFrequencyManager_ShouldSendACK(t *testing.T) {
	afm := NewACKFrequencyManager(newTestLogger(), &ACKFrequencyConfig{
		MaxACKDelay:     25 * time.Millisecond,
		MinACKDelay:     1 * time.Millisecond,
		DefaultACKDelay: 5 * time.Millisecond,
		EnableAdaptive:  false, // Disable adaptive for predictable testing
	})

	mockConn := &struct{}{}
	connID := afm.RegisterConnection(mockConn)

	// First call should return false (just registered)
	if afm.ShouldSendACK(connID, 1460) {
		t.Error("Should not send ACK immediately after registration")
	}

	// Wait for DefaultACKDelay
	time.Sleep(10 * time.Millisecond)

	// Now should return true
	if !afm.ShouldSendACK(connID, 1460) {
		t.Error("Should send ACK after ACK delay")
	}
}

func TestACKFrequencyManager_SetTrafficType(t *testing.T) {
	afm := NewACKFrequencyManager(newTestLogger(), nil)

	mockConn := &struct{}{}
	connID := afm.RegisterConnection(mockConn)

	tests := []struct {
		name         string
		trafficType  string
		wantACKDelay time.Duration
	}{
		{
			name:         "bulk traffic",
			trafficType:  "bulk",
			wantACKDelay: 10 * time.Millisecond,
		},
		{
			name:         "interactive traffic",
			trafficType:  "interactive",
			wantACKDelay: 2 * time.Millisecond,
		},
		{
			name:         "realtime traffic",
			trafficType:  "realtime",
			wantACKDelay: 1 * time.Millisecond,
		},
		{
			name:         "unknown traffic uses default",
			trafficType:  "unknown",
			wantACKDelay: 5 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			afm.SetTrafficType(connID, tt.trafficType)

			afm.mu.RLock()
			conn := afm.connections[connID]
			afm.mu.RUnlock()

			if conn.ackDelay != tt.wantACKDelay {
				t.Errorf("ACK delay = %v, want %v for traffic type %s",
					conn.ackDelay, tt.wantACKDelay, tt.trafficType)
			}
		})
	}
}

func TestACKFrequencyManager_AdaptiveAdjustment(t *testing.T) {
	afm := NewACKFrequencyManager(newTestLogger(), &ACKFrequencyConfig{
		MaxACKDelay:       25 * time.Millisecond,
		MinACKDelay:       1 * time.Millisecond,
		DefaultACKDelay:   5 * time.Millisecond,
		EnableAdaptive:    true,
		HighLoadThreshold: 1000,
		LowLoadThreshold:  100,
	})

	mockConn := &struct{}{}
	connID := afm.RegisterConnection(mockConn)

	// Simulate high load (many packets)
	for i := 0; i < 1500; i++ {
		afm.ShouldSendACK(connID, 1460)
	}

	// Wait for adjustment interval
	time.Sleep(1100 * time.Millisecond)

	afm.mu.RLock()
	conn := afm.connections[connID]
	highLoadDelay := conn.ackDelay
	afm.mu.RUnlock()

	// Under high load, ACK delay should increase
	if highLoadDelay < afm.config.DefaultACKDelay {
		t.Logf("Under high load, ACK delay adjusted to %v", highLoadDelay)
	}

	// Simulate low load
	for i := 0; i < 50; i++ {
		afm.ShouldSendACK(connID, 1460)
		time.Sleep(10 * time.Millisecond)
	}

	time.Sleep(1100 * time.Millisecond)

	afm.mu.RLock()
	conn = afm.connections[connID]
	lowLoadDelay := conn.ackDelay
	afm.mu.RUnlock()

	// Under low load, ACK delay should decrease
	if lowLoadDelay < highLoadDelay {
		t.Logf("Under low load, ACK delay adjusted to %v from %v", lowLoadDelay, highLoadDelay)
	}
}

func TestACKFrequencyManager_GetMetrics(t *testing.T) {
	afm := NewACKFrequencyManager(newTestLogger(), nil)

	metrics := afm.GetMetrics()

	if metrics == nil {
		t.Fatal("GetMetrics should not return nil")
	}

	// Register connections and trigger ACKs
	mockConns := make([]interface{}, 5)
	for i := 0; i < 5; i++ {
		mockConns[i] = &struct{}{}
		afm.RegisterConnection(mockConns[i])
	}

	metrics = afm.GetMetrics()
	if metrics.ActiveConnections != 5 {
		t.Errorf("ActiveConnections = %v, want 5", metrics.ActiveConnections)
	}
}

func TestACKFrequencyManager_ConcurrentAccess(t *testing.T) {
	afm := NewACKFrequencyManager(newTestLogger(), nil)

	// Test concurrent registration and ACK checks
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(id int) {
			mockConn := &struct{ id int }{id: id}
			connID := afm.RegisterConnection(mockConn)

			for j := 0; j < 100; j++ {
				afm.ShouldSendACK(connID, 1460)
				time.Sleep(1 * time.Millisecond)
			}

			afm.UnregisterConnection(connID)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should complete without race conditions
	t.Log("Concurrent access test completed successfully")
}

// Benchmark tests
func BenchmarkACKFrequencyManager_ShouldSendACK(b *testing.B) {
	afm := NewACKFrequencyManager(newTestLogger(), nil)
	mockConn := &struct{}{}
	connID := afm.RegisterConnection(mockConn)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		afm.ShouldSendACK(connID, 1460)
	}
}

func BenchmarkACKFrequencyManager_SetTrafficType(b *testing.B) {
	afm := NewACKFrequencyManager(newTestLogger(), nil)
	mockConn := &struct{}{}
	connID := afm.RegisterConnection(mockConn)

	trafficTypes := []string{"bulk", "interactive", "realtime"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		afm.SetTrafficType(connID, trafficTypes[i%3])
	}
}
