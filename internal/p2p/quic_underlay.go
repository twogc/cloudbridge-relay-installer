package p2p

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"net"
	"sync"
	"time"

	"github.com/quic-go/quic-go"
	"go.uber.org/zap"

	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
)

// QUICUnderlayManager управляет QUIC underlay соединениями
type QUICUnderlayManager struct {
	listener    *quic.Listener
	connections map[string]*QUICUnderlayConnection
	streamPool  *StreamPool
	logger      *logger.Logger
	ctx         context.Context
	cancel      context.CancelFunc
	mutex       sync.RWMutex
	tlsConfig   *tls.Config
	quicConfig  *quic.Config
}

// QUICUnderlayConnection представляет QUIC соединение с пулом потоков
type QUICUnderlayConnection struct {
	conn         quic.Conn
	streams      map[quic.StreamID]*QUICStream
	streamMutex  sync.RWMutex
	connectionID string
	tunnelID     string
	peerID       string
	tenantID     string
	isActive     bool
	lastActivity time.Time
	bytesIn      int64
	bytesOut     int64
	createdAt    time.Time
	logger       *logger.Logger
}

// QUICStream представляет QUIC поток с метаданными
type QUICStream struct {
	stream    quic.Stream
	id        quic.StreamID
	tunnelID  string
	isActive  bool
	lastUsed  time.Time
	bytesIn   int64
	bytesOut  int64
	createdAt time.Time
}

// StreamPool управляет пулом QUIC потоков
type StreamPool struct {
	streams    map[string]*QUICStream
	maxStreams int
	logger     *logger.Logger
	mutex      sync.RWMutex
}

// NewQUICUnderlayManager создает новый QUIC Underlay Manager
func NewQUICUnderlayManager(logger *logger.Logger) *QUICUnderlayManager {
	ctx, cancel := context.WithCancel(context.Background())

	// Генерируем TLS конфигурацию
	tlsConfig := generateTLSConfig()

	// Настраиваем QUIC конфигурацию
	quicConfig := &quic.Config{
		MaxIdleTimeout:          30 * time.Second,
		MaxIncomingStreams:      1000,
		MaxIncomingUniStreams:   1000,
		KeepAlivePeriod:         10 * time.Second,
		DisablePathMTUDiscovery: true,
	}

	return &QUICUnderlayManager{
		connections: make(map[string]*QUICUnderlayConnection),
		streamPool:  NewStreamPool(logger, 1000),
		logger:      logger,
		ctx:         ctx,
		cancel:      cancel,
		tlsConfig:   tlsConfig,
		quicConfig:  quicConfig,
	}
}

// Start запускает QUIC Underlay Manager
func (qum *QUICUnderlayManager) Start(addr string) error {
	qum.logger.Info("Starting QUIC Underlay Manager", zap.String("addr", addr))

	// Создаем listener
	listener, err := quic.ListenAddr(addr, qum.tlsConfig, qum.quicConfig)
	if err != nil {
		return fmt.Errorf("failed to create QUIC listener: %v", err)
	}

	qum.listener = listener

	// Запускаем accept loop
	go qum.acceptLoop()

	// Запускаем cleanup loop
	go qum.cleanupLoop()

	qum.logger.Info("QUIC Underlay Manager started successfully")
	return nil
}

// Stop останавливает QUIC Underlay Manager
func (qum *QUICUnderlayManager) Stop() error {
	qum.logger.Info("Stopping QUIC Underlay Manager")
	qum.cancel()

	if qum.listener != nil {
		qum.listener.Close()
	}

	qum.mutex.Lock()
	defer qum.mutex.Unlock()

	// Закрываем все соединения
	for _, conn := range qum.connections {
		conn.Close()
	}

	return nil
}

// acceptLoop принимает новые QUIC соединения
func (qum *QUICUnderlayManager) acceptLoop() {
	for {
		select {
		case <-qum.ctx.Done():
			return
		default:
			conn, err := qum.listener.Accept(qum.ctx)
			if err != nil {
				if qum.ctx.Err() != nil {
					return
				}
				qum.logger.Error("Failed to accept QUIC connection", zap.Error(err))
				continue
			}

			go qum.handleConnection(*conn)
		}
	}
}

