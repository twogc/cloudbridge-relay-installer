package network

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

// SCORETenantNetworkAllocatorProduction продакшен-готовый аллокатор с правильной retry логикой
type SCORETenantNetworkAllocatorProduction struct {
	baseURL        string
	apiKey         string
	httpClient     *http.Client
	logger         *zap.Logger
	retryConfig    *RetryConfigImproved
	circuitBreaker *CircuitBreaker
	singleFlight   *SingleFlight
	subnetCounter  int // Начальное значение для подсетей
}

// NewSCORETenantNetworkAllocatorProduction создает продакшен-готовый аллокатор
func NewSCORETenantNetworkAllocatorProduction(baseURL, apiKey string, logger *zap.Logger) *SCORETenantNetworkAllocatorProduction {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &SCORETenantNetworkAllocatorProduction{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		logger:         logger,
		retryConfig:    DefaultRetryConfigImproved(),
		circuitBreaker: NewCircuitBreaker(5, 30*time.Second),
		singleFlight:   NewSingleFlight(),
		subnetCounter:  80, // Начальное значение для подсетей
	}
}

// ReserveSubnet выделяет подсеть для тенанта с улучшенной обработкой ошибок
func (a *SCORETenantNetworkAllocatorProduction) ReserveSubnet(ctx context.Context, tenantID string, expected ResourceVersion) (*TenantNetwork, error) {
	// Общий таймаут на операцию
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 15*time.Second)
		defer cancel()
	}

	// Используем singleflight для предотвращения дублирования запросов
	key := fmt.Sprintf("reserve_subnet_%s", tenantID)

	result, err := a.singleFlight.Do(ctx, key, func() (interface{}, error) {
		return a.reserveSubnetInternal(ctx, tenantID, expected)
	})

	if err != nil {
		return nil, err
	}

	return result.(*TenantNetwork), nil
}

// reserveSubnetInternal внутренняя логика резервирования подсети
func (a *SCORETenantNetworkAllocatorProduction) reserveSubnetInternal(ctx context.Context, tenantID string, expected ResourceVersion) (*TenantNetwork, error) {
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

	// Сохраняем в SCORE с улучшенной retry логикой
	err = a.executeWithRetryImproved(ctx, func() error {
		_, err := a.saveTenantNetworkWithCAS(ctx, network, expected)
		return err
	})

	if err != nil {
		return nil, fmt.Errorf("failed to save tenant network: %w", err)
	}

	a.logger.Info("Reserved subnet for tenant",
		zap.String("tenant_id", tenantID),
		zap.String("network_cidr", cidr),
		zap.String("expected_version", string(expected)))

	return network, nil
}

// ReleaseSubnet освобождает подсеть тенанта без глобального mutex
func (a *SCORETenantNetworkAllocatorProduction) ReleaseSubnet(ctx context.Context, tenantID string, expected ResourceVersion) (ResourceVersion, error) {
	// Используем singleflight для предотвращения дублирования запросов
	key := fmt.Sprintf("release_subnet_%s", tenantID)

	result, err := a.singleFlight.Do(ctx, key, func() (interface{}, error) {
		return a.releaseSubnetInternal(ctx, tenantID, expected)
	})

	if err != nil {
		return "", err
	}

	return result.(ResourceVersion), nil
}

// releaseSubnetInternal внутренняя логика освобождения подсети
func (a *SCORETenantNetworkAllocatorProduction) releaseSubnetInternal(ctx context.Context, tenantID string, expected ResourceVersion) (ResourceVersion, error) {
	// Получаем текущую конфигурацию сети с retry
	var netCfg *TenantNetwork
	err := a.executeWithRetryImproved(ctx, func() error {
		var err error
		netCfg, err = a.getTenantNetwork(ctx, tenantID)
		return err
	})

	if err != nil {
		return "", fmt.Errorf("failed to get tenant network: %w", err)
	}

	// Проверяем версию для CAS
	if netCfg.ResourceVersion != expected {
		return "", ErrStaleVersion
	}

	// Создаем новую версию с деактивированной сетью
	netCfg.IsActive = false
	netCfg.Version++
	// ResourceVersion для CAS передаём expected — сервер вернёт новую

	// Сохраняем обновленную конфигурацию с retry логикой
	var newVer ResourceVersion
	err = a.executeWithRetryImproved(ctx, func() error {
		var err error
		newVer, err = a.saveTenantNetworkWithCAS(ctx, netCfg, expected)
		return err
	})

	if err != nil {
		return "", fmt.Errorf("failed to deactivate tenant network: %w", err)
	}

	a.logger.Info("Released subnet for tenant",
		zap.String("tenant_id", tenantID),
		zap.String("network_cidr", netCfg.NetworkCIDR),
		zap.String("expected_version", string(expected)))

	return newVer, nil
}

