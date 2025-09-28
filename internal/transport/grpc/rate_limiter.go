package grpc

import (
	"context"
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"go.uber.org/zap"
)

// TenantRateLimiter provides per-tenant rate limiting for gRPC
type TenantRateLimiter struct {
	tenantLimits map[string]*TokenBucket // tenant_id -> rate limiter
	mutex        sync.RWMutex
	logger       *zap.Logger
	
	// Default limits
	defaultRequestsPerMinute int
	defaultBurstSize        int
}

// TokenBucket implements token bucket rate limiting algorithm
type TokenBucket struct {
	capacity     int           // Maximum tokens
	tokens       int           // Current tokens
	refillRate   time.Duration // Refill interval
	lastRefill   time.Time     // Last refill time
	mutex        sync.Mutex
}

// NewTenantRateLimiter creates new per-tenant rate limiter
func NewTenantRateLimiter(logger *zap.Logger) *TenantRateLimiter {
	return &TenantRateLimiter{
		tenantLimits:            make(map[string]*TokenBucket),
		logger:                  logger,
		defaultRequestsPerMinute: 100, // 100 requests per minute per tenant
		defaultBurstSize:        20,   // burst of 20 requests
	}
}

// NewTokenBucket creates new token bucket
func NewTokenBucket(requestsPerMinute, burstSize int) *TokenBucket {
	return &TokenBucket{
		capacity:   burstSize,
		tokens:     burstSize, // Start with full bucket
		refillRate: time.Minute / time.Duration(requestsPerMinute),
		lastRefill: time.Now(),
	}
}

// TryConsume attempts to consume a token, returns true if successful
func (tb *TokenBucket) TryConsume() bool {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()
	
	// Refill tokens based on time passed
	now := time.Now()
	timePassed := now.Sub(tb.lastRefill)
	tokensToAdd := int(timePassed / tb.refillRate)
	
	if tokensToAdd > 0 {
		tb.tokens += tokensToAdd
		if tb.tokens > tb.capacity {
			tb.tokens = tb.capacity
		}
		tb.lastRefill = now
	}
	
	// Try to consume token
	if tb.tokens > 0 {
		tb.tokens--
		return true
	}
	
	return false
}

// CheckRateLimit checks if request should be allowed for tenant
func (rl *TenantRateLimiter) CheckRateLimit(tenantID string) error {
	if tenantID == "" {
		tenantID = "anonymous"
	}
	
	rl.mutex.RLock()
	bucket, exists := rl.tenantLimits[tenantID]
	rl.mutex.RUnlock()
	
	if !exists {
		// Create new bucket for tenant
		rl.mutex.Lock()
		bucket = NewTokenBucket(rl.defaultRequestsPerMinute, rl.defaultBurstSize)
		rl.tenantLimits[tenantID] = bucket
		rl.mutex.Unlock()
		
		rl.logger.Info("Created rate limiter for new tenant",
			zap.String("tenant_id", tenantID),
			zap.Int("requests_per_minute", rl.defaultRequestsPerMinute),
			zap.Int("burst_size", rl.defaultBurstSize),
		)
	}
	
	if !bucket.TryConsume() {
		return fmt.Errorf("rate limit exceeded for tenant %s", tenantID)
	}
	
	return nil
}

// SetTenantLimit sets custom rate limit for specific tenant
func (rl *TenantRateLimiter) SetTenantLimit(tenantID string, requestsPerMinute, burstSize int) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	
	rl.tenantLimits[tenantID] = NewTokenBucket(requestsPerMinute, burstSize)
	
	rl.logger.Info("Set custom rate limit for tenant",
		zap.String("tenant_id", tenantID),
		zap.Int("requests_per_minute", requestsPerMinute),
		zap.Int("burst_size", burstSize),
	)
}

// GetTenantStats returns rate limiting statistics for tenant
func (rl *TenantRateLimiter) GetTenantStats(tenantID string) map[string]interface{} {
	rl.mutex.RLock()
	defer rl.mutex.RUnlock()
	
	bucket, exists := rl.tenantLimits[tenantID]
	if !exists {
		return map[string]interface{}{
			"tenant_id": tenantID,
			"status":    "not_configured",
		}
	}
	
	bucket.mutex.Lock()
	defer bucket.mutex.Unlock()
	
	return map[string]interface{}{
		"tenant_id":       tenantID,
		"current_tokens":  bucket.tokens,
		"capacity":        bucket.capacity,
		"refill_rate_ms":  bucket.refillRate.Milliseconds(),
		"last_refill":     bucket.lastRefill,
	}
}

// GetAllStats returns rate limiting statistics for all tenants
func (rl *TenantRateLimiter) GetAllStats() map[string]interface{} {
	rl.mutex.RLock()
	defer rl.mutex.RUnlock()
	
	stats := make(map[string]interface{})
	stats["total_tenants"] = len(rl.tenantLimits)
	stats["default_rpm"] = rl.defaultRequestsPerMinute
	stats["default_burst"] = rl.defaultBurstSize
	
	tenantStats := make(map[string]interface{})
	for tenantID := range rl.tenantLimits {
		tenantStats[tenantID] = rl.GetTenantStats(tenantID)
	}
	stats["tenants"] = tenantStats
	
	return stats
}

// CleanupInactiveTenants removes rate limiters for inactive tenants
func (rl *TenantRateLimiter) CleanupInactiveTenants(inactiveDuration time.Duration) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	
	cutoff := time.Now().Add(-inactiveDuration)
	removedCount := 0
	
	for tenantID, bucket := range rl.tenantLimits {
		bucket.mutex.Lock()
		lastActivity := bucket.lastRefill
		bucket.mutex.Unlock()
		
		if lastActivity.Before(cutoff) {
			delete(rl.tenantLimits, tenantID)
			removedCount++
		}
	}
	
	if removedCount > 0 {
		rl.logger.Info("Cleaned up inactive tenant rate limiters",
			zap.Int("removed_count", removedCount),
			zap.Duration("inactive_duration", inactiveDuration),
		)
	}
}

// RateLimitingInterceptor creates a gRPC interceptor for rate limiting
func (rl *TenantRateLimiter) RateLimitingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Extract tenant ID from context (should be set by auth middleware)
		tenantID, _ := ctx.Value("tenant_id").(string)
		if tenantID == "" {
			tenantID = "anonymous"
		}
		
		// Check rate limit
		if err := rl.CheckRateLimit(tenantID); err != nil {
			rl.logger.Warn("Rate limit exceeded",
				zap.String("tenant_id", tenantID),
				zap.String("method", info.FullMethod),
				zap.Error(err),
			)
			return nil, status.Errorf(codes.ResourceExhausted, "rate limit exceeded")
		}
		
		return handler(ctx, req)
	}
}

