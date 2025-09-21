package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/twogc/cloudbridge-relay-installer/config"
	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
	"github.com/twogc/cloudbridge-relay-installer/internal/nat"
)

func TestMASQUEICEIntegration(t *testing.T) {
	// Создаем тестовый логгер
	testLogger, err := logger.New(logger.Config{
		Level:  logger.DEBUG,
		JSON:   true,
		Output: "stdout",
	})
	require.NoError(t, err)

	// Создаем тестовую конфигурацию
	cfg := &config.Config{
		MASQUE: struct {
			Enabled         bool     `json:"enabled"`
			ProxyURL        string   `json:"proxy_url"`
			ALPNProtocols   []string `json:"alpn_protocols"`
			FallbackEnabled bool     `json:"fallback_enabled"`
		}{
			Enabled:         true,
			ProxyURL:        "https://test.example.com:8443",
			ALPNProtocols:   []string{"h3", "h3-29"},
			FallbackEnabled: true,
		},
		Server: struct {
			Address         string        `json:"address"`
			Port            int           `json:"port"`
			Timeout         time.Duration `json:"timeout"`
			MaxTunnels      int           `json:"max_tunnels"`
			PingInterval    time.Duration `json:"ping_interval"`
			CleanupInterval time.Duration `json:"cleanup_interval"`
		}{
			Address:         "0.0.0.0",
			Port:            8080,
			Timeout:         30 * time.Second,
			MaxTunnels:      100,
			PingInterval:    30 * time.Second,
			CleanupInterval: 60 * time.Second,
		},
	}

	// Создаем MASQUE сервер
	masqueServer := NewMASQUEServer(testLogger, cfg)
	require.NotNil(t, masqueServer)

	// Создаем ICE агент
	iceAgent := nat.NewICEAgent(
		[]string{"stun1.example.com:3478"},
		[]*nat.TURNServer{},
		zap.NewNop(),
	)
	require.NotNil(t, iceAgent)

	// Тест 1: MASQUE сервер создается корректно
	t.Run("MASQUEServer_Creation", func(t *testing.T) {
		assert.NotNil(t, masqueServer.udpProxy)
		assert.NotNil(t, masqueServer.ipProxy)
		assert.NotNil(t, masqueServer.connections)
		assert.False(t, masqueServer.isRunning)
	})

	// Тест 2: ICE агент создается корректно
	t.Run("ICEAgent_Creation", func(t *testing.T) {
		assert.NotNil(t, iceAgent.GetMetrics())
		assert.Equal(t, nat.ICEStateNew, iceAgent.GetState())
	})

	// Тест 3: MASQUE HTTP Datagrams обработчик
	t.Run("MASQUE_HTTPDatagrams_Handler", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/.well-known/masque/datagrams/", nil)
		req.Header.Set("Capsule-Protocol", "?1")

		w := httptest.NewRecorder()
		masqueServer.handleHTTPDatagrams(w, req)

		// HTTP Datagrams требует HTTP/3 stream, поэтому ожидаем ошибку в тесте
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Not an HTTP/3 stream")
	})

	// Тест 4: ICE кандидаты собираются
	t.Run("ICE_Candidate_Gathering", func(t *testing.T) {
		err := iceAgent.StartGathering()
		assert.NoError(t, err)

		// Проверяем, что состояние изменилось
		assert.Equal(t, nat.ICEStateGathered, iceAgent.GetState())

		// Проверяем, что собрались кандидаты
		candidates := iceAgent.GetLocalCandidates()
		assert.Greater(t, len(candidates), 0)

		// Проверяем, что есть host кандидаты
		hasHostCandidate := false
		for _, candidate := range candidates {
			if candidate.Type == nat.CandidateTypeHost {
				hasHostCandidate = true
				break
			}
		}
		assert.True(t, hasHostCandidate)
	})

	// Тест 5: ICE метрики записываются
	t.Run("ICE_Metrics_Recording", func(t *testing.T) {
		// Записываем тестовые метрики
		iceAgent.RecordConnectionAttempt(true, 100*time.Millisecond)
		iceAgent.RecordConnectionAttempt(false, 200*time.Millisecond)

		// Проверяем статистику
		stats := iceAgent.GetOverallStats()
		assert.NotNil(t, stats)
		assert.Equal(t, int64(2), stats["total_connections"])
		assert.Equal(t, int64(1), stats["successful_connections"])
		assert.Equal(t, 0.5, stats["success_rate"])
	})

	// Тест 6: MASQUE соединения управляются
	t.Run("MASQUE_Connection_Management", func(t *testing.T) {
		// Изначально соединений нет
		count := masqueServer.GetConnectionCount()
		assert.Equal(t, 0, count)

		// Создаем тестовое соединение
		conn := &MASQUEConnection{
			ID:              "test_conn_1",
			DatagramChannel: make(chan []byte, 10),
			CreatedAt:       time.Now(),
			LastActivity:    time.Now(),
		}

		// Добавляем соединение
		masqueServer.connectionsMutex.Lock()
		masqueServer.connections[conn.ID] = conn
		masqueServer.connectionsMutex.Unlock()

		// Проверяем количество соединений
		count = masqueServer.GetConnectionCount()
		assert.Equal(t, 1, count)

		// Проверяем информацию о соединениях
		info := masqueServer.GetConnectionInfo()
		assert.Equal(t, 1, info["total_connections"])
		assert.Equal(t, false, info["is_running"]) // Сервер не запущен в тесте
	})

	// Тест 7: MASQUE datagram маршрутизация
	t.Run("MASQUE_Datagram_Routing", func(t *testing.T) {
		// Очищаем существующие соединения
		masqueServer.connectionsMutex.Lock()
		masqueServer.connections = make(map[string]*MASQUEConnection)
		masqueServer.connectionsMutex.Unlock()

		// Создаем два тестовых соединения
		conn1 := &MASQUEConnection{
			ID:              "peer1",
			DatagramChannel: make(chan []byte, 10),
			CreatedAt:       time.Now(),
			LastActivity:    time.Now(),
		}

		conn2 := &MASQUEConnection{
			ID:              "peer",
			DatagramChannel: make(chan []byte, 10),
			CreatedAt:       time.Now(),
			LastActivity:    time.Now(),
		}

		// Добавляем соединения
		masqueServer.connectionsMutex.Lock()
		masqueServer.connections[conn1.ID] = conn1
		masqueServer.connections[conn2.ID] = conn2
		masqueServer.connectionsMutex.Unlock()

		// Создаем datagram для маршрутизации: [target_peer_id:4][payload:...]
		testPayload := []byte("test message")
		datagram := make([]byte, 4+len(testPayload))
		copy(datagram[:4], "peer") // target peer ID (4 символа)
		copy(datagram[4:], testPayload)

		// Обрабатываем datagram
		err := masqueServer.processDatagram(datagram, "127.0.0.1:12345")
		assert.NoError(t, err)

		// Проверяем, что datagram был доставлен
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		receivedData, err := conn2.ReceiveDatagram(ctx)
		assert.NoError(t, err)
		assert.Equal(t, testPayload, receivedData)
	})

	// Тест 8: Fallback механизм
	t.Run("Fallback_Mechanism", func(t *testing.T) {
		// Создаем FallbackManager
		fallbackManager := NewFallbackManager(testLogger, cfg)
		require.NotNil(t, fallbackManager)

		// Запускаем FallbackManager
		err := fallbackManager.Start()
		assert.NoError(t, err)

		// Проверяем статистику
		stats := fallbackManager.GetConnectionStats()
		assert.NotNil(t, stats)
		assert.Equal(t, true, stats["is_running"])

		// Закрываем FallbackManager
		err = fallbackManager.Close()
		assert.NoError(t, err)
	})

	// Тест 9: Интеграция MASQUE + ICE
	t.Run("MASQUE_ICE_Integration", func(t *testing.T) {
		// Очищаем существующие соединения
		masqueServer.connectionsMutex.Lock()
		masqueServer.connections = make(map[string]*MASQUEConnection)
		masqueServer.connectionsMutex.Unlock()

		// Создаем MASQUE соединение с ICE кандидатами
		iceCandidates := iceAgent.GetLocalCandidates()
		assert.Greater(t, len(iceCandidates), 0)

		// Создаем MASQUE соединение
		masqueConn := &MASQUEConnection{
			ID:              "ice_masque_conn",
			DatagramChannel: make(chan []byte, 10),
			CreatedAt:       time.Now(),
			LastActivity:    time.Now(),
		}

		// Добавляем соединение
		masqueServer.connectionsMutex.Lock()
		masqueServer.connections[masqueConn.ID] = masqueConn
		masqueServer.connectionsMutex.Unlock()

		// Записываем метрики ICE соединения
		iceAgent.RecordConnectionAttempt(true, 150*time.Millisecond)

		// Проверяем, что метрики записались
		iceStats := iceAgent.GetOverallStats()
		assert.Greater(t, iceStats["total_connections"], int64(0))

		// Проверяем MASQUE соединение
		masqueCount := masqueServer.GetConnectionCount()
		assert.Equal(t, 1, masqueCount)
	})

	// Тест 10: Производительность
	t.Run("Performance_Test", func(t *testing.T) {
		// Очищаем существующие соединения
		masqueServer.connectionsMutex.Lock()
		masqueServer.connections = make(map[string]*MASQUEConnection)
		masqueServer.connectionsMutex.Unlock()

		// Создаем много соединений
		const numConnections = 50
		for i := 0; i < numConnections; i++ {
			conn := &MASQUEConnection{
				ID:              fmt.Sprintf("perf_conn_%d", i),
				DatagramChannel: make(chan []byte, 10),
				CreatedAt:       time.Now(),
				LastActivity:    time.Now(),
			}

			masqueServer.connectionsMutex.Lock()
			masqueServer.connections[conn.ID] = conn
			masqueServer.connectionsMutex.Unlock()
		}

		// Проверяем количество соединений
		count := masqueServer.GetConnectionCount()
		assert.Equal(t, numConnections, count)

		// Проверяем производительность отправки datagrams
		start := time.Now()
		for i := 0; i < 100; i++ {
			conn := &MASQUEConnection{
				ID:              fmt.Sprintf("perf_test_%d", i),
				DatagramChannel: make(chan []byte, 100),
				CreatedAt:       time.Now(),
				LastActivity:    time.Now(),
			}

			testData := []byte(fmt.Sprintf("performance test data %d", i))
			err := conn.SendDatagram(testData)
			assert.NoError(t, err)
		}
		duration := time.Since(start)

		// Проверяем, что отправка была быстрой (менее 1 секунды для 100 datagrams)
		assert.Less(t, duration, 1*time.Second)
	})
}

