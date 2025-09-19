package api

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ICECandidate представляет ICE candidate для NAT traversal
type ICECandidate struct {
	Foundation  string `json:"foundation"`
	Component   int    `json:"component"`
	Transport   string `json:"transport"`
	Priority    uint32 `json:"priority"`
	Address     string `json:"address"`
	Port        int    `json:"port"`
	Type        string `json:"type"` // "host", "srflx", "prflx", "relay"
	RelatedAddr string `json:"related_address,omitempty"`
	RelatedPort int    `json:"related_port,omitempty"`
	Username    string `json:"username,omitempty"`
	Password    string `json:"password,omitempty"`
}

// ICESignalingRequest запрос для ICE signaling
type ICESignalingRequest struct {
	SessionID  string          `json:"session_id" binding:"required"`
	PeerID     string          `json:"peer_id" binding:"required"`
	Candidates []*ICECandidate `json:"candidates" binding:"required"`
	Offer      string          `json:"offer,omitempty"`
	Answer     string          `json:"answer,omitempty"`
}

// ICESignalingResponse ответ на ICE signaling
type ICESignalingResponse struct {
	Success    bool            `json:"success"`
	SessionID  string          `json:"session_id"`
	Candidates []*ICECandidate `json:"candidates,omitempty"`
	Offer      string          `json:"offer,omitempty"`
	Answer     string          `json:"answer,omitempty"`
}

// ICEManager управляет ICE signaling через API
type ICEManager struct {
	sessions map[string]*ICESession
	mu       sync.RWMutex
	logger   *zap.Logger
}

// ICESession представляет ICE сессию
type ICESession struct {
	SessionID       string
	Peer1ID         string
	Peer2ID         string
	Peer1Candidates []*ICECandidate
	Peer2Candidates []*ICECandidate
	Offer           string
	Answer          string
	State           ICESessionState
	CreatedAt       time.Time
	LastActivity    time.Time
}

// ICESessionState состояние ICE сессии
type ICESessionState int

const (
	ICESessionStateNew ICESessionState = iota
	ICESessionStateOfferReceived
	ICESessionStateAnswerReceived
	ICESessionStateCompleted
	ICESessionStateFailed
)

// NewICEManager создает новый ICE Manager
func NewICEManager(logger *zap.Logger) *ICEManager {
	return &ICEManager{
		sessions: make(map[string]*ICESession),
		logger:   logger,
	}
}

// ExchangeCandidates обменивается ICE candidates
func (im *ICEManager) ExchangeCandidates(c *gin.Context) {
	var req ICESignalingRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		im.logger.Error("Failed to bind ICE signaling request", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request format",
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	// Валидация JWT токена
	claims, err := im.validateToken(c)
	if err != nil {
		im.logger.Error("Failed to validate token", zap.Error(err))
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Invalid token",
			Code:    http.StatusUnauthorized,
			Message: err.Error(),
		})
		return
	}

	im.mu.Lock()
	defer im.mu.Unlock()

	// Получаем или создаем сессию
	session, exists := im.sessions[req.SessionID]
	if !exists {
		session = &ICESession{
			SessionID:    req.SessionID,
			Peer1ID:      claims.Subject,
			State:        ICESessionStateNew,
			CreatedAt:    time.Now(),
			LastActivity: time.Now(),
		}
		im.sessions[req.SessionID] = session
	}

	// Обновляем candidates для пира
	if session.Peer1ID == claims.Subject {
		session.Peer1Candidates = req.Candidates
	} else if session.Peer2ID == "" {
		session.Peer2ID = claims.Subject
		session.Peer2Candidates = req.Candidates
	} else if session.Peer2ID == claims.Subject {
		session.Peer2Candidates = req.Candidates
	} else {
		c.JSON(http.StatusForbidden, ErrorResponse{
			Error:   "Access denied",
			Code:    http.StatusForbidden,
			Message: "You don't have access to this session",
		})
		return
	}

	session.LastActivity = time.Now()

	// Если у нас есть candidates от обоих пиров, обновляем состояние
	if len(session.Peer1Candidates) > 0 && len(session.Peer2Candidates) > 0 {
		session.State = ICESessionStateCompleted
	}

	// Определяем, какие candidates отправить в ответе
	var responseCandidates []*ICECandidate
	if session.Peer1ID == claims.Subject {
		responseCandidates = session.Peer2Candidates
	} else {
		responseCandidates = session.Peer1Candidates
	}

	response := ICESignalingResponse{
		Success:    true,
		SessionID:  req.SessionID,
		Candidates: responseCandidates,
	}

	im.logger.Info("ICE candidates exchanged",
		zap.String("session_id", req.SessionID),
		zap.String("peer_id", claims.Subject),
		zap.Int("candidates_count", len(req.Candidates)))

	c.JSON(http.StatusOK, response)
}

