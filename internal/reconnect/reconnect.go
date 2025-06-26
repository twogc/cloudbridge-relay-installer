package reconnect

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/internal/errors"
	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
)

// Config represents the reconnection configuration
type Config struct {
	MaxRetries      int
	RetryDelay      time.Duration
	MaxRetryDelay   time.Duration
	BackoffFactor   float64
	Logger          *logger.Logger
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		MaxRetries:    5,
		RetryDelay:    time.Second,
		MaxRetryDelay: time.Minute,
		BackoffFactor: 2.0,
		Logger:        logger.NewLogger(logger.Config{
			Level:    logger.LogLevelInfo,
			Console:  true,
			Format:   "json",
			MaxSize:  100,
			MaxAge:   30,
			MaxBackups: 10,
			Compress: true,
		}),
	}
}

// Connection represents a reconnecting connection
type Connection struct {
	config     *Config
	addr       string
	conn       net.Conn
	onConnect  func(net.Conn) error
	onDisconnect func(error)
}

// NewConnection creates a new reconnecting connection
func NewConnection(addr string, config *Config) *Connection {
	if config == nil {
		config = DefaultConfig()
	}
	return &Connection{
		config: config,
		addr:   addr,
	}
}

// SetOnConnect sets the onConnect callback
func (c *Connection) SetOnConnect(callback func(net.Conn) error) {
	c.onConnect = callback
}

// SetOnDisconnect sets the onDisconnect callback
func (c *Connection) SetOnDisconnect(callback func(error)) {
	c.onDisconnect = callback
}

// Connect establishes a connection with automatic reconnection
func (c *Connection) Connect(ctx context.Context) error {
	var lastErr error
	retryDelay := c.config.RetryDelay

	for i := 0; i < c.config.MaxRetries; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Try to connect
			conn, err := net.Dial("tcp", c.addr)
			if err != nil {
				lastErr = err
				c.config.Logger.Warning("Connection failed", map[string]interface{}{
					"attempt": i + 1,
					"error":   err.Error(),
				})

				// Wait before retrying
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(retryDelay):
				}

				// Increase delay for next retry
				retryDelay = time.Duration(float64(retryDelay) * c.config.BackoffFactor)
				if retryDelay > c.config.MaxRetryDelay {
					retryDelay = c.config.MaxRetryDelay
				}
				continue
			}

			// Connection successful
			c.conn = conn
			c.config.Logger.Info("Connection established", map[string]interface{}{
				"remote_addr": conn.RemoteAddr(),
			})

			// Call onConnect callback
			if c.onConnect != nil {
				if err := c.onConnect(conn); err != nil {
					conn.Close()
					return err
				}
			}

			// Start monitoring connection
			go c.monitorConnection(ctx)
			return nil
		}
	}

	return errors.NewConnectionError(lastErr)
}

// monitorConnection monitors the connection and attempts to reconnect if it's lost
func (c *Connection) monitorConnection(ctx context.Context) {
	buf := make([]byte, 1)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Try to read from connection
			_, err := c.conn.Read(buf)
			if err != nil {
				c.config.Logger.Warning("Connection lost", map[string]interface{}{
					"error": err.Error(),
				})

				// Call onDisconnect callback
				if c.onDisconnect != nil {
					c.onDisconnect(err)
				}

				// Close connection
				c.conn.Close()

				// Try to reconnect
				if err := c.Connect(ctx); err != nil {
					c.config.Logger.Error("Reconnection failed", err, nil)
					return
				}
			}
		}
	}
}

// Close closes the connection
func (c *Connection) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Write writes data to the connection
func (c *Connection) Write(data []byte) (int, error) {
	if c.conn == nil {
		return 0, fmt.Errorf("not connected")
	}
	return c.conn.Write(data)
}

// Read reads data from the connection
func (c *Connection) Read(data []byte) (int, error) {
	if c.conn == nil {
		return 0, fmt.Errorf("not connected")
	}
	return c.conn.Read(data)
} 