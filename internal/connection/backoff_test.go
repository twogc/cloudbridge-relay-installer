package connection

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBackoffHandler(t *testing.T) {
	tests := []struct {
		name           string
		maxRetries     uint
		baseTime       time.Duration
		retryForever   bool
		expectedRetries int
	}{
		{
			name:           "Default configuration",
			maxRetries:     5,
			baseTime:       time.Second,
			retryForever:   false,
			expectedRetries: 5,
		},
		{
			name:           "Retry forever",
			maxRetries:     3,
			baseTime:       time.Second,
			retryForever:   true,
			expectedRetries: 10, // Test more than maxRetries
		},
		{
			name:           "Zero max retries",
			maxRetries:     0,
			baseTime:       time.Second,
			retryForever:   false,
			expectedRetries: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &RetryConfig{
				MaxRetries:   tt.maxRetries,
				BaseTime:     tt.baseTime,
				MaxDelay:     30 * time.Second,
				Jitter:       0.1,
				RetryForever: tt.retryForever,
			}

			handler := NewBackoffHandler(config)
			require.NotNil(t, handler)

			// Test initial state
			assert.Equal(t, tt.maxRetries, handler.maxRetries)
			assert.Equal(t, tt.baseTime, handler.baseTime)
			assert.Equal(t, tt.retryForever, handler.retryForever)
			assert.Equal(t, 0, handler.Retries())

			// Test retry logic
			retryCount := 0
			for i := 0; i < tt.expectedRetries; i++ {
				if handler.ShouldRetry(&ConnectionError{Message: "test", Retryable: true}, i) {
					retryCount++
					
					// Test backoff duration
					duration := handler.GetBackoffDuration(i)
					assert.Greater(t, duration, time.Duration(0))
					
					// Test that duration increases exponentially (without jitter)
					if i > 0 {
						// Calculate expected duration without jitter
						expectedDuration := config.BaseTime * time.Duration(1<<uint(i))
						if expectedDuration > config.MaxDelay {
							expectedDuration = config.MaxDelay
						}
						// With jitter, duration should be within reasonable bounds
						assert.GreaterOrEqual(t, duration, expectedDuration/2)
						assert.LessOrEqual(t, duration, expectedDuration*2)
					}
				} else {
					break
				}
			}

			if tt.retryForever {
				assert.Greater(t, retryCount, int(tt.maxRetries))
			} else {
				assert.LessOrEqual(t, retryCount, int(tt.maxRetries))
			}
		})
	}
}

func TestBackoffHandler_Backoff(t *testing.T) {
	config := DefaultRetryConfig()
	config.MaxRetries = 3
	config.BaseTime = 10 * time.Millisecond // Short for testing
	config.Logger = nil // Avoid nil pointer dereference

	handler := NewBackoffHandler(config)

	// Test backoff with context
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	backoffCount := 0
	for handler.Backoff(ctx) {
		backoffCount++
		if backoffCount > 5 {
			break // Prevent infinite loop
		}
	}

	assert.Greater(t, backoffCount, 0)
	assert.LessOrEqual(t, backoffCount, 4) // Max 3 retries + 1 initial
}

func TestBackoffHandler_GracePeriod(t *testing.T) {
	config := DefaultRetryConfig()
	config.MaxRetries = 5
	config.BaseTime = time.Millisecond

	handler := NewBackoffHandler(config)

	// Set grace period
	gracePeriod := handler.SetGracePeriod()
	assert.Greater(t, gracePeriod, time.Duration(0))

	// Test that retries are reset after grace period
	handler.retries = 3
	handler.SetGracePeriod()

	// Simulate time passing
	handler.resetDeadline = time.Now().Add(-time.Millisecond)

	// Should reset retries
	handler.BackoffTimer()
	// Note: BackoffTimer increments retries, so we expect 1, not 0
	assert.Equal(t, 1, handler.Retries())
}

func TestBackoffHandler_GetStats(t *testing.T) {
	config := DefaultRetryConfig()
	handler := NewBackoffHandler(config)

	stats := handler.GetStats()
	assert.NotNil(t, stats)

	// Check required fields
	assert.Contains(t, stats, "retries")
	assert.Contains(t, stats, "max_retries")
	assert.Contains(t, stats, "retry_forever")
	assert.Contains(t, stats, "base_time")
	assert.Contains(t, stats, "reached_max")
}

func TestBackoffHandler_ErrorTypes(t *testing.T) {
	config := DefaultRetryConfig()
	handler := NewBackoffHandler(config)

	tests := []struct {
		name     string
		err      error
		retryable bool
	}{
		{
			name:     "ConnectionError retryable",
			err:      &ConnectionError{Message: "test", Retryable: true},
			retryable: true,
		},
		{
			name:     "ConnectionError not retryable",
			err:      &ConnectionError{Message: "test", Retryable: false},
			retryable: false,
		},
		{
			name:     "TimeoutError",
			err:      &TimeoutError{Message: "timeout"},
			retryable: true,
		},
		{
			name:     "NetworkError",
			err:      &NetworkError{Message: "network"},
			retryable: true,
		},
		{
			name:     "Nil error",
			err:      nil,
			retryable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.IsRetryableError(tt.err)
			assert.Equal(t, tt.retryable, result)
		})
	}
} 