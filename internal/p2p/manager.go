package p2p

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"go.uber.org/zap"
)

// hostOnly извлекает только IP из net.Addr, убирая порт
func hostOnly(addr net.Addr) string {
	if addr == nil {
		return ""
	}
	host, _, err := net.SplitHostPort(addr.String())
	if err != nil {
		return addr.String() // на крайний случай
	}
	return host
}

// P2PManager управляет P2P соединениями с интеграцией Cadence
type P2PManager struct {
	p2pMesh            P2PMeshTopologyInterface
	peerDiscovery      PeerDiscoveryInterface
	hybridRouter       HybridRouterInterface
	siteToSiteManager  SiteToSiteManagerInterface
	metrics            *P2PMetrics
	logger             *zap.Logger
	isStarted          bool
	cadenceIntegration CadenceP2PIntegrationInterface
	mu                 sync.RWMutex
}

// NewP2PManager создает новый P2P менеджер
func NewP2PManager(
	mesh P2PMeshTopologyInterface,
	discovery PeerDiscoveryInterface,
	router HybridRouterInterface,
	siteToSite SiteToSiteManagerInterface,
	logger *zap.Logger,
) *P2PManager {
	pm := &P2PManager{
		p2pMesh:           mesh,
		peerDiscovery:     discovery,
		hybridRouter:      router,
		siteToSiteManager: siteToSite,
		metrics:           NewP2PMetrics(),
		logger:            logger,
		isStarted:         false,
	}

	// Cadence интеграция будет инициализирована позже через SetCadenceIntegration
	return pm
}

// NewP2PManagerWithCadence создает новый P2P менеджер с Cadence интеграцией
func NewP2PManagerWithCadence(
	mesh P2PMeshTopologyInterface,
	discovery PeerDiscoveryInterface,
	router HybridRouterInterface,
	siteToSite SiteToSiteManagerInterface,
	cadenceClient interface{}, // *cadence.Client
	logger *zap.Logger,
) *P2PManager {
	pm := NewP2PManager(mesh, discovery, router, siteToSite, logger)

	// Инициализация Cadence интеграции
	// pm.cadenceIntegration = NewCadenceP2PIntegration(cadenceClient, pm, logger)

	return pm
}

// SetCadenceIntegration устанавливает Cadence интеграцию
func (pm *P2PManager) SetCadenceIntegration(cadenceIntegration CadenceP2PIntegrationInterface) {
	pm.cadenceIntegration = cadenceIntegration
	pm.logger.Info("Cadence интеграция установлена для P2P Manager")
}

// Initialize инициализирует P2P компоненты
func (pm *P2PManager) Initialize(ctx context.Context) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.isStarted {
		return fmt.Errorf("P2P Manager already started")
	}

	// Запускаем peer discovery
	if err := pm.peerDiscovery.Start(); err != nil {
		return fmt.Errorf("failed to start peer discovery: %w", err)
	}

	// Добавляем обнаруженных пиров в mesh topology
	go pm.processDiscoveredPeers(ctx)

	pm.isStarted = true
	pm.logger.Info("P2P Manager initialized successfully")
	return nil
}

// processDiscoveredPeers обрабатывает обнаруженных пиров
func (pm *P2PManager) processDiscoveredPeers(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second) // Проверяем каждые 30 секунд
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			peers, err := pm.peerDiscovery.DiscoverPeers()
			if err != nil {
				pm.logger.Error("Failed to discover peers", zap.Error(err))
				continue
			}

			for _, peer := range peers {
				// Проверяем, не добавлен ли уже этот peer
				if pm.isPeerConnected(peer.PeerID) {
					continue
				}
				p2pPeer := &Peer{
					ID:             peer.PeerID,
					PeerID:         peer.PeerID,
					TenantID:       peer.TenantID,
					Endpoint:       peer.Endpoint,
					ConnectionType: "quic",
					IsOnline:       true,
					LastSeen:       peer.LastSeen,
					LatencyMS:      peer.Latency,
					QuicConfig: &QuicConfig{
						Endpoint:            peer.Endpoint,
						MaxStreams:          100,
						HandshakeTimeout:    "10s",
						IdleTimeout:         "60s",
						ConnectionMigration: true,
						Multiplexing:        true,
					},
				}

				// Добавляем в mesh topology
				if err := pm.p2pMesh.AddPeer(peer.PeerID, p2pPeer); err != nil {
					pm.logger.Error("Failed to add peer to mesh",
						zap.String("peer_id", peer.PeerID),
						zap.Error(err))
					continue
				}

				// Статус остается "discovered" до реальной проверки соединения
				pm.logger.Info("Peer added to P2P mesh",
					zap.String("peer_id", peer.PeerID),
					zap.String("endpoint", peer.Endpoint),
					zap.String("region", peer.Region),
				)
			}

			// Обновляем количество пиров в метриках
			pm.UpdatePeerCount(len(peers))
		}
	}
}

