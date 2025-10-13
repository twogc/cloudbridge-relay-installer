package quic

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"go.uber.org/zap"
)

// QLogSampler manages qlog sampling and PII redaction
// Addresses security concern: Full qlog dumps contain PII and huge storage
type QLogSampler struct {
	logger *zap.Logger
	mu     sync.RWMutex

	// Sampling configuration
	samplingRate      float64 // 0.0 to 1.0 (default: 0.01 = 1%)
	enabledForTenants map[string]bool
	globalEnabled     bool

	// PII redaction
	redactIPv4     bool
	redactIPv6     bool
	redactPorts    bool
	redactTokens   bool
	piiHashSalt    string

	// Storage configuration
	outputDir      string
	maxFileSize    int64         // bytes
	retentionDays  int
	flushInterval  time.Duration

	// Statistics
	totalConnections  int64
	sampledConnections int64
	droppedConnections int64
	bytesWritten      int64

	// Active qlog writers
	activeWriters map[string]*QLogWriter
	writersMutex  sync.Mutex
}

// QLogWriter handles writing qlog events for a single connection
type QLogWriter struct {
	connectionID string
	tenantID     string
	file         *os.File
	events       []QLogSamplingEvent
	mutex        sync.Mutex
	closed       bool
}

// QLogSamplingEvent represents a single qlog event for sampling
type QLogSamplingEvent struct {
	Time      int64                  `json:"time"`
	Name      string                 `json:"name"`
	Category  string                 `json:"category,omitempty"`
	Type      string                 `json:"type,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// QLogSamplingConfig configuration for qlog sampling
type QLogSamplingConfig struct {
	SamplingRate      float64           // Default: 0.01 (1%)
	EnabledForTenants map[string]bool   // Per-tenant overrides
	GlobalEnabled     bool              // Global on/off

	// PII Redaction
	RedactIPv4    bool   // Default: true
	RedactIPv6    bool   // Default: true
	RedactPorts   bool   // Default: false
	RedactTokens  bool   // Default: true
	PIIHashSalt   string // Salt for hashing PII

	// Storage
	OutputDir     string        // Default: /var/log/cloudbridge/qlog
	MaxFileSize   int64         // Default: 10MB
	RetentionDays int           // Default: 7 days
	FlushInterval time.Duration // Default: 30s
}

// NewQLogSampler creates a new qlog sampler with safe defaults
func NewQLogSampler(logger *zap.Logger, config *QLogSamplingConfig) (*QLogSampler, error) {
	if config == nil {
		config = &QLogSamplingConfig{
			SamplingRate:      0.01, // 1% sampling by default
			GlobalEnabled:     false, // OFF by default (opt-in)
			RedactIPv4:        true,
			RedactIPv6:        true,
			RedactPorts:       false,
			RedactTokens:      true,
			PIIHashSalt:       generateRandomSalt(),
			OutputDir:         "/var/log/cloudbridge/qlog",
			MaxFileSize:       10 * 1024 * 1024, // 10MB
			RetentionDays:     7,
			FlushInterval:     30 * time.Second,
			EnabledForTenants: make(map[string]bool),
		}
	}

	// Create output directory
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create qlog output directory: %w", err)
	}

	sampler := &QLogSampler{
		logger:            logger,
		samplingRate:      config.SamplingRate,
		enabledForTenants: config.EnabledForTenants,
		globalEnabled:     config.GlobalEnabled,
		redactIPv4:        config.RedactIPv4,
		redactIPv6:        config.RedactIPv6,
		redactPorts:       config.RedactPorts,
		redactTokens:      config.RedactTokens,
		piiHashSalt:       config.PIIHashSalt,
		outputDir:         config.OutputDir,
		maxFileSize:       config.MaxFileSize,
		retentionDays:     config.RetentionDays,
		flushInterval:     config.FlushInterval,
		activeWriters:     make(map[string]*QLogWriter),
	}

	// Start background cleanup goroutine
	go sampler.cleanupOldLogs()

	logger.Info("QLog sampler initialized",
		zap.Float64("sampling_rate", config.SamplingRate*100),
		zap.Bool("global_enabled", config.GlobalEnabled),
		zap.Bool("redact_ipv4", config.RedactIPv4),
		zap.Bool("redact_ipv6", config.RedactIPv6),
		zap.String("output_dir", config.OutputDir))

	return sampler, nil
}

// ShouldSample determines if a connection should be sampled
func (qs *QLogSampler) ShouldSample(connectionID, tenantID string) bool {
	qs.mu.RLock()
	defer qs.mu.RUnlock()

	// Check if globally disabled
	if !qs.globalEnabled {
		// Check tenant-specific override
		if !qs.enabledForTenants[tenantID] {
			return false
		}
	}

	// Deterministic sampling based on connection ID hash
	// This ensures same connection always gets same sampling decision
	hash := sha256.Sum256([]byte(connectionID + qs.piiHashSalt))
	hashValue := float64(hash[0]) / 255.0

	sampled := hashValue < qs.samplingRate

	if sampled {
		qs.logger.Debug("Connection sampled for qlog",
			zap.String("connection_id", connectionID),
			zap.String("tenant_id", tenantID))
	}

	return sampled
}

// StartCapture starts capturing qlog for a connection
func (qs *QLogSampler) StartCapture(connectionID, tenantID string) (*QLogWriter, error) {
	qs.totalConnections++

	if !qs.ShouldSample(connectionID, tenantID) {
		qs.droppedConnections++
		return nil, nil // Not sampled, not an error
	}

	qs.sampledConnections++

	// Create qlog file
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("qlog-%s-%s-%s.json", tenantID, connectionID[:8], timestamp)
	filepath := filepath.Join(qs.outputDir, filename)

	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to create qlog file: %w", err)
	}

	writer := &QLogWriter{
		connectionID: connectionID,
		tenantID:     tenantID,
		file:         file,
		events:       make([]QLogSamplingEvent, 0, 1000),
	}

	// Write qlog header
	header := map[string]interface{}{
		"qlog_version": "0.3",
		"title":        "CloudBridge QUIC Connection Log",
		"trace": map[string]interface{}{
			"common_fields": map[string]interface{}{
				"reference_time": time.Now().Unix(),
				"time_format":    "relative",
			},
		},
	}
	headerJSON, _ := json.Marshal(header)
	file.Write(headerJSON)
	file.Write([]byte("\n"))

	qs.writersMutex.Lock()
	qs.activeWriters[connectionID] = writer
	qs.writersMutex.Unlock()

	qs.logger.Info("Started qlog capture",
		zap.String("connection_id", connectionID),
		zap.String("tenant_id", tenantID),
		zap.String("file", filepath))

	return writer, nil
}

// LogEvent logs a qlog event with PII redaction
func (qs *QLogSampler) LogEvent(connectionID string, event QLogEvent) error {
	qs.writersMutex.Lock()
	writer, exists := qs.activeWriters[connectionID]
	qs.writersMutex.Unlock()

	if !exists {
		// Connection not being sampled
		return nil
	}

	// Redact PII from event data
	event.Data = qs.redactPII(event.Data)

	writer.mutex.Lock()
	defer writer.mutex.Unlock()

	if writer.closed {
		return fmt.Errorf("writer already closed")
	}

	// Convert QLogEvent to QLogSamplingEvent
	samplingEvent := QLogSamplingEvent{
		Time:     event.Time.UnixNano() / 1e6, // Convert to milliseconds
		Name:     event.Event,
		Category: event.Category,
		// Type field not available in QLogEvent, leave empty
		Data:     event.Data,
	}

	// Add event to buffer
	writer.events = append(writer.events, samplingEvent)

	// Flush if buffer is large
	if len(writer.events) >= 100 {
		return qs.flushWriter(writer)
	}

	return nil
}

// redactPII redacts personally identifiable information from event data
func (qs *QLogSampler) redactPII(data map[string]interface{}) map[string]interface{} {
	if data == nil {
		return data
	}

	redacted := make(map[string]interface{}, len(data))

	for key, value := range data {
		switch key {
		case "ip", "src_ip", "dst_ip", "remote_ip", "local_ip":
			if qs.redactIPv4 || qs.redactIPv6 {
				redacted[key] = qs.hashValue(fmt.Sprintf("%v", value))
			} else {
				redacted[key] = value
			}

		case "port", "src_port", "dst_port", "remote_port", "local_port":
			if qs.redactPorts {
				redacted[key] = "REDACTED"
			} else {
				redacted[key] = value
			}

		case "token", "session_ticket", "resume_token", "retry_token":
			if qs.redactTokens {
				redacted[key] = qs.hashValue(fmt.Sprintf("%v", value))
			} else {
				redacted[key] = value
			}

		case "connection_id", "dcid", "scid":
			// Always hash connection IDs for privacy
			redacted[key] = qs.hashValue(fmt.Sprintf("%v", value))

		default:
			// Check if value is a nested map
			if nestedMap, ok := value.(map[string]interface{}); ok {
				redacted[key] = qs.redactPII(nestedMap)
			} else {
				redacted[key] = value
			}
		}
	}

	return redacted
}

// hashValue hashes a value with salt for PII redaction
func (qs *QLogSampler) hashValue(value string) string {
	hash := sha256.Sum256([]byte(value + qs.piiHashSalt))
	return "hash:" + hex.EncodeToString(hash[:])[:16]
}

// flushWriter flushes buffered events to file
func (qs *QLogSampler) flushWriter(writer *QLogWriter) error {
	if len(writer.events) == 0 {
		return nil
	}

	// Write events as JSON array
	for _, event := range writer.events {
		eventJSON, err := json.Marshal(event)
		if err != nil {
			return err
		}
		writer.file.Write(eventJSON)
		writer.file.Write([]byte("\n"))
		qs.bytesWritten += int64(len(eventJSON) + 1)
	}

	// Clear buffer
	writer.events = writer.events[:0]

	// Sync to disk
	return writer.file.Sync()
}

// StopCapture stops capturing qlog for a connection
func (qs *QLogSampler) StopCapture(connectionID string) error {
	qs.writersMutex.Lock()
	writer, exists := qs.activeWriters[connectionID]
	if exists {
		delete(qs.activeWriters, connectionID)
	}
	qs.writersMutex.Unlock()

	if !exists {
		return nil // Not being sampled
	}

	writer.mutex.Lock()
	defer writer.mutex.Unlock()

	if writer.closed {
		return nil
	}

	// Flush remaining events
	if err := qs.flushWriter(writer); err != nil {
		qs.logger.Error("Failed to flush qlog events", zap.Error(err))
	}

	// Close file
	writer.closed = true
	if err := writer.file.Close(); err != nil {
		return err
	}

	qs.logger.Info("Stopped qlog capture",
		zap.String("connection_id", connectionID),
		zap.String("tenant_id", writer.tenantID))

	return nil
}

// SetSamplingRate updates the sampling rate
func (qs *QLogSampler) SetSamplingRate(rate float64) error {
	if rate < 0 || rate > 1 {
		return fmt.Errorf("sampling rate must be between 0 and 1")
	}

	qs.mu.Lock()
	defer qs.mu.Unlock()

	qs.samplingRate = rate
	qs.logger.Info("Updated qlog sampling rate", zap.Float64("rate", rate*100))

	return nil
}

// EnableForTenant enables qlog sampling for specific tenant
func (qs *QLogSampler) EnableForTenant(tenantID string) {
	qs.mu.Lock()
	defer qs.mu.Unlock()

	qs.enabledForTenants[tenantID] = true
	qs.logger.Info("Enabled qlog for tenant", zap.String("tenant_id", tenantID))
}

// DisableForTenant disables qlog sampling for specific tenant
func (qs *QLogSampler) DisableForTenant(tenantID string) {
	qs.mu.Lock()
	defer qs.mu.Unlock()

	delete(qs.enabledForTenants, tenantID)
	qs.logger.Info("Disabled qlog for tenant", zap.String("tenant_id", tenantID))
}

// SetGlobalEnabled enables/disables qlog globally
func (qs *QLogSampler) SetGlobalEnabled(enabled bool) {
	qs.mu.Lock()
	defer qs.mu.Unlock()

	qs.globalEnabled = enabled
	qs.logger.Info("Updated global qlog status", zap.Bool("enabled", enabled))
}

// GetStats returns qlog sampler statistics
func (qs *QLogSampler) GetStats() map[string]interface{} {
	qs.mu.RLock()
	defer qs.mu.RUnlock()

	samplingPercent := 0.0
	if qs.totalConnections > 0 {
		samplingPercent = float64(qs.sampledConnections) / float64(qs.totalConnections) * 100
	}

	return map[string]interface{}{
		"total_connections":   qs.totalConnections,
		"sampled_connections": qs.sampledConnections,
		"dropped_connections": qs.droppedConnections,
		"sampling_rate":       qs.samplingRate * 100,
		"actual_sample_rate":  samplingPercent,
		"bytes_written":       qs.bytesWritten,
		"active_captures":     len(qs.activeWriters),
		"global_enabled":      qs.globalEnabled,
		"tenant_overrides":    len(qs.enabledForTenants),
	}
}

// cleanupOldLogs removes qlog files older than retention period
func (qs *QLogSampler) cleanupOldLogs() {
	ticker := time.NewTicker(24 * time.Hour) // Run daily
	defer ticker.Stop()

	for range ticker.C {
		qs.performCleanup()
	}
}

// performCleanup performs the actual cleanup
func (qs *QLogSampler) performCleanup() {
	cutoffTime := time.Now().AddDate(0, 0, -qs.retentionDays)

	files, err := os.ReadDir(qs.outputDir)
	if err != nil {
		qs.logger.Error("Failed to read qlog directory", zap.Error(err))
		return
	}

	deletedCount := 0
	var deletedBytes int64

	qlogPattern := regexp.MustCompile(`^qlog-.*\.json$`)

	for _, file := range files {
		if !qlogPattern.MatchString(file.Name()) {
			continue
		}

		info, err := file.Info()
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoffTime) {
			filepath := filepath.Join(qs.outputDir, file.Name())
			if err := os.Remove(filepath); err != nil {
				qs.logger.Error("Failed to delete old qlog file",
					zap.String("file", filepath),
					zap.Error(err))
				continue
			}

			deletedCount++
			deletedBytes += info.Size()
		}
	}

	if deletedCount > 0 {
		qs.logger.Info("Cleaned up old qlog files",
			zap.Int("deleted_files", deletedCount),
			zap.Int64("deleted_bytes", deletedBytes),
			zap.Int("retention_days", qs.retentionDays))
	}
}

// Close closes the qlog sampler
func (qs *QLogSampler) Close() error {
	qs.logger.Info("Closing qlog sampler")

	// Close all active writers
	qs.writersMutex.Lock()
	writers := make([]*QLogWriter, 0, len(qs.activeWriters))
	for _, writer := range qs.activeWriters {
		writers = append(writers, writer)
	}
	qs.activeWriters = make(map[string]*QLogWriter)
	qs.writersMutex.Unlock()

	for _, writer := range writers {
		if err := qs.StopCapture(writer.connectionID); err != nil {
			qs.logger.Error("Failed to close qlog writer",
				zap.String("connection_id", writer.connectionID),
				zap.Error(err))
		}
	}

	return nil
}

// generateRandomSalt generates a random salt for PII hashing
func generateRandomSalt() string {
	hash := sha256.Sum256([]byte(time.Now().String()))
	return hex.EncodeToString(hash[:])
}

// Errors
var (
	ErrInvalidSamplingRate = fmt.Errorf("sampling rate must be between 0 and 1")
	ErrWriterClosed        = fmt.Errorf("writer already closed")
)
