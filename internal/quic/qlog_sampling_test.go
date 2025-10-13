package quic

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewQLogSampler(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tmpDir := t.TempDir()

	config := &QLogSamplingConfig{
		SamplingRate:  0.01,
		GlobalEnabled: false,
		OutputDir:     tmpDir,
	}

	sampler, err := NewQLogSampler(logger, config)
	require.NoError(t, err)
	require.NotNil(t, sampler)
	defer sampler.Close()

	assert.Equal(t, 0.01, sampler.samplingRate)
	assert.False(t, sampler.globalEnabled)
	assert.True(t, sampler.redactIPv4)
	assert.True(t, sampler.redactIPv6)
}

func TestNewQLogSampler_DefaultConfig(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	// Nil config should use defaults
	sampler, err := NewQLogSampler(logger, nil)
	require.NoError(t, err)
	require.NotNil(t, sampler)
	defer sampler.Close()

	assert.Equal(t, 0.01, sampler.samplingRate) // 1% default
	assert.False(t, sampler.globalEnabled)       // OFF by default
	assert.True(t, sampler.redactIPv4)
	assert.True(t, sampler.redactIPv6)
	assert.True(t, sampler.redactTokens)
}

func TestShouldSample_GlobalDisabled(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tmpDir := t.TempDir()

	config := &QLogSamplingConfig{
		SamplingRate:  1.0, // 100% sampling rate
		GlobalEnabled: false, // But globally disabled
		OutputDir:     tmpDir,
	}

	sampler, err := NewQLogSampler(logger, config)
	require.NoError(t, err)
	defer sampler.Close()

	// Should not sample when globally disabled
	shouldSample := sampler.ShouldSample("conn-1", "tenant-1")
	assert.False(t, shouldSample)
}

func TestShouldSample_TenantOverride(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tmpDir := t.TempDir()

	config := &QLogSamplingConfig{
		SamplingRate:  1.0,
		GlobalEnabled: false,
		OutputDir:     tmpDir,
		EnabledForTenants: map[string]bool{
			"premium-tenant": true,
		},
	}

	sampler, err := NewQLogSampler(logger, config)
	require.NoError(t, err)
	defer sampler.Close()

	// Premium tenant should be sampled (tenant override)
	shouldSample := sampler.ShouldSample("conn-1", "premium-tenant")
	assert.True(t, shouldSample)

	// Regular tenant should not be sampled
	shouldSample = sampler.ShouldSample("conn-2", "regular-tenant")
	assert.False(t, shouldSample)
}

func TestShouldSample_SamplingRate(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tmpDir := t.TempDir()

	config := &QLogSamplingConfig{
		SamplingRate:  0.5, // 50% sampling
		GlobalEnabled: true,
		OutputDir:     tmpDir,
	}

	sampler, err := NewQLogSampler(logger, config)
	require.NoError(t, err)
	defer sampler.Close()

	// Test multiple connections
	sampledCount := 0
	totalCount := 1000

	for i := 0; i < totalCount; i++ {
		connectionID := fmt.Sprintf("conn-%d", i)
		if sampler.ShouldSample(connectionID, "tenant-1") {
			sampledCount++
		}
	}

	// Should be approximately 50% (allow 10% tolerance)
	expectedSampled := float64(totalCount) * 0.5
	tolerance := float64(totalCount) * 0.1

	assert.InDelta(t, expectedSampled, sampledCount, tolerance,
		"Expected ~%d sampled, got %d", int(expectedSampled), sampledCount)
}

func TestShouldSample_Deterministic(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tmpDir := t.TempDir()

	config := &QLogSamplingConfig{
		SamplingRate:  0.5,
		GlobalEnabled: true,
		OutputDir:     tmpDir,
	}

	sampler, err := NewQLogSampler(logger, config)
	require.NoError(t, err)
	defer sampler.Close()

	// Same connection should always get same sampling decision
	connectionID := "test-connection-123"

	result1 := sampler.ShouldSample(connectionID, "tenant-1")
	result2 := sampler.ShouldSample(connectionID, "tenant-1")
	result3 := sampler.ShouldSample(connectionID, "tenant-1")

	assert.Equal(t, result1, result2)
	assert.Equal(t, result2, result3)
}

