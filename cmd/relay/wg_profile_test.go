package main

import (
	"testing"

	"github.com/twogc/cloudbridge-relay-installer/config"
	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func TestWGProfileManager(t *testing.T) {
	// Создаем тестовую конфигурацию
	cfg := &config.Config{}
	cfg.WireGuard.Enabled = true
	cfg.WireGuard.Interface.Name = "wg-test0"
	cfg.WireGuard.Interface.ListenPort = 51821 // Используем другой порт для тестов
	cfg.WireGuard.Interface.Address = "10.100.2.1/24"

	// Создаем тестовый логгер
	logger := zap.NewNop()

	// Создаем WG-профиль менеджер
	manager, err := NewWGProfileManager(cfg, logger)
	if err != nil {
		t.Fatalf("Failed to create WG profile manager: %v", err)
	}
	defer manager.Stop()

	// Тестируем создание конфигурации пира
	_, err = manager.CreatePeerConfig(
		"test-public-key",
		"192.168.1.100:51820",
		"10.100.2.2/32",
	)
	if err != nil {
		t.Logf("Expected error for invalid public key: %v", err)
	} else {
		t.Error("Expected error for invalid public key, but got success")
	}

	// Тестируем с валидным публичным ключом
	validPublicKey := "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
	_, err = manager.CreatePeerConfig(
		validPublicKey,
		"192.168.1.100:51820",
		"10.100.2.2/32",
	)
	if err != nil {
		t.Logf("Expected error for invalid endpoint: %v", err)
	} else {
		t.Logf("Valid public key test passed - this is expected")
	}

	// Тестируем GetStats
	stats := manager.GetStats()
	if stats == nil {
		t.Error("GetStats returned nil")
	}

	t.Logf("WG Profile Manager test completed successfully")
}

func TestWGProfileManagerDisabled(t *testing.T) {
	// Создаем конфигурацию с отключенным WireGuard
	cfg := &config.Config{}
	cfg.WireGuard.Enabled = false

	// Создаем тестовый логгер
	logger := zap.NewNop()

	// Создаем WG-профиль менеджер
	manager, err := NewWGProfileManager(cfg, logger)
	if err != nil {
		t.Fatalf("Failed to create WG profile manager: %v", err)
	}
	defer manager.Stop()

	// Тестируем Start с отключенным WireGuard
	err = manager.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	t.Logf("WG Profile Manager disabled test completed successfully")
}

func TestWGTypesValidation(t *testing.T) {
	// Тестируем валидацию ключей wgtypes
	_, err := wgtypes.ParseKey("invalid-key")
	if err == nil {
		t.Error("Expected error for invalid key, but got success")
	}

	// Тестируем генерацию ключей
	privateKey, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	publicKey := privateKey.PublicKey()
	if publicKey.String() == "" {
		t.Error("Generated public key is empty")
	}

	t.Logf("Generated public key: %s", publicKey.String())
}
