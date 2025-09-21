package ai

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetDefaultConfig(t *testing.T) {
	config := GetDefaultConfig()

	assert.NotNil(t, config)
	assert.True(t, config.AI.Enabled)
	assert.Equal(t, "cpu-only", config.AI.Mode)
	assert.Equal(t, "info", config.AI.LogLevel)
	assert.Equal(t, 100, config.AI.MaxConcurrentRequests)
	assert.Equal(t, 30*time.Second, config.AI.RequestTimeout)
	assert.Equal(t, 5*time.Minute, config.AI.CacheTTL)

	// Test Models configuration
	assert.True(t, config.AI.Models.AnomalyDetection.Enabled)
	assert.Equal(t, "isolation_forest", config.AI.Models.AnomalyDetection.Algorithm)
	assert.Equal(t, 0.8, config.AI.Models.AnomalyDetection.Threshold)
	assert.Equal(t, 0.01, config.AI.Models.AnomalyDetection.LearningRate)
	assert.Equal(t, 10000, config.AI.Models.AnomalyDetection.MaxSamples)
	assert.Equal(t, 0.1, config.AI.Models.AnomalyDetection.Contamination)

	assert.True(t, config.AI.Models.LoadPrediction.Enabled)
	assert.Equal(t, "arima", config.AI.Models.LoadPrediction.Algorithm)
	assert.Equal(t, 2, config.AI.Models.LoadPrediction.ARIMAParams.P)
	assert.Equal(t, 1, config.AI.Models.LoadPrediction.ARIMAParams.D)
	assert.Equal(t, 2, config.AI.Models.LoadPrediction.ARIMAParams.Q)
	assert.Equal(t, 24, config.AI.Models.LoadPrediction.ForecastHorizon)
	assert.Equal(t, 0.8, config.AI.Models.LoadPrediction.ConfidenceLevel)

	assert.True(t, config.AI.Models.RouteOptimization.Enabled)
	assert.Equal(t, "multi_armed_bandit", config.AI.Models.RouteOptimization.Algorithm)
	assert.Equal(t, 0.1, config.AI.Models.RouteOptimization.ExplorationRate)
	assert.Equal(t, 0.05, config.AI.Models.RouteOptimization.LearningRate)
	assert.Equal(t, time.Minute, config.AI.Models.RouteOptimization.UpdateInterval)

	// Test DataPipeline configuration
	assert.True(t, config.AI.DataPipeline.Enabled)
	assert.Equal(t, 1000, config.AI.DataPipeline.BatchSize)
	assert.Equal(t, 60*time.Second, config.AI.DataPipeline.ProcessingInterval)
	assert.Equal(t, 30, config.AI.DataPipeline.RetentionDays)
	assert.NotEmpty(t, config.AI.DataPipeline.Features.Network)
	assert.NotEmpty(t, config.AI.DataPipeline.Features.System)
	assert.NotEmpty(t, config.AI.DataPipeline.Features.Security)

	// Test Monitoring configuration
	assert.True(t, config.AI.Monitoring.Enabled)
	assert.Equal(t, 30*time.Second, config.AI.Monitoring.MetricsInterval)
	assert.Equal(t, 0.9, config.AI.Monitoring.AlertThresholds.AnomalyScore)
	assert.Equal(t, 0.7, config.AI.Monitoring.AlertThresholds.PredictionAccuracy)
	assert.Equal(t, 100*time.Millisecond, config.AI.Monitoring.AlertThresholds.ProcessingLatency)

	// Test Storage configuration
	assert.Equal(t, "file", config.AI.Storage.Type)
	assert.Equal(t, "/data/ai", config.AI.Storage.Path)
	assert.Equal(t, "1GB", config.AI.Storage.MaxSize)
	assert.True(t, config.AI.Storage.Compression)

	// Test API configuration
	assert.True(t, config.AI.API.Enabled)
	assert.Equal(t, 8083, config.AI.API.Port)
	assert.True(t, config.AI.API.CORSEnabled)
	assert.Equal(t, 1000, config.AI.API.RateLimit)

	// Test Development configuration
	assert.False(t, config.AI.Development.Debug)
	assert.False(t, config.AI.Development.MockData)
	assert.False(t, config.AI.Development.TestMode)
}

