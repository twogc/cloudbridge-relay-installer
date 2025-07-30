package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/twogc/cloudbridge-relay-installer/config"
	"github.com/twogc/cloudbridge-relay-installer/internal/auth"
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
	
	// Prometheus metrics
	connectionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "relay_connections_total",
			Help: "Total number of connections",
		},
		[]string{"status"},
	)
	
	authAttemptsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "relay_auth_attempts_total",
			Help: "Total number of authentication attempts",
		},
		[]string{"status", "provider"},
	)
	
	tunnelsCreatedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "relay_tunnels_created_total",
			Help: "Total number of tunnels created",
		},
		[]string{"status"},
	)
	
	activeConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "relay_active_connections",
			Help: "Number of active connections",
		},
	)
)

// Server represents the main relay server
type Server struct {
	config         *config.Config
	logger         *logger.Logger
	listener       net.Listener
	httpServer     *http.Server
	tunnelManager  *tunnel.Manager
	metricsServer  *metrics.Server
	authAdapter    *auth.Adapter
	
	// Zero Trust components
	circuitBreaker *circuitbreaker.CircuitBreaker
	healthChecker  *health.HealthChecker
	
	ctx    context.Context
	cancel context.CancelFunc
	startTime time.Time
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
	
	// Initialize auth adapter
	authAdapter, err := auth.NewAdapter(cfg, loggerInstance)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize auth adapter: %v", err)
	}
	
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
		authAdapter:    authAdapter,
		circuitBreaker: circuitBreaker,
		healthChecker:  healthChecker,
		ctx:            ctx,
		cancel:         cancel,
		startTime:      time.Now(),
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
	
	connectionsTotal.WithLabelValues("accepted").Inc()
	activeConnections.Inc()
	defer activeConnections.Dec()
	
	// Use circuit breaker if enabled
	if s.circuitBreaker != nil {
		err := s.circuitBreaker.Execute(s.ctx, func() error {
			return s.processConnection(conn)
		})
		if err != nil {
			s.logger.Error("Circuit breaker blocked connection", zap.Error(err))
			connectionsTotal.WithLabelValues("blocked").Inc()
			return
		}
	} else {
		if err := s.processConnection(conn); err != nil {
			s.logger.Error("Failed to process connection", zap.Error(err))
			connectionsTotal.WithLabelValues("failed").Inc()
		}
	}
	
	connectionsTotal.WithLabelValues("completed").Inc()
}

