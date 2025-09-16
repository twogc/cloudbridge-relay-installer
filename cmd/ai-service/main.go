package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Загружаем конфигурацию AI
	configPath := os.Getenv("AI_CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/ai-config.yaml" // Используем путь из ConfigMap
	}

	config, err := ai.LoadConfig(configPath)
	if err != nil {
		logger.Warn("Failed to load AI config, using defaults", zap.Error(err))
		config = ai.GetDefaultConfig()
	}

	// Создаем AI сервис
	aiService := ai.NewService(config, logger)

	// Инициализируем AI сервис
	if err := aiService.Initialize(); err != nil {
		logger.Fatal("Failed to initialize AI service", zap.Error(err))
	}

	// Запускаем AI сервис
	if err := aiService.Start(); err != nil {
		logger.Fatal("Failed to start AI service", zap.Error(err))
	}
	defer aiService.Stop()

	// Создаем метрики
	metrics := NewMetrics()
	metrics.StartMetricsUpdater()

	// Настраиваем Gin router
	r := gin.Default()

	// Добавляем Prometheus middleware
	r.Use(metrics.PrometheusMiddleware())

	// Создаем AI API сервис
	apiService := NewAIService(aiService, logger)

	// Настраиваем маршруты
	apiService.SetupRoutes(r)

	// Health checks
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy", "service": "ai"})
	})

	// Legacy endpoints для совместимости
	r.GET("/anomalies", func(c *gin.Context) {
		c.JSON(200, gin.H{"anomalies": []string{}})
	})

	r.POST("/analyze", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "analysis completed"})
	})

	// Prometheus metrics endpoint
	r.GET("/metrics", metrics.GetPrometheusHandler())

	port := os.Getenv("AI_PORT")
	if port == "" {
		port = "8083"
	}

	logger.Info("AI Service starting", zap.String("port", port))
	if err := r.Run(":" + port); err != nil {
		logger.Fatal("Failed to start AI service", zap.Error(err))
	}
}
