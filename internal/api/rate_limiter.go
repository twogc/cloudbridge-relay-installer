package api

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RateLimiterConfig конфигурация для rate limiter
type RateLimiterConfig struct {
	// Общие лимиты
	GlobalRequestsPerMinute int `json:"global_requests_per_minute"`
	GlobalBurstSize         int `json:"global_burst_size"`

	// Лимиты для AI endpoints
	AIRequestsPerMinute int `json:"ai_requests_per_minute"`
	AIBurstSize         int `json:"ai_burst_size"`

	// Лимиты для Data Pipeline endpoints
	PipelineRequestsPerMinute int `json:"pipeline_requests_per_minute"`
	PipelineBurstSize         int `json:"pipeline_burst_size"`

	// Лимиты для Monitoring endpoints
	MonitoringRequestsPerMinute int `json:"monitoring_requests_per_minute"`
	MonitoringBurstSize         int `json:"monitoring_burst_size"`

	// Лимиты по tenant
	TenantRequestsPerMinute int `json:"tenant_requests_per_minute"`
	TenantBurstSize         int `json:"tenant_burst_size"`

	// Лимиты по пользователю
	UserRequestsPerMinute int `json:"user_requests_per_minute"`
	UserBurstSize         int `json:"user_burst_size"`
}

// DefaultRateLimiterConfig возвращает конфигурацию по умолчанию
func DefaultRateLimiterConfig() *RateLimiterConfig {
	return &RateLimiterConfig{
		GlobalRequestsPerMinute:     1000,
		GlobalBurstSize:             100,
		AIRequestsPerMinute:         100,
		AIBurstSize:                 20,
		PipelineRequestsPerMinute:   200,
		PipelineBurstSize:           50,
		MonitoringRequestsPerMinute: 50,
		MonitoringBurstSize:         10,
		TenantRequestsPerMinute:     500,
		TenantBurstSize:             50,
		UserRequestsPerMinute:       100,
		UserBurstSize:               20,
	}
}

// TokenBucket представляет токен bucket для rate limiting
type TokenBucket struct {
	capacity   int
	tokens     int
	lastRefill time.Time
	refillRate int // tokens per minute
	mutex      sync.Mutex
}

// NewTokenBucket создает новый токен bucket
func NewTokenBucket(capacity, refillRate int) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		tokens:     capacity,
		lastRefill: time.Now(),
		refillRate: refillRate,
	}
}

// TryConsume пытается потребить токены
func (tb *TokenBucket) TryConsume(tokens int) bool {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	// Пополняем bucket
	tb.refill()

	if tb.tokens >= tokens {
		tb.tokens -= tokens
		return true
	}
	return false
}

// refill пополняет bucket токенами
func (tb *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)

	// Вычисляем количество токенов для пополнения
	tokensToAdd := int(elapsed.Minutes() * float64(tb.refillRate))

	if tokensToAdd > 0 {
		tb.tokens = min(tb.capacity, tb.tokens+tokensToAdd)
		tb.lastRefill = now
	}
}

// min возвращает минимальное из двух чисел
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// RateLimiter управляет rate limiting для API
type RateLimiter struct {
	config *RateLimiterConfig
	logger *zap.Logger

	// Buckets для разных уровней
	globalBucket     *TokenBucket
	aiBucket         *TokenBucket
	pipelineBucket   *TokenBucket
	monitoringBucket *TokenBucket

	// Buckets по tenant и user
	tenantBuckets map[string]*TokenBucket
	userBuckets   map[string]*TokenBucket
	bucketsMutex  sync.RWMutex
}

// NewRateLimiter создает новый rate limiter
func NewRateLimiter(config *RateLimiterConfig, logger *zap.Logger) *RateLimiter {
	if config == nil {
		config = DefaultRateLimiterConfig()
	}

	return &RateLimiter{
		config:           config,
		logger:           logger,
		globalBucket:     NewTokenBucket(config.GlobalBurstSize, config.GlobalRequestsPerMinute),
		aiBucket:         NewTokenBucket(config.AIBurstSize, config.AIRequestsPerMinute),
		pipelineBucket:   NewTokenBucket(config.PipelineBurstSize, config.PipelineRequestsPerMinute),
		monitoringBucket: NewTokenBucket(config.MonitoringBurstSize, config.MonitoringRequestsPerMinute),
		tenantBuckets:    make(map[string]*TokenBucket),
		userBuckets:      make(map[string]*TokenBucket),
	}
}

