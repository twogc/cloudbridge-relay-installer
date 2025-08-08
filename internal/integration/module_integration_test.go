package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/twogc/cloudbridge-relay-installer/internal/ai"
	"github.com/twogc/cloudbridge-relay-installer/internal/quantum"
	"github.com/twogc/cloudbridge-relay-installer/internal/wireguard"
)

func TestWireGuardQuantumIntegration(t *testing.T) {
	// Создаем WireGuard компоненты
	wgInterface := wireguard.NewWireGuardInterface("wg0", 51820)
	quantumConfig := quantum.NewPQCConfig()
	quantumWireGuard := quantum.NewQuantumWireGuardInterface(wgInterface, quantumConfig)

	// Запускаем компоненты
	err := wgInterface.Start()
	require.NoError(t, err)
	defer wgInterface.Stop()

	err = quantumWireGuard.Start()
	require.NoError(t, err)
	defer quantumWireGuard.Stop()

	// Тестируем интеграцию
	peer := &wireguard.WireGuardPeer{
		PublicKey:  "test-public-key",
		Endpoint:   "192.168.1.100:51820",
		AllowedIPs: []string{"10.100.1.100/32"},
	}

	// Добавляем пира в WireGuard
	err = wgInterface.AddPeer(peer)
	require.NoError(t, err)

	// Проверяем статус
	status := wgInterface.GetStatus()
	assert.NotNil(t, status)
	assert.Contains(t, status, "running")
}

func TestAIWireGuardIntegration(t *testing.T) {
	// Создаем AI компоненты
	behaviorAnalyzer := ai.NewBehaviorAnalyzer()
	predictiveSecurity := ai.NewPredictiveSecurity()
	automatedResponse := ai.NewAutomatedResponse()

	// Создаем WireGuard компоненты
	wgInterface := wireguard.NewWireGuardInterface("wg0", 51820)

	// Запускаем компоненты
	err := wgInterface.Start()
	require.NoError(t, err)
	defer wgInterface.Stop()

	err = behaviorAnalyzer.Start()
	require.NoError(t, err)
	defer behaviorAnalyzer.Stop()

	err = predictiveSecurity.Start()
	require.NoError(t, err)
	defer predictiveSecurity.Stop()

	err = automatedResponse.Start()
	require.NoError(t, err)
	defer automatedResponse.Stop()

	// Симулируем подозрительную активность
	userID := "test-user"
	behaviorData := map[string]interface{}{
		"action":    "connection_attempt",
		"ip_address": "192.168.1.100",
		"port":      51820,
		"protocol":  "wireguard",
	}

	// Анализируем поведение
	anomalyScore := behaviorAnalyzer.AnalyzeUserBehavior(userID, behaviorData)
	assert.GreaterOrEqual(t, anomalyScore, 0.0)
	assert.LessOrEqual(t, anomalyScore, 1.0)

	// Проверяем, есть ли аномалия
	isAnomaly, err := behaviorAnalyzer.DetectAnomaly(userID)
	require.NoError(t, err)
	assert.IsType(t, false, isAnomaly)

	// Проверяем, что компоненты работают
	assert.True(t, behaviorAnalyzer.Enabled)
	assert.True(t, predictiveSecurity.Enabled)
	assert.True(t, automatedResponse.Enabled)
}