func TestConfig_Validate_ValidConfig(t *testing.T) {
	config := GetDefaultConfig()

	err := config.Validate()
	assert.NoError(t, err)
}

func TestConfig_Validate_DisabledAI(t *testing.T) {
	config := GetDefaultConfig()
	config.AI.Enabled = false

	err := config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "AI is disabled")
}

func TestConfig_Validate_InvalidMode(t *testing.T) {
	config := GetDefaultConfig()
	config.AI.Mode = "invalid-mode"

	err := config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid AI mode")
}

func TestConfig_Validate_ValidModes(t *testing.T) {
	validModes := []string{"cpu-only", "gpu", "hybrid", "kubernetes"}

	for _, mode := range validModes {
		config := GetDefaultConfig()
		config.AI.Mode = mode

		err := config.Validate()
		assert.NoError(t, err, "Mode %s should be valid", mode)
	}
}

func TestConfig_Validate_InvalidAnomalyThreshold(t *testing.T) {
	config := GetDefaultConfig()

	// Test negative threshold
	config.AI.Models.AnomalyDetection.Threshold = -0.1
	err := config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "anomaly detection threshold must be between 0 and 1")

	// Test threshold > 1
	config.AI.Models.AnomalyDetection.Threshold = 1.1
	err = config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "anomaly detection threshold must be between 0 and 1")

	// Test valid threshold
	config.AI.Models.AnomalyDetection.Threshold = 0.5
	err = config.Validate()
	assert.NoError(t, err)
}

func TestConfig_Validate_InvalidConfidenceLevel(t *testing.T) {
	config := GetDefaultConfig()

	// Test negative confidence level
	config.AI.Models.LoadPrediction.ConfidenceLevel = -0.1
	err := config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "load prediction confidence level must be between 0 and 1")

	// Test confidence level > 1
	config.AI.Models.LoadPrediction.ConfidenceLevel = 1.1
	err = config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "load prediction confidence level must be between 0 and 1")

	// Test valid confidence level
	config.AI.Models.LoadPrediction.ConfidenceLevel = 0.5
	err = config.Validate()
	assert.NoError(t, err)
}

func TestConfig_Validate_InvalidAPIPort(t *testing.T) {
	config := GetDefaultConfig()

	// Test port <= 0
	config.AI.API.Port = 0
	err := config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid API port")

	// Test port > 65535
	config.AI.API.Port = 65536
	err = config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid API port")

	// Test valid port
	config.AI.API.Port = 8080
	err = config.Validate()
	assert.NoError(t, err)
}

func TestLoadConfig_DefaultPath(t *testing.T) {
	// Test with empty path (should use default)
	config, err := LoadConfig("")
	assert.Error(t, err) // Should fail because default file doesn't exist
	assert.Nil(t, config)
}

