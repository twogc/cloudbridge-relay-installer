package nat

import (
	"context"
	"testing"
	"time"

	"github.com/pion/ice/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
)

func TestNewPionICEAdapter(t *testing.T) {
	logger := logger.NewLogger()
	config := DefaultPionICEConfig()
	metrics := NewNATMetrics()

	adapter, err := NewPionICEAdapter(logger, config, metrics)
	require.NoError(t, err)
	require.NotNil(t, adapter)
	defer adapter.Close()

	assert.Equal(t, ice.ConnectionStateNew, adapter.GetConnectionState())
	assert.Equal(t, ice.GatheringStateNew, adapter.GetGatheringState())
}

func TestPionICEAdapter_GatherCandidates(t *testing.T) {
	logger := logger.NewLogger()
	config := DefaultPionICEConfig()
	config.STUNServers = []string{
		"stun:stun.l.google.com:19302",
	}
	metrics := NewNATMetrics()

	adapter, err := NewPionICEAdapter(logger, config, metrics)
	require.NoError(t, err)
	defer adapter.Close()

	// Канал для сбора кандидатов
	candidatesReceived := make(chan *ice.Candidate, 10)
	adapter.OnCandidate(func(candidate *ice.Candidate) {
		candidatesReceived <- candidate
	})

	// Канал для завершения сбора
	gatheringComplete := make(chan bool, 1)
	adapter.OnGatheringComplete(func() {
		gatheringComplete <- true
	})

	// Запускаем сбор кандидатов
	err = adapter.GatherCandidates()
	require.NoError(t, err)

	// Ждем завершения сбора (максимум 10 секунд)
	timeout := time.After(10 * time.Second)
	candidatesCollected := 0

collecting:
	for {
		select {
		case <-candidatesReceived:
			candidatesCollected++
		case <-gatheringComplete:
			break collecting
		case <-timeout:
			t.Fatal("Timeout waiting for candidate gathering")
		}
	}

	// Проверяем, что собрали хотя бы 1 кандидат
	assert.Greater(t, candidatesCollected, 0, "Should have collected at least one candidate")

	// Проверяем состояние
	assert.True(t, adapter.IsGatheringComplete())
	assert.Equal(t, ice.GatheringStateComplete, adapter.GetGatheringState())

	// Проверяем локальные кандидаты
	localCandidates := adapter.GetLocalCandidates()
	assert.Greater(t, len(localCandidates), 0, "Should have local candidates")

	// Проверяем метрики
	metrics = adapter.GetMetrics()
	totalCandidates := metrics.GetTotalCandidates()
	assert.Greater(t, totalCandidates, 0, "Metrics should show candidates")
	assert.Greater(t, metrics.HostCandidatesGathered, 0, "Should have host candidates")

	t.Logf("Gathered %d candidates (Host: %d, Srflx: %d, Relay: %d)",
		totalCandidates,
		metrics.HostCandidatesGathered,
		metrics.ServerReflexiveCandidatesGathered,
		metrics.RelayCandidatesGathered)
}

func TestPionICEAdapter_LocalParameters(t *testing.T) {
	logger := logger.NewLogger()
	adapter, err := NewPionICEAdapter(logger, nil, nil)
	require.NoError(t, err)
	defer adapter.Close()

	role, ufrag, pwd, err := adapter.GetLocalParameters()
	require.NoError(t, err)
	assert.NotEmpty(t, ufrag)
	assert.NotEmpty(t, pwd)
	assert.Equal(t, ice.ConnectionRoleControlling, role)

	t.Logf("Local parameters: role=%s, ufrag=%s, pwd=%s", role, ufrag, pwd)
}

func TestPionICEAdapter_ConnectionStateChanges(t *testing.T) {
	logger := logger.NewLogger()
	adapter, err := NewPionICEAdapter(logger, nil, nil)
	require.NoError(t, err)
	defer adapter.Close()

	// Канал для отслеживания изменений состояния
	stateChanges := make(chan ice.ConnectionState, 10)
	adapter.OnConnectionStateChange(func(state ice.ConnectionState) {
		stateChanges <- state
	})

	// Начальное состояние должно быть New
	assert.Equal(t, ice.ConnectionStateNew, adapter.GetConnectionState())

	// Запускаем сбор кандидатов
	err = adapter.GatherCandidates()
	require.NoError(t, err)

	// Ждем изменений состояния (или таймаут)
	select {
	case state := <-stateChanges:
		t.Logf("Connection state changed to: %s", state)
	case <-time.After(5 * time.Second):
		t.Log("No state changes within 5 seconds (expected for gathering only)")
	}
}

