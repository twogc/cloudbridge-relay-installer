package main

import (
	"fmt"
	"net"
	"time"

	"go.uber.org/zap"
)

// handleWireGuardPacket обрабатывает входящие WireGuard пакеты
func (s *Server) handleWireGuardPacket(conn *net.UDPConn, addr *net.UDPAddr, data []byte) {
	s.logger.Debug("Received WireGuard packet",
		zap.String("from", addr.String()),
		zap.Int("size", len(data)),
	)

	// Basic WireGuard packet validation
	if len(data) < 4 {
		s.logger.Warn("Invalid WireGuard packet size", zap.Int("size", len(data)))
		return
	}

	// Check if it's a WireGuard handshake packet
	packetType := data[0]
	switch packetType {
	case 1: // Initiation
		s.logger.Info("WireGuard initiation packet received",
			zap.String("from", addr.String()),
			zap.Int("size", len(data)),
			zap.String("hex_data", fmt.Sprintf("%x", data)),
		)
		// Генерируем Response пакет
		responsePacket := s.generateWireGuardResponse(data, addr)
		if responsePacket != nil {
			s.logger.Info("Sending WireGuard response",
				zap.String("to", addr.String()),
				zap.Int("response_size", len(responsePacket)),
			)
			_, err := conn.WriteToUDP(responsePacket, addr)
			if err != nil {
				s.logger.Error("Failed to send WireGuard response",
					zap.String("to", addr.String()),
					zap.Error(err),
				)
			} else {
				s.logger.Info("WireGuard response sent successfully",
					zap.String("to", addr.String()),
					zap.Int("bytes_sent", len(responsePacket)),
				)
			}
		} else {
			s.logger.Error("Failed to generate WireGuard response",
				zap.String("to", addr.String()),
			)
		}
	case 2: // Response
		s.logger.Info("WireGuard response packet received", zap.String("from", addr.String()))
		// Обрабатываем Response (завершение handshake)
		s.handleWireGuardResponse(data, addr)
	case 3: // Cookie Reply
		s.logger.Debug("WireGuard cookie reply received", zap.String("from", addr.String()))
		// Обрабатываем Cookie Reply
		s.handleWireGuardCookieReply(data, addr)
	case 4: // Transport Data
		s.logger.Debug("WireGuard transport data received", zap.String("from", addr.String()))
		// Обрабатываем транспортные данные
		s.handleWireGuardTransport(data, addr)
	default:
		s.logger.Warn("Unknown WireGuard packet type", zap.Int("type", int(packetType)))
	}
}

// generateWireGuardResponse генерирует Response пакет на Initiation
func (s *Server) generateWireGuardResponse(initiationData []byte, addr *net.UDPAddr) []byte {
	// Простая реализация Response пакета для тестирования
	// В реальной реализации здесь будет полноценный WireGuard handshake

	// Формат Response: [тип][reserved][sender_index][receiver_index][reserved]
	response := make([]byte, 6)
	response[0] = 2 // Response тип
	response[1] = 0 // Reserved

	// Копируем sender_index из initiation пакета
	if len(initiationData) >= 4 {
		response[2] = initiationData[2]
		s.logger.Debug("Copied sender_index from initiation",
			zap.Uint8("sender_index", response[2]))
	}

	// Генерируем receiver_index (в реальности это будет уникальный ID)
	response[3] = byte(time.Now().Unix() % 256)
	response[4] = 0 // Reserved
	response[5] = 0 // Reserved

	s.logger.Info("Generated WireGuard response packet",
		zap.String("to", addr.String()),
		zap.Int("size", len(response)),
		zap.Uint8("response_type", response[0]),
		zap.Uint8("sender_index", response[2]),
		zap.Uint8("receiver_index", response[3]),
		zap.String("hex_data", fmt.Sprintf("%x", response)),
	)

	return response
}

// handleWireGuardResponse обрабатывает Response пакет
func (s *Server) handleWireGuardResponse(data []byte, addr *net.UDPAddr) {
	s.logger.Info("Processing WireGuard response",
		zap.String("from", addr.String()),
		zap.Int("size", len(data)),
	)

	// В реальной реализации здесь будет завершение handshake
	// и установка защищенного соединения
}

// handleWireGuardCookieReply обрабатывает Cookie Reply
func (s *Server) handleWireGuardCookieReply(data []byte, addr *net.UDPAddr) {
	s.logger.Debug("Processing WireGuard cookie reply",
		zap.String("from", addr.String()),
		zap.Int("size", len(data)),
	)
}

// handleWireGuardTransport обрабатывает транспортные данные
func (s *Server) handleWireGuardTransport(data []byte, addr *net.UDPAddr) {
	s.logger.Debug("Processing WireGuard transport data",
		zap.String("from", addr.String()),
		zap.Int("size", len(data)),
	)

	// В реальной реализации здесь будет обработка зашифрованных данных
	// и маршрутизация к соответствующим пирам
}
