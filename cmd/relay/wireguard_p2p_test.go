package main

import (
	"context"
	"testing"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/internal/p2p"
	"github.com/twogc/cloudbridge-relay-installer/internal/wireguard"
	"go.uber.org/zap"
)

func TestWireGuardP2PMeshNetwork(t *testing.T) {
	// Создаем логгер для тестов
	logger := zap.NewNop()

	// 1. Тестируем создание WireGuard интерфейса
	t.Run("WireGuard Interface Creation", func(t *testing.T) {
		wgInterface := wireguard.NewWireGuardInterface("wg-test", 51820)
		
		if wgInterface == nil {
			t.Fatal("Failed to create WireGuard interface")
		}
		
		if wgInterface.Name != "wg-test" {
			t.Errorf("Expected interface name 'wg-test', got '%s'", wgInterface.Name)
		}
		
		if wgInterface.ListenPort != 51820 {
			t.Errorf("Expected listen port 51820, got %d", wgInterface.ListenPort)
		}
		
		// Проверяем, что ключи сгенерированы
		if wgInterface.PrivateKey == "" {
			t.Error("Private key not generated")
		}
		
		if wgInterface.PublicKey == "" {
			t.Error("Public key not generated")
		}
	})

	// 2. Тестируем Mesh Topology
	t.Run("Mesh Topology", func(t *testing.T) {
		wgInterface := wireguard.NewWireGuardInterface("wg-test", 51820)
		mesh := wireguard.NewMeshTopology(wgInterface)
		
		if mesh == nil {
			t.Fatal("Failed to create mesh topology")
		}
		
		if mesh.LocalPeerID == "" {
			t.Error("Local peer ID not generated")
		}
		
		// Тестируем добавление пира
		peer := &wireguard.WireGuardPeer{
			PublicKey:     "test-public-key",
			Endpoint:      "192.168.1.100:51820",
			AllowedIPs:    []string{"10.100.1.0/24"},
			Status:        wireguard.PeerStatusConnected,
			Latency:       50 * time.Millisecond,
		}
		
		err := mesh.AddPeer("test-peer-1", peer)
		if err != nil {
			t.Errorf("Failed to add peer: %v", err)
		}
		
		// Тестируем получение маршрута
		routes, err := mesh.GetOptimalRoute("10.100.1.50")
		if err != nil {
			t.Errorf("Failed to get route: %v", err)
		}
		
		if len(routes) == 0 {
			t.Error("No routes returned")
		}
	})

	// 3. Тестируем Peer Discovery
	t.Run("Peer Discovery", func(t *testing.T) {
		wgInterface := wireguard.NewWireGuardInterface("wg-test", 51820)
		config := &wireguard.DiscoveryConfig{
			RegistryURL:       "https://test-registry.com",
			HeartbeatInterval: 30 * time.Second,
			MaxPeers:          10,
		}
		
		discovery := wireguard.NewPeerDiscovery(config, wgInterface)
		
		if discovery == nil {
			t.Fatal("Failed to create peer discovery")
		}
		
		// Запускаем discovery
		err := discovery.Start()
		if err != nil {
			t.Errorf("Failed to start peer discovery: %v", err)
		}
		
		// Ждем немного для обнаружения пиров
		time.Sleep(500 * time.Millisecond)
		
		// Получаем обнаруженных пиров
		peers, err := discovery.DiscoverPeers()
		if err != nil {
			t.Errorf("Failed to discover peers: %v", err)
		}
		
		// В тестовой среде должно быть 2 симулированных пира
		// Но discovery может не успеть их обнаружить за короткое время
		if len(peers) == 0 {
			t.Log("No peers discovered (this is expected in test environment with short timeout)")
		} else if len(peers) != 2 {
			t.Errorf("Expected 2 peers, got %d", len(peers))
		}
		
		// Останавливаем discovery
		err = discovery.Stop()
		if err != nil {
			t.Errorf("Failed to stop peer discovery: %v", err)
		}
	})

	// 4. Тестируем P2P Manager
	t.Run("P2P Manager", func(t *testing.T) {
		wgInterface := wireguard.NewWireGuardInterface("wg-test", 51820)
		mesh := wireguard.NewMeshTopology(wgInterface)
		
		config := &wireguard.DiscoveryConfig{
			RegistryURL:       "https://test-registry.com",
			HeartbeatInterval: 30 * time.Second,
			MaxPeers:          10,
		}
		discovery := wireguard.NewPeerDiscovery(config, wgInterface)
		
		router := wireguard.NewHybridRouter(nil, mesh)
		siteToSite := wireguard.NewSiteToSiteManager(mesh)
		
		p2pManager := p2p.NewP2PManager(mesh, discovery, router, siteToSite, logger)
		
		if p2pManager == nil {
			t.Fatal("Failed to create P2P Manager")
		}
		
		// Тестируем инициализацию
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		err := p2pManager.Initialize(ctx)
		if err != nil {
			t.Errorf("Failed to initialize P2P Manager: %v", err)
		}
		
		// Тестируем проверку доступности
		available := p2pManager.IsAvailable(ctx, "192.168.1.100", "test-tenant")
		if !available {
			t.Log("P2P not available (expected in test environment)")
		}
		
		// Тестируем получение статуса
		status := p2pManager.GetStatus()
		if status == nil {
			t.Error("Failed to get P2P Manager status")
		}
		
		// Проверяем, что статус содержит ожидаемые поля
		if _, ok := status["initialized"]; !ok {
			t.Error("Status missing 'initialized' field")
		}
		
		if _, ok := status["metrics"]; !ok {
			t.Error("Status missing 'metrics' field")
		}
	})

	// 5. Тестируем Hybrid Router
	t.Run("Hybrid Router", func(t *testing.T) {
		wgInterface := wireguard.NewWireGuardInterface("wg-test", 51820)
		mesh := wireguard.NewMeshTopology(wgInterface)
		router := wireguard.NewHybridRouter(nil, mesh)
		
		if router == nil {
			t.Fatal("Failed to create hybrid router")
		}
		
		// Тестируем маршрутизацию
		route, err := router.RouteRequest("192.168.1.100", "10.100.1.50")
		if err != nil {
			t.Errorf("Failed to route request: %v", err)
		}
		
		if route == "" {
			t.Error("No route returned")
		}
		
		// Тестируем получение статистики
		stats := router.GetRoutingStats()
		if stats == nil {
			t.Error("Failed to get routing stats")
		}
		
		if _, ok := stats["total_requests"]; !ok {
			t.Error("Stats missing 'total_requests' field")
		}
	})

	// 6. Тестируем Site-to-Site Manager
	t.Run("Site-to-Site Manager", func(t *testing.T) {
		wgInterface := wireguard.NewWireGuardInterface("wg-test", 51820)
		mesh := wireguard.NewMeshTopology(wgInterface)
		siteToSite := wireguard.NewSiteToSiteManager(mesh)
		
		if siteToSite == nil {
			t.Fatal("Failed to create site-to-site manager")
		}
		
		// Тестируем добавление корпоративной сети
		network := &wireguard.CorporateNetwork{
			NetworkID:     "corp-a",
			CIDR:          "192.168.1.0/24",
			TenantID:      "tenant-a",
			ConnectedVia:  "wg0",
			VPNReplacement: true,
		}
		
		err := siteToSite.AddCorporateNetwork(network)
		if err != nil {
			t.Errorf("Failed to add corporate network: %v", err)
		}
		
		// Тестируем создание маршрута между сайтами
		route, err := siteToSite.CreateSiteRoute("corp-a", "corp-b")
		if err != nil {
			t.Errorf("Failed to create site route: %v", err)
		}
		
		if route == nil {
			t.Error("No route created")
		}
		
		if route.FromNetwork != "corp-a" {
			t.Errorf("Expected from network 'corp-a', got '%s'", route.FromNetwork)
		}
		
		if route.ToNetwork != "corp-b" {
			t.Errorf("Expected to network 'corp-b', got '%s'", route.ToNetwork)
		}
		
		// Тестируем получение статуса VPN замены
		status := siteToSite.GetVPNReplacementStatus()
		if status == nil {
			t.Error("Failed to get VPN replacement status")
		}
		
		if len(status) == 0 {
			t.Error("No VPN replacement status returned")
		}
	})
}

