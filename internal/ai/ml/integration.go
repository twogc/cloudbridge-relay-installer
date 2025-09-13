package ml

import (
	"context"
	"fmt"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/internal/ai/anomaly"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai/predictive"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai/routing"
	"go.uber.org/zap"
)

// RealAIIntegration интегрирует реальные ML модели с существующим AI сервисом
type RealAIIntegration struct {
	realModels *RealMLModels
	logger     *zap.Logger
}

// NewRealAIIntegration создает новую интеграцию реальных ML моделей
func NewRealAIIntegration(logger *zap.Logger) *RealAIIntegration {
	return &RealAIIntegration{
		realModels: NewRealMLModels(logger),
		logger:     logger,
	}
}

// DetectAnomalyReal обнаруживает аномалии с помощью реальных ML моделей
func (rai *RealAIIntegration) DetectAnomalyReal(ctx context.Context, metrics *anomaly.NetworkMetrics) (*anomaly.AnomalyResult, error) {
	// Конвертируем NetworkMetrics в признаки для ML модели
	features := rai.convertNetworkMetricsToFeatures(metrics)

	// Используем реальный детектор аномалий
	score, isAnomaly, err := rai.realModels.anomalyDetector.DetectAnomaly(features)
	if err != nil {
		rai.logger.Error("Real anomaly detection failed", zap.Error(err))
		return nil, err
	}

	// Определяем тип и серьезность аномалии
	anomalyType := rai.determineAnomalyType(features, score)
	severity := rai.determineSeverity(score)

	// Создаем результат
	result := &anomaly.AnomalyResult{
		IsAnomaly:   isAnomaly,
		Score:       score,
		AnomalyType: anomalyType,
		Severity:    severity,
		Timestamp:   time.Now(),
		Features: map[string]float64{
			"connections_per_second": features[0],
			"bytes_per_second":       features[1],
			"packet_loss_rate":       features[2],
			"average_latency":        features[3],
			"cpu_usage":              features[4],
			"memory_usage":           features[5],
			"active_connections":     features[6],
			"failed_connections":     features[7],
			"traffic_intensity":      features[8],
			"resource_utilization":   features[9],
		},
		Confidence: rai.calculateConfidence(score),
		ModelID:    "isolation_forest",
	}

	rai.logger.Debug("Real anomaly detection completed",
		zap.Float64("score", score),
		zap.Bool("is_anomaly", isAnomaly),
		zap.String("type", anomalyType),
		zap.String("severity", severity),
	)

	return result, nil
}

// PredictLoadReal предсказывает нагрузку с помощью реальных ML моделей
func (rai *RealAIIntegration) PredictLoadReal(ctx context.Context, horizon int) (*predictive.LoadPrediction, error) {
	// Получаем исторические данные (в реальной системе это было бы из базы данных)
	historicalData := rai.getHistoricalLoadData()

	if len(historicalData) < 10 {
		rai.logger.Warn("Insufficient historical data for load prediction", zap.Int("data_points", len(historicalData)))
		// Возвращаем базовое предсказание
		return rai.generateBasicLoadPrediction(horizon), nil
	}

	// Используем реальный предиктор нагрузки
	predictions, err := rai.realModels.loadPredictor.PredictLoad(historicalData, horizon)
	if err != nil {
		rai.logger.Error("Real load prediction failed", zap.Error(err))
		return nil, err
	}

	// Создаем результат
	result := &predictive.LoadPrediction{
		Timestamp:  time.Now(),
		Horizon:    horizon,
		Forecast:   predictions,
		Confidence: rai.calculatePredictionConfidence(predictions),
		ModelID:    "ensemble",
		Features: map[string]float64{
			"trend":       rai.analyzeTrendValue(predictions),
			"seasonality": rai.detectSeasonalityValue(predictions),
			"variance":    rai.calculateVariance(predictions),
		},
		ProcessingTime: time.Since(time.Now()),
		Seasonality:    rai.detectSeasonalityValue(predictions),
		Trend:          rai.analyzeTrendValue(predictions),
	}

	rai.logger.Debug("Real load prediction completed",
		zap.Int("horizon", horizon),
		zap.Float64("avg_prediction", rai.average(predictions)),
		zap.Float64("confidence", result.Confidence),
	)

	return result, nil
}

// OptimizeRouteReal оптимизирует маршрут с помощью реальных ML моделей
func (rai *RealAIIntegration) OptimizeRouteReal(ctx context.Context, source, destination string, constraints map[string]interface{}) (*routing.OptimizedRoute, error) {
	// Используем реальный оптимизатор маршрутов
	optimizedRoute, err := rai.realModels.routeOptimizer.OptimizeRoute(source, destination, constraints)
	if err != nil {
		rai.logger.Error("Real route optimization failed", zap.Error(err))
		return nil, err
	}

	// Конвертируем в формат routing.OptimizedRoute
	result := &routing.OptimizedRoute{
		Path:             []string{optimizedRoute.Path}, // Преобразуем string в []string
		Protocol:         optimizedRoute.Protocol,
		PredictedLatency: optimizedRoute.PredictedLatency,
		Confidence:       optimizedRoute.Confidence,
		Timestamp:        time.Now(),
	}

	rai.logger.Debug("Real route optimization completed",
		zap.String("source", source),
		zap.String("destination", destination),
		zap.Float64("predicted_latency", optimizedRoute.PredictedLatency),
		zap.Float64("optimization_score", optimizedRoute.OptimizationScore),
	)

	return result, nil
}

