package grpc

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"github.com/twogc/cloudbridge-relay-installer/internal/p2p"
)

// WireGuardService реализует gRPC сервис для управления WireGuard туннелями
type WireGuardService struct {
	relayManager *p2p.RelayManager
	logger       *zap.Logger
}

// NewWireGuardService создает новый WireGuard gRPC сервис
func NewWireGuardService(relayManager *p2p.RelayManager, logger *zap.Logger) *WireGuardService {
	return &WireGuardService{
		relayManager: relayManager,
		logger:       logger,
	}
}

// CreateTunnel создает WireGuard туннель для tenant'а
func (wgs *WireGuardService) CreateTunnel(ctx context.Context, req *CreateWireGuardTunnelRequest) (*CreateWireGuardTunnelResponse, error) {
	wgs.logger.Info("Creating WireGuard tunnel",
		zap.String("tenant_id", req.TenantId),
		zap.String("request_id", req.RequestId))
	
	// Создаем туннель
	wgTenant, err := wgs.relayManager.CreateWireGuardTunnel(req.TenantId)
	if err != nil {
		wgs.logger.Error("Failed to create WireGuard tunnel",
			zap.String("tenant_id", req.TenantId),
			zap.Error(err))
		return &CreateWireGuardTunnelResponse{
			Success:   false,
			Message:   fmt.Sprintf("Failed to create tunnel: %v", err),
			RequestId: req.RequestId,
		}, nil
	}
	
	// Получаем конфигурацию клиента
	clientConfig, err := wgs.relayManager.GetWireGuardClientConfig(req.TenantId)
	if err != nil {
		wgs.logger.Error("Failed to get client config",
			zap.String("tenant_id", req.TenantId),
			zap.Error(err))
		return &CreateWireGuardTunnelResponse{
			Success:   false,
			Message:   fmt.Sprintf("Failed to get client config: %v", err),
			RequestId: req.RequestId,
		}, nil
	}
	
	wgs.logger.Info("WireGuard tunnel created successfully",
		zap.String("tenant_id", req.TenantId),
		zap.String("public_key", wgTenant.PublicKey))
	
	return &CreateWireGuardTunnelResponse{
		Success:      true,
		Message:      "Tunnel created successfully",
		RequestId:    req.RequestId,
		PublicKey:    wgTenant.PublicKey,
		Endpoint:     wgTenant.Endpoint,
		Port:         int32(wgTenant.Port),
		ClientConfig: clientConfig,
	}, nil
}

// DeleteTunnel удаляет WireGuard туннель
func (wgs *WireGuardService) DeleteTunnel(ctx context.Context, req *DeleteWireGuardTunnelRequest) (*DeleteWireGuardTunnelResponse, error) {
	wgs.logger.Info("Deleting WireGuard tunnel",
		zap.String("tenant_id", req.TenantId),
		zap.String("request_id", req.RequestId))
	
	err := wgs.relayManager.DeleteWireGuardTunnel(req.TenantId)
	if err != nil {
		wgs.logger.Error("Failed to delete WireGuard tunnel",
			zap.String("tenant_id", req.TenantId),
			zap.Error(err))
		return &DeleteWireGuardTunnelResponse{
			Success:   false,
			Message:   fmt.Sprintf("Failed to delete tunnel: %v", err),
			RequestId: req.RequestId,
		}, nil
	}
	
	wgs.logger.Info("WireGuard tunnel deleted successfully",
		zap.String("tenant_id", req.TenantId))
	
	return &DeleteWireGuardTunnelResponse{
		Success:   true,
		Message:   "Tunnel deleted successfully",
		RequestId: req.RequestId,
	}, nil
}

// GetTunnel возвращает информацию о туннеле
func (wgs *WireGuardService) GetTunnel(ctx context.Context, req *GetWireGuardTunnelRequest) (*GetWireGuardTunnelResponse, error) {
	wgs.logger.Debug("Getting WireGuard tunnel info",
		zap.String("tenant_id", req.TenantId),
		zap.String("request_id", req.RequestId))
	
	wgTenant, err := wgs.relayManager.GetWireGuardTunnel(req.TenantId)
	if err != nil {
		wgs.logger.Error("Failed to get WireGuard tunnel",
			zap.String("tenant_id", req.TenantId),
			zap.Error(err))
		return &GetWireGuardTunnelResponse{
			Success:   false,
			Message:   fmt.Sprintf("Failed to get tunnel: %v", err),
			RequestId: req.RequestId,
		}, nil
	}
	
	return &GetWireGuardTunnelResponse{
		Success:   true,
		Message:   "Tunnel info retrieved successfully",
		RequestId: req.RequestId,
		Tunnel: &WireGuardTunnelInfo{
			TenantId:        wgTenant.TenantID,
			PublicKey:       wgTenant.PublicKey,
			Endpoint:        wgTenant.Endpoint,
			Port:            int32(wgTenant.Port),
			IsActive:        wgTenant.IsActive,
			BytesReceived:   wgTenant.BytesReceived,
			BytesSent:       wgTenant.BytesSent,
			LastConnected:   wgTenant.LastConnected.Unix(),
		},
	}, nil
}

