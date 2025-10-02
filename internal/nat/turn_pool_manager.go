package nat

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"

	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
)

// TURNPoolManager управляет несколькими пулами TURN серверов
type TURNPoolManager struct {
	logger *zap.Logger

	// Пулы серверов по регионам/типам
	pools      map[string]*TURNPool
	poolsMutex sync.RWMutex

	// Активные allocations
	allocations      map[string]*TURNAllocation
	allocationsMutex sync.RWMutex

	// Конфигурация
	config *TURNPoolConfig

	// Метрики
	metrics *TURNPoolMetrics

	// Контекст
	ctx    context.Context
	cancel context.CancelFunc
}

// TURNServerPool представляет TURN сервер в пуле
type TURNServerPool struct {
	Address        string            `json:"address"`
	Username       string            `json:"username"`
	Password       string            `json:"password"`
	Realm          string            `json:"realm"`
	Priority       int               `json:"priority"`
	Weight         int               `json:"weight"`
	MaxConnections int               `json:"max_connections"`
	Region         string            `json:"region"`
	Provider       string            `json:"provider"`
	Capabilities   []string          `json:"capabilities"`
	Tags           map[string]string `json:"tags"`
}

// TURNPool представляет пул TURN серверов
type TURNPool struct {
	Name    string            `json:"name"`
	Region  string            `json:"region"`
	Servers []*TURNServerInfo `json:"servers"`
	mu      sync.RWMutex
}

// TURNAllocation представляет TURN allocation
type TURNAllocation struct {
	ID           string            `json:"id"`
	ClientID     string            `json:"client_id"`
	Server       *TURNServerInfo   `json:"server"`
	RelayAddress string            `json:"relay_address"`
	RelayPort    int               `json:"relay_port"`
	CreatedAt    time.Time         `json:"created_at"`
	ExpiresAt    time.Time         `json:"expires_at"`
	Lifetime     time.Duration     `json:"lifetime"`
	Bandwidth    int64             `json:"bandwidth"`
	Permissions  []string          `json:"permissions"`
	Channels     []string          `json:"channels"`
	Tags         map[string]string `json:"tags"`
	Status       string            `json:"status"` // active, expired, failed
	LastActivity time.Time         `json:"last_activity"`
}

// TURNPoolConfig конфигурация TURN pool manager
type TURNPoolConfig struct {
	// Основные настройки
	DefaultPoolName         string        `json:"default_pool_name"`
	AllocationTimeout       time.Duration `json:"allocation_timeout"`
	HealthCheckInterval     time.Duration `json:"health_check_interval"`
	HealthCheckTimeout      time.Duration `json:"health_check_timeout"`
	MaxAllocationsPerClient int           `json:"max_allocations_per_client"`

	// Балансировка
	LoadBalancingStrategy string  `json:"load_balancing_strategy"` // round_robin, weighted, least_connections
	WeightedThreshold     float64 `json:"weighted_threshold"`

	// Мониторинг
	MetricsEnabled     bool          `json:"metrics_enabled"`
	StatsInterval      time.Duration `json:"stats_interval"`
	LatencyHistorySize int           `json:"latency_history_size"`

	// Failover
	FailoverEnabled   bool          `json:"failover_enabled"`
	FailoverTimeout   time.Duration `json:"failover_timeout"`
	MaxRetries        int           `json:"max_retries"`
	BackoffMultiplier float64       `json:"backoff_multiplier"`
}

// TURNPoolMetrics метрики для TURN pool
type TURNPoolMetrics struct {
	// Счетчики
	TotalServers      prometheus.Gauge
	HealthyServers    prometheus.Gauge
	TotalAllocations  prometheus.Gauge
	ActiveAllocations prometheus.Gauge

	// Счетчики операций
	AllocationRequests prometheus.Counter
	AllocationFailures prometheus.Counter
	HealthCheckTotal   prometheus.Counter
	HealthCheckFailed  prometheus.Counter

	// Гистограммы
	AllocationDuration prometheus.Histogram
	HealthCheckLatency prometheus.Histogram

	// Векторы метрик по серверам
	ServerConnections *prometheus.GaugeVec
	ServerLatency     *prometheus.GaugeVec
	ServerStatus      *prometheus.GaugeVec
}

