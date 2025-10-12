package nat

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/pion/ice/v3"
	"go.uber.org/zap"

	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
)

// PionICEAdapter адаптер для интеграции pion/ice с CloudBridge
type PionICEAdapter struct {
	logger *logger.Logger

	// Pion ICE Agent
	agent *ice.Agent

	// Конфигурация
	config *PionICEConfig

	// Метрики
	metrics *NATMetrics

	// Callbacks
	onCandidateFunc      func(*ice.Candidate)
	onConnectionFunc     func(ice.ConnectionState)
	onSelectedPairFunc   func(*ice.CandidatePair)
	onGatheringComplete  func()

	// Состояние
	mu                  sync.RWMutex
	connectionState     ice.ConnectionState
	gatheringState      ice.GatheringState
	localCandidates     []*ice.Candidate
	remoteCandidates    []*ice.Candidate
	selectedPair        *ice.CandidatePair
	gatheringComplete   bool

	// Контекст
	ctx    context.Context
	cancel context.CancelFunc
}

// PionICEConfig конфигурация для pion/ice
type PionICEConfig struct {
	// STUN серверы
	STUNServers []string

	// TURN серверы
	TURNServers []TURNServerConfig

	// Опции
	PortMin              uint16
	PortMax              uint16
	DisableTCP           bool
	DisableIPv6          bool
	NetworkTypes         []ice.NetworkType
	CandidateTypes       []ice.CandidateType
	InterfaceFilter      func(string) bool
	IPFilter             func(net.IP) bool
	IncludeLoopback      bool
	DisableActiveTCP     bool
	CheckInterval        time.Duration
	KeepaliveInterval    time.Duration
	DisconnectedTimeout  time.Duration
	FailedTimeout        time.Duration
	MaxBindingRequests   uint16
}

// TURNServerConfig конфигурация TURN сервера для pion/ice
type TURNServerConfig struct {
	URLs       []string
	Username   string
	Credential string
	Realm      string
}

// DefaultPionICEConfig возвращает конфигурацию по умолчанию
func DefaultPionICEConfig() *PionICEConfig {
	return &PionICEConfig{
		STUNServers: []string{
			"stun:stun.l.google.com:19302",
			"stun:stun1.l.google.com:19302",
		},
		TURNServers:          []TURNServerConfig{},
		PortMin:              16384,
		PortMax:              32768,
		DisableTCP:           false,
		DisableIPv6:          false,
		NetworkTypes:         []ice.NetworkType{ice.NetworkTypeUDP4, ice.NetworkTypeUDP6, ice.NetworkTypeTCP4, ice.NetworkTypeTCP6},
		CandidateTypes:       []ice.CandidateType{ice.CandidateTypeHost, ice.CandidateTypeServerReflexive, ice.CandidateTypeRelay},
		IncludeLoopback:      false,
		DisableActiveTCP:     false,
		CheckInterval:        200 * time.Millisecond,
		KeepaliveInterval:    2 * time.Second,
		DisconnectedTimeout:  5 * time.Second,
		FailedTimeout:        30 * time.Second,
		MaxBindingRequests:   7,
	}
}

// NewPionICEAdapter создает новый адаптер pion/ice
func NewPionICEAdapter(logger *logger.Logger, config *PionICEConfig, metrics *NATMetrics) (*PionICEAdapter, error) {
	if config == nil {
		config = DefaultPionICEConfig()
	}

	if metrics == nil {
		metrics = NewNATMetrics()
	}

	ctx, cancel := context.WithCancel(context.Background())

	adapter := &PionICEAdapter{
		logger:              logger,
		config:              config,
		metrics:             metrics,
		localCandidates:     make([]*ice.Candidate, 0),
		remoteCandidates:    make([]*ice.Candidate, 0),
		connectionState:     ice.ConnectionStateNew,
		gatheringState:      ice.GatheringStateNew,
		gatheringComplete:   false,
		ctx:                 ctx,
		cancel:              cancel,
	}

	// Создаем pion ICE agent
	if err := adapter.createAgent(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create pion ICE agent: %w", err)
	}

	return adapter, nil
}

