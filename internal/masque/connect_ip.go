package masque

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/quic-go/quicvarint"
	"go.uber.org/zap"
)

const (
	capsuleProtocolHeaderValue = "?1"
)

// ConnectIPProxy represents a CONNECT-IP proxy (RFC 9484)
type ConnectIPProxy struct {
	logger  *zap.Logger
	mx      sync.Mutex
	closed  bool
	closers map[*ConnectIPConnection]struct{}
}

// ConnectIPConnection represents a CONNECT-IP connection
type ConnectIPConnection struct {
	stream   *http3.Stream
	rawConn  net.Conn
	targetIP string
	ctx      context.Context
	cancel   context.CancelFunc
}

// ConnectIPRequest represents a CONNECT-IP request
type ConnectIPRequest struct {
	TargetIP string
}

// NewConnectIPProxy creates a new CONNECT-IP proxy
func NewConnectIPProxy(logger *zap.Logger) *ConnectIPProxy {
	return &ConnectIPProxy{
		logger:  logger,
		closers: make(map[*ConnectIPConnection]struct{}),
	}
}

// Proxy handles CONNECT-IP request (based on connect-ip-go reference)
func (p *ConnectIPProxy) Proxy(w http.ResponseWriter, r *ConnectIPRequest) (*ConnectIPConnection, error) {
	p.mx.Lock()
	if p.closed {
		p.mx.Unlock()
		w.WriteHeader(http.StatusServiceUnavailable)
		return nil, net.ErrClosed
	}
	p.mx.Unlock()

	// Set Capsule-Protocol header as required by RFC 9297
	w.Header().Set(http3.CapsuleProtocolHeader, capsuleProtocolHeaderValue)
	w.WriteHeader(http.StatusOK)

	// Get HTTP/3 stream
	str := w.(http3.HTTPStreamer).HTTPStream()

	// Create connection context
	ctx, cancel := context.WithCancel(context.Background())

	// Create CONNECT-IP connection
	conn := &ConnectIPConnection{
		stream:   str,
		targetIP: r.TargetIP,
		ctx:      ctx,
		cancel:   cancel,
	}

	// Register connection for cleanup
	p.mx.Lock()
	if p.closed {
		p.mx.Unlock()
		cancel()
		return nil, net.ErrClosed
	}
	p.closers[conn] = struct{}{}
	p.mx.Unlock()

	p.logger.Info("CONNECT-IP proxy established", zap.String("target_ip", r.TargetIP))
	return conn, nil
}

// Close closes the CONNECT-IP proxy
func (p *ConnectIPProxy) Close() error {
	p.mx.Lock()
	if p.closed {
		p.mx.Unlock()
		return nil
	}
	p.closed = true

	// Close all connections
	for conn := range p.closers {
		conn.Close()
	}
	p.closers = make(map[*ConnectIPConnection]struct{})
	p.mx.Unlock()

	return nil
}

// Close closes the CONNECT-IP connection
func (c *ConnectIPConnection) Close() error {
	c.cancel()
	if c.rawConn != nil {
		return c.rawConn.Close()
	}
	return c.stream.Close()
}

// Read reads data from the connection (implements net.Conn interface)
func (c *ConnectIPConnection) Read(b []byte) (int, error) {
	// For CONNECT-IP, we read from HTTP datagrams or capsules
	data, err := c.stream.ReceiveDatagram(c.ctx)
	if err != nil {
		return 0, err
	}

	// Parse context ID
	contextID, n, err := quicvarint.Parse(data)
	if err != nil {
		return 0, fmt.Errorf("malformed datagram: %w", err)
	}

	// For CONNECT-IP, context ID should be 0
	if contextID != 0 {
		return 0, fmt.Errorf("unexpected context ID: %d", contextID)
	}

	// Copy payload to buffer
	payload := data[n:]
	if len(payload) > len(b) {
		return 0, fmt.Errorf("buffer too small: need %d, have %d", len(payload), len(b))
	}

	copy(b, payload)
	return len(payload), nil
}

// Write writes data to the connection (implements net.Conn interface)
func (c *ConnectIPConnection) Write(b []byte) (int, error) {
	// Prepare HTTP datagram with context ID 0
	data := make([]byte, 0, len(contextIDZero)+len(b))
	data = append(data, contextIDZero...)
	data = append(data, b...)

	// Send HTTP datagram
	err := c.stream.SendDatagram(data)
	if err != nil {
		return 0, err
	}

	return len(b), nil
}

// LocalAddr returns the local network address (implements net.Conn interface)
func (c *ConnectIPConnection) LocalAddr() net.Addr {
	return &net.IPAddr{IP: net.IPv4zero}
}

// RemoteAddr returns the remote network address (implements net.Conn interface)
func (c *ConnectIPConnection) RemoteAddr() net.Addr {
	ip := net.ParseIP(c.targetIP)
	if ip == nil {
		return &net.IPAddr{IP: net.IPv4zero}
	}
	return &net.IPAddr{IP: ip}
}

// SetDeadline sets read and write deadlines (implements net.Conn interface)
func (c *ConnectIPConnection) SetDeadline(t time.Time) error {
	// HTTP/3 streams don't support deadlines directly
	return nil
}

// SetReadDeadline sets read deadline (implements net.Conn interface)
func (c *ConnectIPConnection) SetReadDeadline(t time.Time) error {
	return nil
}

// SetWriteDeadline sets write deadline (implements net.Conn interface)
func (c *ConnectIPConnection) SetWriteDeadline(t time.Time) error {
	return nil
}