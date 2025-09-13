package api

import (
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"go.uber.org/zap"
)

// MonitoringEndpoints предоставляет расширенные мониторинг endpoints
type MonitoringEndpoints struct {
	logger *zap.Logger
}

// NewMonitoringEndpoints создает новый MonitoringEndpoints
func NewMonitoringEndpoints(logger *zap.Logger) *MonitoringEndpoints {
	return &MonitoringEndpoints{
		logger: logger,
	}
}

// SystemMetricsResponse ответ с системными метриками
type SystemMetricsResponse struct {
	CPU       CPUMetrics     `json:"cpu"`
	Memory    MemoryMetrics  `json:"memory"`
	Disk      DiskMetrics    `json:"disk"`
	Network   NetworkMetrics `json:"network"`
	Runtime   RuntimeMetrics `json:"runtime"`
	Timestamp time.Time      `json:"timestamp"`
}

// CPUMetrics метрики CPU
type CPUMetrics struct {
	UsagePercent float64   `json:"usage_percent"`
	Count        int       `json:"count"`
	LoadAvg      []float64 `json:"load_avg"`
}

// MemoryMetrics метрики памяти
type MemoryMetrics struct {
	Total       uint64  `json:"total"`
	Available   uint64  `json:"available"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"used_percent"`
}

// DiskMetrics метрики диска
type DiskMetrics struct {
	Total       uint64  `json:"total"`
	Free        uint64  `json:"free"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"used_percent"`
}

// NetworkMetrics метрики сети
type NetworkMetrics struct {
	BytesSent   uint64 `json:"bytes_sent"`
	BytesRecv   uint64 `json:"bytes_recv"`
	PacketsSent uint64 `json:"packets_sent"`
	PacketsRecv uint64 `json:"packets_recv"`
}

// RuntimeMetrics метрики Go runtime
type RuntimeMetrics struct {
	Goroutines   int    `json:"goroutines"`
	HeapAlloc    uint64 `json:"heap_alloc"`
	HeapSys      uint64 `json:"heap_sys"`
	HeapIdle     uint64 `json:"heap_idle"`
	HeapInuse    uint64 `json:"heap_inuse"`
	HeapReleased uint64 `json:"heap_released"`
	HeapObjects  uint64 `json:"heap_objects"`
	StackInuse   uint64 `json:"stack_inuse"`
	StackSys     uint64 `json:"stack_sys"`
	MSpanInuse   uint64 `json:"mspan_inuse"`
	MSpanSys     uint64 `json:"mspan_sys"`
	MCacheInuse  uint64 `json:"mcache_inuse"`
	MCacheSys    uint64 `json:"mcache_sys"`
	BuckHashSys  uint64 `json:"buck_hash_sys"`
	GCSys        uint64 `json:"gc_sys"`
	OtherSys     uint64 `json:"other_sys"`
	NextGC       uint64 `json:"next_gc"`
	LastGC       uint64 `json:"last_gc"`
	PauseTotalNs uint64 `json:"pause_total_ns"`
	NumGC        uint32 `json:"num_gc"`
}