// createAgent создает pion ICE agent с конфигурацией
func (pa *PionICEAdapter) createAgent() error {
	// Создаем конфигурацию pion ICE
	agentConfig := &ice.AgentConfig{
		NetworkTypes:        pa.config.NetworkTypes,
		CandidateTypes:      pa.config.CandidateTypes,
		InterfaceFilter:     pa.config.InterfaceFilter,
		IPFilter:            pa.config.IPFilter,
		IncludeLoopback:     pa.config.IncludeLoopback,
		DisableActiveTCP:    pa.config.DisableActiveTCP,
		CheckInterval:       &pa.config.CheckInterval,
		KeepaliveInterval:   &pa.config.KeepaliveInterval,
		DisconnectedTimeout: &pa.config.DisconnectedTimeout,
		FailedTimeout:       &pa.config.FailedTimeout,
		MaxBindingRequests:  &pa.config.MaxBindingRequests,
	}

	// Добавляем STUN серверы
	if len(pa.config.STUNServers) > 0 {
		agentConfig.Urls = make([]*ice.URL, 0, len(pa.config.STUNServers))
		for _, stunURL := range pa.config.STUNServers {
			url, err := ice.ParseURL(stunURL)
			if err != nil {
				pa.logger.Warn("Failed to parse STUN URL",
					zap.String("url", stunURL),
					zap.Error(err))
				continue
			}
			agentConfig.Urls = append(agentConfig.Urls, url)
		}
	}

	// Добавляем TURN серверы
	for _, turnConfig := range pa.config.TURNServers {
		for _, turnURL := range turnConfig.URLs {
			url, err := ice.ParseURL(turnURL)
			if err != nil {
				pa.logger.Warn("Failed to parse TURN URL",
					zap.String("url", turnURL),
					zap.Error(err))
				continue
			}
			url.Username = turnConfig.Username
			url.Password = turnConfig.Credential
			agentConfig.Urls = append(agentConfig.Urls, url)
		}
	}

	// Создаем agent
	agent, err := ice.NewAgent(agentConfig)
	if err != nil {
		return fmt.Errorf("failed to create ICE agent: %w", err)
	}

	pa.agent = agent

	// Устанавливаем callbacks
	pa.setupCallbacks()

	return nil
}

// setupCallbacks настраивает callbacks для pion ICE agent
func (pa *PionICEAdapter) setupCallbacks() {
	// Callback для новых кандидатов
	pa.agent.OnCandidate(func(candidate ice.Candidate) {
		if candidate == nil {
			pa.mu.Lock()
			pa.gatheringComplete = true
			pa.gatheringState = ice.GatheringStateComplete
			pa.mu.Unlock()

			pa.metrics.FinishGathering()
			pa.logger.Info("ICE candidate gathering completed",
				zap.Int("total_candidates", len(pa.localCandidates)))

			if pa.onGatheringComplete != nil {
				pa.onGatheringComplete()
			}
			return
		}

		pa.mu.Lock()
		pa.localCandidates = append(pa.localCandidates, &candidate)
		pa.mu.Unlock()

		// Обновляем метрики
		switch candidate.Type() {
		case ice.CandidateTypeHost:
			pa.metrics.mu.Lock()
			pa.metrics.HostCandidatesGathered++
			pa.metrics.mu.Unlock()
		case ice.CandidateTypeServerReflexive:
			pa.metrics.mu.Lock()
			pa.metrics.ServerReflexiveCandidatesGathered++
			pa.metrics.mu.Unlock()
		case ice.CandidateTypeRelay:
			pa.metrics.mu.Lock()
			pa.metrics.RelayCandidatesGathered++
			pa.metrics.mu.Unlock()
		}

		pa.logger.Debug("New ICE candidate",
			zap.String("type", candidate.Type().String()),
			zap.String("address", candidate.Address()),
			zap.Int("port", candidate.Port()),
			zap.String("protocol", candidate.NetworkType().NetworkShort()),
			zap.Uint32("priority", candidate.Priority()))

		if pa.onCandidateFunc != nil {
			pa.onCandidateFunc(&candidate)
		}
	})

	// Callback для изменения состояния соединения
	pa.agent.OnConnectionStateChange(func(state ice.ConnectionState) {
		pa.mu.Lock()
		oldState := pa.connectionState
		pa.connectionState = state
		pa.mu.Unlock()

		pa.logger.Info("ICE connection state changed",
			zap.String("old_state", oldState.String()),
			zap.String("new_state", state.String()))

		// Обновляем метрики
		switch state {
		case ice.ConnectionStateConnected, ice.ConnectionStateCompleted:
			pa.metrics.UpdateConnectivityMetrics(true, time.Since(pa.metrics.GatheringStartTime))
		case ice.ConnectionStateFailed, ice.ConnectionStateDisconnected:
			pa.metrics.UpdateConnectivityMetrics(false, time.Since(pa.metrics.GatheringStartTime))
		}

		if pa.onConnectionFunc != nil {
			pa.onConnectionFunc(state)
		}
	})

	// Callback для выбранной пары кандидатов
	pa.agent.OnSelectedCandidatePairChange(func(local, remote ice.Candidate) {
		pair := &ice.CandidatePair{
			Local:  local,
			Remote: remote,
		}

		pa.mu.Lock()
		pa.selectedPair = pair
		pa.mu.Unlock()

		pa.logger.Info("Selected ICE candidate pair",
			zap.String("local_type", local.Type().String()),
			zap.String("local_address", local.Address()),
			zap.Int("local_port", local.Port()),
			zap.String("remote_type", remote.Type().String()),
			zap.String("remote_address", remote.Address()),
			zap.Int("remote_port", remote.Port()))

		if pa.onSelectedPairFunc != nil {
			pa.onSelectedPairFunc(pair)
		}
	})
}

