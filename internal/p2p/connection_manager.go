package p2p

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/internal/nat"
	"go.uber.org/zap"
)

// P2PConnectionManagerInterface интерфейс для управления P2P соединениями
type P2PConnectionManagerInterface interface {
	RegisterPeer(peer *Peer) error
	EstablishConnection(ctx context.Context, sessionID, peer1ID, peer2ID, protocol string) (*P2PConnection, error)
	GetConnection(sessionID string) (*P2PConnection, bool)
	CloseConnection(sessionID string) error
	GetActiveConnections() []*P2PConnection
	Close() error
}

// P2PConnectionManager управляет P2P соединениями
type P2PConnectionManager struct {
	activeConnections map[string]*P2PConnection
	iceAgents         map[string]ICEAgentInterface
	relayConnections  map[string]*RelayConnection
	peerStore         *PeerStore
	mu                sync.RWMutex
	logger            *zap.Logger
	ctx               context.Context
	cancel            context.CancelFunc
}

// P2PConnection представляет P2P соединение
type P2PConnection struct {
	SessionID        string
	Peer1ID          string
	Peer2ID          string
	ConnectionType   string // "direct", "relay", "masque"
	Protocol         string // "quic", "websocket", "tcp"
	State            ConnectionState
	ICEPair          *nat.CandidatePair
	EstablishedAt    time.Time
	LastActivity     time.Time
	BytesTransferred int64
	Latency          time.Duration
}

// ConnectionState состояние P2P соединения
type ConnectionState int

const (
	ConnectionStateNew ConnectionState = iota
	ConnectionStateConnecting
	ConnectionStateConnected
	ConnectionStateFailed
	ConnectionStateClosed
)

// RelayConnection представляет relay соединение
type RelayConnection struct {
	SessionID     string
	Peer1ID       string
	Peer2ID       string
	Protocol      string
	StreamBridge  *StreamBridge
	EstablishedAt time.Time
	LastActivity  time.Time
}

// StreamBridge мост для потоков через relay
type StreamBridge struct {
	SessionID    string
	Peer1Stream  interface{} // QUIC Stream или WebSocket Conn
	Peer2Stream  interface{} // QUIC Stream или WebSocket Conn
	BytesRelayed int64
	StartTime    time.Time
}

// PeerStore хранилище пиров
type PeerStore struct {
	peers map[string]*Peer
	mu    sync.RWMutex
}

// NewP2PConnectionManager создает новый P2P Connection Manager
func NewP2PConnectionManager(logger *zap.Logger) *P2PConnectionManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &P2PConnectionManager{
		activeConnections: make(map[string]*P2PConnection),
		iceAgents:         make(map[string]ICEAgentInterface),
		relayConnections:  make(map[string]*RelayConnection),
		peerStore:         &PeerStore{peers: make(map[string]*Peer)},
		logger:            logger,
		ctx:               ctx,
		cancel:            cancel,
	}
}

// RegisterPeer регистрирует пира в P2P Connection Manager
func (pcm *P2PConnectionManager) RegisterPeer(peer *Peer) error {
	pcm.peerStore.mu.Lock()
	defer pcm.peerStore.mu.Unlock()

	pcm.peerStore.peers[peer.PeerID] = peer

	// Создаем ICE агент для пира
	stunServers := peer.STUNServers
	if len(stunServers) == 0 {
		stunServers = []string{
			"stun.l.google.com:19302",
			"stun1.l.google.com:19302",
		}
	}

	iceAgent := nat.NewICEAgent(stunServers, nil, pcm.logger)
	pcm.iceAgents[peer.PeerID] = iceAgent

	if pcm.logger != nil {
		pcm.logger.Info("Peer registered in P2P Connection Manager",
			zap.String("peer_id", peer.PeerID),
			zap.Strings("stun_servers", stunServers),
			zap.Bool("masque_support", peer.MASQUESupport))
	}

	return nil
}