func TestLoadConfig_ValidFile(t *testing.T) {
	// Create temporary config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test-config.yaml")

	configContent := `
ai:
  enabled: true
  mode: "cpu-only"
  log_level: "debug"
  max_concurrent_requests: 50
  request_timeout: "15s"
  cache_ttl: "2m"
  models:
    anomaly_detection:
      enabled: true
      algorithm: "isolation_forest"
      threshold: 0.7
      learning_rate: 0.02
      max_samples: 5000
      contamination: 0.05
    load_prediction:
      enabled: true
      algorithm: "arima"
      arima_params:
        p: 1
        d: 1
        q: 1
      forecast_horizon: 12
      confidence_level: 0.75
    route_optimization:
      enabled: true
      algorithm: "multi_armed_bandit"
      exploration_rate: 0.15
      learning_rate: 0.03
      update_interval: "30s"
  data_pipeline:
    enabled: true
    batch_size: 500
    processing_interval: "30s"
    retention_days: 15
    features:
      network: ["bandwidth", "latency"]
      system: ["cpu", "memory"]
      security: ["failures"]
  monitoring:
    enabled: true
    metrics_interval: "15s"
    alert_thresholds:
      anomaly_score: 0.8
      prediction_accuracy: 0.6
      processing_latency: "50ms"
  storage:
    type: "file"
    path: "/tmp/ai"
    max_size: "500MB"
    compression: false
  api:
    enabled: true
    port: 8080
    cors_enabled: false
    rate_limit: 500
  development:
    debug: true
    mock_data: true
    test_mode: true
`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	config, err := LoadConfig(configFile)
	require.NoError(t, err)
	assert.NotNil(t, config)

	// Verify loaded values
	assert.True(t, config.AI.Enabled)
	assert.Equal(t, "cpu-only", config.AI.Mode)
	assert.Equal(t, "debug", config.AI.LogLevel)
	assert.Equal(t, 50, config.AI.MaxConcurrentRequests)
	assert.Equal(t, 15*time.Second, config.AI.RequestTimeout)
	assert.Equal(t, 2*time.Minute, config.AI.CacheTTL)

	// Verify models
	assert.True(t, config.AI.Models.AnomalyDetection.Enabled)
	assert.Equal(t, 0.7, config.AI.Models.AnomalyDetection.Threshold)
	assert.Equal(t, 0.02, config.AI.Models.AnomalyDetection.LearningRate)

	assert.True(t, config.AI.Models.LoadPrediction.Enabled)
	assert.Equal(t, 1, config.AI.Models.LoadPrediction.ARIMAParams.P)
	assert.Equal(t, 12, config.AI.Models.LoadPrediction.ForecastHorizon)

	// Verify API
	assert.Equal(t, 8080, config.AI.API.Port)
	assert.False(t, config.AI.API.CORSEnabled)

	// Verify development
	assert.True(t, config.AI.Development.Debug)
	assert.True(t, config.AI.Development.MockData)
	assert.True(t, config.AI.Development.TestMode)
}

func TestLoadConfig_InvalidFile(t *testing.T) {
	// Test with non-existent file
	config, err := LoadConfig("/non/existent/file.yaml")
	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "failed to read config file")
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	// Create temporary config file with invalid YAML
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "invalid-config.yaml")

	invalidContent := `
ai:
  enabled: true
  mode: "cpu-only"
  invalid_yaml: [unclosed
`

	err := os.WriteFile(configFile, []byte(invalidContent), 0644)
	require.NoError(t, err)

	config, err := LoadConfig(configFile)
	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "failed to unmarshal config")
}

func TestLoadConfig_InvalidConfig(t *testing.T) {
	// Create temporary config file with invalid configuration
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "invalid-config.yaml")

	invalidConfigContent := `
ai:
  enabled: false
  mode: "invalid-mode"
  models:
    anomaly_detection:
      threshold: -1.0
  api:
    port: 0
`

	err := os.WriteFile(configFile, []byte(invalidConfigContent), 0644)
	require.NoError(t, err)

	config, err := LoadConfig(configFile)
	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "invalid configuration")
}

func TestLoadConfig_PathTraversal(t *testing.T) {
	// Test path traversal protection
	config, err := LoadConfig("../../../etc/passwd")
	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "invalid config path")
}

func TestLoadConfig_AbsolutePath(t *testing.T) {
	// Create temporary config file with absolute path
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test-config.yaml")

	configContent := `
ai:
  enabled: true
  mode: "cpu-only"
  api:
    port: 8080
`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	// Test with absolute path
	config, err := LoadConfig(configFile)
	require.NoError(t, err)
	assert.NotNil(t, config)
	assert.True(t, config.AI.Enabled)
}

func TestConfig_EdgeCases(t *testing.T) {
	// Test with minimal valid config
	config := &Config{
		AI: AI{
			Enabled: true,
			Mode:    "cpu-only",
			Models: Models{
				AnomalyDetection: AnomalyDetection{
					Threshold: 0.5,
				},
				LoadPrediction: LoadPrediction{
					ConfidenceLevel: 0.5,
				},
			},
			API: API{
				Port: 8080,
			},
		},
	}

	err := config.Validate()
	assert.NoError(t, err)
}

func TestConfig_ZeroValues(t *testing.T) {
	// Test with zero values
	config := &Config{
		AI: AI{
			Enabled: true,
			Mode:    "cpu-only",
			Models: Models{
				AnomalyDetection: AnomalyDetection{
					Threshold: 0.0, // Zero threshold
				},
				LoadPrediction: LoadPrediction{
					ConfidenceLevel: 0.0, // Zero confidence
				},
			},
			API: API{
				Port: 1, // Minimum valid port
			},
		},
	}

	err := config.Validate()
	assert.NoError(t, err)
}