// GatherCandidates запускает сбор ICE кандидатов
func (pa *PionICEAdapter) GatherCandidates() error {
	pa.mu.Lock()
	if pa.gatheringState != ice.GatheringStateNew {
		pa.mu.Unlock()
		return fmt.Errorf("gathering already started or completed")
	}
	pa.gatheringState = ice.GatheringStateGathering
	pa.metrics.GatheringStartTime = time.Now()
	pa.mu.Unlock()

	pa.logger.Info("Starting ICE candidate gathering")

	// Запускаем сбор кандидатов
	if err := pa.agent.GatherCandidates(); err != nil {
		return fmt.Errorf("failed to gather candidates: %w", err)
	}

	return nil
}

// AddRemoteCandidate добавляет удаленный кандидат
func (pa *PionICEAdapter) AddRemoteCandidate(candidateStr string) error {
	candidate, err := ice.UnmarshalCandidate(candidateStr)
	if err != nil {
		return fmt.Errorf("failed to unmarshal candidate: %w", err)
	}

	if err := pa.agent.AddRemoteCandidate(candidate); err != nil {
		return fmt.Errorf("failed to add remote candidate: %w", err)
	}

	pa.mu.Lock()
	pa.remoteCandidates = append(pa.remoteCandidates, &candidate)
	pa.mu.Unlock()

	pa.logger.Debug("Added remote ICE candidate",
		zap.String("type", candidate.Type().String()),
		zap.String("address", candidate.Address()),
		zap.Int("port", candidate.Port()))

	return nil
}

// GetLocalCandidates возвращает локальные кандидаты
func (pa *PionICEAdapter) GetLocalCandidates() []ice.Candidate {
	pa.mu.RLock()
	defer pa.mu.RUnlock()

	candidates := make([]ice.Candidate, 0, len(pa.localCandidates))
	for _, candidate := range pa.localCandidates {
		candidates = append(candidates, *candidate)
	}
	return candidates
}

// GetConnectionState возвращает текущее состояние соединения
func (pa *PionICEAdapter) GetConnectionState() ice.ConnectionState {
	pa.mu.RLock()
	defer pa.mu.RUnlock()
	return pa.connectionState
}

// GetGatheringState возвращает текущее состояние сбора кандидатов
func (pa *PionICEAdapter) GetGatheringState() ice.GatheringState {
	pa.mu.RLock()
	defer pa.mu.RUnlock()
	return pa.gatheringState
}

// GetSelectedPair возвращает выбранную пару кандидатов
func (pa *PionICEAdapter) GetSelectedPair() *ice.CandidatePair {
	pa.mu.RLock()
	defer pa.mu.RUnlock()
	return pa.selectedPair
}

// GetLocalParameters возвращает локальные ICE параметры
func (pa *PionICEAdapter) GetLocalParameters() (ice.Role, string, string, error) {
	ufrag, pwd, err := pa.agent.GetLocalUserCredentials()
	if err != nil {
		return ice.Controlled, "", "", fmt.Errorf("failed to get local credentials: %w", err)
	}
	return ice.Controlling, ufrag, pwd, nil
}

// SetRemoteCredentials устанавливает удаленные ICE credentials
func (pa *PionICEAdapter) SetRemoteCredentials(ufrag, pwd string) error {
	if err := pa.agent.SetRemoteCredentials(ufrag, pwd); err != nil {
		return fmt.Errorf("failed to set remote credentials: %w", err)
	}
	pa.logger.Debug("Set remote ICE credentials",
		zap.String("ufrag", ufrag))
	return nil
}

