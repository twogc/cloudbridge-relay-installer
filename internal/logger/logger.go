package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// LogLevel represents the logging level
type LogLevel string

const (
	DEBUG   LogLevel = "DEBUG"
	INFO    LogLevel = "INFO"
	WARNING LogLevel = "WARNING"
	ERROR   LogLevel = "ERROR"
)

// Config represents logger configuration
type Config struct {
	Level      LogLevel `json:"level" yaml:"level"`
	Output     string   `json:"output" yaml:"output"` // "stdout", "file", "both"
	FilePath   string   `json:"file_path" yaml:"file_path"`
	MaxSize    int      `json:"max_size" yaml:"max_size"`       // MB
	MaxBackups int      `json:"max_backups" yaml:"max_backups"` // number of backup files
	MaxAge     int      `json:"max_age" yaml:"max_age"`         // days
	Compress   bool     `json:"compress" yaml:"compress"`
	JSON       bool     `json:"json" yaml:"json"` // structured logging
}

// Logger represents the application logger
type Logger struct {
	zapLogger *zap.Logger
	config    Config
}

// New creates a new logger instance
func New(config Config) (*Logger, error) {
	var cores []zapcore.Core

	// Create encoder
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	var encoder zapcore.Encoder
	if config.JSON {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// Parse log level
	level, err := zapcore.ParseLevel(strings.ToLower(string(config.Level)))
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %s", config.Level)
	}

	// Console output
	if config.Output == "stdout" || config.Output == "both" {
		consoleCore := zapcore.NewCore(
			encoder,
			zapcore.AddSync(os.Stdout),
			level,
		)
		cores = append(cores, consoleCore)
	}

	// File output
	if config.Output == "file" || config.Output == "both" {
		if config.FilePath == "" {
			config.FilePath = "logs/relay.log"
		}

		// Ensure directory exists with secure permissions
		dir := filepath.Dir(config.FilePath)
		if err := os.MkdirAll(dir, 0750); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}

		// Configure log rotation
		writer := &lumberjack.Logger{
			Filename:   config.FilePath,
			MaxSize:    config.MaxSize,    // MB
			MaxBackups: config.MaxBackups, // number of backup files
			MaxAge:     config.MaxAge,     // days
			Compress:   config.Compress,   // compress rotated files
		}

		fileCore := zapcore.NewCore(
			encoder,
			zapcore.AddSync(writer),
			level,
		)
		cores = append(cores, fileCore)
	}

	if len(cores) == 0 {
		return nil, fmt.Errorf("no output configured")
	}

	// Create core
	core := zapcore.NewTee(cores...)

	// Create logger
	zapLogger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return &Logger{
		zapLogger: zapLogger,
		config:    config,
	}, nil
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.zapLogger.Debug(msg, fields...)
}

// Info logs an info message
func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.zapLogger.Info(msg, fields...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.zapLogger.Warn(msg, fields...)
}

// Error logs an error message
func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.zapLogger.Error(msg, fields...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.zapLogger.Fatal(msg, fields...)
}

// With creates a child logger with additional fields
func (l *Logger) With(fields ...zap.Field) *Logger {
	return &Logger{
		zapLogger: l.zapLogger.With(fields...),
		config:    l.config,
	}
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	return l.zapLogger.Sync()
}

// Helper methods for common logging patterns

// WithRequestID creates a logger with request ID
func (l *Logger) WithRequestID(requestID string) *Logger {
	return l.With(zap.String("request_id", requestID))
}

// WithUserID creates a logger with user ID
func (l *Logger) WithUserID(userID string) *Logger {
	return l.With(zap.String("user_id", userID))
}

// WithConnectionID creates a logger with connection ID
func (l *Logger) WithConnectionID(connID string) *Logger {
	return l.With(zap.String("connection_id", connID))
}

// WithDuration logs a message with duration
func (l *Logger) WithDuration(duration time.Duration) *Logger {
	return l.With(zap.Duration("duration", duration))
}

// WithError logs a message with error
func (l *Logger) WithError(err error) *Logger {
	return l.With(zap.Error(err))
}

