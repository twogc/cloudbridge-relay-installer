package zitadel

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
)

// Config конфигурация Zitadel клиента
type Config struct {
	Issuer       string        `json:"issuer"`    // e.g. https://zitadel.example.com
	Audience     string        `json:"audience"`  // ожидаемая aud (project/client ID)
	TokenURL     string        `json:"token_url"` // обычно issuer + /oauth/v2/token
	ClientID     string        `json:"client_id"`
	ClientSecret string        `json:"client_secret"`
	GRPCHost     string        `json:"grpc_host"` // e.g. zitadel.example.com:443
	Timeout      time.Duration `json:"timeout"`
	HTTPTimeout  time.Duration `json:"http_timeout"`
	Insecure     bool          `json:"insecure"` // dev only

	// Optional (только если реально используете интерактивный авторизационный код):
	RedirectURL string   `json:"redirect_url,omitempty"`
	Scopes      []string `json:"scopes,omitempty"`
}

// Client представляет Zitadel клиент для полной интеграции
type Client struct {
	logger           *zap.Logger
	config           *Config
	httpClient       *http.Client
	oidcProvider     *oidc.Provider
	oidcVerifier     *oidc.IDTokenVerifier
	grpcConn         *grpc.ClientConn
	managementClient interface{} // TODO: заменить на management.ManagementServiceClient
	orgClient        interface{} // TODO: заменить на org.OrganizationServiceClient
	tenantManager    *TenantManager
}

// TokenClaims - ЕДИНСТВЕННЫЙ канонический тип claims
type TokenClaims struct {
	// Стандартные JWT claims
	Subject   string   `json:"sub"`
	Issuer    string   `json:"iss"`
	Audience  []string `json:"aud"`
	Expires   int64    `json:"exp"`
	NotBefore int64    `json:"nbf"`
	IssuedAt  int64    `json:"iat"`

	// Zitadel-specific claims
	OrgID             string                 `json:"org_id,omitempty"` // фиксированный claim для TenantID
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

// NewClient создает новый Zitadel клиент с полной интеграцией
func NewClient(ctx context.Context, cfg *Config, logger *zap.Logger) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}
	if cfg.Issuer == "" {
		return nil, fmt.Errorf("issuer is required")
	}
	if cfg.GRPCHost == "" {
		return nil, fmt.Errorf("grpc_host is required")
	}
	if cfg.TokenURL == "" {
		return nil, fmt.Errorf("token_url is required")
	}
	if cfg.ClientID == "" || cfg.ClientSecret == "" {
		return nil, fmt.Errorf("client credentials required")
	}
	if cfg.Audience == "" {
		return nil, fmt.Errorf("audience is required")
	}

	// Create HTTP client with timeout
	httpClient := &http.Client{Timeout: cfg.Timeout}
	if httpClient.Timeout == 0 {
		httpClient.Timeout = 30 * time.Second
	}

	// OIDC provider/verifier
	provider, err := oidc.NewProvider(ctx, cfg.Issuer)
	if err != nil {
		return nil, fmt.Errorf("create OIDC provider: %w", err)
	}
	verifier := provider.Verifier(&oidc.Config{
		ClientID:          cfg.Audience,
		SkipIssuerCheck:   false,
		SkipClientIDCheck: false,
	})

	// gRPC с per-RPC OAuth2 Client Credentials
	conn, err := dialZitadel(ctx, cfg, logger)
	if err != nil {
		return nil, err
	}

	c := &Client{
		logger:       logger,
		config:       cfg,
		httpClient:   httpClient,
		oidcProvider: provider,
		oidcVerifier: verifier,
		grpcConn:     conn,
		// TODO: заменить на реальные клиенты когда zitadel-go будет доступен
		managementClient: nil,
		orgClient:        nil,
	}
	c.tenantManager = NewTenantManager(c, logger)

	logger.Info("Zitadel client initialized",
		zap.String("issuer", cfg.Issuer),
		zap.String("audience", cfg.Audience),
		zap.String("grpc_host", cfg.GRPCHost),
	)
	return c, nil
}

// dialZitadel создает gRPC соединение с OAuth2 аутентификацией
func dialZitadel(ctx context.Context, cfg *Config, logger *zap.Logger) (*grpc.ClientConn, error) {
	cc := clientcredentials.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		TokenURL:     cfg.TokenURL,
		Scopes:       []string{"openid", "profile", "email"},
		AuthStyle:    oauth2.AuthStyleInParams,
		EndpointParams: url.Values{
			"audience": []string{cfg.Audience},
		},
	}
	perRPC := oauth.TokenSource{
		TokenSource: oauth2.ReuseTokenSource(nil, cc.TokenSource(ctx)),
	}

	tlsCfg := &tls.Config{MinVersion: tls.VersionTLS12}
	if cfg.Insecure {
		tlsCfg.InsecureSkipVerify = true // dev only
	}
	creds := credentials.NewTLS(tlsCfg)

	return grpc.DialContext(
		ctx,
		cfg.GRPCHost,
		grpc.WithTransportCredentials(creds),
		grpc.WithPerRPCCredentials(perRPC),
		grpc.WithDefaultCallOptions(grpc.WaitForReady(true)),
	)
}

// ValidateToken - ЕДИНСТВЕННЫЙ метод валидации через go-oidc
func (c *Client) ValidateToken(ctx context.Context, raw string) (*TokenClaims, error) {
	raw = strings.TrimPrefix(raw, "Bearer ")
	idToken, err := c.oidcVerifier.Verify(ctx, raw)
	if err != nil {
		return nil, fmt.Errorf("oidc verify: %w", err)
	}

	var claims TokenClaims
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("decode claims: %w", err)
	}

	if claims.OrgID == "" {
		// альтернативно: попробуйте извлечь из namespaced claim Zitadel, если вы так настроили
		return nil, fmt.Errorf("missing org_id claim")
	}

	c.logger.Debug("token validated",
		zap.String("sub", claims.Subject),
		zap.String("org_id", claims.OrgID))

	return &claims, nil
}

// GetUserInfo получает информацию о пользователе из токена
func (c *Client) GetUserInfo(ctx context.Context, token string) (*UserInfo, error) {
	claims, err := c.ValidateToken(ctx, token)
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
		Provider: "zitadel",
		Claims:   map[string]interface{}{}, // при желании наполните из idToken (raw)
	}

	return userInfo, nil
}

// GetTenantManager возвращает менеджер тенантов
func (c *Client) GetTenantManager() *TenantManager {
	return c.tenantManager
}

// Close - ЕДИНСТВЕННЫЙ метод закрытия
func (c *Client) Close() error {
	if c.grpcConn != nil {
		return c.grpcConn.Close()
	}
	return nil
}
