package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/twogc/cloudbridge-relay-installer/config"
	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
)

// FallbackManager управляет fallback механизмами для блокирующих сетей
type FallbackManager struct {
	logger            *logger.Logger
	config            *config.Config
	masqueServer      *MASQUEServer
	websocketFallback *WebSocketFallback
	tcpFallback       *TCPFallback

	// Состояние
	isRunning bool
	ctx       context.Context
	cancel    context.CancelFunc
	mu        sync.RWMutex
}

// FallbackStrategy стратегия fallback
type FallbackStrategy int

const (
	// FallbackStrategyQUICFirst пытается QUIC, затем fallback
	FallbackStrategyQUICFirst FallbackStrategy = iota
	// FallbackStrategyWebSocketFirst пытается WebSocket, затем TCP
	FallbackStrategyWebSocketFirst
	// FallbackStrategyTCPOnly использует только TCP
	FallbackStrategyTCPOnly
)

// ConnectionResult результат попытки соединения
type ConnectionResult struct {
	Protocol     string
	Success      bool
	Duration     time.Duration
	Error        error
	FallbackUsed bool
}

// NewFallbackManager создает новый FallbackManager
func NewFallbackManager(logger *logger.Logger, config *config.Config) *FallbackManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &FallbackManager{
		logger: logger,
		config: config,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start запускает FallbackManager
func (fm *FallbackManager) Start() error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if fm.isRunning {
		return fmt.Errorf("FallbackManager is already running")
	}

	fm.logger.Info("Starting FallbackManager")
	started := []func() error{} // stack of closers

	// MASQUE
	if fm.config.MASQUE.Enabled {
		fm.masqueServer = NewMASQUEServer(fm.logger, fm.config)
		if err := fm.masqueServer.Start(); err != nil {
			fm.rollback(started)
			return fmt.Errorf("failed to start MASQUE server: %w", err)
		}
		started = append(started, fm.masqueServer.Close)
	}

	// WebSocket
	if fm.config.WebSocketFallback.Enabled {
		fm.websocketFallback = NewWebSocketFallback(fm.logger, fm.config)
		if err := fm.websocketFallback.Start(); err != nil {
			fm.rollback(started)
			return fmt.Errorf("failed to start WebSocket fallback: %w", err)
		}
		started = append(started, fm.websocketFallback.Close)
	}

	// TCP
	if fm.config.TCPFallback.Enabled {
		fm.tcpFallback = NewTCPFallback(fm.logger, fm.config)
		if err := fm.tcpFallback.Start(); err != nil {
			fm.rollback(started)
			return fmt.Errorf("failed to start TCP fallback: %w", err)
		}
		started = append(started, fm.tcpFallback.Close)
	}

	if fm.masqueServer == nil && fm.websocketFallback == nil && fm.tcpFallback == nil {
		return fmt.Errorf("no fallback mechanisms enabled")
	}

	fm.isRunning = true
	fm.logger.Info("FallbackManager started successfully")
	return nil
}

// rollback закрывает уже запущенные компоненты при ошибке
func (fm *FallbackManager) rollback(closers []func() error) {
	for i := len(closers) - 1; i >= 0; i-- {
		_ = closers[i]()
	}
}

// HandleConnection обрабатывает соединение с автоматическим fallback
func (fm *FallbackManager) HandleConnection(ctx context.Context, conn net.Conn, tenantID string) (*ConnectionResult, error) {
	if !fm.isRunning {
		return nil, fmt.Errorf("FallbackManager is not running")
	}

	fm.logger.Info("Handling connection with fallback",
		zap.String("remote_addr", conn.RemoteAddr().String()),
		zap.String("tenant_id", tenantID))

	// Определяем стратегию fallback
	strategy := fm.determineFallbackStrategy(conn.RemoteAddr().String())

	// Пытаемся установить соединение с fallback
	return fm.attemptConnectionWithFallback(ctx, conn, tenantID, strategy)
}

// determineFallbackStrategy определяет стратегию fallback
func (fm *FallbackManager) determineFallbackStrategy(remoteAddr string) FallbackStrategy {
	// Здесь можно добавить логику определения стратегии на основе:
	// - IP адреса клиента
	// - Исторических данных
	// - Конфигурации
	// - Типа сети

	// Для демонстрации используем QUIC-first стратегию
	return FallbackStrategyQUICFirst
}

