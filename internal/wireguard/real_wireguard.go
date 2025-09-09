package wireguard

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// RealWireGuardAdapter адаптер для работы с реальным WireGuard
type RealWireGuardAdapter struct {
	interfaceName string
	configPath    string
	socketPath    string
	isRunning     bool
	mu            sync.RWMutex
	logger        *zap.Logger
	ctx           context.Context
	cancel        context.CancelFunc
}

// WireGuardConfig конфигурация для реального WireGuard
type WireGuardConfig struct {
	Interface struct {
		Name       string `json:"name"`
		ListenPort int    `json:"listen_port"`
		Address    string `json:"address"`
		PrivateKey string `json:"private_key"`
		MTU        int    `json:"mtu"`
	} `json:"interface"`
	Peers []WireGuardPeerConfig `json:"peers"`
}

// WireGuardPeerConfig конфигурация пира для реального WireGuard
type WireGuardPeerConfig struct {
	PublicKey           string   `json:"public_key"`
	AllowedIPs          []string `json:"allowed_ips"`
	Endpoint            string   `json:"endpoint"`
	PresharedKey        string   `json:"preshared_key,omitempty"`
	PersistentKeepalive int      `json:"persistent_keepalive,omitempty"`
}

// NewRealWireGuardAdapter создает новый адаптер для реального WireGuard
func NewRealWireGuardAdapter(interfaceName string, logger *zap.Logger) *RealWireGuardAdapter {
	ctx, cancel := context.WithCancel(context.Background())

	return &RealWireGuardAdapter{
		interfaceName: interfaceName,
		configPath:    fmt.Sprintf("/etc/wireguard/%s.conf", interfaceName),
		socketPath:    fmt.Sprintf("/var/run/wireguard/%s.sock", interfaceName),
		logger:        logger,
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Start запускает реальный WireGuard интерфейс
func (rwa *RealWireGuardAdapter) Start(config *WireGuardConfig) error {
	rwa.mu.Lock()
	defer rwa.mu.Unlock()

	if rwa.isRunning {
		return fmt.Errorf("WireGuard interface %s already running", rwa.interfaceName)
	}

	// Создаем конфигурационный файл
	if err := rwa.createConfigFile(config); err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}

	// Запускаем wireguard-go
	if err := rwa.startWireGuardGo(); err != nil {
		return fmt.Errorf("failed to start wireguard-go: %w", err)
	}

	// Ждем создания интерфейса
	if err := rwa.waitForInterface(); err != nil {
		return fmt.Errorf("failed to wait for interface: %w", err)
	}

	// Настраиваем интерфейс
	if err := rwa.configureInterface(config); err != nil {
		return fmt.Errorf("failed to configure interface: %w", err)
	}

	rwa.isRunning = true
	rwa.logger.Info("Real WireGuard interface started",
		zap.String("interface", rwa.interfaceName),
		zap.String("config_path", rwa.configPath),
	)

	return nil
}

// Stop останавливает реальный WireGuard интерфейс
func (rwa *RealWireGuardAdapter) Stop() error {
	rwa.mu.Lock()
	defer rwa.mu.Unlock()

	if !rwa.isRunning {
		return fmt.Errorf("WireGuard interface %s not running", rwa.interfaceName)
	}

	// Удаляем интерфейс
	if err := rwa.removeInterface(); err != nil {
		rwa.logger.Error("Failed to remove interface", zap.Error(err))
	}

	// Останавливаем wireguard-go
	if err := rwa.stopWireGuardGo(); err != nil {
		rwa.logger.Error("Failed to stop wireguard-go", zap.Error(err))
	}

	// Удаляем конфигурационный файл
	if err := os.Remove(rwa.configPath); err != nil {
		rwa.logger.Error("Failed to remove config file", zap.Error(err))
	}

	rwa.isRunning = false
	rwa.logger.Info("Real WireGuard interface stopped",
		zap.String("interface", rwa.interfaceName),
	)

	return nil
}

// AddPeer добавляет пира в реальный WireGuard
func (rwa *RealWireGuardAdapter) AddPeer(peer *WireGuardPeerConfig) error {
	rwa.mu.RLock()
	defer rwa.mu.RUnlock()

	if !rwa.isRunning {
		return fmt.Errorf("WireGuard interface %s not running", rwa.interfaceName)
	}

	// Используем wg команду для добавления пира
	// Валидация входных данных
	if !isValidInterfaceName(rwa.interfaceName) {
		return fmt.Errorf("invalid interface name: %s", rwa.interfaceName)
	}
	if !isValidPublicKey(peer.PublicKey) {
		return fmt.Errorf("invalid public key: %s", peer.PublicKey)
	}
	cmd := exec.Command("wg", "set", rwa.interfaceName, "peer", peer.PublicKey) // #nosec G204 -- inputs validated above

	if len(peer.AllowedIPs) > 0 {
		cmd.Args = append(cmd.Args, "allowed-ips", strings.Join(peer.AllowedIPs, ","))
	}

	if peer.Endpoint != "" {
		cmd.Args = append(cmd.Args, "endpoint", peer.Endpoint)
	}

	if peer.PresharedKey != "" {
		cmd.Args = append(cmd.Args, "preshared-key", peer.PresharedKey)
	}

	if peer.PersistentKeepalive > 0 {
		cmd.Args = append(cmd.Args, "persistent-keepalive", strconv.Itoa(peer.PersistentKeepalive))
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to add peer: %w, output: %s", err, string(output))
	}

	rwa.logger.Info("Peer added to real WireGuard",
		zap.String("interface", rwa.interfaceName),
		zap.String("public_key", peer.PublicKey),
		zap.Strings("allowed_ips", peer.AllowedIPs),
	)

	return nil
}

// RemovePeer удаляет пира из реального WireGuard
func (rwa *RealWireGuardAdapter) RemovePeer(publicKey string) error {
	rwa.mu.RLock()
	defer rwa.mu.RUnlock()

	if !rwa.isRunning {
		return fmt.Errorf("WireGuard interface %s not running", rwa.interfaceName)
	}

	// Валидация входных данных
	if !isValidInterfaceName(rwa.interfaceName) {
		return fmt.Errorf("invalid interface name: %s", rwa.interfaceName)
	}
	if !isValidPublicKey(publicKey) {
		return fmt.Errorf("invalid public key: %s", publicKey)
	}
	cmd := exec.Command("wg", "set", rwa.interfaceName, "peer", publicKey, "remove") // #nosec G204 -- inputs validated above
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to remove peer: %w, output: %s", err, string(output))
	}

	rwa.logger.Info("Peer removed from real WireGuard",
		zap.String("interface", rwa.interfaceName),
		zap.String("public_key", publicKey),
	)

	return nil
}

