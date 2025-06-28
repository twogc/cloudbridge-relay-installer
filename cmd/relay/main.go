package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/config"
	"github.com/twogc/cloudbridge-relay-installer/internal/django"
	"github.com/twogc/cloudbridge-relay-installer/internal/keycloak"
	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
	"github.com/twogc/cloudbridge-relay-installer/internal/metrics"
	"github.com/twogc/cloudbridge-relay-installer/internal/proto"
	"github.com/twogc/cloudbridge-relay-installer/internal/ratelimit"
	"github.com/twogc/cloudbridge-relay-installer/internal/token"
	"github.com/twogc/cloudbridge-relay-installer/internal/tunnel"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

var (
	configFile   = flag.String("config", "config/config.json", "Path to configuration file")
	workerPoolSize = flag.Int("workers", runtime.NumCPU()*2, "Number of worker goroutines")
	timeout      = flag.Duration("timeout", 30*time.Second, "Connection timeout")
	logLevel     = flag.String("log-level", "INFO", "Log level (DEBUG, INFO, WARNING, ERROR)")
	logFile      = flag.String("log-file", "", "Log file path (empty for stdout)")
	rateLimit    = flag.Float64("rate-limit", 100, "Rate limit per second")
	rateBurst    = flag.Int("rate-burst", 200, "Rate limit burst size")
)

// Connection represents a client connection
type Connection struct {
	conn      net.Conn
	token     string
	clientID  string
	lastPing  time.Time
	limiter   *ratelimit.Limiter
}

// Server represents the relay server
type Server struct {
	port           int
	tokenValidator *token.Validator
	keycloakClient *keycloak.Client
	limiter        *ratelimit.IPLimiter
	userLimiter    *ratelimit.UserLimiter
	auditLogger    *logger.AuditLogger
	maxConnections int
	connCount      int32
	metrics        *metrics.Metrics
	tunnelManager  *tunnel.Manager
	mu             sync.RWMutex
	connections    map[string]*Connection
	listener       net.Listener
	logger         *logger.Logger
	workerPool     chan struct{}
	ctx            context.Context
	cancel         context.CancelFunc
}

