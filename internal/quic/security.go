package quic

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
)

// QUICSecurity реализует QUIC Security согласно RFC 9000
// Включает address validation, Retry, anti-amplification protection
type QUICSecurity struct {
	logger *logger.Logger
	config *SecurityConfig

	// Address validation state
	validationTokens map[string]*ValidationToken
	validationMutex  sync.RWMutex

	// Anti-amplification protection
	amplificationLimits map[string]*AmplificationLimit
	amplificationMutex  sync.RWMutex

	// Stateless reset tokens
	resetTokens map[string]*ResetToken
	resetMutex  sync.RWMutex
}

// SecurityConfig конфигурация для QUIC Security
type SecurityConfig struct {
	// Address validation
	ValidationTokenLifetime time.Duration
	MaxValidationAttempts   int
	ValidationRetryDelay    time.Duration

	// Anti-amplification protection
	MaxAmplificationFactor float64
	AmplificationWindow    time.Duration
	MaxUnvalidatedPackets  int

	// Stateless reset
	ResetTokenLifetime time.Duration
	ResetTokenSecret   []byte
}

// ValidationToken токен для валидации адреса
type ValidationToken struct {
	Token     []byte
	ExpiresAt time.Time
	ClientIP  string
}

// AmplificationLimit лимит для anti-amplification защиты
type AmplificationLimit struct {
	BytesReceived    int64
	BytesSent        int64
	LastActivity     time.Time
	UnvalidatedCount int
}

// ResetToken токен для stateless reset
type ResetToken struct {
	Token        []byte
	ExpiresAt    time.Time
	ConnectionID []byte
}

// DefaultSecurityConfig возвращает конфигурацию по умолчанию
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		ValidationTokenLifetime: 10 * time.Second,
		MaxValidationAttempts:   3,
		ValidationRetryDelay:    100 * time.Millisecond,
		MaxAmplificationFactor:  3.0, // RFC 9000 Section 8.1
		AmplificationWindow:     1 * time.Second,
		MaxUnvalidatedPackets:   10,
		ResetTokenLifetime:      60 * time.Second,
		ResetTokenSecret:        make([]byte, 32), // 256-bit secret
	}
}

// NewQUICSecurity создает новый QUIC Security компонент
func NewQUICSecurity(logger *logger.Logger, config *SecurityConfig) *QUICSecurity {
	if config == nil {
		config = DefaultSecurityConfig()
	}

	// Генерируем случайный секрет для reset tokens если не предоставлен
	if len(config.ResetTokenSecret) == 0 {
		config.ResetTokenSecret = make([]byte, 32)
		if _, err := rand.Read(config.ResetTokenSecret); err != nil {
			logger.Error("Failed to generate reset token secret", zap.Error(err))
		}
	}

	return &QUICSecurity{
		logger:              logger,
		config:              config,
		validationTokens:    make(map[string]*ValidationToken),
		amplificationLimits: make(map[string]*AmplificationLimit),
		resetTokens:         make(map[string]*ResetToken),
	}
}

// ValidateAddress выполняет address validation согласно RFC 9000 Section 8
func (qs *QUICSecurity) ValidateAddress(ctx context.Context, clientAddr net.Addr) (*ValidationToken, error) {
	clientIP := clientAddr.String()

	qs.validationMutex.Lock()
	defer qs.validationMutex.Unlock()

	// Проверяем существующий токен
	if token, exists := qs.validationTokens[clientIP]; exists {
		if time.Now().Before(token.ExpiresAt) {
			qs.logger.Debug("Using existing validation token",
				zap.String("client_ip", clientIP),
				zap.Time("expires_at", token.ExpiresAt),
			)
			return token, nil
		}
		// Удаляем истекший токен
		delete(qs.validationTokens, clientIP)
	}

	// Создаем новый токен валидации
	token, err := qs.generateValidationToken(clientIP)
	if err != nil {
		return nil, fmt.Errorf("failed to generate validation token: %w", err)
	}

	// Сохраняем токен
	qs.validationTokens[clientIP] = token

	qs.logger.Info("Generated new validation token",
		zap.String("client_ip", clientIP),
		zap.Time("expires_at", token.ExpiresAt),
	)

	return token, nil
}