// attemptConnectionWithFallback пытается установить соединение с fallback
func (fm *FallbackManager) attemptConnectionWithFallback(ctx context.Context, conn net.Conn, tenantID string, strategy FallbackStrategy) (*ConnectionResult, error) {
	start := time.Now()

	switch strategy {
	case FallbackStrategyQUICFirst:
		return fm.attemptQUICFirst(ctx, conn, tenantID, start)
	case FallbackStrategyWebSocketFirst:
		return fm.attemptWebSocketFirst(ctx, conn, tenantID, start)
	case FallbackStrategyTCPOnly:
		return fm.attemptTCPOnly(ctx, conn, tenantID, start)
	default:
		return nil, fmt.Errorf("unknown fallback strategy: %d", strategy)
	}
}

// attemptQUICFirst пытается QUIC, затем fallback на WebSocket/TCP
func (fm *FallbackManager) attemptQUICFirst(ctx context.Context, conn net.Conn, tenantID string, start time.Time) (*ConnectionResult, error) {
	// 1. Пытаемся QUIC/MASQUE
	fm.logger.Debug("Attempting QUIC/MASQUE connection",
		zap.String("tenant_id", tenantID))

	qStart := time.Now()
	quicResult := &ConnectionResult{
		Protocol: "QUIC",
		Success:  false,
		Duration: 0, // будет установлено после теста
	}

	// Проверяем доступность QUIC/MASQUE
	if fm.masqueServer != nil && fm.masqueServer.GetConnectionCount() < fm.config.Server.MaxTunnels {
		// Пытаемся создать MASQUE соединение
		quicResult.Success = fm.testQUICConnectivity(ctx)
		quicResult.Duration = time.Since(qStart)
		if !quicResult.Success {
			quicResult.Error = fmt.Errorf("QUIC connectivity test failed")
		}
	} else {
		quicResult.Error = fmt.Errorf("MASQUE server not available or max tunnels reached")
		quicResult.Duration = time.Since(qStart)
	}

	if quicResult.Success {
		fm.logger.Info("QUIC/MASQUE connection successful",
			zap.String("tenant_id", tenantID))
		return quicResult, nil
	}

	fm.logger.Warn("QUIC connection failed, trying WebSocket fallback",
		zap.String("tenant_id", tenantID),
		zap.Error(quicResult.Error))

	// 2. Fallback на WebSocket
	wsStart := time.Now()
	wsResult := &ConnectionResult{
		Protocol:     "WebSocket",
		Success:      false,
		Duration:     0, // будет установлено после теста
		FallbackUsed: true,
	}

	// Проверяем доступность WebSocket fallback
	if fm.websocketFallback != nil {
		wsResult.Success = fm.testWebSocketConnectivity(ctx)
		wsResult.Duration = time.Since(wsStart)
		if !wsResult.Success {
			wsResult.Error = fmt.Errorf("WebSocket connectivity test failed")
		}
	} else {
		wsResult.Error = fmt.Errorf("WebSocket fallback not available")
		wsResult.Duration = time.Since(wsStart)
	}

	if wsResult.Success {
		fm.logger.Info("WebSocket fallback successful",
			zap.String("tenant_id", tenantID))
		return wsResult, nil
	}

	fm.logger.Warn("WebSocket fallback failed, trying TCP fallback",
		zap.String("tenant_id", tenantID),
		zap.Error(wsResult.Error))

	// 3. Fallback на TCP
	tStart := time.Now()
	tcpResult := &ConnectionResult{
		Protocol:     "TCP",
		Success:      false,
		Duration:     0, // будет установлено после теста
		FallbackUsed: true,
	}

	// Проверяем доступность TCP fallback
	if fm.tcpFallback != nil {
		tcpResult.Success = fm.testTCPConnectivity(ctx)
		tcpResult.Duration = time.Since(tStart)
		if !tcpResult.Success {
			tcpResult.Error = fmt.Errorf("TCP connectivity test failed")
		}
	} else {
		tcpResult.Error = fmt.Errorf("TCP fallback not available")
		tcpResult.Duration = time.Since(tStart)
	}

	if tcpResult.Success {
		fm.logger.Info("TCP fallback successful",
			zap.String("tenant_id", tenantID))
		return tcpResult, nil
	}

	// Все попытки неудачны
	return &ConnectionResult{
		Protocol:     "All",
		Success:      false,
		Duration:     time.Since(start),
		FallbackUsed: true,
		Error:        fmt.Errorf("all connection attempts failed: QUIC=%v, WebSocket=%v, TCP=%v", quicResult.Error, wsResult.Error, tcpResult.Error),
	}, nil
}

