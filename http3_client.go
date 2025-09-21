package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/quic-go/quic-go/http3"
)

func main() {
	fmt.Println("🚀 HTTP/3 клиент для тестирования MASQUE сервера")
	fmt.Println(strings.Repeat("=", 50))

	// Тестовые URL
	testURLs := []string{
		"https://192.168.49.2:30443/health",
		"https://edge.2gc.ru:8443/health",
		"https://192.168.49.2:30443/.well-known/masque/udp/",
		"https://edge.2gc.ru:8443/.well-known/masque/udp/",
	}

	// Создаем HTTP/3 клиент
	client := &http.Client{
		Transport: &http3.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // Для тестирования
			},
		},
		Timeout: 10 * time.Second,
	}

	for _, url := range testURLs {
		fmt.Printf("\n📡 Тестируем: %s\n", url)
		fmt.Println(strings.Repeat("-", 30))

		// Создаем контекст с таймаутом
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Создаем запрос
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			fmt.Printf("❌ Ошибка создания запроса: %v\n", err)
			continue
		}

		// Добавляем заголовки
		req.Header.Set("User-Agent", "HTTP3-Test-Client/1.0")
		req.Header.Set("Accept", "*/*")

		// Выполняем запрос
		resp, err := client.Do(req)
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

		// Выводим результат
		fmt.Printf("✅ Статус: %d %s\n", resp.StatusCode, resp.Status)
		fmt.Printf("✅ Заголовки: %v\n", resp.Header)
		fmt.Printf("✅ Тело ответа: %s\n", string(body[:min(len(body), 200)]))

		time.Sleep(1 * time.Second)
	}

	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("🏁 Тестирование завершено")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
