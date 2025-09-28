package wireguard

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
	"github.com/twogc/cloudbridge-relay-installer/internal/database"
)

// WireGuardTenantManager управляет WireGuard туннелями для VIP tenant'ов
type WireGuardTenantManager struct {
	tenants     map[string]*WireGuardTenant
	repository  *database.Repository
	logger      *zap.Logger
	mutex       sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
}

// WireGuardTenant представляет WireGuard конфигурацию для tenant'а
type WireGuardTenant struct {
	TenantID        string    `json:"tenant_id"`
	PublicKey       string    `json:"public_key"`
	PrivateKey      string    `json:"private_key"`
	AllowedIPs      []string  `json:"allowed_ips"`
	Endpoint        string    `json:"endpoint"`
	Port            int       `json:"port"`
	PersistentKeepalive int   `json:"persistent_keepalive"`
	IsActive        bool      `json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
	LastConnected   time.Time `json:"last_connected"`
	BytesReceived   int64     `json:"bytes_received"`
	BytesSent       int64     `json:"bytes_sent"`
}

// WireGuardConfig представляет полную конфигурацию WireGuard
type WireGuardConfig struct {
	InterfaceName string            `json:"interface_name"`
	PrivateKey    string            `json:"private_key"`
	Address       string            `json:"address"`
	ListenPort     int              `json:"listen_port"`
	Peers          []WireGuardPeer  `json:"peers"`
}

// WireGuardPeer представляет peer в WireGuard конфигурации
type WireGuardPeer struct {
	PublicKey           string   `json:"public_key"`
	AllowedIPs          []string `json:"allowed_ips"`
	Endpoint            string   `json:"endpoint"`
	PersistentKeepalive int      `json:"persistent_keepalive"`
}

// NewWireGuardTenantManager создает новый менеджер WireGuard tenant'ов
func NewWireGuardTenantManager(repository *database.Repository, logger *zap.Logger) *WireGuardTenantManager {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &WireGuardTenantManager{
		tenants:    make(map[string]*WireGuardTenant),
		repository: repository,
		logger:     logger,
		ctx:        ctx,
		cancel:     cancel,
	}
}

// CreateTenantTunnel создает WireGuard туннель для tenant'а
func (wgtm *WireGuardTenantManager) CreateTenantTunnel(tenantID string) (*WireGuardTenant, error) {
	wgtm.mutex.Lock()
	defer wgtm.mutex.Unlock()
	
	// Проверяем, не существует ли уже туннель
	if existing, exists := wgtm.tenants[tenantID]; exists && existing.IsActive {
		return existing, fmt.Errorf("tunnel for tenant %s already exists", tenantID)
	}
	
	// Генерируем ключи
	privateKey, publicKey, err := wgtm.generateKeyPair()
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %w", err)
	}
	
	// Получаем IP адрес из IPAM
	tenantNetwork := wgtm.getTenantNetwork(tenantID)
	if tenantNetwork == "" {
		return nil, fmt.Errorf("failed to get network for tenant %s", tenantID)
	}
	
	// Создаем WireGuard tenant
	wgTenant := &WireGuardTenant{
		TenantID:            tenantID,
		PublicKey:           publicKey,
		PrivateKey:          privateKey,
		AllowedIPs:          []string{tenantNetwork},
		Endpoint:            wgtm.getWireGuardEndpoint(),
		Port:                wgtm.getAvailablePort(),
		PersistentKeepalive: 25,
		IsActive:            true,
		CreatedAt:           time.Now(),
	}
	
	// Сохраняем в памяти
	wgtm.tenants[tenantID] = wgTenant
	
	// Сохраняем в базе данных
	if wgtm.repository != nil {
		err := wgtm.saveTenantToDatabase(wgTenant)
		if err != nil {
			wgtm.logger.Error("Failed to save tenant to database", 
				zap.String("tenant_id", tenantID), 
				zap.Error(err))
		}
	}
	
	wgtm.logger.Info("Created WireGuard tunnel for tenant",
		zap.String("tenant_id", tenantID),
		zap.String("public_key", publicKey),
		zap.String("endpoint", wgTenant.Endpoint),
		zap.Int("port", wgTenant.Port))
	
	return wgTenant, nil
}

// DeleteTenantTunnel удаляет WireGuard туннель для tenant'а
func (wgtm *WireGuardTenantManager) DeleteTenantTunnel(tenantID string) error {
	wgtm.mutex.Lock()
	defer wgtm.mutex.Unlock()
	
	tenant, exists := wgtm.tenants[tenantID]
	if !exists {
		return fmt.Errorf("tunnel for tenant %s not found", tenantID)
	}
	
	// Деактивируем туннель
	tenant.IsActive = false
	
	// Удаляем из базы данных
	if wgtm.repository != nil {
		err := wgtm.deleteTenantFromDatabase(tenantID)
		if err != nil {
			wgtm.logger.Error("Failed to delete tenant from database",
				zap.String("tenant_id", tenantID),
				zap.Error(err))
		}
	}
	
	// Удаляем из памяти
	delete(wgtm.tenants, tenantID)
	
	wgtm.logger.Info("Deleted WireGuard tunnel for tenant",
		zap.String("tenant_id", tenantID))
	
	return nil
}

// GetTenantTunnel возвращает информацию о туннеле tenant'а
func (wgtm *WireGuardTenantManager) GetTenantTunnel(tenantID string) (*WireGuardTenant, error) {
	wgtm.mutex.RLock()
	defer wgtm.mutex.RUnlock()
	
	tenant, exists := wgtm.tenants[tenantID]
	if !exists {
		return nil, fmt.Errorf("tunnel for tenant %s not found", tenantID)
	}
	
	return tenant, nil
}

// ListActiveTunnels возвращает список активных туннелей
func (wgtm *WireGuardTenantManager) ListActiveTunnels() []*WireGuardTenant {
	wgtm.mutex.RLock()
	defer wgtm.mutex.RUnlock()
	
	var activeTunnels []*WireGuardTenant
	for _, tenant := range wgtm.tenants {
		if tenant.IsActive {
			activeTunnels = append(activeTunnels, tenant)
		}
	}
	
	return activeTunnels
}

// generateKeyPair генерирует пару ключей WireGuard
func (wgtm *WireGuardTenantManager) generateKeyPair() (privateKey, publicKey string, err error) {
	// Генерируем приватный ключ (32 байта)
	privateKeyBytes := make([]byte, 32)
	_, err = rand.Read(privateKeyBytes)
	if err != nil {
		return "", "", err
	}
	
	privateKey = base64.StdEncoding.EncodeToString(privateKeyBytes)
	
	// Для упрощения, используем тот же ключ как публичный
	// В реальной реализации нужно использовать криптографические функции
	publicKey = base64.StdEncoding.EncodeToString(privateKeyBytes)
	
	return privateKey, publicKey, nil
}

// getTenantNetwork получает сеть tenant'а из IPAM
func (wgtm *WireGuardTenantManager) getTenantNetwork(tenantID string) string {
	// Интеграция с существующей IPAM системой
	// Здесь должен быть вызов к RelayManager.GetTenantNetwork()
	// Для упрощения возвращаем заглушку
	return "10.0.100.0/24" // Пример сети для WireGuard
}

// getWireGuardEndpoint возвращает endpoint для WireGuard
func (wgtm *WireGuardTenantManager) getWireGuardEndpoint() string {
	// В реальной реализации это должен быть внешний IP сервера
	return "cloudbridge.example.com"
}

// getAvailablePort возвращает доступный порт для WireGuard
func (wgtm *WireGuardTenantManager) getAvailablePort() int {
	// Простая реализация - в production нужен более сложный алгоритм
	return 51820 + len(wgtm.tenants)
}

// saveTenantToDatabase сохраняет tenant в базу данных
func (wgtm *WireGuardTenantManager) saveTenantToDatabase(tenant *WireGuardTenant) error {
	// Здесь должна быть реализация сохранения в PostgreSQL
	// Пока что заглушка
	return nil
}

// deleteTenantFromDatabase удаляет tenant из базы данных
func (wgtm *WireGuardTenantManager) deleteTenantFromDatabase(tenantID string) error {
	// Здесь должна быть реализация удаления из PostgreSQL
	// Пока что заглушка
	return nil
}

// Start запускает WireGuard менеджер
func (wgtm *WireGuardTenantManager) Start() error {
	wgtm.logger.Info("Starting WireGuard Tenant Manager")
	
	// Загружаем существующие туннели из базы данных
	if wgtm.repository != nil {
		err := wgtm.loadTenantsFromDatabase()
		if err != nil {
			wgtm.logger.Error("Failed to load tenants from database", zap.Error(err))
		}
	}
	
	// Запускаем мониторинг туннелей
	go wgtm.monitorTunnels()
	
	return nil
}

// Stop останавливает WireGuard менеджер
func (wgtm *WireGuardTenantManager) Stop() {
	wgtm.logger.Info("Stopping WireGuard Tenant Manager")
	wgtm.cancel()
}

// loadTenantsFromDatabase загружает tenant'ов из базы данных
func (wgtm *WireGuardTenantManager) loadTenantsFromDatabase() error {
	// Реализация загрузки из PostgreSQL
	return nil
}

// monitorTunnels мониторит состояние туннелей
func (wgtm *WireGuardTenantManager) monitorTunnels() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-wgtm.ctx.Done():
			return
		case <-ticker.C:
			wgtm.checkTunnelHealth()
		}
	}
}

// checkTunnelHealth проверяет здоровье туннелей
func (wgtm *WireGuardTenantManager) checkTunnelHealth() {
	wgtm.mutex.RLock()
	defer wgtm.mutex.RUnlock()
	
	for tenantID, tenant := range wgtm.tenants {
		if tenant.IsActive {
			// Здесь должна быть проверка состояния туннеля
			// Пока что просто логируем
			wgtm.logger.Debug("Checking tunnel health",
				zap.String("tenant_id", tenantID),
				zap.Bool("is_active", tenant.IsActive))
		}
	}
}