// WithMetrics logs a message with metrics
func (l *Logger) WithMetrics(metrics map[string]interface{}) *Logger {
	fields := make([]zap.Field, 0, len(metrics))
	for k, v := range metrics {
		fields = append(fields, zap.Any(k, v))
	}
	return l.With(fields...)
}

// LogConnection logs connection events
func (l *Logger) LogConnection(connID, userID, remoteAddr string, success bool) {
	fields := []zap.Field{
		zap.String("connection_id", connID),
		zap.String("user_id", userID),
		zap.String("remote_addr", remoteAddr),
		zap.Bool("success", success),
	}

	if success {
		l.Info("Connection established", fields...)
	} else {
		l.Error("Connection failed", fields...)
	}
}

// LogAuthentication logs authentication events
func (l *Logger) LogAuthentication(userID, method string, success bool, duration time.Duration) {
	fields := []zap.Field{
		zap.String("user_id", userID),
		zap.String("method", method),
		zap.Bool("success", success),
		zap.Duration("duration", duration),
	}

	if success {
		l.Info("Authentication successful", fields...)
	} else {
		l.Error("Authentication failed", fields...)
	}
}

// LogTokenValidation logs token validation events
func (l *Logger) LogTokenValidation(tokenID, method string, success bool, duration time.Duration, claims map[string]interface{}) {
	fields := []zap.Field{
		zap.String("token_id", tokenID),
		zap.String("method", method),
		zap.Bool("success", success),
		zap.Duration("duration", duration),
	}

	if len(claims) > 0 {
		fields = append(fields, zap.Any("claims", claims))
	}

	if success {
		l.Info("Token validation successful", fields...)
	} else {
		l.Error("Token validation failed", fields...)
	}
}

// LogPermissionCheck logs permission check events
func (l *Logger) LogPermissionCheck(userID, resource, action string, granted bool, duration time.Duration) {
	fields := []zap.Field{
		zap.String("user_id", userID),
		zap.String("resource", resource),
		zap.String("action", action),
		zap.Bool("granted", granted),
		zap.Duration("duration", duration),
	}

	if granted {
		l.Info("Permission granted", fields...)
	} else {
		l.Warn("Permission denied", fields...)
	}
}

// LogPerformance logs performance metrics
func (l *Logger) LogPerformance(operation string, duration time.Duration, metadata map[string]interface{}) {
	fields := []zap.Field{
		zap.String("operation", operation),
		zap.Duration("duration", duration),
	}

	if len(metadata) > 0 {
		fields = append(fields, zap.Any("metadata", metadata))
	}

	if duration > time.Second {
		l.Warn("Slow operation detected", fields...)
	} else {
		l.Debug("Operation completed", fields...)
	}
}

// LogSecurity logs security events
func (l *Logger) LogSecurity(event string, userID, remoteAddr string, details map[string]interface{}) {
	fields := []zap.Field{
		zap.String("event", event),
		zap.String("user_id", userID),
		zap.String("remote_addr", remoteAddr),
	}

	if len(details) > 0 {
		fields = append(fields, zap.Any("details", details))
	}

	l.Warn("Security event detected", fields...)
}

// LogBusinessEvent logs business events
func (l *Logger) LogBusinessEvent(event string, userID string, data map[string]interface{}) {
	fields := []zap.Field{
		zap.String("event", event),
		zap.String("user_id", userID),
	}

	if len(data) > 0 {
		fields = append(fields, zap.Any("data", data))
	}

	l.Info("Business event", fields...)
}

// GetZapLogger returns the underlying zap logger
func (l *Logger) GetZapLogger() *zap.Logger {
	return l.zapLogger
}

// Default logger instance
var defaultLogger *Logger

// InitDefault initializes the default logger
func InitDefault(config Config) error {
	logger, err := New(config)
	if err != nil {
		return err
	}
	defaultLogger = logger
	return nil
}

// GetDefault returns the default logger
func GetDefault() *Logger {
	if defaultLogger == nil {
		// Fallback to basic logger
		config := Config{
			Level:  INFO,
			Output: "stdout",
			JSON:   false,
		}
		logger, _ := New(config)
		defaultLogger = logger
	}
	return defaultLogger
}