func TestConfig_MaximumValues(t *testing.T) {
	// Test with maximum valid values
	config := &Config{
		AI: AI{
			Enabled: true,
			Mode:    "cpu-only",
			Models: Models{
				AnomalyDetection: AnomalyDetection{
					Threshold: 1.0, // Maximum threshold
				},
				LoadPrediction: LoadPrediction{
					ConfidenceLevel: 1.0, // Maximum confidence
				},
			},
			API: API{
				Port: 65535, // Maximum valid port
			},
		},
	}

	err := config.Validate()
	assert.NoError(t, err)
}

func TestConfig_TimeDurations(t *testing.T) {
	// Test time duration parsing
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "time-config.yaml")

	configContent := `
ai:
  enabled: true
  mode: "cpu-only"
  request_timeout: "1h"
  cache_ttl: "30m"
  models:
    route_optimization:
      update_interval: "5m"
  data_pipeline:
    processing_interval: "2m"
  monitoring:
    metrics_interval: "10s"
    alert_thresholds:
      processing_latency: "1s"
  api:
    port: 8080
`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	config, err := LoadConfig(configFile)
	require.NoError(t, err)
	assert.NotNil(t, config)

	assert.Equal(t, time.Hour, config.AI.RequestTimeout)
	assert.Equal(t, 30*time.Minute, config.AI.CacheTTL)
	assert.Equal(t, 5*time.Minute, config.AI.Models.RouteOptimization.UpdateInterval)
	assert.Equal(t, 2*time.Minute, config.AI.DataPipeline.ProcessingInterval)
	assert.Equal(t, 10*time.Second, config.AI.Monitoring.MetricsInterval)
	assert.Equal(t, time.Second, config.AI.Monitoring.AlertThresholds.ProcessingLatency)
}

func TestConfig_FeatureLists(t *testing.T) {
	// Test feature lists
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "features-config.yaml")

	configContent := `
ai:
  enabled: true
  mode: "cpu-only"
  data_pipeline:
    features:
      network: ["bandwidth", "latency", "packet_loss", "connections"]
      system: ["cpu", "memory", "disk", "network"]
      security: ["failures", "patterns", "rate_limiting", "auth"]
  api:
    port: 8080
`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	config, err := LoadConfig(configFile)
	require.NoError(t, err)
	assert.NotNil(t, config)

	assert.Len(t, config.AI.DataPipeline.Features.Network, 4)
	assert.Len(t, config.AI.DataPipeline.Features.System, 4)
	assert.Len(t, config.AI.DataPipeline.Features.Security, 4)

	assert.Contains(t, config.AI.DataPipeline.Features.Network, "bandwidth")
	assert.Contains(t, config.AI.DataPipeline.Features.System, "cpu")
	assert.Contains(t, config.AI.DataPipeline.Features.Security, "failures")
}

func TestConfig_AllModes(t *testing.T) {
	modes := []string{"cpu-only", "gpu", "hybrid", "kubernetes"}

	for _, mode := range modes {
		config := GetDefaultConfig()
		config.AI.Mode = mode

		err := config.Validate()
		assert.NoError(t, err, "Mode %s should be valid", mode)
	}
}

func TestConfig_DisabledComponents(t *testing.T) {
	// Test with all components disabled
	config := GetDefaultConfig()
	config.AI.Models.AnomalyDetection.Enabled = false
	config.AI.Models.LoadPrediction.Enabled = false
	config.AI.Models.RouteOptimization.Enabled = false
	config.AI.DataPipeline.Enabled = false
	config.AI.Monitoring.Enabled = false
	config.AI.API.Enabled = false

	err := config.Validate()
	assert.NoError(t, err)
}

func TestConfig_FilePermissions(t *testing.T) {
	// Test with file that has no read permissions
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "no-read.yaml")

	configContent := `
ai:
  enabled: true
  mode: "cpu-only"
  api:
    port: 8080
`

	err := os.WriteFile(configFile, []byte(configContent), 0000) // No permissions
	require.NoError(t, err)

	config, err := LoadConfig(configFile)
	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "failed to read config file")
}
