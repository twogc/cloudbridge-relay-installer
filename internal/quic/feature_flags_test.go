package quic

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewFeatureFlags(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ff := NewFeatureFlags(logger)

	assert.NotNil(t, ff)
	assert.NotNil(t, ff.logger)

	// Verify safe defaults (all experimental features OFF)
	assert.False(t, ff.globalConfig.BBRv2Enabled)
	assert.False(t, ff.globalConfig.ACKFrequencyEnabled)
	assert.False(t, ff.globalConfig.FECEnabled)
	assert.True(t, ff.globalConfig.GreasingEnabled) // Safe: RFC 9287
	assert.False(t, ff.globalConfig.ZeroRTTEnabled)
}

func TestGetConfig_DefaultGlobal(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ff := NewFeatureFlags(logger)

	config := ff.GetConfig("tenant-1")

	// Should return global defaults
	assert.False(t, config.BBRv2Enabled)
	assert.False(t, config.ACKFrequencyEnabled)
	assert.False(t, config.FECEnabled)
	assert.True(t, config.GreasingEnabled)
}

func TestSetGlobalConfig(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ff := NewFeatureFlags(logger)

	newConfig := FeatureConfig{
		BBRv2Enabled:        true,
		ACKFrequencyEnabled: true,
		FECEnabled:          false,
		GreasingEnabled:     true,
	}

	ff.SetGlobalConfig(newConfig, "admin", "enable BBRv2 globally")

	// Verify update
	config := ff.GetConfig("any-tenant")
	assert.True(t, config.BBRv2Enabled)
	assert.True(t, config.ACKFrequencyEnabled)
	assert.False(t, config.FECEnabled)
	assert.Equal(t, "admin", config.UpdatedBy)
	assert.Equal(t, "enable BBRv2 globally", config.Reason)
}

func TestSetTenantConfig(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ff := NewFeatureFlags(logger)

	// Set tenant-specific override
	tenantConfig := FeatureConfig{
		BBRv2Enabled:        true,
		ACKFrequencyEnabled: true,
		FECEnabled:          true,
		GreasingEnabled:     true,
	}

	ff.SetTenantConfig("premium-tenant", tenantConfig, "admin", "premium features")

	// Verify tenant override
	config := ff.GetConfig("premium-tenant")
	assert.True(t, config.BBRv2Enabled)
	assert.True(t, config.FECEnabled)
	assert.Equal(t, "premium features", config.Reason)

	// Other tenants should use global defaults
	otherConfig := ff.GetConfig("standard-tenant")
	assert.False(t, otherConfig.BBRv2Enabled)
	assert.False(t, otherConfig.FECEnabled)
}

func TestRemoveTenantConfig(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ff := NewFeatureFlags(logger)

	// Set tenant config
	tenantConfig := FeatureConfig{BBRv2Enabled: true}
	ff.SetTenantConfig("tenant-1", tenantConfig, "admin", "test")

	// Verify it exists
	config := ff.GetConfig("tenant-1")
	assert.True(t, config.BBRv2Enabled)

	// Remove override
	ff.RemoveTenantConfig("tenant-1")

	// Should fall back to global defaults
	config = ff.GetConfig("tenant-1")
	assert.False(t, config.BBRv2Enabled)
}

func TestEnableCanary(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ff := NewFeatureFlags(logger)

	canaryConfig := FeatureConfig{
		BBRv2Enabled: true,
	}

	err := ff.EnableCanary(5, canaryConfig)
	require.NoError(t, err)

	// Verify canary percentage
	stats := ff.GetStats()
	assert.Equal(t, 5, stats["canary_percentage"])
}

func TestEnableCanary_InvalidPercentage(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ff := NewFeatureFlags(logger)

	canaryConfig := FeatureConfig{}

	// Test invalid percentages
	err := ff.EnableCanary(-1, canaryConfig)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidCanaryPercentage, err)

	err = ff.EnableCanary(101, canaryConfig)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidCanaryPercentage, err)

	// Valid percentages
	err = ff.EnableCanary(0, canaryConfig)
	assert.NoError(t, err)

	err = ff.EnableCanary(100, canaryConfig)
	assert.NoError(t, err)
}

func TestAddToCanary(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ff := NewFeatureFlags(logger)

	ff.AddToCanary("canary-tenant-1")
	ff.AddToCanary("canary-tenant-2")

	// Verify canary tenants
	stats := ff.GetStats()
	assert.Equal(t, 2, stats["canary_tenants_count"])

	// Canary tenants should get canary config
	config := ff.GetConfig("canary-tenant-1")
	assert.True(t, config.BBRv2Enabled) // Canary config enables BBRv2
}

func TestRemoveFromCanary(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ff := NewFeatureFlags(logger)

	ff.AddToCanary("tenant-1")
	ff.RemoveFromCanary("tenant-1")

	// Should use global defaults
	config := ff.GetConfig("tenant-1")
	assert.False(t, config.BBRv2Enabled)
}

func TestIsBBRv2Enabled(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ff := NewFeatureFlags(logger)

	// Default: OFF
	assert.False(t, ff.IsBBRv2Enabled("tenant-1"))

	// Enable for specific tenant
	config := FeatureConfig{BBRv2Enabled: true}
	ff.SetTenantConfig("tenant-1", config, "admin", "test")

	assert.True(t, ff.IsBBRv2Enabled("tenant-1"))
	assert.False(t, ff.IsBBRv2Enabled("tenant-2"))
}

