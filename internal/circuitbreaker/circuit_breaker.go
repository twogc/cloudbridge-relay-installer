package circuitbreaker

import (
	"context"
	"errors"
	"sync"
	"time"

	"go.uber.org/zap"
)

// State represents the circuit breaker state
type State int

const (
	Closed State = iota
	Open
	HalfOpen
)

func (s State) String() string {
	switch s {
	case Closed:
		return "CLOSED"
	case Open:
		return "OPEN"
	case HalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	name           string
	state          State
	failureCount   int64
	successCount   int64
	lastFailure    time.Time
	lastSuccess    time.Time
	
	// Configuration
	failureThreshold int64
	successThreshold int64
	timeout          time.Duration
	windowSize       time.Duration
	autoRecovery     bool
	
	// Metrics
	totalRequests    int64
	totalFailures    int64
	totalSuccesses   int64
	totalTimeouts    int64
	stateChanges     int64
	recoveryAttempts int64
	
	// Auto recovery
	recoveryStartTime time.Time
	isRecovering      bool
	
	mu               sync.RWMutex
	logger           *zap.Logger
}

// Config holds circuit breaker configuration
type Config struct {
	Name              string        `json:"name"`
	FailureThreshold  int64         `json:"failure_threshold"`
	SuccessThreshold  int64         `json:"success_threshold"`
	Timeout           time.Duration `json:"timeout"`
	WindowSize        time.Duration `json:"window_size"`
	AutoRecovery      bool          `json:"auto_recovery"`
	Logger            *zap.Logger   `json:"-"`
}

// DefaultConfig returns default circuit breaker configuration
func DefaultConfig(name string) *Config {
	return &Config{
		Name:             name,
		FailureThreshold: 5,
		SuccessThreshold: 3,
		Timeout:          30 * time.Second,
		WindowSize:       60 * time.Second,
		AutoRecovery:     true,
	}
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config *Config) *CircuitBreaker {
	if config == nil {
		config = DefaultConfig("default")
	}
	
	if config.Logger == nil {
		config.Logger = zap.NewNop()
	}
	
	return &CircuitBreaker{
		name:             config.Name,
		state:            Closed,
		failureThreshold: config.FailureThreshold,
		successThreshold: config.SuccessThreshold,
		timeout:          config.Timeout,
		windowSize:       config.WindowSize,
		autoRecovery:     config.AutoRecovery,
		logger:           config.Logger,
	}
}

// Execute runs a function with circuit breaker protection
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() error) error {
	if !cb.Ready() {
		return errors.New("circuit breaker is open")
	}
	
	cb.mu.Lock()
	cb.totalRequests++
	cb.mu.Unlock()
	
	// Execute with timeout
	done := make(chan error, 1)
	go func() {
		done <- fn()
	}()
	
	select {
	case err := <-done:
		cb.recordResult(err)
		return err
	case <-ctx.Done():
		cb.recordTimeout()
		return ctx.Err()
	case <-time.After(cb.timeout):
		cb.recordTimeout()
		return errors.New("circuit breaker timeout")
	}
}

// ExecuteWithResult runs a function that returns a result with circuit breaker protection
func ExecuteWithResult[T any](cb *CircuitBreaker, ctx context.Context, fn func() (T, error)) (T, error) {
	var zero T
	
	if !cb.Ready() {
		return zero, errors.New("circuit breaker is open")
	}
	
	cb.mu.Lock()
	cb.totalRequests++
	cb.mu.Unlock()
	
	// Execute with timeout
	done := make(chan struct {
		result T
		err    error
	}, 1)
	
	go func() {
		result, err := fn()
		done <- struct {
			result T
			err    error
		}{result, err}
	}()
	
	select {
	case res := <-done:
		cb.recordResult(res.err)
		return res.result, res.err
	case <-ctx.Done():
		cb.recordTimeout()
		return zero, ctx.Err()
	case <-time.After(cb.timeout):
		cb.recordTimeout()
		return zero, errors.New("circuit breaker timeout")
	}
}

// Ready checks if the circuit breaker is ready to execute
func (cb *CircuitBreaker) Ready() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	
	switch cb.state {
	case Closed:
		return true
	case Open:
		// Check if timeout has passed
		if time.Since(cb.lastFailure) > cb.timeout {
			cb.mu.RUnlock()
			cb.mu.Lock()
			cb.state = HalfOpen
			cb.logger.Info("Circuit breaker transitioning to half-open",
				zap.String("name", cb.name))
			cb.mu.Unlock()
			cb.mu.RLock()
			return true
		}
		return false
	case HalfOpen:
		return true
	default:
		return false
	}
}

