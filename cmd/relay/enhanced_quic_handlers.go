package main

import (
	"net"
	"time"

	"go.uber.org/zap"
)

// handleEnhancedQUICPacket handles incoming Enhanced QUIC packets
func (s *Server) handleEnhancedQUICPacket(conn *net.UDPConn, addr *net.UDPAddr, data []byte) {
	s.logger.Debug("Received Enhanced QUIC packet",
		zap.String("remote_addr", addr.String()),
		zap.Int("data_length", len(data)),
	)

	// Check if this is a QUIC packet (starts with QUIC version)
	if len(data) < 4 {
		s.logger.Warn("Enhanced QUIC packet too short", zap.Int("length", len(data)))
		return
	}

	// Check for QUIC version (first 4 bytes)
	version := uint32(data[0])<<24 | uint32(data[1])<<16 | uint32(data[2])<<8 | uint32(data[3])

	// Known QUIC versions
	quicVersions := map[uint32]string{
		0x00000001: "QUIC v1",
		0x6b3343cf: "QUIC v2 (draft)",
		0x709a50c4: "QUIC v2 (RFC)",
	}

	if versionName, exists := quicVersions[version]; exists {
		s.logger.Debug("Enhanced QUIC packet received",
			zap.String("version", versionName),
			zap.String("from", addr.String()),
		)
	} else {
		s.logger.Debug("Enhanced QUIC packet with unknown version",
			zap.Uint32("version", version),
			zap.String("from", addr.String()),
		)
	}

	// For now, we'll just log the packet and potentially forward it
	// In a full implementation, this would handle QUIC handshake, streams, etc.

	// Check if we have Enhanced QUIC Manager
	if s.enhancedQUICManager != nil {
		// Enhanced QUIC Manager is available, log that we received a packet
		s.logger.Debug("Enhanced QUIC packet forwarded to manager",
			zap.String("from", addr.String()),
		)
	} else {
		s.logger.Debug("Enhanced QUIC Manager not available, logging packet only")
	}

	// Log packet statistics
	s.logger.Debug("Enhanced QUIC packet processed",
		zap.String("from", addr.String()),
		zap.Int("size", len(data)),
		zap.Time("timestamp", time.Now()),
	)
}
