package wireguard

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"go.uber.org/zap"
)

// WireGuardInterface представляет WireGuard интерфейс
type WireGuardInterface struct {
	Name       string
	PrivateKey string
	PublicKey  string
	ListenPort int
	Address    string
	Peers      map[string]*WireGuardPeer
	mu         sync.RWMutex
	logger     *zap.Logger
	isRunning  bool

	// Интеграция с WireGuard
	wgctrlAdapter      *WgctrlAdapter
	useKernelWireGuard bool
	realAdapter        *RealWireGuardAdapter
	useRealWireGuard   bool

	// Peer whitelist
	peerWhitelist    *PeerWhitelist
	whitelistEnabled bool

	// Notifications
	notificationManager *P2PNotificationManager
}

// WireGuardPeer представляет WireGuard peer
type WireGuardPeer struct {
	PublicKey     string
	Endpoint      string
	AllowedIPs    []string
	LastHandshake time.Time
	Status        PeerStatus
	Latency       time.Duration
}

type PeerStatus int

const (
	PeerStatusDisconnected PeerStatus = iota
	PeerStatusConnecting
	PeerStatusConnected
	PeerStatusFailed
)

// MeshTopology представляет WireGuard mesh topology
type MeshTopology struct {
	LocalPeerID    string
	ConnectedPeers map[string]*WireGuardPeer
	RoutingTable   *RoutingTable
	mu             sync.RWMutex
	logger         *zap.Logger
}

// RoutingTable представляет routing table для WireGuard mesh
type RoutingTable struct {
	DirectRoutes map[string]string
	MeshRoutes   map[string][]string
	HubFallback  map[string]string
	mu           sync.RWMutex
}

// PeerDiscovery представляет автоматическое обнаружение peers
type PeerDiscovery struct {
	Config          *DiscoveryConfig
	LocalInterface  *WireGuardInterface
	DiscoveredPeers map[string]*DiscoveredPeer
	mu              sync.RWMutex
	logger          *zap.Logger
	ctx             context.Context
	cancel          context.CancelFunc
	isRunning       bool
}

// DiscoveryConfig конфигурация для peer discovery
type DiscoveryConfig struct {
	RegistryURL       string
	HeartbeatInterval time.Duration
	MaxPeers          int
}

// DiscoveredPeer представляет обнаруженный peer
type DiscoveredPeer struct {
	PeerID       string
	PublicKey    string
	Endpoint     string
	Region       string
	Capabilities []string
	LastSeen     time.Time
	TrustLevel   TrustLevel
	Latency      time.Duration
}

type TrustLevel int

const (
	TrustNone TrustLevel = iota
	TrustBasic
	TrustFull
)

// QUICOverWireGuard представляет Enhanced QUIC поверх WireGuard
type QUICOverWireGuard struct {
	WgInterface    *WireGuardInterface
	MeshTopology   *MeshTopology
	EnhancedConfig interface{} // *quic.EnhancedConfig
	connections    map[string]interface{}
	mu             sync.RWMutex
	logger         *zap.Logger
}

// HybridRouter представляет hybrid routing (WireGuard + Enhanced QUIC)
type HybridRouter struct {
	EnhancedBalancer interface{} // *balancer.Balancer
	WireGuardMesh    *MeshTopology
	RoutingStrategy  RoutingStrategy
	stats            *RoutingStats
	mu               sync.RWMutex
	logger           *zap.Logger
}

type RoutingStats struct {
	TotalRequests      int64
	WireGuardRoutes    int64
	EnhancedQUICRoutes int64
	LastReset          time.Time
}

// RoutingStrategy тип стратегии маршрутизации
type RoutingStrategy int

const (
	StrategyWireGuardFirst RoutingStrategy = iota
	StrategyPerformanceOptimal
	StrategyGeographicOptimal
	StrategyHybrid
)

// SiteToSiteManager управляет site-to-site connectivity
type SiteToSiteManager struct {
	WireGuardMesh     *MeshTopology
	CorporateNetworks map[string]*CorporateNetwork
	SiteRoutes        map[string]*SiteRoute
	mu                sync.RWMutex
	logger            *zap.Logger
}