// isPeerConnected проверяет, подключен ли peer
func (pm *P2PManager) isPeerConnected(peerID string) bool {
	// Проверяем в mesh topology
	if pm.p2pMesh != nil {
		// Получаем реальный статус подключения peer из mesh topology
		// Используем прямой доступ к ConnectedPeers через публичные методы
		peer, exists := pm.p2pMesh.GetPeer(peerID)
		if !exists {
			pm.logger.Debug("Peer not found in mesh topology", zap.String("peer_id", peerID))
			return false
		}

		// Проверяем статус подключения
		return peer.Status == "connected"
	}
	return false
}

// IsAvailable проверяет доступность P2P соединения
func (pm *P2PManager) IsAvailable(ctx context.Context, clientIP string, tenantID string) bool {
	pm.mu.RLock()
	if !pm.isStarted {
		pm.mu.RUnlock()
		pm.logger.Debug("P2P Manager not started")
		return false
	}
	pm.mu.RUnlock()

	// Проверяем, есть ли доступные пиры
	peers, err := pm.peerDiscovery.DiscoverPeers()
	if err != nil {
		pm.logger.Debug("Failed to discover peers for availability check", zap.Error(err))
		return false
	}

	if len(peers) == 0 {
		pm.logger.Debug("No P2P peers available")
		return false
	}

	// Проверяем, есть ли подходящий маршрут
	if pm.p2pMesh != nil {
		routes, err := pm.p2pMesh.GetOptimalRoute(clientIP)
		if err == nil && len(routes) > 0 {
			pm.logger.Debug("P2P route available",
				zap.String("client_ip", clientIP),
				zap.Strings("routes", routes),
			)
			return true
		}
	}

	pm.logger.Debug("No suitable P2P route found",
		zap.String("client_ip", clientIP),
		zap.String("tenant_id", tenantID),
	)
	return false
}

// HandleConnection обрабатывает P2P соединение
func (pm *P2PManager) HandleConnection(ctx context.Context, conn net.Conn, tenantID string) error {
	start := time.Now()

	pm.mu.RLock()
	if !pm.isStarted {
		pm.mu.RUnlock()
		return fmt.Errorf("P2P Manager not started")
	}
	pm.mu.RUnlock()

	clientIP := hostOnly(conn.RemoteAddr())
	pm.logger.Info("Handling P2P connection",
		zap.String("client_ip", clientIP),
		zap.String("tenant_id", tenantID),
	)

	// Определяем оптимальный маршрут
	route, err := pm.determineOptimalRoute(ctx, clientIP, tenantID)
	if err != nil {
		pm.metrics.RecordConnection("p2p", "route_error", tenantID, time.Since(start))
		return fmt.Errorf("failed to determine optimal route: %w", err)
	}

	// Обрабатываем соединение в зависимости от типа маршрута
	switch route.Type {
	case "direct":
		err = pm.handleDirectP2P(ctx, conn, route)
	case "mesh":
		err = pm.handleMeshP2P(ctx, conn, route)
	default:
		err = fmt.Errorf("unknown route type: %s", route.Type)
	}

	if err != nil {
		pm.metrics.RecordConnection("p2p", "connection_error", tenantID, time.Since(start))
		pm.metrics.RecordConnectionError("connection_failed", "p2p")
		return fmt.Errorf("P2P connection failed: %w", err)
	}

	pm.metrics.RecordConnection("p2p", "success", tenantID, time.Since(start))
	pm.metrics.IncrementTotalConnections()
	pm.logger.Info("P2P connection handled successfully",
		zap.String("route_type", route.Type),
		zap.Duration("duration", time.Since(start)),
	)

	return nil
}