// EstablishConnection устанавливает P2P соединение
func (pcm *P2PConnectionManager) EstablishConnection(
	ctx context.Context,
	sessionID string,
	peer1ID string,
	peer2ID string,
	protocol string,
) (*P2PConnection, error) {
	pcm.mu.Lock()
	defer pcm.mu.Unlock()

	// Проверяем, что соединение еще не существует
	if _, exists := pcm.activeConnections[sessionID]; exists {
		return nil, fmt.Errorf("connection with session ID %s already exists", sessionID)
	}

	// Получаем информацию о пирах
	peer1, err := pcm.getPeer(peer1ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get peer1: %w", err)
	}

	peer2, err := pcm.getPeer(peer2ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get peer2: %w", err)
	}

	// Создаем P2P соединение
	connection := &P2PConnection{
		SessionID:     sessionID,
		Peer1ID:       peer1ID,
		Peer2ID:       peer2ID,
		Protocol:      protocol,
		State:         ConnectionStateNew,
		EstablishedAt: time.Now(),
		LastActivity:  time.Now(),
	}

	pcm.activeConnections[sessionID] = connection

	// Пытаемся установить прямое соединение
	if err := pcm.establishDirectConnection(ctx, connection, peer1, peer2); err != nil {
		if pcm.logger != nil {
			pcm.logger.Warn("Failed to establish direct connection, trying relay",
				zap.String("session_id", sessionID),
				zap.Error(err))
		}

		// Fallback на relay соединение
		if err := pcm.establishRelayConnection(ctx, connection, peer1, peer2); err != nil {
			connection.State = ConnectionStateFailed
			return nil, fmt.Errorf("failed to establish relay connection: %w", err)
		}
	}

	connection.State = ConnectionStateConnected
	connection.LastActivity = time.Now()

	if pcm.logger != nil {
		pcm.logger.Info("P2P connection established",
			zap.String("session_id", sessionID),
			zap.String("connection_type", connection.ConnectionType),
			zap.String("protocol", protocol))
	}

	return connection, nil
}

// establishDirectConnection пытается установить прямое P2P соединение
func (pcm *P2PConnectionManager) establishDirectConnection(
	ctx context.Context,
	connection *P2PConnection,
	peer1, peer2 *Peer,
) error {
	// Получаем ICE агенты для обоих пиров
	iceAgent1, exists1 := pcm.iceAgents[peer1.PeerID]
	iceAgent2, exists2 := pcm.iceAgents[peer2.PeerID]

	if !exists1 || !exists2 {
		return fmt.Errorf("ICE agents not found for peers")
	}

	// Начинаем сбор candidates для обоих пиров
	if err := iceAgent1.StartGathering(); err != nil {
		return fmt.Errorf("failed to start gathering for peer1: %w", err)
	}

	if err := iceAgent2.StartGathering(); err != nil {
		return fmt.Errorf("failed to start gathering for peer2: %w", err)
	}

	// Обмениваемся candidates
	candidates1 := iceAgent1.GetLocalCandidates()
	candidates2 := iceAgent2.GetLocalCandidates()

	for _, candidate := range candidates1 {
		iceAgent2.AddRemoteCandidate(candidate)
	}

	for _, candidate := range candidates2 {
		iceAgent1.AddRemoteCandidate(candidate)
	}

	// Начинаем connectivity checks
	if err := iceAgent1.StartConnectivityChecks(); err != nil {
		return fmt.Errorf("failed to start connectivity checks for peer1: %w", err)
	}

	if err := iceAgent2.StartConnectivityChecks(); err != nil {
		return fmt.Errorf("failed to start connectivity checks for peer2: %w", err)
	}

	// Ждем завершения ICE процесса
	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return fmt.Errorf("ICE connectivity checks timeout")
		case <-ticker.C:
			state1 := iceAgent1.GetState()
			state2 := iceAgent2.GetState()

			if state1 == nat.ICEStateConnected && state2 == nat.ICEStateConnected {
				// Получаем выбранную пару candidates
				pair1 := iceAgent1.GetSelectedPair()
				if pair1 != nil {
					connection.ICEPair = pair1
					connection.ConnectionType = "direct"
					return nil
				}
			}

			if state1 == nat.ICEStateFailed || state2 == nat.ICEStateFailed {
				return fmt.Errorf("ICE connectivity checks failed")
			}
		}
	}
}

// establishRelayConnection устанавливает relay соединение
func (pcm *P2PConnectionManager) establishRelayConnection(
	ctx context.Context,
	connection *P2PConnection,
	peer1, peer2 *Peer,
) error {
	// Создаем relay соединение
	relayConn := &RelayConnection{
		SessionID:     connection.SessionID,
		Peer1ID:       peer1.PeerID,
		Peer2ID:       peer2.PeerID,
		Protocol:      connection.Protocol,
		EstablishedAt: time.Now(),
		LastActivity:  time.Now(),
	}

	// Создаем stream bridge
	streamBridge := &StreamBridge{
		SessionID: connection.SessionID,
		StartTime: time.Now(),
	}

	relayConn.StreamBridge = streamBridge
	pcm.relayConnections[connection.SessionID] = relayConn

	connection.ConnectionType = "relay"

	if pcm.logger != nil {
		pcm.logger.Info("Relay connection established",
			zap.String("session_id", connection.SessionID),
			zap.String("protocol", connection.Protocol))
	}

	return nil
}

// GetConnection возвращает P2P соединение по session ID
func (pcm *P2PConnectionManager) GetConnection(sessionID string) (*P2PConnection, bool) {
	pcm.mu.RLock()
	defer pcm.mu.RUnlock()

	connection, exists := pcm.activeConnections[sessionID]
	return connection, exists
}

