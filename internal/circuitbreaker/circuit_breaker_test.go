package circuitbreaker

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestNewCircuitBreaker(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		want   func(*CircuitBreaker) bool
	}{
		{
			name:   "default config",
			config: nil,
			want: func(cb *CircuitBreaker) bool {
				return cb.name == "default" &&
					cb.state == Closed &&
					cb.failureThreshold == 5 &&
					cb.successThreshold == 3 &&
					cb.timeout == 30*time.Second &&
					cb.windowSize == 60*time.Second &&
					cb.autoRecovery == true
			},
		},
		{
			name: "custom config",
			config: &Config{
				Name:             "test-cb",
				FailureThreshold: 10,
				SuccessThreshold: 5,
				Timeout:          60 * time.Second,
				WindowSize:       120 * time.Second,
				AutoRecovery:     false,
				Logger:           zaptest.NewLogger(t),
			},
			want: func(cb *CircuitBreaker) bool {
				return cb.name == "test-cb" &&
					cb.state == Closed &&
					cb.failureThreshold == 10 &&
					cb.successThreshold == 5 &&
					cb.timeout == 60*time.Second &&
					cb.windowSize == 120*time.Second &&
					cb.autoRecovery == false
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cb := NewCircuitBreaker(tt.config)
			assert.True(t, tt.want(cb))
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig("test")

	assert.Equal(t, "test", config.Name)
	assert.Equal(t, int64(5), config.FailureThreshold)
	assert.Equal(t, int64(3), config.SuccessThreshold)
	assert.Equal(t, 30*time.Second, config.Timeout)
	assert.Equal(t, 60*time.Second, config.WindowSize)
	assert.True(t, config.AutoRecovery)
}

func TestState_String(t *testing.T) {
	tests := []struct {
		state State
		want  string
	}{
		{Closed, "CLOSED"},
		{Open, "OPEN"},
		{HalfOpen, "HALF_OPEN"},
		{State(999), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.state.String())
		})
	}
}