// Вспомогательные методы

// convertNetworkMetricsToFeatures конвертирует NetworkMetrics в признаки
func (rai *RealAIIntegration) convertNetworkMetricsToFeatures(metrics *anomaly.NetworkMetrics) []float64 {
	features := make([]float64, 10)

	// Нормализуем признаки
	features[0] = float64(metrics.ConnectionsPerSecond) / 1000.0 // Нормализуем к 1000
	features[1] = float64(metrics.BytesPerSecond) / 1000000.0    // Нормализуем к MB/s
	features[2] = metrics.PacketLossRate
	features[3] = metrics.AverageLatency / 100.0 // Нормализуем к 100ms
	features[4] = metrics.CPUUsage / 100.0
	features[5] = metrics.MemoryUsage / 100.0
	features[6] = float64(metrics.ActiveConnections) / 1000.0
	features[7] = float64(metrics.FailedConnections) / 100.0

	// Вычисляем дополнительные признаки
	features[8] = rai.calculateTrafficIntensity(metrics)
	features[9] = rai.calculateResourceUtilization(metrics)

	return features
}

// calculateTrafficIntensity вычисляет интенсивность трафика
func (rai *RealAIIntegration) calculateTrafficIntensity(metrics *anomaly.NetworkMetrics) float64 {
	if metrics.ActiveConnections == 0 {
		return 0
	}
	return float64(metrics.ConnectionsPerSecond) / float64(metrics.ActiveConnections)
}

// calculateResourceUtilization вычисляет использование ресурсов
func (rai *RealAIIntegration) calculateResourceUtilization(metrics *anomaly.NetworkMetrics) float64 {
	return (metrics.CPUUsage + metrics.MemoryUsage) / 2.0
}

// determineAnomalyType определяет тип аномалии
func (rai *RealAIIntegration) determineAnomalyType(features []float64, score float64) string {
	if len(features) < 10 {
		return "unknown"
	}

	// Анализируем признаки для определения типа
	if features[0] > 0.8 { // Высокая интенсивность соединений
		return "connection_flood"
	}
	if features[1] > 0.9 { // Высокий трафик
		return "bandwidth_anomaly"
	}
	if features[2] > 0.1 { // Высокая потеря пакетов
		return "packet_loss_anomaly"
	}
	if features[3] > 0.8 { // Высокая задержка
		return "latency_anomaly"
	}
	if features[4] > 0.9 || features[5] > 0.9 { // Высокое использование ресурсов
		return "resource_anomaly"
	}

	return "general_anomaly"
}

// determineSeverity определяет серьезность аномалии
func (rai *RealAIIntegration) determineSeverity(score float64) string {
	if score > 0.9 {
		return "critical"
	} else if score > 0.7 {
		return "high"
	} else if score > 0.5 {
		return "medium"
	} else {
		return "low"
	}
}

// calculateConfidence вычисляет уверенность в результате
func (rai *RealAIIntegration) calculateConfidence(score float64) float64 {
	// Чем дальше от 0.5, тем выше уверенность
	return 1.0 - 2.0*abs(score-0.5)
}

// generateRecommendations генерирует рекомендации
func (rai *RealAIIntegration) generateRecommendations(anomalyType, severity string) []string {
	recommendations := []string{}

	switch anomalyType {
	case "connection_flood":
		recommendations = append(recommendations, "Implement rate limiting")
		recommendations = append(recommendations, "Check for DDoS attack")
	case "bandwidth_anomaly":
		recommendations = append(recommendations, "Monitor bandwidth usage")
		recommendations = append(recommendations, "Consider traffic shaping")
	case "packet_loss_anomaly":
		recommendations = append(recommendations, "Check network connectivity")
		recommendations = append(recommendations, "Verify routing configuration")
	case "latency_anomaly":
		recommendations = append(recommendations, "Check network latency")
		recommendations = append(recommendations, "Consider route optimization")
	case "resource_anomaly":
		recommendations = append(recommendations, "Monitor system resources")
		recommendations = append(recommendations, "Consider scaling")
	}

	if severity == "critical" || severity == "high" {
		recommendations = append(recommendations, "Immediate investigation required")
	}

	return recommendations
}

