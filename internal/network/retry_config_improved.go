package network

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"net"
	"net/url"
	"sync"
	"time"
)

// RetryConfigImproved — конфигурация retry/backoff с безопасным джиттером
type RetryConfigImproved struct {
	MaxAttempts     int           `yaml:"max_attempts"`
	InitialDelay    time.Duration `yaml:"initial_delay"`
	MaxDelay        time.Duration `yaml:"max_delay"`
	MinDelay        time.Duration `yaml:"min_delay"`        // минимальная задержка
	BackoffFactor   float64       `yaml:"backoff_factor"`
	Jitter          bool          `yaml:"jitter"`
	JitterRatio     float64       `yaml:"jitter_ratio"`      // доля джиттера, например 0.1 = ±10%
	RetryableErrors []int         `yaml:"retryable_errors"`

	// внутренний генератор джиттера
	rnd   *rand.Rand
	rndMu sync.Mutex
}

// DefaultRetryConfigImproved — разумные дефолты
func DefaultRetryConfigImproved() *RetryConfigImproved {
	rc := &RetryConfigImproved{
		MaxAttempts:     3,
		InitialDelay:    time.Second,
		MaxDelay:        30 * time.Second,
		MinDelay:        250 * time.Millisecond,
		BackoffFactor:   2.0,
		Jitter:          true,
		JitterRatio:     0.10,
		RetryableErrors: []int{500, 502, 503, 504, 429},
	}
	rc.initRand()
	return rc
}

// initRand безопасно инициализирует источник случайности
func (rc *RetryConfigImproved) initRand() {
	if rc.rnd == nil {
		// seeding временем достаточно для джиттера (не крипто)
		rc.rnd = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
}

// Normalize приводит конфиг к безопасным значениям
func (rc *RetryConfigImproved) Normalize() {
	if rc.MaxAttempts <= 0 {
		rc.MaxAttempts = 1
	}
	if rc.InitialDelay <= 0 {
		rc.InitialDelay = time.Second
	}
	if rc.MinDelay < 0 {
		rc.MinDelay = 0
	}
	if rc.MaxDelay <= 0 {
		rc.MaxDelay = 30 * time.Second
	}
	if rc.MaxDelay < rc.InitialDelay {
		rc.MaxDelay = rc.InitialDelay
	}
	if rc.BackoffFactor < 1.0 || math.IsNaN(rc.BackoffFactor) || math.IsInf(rc.BackoffFactor, 0) {
		rc.BackoffFactor = 2.0
	}
	if rc.JitterRatio < 0 || rc.JitterRatio > 1 {
		rc.JitterRatio = 0.10
	}
	rc.initRand()
}

// IsRetryableError — по HTTP-коду
func (rc *RetryConfigImproved) IsRetryableError(statusCode int) bool {
	for _, code := range rc.RetryableErrors {
		if statusCode == code {
			return true
		}
	}
	return false
}

// IsRetryableNetError — по ошибке сети/контекста
func (rc *RetryConfigImproved) IsRetryableNetError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		// Canceled обычно не ретраим, но оставьте на усмотрение вызывающего
		return errors.Is(err, context.DeadlineExceeded)
	}
	var nerr net.Error
	if errors.As(err, &nerr) {
		return nerr.Timeout() || nerr.Temporary()
	}
	// DNS/URL и т.п. — неретраибл
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		// В url.Error может лежать net.Error — его уже обработали
		// Остальные — обычно конфиг-ошибки => false
	}
	return false
}

// CalculateDelay — экспоненциальный backoff с джиттером
func (rc *RetryConfigImproved) CalculateDelay(attempt int) time.Duration {
	rc.Normalize()
	if attempt < 1 {
		attempt = 1
	}
	base := float64(rc.InitialDelay) * math.Pow(rc.BackoffFactor, float64(attempt-1))
	if base > float64(rc.MaxDelay) {
		base = float64(rc.MaxDelay)
	}
	if rc.Jitter {
		// ±JitterRatio: base * (1 ± r*JitterRatio)
		rc.rndMu.Lock()
		r := rc.rnd.Float64()
		rc.rndMu.Unlock()
		// r в [0,1) -> сместим к [-0.5, +0.5) и умножим на (2*JitterRatio)
		j := (r - 0.5) * 2 * rc.JitterRatio
		base = base * (1 + j)
	}
	delay := time.Duration(base)
	if delay < rc.MinDelay {
		delay = rc.MinDelay
	}
	return delay
}

// CalculateDelayWithRetryAfter — учитывает заголовок Retry-After (секунды или HTTP-дата)
func (rc *RetryConfigImproved) CalculateDelayWithRetryAfter(attempt int, retryAfterHeader string) time.Duration {
	// Если сервер явно попросил подождать — слушаемся
	if retryAfterHeader != "" {
		// формат: число секунд
		if secs, err := time.ParseDuration(retryAfterHeader + "s"); err == nil && secs > 0 {
			return clamp(secs, rc.MinDelay, rc.MaxDelay)
		}
		// формат: HTTP-date
		if t, err := httpTimeParse(retryAfterHeader); err == nil {
			wait := time.Until(t)
			if wait > 0 {
				return clamp(wait, rc.MinDelay, rc.MaxDelay)
			}
		}
	}
	return rc.CalculateDelay(attempt)
}

// clamp ограничивает duration в диапазоне [min,max]
func clamp(d, min, max time.Duration) time.Duration {
	if d < min {
		return min
	}
	if d > max {
		return max
	}
	return d
}

// httpTimeParse — разбирает несколько форматов HTTP-даты
// (минималистично: RFC1123)
func httpTimeParse(s string) (time.Time, error) {
	// Mon, 02 Jan 2006 15:04:05 GMT
	const rfc1123 = time.RFC1123
	if t, err := time.Parse(rfc1123, s); err == nil {
		return t, nil
	}
	// RFC1123Z — на всякий случай
	if t, err := time.Parse(time.RFC1123Z, s); err == nil {
		return t, nil
	}
	return time.Time{}, errors.New("unsupported http time format")
}
