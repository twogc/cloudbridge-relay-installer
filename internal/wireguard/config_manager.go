package wireguard

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
)

// ConfigManager управляет WireGuard конфигурациями
type ConfigManager struct {
	configPath string
	logger     *zap.Logger
}

// NewConfigManager создает новый менеджер конфигураций
func NewConfigManager(configPath string, logger *zap.Logger) *ConfigManager {
	return &ConfigManager{
		configPath: configPath,
		logger:     logger,
	}
}

// GenerateServerConfig генерирует конфигурацию сервера WireGuard
func (cm *ConfigManager) GenerateServerConfig(tenants []*WireGuardTenant) (*WireGuardConfig, error) {
	config := &WireGuardConfig{
		InterfaceName: "wg0",
		PrivateKey:    cm.generateServerPrivateKey(),
		Address:       "10.0.0.1/24", // Серверный IP
		ListenPort:    51820,
		Peers:        []WireGuardPeer{},
	}
	
	// Добавляем peer'ов для каждого tenant'а
	for _, tenant := range tenants {
		if tenant.IsActive {
			peer := WireGuardPeer{
				PublicKey:           tenant.PublicKey,
				AllowedIPs:          tenant.AllowedIPs,
				Endpoint:            fmt.Sprintf("%s:%d", tenant.Endpoint, tenant.Port),
				PersistentKeepalive: tenant.PersistentKeepalive,
			}
			config.Peers = append(config.Peers, peer)
		}
	}
	
	return config, nil
}

// GenerateClientConfig генерирует конфигурацию клиента для tenant'а
func (cm *ConfigManager) GenerateClientConfig(tenant *WireGuardTenant, serverPublicKey string) string {
	config := fmt.Sprintf(`[Interface]
PrivateKey = %s
Address = %s
DNS = 8.8.8.8

[Peer]
PublicKey = %s
Endpoint = %s
AllowedIPs = 0.0.0.0/0
PersistentKeepalive = %d
`, 
		tenant.PrivateKey,
		cm.getClientIP(tenant),
		serverPublicKey,
		tenant.Endpoint,
		tenant.PersistentKeepalive)
	
	return config
}

// SaveConfig сохраняет конфигурацию в файл
func (cm *ConfigManager) SaveConfig(config *WireGuardConfig, filename string) error {
	configContent := cm.configToString(config)
	
	filePath := filepath.Join(cm.configPath, filename)
	
	err := os.WriteFile(filePath, []byte(configContent), 0600)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	
	cm.logger.Info("Saved WireGuard config",
		zap.String("file", filePath),
		zap.Int("peers", len(config.Peers)))
	
	return nil
}

// LoadConfig загружает конфигурацию из файла
func (cm *ConfigManager) LoadConfig(filename string) (*WireGuardConfig, error) {
	filePath := filepath.Join(cm.configPath, filename)
	
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	
	config, err := cm.parseConfig(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	
	return config, nil
}

// ApplyConfig применяет конфигурацию WireGuard
func (cm *ConfigManager) ApplyConfig(config *WireGuardConfig) error {
	// Сохраняем конфигурацию
	configFile := fmt.Sprintf("wg0-%d.conf", time.Now().Unix())
	err := cm.SaveConfig(config, configFile)
	if err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	
	// Применяем конфигурацию через wg-quick
	filePath := filepath.Join(cm.configPath, configFile)
	cmd := exec.Command("wg-quick", "up", filePath)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		cm.logger.Error("Failed to apply WireGuard config",
			zap.String("config", configFile),
			zap.String("output", string(output)),
			zap.Error(err))
		return fmt.Errorf("failed to apply config: %w", err)
	}
	
	cm.logger.Info("Applied WireGuard config",
		zap.String("config", configFile),
		zap.String("interface", config.InterfaceName))
	
	return nil
}

// RemoveConfig удаляет конфигурацию WireGuard
func (cm *ConfigManager) RemoveConfig(interfaceName string) error {
	cmd := exec.Command("wg-quick", "down", interfaceName)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		cm.logger.Error("Failed to remove WireGuard config",
			zap.String("interface", interfaceName),
			zap.String("output", string(output)),
			zap.Error(err))
		return fmt.Errorf("failed to remove config: %w", err)
	}
	
	cm.logger.Info("Removed WireGuard config",
		zap.String("interface", interfaceName))
	
	return nil
}

// GetInterfaceStatus возвращает статус WireGuard интерфейса
func (cm *ConfigManager) GetInterfaceStatus(interfaceName string) (map[string]interface{}, error) {
	cmd := exec.Command("wg", "show", interfaceName)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get interface status: %w", err)
	}
	
	status := cm.parseWireGuardOutput(string(output))
	return status, nil
}