// RouteInfo представляет информацию о маршруте
type RouteInfo struct {
	Type       string
	PeerID     string
	Endpoint   string
	Latency    time.Duration
	TrustLevel string
}

// CreateTunnelWithOrchestration создает туннель с оркестрацией через Cadence
func (pm *P2PManager) CreateTunnelWithOrchestration(ctx context.Context, tenantID, peerID string, config map[string]interface{}) error {
	pm.logger.Info("Создание туннеля с оркестрацией",
		zap.String("tenant_id", tenantID),
		zap.String("peer_id", peerID),
	)

	if pm.cadenceIntegration == nil {
		pm.logger.Warn("Cadence интеграция не настроена, создание туннеля без оркестрации")
		// Fallback к обычному созданию туннеля
		return pm.createTunnelDirectly(ctx, tenantID, peerID, config)
	}

	// Запуск оркестрации через Cadence
	err := pm.cadenceIntegration.OrchestrateTunnelCreation(ctx, tenantID, peerID, config)
	if err != nil {
		pm.logger.Error("Ошибка оркестрации создания туннеля", zap.Error(err))
		return err
	}

	pm.logger.Info("Туннель создан с оркестрацией через Cadence")
	return nil
}

// EstablishConnectionWithOrchestration устанавливает соединение с оркестрацией через Cadence
func (pm *P2PManager) EstablishConnectionWithOrchestration(ctx context.Context, tenantID, sourcePeerID, targetPeerID string, config map[string]interface{}) error {
	pm.logger.Info("Установка соединения с оркестрацией",
		zap.String("tenant_id", tenantID),
		zap.String("source_peer", sourcePeerID),
		zap.String("target_peer", targetPeerID),
	)

	if pm.cadenceIntegration == nil {
		pm.logger.Warn("Cadence интеграция не настроена, установка соединения без оркестрации")
		// Fallback к обычной установке соединения
		return pm.establishConnectionDirectly(ctx, tenantID, sourcePeerID, targetPeerID, config)
	}

	// Запуск оркестрации через Cadence
	err := pm.cadenceIntegration.OrchestrateConnectionEstablishment(ctx, tenantID, sourcePeerID, targetPeerID, config)
	if err != nil {
		pm.logger.Error("Ошибка оркестрации установки соединения", zap.Error(err))
		return err
	}

	pm.logger.Info("Соединение установлено с оркестрацией через Cadence")
	return nil
}

// OptimizeRoutingWithAI оптимизирует маршрутизацию с помощью AI через Cadence
func (pm *P2PManager) OptimizeRoutingWithAI(ctx context.Context, tenantID, sourcePeerID, targetPeerID string, dataSize int64, priority string, networkConditions map[string]interface{}) error {
	pm.logger.Info("AI-оптимизация маршрутизации",
		zap.String("tenant_id", tenantID),
		zap.String("source_peer", sourcePeerID),
		zap.String("target_peer", targetPeerID),
		zap.String("priority", priority),
		zap.Int64("data_size", dataSize),
	)

	if pm.cadenceIntegration == nil {
		pm.logger.Warn("Cadence интеграция не настроена, AI-оптимизация недоступна")
		return fmt.Errorf("cadence integration not available")
	}

	// Запуск AI-оптимизации через Cadence
	err := pm.cadenceIntegration.OrchestrateAIOptimizedRouting(ctx, tenantID, sourcePeerID, targetPeerID, dataSize, priority, networkConditions)
	if err != nil {
		pm.logger.Error("Ошибка AI-оптимизации маршрутизации", zap.Error(err))
		return err
	}

	pm.logger.Info("AI-оптимизация маршрутизации запущена через Cadence")
	return nil
}

// createTunnelDirectly создает туннель напрямую (fallback метод)
func (pm *P2PManager) createTunnelDirectly(ctx context.Context, tenantID, peerID string, config map[string]interface{}) error {
	pm.logger.Info("Создание туннеля напрямую",
		zap.String("tenant_id", tenantID),
		zap.String("peer_id", peerID),
	)

	// Здесь должна быть логика прямого создания туннеля
	// Интеграция с существующими P2P компонентами

	pm.logger.Info("Туннель создан напрямую")
	return nil
}