func TestStartCapture(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tmpDir := t.TempDir()

	config := &QLogSamplingConfig{
		SamplingRate:  1.0, // 100% for testing
		GlobalEnabled: true,
		OutputDir:     tmpDir,
	}

	sampler, err := NewQLogSampler(logger, config)
	require.NoError(t, err)
	defer sampler.Close()

	writer, err := sampler.StartCapture("conn-1", "tenant-1")
	require.NoError(t, err)
	require.NotNil(t, writer)

	assert.Equal(t, "conn-1", writer.connectionID)
	assert.Equal(t, "tenant-1", writer.tenantID)
	assert.NotNil(t, writer.file)

	// Verify file was created
	files, err := os.ReadDir(tmpDir)
	require.NoError(t, err)
	assert.Greater(t, len(files), 0)
}

func TestStartCapture_NotSampled(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tmpDir := t.TempDir()

	config := &QLogSamplingConfig{
		SamplingRate:  0.0, // 0% sampling
		GlobalEnabled: true,
		OutputDir:     tmpDir,
	}

	sampler, err := NewQLogSampler(logger, config)
	require.NoError(t, err)
	defer sampler.Close()

	writer, err := sampler.StartCapture("conn-1", "tenant-1")
	require.NoError(t, err)
	assert.Nil(t, writer) // Should be nil when not sampled

	// Verify no files were created
	files, err := os.ReadDir(tmpDir)
	require.NoError(t, err)
	assert.Equal(t, 0, len(files))
}

func TestLogEvent(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tmpDir := t.TempDir()

	config := &QLogSamplingConfig{
		SamplingRate:  1.0,
		GlobalEnabled: true,
		OutputDir:     tmpDir,
		RedactIPv4:    true,
		RedactIPv6:    true,
		RedactTokens:  true,
	}

	sampler, err := NewQLogSampler(logger, config)
	require.NoError(t, err)
	defer sampler.Close()

	writer, err := sampler.StartCapture("conn-1", "tenant-1")
	require.NoError(t, err)
	require.NotNil(t, writer)

	// Log event
	event := QLogEvent{
		Time:     time.Now(),
		Event:    "connectivity:connection_started",
		Category: "connectivity",
		Data: map[string]interface{}{
			"src_ip":   "192.168.1.100",
			"dst_ip":   "10.0.0.1",
			"src_port": 12345,
			"dst_port": 443,
		},
	}

	err = sampler.LogEvent("conn-1", event)
	require.NoError(t, err)

	// Stop capture to flush
	err = sampler.StopCapture("conn-1")
	require.NoError(t, err)

	// Verify file contains event
	files, err := os.ReadDir(tmpDir)
	require.NoError(t, err)
	assert.Equal(t, 1, len(files))
}

func TestLogEvent_NotSampled(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tmpDir := t.TempDir()

	config := &QLogSamplingConfig{
		SamplingRate:  0.0, // Not sampled
		GlobalEnabled: true,
		OutputDir:     tmpDir,
	}

	sampler, err := NewQLogSampler(logger, config)
	require.NoError(t, err)
	defer sampler.Close()

	// Try to log event for non-sampled connection
	event := QLogEvent{
		Time:  time.Now(),
		Event: "test_event",
	}

	err = sampler.LogEvent("conn-1", event)
	require.NoError(t, err) // Should not error, just skip

	// No files should be created
	files, err := os.ReadDir(tmpDir)
	require.NoError(t, err)
	assert.Equal(t, 0, len(files))
}

