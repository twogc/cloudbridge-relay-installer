package routing

import (
	"fmt"
	"math"
	"sync"
	"time"

	"go.uber.org/zap"
)

// RouteOptimizer оптимизирует маршруты P2P соединений
type RouteOptimizer struct {
	config           *Config
	logger           *zap.Logger
	bandit           *MultiArmedBandit
	latencyPredictor *LatencyPredictor
	pathFinder       *PathFinder

	// Metrics
	optimizationCount int64
	averageGain       float64

	// Thread safety
	mu sync.RWMutex
}

// Config конфигурация для оптимизатора маршрутов
type Config struct {
	Enabled         bool          `yaml:"enabled"`
	Algorithm       string        `yaml:"algorithm"`
	ExplorationRate float64       `yaml:"exploration_rate"`
	LearningRate    float64       `yaml:"learning_rate"`
	UpdateInterval  time.Duration `yaml:"update_interval"`
}

// OptimizedRoute оптимизированный маршрут
type OptimizedRoute struct {
	Protocol         string    `json:"protocol"`
	Path             []string  `json:"path"`
	PredictedLatency float64   `json:"predicted_latency"`
	Confidence       float64   `json:"confidence"`
	Timestamp        time.Time `json:"timestamp"`
}

// NewRouteOptimizer создает новый оптимизатор маршрутов
func NewRouteOptimizer(config *Config, logger *zap.Logger) *RouteOptimizer {
	return &RouteOptimizer{
		config:           config,
		logger:           logger,
		bandit:           NewMultiArmedBandit(config.ExplorationRate),
		latencyPredictor: NewLatencyPredictor(logger),
		pathFinder:       NewPathFinder(logger),
	}
}

// Initialize инициализирует оптимизатор маршрутов
func (ro *RouteOptimizer) Initialize() error {
	ro.logger.Info("Initializing Route Optimizer...")

	// Инициализируем multi-armed bandit
	ro.bandit.Initialize()

	// Инициализируем предсказатель латентности
	if err := ro.latencyPredictor.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize latency predictor: %w", err)
	}

	// Инициализируем поисковик путей
	if err := ro.pathFinder.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize path finder: %w", err)
	}

	ro.logger.Info("Route Optimizer initialized successfully")
	return nil
}

// Start запускает оптимизатор маршрутов
func (ro *RouteOptimizer) Start() error {
	ro.logger.Info("Starting Route Optimizer...")

	// Запускаем фоновое обновление
	go ro.backgroundUpdate()

	ro.logger.Info("Route Optimizer started successfully")
	return nil
}

// Stop останавливает оптимизатор маршрутов
func (ro *RouteOptimizer) Stop() error {
	ro.logger.Info("Stopping Route Optimizer...")

	// Сохраняем состояние
	if err := ro.saveState(); err != nil {
		ro.logger.Error("Failed to save state", zap.Error(err))
	}

	ro.logger.Info("Route Optimizer stopped successfully")
	return nil
}

// OptimizeRoute оптимизирует маршрут между источником и назначением
func (ro *RouteOptimizer) OptimizeRoute(source, destination string) (*OptimizedRoute, error) {
	ro.mu.Lock()
	defer ro.mu.Unlock()

	// Выбор оптимального протокола
	protocol := ro.bandit.Select(source, destination)

	// Предсказание латентности
	predictedLatency, err := ro.latencyPredictor.Predict(source, destination, protocol)
	if err != nil {
		return nil, fmt.Errorf("failed to predict latency: %w", err)
	}

	// Поиск оптимального пути
	path, err := ro.pathFinder.FindOptimalPath(source, destination, protocol)
	if err != nil {
		return nil, fmt.Errorf("failed to find optimal path: %w", err)
	}

	// Расчет уверенности
	confidence := ro.calculateConfidence(protocol, predictedLatency, path)

	// Обновляем метрики
	ro.optimizationCount++
	
	// Обновляем средний gain на основе confidence
	ro.updateAverageGain(confidence)

	result := &OptimizedRoute{
		Protocol:         protocol,
		Path:             path,
		PredictedLatency: predictedLatency,
		Confidence:       confidence,
		Timestamp:        time.Now(),
	}

	ro.logger.Info("Route optimization completed",
		zap.String("protocol", protocol),
		zap.Float64("latency", predictedLatency),
		zap.Float64("confidence", confidence))

	return result, nil
}

