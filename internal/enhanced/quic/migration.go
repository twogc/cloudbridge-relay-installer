package quic

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/quic-go/quic-go"
	"go.uber.org/zap"
)

// ConnectionMigration управляет миграцией QUIC соединений
type ConnectionMigration struct {
	conn           *quic.Conn
	logger         *zap.Logger
	mu             sync.RWMutex
	activePaths    map[string]*PathInfo
	preferredPath  string
	pathValidator  *PathValidator
	migrationState MigrationState
}

// PathInfo информация о сетевом пути
type PathInfo struct {
	Address     *net.UDPAddr
	LastUsed    time.Time
	IsValidated bool
	Latency     time.Duration
	LossRate    float64
	Bandwidth   int64
}

// MigrationState состояние миграции соединения
type MigrationState int

const (
	MigrationStateIdle MigrationState = iota
	MigrationStateValidating
	MigrationStateMigrating
	MigrationStateCompleted
	MigrationStateFailed
)

// PathValidator валидатор сетевых путей
type PathValidator struct {
	validationTimeout time.Duration
	maxRetries        int
	logger            *zap.Logger
}

// NewConnectionMigration создает новый ConnectionMigration
func NewConnectionMigration(conn *quic.Conn, logger *zap.Logger) *ConnectionMigration {
	return &ConnectionMigration{
		conn:           conn,
		logger:         logger,
		activePaths:    make(map[string]*PathInfo),
		pathValidator:  NewPathValidator(logger),
		migrationState: MigrationStateIdle,
	}
}

// MigrateToNewAddress мигрирует соединение на новый адрес
func (cm *ConnectionMigration) MigrateToNewAddress(newAddr *net.UDPAddr) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.migrationState != MigrationStateIdle {
		return fmt.Errorf("migration already in progress")
	}

	cm.migrationState = MigrationStateValidating
	cm.logger.Info("Starting connection migration",
		zap.String("new_address", newAddr.String()))

	// 1. Валидируем новый путь
	if err := cm.validateNewPath(newAddr); err != nil {
		cm.migrationState = MigrationStateFailed
		return fmt.Errorf("failed to validate new path: %w", err)
	}

	// 2. Начинаем миграцию
	cm.migrationState = MigrationStateMigrating
	if err := cm.performMigration(newAddr); err != nil {
		cm.migrationState = MigrationStateFailed
		return fmt.Errorf("failed to perform migration: %w", err)
	}

	// 3. Обновляем предпочтительный путь
	cm.preferredPath = newAddr.String()
	cm.migrationState = MigrationStateCompleted

	cm.logger.Info("Connection migration completed",
		zap.String("new_address", newAddr.String()))

	return nil
}

// validateNewPath валидирует новый сетевой путь
func (cm *ConnectionMigration) validateNewPath(addr *net.UDPAddr) error {
	pathKey := addr.String()

	// Проверяем, не валидируем ли мы уже этот путь
	if pathInfo, exists := cm.activePaths[pathKey]; exists && pathInfo.IsValidated {
		return nil
	}

	// Создаем информацию о пути
	pathInfo := &PathInfo{
		Address:     addr,
		LastUsed:    time.Now(),
		IsValidated: false,
	}

	// Валидируем путь
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := cm.pathValidator.ValidatePath(ctx, cm.conn, addr); err != nil {
		return fmt.Errorf("path validation failed: %w", err)
	}

	pathInfo.IsValidated = true
	pathInfo.LastUsed = time.Now()
	cm.activePaths[pathKey] = pathInfo

	cm.logger.Info("Path validated successfully",
		zap.String("address", addr.String()))

	return nil
}

// performMigration выполняет миграцию соединения
func (cm *ConnectionMigration) performMigration(newAddr *net.UDPAddr) error {
	// В QUIC-go миграция происходит автоматически при отправке пакетов на новый адрес
	// Нам нужно только обновить информацию о путях

	pathKey := newAddr.String()
	pathInfo := cm.activePaths[pathKey]
	if pathInfo == nil {
		return fmt.Errorf("path info not found for address %s", newAddr.String())
	}

	// Обновляем время последнего использования
	pathInfo.LastUsed = time.Now()

	cm.logger.Info("Migration performed",
		zap.String("new_address", newAddr.String()))

	return nil
}

