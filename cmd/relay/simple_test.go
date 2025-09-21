package main

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/twogc/cloudbridge-relay-installer/config"
	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
)

func TestSimpleFallbackManager(t *testing.T) {
	testLogger, err := logger.New(logger.Config{
		Level:    logger.DEBUG,
		JSON:     true,
		Output:   "stdout",
		FilePath: "/tmp/test.log",
	})
	assert.NoError(t, err)

	cfg := &config.Config{}
	manager := NewFallbackManager(testLogger, cfg)

	assert.NotNil(t, manager)
	assert.NotNil(t, manager.logger)
	assert.NotNil(t, manager.config)
	assert.False(t, manager.isRunning)
}

func TestSimpleMASQUEServer(t *testing.T) {
	testLogger, err := logger.New(logger.Config{
		Level:    logger.DEBUG,
		JSON:     true,
		Output:   "stdout",
		FilePath: "/tmp/test.log",
	})
	assert.NoError(t, err)

	cfg := &config.Config{
		MASQUE: struct {
			Enabled         bool     `json:"enabled"`
			ProxyURL        string   `json:"proxy_url"`
			ALPNProtocols   []string `json:"alpn_protocols"`
			FallbackEnabled bool     `json:"fallback_enabled"`
		}{
			Enabled: true,
		},
	}

	server := NewMASQUEServer(testLogger, cfg)
	assert.NotNil(t, server)
	assert.NotNil(t, server.logger)
	assert.NotNil(t, server.config)
}

func TestSimpleWebSocketFallback(t *testing.T) {
	testLogger, err := logger.New(logger.Config{
		Level:    logger.DEBUG,
		JSON:     true,
		Output:   "stdout",
		FilePath: "/tmp/test.log",
	})
	assert.NoError(t, err)

	cfg := &config.Config{
		WebSocketFallback: struct {
			Enabled    bool   `json:"enabled"`
			ListenAddr string `json:"listen_addr"`
		}{
			Enabled:    true,
			ListenAddr: "127.0.0.1:0",
		},
	}

	fallback := NewWebSocketFallback(testLogger, cfg)
	assert.NotNil(t, fallback)
	assert.NotNil(t, fallback.logger)
	assert.NotNil(t, fallback.config)
}
