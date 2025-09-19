package main

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func main() {
	secret := "85Sk/NfLq3gqzCXzmBKbJCpL+f5BssXz3G8dVi3sPiE="
	now := time.Now().Unix()

	claims := jwt.MapClaims{
		"protocol_type":   "p2p-mesh",
		"scope":          "p2p-mesh-claims",
		"tenant_id":      "tenant-216420165",
		"server_id":      "server-17582480805566",
		"connection_type": "quic",
		"permissions":    []string{"p2p_connect", "mesh_join", "mesh_manage"},
		"iat":            now,
		"iss":            "https://auth.2gc.ru/realms/cloudbridge",
		"aud":            "account",
		"sub":            "server-17582480805566",
		"jti":            fmt.Sprintf("jwt_%d_test", now),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		panic(err)
	}

	fmt.Println(tokenString)
}
