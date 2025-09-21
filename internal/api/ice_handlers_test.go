package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// ICESessionInterface интерфейс для ICESession
type ICESessionInterface interface {
	AddCandidate(candidate string) error
	SetOffer(offer string) error
	SetAnswer(answer string) error
	GetStatus() string
	GetCandidates() []string
	GetOffer() string
	GetAnswer() string
	IsExpired() bool
}

// TestICEManager тестовая версия ICEManager с интерфейсом
type TestICEManager struct {
	sessions map[string]ICESessionInterface
	logger   *zap.Logger
}

// NewTestICEManager создает новый тестовый ICE Manager
func NewTestICEManager(logger *zap.Logger) *TestICEManager {
	return &TestICEManager{
		sessions: make(map[string]ICESessionInterface),
		logger:   logger,
	}
}

// ExchangeCandidates обменивается ICE candidates
func (tim *TestICEManager) ExchangeCandidates(c *gin.Context) {
	var req struct {
		SessionID string `json:"session_id" binding:"required"`
		Candidate string `json:"candidate" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid input",
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	session, exists := tim.sessions[req.SessionID]
	if !exists {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "Session not found",
			Code:    http.StatusNotFound,
			Message: "Session not found",
		})
		return
	}

	if err := session.AddCandidate(req.Candidate); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal server error",
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, map[string]string{"message": "Candidate exchanged successfully"})
}

// HandleOffer обрабатывает WebRTC offer
func (tim *TestICEManager) HandleOffer(c *gin.Context) {
	var req struct {
		SessionID string `json:"session_id" binding:"required"`
		Offer     string `json:"offer" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid input",
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	session, exists := tim.sessions[req.SessionID]
	if !exists {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "Session not found",
			Code:    http.StatusNotFound,
			Message: "Session not found",
		})
		return
	}

	if err := session.SetOffer(req.Offer); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal server error",
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, map[string]string{"message": "Offer handled successfully"})
}