// ListTunnels возвращает список всех туннелей
func (wgs *WireGuardService) ListTunnels(ctx context.Context, req *ListWireGuardTunnelsRequest) (*ListWireGuardTunnelsResponse, error) {
	wgs.logger.Debug("Listing WireGuard tunnels",
		zap.String("request_id", req.RequestId))
	
	tunnels := wgs.relayManager.ListWireGuardTunnels()
	
	var tunnelInfos []*WireGuardTunnelInfo
	for _, tunnel := range tunnels {
		tunnelInfos = append(tunnelInfos, &WireGuardTunnelInfo{
			TenantId:        tunnel.TenantID,
			PublicKey:       tunnel.PublicKey,
			Endpoint:        tunnel.Endpoint,
			Port:            int32(tunnel.Port),
			IsActive:        tunnel.IsActive,
			BytesReceived:   tunnel.BytesReceived,
			BytesSent:       tunnel.BytesSent,
			LastConnected:   tunnel.LastConnected.Unix(),
		})
	}
	
	wgs.logger.Debug("Listed WireGuard tunnels",
		zap.Int("count", len(tunnelInfos)))
	
	return &ListWireGuardTunnelsResponse{
		Success:   true,
		Message:   "Tunnels listed successfully",
		RequestId: req.RequestId,
		Tunnels:   tunnelInfos,
	}, nil
}

// GetStatus возвращает статус WireGuard сервера
func (wgs *WireGuardService) GetStatus(ctx context.Context, req *GetWireGuardStatusRequest) (*GetWireGuardStatusResponse, error) {
	wgs.logger.Debug("Getting WireGuard status",
		zap.String("request_id", req.RequestId))
	
	status, err := wgs.relayManager.GetWireGuardStatus()
	if err != nil {
		wgs.logger.Error("Failed to get WireGuard status",
			zap.Error(err))
		return &GetWireGuardStatusResponse{
			Success:   false,
			Message:   fmt.Sprintf("Failed to get status: %v", err),
			RequestId: req.RequestId,
		}, nil
	}
	
	return &GetWireGuardStatusResponse{
		Success:   true,
		Message:   "Status retrieved successfully",
		RequestId: req.RequestId,
		Status:    status,
	}, nil
}

// GetMetrics возвращает метрики WireGuard
func (wgs *WireGuardService) GetMetrics(ctx context.Context, req *GetWireGuardMetricsRequest) (*GetWireGuardMetricsResponse, error) {
	wgs.logger.Debug("Getting WireGuard metrics",
		zap.String("request_id", req.RequestId))
	
	metrics := wgs.relayManager.GetWireGuardMetrics()
	
	return &GetWireGuardMetricsResponse{
		Success:   true,
		Message:   "Metrics retrieved successfully",
		RequestId: req.RequestId,
		Metrics:   metrics,
	}, nil
}

// UpdateConfig обновляет конфигурацию WireGuard сервера
func (wgs *WireGuardService) UpdateConfig(ctx context.Context, req *UpdateWireGuardConfigRequest) (*UpdateWireGuardConfigResponse, error) {
	wgs.logger.Info("Updating WireGuard configuration",
		zap.String("request_id", req.RequestId))
	
	err := wgs.relayManager.UpdateWireGuardConfig()
	if err != nil {
		wgs.logger.Error("Failed to update WireGuard config",
			zap.Error(err))
		return &UpdateWireGuardConfigResponse{
			Success:   false,
			Message:   fmt.Sprintf("Failed to update config: %v", err),
			RequestId: req.RequestId,
		}, nil
	}
	
	wgs.logger.Info("WireGuard configuration updated successfully")
	
	return &UpdateWireGuardConfigResponse{
		Success:   true,
		Message:   "Configuration updated successfully",
		RequestId: req.RequestId,
	}, nil
}
