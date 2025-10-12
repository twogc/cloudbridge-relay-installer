package quic

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// ExperimentalFeaturesExample демонстрирует использование экспериментальных функций
// Этот файл служит справочным материалом для разработчиков
type ExperimentalFeaturesExample struct {
	logger *zap.Logger
	config *Config
}

// Example1_BasicSetup показывает базовую настройку экспериментальных функций
func Example1_BasicSetup() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// 1. Создаем базовую конфигурацию
	config := DefaultConfig()

	// 2. Включаем экспериментальные функции
	config.EnableExperimental = true
	config.ExperimentalFeatures = DefaultExperimentalFeaturesConfig()

	// 3. Настраиваем BBRv2
	config.ExperimentalFeatures.EnableBBRv2 = true
	config.ExperimentalFeatures.CongestionControl = "bbrv2"

	// 4. Настраиваем ACK Frequency
	config.ExperimentalFeatures.EnableACKFrequency = true

	// 5. Greasing (RFC 9287)
	config.ExperimentalFeatures.EnableGreasing = true

	// 6. FEC (опционально, только для высокого packet loss)
	config.ExperimentalFeatures.EnableFEC = false // По умолчанию отключено

	logger.Info("Experimental features configured",
		zap.Bool("bbrv2", config.ExperimentalFeatures.EnableBBRv2),
		zap.Bool("ack_freq", config.ExperimentalFeatures.EnableACKFrequency),
		zap.Bool("greasing", config.ExperimentalFeatures.EnableGreasing))
}

// Example2_EnhancedQUICWithExperimental показывает использование Enhanced QUIC с экспериментальными функциями
func Example2_EnhancedQUICWithExperimental() error {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// 1. Создаем Enhanced конфигурацию
	config := DefaultEnhancedConfig()

	// 2. Настраиваем экспериментальные функции
	config.ExperimentalFeatures = &ExperimentalFeaturesConfig{
		EnableBBRv2:        true,
		CongestionControl:  "bbrv2",
		EnableACKFrequency: true,
		ACKFrequencyConfig: &ACKFrequencyConfig{
			MaxACKDelay:       25 * time.Millisecond,
			MinACKDelay:       1 * time.Millisecond,
			DefaultACKDelay:   5 * time.Millisecond,
			EnableAdaptive:    true,
			HighLoadThreshold: 1000,
			LowLoadThreshold:  100,
		},
		EnableFEC:      false,
		FECRedundancy:  0.1,
		EnableGreasing: true,
	}

	// 3. Инициализируем экспериментальные функции
	if err := config.InitializeExperimentalFeatures(logger); err != nil {
		return fmt.Errorf("failed to initialize experimental features: %w", err)
	}

	logger.Info("Enhanced QUIC with experimental features initialized")

	// 4. При завершении работы
	defer config.StopExperimentalFeatures()

	return nil
}

// Example3_UseBBRv2InConnection показывает использование BBRv2 в соединении
func Example3_UseBBRv2InConnection() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Создаем Enhanced соединение с BBRv2
	config := DefaultEnhancedConfig()
	config.ExperimentalFeatures = DefaultExperimentalFeaturesConfig()
	config.InitializeExperimentalFeatures(logger)

	// Создаем соединение
	conn, err := NewEnhancedPolicies("tenant-123", config)
	if err != nil {
		logger.Error("Failed to create connection", zap.Error(err))
		return
	}
	defer conn.Close()

	// Обработка ACK от удаленного peer
	// Вызывается когда получаем подтверждение о доставке пакета
	bytesAcked := 1460 // MTU размер
	rtt := 50 * time.Millisecond
	conn.OnPacketAck(bytesAcked, rtt)

	// Обработка потери пакета
	// Вызывается когда обнаруживается потеря
	bytesLost := 1460
	conn.OnPacketLoss(bytesLost)

	logger.Info("BBRv2 congestion control in action",
		zap.Int("bytes_acked", bytesAcked),
		zap.Duration("rtt", rtt))
}

