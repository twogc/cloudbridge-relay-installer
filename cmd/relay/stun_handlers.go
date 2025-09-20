package main

import (
	"encoding/binary"
	"net"

	"go.uber.org/zap"
)

// STUN message types
const (
	STUNBindingRequest  = 0x0001
	STUNBindingResponse = 0x0101
	STUNBindingError    = 0x0111
)

// STUN attribute types
const (
	STUNAttrMappedAddress    = 0x0001
	STUNAttrXORMappedAddress = 0x0020
	STUNAttrSoftware         = 0x8022
	STUNAttrFingerprint      = 0x8028
)

// STUN header structure
type STUNHeader struct {
	MessageType   uint16
	MessageLength uint16
	MagicCookie   uint32
	TransactionID [12]byte
}

// STUN attribute structure
type STUNAttribute struct {
	Type   uint16
	Length uint16
	Value  []byte
}

// handleSTUNPacket handles incoming STUN packets
func (s *Server) handleSTUNPacket(conn *net.UDPConn, addr *net.UDPAddr, data []byte) {
	s.logger.Debug("Received STUN packet",
		zap.String("remote_addr", addr.String()),
		zap.Int("data_length", len(data)),
	)

	// Check minimum STUN packet size (20 bytes header)
	if len(data) < 20 {
		s.logger.Warn("STUN packet too short", zap.Int("length", len(data)))
		return
	}

	// Parse STUN header
	header := STUNHeader{
		MessageType:   binary.BigEndian.Uint16(data[0:2]),
		MessageLength: binary.BigEndian.Uint16(data[2:4]),
		MagicCookie:   binary.BigEndian.Uint32(data[4:8]),
	}
	copy(header.TransactionID[:], data[8:20])

	// Validate STUN magic cookie (0x2112A442)
	if header.MagicCookie != 0x2112A442 {
		s.logger.Warn("Invalid STUN magic cookie",
			zap.Uint32("received", header.MagicCookie),
			zap.Uint32("expected", 0x2112A442),
		)
		return
	}

	// Handle different STUN message types
	switch header.MessageType {
	case STUNBindingRequest:
		s.handleSTUNBindingRequest(conn, addr, &header, data[20:])
	case STUNBindingResponse:
		s.logger.Debug("Received STUN Binding Response", zap.String("from", addr.String()))
	case STUNBindingError:
		s.logger.Debug("Received STUN Binding Error", zap.String("from", addr.String()))
	default:
		s.logger.Warn("Unknown STUN message type",
			zap.Uint16("type", header.MessageType),
			zap.String("from", addr.String()),
		)
	}
}

// handleSTUNBindingRequest handles STUN Binding Request
func (s *Server) handleSTUNBindingRequest(conn *net.UDPConn, addr *net.UDPAddr, header *STUNHeader, attributes []byte) {
	s.logger.Debug("Handling STUN Binding Request",
		zap.String("from", addr.String()),
		zap.Uint16("message_length", header.MessageLength),
	)

	// Create STUN Binding Response
	response := s.createSTUNBindingResponse(header, addr)

	// Send response
	if _, err := conn.WriteToUDP(response, addr); err != nil {
		s.logger.Error("Failed to send STUN response",
			zap.String("to", addr.String()),
			zap.Error(err),
		)
		return
	}

	s.logger.Debug("STUN Binding Response sent",
		zap.String("to", addr.String()),
		zap.Int("response_length", len(response)),
	)
}

