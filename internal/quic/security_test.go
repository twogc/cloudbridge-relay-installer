package quic

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
)

func TestQUICSecurity_ValidateAddress(t *testing.T) {
	logger, err := logger.New(logger.Config{
		Level:  logger.DEBUG,
		JSON:   true,
		Output: "stdout",
	})
	require.NoError(t, err)

	config := DefaultSecurityConfig()
	security := NewQUICSecurity(logger, config)

	// Тестируем валидацию адреса
	clientAddr := &net.UDPAddr{IP: net.ParseIP("192.168.1.100"), Port: 12345}

	token, err := security.ValidateAddress(context.Background(), clientAddr)
	require.NoError(t, err)
	assert.NotNil(t, token)
	assert.Equal(t, clientAddr.String(), token.ClientIP)
	assert.True(t, time.Now().Before(token.ExpiresAt))

	// Проверяем, что токен верифицируется
	verified := security.VerifyValidationToken(token.Token, clientAddr.String())
	assert.True(t, verified)

	// Проверяем, что неверный токен не проходит
	verified = security.VerifyValidationToken([]byte("invalid"), clientAddr.String())
	assert.False(t, verified)
}

func TestQUICSecurity_CheckAmplificationLimit(t *testing.T) {
	logger, err := logger.New(logger.Config{
		Level:  logger.DEBUG,
		JSON:   true,
		Output: "stdout",
	})
	require.NoError(t, err)

	config := DefaultSecurityConfig()
	config.MaxAmplificationFactor = 2.0
	security := NewQUICSecurity(logger, config)

	clientIP := "192.168.1.100"

	// Тестируем нормальный amplification factor
	err = security.CheckAmplificationLimit(clientIP, 1000, 1500) // 1.5x
	assert.NoError(t, err)

	// Тестируем превышение amplification factor
	err = security.CheckAmplificationLimit(clientIP, 1000, 3000) // 3.0x > 2.0x
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "amplification limit exceeded")
}

func TestQUICSecurity_GenerateResetToken(t *testing.T) {
	logger, err := logger.New(logger.Config{
		Level:  logger.DEBUG,
		JSON:   true,
		Output: "stdout",
	})
	require.NoError(t, err)

	config := DefaultSecurityConfig()
	security := NewQUICSecurity(logger, config)

	connectionID := []byte("test-connection-id")

	token, err := security.GenerateResetToken(connectionID)
	require.NoError(t, err)
	assert.NotNil(t, token)
	assert.Equal(t, connectionID, token.ConnectionID)
	assert.True(t, time.Now().Before(token.ExpiresAt))

	// Проверяем, что токен верифицируется
	verified := security.VerifyResetToken(token.Token, connectionID)
	assert.True(t, verified)

	// Проверяем, что неверный токен не проходит
	verified = security.VerifyResetToken([]byte("invalid"), connectionID)
	assert.False(t, verified)
}

func TestQUICSecurity_CleanupExpiredTokens(t *testing.T) {
	logger, err := logger.New(logger.Config{
		Level:  logger.DEBUG,
		JSON:   true,
		Output: "stdout",
	})
	require.NoError(t, err)

	config := DefaultSecurityConfig()
	config.ValidationTokenLifetime = 100 * time.Millisecond
	config.ResetTokenLifetime = 100 * time.Millisecond
	security := NewQUICSecurity(logger, config)

	// Создаем токены
	clientAddr := &net.UDPAddr{IP: net.ParseIP("192.168.1.100"), Port: 12345}
	_, err = security.ValidateAddress(context.Background(), clientAddr)
	require.NoError(t, err)

	connectionID := []byte("test-connection-id")
	_, err = security.GenerateResetToken(connectionID)
	require.NoError(t, err)

	// Ждем истечения токенов
	time.Sleep(200 * time.Millisecond)

	// Очищаем истекшие токены
	security.CleanupExpiredTokens()

	// Проверяем, что токены удалены
	verified := security.VerifyValidationToken([]byte("any"), clientAddr.String())
	assert.False(t, verified)

	verified = security.VerifyResetToken([]byte("any"), connectionID)
	assert.False(t, verified)
}

func TestQUICSecurity_GetSecurityMetrics(t *testing.T) {
	logger, err := logger.New(logger.Config{
		Level:  logger.DEBUG,
		JSON:   true,
		Output: "stdout",
	})
	require.NoError(t, err)

	config := DefaultSecurityConfig()
	security := NewQUICSecurity(logger, config)

	// Создаем несколько токенов
	clientAddr1 := &net.UDPAddr{IP: net.ParseIP("192.168.1.100"), Port: 12345}
	clientAddr2 := &net.UDPAddr{IP: net.ParseIP("192.168.1.101"), Port: 12346}

	_, err = security.ValidateAddress(context.Background(), clientAddr1)
	require.NoError(t, err)
	_, err = security.ValidateAddress(context.Background(), clientAddr2)
	require.NoError(t, err)

	// Получаем метрики
	metrics := security.GetSecurityMetrics()

	assert.Equal(t, 2, metrics["validation_tokens_count"])
	assert.Equal(t, 0, metrics["amplification_limits_count"])
	assert.Equal(t, 0, metrics["reset_tokens_count"])
	assert.Equal(t, config.MaxAmplificationFactor, metrics["max_amplification_factor"])
	assert.Equal(t, config.ValidationTokenLifetime, metrics["validation_token_lifetime"])
	assert.Equal(t, config.ResetTokenLifetime, metrics["reset_token_lifetime"])
}