// CorporateNetwork представляет корпоративную сеть
type CorporateNetwork struct {
	NetworkID      string
	CIDR           string
	TenantID       string
	ConnectedVia   string
	VPNReplacement bool
	Status         NetworkStatus
}

type NetworkStatus int

const (
	NetworkStatusDisconnected NetworkStatus = iota
	NetworkStatusConnecting
	NetworkStatusConnected
	NetworkStatusFailed
)

// SiteRoute представляет маршрут между сайтами
type SiteRoute struct {
	FromNetwork  string
	ToNetwork    string
	ViaWireGuard string
	BackupRoute  string
	Status       RouteStatus
	Latency      time.Duration
}

type RouteStatus int

const (
	RouteStatusInactive RouteStatus = iota
	RouteStatusActive
	RouteStatusFailed
)

// NewWireGuardInterface создает новый WireGuard интерфейс
func NewWireGuardInterface(name string, port int) *WireGuardInterface {
	// Генерируем ключи
	privateKey, publicKey := generateKeyPair()

	// Пытаемся использовать kernel WireGuard через wgctrl
	var (
		wgctrlAdapter      *WgctrlAdapter
		useKernelWireGuard bool
	)
	if wa, err := NewWgctrlAdapter(name, zap.NewNop()); err == nil {
		wgctrlAdapter = wa
		useKernelWireGuard = true
	}

	// Если wgctrl недоступен, проверим userspace wireguard-go
	var realAdapter *RealWireGuardAdapter
	useRealWireGuard := false
	if !useKernelWireGuard && checkRealWireGuardAvailability() {
		realAdapter = NewRealWireGuardAdapter(name, zap.NewNop())
		useRealWireGuard = true
	}

	// Создаем peer whitelist
	peerWhitelist := NewPeerWhitelist(zap.NewNop())

	// Создаем notification manager
	notificationManager := NewP2PNotificationManager(zap.NewNop())

	return &WireGuardInterface{
		Name:                name,
		ListenPort:          port,
		PrivateKey:          privateKey,
		PublicKey:           publicKey,
		Peers:               make(map[string]*WireGuardPeer),
		logger:              zap.NewNop(),
		wgctrlAdapter:       wgctrlAdapter,
		useKernelWireGuard:  useKernelWireGuard,
		realAdapter:         realAdapter,
		useRealWireGuard:    useRealWireGuard,
		peerWhitelist:       peerWhitelist,
		whitelistEnabled:    false, // По умолчанию отключен
		notificationManager: notificationManager,
	}
}

// generateKeyPair генерирует пару ключей WireGuard
func generateKeyPair() (string, string) {
	// В реальной реализации здесь будет генерация WireGuard ключей
	// Пока используем простую симуляцию
	privateKeyBytes := make([]byte, 32)
	if _, err := rand.Read(privateKeyBytes); err != nil {
		return "", ""
	}

	publicKeyBytes := sha256.Sum256(privateKeyBytes)

	return base64.StdEncoding.EncodeToString(privateKeyBytes),
		base64.StdEncoding.EncodeToString(publicKeyBytes[:])
}

// checkRealWireGuardAvailability проверяет доступность реального WireGuard
func checkRealWireGuardAvailability() bool {
	// Проверяем наличие wireguard-go
	if _, err := exec.LookPath("wireguard-go"); err != nil {
		return false
	}

	// Проверяем наличие wg команды
	if _, err := exec.LookPath("wg"); err != nil {
		return false
	}

	// Проверяем наличие ip команды
	if _, err := exec.LookPath("ip"); err != nil {
		return false
	}

	return true
}

// CheckRealWireGuardAvailability экспортированная функция для проверки доступности
func CheckRealWireGuardAvailability() bool {
	return checkRealWireGuardAvailability()
}

// NewMeshTopology создает новую mesh topology
func NewMeshTopology(wgInterface *WireGuardInterface) *MeshTopology {
	return &MeshTopology{
		LocalPeerID:    generatePeerID(),
		ConnectedPeers: make(map[string]*WireGuardPeer),
		RoutingTable: &RoutingTable{
			DirectRoutes: make(map[string]string),
			MeshRoutes:   make(map[string][]string),
			HubFallback:  make(map[string]string),
		},
		logger: zap.NewNop(),
	}
}

