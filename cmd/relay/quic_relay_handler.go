package main

import (
	"context"
	"fmt"
	"io"

	"github.com/quic-go/quic-go"
	"github.com/twogc/cloudbridge-relay-installer/internal/api"
	"github.com/twogc/cloudbridge-relay-installer/internal/p2p"
	"github.com/twogc/cloudbridge-relay-installer/internal/token"
	"go.uber.org/zap"
)

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// QUICRelayHandler обрабатывает QUIC соединения для P2P relay
type QUICRelayHandler struct {
	relayManager   *p2p.RelayManager
	peerManager    *api.PeerManager
	tokenValidator *token.Validator
	logger         *zap.Logger
	ctx            context.Context
	cancel         context.CancelFunc
}

// NewQUICRelayHandler создает новый QUIC Relay Handler
func NewQUICRelayHandler(relayManager *p2p.RelayManager, peerManager *api.PeerManager, tokenValidator *token.Validator, logger *zap.Logger) *QUICRelayHandler {
	ctx, cancel := context.WithCancel(context.Background())

	return &QUICRelayHandler{
		relayManager:   relayManager,
		peerManager:    peerManager,
		tokenValidator: tokenValidator,
		logger:         logger,
		ctx:            ctx,
		cancel:         cancel,
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
	qrh.logger.Info("Starting QUIC connection handler")

	for {
		select {
		case <-qrh.ctx.Done():
			qrh.logger.Info("QUIC connection handler stopped")
			return
		default:
			qrh.logger.Debug("Waiting for QUIC connection...")
			conn, err := quicServer.Accept(qrh.ctx)
			if err != nil {
				if qrh.ctx.Err() != nil {
					qrh.logger.Info("Context cancelled, stopping QUIC connection handler")
					return
				}
				qrh.logger.Error("Failed to accept QUIC connection", zap.Error(err))
				continue
			}

			qrh.logger.Info("QUIC connection attempt received",
				zap.String("remote_addr", conn.RemoteAddr().String()),
				zap.String("local_addr", conn.LocalAddr().String()))

			qrh.logger.Info("QUIC connection accepted",
				zap.String("remote_addr", conn.RemoteAddr().String()),
				zap.String("local_addr", conn.LocalAddr().String()))

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

	// Аутентификация через первый поток
	peerID, tenantID, err := qrh.authenticateConnection(conn)
	if err != nil {
		qrh.logger.Error("Failed to authenticate QUIC connection",
			zap.String("client_addr", clientAddr),
			zap.Error(err))
		return
	}

	qrh.logger.Info("QUIC connection authenticated",
		zap.String("peer_id", peerID),
		zap.String("tenant_id", tenantID),
		zap.String("client_addr", clientAddr))

	// Регистрируем QUIC соединение
	err = qrh.relayManager.RegisterQUICConnection(peerID, tenantID, conn)
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

			go qrh.handleStream(stream, peerID)
		}
	}
}

// handleStream обрабатывает QUIC поток
func (qrh *QUICRelayHandler) handleStream(stream *quic.Stream, peerID string) {
	defer (*stream).Close()

	streamID := (*stream).StreamID()
	if streamID < 0 {
		qrh.logger.Error("Invalid stream ID", zap.Int64("stream_id", int64(streamID)))
		return
	}
	// Safe conversion for logging
	var streamIDUint uint64
	if streamID >= 0 {
		streamIDUint = uint64(streamID) // #nosec G115 -- streamID is validated to be non-negative
	}
	qrh.logger.Debug("New stream accepted",
		zap.String("peer_id", peerID),
		zap.Uint64("stream_id", streamIDUint))

	// Читаем данные из потока
	buffer := make([]byte, 4096)
	for {
		select {
		case <-qrh.ctx.Done():
			return
		default:
			n, err := (*stream).Read(buffer)
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

// authenticateConnection аутентифицирует QUIC соединение через JWT токен
func (qrh *QUICRelayHandler) authenticateConnection(conn *quic.Conn) (string, string, error) {
	qrh.logger.Debug("Starting authentication process")

	// Принимаем первый поток для аутентификации
	stream, err := conn.AcceptStream(qrh.ctx)
	if err != nil {
		qrh.logger.Error("Failed to accept auth stream", zap.Error(err))
		return "", "", fmt.Errorf("failed to accept auth stream: %v", err)
	}
	defer stream.Close()

	qrh.logger.Debug("Auth stream accepted, waiting for token")

	// Читаем JWT токен из потока
	buffer := make([]byte, 4096)
	n, err := stream.Read(buffer)
	if err != nil {
		qrh.logger.Error("Failed to read auth token", zap.Error(err))
		return "", "", fmt.Errorf("failed to read auth token: %v", err)
	}

	// Парсим токен (ожидаем формат: "AUTH <jwt_token>")
	authData := string(buffer[:n])
	qrh.logger.Debug("Received auth data", zap.String("data", authData))

	if len(authData) < 6 || authData[:5] != "AUTH " {
		qrh.logger.Error("Invalid auth format", zap.String("received", authData))
		return "", "", fmt.Errorf("invalid auth format, expected 'AUTH <token>'")
	}

	tokenString := authData[5:]
	qrh.logger.Debug("Extracted token", zap.String("token", tokenString[:min(20, len(tokenString))]+"..."))

	// Валидируем JWT токен (если tokenValidator доступен)
	var claims *token.P2PClaims
	var validationErr error

	if qrh.tokenValidator != nil {
		qrh.logger.Debug("Validating JWT token")
		claims, validationErr = qrh.tokenValidator.ValidateP2PToken(tokenString)
		if validationErr != nil {
			qrh.logger.Error("JWT token validation failed", zap.Error(validationErr))
			return "", "", fmt.Errorf("invalid JWT token: %v", validationErr)
		}
		qrh.logger.Debug("JWT token validated successfully")
	} else {
		// Fallback: создаем простые claims для тестирования
		claims = &token.P2PClaims{
			ServerID: "test-server",
			TenantID: "test-tenant",
		}
		qrh.logger.Warn("Token validator not available, using test claims")
	}

	// Генерируем peer ID из server_id
	peerID := fmt.Sprintf("peer_%s", claims.ServerID)
	qrh.logger.Debug("Generated peer ID", zap.String("peer_id", peerID), zap.String("tenant_id", claims.TenantID))

	// Отправляем подтверждение
	_, err = stream.Write([]byte("AUTH_OK"))
	if err != nil {
		qrh.logger.Warn("Failed to send auth confirmation", zap.Error(err))
	} else {
		qrh.logger.Debug("Auth confirmation sent")
	}

	qrh.logger.Info("Authentication completed successfully",
		zap.String("peer_id", peerID),
		zap.String("tenant_id", claims.TenantID))

	return peerID, claims.TenantID, nil
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
