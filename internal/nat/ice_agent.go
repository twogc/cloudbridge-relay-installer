package nat

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
)

// ICEAgent представляет ICE агент для NAT traversal
type ICEAgent struct {
	logger *logger.Logger

	// STUN и TURN серверы
	stunServers     []string
	turnPool        *TURNPool
	turnPoolManager *TURNPoolManager

	// ICE кандидаты
	localCandidates  []*ICECandidate
	remoteCandidates []*ICECandidate
	selectedPair     *ICECandidatePair

	// TURN allocations
	turnAllocations map[string]*TURNAllocation
	allocationsMu   sync.RWMutex

	// Состояние
	state ICEState
	mu    sync.RWMutex

	// Метрики
	Metrics *NATMetrics

	// Контекст
	ctx    context.Context
	cancel context.CancelFunc
}

// ICEState представляет состояние ICE агента
type ICEState int

const (
	ICEStateNew ICEState = iota
	ICEStateGathering
	ICEStateGathered
	ICEStateChecking
	ICEStateConnected
	ICEStateCompleted
	ICEStateFailed
	ICEStateDisconnected
)

// ICECandidate представляет ICE кандидат
type ICECandidate struct {
	Foundation  string
	Component   int
	Transport   string
	Priority    uint32
	Address     string
	Port        int
	Type        ICECandidateType
	RelatedAddr string
	RelatedPort int
}

// ICECandidateType тип ICE кандидата
type ICECandidateType int

const (
	ICECandidateTypeHost ICECandidateType = iota
	ICECandidateTypeServerReflexive
	ICECandidateTypePeerReflexive
	ICECandidateTypeRelay
)

// ICECandidatePair представляет пару ICE кандидатов
type ICECandidatePair struct {
	LocalCandidate  *ICECandidate
	RemoteCandidate *ICECandidate
	Priority        uint64
	State           ICECandidatePairState
}

// ICECandidatePairState состояние пары кандидатов
type ICECandidatePairState int

const (
	ICECandidatePairStateWaiting ICECandidatePairState = iota
	ICECandidatePairStateInProgress
	ICECandidatePairStateSucceeded
	ICECandidatePairStateFailed
	ICECandidatePairStateFrozen
)

// NewICEAgent создает новый ICE агент
func NewICEAgent(logger *logger.Logger, stunServers []string, turnPool *TURNPool) *ICEAgent {
	ctx, cancel := context.WithCancel(context.Background())

	ia := &ICEAgent{
		logger:          logger,
		stunServers:     stunServers,
		turnPool:        turnPool,
		turnAllocations: make(map[string]*TURNAllocation),
		state:           ICEStateNew,
		Metrics:         NewNATMetrics(logger),
		ctx:             ctx,
		cancel:          cancel,
	}

	return ia
}

// NewICEAgentWithTURNPoolManager создает новый ICE агент с TURN pool manager
func NewICEAgentWithTURNPoolManager(logger *logger.Logger, stunServers []string, turnPoolManager *TURNPoolManager) *ICEAgent {
	ctx, cancel := context.WithCancel(context.Background())

	ia := &ICEAgent{
		logger:          logger,
		stunServers:     stunServers,
		turnPoolManager: turnPoolManager,
		turnAllocations: make(map[string]*TURNAllocation),
		state:           ICEStateNew,
		Metrics:         NewNATMetrics(logger),
		ctx:             ctx,
		cancel:          cancel,
	}

	return ia
}

// AllocateTURNRelay создает TURN allocation через pool manager
func (ia *ICEAgent) AllocateTURNRelay(clientID string, preferences *AllocationPreferences) (*TURNAllocation, error) {
	if ia.turnPoolManager == nil {
		return nil, fmt.Errorf("TURN pool manager not configured")
	}

	allocation, err := ia.turnPoolManager.AllocateTURN(clientID, preferences)
	if err != nil {
		ia.logger.Error("Failed to allocate TURN relay",
			zap.String("client_id", clientID),
			zap.Error(err))
		return nil, err
	}

	// Сохраняем allocation в ICE agent
	ia.allocationsMu.Lock()
	ia.turnAllocations[allocation.AllocationID] = allocation
	ia.allocationsMu.Unlock()

	ia.logger.Info("TURN relay allocated",
		zap.String("allocation_id", allocation.AllocationID),
		zap.String("client_id", clientID),
		zap.String("server", allocation.Server.Address))

	return allocation, nil
}