// generatePeerID генерирует уникальный ID для peer
func generatePeerID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("peer-%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("peer-%x", bytes)
}

// NewPeerDiscovery создает новый peer discovery
func NewPeerDiscovery(config *DiscoveryConfig, wgInterface *WireGuardInterface) *PeerDiscovery {
	ctx, cancel := context.WithCancel(context.Background())

	return &PeerDiscovery{
		Config:          config,
		LocalInterface:  wgInterface,
		DiscoveredPeers: make(map[string]*DiscoveredPeer),
		ctx:             ctx,
		cancel:          cancel,
		logger:          zap.NewNop(),
	}
}

// NewQUICOverWireGuard создает новый QUIC over WireGuard
func NewQUICOverWireGuard(wgInterface *WireGuardInterface, enhancedConfig interface{}) *QUICOverWireGuard {
	return &QUICOverWireGuard{
		WgInterface:    wgInterface,
		EnhancedConfig: enhancedConfig,
		connections:    make(map[string]interface{}),
		logger:         zap.NewNop(),
	}
}

// NewHybridRouter создает новый hybrid router
func NewHybridRouter(enhancedBalancer interface{}, wireGuardMesh *MeshTopology) *HybridRouter {
	return &HybridRouter{
		EnhancedBalancer: enhancedBalancer,
		WireGuardMesh:    wireGuardMesh,
		RoutingStrategy:  StrategyHybrid,
		stats: &RoutingStats{
			LastReset: time.Now(),
		},
		logger: zap.NewNop(),
	}
}

// NewSiteToSiteManager создает новый site-to-site manager
func NewSiteToSiteManager(wireGuardMesh *MeshTopology) *SiteToSiteManager {
	return &SiteToSiteManager{
		WireGuardMesh:     wireGuardMesh,
		CorporateNetworks: make(map[string]*CorporateNetwork),
		SiteRoutes:        make(map[string]*SiteRoute),
		logger:            zap.NewNop(),
	}
}

// WireGuardInterface методы
func (wgi *WireGuardInterface) Start() error {
	wgi.mu.Lock()
	defer wgi.mu.Unlock()

	if wgi.isRunning {
		return fmt.Errorf("WireGuard interface already running")
	}

	// 1) Пытаемся запустить kernel WireGuard через wgctrl
	if wgi.useKernelWireGuard && wgi.wgctrlAdapter != nil {
		cfg := &WireGuardConfig{
			Interface: struct {
				Name       string `json:"name"`
				ListenPort int    `json:"listen_port"`
				Address    string `json:"address"`
				PrivateKey string `json:"private_key"`
				MTU        int    `json:"mtu"`
			}{
				Name:       wgi.Name,
				ListenPort: wgi.ListenPort,
				Address:    wgi.Address,
				PrivateKey: wgi.PrivateKey,
				MTU:        1420,
			},
			Peers: []WireGuardPeerConfig{},
		}
		for _, peer := range wgi.Peers {
			cfg.Peers = append(cfg.Peers, WireGuardPeerConfig{
				PublicKey:  peer.PublicKey,
				AllowedIPs: peer.AllowedIPs,
				Endpoint:   peer.Endpoint,
			})
		}
		if err := wgi.wgctrlAdapter.Start(cfg); err != nil {
			wgi.logger.Error("Failed to start kernel WireGuard via wgctrl, trying userspace", zap.Error(err))
		} else {
			wgi.isRunning = true
			wgi.logger.Info("Kernel WireGuard interface started (wgctrl)",
				zap.String("name", wgi.Name),
				zap.Int("port", wgi.ListenPort),
			)
			return nil
		}
	}

	// 2) Fallback: userspace wireguard-go
	if wgi.useRealWireGuard && wgi.realAdapter != nil {
		// Используем реальный WireGuard
		config := &WireGuardConfig{
			Interface: struct {
				Name       string `json:"name"`
				ListenPort int    `json:"listen_port"`
				Address    string `json:"address"`
				PrivateKey string `json:"private_key"`
				MTU        int    `json:"mtu"`
			}{
				Name:       wgi.Name,
				ListenPort: wgi.ListenPort,
				Address:    wgi.Address,
				PrivateKey: wgi.PrivateKey,
				MTU:        1420,
			},
			Peers: []WireGuardPeerConfig{},
		}

		// Добавляем существующих пиров
		for _, peer := range wgi.Peers {
			config.Peers = append(config.Peers, WireGuardPeerConfig{
				PublicKey:  peer.PublicKey,
				AllowedIPs: peer.AllowedIPs,
				Endpoint:   peer.Endpoint,
			})
		}

		if err := wgi.realAdapter.Start(config); err != nil {
			wgi.logger.Error("Failed to start real WireGuard, falling back to simulation", zap.Error(err))
			wgi.useRealWireGuard = false
		} else {
			wgi.isRunning = true
			wgi.logger.Info("Real WireGuard interface started",
				zap.String("name", wgi.Name),
				zap.Int("port", wgi.ListenPort),
			)
			return nil
		}
	}

	// 3) Fallback на симуляцию
	wgi.isRunning = true
	wgi.logger.Info("Simulated WireGuard interface started",
		zap.String("name", wgi.Name),
		zap.Int("port", wgi.ListenPort),
	)

	return nil
}

