package main

import (
	"context"
	"crypto/tls"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/twogc/cloudbridge-relay-installer/config"
	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
)

// TestMASQUEHTTP3ServerConfiguration проверяет правильность конфигурации HTTP/3 сервера
func TestMASQUEHTTP3ServerConfiguration(t *testing.T) {
	testLogger, err := logger.New(logger.Config{Level: logger.DEBUG, JSON: true, Output: "stdout"})
	require.NoError(t, err)

	cfg := &config.Config{
		MASQUE: struct {
			Enabled         bool     `json:"enabled"`
			ProxyURL        string   `json:"proxy_url"`
			ALPNProtocols   []string `json:"alpn_protocols"`
			FallbackEnabled bool     `json:"fallback_enabled"`
		}{
			Enabled:       true,
			ALPNProtocols: []string{"h3", "h3-29", "h3-28", "h3-27"},
		},
	}

	server := NewMASQUEServer(testLogger, cfg)
	require.NotNil(t, server)

	// Проверяем, что сервер создается с правильной конфигурацией
	assert.True(t, server.config.MASQUE.Enabled)
	assert.Equal(t, []string{"h3", "h3-29", "h3-28", "h3-27"}, server.config.MASQUE.ALPNProtocols)
}

// TestMASQUEHTTP3ServerStart проверяет запуск HTTP/3 сервера
func TestMASQUEHTTP3ServerStart(t *testing.T) {
	testLogger, err := logger.New(logger.Config{Level: logger.DEBUG, JSON: true, Output: "stdout"})
	require.NoError(t, err)

	cfg := &config.Config{
		MASQUE: struct {
			Enabled         bool     `json:"enabled"`
			ProxyURL        string   `json:"proxy_url"`
			ALPNProtocols   []string `json:"alpn_protocols"`
			FallbackEnabled bool     `json:"fallback_enabled"`
		}{
			Enabled:       true,
			ALPNProtocols: []string{"h3", "h3-29", "h3-28", "h3-27"},
		},
	}

	server := NewMASQUEServer(testLogger, cfg)
	require.NotNil(t, server)

	// Запускаем сервер
	err = server.Start()
	require.NoError(t, err)

	// Проверяем, что сервер запущен
	assert.True(t, server.isRunning)
	assert.NotNil(t, server.http3Server)

	// Проверяем конфигурацию HTTP/3 сервера
	assert.True(t, server.http3Server.EnableDatagrams, "HTTP/3 server должен иметь включенные datagrams")
	assert.NotNil(t, server.http3Server.QUICConfig, "QUIC config должен быть установлен")
	assert.True(t, server.http3Server.QUICConfig.EnableDatagrams, "QUIC config должен иметь включенные datagrams")

	// Останавливаем сервер
	err = server.Close()
	require.NoError(t, err)
}

// TestMASQUEHTTP3ServerTLSConfig проверяет TLS конфигурацию
func TestMASQUEHTTP3ServerTLSConfig(t *testing.T) {
	testLogger, err := logger.New(logger.Config{Level: logger.DEBUG, JSON: true, Output: "stdout"})
	require.NoError(t, err)

	cfg := &config.Config{
		MASQUE: struct {
			Enabled         bool     `json:"enabled"`
			ProxyURL        string   `json:"proxy_url"`
			ALPNProtocols   []string `json:"alpn_protocols"`
			FallbackEnabled bool     `json:"fallback_enabled"`
		}{
			Enabled:       true,
			ALPNProtocols: []string{"h3", "h3-29", "h3-28", "h3-27"},
		},
	}

	server := NewMASQUEServer(testLogger, cfg)
	require.NotNil(t, server)

	// Создаем TLS конфигурацию
	tlsConfig, err := server.createTLSConfig()
	require.NoError(t, err)
	require.NotNil(t, tlsConfig)

	// Проверяем TLS конфигурацию
	assert.Equal(t, tls.VersionTLS12, tlsConfig.MinVersion, "Минимальная версия TLS должна быть 1.2")
	assert.NotEmpty(t, tlsConfig.CipherSuites, "Cipher suites должны быть настроены")
	assert.NotEmpty(t, tlsConfig.Certificates, "Сертификаты должны быть загружены")
	assert.Equal(t, []string{"h3", "h3-29", "h3-28", "h3-27"}, tlsConfig.NextProtos, "ALPN протоколы должны быть настроены")
}