// ReleaseTURNRelay освобождает TURN allocation
func (ia *ICEAgent) ReleaseTURNRelay(allocationID string) error {
	ia.allocationsMu.Lock()
	allocation, exists := ia.turnAllocations[allocationID]
	if exists {
		delete(ia.turnAllocations, allocationID)
	}
	ia.allocationsMu.Unlock()

	if !exists {
		return fmt.Errorf("allocation %s not found in ICE agent", allocationID)
	}

	if ia.turnPoolManager != nil {
		return ia.turnPoolManager.ReleaseTURN(allocationID)
	}

	ia.logger.Info("TURN relay released",
		zap.String("allocation_id", allocationID),
		zap.String("client_id", allocation.ClientID))

	return nil
}

// GetTURNAllocations возвращает список активных TURN allocations
func (ia *ICEAgent) GetTURNAllocations() []*TURNAllocation {
	ia.allocationsMu.RLock()
	defer ia.allocationsMu.RUnlock()

	allocations := make([]*TURNAllocation, 0, len(ia.turnAllocations))
	for _, allocation := range ia.turnAllocations {
		allocations = append(allocations, allocation)
	}

	return allocations
}

// StartGathering запускает сбор ICE кандидатов
func (ia *ICEAgent) StartGathering() error {
	ia.mu.Lock()
	defer ia.mu.Unlock()

	if ia.state != ICEStateNew {
		return fmt.Errorf("ICE agent is not in new state")
	}

	ia.state = ICEStateGathering
	ia.logger.Info("Starting ICE candidate gathering")

	// Запускаем сбор кандидатов в отдельной горутине
	go ia.gatherCandidates()

	return nil
}

// gatherCandidates собирает ICE кандидаты
func (ia *ICEAgent) gatherCandidates() {
	defer func() {
		ia.mu.Lock()
		if ia.state == ICEStateGathering {
			ia.state = ICEStateGathered
		}
		ia.mu.Unlock()
	}()

	// 1. Собираем host candidates
	ia.gatherHostCandidates()

	// 2. Собираем server reflexive candidates через STUN
	ia.gatherServerReflexiveCandidates()

	// 3. Собираем relay candidates через TURN
	ia.gatherRelayCandidates()

	ia.logger.Info("ICE candidate gathering completed",
		zap.Int("local_candidates", len(ia.localCandidates)),
	)
}

// gatherHostCandidates собирает host кандидаты
func (ia *ICEAgent) gatherHostCandidates() {
	interfaces, err := net.Interfaces()
	if err != nil {
		ia.logger.Error("Failed to get network interfaces", zap.Error(err))
		return
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil { // Только IPv4
					candidate := &ICECandidate{
						Foundation: fmt.Sprintf("host-%s", ipnet.IP.String()),
						Component:  1,
						Transport:  "UDP",
						Priority:   ia.calculatePriority(ICECandidateTypeHost, 0),
						Address:    ipnet.IP.String(),
						Port:       0, // Будет установлен при создании соединения
						Type:       ICECandidateTypeHost,
					}
					ia.localCandidates = append(ia.localCandidates, candidate)
				}
			}
		}
	}
}

// gatherServerReflexiveCandidates собирает server reflexive кандидаты через STUN
func (ia *ICEAgent) gatherServerReflexiveCandidates() {
	for _, stunServer := range ia.stunServers {
		candidate := ia.performSTUNBinding(stunServer)
		if candidate != nil {
			ia.localCandidates = append(ia.localCandidates, candidate)
		}
	}
}

// gatherRelayCandidates собирает relay кандидаты через TURN
func (ia *ICEAgent) gatherRelayCandidates() {
	if ia.turnPool == nil {
		return
	}

	server, err := ia.turnPool.GetBestServer()
	if err != nil {
		ia.logger.Warn("No TURN servers available for relay candidates")
		return
	}

	// Создаем TURN allocation для relay кандидата
	allocation, err := ia.turnPool.AllocateTURN("ice-agent")
	if err != nil {
		ia.logger.Error("Failed to allocate TURN", zap.Error(err))
		return
	}

	candidate := &ICECandidate{
		Foundation: fmt.Sprintf("relay-%s", server.Address),
		Component:  1,
		Transport:  "UDP",
		Priority:   ia.calculatePriority(ICECandidateTypeRelay, 0),
		Address:    allocation.RelayAddress,
		Port:       0, // Будет установлен при создании соединения
		Type:       ICECandidateTypeRelay,
	}
	ia.localCandidates = append(ia.localCandidates, candidate)
}