// HandleOffer обрабатывает WebRTC offer
func (im *ICEManager) HandleOffer(c *gin.Context) {
	var req ICESignalingRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		im.logger.Error("Failed to bind ICE offer request", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request format",
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	// Валидация JWT токена
	claims, err := im.validateToken(c)
	if err != nil {
		im.logger.Error("Failed to validate token", zap.Error(err))
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Invalid token",
			Code:    http.StatusUnauthorized,
			Message: err.Error(),
		})
		return
	}

	im.mu.Lock()
	defer im.mu.Unlock()

	// Получаем или создаем сессию
	session, exists := im.sessions[req.SessionID]
	if !exists {
		session = &ICESession{
			SessionID:    req.SessionID,
			Peer1ID:      claims.Subject,
			State:        ICESessionStateNew,
			CreatedAt:    time.Now(),
			LastActivity: time.Now(),
		}
		im.sessions[req.SessionID] = session
	}

	// Проверяем права доступа
	if session.Peer1ID != claims.Subject && session.Peer2ID != claims.Subject {
		c.JSON(http.StatusForbidden, ErrorResponse{
			Error:   "Access denied",
			Code:    http.StatusForbidden,
			Message: "You don't have access to this session",
		})
		return
	}

	// Сохраняем offer
	session.Offer = req.Offer
	session.State = ICESessionStateOfferReceived
	session.LastActivity = time.Now()

	// Если у нас есть answer, отправляем его
	var responseAnswer string
	if session.Answer != "" {
		responseAnswer = session.Answer
	}

	response := ICESignalingResponse{
		Success:   true,
		SessionID: req.SessionID,
		Answer:    responseAnswer,
	}

	im.logger.Info("ICE offer handled",
		zap.String("session_id", req.SessionID),
		zap.String("peer_id", claims.Subject))

	c.JSON(http.StatusOK, response)
}

// HandleAnswer обрабатывает WebRTC answer
func (im *ICEManager) HandleAnswer(c *gin.Context) {
	var req ICESignalingRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		im.logger.Error("Failed to bind ICE answer request", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request format",
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	// Валидация JWT токена
	claims, err := im.validateToken(c)
	if err != nil {
		im.logger.Error("Failed to validate token", zap.Error(err))
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Invalid token",
			Code:    http.StatusUnauthorized,
			Message: err.Error(),
		})
		return
	}

	im.mu.Lock()
	defer im.mu.Unlock()

	// Получаем сессию
	session, exists := im.sessions[req.SessionID]
	if !exists {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "Session not found",
			Code:    http.StatusNotFound,
			Message: "Session with specified ID not found",
		})
		return
	}

	// Проверяем права доступа
	if session.Peer1ID != claims.Subject && session.Peer2ID != claims.Subject {
		c.JSON(http.StatusForbidden, ErrorResponse{
			Error:   "Access denied",
			Code:    http.StatusForbidden,
			Message: "You don't have access to this session",
		})
		return
	}

	// Сохраняем answer
	session.Answer = req.Answer
	session.State = ICESessionStateAnswerReceived
	session.LastActivity = time.Now()

	// Отправляем offer в ответе
	response := ICESignalingResponse{
		Success:   true,
		SessionID: req.SessionID,
		Offer:     session.Offer,
	}

	im.logger.Info("ICE answer handled",
		zap.String("session_id", req.SessionID),
		zap.String("peer_id", claims.Subject))

	c.JSON(http.StatusOK, response)
}

// GetSessionStatus возвращает статус ICE сессии
func (im *ICEManager) GetSessionStatus(c *gin.Context) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Missing session_id",
			Code:    http.StatusBadRequest,
			Message: "session_id parameter is required",
		})
		return
	}

	// Валидация JWT токена
	claims, err := im.validateToken(c)
	if err != nil {
		im.logger.Error("Failed to validate token", zap.Error(err))
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Invalid token",
			Code:    http.StatusUnauthorized,
			Message: err.Error(),
		})
		return
	}

	im.mu.RLock()
	defer im.mu.RUnlock()

	// Получаем сессию
	session, exists := im.sessions[sessionID]
	if !exists {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "Session not found",
			Code:    http.StatusNotFound,
			Message: "Session with specified ID not found",
		})
		return
	}

	// Проверяем права доступа
	if session.Peer1ID != claims.Subject && session.Peer2ID != claims.Subject {
		c.JSON(http.StatusForbidden, ErrorResponse{
			Error:   "Access denied",
			Code:    http.StatusForbidden,
			Message: "You don't have access to this session",
		})
		return
	}

	status := map[string]interface{}{
		"session_id":       session.SessionID,
		"peer1_id":         session.Peer1ID,
		"peer2_id":         session.Peer2ID,
		"state":            session.State.String(),
		"created_at":       session.CreatedAt,
		"last_activity":    session.LastActivity,
		"peer1_candidates": len(session.Peer1Candidates),
		"peer2_candidates": len(session.Peer2Candidates),
		"has_offer":        session.Offer != "",
		"has_answer":       session.Answer != "",
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"status":  status,
	})
}

// CleanupInactiveSessions очищает неактивные сессии
func (im *ICEManager) CleanupInactiveSessions(timeout time.Duration) {
	im.mu.Lock()
	defer im.mu.Unlock()

	now := time.Now()
	for sessionID, session := range im.sessions {
		if now.Sub(session.LastActivity) > timeout {
			delete(im.sessions, sessionID)
			im.logger.Debug("Cleaned up inactive ICE session",
				zap.String("session_id", sessionID))
		}
	}
}

// validateToken валидирует JWT токен
func (im *ICEManager) validateToken(c *gin.Context) (*P2PClaims, error) {
	// Получаем токен из заголовка Authorization
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("missing authorization header")
	}

	// Извлекаем токен
	_ = authHeader
	// if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
	//	token = authHeader[7:]
	// }

	// Здесь должна быть валидация токена
	// Для демонстрации возвращаем mock claims
	return &P2PClaims{
		Subject:  "peer-123",
		TenantID: "tenant-456",
	}, nil
}

// String возвращает строковое представление ICESessionState
func (s ICESessionState) String() string {
	switch s {
	case ICESessionStateNew:
		return "new"
	case ICESessionStateOfferReceived:
		return "offer_received"
	case ICESessionStateAnswerReceived:
		return "answer_received"
	case ICESessionStateCompleted:
		return "completed"
	case ICESessionStateFailed:
		return "failed"
	default:
		return "unknown"
	}
}
