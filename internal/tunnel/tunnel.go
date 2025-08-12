package tunnel

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
	"github.com/twogc/cloudbridge-relay-installer/internal/metrics"
	"go.uber.org/zap"
)

// Proxy represents a TCP proxy for tunneling
type Proxy struct {
	clientConn net.Conn
	serverConn net.Conn
	remoteHost string
	remotePort int
	tunnelID   string
	logger     *logger.Logger
	metrics    *metrics.Metrics
	ctx        context.Context
	cancel     context.CancelFunc
	mu         sync.RWMutex
	closed     bool
	stats      *ProxyStats
}

// ProxyStats represents proxy statistics
type ProxyStats struct {
	BytesFromClient int64
	BytesToClient   int64
	StartTime       time.Time
	LastActivity    time.Time
	mu              sync.RWMutex
}

// NewProxy creates a new TCP proxy
func NewProxy(
	clientConn net.Conn,
	remoteHost string,
	remotePort int,
	tunnelID string,
	logger *logger.Logger,
	metrics *metrics.Metrics,
) *Proxy {
	ctx, cancel := context.WithCancel(context.Background())

	return &Proxy{
		clientConn: clientConn,
		remoteHost: remoteHost,
		remotePort: remotePort,
		tunnelID:   tunnelID,
		logger:     logger,
		metrics:    metrics,
		ctx:        ctx,
		cancel:     cancel,
		stats: &ProxyStats{
			StartTime:    time.Now(),
			LastActivity: time.Now(),
		},
	}
}

// Start starts the proxy
func (p *Proxy) Start() error {
	p.logger.Info("Starting TCP proxy",
		zap.String("tunnel_id", p.tunnelID),
		zap.String("remote_host", p.remoteHost),
		zap.Int("remote_port", p.remotePort),
	)

	// Connect to remote server
	serverConn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", p.remoteHost, p.remotePort), 30*time.Second)
	if err != nil {
		p.logger.Error("Failed to connect to remote server",
			zap.String("tunnel_id", p.tunnelID),
			zap.String("remote_host", p.remoteHost),
			zap.Int("remote_port", p.remotePort),
			zap.Error(err),
		)
		return fmt.Errorf("failed to connect to remote server: %v", err)
	}

	p.serverConn = serverConn

	// Start bidirectional data transfer
	go p.transferData(p.clientConn, p.serverConn, "client->server")
	go p.transferData(p.serverConn, p.clientConn, "server->client")

	p.logger.Info("TCP proxy started successfully",
		zap.String("tunnel_id", p.tunnelID),
	)

	return nil
}

// transferData transfers data between two connections
func (p *Proxy) transferData(src, dst net.Conn, direction string) {
	defer func() {
		p.Close()
	}()

	buffer := make([]byte, 32*1024) // 32KB buffer
	startTime := time.Now()

	for {
		select {
		case <-p.ctx.Done():
			return
		default:
			// Set read timeout
			if err := src.SetReadDeadline(time.Now().Add(30 * time.Second)); err != nil {
				p.logger.Error("Failed to set read deadline",
					zap.String("tunnel_id", p.tunnelID),
					zap.String("direction", direction),
					zap.Error(err),
				)
				return
			}

			// Read data from source
			n, err := src.Read(buffer)
			if err != nil {
				if err == io.EOF {
					p.logger.Debug("Connection closed by peer",
						zap.String("tunnel_id", p.tunnelID),
						zap.String("direction", direction),
					)
				} else {
					p.logger.Error("Error reading from connection",
						zap.String("tunnel_id", p.tunnelID),
						zap.String("direction", direction),
						zap.Error(err),
					)
				}
				return
			}

			if n == 0 {
				continue
			}

			// Update statistics
			p.updateStats(direction, int64(n))

			// Set write timeout
			if err := dst.SetWriteDeadline(time.Now().Add(30 * time.Second)); err != nil {
				p.logger.Error("Failed to set write deadline",
					zap.String("tunnel_id", p.tunnelID),
					zap.String("direction", direction),
					zap.Error(err),
				)
				return
			}

			// Write data to destination
			_, err = dst.Write(buffer[:n])
			if err != nil {
				p.logger.Error("Error writing to connection",
					zap.String("tunnel_id", p.tunnelID),
					zap.String("direction", direction),
					zap.Error(err),
				)
				return
			}

			// Update metrics
			if p.metrics != nil {
				if direction == "client->server" {
					p.metrics.ObserveNetworkLatency(time.Since(startTime))
					p.metrics.IncDiskIO()
				}
			}
		}
	}
}

// updateStats updates proxy statistics
func (p *Proxy) updateStats(direction string, bytes int64) {
	p.stats.mu.Lock()
	defer p.stats.mu.Unlock()

	p.stats.LastActivity = time.Now()

	if direction == "client->server" {
		p.stats.BytesFromClient += bytes
	} else {
		p.stats.BytesToClient += bytes
	}
}

// Close closes the proxy
func (p *Proxy) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return
	}

	p.closed = true
	p.cancel()

	if p.clientConn != nil {
		if err := p.clientConn.Close(); err != nil {
			p.logger.Error("Failed to close client connection",
				zap.String("tunnel_id", p.tunnelID),
				zap.Error(err),
			)
		}
	}

	if p.serverConn != nil {
		if err := p.serverConn.Close(); err != nil {
			p.logger.Error("Failed to close server connection",
				zap.String("tunnel_id", p.tunnelID),
				zap.Error(err),
			)
		}
	}

	duration := time.Since(p.stats.StartTime)
	p.logger.Info("TCP proxy closed",
		zap.String("tunnel_id", p.tunnelID),
		zap.Duration("duration", duration),
		zap.Int64("bytes_from_client", p.stats.BytesFromClient),
		zap.Int64("bytes_to_client", p.stats.BytesToClient),
	)

	// Update metrics
	if p.metrics != nil {
		p.metrics.ObserveConnectionDuration(duration)
	}
}

// GetStats returns proxy statistics
func (p *Proxy) GetStats() *ProxyStats {
	p.stats.mu.RLock()
	defer p.stats.mu.RUnlock()

	return &ProxyStats{
		BytesFromClient: p.stats.BytesFromClient,
		BytesToClient:   p.stats.BytesToClient,
		StartTime:       p.stats.StartTime,
		LastActivity:    p.stats.LastActivity,
	}
}

// IsActive checks if the proxy is still active
func (p *Proxy) IsActive() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return false
	}

	// Check if connections are still alive
	if p.clientConn != nil {
		if _, err := p.clientConn.Write([]byte{}); err != nil {
			return false
		}
	}

	if p.serverConn != nil {
		if _, err := p.serverConn.Write([]byte{}); err != nil {
			return false
		}
	}

	return true
}

// GetLastActivity returns the last activity time
func (p *Proxy) GetLastActivity() time.Time {
	p.stats.mu.RLock()
	defer p.stats.mu.RUnlock()
	return p.stats.LastActivity
}
