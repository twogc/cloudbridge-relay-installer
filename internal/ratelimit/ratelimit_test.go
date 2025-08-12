package ratelimit

import (
	"context"
	"testing"
	"time"
)

func TestLimiter(t *testing.T) {
	limiter := NewLimiter(10, 5)

	// Test burst
	for i := 0; i < 5; i++ {
		if !limiter.Allow() {
			t.Errorf("Expected burst of 5 to be allowed")
		}
	}

	// Test rate limiting
	if limiter.Allow() {
		t.Errorf("Expected rate limit to be hit after burst")
	}

	// Test rate recovery
	time.Sleep(100 * time.Millisecond)
	if !limiter.Allow() {
		t.Errorf("Expected rate to recover after waiting")
	}
}

func TestLimiterN(t *testing.T) {
	limiter := NewLimiter(10, 5)

	// Test burst
	if !limiter.AllowN(5) {
		t.Errorf("Expected burst of 5 to be allowed")
	}

	// Test rate limiting
	if limiter.AllowN(1) {
		t.Errorf("Expected rate limit to be hit after burst")
	}

	// Test rate recovery
	time.Sleep(100 * time.Millisecond)
	if !limiter.AllowN(1) {
		t.Errorf("Expected rate to recover after waiting")
	}
}

func TestLimiterWait(t *testing.T) {
	limiter := NewLimiter(10, 5)
	ctx := context.Background()

	// Test burst
	for i := 0; i < 5; i++ {
		if err := limiter.Wait(ctx); err != nil {
			t.Errorf("Expected burst of 5 to be allowed")
		}
	}

	// Test rate limiting
	done := make(chan struct{})
	go func() {
		if err := limiter.Wait(ctx); err != nil {
			t.Errorf("Expected rate to recover after waiting")
		}
		close(done)
	}()

	select {
	case <-done:
		t.Errorf("Expected rate limit to be hit after burst")
	case <-time.After(50 * time.Millisecond):
		// Expected
	}

	// Test rate recovery
	time.Sleep(100 * time.Millisecond)
	select {
	case <-done:
		// Expected
	case <-time.After(50 * time.Millisecond):
		t.Errorf("Expected rate to recover after waiting")
	}
}

func TestIPLimiter(t *testing.T) {
	limiter := NewIPLimiter(10, 5)

	// Test different IPs
	ips := []string{"1.1.1.1", "2.2.2.2", "3.3.3.3"}
	for _, ip := range ips {
		// Test burst
		for i := 0; i < 5; i++ {
			if !limiter.Allow(ip) {
				t.Errorf("Expected burst of 5 to be allowed for IP %s", ip)
			}
		}

		// Test rate limiting
		if limiter.Allow(ip) {
			t.Errorf("Expected rate limit to be hit after burst for IP %s", ip)
		}

		// Test rate recovery
		time.Sleep(100 * time.Millisecond)
		if !limiter.Allow(ip) {
			t.Errorf("Expected rate to recover after waiting for IP %s", ip)
		}
	}
}

func TestIPLimiterConcurrent(t *testing.T) {
	limiter := NewIPLimiter(10, 5)
	done := make(chan struct{})

	// Test concurrent access
	for i := 0; i < 10; i++ {
		go func() {
			ip := "1.1.1.1"
			for j := 0; j < 10; j++ {
				limiter.Allow(ip)
			}
			done <- struct{}{}
		}()
	}

	// Wait for all goroutines to finish
	for i := 0; i < 10; i++ {
		select {
		case <-done:
			// Expected
		case <-time.After(time.Second):
			t.Errorf("Expected goroutines to finish within 1 second")
		}
	}
}