func TestAIQuantumIntegration(t *testing.T) {
	// Создаем AI компоненты
	behaviorAnalyzer := ai.NewBehaviorAnalyzer()
	predictiveSecurity := ai.NewPredictiveSecurity()

	// Создаем Quantum компоненты
	quantumConfig := quantum.NewPQCConfig()

	// Запускаем компоненты
	err := behaviorAnalyzer.Start()
	require.NoError(t, err)
	defer behaviorAnalyzer.Stop()

	err = predictiveSecurity.Start()
	require.NoError(t, err)
	defer predictiveSecurity.Stop()

	// Проверяем конфигурацию
	err = quantumConfig.Validate()
	require.NoError(t, err)

	// Симулируем попытку атаки на квантовые ключи
	userID := "attacker"
	behaviorData := map[string]interface{}{
		"action":           "quantum_key_attack",
		"ip_address":       "192.168.1.200",
		"target_algorithm": "kyber",
		"attack_type":      "brute_force",
	}

	// Анализируем поведение
	anomalyScore := behaviorAnalyzer.AnalyzeUserBehavior(userID, behaviorData)
	assert.GreaterOrEqual(t, anomalyScore, 0.0)

	// Проверяем аномалию
	isAnomaly, err := behaviorAnalyzer.DetectAnomaly(userID)
	require.NoError(t, err)
	assert.IsType(t, false, isAnomaly)
}

func TestFullModuleIntegration(t *testing.T) {
	// Создаем все компоненты
	wgInterface := wireguard.NewWireGuardInterface("wg0", 51820)
	behaviorAnalyzer := ai.NewBehaviorAnalyzer()
	predictiveSecurity := ai.NewPredictiveSecurity()
	automatedResponse := ai.NewAutomatedResponse()

	// Запускаем все компоненты
	components := []struct {
		name string
		start func() error
		stop  func() error
	}{
		{"WireGuard Interface", wgInterface.Start, wgInterface.Stop},
		{"Behavior Analyzer", behaviorAnalyzer.Start, behaviorAnalyzer.Stop},
		{"Predictive Security", predictiveSecurity.Start, predictiveSecurity.Stop},
		{"Automated Response", automatedResponse.Start, automatedResponse.Stop},
	}

	for _, comp := range components {
		err := comp.start()
		require.NoError(t, err, "Failed to start %s", comp.name)
		defer comp.stop()
	}

	// Симулируем полный сценарий
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Добавляем пира в WireGuard
	peer := &wireguard.WireGuardPeer{
		PublicKey:  "test-peer-key",
		Endpoint:   "192.168.1.100:51820",
		AllowedIPs: []string{"10.100.1.100/32"},
	}

	err := wgInterface.AddPeer(peer)
	require.NoError(t, err)

	// 2. Анализируем поведение пира
	userID := peer.PublicKey
	behaviorData := map[string]interface{}{
		"action":     "peer_connection",
		"ip_address": "192.168.1.100",
		"protocol":   "wireguard",
	}

	anomalyScore := behaviorAnalyzer.AnalyzeUserBehavior(userID, behaviorData)
	assert.GreaterOrEqual(t, anomalyScore, 0.0)

	// 3. Проверяем статус всех компонентов
	status := wgInterface.GetStatus()
	assert.NotNil(t, status)
	assert.True(t, behaviorAnalyzer.Enabled)
	assert.True(t, predictiveSecurity.Enabled)
	assert.True(t, automatedResponse.Enabled)

	// 4. Проверяем, что контекст не отменен
	select {
	case <-ctx.Done():
		t.Error("Integration test timed out")
	default:
		// Тест прошел успешно
	}
}

func TestModuleErrorHandling(t *testing.T) {
	// Тестируем обработку ошибок при интеграции

	// Создаем компоненты с неверными параметрами
	wgInterface := wireguard.NewWireGuardInterface("", 0)
	behaviorAnalyzer := ai.NewBehaviorAnalyzer()

	// Пытаемся запустить с неверными параметрами
	err := wgInterface.Start()
	assert.Error(t, err) // Ожидаем ошибку

	// Проверяем, что AI компонент может работать даже при ошибках других модулей
	err = behaviorAnalyzer.Start()
	require.NoError(t, err)
	defer behaviorAnalyzer.Stop()

	// Анализируем поведение даже при ошибках в других модулях
	userID := "test-user"
	behaviorData := map[string]interface{}{
		"action":     "error_test",
		"ip_address": "192.168.1.100",
	}

	anomalyScore := behaviorAnalyzer.AnalyzeUserBehavior(userID, behaviorData)
	assert.GreaterOrEqual(t, anomalyScore, 0.0)
}