// generateServerPrivateKey генерирует приватный ключ сервера
func (cm *ConfigManager) generateServerPrivateKey() string {
	// В реальной реализации здесь должна быть генерация ключа
	// Для упрощения возвращаем заглушку
	return "SERVER_PRIVATE_KEY_PLACEHOLDER"
}

// getClientIP возвращает IP адрес клиента
func (cm *ConfigManager) getClientIP(tenant *WireGuardTenant) string {
	// Простая логика назначения IP - в production нужна более сложная
	// Используем последние цифры tenant ID для уникальности
	hash := len(tenant.TenantID) % 254
	return fmt.Sprintf("10.0.0.%d/24", hash+2)
}

// configToString конвертирует конфигурацию в строку
func (cm *ConfigManager) configToString(config *WireGuardConfig) string {
	var sb strings.Builder
	
	sb.WriteString(fmt.Sprintf("[Interface]\n"))
	sb.WriteString(fmt.Sprintf("PrivateKey = %s\n", config.PrivateKey))
	sb.WriteString(fmt.Sprintf("Address = %s\n", config.Address))
	sb.WriteString(fmt.Sprintf("ListenPort = %d\n", config.ListenPort))
	sb.WriteString(fmt.Sprintf("PostUp = iptables -A FORWARD -i %s -j ACCEPT; iptables -A FORWARD -o %s -j ACCEPT; iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE\n", config.InterfaceName, config.InterfaceName))
	sb.WriteString(fmt.Sprintf("PostDown = iptables -D FORWARD -i %s -j ACCEPT; iptables -D FORWARD -o %s -j ACCEPT; iptables -t nat -D POSTROUTING -o eth0 -j MASQUERADE\n", config.InterfaceName, config.InterfaceName))
	sb.WriteString(fmt.Sprintf("\n"))
	
	for i, peer := range config.Peers {
		sb.WriteString(fmt.Sprintf("# Peer %d\n", i+1))
		sb.WriteString(fmt.Sprintf("[Peer]\n"))
		sb.WriteString(fmt.Sprintf("PublicKey = %s\n", peer.PublicKey))
		sb.WriteString(fmt.Sprintf("AllowedIPs = %s\n", strings.Join(peer.AllowedIPs, ", ")))
		if peer.Endpoint != "" {
			sb.WriteString(fmt.Sprintf("Endpoint = %s\n", peer.Endpoint))
		}
		if peer.PersistentKeepalive > 0 {
			sb.WriteString(fmt.Sprintf("PersistentKeepalive = %d\n", peer.PersistentKeepalive))
		}
		sb.WriteString(fmt.Sprintf("\n"))
	}
	
	return sb.String()
}

// parseConfig парсит конфигурацию из строки
func (cm *ConfigManager) parseConfig(content string) (*WireGuardConfig, error) {
	// Простая реализация парсинга - в production нужна более надежная
	config := &WireGuardConfig{
		InterfaceName: "wg0",
		Peers:        []WireGuardPeer{},
	}
	
	lines := strings.Split(content, "\n")
	var currentSection string
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = line[1 : len(line)-1]
			continue
		}
		
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		
		switch currentSection {
		case "Interface":
			switch key {
			case "PrivateKey":
				config.PrivateKey = value
			case "Address":
				config.Address = value
			case "ListenPort":
				// Парсинг порта
			}
		case "Peer":
			// Обработка peer'ов
		}
	}
	
	return config, nil
}

// parseWireGuardOutput парсит вывод команды wg show
func (cm *ConfigManager) parseWireGuardOutput(output string) map[string]interface{} {
	status := make(map[string]interface{})
	
	lines := strings.Split(output, "\n")
	var currentPeer string
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		if strings.HasPrefix(line, "interface:") {
			status["interface"] = strings.TrimSpace(strings.TrimPrefix(line, "interface:"))
		} else if strings.HasPrefix(line, "public key:") {
			currentPeer = strings.TrimSpace(strings.TrimPrefix(line, "public key:"))
			status["peers"] = make(map[string]interface{})
		} else if strings.HasPrefix(line, "endpoint:") {
			if currentPeer != "" {
				if peers, ok := status["peers"].(map[string]interface{}); ok {
					peers[currentPeer] = map[string]interface{}{
						"endpoint": strings.TrimSpace(strings.TrimPrefix(line, "endpoint:")),
					}
				}
			}
		}
	}
	
	return status
}
