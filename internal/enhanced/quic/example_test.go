package quic

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/twogc/cloudbridge-relay-installer/internal/enhanced/metrics"
)

// ExampleQUICServer демонстрирует использование QUIC сервера
func ExampleQUICServer() {
	// Создаем логгер
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal("Failed to create logger:", err)
	}
	defer logger.Sync()

	// Создаем метрики
	registry := prometheus.NewRegistry()
	enhancedMetrics := metrics.NewEnhancedMetrics(registry)

	// Создаем конфигурацию QUIC сервера
	config := &QUICServerConfig{
		Address:           "127.0.0.1",
		Port:              8082,
		MaxConnections:    100,
		ConnectionTimeout: 60 * time.Second,
		EnableTLS:         false, // Для примера отключаем TLS
		EnableMetrics:     true,
		EnableTracing:     false,
	}

	// Создаем QUIC сервер
	server := NewQUICServer(config, logger, enhancedMetrics)

	// Запускаем сервер
	if err := server.Start(); err != nil {
		logger.Fatal("Failed to start QUIC server", zap.Error(err))
	}

	logger.Info("QUIC server started", zap.String("address", server.GetAddress()))

	// Ждем некоторое время для демонстрации
	time.Sleep(5 * time.Second)

	// Получаем статистику
	stats := server.GetStats()
	logger.Info("Server statistics", zap.Any("stats", stats))

	// Останавливаем сервер
	if err := server.Stop(); err != nil {
		logger.Error("Failed to stop QUIC server", zap.Error(err))
	}

	logger.Info("QUIC server stopped")
}

// ExampleQUICConfig демонстрирует различные конфигурации QUIC
func ExampleQUICConfig() {
	// Конфигурация по умолчанию
	defaultConfig := DefaultQUICConfig()
	fmt.Printf("Default config: MaxStreams=%d, KeepAlive=%v\n", 
		defaultConfig.MaxStreams, defaultConfig.KeepAliveInterval)

	// Пользовательская конфигурация для высоконагруженного сервера
	highLoadConfig := &QUICConfig{
		MaxStreams:        1000,
		KeepAliveInterval: 15 * time.Second,
		IdleTimeout:       30 * time.Second,
		MaxIdleTime:       180 * time.Second,
		EnableMetrics:     true,
		EnableTracing:     true,
	}
	fmt.Printf("High load config: MaxStreams=%d, KeepAlive=%v\n", 
		highLoadConfig.MaxStreams, highLoadConfig.KeepAliveInterval)

	// Конфигурация для тестирования
	testConfig := &QUICConfig{
		MaxStreams:        10,
		KeepAliveInterval: 0, // Отключаем keep-alive
		IdleTimeout:       10 * time.Second,
		MaxIdleTime:       60 * time.Second,
		EnableMetrics:     false,
		EnableTracing:     false,
	}
	fmt.Printf("Test config: MaxStreams=%d, KeepAlive=%v\n", 
		testConfig.MaxStreams, testConfig.KeepAliveInterval)
}

// ExampleQUICServerConfig демонстрирует различные конфигурации QUIC сервера
func ExampleQUICServerConfig() {
	// Конфигурация по умолчанию
	defaultServerConfig := DefaultQUICServerConfig()
	fmt.Printf("Default server config: Address=%s, Port=%d, MaxConnections=%d\n",
		defaultServerConfig.Address, defaultServerConfig.Port, defaultServerConfig.MaxConnections)

	// Конфигурация для продакшена
	prodServerConfig := &QUICServerConfig{
		Address:           "0.0.0.0",
		Port:              443,
		MaxConnections:    10000,
		ConnectionTimeout: 120 * time.Second,
		EnableTLS:         true,
		CertFile:          "/path/to/cert.pem",
		KeyFile:           "/path/to/key.pem",
		EnableMetrics:     true,
		EnableTracing:     true,
	}
	fmt.Printf("Production server config: Address=%s, Port=%d, MaxConnections=%d\n",
		prodServerConfig.Address, prodServerConfig.Port, prodServerConfig.MaxConnections)

	// Конфигурация для разработки
	devServerConfig := &QUICServerConfig{
		Address:           "localhost",
		Port:              8080,
		MaxConnections:    100,
		ConnectionTimeout: 30 * time.Second,
		EnableTLS:         false,
		EnableMetrics:     true,
		EnableTracing:     false,
	}
	fmt.Printf("Development server config: Address=%s, Port=%d, MaxConnections=%d\n",
		devServerConfig.Address, devServerConfig.Port, devServerConfig.MaxConnections)
}

