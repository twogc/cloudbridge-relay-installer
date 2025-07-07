package django

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client represents a Django API client
type Client struct {
	client      *http.Client
	apiEndpoint string
	jwtSecret   string
}

// Config represents the Django client configuration
type Config struct {
	APIEndpoint string
	JWTSecret   string
	Timeout     time.Duration
}

// NewClient creates a new Django API client
func NewClient(config Config) *Client {
	return &Client{
		client: &http.Client{
			Timeout: config.Timeout,
		},
		apiEndpoint: config.APIEndpoint,
		jwtSecret:   config.JWTSecret,
	}
}

// ValidateTokenRequest represents a token validation request
type ValidateTokenRequest struct {
	Token string `json:"token"`
}

// ValidateTokenResponse represents a token validation response
type ValidateTokenResponse struct {
	Valid    bool   `json:"valid"`
	ClientID string `json:"client_id,omitempty"`
	Error    string `json:"error,omitempty"`
}

// ValidateToken validates a token using the Django API
func (c *Client) ValidateToken(ctx context.Context, token string) (string, error) {
	reqBody := ValidateTokenRequest{
		Token: token,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.apiEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "2GCClient/test-client")
	req.Host = "lk-dasjboard-django-lk-app-1"

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response ValidateTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	if response.Error != "" {
		return "", fmt.Errorf("validation error: %s", response.Error)
	}

	if !response.Valid {
		return "", fmt.Errorf("invalid token")
	}

	if response.ClientID == "" {
		return "", fmt.Errorf("missing client_id in response")
	}

	return response.ClientID, nil
} 