// generateValidationToken генерирует токен валидации адреса
func (qs *QUICSecurity) generateValidationToken(clientIP string) (*ValidationToken, error) {
	// Создаем токен на основе IP адреса, времени и секрета
	tokenData := make([]byte, 32)

	// Добавляем IP адрес
	copy(tokenData[:16], []byte(clientIP))

	// Добавляем timestamp
	timestamp := time.Now().Unix()
	binary.BigEndian.PutUint64(tokenData[16:24], uint64(timestamp))

	// Добавляем случайные данные
	if _, err := rand.Read(tokenData[24:]); err != nil {
		return nil, fmt.Errorf("failed to generate random token data: %w", err)
	}

	// Хешируем токен
	hash := sha256.Sum256(tokenData)

	return &ValidationToken{
		Token:     hash[:],
		ExpiresAt: time.Now().Add(qs.config.ValidationTokenLifetime),
		ClientIP:  clientIP,
	}, nil
}

// VerifyValidationToken проверяет токен валидации
func (qs *QUICSecurity) VerifyValidationToken(token []byte, clientIP string) bool {
	qs.validationMutex.RLock()
	defer qs.validationMutex.RUnlock()

	storedToken, exists := qs.validationTokens[clientIP]
	if !exists {
		qs.logger.Debug("No validation token found for client",
			zap.String("client_ip", clientIP),
		)
		return false
	}

	if time.Now().After(storedToken.ExpiresAt) {
		qs.logger.Debug("Validation token expired",
			zap.String("client_ip", clientIP),
			zap.Time("expires_at", storedToken.ExpiresAt),
		)
		return false
	}

	// Сравниваем токены
	if len(token) != len(storedToken.Token) {
		return false
	}

	for i := range token {
		if token[i] != storedToken.Token[i] {
			return false
		}
	}

	qs.logger.Debug("Validation token verified successfully",
		zap.String("client_ip", clientIP),
	)

	return true
}

// CheckAmplificationLimit проверяет anti-amplification лимиты согласно RFC 9000 Section 8.1
func (qs *QUICSecurity) CheckAmplificationLimit(clientIP string, bytesReceived, bytesSent int64) error {
	qs.amplificationMutex.Lock()
	defer qs.amplificationMutex.Unlock()

	limit, exists := qs.amplificationLimits[clientIP]
	if !exists {
		limit = &AmplificationLimit{
			LastActivity: time.Now(),
		}
		qs.amplificationLimits[clientIP] = limit
	}

	// Обновляем счетчики
	limit.BytesReceived += bytesReceived
	limit.BytesSent += bytesSent
	limit.LastActivity = time.Now()

	// Проверяем amplification factor
	if limit.BytesReceived > 0 {
		amplificationFactor := float64(limit.BytesSent) / float64(limit.BytesReceived)
		if amplificationFactor > qs.config.MaxAmplificationFactor {
			qs.logger.Warn("Amplification limit exceeded",
				zap.String("client_ip", clientIP),
				zap.Float64("amplification_factor", amplificationFactor),
				zap.Float64("max_factor", qs.config.MaxAmplificationFactor),
				zap.Int64("bytes_sent", limit.BytesSent),
				zap.Int64("bytes_received", limit.BytesReceived),
			)
			return fmt.Errorf("amplification limit exceeded: %.2f > %.2f",
				amplificationFactor, qs.config.MaxAmplificationFactor)
		}
	}

	// Проверяем количество невалидированных пакетов
	if limit.UnvalidatedCount > qs.config.MaxUnvalidatedPackets {
		qs.logger.Warn("Too many unvalidated packets",
			zap.String("client_ip", clientIP),
			zap.Int("unvalidated_count", limit.UnvalidatedCount),
			zap.Int("max_count", qs.config.MaxUnvalidatedPackets),
		)
		return fmt.Errorf("too many unvalidated packets: %d > %d",
			limit.UnvalidatedCount, qs.config.MaxUnvalidatedPackets)
	}

	return nil
}

// IncrementUnvalidatedPackets увеличивает счетчик невалидированных пакетов
func (qs *QUICSecurity) IncrementUnvalidatedPackets(clientIP string) {
	qs.amplificationMutex.Lock()
	defer qs.amplificationMutex.Unlock()

	limit, exists := qs.amplificationLimits[clientIP]
	if !exists {
		limit = &AmplificationLimit{
			LastActivity: time.Now(),
		}
		qs.amplificationLimits[clientIP] = limit
	}

	limit.UnvalidatedCount++
	limit.LastActivity = time.Now()
}

