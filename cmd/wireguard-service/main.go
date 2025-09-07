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
		c.JSON(200, gin.H{"status": "healthy", "service": "wireguard"})
	})

	// WireGuard API endpoints
	r.GET("/tunnels", func(c *gin.Context) {
		c.JSON(200, gin.H{"tunnels": []string{}})
	})

	r.POST("/tunnels", func(c *gin.Context) {
		c.JSON(201, gin.H{"message": "tunnel created"})
	})

	// Metrics
	r.GET("/metrics", func(c *gin.Context) {
		c.JSON(200, gin.H{"metrics": "wireguard_metrics"})
	})

	port := os.Getenv("WIREGUARD_PORT")
	if port == "" {
		port = "8080"
	}

	logger.Info("WireGuard Service starting", zap.String("port", port))
	if err := r.Run(":" + port); err != nil {
		logger.Fatal("Failed to start WireGuard service", zap.Error(err))
	}
}
