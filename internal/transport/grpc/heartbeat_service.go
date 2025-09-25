package grpc

import (
	"context"
	"io"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/internal/auth"
	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
	"github.com/twogc/cloudbridge-relay-installer/internal/p2p"
	relayv1 "github.com/twogc/cloudbridge-relay-installer/internal/transport/grpc/proto/relay/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// HeartbeatService implements the HeartbeatService gRPC service
type HeartbeatService struct {
	relayv1.UnimplementedHeartbeatServiceServer
	heartbeatManager *p2p.HeartbeatManager
	authAdapter      *auth.Adapter
	logger           *logger.Logger
}

// NewHeartbeatService creates a new HeartbeatService
func NewHeartbeatService(heartbeatManager *p2p.HeartbeatManager, authAdapter *auth.Adapter, logger *logger.Logger) *HeartbeatService {
	return &HeartbeatService{
		heartbeatManager: heartbeatManager,
		authAdapter:      authAdapter,
		logger:           logger,
	}
}

// SendHeartbeat implements the SendHeartbeat RPC method
func (s *HeartbeatService) SendHeartbeat(ctx context.Context, req *relayv1.HeartbeatRequest) (*relayv1.HeartbeatResponse, error) {
	s.logger.Debug("Received SendHeartbeat request",
		zap.String("client_id", req.ClientId),
		zap.String("tenant_id", req.TenantId),
		zap.String("transport_mode", req.TransportMode),
	)

	if req.ClientId == "" {
		return nil, status.Error(codes.InvalidArgument, "client_id is required")
	}

	if req.TenantId == "" {
		req.TenantId = "default"
	}

	// Process heartbeat using existing manager
	heartbeatResponse, err := s.heartbeatManager.ProcessHeartbeat(req.ClientId, req.TenantId)
	if err != nil {
		s.logger.Error("Failed to process heartbeat",
			zap.String("client_id", req.ClientId),
			zap.String("tenant_id", req.TenantId),
			zap.Error(err),
		)
		return &relayv1.HeartbeatResponse{
			Status:       "error",
			ErrorMessage: err.Error(),
		}, status.Error(codes.Internal, err.Error())
	}

	// Convert manager response to gRPC response
	resp := &relayv1.HeartbeatResponse{
		Status:          heartbeatResponse.Status,
		ServerTimestamp: timestamppb.New(time.Unix(heartbeatResponse.Timestamp, 0)),
		IntervalSeconds: 30, // Default interval
		ServerMetrics: &relayv1.ServerMetrics{
			ConnectedClients: 1,   // Would need to track this
			ActiveTunnels:    1,   // Would need to track this
			ServerLoad:       0.5, // Would need to calculate this
			Timestamp:        timestamppb.New(time.Now()),
		},
	}

	if heartbeatResponse.Online {
		resp.Status = "ok"
	}

	s.logger.Debug("Sent SendHeartbeat response",
		zap.String("client_id", req.ClientId),
		zap.String("status", resp.Status),
	)

	return resp, nil
}

// StreamHeartbeat implements the StreamHeartbeat RPC method for bidirectional heartbeat streaming
func (s *HeartbeatService) StreamHeartbeat(stream relayv1.HeartbeatService_StreamHeartbeatServer) error {
	s.logger.Info("Started StreamHeartbeat session")

	ctx := stream.Context()

	// Channel for heartbeat responses
	responsesChan := make(chan *relayv1.HeartbeatResponse, 10)
	defer close(responsesChan)

	// Start goroutine to send responses
	go func() {
		for resp := range responsesChan {
			if err := stream.Send(resp); err != nil {
				s.logger.Error("Failed to send heartbeat response", zap.Error(err))
				return
			}
		}
	}()

	// Start periodic server heartbeats (every 30 seconds)
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	serverHeartbeatChan := make(chan struct{}, 1)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				select {
				case serverHeartbeatChan <- struct{}{}:
				default:
				}
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("StreamHeartbeat session ended", zap.Error(ctx.Err()))
			return ctx.Err()

		case <-serverHeartbeatChan:
			// Send server-initiated heartbeat
			serverHeartbeat := &relayv1.HeartbeatResponse{
				Status:          "ping",
				ServerTimestamp: timestamppb.New(time.Now()),
				IntervalSeconds: 30,
				ServerMetrics: &relayv1.ServerMetrics{
					ConnectedClients: 1,
					ActiveTunnels:    1,
					ServerLoad:       0.5,
					Timestamp:        timestamppb.New(time.Now()),
				},
			}

			select {
			case responsesChan <- serverHeartbeat:
				s.logger.Debug("Sent server heartbeat")
			case <-ctx.Done():
				return ctx.Err()
			}

		default:
			// Receive heartbeat from client
			req, err := stream.Recv()
			if err == io.EOF {
				s.logger.Info("StreamHeartbeat client closed connection")
				return nil
			}
			if err != nil {
				s.logger.Error("Failed to receive heartbeat", zap.Error(err))
				return err
			}

			s.logger.Debug("Received client heartbeat",
				zap.String("client_id", req.ClientId),
				zap.String("tenant_id", req.TenantId),
				zap.String("transport_mode", req.TransportMode),
			)

			// Process heartbeat
			if req.TenantId == "" {
				req.TenantId = "default"
			}

			heartbeatResponse, err := s.heartbeatManager.ProcessHeartbeat(req.ClientId, req.TenantId)
			if err != nil {
				s.logger.Error("Failed to process streaming heartbeat",
					zap.String("client_id", req.ClientId),
					zap.Error(err),
				)
				continue
			}

			// Send response
			resp := &relayv1.HeartbeatResponse{
				Status:          heartbeatResponse.Status,
				ServerTimestamp: timestamppb.New(time.Unix(heartbeatResponse.Timestamp, 0)),
				IntervalSeconds: 30,
				ServerMetrics: &relayv1.ServerMetrics{
					ConnectedClients: 1,
					ActiveTunnels:    1,
					ServerLoad:       0.5,
					Timestamp:        timestamppb.New(time.Now()),
				},
			}

			if heartbeatResponse.Online {
				resp.Status = "pong"
			}

			select {
			case responsesChan <- resp:
				s.logger.Debug("Queued heartbeat response",
					zap.String("client_id", req.ClientId),
					zap.String("status", resp.Status),
				)
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
}

// GetHealth implements the GetHealth RPC method
func (s *HeartbeatService) GetHealth(ctx context.Context, req *relayv1.HealthRequest) (*relayv1.HealthResponse, error) {
	s.logger.Debug("Received GetHealth request",
		zap.String("client_id", req.ClientId),
		zap.Bool("include_metrics", req.IncludeMetrics),
	)

	resp := &relayv1.HealthResponse{
		Status:        "healthy",
		Healthy:       true,
		LastHeartbeat: timestamppb.New(time.Now()),
		UptimeSeconds: int32(time.Since(time.Now().Add(-time.Hour)).Seconds()), // Placeholder
	}

	if req.IncludeMetrics {
		resp.ClientMetrics = &relayv1.ClientMetrics{
			BytesSent:         0,
			BytesReceived:     0,
			PacketsSent:       0,
			PacketsReceived:   0,
			ActiveTunnels:     1,
			ActiveP2PSessions: 0,
			CpuUsage:          0.1,
			MemoryUsage:       0.2,
			TransportMode:     "grpc",
			LastSwitch:        timestamppb.New(time.Now()),
		}

		resp.ServerMetrics = &relayv1.ServerMetrics{
			ConnectedClients: 1,
			ActiveTunnels:    1,
			ServerLoad:       0.5,
			Timestamp:        timestamppb.New(time.Now()),
		}
	}

	return resp, nil
}
