package ai

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/internal/ai/anomaly"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai/pipeline"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai/predictive"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai/routing"
	"go.uber.org/zap"
)

// Service represents the main AI service
type Service struct {
	config          *Config
	logger          *zap.Logger
	anomalyDetector *anomaly.NetworkAnomalyDetector
	loadPredictor   *predictive.LoadPredictor
	routeOptimizer  *routing.RouteOptimizer
	dataPipeline    *pipeline.DataPipeline
	monitoring      *AIMonitoring
	metrics         *AIMetrics

	// Service state
	started   bool
	startTime time.Time
	stopCh    chan struct{}
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewService creates a new AI service
func NewService(config *Config, logger *zap.Logger) *Service {
	ctx, cancel := context.WithCancel(context.Background())

	return &Service{
		config: config,
		logger: logger,
		stopCh: make(chan struct{}),
		ctx:    ctx,
		cancel: cancel,
	}
}

// Initialize initializes the AI service components
func (s *Service) Initialize() error {
	s.logger.Info("Initializing AI service...")

	// Initialize data pipeline
	pipelineConfig := &pipeline.Config{
		Enabled:            s.config.AI.DataPipeline.Enabled,
		BatchSize:          s.config.AI.DataPipeline.BatchSize,
		ProcessingInterval: s.config.AI.DataPipeline.ProcessingInterval,
		RetentionDays:      s.config.AI.DataPipeline.RetentionDays,
	}
	s.dataPipeline = pipeline.NewDataPipeline(pipelineConfig, s.logger)
	if err := s.dataPipeline.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize data pipeline: %w", err)
	}

	// Initialize anomaly detector
	if s.config.AI.Models.AnomalyDetection.Enabled {
		anomalyConfig := &anomaly.Config{
			Threshold:     s.config.AI.Models.AnomalyDetection.Threshold,
			LearningRate:  s.config.AI.Models.AnomalyDetection.LearningRate,
			MaxSamples:    s.config.AI.Models.AnomalyDetection.MaxSamples,
			Contamination: s.config.AI.Models.AnomalyDetection.Contamination,
		}
		s.anomalyDetector = anomaly.NewNetworkAnomalyDetector(anomalyConfig, s.logger)
		if err := s.anomalyDetector.Initialize(); err != nil {
			return fmt.Errorf("failed to initialize anomaly detector: %w", err)
		}
	}

	// Initialize load predictor
	if s.config.AI.Models.LoadPrediction.Enabled {
		predictiveConfig := &predictive.Config{
			Enabled:         s.config.AI.Models.LoadPrediction.Enabled,
			Algorithm:       s.config.AI.Models.LoadPrediction.Algorithm,
			ForecastHorizon: s.config.AI.Models.LoadPrediction.ForecastHorizon,
			ConfidenceLevel: s.config.AI.Models.LoadPrediction.ConfidenceLevel,
		}
		s.loadPredictor = predictive.NewLoadPredictor(predictiveConfig, s.logger)
		if err := s.loadPredictor.Initialize(); err != nil {
			return fmt.Errorf("failed to initialize load predictor: %w", err)
		}
	}

	// Initialize route optimizer
	if s.config.AI.Models.RouteOptimization.Enabled {
		routingConfig := &routing.Config{
			Enabled:         s.config.AI.Models.RouteOptimization.Enabled,
			Algorithm:       s.config.AI.Models.RouteOptimization.Algorithm,
			ExplorationRate: s.config.AI.Models.RouteOptimization.ExplorationRate,
			LearningRate:    s.config.AI.Models.RouteOptimization.LearningRate,
			UpdateInterval:  s.config.AI.Models.RouteOptimization.UpdateInterval,
		}
		s.routeOptimizer = routing.NewRouteOptimizer(routingConfig, s.logger)
		if err := s.routeOptimizer.Initialize(); err != nil {
			return fmt.Errorf("failed to initialize route optimizer: %w", err)
		}
	}

	// Initialize monitoring
	s.monitoring = NewAIMonitoring()
	s.metrics = NewAIMetrics()

	s.logger.Info("AI service initialized successfully")
	return nil
}

// Start starts the AI service
func (s *Service) Start() error {
	if s.started {
		return fmt.Errorf("AI service is already started")
	}

	s.logger.Info("Starting AI service...")

	// Start data pipeline
	if err := s.dataPipeline.Start(); err != nil {
		return fmt.Errorf("failed to start data pipeline: %w", err)
	}

	// Start anomaly detector
	if s.anomalyDetector != nil {
		if err := s.anomalyDetector.Start(); err != nil {
			return fmt.Errorf("failed to start anomaly detector: %w", err)
		}
	}

	// Start load predictor
	if s.loadPredictor != nil {
		if err := s.loadPredictor.Start(); err != nil {
			return fmt.Errorf("failed to start load predictor: %w", err)
		}
	}

	// Start route optimizer
	if s.routeOptimizer != nil {
		if err := s.routeOptimizer.Start(); err != nil {
			return fmt.Errorf("failed to start route optimizer: %w", err)
		}
	}

	// Start monitoring
	if err := s.monitoring.Start(); err != nil {
		return fmt.Errorf("failed to start monitoring: %w", err)
	}

	// Start background processing
	s.wg.Add(1)
	go s.backgroundProcessing()

	s.started = true
	s.startTime = time.Now()
	s.logger.Info("AI service started successfully")
	return nil
}

