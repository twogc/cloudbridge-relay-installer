package monitoring

import (
	"context"
	"fmt"
	"runtime"
	"syscall"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/internal/ai/analytics/notifications"
)

// ResourceMonitor мониторинг ресурсов системы
type ResourceMonitor struct {
	notificationSystem *notifications.NotificationSystem
	thresholds         ResourceThresholds
	lastAlerts         map[string]time.Time
	alertCooldown      time.Duration
}

// ResourceThresholds пороги для алертов
type ResourceThresholds struct {
	DiskUsage   float64 `yaml:"disk_usage"`
	CPUUsage    float64 `yaml:"cpu_usage"`
	MemoryUsage float64 `yaml:"memory_usage"`
}

// ResourceUsage текущее использование ресурсов
type ResourceUsage struct {
	DiskUsage   float64            `json:"disk_usage"`
	CPUUsage    float64            `json:"cpu_usage"`
	MemoryUsage float64            `json:"memory_usage"`
	Details     map[string]interface{} `json:"details"`
	Timestamp   time.Time          `json:"timestamp"`
}

// NewResourceMonitor создает новый монитор ресурсов
func NewResourceMonitor(notificationSystem *notifications.NotificationSystem, thresholds ResourceThresholds) *ResourceMonitor {
	return &ResourceMonitor{
		notificationSystem: notificationSystem,
		thresholds:         thresholds,
		lastAlerts:         make(map[string]time.Time),
		alertCooldown:      5 * time.Minute, // 5 минут между алертами
	}
}

// MonitorResources запускает мониторинг ресурсов
func (rm *ResourceMonitor) MonitorResources(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			rm.checkResources()
		}
	}
}

// checkResources проверяет использование ресурсов
func (rm *ResourceMonitor) checkResources() {
	usage, err := rm.getResourceUsage()
	if err != nil {
		fmt.Printf("Error getting resource usage: %v\n", err)
		return
	}

	// Проверка диска
	if usage.DiskUsage >= rm.thresholds.DiskUsage {
		rm.sendResourceAlert("Disk Usage", usage.DiskUsage, rm.thresholds.DiskUsage, usage.Details)
	}

	// Проверка CPU
	if usage.CPUUsage >= rm.thresholds.CPUUsage {
		rm.sendResourceAlert("CPU Usage", usage.CPUUsage, rm.thresholds.CPUUsage, usage.Details)
	}

	// Проверка памяти
	if usage.MemoryUsage >= rm.thresholds.MemoryUsage {
		rm.sendResourceAlert("Memory Usage", usage.MemoryUsage, rm.thresholds.MemoryUsage, usage.Details)
	}
}

// getResourceUsage получает текущее использование ресурсов
func (rm *ResourceMonitor) getResourceUsage() (*ResourceUsage, error) {
	usage := &ResourceUsage{
		Timestamp: time.Now(),
		Details:   make(map[string]interface{}),
	}

	// Получение информации о памяти
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	// Использование памяти в процентах (примерное)
	usage.MemoryUsage = float64(m.Alloc) / float64(m.Sys)
	usage.Details["used_memory_mb"] = float64(m.Alloc) / 1024 / 1024
	usage.Details["total_memory_mb"] = float64(m.Sys) / 1024 / 1024
	usage.Details["available_memory_mb"] = float64(m.Sys-m.Alloc) / 1024 / 1024

	// Получение информации о диске
	diskUsage, err := rm.getDiskUsage()
	if err != nil {
		return nil, fmt.Errorf("failed to get disk usage: %w", err)
	}
	usage.DiskUsage = diskUsage

	// Получение информации о CPU (упрощенное)
	cpuUsage, err := rm.getCPUUsage()
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU usage: %w", err)
	}
	usage.CPUUsage = cpuUsage

	return usage, nil
}

// getDiskUsage получает использование диска
func (rm *ResourceMonitor) getDiskUsage() (float64, error) {
	var stat syscall.Statfs_t
	err := syscall.Statfs("/", &stat)
	if err != nil {
		return 0, err
	}

	// Расчет использования диска в процентах
	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bavail * uint64(stat.Bsize)
	used := total - free

	usage := float64(used) / float64(total)
	return usage, nil
}

// getCPUUsage получает использование CPU (упрощенное)
func (rm *ResourceMonitor) getCPUUsage() (float64, error) {
	// Упрощенная реализация - в реальном проекте нужно использовать
	// более точные методы измерения CPU
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	// Примерное использование CPU на основе количества горутин
	numGoroutines := runtime.NumGoroutine()
	cpuUsage := float64(numGoroutines) / 100.0 // Нормализация
	
	if cpuUsage > 1.0 {
		cpuUsage = 1.0
	}
	
	return cpuUsage, nil
}

// sendResourceAlert отправляет алерт о ресурсах
func (rm *ResourceMonitor) sendResourceAlert(alertType string, usage float64, threshold float64, details map[string]interface{}) {
	// Проверка cooldown для предотвращения спама
	now := time.Now()
	if lastAlert, exists := rm.lastAlerts[alertType]; exists {
		if now.Sub(lastAlert) < rm.alertCooldown {
			return // Пропускаем алерт из-за cooldown
		}
	}

	// Отправка алерта
	err := rm.notificationSystem.SendResourceAlert(alertType, usage, threshold, details)
	if err != nil {
		fmt.Printf("Failed to send resource alert: %v\n", err)
		return
	}

	// Обновление времени последнего алерта
	rm.lastAlerts[alertType] = now
	fmt.Printf("Resource alert sent: %s (%.1f%% > %.1f%%)\n", alertType, usage*100, threshold*100)
}

// GetResourceUsage возвращает текущее использование ресурсов
func (rm *ResourceMonitor) GetResourceUsage() (*ResourceUsage, error) {
	return rm.getResourceUsage()
}

// SetThresholds устанавливает новые пороги
func (rm *ResourceMonitor) SetThresholds(thresholds ResourceThresholds) {
	rm.thresholds = thresholds
}

// SetAlertCooldown устанавливает время между алертами
func (rm *ResourceMonitor) SetAlertCooldown(cooldown time.Duration) {
	rm.alertCooldown = cooldown
}
