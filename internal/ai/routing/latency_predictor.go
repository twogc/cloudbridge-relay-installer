package routing

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"go.uber.org/zap"
)

// LatencyPredictor предсказывает латентность для различных протоколов
type LatencyPredictor struct {
	logger           *zap.Logger
	models           map[string]*LatencyModel
	metricsCollector *NetworkMetricsCollector
	mu               sync.RWMutex
}

// LatencyModel модель для предсказания латентности
type LatencyModel struct {
	Protocol    string
	BaseLatency float64
	Variation   float64
	LastUpdate  time.Time
}

// NewLatencyPredictor создает новый предсказатель латентности
func NewLatencyPredictor(logger *zap.Logger, prometheusURL, cloudbridgeURL string) *LatencyPredictor {
	return &LatencyPredictor{
		logger:           logger,
		models:           make(map[string]*LatencyModel),
		metricsCollector: NewNetworkMetricsCollector(logger, prometheusURL, cloudbridgeURL),
	}
}

// Initialize инициализирует предсказатель латентности
func (lp *LatencyPredictor) Initialize() error {
	lp.logger.Info("Initializing Latency Predictor...")

	// Инициализируем модели для различных протоколов
	protocols := []string{"QUIC", "P2P_QUIC", "TCP", "UDP"}

	for _, protocol := range protocols {
		lp.models[protocol] = &LatencyModel{
			Protocol:    protocol,
			BaseLatency: lp.getBaseLatency(protocol),
			Variation:   lp.getVariation(protocol),
			LastUpdate:  time.Now(),
		}
	}

	// Собираем начальные метрики
	if err := lp.metricsCollector.CollectMetrics(); err != nil {
		lp.logger.Warn("Failed to collect initial metrics", zap.Error(err))
	}

	lp.logger.Info("Latency Predictor initialized successfully")
	return nil
}

// Predict предсказывает латентность для заданного маршрута
func (lp *LatencyPredictor) Predict(source, destination string, protocol string) (float64, error) {
	lp.mu.RLock()
	defer lp.mu.RUnlock()

	model, exists := lp.models[protocol]
	if !exists {
		return 0, fmt.Errorf("unknown protocol: %s", protocol)
	}

	// Базовая латентность протокола
	baseLatency := model.BaseLatency

	// Корректировка на основе расстояния (эвристика)
	distance := lp.calculateDistance(source, destination)
	distanceLatency := distance * 0.1 // 0.1ms на км

	// Корректировка на основе времени суток
	timeLatency := lp.getTimeBasedLatency()

	// Корректировка на основе нагрузки
	loadLatency := lp.getLoadBasedLatency()

	// Итоговая латентность
	totalLatency := baseLatency + distanceLatency + timeLatency + loadLatency

	// Добавляем вариацию на основе реальных метрик сети
	variation := lp.calculateRealVariation(source, destination, model)
	totalLatency += variation

	// Ограничиваем значения
	if totalLatency < 1 {
		totalLatency = 1
	}
	if totalLatency > 1000 {
		totalLatency = 1000
	}

	return totalLatency, nil
}

// Update обновляет модели предсказания
func (lp *LatencyPredictor) Update() {
	lp.mu.Lock()
	defer lp.mu.Unlock()

	// Обновляем модели на основе текущих условий
	for _, model := range lp.models {
		model.LastUpdate = time.Now()

		// Простое обновление базовой латентности
		model.BaseLatency = lp.getBaseLatency(model.Protocol)
		model.Variation = lp.getVariation(model.Protocol)
	}
}

// getBaseLatency возвращает базовую латентность для протокола
func (lp *LatencyPredictor) getBaseLatency(protocol string) float64 {
	switch protocol {
	case "QUIC":
		return 10.0 // QUIC обычно быстрее
	case "P2P_QUIC":
		return 12.0 // P2P QUIC быстрый
	case "TCP":
		return 25.0 // TCP стандартная латентность
	case "UDP":
		return 20.0 // UDP быстрее TCP
	default:
		return 30.0 // По умолчанию
	}
}

// getVariation возвращает вариацию латентности для протокола
func (lp *LatencyPredictor) getVariation(protocol string) float64 {
	switch protocol {
	case "QUIC":
		return 2.0 // QUIC стабильный
	case "P2P_QUIC":
		return 2.5 // P2P QUIC стабильный
	case "TCP":
		return 5.0 // TCP может иметь вариации
	case "UDP":
		return 4.0 // UDP умеренная вариация
	default:
		return 6.0 // По умолчанию
	}
}

