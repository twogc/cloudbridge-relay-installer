package quic

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/quic-go/quic-go"
	"go.uber.org/zap"

	"github.com/twogc/cloudbridge-relay-installer/internal/enhanced/metrics"
)

// QUICConnection представляет QUIC соединение для CloudBridge Relay
type QUICConnection struct {
	conn          *quic.Conn
	logger        *zap.Logger
	metrics       *metrics.EnhancedMetrics
	connectionID  uuid.UUID
	streams       map[quic.StreamID]*quic.Stream
	mu            sync.RWMutex
	config        *QUICConfig
	startTime     time.Time
	activeStreams int32
	totalStreams  int64
}

// QUICConfig конфигурация для QUIC соединения
type QUICConfig struct {
	MaxStreams        int
	KeepAliveInterval time.Duration
	IdleTimeout       time.Duration
	MaxIdleTime       time.Duration
	EnableMetrics     bool
	EnableTracing     bool
}

// DefaultQUICConfig возвращает конфигурацию по умолчанию
func DefaultQUICConfig() *QUICConfig {
	return &QUICConfig{
		MaxStreams:        100,
		KeepAliveInterval: 30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxIdleTime:       300 * time.Second,
		EnableMetrics:     true,
		EnableTracing:     false,
	}
}

// NewQUICConnection создает новое QUIC соединение
func NewQUICConnection(
	conn *quic.Conn,
	logger *zap.Logger,
	metrics *metrics.EnhancedMetrics,
	config *QUICConfig,
) *QUICConnection {
	if config == nil {
		config = DefaultQUICConfig()
	}

	return &QUICConnection{
		conn:         conn,
		logger:       logger,
		metrics:      metrics,
		connectionID: uuid.New(),
		streams:      make(map[quic.StreamID]*quic.Stream),
		config:       config,
		startTime:    time.Now(),
	}
}

// Serve запускает QUIC сервер для обработки соединений
func (qc *QUICConnection) Serve(ctx context.Context) error {
	qc.logger.Info("Starting QUIC connection server",
		zap.String("connection_id", qc.connectionID.String()),
		zap.String("remote_addr", qc.conn.RemoteAddr().String()),
	)

	// Запускаем keep-alive механизм
	if qc.config.KeepAliveInterval > 0 {
		go qc.keepAlive(ctx)
	}

	// Запускаем сборщик метрик
	if qc.config.EnableMetrics {
		go qc.metricsCollector(ctx)
	}

	// Основной цикл обработки потоков
	for {
		select {
		case <-ctx.Done():
			qc.logger.Info("QUIC connection context cancelled",
				zap.String("connection_id", qc.connectionID.String()),
			)
			return ctx.Err()

		default:
			// Принимаем новый поток с таймаутом
			acceptCtx, cancel := context.WithTimeout(ctx, qc.config.IdleTimeout)
			stream, err := qc.conn.AcceptStream(acceptCtx)
			cancel()

			if err != nil {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				qc.logger.Debug("Failed to accept QUIC stream",
					zap.Error(err),
					zap.String("connection_id", qc.connectionID.String()),
				)
				continue
			}

			// Обрабатываем поток в отдельной горутине
			go qc.handleStream(ctx, stream)
		}
	}
}

// handleStream обрабатывает отдельный QUIC поток
func (qc *QUICConnection) handleStream(ctx context.Context, stream *quic.Stream) {
	streamID := stream.StreamID()
	startTime := time.Now()

	qc.mu.Lock()
	qc.streams[streamID] = stream
	qc.activeStreams++
	qc.totalStreams++
	qc.mu.Unlock()

	// Обновляем метрики
	if qc.metrics != nil {
		qc.metrics.IncQUICStreams("inbound", "data")
		qc.metrics.SetQUICStreamsActive("inbound", "data", int(qc.activeStreams))
	}

	qc.logger.Debug("Started handling QUIC stream",
		zap.Uint64("stream_id", uint64(streamID)),
		zap.String("connection_id", qc.connectionID.String()),
	)

	defer func() {
		// Очищаем ресурсы
		qc.mu.Lock()
		delete(qc.streams, streamID)
		qc.activeStreams--
		qc.mu.Unlock()

		if err := stream.Close(); err != nil {
			qc.logger.Warn("Failed to close QUIC stream", zap.Error(err))
		}
		duration := time.Since(startTime)

		// Обновляем метрики
		if qc.metrics != nil {
			qc.metrics.ObserveQUICStreamLatency("inbound", "data", duration)
		}

		qc.logger.Debug("Finished handling QUIC stream",
			zap.Uint64("stream_id", uint64(streamID)),
			zap.Duration("duration", duration),
			zap.String("connection_id", qc.connectionID.String()),
		)
	}()

	// Обрабатываем поток
	if err := qc.processStream(ctx, stream); err != nil {
		qc.logger.Error("Error processing QUIC stream",
			zap.Error(err),
			zap.Uint64("stream_id", uint64(streamID)),
			zap.String("connection_id", qc.connectionID.String()),
		)

		// Обновляем метрики ошибок
		if qc.metrics != nil {
			qc.metrics.IncQUICConnectionErrors("stream_error")
		}

		// Отменяем запись в поток
		stream.CancelWrite(0)
	}
}