// GetStatus получает статус реального WireGuard интерфейса
func (rwa *RealWireGuardAdapter) GetStatus() (map[string]interface{}, error) {
	rwa.mu.RLock()
	defer rwa.mu.RUnlock()

	if !rwa.isRunning {
		return map[string]interface{}{
			"interface": rwa.interfaceName,
			"status":    "stopped",
		}, nil
	}

	// Получаем информацию о интерфейсе
	// Валидация имени интерфейса
	if !isValidInterfaceName(rwa.interfaceName) {
		return nil, fmt.Errorf("invalid interface name: %s", rwa.interfaceName)
	}
	cmd := exec.Command("wg", "show", rwa.interfaceName) // #nosec G204 -- interface name validated above
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	// Парсим вывод wg show
	status := rwa.parseWgShowOutput(string(output))
	status["interface"] = rwa.interfaceName
	status["status"] = "running"
	status["config_path"] = rwa.configPath
	status["socket_path"] = rwa.socketPath

	return status, nil
}

// createConfigFile создает конфигурационный файл для WireGuard
func (rwa *RealWireGuardAdapter) createConfigFile(config *WireGuardConfig) error {
	// Создаем директорию если не существует
	configDir := filepath.Dir(rwa.configPath)
	if err := os.MkdirAll(configDir, 0750); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Формируем содержимое конфигурационного файла
	content := fmt.Sprintf(`[Interface]
PrivateKey = %s
Address = %s
ListenPort = %d
MTU = %d

`, config.Interface.PrivateKey, config.Interface.Address, config.Interface.ListenPort, config.Interface.MTU)

	// Добавляем пиров
	for _, peer := range config.Peers {
		content += fmt.Sprintf(`[Peer]
PublicKey = %s
AllowedIPs = %s
`, peer.PublicKey, strings.Join(peer.AllowedIPs, ","))

		if peer.Endpoint != "" {
			content += fmt.Sprintf("Endpoint = %s\n", peer.Endpoint)
		}
		if peer.PresharedKey != "" {
			content += fmt.Sprintf("PresharedKey = %s\n", peer.PresharedKey)
		}
		if peer.PersistentKeepalive > 0 {
			content += fmt.Sprintf("PersistentKeepalive = %d\n", peer.PersistentKeepalive)
		}
		content += "\n"
	}

	// Записываем файл
	if err := os.WriteFile(rwa.configPath, []byte(content), 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	rwa.logger.Info("WireGuard config file created",
		zap.String("path", rwa.configPath),
	)

	return nil
}

// startWireGuardGo запускает wireguard-go процесс
func (rwa *RealWireGuardAdapter) startWireGuardGo() error {
	// Проверяем наличие wireguard-go
	_, err := exec.LookPath("wireguard-go")
	if err != nil {
		return fmt.Errorf("wireguard-go not found in PATH: %w", err)
	}

	// Создаем директорию для сокета
	socketDir := filepath.Dir(rwa.socketPath)
	if err := os.MkdirAll(socketDir, 0750); err != nil {
		return fmt.Errorf("failed to create socket directory: %w", err)
	}

	// Запускаем wireguard-go в фоне
	// Валидация имени интерфейса для предотвращения injection
	if !isValidInterfaceName(rwa.interfaceName) {
		return fmt.Errorf("invalid interface name: %s", rwa.interfaceName)
	}
	cmd := exec.Command("wireguard-go", rwa.interfaceName) // #nosec G204 -- interface name validated above
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start wireguard-go: %w", err)
	}

	// Запускаем goroutine для мониторинга процесса
	go func() {
		if err := cmd.Wait(); err != nil {
			rwa.logger.Error("wireguard-go process exited with error", zap.Error(err))
		}
	}()

	rwa.logger.Info("wireguard-go process started",
		zap.String("interface", rwa.interfaceName),
		zap.Int("pid", cmd.Process.Pid),
	)

	return nil
}

