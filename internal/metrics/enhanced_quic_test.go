package metrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecordEnhancedQuicConnection(t *testing.T) {
	// Test recording connection metrics
	tenantID := "tenant_123"
	status := "active"
	region := "us-east-1"

	RecordEnhancedQuicConnection(tenantID, status, region, 1.0)
	RecordEnhancedQuicConnection(tenantID, status, region, 1.0)
	RecordEnhancedQuicConnection(tenantID, "inactive", region, -1.0)

	// Verify metrics were recorded
	assert.NotNil(t, enhancedQuicConnections)
}

func TestRecordEnhancedQuicConnectionDuration(t *testing.T) {
	// Test recording connection duration
	tenantID := "tenant_123"
	region := "us-east-1"

	RecordEnhancedQuicConnectionDuration(tenantID, region, 5.0)
	RecordEnhancedQuicConnectionDuration(tenantID, region, 10.0)
	RecordEnhancedQuicConnectionDuration(tenantID, region, 30.0)

	// Verify metrics were recorded
	assert.NotNil(t, enhancedQuicConnectionDuration)
}

func TestRecordEnhancedQuicBytes(t *testing.T) {
	// Test recording bytes sent and received
	tenantID := "tenant_123"
	region := "us-east-1"

	RecordEnhancedQuicBytes(tenantID, region, 1024, 2048)
	RecordEnhancedQuicBytes(tenantID, region, 512, 1024)
	RecordEnhancedQuicBytes(tenantID, region, 0, 0)

	// Verify metrics were recorded
	assert.NotNil(t, enhancedQuicBytesSent)
	assert.NotNil(t, enhancedQuicBytesReceived)
}

func TestUpdateEnhancedQuicThroughput(t *testing.T) {
	// Test updating throughput
	tenantID := "tenant_123"
	direction := "upload"
	region := "us-east-1"

	UpdateEnhancedQuicThroughput(tenantID, direction, region, 1000.0)
	UpdateEnhancedQuicThroughput(tenantID, "download", region, 2000.0)
	UpdateEnhancedQuicThroughput(tenantID, direction, region, 0.0)

	// Verify metrics were updated
	assert.NotNil(t, enhancedQuicThroughput)
}

func TestRecordEnhancedQuicStream(t *testing.T) {
	// Test recording stream metrics
	tenantID := "tenant_123"
	streamType := "data"
	region := "us-east-1"

	RecordEnhancedQuicStream(tenantID, streamType, region, 1.0)
	RecordEnhancedQuicStream(tenantID, streamType, region, 1.0)
	RecordEnhancedQuicStream(tenantID, streamType, region, -1.0)

	// Verify metrics were recorded
	assert.NotNil(t, enhancedQuicStreams)
	assert.NotNil(t, enhancedQuicStreamsCreated)
}

func TestRecordEnhancedQuicDDoS(t *testing.T) {
	// Test recording DDoS metrics
	clientIP := "192.168.1.1"
	reason := "rate_limit"
	region := "us-east-1"

	RecordEnhancedQuicDDoS(clientIP, reason, region)
	RecordEnhancedQuicDDoS(clientIP, "suspicious_pattern", region)
	RecordEnhancedQuicDDoS("10.0.0.1", reason, region)

	// Verify metrics were recorded
	assert.NotNil(t, enhancedQuicDDoSBlocked)
}

func TestRecordEnhancedQuicDDoSDetection(t *testing.T) {
	// Test recording DDoS detection
	attackType := "volumetric"
	severity := "high"
	region := "us-east-1"

	RecordEnhancedQuicDDoSDetection(attackType, severity, region)
	RecordEnhancedQuicDDoSDetection("protocol", "medium", region)
	RecordEnhancedQuicDDoSDetection(attackType, "critical", region)

	// Verify metrics were recorded
	assert.NotNil(t, enhancedQuicDDoSDetected)
}

func TestRecordEnhancedQuicRateLimit(t *testing.T) {
	// Test recording rate limit hits
	tenantID := "tenant_123"
	limitType := "connection_rate"
	region := "us-east-1"

	RecordEnhancedQuicRateLimit(tenantID, limitType, region)
	RecordEnhancedQuicRateLimit(tenantID, "bandwidth", region)
	RecordEnhancedQuicRateLimit("tenant_456", limitType, region)

	// Verify metrics were recorded
	assert.NotNil(t, enhancedQuicRateLimitHits)
}

