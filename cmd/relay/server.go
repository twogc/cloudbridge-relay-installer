package main

import (
	"go.uber.org/zap"
)

// StopServer останавливает сервер (дополнительные методы остановки)
func (s *Server) StopServer() {
	s.logger.Info("Stopping additional server components...")

	// Stop P2P API Server
	if s.p2pAPIServer != nil {
		if err := s.p2pAPIServer.Stop(); err != nil {
			s.logger.Error("Failed to stop P2P API Server", zap.Error(err))
		}
	}

	s.logger.Info("Additional server components stopped")
}
