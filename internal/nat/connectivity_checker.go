package nat

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/pion/stun"
	"go.uber.org/zap"

	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
)

// ConnectivityChecker выполняет ICE connectivity checks
type ConnectivityChecker struct {
	logger  *logger.Logger
	metrics *NATMetrics
	
	// Configuration
	checkInterval    time.Duration
	checkTimeout     time.Duration
	maxConcurrentChecks int
	
	// State
	activePairs   map[string]*CandidatePair
	checkedPairs  map[string]*CheckResult
	selectedPair  *CandidatePair
	mu            sync.RWMutex
	
	// Context for cancellation
	ctx    context.Context
	cancel context.CancelFunc
}

// CandidatePair представляет пару кандидатов для connectivity check
type CandidatePair struct {
	ID            string
	Local         *ICECandidate
	Remote        *ICECandidate
	Priority      uint64
	State         PairState
	Foundation    string
	CreatedAt     time.Time
	LastChecked   time.Time
}

// PairState состояние пары кандидатов
type PairState int

const (
	PairStateWaiting PairState = iota
	PairStateInProgress
	PairStateSucceeded
	PairStateFailed
	PairStateFrozen
)

func (ps PairState) String() string {
	switch ps {
	case PairStateWaiting:
		return "Waiting"
	case PairStateInProgress:
		return "InProgress"
	case PairStateSucceeded:
		return "Succeeded"
	case PairStateFailed:
		return "Failed"
	case PairStateFrozen:
		return "Frozen"
	default:
		return "Unknown"
	}
}

// CheckResult результат connectivity check
type CheckResult struct {
	PairID      string
	Success     bool
	RTT         time.Duration
	Error       error
	Timestamp   time.Time
	RetryCount  int
}

// NewConnectivityChecker создает новый connectivity checker
func NewConnectivityChecker(logger *logger.Logger, metrics *NATMetrics) *ConnectivityChecker {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &ConnectivityChecker{
		logger:              logger,
		metrics:             metrics,
		checkInterval:       200 * time.Millisecond,
		checkTimeout:        5 * time.Second,
		maxConcurrentChecks: 5,
		activePairs:         make(map[string]*CandidatePair),
		checkedPairs:        make(map[string]*CheckResult),
		ctx:                 ctx,
		cancel:              cancel,
	}
}

// StartConnectivityChecks запускает процесс connectivity checks
func (cc *ConnectivityChecker) StartConnectivityChecks(localCandidates, remoteCandidates []*ICECandidate) error {
	cc.logger.Info("Starting ICE connectivity checks",
		zap.Int("local_candidates", len(localCandidates)),
		zap.Int("remote_candidates", len(remoteCandidates)))
	
	// Generate candidate pairs
	pairs := cc.generateCandidatePairs(localCandidates, remoteCandidates)
	cc.logger.Debug("Generated candidate pairs", zap.Int("count", len(pairs)))
	
	// Store pairs
	cc.mu.Lock()
	for _, pair := range pairs {
		cc.activePairs[pair.ID] = pair
	}
	cc.mu.Unlock()
	
	// Start checking pairs
	go cc.performConnectivityChecks(pairs)
	
	return nil
}

// generateCandidatePairs генерирует пары кандидатов
func (cc *ConnectivityChecker) generateCandidatePairs(local, remote []*ICECandidate) []*CandidatePair {
	var pairs []*CandidatePair
	
	for i, localCand := range local {
		for j, remoteCand := range remote {
			pair := &CandidatePair{
				ID:         fmt.Sprintf("pair_%d_%d", i, j),
				Local:      localCand,
				Remote:     remoteCand,
				Priority:   cc.calculatePairPriority(localCand, remoteCand),
				State:      PairStateWaiting,
				Foundation: fmt.Sprintf("%s_%s", localCand.Foundation, remoteCand.Foundation),
				CreatedAt:  time.Now(),
			}
			pairs = append(pairs, pair)
		}
	}
	
	// Sort pairs by priority (higher priority first)
	cc.sortPairsByPriority(pairs)
	
	return pairs
}

// calculatePairPriority вычисляет приоритет пары (RFC 8445)
func (cc *ConnectivityChecker) calculatePairPriority(local, remote *ICECandidate) uint64 {
	// G = controlling agent priority, D = controlled agent priority
	G := uint64(local.Priority)
	D := uint64(remote.Priority)
	
	// pair priority = 2^32 * MIN(G, D) + 2 * MAX(G, D) + (G > D ? 1 : 0)
	var min, max uint64
	var tieBreaker uint64
	
	if G > D {
		max = G
		min = D
		tieBreaker = 1
	} else {
		max = D
		min = G
		tieBreaker = 0
	}
	
	return (1<<32)*min + 2*max + tieBreaker
}

