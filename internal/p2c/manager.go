package p2c

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/internal/tenant"
	"github.com/twogc/cloudbridge-relay-installer/internal/balancer"
	"github.com/twogc/cloudbridge-relay-installer/internal/ratelimit"
	"go.uber.org/zap"
)

// EnhancedQUICManager представляет менеджер Enhanced QUIC
type EnhancedQUICManager struct {
	logger *zap.Logger
}

// NewEnhancedQUICManager создает новый Enhanced QUIC менеджер
func NewEnhancedQUICManager(logger *zap.Logger) *EnhancedQUICManager {
	return &EnhancedQUICManager{
		logger: logger,
	}
}

// HandleConnection обрабатывает соединение через Enhanced QUIC
func (eqm *EnhancedQUICManager) HandleConnection(ctx context.Context, conn net.Conn, targetServer string, tenantID string) error {
	eqm.logger.Info("Handling connection via Enhanced QUIC",
		zap.String("target_server", targetServer),
		zap.String("tenant_id", tenantID),
	)
	// TODO: Реализовать обработку через Enhanced QUIC
	return nil
}

// P2CManager управляет P2C соединениями
type P2CManager struct {
	enhancedQUICManager *EnhancedQUICManager
	tenantManager       *tenant.Manager
	loadBalancer        *balancer.Balancer
	ipRateLimit         *ratelimit.IPLimiter
	metrics             *P2CMetrics
	logger              *zap.Logger
}

// NewP2CManager создает новый P2C менеджер
func NewP2CManager(
	quicManager *EnhancedQUICManager,
	tenantManager *tenant.Manager,
	loadBalancer *balancer.Balancer,
	ipRateLimit *ratelimit.IPLimiter,
	logger *zap.Logger,
) *P2CManager {
	return &P2CManager{
		enhancedQUICManager: quicManager,
		tenantManager:       tenantManager,
		loadBalancer:        loadBalancer,
		ipRateLimit:         ipRateLimit,
		metrics:             NewP2CMetrics(),
		logger:              logger,
	}
}

// HandleConnection обрабатывает P2C соединение
func (pc *P2CManager) HandleConnection(ctx context.Context, conn net.Conn, tenantID string) error {
	start := time.Now()

	// 1. Проверка лимитов тенанта
	if err := pc.tenantManager.CheckTenantLimits(tenantID, 1, 1); err != nil {
		pc.metrics.RecordConnection("p2c", "tenant_limit_exceeded", tenantID, time.Since(start))
		return fmt.Errorf("tenant limits exceeded: %w", err)
	}

	// 2. Проверка rate limiting
	clientIP := conn.RemoteAddr().String()
	if !pc.ipRateLimit.Allow(clientIP) {
		pc.metrics.RecordConnection("p2c", "rate_limited", tenantID, time.Since(start))
		return fmt.Errorf("rate limit exceeded for IP %s", clientIP)
	}

	// 3. Получение оптимального сервера
	// Временная заглушка - используем дефолтный сервер
	// TODO: Реализовать получение оптимального сервера через load balancer
	targetServer := "default-server"
	pc.logger.Debug("Using default server for P2C connection",
		zap.String("target_server", targetServer),
		zap.String("tenant_id", tenantID),
	)

	// 4. Обработка через Enhanced QUIC
	err := pc.enhancedQUICManager.HandleConnection(ctx, conn, targetServer, tenantID)

	// 5. Обновление метрик
	status := "success"
	if err != nil {
		status = "failed"
	}
	pc.metrics.RecordConnection("p2c", status, tenantID, time.Since(start))

	// 6. Обновление активности тенанта
	pc.tenantManager.UpdateTenantActivity(tenantID)

	return err
}

// GetMetrics возвращает метрики P2C
func (pc *P2CManager) GetMetrics() *P2CMetrics {
	return pc.metrics
}
