package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/twogc/cloudbridge-relay-installer/internal/quic"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Initialize experimental metrics
	experimentalMetrics := quic.NewExperimentalMetrics(logger)
	logger.Info("Experimental metrics initialized",
		zap.Any("metrics_snapshot", experimentalMetrics.GetMetricsSnapshot()))

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

	// Prometheus Metrics Endpoint
	// Serves all Prometheus metrics including experimental QUIC features:
	// - BBRv2 congestion control (cloudbridge_quic_bbrv2_*)
	// - ACK Frequency optimization (cloudbridge_quic_ack_freq_*)
	// - Forward Error Correction (cloudbridge_quic_fec_*)
	// - QUIC Bit Greasing (cloudbridge_quic_greasing_*)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Metrics summary endpoint (human-readable)
	r.GET("/metrics/summary", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"service": "cloudbridge-quic",
			"version": "1.0.0",
			"experimental_features": experimentalMetrics.GetMetricsSnapshot(),
			"endpoints": map[string]string{
				"prometheus_metrics": "/metrics",
				"metrics_summary":    "/metrics/summary",
			},
			"documentation": map[string]string{
				"integration_report": "docs/integration/EXPERIMENTAL_FEATURES_INTEGRATION_REPORT.md",
				"configuration":      "config/experimental-features.yaml",
				"examples":           "internal/quic/experimental_example.go",
			},
		})
	})

	port := os.Getenv("QUIC_PORT")
	if port == "" {
		port = "8081"
	}

	logger.Info("QUIC Service starting",
		zap.String("port", port),
		zap.String("metrics_endpoint", "http://localhost:"+port+"/metrics"),
		zap.String("summary_endpoint", "http://localhost:"+port+"/metrics/summary"))

	if err := r.Run(":" + port); err != nil {
		logger.Fatal("Failed to start QUIC service", zap.Error(err))
	}
}
