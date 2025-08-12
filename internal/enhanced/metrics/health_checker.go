package metrics

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// HealthCheck интерфейс для health checks
type HealthCheck interface {
	Check(ctx context.Context) error
	Name() string
	Description() string
}

// HealthStatus статус health check
type HealthStatus struct {
	Status      string                 `json:"status"`
	Timestamp   time.Time              `json:"timestamp"`
	Uptime      string                 `json:"uptime"`
	Version     string                 `json:"version"`
	Checks      map[string]CheckResult `json:"checks"`
	Ready       bool                   `json:"ready"`
	ReadyReason string                 `json:"ready_reason,omitempty"`
}

// CheckResult результат отдельного health check
type CheckResult struct {
	Status      string        `json:"status"`
	Duration    time.Duration `json:"duration"`
	Error       string        `json:"error,omitempty"`
	LastCheck   time.Time     `json:"last_check"`
	Description string        `json:"description"`
}

// EnhancedHealthChecker расширенный health checker с дополнительными возможностями
type EnhancedHealthChecker struct {
	mu          sync.RWMutex
	startTime   time.Time
	version     string
	checks      map[string]HealthCheck
	status      HealthStatus
	ready       bool
	readyReason string
	logger      *zap.Logger
	metrics     *EnhancedMetrics
	connectorID uuid.UUID
	readyChecks map[string]bool
}

// NewEnhancedHealthChecker создает новый расширенный health checker
func NewEnhancedHealthChecker(version string, logger *zap.Logger, metrics *EnhancedMetrics) *EnhancedHealthChecker {
	now := time.Now()
	checker := &EnhancedHealthChecker{
		startTime:   now,
		version:     version,
		checks:      make(map[string]HealthCheck),
		readyChecks: make(map[string]bool),
		logger:      logger,
		metrics:     metrics,
		connectorID: uuid.New(),
		status: HealthStatus{
			Status:    "ok",
			Timestamp: now,
			Uptime:    "0s",
			Version:   version,
			Checks:    make(map[string]CheckResult),
			Ready:     false,
		},
	}

	// Добавляем базовые health checks
	checker.addDefaultChecks()

	return checker
}

// addDefaultChecks добавляет базовые health checks
func (hc *EnhancedHealthChecker) addDefaultChecks() {
	// Database check
	hc.AddCheck(&DatabaseHealthCheck{
		name:   "database",
		logger: hc.logger,
	})

	// Redis check
	hc.AddCheck(&RedisHealthCheck{
		name:   "redis",
		logger: hc.logger,
	})

	// Keycloak check
	hc.AddCheck(&KeycloakHealthCheck{
		name:   "keycloak",
		logger: hc.logger,
	})

	// TCP Proxy check
	hc.AddCheck(&TCPProxyHealthCheck{
		name:   "tcp_proxy",
		logger: hc.logger,
	})

	// System check
	hc.AddCheck(&SystemHealthCheck{
		name:   "system",
		logger: hc.logger,
	})
}

// AddCheck добавляет новый health check
func (hc *EnhancedHealthChecker) AddCheck(check HealthCheck) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	hc.checks[check.Name()] = check
}

// RemoveCheck удаляет health check
func (hc *EnhancedHealthChecker) RemoveCheck(name string) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	delete(hc.checks, name)
	delete(hc.readyChecks, name)
}

// RunChecks выполняет все health checks
func (hc *EnhancedHealthChecker) RunChecks(ctx context.Context) HealthStatus {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	// Обновляем uptime
	hc.status.Uptime = time.Since(hc.startTime).String()
	hc.status.Timestamp = time.Now()

	// Выполняем все checks
	var allHealthy bool = true
	var readyChecks int = 0

	for name, check := range hc.checks {
		start := time.Now()
		err := check.Check(ctx)
		duration := time.Since(start)

		result := CheckResult{
			Status:      "ok",
			Duration:    duration,
			LastCheck:   time.Now(),
			Description: check.Description(),
		}

		if err != nil {
			result.Status = "error"
			result.Error = err.Error()
			allHealthy = false
			hc.readyChecks[name] = false
		} else {
			hc.readyChecks[name] = true
			readyChecks++
		}

		hc.status.Checks[name] = result

		// Обновляем метрики
		if hc.metrics != nil {
			hc.metrics.SetHealthCheckStatus(name, "component", err == nil)
			hc.metrics.ObserveHealthCheckDuration(name, "component", duration)
		}
	}

	// Определяем готовность системы
	hc.status.Ready = allHealthy && readyChecks > 0
	if hc.status.Ready {
		hc.status.Status = "ok"
		hc.status.ReadyReason = fmt.Sprintf("All %d checks passed", readyChecks)
	} else {
		hc.status.Status = "error"
		hc.status.ReadyReason = fmt.Sprintf("%d/%d checks failed", len(hc.checks)-readyChecks, len(hc.checks))
	}

	hc.ready = hc.status.Ready
	hc.readyReason = hc.status.ReadyReason

	return hc.status
}

