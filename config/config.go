package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"
)

var (
	// Security settings
	AuthToken = flag.String("auth-token", "", "Authentication token for initial connection")
	AllowIPs  = flag.String("allow-ips", "", "Comma-separated list of allowed IP addresses (empty for all)")
	
	// TLS settings
	UseTLS     = flag.Bool("tls", false, "Enable TLS encryption")
	CertFile   = flag.String("cert", "/etc/relay/certs/fullchain.pem", "TLS certificate file")
	KeyFile    = flag.String("key", "/etc/relay/certs/privkey.pem", "TLS key file")
	Domain     = flag.String("domain", "", "Domain name for Let's Encrypt certificate")
)

// Config represents the server configuration
type Config struct {
	Server struct {
		Address         string        `json:"address"`
		Port           int           `json:"port"`
		Timeout        time.Duration `json:"timeout"`
		MaxTunnels     int           `json:"max_tunnels"`
		PingInterval   time.Duration `json:"ping_interval"`
	CleanupInterval time.Duration `json:"cleanup_interval"`
	} `json:"server"`

	Django struct {
		APIEndpoint    string        `json:"api_endpoint"`
		JWTSecret      string        `json:"jwt_secret"`
		JWTExpiry      time.Duration `json:"jwt_expiry"`
		RESTTimeout    time.Duration `json:"rest_timeout"`
		RESTCacheTTL   time.Duration `json:"rest_cache_ttl"`
		RESTMaxRetries int           `json:"rest_max_retries"`
		RESTRetryDelay time.Duration `json:"rest_retry_delay"`
	} `json:"django"`

	Security struct {
		TLSEnabled bool   `json:"tls_enabled"`
		CertFile   string `json:"cert_file"`
		KeyFile    string `json:"key_file"`
		AllowedIPs string `json:"allowed_ips"`
		RateLimit  struct {
			Enabled          bool    `json:"enabled"`
			RequestsPerSecond float64 `json:"requests_per_second"`
			Burst           int     `json:"burst"`
		} `json:"rate_limit"`
	} `json:"security"`

	Metrics struct {
		Enabled bool   `json:"enabled"`
		Port    string `json:"port"`
		Path    string `json:"path"`
	} `json:"metrics"`

	Logging struct {
		Level  string `json:"level"`
		Format string `json:"format"`
		Output string `json:"output"`
	} `json:"logging"`
}

// Load loads configuration from a file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	// Set default values
	if config.Server.Address == "" {
		config.Server.Address = ":3456"
	}
	if config.Server.Timeout == 0 {
		config.Server.Timeout = 30 * time.Second
	}
	if config.Server.MaxTunnels == 0 {
		config.Server.MaxTunnels = 1000
	}
	if config.Server.PingInterval == 0 {
		config.Server.PingInterval = 30 * time.Second
	}
	if config.Server.CleanupInterval == 0 {
		config.Server.CleanupInterval = 1 * time.Minute
	}

	if config.Django.RESTTimeout == 0 {
		config.Django.RESTTimeout = 5 * time.Second
	}
	if config.Django.RESTCacheTTL == 0 {
		config.Django.RESTCacheTTL = 5 * time.Minute
	}
	if config.Django.RESTMaxRetries == 0 {
		config.Django.RESTMaxRetries = 3
	}
	if config.Django.RESTRetryDelay == 0 {
		config.Django.RESTRetryDelay = 100 * time.Millisecond
	}

	return &config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.MaxTunnels < 0 {
		return fmt.Errorf("max_tunnels must be positive")
	}
	if c.Server.PingInterval < time.Second {
		return fmt.Errorf("ping_interval must be at least 1 second")
	}
	if c.Server.CleanupInterval < time.Second {
		return fmt.Errorf("cleanup_interval must be at least 1 second")
	}

	if c.Django.APIEndpoint == "" {
		return fmt.Errorf("django.api_endpoint is required")
	}
	if c.Django.RESTTimeout < time.Second {
		return fmt.Errorf("rest_timeout must be at least 1 second")
	}
	if c.Django.RESTCacheTTL < time.Second {
		return fmt.Errorf("rest_cache_ttl must be at least 1 second")
	}
	if c.Django.RESTMaxRetries < 0 {
		return fmt.Errorf("rest_max_retries must be non-negative")
	}
	if c.Django.RESTRetryDelay < 0 {
		return fmt.Errorf("rest_retry_delay must be non-negative")
	}

	if c.Security.TLSEnabled {
		if c.Security.CertFile == "" {
			return fmt.Errorf("cert_file is required when tls_enabled is true")
		}
		if c.Security.KeyFile == "" {
			return fmt.Errorf("key_file is required when tls_enabled is true")
		}
	}

	return nil
} 