// CheckRateLimit проверяет rate limit для запроса
func (rl *RateLimiter) CheckRateLimit(c *gin.Context, endpointType string) bool {
	// Проверяем глобальный лимит
	if !rl.globalBucket.TryConsume(1) {
		rl.logger.Warn("Global rate limit exceeded",
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
			zap.String("remote_addr", c.ClientIP()),
		)
		return false
	}

	// Проверяем лимит для конкретного типа endpoint
	var bucket *TokenBucket
	switch endpointType {
	case "ai":
		bucket = rl.aiBucket
	case "pipeline":
		bucket = rl.pipelineBucket
	case "monitoring":
		bucket = rl.monitoringBucket
	default:
		// Для других endpoints используем глобальный bucket
		return true
	}

	if !bucket.TryConsume(1) {
		rl.logger.Warn("Endpoint rate limit exceeded",
			zap.String("endpoint_type", endpointType),
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
			zap.String("remote_addr", c.ClientIP()),
		)
		return false
	}

	// Проверяем лимит по tenant (если есть claims)
	if claims, exists := GetClaimsFromContext(c); exists {
		tenantBucket := rl.getTenantBucket(claims.TenantID)
		if !tenantBucket.TryConsume(1) {
			rl.logger.Warn("Tenant rate limit exceeded",
				zap.String("tenant_id", claims.TenantID),
				zap.String("endpoint_type", endpointType),
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
				zap.String("remote_addr", c.ClientIP()),
			)
			return false
		}

		// Проверяем лимит по пользователю
		userBucket := rl.getUserBucket(claims.Subject)
		if !userBucket.TryConsume(1) {
			rl.logger.Warn("User rate limit exceeded",
				zap.String("user_id", claims.Subject),
				zap.String("tenant_id", claims.TenantID),
				zap.String("endpoint_type", endpointType),
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
				zap.String("remote_addr", c.ClientIP()),
			)
			return false
		}
	}

	return true
}

// getTenantBucket возвращает bucket для tenant
func (rl *RateLimiter) getTenantBucket(tenantID string) *TokenBucket {
	rl.bucketsMutex.RLock()
	bucket, exists := rl.tenantBuckets[tenantID]
	rl.bucketsMutex.RUnlock()

	if !exists {
		rl.bucketsMutex.Lock()
		// Двойная проверка
		bucket, exists = rl.tenantBuckets[tenantID]
		if !exists {
			bucket = NewTokenBucket(rl.config.TenantBurstSize, rl.config.TenantRequestsPerMinute)
			rl.tenantBuckets[tenantID] = bucket
		}
		rl.bucketsMutex.Unlock()
	}

	return bucket
}

// getUserBucket возвращает bucket для пользователя
func (rl *RateLimiter) getUserBucket(userID string) *TokenBucket {
	rl.bucketsMutex.RLock()
	bucket, exists := rl.userBuckets[userID]
	rl.bucketsMutex.RUnlock()

	if !exists {
		rl.bucketsMutex.Lock()
		// Двойная проверка
		bucket, exists = rl.userBuckets[userID]
		if !exists {
			bucket = NewTokenBucket(rl.config.UserBurstSize, rl.config.UserRequestsPerMinute)
			rl.userBuckets[userID] = bucket
		}
		rl.bucketsMutex.Unlock()
	}

	return bucket
}

// CreateRateLimitMiddleware создает middleware для rate limiting
func (rl *RateLimiter) CreateRateLimitMiddleware(endpointType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !rl.CheckRateLimit(c, endpointType) {
			c.JSON(http.StatusTooManyRequests, ErrorResponse{
				Error:   "Rate limit exceeded",
				Code:    http.StatusTooManyRequests,
				Message: fmt.Sprintf("Rate limit exceeded for %s endpoints", endpointType),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetStats возвращает статистику rate limiter
func (rl *RateLimiter) GetStats() map[string]interface{} {
	rl.bucketsMutex.RLock()
	defer rl.bucketsMutex.RUnlock()

	return map[string]interface{}{
		"global_tokens":     rl.globalBucket.tokens,
		"ai_tokens":         rl.aiBucket.tokens,
		"pipeline_tokens":   rl.pipelineBucket.tokens,
		"monitoring_tokens": rl.monitoringBucket.tokens,
		"tenant_buckets":    len(rl.tenantBuckets),
		"user_buckets":      len(rl.userBuckets),
	}
}

// CleanupOldBuckets очищает старые bucket'ы (вызывается периодически)
func (rl *RateLimiter) CleanupOldBuckets() {
	rl.bucketsMutex.Lock()
	defer rl.bucketsMutex.Unlock()

	// В реальной реализации здесь должна быть логика очистки неиспользуемых bucket'ов
	// Пока что просто логируем количество bucket'ов
	rl.logger.Debug("Rate limiter cleanup",
		zap.Int("tenant_buckets", len(rl.tenantBuckets)),
		zap.Int("user_buckets", len(rl.userBuckets)),
	)
}
