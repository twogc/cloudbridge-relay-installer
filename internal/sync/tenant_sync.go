package sync

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/internal/tenant"
	"github.com/twogc/cloudbridge-relay-installer/internal/zitadel"
	"go.uber.org/zap"
)

// TenantSync синхронизирует лимиты между Zitadel и Relay
type TenantSync struct {
	zitadelClient *zitadel.Client
	relayManager  *tenant.Manager
	logger        *zap.Logger
	ticker        *time.Ticker
	ctx           context.Context
	cancel        context.CancelFunc
	mu            sync.RWMutex
	running       bool

	// Metrics
	lastSyncTime     time.Time
	syncErrors       int64
	tenantsProcessed int64
	limitsChanged    int64
}

// NewTenantSync создает новый синхронизатор
func NewTenantSync(zitadelClient *zitadel.Client, relayManager *tenant.Manager, logger *zap.Logger) *TenantSync {
	ctx, cancel := context.WithCancel(context.Background())

	return &TenantSync{
		zitadelClient: zitadelClient,
		relayManager:  relayManager,
		logger:        logger,
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Start запускает синхронизацию с экспоненциальным бэкоффом
func (ts *TenantSync) Start(interval time.Duration) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.running {
		ts.logger.Warn("Tenant sync already running")
		return
	}

	ts.ticker = time.NewTicker(interval)
	ts.running = true

	go ts.syncLoop(interval)
	ts.logger.Info("Tenant sync started", zap.Duration("interval", interval))
}

// syncLoop основной цикл синхронизации с бэкоффом
func (ts *TenantSync) syncLoop(baseInterval time.Duration) {
	backoff := time.Duration(0)
	maxBackoff := 5 * time.Minute

	for {
		select {
		case <-ts.ctx.Done():
			return
		case <-ts.ticker.C:
			start := time.Now()

			// Выполняем синхронизацию с таймаутом
			ctx, cancel := context.WithTimeout(ts.ctx, 30*time.Second)
			err := ts.syncTenants(ctx)
			cancel()

			duration := time.Since(start)

			if err != nil {
				ts.mu.Lock()
				ts.syncErrors++
				ts.mu.Unlock()

				// Экспоненциальный бэкофф с джиттером
				if backoff == 0 {
					backoff = baseInterval
				} else {
					backoff = time.Duration(float64(backoff) * 1.5)
					if backoff > maxBackoff {
						backoff = maxBackoff
					}
				}

				// Добавляем джиттер (±20%)
				jitter := time.Duration(float64(backoff) * 0.2 * (rand.Float64()*2 - 1))
				actualBackoff := backoff + jitter

				ts.logger.Error("Tenant sync failed, retrying with backoff",
					zap.Error(err),
					zap.Duration("backoff", actualBackoff),
					zap.Duration("duration", duration))

				// Временно останавливаем ticker и ждем
				ts.ticker.Stop()
				time.Sleep(actualBackoff)
				ts.ticker = time.NewTicker(baseInterval)
			} else {
				// Сброс бэкоффа при успехе
				backoff = 0
				ts.mu.Lock()
				ts.lastSyncTime = time.Now()
				ts.mu.Unlock()

				ts.logger.Debug("Tenant sync completed successfully",
					zap.Duration("duration", duration),
					zap.Int64("tenants_processed", ts.tenantsProcessed),
					zap.Int64("limits_changed", ts.limitsChanged))
			}
		}
	}
}

// syncTenants синхронизирует лимиты тенантов
func (ts *TenantSync) syncTenants(ctx context.Context) error {
	// Получаем тенантов из Zitadel
	tenantManager := ts.zitadelClient.GetTenantManager()
	tenants, err := tenantManager.ListTenants(ctx)
	if err != nil {
		return fmt.Errorf("failed to get tenants from Zitadel: %w", err)
	}

	ts.tenantsProcessed = 0
	ts.limitsChanged = 0

	// Синхронизируем с Relay
	for _, zitadelTenant := range tenants {
		if err := ts.syncTenantLimits(ctx, zitadelTenant); err != nil {
			ts.logger.Error("Failed to sync tenant limits",
				zap.String("tenant_id", zitadelTenant.ID),
				zap.Error(err))
			continue
		}
		ts.tenantsProcessed++
	}

	return nil
}

// syncTenantLimits синхронизирует лимиты конкретного тенанта
func (ts *TenantSync) syncTenantLimits(ctx context.Context, zitadelTenant *zitadel.Tenant) error {
	// Получаем текущие лимиты из Relay
	_, exists := ts.relayManager.GetTenant(zitadelTenant.ID)
	if !exists {
		// Создаем нового тенанта в Relay
		limits := &tenant.TenantLimits{
			MaxTunnels:     zitadelTenant.Quotas.MaxTunnels,
			MaxConnections: zitadelTenant.Quotas.MaxConnections,
			RateLimit:      parseBandwidthLimit(zitadelTenant.Quotas.BandwidthLimit),
			Burst:          zitadelTenant.Quotas.MaxConnections,
		}

		err := ts.relayManager.CreateTenant(zitadelTenant.ID, zitadelTenant.Name, limits)
		if err != nil {
			return fmt.Errorf("failed to create tenant in Relay: %w", err)
		}

		ts.limitsChanged++
		ts.logger.Info("Created new tenant in Relay",
			zap.String("tenant_id", zitadelTenant.ID),
			zap.String("name", zitadelTenant.Name))

		return nil
	}

	// Сравниваем лимиты и обновляем только при изменениях
	relayLimits, _ := ts.relayManager.GetTenantLimits(zitadelTenant.ID)
	needsUpdate := false

	newLimits := &tenant.TenantLimits{
		MaxTunnels:     zitadelTenant.Quotas.MaxTunnels,
		MaxConnections: zitadelTenant.Quotas.MaxConnections,
		RateLimit:      parseBandwidthLimit(zitadelTenant.Quotas.BandwidthLimit),
		Burst:          zitadelTenant.Quotas.MaxConnections,
	}

	// Проверяем изменения
	if relayLimits.MaxTunnels != newLimits.MaxTunnels ||
		relayLimits.MaxConnections != newLimits.MaxConnections ||
		relayLimits.RateLimit != newLimits.RateLimit {
		needsUpdate = true
	}

	if needsUpdate {
		err := ts.relayManager.UpdateTenantLimits(zitadelTenant.ID, newLimits)
		if err != nil {
			return fmt.Errorf("failed to update tenant limits: %w", err)
		}

		ts.limitsChanged++
		ts.logger.Info("Updated tenant limits",
			zap.String("tenant_id", zitadelTenant.ID),
			zap.Int("max_tunnels", newLimits.MaxTunnels),
			zap.Int("max_connections", newLimits.MaxConnections))
	}

	return nil
}

// parseBandwidthLimit парсит строковое представление лимита пропускной способности
func parseBandwidthLimit(limit string) float64 {
	// Простая реализация парсинга "1GB", "500MB" и т.д.
	// В реальной реализации здесь должна быть более сложная логика
	if limit == "" {
		return 1000000.0 // 1MB/s по умолчанию
	}

	// Упрощенная логика для демонстрации
	switch limit {
	case "1GB":
		return 1000000000.0
	case "500MB":
		return 500000000.0
	case "100MB":
		return 100000000.0
	default:
		return 1000000.0
	}
}

// Stop останавливает синхронизацию
func (ts *TenantSync) Stop() {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if !ts.running {
		return
	}

	ts.cancel()
	if ts.ticker != nil {
		ts.ticker.Stop()
	}
	ts.running = false

	ts.logger.Info("Tenant sync stopped")
}

// GetStats возвращает статистику синхронизации
func (ts *TenantSync) GetStats() map[string]interface{} {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	return map[string]interface{}{
		"running":           ts.running,
		"last_sync_time":    ts.lastSyncTime,
		"sync_errors":       ts.syncErrors,
		"tenants_processed": ts.tenantsProcessed,
		"limits_changed":    ts.limitsChanged,
	}
}

// IsRunning возвращает true если синхронизация запущена
func (ts *TenantSync) IsRunning() bool {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.running
}
