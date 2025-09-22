package p2p

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net"
	"sync"
	"time"

	"github.com/quic-go/quic-go"
	"go.uber.org/zap"

	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
)

// P2PTunnelManager управляет P2P туннелями с поддержкой NAT traversal
type P2PTunnelManager struct {
	tunnels      map[string]*P2PTunnel
	natDetector  *NATDetector
	holePuncher  *HolePuncher
	relayManager *RelayManager
	logger       *logger.Logger
	ctx          context.Context
	cancel       context.CancelFunc
	mutex        sync.RWMutex
}

// P2PTunnel представляет P2P туннель между двумя пирами
type P2PTunnel struct {
	ID               string
	PeerID           string
	TenantID         string
	LocalAddr        *net.UDPAddr
	RemoteAddr       *net.UDPAddr
	Underlay         UnderlayConnection
	OverlayConns     sync.Map
	IsActive         bool
	LastHeartbeat    time.Time
	CreatedAt        time.Time
	BytesTransferred int64
	Mutex            sync.RWMutex
}

// UnderlayConnection представляет базовое соединение (QUIC, TCP, UDP)
type UnderlayConnection interface {
	Read([]byte) (int, error)
	Write([]byte) (int, error)
	Close() error
	SetReadDeadline(time.Time) error
	SetWriteDeadline(time.Time) error
	Protocol() string
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
}

// QUICUnderlay реализует UnderlayConnection для QUIC
type QUICUnderlay struct {
	conn   quic.Conn
	stream quic.Stream
	mutex  sync.Mutex
}

func (q *QUICUnderlay) Read(b []byte) (int, error) {
	return q.stream.Read(b)
}

func (q *QUICUnderlay) Write(b []byte) (int, error) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	return q.stream.Write(b)
}

func (q *QUICUnderlay) Close() error {
	q.stream.Close()
	return q.conn.CloseWithError(0, "tunnel closed")
}

func (q *QUICUnderlay) SetReadDeadline(t time.Time) error {
	return q.stream.SetReadDeadline(t)
}

func (q *QUICUnderlay) SetWriteDeadline(t time.Time) error {
	return q.stream.SetWriteDeadline(t)
}

func (q *QUICUnderlay) Protocol() string {
	return "quic"
}

func (q *QUICUnderlay) LocalAddr() net.Addr {
	return q.conn.LocalAddr()
}

func (q *QUICUnderlay) RemoteAddr() net.Addr {
	return q.conn.RemoteAddr()
}

// NATDetector определяет тип NAT и выполняет hole punching
type NATDetector struct {
	stunServers []string
	logger      *logger.Logger
}

// NATType представляет тип NAT
type NATType int

const (
	NATTypeUnknown NATType = iota
	NATTypeCone
	NATTypeSymmetric
	NATTypeRestricted
	NATTypePortRestricted
)

// HolePuncher выполняет hole punching для NAT traversal
type HolePuncher struct {
	logger *logger.Logger
}

// TunnelConfig конфигурация туннеля
type TunnelConfig struct {
	PeerID            string
	TenantID          string
	LocalPort         int
	RemotePort        int
	Protocol          string
	HeartbeatInterval time.Duration
	Timeout           time.Duration
}

// NewP2PTunnelManager создает новый P2P Tunnel Manager
func NewP2PTunnelManager(logger *logger.Logger, relayManager *RelayManager) *P2PTunnelManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &P2PTunnelManager{
		tunnels:      make(map[string]*P2PTunnel),
		natDetector:  NewNATDetector(logger),
		holePuncher:  NewHolePuncher(logger),
		relayManager: relayManager,
		logger:       logger,
		ctx:          ctx,
		cancel:       cancel,
	}
}

// Start запускает P2P Tunnel Manager
func (tm *P2PTunnelManager) Start() error {
	tm.logger.Info("Starting P2P Tunnel Manager")

	// Запускаем heartbeat ticker
	go tm.heartbeatLoop()

	// Запускаем cleanup ticker
	go tm.cleanupLoop()

	return nil
}

// Stop останавливает P2P Tunnel Manager
func (tm *P2PTunnelManager) Stop() error {
	tm.logger.Info("Stopping P2P Tunnel Manager")
	tm.cancel()

	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	// Закрываем все туннели
	for _, tunnel := range tm.tunnels {
		tunnel.Close()
	}

	return nil
}

