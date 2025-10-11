package network

import (
	"math"
	"math/rand"
	"time"
)

// RetryConfig конфигурация для retry политик
type RetryConfig struct {
	MaxAttempts     int           `yaml:"max_attempts"`
	InitialDelay    time.Duration `yaml:"initial_delay"`
	MaxDelay        time.Duration `yaml:"max_delay"`
	BackoffFactor   float64       `yaml:"backoff_factor"`
	Jitter          bool          `yaml:"jitter"`
	RetryableErrors []int         `yaml:"retryable_errors"`
}

// DefaultRetryConfig возвращает конфигурацию по умолчанию
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:     3,
		InitialDelay:    time.Second,
		MaxDelay:        30 * time.Second,
		BackoffFactor:   2.0,
		Jitter:          true,
		RetryableErrors: []int{500, 502, 503, 504, 429},
	}
}

// IsRetryableError проверяет, можно ли повторить запрос при данной ошибке
func (rc *RetryConfig) IsRetryableError(statusCode int) bool {
	for _, code := range rc.RetryableErrors {
		if statusCode == code {
			return true
		}
	}
	return false
}

// CalculateDelay вычисляет задержку для попытки
func (rc *RetryConfig) CalculateDelay(attempt int) time.Duration {
	base := float64(rc.InitialDelay) * math.Pow(rc.BackoffFactor, float64(attempt-1))

	if base > float64(rc.MaxDelay) {
		base = float64(rc.MaxDelay)
	}

	if rc.Jitter {
		// Добавляем ±10% джиттер для избежания thundering herd
		jitter := base * (rand.Float64()*0.2 - 0.1)
		base += jitter
	}

	return time.Duration(base)
}
