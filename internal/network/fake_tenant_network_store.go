package network

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// FakeTenantNetworkStore заглушка для тестирования и локальной разработки
type FakeTenantNetworkStore struct {
	networks        map[string]*TenantNetwork
	resourceVersion ResourceVersion
	mutex           sync.RWMutex
	logger          *zap.Logger
	watchers        map[string]chan TenantNetworkEvent
	watchersMu      sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
}

// NewFakeTenantNetworkStore создает новое fake хранилище
func NewFakeTenantNetworkStore(logger *zap.Logger) *FakeTenantNetworkStore {
	ctx, cancel := context.WithCancel(context.Background())

	return &FakeTenantNetworkStore{
		networks:        make(map[string]*TenantNetwork),
		resourceVersion: "fake-1",
		logger:          logger,
		watchers:        make(map[string]chan TenantNetworkEvent),
		ctx:             ctx,
		cancel:          cancel,
	}
}

// ListActive возвращает все активные сети тенантов с версией
func (f *FakeTenantNetworkStore) ListActive(ctx context.Context) ([]*TenantNetwork, ResourceVersion, error) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	var activeNetworks []*TenantNetwork
	for _, network := range f.networks {
		if network.IsActive {
			activeNetworks = append(activeNetworks, network)
		}
	}

	f.logger.Debug("Retrieved active tenant networks from fake store",
		zap.Int("count", len(activeNetworks)),
		zap.String("resource_version", string(f.resourceVersion)))

	return activeNetworks, f.resourceVersion, nil
}

// Get возвращает сеть конкретного тенанта
func (f *FakeTenantNetworkStore) Get(ctx context.Context, tenantID string) (*TenantNetwork, error) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	network, exists := f.networks[tenantID]
	if !exists {
		return nil, ErrNotFound
	}

	return network, nil
}

// Watch подписывается на изменения сетей тенантов от указанной версии
func (f *FakeTenantNetworkStore) Watch(ctx context.Context, from ResourceVersion) (<-chan TenantNetworkEvent, error) {
	watcherID := fmt.Sprintf("watcher_%d", time.Now().UnixNano())
	eventChan := make(chan TenantNetworkEvent, 1024) // Буферизированный канал

	f.watchersMu.Lock()
	f.watchers[watcherID] = eventChan
	f.watchersMu.Unlock()

	// Запускаем горутину для обработки событий
	go f.handleWatchEvents(ctx, watcherID, eventChan)

	return eventChan, nil
}

// handleWatchEvents обрабатывает события (заглушка)
func (f *FakeTenantNetworkStore) handleWatchEvents(ctx context.Context, watcherID string, eventChan chan TenantNetworkEvent) {
	defer func() {
		f.watchersMu.Lock()
		delete(f.watchers, watcherID)
		close(eventChan)
		f.watchersMu.Unlock()
	}()

	// В fake store просто ждем отмены контекста
	<-ctx.Done()
}

// Close закрывает соединение
func (f *FakeTenantNetworkStore) Close() error {
	f.cancel()

	f.watchersMu.Lock()
	defer f.watchersMu.Unlock()

	// Закрываем все watcher'ы
	for _, eventChan := range f.watchers {
		close(eventChan)
	}
	f.watchers = make(map[string]chan TenantNetworkEvent)

	return nil
}

// AddNetwork добавляет сеть в fake store (для тестирования)
func (f *FakeTenantNetworkStore) AddNetwork(network *TenantNetwork) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	network.ResourceVersion = f.resourceVersion
	f.networks[network.TenantID] = network

	// Отправляем событие
	f.broadcastEvent(TenantNetworkEvent{
		Type:            EventCreated,
		Tenant:          network,
		ResourceVersion: f.resourceVersion,
		Reason:          "fake-store",
	})

	f.logger.Debug("Added network to fake store",
		zap.String("tenant_id", network.TenantID),
		zap.String("network_cidr", network.NetworkCIDR))
}

// UpdateNetwork обновляет сеть в fake store (для тестирования)
func (f *FakeTenantNetworkStore) UpdateNetwork(network *TenantNetwork) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	oldNetwork := f.networks[network.TenantID]
	network.ResourceVersion = f.resourceVersion
	f.networks[network.TenantID] = network

	// Отправляем событие
	f.broadcastEvent(TenantNetworkEvent{
		Type:            EventUpdated,
		Tenant:          network,
		PrevVersion:     oldNetwork.Version,
		ResourceVersion: f.resourceVersion,
		Reason:          "fake-store",
	})

	f.logger.Debug("Updated network in fake store",
		zap.String("tenant_id", network.TenantID),
		zap.String("network_cidr", network.NetworkCIDR))
}

// RemoveNetwork удаляет сеть из fake store (для тестирования)
func (f *FakeTenantNetworkStore) RemoveNetwork(tenantID string) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	network, exists := f.networks[tenantID]
	if !exists {
		return
	}

	delete(f.networks, tenantID)

	// Отправляем событие
	f.broadcastEvent(TenantNetworkEvent{
		Type:            EventDeleted,
		Tenant:          network,
		ResourceVersion: f.resourceVersion,
		Reason:          "fake-store",
	})

	f.logger.Debug("Removed network from fake store",
		zap.String("tenant_id", tenantID))
}

// broadcastEvent отправляет событие всем watcher'ам
func (f *FakeTenantNetworkStore) broadcastEvent(event TenantNetworkEvent) {
	f.watchersMu.RLock()
	defer f.watchersMu.RUnlock()

	for _, eventChan := range f.watchers {
		select {
		case eventChan <- event:
		default:
			f.logger.Warn("Event channel full, dropping event",
				zap.String("type", string(event.Type)),
				zap.String("tenant_id", event.Tenant.TenantID))
		}
	}
}

// GetNetworks возвращает все сети (для тестирования)
func (f *FakeTenantNetworkStore) GetNetworks() map[string]*TenantNetwork {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	networks := make(map[string]*TenantNetwork)
	for id, network := range f.networks {
		networks[id] = network
	}

	return networks
}

// SetResourceVersion устанавливает resource version (для тестирования)
func (f *FakeTenantNetworkStore) SetResourceVersion(version ResourceVersion) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.resourceVersion = version
}

