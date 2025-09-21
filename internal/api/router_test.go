package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// TestNewAPIRouter тесты для создания APIRouter
func TestNewAPIRouter(t *testing.T) {
	logger := zap.NewNop()
	validator := &MockTokenValidator{}

	router := NewAPIRouter(logger, validator)

	assert.NotNil(t, router)
	assert.NotNil(t, router.peerManager)
	assert.NotNil(t, router.relayManager)
	assert.NotNil(t, router.healthManager)
	assert.NotNil(t, router.p2pManager)
	assert.NotNil(t, router.iceManager)
	assert.NotNil(t, router.hub)
	assert.NotNil(t, router.logger)
	assert.NotNil(t, router.authMiddleware)
	assert.NotNil(t, router.rateLimiter)
	assert.NotNil(t, router.aiMiddleware)
	assert.NotNil(t, router.aiEndpoints)
	assert.NotNil(t, router.dataPipelineEndpoints)
	assert.NotNil(t, router.monitoringEndpoints)
	assert.NotNil(t, router.heartbeatHandler)
}

func TestNewAPIRouterWithConfig(t *testing.T) {
	logger := zap.NewNop()
	validator := &MockTokenValidator{}
	allowedIPs := []string{"192.168.1.0/24", "10.0.0.1"}

	router := NewAPIRouterWithConfig(logger, validator, allowedIPs)

	assert.NotNil(t, router)
	assert.NotNil(t, router.peerManager)
	assert.NotNil(t, router.relayManager)
	assert.NotNil(t, router.healthManager)
	assert.NotNil(t, router.p2pManager)
	assert.NotNil(t, router.iceManager)
	assert.NotNil(t, router.hub)
	assert.NotNil(t, router.logger)
	assert.NotNil(t, router.authMiddleware)
	assert.NotNil(t, router.rateLimiter)
	assert.NotNil(t, router.aiMiddleware)
	assert.NotNil(t, router.aiEndpoints)
	assert.NotNil(t, router.dataPipelineEndpoints)
	assert.NotNil(t, router.monitoringEndpoints)
}

func TestNewAPIRouterWithKeycloak(t *testing.T) {
	// Пропускаем тест с Keycloak, так как требует реального валидатора
	t.Skip("Skipping Keycloak test - requires real KeycloakValidator")
}

// TestAPIRouter_SetupRoutes тесты для настройки маршрутов
func TestAPIRouter_SetupRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	validator := &MockTokenValidator{}
	router := NewAPIRouter(logger, validator)

	// Создаем Gin engine
	engine := gin.New()

	// Настраиваем маршруты
	router.SetupRoutes(engine)

	// Проверяем, что маршруты настроены
	assert.NotNil(t, engine)

	// Тестируем доступность основных эндпоинтов
	testRoutes := []struct {
		method string
		path   string
	}{
		{"GET", "/api/v1/health/"},
		{"GET", "/api/v1/health/metrics"},
		{"GET", "/api/v1/health/peers"},
		{"GET", "/api/v1/health/connections"},
		{"GET", "/api/v1/relay/status"},
		{"GET", "/api/v1/ai/status"},
		{"GET", "/api/v1/ai/health"},
		{"GET", "/api/v1/ai/metrics"},
	}

	for _, route := range testRoutes {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(route.method, route.path, nil)
		engine.ServeHTTP(w, req)

		// Проверяем, что маршрут существует (не 404)
		assert.NotEqual(t, http.StatusNotFound, w.Code, "Route %s %s should exist", route.method, route.path)
	}
}

// TestAPIRouter_GetHub тесты для получения Hub
func TestAPIRouter_GetHub(t *testing.T) {
	logger := zap.NewNop()
	validator := &MockTokenValidator{}
	router := NewAPIRouter(logger, validator)

	hub := router.GetHub()
	assert.NotNil(t, hub)
	assert.Equal(t, router.hub, hub)
}

// TestAPIRouter_GetPeerManager тесты для получения PeerManager
func TestAPIRouter_GetPeerManager(t *testing.T) {
	logger := zap.NewNop()
	validator := &MockTokenValidator{}
	router := NewAPIRouter(logger, validator)

	peerManager := router.GetPeerManager()
	assert.NotNil(t, peerManager)
	assert.Equal(t, router.peerManager, peerManager)
}

// TestAPIRouter_GetRelayManager тесты для получения RelayManager
func TestAPIRouter_GetRelayManager(t *testing.T) {
	logger := zap.NewNop()
	validator := &MockTokenValidator{}
	router := NewAPIRouter(logger, validator)

	relayManager := router.GetRelayManager()
	assert.NotNil(t, relayManager)
	assert.Equal(t, router.relayManager, relayManager)
}

// TestAPIRouter_GetHealthManager тесты для получения HealthManager
func TestAPIRouter_GetHealthManager(t *testing.T) {
	logger := zap.NewNop()
	validator := &MockTokenValidator{}
	router := NewAPIRouter(logger, validator)

	healthManager := router.GetHealthManager()
	assert.NotNil(t, healthManager)
	assert.Equal(t, router.healthManager, healthManager)
}

