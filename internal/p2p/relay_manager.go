package p2p

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/quic-go/quic-go"
	"go.uber.org/zap"

	"github.com/twogc/cloudbridge-relay-installer/internal/database"
	"github.com/twogc/cloudbridge-relay-installer/internal/wireguard"
)

// IPPool представляет пул свободных IP адресов для подсети
type IPPool struct {
	subnetCIDR string
	freeIPs    []string
	allocated  map[string]bool
	mutex      sync.Mutex
}

// NewIPPool создает новый пул IP адресов для подсети
func NewIPPool(subnetCIDR string) *IPPool {
	_, ipNet, err := net.ParseCIDR(subnetCIDR)
	if err != nil {
		return nil
	}

	baseIP := ipNet.IP.To4()
	if baseIP == nil {
		return nil
	}

	// Инициализируем пул с IP адресами 2-253 для /24 подсети
	var freeIPs []string
	for i := 2; i <= 253; i++ {
		ip := net.IPv4(baseIP[0], baseIP[1], baseIP[2], byte(i))
		freeIPs = append(freeIPs, ip.String())
	}

	return &IPPool{
		subnetCIDR: subnetCIDR,
		freeIPs:    freeIPs,
		allocated:  make(map[string]bool),
	}
}

// AllocateIP выделяет IP адрес из пула
func (p *IPPool) AllocateIP() (string, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if len(p.freeIPs) == 0 {
		return "", fmt.Errorf("no free IPs available in pool for subnet %s", p.subnetCIDR)
	}

	// Берем первый доступный IP
	ip := p.freeIPs[0]
	p.freeIPs = p.freeIPs[1:]
	p.allocated[ip] = true

	return ip, nil
}

// DeallocateIP возвращает IP адрес в пул
func (p *IPPool) DeallocateIP(ip string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.allocated[ip] {
		return fmt.Errorf("IP %s is not allocated in pool", ip)
	}

	delete(p.allocated, ip)
	p.freeIPs = append(p.freeIPs, ip)

	return nil
}

// IsAllocated проверяет, выделен ли IP адрес
func (p *IPPool) IsAllocated(ip string) bool {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.allocated[ip]
}

// GetStats возвращает статистику пула
func (p *IPPool) GetStats() (total, allocated, free int) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	total = len(p.freeIPs) + len(p.allocated)
	allocated = len(p.allocated)
	free = len(p.freeIPs)

	return
}

// RelayManager управляет P2P relay соединениями
type RelayManager struct {
	tenants          map[string]*Tenant
	connections      map[string]*RelayP2PConnection
	quicConnections  map[string]*QUICConnection
	subnetCounter    int32                             // Счетчик для динамического выделения подсетей
	allocatedNets    map[string]string                 // tenantID -> network CIDR
	repository       *database.Repository              // Database repository for IPAM persistence
	wireguardManager *wireguard.WireGuardTenantManager // WireGuard management for VIP tenants

	// IPAM состояние - все под одним мьютексом для избежания гонок
	allocatedIPs        map[string]map[string]string // tenantID -> ip -> peerID
	peerIndex           map[string]string            // peerID -> ip
	freePools           map[string]*IPPool           // tenantID -> pool of free IPs
	allocationFailures  int64                        // счетчик ошибок аллокации
	deallocationCount   int64                        // счетчик деаллокаций
	avgAllocationTime   time.Duration                // среднее время аллокации
	avgDeallocationTime time.Duration                // среднее время деаллокации

	mutex  sync.RWMutex // единый мьютекс для всего IPAM состояния
	logger *zap.Logger
	ctx    context.Context
	cancel context.CancelFunc
}

// Tenant представляет изолированного пользователя
type Tenant struct {
	TenantID       string                         `json:"tenant_id"`
	NetworkCIDR    string                         `json:"network_cidr"`
	Peers          map[string]*Peer               `json:"peers"`
	Connections    map[string]*RelayP2PConnection `json:"connections"`
	MaxPeers       int                            `json:"max_peers"`
	MaxConnections int                            `json:"max_connections"`
	CreatedAt      time.Time                      `json:"created_at"`
	LastActivity   time.Time                      `json:"last_activity"`
	Mutex          sync.RWMutex                   `json:"-"`
}

