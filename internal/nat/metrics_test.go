package nat

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewNATMetrics(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewNATMetrics(logger)

	assert.NotNil(t, metrics)
	assert.NotNil(t, metrics.iceCandidatesTotal)
	assert.NotNil(t, metrics.iceConnectionsTotal)
	assert.NotNil(t, metrics.iceConnectionDuration)
	assert.NotNil(t, metrics.natTypeDetected)
	assert.NotNil(t, metrics.turnAllocationsTotal)
	assert.NotNil(t, metrics.stunRequestsTotal)
	assert.NotNil(t, metrics.connectionSuccessRate)
	assert.NotNil(t, metrics.natTypeStats)
	assert.NotNil(t, metrics.connectionStats)
	assert.NotNil(t, metrics.candidateStats)
	assert.NotNil(t, metrics.turnServerStats)
}

func TestNATMetrics_RecordICECandidate(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewNATMetrics(logger)

	// Записываем различные типы кандидатов
	metrics.RecordICECandidate("host", "UDP")
	metrics.RecordICECandidate("srflx", "UDP")
	metrics.RecordICECandidate("relay", "UDP")
	metrics.RecordICECandidate("prflx", "UDP")

	// Проверяем статистику кандидатов
	stats := metrics.GetCandidateStats()
	assert.Len(t, stats, 1) // Только UDP transport

	udpStats := stats["UDP"]
	assert.NotNil(t, udpStats)
	assert.Equal(t, int64(1), udpStats.HostCandidates)
	assert.Equal(t, int64(1), udpStats.ServerReflexiveCands)
	assert.Equal(t, int64(1), udpStats.RelayCandidates)
	assert.Equal(t, int64(1), udpStats.PeerReflexiveCands)
	assert.Equal(t, int64(4), udpStats.TotalCandidates)
}

func TestNATMetrics_RecordICEConnection(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewNATMetrics(logger)

	// Записываем успешное соединение
	metrics.RecordICEConnection("success", "cone_nat", 100*time.Millisecond)

	// Записываем неудачное соединение
	metrics.RecordICEConnection("failed", "symmetric_nat", 500*time.Millisecond)

	// Проверяем статистику соединений
	stats := metrics.GetConnectionStats()
	assert.Len(t, stats, 2)

	coneStats := stats["cone_nat"]
	assert.NotNil(t, coneStats)
	assert.Equal(t, int64(1), coneStats.TotalAttempts)
	assert.Equal(t, int64(1), coneStats.SuccessfulConns)
	assert.Equal(t, int64(0), coneStats.FailedConns)
	assert.Equal(t, 100*time.Millisecond, coneStats.AverageLatency)

	symmetricStats := stats["symmetric_nat"]
	assert.NotNil(t, symmetricStats)
	assert.Equal(t, int64(1), symmetricStats.TotalAttempts)
	assert.Equal(t, int64(0), symmetricStats.SuccessfulConns)
	assert.Equal(t, int64(1), symmetricStats.FailedConns)
	assert.Equal(t, 500*time.Millisecond, symmetricStats.AverageLatency)
}

func TestNATMetrics_RecordNATType(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewNATMetrics(logger)

	// Записываем различные типы NAT
	metrics.RecordNATType("no_nat")
	metrics.RecordNATType("cone_nat")
	metrics.RecordNATType("symmetric_nat")
	metrics.RecordNATType("cone_nat") // Дублируем

	// Проверяем статистику типов NAT
	stats := metrics.GetNATTypeStats()
	assert.Equal(t, int64(1), stats["no_nat"])
	assert.Equal(t, int64(2), stats["cone_nat"])
	assert.Equal(t, int64(1), stats["symmetric_nat"])
}

func TestNATMetrics_RecordTURNAllocation(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewNATMetrics(logger)

	// Записываем успешные TURN allocations
	metrics.RecordTURNAllocation("turn1.example.com", "success", 200*time.Millisecond)
	metrics.RecordTURNAllocation("turn2.example.com", "success", 300*time.Millisecond)

	// Записываем неудачную TURN allocation
	metrics.RecordTURNAllocation("turn1.example.com", "failed", 1000*time.Millisecond)

	// Проверяем статистику TURN серверов
	stats := metrics.GetTURNServerStats()
	assert.Len(t, stats, 2)

	turn1Stats := stats["turn1.example.com"]
	assert.NotNil(t, turn1Stats)
	assert.Equal(t, int64(2), turn1Stats.TotalAllocations)
	assert.Equal(t, int64(1), turn1Stats.ActiveAllocations)
	assert.Equal(t, int64(1), turn1Stats.FailedAllocations)
	assert.Equal(t, 200*time.Millisecond, turn1Stats.AverageLatency)

	turn2Stats := stats["turn2.example.com"]
	assert.NotNil(t, turn2Stats)
	assert.Equal(t, int64(1), turn2Stats.TotalAllocations)
	assert.Equal(t, int64(1), turn2Stats.ActiveAllocations)
	assert.Equal(t, int64(0), turn2Stats.FailedAllocations)
	assert.Equal(t, 300*time.Millisecond, turn2Stats.AverageLatency)
}