func (wgi *WireGuardInterface) Stop() error {
	wgi.mu.Lock()
	defer wgi.mu.Unlock()

	if !wgi.isRunning {
		return fmt.Errorf("WireGuard interface not running")
	}

	if wgi.useRealWireGuard && wgi.realAdapter != nil {
		// Останавливаем реальный WireGuard
		if err := wgi.realAdapter.Stop(); err != nil {
			wgi.logger.Error("Failed to stop real WireGuard", zap.Error(err))
		}
	}

	wgi.isRunning = false
	wgi.logger.Info("WireGuard interface stopped",
		zap.String("name", wgi.Name),
	)

	return nil
}

func (wgi *WireGuardInterface) AddPeer(peer *WireGuardPeer) error {
	wgi.mu.Lock()
	defer wgi.mu.Unlock()

	// Проверяем whitelist если включен
	if wgi.whitelistEnabled && wgi.peerWhitelist != nil {
		if !wgi.peerWhitelist.IsPeerAllowed(peer.PublicKey) {
			wgi.logger.Warn("Peer not allowed by whitelist",
				zap.String("public_key", peer.PublicKey),
			)
			return fmt.Errorf("peer %s not in whitelist", peer.PublicKey)
		}
	}

	if wgi.useRealWireGuard && wgi.realAdapter != nil && wgi.isRunning {
		// Добавляем пира в реальный WireGuard
		realPeer := &WireGuardPeerConfig{
			PublicKey:  peer.PublicKey,
			AllowedIPs: peer.AllowedIPs,
			Endpoint:   peer.Endpoint,
		}

		if err := wgi.realAdapter.AddPeer(realPeer); err != nil {
			wgi.logger.Error("Failed to add peer to real WireGuard", zap.Error(err))
			// Fallback на симуляцию
			wgi.Peers[peer.PublicKey] = peer
		} else {
			wgi.logger.Info("Peer added to real WireGuard",
				zap.String("public_key", peer.PublicKey),
				zap.String("endpoint", peer.Endpoint),
			)
			return nil
		}
	} else {
		// Используем симуляцию
		wgi.Peers[peer.PublicKey] = peer
		wgi.logger.Info("Peer added to simulated WireGuard",
			zap.String("public_key", peer.PublicKey),
			zap.String("endpoint", peer.Endpoint),
		)
	}

	return nil
}

func (wgi *WireGuardInterface) RemovePeer(publicKey string) error {
	wgi.mu.Lock()
	defer wgi.mu.Unlock()

	if wgi.useRealWireGuard && wgi.realAdapter != nil && wgi.isRunning {
		// Удаляем пира из реального WireGuard
		if err := wgi.realAdapter.RemovePeer(publicKey); err != nil {
			wgi.logger.Error("Failed to remove peer from real WireGuard", zap.Error(err))
		}
	}

	// Удаляем из симуляции
	delete(wgi.Peers, publicKey)
	wgi.logger.Info("WireGuard peer removed",
		zap.String("public_key", publicKey),
	)

	return nil
}

