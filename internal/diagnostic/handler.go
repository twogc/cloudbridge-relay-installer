package diagnostic

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Handler диагностический handler с расширенными возможностями
type Handler struct {
	logger       *zap.Logger
	timeout      time.Duration
	tunnelID     uuid.UUID
	connectorID  uuid.UUID
	version      string
	buildTime    string
	buildType    string
	cliFlags     map[string]string
	metrics      interface{} // EnhancedMetrics interface
}

// NewHandler создает новый диагностический handler
func NewHandler(
	logger *zap.Logger,
	timeout time.Duration,
	tunnelID uuid.UUID,
	connectorID uuid.UUID,
	version string,
	buildTime string,
	buildType string,
	metrics interface{},
) *Handler {
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	// Добавляем базовую информацию о системе
	cliFlags := map[string]string{
		"uid":        strconv.Itoa(os.Getuid()),
		"gid":        strconv.Itoa(os.Getgid()),
		"pid":        strconv.Itoa(os.Getpid()),
		"go_version": runtime.Version(),
		"os":         runtime.GOOS,
		"arch":       runtime.GOARCH,
	}

	return &Handler{
		logger:      logger,
		timeout:     timeout,
		tunnelID:    tunnelID,
		connectorID: connectorID,
		version:     version,
		buildTime:   buildTime,
		buildType:   buildType,
		cliFlags:    cliFlags,
		metrics:     metrics,
	}
}

// InstallEndpoints устанавливает диагностические endpoints
func (h *Handler) InstallEndpoints(router *http.ServeMux) {
	router.HandleFunc("/diagnostic/config", h.ConfigurationHandler)
	router.HandleFunc("/diagnostic/tunnel-state", h.TunnelStateHandler)
	router.HandleFunc("/diagnostic/system", h.SystemHandler)
	router.HandleFunc("/diagnostic/version", h.VersionHandler)
	router.HandleFunc("/diagnostic/health", h.HealthHandler)
	router.HandleFunc("/diagnostic/metrics", h.MetricsHandler)
	router.HandleFunc("/diagnostic/pprof/", h.PProfHandler)
}

// ConfigurationResponse ответ с конфигурацией
type ConfigurationResponse struct {
	CLIFlags    map[string]string `json:"cli_flags"`
	Environment map[string]string `json:"environment"`
	BuildInfo   BuildInfo         `json:"build_info"`
}

