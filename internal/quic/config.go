package quic

import (
	"crypto/tls"
	"time"
)

// Config holds QUIC configuration
type Config struct {
	// Connection settings
	MaxStreams        int           `json:"max_streams"`
	KeepAlive         bool          `json:"keep_alive"`
	IdleTimeout       time.Duration `json:"idle_timeout"`
	HandshakeTimeout  time.Duration `json:"handshake_timeout"`
	
	// Performance settings
	MaxIncomingStreams int           `json:"max_incoming_streams"`
	MaxIncomingUniStreams int        `json:"max_incoming_uni_streams"`
	
	// TLS settings
	TLSConfig         *tls.Config   `json:"-"`
	InsecureSkipVerify bool         `json:"insecure_skip_verify"`
	
	// Flow control
	InitialStreamReceiveWindow     uint64 `json:"initial_stream_receive_window"`
	MaxStreamReceiveWindow         uint64 `json:"max_stream_receive_window"`
	InitialConnectionReceiveWindow uint64 `json:"initial_connection_receive_window"`
	MaxConnectionReceiveWindow     uint64 `json:"max_connection_receive_window"`
	
	// Congestion control
	CongestionControlAlgorithm string `json:"congestion_control_algorithm"`
	
	// Debug settings
	EnableDebugLogs bool `json:"enable_debug_logs"`
}

// DefaultConfig returns default QUIC configuration
func DefaultConfig() *Config {
	return &Config{
		MaxStreams:        100,
		KeepAlive:         true,
		IdleTimeout:       30 * time.Second,
		HandshakeTimeout:  10 * time.Second,
		MaxIncomingStreams: 100,
		MaxIncomingUniStreams: 100,
		InsecureSkipVerify: false,
		InitialStreamReceiveWindow: 512 * 1024, // 512KB
		MaxStreamReceiveWindow:     2 * 1024 * 1024, // 2MB
		InitialConnectionReceiveWindow: 1024 * 1024, // 1MB
		MaxConnectionReceiveWindow:     5 * 1024 * 1024, // 5MB
		CongestionControlAlgorithm: "cubic",
		EnableDebugLogs: false,
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