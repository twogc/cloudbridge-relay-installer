package quic

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/quic-go/quic-go"
	"go.uber.org/zap"
)

// QUICConnection represents a QUIC connection
type QUICConnection struct {
	conn      *quic.Conn
	streams   map[uint64]*Stream
	config    *Config
	tlsConfig *tls.Config
	logger    *zap.Logger
	mu        sync.RWMutex
	closed    bool
	ctx       context.Context
	cancel    context.CancelFunc

	// Metrics
	bytesReceived int64
	bytesSent     int64
	streamCount   int32
	lastActivity  time.Time

	// Enhanced QUIC поддержка
	enhancedConn   *EnhancedConnection
	tenantID       string
	enhancedConfig *EnhancedConfig
	enhancedMode   bool

	// QLog tracing support
	qlogTracer *QLogTracer
	connID     string

	// Network connection management
	pktConn net.PacketConn
}

// Stream represents a QUIC stream
type Stream struct {
	stream       *quic.Stream
	id           uint64
	conn         *QUICConnection
	closed       bool
	mu           sync.RWMutex
	lastActivity time.Time
}

// NewQUICConnection creates a new QUIC connection
func NewQUICConnection(config *Config, logger *zap.Logger) *QUICConnection {
	if config == nil {
		config = DefaultConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Генерируем уникальный ID для соединения
	connID := fmt.Sprintf("conn_%d", time.Now().UnixNano())

	// Создаем qlog трейсер если включен
	var qlogTracer *QLogTracer
	if config.EnableQLogTracing {
		qlogTracer = NewQLogTracer(logger, config.QLogConfig)
	}

	return &QUICConnection{
		streams:      make(map[uint64]*Stream),
		config:       config,
		tlsConfig:    config.GetTLSConfig(),
		logger:       logger,
		ctx:          ctx,
		cancel:       cancel,
		lastActivity: time.Now(),
		enhancedMode: false, // По умолчанию обычный режим
		qlogTracer:   qlogTracer,
		connID:       connID,
	}
}

// NewEnhancedQUICConnection создает новое QUIC соединение с Enhanced функциями
func NewEnhancedQUICConnection(config *Config, enhancedConfig *EnhancedConfig, tenantID string, logger *zap.Logger) *QUICConnection {
	if config == nil {
		config = DefaultConfig()
	}

	if enhancedConfig == nil {
		enhancedConfig = DefaultEnhancedConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Генерируем уникальный ID для соединения
	connID := fmt.Sprintf("enhanced_conn_%s_%d", tenantID, time.Now().UnixNano())

	// Создаем qlog трейсер если включен
	var qlogTracer *QLogTracer
	if config.EnableQLogTracing {
		qlogTracer = NewQLogTracer(logger, config.QLogConfig)
	}

	return &QUICConnection{
		streams:        make(map[uint64]*Stream),
		config:         config,
		tlsConfig:      config.GetTLSConfig(),
		logger:         logger,
		ctx:            ctx,
		cancel:         cancel,
		lastActivity:   time.Now(),
		tenantID:       tenantID,
		enhancedConfig: enhancedConfig,
		enhancedMode:   true,
		qlogTracer:     qlogTracer,
		connID:         connID,
	}
}

// Dial establishes a QUIC connection to the specified address
func (qc *QUICConnection) Dial(ctx context.Context, addr string) error {
	qc.logger.Info("Dialing QUIC connection", zap.String("addr", addr))

	// Если включен Enhanced режим, используем Enhanced соединение
	if qc.enhancedMode && qc.enhancedConfig != nil {
		return qc.dialEnhanced(ctx, addr)
	}

	// Обычное QUIC соединение
	return qc.dialRegular(ctx, addr)
}

// dialRegular устанавливает обычное QUIC соединение
func (qc *QUICConnection) dialRegular(ctx context.Context, addr string) error {
	// Create QUIC transport
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return fmt.Errorf("failed to resolve UDP address: %v", err)
	}

	udpConn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		return fmt.Errorf("failed to create UDP connection: %v", err)
	}
	qc.pktConn = udpConn

	// Create QUIC config
	quicConfig := &quic.Config{
		MaxIncomingStreams:             int64(qc.config.MaxIncomingStreams),
		MaxIncomingUniStreams:          int64(qc.config.MaxIncomingUniStreams),
		HandshakeIdleTimeout:           qc.config.HandshakeTimeout,
		MaxIdleTimeout:                 qc.config.IdleTimeout,
		KeepAlivePeriod:                20 * time.Second,
		InitialStreamReceiveWindow:     qc.config.InitialStreamReceiveWindow,
		MaxStreamReceiveWindow:         qc.config.MaxStreamReceiveWindow,
		InitialConnectionReceiveWindow: qc.config.InitialConnectionReceiveWindow,
		MaxConnectionReceiveWindow:     qc.config.MaxConnectionReceiveWindow,
	}

	// Establish QUIC connection
	conn, err := quic.Dial(ctx, udpConn, udpAddr, qc.tlsConfig, quicConfig)
	if err != nil {
		return fmt.Errorf("failed to establish QUIC connection: %v", err)
	}

	qc.conn = conn
	qc.lastActivity = time.Now()

	// Начинаем qlog трейсинг если включен
	if qc.qlogTracer != nil {
		qc.qlogTracer.StartConnection(qc.connID, conn)
	}

	qc.logger.Info("QUIC connection established",
		zap.String("addr", addr),
		zap.String("conn_id", qc.connID),
	)

	// Start stream management
	go qc.manageStreams()

	return nil
}

