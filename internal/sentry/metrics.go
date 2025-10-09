package sentry

import (
	"context"
	"fmt"
	"sync"
	"time"

	sentrygo "github.com/getsentry/sentry-go"
	"go.uber.org/zap"
)

// MetricsCollector собирает метрики для отправки в Sentry
type MetricsCollector struct {
	mu       sync.RWMutex
	metrics  map[string]interface{}
	counters map[string]int64
	gauges   map[string]float64
	timers   map[string][]time.Duration
}

// NewMetricsCollector создает новый коллектор метрик
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		metrics:  make(map[string]interface{}),
		counters: make(map[string]int64),
		gauges:   make(map[string]float64),
		timers:   make(map[string][]time.Duration),
	}
}

// IncrementCounter увеличивает счетчик
func (mc *MetricsCollector) IncrementCounter(name string, value int64) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.counters[name] += value

	// Отправляем метрику в Sentry
	sentrygo.ConfigureScope(func(scope *sentrygo.Scope) {
		scope.SetExtra(fmt.Sprintf("counter_%s", name), mc.counters[name])
		scope.SetTag("metric_type", "counter")
		scope.SetTag("metric_name", name)
	})

	// Логируем метрику (с проверкой на nil)
	if logger := zap.L(); logger != nil {
		logger.Info("Counter incremented",
			zap.String("metric", name),
			zap.Int64("value", value),
			zap.Int64("total", mc.counters[name]),
		)
	}
}

// SetGauge устанавливает значение gauge
func (mc *MetricsCollector) SetGauge(name string, value float64) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.gauges[name] = value

	// Отправляем метрику в Sentry
	sentrygo.ConfigureScope(func(scope *sentrygo.Scope) {
		scope.SetExtra(fmt.Sprintf("gauge_%s", name), value)
		scope.SetTag("metric_type", "gauge")
		scope.SetTag("metric_name", name)
	})

	// Логируем метрику (с проверкой на nil)
	if logger := zap.L(); logger != nil {
		logger.Info("Gauge set",
			zap.String("metric", name),
			zap.Float64("value", value),
		)
	}
}

// RecordTimer записывает время выполнения
func (mc *MetricsCollector) RecordTimer(name string, duration time.Duration) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.timers[name] = append(mc.timers[name], duration)

	// Ограничиваем количество записей
	if len(mc.timers[name]) > 100 {
		mc.timers[name] = mc.timers[name][1:]
	}

	// Отправляем метрику в Sentry
	sentrygo.ConfigureScope(func(scope *sentrygo.Scope) {
		scope.SetExtra(fmt.Sprintf("timer_%s", name), duration.Milliseconds())
		scope.SetTag("metric_type", "timer")
		scope.SetTag("metric_name", name)
	})

	// Логируем метрику (с проверкой на nil)
	if logger := zap.L(); logger != nil {
		logger.Info("Timer recorded",
			zap.String("metric", name),
			zap.Duration("duration", duration),
		)
	}
}

// GetCounter возвращает значение счетчика
func (mc *MetricsCollector) GetCounter(name string) int64 {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.counters[name]
}

// GetGauge возвращает значение gauge
func (mc *MetricsCollector) GetGauge(name string) float64 {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.gauges[name]
}

// GetTimerStats возвращает статистику таймера
func (mc *MetricsCollector) GetTimerStats(name string) (count int, avg, min, max time.Duration) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	timers := mc.timers[name]
	if len(timers) == 0 {
		return 0, 0, 0, 0
	}

	count = len(timers)
	min = timers[0]
	max = timers[0]
	total := time.Duration(0)

	for _, timer := range timers {
		if timer < min {
			min = timer
		}
		if timer > max {
			max = timer
		}
		total += timer
	}

	avg = total / time.Duration(count)
	return count, avg, min, max
}

