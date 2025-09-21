package main

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
	"github.com/twogc/cloudbridge-relay-installer/internal/nat"
)

// ICEMASQUEBridge реализует мост между ICE и MASQUE для автоматического выбора протокола
type ICEMASQUEBridge struct {
	logger *logger.Logger
	config *BridgeConfig

	// ICE компоненты
	iceAgent   *nat.ICEAgent
	iceSupport *ICESupport

	// MASQUE компоненты
	masqueServer *MASQUEServer

	// Fallback компоненты
	websocketFallback *WebSocketFallback
	tcpFallback       *TCPFallback

	// Управление соединениями
	connections map[string]*BridgeConnection
	connMutex   sync.RWMutex

	// Метрики и статистика
	stats *BridgeStats
}

// BridgeConfig конфигурация для ICE↔MASQUE моста
type BridgeConfig struct {
	// Протоколы
	EnableICE       bool
	EnableMASQUE    bool
	EnableWebSocket bool
	EnableTCP       bool

	// Таймауты
	ICETimeout       time.Duration
	MASQUETimeout    time.Duration
	WebSocketTimeout time.Duration
	TCPTimeout       time.Duration

	// Селекция протокола
	ProtocolSelectionTimeout time.Duration
	FallbackThreshold        int
	EnableProtocolMetrics    bool

	// Приоритеты протоколов
	ProtocolPriorities []ProtocolType
}

// ProtocolType тип протокола
type ProtocolType string

const (
	ProtocolICE       ProtocolType = "ice"
	ProtocolMASQUE    ProtocolType = "masque"
	ProtocolWebSocket ProtocolType = "websocket"
	ProtocolTCP       ProtocolType = "tcp"
)

// BridgeConnection представляет соединение через мост
type BridgeConnection struct {
	ID           string
	RemoteAddr   net.Addr
	CreatedAt    time.Time
	LastActivity time.Time

	// Выбранный протокол
	SelectedProtocol ProtocolType
	ProtocolAttempts map[ProtocolType]int

	// Состояние соединения
	IsConnected    bool
	ConnectionTime time.Duration

	// Метрики
	BytesReceived   int64
	BytesSent       int64
	PacketsReceived int64
	PacketsSent     int64
}

// BridgeStats статистика моста
type BridgeStats struct {
	TotalConnections      int64
	ActiveConnections     int64
	SuccessfulConnections int64
	FailedConnections     int64

	// Статистика по протоколам
	ICEConnections       int64
	MASQUEConnections    int64
	WebSocketConnections int64
	TCPConnections       int64

	// Метрики производительности
	AverageConnectionTime time.Duration
	TotalBytesReceived    int64
	TotalBytesSent        int64
	TotalPacketsReceived  int64
	TotalPacketsSent      int64

	// Ошибки
	ProtocolErrors map[ProtocolType]int64
	FallbackErrors int64
	TimeoutErrors  int64
}

// DefaultBridgeConfig возвращает конфигурацию по умолчанию
func DefaultBridgeConfig() *BridgeConfig {
	return &BridgeConfig{
		EnableICE:                true,
		EnableMASQUE:             true,
		EnableWebSocket:          true,
		EnableTCP:                true,
		ICETimeout:               10 * time.Second,
		MASQUETimeout:            5 * time.Second,
		WebSocketTimeout:         3 * time.Second,
		TCPTimeout:               2 * time.Second,
		ProtocolSelectionTimeout: 15 * time.Second,
		FallbackThreshold:        3,
		EnableProtocolMetrics:    true,
		ProtocolPriorities: []ProtocolType{
			ProtocolICE,
			ProtocolMASQUE,
			ProtocolWebSocket,
			ProtocolTCP,
		},
	}
}

// NewICEMASQUEBridge создает новый ICE↔MASQUE мост
func NewICEMASQUEBridge(logger *logger.Logger, config *BridgeConfig) *ICEMASQUEBridge {
	if config == nil {
		config = DefaultBridgeConfig()
	}

	return &ICEMASQUEBridge{
		logger:      logger,
		config:      config,
		connections: make(map[string]*BridgeConnection),
		stats: &BridgeStats{
			ProtocolErrors: make(map[ProtocolType]int64),
		},
	}
}

// SetICEComponents устанавливает ICE компоненты
func (imb *ICEMASQUEBridge) SetICEComponents(iceAgent *nat.ICEAgent, iceSupport *ICESupport) {
	imb.iceAgent = iceAgent
	imb.iceSupport = iceSupport
}

// SetMASQUEComponents устанавливает MASQUE компоненты
func (imb *ICEMASQUEBridge) SetMASQUEComponents(masqueServer *MASQUEServer) {
	imb.masqueServer = masqueServer
}

