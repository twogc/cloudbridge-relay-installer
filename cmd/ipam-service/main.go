package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/twogc/cloudbridge-relay-installer/internal/database"
	"github.com/twogc/cloudbridge-relay-installer/internal/ipam"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer logger.Sync()

	logger.Info("Starting CloudBridge IPAM Service")

	// Initialize database connection
	dbConfig := &database.Config{
		Host:     "postgres.cloudbridge.svc.cluster.local",
		Port:     5432,
		Database: "cadence",
		Username: "cadence",
		Password: "cadence123",
		SSLMode:  "disable",
	}

	dbClient, err := database.NewClient(dbConfig, logger)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer dbClient.Close()

	// Create repository
	repo := database.NewRepository(dbClient.GetDB(), logger)

	// Initialize IPAM service
	ipamService := ipam.NewService(repo, logger)

	// Initialize HTTP server
	router := setupRouter(ipamService, logger)

	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		logger.Info("Starting IPAM HTTP server",
			zap.String("addr", server.Addr))

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down IPAM service...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("IPAM service stopped")
}

func setupRouter(ipamService *ipam.Service, logger *zap.Logger) *gin.Engine {
	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())
	router.Use(loggingMiddleware(logger))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "ipam",
			"timestamp": time.Now().UTC(),
		})
	})

	// API routes
	api := router.Group("/api/v1")
	{
		// IPAM routes
		ipam := api.Group("/ipam")
		{
			ipam.POST("/tenants", ipamService.CreateTenant)
			ipam.GET("/tenants/:tenant_id", ipamService.GetTenant)
			ipam.DELETE("/tenants/:tenant_id", ipamService.DeleteTenant)
			ipam.GET("/tenants", ipamService.ListTenants)

			ipam.POST("/tenants/:tenant_id/allocate", ipamService.AllocateSubnet)
			ipam.POST("/tenants/:tenant_id/deallocate", ipamService.DeallocateSubnet)
			ipam.GET("/tenants/:tenant_id/subnets", ipamService.ListTenantSubnets)

			ipam.GET("/stats", ipamService.GetStats)
			ipam.GET("/metrics", ipamService.GetMetrics)
		}
	}

	return router
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func loggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		logger.Info("HTTP Request",
			zap.String("method", param.Method),
			zap.String("path", param.Path),
			zap.Int("status", param.StatusCode),
			zap.Duration("latency", param.Latency),
			zap.String("client_ip", param.ClientIP),
		)
		return ""
	})
}