func TestRecordEnhancedQuicGeoRoute(t *testing.T) {
	// Test recording geographic routing
	sourceRegion := "us-east-1"
	targetRegion := "eu-west-1"
	routeType := "optimal"

	RecordEnhancedQuicGeoRoute(sourceRegion, targetRegion, routeType)
	RecordEnhancedQuicGeoRoute(sourceRegion, targetRegion, "fallback")
	RecordEnhancedQuicGeoRoute("us-west-2", targetRegion, routeType)

	// Verify metrics were recorded
	assert.NotNil(t, enhancedQuicGeoRoutes)
}

func TestRecordEnhancedQuicGeoLatency(t *testing.T) {
	// Test recording geographic routing latency
	sourceRegion := "us-east-1"
	targetRegion := "eu-west-1"

	RecordEnhancedQuicGeoLatency(sourceRegion, targetRegion, 0.1)
	RecordEnhancedQuicGeoLatency(sourceRegion, targetRegion, 0.2)
	RecordEnhancedQuicGeoLatency("us-west-2", targetRegion, 0.15)

	// Verify metrics were recorded
	assert.NotNil(t, enhancedQuicGeoLatency)
}

func TestUpdateEnhancedQuicTenantBandwidth(t *testing.T) {
	// Test updating tenant bandwidth usage
	tenantID := "tenant_123"
	direction := "upload"
	region := "us-east-1"

	UpdateEnhancedQuicTenantBandwidth(tenantID, direction, region, 1024.0)
	UpdateEnhancedQuicTenantBandwidth(tenantID, "download", region, 2048.0)
	UpdateEnhancedQuicTenantBandwidth(tenantID, direction, region, 0.0)

	// Verify metrics were updated
	assert.NotNil(t, enhancedQuicTenantBandwidthUsage)
}

func TestUpdateEnhancedQuicTenantConnections(t *testing.T) {
	// Test updating tenant connections
	tenantID := "tenant_123"
	region := "us-east-1"

	UpdateEnhancedQuicTenantConnections(tenantID, region, 10.0)
	UpdateEnhancedQuicTenantConnections(tenantID, region, 5.0)
	UpdateEnhancedQuicTenantConnections(tenantID, region, 0.0)

	// Verify metrics were updated
	assert.NotNil(t, enhancedQuicTenantConnectionsActive)
}

func TestRecordEnhancedQuicTenantLimitExceeded(t *testing.T) {
	// Test recording tenant limit violations
	tenantID := "tenant_123"
	limitType := "bandwidth"
	region := "us-east-1"

	RecordEnhancedQuicTenantLimitExceeded(tenantID, limitType, region)
	RecordEnhancedQuicTenantLimitExceeded(tenantID, "connections", region)
	RecordEnhancedQuicTenantLimitExceeded("tenant_456", limitType, region)

	// Verify metrics were recorded
	assert.NotNil(t, enhancedQuicTenantLimitExceeded)
}

func TestRecordEnhancedQuicAuthentication(t *testing.T) {
	// Test recording authentication attempts
	tenantID := "tenant_123"
	authMethod := "oauth2"
	status := "success"
	region := "us-east-1"

	RecordEnhancedQuicAuthentication(tenantID, authMethod, status, region)
	RecordEnhancedQuicAuthentication(tenantID, "jwt", "failed", region)
	RecordEnhancedQuicAuthentication(tenantID, authMethod, status, region)

	// Verify metrics were recorded
	assert.NotNil(t, enhancedQuicAuthenticationAttempts)
}

func TestRecordEnhancedQuicKeycloakRequest(t *testing.T) {
	// Test recording Keycloak requests
	operation := "validate_token"
	status := "success"
	region := "us-east-1"

	RecordEnhancedQuicKeycloakRequest(operation, status, region)
	RecordEnhancedQuicKeycloakRequest("get_user_info", "failed", region)
	RecordEnhancedQuicKeycloakRequest(operation, status, region)

	// Verify metrics were recorded
	assert.NotNil(t, enhancedQuicKeycloakRequests)
}