func TestCircuitBreaker_Execute(t *testing.T) {
	cb := NewCircuitBreaker(&Config{
		Name:             "test",
		FailureThreshold: 2,
		SuccessThreshold: 1,
		Timeout:          100 * time.Millisecond,
		Logger:           zaptest.NewLogger(t),
	})

	t.Run("successful execution", func(t *testing.T) {
		ctx := context.Background()
		err := cb.Execute(ctx, func() error {
			return nil
		})
		assert.NoError(t, err)
		assert.Equal(t, Closed, cb.GetState())
	})

	t.Run("failed execution", func(t *testing.T) {
		ctx := context.Background()
		err := cb.Execute(ctx, func() error {
			return errors.New("test error")
		})
		assert.Error(t, err)
		assert.Equal(t, "test error", err.Error())
	})

	t.Run("timeout execution", func(t *testing.T) {
		ctx := context.Background()
		err := cb.Execute(ctx, func() error {
			time.Sleep(200 * time.Millisecond)
			return nil
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "timeout")
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := cb.Execute(ctx, func() error {
			time.Sleep(100 * time.Millisecond)
			return nil
		})
		assert.Error(t, err)
		// Circuit breaker might be open from previous test, so check for either error
		assert.True(t, err == context.Canceled || err.Error() == "circuit breaker is open")
	})
}

func TestExecuteWithResult(t *testing.T) {
	cb := NewCircuitBreaker(&Config{
		Name:             "test",
		FailureThreshold: 2,
		SuccessThreshold: 1,
		Timeout:          100 * time.Millisecond,
		Logger:           zaptest.NewLogger(t),
	})

	t.Run("successful execution with result", func(t *testing.T) {
		ctx := context.Background()
		result, err := ExecuteWithResult(cb, ctx, func() (string, error) {
			return "success", nil
		})
		assert.NoError(t, err)
		assert.Equal(t, "success", result)
	})

	t.Run("failed execution with result", func(t *testing.T) {
		ctx := context.Background()
		result, err := ExecuteWithResult(cb, ctx, func() (string, error) {
			return "", errors.New("test error")
		})
		assert.Error(t, err)
		assert.Equal(t, "", result)
		assert.Equal(t, "test error", err.Error())
	})

	t.Run("timeout execution with result", func(t *testing.T) {
		ctx := context.Background()
		result, err := ExecuteWithResult(cb, ctx, func() (string, error) {
			time.Sleep(200 * time.Millisecond)
			return "timeout", nil
		})
		assert.Error(t, err)
		assert.Equal(t, "", result)
		assert.Contains(t, err.Error(), "timeout")
	})
}

func TestCircuitBreaker_Ready(t *testing.T) {
	cb := NewCircuitBreaker(&Config{
		Name:             "test",
		FailureThreshold: 2,
		SuccessThreshold: 1,
		Timeout:          100 * time.Millisecond,
		Logger:           zaptest.NewLogger(t),
	})

	t.Run("closed state is ready", func(t *testing.T) {
		assert.True(t, cb.Ready())
	})

	t.Run("open state is not ready initially", func(t *testing.T) {
		// Force open state
		cb.ForceOpen()
		assert.False(t, cb.Ready())
	})

	t.Run("open state becomes ready after timeout", func(t *testing.T) {
		cb.ForceOpen()
		time.Sleep(150 * time.Millisecond)
		assert.True(t, cb.Ready())
		assert.Equal(t, HalfOpen, cb.GetState())
	})

	t.Run("half-open state is ready", func(t *testing.T) {
		cb.ForceOpen()
		time.Sleep(150 * time.Millisecond)
		cb.Ready() // This should transition to half-open
		assert.True(t, cb.Ready())
	})
}

func TestCircuitBreaker_GetState(t *testing.T) {
	cb := NewCircuitBreaker(&Config{
		Name:             "test",
		FailureThreshold: 1,
		SuccessThreshold: 1,
		Timeout:          100 * time.Millisecond,
		Logger:           zaptest.NewLogger(t),
	})

	assert.Equal(t, Closed, cb.GetState())

	// Trigger failure to open circuit
	cb.Execute(context.Background(), func() error {
		return errors.New("test error")
	})
	assert.Equal(t, Open, cb.GetState())
}

func TestCircuitBreaker_GetStats(t *testing.T) {
	cb := NewCircuitBreaker(&Config{
		Name:             "test",
		FailureThreshold: 2,
		SuccessThreshold: 1,
		Timeout:          100 * time.Millisecond,
		Logger:           zaptest.NewLogger(t),
	})

	// Execute some operations
	cb.Execute(context.Background(), func() error { return nil })
	cb.Execute(context.Background(), func() error { return errors.New("error") })

	stats := cb.GetStats()

	assert.Equal(t, "test", stats["name"])
	assert.Equal(t, "CLOSED", stats["state"])
	assert.Equal(t, int64(2), stats["total_requests"])
	assert.Equal(t, int64(1), stats["total_failures"])
	assert.Equal(t, int64(1), stats["total_successes"])
	assert.Equal(t, int64(2), stats["failure_threshold"])
	assert.Equal(t, int64(1), stats["success_threshold"])
	assert.Equal(t, "100ms", stats["timeout"])
	assert.Equal(t, "0s", stats["window_size"]) // Default window size is 0
}

func TestCircuitBreaker_ForceOpen(t *testing.T) {
	cb := NewCircuitBreaker(&Config{
		Name:             "test",
		FailureThreshold: 2,
		SuccessThreshold: 1,
		Timeout:          100 * time.Millisecond,
		Logger:           zaptest.NewLogger(t),
	})

	assert.Equal(t, Closed, cb.GetState())

	cb.ForceOpen()
	assert.Equal(t, Open, cb.GetState())
	assert.False(t, cb.Ready())
}

func TestCircuitBreaker_ForceClose(t *testing.T) {
	cb := NewCircuitBreaker(&Config{
		Name:             "test",
		FailureThreshold: 1,
		SuccessThreshold: 1,
		Timeout:          100 * time.Millisecond,
		Logger:           zaptest.NewLogger(t),
	})

	// Open the circuit
	cb.Execute(context.Background(), func() error {
		return errors.New("test error")
	})
	assert.Equal(t, Open, cb.GetState())

	// Force close
	cb.ForceClose()
	assert.Equal(t, Closed, cb.GetState())
	assert.True(t, cb.Ready())
}

func TestCircuitBreaker_Reset(t *testing.T) {
	cb := NewCircuitBreaker(&Config{
		Name:             "test",
		FailureThreshold: 1,
		SuccessThreshold: 1,
		Timeout:          100 * time.Millisecond,
		Logger:           zaptest.NewLogger(t),
	})

	// Execute some operations
	cb.Execute(context.Background(), func() error { return nil })
	cb.Execute(context.Background(), func() error { return errors.New("error") })

	// Reset
	cb.Reset()

	stats := cb.GetStats()
	assert.Equal(t, Closed, cb.GetState())
	assert.Equal(t, int64(0), stats["total_requests"])
	assert.Equal(t, int64(0), stats["total_failures"])
	assert.Equal(t, int64(0), stats["total_successes"])
}

func TestCircuitBreaker_IsHealthy(t *testing.T) {
	cb := NewCircuitBreaker(&Config{
		Name:             "test",
		FailureThreshold: 1,
		SuccessThreshold: 1,
		Timeout:          100 * time.Millisecond,
		Logger:           zaptest.NewLogger(t),
	})

	t.Run("closed state is healthy", func(t *testing.T) {
		assert.True(t, cb.IsHealthy())
	})

	t.Run("open state is not healthy", func(t *testing.T) {
		cb.ForceOpen()
		assert.False(t, cb.IsHealthy())
	})

	t.Run("half-open state with recent success is healthy", func(t *testing.T) {
		cb.ForceOpen()
		time.Sleep(150 * time.Millisecond)
		cb.Ready() // Transition to half-open

		// Execute successful operation
		cb.Execute(context.Background(), func() error { return nil })
		assert.True(t, cb.IsHealthy())
	})
}

func TestCircuitBreaker_GetName(t *testing.T) {
	cb := NewCircuitBreaker(&Config{
		Name:   "test-circuit-breaker",
		Logger: zaptest.NewLogger(t),
	})

	assert.Equal(t, "test-circuit-breaker", cb.GetName())
}

func TestCircuitBreaker_StartRecovery(t *testing.T) {
	t.Run("auto recovery enabled", func(t *testing.T) {
		cb := NewCircuitBreaker(&Config{
			Name:             "test",
			FailureThreshold: 1,
			SuccessThreshold: 1,
			Timeout:          100 * time.Millisecond,
			AutoRecovery:     true,
			Logger:           zaptest.NewLogger(t),
		})

		ctx := context.Background()
		err := cb.StartRecovery(ctx)
		assert.NoError(t, err)
		assert.True(t, cb.IsRecovering())
	})

	t.Run("auto recovery disabled", func(t *testing.T) {
		cb := NewCircuitBreaker(&Config{
			Name:             "test",
			FailureThreshold: 1,
			SuccessThreshold: 1,
			Timeout:          100 * time.Millisecond,
			AutoRecovery:     false,
			Logger:           zaptest.NewLogger(t),
		})

		ctx := context.Background()
		err := cb.StartRecovery(ctx)
		assert.Error(t, err)
		assert.Equal(t, "auto recovery is disabled", err.Error())
	})

	t.Run("recovery already in progress", func(t *testing.T) {
		cb := NewCircuitBreaker(&Config{
			Name:             "test",
			FailureThreshold: 1,
			SuccessThreshold: 1,
			Timeout:          100 * time.Millisecond,
			AutoRecovery:     true,
			Logger:           zaptest.NewLogger(t),
		})

		ctx := context.Background()
		cb.StartRecovery(ctx)

		err := cb.StartRecovery(ctx)
		assert.Error(t, err)
		assert.Equal(t, "recovery already in progress", err.Error())
	})
}

func TestCircuitBreaker_GetRecoveryProgress(t *testing.T) {
	cb := NewCircuitBreaker(&Config{
		Name:             "test",
		FailureThreshold: 1,
		SuccessThreshold: 1,
		Timeout:          100 * time.Millisecond,
		AutoRecovery:     true,
		Logger:           zaptest.NewLogger(t),
	})

	t.Run("not recovering", func(t *testing.T) {
		progress := cb.GetRecoveryProgress()
		assert.Equal(t, 0.0, progress)
	})

	t.Run("recovering", func(t *testing.T) {
		ctx := context.Background()
		cb.StartRecovery(ctx)

		progress := cb.GetRecoveryProgress()
		assert.GreaterOrEqual(t, progress, 0.0)
		assert.LessOrEqual(t, progress, 1.0)
	})
}

func TestCircuitBreaker_GetMetrics(t *testing.T) {
	cb := NewCircuitBreaker(&Config{
		Name:             "test",
		FailureThreshold: 2,
		SuccessThreshold: 1,
		Timeout:          100 * time.Millisecond,
		Logger:           zaptest.NewLogger(t),
	})

	// Execute some operations
	cb.Execute(context.Background(), func() error { return nil })
	cb.Execute(context.Background(), func() error { return errors.New("error") })

	metrics := cb.GetMetrics()

	assert.Equal(t, "test", metrics["name"])
	assert.Equal(t, "CLOSED", metrics["state"])
	assert.Equal(t, int64(2), metrics["total_requests"])
	assert.Equal(t, int64(1), metrics["total_failures"])
	assert.Equal(t, int64(1), metrics["total_successes"])
	assert.Equal(t, false, metrics["is_recovering"])
}

func TestCircuitBreaker_GetPrometheusMetrics(t *testing.T) {
	cb := NewCircuitBreaker(&Config{
		Name:             "test",
		FailureThreshold: 2,
		SuccessThreshold: 1,
		Timeout:          100 * time.Millisecond,
		Logger:           zaptest.NewLogger(t),
	})

	// Execute some operations
	cb.Execute(context.Background(), func() error { return nil })
	cb.Execute(context.Background(), func() error { return errors.New("error") })

	metrics := cb.GetPrometheusMetrics()

	assert.Equal(t, float64(Closed), metrics["circuit_breaker_state"])
	assert.Equal(t, float64(2), metrics["circuit_breaker_total_requests"])
	assert.Equal(t, float64(1), metrics["circuit_breaker_total_failures"])
	assert.Equal(t, float64(1), metrics["circuit_breaker_total_successes"])
}

func TestCircuitBreaker_StateTransitions(t *testing.T) {
	cb := NewCircuitBreaker(&Config{
		Name:             "test",
		FailureThreshold: 2,
		SuccessThreshold: 1,
		Timeout:          100 * time.Millisecond,
		Logger:           zaptest.NewLogger(t),
	})

	// Start in closed state
	assert.Equal(t, Closed, cb.GetState())

	// Trigger failures to open circuit
	cb.Execute(context.Background(), func() error { return errors.New("error1") })
	assert.Equal(t, Closed, cb.GetState()) // Still closed, need more failures

	cb.Execute(context.Background(), func() error { return errors.New("error2") })
	assert.Equal(t, Open, cb.GetState()) // Now open

	// Wait for timeout to transition to half-open
	time.Sleep(150 * time.Millisecond)
	cb.Ready() // This should transition to half-open
	assert.Equal(t, HalfOpen, cb.GetState())

	// Success should close the circuit
	cb.Execute(context.Background(), func() error { return nil })
	assert.Equal(t, Closed, cb.GetState())
}

func TestCircuitBreaker_ConcurrentAccess(t *testing.T) {
	cb := NewCircuitBreaker(&Config{
		Name:             "test",
		FailureThreshold: 10,
		SuccessThreshold: 5,
		Timeout:          100 * time.Millisecond,
		Logger:           zaptest.NewLogger(t),
	})

	// Run concurrent operations
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()

			cb.Execute(context.Background(), func() error {
				if time.Now().UnixNano()%2 == 0 {
					return nil
				}
				return errors.New("random error")
			})
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify circuit breaker is still functional
	stats := cb.GetStats()
	assert.NotNil(t, stats)
	assert.Equal(t, "test", stats["name"])
}

func TestCircuitBreaker_EdgeCases(t *testing.T) {
	t.Run("zero timeout", func(t *testing.T) {
		cb := NewCircuitBreaker(&Config{
			Name:             "test",
			FailureThreshold: 1,
			SuccessThreshold: 1,
			Timeout:          0,
			Logger:           zaptest.NewLogger(t),
		})

		ctx := context.Background()
		err := cb.Execute(ctx, func() error {
			time.Sleep(10 * time.Millisecond)
			return nil
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "timeout")
	})

	t.Run("very high thresholds", func(t *testing.T) {
		cb := NewCircuitBreaker(&Config{
			Name:             "test",
			FailureThreshold: 1000,
			SuccessThreshold: 1000,
			Timeout:          100 * time.Millisecond,
			Logger:           zaptest.NewLogger(t),
		})

		// Should stay closed with high thresholds
		for i := 0; i < 100; i++ {
			cb.Execute(context.Background(), func() error {
				return errors.New("error")
			})
		}
		assert.Equal(t, Closed, cb.GetState())
	})
}
