package main

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// P2PClaims структура для P2P JWT токена
type P2PClaims struct {
	ProtocolType   string                 `json:"protocol_type"`
	Scope          string                 `json:"scope"`
	OrgID          string                 `json:"org_id"`
	TenantID       string                 `json:"tenant_id"`
	ServerID       string                 `json:"server_id"`
	ConnectionType string                 `json:"connection_type"`
	MaxPeers       string                 `json:"max_peers"`
	Permissions    []string               `json:"permissions"`
	NetworkConfig  map[string]interface{} `json:"network_config"`
	WireGuardConfig map[string]interface{} `json:"wireguard_config"`
	MeshConfig     map[string]interface{} `json:"mesh_config"`
	PeerWhitelist  []string               `json:"peer_whitelist"`
	jwt.RegisteredClaims
}

func main() {
	// JWT secret из Kubernetes secret
	jwtSecret := "eozy96a8+j125pOpIhCyytge1rR0MTiG4wBi/J9zpew="
	
	// Создаем claims с правильным whitelist
	claims := P2PClaims{
		ProtocolType:   "p2p-mesh",
		Scope:          "p2p-mesh-claims",
		OrgID:          "tenant-216420165",
		TenantID:       "tenant-216420165",
		ServerID:       "server-1758051692753",
		ConnectionType: "wireguard",
		MaxPeers:       "10",
		Permissions:    []string{"mesh_join", "mesh_manage"},
		NetworkConfig: map[string]interface{}{
			"subnet":        "10.0.0.0/24",
			"gateway":       "10.0.0.1",
			"dns":           []string{"8.8.8.8", "1.1.1.1"},
			"mtu":           1420,
			"firewall_rules": []string{"allow_ssh", "allow_http"},
			"enable_ipv6":   false,
		},
		WireGuardConfig: map[string]interface{}{
			"interface_name": "wg0",
			"listen_port":    51820,
			"address":        "10.0.0.100/24",
			"mtu":           1420,
			"allowed_ips":   []string{"10.0.0.0/24", "192.168.1.0/24"},
		},
		MeshConfig: map[string]interface{}{
			"network_id":           "mesh-network-001",
			"subnet":               "10.0.0.0/16",
			"registry_url":         "https://mesh-registry.2gc.ru",
			"heartbeat_interval":   "30s",
			"max_peers":           10,
			"routing_strategy":     "performance_optimal",
			"enable_auto_discovery": true,
			"trust_level":          "basic",
		},
		// ИСПРАВЛЕННЫЙ WHITELIST - добавляем наш пир
		PeerWhitelist: []string{
			"peer-001",
			"peer-002", 
			"peer-003",
			"peer_server-client-server-1758051692753", // Добавляем наш пир
		},
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "https://auth.2gc.ru/realms/cloudbridge",
			Audience:  []string{"account"},
			Subject:   "server-client-server-1758051692753",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			ID:        fmt.Sprintf("jwt_%d_%s", time.Now().Unix(), "test"),
		},
	}

	// Создаем токен
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	
	// Подписываем токен
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		fmt.Printf("Error creating token: %v\n", err)
		return
	}

	fmt.Println("Generated JWT Token with corrected whitelist:")
	fmt.Println("=============================================")
	fmt.Println(tokenString)
	fmt.Println("")
	
	// Декодируем и показываем payload для проверки
	parser := jwt.NewParser()
	parts, _, err := parser.ParseUnverified(tokenString, &P2PClaims{})
	if err == nil && parts != nil {
		payload, _ := parts.Claims.(*P2PClaims)
		fmt.Println("Token payload (for verification):")
		fmt.Printf("  Tenant ID: %s\n", payload.TenantID)
		fmt.Printf("  Subject: %s\n", payload.Subject)
		fmt.Printf("  Peer Whitelist: %v\n", payload.PeerWhitelist)
		fmt.Printf("  Connection Type: %s\n", payload.ConnectionType)
	}
}
