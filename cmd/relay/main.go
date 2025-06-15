package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/maxlanies/cloudbridge-relay-installer/config"
	"github.com/maxlanies/cloudbridge-relay-installer/internal/django"
	"github.com/maxlanies/cloudbridge-relay-installer/internal/logger"
	"github.com/maxlanies/cloudbridge-relay-installer/internal/metrics"
	"github.com/maxlanies/cloudbridge-relay-installer/internal/proto"
	"github.com/maxlanies/cloudbridge-relay-installer/internal/ratelimit"
	"github.com/maxlanies/cloudbridge-relay-installer/internal/token"
)

var (
	configFile     = flag.String("config", "config/config.json", "Path to configuration file")
	workerPoolSize = flag.Int("workers", runtime.NumCPU()*2, "Number of worker goroutines")
	timeout        = flag.Duration("timeout", 30*time.Second, "Connection timeout")
	rateLimit      = flag.Int("rate-limit", 100, "Rate limit (requests per second)")
	rateBurst      = flag.Int("rate-burst", 200, "Rate limit burst")
	maxTunnels     = flag.Int("max-tunnels", 1000, "Maximum number of tunnels")
	logLevel       = flag.String("log-level", "INFO", "Log level (DEBUG, INFO, WARNING, ERROR)")
	logFile        = flag.String("log-file", "", "Log file path (default: stdout)")
)

// Connection represents a client connection
type Connection struct {
	conn      net.Conn
	token     string
	lastPing  time.Time
	limiter   *ratelimit.Limiter
}

// Tunnel represents a client tunnel
type Tunnel struct {
	conn      net.Conn
	token     string
	lastPing  time.Time
	limiter   *ratelimit.Limiter
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

// NewServer creates a new server
func NewServer(cfg *config.Config) (*Server, error) {
	// Create logger
	var logOutput *os.File
	var err error
	if *logFile != "" {
		logOutput, err = os.OpenFile(*logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %v", err)
		}
	}

	logger := logger.NewLogger(logOutput, logger.LogLevel(*logLevel))

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
		metrics:        metrics.NewMetrics(),
		connections:    make(map[string]*Connection),
		tunnels:        make(map[string]*Tunnel),
		logger:         logger,
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
				s.logger.Error("Failed to start metrics server", err, nil)
			}
		}()
	}

	// Create listener
	var err error
	s.listener, err = net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("failed to create listener: %v", err)
	}

	s.logger.Info("Server started", map[string]interface{}{
		"port":     s.port,
		"tls":      *config.UseTLS,
		"workers":  *workerPoolSize,
		"timeout":  *timeout,
		"logLevel": *logLevel,
	})

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
				s.logger.Error("Failed to accept connection", err, nil)
				continue
			}

			// Check connection limit
			if atomic.LoadInt32(&s.connCount) >= int32(s.maxConnections) {
				s.logger.Warning("Connection limit reached", map[string]interface{}{
					"limit": s.maxConnections,
				})
				conn.Close()
				s.metrics.IncRejectedConnections()
				continue
			}

			// Check rate limit
			remoteAddr := conn.RemoteAddr().String()
			if !s.limiter.Allow(remoteAddr) {
				s.logger.Warning("Rate limit exceeded", map[string]interface{}{
					"remote_addr": remoteAddr,
				})
				conn.Close()
				s.metrics.IncRejectedConnections()
				continue
			}

			// Acquire worker slot
			select {
			case s.workerPool <- struct{}{}:
				atomic.AddInt32(&s.connCount, 1)
				s.metrics.IncActiveConnections()
				go func() {
					defer func() {
						<-s.workerPool // Release worker slot
						atomic.AddInt32(&s.connCount, -1)
						s.metrics.DecActiveConnections()
					}()
					s.handleConnection(conn)
				}()
			default:
				// No available workers, reject connection
				s.logger.Warning("No available workers", map[string]interface{}{
					"remote_addr": remoteAddr,
				})
				conn.Close()
				s.metrics.IncRejectedConnections()
			}
		}
	}
}

