//go:build ignore

package main

// Этот файл демонстрирует использование enhanced компонентов
// Для запуска используйте: go run cmd/relay/enhanced_components_demo_test.go
// Убедитесь, что вы находитесь в корневой директории проекта

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/config"
	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
)

// Демонстрация использования улучшенных компонентов P2P QUIC сервера
func main() {
	fmt.Println("CloudBridge Relay - Enhanced Components Demo")
	fmt.Println("================================================")

	// Создаем логгер
	logger, err := logger.New(logger.Config{
		Level:  logger.INFO,
		Output: "stdout",
		JSON:   false,
	})
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}

	// Создаем конфигурацию
	cfg := createDemoConfig()

	// Демонстрация StatelessResetManager
	fmt.Println("\n1. StatelessResetManager Demo")
	fmt.Println("--------------------------------")
	demoStatelessResetManager(logger, cfg)

	// Демонстрация ICE Support
	fmt.Println("\n2. ICE/STUN Support Demo")
	fmt.Println("---------------------------")
	demoICESupport(logger, cfg)

	// Демонстрация MASQUE Support
	fmt.Println("\n3. MASQUE Support Demo")
	fmt.Println("-------------------------")
	demoMASQUESupport(logger, cfg)

	fmt.Println("\nDemo completed successfully!")
}

func createDemoConfig() *config.Config {
	cfg := &config.Config{}

	// ICE/STUN конфигурация
	cfg.ICEStunTurn.Enabled = true
	cfg.ICEStunTurn.StunServers = []string{
		"stun.l.google.com:19302",
		"stun1.l.google.com:19302",
	}
	cfg.ICEStunTurn.ICETimeout = 30 * time.Second
	cfg.ICEStunTurn.ConnectionTimeout = 10 * time.Second
	cfg.ICEStunTurn.GatheringTimeout = 5 * time.Second
	cfg.ICEStunTurn.CandidateTypes = []string{"host", "srflx", "relay"}

	// MASQUE конфигурация
	cfg.MASQUE.Enabled = true
	cfg.MASQUE.ProxyURL = "https://auth.2gc.ru:8443"
	cfg.MASQUE.ALPNProtocols = []string{"h3", "h3-29", "h3-28", "h3-27"}
	cfg.MASQUE.FallbackEnabled = true

	return cfg
}

func demoStatelessResetManager(logger *logger.Logger, cfg *config.Config) {
	fmt.Println("Creating StatelessResetManager...")

	// Создаем StatelessResetManager
	srm := NewStatelessResetManager(logger, "/tmp/demo_stateless_reset.key")

	// Загружаем или генерируем ключ
	if err := srm.LoadOrGenerateKey(); err != nil {
		fmt.Printf("Failed to load/generate key: %v\n", err)
		return
	}

	// Получаем информацию о ключе
	info := srm.GetKeyInfo()
	fmt.Printf("StatelessResetKey created successfully\n")
	fmt.Printf("   - Key file: %s\n", info["key_file"])
	fmt.Printf("   - Key length: %d bytes\n", info["key_length"])
	fmt.Printf("   - Has file: %t\n", info["has_file"])

	// Валидируем ключ
	if err := srm.ValidateKey(); err != nil {
		fmt.Printf("Key validation failed: %v\n", err)
		return
	}

	fmt.Printf("Key validation passed\n")
}

func demoICESupport(logger *logger.Logger, cfg *config.Config) {
	fmt.Println("Creating ICE Support...")

	// Создаем ICE поддержку
	iceSupport, err := NewICESupport(logger, cfg)
	if err != nil {
		fmt.Printf("Failed to create ICE support: %v\n", err)
		return
	}

	// Запускаем ICE поддержку
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := iceSupport.Start(ctx); err != nil {
		fmt.Printf("Failed to start ICE support: %v\n", err)
		return
	}

	fmt.Printf("ICE support started successfully\n")
	fmt.Printf("   - STUN servers: %v\n", cfg.ICEStunTurn.StunServers)
	fmt.Printf("   - ICE timeout: %v\n", cfg.ICEStunTurn.ICETimeout)

	// Ждем немного для сбора кандидатов
	fmt.Println("   - Gathering ICE candidates...")
	time.Sleep(3 * time.Second)

	// Получаем кандидаты
	candidates := iceSupport.GetLocalCandidates()
	fmt.Printf("Gathered %d ICE candidates\n", len(candidates))

	// Закрываем ICE поддержку
	if err := iceSupport.Close(); err != nil {
		fmt.Printf("Failed to close ICE support: %v\n", err)
		return
	}

	fmt.Printf("ICE support closed successfully\n")
}

func demoMASQUESupport(logger *logger.Logger, cfg *config.Config) {
	fmt.Println("Creating MASQUE Server...")

	// Создаем MASQUE сервер
	masqueServer := NewMASQUEServer(logger, cfg)

	// Запускаем MASQUE сервер
	if err := masqueServer.Start(); err != nil {
		fmt.Printf("Failed to start MASQUE server: %v\n", err)
		return
	}

	fmt.Printf("MASQUE server started successfully\n")
	fmt.Printf("   - Proxy URL: %s\n", cfg.MASQUE.ProxyURL)
	fmt.Printf("   - ALPN protocols: %v\n", cfg.MASQUE.ALPNProtocols)
	fmt.Printf("   - Fallback enabled: %t\n", cfg.MASQUE.FallbackEnabled)

	// Получаем информацию о соединениях
	info := masqueServer.GetConnectionInfo()
	fmt.Printf("   - Total connections: %d\n", info["total_connections"])
	fmt.Printf("   - Server running: %t\n", info["is_running"])

	// Закрываем MASQUE сервер
	if err := masqueServer.Close(); err != nil {
		fmt.Printf("Failed to close MASQUE server: %v\n", err)
		return
	}

	fmt.Printf("MASQUE server closed successfully\n")
}
