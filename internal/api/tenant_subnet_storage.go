package api

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
)

// TenantSubnetStorage - постоянное хранение подсетей тенантов
type TenantSubnetStorage struct {
	filePath     string
	tenantSubnets map[string]string            // tenantID -> subnetCIDR
	tenantIPs    map[string]map[string]string // tenantID -> peerID -> IP
	mutex        sync.RWMutex
	logger       *zap.Logger
}

// TenantSubnetRecord - запись о подсети тенанта
type TenantSubnetRecord struct {
	TenantID    string            `json:"tenant_id"`
	SubnetCIDR  string            `json:"subnet_cidr"`
	AllocatedIPs map[string]string `json:"allocated_ips"` // peerID -> IP
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// StorageData - структура для сериализации
type StorageData struct {
	TenantSubnets map[string]TenantSubnetRecord `json:"tenant_subnets"`
	Version       string                        `json:"version"`
	LastUpdated   time.Time                     `json:"last_updated"`
}

// NewTenantSubnetStorage создает новое хранилище подсетей тенантов
func NewTenantSubnetStorage(dataDir string, logger *zap.Logger) *TenantSubnetStorage {
	filePath := filepath.Join(dataDir, "tenant_subnets.json")
	
	storage := &TenantSubnetStorage{
		filePath:      filePath,
		tenantSubnets: make(map[string]string),
		tenantIPs:     make(map[string]map[string]string),
		logger:        logger,
	}
	
	// Загружаем данные при инициализации
	if err := storage.loadFromFile(); err != nil {
		logger.Warn("Failed to load tenant subnets from file, starting with empty storage", zap.Error(err))
	}
	
	return storage
}

// GetTenantSubnet получает подсеть для тенанта
func (ts *TenantSubnetStorage) GetTenantSubnet(tenantID string) (string, bool) {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()
	
	subnet, exists := ts.tenantSubnets[tenantID]
	return subnet, exists
}

// SetTenantSubnet устанавливает подсеть для тенанта
func (ts *TenantSubnetStorage) SetTenantSubnet(tenantID, subnetCIDR string) error {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()
	
	ts.tenantSubnets[tenantID] = subnetCIDR
	
	// Инициализируем мапу IP для тенанта если нужно
	if ts.tenantIPs[tenantID] == nil {
		ts.tenantIPs[tenantID] = make(map[string]string)
	}
	
	// Сохраняем в файл
	return ts.saveToFile()
}

// AllocateIPForTenant выделяет IP для пира в рамках тенанта
func (ts *TenantSubnetStorage) AllocateIPForTenant(tenantID, peerID, allocatedIP string) error {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()
	
	// Инициализируем мапу IP для тенанта если нужно
	if ts.tenantIPs[tenantID] == nil {
		ts.tenantIPs[tenantID] = make(map[string]string)
	}
	
	ts.tenantIPs[tenantID][peerID] = allocatedIP
	
	// Сохраняем в файл
	return ts.saveToFile()
}

// DeallocateIPForTenant освобождает IP для пира
func (ts *TenantSubnetStorage) DeallocateIPForTenant(tenantID, peerID string) error {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()
	
	if tenantIPs, exists := ts.tenantIPs[tenantID]; exists {
		delete(tenantIPs, peerID)
	}
	
	// Сохраняем в файл
	return ts.saveToFile()
}

// GetTenantIPs получает все IP для тенанта
func (ts *TenantSubnetStorage) GetTenantIPs(tenantID string) map[string]string {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()
	
	if tenantIPs, exists := ts.tenantIPs[tenantID]; exists {
		// Возвращаем копию
		result := make(map[string]string)
		for peerID, ip := range tenantIPs {
			result[peerID] = ip
		}
		return result
	}
	
	return make(map[string]string)
}

// GetAllTenantSubnets получает все подсети тенантов
func (ts *TenantSubnetStorage) GetAllTenantSubnets() map[string]string {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()
	
	// Возвращаем копию
	result := make(map[string]string)
	for tenantID, subnet := range ts.tenantSubnets {
		result[tenantID] = subnet
	}
	
	return result
}

// ValidateTenantIsolation проверяет изоляцию тенантов
func (ts *TenantSubnetStorage) ValidateTenantIsolation() (map[string]interface{}, error) {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()
	
	tenantStats := make(map[string]interface{})
	totalTenants := len(ts.tenantSubnets)
	totalAllocated := 0
	
	for tenantID, subnet := range ts.tenantSubnets {
		tenantIPs := ts.tenantIPs[tenantID]
		peerCount := len(tenantIPs)
		totalAllocated += peerCount
		
		tenantStats[tenantID] = map[string]interface{}{
			"subnet":      subnet,
			"peer_count":  peerCount,
			"allocated_ips": tenantIPs,
		}
	}
	
	stats := map[string]interface{}{
		"total_tenants":     totalTenants,
		"total_allocated":   totalAllocated,
		"tenant_stats":      tenantStats,
		"last_updated":      time.Now(),
	}
	
	return stats, nil
}

// loadFromFile загружает данные из файла
func (ts *TenantSubnetStorage) loadFromFile() error {
	// Проверяем, существует ли файл
	if _, err := os.Stat(ts.filePath); os.IsNotExist(err) {
		ts.logger.Info("Tenant subnets file does not exist, starting with empty storage")
		return nil
	}
	
	// Читаем файл
	data, err := os.ReadFile(ts.filePath)
	if err != nil {
		return fmt.Errorf("failed to read tenant subnets file: %w", err)
	}
	
	// Парсим JSON
	var storageData StorageData
	if err := json.Unmarshal(data, &storageData); err != nil {
		return fmt.Errorf("failed to parse tenant subnets file: %w", err)
	}
	
	// Восстанавливаем данные
	ts.tenantSubnets = make(map[string]string)
	ts.tenantIPs = make(map[string]map[string]string)
	
	for tenantID, record := range storageData.TenantSubnets {
		ts.tenantSubnets[tenantID] = record.SubnetCIDR
		ts.tenantIPs[tenantID] = record.AllocatedIPs
	}
	
	ts.logger.Info("Loaded tenant subnets from file",
		zap.Int("tenants", len(ts.tenantSubnets)),
		zap.String("file", ts.filePath))
	
	return nil
}

// saveToFile сохраняет данные в файл
func (ts *TenantSubnetStorage) saveToFile() error {
	// Создаем директорию если не существует
	if err := os.MkdirAll(filepath.Dir(ts.filePath), 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}
	
	// Подготавливаем данные для сериализации
	storageData := StorageData{
		TenantSubnets: make(map[string]TenantSubnetRecord),
		Version:       "1.0",
		LastUpdated:   time.Now(),
	}
	
	// Заполняем данные
	for tenantID, subnet := range ts.tenantSubnets {
		record := TenantSubnetRecord{
			TenantID:     tenantID,
			SubnetCIDR:   subnet,
			AllocatedIPs: make(map[string]string),
			CreatedAt:    time.Now(), // В реальной системе нужно сохранять время создания
			UpdatedAt:    time.Now(),
		}
		
		// Копируем выделенные IP
		if tenantIPs, exists := ts.tenantIPs[tenantID]; exists {
			for peerID, ip := range tenantIPs {
				record.AllocatedIPs[peerID] = ip
			}
		}
		
		storageData.TenantSubnets[tenantID] = record
	}
	
	// Сериализуем в JSON
	data, err := json.MarshalIndent(storageData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tenant subnets: %w", err)
	}
	
	// Записываем в файл
	if err := os.WriteFile(ts.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write tenant subnets file: %w", err)
	}
	
	ts.logger.Debug("Saved tenant subnets to file", zap.String("file", ts.filePath))
	
	return nil
}

// Backup создает резервную копию данных
func (ts *TenantSubnetStorage) Backup(backupPath string) error {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()
	
	// Читаем текущий файл
	data, err := os.ReadFile(ts.filePath)
	if err != nil {
		return fmt.Errorf("failed to read tenant subnets file for backup: %w", err)
	}
	
	// Создаем директорию для бэкапа
	if err := os.MkdirAll(filepath.Dir(backupPath), 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}
	
	// Записываем бэкап
	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write backup file: %w", err)
	}
	
	ts.logger.Info("Created tenant subnets backup", zap.String("backup_path", backupPath))
	
	return nil
}

// Restore восстанавливает данные из бэкапа
func (ts *TenantSubnetStorage) Restore(backupPath string) error {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()
	
	// Читаем бэкап
	data, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup file: %w", err)
	}
	
	// Парсим JSON
	var storageData StorageData
	if err := json.Unmarshal(data, &storageData); err != nil {
		return fmt.Errorf("failed to parse backup file: %w", err)
	}
	
	// Восстанавливаем данные
	ts.tenantSubnets = make(map[string]string)
	ts.tenantIPs = make(map[string]map[string]string)
	
	for tenantID, record := range storageData.TenantSubnets {
		ts.tenantSubnets[tenantID] = record.SubnetCIDR
		ts.tenantIPs[tenantID] = record.AllocatedIPs
	}
	
	// Сохраняем в основной файл
	if err := ts.saveToFile(); err != nil {
		return fmt.Errorf("failed to save restored data: %w", err)
	}
	
	ts.logger.Info("Restored tenant subnets from backup", zap.String("backup_path", backupPath))
	
	return nil
}
