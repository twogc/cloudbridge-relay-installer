package main

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/config"
	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// SiteToSiteConnection представляет site-to-site VPN соединение
type SiteToSiteConnection struct {
	ID         string
	Name       string
	LocalSite  *Site
	RemoteSite *Site
	Status     string
	CreatedAt  time.Time
	LastSeen   time.Time
	Bandwidth  uint64
	IsActive   bool
}

// Site представляет VPN site
type Site struct {
	ID          string
	Name        string
	PublicKey   string
	Endpoint    string
	AllowedIPs  []string
	Description string
}

// RoutingRule представляет правило маршрутизации
type RoutingRule struct {
	ID          string
	Source      string
	Destination string
	Gateway     string
	Interface   string
	Priority    int
	IsActive    bool
}

// WGProfileManager управляет WireGuard интерфейсами и пирами (на основе wgctrl-go)
type WGProfileManager struct {
	client  *wgctrl.Client
	config  *config.Config
	logger  *zap.Logger
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	mu      sync.RWMutex
	devices map[string]*wgtypes.Device
	peers   map[string]wgtypes.PeerConfig

	// Site-to-site VPN management
	siteToSiteConnections map[string]*SiteToSiteConnection
	routingRules          map[string]*RoutingRule
}

// NewWGProfileManager создает новый менеджер WG-профилей
func NewWGProfileManager(cfg *config.Config, logger *zap.Logger) (*WGProfileManager, error) {
	client, err := wgctrl.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create wgctrl client: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &WGProfileManager{
		client:                client,
		config:                cfg,
		logger:                logger,
		ctx:                   ctx,
		cancel:                cancel,
		devices:               make(map[string]*wgtypes.Device),
		peers:                 make(map[string]wgtypes.PeerConfig),
		siteToSiteConnections: make(map[string]*SiteToSiteConnection),
		routingRules:          make(map[string]*RoutingRule),
	}, nil
}

// Start запускает WG-профиль менеджер
func (w *WGProfileManager) Start() error {
	if !w.config.WireGuard.Enabled {
		w.logger.Info("WireGuard disabled, skipping WG profile manager")
		return nil
	}

	w.logger.Info("Starting WG Profile Manager",
		zap.String("interface", w.config.WireGuard.Interface.Name),
		zap.Int("port", w.config.WireGuard.Interface.ListenPort),
		zap.String("address", w.config.WireGuard.Interface.Address),
	)

	// Запускаем мониторинг устройств
	w.wg.Add(1)
	go w.monitorDevices()

	// Инициализируем основной интерфейс
	if err := w.initializeMainInterface(); err != nil {
		w.logger.Error("Failed to initialize main interface", zap.Error(err))
		return err
	}

	w.logger.Info("WG Profile Manager started successfully")
	return nil
}

// Stop останавливает WG-профиль менеджер
func (w *WGProfileManager) Stop() error {
	w.logger.Info("Stopping WG Profile Manager")

	w.cancel()
	w.wg.Wait()

	if err := w.client.Close(); err != nil {
		w.logger.Error("Failed to close wgctrl client", zap.Error(err))
		return err
	}

	w.logger.Info("WG Profile Manager stopped")
	return nil
}

// initializeMainInterface инициализирует основной WireGuard интерфейс
func (w *WGProfileManager) initializeMainInterface() error {
	interfaceName := w.config.WireGuard.Interface.Name

	// Проверяем, существует ли уже интерфейс
	device, err := w.client.Device(interfaceName)
	if err == nil {
		w.logger.Info("WireGuard interface already exists",
			zap.String("interface", interfaceName),
			zap.String("public_key", device.PublicKey.String()),
		)
		w.mu.Lock()
		w.devices[interfaceName] = device
		w.mu.Unlock()
		return nil
	}

	// Создаем новый интерфейс
	w.logger.Info("Creating new WireGuard interface",
		zap.String("interface", interfaceName),
		zap.Int("port", w.config.WireGuard.Interface.ListenPort),
	)

	// Генерируем ключи
	privateKey, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		return fmt.Errorf("failed to generate private key: %w", err)
	}

	publicKey := privateKey.PublicKey()

	// Конфигурируем устройство
	config := wgtypes.Config{
		PrivateKey:   &privateKey,
		ListenPort:   &w.config.WireGuard.Interface.ListenPort,
		ReplacePeers: true,
	}

	err = w.client.ConfigureDevice(interfaceName, config)
	if err != nil {
		return fmt.Errorf("failed to configure device: %w", err)
	}

	// Получаем информацию о созданном устройстве
	device, err = w.client.Device(interfaceName)
	if err != nil {
		return fmt.Errorf("failed to get device info: %w", err)
	}

	w.mu.Lock()
	w.devices[interfaceName] = device
	w.mu.Unlock()

	w.logger.Info("WireGuard interface created successfully",
		zap.String("interface", interfaceName),
		zap.String("public_key", publicKey.String()),
		zap.Int("port", w.config.WireGuard.Interface.ListenPort),
	)

	return nil
}

