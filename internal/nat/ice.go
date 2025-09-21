package nat

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ICEAgent представляет ICE агент для NAT traversal
type ICEAgent struct {
	stunServers      []string
	turnServers      []*TURNServer
	turnRelayServer  *TURNRelayServer
	localCandidates  []*Candidate
	remoteCandidates []*Candidate
	selectedPair     *CandidatePair
	state            ICEState
	mu               sync.RWMutex
	logger           *zap.Logger
	ctx              context.Context
	cancel           context.CancelFunc
	metrics          *NATMetrics
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

// Candidate представляет ICE candidate
type Candidate struct {
	Foundation  string
	Component   int
	Transport   string
	Priority    uint32
	Address     string
	Port        int
	Type        CandidateType
	RelatedAddr string
	RelatedPort int
	Username    string
	Password    string
}

// CandidateType тип ICE candidate
type CandidateType int

const (
	CandidateTypeHost CandidateType = iota
	CandidateTypeServerReflexive
	CandidateTypePeerReflexive
	CandidateTypeRelay
)

// CandidatePair представляет пару candidates
type CandidatePair struct {
	Local  *Candidate
	Remote *Candidate
	State  CandidatePairState
}

// CandidatePairState состояние пары candidates
type CandidatePairState int

const (
	CandidatePairStateWaiting CandidatePairState = iota
	CandidatePairStateInProgress
	CandidatePairStateSucceeded
	CandidatePairStateFailed
	CandidatePairStateFrozen
)

// TURNServer конфигурация TURN сервера
type TURNServer struct {
	Address  string
	Username string
	Password string
	Realm    string
}

// NewICEAgent создает новый ICE агент
func NewICEAgent(stunServers []string, turnServers []*TURNServer, logger *zap.Logger) *ICEAgent {
	ctx, cancel := context.WithCancel(context.Background())

	// Создаем TURN relay сервер
	// TODO: Создать адаптер для zap.Logger -> logger.Logger
	// turnRelayServer := NewTURNRelayServer(logger, nil)
	var turnRelayServer *TURNRelayServer

	return &ICEAgent{
		stunServers:      stunServers,
		turnServers:      turnServers,
		turnRelayServer:  turnRelayServer,
		localCandidates:  make([]*Candidate, 0),
		remoteCandidates: make([]*Candidate, 0),
		state:            ICEStateNew,
		logger:           logger,
		ctx:              ctx,
		cancel:           cancel,
		metrics:          NewNATMetrics(logger),
	}
}

// StartGathering начинает сбор ICE candidates
func (ia *ICEAgent) StartGathering() error {
	ia.mu.Lock()
	ia.state = ICEStateGathering
	ia.mu.Unlock()

	ia.logger.Info("Starting ICE candidate gathering")

	// 1. Собираем host candidates
	if err := ia.gatherHostCandidates(); err != nil {
		return fmt.Errorf("failed to gather host candidates: %w", err)
	}

	// 2. Собираем server reflexive candidates через STUN
	if err := ia.gatherServerReflexiveCandidates(); err != nil {
		ia.logger.Warn("Failed to gather server reflexive candidates", zap.Error(err))
	}

	// 3. Собираем relay candidates через TURN
	if err := ia.gatherRelayCandidates(); err != nil {
		ia.logger.Warn("Failed to gather relay candidates", zap.Error(err))
	}

	ia.mu.Lock()
	ia.state = ICEStateGathered
	ia.mu.Unlock()

	ia.logger.Info("ICE candidate gathering completed",
		zap.Int("candidates_count", len(ia.localCandidates)))

	return nil
}

// StartTURNServer запускает TURN relay сервер
func (ia *ICEAgent) StartTURNServer() error {
	if ia.turnRelayServer == nil {
		return fmt.Errorf("TURN relay server not initialized")
	}

	ia.logger.Info("Starting TURN relay server")

	if err := ia.turnRelayServer.Start(); err != nil {
		return fmt.Errorf("failed to start TURN relay server: %w", err)
	}

	ia.logger.Info("TURN relay server started successfully")
	return nil
}

// StopTURNServer останавливает TURN relay сервер
func (ia *ICEAgent) StopTURNServer() error {
	if ia.turnRelayServer == nil {
		return nil
	}

	ia.logger.Info("Stopping TURN relay server")

	if err := ia.turnRelayServer.Close(); err != nil {
		return fmt.Errorf("failed to stop TURN relay server: %w", err)
	}

	ia.logger.Info("TURN relay server stopped")
	return nil
}

// gatherHostCandidates собирает host candidates
func (ia *ICEAgent) gatherHostCandidates() error {
	interfaces, err := net.Interfaces()
	if err != nil {
		return fmt.Errorf("failed to get network interfaces: %w", err)
	}

	for _, iface := range interfaces {
		// Пропускаем неактивные интерфейсы
		if iface.Flags&net.FlagUp == 0 {
			continue
		}

		// Пропускаем loopback интерфейсы (кроме случаев, когда они явно разрешены)
		if iface.Flags&net.FlagLoopback != 0 {
			// Можно добавить конфигурацию для включения loopback
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			ia.logger.Warn("Failed to get addresses for interface",
				zap.String("interface", iface.Name),
				zap.Error(err))
			continue
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}

			ip := ipNet.IP

			// Пропускаем невалидные IP адреса
			if ip == nil || ip.IsLoopback() || ip.IsMulticast() || ip.IsUnspecified() {
				continue
			}

			// Пропускаем IPv6 link-local адреса для избежания location tracking
			if ip.To4() == nil && (ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast()) {
				ia.logger.Debug("Skipping IPv6 link-local address for location tracking",
					zap.String("address", ip.String()))
				continue
			}

			// Создаем UDP соединение для host candidate
			conn, err := net.ListenUDP("udp", &net.UDPAddr{
				IP:   ip,
				Port: 0, // Система выберет свободный порт
			})
			if err != nil {
				ia.logger.Warn("Failed to create UDP connection for host candidate",
					zap.String("address", ip.String()),
					zap.Error(err))
				continue
			}

			// Получаем реальный адрес соединения
			localAddr := conn.LocalAddr().(*net.UDPAddr)

			// Создаем host candidate
			candidate := &Candidate{
				Foundation:  fmt.Sprintf("host-%s-%d", ip.String(), localAddr.Port),
				Component:   1,
				Transport:   "UDP",
				Priority:    ia.calculatePriority(CandidateTypeHost, 1),
				Address:     localAddr.IP.String(),
				Port:        localAddr.Port,
				Type:        CandidateTypeHost,
				RelatedAddr: "",
				RelatedPort: 0,
			}

			ia.mu.Lock()
			ia.localCandidates = append(ia.localCandidates, candidate)
			ia.mu.Unlock()

			ia.logger.Debug("Added host candidate",
				zap.String("address", candidate.Address),
				zap.Int("port", candidate.Port),
				zap.Uint32("priority", candidate.Priority),
				zap.String("interface", iface.Name))

			// Записываем метрику кандидата
			ia.metrics.RecordICECandidate("host", "UDP")

			// Закрываем соединение после создания candidate
			// В реальной реализации соединение должно храниться для использования
			conn.Close()
		}
	}

	ia.logger.Info("Host candidates gathering completed",
		zap.Int("candidates_count", len(ia.localCandidates)))

	return nil
}