// SetWhitelistEnabled enables or disables peer whitelist
func (wgi *WireGuardInterface) SetWhitelistEnabled(enabled bool) {
	wgi.mu.Lock()
	defer wgi.mu.Unlock()

	wgi.whitelistEnabled = enabled
	wgi.logger.Info("Peer whitelist status changed",
		zap.Bool("enabled", enabled),
	)
}

// LoadPeerWhitelist loads peer whitelist from JWT claims
func (wgi *WireGuardInterface) LoadPeerWhitelist(peerWhitelist []string) error {
	if wgi.peerWhitelist == nil {
		return fmt.Errorf("peer whitelist not initialized")
	}

	return wgi.peerWhitelist.LoadFromJWT(peerWhitelist)
}

// GetPeerWhitelist returns the peer whitelist
func (wgi *WireGuardInterface) GetPeerWhitelist() *PeerWhitelist {
	return wgi.peerWhitelist
}

// GetNotificationManager returns the notification manager
func (wgi *WireGuardInterface) GetNotificationManager() *P2PNotificationManager {
	return wgi.notificationManager
}

// NotifyP2PUnavailable notifies about P2P unavailability
func (wgi *WireGuardInterface) NotifyP2PUnavailable(reason, fallback string) {
	if wgi.notificationManager != nil {
		wgi.notificationManager.NotifyP2PUnavailable(reason, fallback)
	}
}

func (wgi *WireGuardInterface) GetStatus() map[string]interface{} {
	wgi.mu.RLock()
	defer wgi.mu.RUnlock()

	if wgi.useRealWireGuard && wgi.realAdapter != nil && wgi.isRunning {
		// Получаем статус реального WireGuard
		status, err := wgi.realAdapter.GetStatus()
		if err != nil {
			wgi.logger.Error("Failed to get real WireGuard status", zap.Error(err))
		} else {
			status["mode"] = "real"
			return status
		}
	}

	// Fallback на симуляцию
	peers := make([]map[string]interface{}, 0)
	for _, peer := range wgi.Peers {
		peers = append(peers, map[string]interface{}{
			"public_key":     peer.PublicKey,
			"endpoint":       peer.Endpoint,
			"status":         peer.Status,
			"last_handshake": peer.LastHandshake,
			"latency":        peer.Latency,
		})
	}

	return map[string]interface{}{
		"name":                     wgi.Name,
		"listen_port":              wgi.ListenPort,
		"address":                  wgi.Address,
		"peers_count":              len(wgi.Peers),
		"status":                   "running",
		"is_running":               wgi.isRunning,
		"peers":                    peers,
		"mode":                     "simulated",
		"real_wireguard_available": wgi.useRealWireGuard,
	}
}

// MeshTopology методы
func (mt *MeshTopology) AddPeer(peerID string, peer *WireGuardPeer) error {
	mt.mu.Lock()
	defer mt.mu.Unlock()

	mt.ConnectedPeers[peerID] = peer
	mt.logger.Info("Peer added to mesh topology",
		zap.String("peer_id", peerID),
		zap.String("public_key", peer.PublicKey),
	)

	// Обновляем routing table
	return mt.updateRoutingTable()
}

func (mt *MeshTopology) RemovePeer(peerID string) error {
	mt.mu.Lock()
	defer mt.mu.Unlock()

	delete(mt.ConnectedPeers, peerID)
	mt.logger.Info("Peer removed from mesh topology",
		zap.String("peer_id", peerID),
	)

	// Обновляем routing table
	return mt.updateRoutingTable()
}

func (mt *MeshTopology) GetOptimalRoute(destination string) ([]string, error) {
	mt.mu.RLock()
	defer mt.mu.RUnlock()

	// Проверяем прямые маршруты
	if route, exists := mt.RoutingTable.DirectRoutes[destination]; exists {
		return []string{route}, nil
	}

	// Проверяем mesh маршруты
	if routes, exists := mt.RoutingTable.MeshRoutes[destination]; exists {
		return routes, nil
	}

	// Fallback маршрут
	if fallback, exists := mt.RoutingTable.HubFallback[destination]; exists {
		return []string{fallback}, nil
	}

	// Дефолтный маршрут
	return []string{"default"}, nil
}