// getHistoricalLoadData получает исторические данные нагрузки
func (rai *RealAIIntegration) getHistoricalLoadData() []float64 {
	// В реальной системе это было бы из базы данных
	// Для демонстрации генерируем синтетические данные
	data := make([]float64, 100)
	for i := 0; i < 100; i++ {
		// Генерируем данные с трендом и сезонностью
		trend := float64(i) * 0.1
		seasonal := 10 * sin(2*3.14159*float64(i)/24) // 24-часовая сезонность
		noise := (sin(float64(i)*0.5) * 2)            // Шум
		data[i] = 50 + trend + seasonal + noise
	}
	return data
}

// generateBasicLoadPrediction генерирует базовое предсказание нагрузки
func (rai *RealAIIntegration) generateBasicLoadPrediction(horizon int) *predictive.LoadPrediction {
	predictions := make([]float64, horizon)
	for i := 0; i < horizon; i++ {
		predictions[i] = 50.0 // Базовое значение
	}

	return &predictive.LoadPrediction{
		Timestamp:      time.Now(),
		Horizon:        horizon,
		Forecast:       predictions,
		Confidence:     0.5,
		ModelID:        "basic",
		Features:       map[string]float64{"trend": 0.0, "seasonality": 0.0},
		ProcessingTime: time.Since(time.Now()),
		Seasonality:    0.0,
		Trend:          0.0,
	}
}

// calculatePredictionConfidence вычисляет уверенность в предсказании
func (rai *RealAIIntegration) calculatePredictionConfidence(predictions []float64) float64 {
	if len(predictions) == 0 {
		return 0.5
	}

	// Вычисляем вариацию предсказаний
	variance := rai.calculateVariance(predictions)

	// Чем меньше вариация, тем выше уверенность
	confidence := 1.0 - (variance / 100.0)
	return max(0.1, min(1.0, confidence))
}

// calculatePredictionAccuracy вычисляет точность предсказания
func (rai *RealAIIntegration) calculatePredictionAccuracy(predictions []float64) float64 {
	// В реальной системе это было бы на основе валидации
	// Для демонстрации возвращаем фиксированное значение
	return 0.85
}

// analyzeTrendValue анализирует тренд и возвращает числовое значение
func (rai *RealAIIntegration) analyzeTrendValue(predictions []float64) float64 {
	if len(predictions) < 2 {
		return 0.0
	}

	// Вычисляем средний тренд
	trend := (predictions[len(predictions)-1] - predictions[0]) / float64(len(predictions)-1)
	return trend
}

// detectSeasonalityValue обнаруживает сезонность и возвращает числовое значение
func (rai *RealAIIntegration) detectSeasonalityValue(predictions []float64) float64 {
	if len(predictions) < 24 {
		return 0.0
	}

	// Возвращаем коэффициент вариации как меру сезонности
	return rai.calculateVariance(predictions)
}

// generateLoadRecommendations генерирует рекомендации по нагрузке
func (rai *RealAIIntegration) generateLoadRecommendations(predictions []float64) []string {
	recommendations := []string{}

	avgLoad := rai.average(predictions)
	maxLoad := rai.max(predictions)

	if maxLoad > 80.0 {
		recommendations = append(recommendations, "Consider scaling up resources")
		recommendations = append(recommendations, "Monitor for potential overload")
	}

	if avgLoad < 20.0 {
		recommendations = append(recommendations, "Consider scaling down resources")
		recommendations = append(recommendations, "Optimize resource utilization")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Current load levels are optimal")
	}

	return recommendations
}

// generateAlternativeRoutes генерирует альтернативные маршруты
func (rai *RealAIIntegration) generateAlternativeRoutes(source, destination string) []string {
	return []string{
		fmt.Sprintf("%s->intermediate1->%s", source, destination),
		fmt.Sprintf("%s->intermediate2->%s", source, destination),
		fmt.Sprintf("%s->backup->%s", source, destination),
	}
}

// generateRouteRecommendations генерирует рекомендации по маршруту
func (rai *RealAIIntegration) generateRouteRecommendations(route *OptimizedRoute) []string {
	recommendations := []string{}

	if route.PredictedLatency > 50.0 {
		recommendations = append(recommendations, "Consider alternative routes with lower latency")
	}

	if route.Bandwidth < 1000.0 {
		recommendations = append(recommendations, "Monitor bandwidth utilization")
	}

	if route.Confidence < 0.7 {
		recommendations = append(recommendations, "Route prediction has low confidence, monitor closely")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Route is optimal")
	}

	return recommendations
}

// Вспомогательные математические функции

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func sin(x float64) float64 {
	// Упрощенная реализация sin для демонстрации
	return x - (x*x*x)/6 + (x*x*x*x*x)/120
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func (rai *RealAIIntegration) average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func (rai *RealAIIntegration) max(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	maxVal := values[0]
	for _, v := range values {
		if v > maxVal {
			maxVal = v
		}
	}
	return maxVal
}

func (rai *RealAIIntegration) calculateVariance(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}

	mean := rai.average(values)
	sumSquaredDiff := 0.0
	for _, v := range values {
		diff := v - mean
		sumSquaredDiff += diff * diff
	}
	return sumSquaredDiff / float64(len(values)-1)
}
