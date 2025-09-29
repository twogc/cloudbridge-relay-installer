package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/twogc/cloudbridge-relay-installer/internal/ai/analytics"
)

func TestLoadConfigWithNestedRoot(t *testing.T) {
	configContent := `ai_analytics:
  enabled: true
  schedule: "0 * * * *"
  max_reports: 5
  retention_days: 30
  cloudflare_ai:
    account_id: "acc"
    api_token: "token"
    model: "cf_model"
    max_tokens: 1024
  notifications:
    slack:
      enabled: true
      webhook_url: "https://hooks.slack.com/services/foo"
      channel: "#alerts"
    email:
      enabled: true
      smtp_host: "smtp.example.com"
      smtp_port: 587
      username: "user"
      password: "pass"
      recipients:
        - "a@example.com"
        - "b@example.com"
    webhooks:
      - url: "https://example.com/webhook"
        enabled: true
  data_sources:
    prometheus:
      enabled: true
      url: "http://prometheus:9090"
    zabbix:
      enabled: false
      url: ""
    ai_service:
      enabled: true
      url: "http://ai-service"`

	path := writeTempConfigFile(t, "nested", configContent)
	t.Setenv("AI_ANALYTICS_CONFIG_PATH", path)

	got, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}

	expected := &analytics.AnalyticsConfig{
		Enabled:       true,
		Schedule:      "0 * * * *",
		MaxReports:    5,
		RetentionDays: 30,
		CloudflareAI: analytics.CloudflareAIConfig{
			AccountID: "acc",
			APIToken:  "token",
			Model:     "cf_model",
			MaxTokens: 1024,
		},
		Notifications: analytics.NotificationConfig{
			Slack: analytics.SlackConfig{
				Enabled:    true,
				WebhookURL: "https://hooks.slack.com/services/foo",
				Channel:    "#alerts",
			},
			Email: analytics.EmailConfig{
				Enabled:    true,
				SMTPHost:   "smtp.example.com",
				SMTPPort:   587,
				Username:   "user",
				Password:   "pass",
				Recipients: []string{"a@example.com", "b@example.com"},
			},
			Webhooks: []analytics.WebhookConfig{{
				URL:     "https://example.com/webhook",
				Enabled: true,
			}},
		},
		DataSources: analytics.DataSourcesConfig{
			Prometheus: analytics.PrometheusConfig{
				Enabled: true,
				URL:     "http://prometheus:9090",
			},
			Zabbix: analytics.ZabbixConfig{
				Enabled: false,
				URL:     "",
			},
			AIService: analytics.AIServiceConfig{
				Enabled: true,
				URL:     "http://ai-service",
			},
		},
	}

	if got == nil {
		t.Fatal("loadConfig() returned nil config")
	}

	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("loadConfig() = %+v, want %+v", got, expected)
	}
}

func TestLoadConfigWithFlatStructure(t *testing.T) {
	configContent := `enabled: false
schedule: "*/15 * * * *"
max_reports: 10
retention_days: 7
cloudflare_ai:
  account_id: "flat-acc"
  api_token: "flat-token"
  model: "flat-model"
  max_tokens: 2048
notifications:
  slack:
    enabled: false
    webhook_url: ""
    channel: ""
  email:
    enabled: false
    smtp_host: ""
    smtp_port: 0
    username: ""
    password: ""
    recipients: []
  webhooks: []
data_sources:
  prometheus:
    enabled: false
    url: ""
  zabbix:
    enabled: true
    url: "http://zabbix"
  ai_service:
    enabled: false
    url: ""`

	path := writeTempConfigFile(t, "flat", configContent)
	t.Setenv("AI_ANALYTICS_CONFIG_PATH", path)

	got, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}

	expected := &analytics.AnalyticsConfig{
		Enabled:       false,
		Schedule:      "*/15 * * * *",
		MaxReports:    10,
		RetentionDays: 7,
		CloudflareAI: analytics.CloudflareAIConfig{
			AccountID: "flat-acc",
			APIToken:  "flat-token",
			Model:     "flat-model",
			MaxTokens: 2048,
		},
		Notifications: analytics.NotificationConfig{
			Slack: analytics.SlackConfig{
				Enabled:    false,
				WebhookURL: "",
				Channel:    "",
			},
			Email: analytics.EmailConfig{
				Enabled:    false,
				SMTPHost:   "",
				SMTPPort:   0,
				Username:   "",
				Password:   "",
				Recipients: []string{},
			},
			Webhooks: []analytics.WebhookConfig{},
		},
		DataSources: analytics.DataSourcesConfig{
			Prometheus: analytics.PrometheusConfig{
				Enabled: false,
				URL:     "",
			},
			Zabbix: analytics.ZabbixConfig{
				Enabled: true,
				URL:     "http://zabbix",
			},
			AIService: analytics.AIServiceConfig{
				Enabled: false,
				URL:     "",
			},
		},
	}

	if got == nil {
		t.Fatal("loadConfig() returned nil config")
	}

	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("loadConfig() = %+v, want %+v", got, expected)
	}
}

func TestLoadConfigUnreadableFile(t *testing.T) {
	path := writeTempConfigFile(t, "unreadable", "temporary")
	if err := os.Chmod(path, 0o000); err != nil {
		t.Fatalf("failed to chmod temp file: %v", err)
	}
	t.Cleanup(func() {
		os.Chmod(path, 0o600)
	})

	t.Setenv("AI_ANALYTICS_CONFIG_PATH", path)

	cfg, err := loadConfig()
	if err == nil {
		t.Fatal("expected error when loading unreadable config file, got nil")
	}
	if cfg != nil {
		t.Fatalf("expected nil config when loadConfig fails, got %+v", cfg)
	}
}

func writeTempConfigFile(t *testing.T, pattern, content string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, pattern+".yaml")

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write temp config file: %v", err)
	}

	t.Cleanup(func() {
		os.Remove(path)
	})

	return path
}