// Benchmark тесты

func BenchmarkICEAgent_DetectNATType(b *testing.B) {
	iceAgent := nat.NewICEAgent(
		[]string{"stun1.example.com:3478"},
		[]*nat.TURNServer{},
		zap.NewNop(),
	)

	// Добавляем тестовые кандидаты
	iceAgent.AddLocalCandidate(&nat.Candidate{
		Type:      nat.CandidateTypeHost,
		Address:   "192.168.1.1",
		Port:      8080,
		Transport: "UDP",
	})
	iceAgent.AddLocalCandidate(&nat.Candidate{
		Type:      nat.CandidateTypeServerReflexive,
		Address:   "203.0.113.1",
		Port:      8080,
		Transport: "UDP",
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iceAgent.DetectNATType()
	}
}

func BenchmarkMASQUEProcessDatagram(b *testing.B) {
	testLogger, _ := logger.New(logger.Config{Level: logger.DEBUG, JSON: true})
	cfg := &config.Config{
		MASQUE: struct {
			Enabled         bool     `json:"enabled"`
			ProxyURL        string   `json:"proxy_url"`
			ALPNProtocols   []string `json:"alpn_protocols"`
			FallbackEnabled bool     `json:"fallback_enabled"`
		}{
			Enabled: true,
		},
	}

	masqueServer := NewMASQUEServer(testLogger, cfg)

	// Создаем тестовое соединение
	conn := &MASQUEConnection{
		ID:              "bench_target",
		DatagramChannel: make(chan []byte, 1000),
		CreatedAt:       time.Now(),
		LastActivity:    time.Now(),
	}

	masqueServer.connectionsMutex.Lock()
	masqueServer.connections[conn.ID] = conn
	masqueServer.connectionsMutex.Unlock()

	// Создаем тестовый datagram
	testPayload := []byte("benchmark payload")
	datagram := make([]byte, 4+len(testPayload))
	copy(datagram[:4], "bench_target")
	copy(datagram[4:], testPayload)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		masqueServer.processDatagram(datagram, "127.0.0.1:12345")
	}
}