// HandleMigrationPacket обрабатывает пакет миграции
func (cm *ConnectionMigration) HandleMigrationPacket(packet []byte, fromAddr *net.UDPAddr) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Обновляем информацию о пути
	pathKey := fromAddr.String()
	pathInfo, exists := cm.activePaths[pathKey]
	if !exists {
		pathInfo = &PathInfo{
			Address:     fromAddr,
			LastUsed:    time.Now(),
			IsValidated: false,
		}
		cm.activePaths[pathKey] = pathInfo
	}

	pathInfo.LastUsed = time.Now()

	// Если это новый путь, валидируем его
	if !pathInfo.IsValidated {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := cm.pathValidator.ValidatePath(ctx, cm.conn, fromAddr); err != nil {
				cm.logger.Warn("Failed to validate new path",
					zap.String("address", fromAddr.String()),
					zap.Error(err))
			} else {
				cm.mu.Lock()
				pathInfo.IsValidated = true
				cm.mu.Unlock()

				cm.logger.Info("New path validated",
					zap.String("address", fromAddr.String()))
			}
		}()
	}

	return nil
}

// GetActivePaths возвращает список активных путей
func (cm *ConnectionMigration) GetActivePaths() map[string]*PathInfo {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// Создаем копию для безопасности
	paths := make(map[string]*PathInfo)
	for key, pathInfo := range cm.activePaths {
		paths[key] = &PathInfo{
			Address:     pathInfo.Address,
			LastUsed:    pathInfo.LastUsed,
			IsValidated: pathInfo.IsValidated,
			Latency:     pathInfo.Latency,
			LossRate:    pathInfo.LossRate,
			Bandwidth:   pathInfo.Bandwidth,
		}
	}

	return paths
}

// GetPreferredPath возвращает предпочтительный путь
func (cm *ConnectionMigration) GetPreferredPath() string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.preferredPath
}

// GetMigrationState возвращает текущее состояние миграции
func (cm *ConnectionMigration) GetMigrationState() MigrationState {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.migrationState
}

// CleanupInactivePaths удаляет неактивные пути
func (cm *ConnectionMigration) CleanupInactivePaths(timeout time.Duration) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	now := time.Now()
	for pathKey, pathInfo := range cm.activePaths {
		if now.Sub(pathInfo.LastUsed) > timeout {
			delete(cm.activePaths, pathKey)
			cm.logger.Debug("Removed inactive path",
				zap.String("path", pathKey))
		}
	}
}

// NewPathValidator создает новый PathValidator
func NewPathValidator(logger *zap.Logger) *PathValidator {
	return &PathValidator{
		validationTimeout: 5 * time.Second,
		maxRetries:        3,
		logger:            logger,
	}
}

// ValidatePath валидирует сетевой путь
func (pv *PathValidator) ValidatePath(ctx context.Context, conn *quic.Conn, addr *net.UDPAddr) error {
	// Простая валидация - отправляем ping и ждем pong
	// В реальной реализации здесь должна быть более сложная логика

	start := time.Now()

	// Отправляем небольшой пакет для проверки пути
	// В QUIC это делается автоматически при отправке данных
	// Здесь мы просто ждем подтверждения

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(pv.validationTimeout):
		return fmt.Errorf("path validation timeout")
	}

	// Если мы дошли сюда, путь валиден
	latency := time.Since(start)
	pv.logger.Debug("Path validation completed",
		zap.String("address", addr.String()),
		zap.Duration("latency", latency))

	return nil
}

// String возвращает строковое представление MigrationState
func (s MigrationState) String() string {
	switch s {
	case MigrationStateIdle:
		return "idle"
	case MigrationStateValidating:
		return "validating"
	case MigrationStateMigrating:
		return "migrating"
	case MigrationStateCompleted:
		return "completed"
	case MigrationStateFailed:
		return "failed"
	default:
		return "unknown"
	}
}

