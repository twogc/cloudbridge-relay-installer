package p2p

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"github.com/twogc/cloudbridge-relay-installer/internal/wireguard"
)

// CreateWireGuardTunnel создает WireGuard туннель для VIP tenant'а
func (rm *RelayManager) CreateWireGuardTunnel(tenantID string) (*wireguard.WireGuardTenant, error) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	
	// Проверяем, что tenant существует
	tenant, exists := rm.tenants[tenantID]
	if !exists {
		return nil, fmt.Errorf("tenant %s not found", tenantID)
	}
	
	// Проверяем, что tenant имеет VIP статус (здесь должна быть проверка RBAC)
	if !rm.isVIPTenant(tenantID) {
		return nil, fmt.Errorf("tenant %s is not a VIP tenant", tenantID)
	}
	
	// Создаем WireGuard туннель
	wgTenant, err := rm.wireguardManager.CreateTenantTunnel(tenantID)
	if err != nil {
		rm.logger.Error("Failed to create WireGuard tunnel",
			zap.String("tenant_id", tenantID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to create WireGuard tunnel: %w", err)
	}
	
	// Обновляем статус tenant'а
	tenant.LastActivity = time.Now()
	
	rm.logger.Info("Created WireGuard tunnel for VIP tenant",
		zap.String("tenant_id", tenantID),
		zap.String("public_key", wgTenant.PublicKey),
		zap.String("endpoint", wgTenant.Endpoint))
	
	return wgTenant, nil
}

// DeleteWireGuardTunnel удаляет WireGuard туннель для tenant'а
func (rm *RelayManager) DeleteWireGuardTunnel(tenantID string) error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	
	// Удаляем WireGuard туннель
	err := rm.wireguardManager.DeleteTenantTunnel(tenantID)
	if err != nil {
		rm.logger.Error("Failed to delete WireGuard tunnel",
			zap.String("tenant_id", tenantID),
			zap.Error(err))
		return fmt.Errorf("failed to delete WireGuard tunnel: %w", err)
	}
	
	rm.logger.Info("Deleted WireGuard tunnel for tenant",
		zap.String("tenant_id", tenantID))
	
	return nil
}

// GetWireGuardTunnel возвращает информацию о WireGuard туннеле
func (rm *RelayManager) GetWireGuardTunnel(tenantID string) (*wireguard.WireGuardTenant, error) {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()
	
	return rm.wireguardManager.GetTenantTunnel(tenantID)
}

// ListWireGuardTunnels возвращает список всех WireGuard туннелей
func (rm *RelayManager) ListWireGuardTunnels() []*wireguard.WireGuardTenant {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()
	
	return rm.wireguardManager.ListActiveTunnels()
}

// GetWireGuardClientConfig возвращает конфигурацию клиента для WireGuard
func (rm *RelayManager) GetWireGuardClientConfig(tenantID string) (string, error) {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()
	
	// Получаем информацию о туннеле
	wgTenant, err := rm.wireguardManager.GetTenantTunnel(tenantID)
	if err != nil {
		return "", fmt.Errorf("failed to get WireGuard tunnel: %w", err)
	}
	
	// Получаем публичный ключ сервера (здесь должна быть реальная логика)
	serverPublicKey := "SERVER_PUBLIC_KEY_PLACEHOLDER"
	
	// Создаем конфигурационный менеджер
	configManager := wireguard.NewConfigManager("/etc/wireguard", rm.logger)
	
	// Генерируем конфигурацию клиента
	clientConfig := configManager.GenerateClientConfig(wgTenant, serverPublicKey)
	
	return clientConfig, nil
}

// UpdateWireGuardConfig обновляет конфигурацию WireGuard сервера
func (rm *RelayManager) UpdateWireGuardConfig() error {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()
	
	// Получаем список активных туннелей
	activeTunnels := rm.wireguardManager.ListActiveTunnels()
	
	// Создаем конфигурационный менеджер
	configManager := wireguard.NewConfigManager("/etc/wireguard", rm.logger)
	
	// Генерируем конфигурацию сервера
	serverConfig, err := configManager.GenerateServerConfig(activeTunnels)
	if err != nil {
		rm.logger.Error("Failed to generate server config",
			zap.Error(err))
		return fmt.Errorf("failed to generate server config: %w", err)
	}
	
	// Применяем конфигурацию
	err = configManager.ApplyConfig(serverConfig)
	if err != nil {
		rm.logger.Error("Failed to apply WireGuard config",
			zap.Error(err))
		return fmt.Errorf("failed to apply WireGuard config: %w", err)
	}
	
	rm.logger.Info("Updated WireGuard server configuration",
		zap.Int("active_tunnels", len(activeTunnels)))
	
	return nil
}

// GetWireGuardStatus возвращает статус WireGuard интерфейса
func (rm *RelayManager) GetWireGuardStatus() (map[string]interface{}, error) {
	configManager := wireguard.NewConfigManager("/etc/wireguard", rm.logger)
	
	status, err := configManager.GetInterfaceStatus("wg0")
	if err != nil {
		rm.logger.Error("Failed to get WireGuard status",
			zap.Error(err))
		return nil, fmt.Errorf("failed to get WireGuard status: %w", err)
	}
	
	return status, nil
}

// isVIPTenant проверяет, является ли tenant VIP
func (rm *RelayManager) isVIPTenant(tenantID string) bool {
	// Простая проверка - в production должна быть интеграция с RBAC
	// Проверяем по префиксу или другим критериям
	return len(tenantID) > 10 && tenantID[:4] == "vip-"
}

// StartWireGuardManager запускает WireGuard менеджер
func (rm *RelayManager) StartWireGuardManager() error {
	if rm.wireguardManager == nil {
		return fmt.Errorf("WireGuard manager not initialized")
	}
	
	err := rm.wireguardManager.Start()
	if err != nil {
		rm.logger.Error("Failed to start WireGuard manager",
			zap.Error(err))
		return fmt.Errorf("failed to start WireGuard manager: %w", err)
	}
	
	rm.logger.Info("WireGuard manager started successfully")
	return nil
}

// StopWireGuardManager останавливает WireGuard менеджер
func (rm *RelayManager) StopWireGuardManager() {
	if rm.wireguardManager != nil {
		rm.wireguardManager.Stop()
		rm.logger.Info("WireGuard manager stopped")
	}
}

// GetWireGuardMetrics возвращает метрики WireGuard туннелей
func (rm *RelayManager) GetWireGuardMetrics() map[string]interface{} {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()
	
	activeTunnels := rm.wireguardManager.ListActiveTunnels()
	
	metrics := map[string]interface{}{
		"active_tunnels": len(activeTunnels),
		"total_bytes_received": int64(0),
		"total_bytes_sent": int64(0),
		"tunnels": make([]map[string]interface{}, 0),
	}
	
	for _, tunnel := range activeTunnels {
		tunnelMetrics := map[string]interface{}{
			"tenant_id": tunnel.TenantID,
			"is_active": tunnel.IsActive,
			"bytes_received": tunnel.BytesReceived,
			"bytes_sent": tunnel.BytesSent,
			"last_connected": tunnel.LastConnected,
		}
		
		metrics["tunnels"] = append(metrics["tunnels"].([]map[string]interface{}), tunnelMetrics)
		metrics["total_bytes_received"] = metrics["total_bytes_received"].(int64) + tunnel.BytesReceived
		metrics["total_bytes_sent"] = metrics["total_bytes_sent"].(int64) + tunnel.BytesSent
	}
	
	return metrics
}

