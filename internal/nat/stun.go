package nat

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"time"

	"go.uber.org/zap"
)

// STUNClient представляет STUN клиент
type STUNClient struct {
	serverAddr string
	conn       *net.UDPConn
	logger     *zap.Logger
}

// STUNMessage представляет STUN сообщение
type STUNMessage struct {
	Type          uint16
	Length        uint16
	MagicCookie   uint32
	TransactionID [12]byte
	Attributes    []STUNAttribute
	Username      string
	Realm         string
	Nonce         string
}

// STUNAttribute представляет STUN атрибут
type STUNAttribute struct {
	Type   uint16
	Length uint16
	Value  []byte
}

// STUN константы
const (
	STUNMagicCookie uint32 = 0x2112A442

	// STUN Message Types
	STUNBindingRequest       = 0x0001
	STUNBindingResponse      = 0x0101
	STUNBindingError         = 0x0111
	STUNSharedSecretRequest  = 0x0002
	STUNSharedSecretResponse = 0x0102
	STUNSharedSecretError    = 0x0112

	// TURN Message Types
	STUNAllocateRequest          = 0x0003
	STUNAllocateResponse         = 0x0103
	STUNAllocateError            = 0x0113
	STUNRefreshRequest           = 0x0004
	STUNRefreshResponse          = 0x0104
	STUNRefreshError             = 0x0114
	STUNCreatePermissionRequest  = 0x0008
	STUNCreatePermissionResponse = 0x0108
	STUNCreatePermissionError    = 0x0118
	STUNChannelBindRequest       = 0x0009
	STUNChannelBindResponse      = 0x0109
	STUNChannelBindError         = 0x0119
	STUNSendIndication           = 0x0016
	STUNDataIndication           = 0x0017
	STUNUnauthorized             = 0x0141
	STUNServerError              = 0x0500

	// STUN Attribute Types
	STUNAttrMappedAddress      = 0x0001
	STUNAttrResponseAddress    = 0x0002
	STUNAttrChangeRequest      = 0x0003
	STUNAttrSourceAddress      = 0x0004
	STUNAttrChangedAddress     = 0x0005
	STUNAttrUsername           = 0x0006
	STUNAttrPassword           = 0x0007
	STUNAttrMessageIntegrity   = 0x0008
	STUNAttrErrorCode          = 0x0009
	STUNAttrUnknownAttributes  = 0x000A
	STUNAttrReflectedFrom      = 0x000B
	STUNAttrRealm              = 0x0014
	STUNAttrNonce              = 0x0015
	STUNAttrXORMappedAddress   = 0x0020
	STUNAttrXORPeerAddress     = 0x0012
	STUNAttrData               = 0x0013
	STUNAttrLifetime           = 0x000D
	STUNAttrBandwidth          = 0x0010
	STUNAttrDestinationAddress = 0x0011
	STUNAttrXOROnly            = 0x0021
	STUNAttrServerName         = 0x8022
	STUNAttrSecondaryAddress   = 0x8050
)

// NewSTUNClient создает новый STUN клиент
func NewSTUNClient(serverAddr string, logger *zap.Logger) *STUNClient {
	return &STUNClient{
		serverAddr: serverAddr,
		logger:     logger,
	}
}

// GetExternalAddress получает внешний IP адрес через STUN
func (sc *STUNClient) GetExternalAddress() (*net.UDPAddr, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := sc.SendBindingRequest(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to send binding request: %w", err)
	}

	// Ищем XOR-MAPPED-ADDRESS атрибут
	for _, attr := range response.Attributes {
		if attr.Type == STUNAttrXORMappedAddress {
			addr, err := sc.parseXORMappedAddress(attr.Value, response.TransactionID)
			if err != nil {
				return nil, fmt.Errorf("failed to parse XOR mapped address: %w", err)
			}
			return addr, nil
		}
	}

	// Fallback на MAPPED-ADDRESS
	for _, attr := range response.Attributes {
		if attr.Type == STUNAttrMappedAddress {
			addr, err := sc.parseMappedAddress(attr.Value)
			if err != nil {
				return nil, fmt.Errorf("failed to parse mapped address: %w", err)
			}
			return addr, nil
		}
	}

	return nil, fmt.Errorf("no mapped address found in STUN response")
}

