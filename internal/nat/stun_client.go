package nat

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"

	"go.uber.org/zap"
)

// STUNClient представляет простой STUN клиент
type STUNClient struct {
	conn       net.PacketConn
	serverAddr *net.UDPAddr
	logger     *zap.Logger
	timeout    time.Duration
}

// NewSTUNClient создает новый STUN клиент
func NewSTUNClient(serverAddr string, logger *zap.Logger) (*STUNClient, error) {
	// Resolve STUN server address
	udpAddr, err := net.ResolveUDPAddr("udp", serverAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve STUN server address: %w", err)
	}

	// Create UDP connection
	conn, err := net.ListenPacket("udp", ":0")
	if err != nil {
		return nil, fmt.Errorf("failed to create UDP connection: %w", err)
	}

	return &STUNClient{
		conn:       conn,
		serverAddr: udpAddr,
		logger:     logger,
		timeout:    5 * time.Second,
	}, nil
}

// GetExternalAddress получает внешний IP адрес через STUN
func (sc *STUNClient) GetExternalAddress() (*net.UDPAddr, error) {
	sc.logger.Debug("Performing STUN binding request",
		zap.String("server", sc.serverAddr.String()))

	start := time.Now()

	// Create simple STUN binding request
	request := sc.createBindingRequest()

	// Send request
	_, err := sc.conn.WriteTo(request, sc.serverAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to send STUN request: %w", err)
	}

	// Set read timeout
	sc.conn.SetReadDeadline(time.Now().Add(sc.timeout))

	// Read response
	buffer := make([]byte, 1024)
	n, _, err := sc.conn.ReadFrom(buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to read STUN response: %w", err)
	}

	// Parse response
	externalAddr, err := sc.parseBindingResponse(buffer[:n])
	if err != nil {
		return nil, fmt.Errorf("failed to parse STUN response: %w", err)
	}

	duration := time.Since(start)
	sc.logger.Info("STUN binding request successful",
		zap.String("server", sc.serverAddr.String()),
		zap.String("external_addr", externalAddr.String()),
		zap.Duration("duration", duration))

	return externalAddr, nil
}

// createBindingRequest создает STUN binding request
func (sc *STUNClient) createBindingRequest() []byte {
	// STUN message format:
	// 0                   1                   2                   3
	// 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |0 0|     STUN Message Type     |         Message Length        |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |                         Magic Cookie                          |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |                                                               |
	// |                     Transaction ID (96 bits)                 |
	// |                                                               |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

	request := make([]byte, 20) // STUN header is 20 bytes

	// Message Type: Binding Request (0x0001)
	binary.BigEndian.PutUint16(request[0:2], 0x0001)

	// Message Length: 0 (no attributes)
	binary.BigEndian.PutUint16(request[2:4], 0x0000)

	// Magic Cookie: 0x2112A442
	binary.BigEndian.PutUint32(request[4:8], 0x2112A442)

	// Transaction ID: 96 bits (12 bytes) - use current time
	now := time.Now().UnixNano()
	binary.BigEndian.PutUint64(request[8:16], uint64(now))
	binary.BigEndian.PutUint32(request[16:20], uint32(now&0xFFFFFFFF))

	return request
}

// parseBindingResponse парсит STUN binding response
func (sc *STUNClient) parseBindingResponse(data []byte) (*net.UDPAddr, error) {
	if len(data) < 20 {
		return nil, fmt.Errorf("STUN response too short")
	}

	// Check message type (should be Binding Success Response 0x0101)
	msgType := binary.BigEndian.Uint16(data[0:2])
	if msgType != 0x0101 {
		return nil, fmt.Errorf("unexpected STUN message type: 0x%04x", msgType)
	}

	// Get message length
	msgLength := binary.BigEndian.Uint16(data[2:4])
	if len(data) < int(20+msgLength) {
		return nil, fmt.Errorf("incomplete STUN response")
	}

	// Parse attributes
	offset := 20
	for offset < len(data) {
		if offset+4 > len(data) {
			break
		}

		attrType := binary.BigEndian.Uint16(data[offset : offset+2])
		attrLength := binary.BigEndian.Uint16(data[offset+2 : offset+4])
		offset += 4

		if offset+int(attrLength) > len(data) {
			break
		}

		// Look for XOR-MAPPED-ADDRESS (0x0020) or MAPPED-ADDRESS (0x0001)
		if attrType == 0x0020 || attrType == 0x0001 {
			return sc.parseAddressAttribute(data[offset:offset+int(attrLength)], attrType == 0x0020)
		}

		// Move to next attribute (with padding)
		offset += int(attrLength)
		// Attributes are padded to 4-byte boundaries
		if offset%4 != 0 {
			offset += 4 - (offset % 4)
		}
	}

	return nil, fmt.Errorf("no address attribute found in STUN response")
}

