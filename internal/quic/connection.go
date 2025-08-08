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
	conn        *quic.Conn
	streams     map[uint64]*Stream
	config      *Config
	tlsConfig   *tls.Config
	logger      *zap.Logger
	mu          sync.RWMutex
	closed      bool
	ctx         context.Context
	cancel      context.CancelFunc

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
}

// Stream represents a QUIC stream
type Stream struct {
	stream      *quic.Stream
	id          uint64
	conn        *QUICConnection
	closed      bool
	mu          sync.RWMutex
	lastActivity time.Time
}

// NewQUICConnection creates a new QUIC connection
func NewQUICConnection(config *Config, logger *zap.Logger) *QUICConnection {
	if config == nil {
		config = DefaultConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &QUICConnection{
		streams:    make(map[uint64]*Stream),
		config:     config,
		tlsConfig:  config.GetTLSConfig(),
		logger:     logger,
		ctx:        ctx,
		cancel:     cancel,
		lastActivity: time.Now(),
		enhancedMode: false, // По умолчанию обычный режим
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

	// Create QUIC config
	quicConfig := &quic.Config{
		MaxIncomingStreams:    int64(qc.config.MaxIncomingStreams),
		MaxIncomingUniStreams: int64(qc.config.MaxIncomingUniStreams),
		HandshakeIdleTimeout:  qc.config.HandshakeTimeout,
		MaxIdleTimeout:        qc.config.IdleTimeout,
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

	qc.logger.Info("QUIC connection established",
		zap.String("addr", addr),
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

	// Создаем Enhanced соединение
	enhancedConn, err := DialWithTenant("udp", addr, qc.tenantID, qc.enhancedConfig)
	if err != nil {
		return fmt.Errorf("failed to establish Enhanced QUIC connection: %v", err)
	}

	qc.enhancedConn = enhancedConn
	qc.lastActivity = time.Now()

	qc.logger.Info("Enhanced QUIC connection established",
		zap.String("addr", addr),
		zap.String("tenant_id", qc.tenantID),
	)

	// Start stream management
	go qc.manageStreams()

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
	if streamID > 9223372036854775807 { // Max int64
		return nil, fmt.Errorf("stream ID too large: %d", streamID)
	}
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
	if streamID > 9223372036854775807 { // Max int64
		return nil, fmt.Errorf("stream ID too large: %d", streamID)
	}
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
		if stream.closed || stream.lastActivity.Before(cutoff) {
			delete(qc.streams, id)
			qc.streamCount--
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

	qc.logger.Info("QUIC connection closed")
	return nil
}

// IsHealthy checks if the connection is healthy
func (qc *QUICConnection) IsHealthy() bool {
	qc.mu.RLock()
	defer qc.mu.RUnlock()

	if qc.closed || qc.conn == nil {
		return false
	}

	// Check if connection is still active
	return time.Since(qc.lastActivity) < qc.config.IdleTimeout
}

// GetStats returns connection statistics
func (qc *QUICConnection) GetStats() map[string]interface{} {
	qc.mu.RLock()
	defer qc.mu.RUnlock()

	stats := map[string]interface{}{
		"bytes_received":   qc.bytesReceived,
		"bytes_sent":       qc.bytesSent,
		"stream_count":     qc.streamCount,
		"last_activity":    qc.lastActivity,
		"closed":           qc.closed,
		"enhanced_mode":    qc.enhancedMode,
		"tenant_id":        qc.tenantID,
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
	s.mu.Lock()
	s.lastActivity = time.Now()
	s.mu.Unlock()

	// Используем Enhanced соединение, если доступно
	if s.conn.enhancedMode && s.conn.enhancedConn != nil {
		n, err = s.conn.enhancedConn.Read(p)
	} else {
		n, err = (*s.stream).Read(p)
	}
	
	if err != nil {
		if err == io.EOF {
			if closeErr := s.Close(); closeErr != nil {
				s.conn.logger.Error("Error closing stream on EOF", zap.Error(closeErr))
			}
		}
		return n, err
	}

	// Update connection metrics
	s.conn.mu.Lock()
	s.conn.bytesReceived += int64(n)
	s.conn.lastActivity = time.Now()
	s.conn.mu.Unlock()

	return n, nil
}

// Write writes data to the stream
func (s *Stream) Write(p []byte) (n int, err error) {
	s.mu.Lock()
	s.lastActivity = time.Now()
	s.mu.Unlock()

	// Используем Enhanced соединение, если доступно
	if s.conn.enhancedMode && s.conn.enhancedConn != nil {
		n, err = s.conn.enhancedConn.Write(p)
	} else {
		n, err = (*s.stream).Write(p)
	}
	
	if err != nil {
		return n, err
	}

	// Update connection metrics
	s.conn.mu.Lock()
	s.conn.bytesSent += int64(n)
	s.conn.lastActivity = time.Now()
	s.conn.mu.Unlock()

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
		return fmt.Errorf("Enhanced connection already exists")
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

	// Создаем Enhanced соединение
	enhancedConn, err := DialWithTenant("udp", qc.conn.RemoteAddr().String(), tenantID, config)
	if err != nil {
		return fmt.Errorf("failed to create Enhanced connection: %v", err)
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