// RelayP2PConnection представляет P2P соединение между пирами для relay
type RelayP2PConnection struct {
	ConnectionID     string       `json:"connection_id"`
	TenantID         string       `json:"tenant_id"`
	SourcePeer       string       `json:"source_peer"`
	TargetPeer       string       `json:"target_peer"`
	SourceIP         string       `json:"source_ip"`
	TargetIP         string       `json:"target_ip"`
	ConnectionType   string       `json:"connection_type"` // "quic", "quic-ice"
	Status           string       `json:"status"`          // "connecting", "established", "disconnected"
	CreatedAt        time.Time    `json:"created_at"`
	LastActivity     time.Time    `json:"last_activity"`
	BytesTransferred int64        `json:"bytes_transferred"`
	QUICConn         *quic.Conn   `json:"-"`
	ICEConn          *net.Conn    `json:"-"`
	Mutex            sync.RWMutex `json:"-"`
}

// QUICConnection представляет QUIC соединение с пиром
type QUICConnection struct {
	PeerID   string
	TenantID string
	Conn     *quic.Conn
	Streams  map[quic.StreamID]quic.Stream
	LastSeen time.Time
	IsActive bool
	Mutex    sync.RWMutex
}

// NewRelayManager создает новый Relay Manager
func NewRelayManager(logger *zap.Logger, repository *database.Repository) *RelayManager {
	ctx, cancel := context.WithCancel(context.Background())

	rm := &RelayManager{
		tenants:         make(map[string]*Tenant),
		connections:     make(map[string]*RelayP2PConnection),
		quicConnections: make(map[string]*QUICConnection),
		repository:      repository,
		logger:          logger,
		ctx:             ctx,
		cancel:          cancel,
	}

	// Инициализируем WireGuard менеджер для VIP tenant'ов
	rm.wireguardManager = wireguard.NewWireGuardTenantManager(repository, logger)

	// Инициализируем IPAM данные из базы данных
	err := rm.initializeIPAMFromDB()
	if err != nil {
		logger.Error("Failed to initialize IPAM from database", zap.Error(err))
	}

	return rm
}

// initializeIPAMFromDB загружает IPAM данные из базы данных при запуске
func (rm *RelayManager) initializeIPAMFromDB() error {
	rm.allocatedNets = make(map[string]string)

	// Если repository не инициализирован, используем in-memory режим
	if rm.repository == nil {
		if rm.logger != nil {
			rm.logger.Warn("Repository not initialized, using in-memory IPAM mode")
		}
		// Инициализируем базовые tenant'ы в памяти
		rm.allocatedNets["tenant-216420165"] = "10.100.77.0/24"
		rm.allocatedNets["tenant-123456789"] = "10.100.78.0/24"
		rm.allocatedNets["tenant-987654321"] = "10.100.79.0/24"
		rm.subnetCounter = 79
		return nil
	}

	// Загружаем все активные сети из базы данных
	networks, err := rm.repository.GetAllTenantNetworks()
	if err != nil {
		return fmt.Errorf("failed to load tenant networks from database: %w", err)
	}

	// Заполняем in-memory карту
	for _, network := range networks {
		rm.allocatedNets[network.TenantID] = network.NetworkCIDR
	}

	// Получаем текущий максимальный счетчик подсетей
	maxCounter, err := rm.repository.GetMaxSubnetCounter()
	if err != nil {
		return fmt.Errorf("failed to get max subnet counter: %w", err)
	}
	rm.subnetCounter = int32(maxCounter)

	rm.logger.Info("IPAM initialized from database",
		zap.Int("tenant_networks_loaded", len(networks)),
		zap.Int32("subnet_counter", rm.subnetCounter),
	)

	return nil
}

// RegisterPeer регистрирует пира в tenant
func (rm *RelayManager) RegisterPeer(tenantID, peerID, publicKey string, allowedIPs []string) (*Peer, error) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	// Создаем tenant если не существует
	if _, exists := rm.tenants[tenantID]; !exists {
		networkCIDR, err := rm.GetTenantNetwork(tenantID)
		if err != nil {
			rm.logger.Error("Failed to get tenant network",
				zap.String("tenant_id", tenantID),
				zap.Error(err))
			return nil, err
		}

		rm.tenants[tenantID] = &Tenant{
			TenantID:       tenantID,
			NetworkCIDR:    networkCIDR,
			Peers:          make(map[string]*Peer),
			Connections:    make(map[string]*RelayP2PConnection),
			MaxPeers:       100,
			MaxConnections: 1000,
			CreatedAt:      time.Now(),
		}
	}

	tenant := rm.tenants[tenantID]

	// Проверяем лимиты
	if len(tenant.Peers) >= tenant.MaxPeers {
		return nil, fmt.Errorf("tenant %s has reached max peers limit: %d", tenantID, tenant.MaxPeers)
	}

	// Выделяем IP адрес
	allocatedIP, err := rm.allocateIPForTenant(tenantID)
	if err != nil {
		rm.logger.Error("Failed to allocate IP for tenant",
			zap.String("tenant_id", tenantID),
			zap.Error(err))
		return nil, err
	}

	// Создаем peer
	peer := &Peer{
		PeerID:       peerID,
		TenantID:     tenantID,
		IsOnline:     true,
		LastSeen:     time.Now(),
		RegisteredAt: time.Now(),
	}

	// Добавляем в tenant
	tenant.Peers[peerID] = peer

	rm.logger.Info("Peer registered",
		zap.String("peer_id", peerID),
		zap.String("tenant_id", tenantID),
		zap.String("allocated_ip", allocatedIP))

	return peer, nil
}

