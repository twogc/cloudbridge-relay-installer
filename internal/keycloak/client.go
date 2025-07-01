package keycloak

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
)

// Config represents Keycloak configuration
type Config struct {
	ServerURL    string        `json:"server_url"`
	Realm        string        `json:"realm"`
	ClientID     string        `json:"client_id"`
	ClientSecret string        `json:"client_secret"`
	PublicKey    string        `json:"public_key"`
	TokenValidation struct {
		Issuer     string        `json:"issuer"`
		Audience   string        `json:"audience"`
		ClockSkew  time.Duration `json:"clock_skew"`
	} `json:"token_validation"`
	HTTPClient *http.Client
	Logger     *logger.Logger
}

// Client represents Keycloak client
type Client struct {
	config     *Config
	publicKey  *rsa.PublicKey
	httpClient *http.Client
	logger     *logger.Logger
	mu         sync.RWMutex
	// Кэш для JWKS (JSON Web Key Set)
	jwksCache map[string]*rsa.PublicKey
	cacheTTL  time.Duration
	lastFetch time.Time
}

// TokenInfo represents validated token information
type TokenInfo struct {
	Subject    string            `json:"sub"`
	Issuer     string            `json:"iss"`
	Audience   []string          `json:"aud"`
	ExpiresAt  time.Time         `json:"exp"`
	IssuedAt   time.Time         `json:"iat"`
	NotBefore  time.Time         `json:"nbf"`
	Username   string            `json:"preferred_username"`
	Email      string            `json:"email"`
	FullName   string            `json:"name"`
	Groups     []string          `json:"groups"`
	Roles      []string          `json:"roles"`
	RealmRoles map[string]string `json:"realm_access"`
	ClientRoles map[string]string `json:"resource_access"`
	Claims     map[string]interface{} `json:"-"`
}

// JWKS represents JSON Web Key Set
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// JWK represents JSON Web Key
type JWK struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// NewClient creates a new Keycloak client
func NewClient(config *Config) (*Client, error) {
	if config.HTTPClient == nil {
		config.HTTPClient = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	client := &Client{
		config:     config,
		httpClient: config.HTTPClient,
		logger:     config.Logger,
		jwksCache:  make(map[string]*rsa.PublicKey),
		cacheTTL:   1 * time.Hour,
	}

	// Parse public key if provided
	if config.PublicKey != "" {
		publicKey, err := parseRSAPublicKey(config.PublicKey)
		if err != nil {
			return nil, fmt.Errorf("failed to parse public key: %v", err)
		}
		client.publicKey = publicKey
	}

	return client, nil
}

// ValidateToken validates JWT token from Keycloak
func (c *Client) ValidateToken(ctx context.Context, tokenString string) (*TokenInfo, error) {
	// Parse token without validation first
	token, err := jwt.Parse(tokenString, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %v", err)
	}

	// Get key ID from token header
	kid, ok := token.Header["kid"].(string)
	if !ok {
		return nil, fmt.Errorf("missing kid in token header")
	}

	// Get public key for this key ID
	publicKey, err := c.getPublicKey(ctx, kid)
	if err != nil {
		return nil, fmt.Errorf("failed to get public key: %v", err)
	}

	// Parse and validate token
	parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate algorithm
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to validate token: %v", err)
	}

	if !parsedToken.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Extract claims
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Validate claims
	if err := c.validateClaims(claims); err != nil {
		return nil, fmt.Errorf("token validation failed: %v", err)
	}

	// Convert to TokenInfo
	tokenInfo, err := c.claimsToTokenInfo(claims)
	if err != nil {
		return nil, fmt.Errorf("failed to convert claims: %v", err)
	}

	return tokenInfo, nil
}

// getPublicKey retrieves public key from cache or fetches from Keycloak
func (c *Client) getPublicKey(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	c.mu.RLock()
	if key, exists := c.jwksCache[kid]; exists && time.Since(c.lastFetch) < c.cacheTTL {
		c.mu.RUnlock()
		return key, nil
	}
	c.mu.RUnlock()

	// Fetch JWKS from Keycloak
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring write lock
	if key, exists := c.jwksCache[kid]; exists && time.Since(c.lastFetch) < c.cacheTTL {
		return key, nil
	}

	// Fetch JWKS
	jwks, err := c.fetchJWKS(ctx)
	if err != nil {
		return nil, err
	}

	// Parse and cache keys
	for _, jwk := range jwks.Keys {
		if jwk.Kid == kid {
			publicKey, err := jwkToRSAPublicKey(jwk)
			if err != nil {
				return nil, err
			}
			c.jwksCache[kid] = publicKey
			c.lastFetch = time.Now()
			return publicKey, nil
		}
	}

	return nil, fmt.Errorf("key with kid %s not found", kid)
}