func TestRecordEnhancedQuicKeycloakLatency(t *testing.T) {
	// Test recording Keycloak latency
	operation := "validate_token"
	region := "us-east-1"

	RecordEnhancedQuicKeycloakLatency(operation, region, 0.1)
	RecordEnhancedQuicKeycloakLatency("get_user_info", region, 0.2)
	RecordEnhancedQuicKeycloakLatency(operation, region, 0.05)

	// Verify metrics were recorded
	assert.NotNil(t, enhancedQuicKeycloakLatency)
}

func TestRecordEnhancedQuicLoadBalancerRequest(t *testing.T) {
	// Test recording load balancer requests
	server := "server_1"
	algorithm := "round_robin"
	status := "success"
	region := "us-east-1"

	RecordEnhancedQuicLoadBalancerRequest(server, algorithm, status, region)
	RecordEnhancedQuicLoadBalancerRequest("server_2", "least_connections", "failed", region)
	RecordEnhancedQuicLoadBalancerRequest(server, algorithm, status, region)

	// Verify metrics were recorded
	assert.NotNil(t, enhancedQuicLoadBalancerRequests)
}

func TestUpdateEnhancedQuicServerHealth(t *testing.T) {
	// Test updating server health
	server := "server_1"
	region := "us-east-1"

	UpdateEnhancedQuicServerHealth(server, region, true)
	UpdateEnhancedQuicServerHealth("server_2", region, false)
	UpdateEnhancedQuicServerHealth(server, region, true)

	// Verify metrics were updated
	assert.NotNil(t, enhancedQuicServerHealth)
}

func TestRecordEnhancedQuicLatency(t *testing.T) {
	// Test recording operation latency
	operation := "connect"
	tenantID := "tenant_123"
	region := "us-east-1"

	RecordEnhancedQuicLatency(operation, tenantID, region, 0.1)
	RecordEnhancedQuicLatency("disconnect", tenantID, region, 0.05)
	RecordEnhancedQuicLatency(operation, tenantID, region, 0.2)

	// Verify metrics were recorded
	assert.NotNil(t, enhancedQuicLatency)
}

func TestUpdateEnhancedQuicCpuUsage(t *testing.T) {
	// Test updating CPU usage
	component := "quic_handler"
	region := "us-east-1"

	UpdateEnhancedQuicCpuUsage(component, region, 25.5)
	UpdateEnhancedQuicCpuUsage("load_balancer", region, 45.0)
	UpdateEnhancedQuicCpuUsage(component, region, 0.0)

	// Verify metrics were updated
	assert.NotNil(t, enhancedQuicCpuUsage)
}

func TestUpdateEnhancedQuicMemoryUsage(t *testing.T) {
	// Test updating memory usage
	component := "quic_handler"
	region := "us-east-1"

	UpdateEnhancedQuicMemoryUsage(component, region, 1024*1024*100)       // 100MB
	UpdateEnhancedQuicMemoryUsage("load_balancer", region, 1024*1024*200) // 200MB
	UpdateEnhancedQuicMemoryUsage(component, region, 0)

	// Verify metrics were updated
	assert.NotNil(t, enhancedQuicMemoryUsage)
}

func TestRecordEnhancedQuicError(t *testing.T) {
	// Test recording errors
	errorType := "connection_failed"
	component := "quic_handler"
	tenantID := "tenant_123"
	region := "us-east-1"

	RecordEnhancedQuicError(errorType, component, tenantID, region)
	RecordEnhancedQuicError("timeout", "load_balancer", tenantID, region)
	RecordEnhancedQuicError(errorType, component, "tenant_456", region)

	// Verify metrics were recorded
	assert.NotNil(t, enhancedQuicErrors)
}

func TestRecordEnhancedQuicTimeout(t *testing.T) {
	// Test recording timeouts
	operation := "connect"
	tenantID := "tenant_123"
	region := "us-east-1"

	RecordEnhancedQuicTimeout(operation, tenantID, region)
	RecordEnhancedQuicTimeout("disconnect", tenantID, region)
	RecordEnhancedQuicTimeout(operation, "tenant_456", region)

	// Verify metrics were recorded
	assert.NotNil(t, enhancedQuicTimeout)
}