// executeWithRetryImproved выполняет операцию с улучшенной retry политикой
func (a *SCORETenantNetworkAllocatorProduction) executeWithRetryImproved(ctx context.Context, operation func() error) error {
	var lastErr error

	a.retryConfig.Normalize()

	for attempt := 1; attempt <= a.retryConfig.MaxAttempts; attempt++ {
		// Быстрая проверка: открыт ли CB?
		if a.circuitBreaker != nil && a.circuitBreaker.IsOpen() {
			return ErrCircuitOpen
		}

		err := operation()
		if err == nil {
			if a.circuitBreaker != nil {
				a.circuitBreaker.OnSuccess()
			}
			return nil
		}

		lastErr = err
		if a.circuitBreaker != nil {
			a.circuitBreaker.OnFailure()
		}

		// Определяем, ретраибл ли
		var httpErr *HTTPError
		retryable := false
		delay := a.retryConfig.CalculateDelay(attempt)

		if errors.As(err, &httpErr) {
			retryable = a.retryConfig.IsRetryableError(httpErr.StatusCode)
			// Уважаем Retry-After на 429/503
			if httpErr.StatusCode == http.StatusTooManyRequests || httpErr.StatusCode == http.StatusServiceUnavailable {
				if ra := httpErr.RetryAfter; ra != "" {
					delay = a.retryConfig.CalculateDelayWithRetryAfter(attempt, ra)
				}
			}
		} else {
			// Сетевые и таймауты
			retryable = a.retryConfig.IsRetryableNetError(err)
		}

		if !retryable || attempt == a.retryConfig.MaxAttempts {
			return fmt.Errorf("SCORE operation failed after %d attempts: %w", attempt, lastErr)
		}

		a.logger.Warn("SCORE operation failed, retrying",
			zap.Int("attempt", attempt),
			zap.Duration("delay", delay),
			zap.Error(err))

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}

	return fmt.Errorf("SCORE operation failed after %d attempts: %w", a.retryConfig.MaxAttempts, lastErr)
}

// GetNextAvailableSubnet возвращает следующую доступную подсеть с O(1) поиском
func (a *SCORETenantNetworkAllocatorProduction) GetNextAvailableSubnet(ctx context.Context) (string, error) {
	// Получаем глобальный реестр занятых подсетей
	allocatedSubnets, err := a.getAllocatedSubnets(ctx)
	if err != nil {
		return "", fmt.Errorf("get allocated subnets: %w", err)
	}

	// Преобразуем в map для O(1) поиска
	used := make(map[string]struct{}, len(allocatedSubnets))
	for _, s := range allocatedSubnets {
		used[s] = struct{}{}
	}

	// Ищем доступную подсеть
	subnetCounter := a.subnetCounter
	for subnetCounter <= 65535 {
		cidr := fmt.Sprintf("10.%d.%d.0/24", subnetCounter/256, subnetCounter%256)
		if _, ok := used[cidr]; !ok {
			return cidr, nil
		}
		subnetCounter++
	}

	return "", fmt.Errorf("no available subnets")
}

