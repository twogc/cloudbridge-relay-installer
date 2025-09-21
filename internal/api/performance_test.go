package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
	"go.uber.org/zap"
)

// BenchmarkRateLimiter_Performance тестирует производительность rate limiter
func BenchmarkRateLimiter_Performance(b *testing.B) {
	config := DefaultRateLimiterConfig()
	logger := zap.NewNop()
	rateLimiter := NewRateLimiter(config, logger)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(rateLimiter.CreateRateLimitMiddleware("test"))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/test", nil)
			router.ServeHTTP(w, req)
		}
	})
}

// BenchmarkWebSocketHub_ConcurrentConnections тестирует производительность WebSocket Hub при множественных соединениях
func BenchmarkWebSocketHub_ConcurrentConnections(b *testing.B) {
	logger := zap.NewNop()
	hub := NewHub(logger)

	// Запускаем hub
	go hub.Run()
	defer func() {
		close(hub.Register)
		close(hub.Unregister)
		close(hub.Broadcast)
	}()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		connID := 0
		for pb.Next() {
			connID++
			conn := &WebSocketConnection{
				PeerID: fmt.Sprintf("peer_%d", connID),
				Send:   make(chan []byte, 256),
			}

			// Регистрируем соединение
			select {
			case hub.Register <- conn:
			case <-time.After(100 * time.Millisecond):
				b.Fatal("Timeout registering connection")
			}

			// Отменяем регистрацию
			select {
			case hub.Unregister <- conn:
			case <-time.After(100 * time.Millisecond):
				b.Fatal("Timeout unregistering connection")
			}
		}
	})
}

// BenchmarkAPIServer_ConcurrentRequests тестирует производительность API сервера при множественных запросах
func BenchmarkAPIServer_ConcurrentRequests(b *testing.B) {
	config := Config{
		Port:    ":8080",
		Logger:  createPerformanceTestLogger(),
		Version: "1.0.0",
	}
	server := NewServer(config)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/health", gin.WrapF(server.handleHealth))
	router.GET("/api/v1/stats", gin.WrapF(server.handleStats))

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/health", nil)
			router.ServeHTTP(w, req)
		}
	})
}

// TestPerformance_WebSocketHub_MessageThroughput тестирует пропускную способность WebSocket Hub
func TestPerformance_WebSocketHub_MessageThroughput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	logger := zap.NewNop()
	hub := NewHub(logger)

	// Запускаем hub
	go hub.Run()
	defer func() {
		close(hub.Register)
		close(hub.Unregister)
		close(hub.Broadcast)
	}()

	// Создаем множество соединений
	numConnections := 100
	connections := make([]*WebSocketConnection, numConnections)

	for i := 0; i < numConnections; i++ {
		conn := &WebSocketConnection{
			PeerID: fmt.Sprintf("peer_%d", i),
			Send:   make(chan []byte, 1000),
		}
		connections[i] = conn
		hub.Register <- conn
	}

	// Ждем регистрации всех соединений
	time.Sleep(100 * time.Millisecond)

	// Измеряем время отправки сообщений
	start := time.Now()
	numMessages := 1000

	for i := 0; i < numMessages; i++ {
		message := []byte(fmt.Sprintf("message_%d", i))
		hub.Broadcast <- message
	}

	// Ждем обработки всех сообщений
	time.Sleep(500 * time.Millisecond)
	duration := time.Since(start)

	// Вычисляем пропускную способность
	throughput := float64(numMessages) / duration.Seconds()
	t.Logf("Message throughput: %.2f messages/second", throughput)

	// Проверяем, что пропускная способность разумная (больше 1000 сообщений/сек)
	assert.Greater(t, throughput, 1000.0, "Throughput should be greater than 1000 messages/second")

	// Очищаем соединения
	for _, conn := range connections {
		hub.Unregister <- conn
	}
}

// TestPerformance_RateLimiter_ConcurrentAccess тестирует производительность rate limiter при конкурентном доступе
func TestPerformance_RateLimiter_ConcurrentAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	config := DefaultRateLimiterConfig()
	logger := zap.NewNop()
	rateLimiter := NewRateLimiter(config, logger)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(rateLimiter.CreateRateLimitMiddleware("test"))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	numGoroutines := 100
	numRequests := 1000
	var wg sync.WaitGroup

	start := time.Now()

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numRequests/numGoroutines; j++ {
				w := httptest.NewRecorder()
				req := httptest.NewRequest("GET", "/test", nil)
				router.ServeHTTP(w, req)
			}
		}()
	}

	wg.Wait()
	duration := time.Since(start)

	throughput := float64(numRequests) / duration.Seconds()
	t.Logf("Rate limiter throughput: %.2f requests/second", throughput)

	// Проверяем, что пропускная способность разумная
	assert.Greater(t, throughput, 100.0, "Rate limiter throughput should be greater than 100 requests/second")
}