// handleConnection handles a client connection
func (s *Server) handleConnection(conn net.Conn) {
	defer func() {
		conn.Close()
		atomic.AddInt32(&s.connCount, -1)
		s.metrics.DecActiveConnections()
	}()

	// Set read deadline
	conn.SetReadDeadline(time.Now().Add(*timeout))

	// Read message
	msg, err := proto.ReadMessage(conn)
	if err != nil {
		s.logger.Error("Failed to read message", err, map[string]interface{}{
			"remote_addr": conn.RemoteAddr(),
		})
		return
	}

	// Handle message based on type
	switch msg.Type {
	case proto.MessageTypeRegister:
		// Check authentication token
		if *config.AuthToken != "" && msg.Token != *config.AuthToken {
			s.logger.Warning("Invalid authentication token", map[string]interface{}{
				"remote_addr": conn.RemoteAddr(),
				"token":       msg.Token,
			})
			proto.WriteError(conn, "invalid authentication token")
			return
		}

		// Validate token
		if s.tokenValidator != nil {
			if !s.tokenValidator.Validate(msg.Token) {
				s.logger.Warning("Invalid token", map[string]interface{}{
					"remote_addr": conn.RemoteAddr(),
					"token":       msg.Token,
				})
				proto.WriteError(conn, "invalid token")
				return
			}
		}

		// Create tunnel
		tunnel := &Tunnel{
			conn:     conn,
			token:    msg.Token,
			lastPing: time.Now(),
			limiter:  ratelimit.NewLimiter(float64(*rateLimit), *rateBurst),
		}

		// Add tunnel
		s.mu.Lock()
		if len(s.tunnels) >= *maxTunnels {
			s.mu.Unlock()
			s.logger.Warning("Maximum number of tunnels reached", nil)
			proto.WriteError(conn, "maximum number of tunnels reached")
			return
		}
		s.tunnels[msg.Token] = tunnel
		s.mu.Unlock()

		s.logger.Info("Tunnel registered", map[string]interface{}{
			"remote_addr": conn.RemoteAddr(),
			"token":       msg.Token,
		})

		// Send success response
		proto.WriteMessage(conn, &proto.Message{
			Type: proto.MessageTypeConfig,
			Config: &proto.ConfigMessage{
				Token: msg.Token,
			},
		})

		// Handle tunnel
		s.handleTunnel(tunnel)

	case proto.MessageTypeConnect:
		// Find tunnel
		s.mu.RLock()
		tunnel, ok := s.tunnels[msg.Token]
		s.mu.RUnlock()

		if !ok {
			s.logger.Warning("Tunnel not found", map[string]interface{}{
				"remote_addr": conn.RemoteAddr(),
				"token":       msg.Token,
			})
			proto.WriteError(conn, "tunnel not found")
			return
		}

		// Check rate limit
		if !tunnel.limiter.Allow() {
			s.logger.Warning("Rate limit exceeded", map[string]interface{}{
				"remote_addr": conn.RemoteAddr(),
				"token":       msg.Token,
			})
			proto.WriteError(conn, "rate limit exceeded")
			return
		}

		// Handle connection
		s.handleConnection(conn)

	case proto.MessageTypePing:
		// Find tunnel
		s.mu.RLock()
		tunnel, ok := s.tunnels[msg.Token]
		s.mu.RUnlock()

		if !ok {
			s.logger.Warning("Tunnel not found", map[string]interface{}{
				"remote_addr": conn.RemoteAddr(),
				"token":       msg.Token,
			})
			proto.WriteError(conn, "tunnel not found")
			return
		}

		// Update last ping
		tunnel.lastPing = time.Now()

		// Send pong
		proto.WriteMessage(conn, &proto.Message{
			Type: proto.MessageTypePong,
			Pong: &proto.PongMessage{
				Token: msg.Token,
			},
		})

	default:
		s.logger.Warning("Unknown message type", map[string]interface{}{
			"remote_addr": conn.RemoteAddr(),
			"type":        msg.Type,
		})
		proto.WriteError(conn, "unknown message type")
	}
}

// handleTunnel handles a tunnel connection
func (s *Server) handleTunnel(tunnel *Tunnel) {
	for {
		// Set read deadline
		tunnel.conn.SetReadDeadline(time.Now().Add(*timeout))

		// Read message
		msg, err := proto.ReadMessage(tunnel.conn)
		if err != nil {
			s.logger.Error("Failed to read message", err, map[string]interface{}{
				"token": tunnel.token,
			})
			break
		}

		// Handle message based on type
		switch msg.Type {
		case proto.MessageTypePing:
			// Update last ping
			tunnel.lastPing = time.Now()

			// Send pong
			proto.WriteMessage(tunnel.conn, &proto.Message{
				Type: proto.MessageTypePong,
				Pong: &proto.PongMessage{
					Token: tunnel.token,
				},
			})

		default:
			s.logger.Warning("Unknown message type", map[string]interface{}{
				"token": tunnel.token,
				"type":  msg.Type,
			})
			proto.WriteError(tunnel.conn, "unknown message type")
		}
	}

	// Remove tunnel
	s.mu.Lock()
	delete(s.tunnels, tunnel.token)
	s.mu.Unlock()

	s.logger.Info("Tunnel closed", map[string]interface{}{
		"token": tunnel.token,
	})
}

// cleanupLoop periodically cleans up inactive tunnels
func (s *Server) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for token, tunnel := range s.tunnels {
			if now.Sub(tunnel.lastPing) > *timeout {
				s.logger.Warning("Removing inactive tunnel", map[string]interface{}{
					"token": token,
					"age":   now.Sub(tunnel.lastPing).String(),
				})
				tunnel.conn.Close()
				delete(s.tunnels, token)
			}
		}
		s.mu.Unlock()
	}
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.cancel() // Cancel the context to stop accepting new connections
	
	// Close the listener
	if s.listener != nil {
		s.listener.Close()
	}

	// Wait for all connections to close
	done := make(chan struct{})
	go func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		
		// Close all active connections
		for _, conn := range s.connections {
			conn.conn.Close()
		}
		for _, tunnel := range s.tunnels {
			tunnel.conn.Close()
		}
		
		close(done)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}

func main() {
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create server
	server, err := NewServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Start server
	if err := server.Start(cfg); err != nil {
		log.Fatalf("Server error: %v", err)
	}

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	// Shutdown server
	server.Shutdown(context.Background())
} 