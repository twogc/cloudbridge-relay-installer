package main

import (
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// P2PClaims структура для P2P токенов
type P2PClaims struct {
	Subject        string   `json:"sub"`
	TenantID       string   `json:"tenant_id"`
	ServerID       string   `json:"server_id"`
	ConnectionType string   `json:"connection_type"`
	ProtocolType   string   `json:"protocol_type"`
	Permissions    []string `json:"permissions"`
	jwt.RegisteredClaims
}

func main() {
	// JWT секрет (должен совпадать с сервером)
	secret := "Wahm7Iti1VAoXsFAV1MBeMfPv1yope2IJ47utB30Pk4FgZyJFZK0ayfX4zNstkUh"
	
	// Создаем токен для Server A
	serverAToken := createP2PToken("server-a", "mesh-network-test", secret)
	
	// Создаем токен для Server B  
	serverBToken := createP2PToken("server-b", "mesh-network-test", secret)
	
	fmt.Println("=== CloudBridge P2P Tokens ===")
	fmt.Println()
	fmt.Println("Server A Token:")
	fmt.Println(serverAToken)
	fmt.Println()
	fmt.Println("Server B Token:")
	fmt.Println(serverBToken)
	fmt.Println()
	fmt.Println("=== Configuration Files ===")
	fmt.Println()
	fmt.Println("config-test-server-a.yaml:")
	fmt.Println("auth:")
	fmt.Println("  type: \"jwt\"")
	fmt.Printf("  token: \"%s\"\n", serverAToken)
	fmt.Println("  secret: \"Wahm7Iti1VAoXsFAV1MBeMfPv1yope2IJ47utB30Pk4FgZyJFZK0ayfX4zNstkUh\"")
	fmt.Println()
	fmt.Println("config-test-server-b.yaml:")
	fmt.Println("auth:")
	fmt.Println("  type: \"jwt\"")
	fmt.Printf("  token: \"%s\"\n", serverBToken)
	fmt.Println("  secret: \"Wahm7Iti1VAoXsFAV1MBeMfPv1yope2IJ47utB30Pk4FgZyJFZK0ayfX4zNstkUh\"")
	fmt.Println()
	fmt.Println("=== Token Details ===")
	fmt.Println("Audience: relay-test-tenant")
	fmt.Println("Expires: Never (no expiration)")
	fmt.Println("Algorithm: HS256")
	fmt.Println("Purpose: P2P mesh connectivity")
}

func createP2PToken(serverID, tenantID, secret string) string {
	// Создаем claims без срока истечения
	claims := P2PClaims{
		Subject:        serverID,
		TenantID:       tenantID,
		ServerID:       serverID,
		ConnectionType: "quic",
		ProtocolType:   "p2p-mesh",
		Permissions:    []string{"relay_connect", "relay_tunnel", "p2p_discover", "p2p_connect"},
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "cloudbridge-relay",
			Audience:  []string{"relay-test-tenant"},
			Subject:   serverID,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			// НЕ устанавливаем ExpiresAt - токен бессрочный
		},
	}
	
	// Создаем токен
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	
	// Подписываем токен
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		log.Fatalf("Error creating token for %s: %v", serverID, err)
	}
	
	return tokenString
}