func TestRedactPII_IPv4(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tmpDir := t.TempDir()

	config := &QLogSamplingConfig{
		SamplingRate:  1.0,
		GlobalEnabled: true,
		OutputDir:     tmpDir,
		RedactIPv4:    true,
		PIIHashSalt:   "test-salt",
	}

	sampler, err := NewQLogSampler(logger, config)
	require.NoError(t, err)
	defer sampler.Close()

	data := map[string]interface{}{
		"src_ip": "192.168.1.100",
		"dst_ip": "10.0.0.1",
	}

	redacted := sampler.redactPII(data)

	// IPs should be hashed
	assert.Contains(t, redacted["src_ip"].(string), "hash:")
	assert.Contains(t, redacted["dst_ip"].(string), "hash:")
	assert.NotEqual(t, "192.168.1.100", redacted["src_ip"])
	assert.NotEqual(t, "10.0.0.1", redacted["dst_ip"])
}

func TestRedactPII_Ports(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tmpDir := t.TempDir()

	config := &QLogSamplingConfig{
		SamplingRate:  1.0,
		GlobalEnabled: true,
		OutputDir:     tmpDir,
		RedactPorts:   true,
	}

	sampler, err := NewQLogSampler(logger, config)
	require.NoError(t, err)
	defer sampler.Close()

	data := map[string]interface{}{
		"src_port": 12345,
		"dst_port": 443,
	}

	redacted := sampler.redactPII(data)

	// Ports should be redacted
	assert.Equal(t, "REDACTED", redacted["src_port"])
	assert.Equal(t, "REDACTED", redacted["dst_port"])
}

func TestRedactPII_Tokens(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tmpDir := t.TempDir()

	config := &QLogSamplingConfig{
		SamplingRate:  1.0,
		GlobalEnabled: true,
		OutputDir:     tmpDir,
		RedactTokens:  true,
		PIIHashSalt:   "test-salt",
	}

	sampler, err := NewQLogSampler(logger, config)
	require.NoError(t, err)
	defer sampler.Close()

	data := map[string]interface{}{
		"token":         "secret-token-123",
		"session_ticket": "ticket-456",
		"resume_token":  "resume-789",
	}

	redacted := sampler.redactPII(data)

	// Tokens should be hashed
	assert.Contains(t, redacted["token"].(string), "hash:")
	assert.Contains(t, redacted["session_ticket"].(string), "hash:")
	assert.Contains(t, redacted["resume_token"].(string), "hash:")
}

func TestRedactPII_ConnectionID(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tmpDir := t.TempDir()

	config := &QLogSamplingConfig{
		SamplingRate:  1.0,
		GlobalEnabled: true,
		OutputDir:     tmpDir,
		PIIHashSalt:   "test-salt",
	}

	sampler, err := NewQLogSampler(logger, config)
	require.NoError(t, err)
	defer sampler.Close()

	data := map[string]interface{}{
		"connection_id": "conn-123-456-789",
		"dcid":          "dest-conn-id",
		"scid":          "src-conn-id",
	}

	redacted := sampler.redactPII(data)

	// Connection IDs always hashed
	assert.Contains(t, redacted["connection_id"].(string), "hash:")
	assert.Contains(t, redacted["dcid"].(string), "hash:")
	assert.Contains(t, redacted["scid"].(string), "hash:")
}

func TestRedactPII_NestedData(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tmpDir := t.TempDir()

	config := &QLogSamplingConfig{
		SamplingRate:  1.0,
		GlobalEnabled: true,
		OutputDir:     tmpDir,
		RedactIPv4:    true,
		PIIHashSalt:   "test-salt",
	}

	sampler, err := NewQLogSampler(logger, config)
	require.NoError(t, err)
	defer sampler.Close()

	data := map[string]interface{}{
		"header": map[string]interface{}{
			"src_ip": "192.168.1.100",
			"dst_ip": "10.0.0.1",
		},
		"normal_field": "not-redacted",
	}

	redacted := sampler.redactPII(data)

	// Nested IPs should be hashed
	header := redacted["header"].(map[string]interface{})
	assert.Contains(t, header["src_ip"].(string), "hash:")
	assert.Contains(t, header["dst_ip"].(string), "hash:")

	// Normal field should not be touched
	assert.Equal(t, "not-redacted", redacted["normal_field"])
}

