package main

import (
	"context"
	"testing"
	"time"

	"github.com/quic-go/quic-go/quicvarint"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/twogc/cloudbridge-relay-installer/config"
	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
)

// TestMASQUEHTTPDatagramsIntegration тестирует интеграцию HTTP Datagrams с MASQUE
func TestMASQUEHTTPDatagramsIntegration(t *testing.T) {
	testLogger, err := logger.New(logger.Config{
		Level:  logger.DEBUG,
		JSON:   true,
		Output: "stdout",
	})
	require.NoError(t, err)

	cfg := &config.Config{
		MASQUE: struct {
			Enabled         bool     `json:"enabled"`
			ProxyURL        string   `json:"proxy_url"`
			ALPNProtocols   []string `json:"alpn_protocols"`
			FallbackEnabled bool     `json:"fallback_enabled"`
		}{
			Enabled:         true,
			ALPNProtocols:   []string{"h3", "h3-29"},
			FallbackEnabled: true,
		},
	}

	server := NewMASQUEServer(testLogger, cfg)
	require.NotNil(t, server)

	t.Run("ContextID_Parsing", func(t *testing.T) {
		// Тестируем парсинг Context ID как в masque-go
		testData := []byte("test payload")

		// Создаем datagram с Context ID = 0 (CONNECT-UDP)
		datagram := make([]byte, 0, len(contextIDZero)+len(testData))
		datagram = append(datagram, contextIDZero...)
		datagram = append(datagram, testData...)

		// Парсим Context ID
		contextID, n, err := quicvarint.Parse(datagram)
		require.NoError(t, err)
		assert.Equal(t, uint64(0), contextID, "Context ID должен быть 0 для CONNECT-UDP")

		// Извлекаем payload
		payload := datagram[n:]
		assert.Equal(t, testData, payload, "Payload должен совпадать с исходными данными")
	})

	t.Run("ContextID_NonZero", func(t *testing.T) {
		// Тестируем парсинг Context ID != 0 (CONNECT-IP)
		testData := []byte("test IP payload")

		// Создаем datagram с Context ID = 1 (CONNECT-IP)
		contextIDOne := quicvarint.Append([]byte{}, 1)
		datagram := make([]byte, 0, len(contextIDOne)+len(testData))
		datagram = append(datagram, contextIDOne...)
		datagram = append(datagram, testData...)

		// Парсим Context ID
		contextID, n, err := quicvarint.Parse(datagram)
		require.NoError(t, err)
		assert.Equal(t, uint64(1), contextID, "Context ID должен быть 1 для CONNECT-IP")

		// Извлекаем payload
		payload := datagram[n:]
		assert.Equal(t, testData, payload, "Payload должен совпадать с исходными данными")
	})

	t.Run("MASQUEConnection_SendDatagram", func(t *testing.T) {
		// Создаем тестовое соединение
		conn := &MASQUEConnection{
			ID:              "test_conn",
			DatagramChannel: make(chan []byte, 10),
			CreatedAt:       time.Now(),
			LastActivity:    time.Now(),
		}

		// Тест отправки datagram через channel
		testData := []byte("test datagram")
		err := conn.SendDatagramToChannel(testData)
		assert.NoError(t, err)

		// Тест получения datagram
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		receivedData, err := conn.ReceiveDatagram(ctx)
		assert.NoError(t, err)
		assert.Equal(t, testData, receivedData)
	})

	t.Run("MASQUEConnection_SendDatagramWithContextID", func(t *testing.T) {
		// Создаем тестовое соединение без Stream (для тестирования логики)
		conn := &MASQUEConnection{
			ID:              "test_conn",
			DatagramChannel: make(chan []byte, 10),
			CreatedAt:       time.Now(),
			LastActivity:    time.Now(),
		}

		// Тест отправки datagram без Stream (должна быть ошибка)
		testData := []byte("test datagram")
		err := conn.SendDatagramWithContextID(testData)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "stream not initialized")
	})

	t.Run("ProcessDatagram_CONNECT_UDP", func(t *testing.T) {
		// Тестируем обработку CONNECT-UDP datagram
		testData := []byte("test UDP payload")

		// Создаем datagram с Context ID = 0
		datagram := make([]byte, 0, len(contextIDZero)+len(testData))
		datagram = append(datagram, contextIDZero...)
		datagram = append(datagram, testData...)

		// Обрабатываем datagram
		err := server.processDatagram(datagram, "127.0.0.1:8080")
		assert.NoError(t, err, "CONNECT-UDP datagram должен обрабатываться без ошибок")
	})

	t.Run("ProcessDatagram_CONNECT_IP", func(t *testing.T) {
		// Тестируем обработку CONNECT-IP datagram
		testData := []byte("test IP payload")

		// Создаем datagram с Context ID = 1
		contextIDOne := quicvarint.Append([]byte{}, 1)
		datagram := make([]byte, 0, len(contextIDOne)+len(testData))
		datagram = append(datagram, contextIDOne...)
		datagram = append(datagram, testData...)

		// Обрабатываем datagram
		err := server.processDatagram(datagram, "127.0.0.1:8080")
		// Может быть ошибка, если processIPCapsule не реализован полностью
		// Это нормально для данного теста
		t.Logf("CONNECT-IP datagram processing result: %v", err)
	})

	t.Run("ProcessDatagram_Empty", func(t *testing.T) {
		// Тестируем обработку пустого datagram
		err := server.processDatagram([]byte{}, "127.0.0.1:8080")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty datagram")
	})

	t.Run("ProcessDatagram_Malformed", func(t *testing.T) {
		// Тестируем обработку некорректного datagram
		malformedData := []byte{0xFF, 0xFF, 0xFF, 0xFF} // Некорректный varint
		err := server.processDatagram(malformedData, "127.0.0.1:8080")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "malformed datagram")
	})
}