// NewServer creates a new server instance
func NewServer(cfg *config.Config) (*Server, error) {
	// Setup logging
	if *logFile != "" {
		if _, err := os.OpenFile(*logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err != nil {
			return nil, fmt.Errorf("failed to open log file: %v", err)
		}
	}

	// Определяем output в зависимости от настроек
	output := "stdout"
	if *logFile != "" {
		output = "both"
	}

	loggerInstance, err := logger.New(logger.Config{
		Level:      logger.LogLevel(*logLevel),
		Output:     output,
		FilePath:   *logFile,
		JSON:       true,
		MaxSize:    100,
		MaxAge:     30,
		MaxBackups: 10,
		Compress:   true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %v", err)
	}

	// Create Django client
	djangoClient := django.NewClient(django.Config{
		APIEndpoint: cfg.Django.APIEndpoint,
		JWTSecret:   cfg.Django.JWTSecret,
		Timeout:     cfg.Django.RESTTimeout,
	})

	// Create Keycloak client if enabled
	var keycloakClient *keycloak.Client
	if cfg.Keycloak.Enabled {
		keycloakClient, err = keycloak.NewClient(&keycloak.Config{
			ServerURL:    cfg.Keycloak.ServerURL,
			Realm:        cfg.Keycloak.Realm,
			ClientID:     cfg.Keycloak.ClientID,
			ClientSecret: cfg.Keycloak.ClientSecret,
			PublicKey:    cfg.Keycloak.PublicKey,
			TokenValidation: struct {
				Issuer     string        `json:"issuer"`
				Audience   string        `json:"audience"`
				ClockSkew  time.Duration `json:"clock_skew"`
			}{
				Issuer:     cfg.Keycloak.TokenValidation.Issuer,
				Audience:   cfg.Keycloak.TokenValidation.Audience,
				ClockSkew:  cfg.Keycloak.TokenValidation.ClockSkew,
			},
			HTTPClient: &http.Client{Timeout: cfg.Keycloak.HTTPTimeout},
			Logger:     loggerInstance,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create Keycloak client: %v", err)
		}
	}

	// Create token validator
	tokenValidator := token.NewValidator(token.Config{
		DjangoClient: djangoClient,
		JWTSecret:    cfg.Django.JWTSecret,
		CacheTTL:     cfg.Django.RESTCacheTTL,
	})

	// Create rate limiter
	limiter := ratelimit.NewIPLimiter(
		cfg.Security.RateLimit.RequestsPerSecond,
		cfg.Security.RateLimit.Burst,
	)

	// Create user-based rate limiter
	userLimiter := ratelimit.NewUserLimiter(
		cfg.Security.RateLimit.RequestsPerSecond,
		cfg.Security.RateLimit.Burst,
	)

	// Create audit logger
	var auditLogger *logger.AuditLogger
	if cfg.Security.Audit.Enabled {
		auditLogger, err = logger.NewAuditLogger(logger.Config{
			Level:      logger.LogLevel(*logLevel),
			Output:     "file", // Audit logs go to file by default
			FilePath:   cfg.Security.Audit.LogFile,
			JSON:       true,
			MaxSize:    100,    // Default max size in MB
			MaxAge:     7,      // Default max age in days
			MaxBackups: 10,     // Default max backups
			Compress:   true,   // Default compress
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create audit logger: %v", err)
		}
	}

	// Create metrics
	metricsInstance := metrics.NewMetrics(prometheus.DefaultRegisterer)

	// Create tunnel manager
	tunnelManager := tunnel.NewManager(cfg.Server.MaxTunnels, loggerInstance, metricsInstance)

	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		port:           cfg.Server.Port,
		tokenValidator: tokenValidator,
		keycloakClient: keycloakClient,
		limiter:        limiter,
		userLimiter:    userLimiter,
		auditLogger:    auditLogger,
		maxConnections: cfg.Server.MaxTunnels,
		metrics:        metricsInstance,
		tunnelManager:  tunnelManager,
		connections:    make(map[string]*Connection),
		logger:         loggerInstance,
		workerPool:     make(chan struct{}, runtime.NumCPU()*2),
		ctx:            ctx,
		cancel:         cancel,
	}, nil
}

// isIPAllowed checks if the IP address is in the allowed list
func isIPAllowed(ip string, allowedIPs string) bool {
	if allowedIPs == "" {
		return true
	}
	
	ipList := strings.Split(allowedIPs, ",")
	for _, allowedIP := range ipList {
		if strings.TrimSpace(allowedIP) == ip {
			return true
		}
	}
	return false
}

// Start starts the server
func (s *Server) Start(cfg *config.Config) error {
	// Start metrics server if enabled
	if cfg.Metrics.Enabled {
		metricsServer := metrics.NewServer(fmt.Sprintf(":%d", cfg.Metrics.Port), cfg.Metrics.Path)
		go func() {
			if err := metricsServer.Start(); err != nil {
				s.logger.Error("Failed to start metrics server", zap.Error(err))
			}
		}()
	}

	// Start tunnel manager cleanup loop
	s.tunnelManager.StartCleanupLoop(1*time.Minute, 5*time.Minute)

	// Create listener
	var err error
	
	// Check if TLS is enabled
	if cfg.Security.TLSEnabled {
		// Create TLS configuration
		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{},
			MinVersion:   tls.VersionTLS12,
		}
		
		// Load certificate if files are specified
		if cfg.Security.CertFile != "" && cfg.Security.KeyFile != "" {
			cert, err := tls.LoadX509KeyPair(cfg.Security.CertFile, cfg.Security.KeyFile)
			if err != nil {
				s.logger.Error("Failed to load TLS certificate", 
					zap.String("cert_file", cfg.Security.CertFile),
					zap.String("key_file", cfg.Security.KeyFile),
					zap.Error(err),
				)
				return fmt.Errorf("failed to load TLS certificate: %v", err)
			}
			tlsConfig.Certificates = []tls.Certificate{cert}
		}
		
		// Create TCP listener first
		tcpListener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
		if err != nil {
			return fmt.Errorf("failed to create TCP listener: %v", err)
		}
		
		// Wrap with TLS listener
		s.listener = tls.NewListener(tcpListener, tlsConfig)
		s.logger.Info("TLS listener created", 
			zap.String("cert_file", cfg.Security.CertFile),
			zap.String("key_file", cfg.Security.KeyFile),
		)
	} else {
		// Create regular TCP listener
		s.listener, err = net.Listen("tcp", fmt.Sprintf(":%d", s.port))
		if err != nil {
			return fmt.Errorf("failed to create listener: %v", err)
		}
	}

	s.logger.Info("Server started", 
		zap.Int("port", s.port),
		zap.Bool("tls", cfg.Security.TLSEnabled),
		zap.Int("workers", *workerPoolSize),
		zap.Duration("timeout", *timeout),
		zap.String("logLevel", *logLevel),
	)

	// Start cleanup loop
	go s.cleanupLoop()

	// Start accepting connections
	for {
		select {
		case <-s.ctx.Done():
			return nil
		default:
			conn, err := s.listener.Accept()
			if err != nil {
				s.logger.Error("Failed to accept connection", zap.Error(err))
				continue
			}

			// Check connection limit
			if atomic.LoadInt32(&s.connCount) >= int32(s.maxConnections) {
				s.logger.Warn("Connection limit reached", 
					zap.Int("limit", s.maxConnections),
				)
				conn.Close()
				s.metrics.IncRejectedConnections()
				continue
			}

			// Check rate limit
			remoteAddr := conn.RemoteAddr().String()
			if !s.limiter.Allow(remoteAddr) {
				s.logger.Warn("Rate limit exceeded", 
					zap.String("remote_addr", remoteAddr),
				)
				conn.Close()
				s.metrics.IncRejectedConnections()
				continue
			}

			// Acquire worker slot
			select {
			case s.workerPool <- struct{}{}:
				atomic.AddInt32(&s.connCount, 1)
				s.metrics.IncConnections()
				go func() {
					defer func() {
						<-s.workerPool // Release worker slot
						atomic.AddInt32(&s.connCount, -1)
						s.metrics.DecConnections()
					}()
					s.handleConnection(conn)
				}()
			default:
				// No available workers, reject connection
				s.logger.Warn("No available workers", 
					zap.String("remote_addr", remoteAddr),
				)
				conn.Close()
				s.metrics.IncRejectedConnections()
			}
		}
	}
}

// handleConnection handles a client connection
func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Audit TLS connection info if TLS is enabled
	if tlsConn, ok := conn.(*tls.Conn); ok {
		state := tlsConn.ConnectionState()
		if s.auditLogger != nil {
			s.auditLogger.AuditSecurity(
				"tls_connection_established",
				"unknown", // userID not known yet
				conn.RemoteAddr().String(),
				"info",
				map[string]interface{}{
					"tls_version":     getTLSVersionString(state.Version),
					"cipher_suite":    getCipherSuiteString(state.CipherSuite),
					"server_name":     state.ServerName,
					"negotiated_protocol": state.NegotiatedProtocol,
				},
			)
		}
	}

	// Устанавливаем таймауты
	conn.SetDeadline(time.Now().Add(*timeout))

	// Создаем буферизованные reader/writer
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	// Отправляем hello
	if err := proto.WriteHello(writer, "1.0", []string{"tls", "heartbeat", "tunnel_info"}); err != nil {
		s.logger.Error("Failed to send hello", zap.Error(err))
		return
	}

	// Читаем auth
	msg, err := proto.ReadMessage(reader)
	if err != nil {
		s.logger.Error("Failed to read auth message", zap.Error(err))
		return
	}

	if msg.Type != proto.MessageTypeAuth {
		s.logger.Error("Expected auth message", zap.Error(fmt.Errorf("got %s", msg.Type)))
		return
	}

	// Валидируем токен
	authStart := time.Now()
	var clientID string
	
	// Try Keycloak validation first if available
	if s.keycloakClient != nil {
		tokenInfo, err := s.keycloakClient.ValidateToken(context.Background(), msg.Token)
		if err == nil {
			clientID = tokenInfo.Subject
			s.metrics.IncKeycloakTokenValidations("success", "keycloak")
		} else {
			s.logger.Warn("Keycloak validation failed, falling back to Django", zap.Error(err))
			s.metrics.IncKeycloakTokenValidations("failure", "keycloak")
			// Fall back to Django validation
			clientID, err = s.tokenValidator.ValidateToken(msg.Token)
			if err != nil {
				s.logger.Error("Both Keycloak and Django validation failed", zap.Error(err))
				proto.WriteError(writer, "invalid_token", "Invalid or expired token")
				return
			}
		}
	} else {
		// Use Django validation only
		clientID, err = s.tokenValidator.ValidateToken(msg.Token)
		if err != nil {
			s.logger.Error("Invalid token", zap.Error(err))
			proto.WriteError(writer, "invalid_token", "Invalid or expired token")
			return
		}
	}
	
	authDuration := time.Since(authStart)

	// Audit successful authentication
	if s.auditLogger != nil {
		s.auditLogger.AuditAuthentication(
			clientID,
			"jwt",
			conn.RemoteAddr().String(),
			true,
			authDuration,
			map[string]interface{}{
				"token_length": len(msg.Token),
			},
		)
	}

	// Rate limiting per user (clientID)
	if !s.userLimiter.Allow(clientID) {
		s.logger.Warn("User rate limit exceeded", zap.String("client_id", clientID))
		// Audit rate limit event
		if s.auditLogger != nil {
			s.auditLogger.AuditRateLimit(
				clientID,
				conn.RemoteAddr().String(),
				"user",
				map[string]interface{}{
					"rate_limit": s.userLimiter.Rate,
					"burst": s.userLimiter.Burst,
				},
			)
		}
		proto.WriteError(writer, "rate_limit_exceeded", "Rate limit exceeded for user")
		return
	}

	// Отправляем auth_response
	if err := proto.WriteAuthResponse(writer, "ok", clientID); err != nil {
		s.logger.Error("Failed to send auth response", zap.Error(err))
		return
	}

	// Создаем соединение
	connection := &Connection{
		conn:     conn,
		token:    msg.Token,
		clientID: clientID,
		lastPing: time.Now(),
		limiter:  ratelimit.NewLimiter(*rateLimit, *rateBurst),
	}

	// Добавляем в список соединений
	s.mu.Lock()
	s.connections[clientID] = connection
	s.mu.Unlock()

	// Обрабатываем сообщения
	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			msg, err := proto.ReadMessage(reader)
			if err != nil {
				s.logger.Error("Failed to read message", zap.Error(err))
				return
			}

			switch msg.Type {
			case proto.MessageTypeHeartbeat:
				s.handleHeartbeat(writer, connection)
			case proto.MessageTypeTunnelInfo:
				s.handleTunnelInfo(writer, connection, msg.TunnelInfo)
			default:
				s.logger.Error("Unknown message type", zap.Error(fmt.Errorf("got %s", msg.Type)))
				proto.WriteError(writer, "unknown_message_type", "Unknown message type")
			}
		}
	}
}