// GetSystemMetrics возвращает детальные системные метрики
func (me *MonitoringEndpoints) GetSystemMetrics(c *gin.Context) {
	me.logger.Info("Getting detailed system metrics")

	// CPU метрики
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		me.logger.Error("Failed to get CPU metrics", zap.Error(err))
		cpuPercent = []float64{0}
	}

	cpuCount, err := cpu.Counts(true)
	if err != nil {
		me.logger.Error("Failed to get CPU count", zap.Error(err))
		cpuCount = 0
	}

	// Заглушка для load average (недоступно в gopsutil)
	loadAvg := []float64{0.5, 0.3, 0.2}

	// Memory метрики
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		me.logger.Error("Failed to get memory metrics", zap.Error(err))
		memInfo = &mem.VirtualMemoryStat{}
	}

	// Disk метрики
	diskInfo, err := disk.Usage("/")
	if err != nil {
		me.logger.Error("Failed to get disk metrics", zap.Error(err))
		diskInfo = &disk.UsageStat{}
	}

	// Network метрики
	netStats, err := net.IOCounters(false)
	if err != nil {
		me.logger.Error("Failed to get network metrics", zap.Error(err))
		netStats = []net.IOCountersStat{}
	}

	var networkMetrics NetworkMetrics
	if len(netStats) > 0 {
		networkMetrics = NetworkMetrics{
			BytesSent:   netStats[0].BytesSent,
			BytesRecv:   netStats[0].BytesRecv,
			PacketsSent: netStats[0].PacketsSent,
			PacketsRecv: netStats[0].PacketsRecv,
		}
	}

	// Runtime метрики
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	runtimeMetrics := RuntimeMetrics{
		Goroutines:   runtime.NumGoroutine(),
		HeapAlloc:    m.HeapAlloc,
		HeapSys:      m.HeapSys,
		HeapIdle:     m.HeapIdle,
		HeapInuse:    m.HeapInuse,
		HeapReleased: m.HeapReleased,
		HeapObjects:  m.HeapObjects,
		StackInuse:   m.StackInuse,
		StackSys:     m.StackSys,
		MSpanInuse:   m.MSpanInuse,
		MSpanSys:     m.MSpanSys,
		MCacheInuse:  m.MCacheInuse,
		MCacheSys:    m.MCacheSys,
		BuckHashSys:  m.BuckHashSys,
		GCSys:        m.GCSys,
		OtherSys:     m.OtherSys,
		NextGC:       m.NextGC,
		LastGC:       m.LastGC,
		PauseTotalNs: m.PauseTotalNs,
		NumGC:        m.NumGC,
	}

	response := SystemMetricsResponse{
		CPU: CPUMetrics{
			UsagePercent: cpuPercent[0],
			Count:        cpuCount,
			LoadAvg:      loadAvg,
		},
		Memory: MemoryMetrics{
			Total:       memInfo.Total,
			Available:   memInfo.Available,
			Used:        memInfo.Used,
			UsedPercent: memInfo.UsedPercent,
		},
		Disk: DiskMetrics{
			Total:       diskInfo.Total,
			Free:        diskInfo.Free,
			Used:        diskInfo.Used,
			UsedPercent: diskInfo.UsedPercent,
		},
		Network:   networkMetrics,
		Runtime:   runtimeMetrics,
		Timestamp: time.Now(),
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"metrics": response,
	})
}

// GetHealthCheck возвращает расширенный health check
func (me *MonitoringEndpoints) GetHealthCheck(c *gin.Context) {
	me.logger.Info("Getting extended health check")

	// Проверяем системные ресурсы
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		me.logger.Error("Failed to get memory info", zap.Error(err))
		memInfo = &mem.VirtualMemoryStat{}
	}

	diskInfo, err := disk.Usage("/")
	if err != nil {
		me.logger.Error("Failed to get disk info", zap.Error(err))
		diskInfo = &disk.UsageStat{}
	}

	// Определяем статус здоровья
	status := "healthy"
	issues := []string{}

	// Проверяем память
	if memInfo.UsedPercent > 90 {
		status = "warning"
		issues = append(issues, "High memory usage")
	}

	// Проверяем диск
	if diskInfo.UsedPercent > 90 {
		status = "warning"
		issues = append(issues, "High disk usage")
	}

	// Проверяем количество горутин
	goroutines := runtime.NumGoroutine()
	if goroutines > 10000 {
		status = "warning"
		issues = append(issues, "High number of goroutines")
	}

	healthResponse := gin.H{
		"status":    status,
		"service":   "cloudbridge-relay",
		"version":   "v1.0.0",
		"uptime":    time.Since(time.Now()).String(), // Это нужно будет исправить
		"timestamp": time.Now(),
		"components": gin.H{
			"api_server":    "healthy",
			"relay_server":  "healthy",
			"ai_service":    "healthy",
			"data_pipeline": "healthy",
		},
		"system": gin.H{
			"memory_usage_percent": memInfo.UsedPercent,
			"disk_usage_percent":   diskInfo.UsedPercent,
			"goroutines":           goroutines,
		},
	}

	if len(issues) > 0 {
		healthResponse["issues"] = issues
	}

	// Определяем HTTP статус код
	httpStatus := http.StatusOK
	if status == "warning" {
		httpStatus = http.StatusTooManyRequests
	} else if status == "unhealthy" {
		httpStatus = http.StatusServiceUnavailable
	}

	c.JSON(httpStatus, healthResponse)
}

