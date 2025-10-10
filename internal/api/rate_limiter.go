package api

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RateLimiterConfig конфигурация для rate limiter с QoS
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

	// Лимиты по tenant с QoS
	TenantRequestsPerMinute int `json:"tenant_requests_per_minute"`
	TenantBurstSize         int `json:"tenant_burst_size"`
	
	// QoS настройки
	QoSEnabled              bool    `json:"qos_enabled"`
	TenantBandwidthLimit    int64   `json:"tenant_bandwidth_limit"`    // bytes per second
	TenantConnectionLimit   int     `json:"tenant_connection_limit"`
	TenantPriorityWeight    float64 `json:"tenant_priority_weight"`     // 0.0-1.0
	
	// Circuit breaker для tenant
	TenantCircuitBreaker struct {
		Enabled          bool    `json:"enabled"`
		FailureThreshold int     `json:"failure_threshold"`
		SuccessThreshold int     `json:"success_threshold"`
		Timeout          int     `json:"timeout"` // seconds
	} `json:"tenant_circuit_breaker"`

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
		
		// QoS настройки
		QoSEnabled:              true,
		TenantBandwidthLimit:    10 * 1024 * 1024, // 10 MB/s
		TenantConnectionLimit:   100,
		TenantPriorityWeight:   0.5,
		
		// Circuit breaker
		TenantCircuitBreaker: struct {
			Enabled          bool `json:"enabled"`
			FailureThreshold int  `json:"failure_threshold"`
			SuccessThreshold int  `json:"success_threshold"`
			Timeout          int  `json:"timeout"`
		}{
			Enabled:          true,
			FailureThreshold: 5,
			SuccessThreshold: 3,
			Timeout:          30,
		},
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

// RateLimiter управляет rate limiting для API с QoS
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
	
	// QoS компоненты
	qosManager      *QoSManager
	circuitBreaker *CircuitBreakerManager
}

// QoSManager управляет QoS для tenant'ов
type QoSManager struct {
	config *RateLimiterConfig
	logger *zap.Logger
	
	// Tenant QoS состояние
	tenantQoS    map[string]*TenantQoS
	tenantMutex  sync.RWMutex
	
	// Bandwidth tracking
	bandwidthUsage map[string]*BandwidthTracker
	bandwidthMutex sync.RWMutex
}

// TenantQoS представляет QoS состояние tenant'а
type TenantQoS struct {
	TenantID          string    `json:"tenant_id"`
	Priority          float64   `json:"priority"`           // 0.0-1.0
	BandwidthLimit    int64     `json:"bandwidth_limit"`    // bytes/s
	ConnectionLimit   int       `json:"connection_limit"`
	CurrentConnections int      `json:"current_connections"`
	BandwidthUsage    int64     `json:"bandwidth_usage"`    // bytes/s
	LastUpdate        time.Time `json:"last_update"`
	
	// QoS метрики
	LatencyP99        float64   `json:"latency_p99"`        // ms
	ErrorRate         float64   `json:"error_rate"`         // %
	Throughput        float64   `json:"throughput"`          // req/s
	
	// Circuit breaker состояние
	CircuitState      string    `json:"circuit_state"`      // closed, open, half-open
	FailureCount      int       `json:"failure_count"`
	SuccessCount      int       `json:"success_count"`
	LastFailure       time.Time `json:"last_failure"`
}

// BandwidthTracker отслеживает использование bandwidth
type BandwidthTracker struct {
	TenantID     string    `json:"tenant_id"`
	BytesPerSec  int64     `json:"bytes_per_sec"`
	TotalBytes   int64     `json:"total_bytes"`
	LastReset    time.Time `json:"last_reset"`
	WindowSize   time.Duration `json:"window_size"`
}

// CircuitBreakerManager управляет circuit breaker'ами
type CircuitBreakerManager struct {
	config *RateLimiterConfig
	logger *zap.Logger
	
	// Circuit breaker состояние
	breakers    map[string]*CircuitBreaker
	breakersMux sync.RWMutex
}

// CircuitBreaker представляет circuit breaker
type CircuitBreaker struct {
	TenantID         string    `json:"tenant_id"`
	State            string    `json:"state"`            // closed, open, half-open
	FailureCount     int       `json:"failure_count"`
	SuccessCount     int       `json:"success_count"`
	LastFailure      time.Time `json:"last_failure"`
	LastSuccess      time.Time `json:"last_success"`
	NextAttempt      time.Time `json:"next_attempt"`
}