// HandleAnswer обрабатывает WebRTC answer
func (tim *TestICEManager) HandleAnswer(c *gin.Context) {
	var req struct {
		SessionID string `json:"session_id" binding:"required"`
		Answer    string `json:"answer" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid input",
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	session, exists := tim.sessions[req.SessionID]
	if !exists {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "Session not found",
			Code:    http.StatusNotFound,
			Message: "Session not found",
		})
		return
	}

	if err := session.SetAnswer(req.Answer); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal server error",
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, map[string]string{"message": "Answer handled successfully"})
}

// GetSessionStatus возвращает статус ICE сессии
func (tim *TestICEManager) GetSessionStatus(c *gin.Context) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Missing session_id",
			Code:    http.StatusBadRequest,
			Message: "session_id parameter is required",
		})
		return
	}

	session, exists := tim.sessions[sessionID]
	if !exists {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "Session not found",
			Code:    http.StatusNotFound,
			Message: "Session not found",
		})
		return
	}

	status := map[string]interface{}{
		"session_id": sessionID,
		"status":     session.GetStatus(),
		"candidates": session.GetCandidates(),
		"offer":      session.GetOffer(),
		"answer":     session.GetAnswer(),
	}

	c.JSON(http.StatusOK, status)
}

// CreateSession создает новую ICE сессию
func (tim *TestICEManager) CreateSession(c *gin.Context) {
	sessionID := "session_" + time.Now().Format("20060102150405")

	// Создаем новую сессию
	session := &MockICESession{}
	tim.sessions[sessionID] = session

	c.JSON(http.StatusOK, map[string]string{"session_id": sessionID})
}

// CleanupInactiveSessions очищает неактивные сессии
func (tim *TestICEManager) CleanupInactiveSessions() {
	for sessionID, session := range tim.sessions {
		if session.IsExpired() {
			delete(tim.sessions, sessionID)
		}
	}
}

// MockICESession мок для ICESession
type MockICESession struct {
	mock.Mock
}

func (m *MockICESession) AddCandidate(candidate string) error {
	args := m.Called(candidate)
	return args.Error(0)
}

func (m *MockICESession) SetOffer(offer string) error {
	args := m.Called(offer)
	return args.Error(0)
}

func (m *MockICESession) SetAnswer(answer string) error {
	args := m.Called(answer)
	return args.Error(0)
}

func (m *MockICESession) GetStatus() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockICESession) GetCandidates() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *MockICESession) GetOffer() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockICESession) GetAnswer() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockICESession) IsExpired() bool {
	args := m.Called()
	return args.Bool(0)
}

func TestNewTestICEManager(t *testing.T) {
	logger := zap.NewNop()
	manager := NewTestICEManager(logger)

	assert.NotNil(t, manager)
	assert.Equal(t, logger, manager.logger)
	assert.NotNil(t, manager.sessions)
}

func TestICEManager_ExchangeCandidates(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*MockICESession)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful candidate exchange",
			requestBody: map[string]interface{}{
				"session_id": "session1",
				"candidate":  "candidate:1 1 UDP 2113667326 192.168.1.1 54400 typ host",
			},
			mockSetup: func(s *MockICESession) {
				s.On("AddCandidate", "candidate:1 1 UDP 2113667326 192.168.1.1 54400 typ host").Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"Candidate exchanged successfully"}`,
		},
		{
			name: "session not found",
			requestBody: map[string]interface{}{
				"session_id": "nonexistent",
				"candidate":  "candidate:1 1 UDP 2113667326 192.168.1.1 54400 typ host",
			},
			mockSetup: func(s *MockICESession) {
				// No mock setup needed for nonexistent session
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"code":404,"error":"Session not found","message":"Session not found"}`,
		},
		{
			name: "candidate exchange error",
			requestBody: map[string]interface{}{
				"session_id": "session1",
				"candidate":  "invalid candidate",
			},
			mockSetup: func(s *MockICESession) {
				s.On("AddCandidate", "invalid candidate").Return(assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"code":500,"error":"Internal server error","message":"assert.AnError general error for testing"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			requestBody, _ := json.Marshal(tt.requestBody)
			c.Request = httptest.NewRequest("POST", "/ice/candidate", bytes.NewBuffer(requestBody))

			manager := NewTestICEManager(zap.NewNop())

			// Create session if needed
			if tt.name == "successful candidate exchange" || tt.name == "candidate exchange error" {
				session := &MockICESession{}
				tt.mockSetup(session)
				manager.sessions["session1"] = session
			}

			manager.ExchangeCandidates(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
		})
	}
}

func TestICEManager_HandleOffer(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*MockICESession)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful offer handling",
			requestBody: map[string]interface{}{
				"session_id": "session1",
				"offer":      "v=0\r\no=- 123456789 123456789 IN IP4 192.168.1.1\r\ns=-\r\nt=0 0\r\n",
			},
			mockSetup: func(s *MockICESession) {
				s.On("SetOffer", "v=0\r\no=- 123456789 123456789 IN IP4 192.168.1.1\r\ns=-\r\nt=0 0\r\n").Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"Offer handled successfully"}`,
		},
		{
			name: "session not found",
			requestBody: map[string]interface{}{
				"session_id": "nonexistent",
				"offer":      "v=0\r\no=- 123456789 123456789 IN IP4 192.168.1.1\r\ns=-\r\nt=0 0\r\n",
			},
			mockSetup: func(s *MockICESession) {
				// No mock setup needed for nonexistent session
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"code":404,"error":"Session not found","message":"Session not found"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			requestBody, _ := json.Marshal(tt.requestBody)
			c.Request = httptest.NewRequest("POST", "/ice/offer", bytes.NewBuffer(requestBody))

			manager := NewTestICEManager(zap.NewNop())

			// Create session if needed
			if tt.name == "successful offer handling" {
				session := &MockICESession{}
				tt.mockSetup(session)
				manager.sessions["session1"] = session
			}

			manager.HandleOffer(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
		})
	}
}

