package routing

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"go.uber.org/zap"
)

// PathFinder находит оптимальные пути между узлами
type PathFinder struct {
	logger           *zap.Logger
	topology         *NetworkTopology
	metricsCollector *NetworkMetricsCollector
	mu               sync.RWMutex
}

// NetworkTopology представляет топологию сети
type NetworkTopology struct {
	Nodes map[string]*Node
	Edges map[string][]*Edge
}

// Node представляет узел в сети
type Node struct {
	ID       string
	Location string
	Capacity int
	Load     int
}

// Edge представляет соединение между узлами
type Edge struct {
	From      string
	To        string
	Latency   float64
	Bandwidth int
	Protocol  string
}

// NewPathFinder создает новый поисковик путей
func NewPathFinder(logger *zap.Logger, prometheusURL, cloudbridgeURL string) *PathFinder {
	return &PathFinder{
		logger: logger,
		topology: &NetworkTopology{
			Nodes: make(map[string]*Node),
			Edges: make(map[string][]*Edge),
		},
		metricsCollector: NewNetworkMetricsCollector(logger, prometheusURL, cloudbridgeURL),
	}
}

// Initialize инициализирует поисковик путей
func (pf *PathFinder) Initialize() error {
	pf.logger.Info("Initializing Path Finder...")

	// Инициализируем базовую топологию
	if err := pf.initializeTopology(); err != nil {
		return fmt.Errorf("failed to initialize topology: %w", err)
	}

	// Инициализируем коллектор метрик
	nodes := make([]*Node, 0, len(pf.topology.Nodes))
	for _, node := range pf.topology.Nodes {
		nodes = append(nodes, node)
	}
	pf.metricsCollector.InitializeNodes(nodes)

	// Собираем начальные метрики
	if err := pf.metricsCollector.CollectMetrics(); err != nil {
		pf.logger.Warn("Failed to collect initial metrics", zap.Error(err))
	}

	pf.logger.Info("Path Finder initialized successfully")
	return nil
}

// FindOptimalPath находит оптимальный путь между узлами
func (pf *PathFinder) FindOptimalPath(source, destination string, protocol string) ([]string, error) {
	pf.mu.RLock()
	defer pf.mu.RUnlock()

	// Проверяем существование узлов
	if _, exists := pf.topology.Nodes[source]; !exists {
		return nil, fmt.Errorf("source node not found: %s", source)
	}
	if _, exists := pf.topology.Nodes[destination]; !exists {
		return nil, fmt.Errorf("destination node not found: %s", destination)
	}

	// Если источник и назначение одинаковые
	if source == destination {
		return []string{source}, nil
	}

	// Ищем оптимальный путь
	path := pf.dijkstra(source, destination, protocol)

	if len(path) == 0 {
		return nil, fmt.Errorf("no path found from %s to %s", source, destination)
	}

	return path, nil
}

// Update обновляет топологию сети
func (pf *PathFinder) Update() {
	pf.mu.Lock()
	defer pf.mu.Unlock()

	// Собираем актуальные метрики
	if err := pf.metricsCollector.CollectMetrics(); err != nil {
		pf.logger.Warn("Failed to collect metrics during update", zap.Error(err))
	}

	// Обновляем загрузку узлов на основе реальных метрик
	for _, node := range pf.topology.Nodes {
		node.Load = pf.calculateNodeLoad(node.ID)
	}

	// Обновляем латентность соединений
	for _, edges := range pf.topology.Edges {
		for _, edge := range edges {
			edge.Latency = pf.calculateEdgeLatency(edge)
		}
	}
}

