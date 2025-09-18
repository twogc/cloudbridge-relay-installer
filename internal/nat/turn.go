package nat

import (
	"context"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"net"
	"time"

	"go.uber.org/zap"
)

// TURNClient представляет TURN клиент
type TURNClient struct {
	server   *TURNServer
	conn     *net.UDPConn
	logger   *zap.Logger
	realm    string
	nonce    string
	username string
	password string
}

// TURN константы
const (
	// TURN Message Types
	TURNAllocateRequest          = 0x0003
	TURNAllocateResponse         = 0x0103
	TURNAllocateError            = 0x0113
	TURNRefreshRequest           = 0x0004
	TURNRefreshResponse          = 0x0104
	TURNRefreshError             = 0x0114
	TURNSendIndication           = 0x0016
	TURNDataIndication           = 0x0017
	TURNCreatePermissionRequest  = 0x0008
	TURNCreatePermissionResponse = 0x0108
	TURNCreatePermissionError    = 0x0118
	TURNChannelBindRequest       = 0x0009
	TURNChannelBindResponse      = 0x0109
	TURNChannelBindError         = 0x0119

	// TURN Attribute Types
	TURNAttrRequestedTransport = 0x0019
	TURNAttrXORRelayedAddress  = 0x0016
	TURNAttrXORMappedAddress   = 0x0020
	TURNAttrLifetime           = 0x000D
	TURNAttrBandwidth          = 0x000E
	TURNAttrDestinationAddress = 0x0011
	TURNAttrData               = 0x0013
	TURNAttrXORPeerAddress     = 0x0012
	TURNAttrRealm              = 0x0014
	TURNAttrNonce              = 0x0015
	TURNAttrSoftware           = 0x8022
	TURNAttrAlternateServer    = 0x8023
	TURNAttrFingerprint        = 0x8028

	// TURN Transport Protocols
	TURNTransportUDP = 17
)

// NewTURNClient создает новый TURN клиент
func NewTURNClient(server *TURNServer, logger *zap.Logger) *TURNClient {
	return &TURNClient{
		server:   server,
		logger:   logger,
		username: server.Username,
		password: server.Password,
	}
}

// AllocateRelayCandidate выделяет relay candidate через TURN сервер
func (tc *TURNClient) AllocateRelayCandidate() (*Candidate, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 1. Подключаемся к TURN серверу
	if err := tc.connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to TURN server: %w", err)
	}
	defer tc.Close()

	// 2. Отправляем ALLOCATE request
	relayAddr, err := tc.allocate(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to allocate relay address: %w", err)
	}

	// 3. Создаем relay candidate
	candidate := &Candidate{
		Foundation: fmt.Sprintf("relay-%s", tc.server.Address),
		Component:  1,
		Transport:  "UDP",
		Priority:   0, // Will be calculated by ICE agent
		Address:    relayAddr.IP.String(),
		Port:       relayAddr.Port,
		Type:       CandidateTypeRelay,
		Username:   tc.username,
		Password:   tc.password,
	}

	tc.logger.Info("Allocated relay candidate",
		zap.String("address", candidate.Address),
		zap.Int("port", candidate.Port))

	return candidate, nil
}

// connect подключается к TURN серверу
func (tc *TURNClient) connect() error {
	conn, err := net.Dial("udp", tc.server.Address)
	if err != nil {
		return fmt.Errorf("failed to dial TURN server: %w", err)
	}

	tc.conn = conn.(*net.UDPConn)
	return nil
}

