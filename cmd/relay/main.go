package main

import (
	"context"
	"crypto/tls"
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
	"github.com/maxlanies/cloudbridge-relay-installer/internal/metrics"
	"github.com/maxlanies/cloudbridge-relay-installer/internal/proto"
	"github.com/maxlanies/cloudbridge-relay-installer/internal/ratelimit"
	"github.com/maxlanies/cloudbridge-relay-installer/internal/token"
)

var (
	addr            = flag.String("addr", ":3456", "Server address")
	timeout         = flag.Duration("timeout", 30*time.Second, "Connection timeout")
	maxTunnels      = flag.Int("max-tunnels", 10000, "Maximum number of concurrent tunnels")
	restEndpoint    = flag.String("rest-endpoint", "", "REST API endpoint for token validation")
	jwtSecret       = flag.String("jwt-secret", "", "JWT secret for token validation")
	rateLimit       = flag.Float64("rate-limit", 1000, "Rate limit per second")
	rateBurst       = flag.Int("rate-burst", 500, "Rate limit burst size")
	workerPoolSize  = flag.Int("workers", runtime.NumCPU()*2, "Number of worker goroutines")
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
	port            int
	tokenValidator  *token.Validator
	limiter         *ratelimit.IPLimiter
	maxConnections  int
	connCount       int32
	metrics         *metrics.Metrics
	mu              sync.RWMutex
	connections     map[string]*Connection
	listener        net.Listener
	tunnels         map[string]*Tunnel
	logger          *log.Logger
	workerPool      chan struct{}
	ctx             context.Context
	cancel          context.CancelFunc
}

// NewServer creates a new server
func NewServer(port int, tokenValidator *token.Validator, limiter *ratelimit.IPLimiter, maxConnections int) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		port:           port,
		tokenValidator: tokenValidator,
		limiter:        limiter,
		maxConnections: maxConnections,
		metrics:        metrics.NewMetrics(),
		connections:    make(map[string]*Connection),
		tunnels:        make(map[string]*Tunnel),
		logger:         log.New(os.Stdout, "", log.LstdFlags),
		workerPool:     make(chan struct{}, *workerPoolSize),
		ctx:            ctx,
		cancel:         cancel,
	}
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
func (s *Server) Start() error {
	var err error
	var listener net.Listener

	if *config.UseTLS {
		// Load TLS certificate
		cert, err := tls.LoadX509KeyPair(*config.CertFile, *config.KeyFile)
		if err != nil {
			return fmt.Errorf("failed to load TLS certificate: %v", err)
		}

		// Configure TLS
		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS12,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			},
		}

		// Create TLS listener
		listener, err = tls.Listen("tcp", fmt.Sprintf(":%d", s.port), tlsConfig)
		if err != nil {
			return fmt.Errorf("failed to start TLS server: %v", err)
		}
	} else {
		// Create regular TCP listener
		listener, err = net.Listen("tcp", fmt.Sprintf(":%d", s.port))
		if err != nil {
			return fmt.Errorf("failed to start server: %v", err)
		}
	}

	s.listener = listener
	defer s.listener.Close()

	log.Printf("Server started on port %d (TLS: %v, Workers: %d)", s.port, *config.UseTLS, *workerPoolSize)

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
				log.Printf("Failed to accept connection: %v", err)
				continue
			}

			// Check connection limit
			if atomic.LoadInt32(&s.connCount) >= int32(s.maxConnections) {
				log.Printf("Connection limit reached (%d), rejecting new connection", s.maxConnections)
				conn.Close()
				s.metrics.IncRejectedConnections()
				continue
			}

			// Check rate limit
			remoteAddr := conn.RemoteAddr().String()
			if !s.limiter.Allow(remoteAddr) {
				log.Printf("Rate limit exceeded for %s", remoteAddr)
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
				log.Printf("No available workers, rejecting connection from %s", remoteAddr)
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
		s.logger.Printf("Failed to read message: %v", err)
		return
	}

	// Handle message based on type
	switch msg.Type {
	case proto.MessageTypeRegister:
		// Check authentication token
		if *config.AuthToken != "" && msg.Token != *config.AuthToken {
			s.logger.Printf("Invalid authentication token from %s", conn.RemoteAddr())
			proto.WriteError(conn, "invalid authentication token")
			return
		}

		// Validate token
		if s.tokenValidator != nil {
			if !s.tokenValidator.Validate(msg.Token) {
				s.logger.Printf("Invalid token: %s", msg.Token)
				proto.WriteError(conn, "invalid token")
				return
			}
		}

		// Create tunnel
		tunnel := &Tunnel{
			conn:     conn,
			token:    msg.Token,
			lastPing: time.Now(),
			limiter:  ratelimit.NewLimiter(*rateLimit, *rateBurst),
		}

		// Add tunnel
		s.mu.Lock()
		if len(s.tunnels) >= *maxTunnels {
			s.mu.Unlock()
			s.logger.Printf("Maximum number of tunnels reached")
			proto.WriteError(conn, "maximum number of tunnels reached")
			return
		}
		s.tunnels[msg.Token] = tunnel
		s.mu.Unlock()

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
			s.logger.Printf("Tunnel not found: %s", msg.Token)
			proto.WriteError(conn, "tunnel not found")
			return
		}

		// Check rate limit
		if !tunnel.limiter.Allow() {
			s.logger.Printf("Rate limit exceeded for tunnel %s", msg.Token)
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
			s.logger.Printf("Tunnel not found: %s", msg.Token)
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
		s.logger.Printf("Unknown message type: %d", msg.Type)
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
			s.logger.Printf("Failed to read message: %v", err)
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
			s.logger.Printf("Unknown message type: %d", msg.Type)
			proto.WriteError(tunnel.conn, "unknown message type")
		}
	}

	// Remove tunnel
	s.mu.Lock()
	delete(s.tunnels, tunnel.token)
	s.mu.Unlock()
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
				s.logger.Printf("Removing inactive tunnel: %s", token)
				tunnel.conn.Close()
				delete(s.tunnels, token)
			}
		}
		s.mu.Unlock()
	}
}

func main() {
	flag.Parse()

	// Create rate limiter
	limiter := ratelimit.NewIPLimiter(*rateLimit, *rateBurst)

	// Create server
	server := NewServer(3456, nil, limiter, 1000)

	// Handle signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Start server
	go func() {
		if err := server.Start(); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for signal
	<-sigCh
	log.Println("Shutting down...")
} 