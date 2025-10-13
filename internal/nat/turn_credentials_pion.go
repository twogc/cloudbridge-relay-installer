package nat

import (
	"fmt"
	"time"

	pionTurn "github.com/pion/turn/v3"
	"go.uber.org/zap"
)

// TURNCredentialService wraps pion/turn standard implementation with CloudBridge business logic
// This replaces the custom turn_credentials.go with RFC 5766 compliant pion/turn library
type TURNCredentialService struct {
	logger       *zap.Logger
	sharedSecret string
	realm        string
	defaultTTL   time.Duration
	minTTL       time.Duration
	maxTTL       time.Duration

	// Business logic: quota management (not in pion/turn)
	quotaManager *TURNQuotaManager
}

// NewTURNCredentialService creates a service using pion/turn standards
func NewTURNCredentialService(sharedSecret string, realm string, logger *zap.Logger) *TURNCredentialService {
	return &TURNCredentialService{
		logger:       logger,
		sharedSecret: sharedSecret,
		realm:        realm,
		defaultTTL:   10 * time.Minute,
		minTTL:       1 * time.Minute,
		maxTTL:       60 * time.Minute,
		quotaManager: NewTURNQuotaManager(logger),
	}
}

// GenerateCredential generates time-limited TURN credentials using pion/turn standards
// Implements draft-uberti-behave-turn-rest-00 (TURN REST API)
func (tcs *TURNCredentialService) GenerateCredential(
	tenantID string,
	peerID string,
	requestedTTL time.Duration,
) (*TURNCredentialResponse, error) {

	// Validate and normalize TTL
	ttl := tcs.normalizeTTL(requestedTTL)

	// Check business logic: quota
	if err := tcs.quotaManager.CheckQuota(tenantID, peerID); err != nil {
		tcs.logger.Warn("TURN quota exceeded",
			zap.String("tenant_id", tenantID),
			zap.String("peer_id", peerID),
			zap.Error(err))
		return nil, err
	}

	// Use pion/turn standard implementation (RFC 5766 compliant)
	// Format: <timestamp>:<user>
	// Password: HMAC-SHA1(shared_secret, username)
	username, password, err := pionTurn.GenerateLongTermTURNRESTCredentials(
		tcs.sharedSecret,
		peerID,
		ttl,
	)
	if err != nil {
		return nil, fmt.Errorf("pion/turn credential generation failed: %w", err)
	}

	// Increment business logic quota
	tcs.quotaManager.IncrementAllocation(tenantID, peerID)

	// Get tenant quota info for response
	quota := tcs.quotaManager.GetTenantQuota(tenantID)
	expiresAt := time.Now().Add(ttl)

	tcs.logger.Info("Generated TURN credential using pion/turn standards",
		zap.String("tenant_id", tenantID),
		zap.String("peer_id", peerID),
		zap.Duration("ttl", ttl),
		zap.Time("expires_at", expiresAt),
		zap.Int("quota_used", quota.ActiveAllocations),
		zap.String("standard", "RFC 5766 + draft-uberti-behave-turn-rest-00"))

	return &TURNCredentialResponse{
		Username:   username,
		Password:   password,
		TTL:        int(ttl.Seconds()),
		ExpiresAt:  expiresAt,
		URIs:       tcs.getTURNURIs(),
		Realm:      tcs.realm,
		QuotaUsed:  quota.ActiveAllocations,
		QuotaLimit: quota.QuotaLimit,
	}, nil
}

// normalizeTTL validates and normalizes TTL to acceptable range
func (tcs *TURNCredentialService) normalizeTTL(requestedTTL time.Duration) time.Duration {
	if requestedTTL <= 0 {
		return tcs.defaultTTL
	}
	if requestedTTL < tcs.minTTL {
		return tcs.minTTL
	}
	if requestedTTL > tcs.maxTTL {
		return tcs.maxTTL
	}
	return requestedTTL
}

// getTURNURIs returns configured TURN server URIs
// In production, load from config or service discovery
func (tcs *TURNCredentialService) getTURNURIs() []string {
	// TODO: Load from environment/config
	return []string{
		"turn:turn.cloudbridge.2gc.ru:3478?transport=udp",
		"turn:turn.cloudbridge.2gc.ru:3478?transport=tcp",
		"turns:turn.cloudbridge.2gc.ru:5349?transport=tcp",
	}
}

// GetQuotaManager returns the quota manager for advanced operations
func (tcs *TURNCredentialService) GetQuotaManager() *TURNQuotaManager {
	return tcs.quotaManager
}

