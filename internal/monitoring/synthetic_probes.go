package monitoring

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

// SyntheticProbeManager управляет synthetic probes для SLI измерений
type SyntheticProbeManager struct {
	logger *zap.Logger
	config *SyntheticProbeConfig

	// Probes
	probes    map[string]*SyntheticProbe
	probesMux sync.RWMutex

	// Results
	results    map[string]*ProbeResult
	resultsMux sync.RWMutex

	// Metrics
	metrics *ProbeMetrics

	// Concurrency control
	sem       chan struct{}
	parentCtx context.Context
}

// SyntheticProbeConfig конфигурация synthetic probes
type SyntheticProbeConfig struct {
	// Общие настройки
	Enabled       bool          `json:"enabled"`
	ProbeInterval time.Duration `json:"probe_interval"` // Интервал между probes
	ProbeTimeout  time.Duration `json:"probe_timeout"`  // Timeout для probe
	MaxConcurrent int           `json:"max_concurrent"` // Максимум concurrent probes

	// SLI настройки
	SLIEnabled    bool          `json:"sli_enabled"`
	SLIThresholds SLIThresholds `json:"sli_thresholds"`

	// Retry настройки
	RetryEnabled bool          `json:"retry_enabled"`
	MaxRetries   int           `json:"max_retries"`
	RetryDelay   time.Duration `json:"retry_delay"`

	// Alerting
	AlertingEnabled bool            `json:"alerting_enabled"`
	AlertThresholds AlertThresholds `json:"alert_thresholds"`
}

// DefaultSyntheticProbeConfig возвращает конфигурацию по умолчанию
func DefaultSyntheticProbeConfig() *SyntheticProbeConfig {
	return &SyntheticProbeConfig{
		Enabled:       true,
		ProbeInterval: 30 * time.Second,
		ProbeTimeout:  10 * time.Second,
		MaxConcurrent: 10,

		SLIEnabled: true,
		SLIThresholds: SLIThresholds{
			LatencyP99:   100.0,  // 100ms
			ErrorRate:    1.0,    // 1%
			Availability: 99.9,   // 99.9%
			Throughput:   1000.0, // 1000 req/s
		},

		RetryEnabled: true,
		MaxRetries:   3,
		RetryDelay:   1 * time.Second,

		AlertingEnabled: true,
		AlertThresholds: AlertThresholds{
			LatencyAlert:      200.0, // 200ms
			ErrorRateAlert:    5.0,   // 5%
			AvailabilityAlert: 95.0,  // 95%
		},
	}
}

// SLIThresholds представляет пороги для SLI
type SLIThresholds struct {
	LatencyP99   float64 `json:"latency_p99"`  // ms
	ErrorRate    float64 `json:"error_rate"`   // %
	Availability float64 `json:"availability"` // %
	Throughput   float64 `json:"throughput"`   // req/s
}

// AlertThresholds представляет пороги для алертов
type AlertThresholds struct {
	LatencyAlert      float64 `json:"latency_alert"`      // ms
	ErrorRateAlert    float64 `json:"error_rate_alert"`   // %
	AvailabilityAlert float64 `json:"availability_alert"` // %
}

// SyntheticProbe представляет synthetic probe
type SyntheticProbe struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Type         string            `json:"type"`   // http, tcp, udp, icmp
	Target       string            `json:"target"` // URL или адрес
	Method       string            `json:"method"` // HTTP метод
	Headers      map[string]string `json:"headers"`
	Body         string            `json:"body"`
	ExpectedCode int               `json:"expected_code"`
	Timeout      time.Duration     `json:"timeout"`
	Interval     time.Duration     `json:"interval"`
	Enabled      bool              `json:"enabled"`
	CreatedAt    time.Time         `json:"created_at"`
	LastRun      time.Time         `json:"last_run"`

	// SLI настройки
	SLIEnabled bool    `json:"sli_enabled"`
	SLIWeight  float64 `json:"sli_weight"` // Вес в общем SLI
}

// ProbeResult представляет результат probe
type ProbeResult struct {
	ProbeID      string    `json:"probe_id"`
	Timestamp    time.Time `json:"timestamp"`
	Success      bool      `json:"success"`
	Latency      float64   `json:"latency"` // ms
	StatusCode   int       `json:"status_code"`
	Error        string    `json:"error"`
	ResponseSize int64     `json:"response_size"`

	// SLI метрики
	SLIMetrics *SLIMetrics `json:"sli_metrics"`
}

