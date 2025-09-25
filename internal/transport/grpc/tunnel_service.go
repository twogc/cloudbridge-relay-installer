package grpc

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/internal/auth"
	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
	relayv1 "github.com/twogc/cloudbridge-relay-installer/internal/transport/grpc/proto/relay/v1"
	"github.com/twogc/cloudbridge-relay-installer/internal/tunnel"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TunnelService implements the TunnelService gRPC service
type TunnelService struct {
	relayv1.UnimplementedTunnelServiceServer
	tunnelManager *tunnel.Manager
	authAdapter   *auth.Adapter
	logger        *logger.Logger
}

// NewTunnelService creates a new TunnelService
func NewTunnelService(tunnelManager *tunnel.Manager, authAdapter *auth.Adapter, logger *logger.Logger) *TunnelService {
	return &TunnelService{
		tunnelManager: tunnelManager,
		authAdapter:   authAdapter,
		logger:        logger,
	}
}

// CreateTunnel implements the CreateTunnel RPC method
func (s *TunnelService) CreateTunnel(ctx context.Context, req *relayv1.CreateTunnelRequest) (*relayv1.CreateTunnelResponse, error) {
	s.logger.Info("Received CreateTunnel request",
		zap.String("tunnel_id", req.TunnelId),
		zap.String("tenant_id", req.TenantId),
		zap.Int32("local_port", req.LocalPort),
		zap.String("remote_host", req.RemoteHost),
		zap.Int32("remote_port", req.RemotePort),
	)

	// Validate request
	if req.TunnelId == "" {
		return nil, status.Error(codes.InvalidArgument, "tunnel_id is required")
	}
	if req.RemoteHost == "" {
		return nil, status.Error(codes.InvalidArgument, "remote_host is required")
	}
	if req.RemotePort <= 0 {
		return nil, status.Error(codes.InvalidArgument, "remote_port must be positive")
	}

	// For gRPC, we need to create a virtual connection
	// This is different from HTTP where we have a real TCP connection
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close() // We'll return the server end

	// Create tunnel using existing tunnel manager
	proxy, err := s.tunnelManager.CreateTunnel(
		req.TunnelId,
		serverConn,
		req.RemoteHost,
		int(req.RemotePort),
	)
	if err != nil {
		s.logger.Error("Failed to create tunnel",
			zap.String("tunnel_id", req.TunnelId),
			zap.Error(err),
		)
		return &relayv1.CreateTunnelResponse{
			Status:       "error",
			ErrorMessage: fmt.Sprintf("Failed to create tunnel: %v", err),
		}, status.Error(codes.Internal, err.Error())
	}

	// Create tunnel info
	tunnelInfo := &relayv1.TunnelInfo{
		TunnelId:     req.TunnelId,
		TenantId:     req.TenantId,
		LocalPort:    req.LocalPort,
		RemoteHost:   req.RemoteHost,
		RemotePort:   req.RemotePort,
		Status:       "active",
		CreatedAt:    timestamppb.New(time.Now()),
		LastActivity: timestamppb.New(time.Now()),
		Config:       req.Config,
	}

	resp := &relayv1.CreateTunnelResponse{
		Status:     "ok",
		TunnelId:   req.TunnelId,
		Endpoint:   fmt.Sprintf("%s:%d", req.RemoteHost, req.RemotePort),
		TunnelInfo: tunnelInfo,
	}

	s.logger.Info("Tunnel created successfully",
		zap.String("tunnel_id", req.TunnelId),
		zap.String("endpoint", resp.Endpoint),
	)

	// Start tracking the proxy if needed
	_ = proxy // proxy can be used for additional operations

	return resp, nil
}

// CloseTunnel implements the CloseTunnel RPC method
func (s *TunnelService) CloseTunnel(ctx context.Context, req *relayv1.CloseTunnelRequest) (*relayv1.CloseTunnelResponse, error) {
	s.logger.Info("Received CloseTunnel request",
		zap.String("tunnel_id", req.TunnelId),
		zap.String("tenant_id", req.TenantId),
		zap.Bool("force", req.Force),
	)

	if req.TunnelId == "" {
		return nil, status.Error(codes.InvalidArgument, "tunnel_id is required")
	}

	// Close tunnel using tunnel manager
	err := s.tunnelManager.CloseTunnel(req.TunnelId)
	if err != nil {
		s.logger.Error("Failed to close tunnel",
			zap.String("tunnel_id", req.TunnelId),
			zap.Error(err),
		)
		return &relayv1.CloseTunnelResponse{
			Status:       "error",
			TunnelId:     req.TunnelId,
			ErrorMessage: fmt.Sprintf("Failed to close tunnel: %v", err),
		}, status.Error(codes.Internal, err.Error())
	}

	s.logger.Info("Tunnel closed successfully",
		zap.String("tunnel_id", req.TunnelId),
	)

	return &relayv1.CloseTunnelResponse{
		Status:   "ok",
		TunnelId: req.TunnelId,
	}, nil
}

