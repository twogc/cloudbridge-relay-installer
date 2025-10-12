package quic

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/quic-go/quic-go"
	"go.uber.org/zap"
)

// QLogTracer представляет qlog трейсер для QUIC соединений
type QLogTracer struct {
	logger     *zap.Logger
	outputDir  string
	enabled    bool
	mu         sync.RWMutex
	connections map[string]*QLogConnection
}

// QLogConnection представляет qlog соединение
type QLogConnection struct {
	ConnectionID string
	StartTime    time.Time
	Events       []QLogEvent
	mu           sync.RWMutex
}

// QLogEvent представляет qlog событие
type QLogEvent struct {
	Time     time.Time              `json:"time"`
	Category string                 `json:"category"`
	Event    string                 `json:"event"`
	Data     map[string]interface{} `json:"data,omitempty"`
}

// QLogConfig конфигурация для qlog трейсера
type QLogConfig struct {
	Enabled    bool   `json:"enabled"`
	OutputDir  string `json:"output_dir"`
	MaxEvents  int    `json:"max_events"`
	FlushInterval time.Duration `json:"flush_interval"`
}

// DefaultQLogConfig возвращает конфигурацию по умолчанию
func DefaultQLogConfig() *QLogConfig {
	return &QLogConfig{
		Enabled:       true,
		OutputDir:     "./qlog",
		MaxEvents:     10000,
		FlushInterval: 30 * time.Second,
	}
}

// NewQLogTracer создает новый qlog трейсер
func NewQLogTracer(logger *zap.Logger, config *QLogConfig) *QLogTracer {
	if config == nil {
		config = DefaultQLogConfig()
	}

	// Создаем директорию для qlog файлов
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		logger.Error("Failed to create qlog directory", zap.Error(err))
	}

	return &QLogTracer{
		logger:      logger,
		outputDir:   config.OutputDir,
		enabled:     config.Enabled,
		connections: make(map[string]*QLogConnection),
	}
}

// StartConnection начинает трейсинг QUIC соединения
func (qt *QLogTracer) StartConnection(connID string, conn *quic.Conn) {
	if !qt.enabled {
		return
	}

	qt.mu.Lock()
	defer qt.mu.Unlock()

	qconn := &QLogConnection{
		ConnectionID: connID,
		StartTime:    time.Now(),
		Events:       make([]QLogEvent, 0),
	}

	qt.connections[connID] = qconn

	// Получаем адреса если соединение не nil
	var remoteAddr, localAddr string
	if conn != nil {
		remoteAddr = conn.RemoteAddr().String()
		localAddr = conn.LocalAddr().String()
	} else {
		remoteAddr = "unknown"
		localAddr = "unknown"
	}

	// Логируем начало соединения
	qt.logEvent(connID, "connectivity", "connection_started", map[string]interface{}{
		"connection_id": connID,
		"remote_addr":   remoteAddr,
		"local_addr":    localAddr,
	})

	qt.logger.Info("Started qlog tracing for connection",
		zap.String("connection_id", connID),
		zap.String("remote_addr", remoteAddr),
	)
}

// EndConnection завершает трейсинг QUIC соединения
func (qt *QLogTracer) EndConnection(connID string, reason string) {
	if !qt.enabled {
		return
	}

	qt.mu.Lock()
	defer qt.mu.Unlock()

	if qconn, exists := qt.connections[connID]; exists {
		// Логируем завершение соединения
		qt.logEvent(connID, "connectivity", "connection_ended", map[string]interface{}{
			"connection_id": connID,
			"reason":        reason,
			"duration":      time.Since(qconn.StartTime).String(),
		})

		// Сохраняем qlog файл
		qt.saveQLogFile(connID, qconn)

		// Удаляем из активных соединений
		delete(qt.connections, connID)
	}

	qt.logger.Info("Ended qlog tracing for connection",
		zap.String("connection_id", connID),
		zap.String("reason", reason),
	)
}

// LogPacketReceived логирует получение пакета
func (qt *QLogTracer) LogPacketReceived(connID string, packetType string, size int, data map[string]interface{}) {
	if !qt.enabled {
		return
	}

	eventData := map[string]interface{}{
		"packet_type": packetType,
		"size":        size,
	}
	
	// Добавляем дополнительные данные
	for k, v := range data {
		eventData[k] = v
	}

	qt.logEvent(connID, "transport", "packet_received", eventData)
}

// LogPacketSent логирует отправку пакета
func (qt *QLogTracer) LogPacketSent(connID string, packetType string, size int, data map[string]interface{}) {
	if !qt.enabled {
		return
	}

	eventData := map[string]interface{}{
		"packet_type": packetType,
		"size":        size,
	}
	
	// Добавляем дополнительные данные
	for k, v := range data {
		eventData[k] = v
	}

	qt.logEvent(connID, "transport", "packet_sent", eventData)
}

