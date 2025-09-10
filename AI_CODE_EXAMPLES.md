# 💻 Примеры кода для AI интеграции

## 🎯 Демонстрационные примеры

### **1. Network Anomaly Detection**

```go
// internal/ai/anomaly/detector.go
package anomaly

import (
    "github.com/sjwhitworth/golearn/base"
    "github.com/sjwhitworth/golearn/ensemble"
    "time"
)

type NetworkAnomalyDetector struct {
    isolationForest *ensemble.IsolationForest
    threshold       float64
    metrics         *metrics.AIMetrics
}

func NewNetworkAnomalyDetector(threshold float64) *NetworkAnomalyDetector {
    return &NetworkAnomalyDetector{
        isolationForest: ensemble.NewIsolationForest(100, 0.1),
        threshold:       threshold,
        metrics:         metrics.NewAIMetrics(),
    }
}

func (nad *NetworkAnomalyDetector) DetectAnomaly(metrics *NetworkMetrics) (*AnomalyResult, error) {
    // Подготовка данных
    features := nad.extractFeatures(metrics)
    
    // Обнаружение аномалий
    score := nad.isolationForest.Predict(features)
    
    isAnomaly := score > nad.threshold
    
    // Обновление метрик
    if isAnomaly {
        nad.metrics.IncrementAnomalyDetections("network", "high")
    }
    
    return &AnomalyResult{
        Score:     score,
        IsAnomaly: isAnomaly,
        Timestamp: time.Now(),
        Features:  features,
    }, nil
}

func (nad *NetworkAnomalyDetector) extractFeatures(metrics *NetworkMetrics) []float64 {
    return []float64{
        float64(metrics.ConnectionsPerSecond),
        float64(metrics.BytesPerSecond),
        float64(metrics.PacketLossRate),
        float64(metrics.AverageLatency),
        float64(metrics.CPUUsage),
        float64(metrics.MemoryUsage),
    }
}
```

### **2. Predictive Load Analytics**

```go
// internal/ai/predictive/load_predictor.go
package predictive

import (
    "github.com/gonum/stat"
    "time"
)

type LoadPredictor struct {
    history    *TimeSeriesBuffer
    arimaModel *ARIMAModel
    metrics    *metrics.AIMetrics
}

type LoadPrediction struct {
    Forecast    []float64 `json:"forecast"`
    Confidence  float64   `json:"confidence"`
    Horizon     int       `json:"horizon"`
    Timestamp   time.Time `json:"timestamp"`
}

func (lp *LoadPredictor) PredictLoad(horizon int) (*LoadPrediction, error) {
    // Получение исторических данных
    historicalData := lp.history.GetLast(24 * 7) // 7 дней
    
    // ARIMA прогнозирование
    forecast := lp.arimaModel.Forecast(historicalData, horizon)
    
    // Расчет доверительного интервала
    confidence := lp.calculateConfidence(forecast, historicalData)
    
    // Обновление метрик
    lp.metrics.UpdatePredictionAccuracy("load", "v1.0", confidence)
    
    return &LoadPrediction{
        Forecast:   forecast,
        Confidence: confidence,
        Horizon:    horizon,
        Timestamp:  time.Now(),
    }, nil
}

func (lp *LoadPredictor) calculateConfidence(forecast []float64, historical []float64) float64 {
    // Простой расчет доверительного интервала
    mean, _ := stat.Mean(historical, nil)
    stdDev, _ := stat.StdDev(historical, nil)
    
    // Чем меньше стандартное отклонение, тем выше уверенность
    confidence := 1.0 - (stdDev / mean)
    if confidence < 0 {
        confidence = 0
    }
    if confidence > 1 {
        confidence = 1
    }
    
    return confidence
}
```

### **3. Intelligent Routing**

