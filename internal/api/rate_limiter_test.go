package api

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewRateLimiter(t *testing.T) {
	config := DefaultRateLimiterConfig()
	logger := zap.NewNop()

	rateLimiter := NewRateLimiter(config, logger)

	assert.NotNil(t, rateLimiter)
}

func TestDefaultRateLimiterConfig(t *testing.T) {
	config := DefaultRateLimiterConfig()

	assert.NotNil(t, config)
	assert.Greater(t, config.GlobalRequestsPerMinute, 0)
	assert.Greater(t, config.GlobalBurstSize, 0)
}

func TestRateLimiter_CreateRateLimitMiddleware(t *testing.T) {
	config := &RateLimiterConfig{
		GlobalRequestsPerMinute: 10,
		GlobalBurstSize:         20,
		AIRequestsPerMinute:     5,
		AIBurstSize:             10,
	}
	logger := zap.NewNop()

	rateLimiter := NewRateLimiter(config, logger)
	middleware := rateLimiter.CreateRateLimitMiddleware("test")

	assert.NotNil(t, middleware)

	// Test middleware function
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	middleware(c)

	// Should not abort under normal conditions
	assert.False(t, c.IsAborted())
}

func TestRateLimiter_CheckRateLimit(t *testing.T) {
	config := &RateLimiterConfig{
		GlobalRequestsPerMinute: 10,
		GlobalBurstSize:         20,
		AIRequestsPerMinute:     5,
		AIBurstSize:             10,
	}
	logger := zap.NewNop()

	rateLimiter := NewRateLimiter(config, logger)

	// Test that requests are initially allowed
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	allowed := rateLimiter.CheckRateLimit(c, "ai")
	assert.True(t, allowed)
}

func TestRateLimiter_GetStats(t *testing.T) {
	config := DefaultRateLimiterConfig()
	logger := zap.NewNop()

	rateLimiter := NewRateLimiter(config, logger)
	stats := rateLimiter.GetStats()

	assert.NotNil(t, stats)
	assert.Contains(t, stats, "global_tokens")
	assert.Contains(t, stats, "ai_tokens")
	assert.Contains(t, stats, "tenant_buckets")
	assert.Contains(t, stats, "user_buckets")
}

func TestRateLimiter_CleanupOldBuckets(t *testing.T) {
	config := DefaultRateLimiterConfig()
	logger := zap.NewNop()

	rateLimiter := NewRateLimiter(config, logger)

	// Should not panic
	assert.NotPanics(t, func() {
		rateLimiter.CleanupOldBuckets()
	})
}

func TestNewTokenBucket(t *testing.T) {
	bucket := NewTokenBucket(10, 5)

	assert.NotNil(t, bucket)
	assert.Equal(t, 10, bucket.capacity)
	assert.Equal(t, 10, bucket.tokens) // Initially full
	assert.Equal(t, 5, bucket.refillRate)
	assert.False(t, bucket.lastRefill.IsZero())
}