func (mt *MeshTopology) UpdateRoutingTable() error {
	mt.mu.Lock()
	defer mt.mu.Unlock()

	return mt.updateRoutingTable()
}

func (mt *MeshTopology) updateRoutingTable() error {
	// Очищаем старые маршруты
	mt.RoutingTable.mu.Lock()
	defer mt.RoutingTable.mu.Unlock()

	mt.RoutingTable.DirectRoutes = make(map[string]string)
	mt.RoutingTable.MeshRoutes = make(map[string][]string)
	mt.RoutingTable.HubFallback = make(map[string]string)

	// Строим новые маршруты на основе подключенных пиров
	for peerID, peer := range mt.ConnectedPeers {
		if peer.Status == PeerStatusConnected {
			// Добавляем прямые маршруты для каждого peer
			for _, allowedIP := range peer.AllowedIPs {
				mt.RoutingTable.DirectRoutes[allowedIP] = peerID
			}
		}
	}

	mt.logger.Info("Mesh routing table updated",
		zap.Int("direct_routes", len(mt.RoutingTable.DirectRoutes)),
		zap.Int("mesh_routes", len(mt.RoutingTable.MeshRoutes)),
	)

	return nil
}

// PeerDiscovery методы
func (pd *PeerDiscovery) Start() error {
	pd.mu.Lock()
	defer pd.mu.Unlock()

	if pd.isRunning {
		return fmt.Errorf("Peer discovery already running")
	}

	pd.isRunning = true

	// Запускаем goroutine для peer discovery
	go pd.discoveryLoop()

	pd.logger.Info("Peer discovery started",
		zap.String("registry_url", pd.Config.RegistryURL),
		zap.Duration("heartbeat_interval", pd.Config.HeartbeatInterval),
	)

	return nil
}

func (pd *PeerDiscovery) Stop() error {
	pd.mu.Lock()
	defer pd.mu.Unlock()

	if !pd.isRunning {
		return fmt.Errorf("Peer discovery not running")
	}

	pd.cancel()
	pd.isRunning = false

	pd.logger.Info("Peer discovery stopped")

	return nil
}

func (pd *PeerDiscovery) discoveryLoop() {
	ticker := time.NewTicker(pd.Config.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-pd.ctx.Done():
			return
		case <-ticker.C:
			pd.discoverPeers()
			pd.registerLocalPeer()
		}
	}
}

func (pd *PeerDiscovery) discoverPeers() {
	// В реальной реализации здесь будет HTTP запрос к registry
	// для получения списка доступных пиров

	pd.mu.Lock()
	defer pd.mu.Unlock()

	// Симуляция обнаружения пиров
	discoveredPeers := []*DiscoveredPeer{
		{
			PeerID:       "peer-remote-1",
			PublicKey:    "remote-public-key-1",
			Endpoint:     "192.168.1.100:51820",
			Region:       "eu-west",
			Capabilities: []string{"p2p", "mesh", "site-to-site"},
			LastSeen:     time.Now(),
			TrustLevel:   TrustBasic,
			Latency:      50 * time.Millisecond,
		},
		{
			PeerID:       "peer-remote-2",
			PublicKey:    "remote-public-key-2",
			Endpoint:     "192.168.1.101:51820",
			Region:       "eu-central",
			Capabilities: []string{"p2p", "mesh"},
			LastSeen:     time.Now(),
			TrustLevel:   TrustBasic,
			Latency:      75 * time.Millisecond,
		},
	}

	for _, peer := range discoveredPeers {
		pd.DiscoveredPeers[peer.PeerID] = peer
	}

	pd.logger.Debug("Discovered peers",
		zap.Int("count", len(discoveredPeers)),
	)
}

func (pd *PeerDiscovery) registerLocalPeer() {
	// В реальной реализации здесь будет регистрация локального пира в registry
	pd.logger.Debug("Registering local peer with registry")
}

func (pd *PeerDiscovery) DiscoverPeers() ([]*DiscoveredPeer, error) {
	pd.mu.RLock()
	defer pd.mu.RUnlock()

	peers := make([]*DiscoveredPeer, 0, len(pd.DiscoveredPeers))
	for _, peer := range pd.DiscoveredPeers {
		peers = append(peers, peer)
	}

	return peers, nil
}