// RegisterQUICConnection регистрирует QUIC соединение с пиром
func (rm *RelayManager) RegisterQUICConnection(peerID, tenantID string, conn *quic.Conn) error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	quicConn := &QUICConnection{
		PeerID:   peerID,
		TenantID: tenantID,
		Conn:     conn,
		Streams:  make(map[quic.StreamID]quic.Stream),
		LastSeen: time.Now(),
		IsActive: true,
	}

	rm.quicConnections[peerID] = quicConn

	rm.logger.Info("QUIC connection registered",
		zap.String("peer_id", peerID),
		zap.String("tenant_id", tenantID))

	return nil
}

// RelayPacket пересылает пакет между пирами
func (rm *RelayManager) RelayPacket(sourcePeerID string, data []byte) error {
	// Безопасно получаем информацию об отправителе
	rm.mutex.RLock()
	sourceConn, exists := rm.quicConnections[sourcePeerID]
	rm.mutex.RUnlock()
	if !exists {
		return fmt.Errorf("source peer connection not found: %s", sourcePeerID)
	}

	// Парсим пакет для определения получателя
	targetPeerID, err := rm.parsePacketTarget(data)
	if err != nil {
		return fmt.Errorf("failed to parse packet target: %v", err)
	}

	// Безопасно получаем информацию о получателе
	rm.mutex.RLock()
	targetConn, exists := rm.quicConnections[targetPeerID]
	sameTenant := exists && (targetConn.TenantID == sourceConn.TenantID)
	rm.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("target peer connection not found: %s", targetPeerID)
	}

	// Проверяем изоляцию tenant
	if !sameTenant {
		return fmt.Errorf("cross-tenant relay not allowed: %s -> %s",
			sourceConn.TenantID, targetConn.TenantID)
	}

	// Гарантируем наличие P2P соединения только через публичный метод
	connectionID := rm.generateConnectionID(sourcePeerID, targetPeerID)
	rm.mutex.RLock()
	_, exists = rm.connections[connectionID]
	rm.mutex.RUnlock()
	if !exists {
		if _, err := rm.EstablishP2PConnection(sourcePeerID, targetPeerID); err != nil {
			return fmt.Errorf("failed to establish P2P connection: %v", err)
		}
	}

	// Пересылаем пакет
	err = rm.forwardPacket(targetConn, data)
	if err != nil {
		return fmt.Errorf("failed to forward packet: %v", err)
	}

	// Обновляем статистику (под lock менеджера, затем локальный lock соединения)
	rm.mutex.RLock()
	connection := rm.connections[connectionID]
	rm.mutex.RUnlock()
	if connection != nil {
		connection.Mutex.Lock()
		connection.BytesTransferred += int64(len(data))
		connection.LastActivity = time.Now()
		connection.Mutex.Unlock()
	}

	rm.logger.Debug("Packet relayed",
		zap.String("source_peer", sourcePeerID),
		zap.String("target_peer", targetPeerID),
		zap.Int("bytes", len(data)))

	return nil
}

// parsePacketTarget парсит пакет для определения получателя
func (rm *RelayManager) parsePacketTarget(data []byte) (string, error) {
	// Простой протокол P2P сообщений:
	// Формат: "TO:<peer_id>:<payload>"
	if len(data) < 4 {
		return "", fmt.Errorf("packet too short")
	}

	// Ищем префикс "TO:"
	if string(data[:3]) != "TO:" {
		return "", fmt.Errorf("invalid packet format")
	}

	// Ищем разделитель ':'
	colonIndex := -1
	for i := 3; i < len(data) && i < 100; i++ { // Ограничиваем поиск первыми 100 байтами
		if data[i] == ':' {
			colonIndex = i
			break
		}
	}

	if colonIndex == -1 {
		return "", fmt.Errorf("peer ID separator not found")
	}

	// Извлекаем peer ID
	peerID := string(data[3:colonIndex])

	// Безопасная проверка существования целевого пира
	rm.mutex.RLock()
	_, exists := rm.quicConnections[peerID]
	rm.mutex.RUnlock()

	if exists {
		return peerID, nil
	}

	return "", fmt.Errorf("target peer not found: %s", peerID)
}