// ListTunnels implements the ListTunnels RPC method
func (s *TunnelService) ListTunnels(ctx context.Context, req *relayv1.ListTunnelsRequest) (*relayv1.ListTunnelsResponse, error) {
	s.logger.Debug("Received ListTunnels request",
		zap.String("tenant_id", req.TenantId),
		zap.String("filter", req.Filter),
	)

	// Get tunnel stats from manager
	stats := s.tunnelManager.GetTunnelStats()

	var tunnels []*relayv1.TunnelInfo
	for tunnelID, stat := range stats {
		// Filter by tenant if specified
		if req.TenantId != "" && req.TenantId != "default" {
			// For now, we don't have tenant info in stats, so skip filtering
		}

		// Convert tunnel stats to TunnelInfo
		tunnelInfo := &relayv1.TunnelInfo{
			TunnelId:     tunnelID,
			TenantId:     "default", // Would need to track this
			LocalPort:    0,         // Would need to track this
			RemoteHost:   "",        // Would need to track this
			RemotePort:   0,         // Would need to track this
			Status:       "active",
			CreatedAt:    timestamppb.New(stat.StartTime),
			LastActivity: timestamppb.New(stat.LastActivity),
		}

		tunnels = append(tunnels, tunnelInfo)
	}

	return &relayv1.ListTunnelsResponse{
		Tunnels:    tunnels,
		TotalCount: int32(len(tunnels)),
	}, nil
}

// GetTunnelStatus implements the GetTunnelStatus RPC method
func (s *TunnelService) GetTunnelStatus(ctx context.Context, req *relayv1.TunnelStatusRequest) (*relayv1.TunnelStatusResponse, error) {
	s.logger.Debug("Received GetTunnelStatus request",
		zap.String("tunnel_id", req.TunnelId),
		zap.String("tenant_id", req.TenantId),
	)

	if req.TunnelId == "" {
		return nil, status.Error(codes.InvalidArgument, "tunnel_id is required")
	}

	// Get tunnel from manager
	tunnelProxy, exists := s.tunnelManager.GetTunnel(req.TunnelId)
	if !exists {
		return nil, status.Error(codes.NotFound, "tunnel not found")
	}

	// Get tunnel stats
	stats := tunnelProxy.GetStats()

	tunnelInfo := &relayv1.TunnelInfo{
		TunnelId:     req.TunnelId,
		TenantId:     req.TenantId,
		Status:       "active",
		CreatedAt:    timestamppb.New(stats.StartTime),
		LastActivity: timestamppb.New(stats.LastActivity),
	}

	tunnelStats := &relayv1.TunnelStats{
		BytesSent:       stats.BytesToClient,
		BytesReceived:   stats.BytesFromClient,
		PacketsSent:     0, // Not tracked in current implementation
		PacketsReceived: 0, // Not tracked in current implementation
		Errors:          0, // Not tracked in current ProxyStats
		LastReset:       timestamppb.New(stats.StartTime),
	}

	return &relayv1.TunnelStatusResponse{
		Status:     "ok",
		TunnelInfo: tunnelInfo,
		Stats:      tunnelStats,
	}, nil
}

// StreamData implements the StreamData RPC method for bidirectional data streaming
func (s *TunnelService) StreamData(stream relayv1.TunnelService_StreamDataServer) error {
	s.logger.Info("Started StreamData session")

	ctx := stream.Context()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("StreamData session ended", zap.Error(ctx.Err()))
			return ctx.Err()
		default:
			// Receive data packet from client
			packet, err := stream.Recv()
			if err == io.EOF {
				s.logger.Info("StreamData client closed connection")
				return nil
			}
			if err != nil {
				s.logger.Error("Failed to receive data packet", zap.Error(err))
				return err
			}

			s.logger.Debug("Received data packet",
				zap.String("tunnel_id", packet.TunnelId),
				zap.Int("data_size", len(packet.Data)),
				zap.Int64("sequence", packet.SequenceNumber),
				zap.String("type", packet.PacketType.String()),
			)

			// Echo the packet back (for testing)
			// In a real implementation, this would forward data to the target
			responsePacket := &relayv1.DataPacket{
				TunnelId:       packet.TunnelId,
				Data:           packet.Data,
				SequenceNumber: packet.SequenceNumber,
				Timestamp:      timestamppb.New(time.Now()),
				PacketType:     relayv1.PacketType_DATA,
			}

			if err := stream.Send(responsePacket); err != nil {
				s.logger.Error("Failed to send data packet", zap.Error(err))
				return err
			}

			s.logger.Debug("Sent data packet response",
				zap.String("tunnel_id", responsePacket.TunnelId),
				zap.Int("data_size", len(responsePacket.Data)),
			)
		}
	}
}
