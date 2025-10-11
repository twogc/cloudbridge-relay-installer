package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/twogc/cloudbridge-relay-installer/config"
	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
)

func TestNewAdapter(t *testing.T) {
	t.Run("no providers enabled", func(t *testing.T) {
		cfg := &config.Config{}
		logger, err := logger.New(logger.Config{
			Level:  logger.INFO,
			Output: "stdout",
			JSON:   false,
		})
		require.NoError(t, err)
		adapter, err := NewAdapter(cfg, logger)
		require.NoError(t, err)
		assert.NotNil(t, adapter)
		assert.Equal(t, "noop", adapter.GetProvider())
	})

	t.Run("nil logger", func(t *testing.T) {
		cfg := &config.Config{}
		adapter, err := NewAdapter(cfg, nil)
		assert.Error(t, err)
		assert.Nil(t, adapter)
		assert.Contains(t, err.Error(), "logger is required")
	})
}

func TestAdapter_ValidateToken(t *testing.T) {
	t.Run("noop provider", func(t *testing.T) {
		cfg := &config.Config{}
		logger, err := logger.New(logger.Config{
			Level:  logger.INFO,
			Output: "stdout",
			JSON:   false,
		})
		require.NoError(t, err)
		adapter, err := NewAdapter(cfg, logger)
		require.NoError(t, err)

		// Test with invalid token - should work with noop provider
		ctx := context.Background()
		claims, err := adapter.ValidateToken(ctx, "invalid-token")
		assert.NoError(t, err)
		assert.NotNil(t, claims)
		assert.Equal(t, "noop", claims.Provider)
	})
}

func TestAdapter_ContextCancellation(t *testing.T) {
	cfg := &config.Config{}
	logger, err := logger.New(logger.Config{
		Level:  logger.INFO,
		Output: "stdout",
		JSON:   false,
	})
	require.NoError(t, err)
	adapter, err := NewAdapter(cfg, logger)
	require.NoError(t, err)

	// Test context cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	claims, err := adapter.ValidateToken(ctx, "test-token")
	assert.NoError(t, err)
	assert.NotNil(t, claims)
}