// processConnection processes a connection
func (s *Server) processConnection(conn net.Conn) error {
    reader := bufio.NewReader(conn)
    writer := bufio.NewWriter(conn)

    // 1. Прочитать первое сообщение (ожидаем hello)
    line, err := reader.ReadString('\n')
    if err != nil {
        s.logger.Error("Failed to read from connection", zap.Error(err))
        return err
    }
    s.logger.Info("Received raw message", zap.String("data", line))

    // Check if it's an HTTP request (healthcheck)
    if len(line) > 0 && (strings.HasPrefix(line, "GET ") || strings.HasPrefix(line, "POST ") || 
        strings.HasPrefix(line, "PUT ") || strings.HasPrefix(line, "DELETE ") || 
        strings.HasPrefix(line, "HEAD ") || strings.HasPrefix(line, "OPTIONS ") ||
        strings.HasPrefix(line, "HTTP/")) {
        s.logger.Debug("Received HTTP request, handling as HTTP", zap.String("data", line))
        return s.handleHTTPRequest(conn, line, reader)
    }
    
    // Additional check for HTTP requests that might not start with method
    if len(line) > 0 && (strings.Contains(line, "HTTP/") || strings.Contains(line, "Host:") || 
        strings.Contains(line, "User-Agent:")) {
        s.logger.Debug("Detected HTTP request by content, handling as HTTP", zap.String("data", line))
        return s.handleHTTPRequest(conn, line, reader)
    }

    var msg map[string]interface{}
    if err := json.Unmarshal([]byte(line), &msg); err != nil {
        s.logger.Error("Failed to parse JSON", zap.Error(err))
        return err
    }
    msgType, _ := msg["type"].(string)
    s.logger.Info("Received message", zap.String("type", msgType))

    if msgType != "hello" {
        s.logger.Warn("Expected hello message", zap.String("type", msgType))
        return nil
    }

    // 2. Отправить hello-ответ
    respHello := map[string]interface{}{
        "type":     "hello",
        "version":  "1.0.0",
        "features": []string{"tls", "jwt", "tunneling"},
    }
    if err := writeJSON(writer, respHello); err != nil {
        s.logger.Error("Failed to send hello response", zap.Error(err))
        return err
    }
    s.logger.Info("Sent hello response")

    // 3. Прочитать auth
    line, err = reader.ReadString('\n')
    if err != nil {
        s.logger.Error("Failed to read auth", zap.Error(err))
        return err
    }
    s.logger.Info("Received raw auth", zap.String("data", line))
    msg = map[string]interface{}{}
    if err := json.Unmarshal([]byte(line), &msg); err != nil {
        s.logger.Error("Failed to parse auth JSON", zap.Error(err))
        return err
    }
    msgType, _ = msg["type"].(string)
    if msgType != "auth" {
        s.logger.Warn("Expected auth message", zap.String("type", msgType))
        return nil
    }
    token, _ := msg["token"].(string)
    if token == "" {
        resp := map[string]interface{}{
            "type":    "auth_response",
            "status":  "error",
            "message": "Token required",
        }
        writeJSON(writer, resp)
        s.logger.Warn("Auth failed: token required")
        return nil
    }
    
    // Validate JWT token using auth adapter
    /*
    tokenInfo, err := s.authAdapter.ValidateToken(s.ctx, token)
    if err != nil {
        resp := map[string]interface{}{
            "type":    "auth_response",
            "status":  "error",
            "message": "Invalid token: " + err.Error(),
        }
        writeJSON(writer, resp)
        s.logger.Warn("Auth failed: invalid token", zap.Error(err))
        authAttemptsTotal.WithLabelValues("failed", "unknown").Inc()
        return nil
    }
    
    s.logger.Info("Token validated successfully", 
        zap.String("subject", tokenInfo.Subject),
        zap.String("username", tokenInfo.Username),
        zap.String("provider", tokenInfo.Provider),
    )
    
    authAttemptsTotal.WithLabelValues("success", tokenInfo.Provider).Inc()
    */
    
    // Temporary: accept any token for testing
    var tokenLog string
    if len(token) > 20 {
        tokenLog = token[:20] + "..."
    } else {
        tokenLog = token
    }
    s.logger.Info("Token accepted (validation disabled)", zap.String("token", tokenLog))
    authAttemptsTotal.WithLabelValues("success", "test").Inc()
    
    // Можно добавить валидацию токена
    respAuth := map[string]interface{}{
        "type":     "auth_response",
        "status":   "ok",
        "client_id": "relay-client-001",
    }
    if err := writeJSON(writer, respAuth); err != nil {
        s.logger.Error("Failed to send auth response", zap.Error(err))
        return err
    }
    s.logger.Info("Sent auth response")

    // 4. Ждем tunnel_info
    for {
        line, err = reader.ReadString('\n')
        if err != nil {
            s.logger.Error("Connection closed or error reading tunnel_info", zap.Error(err))
            break
        }
        s.logger.Info("Received raw tunnel_info", zap.String("data", line))
        msg = map[string]interface{}{}
        if err := json.Unmarshal([]byte(line), &msg); err != nil {
            s.logger.Error("Failed to parse tunnel_info JSON", zap.Error(err))
            continue
        }
        msgType, _ = msg["type"].(string)
        if msgType == "tunnel_info" {
            // Можно добавить обработку параметров туннеля
            respTunnel := map[string]interface{}{
                "type":       "tunnel_response",
                "status":     "ok",
                "tunnel_id":  "tunnel_001",
                "local_port": msg["local_port"],
                "remote_host": msg["remote_host"],
                "remote_port": msg["remote_port"],
                "protocol":   msg["protocol"],
            }
            if err := writeJSON(writer, respTunnel); err != nil {
                s.logger.Error("Failed to send tunnel response", zap.Error(err))
                tunnelsCreatedTotal.WithLabelValues("failed").Inc()
            } else {
                s.logger.Info("Sent tunnel response")
                tunnelsCreatedTotal.WithLabelValues("success").Inc()
            }
        } else {
            s.logger.Warn("Unknown message type", zap.String("type", msgType))
        }
    }
    return nil
}

