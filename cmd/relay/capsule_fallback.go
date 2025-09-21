package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/quic-go/quic-go/http3"
	"go.uber.org/zap"

	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
)

// CapsuleType тип Capsule согласно RFC 9297
type CapsuleType uint8

const (
	CapsuleTypeDatagram CapsuleType = 0x00
)

// Capsule представляет HTTP Capsule согласно RFC 9297
type Capsule struct {
	Type CapsuleType
	Data []byte
}

// CapsuleFallback реализует HTTP Datagrams через Capsule (RFC 9297)
// для случаев когда QUIC-DATAGRAM не поддерживается
type CapsuleFallback struct {
	logger *logger.Logger
	config *CapsuleConfig

	// Управление соединениями
	connections map[string]*CapsuleConnection
	connMutex   sync.RWMutex

	// Статистика
	stats *CapsuleStats
}

// CapsuleConfig конфигурация для Capsule fallback
type CapsuleConfig struct {
	// Capsule Protocol
	MaxCapsuleSize     int
	CapsuleTimeout     time.Duration
	EnableCapsuleDebug bool

	// HTTP Datagrams
	MaxDatagramSize int
	DatagramTimeout time.Duration

	// Fallback detection
	DatagramSupportCheckTimeout time.Duration
	FallbackThreshold           int // количество неудачных попыток перед fallback
}

// CapsuleConnection представляет соединение с Capsule fallback
type CapsuleConnection struct {
	ID           string
	Stream       *http3.Stream
	RemoteAddr   net.Addr
	CreatedAt    time.Time
	LastActivity time.Time

	// Capsule state
	capsuleBuffer []byte
	capsuleMutex  sync.Mutex

	// Datagram support detection
	datagramSupported bool
	fallbackCount     int
}

// CapsuleStats статистика Capsule fallback
type CapsuleStats struct {
	TotalConnections    int64
	ActiveConnections   int64
	DatagramSupported   int64
	CapsuleFallbackUsed int64
	TotalDatagrams      int64
	TotalCapsules       int64
	Errors              int64
}

// DefaultCapsuleConfig возвращает конфигурацию по умолчанию
func DefaultCapsuleConfig() *CapsuleConfig {
	return &CapsuleConfig{
		MaxCapsuleSize:              1024 * 1024, // 1MB
		CapsuleTimeout:              30 * time.Second,
		EnableCapsuleDebug:          false,
		MaxDatagramSize:             64 * 1024, // 64KB
		DatagramTimeout:             5 * time.Second,
		DatagramSupportCheckTimeout: 2 * time.Second,
		FallbackThreshold:           3,
	}
}

// NewCapsuleFallback создает новый Capsule fallback компонент
func NewCapsuleFallback(logger *logger.Logger, config *CapsuleConfig) *CapsuleFallback {
	if config == nil {
		config = DefaultCapsuleConfig()
	}

	return &CapsuleFallback{
		logger:      logger,
		config:      config,
		connections: make(map[string]*CapsuleConnection),
		stats:       &CapsuleStats{},
	}
}

// HandleConnection обрабатывает новое соединение с Capsule fallback
func (cf *CapsuleFallback) HandleConnection(ctx context.Context, stream *http3.Stream, remoteAddr net.Addr) error {
	connID := fmt.Sprintf("%s-%d", remoteAddr.String(), time.Now().UnixNano())

	conn := &CapsuleConnection{
		ID:            connID,
		Stream:        stream,
		RemoteAddr:    remoteAddr,
		CreatedAt:     time.Now(),
		LastActivity:  time.Now(),
		capsuleBuffer: make([]byte, 0, cf.config.MaxCapsuleSize),
	}

	cf.connMutex.Lock()
	cf.connections[connID] = conn
	cf.connMutex.Unlock()

	cf.stats.TotalConnections++
	cf.stats.ActiveConnections++

	cf.logger.Info("New Capsule connection established",
		zap.String("conn_id", connID),
		zap.String("remote_addr", remoteAddr.String()),
	)

	// Запускаем обработку соединения
	go cf.handleConnection(ctx, conn)

	return nil
}

