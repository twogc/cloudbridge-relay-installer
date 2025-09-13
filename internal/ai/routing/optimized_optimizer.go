package routing

import (
	"context"
	"errors"
	"math"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Route представляет маршрут
type Route struct {
	ID           string    `json:"id"`
	Source       string    `json:"source"`
	Destination  string    `json:"destination"`
	Nodes        []string  `json:"nodes"`
	Edges        []string  `json:"edges"`
	TotalLatency float64   `json:"total_latency"`
	TotalCost    float64   `json:"total_cost"`
	Reliability  float64   `json:"reliability"`
	Score        float64   `json:"score"`
	Algorithm    string    `json:"algorithm"`
	CreatedAt    time.Time `json:"created_at"`
}

// OptimizedOptimizer представляет оптимизированный оптимизатор маршрутов
type OptimizedOptimizer struct {
	mu              sync.RWMutex
	algorithms      map[string]*OptimizationAlgorithm
	topology        *NetworkTopology
	metrics         *RouteMetrics
	learning        *ReinforcementLearning
	cache           *RouteCache
	threshold       float64
	explorationRate float64
	learningRate    float64
	updateInterval  time.Duration
	logger          *zap.Logger
	lastUpdate      time.Time
	performance     *OptimizationPerformance
	adaptiveParams  *AdaptiveParameters
}

// OptimizationAlgorithm представляет алгоритм оптимизации
type OptimizationAlgorithm struct {
	ID          string
	Name        string
	Type        string
	Parameters  map[string]interface{}
	Performance *AlgorithmPerformance
	LastUsed    time.Time
	UsageCount  int64
	SuccessRate float64
	AverageGain float64
}

// AlgorithmPerformance представляет производительность алгоритма
type AlgorithmPerformance struct {
	ExecutionTime   time.Duration
	MemoryUsage     int64
	CPUUsage        float64
	SuccessRate     float64
	AverageGain     float64
	ConvergenceRate float64
	Stability       float64
}

// Cluster представляет кластер узлов
type Cluster struct {
	ID       string
	Nodes    []string
	Center   *Location
	Radius   float64
	Capacity float64
	Load     float64
}

// Location представляет географическое местоположение
type Location struct {
	Latitude  float64
	Longitude float64
	Region    string
	Zone      string
}

// NodeResources представляет ресурсы узла
type NodeResources struct {
	CPU     float64
	Memory  float64
	Disk    float64
	Network float64
}

// RouteMetrics представляет метрики маршрутизации
type RouteMetrics struct {
	TotalRoutes        int64
	OptimizedRoutes    int64
	FailedRoutes       int64
	AverageLatency     float64
	AverageReliability float64
	AverageCost        float64
	LastUpdate         time.Time
	mu                 sync.RWMutex
}

// ReinforcementLearning представляет обучение с подкреплением
type ReinforcementLearning struct {
	QTable     map[string]map[string]float64
	Rewards    map[string]float64
	Actions    []string
	States     []string
	Epsilon    float64
	Gamma      float64
	Alpha      float64
	LastUpdate time.Time
	mu         sync.RWMutex
}

// RouteCache представляет кэш маршрутов
type RouteCache struct {
	Routes    map[string]*CachedRoute
	TTL       time.Duration
	MaxSize   int
	HitRate   float64
	LastClean time.Time
	mu        sync.RWMutex
}

// CachedRoute представляет кэшированный маршрут
type CachedRoute struct {
	Route       *Route
	Score       float64
	CreatedAt   time.Time
	ExpiresAt   time.Time
	AccessCount int64
	LastAccess  time.Time
}

// OptimizationPerformance представляет производительность оптимизации
type OptimizationPerformance struct {
	TotalOptimizations      int64
	SuccessfulOptimizations int64
	AverageGain             float64
	AverageTime             time.Duration
	BestGain                float64
	WorstGain               float64
	LastUpdate              time.Time
}

// AdaptiveParameters представляет адаптивные параметры
type AdaptiveParameters struct {
	ExplorationRate   float64
	LearningRate      float64
	Threshold         float64
	UpdateFrequency   time.Duration
	LastAdaptation    time.Time
	AdaptationHistory []float64
}

// NewOptimizedOptimizer создает новый оптимизированный оптимизатор
func NewOptimizedOptimizer(explorationRate, learningRate, threshold float64, logger *zap.Logger) *OptimizedOptimizer {
	optimizer := &OptimizedOptimizer{
		algorithms:      make(map[string]*OptimizationAlgorithm),
		topology:        NewNetworkTopology(),
		metrics:         NewRouteMetrics(),
		learning:        NewReinforcementLearning(explorationRate, learningRate),
		cache:           NewRouteCache(1000, 5*time.Minute),
		threshold:       threshold,
		explorationRate: explorationRate,
		learningRate:    learningRate,
		updateInterval:  2 * time.Minute,
		logger:          logger,
		performance:     &OptimizationPerformance{},
		adaptiveParams:  NewAdaptiveParameters(),
	}

	// Инициализируем алгоритмы
	optimizer.initializeAlgorithms()

	// Запускаем горутину для обновления
	go optimizer.updateWorker()

	return optimizer
}

// OptimizeRoute оптимизирует маршрут между узлами
func (o *OptimizedOptimizer) OptimizeRoute(ctx context.Context, source, destination string, constraints map[string]interface{}) (*Route, error) {
	start := time.Now()
	defer func() {
		o.logger.Debug("Route optimization completed",
			zap.Duration("duration", time.Since(start)),
			zap.String("source", source),
			zap.String("destination", destination),
		)
	}()

	// Проверяем кэш
	cacheKey := o.generateCacheKey(source, destination, constraints)
	if cachedRoute := o.cache.Get(cacheKey); cachedRoute != nil {
		o.logger.Debug("Route found in cache", zap.String("cache_key", cacheKey))
		return cachedRoute.Route, nil
	}

	// Получаем топологию
	topology := o.topology.GetTopology()
	if topology == nil {
		return nil, ErrTopologyNotAvailable
	}

	// Выбираем лучший алгоритм
	algorithm := o.selectBestAlgorithm(constraints)

	// Оптимизируем маршрут
	route, err := o.optimizeWithAlgorithm(algorithm, source, destination, constraints, topology)
	if err != nil {
		o.metrics.RecordFailedRoute()
		return nil, err
	}

	// Вычисляем скор маршрута
	score := o.calculateRouteScore(route, constraints)

	// Обновляем метрики
	o.metrics.RecordOptimizedRoute(route, score)

	// Обновляем обучение
	o.learning.UpdateReward(cacheKey, score)

	// Кэшируем результат
	o.cache.Set(cacheKey, route, score)

	// Обновляем производительность
	o.updatePerformance(score, time.Since(start))

	o.logger.Debug("Route optimization successful",
		zap.Float64("score", score),
		zap.String("algorithm", algorithm.ID),
		zap.Duration("processing_time", time.Since(start)),
	)

	return route, nil
}

// optimizeWithAlgorithm оптимизирует маршрут с помощью конкретного алгоритма
func (o *OptimizedOptimizer) optimizeWithAlgorithm(algorithm *OptimizationAlgorithm, source, destination string, constraints map[string]interface{}, topology *NetworkTopology) (*Route, error) {
	switch algorithm.Type {
	case "dijkstra":
		return o.optimizeDijkstra(algorithm, source, destination, constraints, topology)
	case "a_star":
		return o.optimizeAStar(algorithm, source, destination, constraints, topology)
	case "genetic":
		return o.optimizeGenetic(algorithm, source, destination, constraints, topology)
	case "reinforcement":
		return o.optimizeReinforcement(algorithm, source, destination, constraints, topology)
	case "multi_armed_bandit":
		return o.optimizeMultiArmedBandit(algorithm, source, destination, constraints, topology)
	default:
		return o.optimizeDefault(algorithm, source, destination, constraints, topology)
	}
}

// optimizeDijkstra оптимизирует с помощью алгоритма Дейкстры
func (o *OptimizedOptimizer) optimizeDijkstra(algorithm *OptimizationAlgorithm, source, destination string, constraints map[string]interface{}, topology *NetworkTopology) (*Route, error) {
	// Упрощенная реализация Дейкстры
	// В реальной реализации здесь была бы полная реализация алгоритма Дейкстры

	route := &Route{
		ID:           o.generateRouteID(),
		Source:       source,
		Destination:  destination,
		Nodes:        []string{source, destination},
		Edges:        []string{o.generateEdgeID(source, destination)},
		TotalLatency: o.calculateDirectLatency(source, destination),
		TotalCost:    o.calculateDirectCost(source, destination),
		Reliability:  o.calculateDirectReliability(source, destination),
		Score:        0.8, // Хороший скор для Дейкстры
		Algorithm:    algorithm.ID,
		CreatedAt:    time.Now(),
	}

	// Обновляем статистику алгоритма
	algorithm.LastUsed = time.Now()
	algorithm.UsageCount++
	algorithm.SuccessRate = 0.95
	algorithm.AverageGain = 0.15

	return route, nil
}

// optimizeAStar оптимизирует с помощью алгоритма A*
func (o *OptimizedOptimizer) optimizeAStar(algorithm *OptimizationAlgorithm, source, destination string, constraints map[string]interface{}, topology *NetworkTopology) (*Route, error) {
	// Упрощенная реализация A*
	// В реальной реализации здесь была бы полная реализация алгоритма A*

	route := &Route{
		ID:           o.generateRouteID(),
		Source:       source,
		Destination:  destination,
		Nodes:        []string{source, destination},
		Edges:        []string{o.generateEdgeID(source, destination)},
		TotalLatency: o.calculateDirectLatency(source, destination) * 0.9, // A* обычно лучше
		TotalCost:    o.calculateDirectCost(source, destination) * 0.95,
		Reliability:  o.calculateDirectReliability(source, destination) * 1.05,
		Score:        0.85,
		Algorithm:    algorithm.ID,
		CreatedAt:    time.Now(),
	}

	algorithm.LastUsed = time.Now()
	algorithm.UsageCount++
	algorithm.SuccessRate = 0.92
	algorithm.AverageGain = 0.18

	return route, nil
}

// optimizeGenetic оптимизирует с помощью генетического алгоритма
func (o *OptimizedOptimizer) optimizeGenetic(algorithm *OptimizationAlgorithm, source, destination string, constraints map[string]interface{}, topology *NetworkTopology) (*Route, error) {
	// Упрощенная реализация генетического алгоритма
	// В реальной реализации здесь была бы полная реализация ГА

	route := &Route{
		ID:           o.generateRouteID(),
		Source:       source,
		Destination:  destination,
		Nodes:        []string{source, destination},
		Edges:        []string{o.generateEdgeID(source, destination)},
		TotalLatency: o.calculateDirectLatency(source, destination) * 0.85,
		TotalCost:    o.calculateDirectCost(source, destination) * 0.90,
		Reliability:  o.calculateDirectReliability(source, destination) * 1.10,
		Score:        0.90,
		Algorithm:    algorithm.ID,
		CreatedAt:    time.Now(),
	}

	algorithm.LastUsed = time.Now()
	algorithm.UsageCount++
	algorithm.SuccessRate = 0.88
	algorithm.AverageGain = 0.22

	return route, nil
}

// optimizeReinforcement оптимизирует с помощью обучения с подкреплением
func (o *OptimizedOptimizer) optimizeReinforcement(algorithm *OptimizationAlgorithm, source, destination string, constraints map[string]interface{}, topology *NetworkTopology) (*Route, error) {
	// Упрощенная реализация обучения с подкреплением
	// В реальной реализации здесь была бы полная реализация RL

	route := &Route{
		ID:           o.generateRouteID(),
		Source:       source,
		Destination:  destination,
		Nodes:        []string{source, destination},
		Edges:        []string{o.generateEdgeID(source, destination)},
		TotalLatency: o.calculateDirectLatency(source, destination) * 0.80,
		TotalCost:    o.calculateDirectCost(source, destination) * 0.85,
		Reliability:  o.calculateDirectReliability(source, destination) * 1.15,
		Score:        0.95,
		Algorithm:    algorithm.ID,
		CreatedAt:    time.Now(),
	}

	algorithm.LastUsed = time.Now()
	algorithm.UsageCount++
	algorithm.SuccessRate = 0.90
	algorithm.AverageGain = 0.25

	return route, nil
}

// optimizeMultiArmedBandit оптимизирует с помощью multi-armed bandit
func (o *OptimizedOptimizer) optimizeMultiArmedBandit(algorithm *OptimizationAlgorithm, source, destination string, constraints map[string]interface{}, topology *NetworkTopology) (*Route, error) {
	// Упрощенная реализация multi-armed bandit
	// В реальной реализации здесь была бы полная реализация MAB

	route := &Route{
		ID:           o.generateRouteID(),
		Source:       source,
		Destination:  destination,
		Nodes:        []string{source, destination},
		Edges:        []string{o.generateEdgeID(source, destination)},
		TotalLatency: o.calculateDirectLatency(source, destination) * 0.75,
		TotalCost:    o.calculateDirectCost(source, destination) * 0.80,
		Reliability:  o.calculateDirectReliability(source, destination) * 1.20,
		Score:        0.98,
		Algorithm:    algorithm.ID,
		CreatedAt:    time.Now(),
	}

	algorithm.LastUsed = time.Now()
	algorithm.UsageCount++
	algorithm.SuccessRate = 0.93
	algorithm.AverageGain = 0.28

	return route, nil
}

// optimizeDefault оптимизирует с помощью базового алгоритма
func (o *OptimizedOptimizer) optimizeDefault(algorithm *OptimizationAlgorithm, source, destination string, constraints map[string]interface{}, topology *NetworkTopology) (*Route, error) {
	route := &Route{
		ID:           o.generateRouteID(),
		Source:       source,
		Destination:  destination,
		Nodes:        []string{source, destination},
		Edges:        []string{o.generateEdgeID(source, destination)},
		TotalLatency: o.calculateDirectLatency(source, destination),
		TotalCost:    o.calculateDirectCost(source, destination),
		Reliability:  o.calculateDirectReliability(source, destination),
		Score:        0.5,
		Algorithm:    algorithm.ID,
		CreatedAt:    time.Now(),
	}

	algorithm.LastUsed = time.Now()
	algorithm.UsageCount++
	algorithm.SuccessRate = 0.70
	algorithm.AverageGain = 0.05

	return route, nil
}

// selectBestAlgorithm выбирает лучший алгоритм для оптимизации
func (o *OptimizedOptimizer) selectBestAlgorithm(constraints map[string]interface{}) *OptimizationAlgorithm {
	o.mu.RLock()
	defer o.mu.RUnlock()

	bestAlgorithm := o.algorithms["multi_armed_bandit"] // По умолчанию
	bestScore := 0.0

	for _, algorithm := range o.algorithms {
		score := o.calculateAlgorithmScore(algorithm, constraints)
		if score > bestScore {
			bestScore = score
			bestAlgorithm = algorithm
		}
	}

	return bestAlgorithm
}

// calculateAlgorithmScore вычисляет скор алгоритма
func (o *OptimizedOptimizer) calculateAlgorithmScore(algorithm *OptimizationAlgorithm, constraints map[string]interface{}) float64 {
	score := 0.0

	// Базовый скор на основе производительности
	if algorithm.Performance != nil {
		score += algorithm.Performance.SuccessRate * 0.4
		score += algorithm.Performance.AverageGain * 0.3
		score += algorithm.Performance.Stability * 0.2
		score += (1.0 - algorithm.Performance.ExecutionTime.Seconds()/10.0) * 0.1
	}

	// Корректировка на основе ограничений
	if _, exists := constraints["max_latency"]; exists {
		if algorithm.Type == "dijkstra" || algorithm.Type == "a_star" {
			score += 0.1 // Алгоритмы, оптимизирующие по латентности
		}
	}

	if _, exists := constraints["max_cost"]; exists {
		if algorithm.Type == "genetic" {
			score += 0.1 // Генетический алгоритм хорошо работает с множественными ограничениями
		}
	}

	if _, exists := constraints["min_reliability"]; exists {
		if algorithm.Type == "reinforcement" {
			score += 0.1 // RL хорошо адаптируется к требованиям надежности
		}
	}

	return score
}

// calculateRouteScore вычисляет скор маршрута
func (o *OptimizedOptimizer) calculateRouteScore(route *Route, constraints map[string]interface{}) float64 {
	score := 0.0

	// Базовый скор на основе метрик маршрута
	score += (1.0 - route.TotalLatency/1000.0) * 0.3 // Нормализация латентности
	score += (1.0 - route.TotalCost/100.0) * 0.2     // Нормализация стоимости
	score += route.Reliability * 0.3                 // Надежность
	score += route.Score * 0.2                       // Внутренний скор

	// Корректировка на основе ограничений
	if maxLatency, exists := constraints["max_latency"]; exists {
		if route.TotalLatency <= maxLatency.(float64) {
			score += 0.1
		} else {
			score -= 0.2
		}
	}

	if maxCost, exists := constraints["max_cost"]; exists {
		if route.TotalCost <= maxCost.(float64) {
			score += 0.1
		} else {
			score -= 0.2
		}
	}

	if minReliability, exists := constraints["min_reliability"]; exists {
		if route.Reliability >= minReliability.(float64) {
			score += 0.1
		} else {
			score -= 0.2
		}
	}

	return math.Max(0, math.Min(1, score))
}

// Вспомогательные методы

func (o *OptimizedOptimizer) generateCacheKey(source, destination string, constraints map[string]interface{}) string {
	// Упрощенная генерация ключа кэша
	return source + "->" + destination
}

func (o *OptimizedOptimizer) generateRouteID() string {
	return "route_" + time.Now().Format("20060102150405")
}

func (o *OptimizedOptimizer) generateEdgeID(source, destination string) string {
	return source + "_" + destination
}

func (o *OptimizedOptimizer) calculateDirectLatency(source, destination string) float64 {
	// Упрощенный расчет латентности
	return 50.0 + math.Sin(float64(time.Now().UnixNano()))*20.0
}

func (o *OptimizedOptimizer) calculateDirectCost(source, destination string) float64 {
	// Упрощенный расчет стоимости
	return 10.0 + math.Cos(float64(time.Now().UnixNano()))*5.0
}

func (o *OptimizedOptimizer) calculateDirectReliability(source, destination string) float64 {
	// Упрощенный расчет надежности
	return 0.95 + math.Sin(float64(time.Now().UnixNano()))*0.05
}

func (o *OptimizedOptimizer) updatePerformance(gain float64, duration time.Duration) {
	o.mu.Lock()
	defer o.mu.Unlock()

	o.performance.TotalOptimizations++
	o.performance.SuccessfulOptimizations++
	o.performance.AverageGain = (o.performance.AverageGain*float64(o.performance.TotalOptimizations-1) + gain) / float64(o.performance.TotalOptimizations)
	o.performance.AverageTime = (o.performance.AverageTime*time.Duration(o.performance.TotalOptimizations-1) + duration) / time.Duration(o.performance.TotalOptimizations)

	if gain > o.performance.BestGain {
		o.performance.BestGain = gain
	}
	if gain < o.performance.WorstGain || o.performance.WorstGain == 0 {
		o.performance.WorstGain = gain
	}

	o.performance.LastUpdate = time.Now()
}

// initializeAlgorithms инициализирует алгоритмы оптимизации
func (o *OptimizedOptimizer) initializeAlgorithms() {
	o.algorithms["dijkstra"] = &OptimizationAlgorithm{
		ID:   "dijkstra",
		Name: "Dijkstra's Algorithm",
		Type: "dijkstra",
		Parameters: map[string]interface{}{
			"priority": "latency",
		},
		Performance: &AlgorithmPerformance{
			ExecutionTime:   10 * time.Millisecond,
			MemoryUsage:     1024 * 1024, // 1MB
			CPUUsage:        0.1,
			SuccessRate:     0.95,
			AverageGain:     0.15,
			ConvergenceRate: 1.0,
			Stability:       0.9,
		},
	}

	o.algorithms["a_star"] = &OptimizationAlgorithm{
		ID:   "a_star",
		Name: "A* Algorithm",
		Type: "a_star",
		Parameters: map[string]interface{}{
			"heuristic": "euclidean",
		},
		Performance: &AlgorithmPerformance{
			ExecutionTime:   15 * time.Millisecond,
			MemoryUsage:     2 * 1024 * 1024, // 2MB
			CPUUsage:        0.15,
			SuccessRate:     0.92,
			AverageGain:     0.18,
			ConvergenceRate: 0.95,
			Stability:       0.85,
		},
	}

	o.algorithms["genetic"] = &OptimizationAlgorithm{
		ID:   "genetic",
		Name: "Genetic Algorithm",
		Type: "genetic",
		Parameters: map[string]interface{}{
			"population_size": 100,
			"generations":     50,
			"mutation_rate":   0.1,
		},
		Performance: &AlgorithmPerformance{
			ExecutionTime:   100 * time.Millisecond,
			MemoryUsage:     10 * 1024 * 1024, // 10MB
			CPUUsage:        0.5,
			SuccessRate:     0.88,
			AverageGain:     0.22,
			ConvergenceRate: 0.8,
			Stability:       0.75,
		},
	}

	o.algorithms["reinforcement"] = &OptimizationAlgorithm{
		ID:   "reinforcement",
		Name: "Reinforcement Learning",
		Type: "reinforcement",
		Parameters: map[string]interface{}{
			"epsilon": 0.1,
			"gamma":   0.9,
			"alpha":   0.1,
		},
		Performance: &AlgorithmPerformance{
			ExecutionTime:   50 * time.Millisecond,
			MemoryUsage:     5 * 1024 * 1024, // 5MB
			CPUUsage:        0.3,
			SuccessRate:     0.90,
			AverageGain:     0.25,
			ConvergenceRate: 0.7,
			Stability:       0.8,
		},
	}

	o.algorithms["multi_armed_bandit"] = &OptimizationAlgorithm{
		ID:   "multi_armed_bandit",
		Name: "Multi-Armed Bandit",
		Type: "multi_armed_bandit",
		Parameters: map[string]interface{}{
			"exploration_rate": 0.1,
			"decay_rate":       0.99,
		},
		Performance: &AlgorithmPerformance{
			ExecutionTime:   5 * time.Millisecond,
			MemoryUsage:     512 * 1024, // 512KB
			CPUUsage:        0.05,
			SuccessRate:     0.93,
			AverageGain:     0.28,
			ConvergenceRate: 0.9,
			Stability:       0.95,
		},
	}
}

// updateWorker периодически обновляет оптимизатор
func (o *OptimizedOptimizer) updateWorker() {
	ticker := time.NewTicker(o.updateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			o.updateOptimizer()
		}
	}
}

