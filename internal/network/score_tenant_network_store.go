package network

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

// SCORETenantNetworkStoreFixed исправленная реализация TenantNetworkStore для SCORE
type SCORETenantNetworkStoreFixed struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	logger     *zap.Logger
	watchers   map[string]chan TenantNetworkEvent
	watchersMu sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewSCORETenantNetworkStoreFixed создает новое хранилище для SCORE
func NewSCORETenantNetworkStoreFixed(baseURL, apiKey string, logger *zap.Logger) *SCORETenantNetworkStoreFixed {
	ctx, cancel := context.WithCancel(context.Background())

	return &SCORETenantNetworkStoreFixed{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger:   logger,
		watchers: make(map[string]chan TenantNetworkEvent),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// ListActive возвращает все активные сети тенантов с версией
func (s *SCORETenantNetworkStoreFixed) ListActive(ctx context.Context) ([]*TenantNetwork, ResourceVersion, error) {
	url := fmt.Sprintf("%s/api/v1/config/tenants/networks", s.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("X-Service-Name", "cloudbridge-relay")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get tenant networks: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("failed to get tenant networks: status %d", resp.StatusCode)
	}

	var result struct {
		Success bool             `json:"success"`
		Data    []*TenantNetwork `json:"data"`
		Error   string           `json:"error,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, "", fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.Success {
		return nil, "", fmt.Errorf("SCORE API error: %s", result.Error)
	}

	// Фильтруем только активные сети
	var activeNetworks []*TenantNetwork
	for _, network := range result.Data {
		if network.IsActive {
			activeNetworks = append(activeNetworks, network)
		}
	}

	s.logger.Debug("Retrieved active tenant networks from SCORE",
		zap.Int("count", len(activeNetworks)))

	return activeNetworks, "score-1", nil
}

// Get возвращает сеть конкретного тенанта
func (s *SCORETenantNetworkStoreFixed) Get(ctx context.Context, tenantID string) (*TenantNetwork, error) {
	url := fmt.Sprintf("%s/api/v1/config/tenants/%s/network", s.baseURL, tenantID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("X-Service-Name", "cloudbridge-relay")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant network: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get tenant network: status %d", resp.StatusCode)
	}

	var result struct {
		Success bool           `json:"success"`
		Data    *TenantNetwork `json:"data"`
		Error   string         `json:"error,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("SCORE API error: %s", result.Error)
	}

	return result.Data, nil
}

// Watch подписывается на изменения сетей тенантов от указанной версии
func (s *SCORETenantNetworkStoreFixed) Watch(ctx context.Context, from ResourceVersion) (<-chan TenantNetworkEvent, error) {
	// Создаем канал для этого watcher'а
	watcherID := fmt.Sprintf("watcher_%d", time.Now().UnixNano())
	eventChan := make(chan TenantNetworkEvent, 1024) // Буферизированный канал

	s.watchersMu.Lock()
	s.watchers[watcherID] = eventChan
	s.watchersMu.Unlock()

	// Запускаем горутину для обработки событий
	go s.handleWatchEvents(ctx, watcherID, eventChan, from)

	return eventChan, nil
}

// handleWatchEvents обрабатывает события от SCORE
func (s *SCORETenantNetworkStoreFixed) handleWatchEvents(ctx context.Context, watcherID string, eventChan chan TenantNetworkEvent, from ResourceVersion) {
	defer func() {
		s.watchersMu.Lock()
		delete(s.watchers, watcherID)
		close(eventChan)
		s.watchersMu.Unlock()
	}()

	url := fmt.Sprintf("%s/api/v1/config/tenants/networks/watch?from=%s", s.baseURL, string(from))

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		s.logger.Error("Failed to create watch request", zap.Error(err))
		return
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("X-Service-Name", "cloudbridge-relay")
	req.Header.Set("Accept", "text/event-stream")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Error("Failed to start watching", zap.Error(err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.logger.Error("Failed to start watching", zap.Int("status", resp.StatusCode))
		return
	}

	// Обрабатываем Server-Sent Events
	decoder := json.NewDecoder(resp.Body)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			var event struct {
				Type            string          `json:"type"`
				Data            *TenantNetwork  `json:"data"`
				PrevVersion     int             `json:"prev_version,omitempty"`
				ResourceVersion ResourceVersion `json:"resource_version"`
				Reason          string          `json:"reason,omitempty"`
			}

			if err := decoder.Decode(&event); err != nil {
				s.logger.Error("Failed to decode watch event", zap.Error(err))
				continue
			}

			if event.Data != nil {
				networkEvent := TenantNetworkEvent{
					Type:            TenantEventType(event.Type),
					Tenant:          event.Data,
					PrevVersion:     event.PrevVersion,
					ResourceVersion: event.ResourceVersion,
					Reason:          event.Reason,
				}

				select {
				case eventChan <- networkEvent:
				case <-ctx.Done():
					return
				default:
					s.logger.Warn("Event channel full, dropping event",
						zap.String("type", event.Type),
						zap.String("tenant_id", event.Data.TenantID),
						zap.String("reason", event.Reason))
				}
			}
		}
	}
}

// Close закрывает соединение с SCORE
func (s *SCORETenantNetworkStoreFixed) Close() error {
	s.cancel()

	s.watchersMu.Lock()
	defer s.watchersMu.Unlock()

	// Закрываем все watcher'ы
	for _, eventChan := range s.watchers {
		close(eventChan)
	}
	s.watchers = make(map[string]chan TenantNetworkEvent)

	return nil
}
