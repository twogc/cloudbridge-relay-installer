package zitadel

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// K8sConfig конфигурация для локального k8s окружения
type K8sConfig struct {
	// Kubernetes service names
	ZitadelServiceName string `json:"zitadel_service_name"`
	ZitadelNamespace   string `json:"zitadel_namespace"`
	ZitadelPort        int    `json:"zitadel_port"`

	// Internal communication
	UseInternalDNS bool   `json:"use_internal_dns"`
	InternalDomain string `json:"internal_domain"`

	// TLS settings for k8s
	InsecureSkipVerify bool `json:"insecure_skip_verify"`
	UseTLS             bool `json:"use_tls"`

	// Service discovery
	EnableServiceDiscovery bool `json:"enable_service_discovery"`

	// Health check settings
	HealthCheckPath     string        `json:"health_check_path"`
	HealthCheckInterval time.Duration `json:"health_check_interval"`

	*Config // Embed base config
}

// NewK8sClient создает Zitadel клиент для k8s окружения
func NewK8sClient(ctx context.Context, config *K8sConfig, logger *zap.Logger) (*Client, error) {
	// Resolve Zitadel service address in k8s
	zitadelURL, err := resolveZitadelService(config)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve Zitadel service: %w", err)
	}

	logger.Info("Resolved Zitadel service in k8s",
		zap.String("service_url", zitadelURL),
		zap.String("service_name", config.ZitadelServiceName),
		zap.String("namespace", config.ZitadelNamespace))

	// Update base config with resolved URL
	config.Config.Domain = zitadelURL

	// Create HTTP client with k8s-specific settings
	httpClient := createK8sHTTPClient(config)

	// Create base client with k8s-specific config
	baseConfig := config.Config
	baseConfig.HTTPTimeout = 15 * time.Second // Shorter timeout for local k8s

	client, err := NewClient(ctx, baseConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create base Zitadel client: %w", err)
	}

	// Override HTTP client with k8s-specific one
	client.httpClient = httpClient

	// Start health monitoring if enabled
	if config.EnableServiceDiscovery {
		go client.monitorZitadelHealth(ctx, config, logger)
	}

	return client, nil
}

// resolveZitadelService разрешает адрес сервиса Zitadel в k8s
func resolveZitadelService(config *K8sConfig) (string, error) {
	if config.UseInternalDNS {
		// Use Kubernetes internal DNS
		serviceDNS := fmt.Sprintf("%s.%s.svc.cluster.local",
			config.ZitadelServiceName,
			config.ZitadelNamespace)

		protocol := "http"
		if config.UseTLS {
			protocol = "https"
		}

		return fmt.Sprintf("%s://%s:%d", protocol, serviceDNS, config.ZitadelPort), nil
	}

	// Use external domain if provided
	if config.InternalDomain != "" {
		protocol := "http"
		if config.UseTLS {
			protocol = "https"
		}
		return fmt.Sprintf("%s://%s:%d", protocol, config.InternalDomain, config.ZitadelPort), nil
	}

	return "", fmt.Errorf("no valid service resolution method configured")
}

// createK8sHTTPClient создает HTTP клиент для k8s окружения
func createK8sHTTPClient(config *K8sConfig) *http.Client {
	transport := &http.Transport{
		MaxIdleConns:        10,
		IdleConnTimeout:     30 * time.Second,
		DisableCompression:  false,
		DisableKeepAlives:   false,
		MaxIdleConnsPerHost: 5,
	}

	// For local k8s, we might need to skip TLS verification
	if config.InsecureSkipVerify {
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	return &http.Client{
		Transport: transport,
		Timeout:   config.HTTPTimeout,
	}
}

// monitorZitadelHealth мониторит состояние Zitadel сервиса
func (c *Client) monitorZitadelHealth(ctx context.Context, config *K8sConfig, logger *zap.Logger) {
	if config.HealthCheckInterval == 0 {
		config.HealthCheckInterval = 30 * time.Second
	}

	ticker := time.NewTicker(config.HealthCheckInterval)
	defer ticker.Stop()

	healthURL := fmt.Sprintf("%s%s", config.Domain, config.HealthCheckPath)
	if config.HealthCheckPath == "" {
		healthURL = fmt.Sprintf("%s/debug/healthz", config.Domain)
	}

	logger.Info("Starting Zitadel health monitoring",
		zap.String("health_url", healthURL),
		zap.Duration("interval", config.HealthCheckInterval))

	for {
		select {
		case <-ctx.Done():
			logger.Info("Stopping Zitadel health monitoring")
			return
		case <-ticker.C:
			if err := c.checkZitadelHealth(healthURL); err != nil {
				logger.Warn("Zitadel health check failed", zap.Error(err))
			} else {
				logger.Debug("Zitadel health check passed")
			}
		}
	}
}

// checkZitadelHealth проверяет здоровье Zitadel сервиса
func (c *Client) checkZitadelHealth(healthURL string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("health check request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status: %d", resp.StatusCode)
	}

	return nil
}

// GetK8sServiceInfo возвращает информацию о сервисе в k8s
func GetK8sServiceInfo(config *K8sConfig) map[string]interface{} {
	return map[string]interface{}{
		"service_name":         config.ZitadelServiceName,
		"namespace":            config.ZitadelNamespace,
		"port":                 config.ZitadelPort,
		"use_internal_dns":     config.UseInternalDNS,
		"use_tls":              config.UseTLS,
		"insecure_skip_verify": config.InsecureSkipVerify,
		"service_discovery":    config.EnableServiceDiscovery,
	}
}
