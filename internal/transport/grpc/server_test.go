package grpc

import (
	"context"
	"testing"
	"time"

	"net"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twogc/cloudbridge-relay-installer/internal/auth"
	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
	"github.com/twogc/cloudbridge-relay-installer/internal/p2p"
	relayv1 "github.com/twogc/cloudbridge-relay-installer/internal/transport/grpc/proto/relay/v1"
	"github.com/twogc/cloudbridge-relay-installer/internal/tunnel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func init() {
	lis = bufconn.Listen(bufSize)
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func TestGRPCServer_Hello(t *testing.T) {
	// Create test components
	loggerInstance, err := logger.New(logger.Config{
		Level:  logger.INFO,
		Output: "stdout",
		JSON:   false,
	})
	require.NoError(t, err)

	config := Config{
		Port:             0, // Use bufconn listener instead
		TLSEnabled:       false,
		KeepalivePing:    15 * time.Second,
		KeepaliveTimeout: 10 * time.Second,
		MaxRecvMsgSize:   4 * 1024 * 1024,
		MaxSendMsgSize:   4 * 1024 * 1024,
	}

	// Create mock auth adapter
	authAdapter := &auth.Adapter{} // This would normally be properly initialized

	// Create tunnel manager
	tunnelManager := tunnel.NewManager(10, loggerInstance, nil)

	// Create heartbeat manager
	heartbeatManager := p2p.NewHeartbeatManager(nil, loggerInstance.GetZapLogger())

	// Create gRPC server
	server, err := NewServer(config, loggerInstance, authAdapter, tunnelManager, heartbeatManager)
	require.NoError(t, err)

	// Manually set the listener to bufconn for testing
	server.listener = lis

	// Start the gRPC server
	go func() {
		if err := server.grpcServer.Serve(lis); err != nil {
			t.Logf("gRPC server error: %v", err)
		}
	}()

	// Create gRPC client
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer conn.Close()

	client := relayv1.NewControlServiceClient(conn)

	// Test Hello RPC
	helloReq := &relayv1.HelloRequest{
		Version:   "1.0.0",
		Features:  []string{"tls", "tunneling"},
		ClientId:  "test-client-1",
		Timestamp: nil,
	}

	helloResp, err := client.Hello(ctx, helloReq)
	require.NoError(t, err)
	assert.NotNil(t, helloResp)
	assert.Equal(t, "ok", helloResp.Status)
	assert.Equal(t, "1.0.0", helloResp.ServerVersion)
	assert.NotEmpty(t, helloResp.SupportedFeatures)
	assert.Contains(t, helloResp.SupportedFeatures, "tls")
	assert.Contains(t, helloResp.SupportedFeatures, "tunneling")
	assert.NotEmpty(t, helloResp.SessionId)

	t.Logf("Hello response: %+v", helloResp)
}

func TestGRPCServer_Integration(t *testing.T) {
	// Create integration
	loggerInstance, err := logger.New(logger.Config{
		Level:  logger.INFO,
		Output: "stdout",
		JSON:   false,
	})
	require.NoError(t, err)

	config := DefaultConfig()
	config.Port = 0 // Will use random port
	config.TLSEnabled = false

	// Create basic components
	authAdapter := &auth.Adapter{}
	tunnelManager := tunnel.NewManager(10, loggerInstance, nil)
	heartbeatManager := p2p.NewHeartbeatManager(nil, loggerInstance.GetZapLogger())

	integration, err := NewIntegration(config, loggerInstance, authAdapter, tunnelManager, heartbeatManager)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start integration
	err = integration.Start(ctx)
	require.NoError(t, err)

	// Stop integration
	err = integration.Stop()
	require.NoError(t, err)

	t.Log("Integration test completed successfully")
}
