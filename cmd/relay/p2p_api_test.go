package main

import (
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestNewP2PAPIServer(t *testing.T) {
	// Создаем тестовый логгер
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// Тестируем создание P2P API Server
	jwtSecret := []byte("test-jwt-secret")
	server := NewP2PAPIServer(logger, jwtSecret, 8080)

	// Проверяем, что сервер создался
	if server == nil {
		t.Fatal("NewP2PAPIServer returned nil")
	}

	// Проверяем, что все компоненты инициализированы
	if server.logger == nil {
		t.Error("Logger is nil")
	}

	if server.router == nil {
		t.Error("Router is nil")
	}

	if server.apiRouter == nil {
		t.Error("API Router is nil")
	}

	if server.validator == nil {
		t.Error("Validator is nil")
	}

	if server.server == nil {
		t.Error("HTTP Server is nil")
	}

	// Проверяем, что сервер настроен на правильный порт
	if server.server.Addr != ":8080" {
		t.Errorf("Expected server address :8080, got %s", server.server.Addr)
	}

	t.Log("P2P API Server created successfully")
}

func TestP2PAPIServerStartStop(t *testing.T) {
	// Создаем тестовый логгер
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// Создаем P2P API Server
	jwtSecret := []byte("test-jwt-secret")
	server := NewP2PAPIServer(logger, jwtSecret, 8080)

	if server == nil {
		t.Fatal("Failed to create P2P API Server")
	}

	// Запускаем сервер в горутине
	done := make(chan error, 1)
	go func() {
		done <- server.Start()
	}()

	// Ждем немного, чтобы сервер запустился
	time.Sleep(100 * time.Millisecond)

	// Останавливаем сервер
	err := server.Stop()
	if err != nil {
		t.Errorf("Failed to stop P2P API Server: %v", err)
	}

	// Ждем завершения горутины
	select {
	case err := <-done:
		if err != nil && err.Error() != "http: Server closed" {
			t.Errorf("Unexpected error from server: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Error("Server did not stop within timeout")
	}

	t.Log("P2P API Server start/stop test completed")
}