// handleHTTPRequest handles HTTP requests (like health checks)
func (s *Server) handleHTTPRequest(conn net.Conn, firstLine string, reader *bufio.Reader) error {
    // Read the rest of the HTTP request headers
    var lines []string
    lines = append(lines, firstLine)
    
    // Read headers until empty line
    for {
        line, err := reader.ReadString('\n')
        if err != nil {
            // If we can't read more, just proceed with what we have
            break
        }
        lines = append(lines, line)
        if line == "\r\n" || line == "\n" {
            break // End of headers
        }
    }
    
    // Skip body for now - we don't need it for health checks and 404 responses
    
    // Parse the first line to get method and path
    parts := strings.Fields(firstLine)
    if len(parts) < 2 {
        // If we can't parse the request line, return 400 Bad Request
        errorResponse := `{"error": "Invalid request line"}`
        response := "HTTP/1.1 400 Bad Request\r\n"
        response += "Content-Type: application/json\r\n"
        response += "Connection: close\r\n"
        response += fmt.Sprintf("Content-Length: %d\r\n", len(errorResponse))
        response += "\r\n"
        response += errorResponse
        
        _, writeErr := conn.Write([]byte(response))
        if writeErr != nil {
            return writeErr
        }
        
        // Ensure all data is written before closing
        if tcpConn, ok := conn.(*net.TCPConn); ok {
            tcpConn.CloseWrite()
        }
        return nil
    }
    
    method := parts[0]
    path := parts[1]
    
    s.logger.Debug("HTTP request", zap.String("method", method), zap.String("path", path))
    
    // Handle OPTIONS request (CORS preflight)
    if method == "OPTIONS" {
        response := "HTTP/1.1 200 OK\r\n"
        response += "Access-Control-Allow-Origin: *\r\n"
        response += "Access-Control-Allow-Methods: GET, OPTIONS\r\n"
        response += "Access-Control-Allow-Headers: Content-Type\r\n"
        response += "Connection: close\r\n"
        response += "Content-Length: 0\r\n"
        response += "\r\n"
        
        _, writeErr := conn.Write([]byte(response))
        if writeErr != nil {
            return writeErr
        }
        
        // Ensure all data is written before closing
        if tcpConn, ok := conn.(*net.TCPConn); ok {
            tcpConn.CloseWrite()
        }
        return nil
    }
    
    // Handle health endpoint
    if path == "/health" {
        return s.handleHealthHTTP(conn)
    }
    
    // Return 404 for unknown paths
    errorResponse := `{"error": "Not Found"}`
    response := "HTTP/1.1 404 Not Found\r\n"
    response += "Content-Type: application/json\r\n"
    response += "Access-Control-Allow-Origin: *\r\n"
    response += "Connection: close\r\n"
    response += fmt.Sprintf("Content-Length: %d\r\n", len(errorResponse))
    response += "\r\n"
    response += errorResponse
    
    _, writeErr := conn.Write([]byte(response))
    if writeErr != nil {
        return writeErr
    }
    
    // Ensure all data is written before closing
    if tcpConn, ok := conn.(*net.TCPConn); ok {
        tcpConn.CloseWrite()
    }
    return nil
}

// handleHealthHTTP handles HTTP health check requests
func (s *Server) handleHealthHTTP(conn net.Conn) error {
    status := map[string]interface{}{
        "status":    "ok",
        "version":   s.config.API.Version,
        "timestamp": time.Now().Format(time.RFC3339),
        "service":   "cloudbridge-relay",
        "uptime":    time.Since(s.startTime).String(),
        "components": map[string]interface{}{
            "tcp_server":     "running",
            "metrics":        s.metricsServer != nil,
            "health_checker": s.healthChecker != nil,
            "circuit_breaker": s.circuitBreaker != nil,
        },
    }

    jsonData, err := json.Marshal(status)
    if err != nil {
        return err
    }

    response := "HTTP/1.1 200 OK\r\n"
    response += "Content-Type: application/json\r\n"
    response += "Access-Control-Allow-Origin: *\r\n"
    response += "Access-Control-Allow-Methods: GET, OPTIONS\r\n"
    response += "Access-Control-Allow-Headers: Content-Type\r\n"
    response += "Connection: close\r\n"
    response += fmt.Sprintf("Content-Length: %d\r\n", len(jsonData))
    response += "\r\n"
    response += string(jsonData)

    _, err = conn.Write([]byte(response))
    if err != nil {
        return err
    }
    
    // Ensure all data is written before closing
    if tcpConn, ok := conn.(*net.TCPConn); ok {
        tcpConn.CloseWrite()
    }
    return nil
}

func writeJSON(writer *bufio.Writer, msg interface{}) error {
    data, err := json.Marshal(msg)
    if err != nil {
        return err
    }
    data = append(data, '\n')
    if _, err := writer.Write(data); err != nil {
        return err
    }
    return writer.Flush()
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
	
	// Close HTTP server
	if s.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.httpServer.Shutdown(ctx); err != nil {
			s.logger.Error("Error shutting down HTTP server", zap.Error(err))
		}
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