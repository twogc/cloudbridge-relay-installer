package quic

import (
	"fmt"
	"log"
	"time"
)

// ExampleEnhancedQUIC демонстрирует использование Enhanced QUIC
func ExampleEnhancedQUIC() {
	// Создаем менеджер тенантов
	tenantManager := NewTenantManager()

	// Добавляем тенантов
	tenantManager.AddTenant("tenant1", 1024*1024*100, 10) // 100MB, 10 соединений
	tenantManager.AddTenant("tenant2", 1024*1024*50, 5)   // 50MB, 5 соединений

	// Создаем конфигурацию Enhanced QUIC
	config := &EnhancedConfig{
		MaxStreams:            100,
		KeepAlive:             true,
		IdleTimeout:           30 * time.Second,
		HandshakeTimeout:      10 * time.Second,
		MaxIncomingStreams:    100,
		MaxIncomingUniStreams: 100,

		// Enhanced features
		TenantManager:           tenantManager,
		EnableDDoSProtection:    true,
		EnableGeographicRouting: true,
	}

	// Пример создания соединения для тенанта
	conn, err := NewEnhancedPolicies("tenant1", config)
	if err != nil {
		log.Fatalf("Failed to create enhanced connection: %v", err)
	}
	defer conn.Close()

	// Пример использования хуков для отправки данных
	data := []byte("Hello, Enhanced QUIC!")
	remoteAddr := "192.168.1.100:8080"

	// Проверяем лимиты перед записью
	if err := conn.BeforeWrite(remoteAddr, len(data)); err != nil {
		log.Printf("Write blocked by policies: %v", err)
	} else {
		// В реальном коде здесь был бы вызов quic.Stream.Write(data)
		fmt.Printf("Data would be written: %s\n", string(data))
		conn.AfterWrite(len(data))
	}

	// Пример использования хуков для чтения данных
	buffer := make([]byte, 1024)

	// Проверяем лимиты перед чтением
	if err := conn.BeforeRead(remoteAddr, len(buffer)); err != nil {
		log.Printf("Read blocked by policies: %v", err)
	} else {
		// В реальном коде здесь был бы вызов quic.Stream.Read(buffer)
		// Симулируем успешное чтение
		n := 20
		fmt.Printf("Data would be read: %d bytes\n", n)
		conn.AfterRead(n)
	}

	// Получаем статистику
	stats := conn.GetStats()
	fmt.Printf("Connection stats: %+v\n", stats)

	// Пример географической маршрутизации
	if conn.geoRouter != nil {
		route := conn.geoRouter.RouteByLocation("192.168.1.1")
		fmt.Printf("Geographic route: %s\n", route)
	}
}

// ExampleDDoSProtection демонстрирует DDoS защиту
func ExampleDDoSProtection() {
	ddosProtection := NewDDoSProtection()

	// Симулируем атаку
	clientIP := "192.168.1.100"

	for i := 0; i < 150; i++ {
		err := ddosProtection.CheckRateLimit(clientIP)
		if err != nil {
			fmt.Printf("Rate limit exceeded after %d requests: %v\n", i+1, err)
			break
		}
	}
}

// ExampleTenantManagement демонстрирует управление тенантами
func ExampleTenantManagement() {
	tenantManager := NewTenantManager()

	// Добавляем тенантов
	tenantManager.AddTenant("premium", 1024*1024*1000, 100) // 1GB, 100 соединений
	tenantManager.AddTenant("standard", 1024*1024*100, 10)  // 100MB, 10 соединений

	// Проверяем лимиты
	err := tenantManager.CheckTenantLimits("premium")
	if err != nil {
		fmt.Printf("Premium tenant limit check: %v\n", err)
	} else {
		fmt.Println("Premium tenant limit check: OK")
	}

	// Получаем информацию о тенанте
	if tenant, exists := tenantManager.GetTenant("standard"); exists {
		fmt.Printf("Standard tenant: %+v\n", tenant)
	}
}

// ExampleGeographicRouting демонстрирует географическую маршрутизацию
func ExampleGeographicRouting() {
	router := NewGeographicRouter()

	// Тестируем маршрутизацию для разных IP
	testIPs := []string{"192.168.1.1", "10.0.0.1", "172.16.0.1"}

	for _, ip := range testIPs {
		route := router.RouteByLocation(ip)
		fmt.Printf("IP %s -> Route: %s\n", ip, route)
	}
}
