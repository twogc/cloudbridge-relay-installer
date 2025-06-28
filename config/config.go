package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"regexp"
	"time"
	"github.com/twogc/cloudbridge-relay-installer/internal/tls"
)

var (
	// Security settings
	AuthToken = flag.String("auth-token", "", "Authentication token for initial connection")
	AllowIPs  = flag.String("allow-ips", "", "Comma-separated list of allowed IP addresses (empty for all)")
	
	// TLS settings
	UseTLS     = flag.Bool("tls", true, "Enable TLS encryption")
	CertFile   = flag.String("cert", "/etc/relay/certs/fullchain.pem", "TLS certificate file")
	KeyFile    = flag.String("key", "/etc/relay/certs/privkey.pem", "TLS key file")
	Domain     = flag.String("domain", "", "Domain name for Let's Encrypt certificate")
)

// Промежуточные структуры для парсинга time.Duration из строк
type configRaw struct {
	Server struct {
		Address         string `json:"address"`
		Port           int    `json:"port"`
		Timeout        string `json:"timeout"`
		MaxTunnels     int    `json:"max_tunnels"`
		PingInterval   string `json:"ping_interval"`
		CleanupInterval string `json:"cleanup_interval"`
	} `json:"server"`

	Django struct {
		APIEndpoint    string `json:"api_endpoint"`
		JWTSecret      string `json:"jwt_secret"`
		JWTExpiry      string `json:"jwt_expiry"`
		RESTTimeout    string `json:"rest_timeout"`
		RESTCacheTTL   string `json:"rest_cache_ttl"`
		RESTMaxRetries int    `json:"rest_max_retries"`
		RESTRetryDelay string `json:"rest_retry_delay"`
	} `json:"django"`

	Keycloak struct {
		Enabled        bool   `json:"enabled"`
		ServerURL      string `json:"server_url"`
		Realm          string `json:"realm"`
		ClientID       string `json:"client_id"`
		ClientSecret   string `json:"client_secret"`
		PublicKey      string `json:"public_key"`
		TokenValidation struct {
			Issuer     string `json:"issuer"`
			Audience   string `json:"audience"`
			ClockSkew  string `json:"clock_skew"`
		} `json:"token_validation"`
		JWKSCacheTTL   string `json:"jwks_cache_ttl"`
		HTTPTimeout    string `json:"http_timeout"`
	} `json:"keycloak"`

	Security struct {
		TLSEnabled   bool     `json:"tls_enabled"`
		CertFile     string   `json:"cert_file"`
		KeyFile      string   `json:"key_file"`
		AllowedIPs   []string `json:"allowed_ips"`
		RateLimit    struct {
			Enabled           bool    `json:"enabled"`
			RequestsPerSecond float64 `json:"requests_per_second"`
			Burst             int     `json:"burst"`
		} `json:"rate_limit"`
		MTLS struct {
			Enabled     bool   `json:"enabled"`
			CACert      string `json:"ca_file"`
		} `json:"mtls"`
		CipherSuites   []string `json:"cipher_suites"`
		MinTLSVersion  string   `json:"min_tls_version"`
		Audit          struct {
			Enabled    bool   `json:"enabled"`
			LogFile    string `json:"log_file"`
		} `json:"audit"`
	} `json:"security"`

	Metrics struct {
		Enabled bool   `json:"enabled"`
		Port    int    `json:"port"`
		Path    string `json:"path"`
	} `json:"metrics"`

	Logging struct {
		Level      string `json:"level"`
		Format     string `json:"format"`
		Output     string `json:"output"`
		File       string `json:"file"`
		MaxSize    string `json:"max_size"`
		MaxAge     string `json:"max_age"`
		MaxBackups int    `json:"max_backups"`
	} `json:"logging"`
}

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

	// Новый блок конфигурации для Keycloak
	Keycloak struct {
		Enabled        bool   `json:"enabled"`
		ServerURL      string `json:"server_url"`
		Realm          string `json:"realm"`
		ClientID       string `json:"client_id"`
		ClientSecret   string `json:"client_secret"`
		PublicKey      string `json:"public_key"`
		TokenValidation struct {
			Issuer     string        `json:"issuer"`
			Audience   string        `json:"audience"`
			ClockSkew  time.Duration `json:"clock_skew"`
		} `json:"token_validation"`
		JWKSCacheTTL   time.Duration `json:"jwks_cache_ttl"`
		HTTPTimeout    time.Duration `json:"http_timeout"`
	} `json:"keycloak"`

	Security struct {
		TLSEnabled   bool     `json:"tls_enabled"`
		CertFile     string   `json:"cert_file"`
		KeyFile      string   `json:"key_file"`
		AllowedIPs   []string `json:"allowed_ips"`
		RateLimit    struct {
			Enabled           bool    `json:"enabled"`
			RequestsPerSecond float64 `json:"requests_per_second"`
			Burst             int     `json:"burst"`
		} `json:"rate_limit"`
		MTLS struct {
			Enabled     bool   `json:"enabled"`
			CACert      string `json:"ca_file"`
		} `json:"mtls"`
		CipherSuites   []string `json:"cipher_suites"`
		MinTLSVersion  string   `json:"min_tls_version"`
		Audit          struct {
			Enabled    bool   `json:"enabled"`
			LogFile    string `json:"log_file"`
		} `json:"audit"`
	} `json:"security"`

	Metrics struct {
		Enabled bool   `json:"enabled"`
		Port    int    `json:"port"`
		Path    string `json:"path"`
	} `json:"metrics"`

	Logging struct {
		Level      string `json:"level"`
		Format     string `json:"format"`
		Output     string `json:"output"`
		File       string `json:"file"`
		MaxSize    string `json:"max_size"`
		MaxAge     string `json:"max_age"`
		MaxBackups int    `json:"max_backups"`
	} `json:"logging"`
}