// AddPeer добавляет нового пира к интерфейсу
func (w *WGProfileManager) AddPeer(interfaceName string, peerConfig wgtypes.PeerConfig) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	device, exists := w.devices[interfaceName]
	if !exists {
		return fmt.Errorf("interface %s not found", interfaceName)
	}

	// Получаем текущую конфигурацию
	var existingPeers []wgtypes.PeerConfig
	for _, peer := range device.Peers {
		existingPeers = append(existingPeers, wgtypes.PeerConfig{
			PublicKey:  peer.PublicKey,
			Endpoint:   peer.Endpoint,
			AllowedIPs: peer.AllowedIPs,
		})
	}

	config := wgtypes.Config{
		PrivateKey:   &device.PrivateKey,
		ListenPort:   &device.ListenPort,
		ReplacePeers: false,
		Peers:        append(existingPeers, peerConfig),
	}

	err := w.client.ConfigureDevice(interfaceName, config)
	if err != nil {
		return fmt.Errorf("failed to add peer: %w", err)
	}

	// Обновляем локальную информацию
	device, err = w.client.Device(interfaceName)
	if err != nil {
		return fmt.Errorf("failed to get updated device info: %w", err)
	}

	w.devices[interfaceName] = device
	w.peers[peerConfig.PublicKey.String()] = peerConfig

	w.logger.Info("Peer added successfully",
		zap.String("interface", interfaceName),
		zap.String("public_key", peerConfig.PublicKey.String()),
		zap.String("endpoint", peerConfig.Endpoint.String()),
	)

	return nil
}

// RemovePeer удаляет пира из интерфейса
func (w *WGProfileManager) RemovePeer(interfaceName string, publicKey wgtypes.Key) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	device, exists := w.devices[interfaceName]
	if !exists {
		return fmt.Errorf("interface %s not found", interfaceName)
	}

	// Создаем новый список пиров без удаляемого
	var newPeers []wgtypes.PeerConfig
	for _, peer := range device.Peers {
		if peer.PublicKey != publicKey {
			newPeers = append(newPeers, wgtypes.PeerConfig{
				PublicKey:  peer.PublicKey,
				Endpoint:   peer.Endpoint,
				AllowedIPs: peer.AllowedIPs,
			})
		}
	}

	// Конфигурируем устройство с обновленным списком пиров
	config := wgtypes.Config{
		PrivateKey:   &device.PrivateKey,
		ListenPort:   &device.ListenPort,
		ReplacePeers: true,
		Peers:        newPeers,
	}

	err := w.client.ConfigureDevice(interfaceName, config)
	if err != nil {
		return fmt.Errorf("failed to remove peer: %w", err)
	}

	// Обновляем локальную информацию
	device, err = w.client.Device(interfaceName)
	if err != nil {
		return fmt.Errorf("failed to get updated device info: %w", err)
	}

	w.devices[interfaceName] = device
	delete(w.peers, publicKey.String())

	w.logger.Info("Peer removed successfully",
		zap.String("interface", interfaceName),
		zap.String("public_key", publicKey.String()),
	)

	return nil
}

// GetDeviceStatus возвращает статус устройства
func (w *WGProfileManager) GetDeviceStatus(interfaceName string) (*wgtypes.Device, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	device, exists := w.devices[interfaceName]
	if !exists {
		return nil, fmt.Errorf("interface %s not found", interfaceName)
	}

	return device, nil
}