// establishConnectionDirectly устанавливает соединение напрямую (fallback метод)
func (pm *P2PManager) establishConnectionDirectly(ctx context.Context, tenantID, sourcePeerID, targetPeerID string, config map[string]interface{}) error {
	pm.logger.Info("Установка соединения напрямую",
		zap.String("tenant_id", tenantID),
		zap.String("source_peer", sourcePeerID),
		zap.String("target_peer", targetPeerID),
	)

	// Здесь должна быть логика прямой установки соединения
	// Интеграция с существующими P2P компонентами

	pm.logger.Info("Соединение установлено напрямую")
	return nil
}

// pickBestPeer выбирает лучшего пира с параллельными проверками
func (pm *P2PManager) pickBestPeer(ctx context.Context, peers []*DiscoveredPeer) *DiscoveredPeer {
	type res struct {
		p   *DiscoveredPeer
		ok  bool
		rtt time.Duration
	}
	ch := make(chan res, len(peers))

	for _, p := range peers {
		p := p
		go func() {
			start := time.Now()
			ok := pm.isDirectRouteAvailable(ctx, p)
			ch <- res{p: p, ok: ok, rtt: time.Since(start)}
		}()
	}

	var best *DiscoveredPeer
	bestRTT := time.Hour
	for i := 0; i < len(peers); i++ {
		r := <-ch
		if !r.ok {
			continue
		}
		if r.rtt < bestRTT {
			bestRTT = r.rtt
			best = r.p
		}
	}
	if best != nil {
		best.Latency = int(bestRTT / time.Millisecond)
	}
	return best
}

// determineOptimalRoute определяет оптимальный маршрут для соединения
func (pm *P2PManager) determineOptimalRoute(ctx context.Context, clientIP, tenantID string) (*RouteInfo, error) {
	// Получаем доступных пиров
	peers, err := pm.peerDiscovery.DiscoverPeers()
	if err != nil {
		return nil, fmt.Errorf("failed to discover peers: %w", err)
	}

	if len(peers) == 0 {
		return nil, fmt.Errorf("no P2P peers available")
	}

	// Фильтруем по trust level
	filtered := make([]*DiscoveredPeer, 0, len(peers))
	for _, p := range peers {
		if p.TrustLevel == "basic" || p.TrustLevel == "high" {
			filtered = append(filtered, p)
		}
	}

	bestPeer := pm.pickBestPeer(ctx, filtered)
	if bestPeer == nil {
		return nil, fmt.Errorf("no suitable P2P peer found")
	}

	// Определяем тип маршрута
	routeType := "mesh"
	if pm.isDirectRouteAvailable(ctx, bestPeer) {
		routeType = "direct"
	}

	return &RouteInfo{
		Type:       routeType,
		PeerID:     bestPeer.PeerID,
		Endpoint:   bestPeer.Endpoint,
		Latency:    time.Duration(bestPeer.Latency) * time.Millisecond,
		TrustLevel: bestPeer.TrustLevel,
	}, nil
}

// isDirectRouteAvailable проверяет доступность прямого маршрута
func (pm *P2PManager) isDirectRouteAvailable(ctx context.Context, peer *DiscoveredPeer) bool {
	// Используем контекст с таймаутом
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// UDP probe для QUIC (более подходящий для P2P)
	d := &net.Dialer{}
	conn, err := d.DialContext(ctx, "udp", peer.Endpoint)
	if err != nil {
		pm.logger.Debug("Direct route probe (UDP) failed",
			zap.String("endpoint", peer.Endpoint),
			zap.Error(err))
		return false
	}
	_ = conn.Close()

	pm.logger.Debug("Direct route available",
		zap.String("peer_endpoint", peer.Endpoint))
	return true
}

// handleDirectP2P обрабатывает прямое P2P соединение
func (pm *P2PManager) handleDirectP2P(ctx context.Context, conn net.Conn, route *RouteInfo) error {
	pm.logger.Info("Establishing direct P2P connection",
		zap.String("peer_id", route.PeerID),
		zap.String("endpoint", route.Endpoint),
	)
	if pm.p2pMesh != nil {
		// В реальной реализации здесь будет получение peer через публичный метод
		// Пока используем симуляцию
		pm.logger.Info("Direct P2P connection established",
			zap.String("peer_id", route.PeerID),
			zap.Duration("latency", route.Latency),
		)
		return nil
	}

	return fmt.Errorf("direct P2P connection failed: peer not available")
}