// parseDuration парсит строку в time.Duration с значением по умолчанию
func parseDuration(s string, defaultVal time.Duration) time.Duration {
	if s == "" {
		return defaultVal
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return defaultVal
	}
	return d
}

// replaceEnvVars заменяет переменные окружения в строке
func replaceEnvVars(s string) string {
	if s == "" {
		return s
	}
	
	// Регулярное выражение для поиска ${VAR_NAME}
	re := regexp.MustCompile(`\$\{([^}]+)\}`)
	
	return re.ReplaceAllStringFunc(s, func(match string) string {
		// Извлекаем имя переменной из ${VAR_NAME}
		varName := match[2 : len(match)-1]
		
		// Получаем значение из переменной окружения
		if value := os.Getenv(varName); value != "" {
			return value
		}
		
		// Если переменная не найдена, возвращаем оригинальную строку
		return match
	})
}

// Load loads configuration from a file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	// Заменяем переменные окружения в JSON
	configStr := string(data)
	configStr = replaceEnvVars(configStr)
	
	var raw configRaw
	if err := json.Unmarshal([]byte(configStr), &raw); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	config := &Config{}

	// Server
	config.Server.Address = raw.Server.Address
	config.Server.Port = raw.Server.Port
	config.Server.Timeout = parseDuration(raw.Server.Timeout, 30*time.Second)
	config.Server.MaxTunnels = raw.Server.MaxTunnels
	config.Server.PingInterval = parseDuration(raw.Server.PingInterval, 30*time.Second)
	config.Server.CleanupInterval = parseDuration(raw.Server.CleanupInterval, time.Minute)

	// Django
	config.Django.APIEndpoint = raw.Django.APIEndpoint
	config.Django.JWTSecret = raw.Django.JWTSecret
	config.Django.JWTExpiry = parseDuration(raw.Django.JWTExpiry, 24*time.Hour)
	config.Django.RESTTimeout = parseDuration(raw.Django.RESTTimeout, 5*time.Second)
	config.Django.RESTCacheTTL = parseDuration(raw.Django.RESTCacheTTL, 5*time.Minute)
	config.Django.RESTMaxRetries = raw.Django.RESTMaxRetries
	config.Django.RESTRetryDelay = parseDuration(raw.Django.RESTRetryDelay, 100*time.Millisecond)

	// Keycloak
	config.Keycloak.Enabled = raw.Keycloak.Enabled
	config.Keycloak.ServerURL = raw.Keycloak.ServerURL
	config.Keycloak.Realm = raw.Keycloak.Realm
	config.Keycloak.ClientID = raw.Keycloak.ClientID
	config.Keycloak.ClientSecret = raw.Keycloak.ClientSecret
	config.Keycloak.PublicKey = raw.Keycloak.PublicKey
	config.Keycloak.TokenValidation.Issuer = raw.Keycloak.TokenValidation.Issuer
	config.Keycloak.TokenValidation.Audience = raw.Keycloak.TokenValidation.Audience
	config.Keycloak.TokenValidation.ClockSkew = parseDuration(raw.Keycloak.TokenValidation.ClockSkew, 30*time.Second)
	config.Keycloak.JWKSCacheTTL = parseDuration(raw.Keycloak.JWKSCacheTTL, time.Hour)
	config.Keycloak.HTTPTimeout = parseDuration(raw.Keycloak.HTTPTimeout, 30*time.Second)

	// Security
	config.Security.TLSEnabled = raw.Security.TLSEnabled
	config.Security.CertFile = raw.Security.CertFile
	config.Security.KeyFile = raw.Security.KeyFile
	config.Security.AllowedIPs = raw.Security.AllowedIPs
	config.Security.RateLimit = raw.Security.RateLimit
	config.Security.MTLS = raw.Security.MTLS
	config.Security.CipherSuites = raw.Security.CipherSuites
	config.Security.MinTLSVersion = raw.Security.MinTLSVersion
	config.Security.Audit = raw.Security.Audit

	// Metrics
	config.Metrics = raw.Metrics

	// Logging
	config.Logging = raw.Logging

	// Load JWT secret from environment if not set
	if config.Django.JWTSecret == "${RELAY_JWT_SECRET}" {
		if secret := os.Getenv("RELAY_JWT_SECRET"); secret != "" {
			config.Django.JWTSecret = secret
		} else {
			return nil, fmt.Errorf("JWT secret must be set via RELAY_JWT_SECRET environment variable")
		}
	}

	// Load Keycloak configuration from environment
	if config.Keycloak.Enabled {
		if config.Keycloak.ServerURL == "${KEYCLOAK_SERVER_URL}" {
			if url := os.Getenv("KEYCLOAK_SERVER_URL"); url != "" {
				config.Keycloak.ServerURL = url
			}
		}
		if config.Keycloak.Realm == "${KEYCLOAK_REALM}" {
			if realm := os.Getenv("KEYCLOAK_REALM"); realm != "" {
				config.Keycloak.Realm = realm
			}
		}
		if config.Keycloak.ClientID == "${KEYCLOAK_CLIENT_ID}" {
			if clientID := os.Getenv("KEYCLOAK_CLIENT_ID"); clientID != "" {
				config.Keycloak.ClientID = clientID
			}
		}
		if config.Keycloak.ClientSecret == "${KEYCLOAK_CLIENT_SECRET}" {
			if secret := os.Getenv("KEYCLOAK_CLIENT_SECRET"); secret != "" {
				config.Keycloak.ClientSecret = secret
			}
		}
		if config.Keycloak.PublicKey == "${KEYCLOAK_PUBLIC_KEY}" {
			if key := os.Getenv("KEYCLOAK_PUBLIC_KEY"); key != "" {
				config.Keycloak.PublicKey = key
			}
		}
		if config.Keycloak.TokenValidation.Issuer == "${KEYCLOAK_ISSUER}" {
			if issuer := os.Getenv("KEYCLOAK_ISSUER"); issuer != "" {
				config.Keycloak.TokenValidation.Issuer = issuer
			}
		}
		if config.Keycloak.TokenValidation.Audience == "${KEYCLOAK_AUDIENCE}" {
			if audience := os.Getenv("KEYCLOAK_AUDIENCE"); audience != "" {
				config.Keycloak.TokenValidation.Audience = audience
			}
		}
	}

	// Load TLS configuration from environment
	if config.Security.CertFile == "${RELAY_TLS_CERT_FILE}" {
		if certFile := os.Getenv("RELAY_TLS_CERT_FILE"); certFile != "" {
			config.Security.CertFile = certFile
		}
	}
	if config.Security.KeyFile == "${RELAY_TLS_KEY_FILE}" {
		if keyFile := os.Getenv("RELAY_TLS_KEY_FILE"); keyFile != "" {
			config.Security.KeyFile = keyFile
		}
	}
	if config.Security.MTLS.CACert == "${RELAY_MTLS_CA_FILE}" {
		if caFile := os.Getenv("RELAY_MTLS_CA_FILE"); caFile != "" {
			config.Security.MTLS.CACert = caFile
		}
	}

	// Set default values
	if config.Server.Address == "" {
		config.Server.Address = ":3456"
	}
	if config.Server.MaxTunnels == 0 {
		config.Server.MaxTunnels = 1000
	}

	return config, nil
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

	// Validate Django configuration (for backward compatibility)
	if !c.Keycloak.Enabled {
		if c.Django.APIEndpoint == "" {
			return fmt.Errorf("django.api_endpoint is required when keycloak is disabled")
		}
		if c.Django.JWTSecret == "" {
			return fmt.Errorf("django.jwt_secret is required when keycloak is disabled")
		}
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

	// Validate Keycloak configuration
	if c.Keycloak.Enabled {
		if c.Keycloak.ServerURL == "" {
			return fmt.Errorf("keycloak.server_url is required when keycloak is enabled")
		}
		if c.Keycloak.Realm == "" {
			return fmt.Errorf("keycloak.realm is required when keycloak is enabled")
		}
		if c.Keycloak.ClientID == "" {
			return fmt.Errorf("keycloak.client_id is required when keycloak is enabled")
		}
		if c.Keycloak.TokenValidation.Issuer == "" {
			return fmt.Errorf("keycloak.token_validation.issuer is required when keycloak is enabled")
		}
		if c.Keycloak.TokenValidation.Audience == "" {
			return fmt.Errorf("keycloak.token_validation.audience is required when keycloak is enabled")
		}
		if c.Keycloak.TokenValidation.ClockSkew < 0 {
			return fmt.Errorf("keycloak.token_validation.clock_skew must be non-negative")
		}
		if c.Keycloak.JWKSCacheTTL < time.Second {
			return fmt.Errorf("keycloak.jwks_cache_ttl must be at least 1 second")
		}
		if c.Keycloak.HTTPTimeout < time.Second {
			return fmt.Errorf("keycloak.http_timeout must be at least 1 second")
		}
	}

	if c.Security.TLSEnabled {
		if c.Security.CertFile == "" {
			return fmt.Errorf("cert_file is required when tls_enabled is true")
		}
		if c.Security.KeyFile == "" {
			return fmt.Errorf("key_file is required when tls_enabled is true")
		}
	}

	if c.Security.MTLS.Enabled {
		if c.Security.MTLS.CACert == "" {
			return fmt.Errorf("ca_file is required when mtls.enabled is true")
		}
	}

	// Validate TLS configuration
	if err := tls.ValidateTLSConfig(tls.Config{
		Enabled:      c.Security.TLSEnabled,
		CertFile:     c.Security.CertFile,
		KeyFile:      c.Security.KeyFile,
		MinVersion:   c.Security.MinTLSVersion,
		CipherSuites: c.Security.CipherSuites,
		MTLS: tls.MTLSConfig{
			Enabled:    c.Security.MTLS.Enabled,
			CACert:     c.Security.MTLS.CACert,
		},
	}); err != nil {
		return fmt.Errorf("invalid TLS configuration: %v", err)
	}

	return nil
} 