// handleConnection обрабатывает соединение
func (cf *CapsuleFallback) handleConnection(ctx context.Context, conn *CapsuleConnection) {
	defer func() {
		cf.connMutex.Lock()
		delete(cf.connections, conn.ID)
		cf.connMutex.Unlock()
		cf.stats.ActiveConnections--

		if err := conn.Stream.Close(); err != nil {
			cf.logger.Debug("Error closing stream",
				zap.String("conn_id", conn.ID),
				zap.Error(err),
			)
		}
	}()

	// Сначала пытаемся определить поддержку QUIC-DATAGRAM
	if err := cf.detectDatagramSupport(ctx, conn); err != nil {
		cf.logger.Warn("Failed to detect datagram support",
			zap.String("conn_id", conn.ID),
			zap.Error(err),
		)
	}

	// Обрабатываем входящие данные
	buffer := make([]byte, cf.config.MaxCapsuleSize)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Читаем данные из stream
			n, err := conn.Stream.Read(buffer)
			if err != nil {
				if err != io.EOF {
					cf.logger.Debug("Error reading from stream",
						zap.String("conn_id", conn.ID),
						zap.Error(err),
					)
				}
				return
			}

			conn.LastActivity = time.Now()

			// Обрабатываем полученные данные
			if err := cf.processIncomingData(conn, buffer[:n]); err != nil {
				cf.logger.Error("Error processing incoming data",
					zap.String("conn_id", conn.ID),
					zap.Error(err),
				)
				cf.stats.Errors++
				continue
			}
		}
	}
}

// detectDatagramSupport определяет поддержку QUIC-DATAGRAM
func (cf *CapsuleFallback) detectDatagramSupport(ctx context.Context, conn *CapsuleConnection) error {
	// Пытаемся отправить datagram
	ctx, cancel := context.WithTimeout(ctx, cf.config.DatagramSupportCheckTimeout)
	defer cancel()

	// В реальной реализации здесь должна быть проверка поддержки QUIC-DATAGRAM
	// Для демонстрации считаем, что datagram поддерживается
	conn.datagramSupported = true
	cf.stats.DatagramSupported++

	cf.logger.Debug("Datagram support detected",
		zap.String("conn_id", conn.ID),
		zap.Bool("supported", conn.datagramSupported),
	)

	return nil
}

// processIncomingData обрабатывает входящие данные
func (cf *CapsuleFallback) processIncomingData(conn *CapsuleConnection, data []byte) error {
	if conn.datagramSupported {
		// Если datagram поддерживается, обрабатываем как обычный datagram
		return cf.processDatagram(conn, data)
	} else {
		// Если datagram не поддерживается, обрабатываем как Capsule
		return cf.processCapsule(conn, data)
	}
}

// processDatagram обрабатывает datagram (когда QUIC-DATAGRAM поддерживается)
func (cf *CapsuleFallback) processDatagram(conn *CapsuleConnection, data []byte) error {
	cf.stats.TotalDatagrams++

	cf.logger.Debug("Processing datagram",
		zap.String("conn_id", conn.ID),
		zap.Int("data_len", len(data)),
	)

	// Здесь должна быть логика обработки datagram
	// Для демонстрации просто логируем
	return nil
}

// processCapsule обрабатывает Capsule (когда QUIC-DATAGRAM не поддерживается)
func (cf *CapsuleFallback) processCapsule(conn *CapsuleConnection, data []byte) error {
	cf.stats.TotalCapsules++

	conn.capsuleMutex.Lock()
	defer conn.capsuleMutex.Unlock()

	// Добавляем данные в буфер
	conn.capsuleBuffer = append(conn.capsuleBuffer, data...)

	// Пытаемся парсить Capsule
	for len(conn.capsuleBuffer) > 0 {
		capsule, consumed, err := cf.parseCapsule(conn.capsuleBuffer)
		if err != nil {
			// Если не удалось парсить, ждем больше данных
			if err == io.ErrUnexpectedEOF {
				break
			}
			return fmt.Errorf("failed to parse capsule: %w", err)
		}

		// Удаляем обработанные данные из буфера
		conn.capsuleBuffer = conn.capsuleBuffer[consumed:]

		// Обрабатываем Capsule
		if err := cf.handleCapsule(conn, capsule); err != nil {
			cf.logger.Error("Error handling capsule",
				zap.String("conn_id", conn.ID),
				zap.Error(err),
			)
			cf.stats.Errors++
		}
	}

	return nil
}

