package zitadel

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestNewClient(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping OIDC discovery against live Zitadel in -short mode")
	}

	logger := zap.NewNop()

	config := &Config{
		Issuer:       "https://test.zitadel.cloud",
		Audience:     "test-project",
		TokenURL:     "https://test.zitadel.cloud/oauth/v2/token",
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		GRPCHost:     "test.zitadel.cloud:443",
		Timeout:      30 * time.Second,
		Insecure:     true,
	}

	client, err := NewClient(context.Background(), config, logger)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	if client == nil {
		t.Fatal("Client is nil")
	}

	if client.config != config {
		t.Error("Config not set correctly")
	}

	if client.logger != logger {
		t.Error("Logger not set correctly")
	}

	client.Close()
}

func TestNewClientValidation(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		live    bool // requires network / OIDC discovery
	}{
		{
			name: "missing issuer",
			config: &Config{
				Audience: "test-project",
				ClientID: "test-client",
			},
			wantErr: true,
		},
		{
			name: "missing grpc host",
			config: &Config{
				Issuer:   "https://test.zitadel.cloud",
				Audience: "test-project",
				ClientID: "test-client",
			},
			wantErr: true,
		},
		{
			name: "valid config",
			config: &Config{
				Issuer:       "https://test.zitadel.cloud",
				Audience:     "test-project",
				TokenURL:     "https://test.zitadel.cloud/oauth/v2/token",
				ClientID:     "test-client",
				ClientSecret: "test-secret",
				GRPCHost:     "test.zitadel.cloud:443",
				Insecure:     true,
			},
			wantErr: false,
			live:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.live && testing.Short() {
				t.Skip("skipping live Zitadel OIDC discovery in -short mode")
			}
			client, err := NewClient(context.Background(), tt.config, logger)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if client != nil {
				client.Close()
			}
		})
	}
}

func TestTokenClaims(t *testing.T) {
	claims := &TokenClaims{
		Subject:   "user123",
		Issuer:    "https://test.zitadel.cloud",
		Audience:  []string{"test-project"},
		Expires:   1234567890,
		NotBefore: 1234567800,
		OrgID:     "org123",
	}

	if claims.Subject != "user123" {
		t.Error("Subject not set correctly")
	}

	if claims.OrgID != "org123" {
		t.Error("OrgID not set correctly")
	}

	if len(claims.Audience) != 1 || claims.Audience[0] != "test-project" {
		t.Error("Audience not set correctly")
	}
}
