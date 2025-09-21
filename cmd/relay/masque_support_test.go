package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/twogc/cloudbridge-relay-installer/config"
	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
)

func TestMASQUEServer_Creation(t *testing.T) {
	var testLogger *logger.Logger
	var err error
	testLogger, err = logger.New(logger.Config{Level: logger.DEBUG, JSON: true})
	require.NoError(t, err)
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

	server := NewMASQUEServer(testLogger, cfg)
	assert.NotNil(t, server)
	assert.NotNil(t, server.udpProxy)
	assert.NotNil(t, server.ipProxy)
	assert.NotNil(t, server.connections)
}

func TestMASQUEConnection_SendReceiveDatagram(t *testing.T) {
	var testLogger *logger.Logger
	var err error
	testLogger, err = logger.New(logger.Config{Level: logger.DEBUG, JSON: true})
	require.NoError(t, err)
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

	_ = NewMASQUEServer(testLogger, cfg)

	// Создаем тестовое соединение
	conn := &MASQUEConnection{
		ID:              "test_conn",
		DatagramChannel: make(chan []byte, 10),
		CreatedAt:       time.Now(),
		LastActivity:    time.Now(),
	}

	// Тест отправки datagram
	testData := []byte("test datagram")
	err = conn.SendDatagram(testData)
	assert.NoError(t, err)

	// Тест получения datagram
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	receivedData, err := conn.ReceiveDatagram(ctx)
	assert.NoError(t, err)
	assert.Equal(t, testData, receivedData)
}

func TestMASQUEConnection_DatagramChannelFull(t *testing.T) {
	var testLogger *logger.Logger
	var err error
	testLogger, err = logger.New(logger.Config{Level: logger.DEBUG, JSON: true})
	require.NoError(t, err)
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

	_ = NewMASQUEServer(testLogger, cfg)

	// Создаем соединение с маленьким буфером
	conn := &MASQUEConnection{
		ID:              "test_conn",
		DatagramChannel: make(chan []byte, 1), // Буфер на 1 элемент
		CreatedAt:       time.Now(),
		LastActivity:    time.Now(),
	}

	// Заполняем буфер
	err = conn.SendDatagram([]byte("first"))
	assert.NoError(t, err)

	// Пытаемся отправить еще один - должен вернуть ошибку
	err = conn.SendDatagram([]byte("second"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "datagram channel is full")
}

func TestMASQUEConnection_ReceiveDatagramTimeout(t *testing.T) {
	var testLogger *logger.Logger
	var err error
	testLogger, err = logger.New(logger.Config{Level: logger.DEBUG, JSON: true})
	require.NoError(t, err)
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

	_ = NewMASQUEServer(testLogger, cfg)

	conn := &MASQUEConnection{
		ID:              "test_conn",
		DatagramChannel: make(chan []byte, 10),
		CreatedAt:       time.Now(),
		LastActivity:    time.Now(),
	}

	// Тест таймаута при получении datagram
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err = conn.ReceiveDatagram(ctx)
	assert.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)
}

func TestMASQUEServer_HTTPDatagramsHandler(t *testing.T) {
	var testLogger *logger.Logger
	var err error
	testLogger, err = logger.New(logger.Config{Level: logger.DEBUG, JSON: true})
	require.NoError(t, err)
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

	server := NewMASQUEServer(testLogger, cfg)

	// Создаем тестовый HTTP запрос
	req := httptest.NewRequest("GET", "/.well-known/masque/datagrams/", nil)
	req.Header.Set("Capsule-Protocol", "?1")

	w := httptest.NewRecorder()

	// Вызываем обработчик
	server.handleHTTPDatagrams(w, req)

	// Проверяем ответ
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "?1", w.Header().Get("Capsule-Protocol"))
}

func TestMASQUEServer_HTTPDatagramsHandler_MissingHeader(t *testing.T) {
	var testLogger *logger.Logger
	var err error
	testLogger, err = logger.New(logger.Config{Level: logger.DEBUG, JSON: true})
	require.NoError(t, err)
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

	server := NewMASQUEServer(testLogger, cfg)

	// Создаем тестовый HTTP запрос без Capsule-Protocol заголовка
	req := httptest.NewRequest("GET", "/.well-known/masque/datagrams/", nil)
	w := httptest.NewRecorder()

	// Вызываем обработчик
	server.handleHTTPDatagrams(w, req)

	// Проверяем ошибку
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Capsule-Protocol header required")
}

func TestMASQUEServer_HealthHandler(t *testing.T) {
	var testLogger *logger.Logger
	var err error
	testLogger, err = logger.New(logger.Config{Level: logger.DEBUG, JSON: true})
	require.NoError(t, err)
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

	server := NewMASQUEServer(testLogger, cfg)

	// Создаем тестовый HTTP запрос
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// Вызываем обработчик
	server.handleHealth(w, req)

	// Проверяем ответ
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Body.String(), "healthy")
	assert.Contains(t, w.Body.String(), "masque_enabled")
	assert.Contains(t, w.Body.String(), "connections")
}

