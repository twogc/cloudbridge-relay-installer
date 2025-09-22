package ai

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai/anomaly"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai/cache"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai/predictive"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai/routing"
)

// AICircuitBreaker представляет circuit breaker для AI сервиса
type AICircuitBreaker struct {
	mu           sync.RWMutex
	state        string // "closed", "open", "half-open"
	failures     int
	threshold    int
	timeout      time.Duration
	lastFailTime time.Time
	logger       *zap.Logger
}

// NewAICircuitBreaker создает новый circuit breaker
func NewAICircuitBreaker(timeout time.Duration, logger *zap.Logger) *AICircuitBreaker {
	return &AICircuitBreaker{
		state:     "closed",
		threshold: 5,
		timeout:   timeout,
		logger:    logger,
	}
}

// CanExecute проверяет, можно ли выполнить запрос
func (cb *AICircuitBreaker) CanExecute() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	if cb.state == "closed" {
		return true
	}

	if cb.state == "open" {
		if time.Since(cb.lastFailTime) > cb.timeout {
			cb.state = "half-open"
			return true
		}
		return false
	}

	return true // half-open
}

// RecordSuccess записывает успешный запрос
func (cb *AICircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures = 0
	cb.state = "closed"
}

// RecordFailure записывает неудачный запрос
func (cb *AICircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.lastFailTime = time.Now()

	if cb.failures >= cb.threshold {
		cb.state = "open"
	}
}

// OptimizedService представляет оптимизированный AI сервис
type OptimizedService struct {
	mu              sync.RWMutex
	config          *Config
	logger          *zap.Logger
	anomalyDetector *anomaly.OptimizedDetector
	loadPredictor   *predictive.OptimizedPredictor
	routeOptimizer  *routing.OptimizedOptimizer
	cache           *cache.OptimizedCache
	metrics         *AIMetrics
	circuitBreaker  *AICircuitBreaker
	performance     *ServicePerformance
	healthChecker   *HealthChecker
	resourceManager *ResourceManager
	lastUpdate      time.Time
	updateInterval  time.Duration
	shutdownChan    chan struct{}
	wg              sync.WaitGroup
}

// ServicePerformance представляет производительность сервиса
type ServicePerformance struct {
	TotalRequests      int64
	SuccessfulRequests int64
	FailedRequests     int64
	AverageLatency     time.Duration
	PeakLatency        time.Duration
	Throughput         float64
	ErrorRate          float64
	LastUpdate         time.Time
	mu                 sync.RWMutex
}

// HealthChecker проверяет здоровье сервиса
type HealthChecker struct {
	checks    map[string]*HealthCheck
	status    map[string]bool
	lastCheck time.Time
	mu        sync.RWMutex
}

// HealthCheck представляет проверку здоровья
type HealthCheck struct {
	Name                string
	CheckFunc           func() error
	Interval            time.Duration
	LastCheck           time.Time
	LastError           error
	ConsecutiveFailures int
	Enabled             bool
}

// ResourceManager управляет ресурсами сервиса
type ResourceManager struct {
	CPUUsage    float64
	MemoryUsage float64
	DiskUsage   float64
	NetworkIO   float64
	Limits      *ResourceLimits
	mu          sync.RWMutex
}

// ResourceLimits представляет лимиты ресурсов
type ResourceLimits struct {
	MaxCPUUsage    float64
	MaxMemoryUsage float64
	MaxDiskUsage   float64
	MaxNetworkIO   float64
}

// NewOptimizedService создает новый оптимизированный AI сервис
func NewOptimizedService(config *Config, logger *zap.Logger) (*OptimizedService, error) {
	// Проверяем конфигурацию и logger
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	service := &OptimizedService{
		config:          config,
		logger:          logger,
		metrics:         NewAIMetrics(),
		performance:     &ServicePerformance{},
		healthChecker:   NewHealthChecker(),
		resourceManager: NewResourceManager(),
		updateInterval:  30 * time.Second,
		shutdownChan:    make(chan struct{}),
	}

	// Инициализируем компоненты
	if err := service.initializeComponents(); err != nil {
		return nil, err
	}

	// Инициализируем circuit breaker
	service.circuitBreaker = NewAICircuitBreaker(config.AI.RequestTimeout, logger)

	// Запускаем фоновые задачи
	service.startBackgroundTasks()

	service.logger.Info("Optimized AI service initialized successfully")
	return service, nil
}

