package identity

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// ZitadelClient клиент для работы с Zitadel
type ZitadelClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	logger     *zap.Logger
}

// TenantNetworkConfig конфигурация сети тенанта в Zitadel
type TenantNetworkConfig struct {
	TenantID      string    `json:"tenant_id"`
	NetworkCIDR   string    `json:"network_cidr"`
	SubnetCounter int       `json:"subnet_counter"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ZitadelTenantInfo информация о тенанте из Zitadel
type ZitadelTenantInfo struct {
	TenantID      string                 `json:"tenant_id"`
	Name          string                 `json:"name"`
	NetworkConfig *TenantNetworkConfig   `json:"network_config"`
	Attributes    map[string]interface{} `json:"attributes"`
}

// NewZitadelClient создает новый клиент Zitadel
func NewZitadelClient(baseURL, apiKey string, logger *zap.Logger) *ZitadelClient {
	return &ZitadelClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// GetTenantInfo получает информацию о тенанте из Zitadel
func (zc *ZitadelClient) GetTenantInfo(ctx context.Context, tenantID string) (*ZitadelTenantInfo, error) {
	url := fmt.Sprintf("%s/api/v1/tenants/%s", zc.baseURL, tenantID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+zc.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := zc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get tenant info: status %d", resp.StatusCode)
	}

	var tenantInfo ZitadelTenantInfo
	if err := json.NewDecoder(resp.Body).Decode(&tenantInfo); err != nil {
		return nil, fmt.Errorf("failed to decode tenant info: %w", err)
	}

	zc.logger.Debug("Retrieved tenant info from Zitadel",
		zap.String("tenant_id", tenantID),
		zap.String("name", tenantInfo.Name))

	return &tenantInfo, nil
}

// GetTenantNetworkConfig получает конфигурацию сети тенанта
func (zc *ZitadelClient) GetTenantNetworkConfig(ctx context.Context, tenantID string) (*TenantNetworkConfig, error) {
	tenantInfo, err := zc.GetTenantInfo(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant info: %w", err)
	}

	if tenantInfo.NetworkConfig == nil {
		// Если конфигурации сети нет, создаем новую
		return zc.createTenantNetworkConfig(ctx, tenantID)
	}

	return tenantInfo.NetworkConfig, nil
}

// createTenantNetworkConfig создает новую конфигурацию сети для тенанта
func (zc *ZitadelClient) createTenantNetworkConfig(ctx context.Context, tenantID string) (*TenantNetworkConfig, error) {
	// Получаем следующий доступный subnet counter
	subnetCounter, err := zc.getNextSubnetCounter(ctx)
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
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Сохраняем конфигурацию в Zitadel
	if err := zc.updateTenantNetworkConfig(ctx, tenantID, config); err != nil {
		return nil, fmt.Errorf("failed to save network config: %w", err)
	}

	zc.logger.Info("Created new tenant network config",
		zap.String("tenant_id", tenantID),
		zap.String("network_cidr", networkCIDR),
		zap.Int("subnet_counter", subnetCounter))

	return config, nil
}

// getNextSubnetCounter получает следующий доступный subnet counter
func (zc *ZitadelClient) getNextSubnetCounter(ctx context.Context) (int, error) {
	url := fmt.Sprintf("%s/api/v1/tenants/network-counters/next", zc.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+zc.apiKey)

	resp, err := zc.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to get next subnet counter: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("failed to get next subnet counter: status %d", resp.StatusCode)
	}

	var result struct {
		NextCounter int `json:"next_counter"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("failed to decode counter response: %w", err)
	}

	return result.NextCounter, nil
}

// updateTenantNetworkConfig обновляет конфигурацию сети тенанта в Zitadel
func (zc *ZitadelClient) updateTenantNetworkConfig(ctx context.Context, tenantID string, config *TenantNetworkConfig) error {
	url := fmt.Sprintf("%s/api/v1/tenants/%s/network-config", zc.baseURL, tenantID)

	jsonData, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal network config: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+zc.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := zc.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to update network config: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to update network config: status %d", resp.StatusCode)
	}

	zc.logger.Debug("Updated tenant network config in Zitadel",
		zap.String("tenant_id", tenantID),
		zap.String("network_cidr", config.NetworkCIDR))

	return nil
}

// ValidateTenantAccess проверяет доступ тенанта к ресурсам
func (zc *ZitadelClient) ValidateTenantAccess(ctx context.Context, tenantID, resource string) (bool, error) {
	url := fmt.Sprintf("%s/api/v1/tenants/%s/access/%s", zc.baseURL, tenantID, resource)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+zc.apiKey)

	resp, err := zc.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to validate access: %w", err)
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}

// GetTenantPeers получает список пиров тенанта из Zitadel
func (zc *ZitadelClient) GetTenantPeers(ctx context.Context, tenantID string) ([]string, error) {
	url := fmt.Sprintf("%s/api/v1/tenants/%s/peers", zc.baseURL, tenantID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+zc.apiKey)

	resp, err := zc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant peers: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get tenant peers: status %d", resp.StatusCode)
	}

	var result struct {
		Peers []string `json:"peers"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode peers response: %w", err)
	}

	return result.Peers, nil
}