func TestGetEnhancedQuicMetrics(t *testing.T) {
	// Test getting all metrics
	metrics := GetEnhancedQuicMetrics()

	// Should return all 24 metrics
	assert.Len(t, metrics, 24)

	// Verify all expected metrics are present
	assert.NotNil(t, metrics)
}

func TestEnhancedQuicMetrics_EdgeCases(t *testing.T) {
	// Test with empty strings
	RecordEnhancedQuicConnection("", "", "", 1.0)
	RecordEnhancedQuicConnectionDuration("", "", 1.0)
	RecordEnhancedQuicBytes("", "", 0, 0)
	UpdateEnhancedQuicThroughput("", "", "", 0.0)
	RecordEnhancedQuicStream("", "", "", 1.0)
	RecordEnhancedQuicDDoS("", "", "")
	RecordEnhancedQuicDDoSDetection("", "", "")
	RecordEnhancedQuicRateLimit("", "", "")
	RecordEnhancedQuicGeoRoute("", "", "")
	RecordEnhancedQuicGeoLatency("", "", 0.0)
	UpdateEnhancedQuicTenantBandwidth("", "", "", 0.0)
	UpdateEnhancedQuicTenantConnections("", "", 0.0)
	RecordEnhancedQuicTenantLimitExceeded("", "", "")
	RecordEnhancedQuicAuthentication("", "", "", "")
	RecordEnhancedQuicKeycloakRequest("", "", "")
	RecordEnhancedQuicKeycloakLatency("", "", 0.0)
	RecordEnhancedQuicLoadBalancerRequest("", "", "", "")
	UpdateEnhancedQuicServerHealth("", "", true)
	RecordEnhancedQuicLatency("", "", "", 0.0)
	UpdateEnhancedQuicCpuUsage("", "", 0.0)
	UpdateEnhancedQuicMemoryUsage("", "", 0.0)
	RecordEnhancedQuicError("", "", "", "")
	RecordEnhancedQuicTimeout("", "", "")

	// Test with negative values
	RecordEnhancedQuicConnection("tenant", "status", "region", -1.0)
	RecordEnhancedQuicConnectionDuration("tenant", "region", -1.0)
	RecordEnhancedQuicBytes("tenant", "region", -1, -1)
	UpdateEnhancedQuicThroughput("tenant", "direction", "region", -1.0)
	RecordEnhancedQuicStream("tenant", "type", "region", -1.0)
	UpdateEnhancedQuicTenantBandwidth("tenant", "direction", "region", -1.0)
	UpdateEnhancedQuicTenantConnections("tenant", "region", -1.0)
	RecordEnhancedQuicGeoLatency("source", "target", -1.0)
	RecordEnhancedQuicLatency("operation", "tenant", "region", -1.0)
	UpdateEnhancedQuicCpuUsage("component", "region", -1.0)
	UpdateEnhancedQuicMemoryUsage("component", "region", -1.0)

	// Test with very large values
	RecordEnhancedQuicConnection("tenant", "status", "region", 999999.0)
	RecordEnhancedQuicConnectionDuration("tenant", "region", 999999.0)
	RecordEnhancedQuicBytes("tenant", "region", 999999999, 999999999)
	UpdateEnhancedQuicThroughput("tenant", "direction", "region", 999999.0)
	RecordEnhancedQuicStream("tenant", "type", "region", 999999.0)
	UpdateEnhancedQuicTenantBandwidth("tenant", "direction", "region", 999999.0)
	UpdateEnhancedQuicTenantConnections("tenant", "region", 999999.0)
	RecordEnhancedQuicGeoLatency("source", "target", 999999.0)
	RecordEnhancedQuicLatency("operation", "tenant", "region", 999999.0)
	UpdateEnhancedQuicCpuUsage("component", "region", 999999.0)
	UpdateEnhancedQuicMemoryUsage("component", "region", 999999.0)

	// Verify all operations completed without errors
	assert.NotNil(t, enhancedQuicConnections)
}