// TURNServerInfo расширенная информация о TURN сервере
type TURNServerInfo struct {
	*TURNServerPool

	// Дополнительные поля
	Region       string            `json:"region"`
	Provider     string            `json:"provider"`
	Capabilities []string          `json:"capabilities"`
	Tags         map[string]string `json:"tags"`

	// Статистика
	LatencyHistory []time.Duration `json:"latency_history"`
	ErrorRate      float64         `json:"error_rate"`
	SuccessRate    float64         `json:"success_rate"`
	LastError      error           `json:"last_error,omitempty"`

	// Состояние
	Status        string    `json:"status"` // active, degraded, maintenance, failed
	StatusChanged time.Time `json:"status_changed"`
}

// NewTURNPool создает новый TURN пул
func NewTURNPool(logger *zap.Logger) *TURNPool {
	return &TURNPool{
		Name:    "default",
		Region:  "global",
		Servers: make([]*TURNServerInfo, 0),
	}
}

// NewTURNPoolManager создает новый TURN pool manager
func NewTURNPoolManager(config *TURNPoolConfig, logger *zap.Logger) *TURNPoolManager {
	ctx, cancel := context.WithCancel(context.Background())

	// Установка значений по умолчанию
	if config == nil {
		config = &TURNPoolConfig{
			DefaultPoolName:         "default",
			AllocationTimeout:       10 * time.Minute,
			HealthCheckInterval:     30 * time.Second,
			HealthCheckTimeout:      5 * time.Second,
			MaxAllocationsPerClient: 10,
			LoadBalancingStrategy:   "weighted",
			WeightedThreshold:       0.8,
			MetricsEnabled:          true,
			StatsInterval:           60 * time.Second,
			LatencyHistorySize:      100,
			FailoverEnabled:         true,
			FailoverTimeout:         30 * time.Second,
			MaxRetries:              3,
			BackoffMultiplier:       2.0,
		}
	}

	manager := &TURNPoolManager{
		logger:      logger,
		pools:       make(map[string]*TURNPool),
		allocations: make(map[string]*TURNAllocation),
		config:      config,
		ctx:         ctx,
		cancel:      cancel,
	}

	// Инициализация метрик
	if config.MetricsEnabled {
		manager.initMetrics()
	}

	// Запуск фоновых процессов
	go manager.startStatsCollector()
	go manager.startAllocationCleanup()

	logger.Info("TURN Pool Manager initialized",
		zap.String("default_pool", config.DefaultPoolName),
		zap.String("strategy", config.LoadBalancingStrategy),
		zap.Bool("metrics_enabled", config.MetricsEnabled))

	return manager
}

// initMetrics инициализирует Prometheus метрики
func (tpm *TURNPoolManager) initMetrics() {
	tpm.metrics = &TURNPoolMetrics{
		TotalServers: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "turn_pool_total_servers",
			Help: "Total number of TURN servers in all pools",
		}),
		HealthyServers: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "turn_pool_healthy_servers",
			Help: "Number of healthy TURN servers",
		}),
		TotalAllocations: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "turn_pool_total_allocations",
			Help: "Total number of TURN allocations",
		}),
		ActiveAllocations: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "turn_pool_active_allocations",
			Help: "Number of active TURN allocations",
		}),
		AllocationRequests: promauto.NewCounter(prometheus.CounterOpts{
			Name: "turn_pool_allocation_requests_total",
			Help: "Total number of TURN allocation requests",
		}),
		AllocationFailures: promauto.NewCounter(prometheus.CounterOpts{
			Name: "turn_pool_allocation_failures_total",
			Help: "Total number of failed TURN allocation requests",
		}),
		HealthCheckTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "turn_pool_health_checks_total",
			Help: "Total number of health checks performed",
		}),
		HealthCheckFailed: promauto.NewCounter(prometheus.CounterOpts{
			Name: "turn_pool_health_checks_failed_total",
			Help: "Total number of failed health checks",
		}),
		AllocationDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "turn_pool_allocation_duration_seconds",
			Help:    "Duration of TURN allocation requests",
			Buckets: prometheus.DefBuckets,
		}),
		HealthCheckLatency: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "turn_pool_health_check_latency_seconds",
			Help:    "Latency of TURN server health checks",
			Buckets: prometheus.DefBuckets,
		}),
		ServerConnections: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "turn_server_connections",
			Help: "Number of active connections per TURN server",
		}, []string{"server", "pool", "region"}),
		ServerLatency: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "turn_server_latency_seconds",
			Help: "Average latency per TURN server",
		}, []string{"server", "pool", "region"}),
		ServerStatus: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "turn_server_status",
			Help: "Status of TURN server (1=healthy, 0=unhealthy)",
		}, []string{"server", "pool", "region"}),
	}
}

