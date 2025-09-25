package zitadel

import (
	"context"
	"crypto/x509"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/oauth2"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Client представляет Zitadel клиент для полной интеграции
type Client struct {
	logger     *zap.Logger
	config     *Config
	httpClient *http.Client

	// OAuth2 and JWKS components
	oauth2Config *oauth2.Config
	jwksProvider *JWKSProvider

	// gRPC connection for management API
	grpcConn *grpc.ClientConn

	// Tenant management
	tenantManager *TenantManager
}

// Config конфигурация Zitadel клиента
type Config struct {
	Domain         string        `json:"domain"`
	ProjectID      string        `json:"project_id"`
	ClientID       string        `json:"client_id"`
	ClientSecret   string        `json:"client_secret"`
	FallbackSecret string        `json:"fallback_secret"`
	Scopes         []string      `json:"scopes"`
	RedirectURL    string        `json:"redirect_url"`
	HTTPTimeout    time.Duration `json:"http_timeout"`
	Insecure       bool          `json:"insecure"`
}

// NewClient создает новый Zitadel клиент с полной интеграцией
func NewClient(ctx context.Context, config *Config, logger *zap.Logger) (*Client, error) {
	if config.Domain == "" {
		return nil, fmt.Errorf("zitadel domain is required")
	}

	// Create HTTP client
	httpClient := &http.Client{
		Timeout: config.HTTPTimeout,
	}
	if config.HTTPTimeout == 0 {
		httpClient.Timeout = 10 * time.Second
	}

	client := &Client{
		logger:     logger,
		config:     config,
		httpClient: httpClient,
	}

	// Initialize OAuth2 configuration
	if err := client.initOAuth2Config(); err != nil {
		return nil, fmt.Errorf("failed to initialize OAuth2 config: %w", err)
	}

	// Initialize JWKS provider
	if err := client.initJWKSProvider(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize JWKS provider: %w", err)
	}

	// Initialize gRPC connection for management API
	if err := client.initGRPCConnection(ctx); err != nil {
		logger.Warn("Failed to initialize gRPC connection, management API will not be available", zap.Error(err))
	}

	// Initialize tenant manager
	client.tenantManager = NewTenantManager(client, logger)

	logger.Info("Zitadel client initialized successfully",
		zap.String("domain", config.Domain),
		zap.String("project_id", config.ProjectID),
		zap.String("client_id", config.ClientID))

	return client, nil
}

// initOAuth2Config инициализирует OAuth2 конфигурацию
func (c *Client) initOAuth2Config() error {
	issuerURL := fmt.Sprintf("https://%s", c.config.Domain)

	// Configure OAuth2 endpoints manually
	scopes := c.config.Scopes
	if len(scopes) == 0 {
		scopes = []string{"openid", "profile", "email"}
	}

	c.oauth2Config = &oauth2.Config{
		ClientID:     c.config.ClientID,
		ClientSecret: c.config.ClientSecret,
		RedirectURL:  c.config.RedirectURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("%s/oauth/v2/authorize", issuerURL),
			TokenURL: fmt.Sprintf("%s/oauth/v2/token", issuerURL),
		},
		Scopes: scopes,
	}

	c.logger.Debug("OAuth2 config initialized",
		zap.String("issuer", issuerURL),
		zap.Strings("scopes", scopes))

	return nil
}

// initJWKSProvider инициализирует JWKS провайдер
func (c *Client) initJWKSProvider(ctx context.Context) error {
	jwksProvider, err := NewJWKSProvider(c.config.Domain, c.httpClient, c.logger)
	if err != nil {
		return fmt.Errorf("failed to create JWKS provider: %w", err)
	}

	c.jwksProvider = jwksProvider

	// Prefetch keys
	if err := c.jwksProvider.RefreshKeys(ctx); err != nil {
		c.logger.Warn("Failed to prefetch JWKS keys", zap.Error(err))
	}

	return nil
}