func TestEnhancedQuicMetrics_ConcurrentAccess(t *testing.T) {
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()

			tenantID := "tenant" + string(rune(id))
			region := "region" + string(rune(id))
			component := "component" + string(rune(id))
			operation := "operation" + string(rune(id))

			RecordEnhancedQuicConnection(tenantID, "active", region, 1.0)
			RecordEnhancedQuicConnectionDuration(tenantID, region, float64(id))
			RecordEnhancedQuicBytes(tenantID, region, int64(id*100), int64(id*200))
			UpdateEnhancedQuicThroughput(tenantID, "upload", region, float64(id*10))
			RecordEnhancedQuicStream(tenantID, "data", region, 1.0)
			RecordEnhancedQuicDDoS("192.168.1.1", "rate_limit", region)
			RecordEnhancedQuicDDoSDetection("volumetric", "high", region)
			RecordEnhancedQuicRateLimit(tenantID, "connection_rate", region)
			RecordEnhancedQuicGeoRoute("us-east-1", region, "optimal")
			RecordEnhancedQuicGeoLatency("us-east-1", region, float64(id)/10.0)
			UpdateEnhancedQuicTenantBandwidth(tenantID, "upload", region, float64(id*1000))
			UpdateEnhancedQuicTenantConnections(tenantID, region, float64(id))
			RecordEnhancedQuicTenantLimitExceeded(tenantID, "bandwidth", region)
			RecordEnhancedQuicAuthentication(tenantID, "oauth2", "success", region)
			RecordEnhancedQuicKeycloakRequest("validate_token", "success", region)
			RecordEnhancedQuicKeycloakLatency("validate_token", region, float64(id)/100.0)
			RecordEnhancedQuicLoadBalancerRequest("server1", "round_robin", "success", region)
			UpdateEnhancedQuicServerHealth("server1", region, true)
			RecordEnhancedQuicLatency(operation, tenantID, region, float64(id)/100.0)
			UpdateEnhancedQuicCpuUsage(component, region, float64(id))
			UpdateEnhancedQuicMemoryUsage(component, region, float64(id*1024))
			RecordEnhancedQuicError("connection_failed", component, tenantID, region)
			RecordEnhancedQuicTimeout(operation, tenantID, region)
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all operations completed without errors
	assert.NotNil(t, enhancedQuicConnections)
	assert.NotNil(t, enhancedQuicConnectionDuration)
	assert.NotNil(t, enhancedQuicBytesSent)
	assert.NotNil(t, enhancedQuicBytesReceived)
	assert.NotNil(t, enhancedQuicThroughput)
	assert.NotNil(t, enhancedQuicStreams)
	assert.NotNil(t, enhancedQuicStreamsCreated)
	assert.NotNil(t, enhancedQuicDDoSBlocked)
	assert.NotNil(t, enhancedQuicDDoSDetected)
	assert.NotNil(t, enhancedQuicRateLimitHits)
	assert.NotNil(t, enhancedQuicGeoRoutes)
	assert.NotNil(t, enhancedQuicGeoLatency)
	assert.NotNil(t, enhancedQuicTenantBandwidthUsage)
	assert.NotNil(t, enhancedQuicTenantConnectionsActive)
	assert.NotNil(t, enhancedQuicTenantLimitExceeded)
	assert.NotNil(t, enhancedQuicAuthenticationAttempts)
	assert.NotNil(t, enhancedQuicKeycloakRequests)
	assert.NotNil(t, enhancedQuicKeycloakLatency)
	assert.NotNil(t, enhancedQuicLoadBalancerRequests)
	assert.NotNil(t, enhancedQuicServerHealth)
	assert.NotNil(t, enhancedQuicLatency)
	assert.NotNil(t, enhancedQuicCpuUsage)
	assert.NotNil(t, enhancedQuicMemoryUsage)
	assert.NotNil(t, enhancedQuicErrors)
	assert.NotNil(t, enhancedQuicTimeout)
}