// gatherServerReflexiveCandidates собирает server reflexive candidates через STUN
func (ia *ICEAgent) gatherServerReflexiveCandidates() error {
	for _, stunServer := range ia.stunServers {
		stunClient := NewSTUNClient(stunServer, ia.logger)

		// Для каждого host candidate получаем server reflexive candidate
		for _, hostCandidate := range ia.localCandidates {
			if hostCandidate.Type == CandidateTypeHost {
				srCandidate, err := stunClient.GetServerReflexiveCandidate(
					hostCandidate.Address, hostCandidate.Port)
				if err != nil {
					ia.logger.Warn("Failed to get server reflexive candidate",
						zap.String("stun_server", stunServer),
						zap.String("host_address", hostCandidate.Address),
						zap.Error(err))
					continue
				}

				srCandidate.RelatedAddr = hostCandidate.Address
				srCandidate.RelatedPort = hostCandidate.Port
				ia.localCandidates = append(ia.localCandidates, srCandidate)

				// Записываем метрику server reflexive кандидата
				ia.metrics.RecordICECandidate("srflx", "UDP")
				ia.metrics.RecordSTUNRequest(stunServer, "success")
			}
		}
	}

	return nil
}

// gatherRelayCandidates собирает relay candidates через TURN
func (ia *ICEAgent) gatherRelayCandidates() error {
	for _, turnServer := range ia.turnServers {
		turnClient := NewTURNClient(turnServer, ia.logger)

		relayCandidate, err := turnClient.AllocateRelayCandidate()
		if err != nil {
			ia.logger.Warn("Failed to allocate relay candidate",
				zap.String("turn_server", turnServer.Address),
				zap.Error(err))
			continue
		}

		ia.localCandidates = append(ia.localCandidates, relayCandidate)
	}

	return nil
}