// SLIMetrics представляет SLI метрики
type SLIMetrics struct {
	LatencyP99   float64 `json:"latency_p99"`  // ms
	ErrorRate    float64 `json:"error_rate"`   // %
	Availability float64 `json:"availability"` // %
	Throughput   float64 `json:"throughput"`   // req/s
	SLIScore     float64 `json:"sli_score"`    // 0.0-1.0
}

// ProbeMetrics представляет метрики probes
type ProbeMetrics struct {
	mu               sync.Mutex
	TotalProbes      int64   `json:"total_probes"`
	SuccessfulProbes int64   `json:"successful_probes"`
	FailedProbes     int64   `json:"failed_probes"`
	AverageLatency   float64 `json:"average_latency"`
	ErrorRate        float64 `json:"error_rate"`
	Availability     float64 `json:"availability"`
	SLIScore         float64 `json:"sli_score"`
}

// NewSyntheticProbeManager создает новый менеджер synthetic probes
func NewSyntheticProbeManager(logger *zap.Logger, config *SyntheticProbeConfig) *SyntheticProbeManager {
	if config == nil {
		config = DefaultSyntheticProbeConfig()
	}

	max := config.MaxConcurrent
	if max <= 0 {
		max = 1
	}

	return &SyntheticProbeManager{
		logger:  logger,
		config:  config,
		probes:  make(map[string]*SyntheticProbe),
		results: make(map[string]*ProbeResult),
		metrics: &ProbeMetrics{},
		sem:     make(chan struct{}, max),
	}
}

// AddProbe добавляет новый synthetic probe
func (spm *SyntheticProbeManager) AddProbe(probe *SyntheticProbe) error {
	spm.probesMux.Lock()
	defer spm.probesMux.Unlock()

	// Валидация probe
	if err := spm.validateProbe(probe); err != nil {
		return fmt.Errorf("invalid probe: %w", err)
	}

	// Устанавливаем значения по умолчанию
	if probe.Timeout == 0 {
		probe.Timeout = spm.config.ProbeTimeout
	}
	if probe.Interval == 0 {
		probe.Interval = spm.config.ProbeInterval
	}
	if probe.CreatedAt.IsZero() {
		probe.CreatedAt = time.Now()
	}

	spm.probes[probe.ID] = probe

	spm.logger.Info("Synthetic probe added",
		zap.String("probe_id", probe.ID),
		zap.String("name", probe.Name),
		zap.String("type", probe.Type),
		zap.String("target", probe.Target))

	return nil
}

// validateProbe валидирует probe
func (spm *SyntheticProbeManager) validateProbe(probe *SyntheticProbe) error {
	if probe.ID == "" {
		return fmt.Errorf("probe ID is required")
	}
	if probe.Name == "" {
		return fmt.Errorf("probe name is required")
	}
	if probe.Type == "" {
		return fmt.Errorf("probe type is required")
	}
	if probe.Target == "" {
		return fmt.Errorf("probe target is required")
	}

	// Валидация типа probe
	switch probe.Type {
	case "http", "https":
		if probe.Method == "" {
			probe.Method = "GET"
		}
	case "tcp", "udp":
		// Проверяем, что target - это адрес:порт
		if _, _, err := net.SplitHostPort(probe.Target); err != nil {
			return fmt.Errorf("invalid target format for %s probe: %w", probe.Type, err)
		}
	case "icmp":
		// Проверяем, что target - это IP адрес
		if net.ParseIP(probe.Target) == nil {
			return fmt.Errorf("invalid IP address for ICMP probe")
		}
	default:
		return fmt.Errorf("unsupported probe type: %s", probe.Type)
	}

	return nil
}

// Start запускает synthetic probes
func (spm *SyntheticProbeManager) Start(ctx context.Context) error {
	if !spm.config.Enabled {
		spm.logger.Info("Synthetic probes are disabled")
		return nil
	}

	spm.parentCtx = ctx

	spm.logger.Info("Starting synthetic probes",
		zap.Int("probe_count", len(spm.probes)),
		zap.Duration("interval", spm.config.ProbeInterval),
		zap.Int("max_concurrent", spm.config.MaxConcurrent))

	// Запускаем probes в отдельных горутинах
	for _, probe := range spm.probes {
		if probe.Enabled {
			go spm.runProbe(ctx, probe)
		}
	}

	return nil
}