// SetFallbackComponents устанавливает fallback компоненты
func (imb *ICEMASQUEBridge) SetFallbackComponents(websocketFallback *WebSocketFallback, tcpFallback *TCPFallback) {
	imb.websocketFallback = websocketFallback
	imb.tcpFallback = tcpFallback
}

// HandleConnection обрабатывает новое соединение с автоматическим выбором протокола
func (imb *ICEMASQUEBridge) HandleConnection(ctx context.Context, conn net.Conn) (*BridgeConnection, error) {
	connID := fmt.Sprintf("%s-%d", conn.RemoteAddr().String(), time.Now().UnixNano())

	bridgeConn := &BridgeConnection{
		ID:               connID,
		RemoteAddr:       conn.RemoteAddr(),
		CreatedAt:        time.Now(),
		LastActivity:     time.Now(),
		ProtocolAttempts: make(map[ProtocolType]int),
	}

	imb.connMutex.Lock()
	imb.connections[connID] = bridgeConn
	imb.connMutex.Unlock()

	imb.stats.TotalConnections++
	imb.stats.ActiveConnections++

	imb.logger.Info("New bridge connection established",
		zap.String("conn_id", connID),
		zap.String("remote_addr", conn.RemoteAddr().String()),
	)

	// Выбираем лучший протокол
	selectedProtocol, err := imb.SelectBestProtocol(ctx, bridgeConn, conn)
	if err != nil {
		imb.logger.Error("Failed to select protocol",
			zap.String("conn_id", connID),
			zap.Error(err),
		)
		imb.stats.FailedConnections++
		return nil, fmt.Errorf("failed to select protocol: %w", err)
	}

	bridgeConn.SelectedProtocol = selectedProtocol
	imb.stats.SuccessfulConnections++

	// Обновляем статистику по протоколам
	imb.updateProtocolStats(selectedProtocol)

	imb.logger.Info("Protocol selected for connection",
		zap.String("conn_id", connID),
		zap.String("protocol", string(selectedProtocol)),
	)

	// Запускаем обработку соединения
	go imb.handleConnection(ctx, bridgeConn, conn)

	return bridgeConn, nil
}

// SelectBestProtocol выбирает лучший протокол для соединения
func (imb *ICEMASQUEBridge) SelectBestProtocol(ctx context.Context, bridgeConn *BridgeConnection, conn net.Conn) (ProtocolType, error) {
	ctx, cancel := context.WithTimeout(ctx, imb.config.ProtocolSelectionTimeout)
	defer cancel()

	// Пробуем протоколы в порядке приоритета
	for _, protocol := range imb.config.ProtocolPriorities {
		if !imb.isProtocolEnabled(protocol) {
			continue
		}

		imb.logger.Debug("Attempting protocol",
			zap.String("conn_id", bridgeConn.ID),
			zap.String("protocol", string(protocol)),
		)

		// Пытаемся установить соединение через выбранный протокол
		success, err := imb.attemptProtocol(ctx, protocol, bridgeConn, conn)
		if err != nil {
			imb.logger.Debug("Protocol attempt failed",
				zap.String("conn_id", bridgeConn.ID),
				zap.String("protocol", string(protocol)),
				zap.Error(err),
			)
			imb.stats.ProtocolErrors[protocol]++
			bridgeConn.ProtocolAttempts[protocol]++
			continue
		}

		if success {
			imb.logger.Info("Protocol selected successfully",
				zap.String("conn_id", bridgeConn.ID),
				zap.String("protocol", string(protocol)),
			)
			return protocol, nil
		}
	}

	return "", fmt.Errorf("all protocol attempts failed")
}

// isProtocolEnabled проверяет, включен ли протокол
func (imb *ICEMASQUEBridge) isProtocolEnabled(protocol ProtocolType) bool {
	switch protocol {
	case ProtocolICE:
		return imb.config.EnableICE && imb.iceAgent != nil
	case ProtocolMASQUE:
		return imb.config.EnableMASQUE && imb.masqueServer != nil
	case ProtocolWebSocket:
		return imb.config.EnableWebSocket && imb.websocketFallback != nil
	case ProtocolTCP:
		return imb.config.EnableTCP && imb.tcpFallback != nil
	default:
		return false
	}
}

// attemptProtocol пытается установить соединение через указанный протокол
func (imb *ICEMASQUEBridge) attemptProtocol(ctx context.Context, protocol ProtocolType, bridgeConn *BridgeConnection, conn net.Conn) (bool, error) {
	switch protocol {
	case ProtocolICE:
		return imb.attemptICE(ctx, bridgeConn, conn)
	case ProtocolMASQUE:
		return imb.attemptMASQUE(ctx, bridgeConn, conn)
	case ProtocolWebSocket:
		return imb.attemptWebSocket(ctx, bridgeConn, conn)
	case ProtocolTCP:
		return imb.attemptTCP(ctx, bridgeConn, conn)
	default:
		return false, fmt.Errorf("unknown protocol: %s", protocol)
	}
}

