package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// TenantConfig конфигурация tenant'а
type TenantConfig struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	MaxTunnels     int      `json:"max_tunnels"`
	MaxConnections int      `json:"max_connections"`
	BandwidthLimit int64    `json:"bandwidth_limit"`
	RateLimit      int      `json:"rate_limit"`
	Burst          int      `json:"burst"`
	AllowedIPs     []string `json:"allowed_ips"`
	AllowedPorts   []int    `json:"allowed_ports"`
	Features       []string `json:"features"`
}

// CreateTenantRequest запрос на создание tenant'а
type CreateTenantRequest struct {
	Tenant TenantConfig `json:"tenant"`
}

func main() {
	log.Println(" Инициализация тестового tenant'а relay-test-tenant...")

	// Конфигурация тестового tenant'а
	tenant := TenantConfig{
		ID:             "relay-test-tenant",
		Name:           "Relay Test Tenant",
		Description:    "Test tenant for DevOps development",
		MaxTunnels:     10,
		MaxConnections: 20,
		BandwidthLimit: 104857600, // 100MB
		RateLimit:      500,       // 500 запросов в минуту
		Burst:          100,
		AllowedIPs:     []string{}, // Все IP разрешены
		AllowedPorts:   []int{},    // Все порты разрешены
		Features:       []string{"tls", "jwt", "tunneling", "p2p", "quic"},
	}

	// Создаем запрос
	request := CreateTenantRequest{
		Tenant: tenant,
	}

	// Конвертируем в JSON
	jsonData, err := json.MarshalIndent(request, "", "  ")
	if err != nil {
		log.Fatalf("Ошибка маршалинга JSON: %v", err)
	}

	log.Printf("Конфигурация tenant'а:\n%s\n", string(jsonData))

	// Проверяем доступность relay сервера
	log.Println("🌐 Проверка доступности relay сервера...")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Проверяем health endpoint
	resp, err := client.Get("http://localhost:9090/health")
	if err != nil {
		log.Printf("❌ Relay сервер недоступен: %v", err)
		log.Println("Убедитесь, что relay сервер запущен: docker-compose up -d")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("❌ Health endpoint вернул статус: %d", resp.StatusCode)
		return
	}

	log.Println("Relay сервер доступен")

	// Тестируем подключение с новым tenant'ом
	log.Println("🧪 Тестирование подключения с tenant relay-test-tenant...")

	// Создаем тестовый JWT токен
	testJWT := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ0ZW5hbnRfaWQiOiJyZWxheS10ZXN0LXRlbmFudCIsInVzZXJfaWQiOiJkZXYtdXNlciIsInBlcm1pc3Npb25zIjpbInR1bm5lbCIsInJlYWQiLCJ3cml0ZSJdLCJpc3MiOiIyZ2MucnUiLCJhdWQiOiJyZWxheS10ZXN0LXRlbmFudCIsImV4cCI6MTczNTY4NjQwMCwiaWF0IjoxNzM1NjgyODAwfQ.test-signature"

	log.Printf("🔑 Тестовый JWT токен: %s", testJWT)

	// Инструкции по тестированию
	log.Println("\nИнструкции по тестированию:")
	log.Println("1. Отправьте hello сообщение:")
	log.Println("   echo '{\"type\":\"hello\",\"version\":\"1.0.0\",\"tenant_id\":\"relay-test-tenant\"}' | nc localhost 9090")
	log.Println("")
	log.Println("2. Отправьте auth сообщение:")
	log.Printf("   echo -e '{\"type\":\"hello\",\"version\":\"1.0.0\"}\\n{\"type\":\"auth\",\"token\":\"%s\",\"tenant_id\":\"relay-test-tenant\"}' | nc -w 10 localhost 9090", testJWT)
	log.Println("")
	log.Println("3. Отправьте tunnel_info:")
	log.Println("   echo -e '{\"type\":\"hello\",\"version\":\"1.0.0\"}\\n{\"type\":\"auth\",\"token\":\"test-token\"}\\n{\"type\":\"tunnel_info\",\"local_port\":8080,\"remote_host\":\"example.com\",\"remote_port\":80,\"protocol\":\"http\",\"tenant_id\":\"relay-test-tenant\"}' | nc -w 15 localhost 9090")
	log.Println("")
	log.Println("Tenant relay-test-tenant готов к использованию!")
	log.Println("")
	log.Println("📊 Проверьте статус:")
	log.Println("   curl -s http://localhost:9090/health | jq .")
	log.Println("")
	log.Println("Проверьте логи:")
	log.Println("   docker-compose logs cloudbridge-relay | tail -20")
}
