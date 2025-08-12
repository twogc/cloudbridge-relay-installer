package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/twogc/cloudbridge-relay-installer/config"
	"github.com/twogc/cloudbridge-relay-installer/internal/api"
	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
	"github.com/twogc/cloudbridge-relay-installer/internal/metrics"
	"github.com/twogc/cloudbridge-relay-installer/internal/tunnel"
	"go.uber.org/zap"
)

var (
	configFile = flag.String("config", "config/config.json", "Path to configuration file")
	logLevel   = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	logFile    = flag.String("log-file", "", "Log file path")
)

func main() {
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Debug: Print API configuration
	log.Printf("API Config: Enabled=%v, Port=%d, Version=%s", cfg.API.Enabled, cfg.API.Port, cfg.API.Version)

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

	if *logFile != "" {
		loggerConfig.Output = "both"
		loggerConfig.FilePath = *logFile
	}

	loggerInstance, err := logger.New(loggerConfig)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// Initialize metrics
	var metricsInstance *metrics.Metrics
	if cfg.Metrics.Enabled {
		metricsInstance = metrics.NewMetrics(prometheus.DefaultRegisterer)
	}

	// Initialize tunnel manager
	tunnelManager := tunnel.NewManager(cfg.Server.MaxTunnels, loggerInstance, metricsInstance)

	// Create API server
	apiServer := api.NewServer(api.Config{
		Port:          fmt.Sprintf(":%d", cfg.API.Port),
		TunnelManager: tunnelManager,
		Metrics:       metricsInstance,
		Logger:        loggerInstance,
		Version:       cfg.API.Version,
	})

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		loggerInstance.Info("Received shutdown signal, stopping API server...")
		cancel()
	}()

	// Start API server
	loggerInstance.Info("Starting API server",
		zap.String("port", fmt.Sprintf(":%d", cfg.API.Port)),
		zap.String("version", cfg.API.Version),
	)

	if err := apiServer.Start(ctx); err != nil {
		loggerInstance.Error("Failed to start API server", zap.Error(err))
		os.Exit(1)
	}
}
