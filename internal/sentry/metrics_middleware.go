package sentry

import (
	"runtime"
	"strconv"
	"time"

	sentrygo "github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// MetricsMiddleware middleware для сбора метрик
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Получаем информацию о запросе
		method := c.Request.Method
		path := c.Request.URL.Path

		// Записываем начало запроса
		IncrementCounter("requests_started_total", 1)

		// Обрабатываем запрос
		c.Next()

		// Вычисляем метрики
		duration := time.Since(start)
		statusCode := c.Writer.Status()

		// Записываем метрики запроса
		RecordRequestMetrics(c.Request.Context(), method, path, statusCode, duration)

		// Записываем метрики производительности
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		RecordPerformanceMetrics(c.Request.Context(), "http_request", duration, int64(m.Alloc))

		// Записываем метрики ошибок если есть
		if statusCode >= 400 {
			errorType := "http_error"
			if statusCode >= 500 {
				errorType = "server_error"
			} else if statusCode >= 400 {
				errorType = "client_error"
			}

			RecordErrorMetrics(c.Request.Context(), errorType, "http", 1)
		}
	}
}

// LoggingMiddleware middleware для логирования
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Логируем начало запроса (с проверкой на nil)
		if logger := zap.L(); logger != nil {
			logger.Info("Request started",
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.String("remote_addr", c.ClientIP()),
				zap.String("user_agent", c.Request.UserAgent()),
			)
		}

		// Обрабатываем запрос
		c.Next()

		// Логируем завершение запроса
		duration := time.Since(start)
		statusCode := c.Writer.Status()

		// Логируем завершение запроса (с проверкой на nil)
		if logger := zap.L(); logger != nil {
			logger.Info("Request completed",
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.Int("status_code", statusCode),
				zap.Duration("duration", duration),
				zap.Int("response_size", c.Writer.Size()),
			)
		}

		// Логируем ошибки (с проверкой на nil)
		if statusCode >= 400 {
			if logger := zap.L(); logger != nil {
				logger.Warn("Request error",
					zap.String("method", c.Request.Method),
					zap.String("path", c.Request.URL.Path),
					zap.Int("status_code", statusCode),
					zap.Duration("duration", duration),
				)
			}
		}
	}
}

// SentryContextMiddleware middleware для добавления контекста в Sentry
func SentryContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Добавляем контекст в Sentry
		sentrygo.ConfigureScope(func(scope *sentrygo.Scope) {
			scope.SetTag("request_method", c.Request.Method)
			scope.SetTag("request_path", c.Request.URL.Path)
			scope.SetTag("remote_addr", c.ClientIP())
			scope.SetTag("user_agent", c.Request.UserAgent())

			// Добавляем параметры запроса
			if len(c.Request.URL.RawQuery) > 0 {
				scope.SetExtra("query_params", c.Request.URL.RawQuery)
			}

			// Добавляем заголовки
			scope.SetExtra("content_type", c.Request.Header.Get("Content-Type"))
			scope.SetExtra("content_length", c.Request.ContentLength)
		})

		// Обрабатываем запрос
		c.Next()

		// Добавляем информацию о ответе
		sentrygo.ConfigureScope(func(scope *sentrygo.Scope) {
			scope.SetTag("response_status", strconv.Itoa(c.Writer.Status()))
			scope.SetExtra("response_size", c.Writer.Size())
		})
	}
}

// HealthCheckMiddleware middleware для проверки здоровья
func HealthCheckMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Проверяем здоровье системы
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		// Устанавливаем метрики здоровья
		SetGauge("memory_alloc_bytes", float64(m.Alloc))
		SetGauge("memory_total_alloc_bytes", float64(m.TotalAlloc))
		SetGauge("memory_sys_bytes", float64(m.Sys))
		SetGauge("memory_num_gc", float64(m.NumGC))
		SetGauge("goroutines", float64(runtime.NumGoroutine()))

		// Логируем метрики здоровья (с проверкой на nil)
		if logger := zap.L(); logger != nil {
			logger.Debug("Health check metrics",
				zap.Uint64("memory_alloc", m.Alloc),
				zap.Uint64("memory_total_alloc", m.TotalAlloc),
				zap.Uint64("memory_sys", m.Sys),
				zap.Uint32("num_gc", m.NumGC),
				zap.Int("goroutines", runtime.NumGoroutine()),
			)
		}

		// Обрабатываем запрос
		c.Next()
	}
}

// ErrorHandlingMiddleware middleware для обработки ошибок
func ErrorHandlingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Обрабатываем запрос
		c.Next()

		// Проверяем на ошибки
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				// Логируем ошибку (с проверкой на nil)
				if logger := zap.L(); logger != nil {
					logger.Error("Request error",
						zap.String("method", c.Request.Method),
						zap.String("path", c.Request.URL.Path),
						zap.String("error", err.Error()),
						zap.String("error_type", string(err.Type)),
					)
				}

				// Отправляем в Sentry
				sentrygo.ConfigureScope(func(scope *sentrygo.Scope) {
					scope.SetTag("error_type", string(err.Type))
					scope.SetTag("request_method", c.Request.Method)
					scope.SetTag("request_path", c.Request.URL.Path)
				})

				sentrygo.CaptureException(err.Err)
			}
		}
	}
}

// PerformanceMonitoringMiddleware middleware для мониторинга производительности
func PerformanceMonitoringMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Обрабатываем запрос
		c.Next()

		duration := time.Since(start)

		// Записываем метрики производительности
		RecordTimer("request_duration_total", duration)

		// Проверяем на медленные запросы (с проверкой на nil)
		if duration > 1*time.Second {
			if logger := zap.L(); logger != nil {
				logger.Warn("Slow request detected",
					zap.String("method", c.Request.Method),
					zap.String("path", c.Request.URL.Path),
					zap.Duration("duration", duration),
				)
			}

			// Отправляем в Sentry
			sentrygo.ConfigureScope(func(scope *sentrygo.Scope) {
				scope.SetTag("slow_request", "true")
				scope.SetTag("request_method", c.Request.Method)
				scope.SetTag("request_path", c.Request.URL.Path)
				scope.SetExtra("duration_ms", duration.Milliseconds())
			})

			sentrygo.CaptureMessage("Slow request detected")
		}
	}
}
