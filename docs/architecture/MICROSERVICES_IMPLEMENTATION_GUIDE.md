# 🛠️ CloudBridge Relay - Руководство по реализации микросервисов

## 📋 Обзор

Данное руководство содержит практические примеры реализации микросервисов для CloudBridge Relay на основе анализа существующего кода.

## 🎯 Фаза 1: Извлечение Auth Service

### **Структура Auth Service**

```bash
cmd/auth-service/
├── main.go
├── Dockerfile
└── config.yaml

internal/auth/
├── service.go
├── middleware.go
├── handlers.go
└── models.go
```

### **main.go для Auth Service**

```go
package main

import (
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "go.uber.org/zap"
    
    "github.com/twogc/cloudbridge-relay-installer/internal/auth"
    "github.com/twogc/cloudbridge-relay-installer/internal/health"
    "github.com/twogc/cloudbridge-relay-installer/internal/metrics"
)

func main() {
    // Инициализация логгера
    logger, _ := zap.NewProduction()
    defer logger.Sync()

    // Инициализация метрик
    metricsService := metrics.NewService("auth-service")
    
    // Инициализация auth service
    authService := auth.NewService(logger, metricsService)
    
    // Настройка Gin router
    router := gin.New()
    router.Use(gin.Recovery())
    router.Use(authService.Middleware())
    
    // Health check endpoints
    healthService := health.NewService(logger)
    router.GET("/health", healthService.HealthCheck)
    router.GET("/ready", healthService.ReadyCheck)
    
    // Prometheus metrics
    router.GET("/metrics", gin.WrapH(promhttp.Handler()))
    
    // Auth endpoints
    authGroup := router.Group("/auth")
    {
        authGroup.POST("/login", authService.Login)
        authGroup.POST("/logout", authService.Logout)
        authGroup.GET("/profile", authService.GetProfile)
        authGroup.POST("/refresh", authService.RefreshToken)
    }
    
    // Graceful shutdown
    srv := &http.Server{
        Addr:    ":8080",
        Handler: router,
    }
    
    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            logger.Fatal("Failed to start auth service", zap.Error(err))
        }
    }()
    
    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    logger.Info("Shutting down auth service...")
    
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    if err := srv.Shutdown(ctx); err != nil {
        logger.Fatal("Auth service forced to shutdown", zap.Error(err))
    }
    
    logger.Info("Auth service exited")
}
```

### **Auth Service Implementation**

```go
// internal/auth/service.go
package auth

import (
    "context"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
    "go.uber.org/zap"
    
    "github.com/twogc/cloudbridge-relay-installer/internal/metrics"
)

type Service struct {
    logger        *zap.Logger
    metrics       *metrics.Service
    keycloakURL   string
    jwtSecret     []byte
}

func NewService(logger *zap.Logger, metrics *metrics.Service) *Service {
    return &Service{
        logger:      logger,
        metrics:     metrics,
        keycloakURL: os.Getenv("KEYCLOAK_URL"),
        jwtSecret:   []byte(os.Getenv("JWT_SECRET")),
    }
}

func (s *Service) Login(c *gin.Context) {
    start := time.Now()
    defer func() {
        s.metrics.RecordRequestDuration("login", time.Since(start))
    }()
    
    var loginReq LoginRequest
    if err := c.ShouldBindJSON(&loginReq); err != nil {
        s.logger.Error("Invalid login request", zap.Error(err))
        c.JSON(400, gin.H{"error": "Invalid request"})
        return
    }
    
    // Интеграция с Keycloak
    token, err := s.authenticateWithKeycloak(loginReq.Username, loginReq.Password)
    if err != nil {
        s.logger.Error("Authentication failed", 
            zap.String("username", loginReq.Username),
            zap.Error(err))
        s.metrics.IncrementError("login_failed")
        c.JSON(401, gin.H{"error": "Invalid credentials"})
        return
    }
    
    s.metrics.IncrementSuccess("login")
    c.JSON(200, gin.H{"token": token})
}

func (s *Service) Middleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.JSON(401, gin.H{"error": "No token provided"})
            c.Abort()
            return
        }
        
        // Валидация JWT токена
        claims, err := s.validateToken(token)
        if err != nil {
            s.logger.Error("Token validation failed", zap.Error(err))
            c.JSON(401, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }
        
        // Добавляем claims в контекст
        c.Set("user", claims)
        c.Next()
    }
}
```