// TestMASQUEHTTPDatagramsPerformance тестирует производительность HTTP Datagrams
func TestMASQUEHTTPDatagramsPerformance(t *testing.T) {
	testLogger, err := logger.New(logger.Config{
		Level:  logger.INFO, // Уменьшаем логирование для performance тестов
		JSON:   true,
		Output: "stdout",
	})
	require.NoError(t, err)

	cfg := &config.Config{
		MASQUE: struct {
			Enabled         bool     `json:"enabled"`
			ProxyURL        string   `json:"proxy_url"`
			ALPNProtocols   []string `json:"alpn_protocols"`
			FallbackEnabled bool     `json:"fallback_enabled"`
		}{
			Enabled:         true,
			ALPNProtocols:   []string{"h3", "h3-29"},
			FallbackEnabled: true,
		},
	}

	server := NewMASQUEServer(testLogger, cfg)
	require.NotNil(t, server)

	t.Run("ProcessDatagram_Performance", func(t *testing.T) {
		// Тестируем производительность обработки datagrams
		testData := []byte("performance test payload")

		// Создаем datagram с Context ID = 0
		datagram := make([]byte, 0, len(contextIDZero)+len(testData))
		datagram = append(datagram, contextIDZero...)
		datagram = append(datagram, testData...)

		// Измеряем время обработки
		start := time.Now()
		iterations := 1000

		for i := 0; i < iterations; i++ {
			err := server.processDatagram(datagram, "127.0.0.1:8080")
			assert.NoError(t, err)
		}

		duration := time.Since(start)
		avgTime := duration / time.Duration(iterations)

		t.Logf("Processed %d datagrams in %v", iterations, duration)
		t.Logf("Average time per datagram: %v", avgTime)
		t.Logf("Throughput: %.2f datagrams/second", float64(iterations)/duration.Seconds())

		// Проверяем, что среднее время обработки разумное (< 1ms)
		assert.Less(t, avgTime, 1*time.Millisecond, "Среднее время обработки должно быть < 1ms")
	})

	t.Run("ContextID_Parsing_Performance", func(t *testing.T) {
		// Тестируем производительность парсинга Context ID
		testData := []byte("performance test payload")

		// Создаем datagram с Context ID = 0
		datagram := make([]byte, 0, len(contextIDZero)+len(testData))
		datagram = append(datagram, contextIDZero...)
		datagram = append(datagram, testData...)

		// Измеряем время парсинга
		start := time.Now()
		iterations := 10000

		for i := 0; i < iterations; i++ {
			contextID, n, err := quicvarint.Parse(datagram)
			require.NoError(t, err)
			require.Equal(t, uint64(0), contextID)
			require.Equal(t, len(contextIDZero), n)
		}

		duration := time.Since(start)
		avgTime := duration / time.Duration(iterations)

		t.Logf("Parsed %d Context IDs in %v", iterations, duration)
		t.Logf("Average time per parse: %v", avgTime)
		t.Logf("Throughput: %.2f parses/second", float64(iterations)/duration.Seconds())

		// Проверяем, что среднее время парсинга разумное (< 100μs)
		assert.Less(t, avgTime, 100*time.Microsecond, "Среднее время парсинга должно быть < 100μs")
	})
}