// performSTUNBinding выполняет STUN binding request
func (ia *ICEAgent) performSTUNBinding(stunServer string) *ICECandidate {
	start := time.Now()

	// Create STUN client
	stunClient, err := NewSTUNClient(stunServer, ia.logger.GetZapLogger())
	if err != nil {
		ia.logger.Debug("Failed to create STUN client", zap.String("server", stunServer), zap.Error(err))
		ia.Metrics.RecordSTUNRequest("ice-agent", stunServer, false, time.Since(start))
		return nil
	}
	defer stunClient.Close()

	// Get external address through STUN
	externalAddr, err := stunClient.GetExternalAddress()
	if err != nil {
		ia.logger.Debug("STUN binding failed", zap.String("server", stunServer), zap.Error(err))
		ia.Metrics.RecordSTUNRequest("ice-agent", stunServer, false, time.Since(start))
		return nil
	}

	ia.Metrics.RecordSTUNRequest("ice-agent", stunServer, true, time.Since(start))

	// Get local address for related address
	localAddr := "192.168.1.100" // Default fallback
	localPort := 12345

	// Try to get actual local address
	if stunClient.conn != nil {
		if addr := stunClient.conn.LocalAddr(); addr != nil {
			if udpAddr, ok := addr.(*net.UDPAddr); ok {
				localAddr = udpAddr.IP.String()
				localPort = udpAddr.Port
			}
		}
	}

	return &ICECandidate{
		Foundation:  fmt.Sprintf("srflx-%s", stunServer),
		Component:   1,
		Transport:   "UDP",
		Priority:    ia.calculatePriority(ICECandidateTypeServerReflexive, 0),
		Address:     externalAddr.IP.String(),
		Port:        externalAddr.Port,
		Type:        ICECandidateTypeServerReflexive,
		RelatedAddr: localAddr,
		RelatedPort: localPort,
	}
}

// calculatePriority вычисляет приоритет кандидата
func (ia *ICEAgent) calculatePriority(candidateType ICECandidateType, localPref int) uint32 {
	// RFC 5245: priority = (2^24) * (type preference) + (2^8) * (local preference) + (2^0) * (component ID)
	typePreference := uint32(126) // Host
	switch candidateType {
	case ICECandidateTypeHost:
		typePreference = 126
	case ICECandidateTypeServerReflexive:
		typePreference = 100
	case ICECandidateTypePeerReflexive:
		typePreference = 110
	case ICECandidateTypeRelay:
		typePreference = 0
	}

	return (1<<24)*typePreference + (1<<8)*uint32(localPref) + 1
}

// AddRemoteCandidate добавляет удаленный кандидат
func (ia *ICEAgent) AddRemoteCandidate(candidate *ICECandidate) error {
	ia.mu.Lock()
	defer ia.mu.Unlock()

	ia.remoteCandidates = append(ia.remoteCandidates, candidate)
	ia.logger.Debug("Added remote ICE candidate",
		zap.String("address", candidate.Address),
		zap.Int("port", candidate.Port),
		zap.String("type", candidate.Type.String()),
	)

	return nil
}

// StartConnectivityChecks запускает connectivity checks
func (ia *ICEAgent) StartConnectivityChecks() error {
	ia.mu.Lock()
	defer ia.mu.Unlock()

	if ia.state != ICEStateGathered {
		return fmt.Errorf("ICE agent is not in gathered state")
	}

	ia.state = ICEStateChecking
	ia.logger.Info("Starting ICE connectivity checks")

	// Запускаем connectivity checks в отдельной горутине
	go ia.performConnectivityChecks()

	return nil
}

// performConnectivityChecks выполняет connectivity checks
func (ia *ICEAgent) performConnectivityChecks() {
	defer func() {
		ia.mu.Lock()
		if ia.state == ICEStateChecking {
			ia.state = ICEStateConnected
		}
		ia.mu.Unlock()
	}()

	// Создаем пары кандидатов
	pairs := ia.createCandidatePairs()

	// Сортируем пары по приоритету
	ia.sortCandidatePairs(pairs)

	// Выполняем connectivity checks
	for _, pair := range pairs {
		if ia.performConnectivityCheck(pair) {
			ia.selectedPair = pair
			ia.logger.Info("ICE connectivity check succeeded",
				zap.String("local", pair.LocalCandidate.Address),
				zap.String("remote", pair.RemoteCandidate.Address),
			)
			return
		}
	}

	ia.logger.Error("All ICE connectivity checks failed")
	ia.mu.Lock()
	ia.state = ICEStateFailed
	ia.mu.Unlock()
}

// createCandidatePairs создает пары кандидатов
func (ia *ICEAgent) createCandidatePairs() []*ICECandidatePair {
	var pairs []*ICECandidatePair

	for _, local := range ia.localCandidates {
		for _, remote := range ia.remoteCandidates {
			pair := &ICECandidatePair{
				LocalCandidate:  local,
				RemoteCandidate: remote,
				Priority:        ia.calculatePairPriority(local, remote),
				State:           ICECandidatePairStateWaiting,
			}
			pairs = append(pairs, pair)
		}
	}

	return pairs
}