// Dial создает исходящее соединение
func (pa *PionICEAdapter) Dial(ctx context.Context, remoteUfrag, remotePwd string) (*ice.Conn, error) {
	// Устанавливаем удаленные credentials
	if err := pa.SetRemoteCredentials(remoteUfrag, remotePwd); err != nil {
		return nil, err
	}

	// Создаем соединение
	conn, err := pa.agent.Dial(ctx, remoteUfrag, remotePwd)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %w", err)
	}

	pa.logger.Info("ICE connection established (dial)")
	return conn, nil
}

// Accept принимает входящее соединение
func (pa *PionICEAdapter) Accept(ctx context.Context, remoteUfrag, remotePwd string) (*ice.Conn, error) {
	// Устанавливаем удаленные credentials
	if err := pa.SetRemoteCredentials(remoteUfrag, remotePwd); err != nil {
		return nil, err
	}

	// Принимаем соединение
	conn, err := pa.agent.Accept(ctx, remoteUfrag, remotePwd)
	if err != nil {
		return nil, fmt.Errorf("failed to accept: %w", err)
	}

	pa.logger.Info("ICE connection established (accept)")
	return conn, nil
}

// OnCandidate устанавливает callback для новых кандидатов
func (pa *PionICEAdapter) OnCandidate(f func(*ice.Candidate)) {
	pa.onCandidateFunc = f
}

// OnConnectionStateChange устанавливает callback для изменения состояния соединения
func (pa *PionICEAdapter) OnConnectionStateChange(f func(ice.ConnectionState)) {
	pa.onConnectionFunc = f
}

// OnSelectedCandidatePair устанавливает callback для выбранной пары кандидатов
func (pa *PionICEAdapter) OnSelectedCandidatePair(f func(*ice.CandidatePair)) {
	pa.onSelectedPairFunc = f
}

// OnGatheringComplete устанавливает callback для завершения сбора кандидатов
func (pa *PionICEAdapter) OnGatheringComplete(f func()) {
	pa.onGatheringComplete = f
}

// GetMetrics возвращает метрики NAT traversal
func (pa *PionICEAdapter) GetMetrics() *NATMetrics {
	return pa.metrics.Clone()
}

// GetStats возвращает статистику pion ICE agent
func (pa *PionICEAdapter) GetStats() map[string]interface{} {
	// В pion/ice v3 нет метода GetStats, возвращаем базовую информацию
	pa.mu.RLock()
	defer pa.mu.RUnlock()
	
	return map[string]interface{}{
		"connection_state":     pa.connectionState.String(),
		"gathering_state":     pa.gatheringState.String(),
		"local_candidates":     len(pa.localCandidates),
		"remote_candidates":    len(pa.remoteCandidates),
		"gathering_complete":  pa.gatheringComplete,
		"selected_pair":       pa.selectedPair != nil,
	}
}

// Restart перезапускает ICE agent
func (pa *PionICEAdapter) Restart() error {
	pa.mu.Lock()
	defer pa.mu.Unlock()

	if err := pa.agent.Restart("", ""); err != nil {
		return fmt.Errorf("failed to restart ICE agent: %w", err)
	}

	pa.localCandidates = make([]*ice.Candidate, 0)
	pa.remoteCandidates = make([]*ice.Candidate, 0)
	pa.selectedPair = nil
	pa.gatheringComplete = false
	pa.gatheringState = ice.GatheringStateNew
	pa.connectionState = ice.ConnectionStateNew

	pa.logger.Info("ICE agent restarted")
	return nil
}

// Close закрывает адаптер и освобождает ресурсы
func (pa *PionICEAdapter) Close() error {
	pa.cancel()

	if pa.agent != nil {
		if err := pa.agent.Close(); err != nil {
			pa.logger.Error("Failed to close ICE agent", zap.Error(err))
			return err
		}
	}

	pa.logger.Info("Pion ICE adapter closed")
	return nil
}

// IsGatheringComplete проверяет, завершен ли сбор кандидатов
func (pa *PionICEAdapter) IsGatheringComplete() bool {
	pa.mu.RLock()
	defer pa.mu.RUnlock()
	return pa.gatheringComplete
}

// WaitForGatheringComplete ожидает завершения сбора кандидатов
func (pa *PionICEAdapter) WaitForGatheringComplete(timeout time.Duration) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	timeoutTimer := time.NewTimer(timeout)
	defer timeoutTimer.Stop()

	for {
		select {
		case <-pa.ctx.Done():
			return fmt.Errorf("context cancelled")
		case <-timeoutTimer.C:
			return fmt.Errorf("gathering timeout after %v", timeout)
		case <-ticker.C:
			if pa.IsGatheringComplete() {
				return nil
			}
		}
	}
}
