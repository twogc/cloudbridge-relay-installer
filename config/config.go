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

// Config holds the server configuration
type Config struct {
	// Server settings
	Address string `json:"address"`
	Timeout time.Duration `json:"timeout"`

	// Tunnel settings
	MaxTunnels int `json:"max_tunnels"`
	PingInterval time.Duration `json:"ping_interval"`
	CleanupInterval time.Duration `json:"cleanup_interval"`

	// Token validation settings
	TokenValidation struct {
		RESTEnabled     bool          `json:"rest_enabled"`
		RESTEndpoint    string        `json:"rest_endpoint"`
		RESTTimeout     time.Duration `json:"rest_timeout"`
		RESTCacheTTL    time.Duration `json:"rest_cache_ttl"`
		RESTMaxRetries  int          `json:"rest_max_retries"`
		RESTRetryDelay  time.Duration `json:"rest_retry_delay"`

		JWTEnabled      bool          `json:"jwt_enabled"`
		JWTSigningKey   string        `json:"jwt_signing_key"`
		JWTIssuer       string        `json:"jwt_issuer"`
		JWTMaxAge       time.Duration `json:"jwt_max_age"`
	} `json:"token_validation"`

	// Logging settings
	Debug bool `json:"debug"`
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
	if config.Address == "" {
		config.Address = ":8080"
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.MaxTunnels == 0 {
		config.MaxTunnels = 1000
	}
	if config.PingInterval == 0 {
		config.PingInterval = 30 * time.Second
	}
	if config.CleanupInterval == 0 {
		config.CleanupInterval = 1 * time.Minute
	}

	return &config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.MaxTunnels < 0 {
		return fmt.Errorf("max_tunnels must be positive")
	}
	if c.PingInterval < time.Second {
		return fmt.Errorf("ping_interval must be at least 1 second")
	}
	if c.CleanupInterval < time.Second {
		return fmt.Errorf("cleanup_interval must be at least 1 second")
	}

	if c.TokenValidation.RESTEnabled {
		if c.TokenValidation.RESTEndpoint == "" {
			return fmt.Errorf("rest_endpoint is required when rest_enabled is true")
		}
		if c.TokenValidation.RESTTimeout < time.Second {
			return fmt.Errorf("rest_timeout must be at least 1 second")
		}
		if c.TokenValidation.RESTCacheTTL < time.Second {
			return fmt.Errorf("rest_cache_ttl must be at least 1 second")
		}
		if c.TokenValidation.RESTMaxRetries < 0 {
			return fmt.Errorf("rest_max_retries must be non-negative")
		}
		if c.TokenValidation.RESTRetryDelay < 0 {
			return fmt.Errorf("rest_retry_delay must be non-negative")
		}
	}

	if c.TokenValidation.JWTEnabled {
		if c.TokenValidation.JWTSigningKey == "" {
			return fmt.Errorf("jwt_signing_key is required when jwt_enabled is true")
		}
		if c.TokenValidation.JWTMaxAge < time.Second {
			return fmt.Errorf("jwt_max_age must be at least 1 second")
		}
	}

	return nil
} 