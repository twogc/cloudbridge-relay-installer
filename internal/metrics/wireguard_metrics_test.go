package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestNewWireGuardMetrics(t *testing.T) {
	// Создаем новый регистр для тестов
	reg := prometheus.NewRegistry()

	// Создаем WireGuard метрики
	wireguardMetrics := NewWireGuardMetrics(reg)

	// Проверяем, что метрики созданы
	assert.NotNil(t, wireguardMetrics)
	assert.NotNil(t, wireguardMetrics.interfacesTotal)
	assert.NotNil(t, wireguardMetrics.interfacesActive)
	assert.NotNil(t, wireguardMetrics.peersTotal)
	assert.NotNil(t, wireguardMetrics.peersActive)
	assert.NotNil(t, wireguardMetrics.trafficBytesReceived)
	assert.NotNil(t, wireguardMetrics.handshakeDuration)
}

func TestWireGuardInterfaceMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	wireguardMetrics := NewWireGuardMetrics(reg)

	// Тестируем инкремент интерфейсов
	wireguardMetrics.IncInterfaces()
	wireguardMetrics.IncInterfaces()

	// Тестируем декремент интерфейсов
	wireguardMetrics.DecInterfaces()

	// Тестируем ошибки интерфейсов
	wireguardMetrics.IncInterfacesFailed("wg0", "creation_failed")
	wireguardMetrics.IncInterfacesFailed("wg1", "configuration_error")

	// Тестируем создание интерфейсов
	wireguardMetrics.IncInterfacesCreated("wg0", "main")
	wireguardMetrics.IncInterfacesCreated("wg1", "tenant")

	// Тестируем уничтожение интерфейсов
	wireguardMetrics.IncInterfacesDestroyed("wg0", "manual")
	wireguardMetrics.IncInterfacesDestroyed("wg1", "error")

	// Проверяем, что метрики зарегистрированы
	metrics, err := reg.Gather()
	assert.NoError(t, err)
	assert.Greater(t, len(metrics), 0)
}

func TestWireGuardPeerMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	wireguardMetrics := NewWireGuardMetrics(reg)

	// Тестируем инкремент пиров
	wireguardMetrics.IncPeers()
	wireguardMetrics.IncPeers()

	// Тестируем декремент пиров
	wireguardMetrics.DecPeers()

	// Тестируем подключенных пиров
	wireguardMetrics.SetPeersConnected("wg0", "peer1", 1)
	wireguardMetrics.SetPeersConnected("wg0", "peer2", 1)
	wireguardMetrics.SetPeersConnected("wg1", "peer3", 0)

	// Тестируем добавление пиров
	wireguardMetrics.IncPeersAdded("wg0", "peer1")
	wireguardMetrics.IncPeersAdded("wg0", "peer2")

	// Тестируем удаление пиров
	wireguardMetrics.IncPeersRemoved("wg0", "peer1", "manual")
	wireguardMetrics.IncPeersRemoved("wg0", "peer2", "timeout")

	// Тестируем ошибки пиров
	wireguardMetrics.IncPeersFailed("wg0", "peer1", "handshake_failed")
	wireguardMetrics.IncPeersFailed("wg1", "peer2", "connection_timeout")

	// Проверяем, что метрики зарегистрированы
	metrics, err := reg.Gather()
	assert.NoError(t, err)
	assert.Greater(t, len(metrics), 0)
}

func TestWireGuardTrafficMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	wireguardMetrics := NewWireGuardMetrics(reg)

	// Тестируем трафик байтов
	wireguardMetrics.IncTrafficBytesReceived("wg0", "peer1", 1024)
	wireguardMetrics.IncTrafficBytesReceived("wg0", "peer1", 2048)
	wireguardMetrics.IncTrafficBytesSent("wg0", "peer1", 512)
	wireguardMetrics.IncTrafficBytesSent("wg0", "peer1", 1024)

	// Тестируем трафик пакетов
	wireguardMetrics.IncTrafficPacketsReceived("wg0", "peer1", 10)
	wireguardMetrics.IncTrafficPacketsReceived("wg0", "peer1", 20)
	wireguardMetrics.IncTrafficPacketsSent("wg0", "peer1", 5)
	wireguardMetrics.IncTrafficPacketsSent("wg0", "peer1", 15)

	// Тестируем скорость трафика
	wireguardMetrics.SetTrafficRate("wg0", "peer1", "incoming", 1000.0)
	wireguardMetrics.SetTrafficRate("wg0", "peer1", "outgoing", 500.0)

	// Проверяем, что метрики зарегистрированы
	metrics, err := reg.Gather()
	assert.NoError(t, err)
	assert.Greater(t, len(metrics), 0)
}

func TestWireGuardPerformanceMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	wireguardMetrics := NewWireGuardMetrics(reg)

	// Тестируем длительность handshake
	wireguardMetrics.ObserveHandshakeDuration("wg0", "peer1", 0.1)
	wireguardMetrics.ObserveHandshakeDuration("wg0", "peer1", 0.2)
	wireguardMetrics.ObserveHandshakeDuration("wg0", "peer2", 0.15)

	// Тестируем латентность соединения
	wireguardMetrics.ObserveConnectionLatency("wg0", "peer1", 0.05)
	wireguardMetrics.ObserveConnectionLatency("wg0", "peer1", 0.08)
	wireguardMetrics.ObserveConnectionLatency("wg0", "peer2", 0.06)

	// Тестируем пропускную способность
	wireguardMetrics.IncThroughputBytes("wg0", "peer1", "incoming", 1024)
	wireguardMetrics.IncThroughputBytes("wg0", "peer1", "outgoing", 2048)
	wireguardMetrics.IncThroughputPackets("wg0", "peer1", "incoming", 10)
	wireguardMetrics.IncThroughputPackets("wg0", "peer1", "outgoing", 20)

	// Тестируем утилизацию полосы пропускания
	wireguardMetrics.SetBandwidthUtilization("wg0", "peer1", 75.5)
	wireguardMetrics.SetBandwidthUtilization("wg0", "peer2", 45.2)

	// Проверяем, что метрики зарегистрированы
	metrics, err := reg.Gather()
	assert.NoError(t, err)
	assert.Greater(t, len(metrics), 0)
}

func TestWireGuardConnectionQualityMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	wireguardMetrics := NewWireGuardMetrics(reg)

	// Тестируем интервал keepalive
	wireguardMetrics.SetKeepaliveInterval("wg0", "peer1", 25.0)
	wireguardMetrics.SetKeepaliveInterval("wg0", "peer2", 30.0)

	// Тестируем последний handshake
	wireguardMetrics.SetLastHandshake("wg0", "peer1", 1640995200.0)
	wireguardMetrics.SetLastHandshake("wg0", "peer2", 1640995300.0)

	// Тестируем попытки handshake
	wireguardMetrics.IncHandshakeAttempts("wg0", "peer1")
	wireguardMetrics.IncHandshakeAttempts("wg0", "peer1")
	wireguardMetrics.IncHandshakeAttempts("wg0", "peer2")

	// Тестируем ошибки handshake
	wireguardMetrics.IncHandshakeFailures("wg0", "peer1", "timeout")
	wireguardMetrics.IncHandshakeFailures("wg0", "peer2", "crypto_error")

	// Тестируем время работы соединения
	wireguardMetrics.SetConnectionUptime("wg0", "peer1", 3600.0)
	wireguardMetrics.SetConnectionUptime("wg0", "peer2", 7200.0)

	// Проверяем, что метрики зарегистрированы
	metrics, err := reg.Gather()
	assert.NoError(t, err)
	assert.Greater(t, len(metrics), 0)
}

func TestWireGuardErrorMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	wireguardMetrics := NewWireGuardMetrics(reg)

	// Тестируем ошибки интерфейсов
	wireguardMetrics.IncInterfaceErrors("wg0", "creation_failed")
	wireguardMetrics.IncInterfaceErrors("wg1", "configuration_error")

	// Тестируем ошибки пиров
	wireguardMetrics.IncPeerErrors("wg0", "peer1", "connection_failed")
	wireguardMetrics.IncPeerErrors("wg0", "peer2", "handshake_timeout")

	// Тестируем ошибки handshake
	wireguardMetrics.IncHandshakeErrors("wg0", "peer1", "crypto_error")
	wireguardMetrics.IncHandshakeErrors("wg0", "peer2", "timeout")

	// Тестируем события таймаута
	wireguardMetrics.IncTimeoutEvents("wg0", "peer1", "idle_timeout")
	wireguardMetrics.IncTimeoutEvents("wg0", "peer2", "handshake_timeout")

	// Тестируем попытки переподключения
	wireguardMetrics.IncReconnectionAttempts("wg0", "peer1")
	wireguardMetrics.IncReconnectionAttempts("wg0", "peer2")

	// Проверяем, что метрики зарегистрированы
	metrics, err := reg.Gather()
	assert.NoError(t, err)
	assert.Greater(t, len(metrics), 0)
}

func TestWireGuardSecurityMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	wireguardMetrics := NewWireGuardMetrics(reg)

	// Тестируем события ротации ключей
	wireguardMetrics.IncKeyRotationEvents("wg0", "private_key")
	wireguardMetrics.IncKeyRotationEvents("wg0", "public_key")

	// Тестируем ошибки аутентификации
	wireguardMetrics.IncAuthenticationFailures("wg0", "peer1", "invalid_key")
	wireguardMetrics.IncAuthenticationFailures("wg0", "peer2", "expired_certificate")

	// Тестируем ошибки шифрования
	wireguardMetrics.IncEncryptionFailures("wg0", "peer1", "cipher_error")
	wireguardMetrics.IncEncryptionFailures("wg0", "peer2", "key_rotation")

	// Тестируем неавторизованный доступ
	wireguardMetrics.IncUnauthorizedAccess("wg0", "peer1", "invalid_peer")
	wireguardMetrics.IncUnauthorizedAccess("wg0", "peer2", "blocked_ip")

	// Проверяем, что метрики зарегистрированы
	metrics, err := reg.Gather()
	assert.NoError(t, err)
	assert.Greater(t, len(metrics), 0)
}

func TestWireGuardSiteToSiteMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	wireguardMetrics := NewWireGuardMetrics(reg)

	// Тестируем site-to-site соединения
	wireguardMetrics.SetSiteToSiteConnections("conn1", "active", 1)
	wireguardMetrics.SetSiteToSiteConnections("conn2", "inactive", 0)

	// Тестируем активные соединения
	wireguardMetrics.SetSiteToSiteActive("conn1", 1)
	wireguardMetrics.SetSiteToSiteActive("conn2", 0)

	// Тестируем полосу пропускания
	wireguardMetrics.SetSiteToSiteBandwidth("conn1", 1000.0)
	wireguardMetrics.SetSiteToSiteBandwidth("conn2", 500.0)

	// Тестируем латентность
	wireguardMetrics.ObserveSiteToSiteLatency("conn1", 0.05)
	wireguardMetrics.ObserveSiteToSiteLatency("conn2", 0.1)

	// Проверяем, что метрики зарегистрированы
	metrics, err := reg.Gather()
	assert.NoError(t, err)
	assert.Greater(t, len(metrics), 0)
}

func TestWireGuardRoutingMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	wireguardMetrics := NewWireGuardMetrics(reg)

	// Тестируем правила маршрутизации
	wireguardMetrics.SetRoutingRules("wg0", "static", 5)
	wireguardMetrics.SetRoutingRules("wg0", "dynamic", 3)

	// Тестируем обновления маршрутизации
	wireguardMetrics.IncRoutingUpdates("wg0", "route_added")
	wireguardMetrics.IncRoutingUpdates("wg0", "route_removed")

	// Тестируем ошибки маршрутизации
	wireguardMetrics.IncRoutingFailures("wg0", "route_conflict")
	wireguardMetrics.IncRoutingFailures("wg0", "invalid_route")

	// Тестируем изменения маршрутов
	wireguardMetrics.IncRouteChanges("wg0", "route_updated")
	wireguardMetrics.IncRouteChanges("wg0", "route_deleted")

	// Проверяем, что метрики зарегистрированы
	metrics, err := reg.Gather()
	assert.NoError(t, err)
	assert.Greater(t, len(metrics), 0)
}

func TestWireGuardResourceMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	wireguardMetrics := NewWireGuardMetrics(reg)

	// Тестируем использование памяти
	wireguardMetrics.SetMemoryUsage("wg0", "interface", 1024*1024)
	wireguardMetrics.SetMemoryUsage("wg0", "peers", 512*1024)

	// Тестируем использование CPU
	wireguardMetrics.SetCPUUsage("wg0", "interface", 25.5)
	wireguardMetrics.SetCPUUsage("wg0", "peers", 15.2)

	// Тестируем количество горутин
	wireguardMetrics.SetGoroutineCount("wg0", "interface", 10)
	wireguardMetrics.SetGoroutineCount("wg0", "peers", 5)

	// Тестируем количество файловых дескрипторов
	wireguardMetrics.SetFileDescriptorCount("wg0", 20)
	wireguardMetrics.SetFileDescriptorCount("wg1", 15)

	// Проверяем, что метрики зарегистрированы
	metrics, err := reg.Gather()
	assert.NoError(t, err)
	assert.Greater(t, len(metrics), 0)
}

func TestWireGuardConfigurationMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	wireguardMetrics := NewWireGuardMetrics(reg)

	// Тестируем обновления конфигурации
	wireguardMetrics.IncConfigUpdates("wg0", "peer_added")
	wireguardMetrics.IncConfigUpdates("wg0", "peer_removed")

	// Тестируем ошибки конфигурации
	wireguardMetrics.IncConfigErrors("wg0", "invalid_config")
	wireguardMetrics.IncConfigErrors("wg0", "parse_error")

	// Тестируем события генерации ключей
	wireguardMetrics.IncKeyGenerationEvents("wg0", "private_key")
	wireguardMetrics.IncKeyGenerationEvents("wg0", "public_key")

	// Тестируем статус интерфейса
	wireguardMetrics.SetInterfaceStatus("wg0", "up", true)
	wireguardMetrics.SetInterfaceStatus("wg1", "down", false)

	// Проверяем, что метрики зарегистрированы
	metrics, err := reg.Gather()
	assert.NoError(t, err)
	assert.Greater(t, len(metrics), 0)
}

func TestWireGuardMetricsIntegration(t *testing.T) {
	// Создаем основной Metrics с WireGuard метриками
	reg := prometheus.NewRegistry()
	mainMetrics := NewMetrics(reg)

	// Проверяем, что WireGuard метрики доступны
	wireguardMetrics := mainMetrics.GetWireGuardMetrics()
	assert.NotNil(t, wireguardMetrics)

	// Тестируем методы основного Metrics
	mainMetrics.IncWireGuardInterfaces()
	mainMetrics.DecWireGuardInterfaces()
	mainMetrics.IncWireGuardInterfacesFailed("wg0", "test_reason")
	mainMetrics.IncWireGuardPeers()
	mainMetrics.DecWireGuardPeers()
	mainMetrics.SetWireGuardPeersConnected("wg0", "peer1", 1)
	mainMetrics.IncWireGuardTrafficBytesReceived("wg0", "peer1", 1024)
	mainMetrics.IncWireGuardTrafficBytesSent("wg0", "peer1", 2048)
	mainMetrics.ObserveWireGuardHandshakeDuration("wg0", "peer1", 0.1)
	mainMetrics.SetWireGuardBandwidthUtilization("wg0", "peer1", 75.5)
	mainMetrics.IncWireGuardInterfaceErrors("wg0", "test_error")
	mainMetrics.IncWireGuardKeyRotationEvents("wg0", "test_key")
	mainMetrics.SetWireGuardSiteToSiteConnections("conn1", "active", 1)
	mainMetrics.SetWireGuardRoutingRules("wg0", "static", 5)
	mainMetrics.SetWireGuardMemoryUsage("wg0", "test", 1024)
	mainMetrics.IncWireGuardConfigUpdates("wg0", "test_update")

	// Проверяем, что метрики зарегистрированы
	metrics, err := reg.Gather()
	assert.NoError(t, err)
	assert.Greater(t, len(metrics), 0)
}



