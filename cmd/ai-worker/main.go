package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/internal/cadence"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	logger.Info("Starting AI Worker...")

	// Create Cadence configuration
	cadenceConfig := cadence.Config{
		HostPort: "cadence-service:7933",
		Domain:   "cloudbridge",
		TaskList: "ai-task-list",
	}

	// Create Cadence client
	cadenceClient, err := cadence.NewClient(cadenceConfig, logger)
	if err != nil {
		logger.Fatal("Failed to create Cadence client", zap.Error(err))
	}
	defer cadenceClient.Close()

	// Start worker
	err = cadenceClient.Start()
	if err != nil {
		logger.Fatal("Failed to start worker", zap.Error(err))
	}
	defer cadenceClient.Stop()

	logger.Info("AI Worker connected to Cadence successfully")

	// Start HTTP server for health checks
	server := &http.Server{
		Addr: ":8085",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		}),
	}

	go func() {
		logger.Info("Starting health check server on :8085")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Health check server failed", zap.Error(err))
		}
	}()

	// Give the server time to start
	time.Sleep(1 * time.Second)

	logger.Info("AI Worker started successfully")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Keep the worker running
	select {
	case <-sigChan:
		logger.Info("Received shutdown signal")
	}

	logger.Info("AI Worker shutting down...")

	// Shutdown HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)
}
