package config

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// SCOREClient клиент для работы с SCORE (Service Configuration Repository)
type SCOREClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	logger     *zap.Logger
}

// TenantNetworkConfig конфигурация сети тенанта в SCORE
type TenantNetworkConfig struct {
	TenantID      string                 `json:"tenant_id"`
	NetworkCIDR   string                 `json:"network_cidr"`
	SubnetCounter int                    `json:"subnet_counter"`
	IsActive      bool                   `json:"is_active"`
	Metadata      map[string]interface{} `json:"metadata"`
	Version       int                    `json:"version"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// SCOREResponse стандартный ответ SCORE API
type SCOREResponse struct {
	Success bool                   `json:"success"`
	Data    interface{}            `json:"data"`
	Error   string                 `json:"error,omitempty"`
	Meta    map[string]interface{} `json:"meta,omitempty"`
}

// NewSCOREClient создает новый клиент SCORE
func NewSCOREClient(baseURL, apiKey string, logger *zap.Logger) *SCOREClient {
	return &SCOREClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// GetTenantNetworkConfig получает конфигурацию сети тенанта из SCORE
func (sc *SCOREClient) GetTenantNetworkConfig(ctx context.Context, tenantID string) (*TenantNetworkConfig, error) {
	url := fmt.Sprintf("%s/api/v1/config/tenants/%s/network", sc.baseURL, tenantID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+sc.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Name", "cloudbridge-relay")

	resp, err := sc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant network config: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		// Конфигурации нет, создаем новую
		return sc.createTenantNetworkConfig(ctx, tenantID)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get tenant network config: status %d", resp.StatusCode)
	}

	var scoreResp SCOREResponse
	if err := json.NewDecoder(resp.Body).Decode(&scoreResp); err != nil {
		return nil, fmt.Errorf("failed to decode SCORE response: %w", err)
	}

	if !scoreResp.Success {
		return nil, fmt.Errorf("SCORE API error: %s", scoreResp.Error)
	}

	// Парсим данные конфигурации
	configData, err := json.Marshal(scoreResp.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config data: %w", err)
	}

	var config TenantNetworkConfig
	if err := json.Unmarshal(configData, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal network config: %w", err)
	}

	sc.logger.Debug("Retrieved tenant network config from SCORE",
		zap.String("tenant_id", tenantID),
		zap.String("network_cidr", config.NetworkCIDR),
		zap.Int("version", config.Version))

	return &config, nil
}

// createTenantNetworkConfig создает новую конфигурацию сети для тенанта
func (sc *SCOREClient) createTenantNetworkConfig(ctx context.Context, tenantID string) (*TenantNetworkConfig, error) {
	// Получаем следующий доступный subnet counter из SCORE
	subnetCounter, err := sc.getNextSubnetCounter(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get next subnet counter: %w", err)
	}

	// Генерируем CIDR для тенанта
	networkCIDR := fmt.Sprintf("10.%d.%d.0/24", subnetCounter/256, subnetCounter%256)

	config := &TenantNetworkConfig{
		TenantID:      tenantID,
		NetworkCIDR:   networkCIDR,
		SubnetCounter: subnetCounter,
		IsActive:      true,
		Version:       1,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		Metadata: map[string]interface{}{
			"service":     "cloudbridge-relay",
			"environment": "production",
			"region":      "us-east-1",
		},
	}

	// Сохраняем конфигурацию в SCORE
	if err := sc.saveTenantNetworkConfig(ctx, config); err != nil {
		return nil, fmt.Errorf("failed to save network config: %w", err)
	}

	sc.logger.Info("Created new tenant network config in SCORE",
		zap.String("tenant_id", tenantID),
		zap.String("network_cidr", networkCIDR),
		zap.Int("subnet_counter", subnetCounter))

	return config, nil
}

// getNextSubnetCounter получает следующий доступный subnet counter из SCORE
func (sc *SCOREClient) getNextSubnetCounter(ctx context.Context) (int, error) {
	url := fmt.Sprintf("%s/api/v1/config/network-counters/next", sc.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+sc.apiKey)
	req.Header.Set("X-Service-Name", "cloudbridge-relay")

	resp, err := sc.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to get next subnet counter: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("failed to get next subnet counter: status %d", resp.StatusCode)
	}

	var scoreResp SCOREResponse
	if err := json.NewDecoder(resp.Body).Decode(&scoreResp); err != nil {
		return 0, fmt.Errorf("failed to decode counter response: %w", err)
	}

	if !scoreResp.Success {
		return 0, fmt.Errorf("SCORE API error: %s", scoreResp.Error)
	}

	// Парсим counter из данных
	counterData, err := json.Marshal(scoreResp.Data)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal counter data: %w", err)
	}

	var result struct {
		NextCounter int `json:"next_counter"`
	}

	if err := json.Unmarshal(counterData, &result); err != nil {
		return 0, fmt.Errorf("failed to unmarshal counter: %w", err)
	}

	return result.NextCounter, nil
}

// saveTenantNetworkConfig сохраняет конфигурацию сети тенанта в SCORE
func (sc *SCOREClient) saveTenantNetworkConfig(ctx context.Context, config *TenantNetworkConfig) error {
	url := fmt.Sprintf("%s/api/v1/config/tenants/%s/network", sc.baseURL, config.TenantID)

	jsonData, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal network config: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+sc.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Name", "cloudbridge-relay")
	req.Header.Set("X-Config-Version", fmt.Sprintf("%d", config.Version))

	resp, err := sc.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to save network config: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to save network config: status %d", resp.StatusCode)
	}

	sc.logger.Debug("Saved tenant network config to SCORE",
		zap.String("tenant_id", config.TenantID),
		zap.String("network_cidr", config.NetworkCIDR),
		zap.Int("version", config.Version))

	return nil
}

// UpdateTenantNetworkConfig обновляет конфигурацию сети тенанта
func (sc *SCOREClient) UpdateTenantNetworkConfig(ctx context.Context, config *TenantNetworkConfig) error {
	config.UpdatedAt = time.Now()
	config.Version++

	return sc.saveTenantNetworkConfig(ctx, config)
}

// DeactivateTenantNetworkConfig деактивирует конфигурацию сети тенанта
func (sc *SCOREClient) DeactivateTenantNetworkConfig(ctx context.Context, tenantID string) error {
	config, err := sc.GetTenantNetworkConfig(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("failed to get tenant network config: %w", err)
	}

	config.IsActive = false
	config.UpdatedAt = time.Now()
	config.Version++

	return sc.saveTenantNetworkConfig(ctx, config)
}

// GetAllTenantNetworkConfigs получает все активные конфигурации сетей
func (sc *SCOREClient) GetAllTenantNetworkConfigs(ctx context.Context) ([]*TenantNetworkConfig, error) {
	url := fmt.Sprintf("%s/api/v1/config/tenants/networks", sc.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+sc.apiKey)
	req.Header.Set("X-Service-Name", "cloudbridge-relay")

	resp, err := sc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get all tenant network configs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get all tenant network configs: status %d", resp.StatusCode)
	}

	var scoreResp SCOREResponse
	if err := json.NewDecoder(resp.Body).Decode(&scoreResp); err != nil {
		return nil, fmt.Errorf("failed to decode SCORE response: %w", err)
	}

	if !scoreResp.Success {
		return nil, fmt.Errorf("SCORE API error: %s", scoreResp.Error)
	}

	// Парсим список конфигураций
	configsData, err := json.Marshal(scoreResp.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal configs data: %w", err)
	}

	var configs []*TenantNetworkConfig
	if err := json.Unmarshal(configsData, &configs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal network configs: %w", err)
	}

	sc.logger.Debug("Retrieved all tenant network configs from SCORE",
		zap.Int("count", len(configs)))

	return configs, nil
}

// WatchTenantNetworkConfigs подписывается на изменения конфигураций сетей
func (sc *SCOREClient) WatchTenantNetworkConfigs(ctx context.Context, callback func(*TenantNetworkConfig)) error {
	url := fmt.Sprintf("%s/api/v1/config/tenants/networks/watch", sc.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+sc.apiKey)
	req.Header.Set("X-Service-Name", "cloudbridge-relay")
	req.Header.Set("Accept", "text/event-stream")

	resp, err := sc.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to watch network configs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to watch network configs: status %d", resp.StatusCode)
	}

	// Обрабатываем Server-Sent Events
	decoder := json.NewDecoder(resp.Body)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			var event struct {
				Type string               `json:"type"`
				Data *TenantNetworkConfig `json:"data"`
			}

			if err := decoder.Decode(&event); err != nil {
				sc.logger.Error("Failed to decode watch event", zap.Error(err))
				continue
			}

			if event.Data != nil {
				callback(event.Data)
			}
		}
	}
}