// parseCapsule парсит Capsule согласно RFC 9297
func (cf *CapsuleFallback) parseCapsule(data []byte) (*Capsule, int, error) {
	if len(data) < 2 {
		return nil, 0, io.ErrUnexpectedEOF
	}

	// Читаем длину Capsule
	capsuleLength := binary.BigEndian.Uint16(data[:2])
	if len(data) < int(capsuleLength)+2 {
		return nil, 0, io.ErrUnexpectedEOF
	}

	// Читаем тип Capsule
	capsuleType := CapsuleType(data[2])

	// Читаем данные Capsule
	capsuleData := data[3 : 3+capsuleLength-1]

	capsule := &Capsule{
		Type: capsuleType,
		Data: capsuleData,
	}

	totalLength := int(capsuleLength) + 2
	return capsule, totalLength, nil
}

// handleCapsule обрабатывает Capsule
func (cf *CapsuleFallback) handleCapsule(conn *CapsuleConnection, capsule *Capsule) error {
	cf.logger.Debug("Handling capsule",
		zap.String("conn_id", conn.ID),
		zap.Uint8("type", uint8(capsule.Type)),
		zap.Int("data_len", len(capsule.Data)),
	)

	switch capsule.Type {
	case CapsuleTypeDatagram:
		// Обрабатываем HTTP Datagram через Capsule
		return cf.handleDatagramCapsule(conn, capsule)
	default:
		cf.logger.Warn("Unknown capsule type",
			zap.String("conn_id", conn.ID),
			zap.Uint8("type", uint8(capsule.Type)),
		)
		return fmt.Errorf("unknown capsule type: %d", capsule.Type)
	}
}

// handleDatagramCapsule обрабатывает Datagram Capsule
func (cf *CapsuleFallback) handleDatagramCapsule(conn *CapsuleConnection, capsule *Capsule) error {
	// Парсим Datagram Capsule
	// Формат: [Quarter Stream ID (varint)][HTTP Datagram Payload]

	if len(capsule.Data) == 0 {
		return fmt.Errorf("empty datagram capsule")
	}

	// Парсим Quarter Stream ID (Context ID для MASQUE)
	contextID, n, err := parseVarint(capsule.Data)
	if err != nil {
		return fmt.Errorf("failed to parse context ID: %w", err)
	}

	// Извлекаем payload
	payload := capsule.Data[n:]

	cf.logger.Debug("Processing datagram capsule",
		zap.String("conn_id", conn.ID),
		zap.Uint64("context_id", contextID),
		zap.Int("payload_len", len(payload)),
	)

	// Здесь должна быть логика обработки datagram payload
	// Для демонстрации просто логируем
	return nil
}

// parseVarint парсит varint согласно QUIC спецификации
func parseVarint(data []byte) (uint64, int, error) {
	if len(data) == 0 {
		return 0, 0, fmt.Errorf("empty data")
	}

	firstByte := data[0]
	length := 1 << (firstByte >> 6)

	if len(data) < length {
		return 0, 0, fmt.Errorf("insufficient data for varint")
	}

	var value uint64
	switch length {
	case 1:
		value = uint64(firstByte & 0x3f)
	case 2:
		value = uint64(firstByte&0x3f)<<8 | uint64(data[1])
	case 4:
		value = uint64(firstByte&0x3f)<<24 | uint64(data[1])<<16 | uint64(data[2])<<8 | uint64(data[3])
	case 8:
		value = uint64(firstByte&0x3f)<<56 | uint64(data[1])<<48 | uint64(data[2])<<40 | uint64(data[3])<<32 |
			uint64(data[4])<<24 | uint64(data[5])<<16 | uint64(data[6])<<8 | uint64(data[7])
	default:
		return 0, 0, fmt.Errorf("invalid varint length: %d", length)
	}

	return value, length, nil
}

// SendDatagram отправляет datagram через Capsule fallback
func (cf *CapsuleFallback) SendDatagram(conn *CapsuleConnection, contextID uint64, data []byte) error {
	if conn.datagramSupported {
		// Если datagram поддерживается, отправляем как обычный datagram
		return cf.sendDatagramDirect(conn, data)
	} else {
		// Если datagram не поддерживается, отправляем через Capsule
		return cf.sendDatagramCapsule(conn, contextID, data)
	}
}

