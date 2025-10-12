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

// ConnectivityChecker выполняет connectivity checks для ICE
type ConnectivityChecker struct {
	logger *logger.Logger

	// Метрики
	metrics           *NATMetrics
	prometheusMetrics *NATPrometheusMetrics

	// Конфигурация
	checkTimeout    time.Duration
	checkInterval   time.Duration
	maxRetries      int
	parallelChecks  int

	// Результаты
	checkResults      map[string]*CheckResult
	checkResultsMutex sync.RWMutex

	ctx    context.Context
	cancel context.CancelFunc
}

// CheckResult результат connectivity check
type CheckResult struct {
	LocalCandidate  *ICECandidate
	RemoteCandidate *ICECandidate
	Success         bool
	Latency         time.Duration
	Error           error
	Timestamp       time.Time
	RetryCount      int
}

// NewConnectivityChecker создает новый connectivity checker
func NewConnectivityChecker(logger *logger.Logger, metrics *NATMetrics) *ConnectivityChecker {
	ctx, cancel := context.WithCancel(context.Background())

	prometheusMetrics := NewNATPrometheusMetrics()

	return &ConnectivityChecker{
		logger:            logger,
		metrics:           metrics,
		prometheusMetrics: prometheusMetrics,
		checkTimeout:      5 * time.Second,
		checkInterval:     200 * time.Millisecond,
		maxRetries:        3,
		parallelChecks:    10,
		checkResults:      make(map[string]*CheckResult),
		ctx:               ctx,
		cancel:            cancel,
	}
}

// StartConnectivityChecks запускает connectivity checks для пар кандидатов
func (cc *ConnectivityChecker) StartConnectivityChecks(localCandidates, remoteCandidates []*ICECandidate) error {
	if len(localCandidates) == 0 || len(remoteCandidates) == 0 {
		return fmt.Errorf("no candidates provided for connectivity checks")
	}

	cc.logger.Info("Starting connectivity checks",
		zap.Int("local_candidates", len(localCandidates)),
		zap.Int("remote_candidates", len(remoteCandidates)))

	// Создаем пары кандидатов
	pairs := cc.createCandidatePairs(localCandidates, remoteCandidates)
	cc.logger.Info("Created candidate pairs", zap.Int("pairs", len(pairs)))

	// Сортируем по приоритету
	cc.sortPairsByPriority(pairs)

	// Запускаем проверки параллельно
	return cc.performChecksParallel(pairs)
}

// createCandidatePairs создает все возможные пары кандидатов
func (cc *ConnectivityChecker) createCandidatePairs(local, remote []*ICECandidate) []*CandidatePair {
	pairs := make([]*CandidatePair, 0, len(local)*len(remote))

	for _, l := range local {
		for _, r := range remote {
			priority := cc.calculatePairPriority(l, r)
			pair := &CandidatePair{
				LocalCandidate:  l,
				RemoteCandidate: r,
				Priority:        priority,
				State:           ICECandidatePairStateWaiting,
			}
			pairs = append(pairs, pair)
		}
	}

	return pairs
}

// calculatePairPriority вычисляет приоритет пары (RFC 5245)
func (cc *ConnectivityChecker) calculatePairPriority(local, remote *ICECandidate) uint64 {
	g := uint64(local.Priority)
	d := uint64(remote.Priority)

	if g > d {
		return (1<<32)*d + 2*g + 1
	}
	return (1<<32)*g + 2*d + 0
}

// sortPairsByPriority сортирует пары по приоритету (по убыванию)
func (cc *ConnectivityChecker) sortPairsByPriority(pairs []*CandidatePair) {
	// Простая сортировка пузырьком
	for i := 0; i < len(pairs)-1; i++ {
		for j := i + 1; j < len(pairs); j++ {
			if pairs[i].Priority < pairs[j].Priority {
				pairs[i], pairs[j] = pairs[j], pairs[i]
			}
		}
	}
}

// performChecksParallel выполняет проверки параллельно
func (cc *ConnectivityChecker) performChecksParallel(pairs []*CandidatePair) error {
	// Канал для ограничения количества параллельных проверок
	semaphore := make(chan struct{}, cc.parallelChecks)

	// WaitGroup для ожидания завершения всех проверок
	var wg sync.WaitGroup

	// Канал для результатов
	resultsChan := make(chan *CheckResult, len(pairs))

	// Запускаем проверки
	for _, pair := range pairs {
		wg.Add(1)
		go func(p *CandidatePair) {
			defer wg.Done()

			// Ждем свободного слота
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Выполняем проверку
			result := cc.performConnectivityCheck(p)
			resultsChan <- result
		}(pair)
	}

	// Горутина для ожидания завершения
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Собираем результаты
	successCount := 0
	failureCount := 0

	for result := range resultsChan {
		key := fmt.Sprintf("%s:%d->%s:%d",
			result.LocalCandidate.Address,
			result.LocalCandidate.Port,
			result.RemoteCandidate.Address,
			result.RemoteCandidate.Port)

		cc.checkResultsMutex.Lock()
		cc.checkResults[key] = result
		cc.checkResultsMutex.Unlock()

		if result.Success {
			successCount++
			cc.metrics.UpdateConnectivityMetrics(true, result.Latency)
			cc.prometheusMetrics.RecordConnectivityCheck(true, result.Latency.Seconds())
		} else {
			failureCount++
			cc.metrics.UpdateConnectivityMetrics(false, 0)
			cc.prometheusMetrics.RecordConnectivityCheck(false, 0)
		}
	}

	cc.logger.Info("Connectivity checks completed",
		zap.Int("success", successCount),
		zap.Int("failure", failureCount),
		zap.Float64("success_rate", float64(successCount)/float64(successCount+failureCount)*100))

	return nil
}