func TestIsACKFrequencyEnabled(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ff := NewFeatureFlags(logger)

	// Default: OFF
	assert.False(t, ff.IsACKFrequencyEnabled("tenant-1"))

	// Enable globally
	config := FeatureConfig{ACKFrequencyEnabled: true}
	ff.SetGlobalConfig(config, "admin", "test")

	assert.True(t, ff.IsACKFrequencyEnabled("tenant-1"))
	assert.True(t, ff.IsACKFrequencyEnabled("tenant-2"))
}

func TestIsFECEnabled(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ff := NewFeatureFlags(logger)

	// Default: OFF
	assert.False(t, ff.IsFECEnabled("tenant-1"))

	// Enable via canary
	ff.AddToCanary("tenant-1")

	// Canary config doesn't enable FEC by default
	assert.False(t, ff.IsFECEnabled("tenant-1"))
}

func TestIsGreasingEnabled(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ff := NewFeatureFlags(logger)

	// Greasing is ON by default (safe, RFC 9287)
	assert.True(t, ff.IsGreasingEnabled("tenant-1"))
}

func TestGetStats(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ff := NewFeatureFlags(logger)

	// Add some configs
	ff.SetTenantConfig("tenant-1", FeatureConfig{BBRv2Enabled: true}, "admin", "test")
	ff.SetTenantConfig("tenant-2", FeatureConfig{FECEnabled: true}, "admin", "test")
	ff.AddToCanary("canary-tenant")
	ff.EnableCanary(10, FeatureConfig{})

	stats := ff.GetStats()

	assert.Equal(t, 2, stats["tenant_overrides_count"])
	assert.Equal(t, 10, stats["canary_percentage"])
	assert.Equal(t, 1, stats["canary_tenants_count"])

	globalConfig := stats["global_config"].(map[string]interface{})
	assert.False(t, globalConfig["bbrv2_enabled"].(bool))
	assert.True(t, globalConfig["greasing_enabled"].(bool))
}

func TestExportImportConfig(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ff := NewFeatureFlags(logger)

	// Setup some configuration
	ff.SetGlobalConfig(FeatureConfig{BBRv2Enabled: true}, "admin", "test")
	ff.SetTenantConfig("tenant-1", FeatureConfig{FECEnabled: true}, "admin", "test")
	ff.AddToCanary("canary-tenant")
	ff.EnableCanary(5, FeatureConfig{})

	// Export
	data, err := ff.ExportConfig()
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Create new instance and import
	ff2 := NewFeatureFlags(logger)
	err = ff2.ImportConfig(data)
	require.NoError(t, err)

	// Verify imported data
	assert.True(t, ff2.GetConfig("any-tenant").BBRv2Enabled)
	assert.True(t, ff2.GetConfig("tenant-1").FECEnabled)
	assert.True(t, ff2.GetConfig("canary-tenant").BBRv2Enabled)

	stats := ff2.GetStats()
	assert.Equal(t, 5, stats["canary_percentage"])
}

func TestLoadSaveFile(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ff := NewFeatureFlags(logger)

	// Setup configuration
	ff.SetGlobalConfig(FeatureConfig{BBRv2Enabled: true, ACKFrequencyEnabled: true}, "admin", "test")
	ff.SetTenantConfig("premium", FeatureConfig{FECEnabled: true}, "admin", "premium tier")

	// Save to temp file
	tmpFile := "/tmp/quic-feature-flags-test.json"
	defer os.Remove(tmpFile)

	err := ff.SaveToFile(tmpFile)
	require.NoError(t, err)

	// Load into new instance
	ff2 := NewFeatureFlags(logger)
	err = ff2.LoadFromFile(tmpFile)
	require.NoError(t, err)

	// Verify
	assert.True(t, ff2.GetConfig("any-tenant").BBRv2Enabled)
	assert.True(t, ff2.GetConfig("any-tenant").ACKFrequencyEnabled)
	assert.True(t, ff2.GetConfig("premium").FECEnabled)
}

func TestTenantConfigPriority(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ff := NewFeatureFlags(logger)

	// Set global: BBRv2=ON
	ff.SetGlobalConfig(FeatureConfig{BBRv2Enabled: true}, "admin", "global")

	// Set tenant override: BBRv2=OFF
	ff.SetTenantConfig("tenant-1", FeatureConfig{BBRv2Enabled: false}, "admin", "tenant specific")

	// Add to canary (would normally enable BBRv2)
	ff.AddToCanary("tenant-1")

	// Tenant override should win
	config := ff.GetConfig("tenant-1")
	assert.False(t, config.BBRv2Enabled)
}

func TestCanaryConfig(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ff := NewFeatureFlags(logger)

	// Add tenant to canary
	ff.AddToCanary("canary-tenant")

	config := ff.GetConfig("canary-tenant")

	// Canary config: BBRv2=ON, others OFF (test one feature at a time)
	assert.True(t, config.BBRv2Enabled)
	assert.False(t, config.ACKFrequencyEnabled)
	assert.False(t, config.FECEnabled)
	assert.True(t, config.GreasingEnabled)
	assert.Equal(t, "canary", config.UpdatedBy)
}

func BenchmarkGetConfig(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	ff := NewFeatureFlags(logger)

	// Add some tenant configs
	for i := 0; i < 100; i++ {
		ff.SetTenantConfig(fmt.Sprintf("tenant-%d", i), FeatureConfig{BBRv2Enabled: true}, "admin", "test")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ff.GetConfig("tenant-50")
	}
}

func BenchmarkIsBBRv2Enabled(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	ff := NewFeatureFlags(logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ff.IsBBRv2Enabled("tenant-1")
	}
}