// SendMetricsToSentry отправляет метрики в Sentry
func (mc *MetricsCollector) SendMetricsToSentry() {
	// Проверяем, что Sentry DSN настроен
	dsn := getEnvOrDefault("SENTRY_DSN", "")
	if dsn == "" {
		// Не отправляем метрики без Sentry
		return
	}
	
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	// Создаем сообщение с метриками
	message := "CloudBridge Relay Metrics Report"

	// Добавляем метрики в scope
	sentrygo.ConfigureScope(func(scope *sentrygo.Scope) {
		scope.SetTag("component", "metrics")
		scope.SetTag("report_type", "periodic")

		// Добавляем счетчики
		for name, value := range mc.counters {
			scope.SetExtra(fmt.Sprintf("counter_%s", name), value)
		}

		// Добавляем gauges
		for name, value := range mc.gauges {
			scope.SetExtra(fmt.Sprintf("gauge_%s", name), value)
		}

		// Добавляем статистику таймеров
		for name := range mc.timers {
			count, avg, min, max := mc.GetTimerStats(name)
			scope.SetExtra(fmt.Sprintf("timer_%s_count", name), count)
			scope.SetExtra(fmt.Sprintf("timer_%s_avg_ms", name), avg.Milliseconds())
			scope.SetExtra(fmt.Sprintf("timer_%s_min_ms", name), min.Milliseconds())
			scope.SetExtra(fmt.Sprintf("timer_%s_max_ms", name), max.Milliseconds())
		}
	})

	// Отправляем в Sentry
	sentrygo.CaptureMessage(message)

	// Логируем отправку метрик (с проверкой на nil)
	if logger := zap.L(); logger != nil {
		logger.Info("Metrics sent to Sentry",
			zap.Int("counters", len(mc.counters)),
			zap.Int("gauges", len(mc.gauges)),
			zap.Int("timers", len(mc.timers)),
		)
	}
}

// StartMetricsReporter запускает периодическую отправку метрик
func (mc *MetricsCollector) StartMetricsReporter(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			mc.SendMetricsToSentry()
		}
	}()
}

// Global metrics collector
var (
	globalMetrics = NewMetricsCollector()
)

// IncrementCounter увеличивает глобальный счетчик
func IncrementCounter(name string, value int64) {
	globalMetrics.IncrementCounter(name, value)
}

// SetGauge устанавливает глобальный gauge
func SetGauge(name string, value float64) {
	globalMetrics.SetGauge(name, value)
}

// RecordTimer записывает глобальный таймер
func RecordTimer(name string, duration time.Duration) {
	globalMetrics.RecordTimer(name, duration)
}

// StartGlobalMetricsReporter запускает глобальный репортер метрик
func StartGlobalMetricsReporter(interval time.Duration) {
	// Проверяем, что Sentry DSN настроен
	dsn := getEnvOrDefault("SENTRY_DSN", "")
	if dsn == "" {
		// Не запускаем метрики без Sentry
		return
	}
	
	globalMetrics.StartMetricsReporter(interval)
}

// SendGlobalMetricsToSentry отправляет глобальные метрики в Sentry
func SendGlobalMetricsToSentry() {
	globalMetrics.SendMetricsToSentry()
}

// Custom metrics for CloudBridge Relay
func RecordRequestMetrics(ctx context.Context, method, path string, statusCode int, duration time.Duration) {
	// Записываем метрики запроса
	IncrementCounter("requests_total", 1)
	IncrementCounter(fmt.Sprintf("requests_%s_total", method), 1)
	IncrementCounter(fmt.Sprintf("requests_status_%d_total", statusCode), 1)

	RecordTimer("request_duration", duration)
	RecordTimer(fmt.Sprintf("request_duration_%s", method), duration)

	// Устанавливаем gauge для активных соединений
	SetGauge("active_connections", 1)

	// Логируем метрики запроса (с проверкой на nil)
	if logger := zap.L(); logger != nil {
		logger.Info("Request metrics recorded",
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status_code", statusCode),
			zap.Duration("duration", duration),
		)
	}
}

// RecordErrorMetrics записывает метрики ошибок
func RecordErrorMetrics(ctx context.Context, errorType, component string, count int64) {
	IncrementCounter("errors_total", count)
	IncrementCounter(fmt.Sprintf("errors_%s_total", errorType), count)
	IncrementCounter(fmt.Sprintf("errors_%s_total", component), count)

	// Отправляем в Sentry с контекстом
	sentrygo.ConfigureScope(func(scope *sentrygo.Scope) {
		scope.SetTag("error_type", errorType)
		scope.SetTag("component", component)
		scope.SetExtra("error_count", count)
	})

	sentrygo.CaptureMessage(fmt.Sprintf("Error metrics: %s in %s", errorType, component))

	// Логируем метрики ошибок (с проверкой на nil)
	if logger := zap.L(); logger != nil {
		logger.Warn("Error metrics recorded",
			zap.String("error_type", errorType),
			zap.String("component", component),
			zap.Int64("count", count),
		)
	}
}

// RecordPerformanceMetrics записывает метрики производительности
func RecordPerformanceMetrics(ctx context.Context, component string, duration time.Duration, memoryUsage int64) {
	RecordTimer(fmt.Sprintf("performance_%s", component), duration)
	SetGauge(fmt.Sprintf("memory_usage_%s", component), float64(memoryUsage))

	// Логируем метрики производительности (с проверкой на nil)
	if logger := zap.L(); logger != nil {
		logger.Info("Performance metrics recorded",
			zap.String("component", component),
			zap.Duration("duration", duration),
			zap.Int64("memory_usage", memoryUsage),
		)
	}
}
