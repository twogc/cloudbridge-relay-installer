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
		c.JSON(200, gin.H{"status": "healthy", "service": "quic"})
	})

	// QUIC API endpoints
	r.GET("/connections", func(c *gin.Context) {
		c.JSON(200, gin.H{"connections": []string{}})
	})

	r.POST("/connections", func(c *gin.Context) {
		c.JSON(201, gin.H{"message": "connection created"})
	})

	// Metrics
	r.GET("/metrics", func(c *gin.Context) {
		c.JSON(200, gin.H{"metrics": "quic_metrics"})
	})

	port := os.Getenv("QUIC_PORT")
	if port == "" {
		port = "8081"
	}

	logger.Info("QUIC Service starting", zap.String("port", port))
	if err := r.Run(":" + port); err != nil {
		logger.Fatal("Failed to start QUIC service", zap.Error(err))
	}
}