// updateAverageGain обновляет средний gain на основе confidence
func (ro *RouteOptimizer) updateAverageGain(confidence float64) {
	if ro.optimizationCount == 1 {
		ro.averageGain = confidence
	} else {
		// Экспоненциальное скользящее среднее
		alpha := 0.1 // Коэффициент сглаживания
		ro.averageGain = alpha*confidence + (1-alpha)*ro.averageGain
	}
}

// UpdateReward обновляет награду для выбранного протокола
func (ro *RouteOptimizer) UpdateReward(protocol string, actualLatency float64, predictedLatency float64) {
	ro.mu.Lock()
	defer ro.mu.Unlock()

	// Рассчитываем награду на основе точности предсказания
	reward := 1.0 - math.Abs(actualLatency-predictedLatency)/predictedLatency
	if reward < 0 {
		reward = 0
	}

	// Обновляем multi-armed bandit
	ro.bandit.UpdateReward(protocol, reward)

	// Обновляем средний выигрыш
	ro.averageGain = (ro.averageGain*float64(ro.optimizationCount-1) + reward) / float64(ro.optimizationCount)

	ro.logger.Debug("Updated reward for protocol",
		zap.String("protocol", protocol),
		zap.Float64("reward", reward),
		zap.Float64("average_gain", ro.averageGain))
}

// calculateConfidence рассчитывает уверенность в результате
func (ro *RouteOptimizer) calculateConfidence(protocol string, latency float64, path []string) float64 {
	// Базовая уверенность на основе количества использований протокола
	usageCount := ro.bandit.GetUsageCount(protocol)
	usageConfidence := math.Min(float64(usageCount)/100.0, 1.0)

	// Уверенность на основе латентности
	latencyConfidence := 1.0 - math.Min(latency/1000.0, 1.0) // Чем меньше латентность, тем выше уверенность

	// Уверенность на основе длины пути
	pathConfidence := 1.0 - math.Min(float64(len(path))/10.0, 1.0) // Чем короче путь, тем выше уверенность

	// Итоговая уверенность
	confidence := (usageConfidence + latencyConfidence + pathConfidence) / 3.0

	return math.Min(confidence, 1.0)
}

// backgroundUpdate выполняет фоновое обновление
func (ro *RouteOptimizer) backgroundUpdate() {
	ticker := time.NewTicker(ro.config.UpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ro.updateModels()
		}
	}
}

// updateModels обновляет модели
func (ro *RouteOptimizer) updateModels() {
	ro.mu.Lock()
	defer ro.mu.Unlock()

	// Обновляем предсказатель латентности
	ro.latencyPredictor.Update()

	// Обновляем поисковик путей
	ro.pathFinder.Update()

	ro.logger.Debug("Models updated")
}

// saveState сохраняет состояние оптимизатора
func (ro *RouteOptimizer) saveState() error {
	ro.logger.Info("Saving route optimizer state",
		zap.Int64("optimizations", ro.optimizationCount),
		zap.Float64("average_gain", ro.averageGain))

	return nil
}

// GetMetrics возвращает метрики оптимизатора
func (ro *RouteOptimizer) GetMetrics() map[string]interface{} {
	ro.mu.RLock()
	defer ro.mu.RUnlock()

	return map[string]interface{}{
		"optimization_count": ro.optimizationCount,
		"average_gain":       ro.averageGain,
		"exploration_rate":   ro.config.ExplorationRate,
		"learning_rate":      ro.config.LearningRate,
	}
}