func TestICEManager_HandleAnswer(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*MockICESession)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful answer handling",
			requestBody: map[string]interface{}{
				"session_id": "session1",
				"answer":     "v=0\r\no=- 123456789 123456789 IN IP4 192.168.1.2\r\ns=-\r\nt=0 0\r\n",
			},
			mockSetup: func(s *MockICESession) {
				s.On("SetAnswer", "v=0\r\no=- 123456789 123456789 IN IP4 192.168.1.2\r\ns=-\r\nt=0 0\r\n").Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"Answer handled successfully"}`,
		},
		{
			name: "session not found",
			requestBody: map[string]interface{}{
				"session_id": "nonexistent",
				"answer":     "v=0\r\no=- 123456789 123456789 IN IP4 192.168.1.2\r\ns=-\r\nt=0 0\r\n",
			},
			mockSetup: func(s *MockICESession) {
				// No mock setup needed for nonexistent session
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"code":404,"error":"Session not found","message":"Session not found"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			requestBody, _ := json.Marshal(tt.requestBody)
			c.Request = httptest.NewRequest("POST", "/ice/answer", bytes.NewBuffer(requestBody))

			manager := NewTestICEManager(zap.NewNop())

			// Create session if needed
			if tt.name == "successful answer handling" {
				session := &MockICESession{}
				tt.mockSetup(session)
				manager.sessions["session1"] = session
			}

			manager.HandleAnswer(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
		})
	}
}

func TestICEManager_GetSessionStatus(t *testing.T) {
	tests := []struct {
		name           string
		sessionID      string
		mockSetup      func(*MockICESession)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:      "successful status retrieval",
			sessionID: "session1",
			mockSetup: func(s *MockICESession) {
				s.On("GetStatus").Return("connected")
				s.On("GetCandidates").Return([]string{"candidate:1 1 UDP 2113667326 192.168.1.1 54400 typ host"})
				s.On("GetOffer").Return("v=0\r\no=- 123456789 123456789 IN IP4 192.168.1.1\r\ns=-\r\nt=0 0\r\n")
				s.On("GetAnswer").Return("v=0\r\no=- 123456789 123456789 IN IP4 192.168.1.2\r\ns=-\r\nt=0 0\r\n")
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"session_id":"session1","status":"connected","candidates":["candidate:1 1 UDP 2113667326 192.168.1.1 54400 typ host"],"offer":"v=0\r\no=- 123456789 123456789 IN IP4 192.168.1.1\r\ns=-\r\nt=0 0\r\n","answer":"v=0\r\no=- 123456789 123456789 IN IP4 192.168.1.2\r\ns=-\r\nt=0 0\r\n"}`,
		},
		{
			name:      "session not found",
			sessionID: "nonexistent",
			mockSetup: func(s *MockICESession) {
				// No mock setup needed for nonexistent session
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"code":404,"error":"Session not found","message":"Session not found"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = gin.Params{{Key: "session_id", Value: tt.sessionID}}
			c.Request = httptest.NewRequest("GET", "/ice/status/"+tt.sessionID, nil)

			manager := NewTestICEManager(zap.NewNop())

			// Create session if needed
			if tt.name == "successful status retrieval" {
				session := &MockICESession{}
				tt.mockSetup(session)
				manager.sessions["session1"] = session
			}

			manager.GetSessionStatus(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
		})
	}
}

func TestICEManager_CreateSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	manager := NewTestICEManager(zap.NewNop())

	manager.CreateSession(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "session_id")
	assert.NotEmpty(t, response["session_id"])
}

func TestICEManager_CleanupInactiveSessions(t *testing.T) {
	manager := NewTestICEManager(zap.NewNop())

	// Create a mock session that is expired
	session := &MockICESession{}
	session.On("IsExpired").Return(true)
	manager.sessions["expired_session"] = session

	// Create a mock session that is not expired
	activeSession := &MockICESession{}
	activeSession.On("IsExpired").Return(false)
	manager.sessions["active_session"] = activeSession

	// Run cleanup
	manager.CleanupInactiveSessions()

	// Check that expired session was removed
	_, exists := manager.sessions["expired_session"]
	assert.False(t, exists)

	// Check that active session still exists
	_, exists = manager.sessions["active_session"]
	assert.True(t, exists)

	session.AssertExpectations(t)
	activeSession.AssertExpectations(t)
}

func TestICEManager_InvalidInput(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	requestBody := `{"invalid": "json"`
	c.Request = httptest.NewRequest("POST", "/ice/candidate", bytes.NewBufferString(requestBody))

	manager := NewTestICEManager(zap.NewNop())

	manager.ExchangeCandidates(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"code":400,"error":"Invalid input","message":"unexpected EOF"}`, w.Body.String())
}

func TestICEManager_MissingFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	requestBody := `{"session_id": "session1"}`
	c.Request = httptest.NewRequest("POST", "/ice/candidate", bytes.NewBufferString(requestBody))

	manager := NewTestICEManager(zap.NewNop())

	manager.ExchangeCandidates(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"code":400,"error":"Invalid input","message":"Key: 'Candidate' Error:Field validation for 'Candidate' failed on the 'required' tag"}`, w.Body.String())
}

func TestICEManager_SessionExpiration(t *testing.T) {
	manager := NewTestICEManager(zap.NewNop())

	// Create a session
	session := &MockICESession{}
	// Устанавливаем ожидание, что IsExpired будет вызван и вернет true
	session.On("IsExpired").Return(true).Once()
	manager.sessions["test_session"] = session

	// Verify session exists
	_, exists := manager.sessions["test_session"]
	assert.True(t, exists)

	// Run cleanup
	manager.CleanupInactiveSessions()

	// Verify session was removed
	_, exists = manager.sessions["test_session"]
	assert.False(t, exists, "Session should be removed after cleanup")

	// Verify mock expectations were met
	session.AssertExpectations(t)
}
