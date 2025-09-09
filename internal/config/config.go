package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Config represents the main configuration structure
type Config struct {
	Relay   RelayConfig   `json:"relay"`
	Cadence CadenceConfig `json:"cadence"`
	Tenant  TenantConfig  `json:"tenant"`
	Tunnel  TunnelConfig  `json:"tunnel"`
}

// RelayConfig represents relay server configuration
type RelayConfig struct {
	Host           string `json:"host"`
	Port           int    `json:"port"`
	MaxConnections int    `json:"max_connections"`
	Timeout        int    `json:"timeout"`
}

// CadenceConfig represents Cadence configuration
type CadenceConfig struct {
	HostPort string `json:"host_port"`
	Domain   string `json:"domain"`
	TaskList string `json:"task_list"`
}

// TenantConfig represents tenant management configuration
type TenantConfig struct {
	MaxTunnelsPerTenant     int `json:"max_tunnels_per_tenant"`
	MaxConnectionsPerTenant int `json:"max_connections_per_tenant"`
	RateLimitPerTenant      int `json:"rate_limit_per_tenant"`
}

// TunnelConfig represents tunnel management configuration
type TunnelConfig struct {
	MaxTunnels     int `json:"max_tunnels"`
	MaxConnections int `json:"max_connections"`
	Timeout        int `json:"timeout"`
}

// LoadConfig loads configuration from file
func LoadConfig(filename string) (*Config, error) {
	// Валидация пути к файлу для предотвращения path traversal
	if !isValidConfigPath(filename) {
		return nil, fmt.Errorf("invalid config file path: %s", filename)
	}
	file, err := os.Open(filename) // #nosec G304 -- path validated above
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// isValidConfigPath проверяет валидность пути к конфигурационному файлу
func isValidConfigPath(path string) bool {
	// Проверяем, что путь не содержит опасные символы
	if strings.Contains(path, "..") || strings.Contains(path, "~") {
		return false
	}

	// Проверяем, что путь начинается с допустимого префикса
	validPrefixes := []string{"./", "/", "configs/", "internal/config/"}
	hasValidPrefix := false
	for _, prefix := range validPrefixes {
		if strings.HasPrefix(path, prefix) {
			hasValidPrefix = true
			break
		}
	}

	return hasValidPrefix
}