// CreatePool создает новый пул TURN серверов
func (tpm *TURNPoolManager) CreatePool(name string) error {
	tpm.poolsMutex.Lock()
	defer tpm.poolsMutex.Unlock()

	if _, exists := tpm.pools[name]; exists {
		return fmt.Errorf("pool %s already exists", name)
	}

	// Создаем logger для пула - используем существующий logger с конфигурацией
	poolLogger, err := logger.New(logger.Config{
		Level:  logger.ERROR,
		Output: "stdout",
		JSON:   false,
	})
	if err != nil {
		return fmt.Errorf("failed to create pool logger: %w", err)
	}

	pool := NewTURNPool(poolLogger.GetZapLogger())
	tpm.pools[name] = pool

	tpm.logger.Info("Created TURN pool", zap.String("name", name))
	return nil
}

// AddServerToPool добавляет сервер в указанный пул
func (tpm *TURNPoolManager) AddServerToPool(poolName string, server *TURNServerInfo) error {
	tpm.poolsMutex.RLock()
	pool, exists := tpm.pools[poolName]
	tpm.poolsMutex.RUnlock()

	if !exists {
		return fmt.Errorf("pool %s does not exist", poolName)
	}

	// Добавляем сервер в пул
	pool.mu.Lock()
	pool.Servers = append(pool.Servers, server)
	pool.mu.Unlock()

	// Обновляем метрики
	if tpm.metrics != nil {
		tpm.updateServerMetrics()
	}

	tpm.logger.Info("Added server to pool",
		zap.String("pool", poolName),
		zap.String("server", server.Address),
		zap.String("region", server.Region))

	return nil
}

// AllocateTURN создает TURN allocation с автоматическим выбором лучшего сервера
func (tpm *TURNPoolManager) AllocateTURN(clientID string, preferences *AllocationPreferences) (*TURNAllocation, error) {
	start := time.Now()

	if tpm.metrics != nil {
		tpm.metrics.AllocationRequests.Inc()
	}

	// Проверяем лимит allocations для клиента
	if err := tpm.checkAllocationLimit(clientID); err != nil {
		if tpm.metrics != nil {
			tpm.metrics.AllocationFailures.Inc()
		}
		return nil, err
	}

	// Выбираем лучший пул и сервер
	allocation, err := tpm.selectAndAllocate(clientID, preferences)
	if err != nil {
		if tpm.metrics != nil {
			tpm.metrics.AllocationFailures.Inc()
		}
		return nil, err
	}

	// Сохраняем allocation
	tpm.allocationsMutex.Lock()
	tpm.allocations[allocation.ID] = allocation
	tpm.allocationsMutex.Unlock()

	// Обновляем метрики
	if tpm.metrics != nil {
		tpm.metrics.AllocationDuration.Observe(time.Since(start).Seconds())
		tpm.updateAllocationMetrics()
	}

	tpm.logger.Info("TURN allocation created",
		zap.String("allocation_id", allocation.ID),
		zap.String("client_id", clientID),
		zap.String("server", allocation.Server.Address),
		zap.Duration("duration", time.Since(start)))

	return allocation, nil
}

// AllocationPreferences предпочтения для создания allocation
type AllocationPreferences struct {
	PreferredRegion   string            `json:"preferred_region,omitempty"`
	PreferredProvider string            `json:"preferred_provider,omitempty"`
	RequiredTags      map[string]string `json:"required_tags,omitempty"`
	MaxLatency        time.Duration     `json:"max_latency,omitempty"`
	PoolName          string            `json:"pool_name,omitempty"`
}

// selectAndAllocate выбирает лучший сервер и создает allocation
func (tpm *TURNPoolManager) selectAndAllocate(clientID string, preferences *AllocationPreferences) (*TURNAllocation, error) {
	poolName := tpm.config.DefaultPoolName
	if preferences != nil && preferences.PoolName != "" {
		poolName = preferences.PoolName
	}

	tpm.poolsMutex.RLock()
	pool, exists := tpm.pools[poolName]
	tpm.poolsMutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("pool %s not found", poolName)
	}

	// Выбираем лучший сервер из пула
	if len(pool.Servers) == 0 {
		return nil, fmt.Errorf("no servers available in pool %s", poolName)
	}

	// Выбираем первый доступный сервер
	server := pool.Servers[0]

	// Создаем allocation
	allocation := &TURNAllocation{
		ID:           tpm.generateAllocationID(),
		ClientID:     clientID,
		Server:       server,
		RelayAddress: "127.0.0.1", // Будет установлен реальный адрес
		RelayPort:    49152,       // Будет установлен реальный порт
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(tpm.config.AllocationTimeout),
		Lifetime:     tpm.config.AllocationTimeout,
		Bandwidth:    0,
		Permissions:  []string{},
		Channels:     []string{},
		Tags:         make(map[string]string),
		Status:       "active",
		LastActivity: time.Now(),
	}

	// Сохраняем allocation
	tpm.allocationsMutex.Lock()
	tpm.allocations[allocation.ID] = allocation
	tpm.allocationsMutex.Unlock()

	return allocation, nil
}

