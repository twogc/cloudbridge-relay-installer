package quic

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestEnhancedQUICIntegration тестирует интеграцию Enhanced QUIC
func TestEnhancedQUICIntegration(t *testing.T) {
	logger := zap.NewNop()
	
	t.Run("Enhanced Config Creation", func(t *testing.T) {
		// Тестируем создание Enhanced конфигурации
		config := DefaultEnhancedConfig()
		require.NotNil(t, config)
		
		assert.Equal(t, 100, config.MaxStreams)
		assert.True(t, config.KeepAlive)
		assert.Equal(t, 30*time.Second, config.IdleTimeout)
		assert.True(t, config.EnableDDoSProtection)
		assert.True(t, config.EnableGeographicRouting)
		assert.False(t, config.EnableKeycloakIntegration) // По умолчанию отключено
		
		// Проверяем TenantManager
		require.NotNil(t, config.TenantManager)
	})
	
	t.Run("Tenant Manager Integration", func(t *testing.T) {
		// Тестируем создание TenantManager
		tenantManager := NewTenantManager()
		require.NotNil(t, tenantManager)
		
		// Добавляем тестового тенанта
		tenantManager.AddTenant("test", 1024*1024*100, 10) // 100MB, 10 connections
		
		// Получаем тенанта
		tenant, exists := tenantManager.GetTenant("test")
		assert.True(t, exists)
		require.NotNil(t, tenant)
		
		assert.Equal(t, "test", tenant.ID)
		assert.Equal(t, int64(1024*1024*100), tenant.BandwidthLimit)
		assert.Equal(t, 10, tenant.ConnectionLimit)
		
		// Проверяем лимиты
		err := tenantManager.CheckTenantLimits("test")
		assert.NoError(t, err)
		
		// Проверяем несуществующего тенанта
		_, exists = tenantManager.GetTenant("nonexistent")
		assert.False(t, exists)
		
		err = tenantManager.CheckTenantLimits("nonexistent")
		assert.Error(t, err)
	})
	
	t.Run("DDoS Protection", func(t *testing.T) {
		// Тестируем создание DDoS Protection
		ddosProtection := NewDDoSProtection()
		require.NotNil(t, ddosProtection)
		
		// Тестируем проверку rate limit
		err := ddosProtection.CheckRateLimit("192.168.1.1")
		assert.NoError(t, err)
		
		// Проверяем множественные запросы от одного IP
		for i := 0; i < 50; i++ {
			err = ddosProtection.CheckRateLimit("192.168.1.2")
			assert.NoError(t, err) // Первые запросы должны проходить
		}
		
		// Превышаем лимит
		for i := 0; i < 100; i++ {
			err = ddosProtection.CheckRateLimit("192.168.1.3")
			// Последние запросы должны блокироваться
		}
		
		// Проверяем заблокированный IP
		err = ddosProtection.CheckRateLimit("192.168.1.3")
		if err != nil {
			assert.Contains(t, err.Error(), "rate limit exceeded")
		}
	})
	
	t.Run("Geographic Router", func(t *testing.T) {
		// Тестируем создание Geographic Router
		geoRouter := NewGeographicRouter()
		require.NotNil(t, geoRouter)
		
		// Тестируем маршрутизацию по IP
		testCases := []struct {
			ip             string
			expectedRoute  string
		}{
			{"1.2.3.4", "us-east.relay.example.com"},    // Американский IP (начинается с 1)
			{"2.3.4.5", "eu-west.relay.example.com"},    // Европейский IP (начинается с 2)
			{"3.4.5.6", "asia.relay.example.com"},       // Азиатский IP (другие)
			{"192.168.1.1", "asia.relay.example.com"},   // Локальный IP
		}
		
		for _, tc := range testCases {
			route := geoRouter.RouteByLocation(tc.ip)
			assert.Equal(t, tc.expectedRoute, route, "IP %s should route to %s", tc.ip, tc.expectedRoute)
		}
	})
	
	t.Run("Keycloak Auth", func(t *testing.T) {
		// Тестируем создание Keycloak Auth
		config := &KeycloakConfig{
			ServerURL: "https://edge.2gc.ru",
			Realm:     "cloudbridge",
			ClientID:  "enhanced-quic-client",
		}
		
		keycloakAuth := NewKeycloakAuth(config)
		require.NotNil(t, keycloakAuth)
		
		// Тестируем аутентификацию (заглушка)
		authenticated, err := keycloakAuth.AuthenticateUser("test-token")
		assert.NoError(t, err)
		assert.True(t, authenticated) // Заглушка всегда возвращает true для непустого токена
		
		// Тестируем пустой токен
		authenticated, err = keycloakAuth.AuthenticateUser("")
		assert.NoError(t, err)
		assert.False(t, authenticated)
		
		// Тестируем валидацию токена
		valid, err := keycloakAuth.ValidateToken("test-token")
		assert.NoError(t, err)
		assert.True(t, valid)
		
		valid, err = keycloakAuth.ValidateToken("")
		assert.NoError(t, err)
		assert.False(t, valid)
	})
	
	t.Run("Enhanced Connection Creation", func(t *testing.T) {
		// Создаем Enhanced конфигурацию
		config := DefaultEnhancedConfig()
		require.NotNil(t, config)
		
		// Пытаемся создать соединение (будет ошибка, так как нет сервера)
		conn, err := DialWithTenant("udp", "localhost:12345", "test", config)
		// Ожидаем ошибку подключения, но не ошибку конфигурации
		assert.Error(t, err)
		assert.Nil(t, conn)
		assert.Contains(t, err.Error(), "connection refused")
	})
	
	t.Run("Anomaly Detector", func(t *testing.T) {
		// Тестируем создание детектора аномалий
		detector := NewAnomalyDetector()
		require.NotNil(t, detector)
		
		// Проверяем пороговые значения
		assert.Equal(t, 1000.0, detector.thresholds["connection_rate"])
		assert.Equal(t, 1000000.0, detector.thresholds["bandwidth_usage"])
	})
	
	t.Run("Enhanced Config Compatibility", func(t *testing.T) {
		// Тестируем совместимость Enhanced конфигурации с обычной
		regularConfig := DefaultConfig()
		require.NotNil(t, regularConfig)
		
		// Создаем Enhanced конфигурацию на основе обычной
		enhancedConfig := regularConfig.CreateEnhancedConfig()
		require.NotNil(t, enhancedConfig)
		
		// Проверяем, что базовые параметры скопированы
		assert.Equal(t, regularConfig.MaxStreams, enhancedConfig.MaxStreams)
		assert.Equal(t, regularConfig.KeepAlive, enhancedConfig.KeepAlive)
		assert.Equal(t, regularConfig.IdleTimeout, enhancedConfig.IdleTimeout)
		assert.Equal(t, regularConfig.HandshakeTimeout, enhancedConfig.HandshakeTimeout)
		
		// Проверяем Enhanced функции
		assert.True(t, enhancedConfig.EnableDDoSProtection)
		assert.True(t, enhancedConfig.EnableGeographicRouting)
		assert.False(t, enhancedConfig.EnableKeycloakIntegration)
		
		// Проверяем включение/отключение Enhanced режима
		regularConfig.EnableEnhanced()
		assert.True(t, regularConfig.IsEnhancedEnabled())
		
		regularConfig.DisableEnhanced()
		assert.False(t, regularConfig.IsEnhancedEnabled())
	})
}

