package sentry

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// OAuthConfig конфигурация OAuth для Sentry
type OAuthConfig struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	AuthURL      string `json:"auth_url"`
	TokenURL     string `json:"token_url"`
	Scopes       string `json:"scopes"`
	RedirectURL  string `json:"redirect_url"`
}

// OAuthToken токен OAuth
type OAuthToken struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in"`
	RefreshToken string    `json:"refresh_token"`
	Scope        string    `json:"scope"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// OAuthClient клиент для OAuth операций
type OAuthClient struct {
	config *OAuthConfig
	client *http.Client
}

// NewOAuthClient создает новый OAuth клиент
func NewOAuthClient(config *OAuthConfig) *OAuthClient {
	return &OAuthClient{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetAuthURL возвращает URL для авторизации
func (c *OAuthClient) GetAuthURL(state string) string {
	params := url.Values{}
	params.Set("client_id", c.config.ClientID)
	params.Set("redirect_uri", c.config.RedirectURL)
	params.Set("response_type", "code")
	params.Set("scope", c.config.Scopes)
	params.Set("state", state)

	return fmt.Sprintf("%s?%s", c.config.AuthURL, params.Encode())
}

// ExchangeCodeForToken обменивает код авторизации на токен
func (c *OAuthClient) ExchangeCodeForToken(ctx context.Context, code string) (*OAuthToken, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", c.config.ClientID)
	data.Set("client_secret", c.config.ClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", c.config.RedirectURL)

	req, err := http.NewRequestWithContext(ctx, "POST", c.config.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed with status: %d", resp.StatusCode)
	}

	var token OAuthToken
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	// Устанавливаем время истечения токена
	token.ExpiresAt = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)

	return &token, nil
}

// RefreshToken обновляет токен
func (c *OAuthClient) RefreshToken(ctx context.Context, refreshToken string) (*OAuthToken, error) {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("client_id", c.config.ClientID)
	data.Set("client_secret", c.config.ClientSecret)
	data.Set("refresh_token", refreshToken)

	req, err := http.NewRequestWithContext(ctx, "POST", c.config.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token refresh failed with status: %d", resp.StatusCode)
	}

	var token OAuthToken
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	// Устанавливаем время истечения токена
	token.ExpiresAt = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)

	return &token, nil
}

// IsTokenValid проверяет, действителен ли токен
func (t *OAuthToken) IsTokenValid() bool {
	return time.Now().Before(t.ExpiresAt)
}

// GetAuthHeader возвращает заголовок авторизации
func (t *OAuthToken) GetAuthHeader() string {
	return fmt.Sprintf("Bearer %s", t.AccessToken)
}

// DefaultOAuthConfig возвращает конфигурацию OAuth по умолчанию
func DefaultOAuthConfig() *OAuthConfig {
	return &OAuthConfig{
		ClientID:     getEnvOrDefault("SENTRY_CLIENT_ID", ""),
		ClientSecret: getEnvOrDefault("SENTRY_CLIENT_SECRET", ""),
		AuthURL:      getEnvOrDefault("SENTRY_AUTH_URL", "https://sentry.io/oauth/authorize/"),
		TokenURL:     getEnvOrDefault("SENTRY_TOKEN_URL", "https://sentry.io/oauth/token/"),
		Scopes:       getEnvOrDefault("SENTRY_OAUTH_SCOPES", "project:read,project:write,event:read,event:write,org:read"),
		RedirectURL:  getEnvOrDefault("SENTRY_REDIRECT_URL", "http://localhost:8080/auth/sentry/callback"),
	}
}



