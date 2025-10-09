package main

import (
	"fmt"
	"time"
	"github.com/golang-jwt/jwt/v5"
)

type P2PClaims struct {
	jwt.RegisteredClaims
	ProtocolType   string   `json:"protocol_type"`
	Scope          string   `json:"scope"`
	TenantID       string   `json:"tenant_id"`
	ServerID       string   `json:"server_id"`
	ConnectionType string   `json:"connection_type"`
	Permissions    []string `json:"permissions"`
}

func main() {
	// Используем тот же секрет, что и в конфигурации
	secret := "85Sk/NfLq3gqzCXzmBKbJCpL+f5BssXz3G8dVi3sPiE="
	
	// Создаем claims с правильными permissions
	claims := &P2PClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "https://edge.2gc.ru/oauth/v2",
			Subject:   "server-server-a-12345",
			Audience:  []string{"cloudbridge-relay"},
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			ID:        "jwt_1759875971954_10hg68x0s",
		},
		ProtocolType:   "p2p-mesh",
		Scope:          "relay:connect relay:manage tenant:read tenant:write",
		TenantID:       "test-tenant-dev",
		ServerID:       "server-a-12345",
		ConnectionType: "quic",
		Permissions:    []string{"p2p_connect", "relay:manage"}, // Исправляем permissions
	}
	
	// Создаем токен с HMAC подписью
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		panic(err)
	}
	
	fmt.Println(tokenString)
}