// LogStreamEvent логирует событие потока
func (qt *QLogTracer) LogStreamEvent(connID string, streamID uint64, event string, data map[string]interface{}) {
	if !qt.enabled {
		return
	}

	eventData := map[string]interface{}{
		"stream_id": streamID,
	}
	
	// Добавляем дополнительные данные
	for k, v := range data {
		eventData[k] = v
	}

	qt.logEvent(connID, "transport", event, eventData)
}

// LogSecurityEvent логирует событие безопасности
func (qt *QLogTracer) LogSecurityEvent(connID string, event string, data map[string]interface{}) {
	if !qt.enabled {
		return
	}

	qt.logEvent(connID, "security", event, data)
}

// logEvent добавляет событие в qlog
func (qt *QLogTracer) logEvent(connID string, category string, event string, data map[string]interface{}) {
	qt.mu.RLock()
	qconn, exists := qt.connections[connID]
	qt.mu.RUnlock()

	if !exists {
		return
	}

	qconn.mu.Lock()
	defer qconn.mu.Unlock()

	qlogEvent := QLogEvent{
		Time:     time.Now(),
		Category: category,
		Event:    event,
		Data:     data,
	}

	qconn.Events = append(qconn.Events, qlogEvent)

	// Ограничиваем количество событий
	if len(qconn.Events) > 10000 {
		qconn.Events = qconn.Events[1000:] // Удаляем старые события
	}
}

// saveQLogFile сохраняет qlog файл
func (qt *QLogTracer) saveQLogFile(connID string, qconn *QLogConnection) {
	filename := fmt.Sprintf("qlog_%s_%d.json", connID, time.Now().Unix())
	filepath := filepath.Join(qt.outputDir, filename)

	qconn.mu.RLock()
	events := make([]QLogEvent, len(qconn.Events))
	copy(events, qconn.Events)
	qconn.mu.RUnlock()

	// Создаем qlog структуру
	qlogData := map[string]interface{}{
		"qlog_version": "0.3",
		"title":        "CloudBridge Relay QUIC qlog",
		"description":  "QUIC connection trace for CloudBridge Relay",
		"traces": []map[string]interface{}{
			{
				"vantage_point": map[string]interface{}{
					"type": "server",
					"name": "CloudBridge Relay",
				},
				"common_fields": map[string]interface{}{
					"reference_time": qconn.StartTime.Format(time.RFC3339Nano),
				},
				"events": events,
			},
		},
	}

	// Сохраняем в файл
	file, err := os.Create(filepath)
	if err != nil {
		qt.logger.Error("Failed to create qlog file", zap.Error(err))
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(qlogData); err != nil {
		qt.logger.Error("Failed to encode qlog data", zap.Error(err))
		return
	}

	qt.logger.Info("Saved qlog file",
		zap.String("filename", filename),
		zap.Int("events_count", len(events)),
	)
}

// GetConnectionStats возвращает статистику соединения
func (qt *QLogTracer) GetConnectionStats(connID string) map[string]interface{} {
	qt.mu.RLock()
	defer qt.mu.RUnlock()

	if qconn, exists := qt.connections[connID]; exists {
		qconn.mu.RLock()
		defer qconn.mu.RUnlock()

		return map[string]interface{}{
			"connection_id": connID,
			"start_time":    qconn.StartTime,
			"duration":      time.Since(qconn.StartTime).String(),
			"events_count":  len(qconn.Events),
		}
	}

	return nil
}

// GetAllConnections возвращает все активные соединения
func (qt *QLogTracer) GetAllConnections() []string {
	qt.mu.RLock()
	defer qt.mu.RUnlock()

	connections := make([]string, 0, len(qt.connections))
	for connID := range qt.connections {
		connections = append(connections, connID)
	}

	return connections
}

// FlushAllConnections принудительно сохраняет все активные соединения
func (qt *QLogTracer) FlushAllConnections() {
	qt.mu.Lock()
	defer qt.mu.Unlock()

	for connID, qconn := range qt.connections {
		qt.saveQLogFile(connID, qconn)
	}

	qt.logger.Info("Flushed all qlog connections",
		zap.Int("connections_count", len(qt.connections)),
	)
}

// StartFlushRoutine запускает горутину для периодического сохранения
func (qt *QLogTracer) StartFlushRoutine(ctx context.Context) {
	if !qt.enabled {
		return
	}

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			qt.FlushAllConnections()
			return
		case <-ticker.C:
			qt.FlushAllConnections()
		}
	}
}
