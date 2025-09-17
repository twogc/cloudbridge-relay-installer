package connection

import (
	"context"
	crand "crypto/rand"
	"math/big"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
	"go.uber.org/zap"
)

const (
	DefaultBaseTime time.Duration = time.Second
)

// Clock interface for time operations (allows testing)
type Clock struct {
	Now   func() time.Time
	After func(d time.Duration) <-chan time.Time
}

// RetryStrategy defines the retry behavior for connections
type RetryStrategy interface {
	ShouldRetry(err error, attempt int) bool
	GetBackoffDuration(attempt int) time.Duration
	Reset()
	GetMaxAttempts() int
	IsRetryableError(err error) bool
}

// BackoffHandler manages exponential backoff and limits the maximum number of retries.
// Adapted from cloudflared retry/backoffhandler.go
type BackoffHandler struct {
	// MaxRetries sets the maximum number of retries to perform
	maxRetries uint
	// RetryForever caps the exponential backoff period but allows indefinite retries
	retryForever bool
	// BaseTime sets the initial backoff period
	baseTime time.Duration
	// MaxDelay sets the maximum delay between retries
	maxDelay time.Duration
	// Jitter factor for randomization (0.0 to 1.0)
	jitter float64

	retries       uint
	resetDeadline time.Time

	Clock  Clock
	logger *logger.Logger
}

// RetryConfig holds configuration for retry behavior
type RetryConfig struct {
	MaxRetries   uint           `json:"max_retries"`
	BaseTime     time.Duration  `json:"base_time"`
	MaxDelay     time.Duration  `json:"max_delay"`
	Jitter       float64        `json:"jitter"`
	RetryForever bool           `json:"retry_forever"`
	Logger       *logger.Logger `json:"-"`
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:   5,
		BaseTime:     DefaultBaseTime,
		MaxDelay:     30 * time.Second,
		Jitter:       0.1,
		RetryForever: false,
	}
}

// NewBackoffHandler creates a new backoff handler
func NewBackoffHandler(config *RetryConfig) *BackoffHandler {
	if config == nil {
		config = DefaultRetryConfig()
	}

	if config.Logger == nil {
		// Logger will be handled by the caller
		config.Logger = nil
	}

	return &BackoffHandler{
		maxRetries:   config.MaxRetries,
		baseTime:     config.BaseTime,
		maxDelay:     config.MaxDelay,
		jitter:       config.Jitter,
		retryForever: config.RetryForever,
		Clock:        Clock{Now: time.Now, After: time.After},
		logger:       config.Logger,
	}
}

// ShouldRetry determines if a retry should be attempted
func (b *BackoffHandler) ShouldRetry(err error, attempt int) bool {
	if !b.IsRetryableError(err) {
		return false
	}

	if b.retryForever {
		return true
	}

	return uint(attempt) < b.maxRetries
}

// GetBackoffDuration calculates the backoff duration for the given attempt
func (b *BackoffHandler) GetBackoffDuration(attempt int) time.Duration {
	if !b.resetDeadline.IsZero() && b.Clock.Now().After(b.resetDeadline) {
		b.retries = 0
		b.resetDeadline = time.Time{}
	}

	if uint(attempt) >= b.maxRetries && !b.retryForever {
		return 0
	}

	// Calculate exponential backoff
	backoff := b.baseTime
	for i := 0; i < attempt; i++ {
		backoff *= 2
		if backoff > b.maxDelay {
			backoff = b.maxDelay
			break
		}
	}

	// Apply jitter
	if b.jitter > 0 {
		jitterRange := float64(backoff) * b.jitter
		// Secure random in [0,1) approximated via big.Int
		// Convert jitterRange to nanoseconds to avoid float reuse
		maxNanos := int64(jitterRange)
		if maxNanos > 0 {
			if n, err := crand.Int(crand.Reader, big.NewInt(maxNanos)); err == nil {
				backoff += time.Duration(n.Int64())
			}
		}
	}

	return backoff
}

// Reset resets the retry counter and backoff state
func (b *BackoffHandler) Reset() {
	b.retries = 0
	b.resetDeadline = time.Time{}
}

// GetMaxAttempts returns the maximum number of retry attempts
func (b *BackoffHandler) GetMaxAttempts() int {
	return int(b.maxRetries)
}

// IsRetryableError determines if an error is retryable
func (b *BackoffHandler) IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check for specific retryable error types
	switch e := err.(type) {
	case *ConnectionError:
		return e.Retryable
	case *TimeoutError:
		return true
	case *NetworkError:
		return true
	default:
		// Default behavior: retry on most errors except context cancellation
		return true
	}
}

