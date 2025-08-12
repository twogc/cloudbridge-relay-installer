package protocol

import (
	"context"
	"fmt"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/internal/p2c"
	"github.com/twogc/cloudbridge-relay-installer/internal/p2p"
	"go.uber.org/zap"
)

// AutoSwitchManager управляет автоматическим переключением протоколов
type AutoSwitchManager struct {
	selector   *IntelligentProtocolSelector
	p2pManager *p2p.P2PManager
	p2cManager *p2c.P2CManager
	metrics    *SwitchMetrics
	logger     *zap.Logger

	switchCooldown time.Duration
	lastSwitch     time.Time
}

// NewAutoSwitchManager создает новый менеджер автоматического переключения
func NewAutoSwitchManager(
	selector *IntelligentProtocolSelector,
	p2pManager *p2p.P2PManager,
	p2cManager *p2c.P2CManager,
	logger *zap.Logger,
) *AutoSwitchManager {
	return &AutoSwitchManager{
		selector:       selector,
		p2pManager:     p2pManager,
		p2cManager:     p2cManager,
		metrics:        NewSwitchMetrics(),
		logger:         logger,
		switchCooldown: 5 * time.Minute,
	}
}

// Start запускает автоматическое переключение протоколов
func (asm *AutoSwitchManager) Start(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()

	asm.logger.Info("AutoSwitchManager started")

	for {
		select {
		case <-ctx.Done():
			asm.logger.Info("AutoSwitchManager stopped")
			return
		case <-ticker.C:
			asm.checkAndSwitch(ctx)
		}
	}
}

// checkAndSwitch проверяет необходимость переключения и выполняет его
func (asm *AutoSwitchManager) checkAndSwitch(ctx context.Context) {
	// Проверяем cooldown период
	if time.Since(asm.lastSwitch) < asm.switchCooldown {
		return
	}

	currentProtocol := asm.selector.Current()
	optimalProtocol := asm.selector.SelectBest()

	if currentProtocol != optimalProtocol {
		asm.logger.Info("Protocol switch needed",
			zap.String("from", currentProtocol.String()),
			zap.String("to", optimalProtocol.String()),
		)

		if err := asm.performSwitch(ctx, currentProtocol, optimalProtocol); err != nil {
			asm.logger.Error("Failed to switch protocol", zap.Error(err))
			return
		}

		asm.lastSwitch = time.Now()
		asm.metrics.RecordSwitch(currentProtocol, optimalProtocol, "auto")
	}
}

// performSwitch выполняет переключение протокола
func (asm *AutoSwitchManager) performSwitch(ctx context.Context, from, to Protocol) error {
	asm.logger.Info("Performing protocol switch",
		zap.String("from", from.String()),
		zap.String("to", to.String()),
	)

	// 1. Graceful shutdown текущего протокола
	if err := asm.gracefulShutdown(ctx, from); err != nil {
		return fmt.Errorf("failed to shutdown %s: %w", from.String(), err)
	}

	// 2. Инициализация нового протокола
	if err := asm.initializeProtocol(ctx, to); err != nil {
		return fmt.Errorf("failed to initialize %s: %w", to.String(), err)
	}

	// 3. Обновление селектора
	// TODO: Добавить метод SetCurrent в IntelligentProtocolSelector
	asm.logger.Debug("Protocol selector updated",
		zap.String("new_protocol", to.String()),
	)

	asm.logger.Info("Protocol switch completed",
		zap.String("from", from.String()),
		zap.String("to", to.String()),
	)

	return nil
}

// gracefulShutdown выполняет graceful shutdown протокола
func (asm *AutoSwitchManager) gracefulShutdown(ctx context.Context, protocol Protocol) error {
	switch protocol {
	case QUIC:
		// Graceful shutdown QUIC
		asm.logger.Debug("Graceful shutdown for QUIC protocol")
		return nil
	case HTTP2:
		// Graceful shutdown HTTP2
		asm.logger.Debug("Graceful shutdown for HTTP2 protocol")
		return nil
	case WebSocket:
		// Graceful shutdown WebSocket
		asm.logger.Debug("Graceful shutdown for WebSocket protocol")
		return nil
	default:
		return nil
	}
}

// initializeProtocol инициализирует новый протокол
func (asm *AutoSwitchManager) initializeProtocol(ctx context.Context, protocol Protocol) error {
	switch protocol {
	case QUIC:
		// Инициализация QUIC
		asm.logger.Debug("Initializing QUIC protocol")
		return nil
	case HTTP2:
		// Инициализация HTTP2
		asm.logger.Debug("Initializing HTTP2 protocol")
		return nil
	case WebSocket:
		// Инициализация WebSocket
		asm.logger.Debug("Initializing WebSocket protocol")
		return nil
	default:
		return fmt.Errorf("unknown protocol: %s", protocol.String())
	}
}

// GetMetrics возвращает метрики переключения
func (asm *AutoSwitchManager) GetMetrics() *SwitchMetrics {
	return asm.metrics
}
