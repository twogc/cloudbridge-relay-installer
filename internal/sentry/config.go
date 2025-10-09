package sentry

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/getsentry/sentry-go"
)

// Config конфигурация Sentry
type Config struct {
	DSN              string  `yaml:"dsn" json:"dsn"`
	Environment      string  `yaml:"environment" json:"environment"`
	Release          string  `yaml:"release" json:"release"`
	TracesSampleRate float64 `yaml:"traces_sample_rate" json:"traces_sample_rate"`
	EnableTracing    bool    `yaml:"enable_tracing" json:"enable_tracing"`
	EnableLogs       bool    `yaml:"enable_logs" json:"enable_logs"`
	Debug            bool    `yaml:"debug" json:"debug"`
	ServerName       string  `yaml:"server_name" json:"server_name"`
	AttachStacktrace bool    `yaml:"attach_stacktrace" json:"attach_stacktrace"`
	MaxBreadcrumbs   int     `yaml:"max_breadcrumbs" json:"max_breadcrumbs"`
	SampleRate       float64 `yaml:"sample_rate" json:"sample_rate"`
	BeforeSend       func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig() *Config {
	// Парсим TracesSampleRate из переменной окружения
	tracesSampleRate := 0.1
	if rateStr := getEnvOrDefault("SENTRY_TRACES_SAMPLE_RATE", ""); rateStr != "" {
		if rate, err := strconv.ParseFloat(rateStr, 64); err == nil {
			tracesSampleRate = rate
		}
	}

	// Парсим SampleRate из переменной окружения
	sampleRate := 1.0
	if rateStr := getEnvOrDefault("SENTRY_SAMPLE_RATE", ""); rateStr != "" {
		if rate, err := strconv.ParseFloat(rateStr, 64); err == nil {
			sampleRate = rate
		}
	}

	// Парсим MaxBreadcrumbs из переменной окружения
	maxBreadcrumbs := 50
	if breadcrumbsStr := getEnvOrDefault("SENTRY_MAX_BREADCRUMBS", ""); breadcrumbsStr != "" {
		if breadcrumbs, err := strconv.Atoi(breadcrumbsStr); err == nil {
			maxBreadcrumbs = breadcrumbs
		}
	}

	return &Config{
		DSN:              getEnvOrDefault("SENTRY_DSN", ""),
		Environment:      getEnvOrDefault("SENTRY_ENVIRONMENT", "production"),
		Release:          getEnvOrDefault("SENTRY_RELEASE", "cloudbridge-relay@1.0.0"),
		TracesSampleRate: tracesSampleRate,
		EnableTracing:    true,
		EnableLogs:       true,
		Debug:            getEnvOrDefault("SENTRY_DEBUG", "false") == "true",
		ServerName:       getEnvOrDefault("SENTRY_SERVER_NAME", "cloudbridge-relay"),
		AttachStacktrace: true,
		MaxBreadcrumbs:   maxBreadcrumbs,
		SampleRate:       sampleRate,
	}
}

// Init инициализирует Sentry
func Init(config *Config) error {
	if config.DSN == "" {
		return fmt.Errorf("Sentry DSN is required")
	}

	clientOptions := sentry.ClientOptions{
		Dsn:              config.DSN,
		Environment:      config.Environment,
		Release:          config.Release,
		TracesSampleRate: config.TracesSampleRate,
		EnableTracing:    config.EnableTracing,
		Debug:            config.Debug,
		ServerName:       config.ServerName,
		AttachStacktrace: config.AttachStacktrace,
		MaxBreadcrumbs:   config.MaxBreadcrumbs,
		SampleRate:       config.SampleRate,
		BeforeSend:       config.BeforeSend,
		// Integrations будут добавлены автоматически
		Transport: &sentry.HTTPTransport{
			Timeout: 30 * time.Second,
			BufferSize: 1000, // Увеличиваем размер буфера
		},
	}

	return sentry.Init(clientOptions)
}

// getEnvOrDefault возвращает значение переменной окружения или значение по умолчанию
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Flush ждет отправки всех событий
func Flush(timeout time.Duration) bool {
	return sentry.Flush(timeout)
}

// Recover восстанавливает панику и отправляет в Sentry
func Recover() {
	if r := recover(); r != nil {
		sentry.CurrentHub().Recover(r)
		panic(r)
	}
}

// CaptureException отправляет исключение в Sentry
func CaptureException(err error) *sentry.EventID {
	return sentry.CaptureException(err)
}

// CaptureMessage отправляет сообщение в Sentry
func CaptureMessage(message string) *sentry.EventID {
	return sentry.CaptureMessage(message)
}

// AddBreadcrumb добавляет хлебные крошки
func AddBreadcrumb(breadcrumb *sentry.Breadcrumb) {
	sentry.AddBreadcrumb(breadcrumb)
}

// SetTag устанавливает тег
func SetTag(key, value string) {
	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetTag(key, value)
	})
}

// SetUser устанавливает пользователя
func SetUser(user sentry.User) {
	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetUser(user)
	})
}

// SetContext устанавливает контекст
func SetContext(key string, value interface{}) {
	sentry.ConfigureScope(func(scope *sentry.Scope) {
		// Преобразуем interface{} в sentry.Context
		if ctx, ok := value.(sentry.Context); ok {
			scope.SetContext(key, ctx)
		} else {
			// Если не удается преобразовать, используем Extra
			scope.SetExtra(key, value)
		}
	})
}
