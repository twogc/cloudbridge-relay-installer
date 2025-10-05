package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestNewQUICMetrics(t *testing.T) {
	// Создаем новый регистр для тестов
	reg := prometheus.NewRegistry()

	// Создаем QUIC метрики
	quicMetrics := NewQUICMetrics(reg)

	// Проверяем, что метрики созданы
	assert.NotNil(t, quicMetrics)
	assert.NotNil(t, quicMetrics.connectionsTotal)
	assert.NotNil(t, quicMetrics.connectionsActive)
	assert.NotNil(t, quicMetrics.streamsTotal)
	assert.NotNil(t, quicMetrics.streamsActive)
	assert.NotNil(t, quicMetrics.handshakeDuration)
}

func TestQUICConnectionMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	quicMetrics := NewQUICMetrics(reg)

	// Тестируем инкремент соединений
	quicMetrics.IncConnections()
	quicMetrics.IncConnections()

	// Тестируем декремент соединений
	quicMetrics.DecConnections()

	// Тестируем ошибки соединений
	quicMetrics.IncConnectionsFailed("timeout", "v1")
	quicMetrics.IncConnectionsFailed("handshake_failed", "v1")
	quicMetrics.IncConnectionsDropped("network_error")

	// Проверяем, что метрики зарегистрированы
	metrics, err := reg.Gather()
	assert.NoError(t, err)
	assert.Greater(t, len(metrics), 0)
}

func TestQUICStreamMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	quicMetrics := NewQUICMetrics(reg)

	// Тестируем инкремент потоков
	quicMetrics.IncStreams()
	quicMetrics.IncStreams()

	// Тестируем декремент потоков
	quicMetrics.DecStreams()

	// Тестируем создание потоков
	quicMetrics.IncStreamsCreated("unidirectional", "incoming")
	quicMetrics.IncStreamsCreated("bidirectional", "outgoing")

	// Тестируем закрытие потоков
	quicMetrics.IncStreamsClosed("unidirectional", "normal")
	quicMetrics.IncStreamsClosed("bidirectional", "error")

	// Тестируем сброс потоков
	quicMetrics.IncStreamsReset("bidirectional", "timeout")

	// Проверяем, что метрики зарегистрированы
	metrics, err := reg.Gather()
	assert.NoError(t, err)
	assert.Greater(t, len(metrics), 0)
}

func TestQUICPerformanceMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	quicMetrics := NewQUICMetrics(reg)

	// Тестируем длительность handshake
	quicMetrics.ObserveHandshakeDuration(0.1)
	quicMetrics.ObserveHandshakeDuration(0.5)
	quicMetrics.ObserveHandshakeDuration(1.0)

	// Тестируем ошибки handshake
	quicMetrics.IncHandshakeFailures("timeout", "v1")
	quicMetrics.IncHandshakeFailures("crypto_error", "v1")

	// Тестируем латентность соединения
	quicMetrics.ObserveConnectionLatency("conn1", 0.05)
	quicMetrics.ObserveConnectionLatency("conn2", 0.1)

	// Тестируем пропускную способность
	quicMetrics.IncThroughputBytes("conn1", "incoming", 1024)
	quicMetrics.IncThroughputBytes("conn1", "outgoing", 2048)
	quicMetrics.IncThroughputPackets("conn1", "incoming", 10)
	quicMetrics.IncThroughputPackets("conn1", "outgoing", 15)

	// Тестируем потерю пакетов
	quicMetrics.SetPacketLossRate("conn1", 0.01)
	quicMetrics.IncPacketLossEvents("conn1", "network_congestion")

	// Проверяем, что метрики зарегистрированы
	metrics, err := reg.Gather()
	assert.NoError(t, err)
	assert.Greater(t, len(metrics), 0)
}

func TestQUICConnectionQualityMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	quicMetrics := NewQUICMetrics(reg)

	// Тестируем RTT
	quicMetrics.ObserveRTT("conn1", 50.0)
	quicMetrics.ObserveRTT("conn1", 75.0)
	quicMetrics.ObserveRTT("conn2", 100.0)

	// Тестируем вариацию RTT
	quicMetrics.SetRTTVariance("conn1", 10.0)
	quicMetrics.SetRTTVariance("conn2", 15.0)

	// Тестируем окно перегрузки
	quicMetrics.SetCongestionWindow("conn1", 1024)
	quicMetrics.SetCongestionWindow("conn2", 2048)

	// Тестируем байты в полете
	quicMetrics.SetBytesInFlight("conn1", 512)
	quicMetrics.SetBytesInFlight("conn2", 1024)

	// Тестируем пакеты в полете
	quicMetrics.SetPacketsInFlight("conn1", 5)
	quicMetrics.SetPacketsInFlight("conn2", 10)

	// Проверяем, что метрики зарегистрированы
	metrics, err := reg.Gather()
	assert.NoError(t, err)
	assert.Greater(t, len(metrics), 0)
}

func TestQUICErrorMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	quicMetrics := NewQUICMetrics(reg)

	// Тестируем ошибки пакетов
	quicMetrics.IncPacketErrors("corruption", "conn1")
	quicMetrics.IncPacketErrors("duplicate", "conn2")

	// Тестируем ошибки соединений
	quicMetrics.IncConnectionErrors("timeout", "conn1")
	quicMetrics.IncConnectionErrors("reset", "conn2")

	// Тестируем ошибки потоков
	quicMetrics.IncStreamErrors("reset", "stream1")
	quicMetrics.IncStreamErrors("timeout", "stream2")

	// Тестируем события таймаута
	quicMetrics.IncTimeoutEvents("idle", "conn1")
	quicMetrics.IncTimeoutEvents("handshake", "conn2")

	// Тестируем ретрансмиссии
	quicMetrics.IncRetransmissions("conn1", "packet_loss")
	quicMetrics.IncRetransmissions("conn2", "timeout")

	// Проверяем, что метрики зарегистрированы
	metrics, err := reg.Gather()
	assert.NoError(t, err)
	assert.Greater(t, len(metrics), 0)
}

func TestQUICSecurityMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	quicMetrics := NewQUICMetrics(reg)

	// Тестируем ошибки шифрования
	quicMetrics.IncEncryptionFailures("key_rotation", "conn1")
	quicMetrics.IncEncryptionFailures("cipher_error", "conn2")

	// Тестируем ошибки аутентификации
	quicMetrics.IncAuthenticationFailures("token_invalid", "conn1")
	quicMetrics.IncAuthenticationFailures("signature_error", "conn2")

	// Тестируем события ротации ключей
	quicMetrics.IncKeyRotationEvents("conn1", "initial")
	quicMetrics.IncKeyRotationEvents("conn2", "handshake")

	// Проверяем, что метрики зарегистрированы
	metrics, err := reg.Gather()
	assert.NoError(t, err)
	assert.Greater(t, len(metrics), 0)
}

func TestQUICConfigurationMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	quicMetrics := NewQUICMetrics(reg)

	// Тестируем использование версий
	quicMetrics.SetVersionUsage("v1", 5)
	quicMetrics.SetVersionUsage("draft-29", 3)

	// Тестируем переговоры параметров
	quicMetrics.IncParameterNegotiation("max_streams", "success")
	quicMetrics.IncParameterNegotiation("initial_window", "failure")

	// Тестируем использование расширений
	quicMetrics.IncExtensionUsage("multipath", "conn1")
	quicMetrics.IncExtensionUsage("0rtt", "conn2")

	// Проверяем, что метрики зарегистрированы
	metrics, err := reg.Gather()
	assert.NoError(t, err)
	assert.Greater(t, len(metrics), 0)
}

func TestQUICResourceMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	quicMetrics := NewQUICMetrics(reg)

	// Тестируем использование памяти
	quicMetrics.SetMemoryUsage("conn1", "connection", 1024*1024)
	quicMetrics.SetMemoryUsage("conn1", "streams", 512*1024)

	// Тестируем использование CPU
	quicMetrics.SetCPUUsage("conn1", "connection", 25.5)
	quicMetrics.SetCPUUsage("conn1", "crypto", 15.2)

	// Тестируем количество горутин
	quicMetrics.SetGoroutineCount("conn1", "connection", 10)
	quicMetrics.SetGoroutineCount("conn1", "streams", 5)

	// Проверяем, что метрики зарегистрированы
	metrics, err := reg.Gather()
	assert.NoError(t, err)
	assert.Greater(t, len(metrics), 0)
}

func TestQUICMetricsIntegration(t *testing.T) {
	// Создаем основной Metrics с QUIC метриками
	reg := prometheus.NewRegistry()
	mainMetrics := NewMetrics(reg)

	// Проверяем, что QUIC метрики доступны
	quicMetrics := mainMetrics.GetQUICMetrics()
	assert.NotNil(t, quicMetrics)

	// Тестируем методы основного Metrics
	mainMetrics.IncQUICConnections()
	mainMetrics.DecQUICConnections()
	mainMetrics.IncQUICConnectionsFailed("test_reason", "v1")
	mainMetrics.IncQUICStreams()
	mainMetrics.DecQUICStreams()
	mainMetrics.ObserveQUICHandshakeDuration(0.1)
	mainMetrics.ObserveQUICRTT("conn1", 50.0)
	mainMetrics.SetQUICPacketLossRate("conn1", 0.01)
	mainMetrics.IncQUICPacketErrors("test_error", "conn1")
	mainMetrics.IncQUICEncryptionFailures("test_failure", "conn1")
	mainMetrics.SetQUICVersionUsage("v1", 1)
	mainMetrics.SetQUICMemoryUsage("conn1", "test", 1024)

	// Проверяем, что метрики зарегистрированы
	metrics, err := reg.Gather()
	assert.NoError(t, err)
	assert.Greater(t, len(metrics), 0)
}
