package nat

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewTURNCredentialManager(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tcm := NewTURNCredentialManager("test-secret", logger)

	assert.NotNil(t, tcm)
	assert.Equal(t, "test-secret", tcm.sharedSecret)
	assert.Equal(t, 10*time.Minute, tcm.defaultTTL)
	assert.NotNil(t, tcm.quotaManager)
}

func TestGenerateShortLivedCredential_Success(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tcm := NewTURNCredentialManager("test-secret", logger)

	cred, err := tcm.GenerateShortLivedCredential("tenant-1", "peer-123", 5*time.Minute)
	require.NoError(t, err)
	require.NotNil(t, cred)

	// Verify fields
	assert.NotEmpty(t, cred.Username)
	assert.NotEmpty(t, cred.Password)
	assert.Equal(t, 300, cred.TTL) // 5 minutes = 300 seconds
	assert.False(t, cred.ExpiresAt.IsZero())
	assert.Greater(t, len(cred.URIs), 0)
	assert.Equal(t, "cloudbridge.2gc.ru", cred.Realm)

	// Username should contain peer_id
	assert.Contains(t, cred.Username, "peer-123")

	// ExpiresAt should be ~5 minutes from now
	expectedExpiry := time.Now().Add(5 * time.Minute)
	assert.WithinDuration(t, expectedExpiry, cred.ExpiresAt, 5*time.Second)
}

func TestGenerateShortLivedCredential_DefaultTTL(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tcm := NewTURNCredentialManager("test-secret", logger)

	cred, err := tcm.GenerateShortLivedCredential("tenant-1", "peer-123", 0)
	require.NoError(t, err)

	// Should use default TTL (10 minutes)
	assert.Equal(t, 600, cred.TTL)
}

func TestGenerateShortLivedCredential_TTLBounds(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tcm := NewTURNCredentialManager("test-secret", logger)

	// Test minimum TTL
	cred, err := tcm.GenerateShortLivedCredential("tenant-1", "peer-123", 30*time.Second)
	require.NoError(t, err)
	assert.Equal(t, 60, cred.TTL) // Should be clamped to minTTL (1 minute)

	// Test maximum TTL
	cred, err = tcm.GenerateShortLivedCredential("tenant-1", "peer-456", 120*time.Minute)
	require.NoError(t, err)
	assert.Equal(t, 3600, cred.TTL) // Should be clamped to maxTTL (60 minutes)
}

func TestValidateCredential_Success(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tcm := NewTURNCredentialManager("test-secret", logger)

	// Generate credential
	cred, err := tcm.GenerateShortLivedCredential("tenant-1", "peer-123", 5*time.Minute)
	require.NoError(t, err)

	// Validate it
	valid, err := tcm.ValidateCredential(cred.Username, cred.Password)
	assert.NoError(t, err)
	assert.True(t, valid)
}

func TestValidateCredential_InvalidPassword(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tcm := NewTURNCredentialManager("test-secret", logger)

	cred, err := tcm.GenerateShortLivedCredential("tenant-1", "peer-123", 5*time.Minute)
	require.NoError(t, err)

	// Try with wrong password
	valid, err := tcm.ValidateCredential(cred.Username, "wrong-password")
	assert.Error(t, err)
	assert.False(t, valid)
	assert.Contains(t, err.Error(), "invalid password")
}

func TestValidateCredential_InvalidUsername(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tcm := NewTURNCredentialManager("test-secret", logger)

	valid, err := tcm.ValidateCredential("invalid-format", "password")
	assert.Error(t, err)
	assert.False(t, valid)
	assert.Contains(t, err.Error(), "invalid username format")
}

func TestValidateCredential_Expired(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tcm := NewTURNCredentialManager("test-secret", logger)

	// Create credential with past expiry
	pastTimestamp := time.Now().Add(-1 * time.Minute).Unix()
	username := fmt.Sprintf("%d:peer-123", pastTimestamp)
	password := tcm.generateHMACPassword(username)

	valid, err := tcm.ValidateCredential(username, password)
	assert.Error(t, err)
	assert.False(t, valid)
	assert.Contains(t, err.Error(), "expired")
}

func TestTURNQuotaManager_CheckQuota_Success(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	qm := NewTURNQuotaManager(logger)

	err := qm.CheckQuota("tenant-1", "peer-123")
	assert.NoError(t, err)
}

func TestTURNQuotaManager_CheckQuota_TenantExceeded(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	qm := NewTURNQuotaManager(logger)

	// Set low quota
	qm.SetTenantQuota("tenant-1", 2)

	// Use up quota
	qm.IncrementAllocation("tenant-1", "peer-1")
	qm.IncrementAllocation("tenant-1", "peer-2")

	// Should fail
	err := qm.CheckQuota("tenant-1", "peer-3")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tenant quota exceeded")
}