func TestStopCapture(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tmpDir := t.TempDir()

	config := &QLogSamplingConfig{
		SamplingRate:  1.0,
		GlobalEnabled: true,
		OutputDir:     tmpDir,
	}

	sampler, err := NewQLogSampler(logger, config)
	require.NoError(t, err)
	defer sampler.Close()

	writer, err := sampler.StartCapture("conn-1", "tenant-1")
	require.NoError(t, err)
	require.NotNil(t, writer)

	// Log some events
	for i := 0; i < 10; i++ {
		event := QLogEvent{
			Time:  time.Now(),
			Event: fmt.Sprintf("event_%d", i),
		}
		err = sampler.LogEvent("conn-1", event)
		require.NoError(t, err)
	}

	// Stop capture
	err = sampler.StopCapture("conn-1")
	require.NoError(t, err)

	// Writer should be closed
	assert.True(t, writer.closed)

	// Should not be in active writers
	sampler.writersMutex.Lock()
	_, exists := sampler.activeWriters["conn-1"]
	sampler.writersMutex.Unlock()
	assert.False(t, exists)
}

func TestSetSamplingRate(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tmpDir := t.TempDir()

	config := &QLogSamplingConfig{
		SamplingRate:  0.01,
		GlobalEnabled: true,
		OutputDir:     tmpDir,
	}

	sampler, err := NewQLogSampler(logger, config)
	require.NoError(t, err)
	defer sampler.Close()

	// Update sampling rate
	err = sampler.SetSamplingRate(0.05)
	require.NoError(t, err)
	assert.Equal(t, 0.05, sampler.samplingRate)

	// Invalid rates
	err = sampler.SetSamplingRate(-0.1)
	assert.Error(t, err)

	err = sampler.SetSamplingRate(1.1)
	assert.Error(t, err)
}

func TestEnableDisableForTenant(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tmpDir := t.TempDir()

	config := &QLogSamplingConfig{
		SamplingRate:  1.0,
		GlobalEnabled: false,
		OutputDir:     tmpDir,
	}

	sampler, err := NewQLogSampler(logger, config)
	require.NoError(t, err)
	defer sampler.Close()

	// Initially should not sample
	shouldSample := sampler.ShouldSample("conn-1", "tenant-1")
	assert.False(t, shouldSample)

	// Enable for tenant
	sampler.EnableForTenant("tenant-1")
	shouldSample = sampler.ShouldSample("conn-1", "tenant-1")
	assert.True(t, shouldSample)

	// Disable for tenant
	sampler.DisableForTenant("tenant-1")
	shouldSample = sampler.ShouldSample("conn-1", "tenant-1")
	assert.False(t, shouldSample)
}

func TestSetGlobalEnabled(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tmpDir := t.TempDir()

	config := &QLogSamplingConfig{
		SamplingRate:  1.0,
		GlobalEnabled: false,
		OutputDir:     tmpDir,
	}

	sampler, err := NewQLogSampler(logger, config)
	require.NoError(t, err)
	defer sampler.Close()

	// Initially disabled
	shouldSample := sampler.ShouldSample("conn-1", "tenant-1")
	assert.False(t, shouldSample)

	// Enable globally
	sampler.SetGlobalEnabled(true)
	shouldSample = sampler.ShouldSample("conn-1", "tenant-1")
	assert.True(t, shouldSample)

	// Disable globally
	sampler.SetGlobalEnabled(false)
	shouldSample = sampler.ShouldSample("conn-1", "tenant-1")
	assert.False(t, shouldSample)
}