// GetPublicKey возвращает публичный ключ интерфейса
func (w *WGProfileManager) GetPublicKey(interfaceName string) (wgtypes.Key, error) {
	device, err := w.GetDeviceStatus(interfaceName)
	if err != nil {
		return wgtypes.Key{}, err
	}

	return device.PublicKey, nil
}

// monitorDevices мониторит состояние WireGuard устройств
func (w *WGProfileManager) monitorDevices() {
	defer w.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-ticker.C:
			w.updateDeviceStatus()
		}
	}
}

// updateDeviceStatus обновляет статус всех устройств
func (w *WGProfileManager) updateDeviceStatus() {
	w.mu.Lock()
	defer w.mu.Unlock()

	for interfaceName := range w.devices {
		device, err := w.client.Device(interfaceName)
		if err != nil {
			w.logger.Warn("Failed to get device status",
				zap.String("interface", interfaceName),
				zap.Error(err),
			)
			continue
		}

		w.devices[interfaceName] = device

		// Логируем статистику пиров
		for _, peer := range device.Peers {
			w.logger.Debug("Peer status",
				zap.String("interface", interfaceName),
				zap.String("public_key", peer.PublicKey.String()),
				zap.String("endpoint", peer.Endpoint.String()),
				zap.Int64("rx_bytes", peer.ReceiveBytes),
				zap.Int64("tx_bytes", peer.TransmitBytes),
				zap.Time("last_handshake", peer.LastHandshakeTime),
			)
		}
	}
}

// CreatePeerConfig создает конфигурацию пира
func (w *WGProfileManager) CreatePeerConfig(publicKeyStr, endpoint, allowedIPs string) (wgtypes.PeerConfig, error) {
	// Парсим публичный ключ
	publicKey, err := wgtypes.ParseKey(publicKeyStr)
	if err != nil {
		return wgtypes.PeerConfig{}, fmt.Errorf("invalid public key: %w", err)
	}

	// Парсим endpoint
	udpAddr, err := net.ResolveUDPAddr("udp", endpoint)
	if err != nil {
		return wgtypes.PeerConfig{}, fmt.Errorf("invalid endpoint: %w", err)
	}

	// Парсим allowed IPs
	_, allowedIPNet, err := net.ParseCIDR(allowedIPs)
	if err != nil {
		return wgtypes.PeerConfig{}, fmt.Errorf("invalid allowed IPs: %w", err)
	}

	return wgtypes.PeerConfig{
		PublicKey:  publicKey,
		Endpoint:   udpAddr,
		AllowedIPs: []net.IPNet{*allowedIPNet},
	}, nil
}

// GetStats возвращает статистику по всем интерфейсам
func (w *WGProfileManager) GetStats() map[string]interface{} {
	w.mu.RLock()
	defer w.mu.RUnlock()

	stats := make(map[string]interface{})

	for interfaceName, device := range w.devices {
		interfaceStats := map[string]interface{}{
			"name":        device.Name,
			"public_key":  device.PublicKey.String(),
			"listen_port": device.ListenPort,
			"peers_count": len(device.Peers),
			"peers":       make([]map[string]interface{}, 0),
		}

		for _, peer := range device.Peers {
			peerStats := map[string]interface{}{
				"public_key":     peer.PublicKey.String(),
				"endpoint":       peer.Endpoint.String(),
				"allowed_ips":    peer.AllowedIPs,
				"rx_bytes":       peer.ReceiveBytes,
				"tx_bytes":       peer.TransmitBytes,
				"last_handshake": peer.LastHandshakeTime,
			}
			interfaceStats["peers"] = append(interfaceStats["peers"].([]map[string]interface{}), peerStats)
		}

		stats[interfaceName] = interfaceStats
	}

	return stats
}

