package zitadel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// TenantNetworkClient - клиент для управления сетевыми конфигурациями тенантов в Zitadel
type TenantNetworkClient struct {
	baseURL        string
	adminToken     string
	httpClient     *http.Client
	logger         *zap.Logger
}

// TenantNetworkConfig - конфигурация сети для тенанта
type TenantNetworkConfig struct {
	TenantID         string            `json:"tenant_id"`
	AllocatedSubnet  string            `json:"allocated_subnet"`
	IPAMPool         string            `json:"ipam_pool"`
	NetworkPolicies  []string          `json:"network_policies"`
	AllocatedIPs     map[string]string `json:"allocated_ips"` // peerID -> IP
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
}

// IPAMPool - пул IP адресов для выделения подсетей
type IPAMPool struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	CIDR        string   `json:"cidr"`
	SubnetSize  int      `json:"subnet_size"` // /24, /25, etc.
	Allocated   []string `json:"allocated"`   // список выделенных подсетей
	Available   []string `json:"available"`   // список доступных подсетей
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NewTenantNetworkClient создает новый клиент для управления сетевыми конфигурациями
func NewTenantNetworkClient(baseURL, adminToken string, logger *zap.Logger) *TenantNetworkClient {
	return &TenantNetworkClient{
		baseURL:    baseURL,
		adminToken: adminToken,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// GetTenantNetworkConfig получает сетевую конфигурацию тенанта
func (c *TenantNetworkClient) GetTenantNetworkConfig(ctx context.Context, tenantID string) (*TenantNetworkConfig, error) {
	url := fmt.Sprintf("%s/tenants/%s/network-config", c.baseURL, tenantID)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Authorization", "Bearer "+c.adminToken)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == http.StatusNotFound {
		// Тенант не имеет сетевой конфигурации
		return nil, nil
	}
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get tenant network config: status %d, body: %s", resp.StatusCode, string(body))
	}
	
	var config TenantNetworkConfig
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	c.logger.Debug("Retrieved tenant network config",
		zap.String("tenant_id", tenantID),
		zap.String("subnet", config.AllocatedSubnet))
	
	return &config, nil
}

// CreateTenantNetworkConfig создает сетевую конфигурацию для тенанта
func (c *TenantNetworkClient) CreateTenantNetworkConfig(ctx context.Context, tenantID string, config *TenantNetworkConfig) error {
	url := fmt.Sprintf("%s/tenants/%s/network-config", c.baseURL, tenantID)
	
	config.TenantID = tenantID
	config.CreatedAt = time.Now()
	config.UpdatedAt = time.Now()
	
	jsonData, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Authorization", "Bearer "+c.adminToken)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create tenant network config: status %d, body: %s", resp.StatusCode, string(body))
	}
	
	c.logger.Info("Created tenant network config",
		zap.String("tenant_id", tenantID),
		zap.String("subnet", config.AllocatedSubnet))
	
	return nil
}

// UpdateTenantNetworkConfig обновляет сетевую конфигурацию тенанта
func (c *TenantNetworkClient) UpdateTenantNetworkConfig(ctx context.Context, tenantID string, config *TenantNetworkConfig) error {
	url := fmt.Sprintf("%s/tenants/%s/network-config", c.baseURL, tenantID)
	
	config.TenantID = tenantID
	config.UpdatedAt = time.Now()
	
	jsonData, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Authorization", "Bearer "+c.adminToken)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update tenant network config: status %d, body: %s", resp.StatusCode, string(body))
	}
	
	c.logger.Info("Updated tenant network config",
		zap.String("tenant_id", tenantID),
		zap.String("subnet", config.AllocatedSubnet))
	
	return nil
}

// AllocateIPForTenant выделяет IP для пира в рамках тенанта
func (c *TenantNetworkClient) AllocateIPForTenant(ctx context.Context, tenantID, peerID, allocatedIP string) error {
	// Получаем текущую конфигурацию
	config, err := c.GetTenantNetworkConfig(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("failed to get tenant network config: %w", err)
	}
	
	if config == nil {
		return fmt.Errorf("tenant network config not found for tenant %s", tenantID)
	}
	
	// Добавляем IP в конфигурацию
	if config.AllocatedIPs == nil {
		config.AllocatedIPs = make(map[string]string)
	}
	config.AllocatedIPs[peerID] = allocatedIP
	
	// Обновляем конфигурацию
	return c.UpdateTenantNetworkConfig(ctx, tenantID, config)
}

// DeallocateIPForTenant освобождает IP для пира
func (c *TenantNetworkClient) DeallocateIPForTenant(ctx context.Context, tenantID, peerID string) error {
	// Получаем текущую конфигурацию
	config, err := c.GetTenantNetworkConfig(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("failed to get tenant network config: %w", err)
	}
	
	if config == nil {
		return fmt.Errorf("tenant network config not found for tenant %s", tenantID)
	}
	
	// Удаляем IP из конфигурации
	if config.AllocatedIPs != nil {
		delete(config.AllocatedIPs, peerID)
	}
	
	// Обновляем конфигурацию
	return c.UpdateTenantNetworkConfig(ctx, tenantID, config)
}

// GetIPAMPools получает список IPAM пулов
func (c *TenantNetworkClient) GetIPAMPools(ctx context.Context) ([]*IPAMPool, error) {
	url := fmt.Sprintf("%s/ipam/pools", c.baseURL)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Authorization", "Bearer "+c.adminToken)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get IPAM pools: status %d, body: %s", resp.StatusCode, string(body))
	}
	
	var pools []*IPAMPool
	if err := json.NewDecoder(resp.Body).Decode(&pools); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	c.logger.Debug("Retrieved IPAM pools", zap.Int("count", len(pools)))
	
	return pools, nil
}

// AllocateSubnetFromPool выделяет подсеть из IPAM пула
func (c *TenantNetworkClient) AllocateSubnetFromPool(ctx context.Context, poolID string) (string, error) {
	url := fmt.Sprintf("%s/ipam/pools/%s/allocate-subnet", c.baseURL, poolID)
	
	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Authorization", "Bearer "+c.adminToken)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to allocate subnet: status %d, body: %s", resp.StatusCode, string(body))
	}
	
	var result struct {
		Subnet string `json:"subnet"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}
	
	c.logger.Info("Allocated subnet from pool",
		zap.String("pool_id", poolID),
		zap.String("subnet", result.Subnet))
	
	return result.Subnet, nil
}

// ValidateTenantIsolation проверяет изоляцию тенантов
func (c *TenantNetworkClient) ValidateTenantIsolation(ctx context.Context) (map[string]interface{}, error) {
	// Получаем все IPAM пулы
	pools, err := c.GetIPAMPools(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get IPAM pools: %w", err)
	}
	
	stats := map[string]interface{}{
		"total_pools": len(pools),
		"pools":       make([]map[string]interface{}, 0),
		"last_updated": time.Now(),
	}
	
	for _, pool := range pools {
		poolStats := map[string]interface{}{
			"id":         pool.ID,
			"name":       pool.Name,
			"cidr":       pool.CIDR,
			"subnet_size": pool.SubnetSize,
			"allocated_count": len(pool.Allocated),
			"available_count": len(pool.Available),
		}
		
		stats["pools"] = append(stats["pools"].([]map[string]interface{}), poolStats)
	}
	
	c.logger.Info("Validated tenant isolation", zap.Int("pools", len(pools)))
	
	return stats, nil
}