// initGRPCConnection инициализирует gRPC соединение для management API
func (c *Client) initGRPCConnection(ctx context.Context) error {
	apiURL := fmt.Sprintf("%s:443", c.config.Domain)

	var opts []grpc.DialOption

	if c.config.Insecure {
		opts = append(opts, grpc.WithInsecure())
	} else {
		creds, err := c.createTLSCredentials()
		if err != nil {
			return fmt.Errorf("failed to create TLS credentials: %w", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	}

	// Add authentication interceptor
	authInterceptor, err := c.createAuthInterceptor(ctx)
	if err != nil {
		return fmt.Errorf("failed to create auth interceptor: %w", err)
	}

	opts = append(opts,
		grpc.WithUnaryInterceptor(authInterceptor),
		grpc.WithDefaultCallOptions(grpc.WaitForReady(true)))

	conn, err := grpc.DialContext(ctx, apiURL, opts...)
	if err != nil {
		return fmt.Errorf("failed to connect to Zitadel API: %w", err)
	}

	c.grpcConn = conn
	c.logger.Debug("gRPC connection established", zap.String("api_url", apiURL))

	return nil
}

// createTLSCredentials создает TLS credentials
func (c *Client) createTLSCredentials() (credentials.TransportCredentials, error) {
	certPool, err := x509.SystemCertPool()
	if err != nil {
		return nil, err
	}
	if certPool == nil {
		certPool = x509.NewCertPool()
	}

	serverName := strings.Split(c.config.Domain, ":")[0]
	return credentials.NewClientTLSFromCert(certPool, serverName), nil
}

// createAuthInterceptor создает аутентификационный interceptor
func (c *Client) createAuthInterceptor(ctx context.Context) (grpc.UnaryClientInterceptor, error) {
	// For service account authentication, we would use JWT profile
	// For now, implement basic token-based auth
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// Add authorization header if available
		// In a full implementation, this would handle JWT profile or service account tokens
		return invoker(ctx, method, req, reply, cc, opts...)
	}, nil
}

// ValidateToken валидирует JWT токен с использованием JWKS
func (c *Client) ValidateToken(ctx context.Context, tokenString string) (*TokenClaims, error) {
	// Remove Bearer prefix if present
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	// Parse token without verification first to get kid
	token, err := jwt.Parse(tokenString, nil)
	if token == nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// Get key ID from token header
	var keyID string
	if kid, ok := token.Header["kid"].(string); ok {
		keyID = kid
	}

	// Get public key from JWKS
	publicKey, err := c.jwksProvider.GetPublicKey(ctx, keyID)
	if err != nil {
		c.logger.Warn("Failed to get public key from JWKS, falling back to validation",
			zap.String("kid", keyID), zap.Error(err))

		// Fallback to basic validation
		return c.validateTokenFallback(tokenString)
	}

	// Parse and validate token with public key
	claims := &TokenClaims{}
	token, err = jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is invalid")
	}

	// Verify issuer
	expectedIssuer := fmt.Sprintf("https://%s", c.config.Domain)
	if claims.Issuer != expectedIssuer {
		return nil, fmt.Errorf("invalid issuer: expected %s, got %s", expectedIssuer, claims.Issuer)
	}

	// Verify audience manually
	validAudience := false
	for _, aud := range claims.Audience {
		if aud == c.config.ClientID {
			validAudience = true
			break
		}
	}
	if !validAudience {
		return nil, fmt.Errorf("invalid audience: expected %s", c.config.ClientID)
	}

	c.logger.Debug("Token validated successfully",
		zap.String("subject", claims.Subject),
		zap.String("issuer", claims.Issuer))

	return claims, nil
}

