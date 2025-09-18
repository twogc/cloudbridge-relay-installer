package masque

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/quic-go/quic-go"
	"go.uber.org/zap"
)

// MASQUEProxy представляет MASQUE CONNECT-UDP прокси
type MASQUEProxy struct {
	quicServer    *quic.Server
	udpConn       *net.UDPConn
	activeTunnels map[string]*UDPTunnel
	mu            sync.RWMutex
	logger        *zap.Logger
	ctx           context.Context
	cancel        context.CancelFunc
}

// UDPTunnel представляет UDP туннель через MASQUE
type UDPTunnel struct {
	SessionID    string
	TargetAddr   *net.UDPAddr
	QUICStream   quic.Stream
	UDPConn      *net.UDPConn
	BytesRelayed int64
	StartTime    time.Time
	LastActivity time.Time
	IsActive     bool
}

// MASQUEConfig конфигурация MASQUE прокси
type MASQUEConfig struct {
	ListenAddr    string
	UDPListenAddr string
	CertFile      string
	KeyFile       string
	MaxTunnels    int
}

// NewMASQUEProxy создает новый MASQUE прокси
func NewMASQUEProxy(config *MASQUEConfig, logger *zap.Logger) (*MASQUEProxy, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Создаем UDP соединение для проксирования
	udpAddr, err := net.ResolveUDPAddr("udp", config.UDPListenAddr)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to resolve UDP address: %w", err)
	}

	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to listen on UDP: %w", err)
	}

	// Создаем QUIC сервер
	tlsConfig, err := createTLSConfig(config.CertFile, config.KeyFile)
	if err != nil {
		cancel()
		udpConn.Close()
		return nil, fmt.Errorf("failed to create TLS config: %w", err)
	}

	quicConfig := &quic.Config{
		MaxIncomingStreams:    1000,
		MaxIncomingUniStreams: 1000,
	}

	quicServer, err := quic.ListenAddr(config.ListenAddr, tlsConfig, quicConfig)
	if err != nil {
		cancel()
		udpConn.Close()
		return nil, fmt.Errorf("failed to create QUIC server: %w", err)
	}

	proxy := &MASQUEProxy{
		quicServer:    quicServer,
		udpConn:       udpConn,
		activeTunnels: make(map[string]*UDPTunnel),
		logger:        logger,
		ctx:           ctx,
		cancel:        cancel,
	}

	logger.Info("MASQUE proxy created",
		zap.String("quic_addr", config.ListenAddr),
		zap.String("udp_addr", config.UDPListenAddr))

	return proxy, nil
}

// Start запускает MASQUE прокси
func (mp *MASQUEProxy) Start() error {
	mp.logger.Info("Starting MASQUE proxy")

	// Запускаем обработку QUIC соединений
	go mp.handleQUICConnections()

	// Запускаем обработку UDP пакетов
	go mp.handleUDPPackets()

	// Запускаем очистку неактивных туннелей
	go mp.cleanupInactiveTunnels()

	mp.logger.Info("MASQUE proxy started")
	return nil
}

// handleQUICConnections обрабатывает входящие QUIC соединения
func (mp *MASQUEProxy) handleQUICConnections() {
	for {
		select {
		case <-mp.ctx.Done():
			return
		default:
			conn, err := mp.quicServer.Accept(mp.ctx)
			if err != nil {
				if mp.ctx.Err() != nil {
					return
				}
				mp.logger.Error("Failed to accept QUIC connection", zap.Error(err))
				continue
			}

			go mp.handleQUICConnection(conn)
		}
	}
}

// handleQUICConnection обрабатывает QUIC соединение
func (mp *MASQUEProxy) handleQUICConnection(conn quic.Connection) {
	defer conn.CloseWithError(0, "connection closed")

	mp.logger.Info("New QUIC connection",
		zap.String("remote_addr", conn.RemoteAddr().String()))

	for {
		select {
		case <-mp.ctx.Done():
			return
		default:
			stream, err := conn.AcceptStream(mp.ctx)
			if err != nil {
				if mp.ctx.Err() != nil {
					return
				}
				mp.logger.Error("Failed to accept stream", zap.Error(err))
				continue
			}

			go mp.handleStream(stream)
		}
	}
}

// handleStream обрабатывает QUIC поток
func (mp *MASQUEProxy) handleStream(stream quic.Stream) {
	defer stream.Close()

	// Читаем CONNECT-UDP запрос
	request, err := mp.readCONNECTUDPRequest(stream)
	if err != nil {
		mp.logger.Error("Failed to read CONNECT-UDP request", zap.Error(err))
		mp.sendErrorResponse(stream, "invalid_request")
		return
	}

	// Создаем UDP туннель
	tunnel, err := mp.createUDPTunnel(request, stream)
	if err != nil {
		mp.logger.Error("Failed to create UDP tunnel", zap.Error(err))
		mp.sendErrorResponse(stream, "tunnel_creation_failed")
		return
	}

	// Отправляем успешный ответ
	if err := mp.sendSuccessResponse(stream, tunnel); err != nil {
		mp.logger.Error("Failed to send success response", zap.Error(err))
		return
	}

	// Запускаем релеирование данных
	mp.startTunnelRelay(tunnel)
}