// initializeComponents инициализирует компоненты сервиса
func (s *OptimizedService) initializeComponents() error {
	// Инициализируем детектор аномалий
	s.anomalyDetector = anomaly.NewOptimizedDetector(
		s.config.AI.Models.AnomalyDetection.Threshold,
		s.config.AI.Models.AnomalyDetection.LearningRate,
		s.config.AI.Models.AnomalyDetection.MaxSamples,
		s.config.AI.Models.AnomalyDetection.Contamination,
		s.logger,
	)

	// Инициализируем предиктор нагрузки
	s.loadPredictor = predictive.NewOptimizedPredictor(
		s.config.AI.Models.LoadPrediction.ForecastHorizon,
		s.config.AI.Models.LoadPrediction.ConfidenceLevel,
		s.logger,
	)

	// Инициализируем оптимизатор маршрутов
	s.routeOptimizer = routing.NewOptimizedOptimizer(
		s.config.AI.Models.RouteOptimization.ExplorationRate,
		s.config.AI.Models.RouteOptimization.LearningRate,
		0.1, // threshold
		s.logger,
	)

	// Инициализируем кэш
	s.cache = cache.NewOptimizedCache(
		s.config.AI.MaxConcurrentRequests,
		s.config.AI.CacheTTL,
		s.logger,
	)

	// Инициализируем проверки здоровья
	s.initializeHealthChecks()

	return nil
}

// initializeHealthChecks инициализирует проверки здоровья
func (s *OptimizedService) initializeHealthChecks() {
	s.healthChecker.AddCheck("anomaly_detector", func() error {
		if s.anomalyDetector == nil {
			return errors.New("anomaly detector not initialized")
		}
		return nil
	}, 30*time.Second)

	s.healthChecker.AddCheck("load_predictor", func() error {
		if s.loadPredictor == nil {
			return errors.New("load predictor not initialized")
		}
		return nil
	}, 30*time.Second)

	s.healthChecker.AddCheck("route_optimizer", func() error {
		if s.routeOptimizer == nil {
			return errors.New("route optimizer not initialized")
		}
		return nil
	}, 30*time.Second)

	s.healthChecker.AddCheck("cache", func() error {
		if s.cache == nil {
			return errors.New("cache not initialized")
		}
		return nil
	}, 30*time.Second)

	s.healthChecker.AddCheck("circuit_breaker", func() error {
		if s.circuitBreaker == nil {
			return errors.New("circuit breaker not initialized")
		}
		return nil
	}, 30*time.Second)
}

// startBackgroundTasks запускает фоновые задачи
func (s *OptimizedService) startBackgroundTasks() {
	// Запускаем обновление метрик
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.updateMetricsWorker()
	}()

	// Запускаем проверки здоровья
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.healthCheckWorker()
	}()

	// Запускаем мониторинг ресурсов
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.resourceMonitoringWorker()
	}()

	// Запускаем очистку кэша
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.cacheCleanupWorker()
	}()
}