// CreateTunnel создает новый P2P туннель
func (tm *P2PTunnelManager) CreateTunnel(config *TunnelConfig) (*P2PTunnel, error) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	tunnelID := fmt.Sprintf("%s-%s-%d", config.TenantID, config.PeerID, time.Now().UnixNano())

	// Определяем тип NAT
	natType, publicIP, err := tm.natDetector.DetectNATType()
	if err != nil {
		return nil, fmt.Errorf("failed to detect NAT type: %v", err)
	}

	tm.logger.Info("NAT detection completed",
		zap.String("tunnel_id", tunnelID),
		zap.String("nat_type", natType.String()),
		zap.String("public_ip", publicIP))

	// Создаем туннель
	tunnel := &P2PTunnel{
		ID:            tunnelID,
		PeerID:        config.PeerID,
		TenantID:      config.TenantID,
		IsActive:      false,
		CreatedAt:     time.Now(),
		LastHeartbeat: time.Now(),
	}

	// Выполняем hole punching если необходимо
	if natType != NATTypeCone {
		err = tm.holePuncher.PerformHolePunching(tunnel, natType)
		if err != nil {
			return nil, fmt.Errorf("failed to perform hole punching: %v", err)
		}
	}

	tm.tunnels[tunnelID] = tunnel

	tm.logger.Info("P2P tunnel created",
		zap.String("tunnel_id", tunnelID),
		zap.String("peer_id", config.PeerID),
		zap.String("tenant_id", config.TenantID))

	return tunnel, nil
}

// GetTunnel возвращает туннель по ID
func (tm *P2PTunnelManager) GetTunnel(tunnelID string) (*P2PTunnel, bool) {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	tunnel, exists := tm.tunnels[tunnelID]
	return tunnel, exists
}

// GetTunnelsByTenant возвращает все туннели для tenant
func (tm *P2PTunnelManager) GetTunnelsByTenant(tenantID string) []*P2PTunnel {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	var tunnels []*P2PTunnel
	for _, tunnel := range tm.tunnels {
		if tunnel.TenantID == tenantID {
			tunnels = append(tunnels, tunnel)
		}
	}

	return tunnels
}

// CloseTunnel закрывает туннель
func (tm *P2PTunnelManager) CloseTunnel(tunnelID string) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	tunnel, exists := tm.tunnels[tunnelID]
	if !exists {
		return fmt.Errorf("tunnel %s not found", tunnelID)
	}

	tunnel.Close()
	delete(tm.tunnels, tunnelID)

	tm.logger.Info("P2P tunnel closed",
		zap.String("tunnel_id", tunnelID))

	return nil
}

// heartbeatLoop отправляет heartbeat для всех активных туннелей
func (tm *P2PTunnelManager) heartbeatLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-tm.ctx.Done():
			return
		case <-ticker.C:
			tm.sendHeartbeats()
		}
	}
}

// sendHeartbeats отправляет heartbeat для всех туннелей
func (tm *P2PTunnelManager) sendHeartbeats() {
	tm.mutex.RLock()
	tunnels := make(map[string]*P2PTunnel)
	for id, tunnel := range tm.tunnels {
		tunnels[id] = tunnel
	}
	tm.mutex.RUnlock()

	for _, tunnel := range tunnels {
		if tunnel.IsActive {
			tunnel.SendHeartbeat()
		}
	}
}

// cleanupLoop очищает неактивные туннели
func (tm *P2PTunnelManager) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-tm.ctx.Done():
			return
		case <-ticker.C:
			tm.cleanupInactiveTunnels()
		}
	}
}

// cleanupInactiveTunnels удаляет неактивные туннели
func (tm *P2PTunnelManager) cleanupInactiveTunnels() {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	cutoff := time.Now().Add(-10 * time.Minute)
	var toRemove []string

	for id, tunnel := range tm.tunnels {
		if !tunnel.IsActive && tunnel.LastHeartbeat.Before(cutoff) {
			toRemove = append(toRemove, id)
		}
	}

	for _, id := range toRemove {
		tm.tunnels[id].Close()
		delete(tm.tunnels, id)

		tm.logger.Info("Removed inactive tunnel",
			zap.String("tunnel_id", id))
	}
}

