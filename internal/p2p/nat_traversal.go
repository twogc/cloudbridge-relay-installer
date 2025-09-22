package p2p

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
)

// NATTraversalManager управляет NAT traversal и hole punching
type NATTraversalManager struct {
	stunServers   []string
	holePunchers  map[string]*HolePuncher
	logger        *logger.Logger
	mutex         sync.RWMutex
	activePunches map[string]*PunchSession
	punchMutex    sync.RWMutex
}

// PunchSession представляет сессию hole punching
type PunchSession struct {
	ID             string
	LocalAddr      *net.UDPAddr
	RemoteAddr     *net.UDPAddr
	StartTime      time.Time
	LastActivity   time.Time
	Attempts       int
	MaxAttempts    int
	IsSuccessful   bool
	ConnectionType string // "udp", "tcp"
	Mutex          sync.RWMutex
}

// STUNMessage представляет STUN сообщение
type STUNMessage struct {
	Type       uint16
	Length     uint16
	Magic      uint32
	ID         [12]byte
	Attributes []STUNAttribute
}

// STUNAttribute представляет STUN атрибут
type STUNAttribute struct {
	Type   uint16
	Length uint16
	Value  []byte
}

// STUNResponse представляет ответ STUN сервера
type STUNResponse struct {
	PublicIP   string
	PublicPort int
	LocalIP    string
	LocalPort  int
	NatType    NATType
}

// HolePunchingConfig конфигурация hole punching
type HolePunchingConfig struct {
	MaxAttempts     int
	AttemptInterval time.Duration
	Timeout         time.Duration
	RetryCount      int
}

// NewNATTraversalManager создает новый NAT Traversal Manager
func NewNATTraversalManager(logger *logger.Logger) *NATTraversalManager {
	return &NATTraversalManager{
		stunServers: []string{
			"stun.l.google.com:19302",
			"stun1.l.google.com:19302",
			"stun2.l.google.com:19302",
			"stun3.l.google.com:19302",
			"stun4.l.google.com:19302",
			"auth.2gc.ru:19302",
		},
		holePunchers:  make(map[string]*HolePuncher),
		logger:        logger,
		activePunches: make(map[string]*PunchSession),
	}
}

// DetectNATType определяет тип NAT
func (ntm *NATTraversalManager) DetectNATType() (NATType, *STUNResponse, error) {
	ntm.logger.Info("Starting NAT type detection")

	// Выбираем случайный локальный порт
	randNum, err := rand.Int(rand.Reader, big.NewInt(15000))
	if err != nil {
		return NATTypeUnknown, nil, fmt.Errorf("failed to generate random number: %v", err)
	}
	localPort := int(randNum.Int64()) + 50000

	// Тестируем с первым STUN сервером
	response1, err := ntm.performSTUNTest(ntm.stunServers[0], localPort)
	if err != nil {
		return NATTypeUnknown, nil, fmt.Errorf("STUN test 1 failed: %v", err)
	}

	// Тестируем со вторым STUN сервером
	response2, err := ntm.performSTUNTest(ntm.stunServers[1], localPort)
	if err != nil {
		return NATTypeUnknown, nil, fmt.Errorf("STUN test 2 failed: %v", err)
	}

	// Определяем тип NAT
	natType := ntm.determineNATType(response1, response2)

	ntm.logger.Info("NAT type detection completed",
		zap.String("nat_type", natType.String()),
		zap.String("public_ip", response1.PublicIP),
		zap.Int("public_port_1", response1.PublicPort),
		zap.Int("public_port_2", response2.PublicPort))

	return natType, response1, nil
}

// performSTUNTest выполняет STUN тест
func (ntm *NATTraversalManager) performSTUNTest(serverAddr string, localPort int) (*STUNResponse, error) {
	ntm.logger.Debug("Performing STUN test",
		zap.String("server", serverAddr),
		zap.Int("local_port", localPort))

	// Создаем UDP соединение
	conn, err := net.ListenPacket("udp", fmt.Sprintf(":%d", localPort))
	if err != nil {
		return nil, fmt.Errorf("failed to listen on port %d: %v", localPort, err)
	}
	defer conn.Close()

	// Разрешаем адрес сервера
	dst, err := net.ResolveUDPAddr("udp", serverAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve server address: %v", err)
	}

	// Создаем STUN Binding Request
	stunRequest := ntm.createSTUNBindingRequest()

	// Отправляем запрос
	_, err = conn.WriteTo(stunRequest, dst)
	if err != nil {
		return nil, fmt.Errorf("failed to send STUN request: %v", err)
	}

	// Устанавливаем timeout для чтения
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	// Читаем ответ
	buffer := make([]byte, 1024)
	n, _, err := conn.ReadFrom(buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to read STUN response: %v", err)
	}

	// Парсим ответ
	response, err := ntm.parseSTUNResponse(buffer[:n])
	if err != nil {
		return nil, fmt.Errorf("failed to parse STUN response: %v", err)
	}

	// Добавляем локальную информацию
	response.LocalIP = "127.0.0.1" // Упрощение
	response.LocalPort = localPort

	return response, nil
}