// allocate отправляет ALLOCATE request
func (tc *TURNClient) allocate(ctx context.Context) (*net.UDPAddr, error) {
	// Создаем ALLOCATE request
	request := &STUNMessage{
		Type:        TURNAllocateRequest,
		Length:      0,
		MagicCookie: STUNMagicCookie,
	}

	// Генерируем transaction ID
	copy(request.TransactionID[:], []byte("123456789012")) // Simplified for demo

	// Добавляем REQUESTED-TRANSPORT атрибут
	transport := make([]byte, 4)
	binary.BigEndian.PutUint32(transport, TURNTransportUDP)
	transport[3] = 0 // Reserved

	request.Attributes = append(request.Attributes, STUNAttribute{
		Type:   TURNAttrRequestedTransport,
		Length: uint16(len(transport)),
		Value:  transport,
	})

	// Добавляем LIFETIME атрибут (3600 секунд)
	lifetime := make([]byte, 4)
	binary.BigEndian.PutUint32(lifetime, 3600)

	request.Attributes = append(request.Attributes, STUNAttribute{
		Type:   TURNAttrLifetime,
		Length: uint16(len(lifetime)),
		Value:  lifetime,
	})

	// Если нужна аутентификация, добавляем USERNAME
	if tc.username != "" {
		request.Attributes = append(request.Attributes, STUNAttribute{
			Type:   STUNAttrUsername,
			Length: uint16(len(tc.username)),
			Value:  []byte(tc.username),
		})
	}

	// Отправляем запрос
	message, err := tc.encodeMessage(request)
	if err != nil {
		return nil, fmt.Errorf("failed to encode ALLOCATE request: %w", err)
	}

	if _, err := tc.conn.Write(message); err != nil {
		return nil, fmt.Errorf("failed to send ALLOCATE request: %w", err)
	}

	// Читаем ответ
	tc.conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	buffer := make([]byte, 1024)
	n, err := tc.conn.Read(buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to read ALLOCATE response: %w", err)
	}

	// Декодируем ответ
	response, err := tc.decodeMessage(buffer[:n])
	if err != nil {
		return nil, fmt.Errorf("failed to decode ALLOCATE response: %w", err)
	}

	// Проверяем тип ответа
	if response.Type == TURNAllocateError {
		return nil, fmt.Errorf("TURN allocation failed")
	}

	if response.Type != TURNAllocateResponse {
		return nil, fmt.Errorf("unexpected TURN message type: %d", response.Type)
	}

	// Ищем XOR-RELAYED-ADDRESS атрибут
	for _, attr := range response.Attributes {
		if attr.Type == TURNAttrXORRelayedAddress {
			addr, err := tc.parseXORRelayedAddress(attr.Value, response.TransactionID)
			if err != nil {
				return nil, fmt.Errorf("failed to parse XOR relayed address: %w", err)
			}
			return addr, nil
		}
	}

	return nil, fmt.Errorf("no relayed address found in TURN response")
}

// CreatePermission создает permission для отправки данных на указанный адрес
func (tc *TURNClient) CreatePermission(peerAddr *net.UDPAddr) error {
	request := &STUNMessage{
		Type:        TURNCreatePermissionRequest,
		Length:      0,
		MagicCookie: STUNMagicCookie,
	}

	// Генерируем transaction ID
	copy(request.TransactionID[:], []byte("123456789012")) // Simplified for demo

	// Добавляем XOR-PEER-ADDRESS атрибут
	peerAddrBytes := make([]byte, 8)
	peerAddrBytes[1] = 0x01 // IPv4
	binary.BigEndian.PutUint16(peerAddrBytes[2:4], uint16(peerAddr.Port))
	copy(peerAddrBytes[4:8], peerAddr.IP.To4())

	// XOR с magic cookie
	peerAddrBytes[2] ^= byte((STUNMagicCookie >> 16) & 0xFF)
	peerAddrBytes[3] ^= byte((STUNMagicCookie >> 8) & 0xFF)
	peerAddrBytes[4] ^= byte((STUNMagicCookie >> 0) & 0xFF)
	peerAddrBytes[5] ^= byte((STUNMagicCookie >> 24) & 0xFF)

	request.Attributes = append(request.Attributes, STUNAttribute{
		Type:   TURNAttrXORPeerAddress,
		Length: uint16(len(peerAddrBytes)),
		Value:  peerAddrBytes,
	})

	// Отправляем запрос
	message, err := tc.encodeMessage(request)
	if err != nil {
		return fmt.Errorf("failed to encode CREATE-PERMISSION request: %w", err)
	}

	if _, err := tc.conn.Write(message); err != nil {
		return fmt.Errorf("failed to send CREATE-PERMISSION request: %w", err)
	}

	// Читаем ответ
	tc.conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	buffer := make([]byte, 1024)
	n, err := tc.conn.Read(buffer)
	if err != nil {
		return fmt.Errorf("failed to read CREATE-PERMISSION response: %w", err)
	}

	response, err := tc.decodeMessage(buffer[:n])
	if err != nil {
		return fmt.Errorf("failed to decode CREATE-PERMISSION response: %w", err)
	}

	if response.Type == TURNCreatePermissionError {
		return fmt.Errorf("TURN create permission failed")
	}

	return nil
}

