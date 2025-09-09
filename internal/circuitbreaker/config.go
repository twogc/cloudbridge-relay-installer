package circuitbreaker

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// CircuitBreakerConfig holds configuration for multiple circuit breakers
type CircuitBreakerConfig struct {
	DefaultConfig *Config                     `json:"default_config"`
	Breakers      map[string]*Config          `json:"breakers"`
	GlobalConfig  *GlobalCircuitBreakerConfig `json:"global_config"`
}

// GlobalCircuitBreakerConfig holds global configuration for circuit breakers
type GlobalCircuitBreakerConfig struct {
	EnableMetrics     bool          `json:"enable_metrics"`
	MetricsInterval   time.Duration `json:"metrics_interval"`
	DefaultTimeout    time.Duration `json:"default_timeout"`
	DefaultWindowSize time.Duration `json:"default_window_size"`
	MaxBreakers       int           `json:"max_breakers"`
}

// DefaultGlobalConfig returns default global configuration
func DefaultGlobalConfig() *GlobalCircuitBreakerConfig {
	return &GlobalCircuitBreakerConfig{
		EnableMetrics:     true,
		MetricsInterval:   30 * time.Second,
		DefaultTimeout:    30 * time.Second,
		DefaultWindowSize: 60 * time.Second,
		MaxBreakers:       100,
	}
}

// LoadConfigFromFile loads circuit breaker configuration from a JSON file
func LoadConfigFromFile(filename string) (*CircuitBreakerConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config CircuitBreakerConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults if not provided
	if config.DefaultConfig == nil {
		config.DefaultConfig = DefaultConfig("default")
	}

	if config.GlobalConfig == nil {
		config.GlobalConfig = DefaultGlobalConfig()
	}

	if config.Breakers == nil {
		config.Breakers = make(map[string]*Config)
	}

	return &config, nil
}

// SaveConfigToFile saves circuit breaker configuration to a JSON file
func SaveConfigToFile(config *CircuitBreakerConfig, filename string) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(filename, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetBreakerConfig returns configuration for a specific circuit breaker
func (cbc *CircuitBreakerConfig) GetBreakerConfig(name string) *Config {
	// Check if specific config exists
	if config, exists := cbc.Breakers[name]; exists {
		return config
	}

	// Return default config
	return cbc.DefaultConfig
}

// AddBreakerConfig adds configuration for a specific circuit breaker
func (cbc *CircuitBreakerConfig) AddBreakerConfig(name string, config *Config) error {
	if cbc.GlobalConfig.MaxBreakers > 0 && len(cbc.Breakers) >= cbc.GlobalConfig.MaxBreakers {
		return fmt.Errorf("maximum number of circuit breakers reached (%d)", cbc.GlobalConfig.MaxBreakers)
	}

	cbc.Breakers[name] = config
	return nil
}

// RemoveBreakerConfig removes configuration for a specific circuit breaker
func (cbc *CircuitBreakerConfig) RemoveBreakerConfig(name string) {
	delete(cbc.Breakers, name)
}

// ValidateConfig validates circuit breaker configuration
func (cbc *CircuitBreakerConfig) ValidateConfig() error {
	// Validate default config
	if err := validateConfig(cbc.DefaultConfig); err != nil {
		return fmt.Errorf("default config validation failed: %w", err)
	}

	// Validate individual breaker configs
	for name, config := range cbc.Breakers {
		if err := validateConfig(config); err != nil {
			return fmt.Errorf("breaker config '%s' validation failed: %w", name, err)
		}
	}

	// Validate global config
	if cbc.GlobalConfig.MetricsInterval <= 0 {
		return fmt.Errorf("metrics interval must be positive")
	}

	if cbc.GlobalConfig.DefaultTimeout <= 0 {
		return fmt.Errorf("default timeout must be positive")
	}

	if cbc.GlobalConfig.DefaultWindowSize <= 0 {
		return fmt.Errorf("default window size must be positive")
	}

	if cbc.GlobalConfig.MaxBreakers < 0 {
		return fmt.Errorf("max breakers cannot be negative")
	}

	return nil
}

// validateConfig validates a single circuit breaker configuration
func validateConfig(config *Config) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	if config.Name == "" {
		return fmt.Errorf("name cannot be empty")
	}

	if config.FailureThreshold <= 0 {
		return fmt.Errorf("failure threshold must be positive")
	}

	if config.SuccessThreshold <= 0 {
		return fmt.Errorf("success threshold must be positive")
	}

	if config.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}

	if config.WindowSize <= 0 {
		return fmt.Errorf("window size must be positive")
	}

	return nil
}

// MergeConfigs merges multiple circuit breaker configurations
func MergeConfigs(configs ...*CircuitBreakerConfig) *CircuitBreakerConfig {
	if len(configs) == 0 {
		return &CircuitBreakerConfig{
			DefaultConfig: DefaultConfig("default"),
			GlobalConfig:  DefaultGlobalConfig(),
			Breakers:      make(map[string]*Config),
		}
	}

	merged := &CircuitBreakerConfig{
		DefaultConfig: configs[0].DefaultConfig,
		GlobalConfig:  configs[0].GlobalConfig,
		Breakers:      make(map[string]*Config),
	}

	// Merge breaker configs from all sources
	for _, config := range configs {
		for name, breakerConfig := range config.Breakers {
			merged.Breakers[name] = breakerConfig
		}
	}

	return merged
}

// CreateDefaultConfig creates a default circuit breaker configuration
func CreateDefaultConfig() *CircuitBreakerConfig {
	return &CircuitBreakerConfig{
		DefaultConfig: DefaultConfig("default"),
		GlobalConfig:  DefaultGlobalConfig(),
		Breakers: map[string]*Config{
			"database": {
				Name:             "database",
				FailureThreshold: 5,
				SuccessThreshold: 3,
				Timeout:          30 * time.Second,
				WindowSize:       60 * time.Second,
				AutoRecovery:     true,
			},
			"redis": {
				Name:             "redis",
				FailureThreshold: 3,
				SuccessThreshold: 2,
				Timeout:          10 * time.Second,
				WindowSize:       30 * time.Second,
				AutoRecovery:     true,
			},
			"external_api": {
				Name:             "external_api",
				FailureThreshold: 10,
				SuccessThreshold: 5,
				Timeout:          60 * time.Second,
				WindowSize:       120 * time.Second,
				AutoRecovery:     true,
			},
		},
	}
}