// readCONNECTUDPRequest читает CONNECT-UDP запрос
func (mp *MASQUEProxy) readCONNECTUDPRequest(stream quic.Stream) (*CONNECTUDPRequest, error) {
	// Читаем заголовок запроса
	buffer := make([]byte, 1024)
	n, err := stream.Read(buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to read request: %w", err)
	}

	// Парсим CONNECT-UDP запрос
	request, err := mp.parseCONNECTUDPRequest(buffer[:n])
	if err != nil {
		return nil, fmt.Errorf("failed to parse request: %w", err)
	}

	return request, nil
}

// CONNECTUDPRequest представляет CONNECT-UDP запрос
type CONNECTUDPRequest struct {
	TargetHost string
	TargetPort int
	SessionID  string
}

// parseCONNECTUDPRequest парсит CONNECT-UDP запрос
func (mp *MASQUEProxy) parseCONNECTUDPRequest(data []byte) (*CONNECTUDPRequest, error) {
	// Упрощенная реализация - в реальности нужно парсить HTTP/3 CONNECT-UDP
	// Здесь мы просто извлекаем информацию из данных

	// Для демонстрации используем простой формат: "host:port:session_id"
	_ = string(data)

	// В реальной реализации здесь должен быть парсинг HTTP/3 CONNECT-UDP
	// Пока используем упрощенный формат
	return &CONNECTUDPRequest{
		TargetHost: "127.0.0.1", // По умолчанию
		TargetPort: 8080,        // По умолчанию
		SessionID:  fmt.Sprintf("session_%d", time.Now().UnixNano()),
	}, nil
}

// createUDPTunnel создает UDP туннель
func (mp *MASQUEProxy) createUDPTunnel(request *CONNECTUDPRequest, stream quic.Stream) (*UDPTunnel, error) {
	// Разрешаем целевой адрес
	targetAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", request.TargetHost, request.TargetPort))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve target address: %w", err)
	}

	// Создаем UDP соединение к целевому адресу
	udpConn, err := net.DialUDP("udp", nil, targetAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create UDP connection: %w", err)
	}

	tunnel := &UDPTunnel{
		SessionID:    request.SessionID,
		TargetAddr:   targetAddr,
		QUICStream:   stream,
		UDPConn:      udpConn,
		StartTime:    time.Now(),
		LastActivity: time.Now(),
		IsActive:     true,
	}

	// Добавляем туннель в активные
	mp.mu.Lock()
	mp.activeTunnels[request.SessionID] = tunnel
	mp.mu.Unlock()

	mp.logger.Info("UDP tunnel created",
		zap.String("session_id", request.SessionID),
		zap.String("target_addr", targetAddr.String()))

	return tunnel, nil
}

// sendSuccessResponse отправляет успешный ответ
func (mp *MASQUEProxy) sendSuccessResponse(stream quic.Stream, tunnel *UDPTunnel) error {
	// Отправляем HTTP/3 200 OK ответ
	response := fmt.Sprintf("HTTP/3 200 OK\r\nSession-ID: %s\r\n\r\n", tunnel.SessionID)
	_, err := stream.Write([]byte(response))
	return err
}

// sendErrorResponse отправляет ответ об ошибке
func (mp *MASQUEProxy) sendErrorResponse(stream quic.Stream, errorCode string) error {
	response := fmt.Sprintf("HTTP/3 400 Bad Request\r\nError: %s\r\n\r\n", errorCode)
	_, err := stream.Write([]byte(response))
	return err
}

// startTunnelRelay запускает релеирование данных через туннель
func (mp *MASQUEProxy) startTunnelRelay(tunnel *UDPTunnel) {
	// Релеирование QUIC -> UDP
	go mp.relayQUICToUDP(tunnel)

	// Релеирование UDP -> QUIC
	go mp.relayUDPToQUIC(tunnel)
}