// Convenience functions for default logger
func Debug(msg string, fields ...zap.Field) {
	GetDefault().Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	GetDefault().Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	GetDefault().Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	GetDefault().Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	GetDefault().Fatal(msg, fields...)
}

// AuditLogger represents the audit logger for security events
type AuditLogger struct {
	zapLogger *zap.Logger
	config    Config
}

// NewAuditLogger creates a new audit logger instance
func NewAuditLogger(config Config) (*AuditLogger, error) {
	// Override config for audit logging
	auditConfig := config
	auditConfig.JSON = true // Always use JSON for audit logs
	auditConfig.Level = INFO // Audit logs should be at least INFO level
	
	// Set default audit file path if not specified
	if auditConfig.FilePath == "" {
		auditConfig.FilePath = "logs/audit.log"
	}
	
	// Ensure audit logs are kept longer
	if auditConfig.MaxAge == 0 {
		auditConfig.MaxAge = 365 // Keep audit logs for 1 year by default
	}
	if auditConfig.MaxBackups == 0 {
		auditConfig.MaxBackups = 50 // Keep more backup files for audit
	}

	var cores []zapcore.Core

	// Create encoder for audit logs
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	encoder := zapcore.NewJSONEncoder(encoderConfig)

	// Parse log level
	level, err := zapcore.ParseLevel(strings.ToLower(string(auditConfig.Level)))
	if err != nil {
		return nil, fmt.Errorf("invalid audit log level: %s", auditConfig.Level)
	}

	// File output for audit logs
	if auditConfig.FilePath != "" {
		// Ensure directory exists with secure permissions
		dir := filepath.Dir(auditConfig.FilePath)
		if err := os.MkdirAll(dir, 0750); err != nil {
			return nil, fmt.Errorf("failed to create audit log directory: %w", err)
		}

		// Configure log rotation for audit
		writer := &lumberjack.Logger{
			Filename:   auditConfig.FilePath,
			MaxSize:    auditConfig.MaxSize,    // MB
			MaxBackups: auditConfig.MaxBackups, // number of backup files
			MaxAge:     auditConfig.MaxAge,     // days
			Compress:   auditConfig.Compress,   // compress rotated files
		}

		fileCore := zapcore.NewCore(
			encoder,
			zapcore.AddSync(writer),
			level,
		)
		cores = append(cores, fileCore)
	}

	// Also log to stdout for audit events (optional)
	if auditConfig.Output == "stdout" || auditConfig.Output == "both" {
		consoleCore := zapcore.NewCore(
			encoder,
			zapcore.AddSync(os.Stdout),
			level,
		)
		cores = append(cores, consoleCore)
	}

	if len(cores) == 0 {
		return nil, fmt.Errorf("no audit output configured")
	}

	// Create core
	core := zapcore.NewTee(cores...)

	// Create audit logger
	zapLogger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return &AuditLogger{
		zapLogger: zapLogger,
		config:    auditConfig,
	}, nil
}

// AuditEvent logs an audit event
func (al *AuditLogger) AuditEvent(event string, userID, remoteAddr string, details map[string]interface{}) {
	fields := []zap.Field{
		zap.String("event_type", "audit"),
		zap.String("event", event),
		zap.String("user_id", userID),
		zap.String("remote_addr", remoteAddr),
		zap.Time("timestamp", time.Now()),
	}
	
	// Add additional details
	for k, v := range details {
		fields = append(fields, zap.Any(k, v))
	}
	
	al.zapLogger.Info("AUDIT_EVENT", fields...)
}

// AuditAuthentication logs authentication events
func (al *AuditLogger) AuditAuthentication(userID, method, remoteAddr string, success bool, duration time.Duration, details map[string]interface{}) {
	event := "authentication_success"
	if !success {
		event = "authentication_failure"
	}
	
	fields := []zap.Field{
		zap.String("event_type", "audit"),
		zap.String("event", event),
		zap.String("user_id", userID),
		zap.String("method", method),
		zap.String("remote_addr", remoteAddr),
		zap.Bool("success", success),
		zap.Duration("duration", duration),
		zap.Time("timestamp", time.Now()),
	}
	
	// Add additional details
	for k, v := range details {
		fields = append(fields, zap.Any(k, v))
	}
	
	al.zapLogger.Info("AUDIT_AUTHENTICATION", fields...)
}

