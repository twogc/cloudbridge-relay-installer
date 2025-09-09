package main

import (
	"net"

	"go.uber.org/zap"
)

// handleQUICPacket обрабатывает входящие QUIC пакеты
func (s *Server) handleQUICPacket(conn *net.UDPConn, addr *net.UDPAddr, data []byte) {
	s.logger.Debug("Received QUIC packet",
		zap.String("from", addr.String()),
		zap.Int("size", len(data)),
	)

	// Basic QUIC packet validation
	if len(data) < 1 {
		s.logger.Warn("Invalid QUIC packet size", zap.Int("size", len(data)))
		return
	}

	// Check QUIC packet type (first byte)
	packetType := data[0] & 0x30 >> 4 // Extract packet type from first byte
	switch packetType {
	case 0: // Initial
		s.logger.Info("QUIC initial packet received", zap.String("from", addr.String()))
	case 1: // 0-RTT
		s.logger.Info("QUIC 0-RTT packet received", zap.String("from", addr.String()))
	case 2: // Handshake
		s.logger.Info("QUIC handshake packet received", zap.String("from", addr.String()))
	case 3: // Retry
		s.logger.Debug("QUIC retry packet received", zap.String("from", addr.String()))
	default:
		s.logger.Debug("QUIC short header packet received", zap.String("from", addr.String()))
	}

	// Echo back for testing (remove in production)
	_, err := conn.WriteToUDP(data, addr)
	if err != nil {
		s.logger.Error("Failed to send QUIC response", zap.Error(err))
	}
}
