package p2p

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// HandoverManager управляет seamless handover между relay серверами
type HandoverManager struct {
	logger *zap.Logger

	// Состояния соединений
	connections    map[string]*HandoverConnectionState
	connectionsMux sync.RWMutex

	// Continuation tokens
	tokens    map[string]*ContinuationToken
	tokensMux sync.RWMutex

	// Handover configuration
	config *HandoverConfig
}

// HandoverConnectionState представляет состояние соединения для handover
type HandoverConnectionState struct {
	ID                string    `json:"id"`
	TenantID          string    `json:"tenant_id"`
	UserID            string    `json:"user_id"`
	CurrentRelay      string    `json:"current_relay"`
	PreviousRelay     string    `json:"previous_relay,omitempty"`
	State             string    `json:"state"` // good, degraded, switching, stabilize
	CreatedAt         time.Time `json:"created_at"`
	LastActivity      time.Time `json:"last_activity"`
	ContinuationToken string    `json:"continuation_token,omitempty"`

	// Handover metrics
	HandoverCount int       `json:"handover_count"`
	LastHandover  time.Time `json:"last_handover,omitempty"`
	CooldownUntil time.Time `json:"cooldown_until,omitempty"`
}

// ContinuationToken представляет токен для продолжения сессии
type ContinuationToken struct {
	Token        string                 `json:"token"`
	ConnectionID string                 `json:"connection_id"`
	TenantID     string                 `json:"tenant_id"`
	UserID       string                 `json:"user_id"`
	RelayID      string                 `json:"relay_id"`
	ExpiresAt    time.Time              `json:"expires_at"`
	CreatedAt    time.Time              `json:"created_at"`
	SessionData  map[string]interface{} `json:"session_data"`
}

// HandoverConfig конфигурация для handover
type HandoverConfig struct {
	// Лимиты на переключения
	MaxHandoversPerMinute int           `json:"max_handovers_per_minute"`
	CooldownPeriod        time.Duration `json:"cooldown_period"`

	// Timeouts
	TokenExpiry     time.Duration `json:"token_expiry"`
	HandoverTimeout time.Duration `json:"handover_timeout"`

	// Hysteresis settings
	HysteresisThreshold float64       `json:"hysteresis_threshold"`
	StabilizationTime   time.Duration `json:"stabilization_time"`
}

// DefaultHandoverConfig возвращает конфигурацию по умолчанию
func DefaultHandoverConfig() *HandoverConfig {
	return &HandoverConfig{
		MaxHandoversPerMinute: 3,                // Максимум 3 переключения в минуту
		CooldownPeriod:        30 * time.Second, // 30 секунд cooldown
		TokenExpiry:           5 * time.Minute,  // Токен действует 5 минут
		HandoverTimeout:       10 * time.Second, // 10 секунд на handover
		HysteresisThreshold:   0.1,              // 10% гистерезис
		StabilizationTime:     2 * time.Minute,  // 2 минуты стабилизации
	}
}

// NewHandoverManager создает новый менеджер handover
func NewHandoverManager(logger *zap.Logger, config *HandoverConfig) *HandoverManager {
	if config == nil {
		config = DefaultHandoverConfig()
	}

	return &HandoverManager{
		logger:      logger,
		connections: make(map[string]*HandoverConnectionState),
		tokens:      make(map[string]*ContinuationToken),
		config:      config,
	}
}

// CreateConnection создает новое соединение с continuation token
func (hm *HandoverManager) CreateConnection(tenantID, userID, relayID string) (*HandoverConnectionState, error) {
	connectionID := hm.generateConnectionID()

	// Создаем continuation token
	token, err := hm.createContinuationToken(connectionID, tenantID, userID, relayID)
	if err != nil {
		return nil, fmt.Errorf("failed to create continuation token: %w", err)
	}

	// Создаем состояние соединения
	state := &HandoverConnectionState{
		ID:                connectionID,
		TenantID:          tenantID,
		UserID:            userID,
		CurrentRelay:      relayID,
		State:             "good",
		CreatedAt:         time.Now(),
		LastActivity:      time.Now(),
		ContinuationToken: token.Token,
		HandoverCount:     0,
	}

	hm.connectionsMux.Lock()
	hm.connections[connectionID] = state
	hm.connectionsMux.Unlock()

	hm.logger.Info("Connection created with continuation token",
		zap.String("connection_id", connectionID),
		zap.String("tenant_id", tenantID),
		zap.String("user_id", userID),
		zap.String("relay_id", relayID),
		zap.String("token", token.Token))

	return state, nil
}