// parseAddressAttribute парсит address attribute
func (sc *STUNClient) parseAddressAttribute(data []byte, isXOR bool) (*net.UDPAddr, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("address attribute too short")
	}

	// Skip reserved byte
	family := data[1]
	port := binary.BigEndian.Uint16(data[2:4])

	var ip net.IP
	if family == 0x01 { // IPv4
		if len(data) < 8 {
			return nil, fmt.Errorf("IPv4 address attribute too short")
		}
		ip = net.IP(data[4:8])

		if isXOR {
			// XOR with magic cookie for XOR-MAPPED-ADDRESS
			magicCookie := uint32(0x2112A442)
			port ^= uint16(magicCookie >> 16)

			ipBytes := []byte(ip)
			ipInt := binary.BigEndian.Uint32(ipBytes)
			ipInt ^= magicCookie
			binary.BigEndian.PutUint32(ipBytes, ipInt)
			ip = net.IP(ipBytes)
		}
	} else {
		return nil, fmt.Errorf("unsupported address family: %d", family)
	}

	return &net.UDPAddr{
		IP:   ip,
		Port: int(port),
	}, nil
}

// DiscoverNATBehavior определяет тип NAT (упрощенная версия)
func (sc *STUNClient) DiscoverNATBehavior() (NATType, error) {
	sc.logger.Debug("Starting NAT behavior discovery")

	// Step 1: Basic binding request
	externalAddr1, err := sc.GetExternalAddress()
	if err != nil {
		return NATTypeUnknown, fmt.Errorf("failed to get external address: %w", err)
	}

	// Step 2: Binding request from different local port
	conn2, err := net.ListenPacket("udp", ":0")
	if err != nil {
		return NATTypeUnknown, fmt.Errorf("failed to create second UDP connection: %w", err)
	}
	defer conn2.Close()

	client2 := &STUNClient{
		conn:       conn2,
		serverAddr: sc.serverAddr,
		logger:     sc.logger,
		timeout:    sc.timeout,
	}

	externalAddr2, err := client2.GetExternalAddress()
	if err != nil {
		sc.logger.Warn("Failed to get second external address, assuming symmetric NAT")
		return NATTypeSymmetric, nil
	}

	// Analyze NAT behavior
	var natType NATType
	if externalAddr1.IP.Equal(externalAddr2.IP) {
		if externalAddr1.Port == externalAddr2.Port {
			natType = NATTypeCone
		} else {
			natType = NATTypeRestrictedCone
		}
	} else {
		natType = NATTypeSymmetric
	}

	sc.logger.Info("NAT behavior discovered",
		zap.String("nat_type", string(natType)),
		zap.String("external_addr1", externalAddr1.String()),
		zap.String("external_addr2", externalAddr2.String()))

	return natType, nil
}

// Close закрывает STUN клиент
func (sc *STUNClient) Close() error {
	if sc.conn != nil {
		return sc.conn.Close()
	}
	return nil
}

// NATType представляет тип NAT
type NATType string

const (
	NATTypeUnknown        NATType = "unknown"
	NATTypeCone           NATType = "cone"
	NATTypeRestrictedCone NATType = "restricted_cone"
	NATTypeSymmetric      NATType = "symmetric"
)