// dialEnhanced устанавливает Enhanced QUIC соединение
func (qc *QUICConnection) dialEnhanced(ctx context.Context, addr string) error {
	qc.logger.Info("Dialing Enhanced QUIC connection",
		zap.String("addr", addr),
		zap.String("tenant_id", qc.tenantID),
	)

	// Сначала устанавливаем обычное QUIC соединение
	err := qc.dialRegular(ctx, addr)
	if err != nil {
		return fmt.Errorf("failed to establish base QUIC connection for enhanced mode: %v", err)
	}

	// Затем создаем Enhanced политики
	enhancedConn, err := NewEnhancedPolicies(qc.tenantID, qc.enhancedConfig)
	if err != nil {
		// Если Enhanced политики не удалось создать, закрываем базовое
		if qc.conn != nil {
			_ = qc.conn.CloseWithError(0, "enhanced policies failed")
		}
		return fmt.Errorf("failed to create Enhanced policies: %v", err)
	}

	qc.enhancedConn = enhancedConn
	qc.lastActivity = time.Now()

	// Начинаем qlog трейсинг если включен
	if qc.qlogTracer != nil {
		qc.qlogTracer.StartConnection(qc.connID, qc.conn)
	}

	qc.logger.Info("Enhanced QUIC connection established",
		zap.String("addr", addr),
		zap.String("tenant_id", qc.tenantID),
		zap.String("conn_id", qc.connID),
	)

	return nil
}

// AcceptStream accepts a new stream from the connection
func (qc *QUICConnection) AcceptStream(ctx context.Context) (*Stream, error) {
	if qc.conn == nil {
		return nil, fmt.Errorf("connection not established")
	}

	stream, err := qc.conn.AcceptStream(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to accept stream: %v", err)
	}

	streamID := stream.StreamID()
	if streamID < 0 {
		return nil, fmt.Errorf("invalid stream ID: %d", streamID)
	}
	// streamID is int64, so it cannot be greater than math.MaxInt64
	// This check is redundant but kept for clarity
	//nolint:gosec // #nosec G115: streamID не может быть отрицательным, переполнения не будет
	id := uint64(streamID) // #nosec G115
	quicStream := &Stream{
		stream:       stream,
		id:           id,
		conn:         qc,
		lastActivity: time.Now(),
	}

	qc.mu.Lock()
	qc.streams[id] = quicStream
	qc.streamCount++
	qc.lastActivity = time.Now()
	qc.mu.Unlock()

	qc.logger.Debug("Accepted QUIC stream",
		zap.Uint64("stream_id", id),
	)

	return quicStream, nil
}

