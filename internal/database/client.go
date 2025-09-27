package database

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver
	"go.uber.org/zap"
)

// Client represents a database client
type Client struct {
	db     *sqlx.DB
	config *Config
	logger *zap.Logger
}

// NewClient creates a new database client
func NewClient(config *Config, logger *zap.Logger) (*Client, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid database config: %w", err)
	}

	// Open database connection
	db, err := sqlx.Connect("postgres", config.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)
	db.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	client := &Client{
		db:     db,
		config: config,
		logger: logger,
	}

	logger.Info("Database client created successfully",
		zap.String("host", config.Host),
		zap.Int("port", config.Port),
		zap.String("database", config.Database),
		zap.String("username", config.Username),
		zap.String("ssl_mode", config.SSLMode),
		zap.Int("max_open_conns", config.MaxOpenConns),
		zap.Int("max_idle_conns", config.MaxIdleConns),
	)

	return client, nil
}

// GetDB returns the underlying database connection
func (c *Client) GetDB() *sqlx.DB {
	return c.db
}

// Close closes the database connection
func (c *Client) Close() error {
	if c.db != nil {
		err := c.db.Close()
		if err != nil {
			c.logger.Error("Failed to close database connection", zap.Error(err))
			return err
		}
		c.logger.Info("Database connection closed")
	}
	return nil
}

// Health checks database health
func (c *Client) Health() error {
	ctx, cancel := context.WithTimeout(context.Background(), c.config.QueryTimeout)
	defer cancel()

	if err := c.db.PingContext(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	return nil
}

// Stats returns database connection statistics
func (c *Client) Stats() sql.DBStats {
	return c.db.Stats()
}

// Migrate runs database migrations
func (c *Client) Migrate() error {
	c.logger.Info("Running database migrations...")

	// Read schema file
	schema, err := ioutil.ReadFile("internal/database/schema.sql")
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	// Execute schema
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	_, err = c.db.ExecContext(ctx, string(schema))
	if err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	c.logger.Info("Database schema migration completed successfully")
	return nil
}

// MigrateIPAM runs IPAM-specific migrations
func (c *Client) MigrateIPAM() error {
	c.logger.Info("Running IPAM migrations...")

	// Read migrations file
	migrations, err := ioutil.ReadFile("internal/database/migrations.sql")
	if err != nil {
		return fmt.Errorf("failed to read migrations file: %w", err)
	}

	// Execute migrations
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	_, err = c.db.ExecContext(ctx, string(migrations))
	if err != nil {
		return fmt.Errorf("failed to execute IPAM migrations: %w", err)
	}

	c.logger.Info("IPAM migrations completed successfully")
	return nil
}

// Transaction executes a function within a database transaction
func (c *Client) Transaction(fn func(*sqlx.Tx) error) error {
	tx, err := c.db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	err = fn(tx)
	return err
}

// WithTimeout executes a function with a timeout
func (c *Client) WithTimeout(timeout time.Duration, fn func(*sqlx.DB) error) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- fn(c.db)
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return fmt.Errorf("operation timed out after %v", timeout)
	}
}
