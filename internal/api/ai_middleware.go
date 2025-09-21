package api

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai/anomaly"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai/predictive"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai/routing"
	"go.uber.org/zap"
)

// AIMiddleware обеспечивает интеграцию AI с Relay Server
type AIMiddleware struct {
	aiClient       ai.AIClientInterface
	config         *ai.Config
	logger         *zap.Logger
	cache          *AICache
	metrics        *AIMetrics
	circuitBreaker *AICircuitBreaker
}

// NewAIMiddleware создает новый AI middleware
func NewAIMiddleware(aiClient ai.AIClientInterface, config *ai.Config, logger *zap.Logger) *AIMiddleware {
	return &AIMiddleware{
		aiClient:       aiClient,
		config:         config,
		logger:         logger,
		cache:          NewAICache(),
		metrics:        NewAIMetrics(),
		circuitBreaker: NewAICircuitBreaker(),
	}
}

// Connection представляет соединение для AI обработки
type Connection struct {
	ID               string
	RemoteAddr       string
	Source           string
	Destination      string
	Protocol         string
	Route            string
	Timestamp        time.Time
	PacketsPerSecond int
	BytesPerSecond   int64
	Latency          time.Duration
	ConnectionCount  int
	ThrottleFactor   float64
}

// ProcessConnection обрабатывает соединение через AI
func (m *AIMiddleware) ProcessConnection(ctx context.Context, conn *Connection) error {
	start := time.Now()
	defer func() {
		m.metrics.RecordProcessingTime(time.Since(start))
		m.metrics.IncrementRequestsProcessed()
	}()

	// Проверяем circuit breaker
	if !m.circuitBreaker.CanExecute() {
		m.logger.Warn("AI service circuit breaker is open, skipping AI processing")
		return nil
	}

	// 1. Анализ аномалий
	anomalyResult, err := m.analyzeAnomalies(ctx, conn)
	if err != nil {
		m.logger.Error("Failed to analyze anomalies", zap.Error(err))
		m.circuitBreaker.RecordFailure()
		return err
	}

	// 2. Предсказание нагрузки
	loadPrediction, err := m.predictLoad(ctx, conn)
	if err != nil {
		m.logger.Error("Failed to predict load", zap.Error(err))
		m.circuitBreaker.RecordFailure()
		return err
	}

	// 3. Оптимизация маршрута
	optimizedRoute, err := m.optimizeRoute(ctx, conn)
	if err != nil {
		m.logger.Error("Failed to optimize route", zap.Error(err))
		m.circuitBreaker.RecordFailure()
		return err
	}

	// 4. Применение результатов
	if err := m.applyAIResults(ctx, conn, anomalyResult, loadPrediction, optimizedRoute); err != nil {
		m.logger.Error("Failed to apply AI results", zap.Error(err))
		return err
	}

	// Записываем успех в circuit breaker
	m.circuitBreaker.RecordSuccess()

	return nil
}

// analyzeAnomalies анализирует аномалии в соединении
func (m *AIMiddleware) analyzeAnomalies(ctx context.Context, conn *Connection) (*anomaly.AnomalyResult, error) {
	// Проверяем кэш
	cacheKey := fmt.Sprintf("anomaly:%s:%d", conn.RemoteAddr, time.Now().Unix()/300)

	if result, exists := m.cache.GetAnomalyResult(cacheKey); exists {
		m.metrics.IncrementCacheHits()
		return result, nil
	}

	m.metrics.IncrementCacheMisses()

	// Собираем метрики соединения
	metrics := &anomaly.NetworkMetrics{
		Timestamp:            time.Now(),
		ConnectionsPerSecond: conn.PacketsPerSecond,
		BytesPerSecond:       conn.BytesPerSecond,
		PacketLossRate:       0.01,
		AverageLatency:       float64(conn.Latency.Milliseconds()),
		CPUUsage:             50.0,
		MemoryUsage:          45.0,
		ActiveConnections:    conn.ConnectionCount,
		FailedConnections:    0,
		ProtocolDistribution: map[string]int{
			conn.Protocol: 1,
		},
	}

	// Вызываем AI Service
	result, err := m.aiClient.DetectAnomaly(ctx, metrics)
	if err != nil {
		return nil, err
	}

	// Кэшируем результат
	m.cache.SetAnomalyResult(cacheKey, result)

	// Обновляем метрики
	if result.IsAnomaly {
		m.metrics.IncrementAnomaliesDetected()
	}

	return result, nil
}

// predictLoad предсказывает нагрузку
func (m *AIMiddleware) predictLoad(ctx context.Context, conn *Connection) (*predictive.LoadPrediction, error) {
	// Проверяем кэш
	cacheKey := fmt.Sprintf("load:%d", time.Now().Unix()/300)

	if result, exists := m.cache.GetLoadPrediction(cacheKey); exists {
		m.metrics.IncrementCacheHits()
		return result, nil
	}

	m.metrics.IncrementCacheMisses()

	// Вызываем AI Service
	result, err := m.aiClient.PredictLoad(ctx, 24) // 24 часа
	if err != nil {
		return nil, err
	}

	// Кэшируем результат
	m.cache.SetLoadPrediction(cacheKey, result)
	m.metrics.IncrementPredictionsMade()

	return result, nil
}