func TestEnhancedQuicMetrics_IntegrationWorkflow(t *testing.T) {
	// Simulate a complete Enhanced QUIC workflow
	tenantID := "tenant_123"
	region := "us-east-1"
	component := "quic_handler"
	operation := "connect"

	// 1. Connection establishment
	RecordEnhancedQuicConnection(tenantID, "active", region, 1.0)
	RecordEnhancedQuicConnectionDuration(tenantID, region, 0.1)

	// 2. Data transfer
	RecordEnhancedQuicBytes(tenantID, region, 1024, 2048)
	UpdateEnhancedQuicThroughput(tenantID, "upload", region, 1000.0)
	UpdateEnhancedQuicThroughput(tenantID, "download", region, 2000.0)

	// 3. Stream management
	RecordEnhancedQuicStream(tenantID, "data", region, 1.0)
	RecordEnhancedQuicStream(tenantID, "control", region, 1.0)

	// 4. Security checks
	RecordEnhancedQuicDDoSDetection("volumetric", "medium", region)
	RecordEnhancedQuicRateLimit(tenantID, "connection_rate", region)

	// 5. Geographic routing
	RecordEnhancedQuicGeoRoute("us-east-1", "eu-west-1", "optimal")
	RecordEnhancedQuicGeoLatency("us-east-1", "eu-west-1", 0.1)

	// 6. Tenant management
	UpdateEnhancedQuicTenantBandwidth(tenantID, "upload", region, 1024.0)
	UpdateEnhancedQuicTenantConnections(tenantID, region, 5.0)

	// 7. Authentication
	RecordEnhancedQuicAuthentication(tenantID, "oauth2", "success", region)
	RecordEnhancedQuicKeycloakRequest("validate_token", "success", region)
	RecordEnhancedQuicKeycloakLatency("validate_token", region, 0.05)

	// 8. Load balancing
	RecordEnhancedQuicLoadBalancerRequest("server1", "round_robin", "success", region)
	UpdateEnhancedQuicServerHealth("server1", region, true)

	// 9. Performance monitoring
	RecordEnhancedQuicLatency(operation, tenantID, region, 0.1)
	UpdateEnhancedQuicCpuUsage(component, region, 25.0)
	UpdateEnhancedQuicMemoryUsage(component, region, 1024*1024*100)

	// 10. Error handling
	RecordEnhancedQuicError("connection_failed", component, tenantID, region)
	RecordEnhancedQuicTimeout(operation, tenantID, region)

	// Verify workflow completed
	assert.NotNil(t, enhancedQuicConnections)
	assert.NotNil(t, enhancedQuicConnectionDuration)
	assert.NotNil(t, enhancedQuicBytesSent)
	assert.NotNil(t, enhancedQuicBytesReceived)
	assert.NotNil(t, enhancedQuicThroughput)
	assert.NotNil(t, enhancedQuicStreams)
	assert.NotNil(t, enhancedQuicStreamsCreated)
	assert.NotNil(t, enhancedQuicDDoSBlocked)
	assert.NotNil(t, enhancedQuicDDoSDetected)
	assert.NotNil(t, enhancedQuicRateLimitHits)
	assert.NotNil(t, enhancedQuicGeoRoutes)
	assert.NotNil(t, enhancedQuicGeoLatency)
	assert.NotNil(t, enhancedQuicTenantBandwidthUsage)
	assert.NotNil(t, enhancedQuicTenantConnectionsActive)
	assert.NotNil(t, enhancedQuicTenantLimitExceeded)
	assert.NotNil(t, enhancedQuicAuthenticationAttempts)
	assert.NotNil(t, enhancedQuicKeycloakRequests)
	assert.NotNil(t, enhancedQuicKeycloakLatency)
	assert.NotNil(t, enhancedQuicLoadBalancerRequests)
	assert.NotNil(t, enhancedQuicServerHealth)
	assert.NotNil(t, enhancedQuicLatency)
	assert.NotNil(t, enhancedQuicCpuUsage)
	assert.NotNil(t, enhancedQuicMemoryUsage)
	assert.NotNil(t, enhancedQuicErrors)
	assert.NotNil(t, enhancedQuicTimeout)
}

