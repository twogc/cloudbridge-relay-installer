package database

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

// Minimal peer IPAM table for hybrid SCORE path (no FK deps on peers/tenant_networks).
const peerIPAMSchemaPostgres = `
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS peer_ip_allocations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id VARCHAR(255) NOT NULL,
    peer_id VARCHAR(255) NOT NULL,
    allocated_ip VARCHAR(45) NOT NULL,
    allocated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    released_at TIMESTAMPTZ,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    version INTEGER NOT NULL DEFAULT 1
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_peer_ip_active_pg
    ON peer_ip_allocations (tenant_id, allocated_ip)
    WHERE is_active = true;

CREATE UNIQUE INDEX IF NOT EXISTS uq_peer_active_allocation_pg
    ON peer_ip_allocations (peer_id)
    WHERE is_active = true;
`

// PostgresPeerIPAMConfig is resolved from env for SCORE hybrid production IPAM.
//
// Priority:
//  1. SCORE_DATABASE_URL or DATABASE_URL (libpq URL or postgres:// DSN)
//  2. SCORE_PG_HOST (+ SCORE_PG_PORT/USER/PASSWORD/DATABASE/SSLMODE)
type PostgresPeerIPAMConfig struct {
	// URL is a full postgres DSN or URL (optional if Host is set).
	URL string
	// Discrete fields used when URL is empty.
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string
}

// PostgresPeerIPAMFromEnv reads SCORE_DATABASE_URL, DATABASE_URL, or SCORE_PG_*.
func PostgresPeerIPAMFromEnv() PostgresPeerIPAMConfig {
	cfg := PostgresPeerIPAMConfig{
		URL: strings.TrimSpace(firstNonEmpty(
			os.Getenv("SCORE_DATABASE_URL"),
			os.Getenv("DATABASE_URL"),
		)),
		Host:     strings.TrimSpace(os.Getenv("SCORE_PG_HOST")),
		User:     strings.TrimSpace(firstNonEmpty(os.Getenv("SCORE_PG_USER"), os.Getenv("SCORE_PG_USERNAME"))),
		Password: strings.TrimSpace(os.Getenv("SCORE_PG_PASSWORD")),
		Database: strings.TrimSpace(firstNonEmpty(os.Getenv("SCORE_PG_DATABASE"), os.Getenv("SCORE_PG_DB"))),
		SSLMode:  strings.TrimSpace(os.Getenv("SCORE_PG_SSLMODE")),
		Port:     5432,
	}
	if p := strings.TrimSpace(os.Getenv("SCORE_PG_PORT")); p != "" {
		if n, err := strconv.Atoi(p); err == nil {
			cfg.Port = n
		}
	}
	if cfg.SSLMode == "" {
		cfg.SSLMode = "disable"
	}
	if cfg.Database == "" {
		cfg.Database = "cloudbridge_p2p"
	}
	if cfg.User == "" {
		cfg.User = "cloudbridge"
	}
	return cfg
}

// Enabled reports whether enough config is present to attempt Postgres.
func (c PostgresPeerIPAMConfig) Enabled() bool {
	if c.URL != "" {
		return true
	}
	return c.Host != ""
}

// DSN builds a libpq keyword DSN or returns the URL as-is.
func (c PostgresPeerIPAMConfig) DSN() (string, error) {
	if c.URL != "" {
		return normalizePostgresURL(c.URL), nil
	}
	if c.Host == "" {
		return "", fmt.Errorf("postgres peer IPAM: host or URL required")
	}
	if c.Port <= 0 {
		c.Port = 5432
	}
	return fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=%s connect_timeout=10",
		c.Host, c.Port, c.Database, c.User, c.Password, c.SSLMode,
	), nil
}

func normalizePostgresURL(raw string) string {
	// sqlx/libpq accept postgres:// and postgresql://
	u := strings.TrimSpace(raw)
	if strings.HasPrefix(u, "postgres://") || strings.HasPrefix(u, "postgresql://") {
		if _, err := url.Parse(u); err == nil {
			return u
		}
	}
	return u
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

// OpenPostgresPeerIPAM connects to Postgres, ensures peer_ip_allocations exists, returns Repository.
func OpenPostgresPeerIPAM(cfg PostgresPeerIPAMConfig, logger *zap.Logger) (*Repository, error) {
	if logger == nil {
		logger = zap.NewNop()
	}
	if !cfg.Enabled() {
		return nil, fmt.Errorf("postgres peer IPAM not configured")
	}
	dsn, err := cfg.DSN()
	if err != nil {
		return nil, err
	}

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	if _, err := db.Exec(peerIPAMSchemaPostgres); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("apply peer IPAM schema: %w", err)
	}

	logger.Info("Postgres peer IPAM repository ready",
		zap.Bool("url_mode", cfg.URL != ""),
		zap.String("host", cfg.Host),
		zap.String("database", cfg.Database),
	)
	return NewRepository(db, logger), nil
}