// EstablishP2PConnection создает P2P соединение (публичный метод)
func (rm *RelayManager) EstablishP2PConnection(sourcePeerID, targetPeerID string) (*RelayP2PConnection, error) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	return rm.establishP2PConnection(sourcePeerID, targetPeerID)
}

// establishP2PConnection создает P2P соединение (приватный метод)
func (rm *RelayManager) establishP2PConnection(sourcePeerID, targetPeerID string) (*RelayP2PConnection, error) {
	// Получаем информацию о пирах
	sourceConn := rm.quicConnections[sourcePeerID]
	targetConn := rm.quicConnections[targetPeerID]

	// защита от гонок при повторном создании
	id := rm.generateConnectionID(sourcePeerID, targetPeerID)
	if existing, ok := rm.connections[id]; ok {
		return existing, nil
	}

	if sourceConn == nil || targetConn == nil {
		return nil, fmt.Errorf("peer connections not found")
	}

	// Проверяем изоляцию тенантов
	if sourceConn.TenantID != targetConn.TenantID {
		return nil, fmt.Errorf("cross-tenant relay not allowed: %s -> %s",
			sourceConn.TenantID, targetConn.TenantID)
	}

	// Получаем tenant
	tenant := rm.tenants[sourceConn.TenantID]
	if tenant == nil {
		return nil, fmt.Errorf("tenant not found: %s", sourceConn.TenantID)
	}

	// Проверяем лимиты соединений
	if len(tenant.Connections) >= tenant.MaxConnections {
		return nil, fmt.Errorf("tenant %s has reached max connections limit: %d",
			tenant.TenantID, tenant.MaxConnections)
	}

	// Создаем соединение
	connection := &RelayP2PConnection{
		ConnectionID:   id,
		TenantID:       sourceConn.TenantID,
		SourcePeer:     sourcePeerID,
		TargetPeer:     targetPeerID,
		SourceIP:       rm.getPeerIPSafe(sourcePeerID),
		TargetIP:       rm.getPeerIPSafe(targetPeerID),
		ConnectionType: "quic",
		Status:         "established",
		CreatedAt:      time.Now(),
		LastActivity:   time.Now(),
		QUICConn:       targetConn.Conn,
	}

	// Добавляем в tenant
	tenant.Connections[connection.ConnectionID] = connection
	// и в глобальную карту соединений
	rm.connections[connection.ConnectionID] = connection

	rm.logger.Info("P2P connection established",
		zap.String("connection_id", connection.ConnectionID),
		zap.String("source_peer", sourcePeerID),
		zap.String("target_peer", targetPeerID),
		zap.String("tenant_id", sourceConn.TenantID))

	return connection, nil
}

// forwardPacket пересылает пакет получателю
func (rm *RelayManager) forwardPacket(targetConn *QUICConnection, data []byte) error {
	targetConn.Mutex.Lock()
	defer targetConn.Mutex.Unlock()

	if !targetConn.IsActive || targetConn.Conn == nil {
		return fmt.Errorf("target connection is not active")
	}

	// Создаем новый поток для пересылки данных
	stream, err := targetConn.Conn.OpenStream()
	if err != nil {
		return fmt.Errorf("failed to open stream: %v", err)
	}
	defer stream.Close()

	// Отправляем данные
	_, err = stream.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write to stream: %v", err)
	}

	targetConn.LastSeen = time.Now()

	return nil
}

// getTenantNetwork возвращает сеть для tenant
func (rm *RelayManager) GetTenantNetwork(tenantID string) (string, error) {
	// Проверяем в allocatedNets (единый источник истины)
	rm.mutex.RLock()
	if network, exists := rm.allocatedNets[tenantID]; exists {
		rm.mutex.RUnlock()
		return network, nil
	}
	rm.mutex.RUnlock()

	// Генерируем новую подсеть для нового tenant
	return rm.generateNewTenantNetwork(tenantID)
}

