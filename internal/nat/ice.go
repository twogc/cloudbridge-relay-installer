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
	localCandidates  []*Candidate
	remoteCandidates []*Candidate
	selectedPair     *CandidatePair
	state            ICEState
	mu               sync.RWMutex
	logger           *zap.Logger
	ctx              context.Context
	cancel           context.CancelFunc
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

	return &ICEAgent{
		stunServers:      stunServers,
		turnServers:      turnServers,
		localCandidates:  make([]*Candidate, 0),
		remoteCandidates: make([]*Candidate, 0),
		state:            ICEStateNew,
		logger:           logger,
		ctx:              ctx,
		cancel:           cancel,
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

// gatherHostCandidates собирает host candidates
func (ia *ICEAgent) gatherHostCandidates() error {
	interfaces, err := net.Interfaces()
	if err != nil {
		return fmt.Errorf("failed to get network interfaces: %w", err)
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
				if ipnet.IP.To4() != nil { // IPv4 only for now
					candidate := &Candidate{
						Foundation: fmt.Sprintf("%d", iface.Index),
						Component:  1,
						Transport:  "UDP",
						Priority:   ia.calculatePriority(CandidateTypeHost, 0),
						Address:    ipnet.IP.String(),
						Port:       0, // Will be set by QUIC connection
						Type:       CandidateTypeHost,
					}
					ia.localCandidates = append(ia.localCandidates, candidate)
				}
			}
		}
	}

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

// calculatePriority вычисляет приоритет candidate
func (ia *ICEAgent) calculatePriority(candidateType CandidateType, localPref int) uint32 {
	// RFC 5245 priority formula
	var typePref uint32
	switch candidateType {
	case CandidateTypeHost:
		typePref = 126
	case CandidateTypeServerReflexive:
		typePref = 100
	case CandidateTypePeerReflexive:
		typePref = 110
	case CandidateTypeRelay:
		typePref = 0
	}

	return (typePref << 24) | (uint32(localPref) << 8) | (256 - 1)
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