func TestQLogSampler_GetStats(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tmpDir := t.TempDir()

	config := &QLogSamplingConfig{
		SamplingRate:  0.5,
		GlobalEnabled: true,
		OutputDir:     tmpDir,
	}

	sampler, err := NewQLogSampler(logger, config)
	require.NoError(t, err)
	defer sampler.Close()

	// Start some captures
	for i := 0; i < 100; i++ {
		connectionID := fmt.Sprintf("conn-%d", i)
		sampler.StartCapture(connectionID, "tenant-1")
	}

	stats := sampler.GetStats()

	assert.Equal(t, int64(100), stats["total_connections"])
	assert.Greater(t, stats["sampled_connections"].(int64), int64(0))
	assert.Greater(t, stats["dropped_connections"].(int64), int64(0))
	assert.Equal(t, 50.0, stats["sampling_rate"])
	assert.True(t, stats["global_enabled"].(bool))
}

func TestCleanupOldLogs(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tmpDir := t.TempDir()

	config := &QLogSamplingConfig{
		SamplingRate:  1.0,
		GlobalEnabled: true,
		OutputDir:     tmpDir,
		RetentionDays: 1, // 1 day retention for testing
	}

	sampler, err := NewQLogSampler(logger, config)
	require.NoError(t, err)
	defer sampler.Close()

	// Create old qlog file
	oldFile := filepath.Join(tmpDir, "qlog-tenant-conn-20200101-120000.json")
	err = os.WriteFile(oldFile, []byte("old log"), 0644)
	require.NoError(t, err)

	// Set modification time to 2 days ago
	twoDaysAgo := time.Now().AddDate(0, 0, -2)
	err = os.Chtimes(oldFile, twoDaysAgo, twoDaysAgo)
	require.NoError(t, err)

	// Create recent file
	recentFile := filepath.Join(tmpDir, "qlog-tenant-conn-20250101-120000.json")
	err = os.WriteFile(recentFile, []byte("recent log"), 0644)
	require.NoError(t, err)

	// Run cleanup
	sampler.performCleanup()

	// Old file should be deleted
	_, err = os.Stat(oldFile)
	assert.True(t, os.IsNotExist(err))

	// Recent file should still exist
	_, err = os.Stat(recentFile)
	assert.NoError(t, err)
}

func TestClose(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tmpDir := t.TempDir()

	config := &QLogSamplingConfig{
		SamplingRate:  1.0,
		GlobalEnabled: true,
		OutputDir:     tmpDir,
	}

	sampler, err := NewQLogSampler(logger, config)
	require.NoError(t, err)

	// Start multiple captures
	for i := 0; i < 5; i++ {
		connectionID := fmt.Sprintf("conn-%d", i)
		writer, err := sampler.StartCapture(connectionID, "tenant-1")
		require.NoError(t, err)
		require.NotNil(t, writer)
	}

	// Close sampler
	err = sampler.Close()
	require.NoError(t, err)

	// All writers should be closed
	sampler.writersMutex.Lock()
	activeCount := len(sampler.activeWriters)
	sampler.writersMutex.Unlock()
	assert.Equal(t, 0, activeCount)
}

func BenchmarkShouldSample(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	tmpDir := b.TempDir()

	config := &QLogSamplingConfig{
		SamplingRate:  0.01,
		GlobalEnabled: true,
		OutputDir:     tmpDir,
	}

	sampler, _ := NewQLogSampler(logger, config)
	defer sampler.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		connectionID := fmt.Sprintf("conn-%d", i)
		sampler.ShouldSample(connectionID, "tenant-1")
	}
}

func BenchmarkRedactPII(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	tmpDir := b.TempDir()

	config := &QLogSamplingConfig{
		SamplingRate:  1.0,
		GlobalEnabled: true,
		OutputDir:     tmpDir,
		RedactIPv4:    true,
		RedactIPv6:    true,
		RedactTokens:  true,
	}

	sampler, _ := NewQLogSampler(logger, config)
	defer sampler.Close()

	data := map[string]interface{}{
		"src_ip":        "192.168.1.100",
		"dst_ip":        "10.0.0.1",
		"src_port":      12345,
		"dst_port":      443,
		"connection_id": "conn-123-456-789",
		"token":         "secret-token",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sampler.redactPII(data)
	}
}
