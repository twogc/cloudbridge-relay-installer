package relay

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/quic-go/quic-go"
	"go.uber.org/zap"
)

// StreamBridge мост для потоков через relay
type StreamBridge struct {
	SessionID    string
	Peer1Stream  quic.Stream
	Peer2Stream  quic.Stream
	BytesRelayed int64
	StartTime    time.Time
	LastActivity time.Time
	IsActive     bool
	mu           sync.RWMutex
	logger       *zap.Logger
	ctx          context.Context
	cancel       context.CancelFunc
}

// NewStreamBridge создает новый StreamBridge
func NewStreamBridge(sessionID string, peer1Stream, peer2Stream quic.Stream, logger *zap.Logger) *StreamBridge {
	ctx, cancel := context.WithCancel(context.Background())

	return &StreamBridge{
		SessionID:    sessionID,
		Peer1Stream:  peer1Stream,
		Peer2Stream:  peer2Stream,
		StartTime:    time.Now(),
		LastActivity: time.Now(),
		IsActive:     true,
		logger:       logger,
		ctx:          ctx,
		cancel:       cancel,
	}
}

// Start запускает релеирование потоков
func (sb *StreamBridge) Start() error {
	sb.logger.Info("Starting stream bridge",
		zap.String("session_id", sb.SessionID))

	// Запускаем релеирование в обе стороны
	go sb.relayPeer1ToPeer2()
	go sb.relayPeer2ToPeer1()

	return nil
}

// relayPeer1ToPeer2 релеирует данные от Peer1 к Peer2
func (sb *StreamBridge) relayPeer1ToPeer2() {
	defer func() {
		sb.mu.Lock()
		sb.IsActive = false
		sb.mu.Unlock()
	}()

	buffer := make([]byte, 4096)
	for sb.isActive() {
		select {
		case <-sb.ctx.Done():
			return
		default:
			// Читаем из Peer1
			n, err := sb.Peer1Stream.Read(buffer)
			if err != nil {
				if err == io.EOF {
					sb.logger.Debug("Peer1 stream closed",
						zap.String("session_id", sb.SessionID))
				} else {
					sb.logger.Error("Error reading from Peer1 stream",
						zap.String("session_id", sb.SessionID),
						zap.Error(err))
				}
				return
			}

			// Записываем в Peer2
			_, err = sb.Peer2Stream.Write(buffer[:n])
			if err != nil {
				sb.logger.Error("Error writing to Peer2 stream",
					zap.String("session_id", sb.SessionID),
					zap.Error(err))
				return
			}

			// Обновляем статистику
			sb.mu.Lock()
			sb.BytesRelayed += int64(n)
			sb.LastActivity = time.Now()
			sb.mu.Unlock()

			sb.logger.Debug("Relayed Peer1->Peer2",
				zap.String("session_id", sb.SessionID),
				zap.Int("bytes", n))
		}
	}
}

// relayPeer2ToPeer1 релеирует данные от Peer2 к Peer1
func (sb *StreamBridge) relayPeer2ToPeer1() {
	defer func() {
		sb.mu.Lock()
		sb.IsActive = false
		sb.mu.Unlock()
	}()

	buffer := make([]byte, 4096)
	for sb.isActive() {
		select {
		case <-sb.ctx.Done():
			return
		default:
			// Читаем из Peer2
			n, err := sb.Peer2Stream.Read(buffer)
			if err != nil {
				if err == io.EOF {
					sb.logger.Debug("Peer2 stream closed",
						zap.String("session_id", sb.SessionID))
				} else {
					sb.logger.Error("Error reading from Peer2 stream",
						zap.String("session_id", sb.SessionID),
						zap.Error(err))
				}
				return
			}

			// Записываем в Peer1
			_, err = sb.Peer1Stream.Write(buffer[:n])
			if err != nil {
				sb.logger.Error("Error writing to Peer1 stream",
					zap.String("session_id", sb.SessionID),
					zap.Error(err))
				return
			}

			// Обновляем статистику
			sb.mu.Lock()
			sb.BytesRelayed += int64(n)
			sb.LastActivity = time.Now()
			sb.mu.Unlock()

			sb.logger.Debug("Relayed Peer2->Peer1",
				zap.String("session_id", sb.SessionID),
				zap.Int("bytes", n))
		}
	}
}

// isActive проверяет, активен ли bridge
func (sb *StreamBridge) isActive() bool {
	sb.mu.RLock()
	defer sb.mu.RUnlock()
	return sb.IsActive
}

// GetStats возвращает статистику bridge
func (sb *StreamBridge) GetStats() *StreamBridgeStats {
	sb.mu.RLock()
	defer sb.mu.RUnlock()

	return &StreamBridgeStats{
		SessionID:    sb.SessionID,
		BytesRelayed: sb.BytesRelayed,
		StartTime:    sb.StartTime,
		LastActivity: sb.LastActivity,
		IsActive:     sb.IsActive,
		Duration:     time.Since(sb.StartTime),
	}
}

// Close закрывает stream bridge
func (sb *StreamBridge) Close() error {
	sb.cancel()

	sb.mu.Lock()
	sb.IsActive = false
	sb.mu.Unlock()

	// Закрываем потоки
	sb.Peer1Stream.Close()
	sb.Peer2Stream.Close()

	sb.logger.Info("Stream bridge closed",
		zap.String("session_id", sb.SessionID))

	return nil
}

// StreamBridgeStats статистика stream bridge
type StreamBridgeStats struct {
	SessionID    string
	BytesRelayed int64
	StartTime    time.Time
	LastActivity time.Time
	IsActive     bool
	Duration     time.Duration
}

// DatagramRelay будет реализован позже с правильными QUIC API
