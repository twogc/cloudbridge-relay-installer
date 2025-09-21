package main

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/pion/ice/v2"
	"github.com/pion/stun"
	"github.com/quic-go/quic-go"
	"go.uber.org/zap"

	"github.com/twogc/cloudbridge-relay-installer/config"
	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
)

// ICESupport представляет поддержку ICE/STUN для NAT traversal
type ICESupport struct {
	logger      *logger.Logger
	config      *config.Config
	agent       *ice.Agent
	stunServers []string
	turnServers []string

	// Управление соединениями
	connections      map[string]*ice.Conn
	connectionsMutex sync.RWMutex

	// Кандидаты
	localCandidates  []ice.Candidate
	remoteCandidates []ice.Candidate

	// Состояние
	isGathering bool
	isConnected bool
}

// NewICESupport создает новый ICE поддержку
func NewICESupport(logger *logger.Logger, config *config.Config) (*ICESupport, error) {
	iceConfig := &ice.AgentConfig{
		NetworkTypes: []ice.NetworkType{ice.NetworkTypeUDP4},
		Urls:         []*stun.URI{},
	}

	// Добавляем STUN серверы из конфигурации
	if config.ICEStunTurn.Enabled {
		for _, stunServer := range config.ICEStunTurn.StunServers {
			url, err := stun.ParseURI(fmt.Sprintf("stun:%s", stunServer))
			if err != nil {
				logger.Warn("Failed to parse STUN server URL", zap.String("server", stunServer), zap.Error(err))
				continue
			}
			iceConfig.Urls = append(iceConfig.Urls, url)
		}

		// Добавляем TURN серверы
		for _, turnServer := range config.ICEStunTurn.TurnServers {
			url, err := stun.ParseURI(fmt.Sprintf("turn:%s", turnServer))
			if err != nil {
				logger.Warn("Failed to parse TURN server URL", zap.String("server", turnServer), zap.Error(err))
				continue
			}
			iceConfig.Urls = append(iceConfig.Urls, url)
		}
	}

	agent, err := ice.NewAgent(iceConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create ICE agent: %v", err)
	}

	return &ICESupport{
		logger:           logger,
		config:           config,
		agent:            agent,
		stunServers:      config.ICEStunTurn.StunServers,
		turnServers:      config.ICEStunTurn.TurnServers,
		connections:      make(map[string]*ice.Conn),
		localCandidates:  make([]ice.Candidate, 0),
		remoteCandidates: make([]ice.Candidate, 0),
	}, nil
}

// Start запускает ICE поддержку
func (is *ICESupport) Start(ctx context.Context) error {
	is.logger.Info("Starting ICE support",
		zap.Strings("stun_servers", is.stunServers),
		zap.Strings("turn_servers", is.turnServers))

	// Настраиваем обработчики событий
	if err := is.setupEventHandlers(); err != nil {
		return fmt.Errorf("failed to setup ICE event handlers: %v", err)
	}

	// Начинаем сбор кандидатов
	if err := is.agent.GatherCandidates(); err != nil {
		return fmt.Errorf("failed to gather ICE candidates: %v", err)
	}

	is.isGathering = true
	is.logger.Info("ICE candidate gathering started")

	return nil
}

// setupEventHandlers настраивает обработчики событий ICE
func (is *ICESupport) setupEventHandlers() error {
	// Обработчик новых кандидатов
	if err := is.agent.OnCandidate(func(c ice.Candidate) {
		if c == nil {
			return
		}

		is.logger.Debug("ICE candidate gathered",
			zap.String("candidate", c.String()),
			zap.String("type", c.Type().String()),
			zap.String("address", c.Address()),
			zap.Int("port", c.Port()))

		is.localCandidates = append(is.localCandidates, c)
	}); err != nil {
		return err
	}

	// Обработчик изменения состояния соединения
	if err := is.agent.OnConnectionStateChange(func(c ice.ConnectionState) {
		is.logger.Info("ICE connection state changed", zap.String("state", c.String()))

		switch c {
		case ice.ConnectionStateConnected:
			is.isConnected = true
			is.logger.Info("ICE connection established successfully")
		case ice.ConnectionStateDisconnected:
			is.isConnected = false
			is.logger.Warn("ICE connection disconnected")
		case ice.ConnectionStateFailed:
			is.isConnected = false
			is.logger.Error("ICE connection failed")
		}
	}); err != nil {
		return err
	}

	// Обработчик завершения сбора кандидатов
	// В новой версии pion/ice этот метод может не существовать
	// Используем таймер для отслеживания завершения сбора
	go func() {
		time.Sleep(is.config.ICEStunTurn.GatheringTimeout)
		is.isGathering = false
		is.logger.Info("ICE candidate gathering completed",
			zap.Int("candidates_count", len(is.localCandidates)))
	}()

	return nil
}

// GetLocalCandidates возвращает локальные кандидаты
func (is *ICESupport) GetLocalCandidates() []ice.Candidate {
	is.connectionsMutex.RLock()
	defer is.connectionsMutex.RUnlock()

	return is.localCandidates
}