// relayQUICToUDP релеирует данные из QUIC потока в UDP
func (mp *MASQUEProxy) relayQUICToUDP(tunnel *UDPTunnel) {
	defer func() {
		tunnel.IsActive = false
		tunnel.UDPConn.Close()
	}()

	buffer := make([]byte, 4096)
	for tunnel.IsActive {
		select {
		case <-mp.ctx.Done():
			return
		default:
			// Читаем из QUIC потока
			n, err := tunnel.QUICStream.Read(buffer)
			if err != nil {
				mp.logger.Debug("QUIC stream read error",
					zap.String("session_id", tunnel.SessionID),
					zap.Error(err))
				return
			}

			// Отправляем в UDP
			_, err = tunnel.UDPConn.Write(buffer[:n])
			if err != nil {
				mp.logger.Error("UDP write error",
					zap.String("session_id", tunnel.SessionID),
					zap.Error(err))
				return
			}

			tunnel.BytesRelayed += int64(n)
			tunnel.LastActivity = time.Now()

			mp.logger.Debug("Relayed QUIC->UDP",
				zap.String("session_id", tunnel.SessionID),
				zap.Int("bytes", n))
		}
	}
}

// relayUDPToQUIC релеирует данные из UDP в QUIC поток
func (mp *MASQUEProxy) relayUDPToQUIC(tunnel *UDPTunnel) {
	defer func() {
		tunnel.IsActive = false
		tunnel.UDPConn.Close()
	}()

	buffer := make([]byte, 4096)
	for tunnel.IsActive {
		select {
		case <-mp.ctx.Done():
			return
		default:
			// Читаем из UDP
			tunnel.UDPConn.SetReadDeadline(time.Now().Add(1 * time.Second))
			n, err := tunnel.UDPConn.Read(buffer)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				mp.logger.Debug("UDP read error",
					zap.String("session_id", tunnel.SessionID),
					zap.Error(err))
				return
			}

			// Отправляем в QUIC поток
			_, err = tunnel.QUICStream.Write(buffer[:n])
			if err != nil {
				mp.logger.Error("QUIC stream write error",
					zap.String("session_id", tunnel.SessionID),
					zap.Error(err))
				return
			}

			tunnel.BytesRelayed += int64(n)
			tunnel.LastActivity = time.Now()

			mp.logger.Debug("Relayed UDP->QUIC",
				zap.String("session_id", tunnel.SessionID),
				zap.Int("bytes", n))
		}
	}
}

// handleUDPPackets обрабатывает входящие UDP пакеты
func (mp *MASQUEProxy) handleUDPPackets() {
	buffer := make([]byte, 4096)
	for {
		select {
		case <-mp.ctx.Done():
			return
		default:
			n, addr, err := mp.udpConn.ReadFromUDP(buffer)
			if err != nil {
				if mp.ctx.Err() != nil {
					return
				}
				mp.logger.Error("UDP read error", zap.Error(err))
				continue
			}

			// Обрабатываем UDP пакет
			go mp.handleUDPPacket(buffer[:n], addr)
		}
	}
}

// handleUDPPacket обрабатывает UDP пакет
func (mp *MASQUEProxy) handleUDPPacket(data []byte, fromAddr *net.UDPAddr) {
	// Здесь должна быть логика маршрутизации UDP пакетов к соответствующим туннелям
	// Для демонстрации просто логируем пакет
	mp.logger.Debug("Received UDP packet",
		zap.String("from_addr", fromAddr.String()),
		zap.Int("size", len(data)))
}

// cleanupInactiveTunnels очищает неактивные туннели
func (mp *MASQUEProxy) cleanupInactiveTunnels() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-mp.ctx.Done():
			return
		case <-ticker.C:
			mp.mu.Lock()
			now := time.Now()
			for sessionID, tunnel := range mp.activeTunnels {
				if now.Sub(tunnel.LastActivity) > 5*time.Minute {
					tunnel.IsActive = false
					tunnel.UDPConn.Close()
					delete(mp.activeTunnels, sessionID)

					mp.logger.Info("Cleaned up inactive tunnel",
						zap.String("session_id", sessionID))
				}
			}
			mp.mu.Unlock()
		}
	}
}

// GetActiveTunnels возвращает количество активных туннелей
func (mp *MASQUEProxy) GetActiveTunnels() int {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	return len(mp.activeTunnels)
}

// Close закрывает MASQUE прокси
func (mp *MASQUEProxy) Close() error {
	mp.cancel()

	// Закрываем все активные туннели
	mp.mu.Lock()
	for _, tunnel := range mp.activeTunnels {
		tunnel.IsActive = false
		tunnel.UDPConn.Close()
	}
	mp.activeTunnels = make(map[string]*UDPTunnel)
	mp.mu.Unlock()

	// Закрываем UDP соединение
	if mp.udpConn != nil {
		mp.udpConn.Close()
	}

	// Закрываем QUIC сервер
	if mp.quicServer != nil {
		mp.quicServer.Close()
	}

	mp.logger.Info("MASQUE proxy closed")
	return nil
}

// createTLSConfig создает TLS конфигурацию
func createTLSConfig(certFile, keyFile string) (*tls.Config, error) {
	// Упрощенная реализация - в реальности нужно загружать сертификаты
	return &tls.Config{
		InsecureSkipVerify: true, // Только для демонстрации
	}, nil
}
