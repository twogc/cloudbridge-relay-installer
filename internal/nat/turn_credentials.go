package nat

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// TURNCredentialManager manages time-limited TURN credentials per RFC 5766 Section 10.2
type TURNCredentialManager struct {
	logger       *zap.Logger
	sharedSecret string // HMAC shared secret
	defaultTTL   time.Duration
	minTTL       time.Duration
	maxTTL       time.Duration

	// Per-tenant quota tracking
	quotaManager *TURNQuotaManager
}

// TURNCredentialResponse represents a time-limited TURN credential
type TURNCredentialResponse struct {
	Username   string    `json:"username"`
	Password   string    `json:"password"`
	TTL        int       `json:"ttl"`        // Seconds
	ExpiresAt  time.Time `json:"expires_at"` // ISO 8601
	URIs       []string  `json:"uris"`
	Realm      string    `json:"realm,omitempty"`
	QuotaUsed  int       `json:"quota_used,omitempty"`
	QuotaLimit int       `json:"quota_limit,omitempty"`
}

// TURNQuotaManager tracks per-tenant TURN allocation quotas
type TURNQuotaManager struct {
	logger *zap.Logger

	// Quota limits per tenant
	defaultQuotaPerTenant    int
	defaultQuotaPerPeer      int
	defaultBandwidthQuotaMbps int

	// Current usage (in production use Redis/distributed cache)
	tenantAllocations map[string]*TenantQuota
	peerAllocations   map[string]*PeerQuota
}

// TenantQuota tracks TURN usage for a tenant
type TenantQuota struct {
	TenantID          string
	ActiveAllocations int
	TotalAllocations  int64
	BandwidthUsedMB   int64
	LastReset         time.Time
	QuotaLimit        int
}

// PeerQuota tracks TURN usage for a specific peer
type PeerQuota struct {
	PeerID            string
	TenantID          string
	ActiveAllocations int
	TotalAllocations  int64
	BandwidthUsedMB   int64
	FirstAllocation   time.Time
	LastAllocation    time.Time
}

// NewTURNCredentialManager creates a new credential manager
func NewTURNCredentialManager(sharedSecret string, logger *zap.Logger) *TURNCredentialManager {
	return &TURNCredentialManager{
		logger:       logger,
		sharedSecret: sharedSecret,
		defaultTTL:   10 * time.Minute, // 10 minutes per security best practice
		minTTL:       1 * time.Minute,
		maxTTL:       60 * time.Minute,
		quotaManager: NewTURNQuotaManager(logger),
	}
}

// NewTURNQuotaManager creates a new quota manager
func NewTURNQuotaManager(logger *zap.Logger) *TURNQuotaManager {
	return &TURNQuotaManager{
		logger:                    logger,
		defaultQuotaPerTenant:     100,  // 100 concurrent allocations per tenant
		defaultQuotaPerPeer:       10,   // 10 concurrent allocations per peer
		defaultBandwidthQuotaMbps: 1000, // 1 Gbps per tenant
		tenantAllocations:         make(map[string]*TenantQuota),
		peerAllocations:           make(map[string]*PeerQuota),
	}
}

// GenerateShortLivedCredential generates time-limited TURN credentials using HMAC
// Implements RFC 5766 Section 10.2 (short-term credential mechanism)
func (tcm *TURNCredentialManager) GenerateShortLivedCredential(
	tenantID string,
	peerID string,
	requestedTTL time.Duration,
) (*TURNCredentialResponse, error) {

	// Validate TTL
	ttl := tcm.defaultTTL
	if requestedTTL > 0 {
		if requestedTTL < tcm.minTTL {
			ttl = tcm.minTTL
		} else if requestedTTL > tcm.maxTTL {
			ttl = tcm.maxTTL
		} else {
			ttl = requestedTTL
		}
	}

	// Check quota
	if err := tcm.quotaManager.CheckQuota(tenantID, peerID); err != nil {
		tcm.logger.Warn("TURN quota exceeded",
			zap.String("tenant_id", tenantID),
			zap.String("peer_id", peerID),
			zap.Error(err))
		return nil, err
	}

	// Calculate expiration timestamp
	expiresAt := time.Now().Add(ttl)
	expiryTimestamp := expiresAt.Unix()

	// Generate username: <expiry_timestamp>:<peer_id>
	// Format per RFC 5766: timestamp ensures credentials expire automatically
	username := fmt.Sprintf("%d:%s", expiryTimestamp, peerID)

	// Generate password using HMAC-SHA1(shared_secret, username)
	// This allows the TURN server to validate credentials without central coordination
	password := tcm.generateHMACPassword(username)

	// Increment quota
	tcm.quotaManager.IncrementAllocation(tenantID, peerID)

	// Get tenant quota info
	quota := tcm.quotaManager.GetTenantQuota(tenantID)

	tcm.logger.Info("Generated short-lived TURN credential",
		zap.String("tenant_id", tenantID),
		zap.String("peer_id", peerID),
		zap.Duration("ttl", ttl),
		zap.Time("expires_at", expiresAt),
		zap.Int("quota_used", quota.ActiveAllocations))

	return &TURNCredentialResponse{
		Username:   username,
		Password:   password,
		TTL:        int(ttl.Seconds()),
		ExpiresAt:  expiresAt,
		URIs:       tcm.getTURNURIs(),
		Realm:      "cloudbridge.2gc.ru",
		QuotaUsed:  quota.ActiveAllocations,
		QuotaLimit: quota.QuotaLimit,
	}, nil
}