// generateNewTenantNetwork генерирует новую сеть для tenant
func (rm *RelayManager) generateNewTenantNetwork(tenantID string) (string, error) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	// Проверяем, есть ли уже выделенная подсеть для этого tenant'а
	if existingNet, exists := rm.allocatedNets[tenantID]; exists {
		return existingNet, nil
	}

	// Инициализируем allocatedNets если нужно
	if rm.allocatedNets == nil {
		rm.allocatedNets = make(map[string]string)
		// Резервируем известные подсети (мигрируем в новый диапазон)
		rm.allocatedNets["tenant-216420165"] = "10.0.77.0/24"
		rm.allocatedNets["tenant-123456789"] = "10.0.78.0/24"
		rm.allocatedNets["tenant-987654321"] = "10.0.79.0/24"
		rm.subnetCounter = 79 // Начинаем с 80
	}

	// Ищем следующую доступную подсеть
	for {
		rm.subnetCounter++
		// Новый алгоритм: используем весь 10.0.0.0/8 диапазон
		// Формат: 10.x.y.0/24 где x = subnetCounter/256, y = subnetCounter%256
		x := rm.subnetCounter / 256
		y := rm.subnetCounter % 256
		candidateNet := fmt.Sprintf("10.%d.%d.0/24", x, y)

		// Проверяем что эта подсеть не занята (в памяти и в БД)
		occupied := false

		// Проверяем в памяти
		for _, allocatedNet := range rm.allocatedNets {
			if allocatedNet == candidateNet {
				occupied = true
				break
			}
		}

		// Если не занята в памяти, проверяем в БД
		if !occupied && rm.repository != nil {
			exists, err := rm.repository.IsTenantNetworkAllocated(candidateNet)
			if err != nil {
				if rm.logger != nil {
					rm.logger.Error("Failed to check network allocation in database",
						zap.String("candidate_net", candidateNet),
						zap.Error(err))
				}
				occupied = true // Безопасно считаем занятой при ошибке
			} else if exists {
				occupied = true
			}
		}

		if !occupied && rm.subnetCounter < 65535 { // 256*256-1 = 65535 (максимум для 10.0.0.0/8)
			// Выделяем подсеть в памяти
			rm.allocatedNets[tenantID] = candidateNet

			// Сохраняем в базе данных если repository доступен
			if rm.repository != nil {
				tenantNetwork := &database.TenantNetwork{
					TenantID:      tenantID,
					NetworkCIDR:   candidateNet,
					SubnetCounter: int(rm.subnetCounter),
					AllocatedAt:   time.Now(),
					IsActive:      true,
				}

				err := rm.repository.CreateTenantNetwork(tenantNetwork)
				if err != nil {
					rm.logger.Error("Failed to save tenant network to database",
						zap.String("tenant_id", tenantID),
						zap.String("subnet", candidateNet),
						zap.Error(err))
					// Продолжаем работу даже если сохранение в БД не удалось
				}
			} else {
				rm.logger.Warn("Repository not available, tenant network not persisted to database",
					zap.String("tenant_id", tenantID),
					zap.String("subnet", candidateNet))
			}

			rm.logger.Info("Allocated new subnet for tenant",
				zap.String("tenant_id", tenantID),
				zap.String("subnet", candidateNet),
				zap.Int32("subnet_counter", rm.subnetCounter))
			return candidateNet, nil
		}

		// Защита от бесконечного цикла
		if rm.subnetCounter >= 65535 {
			rm.logger.Error("Subnet pool exhausted - no more /24 subnets available",
				zap.String("tenant_id", tenantID),
				zap.Int32("current_counter", rm.subnetCounter),
				zap.Int32("max_subnets", 65535))

			// Отправляем алерт
			rm.sendPoolExhaustionAlert(tenantID)

			// Возвращаем ошибку вместо фиксированной подсети
			return "", fmt.Errorf("subnet pool exhausted: no more /24 subnets available (current: %d, max: 65535)", rm.subnetCounter)
		}
	}
}

// allocateIPForTenant выделяет IP адрес для tenant
func (rm *RelayManager) allocateIPForTenant(tenantID string) (string, error) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	// Получаем подсеть тенанта
	tenantSubnet, exists := rm.allocatedNets[tenantID]
	if !exists {
		return "", fmt.Errorf("tenant %s has no allocated subnet", tenantID)
	}

	// Инициализируем пул для тенанта если его нет
	if rm.freePools == nil {
		rm.freePools = make(map[string]*IPPool)
	}
	if rm.allocatedIPs == nil {
		rm.allocatedIPs = make(map[string]map[string]string)
	}
	if rm.peerIndex == nil {
		rm.peerIndex = make(map[string]string)
	}

	pool, exists := rm.freePools[tenantID]
	if !exists {
		pool = NewIPPool(tenantSubnet)
		if pool == nil {
			return "", fmt.Errorf("failed to create IP pool for subnet %s", tenantSubnet)
		}
		rm.freePools[tenantID] = pool
	}

	// Выделяем IP из пула
	ip, err := pool.AllocateIP()
	if err != nil {
		rm.allocationFailures++
		rm.logger.Error("Failed to allocate IP from pool",
			zap.String("tenant_id", tenantID),
			zap.String("subnet", tenantSubnet),
			zap.Error(err))
		return "", fmt.Errorf("no available IPs in tenant subnet %s: %w", tenantSubnet, err)
	}

	// Резервируем IP (без peerID пока)
	if rm.allocatedIPs[tenantID] == nil {
		rm.allocatedIPs[tenantID] = make(map[string]string)
	}
	rm.allocatedIPs[tenantID][ip] = "reserved"

	rm.logger.Info("IP allocated for tenant",
		zap.String("tenant_id", tenantID),
		zap.String("ip", ip),
		zap.String("subnet", tenantSubnet))

	return ip, nil
}