func TestTURNQuotaManager_CheckQuota_PeerExceeded(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	qm := NewTURNQuotaManager(logger)
	qm.defaultQuotaPerPeer = 2

	// Use up peer quota
	qm.IncrementAllocation("tenant-1", "peer-1")
	qm.IncrementAllocation("tenant-1", "peer-1")

	// Should fail
	err := qm.CheckQuota("tenant-1", "peer-1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "peer quota exceeded")
}

func TestTURNQuotaManager_IncrementDecrement(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	qm := NewTURNQuotaManager(logger)

	// Increment
	qm.IncrementAllocation("tenant-1", "peer-123")

	tenantQuota := qm.GetTenantQuota("tenant-1")
	assert.Equal(t, 1, tenantQuota.ActiveAllocations)
	assert.Equal(t, int64(1), tenantQuota.TotalAllocations)

	peerQuota := qm.GetPeerQuota("peer-123")
	assert.Equal(t, 1, peerQuota.ActiveAllocations)
	assert.Equal(t, "tenant-1", peerQuota.TenantID)

	// Decrement
	qm.DecrementAllocation("tenant-1", "peer-123")

	tenantQuota = qm.GetTenantQuota("tenant-1")
	assert.Equal(t, 0, tenantQuota.ActiveAllocations)
	assert.Equal(t, int64(1), tenantQuota.TotalAllocations) // Total shouldn't decrease

	peerQuota = qm.GetPeerQuota("peer-123")
	assert.Equal(t, 0, peerQuota.ActiveAllocations)
}

func TestTURNQuotaManager_GetQuotaStats(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	qm := NewTURNQuotaManager(logger)

	qm.IncrementAllocation("tenant-1", "peer-1")
	qm.IncrementAllocation("tenant-1", "peer-2")
	qm.IncrementAllocation("tenant-2", "peer-3")

	stats := qm.GetQuotaStats()

	assert.Equal(t, 2, stats["tenants_count"])
	assert.Equal(t, 3, stats["peers_count"])
	assert.Equal(t, 3, stats["active_allocations"])
	assert.Equal(t, int64(3), stats["total_allocations"])
}

func TestTURNQuotaManager_CleanupExpiredQuotas(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	qm := NewTURNQuotaManager(logger)

	// Create some allocations
	qm.IncrementAllocation("tenant-1", "peer-1")
	qm.IncrementAllocation("tenant-1", "peer-2")

	// Release one
	qm.DecrementAllocation("tenant-1", "peer-1")

	// Make peer-1 look old
	if quota, exists := qm.peerAllocations["peer-1"]; exists {
		quota.LastAllocation = time.Now().Add(-2 * time.Hour)
		qm.peerAllocations["peer-1"] = quota
	}

	// Cleanup with 1 hour threshold
	cleaned := qm.CleanupExpiredQuotas(1 * time.Hour)

	assert.Equal(t, 1, cleaned)
	assert.Equal(t, 1, len(qm.peerAllocations)) // Only peer-2 should remain
}

func TestExtractPeerIDFromUsername(t *testing.T) {
	username := "1234567890:peer-123"
	peerID, err := ExtractPeerIDFromUsername(username)
	require.NoError(t, err)
	assert.Equal(t, "peer-123", peerID)
}

func TestExtractPeerIDFromUsername_Invalid(t *testing.T) {
	_, err := ExtractPeerIDFromUsername("invalid-format")
	assert.Error(t, err)
}

func TestParseExpiryFromUsername(t *testing.T) {
	expectedTime := time.Unix(1234567890, 0)
	username := "1234567890:peer-123"

	expiresAt, err := ParseExpiryFromUsername(username)
	require.NoError(t, err)
	assert.Equal(t, expectedTime, expiresAt)
}

func TestIsCredentialExpired(t *testing.T) {
	// Future timestamp
	futureTimestamp := time.Now().Add(1 * time.Hour).Unix()
	futureUsername := fmt.Sprintf("%d:peer-123", futureTimestamp)
	assert.False(t, IsCredentialExpired(futureUsername))

	// Past timestamp
	pastTimestamp := time.Now().Add(-1 * time.Hour).Unix()
	pastUsername := fmt.Sprintf("%d:peer-123", pastTimestamp)
	assert.True(t, IsCredentialExpired(pastUsername))
}

func TestGenerateShortLivedCredential_QuotaTracking(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tcm := NewTURNCredentialManager("test-secret", logger)

	// Generate multiple credentials
	cred1, err := tcm.GenerateShortLivedCredential("tenant-1", "peer-1", 5*time.Minute)
	require.NoError(t, err)
	assert.Equal(t, 1, cred1.QuotaUsed)

	cred2, err := tcm.GenerateShortLivedCredential("tenant-1", "peer-2", 5*time.Minute)
	require.NoError(t, err)
	assert.Equal(t, 2, cred2.QuotaUsed)

	assert.Equal(t, cred1.QuotaLimit, cred2.QuotaLimit)
}

func BenchmarkGenerateShortLivedCredential(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	tcm := NewTURNCredentialManager("test-secret", logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tcm.GenerateShortLivedCredential("tenant-1", fmt.Sprintf("peer-%d", i), 5*time.Minute)
	}
}

func BenchmarkValidateCredential(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	tcm := NewTURNCredentialManager("test-secret", logger)

	cred, _ := tcm.GenerateShortLivedCredential("tenant-1", "peer-123", 5*time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tcm.ValidateCredential(cred.Username, cred.Password)
	}
}