// OpenStream opens a new stream on the connection
func (qc *QUICConnection) OpenStream(ctx context.Context) (*Stream, error) {
	if qc.conn == nil {
		return nil, fmt.Errorf("connection not established")
	}

	stream, err := qc.conn.OpenStreamSync(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to open stream: %v", err)
	}

	streamID := stream.StreamID()
	if streamID < 0 {
		return nil, fmt.Errorf("invalid stream ID: %d", streamID)
	}
	// streamID is int64, so it cannot be greater than math.MaxInt64
	// This check is redundant but kept for clarity
	//nolint:gosec // #nosec G115: streamID не может быть отрицательным, переполнения не будет
	id := uint64(streamID) // #nosec G115
	quicStream := &Stream{
		stream:       stream,
		id:           id,
		conn:         qc,
		lastActivity: time.Now(),
	}

	qc.mu.Lock()
	qc.streams[id] = quicStream
	qc.streamCount++
	qc.lastActivity = time.Now()
	qc.mu.Unlock()

	qc.logger.Debug("Opened QUIC stream",
		zap.Uint64("stream_id", id),
	)

	return quicStream, nil
}

// manageStreams manages active streams
func (qc *QUICConnection) manageStreams() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-qc.ctx.Done():
			return
		case <-ticker.C:
			qc.cleanupStreams()
		}
	}
}

// cleanupStreams removes closed streams
func (qc *QUICConnection) cleanupStreams() {
	qc.mu.Lock()
	defer qc.mu.Unlock()

	cutoff := time.Now().Add(-5 * time.Minute)
	for id, stream := range qc.streams {
		stream.mu.RLock()
		closed := stream.closed
		inactive := stream.lastActivity.Before(cutoff)
		stream.mu.RUnlock()
		
		if closed || inactive {
			_ = stream.Close() // аккуратно закрыть перед delete
			delete(qc.streams, id)
			if qc.streamCount > 0 {
				qc.streamCount--
			}
			qc.logger.Debug("Cleaned up QUIC stream",
				zap.Uint64("stream_id", id),
			)
		}
	}
}

// Close closes the QUIC connection
func (qc *QUICConnection) Close() error {
	qc.mu.Lock()
	defer qc.mu.Unlock()

	if qc.closed {
		return nil
	}

	qc.closed = true
	qc.cancel()

	// Завершаем qlog трейсинг если включен
	if qc.qlogTracer != nil {
		qc.qlogTracer.EndConnection(qc.connID, "connection_closed")
	}

	// Close all streams
	for _, stream := range qc.streams {
		if err := stream.Close(); err != nil {
			qc.logger.Error("Error closing QUIC stream", zap.Error(err))
		}
	}

	// Close Enhanced connection if available
	if qc.enhancedConn != nil {
		if err := qc.enhancedConn.Close(); err != nil {
			qc.logger.Error("Error closing Enhanced QUIC connection", zap.Error(err))
		}
	}

	// Close regular connection
	if qc.conn != nil {
		if err := qc.conn.CloseWithError(0, "connection closed"); err != nil {
			qc.logger.Error("Error closing QUIC connection", zap.Error(err))
		}
	}

	// Close packet connection
	if qc.pktConn != nil {
		_ = qc.pktConn.Close()
		qc.pktConn = nil
	}

	qc.logger.Info("QUIC connection closed",
		zap.String("conn_id", qc.connID),
	)
	return nil
}