## 🎯 Фаза 2: Извлечение WireGuard Service

### **WireGuard Service с привилегированными правами**

```go
// cmd/wireguard-service/main.go
package main

import (
    "log"
    "os"
    "os/signal"
    "syscall"
    
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
    
    "github.com/twogc/cloudbridge-relay-installer/internal/wireguard"
    "github.com/twogc/cloudbridge-relay-installer/internal/health"
    "github.com/twogc/cloudbridge-relay-installer/internal/metrics"
)

func main() {
    logger, _ := zap.NewProduction()
    defer logger.Sync()
    
    // Проверка привилегированного режима
    if !isPrivileged() {
        logger.Fatal("WireGuard service requires privileged mode")
    }
    
    metricsService := metrics.NewService("wireguard-service")
    wgService := wireguard.NewService(logger, metricsService)
    
    router := gin.New()
    router.Use(gin.Recovery())
    
    // Health checks
    healthService := health.NewService(logger)
    router.GET("/health", healthService.HealthCheck)
    router.GET("/ready", healthService.ReadyCheck)
    router.GET("/metrics", gin.WrapH(promhttp.Handler()))
    
    // WireGuard endpoints
    wgGroup := router.Group("/wireguard")
    {
        wgGroup.POST("/peers", wgService.AddPeer)
        wgGroup.DELETE("/peers/:id", wgService.RemovePeer)
        wgGroup.GET("/peers", wgService.ListPeers)
        wgGroup.GET("/status", wgService.GetStatus)
        wgGroup.POST("/config", wgService.UpdateConfig)
    }
    
    // Запуск в отдельной горутине
    go wgService.Start()
    
    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    logger.Info("Shutting down WireGuard service...")
    wgService.Stop()
}

func isPrivileged() bool {
    return os.Geteuid() == 0
}
```

### **WireGuard Service Implementation**

```go
// internal/wireguard/service.go
package wireguard

import (
    "context"
    "fmt"
    "net"
    "sync"
    "time"
    
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
    
    "github.com/twogc/cloudbridge-relay-installer/internal/metrics"
)

type Service struct {
    logger    *zap.Logger
    metrics   *metrics.Service
    config    *Config
    peers     map[string]*Peer
    mutex     sync.RWMutex
    ctx       context.Context
    cancel    context.CancelFunc
}

type Config struct {
    Interface string `yaml:"interface"`
    Port      int    `yaml:"port"`
    PrivateKey string `yaml:"private_key"`
    PublicKey  string `yaml:"public_key"`
}

type Peer struct {
    ID          string    `json:"id"`
    PublicKey   string    `json:"public_key"`
    AllowedIPs  []string  `json:"allowed_ips"`
    Endpoint    string    `json:"endpoint"`
    LastSeen    time.Time `json:"last_seen"`
    BytesSent   int64     `json:"bytes_sent"`
    BytesRecv   int64     `json:"bytes_recv"`
}

func NewService(logger *zap.Logger, metrics *metrics.Service) *Service {
    ctx, cancel := context.WithCancel(context.Background())
    
    return &Service{
        logger:  logger,
        metrics: metrics,
        peers:   make(map[string]*Peer),
        ctx:     ctx,
        cancel:  cancel,
    }
}

func (s *Service) Start() error {
    s.logger.Info("Starting WireGuard service")
    
    // Инициализация WireGuard интерфейса
    if err := s.initInterface(); err != nil {
        return fmt.Errorf("failed to init interface: %w", err)
    }
    
    // Запуск мониторинга пиров
    go s.monitorPeers()
    
    // Запуск метрик сбора
    go s.collectMetrics()
    
    s.logger.Info("WireGuard service started successfully")
    return nil
}

func (s *Service) Stop() {
    s.logger.Info("Stopping WireGuard service")
    s.cancel()
    
    // Очистка интерфейса
    if err := s.cleanupInterface(); err != nil {
        s.logger.Error("Failed to cleanup interface", zap.Error(err))
    }
}

func (s *Service) AddPeer(c *gin.Context) {
    var peerReq AddPeerRequest
    if err := c.ShouldBindJSON(&peerReq); err != nil {
        c.JSON(400, gin.H{"error": "Invalid request"})
        return
    }
    
    s.mutex.Lock()
    defer s.mutex.Unlock()
    
    peer := &Peer{
        ID:         peerReq.ID,
        PublicKey:  peerReq.PublicKey,
        AllowedIPs: peerReq.AllowedIPs,
        Endpoint:   peerReq.Endpoint,
        LastSeen:   time.Now(),
    }
    
    s.peers[peer.ID] = peer
    
    // Добавление пира в WireGuard
    if err := s.addPeerToWireGuard(peer); err != nil {
        s.logger.Error("Failed to add peer to WireGuard", 
            zap.String("peer_id", peer.ID),
            zap.Error(err))
        c.JSON(500, gin.H{"error": "Failed to add peer"})
        return
    }
    
    s.metrics.IncrementSuccess("peer_added")
    s.logger.Info("Peer added successfully", zap.String("peer_id", peer.ID))
    
    c.JSON(200, gin.H{"message": "Peer added successfully"})
}

func (s *Service) GetStatus(c *gin.Context) {
    s.mutex.RLock()
    defer s.mutex.RUnlock()
    
    status := &Status{
        Interface: s.config.Interface,
        Port:      s.config.Port,
        Peers:     len(s.peers),
        Uptime:    time.Since(s.startTime).String(),
    }
    
    c.JSON(200, status)
}

func (s *Service) monitorPeers() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-s.ctx.Done():
            return
        case <-ticker.C:
            s.updatePeerStats()
        }
    }
}

func (s *Service) collectMetrics() {
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-s.ctx.Done():
            return
        case <-ticker.C:
            s.metrics.SetGauge("wireguard_peers_total", float64(len(s.peers)))
            s.metrics.SetGauge("wireguard_interface_up", 1.0)
        }
    }
}
```

