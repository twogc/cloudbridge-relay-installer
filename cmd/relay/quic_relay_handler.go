package main

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/twogc/cloudbridge-relay-installer/internal/p2p"
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
	peerManager    *p2p.SimplePeerManager
	tokenValidator *p2p.TokenValidator
	logger         *zap.Logger
	ctx            context.Context
	cancel         context.CancelFunc
}

// NewQUICRelayHandler создает новый QUIC Relay Handler
func NewQUICRelayHandler(relayManager *p2p.RelayManager, peerManager *p2p.SimplePeerManager, tokenValidator *p2p.TokenValidator, logger *zap.Logger) *QUICRelayHandler {
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

// handleStream обрабатывает QUIC поток после AUTH.
// Control frames (single-peer liveness / local smoke):
//
//	PING[ payload]  →  PONG[ same payload]
//	ECHO[ payload]  →  ECHO[ same payload]
//
// Everything else is peer relay (TO:<peer>:...).
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

			msg := buffer[:n]
			if handled := qrh.handleControlFrame(stream, peerID, msg); handled {
				continue
			}

			// Пересылаем пакет
			err = qrh.relayManager.RelayPacket(peerID, msg)
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

// handleControlFrame answers PING/ECHO on the same stream. Returns true if consumed.
func (qrh *QUICRelayHandler) handleControlFrame(stream *quic.Stream, peerID string, msg []byte) bool {
	// Trim optional trailing newline for line-oriented smokes
	raw := msg
	if len(raw) > 0 && raw[len(raw)-1] == '\n' {
		raw = raw[:len(raw)-1]
	}
	s := string(raw)

	var reply string
	switch {
	case s == "PING" || len(s) >= 5 && s[:5] == "PING ":
		// PING → PONG; preserve optional payload after "PING"
		if s == "PING" {
			reply = "PONG"
		} else {
			reply = "PONG" + s[4:] // keep leading space + payload
		}
	case s == "ECHO" || len(s) >= 5 && s[:5] == "ECHO ":
		reply = s // echo back as-is
	default:
		return false
	}

	if _, err := (*stream).Write([]byte(reply)); err != nil {
		qrh.logger.Error("Failed to write control response",
			zap.String("peer_id", peerID),
			zap.Error(err))
		return true
	}
	qrh.logger.Debug("Control frame answered",
		zap.String("peer_id", peerID),
		zap.String("request", s),
		zap.String("reply", reply))
	return true
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
	// stream closed after AUTH_OK write

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
	var peerID, tenantID string

	if qrh.tokenValidator != nil {
		qrh.logger.Debug("Validating JWT token")
		result, validationErr := qrh.tokenValidator.ValidateP2PToken(tokenString, "", "")
		if validationErr != nil {
			qrh.logger.Error("JWT token validation failed", zap.Error(validationErr))
			return "", "", fmt.Errorf("invalid JWT token: %v", validationErr)
		}

		if !result.Valid {
			qrh.logger.Error("Token validation failed", zap.String("error", result.Error))
			return "", "", fmt.Errorf("invalid JWT token: %s", result.Error)
		}

		peerID = result.PeerID
		tenantID = result.TenantID
		qrh.logger.Debug("JWT token validated successfully")
	} else {
		// Fallback: создаем простые значения для тестирования
		peerID = "test-peer"
		tenantID = "test-tenant"
		qrh.logger.Warn("Token validator not available, using test values")
	}

	qrh.logger.Debug("Generated peer ID", zap.String("peer_id", peerID), zap.String("tenant_id", tenantID))

	// Fallback peer_id if claim missing (local smoke / partial JWTs)
	if peerID == "" {
		peerID = "anonymous-peer"
	}
	if tenantID == "" {
		tenantID = "default"
	}

	// Отправляем подтверждение. Keep stream open briefly so peer can read before FIN.
	if _, err = stream.Write([]byte("AUTH_OK")); err != nil {
		qrh.logger.Warn("Failed to send auth confirmation", zap.Error(err))
	} else {
		qrh.logger.Info("Auth confirmation AUTH_OK sent",
			zap.String("peer_id", peerID),
			zap.String("tenant_id", tenantID))
	}
	// Cancel read side only; leave write data deliverable. Full close after short delay.
	stream.CancelRead(0)
	go func() {
		time.Sleep(200 * time.Millisecond)
		_ = stream.Close()
	}()

	qrh.logger.Info("Authentication completed successfully",
		zap.String("peer_id", peerID),
		zap.String("tenant_id", tenantID))

	return peerID, tenantID, nil
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