// generateHMACPassword generates HMAC-SHA1 password for TURN authentication
func (tcm *TURNCredentialManager) generateHMACPassword(username string) string {
	h := hmac.New(sha1.New, []byte(tcm.sharedSecret))
	h.Write([]byte(username))
	hash := h.Sum(nil)
	return base64.StdEncoding.EncodeToString(hash)
}

// ValidateCredential validates a TURN credential (used by TURN server)
func (tcm *TURNCredentialManager) ValidateCredential(username, password string) (bool, error) {
	// Parse username to extract timestamp and peer_id
	var expiryTimestamp int64
	var peerID string

	// Username format: <timestamp>:<peer_id>
	n, err := fmt.Sscanf(username, "%d:%s", &expiryTimestamp, &peerID)
	if err != nil || n != 2 {
		tcm.logger.Warn("Invalid TURN username format", zap.String("username", username))
		return false, fmt.Errorf("invalid username format")
	}

	// Check expiration
	expiresAt := time.Unix(expiryTimestamp, 0)
	if time.Now().After(expiresAt) {
		tcm.logger.Info("TURN credential expired",
			zap.String("username", username),
			zap.Time("expires_at", expiresAt))
		return false, fmt.Errorf("credential expired")
	}

	// Validate password using HMAC
	expectedPassword := tcm.generateHMACPassword(username)
	if !hmac.Equal([]byte(password), []byte(expectedPassword)) {
		tcm.logger.Warn("Invalid TURN password", zap.String("username", username))
		return false, fmt.Errorf("invalid password")
	}

	return true, nil
}

// getTURNURIs returns list of TURN server URIs
func (tcm *TURNCredentialManager) getTURNURIs() []string {
	// In production, load from config or service discovery
	return []string{
		"turn:turn.cloudbridge.2gc.ru:3478?transport=udp",
		"turn:turn.cloudbridge.2gc.ru:3478?transport=tcp",
		"turns:turn.cloudbridge.2gc.ru:5349?transport=tcp",
	}
}

// CheckQuota checks if tenant/peer can create new allocation
func (qm *TURNQuotaManager) CheckQuota(tenantID, peerID string) error {
	// Check tenant quota
	tenantQuota := qm.GetTenantQuota(tenantID)
	if tenantQuota.ActiveAllocations >= tenantQuota.QuotaLimit {
		return fmt.Errorf("tenant quota exceeded: %d/%d allocations",
			tenantQuota.ActiveAllocations, tenantQuota.QuotaLimit)
	}

	// Check peer quota
	peerQuota := qm.GetPeerQuota(peerID)
	if peerQuota.ActiveAllocations >= qm.defaultQuotaPerPeer {
		return fmt.Errorf("peer quota exceeded: %d/%d allocations",
			peerQuota.ActiveAllocations, qm.defaultQuotaPerPeer)
	}

	return nil
}

// IncrementAllocation increments allocation counters
func (qm *TURNQuotaManager) IncrementAllocation(tenantID, peerID string) {
	// Tenant
	tenantQuota := qm.GetTenantQuota(tenantID)
	tenantQuota.ActiveAllocations++
	tenantQuota.TotalAllocations++
	qm.tenantAllocations[tenantID] = tenantQuota

	// Peer
	peerQuota := qm.GetPeerQuota(peerID)
	peerQuota.TenantID = tenantID
	peerQuota.ActiveAllocations++
	peerQuota.TotalAllocations++
	peerQuota.LastAllocation = time.Now()
	if peerQuota.FirstAllocation.IsZero() {
		peerQuota.FirstAllocation = time.Now()
	}
	qm.peerAllocations[peerID] = peerQuota
}