// sortPairsByPriority сортирует пары по приоритету
func (cc *ConnectivityChecker) sortPairsByPriority(pairs []*CandidatePair) {
	// Simple bubble sort for demonstration - in production use sort.Slice
	n := len(pairs)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if pairs[j].Priority < pairs[j+1].Priority {
				pairs[j], pairs[j+1] = pairs[j+1], pairs[j]
			}
		}
	}
}

// performConnectivityChecks выполняет connectivity checks
func (cc *ConnectivityChecker) performConnectivityChecks(pairs []*CandidatePair) {
	semaphore := make(chan struct{}, cc.maxConcurrentChecks)
	var wg sync.WaitGroup
	
	for _, pair := range pairs {
		select {
		case <-cc.ctx.Done():
			cc.logger.Info("Connectivity checks cancelled")
			return
		case semaphore <- struct{}{}:
			wg.Add(1)
			go func(p *CandidatePair) {
				defer func() {
					<-semaphore
					wg.Done()
				}()
				
				cc.performPairCheck(p)
			}(pair)
		}
	}
	
	wg.Wait()
	cc.logger.Info("All connectivity checks completed")
	
	// Select the best pair
	cc.selectBestPair()
}

// performPairCheck выполняет connectivity check для пары
func (cc *ConnectivityChecker) performPairCheck(pair *CandidatePair) {
	cc.logger.Debug("Checking candidate pair",
		zap.String("pair_id", pair.ID),
		zap.String("local", fmt.Sprintf("%s:%d", pair.Local.Address, pair.Local.Port)),
		zap.String("remote", fmt.Sprintf("%s:%d", pair.Remote.Address, pair.Remote.Port)))
	
	// Update pair state
	cc.mu.Lock()
	pair.State = PairStateInProgress
	pair.LastChecked = time.Now()
	cc.mu.Unlock()
	
	startTime := time.Now()
	result := &CheckResult{
		PairID:    pair.ID,
		Timestamp: startTime,
	}
	
	// Perform the actual connectivity check
	success, rtt, err := cc.performSTUNCheck(pair)
	
	result.Success = success
	result.RTT = rtt
	result.Error = err
	
	// Update pair state based on result
	cc.mu.Lock()
	if success {
		pair.State = PairStateSucceeded
		cc.logger.Debug("Connectivity check succeeded",
			zap.String("pair_id", pair.ID),
			zap.Duration("rtt", rtt))
	} else {
		pair.State = PairStateFailed
		cc.logger.Debug("Connectivity check failed",
			zap.String("pair_id", pair.ID),
			zap.Error(err))
	}
	
	cc.checkedPairs[pair.ID] = result
	cc.mu.Unlock()
	
	// Update metrics
	connectionTime := time.Since(startTime)
	cc.metrics.UpdateConnectivityMetrics(success, connectionTime)
}

// performSTUNCheck выполняет STUN connectivity check
func (cc *ConnectivityChecker) performSTUNCheck(pair *CandidatePair) (bool, time.Duration, error) {
	// Create local address for binding
	localAddr := &net.UDPAddr{
		IP:   net.ParseIP(pair.Local.Address),
		Port: pair.Local.Port,
	}
	
	// Create remote address
	remoteAddr := &net.UDPAddr{
		IP:   net.ParseIP(pair.Remote.Address),
		Port: pair.Remote.Port,
	}
	
	// For host candidates, we can try direct UDP connection
	if pair.Local.Type == ICECandidateTypeHost && pair.Remote.Type == ICECandidateTypeHost {
		return cc.performDirectUDPCheck(localAddr, remoteAddr)
	}
	
	// For server-reflexive or relay candidates, use STUN binding
	return cc.performSTUNBindingCheck(localAddr, remoteAddr)
}

// performDirectUDPCheck выполняет прямую UDP проверку
func (cc *ConnectivityChecker) performDirectUDPCheck(localAddr, remoteAddr *net.UDPAddr) (bool, time.Duration, error) {
	startTime := time.Now()
	
	// Create UDP connection
	conn, err := net.DialUDP("udp", localAddr, remoteAddr)
	if err != nil {
		return false, 0, fmt.Errorf("failed to create UDP connection: %w", err)
	}
	defer conn.Close()
	
	// Set timeout
	conn.SetDeadline(time.Now().Add(cc.checkTimeout))
	
	// Send test packet
	testData := []byte("ICE connectivity check")
	_, err = conn.Write(testData)
	if err != nil {
		return false, 0, fmt.Errorf("failed to send test packet: %w", err)
	}
	
	// Try to read response (this is simplified - in real ICE we'd expect STUN response)
	buffer := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	_, err = conn.Read(buffer)
	
	rtt := time.Since(startTime)
	
	// For this simplified check, we consider it successful if we can send
	// In real ICE implementation, we'd need proper STUN binding request/response
	return err == nil, rtt, nil
}