// calculateDistance вычисляет расстояние между узлами (эвристика)
func (lp *LatencyPredictor) calculateDistance(source, destination string) float64 {
	// Эвристика на основе IP адресов

	// Простая эвристика: считаем количество различий в IP
	distance := 0.0

	// Если это локальные адреса
	if lp.isLocalAddress(source) && lp.isLocalAddress(destination) {
		distance = 1.0 // 1 км
	} else if lp.isLocalAddress(source) || lp.isLocalAddress(destination) {
		distance = 100.0 // 100 км
	} else {
		distance = 1000.0 // 1000 км
	}

	return distance
}

// isLocalAddress проверяет, является ли адрес локальным
func (lp *LatencyPredictor) isLocalAddress(address string) bool {
	// Простая проверка на локальные адреса
	return len(address) > 0 && (address[0] == '1' || address[0] == '2')
}

// getTimeBasedLatency возвращает латентность на основе времени суток
func (lp *LatencyPredictor) getTimeBasedLatency() float64 {
	hour := time.Now().Hour()

	// Пиковые часы (9-17) имеют большую латентность
	if hour >= 9 && hour <= 17 {
		return 5.0 // +5ms в пиковые часы
	}

	// Ночные часы (0-6) имеют меньшую латентность
	if hour >= 0 && hour <= 6 {
		return -2.0 // -2ms в ночные часы
	}

	return 0.0 // Обычная латентность
}

// getLoadBasedLatency возвращает латентность на основе реальной нагрузки
func (lp *LatencyPredictor) getLoadBasedLatency() float64 {
	// Получаем реальные метрики сети
	allMetrics := lp.metricsCollector.GetAllNodeMetrics()
	if len(allMetrics) == 0 {
		return 0.0
	}

	// Вычисляем среднюю нагрузку по всем узлам
	totalLoad := 0.0
	nodeCount := 0

	for _, metrics := range allMetrics {
		// Комбинируем различные метрики для оценки нагрузки
		nodeLoad := (metrics.CPUUsage + metrics.MemoryUsage + metrics.NetworkLoad) / 3.0
		totalLoad += nodeLoad
		nodeCount++
	}

	if nodeCount == 0 {
		return 0.0
	}

	averageLoad := totalLoad / float64(nodeCount)

	// Корректируем латентность на основе реальной нагрузки
	// Нормализуем к разумному диапазону
	return averageLoad * 0.1 // До +10ms при высокой нагрузке
}

// calculateRealVariation вычисляет вариацию латентности на основе реальных метрик
func (lp *LatencyPredictor) calculateRealVariation(source, destination string, model *LatencyModel) float64 {
	// Получаем метрики узлов
	sourceMetrics, sourceExists := lp.metricsCollector.GetNodeMetrics(source)
	destMetrics, destExists := lp.metricsCollector.GetNodeMetrics(destination)

	if !sourceExists || !destExists {
		// Если нет метрик, используем базовую вариацию
		return model.Variation * 0.1
	}

	// Вычисляем вариацию на основе нестабильности метрик
	sourceStability := lp.calculateNodeStability(sourceMetrics)
	destStability := lp.calculateNodeStability(destMetrics)

	// Средняя нестабильность
	averageStability := (sourceStability + destStability) / 2.0

	// Вариация пропорциональна нестабильности
	variation := model.Variation * averageStability

	return variation
}

// calculateNodeStability вычисляет стабильность узла на основе его метрик
func (lp *LatencyPredictor) calculateNodeStability(metrics *NodeMetrics) float64 {
	// Вычисляем коэффициент вариации для различных метрик
	cpuStability := 1.0 - math.Min(1.0, metrics.CPUUsage/100.0)
	memoryStability := 1.0 - math.Min(1.0, metrics.MemoryUsage/100.0)
	networkStability := 1.0 - math.Min(1.0, metrics.NetworkLoad/100.0)
	errorStability := 1.0 - math.Min(1.0, metrics.ErrorRate)

	// Средняя стабильность (чем выше, тем стабильнее)
	averageStability := (cpuStability + memoryStability + networkStability + errorStability) / 4.0

	// Возвращаем нестабильность (обратное значение)
	return 1.0 - averageStability
}

// StartMetricsCollection запускает периодический сбор метрик
func (lp *LatencyPredictor) StartMetricsCollection(ctx context.Context) {
	lp.metricsCollector.StartPeriodicCollection(ctx)
}

// GetMetricsCollector возвращает коллектор метрик
func (lp *LatencyPredictor) GetMetricsCollector() *NetworkMetricsCollector {
	return lp.metricsCollector
}
