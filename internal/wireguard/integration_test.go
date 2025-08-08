package wireguard

import (
	"testing"
	"time"
)

func TestWireGuardMeshIntegration(t *testing.T) {
	// Создаем WireGuard интерфейс
	wgInterface := NewWireGuardInterface("wg0", 51820)
	
	// Создаем mesh topology
	mesh := NewMeshTopology(wgInterface)
	
	// Создаем peer discovery
	config := &DiscoveryConfig{
		RegistryURL:       "https://registry.cloudbridge.com",
		HeartbeatInterval: 30 * time.Second,
		MaxPeers:          50,
	}
	discovery := NewPeerDiscovery(config, wgInterface)
	
	// Тестируем запуск компонентов
	if err := wgInterface.Start(); err != nil {
		t.Errorf("Failed to start WireGuard interface: %v", err)
	}
	
	if err := discovery.Start(); err != nil {
		t.Errorf("Failed to start peer discovery: %v", err)
	}
	
	// Тестируем добавление peer
	peer := &WireGuardPeer{
		PublicKey:    "test-public-key",
		Endpoint:     "192.168.1.100:51820",
		AllowedIPs:   []string{"10.100.1.2/32"},
		LastHandshake: time.Now(),
	}
	
	if err := wgInterface.AddPeer(peer); err != nil {
		t.Errorf("Failed to add peer: %v", err)
	}
	
	if err := mesh.AddPeer("test-peer", peer); err != nil {
		t.Errorf("Failed to add peer to mesh: %v", err)
	}
	
	// Тестируем получение статуса
	status := wgInterface.GetStatus()
	if status["peers_count"] != 1 {
		t.Errorf("Expected 1 peer, got %v", status["peers_count"])
	}
	
	// Тестируем маршрутизацию
	route, err := mesh.GetOptimalRoute("10.100.1.2")
	if err != nil {
		t.Errorf("Failed to get optimal route: %v", err)
	}
	if len(route) == 0 {
		t.Error("Expected non-empty route")
	}
	
	// Очистка
	discovery.Stop()
	wgInterface.Stop()
}

func TestQUICOverWireGuardIntegration(t *testing.T) {
	// Создаем WireGuard интерфейс
	wgInterface := NewWireGuardInterface("wg0", 51820)
	
	// Создаем QUIC over WireGuard
	quicOverWG := NewQUICOverWireGuard(wgInterface, nil)
	
	// Тестируем запуск
	if err := quicOverWG.Start(); err != nil {
		t.Errorf("Failed to start QUIC over WireGuard: %v", err)
	}
	
	// Тестируем создание соединения
	conn, err := quicOverWG.CreateConnection("test-peer")
	if err != nil {
		t.Errorf("Failed to create connection: %v", err)
	}
	if conn == nil {
		t.Error("Expected non-nil connection")
	}
	
	// Очистка
	quicOverWG.Stop()
}

func TestHybridRoutingIntegration(t *testing.T) {
	// Создаем WireGuard интерфейс и mesh
	wgInterface := NewWireGuardInterface("wg0", 51820)
	mesh := NewMeshTopology(wgInterface)
	
	// Создаем hybrid router
	router := NewHybridRouter(nil, mesh)
	
	// Тестируем маршрутизацию
	route, err := router.RouteRequest("192.168.1.100", "10.100.1.2")
	if err != nil {
		t.Errorf("Failed to route request: %v", err)
	}
	if route == "" {
		t.Error("Expected non-empty route")
	}
	
	// Тестируем установку стратегии
	if err := router.SetRoutingStrategy(StrategyPerformanceOptimal); err != nil {
		t.Errorf("Failed to set routing strategy: %v", err)
	}
	
	// Тестируем получение статистики
	stats := router.GetRoutingStats()
	if stats["strategy"] != StrategyPerformanceOptimal {
		t.Errorf("Expected strategy %v, got %v", StrategyPerformanceOptimal, stats["strategy"])
	}
}

func TestSiteToSiteIntegration(t *testing.T) {
	// Создаем WireGuard mesh
	wgInterface := NewWireGuardInterface("wg0", 51820)
	mesh := NewMeshTopology(wgInterface)
	
	// Создаем site-to-site manager
	ssm := NewSiteToSiteManager(mesh)
	
	// Тестируем добавление корпоративной сети
	network := &CorporateNetwork{
		NetworkID:     "office-moscow",
		CIDR:          "192.168.1.0/24",
		TenantID:      "moscow-tenant",
		ConnectedVia:  "wg0",
		VPNReplacement: true,
	}
	
	if err := ssm.AddCorporateNetwork(network); err != nil {
		t.Errorf("Failed to add corporate network: %v", err)
	}
	
	// Тестируем создание маршрута между сайтами
	route, err := ssm.CreateSiteRoute("office-moscow", "office-spb")
	if err != nil {
		t.Errorf("Failed to create site route: %v", err)
	}
	if route.FromNetwork != "office-moscow" {
		t.Errorf("Expected from network 'office-moscow', got %s", route.FromNetwork)
	}
	
	// Тестируем получение статуса VPN замены
	status := ssm.GetVPNReplacementStatus()
	if !status["office-moscow"] {
		t.Error("Expected VPN replacement to be enabled for office-moscow")
	}
}
