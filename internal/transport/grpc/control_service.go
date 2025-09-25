package grpc

import (
	"context"
	"fmt"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/internal/auth"
	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
	relayv1 "github.com/twogc/cloudbridge-relay-installer/internal/transport/grpc/proto/relay/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ControlService implements the ControlService gRPC service
type ControlService struct {
	relayv1.UnimplementedControlServiceServer
	authAdapter *auth.Adapter
	logger      *logger.Logger
}

// NewControlService creates a new ControlService
func NewControlService(authAdapter *auth.Adapter, logger *logger.Logger) *ControlService {
	return &ControlService{
		authAdapter: authAdapter,
		logger:      logger,
	}
}

// Hello implements the Hello RPC method
func (s *ControlService) Hello(ctx context.Context, req *relayv1.HelloRequest) (*relayv1.HelloResponse, error) {
	s.logger.Info("Received Hello request",
		zap.String("client_id", req.ClientId),
		zap.String("version", req.Version),
		zap.Strings("features", req.Features),
	)

	// Create response with server information
	resp := &relayv1.HelloResponse{
		Status:        "ok",
		ServerVersion: "1.0.0",
		SupportedFeatures: []string{
			"tls",
			"mtls",
			"heartbeat",
			"tunneling",
			"streaming",
		},
		SessionId: generateSessionID(),
		Timestamp: timestamppb.New(time.Now()),
	}

	s.logger.Info("Sent Hello response",
		zap.String("session_id", resp.SessionId),
		zap.Strings("supported_features", resp.SupportedFeatures),
	)

	return resp, nil
}

// Authenticate implements the Authenticate RPC method
func (s *ControlService) Authenticate(ctx context.Context, req *relayv1.AuthRequest) (*relayv1.AuthResponse, error) {
	s.logger.Info("Received Auth request",
		zap.String("client_id", req.ClientId),
		zap.String("auth_type", req.AuthType),
	)

	if req.Token == "" {
		s.logger.Warn("Authentication failed: empty token")
		return &relayv1.AuthResponse{
			Status:       "error",
			ErrorMessage: "Token required",
		}, status.Error(codes.Unauthenticated, "token required")
	}

	// Validate token using auth adapter
	tokenInfo, err := s.authAdapter.ValidateToken(ctx, req.Token)
	if err != nil {
		s.logger.Warn("Authentication failed: invalid token",
			zap.Error(err),
			zap.String("client_id", req.ClientId),
		)
		return &relayv1.AuthResponse{
			Status:       "error",
			ErrorMessage: "Invalid token: " + err.Error(),
		}, status.Error(codes.Unauthenticated, "invalid token")
	}

	s.logger.Info("Authentication successful",
		zap.String("client_id", req.ClientId),
		zap.String("subject", tokenInfo.Subject),
		zap.String("username", tokenInfo.Username),
		zap.String("provider", tokenInfo.Provider),
	)

	// Extract tenant ID from token info or use default
	tenantID := "default"
	if len(tokenInfo.Groups) > 0 {
		tenantID = tokenInfo.Groups[0] // Use first group as tenant ID
	}

	// Create response
	resp := &relayv1.AuthResponse{
		Status:       "ok",
		ClientId:     tokenInfo.Subject,
		TenantId:     tenantID,
		SessionToken: generateSessionToken(),
		ExpiresAt:    timestamppb.New(tokenInfo.ExpiresAt),
	}

	return resp, nil
}

// GetStatus implements the GetStatus RPC method
func (s *ControlService) GetStatus(ctx context.Context, req *relayv1.StatusRequest) (*relayv1.StatusResponse, error) {
	s.logger.Debug("Received Status request",
		zap.String("client_id", req.ClientId),
	)

	// Create status response
	resp := &relayv1.StatusResponse{
		Status:        "connected",
		Connected:     true,
		TransportMode: "grpc",
		LastHeartbeat: timestamppb.New(time.Now()),
		BytesSent:     0, // Could be tracked if needed
		BytesReceived: 0, // Could be tracked if needed
	}

	return resp, nil
}

// generateSessionID generates a unique session ID
func generateSessionID() string {
	return fmt.Sprintf("session_%d", time.Now().Unix())
}

// generateSessionToken generates a session token
func generateSessionToken() string {
	return fmt.Sprintf("token_%d", time.Now().UnixNano())
}