// IsReady проверяет готовность системы
func (hc *EnhancedHealthChecker) IsReady() bool {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	return hc.ready
}

// GetStatus возвращает текущий статус
func (hc *EnhancedHealthChecker) GetStatus() HealthStatus {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	return hc.status
}

// GetConnectorID возвращает ID коннектора
func (hc *EnhancedHealthChecker) GetConnectorID() uuid.UUID {
	return hc.connectorID
}

// CountReadyChecks возвращает количество готовых checks
func (hc *EnhancedHealthChecker) CountReadyChecks() int {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	count := 0
	for _, ready := range hc.readyChecks {
		if ready {
			count++
		}
	}
	return count
}

// ServeHTTP реализует http.Handler для health check endpoints
func (hc *EnhancedHealthChecker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Обновляем метрики
	if hc.metrics != nil {
		hc.metrics.IncDiagnosticRequests("healthcheck", r.Method)
	}

	// Выполняем checks
	status := hc.RunChecks(ctx)

	// Определяем HTTP статус
	httpStatus := http.StatusOK
	if !status.Ready {
		httpStatus = http.StatusServiceUnavailable
	}

	// Устанавливаем заголовки
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)

	// Кодируем ответ
	if err := json.NewEncoder(w).Encode(status); err != nil {
		hc.logger.Error("Failed to encode health status", zap.Error(err))
		if hc.metrics != nil {
			hc.metrics.IncDiagnosticErrors("healthcheck", "encode_error")
		}
		http.Error(w, "Failed to encode health status", http.StatusInternalServerError)
		return
	}
}

// StartPeriodicChecks запускает периодические health checks
func (hc *EnhancedHealthChecker) StartPeriodicChecks(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	hc.logger.Info("Starting periodic health checks", zap.Duration("interval", interval))

	for {
		select {
		case <-ctx.Done():
			hc.logger.Info("Stopping periodic health checks")
			return
		case <-ticker.C:
			hc.RunChecks(ctx)
		}
	}
}

// DatabaseHealthCheck проверка базы данных
type DatabaseHealthCheck struct {
	name   string
	logger *zap.Logger
}

func (d *DatabaseHealthCheck) Check(ctx context.Context) error {
	// TODO: Реализовать проверку базы данных
	// Пока возвращаем nil (успех)
	return nil
}

func (d *DatabaseHealthCheck) Name() string {
	return d.name
}

func (d *DatabaseHealthCheck) Description() string {
	return "Database connectivity check"
}

// RedisHealthCheck проверка Redis
type RedisHealthCheck struct {
	name   string
	logger *zap.Logger
}

func (r *RedisHealthCheck) Check(ctx context.Context) error {
	// TODO: Реализовать проверку Redis
	// Пока возвращаем nil (успех)
	return nil
}

func (r *RedisHealthCheck) Name() string {
	return r.name
}

func (r *RedisHealthCheck) Description() string {
	return "Redis connectivity check"
}

// KeycloakHealthCheck проверка Keycloak
type KeycloakHealthCheck struct {
	name   string
	logger *zap.Logger
}

func (k *KeycloakHealthCheck) Check(ctx context.Context) error {
	// TODO: Реализовать проверку Keycloak
	// Пока возвращаем nil (успех)
	return nil
}

func (k *KeycloakHealthCheck) Name() string {
	return k.name
}

func (k *KeycloakHealthCheck) Description() string {
	return "Keycloak connectivity check"
}

// TCPProxyHealthCheck проверка TCP прокси
type TCPProxyHealthCheck struct {
	name   string
	logger *zap.Logger
}

func (t *TCPProxyHealthCheck) Check(ctx context.Context) error {
	// TODO: Реализовать проверку TCP прокси
	// Пока возвращаем nil (успех)
	return nil
}

func (t *TCPProxyHealthCheck) Name() string {
	return t.name
}

func (t *TCPProxyHealthCheck) Description() string {
	return "TCP Proxy health check"
}

// SystemHealthCheck проверка системы
type SystemHealthCheck struct {
	name   string
	logger *zap.Logger
}

func (s *SystemHealthCheck) Check(ctx context.Context) error {
	// TODO: Реализовать проверку системы (память, CPU, диск)
	// Пока возвращаем nil (успех)
	return nil
}

func (s *SystemHealthCheck) Name() string {
	return s.name
}

func (s *SystemHealthCheck) Description() string {
	return "System resources check"
}
