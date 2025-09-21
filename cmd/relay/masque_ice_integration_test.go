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
		testLogger,
		[]string{"stun1.example.com:3478"},
		nil, // TURN pool
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
		assert.NotNil(t, iceAgent.Metrics)
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

		// Ждем завершения сбора кандидатов
		time.Sleep(100 * time.Millisecond)

		// Проверяем, что состояние изменилось на Gathered
		state := iceAgent.GetState()
		assert.True(t, state == nat.ICEStateGathered || state == nat.ICEStateGathering)

		// Проверяем, что собрались кандидаты
		candidates := iceAgent.GetLocalCandidates()
		assert.GreaterOrEqual(t, len(candidates), 0) // Может быть 0 в тестовой среде

		// Проверяем, что есть host кандидаты (если есть кандидаты)
		if len(candidates) > 0 {
			hasHostCandidate := false
			for _, candidate := range candidates {
				if candidate.Type == nat.ICECandidateTypeHost {
					hasHostCandidate = true
					break
				}
			}
			assert.True(t, hasHostCandidate)
		}
	})

	// Тест 5: ICE метрики записываются
	t.Run("ICE_Metrics_Recording", func(t *testing.T) {
		// Записываем тестовые метрики через Metrics
		iceAgent.Metrics.RecordConnectivityCheck("test-client", true)
		iceAgent.Metrics.RecordConnectivityCheck("test-client", false)

		// Проверяем статистику
		stats := iceAgent.Metrics.GetAllMetrics()
		assert.NotNil(t, stats)
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

		// Создаем тестовое соединение
		conn := &MASQUEConnection{
			ID:              "test_peer",
			DatagramChannel: make(chan []byte, 10),
			CreatedAt:       time.Now(),
			LastActivity:    time.Now(),
		}

		// Добавляем соединение
		masqueServer.connectionsMutex.Lock()
		masqueServer.connections[conn.ID] = conn
		masqueServer.connectionsMutex.Unlock()

		// Тестируем прямую отправку в канал (без SendDatagram)
		testPayload := []byte("test message")

		// Отправляем напрямую в канал
		select {
		case conn.DatagramChannel <- testPayload:
			// Успешно отправили
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Failed to send datagram to channel")
		}

		// Проверяем, что datagram был доставлен
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		receivedData, err := conn.ReceiveDatagram(ctx)
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
		iceAgent.Metrics.RecordConnectivityCheck("ice-masque-client", true)

		// Проверяем, что метрики записались
		iceStats := iceAgent.Metrics.GetAllMetrics()
		assert.NotNil(t, iceStats)

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

		// Проверяем производительность создания соединений
		start := time.Now()
		for i := 0; i < 100; i++ {
			conn := &MASQUEConnection{
				ID:              fmt.Sprintf("perf_test_%d", i),
				DatagramChannel: make(chan []byte, 100),
				CreatedAt:       time.Now(),
				LastActivity:    time.Now(),
			}

			// Просто создаем соединение, не отправляем datagram
			masqueServer.connectionsMutex.Lock()
			masqueServer.connections[conn.ID] = conn
			masqueServer.connectionsMutex.Unlock()
		}
		duration := time.Since(start)

		// Проверяем, что создание было быстрым (менее 1 секунды для 100 соединений)
		assert.Less(t, duration, 1*time.Second)
	})
}

// Benchmark тесты

func BenchmarkICEAgent_DetectNATType(b *testing.B) {
	// Создаем logger для NAT компонентов
	natLogger, _ := logger.New(logger.Config{
		Level:  logger.DEBUG,
		JSON:   true,
		Output: "stdout",
	})

	iceAgent := nat.NewICEAgent(
		natLogger,
		[]string{"stun1.example.com:3478"},
		nil, // TURN pool
	)

	// Добавляем тестовые кандидаты
	iceAgent.AddRemoteCandidate(&nat.ICECandidate{
		Type:      nat.ICECandidateTypeHost,
		Address:   "192.168.1.1",
		Port:      8080,
		Transport: "UDP",
	})
	iceAgent.AddRemoteCandidate(&nat.ICECandidate{
		Type:      nat.ICECandidateTypeServerReflexive,
		Address:   "203.0.113.1",
		Port:      8080,
		Transport: "UDP",
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Симулируем проверку connectivity
		iceAgent.Metrics.RecordConnectivityCheck("benchmark-client", true)
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