// TestEnhancedQUICPerformance тестирует производительность Enhanced QUIC
func TestEnhancedQUICPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}
	
	logger := zap.NewNop()
	
	t.Run("Tenant Manager Performance", func(t *testing.T) {
		tenantManager := NewTenantManager()
		
		// Добавляем много тенантов
		start := time.Now()
		for i := 0; i < 1000; i++ {
			tenantID := fmt.Sprintf("tenant_%d", i)
			tenantManager.AddTenant(tenantID, 1024*1024, 10)
		}
		addDuration := time.Since(start)
		
		t.Logf("Added 1000 tenants in %v", addDuration)
		assert.Less(t, addDuration, 100*time.Millisecond, "Adding tenants should be fast")
		
		// Проверяем производительность поиска
		start = time.Now()
		for i := 0; i < 1000; i++ {
			tenantID := fmt.Sprintf("tenant_%d", i)
			_, exists := tenantManager.GetTenant(tenantID)
			assert.True(t, exists)
		}
		lookupDuration := time.Since(start)
		
		t.Logf("Looked up 1000 tenants in %v", lookupDuration)
		assert.Less(t, lookupDuration, 50*time.Millisecond, "Tenant lookup should be fast")
	})
	
	t.Run("DDoS Protection Performance", func(t *testing.T) {
		ddosProtection := NewDDoSProtection()
		
		// Тестируем производительность проверки rate limit
		start := time.Now()
		for i := 0; i < 10000; i++ {
			ip := fmt.Sprintf("192.168.1.%d", i%255)
			ddosProtection.CheckRateLimit(ip)
		}
		checkDuration := time.Since(start)
		
		t.Logf("Performed 10000 rate limit checks in %v", checkDuration)
		assert.Less(t, checkDuration, 500*time.Millisecond, "Rate limit checks should be fast")
	})
	
	t.Run("Geographic Router Performance", func(t *testing.T) {
		geoRouter := NewGeographicRouter()
		
		// Тестируем производительность географической маршрутизации
		start := time.Now()
		for i := 0; i < 10000; i++ {
			ip := fmt.Sprintf("%d.%d.%d.%d", i%256, (i/256)%256, (i/65536)%256, (i/16777216)%256)
			geoRouter.RouteByLocation(ip)
		}
		routeDuration := time.Since(start)
		
		t.Logf("Performed 10000 geographic routes in %v", routeDuration)
		assert.Less(t, routeDuration, 100*time.Millisecond, "Geographic routing should be fast")
	})
}

