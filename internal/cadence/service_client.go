package cadence

import (
	"fmt"
	"net"
	"time"

	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/zap"
)

// ServiceClientOptions options for creating a workflow service client.
// Used by ai-worker path (NewClient); production path uses CreateWorkflowServiceClientImproved.
type ServiceClientOptions struct {
	Timeout     time.Duration
	ServiceName string
	EnableTLS   bool
	TLSCertPath string
	TLSKeyPath  string
	TLSCAPath   string
}

// CreateWorkflowServiceClient creates a workflowserviceclient.Interface for Cadence.
// Delegates to the improved factory so ai-worker and relay share one dial/create path.
// Returns (nil, nil) in stub mode when the host is unreachable (legacy behaviour).
func CreateWorkflowServiceClient(hostPort string, logger *zap.Logger) (workflowserviceclient.Interface, error) {
	if logger == nil {
		logger = zap.NewNop()
	}
	logger.Info("Creating workflow service client",
		zap.String("hostPort", hostPort),
		zap.String("factory", "shared→improved"),
	)
	return CreateWorkflowServiceClientImproved(hostPort, logger)
}

// CreateWorkflowServiceClientWithOptions creates a service client with extra options.
// TLS options are logged; transport still uses the shared improved factory until
// full TChannel TLS is implemented. Host probe uses Timeout when set.
func CreateWorkflowServiceClientWithOptions(hostPort string, options ServiceClientOptions, logger *zap.Logger) (workflowserviceclient.Interface, error) {
	if logger == nil {
		logger = zap.NewNop()
	}
	logger.Info("Creating workflow service client with options",
		zap.String("hostPort", hostPort),
		zap.Bool("enableTLS", options.EnableTLS),
		zap.String("serviceName", options.ServiceName),
		zap.String("factory", "shared→improved"),
	)

	// Optional pre-check with custom timeout (improved also probes with default).
	if options.Timeout > 0 {
		if err := checkHostAvailabilityWithTimeout(hostPort, options.Timeout, logger); err != nil {
			logger.Warn("Cadence server not available, falling back to stub mode",
				zap.String("hostPort", hostPort),
				zap.Error(err))
			return nil, nil
		}
	}

	return CreateWorkflowServiceClientImproved(hostPort, logger)
}

// checkHostAvailability probes Cadence frontend TCP reachability (5s default).
func checkHostAvailability(hostPort string, logger *zap.Logger) error {
	return checkHostAvailabilityWithTimeout(hostPort, 5*time.Second, logger)
}

func checkHostAvailabilityWithTimeout(hostPort string, timeout time.Duration, logger *zap.Logger) error {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	conn, err := net.DialTimeout("tcp", hostPort, timeout)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", hostPort, err)
	}
	defer conn.Close()

	if logger != nil {
		logger.Debug("Cadence server is available", zap.String("hostPort", hostPort))
	}
	return nil
}

// HealthCheck checks Cadence service client health (nil = stub mode).
// Prefer HealthCheckImproved when DescribeDomain is needed.
func HealthCheck(client workflowserviceclient.Interface, logger *zap.Logger) error {
	return HealthCheckImproved(client, logger)
}