// TestMASQUEHTTP3ServerHandlers проверяет HTTP обработчики
func TestMASQUEHTTP3ServerHandlers(t *testing.T) {
	testLogger, err := logger.New(logger.Config{Level: logger.DEBUG, JSON: true, Output: "stdout"})
	require.NoError(t, err)

	cfg := &config.Config{
		MASQUE: struct {
			Enabled         bool     `json:"enabled"`
			ProxyURL        string   `json:"proxy_url"`
			ALPNProtocols   []string `json:"alpn_protocols"`
			FallbackEnabled bool     `json:"fallback_enabled"`
		}{
			Enabled:       true,
			ALPNProtocols: []string{"h3", "h3-29", "h3-28", "h3-27"},
		},
	}

	server := NewMASQUEServer(testLogger, cfg)
	require.NotNil(t, server)

	// Создаем HTTP обработчик
	handler := server.createHTTPHandler()
	require.NotNil(t, handler)

	// Проверяем, что обработчик создан
	assert.NotNil(t, handler)
}

// TestMASQUEConnectionDatagramSupport проверяет поддержку datagrams в соединениях
func TestMASQUEConnectionDatagramSupport(t *testing.T) {
	testLogger, err := logger.New(logger.Config{Level: logger.DEBUG, JSON: true, Output: "stdout"})
	require.NoError(t, err)

	cfg := &config.Config{
		MASQUE: struct {
			Enabled         bool     `json:"enabled"`
			ProxyURL        string   `json:"proxy_url"`
			ALPNProtocols   []string `json:"alpn_protocols"`
			FallbackEnabled bool     `json:"fallback_enabled"`
		}{
			Enabled:       true,
			ALPNProtocols: []string{"h3", "h3-29", "h3-28", "h3-27"},
		},
	}

	server := NewMASQUEServer(testLogger, cfg)
	require.NotNil(t, server)

	// Создаем тестовое соединение
	conn := &MASQUEConnection{
		ID:              "test_conn",
		DatagramChannel: make(chan []byte, 10),
		CreatedAt:       time.Now(),
		LastActivity:    time.Now(),
	}

	// Тестируем отправку datagram
	testData := []byte("test datagram")
	err = conn.SendDatagram(testData)
	assert.NoError(t, err)

	// Тестируем получение datagram
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	receivedData, err := conn.ReceiveDatagram(ctx)
	assert.NoError(t, err)
	assert.Equal(t, testData, receivedData)
}

// TestMASQUEHTTP3ServerIntegration проверяет интеграцию HTTP/3 сервера
func TestMASQUEHTTP3ServerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Пропускаем интеграционный тест в коротком режиме")
	}

	testLogger, err := logger.New(logger.Config{Level: logger.DEBUG, JSON: true, Output: "stdout"})
	require.NoError(t, err)

	cfg := &config.Config{
		MASQUE: struct {
			Enabled         bool     `json:"enabled"`
			ProxyURL        string   `json:"proxy_url"`
			ALPNProtocols   []string `json:"alpn_protocols"`
			FallbackEnabled bool     `json:"fallback_enabled"`
		}{
			Enabled:       true,
			ALPNProtocols: []string{"h3", "h3-29", "h3-28", "h3-27"},
		},
	}

	server := NewMASQUEServer(testLogger, cfg)
	require.NotNil(t, server)

	// Запускаем сервер
	err = server.Start()
	require.NoError(t, err)
	defer server.Close()

	// Ждем немного, чтобы сервер запустился
	time.Sleep(100 * time.Millisecond)

	// Проверяем, что сервер запущен
	assert.True(t, server.isRunning)

	// Проверяем конфигурацию HTTP/3 сервера
	assert.NotNil(t, server.http3Server)
	assert.True(t, server.http3Server.EnableDatagrams)
	assert.NotNil(t, server.http3Server.QUICConfig)
	assert.True(t, server.http3Server.QUICConfig.EnableDatagrams)
}

// BenchmarkMASQUEConnectionDatagramThroughput тестирует производительность datagrams
func BenchmarkMASQUEConnectionDatagramThroughput(b *testing.B) {
	conn := &MASQUEConnection{
		ID:              "bench_conn",
		DatagramChannel: make(chan []byte, 1000),
		CreatedAt:       time.Now(),
		LastActivity:    time.Now(),
	}

	testData := []byte("benchmark datagram data for throughput testing")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := conn.SendDatagram(testData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkMASQUEConnectionDatagramLatency тестирует задержку datagrams
func BenchmarkMASQUEConnectionDatagramLatency(b *testing.B) {
	conn := &MASQUEConnection{
		ID:              "bench_conn",
		DatagramChannel: make(chan []byte, 1000),
		CreatedAt:       time.Now(),
		LastActivity:    time.Now(),
	}

	testData := []byte("benchmark datagram data for latency testing")

	// Заполняем канал данными
	for i := 0; i < b.N; i++ {
		conn.SendDatagram(testData)
	}

	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := conn.ReceiveDatagram(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}