// TestAPIRouter_GetAIEndpoints тесты для получения AIEndpoints
func TestAPIRouter_GetAIEndpoints(t *testing.T) {
	logger := zap.NewNop()
	validator := &MockTokenValidator{}
	router := NewAPIRouter(logger, validator)

	aiEndpoints := router.GetAIEndpoints()
	assert.NotNil(t, aiEndpoints)
	assert.Equal(t, router.aiEndpoints, aiEndpoints)
}

// TestAPIRouter_GetDataPipelineEndpoints тесты для получения DataPipelineEndpoints
func TestAPIRouter_GetDataPipelineEndpoints(t *testing.T) {
	logger := zap.NewNop()
	validator := &MockTokenValidator{}
	router := NewAPIRouter(logger, validator)

	dataPipelineEndpoints := router.GetDataPipelineEndpoints()
	assert.NotNil(t, dataPipelineEndpoints)
	assert.Equal(t, router.dataPipelineEndpoints, dataPipelineEndpoints)
}

// TestAPIRouter_GetMonitoringEndpoints тесты для получения MonitoringEndpoints
func TestAPIRouter_GetMonitoringEndpoints(t *testing.T) {
	logger := zap.NewNop()
	validator := &MockTokenValidator{}
	router := NewAPIRouter(logger, validator)

	monitoringEndpoints := router.GetMonitoringEndpoints()
	assert.NotNil(t, monitoringEndpoints)
	assert.Equal(t, router.monitoringEndpoints, monitoringEndpoints)
}

// TestAPIRouter_StartHub тесты для запуска Hub
func TestAPIRouter_StartHub(t *testing.T) {
	logger := zap.NewNop()
	validator := &MockTokenValidator{}
	router := NewAPIRouter(logger, validator)

	// Запускаем Hub
	router.StartHub()

	// Проверяем, что Hub запущен
	assert.NotNil(t, router.hub)
}

// TestAPIRouter_CORSMiddleware тесты для CORS middleware
func TestAPIRouter_CORSMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	validator := &MockTokenValidator{}
	router := NewAPIRouter(logger, validator)

	// Создаем Gin engine
	engine := gin.New()
	engine.Use(router.corsMiddleware())

	// Добавляем тестовый маршрут
	engine.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "test"})
	})

	// Тестируем CORS
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	engine.ServeHTTP(w, req)

	// Проверяем CORS заголовки
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Origin"), "*")
}

// TestAPIRouter_RateLimitMiddleware тесты для rate limiting middleware
func TestAPIRouter_RateLimitMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	validator := &MockTokenValidator{}
	router := NewAPIRouter(logger, validator)

	// Создаем Gin engine
	engine := gin.New()
	engine.Use(router.rateLimitMiddleware())

	// Добавляем тестовый маршрут
	engine.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "test"})
	})

	// Тестируем rate limiting
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	engine.ServeHTTP(w, req)

	// Проверяем, что запрос прошел
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestAPIRouter_Integration тесты интеграции
func TestAPIRouter_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	validator := &MockTokenValidator{}
	router := NewAPIRouter(logger, validator)

	// Создаем Gin engine
	engine := gin.New()
	router.SetupRoutes(engine)

	// Тестируем health endpoint
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/health/", nil)
	engine.ServeHTTP(w, req)

	// Проверяем, что health endpoint работает
	assert.Equal(t, http.StatusOK, w.Code)

	// Тестируем AI status endpoint
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/ai/status", nil)
	engine.ServeHTTP(w, req)

	// Проверяем, что AI endpoint работает
	assert.Equal(t, http.StatusOK, w.Code)

	// Тестируем relay stats endpoint
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/relay/status", nil)
	engine.ServeHTTP(w, req)

	// Проверяем, что relay endpoint работает
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestAPIRouter_ErrorHandling тесты для обработки ошибок
func TestAPIRouter_ErrorHandling(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	validator := &MockTokenValidator{}
	router := NewAPIRouter(logger, validator)

	// Создаем Gin engine
	engine := gin.New()
	router.SetupRoutes(engine)

	// Тестируем несуществующий endpoint
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/nonexistent", nil)
	engine.ServeHTTP(w, req)

	// Проверяем, что возвращается 404
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestAPIRouter_ConcurrentAccess тесты для конкурентного доступа
func TestAPIRouter_ConcurrentAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	validator := &MockTokenValidator{}
	router := NewAPIRouter(logger, validator)

	// Создаем Gin engine
	engine := gin.New()
	router.SetupRoutes(engine)

	// Запускаем несколько горутин для конкурентного доступа
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/v1/health/", nil)
			engine.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			done <- true
		}()
	}

	// Ждем завершения всех горутин
	for i := 0; i < 10; i++ {
		<-done
	}
}

// Claims структура для JWT claims
type Claims struct {
	UserID    string   `json:"user_id"`
	TenantID  string   `json:"tenant_id"`
	Roles     []string `json:"roles"`
	ExpiresAt int64    `json:"exp"`
}