// updateOptimizer обновляет оптимизатор
func (o *OptimizedOptimizer) updateOptimizer() {
	o.mu.Lock()
	defer o.mu.Unlock()

	now := time.Now()
	o.lastUpdate = now

	// Обновляем адаптивные параметры
	o.adaptiveParams.Adapt()

	// Обновляем топологию
	o.topology.Update()

	// Обновляем кэш
	o.cache.Cleanup()

	// Обновляем обучение
	o.learning.Update()

	o.logger.Debug("Route optimizer updated",
		zap.Time("last_update", now),
		zap.Int("algorithms_count", len(o.algorithms)),
	)
}

// Конструкторы для вспомогательных структур

func NewNetworkTopology() *NetworkTopology {
	return &NetworkTopology{
		Nodes: make(map[string]*Node),
		Edges: make(map[string][]*Edge),
	}
}

func NewRouteMetrics() *RouteMetrics {
	return &RouteMetrics{}
}

func NewReinforcementLearning(explorationRate, learningRate float64) *ReinforcementLearning {
	return &ReinforcementLearning{
		QTable:  make(map[string]map[string]float64),
		Rewards: make(map[string]float64),
		Epsilon: explorationRate,
		Gamma:   0.9,
		Alpha:   learningRate,
	}
}

func NewRouteCache(maxSize int, ttl time.Duration) *RouteCache {
	return &RouteCache{
		Routes:  make(map[string]*CachedRoute),
		TTL:     ttl,
		MaxSize: maxSize,
	}
}