## 🎯 Фаза 3: Извлечение AI/ML Service

### **AI/ML Service с GPU поддержкой**

```go
// cmd/ai-service/main.go
package main

import (
    "log"
    "os"
    "os/signal"
    "syscall"
    
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
    
    "github.com/twogc/cloudbridge-relay-installer/internal/ai"
    "github.com/twogc/cloudbridge-relay-installer/internal/health"
    "github.com/twogc/cloudbridge-relay-installer/internal/metrics"
)

func main() {
    logger, _ := zap.NewProduction()
    defer logger.Sync()
    
    metricsService := metrics.NewService("ai-service")
    aiService := ai.NewService(logger, metricsService)
    
    router := gin.New()
    router.Use(gin.Recovery())
    
    // Health checks
    healthService := health.NewService(logger)
    router.GET("/health", healthService.HealthCheck)
    router.GET("/ready", healthService.ReadyCheck)
    router.GET("/metrics", gin.WrapH(promhttp.Handler()))
    
    // AI/ML endpoints
    aiGroup := router.Group("/ai")
    {
        aiGroup.POST("/analyze", aiService.AnalyzeTraffic)
        aiGroup.POST("/detect-anomaly", aiService.DetectAnomaly)
        aiGroup.GET("/models", aiService.ListModels)
        aiGroup.POST("/models/:name/reload", aiService.ReloadModel)
        aiGroup.GET("/predictions", aiService.GetPredictions)
    }
    
    // Запуск AI обработки
    go aiService.StartProcessing()
    
    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    logger.Info("Shutting down AI service...")
    aiService.Stop()
}

func main() {
    // ... implementation
}
```

### **AI/ML Service Implementation**