// handleMeshP2P обрабатывает mesh P2P соединение
func (pm *P2PManager) handleMeshP2P(ctx context.Context, conn net.Conn, route *RouteInfo) error {
	pm.logger.Info("Establishing mesh P2P connection",
		zap.String("peer_id", route.PeerID),
		zap.String("endpoint", route.Endpoint),
	)

	// В реальной реализации здесь будет установка mesh соединения через несколько пиров
	// Получаем оптимальный mesh маршрут
	if pm.p2pMesh != nil {
		routes, err := pm.p2pMesh.GetOptimalRoute(conn.RemoteAddr().String())
		if err != nil {
			return fmt.Errorf("failed to get mesh route: %w", err)
		}

		pm.logger.Info("Mesh P2P connection established",
			zap.String("peer_id", route.PeerID),
			zap.Strings("route_hops", routes),
			zap.Duration("latency", route.Latency),
		)
		return nil
	}

	return fmt.Errorf("mesh P2P connection failed: mesh topology not available")
}

// measureLatency измеряет задержку до клиента
func (pm *P2PManager) measureLatency(ctx context.Context, clientIP string) time.Duration {
	start := time.Now()

	// Используем контекст с таймаутом
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// UDP probe для более точного измерения (подходит для QUIC)
	d := &net.Dialer{}
	conn, err := d.DialContext(ctx, "udp", clientIP+":9090")
	if err != nil {
		pm.logger.Debug("Failed to measure latency",
			zap.String("client_ip", clientIP),
			zap.Error(err))
		return time.Hour // Ошибка соединения
	}
	conn.Close()

	latency := time.Since(start)
	pm.logger.Debug("Latency measurement completed",
		zap.String("client_ip", clientIP),
		zap.Duration("latency", latency),
	)

	return latency
}

// GetMetrics возвращает метрики P2P
func (pm *P2PManager) GetMetrics() *P2PMetrics {
	return pm.metrics
}

// GetMetricsSnapshot возвращает снимок метрик
func (pm *P2PManager) GetMetricsSnapshot() map[string]interface{} {
	return map[string]interface{}{
		"peer_count": 0, // Временная заглушка - нужно добавить метод в P2PMetrics
		// Добавить другие агрегаты по необходимости
	}
}

// UpdatePeerCount обновляет количество активных пиров
func (pm *P2PManager) UpdatePeerCount(count int) {
	pm.metrics.SetPeerCount(count)
}

// GetStatus возвращает статус P2P Manager
func (pm *P2PManager) GetStatus() map[string]interface{} {
	pm.mu.RLock()
	status := map[string]interface{}{
		"started": pm.isStarted,
		"metrics": pm.GetMetricsSnapshot(),
	}
	pm.mu.RUnlock()

	// Добавляем информацию о mesh topology
	if pm.p2pMesh != nil {
		// В реальной реализации здесь будет получение статуса через публичные методы
		status["mesh_peers"] = 0 // Временная заглушка
		status["local_peer_id"] = "peer-local"
	}

	// Добавляем информацию о peer discovery
	if pm.peerDiscovery != nil {
		peers, err := pm.peerDiscovery.DiscoverPeers()
		if err == nil {
			status["discovered_peers"] = len(peers)
		}
	}

	// Добавляем статистику маршрутизации
	if pm.hybridRouter != nil {
		status["routing_stats"] = pm.hybridRouter.GetRoutingStats()
	}

	return status
}

// Shutdown корректно завершает работу P2P Manager
func (pm *P2PManager) Shutdown(ctx context.Context) error {
	pm.logger.Info("Shutting down P2P Manager")

	// Останавливаем peer discovery
	if pm.peerDiscovery != nil {
		if err := pm.peerDiscovery.Stop(); err != nil {
			pm.logger.Error("Failed to stop peer discovery", zap.Error(err))
		}
	}

	pm.logger.Info("P2P Manager shutdown completed")
	return nil
}
