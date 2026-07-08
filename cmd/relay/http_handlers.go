package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	"go.uber.org/zap"
)

// handleHTTPRequest обрабатывает HTTP запросы
func (s *Server) handleHTTPRequest(conn net.Conn, firstLine string, reader *bufio.Reader) error {
	// Read headers until empty line
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			// If we can't read more, just proceed with what we have
			break
		}
		if line == "\r\n" || line == "\n" {
			break // End of headers
		}
	}

	// Skip body for now - we don't need it for health checks and 404 responses

	// Parse the first line to get method and path
	parts := strings.Fields(firstLine)
	if len(parts) < 2 {
		// If we can't parse the request line, return 400 Bad Request
		errorResponse := `{"error": "Invalid request line"}`
		response := "HTTP/1.1 400 Bad Request\r\n"
		response += "Content-Type: application/json\r\n"
		response += "Connection: close\r\n"
		response += fmt.Sprintf("Content-Length: %d\r\n", len(errorResponse))
		response += "\r\n"
		response += errorResponse

		_, writeErr := conn.Write([]byte(response))
		if writeErr != nil {
			return writeErr
		}

		// Ensure all data is written before closing
		if tcpConn, ok := conn.(*net.TCPConn); ok {
			if err := tcpConn.CloseWrite(); err != nil {
				s.logger.Warn("Failed to close TCP write", zap.Error(err))
			}
		}
		return nil
	}

	method := parts[0]
	path := parts[1]

	s.logger.Debug("HTTP request", zap.String("method", method), zap.String("path", path))

	// Handle OPTIONS request (CORS preflight)
	if method == "OPTIONS" {
		response := "HTTP/1.1 200 OK\r\n"
		response += "Access-Control-Allow-Origin: *\r\n"
		response += "Access-Control-Allow-Methods: GET, OPTIONS\r\n"
		response += "Access-Control-Allow-Headers: Content-Type\r\n"
		response += "Connection: close\r\n"
		response += "Content-Length: 0\r\n"
		response += "\r\n"

		_, writeErr := conn.Write([]byte(response))
		if writeErr != nil {
			return writeErr
		}

		// Ensure all data is written before closing
		if tcpConn, ok := conn.(*net.TCPConn); ok {
			if err := tcpConn.CloseWrite(); err != nil {
				s.logger.Warn("Failed to close TCP write", zap.Error(err))
			}
		}
		return nil
	}

	// Handle health endpoint
	if path == "/health" {
		return s.handleHealthHTTP(conn)
	}

	// Return 404 for unknown paths
	errorResponse := `{"error": "Not Found"}`
	response := "HTTP/1.1 404 Not Found\r\n"
	response += "Content-Type: application/json\r\n"
	response += "Access-Control-Allow-Origin: *\r\n"
	response += "Connection: close\r\n"
	response += fmt.Sprintf("Content-Length: %d\r\n", len(errorResponse))
	response += "\r\n"
	response += errorResponse

	_, writeErr := conn.Write([]byte(response))
	if writeErr != nil {
		return writeErr
	}

	// Ensure all data is written before closing
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		if err := tcpConn.CloseWrite(); err != nil {
			s.logger.Error("Failed to close TCP write", zap.Error(err))
		}
	}
	return nil
}

// handleHealthHTTP обрабатывает health check запросы
func (s *Server) handleHealthHTTP(conn net.Conn) error {
	status := map[string]interface{}{
		"status":    "ok",
		"version":   s.config.API.Version,
		"timestamp": time.Now().Format(time.RFC3339),
		"service":   "cloudbridge-relay",
		"uptime":    time.Since(s.startTime).String(),
		"components": map[string]interface{}{
			"tcp_server":        "running",
			"metrics":           s.metricsServer != nil,
			"health_checker":    s.healthChecker != nil,
			"circuit_breaker":   s.circuitBreaker != nil,
			"enhanced_quic":     s.enhancedEnabled,
			"tenant_manager":    s.tenantManager != nil,
			"load_balancer":     s.loadBalancer != nil,
			"rate_limiter":    s.ipRateLimit != nil,
		},
	}

	// Добавляем Enhanced QUIC статистику
	if s.enhancedEnabled && s.enhancedQUICManager != nil {
		status["enhanced_quic_stats"] = s.enhancedQUICManager.GetStats()
	}

	// Добавляем статистику тенантов
	if s.tenantManager != nil {
		status["tenant_stats"] = map[string]interface{}{
			"total_tenants": s.tenantManager.GetTenantCount(),
			"enhanced_mode": s.tenantManager.IsEnhancedEnabled(),
		}
	}

	jsonData, err := json.Marshal(status)
	if err != nil {
		return err
	}

	response := "HTTP/1.1 200 OK\r\n"
	response += "Content-Type: application/json\r\n"
	response += "Access-Control-Allow-Origin: *\r\n"
	response += "Access-Control-Allow-Methods: GET, OPTIONS\r\n"
	response += "Access-Control-Allow-Headers: Content-Type\r\n"
	response += "Connection: close\r\n"
	response += fmt.Sprintf("Content-Length: %d\r\n", len(jsonData))
	response += "\r\n"
	response += string(jsonData)

	_, err = conn.Write([]byte(response))
	if err != nil {
		return err
	}

	// Ensure all data is written before closing
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		if err := tcpConn.CloseWrite(); err != nil {
			s.logger.Error("Failed to close TCP write", zap.Error(err))
		}
	}
	return nil
}

// writeJSON записывает JSON ответ
func writeJSON(writer *bufio.Writer, msg interface{}) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if _, err := writer.Write(data); err != nil {
		return err
	}
	return writer.Flush()
}