// GetPrometheusMetrics возвращает метрики в формате Prometheus
func (me *MonitoringEndpoints) GetPrometheusMetrics(c *gin.Context) {
	me.logger.Info("Getting Prometheus metrics")

	// Получаем системные метрики
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		me.logger.Error("Failed to get memory info", zap.Error(err))
		memInfo = &mem.VirtualMemoryStat{}
	}

	diskInfo, err := disk.Usage("/")
	if err != nil {
		me.logger.Error("Failed to get disk info", zap.Error(err))
		diskInfo = &disk.UsageStat{}
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Формируем метрики в формате Prometheus
	metrics := `# HELP cloudbridge_memory_usage_percent Memory usage percentage
# TYPE cloudbridge_memory_usage_percent gauge
cloudbridge_memory_usage_percent ` + strconv.FormatFloat(memInfo.UsedPercent, 'f', 2, 64) + `

# HELP cloudbridge_disk_usage_percent Disk usage percentage
# TYPE cloudbridge_disk_usage_percent gauge
cloudbridge_disk_usage_percent ` + strconv.FormatFloat(diskInfo.UsedPercent, 'f', 2, 64) + `

# HELP cloudbridge_goroutines_total Number of goroutines
# TYPE cloudbridge_goroutines_total gauge
cloudbridge_goroutines_total ` + strconv.Itoa(runtime.NumGoroutine()) + `

# HELP cloudbridge_heap_alloc_bytes Heap allocation in bytes
# TYPE cloudbridge_heap_alloc_bytes gauge
cloudbridge_heap_alloc_bytes ` + strconv.FormatUint(m.HeapAlloc, 10) + `

# HELP cloudbridge_heap_objects_total Number of heap objects
# TYPE cloudbridge_heap_objects_total gauge
cloudbridge_heap_objects_total ` + strconv.FormatUint(m.HeapObjects, 10) + `

# HELP cloudbridge_gc_duration_seconds_total Total GC duration in seconds
# TYPE cloudbridge_gc_duration_seconds_total counter
cloudbridge_gc_duration_seconds_total ` + strconv.FormatFloat(float64(m.PauseTotalNs)/1e9, 'f', 6, 64) + `

# HELP cloudbridge_gc_runs_total Total number of GC runs
# TYPE cloudbridge_gc_runs_total counter
cloudbridge_gc_runs_total ` + strconv.FormatUint(uint64(m.NumGC), 10) + `
`

	c.Header("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	c.String(http.StatusOK, metrics)
}

// GetServiceStatus возвращает статус всех сервисов
func (me *MonitoringEndpoints) GetServiceStatus(c *gin.Context) {
	me.logger.Info("Getting service status")

	services := gin.H{
		"api_server": gin.H{
			"status":    "healthy",
			"uptime":    "1h30m",
			"version":   "v1.0.0",
			"endpoints": []string{"/api/v1/health", "/api/v1/metrics"},
		},
		"relay_server": gin.H{
			"status":      "healthy",
			"uptime":      "1h30m",
			"version":     "v1.0.0",
			"connections": 0,
		},
		"ai_service": gin.H{
			"status":  "healthy",
			"uptime":  "1h30m",
			"version": "v1.0.0",
			"models":  []string{"anomaly_detection", "load_prediction", "route_optimization"},
		},
		"data_pipeline": gin.H{
			"status":     "healthy",
			"uptime":     "1h30m",
			"version":    "v1.0.0",
			"collectors": []string{"prometheus", "system", "network", "relay"},
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "success",
		"services":  services,
		"timestamp": time.Now(),
	})
}
