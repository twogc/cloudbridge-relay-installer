package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/quic-go/quic-go/http3"
)

func main() {
	fmt.Println("🚀 Продвинутый HTTP/3 клиент для тестирования MASQUE")
	fmt.Println(strings.Repeat("=", 60))

	// Создаем HTTP/3 клиент с расширенными настройками
	client := &http.Client{
		Transport: &http3.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
				ServerName:         "edge.2gc.ru",
			},
		},
		Timeout: 30 * time.Second,
	}

	// Тестовые сценарии
	testScenarios := []struct {
		name string
		url  string
		desc string
	}{
		{
			name: "Health Check",
			url:  "https://192.168.49.2:30443/health",
			desc: "Проверка состояния MASQUE сервера",
		},
		{
			name: "MASQUE UDP Endpoint",
			url:  "https://192.168.49.2:30443/.well-known/masque/udp/",
			desc: "Проверка CONNECT-UDP endpoint",
		},
		{
			name: "MASQUE IP Endpoint", 
			url:  "https://192.168.49.2:30443/.well-known/masque/ip/",
			desc: "Проверка CONNECT-IP endpoint",
		},
		{
			name: "MASQUE Datagrams",
			url:  "https://192.168.49.2:30443/.well-known/masque/datagrams/",
			desc: "Проверка HTTP Datagrams endpoint",
		},
	}

	// Выполняем тесты
	for i, scenario := range testScenarios {
		fmt.Printf("\n📡 Тест %d: %s\n", i+1, scenario.name)
		fmt.Printf("   Описание: %s\n", scenario.desc)
		fmt.Printf("   URL: %s\n", scenario.url)
		fmt.Println(strings.Repeat("-", 50))

		// Создаем запрос
		req, err := http.NewRequest("GET", scenario.url, nil)
		if err != nil {
			fmt.Printf("❌ Ошибка создания запроса: %v\n", err)
			continue
		}

		// Добавляем заголовки
		req.Header.Set("User-Agent", "Advanced-HTTP3-Client/1.0")
		req.Header.Set("Accept", "*/*")
		req.Header.Set("Connection", "keep-alive")

		// Выполняем запрос
		start := time.Now()
		resp, err := client.Do(req)
		duration := time.Since(start)

		if err != nil {
			fmt.Printf("❌ Ошибка HTTP/3 запроса: %v\n", err)
			continue
		}
		defer resp.Body.Close()

		// Читаем ответ
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("❌ Ошибка чтения ответа: %v\n", err)
			continue
		}

		// Выводим результаты
		fmt.Printf("✅ Статус: %d %s\n", resp.StatusCode, resp.Status)
		fmt.Printf("✅ Время ответа: %v\n", duration)
		fmt.Printf("✅ Размер ответа: %d байт\n", len(body))
		fmt.Printf("✅ Заголовки:\n")
		for name, values := range resp.Header {
			for _, value := range values {
				fmt.Printf("   %s: %s\n", name, value)
			}
		}
		fmt.Printf("✅ Тело ответа:\n%s\n", string(body))

		time.Sleep(1 * time.Second)
	}

	// Тестируем CONNECT-UDP
	fmt.Printf("\n📡 Тест CONNECT-UDP\n")
	fmt.Println(strings.Repeat("-", 50))
	testConnectUDP(client, "192.168.49.2:30443")

	// Тестируем CONNECT-IP
	fmt.Printf("\n📡 Тест CONNECT-IP\n")
	fmt.Println(strings.Repeat("-", 50))
	testConnectIP(client, "192.168.49.2:30443")

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🏁 Тестирование завершено")
}

func testConnectUDP(client *http.Client, host string) {
	// Создаем CONNECT-UDP запрос
	req, err := http.NewRequest("CONNECT", fmt.Sprintf("https://%s/.well-known/masque/udp/8.8.8.8:53", host), nil)
	if err != nil {
		fmt.Printf("❌ Ошибка создания CONNECT-UDP запроса: %v\n", err)
		return
	}

	// Добавляем необходимые заголовки для CONNECT-UDP
	req.Header.Set("Capsule-Protocol", "?1")
	req.Header.Set("User-Agent", "Advanced-HTTP3-Client/1.0")

	// Выполняем запрос
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ Ошибка CONNECT-UDP запроса: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("✅ CONNECT-UDP статус: %d %s\n", resp.StatusCode, resp.Status)
	fmt.Printf("✅ Заголовки: %v\n", resp.Header)

	// Читаем ответ
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ Ошибка чтения ответа: %v\n", err)
		return
	}

	fmt.Printf("✅ Тело ответа: %s\n", string(body))
}

func testConnectIP(client *http.Client, host string) {
	// Создаем CONNECT-IP запрос
	req, err := http.NewRequest("CONNECT", fmt.Sprintf("https://%s/.well-known/masque/ip/8.8.8.8", host), nil)
	if err != nil {
		fmt.Printf("❌ Ошибка создания CONNECT-IP запроса: %v\n", err)
		return
	}

	// Добавляем необходимые заголовки для CONNECT-IP
	req.Header.Set("Capsule-Protocol", "?1")
	req.Header.Set("User-Agent", "Advanced-HTTP3-Client/1.0")

	// Выполняем запрос
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ Ошибка CONNECT-IP запроса: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("✅ CONNECT-IP статус: %d %s\n", resp.StatusCode, resp.Status)
	fmt.Printf("✅ Заголовки: %v\n", resp.Header)

	// Читаем ответ
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ Ошибка чтения ответа: %v\n", err)
		return
	}

	fmt.Printf("✅ Тело ответа: %s\n", string(body))
}
