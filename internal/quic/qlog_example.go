package quic

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// QLogExample демонстрирует использование qlog трейсинга
func QLogExample() {
	// Создаем логгер
	logger, _ := zap.NewDevelopment()

	// Создаем конфигурацию с включенным qlog трейсингом
	config := DefaultConfig()
	config.EnableQLogTracing = true
	config.QLogConfig = &QLogConfig{
		Enabled:       true,
		OutputDir:     "./qlog_output",
		MaxEvents:     1000,
		FlushInterval: 10 * time.Second,
	}

	// Создаем QUIC соединение
	conn := NewQUICConnection(config, logger)

	// Устанавливаем соединение
	ctx := context.Background()
	err := conn.Dial(ctx, "localhost:4433")
	if err != nil {
		logger.Error("Failed to dial QUIC connection", zap.Error(err))
		return
	}

	// Открываем поток
	stream, err := conn.OpenStream(ctx)
	if err != nil {
		logger.Error("Failed to open stream", zap.Error(err))
		return
	}

	// Отправляем данные
	testData := []byte("Hello, QUIC with qlog tracing!")
	_, err = stream.Write(testData)
	if err != nil {
		logger.Error("Failed to write to stream", zap.Error(err))
		return
	}

	// Читаем ответ
	buffer := make([]byte, 1024)
	n, err := stream.Read(buffer)
	if err != nil {
		logger.Error("Failed to read from stream", zap.Error(err))
		return
	}

	logger.Info("Received data",
		zap.String("data", string(buffer[:n])),
		zap.String("conn_id", conn.GetConnectionID()),
	)

	// Получаем статистику qlog трейсинга
	if tracer := conn.GetQLogTracer(); tracer != nil {
		stats := tracer.GetConnectionStats(conn.GetConnectionID())
		logger.Info("QLog tracing stats", zap.Any("stats", stats))
	}

	// Закрываем соединение
	err = conn.Close()
	if err != nil {
		logger.Error("Failed to close connection", zap.Error(err))
		return
	}

	fmt.Println("QLog example completed successfully!")
}

// QLogEnhancedExample демонстрирует qlog трейсинг с Enhanced QUIC
func QLogEnhancedExample() {
	// Создаем логгер
	logger, _ := zap.NewDevelopment()

	// Создаем конфигурацию с Enhanced QUIC и qlog трейсингом
	config := DefaultConfig()
	config.EnableEnhancedQUIC = true
	config.EnableQLogTracing = true
	config.QLogConfig = &QLogConfig{
		Enabled:       true,
		OutputDir:     "./qlog_enhanced_output",
		MaxEvents:     2000,
		FlushInterval: 5 * time.Second,
	}

	enhancedConfig := DefaultEnhancedConfig()
	enhancedConfig.EnableDDoSProtection = true
	enhancedConfig.EnableGeographicRouting = true

	// Создаем Enhanced QUIC соединение
	conn := NewEnhancedQUICConnection(config, enhancedConfig, "tenant-123", logger)

	// Устанавливаем соединение
	ctx := context.Background()
	err := conn.Dial(ctx, "localhost:4433")
	if err != nil {
		logger.Error("Failed to dial Enhanced QUIC connection", zap.Error(err))
		return
	}

	// Открываем поток
	stream, err := conn.OpenStream(ctx)
	if err != nil {
		logger.Error("Failed to open stream", zap.Error(err))
		return
	}

	// Отправляем данные
	testData := []byte("Hello, Enhanced QUIC with qlog tracing!")
	_, err = stream.Write(testData)
	if err != nil {
		logger.Error("Failed to write to stream", zap.Error(err))
		return
	}

	// Читаем ответ
	buffer := make([]byte, 1024)
	n, err := stream.Read(buffer)
	if err != nil {
		logger.Error("Failed to read from stream", zap.Error(err))
		return
	}

	logger.Info("Received data from Enhanced QUIC",
		zap.String("data", string(buffer[:n])),
		zap.String("conn_id", conn.GetConnectionID()),
		zap.String("tenant_id", conn.GetTenantID()),
	)

	// Получаем статистику qlog трейсинга
	if tracer := conn.GetQLogTracer(); tracer != nil {
		stats := tracer.GetConnectionStats(conn.GetConnectionID())
		logger.Info("Enhanced QLog tracing stats", zap.Any("stats", stats))
	}

	// Закрываем соединение
	err = conn.Close()
	if err != nil {
		logger.Error("Failed to close Enhanced connection", zap.Error(err))
		return
	}

	fmt.Println("Enhanced QLog example completed successfully!")
}

// QLogSecurityExample демонстрирует qlog трейсинг с security событиями
func QLogSecurityExample() {
	// Создаем логгер
	logger, _ := zap.NewDevelopment()

	// Создаем конфигурацию с qlog трейсингом
	config := DefaultConfig()
	config.EnableQLogTracing = true
	config.QLogConfig = &QLogConfig{
		Enabled:       true,
		OutputDir:     "./qlog_security_output",
		MaxEvents:     5000,
		FlushInterval: 15 * time.Second,
	}

	// Создаем QUIC соединение
	conn := NewQUICConnection(config, logger)

	// Устанавливаем соединение
	ctx := context.Background()
	err := conn.Dial(ctx, "localhost:4433")
	if err != nil {
		logger.Error("Failed to dial QUIC connection", zap.Error(err))
		return
	}

	// Логируем security события
	if tracer := conn.GetQLogTracer(); tracer != nil {
		tracer.LogSecurityEvent(conn.GetConnectionID(), "handshake_started", map[string]interface{}{
			"tls_version": "1.3",
			"cipher_suite": "TLS_AES_256_GCM_SHA384",
		})

		tracer.LogSecurityEvent(conn.GetConnectionID(), "certificate_validated", map[string]interface{}{
			"certificate_chain_length": 3,
			"validation_time_ms": 15,
		})

		tracer.LogSecurityEvent(conn.GetConnectionID(), "handshake_completed", map[string]interface{}{
			"handshake_duration_ms": 250,
			"negotiated_protocol": "h3",
		})
	}

	// Открываем поток
	stream, err := conn.OpenStream(context.Background())
	if err != nil {
		logger.Error("Failed to open stream", zap.Error(err))
		return
	}

	// Отправляем данные
	testData := []byte("Hello, QUIC with security qlog tracing!")
	_, err = stream.Write(testData)
	if err != nil {
		logger.Error("Failed to write to stream", zap.Error(err))
		return
	}

	// Читаем ответ
	buffer := make([]byte, 1024)
	n, err := stream.Read(buffer)
	if err != nil {
		logger.Error("Failed to read from stream", zap.Error(err))
		return
	}

	logger.Info("Received data with security tracing",
		zap.String("data", string(buffer[:n])),
		zap.String("conn_id", conn.GetConnectionID()),
	)

	// Получаем статистику qlog трейсинга
	if tracer := conn.GetQLogTracer(); tracer != nil {
		stats := tracer.GetConnectionStats(conn.GetConnectionID())
		logger.Info("Security QLog tracing stats", zap.Any("stats", stats))
	}

	// Закрываем соединение
	err = conn.Close()
	if err != nil {
		logger.Error("Failed to close connection", zap.Error(err))
		return
	}

	fmt.Println("Security QLog example completed successfully!")
}