// CloseConnection закрывает P2P соединение
func (pcm *P2PConnectionManager) CloseConnection(sessionID string) error {
	pcm.mu.Lock()
	defer pcm.mu.Unlock()

	connection, exists := pcm.activeConnections[sessionID]
	if !exists {
		return fmt.Errorf("connection with session ID %s not found", sessionID)
	}

	connection.State = ConnectionStateClosed

	// Закрываем relay соединение если есть
	if _, exists := pcm.relayConnections[sessionID]; exists {
		// Здесь должна быть логика закрытия stream bridge
		delete(pcm.relayConnections, sessionID)
	}

	delete(pcm.activeConnections, sessionID)

	if pcm.logger != nil {
		pcm.logger.Info("P2P connection closed",
			zap.String("session_id", sessionID))
	}

	return nil
}

// GetActiveConnections возвращает список активных соединений
func (pcm *P2PConnectionManager) GetActiveConnections() []*P2PConnection {
	pcm.mu.RLock()
	defer pcm.mu.RUnlock()

	connections := make([]*P2PConnection, 0, len(pcm.activeConnections))
	for _, conn := range pcm.activeConnections {
		if conn.State == ConnectionStateConnected {
			connections = append(connections, conn)
		}
	}

	return connections
}

// getPeer возвращает пира по ID
func (pcm *P2PConnectionManager) getPeer(peerID string) (*Peer, error) {
	pcm.peerStore.mu.RLock()
	defer pcm.peerStore.mu.RUnlock()

	peer, exists := pcm.peerStore.peers[peerID]
	if !exists {
		return nil, fmt.Errorf("peer %s not found", peerID)
	}

	return peer, nil
}

// convertToNATCandidate конвертирует ICECandidate в NAT Candidate
func convertToNATCandidate(apiCandidate *ICECandidate) *nat.Candidate {
	var candidateType nat.CandidateType
	switch apiCandidate.Type {
	case "host":
		candidateType = nat.CandidateTypeHost
	case "srflx":
		candidateType = nat.CandidateTypeServerReflexive
	case "prflx":
		candidateType = nat.CandidateTypePeerReflexive
	case "relay":
		candidateType = nat.CandidateTypeRelay
	default:
		candidateType = nat.CandidateTypeHost
	}

	return &nat.Candidate{
		Foundation:  apiCandidate.Foundation,
		Component:   apiCandidate.Component,
		Transport:   apiCandidate.Transport,
		Priority:    apiCandidate.Priority,
		Address:     apiCandidate.Address,
		Port:        apiCandidate.Port,
		Type:        candidateType,
		RelatedAddr: apiCandidate.RelatedAddr,
		RelatedPort: apiCandidate.RelatedPort,
		Username:    apiCandidate.Username,
		Password:    apiCandidate.Password,
	}
}

// convertToAPICandidate конвертирует NAT Candidate в ICECandidate
func convertToAPICandidate(natCandidate *nat.Candidate) *ICECandidate {
	var candidateType string
	switch natCandidate.Type {
	case nat.CandidateTypeHost:
		candidateType = "host"
	case nat.CandidateTypeServerReflexive:
		candidateType = "srflx"
	case nat.CandidateTypePeerReflexive:
		candidateType = "prflx"
	case nat.CandidateTypeRelay:
		candidateType = "relay"
	default:
		candidateType = "host"
	}

	return &ICECandidate{
		Foundation:  natCandidate.Foundation,
		Component:   natCandidate.Component,
		Transport:   natCandidate.Transport,
		Priority:    natCandidate.Priority,
		Address:     natCandidate.Address,
		Port:        natCandidate.Port,
		Type:        candidateType,
		RelatedAddr: natCandidate.RelatedAddr,
		RelatedPort: natCandidate.RelatedPort,
		Username:    natCandidate.Username,
		Password:    natCandidate.Password,
	}
}

// Close закрывает P2P Connection Manager
func (pcm *P2PConnectionManager) Close() error {
	pcm.cancel()

	// Закрываем все ICE агенты
	for _, iceAgent := range pcm.iceAgents {
		iceAgent.Close()
	}

	// Закрываем все активные соединения
	for sessionID := range pcm.activeConnections {
		pcm.CloseConnection(sessionID)
	}

	if pcm.logger != nil {
		pcm.logger.Info("P2P Connection Manager closed")
	}
	return nil
}

// String возвращает строковое представление ConnectionState
func (s ConnectionState) String() string {
	switch s {
	case ConnectionStateNew:
		return "new"
	case ConnectionStateConnecting:
		return "connecting"
	case ConnectionStateConnected:
		return "connected"
	case ConnectionStateFailed:
		return "failed"
	case ConnectionStateClosed:
		return "closed"
	default:
		return "unknown"
	}
}