// recordResult records the result of an execution
func (cb *CircuitBreaker) recordResult(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	if err != nil {
		cb.failureCount++
		cb.totalFailures++
		cb.lastFailure = time.Now()
		
		cb.logger.Debug("Circuit breaker recorded failure",
			zap.String("name", cb.name),
			zap.Error(err),
			zap.Int64("failure_count", cb.failureCount))
		
		// Check if we should open the circuit
		if cb.state == Closed && cb.failureCount >= cb.failureThreshold {
			cb.state = Open
			cb.logger.Warn("Circuit breaker opened",
				zap.String("name", cb.name),
				zap.Int64("failure_count", cb.failureCount),
				zap.Duration("timeout", cb.timeout))
		} else if cb.state == HalfOpen {
			// If half-open and we get a failure, go back to open
			cb.state = Open
			cb.logger.Warn("Circuit breaker reopened after failure",
				zap.String("name", cb.name))
		}
	} else {
		cb.successCount++
		cb.totalSuccesses++
		cb.lastSuccess = time.Now()
		
		cb.logger.Debug("Circuit breaker recorded success",
			zap.String("name", cb.name),
			zap.Int64("success_count", cb.successCount))
		
		// Reset failure count on success
		cb.failureCount = 0
		
		// If half-open and we get enough successes, close the circuit
		if cb.state == HalfOpen && cb.successCount >= cb.successThreshold {
			cb.state = Closed
			cb.successCount = 0
			cb.logger.Info("Circuit breaker closed after successful recovery",
				zap.String("name", cb.name))
		}
	}
}

// recordTimeout records a timeout
func (cb *CircuitBreaker) recordTimeout() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	cb.totalTimeouts++
	cb.failureCount++
	cb.totalFailures++
	cb.lastFailure = time.Now()
	
	cb.logger.Debug("Circuit breaker recorded timeout",
		zap.String("name", cb.name),
		zap.Int64("timeout_count", cb.totalTimeouts))
	
	// Check if we should open the circuit
	if cb.state == Closed && cb.failureCount >= cb.failureThreshold {
		cb.state = Open
		cb.logger.Warn("Circuit breaker opened due to timeouts",
			zap.String("name", cb.name),
			zap.Int64("failure_count", cb.failureCount))
	} else if cb.state == HalfOpen {
		cb.state = Open
		cb.logger.Warn("Circuit breaker reopened after timeout",
			zap.String("name", cb.name))
	}
}

// GetState returns the current state
func (cb *CircuitBreaker) GetState() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetStats returns circuit breaker statistics
func (cb *CircuitBreaker) GetStats() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	
	return map[string]interface{}{
		"name":              cb.name,
		"state":             cb.state.String(),
		"failure_count":     cb.failureCount,
		"success_count":     cb.successCount,
		"total_requests":    cb.totalRequests,
		"total_failures":    cb.totalFailures,
		"total_successes":   cb.totalSuccesses,
		"total_timeouts":    cb.totalTimeouts,
		"last_failure":      cb.lastFailure,
		"last_success":      cb.lastSuccess,
		"failure_threshold": cb.failureThreshold,
		"success_threshold": cb.successThreshold,
		"timeout":           cb.timeout.String(),
		"window_size":       cb.windowSize.String(),
		"success_rate":      cb.calculateSuccessRate(),
	}
}

// calculateSuccessRate calculates the success rate
func (cb *CircuitBreaker) calculateSuccessRate() float64 {
	if cb.totalRequests == 0 {
		return 0.0
	}
	return float64(cb.totalSuccesses) / float64(cb.totalRequests) * 100
}

// ForceOpen forces the circuit breaker to open state
func (cb *CircuitBreaker) ForceOpen() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	oldState := cb.state
	cb.state = Open
	cb.lastFailure = time.Now()
	
	cb.logger.Info("Circuit breaker forced open",
		zap.String("name", cb.name),
		zap.String("old_state", oldState.String()))
}

// ForceClose forces the circuit breaker to closed state
func (cb *CircuitBreaker) ForceClose() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	oldState := cb.state
	cb.state = Closed
	cb.failureCount = 0
	cb.successCount = 0
	
	cb.logger.Info("Circuit breaker forced closed",
		zap.String("name", cb.name),
		zap.String("old_state", oldState.String()))
}

