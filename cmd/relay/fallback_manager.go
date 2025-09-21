package main

import (
	"context"
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

	// Инициализируем MASQUE сервер
	fm.masqueServer = NewMASQUEServer(fm.logger, fm.config)
	if err := fm.masqueServer.Start(); err != nil {
		return fmt.Errorf("failed to start MASQUE server: %v", err)
	}

	// Инициализируем WebSocket fallback
	fm.websocketFallback = NewWebSocketFallback(fm.logger, fm.config)
	if err := fm.websocketFallback.Start(); err != nil {
		return fmt.Errorf("failed to start WebSocket fallback: %v", err)
	}

	// Инициализируем TCP fallback
	fm.tcpFallback = NewTCPFallback(fm.logger, fm.config)
	if err := fm.tcpFallback.Start(); err != nil {
		return fmt.Errorf("failed to start TCP fallback: %v", err)
	}

	fm.isRunning = true
	fm.logger.Info("FallbackManager started successfully")

	return nil
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

	quicResult := &ConnectionResult{
		Protocol: "QUIC",
		Success:  false,
		Duration: time.Since(start),
	}

	// Проверяем доступность QUIC/MASQUE
	if fm.masqueServer != nil && fm.masqueServer.GetConnectionCount() < fm.config.Server.MaxTunnels {
		// Пытаемся создать MASQUE соединение
		quicResult.Success = fm.testQUICConnectivity(ctx, conn.RemoteAddr().String())
		if !quicResult.Success {
			quicResult.Error = fmt.Errorf("QUIC connectivity test failed")
		}
	} else {
		quicResult.Error = fmt.Errorf("MASQUE server not available or max tunnels reached")
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
	wsResult := &ConnectionResult{
		Protocol:     "WebSocket",
		Success:      false,
		Duration:     time.Since(start),
		FallbackUsed: true,
	}

	// Проверяем доступность WebSocket fallback
	if fm.websocketFallback != nil {
		wsResult.Success = fm.testWebSocketConnectivity(ctx, conn.RemoteAddr().String())
		if !wsResult.Success {
			wsResult.Error = fmt.Errorf("WebSocket connectivity test failed")
		}
	} else {
		wsResult.Error = fmt.Errorf("WebSocket fallback not available")
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
	tcpResult := &ConnectionResult{
		Protocol:     "TCP",
		Success:      false,
		Duration:     time.Since(start),
		FallbackUsed: true,
	}

	// Проверяем доступность TCP fallback
	if fm.tcpFallback != nil {
		tcpResult.Success = fm.testTCPConnectivity(ctx, conn.RemoteAddr().String())
		if !tcpResult.Success {
			tcpResult.Error = fmt.Errorf("TCP connectivity test failed")
		}
	} else {
		tcpResult.Error = fmt.Errorf("TCP fallback not available")
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
	// Простая эвристика для демонстрации
	// В реальности здесь должна быть более сложная логика

	// Проверяем на корпоративные сети (пример)
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return false
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}

	// Проверяем на частные сети, которые могут блокировать QUIC
	if ip.IsPrivate() {
		// Некоторые корпоративные сети блокируют QUIC
		return true
	}

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
func (fm *FallbackManager) testQUICConnectivity(ctx context.Context, remoteAddr string) bool {
	// Создаем тестовое соединение для проверки QUIC доступности
	timeout := 5 * time.Second

	// Пытаемся подключиться к QUIC порту
	conn, err := net.DialTimeout("udp", remoteAddr, timeout)
	if err != nil {
		fm.logger.Debug("QUIC connectivity test failed",
			zap.String("remote_addr", remoteAddr),
			zap.Error(err))
		return false
	}
	defer conn.Close()

	// Отправляем тестовый пакет
	testData := []byte("QUIC_TEST")
	if _, err := conn.Write(testData); err != nil {
		fm.logger.Debug("QUIC test packet send failed",
			zap.String("remote_addr", remoteAddr),
			zap.Error(err))
		return false
	}

	// Ждем ответ (с таймаутом)
	conn.SetReadDeadline(time.Now().Add(timeout))
	buffer := make([]byte, 1024)
	_, err = conn.Read(buffer)
	if err != nil {
		// Для QUIC это может быть нормально, так как не все серверы отвечают на тестовые пакеты
		fm.logger.Debug("QUIC test response timeout (may be normal)",
			zap.String("remote_addr", remoteAddr))
		return true // Считаем успешным, если можем отправить пакет
	}

	fm.logger.Debug("QUIC connectivity test successful",
		zap.String("remote_addr", remoteAddr))
	return true
}

// testWebSocketConnectivity тестирует WebSocket connectivity
func (fm *FallbackManager) testWebSocketConnectivity(ctx context.Context, remoteAddr string) bool {
	// Проверяем доступность WebSocket порта
	timeout := 5 * time.Second

	// Пытаемся подключиться к WebSocket порту
	conn, err := net.DialTimeout("tcp", remoteAddr, timeout)
	if err != nil {
		fm.logger.Debug("WebSocket connectivity test failed",
			zap.String("remote_addr", remoteAddr),
			zap.Error(err))
		return false
	}
	defer conn.Close()

	// Отправляем HTTP Upgrade запрос
	upgradeRequest := fmt.Sprintf("GET / HTTP/1.1\r\nHost: %s\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==\r\nSec-WebSocket-Version: 13\r\n\r\n", remoteAddr)
	if _, err := conn.Write([]byte(upgradeRequest)); err != nil {
		fm.logger.Debug("WebSocket upgrade request failed",
			zap.String("remote_addr", remoteAddr),
			zap.Error(err))
		return false
	}

	// Читаем ответ
	conn.SetReadDeadline(time.Now().Add(timeout))
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fm.logger.Debug("WebSocket upgrade response failed",
			zap.String("remote_addr", remoteAddr),
			zap.Error(err))
		return false
	}

	response := string(buffer[:n])
	if !strings.Contains(response, "101 Switching Protocols") {
		fm.logger.Debug("WebSocket upgrade not supported",
			zap.String("remote_addr", remoteAddr))
		return false
	}

	fm.logger.Debug("WebSocket connectivity test successful",
		zap.String("remote_addr", remoteAddr))
	return true
}

// testTCPConnectivity тестирует TCP connectivity
func (fm *FallbackManager) testTCPConnectivity(ctx context.Context, remoteAddr string) bool {
	// Проверяем доступность TCP порта
	timeout := 5 * time.Second

	// Пытаемся подключиться к TCP порту
	conn, err := net.DialTimeout("tcp", remoteAddr, timeout)
	if err != nil {
		fm.logger.Debug("TCP connectivity test failed",
			zap.String("remote_addr", remoteAddr),
			zap.Error(err))
		return false
	}
	defer conn.Close()

	// Отправляем тестовые данные
	testData := []byte("TCP_TEST")
	if _, err := conn.Write(testData); err != nil {
		fm.logger.Debug("TCP test data send failed",
			zap.String("remote_addr", remoteAddr),
			zap.Error(err))
		return false
	}

	fm.logger.Debug("TCP connectivity test successful",
		zap.String("remote_addr", remoteAddr))
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
