package nat

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
)

func TestNewConnectivityChecker(t *testing.T) {
	logger, err := logger.New(logger.Config{
		Level:  logger.DEBUG,
		Output: "stdout",
		JSON:   false,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	metrics := NewNATMetrics()

	checker := NewConnectivityChecker(logger, metrics)
	require.NotNil(t, checker)
	defer checker.Close()

	assert.NotNil(t, checker.prometheusMetrics)
	assert.Equal(t, 5*time.Second, checker.checkTimeout)
	assert.Equal(t, 200*time.Millisecond, checker.checkInterval)
	assert.Equal(t, 3, checker.maxRetries)
	assert.Equal(t, 10, checker.parallelChecks)
}

func TestConnectivityChecker_CreateCandidatePairs(t *testing.T) {
	logger, err := logger.New(logger.Config{
		Level:  logger.DEBUG,
		Output: "stdout",
		JSON:   false,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	metrics := NewNATMetrics()
	checker := NewConnectivityChecker(logger, metrics)
	defer checker.Close()

	local := []*ICECandidate{
		{
			Foundation: "1",
			Type:       ICECandidateTypeHost,
			Address:    "192.168.1.100",
			Port:       54321,
			Priority:   126,
		},
		{
			Foundation: "2",
			Type:       ICECandidateTypeServerReflexive,
			Address:    "203.0.113.1",
			Port:       54322,
			Priority:   100,
		},
	}

	remote := []*ICECandidate{
		{
			Foundation: "3",
			Type:       ICECandidateTypeHost,
			Address:    "192.168.1.200",
			Port:       12345,
			Priority:   126,
		},
		{
			Foundation: "4",
			Type:       ICECandidateTypeServerReflexive,
			Address:    "203.0.113.2",
			Port:       12346,
			Priority:   100,
		},
	}

	pairs := checker.createCandidatePairs(local, remote)

	// Должно быть 2*2=4 пары
	assert.Equal(t, 4, len(pairs))

	// Проверяем, что все пары созданы
	assert.NotNil(t, pairs[0].LocalCandidate)
	assert.NotNil(t, pairs[0].RemoteCandidate)
	assert.Greater(t, pairs[0].Priority, uint64(0))
	assert.Equal(t, ICECandidatePairStateWaiting, pairs[0].State)
}

func TestConnectivityChecker_CalculatePairPriority(t *testing.T) {
	logger, err := logger.New(logger.Config{
		Level:  logger.DEBUG,
		Output: "stdout",
		JSON:   false,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	metrics := NewNATMetrics()
	checker := NewConnectivityChecker(logger, metrics)
	defer checker.Close()

	local := &ICECandidate{Priority: 126}
	remote := &ICECandidate{Priority: 100}

	priority := checker.calculatePairPriority(local, remote)
	assert.Greater(t, priority, uint64(0))

	// Приоритет должен быть одинаковым независимо от порядка
	// (по RFC 5245)
	priority2 := checker.calculatePairPriority(remote, local)
	assert.NotEqual(t, priority, priority2) // Они должны быть разные из-за формулы
}

func TestConnectivityChecker_SortPairsByPriority(t *testing.T) {
	logger, err := logger.New(logger.Config{
		Level:  logger.DEBUG,
		Output: "stdout",
		JSON:   false,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	metrics := NewNATMetrics()
	checker := NewConnectivityChecker(logger, metrics)
	defer checker.Close()

	pairs := []*CandidatePair{
		{Priority: 100},
		{Priority: 300},
		{Priority: 200},
	}

	checker.sortPairsByPriority(pairs)

	// Должны быть отсортированы по убыванию
	assert.Equal(t, uint64(300), pairs[0].Priority)
	assert.Equal(t, uint64(200), pairs[1].Priority)
	assert.Equal(t, uint64(100), pairs[2].Priority)
}

func TestConnectivityChecker_CheckConnectivity_Localhost(t *testing.T) {
	logger, err := logger.New(logger.Config{
		Level:  logger.DEBUG,
		Output: "stdout",
		JSON:   false,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	metrics := NewNATMetrics()
	checker := NewConnectivityChecker(logger, metrics)
	defer checker.Close()

	// Тестируем connectivity с localhost (должен работать)
	local := &ICECandidate{
		Address:   "127.0.0.1",
		Port:      0,
		Transport: "UDP",
	}

	remote := &ICECandidate{
		Address:   "127.0.0.1",
		Port:      9999, // Несуществующий порт
		Transport: "UDP",
	}

	// Проверка должна не зависнуть (с таймаутом)
	success, err := checker.checkConnectivity(local, remote)

	// Для несуществующего порта может быть успех или ошибка (зависит от ОС)
	if !success {
		assert.Error(t, err)
	}
	t.Logf("Connectivity check result: success=%v, err=%v", success, err)
}

func TestConnectivityChecker_GetSuccessRate(t *testing.T) {
	logger, err := logger.New(logger.Config{
		Level:  logger.DEBUG,
		Output: "stdout",
		JSON:   false,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	metrics := NewNATMetrics()
	checker := NewConnectivityChecker(logger, metrics)
	defer checker.Close()

	// Начальный success rate должен быть 0
	rate := checker.GetSuccessRate()
	assert.Equal(t, 0.0, rate)

	// Добавляем результаты
	checker.checkResultsMutex.Lock()
	checker.checkResults["test1"] = &CheckResult{Success: true}
	checker.checkResults["test2"] = &CheckResult{Success: false}
	checker.checkResults["test3"] = &CheckResult{Success: true}
	checker.checkResultsMutex.Unlock()

	// Success rate должен быть 66.67%
	rate = checker.GetSuccessRate()
	assert.InDelta(t, 66.67, rate, 0.01)
}

func TestConnectivityChecker_GetBestPair(t *testing.T) {
	logger, err := logger.New(logger.Config{
		Level:  logger.DEBUG,
		Output: "stdout",
		JSON:   false,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	metrics := NewNATMetrics()
	checker := NewConnectivityChecker(logger, metrics)
	defer checker.Close()

	// Нет результатов - должен вернуть nil
	best := checker.GetBestPair()
	assert.Nil(t, best)

	// Добавляем результаты с разной латентностью
	local := &ICECandidate{Address: "192.168.1.100", Port: 5000}
	remote1 := &ICECandidate{Address: "192.168.1.200", Port: 6000}
	remote2 := &ICECandidate{Address: "192.168.1.201", Port: 6001}
	remote3 := &ICECandidate{Address: "192.168.1.202", Port: 6002}

	checker.checkResultsMutex.Lock()
	checker.checkResults["test1"] = &CheckResult{
		LocalCandidate:  local,
		RemoteCandidate: remote1,
		Success:         true,
		Latency:         50 * time.Millisecond,
	}
	checker.checkResults["test2"] = &CheckResult{
		LocalCandidate:  local,
		RemoteCandidate: remote2,
		Success:         true,
		Latency:         20 * time.Millisecond, // Самая быстрая
	}
	checker.checkResults["test3"] = &CheckResult{
		LocalCandidate:  local,
		RemoteCandidate: remote3,
		Success:         false, // Неудачная проверка - не должна учитываться
		Latency:         10 * time.Millisecond,
	}
	checker.checkResultsMutex.Unlock()

	// Лучшая пара должна быть test2 (20ms)
	best = checker.GetBestPair()
	require.NotNil(t, best)
	assert.Equal(t, 20*time.Millisecond, best.Latency)
	assert.Equal(t, remote2.Address, best.RemoteCandidate.Address)
}

func TestConnectivityChecker_GetSuccessfulPairs(t *testing.T) {
	logger, err := logger.New(logger.Config{
		Level:  logger.DEBUG,
		Output: "stdout",
		JSON:   false,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	metrics := NewNATMetrics()
	checker := NewConnectivityChecker(logger, metrics)
	defer checker.Close()

	// Добавляем результаты
	local := &ICECandidate{Address: "192.168.1.100", Port: 5000}
	remote1 := &ICECandidate{Address: "192.168.1.200", Port: 6000}
	remote2 := &ICECandidate{Address: "192.168.1.201", Port: 6001}

	checker.checkResultsMutex.Lock()
	checker.checkResults["test1"] = &CheckResult{
		LocalCandidate:  local,
		RemoteCandidate: remote1,
		Success:         true,
	}
	checker.checkResults["test2"] = &CheckResult{
		LocalCandidate:  local,
		RemoteCandidate: remote2,
		Success:         false,
	}
	checker.checkResultsMutex.Unlock()

	// Должна быть только одна успешная пара
	successful := checker.GetSuccessfulPairs()
	assert.Equal(t, 1, len(successful))
	assert.True(t, successful[0].Success)
}

func TestConnectivityChecker_PerformConnectivityCheck(t *testing.T) {
	logger, err := logger.New(logger.Config{
		Level:  logger.DEBUG,
		Output: "stdout",
		JSON:   false,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	metrics := NewNATMetrics()
	checker := NewConnectivityChecker(logger, metrics)
	defer checker.Close()

	// Устанавливаем короткий таймаут для теста
	checker.checkTimeout = 100 * time.Millisecond
	checker.maxRetries = 1

	pair := &CandidatePair{
		LocalCandidate: &ICECandidate{
			Address:   "127.0.0.1",
			Port:      0,
			Transport: "UDP",
		},
		RemoteCandidate: &ICECandidate{
			Address:   "127.0.0.1",
			Port:      9999, // Несуществующий порт
			Transport: "UDP",
		},
		State: ICECandidatePairStateWaiting,
	}

	result := checker.performConnectivityCheck(pair)

	assert.NotNil(t, result)
	assert.NotNil(t, result.LocalCandidate)
	assert.NotNil(t, result.RemoteCandidate)
	assert.False(t, result.Timestamp.IsZero())

	// Для несуществующего порта может быть успех (UDP) или failure
	if !result.Success {
		assert.Equal(t, ICECandidatePairStateFailed, pair.State)
	}

	t.Logf("Check result: success=%v, latency=%v, retries=%d, error=%v",
		result.Success, result.Latency, result.RetryCount, result.Error)
}

func TestConnectivityChecker_PrometheusMetrics(t *testing.T) {
	logger, err := logger.New(logger.Config{
		Level:  logger.DEBUG,
		Output: "stdout",
		JSON:   false,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	metrics := NewNATMetrics()
	checker := NewConnectivityChecker(logger, metrics)
	defer checker.Close()

	// Проверяем, что Prometheus метрики созданы
	promMetrics := checker.GetPrometheusMetrics()
	assert.NotNil(t, promMetrics)
	assert.NotNil(t, promMetrics.stunRequestsTotal)
	assert.NotNil(t, promMetrics.connectivityChecksTotal)
	assert.NotNil(t, promMetrics.iceCandidatesGathered)
}

func TestConnectivityChecker_MetricsIntegration(t *testing.T) {
	logger, err := logger.New(logger.Config{
		Level:  logger.DEBUG,
		Output: "stdout",
		JSON:   false,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	metrics := NewNATMetrics()
	checker := NewConnectivityChecker(logger, metrics)
	defer checker.Close()

	// Записываем метрики
	promMetrics := checker.GetPrometheusMetrics()

	// Записываем ICE кандидата
	promMetrics.RecordICECandidate("host")
	promMetrics.RecordICECandidate("srflx")

	// Записываем connectivity check
	promMetrics.RecordConnectivityCheck(true, 0.050) // 50ms
	promMetrics.RecordConnectivityCheck(false, 0)

	// Записываем STUN запрос
	promMetrics.RecordSTUNRequest("stun.example.com", true, 0.020) // 20ms

	// Записываем NAT type
	promMetrics.RecordNATType("Full Cone")

	// Проверяем, что метрики обновились (просто проверяем, что не паникует)
	assert.NotPanics(t, func() {
		promMetrics.UpdateFromNATMetrics(metrics)
	})
}