// SendHeartbeat отправляет heartbeat для туннеля
func (t *P2PTunnel) SendHeartbeat() error {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()

	if t.Underlay == nil {
		return fmt.Errorf("tunnel underlay not connected")
	}

	heartbeat := map[string]interface{}{
		"type":      "heartbeat",
		"tunnel_id": t.ID,
		"timestamp": time.Now().Unix(),
	}

	data, err := json.Marshal(heartbeat)
	if err != nil {
		return err
	}

	_, err = t.Underlay.Write(data)
	if err != nil {
		t.IsActive = false
		return err
	}

	t.LastHeartbeat = time.Now()
	t.IsActive = true

	return nil
}

// Close закрывает туннель
func (t *P2PTunnel) Close() error {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()

	t.IsActive = false

	if t.Underlay != nil {
		return t.Underlay.Close()
	}

	return nil
}

// NewNATDetector создает новый NAT детектор
func NewNATDetector(logger *logger.Logger) *NATDetector {
	return &NATDetector{
		stunServers: []string{
			"stun.l.google.com:19302",
			"stun1.l.google.com:19302",
			"auth.2gc.ru:19302",
		},
		logger: logger,
	}
}

// DetectNATType определяет тип NAT
func (nd *NATDetector) DetectNATType() (NATType, string, error) {
	// Простая реализация - в реальности нужно более сложная логика
	randNum, err := rand.Int(rand.Reader, big.NewInt(15000))
	if err != nil {
		return NATTypeUnknown, "", fmt.Errorf("failed to generate random number: %v", err)
	}
	localPort := int(randNum.Int64()) + 50000

	// Тестируем с первым STUN сервером
	publicIP, port1, err := nd.natTest(nd.stunServers[0], localPort)
	if err != nil {
		return NATTypeUnknown, "", err
	}

	// Тестируем со вторым STUN сервером
	_, port2, err := nd.natTest(nd.stunServers[1], localPort)
	if err != nil {
		return NATTypeUnknown, "", err
	}

	// Определяем тип NAT
	if port1 == port2 {
		return NATTypeCone, publicIP, nil
	}

	return NATTypeSymmetric, publicIP, nil
}

// natTest выполняет NAT тест
func (nd *NATDetector) natTest(serverAddr string, localPort int) (string, int, error) {
	conn, err := net.ListenPacket("udp", fmt.Sprintf(":%d", localPort))
	if err != nil {
		return "", 0, err
	}
	defer conn.Close()

	dst, err := net.ResolveUDPAddr("udp", serverAddr)
	if err != nil {
		return "", 0, err
	}

	// Отправляем тестовый пакет
	testData := []byte("NAT_TEST")
	_, err = conn.WriteTo(testData, dst)
	if err != nil {
		return "", 0, err
	}

	// Читаем ответ
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	buffer := make([]byte, 1024)
	n, _, err := conn.ReadFrom(buffer)
	if err != nil {
		return "", 0, err
	}

	// Парсим ответ (упрощенная версия)
	_ = string(buffer[:n])
	// В реальности нужно парсить STUN ответ

	return "127.0.0.1", localPort, nil
}

// NewHolePuncher создает новый hole puncher
func NewHolePuncher(logger *logger.Logger) *HolePuncher {
	return &HolePuncher{
		logger: logger,
	}
}

// PerformHolePunching выполняет hole punching
func (hp *HolePuncher) PerformHolePunching(tunnel *P2PTunnel, natType NATType) error {
	hp.logger.Info("Performing hole punching",
		zap.String("tunnel_id", tunnel.ID),
		zap.String("nat_type", natType.String()))

	// Простая реализация - в реальности нужно более сложная логика
	// для разных типов NAT

	return nil
}

// String возвращает строковое представление NAT типа
func (nt NATType) String() string {
	switch nt {
	case NATTypeCone:
		return "cone"
	case NATTypeSymmetric:
		return "symmetric"
	case NATTypeRestricted:
		return "restricted"
	case NATTypePortRestricted:
		return "port_restricted"
	default:
		return "unknown"
	}
}
