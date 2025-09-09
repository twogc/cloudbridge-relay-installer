package database

import (
	"fmt"
	"time"
)

// Config represents database configuration
type Config struct {
	Host            string        `json:"host" yaml:"host"`
	Port            int           `json:"port" yaml:"port"`
	Database        string        `json:"database" yaml:"database"`
	Username        string        `json:"username" yaml:"username"`
	Password        string        `json:"password" yaml:"password"`
	SSLMode         string        `json:"ssl_mode" yaml:"ssl_mode"`
	MaxOpenConns    int           `json:"max_open_conns" yaml:"max_open_conns"`
	MaxIdleConns    int           `json:"max_idle_conns" yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime" yaml:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `json:"conn_max_idle_time" yaml:"conn_max_idle_time"`
	ConnectTimeout  time.Duration `json:"connect_timeout" yaml:"connect_timeout"`
	QueryTimeout    time.Duration `json:"query_timeout" yaml:"query_timeout"`
}

// DefaultConfig returns default database configuration
func DefaultConfig() *Config {
	return &Config{
		Host:            "localhost",
		Port:            5432,
		Database:        "cloudbridge_p2p",
		Username:        "cloudbridge",
		Password:        "cloudbridge",
		SSLMode:         "disable",
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 1 * time.Minute,
		ConnectTimeout:  30 * time.Second,
		QueryTimeout:    10 * time.Second,
	}
}

// DSN returns the database connection string
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=%s connect_timeout=%d",
		c.Host, c.Port, c.Database, c.Username, c.Password, c.SSLMode, int(c.ConnectTimeout.Seconds()),
	)
}

// Validate validates the database configuration
func (c *Config) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("database host is required")
	}
	
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("database port must be between 1 and 65535")
	}
	
	if c.Database == "" {
		return fmt.Errorf("database name is required")
	}
	
	if c.Username == "" {
		return fmt.Errorf("database username is required")
	}
	
	if c.Password == "" {
		return fmt.Errorf("database password is required")
	}
	
	validSSLModes := []string{"disable", "require", "verify-ca", "verify-full"}
	validSSLMode := false
	for _, mode := range validSSLModes {
		if c.SSLMode == mode {
			validSSLMode = true
			break
		}
	}
	if !validSSLMode {
		return fmt.Errorf("invalid SSL mode: %s, must be one of: %v", c.SSLMode, validSSLModes)
	}
	
	if c.MaxOpenConns <= 0 {
		return fmt.Errorf("max_open_conns must be greater than 0")
	}
	
	if c.MaxIdleConns <= 0 {
		return fmt.Errorf("max_idle_conns must be greater than 0")
	}
	
	if c.MaxIdleConns > c.MaxOpenConns {
		return fmt.Errorf("max_idle_conns cannot be greater than max_open_conns")
	}
	
	if c.ConnMaxLifetime <= 0 {
		return fmt.Errorf("conn_max_lifetime must be greater than 0")
	}
	
	if c.ConnMaxIdleTime <= 0 {
		return fmt.Errorf("conn_max_idle_time must be greater than 0")
	}
	
	if c.ConnectTimeout <= 0 {
		return fmt.Errorf("connect_timeout must be greater than 0")
	}
	
	if c.QueryTimeout <= 0 {
		return fmt.Errorf("query_timeout must be greater than 0")
	}
	
	return nil
}