// IsHealthy checks if the connection is healthy
func (qc *QUICConnection) IsHealthy() bool {
	qc.mu.RLock()
	defer qc.mu.RUnlock()

	if qc.closed {
		return false
	}

	// Проверяем обычное соединение
	if qc.conn != nil && time.Since(qc.lastActivity) < qc.config.IdleTimeout {
		return true
	}

	// Если enhanced-режим, проверяем enhanced соединение
	if qc.enhancedMode && qc.enhancedConn != nil {
		return qc.enhancedConn.IsHealthy()
	}

	return false
}

// GetStats returns connection statistics
func (qc *QUICConnection) GetStats() map[string]interface{} {
	qc.mu.RLock()
	defer qc.mu.RUnlock()

	stats := map[string]interface{}{
		"conn_id":        qc.connID,
		"bytes_received": qc.bytesReceived,
		"bytes_sent":     qc.bytesSent,
		"stream_count":   qc.streamCount,
		"last_activity":  qc.lastActivity,
		"closed":         qc.closed,
		"enhanced_mode":  qc.enhancedMode,
		"tenant_id":      qc.tenantID,
	}

	// Добавляем Enhanced статистику, если доступна
	if qc.enhancedConn != nil {
		enhancedStats := qc.enhancedConn.GetStats()
		stats["enhanced_stats"] = enhancedStats
	}

	return stats
}

// Stream methods

// Read reads data from the stream
func (s *Stream) Read(p []byte) (n int, err error) {
	// Вызываем BeforeRead хук если есть Enhanced политики
	if s.conn.enhancedMode && s.conn.enhancedConn != nil {
		remoteAddr := "unknown"
		if s.conn.conn != nil {
			remoteAddr = s.conn.conn.RemoteAddr().String()
		}
		if err := s.conn.enhancedConn.BeforeRead(remoteAddr, len(p)); err != nil {
			return 0, err
		}
	}

	// Читаем из QUIC потока
	n, err = (*s.stream).Read(p)

	if err != nil {
		if err == io.EOF {
			if closeErr := s.Close(); closeErr != nil {
				s.conn.logger.Error("Error closing stream on EOF", zap.Error(closeErr))
			}
		}
		return n, err
	}

	// Вызываем AfterRead хук если есть Enhanced политики
	if s.conn.enhancedMode && s.conn.enhancedConn != nil {
		s.conn.enhancedConn.AfterRead(n)
	}

	// Update connection metrics
	s.conn.mu.Lock()
	s.conn.bytesReceived += int64(n)
	s.conn.lastActivity = time.Now()
	s.conn.mu.Unlock()

	// Update stream activity after successful read
	s.mu.Lock()
	s.lastActivity = time.Now()
	s.mu.Unlock()

	// Логируем получение данных в qlog если включен
	if s.conn.qlogTracer != nil && n > 0 {
		s.conn.qlogTracer.LogStreamEvent(s.conn.connID, s.id, "stream_data_received", map[string]interface{}{
			"bytes": n,
		})
	}

	return n, nil
}

// Write writes data to the stream
func (s *Stream) Write(p []byte) (n int, err error) {
	// Вызываем BeforeWrite хук если есть Enhanced политики
	if s.conn.enhancedMode && s.conn.enhancedConn != nil {
		remoteAddr := "unknown"
		if s.conn.conn != nil {
			remoteAddr = s.conn.conn.RemoteAddr().String()
		}
		if err := s.conn.enhancedConn.BeforeWrite(remoteAddr, len(p)); err != nil {
			return 0, err
		}
	}

	// Записываем в QUIC поток
	n, err = (*s.stream).Write(p)

	if err != nil {
		return n, err
	}

	// Вызываем AfterWrite хук если есть Enhanced политики
	if s.conn.enhancedMode && s.conn.enhancedConn != nil {
		s.conn.enhancedConn.AfterWrite(n)
	}

	// Update connection metrics
	s.conn.mu.Lock()
	s.conn.bytesSent += int64(n)
	s.conn.lastActivity = time.Now()
	s.conn.mu.Unlock()

	// Update stream activity after successful write
	s.mu.Lock()
	s.lastActivity = time.Now()
	s.mu.Unlock()

	// Логируем отправку данных в qlog если включен
	if s.conn.qlogTracer != nil && n > 0 {
		s.conn.qlogTracer.LogStreamEvent(s.conn.connID, s.id, "stream_data_sent", map[string]interface{}{
			"bytes": n,
		})
	}

	return n, nil
}

