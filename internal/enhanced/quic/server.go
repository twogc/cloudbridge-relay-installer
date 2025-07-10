package quic

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/quic-go/quic-go"
	"go.uber.org/zap"

	"github.com/twogc/cloudbridge-relay-installer/internal/enhanced/metrics"
)

// QUICServer представляет QUIC сервер для CloudBridge Relay
type QUICServer struct {
	listener       *quic.Listener
	logger         *zap.Logger
	metrics        *metrics.EnhancedMetrics
	config         *QUICServerConfig
	connections    map[uuid.UUID]*QUICConnection
	mu             sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
	startTime      time.Time
	activeConns    int32
	totalConns     int64
}

// QUICServerConfig конфигурация для QUIC сервера
type QUICServerConfig struct {
	Address           string
	Port              int
	MaxConnections    int
	ConnectionTimeout time.Duration
	EnableTLS         bool
	CertFile          string
	KeyFile           string
	EnableMetrics     bool
	EnableTracing     bool
}

// DefaultQUICServerConfig возвращает конфигурацию по умолчанию
func DefaultQUICServerConfig() *QUICServerConfig {
	return &QUICServerConfig{
		Address:           "0.0.0.0",
		Port:              8082,
		MaxConnections:    1000,
		ConnectionTimeout: 60 * time.Second,
		EnableTLS:         true,
		EnableMetrics:     true,
		EnableTracing:     false,
	}
}

// NewQUICServer создает новый QUIC сервер
func NewQUICServer(
	config *QUICServerConfig,
	logger *zap.Logger,
	metrics *metrics.EnhancedMetrics,
) *QUICServer {
	if config == nil {
		config = DefaultQUICServerConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &QUICServer{
		logger:      logger,
		metrics:     metrics,
		config:      config,
		connections: make(map[uuid.UUID]*QUICConnection),
		ctx:         ctx,
		cancel:      cancel,
		startTime:   time.Now(),
	}
}

// Start запускает QUIC сервер
func (qs *QUICServer) Start() error {
	// Создаем TLS конфигурацию
	tlsConfig, err := qs.createTLSConfig()
	if err != nil {
		return fmt.Errorf("failed to create TLS config: %w", err)
	}

	// Создаем QUIC конфигурацию
	quicConfig := &quic.Config{
		MaxIdleTimeout:  qs.config.ConnectionTimeout,
		KeepAlivePeriod: 30 * time.Second,
	}

	// Создаем адрес для прослушивания
	addr := fmt.Sprintf("%s:%d", qs.config.Address, qs.config.Port)

	// Создаем listener
	listener, err := quic.ListenAddr(addr, tlsConfig, quicConfig)
	if err != nil {
		return fmt.Errorf("failed to create QUIC listener: %w", err)
	}

	qs.listener = listener
	qs.logger.Info("QUIC server started",
		zap.String("address", addr),
		zap.Int("max_connections", qs.config.MaxConnections),
		zap.Bool("tls_enabled", qs.config.EnableTLS),
	)

	// Запускаем обработку соединений
	go qs.acceptConnections()

	// Запускаем сборщик метрик
	if qs.config.EnableMetrics {
		go qs.metricsCollector()
	}

	return nil
}

// acceptConnections принимает новые соединения
func (qs *QUICServer) acceptConnections() {
	for {
		select {
		case <-qs.ctx.Done():
			qs.logger.Info("QUIC server context cancelled")
			return

		default:
			// Принимаем новое соединение с таймаутом
			acceptCtx, cancel := context.WithTimeout(qs.ctx, qs.config.ConnectionTimeout)
			conn, err := qs.listener.Accept(acceptCtx)
			cancel()

			if err != nil {
				if qs.ctx.Err() != nil {
					return
				}
				qs.logger.Debug("Failed to accept QUIC connection", zap.Error(err))
				continue
			}

			// Проверяем лимит соединений
			if int(qs.getActiveConnections()) >= qs.config.MaxConnections {
				qs.logger.Warn("Max connections reached, rejecting new connection",
					zap.Int32("active", qs.getActiveConnections()),
					zap.Int("max", qs.config.MaxConnections),
				)
				conn.CloseWithError(0, "max connections reached")
				continue
			}

			// Обрабатываем соединение в отдельной горутине
			go qs.handleConnection(conn)
		}
	}
}

// handleConnection обрабатывает новое соединение
func (qs *QUICServer) handleConnection(conn *quic.Conn) {
	connectionID := uuid.New()
	startTime := time.Now()

	// Создаем QUIC соединение
	quicConn := NewQUICConnection(
		conn,
		qs.logger.With(zap.String("connection_id", connectionID.String())),
		qs.metrics,
		&QUICConfig{
			MaxStreams:        100,
			KeepAliveInterval: 30 * time.Second,
			IdleTimeout:       60 * time.Second,
			MaxIdleTime:       300 * time.Second,
			EnableMetrics:     qs.config.EnableMetrics,
			EnableTracing:     qs.config.EnableTracing,
		},
	)

	// Добавляем соединение в список
	qs.mu.Lock()
	qs.connections[connectionID] = quicConn
	qs.activeConns++
	qs.totalConns++
	qs.mu.Unlock()

	// Обновляем метрики
	if qs.metrics != nil {
		qs.metrics.IncProtocolSuccess("quic", "connection_established")
	}

	qs.logger.Info("New QUIC connection established",
		zap.String("connection_id", connectionID.String()),
		zap.String("remote_addr", conn.RemoteAddr().String()),
		zap.Int32("active_connections", qs.getActiveConnections()),
	)

	defer func() {
		// Удаляем соединение из списка
		qs.mu.Lock()
		delete(qs.connections, connectionID)
		qs.activeConns--
		qs.mu.Unlock()

		// Закрываем соединение
		quicConn.Close()
		duration := time.Since(startTime)

		// Обновляем метрики
		if qs.metrics != nil {
			qs.metrics.ObserveProtocolLatency("quic", "connection_lifetime", duration)
		}

		qs.logger.Info("QUIC connection closed",
			zap.String("connection_id", connectionID.String()),
			zap.Duration("duration", duration),
			zap.Int32("active_connections", qs.getActiveConnections()),
		)
	}()

	// Запускаем обработку соединения
	if err := quicConn.Serve(qs.ctx); err != nil {
		qs.logger.Error("QUIC connection serve error",
			zap.Error(err),
			zap.String("connection_id", connectionID.String()),
		)

		// Обновляем метрики ошибок
		if qs.metrics != nil {
			qs.metrics.IncProtocolErrors("quic", "serve_error")
		}
	}
}

// createTLSConfig создает TLS конфигурацию
func (qs *QUICServer) createTLSConfig() (*tls.Config, error) {
	if !qs.config.EnableTLS {
		// Для тестирования без TLS
		return &tls.Config{
			InsecureSkipVerify: true,
		}, nil
	}

	if qs.config.CertFile != "" && qs.config.KeyFile != "" {
		// Загружаем сертификат из файлов
		cert, err := tls.LoadX509KeyPair(qs.config.CertFile, qs.config.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load certificate: %w", err)
		}

		return &tls.Config{
			Certificates: []tls.Certificate{cert},
			NextProtos:   []string{"cloudbridge-quic"},
		}, nil
	}

	// Генерируем самоподписанный сертификат для тестирования
	return qs.generateSelfSignedCert()
}

// generateSelfSignedCert генерирует самоподписанный сертификат
func (qs *QUICServer) generateSelfSignedCert() (*tls.Config, error) {
	// Генерируем приватный ключ
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	// Создаем сертификат
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"CloudBridge Relay"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost", "cloudbridge.local"},
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
	}

	// Подписываем сертификат
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	// Создаем PEM блоки
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: derBytes,
	})

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	// Создаем сертификат
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to create key pair: %w", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{"cloudbridge-quic"},
	}, nil
}