// handleHeartbeat handles heartbeat messages
func (s *Server) handleHeartbeat(writer *bufio.Writer, connection *Connection) {
	connection.lastPing = time.Now()
	s.metrics.IncHeartbeats()

	if err := proto.WriteHeartbeatResponse(writer); err != nil {
		s.logger.Error("Failed to send heartbeat response", zap.Error(err))
	}
}

// handleTunnelInfo handles tunnel info messages and creates TCP proxy
func (s *Server) handleTunnelInfo(writer *bufio.Writer, connection *Connection, tunnelInfo *proto.TunnelInfo) {
	if tunnelInfo == nil {
		s.logger.Error("Tunnel info is nil", zap.Error(fmt.Errorf("tunnel info is nil")))
		proto.WriteError(writer, "invalid_tunnel_info", "Tunnel info is required")
		return
	}

	// Audit tunnel creation
	if s.auditLogger != nil {
		s.auditLogger.AuditTunnelCreation(
			connection.clientID,
			tunnelInfo.TunnelID,
			connection.conn.RemoteAddr().String(),
			map[string]interface{}{
				"local_port": tunnelInfo.LocalPort,
				"remote_host": tunnelInfo.RemoteHost,
				"remote_port": tunnelInfo.RemotePort,
			},
			map[string]interface{}{
				"connection_id": connection.conn.RemoteAddr().String(),
			},
		)
	}

	// Create TCP proxy tunnel
	_, err := s.tunnelManager.CreateTunnel(
		tunnelInfo.TunnelID,
		connection.conn,
		tunnelInfo.RemoteHost,
		tunnelInfo.RemotePort,
	)
	if err != nil {
		s.logger.Error("Failed to create tunnel",
			zap.String("tunnel_id", tunnelInfo.TunnelID),
			zap.Error(err),
		)
		proto.WriteError(writer, "tunnel_creation_failed", err.Error())
		return
	}

	s.logger.Info("Tunnel created successfully", 
		zap.String("tunnel_id", tunnelInfo.TunnelID),
		zap.Int("local_port", tunnelInfo.LocalPort),
		zap.String("remote_host", tunnelInfo.RemoteHost),
		zap.Int("remote_port", tunnelInfo.RemotePort),
	)

	// Send success response
	if err := proto.WriteTunnelInfo(writer, tunnelInfo); err != nil {
		s.logger.Error("Failed to send tunnel response", zap.Error(err))
	}
}