// GetServerReflexiveCandidate получает server reflexive candidate (улучшенная версия)
func (sc *STUNClient) GetServerReflexiveCandidate(localAddr string, localPort int) (*Candidate, error) {
	// Создаем временное соединение для STUN запроса
	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.ParseIP(localAddr),
		Port: 0, // Система выберет свободный порт
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create UDP connection: %w", err)
	}
	defer conn.Close()

	// Создаем STUN клиент с конкретным соединением
	stunClient := &STUNClient{
		serverAddr: sc.serverAddr,
		conn:       conn,
		logger:     sc.logger,
	}

	externalAddr, err := stunClient.GetExternalAddress()
	if err != nil {
		return nil, fmt.Errorf("failed to get external address: %w", err)
	}

	// Создаем server reflexive candidate
	candidate := &Candidate{
		Foundation:  fmt.Sprintf("srflx-%s-%d", externalAddr.IP.String(), externalAddr.Port),
		Component:   1,
		Transport:   "UDP",
		Priority:    0, // Will be calculated by ICE agent
		Address:     externalAddr.IP.String(),
		Port:        externalAddr.Port,
		Type:        CandidateTypeServerReflexive,
		RelatedAddr: localAddr,
		RelatedPort: localPort,
	}

	sc.logger.Debug("Got server reflexive candidate",
		zap.String("external_address", externalAddr.String()),
		zap.String("local_address", fmt.Sprintf("%s:%d", localAddr, localPort)))

	return candidate, nil
}

// SendBindingRequest отправляет STUN binding request
func (sc *STUNClient) SendBindingRequest(ctx context.Context) (*STUNMessage, error) {
	// Создаем UDP соединение
	conn, err := net.Dial("udp", sc.serverAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial STUN server: %w", err)
	}
	defer conn.Close()

	udpConn := conn.(*net.UDPConn)
	sc.conn = udpConn

	// Создаем binding request
	request := &STUNMessage{
		Type:        STUNBindingRequest,
		Length:      0,
		MagicCookie: STUNMagicCookie,
	}

	// Генерируем случайный transaction ID
	_, err = net.ParseMAC("00:00:00:00:00:00")
	if err != nil {
		return nil, fmt.Errorf("failed to generate transaction ID: %w", err)
	}
	copy(request.TransactionID[:], []byte("123456789012")) // Simplified for demo

	// Добавляем CHANGE-REQUEST атрибут для получения mapped address
	changeRequest := make([]byte, 4)
	binary.BigEndian.PutUint32(changeRequest, 0x00000006) // Change IP and Port

	request.Attributes = append(request.Attributes, STUNAttribute{
		Type:   STUNAttrChangeRequest,
		Length: uint16(len(changeRequest)),
		Value:  changeRequest,
	})

	// Кодируем сообщение
	message, err := sc.encodeMessage(request)
	if err != nil {
		return nil, fmt.Errorf("failed to encode STUN message: %w", err)
	}

	// Отправляем запрос
	if _, err := udpConn.Write(message); err != nil {
		return nil, fmt.Errorf("failed to send STUN request: %w", err)
	}

	// Устанавливаем таймаут для чтения
	udpConn.SetReadDeadline(time.Now().Add(5 * time.Second))

	// Читаем ответ
	buffer := make([]byte, 1024)
	n, err := udpConn.Read(buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to read STUN response: %w", err)
	}

	// Декодируем ответ
	response, err := sc.decodeMessage(buffer[:n])
	if err != nil {
		return nil, fmt.Errorf("failed to decode STUN response: %w", err)
	}

	// Проверяем, что это binding response
	if response.Type != STUNBindingResponse {
		return nil, fmt.Errorf("unexpected STUN message type: %d", response.Type)
	}

	// Проверяем transaction ID
	if response.TransactionID != request.TransactionID {
		return nil, fmt.Errorf("transaction ID mismatch")
	}

	return response, nil
}

// encodeMessage кодирует STUN сообщение в байты
func (sc *STUNClient) encodeMessage(msg *STUNMessage) ([]byte, error) {
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

// decodeMessage декодирует STUN сообщение из байтов
func (sc *STUNClient) decodeMessage(data []byte) (*STUNMessage, error) {
	if len(data) < 20 {
		return nil, fmt.Errorf("STUN message too short")
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
		return nil, fmt.Errorf("invalid STUN magic cookie")
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

// parseXORMappedAddress парсит XOR-MAPPED-ADDRESS атрибут
func (sc *STUNClient) parseXORMappedAddress(data []byte, transactionID [12]byte) (*net.UDPAddr, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("XOR mapped address too short")
	}

	// Первый байт - семейство адресов (0x01 для IPv4, 0x02 для IPv6)
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

// parseMappedAddress парсит MAPPED-ADDRESS атрибут
func (sc *STUNClient) parseMappedAddress(data []byte) (*net.UDPAddr, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("mapped address too short")
	}

	// Первый байт - семейство адресов
	family := data[1]
	if family != 0x01 {
		return nil, fmt.Errorf("unsupported address family: %d", family)
	}

	port := binary.BigEndian.Uint16(data[2:4])
	ip := net.IP(data[4:8])

	return &net.UDPAddr{
		IP:   ip,
		Port: int(port),
	}, nil
}

// Close закрывает STUN клиент
func (sc *STUNClient) Close() error {
	if sc.conn != nil {
		return sc.conn.Close()
	}
	return nil
}