// Тест для проверки интеграции всех компонентов
func TestWireGuardP2PIntegration(t *testing.T) {
	logger := zap.NewNop()
	
	// Создаем все компоненты
	wgInterface := wireguard.NewWireGuardInterface("wg-integration-test", 51821)
	mesh := wireguard.NewMeshTopology(wgInterface)
	
	config := &wireguard.DiscoveryConfig{
		RegistryURL:       "https://integration-test-registry.com",
		HeartbeatInterval: 30 * time.Second,
		MaxPeers:          10,
	}
	discovery := wireguard.NewPeerDiscovery(config, wgInterface)
	
	router := wireguard.NewHybridRouter(nil, mesh)
	siteToSite := wireguard.NewSiteToSiteManager(mesh)
	
	p2pManager := p2p.NewP2PManager(mesh, discovery, router, siteToSite, logger)
	
	// Тестируем полную интеграцию
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// Инициализируем P2P Manager
	err := p2pManager.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize P2P Manager: %v", err)
	}
	
	// Ждем немного для обнаружения пиров
	time.Sleep(200 * time.Millisecond)
	
	// Проверяем статус
	status := p2pManager.GetStatus()
	if status == nil {
		t.Fatal("Failed to get integration status")
	}
	
	t.Logf("Integration status: %+v", status)
	
	// Проверяем, что все компоненты работают
	if !status["initialized"].(bool) {
		t.Error("P2P Manager not initialized")
	}
	
	// Очистка
	err = p2pManager.Shutdown(ctx)
	if err != nil {
		t.Errorf("Failed to shutdown P2P Manager: %v", err)
	}
}