```go
// internal/ai/service.go
package ai

import (
    "context"
    "encoding/json"
    "fmt"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/segmentio/kafka-go"
    "go.uber.org/zap"
    
    "github.com/twogc/cloudbridge-relay-installer/internal/metrics"
)

type Service struct {
    logger      *zap.Logger
    metrics     *metrics.Service
    models      map[string]*Model
    kafkaReader *kafka.Reader
    ctx         context.Context
    cancel      context.CancelFunc
}

type Model struct {
    Name       string    `json:"name"`
    Version    string    `json:"version"`
    Type       string    `json:"type"`
    Accuracy   float64   `json:"accuracy"`
    LoadedAt   time.Time `json:"loaded_at"`
    Predictions int64    `json:"predictions"`
}

type TrafficData struct {
    Timestamp   time.Time         `json:"timestamp"`
    SourceIP    string            `json:"source_ip"`
    DestIP      string            `json:"dest_ip"`
    Protocol    string            `json:"protocol"`
    Bytes       int64             `json:"bytes"`
    Packets     int64             `json:"packets"`
    Features    map[string]float64 `json:"features"`
}

type AnomalyResult struct {
    IsAnomaly   bool    `json:"is_anomaly"`
    Confidence  float64 `json:"confidence"`
    Score       float64 `json:"score"`
    Model       string  `json:"model"`
    Timestamp   time.Time `json:"timestamp"`
}

func NewService(logger *zap.Logger, metrics *metrics.Service) *Service {
    ctx, cancel := context.WithCancel(context.Background())
    
    // Инициализация Kafka reader
    reader := kafka.NewReader(kafka.ReaderConfig{
        Brokers: []string{os.Getenv("KAFKA_BROKERS")},
        Topic:   "traffic-data",
        GroupID: "ai-service",
    })
    
    return &Service{
        logger:      logger,
        metrics:     metrics,
        models:      make(map[string]*Model),
        kafkaReader: reader,
        ctx:         ctx,
        cancel:      cancel,
    }
}

func (s *Service) StartProcessing() {
    s.logger.Info("Starting AI processing")
    
    // Загрузка моделей
    if err := s.loadModels(); err != nil {
        s.logger.Error("Failed to load models", zap.Error(err))
        return
    }
    
    // Обработка сообщений из Kafka
    go s.processMessages()
    
    // Обновление метрик
    go s.updateMetrics()
}

func (s *Service) processMessages() {
    for {
        select {
        case <-s.ctx.Done():
            return
        default:
            m, err := s.kafkaReader.ReadMessage(s.ctx)
            if err != nil {
                s.logger.Error("Failed to read message", zap.Error(err))
                continue
            }
            
            var trafficData TrafficData
            if err := json.Unmarshal(m.Value, &trafficData); err != nil {
                s.logger.Error("Failed to unmarshal traffic data", zap.Error(err))
                continue
            }
            
            // Анализ трафика
            result := s.analyzeTraffic(trafficData)
            
            // Отправка результата в Kafka
            s.publishResult(result)
            
            s.metrics.IncrementSuccess("traffic_analyzed")
        }
    }
}

func (s *Service) AnalyzeTraffic(c *gin.Context) {
    var trafficData TrafficData
    if err := c.ShouldBindJSON(&trafficData); err != nil {
        c.JSON(400, gin.H{"error": "Invalid request"})
        return
    }
    
    result := s.analyzeTraffic(trafficData)
    
    s.metrics.IncrementSuccess("traffic_analyzed")
    c.JSON(200, result)
}

func (s *Service) DetectAnomaly(c *gin.Context) {
    var trafficData TrafficData
    if err := c.ShouldBindJSON(&trafficData); err != nil {
        c.JSON(400, gin.H{"error": "Invalid request"})
        return
    }
    
    anomaly := s.detectAnomaly(trafficData)
    
    if anomaly.IsAnomaly {
        s.metrics.IncrementAlert("anomaly_detected")
        s.logger.Warn("Anomaly detected", 
            zap.String("source_ip", trafficData.SourceIP),
            zap.Float64("confidence", anomaly.Confidence))
    }
    
    c.JSON(200, anomaly)
}

func (s *Service) analyzeTraffic(data TrafficData) *AnomalyResult {
    // Реализация анализа трафика с использованием ML моделей
    // Здесь будет интеграция с TensorFlow/PyTorch через CGO или gRPC
    
    return &AnomalyResult{
        IsAnomaly:  false,
        Confidence: 0.95,
        Score:      0.05,
        Model:      "traffic_classifier_v1",
        Timestamp:  time.Now(),
    }
}

func (s *Service) detectAnomaly(data TrafficData) *AnomalyResult {
    // Реализация детекции аномалий
    // Использование изолированных лесов, LSTM или других алгоритмов
    
    return &AnomalyResult{
        IsAnomaly:  false,
        Confidence: 0.98,
        Score:      0.02,
        Model:      "anomaly_detector_v1",
        Timestamp:  time.Now(),
    }
}

func (s *Service) updateMetrics() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-s.ctx.Done():
            return
        case <-ticker.C:
            s.metrics.SetGauge("ai_models_loaded", float64(len(s.models)))
            s.metrics.SetGauge("ai_processing_queue_size", float64(s.getQueueSize()))
        }
    }
}
```

