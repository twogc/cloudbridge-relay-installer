package nat

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewTURNCredentialService(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tcs := NewTURNCredentialService("test-secret", "example.com", logger)

	assert.NotNil(t, tcs)
	assert.Equal(t, "test-secret", tcs.sharedSecret)
	assert.Equal(t, "example.com", tcs.realm)
	assert.Equal(t, 10*time.Minute, tcs.defaultTTL)
	assert.NotNil(t, tcs.quotaManager)
}

func TestGenerateCredential_Success(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tcs := NewTURNCredentialService("test-secret", "cloudbridge.2gc.ru", logger)

	cred, err := tcs.GenerateCredential("tenant-1", "peer-123", 5*time.Minute)
	require.NoError(t, err)
	require.NotNil(t, cred)

	// Verify format matches pion/turn standard
	// Username should be: <timestamp>:<peer_id>
	assert.Contains(t, cred.Username, ":")
	assert.Contains(t, cred.Username, "peer-123")

	// Verify fields
	assert.NotEmpty(t, cred.Password)
	assert.Equal(t, 300, cred.TTL)
	assert.False(t, cred.ExpiresAt.IsZero())
	assert.Greater(t, len(cred.URIs), 0)
	assert.Equal(t, "cloudbridge.2gc.ru", cred.Realm)

	// Verify username format: starts with timestamp
	parts := strings.Split(cred.Username, ":")
	assert.Len(t, parts, 2)
	timestamp := parts[0]
	assert.Regexp(t, `^\d+$`, timestamp) // Should be numeric
}

func TestGenerateCredential_DefaultTTL(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tcs := NewTURNCredentialService("test-secret", "cloudbridge.2gc.ru", logger)

	cred, err := tcs.GenerateCredential("tenant-1", "peer-123", 0)
	require.NoError(t, err)

	// Should use default TTL (10 minutes)
	assert.Equal(t, 600, cred.TTL)
}

func TestGenerateCredential_TTLBounds(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tcs := NewTURNCredentialService("test-secret", "cloudbridge.2gc.ru", logger)

	// Test minimum TTL
	cred, err := tcs.GenerateCredential("tenant-1", "peer-123", 30*time.Second)
	require.NoError(t, err)
	assert.Equal(t, 60, cred.TTL) // Clamped to minTTL (1 minute)

	// Test maximum TTL
	cred, err = tcs.GenerateCredential("tenant-1", "peer-456", 120*time.Minute)
	require.NoError(t, err)
	assert.Equal(t, 3600, cred.TTL) // Clamped to maxTTL (60 minutes)
}

func TestGenerateCredential_QuotaTracking(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tcs := NewTURNCredentialService("test-secret", "cloudbridge.2gc.ru", logger)

	// Generate multiple credentials
	cred1, err := tcs.GenerateCredential("tenant-1", "peer-1", 5*time.Minute)
	require.NoError(t, err)
	assert.Equal(t, 1, cred1.QuotaUsed)

	cred2, err := tcs.GenerateCredential("tenant-1", "peer-2", 5*time.Minute)
	require.NoError(t, err)
	assert.Equal(t, 2, cred2.QuotaUsed)

	// Both should have same limit
	assert.Equal(t, cred1.QuotaLimit, cred2.QuotaLimit)
}

func TestGenerateCredential_QuotaExceeded(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tcs := NewTURNCredentialService("test-secret", "cloudbridge.2gc.ru", logger)

	// Set low quota
	tcs.quotaManager.SetTenantQuota("tenant-1", 2)

	// Use up quota
	_, err := tcs.GenerateCredential("tenant-1", "peer-1", 5*time.Minute)
	require.NoError(t, err)

	_, err = tcs.GenerateCredential("tenant-1", "peer-2", 5*time.Minute)
	require.NoError(t, err)

	// Should fail quota check
	_, err = tcs.GenerateCredential("tenant-1", "peer-3", 5*time.Minute)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "quota exceeded")
}