// fetchJWKS fetches JSON Web Key Set from Keycloak
func (c *Client) fetchJWKS(ctx context.Context) (*JWKS, error) {
	url := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/certs", 
		c.config.ServerURL, c.config.Realm)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch JWKS: %s", resp.Status)
	}

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, err
	}

	return &jwks, nil
}

// validateClaims validates token claims
func (c *Client) validateClaims(claims jwt.MapClaims) error {
	// Validate issuer
	if issuer, ok := claims["iss"].(string); ok {
		if issuer != c.config.TokenValidation.Issuer {
			return fmt.Errorf("invalid issuer: expected %s, got %s", 
				c.config.TokenValidation.Issuer, issuer)
		}
	}

	// Validate audience
	if audience, ok := claims["aud"].(interface{}); ok {
		switch aud := audience.(type) {
		case string:
			if aud != c.config.TokenValidation.Audience {
				return fmt.Errorf("invalid audience: expected %s, got %s", 
					c.config.TokenValidation.Audience, aud)
			}
		case []interface{}:
			found := false
			for _, a := range aud {
				if a.(string) == c.config.TokenValidation.Audience {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("invalid audience: %s not found in %v", 
					c.config.TokenValidation.Audience, aud)
			}
		}
	}

	// Validate expiration with clock skew
	if exp, ok := claims["exp"].(float64); ok {
		expTime := time.Unix(int64(exp), 0)
		if time.Now().Add(c.config.TokenValidation.ClockSkew).After(expTime) {
			return fmt.Errorf("token expired")
		}
	}

	// Validate not before
	if nbf, ok := claims["nbf"].(float64); ok {
		nbfTime := time.Unix(int64(nbf), 0)
		if time.Now().Add(-c.config.TokenValidation.ClockSkew).Before(nbfTime) {
			return fmt.Errorf("token not yet valid")
		}
	}

	return nil
}

// claimsToTokenInfo converts JWT claims to TokenInfo
func (c *Client) claimsToTokenInfo(claims jwt.MapClaims) (*TokenInfo, error) {
	tokenInfo := &TokenInfo{
		Claims: make(map[string]interface{}),
	}

	// Copy all claims
	for key, value := range claims {
		tokenInfo.Claims[key] = value
	}

	// Extract standard claims
	if sub, ok := claims["sub"].(string); ok {
		tokenInfo.Subject = sub
	}
	if iss, ok := claims["iss"].(string); ok {
		tokenInfo.Issuer = iss
	}
	if exp, ok := claims["exp"].(float64); ok {
		tokenInfo.ExpiresAt = time.Unix(int64(exp), 0)
	}
	if iat, ok := claims["iat"].(float64); ok {
		tokenInfo.IssuedAt = time.Unix(int64(iat), 0)
	}
	if nbf, ok := claims["nbf"].(float64); ok {
		tokenInfo.NotBefore = time.Unix(int64(nbf), 0)
	}

	// Extract custom claims
	if username, ok := claims["preferred_username"].(string); ok {
		tokenInfo.Username = username
	}
	if email, ok := claims["email"].(string); ok {
		tokenInfo.Email = email
	}
	if name, ok := claims["name"].(string); ok {
		tokenInfo.FullName = name
	}

	// Extract groups
	if groups, ok := claims["groups"].([]interface{}); ok {
		for _, group := range groups {
			if groupStr, ok := group.(string); ok {
				tokenInfo.Groups = append(tokenInfo.Groups, groupStr)
			}
		}
	}

	// Extract roles
	if realmAccess, ok := claims["realm_access"].(map[string]interface{}); ok {
		if roles, ok := realmAccess["roles"].([]interface{}); ok {
			for _, role := range roles {
				if roleStr, ok := role.(string); ok {
					tokenInfo.Roles = append(tokenInfo.Roles, roleStr)
				}
			}
		}
	}

	return tokenInfo, nil
}

// CheckPermission checks if user has specific permission
func (c *Client) CheckPermission(tokenInfo *TokenInfo, resource, action string) bool {
	// Check realm roles
	for _, role := range tokenInfo.Roles {
		if role == "admin" || role == "relay-admin" {
			return true
		}
		if role == fmt.Sprintf("%s:%s", resource, action) {
			return true
		}
	}

	// Check groups
	for _, group := range tokenInfo.Groups {
		if group == "relay-users" || group == "relay-admins" {
			return true
		}
	}

	return false
}

// Helper functions
func parseRSAPublicKey(publicKeyPEM string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, fmt.Errorf("failed to decode PEM block containing public key")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not RSA public key")
	}
	return rsaPub, nil
}

func jwkToRSAPublicKey(jwk JWK) (*rsa.PublicKey, error) {
	// Decode base64url-encoded N and E
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode N: %v", err)
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode E: %v", err)
	}
	// Convert eBytes to int
	e := 0
	for _, b := range eBytes {
		e = e<<8 + int(b)
	}
	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: e,
	}, nil
} 