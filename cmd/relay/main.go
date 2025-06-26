package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/config"
	"github.com/twogc/cloudbridge-relay-installer/internal/django"
	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
	"github.com/twogc/cloudbridge-relay-installer/internal/metrics"
	"github.com/twogc/cloudbridge-relay-installer/internal/proto"
	"github.com/twogc/cloudbridge-relay-installer/internal/ratelimit"
	"github.com/twogc/cloudbridge-relay-installer/internal/token"
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
	lastPing  time.Time
	limiter   *ratelimit.Limiter
}

// Tunnel represents a tunnel connection
type Tunnel struct {
	conn      net.Conn
	token     string
	lastPing  time.Time
	limiter   *ratelimit.Limiter
	tunnelInfo *proto.TunnelInfo
}

// Server represents the relay server
type Server struct {
	port           int
	tokenValidator *token.Validator
	limiter        *ratelimit.IPLimiter
	maxConnections int
	connCount      int32
	metrics        *metrics.Metrics
	mu             sync.RWMutex
	connections    map[string]*Connection
	listener       net.Listener
	tunnels        map[string]*Tunnel
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

	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		port:           cfg.Server.Port,
		tokenValidator: tokenValidator,
		limiter:        limiter,
		maxConnections: cfg.Server.MaxTunnels,
		metrics:        metrics.NewMetrics(prometheus.DefaultRegisterer),
		connections:    make(map[string]*Connection),
		tunnels:        make(map[string]*Tunnel),
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
		metricsServer := metrics.NewServer(cfg.Metrics.Port, cfg.Metrics.Path)
		go func() {
			if err := metricsServer.Start(); err != nil {
				s.logger.Error("Failed to start metrics server", zap.Error(err))
			}
		}()
	}

	// Create listener
	var err error
	s.listener, err = net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("failed to create listener: %v", err)
	}

	s.logger.Info("Server started", 
		zap.Int("port", s.port),
		zap.Bool("tls", *config.UseTLS),
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
	clientID, err := s.tokenValidator.ValidateToken(msg.Token)
	if err != nil {
		s.logger.Error("Invalid token", zap.Error(err))
		proto.WriteError(writer, "invalid_token", "Invalid or expired token")
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

// handleTunnelInfo handles tunnel info messages
func (s *Server) handleTunnelInfo(writer *bufio.Writer, connection *Connection, tunnelInfo *proto.TunnelInfo) {
	if tunnelInfo == nil {
		s.logger.Error("Tunnel info is nil", zap.Error(fmt.Errorf("tunnel info is nil")))
		proto.WriteError(writer, "invalid_tunnel_info", "Tunnel info is required")
		return
	}

	// Создаем туннель
	tunnel := &Tunnel{
		conn:       connection.conn,
		token:      connection.token,
		lastPing:   time.Now(),
		limiter:    ratelimit.NewLimiter(*rateLimit, *rateBurst),
		tunnelInfo: tunnelInfo,
	}

	// Добавляем в список туннелей
	s.mu.Lock()
	s.tunnels[tunnelInfo.TunnelID] = tunnel
	s.mu.Unlock()

	s.metrics.SetActiveTunnels(len(s.tunnels))

	// Запускаем обработку туннеля
	go s.handleTunnel(tunnel)

	s.logger.Info("Tunnel created", 
		zap.String("tunnel_id", tunnelInfo.TunnelID),
		zap.Int("local_port", tunnelInfo.LocalPort),
		zap.String("remote_host", tunnelInfo.RemoteHost),
		zap.Int("remote_port", tunnelInfo.RemotePort),
	)
}

// handleTunnel handles a tunnel connection
func (s *Server) handleTunnel(tunnel *Tunnel) {
	defer func() {
		s.mu.Lock()
		delete(s.tunnels, tunnel.tunnelInfo.TunnelID)
		s.mu.Unlock()
		s.metrics.SetActiveTunnels(len(s.tunnels))
	}()

	// Здесь должна быть логика обработки туннеля
	// Пока просто логируем
	s.logger.Info("Tunnel started", 
		zap.String("tunnel_id", tunnel.tunnelInfo.TunnelID),
	)

	// Бесконечный цикл для поддержания туннеля
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-time.After(30 * time.Second):
			// Проверяем активность туннеля
			if time.Since(tunnel.lastPing) > 60*time.Second {
				s.logger.Warn("Tunnel inactive", 
					zap.String("tunnel_id", tunnel.tunnelInfo.TunnelID),
				)
				return
			}
		}
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

// cleanup removes inactive connections and tunnels
func (s *Server) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	// Cleanup connections
	for clientID, connection := range s.connections {
		if now.Sub(connection.lastPing) > 5*time.Minute {
			connection.conn.Close()
			delete(s.connections, clientID)
			s.logger.Info("Cleaned up inactive connection", 
				zap.String("client_id", clientID),
			)
		}
	}

	// Cleanup tunnels
	for tunnelID, tunnel := range s.tunnels {
		if now.Sub(tunnel.lastPing) > 5*time.Minute {
			tunnel.conn.Close()
			delete(s.tunnels, tunnelID)
			s.logger.Info("Cleaned up inactive tunnel", 
				zap.String("tunnel_id", tunnelID),
			)
		}
	}

	s.metrics.SetActiveTunnels(len(s.tunnels))
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
	for _, tunnel := range s.tunnels {
		tunnel.conn.Close()
	}
	s.mu.Unlock()

	// Wait for context cancellation or timeout
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(30 * time.Second):
		return fmt.Errorf("shutdown timeout")
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