// getPeerIP возвращает IP адрес пира
func (rm *RelayManager) getPeerIP(peerID string) (string, error) {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	ip, exists := rm.peerIndex[peerID]
	if !exists {
		return "", fmt.Errorf("peer %s not found", peerID)
	}

	return ip, nil
}

// getPeerIPSafe безопасно возвращает IP адрес пира (без ошибки для обратной совместимости)
func (rm *RelayManager) getPeerIPSafe(peerID string) string {
	ip, err := rm.getPeerIP(peerID)
	if err != nil {
		rm.logger.Warn("Failed to get peer IP, using fallback",
			zap.String("peer_id", peerID),
			zap.Error(err))
		return "0.0.0.0" // fallback IP
	}
	return ip
}

// reserveIPForPeer резервирует IP адрес для конкретного пира
func (rm *RelayManager) reserveIPForPeer(tenantID, peerID, ip string) error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	// Проверяем, что IP не занят другим пиром
	if existingPeerID, exists := rm.allocatedIPs[tenantID][ip]; exists && existingPeerID != "reserved" {
		return fmt.Errorf("IP %s is already allocated to peer %s", ip, existingPeerID)
	}

	// Резервируем IP для пира
	if rm.allocatedIPs[tenantID] == nil {
		rm.allocatedIPs[tenantID] = make(map[string]string)
	}
	rm.allocatedIPs[tenantID][ip] = peerID
	rm.peerIndex[peerID] = ip

	rm.logger.Info("IP reserved for peer",
		zap.String("tenant_id", tenantID),
		zap.String("peer_id", peerID),
		zap.String("ip", ip))

	return nil
}

// releaseIPForPeer освобождает IP адрес пира
func (rm *RelayManager) releaseIPForPeer(peerID string) error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	ip, exists := rm.peerIndex[peerID]
	if !exists {
		return fmt.Errorf("peer %s not found in index", peerID)
	}

	// Находим тенанта по IP
	var tenantID string
	for tid, ips := range rm.allocatedIPs {
		if allocatedPeerID, exists := ips[ip]; exists && allocatedPeerID == peerID {
			tenantID = tid
			break
		}
	}

	if tenantID == "" {
		return fmt.Errorf("tenant not found for peer %s", peerID)
	}

	// Освобождаем IP в пуле
	if pool, exists := rm.freePools[tenantID]; exists {
		if err := pool.DeallocateIP(ip); err != nil {
			rm.logger.Warn("Failed to deallocate IP in pool",
				zap.String("peer_id", peerID),
				zap.String("ip", ip),
				zap.Error(err))
		}
	}

	// Удаляем из индексов
	delete(rm.allocatedIPs[tenantID], ip)
	delete(rm.peerIndex, peerID)
	rm.deallocationCount++

	rm.logger.Info("IP released for peer",
		zap.String("tenant_id", tenantID),
		zap.String("peer_id", peerID),
		zap.String("ip", ip))

	return nil
}

// sendPoolExhaustionAlert отправляет алерт при исчерпании пула подсетей
func (rm *RelayManager) sendPoolExhaustionAlert(tenantID string) {
	alert := map[string]interface{}{
		"type":          "subnet_pool_exhaustion",
		"tenant_id":     tenantID,
		"timestamp":     time.Now(),
		"current_usage": rm.subnetCounter,
		"max_capacity":  65535,
		"utilization":   float64(rm.subnetCounter) / 65535.0 * 100.0,
		"severity":      "critical",
	}

	// Логируем критический алерт
	rm.logger.Error("CRITICAL: Subnet pool exhausted", zap.Any("alert", alert))

	// Здесь можно добавить интеграцию с системами мониторинга:
	// - Prometheus alerts
	// - Slack notifications
	// - Email alerts
	// - PagerDuty integration
}

// generateConnectionID генерирует уникальный ID соединения
func (rm *RelayManager) generateConnectionID(sourcePeer, targetPeer string) string {
	if sourcePeer < targetPeer {
		return fmt.Sprintf("conn-%s-%s", sourcePeer, targetPeer)
	}
	return fmt.Sprintf("conn-%s-%s", targetPeer, sourcePeer)
}

