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
		MaxStreams:        100,
		KeepAlive:         true,
		IdleTimeout:       30 * time.Second,
		HandshakeTimeout:  10 * time.Second,
		MaxIncomingStreams: 100,
		MaxIncomingUniStreams: 100,
		
		// Enhanced features
		TenantManager:    tenantManager,
		EnableDDoSProtection: true,
		EnableGeographicRouting: true,
		EnableKeycloakIntegration: true,
		KeycloakConfig: &KeycloakConfig{
			ServerURL:    "https://keycloak.example.com",
			Realm:        "cloudbridge",
			ClientID:     "quic-client",
			ClientSecret: "secret",
		},
	}
	
	// Пример создания соединения для тенанта
	conn, err := DialWithTenant("tcp", "127.0.0.1:8080", "tenant1", config)
	if err != nil {
		log.Fatalf("Failed to create enhanced connection: %v", err)
	}
	defer conn.Close()
	
	// Пример отправки данных
	data := []byte("Hello, Enhanced QUIC!")
	_, err = conn.Write(data)
	if err != nil {
		log.Printf("Failed to write data: %v", err)
	}
	
	// Пример чтения данных
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		log.Printf("Failed to read data: %v", err)
	} else {
		fmt.Printf("Received: %s\n", string(buffer[:n]))
	}
	
	// Получаем статистику
	stats := conn.GetStats()
	fmt.Printf("Connection stats: %+v\n", stats)
	
	// Пример географической маршрутизации
	if conn.geoRouter != nil {
		route := conn.geoRouter.RouteByLocation("192.168.1.1")
		fmt.Printf("Geographic route: %s\n", route)
	}
	
	// Пример аутентификации через Keycloak
	if conn.keycloakAuth != nil {
		authenticated, err := conn.keycloakAuth.AuthenticateUser("sample-token")
		if err != nil {
			log.Printf("Authentication failed: %v", err)
		} else {
			fmt.Printf("Authentication successful: %t\n", authenticated)
		}
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

// ExampleKeycloakIntegration демонстрирует интеграцию с Keycloak
func ExampleKeycloakIntegration() {
	keycloakConfig := &KeycloakConfig{
		ServerURL:    "https://keycloak.example.com",
		Realm:        "cloudbridge",
		ClientID:     "quic-client",
		ClientSecret: "secret",
	}
	
	auth := NewKeycloakAuth(keycloakConfig)
	
	// Тестируем аутентификацию
	token := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."
	authenticated, err := auth.AuthenticateUser(token)
	if err != nil {
		fmt.Printf("Authentication error: %v\n", err)
	} else {
		fmt.Printf("Authentication result: %t\n", authenticated)
	}
	
	// Тестируем валидацию токена
	valid, err := auth.ValidateToken(token)
	if err != nil {
		fmt.Printf("Token validation error: %v\n", err)
	} else {
		fmt.Printf("Token validation result: %t\n", valid)
	}
}