// handleConnection обрабатывает новое QUIC соединение
func (qum *QUICUnderlayManager) handleConnection(conn quic.Conn) {
	connectionID := fmt.Sprintf("%s-%d", conn.RemoteAddr().String(), time.Now().UnixNano())

	qum.logger.Info("New QUIC connection",
		zap.String("connection_id", connectionID),
		zap.String("remote_addr", conn.RemoteAddr().String()))

	// Создаем QUIC соединение
	quicConn := &QUICUnderlayConnection{
		conn:         conn,
		streams:      make(map[quic.StreamID]*QUICStream),
		connectionID: connectionID,
		isActive:     true,
		lastActivity: time.Now(),
		createdAt:    time.Now(),
		logger:       qum.logger,
	}

	// Аутентификация через первый поток
	peerID, tenantID, err := qum.authenticateConnection(quicConn)
	if err != nil {
		qum.logger.Error("Failed to authenticate QUIC connection",
			zap.String("connection_id", connectionID),
			zap.Error(err))
		conn.CloseWithError(0, "authentication failed")
		return
	}

	quicConn.peerID = peerID
	quicConn.tenantID = tenantID

	// Регистрируем соединение
	qum.mutex.Lock()
	qum.connections[connectionID] = quicConn
	qum.mutex.Unlock()

	// Обрабатываем потоки
	qum.handleStreams(quicConn)
}

// handleStreams обрабатывает потоки QUIC соединения
func (qum *QUICUnderlayManager) handleStreams(conn *QUICUnderlayConnection) {
	for {
		select {
		case <-qum.ctx.Done():
			return
		default:
			stream, err := conn.conn.AcceptStream(qum.ctx)
			if err != nil {
				if qum.ctx.Err() != nil {
					return
				}
				conn.logger.Error("Failed to accept stream",
					zap.String("connection_id", conn.connectionID),
					zap.Error(err))
				continue
			}

			go qum.handleStream(conn, *stream)
		}
	}
}

// handleStream обрабатывает QUIC поток
func (qum *QUICUnderlayManager) handleStream(conn *QUICUnderlayConnection, stream quic.Stream) {
	streamID := stream.StreamID()

	if streamID < 0 {
		conn.logger.Error("Invalid stream ID", zap.Int64("stream_id", int64(streamID)))
		return
	}
	// Safe conversion for logging
	var streamIDUint uint64
	if streamID >= 0 {
		streamIDUint = uint64(streamID) // #nosec G115 -- streamID is validated to be non-negative
	}
	conn.logger.Debug("New QUIC stream",
		zap.String("connection_id", conn.connectionID),
		zap.Uint64("stream_id", streamIDUint))

	// Создаем QUIC поток
	quicStream := &QUICStream{
		stream:    stream,
		id:        streamID,
		tunnelID:  conn.connectionID,
		isActive:  true,
		lastUsed:  time.Now(),
		createdAt: time.Now(),
	}

	// Регистрируем поток
	conn.streamMutex.Lock()
	conn.streams[streamID] = quicStream
	conn.streamMutex.Unlock()

	// Обрабатываем данные потока
	qum.processStreamData(conn, quicStream)
}

// processStreamData обрабатывает данные потока
func (qum *QUICUnderlayManager) processStreamData(conn *QUICUnderlayConnection, stream *QUICStream) {
	defer func() {
		stream.stream.Close()
		conn.streamMutex.Lock()
		delete(conn.streams, stream.id)
		conn.streamMutex.Unlock()
	}()

	buffer := make([]byte, 4096)
	for {
		select {
		case <-qum.ctx.Done():
			return
		default:
			n, err := stream.stream.Read(buffer)
			if err != nil {
				if qum.ctx.Err() != nil {
					return
				}
				if stream.id >= 0 {
					conn.logger.Debug("Stream closed",
						zap.String("connection_id", conn.connectionID),
						zap.Uint64("stream_id", uint64(stream.id))) // #nosec G115 -- stream.id is validated to be non-negative
				}
				return
			}

			// Обновляем статистику
			if n > 0 {
				stream.bytesIn += int64(n)
				stream.lastUsed = time.Now()
				conn.bytesIn += int64(n)
				conn.lastActivity = time.Now()
			}

			// Обрабатываем данные (здесь должна быть логика маршрутизации)
			qum.routeData(conn, stream, buffer[:n])
		}
	}
}