// performConnectivityCheck выполняет connectivity check для одной пары
func (cc *ConnectivityChecker) performConnectivityCheck(pair *CandidatePair) *CheckResult {
	result := &CheckResult{
		LocalCandidate:  pair.LocalCandidate,
		RemoteCandidate: pair.RemoteCandidate,
		Timestamp:       time.Now(),
	}

	// Обновляем состояние пары
	pair.State = ICECandidatePairStateInProgress

	// Пробуем несколько раз
	for retry := 0; retry < cc.maxRetries; retry++ {
		start := time.Now()

		success, err := cc.checkConnectivity(pair.LocalCandidate, pair.RemoteCandidate)
		latency := time.Since(start)

		if success {
			result.Success = true
			result.Latency = latency
			pair.State = ICECandidatePairStateSucceeded

			cc.logger.Debug("Connectivity check succeeded",
				zap.String("local", pair.LocalCandidate.Address),
				zap.Int("local_port", pair.LocalCandidate.Port),
				zap.String("remote", pair.RemoteCandidate.Address),
				zap.Int("remote_port", pair.RemoteCandidate.Port),
				zap.Duration("latency", latency),
				zap.Int("retry", retry))

			return result
		}

		result.RetryCount = retry + 1
		result.Error = err

		// Ждем перед следующей попыткой
		if retry < cc.maxRetries-1 {
			time.Sleep(cc.checkInterval)
		}
	}

	// Все попытки не удались
	pair.State = ICECandidatePairStateFailed
	result.Success = false

	cc.logger.Debug("Connectivity check failed",
		zap.String("local", pair.LocalCandidate.Address),
		zap.Int("local_port", pair.LocalCandidate.Port),
		zap.String("remote", pair.RemoteCandidate.Address),
		zap.Int("remote_port", pair.RemoteCandidate.Port),
		zap.Int("retries", cc.maxRetries),
		zap.Error(result.Error))

	return result
}

// checkConnectivity выполняет базовую проверку connectivity
func (cc *ConnectivityChecker) checkConnectivity(local, remote *ICECandidate) (bool, error) {
	ctx, cancel := context.WithTimeout(cc.ctx, cc.checkTimeout)
	defer cancel()

	// Выбираем протокол
	network := "udp"
	if remote.Transport == "TCP" {
		network = "tcp"
	}

	// Пытаемся установить соединение
	dialer := &net.Dialer{
		Timeout: cc.checkTimeout,
	}

	remoteAddr := fmt.Sprintf("%s:%d", remote.Address, remote.Port)
	conn, err := dialer.DialContext(ctx, network, remoteAddr)
	if err != nil {
		return false, err
	}
	defer conn.Close()

	// Для UDP отправляем тестовый пакет
	if network == "udp" {
		// Отправляем STUN Binding Request (упрощенный вариант)
		testData := []byte{0x00, 0x01, 0x00, 0x00, 0x21, 0x12, 0xA4, 0x42} // STUN header
		conn.SetWriteDeadline(time.Now().Add(1 * time.Second))
		if _, err := conn.Write(testData); err != nil {
			return false, err
		}

		// Пытаемся прочитать ответ
		buffer := make([]byte, 1024)
		conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		if _, err := conn.Read(buffer); err != nil {
			// Для UDP если нет ответа, это нормально (порт может быть open)
			return true, nil
		}
	}

	return true, nil
}

// GetSuccessfulPairs возвращает успешные пары кандидатов
func (cc *ConnectivityChecker) GetSuccessfulPairs() []*CheckResult {
	cc.checkResultsMutex.RLock()
	defer cc.checkResultsMutex.RUnlock()

	var successful []*CheckResult
	for _, result := range cc.checkResults {
		if result.Success {
			successful = append(successful, result)
		}
	}

	return successful
}

// GetBestPair возвращает лучшую пару кандидатов (с наименьшей латентностью)
func (cc *ConnectivityChecker) GetBestPair() *CheckResult {
	successful := cc.GetSuccessfulPairs()
	if len(successful) == 0 {
		return nil
	}

	best := successful[0]
	for _, result := range successful[1:] {
		if result.Latency < best.Latency {
			best = result
		}
	}

	return best
}

// GetCheckResults возвращает все результаты проверок
func (cc *ConnectivityChecker) GetCheckResults() map[string]*CheckResult {
	cc.checkResultsMutex.RLock()
	defer cc.checkResultsMutex.RUnlock()

	// Создаем копию для безопасности
	results := make(map[string]*CheckResult, len(cc.checkResults))
	for k, v := range cc.checkResults {
		results[k] = v
	}

	return results
}

// GetSuccessRate возвращает процент успешных проверок
func (cc *ConnectivityChecker) GetSuccessRate() float64 {
	cc.checkResultsMutex.RLock()
	defer cc.checkResultsMutex.RUnlock()

	if len(cc.checkResults) == 0 {
		return 0
	}

	successCount := 0
	for _, result := range cc.checkResults {
		if result.Success {
			successCount++
		}
	}

	return float64(successCount) / float64(len(cc.checkResults)) * 100
}

// GetPrometheusMetrics возвращает Prometheus метрики
func (cc *ConnectivityChecker) GetPrometheusMetrics() *NATPrometheusMetrics {
	return cc.prometheusMetrics
}

// Close закрывает connectivity checker
func (cc *ConnectivityChecker) Close() error {
	cc.cancel()
	cc.logger.Info("Connectivity checker closed")
	return nil
}

// CandidatePair представляет пару кандидатов для connectivity check
type CandidatePair struct {
	LocalCandidate  *ICECandidate
	RemoteCandidate *ICECandidate
	Priority        uint64
	State           ICECandidatePairState
}
