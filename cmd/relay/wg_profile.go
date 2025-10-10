package main

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/config"
	"github.com/twogc/cloudbridge-relay-installer/internal/metrics"
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

	// Метрики
	metrics *metrics.Metrics
}

// NewWGProfileManager создает новый менеджер WG-профилей
func NewWGProfileManager(cfg *config.Config, logger *zap.Logger, metrics *metrics.Metrics) (*WGProfileManager, error) {
	client, err := wgctrl.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create wgctrl client: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &WGProfileManager{
		client:                client,
		config:                cfg,
		logger:                logger,
		metrics:               metrics,
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

	// Записываем метрики запуска интерфейса
	if w.metrics != nil {
		w.metrics.IncWireGuardInterfaces()
		w.metrics.IncWireGuardInterfacesCreated(w.config.WireGuard.Interface.Name, "main")
		w.metrics.SetWireGuardInterfaceStatus(w.config.WireGuard.Interface.Name, "starting", true)
	}

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

	if w.metrics != nil {
		w.metrics.SetWireGuardInterfaceStatus(w.config.WireGuard.Interface.Name, "stopped", false)
	}

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

	// Создаем интерфейс с помощью ip команды
	if err := w.createInterface(interfaceName); err != nil {
		return fmt.Errorf("failed to create interface: %w", err)
	}

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

// createInterface создает WireGuard интерфейс с помощью ip команды
func (w *WGProfileManager) createInterface(interfaceName string) error {
	if _, err := w.client.Device(interfaceName); err == nil {
		w.logger.Info("WireGuard interface already exists", zap.String("interface", interfaceName))
		return nil
	}

	// ip link add <if> type wireguard
	if out, err := exec.Command("ip", "link", "add", interfaceName, "type", "wireguard").CombinedOutput(); err != nil {
		if strings.Contains(string(out), "File exists") {
			w.logger.Info("WireGuard interface already exists", zap.String("interface", interfaceName))
		} else {
			return fmt.Errorf("ip link add failed: %w, output: %s", err, string(out))
		}
	}

	// ip address add <cidr> dev <if>
	addr := w.config.WireGuard.Interface.Address // например "10.20.30.1/24"
	if addr != "" {
		if out, err := exec.Command("ip", "address", "add", addr, "dev", interfaceName).CombinedOutput(); err != nil {
			// адрес мог уже быть назначен — не фейлимся жёстко
			if !strings.Contains(string(out), "File exists") {
				return fmt.Errorf("ip address add failed: %w, output: %s", err, string(out))
			}
		}
	}

	// ip link set <if> up
	if out, err := exec.Command("ip", "link", "set", interfaceName, "up").CombinedOutput(); err != nil {
		return fmt.Errorf("ip link set up failed: %w, output: %s", err, string(out))
	}

	w.logger.Info("WireGuard interface created and brought up", zap.String("interface", interfaceName))
	return nil
}

// AddPeer добавляет нового пира к интерфейсу
func (w *WGProfileManager) AddPeer(interfaceName string, peerConfig wgtypes.PeerConfig) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// убедимся, что интерфейс существует
	if _, exists := w.devices[interfaceName]; !exists {
		return fmt.Errorf("interface %s not found", interfaceName)
	}

	// передаём только то, что меняем
	cfg := wgtypes.Config{
		ReplacePeers: false, // не затирать остальных
		Peers:        []wgtypes.PeerConfig{peerConfig},
	}
	if err := w.client.ConfigureDevice(interfaceName, cfg); err != nil {
		return fmt.Errorf("failed to add peer: %w", err)
	}

	// перечитываем актуальное состояние
	device, err := w.client.Device(interfaceName)
	if err != nil {
		return fmt.Errorf("failed to get updated device info: %w", err)
	}
	w.devices[interfaceName] = device
	w.peers[peerConfig.PublicKey.String()] = peerConfig

	w.logger.Info("Peer added successfully",
		zap.String("interface", interfaceName),
		zap.String("public_key", peerConfig.PublicKey.String()),
		func() zap.Field {
			if peerConfig.Endpoint != nil {
				return zap.String("endpoint", peerConfig.Endpoint.String())
			}
			return zap.String("endpoint", "")
		}(),
	)

	// Записываем метрики добавления пира
	if w.metrics != nil {
		w.metrics.IncWireGuardPeers()
		w.metrics.IncWireGuardPeersAdded(interfaceName, peerConfig.PublicKey.String())
		w.metrics.SetWireGuardPeersConnected(interfaceName, peerConfig.PublicKey.String(), 1)
	}

	return nil
}

// RemovePeer удаляет пира из интерфейса
func (w *WGProfileManager) RemovePeer(interfaceName string, publicKey wgtypes.Key) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if _, exists := w.devices[interfaceName]; !exists {
		return fmt.Errorf("interface %s not found", interfaceName)
	}

	cfg := wgtypes.Config{
		ReplacePeers: false,
		Peers: []wgtypes.PeerConfig{{
			PublicKey: publicKey,
			Remove:    true,
		}},
	}
	if err := w.client.ConfigureDevice(interfaceName, cfg); err != nil {
		return fmt.Errorf("failed to remove peer: %w", err)
	}

	device, err := w.client.Device(interfaceName)
	if err != nil {
		return fmt.Errorf("failed to get updated device info: %w", err)
	}
	w.devices[interfaceName] = device
	delete(w.peers, publicKey.String())

	w.logger.Info("Peer removed successfully",
		zap.String("interface", interfaceName),
		zap.String("public_key", publicKey.String()),
	)

	// Записываем метрики удаления пира
	if w.metrics != nil {
		w.metrics.DecWireGuardPeers()
		w.metrics.IncWireGuardPeersRemoved(interfaceName, publicKey.String(), "manual")
		w.metrics.SetWireGuardPeersConnected(interfaceName, publicKey.String(), 0)
	}

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
			endpoint := ""
			if peer.Endpoint != nil {
				endpoint = peer.Endpoint.String()
			}
			w.logger.Debug("Peer status",
				zap.String("interface", interfaceName),
				zap.String("public_key", peer.PublicKey.String()),
				zap.String("endpoint", endpoint),
				zap.Int64("rx_bytes", peer.ReceiveBytes),
				zap.Int64("tx_bytes", peer.TransmitBytes),
				zap.Time("last_handshake", peer.LastHandshakeTime),
			)
		}
	}
}

