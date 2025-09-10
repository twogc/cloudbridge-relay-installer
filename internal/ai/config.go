package ai

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the AI configuration
type Config struct {
	AI AI `yaml:"ai"`
}

// AI configuration structure
type AI struct {
	Enabled               bool          `yaml:"enabled"`
	Mode                  string        `yaml:"mode"`
	LogLevel              string        `yaml:"log_level"`
	MaxConcurrentRequests int           `yaml:"max_concurrent_requests"`
	RequestTimeout        time.Duration `yaml:"request_timeout"`
	CacheTTL              time.Duration `yaml:"cache_ttl"`
	Models                Models        `yaml:"models"`
	DataPipeline          DataPipeline  `yaml:"data_pipeline"`
	Monitoring            Monitoring    `yaml:"monitoring"`
	Storage               Storage       `yaml:"storage"`
	API                   API           `yaml:"api"`
	Development           Development   `yaml:"development"`
}

// Models configuration
type Models struct {
	AnomalyDetection  AnomalyDetection  `yaml:"anomaly_detection"`
	LoadPrediction    LoadPrediction    `yaml:"load_prediction"`
	RouteOptimization RouteOptimization `yaml:"route_optimization"`
}

// AnomalyDetection configuration
type AnomalyDetection struct {
	Enabled       bool    `yaml:"enabled"`
	Algorithm     string  `yaml:"algorithm"`
	Threshold     float64 `yaml:"threshold"`
	LearningRate  float64 `yaml:"learning_rate"`
	MaxSamples    int     `yaml:"max_samples"`
	Contamination float64 `yaml:"contamination"`
}

// LoadPrediction configuration
type LoadPrediction struct {
	Enabled         bool        `yaml:"enabled"`
	Algorithm       string      `yaml:"algorithm"`
	ARIMAParams     ARIMAParams `yaml:"arima_params"`
	ForecastHorizon int         `yaml:"forecast_horizon"`
	ConfidenceLevel float64     `yaml:"confidence_level"`
}

// ARIMAParams configuration
type ARIMAParams struct {
	P int `yaml:"p"` // autoregressive
	D int `yaml:"d"` // differencing
	Q int `yaml:"q"` // moving average
}

// RouteOptimization configuration
type RouteOptimization struct {
	Enabled         bool          `yaml:"enabled"`
	Algorithm       string        `yaml:"algorithm"`
	ExplorationRate float64       `yaml:"exploration_rate"`
	LearningRate    float64       `yaml:"learning_rate"`
	UpdateInterval  time.Duration `yaml:"update_interval"`
}

// DataPipeline configuration
type DataPipeline struct {
	Enabled            bool          `yaml:"enabled"`
	BatchSize          int           `yaml:"batch_size"`
	ProcessingInterval time.Duration `yaml:"processing_interval"`
	RetentionDays      int           `yaml:"retention_days"`
	Features           Features      `yaml:"features"`
}

// Features configuration
type Features struct {
	Network  []string `yaml:"network"`
	System   []string `yaml:"system"`
	Security []string `yaml:"security"`
}

// Monitoring configuration
type Monitoring struct {
	Enabled         bool            `yaml:"enabled"`
	MetricsInterval time.Duration   `yaml:"metrics_interval"`
	AlertThresholds AlertThresholds `yaml:"alert_thresholds"`
}

// AlertThresholds configuration
type AlertThresholds struct {
	AnomalyScore      float64       `yaml:"anomaly_score"`
	PredictionAccuracy float64      `yaml:"prediction_accuracy"`
	ProcessingLatency time.Duration `yaml:"processing_latency"`
}

// Storage configuration
type Storage struct {
	Type        string `yaml:"type"`
	Path        string `yaml:"path"`
	MaxSize     string `yaml:"max_size"`
	Compression bool   `yaml:"compression"`
}

