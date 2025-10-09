package sentry

import (
	"context"
	"fmt"

	sentrygo "github.com/getsentry/sentry-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// SentryZapCore создает zap core для интеграции с Sentry
type SentryZapCore struct {
	zapcore.Core
	level zapcore.Level
}

// NewSentryZapCore создает новый Sentry zap core
func NewSentryZapCore(level zapcore.Level) *SentryZapCore {
	// Создаем базовый core
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(level)
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	baseCore, err := config.Build()
	if err != nil {
		// Fallback к простому core если не удалось создать
		baseCore, _ = zap.NewProduction()
	}

	return &SentryZapCore{
		Core:  baseCore.Core(),
		level: level,
	}
}

// Enabled проверяет, включен ли уровень логирования
func (s *SentryZapCore) Enabled(level zapcore.Level) bool {
	if s == nil {
		return false
	}
	return level >= s.level
}

// Write записывает лог в Sentry
func (s *SentryZapCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	// Проверяем, что Sentry инициализирован
	if s == nil || s.Core == nil {
		return nil
	}

	// Создаем scope для Sentry
	sentrygo.ConfigureScope(func(scope *sentrygo.Scope) {
		// Устанавливаем уровень логирования
		scope.SetLevel(sentrygo.Level(entry.Level.String()))

		// Добавляем поля как extra данные
		for _, field := range fields {
			scope.SetExtra(field.Key, field.Interface)
		}

		// Устанавливаем теги
		scope.SetTag("logger", "zap")
		scope.SetTag("component", "cloudbridge-relay")

		// Добавляем breadcrumb
		sentrygo.AddBreadcrumb(&sentrygo.Breadcrumb{
			Message:   entry.Message,
			Category:  "log",
			Level:     sentrygo.Level(entry.Level.String()),
			Timestamp: entry.Time,
		})
	})

	// Отправляем в Sentry в зависимости от уровня
	switch entry.Level {
	case zapcore.ErrorLevel, zapcore.FatalLevel, zapcore.PanicLevel:
		// Для ошибок отправляем как исключение
		sentrygo.CaptureMessage(fmt.Sprintf("[%s] %s", entry.Level.String(), entry.Message))
	case zapcore.WarnLevel:
		// Для предупреждений отправляем как сообщение
		sentrygo.CaptureMessage(fmt.Sprintf("[WARN] %s", entry.Message))
	case zapcore.InfoLevel:
		// Для информационных сообщений отправляем как breadcrumb
		sentrygo.CaptureMessage(fmt.Sprintf("[INFO] %s", entry.Message))
	default:
		// Для остальных уровней отправляем как breadcrumb
		sentrygo.CaptureMessage(fmt.Sprintf("[%s] %s", entry.Level.String(), entry.Message))
	}

	return nil
}

// Sync синхронизирует логи
func (s *SentryZapCore) Sync() error {
	return nil
}

// With добавляет поля к core
func (s *SentryZapCore) With(fields []zapcore.Field) zapcore.Core {
	if s == nil {
		return nil
	}
	return s
}

// NewSentryLogger создает новый logger с интеграцией Sentry
func NewSentryLogger(level zapcore.Level) (*zap.Logger, error) {
	// Создаем конфигурацию zap
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(level)
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Создаем Sentry core
	sentryCore := NewSentryZapCore(level)

	// Создаем logger с Sentry core
	logger, err := config.Build(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return zapcore.NewTee(core, sentryCore)
	}))

	if err != nil {
		return nil, fmt.Errorf("failed to create Sentry logger: %w", err)
	}

	return logger, nil
}

// LogLevelFromEnv получает уровень логирования из переменной окружения
func LogLevelFromEnv() zapcore.Level {
	level := getEnvOrDefault("LOG_LEVEL", "info")

	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// SetupLogging настраивает логирование с интеграцией Sentry
func SetupLogging() (*zap.Logger, error) {
	level := LogLevelFromEnv()

	// Проверяем, что Sentry DSN настроен
	dsn := getEnvOrDefault("SENTRY_DSN", "")
	if dsn == "" {
		// Создаем простой logger без Sentry
		config := zap.NewProductionConfig()
		config.Level = zap.NewAtomicLevelAt(level)
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

		logger, err := config.Build()
		if err != nil {
			return nil, fmt.Errorf("failed to create logger: %w", err)
		}

		return logger, nil
	}

	// Создаем logger с Sentry
	logger, err := NewSentryLogger(level)
	if err != nil {
		return nil, fmt.Errorf("failed to setup logging: %w", err)
	}

	// Заменяем стандартный logger
	zap.ReplaceGlobals(logger)

	return logger, nil
}

// LogError логирует ошибку в Sentry и zap
func LogError(err error, message string, fields ...zap.Field) {
	if err != nil {
		// Логируем в zap (с проверкой на nil)
		if logger := zap.L(); logger != nil {
			logger.Error(message, append(fields, zap.Error(err))...)
		}

		// Отправляем в Sentry
		sentrygo.CaptureException(err)
	}
}

// LogInfo логирует информационное сообщение
func LogInfo(message string, fields ...zap.Field) {
	if logger := zap.L(); logger != nil {
		logger.Info(message, fields...)
	}
}

// LogWarn логирует предупреждение
func LogWarn(message string, fields ...zap.Field) {
	if logger := zap.L(); logger != nil {
		logger.Warn(message, fields...)
	}
}

// LogDebug логирует отладочное сообщение
func LogDebug(message string, fields ...zap.Field) {
	if logger := zap.L(); logger != nil {
		logger.Debug(message, fields...)
	}
}

// LogFatal логирует критическую ошибку и завершает программу
func LogFatal(message string, fields ...zap.Field) {
	if logger := zap.L(); logger != nil {
		logger.Fatal(message, fields...)
	}
}

// LogWithContext логирует с контекстом
func LogWithContext(ctx context.Context, level zapcore.Level, message string, fields ...zap.Field) {
	// Добавляем контекст в Sentry
	if hub := sentrygo.GetHubFromContext(ctx); hub != nil {
		hub.ConfigureScope(func(scope *sentrygo.Scope) {
			for _, field := range fields {
				scope.SetExtra(field.Key, field.Interface)
			}
		})
	}

	// Логируем в zap (с проверкой на nil)
	if logger := zap.L(); logger != nil {
		switch level {
		case zapcore.DebugLevel:
			logger.Debug(message, fields...)
		case zapcore.InfoLevel:
			logger.Info(message, fields...)
		case zapcore.WarnLevel:
			logger.Warn(message, fields...)
		case zapcore.ErrorLevel:
			logger.Error(message, fields...)
		case zapcore.FatalLevel:
			logger.Fatal(message, fields...)
		}
	}
}
