package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Gin mode
	if m := os.Getenv("GIN_MODE"); m != "" {
		gin.SetMode(m)
	}

	r := gin.New()
	// базовые middlewares
	r.Use(gin.Recovery())
	r.Use(ginzap(logger)) // см. реализацию ниже

	// liveness / readiness
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "healthy", "service": "cadence"}) })
	r.GET("/ready", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ready"}) })

	// версии/инфо (удобно для отладки)
	r.GET("/version", func(c *gin.Context) {
		c.JSON(200, gin.H{"version": os.Getenv("BUILD_VERSION"), "commit": os.Getenv("GIT_SHA")})
	})

	// Cadence API (заглушки)
	r.GET("/workflows", func(c *gin.Context) { c.JSON(200, gin.H{"workflows": []string{}}) })
	r.POST("/workflows", func(c *gin.Context) { c.JSON(201, gin.H{"message": "workflow started"}) })

	// Prometheus metrics
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// конфиг сервера
	host := getenvDefault("CADENCE_HOST", "0.0.0.0")
	port := getenvDefault("CADENCE_PORT", "8082")
	addr := host + ":" + port

	// http.Server с таймаутами
	srv := &http.Server{
		Addr:              addr,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	logger.Info("Cadence Service starting", zap.String("addr", addr))

	// graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server listen error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down Cadence Service...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Graceful shutdown failed, forcing close", zap.Error(err))
		_ = srv.Close()
	}
	logger.Info("Cadence Service stopped")
}

// --- helpers ---

func getenvDefault(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

// простой zap-логгер для запросов (без дополнительной либы)
func ginzap(l *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		lat := time.Since(start)
		l.Info("http",
			zap.String("method", c.Request.Method),
			zap.String("path", c.FullPath()),
			zap.Int("status", c.Writer.Status()),
			zap.String("ip", c.ClientIP()),
			zap.Duration("latency", lat),
			zap.String("ua", c.Request.UserAgent()),
			zap.String("err", c.Errors.ByType(gin.ErrorTypePrivate).String()),
		)
	}
}