// CreatePeerConfig создает конфигурацию пира
func (w *WGProfileManager) CreatePeerConfig(publicKeyStr, endpoint, allowedIPs string) (wgtypes.PeerConfig, error) {
	pub, err := wgtypes.ParseKey(publicKeyStr)
	if err != nil {
		return wgtypes.PeerConfig{}, fmt.Errorf("invalid public key: %w", err)
	}

	var ep *net.UDPAddr
	if strings.TrimSpace(endpoint) != "" {
		udp, err := net.ResolveUDPAddr("udp", endpoint)
		if err != nil {
			return wgtypes.PeerConfig{}, fmt.Errorf("invalid endpoint: %w", err)
		}
		ep = udp
	}

	var ipnets []net.IPNet
	// поддержим "10.0.0.0/24,10.0.1.0/24" и пробелы
	for _, part := range strings.FieldsFunc(allowedIPs, func(r rune) bool { return r == ',' || r == ' ' || r == ';' }) {
		if part = strings.TrimSpace(part); part == "" {
			continue
		}
		_, n, err := net.ParseCIDR(part)
		if err != nil {
			return wgtypes.PeerConfig{}, fmt.Errorf("invalid allowed IPs %q: %w", part, err)
		}
		ipnets = append(ipnets, *n)
	}
	if len(ipnets) == 0 {
		return wgtypes.PeerConfig{}, fmt.Errorf("allowed IPs cannot be empty")
	}

	return wgtypes.PeerConfig{
		PublicKey:  pub,
		Endpoint:   ep,
		AllowedIPs: ipnets,
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

	if len(connection.RemoteSite.AllowedIPs) == 0 || len(connection.LocalSite.AllowedIPs) == 0 {
		return fmt.Errorf("site-to-site: allowed IPs must be non-empty (local=%d, remote=%d)",
			len(connection.LocalSite.AllowedIPs), len(connection.RemoteSite.AllowedIPs))
	}

	peerConfig, err := w.CreatePeerConfig(
		connection.RemoteSite.PublicKey,
		connection.RemoteSite.Endpoint,
		strings.Join(connection.RemoteSite.AllowedIPs, ","),
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
		w.routingRules[ruleID] = &RoutingRule{
			ID:          ruleID,
			Source:      connection.LocalSite.AllowedIPs[0],
			Destination: allowedIP,
			Gateway:     connection.RemoteSite.Endpoint,
			Interface:   interfaceName,
			Priority:    100,
			IsActive:    true,
		}
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

// RecordTrafficMetrics записывает метрики трафика
func (w *WGProfileManager) RecordTrafficMetrics(interfaceName, peer string, bytesReceived, bytesSent, packetsReceived, packetsSent int64) {
	if w.metrics == nil {
		return
	}

	w.metrics.IncWireGuardTrafficBytesReceived(interfaceName, peer, bytesReceived)
	w.metrics.IncWireGuardTrafficBytesSent(interfaceName, peer, bytesSent)
	w.metrics.IncWireGuardTrafficPacketsReceived(interfaceName, peer, packetsReceived)
	w.metrics.IncWireGuardTrafficPacketsSent(interfaceName, peer, packetsSent)
}

// RecordPerformanceMetrics записывает метрики производительности
func (w *WGProfileManager) RecordPerformanceMetrics(interfaceName, peer string, metrics map[string]interface{}) {
	if w.metrics == nil {
		return
	}

	// Записываем метрики производительности
	if handshakeDuration, ok := metrics["handshake_duration"].(float64); ok {
		w.metrics.ObserveWireGuardHandshakeDuration(interfaceName, peer, handshakeDuration)
	}

	if connectionLatency, ok := metrics["connection_latency"].(float64); ok {
		w.metrics.ObserveWireGuardConnectionLatency(interfaceName, peer, connectionLatency)
	}

	if throughputBytes, ok := metrics["throughput_bytes"].(int64); ok {
		direction := "unknown"
		if dir, exists := metrics["direction"].(string); exists {
			direction = dir
		}
		w.metrics.IncWireGuardThroughputBytes(interfaceName, peer, direction, throughputBytes)
	}

	if throughputPackets, ok := metrics["throughput_packets"].(int64); ok {
		direction := "unknown"
		if dir, exists := metrics["direction"].(string); exists {
			direction = dir
		}
		w.metrics.IncWireGuardThroughputPackets(interfaceName, peer, direction, throughputPackets)
	}

	if bandwidthUtilization, ok := metrics["bandwidth_utilization"].(float64); ok {
		w.metrics.SetWireGuardBandwidthUtilization(interfaceName, peer, bandwidthUtilization)
	}

	if trafficRate, ok := metrics["traffic_rate"].(float64); ok {
		direction := "unknown"
		if dir, exists := metrics["direction"].(string); exists {
			direction = dir
		}
		w.metrics.SetWireGuardTrafficRate(interfaceName, peer, direction, trafficRate)
	}
}

// RecordConnectionQualityMetrics записывает метрики качества соединения
func (w *WGProfileManager) RecordConnectionQualityMetrics(interfaceName, peer string, metrics map[string]interface{}) {
	if w.metrics == nil {
		return
	}

	if keepaliveInterval, ok := metrics["keepalive_interval"].(float64); ok {
		w.metrics.SetWireGuardKeepaliveInterval(interfaceName, peer, keepaliveInterval)
	}

	if lastHandshake, ok := metrics["last_handshake"].(float64); ok {
		w.metrics.SetWireGuardLastHandshake(interfaceName, peer, lastHandshake)
	}

	if handshakeAttempts, ok := metrics["handshake_attempts"].(int); ok {
		for i := 0; i < handshakeAttempts; i++ {
			w.metrics.IncWireGuardHandshakeAttempts(interfaceName, peer)
		}
	}

	if connectionUptime, ok := metrics["connection_uptime"].(float64); ok {
		w.metrics.SetWireGuardConnectionUptime(interfaceName, peer, connectionUptime)
	}
}

// RecordErrorMetrics записывает метрики ошибок
func (w *WGProfileManager) RecordErrorMetrics(interfaceName, peer, errorType string) {
	if w.metrics == nil {
		return
	}

	switch errorType {
	case "interface_error":
		w.metrics.IncWireGuardInterfaceErrors(interfaceName, "interface_failed")
	case "peer_error":
		w.metrics.IncWireGuardPeerErrors(interfaceName, peer, "peer_failed")
	case "handshake_error":
		w.metrics.IncWireGuardHandshakeErrors(interfaceName, peer, "handshake_failed")
	case "timeout":
		w.metrics.IncWireGuardTimeoutEvents(interfaceName, peer, "connection_timeout")
	case "reconnection":
		w.metrics.IncWireGuardReconnectionAttempts(interfaceName, peer)
	case "authentication_failure":
		w.metrics.IncWireGuardAuthenticationFailures(interfaceName, peer, "auth_failed")
	case "encryption_failure":
		w.metrics.IncWireGuardEncryptionFailures(interfaceName, peer, "encryption_failed")
	case "unauthorized_access":
		w.metrics.IncWireGuardUnauthorizedAccess(interfaceName, peer, "access_denied")
	}
}

// RecordSiteToSiteMetrics записывает метрики site-to-site соединений
func (w *WGProfileManager) RecordSiteToSiteMetrics(connectionID string, metrics map[string]interface{}) {
	if w.metrics == nil {
		return
	}

	if status, ok := metrics["status"].(string); ok {
		count := 1
		if c, exists := metrics["count"].(int); exists {
			count = c
		}
		w.metrics.SetWireGuardSiteToSiteConnections(connectionID, status, count)
	}

	if active, ok := metrics["active"].(int); ok {
		w.metrics.SetWireGuardSiteToSiteActive(connectionID, active)
	}

	if bandwidth, ok := metrics["bandwidth"].(float64); ok {
		w.metrics.SetWireGuardSiteToSiteBandwidth(connectionID, bandwidth)
	}

	if latency, ok := metrics["latency"].(float64); ok {
		w.metrics.ObserveWireGuardSiteToSiteLatency(connectionID, latency)
	}
}

// RecordResourceMetrics записывает метрики ресурсов
func (w *WGProfileManager) RecordResourceMetrics(interfaceName, component string, metrics map[string]interface{}) {
	if w.metrics == nil {
		return
	}

	if memoryUsage, ok := metrics["memory_usage"].(int64); ok {
		w.metrics.SetWireGuardMemoryUsage(interfaceName, component, memoryUsage)
	}

	if cpuUsage, ok := metrics["cpu_usage"].(float64); ok {
		w.metrics.SetWireGuardCPUUsage(interfaceName, component, cpuUsage)
	}

	if goroutineCount, ok := metrics["goroutine_count"].(int); ok {
		w.metrics.SetWireGuardGoroutineCount(interfaceName, component, goroutineCount)
	}

	if fileDescriptorCount, ok := metrics["file_descriptor_count"].(int); ok {
		w.metrics.SetWireGuardFileDescriptorCount(interfaceName, fileDescriptorCount)
	}
}
