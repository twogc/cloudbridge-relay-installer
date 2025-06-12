package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/maxlanies/cloudbridge-relay-installer/internal/metrics"
	"github.com/maxlanies/cloudbridge-relay-installer/internal/proto"
	"github.com/maxlanies/cloudbridge-relay-installer/internal/ratelimit"
	"github.com/maxlanies/cloudbridge-relay-installer/internal/token"
)

var (
	addr            = flag.String("addr", ":8080", "Server address")
	timeout         = flag.Duration("timeout", 30*time.Second, "Connection timeout")
	maxTunnels      = flag.Int("max-tunnels", 1000, "Maximum number of concurrent tunnels")
	restEndpoint    = flag.String("rest-endpoint", "", "REST API endpoint for token validation")
	jwtSecret       = flag.String("jwt-secret", "", "JWT secret for token validation")
	rateLimit       = flag.Float64("rate-limit", 100, "Rate limit per second")
	rateBurst       = flag.Int("rate-burst", 50, "Rate limit burst size")
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
}

// NewServer creates a new server
func NewServer(port int, tokenValidator *token.Validator, limiter *ratelimit.IPLimiter, maxConnections int) *Server {
	return &Server{
		port:           port,
		tokenValidator: tokenValidator,
		limiter:        limiter,
		maxConnections: maxConnections,
		metrics:        metrics.NewMetrics(),
		connections:    make(map[string]*Connection),
		tunnels:        make(map[string]*Tunnel),
		logger:         log.New(os.Stdout, "", log.LstdFlags),
	}
}

// Start starts the server
func (s *Server) Start() error {
	var err error
	s.listener, err = net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("failed to start server: %v", err)
	}
	defer s.listener.Close()

	log.Printf("Server started on port %d", s.port)

	// Start cleanup loop
	go s.cleanupLoop()

	// Start accepting connections
	for {
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

		atomic.AddInt32(&s.connCount, 1)
		s.metrics.IncActiveConnections()
		go s.handleConnection(conn)
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

	// Create server
	server := NewServer(8080, nil, nil, 1000)

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