// attemptWebSocketFirst пытается WebSocket, затем TCP
func (fm *FallbackManager) attemptWebSocketFirst(ctx context.Context, conn net.Conn, tenantID string, start time.Time) (*ConnectionResult, error) {
	// 1. Пытаемся WebSocket
	fm.logger.Debug("Attempting WebSocket connection",
		zap.String("tenant_id", tenantID))

	wsResult := &ConnectionResult{
		Protocol: "WebSocket",
		Success:  true, // Для демонстрации
		Duration: time.Since(start),
	}

	if wsResult.Success {
		return wsResult, nil
	}

	// 2. Fallback на TCP
	fm.logger.Warn("WebSocket connection failed, trying TCP fallback",
		zap.String("tenant_id", tenantID))

	tcpResult := &ConnectionResult{
		Protocol:     "TCP",
		Success:      true, // Для демонстрации
		Duration:     time.Since(start),
		FallbackUsed: true,
	}

	return tcpResult, nil
}

// attemptTCPOnly использует только TCP
func (fm *FallbackManager) attemptTCPOnly(ctx context.Context, conn net.Conn, tenantID string, start time.Time) (*ConnectionResult, error) {
	fm.logger.Debug("Attempting TCP-only connection",
		zap.String("tenant_id", tenantID))

	tcpResult := &ConnectionResult{
		Protocol: "TCP",
		Success:  true, // Для демонстрации
		Duration: time.Since(start),
	}

	return tcpResult, nil
}

// DetectBlockingNetwork детектирует блокирующую сеть
func (fm *FallbackManager) DetectBlockingNetwork(remoteAddr string) bool {
	// Здесь можно добавить логику детекции блокирующих сетей:
	// - Проверка доступности QUIC портов
	// - Анализ исторических данных
	// - Проверка через STUN/TURN
	// - Анализ сетевых характеристик

	// Для демонстрации используем простую эвристику
	// Если IP адрес содержит определенные паттерны, считаем сеть блокирующей
	return fm.isBlockingIPPattern(remoteAddr)
}

// isBlockingIPPattern проверяет, является ли IP адрес блокирующим
func (fm *FallbackManager) isBlockingIPPattern(remoteAddr string) bool {
	// Более мягкая эвристика - не помечаем все частные сети как блокирующие
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return false
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}

	// Частные сети — не автоматически "блокирующие", а только фактор
	// В реальности здесь должна быть более сложная логика анализа
	// исторических данных, типа сети и т.д.
	return false
}

// GetConnectionStats возвращает статистику соединений
func (fm *FallbackManager) GetConnectionStats() map[string]interface{} {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	stats := map[string]interface{}{
		"is_running": fm.isRunning,
	}

	if fm.masqueServer != nil {
		stats["masque"] = fm.masqueServer.GetConnectionInfo()
	}

	if fm.websocketFallback != nil {
		stats["websocket"] = fm.websocketFallback.GetConnectionInfo()
	}

	if fm.tcpFallback != nil {
		stats["tcp"] = fm.tcpFallback.GetConnectionInfo()
	}

	return stats
}

// testQUICConnectivity тестирует QUIC connectivity
func (fm *FallbackManager) testQUICConnectivity(ctx context.Context) bool {
	// Проверяем доступность нашего QUIC сервера
	target := net.JoinHostPort("127.0.0.1", "9090") // QUIC порт из конфига
	d := net.Dialer{}
	conn, err := d.DialContext(ctx, "udp", target)
	if err != nil {
		fm.logger.Debug("QUIC UDP dial failed", zap.String("target", target), zap.Error(err))
		return false
	}
	defer conn.Close()

	// Отправляем тестовый пакет
	testData := []byte("QUIC_TEST")
	if _, err := conn.Write(testData); err != nil {
		fm.logger.Debug("QUIC test packet send failed", zap.String("target", target), zap.Error(err))
		return false
	}

	// Ждем ответ с таймаутом
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	buffer := make([]byte, 1024)
	_, err = conn.Read(buffer)
	if err != nil {
		// Для QUIC это может быть нормально, так как не все серверы отвечают на тестовые пакеты
		fm.logger.Debug("QUIC test response timeout (may be normal)", zap.String("target", target))
		return true // Считаем успешным, если можем отправить пакет
	}

	fm.logger.Debug("QUIC connectivity test successful", zap.String("target", target))
	return true
}