func TestNATMetrics_RecordSTUNRequest(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewNATMetrics(logger)

	// Записываем STUN запросы
	metrics.RecordSTUNRequest("stun1.example.com", "success")
	metrics.RecordSTUNRequest("stun2.example.com", "success")
	metrics.RecordSTUNRequest("stun1.example.com", "failed")

	// Метрики STUN запросов проверяются через Prometheus, здесь просто убеждаемся, что нет ошибок
	// В реальном тесте можно было бы использовать prometheus testutil для проверки метрик
}

func TestNATMetrics_GetOverallStats(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewNATMetrics(logger)

	// Записываем тестовые данные
	metrics.RecordICEConnection("success", "cone_nat", 100*time.Millisecond)
	metrics.RecordICEConnection("failed", "symmetric_nat", 500*time.Millisecond)
	metrics.RecordICECandidate("host", "UDP")
	metrics.RecordICECandidate("srflx", "UDP")
	metrics.RecordNATType("cone_nat")
	metrics.RecordNATType("symmetric_nat")
	metrics.RecordTURNAllocation("turn.example.com", "success", 200*time.Millisecond)

	// Получаем общую статистику
	stats := metrics.GetOverallStats()

	assert.Equal(t, int64(2), stats["total_connections"])
	assert.Equal(t, int64(1), stats["successful_connections"])
	assert.Equal(t, 0.5, stats["success_rate"]) // 1 из 2 успешных
	assert.Equal(t, int64(2), stats["total_candidates"])
	assert.Equal(t, 2, stats["nat_types_detected"])
	assert.Equal(t, 1, stats["turn_servers_active"])
}

func TestNATMetrics_Reset(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewNATMetrics(logger)

	// Записываем тестовые данные
	metrics.RecordICEConnection("success", "cone_nat", 100*time.Millisecond)
	metrics.RecordICECandidate("host", "UDP")
	metrics.RecordNATType("cone_nat")
	metrics.RecordTURNAllocation("turn.example.com", "success", 200*time.Millisecond)

	// Проверяем, что данные есть
	stats := metrics.GetOverallStats()
	assert.Equal(t, int64(1), stats["total_connections"])

	// Сбрасываем метрики
	metrics.Reset()

	// Проверяем, что данные очищены
	stats = metrics.GetOverallStats()
	assert.Equal(t, int64(0), stats["total_connections"])
	assert.Equal(t, int64(0), stats["successful_connections"])
	assert.Equal(t, int64(0), stats["total_candidates"])
	assert.Equal(t, 0, stats["nat_types_detected"])
	assert.Equal(t, 0, stats["turn_servers_active"])
}

func TestNATMetrics_ConnectionSuccessRate(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewNATMetrics(logger)

	// Записываем несколько соединений для одного типа NAT
	metrics.RecordICEConnection("success", "cone_nat", 100*time.Millisecond)
	metrics.RecordICEConnection("failed", "cone_nat", 200*time.Millisecond)
	metrics.RecordICEConnection("success", "cone_nat", 150*time.Millisecond)

	// Проверяем статистику соединений
	stats := metrics.GetConnectionStats()
	coneStats := stats["cone_nat"]
	assert.NotNil(t, coneStats)
	assert.Equal(t, int64(3), coneStats.TotalAttempts)
	assert.Equal(t, int64(2), coneStats.SuccessfulConns)
	assert.Equal(t, int64(1), coneStats.FailedConns)

	// Проверяем общую статистику
	overallStats := metrics.GetOverallStats()
	expectedSuccessRate := float64(2) / float64(3) // 2 из 3 успешных
	assert.Equal(t, expectedSuccessRate, overallStats["success_rate"])
}

func TestNATMetrics_AverageLatency(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewNATMetrics(logger)

	// Записываем соединения с разной задержкой
	metrics.RecordICEConnection("success", "cone_nat", 100*time.Millisecond)
	metrics.RecordICEConnection("success", "cone_nat", 300*time.Millisecond)

	// Проверяем среднюю задержку
	stats := metrics.GetConnectionStats()
	coneStats := stats["cone_nat"]
	assert.NotNil(t, coneStats)
	
	// Средняя задержка должна быть (100 + 300) / 2 = 200ms
	expectedAverage := (100*time.Millisecond + 300*time.Millisecond) / 2
	assert.Equal(t, expectedAverage, coneStats.AverageLatency)
}

// Benchmark тесты
func BenchmarkNATMetrics_RecordICECandidate(b *testing.B) {
	logger := zap.NewNop()
	metrics := NewNATMetrics(logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metrics.RecordICECandidate("host", "UDP")
	}
}

func BenchmarkNATMetrics_RecordICEConnection(b *testing.B) {
	logger := zap.NewNop()
	metrics := NewNATMetrics(logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metrics.RecordICEConnection("success", "cone_nat", 100*time.Millisecond)
	}
}

func BenchmarkNATMetrics_GetOverallStats(b *testing.B) {
	logger := zap.NewNop()
	metrics := NewNATMetrics(logger)

	// Заполняем метрики тестовыми данными
	for i := 0; i < 1000; i++ {
		metrics.RecordICEConnection("success", "cone_nat", 100*time.Millisecond)
		metrics.RecordICECandidate("host", "UDP")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metrics.GetOverallStats()
	}
}