// Example4_UseACKFrequency показывает использование ACK Frequency optimization
func Example4_UseACKFrequency() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// 1. Создаем ACK Frequency manager
	ackConfig := &ACKFrequencyConfig{
		MaxACKDelay:       25 * time.Millisecond,
		MinACKDelay:       1 * time.Millisecond,
		DefaultACKDelay:   5 * time.Millisecond,
		EnableAdaptive:    true,
		HighLoadThreshold: 1000,
		LowLoadThreshold:  100,
	}

	ackManager := NewACKFrequencyManager(logger, ackConfig)

	// 2. Регистрируем соединение
	mockConn := &struct{}{}
	connID := ackManager.RegisterConnection(mockConn)

	// 3. Настраиваем тип трафика
	ackManager.SetTrafficType(connID, "interactive") // bulk, interactive, realtime

	// 4. Проверяем, нужно ли отправлять ACK
	packetSize := 1460
	shouldSendACK := ackManager.ShouldSendACK(connID, packetSize)

	if shouldSendACK {
		logger.Info("Time to send ACK", zap.String("conn_id", connID))
		// Отправляем ACK...
	}

	// 5. Получаем метрики
	metrics := ackManager.GetMetrics()
	logger.Info("ACK Frequency metrics",
		zap.Int64("active_connections", metrics.ActiveConnections),
		zap.Int64("total_acks", metrics.TotalACKs),
		zap.Duration("avg_delay", metrics.AverageACKDelay))

	// 6. При завершении
	defer ackManager.UnregisterConnection(connID)
}

// Example5_UseFEC показывает использование Forward Error Correction
func Example5_UseFEC() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// 1. Создаем FEC manager
	fecConfig := &FECConfig{
		RedundancyRatio:    0.1,  // 10% избыточности
		Algorithm:          "xor", // или "reed-solomon"
		GroupSize:          10,
		EnableAdaptive:     true,
		TargetRecoveryRate: 0.95,
	}

	fecManager := NewFECManager(logger, fecConfig)
	defer fecManager.Stop()

	// 2. Подготавливаем данные для отправки
	dataPackets := [][]byte{
		[]byte("packet1"),
		[]byte("packet2"),
		[]byte("packet3"),
		[]byte("packet4"),
		[]byte("packet5"),
	}

	// 3. Создаем FEC группу с избыточными пакетами
	group, err := fecManager.CreateGroup(dataPackets)
	if err != nil {
		logger.Error("Failed to create FEC group", zap.Error(err))
		return
	}

	logger.Info("FEC group created",
		zap.Int("data_packets", len(group.DataPackets)),
		zap.Int("redundant_packets", len(group.RedundantPackets)))

	// 4. Отправляем все пакеты (данные + избыточные)
	// ...

	// 5. При получении на другой стороне, если пакет потерян:
	receivedIndices := []int{0, 1, 3, 4} // Пакет 2 потерян
	recovered, err := fecManager.RecoverLostPackets(group.ID, receivedIndices, group.RedundantPackets)
	if err == nil && len(recovered) > 0 {
		logger.Info("Packet recovered successfully", zap.Int("recovered_count", len(recovered)))
	}

	// 6. Получаем метрики
	metrics := fecManager.GetMetrics()
	logger.Info("FEC metrics",
		zap.Int64("total_groups", metrics.TotalGroups),
		zap.Int64("recovery_events", metrics.RecoveryEvents),
		zap.Float64("recovery_rate", metrics.RecoveryRate))
}

