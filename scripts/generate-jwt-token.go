package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run generate-jwt-token.go <server-id> <tenant-id> [duration]")
		fmt.Println("Example: go run generate-jwt-token.go server-a mesh-network-test 168h")
		os.Exit(1)
	}

	serverID := os.Args[1]
	tenantID := os.Args[2]
	duration := "168h" // default 7 days
	if len(os.Args) > 3 {
		duration = os.Args[3]
	}

	// Parse duration
	d, err := time.ParseDuration(duration)
	if err != nil {
		fmt.Printf("Error parsing duration: %v\n", err)
		os.Exit(1)
	}

	// JWT secret from Minikube
	secret := "85Sk/NfLq3gqzCXzmBKbJCpL+f5BssXz3G8dVi3sPiE="
	secretBytes, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		fmt.Printf("Error decoding secret: %v\n", err)
		os.Exit(1)
	}

	// Create claims
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":              serverID,
		"iss":              "cloudbridge-relay",
		"aud":              "cloudbridge-relay",
		"iat":              now.Unix(),
		"exp":              now.Add(d).Unix(),
		"tenant_id":        tenantID,
		"server_id":        serverID,
		"protocol":         "p2p-mesh",
		"connection_type": "quic",
		"permissions":      []string{"relay_connect", "relay_tunnel"},
		"scopes":           []string{"relay_p2p"},
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secretBytes)
	if err != nil {
		fmt.Printf("Error signing token: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated JWT token for %s (tenant: %s):\n", serverID, tenantID)
	fmt.Printf("Token: %s\n", tokenString)
	fmt.Printf("Expires: %s\n", now.Add(d).Format(time.RFC3339))
}