// performSTUNBindingCheck выполняет STUN binding check
func (cc *ConnectivityChecker) performSTUNBindingCheck(localAddr, remoteAddr *net.UDPAddr) (bool, time.Duration, error) {
	startTime := time.Now()
	
	// Create UDP connection
	conn, err := net.DialUDP("udp", localAddr, remoteAddr)
	if err != nil {
		return false, 0, fmt.Errorf("failed to create UDP connection: %w", err)
	}
	defer conn.Close()
	
	// Set timeout
	conn.SetDeadline(time.Now().Add(cc.checkTimeout))
	
	// Create STUN binding request
	message := stun.MustBuild(stun.TransactionID, stun.BindingRequest)
	
	// Send STUN request
	_, err = conn.Write(message.Raw)
	if err != nil {
		return false, 0, fmt.Errorf("failed to send STUN request: %w", err)
	}
	
	// Read response
	buffer := make([]byte, 1500)
	n, err := conn.Read(buffer)
	if err != nil {
		return false, 0, fmt.Errorf("failed to read STUN response: %w", err)
	}
	
	// Parse STUN response
	response := &stun.Message{Raw: buffer[:n]}
	err = response.Decode()
	if err != nil {
		return false, 0, fmt.Errorf("failed to decode STUN response: %w", err)
	}
	
	rtt := time.Since(startTime)
	
	// Check if it's a successful binding response
	if response.Type == stun.BindingSuccess {
		return true, rtt, nil
	}
	
	return false, rtt, fmt.Errorf("STUN binding failed: %s", response.Type)
}

// selectBestPair выбирает лучшую пару для использования
func (cc *ConnectivityChecker) selectBestPair() {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	
	var bestPair *CandidatePair
	var bestPriority uint64
	
	for _, pair := range cc.activePairs {
		if pair.State == PairStateSucceeded {
			if bestPair == nil || pair.Priority > bestPriority {
				bestPair = pair
				bestPriority = pair.Priority
			}
		}
	}
	
	if bestPair != nil {
		cc.selectedPair = bestPair
		cc.logger.Info("Selected best candidate pair",
			zap.String("pair_id", bestPair.ID),
			zap.String("local", fmt.Sprintf("%s:%d (%s)", 
				bestPair.Local.Address, bestPair.Local.Port, bestPair.Local.Type)),
			zap.String("remote", fmt.Sprintf("%s:%d (%s)", 
				bestPair.Remote.Address, bestPair.Remote.Port, bestPair.Remote.Type)),
			zap.Uint64("priority", bestPair.Priority))
	} else {
		cc.logger.Warn("No successful candidate pairs found")
	}
}

// GetSelectedPair возвращает выбранную пару кандидатов
func (cc *ConnectivityChecker) GetSelectedPair() *CandidatePair {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	return cc.selectedPair
}

// GetCheckResults возвращает результаты всех проверок
func (cc *ConnectivityChecker) GetCheckResults() map[string]*CheckResult {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	
	results := make(map[string]*CheckResult)
	for k, v := range cc.checkedPairs {
		results[k] = v
	}
	return results
}

// GetActivePairs возвращает активные пары кандидатов
func (cc *ConnectivityChecker) GetActivePairs() map[string]*CandidatePair {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	
	pairs := make(map[string]*CandidatePair)
	for k, v := range cc.activePairs {
		pairs[k] = v
	}
	return pairs
}

// Stop останавливает connectivity checker
func (cc *ConnectivityChecker) Stop() {
	cc.cancel()
	cc.logger.Info("Connectivity checker stopped")
}

// GetConnectivityReport генерирует отчет о connectivity
func (cc *ConnectivityChecker) GetConnectivityReport() *ConnectivityReport {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	
	report := &ConnectivityReport{
		TotalPairs:      len(cc.activePairs),
		SuccessfulPairs: 0,
		FailedPairs:     0,
		CheckedPairs:    len(cc.checkedPairs),
		SelectedPair:    cc.selectedPair,
		GeneratedAt:     time.Now(),
	}
	
	for _, pair := range cc.activePairs {
		switch pair.State {
		case PairStateSucceeded:
			report.SuccessfulPairs++
		case PairStateFailed:
			report.FailedPairs++
		}
	}
	
	if report.TotalPairs > 0 {
		report.SuccessRate = float64(report.SuccessfulPairs) / float64(report.TotalPairs) * 100
	}
	
	return report
}

// ConnectivityReport отчет о connectivity checks
type ConnectivityReport struct {
	TotalPairs      int            `json:"total_pairs"`
	CheckedPairs    int            `json:"checked_pairs"`
	SuccessfulPairs int            `json:"successful_pairs"`
	FailedPairs     int            `json:"failed_pairs"`
	SuccessRate     float64        `json:"success_rate"`
	SelectedPair    *CandidatePair `json:"selected_pair"`
	GeneratedAt     time.Time      `json:"generated_at"`
}