```go
// internal/ai/routing/optimizer.go
package routing

import (
    "math/rand"
    "time"
)

type P2PRouteOptimizer struct {
    multiArmedBandit *MultiArmedBandit
    latencyPredictor *LatencyPredictor
    metrics          *metrics.AIMetrics
}

type OptimizedRoute struct {
    Protocol         string    `json:"protocol"`
    Path             []string  `json:"path"`
    PredictedLatency float64   `json:"predicted_latency"`
    Confidence       float64   `json:"confidence"`
    Timestamp        time.Time `json:"timestamp"`
}

func (pro *P2PRouteOptimizer) OptimizeRoute(source, destination string) (*OptimizedRoute, error) {
    // Выбор оптимального протокола
    protocol := pro.multiArmedBandit.Select(source, destination)
    
    // Предсказание латентности
    predictedLatency := pro.latencyPredictor.Predict(source, destination, protocol)
    
    // Поиск оптимального пути
    path := pro.findOptimalPath(source, destination, protocol)
    
    // Расчет уверенности
    confidence := pro.calculateConfidence(protocol, predictedLatency)
    
    // Обновление метрик
    pro.metrics.ObserveRouteOptimization("p2p", time.Since(time.Now()).Seconds())
    
    return &OptimizedRoute{
        Protocol:         protocol,
        Path:             path,
        PredictedLatency: predictedLatency,
        Confidence:       confidence,
        Timestamp:        time.Now(),
    }, nil
}

type MultiArmedBandit struct {
    arms map[string]*Arm
}

type Arm struct {
    Count    int     `json:"count"`
    Reward   float64 `json:"reward"`
    Protocol string  `json:"protocol"`
}

func (mab *MultiArmedBandit) Select(source, destination string) string {
    // UCB1 алгоритм для выбора протокола
    bestArm := ""
    bestValue := -1.0
    
    for protocol, arm := range mab.arms {
        if arm.Count == 0 {
            return protocol // Выбираем неиспользованный протокол
        }
        
        // UCB1 формула
        value := arm.Reward + math.Sqrt(2*math.Log(float64(mab.getTotalCount()))/float64(arm.Count))
        
        if value > bestValue {
            bestValue = value
            bestArm = protocol
        }
    }
    
    return bestArm
}
```

### **4. Enhanced AI Service**

```go
// cmd/ai-service/enhanced_main.go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/twogc/cloudbridge-relay-installer/internal/ai"
    "github.com/twogc/cloudbridge-relay-installer/internal/metrics"
)

type EnhancedAIService struct {
    anomalyDetector    *ai.NetworkAnomalyDetector
    loadPredictor      *ai.LoadPredictor
    routeOptimizer     *ai.P2PRouteOptimizer
    securityAnalyzer   *ai.SecurityIntelligence
    metrics            *metrics.AIMetrics
}

func main() {
    // Инициализация AI сервиса
    aiService := &EnhancedAIService{
        anomalyDetector:  ai.NewNetworkAnomalyDetector(0.8),
        loadPredictor:    ai.NewLoadPredictor(),
        routeOptimizer:   ai.NewP2PRouteOptimizer(),
        securityAnalyzer: ai.NewSecurityIntelligence(),
        metrics:          metrics.NewAIMetrics(),
    }
    
    // Настройка Gin router
    r := gin.Default()
    
    // AI API endpoints
    api := r.Group("/api/v1/ai")
    {
        // Anomaly Detection
        api.POST("/anomaly/detect", aiService.DetectAnomaly)
        api.GET("/anomaly/status", aiService.GetAnomalyStatus)
        
        // Predictive Analytics
        api.POST("/predict/load", aiService.PredictLoad)
        api.POST("/predict/failure", aiService.PredictFailure)
        
        // Intelligent Routing
        api.POST("/route/optimize", aiService.OptimizeRoute)
        api.GET("/route/recommendations", aiService.GetRouteRecommendations)
        
        // Security Intelligence
        api.POST("/security/analyze", aiService.AnalyzeSecurity)
        api.GET("/security/threats", aiService.GetActiveThreats)
        
        // Model Management
        api.GET("/models", aiService.ListModels)
        api.POST("/models/:id/train", aiService.TrainModel)
        api.POST("/models/:id/deploy", aiService.DeployModel)
    }
    
    // Запуск сервера
    r.Run(":8083")
}

func (s *EnhancedAIService) DetectAnomaly(c *gin.Context) {
    var req AnomalyDetectionRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    result, err := s.anomalyDetector.DetectAnomaly(req.Metrics)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(200, result)
}
```