// createSTUNBindingResponse creates a STUN Binding Response
func (s *Server) createSTUNBindingResponse(header *STUNHeader, addr *net.UDPAddr) []byte {
	// Create XOR-MAPPED-ADDRESS attribute
	xorMappedAddr := s.createXORMappedAddress(addr)

	// Create SOFTWARE attribute
	software := []byte("CloudBridge-Relay/1.0")
	softwareAttr := make([]byte, 4+len(software))
	binary.BigEndian.PutUint16(softwareAttr[0:2], STUNAttrSoftware)
	if len(software) > 65535 {
		s.logger.Error("Software attribute too long", zap.Int("length", len(software)))
		return nil
	}
	// Safe conversion with bounds check
	softwareLen := len(software)
	if softwareLen < 0 || softwareLen > 65535 {
		s.logger.Error("Software length out of bounds", zap.Int("length", softwareLen))
		return nil
	}
	binary.BigEndian.PutUint16(softwareAttr[2:4], uint16(softwareLen))
	copy(softwareAttr[4:], software)

	// Calculate total message length
	totalLen := len(xorMappedAddr) + len(softwareAttr)
	if totalLen > 65535 {
		s.logger.Error("Message too long", zap.Int("length", totalLen))
		return nil
	}
	// Safe conversion with bounds check
	if totalLen < 0 || totalLen > 65535 {
		s.logger.Error("Message length out of bounds", zap.Int("length", totalLen))
		return nil
	}
	messageLength := uint16(totalLen)

	// Create response header
	response := make([]byte, 20+messageLength)
	binary.BigEndian.PutUint16(response[0:2], STUNBindingResponse)
	binary.BigEndian.PutUint16(response[2:4], messageLength)
	binary.BigEndian.PutUint32(response[4:8], 0x2112A442) // Magic cookie
	copy(response[8:20], header.TransactionID[:])

	// Add attributes
	offset := 20
	copy(response[offset:], xorMappedAddr)
	offset += len(xorMappedAddr)
	copy(response[offset:], softwareAttr)

	// Calculate and add fingerprint
	fingerprint := s.calculateFingerprint(response[:len(response)-8])
	fingerprintAttr := make([]byte, 8)
	binary.BigEndian.PutUint16(fingerprintAttr[0:2], STUNAttrFingerprint)
	binary.BigEndian.PutUint16(fingerprintAttr[2:4], 4)
	binary.BigEndian.PutUint32(fingerprintAttr[4:8], fingerprint)

	// Update message length to include fingerprint
	messageLength += 8
	binary.BigEndian.PutUint16(response[2:4], messageLength)

	// Append fingerprint
	response = append(response, fingerprintAttr...)

	return response
}

// createXORMappedAddress creates XOR-MAPPED-ADDRESS attribute
func (s *Server) createXORMappedAddress(addr *net.UDPAddr) []byte {
	// XOR-MAPPED-ADDRESS format:
	// 0                   1                   2                   3
	// 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |x x x x x x x x|    Family     |         X-Port                |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |                X-Address (Variable)
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

	var family uint16
	var address []byte

	if addr.IP.To4() != nil {
		// IPv4
		family = 0x01
		address = addr.IP.To4()
	} else {
		// IPv6
		family = 0x02
		address = addr.IP.To16()
	}

	// XOR with magic cookie
	magicCookie := uint32(0x2112A442)
	if addr.Port < 0 || addr.Port > 65535 {
		s.logger.Error("Invalid port", zap.Int("port", addr.Port))
		return nil
	}
	// Safe conversion with bounds check
	port := addr.Port
	if port < 0 || port > 65535 {
		s.logger.Error("Port out of bounds", zap.Int("port", port))
		return nil
	}
	xorPort := uint16(port) ^ uint16(magicCookie>>16) // #nosec G115 -- port is validated to be in range 0-65535

	// Create attribute
	attrLen := 2 + len(address)
	if attrLen > 65535 {
		s.logger.Error("Attribute too long", zap.Int("length", attrLen))
		return nil
	}
	attr := make([]byte, 4+attrLen)
	binary.BigEndian.PutUint16(attr[0:2], STUNAttrXORMappedAddress)
	// Safe conversion with bounds check
	if attrLen < 0 || attrLen > 65535 {
		s.logger.Error("Attribute length out of bounds", zap.Int("length", attrLen))
		return nil
	}
	binary.BigEndian.PutUint16(attr[2:4], uint16(attrLen))
	attr[4] = 0 // Reserved
	attr[5] = byte(family)
	binary.BigEndian.PutUint16(attr[6:8], xorPort)

	// XOR address with magic cookie
	for i, b := range address {
		if i < 4 && 8+i < len(attr) {
			attr[8+i] = b ^ byte(magicCookie>>(8*(3-i)))
		} else if 8+i < len(attr) {
			attr[8+i] = b
		}
	}

	return attr
}

// calculateFingerprint calculates STUN message fingerprint
func (s *Server) calculateFingerprint(data []byte) uint32 {
	// Simple CRC32-like calculation for fingerprint
	// In real implementation, this should be CRC32C
	var crc uint32 = 0xFFFFFFFF
	for _, b := range data {
		crc ^= uint32(b)
		for i := 0; i < 8; i++ {
			if crc&1 != 0 {
				crc = (crc >> 1) ^ 0xEDB88320
			} else {
				crc >>= 1
			}
		}
	}
	return crc ^ 0xFFFFFFFF
}