// NewRateLimiter создает новый rate limiter с QoS
func NewRateLimiter(config *RateLimiterConfig, logger *zap.Logger) *RateLimiter {
	if config == nil {
		config = DefaultRateLimiterConfig()
	}

	// Создаем QoS manager
	qosManager := &QoSManager{
		config:        config,
		logger:        logger,
		tenantQoS:    make(map[string]*TenantQoS),
		bandwidthUsage: make(map[string]*BandwidthTracker),
	}
	
	// Создаем Circuit breaker manager
	circuitBreaker := &CircuitBreakerManager{
		config:   config,
		logger:   logger,
		breakers: make(map[string]*CircuitBreaker),
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
		qosManager:       qosManager,
		circuitBreaker:   circuitBreaker,
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

// CheckQoSLimit проверяет QoS лимиты для tenant'а
func (rl *RateLimiter) CheckQoSLimit(tenantID string, requestSize int64) (bool, string) {
	if !rl.config.QoSEnabled {
		return true, "qos_disabled"
	}
	
	// Получаем QoS состояние tenant'а
	tenantQoS := rl.qosManager.GetTenantQoS(tenantID)
	if tenantQoS == nil {
		// Создаем новое QoS состояние
		tenantQoS = rl.qosManager.CreateTenantQoS(tenantID)
	}
	
	// Проверяем circuit breaker
	if rl.circuitBreaker.IsOpen(tenantID) {
		return false, "circuit_breaker_open"
	}
	
	// Проверяем bandwidth лимит
	if !rl.qosManager.CheckBandwidthLimit(tenantID, requestSize) {
		return false, "bandwidth_limit_exceeded"
	}
	
	// Проверяем connection лимит
	if !rl.qosManager.CheckConnectionLimit(tenantID) {
		return false, "connection_limit_exceeded"
	}
	
	// Обновляем метрики
	rl.qosManager.UpdateTenantMetrics(tenantID, requestSize)
	
	return true, "qos_ok"
}

// UpdateTenantQoS обновляет QoS для tenant'а
func (rl *RateLimiter) UpdateTenantQoS(tenantID string, priority float64, bandwidthLimit int64, connectionLimit int) {
	rl.qosManager.UpdateTenantQoS(tenantID, priority, bandwidthLimit, connectionLimit)
}

// GetTenantQoSStats возвращает QoS статистику для tenant'а
func (rl *RateLimiter) GetTenantQoSStats(tenantID string) map[string]interface{} {
	return rl.qosManager.GetTenantStats(tenantID)
}

// QoSManager методы

// GetTenantQoS возвращает QoS состояние tenant'а
func (qm *QoSManager) GetTenantQoS(tenantID string) *TenantQoS {
	qm.tenantMutex.RLock()
	defer qm.tenantMutex.RUnlock()
	
	return qm.tenantQoS[tenantID]
}

// CreateTenantQoS создает новое QoS состояние для tenant'а
func (qm *QoSManager) CreateTenantQoS(tenantID string) *TenantQoS {
	qm.tenantMutex.Lock()
	defer qm.tenantMutex.Unlock()
	
	tenantQoS := &TenantQoS{
		TenantID:          tenantID,
		Priority:          qm.config.TenantPriorityWeight,
		BandwidthLimit:    qm.config.TenantBandwidthLimit,
		ConnectionLimit:   qm.config.TenantConnectionLimit,
		CurrentConnections: 0,
		BandwidthUsage:    0,
		LastUpdate:        time.Now(),
		CircuitState:      "closed",
	}
	
	qm.tenantQoS[tenantID] = tenantQoS
	return tenantQoS
}

// CheckBandwidthLimit проверяет bandwidth лимит
func (qm *QoSManager) CheckBandwidthLimit(tenantID string, requestSize int64) bool {
	qm.tenantMutex.RLock()
	tenantQoS, exists := qm.tenantQoS[tenantID]
	qm.tenantMutex.RUnlock()
	
	if !exists {
		return true // Если tenant не найден, разрешаем
	}
	
	// Проверяем текущее использование bandwidth
	return tenantQoS.BandwidthUsage+requestSize <= tenantQoS.BandwidthLimit
}

// CheckConnectionLimit проверяет connection лимит
func (qm *QoSManager) CheckConnectionLimit(tenantID string) bool {
	qm.tenantMutex.RLock()
	tenantQoS, exists := qm.tenantQoS[tenantID]
	qm.tenantMutex.RUnlock()
	
	if !exists {
		return true // Если tenant не найден, разрешаем
	}
	
	// Проверяем текущее количество соединений
	return tenantQoS.CurrentConnections < tenantQoS.ConnectionLimit
}

// UpdateTenantMetrics обновляет метрики tenant'а
func (qm *QoSManager) UpdateTenantMetrics(tenantID string, requestSize int64) {
	qm.tenantMutex.Lock()
	defer qm.tenantMutex.Unlock()
	
	tenantQoS, exists := qm.tenantQoS[tenantID]
	if !exists {
		return
	}
	
	// Обновляем bandwidth usage
	tenantQoS.BandwidthUsage += requestSize
	tenantQoS.LastUpdate = time.Now()
	
	// Увеличиваем количество соединений
	tenantQoS.CurrentConnections++
}

// UpdateTenantQoS обновляет QoS настройки для tenant'а
func (qm *QoSManager) UpdateTenantQoS(tenantID string, priority float64, bandwidthLimit int64, connectionLimit int) {
	qm.tenantMutex.Lock()
	defer qm.tenantMutex.Unlock()
	
	tenantQoS, exists := qm.tenantQoS[tenantID]
	if !exists {
		tenantQoS = &TenantQoS{
			TenantID: tenantID,
		}
		qm.tenantQoS[tenantID] = tenantQoS
	}
	
	tenantQoS.Priority = priority
	tenantQoS.BandwidthLimit = bandwidthLimit
	tenantQoS.ConnectionLimit = connectionLimit
	tenantQoS.LastUpdate = time.Now()
}

// GetTenantStats возвращает статистику tenant'а
func (qm *QoSManager) GetTenantStats(tenantID string) map[string]interface{} {
	qm.tenantMutex.RLock()
	defer qm.tenantMutex.RUnlock()
	
	tenantQoS, exists := qm.tenantQoS[tenantID]
	if !exists {
		return map[string]interface{}{
			"error": "tenant_not_found",
		}
	}
	
	return map[string]interface{}{
		"tenant_id":           tenantQoS.TenantID,
		"priority":            tenantQoS.Priority,
		"bandwidth_limit":     tenantQoS.BandwidthLimit,
		"bandwidth_usage":     tenantQoS.BandwidthUsage,
		"connection_limit":    tenantQoS.ConnectionLimit,
		"current_connections": tenantQoS.CurrentConnections,
		"circuit_state":       tenantQoS.CircuitState,
		"latency_p99":         tenantQoS.LatencyP99,
		"error_rate":          tenantQoS.ErrorRate,
		"throughput":          tenantQoS.Throughput,
		"last_update":         tenantQoS.LastUpdate,
	}
}

// CircuitBreakerManager методы

// IsOpen проверяет, открыт ли circuit breaker для tenant'а
func (cbm *CircuitBreakerManager) IsOpen(tenantID string) bool {
	cbm.breakersMux.RLock()
	defer cbm.breakersMux.RUnlock()
	
	breaker, exists := cbm.breakers[tenantID]
	if !exists {
		return false // Если breaker не существует, считаем закрытым
	}
	
	return breaker.State == "open"
}

// RecordSuccess записывает успешный запрос
func (cbm *CircuitBreakerManager) RecordSuccess(tenantID string) {
	cbm.breakersMux.Lock()
	defer cbm.breakersMux.Unlock()
	
	breaker, exists := cbm.breakers[tenantID]
	if !exists {
		breaker = &CircuitBreaker{
			TenantID: tenantID,
			State:    "closed",
		}
		cbm.breakers[tenantID] = breaker
	}
	
	breaker.SuccessCount++
	breaker.LastSuccess = time.Now()
	
	// Если breaker в half-open состоянии и достаточно успехов, закрываем его
	if breaker.State == "half-open" && breaker.SuccessCount >= cbm.config.TenantCircuitBreaker.SuccessThreshold {
		breaker.State = "closed"
		breaker.FailureCount = 0
		breaker.SuccessCount = 0
	}
}

// RecordFailure записывает неудачный запрос
func (cbm *CircuitBreakerManager) RecordFailure(tenantID string) {
	cbm.breakersMux.Lock()
	defer cbm.breakersMux.Unlock()
	
	breaker, exists := cbm.breakers[tenantID]
	if !exists {
		breaker = &CircuitBreaker{
			TenantID: tenantID,
			State:    "closed",
		}
		cbm.breakers[tenantID] = breaker
	}
	
	breaker.FailureCount++
	breaker.LastFailure = time.Now()
	
	// Если достаточно неудач, открываем breaker
	if breaker.FailureCount >= cbm.config.TenantCircuitBreaker.FailureThreshold {
		breaker.State = "open"
		breaker.NextAttempt = time.Now().Add(time.Duration(cbm.config.TenantCircuitBreaker.Timeout) * time.Second)
	}
}