// ResumeConnection возобновляет соединение по continuation token
func (hm *HandoverManager) ResumeConnection(token string, newRelayID string) (*HandoverConnectionState, error) {
	// Находим токен
	hm.tokensMux.RLock()
	continuationToken, exists := hm.tokens[token]
	hm.tokensMux.RUnlock()

	if !exists {
		return nil, fmt.Errorf("continuation token not found")
	}

	// Проверяем срок действия
	if time.Now().After(continuationToken.ExpiresAt) {
		hm.cleanupToken(token)
		return nil, fmt.Errorf("continuation token expired")
	}

	// Находим соединение
	hm.connectionsMux.RLock()
	state, exists := hm.connections[continuationToken.ConnectionID]
	hm.connectionsMux.RUnlock()

	if !exists {
		return nil, fmt.Errorf("connection not found")
	}

	// Проверяем cooldown
	if time.Now().Before(state.CooldownUntil) {
		return nil, fmt.Errorf("connection is in cooldown period")
	}

	// Обновляем состояние
	state.PreviousRelay = state.CurrentRelay
	state.CurrentRelay = newRelayID
	state.LastActivity = time.Now()
	state.HandoverCount++
	state.LastHandover = time.Now()
	state.CooldownUntil = time.Now().Add(hm.config.CooldownPeriod)

	// Создаем новый continuation token
	newToken, err := hm.createContinuationToken(state.ID, state.TenantID, state.UserID, newRelayID)
	if err != nil {
		return nil, fmt.Errorf("failed to create new continuation token: %w", err)
	}

	state.ContinuationToken = newToken.Token

	hm.logger.Info("Connection resumed with new relay",
		zap.String("connection_id", state.ID),
		zap.String("previous_relay", state.PreviousRelay),
		zap.String("new_relay", newRelayID),
		zap.Int("handover_count", state.HandoverCount),
		zap.String("new_token", newToken.Token))

	return state, nil
}

// ShouldHandover определяет, нужно ли переключение на основе метрик
func (hm *HandoverManager) ShouldHandover(connectionID string, currentMetrics, newMetrics map[string]float64) (bool, string) {
	hm.connectionsMux.RLock()
	state, exists := hm.connections[connectionID]
	hm.connectionsMux.RUnlock()

	if !exists {
		return false, "connection not found"
	}

	// Проверяем cooldown
	if time.Now().Before(state.CooldownUntil) {
		return false, "connection in cooldown"
	}

	// Проверяем лимит переключений
	if state.HandoverCount >= hm.config.MaxHandoversPerMinute {
		return false, "handover limit exceeded"
	}

	// Проверяем гистерезис
	latencyImprovement := (currentMetrics["latency"] - newMetrics["latency"]) / currentMetrics["latency"]

	// Применяем гистерезис
	if state.State == "good" && latencyImprovement < hm.config.HysteresisThreshold {
		return false, "improvement below hysteresis threshold"
	}

	// Проверяем стабилизацию
	if state.State == "switching" && time.Since(state.LastHandover) < hm.config.StabilizationTime {
		return false, "stabilization period active"
	}

	return true, "handover recommended"
}

// UpdateConnectionState обновляет состояние соединения
func (hm *HandoverManager) UpdateConnectionState(connectionID, newState string) error {
	hm.connectionsMux.Lock()
	defer hm.connectionsMux.Unlock()

	state, exists := hm.connections[connectionID]
	if !exists {
		return fmt.Errorf("connection not found")
	}

	oldState := state.State
	state.State = newState
	state.LastActivity = time.Now()

	hm.logger.Info("Connection state updated",
		zap.String("connection_id", connectionID),
		zap.String("old_state", oldState),
		zap.String("new_state", newState))

	return nil
}

// GetConnectionState возвращает состояние соединения
func (hm *HandoverManager) GetConnectionState(connectionID string) (*HandoverConnectionState, error) {
	hm.connectionsMux.RLock()
	defer hm.connectionsMux.RUnlock()

	state, exists := hm.connections[connectionID]
	if !exists {
		return nil, fmt.Errorf("connection not found")
	}

	return state, nil
}

// createContinuationToken создает новый continuation token
func (hm *HandoverManager) createContinuationToken(connectionID, tenantID, userID, relayID string) (*ContinuationToken, error) {
	// Генерируем случайный токен
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	token := base64.URLEncoding.EncodeToString(tokenBytes)

	continuationToken := &ContinuationToken{
		Token:        token,
		ConnectionID: connectionID,
		TenantID:     tenantID,
		UserID:       userID,
		RelayID:      relayID,
		ExpiresAt:    time.Now().Add(hm.config.TokenExpiry),
		CreatedAt:    time.Now(),
		SessionData:  make(map[string]interface{}),
	}

	hm.tokensMux.Lock()
	hm.tokens[token] = continuationToken
	hm.tokensMux.Unlock()

	return continuationToken, nil
}

// cleanupToken удаляет токен
func (hm *HandoverManager) cleanupToken(token string) {
	hm.tokensMux.Lock()
	delete(hm.tokens, token)
	hm.tokensMux.Unlock()
}

// generateConnectionID генерирует уникальный ID соединения
func (hm *HandoverManager) generateConnectionID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return fmt.Sprintf("conn_%x", bytes)
}

// CleanupExpiredTokens удаляет истекшие токены
func (hm *HandoverManager) CleanupExpiredTokens() {
	now := time.Now()

	hm.tokensMux.Lock()
	for token, continuationToken := range hm.tokens {
		if now.After(continuationToken.ExpiresAt) {
			delete(hm.tokens, token)
		}
	}
	hm.tokensMux.Unlock()
}

// GetHandoverStats возвращает статистику handover
func (hm *HandoverManager) GetHandoverStats() map[string]interface{} {
	hm.connectionsMux.RLock()
	defer hm.connectionsMux.RUnlock()

	stats := map[string]interface{}{
		"total_connections": len(hm.connections),
		"active_tokens":     len(hm.tokens),
		"states":            make(map[string]int),
		"handover_counts":   make(map[int]int),
	}

	states := make(map[string]int)
	handoverCounts := make(map[int]int)

	for _, state := range hm.connections {
		states[state.State]++
		handoverCounts[state.HandoverCount]++
	}

	stats["states"] = states
	stats["handover_counts"] = handoverCounts

	return stats
}