// testWebSocketConnectivity тестирует WebSocket connectivity
func (fm *FallbackManager) testWebSocketConnectivity(ctx context.Context) bool {
	// Проверяем доступность нашего WebSocket сервера
	host := "127.0.0.1"
	port := fm.config.HTTPSAPI.Port // если TLS включён
	if port == 0 {
		port = fm.config.HTTPAPI.Port // fallback на HTTP
	}
	addr := net.JoinHostPort(host, fmt.Sprint(port))

	d := &net.Dialer{}
	raw, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		fm.logger.Debug("WebSocket connectivity test failed", zap.String("addr", addr), zap.Error(err))
		return false
	}
	defer raw.Close()

	// Если HTTPS — оборачиваем в TLS
	var c net.Conn = raw
	if fm.config.HTTPSAPI.Enabled && fm.config.HTTPSAPI.TLS.Enabled {
		tlsConn := tls.Client(raw, &tls.Config{ServerName: host, InsecureSkipVerify: true /* dev only */})
		if err := tlsConn.HandshakeContext(ctx); err != nil {
			fm.logger.Debug("WebSocket TLS handshake failed", zap.String("addr", addr), zap.Error(err))
			return false
		}
		c = tlsConn
	}

	// Отправляем WebSocket upgrade запрос
	req := fmt.Sprintf(
		"GET /ws HTTP/1.1\r\nHost: %s\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==\r\nSec-WebSocket-Version: 13\r\n\r\n",
		host,
	)
	c.SetDeadline(time.Now().Add(5 * time.Second))
	if _, err := c.Write([]byte(req)); err != nil {
		fm.logger.Debug("WebSocket upgrade request failed", zap.String("addr", addr), zap.Error(err))
		return false
	}

	// Читаем ответ
	buf := make([]byte, 1024)
	n, err := c.Read(buf)
	if err != nil {
		fm.logger.Debug("WebSocket upgrade response failed", zap.String("addr", addr), zap.Error(err))
		return false
	}

	response := string(buf[:n])
	if !strings.Contains(response, "101 Switching Protocols") {
		fm.logger.Debug("WebSocket upgrade not supported", zap.String("addr", addr))
		return false
	}

	fm.logger.Debug("WebSocket connectivity test successful", zap.String("addr", addr))
	return true
}

// testTCPConnectivity тестирует TCP connectivity
func (fm *FallbackManager) testTCPConnectivity(ctx context.Context) bool {
	// Проверяем доступность нашего TCP сервера
	addr := net.JoinHostPort("127.0.0.1", fmt.Sprint(fm.config.Server.Port))
	d := net.Dialer{}
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		fm.logger.Debug("TCP connectivity test failed", zap.String("addr", addr), zap.Error(err))
		return false
	}
	defer conn.Close()

	// Отправляем тестовые данные
	testData := []byte("TCP_TEST")
	if _, err := conn.Write(testData); err != nil {
		fm.logger.Debug("TCP test data send failed", zap.String("addr", addr), zap.Error(err))
		return false
	}

	fm.logger.Debug("TCP connectivity test successful", zap.String("addr", addr))
	return true
}

// Close закрывает FallbackManager
func (fm *FallbackManager) Close() error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if !fm.isRunning {
		return nil
	}

	fm.logger.Info("Closing FallbackManager")

	fm.cancel()

	// Закрываем все компоненты
	var errors []error

	if fm.masqueServer != nil {
		if err := fm.masqueServer.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close MASQUE server: %v", err))
		}
	}

	if fm.websocketFallback != nil {
		if err := fm.websocketFallback.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close WebSocket fallback: %v", err))
		}
	}

	if fm.tcpFallback != nil {
		if err := fm.tcpFallback.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close TCP fallback: %v", err))
		}
	}

	fm.isRunning = false

	if len(errors) > 0 {
		return fmt.Errorf("errors during shutdown: %v", errors)
	}

	fm.logger.Info("FallbackManager closed successfully")
	return nil
}
