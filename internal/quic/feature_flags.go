package quic

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
)

// FeatureFlags manages per-tenant QUIC experimental feature toggles
// Addresses security concern: BBRv2/ACK-freq/FEC can break old stacks
type FeatureFlags struct {
	logger *zap.Logger
	mu     sync.RWMutex

	// Global defaults
	globalConfig FeatureConfig

	// Per-tenant overrides
	tenantConfigs map[string]*FeatureConfig

	// Canary rollout settings
	canaryPercentage int
	canaryTenants    map[string]bool
}

// FeatureConfig represents QUIC feature configuration
type FeatureConfig struct {
	// BBRv2 Congestion Control
	BBRv2Enabled bool `json:"bbrv2_enabled"`

	// ACK Frequency Optimization
	ACKFrequencyEnabled bool `json:"ack_frequency_enabled"`

	// Forward Error Correction
	FECEnabled bool `json:"fec_enabled"`

	// QUIC Bit Greasing (RFC 9287)
	GreasingEnabled bool `json:"greasing_enabled"`

	// 0-RTT (potentially risky)
	ZeroRTTEnabled bool `json:"zero_rtt_enabled"`

	// HTTP/3 QPACK dynamic table
	QPACKDynamicTableEnabled bool `json:"qpack_dynamic_table_enabled"`

	// Metadata
	LastUpdated time.Time `json:"last_updated"`
	UpdatedBy   string    `json:"updated_by"`
	Reason      string    `json:"reason,omitempty"`
}

// NewFeatureFlags creates a new feature flags manager with safe defaults
func NewFeatureFlags(logger *zap.Logger) *FeatureFlags {
	return &FeatureFlags{
		logger: logger,
		// SAFE DEFAULTS: All experimental features OFF
		globalConfig: FeatureConfig{
			BBRv2Enabled:             false, // Default: Cubic
			ACKFrequencyEnabled:      false, // Default: RFC 9000
			FECEnabled:               false, // Default: No FEC
			GreasingEnabled:          true,  // Safe: RFC 9287
			ZeroRTTEnabled:           false, // Security concern
			QPACKDynamicTableEnabled: true,  // Standard HTTP/3
			LastUpdated:              time.Now(),
			UpdatedBy:                "system",
			Reason:                   "safe defaults",
		},
		tenantConfigs:    make(map[string]*FeatureConfig),
		canaryPercentage: 0, // 0% canary by default
		canaryTenants:    make(map[string]bool),
	}
}

// GetConfig returns feature configuration for tenant
func (ff *FeatureFlags) GetConfig(tenantID string) FeatureConfig {
	ff.mu.RLock()
	defer ff.mu.RUnlock()

	// Check tenant-specific override
	if config, exists := ff.tenantConfigs[tenantID]; exists {
		ff.logger.Debug("Using tenant-specific config",
			zap.String("tenant_id", tenantID))
		return *config
	}

	// Check if tenant is in canary group
	if ff.canaryTenants[tenantID] {
		ff.logger.Debug("Tenant in canary group, using canary config",
			zap.String("tenant_id", tenantID))
		return ff.getCanaryConfig()
	}

	// Use global default
	return ff.globalConfig
}

// SetGlobalConfig updates global default configuration
func (ff *FeatureFlags) SetGlobalConfig(config FeatureConfig, updatedBy string, reason string) {
	ff.mu.Lock()
	defer ff.mu.Unlock()

	config.LastUpdated = time.Now()
	config.UpdatedBy = updatedBy
	config.Reason = reason

	ff.globalConfig = config

	ff.logger.Info("Updated global QUIC feature config",
		zap.Bool("bbrv2", config.BBRv2Enabled),
		zap.Bool("ack_freq", config.ACKFrequencyEnabled),
		zap.Bool("fec", config.FECEnabled),
		zap.String("updated_by", updatedBy),
		zap.String("reason", reason))
}

// SetTenantConfig sets per-tenant configuration override
func (ff *FeatureFlags) SetTenantConfig(tenantID string, config FeatureConfig, updatedBy string, reason string) {
	ff.mu.Lock()
	defer ff.mu.Unlock()

	config.LastUpdated = time.Now()
	config.UpdatedBy = updatedBy
	config.Reason = reason

	ff.tenantConfigs[tenantID] = &config

	ff.logger.Info("Updated tenant QUIC feature config",
		zap.String("tenant_id", tenantID),
		zap.Bool("bbrv2", config.BBRv2Enabled),
		zap.Bool("ack_freq", config.ACKFrequencyEnabled),
		zap.Bool("fec", config.FECEnabled),
		zap.String("updated_by", updatedBy))
}

// RemoveTenantConfig removes tenant-specific override
func (ff *FeatureFlags) RemoveTenantConfig(tenantID string) {
	ff.mu.Lock()
	defer ff.mu.Unlock()

	delete(ff.tenantConfigs, tenantID)
	ff.logger.Info("Removed tenant config override", zap.String("tenant_id", tenantID))
}

// EnableCanary enables canary rollout for percentage of traffic
func (ff *FeatureFlags) EnableCanary(percentage int, features FeatureConfig) error {
	if percentage < 0 || percentage > 100 {
		return ErrInvalidCanaryPercentage
	}

	ff.mu.Lock()
	defer ff.mu.Unlock()

	ff.canaryPercentage = percentage

	ff.logger.Info("Enabled canary rollout for QUIC features",
		zap.Int("percentage", percentage),
		zap.Bool("bbrv2", features.BBRv2Enabled),
		zap.Bool("ack_freq", features.ACKFrequencyEnabled),
		zap.Bool("fec", features.FECEnabled))

	return nil
}

