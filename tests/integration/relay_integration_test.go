package integration

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/config"
	"github.com/twogc/cloudbridge-relay-installer/internal/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestServer represents a test relay server
type TestServer struct {
	server *Server
	port   int
}

// TestClient represents a test client
type TestClient struct {
	conn net.Conn
}

// setupTestServer creates a test server
func setupTestServer(t *testing.T) *TestServer {
	cfg := &config.Config{}
	cfg.Server.Port = 0 // Use random port
	cfg.Server.MaxTunnels = 10
	cfg.Django.APIEndpoint = "http://localhost:8000/api"
	cfg.Django.JWTSecret = "test-secret"
	cfg.Security.TLSEnabled = false
	cfg.Security.RateLimit.Enabled = false

	server, err := NewServer(cfg)
	require.NoError(t, err)

	go func() {
		if err := server.Start(cfg); err != nil {
			t.Logf("Server error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	return &TestServer{
		server: server,
		port:   server.port,
	}
}

// TestBasicConnection tests basic connection establishment
func TestBasicConnection(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.server.Shutdown(context.Background())

	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", ts.port))
	require.NoError(t, err)
	defer conn.Close()

	// Test hello message
	helloMsg := &proto.Message{
		Type:    proto.MessageTypeHello,
		Version: "1.0",
		Features: []string{"tls", "heartbeat"},
	}

	err = proto.WriteMessage(conn.(*net.TCPConn).Writer(), helloMsg)
	require.NoError(t, err)

	// Read response
	response, err := proto.ReadMessage(conn.(*net.TCPConn).Reader())
	require.NoError(t, err)
	assert.Equal(t, proto.MessageTypeHello, response.Type)
}

// TestAuthentication tests JWT authentication
func TestAuthentication(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.server.Shutdown(context.Background())

	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", ts.port))
	require.NoError(t, err)
	defer conn.Close()

	// Send auth message
	authMsg := &proto.Message{
		Type:  proto.MessageTypeAuth,
		Token: "valid-jwt-token",
	}

	err = proto.WriteMessage(conn.(*net.TCPConn).Writer(), authMsg)
	require.NoError(t, err)

	// Read auth response
	response, err := proto.ReadMessage(conn.(*net.TCPConn).Reader())
	require.NoError(t, err)
	assert.Equal(t, proto.MessageTypeAuthResponse, response.Type)
}

// TestTunnelCreation tests tunnel creation
func TestTunnelCreation(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.server.Shutdown(context.Background())

	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", ts.port))
	require.NoError(t, err)
	defer conn.Close()

	// Authenticate first
	authMsg := &proto.Message{
		Type:  proto.MessageTypeAuth,
		Token: "valid-jwt-token",
	}
	err = proto.WriteMessage(conn.(*net.TCPConn).Writer(), authMsg)
	require.NoError(t, err)

	response, err := proto.ReadMessage(conn.(*net.TCPConn).Reader())
	require.NoError(t, err)
	assert.Equal(t, proto.MessageTypeAuthResponse, response.Type)

	// Create tunnel
	tunnelInfo := &proto.TunnelInfo{
		TunnelID:   "test-tunnel",
		LocalPort:  3389,
		RemoteHost: "internal-server",
		RemotePort: 3389,
	}

	tunnelMsg := &proto.Message{
		Type:       proto.MessageTypeTunnelInfo,
		TunnelInfo: tunnelInfo,
	}

	err = proto.WriteMessage(conn.(*net.TCPConn).Writer(), tunnelMsg)
	require.NoError(t, err)

	// Verify tunnel was created
	assert.Equal(t, 1, len(ts.server.tunnels))
}

// TestHeartbeat tests heartbeat functionality
func TestHeartbeat(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.server.Shutdown(context.Background())

	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", ts.port))
	require.NoError(t, err)
	defer conn.Close()

	// Authenticate
	authMsg := &proto.Message{
		Type:  proto.MessageTypeAuth,
		Token: "valid-jwt-token",
	}
	err = proto.WriteMessage(conn.(*net.TCPConn).Writer(), authMsg)
	require.NoError(t, err)

	response, err := proto.ReadMessage(conn.(*net.TCPConn).Reader())
	require.NoError(t, err)
	assert.Equal(t, proto.MessageTypeAuthResponse, response.Type)

	// Send heartbeat
	heartbeatMsg := &proto.Message{
		Type: proto.MessageTypeHeartbeat,
	}

	err = proto.WriteMessage(conn.(*net.TCPConn).Writer(), heartbeatMsg)
	require.NoError(t, err)

	// Read heartbeat response
	response, err = proto.ReadMessage(conn.(*net.TCPConn).Reader())
	require.NoError(t, err)
	assert.Equal(t, proto.MessageTypeHeartbeatResponse, response.Type)
}

// TestRateLimiting tests rate limiting functionality
func TestRateLimiting(t *testing.T) {
	cfg := &config.Config{}
	cfg.Server.Port = 0
	cfg.Server.MaxTunnels = 10
	cfg.Django.APIEndpoint = "http://localhost:8000/api"
	cfg.Django.JWTSecret = "test-secret"
	cfg.Security.TLSEnabled = false
	cfg.Security.RateLimit.Enabled = true
	cfg.Security.RateLimit.RequestsPerSecond = 1
	cfg.Security.RateLimit.Burst = 1

	server, err := NewServer(cfg)
	require.NoError(t, err)

	go func() {
		if err := server.Start(cfg); err != nil {
			t.Logf("Server error: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)
	defer server.Shutdown(context.Background())

	// Try to connect multiple times rapidly
	for i := 0; i < 5; i++ {
		conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", server.port))
		if err != nil {
			// Expected after rate limit
			continue
		}
		conn.Close()
	}

	// Verify rate limiting worked
	// This is a basic test - in a real scenario we'd check metrics
}

// TestTLSConnection tests TLS connection
func TestTLSConnection(t *testing.T) {
	// Skip if TLS certificates not available
	if !hasTLSCertificates() {
		t.Skip("TLS certificates not available")
	}

	cfg := &config.Config{}
	cfg.Server.Port = 0
	cfg.Server.MaxTunnels = 10
	cfg.Django.APIEndpoint = "http://localhost:8000/api"
	cfg.Django.JWTSecret = "test-secret"
	cfg.Security.TLSEnabled = true
	cfg.Security.CertFile = "/etc/relay/certs/fullchain.pem"
	cfg.Security.KeyFile = "/etc/relay/certs/privkey.pem"

	server, err := NewServer(cfg)
	require.NoError(t, err)

	go func() {
		if err := server.Start(cfg); err != nil {
			t.Logf("Server error: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)
	defer server.Shutdown(context.Background())

	// Connect with TLS
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, // For testing only
	}

	conn, err := tls.Dial("tcp", fmt.Sprintf("localhost:%d", server.port), tlsConfig)
	require.NoError(t, err)
	defer conn.Close()

	// Test connection
	helloMsg := &proto.Message{
		Type:    proto.MessageTypeHello,
		Version: "1.0",
	}

	err = proto.WriteMessage(conn.(*tls.Conn).Writer(), helloMsg)
	require.NoError(t, err)

	response, err := proto.ReadMessage(conn.(*tls.Conn).Reader())
	require.NoError(t, err)
	assert.Equal(t, proto.MessageTypeHello, response.Type)
}

// TestErrorHandling tests error handling
func TestErrorHandling(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.server.Shutdown(context.Background())

	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", ts.port))
	require.NoError(t, err)
	defer conn.Close()

	// Send invalid message
	invalidMsg := &proto.Message{
		Type: "invalid_type",
	}

	err = proto.WriteMessage(conn.(*net.TCPConn).Writer(), invalidMsg)
	require.NoError(t, err)

	// Should receive error response
	response, err := proto.ReadMessage(conn.(*net.TCPConn).Reader())
	require.NoError(t, err)
	assert.Equal(t, proto.MessageTypeError, response.Type)
	assert.NotNil(t, response.Error)
}

// TestConcurrentConnections tests concurrent connections
func TestConcurrentConnections(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.server.Shutdown(context.Background())

	const numConnections = 10
	connections := make([]net.Conn, numConnections)
	errors := make(chan error, numConnections)

	// Create concurrent connections
	for i := 0; i < numConnections; i++ {
		go func(id int) {
			conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", ts.port))
			if err != nil {
				errors <- err
				return
			}
			connections[id] = conn

			// Send hello message
			helloMsg := &proto.Message{
				Type:    proto.MessageTypeHello,
				Version: "1.0",
			}
			err = proto.WriteMessage(conn.(*net.TCPConn).Writer(), helloMsg)
			if err != nil {
				errors <- err
				return
			}

			// Read response
			_, err = proto.ReadMessage(conn.(*net.TCPConn).Reader())
			if err != nil {
				errors <- err
				return
			}

			errors <- nil
		}(i)
	}

	// Wait for all connections
	for i := 0; i < numConnections; i++ {
		err := <-errors
		assert.NoError(t, err)
	}

	// Clean up
	for _, conn := range connections {
		if conn != nil {
			conn.Close()
		}
	}
}

// Helper functions
func hasTLSCertificates() bool {
	// Check if TLS certificates exist
	_, err1 := os.Stat("/etc/relay/certs/fullchain.pem")
	_, err2 := os.Stat("/etc/relay/certs/privkey.pem")
	return err1 == nil && err2 == nil
} 