func (pd *PeerDiscovery) RegisterLocalPeer() error {
	// Регистрация уже происходит в discoveryLoop
	return nil
}

// QUICOverWireGuard методы
func (qowg *QUICOverWireGuard) Start() error {
	qowg.logger.Info("QUIC over WireGuard started")
	return nil
}

func (qowg *QUICOverWireGuard) Stop() error {
	qowg.logger.Info("QUIC over WireGuard stopped")
	return nil
}

func (qowg *QUICOverWireGuard) CreateConnection(peerID string) (interface{}, error) {
	qowg.mu.Lock()
	defer qowg.mu.Unlock()

	// В реальной реализации здесь будет создание QUIC соединения поверх WireGuard
	connection := fmt.Sprintf("quic-wg-connection-%s", peerID)
	qowg.connections[peerID] = connection

	qowg.logger.Info("QUIC over WireGuard connection created",
		zap.String("peer_id", peerID),
	)

	return connection, nil
}

// HybridRouter методы
func (hr *HybridRouter) RouteRequest(clientIP, destination string) (string, error) {
	hr.mu.Lock()
	defer hr.mu.Unlock()

	hr.stats.TotalRequests++

	// Проверяем доступность WireGuard маршрута
	if hr.WireGuardMesh != nil {
		routes, err := hr.WireGuardMesh.GetOptimalRoute(destination)
		if err == nil && len(routes) > 0 {
			hr.stats.WireGuardRoutes++
			return "wireguard", nil
		}
	}

	// Fallback на Enhanced QUIC
	hr.stats.EnhancedQUICRoutes++
	return "enhanced_quic", nil
}

func (hr *HybridRouter) SetRoutingStrategy(strategy RoutingStrategy) error {
	hr.mu.Lock()
	defer hr.mu.Unlock()

	hr.RoutingStrategy = strategy
	hr.logger.Info("Routing strategy updated",
		zap.Int("strategy", int(strategy)),
	)

	return nil
}

func (hr *HybridRouter) GetRoutingStats() map[string]interface{} {
	hr.mu.RLock()
	defer hr.mu.RUnlock()

	return map[string]interface{}{
		"strategy":             hr.RoutingStrategy,
		"wireguard_routes":     hr.stats.WireGuardRoutes,
		"enhanced_quic_routes": hr.stats.EnhancedQUICRoutes,
		"total_requests":       hr.stats.TotalRequests,
		"last_reset":           hr.stats.LastReset,
	}
}

// SiteToSiteManager методы
func (ssm *SiteToSiteManager) AddCorporateNetwork(network *CorporateNetwork) error {
	ssm.mu.Lock()
	defer ssm.mu.Unlock()

	ssm.CorporateNetworks[network.NetworkID] = network
	ssm.logger.Info("Corporate network added",
		zap.String("network_id", network.NetworkID),
		zap.String("cidr", network.CIDR),
		zap.String("tenant_id", network.TenantID),
	)

	return nil
}

func (ssm *SiteToSiteManager) CreateSiteRoute(from, to string) (*SiteRoute, error) {
	ssm.mu.Lock()
	defer ssm.mu.Unlock()

	route := &SiteRoute{
		FromNetwork:  from,
		ToNetwork:    to,
		ViaWireGuard: "wg0",
		BackupRoute:  "quic",
		Status:       RouteStatusActive,
		Latency:      25 * time.Millisecond,
	}

	routeKey := fmt.Sprintf("%s-%s", from, to)
	ssm.SiteRoutes[routeKey] = route

	ssm.logger.Info("Site-to-site route created",
		zap.String("from", from),
		zap.String("to", to),
		zap.String("via", route.ViaWireGuard),
	)

	return route, nil
}

func (ssm *SiteToSiteManager) GetVPNReplacementStatus() map[string]bool {
	ssm.mu.RLock()
	defer ssm.mu.RUnlock()

	status := make(map[string]bool)
	for _, network := range ssm.CorporateNetworks {
		status[network.NetworkID] = network.VPNReplacement
	}

	return status
}