func TestSetTTLLimits(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tcs := NewTURNCredentialService("test-secret", "cloudbridge.2gc.ru", logger)

	// Set custom limits
	tcs.SetTTLLimits(5*time.Minute, 30*time.Second, 30*time.Minute)

	assert.Equal(t, 5*time.Minute, tcs.defaultTTL)
	assert.Equal(t, 30*time.Second, tcs.minTTL)
	assert.Equal(t, 30*time.Minute, tcs.maxTTL)

	// Test with new limits
	cred, err := tcs.GenerateCredential("tenant-1", "peer-123", 0)
	require.NoError(t, err)
	assert.Equal(t, 300, cred.TTL) // 5 minutes
}

func TestHealthCheck_Healthy(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tcs := NewTURNCredentialService("test-secret", "cloudbridge.2gc.ru", logger)

	health := tcs.HealthCheck()

	assert.Equal(t, "turn_credential_service", health["component"])
	assert.Equal(t, "healthy", health["status"])
	assert.Equal(t, "pion/turn v3 (RFC 5766 compliant)", health["implementation"])
	assert.Equal(t, "draft-uberti-behave-turn-rest-00", health["standard"])
	assert.True(t, health["shared_secret_set"].(bool))
	assert.Equal(t, "cloudbridge.2gc.ru", health["realm"])
}

func TestHealthCheck_Unhealthy_NoSecret(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tcs := NewTURNCredentialService("", "cloudbridge.2gc.ru", logger)

	health := tcs.HealthCheck()

	assert.Equal(t, "unhealthy", health["status"])
	assert.False(t, health["shared_secret_set"].(bool))

	issues := health["issues"].([]string)
	assert.Contains(t, issues, "missing shared secret")
}

func TestCreatePionTURNAuthHandler(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	authHandler := CreatePionTURNAuthHandler("test-secret", logger)

	assert.NotNil(t, authHandler)
	// AuthHandler is a function type, so we can't test much more without a full TURN server
	// Integration tests would validate actual authentication
}

func TestGetQuotaManager(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tcs := NewTURNCredentialService("test-secret", "cloudbridge.2gc.ru", logger)

	qm := tcs.GetQuotaManager()
	assert.NotNil(t, qm)
	assert.Equal(t, tcs.quotaManager, qm)
}

// Integration test: Verify compatibility with pion/turn standards
func TestPionTURNCompatibility(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tcs := NewTURNCredentialService("shared-secret-123", "cloudbridge.2gc.ru", logger)

	// Generate credential
	cred, err := tcs.GenerateCredential("tenant-1", "peer-abc", 10*time.Minute)
	require.NoError(t, err)

	// Verify username format matches pion/turn spec
	// Format: <timestamp>:<user>
	parts := strings.Split(cred.Username, ":")
	require.Len(t, parts, 2, "Username should be timestamp:user format")

	timestamp := parts[0]
	user := parts[1]

	// Timestamp should be numeric and in the future
	assert.Regexp(t, `^\d+$`, timestamp)
	assert.Equal(t, "peer-abc", user)

	// Password should be base64-encoded HMAC-SHA1
	assert.NotEmpty(t, cred.Password)
	// Base64 strings are typically multiples of 4 in length
	assert.Equal(t, 0, len(cred.Password)%4)
}

func TestNormalizeTTL(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tcs := NewTURNCredentialService("test-secret", "cloudbridge.2gc.ru", logger)

	tests := []struct {
		name     string
		input    time.Duration
		expected time.Duration
	}{
		{"Zero returns default", 0, 10 * time.Minute},
		{"Negative returns default", -5 * time.Minute, 10 * time.Minute},
		{"Below min clamped to min", 30 * time.Second, 1 * time.Minute},
		{"Above max clamped to max", 120 * time.Minute, 60 * time.Minute},
		{"Within range unchanged", 15 * time.Minute, 15 * time.Minute},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tcs.normalizeTTL(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func BenchmarkGenerateCredential(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	tcs := NewTURNCredentialService("test-secret", "cloudbridge.2gc.ru", logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tcs.GenerateCredential("tenant-1", "peer-123", 5*time.Minute)
	}
}

func BenchmarkHealthCheck(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	tcs := NewTURNCredentialService("test-secret", "cloudbridge.2gc.ru", logger)

	// Generate some credentials first
	for i := 0; i < 10; i++ {
		tcs.GenerateCredential("tenant-1", fmt.Sprintf("peer-%d", i), 5*time.Minute)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tcs.HealthCheck()
	}
}