// AddRemoteCandidate добавляет remote candidate
func (ia *ICEAgent) AddRemoteCandidate(candidate *Candidate) error {
	ia.mu.Lock()
	defer ia.mu.Unlock()

	ia.remoteCandidates = append(ia.remoteCandidates, candidate)

	ia.logger.Info("Added remote candidate",
		zap.String("address", candidate.Address),
		zap.Int("port", candidate.Port),
		zap.String("type", candidate.Type.String()))

	return nil
}

// StartConnectivityChecks начинает connectivity checks
func (ia *ICEAgent) StartConnectivityChecks() error {
	ia.mu.Lock()
	ia.state = ICEStateChecking
	ia.mu.Unlock()

	ia.logger.Info("Starting ICE connectivity checks")

	// Сортируем candidates по приоритету
	ia.sortCandidatesByPriority()

	// Начинаем connectivity checks
	go ia.performConnectivityChecks()

	return nil
}

// performConnectivityChecks выполняет connectivity checks
func (ia *ICEAgent) performConnectivityChecks() {
	for _, localCandidate := range ia.localCandidates {
		for _, remoteCandidate := range ia.remoteCandidates {
			select {
			case <-ia.ctx.Done():
				return
			default:
				pair := &CandidatePair{
					Local:  localCandidate,
					Remote: remoteCandidate,
					State:  CandidatePairStateInProgress,
				}

				if ia.checkConnectivity(pair) {
					ia.mu.Lock()
					ia.selectedPair = pair
					ia.state = ICEStateConnected
					ia.mu.Unlock()

					ia.logger.Info("ICE connectivity check succeeded",
						zap.String("local_address", localCandidate.Address),
						zap.String("remote_address", remoteCandidate.Address))

					return
				}
			}
		}
	}

	ia.mu.Lock()
	ia.state = ICEStateFailed
	ia.mu.Unlock()

	ia.logger.Error("All ICE connectivity checks failed")
}

// checkConnectivity проверяет connectivity для пары candidates
func (ia *ICEAgent) checkConnectivity(pair *CandidatePair) bool {
	// Простая проверка - пытаемся отправить STUN binding request
	stunClient := NewSTUNClient(fmt.Sprintf("%s:%d", pair.Remote.Address, pair.Remote.Port), ia.logger)

	ctx, cancel := context.WithTimeout(ia.ctx, 5*time.Second)
	defer cancel()

	_, err := stunClient.SendBindingRequest(ctx)
	return err == nil
}

// GetSelectedPair возвращает выбранную пару candidates
func (ia *ICEAgent) GetSelectedPair() *CandidatePair {
	ia.mu.RLock()
	defer ia.mu.RUnlock()
	return ia.selectedPair
}

// GetLocalCandidates возвращает локальные candidates
func (ia *ICEAgent) GetLocalCandidates() []*Candidate {
	ia.mu.RLock()
	defer ia.mu.RUnlock()
	return append([]*Candidate(nil), ia.localCandidates...)
}

// GetState возвращает текущее состояние ICE агента
func (ia *ICEAgent) GetState() ICEState {
	ia.mu.RLock()
	defer ia.mu.RUnlock()
	return ia.state
}

// Close закрывает ICE агент
func (ia *ICEAgent) Close() error {
	ia.cancel()

	ia.mu.Lock()
	ia.state = ICEStateDisconnected
	ia.mu.Unlock()

	ia.logger.Info("ICE agent closed")
	return nil
}

// sortCandidatesByPriority сортирует candidates по приоритету
func (ia *ICEAgent) sortCandidatesByPriority() {
	// Простая сортировка по убыванию приоритета
	for i := 0; i < len(ia.localCandidates)-1; i++ {
		for j := i + 1; j < len(ia.localCandidates); j++ {
			if ia.localCandidates[i].Priority < ia.localCandidates[j].Priority {
				ia.localCandidates[i], ia.localCandidates[j] = ia.localCandidates[j], ia.localCandidates[i]
			}
		}
	}
}