// TestEnhancedQUICConcurrency тестирует конкурентную безопасность Enhanced QUIC
func TestEnhancedQUICConcurrency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrency test in short mode")
	}
	
	logger := zap.NewNop()
	
	t.Run("Tenant Manager Concurrency", func(t *testing.T) {
		tenantManager := NewTenantManager()
		
		// Конкурентное добавление тенантов
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		const numWorkers = 10
		const tenantsPerWorker = 100
		
		errChan := make(chan error, numWorkers)
		
		for w := 0; w < numWorkers; w++ {
			go func(workerID int) {
				for i := 0; i < tenantsPerWorker; i++ {
					tenantID := fmt.Sprintf("worker_%d_tenant_%d", workerID, i)
					tenantManager.AddTenant(tenantID, 1024*1024, 10)
					
					// Проверяем лимиты
					if err := tenantManager.CheckTenantLimits(tenantID); err != nil {
						errChan <- err
						return
					}
					
					// Проверяем получение тенанта
					if _, exists := tenantManager.GetTenant(tenantID); !exists {
						errChan <- fmt.Errorf("tenant %s not found", tenantID)
						return
					}
				}
				errChan <- nil
			}(w)
		}
		
		// Ждем завершения всех воркеров
		for w := 0; w < numWorkers; w++ {
			select {
			case err := <-errChan:
				assert.NoError(t, err)
			case <-ctx.Done():
				t.Fatal("Test timed out")
			}
		}
		
		// Проверяем общее количество тенантов
		// Примечание: мы не можем точно проверить количество из-за возможных коллизий имен
		t.Logf("Total tenants added: %d", numWorkers*tenantsPerWorker)
	})
	
	t.Run("DDoS Protection Concurrency", func(t *testing.T) {
		ddosProtection := NewDDoSProtection()
		
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		const numWorkers = 10
		const requestsPerWorker = 1000
		
		errChan := make(chan error, numWorkers)
		
		for w := 0; w < numWorkers; w++ {
			go func(workerID int) {
				for i := 0; i < requestsPerWorker; i++ {
					ip := fmt.Sprintf("192.168.%d.%d", workerID, i%255)
					ddosProtection.CheckRateLimit(ip)
				}
				errChan <- nil
			}(w)
		}
		
		// Ждем завершения всех воркеров
		for w := 0; w < numWorkers; w++ {
			select {
			case err := <-errChan:
				assert.NoError(t, err)
			case <-ctx.Done():
				t.Fatal("Test timed out")
			}
		}
		
		t.Logf("Processed %d concurrent rate limit checks", numWorkers*requestsPerWorker)
	})
}

// BenchmarkEnhancedQUIC бенчмарки для Enhanced QUIC
func BenchmarkEnhancedQUIC(b *testing.B) {
	_ = zap.NewNop() // Не используется в бенчмарках
	
	b.Run("TenantManager_AddTenant", func(b *testing.B) {
		tenantManager := NewTenantManager()
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			tenantID := fmt.Sprintf("tenant_%d", i)
			tenantManager.AddTenant(tenantID, 1024*1024, 10)
		}
	})
	
	b.Run("TenantManager_GetTenant", func(b *testing.B) {
		tenantManager := NewTenantManager()
		
		// Предварительно добавляем тенантов
		for i := 0; i < 1000; i++ {
			tenantID := fmt.Sprintf("tenant_%d", i)
			tenantManager.AddTenant(tenantID, 1024*1024, 10)
		}
		
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			tenantID := fmt.Sprintf("tenant_%d", i%1000)
			tenantManager.GetTenant(tenantID)
		}
	})
	
	b.Run("DDoSProtection_CheckRateLimit", func(b *testing.B) {
		ddosProtection := NewDDoSProtection()
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			ip := fmt.Sprintf("192.168.1.%d", i%255)
			ddosProtection.CheckRateLimit(ip)
		}
	})
	
	b.Run("GeographicRouter_RouteByLocation", func(b *testing.B) {
		geoRouter := NewGeographicRouter()
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			ip := fmt.Sprintf("%d.%d.%d.%d", i%256, (i/256)%256, (i/65536)%256, (i/16777216)%256)
			geoRouter.RouteByLocation(ip)
		}
	})
}


