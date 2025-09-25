package pipeline

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// DataPipeline обрабатывает данные для AI системы
type DataPipeline struct {
	config     *Config
	logger     *zap.Logger
	collectors map[string]DataCollector
	processors []DataProcessor
	storage    DataStorage

	// State
	started bool
	stopCh  chan struct{}
	wg      sync.WaitGroup
	ctx     context.Context
	cancel  context.CancelFunc
}

// Config конфигурация data pipeline
type Config struct {
	Enabled            bool          `yaml:"enabled"`
	BatchSize          int           `yaml:"batch_size"`
	ProcessingInterval time.Duration `yaml:"processing_interval"`
	RetentionDays      int           `yaml:"retention_days"`
	StoragePath        string        `yaml:"storage_path"`
	PrometheusURL      string        `yaml:"prometheus_url"`
}

// DataCollector интерфейс для сбора данных
type DataCollector interface {
	Collect() ([]DataPoint, error)
	Start() error
	Stop() error
}

// DataProcessor интерфейс для обработки данных
type DataProcessor interface {
	Process(data []DataPoint) ([]DataPoint, error)
}

// DataStorage интерфейс для хранения данных
type DataStorage interface {
	Store(data []DataPoint) error
	Retrieve(timeRange TimeRange) ([]DataPoint, error)
	Query(timeRange TimeRange, dataType string, limit int) ([]DataPoint, error)
}

// DataPoint представляет точку данных
type DataPoint struct {
	Timestamp time.Time              `json:"timestamp"`
	Source    string                 `json:"source"`
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
}