// GetTenantInfo возвращает информацию о tenant
func (rm *RelayManager) GetTenantInfo(tenantID string) (*Tenant, error) {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	tenant, exists := rm.tenants[tenantID]
	if !exists {
		return nil, fmt.Errorf("tenant not found: %s", tenantID)
	}

	return tenant, nil
}

// GetConnectionInfo возвращает информацию о соединении
func (rm *RelayManager) GetConnectionInfo(connectionID string) (*RelayP2PConnection, error) {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	connection, exists := rm.connections[connectionID]
	if !exists {
		return nil, fmt.Errorf("connection not found: %s", connectionID)
	}

	return connection, nil
}

// CleanupOfflinePeers очищает офлайн пиров
func (rm *RelayManager) CleanupOfflinePeers() {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	cutoff := time.Now().Add(-5 * time.Minute) // 5 минут без активности

	for peerID, quicConn := range rm.quicConnections {
		if quicConn.LastSeen.Before(cutoff) {
			quicConn.IsActive = false
			quicConn.Conn.CloseWithError(0, "peer offline")
			delete(rm.quicConnections, peerID)

			rm.logger.Info("Peer marked as offline",
				zap.String("peer_id", peerID),
				zap.String("tenant_id", quicConn.TenantID))
		}
	}
}

// StartCleanupRoutine запускает routine для очистки офлайн пиров
func (rm *RelayManager) StartCleanupRoutine() {
	ticker := time.NewTicker(60 * time.Second)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-rm.ctx.Done():
				return
			case <-ticker.C:
				rm.CleanupOfflinePeers()
			}
		}
	}()
}

// GetPeer возвращает информацию о пире
func (rm *RelayManager) GetPeer(peerID string) *Peer {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	// Ищем пира во всех tenant
	for _, tenant := range rm.tenants {
		if peer, exists := tenant.Peers[peerID]; exists {
			return peer
		}
	}

	return nil
}

// GetConnection возвращает соединение по ID
func (rm *RelayManager) GetConnection(connectionID string) (*RelayP2PConnection, error) {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	connection, exists := rm.connections[connectionID]
	if !exists {
		return nil, fmt.Errorf("connection not found: %s", connectionID)
	}

	return connection, nil
}

// CloseConnection закрывает соединение
func (rm *RelayManager) CloseConnection(connectionID string) error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	connection, exists := rm.connections[connectionID]
	if !exists {
		return fmt.Errorf("connection not found: %s", connectionID)
	}

	// Обновляем статус соединения
	connection.Status = "disconnected"

	// Удаляем из tenant
	if tenant, exists := rm.tenants[connection.TenantID]; exists {
		delete(tenant.Connections, connectionID)
	}

	// Удаляем из глобального списка
	delete(rm.connections, connectionID)

	rm.logger.Info("Connection closed",
		zap.String("connection_id", connectionID),
		zap.String("tenant_id", connection.TenantID))

	return nil
}

// GetActiveConnections возвращает список активных соединений
func (rm *RelayManager) GetActiveConnections() []*RelayP2PConnection {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	var activeConnections []*RelayP2PConnection
	for _, connection := range rm.connections {
		if connection.Status == "established" {
			activeConnections = append(activeConnections, connection)
		}
	}

	return activeConnections
}

// DeactivateTenant деактивирует tenant и освобождает его ресурсы
func (rm *RelayManager) DeactivateTenant(tenantID string) error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	// Получаем информацию о подсети перед удалением
	allocatedNet, wasAllocated := rm.allocatedNets[tenantID]

	// Закрываем QUIC соединения этого тенанта
	for peerID, qc := range rm.quicConnections {
		if qc.TenantID == tenantID {
			if qc.Conn != nil {
				qc.Conn.CloseWithError(0, "tenant deactivated")
			}
			delete(rm.quicConnections, peerID)
			if rm.logger != nil {
				rm.logger.Info("Peer connection closed due to tenant deactivation",
					zap.String("peer_id", peerID),
					zap.String("tenant_id", tenantID))
			}
		}
	}

	// Закрываем/удаляем P2P соединения этого тенанта
	for connID, c := range rm.connections {
		if c.TenantID == tenantID {
			c.Status = "disconnected"
			delete(rm.connections, connID)
		}
	}

	// Удаляем tenant из памяти
	delete(rm.tenants, tenantID)
	delete(rm.allocatedNets, tenantID)

	// Деактивируем в базе данных если repository доступен
	if rm.repository != nil {
		err := rm.repository.DeactivateTenantNetwork(tenantID)
		if err != nil {
			if rm.logger != nil {
				rm.logger.Error("Failed to deactivate tenant network in database",
					zap.String("tenant_id", tenantID),
					zap.Error(err))
			}
			return fmt.Errorf("failed to deactivate tenant network: %w", err)
		}
	} else {
		if rm.logger != nil {
			rm.logger.Warn("Repository not available, tenant network not deactivated in database",
				zap.String("tenant_id", tenantID))
		}
	}

	// Логируем освобождение подсети
	if wasAllocated && rm.logger != nil {
		rm.logger.Info("Tenant deactivated and subnet freed",
			zap.String("tenant_id", tenantID),
			zap.String("freed_subnet", allocatedNet))
	}

	return nil
}