// TestPerformance_RelayManager_ConcurrentRouting тестирует производительность RelayManager при конкурентной маршрутизации
func TestPerformance_RelayManager_ConcurrentRouting(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	logger := zap.NewNop()
	hub := NewHub(logger)
	relayManager := &RelayManager{
		hub:    hub,
		peers:  make(map[string]*Peer),
		logger: logger,
	}

	gin.SetMode(gin.TestMode)

	numGoroutines := 50
	numRequests := 500
	var wg sync.WaitGroup

	start := time.Now()

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < numRequests/numGoroutines; j++ {
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)

				requestBody := map[string]interface{}{
					"from_peer": fmt.Sprintf("peer_%d", goroutineID),
					"to_peer":   fmt.Sprintf("peer_%d", (goroutineID+1)%numGoroutines),
					"data":      fmt.Sprintf("data_%d_%d", goroutineID, j),
				}
				reqBody, _ := json.Marshal(requestBody)
				req := httptest.NewRequest("POST", "/relay", bytes.NewReader(reqBody))
				req.Header.Set("Content-Type", "application/json")
				c.Request = req

				relayManager.RouteThroughRelay(c)
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	throughput := float64(numRequests) / duration.Seconds()
	t.Logf("Relay manager throughput: %.2f requests/second", throughput)

	// Проверяем, что пропускная способность разумная
	assert.Greater(t, throughput, 50.0, "Relay manager throughput should be greater than 50 requests/second")
}

// TestPerformance_MemoryUsage тестирует использование памяти при длительной работе
func TestPerformance_MemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	logger := zap.NewNop()
	hub := NewHub(logger)

	// Запускаем hub
	go hub.Run()
	defer func() {
		close(hub.Register)
		close(hub.Unregister)
		close(hub.Broadcast)
	}()

	// Создаем и удаляем множество соединений
	numIterations := 1000
	numConnectionsPerIteration := 10

	for i := 0; i < numIterations; i++ {
		connections := make([]*WebSocketConnection, numConnectionsPerIteration)

		// Создаем соединения
		for j := 0; j < numConnectionsPerIteration; j++ {
			conn := &WebSocketConnection{
				PeerID: fmt.Sprintf("peer_%d_%d", i, j),
				Send:   make(chan []byte, 256),
			}
			connections[j] = conn
			hub.Register <- conn
		}

		// Ждем регистрации
		time.Sleep(10 * time.Millisecond)

		// Отправляем сообщения
		for k := 0; k < 5; k++ {
			message := []byte(fmt.Sprintf("message_%d_%d", i, k))
			hub.Broadcast <- message
		}

		// Удаляем соединения
		for _, conn := range connections {
			hub.Unregister <- conn
		}

		// Ждем очистки
		time.Sleep(10 * time.Millisecond)

		// Проверяем, что соединения действительно удалены
		if i%100 == 0 {
			connectionCount := hub.GetConnectionCount()
			assert.Equal(t, 0, connectionCount, "All connections should be cleaned up")
		}
	}

	t.Log("Memory usage test completed successfully")
}

// TestPerformance_ConcurrentAPIRequests тестирует производительность API при множественных одновременных запросах
func TestPerformance_ConcurrentAPIRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	config := Config{
		Port:    ":8080",
		Logger:  createPerformanceTestLogger(),
		Version: "1.0.0",
	}
	server := NewServer(config)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/health", gin.WrapF(server.handleHealth))
	router.GET("/api/v1/stats", gin.WrapF(server.handleStats))
	router.GET("/api/v1/tunnels", gin.WrapF(server.handleTunnels))
	router.GET("/api/v1/clients", gin.WrapF(server.handleClients))

	endpoints := []string{"/health", "/api/v1/stats", "/api/v1/tunnels", "/api/v1/clients"}
	numGoroutines := 50
	requestsPerGoroutine := 20
	var wg sync.WaitGroup

	start := time.Now()

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < requestsPerGoroutine; j++ {
				endpoint := endpoints[j%len(endpoints)]
				w := httptest.NewRecorder()
				req := httptest.NewRequest("GET", endpoint, nil)
				router.ServeHTTP(w, req)

				// Проверяем, что запрос обработан успешно
				assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusTooManyRequests)
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	totalRequests := numGoroutines * requestsPerGoroutine
	throughput := float64(totalRequests) / duration.Seconds()
	t.Logf("API throughput: %.2f requests/second", throughput)

	// Проверяем, что пропускная способность разумная
	assert.Greater(t, throughput, 100.0, "API throughput should be greater than 100 requests/second")
}

// createPerformanceTestLogger создает logger для тестов производительности
func createPerformanceTestLogger() *logger.Logger {
	config := logger.Config{
		Level:  logger.INFO, // Используем INFO для тестов производительности
		Output: "stdout",
		JSON:   true,
	}
	log, _ := logger.New(config)
	return log
}