// DetectAnomaly обнаруживает аномалии
func (s *OptimizedService) DetectAnomaly(ctx context.Context, metrics *anomaly.NetworkMetrics) (*anomaly.AnomalyResult, error) {
	start := time.Now()
	defer func() {
		s.recordRequest("detect_anomaly", time.Since(start), nil)
	}()

	// Проверяем входные параметры
	if metrics == nil {
		return nil, errors.New("metrics cannot be nil")
	}

	// Проверяем circuit breaker
	if !s.circuitBreaker.CanExecute() {
		return nil, errors.New("circuit breaker is open")
	}

	// Проверяем кэш
	cacheKey := s.generateCacheKey("anomaly", metrics)
	if cached, exists := s.cache.Get(cacheKey); exists {
		if result, ok := cached.(*anomaly.AnomalyResult); ok {
			s.logger.Debug("Anomaly detection result found in cache")
			return result, nil
		}
	}

	// Выполняем детекцию аномалий
	result, err := s.anomalyDetector.DetectAnomaly(ctx, metrics)
	if err != nil {
		s.circuitBreaker.RecordFailure()
		return nil, err
	}

	// Кэшируем результат
	s.cache.Set(cacheKey, result, 1024) // 1KB

	// Записываем метрики
	s.metrics.RecordAnomalyDetection(result.IsAnomaly, true)

	s.circuitBreaker.RecordSuccess()
	return result, nil
}

// PredictLoad предсказывает нагрузку
func (s *OptimizedService) PredictLoad(ctx context.Context, horizon int) (*predictive.LoadPrediction, error) {
	start := time.Now()
	defer func() {
		s.recordRequest("predict_load", time.Since(start), nil)
	}()

	// Проверяем входные параметры
	if horizon <= 0 {
		return nil, errors.New("horizon must be positive")
	}

	// Проверяем circuit breaker
	if !s.circuitBreaker.CanExecute() {
		return nil, errors.New("circuit breaker is open")
	}

	// Проверяем кэш
	cacheKey := s.generateCacheKey("load_prediction", horizon)
	if cached, exists := s.cache.Get(cacheKey); exists {
		if result, ok := cached.(*predictive.LoadPrediction); ok {
			s.logger.Debug("Load prediction result found in cache")
			return result, nil
		}
	}

	// Выполняем предсказание нагрузки
	result, err := s.loadPredictor.PredictLoad(ctx, horizon)
	if err != nil {
		s.circuitBreaker.RecordFailure()
		return nil, err
	}

	// Кэшируем результат
	s.cache.Set(cacheKey, result, 2048) // 2KB

	s.circuitBreaker.RecordSuccess()
	return result, nil
}

// OptimizeRoute оптимизирует маршрут
func (s *OptimizedService) OptimizeRoute(ctx context.Context, source, destination string, constraints map[string]interface{}) (*routing.Route, error) {
	start := time.Now()
	defer func() {
		s.recordRequest("optimize_route", time.Since(start), nil)
	}()

	// Проверяем circuit breaker
	if !s.circuitBreaker.CanExecute() {
		return nil, errors.New("circuit breaker is open")
	}

	// Проверяем кэш
	cacheKey := s.generateCacheKey("route_optimization", source, destination, constraints)
	if cached, exists := s.cache.Get(cacheKey); exists {
		if result, ok := cached.(*routing.Route); ok {
			s.logger.Debug("Route optimization result found in cache")
			return result, nil
		}
	}

	// Выполняем оптимизацию маршрута
	result, err := s.routeOptimizer.OptimizeRoute(ctx, source, destination, constraints)
	if err != nil {
		s.circuitBreaker.RecordFailure()
		return nil, err
	}

	// Кэшируем результат
	s.cache.Set(cacheKey, result, 1536) // 1.5KB

	// Записываем метрики
	s.metrics.RecordRouteOptimization(1)

	s.circuitBreaker.RecordSuccess()
	return result, nil
}

// GetStatus возвращает статус сервиса
func (s *OptimizedService) GetStatus() *ServiceStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Получаем статус компонентов
	anomalyStatus := s.anomalyDetector != nil
	loadStatus := s.loadPredictor != nil
	routeStatus := s.routeOptimizer != nil
	cacheStatus := s.cache != nil

	// Получаем метрики производительности
	performance := s.getPerformanceMetrics()

	// Получаем статус здоровья
	healthStatus := s.healthChecker.GetOverallStatus()

	// Получаем использование ресурсов
	resourceUsage := s.resourceManager.GetUsage()

	return &ServiceStatus{
		Status: "running",
		Components: map[string]bool{
			"anomaly_detector": anomalyStatus,
			"load_predictor":   loadStatus,
			"route_optimizer":  routeStatus,
			"cache":            cacheStatus,
		},
		Performance: performance,
		Health:      healthStatus,
		Resources:   resourceUsage,
		Uptime:      time.Since(s.lastUpdate),
		LastUpdate:  time.Now(),
	}
}