// GetIPAMStats возвращает статистику по IPAM
func (rm *RelayManager) GetIPAMStats() map[string]interface{} {
	// Собираем snapshot под RLock
	rm.mutex.RLock()
	allocatedTenants := len(rm.allocatedNets)
	currentSubnetCounter := rm.subnetCounter
	allocationFailures := rm.allocationFailures
	deallocationCount := rm.deallocationCount
	avgAllocationTime := rm.avgAllocationTime
	avgDeallocationTime := rm.avgDeallocationTime

	// Собираем статистику по IP пулам
	totalAvailableIPs := 0
	totalAllocatedIPs := 0
	tenantStats := make(map[string]interface{})

	for tenantID, pool := range rm.freePools {
		total, allocated, free := pool.GetStats()
		totalAvailableIPs += total
		totalAllocatedIPs += allocated
		tenantStats[tenantID] = map[string]interface{}{
			"total_ips":     total,
			"allocated_ips": allocated,
			"free_ips":      free,
			"utilization":   float64(allocated) / float64(total) * 100.0,
		}
	}
	rm.mutex.RUnlock()

	// Создаем статистику
	stats := make(map[string]interface{})
	stats["allocated_tenants"] = allocatedTenants
	stats["current_subnet_counter"] = currentSubnetCounter
	stats["max_subnets"] = 65536
	stats["available_subnets"] = 65536 - int(currentSubnetCounter)
	stats["pool_utilization_percent"] = float64(currentSubnetCounter) / 65536.0 * 100.0

	// Статистика по IP
	stats["total_available_ips"] = totalAvailableIPs
	stats["total_allocated_ips"] = totalAllocatedIPs
	stats["ip_utilization_percent"] = float64(totalAllocatedIPs) / float64(totalAvailableIPs) * 100.0

	// Метрики производительности
	stats["allocation_failures"] = allocationFailures
	stats["deallocation_count"] = deallocationCount
	stats["avg_allocation_time_ms"] = avgAllocationTime.Milliseconds()
	stats["avg_deallocation_time_ms"] = avgDeallocationTime.Milliseconds()

	// Статистика по тенантам
	stats["tenant_stats"] = tenantStats

	// Статус синхронизации с БД (вне критической секции)
	if rm.repository != nil {
		networks, err := rm.repository.GetAllTenantNetworks()
		if err == nil {
			stats["db_active_networks"] = len(networks)
			stats["db_sync_status"] = "synced"
		} else {
			stats["db_sync_status"] = "error"
			stats["db_error"] = err.Error()
		}
	} else {
		stats["db_sync_status"] = "disabled"
	}

	return stats
}

// RefreshIPAMFromDB принудительно обновляет IPAM данные из базы данных
func (rm *RelayManager) RefreshIPAMFromDB() error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	return rm.initializeIPAMFromDB()
}

// SyncIPAMWithDB синхронизирует состояние IPAM с базой данных
func (rm *RelayManager) SyncIPAMWithDB() error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	if rm.repository == nil {
		return fmt.Errorf("repository not available")
	}

	// Получаем все активные сети из БД
	networks, err := rm.repository.GetAllTenantNetworks()
	if err != nil {
		return fmt.Errorf("failed to get networks from database: %w", err)
	}

	// Обновляем локальное состояние
	rm.allocatedNets = make(map[string]string)
	maxCounter := int32(0)

	for _, network := range networks {
		rm.allocatedNets[network.TenantID] = network.NetworkCIDR
		if int32(network.SubnetCounter) > maxCounter {
			maxCounter = int32(network.SubnetCounter)
		}
	}

	rm.subnetCounter = maxCounter

	if rm.logger != nil {
		rm.logger.Info("IPAM synchronized with database",
			zap.Int("networks_count", len(networks)),
			zap.Int32("max_subnet_counter", maxCounter))
	}

	return nil
}

// Shutdown останавливает Relay Manager
func (rm *RelayManager) Shutdown() {
	rm.cancel()

	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	// Закрываем все QUIC соединения
	for _, quicConn := range rm.quicConnections {
		quicConn.Conn.CloseWithError(0, "server shutdown")
	}

	rm.logger.Info("Relay Manager shutdown completed")
}