// SetTTLLimits allows customizing TTL constraints
func (tcs *TURNCredentialService) SetTTLLimits(defaultTTL, minTTL, maxTTL time.Duration) {
	tcs.defaultTTL = defaultTTL
	tcs.minTTL = minTTL
	tcs.maxTTL = maxTTL

	tcs.logger.Info("Updated TURN TTL limits",
		zap.Duration("default", defaultTTL),
		zap.Duration("min", minTTL),
		zap.Duration("max", maxTTL))
}

// HealthCheck returns health status of TURN credential service
func (tcs *TURNCredentialService) HealthCheck() map[string]interface{} {
	quotaStats := tcs.quotaManager.GetQuotaStats()

	status := "healthy"
	issues := make([]string, 0)

	if tcs.sharedSecret == "" {
		status = "unhealthy"
		issues = append(issues, "missing shared secret")
	}

	if quotaStats["active_allocations"].(int) > 10000 {
		status = "degraded"
		issues = append(issues, "high allocation count")
	}

	return map[string]interface{}{
		"component":         "turn_credential_service",
		"status":            status,
		"implementation":    "pion/turn v3 (RFC 5766 compliant)",
		"standard":          "draft-uberti-behave-turn-rest-00",
		"shared_secret_set": tcs.sharedSecret != "",
		"realm":             tcs.realm,
		"default_ttl_sec":   int(tcs.defaultTTL.Seconds()),
		"min_ttl_sec":       int(tcs.minTTL.Seconds()),
		"max_ttl_sec":       int(tcs.maxTTL.Seconds()),
		"quota_stats":       quotaStats,
		"issues":            issues,
	}
}

// CreatePionTURNAuthHandler creates a pion/turn AuthHandler for server-side validation
// This should be used by the TURN server to validate incoming credentials
func CreatePionTURNAuthHandler(sharedSecret string, logger *zap.Logger) pionTurn.AuthHandler {
	// Create adapter from zap.Logger to pion logging.LeveledLogger
	pionLogger := NewPionLoggerAdapter(logger)

	// Use pion/turn standard validation handler
	// Automatically validates:
	// - Username format: <timestamp>:<user>
	// - Expiry: timestamp > now
	// - Password: HMAC-SHA1(shared_secret, username)
	return pionTurn.LongTermTURNRESTAuthHandler(sharedSecret, pionLogger)
}

// PionLoggerAdapter adapts zap.Logger to pion logging.LeveledLogger interface
type PionLoggerAdapter struct {
	zapLogger *zap.Logger
}

// NewPionLoggerAdapter creates a new adapter
func NewPionLoggerAdapter(zapLogger *zap.Logger) *PionLoggerAdapter {
	return &PionLoggerAdapter{zapLogger: zapLogger}
}

// Trace logs at trace level
func (p *PionLoggerAdapter) Trace(msg string) {
	p.zapLogger.Debug(msg, zap.String("level", "trace"))
}

// Tracef logs at trace level with formatting
func (p *PionLoggerAdapter) Tracef(format string, args ...interface{}) {
	p.zapLogger.Debug(fmt.Sprintf(format, args...), zap.String("level", "trace"))
}

// Debug logs at debug level
func (p *PionLoggerAdapter) Debug(msg string) {
	p.zapLogger.Debug(msg)
}

// Debugf logs at debug level with formatting
func (p *PionLoggerAdapter) Debugf(format string, args ...interface{}) {
	p.zapLogger.Debug(fmt.Sprintf(format, args...))
}

// Info logs at info level
func (p *PionLoggerAdapter) Info(msg string) {
	p.zapLogger.Info(msg)
}

// Infof logs at info level with formatting
func (p *PionLoggerAdapter) Infof(format string, args ...interface{}) {
	p.zapLogger.Info(fmt.Sprintf(format, args...))
}

// Warn logs at warn level
func (p *PionLoggerAdapter) Warn(msg string) {
	p.zapLogger.Warn(msg)
}

// Warnf logs at warn level with formatting
func (p *PionLoggerAdapter) Warnf(format string, args ...interface{}) {
	p.zapLogger.Warn(fmt.Sprintf(format, args...))
}

// Error logs at error level
func (p *PionLoggerAdapter) Error(msg string) {
	p.zapLogger.Error(msg)
}

// Errorf logs at error level with formatting
func (p *PionLoggerAdapter) Errorf(format string, args ...interface{}) {
	p.zapLogger.Error(fmt.Sprintf(format, args...))
}
