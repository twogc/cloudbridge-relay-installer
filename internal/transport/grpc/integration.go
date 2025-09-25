package grpc

import (
	"context"
	"fmt"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/internal/auth"
	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
	"github.com/twogc/cloudbridge-relay-installer/internal/p2p"
	"github.com/twogc/cloudbridge-relay-installer/internal/tunnel"
	"go.uber.org/zap"
)

// Integration provides gRPC server integration with the main relay server
type Integration struct {
	server           *Server
	config           Config
	logger           *logger.Logger
	authAdapter      *auth.Adapter
	tunnelManager    *tunnel.Manager
	heartbeatManager *p2p.HeartbeatManager
}

// NewIntegration creates a new gRPC integration
func NewIntegration(
	config Config,
	logger *logger.Logger,
	authAdapter *auth.Adapter,
	tunnelManager *tunnel.Manager,
	heartbeatManager *p2p.HeartbeatManager,
) (*Integration, error) {
	server, err := NewServer(config, logger, authAdapter, tunnelManager, heartbeatManager)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC server: %v", err)
	}

	return &Integration{
		server:           server,
		config:           config,
		logger:           logger,
		authAdapter:      authAdapter,
		tunnelManager:    tunnelManager,
		heartbeatManager: heartbeatManager,
	}, nil
}

// Start starts the gRPC server integration
func (i *Integration) Start(ctx context.Context) error {
	i.logger.Info("Starting gRPC server integration",
		zap.Int("port", i.config.Port),
		zap.Bool("tls_enabled", i.config.TLSEnabled),
		zap.Bool("mtls_enabled", i.config.MTLSEnabled),
	)

	// Start the gRPC server in a goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := i.server.StartBlocking(); err != nil {
			errChan <- fmt.Errorf("gRPC server failed: %v", err)
		}
	}()

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	i.logger.Info("gRPC server integration started successfully",
		zap.String("address", i.server.GetAddr()),
	)

	// Set up context cancellation handler
	go func() {
		select {
		case <-ctx.Done():
			i.logger.Info("Context cancelled, stopping gRPC server...")
			i.Stop()
		case err := <-errChan:
			i.logger.Error("gRPC server error", zap.Error(err))
		}
	}()

	return nil
}

// Stop stops the gRPC server integration
func (i *Integration) Stop() error {
	i.logger.Info("Stopping gRPC server integration...")

	if i.server == nil {
		i.logger.Warn("gRPC server is nil, nothing to stop")
		return nil
	}

	if err := i.server.Stop(); err != nil {
		i.logger.Error("Error stopping gRPC server", zap.Error(err))
		return err
	}

	i.logger.Info("gRPC server integration stopped")
	return nil
}

// GetServer returns the underlying gRPC server
func (i *Integration) GetServer() *Server {
	return i.server
}

// GetAddr returns the server address
func (i *Integration) GetAddr() string {
	return i.server.GetAddr()
}

// DefaultConfig returns a default gRPC server configuration
func DefaultConfig() Config {
	return Config{
		Port:             8444, // Default gRPC port as specified in the requirements
		TLSEnabled:       true,
		CertFile:         "cert/server.crt",
		KeyFile:          "cert/server.key",
		MTLSEnabled:      false,
		KeepalivePing:    15 * time.Second,
		KeepaliveTimeout: 10 * time.Second,
		MaxRecvMsgSize:   4 * 1024 * 1024, // 4MB
		MaxSendMsgSize:   4 * 1024 * 1024, // 4MB
	}
}

// TLSConfig returns a TLS-enabled configuration
func TLSConfig(certFile, keyFile string) Config {
	config := DefaultConfig()
	config.CertFile = certFile
	config.KeyFile = keyFile
	return config
}

// MTLSConfig returns a mutual TLS configuration
func MTLSConfig(certFile, keyFile string) Config {
	config := TLSConfig(certFile, keyFile)
	config.MTLSEnabled = true
	return config
}