// Stop stops the AI service
func (s *Service) Stop() error {
	if !s.started {
		return nil
	}

	s.logger.Info("Stopping AI service...")

	// Cancel context
	s.cancel()

	// Stop data pipeline
	if s.dataPipeline != nil {
		if err := s.dataPipeline.Stop(); err != nil {
			s.logger.Error("Failed to stop data pipeline", zap.Error(err))
		}
	}

	// Stop anomaly detector
	if s.anomalyDetector != nil {
		if err := s.anomalyDetector.Stop(); err != nil {
			s.logger.Error("Failed to stop anomaly detector", zap.Error(err))
		}
	}

	// Stop load predictor
	if s.loadPredictor != nil {
		if err := s.loadPredictor.Stop(); err != nil {
			s.logger.Error("Failed to stop load predictor", zap.Error(err))
		}
	}

	// Stop route optimizer
	if s.routeOptimizer != nil {
		if err := s.routeOptimizer.Stop(); err != nil {
			s.logger.Error("Failed to stop route optimizer", zap.Error(err))
		}
	}

	// Stop monitoring
	if s.monitoring != nil {
		if err := s.monitoring.Stop(); err != nil {
			s.logger.Error("Failed to stop monitoring", zap.Error(err))
		}
	}

	// Wait for background processing to finish
	close(s.stopCh)
	s.wg.Wait()

	s.started = false
	s.logger.Info("AI service stopped successfully")
	return nil
}

// backgroundProcessing handles background AI processing
func (s *Service) backgroundProcessing() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.config.AI.DataPipeline.ProcessingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.processData()
		}
	}
}

// processData processes incoming data
func (s *Service) processData() {
	start := time.Now()

	// Collect data from pipeline
	data, err := s.dataPipeline.CollectData()
	if err != nil {
		s.logger.Error("Failed to collect data", zap.Error(err))
		return
	}

	// Process anomaly detection
	if s.anomalyDetector != nil {
		// Создаем простые метрики для тестирования
		metrics := &anomaly.NetworkMetrics{
			Timestamp:            time.Now(),
			ConnectionsPerSecond: 100,
			BytesPerSecond:       1000000,
			PacketLossRate:       0.01,
			AverageLatency:       50.0,
			CPUUsage:             45.0,
			MemoryUsage:          60.0,
			ActiveConnections:    150,
			FailedConnections:    5,
			ProtocolDistribution: map[string]int{
				"QUIC":      50,
				"WireGuard": 30,
				"TCP":       20,
			},
		}

		_, err := s.anomalyDetector.DetectAnomaly(metrics)
		if err != nil {
			s.logger.Error("Failed to detect anomalies", zap.Error(err))
		} else {
			s.metrics.RecordAnomalyDetections(1)
		}
	}

	// Process load prediction
	if s.loadPredictor != nil {
		prediction, err := s.loadPredictor.PredictLoad(24) // 24 часа
		if err != nil {
			s.logger.Error("Failed to predict load", zap.Error(err))
		} else {
			s.metrics.RecordLoadPredictions(1)
			s.logger.Debug("Load prediction completed", zap.Float64("confidence", prediction.Confidence))
		}
	}

	// Process route optimization
	if s.routeOptimizer != nil {
		_, err := s.routeOptimizer.OptimizeRoute("node1", "node2")
		if err != nil {
			s.logger.Error("Failed to optimize routes", zap.Error(err))
		} else {
			s.metrics.RecordRouteOptimizations(1)
		}
	}

	// Record processing time
	processingTime := time.Since(start)
	s.metrics.RecordProcessingTime(processingTime)

	s.logger.Debug("Data processing completed",
		zap.Duration("processing_time", processingTime),
		zap.Int("data_points", len(data)),
	)
}

// GetStatus returns the current status of the AI service
func (s *Service) GetStatus() *ServiceStatus {
	status := "stopped"
	if s.started {
		status = "running"
	}

	components := map[string]bool{
		"anomaly_detector": s.anomalyDetector != nil,
		"load_predictor":   s.loadPredictor != nil,
		"route_optimizer":  s.routeOptimizer != nil,
		"data_pipeline":    s.dataPipeline != nil,
		"monitoring":       s.monitoring != nil,
	}

	performance := map[string]interface{}{
		"metrics": s.metrics.GetMetrics(),
	}

	return &ServiceStatus{
		Status:      status,
		Components:  components,
		Performance: performance,
		Health:      s.started && s.anomalyDetector != nil && s.loadPredictor != nil && s.routeOptimizer != nil,
		Resources:   map[string]interface{}{}, // Заглушка для ресурсов
		Uptime:      time.Since(s.startTime),
		LastUpdate:  time.Now(),
	}
}

// ServiceStatus represents the status of the AI service
type ServiceStatus struct {
	Status      string                 `json:"status"`
	Components  map[string]bool        `json:"components"`
	Performance map[string]interface{} `json:"performance"`
	Health      bool                   `json:"health"`
	Resources   map[string]interface{} `json:"resources"`
	Uptime      time.Duration          `json:"uptime"`
	LastUpdate  time.Time              `json:"last_update"`
}