// DecrementAllocation decrements allocation counters (called when allocation is released)
func (qm *TURNQuotaManager) DecrementAllocation(tenantID, peerID string) {
	if tenantQuota, exists := qm.tenantAllocations[tenantID]; exists {
		if tenantQuota.ActiveAllocations > 0 {
			tenantQuota.ActiveAllocations--
		}
		qm.tenantAllocations[tenantID] = tenantQuota
	}

	if peerQuota, exists := qm.peerAllocations[peerID]; exists {
		if peerQuota.ActiveAllocations > 0 {
			peerQuota.ActiveAllocations--
		}
		qm.peerAllocations[peerID] = peerQuota
	}
}

// GetTenantQuota returns quota info for tenant
func (qm *TURNQuotaManager) GetTenantQuota(tenantID string) *TenantQuota {
	if quota, exists := qm.tenantAllocations[tenantID]; exists {
		return quota
	}

	// Create new quota
	return &TenantQuota{
		TenantID:          tenantID,
		ActiveAllocations: 0,
		TotalAllocations:  0,
		BandwidthUsedMB:   0,
		LastReset:         time.Now(),
		QuotaLimit:        qm.defaultQuotaPerTenant,
	}
}

// GetPeerQuota returns quota info for peer
func (qm *TURNQuotaManager) GetPeerQuota(peerID string) *PeerQuota {
	if quota, exists := qm.peerAllocations[peerID]; exists {
		return quota
	}

	// Create new quota
	return &PeerQuota{
		PeerID:            peerID,
		ActiveAllocations: 0,
		TotalAllocations:  0,
		BandwidthUsedMB:   0,
		FirstAllocation:   time.Time{},
		LastAllocation:    time.Time{},
	}
}

// SetTenantQuota sets custom quota for tenant
func (qm *TURNQuotaManager) SetTenantQuota(tenantID string, limit int) {
	quota := qm.GetTenantQuota(tenantID)
	quota.QuotaLimit = limit
	qm.tenantAllocations[tenantID] = quota

	qm.logger.Info("Updated tenant TURN quota",
		zap.String("tenant_id", tenantID),
		zap.Int("limit", limit))
}

// GetQuotaStats returns quota statistics
func (qm *TURNQuotaManager) GetQuotaStats() map[string]interface{} {
	totalActive := 0
	totalAllocations := int64(0)

	for _, quota := range qm.tenantAllocations {
		totalActive += quota.ActiveAllocations
		totalAllocations += quota.TotalAllocations
	}

	return map[string]interface{}{
		"tenants_count":       len(qm.tenantAllocations),
		"peers_count":         len(qm.peerAllocations),
		"active_allocations":  totalActive,
		"total_allocations":   totalAllocations,
		"default_quota":       qm.defaultQuotaPerTenant,
		"default_peer_quota":  qm.defaultQuotaPerPeer,
		"bandwidth_quota_mbps": qm.defaultBandwidthQuotaMbps,
	}
}

// CleanupExpiredQuotas removes inactive quotas (call periodically)
func (qm *TURNQuotaManager) CleanupExpiredQuotas(inactiveThreshold time.Duration) int {
	cleaned := 0
	now := time.Now()

	// Clean peer quotas
	for peerID, quota := range qm.peerAllocations {
		if quota.ActiveAllocations == 0 &&
			now.Sub(quota.LastAllocation) > inactiveThreshold {
			delete(qm.peerAllocations, peerID)
			cleaned++
		}
	}

	qm.logger.Info("Cleaned up expired TURN quotas",
		zap.Int("cleaned_count", cleaned),
		zap.Duration("threshold", inactiveThreshold))

	return cleaned
}

// ExtractPeerIDFromUsername extracts peer_id from TURN username
func ExtractPeerIDFromUsername(username string) (string, error) {
	var timestamp int64
	var peerID string

	n, err := fmt.Sscanf(username, "%d:%s", &timestamp, &peerID)
	if err != nil || n != 2 {
		return "", fmt.Errorf("invalid username format")
	}

	return peerID, nil
}

// ParseExpiryFromUsername extracts expiry timestamp from TURN username
func ParseExpiryFromUsername(username string) (time.Time, error) {
	var timestamp int64
	var peerID string

	n, err := fmt.Sscanf(username, "%d:%s", &timestamp, &peerID)
	if err != nil || n != 2 {
		return time.Time{}, fmt.Errorf("invalid username format")
	}

	return time.Unix(timestamp, 0), nil
}

// IsCredentialExpired checks if credential is expired based on username
func IsCredentialExpired(username string) bool {
	expiresAt, err := ParseExpiryFromUsername(username)
	if err != nil {
		return true
	}
	return time.Now().After(expiresAt)
}
