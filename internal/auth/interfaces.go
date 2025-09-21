package auth

// KeycloakAuthInterface defines the interface for Keycloak authentication operations
type KeycloakAuthInterface interface {
	AuthenticateUser(token string) (bool, error)
	ValidateToken(token string) (bool, error)
}