// API configuration
type API struct {
	Enabled    bool `yaml:"enabled"`
	Port       int  `yaml:"port"`
	CORSEnabled bool `yaml:"cors_enabled"`
	RateLimit  int  `yaml:"rate_limit"`
}

// Development configuration
type Development struct {
	Debug     bool `yaml:"debug"`
	MockData  bool `yaml:"mock_data"`
	TestMode  bool `yaml:"test_mode"`
}

// LoadConfig loads AI configuration from file
func LoadConfig(configPath string) (*Config, error) {
	if configPath == "" {
		configPath = "config/ai-config.yaml"
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if !c.AI.Enabled {
		return fmt.Errorf("AI is disabled")
	}

	if c.AI.Mode != "cpu-only" && c.AI.Mode != "gpu" && c.AI.Mode != "hybrid" {
		return fmt.Errorf("invalid AI mode: %s", c.AI.Mode)
	}

	if c.AI.Models.AnomalyDetection.Threshold < 0 || c.AI.Models.AnomalyDetection.Threshold > 1 {
		return fmt.Errorf("anomaly detection threshold must be between 0 and 1")
	}

	if c.AI.Models.LoadPrediction.ConfidenceLevel < 0 || c.AI.Models.LoadPrediction.ConfidenceLevel > 1 {
		return fmt.Errorf("load prediction confidence level must be between 0 and 1")
	}

	if c.AI.API.Port <= 0 || c.AI.API.Port > 65535 {
		return fmt.Errorf("invalid API port: %d", c.AI.API.Port)
	}

	return nil
}

// GetDefaultConfig returns default configuration
func GetDefaultConfig() *Config {
	return &Config{
		AI: AI{
			Enabled:               true,
			Mode:                  "cpu-only",
			LogLevel:              "info",
			MaxConcurrentRequests: 100,
			RequestTimeout:        30 * time.Second,
			CacheTTL:              5 * time.Minute,
			Models: Models{
				AnomalyDetection: AnomalyDetection{
					Enabled:       true,
					Algorithm:     "isolation_forest",
					Threshold:     0.8,
					LearningRate:  0.01,
					MaxSamples:    10000,
					Contamination: 0.1,
				},
				LoadPrediction: LoadPrediction{
					Enabled:   true,
					Algorithm: "arima",
					ARIMAParams: ARIMAParams{
						P: 2,
						D: 1,
						Q: 2,
					},
					ForecastHorizon: 24,
					ConfidenceLevel: 0.8,
				},
				RouteOptimization: RouteOptimization{
					Enabled:         true,
					Algorithm:       "multi_armed_bandit",
					ExplorationRate: 0.1,
					LearningRate:    0.05,
					UpdateInterval:  time.Minute,
				},
			},
			DataPipeline: DataPipeline{
				Enabled:            true,
				BatchSize:          1000,
				ProcessingInterval: 10 * time.Second,
				RetentionDays:      30,
				Features: Features{
					Network:  []string{"bandwidth_utilization", "latency", "packet_loss", "connection_count", "protocol_distribution"},
					System:   []string{"cpu_usage", "memory_usage", "disk_io", "network_io"},
					Security: []string{"failed_connections", "suspicious_patterns", "rate_limiting", "authentication_failures"},
				},
			},
			Monitoring: Monitoring{
				Enabled:         true,
				MetricsInterval: 30 * time.Second,
				AlertThresholds: AlertThresholds{
					AnomalyScore:       0.9,
					PredictionAccuracy: 0.7,
					ProcessingLatency:  100 * time.Millisecond,
				},
			},
			Storage: Storage{
				Type:        "file",
				Path:        "/data/ai",
				MaxSize:     "1GB",
				Compression: true,
			},
			API: API{
				Enabled:     true,
				Port:        8083,
				CORSEnabled: true,
				RateLimit:   1000,
			},
			Development: Development{
				Debug:    false,
				MockData: false,
				TestMode: false,
			},
		},
	}
}