// runProbe запускает отдельный probe
func (spm *SyntheticProbeManager) runProbe(ctx context.Context, probe *SyntheticProbe) {
	ticker := time.NewTicker(probe.Interval)
	defer ticker.Stop()

	spm.logger.Info("Starting probe",
		zap.String("probe_id", probe.ID),
		zap.String("name", probe.Name),
		zap.Duration("interval", probe.Interval))

	for {
		select {
		case <-ctx.Done():
			spm.logger.Info("Probe stopped due to context cancellation",
				zap.String("probe_id", probe.ID))
			return
		case <-ticker.C:
			spm.executeProbe(probe)
		}
	}
}

// executeProbe выполняет probe
func (spm *SyntheticProbeManager) executeProbe(probe *SyntheticProbe) {
	// Проверяем лимит параллелизма
	select {
	case spm.sem <- struct{}{}:
		defer func() { <-spm.sem }()
	default:
		spm.logger.Debug("Skipping probe due to MaxConcurrent limit",
			zap.String("probe_id", probe.ID))
		return
	}

	start := time.Now()

	// Выполняем probe с retry (только если включено)
	var result *ProbeResult
	var err error

	attempts := 1
	if spm.config.RetryEnabled && spm.config.MaxRetries > 0 {
		attempts = spm.config.MaxRetries + 1
	}

	for attempt := 1; attempt <= attempts; attempt++ {
		result, err = spm.runProbeOnce(spm.parentCtx, probe)
		if err == nil && result.Success {
			break
		}

		if attempt < attempts {
			spm.logger.Debug("Probe failed, retrying",
				zap.String("probe_id", probe.ID),
				zap.Int("attempt", attempt),
				zap.Error(err))
			time.Sleep(spm.config.RetryDelay)
		}
	}

	// Обновляем время последнего запуска
	probe.LastRun = time.Now()

	// Сохраняем результат
	spm.resultsMux.Lock()
	spm.results[probe.ID] = result
	spm.resultsMux.Unlock()

	// Обновляем метрики
	spm.updateMetrics(result)

	// Проверяем SLI
	if spm.config.SLIEnabled && probe.SLIEnabled {
		spm.checkSLI(probe, result)
	}

	// Проверяем алерты
	if spm.config.AlertingEnabled {
		spm.checkAlerts(probe, result)
	}

	spm.logger.Debug("Probe executed",
		zap.String("probe_id", probe.ID),
		zap.Bool("success", result.Success),
		zap.Float64("latency", result.Latency),
		zap.Duration("duration", time.Since(start)))
}

// runProbeOnce выполняет probe один раз
func (spm *SyntheticProbeManager) runProbeOnce(parentCtx context.Context, probe *SyntheticProbe) (*ProbeResult, error) {
	result := &ProbeResult{
		ProbeID:   probe.ID,
		Timestamp: time.Now(),
	}

	ctx, cancel := context.WithTimeout(parentCtx, probe.Timeout)
	defer cancel()

	start := time.Now()

	switch probe.Type {
	case "http", "https":
		err := spm.executeHTTPProbe(ctx, probe, result)
		if err != nil {
			result.Error = err.Error()
			result.Success = false
		}
	case "tcp":
		err := spm.executeTCPProbe(ctx, probe, result)
		if err != nil {
			result.Error = err.Error()
			result.Success = false
		}
	case "udp":
		err := spm.executeUDPProbe(ctx, probe, result)
		if err != nil {
			result.Error = err.Error()
			result.Success = false
		}
	case "icmp":
		err := spm.executeICMPProbe(ctx, probe, result)
		if err != nil {
			result.Error = err.Error()
			result.Success = false
		}
	default:
		return nil, fmt.Errorf("unsupported probe type: %s", probe.Type)
	}

	result.Latency = float64(time.Since(start).Nanoseconds()) / 1e6 // Convert to ms

	return result, nil
}