// initializeTopology инициализирует базовую топологию сети
func (pf *PathFinder) initializeTopology() error {
	// Создаем базовые узлы
	nodes := []*Node{
		{ID: "node1", Location: "US-East", Capacity: 1000, Load: 0},
		{ID: "node2", Location: "US-West", Capacity: 1000, Load: 0},
		{ID: "node3", Location: "EU-Central", Capacity: 1000, Load: 0},
		{ID: "node4", Location: "Asia-Pacific", Capacity: 1000, Load: 0},
		{ID: "node5", Location: "US-Central", Capacity: 1000, Load: 0},
	}

	for _, node := range nodes {
		pf.topology.Nodes[node.ID] = node
		pf.topology.Edges[node.ID] = make([]*Edge, 0)
	}

	// Создаем соединения между узлами
	edges := []*Edge{
		{From: "node1", To: "node2", Latency: 50, Bandwidth: 1000, Protocol: "QUIC"},
		{From: "node1", To: "node5", Latency: 30, Bandwidth: 1000, Protocol: "QUIC"},
		{From: "node2", To: "node3", Latency: 100, Bandwidth: 800, Protocol: "TCP"},
		{From: "node3", To: "node4", Latency: 80, Bandwidth: 900, Protocol: "P2P_QUIC"},
		{From: "node4", To: "node5", Latency: 120, Bandwidth: 700, Protocol: "TCP"},
		{From: "node5", To: "node1", Latency: 30, Bandwidth: 1000, Protocol: "QUIC"},
	}

	for _, edge := range edges {
		pf.topology.Edges[edge.From] = append(pf.topology.Edges[edge.From], edge)
	}

	return nil
}

// dijkstra реализует алгоритм Дейкстры для поиска кратчайшего пути
func (pf *PathFinder) dijkstra(source, destination string, protocol string) []string {
	// Инициализация
	distances := make(map[string]float64)
	previous := make(map[string]string)
	visited := make(map[string]bool)

	// Инициализируем расстояния
	for nodeID := range pf.topology.Nodes {
		distances[nodeID] = math.Inf(1)
	}
	distances[source] = 0

	// Основной цикл алгоритма Дейкстры
	for {
		// Находим узел с минимальным расстоянием
		current := ""
		minDist := math.Inf(1)

		for nodeID, dist := range distances {
			if !visited[nodeID] && dist < minDist {
				minDist = dist
				current = nodeID
			}
		}

		// Если не нашли узел или достигли назначения
		if current == "" || current == destination {
			break
		}

		visited[current] = true

		// Обновляем расстояния до соседних узлов
		for _, edge := range pf.topology.Edges[current] {
			if !visited[edge.To] {
				// Учитываем протокол при расчете веса
				weight := pf.calculateEdgeWeight(edge, protocol)
				newDist := distances[current] + weight

				if newDist < distances[edge.To] {
					distances[edge.To] = newDist
					previous[edge.To] = current
				}
			}
		}
	}

	// Восстанавливаем путь
	path := make([]string, 0)
	current := destination

	for current != "" {
		path = append([]string{current}, path...)
		current = previous[current]
	}

	return path
}

// calculateEdgeWeight рассчитывает вес ребра с учетом протокола
func (pf *PathFinder) calculateEdgeWeight(edge *Edge, protocol string) float64 {
	// Базовый вес - латентность
	weight := edge.Latency

	// Корректировка на основе протокола
	switch protocol {
	case "QUIC":
		weight *= 0.8 // QUIC быстрее
	case "P2P_QUIC":
		weight *= 0.85 // P2P QUIC быстрый
	case "TCP":
		weight *= 1.0 // TCP стандартный
	case "UDP":
		weight *= 0.95 // UDP немного быстрее
	}

	// Корректировка на основе загрузки узла
	toNode := pf.topology.Nodes[edge.To]
	if toNode != nil {
		loadFactor := float64(toNode.Load) / float64(toNode.Capacity)
		weight *= (1.0 + loadFactor) // Увеличиваем вес при высокой загрузке
	}

	// Корректировка на основе пропускной способности
	bandwidthFactor := 1000.0 / float64(edge.Bandwidth)
	weight *= (1.0 + bandwidthFactor*0.1) // Увеличиваем вес при низкой пропускной способности

	return weight
}