// routeData маршрутизирует данные (заглушка)
func (qum *QUICUnderlayManager) routeData(conn *QUICUnderlayConnection, stream *QUICStream, data []byte) {
	// Здесь должна быть логика маршрутизации данных
	// Пока просто логируем
	if stream.id >= 0 {
		conn.logger.Debug("Routing data",
			zap.String("connection_id", conn.connectionID),
			zap.Uint64("stream_id", uint64(stream.id)), // #nosec G115 -- stream.id is validated to be non-negative
			zap.Int("bytes", len(data)))
	}
}

// authenticateConnection аутентифицирует QUIC соединение
func (qum *QUICUnderlayManager) authenticateConnection(conn *QUICUnderlayConnection) (string, string, error) {
	// Принимаем первый поток для аутентификации
	stream, err := conn.conn.AcceptStream(qum.ctx)
	if err != nil {
		return "", "", err
	}
	defer stream.Close()

	// Читаем данные аутентификации
	buffer := make([]byte, 1024)
	n, err := stream.Read(buffer)
	if err != nil {
		return "", "", err
	}

	// Парсим данные аутентификации (упрощенная версия)
	_ = string(buffer[:n])
	// В реальности нужно парсить JWT токен

	// Извлекаем peerID и tenantID (заглушка)
	peerID := "peer-" + conn.connectionID
	tenantID := "tenant-" + conn.connectionID

	conn.logger.Info("QUIC connection authenticated",
		zap.String("connection_id", conn.connectionID),
		zap.String("peer_id", peerID),
		zap.String("tenant_id", tenantID))

	return peerID, tenantID, nil
}

// cleanupLoop очищает неактивные соединения и потоки
func (qum *QUICUnderlayManager) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-qum.ctx.Done():
			return
		case <-ticker.C:
			qum.cleanupInactiveConnections()
		}
	}
}

// cleanupInactiveConnections удаляет неактивные соединения
func (qum *QUICUnderlayManager) cleanupInactiveConnections() {
	qum.mutex.Lock()
	defer qum.mutex.Unlock()

	cutoff := time.Now().Add(-5 * time.Minute)
	var toRemove []string

	for id, conn := range qum.connections {
		if !conn.isActive || conn.lastActivity.Before(cutoff) {
			toRemove = append(toRemove, id)
		}
	}

	for _, id := range toRemove {
		qum.connections[id].Close()
		delete(qum.connections, id)

		qum.logger.Info("Removed inactive QUIC connection",
			zap.String("connection_id", id))
	}
}

// Close закрывает QUIC соединение
func (conn *QUICUnderlayConnection) Close() error {
	conn.logger.Info("Closing QUIC connection",
		zap.String("connection_id", conn.connectionID))

	conn.isActive = false

	// Закрываем все потоки
	conn.streamMutex.Lock()
	for _, stream := range conn.streams {
		stream.stream.Close()
	}
	conn.streamMutex.Unlock()

	// Закрываем соединение
	return conn.conn.CloseWithError(0, "connection closed")
}

// GetStats возвращает статистику соединения
func (conn *QUICUnderlayConnection) GetStats() map[string]interface{} {
	conn.streamMutex.RLock()
	streamCount := len(conn.streams)
	conn.streamMutex.RUnlock()

	return map[string]interface{}{
		"connection_id": conn.connectionID,
		"peer_id":       conn.peerID,
		"tenant_id":     conn.tenantID,
		"is_active":     conn.isActive,
		"stream_count":  streamCount,
		"bytes_in":      conn.bytesIn,
		"bytes_out":     conn.bytesOut,
		"last_activity": conn.lastActivity,
		"created_at":    conn.createdAt,
	}
}

// NewStreamPool создает новый пул потоков
func NewStreamPool(logger *logger.Logger, maxStreams int) *StreamPool {
	return &StreamPool{
		streams:    make(map[string]*QUICStream),
		maxStreams: maxStreams,
		logger:     logger,
	}
}

// generateTLSConfig генерирует TLS конфигурацию для QUIC
func generateTLSConfig() *tls.Config {
	// Генерируем приватный ключ
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(fmt.Sprintf("Failed to generate private key: %v", err))
	}

	// Создаем сертификат
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:  []string{"CloudBridge"},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{"San Francisco"},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		DNSNames:    []string{"localhost", "b1.2gc.space"},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		panic(fmt.Sprintf("Failed to create certificate: %v", err))
	}

	return &tls.Config{
		Certificates: []tls.Certificate{
			{
				Certificate: [][]byte{certDER},
				PrivateKey:  priv,
			},
		},
		NextProtos: []string{"h3", "h3-29", "h3-28", "h3-27"},
		MinVersion: tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
	}
}