## 🎯 Фаза 4: Извлечение Quantum Service

### **Quantum Service с Post-Quantum криптографией**

```go
// cmd/quantum-service/main.go
package main

import (
    "log"
    "os"
    "os/signal"
    "syscall"
    
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
    
    "github.com/twogc/cloudbridge-relay-installer/internal/quantum"
    "github.com/twogc/cloudbridge-relay-installer/internal/health"
    "github.com/twogc/cloudbridge-relay-installer/internal/metrics"
)

func main() {
    logger, _ := zap.NewProduction()
    defer logger.Sync()
    
    metricsService := metrics.NewService("quantum-service")
    quantumService := quantum.NewService(logger, metricsService)
    
    router := gin.New()
    router.Use(gin.Recovery())
    
    // Health checks
    healthService := health.NewService(logger)
    router.GET("/health", healthService.HealthCheck)
    router.GET("/ready", healthService.ReadyCheck)
    router.GET("/metrics", gin.WrapH(promhttp.Handler()))
    
    // Quantum crypto endpoints
    quantumGroup := router.Group("/quantum")
    {
        quantumGroup.POST("/generate-keys", quantumService.GenerateKeyPair)
        quantumGroup.POST("/encrypt", quantumService.Encrypt)
        quantumGroup.POST("/decrypt", quantumService.Decrypt)
        quantumGroup.POST("/sign", quantumService.Sign)
        quantumGroup.POST("/verify", quantumService.Verify)
        quantumGroup.GET("/algorithms", quantumService.ListAlgorithms)
        quantumGroup.POST("/rotate-keys", quantumService.RotateKeys)
    }
    
    // Запуск автоматической ротации ключей
    go quantumService.StartKeyRotation()
    
    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    logger.Info("Shutting down Quantum service...")
    quantumService.Stop()
}
```

### **Quantum Service Implementation**

```go
// internal/quantum/service.go
package quantum

import (
    "context"
    "crypto/rand"
    "encoding/base64"
    "fmt"
    "time"
    
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
    
    "github.com/twogc/cloudbridge-relay-installer/internal/metrics"
)

type Service struct {
    logger    *zap.Logger
    metrics   *metrics.Service
    algorithms map[string]Algorithm
    keys      map[string]*KeyPair
    ctx       context.Context
    cancel    context.CancelFunc
}

type Algorithm struct {
    Name        string `json:"name"`
    Type        string `json:"type"` // KEM, Signature
    SecurityLevel int  `json:"security_level"`
    KeySize     int    `json:"key_size"`
    Enabled     bool   `json:"enabled"`
}

type KeyPair struct {
    ID        string    `json:"id"`
    Algorithm string    `json:"algorithm"`
    PublicKey string    `json:"public_key"`
    PrivateKey string   `json:"private_key,omitempty"`
    CreatedAt time.Time `json:"created_at"`
    ExpiresAt time.Time `json:"expires_at"`
}

type EncryptRequest struct {
    Algorithm string `json:"algorithm"`
    Data      string `json:"data"`
    PublicKey string `json:"public_key"`
}

type EncryptResponse struct {
    Ciphertext string `json:"ciphertext"`
    Algorithm  string `json:"algorithm"`
    KeyID      string `json:"key_id"`
}

func NewService(logger *zap.Logger, metrics *metrics.Service) *Service {
    ctx, cancel := context.WithCancel(context.Background())
    
    return &Service{
        logger:     logger,
        metrics:    metrics,
        algorithms: make(map[string]Algorithm),
        keys:       make(map[string]*KeyPair),
        ctx:        ctx,
        cancel:     cancel,
    }
}

func (s *Service) StartKeyRotation() {
    s.logger.Info("Starting quantum key rotation service")
    
    // Инициализация алгоритмов
    s.initAlgorithms()
    
    // Генерация начальных ключей
    s.generateInitialKeys()
    
    // Запуск автоматической ротации
    go s.rotateKeysPeriodically()
}

func (s *Service) GenerateKeyPair(c *gin.Context) {
    var req struct {
        Algorithm string `json:"algorithm"`
        KeySize   int    `json:"key_size,omitempty"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": "Invalid request"})
        return
    }
    
    // Проверка поддержки алгоритма
    if !s.isAlgorithmSupported(req.Algorithm) {
        c.JSON(400, gin.H{"error": "Algorithm not supported"})
        return
    }
    
    // Генерация ключевой пары
    keyPair, err := s.generateKeyPair(req.Algorithm, req.KeySize)
    if err != nil {
        s.logger.Error("Failed to generate key pair", 
            zap.String("algorithm", req.Algorithm),
            zap.Error(err))
        c.JSON(500, gin.H{"error": "Failed to generate key pair"})
        return
    }
    
    s.keys[keyPair.ID] = keyPair
    s.metrics.IncrementSuccess("key_pair_generated")
    
    c.JSON(200, keyPair)
}

