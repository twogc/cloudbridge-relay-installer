package main

import (
	"flag"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/maxlanies/bore/internal/proto"
	"github.com/maxlanies/bore/internal/token"
)

var (
	addr       = flag.String("addr", ":12345", "Address to listen on")
	timeout    = flag.Duration("timeout", 10*time.Minute, "Connection timeout")
	maxTunnels = flag.Int("max-tunnels", 1000, "Maximum number of concurrent tunnels")

	// Token validation settings
	restEnabled    = flag.Bool("rest-enabled", false, "Enable REST token validation")
	restEndpoint   = flag.String("rest-endpoint", "", "REST API endpoint for token validation")
	restTimeout    = flag.Duration("rest-timeout", 5*time.Second, "REST API timeout")
	restCacheTTL   = flag.Duration("rest-cache-ttl", 1*time.Minute, "Token cache TTL")
	restMaxRetries = flag.Int("rest-max-retries", 3, "Maximum REST API retries")
	restRetryDelay = flag.Duration("rest-retry-delay", 100*time.Millisecond, "REST API retry delay")

	jwtEnabled    = flag.Bool("jwt-enabled", false, "Enable JWT token validation")
	jwtSigningKey = flag.String("jwt-key", "", "JWT signing key")
	jwtIssuer     = flag.String("jwt-issuer", "", "JWT issuer")
	jwtMaxAge     = flag.Duration("jwt-max-age", 24*time.Hour, "JWT maximum age")
)

// Error codes for better client handling
const (
	ErrEmptyToken     = "EMPTY_TOKEN"
	ErrTunnelNotFound = "TUNNEL_NOT_FOUND"
	ErrAlreadyExists  = "ALREADY_EXISTS"
	ErrInvalidMessage = "INVALID_MESSAGE"
	ErrTooManyTunnels = "TOO_MANY_TUNNELS"
)

type Tunnel struct {
	conn     net.Conn
	created  time.Time
	lastPing time.Time
	mu       sync.Mutex
}

type Server struct {
	listener  net.Listener
	tunnels   map[string]*Tunnel
	mu        sync.RWMutex
	validator *token.Validator
	logger    *log.Logger
}

func NewServer() (*Server, error) {
	validatorConfig := token.Config{
		RESTEnabled:     *restEnabled,
		RESTEndpoint:    *restEndpoint,
		RESTTimeout:     *restTimeout,
		RESTCacheTTL:    *restCacheTTL,
		RESTMaxRetries:  *restMaxRetries,
		RESTRetryDelay:  *restRetryDelay,
		JWTEnabled:      *jwtEnabled,
		JWTSigningKey:   []byte(*jwtSigningKey),
		JWTIssuer:       *jwtIssuer,
		JWTMaxAge:       *jwtMaxAge,
	}

	validator := token.NewValidator(validatorConfig)

	return &Server{
		tunnels:   make(map[string]*Tunnel),
		validator: validator,
		logger:    log.New(os.Stdout, "[relay] ", log.LstdFlags),
	}, nil
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", *addr)
	if err != nil {
		return err
	}
	s.listener = listener

	go s.cleanupLoop()

	s.logger.Printf("Server started on %s", *addr)
	return s.acceptLoop()
}