// Example6_UseGreasing показывает использование QUIC Bit Greasing
func Example6_UseGreasing() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// 1. Создаем Greasing manager
	greasingConfig := DefaultGreasingConfig()
	greasingManager := NewGreasingManager(logger, greasingConfig)
	defer greasingManager.Stop()

	// 2. Применяем greasing к версиям QUIC
	versions := []uint32{1, 2, 3}
	greasedVersions, err := greasingManager.GreaseVersionNegotiation(versions)
	if err != nil {
		logger.Error("Failed to grease versions", zap.Error(err))
		return
	}

	logger.Info("Version greasing applied",
		zap.Int("original", len(versions)),
		zap.Int("greased", len(greasedVersions)))

	// 3. Применяем greasing к transport parameters
	params := map[uint64][]byte{
		0x01: []byte("value1"),
		0x02: []byte("value2"),
	}

	greasedParams, err := greasingManager.GreaseTransportParameters(params)
	if err != nil {
		logger.Error("Failed to grease parameters", zap.Error(err))
		return
	}

	logger.Info("Transport parameter greasing applied",
		zap.Int("original", len(params)),
		zap.Int("greased", len(greasedParams)))

	// 4. Получаем метрики
	metrics := greasingManager.GetMetrics()
	logger.Info("Greasing metrics",
		zap.Int64("version_grease_count", metrics.VersionGreaseCount),
		zap.Int64("param_grease_count", metrics.TransportParamGreaseCount),
		zap.Int64("total_applications", metrics.TotalGreaseApplications))
}

// Example7_CompleteIntegration показывает полную интеграцию всех функций
func Example7_CompleteIntegration() error {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// 1. Создаем полную конфигурацию
	config := DefaultEnhancedConfig()

	// 2. Включаем все экспериментальные функции
	config.ExperimentalFeatures = &ExperimentalFeaturesConfig{
		EnableBBRv2:        true,
		CongestionControl:  "bbrv2",
		EnableACKFrequency: true,
		ACKFrequencyConfig: DefaultACKFrequencyConfig(),
		EnableFEC:          true, // Включаем для демонстрации
		FECRedundancy:      0.1,
		EnableGreasing:     true,
		EnableGSO:          true,
		EnableGRO:          true,
		SocketBufferSize:   1024 * 1024,
	}

	// 3. Инициализируем
	if err := config.InitializeExperimentalFeatures(logger); err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}
	defer config.StopExperimentalFeatures()

	logger.Info("All experimental features initialized successfully")

	// 4. Создаем соединение
	conn, err := NewEnhancedPolicies("production-tenant", config)
	if err != nil {
		return fmt.Errorf("failed to create connection: %w", err)
	}
	defer conn.Close()

	// 5. Симуляция работы соединения
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("Simulation complete")
			return nil

		case <-ticker.C:
			// Симулируем получение ACK
			conn.OnPacketAck(1460, 50*time.Millisecond)

			// Симулируем периодические потери (редко)
			if time.Now().Unix()%10 == 0 {
				conn.OnPacketLoss(1460)
			}

			// Применяем FEC к пакетам
			packets := [][]byte{[]byte("data1"), []byte("data2")}
			fecPackets, _ := conn.ApplyFEC(packets)
			logger.Debug("FEC applied", zap.Int("total_packets", len(fecPackets)))

			// Применяем Greasing
			versions := []uint32{1}
			params := make(map[uint64][]byte)
			greasedVer, greasedParams, _ := conn.GreaseConnection(versions, params)
			logger.Debug("Greasing applied",
				zap.Int("versions", len(greasedVer)),
				zap.Int("params", len(greasedParams)))
		}
	}
}

