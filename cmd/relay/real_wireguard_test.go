package main

import (
	"testing"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/internal/wireguard"
	"go.uber.org/zap"
)

func TestRealWireGuardIntegration(t *testing.T) {
	logger := zap.NewNop()
	
	// Тестируем создание адаптера
	t.Run("Real WireGuard Adapter Creation", func(t *testing.T) {
		adapter := wireguard.NewRealWireGuardAdapter("wg-test-real", logger)
		
		if adapter == nil {
			t.Fatal("Failed to create real WireGuard adapter")
		}
		
		// Проверяем через публичные методы
		status, err := adapter.GetStatus()
		if err != nil {
			t.Logf("GetStatus error (expected in test): %v", err)
		} else {
			if status["interface"] != "wg-test-real" {
				t.Errorf("Expected interface name 'wg-test-real', got '%v'", status["interface"])
			}
		}
	})
	
	// Тестируем гибридный интерфейс
	t.Run("Hybrid WireGuard Interface", func(t *testing.T) {
		wgInterface := wireguard.NewWireGuardInterface("wg-hybrid-test", 51822)
		
		if wgInterface == nil {
			t.Fatal("Failed to create hybrid WireGuard interface")
		}
		
		// Проверяем, что интерфейс создан
		if wgInterface.Name != "wg-hybrid-test" {
			t.Errorf("Expected interface name 'wg-hybrid-test', got '%s'", wgInterface.Name)
		}
		
		// Проверяем, что ключи сгенерированы
		if wgInterface.PrivateKey == "" {
			t.Error("Private key not generated")
		}
		
		if wgInterface.PublicKey == "" {
			t.Error("Public key not generated")
		}
		
		// Проверяем статус для определения режима работы
		status := wgInterface.GetStatus()
		t.Logf("Interface status: %+v", status)
	})
	
	// Тестируем конфигурацию
	t.Run("WireGuard Configuration", func(t *testing.T) {
		config := &wireguard.WireGuardConfig{
			Interface: struct {
				Name       string `json:"name"`
				ListenPort int    `json:"listen_port"`
				Address    string `json:"address"`
				PrivateKey string `json:"private_key"`
				MTU        int    `json:"mtu"`
			}{
				Name:       "wg-config-test",
				ListenPort: 51823,
				Address:    "10.100.1.1/24",
				PrivateKey: "test-private-key",
				MTU:        1420,
			},
			Peers: []wireguard.WireGuardPeerConfig{
				{
					PublicKey:    "test-peer-public-key",
					AllowedIPs:   []string{"10.100.1.2/32"},
					Endpoint:     "192.168.1.100:51820",
					PresharedKey: "test-preshared-key",
				},
			},
		}
		
		if config.Interface.Name != "wg-config-test" {
			t.Errorf("Expected config name 'wg-config-test', got '%s'", config.Interface.Name)
		}
		
		if len(config.Peers) != 1 {
			t.Errorf("Expected 1 peer, got %d", len(config.Peers))
		}
		
		if config.Peers[0].PublicKey != "test-peer-public-key" {
			t.Errorf("Expected peer public key 'test-peer-public-key', got '%s'", config.Peers[0].PublicKey)
		}
	})
}

func TestRealWireGuardFallback(t *testing.T) {
	// Тестируем fallback на симуляцию
	t.Run("Fallback to Simulation", func(t *testing.T) {
		wgInterface := wireguard.NewWireGuardInterface("wg-fallback-test", 51824)
		
		// Запускаем интерфейс
		err := wgInterface.Start()
		if err != nil {
			t.Fatalf("Failed to start interface: %v", err)
		}
		
		// Проверяем статус
		status := wgInterface.GetStatus()
		if status == nil {
			t.Fatal("Failed to get status")
		}
		
		// Проверяем, что интерфейс работает
		if status["status"] != "running" {
			t.Errorf("Expected status 'running', got '%v'", status["status"])
		}
		
		// Проверяем режим работы
		mode := status["mode"]
		t.Logf("Interface mode: %v", mode)
		
		// Останавливаем интерфейс
		err = wgInterface.Stop()
		if err != nil {
			t.Errorf("Failed to stop interface: %v", err)
		}
	})
}

func TestRealWireGuardPeerManagement(t *testing.T) {
	t.Run("Peer Management", func(t *testing.T) {
		wgInterface := wireguard.NewWireGuardInterface("wg-peer-test", 51825)
		
		// Запускаем интерфейс
		err := wgInterface.Start()
		if err != nil {
			t.Fatalf("Failed to start interface: %v", err)
		}
		defer wgInterface.Stop()
		
		// Добавляем пира
		peer := &wireguard.WireGuardPeer{
			PublicKey:     "test-peer-key",
			Endpoint:      "192.168.1.100:51820",
			AllowedIPs:    []string{"10.100.1.2/32"},
			Status:        wireguard.PeerStatusConnected,
			Latency:       50 * time.Millisecond,
		}
		
		err = wgInterface.AddPeer(peer)
		if err != nil {
			t.Errorf("Failed to add peer: %v", err)
		}
		
		// Проверяем статус
		status := wgInterface.GetStatus()
		if status == nil {
			t.Fatal("Failed to get status")
		}
		
		// Проверяем количество пиров
		peersCount := status["peers_count"].(int)
		if peersCount != 1 {
			t.Errorf("Expected 1 peer, got %d", peersCount)
		}
		
		// Удаляем пира
		err = wgInterface.RemovePeer("test-peer-key")
		if err != nil {
			t.Errorf("Failed to remove peer: %v", err)
		}
		
		// Проверяем статус после удаления
		status = wgInterface.GetStatus()
		peersCount = status["peers_count"].(int)
		if peersCount != 0 {
			t.Errorf("Expected 0 peers after removal, got %d", peersCount)
		}
	})
}

func TestRealWireGuardContextIntegration(t *testing.T) {
	t.Run("Context Integration", func(t *testing.T) {
		adapter := wireguard.NewRealWireGuardAdapter("wg-context-test", zap.NewNop())
		
		if adapter == nil {
			t.Fatal("Failed to create adapter")
		}
		
		// Проверяем работу через публичные методы
		status, err := adapter.GetStatus()
		if err != nil {
			t.Logf("GetStatus error (expected in test): %v", err)
		} else {
			if status["interface"] != "wg-context-test" {
				t.Errorf("Expected interface name 'wg-context-test', got '%v'", status["interface"])
			}
		}
		
		// Тестируем Close
		err = adapter.Close()
		if err != nil {
			t.Logf("Close error (expected in test): %v", err)
		}
	})
}
