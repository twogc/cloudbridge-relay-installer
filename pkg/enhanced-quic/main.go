package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/twogc/enhanced-quic/internal/enhanced"
	"github.com/twogc/enhanced-quic/internal/logger"
	"github.com/twogc/enhanced-quic/internal/metrics"
)

var (
	configFile = flag.String("config", "config/config.json", "Path to configuration file")
	logLevel   = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	version    = flag.Bool("version", false, "Show version information")
	
	// Version information
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	flag.Parse()

	// Show version
	if *version {
		fmt.Printf("Enhanced QUIC v%s (built %s)\n", Version, BuildTime)
		os.Exit(0)
	}

	// Initialize logger
	loggerConfig := logger.Config{
		Level:      logger.LogLevel(*logLevel),
		Output:     "stdout",
		JSON:       true,
		MaxSize:    100,
		MaxAge:     30,
		MaxBackups: 10,
		Compress:   true,
	}

	loggerInstance, err := logger.New(loggerConfig)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	loggerInstance.Info("Starting Enhanced QUIC",
		logger.String("version", Version),
		logger.String("build_time", BuildTime),
	)

	// Load configuration
	config, err := enhanced.LoadConfig(*configFile)
	if err != nil {
		loggerInstance.Fatal("Failed to load configuration", logger.Error(err))
	}

	// Initialize metrics
	var metricsInstance *metrics.EnhancedMetrics
	if config.Metrics.Enabled {
		metricsInstance = metrics.NewEnhancedMetrics(nil) // Use default registry
	}

	// Create enhanced QUIC server
	server := enhanced.NewServer(config, loggerInstance, metricsInstance)

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		loggerInstance.Info("Received shutdown signal, stopping server...")
		cancel()
	}()

	// Start server
	if err := server.Start(ctx); err != nil {
		loggerInstance.Fatal("Failed to start server", logger.Error(err))
	}

	// Wait for context cancellation
	<-ctx.Done()

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Stop(shutdownCtx); err != nil {
		loggerInstance.Error("Error during shutdown", logger.Error(err))
	}

	loggerInstance.Info("Enhanced QUIC server stopped")
} 