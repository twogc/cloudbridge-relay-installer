package quic

import (
	"crypto/tls"
	"time"
)

// Config holds QUIC configuration
type Config struct {
	// Connection settings
	MaxStreams       int           `json:"max_streams"`
	KeepAlive        bool          `json:"keep_alive"`
	IdleTimeout      time.Duration `json:"idle_timeout"`
	HandshakeTimeout time.Duration `json:"handshake_timeout"`

	// Performance settings
	MaxIncomingStreams    int `json:"max_incoming_streams"`
	MaxIncomingUniStreams int `json:"max_incoming_uni_streams"`

	// TLS settings
	TLSConfig          *tls.Config `json:"-"`
	InsecureSkipVerify bool        `json:"insecure_skip_verify"`

	// Flow control
	InitialStreamReceiveWindow     uint64 `json:"initial_stream_receive_window"`
	MaxStreamReceiveWindow         uint64 `json:"max_stream_receive_window"`
	InitialConnectionReceiveWindow uint64 `json:"initial_connection_receive_window"`
	MaxConnectionReceiveWindow     uint64 `json:"max_connection_receive_window"`

	// Congestion control
	CongestionControlAlgorithm string `json:"congestion_control_algorithm"`

	// Debug settings
	EnableDebugLogs bool `json:"enable_debug_logs"`

	// Enhanced QUIC settings
	EnableEnhancedQUIC bool            `json:"enable_enhanced_quic"`
	EnhancedConfig     *EnhancedConfig `json:"enhanced_config,omitempty"`

	// QLog tracing settings
	EnableQLogTracing bool        `json:"enable_qlog_tracing"`
	QLogConfig        *QLogConfig `json:"qlog_config,omitempty"`

	// Experimental features (integrated from quic-test)
	EnableExperimental   bool                       `json:"enable_experimental"`
	ExperimentalFeatures *ExperimentalFeaturesConfig `json:"experimental_features,omitempty"`
}

// ExperimentalFeaturesConfig holds experimental QUIC features configuration
type ExperimentalFeaturesConfig struct {
	// BBRv2 congestion control
	EnableBBRv2           bool   `json:"enable_bbrv2"`
	CongestionControl     string `json:"congestion_control"` // cubic, bbr, bbrv2, reno

	// ACK Frequency optimization
	EnableACKFrequency    bool                  `json:"enable_ack_frequency"`
	ACKFrequencyConfig    *ACKFrequencyConfig   `json:"ack_frequency_config,omitempty"`

	// Forward Error Correction (TODO)
	EnableFEC             bool   `json:"enable_fec"`
	FECRedundancy         float64 `json:"fec_redundancy"` // 0.1 = 10% redundancy

	// QUIC Bit Greasing (RFC 9287)
	EnableGreasing        bool   `json:"enable_greasing"`

	// Advanced performance features
	EnableGSO             bool   `json:"enable_gso"` // UDP GSO
	EnableGRO             bool   `json:"enable_gro"` // UDP GRO
	SocketBufferSize      int    `json:"socket_buffer_size"`
	EnableNUMA            bool   `json:"enable_numa"` // NUMA pinning
}

// DefaultConfig returns default QUIC configuration
func DefaultConfig() *Config {
	return &Config{
		MaxStreams:                     100,
		KeepAlive:                      true,
		IdleTimeout:                    30 * time.Second,
		HandshakeTimeout:               10 * time.Second,
		MaxIncomingStreams:             100,
		MaxIncomingUniStreams:          100,
		InsecureSkipVerify:             false,
		InitialStreamReceiveWindow:     512 * 1024,      // 512KB
		MaxStreamReceiveWindow:         2 * 1024 * 1024, // 2MB
		InitialConnectionReceiveWindow: 1024 * 1024,     // 1MB
		MaxConnectionReceiveWindow:     5 * 1024 * 1024, // 5MB
		CongestionControlAlgorithm:     "cubic",
		EnableDebugLogs:                false,
		EnableEnhancedQUIC:             false, // По умолчанию отключено
		EnableQLogTracing:              false, // По умолчанию отключено
		QLogConfig:                     DefaultQLogConfig(),
	}
}

// DefaultEnhancedConfig возвращает конфигурацию Enhanced QUIC по умолчанию
func DefaultEnhancedConfig() *EnhancedConfig {
	return &EnhancedConfig{
		MaxStreams:                     100,
		KeepAlive:                      true,
		IdleTimeout:                    30 * time.Second,
		HandshakeTimeout:               10 * time.Second,
		MaxIncomingStreams:             100,
		MaxIncomingUniStreams:          100,
		InitialStreamReceiveWindow:     512 * 1024,      // 512KB
		MaxStreamReceiveWindow:         2 * 1024 * 1024, // 2MB
		InitialConnectionReceiveWindow: 1024 * 1024,     // 1MB
		MaxConnectionReceiveWindow:     5 * 1024 * 1024, // 5MB
		CongestionControlAlgorithm:     "cubic",

		// Enhanced features
		TenantManager:             NewTenantManager(),
		EnableDDoSProtection:      true,
		EnableGeographicRouting:   true,
		EnableKeycloakIntegration: false, // Отключено по умолчанию
	}
}