// CreateSiteToSiteConnection создает site-to-site VPN соединение
func (w *WGProfileManager) CreateSiteToSiteConnection(name string, localSite, remoteSite *Site) (*SiteToSiteConnection, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	connID := fmt.Sprintf("s2s_%d", time.Now().UnixNano())
	connection := &SiteToSiteConnection{
		ID:         connID,
		Name:       name,
		LocalSite:  localSite,
		RemoteSite: remoteSite,
		Status:     "connecting",
		CreatedAt:  time.Now(),
		LastSeen:   time.Now(),
		IsActive:   false,
	}

	w.siteToSiteConnections[connID] = connection

	w.logger.Info("Site-to-site connection created",
		zap.String("conn_id", connID),
		zap.String("name", name),
		zap.String("local_site", localSite.Name),
		zap.String("remote_site", remoteSite.Name))

	return connection, nil
}

// ActivateSiteToSiteConnection активирует site-to-site соединение
func (w *WGProfileManager) ActivateSiteToSiteConnection(connID string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	connection, exists := w.siteToSiteConnections[connID]
	if !exists {
		return fmt.Errorf("site-to-site connection not found: %s", connID)
	}

	// Создаем peer конфигурацию для remote site
	peerConfig, err := w.CreatePeerConfig(
		connection.RemoteSite.PublicKey,
		connection.RemoteSite.Endpoint,
		connection.RemoteSite.AllowedIPs[0], // Используем первый allowed IP
	)
	if err != nil {
		return fmt.Errorf("failed to create peer config: %w", err)
	}

	// Добавляем peer к основному интерфейсу
	interfaceName := w.config.WireGuard.Interface.Name
	if err := w.AddPeer(interfaceName, peerConfig); err != nil {
		return fmt.Errorf("failed to add peer: %w", err)
	}

	// Создаем правила маршрутизации
	for _, allowedIP := range connection.RemoteSite.AllowedIPs {
		ruleID := fmt.Sprintf("rule_%s_%d", connID, time.Now().UnixNano())
		rule := &RoutingRule{
			ID:          ruleID,
			Source:      connection.LocalSite.AllowedIPs[0],
			Destination: allowedIP,
			Gateway:     connection.RemoteSite.Endpoint,
			Interface:   interfaceName,
			Priority:    100,
			IsActive:    true,
		}
		w.routingRules[ruleID] = rule
	}

	connection.Status = "active"
	connection.IsActive = true
	connection.LastSeen = time.Now()

	w.logger.Info("Site-to-site connection activated",
		zap.String("conn_id", connID),
		zap.String("name", connection.Name))

	return nil
}

// DeactivateSiteToSiteConnection деактивирует site-to-site соединение
func (w *WGProfileManager) DeactivateSiteToSiteConnection(connID string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	connection, exists := w.siteToSiteConnections[connID]
	if !exists {
		return fmt.Errorf("site-to-site connection not found: %s", connID)
	}

	// Удаляем peer
	publicKey, err := wgtypes.ParseKey(connection.RemoteSite.PublicKey)
	if err != nil {
		return fmt.Errorf("failed to parse public key: %w", err)
	}

	interfaceName := w.config.WireGuard.Interface.Name
	if err := w.RemovePeer(interfaceName, publicKey); err != nil {
		w.logger.Warn("Failed to remove peer", zap.Error(err))
	}

	// Удаляем правила маршрутизации
	for ruleID, rule := range w.routingRules {
		if rule.Source == connection.LocalSite.AllowedIPs[0] {
			delete(w.routingRules, ruleID)
		}
	}

	connection.Status = "inactive"
	connection.IsActive = false

	w.logger.Info("Site-to-site connection deactivated",
		zap.String("conn_id", connID),
		zap.String("name", connection.Name))

	return nil
}

// GetSiteToSiteConnections возвращает список site-to-site соединений
func (w *WGProfileManager) GetSiteToSiteConnections() []*SiteToSiteConnection {
	w.mu.RLock()
	defer w.mu.RUnlock()

	connections := make([]*SiteToSiteConnection, 0, len(w.siteToSiteConnections))
	for _, conn := range w.siteToSiteConnections {
		connections = append(connections, conn)
	}

	return connections
}

// GetRoutingRules возвращает список правил маршрутизации
func (w *WGProfileManager) GetRoutingRules() []*RoutingRule {
	w.mu.RLock()
	defer w.mu.RUnlock()

	rules := make([]*RoutingRule, 0, len(w.routingRules))
	for _, rule := range w.routingRules {
		rules = append(rules, rule)
	}

	return rules
}
