package nat

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewICEAgent(t *testing.T) {
	logger := zap.NewNop()
	stunServers := []string{"stun1.example.com:3478", "stun2.example.com:3478"}
	turnServers := []*TURNServer{
		{
			Address:  "turn1.example.com:3478",
			Username: "user1",
			Password: "pass1",
			Realm:    "example.com",
		},
	}

	agent := NewICEAgent(stunServers, turnServers, logger)

	assert.NotNil(t, agent)
	assert.Equal(t, stunServers, agent.stunServers)
	assert.Equal(t, turnServers, agent.turnServers)
	assert.NotNil(t, agent.localCandidates)
	assert.NotNil(t, agent.remoteCandidates)
	assert.Equal(t, ICEStateNew, agent.state)
	assert.NotNil(t, agent.metrics)
}

func TestICEAgent_StartGathering(t *testing.T) {
	logger := zap.NewNop()
	stunServers := []string{"stun1.example.com:3478"}
	turnServers := []*TURNServer{}

	agent := NewICEAgent(stunServers, turnServers, logger)

	// Запускаем сбор кандидатов
	err := agent.StartGathering()
	assert.NoError(t, err)

	// Проверяем, что состояние изменилось
	assert.Equal(t, ICEStateGathered, agent.state)

	// Проверяем, что собрались host кандидаты
	assert.Greater(t, len(agent.localCandidates), 0)

	// Проверяем, что все кандидаты имеют правильный тип
	hasHostCandidate := false
	for _, candidate := range agent.localCandidates {
		if candidate.Type == CandidateTypeHost {
			hasHostCandidate = true
			break
		}
	}
	assert.True(t, hasHostCandidate)
}

func TestICEAgent_RecordConnectionAttempt(t *testing.T) {
	logger := zap.NewNop()
	agent := NewICEAgent([]string{}, []*TURNServer{}, logger)

	// Записываем успешную попытку соединения
	agent.RecordConnectionAttempt(true, 100*time.Millisecond)

	// Проверяем метрики
	stats := agent.GetConnectionStats()
	assert.Len(t, stats, 1)

	// Получаем тип NAT (должен быть "unknown" так как нет кандидатов)
	natType := agent.detectNATType()
	assert.Equal(t, "unknown", natType)

	// Проверяем, что метрика записалась
	natStats := agent.GetNATTypeStats()
	assert.Equal(t, int64(1), natStats[natType])
}

func TestICEAgent_DetectNATType(t *testing.T) {
	logger := zap.NewNop()
	agent := NewICEAgent([]string{}, []*TURNServer{}, logger)

	// Тест 1: Нет кандидатов - должен быть "unknown"
	natType := agent.detectNATType()
	assert.Equal(t, "unknown", natType)

	// Тест 2: Только host кандидаты - должен быть "no_nat"
	agent.localCandidates = []*Candidate{
		{
			Type:      CandidateTypeHost,
			Address:   "192.168.1.1",
			Port:      8080,
			Transport: "UDP",
		},
	}
	natType = agent.detectNATType()
	assert.Equal(t, "no_nat", natType)

	// Тест 3: Host + Server Reflexive - должен быть "cone_nat"
	agent.localCandidates = append(agent.localCandidates, &Candidate{
		Type:      CandidateTypeServerReflexive,
		Address:   "203.0.113.1",
		Port:      8080,
		Transport: "UDP",
	})
	natType = agent.detectNATType()
	assert.Equal(t, "cone_nat", natType)

	// Тест 4: Host + Server Reflexive + Relay - должен быть "symmetric_nat"
	agent.localCandidates = append(agent.localCandidates, &Candidate{
		Type:      CandidateTypeRelay,
		Address:   "198.51.100.1",
		Port:      8080,
		Transport: "UDP",
	})
	natType = agent.detectNATType()
	assert.Equal(t, "symmetric_nat", natType)

	// Тест 5: Только Server Reflexive - должен быть "restricted_nat"
	agent.localCandidates = []*Candidate{
		{
			Type:      CandidateTypeServerReflexive,
			Address:   "203.0.113.1",
			Port:      8080,
			Transport: "UDP",
		},
	}
	natType = agent.detectNATType()
	assert.Equal(t, "restricted_nat", natType)

	// Тест 6: Только Relay - должен быть "blocked_nat"
	agent.localCandidates = []*Candidate{
		{
			Type:      CandidateTypeRelay,
			Address:   "198.51.100.1",
			Port:      8080,
			Transport: "UDP",
		},
	}
	natType = agent.detectNATType()
	assert.Equal(t, "blocked_nat", natType)
}

func TestICEAgent_GetMetrics(t *testing.T) {
	logger := zap.NewNop()
	agent := NewICEAgent([]string{}, []*TURNServer{}, logger)

	metrics := agent.GetMetrics()
	assert.NotNil(t, metrics)
	assert.IsType(t, &NATMetrics{}, metrics)
}

