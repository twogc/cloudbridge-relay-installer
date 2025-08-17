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
		c.JSON(200, gin.H{"status": "healthy", "service": "cadence"})
	})

	// Cadence API endpoints
	r.GET("/workflows", func(c *gin.Context) {
		c.JSON(200, gin.H{"workflows": []string{}})
	})

	r.POST("/workflows", func(c *gin.Context) {
		c.JSON(201, gin.H{"message": "workflow started"})
	})

	// Metrics
	r.GET("/metrics", func(c *gin.Context) {
		c.JSON(200, gin.H{"metrics": "cadence_metrics"})
	})

	port := os.Getenv("CADENCE_PORT")
	if port == "" {
		port = "8082"
	}

	logger.Info("Cadence Service starting", zap.String("port", port))
	r.Run(":" + port)
}