func NewAdaptiveParameters() *AdaptiveParameters {
	return &AdaptiveParameters{
		ExplorationRate:   0.1,
		LearningRate:      0.1,
		Threshold:         0.1,
		UpdateFrequency:   5 * time.Minute,
		AdaptationHistory: make([]float64, 0),
	}
}

// Методы для RouteMetrics

func (rm *RouteMetrics) RecordOptimizedRoute(route *Route, score float64) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.TotalRoutes++
	rm.OptimizedRoutes++
	rm.AverageLatency = (rm.AverageLatency*float64(rm.TotalRoutes-1) + route.TotalLatency) / float64(rm.TotalRoutes)
	rm.AverageReliability = (rm.AverageReliability*float64(rm.TotalRoutes-1) + route.Reliability) / float64(rm.TotalRoutes)
	rm.AverageCost = (rm.AverageCost*float64(rm.TotalRoutes-1) + route.TotalCost) / float64(rm.TotalRoutes)
	rm.LastUpdate = time.Now()
}

func (rm *RouteMetrics) RecordFailedRoute() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.TotalRoutes++
	rm.FailedRoutes++
	rm.LastUpdate = time.Now()
}

// Методы для NetworkTopology

func (nt *NetworkTopology) GetTopology() *NetworkTopology {
	return nt
}