// ExampleQUICServer_WithContext демонстрирует использование QUIC сервера с контекстом
func ExampleQUICServer_WithContext() {
	logger, _ := zap.NewDevelopment()
	registry := prometheus.NewRegistry()
	enhancedMetrics := metrics.NewEnhancedMetrics(registry)

	// Создаем контекст с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Создаем сервер
	server := NewQUICServer(DefaultQUICServerConfig(), logger, enhancedMetrics)

	// Запускаем сервер в горутине
	go func() {
		if err := server.Start(); err != nil {
			logger.Error("Server error", zap.Error(err))
		}
	}()

	// Ждем завершения контекста
	<-ctx.Done()

	// Останавливаем сервер
	if err := server.Stop(); err != nil {
		logger.Error("Failed to stop server", zap.Error(err))
	}

	logger.Info("Server stopped due to context cancellation")
}

// ExampleQUICServer_WithGracefulShutdown демонстрирует graceful shutdown
func ExampleQUICServer_WithGracefulShutdown() {
	logger, _ := zap.NewDevelopment()
	registry := prometheus.NewRegistry()
	enhancedMetrics := metrics.NewEnhancedMetrics(registry)

	// Создаем сервер
	server := NewQUICServer(DefaultQUICServerConfig(), logger, enhancedMetrics)

	// Запускаем сервер
	if err := server.Start(); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}

	// Симулируем работу сервера
	go func() {
		time.Sleep(5 * time.Second)
		logger.Info("Simulating graceful shutdown signal")
	}()

	// Graceful shutdown
	logger.Info("Starting graceful shutdown...")
	
	// Даем время на завершение активных соединений
	time.Sleep(2 * time.Second)
	
	// Останавливаем сервер
	if err := server.Stop(); err != nil {
		logger.Error("Failed to stop server gracefully", zap.Error(err))
	}

	logger.Info("Server stopped gracefully")
}

// ExampleQUICServer_WithMetrics демонстрирует работу с метриками
func ExampleQUICServer_WithMetrics() {
	logger, _ := zap.NewDevelopment()
	registry := prometheus.NewRegistry()
	enhancedMetrics := metrics.NewEnhancedMetrics(registry)

	// Создаем сервер с включенными метриками
	config := DefaultQUICServerConfig()
	config.EnableMetrics = true
	
	server := NewQUICServer(config, logger, enhancedMetrics)

	// Запускаем сервер
	if err := server.Start(); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}

	// Собираем метрики каждые 5 секунд
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			stats := server.GetStats()
			logger.Info("Server metrics", 
				zap.Int32("active_connections", stats["active_connections"].(int32)),
				zap.Int64("total_connections", stats["total_connections"].(int64)),
				zap.String("uptime", stats["uptime"].(string)),
			)
		}
	}()

	// Работаем 15 секунд
	time.Sleep(15 * time.Second)

	// Останавливаем сервер
	server.Stop()
	logger.Info("Server stopped, final metrics collected")
}

// ExampleQUICServer_WithTLS демонстрирует настройку TLS
func ExampleQUICServer_WithTLS() {
	logger, _ := zap.NewDevelopment()
	registry := prometheus.NewRegistry()
	enhancedMetrics := metrics.NewEnhancedMetrics(registry)

	// Конфигурация с TLS
	config := &QUICServerConfig{
		Address:           "127.0.0.1",
		Port:              8443,
		MaxConnections:    100,
		ConnectionTimeout: 60 * time.Second,
		EnableTLS:         true,
		// CertFile и KeyFile оставляем пустыми для использования самоподписанного сертификата
		EnableMetrics:     true,
		EnableTracing:     false,
	}

	server := NewQUICServer(config, logger, enhancedMetrics)

	// Запускаем сервер
	if err := server.Start(); err != nil {
		logger.Fatal("Failed to start TLS server", zap.Error(err))
	}

	logger.Info("TLS QUIC server started", zap.String("address", server.GetAddress()))

	// Работаем некоторое время
	time.Sleep(5 * time.Second)

	// Останавливаем сервер
	server.Stop()
	logger.Info("TLS QUIC server stopped")
} 