// AddToCanary adds tenant to canary group
func (ff *FeatureFlags) AddToCanary(tenantID string) {
	ff.mu.Lock()
	defer ff.mu.Unlock()

	ff.canaryTenants[tenantID] = true
	ff.logger.Info("Added tenant to canary group", zap.String("tenant_id", tenantID))
}

// RemoveFromCanary removes tenant from canary group
func (ff *FeatureFlags) RemoveFromCanary(tenantID string) {
	ff.mu.Lock()
	defer ff.mu.Unlock()

	delete(ff.canaryTenants, tenantID)
	ff.logger.Info("Removed tenant from canary group", zap.String("tenant_id", tenantID))
}

// getCanaryConfig returns configuration for canary tenants
func (ff *FeatureFlags) getCanaryConfig() FeatureConfig {
	// Canary config: enable ONE experimental feature at a time
	return FeatureConfig{
		BBRv2Enabled:             true,  // Test BBRv2 first
		ACKFrequencyEnabled:      false, // Test separately
		FECEnabled:               false, // Test separately
		GreasingEnabled:          true,
		ZeroRTTEnabled:           false,
		QPACKDynamicTableEnabled: true,
		LastUpdated:              time.Now(),
		UpdatedBy:                "canary",
		Reason:                   "canary rollout",
	}
}

// IsBBRv2Enabled checks if BBRv2 is enabled for tenant
func (ff *FeatureFlags) IsBBRv2Enabled(tenantID string) bool {
	config := ff.GetConfig(tenantID)
	return config.BBRv2Enabled
}

// IsACKFrequencyEnabled checks if ACK Frequency is enabled for tenant
func (ff *FeatureFlags) IsACKFrequencyEnabled(tenantID string) bool {
	config := ff.GetConfig(tenantID)
	return config.ACKFrequencyEnabled
}

// IsFECEnabled checks if FEC is enabled for tenant
func (ff *FeatureFlags) IsFECEnabled(tenantID string) bool {
	config := ff.GetConfig(tenantID)
	return config.FECEnabled
}

// IsGreasingEnabled checks if Greasing is enabled for tenant
func (ff *FeatureFlags) IsGreasingEnabled(tenantID string) bool {
	config := ff.GetConfig(tenantID)
	return config.GreasingEnabled
}

// GetStats returns feature flag statistics
func (ff *FeatureFlags) GetStats() map[string]interface{} {
	ff.mu.RLock()
	defer ff.mu.RUnlock()

	return map[string]interface{}{
		"global_config": map[string]interface{}{
			"bbrv2_enabled":        ff.globalConfig.BBRv2Enabled,
			"ack_freq_enabled":     ff.globalConfig.ACKFrequencyEnabled,
			"fec_enabled":          ff.globalConfig.FECEnabled,
			"greasing_enabled":     ff.globalConfig.GreasingEnabled,
			"zero_rtt_enabled":     ff.globalConfig.ZeroRTTEnabled,
			"qpack_dynamic_table":  ff.globalConfig.QPACKDynamicTableEnabled,
			"last_updated":         ff.globalConfig.LastUpdated,
			"updated_by":           ff.globalConfig.UpdatedBy,
		},
		"tenant_overrides_count": len(ff.tenantConfigs),
		"canary_percentage":      ff.canaryPercentage,
		"canary_tenants_count":   len(ff.canaryTenants),
	}
}

// ExportConfig exports configuration as JSON
func (ff *FeatureFlags) ExportConfig() ([]byte, error) {
	ff.mu.RLock()
	defer ff.mu.RUnlock()

	export := map[string]interface{}{
		"global":        ff.globalConfig,
		"tenant_configs": ff.tenantConfigs,
		"canary": map[string]interface{}{
			"percentage": ff.canaryPercentage,
			"tenants":    ff.canaryTenants,
		},
	}

	return json.MarshalIndent(export, "", "  ")
}

// ImportConfig imports configuration from JSON
func (ff *FeatureFlags) ImportConfig(data []byte) error {
	var imported struct {
		Global        FeatureConfig                  `json:"global"`
		TenantConfigs map[string]*FeatureConfig      `json:"tenant_configs"`
		Canary        struct {
			Percentage int               `json:"percentage"`
			Tenants    map[string]bool   `json:"tenants"`
		} `json:"canary"`
	}

	if err := json.Unmarshal(data, &imported); err != nil {
		return err
	}

	ff.mu.Lock()
	defer ff.mu.Unlock()

	ff.globalConfig = imported.Global
	ff.tenantConfigs = imported.TenantConfigs
	ff.canaryPercentage = imported.Canary.Percentage
	ff.canaryTenants = imported.Canary.Tenants

	ff.logger.Info("Imported QUIC feature flags configuration",
		zap.Int("tenant_configs", len(ff.tenantConfigs)),
		zap.Int("canary_percentage", ff.canaryPercentage))

	return nil
}

// LoadFromFile loads configuration from file
func (ff *FeatureFlags) LoadFromFile(filepath string) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	return ff.ImportConfig(data)
}

// SaveToFile saves configuration to file
func (ff *FeatureFlags) SaveToFile(filepath string) error {
	data, err := ff.ExportConfig()
	if err != nil {
		return err
	}

	return os.WriteFile(filepath, data, 0644)
}

// WatchContext monitors context for cancellation and performs cleanup
func (ff *FeatureFlags) WatchContext(ctx context.Context) {
	<-ctx.Done()
	ff.logger.Info("Feature flags context cancelled, cleaning up")
}

// Errors
var (
	ErrInvalidCanaryPercentage = fmt.Errorf("canary percentage must be 0-100")
)
