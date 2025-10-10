package network

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/internal/database"
	"go.uber.org/zap"
)

// DBTenantNetworkStore реализация TenantNetworkStore для базы данных
type DBTenantNetworkStore struct {
	repository *database.Repository
	logger     *zap.Logger
	watchers   map[string]chan TenantNetworkEvent
	watchersMu sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewDBTenantNetworkStore создает новое хранилище для базы данных
func NewDBTenantNetworkStore(repository *database.Repository, logger *zap.Logger) *DBTenantNetworkStore {
	ctx, cancel := context.WithCancel(context.Background())

	return &DBTenantNetworkStore{
		repository: repository,
		logger:     logger,
		watchers:   make(map[string]chan TenantNetworkEvent),
		ctx:        ctx,
		cancel:     cancel,
	}
}

// ListActive возвращает все активные сети тенантов с версией
func (s *DBTenantNetworkStore) ListActive(ctx context.Context) ([]*TenantNetwork, ResourceVersion, error) {
	networks, err := s.repository.GetAllTenantNetworks()
	if err != nil {
		return nil, "", fmt.Errorf("failed to get tenant networks from database: %w", err)
	}

	var result []*TenantNetwork
	for _, network := range networks {
		if network.IsActive {
			result = append(result, &TenantNetwork{
				TenantID:        network.TenantID,
				NetworkCIDR:     network.NetworkCIDR,
				SubnetCounter:   network.SubnetCounter,
				IsActive:        network.IsActive,
				AllocatedAtUnix: network.AllocatedAt.Unix(),
				Version:         network.Version,
			})
		}
	}

	s.logger.Debug("Retrieved active tenant networks from database",
		zap.Int("count", len(result)))

	return result, "db-1", nil
}

// Get возвращает сеть конкретного тенанта
func (s *DBTenantNetworkStore) Get(ctx context.Context, tenantID string) (*TenantNetwork, error) {
	network, err := s.repository.GetTenantNetwork(tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant network from database: %w", err)
	}

	return &TenantNetwork{
		TenantID:        network.TenantID,
		NetworkCIDR:     network.NetworkCIDR,
		SubnetCounter:   network.SubnetCounter,
		IsActive:        network.IsActive,
		AllocatedAtUnix: network.AllocatedAt.Unix(),
		Version:         network.Version,
	}, nil
}

// Watch подписывается на изменения сетей тенантов (заглушка для БД)
func (s *DBTenantNetworkStore) Watch(ctx context.Context, from ResourceVersion) (<-chan TenantNetworkEvent, error) {
	// Для базы данных реализуем простой polling
	watcherID := fmt.Sprintf("watcher_%d", time.Now().UnixNano())
	eventChan := make(chan TenantNetworkEvent, 100)

	s.watchersMu.Lock()
	s.watchers[watcherID] = eventChan
	s.watchersMu.Unlock()

	// Запускаем горутину для polling
	go s.handleWatchEvents(ctx, watcherID, eventChan)

	return eventChan, nil
}

// handleWatchEvents обрабатывает события через polling
func (s *DBTenantNetworkStore) handleWatchEvents(ctx context.Context, watcherID string, eventChan chan TenantNetworkEvent) {
	defer func() {
		s.watchersMu.Lock()
		delete(s.watchers, watcherID)
		close(eventChan)
		s.watchersMu.Unlock()
	}()

	ticker := time.NewTicker(30 * time.Second) // Polling каждые 30 секунд
	defer ticker.Stop()

	// Кэш для отслеживания изменений
	lastNetworks := make(map[string]*TenantNetwork)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Получаем текущие сети
			currentNetworks, _, err := s.ListActive(ctx)
			if err != nil {
				s.logger.Error("Failed to get current networks", zap.Error(err))
				continue
			}

			// Проверяем изменения
			currentMap := make(map[string]*TenantNetwork)
			for _, network := range currentNetworks {
				currentMap[network.TenantID] = network
			}

			// Проверяем новые/обновленные сети
			for tenantID, network := range currentMap {
				if lastNetwork, exists := lastNetworks[tenantID]; !exists {
					// Новая сеть
					eventChan <- TenantNetworkEvent{
						Type:   "created",
						Tenant: network,
					}
				} else if lastNetwork.Version != network.Version {
					// Обновленная сеть
					eventChan <- TenantNetworkEvent{
						Type:   "updated",
						Tenant: network,
					}
				}
			}

			// Проверяем удаленные сети
			for tenantID, lastNetwork := range lastNetworks {
				if _, exists := currentMap[tenantID]; !exists {
					// Удаленная сеть
					eventChan <- TenantNetworkEvent{
						Type:   "deleted",
						Tenant: lastNetwork,
					}
				}
			}

			// Обновляем кэш
			lastNetworks = currentMap
		}
	}
}

// Close закрывает соединение с базой данных
func (s *DBTenantNetworkStore) Close() error {
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
