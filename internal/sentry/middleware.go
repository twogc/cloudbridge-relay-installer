package sentry

import (
	"time"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
)

// MiddlewareOptions опции для Sentry middleware
type MiddlewareOptions struct {
	Repanic         bool          `yaml:"repanic" json:"repanic"`
	WaitForDelivery bool          `yaml:"wait_for_delivery" json:"wait_for_delivery"`
	Timeout         time.Duration `yaml:"timeout" json:"timeout"`
	EnableTracing   bool          `yaml:"enable_tracing" json:"enable_tracing"`
}

// DefaultMiddlewareOptions возвращает опции по умолчанию
func DefaultMiddlewareOptions() *MiddlewareOptions {
	return &MiddlewareOptions{
		Repanic:         true,
		WaitForDelivery: false,
		Timeout:         30 * time.Second,
		EnableTracing:   true,
	}
}

// NewMiddleware создает новый Sentry middleware для Gin
func NewMiddleware(options *MiddlewareOptions) gin.HandlerFunc {
	if options == nil {
		options = DefaultMiddlewareOptions()
	}

	// Создаем Sentry Gin middleware
	sentryMiddleware := sentrygin.New(sentrygin.Options{
		Repanic:         options.Repanic,
		WaitForDelivery: options.WaitForDelivery,
		Timeout:         options.Timeout,
	})

	// Обертываем в дополнительную логику для CloudBridge
	return func(c *gin.Context) {
		// Добавляем контекст CloudBridge
		hub := sentrygin.GetHubFromContext(c)
		if hub != nil {
			// Устанавливаем теги для CloudBridge
			hub.Scope().SetTag("service", "cloudbridge-relay")
			hub.Scope().SetTag("component", "gin-router")
			hub.Scope().SetTag("version", "1.0.0")

			// Добавляем информацию о запросе
			hub.Scope().SetContext("request", map[string]interface{}{
				"method":     c.Request.Method,
				"url":        c.Request.URL.String(),
				"host":       c.Request.Host,
				"user_agent": c.Request.UserAgent(),
			})
		}

		// Выполняем оригинальный Sentry middleware
		sentryMiddleware(c)

		// Добавляем дополнительную логику после обработки
		if hub != nil {
			// Логируем время выполнения
			start := time.Now()
			c.Next()
			duration := time.Since(start)

			// Добавляем метрики производительности
			hub.Scope().SetExtra("request_duration_ms", duration.Milliseconds())
			hub.Scope().SetExtra("status_code", c.Writer.Status())

			// Если есть ошибка, логируем её
			if c.Writer.Status() >= 400 {
				hub.Scope().SetLevel(sentry.LevelError)
				hub.CaptureMessage("HTTP error occurred")
			}
		} else {
			c.Next()
		}
	}
}

// NewCloudBridgeMiddleware создает специализированный middleware для CloudBridge
func NewCloudBridgeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		hub := sentrygin.GetHubFromContext(c)
		if hub == nil {
			c.Next()
			return
		}

		// Добавляем специфичные для CloudBridge теги
		hub.Scope().SetTag("tenant_id", c.GetHeader("X-Tenant-ID"))
		hub.Scope().SetTag("user_id", c.GetHeader("X-User-ID"))
		hub.Scope().SetTag("connection_id", c.GetHeader("X-Connection-ID"))
		hub.Scope().SetTag("protocol", c.GetHeader("X-Protocol"))

		// Добавляем breadcrumb для отслеживания
		hub.AddBreadcrumb(&sentry.Breadcrumb{
			Category: "http",
			Message:  "CloudBridge request started",
			Level:    sentry.LevelInfo,
			Data: map[string]interface{}{
				"method": c.Request.Method,
				"path":   c.Request.URL.Path,
			},
		}, nil)

		c.Next()

		// Логируем завершение запроса
		hub.AddBreadcrumb(&sentry.Breadcrumb{
			Category: "http",
			Message:  "CloudBridge request completed",
			Level:    sentry.LevelInfo,
			Data: map[string]interface{}{
				"status_code": c.Writer.Status(),
				"duration_ms": time.Since(time.Now()).Milliseconds(),
			},
		}, nil)
	}
}

// CaptureCloudBridgeError отправляет ошибку CloudBridge в Sentry
func CaptureCloudBridgeError(c *gin.Context, err error, context map[string]interface{}) {
	hub := sentrygin.GetHubFromContext(c)
	if hub == nil {
		return
	}

	// Добавляем контекст CloudBridge
	hub.Scope().SetTag("error_type", "cloudbridge")
	hub.Scope().SetTag("component", "relay")

	// Добавляем дополнительный контекст
	for key, value := range context {
		hub.Scope().SetExtra(key, value)
	}

	// Отправляем ошибку
	hub.CaptureException(err)
}

// CaptureNetworkError отправляет сетевую ошибку в Sentry
func CaptureNetworkError(c *gin.Context, err error, networkInfo map[string]interface{}) {
	hub := sentrygin.GetHubFromContext(c)
	if hub == nil {
		return
	}

	// Добавляем теги для сетевых ошибок
	hub.Scope().SetTag("error_type", "network")
	hub.Scope().SetTag("component", "networking")

	// Добавляем сетевую информацию
	for key, value := range networkInfo {
		hub.Scope().SetExtra(key, value)
	}

	// Отправляем ошибку
	hub.CaptureException(err)
}

// CaptureAuthError отправляет ошибку аутентификации в Sentry
func CaptureAuthError(c *gin.Context, err error, authInfo map[string]interface{}) {
	hub := sentrygin.GetHubFromContext(c)
	if hub == nil {
		return
	}

	// Добавляем теги для ошибок аутентификации
	hub.Scope().SetTag("error_type", "authentication")
	hub.Scope().SetTag("component", "auth")

	// Добавляем информацию об аутентификации
	for key, value := range authInfo {
		hub.Scope().SetExtra(key, value)
	}

	// Отправляем ошибку
	hub.CaptureException(err)
}

// SetCloudBridgeContext устанавливает контекст CloudBridge
func SetCloudBridgeContext(c *gin.Context, tenantID, userID, connectionID string) {
	hub := sentrygin.GetHubFromContext(c)
	if hub == nil {
		return
	}

	// Устанавливаем пользователя
	hub.Scope().SetUser(sentry.User{
		ID:       userID,
		Username: "cloudbridge-user",
	})

	// Устанавливаем теги
	hub.Scope().SetTag("tenant_id", tenantID)
	hub.Scope().SetTag("user_id", userID)
	hub.Scope().SetTag("connection_id", connectionID)
}
