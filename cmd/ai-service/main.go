package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	r := gin.Default()

	// Health checks
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy", "service": "ai"})
	})

	// AI API endpoints
	r.GET("/anomalies", func(c *gin.Context) {
		c.JSON(200, gin.H{"anomalies": []string{}})
	})

	r.POST("/analyze", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "analysis completed"})
	})

	// Metrics
	r.GET("/metrics", func(c *gin.Context) {
		c.JSON(200, gin.H{"metrics": "ai_metrics"})
	})

	port := os.Getenv("AI_PORT")
	if port == "" {
		port = "8083"
	}

	logger.Info("AI Service starting", zap.String("port", port))
	if err := r.Run(":" + port); err != nil {
		logger.Fatal("Failed to start AI service", zap.Error(err))
	}
}