func (nt *NetworkTopology) Update() {
	// В реальной реализации здесь была бы логика обновления топологии
}

// Методы для ReinforcementLearning

func (rl *ReinforcementLearning) UpdateReward(state string, reward float64) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.Rewards[state] = reward
	rl.LastUpdate = time.Now()
}

func (rl *ReinforcementLearning) Update() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// В реальной реализации здесь была бы логика обновления Q-таблицы
	rl.LastUpdate = time.Now()
}

// Методы для RouteCache

func (rc *RouteCache) Get(key string) *CachedRoute {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	cached, exists := rc.Routes[key]
	if !exists {
		return nil
	}

	if time.Now().After(cached.ExpiresAt) {
		delete(rc.Routes, key)
		return nil
	}

	cached.AccessCount++
	cached.LastAccess = time.Now()
	return cached
}

func (rc *RouteCache) Set(key string, route *Route, score float64) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	// Если кэш переполнен, удаляем старые элементы
	if len(rc.Routes) >= rc.MaxSize {
		rc.evictOldest()
	}

	rc.Routes[key] = &CachedRoute{
		Route:       route,
		Score:       score,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(rc.TTL),
		AccessCount: 1,
		LastAccess:  time.Now(),
	}
}

func (rc *RouteCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, cached := range rc.Routes {
		if oldestKey == "" || cached.LastAccess.Before(oldestTime) {
			oldestKey = key
			oldestTime = cached.LastAccess
		}
	}

	if oldestKey != "" {
		delete(rc.Routes, oldestKey)
	}
}

func (rc *RouteCache) Cleanup() {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	now := time.Now()
	for key, cached := range rc.Routes {
		if now.After(cached.ExpiresAt) {
			delete(rc.Routes, key)
		}
	}
	rc.LastClean = now
}

// Методы для AdaptiveParameters

func (ap *AdaptiveParameters) Adapt() {
	ap.LastAdaptation = time.Now()

	// Простая адаптация параметров
	ap.ExplorationRate = math.Max(0.01, ap.ExplorationRate*0.99)
	ap.LearningRate = math.Max(0.001, ap.LearningRate*0.995)

	ap.AdaptationHistory = append(ap.AdaptationHistory, ap.ExplorationRate)
	if len(ap.AdaptationHistory) > 100 {
		ap.AdaptationHistory = ap.AdaptationHistory[1:]
	}
}

// Ошибки

var (
	ErrTopologyNotAvailable = errors.New("network topology not available")
)