// Close closes the stream
func (s *Stream) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}

	s.closed = true
	return (*s.stream).Close()
}

// StreamID returns the stream ID
func (s *Stream) StreamID() uint64 {
	return s.id
}

// IsClosed returns true if the stream is closed
func (s *Stream) IsClosed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.closed
}

// CreateEnhancedConnection создает Enhanced соединение для существующего QUIC соединения
func (qc *QUICConnection) CreateEnhancedConnection(tenantID string, config *EnhancedConfig) error {
	qc.mu.Lock()
	defer qc.mu.Unlock()

	if qc.enhancedConn != nil {
		return fmt.Errorf("enhanced connection already exists")
	}

	if config == nil {
		config = DefaultEnhancedConfig()
	}

	// Настраиваем Enhanced конфигурацию на основе обычной
	config.MaxStreams = qc.config.MaxStreams
	config.IdleTimeout = qc.config.IdleTimeout
	config.HandshakeTimeout = qc.config.HandshakeTimeout
	config.MaxIncomingStreams = qc.config.MaxIncomingStreams
	config.MaxIncomingUniStreams = qc.config.MaxIncomingUniStreams
	config.TLSConfig = qc.tlsConfig

	// Создаем Enhanced политики
	enhancedConn, err := NewEnhancedPolicies(tenantID, config)
	if err != nil {
		return fmt.Errorf("failed to create Enhanced policies: %v", err)
	}

	qc.enhancedConn = enhancedConn
	qc.tenantID = tenantID
	qc.enhancedConfig = config
	qc.enhancedMode = true

	qc.logger.Info("Enhanced QUIC connection created",
		zap.String("tenant_id", tenantID),
	)

	return nil
}

// GetEnhancedConnection возвращает Enhanced соединение
func (qc *QUICConnection) GetEnhancedConnection() *EnhancedConnection {
	qc.mu.RLock()
	defer qc.mu.RUnlock()
	return qc.enhancedConn
}

// IsEnhancedMode проверяет, включен ли Enhanced режим
func (qc *QUICConnection) IsEnhancedMode() bool {
	qc.mu.RLock()
	defer qc.mu.RUnlock()
	return qc.enhancedMode
}

// GetTenantID возвращает ID тенанта
func (qc *QUICConnection) GetTenantID() string {
	qc.mu.RLock()
	defer qc.mu.RUnlock()
	return qc.tenantID
}

// GetConnectionID возвращает ID соединения
func (qc *QUICConnection) GetConnectionID() string {
	qc.mu.RLock()
	defer qc.mu.RUnlock()
	return qc.connID
}

// GetQLogTracer возвращает qlog трейсер
func (qc *QUICConnection) GetQLogTracer() *QLogTracer {
	qc.mu.RLock()
	defer qc.mu.RUnlock()
	return qc.qlogTracer
}

// EnableQLogTracing включает qlog трейсинг
func (qc *QUICConnection) EnableQLogTracing() {
	qc.mu.Lock()
	defer qc.mu.Unlock()

	if qc.qlogTracer == nil {
		qc.qlogTracer = NewQLogTracer(qc.logger, qc.config.QLogConfig)
		if qc.conn != nil {
			qc.qlogTracer.StartConnection(qc.connID, qc.conn)
		}
	}
}

// DisableQLogTracing отключает qlog трейсинг
func (qc *QUICConnection) DisableQLogTracing() {
	qc.mu.Lock()
	defer qc.mu.Unlock()

	if qc.qlogTracer != nil {
		qc.qlogTracer.EndConnection(qc.connID, "tracing_disabled")
		qc.qlogTracer = nil
	}
}