func TestPionICEAdapter_Restart(t *testing.T) {
	logger := logger.NewLogger()
	adapter, err := NewPionICEAdapter(logger, nil, nil)
	require.NoError(t, err)
	defer adapter.Close()

	// Запускаем сбор кандидатов
	err = adapter.GatherCandidates()
	require.NoError(t, err)

	// Ждем немного для сбора кандидатов
	time.Sleep(2 * time.Second)

	// Проверяем, что есть кандидаты
	candidates := adapter.GetLocalCandidates()
	assert.Greater(t, len(candidates), 0)

	// Перезапускаем
	err = adapter.Restart()
	require.NoError(t, err)

	// Проверяем, что состояние сброшено
	assert.Equal(t, ice.ConnectionStateNew, adapter.GetConnectionState())
	assert.Equal(t, ice.GatheringStateNew, adapter.GetGatheringState())
	assert.False(t, adapter.IsGatheringComplete())

	// Проверяем, что кандидаты очищены
	candidates = adapter.GetLocalCandidates()
	assert.Equal(t, 0, len(candidates))
}

func TestPionICEAdapter_WaitForGatheringComplete(t *testing.T) {
	logger := logger.NewLogger()
	adapter, err := NewPionICEAdapter(logger, nil, nil)
	require.NoError(t, err)
	defer adapter.Close()

	// Запускаем сбор кандидатов
	err = adapter.GatherCandidates()
	require.NoError(t, err)

	// Ждем завершения сбора с таймаутом
	err = adapter.WaitForGatheringComplete(10 * time.Second)
	require.NoError(t, err)

	assert.True(t, adapter.IsGatheringComplete())
}

func TestPionICEAdapter_GetStats(t *testing.T) {
	logger := logger.NewLogger()
	adapter, err := NewPionICEAdapter(logger, nil, nil)
	require.NoError(t, err)
	defer adapter.Close()

	// Запускаем сбор кандидатов
	err = adapter.GatherCandidates()
	require.NoError(t, err)

	// Ждем немного для сбора статистики
	time.Sleep(2 * time.Second)

	// Получаем статистику
	stats := adapter.GetStats()
	assert.NotNil(t, stats)

	t.Logf("ICE Stats: %+v", stats)
}

func TestPionICEAdapter_MultipleAdapters(t *testing.T) {
	logger := logger.NewLogger()

	// Создаем два адаптера для имитации двух пиров
	adapter1, err := NewPionICEAdapter(logger, nil, nil)
	require.NoError(t, err)
	defer adapter1.Close()

	adapter2, err := NewPionICEAdapter(logger, nil, nil)
	require.NoError(t, err)
	defer adapter2.Close()

	// Получаем локальные параметры для обоих
	_, ufrag1, pwd1, err := adapter1.GetLocalParameters()
	require.NoError(t, err)

	_, ufrag2, pwd2, err := adapter2.GetLocalParameters()
	require.NoError(t, err)

	// Проверяем, что параметры разные
	assert.NotEqual(t, ufrag1, ufrag2)
	assert.NotEqual(t, pwd1, pwd2)

	t.Logf("Adapter1: ufrag=%s, pwd=%s", ufrag1, pwd1)
	t.Logf("Adapter2: ufrag=%s, pwd=%s", ufrag2, pwd2)
}

func TestPionICEAdapter_AddRemoteCandidate(t *testing.T) {
	logger := logger.NewLogger()
	adapter, err := NewPionICEAdapter(logger, nil, nil)
	require.NoError(t, err)
	defer adapter.Close()

	// Пример удаленного кандидата (host candidate)
	candidateInit := ice.CandidateInit{
		Candidate: "candidate:1 1 UDP 2130706431 192.168.1.100 54321 typ host",
	}

	// Добавляем удаленный кандидат
	err = adapter.AddRemoteCandidate(candidateInit)
	require.NoError(t, err)

	// Проверяем, что кандидат добавлен
	adapter.mu.RLock()
	remoteCount := len(adapter.remoteCandidates)
	adapter.mu.RUnlock()

	assert.Equal(t, 1, remoteCount, "Should have 1 remote candidate")
}

