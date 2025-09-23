package masque

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/quic-go/quicvarint"
	"go.uber.org/zap"
)

const capsuleProtocolHeaderValue = "?1"

// ConnectIPProxy represents a CONNECT-IP proxy (RFC 9484)
type ConnectIPProxy struct {
	logger *zap.Logger
}

// ConnectIPRequest represents a CONNECT-IP request
type ConnectIPRequest struct {
	TargetIP string
}

// ConnectIPConn represents a CONNECT-IP connection (from connect-ip-go reference)
type ConnectIPConn struct {
	str *http3.Stream
}

// NewConnectIPProxy creates a new CONNECT-IP proxy
func NewConnectIPProxy(logger *zap.Logger) *ConnectIPProxy {
	return &ConnectIPProxy{
		logger: logger,
	}
}

// Proxy handles a CONNECT-IP request (based on connect-ip-go reference)
func (p *ConnectIPProxy) Proxy(w http.ResponseWriter, r *ConnectIPRequest) (*ConnectIPConn, error) {
	p.logger.Info("Processing CONNECT-IP request", zap.String("target_ip", r.TargetIP))

	// Validate target IP
	targetIP := net.ParseIP(r.TargetIP)
	if targetIP == nil {
		http.Error(w, "Invalid target IP", http.StatusBadRequest)
		return nil, fmt.Errorf("invalid target IP: %s", r.TargetIP)
	}

	// Set required headers for CONNECT-IP
	w.Header().Set(http3.CapsuleProtocolHeader, capsuleProtocolHeaderValue)
	w.WriteHeader(http.StatusOK)

	// Get HTTP/3 stream
	str := w.(http3.HTTPStreamer).HTTPStream()

	// Create proxied connection
	conn := newProxiedIPConn(str)

	p.logger.Info("CONNECT-IP proxy established", zap.String("target_ip", r.TargetIP))
	return conn, nil
}

// newProxiedIPConn creates a new proxied IP connection (from connect-ip-go reference)
func newProxiedIPConn(str *http3.Stream) *ConnectIPConn {
	return &ConnectIPConn{
		str: str,
	}
}

// ReadPacket reads an IP packet from the connection
func (c *ConnectIPConn) ReadPacket(ctx context.Context) ([]byte, error) {
	// Receive HTTP datagram
	data, err := c.str.ReceiveDatagram(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to receive datagram: %w", err)
	}

	// Parse context ID (for CONNECT-IP, context ID can be non-zero)
	contextID, n, err := quicvarint.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("malformed datagram: %w", err)
	}

	// Extract IP packet payload
	payload := data[n:]

	// For CONNECT-IP, we might need to handle IP Capsules
	// This is a simplified implementation
	_ = contextID // Context ID handling can be added here

	return payload, nil
}

// WritePacket writes an IP packet to the connection
func (c *ConnectIPConn) WritePacket(packet []byte) error {
	// Prepare HTTP datagram with context ID 0 for simplicity
	// In a full implementation, context ID management would be more sophisticated
	data := make([]byte, 0, len(contextIDZero)+len(packet))
	data = append(data, contextIDZero...)
	data = append(data, packet...)

	// Send HTTP datagram
	return c.str.SendDatagram(data)
}

// Close closes the CONNECT-IP connection
func (c *ConnectIPConn) Close() error {
	return c.str.Close()
}

// LocalAddr returns the local network address
func (c *ConnectIPConn) LocalAddr() net.Addr {
	return &connectIPAddr{addr: "connect-ip"}
}

// RemoteAddr returns the remote network address
func (c *ConnectIPConn) RemoteAddr() net.Addr {
	return &connectIPAddr{addr: "connect-ip"}
}

// connectIPAddr implements net.Addr for CONNECT-IP
type connectIPAddr struct {
	addr string
}

func (a *connectIPAddr) Network() string {
	return "connect-ip"
}

func (a *connectIPAddr) String() string {
	return a.addr
}