func TestMASQUEServer_GetConnectionCount(t *testing.T) {
	var testLogger *logger.Logger
	var err error
	testLogger, err = logger.New(logger.Config{Level: logger.DEBUG, JSON: true})
	require.NoError(t, err)
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

	server := NewMASQUEServer(testLogger, cfg)

	// Изначально соединений нет
	count := server.GetConnectionCount()
	assert.Equal(t, 0, count)

	// Добавляем тестовое соединение
	conn := &MASQUEConnection{
		ID:              "test_conn",
		DatagramChannel: make(chan []byte, 10),
		CreatedAt:       time.Now(),
		LastActivity:    time.Now(),
	}

	server.connectionsMutex.Lock()
	server.connections[conn.ID] = conn
	server.connectionsMutex.Unlock()

	// Проверяем количество соединений
	count = server.GetConnectionCount()
	assert.Equal(t, 1, count)
}

func TestMASQUEServer_GetConnectionInfo(t *testing.T) {
	var testLogger *logger.Logger
	var err error
	testLogger, err = logger.New(logger.Config{Level: logger.DEBUG, JSON: true})
	require.NoError(t, err)
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

	server := NewMASQUEServer(testLogger, cfg)

	// Добавляем тестовое соединение
	conn := &MASQUEConnection{
		ID:              "test_conn",
		DatagramChannel: make(chan []byte, 10),
		CreatedAt:       time.Now(),
		LastActivity:    time.Now(),
	}

	server.connectionsMutex.Lock()
	server.connections[conn.ID] = conn
	server.connectionsMutex.Unlock()

	// Получаем информацию о соединениях
	info := server.GetConnectionInfo()

	assert.Equal(t, 1, info["total_connections"])
	assert.Equal(t, true, info["is_running"])

	connections, ok := info["connections"].([]map[string]interface{})
	require.True(t, ok)
	require.Len(t, connections, 1)

	assert.Equal(t, "test_conn", connections[0]["id"])
}

func TestMASQUEServer_Close(t *testing.T) {
	var testLogger *logger.Logger
	var err error
	testLogger, err = logger.New(logger.Config{Level: logger.DEBUG, JSON: true})
	require.NoError(t, err)
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

	server := NewMASQUEServer(testLogger, cfg)

	// Добавляем тестовое соединение
	conn := &MASQUEConnection{
		ID:              "test_conn",
		DatagramChannel: make(chan []byte, 10),
		CreatedAt:       time.Now(),
		LastActivity:    time.Now(),
	}

	server.connectionsMutex.Lock()
	server.connections[conn.ID] = conn
	server.connectionsMutex.Unlock()

	// Закрываем сервер
	err = server.Close()
	assert.NoError(t, err)

	// Проверяем, что соединения очищены
	count := server.GetConnectionCount()
	assert.Equal(t, 0, count)
}

func TestMASQUEUDPProxy_HandleConnection(t *testing.T) {
	var testLogger *logger.Logger
	var err error
	testLogger, err = logger.New(logger.Config{Level: logger.DEBUG, JSON: true})
	require.NoError(t, err)
	proxy := &MASQUEUDPProxy{logger: testLogger}

	// Создаем тестовое соединение
	conn := &MASQUEConnection{
		ID:              "test_conn",
		DatagramChannel: make(chan []byte, 10),
		CreatedAt:       time.Now(),
		LastActivity:    time.Now(),
	}

	// Создаем тестовый HTTP response writer
	w := httptest.NewRecorder()

	// Вызываем обработчик
	err = proxy.HandleConnection(w, conn)
	assert.NoError(t, err)

	// Проверяем заголовки ответа
	assert.Equal(t, "?1", w.Header().Get("Capsule-Protocol"))
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMASQUEIPProxy_HandleConnection(t *testing.T) {
	var testLogger *logger.Logger
	var err error
	testLogger, err = logger.New(logger.Config{Level: logger.DEBUG, JSON: true})
	require.NoError(t, err)
	proxy := &MASQUEIPProxy{logger: testLogger}

	// Создаем тестовое соединение
	conn := &MASQUEConnection{
		ID:              "test_conn",
		DatagramChannel: make(chan []byte, 10),
		CreatedAt:       time.Now(),
		LastActivity:    time.Now(),
	}

	// Создаем тестовый HTTP response writer
	w := httptest.NewRecorder()

	// Вызываем обработчик
	err = proxy.HandleConnection(w, conn)
	assert.NoError(t, err)

	// Проверяем заголовки ответа
	assert.Equal(t, "?1", w.Header().Get("Capsule-Protocol"))
	assert.Equal(t, http.StatusOK, w.Code)
}

// Benchmark тесты
func BenchmarkMASQUEConnection_SendDatagram(b *testing.B) {
	conn := &MASQUEConnection{
		ID:              "bench_conn",
		DatagramChannel: make(chan []byte, 1000),
		CreatedAt:       time.Now(),
		LastActivity:    time.Now(),
	}

	testData := []byte("benchmark datagram")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conn.SendDatagram(testData)
	}
}

func BenchmarkMASQUEConnection_ReceiveDatagram(b *testing.B) {
	conn := &MASQUEConnection{
		ID:              "bench_conn",
		DatagramChannel: make(chan []byte, 1000),
		CreatedAt:       time.Now(),
		LastActivity:    time.Now(),
	}

	testData := []byte("benchmark datagram")

	// Заполняем канал данными
	for i := 0; i < b.N; i++ {
		conn.SendDatagram(testData)
	}

	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conn.ReceiveDatagram(ctx)
	}
}