// cleanupLoop periodically cleans up inactive connections
func (s *Server) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.cleanup()
		}
	}
}

// cleanup removes inactive connections
func (s *Server) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	// Cleanup connections
	for clientID, connection := range s.connections {
		if now.Sub(connection.lastPing) > 5*time.Minute {
			// Audit connection cleanup
			if s.auditLogger != nil {
				s.auditLogger.AuditEvent(
					"connection_cleanup",
					connection.clientID,
					connection.conn.RemoteAddr().String(),
					map[string]interface{}{
						"reason": "inactive_timeout",
						"inactive_duration": now.Sub(connection.lastPing).String(),
					},
				)
			}
			connection.conn.Close()
			delete(s.connections, clientID)
			s.logger.Info("Cleaned up inactive connection", 
				zap.String("client_id", clientID),
			)
		}
	}
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down server")

	// Cancel context to stop accepting new connections
	s.cancel()

	// Close listener
	if s.listener != nil {
		s.listener.Close()
	}

	// Close all connections
	s.mu.Lock()
	for _, connection := range s.connections {
		connection.conn.Close()
	}
	s.mu.Unlock()

	// Shutdown tunnel manager
	s.tunnelManager.Shutdown()

	s.logger.Info("Server shutdown completed")
	return nil
}

// Helper functions for TLS audit logging
func getTLSVersionString(version uint16) string {
	switch version {
	case tls.VersionTLS10:
		return "1.0"
	case tls.VersionTLS11:
		return "1.1"
	case tls.VersionTLS12:
		return "1.2"
	case tls.VersionTLS13:
		return "1.3"
	default:
		return "unknown"
	}
}

func getCipherSuiteString(suite uint16) string {
	switch suite {
	case tls.TLS_AES_256_GCM_SHA384:
		return "TLS_AES_256_GCM_SHA384"
	case tls.TLS_CHACHA20_POLY1305_SHA256:
		return "TLS_CHACHA20_POLY1305_SHA256"
	case tls.TLS_AES_128_GCM_SHA256:
		return "TLS_AES_128_GCM_SHA256"
	case tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384:
		return "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384"
	case tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256:
		return "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"
	default:
		return fmt.Sprintf("unknown_%d", suite)
	}
}

func main() {
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid configuration: %v\n", err)
		os.Exit(1)
	}

	// Create server
	server, err := NewServer(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create server: %v\n", err)
		os.Exit(1)
	}

	// Start server
	if err := server.Start(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start server: %v\n", err)
		os.Exit(1)
	}
} 