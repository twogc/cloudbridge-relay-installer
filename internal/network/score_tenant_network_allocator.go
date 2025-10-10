package network

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

// SCORETenantNetworkAllocatorV2 аллокатор подсетей через SCORE с CAS-защитой
type SCORETenantNetworkAllocatorV2 struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	logger     *zap.Logger
	mutex      sync.Mutex
}

// NewSCORETenantNetworkAllocatorV2 создает новый аллокатор для SCORE
func NewSCORETenantNetworkAllocatorV2(baseURL, apiKey string, logger *zap.Logger) *SCORETenantNetworkAllocatorV2 {
	return &SCORETenantNetworkAllocatorV2{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// ReserveSubnet выделяет подсеть для тенанта с CAS-защитой
func (a *SCORETenantNetworkAllocatorV2) ReserveSubnet(ctx context.Context, tenantID string, expected ResourceVersion) (*TenantNetwork, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	// Получаем следующую доступную подсеть
	cidr, err := a.GetNextAvailableSubnet(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get next available subnet: %w", err)
	}

	// Создаем конфигурацию сети
	network := &TenantNetwork{
		TenantID:        tenantID,
		NetworkCIDR:     cidr,
		SubnetCounter:   a.extractSubnetCounter(cidr),
		IsActive:        true,
		AllocatedAtUnix: time.Now().Unix(),
		Version:         1,
		ResourceVersion: expected,
	}

	// Сохраняем в SCORE с CAS-защитой
	_, err = a.saveTenantNetworkWithCAS(ctx, network, expected)
	if err != nil {
		return nil, fmt.Errorf("failed to save tenant network: %w", err)
	}

	a.logger.Info("Reserved subnet for tenant",
		zap.String("tenant_id", tenantID),
		zap.String("network_cidr", cidr),
		zap.String("expected_version", string(expected)))

	return network, nil
}

// ReleaseSubnet освобождает подсеть тенанта с CAS-защитой
func (a *SCORETenantNetworkAllocatorV2) ReleaseSubnet(ctx context.Context, tenantID string, expected ResourceVersion) (ResourceVersion, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	// Получаем текущую конфигурацию
	url := fmt.Sprintf("%s/api/v1/config/tenants/%s/network", a.baseURL, tenantID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	req.Header.Set("X-Service-Name", "cloudbridge-allocator")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get tenant network: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		a.logger.Warn("Tenant network not found for release",
			zap.String("tenant_id", tenantID))
		return "", ErrNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get tenant network: status %d", resp.StatusCode)
	}

	var result struct {
		Success bool           `json:"success"`
		Data    *TenantNetwork `json:"data"`
		Error   string         `json:"error,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.Success {
		return "", fmt.Errorf("SCORE API error: %s", result.Error)
	}

	// Проверяем версию для CAS
	if result.Data.ResourceVersion != expected {
		return "", ErrStaleVersion
	}

	// Деактивируем сеть
	result.Data.IsActive = false
	result.Data.Version++
	result.Data.ResourceVersion = expected

	newVersion, err := a.saveTenantNetworkWithCAS(ctx, result.Data, expected)
	if err != nil {
		return "", fmt.Errorf("failed to deactivate tenant network: %w", err)
	}

	a.logger.Info("Released subnet for tenant",
		zap.String("tenant_id", tenantID),
		zap.String("network_cidr", result.Data.NetworkCIDR),
		zap.String("expected_version", string(expected)))

	return newVersion, nil
}

// GetNextAvailableSubnet возвращает следующую доступную подсеть
func (a *SCORETenantNetworkAllocatorV2) GetNextAvailableSubnet(ctx context.Context) (string, error) {
	// Получаем глобальный реестр занятых подсетей
	allocatedSubnets, err := a.getAllocatedSubnets(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get allocated subnets: %w", err)
	}

	// Генерируем следующую доступную подсеть
	subnetCounter := 80 // Начальное значение
	for {
		cidr := fmt.Sprintf("10.%d.%d.0/24", subnetCounter/256, subnetCounter%256)

		// Проверяем, не занята ли подсеть
		if !a.isSubnetAllocated(cidr, allocatedSubnets) {
			return cidr, nil
		}

		subnetCounter++

		// Защита от бесконечного цикла
		if subnetCounter > 65535 {
			return "", fmt.Errorf("no available subnets")
		}
	}
}

// getAllocatedSubnets получает список занятых подсетей
func (a *SCORETenantNetworkAllocatorV2) getAllocatedSubnets(ctx context.Context) ([]string, error) {
	url := fmt.Sprintf("%s/api/v1/config/allocated-subnets", a.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	req.Header.Set("X-Service-Name", "cloudbridge-allocator")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get allocated subnets: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get allocated subnets: status %d", resp.StatusCode)
	}

	var result struct {
		Success bool `json:"success"`
		Data    struct {
			Items []struct {
				TenantID string `json:"tenantId"`
				CIDR     string `json:"cidr"`
			} `json:"items"`
		} `json:"data"`
		Error string `json:"error,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("SCORE API error: %s", result.Error)
	}

	var subnets []string
	for _, item := range result.Data.Items {
		subnets = append(subnets, item.CIDR)
	}

	return subnets, nil
}

// isSubnetAllocated проверяет, занята ли подсеть
func (a *SCORETenantNetworkAllocatorV2) isSubnetAllocated(cidr string, allocatedSubnets []string) bool {
	for _, allocated := range allocatedSubnets {
		if allocated == cidr {
			return true
		}
	}
	return false
}

// extractSubnetCounter извлекает subnet counter из CIDR
func (a *SCORETenantNetworkAllocatorV2) extractSubnetCounter(cidr string) int {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return 0
	}

	ip := ipNet.IP.To4()
	if ip == nil {
		return 0
	}

	// Формат: 10.x.y.0/24 -> counter = x*256 + y
	return int(ip[1])*256 + int(ip[2])
}

// saveTenantNetworkWithCAS сохраняет конфигурацию сети в SCORE с CAS-защитой
func (a *SCORETenantNetworkAllocatorV2) saveTenantNetworkWithCAS(ctx context.Context, network *TenantNetwork, expected ResourceVersion) (ResourceVersion, error) {
	url := fmt.Sprintf("%s/api/v1/config/tenants/%s/network", a.baseURL, network.TenantID)

	jsonData, err := json.Marshal(network)
	if err != nil {
		return "", fmt.Errorf("failed to marshal network config: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewReader(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Name", "cloudbridge-allocator")
	req.Header.Set("X-Config-Version", fmt.Sprintf("%d", network.Version))
	req.Header.Set("X-Expected-Version", string(expected))

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to save network config: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		return "", ErrConflict
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("failed to save network config: status %d", resp.StatusCode)
	}

	// Получаем новую версию из ответа
	var result struct {
		Success         bool            `json:"success"`
		ResourceVersion ResourceVersion `json:"resource_version"`
		Error           string          `json:"error,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.Success {
		return "", fmt.Errorf("SCORE API error: %s", result.Error)
	}

	return result.ResourceVersion, nil
}