// String возвращает строковое представление ICEState
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

// String возвращает строковое представление CandidateType
func (t CandidateType) String() string {
	switch t {
	case CandidateTypeHost:
		return "host"
	case CandidateTypeServerReflexive:
		return "srflx"
	case CandidateTypePeerReflexive:
		return "prflx"
	case CandidateTypeRelay:
		return "relay"
	default:
		return "unknown"
	}
}

// GetMetrics возвращает метрики ICE агента
func (ia *ICEAgent) GetMetrics() *NATMetrics {
	return ia.metrics
}

// RecordConnectionAttempt записывает попытку соединения
func (ia *ICEAgent) RecordConnectionAttempt(success bool, duration time.Duration) {
	natType := ia.detectNATType()
	status := "failed"
	if success {
		status = "success"
	}

	ia.metrics.RecordICEConnection(status, natType, duration)
	ia.metrics.RecordNATType(natType)
}

// detectNATType определяет тип NAT на основе собранных кандидатов
func (ia *ICEAgent) detectNATType() string {
	ia.mu.RLock()
	defer ia.mu.RUnlock()

	hasHost := false
	hasServerReflexive := false
	hasRelay := false

	for _, candidate := range ia.localCandidates {
		switch candidate.Type {
		case CandidateTypeHost:
			hasHost = true
		case CandidateTypeServerReflexive:
			hasServerReflexive = true
		case CandidateTypeRelay:
			hasRelay = true
		}
	}

	// Определяем тип NAT на основе доступных кандидатов
	if hasHost && !hasServerReflexive && !hasRelay {
		return "no_nat" // Прямое соединение
	} else if hasHost && hasServerReflexive && !hasRelay {
		return "cone_nat" // Cone NAT
	} else if hasHost && hasServerReflexive && hasRelay {
		return "symmetric_nat" // Symmetric NAT
	} else if !hasHost && hasServerReflexive {
		return "restricted_nat" // Restricted NAT
	} else if !hasHost && !hasServerReflexive && hasRelay {
		return "blocked_nat" // Полностью заблокированный NAT
	}

	return "unknown"
}

// GetNATTypeStats возвращает статистику типов NAT
func (ia *ICEAgent) GetNATTypeStats() map[string]int64 {
	return ia.metrics.GetNATTypeStats()
}

// GetConnectionStats возвращает статистику соединений
func (ia *ICEAgent) GetConnectionStats() map[string]*ConnectionStats {
	return ia.metrics.GetConnectionStats()
}

// GetOverallStats возвращает общую статистику
func (ia *ICEAgent) GetOverallStats() map[string]interface{} {
	return ia.metrics.GetOverallStats()
}

// calculatePriority рассчитывает приоритет кандидата согласно RFC 8445
func (ia *ICEAgent) calculatePriority(candidateType CandidateType, componentID int) uint32 {
	// Приоритет рассчитывается по формуле: priority = (2^24)*(type preference) + (2^8)*(local preference) + (2^0)*(256 - component ID)

	var typePreference uint32
	switch candidateType {
	case CandidateTypeHost:
		typePreference = 126 // Host candidates имеют наивысший приоритет
	case CandidateTypePeerReflexive:
		typePreference = 110 // Peer reflexive candidates
	case CandidateTypeServerReflexive:
		typePreference = 100 // Server reflexive candidates
	case CandidateTypeRelay:
		typePreference = 0 // Relay candidates имеют наименьший приоритет
	default:
		typePreference = 0
	}

	localPreference := uint32(65535) // Максимальное значение local preference

	priority := (typePreference << 24) + (localPreference << 8) + uint32(256-componentID)
	return priority
}

// AddLocalCandidate добавляет локальный кандидат
func (ia *ICEAgent) AddLocalCandidate(candidate *Candidate) {
	ia.mu.Lock()
	defer ia.mu.Unlock()

	ia.localCandidates = append(ia.localCandidates, candidate)

	// Записываем метрику
	ia.metrics.RecordICECandidate(candidate.Type.String(), candidate.Transport)
}

// DetectNATType определяет тип NAT (публичный метод)
func (ia *ICEAgent) DetectNATType() string {
	return ia.detectNATType()
}