func TestICEAgent_GetNATTypeStats(t *testing.T) {
	logger := zap.NewNop()
	agent := NewICEAgent([]string{}, []*TURNServer{}, logger)

	// Записываем несколько попыток соединения
	agent.RecordConnectionAttempt(true, 100*time.Millisecond)
	agent.RecordConnectionAttempt(false, 200*time.Millisecond)

	stats := agent.GetNATTypeStats()
	assert.NotNil(t, stats)
	assert.Equal(t, int64(2), stats["unknown"]) // Оба раза определился как "unknown"
}

func TestICEAgent_GetConnectionStats(t *testing.T) {
	logger := zap.NewNop()
	agent := NewICEAgent([]string{}, []*TURNServer{}, logger)

	// Записываем несколько попыток соединения
	agent.RecordConnectionAttempt(true, 100*time.Millisecond)
	agent.RecordConnectionAttempt(false, 200*time.Millisecond)

	stats := agent.GetConnectionStats()
	assert.NotNil(t, stats)
	assert.Len(t, stats, 1) // Только один тип NAT

	unknownNATStats := stats["unknown"]
	assert.NotNil(t, unknownNATStats)
	assert.Equal(t, int64(2), unknownNATStats.TotalAttempts)
	assert.Equal(t, int64(1), unknownNATStats.SuccessfulConns)
	assert.Equal(t, int64(1), unknownNATStats.FailedConns)
}

func TestICEAgent_GetOverallStats(t *testing.T) {
	logger := zap.NewNop()
	agent := NewICEAgent([]string{}, []*TURNServer{}, logger)

	// Записываем тестовые данные
	agent.RecordConnectionAttempt(true, 100*time.Millisecond)
	agent.RecordConnectionAttempt(false, 200*time.Millisecond)

	stats := agent.GetOverallStats()
	assert.NotNil(t, stats)
	assert.Equal(t, int64(2), stats["total_connections"])
	assert.Equal(t, int64(1), stats["successful_connections"])
	assert.Equal(t, 0.5, stats["success_rate"]) // 1 из 2 успешных
}

func TestCandidateType_String(t *testing.T) {
	tests := []struct {
		candidateType CandidateType
		expected      string
	}{
		{CandidateTypeHost, "host"},
		{CandidateTypeServerReflexive, "srflx"},
		{CandidateTypePeerReflexive, "prflx"},
		{CandidateTypeRelay, "relay"},
		{CandidateType(999), "unknown"},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, test.candidateType.String())
	}
}

func TestICEAgent_GatherHostCandidates(t *testing.T) {
	logger := zap.NewNop()
	agent := NewICEAgent([]string{}, []*TURNServer{}, logger)

	// Запускаем сбор host кандидатов
	err := agent.gatherHostCandidates()
	assert.NoError(t, err)

	// Проверяем, что собрались кандидаты
	assert.Greater(t, len(agent.localCandidates), 0)

	// Проверяем, что все кандидаты имеют правильный тип
	for _, candidate := range agent.localCandidates {
		assert.Equal(t, CandidateTypeHost, candidate.Type)
		assert.Equal(t, "UDP", candidate.Transport)
		assert.NotEmpty(t, candidate.Address)
		assert.Greater(t, candidate.Priority, uint32(0))
	}
}

func TestICEAgent_CalculatePriority(t *testing.T) {
	logger := zap.NewNop()
	agent := NewICEAgent([]string{}, []*TURNServer{}, logger)

	// Тестируем расчет приоритета для разных типов кандидатов
	hostPriority := agent.calculatePriority(CandidateTypeHost, 0)
	srflxPriority := agent.calculatePriority(CandidateTypeServerReflexive, 0)
	relayPriority := agent.calculatePriority(CandidateTypeRelay, 0)

	// Host кандидаты должны иметь самый высокий приоритет
	assert.Greater(t, hostPriority, srflxPriority)
	assert.Greater(t, srflxPriority, relayPriority)
}

// Benchmark тесты
func BenchmarkICEAgent_DetectNATType(b *testing.B) {
	logger := zap.NewNop()
	agent := NewICEAgent([]string{}, []*TURNServer{}, logger)

	// Добавляем тестовые кандидаты
	agent.localCandidates = []*Candidate{
		{Type: CandidateTypeHost, Address: "192.168.1.1", Port: 8080, Transport: "UDP"},
		{Type: CandidateTypeServerReflexive, Address: "203.0.113.1", Port: 8080, Transport: "UDP"},
		{Type: CandidateTypeRelay, Address: "198.51.100.1", Port: 8080, Transport: "UDP"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		agent.detectNATType()
	}
}

func BenchmarkICEAgent_RecordConnectionAttempt(b *testing.B) {
	logger := zap.NewNop()
	agent := NewICEAgent([]string{}, []*TURNServer{}, logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		agent.RecordConnectionAttempt(i%2 == 0, 100*time.Millisecond)
	}
}

func BenchmarkICEAgent_GetOverallStats(b *testing.B) {
	logger := zap.NewNop()
	agent := NewICEAgent([]string{}, []*TURNServer{}, logger)

	// Заполняем метрики тестовыми данными
	for i := 0; i < 1000; i++ {
		agent.RecordConnectionAttempt(i%2 == 0, 100*time.Millisecond)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		agent.GetOverallStats()
	}
}