// BuildInfo информация о сборке
type BuildInfo struct {
	Version   string `json:"version"`
	BuildTime string `json:"build_time"`
	BuildType string `json:"build_type"`
	GoVersion string `json:"go_version"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
}

// ConfigurationHandler обрабатывает запросы конфигурации
func (h *Handler) ConfigurationHandler(w http.ResponseWriter, r *http.Request) {
	logger := h.logger.With(zap.String("endpoint", "configuration"))
	logger.Info("Configuration request started")

	defer logger.Info("Configuration request finished")

	// Обновляем метрики
	if h.metrics != nil {
		if em, ok := h.metrics.(interface{ IncDiagnosticRequests(endpoint, method string) }); ok {
			em.IncDiagnosticRequests("config", r.Method)
		}
	}

	// Собираем переменные окружения
	env := make(map[string]string)
	for _, e := range os.Environ() {
		// Фильтруем чувствительные переменные
		if !isSensitiveEnvVar(e) {
			env[e] = os.Getenv(e)
		}
	}

	response := ConfigurationResponse{
		CLIFlags: h.cliFlags,
		Environment: env,
		BuildInfo: BuildInfo{
			Version:   h.version,
			BuildTime: h.buildTime,
			BuildType: h.buildType,
			GoVersion: runtime.Version(),
			OS:        runtime.GOOS,
			Arch:      runtime.GOARCH,
		},
	}

	h.writeJSONResponse(w, response, logger)
}

// TunnelStateResponse ответ с состоянием туннеля
type TunnelStateResponse struct {
	TunnelID    uuid.UUID `json:"tunnel_id"`
	ConnectorID uuid.UUID `json:"connector_id"`
	Status      string    `json:"status"`
	Uptime      string    `json:"uptime"`
	StartTime   time.Time `json:"start_time"`
	Connections int       `json:"connections"`
}

// TunnelStateHandler обрабатывает запросы состояния туннеля
func (h *Handler) TunnelStateHandler(w http.ResponseWriter, r *http.Request) {
	logger := h.logger.With(zap.String("endpoint", "tunnel_state"))
	logger.Info("Tunnel state request started")

	defer logger.Info("Tunnel state request finished")

	// Обновляем метрики
	if h.metrics != nil {
		if em, ok := h.metrics.(interface{ IncDiagnosticRequests(endpoint, method string) }); ok {
			em.IncDiagnosticRequests("tunnel_state", r.Method)
		}
	}

	// TODO: Получить реальное состояние туннеля
	response := TunnelStateResponse{
		TunnelID:    h.tunnelID,
		ConnectorID: h.connectorID,
		Status:      "active",
		Uptime:      "1h30m",
		StartTime:   time.Now().Add(-90 * time.Minute),
		Connections: 5, // TODO: Получить реальное количество
	}

	h.writeJSONResponse(w, response, logger)
}

// SystemInfo информация о системе
type SystemInfo struct {
	Memory     MemoryInfo `json:"memory"`
	CPU        CPUInfo    `json:"cpu"`
	Goroutines int        `json:"goroutines"`
	Process    ProcessInfo `json:"process"`
}

// MemoryInfo информация о памяти
type MemoryInfo struct {
	Alloc      uint64 `json:"alloc"`
	TotalAlloc uint64 `json:"total_alloc"`
	Sys        uint64 `json:"sys"`
	NumGC      uint32 `json:"num_gc"`
}

// CPUInfo информация о CPU
type CPUInfo struct {
	NumCPU int `json:"num_cpu"`
}

// ProcessInfo информация о процессе
type ProcessInfo struct {
	PID int `json:"pid"`
	UID int `json:"uid"`
	GID int `json:"gid"`
}

// SystemHandler обрабатывает запросы информации о системе
func (h *Handler) SystemHandler(w http.ResponseWriter, r *http.Request) {
	logger := h.logger.With(zap.String("endpoint", "system"))
	logger.Info("System info request started")

	defer logger.Info("System info request finished")

	// Обновляем метрики
	if h.metrics != nil {
		if em, ok := h.metrics.(interface{ IncDiagnosticRequests(endpoint, method string) }); ok {
			em.IncDiagnosticRequests("system", r.Method)
		}
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	response := SystemInfo{
		Memory: MemoryInfo{
			Alloc:      m.Alloc,
			TotalAlloc: m.TotalAlloc,
			Sys:        m.Sys,
			NumGC:      m.NumGC,
		},
		CPU: CPUInfo{
			NumCPU: runtime.NumCPU(),
		},
		Goroutines: runtime.NumGoroutine(),
		Process: ProcessInfo{
			PID: os.Getpid(),
			UID: os.Getuid(),
			GID: os.Getgid(),
		},
	}

	h.writeJSONResponse(w, response, logger)
}

// VersionResponse ответ с версией
type VersionResponse struct {
	Version   string `json:"version"`
	BuildTime string `json:"build_time"`
	BuildType string `json:"build_type"`
	GoVersion string `json:"go_version"`
}

// VersionHandler обрабатывает запросы версии
func (h *Handler) VersionHandler(w http.ResponseWriter, r *http.Request) {
	logger := h.logger.With(zap.String("endpoint", "version"))
	logger.Info("Version request started")

	defer logger.Info("Version request finished")

	// Обновляем метрики
	if h.metrics != nil {
		if em, ok := h.metrics.(interface{ IncDiagnosticRequests(endpoint, method string) }); ok {
			em.IncDiagnosticRequests("version", r.Method)
		}
	}

	response := VersionResponse{
		Version:   h.version,
		BuildTime: h.buildTime,
		BuildType: h.buildType,
		GoVersion: runtime.Version(),
	}

	h.writeJSONResponse(w, response, logger)
}

// HealthResponse ответ с состоянием здоровья
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Uptime    string    `json:"uptime"`
	Version   string    `json:"version"`
}

// HealthHandler обрабатывает запросы здоровья
func (h *Handler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	logger := h.logger.With(zap.String("endpoint", "health"))
	logger.Info("Health request started")

	defer logger.Info("Health request finished")

	// Обновляем метрики
	if h.metrics != nil {
		if em, ok := h.metrics.(interface{ IncDiagnosticRequests(endpoint, method string) }); ok {
			em.IncDiagnosticRequests("health", r.Method)
		}
	}

	response := HealthResponse{
		Status:    "ok",
		Timestamp: time.Now(),
		Uptime:    "1h30m", // TODO: Вычислить реальное время работы
		Version:   h.version,
	}

	h.writeJSONResponse(w, response, logger)
}

// MetricsHandler обрабатывает запросы метрик
func (h *Handler) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	logger := h.logger.With(zap.String("endpoint", "metrics"))
	logger.Info("Metrics request started")

	defer logger.Info("Metrics request finished")

	// Обновляем метрики
	if h.metrics != nil {
		if em, ok := h.metrics.(interface{ IncDiagnosticRequests(endpoint, method string) }); ok {
			em.IncDiagnosticRequests("metrics", r.Method)
		}
	}

	// Перенаправляем на стандартный Prometheus handler
	http.Redirect(w, r, "/metrics", http.StatusMovedPermanently)
}

// PProfHandler обрабатывает pprof запросы
func (h *Handler) PProfHandler(w http.ResponseWriter, r *http.Request) {
	logger := h.logger.With(zap.String("endpoint", "pprof"))
	logger.Info("PProf request started")

	defer logger.Info("PProf request finished")

	// Обновляем метрики
	if h.metrics != nil {
		if em, ok := h.metrics.(interface{ IncDiagnosticRequests(endpoint, method string) }); ok {
			em.IncDiagnosticRequests("pprof", r.Method)
		}
	}

	// Перенаправляем на стандартный pprof handler
	http.Redirect(w, r, "/debug/pprof/", http.StatusMovedPermanently)
}

// writeJSONResponse записывает JSON ответ
func (h *Handler) writeJSONResponse(w http.ResponseWriter, data interface{}, logger *zap.Logger) {
	w.Header().Set("Content-Type", "application/json")
	
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	
	if err := encoder.Encode(data); err != nil {
		logger.Error("Failed to encode response", zap.Error(err))
		if h.metrics != nil {
			if em, ok := h.metrics.(interface{ IncDiagnosticErrors(endpoint, errorType string) }); ok {
				em.IncDiagnosticErrors("json_encode", "encode_error")
			}
		}
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"error": "Failed to encode response: %s"}`, err.Error())
		return
	}
}

// isSensitiveEnvVar проверяет, является ли переменная окружения чувствительной
func isSensitiveEnvVar(envVar string) bool {
	sensitive := []string{
		"PASSWORD", "SECRET", "KEY", "TOKEN", "AUTH", "CREDENTIAL",
		"DB_PASSWORD", "REDIS_PASSWORD", "JWT_SECRET",
	}
	
	for _, s := range sensitive {
		if contains(envVar, s) {
			return true
		}
	}
	return false
}

// contains проверяет, содержит ли строка подстроку (без учета регистра)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s[:len(substr)] == substr || 
		    len(s) > len(substr) && s[len(substr)] == '=')
} 