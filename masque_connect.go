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
	fmt.Println("🚀 MASQUE CONNECT-UDP тест")
	fmt.Println(strings.Repeat("=", 50))

	// Создаем HTTP/3 клиент
	client := &http.Client{
		Transport: &http3.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Timeout: 30 * time.Second,
	}

	// Тестируем CONNECT-UDP
	testConnectUDP(client, "localhost:8443")

	// Тестируем CONNECT-IP
	testConnectIP(client, "localhost:8443")
}

func testConnectUDP(client *http.Client, host string) {
	fmt.Printf("\n📡 Тестируем CONNECT-UDP к %s\n", host)
	fmt.Println(strings.Repeat("-", 30))

	// Создаем CONNECT-UDP запрос
	req, err := http.NewRequest("CONNECT", fmt.Sprintf("https://%s/.well-known/masque/udp/8.8.8.8:53", host), nil)
	if err != nil {
		fmt.Printf("❌ Ошибка создания CONNECT-UDP запроса: %v\n", err)
		return
	}

	// Добавляем необходимые заголовки для CONNECT-UDP
	req.Header.Set("Capsule-Protocol", "?1")
	req.Header.Set("User-Agent", "MASQUE-Test-Client/1.0")

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
	fmt.Printf("\n📡 Тестируем CONNECT-IP к %s\n", host)
	fmt.Println(strings.Repeat("-", 30))

	// Создаем CONNECT-IP запрос
	req, err := http.NewRequest("CONNECT", fmt.Sprintf("https://%s/.well-known/masque/ip/8.8.8.8", host), nil)
	if err != nil {
		fmt.Printf("❌ Ошибка создания CONNECT-IP запроса: %v\n", err)
		return
	}

	// Добавляем необходимые заголовки для CONNECT-IP
	req.Header.Set("Capsule-Protocol", "?1")
	req.Header.Set("User-Agent", "MASQUE-Test-Client/1.0")

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
