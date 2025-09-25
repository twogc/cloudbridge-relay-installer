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

// ICEConnectivityChecker выполняет ICE connectivity checks
type ICEConnectivityChecker struct {
	logger *logger.Logger

	// STUN серверы для проверок
	stunServers []string

	// Активные проверки
	activeChecks      map[string]*ConnectivityCheck
	activeChecksMutex sync.RWMutex

	// Результаты проверок
	checkResults      map[string]*ConnectivityResult
	checkResultsMutex sync.RWMutex

	// Конфигурация
	checkTimeout    time.Duration
	checkInterval   time.Duration
	maxConcurrentChecks int

	ctx    context.Context
	cancel context.CancelFunc
}

// ConnectivityCheck представляет активную проверку connectivity
type ConnectivityCheck struct {
	CheckID      string
	ClientID     string
	LocalAddr    string
	RemoteAddr   string
	CheckType    CheckType
	StartedAt    time.Time
	TimeoutAt    time.Time
	IsActive     bool
}

// ConnectivityResult представляет результат проверки connectivity
type ConnectivityResult struct {
	CheckID        string
	ClientID       string
	LocalAddr      string
	RemoteAddr     string
	CheckType      CheckType
	Success        bool
	Latency        time.Duration
	Error          string
	CompletedAt    time.Time
	PublicIP       string
	NATBehavior    string
}

// CheckType тип проверки connectivity
type CheckType string

const (
	CheckTypeSTUN        CheckType = "stun"
	CheckTypeTURN        CheckType = "turn"
	CheckTypePeerToPeer  CheckType = "peer_to_peer"
	CheckTypeRelay       CheckType = "relay"
)

// NewICEConnectivityChecker создает новый ICE connectivity checker
func NewICEConnectivityChecker(logger *logger.Logger, stunServers []string) *ICEConnectivityChecker {
	ctx, cancel := context.WithCancel(context.Background())
	icc := &ICEConnectivityChecker{
		logger:              logger,
		stunServers:         stunServers,
		activeChecks:        make(map[string]*ConnectivityCheck),
		checkResults:        make(map[string]*ConnectivityResult),
		checkTimeout:        10 * time.Second,
		checkInterval:       30 * time.Second,
		maxConcurrentChecks: 100,
		ctx:                 ctx,
		cancel:              cancel,
	}
	go icc.startCheckCleanup()
	return icc
}

// StartSTUNCheck запускает STUN проверку
func (icc *ICEConnectivityChecker) StartSTUNCheck(clientID, localAddr string) (string, error) {
	checkID := fmt.Sprintf("stun-check-%d", time.Now().UnixNano())
	
	// Выбираем лучший STUN сервер
	stunServer := icc.selectBestSTUNServer()
	if stunServer == "" {
		return "", fmt.Errorf("no STUN servers available")
	}

	check := &ConnectivityCheck{
		CheckID:    checkID,
		ClientID:   clientID,
		LocalAddr:  localAddr,
		RemoteAddr: stunServer,
		CheckType:  CheckTypeSTUN,
		StartedAt:  time.Now(),
		TimeoutAt:  time.Now().Add(icc.checkTimeout),
		IsActive:   true,
	}

	icc.activeChecksMutex.Lock()
	icc.activeChecks[checkID] = check
	icc.activeChecksMutex.Unlock()

	// Запускаем проверку в отдельной горутине
	go icc.performSTUNCheck(check)

	icc.logger.Info("Started STUN connectivity check",
		zap.String("check_id", checkID),
		zap.String("client_id", clientID),
		zap.String("local_addr", localAddr),
		zap.String("stun_server", stunServer),
	)

	return checkID, nil
}

// StartTURNCheck запускает TURN проверку
func (icc *ICEConnectivityChecker) StartTURNCheck(clientID, localAddr, turnServer string) (string, error) {
	checkID := fmt.Sprintf("turn-check-%d", time.Now().UnixNano())

	check := &ConnectivityCheck{
		CheckID:    checkID,
		ClientID:   clientID,
		LocalAddr:  localAddr,
		RemoteAddr: turnServer,
		CheckType:  CheckTypeTURN,
		StartedAt:  time.Now(),
		TimeoutAt:  time.Now().Add(icc.checkTimeout),
		IsActive:   true,
	}

	icc.activeChecksMutex.Lock()
	icc.activeChecks[checkID] = check
	icc.activeChecksMutex.Unlock()

	// Запускаем проверку в отдельной горутине
	go icc.performTURNCheck(check)

	icc.logger.Info("Started TURN connectivity check",
		zap.String("check_id", checkID),
		zap.String("client_id", clientID),
		zap.String("local_addr", localAddr),
		zap.String("turn_server", turnServer),
	)

	return checkID, nil
}