// checkAllocationLimit проверяет лимит allocations для клиента
func (tpm *TURNPoolManager) checkAllocationLimit(clientID string) error {
	tpm.allocationsMutex.RLock()
	defer tpm.allocationsMutex.RUnlock()

	count := 0
	for _, allocation := range tpm.allocations {
		if allocation.ClientID == clientID {
			count++
		}
	}

	if count >= tpm.config.MaxAllocationsPerClient {
		return fmt.Errorf("allocation limit exceeded for client %s: %d/%d",
			clientID, count, tpm.config.MaxAllocationsPerClient)
	}

	return nil
}

// ReleaseTURN освобождает TURN allocation
func (tpm *TURNPoolManager) ReleaseTURN(allocationID string) error {
	tpm.allocationsMutex.Lock()
	allocation, exists := tpm.allocations[allocationID]
	if exists {
		delete(tpm.allocations, allocationID)
	}
	tpm.allocationsMutex.Unlock()

	if !exists {
		return fmt.Errorf("allocation %s not found", allocationID)
	}

	// Allocation уже удален из основного хранилища

	// Обновляем метрики
	if tpm.metrics != nil {
		tpm.updateAllocationMetrics()
	}

	tpm.logger.Info("TURN allocation released",
		zap.String("allocation_id", allocationID),
		zap.String("client_id", allocation.ClientID))

	return nil
}

// findPoolForServer находит пул, содержащий указанный сервер
func (tpm *TURNPoolManager) findPoolForServer(serverAddress string) string {
	tpm.poolsMutex.RLock()
	defer tpm.poolsMutex.RUnlock()

	for poolName, pool := range tpm.pools {
		pool.mu.RLock()
		for _, server := range pool.Servers {
			if server.Address == serverAddress {
				pool.mu.RUnlock()
				return poolName
			}
		}
		pool.mu.RUnlock()
	}
	return ""
}

// GetPoolStats возвращает статистику всех пулов
func (tpm *TURNPoolManager) GetPoolStats() map[string]interface{} {
	tpm.poolsMutex.RLock()
	defer tpm.poolsMutex.RUnlock()

	stats := map[string]interface{}{
		"total_pools":       len(tpm.pools),
		"total_allocations": len(tpm.allocations),
		"pools":             make(map[string]interface{}),
	}

	totalServers := 0
	healthyServers := 0

	for poolName, pool := range tpm.pools {
		pool.mu.RLock()
		servers := pool.Servers
		pool.mu.RUnlock()

		poolStats := map[string]interface{}{
			"total_servers":   len(servers),
			"healthy_servers": 0, // Будет вычислено ниже
		}

		healthyCount := 0
		for _, server := range servers {
			if server.Status == "active" {
				healthyCount++
			}
		}
		poolStats["healthy_servers"] = healthyCount

		stats["pools"].(map[string]interface{})[poolName] = poolStats

		totalServers += len(servers)
		healthyServers += healthyCount
	}

	stats["total_servers"] = totalServers
	stats["healthy_servers"] = healthyServers

	return stats
}

// GetAllocationInfo возвращает информацию об allocation
func (tpm *TURNPoolManager) GetAllocationInfo(allocationID string) (*TURNAllocation, error) {
	tpm.allocationsMutex.RLock()
	defer tpm.allocationsMutex.RUnlock()

	allocation, exists := tpm.allocations[allocationID]
	if !exists {
		return nil, fmt.Errorf("allocation %s not found", allocationID)
	}

	return allocation, nil
}

// ListAllocations возвращает список всех активных allocations
func (tpm *TURNPoolManager) ListAllocations() []*TURNAllocation {
	tpm.allocationsMutex.RLock()
	defer tpm.allocationsMutex.RUnlock()

	allocations := make([]*TURNAllocation, 0, len(tpm.allocations))
	for _, allocation := range tpm.allocations {
		allocations = append(allocations, allocation)
	}

	return allocations
}

