package main

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/twogc/cloudbridge-relay-installer/config"
	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
)

func TestStatelessResetManager(t *testing.T) {
	// Создаем тестовый логгер
	logger, err := logger.New(logger.Config{
		Level:  logger.DEBUG,
		Output: "stdout",
		JSON:   false,
	})
	require.NoError(t, err)

	// Создаем StatelessResetManager
	srm := NewStatelessResetManager(logger, "/tmp/test_stateless_reset.key")
	require.NotNil(t, srm)

	// Тестируем загрузку/генерацию ключа
	err = srm.LoadOrGenerateKey()
	assert.NoError(t, err)

	// Проверяем, что ключ сгенерирован
	key := srm.GetKey()
	assert.NotNil(t, key)
	assert.NotEqual(t, [32]byte{}, *key)

	// Проверяем валидность ключа
	err = srm.ValidateKey()
	assert.NoError(t, err)

	// Получаем информацию о ключе
	info := srm.GetKeyInfo()
	assert.NotNil(t, info)
	assert.Equal(t, "/tmp/test_stateless_reset.key", info["key_file"])
	assert.Equal(t, 32, info["key_length"])
	assert.True(t, info["has_file"].(bool))
}

func TestICESupport(t *testing.T) {
	// Создаем тестовый логгер
	logger, err := logger.New(logger.Config{
		Level:  logger.DEBUG,
		Output: "stdout",
		JSON:   false,
	})
	require.NoError(t, err)

	// Создаем тестовую конфигурацию
	cfg := &config.Config{
		ICEStunTurn: config.Config{}.ICEStunTurn,
	}
	cfg.ICEStunTurn.Enabled = true
	cfg.ICEStunTurn.StunServers = []string{"stun.l.google.com:19302"}
	cfg.ICEStunTurn.ICETimeout = 30 * time.Second
	cfg.ICEStunTurn.ConnectionTimeout = 10 * time.Second
	cfg.ICEStunTurn.GatheringTimeout = 5 * time.Second

	// Создаем ICE поддержку
	iceSupport, err := NewICESupport(logger, cfg)
	require.NoError(t, err)
	require.NotNil(t, iceSupport)

	// Тестируем запуск ICE поддержки
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = iceSupport.Start(ctx)
	assert.NoError(t, err)

	// Ждем немного для сбора кандидатов
	time.Sleep(2 * time.Second)

	// Проверяем, что кандидаты собраны
	candidates := iceSupport.GetLocalCandidates()
	t.Logf("Gathered %d ICE candidates", len(candidates))

	// Закрываем ICE поддержку
	err = iceSupport.Close()
	assert.NoError(t, err)
}

func TestMASQUEServer(t *testing.T) {
	// Создаем тестовый логгер
	logger, err := logger.New(logger.Config{
		Level:  logger.DEBUG,
		Output: "stdout",
		JSON:   false,
	})
	require.NoError(t, err)

	// Создаем тестовую конфигурацию
	cfg := &config.Config{
		MASQUE: config.Config{}.MASQUE,
	}
	cfg.MASQUE.Enabled = true
	cfg.MASQUE.ProxyURL = "https://test.example.com:8443"
	cfg.MASQUE.ALPNProtocols = []string{"h3", "h3-29"}
	cfg.MASQUE.FallbackEnabled = true

	// Создаем MASQUE сервер
	masqueServer := NewMASQUEServer(logger, cfg)
	require.NotNil(t, masqueServer)

	// Тестируем запуск MASQUE сервера
	err = masqueServer.Start()
	assert.NoError(t, err)

	// Проверяем информацию о соединениях
	info := masqueServer.GetConnectionInfo()
	assert.NotNil(t, info)
	assert.Equal(t, 0, info["total_connections"])
	assert.True(t, info["is_running"].(bool))

	// Закрываем MASQUE сервер
	err = masqueServer.Close()
	assert.NoError(t, err)
}

func TestP2PQUICServerProperIntegration(t *testing.T) {
	// Создаем тестовый логгер
	logger, err := logger.New(logger.Config{
		Level:  logger.DEBUG,
		Output: "stdout",
		JSON:   false,
	})
	require.NoError(t, err)

	// Создаем тестовую конфигурацию
	cfg := &config.Config{
		ICEStunTurn: config.Config{}.ICEStunTurn,
		MASQUE:      config.Config{}.MASQUE,
	}
	cfg.ICEStunTurn.Enabled = true
	cfg.ICEStunTurn.StunServers = []string{"stun.l.google.com:19302"}
	cfg.MASQUE.Enabled = true
	cfg.MASQUE.ProxyURL = "https://test.example.com:8443"

	// Создаем контекст
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Создаем P2P QUIC сервер
	quicRelayHandler := &QUICRelayHandler{} // Заглушка
	server := NewP2PQUICServerProper(logger, ctx, quicRelayHandler, cfg)
	require.NotNil(t, server)

	// Проверяем, что компоненты созданы
	assert.NotNil(t, server.statelessResetManager)
	assert.NotNil(t, server.iceSupport)
	assert.NotNil(t, server.masqueServer)

	// Тестируем запуск сервера (только инициализацию, без сетевых операций)
	// В реальном тесте здесь был бы полный запуск, но для unit теста достаточно проверки инициализации

	// Проверяем StatelessResetManager
	err = server.statelessResetManager.LoadOrGenerateKey()
	assert.NoError(t, err)

	// Проверяем ICE поддержку
	err = server.iceSupport.Start(ctx)
	assert.NoError(t, err)

	// Проверяем MASQUE сервер
	err = server.masqueServer.Start()
	assert.NoError(t, err)

	// Останавливаем сервер
	err = server.Stop()
	assert.NoError(t, err)
}

// Benchmark тесты для проверки производительности
func BenchmarkStatelessResetManager(b *testing.B) {
	logger, _ := logger.New(logger.Config{
		Level:  logger.INFO,
		Output: "stdout",
		JSON:   false,
	})

	srm := NewStatelessResetManager(logger, "/tmp/bench_stateless_reset.key")
	srm.LoadOrGenerateKey()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		srm.GetKey()
	}
}

func BenchmarkICESupport(b *testing.B) {
	logger, _ := logger.New(logger.Config{
		Level:  logger.INFO,
		Output: "stdout",
		JSON:   false,
	})

	cfg := &config.Config{
		ICEStunTurn: config.Config{}.ICEStunTurn,
	}
	cfg.ICEStunTurn.Enabled = true
	cfg.ICEStunTurn.StunServers = []string{"stun.l.google.com:19302"}

	iceSupport, _ := NewICESupport(logger, cfg)
	ctx := context.Background()
	iceSupport.Start(ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iceSupport.GetLocalCandidates()
	}

	iceSupport.Close()
}
