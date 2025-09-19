package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/quic-go/quic-go"
)

const (
	serverAddr = "localhost:9091" // P2P QUIC порт сервера
)

func main() {
	log.Println("🚀 Запуск P2P тестового клиента...")

	// Токены для двух машин
	token1 := "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJwcm90b2NvbF90eXBlIjoicDJwLW1lc2giLCJzY29wZSI6InAycC1tZXNoLWNsYWltcyIsInRlbmFudF9pZCI6InRlbmFudC0yMTY0MjAxNjUiLCJzZXJ2ZXJfaWQiOiJzZXJ2ZXItMTc1ODI0ODA0MTQ3NCIsImNvbm5lY3Rpb25fdHlwZSI6InF1aWMiLCJwZXJtaXNzaW9ucyI6WyJwMnBfY29ubmVjdCIsIm1lc2hfam9pbiIsIm1lc2hfbWFuYWdlIl0sImlhdCI6MTc1ODI0OTEzMSwiaXNzIjoiaHR0cHM6Ly9hdXRoLjJnYy5ydS9yZWFsbXMvY2xvdWRicmlkZ2UiLCJhdWQiOiJhY2NvdW50Iiwic3ViIjoic2VydmVyLTE3NTgyNDgwNDE0NzQiLCJqdGkiOiJqd3RfMTc1ODI0OTEzMV9hZmFuaDBsYSJ9.1hzU8nXygHs95fAhLV2ENfgVZQOR0HPuzFiXIHJ3loc"
	token2 := "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJwcm90b2NvbF90eXBlIjoicDJwLW1lc2giLCJzY29wZSI6InAycC1tZXNoLWNsYWltcyIsInRlbmFudF9pZCI6InRlbmFudC0yMTY0MjAxNjUiLCJzZXJ2ZXJfaWQiOiJzZXJ2ZXItMTc1ODI0ODA3NTU2NiIsImNvbm5lY3Rpb25fdHlwZSI6InF1aWMiLCJwZXJtaXNzaW9ucyI6WyJwMnBfY29ubmVjdCIsIm1lc2hfam9pbiIsIm1lc2hfbWFuYWdlIl0sImlhdCI6MTc1ODI0ODA3NywiaXNzIjoiaHR0cHM6Ly9hdXRoLjJnYy5ydS9yZWFsbXMvY2xvdWRicmlkZ2UiLCJhdWQiOiJhY2NvdW50Iiwic3ViIjoic2VydmVyLTE3NTgyNDgwNzU1NjYiLCJqdGkiOiJqd3RfMTc1ODI0ODA3N19sc3pvYm1oMiJ9.wxPDSKGMG-vI5cB-XoNlx8C5rFQslO2cSinklylfPMA"

	// Целевые peer ID для отправки сообщений
	targetPeer1 := "peer_server-1758248041474" // Машина 1
	targetPeer2 := "peer_server-1758248075566" // Машина 2

	// Тестируем соединение с токеном Машины 1
	log.Println("📡 Тестирование соединения Машины 1...")
	testConnection(token1, targetPeer2, "Машина 1")

	time.Sleep(2 * time.Second)

	// Тестируем соединение с токеном Машины 2
	log.Println("📡 Тестирование соединения Машины 2...")
	testConnection(token2, targetPeer1, "Машина 2")
}

func testConnection(token, targetPeer, machineName string) {
	// TLS конфигурация для QUIC
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, // Для тестирования
		NextProtos:         []string{"quic-p2p"},
	}

	// Подключение к QUIC серверу
	conn, err := quic.DialAddr(context.Background(), serverAddr, tlsConfig, nil)
	if err != nil {
		log.Printf("❌ %s: Ошибка подключения к QUIC серверу: %v", machineName, err)
		return
	}
	defer conn.CloseWithError(0, "client closed")
	log.Printf("✅ %s: Подключен к QUIC серверу: %s", machineName, conn.RemoteAddr())

	// 1. Аутентификация
	authStream, err := conn.OpenStreamSync(context.Background())
	if err != nil {
		log.Printf("❌ %s: Ошибка открытия auth stream: %v", machineName, err)
		return
	}
	defer authStream.Close()

	// Отправляем токен аутентификации
	authMessage := "AUTH " + token
	_, err = authStream.Write([]byte(authMessage))
	if err != nil {
		log.Printf("❌ %s: Ошибка отправки токена: %v", machineName, err)
		return
	}

	// Читаем ответ аутентификации
	authResponse := make([]byte, 1024)
	n, err := authStream.Read(authResponse)
	if err != nil {
		log.Printf("❌ %s: Ошибка чтения auth ответа: %v", machineName, err)
		return
	}

	response := string(authResponse[:n])
	if response != "AUTH_OK" {
		log.Printf("❌ %s: Аутентификация не удалась: %s", machineName, response)
		return
	}
	log.Printf("✅ %s: Аутентификация успешна!", machineName)

	// 2. Отправка P2P сообщений
	for i := 1; i <= 3; i++ {
		message := fmt.Sprintf("TO:%s:Привет от %s! Сообщение #%d - %s", 
			targetPeer, machineName, i, time.Now().Format("15:04:05"))
		
		stream, err := conn.OpenStreamSync(context.Background())
		if err != nil {
			log.Printf("❌ %s: Ошибка открытия stream: %v", machineName, err)
			continue
		}

		_, err = stream.Write([]byte(message))
		if err != nil {
			log.Printf("❌ %s: Ошибка отправки сообщения: %v", machineName, err)
			stream.Close()
			continue
		}
		
		log.Printf("📤 %s: Отправлено сообщение: %s", machineName, message)
		stream.Close()

		time.Sleep(1 * time.Second)
	}

	// 3. Слушаем входящие сообщения
	log.Printf("👂 %s: Слушаем входящие сообщения...", machineName)
	go func() {
		for {
			stream, err := conn.AcceptStream(context.Background())
			if err != nil {
				log.Printf("❌ %s: Ошибка принятия входящего stream: %v", machineName, err)
				return
			}
			
			go func(s *quic.Stream) {
				defer s.Close()
				buf := make([]byte, 4096)
				n, err := s.Read(buf)
				if err != nil && err != io.EOF {
					log.Printf("❌ %s: Ошибка чтения входящего сообщения: %v", machineName, err)
					return
				}
				log.Printf("📥 %s: Получено сообщение: %s", machineName, string(buf[:n]))
			}(stream)
		}
	}()

	// Ждем некоторое время для получения сообщений
	time.Sleep(5 * time.Second)
	log.Printf("🏁 %s: Тест завершен", machineName)
}