// getAllocatedSubnets получает список занятых подсетей
func (a *SCORETenantNetworkAllocatorProduction) getAllocatedSubnets(ctx context.Context) ([]string, error) {
	url := fmt.Sprintf("%s/api/v1/config/allocated-subnets", a.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	req.Header.Set("X-Service-Name", "cloudbridge-allocator")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
		return nil, NewHTTPErrorWithRetryAfter(resp.StatusCode, string(b), resp.Header.Get("Retry-After"))
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
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if !result.Success {
		return nil, NewHTTPErrorWithRetryAfter(resp.StatusCode, result.Error, resp.Header.Get("Retry-After"))
	}

	var subnets []string
	for _, item := range result.Data.Items {
		subnets = append(subnets, item.CIDR)
	}

	return subnets, nil
}

// getTenantNetwork получает конфигурацию сети тенанта
func (a *SCORETenantNetworkAllocatorProduction) getTenantNetwork(ctx context.Context, tenantID string) (*TenantNetwork, error) {
	url := fmt.Sprintf("%s/api/v1/config/tenants/%s/network", a.baseURL, tenantID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	req.Header.Set("X-Service-Name", "cloudbridge-allocator")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
		return nil, NewHTTPErrorWithRetryAfter(resp.StatusCode, string(b), resp.Header.Get("Retry-After"))
	}

	var result struct {
		Success bool           `json:"success"`
		Data    *TenantNetwork `json:"data"`
		Error   string         `json:"error,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if !result.Success {
		return nil, NewHTTPErrorWithRetryAfter(resp.StatusCode, result.Error, resp.Header.Get("Retry-After"))
	}

	return result.Data, nil
}

// saveTenantNetworkWithCAS сохраняет конфигурацию сети в SCORE с CAS-защитой
func (a *SCORETenantNetworkAllocatorProduction) saveTenantNetworkWithCAS(ctx context.Context, network *TenantNetwork, expected ResourceVersion) (ResourceVersion, error) {
	url := fmt.Sprintf("%s/api/v1/config/tenants/%s/network", a.baseURL, network.TenantID)

	jsonData, err := json.Marshal(network)
	if err != nil {
		return "", fmt.Errorf("marshal network: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(jsonData))
	if err != nil {
		return "", fmt.Errorf("new request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Name", "cloudbridge-allocator")
	req.Header.Set("X-Config-Version", fmt.Sprintf("%d", network.Version))
	req.Header.Set("X-Expected-Version", string(expected))

	resp, err := a.httpClient.Do(req)
	if err != nil {
		// Сетевые/таймаут — пусть наверх, там IsRetryableNetError решит
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		return "", ErrConflict
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10)) // до 4КБ в лог/ошибку
		return "", NewHTTPErrorWithRetryAfter(resp.StatusCode, strings.TrimSpace(string(b)), resp.Header.Get("Retry-After"))
	}

	var out struct {
		Success         bool            `json:"success"`
		ResourceVersion ResourceVersion `json:"resource_version"`
		Error           string          `json:"error,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	if !out.Success {
		return "", NewHTTPErrorWithRetryAfter(resp.StatusCode, out.Error, resp.Header.Get("Retry-After"))
	}
	return out.ResourceVersion, nil
}

// extractSubnetCounter извлекает subnet counter из CIDR
func (a *SCORETenantNetworkAllocatorProduction) extractSubnetCounter(cidr string) int {
	// Упрощенная реализация для IPv4 /24
	// В реальном проекте здесь будет более сложная логика
	parts := strings.Split(cidr, ".")
	if len(parts) >= 3 {
		// Формат: 10.x.y.0/24 -> counter = x*256 + y
		var x, y int
		if _, err := fmt.Sscanf(parts[1], "%d", &x); err == nil {
			if _, err := fmt.Sscanf(parts[2], "%d", &y); err == nil {
				return x*256 + y
			}
		}
	}
	return 0
}

// GetCircuitBreaker возвращает circuit breaker для мониторинга
func (a *SCORETenantNetworkAllocatorProduction) GetCircuitBreaker() *CircuitBreaker {
	return a.circuitBreaker
}

// GetRetryConfig возвращает retry конфигурацию для мониторинга
func (a *SCORETenantNetworkAllocatorProduction) GetRetryConfig() *RetryConfigImproved {
	return a.retryConfig
}