// SendData отправляет данные через TURN сервер
func (tc *TURNClient) SendData(data []byte, peerAddr *net.UDPAddr) error {
	// Создаем SEND indication
	indication := &STUNMessage{
		Type:        TURNSendIndication,
		Length:      0,
		MagicCookie: STUNMagicCookie,
	}

	// Генерируем transaction ID
	copy(indication.TransactionID[:], []byte("123456789012")) // Simplified for demo

	// Добавляем XOR-PEER-ADDRESS атрибут
	peerAddrBytes := make([]byte, 8)
	peerAddrBytes[1] = 0x01 // IPv4
	binary.BigEndian.PutUint16(peerAddrBytes[2:4], uint16(peerAddr.Port))
	copy(peerAddrBytes[4:8], peerAddr.IP.To4())

	// XOR с magic cookie
	peerAddrBytes[2] ^= byte((STUNMagicCookie >> 16) & 0xFF)
	peerAddrBytes[3] ^= byte((STUNMagicCookie >> 8) & 0xFF)
	peerAddrBytes[4] ^= byte((STUNMagicCookie >> 0) & 0xFF)
	peerAddrBytes[5] ^= byte((STUNMagicCookie >> 24) & 0xFF)

	indication.Attributes = append(indication.Attributes, STUNAttribute{
		Type:   TURNAttrXORPeerAddress,
		Length: uint16(len(peerAddrBytes)),
		Value:  peerAddrBytes,
	})

	// Добавляем DATA атрибут
	indication.Attributes = append(indication.Attributes, STUNAttribute{
		Type:   TURNAttrData,
		Length: uint16(len(data)),
		Value:  data,
	})

	// Отправляем indication
	message, err := tc.encodeMessage(indication)
	if err != nil {
		return fmt.Errorf("failed to encode SEND indication: %w", err)
	}

	if _, err := tc.conn.Write(message); err != nil {
		return fmt.Errorf("failed to send SEND indication: %w", err)
	}

	return nil
}

// parseXORRelayedAddress парсит XOR-RELAYED-ADDRESS атрибут
func (tc *TURNClient) parseXORRelayedAddress(data []byte, transactionID [12]byte) (*net.UDPAddr, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("XOR relayed address too short")
	}

	family := data[1]
	if family != 0x01 {
		return nil, fmt.Errorf("unsupported address family: %d", family)
	}

	// Порты и адреса зашифрованы XOR с magic cookie и transaction ID
	port := binary.BigEndian.Uint16(data[2:4])
	port ^= uint16(STUNMagicCookie >> 16)

	ipBytes := make([]byte, 4)
	copy(ipBytes, data[4:8])

	// XOR с magic cookie
	for i := 0; i < 4; i++ {
		ipBytes[i] ^= byte((STUNMagicCookie >> (8 * (3 - i))) & 0xFF)
	}

	return &net.UDPAddr{
		IP:   net.IP(ipBytes),
		Port: int(port),
	}, nil
}