// GetMetrics возвращает метрики сервиса
func (s *OptimizedService) GetMetrics() *AIMetrics {
	return s.metrics
}

// HealthCheck выполняет проверку здоровья
func (s *OptimizedService) HealthCheck() error {
	// Проверяем все компоненты
	if s.anomalyDetector == nil {
		return errors.New("anomaly detector not available")
	}
	if s.loadPredictor == nil {
		return errors.New("load predictor not available")
	}
	if s.routeOptimizer == nil {
		return errors.New("route optimizer not available")
	}
	if s.cache == nil {
		return errors.New("cache not available")
	}

	// Проверяем circuit breaker
	if !s.circuitBreaker.CanExecute() {
		return errors.New("circuit breaker is open")
	}

	return nil
}

// Shutdown корректно завершает работу сервиса
func (s *OptimizedService) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down optimized AI service")

	// Сигнализируем о завершении
	close(s.shutdownChan)

	// Ждем завершения всех горутин
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	// Ждем завершения или таймаут
	select {
	case <-done:
		s.logger.Info("All background tasks completed")
	case <-ctx.Done():
		s.logger.Warn("Shutdown timeout exceeded")
		return ctx.Err()
	}

	s.logger.Info("Optimized AI service shutdown completed")
	return nil
}

// Вспомогательные методы

func (s *OptimizedService) generateCacheKey(operation string, params ...interface{}) string {
	// Упрощенная генерация ключа кэша
	key := operation
	for _, param := range params {
		key += "_" + fmt.Sprintf("%v", param)
	}
	return key
}

func (s *OptimizedService) recordRequest(operation string, duration time.Duration, err error) {
	s.performance.mu.Lock()
	defer s.performance.mu.Unlock()

	s.performance.TotalRequests++
	if err != nil {
		s.performance.FailedRequests++
	} else {
		s.performance.SuccessfulRequests++
	}

	// Обновляем среднюю латентность
	s.performance.AverageLatency = (s.performance.AverageLatency*time.Duration(s.performance.TotalRequests-1) + duration) / time.Duration(s.performance.TotalRequests)

	// Обновляем пиковую латентность
	if duration > s.performance.PeakLatency {
		s.performance.PeakLatency = duration
	}

	// Вычисляем throughput (запросов в секунду)
	s.performance.Throughput = float64(s.performance.SuccessfulRequests) / time.Since(s.performance.LastUpdate).Seconds()

	// Вычисляем error rate
	s.performance.ErrorRate = float64(s.performance.FailedRequests) / float64(s.performance.TotalRequests)

	s.performance.LastUpdate = time.Now()
}

func (s *OptimizedService) getPerformanceMetrics() map[string]interface{} {
	s.performance.mu.RLock()
	defer s.performance.mu.RUnlock()

	return map[string]interface{}{
		"total_requests":      s.performance.TotalRequests,
		"successful_requests": s.performance.SuccessfulRequests,
		"failed_requests":     s.performance.FailedRequests,
		"average_latency":     s.performance.AverageLatency.Milliseconds(),
		"peak_latency":        s.performance.PeakLatency.Milliseconds(),
		"throughput":          s.performance.Throughput,
		"error_rate":          s.performance.ErrorRate,
	}
}

// Фоновые задачи

func (s *OptimizedService) updateMetricsWorker() {
	ticker := time.NewTicker(s.updateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.updateMetrics()
		case <-s.shutdownChan:
			return
		}
	}
}

func (s *OptimizedService) healthCheckWorker() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.healthChecker.RunChecks()
		case <-s.shutdownChan:
			return
		}
	}
}