// stopWireGuardGo останавливает wireguard-go процесс
func (rwa *RealWireGuardAdapter) stopWireGuardGo() error {
	// Удаляем сокет для остановки wireguard-go
	if err := os.Remove(rwa.socketPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove socket: %w", err)
	}

	rwa.logger.Info("wireguard-go process stopped",
		zap.String("interface", rwa.interfaceName),
	)

	return nil
}

// waitForInterface ждет создания интерфейса
func (rwa *RealWireGuardAdapter) waitForInterface() error {
	timeout := 10 * time.Second
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if _, err := net.InterfaceByName(rwa.interfaceName); err == nil {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("timeout waiting for interface %s", rwa.interfaceName)
}

// configureInterface настраивает сетевой интерфейс
func (rwa *RealWireGuardAdapter) configureInterface(config *WireGuardConfig) error {
	// Включаем интерфейс
	// Валидация имени интерфейса
	if !isValidInterfaceName(rwa.interfaceName) {
		return fmt.Errorf("invalid interface name: %s", rwa.interfaceName)
	}
	cmd := exec.Command("ip", "link", "set", rwa.interfaceName, "up") // #nosec G204 -- interface name validated above
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to bring up interface: %w, output: %s", err, string(output))
	}

	// Устанавливаем MTU
	if config.Interface.MTU > 0 {
		// Валидация имени интерфейса уже выполнена выше
		cmd = exec.Command("ip", "link", "set", rwa.interfaceName, "mtu", strconv.Itoa(config.Interface.MTU)) // #nosec G204 -- interface name validated above
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to set MTU: %w, output: %s", err, string(output))
		}
	}

	rwa.logger.Info("Network interface configured",
		zap.String("interface", rwa.interfaceName),
		zap.Int("mtu", config.Interface.MTU),
	)

	return nil
}

// removeInterface удаляет сетевой интерфейс
func (rwa *RealWireGuardAdapter) removeInterface() error {
	// Валидация имени интерфейса
	if !isValidInterfaceName(rwa.interfaceName) {
		return fmt.Errorf("invalid interface name: %s", rwa.interfaceName)
	}
	cmd := exec.Command("ip", "link", "del", rwa.interfaceName) // #nosec G204 -- interface name validated above
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to remove interface: %w, output: %s", err, string(output))
	}

	rwa.logger.Info("Network interface removed",
		zap.String("interface", rwa.interfaceName),
	)

	return nil
}

// isValidInterfaceName проверяет валидность имени интерфейса
func isValidInterfaceName(name string) bool {
	// Проверяем, что имя содержит только допустимые символы
	if len(name) == 0 || len(name) > 15 {
		return false
	}

	// Проверяем, что имя начинается с буквы и содержит только буквы, цифры и дефисы
	if !((name[0] >= 'a' && name[0] <= 'z') || (name[0] >= 'A' && name[0] <= 'Z')) {
		return false
	}

	for _, char := range name {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') || char == '-') {
			return false
		}
	}

	return true
}

// isValidPublicKey проверяет валидность публичного ключа WireGuard
func isValidPublicKey(key string) bool {
	// WireGuard публичные ключи должны быть base64 строкой длиной 44 символа
	if len(key) != 44 {
		return false
	}

	// Проверяем, что это валидная base64 строка
	_, err := base64.StdEncoding.DecodeString(key)
	return err == nil
}

// parseWgShowOutput парсит вывод команды wg show
func (rwa *RealWireGuardAdapter) parseWgShowOutput(output string) map[string]interface{} {
	status := make(map[string]interface{})
	lines := strings.Split(output, "\n")

	var currentPeer string
	peers := make(map[string]map[string]interface{})

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "peer:") {
			// Новый пир
			currentPeer = strings.TrimSpace(strings.TrimPrefix(line, "peer:"))
			peers[currentPeer] = make(map[string]interface{})
		} else if currentPeer != "" && strings.Contains(line, ":") {
			// Свойство пира
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				peers[currentPeer][key] = value
			}
		} else if strings.Contains(line, ":") {
			// Свойство интерфейса
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				status[key] = value
			}
		}
	}

	status["peers"] = peers
	return status
}

// Close корректно завершает работу адаптера
func (rwa *RealWireGuardAdapter) Close() error {
	rwa.cancel()
	return rwa.Stop()
}