// StartPeerToPeerCheck запускает peer-to-peer проверку
func (icc *ICEConnectivityChecker) StartPeerToPeerCheck(clientID, localAddr, remoteAddr string) (string, error) {
	checkID := fmt.Sprintf("p2p-check-%d", time.Now().UnixNano())

	check := &ConnectivityCheck{
		CheckID:    checkID,
		ClientID:   clientID,
		LocalAddr:  localAddr,
		RemoteAddr: remoteAddr,
		CheckType:  CheckTypePeerToPeer,
		StartedAt:  time.Now(),
		TimeoutAt:  time.Now().Add(icc.checkTimeout),
		IsActive:   true,
	}

	icc.activeChecksMutex.Lock()
	icc.activeChecks[checkID] = check
	icc.activeChecksMutex.Unlock()

	// Запускаем проверку в отдельной горутине
	go icc.performPeerToPeerCheck(check)

	icc.logger.Info("Started peer-to-peer connectivity check",
		zap.String("check_id", checkID),
		zap.String("client_id", clientID),
		zap.String("local_addr", localAddr),
		zap.String("remote_addr", remoteAddr),
	)

	return checkID, nil
}

// performSTUNCheck выполняет STUN проверку
func (icc *ICEConnectivityChecker) performSTUNCheck(check *ConnectivityCheck) {
	defer icc.completeCheck(check.CheckID)

	start := time.Now()
	
	// В реальной реализации здесь должна быть логика STUN проверки
	// Пока просто проверяем доступность порта
	conn, err := net.DialTimeout("udp", check.RemoteAddr, icc.checkTimeout)
	latency := time.Since(start)
	
	result := &ConnectivityResult{
		CheckID:     check.CheckID,
		ClientID:    check.ClientID,
		LocalAddr:   check.LocalAddr,
		RemoteAddr:  check.RemoteAddr,
		CheckType:   check.CheckType,
		CompletedAt: time.Now(),
		Latency:     latency,
	}

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		icc.logger.Debug("STUN check failed",
			zap.String("check_id", check.CheckID),
			zap.Error(err),
		)
	} else {
		result.Success = true
		conn.Close()
		icc.logger.Debug("STUN check succeeded",
			zap.String("check_id", check.CheckID),
			zap.Duration("latency", latency),
		)
	}

	icc.recordCheckResult(result)
}

// performTURNCheck выполняет TURN проверку
func (icc *ICEConnectivityChecker) performTURNCheck(check *ConnectivityCheck) {
	defer icc.completeCheck(check.CheckID)

	start := time.Now()
	
	// В реальной реализации здесь должна быть логика TURN проверки
	// Пока просто проверяем доступность порта
	conn, err := net.DialTimeout("tcp", check.RemoteAddr, icc.checkTimeout)
	latency := time.Since(start)
	
	result := &ConnectivityResult{
		CheckID:     check.CheckID,
		ClientID:    check.ClientID,
		LocalAddr:   check.LocalAddr,
		RemoteAddr:  check.RemoteAddr,
		CheckType:   check.CheckType,
		CompletedAt: time.Now(),
		Latency:     latency,
	}

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		icc.logger.Debug("TURN check failed",
			zap.String("check_id", check.CheckID),
			zap.Error(err),
		)
	} else {
		result.Success = true
		conn.Close()
		icc.logger.Debug("TURN check succeeded",
			zap.String("check_id", check.CheckID),
			zap.Duration("latency", latency),
		)
	}

	icc.recordCheckResult(result)
}

// performPeerToPeerCheck выполняет peer-to-peer проверку
func (icc *ICEConnectivityChecker) performPeerToPeerCheck(check *ConnectivityCheck) {
	defer icc.completeCheck(check.CheckID)

	start := time.Now()
	
	// В реальной реализации здесь должна быть логика peer-to-peer проверки
	// Пока просто проверяем доступность порта
	conn, err := net.DialTimeout("udp", check.RemoteAddr, icc.checkTimeout)
	latency := time.Since(start)
	
	result := &ConnectivityResult{
		CheckID:     check.CheckID,
		ClientID:    check.ClientID,
		LocalAddr:   check.LocalAddr,
		RemoteAddr:  check.RemoteAddr,
		CheckType:   check.CheckType,
		CompletedAt: time.Now(),
		Latency:     latency,
	}

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		icc.logger.Debug("Peer-to-peer check failed",
			zap.String("check_id", check.CheckID),
			zap.Error(err),
		)
	} else {
		result.Success = true
		conn.Close()
		icc.logger.Debug("Peer-to-peer check succeeded",
			zap.String("check_id", check.CheckID),
			zap.Duration("latency", latency),
		)
	}

	icc.recordCheckResult(result)
}

// completeCheck завершает проверку
func (icc *ICEConnectivityChecker) completeCheck(checkID string) {
	icc.activeChecksMutex.Lock()
	if check, exists := icc.activeChecks[checkID]; exists {
		check.IsActive = false
		delete(icc.activeChecks, checkID)
	}
	icc.activeChecksMutex.Unlock()
}