// AuditAuthorization logs authorization events
func (al *AuditLogger) AuditAuthorization(userID, resource, action, remoteAddr string, granted bool, duration time.Duration, details map[string]interface{}) {
	event := "authorization_granted"
	if !granted {
		event = "authorization_denied"
	}
	
	fields := []zap.Field{
		zap.String("event_type", "audit"),
		zap.String("event", event),
		zap.String("user_id", userID),
		zap.String("resource", resource),
		zap.String("action", action),
		zap.String("remote_addr", remoteAddr),
		zap.Bool("granted", granted),
		zap.Duration("duration", duration),
		zap.Time("timestamp", time.Now()),
	}
	
	// Add additional details
	for k, v := range details {
		fields = append(fields, zap.Any(k, v))
	}
	
	al.zapLogger.Info("AUDIT_AUTHORIZATION", fields...)
}

// AuditTunnelCreation logs tunnel creation events
func (al *AuditLogger) AuditTunnelCreation(userID, tunnelID, remoteAddr string, tunnelInfo map[string]interface{}, details map[string]interface{}) {
	fields := []zap.Field{
		zap.String("event_type", "audit"),
		zap.String("event", "tunnel_created"),
		zap.String("user_id", userID),
		zap.String("tunnel_id", tunnelID),
		zap.String("remote_addr", remoteAddr),
		zap.Time("timestamp", time.Now()),
	}
	
	// Add tunnel info
	for k, v := range tunnelInfo {
		fields = append(fields, zap.Any("tunnel_"+k, v))
	}
	
	// Add additional details
	for k, v := range details {
		fields = append(fields, zap.Any(k, v))
	}
	
	al.zapLogger.Info("AUDIT_TUNNEL_CREATION", fields...)
}

// AuditTunnelDeletion logs tunnel deletion events
func (al *AuditLogger) AuditTunnelDeletion(userID, tunnelID, remoteAddr string, reason string, details map[string]interface{}) {
	fields := []zap.Field{
		zap.String("event_type", "audit"),
		zap.String("event", "tunnel_deleted"),
		zap.String("user_id", userID),
		zap.String("tunnel_id", tunnelID),
		zap.String("remote_addr", remoteAddr),
		zap.String("reason", reason),
		zap.Time("timestamp", time.Now()),
	}
	
	// Add additional details
	for k, v := range details {
		fields = append(fields, zap.Any(k, v))
	}
	
	al.zapLogger.Info("AUDIT_TUNNEL_DELETION", fields...)
}

// AuditRateLimit logs rate limit events
func (al *AuditLogger) AuditRateLimit(userID, remoteAddr, limitType string, details map[string]interface{}) {
	fields := []zap.Field{
		zap.String("event_type", "audit"),
		zap.String("event", "rate_limit_exceeded"),
		zap.String("user_id", userID),
		zap.String("remote_addr", remoteAddr),
		zap.String("limit_type", limitType),
		zap.Time("timestamp", time.Now()),
	}
	
	// Add additional details
	for k, v := range details {
		fields = append(fields, zap.Any(k, v))
	}
	
	al.zapLogger.Warn("AUDIT_RATE_LIMIT", fields...)
}

// AuditSecurity logs security events
func (al *AuditLogger) AuditSecurity(event, userID, remoteAddr string, severity string, details map[string]interface{}) {
	fields := []zap.Field{
		zap.String("event_type", "audit"),
		zap.String("event", event),
		zap.String("user_id", userID),
		zap.String("remote_addr", remoteAddr),
		zap.String("severity", severity),
		zap.Time("timestamp", time.Now()),
	}
	
	// Add additional details
	for k, v := range details {
		fields = append(fields, zap.Any(k, v))
	}
	
	al.zapLogger.Warn("AUDIT_SECURITY", fields...)
}

// Sync flushes any buffered audit log entries
func (al *AuditLogger) Sync() error {
	return al.zapLogger.Sync()
} 