### **5. Data Pipeline**

```go
// internal/ai/pipeline/data_pipeline.go
package pipeline

import (
    "github.com/segmentio/kafka-go"
    "time"
)

type DataPipeline struct {
    collectors  map[string]DataCollector
    processors  []DataProcessor
    storage     DataStorage
    kafkaWriter *kafka.Writer
}

type NetworkMetrics struct {
    Timestamp           time.Time `json:"timestamp"`
    ConnectionsPerSecond int      `json:"connections_per_second"`
    BytesPerSecond      int64     `json:"bytes_per_second"`
    PacketLossRate      float64   `json:"packet_loss_rate"`
    AverageLatency      float64   `json:"average_latency"`
    CPUUsage            float64   `json:"cpu_usage"`
    MemoryUsage         float64   `json:"memory_usage"`
}

func (dp *DataPipeline) ProcessMetrics(metrics *NetworkMetrics) error {
    // 1. Сбор данных
    rawData := dp.collect(metrics)
    
    // 2. Предобработка
    processedData := dp.preprocess(rawData)
    
    // 3. Feature extraction
    features := dp.extractFeatures(processedData)
    
    // 4. Отправка в Kafka для real-time обработки
    if err := dp.sendToKafka(features); err != nil {
        return err
    }
    
    // 5. Сохранение в storage
    return dp.storage.Store(features)
}

func (dp *DataPipeline) extractFeatures(data *NetworkMetrics) []float64 {
    return []float64{
        float64(data.ConnectionsPerSecond),
        float64(data.BytesPerSecond),
        data.PacketLossRate,
        data.AverageLatency,
        data.CPUUsage,
        data.MemoryUsage,
    }
}
```

### **6. Integration with Main Server**

```go
// cmd/relay/ai_integration.go
package main

import (
    "github.com/twogc/cloudbridge-relay-installer/internal/ai"
    "github.com/twogc/cloudbridge-relay-installer/internal/metrics"
)

func (s *Server) initializeAIComponents() error {
    if !s.config.AI.Enabled {
        return nil
    }
    
    // Инициализация AI компонентов
    s.aiAnomalyDetector = ai.NewNetworkAnomalyDetector(s.config.AI.AnomalyThreshold)
    s.aiLoadPredictor = ai.NewLoadPredictor()
    s.aiRouteOptimizer = ai.NewP2PRouteOptimizer()
    s.aiSecurityAnalyzer = ai.NewSecurityIntelligence()
    
    // Запуск AI обработки
    go s.startAIProcessing()
    
    s.logger.Info("AI components initialized successfully")
    return nil
}

func (s *Server) startAIProcessing() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-s.ctx.Done():
            return
        case <-ticker.C:
            s.processAIMetrics()
        }
    }
}

func (s *Server) processAIMetrics() {
    // Сбор метрик
    metrics := s.collectMetrics()
    
    // Обнаружение аномалий
    if result, err := s.aiAnomalyDetector.DetectAnomaly(metrics); err == nil {
        if result.IsAnomaly {
            s.logger.Warn("Network anomaly detected", 
                zap.Float64("score", result.Score),
                zap.Time("timestamp", result.Timestamp))
            
            // Триггер автоматического реагирования
            s.triggerAutomatedResponse(result)
        }
    }
    
    // Предсказание нагрузки
    if prediction, err := s.aiLoadPredictor.PredictLoad(24); err == nil {
        s.logger.Info("Load prediction", 
            zap.Float64("confidence", prediction.Confidence),
            zap.Int("horizon", prediction.Horizon))
    }
}
```

---

## 🎯 Заключение

Представленные примеры кода демонстрируют:

1. **Реальную реализацию** AI функций вместо заглушек
2. **Интеграцию с ML библиотеками** (GoLearn, Gonum)
3. **Правильную архитектуру** для production использования
4. **Метрики и мониторинг** AI компонентов
5. **Интеграцию с существующей системой**

Эти примеры служат основой для полной реализации AI интеграции в CloudBridge Relay.