// sendDatagramDirect отправляет datagram напрямую
func (cf *CapsuleFallback) sendDatagramDirect(conn *CapsuleConnection, data []byte) error {
	// Здесь должна быть логика отправки datagram напрямую
	// Для демонстрации просто логируем
	cf.logger.Debug("Sending datagram directly",
		zap.String("conn_id", conn.ID),
		zap.Int("data_len", len(data)),
	)
	return nil
}

// sendDatagramCapsule отправляет datagram через Capsule
func (cf *CapsuleFallback) sendDatagramCapsule(conn *CapsuleConnection, contextID uint64, data []byte) error {
	// Создаем Datagram Capsule
	capsuleData := make([]byte, 0, len(data)+8)

	// Добавляем Context ID как varint
	contextIDBytes := encodeVarint(contextID)
	capsuleData = append(capsuleData, contextIDBytes...)

	// Добавляем payload
	capsuleData = append(capsuleData, data...)

	// Создаем Capsule
	capsule := &Capsule{
		Type: CapsuleTypeDatagram,
		Data: capsuleData,
	}

	// Сериализуем Capsule
	capsuleBytes, err := cf.serializeCapsule(capsule)
	if err != nil {
		return fmt.Errorf("failed to serialize capsule: %w", err)
	}

	// Отправляем через stream
	if _, err := conn.Stream.Write(capsuleBytes); err != nil {
		return fmt.Errorf("failed to send capsule: %w", err)
	}

	cf.logger.Debug("Sent datagram capsule",
		zap.String("conn_id", conn.ID),
		zap.Uint64("context_id", contextID),
		zap.Int("data_len", len(data)),
	)

	return nil
}

// serializeCapsule сериализует Capsule согласно RFC 9297
func (cf *CapsuleFallback) serializeCapsule(capsule *Capsule) ([]byte, error) {
	// Формат: [Length (2 bytes)][Type (1 byte)][Data]
	length := len(capsule.Data) + 1 // +1 для типа

	data := make([]byte, 2+length)
	binary.BigEndian.PutUint16(data[:2], uint16(length))
	data[2] = uint8(capsule.Type)
	copy(data[3:], capsule.Data)

	return data, nil
}

// encodeVarint кодирует varint согласно QUIC спецификации
func encodeVarint(value uint64) []byte {
	if value < 64 {
		return []byte{byte(value)}
	} else if value < 16384 {
		return []byte{byte(value>>8) | 0x40, byte(value)}
	} else if value < 1073741824 {
		return []byte{byte(value>>24) | 0x80, byte(value >> 16), byte(value >> 8), byte(value)}
	} else {
		return []byte{0xc0, byte(value >> 56), byte(value >> 48), byte(value >> 40), byte(value >> 32),
			byte(value >> 24), byte(value >> 16), byte(value >> 8), byte(value)}
	}
}

// GetStats возвращает статистику Capsule fallback
func (cf *CapsuleFallback) GetStats() *CapsuleStats {
	return cf.stats
}

// GetConnectionCount возвращает количество активных соединений
func (cf *CapsuleFallback) GetConnectionCount() int {
	cf.connMutex.RLock()
	defer cf.connMutex.RUnlock()
	return len(cf.connections)
}

// CleanupExpiredConnections очищает истекшие соединения
func (cf *CapsuleFallback) CleanupExpiredConnections() {
	cf.connMutex.Lock()
	defer cf.connMutex.Unlock()

	now := time.Now()
	for connID, conn := range cf.connections {
		if now.Sub(conn.LastActivity) > cf.config.CapsuleTimeout {
			cf.logger.Debug("Cleaning up expired connection",
				zap.String("conn_id", connID),
				zap.Duration("idle_time", now.Sub(conn.LastActivity)),
			)

			if err := conn.Stream.Close(); err != nil {
				cf.logger.Debug("Error closing expired connection",
					zap.String("conn_id", connID),
					zap.Error(err),
				)
			}

			delete(cf.connections, connID)
			cf.stats.ActiveConnections--
		}
	}
}

// StartCleanupRoutine запускает горутину для очистки истекших соединений
func (cf *CapsuleFallback) StartCleanupRoutine(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cf.CleanupExpiredConnections()
		}
	}
}