// createSTUNBindingRequest создает STUN Binding Request
func (ntm *NATTraversalManager) createSTUNBindingRequest() []byte {
	// Упрощенная реализация STUN Binding Request
	// В реальности нужно создавать правильный STUN пакет

	request := make([]byte, 20) // Минимальный размер STUN сообщения

	// STUN Header
	request[0] = 0x00 // Message Type (Binding Request)
	request[1] = 0x01
	request[2] = 0x00 // Message Length
	request[3] = 0x00
	request[4] = 0x21 // Magic Cookie
	request[5] = 0x12
	request[6] = 0xA4
	request[7] = 0x42

	// Transaction ID (случайные байты)
	for i := 8; i < 20; i++ {
		randByte := make([]byte, 1)
		if _, err := rand.Read(randByte); err != nil {
			return nil
		}
		request[i] = randByte[0]
	}

	return request
}

// parseSTUNResponse парсит STUN ответ
func (ntm *NATTraversalManager) parseSTUNResponse(data []byte) (*STUNResponse, error) {
	if len(data) < 20 {
		return nil, fmt.Errorf("STUN response too short")
	}

	response := &STUNResponse{}

	// Парсим заголовок
	messageType := uint16(data[0])<<8 | uint16(data[1])
	messageLength := uint16(data[2])<<8 | uint16(data[3])

	if messageType != 0x0101 { // Binding Response
		return nil, fmt.Errorf("unexpected STUN message type: %d", messageType)
	}

	// Парсим атрибуты
	offset := 20
	for offset < int(messageLength)+20 {
		if offset+4 > len(data) {
			break
		}

		attrType := uint16(data[offset])<<8 | uint16(data[offset+1])
		attrLength := uint16(data[offset+2])<<8 | uint16(data[offset+3])

		offset += 4

		if offset+int(attrLength) > len(data) {
			break
		}

		// Обрабатываем MAPPED-ADDRESS атрибут
		if attrType == 0x0001 { // MAPPED-ADDRESS
			if attrLength >= 8 {
				family := data[offset+1]
				if family == 0x01 { // IPv4
					port := uint16(data[offset+2])<<8 | uint16(data[offset+3])
					ip := net.IP(data[offset+4 : offset+8])
					response.PublicIP = ip.String()
					response.PublicPort = int(port)
				}
			}
		}

		offset += int(attrLength)
		// Выравнивание на 4 байта
		if attrLength%4 != 0 {
			offset += 4 - int(attrLength%4)
		}
	}

	return response, nil
}

// determineNATType определяет тип NAT на основе ответов STUN
func (ntm *NATTraversalManager) determineNATType(response1, response2 *STUNResponse) NATType {
	// Если порты одинаковые - это Cone NAT
	if response1.PublicPort == response2.PublicPort {
		return NATTypeCone
	}

	// Если порты разные - это Symmetric NAT
	return NATTypeSymmetric
}

// StartHolePunching начинает hole punching
func (ntm *NATTraversalManager) StartHolePunching(sessionID string, localAddr, remoteAddr *net.UDPAddr, config *HolePunchingConfig) (*PunchSession, error) {
	ntm.logger.Info("Starting hole punching",
		zap.String("session_id", sessionID),
		zap.String("local_addr", localAddr.String()),
		zap.String("remote_addr", remoteAddr.String()))

	// Создаем сессию hole punching
	session := &PunchSession{
		ID:             sessionID,
		LocalAddr:      localAddr,
		RemoteAddr:     remoteAddr,
		StartTime:      time.Now(),
		LastActivity:   time.Now(),
		MaxAttempts:    config.MaxAttempts,
		ConnectionType: "udp",
	}

	// Регистрируем сессию
	ntm.punchMutex.Lock()
	ntm.activePunches[sessionID] = session
	ntm.punchMutex.Unlock()

	// Запускаем hole punching
	go ntm.performHolePunching(session, config)

	return session, nil
}

