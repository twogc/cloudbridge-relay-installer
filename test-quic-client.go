package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"time"

	"github.com/quic-go/quic-go"
)

func main() {
	// Создаем QUIC соединение
	addr := "localhost:9090"
	
	// Конфигурация TLS (без проверки сертификата для тестирования)
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"cloudbridge-p2p"},
	}

	// Подключаемся к серверу
	conn, err := quic.DialAddr(context.Background(), addr, tlsConfig, nil)
	if err != nil {
		log.Fatalf("Failed to connect to QUIC server: %v", err)
	}
	defer conn.CloseWithError(0, "client closing")

	fmt.Println("Connected to QUIC server")

	// Открываем поток для аутентификации
	authStream, err := conn.OpenStreamSync(context.Background())
	if err != nil {
		log.Fatalf("Failed to open auth stream: %v", err)
	}
	defer authStream.Close()

	// Отправляем JWT токен для аутентификации
	jwtToken := "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJwcm90b2NvbF90eXBlIjoicDJwLW1lc2giLCJzY29wZSI6InAycC1tZXNoLWNsYWltcyIsInRlbmFudF9pZCI6InRlbmFudC0yMTY0MjAxNjUiLCJzZXJ2ZXJfaWQiOiJzZXJ2ZXItY2xpZW50LXNlcnZlci0xNzU4MTA1Njg5MTY5IiwiY29ubmVjdGlvbl90eXBlIjoicXVpYyIsInBlcm1pc3Npb25zIjpbInAycF9jb25uZWN0IiwibWVzaF9qb2luIl0sImlhdCI6MTc1ODI0NTQ5OCwiaXNzIjoiaHR0cHM6Ly9hdXRoLjJnYy5ydS9yZWFsbXMvY2xvdWRicmlkZ2UiLCJhdWQiOiJhY2NvdW50Iiwic3ViIjoic2VydmVyLWNsaWVudC1zZXJ2ZXItMTc1ODEwNTY4OTE2OSIsImp0aSI6Imp3dF8xNzU4MjQ1NDk4X3VuaWZpZWQifQ.P37_9cRFA-nW_BN99K_v1QXWqA_Q8Y7m1MpbWFJ9b_Y"
	
	authMessage := fmt.Sprintf("AUTH %s", jwtToken)
	_, err = authStream.Write([]byte(authMessage))
	if err != nil {
		log.Fatalf("Failed to send auth message: %v", err)
	}

	// Читаем подтверждение аутентификации
	buffer := make([]byte, 1024)
	n, err := authStream.Read(buffer)
	if err != nil {
		log.Fatalf("Failed to read auth response: %v", err)
	}

	response := string(buffer[:n])
	fmt.Printf("Auth response: %s\n", response)

	if response != "AUTH_OK" {
		log.Fatalf("Authentication failed: %s", response)
	}

	fmt.Println("Authentication successful!")

	// Открываем поток для отправки P2P сообщения
	dataStream, err := conn.OpenStreamSync(context.Background())
	if err != nil {
		log.Fatalf("Failed to open data stream: %v", err)
	}
	defer dataStream.Close()

	// Отправляем тестовое P2P сообщение
	// Формат: "TO:<peer_id>:<payload>"
	targetPeerID := "peer_server-client-server-1758105689169"
	message := fmt.Sprintf("TO:%s:Hello from QUIC client!", targetPeerID)
	
	_, err = dataStream.Write([]byte(message))
	if err != nil {
		log.Fatalf("Failed to send P2P message: %v", err)
	}

	fmt.Printf("Sent P2P message: %s\n", message)

	// Ждем немного
	time.Sleep(2 * time.Second)

	fmt.Println("Test completed successfully!")
}