// Validate validates the QUIC configuration
func (c *Config) Validate() error {
	if c.MaxStreams <= 0 {
		c.MaxStreams = 100
	}

	if c.IdleTimeout <= 0 {
		c.IdleTimeout = 30 * time.Second
	}

	if c.HandshakeTimeout <= 0 {
		c.HandshakeTimeout = 10 * time.Second
	}

	if c.MaxIncomingStreams <= 0 {
		c.MaxIncomingStreams = 100
	}

	if c.MaxIncomingUniStreams <= 0 {
		c.MaxIncomingUniStreams = 100
	}

	if c.InitialStreamReceiveWindow == 0 {
		c.InitialStreamReceiveWindow = 512 * 1024
	}

	if c.MaxStreamReceiveWindow == 0 {
		c.MaxStreamReceiveWindow = 2 * 1024 * 1024
	}

	if c.InitialConnectionReceiveWindow == 0 {
		c.InitialConnectionReceiveWindow = 1024 * 1024
	}

	if c.MaxConnectionReceiveWindow == 0 {
		c.MaxConnectionReceiveWindow = 5 * 1024 * 1024
	}

	return nil
}

// GetTLSConfig returns TLS configuration for QUIC
func (c *Config) GetTLSConfig() *tls.Config {
	if c.TLSConfig != nil {
		return c.TLSConfig
	}

	config := &tls.Config{
		// #nosec G402 -- InsecureSkipVerify is configurable for development/testing
		// In production, this should be set to false and proper certificates should be used
		InsecureSkipVerify: c.InsecureSkipVerify,
		NextProtos:         []string{"h3", "h3-29", "h3-28", "h3-27"},
	}

	// Security warning for development/testing only
	if c.InsecureSkipVerify {
		// In production, this should be set to false and proper certificates should be used
	}

	return config
}

// CreateEnhancedConfig создает Enhanced конфигурацию на основе обычной
func (c *Config) CreateEnhancedConfig() *EnhancedConfig {
	return &EnhancedConfig{
		MaxStreams:                     c.MaxStreams,
		KeepAlive:                      c.KeepAlive,
		IdleTimeout:                    c.IdleTimeout,
		HandshakeTimeout:               c.HandshakeTimeout,
		MaxIncomingStreams:             c.MaxIncomingStreams,
		MaxIncomingUniStreams:          c.MaxIncomingUniStreams,
		TLSConfig:                      c.TLSConfig,
		InitialStreamReceiveWindow:     c.InitialStreamReceiveWindow,
		MaxStreamReceiveWindow:         c.MaxStreamReceiveWindow,
		InitialConnectionReceiveWindow: c.InitialConnectionReceiveWindow,
		MaxConnectionReceiveWindow:     c.MaxConnectionReceiveWindow,
		CongestionControlAlgorithm:     c.CongestionControlAlgorithm,

		// Enhanced features (по умолчанию включены)
		TenantManager:             NewTenantManager(),
		EnableDDoSProtection:      true,
		EnableGeographicRouting:   true,
		EnableKeycloakIntegration: false,
	}
}

// EnableEnhanced включает Enhanced QUIC режим
func (c *Config) EnableEnhanced() {
	c.EnableEnhancedQUIC = true
	if c.EnhancedConfig == nil {
		c.EnhancedConfig = c.CreateEnhancedConfig()
	}
}

// DisableEnhanced отключает Enhanced QUIC режим
func (c *Config) DisableEnhanced() {
	c.EnableEnhancedQUIC = false
	c.EnhancedConfig = nil
}

// IsEnhancedEnabled проверяет, включен ли Enhanced QUIC
func (c *Config) IsEnhancedEnabled() bool {
	return c.EnableEnhancedQUIC && c.EnhancedConfig != nil
}

// DefaultExperimentalFeaturesConfig возвращает конфигурацию экспериментальных функций по умолчанию
func DefaultExperimentalFeaturesConfig() *ExperimentalFeaturesConfig {
	return &ExperimentalFeaturesConfig{
		// BBRv2 congestion control
		EnableBBRv2:       true,
		CongestionControl: "bbrv2", // Use BBRv2 by default

		// ACK Frequency optimization
		EnableACKFrequency: true,
		ACKFrequencyConfig: DefaultACKFrequencyConfig(),

		// FEC (disabled by default - adds overhead)
		EnableFEC:     false,
		FECRedundancy: 0.1, // 10% redundancy

		// QUIC Bit Greasing (RFC 9287)
		EnableGreasing: true,

		// Advanced performance features
		EnableGSO:        true,  // Enable if supported by OS
		EnableGRO:        true,  // Enable if supported by OS
		SocketBufferSize: 1024 * 1024, // 1MB socket buffer
		EnableNUMA:       false, // NUMA pinning disabled by default
	}
}

// EnableExperimentalFeatures включает экспериментальные функции
func (c *Config) EnableExperimentalFeatures() {
	c.EnableExperimental = true
	if c.ExperimentalFeatures == nil {
		c.ExperimentalFeatures = DefaultExperimentalFeaturesConfig()
	}

	// Update congestion control algorithm if BBRv2 is enabled
	if c.ExperimentalFeatures.EnableBBRv2 {
		c.CongestionControlAlgorithm = "bbrv2"
	}
}

// DisableExperimentalFeatures отключает экспериментальные функции
func (c *Config) DisableExperimentalFeatures() {
	c.EnableExperimental = false
	c.ExperimentalFeatures = nil
	c.CongestionControlAlgorithm = "cubic" // Reset to default
}

// IsExperimentalEnabled проверяет, включены ли экспериментальные функции
func (c *Config) IsExperimentalEnabled() bool {
	return c.EnableExperimental && c.ExperimentalFeatures != nil
}
