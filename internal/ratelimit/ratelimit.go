package ratelimit

import (
	"context"
	"sync"
	"time"
)

// Limiter implements rate limiting
type Limiter struct {
	rate       float64
	burst      int
	tokens     float64
	lastUpdate time.Time
	mu         sync.Mutex
}

// NewLimiter creates a new rate limiter
func NewLimiter(rate float64, burst int) *Limiter {
	return &Limiter{
		rate:       rate,
		burst:      burst,
		tokens:     float64(burst),
		lastUpdate: time.Now(),
	}
}

// Allow checks if a request is allowed
func (l *Limiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(l.lastUpdate).Seconds()
	l.tokens = min(float64(l.burst), l.tokens+elapsed*l.rate)
	l.lastUpdate = now

	if l.tokens >= 1 {
		l.tokens--
		return true
	}
	return false
}

// AllowN checks if n requests are allowed
func (l *Limiter) AllowN(n int) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(l.lastUpdate).Seconds()
	l.tokens = min(float64(l.burst), l.tokens+elapsed*l.rate)
	l.lastUpdate = now

	if l.tokens >= float64(n) {
		l.tokens -= float64(n)
		return true
	}
	return false
}

// Wait blocks until a request is allowed
func (l *Limiter) Wait(ctx context.Context) error {
	for {
		if l.Allow() {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second / time.Duration(l.rate)):
		}
	}
}

// WaitN blocks until n requests are allowed
func (l *Limiter) WaitN(ctx context.Context, n int) error {
	for {
		if l.AllowN(n) {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(float64(n)/l.rate) * time.Second):
		}
	}
}

// IPLimiter implements rate limiting per IP
type IPLimiter struct {
	limiters map[string]*Limiter
	rate     float64
	burst    int
	mu       sync.RWMutex
}

// NewIPLimiter creates a new IP-based rate limiter
func NewIPLimiter(rate float64, burst int) *IPLimiter {
	return &IPLimiter{
		limiters: make(map[string]*Limiter),
		rate:     rate,
		burst:    burst,
	}
}

// Allow checks if a request from an IP is allowed
func (l *IPLimiter) Allow(ip string) bool {
	l.mu.RLock()
	limiter, ok := l.limiters[ip]
	l.mu.RUnlock()

	if !ok {
		l.mu.Lock()
		limiter, ok = l.limiters[ip]
		if !ok {
			limiter = NewLimiter(l.rate, l.burst)
			l.limiters[ip] = limiter
		}
		l.mu.Unlock()
	}

	return limiter.Allow()
}

// Cleanup removes old limiters
func (l *IPLimiter) Cleanup() {
	l.mu.Lock()
	defer l.mu.Unlock()

	// TODO: Implement cleanup logic
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
} 