// executeHTTPProbe выполняет HTTP probe
func (spm *SyntheticProbeManager) executeHTTPProbe(ctx context.Context, probe *SyntheticProbe, result *ProbeResult) error {
	// Используем общий Transport с keep-alive
	tr := &http.Transport{
		Proxy:              http.ProxyFromEnvironment,
		MaxIdleConns:       100,
		IdleConnTimeout:    90 * time.Second,
		DisableCompression: false,
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   probe.Timeout,
	}

	req, err := http.NewRequestWithContext(ctx, probe.Method, probe.Target, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Добавляем заголовки
	for key, value := range probe.Headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode

	// Читаем тело для получения реального размера и освобождения соединения
	n, _ := io.Copy(io.Discard, resp.Body)
	result.ResponseSize = n

	// Проверяем ожидаемый код ответа
	if probe.ExpectedCode != 0 && resp.StatusCode != probe.ExpectedCode {
		return fmt.Errorf("unexpected status code: got %d, expected %d", resp.StatusCode, probe.ExpectedCode)
	}

	// Проверяем успешность
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		result.Success = true
	} else {
		result.Success = false
		result.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
	}

	return nil
}

// executeTCPProbe выполняет TCP probe
func (spm *SyntheticProbeManager) executeTCPProbe(ctx context.Context, probe *SyntheticProbe, result *ProbeResult) error {
	conn, err := net.DialTimeout("tcp", probe.Target, probe.Timeout)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	result.Success = true
	return nil
}

// executeUDPProbe выполняет UDP probe
func (spm *SyntheticProbeManager) executeUDPProbe(ctx context.Context, probe *SyntheticProbe, result *ProbeResult) error {
	d := &net.Dialer{}
	conn, err := d.DialContext(ctx, "udp", probe.Target)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	// Устанавливаем deadline для записи
	if err := conn.SetWriteDeadline(time.Now().Add(probe.Timeout)); err != nil {
		return fmt.Errorf("failed to set write deadline: %w", err)
	}

	// Отправляем тестовые данные
	_, err = conn.Write([]byte("ping"))
	if err != nil {
		return fmt.Errorf("failed to send data: %w", err)
	}

	// UDP probe считается успешным если пакет отправлен
	// (не ждем ответ, так как это может быть "black hole")
	result.Success = true
	return nil
}

// executeICMPProbe выполняет TCP ping probe (не настоящий ICMP)
func (spm *SyntheticProbeManager) executeICMPProbe(ctx context.Context, probe *SyntheticProbe, result *ProbeResult) error {
	// TCP ping через connect к порту 80 (не настоящий ICMP)
	d := &net.Dialer{}
	conn, err := d.DialContext(ctx, "tcp", probe.Target+":80")
	if err != nil {
		return fmt.Errorf("failed to tcp ping: %w", err)
	}
	defer conn.Close()

	result.Success = true
	return nil
}

// updateMetrics обновляет метрики
func (spm *SyntheticProbeManager) updateMetrics(result *ProbeResult) {
	spm.metrics.mu.Lock()
	defer spm.metrics.mu.Unlock()

	spm.metrics.TotalProbes++

	if result.Success {
		spm.metrics.SuccessfulProbes++
	} else {
		spm.metrics.FailedProbes++
	}

	// Обновляем среднюю латентность
	total := float64(spm.metrics.TotalProbes)
	spm.metrics.AverageLatency = (spm.metrics.AverageLatency*(total-1) + result.Latency) / total

	// Обновляем error rate
	spm.metrics.ErrorRate = float64(spm.metrics.FailedProbes) / total * 100

	// Обновляем availability
	spm.metrics.Availability = float64(spm.metrics.SuccessfulProbes) / total * 100
}

// checkSLI проверяет SLI
func (spm *SyntheticProbeManager) checkSLI(probe *SyntheticProbe, result *ProbeResult) {
	// Получаем метрики с синхронизацией
	spm.metrics.mu.Lock()
	errorRate := spm.metrics.ErrorRate
	availability := spm.metrics.Availability
	successfulProbes := spm.metrics.SuccessfulProbes
	spm.metrics.mu.Unlock()

	// Вычисляем SLI метрики
	sliMetrics := &SLIMetrics{
		LatencyP99:   result.Latency,
		ErrorRate:    errorRate,
		Availability: availability,
		Throughput:   float64(successfulProbes) / time.Since(probe.CreatedAt).Seconds(),
	}

	// Вычисляем SLI score
	sliScore := spm.calculateSLIScore(sliMetrics)
	sliMetrics.SLIScore = sliScore

	result.SLIMetrics = sliMetrics
	spm.metrics.SLIScore = sliScore

	spm.logger.Debug("SLI metrics calculated",
		zap.String("probe_id", probe.ID),
		zap.Float64("sli_score", sliScore),
		zap.Float64("latency_p99", sliMetrics.LatencyP99),
		zap.Float64("error_rate", sliMetrics.ErrorRate),
		zap.Float64("availability", sliMetrics.Availability))
}