// ResetUnvalidatedPackets сбрасывает счетчик невалидированных пакетов после валидации
func (qs *QUICSecurity) ResetUnvalidatedPackets(clientIP string) {
	qs.amplificationMutex.Lock()
	defer qs.amplificationMutex.Unlock()

	if limit, exists := qs.amplificationLimits[clientIP]; exists {
		limit.UnvalidatedCount = 0
		limit.LastActivity = time.Now()
	}
}

// GenerateResetToken генерирует stateless reset token
func (qs *QUICSecurity) GenerateResetToken(connectionID []byte) (*ResetToken, error) {
	// Создаем токен на основе connection ID и секрета
	tokenData := make([]byte, len(connectionID)+len(qs.config.ResetTokenSecret))
	copy(tokenData, connectionID)
	copy(tokenData[len(connectionID):], qs.config.ResetTokenSecret)

	// Хешируем токен
	hash := sha256.Sum256(tokenData)

	token := &ResetToken{
		Token:        hash[:],
		ExpiresAt:    time.Now().Add(qs.config.ResetTokenLifetime),
		ConnectionID: connectionID,
	}

	qs.resetMutex.Lock()
	defer qs.resetMutex.Unlock()

	// Сохраняем токен
	qs.resetTokens[string(connectionID)] = token

	qs.logger.Debug("Generated reset token",
		zap.String("connection_id", string(connectionID)),
		zap.Time("expires_at", token.ExpiresAt),
	)

	return token, nil
}

// VerifyResetToken проверяет stateless reset token
func (qs *QUICSecurity) VerifyResetToken(token []byte, connectionID []byte) bool {
	qs.resetMutex.RLock()
	defer qs.resetMutex.RUnlock()

	storedToken, exists := qs.resetTokens[string(connectionID)]
	if !exists {
		return false
	}

	if time.Now().After(storedToken.ExpiresAt) {
		return false
	}

	// Сравниваем токены
	if len(token) != len(storedToken.Token) {
		return false
	}

	for i := range token {
		if token[i] != storedToken.Token[i] {
			return false
		}
	}

	return true
}

// CleanupExpiredTokens очищает истекшие токены
func (qs *QUICSecurity) CleanupExpiredTokens() {
	now := time.Now()

	// Очищаем validation tokens
	qs.validationMutex.Lock()
	for clientIP, token := range qs.validationTokens {
		if now.After(token.ExpiresAt) {
			delete(qs.validationTokens, clientIP)
		}
	}
	qs.validationMutex.Unlock()

	// Очищаем amplification limits
	qs.amplificationMutex.Lock()
	for clientIP, limit := range qs.amplificationLimits {
		if now.After(limit.LastActivity.Add(qs.config.AmplificationWindow)) {
			delete(qs.amplificationLimits, clientIP)
		}
	}
	qs.amplificationMutex.Unlock()

	// Очищаем reset tokens
	qs.resetMutex.Lock()
	for connectionID, token := range qs.resetTokens {
		if now.After(token.ExpiresAt) {
			delete(qs.resetTokens, connectionID)
		}
	}
	qs.resetMutex.Unlock()
}

// StartCleanupRoutine запускает горутину для очистки истекших токенов
func (qs *QUICSecurity) StartCleanupRoutine(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			qs.CleanupExpiredTokens()
		}
	}
}

// GetSecurityMetrics возвращает метрики безопасности
func (qs *QUICSecurity) GetSecurityMetrics() map[string]interface{} {
	qs.validationMutex.RLock()
	validationCount := len(qs.validationTokens)
	qs.validationMutex.RUnlock()

	qs.amplificationMutex.RLock()
	amplificationCount := len(qs.amplificationLimits)
	qs.amplificationMutex.RUnlock()

	qs.resetMutex.RLock()
	resetCount := len(qs.resetTokens)
	qs.resetMutex.RUnlock()

	return map[string]interface{}{
		"validation_tokens_count":    validationCount,
		"amplification_limits_count": amplificationCount,
		"reset_tokens_count":         resetCount,
		"max_amplification_factor":   qs.config.MaxAmplificationFactor,
		"validation_token_lifetime":  qs.config.ValidationTokenLifetime,
		"reset_token_lifetime":       qs.config.ResetTokenLifetime,
	}
}