// metricsCollector собирает метрики сервера
func (qs *QUICServer) metricsCollector() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-qs.ctx.Done():
			return
		case <-ticker.C:
			qs.updateMetrics()
		}
	}
}

// updateMetrics обновляет метрики сервера
func (qs *QUICServer) updateMetrics() {
	if qs.metrics == nil {
		return
	}

	activeConns := qs.getActiveConnections()
	totalConns := qs.getTotalConnections()
	uptime := time.Since(qs.startTime)

	// Обновляем метрики
	qs.metrics.SetConnectionPoolActive("quic", "quic", int(activeConns))
	qs.metrics.ObserveProtocolLatency("quic", "server_uptime", uptime)

	// Логируем статистику
	qs.logger.Debug("QUIC server metrics",
		zap.Int32("active_connections", activeConns),
		zap.Int64("total_connections", totalConns),
		zap.Duration("uptime", uptime),
	)
}

// Stop останавливает QUIC сервер
func (qs *QUICServer) Stop() error {
	qs.logger.Info("Stopping QUIC server")

	// Отменяем контекст
	qs.cancel()

	// Закрываем все соединения
	qs.mu.Lock()
	for _, conn := range qs.connections {
		conn.Close()
	}
	qs.connections = make(map[uuid.UUID]*QUICConnection)
	qs.mu.Unlock()

	// Закрываем listener
	if qs.listener != nil {
		return qs.listener.Close()
	}

	return nil
}

// GetStats возвращает статистику сервера
func (qs *QUICServer) GetStats() map[string]interface{} {
	qs.mu.RLock()
	defer qs.mu.RUnlock()

	return map[string]interface{}{
		"address":           fmt.Sprintf("%s:%d", qs.config.Address, qs.config.Port),
		"active_connections": qs.activeConns,
		"total_connections":  qs.totalConns,
		"max_connections":    qs.config.MaxConnections,
		"uptime":            time.Since(qs.startTime).String(),
		"tls_enabled":       qs.config.EnableTLS,
		"enable_metrics":    qs.config.EnableMetrics,
		"enable_tracing":    qs.config.EnableTracing,
	}
}

// GetConnections возвращает список активных соединений
func (qs *QUICServer) GetConnections() map[uuid.UUID]*QUICConnection {
	qs.mu.RLock()
	defer qs.mu.RUnlock()

	// Создаем копию для безопасности
	connections := make(map[uuid.UUID]*QUICConnection)
	for id, conn := range qs.connections {
		connections[id] = conn
	}

	return connections
}

// getActiveConnections возвращает количество активных соединений
func (qs *QUICServer) getActiveConnections() int32 {
	qs.mu.RLock()
	defer qs.mu.RUnlock()
	return qs.activeConns
}

// getTotalConnections возвращает общее количество соединений
func (qs *QUICServer) getTotalConnections() int64 {
	qs.mu.RLock()
	defer qs.mu.RUnlock()
	return qs.totalConns
}

// IsRunning проверяет, запущен ли сервер
func (qs *QUICServer) IsRunning() bool {
	return qs.listener != nil && qs.ctx.Err() == nil
}

// GetAddress возвращает адрес сервера
func (qs *QUICServer) GetAddress() string {
	return fmt.Sprintf("%s:%d", qs.config.Address, qs.config.Port)
} 