func (s *OptimizedService) resourceMonitoringWorker() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.resourceManager.UpdateUsage()
		case <-s.shutdownChan:
			return
		}
	}
}

func (s *OptimizedService) cacheCleanupWorker() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Кэш сам управляет очисткой, но можем добавить дополнительную логику
			s.logger.Debug("Cache cleanup completed", zap.Int("size", s.cache.Size()))
		case <-s.shutdownChan:
			return
		}
	}
}

func (s *OptimizedService) updateMetrics() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.lastUpdate = time.Now()

	// Обновляем метрики компонентов
	// В реальной реализации здесь была бы логика обновления метрик

	s.logger.Debug("Service metrics updated",
		zap.Time("last_update", s.lastUpdate),
	)
}

// Конструкторы для вспомогательных структур

func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		checks: make(map[string]*HealthCheck),
		status: make(map[string]bool),
	}
}

func NewResourceManager() *ResourceManager {
	return &ResourceManager{
		Limits: &ResourceLimits{
			MaxCPUUsage:    80.0,
			MaxMemoryUsage: 80.0,
			MaxDiskUsage:   90.0,
			MaxNetworkIO:   1000.0, // MB/s
		},
	}
}

// Методы для HealthChecker

func (hc *HealthChecker) AddCheck(name string, checkFunc func() error, interval time.Duration) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	hc.checks[name] = &HealthCheck{
		Name:      name,
		CheckFunc: checkFunc,
		Interval:  interval,
		Enabled:   true,
	}
}

func (hc *HealthChecker) RunChecks() {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	for name, check := range hc.checks {
		if !check.Enabled {
			continue
		}

		if time.Since(check.LastCheck) < check.Interval {
			continue
		}

		err := check.CheckFunc()
		check.LastCheck = time.Now()
		check.LastError = err

		if err != nil {
			check.ConsecutiveFailures++
			hc.status[name] = false
		} else {
			check.ConsecutiveFailures = 0
			hc.status[name] = true
		}
	}

	hc.lastCheck = time.Now()
}

func (hc *HealthChecker) GetOverallStatus() bool {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	for _, status := range hc.status {
		if !status {
			return false
		}
	}
	return true
}

// Методы для ResourceManager

func (rm *ResourceManager) UpdateUsage() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// CPU usage (overall average across cores for 0-second interval)
	if cpuPercents, err := cpu.Percent(0, true); err == nil && len(cpuPercents) > 0 {
		total := 0.0
		for _, p := range cpuPercents {
			total += p
		}
		rm.CPUUsage = total / float64(len(cpuPercents))
	} else if err != nil {
		// Log error but don't fail
		rm.CPUUsage = 0.0
	}

	// Memory usage
	if vm, err := mem.VirtualMemory(); err == nil && vm != nil {
		rm.MemoryUsage = vm.UsedPercent
	} else if err != nil {
		// Log error but don't fail
		rm.MemoryUsage = 0.0
	}

	// Disk usage (root mountpoint)
	if du, err := disk.Usage("/"); err == nil && du != nil {
		rm.DiskUsage = du.UsedPercent
	} else if err != nil {
		// Log error but don't fail
		rm.DiskUsage = 0.0
	}

	// Network IO (sum of bytes counters across interfaces)
	if ioStats, err := net.IOCounters(true); err == nil {
		var bytesTotal uint64
		for _, s := range ioStats {
			// Skip loopback
			if s.Name == "lo" || s.Name == "lo0" {
				continue
			}
			bytesTotal += s.BytesRecv + s.BytesSent
		}
		// Store as MB
		rm.NetworkIO = float64(bytesTotal) / (1024.0 * 1024.0)
	}
}

func (rm *ResourceManager) GetUsage() map[string]interface{} {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	return map[string]interface{}{
		"cpu_usage":    rm.CPUUsage,
		"memory_usage": rm.MemoryUsage,
		"disk_usage":   rm.DiskUsage,
		"network_io":   rm.NetworkIO,
		"limits":       rm.Limits,
	}
}
