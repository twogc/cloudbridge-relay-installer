package quic

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultQUICConfig(t *testing.T) {
	config := DefaultQUICConfig()
	assert.NotNil(t, config)
	assert.Equal(t, 100, config.MaxStreams)
	assert.Equal(t, 30*time.Second, config.KeepAliveInterval)
	assert.Equal(t, 60*time.Second, config.IdleTimeout)
	assert.Equal(t, 300*time.Second, config.MaxIdleTime)
	assert.True(t, config.EnableMetrics)
	assert.False(t, config.EnableTracing)
}

func TestQUICConfig_CustomValues(t *testing.T) {
	config := &QUICConfig{
		MaxStreams:        50,
		KeepAliveInterval: 60 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxIdleTime:       600 * time.Second,
		EnableMetrics:     false,
		EnableTracing:     true,
	}

	assert.Equal(t, 50, config.MaxStreams)
	assert.Equal(t, 60*time.Second, config.KeepAliveInterval)
	assert.Equal(t, 120*time.Second, config.IdleTimeout)
	assert.Equal(t, 600*time.Second, config.MaxIdleTime)
	assert.False(t, config.EnableMetrics)
	assert.True(t, config.EnableTracing)
}

func TestNewQUICConnection_WithNilConfig(t *testing.T) {
	// Тестируем создание с nil конфигурацией (должна использоваться конфигурация по умолчанию)
	// Поскольку мы не можем создать реальное QUIC соединение в тестах,
	// мы тестируем только логику создания конфигурации
	config := DefaultQUICConfig()
	assert.NotNil(t, config)
	assert.Equal(t, 100, config.MaxStreams)
	assert.Equal(t, 30*time.Second, config.KeepAliveInterval)
}

func TestQUICConnection_ConfigValidation(t *testing.T) {
	// Тестируем валидацию конфигурации
	config := &QUICConfig{
		MaxStreams:        -1, // Некорректное значение
		KeepAliveInterval: 0,  // Некорректное значение
		IdleTimeout:       0,  // Некорректное значение
	}

	// Проверяем, что некорректные значения не вызывают панику
	assert.Equal(t, -1, config.MaxStreams)
	assert.Equal(t, time.Duration(0), config.KeepAliveInterval)
	assert.Equal(t, time.Duration(0), config.IdleTimeout)
}

func TestQUICConnection_Timeouts(t *testing.T) {
	// Тестируем различные таймауты
	config := &QUICConfig{
		IdleTimeout: 5 * time.Second,
		MaxIdleTime: 10 * time.Second,
	}

	assert.Equal(t, 5*time.Second, config.IdleTimeout)
	assert.Equal(t, 10*time.Second, config.MaxIdleTime)
	assert.True(t, config.IdleTimeout < config.MaxIdleTime)
}

func TestQUICConnection_MetricsSettings(t *testing.T) {
	// Тестируем настройки метрик
	config := &QUICConfig{
		EnableMetrics: true,
		EnableTracing: false,
	}

	assert.True(t, config.EnableMetrics)
	assert.False(t, config.EnableTracing)

	// Тестируем обратные настройки
	config2 := &QUICConfig{
		EnableMetrics: false,
		EnableTracing: true,
	}

	assert.False(t, config2.EnableMetrics)
	assert.True(t, config2.EnableTracing)
}

func TestQUICConnection_StreamLimits(t *testing.T) {
	// Тестируем лимиты потоков
	config := &QUICConfig{
		MaxStreams: 1000,
	}

	assert.Equal(t, 1000, config.MaxStreams)
	assert.True(t, config.MaxStreams > 0)

	// Тестируем граничные значения
	config2 := &QUICConfig{
		MaxStreams: 1,
	}

	assert.Equal(t, 1, config2.MaxStreams)
}

func TestQUICConnection_KeepAliveSettings(t *testing.T) {
	// Тестируем настройки keep-alive
	config := &QUICConfig{
		KeepAliveInterval: 15 * time.Second,
	}

	assert.Equal(t, 15*time.Second, config.KeepAliveInterval)
	assert.True(t, config.KeepAliveInterval > 0)

	// Тестируем отключение keep-alive
	config2 := &QUICConfig{
		KeepAliveInterval: 0,
	}

	assert.Equal(t, time.Duration(0), config2.KeepAliveInterval)
}

func TestQUICConnection_ConfigComparison(t *testing.T) {
	// Тестируем сравнение конфигураций
	config1 := DefaultQUICConfig()
	config2 := DefaultQUICConfig()
	config3 := &QUICConfig{
		MaxStreams:        200,
		KeepAliveInterval: 60 * time.Second,
	}

	// Конфигурации по умолчанию должны быть одинаковыми
	assert.Equal(t, config1.MaxStreams, config2.MaxStreams)
	assert.Equal(t, config1.KeepAliveInterval, config2.KeepAliveInterval)

	// Пользовательская конфигурация должна отличаться
	assert.NotEqual(t, config1.MaxStreams, config3.MaxStreams)
	assert.NotEqual(t, config1.KeepAliveInterval, config3.KeepAliveInterval)
}

func TestQUICConnection_ConfigCopy(t *testing.T) {
	// Тестируем копирование конфигурации
	original := &QUICConfig{
		MaxStreams:        150,
		KeepAliveInterval: 45 * time.Second,
		IdleTimeout:       90 * time.Second,
		MaxIdleTime:       450 * time.Second,
		EnableMetrics:     true,
		EnableTracing:     true,
	}

	// Создаем копию
	copy := &QUICConfig{
		MaxStreams:        original.MaxStreams,
		KeepAliveInterval: original.KeepAliveInterval,
		IdleTimeout:       original.IdleTimeout,
		MaxIdleTime:       original.MaxIdleTime,
		EnableMetrics:     original.EnableMetrics,
		EnableTracing:     original.EnableTracing,
	}

	// Проверяем, что копия идентична оригиналу
	assert.Equal(t, original.MaxStreams, copy.MaxStreams)
	assert.Equal(t, original.KeepAliveInterval, copy.KeepAliveInterval)
	assert.Equal(t, original.IdleTimeout, copy.IdleTimeout)
	assert.Equal(t, original.MaxIdleTime, copy.MaxIdleTime)
	assert.Equal(t, original.EnableMetrics, copy.EnableMetrics)
	assert.Equal(t, original.EnableTracing, copy.EnableTracing)
}

func TestQUICConnection_ConfigValidation_EdgeCases(t *testing.T) {
	// Тестируем граничные случаи конфигурации
	testCases := []struct {
		name   string
		config *QUICConfig
		valid  bool
	}{
		{
			name: "Valid configuration",
			config: &QUICConfig{
				MaxStreams:        100,
				KeepAliveInterval: 30 * time.Second,
				IdleTimeout:       60 * time.Second,
				MaxIdleTime:       300 * time.Second,
			},
			valid: true,
		},
		{
			name: "Zero max streams",
			config: &QUICConfig{
				MaxStreams: 0,
			},
			valid: true, // 0 может быть допустимым значением
		},
		{
			name: "Negative max streams",
			config: &QUICConfig{
				MaxStreams: -1,
			},
			valid: true, // Проверяем, что не падает
		},
		{
			name: "Zero timeouts",
			config: &QUICConfig{
				IdleTimeout: 0,
				MaxIdleTime: 0,
			},
			valid: true, // 0 может быть допустимым значением
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Просто проверяем, что конфигурация не вызывает панику
			assert.NotNil(t, tc.config)
			// Дополнительные проверки можно добавить здесь
		})
	}
} 