func TestTokenBucket_TryConsume(t *testing.T) {
	bucket := NewTokenBucket(10, 5)

	tests := []struct {
		name     string
		tokens   int
		expected bool
	}{
		{
			name:     "consume available tokens",
			tokens:   5,
			expected: true,
		},
		{
			name:     "consume all tokens",
			tokens:   10,
			expected: true,
		},
		{
			name:     "consume more than available",
			tokens:   15,
			expected: false,
		},
		{
			name:     "consume zero tokens",
			tokens:   0,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset bucket for each test
			bucket.tokens = 10
			result := bucket.TryConsume(tt.tokens)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTokenBucket_Refill(t *testing.T) {
	bucket := NewTokenBucket(10, 5)

	// Consume all tokens
	bucket.tokens = 0
	bucket.lastRefill = time.Now().Add(-2 * time.Minute) // 2 minutes ago

	// Try to consume 1 token - should trigger refill
	result := bucket.TryConsume(1)
	assert.True(t, result)
	assert.Equal(t, 9, bucket.tokens) // Should have 9 tokens (10 - 1 consumed)
}

func TestTokenBucket_RefillRate(t *testing.T) {
	bucket := NewTokenBucket(10, 5)

	// Consume all tokens
	bucket.tokens = 0
	bucket.lastRefill = time.Now().Add(-1 * time.Minute) // 1 minute ago

	// Try to consume 1 token - should trigger refill
	result := bucket.TryConsume(1)
	assert.True(t, result)
	assert.Equal(t, 4, bucket.tokens) // Should have 4 tokens (5 refilled - 1 consumed)
}

func TestTokenBucket_CapacityLimit(t *testing.T) {
	bucket := NewTokenBucket(5, 10) // High refill rate, low capacity

	// Consume all tokens
	bucket.tokens = 0
	bucket.lastRefill = time.Now().Add(-1 * time.Minute) // 1 minute ago

	// Try to consume 1 token - should trigger refill
	result := bucket.TryConsume(1)
	assert.True(t, result)
	assert.Equal(t, 4, bucket.tokens) // Should be capped at capacity (5 - 1 consumed)
}

func TestRateLimiter_CheckRateLimit_GlobalLimit(t *testing.T) {
	config := &RateLimiterConfig{
		GlobalRequestsPerMinute: 2,
		GlobalBurstSize:         2,
	}
	logger := zap.NewNop()
	rateLimiter := NewRateLimiter(config, logger)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	// First two requests should be allowed
	assert.True(t, rateLimiter.CheckRateLimit(c, "test"))
	assert.True(t, rateLimiter.CheckRateLimit(c, "test"))

	// Third request should be denied (burst limit exceeded)
	assert.False(t, rateLimiter.CheckRateLimit(c, "test"))
}

func TestRateLimiter_CheckRateLimit_EndpointSpecific(t *testing.T) {
	config := &RateLimiterConfig{
		GlobalRequestsPerMinute: 100,
		GlobalBurstSize:         100,
		AIRequestsPerMinute:     2,
		AIBurstSize:             2,
	}
	logger := zap.NewNop()
	rateLimiter := NewRateLimiter(config, logger)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/ai/test", nil)

	// First two AI requests should be allowed
	assert.True(t, rateLimiter.CheckRateLimit(c, "ai"))
	assert.True(t, rateLimiter.CheckRateLimit(c, "ai"))

	// Third AI request should be denied
	assert.False(t, rateLimiter.CheckRateLimit(c, "ai"))

	// But other endpoint types should still work
	assert.True(t, rateLimiter.CheckRateLimit(c, "pipeline"))
}

func TestRateLimiter_CheckRateLimit_WithClaims(t *testing.T) {
	config := &RateLimiterConfig{
		GlobalRequestsPerMinute: 100,
		GlobalBurstSize:         100,
		TenantRequestsPerMinute: 2,
		TenantBurstSize:         2,
		UserRequestsPerMinute:   1,
		UserBurstSize:           1,
	}
	logger := zap.NewNop()
	rateLimiter := NewRateLimiter(config, logger)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	// Mock claims in context
	claims := &P2PClaims{
		Subject:  "user1",
		TenantID: "tenant1",
	}
	c.Set("claims", claims)

	// First request should be allowed
	assert.True(t, rateLimiter.CheckRateLimit(c, "test"))

	// Second request should be denied (user limit exceeded)
	assert.False(t, rateLimiter.CheckRateLimit(c, "test"))
}

func TestRateLimiter_GetTenantBucket(t *testing.T) {
	config := DefaultRateLimiterConfig()
	logger := zap.NewNop()
	rateLimiter := NewRateLimiter(config, logger)

	// Get bucket for tenant1
	bucket1 := rateLimiter.getTenantBucket("tenant1")
	assert.NotNil(t, bucket1)

	// Get bucket for tenant2
	bucket2 := rateLimiter.getTenantBucket("tenant2")
	assert.NotNil(t, bucket2)

	// Should be different buckets
	assert.NotEqual(t, bucket1, bucket2)

	// Getting same tenant should return same bucket
	bucket1Again := rateLimiter.getTenantBucket("tenant1")
	assert.Equal(t, bucket1, bucket1Again)
}

func TestRateLimiter_GetUserBucket(t *testing.T) {
	config := DefaultRateLimiterConfig()
	logger := zap.NewNop()
	rateLimiter := NewRateLimiter(config, logger)

	// Get bucket for user1
	bucket1 := rateLimiter.getUserBucket("user1")
	assert.NotNil(t, bucket1)

	// Get bucket for user2
	bucket2 := rateLimiter.getUserBucket("user2")
	assert.NotNil(t, bucket2)

	// Should be different buckets
	assert.NotEqual(t, bucket1, bucket2)

	// Getting same user should return same bucket
	bucket1Again := rateLimiter.getUserBucket("user1")
	assert.Equal(t, bucket1, bucket1Again)
}

func TestRateLimiter_CreateRateLimitMiddleware_Detailed(t *testing.T) {
	config := &RateLimiterConfig{
		GlobalRequestsPerMinute: 1,
		GlobalBurstSize:         1,
	}
	logger := zap.NewNop()
	rateLimiter := NewRateLimiter(config, logger)

	middleware := rateLimiter.CreateRateLimitMiddleware("test")
	assert.NotNil(t, middleware)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	// First request should pass
	middleware(c)
	assert.False(t, c.IsAborted())

	// Second request should be rate limited
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Request = httptest.NewRequest("GET", "/test", nil)
	middleware(c2)
	assert.True(t, c2.IsAborted())
	assert.Equal(t, 429, w2.Code)
}

func TestRateLimiter_GetStats_Detailed(t *testing.T) {
	config := DefaultRateLimiterConfig()
	logger := zap.NewNop()
	rateLimiter := NewRateLimiter(config, logger)

	stats := rateLimiter.GetStats()
	assert.NotNil(t, stats)
	assert.Contains(t, stats, "global_tokens")
	assert.Contains(t, stats, "ai_tokens")
	assert.Contains(t, stats, "pipeline_tokens")
	assert.Contains(t, stats, "monitoring_tokens")
	assert.Contains(t, stats, "tenant_buckets")
	assert.Contains(t, stats, "user_buckets")

	// Initially should have 0 tenant and user buckets
	assert.Equal(t, 0, stats["tenant_buckets"])
	assert.Equal(t, 0, stats["user_buckets"])
}

func TestRateLimiter_ConcurrentAccess(t *testing.T) {
	config := &RateLimiterConfig{
		GlobalRequestsPerMinute: 100,
		GlobalBurstSize:         100,
	}
	logger := zap.NewNop()
	rateLimiter := NewRateLimiter(config, logger)

	gin.SetMode(gin.TestMode)

	// Test concurrent access
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test", nil)

			// This should not panic
			rateLimiter.CheckRateLimit(c, "test")
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestMinFunction(t *testing.T) {
	tests := []struct {
		name     string
		a        int
		b        int
		expected int
	}{
		{
			name:     "a is smaller",
			a:        5,
			b:        10,
			expected: 5,
		},
		{
			name:     "b is smaller",
			a:        10,
			b:        5,
			expected: 5,
		},
		{
			name:     "equal values",
			a:        5,
			b:        5,
			expected: 5,
		},
		{
			name:     "negative values",
			a:        -5,
			b:        -10,
			expected: -10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := min(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRateLimiter_EdgeCases(t *testing.T) {
	// Test with nil config
	logger := zap.NewNop()
	rateLimiter := NewRateLimiter(nil, logger)
	assert.NotNil(t, rateLimiter)

	// Test with zero values
	config := &RateLimiterConfig{
		GlobalRequestsPerMinute: 0,
		GlobalBurstSize:         0,
	}
	rateLimiter2 := NewRateLimiter(config, logger)
	assert.NotNil(t, rateLimiter2)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	// Should not panic even with zero config
	assert.False(t, rateLimiter2.CheckRateLimit(c, "test"))
}

func TestTokenBucket_EdgeCases(t *testing.T) {
	// Test with zero capacity
	bucket := NewTokenBucket(0, 5)
	assert.Equal(t, 0, bucket.capacity)
	assert.Equal(t, 0, bucket.tokens)

	// Try to consume from empty bucket
	result := bucket.TryConsume(1)
	assert.False(t, result)

	// Test with zero refill rate
	bucket2 := NewTokenBucket(10, 0)
	bucket2.tokens = 0
	bucket2.lastRefill = time.Now().Add(-1 * time.Minute)

	// Should not refill
	result = bucket2.TryConsume(1)
	assert.False(t, result)
}