// startStatsCollector запускает сборщик статистики
func (tpm *TURNPoolManager) startStatsCollector() {
	// Проверяем, что интервал положительный
	if tpm.config.StatsInterval <= 0 {
		tpm.logger.Debug("Stats interval is zero or negative, skipping stats collector")
		return
	}

	ticker := time.NewTicker(tpm.config.StatsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-tpm.ctx.Done():
			return
		case <-ticker.C:
			tpm.updateMetrics()
		}
	}
}

// updateMetrics обновляет все метрики
func (tpm *TURNPoolManager) updateMetrics() {
	if tpm.metrics == nil {
		return
	}

	tpm.updateServerMetrics()
	tpm.updateAllocationMetrics()
}

// updateServerMetrics обновляет метрики серверов
func (tpm *TURNPoolManager) updateServerMetrics() {
	if tpm.metrics == nil {
		return
	}

	tpm.poolsMutex.RLock()
	defer tpm.poolsMutex.RUnlock()

	totalServers := 0
	healthyServers := 0

	for _, pool := range tpm.pools {
		pool.mu.RLock()
		servers := pool.Servers
		totalServers += len(servers)

		for _, server := range servers {
			if server.Status == "active" {
				healthyServers++
			}
		}
		pool.mu.RUnlock()
	}

	tpm.metrics.TotalServers.Set(float64(totalServers))
	tpm.metrics.HealthyServers.Set(float64(healthyServers))
}

// updateAllocationMetrics обновляет метрики allocations
func (tpm *TURNPoolManager) updateAllocationMetrics() {
	if tpm.metrics == nil {
		return
	}

	tpm.allocationsMutex.RLock()
	totalAllocations := len(tpm.allocations)

	// Считаем активные allocations (использованные в последний час)
	activeAllocations := 0
	now := time.Now()
	for _, allocation := range tpm.allocations {
		if now.Sub(allocation.LastActivity) < time.Hour {
			activeAllocations++
		}
	}
	tpm.allocationsMutex.RUnlock()

	tpm.metrics.TotalAllocations.Set(float64(totalAllocations))
	tpm.metrics.ActiveAllocations.Set(float64(activeAllocations))
}

// startAllocationCleanup запускает очистку устаревших allocations
func (tpm *TURNPoolManager) startAllocationCleanup() {
	// Используем фиксированный интервал в 1 минуту для очистки
	cleanupInterval := time.Minute
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-tpm.ctx.Done():
			return
		case <-ticker.C:
			tpm.cleanupAllocations()
		}
	}
}

// cleanupAllocations очищает устаревшие allocations
func (tpm *TURNPoolManager) cleanupAllocations() {
	tpm.allocationsMutex.Lock()
	defer tpm.allocationsMutex.Unlock()

	now := time.Now()
	cleaned := 0

	for allocID, allocation := range tpm.allocations {
		if now.Sub(allocation.LastActivity) > tpm.config.AllocationTimeout {
			delete(tpm.allocations, allocID)
			cleaned++

			tpm.logger.Debug("Cleaned up expired allocation",
				zap.String("allocation_id", allocID),
				zap.String("client_id", allocation.ClientID),
				zap.Duration("age", now.Sub(allocation.LastActivity)))
		}
	}

	if cleaned > 0 {
		tpm.logger.Info("Cleaned up expired allocations", zap.Int("count", cleaned))
	}
}

// generateAllocationID генерирует уникальный ID для allocation
func (tpm *TURNPoolManager) generateAllocationID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return "turn-" + hex.EncodeToString(bytes)
}

// Stop останавливает TURN pool manager
func (tpm *TURNPoolManager) Stop() {
	tpm.logger.Info("Stopping TURN Pool Manager...")

	// Останавливаем все пулы
	tpm.poolsMutex.RLock()
	for name := range tpm.pools {
		tpm.logger.Info("Stopping pool", zap.String("name", name))
	}
	tpm.poolsMutex.RUnlock()

	// Освобождаем все allocations
	tpm.allocationsMutex.Lock()
	allocCount := len(tpm.allocations)
	tpm.allocations = make(map[string]*TURNAllocation)
	tpm.allocationsMutex.Unlock()

	tpm.cancel()

	tpm.logger.Info("TURN Pool Manager stopped",
		zap.Int("released_allocations", allocCount))
}