// TimeRange представляет временной диапазон
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// NewDataPipeline создает новый data pipeline
func NewDataPipeline(config *Config, logger *zap.Logger) *DataPipeline {
	ctx, cancel := context.WithCancel(context.Background())

	return &DataPipeline{
		config:     config,
		logger:     logger,
		collectors: make(map[string]DataCollector),
		processors: make([]DataProcessor, 0),
		stopCh:     make(chan struct{}),
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Initialize инициализирует data pipeline
func (dp *DataPipeline) Initialize() error {
	dp.logger.Info("Initializing Data Pipeline...")

	// Инициализируем коллекторы
	if err := dp.initializeCollectors(); err != nil {
		return fmt.Errorf("failed to initialize collectors: %w", err)
	}

	// Инициализируем процессоры
	if err := dp.initializeProcessors(); err != nil {
		return fmt.Errorf("failed to initialize processors: %w", err)
	}

	// Инициализируем хранилище
	if err := dp.initializeStorage(); err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	dp.logger.Info("Data Pipeline initialized successfully")
	return nil
}

// Start запускает data pipeline
func (dp *DataPipeline) Start() error {
	if dp.started {
		return fmt.Errorf("data pipeline is already started")
	}

	dp.logger.Info("Starting Data Pipeline...")

	// Запускаем коллекторы
	for name, collector := range dp.collectors {
		if err := collector.Start(); err != nil {
			return fmt.Errorf("failed to start collector %s: %w", name, err)
		}
	}

	// Запускаем обработку данных
	dp.wg.Add(1)
	go dp.dataProcessing()

	dp.started = true
	dp.logger.Info("Data Pipeline started successfully")
	return nil
}

// Stop останавливает data pipeline
func (dp *DataPipeline) Stop() error {
	if !dp.started {
		return nil
	}

	dp.logger.Info("Stopping Data Pipeline...")

	// Останавливаем коллекторы
	for name, collector := range dp.collectors {
		if err := collector.Stop(); err != nil {
			dp.logger.Error("Failed to stop collector", zap.String("name", name), zap.Error(err))
		}
	}

	// Останавливаем обработку данных
	dp.cancel()
	select {
	case <-dp.stopCh:
		// Канал уже закрыт
	default:
		close(dp.stopCh)
	}
	dp.wg.Wait()

	dp.started = false
	dp.logger.Info("Data Pipeline stopped successfully")
	return nil
}

// CollectData собирает данные из всех коллекторов
func (dp *DataPipeline) CollectData() ([]DataPoint, error) {
	var allData []DataPoint

	for name, collector := range dp.collectors {
		data, err := collector.Collect()
		if err != nil {
			dp.logger.Error("Failed to collect data", zap.String("collector", name), zap.Error(err))
			continue
		}

		allData = append(allData, data...)
	}

	return allData, nil
}

// dataProcessing обрабатывает данные в фоновом режиме
func (dp *DataPipeline) dataProcessing() {
	defer dp.wg.Done()

	ticker := time.NewTicker(dp.config.ProcessingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-dp.ctx.Done():
			return
		case <-dp.stopCh:
			return
		case <-ticker.C:
			dp.processData()
		}
	}
}

// processData обрабатывает собранные данные
func (dp *DataPipeline) processData() {
	start := time.Now()

	// Собираем данные
	data, err := dp.CollectData()
	if err != nil {
		dp.logger.Error("Failed to collect data", zap.Error(err))
		return
	}

	if len(data) == 0 {
		return
	}

	// Обрабатываем данные
	processedData := data
	for _, processor := range dp.processors {
		processedData, err = processor.Process(processedData)
		if err != nil {
			dp.logger.Error("Failed to process data", zap.Error(err))
			return
		}
	}

	// Сохраняем данные
	if err := dp.storage.Store(processedData); err != nil {
		dp.logger.Error("Failed to store data", zap.Error(err))
		return
	}

	processingTime := time.Since(start)
	dp.logger.Debug("Data processing completed",
		zap.Int("data_points", len(data)),
		zap.Duration("processing_time", processingTime))
}

// initializeCollectors инициализирует коллекторы данных
func (dp *DataPipeline) initializeCollectors() error {
	// Real Prometheus коллектор
	prometheusURL := dp.config.PrometheusURL
	if prometheusURL == "" {
		prometheusURL = "http://prometheus.cloudbridge.svc.cluster.local:9090"
	}
	prometheusCollector := NewRealPrometheusCollector(dp.logger, prometheusURL)
	dp.collectors["prometheus"] = prometheusCollector

	// Real System коллектор
	systemCollector := NewRealSystemCollector(dp.logger)
	dp.collectors["system"] = systemCollector

	// Real Network коллектор
	networkCollector := NewRealNetworkCollector(dp.logger)
	dp.collectors["network"] = networkCollector

	// Relay Metrics коллектор
	relayCollector := NewRelayMetricsCollector(dp.logger)
	dp.collectors["relay"] = relayCollector

	dp.logger.Info("Collectors initialized",
		zap.Int("count", len(dp.collectors)),
		zap.String("prometheus_url", prometheusURL))
	return nil
}

// initializeProcessors инициализирует процессоры данных
func (dp *DataPipeline) initializeProcessors() error {
	// Real Feature extractor
	featureExtractor := NewRealFeatureExtractor(dp.logger)
	dp.processors = append(dp.processors, featureExtractor)

	// Real Data normalizer
	normalizer := NewRealDataNormalizer(dp.logger)
	dp.processors = append(dp.processors, normalizer)

	// Data aggregator (5-минутные окна)
	aggregator := NewDataAggregator(dp.logger, 5*time.Minute)
	dp.processors = append(dp.processors, aggregator)

	dp.logger.Info("Processors initialized", zap.Int("count", len(dp.processors)))
	return nil
}

// initializeStorage инициализирует хранилище данных
func (dp *DataPipeline) initializeStorage() error {
	// Hybrid storage (file + memory)
	dp.storage = NewHybridStorage(dp.config, dp.logger)

	dp.logger.Info("Storage initialized",
		zap.String("storage_path", dp.config.StoragePath))
	return nil
}

// PrometheusCollector собирает данные из Prometheus
type PrometheusCollector struct {
	logger *zap.Logger
}

// NewPrometheusCollector создает новый Prometheus коллектор
func NewPrometheusCollector(logger *zap.Logger) *PrometheusCollector {
	return &PrometheusCollector{
		logger: logger,
	}
}

// Collect собирает данные из Prometheus
func (pc *PrometheusCollector) Collect() ([]DataPoint, error) {
	// Используйте RealPrometheusCollector через initializeCollectors
	return []DataPoint{}, nil
}

// Start запускает коллектор
func (pc *PrometheusCollector) Start() error {
	pc.logger.Info("Prometheus collector started")
	return nil
}

// Stop останавливает коллектор
func (pc *PrometheusCollector) Stop() error {
	pc.logger.Info("Prometheus collector stopped")
	return nil
}

// SystemCollector собирает системные данные
type SystemCollector struct {
	logger *zap.Logger
}

// NewSystemCollector создает новый системный коллектор
func NewSystemCollector(logger *zap.Logger) *SystemCollector {
	return &SystemCollector{
		logger: logger,
	}
}

// Collect собирает системные данные
func (sc *SystemCollector) Collect() ([]DataPoint, error) {
	// Используйте RealSystemCollector через initializeCollectors
	return []DataPoint{}, nil
}

// Start запускает коллектор
func (sc *SystemCollector) Start() error {
	sc.logger.Info("System collector started")
	return nil
}

// Stop останавливает коллектор
func (sc *SystemCollector) Stop() error {
	sc.logger.Info("System collector stopped")
	return nil
}

// NetworkCollector собирает сетевые данные
type NetworkCollector struct {
	logger *zap.Logger
}

// NewNetworkCollector создает новый сетевой коллектор
func NewNetworkCollector(logger *zap.Logger) *NetworkCollector {
	return &NetworkCollector{
		logger: logger,
	}
}

// Collect собирает сетевые данные
func (nc *NetworkCollector) Collect() ([]DataPoint, error) {
	// Используйте RealNetworkCollector через initializeCollectors
	return []DataPoint{}, nil
}

// Start запускает коллектор
func (nc *NetworkCollector) Start() error {
	nc.logger.Info("Network collector started")
	return nil
}

// Stop останавливает коллектор
func (nc *NetworkCollector) Stop() error {
	nc.logger.Info("Network collector stopped")
	return nil
}

// FeatureExtractor извлекает признаки из данных
type FeatureExtractor struct {
	logger *zap.Logger
}

// NewFeatureExtractor создает новый извлекатель признаков
func NewFeatureExtractor(logger *zap.Logger) *FeatureExtractor {
	return &FeatureExtractor{
		logger: logger,
	}
}

// Process обрабатывает данные и извлекает признаки
func (fe *FeatureExtractor) Process(data []DataPoint) ([]DataPoint, error) {
	// Используйте RealFeatureExtractor через initializeProcessors
	return data, nil
}

// extractFeatures извлекает признаки из данных
func (fe *FeatureExtractor) extractFeatures(data map[string]interface{}) map[string]interface{} {
	features := make(map[string]interface{})

	// Извлекаем числовые признаки
	for key, value := range data {
		if num, ok := value.(float64); ok {
			features[key] = num
		} else if num, ok := value.(int); ok {
			features[key] = float64(num)
		}
	}

	return features
}

// DataNormalizer нормализует данные
type DataNormalizer struct {
	logger *zap.Logger
}

// NewDataNormalizer создает новый нормализатор данных
func NewDataNormalizer(logger *zap.Logger) *DataNormalizer {
	return &DataNormalizer{
		logger: logger,
	}
}

// Process нормализует данные
func (dn *DataNormalizer) Process(data []DataPoint) ([]DataPoint, error) {
	// Используйте RealDataNormalizer через initializeProcessors
	return data, nil
}

// normalizeData нормализует данные
func (dn *DataNormalizer) normalizeData(data map[string]interface{}) map[string]interface{} {
	normalized := make(map[string]interface{})

	for key, value := range data {
		if num, ok := value.(float64); ok {
			// Простая нормализация в диапазон [0, 1]
			normalized[key] = num / 100.0
		} else {
			normalized[key] = value
		}
	}

	return normalized
}

// FileStorage хранит данные в файлах
type FileStorage struct {
	config *Config
	logger *zap.Logger
}

// NewFileStorage создает новое файловое хранилище
func NewFileStorage(config *Config, logger *zap.Logger) *FileStorage {
	return &FileStorage{
		config: config,
		logger: logger,
	}
}

// Store сохраняет данные
func (fs *FileStorage) Store(data []DataPoint) error {
	// Используйте HybridStorage/RealFileStorage в initializeStorage
	return nil
}

// Retrieve извлекает данные
func (fs *FileStorage) Retrieve(timeRange TimeRange) ([]DataPoint, error) {
	// Используйте HybridStorage/RealFileStorage в initializeStorage
	return []DataPoint{}, nil
}

// GetCollector возвращает коллектор по имени
func (dp *DataPipeline) GetCollector(name string) (DataCollector, bool) {
	collector, exists := dp.collectors[name]
	return collector, exists
}

// QueryData запрашивает данные из pipeline
func (dp *DataPipeline) QueryData(timeRange TimeRange, dataType string, limit int) ([]DataPoint, error) {
	if !dp.started {
		return nil, fmt.Errorf("data pipeline not started")
	}

	// Получаем данные из storage
	if dp.storage == nil {
		return nil, fmt.Errorf("storage not available")
	}

	// Запрашиваем данные из storage
	data, err := dp.storage.Query(timeRange, dataType, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query data from storage: %w", err)
	}

	return data, nil
}

// GetMetrics возвращает метрики pipeline
func (dp *DataPipeline) GetMetrics() (map[string]interface{}, error) {
	metrics := make(map[string]interface{})

	// Метрики коллекторов
	collectorMetrics := make(map[string]interface{})
	for name := range dp.collectors {
		collectorMetrics[name] = map[string]interface{}{
			"status": "running",
			"type":   "collector",
		}
	}
	metrics["collectors"] = collectorMetrics

	// Метрики процессоров
	processorMetrics := make(map[string]interface{})
	for i := range dp.processors {
		processorMetrics[fmt.Sprintf("processor_%d", i)] = map[string]interface{}{
			"status": "running",
			"type":   "processor",
		}
	}
	metrics["processors"] = processorMetrics

	// Метрики хранилища
	metrics["storage"] = map[string]interface{}{
		"status": "running",
		"type":   "storage",
	}

	// Общие метрики
	metrics["pipeline_status"] = "healthy"
	if !dp.started {
		metrics["pipeline_status"] = "stopped"
	}

	return metrics, nil
}