func TestPionICEAdapter_MetricsIntegration(t *testing.T) {
	logger := logger.NewLogger()
	metrics := NewNATMetrics()

	adapter, err := NewPionICEAdapter(logger, nil, metrics)
	require.NoError(t, err)
	defer adapter.Close()

	// Запускаем сбор кандидатов
	err = adapter.GatherCandidates()
	require.NoError(t, err)

	// Ждем завершения
	err = adapter.WaitForGatheringComplete(10 * time.Second)
	require.NoError(t, err)

	// Получаем метрики
	finalMetrics := adapter.GetMetrics()
	assert.Greater(t, finalMetrics.GetTotalCandidates(), 0)
	assert.Greater(t, finalMetrics.GatheringDuration, time.Duration(0))

	t.Logf("Metrics: Total candidates=%d, Gathering duration=%v",
		finalMetrics.GetTotalCandidates(),
		finalMetrics.GatheringDuration)
}

func TestPionICEAdapter_CandidateTypes(t *testing.T) {
	logger := logger.NewLogger()
	config := DefaultPionICEConfig()
	// Используем только публичный STUN сервер
	config.STUNServers = []string{"stun:stun.l.google.com:19302"}

	adapter, err := NewPionICEAdapter(logger, config, nil)
	require.NoError(t, err)
	defer adapter.Close()

	// Канал для сбора типов кандидатов
	candidateTypes := make(map[ice.CandidateType]int)
	candidateTypesMu := make(chan struct{}, 1)

	adapter.OnCandidate(func(candidate *ice.Candidate) {
		<-candidateTypesMu
		candidateTypes[(*candidate).Type()]++
		candidateTypesMu <- struct{}{}
	})

	candidateTypesMu <- struct{}{} // Initial lock

	// Запускаем сбор
	err = adapter.GatherCandidates()
	require.NoError(t, err)

	// Ждем завершения
	err = adapter.WaitForGatheringComplete(15 * time.Second)
	require.NoError(t, err)

	<-candidateTypesMu
	t.Logf("Candidate types collected:")
	for candidateType, count := range candidateTypes {
		t.Logf("  - %s: %d", candidateType, count)
	}

	// Должны быть хотя бы host кандидаты
	assert.Greater(t, candidateTypes[ice.CandidateTypeHost], 0, "Should have host candidates")
}

func TestPionICEAdapter_Dial_Accept_Integration(t *testing.T) {
	t.Skip("Skipping full dial/accept integration test - requires network setup")

	logger := logger.NewLogger()

	// Создаем два адаптера
	adapter1, err := NewPionICEAdapter(logger, nil, nil)
	require.NoError(t, err)
	defer adapter1.Close()

	adapter2, err := NewPionICEAdapter(logger, nil, nil)
	require.NoError(t, err)
	defer adapter2.Close()

	// Получаем credentials
	_, ufrag1, pwd1, err := adapter1.GetLocalParameters()
	require.NoError(t, err)

	_, ufrag2, pwd2, err := adapter2.GetLocalParameters()
	require.NoError(t, err)

	// Запускаем сбор кандидатов
	err = adapter1.GatherCandidates()
	require.NoError(t, err)

	err = adapter2.GatherCandidates()
	require.NoError(t, err)

	// Ждем завершения
	err = adapter1.WaitForGatheringComplete(10 * time.Second)
	require.NoError(t, err)

	err = adapter2.WaitForGatheringComplete(10 * time.Second)
	require.NoError(t, err)

	// Обмениваем кандидаты
	candidates1 := adapter1.GetLocalCandidates()
	for _, candidate := range candidates1 {
		init := ice.CandidateInit{
			Candidate: candidate.String(),
		}
		err = adapter2.AddRemoteCandidate(init)
		require.NoError(t, err)
	}

	candidates2 := adapter2.GetLocalCandidates()
	for _, candidate := range candidates2 {
		init := ice.CandidateInit{
			Candidate: candidate.String(),
		}
		err = adapter1.AddRemoteCandidate(init)
		require.NoError(t, err)
	}

	// Устанавливаем соединение
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// adapter1 делает dial
	go func() {
		conn, err := adapter1.Dial(ctx, ufrag2, pwd2)
		if err != nil {
			t.Logf("Dial error: %v", err)
			return
		}
		defer conn.Close()
		t.Log("Dial successful")
	}()

	// adapter2 делает accept
	conn, err := adapter2.Accept(ctx, ufrag1, pwd1)
	if err != nil {
		t.Fatalf("Accept failed: %v", err)
	}
	defer conn.Close()

	t.Log("Connection established successfully")

	// Проверяем состояние соединения
	assert.Equal(t, ice.ConnectionStateConnected, adapter1.GetConnectionState())
	assert.Equal(t, ice.ConnectionStateConnected, adapter2.GetConnectionState())
}