// Example8_MonitoringAndMetrics показывает сбор метрик
func Example8_MonitoringAndMetrics() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	config := DefaultEnhancedConfig()
	config.ExperimentalFeatures = DefaultExperimentalFeaturesConfig()
	config.InitializeExperimentalFeatures(logger)
	defer config.StopExperimentalFeatures()

	// Сбор метрик BBRv2
	if config.CCManager != nil {
		ccMetrics := config.CCManager.GetMetrics()
		logger.Info("BBRv2 metrics",
			zap.String("algorithm", ccMetrics.Algorithm),
			zap.Int64("congestion_window", ccMetrics.CongestionWindow),
			zap.Float64("throughput_mbps", ccMetrics.Throughput),
			zap.Int64("pacing_rate_bps", ccMetrics.PacingRate),
			zap.Float64("loss_rate", ccMetrics.LossRate))
	}

	// Сбор метрик ACK Frequency
	if config.ACKFreqManager != nil {
		ackMetrics := config.ACKFreqManager.GetMetrics()
		logger.Info("ACK Frequency metrics",
			zap.Int64("active_connections", ackMetrics.ActiveConnections),
			zap.Int64("total_acks", ackMetrics.TotalACKs),
			zap.Duration("avg_delay", ackMetrics.AverageACKDelay),
			zap.Int64("bandwidth_saved", ackMetrics.BandwidthSaved))
	}

	// Сбор метрик FEC
	if config.FECManager != nil && config.FECManager.IsActive() {
		fecMetrics := config.FECManager.GetMetrics()
		logger.Info("FEC metrics",
			zap.Int64("total_groups", fecMetrics.TotalGroups),
			zap.Int64("recovery_events", fecMetrics.RecoveryEvents),
			zap.Float64("recovery_rate", fecMetrics.RecoveryRate),
			zap.Int64("bytes_saved", fecMetrics.BytesSaved))
	}

	// Сбор метрик Greasing
	if config.GreasingManager != nil && config.GreasingManager.IsActive() {
		greasingMetrics := config.GreasingManager.GetMetrics()
		logger.Info("Greasing metrics",
			zap.Int64("version_grease", greasingMetrics.VersionGreaseCount),
			zap.Int64("param_grease", greasingMetrics.TransportParamGreaseCount),
			zap.Int64("total_applications", greasingMetrics.TotalGreaseApplications),
			zap.Int64("compatibility_issues", greasingMetrics.CompatibilityIssues))
	}
}

// Example9_ErrorHandling показывает обработку ошибок
func Example9_ErrorHandling() error {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	config := DefaultEnhancedConfig()
	config.ExperimentalFeatures = DefaultExperimentalFeaturesConfig()

	// Инициализация с обработкой ошибок
	if err := config.InitializeExperimentalFeatures(logger); err != nil {
		logger.Error("Failed to initialize experimental features",
			zap.Error(err))
		// Fallback на стандартную конфигурацию
		config.ExperimentalFeatures = nil
		logger.Info("Falling back to standard QUIC configuration")
	}

	// Проверка доступности функций
	if config.CCManager == nil {
		logger.Warn("BBRv2 not available, using default congestion control")
	}

	if config.ACKFreqManager == nil {
		logger.Warn("ACK Frequency optimization not available")
	}

	// Graceful degradation
	conn, err := NewEnhancedPolicies("tenant", config)
	if err != nil {
		return fmt.Errorf("failed to create connection: %w", err)
	}
	defer conn.Close()

	logger.Info("Connection established with available features")
	return nil
}

// Example10_ProductionDeployment показывает рекомендации для production
func Example10_ProductionDeployment() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Рекомендованная production конфигурация
	config := DefaultEnhancedConfig()
	config.ExperimentalFeatures = &ExperimentalFeaturesConfig{
		// BBRv2: РЕКОМЕНДУЕТСЯ включить
		EnableBBRv2:       true,
		CongestionControl: "bbrv2",

		// ACK Frequency: РЕКОМЕНДУЕТСЯ включить
		EnableACKFrequency: true,
		ACKFrequencyConfig: DefaultACKFrequencyConfig(),

		// FEC: Включать только при packet loss > 1%
		EnableFEC:     false,
		FECRedundancy: 0.1,

		// Greasing: РЕКОМЕНДУЕТСЯ включить
		EnableGreasing: true,

		// Advanced features
		EnableGSO:        true, // Если поддерживается
		EnableGRO:        true, // Если поддерживается
		SocketBufferSize: 1024 * 1024,
	}

	config.InitializeExperimentalFeatures(logger)
	defer config.StopExperimentalFeatures()

	logger.Info("Production configuration ready",
		zap.String("recommendation", "Start with BBRv2 and ACK Frequency"),
		zap.String("monitoring", "Enable Prometheus metrics"),
		zap.String("documentation", "docs/integration/UNIT_TESTS_COMPLETION_REPORT.md"))
}