// validateTokenFallback fallback валидация токена
func (c *Client) validateTokenFallback(tokenString string) (*TokenClaims, error) {
	if c.config.FallbackSecret == "" {
		return nil, fmt.Errorf("no fallback secret configured")
	}

	claims := &TokenClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(c.config.FallbackSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("fallback token validation failed: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("fallback token is invalid")
	}

	c.logger.Debug("Token validated using fallback method", zap.String("subject", claims.Subject))
	return claims, nil
}

// GetUserInfo получает информацию о пользователе
func (c *Client) GetUserInfo(ctx context.Context, tokenString string) (*UserInfo, error) {
	claims, err := c.ValidateToken(ctx, tokenString)
	if err != nil {
		return nil, err
	}

	userInfo := &UserInfo{
		ID:       claims.Subject,
		Username: claims.PreferredUsername,
		Email:    claims.Email,
		Name:     claims.Name,
		Roles:    claims.Roles,
		Groups:   claims.Groups,
		Active:   true,
		Claims:   claims.Extra,
		Provider: "zitadel",
	}

	// If we have gRPC connection, we could fetch additional user details
	if c.grpcConn != nil {
		if err := c.enrichUserInfo(ctx, userInfo); err != nil {
			c.logger.Warn("Failed to enrich user info", zap.Error(err))
		}
	}

	return userInfo, nil
}

// enrichUserInfo обогащает информацию о пользователе через Management API
func (c *Client) enrichUserInfo(ctx context.Context, userInfo *UserInfo) error {
	// In a full implementation, this would call Zitadel Management API
	// to get additional user details, organization membership, etc.
	c.logger.Debug("User info enrichment not implemented yet", zap.String("user_id", userInfo.ID))
	return nil
}

// CreateAuthURL создает URL для аутентификации
func (c *Client) CreateAuthURL(state string) string {
	return c.oauth2Config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// ExchangeCode обменивает authorization code на токены
func (c *Client) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	token, err := c.oauth2Config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	c.logger.Debug("Successfully exchanged code for token")
	return token, nil
}

// RefreshToken обновляет access token
func (c *Client) RefreshToken(ctx context.Context, refreshToken string) (*oauth2.Token, error) {
	token := &oauth2.Token{
		RefreshToken: refreshToken,
	}

	tokenSource := c.oauth2Config.TokenSource(ctx, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	c.logger.Debug("Successfully refreshed token")
	return newToken, nil
}

// GetTenantManager возвращает tenant manager
func (c *Client) GetTenantManager() *TenantManager {
	return c.tenantManager
}

// GetJWKSProvider возвращает JWKS provider
func (c *Client) GetJWKSProvider() *JWKSProvider {
	return c.jwksProvider
}

// HealthCheck проверяет доступность Zitadel
func (c *Client) HealthCheck(ctx context.Context) error {
	// Check OAuth2 config
	if c.oauth2Config == nil {
		return fmt.Errorf("OAuth2 config not initialized")
	}

	// Check JWKS keys availability
	if err := c.jwksProvider.RefreshKeys(ctx); err != nil {
		return fmt.Errorf("JWKS keys not available: %w", err)
	}

	// Check gRPC connection if available
	if c.grpcConn != nil {
		state := c.grpcConn.GetState()
		if state.String() != "READY" && state.String() != "IDLE" {
			c.logger.Warn("gRPC connection not ready", zap.String("state", state.String()))
		}
	}

	return nil
}

// Close закрывает соединения
func (c *Client) Close() error {
	if c.grpcConn != nil {
		return c.grpcConn.Close()
	}
	return nil
}

// TokenClaims представляет JWT claims
type TokenClaims struct {
	jwt.RegisteredClaims
	Email             string                 `json:"email,omitempty"`
	EmailVerified     bool                   `json:"email_verified,omitempty"`
	Name              string                 `json:"name,omitempty"`
	PreferredUsername string                 `json:"preferred_username,omitempty"`
	Roles             []string               `json:"roles,omitempty"`
	Groups            []string               `json:"groups,omitempty"`
	Extra             map[string]interface{} `json:"-"`
}

// UserInfo представляет информацию о пользователе
type UserInfo struct {
	ID       string                 `json:"id"`
	Username string                 `json:"username"`
	Email    string                 `json:"email"`
	Name     string                 `json:"name"`
	Roles    []string               `json:"roles"`
	Groups   []string               `json:"groups"`
	Active   bool                   `json:"active"`
	Claims   map[string]interface{} `json:"claims"`
	Provider string                 `json:"provider"`
}
