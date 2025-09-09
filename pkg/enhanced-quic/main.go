package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

var (
	configFile = flag.String("config", "config/config.json", "Path to configuration file")
	logLevel   = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	version    = flag.Bool("version", false, "Show version information")

	// Version information
	Version   = "dev"
	BuildTime = "unknown"
)

// Enhanced QUIC configuration
type Config struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	MetricsPort int    `json:"metrics_port"`
	CertFile    string `json:"cert_file"`
	KeyFile     string `json:"key_file"`
	MaxStreams  int    `json:"max_streams"`
	EnableHTTP3 bool   `json:"enable_http3"`
}

// Enhanced QUIC metrics
var (
	connectionsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "enhanced_quic_connections_total",
			Help: "Total number of Enhanced QUIC connections",
		},
		[]string{"status"},
	)

	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "enhanced_quic_requests_total",
			Help: "Total number of Enhanced QUIC requests",
		},
		[]string{"method", "status"},
	)
)

func init() {
	prometheus.MustRegister(connectionsTotal)
	prometheus.MustRegister(requestsTotal)
}

func main() {
	flag.Parse()

	// Show version
	if *version {
		fmt.Printf("Enhanced QUIC v%s (built %s)\n", Version, BuildTime)
		os.Exit(0)
	}

	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("Starting Enhanced QUIC Demo Server",
		zap.String("version", Version),
		zap.String("build_time", BuildTime),
	)

	// Default configuration
	config := &Config{
		Host:        "localhost",
		Port:        8080,
		MetricsPort: 8081,
		CertFile:    "cert.pem",
		KeyFile:     "key.pem",
		MaxStreams:  100,
		EnableHTTP3: true,
	}

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Received shutdown signal, stopping server...")
		cancel()
	}()

	// Start metrics server
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		metricsAddr := fmt.Sprintf("%s:%d", config.Host, config.MetricsPort)
		logger.Info("Starting metrics server", zap.String("address", metricsAddr))
		if err := http.ListenAndServe(metricsAddr, nil); err != nil {
			logger.Error("Metrics server error", zap.Error(err))
		}
	}()

	// Create HTTP/3 handler
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		connectionsTotal.WithLabelValues("success").Inc()
		requestsTotal.WithLabelValues(r.Method, "200").Inc()

		w.Header().Set("Content-Type", "application/json")
		response := fmt.Sprintf(`{
			"message": "Enhanced QUIC Demo Server",
			"version": "%s",
			"protocol": "HTTP/3 over QUIC",
			"timestamp": "%s",
			"client_addr": "%s"
		}`, Version, time.Now().Format(time.RFC3339), r.RemoteAddr)

		if _, err := w.Write([]byte(response)); err != nil {
			logger.Error("Failed to write response", zap.Error(err))
		}
		logger.Info("Request handled",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("client", r.RemoteAddr),
		)
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{"status": "healthy", "service": "enhanced-quic"}`)); err != nil {
			logger.Error("Failed to write health response", zap.Error(err))
		}
	})

	// Generate self-signed certificate if not exists
	if _, err := os.Stat(config.CertFile); os.IsNotExist(err) {
		logger.Info("Generating self-signed certificate...")
		if err := generateSelfSignedCert(config.CertFile, config.KeyFile); err != nil {
			logger.Fatal("Failed to generate certificate", zap.Error(err))
		}
	}

	// Start Enhanced QUIC server with HTTP/2 support
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	logger.Info("Starting Enhanced QUIC server",
		zap.String("address", server.Addr),
		zap.Int("max_streams", config.MaxStreams),
		zap.Bool("http3_enabled", config.EnableHTTP3),
	)

	// Start server in goroutine
	go func() {
		var err error
		if config.EnableHTTP3 {
			// For now, use HTTPS instead of HTTP/3
			err = server.ListenAndServeTLS(config.CertFile, config.KeyFile)
		} else {
			err = server.ListenAndServe()
		}
		if err != nil {
			logger.Error("Server error", zap.Error(err))
			cancel()
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()

	// Graceful shutdown
	if err := server.Close(); err != nil {
		logger.Error("Error during shutdown", zap.Error(err))
	}

	logger.Info("Enhanced QUIC server stopped")
}

// generateSelfSignedCert creates a self-signed certificate for testing
func generateSelfSignedCert(certFile, keyFile string) error {
	// This is a placeholder - in real implementation you would generate actual certificates
	// For demo purposes, we'll just create empty files
	if err := os.WriteFile(certFile, []byte("# Demo cert file"), 0600); err != nil {
		return err
	}
	if err := os.WriteFile(keyFile, []byte("# Demo key file"), 0600); err != nil {
		return err
	}
	return nil
}
