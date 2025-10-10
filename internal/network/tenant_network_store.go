package network

import (
	"context"
	"errors"
)

// ResourceVersion версия ресурса для snapshot/watch
type ResourceVersion string

// TenantEventType тип события
type TenantEventType string

const (
	EventCreated TenantEventType = "created"
	EventUpdated TenantEventType = "updated"
	EventDeleted TenantEventType = "deleted"
)

// TenantNetwork представляет сеть тенанта
type TenantNetwork struct {
	TenantID        string          `json:"tenant_id"`
	NetworkCIDR     string          `json:"network_cidr"`
	SubnetCounter   int             `json:"subnet_counter"`
	IsActive        bool            `json:"is_active"`
	AllocatedAtUnix int64           `json:"allocated_at_unix"` // устойчивее, чем time.Time в кросс-языках
	Version         int             `json:"version"`           // логическая версия объекта
	ResourceVersion ResourceVersion `json:"resource_version"`  // версия стора/стрима
}

// TenantNetworkEvent событие изменения сети тенанта
type TenantNetworkEvent struct {
	Type            TenantEventType `json:"type"`
	Tenant          *TenantNetwork  `json:"tenant"`
	PrevVersion     int             `json:"prev_version,omitempty"`
	ResourceVersion ResourceVersion `json:"resource_version"`
	Reason          string          `json:"reason,omitempty"` // e.g. "allocator", "sync", "rollback"
}

// TenantNetworkStore интерфейс для работы с хранилищем подсетей тенантов
type TenantNetworkStore interface {
	// ListActive возвращает все активные сети тенантов (snapshot для старта) с версией
	ListActive(ctx context.Context) (items []*TenantNetwork, at ResourceVersion, err error)

	// Get возвращает сеть конкретного тенанта
	Get(ctx context.Context, tenantID string) (*TenantNetwork, error)

	// Watch подписывается на изменения сетей тенантов от указанной версии
	Watch(ctx context.Context, from ResourceVersion) (<-chan TenantNetworkEvent, error)

	// Close закрывает соединение с хранилищем
	Close() error
}

// TenantNetworkAllocator интерфейс для выделения/освобождения подсетей
type TenantNetworkAllocator interface {
	// ReserveSubnet выделяет подсеть для тенанта с CAS-защитой
	ReserveSubnet(ctx context.Context, tenantID string, expected ResourceVersion) (*TenantNetwork, error)

	// ReleaseSubnet освобождает подсеть тенанта с CAS-защитой
	ReleaseSubnet(ctx context.Context, tenantID string, expected ResourceVersion) (ResourceVersion, error)

	// GetNextAvailableSubnet возвращает следующую доступную подсеть
	GetNextAvailableSubnet(ctx context.Context) (string, error)
}

// TenantNetworkValidator интерфейс для валидации сетей
type TenantNetworkValidator interface {
	// ValidateCIDR проверяет корректность CIDR
	ValidateCIDR(cidr string) error

	// CheckOverlap проверяет пересечение с существующими сетями
	CheckOverlap(ctx context.Context, cidr string) (bool, error)

	// ValidateTenantAccess проверяет доступ тенанта к сети
	ValidateTenantAccess(ctx context.Context, tenantID string) (bool, error)
}

// Ошибки домена
var (
	ErrNotFound         = errors.New("tenant network not found")
	ErrConflict         = errors.New("version conflict")
	ErrStaleVersion     = errors.New("stale version")
	ErrPermissionDenied = errors.New("permission denied")
	ErrInvalidCIDR      = errors.New("invalid CIDR")
	ErrOverlap          = errors.New("network overlap detected")
)
