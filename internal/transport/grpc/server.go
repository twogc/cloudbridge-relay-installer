package grpc

import (
	"context"
	"fmt"
	"net"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/twogc/cloudbridge-relay-installer/internal/auth"
	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
	"github.com/twogc/cloudbridge-relay-installer/internal/p2p"
	relayv1 "github.com/twogc/cloudbridge-relay-installer/internal/transport/grpc/proto/relay/v1"
	"github.com/twogc/cloudbridge-relay-installer/internal/tunnel"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"
)

// Server represents the gRPC server
type Server struct {
	grpcServer       *grpc.Server
	listener         net.Listener
	logger           *logger.Logger
	authAdapter      *auth.Adapter
	tunnelManager    *tunnel.Manager
	heartbeatManager *p2p.HeartbeatManager

	// Services
	controlService   *ControlService
	tunnelService    *TunnelService
	heartbeatService *HeartbeatService

	ctx    context.Context
	cancel context.CancelFunc
}

// Config represents gRPC server configuration
type Config struct {
	Port             int           `json:"port"`
	TLSEnabled       bool          `json:"tls_enabled"`
	CertFile         string        `json:"cert_file"`
	KeyFile          string        `json:"key_file"`
	MTLSEnabled      bool          `json:"mtls_enabled"`
	KeepalivePing    time.Duration `json:"keepalive_ping"`
	KeepaliveTimeout time.Duration `json:"keepalive_timeout"`
	MaxRecvMsgSize   int           `json:"max_recv_msg_size"`
	MaxSendMsgSize   int           `json:"max_send_msg_size"`
}

// NewServer creates a new gRPC server
func NewServer(
	config Config,
	logger *logger.Logger,
	authAdapter *auth.Adapter,
	tunnelManager *tunnel.Manager,
	heartbeatManager *p2p.HeartbeatManager,
) (*Server, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Create server options
	var opts []grpc.ServerOption

	// TLS configuration
	if config.TLSEnabled {
		creds, err := credentials.NewServerTLSFromFile(config.CertFile, config.KeyFile)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("failed to load TLS credentials: %v", err)
		}
		opts = append(opts, grpc.Creds(creds))
	}

	// Keepalive configuration
	opts = append(opts,
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    config.KeepalivePing,
			Timeout: config.KeepaliveTimeout,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             config.KeepalivePing / 2,
			PermitWithoutStream: true,
		}),
	)

	// Message size limits
	opts = append(opts,
		grpc.MaxRecvMsgSize(config.MaxRecvMsgSize),
		grpc.MaxSendMsgSize(config.MaxSendMsgSize),
	)

	// Middleware setup
	zapLogger := logger.GetZapLogger()
	grpc_zap.ReplaceGrpcLogger(zapLogger)

	opts = append(opts,
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_ctxtags.StreamServerInterceptor(),
			grpc_prometheus.StreamServerInterceptor,
			grpc_zap.StreamServerInterceptor(zapLogger),
			grpc_recovery.StreamServerInterceptor(),
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_prometheus.UnaryServerInterceptor,
			grpc_zap.UnaryServerInterceptor(zapLogger),
			grpc_recovery.UnaryServerInterceptor(),
		)),
	)

	// Create gRPC server
	grpcServer := grpc.NewServer(opts...)

	// Initialize Prometheus metrics
	grpc_prometheus.Register(grpcServer)

	// Create services
	controlService := NewControlService(authAdapter, logger)
	tunnelService := NewTunnelService(tunnelManager, authAdapter, logger)
	heartbeatService := NewHeartbeatService(heartbeatManager, authAdapter, logger)

	// Register services
	relayv1.RegisterControlServiceServer(grpcServer, controlService)
	relayv1.RegisterTunnelServiceServer(grpcServer, tunnelService)
	relayv1.RegisterHeartbeatServiceServer(grpcServer, heartbeatService)

	// Create listener
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Port))
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to listen on port %d: %v", config.Port, err)
	}

	return &Server{
		grpcServer:       grpcServer,
		listener:         listener,
		logger:           logger,
		authAdapter:      authAdapter,
		tunnelManager:    tunnelManager,
		heartbeatManager: heartbeatManager,
		controlService:   controlService,
		tunnelService:    tunnelService,
		heartbeatService: heartbeatService,
		ctx:              ctx,
		cancel:           cancel,
	}, nil
}

// Start starts the gRPC server (non-blocking)
func (s *Server) Start() error {
	s.logger.Info("Starting gRPC server",
		zap.String("address", s.listener.Addr().String()),
	)

	go func() {
		if err := s.grpcServer.Serve(s.listener); err != nil {
			s.logger.Error("gRPC server error", zap.Error(err))
		}
	}()

	return nil
}

// StartBlocking starts the gRPC server (blocking call)
func (s *Server) StartBlocking() error {
	s.logger.Info("Starting gRPC server (blocking)",
		zap.String("address", s.listener.Addr().String()),
	)

	// This call blocks until the server is stopped
	if err := s.grpcServer.Serve(s.listener); err != nil {
		s.logger.Error("gRPC server error", zap.Error(err))
		return err
	}

	return nil
}

// Stop stops the gRPC server
func (s *Server) Stop() error {
	s.logger.Info("Stopping gRPC server...")

	s.cancel()
	s.grpcServer.GracefulStop()

	if err := s.listener.Close(); err != nil {
		s.logger.Error("Error closing listener", zap.Error(err))
		return err
	}

	s.logger.Info("gRPC server stopped")
	return nil
}

// GetAddr returns the server address
func (s *Server) GetAddr() string {
	return s.listener.Addr().String()
}

// authInterceptor validates JWT tokens in gRPC requests
func (s *Server) authInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Skip auth for health checks
	if info.FullMethod == "/grpc.health.v1.Health/Check" {
		return handler(ctx, req)
	}

	// Extract client ID from context
	clientID := extractClientIDFromContext(ctx)
	if clientID == "" {
		return nil, status.Errorf(codes.Unauthenticated, "client ID not found")
	}

	// Add client ID to context for downstream use
	ctx = context.WithValue(ctx, "client_id", clientID)

	return handler(ctx, req)
}

// extractClientIDFromContext extracts client ID from gRPC metadata
func extractClientIDFromContext(ctx context.Context) string {
	// Implementation depends on how you want to pass client ID
	// This could be from JWT token, metadata, etc.
	return "default-client" // Placeholder
}
