package main

import (
	"context"
	"fmt"
	"io"

	"github.com/quic-go/quic-go"
	"github.com/twogc/cloudbridge-relay-installer/internal/p2p"
	"go.uber.org/zap"
)

// QUICRelayHandler обрабатывает QUIC соединения для P2P relay
type QUICRelayHandler struct {
	relayManager *p2p.RelayManager
	logger       *zap.Logger
	ctx          context.Context
	cancel       context.CancelFunc
}

// NewQUICRelayHandler создает новый QUIC Relay Handler
func NewQUICRelayHandler(relayManager *p2p.RelayManager, logger *zap.Logger) *QUICRelayHandler {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &QUICRelayHandler{
		relayManager: relayManager,
		logger:       logger,
		ctx:          ctx,
		cancel:       cancel,
	}
}

// Start запускает QUIC relay handler
func (qrh *QUICRelayHandler) Start(quicServer *quic.Listener) error {
	qrh.logger.Info("Starting QUIC Relay Handler")
	
	// Запускаем cleanup routine
	qrh.relayManager.StartCleanupRoutine()
	
	// Обрабатываем входящие QUIC соединения
	go qrh.handleQUICConnections(quicServer)
	
	return nil
}

// handleQUICConnections обрабатывает входящие QUIC соединения
func (qrh *QUICRelayHandler) handleQUICConnections(quicServer *quic.Listener) {
	for {
		select {
		case <-qrh.ctx.Done():
			return
		default:
			conn, err := quicServer.Accept(qrh.ctx)
			if err != nil {
				if qrh.ctx.Err() != nil {
					return
				}
				qrh.logger.Error("Failed to accept QUIC connection", zap.Error(err))
				continue
			}
			
			go qrh.handleQUICConnection(conn)
		}
	}
}

// handleQUICConnection обрабатывает QUIC соединение
func (qrh *QUICRelayHandler) handleQUICConnection(conn *quic.Conn) {
	defer conn.CloseWithError(0, "connection closed")
	
	clientAddr := conn.RemoteAddr().String()
	qrh.logger.Info("New QUIC connection",
		zap.String("client_addr", clientAddr))
	
	// Извлекаем peer ID из соединения (в реальности из JWT токена)
	peerID := qrh.extractPeerID(conn)
	if peerID == "" {
		qrh.logger.Error("Failed to extract peer ID from connection")
		return
	}
	
	// Извлекаем tenant ID (в реальности из JWT токена)
	tenantID := qrh.extractTenantID(conn)
	if tenantID == "" {
		qrh.logger.Error("Failed to extract tenant ID from connection")
		return
	}
	
	// Регистрируем QUIC соединение
	err := qrh.relayManager.RegisterQUICConnection(peerID, tenantID, conn)
	if err != nil {
		qrh.logger.Error("Failed to register QUIC connection",
			zap.String("peer_id", peerID),
			zap.String("tenant_id", tenantID),
			zap.Error(err))
		return
	}
	
	// Обрабатываем потоки
	qrh.handleStreams(conn, peerID)
}

// handleStreams обрабатывает QUIC потоки
func (qrh *QUICRelayHandler) handleStreams(conn *quic.Conn, peerID string) {
	for {
		select {
		case <-qrh.ctx.Done():
			return
		default:
			stream, err := conn.AcceptStream(qrh.ctx)
			if err != nil {
				if qrh.ctx.Err() != nil {
					return
				}
				qrh.logger.Error("Failed to accept stream",
					zap.String("peer_id", peerID),
					zap.Error(err))
				continue
			}
			
			go qrh.handleStream(*stream, peerID)
		}
	}
}

// handleStream обрабатывает QUIC поток
func (qrh *QUICRelayHandler) handleStream(stream quic.Stream, peerID string) {
	defer stream.Close()
	
	qrh.logger.Debug("New stream accepted",
		zap.String("peer_id", peerID),
		zap.Uint64("stream_id", uint64(stream.StreamID())))
	
	// Читаем данные из потока
	buffer := make([]byte, 4096)
	for {
		select {
		case <-qrh.ctx.Done():
			return
		default:
			n, err := stream.Read(buffer)
			if err != nil {
				if err == io.EOF {
					qrh.logger.Debug("Stream closed by peer",
						zap.String("peer_id", peerID))
					return
				}
				qrh.logger.Error("Failed to read from stream",
					zap.String("peer_id", peerID),
					zap.Error(err))
				return
			}
			
			// Пересылаем пакет
			err = qrh.relayManager.RelayPacket(peerID, buffer[:n])
			if err != nil {
				qrh.logger.Error("Failed to relay packet",
					zap.String("peer_id", peerID),
					zap.Error(err))
				continue
			}
			
			qrh.logger.Debug("Packet relayed",
				zap.String("peer_id", peerID),
				zap.Int("bytes", n))
		}
	}
}

// extractPeerID извлекает peer ID из соединения
func (qrh *QUICRelayHandler) extractPeerID(conn *quic.Conn) string {
	// В реальной реализации здесь должен быть парсинг JWT токена
	// Для демонстрации используем IP адрес
	clientAddr := conn.RemoteAddr().String()
	
	// Простая логика для демонстрации
	// В реальности peer ID должен приходить в JWT токене
	return fmt.Sprintf("peer-%s", clientAddr)
}

// extractTenantID извлекает tenant ID из соединения
func (qrh *QUICRelayHandler) extractTenantID(conn *quic.Conn) string {
	// В реальной реализации здесь должен быть парсинг JWT токена
	// Для демонстрации используем фиксированный tenant ID
	return "tenant-216420165"
}

// Stop останавливает QUIC Relay Handler
func (qrh *QUICRelayHandler) Stop() {
	qrh.logger.Info("Stopping QUIC Relay Handler")
	qrh.cancel()
}

// GetRelayManager возвращает Relay Manager
func (qrh *QUICRelayHandler) GetRelayManager() *p2p.RelayManager {
	return qrh.relayManager
}