// optimizeRoute оптимизирует маршрут
func (m *AIMiddleware) optimizeRoute(ctx context.Context, conn *Connection) (*routing.OptimizedRoute, error) {
	// Проверяем кэш
	cacheKey := fmt.Sprintf("route:%s:%s", conn.Source, conn.Destination)

	if result, exists := m.cache.GetOptimizedRoute(cacheKey); exists {
		m.metrics.IncrementCacheHits()
		return result, nil
	}

	m.metrics.IncrementCacheMisses()

	// Вызываем AI Service
	constraints := map[string]interface{}{
		"max_latency":   100,  // ms
		"min_bandwidth": 1000, // Mbps
	}

	result, err := m.aiClient.OptimizeRoute(ctx, conn.Source, conn.Destination, constraints)
	if err != nil {
		return nil, err
	}

	// Кэшируем результат
	m.cache.SetOptimizedRoute(cacheKey, result)
	m.metrics.IncrementRoutesOptimized()

	return result, nil
}

// applyAIResults применяет результаты AI анализа
func (m *AIMiddleware) applyAIResults(ctx context.Context, conn *Connection,
	anomalyResult *anomaly.AnomalyResult, loadPrediction *predictive.LoadPrediction,
	optimizedRoute *routing.OptimizedRoute) error {

	// 1. Обработка аномалий
	if anomalyResult.IsAnomaly {
		m.logger.Warn("Anomaly detected in connection",
			zap.String("remote_addr", conn.RemoteAddr),
			zap.Float64("score", anomalyResult.Score),
			zap.String("type", anomalyResult.AnomalyType),
		)

		// Применяем политику реагирования
		switch anomalyResult.Severity {
		case "high", "critical":
			return m.blockConnection(ctx, conn, "anomaly_detected")
		case "medium":
			return m.throttleConnection(ctx, conn, 0.5)
		case "low":
			m.logger.Info("Low severity anomaly, continuing")
		}
	}

	// 2. Применение предсказания нагрузки
	if loadPrediction != nil {
		m.adjustResources(ctx, loadPrediction)
	}

	// 3. Применение оптимизированного маршрута
	if optimizedRoute != nil {
		conn.Route = fmt.Sprintf("%v", optimizedRoute.Path)
		conn.Protocol = optimizedRoute.Protocol
	}

	return nil
}

// blockConnection блокирует соединение
func (m *AIMiddleware) blockConnection(ctx context.Context, conn *Connection, reason string) error {
	m.logger.Warn("Blocking connection",
		zap.String("remote_addr", conn.RemoteAddr),
		zap.String("reason", reason),
	)
	return nil
}

// throttleConnection ограничивает пропускную способность
func (m *AIMiddleware) throttleConnection(ctx context.Context, conn *Connection, factor float64) error {
	m.logger.Info("Throttling connection",
		zap.String("remote_addr", conn.RemoteAddr),
		zap.Float64("factor", factor),
	)
	conn.ThrottleFactor = factor
	return nil
}

// adjustResources адаптирует ресурсы на основе предсказания
func (m *AIMiddleware) adjustResources(ctx context.Context, prediction *predictive.LoadPrediction) {
	if len(prediction.Forecast) > 0 {
		avgLoad := 0.0
		for _, load := range prediction.Forecast {
			avgLoad += load
		}
		avgLoad /= float64(len(prediction.Forecast))

		if avgLoad > 80.0 {
			m.logger.Info("High load predicted, scaling up resources",
				zap.Float64("predicted_load", avgLoad),
			)
		} else if avgLoad < 20.0 {
			m.logger.Info("Low load predicted, scaling down resources",
				zap.Float64("predicted_load", avgLoad),
			)
		}
	}
}

// GetMetrics возвращает метрики AI middleware
func (m *AIMiddleware) GetMetrics() *AIMetrics {
	return m.metrics
}

// CleanupCache очищает устаревшие записи кэша
func (m *AIMiddleware) CleanupCache() {
	m.cache.Cleanup()
}

// GetCacheStats возвращает статистику кэша
func (m *AIMiddleware) GetCacheStats() map[string]int {
	return m.cache.GetStats()
}

// GetCircuitBreakerStats возвращает статистику circuit breaker
func (m *AIMiddleware) GetCircuitBreakerStats() map[string]interface{} {
	return m.circuitBreaker.GetStats()
}

// AIMiddlewareGin создает Gin middleware для AI обработки
func AIMiddlewareGin(aiMiddleware *AIMiddleware) gin.HandlerFunc {
	return func(c *gin.Context) {
		conn := &Connection{
			ID:          c.GetString("connection_id"),
			RemoteAddr:  c.ClientIP(),
			Source:      c.GetHeader("X-Forwarded-For"),
			Destination: c.Request.Host,
			Protocol:    c.Request.Proto,
			Timestamp:   time.Now(),
		}

		if err := aiMiddleware.ProcessConnection(c.Request.Context(), conn); err != nil {
			aiMiddleware.logger.Error("AI processing failed", zap.Error(err))
		}

		c.Next()
	}
}