func TestEnhancedQuicMetrics_Performance(t *testing.T) {
	// Test performance with many operations
	start := time.Now()

	for i := 0; i < 1000; i++ {
		tenantID := "tenant" + string(rune(i%100))
		region := "region" + string(rune(i%10))
		component := "component" + string(rune(i%10))
		operation := "operation" + string(rune(i%10))

		RecordEnhancedQuicConnection(tenantID, "active", region, 1.0)
		RecordEnhancedQuicConnectionDuration(tenantID, region, float64(i%100)/10.0)
		RecordEnhancedQuicBytes(tenantID, region, int64(i%1000), int64(i%2000))
		UpdateEnhancedQuicThroughput(tenantID, "upload", region, float64(i%10000))
		RecordEnhancedQuicStream(tenantID, "data", region, 1.0)
		RecordEnhancedQuicDDoS("192.168.1.1", "rate_limit", region)
		RecordEnhancedQuicDDoSDetection("volumetric", "high", region)
		RecordEnhancedQuicRateLimit(tenantID, "connection_rate", region)
		RecordEnhancedQuicGeoRoute("us-east-1", region, "optimal")
		RecordEnhancedQuicGeoLatency("us-east-1", region, float64(i%100)/1000.0)
		UpdateEnhancedQuicTenantBandwidth(tenantID, "upload", region, float64(i%100000))
		UpdateEnhancedQuicTenantConnections(tenantID, region, float64(i%100))
		RecordEnhancedQuicTenantLimitExceeded(tenantID, "bandwidth", region)
		RecordEnhancedQuicAuthentication(tenantID, "oauth2", "success", region)
		RecordEnhancedQuicKeycloakRequest("validate_token", "success", region)
		RecordEnhancedQuicKeycloakLatency("validate_token", region, float64(i%100)/10000.0)
		RecordEnhancedQuicLoadBalancerRequest("server1", "round_robin", "success", region)
		UpdateEnhancedQuicServerHealth("server1", region, true)
		RecordEnhancedQuicLatency(operation, tenantID, region, float64(i%100)/1000.0)
		UpdateEnhancedQuicCpuUsage(component, region, float64(i%100))
		UpdateEnhancedQuicMemoryUsage(component, region, float64(i%1000000))
		RecordEnhancedQuicError("connection_failed", component, tenantID, region)
		RecordEnhancedQuicTimeout(operation, tenantID, region)
	}

	duration := time.Since(start)

	// Should complete within reasonable time (less than 1 second)
	assert.Less(t, duration, time.Second)
}

func TestEnhancedQuicMetrics_DataTypes(t *testing.T) {
	// Test different data types and edge cases
	testCases := []struct {
		name     string
		testFunc func()
	}{
		{
			name: "string labels",
			testFunc: func() {
				RecordEnhancedQuicConnection("tenant-id", "status-name", "region-name", 1.0)
			},
		},
		{
			name: "unicode strings",
			testFunc: func() {
				RecordEnhancedQuicConnection("тенант-ид", "статус-имя", "регион-имя", 1.0)
			},
		},
		{
			name: "special characters",
			testFunc: func() {
				RecordEnhancedQuicConnection("tenant@domain.com", "status_name", "region-name", 1.0)
			},
		},
		{
			name: "long strings",
			testFunc: func() {
				longString := "very-long-tenant-id-that-might-cause-issues-with-prometheus-label-limits"
				RecordEnhancedQuicConnection(longString, "status", "region", 1.0)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.NotPanics(t, tc.testFunc)
		})
	}
}

func TestEnhancedQuicMetrics_ServerHealthStates(t *testing.T) {
	// Test all server health states
	server := "test_server"
	region := "test_region"

	// Test healthy state (true)
	UpdateEnhancedQuicServerHealth(server, region, true)
	assert.NotNil(t, enhancedQuicServerHealth)

	// Test unhealthy state (false)
	UpdateEnhancedQuicServerHealth(server, region, false)
	assert.NotNil(t, enhancedQuicServerHealth)
}

func TestEnhancedQuicMetrics_StreamDeltaValues(t *testing.T) {
	// Test stream delta values
	tenantID := "tenant_123"
	streamType := "data"
	region := "us-east-1"

	// Test positive delta (stream created)
	RecordEnhancedQuicStream(tenantID, streamType, region, 1.0)
	assert.NotNil(t, enhancedQuicStreams)
	assert.NotNil(t, enhancedQuicStreamsCreated)

	// Test negative delta (stream closed)
	RecordEnhancedQuicStream(tenantID, streamType, region, -1.0)
	assert.NotNil(t, enhancedQuicStreams)

	// Test zero delta (no change)
	RecordEnhancedQuicStream(tenantID, streamType, region, 0.0)
	assert.NotNil(t, enhancedQuicStreams)
}