// performHolePunching выполняет hole punching
func (ntm *NATTraversalManager) performHolePunching(session *PunchSession, config *HolePunchingConfig) {
	defer func() {
		ntm.punchMutex.Lock()
		delete(ntm.activePunches, session.ID)
		ntm.punchMutex.Unlock()
	}()

	// Создаем UDP соединение
	conn, err := net.ListenPacket("udp", session.LocalAddr.String())
	if err != nil {
		ntm.logger.Error("Failed to create UDP connection for hole punching",
			zap.String("session_id", session.ID),
			zap.Error(err))
		return
	}
	defer conn.Close()

	// Выполняем hole punching
	for session.Attempts < session.MaxAttempts {
		session.Mutex.Lock()
		session.Attempts++
		session.LastActivity = time.Now()
		session.Mutex.Unlock()

		ntm.logger.Debug("Hole punching attempt",
			zap.String("session_id", session.ID),
			zap.Int("attempt", session.Attempts),
			zap.Int("max_attempts", session.MaxAttempts))

		// Отправляем пакет
		punchData := []byte(fmt.Sprintf("PUNCH_%s_%d", session.ID, session.Attempts))
		_, err := conn.WriteTo(punchData, session.RemoteAddr)
		if err != nil {
			ntm.logger.Error("Failed to send punch packet",
				zap.String("session_id", session.ID),
				zap.Error(err))
		}

		// Пытаемся получить ответ
		conn.SetReadDeadline(time.Now().Add(config.Timeout))
		buffer := make([]byte, 1024)
		_, addr, err := conn.ReadFrom(buffer)
		if err == nil {
			// Получили ответ - hole punching успешен
			session.Mutex.Lock()
			session.IsSuccessful = true
			session.Mutex.Unlock()

			ntm.logger.Info("Hole punching successful",
				zap.String("session_id", session.ID),
				zap.String("remote_addr", addr.String()),
				zap.Int("attempts", session.Attempts))

			return
		}

		// Ждем перед следующей попыткой
		time.Sleep(config.AttemptInterval)
	}

	ntm.logger.Warn("Hole punching failed",
		zap.String("session_id", session.ID),
		zap.Int("attempts", session.Attempts))
}

// GetPunchSession возвращает сессию hole punching
func (ntm *NATTraversalManager) GetPunchSession(sessionID string) (*PunchSession, bool) {
	ntm.punchMutex.RLock()
	defer ntm.punchMutex.RUnlock()

	session, exists := ntm.activePunches[sessionID]
	return session, exists
}

// StopHolePunching останавливает hole punching
func (ntm *NATTraversalManager) StopHolePunching(sessionID string) error {
	ntm.punchMutex.Lock()
	defer ntm.punchMutex.Unlock()

	session, exists := ntm.activePunches[sessionID]
	if !exists {
		return fmt.Errorf("punch session %s not found", sessionID)
	}

	session.Mutex.Lock()
	session.IsSuccessful = false
	session.Mutex.Unlock()

	delete(ntm.activePunches, sessionID)

	ntm.logger.Info("Hole punching stopped",
		zap.String("session_id", sessionID))

	return nil
}

// GetActivePunchSessions возвращает все активные сессии hole punching
func (ntm *NATTraversalManager) GetActivePunchSessions() []*PunchSession {
	ntm.punchMutex.RLock()
	defer ntm.punchMutex.RUnlock()

	var sessions []*PunchSession
	for _, session := range ntm.activePunches {
		sessions = append(sessions, session)
	}

	return sessions
}

// GetStats возвращает статистику NAT traversal
func (ntm *NATTraversalManager) GetStats() map[string]interface{} {
	ntm.punchMutex.RLock()
	activeSessions := len(ntm.activePunches)
	ntm.punchMutex.RUnlock()

	return map[string]interface{}{
		"stun_servers":    ntm.stunServers,
		"active_sessions": activeSessions,
		"hole_punchers":   len(ntm.holePunchers),
	}
}

// DefaultHolePunchingConfig возвращает конфигурацию по умолчанию
func DefaultHolePunchingConfig() *HolePunchingConfig {
	return &HolePunchingConfig{
		MaxAttempts:     10,
		AttemptInterval: 1 * time.Second,
		Timeout:         5 * time.Second,
		RetryCount:      3,
	}
}