// encodeMessage кодирует STUN/TURN сообщение (используем тот же метод что в STUN)
func (tc *TURNClient) encodeMessage(msg *STUNMessage) ([]byte, error) {
	// Вычисляем общую длину атрибутов
	attributesLength := 0
	for _, attr := range msg.Attributes {
		attributesLength += 4 + len(attr.Value) // Type(2) + Length(2) + Value
		// Выравнивание на 4 байта
		if len(attr.Value)%4 != 0 {
			attributesLength += 4 - (len(attr.Value) % 4)
		}
	}

	msg.Length = uint16(attributesLength)

	// Создаем буфер
	buffer := make([]byte, 20+attributesLength) // Header(20) + Attributes
	offset := 0

	// Записываем заголовок
	binary.BigEndian.PutUint16(buffer[offset:], msg.Type)
	offset += 2
	binary.BigEndian.PutUint16(buffer[offset:], msg.Length)
	offset += 2
	binary.BigEndian.PutUint32(buffer[offset:], msg.MagicCookie)
	offset += 4
	copy(buffer[offset:], msg.TransactionID[:])
	offset += 12

	// Записываем атрибуты
	for _, attr := range msg.Attributes {
		binary.BigEndian.PutUint16(buffer[offset:], attr.Type)
		offset += 2
		binary.BigEndian.PutUint16(buffer[offset:], attr.Length)
		offset += 2
		copy(buffer[offset:], attr.Value)
		offset += len(attr.Value)

		// Выравнивание на 4 байта
		padding := (4 - (len(attr.Value) % 4)) % 4
		offset += padding
	}

	return buffer, nil
}

// decodeMessage декодирует STUN/TURN сообщение (используем тот же метод что в STUN)
func (tc *TURNClient) decodeMessage(data []byte) (*STUNMessage, error) {
	if len(data) < 20 {
		return nil, fmt.Errorf("STUN/TURN message too short")
	}

	msg := &STUNMessage{}
	offset := 0

	// Читаем заголовок
	msg.Type = binary.BigEndian.Uint16(data[offset:])
	offset += 2
	msg.Length = binary.BigEndian.Uint16(data[offset:])
	offset += 2
	msg.MagicCookie = binary.BigEndian.Uint32(data[offset:])
	offset += 4
	copy(msg.TransactionID[:], data[offset:offset+12])
	offset += 12

	// Проверяем magic cookie
	if msg.MagicCookie != STUNMagicCookie {
		return nil, fmt.Errorf("invalid STUN/TURN magic cookie")
	}

	// Читаем атрибуты
	attributesEnd := offset + int(msg.Length)
	for offset < attributesEnd {
		if offset+4 > len(data) {
			break
		}

		attr := STUNAttribute{}
		attr.Type = binary.BigEndian.Uint16(data[offset:])
		offset += 2
		attr.Length = binary.BigEndian.Uint16(data[offset:])
		offset += 2

		if offset+int(attr.Length) > len(data) {
			break
		}

		attr.Value = make([]byte, attr.Length)
		copy(attr.Value, data[offset:offset+int(attr.Length)])
		offset += int(attr.Length)

		msg.Attributes = append(msg.Attributes, attr)

		// Выравнивание на 4 байта
		padding := (4 - (int(attr.Length) % 4)) % 4
		offset += padding
	}

	return msg, nil
}

// generateCredentials генерирует credentials для TURN аутентификации
func (tc *TURNClient) generateCredentials() (string, string, error) {
	// Простая реализация - в реальности нужна более сложная аутентификация
	username := tc.username
	password := tc.password

	// Генерируем nonce (упрощенно)
	nonce := fmt.Sprintf("%x", md5.Sum([]byte(username+password+time.Now().String())))

	return username, nonce, nil
}

// Close закрывает TURN клиент
func (tc *TURNClient) Close() error {
	if tc.conn != nil {
		return tc.conn.Close()
	}
	return nil
}
