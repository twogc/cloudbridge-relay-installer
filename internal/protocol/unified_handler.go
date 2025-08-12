package protocol

import (
	"context"
	"net"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/internal/p2c"
	"github.com/twogc/cloudbridge-relay-installer/internal/p2p"
	"go.uber.org/zap"
)

// UnifiedProtocolSelector интерфейс для выбора протокола в unified handler
type UnifiedProtocolSelector interface {
	SelectBest() Protocol
	Current() Protocol
	ShouldSwitch() bool
	GetMetrics() *ProtocolMetrics
	UpdateMetrics(protocol Protocol, latency time.Duration, success bool)
}

// P2PManagerInterface интерфейс для P2P менеджера
type P2PManagerInterface interface {
	IsAvailable(ctx context.Context, clientIP string, tenantID string) bool
	HandleConnection(ctx context.Context, conn net.Conn, tenantID string) error
	GetMetrics() *p2p.P2PMetrics
	UpdatePeerCount(count int)
}

// P2CManagerInterface интерфейс для P2C менеджера
type P2CManagerInterface interface {
	HandleConnection(ctx context.Context, conn net.Conn, tenantID string) error
	GetMetrics() *p2c.P2CMetrics
}

// UnifiedConnectionHandler унифицированный обработчик соединений
type UnifiedConnectionHandler struct {
	protocolSelector UnifiedProtocolSelector
	p2pManager       P2PManagerInterface
	p2cManager       P2CManagerInterface
	metrics          *ProtocolMetricsCollector
	logger           *zap.Logger
}

// NewUnifiedConnectionHandler создает новый унифицированный обработчик соединений
func NewUnifiedConnectionHandler(
	selector UnifiedProtocolSelector,
	p2pManager P2PManagerInterface,
	p2cManager P2CManagerInterface,
	logger *zap.Logger,
) *UnifiedConnectionHandler {
	return &UnifiedConnectionHandler{
		protocolSelector: selector,
		p2pManager:       p2pManager,
		p2cManager:       p2cManager,
		metrics:          NewProtocolMetricsCollector(5*time.Minute, logger),
		logger:           logger,
	}
}

// HandleConnection обрабатывает соединение через унифицированный интерфейс
func (h *UnifiedConnectionHandler) HandleConnection(ctx context.Context, conn net.Conn, tenantID string) error {
	start := time.Now()

	// 1. Определение оптимального протокола
	protocol := h.protocolSelector.SelectBest()
	h.logger.Info("Selected protocol for connection",
		zap.String("protocol", protocol.String()),
		zap.String("client_ip", conn.RemoteAddr().String()),
		zap.String("tenant_id", tenantID),
	)

	// 2. Обработка через выбранный протокол
	var err error
	switch protocol {
	case QUIC:
		err = h.handleQUICConnection(ctx, conn, tenantID)
	case HTTP2:
		err = h.handleHTTP2Connection(ctx, conn, tenantID)
	case WebSocket:
		err = h.handleWebSocketConnection(ctx, conn, tenantID)
	default:
		err = h.handleFallbackConnection(ctx, conn, tenantID)
	}

	// 3. Обновление метрик
	duration := time.Since(start)
	if err == nil {
		h.metrics.RecordSuccess(protocol)
		h.metrics.RecordLatency(protocol, duration)
	} else {
		h.metrics.RecordError(protocol, "connection_failed")
	}

	return err
}

// handleQUICConnection обрабатывает QUIC соединение
func (h *UnifiedConnectionHandler) handleQUICConnection(ctx context.Context, conn net.Conn, tenantID string) error {
	// Проверяем, доступен ли P2P
	if h.p2pManager.IsAvailable(ctx, conn.RemoteAddr().String(), tenantID) {
		h.logger.Info("Using P2P for QUIC connection",
			zap.String("tenant_id", tenantID),
		)
		return h.p2pManager.HandleConnection(ctx, conn, tenantID)
	}

	// Fallback на P2C через QUIC
	h.logger.Info("Using P2C fallback for QUIC connection",
		zap.String("tenant_id", tenantID),
	)
	return h.p2cManager.HandleConnection(ctx, conn, tenantID)
}

// handleHTTP2Connection обрабатывает HTTP2 соединение
func (h *UnifiedConnectionHandler) handleHTTP2Connection(ctx context.Context, conn net.Conn, tenantID string) error {
	h.logger.Info("Handling HTTP2 connection via P2C",
		zap.String("tenant_id", tenantID),
	)
	return h.p2cManager.HandleConnection(ctx, conn, tenantID)
}

// handleWebSocketConnection обрабатывает WebSocket соединение
func (h *UnifiedConnectionHandler) handleWebSocketConnection(ctx context.Context, conn net.Conn, tenantID string) error {
	h.logger.Info("Handling WebSocket connection via P2C",
		zap.String("tenant_id", tenantID),
	)
	return h.p2cManager.HandleConnection(ctx, conn, tenantID)
}

// handleFallbackConnection обрабатывает fallback соединение
func (h *UnifiedConnectionHandler) handleFallbackConnection(ctx context.Context, conn net.Conn, tenantID string) error {
	h.logger.Warn("Using fallback connection",
		zap.String("client_ip", conn.RemoteAddr().String()),
		zap.String("tenant_id", tenantID),
	)
	return h.p2cManager.HandleConnection(ctx, conn, tenantID)
}

// GetMetrics возвращает метрики обработчика
func (h *UnifiedConnectionHandler) GetMetrics() *ProtocolMetricsCollector {
	return h.metrics
}
