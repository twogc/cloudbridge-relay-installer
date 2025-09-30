package main

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	jwtSecret = "85Sk/NfLq3gqzCXzmBKbJCpL+f5BssXz3G8dVi3sPiE=" // Секрет с сервера
	issuer    = "cloudbridge-relay"
	audience  = "relay-test-tenant"
	tenantID  = "mesh-network-test"
)

type P2PClaims struct {
	Subject        string   `json:"sub"`
	TenantID       string   `json:"tenant_id"`
	ServerID       string   `json:"server_id"`
	ConnectionType string   `json:"connection_type"`
	ProtocolType   string   `json:"protocol_type"`
	Permissions    []string `json:"permissions"`
	jwt.RegisteredClaims
}

func generateToken(serverID string) (string, error) {
	claims := P2PClaims{
		Subject:        serverID,
		TenantID:       tenantID,
		ServerID:       serverID,
		ConnectionType: "quic",
		ProtocolType:   "p2p-mesh",
		Permissions:    []string{"relay_connect", "relay_tunnel", "p2p_discover", "p2p_connect"},
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Audience:  jwt.ClaimStrings{audience},
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			// ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // Для постоянного токена, закомментировать эту строку
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return signedToken, nil
}

func main() {
	fmt.Println("=== CloudBridge P2P Tokens (Updated Secret) ===")

	// Generate token for Server A
	tokenA, err := generateToken("server-a")
	if err != nil {
		fmt.Printf("Error generating token for Server A: %v\n", err)
		return
	}
	fmt.Printf("\nServer A Token:\n%s\n", tokenA)

	// Generate token for Server B
	tokenB, err := generateToken("server-b")
	if err != nil {
		fmt.Printf("Error generating token for Server B: %v\n", err)
		return
	}
	fmt.Printf("\nServer B Token:\n%s\n", tokenB)

	fmt.Println("\n=== Configuration Files ===")
	fmt.Printf("\nserver-a-pilot.yaml:\nauth:\n  type: \"jwt\"\n  token: \"%s\"\n  secret: \"%s\"\n", tokenA, jwtSecret)
	fmt.Printf("\nserver-b-pilot.yaml:\nauth:\n  type: \"jwt\"\n  token: \"%s\"\n  secret: \"%s\"\n", tokenB, jwtSecret)

	fmt.Println("\n=== Token Details ===")
	fmt.Printf("Audience: %s\n", audience)
	fmt.Println("Expires: Never (no expiration)")
	fmt.Println("Algorithm: HS256")
	fmt.Println("Purpose: P2P mesh connectivity")
	fmt.Println("Secret: Updated to match server")
}