// Reset resets the circuit breaker to initial state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	cb.state = Closed
	cb.failureCount = 0
	cb.successCount = 0
	cb.totalRequests = 0
	cb.totalFailures = 0
	cb.totalSuccesses = 0
	cb.totalTimeouts = 0
	cb.lastFailure = time.Time{}
	cb.lastSuccess = time.Time{}
	
	cb.logger.Info("Circuit breaker reset",
		zap.String("name", cb.name))
}

// IsHealthy returns true if the circuit breaker is healthy
func (cb *CircuitBreaker) IsHealthy() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	
	// Consider healthy if closed or half-open with recent success
	if cb.state == Closed {
		return true
	}
	
	if cb.state == HalfOpen {
		// Consider healthy if we have recent success
		return cb.successCount > 0 && time.Since(cb.lastSuccess) < cb.windowSize
	}
	
	return false
}

// GetName returns the circuit breaker name
func (cb *CircuitBreaker) GetName() string {
	return cb.name
}

// StartRecovery starts automatic recovery process
func (cb *CircuitBreaker) StartRecovery(ctx context.Context) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	if !cb.autoRecovery {
		return errors.New("auto recovery is disabled")
	}
	
	if cb.isRecovering {
		return errors.New("recovery already in progress")
	}
	
	cb.isRecovering = true
	cb.recoveryStartTime = time.Now()
	cb.recoveryAttempts++
	
	cb.logger.Info("Starting automatic recovery",
		zap.String("name", cb.name),
		zap.Int64("recovery_attempts", cb.recoveryAttempts))
	
	// Start recovery goroutine
	go cb.recoveryLoop(ctx)
	
	return nil
}

// IsRecovering returns true if recovery is in progress
func (cb *CircuitBreaker) IsRecovering() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.isRecovering
}

// GetRecoveryProgress returns recovery progress as a percentage
func (cb *CircuitBreaker) GetRecoveryProgress() float64 {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	
	if !cb.isRecovering {
		return 0.0
	}
	
	elapsed := time.Since(cb.recoveryStartTime)
	progress := float64(elapsed) / float64(cb.timeout)
	
	if progress > 1.0 {
		return 1.0
	}
	
	return progress
}

// recoveryLoop handles automatic recovery
func (cb *CircuitBreaker) recoveryLoop(ctx context.Context) {
	ticker := time.NewTicker(cb.timeout / 10) // Check every 10% of timeout
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			cb.mu.Lock()
			cb.isRecovering = false
			cb.mu.Unlock()
			return
		case <-ticker.C:
			cb.mu.Lock()
			if cb.state == Open && time.Since(cb.lastFailure) > cb.timeout {
				cb.state = HalfOpen
				cb.stateChanges++
				cb.isRecovering = false
				cb.logger.Info("Circuit breaker recovered to half-open state",
					zap.String("name", cb.name))
			}
			cb.mu.Unlock()
		}
	}
}

// GetMetrics returns circuit breaker metrics
func (cb *CircuitBreaker) GetMetrics() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	
	successRate := cb.calculateSuccessRate()
	
	return map[string]interface{}{
		"name":               cb.name,
		"state":              cb.state.String(),
		"total_requests":     cb.totalRequests,
		"total_failures":     cb.totalFailures,
		"total_successes":    cb.totalSuccesses,
		"total_timeouts":     cb.totalTimeouts,
		"state_changes":      cb.stateChanges,
		"recovery_attempts":  cb.recoveryAttempts,
		"success_rate":       successRate,
		"is_recovering":      cb.isRecovering,
		"recovery_progress":  cb.GetRecoveryProgress(),
		"failure_count":      cb.failureCount,
		"success_count":      cb.successCount,
		"last_failure":       cb.lastFailure,
		"last_success":       cb.lastSuccess,
	}
}

// GetPrometheusMetrics returns metrics in Prometheus format
func (cb *CircuitBreaker) GetPrometheusMetrics() map[string]float64 {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return map[string]float64{
		"circuit_breaker_state":           float64(cb.state),
		"circuit_breaker_failure_count":   float64(cb.failureCount),
		"circuit_breaker_success_count":   float64(cb.successCount),
		"circuit_breaker_total_requests":  float64(cb.totalRequests),
		"circuit_breaker_total_failures":  float64(cb.totalFailures),
		"circuit_breaker_total_successes": float64(cb.totalSuccesses),
		"circuit_breaker_total_timeouts":  float64(cb.totalTimeouts),
		"circuit_breaker_state_changes":   float64(cb.stateChanges),
		"circuit_breaker_success_rate":    cb.calculateSuccessRate(),
		"circuit_breaker_recovery_progress": cb.GetRecoveryProgress(),
	}
} 