// AddRemoteCandidate добавляет удаленный кандидат
func (is *ICESupport) AddRemoteCandidate(candidate ice.Candidate) error {
	is.logger.Debug("Adding remote ICE candidate",
		zap.String("candidate", candidate.String()),
		zap.String("type", candidate.Type().String()))

	if err := is.agent.AddRemoteCandidate(candidate); err != nil {
		return fmt.Errorf("failed to add remote candidate: %v", err)
	}

	is.remoteCandidates = append(is.remoteCandidates, candidate)
	return nil
}

// CreateConnection создает ICE соединение
func (is *ICESupport) CreateConnection(ctx context.Context, remoteUfrag, remotePwd string) (*ice.Conn, error) {
	is.logger.Info("Creating ICE connection",
		zap.String("remote_ufrag", remoteUfrag))

	conn, err := is.agent.Dial(ctx, remoteUfrag, remotePwd)
	if err != nil {
		return nil, fmt.Errorf("failed to create ICE connection: %v", err)
	}

	// Сохраняем соединение
	connID := fmt.Sprintf("ice_%d", time.Now().UnixNano())
	is.connectionsMutex.Lock()
	is.connections[connID] = conn
	is.connectionsMutex.Unlock()

	is.logger.Info("ICE connection created successfully", zap.String("conn_id", connID))
	return conn, nil
}

// AcceptConnection принимает входящее ICE соединение
func (is *ICESupport) AcceptConnection(ctx context.Context, remoteUfrag, remotePwd string) (*ice.Conn, error) {
	is.logger.Info("Accepting ICE connection",
		zap.String("remote_ufrag", remoteUfrag))

	conn, err := is.agent.Accept(ctx, remoteUfrag, remotePwd)
	if err != nil {
		return nil, fmt.Errorf("failed to accept ICE connection: %v", err)
	}

	// Сохраняем соединение
	connID := fmt.Sprintf("ice_%d", time.Now().UnixNano())
	is.connectionsMutex.Lock()
	is.connections[connID] = conn
	is.connectionsMutex.Unlock()

	is.logger.Info("ICE connection accepted successfully", zap.String("conn_id", connID))
	return conn, nil
}

// GetSelectedCandidatePair возвращает выбранную пару кандидатов
func (is *ICESupport) GetSelectedCandidatePair() (*ice.CandidatePair, error) {
	pair, err := is.agent.GetSelectedCandidatePair()
	if err != nil {
		return nil, fmt.Errorf("failed to get selected candidate pair: %v", err)
	}

	if pair != nil {
		is.logger.Info("Selected ICE candidate pair",
			zap.String("local", pair.Local.String()),
			zap.String("remote", pair.Remote.String()))
	}

	return pair, nil
}

// CreateQUICTransport создает QUIC Transport поверх ICE соединения
func (is *ICESupport) CreateQUICTransport(iceConn *ice.Conn) (*quic.Transport, error) {
	// Создаем PacketConn обертку для ICE соединения
	packetConn := &ICEPacketConn{Conn: iceConn}

	// Создаем QUIC Transport
	transport := &quic.Transport{
		Conn: packetConn,
	}

	is.logger.Info("QUIC Transport created over ICE connection")
	return transport, nil
}

// Close закрывает ICE поддержку
func (is *ICESupport) Close() error {
	is.logger.Info("Closing ICE support")

	// Закрываем все соединения
	is.connectionsMutex.Lock()
	for connID, conn := range is.connections {
		if err := conn.Close(); err != nil {
			is.logger.Warn("Failed to close ICE connection",
				zap.String("conn_id", connID),
				zap.Error(err))
		}
	}
	is.connections = make(map[string]*ice.Conn)
	is.connectionsMutex.Unlock()

	// Закрываем ICE agent
	if is.agent != nil {
		if err := is.agent.Close(); err != nil {
			return fmt.Errorf("failed to close ICE agent: %v", err)
		}
	}

	is.logger.Info("ICE support closed successfully")
	return nil
}

// ICEPacketConn представляет PacketConn обертку для ICE соединения
type ICEPacketConn struct {
	*ice.Conn
}

// ReadFrom реализует net.PacketConn
func (c *ICEPacketConn) ReadFrom(p []byte) (int, net.Addr, error) {
	n, err := c.Read(p)
	return n, c.RemoteAddr(), err
}

// WriteTo реализует net.PacketConn
func (c *ICEPacketConn) WriteTo(p []byte, addr net.Addr) (int, error) {
	return c.Write(p)
}

// LocalAddr возвращает локальный адрес
func (c *ICEPacketConn) LocalAddr() net.Addr {
	return c.Conn.LocalAddr()
}

// RemoteAddr возвращает удаленный адрес
func (c *ICEPacketConn) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

// SetDeadline устанавливает deadline
func (c *ICEPacketConn) SetDeadline(t time.Time) error {
	return c.Conn.SetDeadline(t)
}

// SetReadDeadline устанавливает read deadline
func (c *ICEPacketConn) SetReadDeadline(t time.Time) error {
	return c.Conn.SetReadDeadline(t)
}

// SetWriteDeadline устанавливает write deadline
func (c *ICEPacketConn) SetWriteDeadline(t time.Time) error {
	return c.Conn.SetWriteDeadline(t)
}