// processStream обрабатывает данные потока
func (qc *QUICConnection) processStream(ctx context.Context, stream *quic.Stream) error {
	// Создаем буфер для чтения данных
	buffer := make([]byte, 4096)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		default:
			// Читаем данные с таймаутом
			if err := stream.SetReadDeadline(time.Now().Add(qc.config.IdleTimeout)); err != nil {
				qc.logger.Warn("Failed to set read deadline", zap.Error(err))
			}
			n, err := stream.Read(buffer)

			if err != nil {
				if err.Error() == "timeout" {
					// Таймаут - это нормально, продолжаем
					continue
				}
				return fmt.Errorf("failed to read from stream: %w", err)
			}

			if n > 0 {
				// Обрабатываем полученные данные
				if err := qc.handleData(buffer[:n], stream); err != nil {
					return fmt.Errorf("failed to handle data: %w", err)
				}
			}
		}
	}
}

// handleData обрабатывает полученные данные
func (qc *QUICConnection) handleData(data []byte, stream *quic.Stream) error {
	// TODO: Реализовать обработку данных в соответствии с протоколом CloudBridge
	// Пока просто эхо-ответ
	_, err := stream.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write to stream: %w", err)
	}

	// Обновляем метрики
	if qc.metrics != nil {
		qc.metrics.IncIngressBandwidth("quic", "in", int64(len(data)))
		qc.metrics.IncIngressBandwidth("quic", "out", int64(len(data)))
	}

	return nil
}

// keepAlive поддерживает соединение активным
func (qc *QUICConnection) keepAlive(ctx context.Context) {
	ticker := time.NewTicker(qc.config.KeepAliveInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Отправляем ping для поддержания соединения
			if err := qc.sendPing(); err != nil {
				qc.logger.Warn("Failed to send keep-alive ping",
					zap.Error(err),
					zap.String("connection_id", qc.connectionID.String()),
				)
			}
		}
	}
}

// sendPing отправляет ping для поддержания соединения
func (qc *QUICConnection) sendPing() error {
	// Открываем новый поток для ping
	stream, err := qc.conn.OpenStream()
	if err != nil {
		return fmt.Errorf("failed to open ping stream: %w", err)
	}
	defer stream.Close()

	// Отправляем ping данные
	pingData := []byte("ping")
	_, err = stream.Write(pingData)
	if err != nil {
		return fmt.Errorf("failed to write ping: %w", err)
	}

	return nil
}

// metricsCollector собирает метрики соединения
func (qc *QUICConnection) metricsCollector(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			qc.updateMetrics()
		}
	}
}

// updateMetrics обновляет метрики соединения
func (qc *QUICConnection) updateMetrics() {
	if qc.metrics == nil {
		return
	}

	qc.mu.RLock()
	activeStreams := qc.activeStreams
	totalStreams := qc.totalStreams
	uptime := time.Since(qc.startTime)
	qc.mu.RUnlock()

	// Обновляем метрики
	qc.metrics.SetQUICStreamsActive("inbound", "data", int(activeStreams))
	qc.metrics.ObserveProtocolLatency("quic", "connection_uptime", uptime)

	// Логируем статистику
	qc.logger.Debug("QUIC connection metrics",
		zap.Int32("active_streams", activeStreams),
		zap.Int64("total_streams", totalStreams),
		zap.Duration("uptime", uptime),
		zap.String("connection_id", qc.connectionID.String()),
	)
}

// Close закрывает QUIC соединение
func (qc *QUICConnection) Close() error {
	qc.logger.Info("Closing QUIC connection",
		zap.String("connection_id", qc.connectionID.String()),
		zap.Duration("uptime", time.Since(qc.startTime)),
	)

	// Закрываем все активные потоки
	qc.mu.Lock()
	for streamID, stream := range qc.streams {
		if err := stream.Close(); err != nil {
			qc.logger.Warn("Failed to close QUIC stream", zap.Error(err))
		}
		delete(qc.streams, streamID)
	}
	qc.mu.Unlock()

	// Закрываем соединение
	return qc.conn.CloseWithError(0, "normal shutdown")
}

// GetStats возвращает статистику соединения
func (qc *QUICConnection) GetStats() map[string]interface{} {
	qc.mu.RLock()
	defer qc.mu.RUnlock()

	return map[string]interface{}{
		"connection_id":  qc.connectionID.String(),
		"remote_addr":    qc.conn.RemoteAddr().String(),
		"active_streams": qc.activeStreams,
		"total_streams":  qc.totalStreams,
		"uptime":         time.Since(qc.startTime).String(),
		"max_streams":    qc.config.MaxStreams,
		"keep_alive":     qc.config.KeepAliveInterval.String(),
		"idle_timeout":   qc.config.IdleTimeout.String(),
		"enable_metrics": qc.config.EnableMetrics,
		"enable_tracing": qc.config.EnableTracing,
	}
}

// IsActive проверяет, активно ли соединение
func (qc *QUICConnection) IsActive() bool {
	qc.mu.RLock()
	defer qc.mu.RUnlock()
	return len(qc.streams) > 0 || time.Since(qc.startTime) < qc.config.MaxIdleTime
}

// GetConnectionID возвращает ID соединения
func (qc *QUICConnection) GetConnectionID() uuid.UUID {
	return qc.connectionID
}

// GetActiveStreams возвращает количество активных потоков
func (qc *QUICConnection) GetActiveStreams() int32 {
	qc.mu.RLock()
	defer qc.mu.RUnlock()
	return qc.activeStreams
}

// GetTotalStreams возвращает общее количество потоков
func (qc *QUICConnection) GetTotalStreams() int64 {
	qc.mu.RLock()
	defer qc.mu.RUnlock()
	return qc.totalStreams
}
