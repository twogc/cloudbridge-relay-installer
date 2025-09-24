package zitadel

import (
	"time"
)

// Config конфигурация для Zitadel валидатора
type Config struct {
	Domain         string `json:"domain"`          // https://zitadel.local
	ProjectID      string `json:"project_id"`      // ID проекта в Zitadel
	ClientID       string `json:"client_id"`       // ID приложения
	ClientSecret   string `json:"client_secret"`   // Секрет приложения
	FallbackSecret string `json:"fallback_secret"` // Fallback для совместимости
}

// BaseTokenClaims базовые claims для всех токенов (совместимость с Keycloak)
type BaseTokenClaims struct {
	ProtocolType string `json:"protocol_type"`
	TenantID     string `json:"tenant_id"`
	ServerID     string `json:"server_id"`
	Issuer       string `json:"iss"`
	Subject      string `json:"sub"`
	ExpiresAt    int64  `json:"exp"`
	IssuedAt     int64  `json:"iat"`
	Scope        string `json:"scope"`
}

// P2PMeshTokenClaims claims для P2P Mesh протокола
type P2PMeshTokenClaims struct {
	BaseTokenClaims
	MeshConfig string `json:"mesh_config"`
	MaxPeers   string `json:"max_peers"`
}

// QuickTokenClaims claims для Quick протокола
type QuickTokenClaims struct {
	BaseTokenClaims
	TargetServer   string `json:"target_server"`
	MaxConnections string `json:"max_connections"`
	MaxTunnels     string `json:"max_tunnels"`
	MaxBandwidth   string `json:"max_bandwidth"`
}

// TenantInfo информация о тенанте
type TenantInfo struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Domain    string                 `json:"domain"`
	State     string                 `json:"state"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`

	// P2P настройки
	MaxPeers       int `json:"max_peers"`
	MaxConnections int `json:"max_connections"`
	MaxTunnels     int `json:"max_tunnels"`

	// Квоты и лимиты
	BandwidthLimit string `json:"bandwidth_limit"`
	StorageLimit   string `json:"storage_limit"`
}

// Вспомогательные функции
func getIntFromConfig(config map[string]interface{}, key string, defaultValue int) int {
	if value, ok := config[key].(int); ok {
		return value
	}
	return defaultValue
}

func getStringFromConfig(config map[string]interface{}, key string, defaultValue string) string {
	if value, ok := config[key].(string); ok {
		return value
	}
	return defaultValue
}