func (s *Service) Encrypt(c *gin.Context) {
    var req EncryptRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": "Invalid request"})
        return
    }
    
    // Декодирование данных
    data, err := base64.StdEncoding.DecodeString(req.Data)
    if err != nil {
        c.JSON(400, gin.H{"error": "Invalid data encoding"})
        return
    }
    
    // Шифрование
    ciphertext, err := s.encryptData(data, req.PublicKey, req.Algorithm)
    if err != nil {
        s.logger.Error("Encryption failed", zap.Error(err))
        c.JSON(500, gin.H{"error": "Encryption failed"})
        return
    }
    
    response := EncryptResponse{
        Ciphertext: base64.StdEncoding.EncodeToString(ciphertext),
        Algorithm:  req.Algorithm,
        KeyID:      s.getKeyID(req.PublicKey),
    }
    
    s.metrics.IncrementSuccess("data_encrypted")
    c.JSON(200, response)
}

func (s *Service) Sign(c *gin.Context) {
    var req struct {
        Algorithm  string `json:"algorithm"`
        Data       string `json:"data"`
        PrivateKey string `json:"private_key"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": "Invalid request"})
        return
    }
    
    // Декодирование данных
    data, err := base64.StdEncoding.DecodeString(req.Data)
    if err != nil {
        c.JSON(400, gin.H{"error": "Invalid data encoding"})
        return
    }
    
    // Подписание
    signature, err := s.signData(data, req.PrivateKey, req.Algorithm)
    if err != nil {
        s.logger.Error("Signing failed", zap.Error(err))
        c.JSON(500, gin.H{"error": "Signing failed"})
        return
    }
    
    s.metrics.IncrementSuccess("data_signed")
    c.JSON(200, gin.H{
        "signature": base64.StdEncoding.EncodeToString(signature),
        "algorithm": req.Algorithm,
    })
}

func (s *Service) initAlgorithms() {
    s.algorithms["CRYSTALS-Kyber-512"] = Algorithm{
        Name:         "CRYSTALS-Kyber-512",
        Type:         "KEM",
        SecurityLevel: 1,
        KeySize:      512,
        Enabled:      true,
    }
    
    s.algorithms["CRYSTALS-Kyber-768"] = Algorithm{
        Name:         "CRYSTALS-Kyber-768",
        Type:         "KEM",
        SecurityLevel: 3,
        KeySize:      768,
        Enabled:      true,
    }
    
    s.algorithms["CRYSTALS-Dilithium2"] = Algorithm{
        Name:         "CRYSTALS-Dilithium2",
        Type:         "Signature",
        SecurityLevel: 2,
        KeySize:      1312,
        Enabled:      true,
    }
    
    s.algorithms["CRYSTALS-Dilithium3"] = Algorithm{
        Name:         "CRYSTALS-Dilithium3",
        Type:         "Signature",
        SecurityLevel: 3,
        KeySize:      1952,
        Enabled:      true,
    }
}

func (s *Service) generateKeyPair(algorithm string, keySize int) (*KeyPair, error) {
    // Реализация генерации ключей для различных алгоритмов
    // Здесь будет интеграция с liboqs или другими библиотеками
    
    id := s.generateKeyID()
    
    // Генерация случайных ключей (заглушка)
    publicKey := make([]byte, keySize/8)
    privateKey := make([]byte, keySize/8)
    
    if _, err := rand.Read(publicKey); err != nil {
        return nil, err
    }
    if _, err := rand.Read(privateKey); err != nil {
        return nil, err
    }
    
    return &KeyPair{
        ID:         id,
        Algorithm:  algorithm,
        PublicKey:  base64.StdEncoding.EncodeToString(publicKey),
        PrivateKey: base64.StdEncoding.EncodeToString(privateKey),
        CreatedAt:  time.Now(),
        ExpiresAt:  time.Now().Add(24 * time.Hour), // 24 часа
    }, nil
}

func (s *Service) rotateKeysPeriodically() {
    ticker := time.NewTicker(1 * time.Hour) // Ротация каждый час
    defer ticker.Stop()
    
    for {
        select {
        case <-s.ctx.Done():
            return
        case <-ticker.C:
            s.rotateExpiredKeys()
        }
    }
}

func (s *Service) rotateExpiredKeys() {
    now := time.Now()
    
    for id, keyPair := range s.keys {
        if now.After(keyPair.ExpiresAt) {
            s.logger.Info("Rotating expired key", zap.String("key_id", id))
            
            // Генерация нового ключа
            newKeyPair, err := s.generateKeyPair(keyPair.Algorithm, 0)
            if err != nil {
                s.logger.Error("Failed to generate new key", zap.Error(err))
                continue
            }
            
            // Замена ключа
            s.keys[id] = newKeyPair
            s.metrics.IncrementSuccess("key_rotated")
        }
    }
}
```

## 🎯 Фаза 5: Интеграция с Cadence Workflows

### **Cadence Workflow для оркестрации микросервисов**

```go
// internal/cadence/workflows/global_mesh_workflow.go
package workflows

import (
    "time"
    
    "go.uber.org/cadence/workflow"
    "go.uber.org/zap"
)

type GlobalMeshInput struct {
    Regions []string `json:"regions"`
    Config  MeshConfig `json:"config"`
}

type MeshConfig struct {
    WireGuardPort int    `json:"wireguard_port"`
    P2PEnabled    bool   `json:"p2p_enabled"`
    QuantumEnabled bool  `json:"quantum_enabled"`
}

type RegionalMeshResult struct {
    Region    string `json:"region"`
    PeerCount int    `json:"peer_count"`
    Status    string `json:"status"`
    Error     string `json:"error,omitempty"`
}

func GlobalP2PMeshWorkflow(ctx workflow.Context, input GlobalMeshInput) error {
    logger := workflow.GetLogger(ctx)
    
    logger.Info("Starting global P2P mesh workflow", 
        zap.Strings("regions", input.Regions))
    
    // Параллельное выполнение в разных регионах
    futures := []workflow.Future{}
    
    for _, region := range input.Regions {
        future := workflow.ExecuteActivity(ctx, 
            activity.Options{
                TaskList: fmt.Sprintf("wireguard-%s", region),
                ScheduleToCloseTimeout: time.Minute * 5,
                HeartbeatTimeout: time.Second * 30,
            },
            "EstablishRegionalMesh", region, input.Config)
        futures = append(futures, future)
    }
    
    // Ожидание завершения во всех регионах
    results := []RegionalMeshResult{}
    for i, future := range futures {
        var result RegionalMeshResult
        if err := future.Get(ctx, &result); err != nil {
            logger.Error("Regional mesh failed", 
                zap.String("region", input.Regions[i]),
                zap.Error(err))
            result = RegionalMeshResult{
                Region: input.Regions[i],
                Status: "failed",
                Error:  err.Error(),
            }
        }
        results = append(results, result)
    }
    
    // Анализ результатов
    successCount := 0
    for _, result := range results {
        if result.Status == "success" {
            successCount++
        }
    }
    
    logger.Info("Global mesh workflow completed",
        zap.Int("total_regions", len(input.Regions)),
        zap.Int("successful_regions", successCount))
    
    // Если большинство регионов успешно, считаем операцию успешной
    if successCount >= len(input.Regions)/2 {
        return nil
    }
    
    return fmt.Errorf("insufficient successful regions: %d/%d", 
        successCount, len(input.Regions))
}

func EstablishRegionalMesh(ctx workflow.Context, region string, config MeshConfig) (*RegionalMeshResult, error) {
    logger := workflow.GetLogger(ctx)
    
    logger.Info("Establishing regional mesh", zap.String("region", region))
    
    // 1. Инициализация WireGuard сервиса
    if err := workflow.ExecuteActivity(ctx, "InitWireGuardService", region, config).Get(ctx, nil); err != nil {
        return nil, fmt.Errorf("failed to init WireGuard: %w", err)
    }
    
    // 2. Настройка P2P discovery
    if config.P2PEnabled {
        if err := workflow.ExecuteActivity(ctx, "SetupP2PDiscovery", region).Get(ctx, nil); err != nil {
            return nil, fmt.Errorf("failed to setup P2P: %w", err)
        }
    }
    
    // 3. Настройка quantum crypto
    if config.QuantumEnabled {
        if err := workflow.ExecuteActivity(ctx, "SetupQuantumCrypto", region).Get(ctx, nil); err != nil {
            return nil, fmt.Errorf("failed to setup quantum: %w", err)
        }
    }
    
    // 4. Подключение к глобальной сети
    peerCount, err := workflow.ExecuteActivity(ctx, "ConnectToGlobalMesh", region).Get(ctx, 0)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to global mesh: %w", err)
    }
    
    logger.Info("Regional mesh established successfully",
        zap.String("region", region),
        zap.Int("peers", peerCount.(int)))
    
    return &RegionalMeshResult{
        Region:    region,
        PeerCount: peerCount.(int),
        Status:    "success",
    }, nil
}
```

## 🚀 Команды для запуска

### **Локальная разработка:**
```bash
# Запуск всех микросервисов
docker-compose up -d

# Проверка здоровья всех сервисов
docker-compose ps

# Просмотр логов конкретного сервиса
docker-compose logs -f wireguard-service

# Масштабирование сервиса
docker-compose up --scale ai-service=3 -d
```

### **Production развертывание:**
```bash
# Развертывание в Kubernetes
kubectl apply -f k8s/

# Проверка статуса всех сервисов
kubectl get pods -l app.kubernetes.io/name=cloudbridge

# Просмотр метрик
kubectl port-forward svc/prometheus 9090:9090
open http://localhost:9090

# Distributed tracing
kubectl port-forward svc/jaeger 16686:16686
open http://localhost:16686
```

## 📊 Мониторинг микросервисной архитектуры

### **Ключевые метрики:**
```promql
# Availability всех сервисов
avg(up{job=~"cloudbridge-.*"}) * 100

# Latency P99 для API Gateway
histogram_quantile(0.99, rate(http_request_duration_seconds_bucket{job="api-gateway"}[5m]))

# Error rate для всех сервисов
sum(rate(http_requests_total{status=~"5.."}[5m])) / sum(rate(http_requests_total[5m])) * 100

# Throughput для WireGuard
rate(wireguard_peers_total[5m])
```

## 🎯 Выводы и рекомендации

### **Преимущества микросервисной архитектуры:**
1. ✅ **Независимое масштабирование** каждого компонента
2. ✅ **Изоляция сбоев** - падение одного сервиса не влияет на другие
3. ✅ **Технологическое разнообразие** - каждый сервис может использовать оптимальные инструменты
4. ✅ **Быстрое развертывание** - независимые релизы
5. ✅ **Команды разработки** могут работать параллельно

### **Следование Золотым правилам:**
- **Отказоустойчивость** встроена в каждый микросервис
- **Безопасность** реализована на каждом уровне
- **Наблюдаемость** обеспечена для всех компонентов
- **Автоматизация** покрывает весь жизненный цикл
- **Производительность** оптимизирована для каждого сервиса

### **Следующие шаги:**
1. 🎯 Начать с извлечения Auth Service
2. 🔧 Настроить CI/CD для микросервисов
3. 📊 Внедрить комплексный мониторинг
4. 🌍 Запланировать multi-region развертывание

Эта архитектура обеспечит CloudBridge Relay способность масштабироваться до глобальной сети с миллионами P2P и P2C клиентов! 🚀