func (s *Server) acceptLoop() error {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return err
		}
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	conn.SetReadDeadline(time.Now().Add(*timeout))

	// 1. Прочитать первый message (handshake)
	msgAny, err := proto.ReadMessage(conn)
	if err != nil {
		s.logger.Printf("Failed to read message: %v", err)
		return
	}

	switch msg := msgAny.(type) {
	case *proto.RegisterMessage:
		if msg.Token == "" {
			proto.WriteError(conn, ErrEmptyToken)
			return
		}
		// Validate token
		if !s.validator.Validate(msg.Token) {
			proto.WriteError(conn, "INVALID_TOKEN")
			return
		}
		// Check and register tunnel
		s.mu.Lock()
		if len(s.tunnels) >= *maxTunnels {
			s.mu.Unlock()
			proto.WriteError(conn, ErrTooManyTunnels)
			return
		}
		if _, exists := s.tunnels[msg.Token]; exists {
			s.mu.Unlock()
			proto.WriteError(conn, ErrAlreadyExists)
			return
		}
		tunnel := &Tunnel{
			conn:     conn,
			created:  time.Now(),
			lastPing: time.Now(),
		}
		s.tunnels[msg.Token] = tunnel
		s.mu.Unlock()
		s.logger.Printf("Tunnel registered: %s", msg.Token)
		// Send config back
		config := &proto.ConfigMessage{
			Token:     msg.Token,
			PublicURL: "bore://" + msg.Token,
			Status:    "ok",
		}
		if err := proto.WriteMessage(conn, config); err != nil {
			s.logger.Printf("Failed to send config: %v", err)
			return
		}
		// Держим туннель в keepalive
		s.handleTunnel(msg.Token, tunnel)
	case *proto.ConnectMessage:
		// Проверка токена
		if msg.Token == "" {
			proto.WriteError(conn, ErrEmptyToken)
			return
		}
		if !s.validator.Validate(msg.Token) {
			proto.WriteError(conn, "INVALID_TOKEN")
			return
		}
		s.mu.RLock()
		tunnel, exists := s.tunnels[msg.Token]
		s.mu.RUnlock()
		if !exists {
			proto.WriteError(conn, ErrTunnelNotFound)
			return
		}
		// Bi-directional relay
		s.logger.Printf("Consumer connected: %s", msg.Token)
		var once sync.Once
		closeBoth := func() { once.Do(func() { conn.Close(); tunnel.conn.Close() }) }
		go func() { defer closeBoth(); _, _ = io.Copy(tunnel.conn, conn) }()
		go func() { defer closeBoth(); _, _ = io.Copy(conn, tunnel.conn) }()
		// Wait (simple block)
		select {}
	case *proto.PingMessage:
		if msg.Token == "" {
			proto.WriteError(conn, ErrEmptyToken)
			return
		}
		if tunnel := s.getTunnel(msg.Token); tunnel != nil {
			tunnel.mu.Lock()
			tunnel.lastPing = time.Now()
			tunnel.mu.Unlock()
		}
		pong := &proto.PongMessage{Token: msg.Token}
		proto.WriteMessage(conn, pong)
	default:
		proto.WriteError(conn, ErrInvalidMessage)
	}
}

func (s *Server) handleTunnel(token string, tunnel *Tunnel) {
	// Keepalive loop
	for {
		tunnel.conn.SetReadDeadline(time.Now().Add(*timeout))
		msgAny, err := proto.ReadMessage(tunnel.conn)
		if err != nil {
			s.logger.Printf("Tunnel closed (%s): %v", token, err)
			s.mu.Lock()
			delete(s.tunnels, token)
			s.mu.Unlock()
			return
		}
		switch msg := msgAny.(type) {
		case *proto.PingMessage:
			tunnel.mu.Lock()
			tunnel.lastPing = time.Now()
			tunnel.mu.Unlock()
			pong := &proto.PongMessage{Token: msg.Token}
			proto.WriteMessage(tunnel.conn, pong)
		default:
			// ignore
		}
	}
}

func (s *Server) getTunnel(token string) *Tunnel {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.tunnels[token]
}

func (s *Server) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for token, tunnel := range s.tunnels {
			tunnel.mu.Lock()
			if now.Sub(tunnel.lastPing) > *timeout {
				tunnel.conn.Close()
				delete(s.tunnels, token)
				s.logger.Printf("Cleaned up tunnel: %s", token)
			}
			tunnel.mu.Unlock()
		}
		s.mu.Unlock()
		// Очистка токен-валидации (если есть cache)
		s.validator.Cleanup()
	}
}

func main() {
	flag.Parse()

	if *restEnabled && *restEndpoint == "" {
		log.Fatal("REST endpoint is required when REST validation is enabled")
	}
	if *jwtEnabled && *jwtSigningKey == "" {
		log.Fatal("JWT signing key is required when JWT validation is enabled")
	}

	server, err := NewServer()
	if err != nil {
		log.Fatal(err)
	}

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		server.logger.Println("Shutting down...")
		if server.listener != nil {
			server.listener.Close()
		}
	}()

	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
} 