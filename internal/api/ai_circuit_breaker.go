package api

import (
	"sync"
	"time"
)

// AICircuitBreaker защищает от сбоев AI сервиса
type AICircuitBreaker struct {
	failureCount    int
	lastFailureTime time.Time
	state           string // "closed", "open", "half-open"
	threshold       int
	timeout         time.Duration
	mu              sync.RWMutex
}

// NewAICircuitBreaker создает новый circuit breaker
func NewAICircuitBreaker() *AICircuitBreaker {
	return &AICircuitBreaker{
		threshold: 5,
		timeout:   30 * time.Second,
		state:     "closed",
	}
}

// CanExecute проверяет, можно ли выполнить запрос
func (cb *AICircuitBreaker) CanExecute() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.state {
	case "closed":
		return true
	case "open":
		// Проверяем, прошло ли достаточно времени
		if time.Since(cb.lastFailureTime) > cb.timeout {
			cb.mu.RUnlock()
			cb.mu.Lock()
			cb.state = "half-open"
			cb.mu.Unlock()
			cb.mu.RLock()
			return true
		}
		return false
	case "half-open":
		return true
	default:
		return false
	}
}

// RecordSuccess записывает успешное выполнение
func (cb *AICircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount = 0
	cb.state = "closed"
}

// RecordFailure записывает неудачное выполнение
func (cb *AICircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount++
	cb.lastFailureTime = time.Now()

	if cb.failureCount >= cb.threshold {
		cb.state = "open"
	}
}

// GetState возвращает текущее состояние circuit breaker
func (cb *AICircuitBreaker) GetState() string {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetFailureCount возвращает количество неудач
func (cb *AICircuitBreaker) GetFailureCount() int {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.failureCount
}

// GetLastFailureTime возвращает время последней неудачи
func (cb *AICircuitBreaker) GetLastFailureTime() time.Time {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.lastFailureTime
}

// GetStats возвращает статистику circuit breaker
func (cb *AICircuitBreaker) GetStats() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return map[string]interface{}{
		"state":             cb.state,
		"failure_count":     cb.failureCount,
		"threshold":         cb.threshold,
		"timeout":           cb.timeout.String(),
		"last_failure_time": cb.lastFailureTime,
	}
}

// Reset сбрасывает circuit breaker
func (cb *AICircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount = 0
	cb.state = "closed"
	cb.lastFailureTime = time.Time{}
}