// calculateNodeLoad рассчитывает загрузку узла на основе реальных метрик
func (pf *PathFinder) calculateNodeLoad(nodeID string) int {
	// Получаем реальные метрики узла
	metrics, exists := pf.metricsCollector.GetNodeMetrics(nodeID)
	if !exists {
		pf.logger.Warn("No metrics found for node", zap.String("node_id", nodeID))
		return 0
	}

	// Рассчитываем загрузку на основе реальных метрик
	// Используем комбинацию CPU, памяти, сетевой нагрузки и активных соединений
	cpuLoad := metrics.CPUUsage * 0.3
	memoryLoad := metrics.MemoryUsage * 0.2
	networkLoad := metrics.NetworkLoad * 0.3
	connectionLoad := float64(metrics.ActiveConns) * 0.2

	totalLoad := cpuLoad + memoryLoad + networkLoad + connectionLoad

	// Нормализуем загрузку к диапазону 0-1000
	normalizedLoad := int(totalLoad * 10) // Масштабируем для получения разумных значений

	// Ограничиваем загрузку
	if normalizedLoad < 0 {
		normalizedLoad = 0
	}
	if normalizedLoad > 1000 {
		normalizedLoad = 1000
	}

	pf.logger.Debug("Calculated node load",
		zap.String("node_id", nodeID),
		zap.Int("load", normalizedLoad),
		zap.Float64("cpu", metrics.CPUUsage),
		zap.Float64("memory", metrics.MemoryUsage),
		zap.Float64("network", metrics.NetworkLoad),
		zap.Int("connections", metrics.ActiveConns))

	return normalizedLoad
}

// calculateEdgeLatency рассчитывает латентность соединения на основе реальных метрик
func (pf *PathFinder) calculateEdgeLatency(edge *Edge) float64 {
	// Базовая латентность из конфигурации
	baseLatency := edge.Latency

	// Получаем метрики узлов для корректировки
	fromMetrics, fromExists := pf.metricsCollector.GetNodeMetrics(edge.From)
	toMetrics, toExists := pf.metricsCollector.GetNodeMetrics(edge.To)

	// Корректировка на основе загрузки узлов
	if fromExists && toExists {
		// Учитываем загрузку узлов
		fromLoad := (fromMetrics.CPUUsage + fromMetrics.MemoryUsage + fromMetrics.NetworkLoad) / 3.0
		toLoad := (toMetrics.CPUUsage + toMetrics.MemoryUsage + toMetrics.NetworkLoad) / 3.0

		// Увеличиваем латентность при высокой загрузке
		loadFactor := (fromLoad + toLoad) / 200.0 // Нормализуем к разумному диапазону
		baseLatency *= (1.0 + loadFactor)
	}

	// Корректировка на основе времени суток (пиковые часы)
	hour := time.Now().Hour()
	if hour >= 9 && hour <= 17 {
		baseLatency *= 1.1 // +10% в пиковые часы (уменьшили с 20%)
	}

	// Корректировка на основе протокола
	switch edge.Protocol {
	case "QUIC":
		baseLatency *= 0.8
	case "P2P_QUIC":
		baseLatency *= 0.9
	case "TCP":
		baseLatency *= 1.0
	case "UDP":
		baseLatency *= 0.95
	}

	// Корректировка на основе ошибок сети
	if fromExists && toExists {
		errorFactor := (fromMetrics.ErrorRate + toMetrics.ErrorRate) / 2.0
		baseLatency *= (1.0 + errorFactor)
	}

	return baseLatency
}

// StartMetricsCollection запускает периодический сбор метрик
func (pf *PathFinder) StartMetricsCollection(ctx context.Context) {
	pf.metricsCollector.StartPeriodicCollection(ctx)
}

// GetMetricsCollector возвращает коллектор метрик
func (pf *PathFinder) GetMetricsCollector() *NetworkMetricsCollector {
	return pf.metricsCollector
}