// calculateSLIScore вычисляет SLI score
func (spm *SyntheticProbeManager) calculateSLIScore(metrics *SLIMetrics) float64 {
	score := 1.0

	// Проверяем латентность
	if metrics.LatencyP99 > spm.config.SLIThresholds.LatencyP99 {
		score -= 0.3
	}

	// Проверяем error rate
	if metrics.ErrorRate > spm.config.SLIThresholds.ErrorRate {
		score -= 0.3
	}

	// Проверяем availability
	if metrics.Availability < spm.config.SLIThresholds.Availability {
		score -= 0.4
	}

	// Проверяем throughput
	if metrics.Throughput < spm.config.SLIThresholds.Throughput {
		score -= 0.1
	}

	if score < 0 {
		score = 0
	}

	return score
}

// checkAlerts проверяет алерты
func (spm *SyntheticProbeManager) checkAlerts(probe *SyntheticProbe, result *ProbeResult) {
	// Получаем метрики с синхронизацией
	spm.metrics.mu.Lock()
	errorRate := spm.metrics.ErrorRate
	availability := spm.metrics.Availability
	spm.metrics.mu.Unlock()

	alerts := []string{}

	// Проверяем латентность
	if result.Latency > spm.config.AlertThresholds.LatencyAlert {
		alerts = append(alerts, fmt.Sprintf("High latency: %.2fms", result.Latency))
	}

	// Проверяем error rate
	if errorRate > spm.config.AlertThresholds.ErrorRateAlert {
		alerts = append(alerts, fmt.Sprintf("High error rate: %.2f%%", errorRate))
	}

	// Проверяем availability
	if availability < spm.config.AlertThresholds.AvailabilityAlert {
		alerts = append(alerts, fmt.Sprintf("Low availability: %.2f%%", availability))
	}

	// Отправляем алерты
	if len(alerts) > 0 {
		spm.logger.Warn("Synthetic probe alerts triggered",
			zap.String("probe_id", probe.ID),
			zap.Strings("alerts", alerts))
	}
}

// GetProbeResults возвращает копии результатов probes
func (spm *SyntheticProbeManager) GetProbeResults() map[string]*ProbeResult {
	spm.resultsMux.RLock()
	defer spm.resultsMux.RUnlock()

	results := make(map[string]*ProbeResult)
	for id, result := range spm.results {
		// Создаем копию результата для безопасности
		resultCopy := *result
		if result.SLIMetrics != nil {
			sliCopy := *result.SLIMetrics
			resultCopy.SLIMetrics = &sliCopy
		}
		results[id] = &resultCopy
	}

	return results
}

// GetProbeMetrics возвращает копию метрик probes
func (spm *SyntheticProbeManager) GetProbeMetrics() *ProbeMetrics {
	spm.metrics.mu.Lock()
	defer spm.metrics.mu.Unlock()

	// Возвращаем копию метрик для безопасности
	return &ProbeMetrics{
		TotalProbes:      spm.metrics.TotalProbes,
		SuccessfulProbes: spm.metrics.SuccessfulProbes,
		FailedProbes:     spm.metrics.FailedProbes,
		AverageLatency:   spm.metrics.AverageLatency,
		ErrorRate:        spm.metrics.ErrorRate,
		Availability:     spm.metrics.Availability,
		SLIScore:         spm.metrics.SLIScore,
	}
}

// GetSLIStatus возвращает SLI статус
func (spm *SyntheticProbeManager) GetSLIStatus() map[string]interface{} {
	spm.metrics.mu.Lock()
	defer spm.metrics.mu.Unlock()

	return map[string]interface{}{
		"enabled":           spm.config.SLIEnabled,
		"thresholds":        spm.config.SLIThresholds,
		"current_score":     spm.metrics.SLIScore,
		"availability":      spm.metrics.Availability,
		"error_rate":        spm.metrics.ErrorRate,
		"average_latency":   spm.metrics.AverageLatency,
		"total_probes":      spm.metrics.TotalProbes,
		"successful_probes": spm.metrics.SuccessfulProbes,
		"failed_probes":     spm.metrics.FailedProbes,
	}
}