// attemptICE пытается установить соединение через ICE
func (imb *ICEMASQUEBridge) attemptICE(ctx context.Context, bridgeConn *BridgeConnection, conn net.Conn) (bool, error) {
	if imb.iceAgent == nil {
		return false, fmt.Errorf("ICE agent not available")
	}

	ctx, cancel := context.WithTimeout(ctx, imb.config.ICETimeout)
	defer cancel()

	// Здесь должна быть логика установки ICE соединения
	// Для демонстрации считаем, что ICE соединение успешно
	time.Sleep(100 * time.Millisecond) // Имитация времени установки соединения

	imb.logger.Debug("ICE connection established",
		zap.String("conn_id", bridgeConn.ID),
	)

	return true, nil
}

// attemptMASQUE пытается установить соединение через MASQUE
func (imb *ICEMASQUEBridge) attemptMASQUE(ctx context.Context, bridgeConn *BridgeConnection, conn net.Conn) (bool, error) {
	if imb.masqueServer == nil {
		return false, fmt.Errorf("MASQUE server not available")
	}

	ctx, cancel := context.WithTimeout(ctx, imb.config.MASQUETimeout)
	defer cancel()

	// Здесь должна быть логика установки MASQUE соединения
	// Для демонстрации считаем, что MASQUE соединение успешно
	time.Sleep(50 * time.Millisecond) // Имитация времени установки соединения

	imb.logger.Debug("MASQUE connection established",
		zap.String("conn_id", bridgeConn.ID),
	)

	return true, nil
}

// attemptWebSocket пытается установить соединение через WebSocket
func (imb *ICEMASQUEBridge) attemptWebSocket(ctx context.Context, bridgeConn *BridgeConnection, conn net.Conn) (bool, error) {
	if imb.websocketFallback == nil {
		return false, fmt.Errorf("WebSocket fallback not available")
	}

	ctx, cancel := context.WithTimeout(ctx, imb.config.WebSocketTimeout)
	defer cancel()

	// Здесь должна быть логика установки WebSocket соединения
	// Для демонстрации считаем, что WebSocket соединение успешно
	time.Sleep(30 * time.Millisecond) // Имитация времени установки соединения

	imb.logger.Debug("WebSocket connection established",
		zap.String("conn_id", bridgeConn.ID),
	)

	return true, nil
}

// attemptTCP пытается установить соединение через TCP
func (imb *ICEMASQUEBridge) attemptTCP(ctx context.Context, bridgeConn *BridgeConnection, conn net.Conn) (bool, error) {
	if imb.tcpFallback == nil {
		return false, fmt.Errorf("TCP fallback not available")
	}

	ctx, cancel := context.WithTimeout(ctx, imb.config.TCPTimeout)
	defer cancel()

	// Здесь должна быть логика установки TCP соединения
	// Для демонстрации считаем, что TCP соединение успешно
	time.Sleep(20 * time.Millisecond) // Имитация времени установки соединения

	imb.logger.Debug("TCP connection established",
		zap.String("conn_id", bridgeConn.ID),
	)

	return true, nil
}

// handleConnection обрабатывает соединение через выбранный протокол
func (imb *ICEMASQUEBridge) handleConnection(ctx context.Context, bridgeConn *BridgeConnection, conn net.Conn) {
	defer func() {
		imb.connMutex.Lock()
		delete(imb.connections, bridgeConn.ID)
		imb.connMutex.Unlock()
		imb.stats.ActiveConnections--

		if err := conn.Close(); err != nil {
			imb.logger.Debug("Error closing connection",
				zap.String("conn_id", bridgeConn.ID),
				zap.Error(err),
			)
		}
	}()

	startTime := time.Now()
	bridgeConn.IsConnected = true

	// Обрабатываем соединение в зависимости от выбранного протокола
	switch bridgeConn.SelectedProtocol {
	case ProtocolICE:
		imb.handleICEConnection(ctx, bridgeConn, conn)
	case ProtocolMASQUE:
		imb.handleMASQUEConnection(ctx, bridgeConn, conn)
	case ProtocolWebSocket:
		imb.handleWebSocketConnection(ctx, bridgeConn, conn)
	case ProtocolTCP:
		imb.handleTCPConnection(ctx, bridgeConn, conn)
	default:
		imb.logger.Error("Unknown protocol for connection",
			zap.String("conn_id", bridgeConn.ID),
			zap.String("protocol", string(bridgeConn.SelectedProtocol)),
		)
	}

	// Обновляем метрики
	bridgeConn.ConnectionTime = time.Since(startTime)
	imb.updateConnectionMetrics(bridgeConn)
}

