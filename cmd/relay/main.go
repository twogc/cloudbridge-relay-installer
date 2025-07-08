package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/twogc/cloudbridge-relay-installer/config"
	"github.com/twogc/cloudbridge-relay-installer/internal/circuitbreaker"
	"github.com/twogc/cloudbridge-relay-installer/internal/health"
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

// Server represents the main relay server
type Server struct {
	config         *config.Config
	logger         *logger.Logger
	listener       net.Listener
	tunnelManager  *tunnel.Manager
	metricsServer  *metrics.Server
	
	// Zero Trust components
	circuitBreaker *circuitbreaker.CircuitBreaker
	healthChecker  *health.HealthChecker
	
	ctx    context.Context
	cancel context.CancelFunc
}

// NewServer creates a new relay server
func NewServer(cfg *config.Config) (*Server, error) {
	ctx, cancel := context.WithCancel(context.Background())
	
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
		cancel()
		return nil, fmt.Errorf("failed to initialize logger: %v", err)
	}
	
	// Initialize tunnel manager
	tunnelManager := tunnel.NewManager(cfg.Server.MaxTunnels, loggerInstance, nil)
	
	// Initialize Zero Trust components
	var circuitBreaker *circuitbreaker.CircuitBreaker
	if cfg.ZeroTrust.CircuitBreaker.Enabled {
		// Convert logger.Logger to *zap.Logger for circuit breaker
		zapLogger := zap.NewNop() // Fallback to no-op logger
		
		circuitBreaker = circuitbreaker.NewCircuitBreaker(&circuitbreaker.Config{
			Name:             "relay-server",
			FailureThreshold: cfg.ZeroTrust.CircuitBreaker.FailureThreshold,
			SuccessThreshold: cfg.ZeroTrust.CircuitBreaker.SuccessThreshold,
			Timeout:          cfg.ZeroTrust.CircuitBreaker.Timeout,
			WindowSize:       cfg.ZeroTrust.CircuitBreaker.WindowSize,
			Logger:           zapLogger,
		})
	}
	
	var healthChecker *health.HealthChecker
	if cfg.ZeroTrust.HealthCheck.Enabled {
		// Convert logger.Logger to *zap.Logger for health checker
		zapLogger := zap.NewNop() // Fallback to no-op logger
		
		healthChecker = health.NewHealthChecker(&health.Config{
			Interval: cfg.ZeroTrust.HealthCheck.Interval,
			Timeout:  cfg.ZeroTrust.HealthCheck.Timeout,
			Logger:   zapLogger,
		})
		
		// Add configured health checks
		for _, checkConfig := range cfg.ZeroTrust.HealthCheck.Checks {
			switch checkConfig.Type {
			case "http":
				if checkConfig.URL != "" {
					healthChecker.AddCheck(checkConfig.Name, health.HTTPHealthCheck(checkConfig.Name, checkConfig.URL))
				}
			case "ping":
				if checkConfig.Host != "" {
					healthChecker.AddCheck(checkConfig.Name, health.PingHealthCheck(checkConfig.Name, checkConfig.Host))
				}
			case "custom":
				// Custom health checks can be added here
				healthChecker.AddCheck(checkConfig.Name, health.CustomHealthCheck(
					checkConfig.Name,
					checkConfig.Description,
					func(ctx context.Context) error {
						// Default custom check - always healthy
						return nil
					},
				))
			}
		}
		
		// Start health checker
		healthChecker.Start()
	}
	
	// Initialize metrics server
	var metricsServer *metrics.Server
	if cfg.Metrics.Enabled {
		metricsServer = metrics.NewServer(fmt.Sprintf(":%d", cfg.Metrics.Port), cfg.Metrics.Path)
	}
	
	return &Server{
		config:         cfg,
		logger:         loggerInstance,
		tunnelManager:  tunnelManager,
		metricsServer:  metricsServer,
		circuitBreaker: circuitBreaker,
		healthChecker:  healthChecker,
		ctx:            ctx,
		cancel:         cancel,
	}, nil
}

// Start starts the server
func (s *Server) Start() error {
	// Start metrics server
	if s.metricsServer != nil {
		go func() {
			if err := s.metricsServer.Start(); err != nil {
				s.logger.Error("Failed to start metrics server", zap.Error(err))
			}
		}()
	}
	
	// Start listening
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.Server.Port))
	if err != nil {
		return fmt.Errorf("failed to start listener: %v", err)
	}
	s.listener = listener
	
	s.logger.Info("Relay server started",
		zap.Int("port", s.config.Server.Port),
		zap.Bool("zero_trust_enabled", s.config.ZeroTrust.Enabled),
		zap.Bool("circuit_breaker_enabled", s.config.ZeroTrust.CircuitBreaker.Enabled),
		zap.Bool("health_check_enabled", s.config.ZeroTrust.HealthCheck.Enabled),
	)
	
	// Handle connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			if s.ctx.Err() != nil {
				return nil // Server is shutting down
			}
			s.logger.Error("Failed to accept connection", zap.Error(err))
			continue
		}
		
		go s.handleConnection(conn)
	}
}

// handleConnection handles a new connection
func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	
	// Use circuit breaker if enabled
	if s.circuitBreaker != nil {
		err := s.circuitBreaker.Execute(s.ctx, func() error {
			return s.processConnection(conn)
		})
		if err != nil {
			s.logger.Error("Circuit breaker blocked connection", zap.Error(err))
			return
		}
	} else {
		if err := s.processConnection(conn); err != nil {
			s.logger.Error("Failed to process connection", zap.Error(err))
		}
	}
}

// processConnection processes a connection
func (s *Server) processConnection(conn net.Conn) error {
	// Example: Check if QUIC is enabled
	if s.config.QUIC.Enabled {
		s.logger.Debug("QUIC enabled for connection")
	}
	
	// Simple echo for demonstration
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		return err
	}
	
	_, err = conn.Write(buffer[:n])
	return err
}

// Stop stops the server
func (s *Server) Stop() {
	s.logger.Info("Stopping relay server...")
	
	// Stop health checker
	if s.healthChecker != nil {
		s.healthChecker.Stop()
	}
	
	// Stop metrics server
	if s.metricsServer != nil {
		// Note: metrics.Server doesn't have Stop method, it will stop when context is cancelled
	}
	
	// Close listener
	if s.listener != nil {
		if err := s.listener.Close(); err != nil {
			s.logger.Error("Error closing listener", zap.Error(err))
		}
	}
	
	// Cancel context
	s.cancel()
	
	s.logger.Info("Relay server stopped")
}

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
	
	// Create server
	server, err := NewServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}
	
	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	go func() {
		<-sigChan
		server.Stop()
	}()
	
	// Start server
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
} 