// SetGracePeriod sets a grace period within which the backoff timer is maintained
func (b *BackoffHandler) SetGracePeriod() time.Duration {
	maxTimeToWait := b.baseTime * 2 << (b.retries + 1)
	if maxTimeToWait > b.maxDelay {
		maxTimeToWait = b.maxDelay
	}

	// Ensure maxTimeToWait is at least 1 nanosecond to avoid panic
	if maxTimeToWait <= 0 {
		maxTimeToWait = time.Nanosecond
	}
	// Secure random duration in [0, maxTimeToWait)
	var timeToWait time.Duration
	if maxTimeToWait > 0 {
		if n, err := crand.Int(crand.Reader, big.NewInt(maxTimeToWait.Nanoseconds())); err == nil {
			timeToWait = time.Duration(n.Int64())
		}
	}
	b.resetDeadline = b.Clock.Now().Add(timeToWait)

	if b.logger != nil {
		b.logger.Info("Grace period set",
			zap.Duration("grace_period", timeToWait),
			zap.Uint("retries", b.retries))
	}

	return timeToWait
}

// Backoff waits according to exponential backoff
func (b *BackoffHandler) Backoff(ctx context.Context) bool {
	if !b.resetDeadline.IsZero() && b.Clock.Now().After(b.resetDeadline) {
		b.retries = 0
		b.resetDeadline = time.Time{}
	}

	if b.retries >= b.maxRetries && !b.retryForever {
		return false
	}

	b.retries++
	maxTimeToWait := b.GetBackoffDuration(int(b.retries))

	// Ensure maxTimeToWait is at least 1 nanosecond to avoid panic
	if maxTimeToWait <= 0 {
		maxTimeToWait = time.Nanosecond
	}
	var timeToWait time.Duration
	if maxTimeToWait > 0 {
		if n, err := crand.Int(crand.Reader, big.NewInt(maxTimeToWait.Nanoseconds())); err == nil {
			timeToWait = time.Duration(n.Int64())
		}
	}

	if b.logger != nil {
		b.logger.Debug("Backoff delay",
			zap.Duration("delay", timeToWait),
			zap.Uint("retry", b.retries))
	}

	select {
	case <-b.Clock.After(timeToWait):
		return true
	case <-ctx.Done():
		return false
	}
}

// Retries returns the number of retries consumed so far
func (b *BackoffHandler) Retries() int {
	return int(b.retries)
}

// ReachedMaxRetries returns true if maximum retries have been reached
func (b *BackoffHandler) ReachedMaxRetries() bool {
	return b.retries >= b.maxRetries && !b.retryForever
}

// GetMaxBackoffDuration returns the maximum backoff duration without modifying state
func (b *BackoffHandler) GetMaxBackoffDuration(ctx context.Context) (time.Duration, bool) {
	select {
	case <-ctx.Done():
		return time.Duration(0), false
	default:
	}

	if !b.resetDeadline.IsZero() && b.Clock.Now().After(b.resetDeadline) {
		// b.retries would be set to 0 at this point
		return time.Second, true
	}

	if b.retries >= b.maxRetries && !b.retryForever {
		return time.Duration(0), false
	}

	maxTimeToWait := b.GetBackoffDuration(int(b.retries + 1))
	return maxTimeToWait, true
}

// BackoffTimer returns a channel that sends the current time when the exponential backoff timeout expires
// Returns nil if the maximum number of retries have been used
func (b *BackoffHandler) BackoffTimer() <-chan time.Time {
	if !b.resetDeadline.IsZero() && b.Clock.Now().After(b.resetDeadline) {
		b.retries = 0
		b.resetDeadline = time.Time{}
	}

	if b.retries >= b.maxRetries {
		if !b.retryForever {
			return nil
		}
	} else {
		b.retries++
	}

	timeToWait := b.GetBackoffDuration(int(b.retries))
	return b.Clock.After(timeToWait)
}

// GetStats returns backoff statistics
func (b *BackoffHandler) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"retries":            b.retries,
		"max_retries":        b.maxRetries,
		"retry_forever":      b.retryForever,
		"base_time":          b.baseTime.String(),
		"max_delay":          b.maxDelay.String(),
		"jitter":             b.jitter,
		"reached_max":        b.ReachedMaxRetries(),
		"reset_deadline":     b.resetDeadline,
		"has_reset_deadline": !b.resetDeadline.IsZero(),
	}
}

// Custom error types for better error handling
type ConnectionError struct {
	Message   string
	Retryable bool
}

func (e *ConnectionError) Error() string {
	return e.Message
}

type TimeoutError struct {
	Message string
}

func (e *TimeoutError) Error() string {
	return e.Message
}

type NetworkError struct {
	Message string
}

func (e *NetworkError) Error() string {
	return e.Message
}