// handleICEConnection обрабатывает ICE соединение
func (imb *ICEMASQUEBridge) handleICEConnection(ctx context.Context, bridgeConn *BridgeConnection, conn net.Conn) {
	imb.logger.Debug("Handling ICE connection",
		zap.String("conn_id", bridgeConn.ID),
	)

	// Здесь должна быть логика обработки ICE соединения
	// Для демонстрации просто ждем завершения контекста
	<-ctx.Done()
}

// handleMASQUEConnection обрабатывает MASQUE соединение
func (imb *ICEMASQUEBridge) handleMASQUEConnection(ctx context.Context, bridgeConn *BridgeConnection, conn net.Conn) {
	imb.logger.Debug("Handling MASQUE connection",
		zap.String("conn_id", bridgeConn.ID),
	)

	// Здесь должна быть логика обработки MASQUE соединения
	// Для демонстрации просто ждем завершения контекста
	<-ctx.Done()
}

// handleWebSocketConnection обрабатывает WebSocket соединение
func (imb *ICEMASQUEBridge) handleWebSocketConnection(ctx context.Context, bridgeConn *BridgeConnection, conn net.Conn) {
	imb.logger.Debug("Handling WebSocket connection",
		zap.String("conn_id", bridgeConn.ID),
	)

	// Здесь должна быть логика обработки WebSocket соединения
	// Для демонстрации просто ждем завершения контекста
	<-ctx.Done()
}

// handleTCPConnection обрабатывает TCP соединение
func (imb *ICEMASQUEBridge) handleTCPConnection(ctx context.Context, bridgeConn *BridgeConnection, conn net.Conn) {
	imb.logger.Debug("Handling TCP connection",
		zap.String("conn_id", bridgeConn.ID),
	)

	// Здесь должна быть логика обработки TCP соединения
	// Для демонстрации просто ждем завершения контекста
	<-ctx.Done()
}

// updateProtocolStats обновляет статистику по протоколам
func (imb *ICEMASQUEBridge) updateProtocolStats(protocol ProtocolType) {
	switch protocol {
	case ProtocolICE:
		imb.stats.ICEConnections++
	case ProtocolMASQUE:
		imb.stats.MASQUEConnections++
	case ProtocolWebSocket:
		imb.stats.WebSocketConnections++
	case ProtocolTCP:
		imb.stats.TCPConnections++
	}
}

// updateConnectionMetrics обновляет метрики соединения
func (imb *ICEMASQUEBridge) updateConnectionMetrics(bridgeConn *BridgeConnection) {
	imb.stats.TotalBytesReceived += bridgeConn.BytesReceived
	imb.stats.TotalBytesSent += bridgeConn.BytesSent
	imb.stats.TotalPacketsReceived += bridgeConn.PacketsReceived
	imb.stats.TotalPacketsSent += bridgeConn.PacketsSent

	// Обновляем среднее время соединения
	if imb.stats.SuccessfulConnections > 0 {
		totalTime := imb.stats.AverageConnectionTime * time.Duration(imb.stats.SuccessfulConnections-1)
		imb.stats.AverageConnectionTime = (totalTime + bridgeConn.ConnectionTime) / time.Duration(imb.stats.SuccessfulConnections)
	}
}

// GetStats возвращает статистику моста
func (imb *ICEMASQUEBridge) GetStats() *BridgeStats {
	return imb.stats
}

// GetConnectionCount возвращает количество активных соединений
func (imb *ICEMASQUEBridge) GetConnectionCount() int {
	imb.connMutex.RLock()
	defer imb.connMutex.RUnlock()
	return len(imb.connections)
}

// GetConnectionByID возвращает соединение по ID
func (imb *ICEMASQUEBridge) GetConnectionByID(connID string) (*BridgeConnection, bool) {
	imb.connMutex.RLock()
	defer imb.connMutex.RUnlock()
	conn, exists := imb.connections[connID]
	return conn, exists
}

// CleanupExpiredConnections очищает истекшие соединения
func (imb *ICEMASQUEBridge) CleanupExpiredConnections() {
	imb.connMutex.Lock()
	defer imb.connMutex.Unlock()

	now := time.Now()
	for connID, conn := range imb.connections {
		if now.Sub(conn.LastActivity) > 30*time.Minute {
			imb.logger.Debug("Cleaning up expired connection",
				zap.String("conn_id", connID),
				zap.Duration("idle_time", now.Sub(conn.LastActivity)),
			)

			delete(imb.connections, connID)
			imb.stats.ActiveConnections--
		}
	}
}

// StartCleanupRoutine запускает горутину для очистки истекших соединений
func (imb *ICEMASQUEBridge) StartCleanupRoutine(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			imb.CleanupExpiredConnections()
		}
	}
}