// calculatePairPriority вычисляет приоритет пары кандидатов
func (ia *ICEAgent) calculatePairPriority(local, remote *ICECandidate) uint64 {
	// RFC 5245: pair priority = 2^32 * min(G,D) + 2 * max(G,D) + (G>D?1:0)
	g := uint64(local.Priority)
	d := uint64(remote.Priority)

	if g < d {
		return (1<<32)*g + 2*d + 0
	}
	return (1<<32)*d + 2*g + 1
}

// sortCandidatePairs сортирует пары кандидатов по приоритету
func (ia *ICEAgent) sortCandidatePairs(pairs []*ICECandidatePair) {
	// Простая сортировка пузырьком по убыванию приоритета
	for i := 0; i < len(pairs)-1; i++ {
		for j := i + 1; j < len(pairs); j++ {
			if pairs[i].Priority < pairs[j].Priority {
				pairs[i], pairs[j] = pairs[j], pairs[i]
			}
		}
	}
}

// performConnectivityCheck выполняет connectivity check для пары кандидатов
func (ia *ICEAgent) performConnectivityCheck(pair *ICECandidatePair) bool {
	pair.State = ICECandidatePairStateInProgress

	// В реальной реализации здесь должна быть логика connectivity check
	// Пока просто проверяем доступность удаленного адреса
	conn, err := net.DialTimeout("udp", fmt.Sprintf("%s:%d", pair.RemoteCandidate.Address, pair.RemoteCandidate.Port), 5*time.Second)
	if err != nil {
		pair.State = ICECandidatePairStateFailed
		ia.Metrics.RecordConnectivityCheck("ice-agent", false)
		return false
	}
	defer conn.Close()

	pair.State = ICECandidatePairStateSucceeded
	ia.Metrics.RecordConnectivityCheck("ice-agent", true)
	return true
}

// GetState возвращает текущее состояние ICE агента
func (ia *ICEAgent) GetState() ICEState {
	ia.mu.RLock()
	defer ia.mu.RUnlock()
	return ia.state
}

// GetLocalCandidates возвращает локальные кандидаты
func (ia *ICEAgent) GetLocalCandidates() []*ICECandidate {
	ia.mu.RLock()
	defer ia.mu.RUnlock()
	return ia.localCandidates
}

// GetSelectedPair возвращает выбранную пару кандидатов
func (ia *ICEAgent) GetSelectedPair() *ICECandidatePair {
	ia.mu.RLock()
	defer ia.mu.RUnlock()
	return ia.selectedPair
}

// Close закрывает ICE агент
func (ia *ICEAgent) Close() error {
	ia.cancel()
	ia.Metrics.Stop()
	ia.logger.Info("ICE agent closed")
	return nil
}

// String возвращает строковое представление ICE состояния
func (s ICEState) String() string {
	switch s {
	case ICEStateNew:
		return "new"
	case ICEStateGathering:
		return "gathering"
	case ICEStateGathered:
		return "gathered"
	case ICEStateChecking:
		return "checking"
	case ICEStateConnected:
		return "connected"
	case ICEStateCompleted:
		return "completed"
	case ICEStateFailed:
		return "failed"
	case ICEStateDisconnected:
		return "disconnected"
	default:
		return "unknown"
	}
}

// String возвращает строковое представление типа кандидата
func (t ICECandidateType) String() string {
	switch t {
	case ICECandidateTypeHost:
		return "host"
	case ICECandidateTypeServerReflexive:
		return "srflx"
	case ICECandidateTypePeerReflexive:
		return "prflx"
	case ICECandidateTypeRelay:
		return "relay"
	default:
		return "unknown"
	}
}

// Stop останавливает ICE agent и освобождает все ресурсы
func (ia *ICEAgent) Stop() error {
	ia.mu.Lock()
	defer ia.mu.Unlock()

	// Освобождаем все TURN allocations
	ia.allocationsMu.Lock()
	for allocationID := range ia.turnAllocations {
		if ia.turnPoolManager != nil {
			if err := ia.turnPoolManager.ReleaseTURN(allocationID); err != nil {
				ia.logger.Error("Failed to release TURN allocation during stop",
					zap.String("allocation_id", allocationID),
					zap.Error(err))
			}
		}
	}
	ia.turnAllocations = make(map[string]*TURNAllocation)
	ia.allocationsMu.Unlock()

	// Останавливаем контекст
	ia.cancel()

	// Обновляем состояние
	ia.state = ICEStateFailed

	ia.logger.Info("ICE agent stopped")
	return nil
}
