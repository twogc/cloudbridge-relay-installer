package network

import (
	"context"
	"errors"
	"sync"
	"time"
)

// CircuitBreakerState состояние circuit breaker
type CircuitBreakerState int

const (
	StateClosed CircuitBreakerState = iota
	StateOpen
	StateHalfOpen
)

// CircuitBreaker реализует circuit breaker паттерн
type CircuitBreaker struct {
	maxFailures      int
	timeout          time.Duration
	halfOpenMaxCalls int
	state            CircuitBreakerState
	failures         int
	halfOpenCalls    int
	lastFailTime     time.Time
	mutex            sync.RWMutex
}

// ErrCircuitOpen ошибка когда circuit breaker открыт
var ErrCircuitOpen = errors.New("circuit breaker is open")

// NewCircuitBreaker создает новый circuit breaker
func NewCircuitBreaker(maxFailures int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:      maxFailures,
		timeout:          timeout,
		halfOpenMaxCalls: 3, // Максимум 3 вызова в half-open состоянии
		state:            StateClosed,
	}
}

// Execute выполняет операцию через circuit breaker
func (cb *CircuitBreaker) Execute(ctx context.Context, operation func() error) error {
	// Быстрая проверка состояния
	cb.mutex.RLock()
	state := cb.state
	lastFailTime := cb.lastFailTime
	cb.mutex.RUnlock()

	// Если circuit открыт и не прошло достаточно времени
	if state == StateOpen && time.Since(lastFailTime) < cb.timeout {
		return ErrCircuitOpen
	}

	// Если в half-open состоянии, проверяем лимит вызовов
	if state == StateHalfOpen {
		cb.mutex.Lock()
		if cb.halfOpenCalls >= cb.halfOpenMaxCalls {
			cb.mutex.Unlock()
			return ErrCircuitOpen
		}
		cb.halfOpenCalls++
		cb.mutex.Unlock()
	}

	// Выполняем операцию
	err := operation()

	// Обновляем состояние
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	if err != nil {
		cb.failures++
		cb.lastFailTime = time.Now()

		// Переходим в открытое состояние при достижении лимита ошибок
		if cb.failures >= cb.maxFailures {
			cb.state = StateOpen
		} else if cb.state == StateHalfOpen {
			// В half-open состоянии любая ошибка переводит в open
			cb.state = StateOpen
		}
		return err
	}

	// Успешный вызов - сбрасываем счетчики и переходим в closed
	cb.failures = 0
	cb.halfOpenCalls = 0
	cb.state = StateClosed
	return nil
}

// IsOpen проверяет, открыт ли circuit breaker
func (cb *CircuitBreaker) IsOpen() bool {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	return cb.state == StateOpen && time.Since(cb.lastFailTime) < cb.timeout
}

// OnSuccess сообщает circuit breaker об успешной операции
func (cb *CircuitBreaker) OnSuccess() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.failures = 0
	cb.halfOpenCalls = 0
	cb.state = StateClosed
}

// OnFailure сообщает circuit breaker о неудачной операции
func (cb *CircuitBreaker) OnFailure() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.failures++
	cb.lastFailTime = time.Now()

	// Переходим в открытое состояние при достижении лимита ошибок
	if cb.failures >= cb.maxFailures {
		cb.state = StateOpen
	} else if cb.state == StateHalfOpen {
		// В half-open состоянии любая ошибка переводит в open
		cb.state = StateOpen
	}
}

// GetState возвращает текущее состояние circuit breaker
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state
}

// GetStats возвращает статистику circuit breaker
func (cb *CircuitBreaker) GetStats() map[string]interface{} {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	return map[string]interface{}{
		"state":           cb.state,
		"failures":        cb.failures,
		"half_open_calls": cb.halfOpenCalls,
		"last_fail_time":  cb.lastFailTime,
	}
}