// recordCheckResult записывает результат проверки
func (icc *ICEConnectivityChecker) recordCheckResult(result *ConnectivityResult) {
	icc.checkResultsMutex.Lock()
	icc.checkResults[result.CheckID] = result
	icc.checkResultsMutex.Unlock()

	icc.logger.Info("Connectivity check completed",
		zap.String("check_id", result.CheckID),
		zap.String("client_id", result.ClientID),
		zap.String("check_type", string(result.CheckType)),
		zap.Bool("success", result.Success),
		zap.Duration("latency", result.Latency),
		zap.String("error", result.Error),
	)
}

// selectBestSTUNServer выбирает лучший STUN сервер
func (icc *ICEConnectivityChecker) selectBestSTUNServer() string {
	if len(icc.stunServers) == 0 {
		return ""
	}
	
	// В реальной реализации здесь должна быть логика выбора лучшего сервера
	// Пока возвращаем первый доступный
	return icc.stunServers[0]
}

// GetCheckResult возвращает результат проверки
func (icc *ICEConnectivityChecker) GetCheckResult(checkID string) (*ConnectivityResult, bool) {
	icc.checkResultsMutex.RLock()
	defer icc.checkResultsMutex.RUnlock()

	result, exists := icc.checkResults[checkID]
	return result, exists
}

// GetClientCheckResults возвращает результаты проверок для клиента
func (icc *ICEConnectivityChecker) GetClientCheckResults(clientID string) []*ConnectivityResult {
	icc.checkResultsMutex.RLock()
	defer icc.checkResultsMutex.RUnlock()

	results := make([]*ConnectivityResult, 0)
	for _, result := range icc.checkResults {
		if result.ClientID == clientID {
			results = append(results, result)
		}
	}

	return results
}

// GetActiveChecks возвращает активные проверки
func (icc *ICEConnectivityChecker) GetActiveChecks() []*ConnectivityCheck {
	icc.activeChecksMutex.RLock()
	defer icc.activeChecksMutex.RUnlock()

	checks := make([]*ConnectivityCheck, 0, len(icc.activeChecks))
	for _, check := range icc.activeChecks {
		checks = append(checks, check)
	}

	return checks
}

// GetConnectivityStats возвращает статистику connectivity checks
func (icc *ICEConnectivityChecker) GetConnectivityStats() map[string]interface{} {
	icc.activeChecksMutex.RLock()
	icc.checkResultsMutex.RLock()
	defer icc.activeChecksMutex.RUnlock()
	defer icc.checkResultsMutex.RUnlock()

	activeChecks := len(icc.activeChecks)
	totalResults := len(icc.checkResults)
	successfulChecks := 0
	failedChecks := 0

	for _, result := range icc.checkResults {
		if result.Success {
			successfulChecks++
		} else {
			failedChecks++
		}
	}

	return map[string]interface{}{
		"active_checks":     activeChecks,
		"total_results":     totalResults,
		"successful_checks": successfulChecks,
		"failed_checks":     failedChecks,
		"success_rate":      float64(successfulChecks) / float64(totalResults) * 100,
	}
}

// startCheckCleanup запускает очистку просроченных проверок
func (icc *ICEConnectivityChecker) startCheckCleanup() {
	ticker := time.NewTicker(icc.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-icc.ctx.Done():
			return
		case <-ticker.C:
			icc.cleanupExpiredChecks()
			icc.cleanupOldResults()
		}
	}
}

// cleanupExpiredChecks очищает просроченные проверки
func (icc *ICEConnectivityChecker) cleanupExpiredChecks() {
	icc.activeChecksMutex.Lock()
	defer icc.activeChecksMutex.Unlock()

	now := time.Now()
	for checkID, check := range icc.activeChecks {
		if now.After(check.TimeoutAt) {
			check.IsActive = false
			delete(icc.activeChecks, checkID)
			icc.logger.Debug("Cleaned up expired connectivity check", zap.String("check_id", checkID))
		}
	}
}

// cleanupOldResults очищает старые результаты
func (icc *ICEConnectivityChecker) cleanupOldResults() {
	icc.checkResultsMutex.Lock()
	defer icc.checkResultsMutex.Unlock()

	expiredTime := time.Now().Add(-1 * time.Hour) // Удаляем результаты старше 1 часа
	for checkID, result := range icc.checkResults {
		if result.CompletedAt.Before(expiredTime) {
			delete(icc.checkResults, checkID)
			icc.logger.Debug("Cleaned up old connectivity result", zap.String("check_id", checkID))
		}
	}
}

// Stop останавливает ICE connectivity checker
func (icc *ICEConnectivityChecker) Stop() {
	icc.cancel()
	icc.logger.Info("ICE connectivity checker stopped")
}
