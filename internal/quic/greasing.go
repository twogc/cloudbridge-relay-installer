package quic

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"

	"go.uber.org/zap"
)

// GreasingManager implements QUIC Bit Greasing (RFC 9287)
// Prevents ossification by using reserved bits in QUIC headers
// Integrated from quic-test experimental features
type GreasingManager struct {
	logger   *zap.Logger
	config   *GreasingConfig
	mu       sync.RWMutex
	metrics  *GreasingMetrics
	isActive bool
}

// GreasingConfig holds greasing configuration
type GreasingConfig struct {
	// Enable greasing for QUIC version negotiation
	EnableVersionGreasing bool

	// Enable greasing for transport parameters
	EnableTransportParamGreasing bool

	// Enable greasing for frame types
	EnableFrameTypeGreasing bool

	// Probability of using grease value (0.0 to 1.0)
	GreaseProbability float64

	// List of reserved values to use for greasing
	GreaseValues []uint64
}

// GreasingMetrics holds greasing metrics
type GreasingMetrics struct {
	VersionGreaseCount        int64 `json:"version_grease_count"`
	TransportParamGreaseCount int64 `json:"transport_param_grease_count"`
	FrameTypeGreaseCount      int64 `json:"frame_type_grease_count"`
	TotalGreaseApplications   int64 `json:"total_grease_applications"`
	CompatibilityIssues       int64 `json:"compatibility_issues"`
}

// NewGreasingManager creates a new greasing manager
func NewGreasingManager(logger *zap.Logger, config *GreasingConfig) *GreasingManager {
	if config == nil {
		config = DefaultGreasingConfig()
	}

	return &GreasingManager{
		logger:   logger,
		config:   config,
		metrics:  &GreasingMetrics{},
		isActive: true,
	}
}

// DefaultGreasingConfig returns default greasing configuration
func DefaultGreasingConfig() *GreasingConfig {
	return &GreasingConfig{
		EnableVersionGreasing:        true,
		EnableTransportParamGreasing: true,
		EnableFrameTypeGreasing:      true,
		GreaseProbability:            0.5, // 50% chance to apply greasing

		// RFC 9287 reserved values for greasing
		// These values are reserved and MUST be ignored by peers
		GreaseValues: []uint64{
			0x0a0a0a0a, // Reserved for greasing
			0x1a1a1a1a,
			0x2a2a2a2a,
			0x3a3a3a3a,
			0x4a4a4a4a,
			0x5a5a5a5a,
			0x6a6a6a6a,
			0x7a7a7a7a,
			0x8a8a8a8a,
			0x9a9a9a9a,
			0xaaaaaaaa,
			0xbabababa,
			0xcacacaca,
			0xdadadada,
			0xeaeaeaea,
			0xfafafafa,
		},
	}
}

// ShouldGrease determines if greasing should be applied
// Uses probability-based decision
func (gm *GreasingManager) ShouldGrease() bool {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	if !gm.isActive {
		return false
	}

	// Generate random number between 0 and 1
	randomValue, err := rand.Int(rand.Reader, big.NewInt(1000))
	if err != nil {
		return false
	}

	probability := float64(randomValue.Int64()) / 1000.0
	return probability < gm.config.GreaseProbability
}

// GetGreaseValue returns a random grease value from the reserved list
func (gm *GreasingManager) GetGreaseValue() (uint64, error) {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	if len(gm.config.GreaseValues) == 0 {
		return 0, fmt.Errorf("no grease values configured")
	}

	// Select random grease value
	idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(gm.config.GreaseValues))))
	if err != nil {
		return 0, fmt.Errorf("failed to generate random index: %w", err)
	}

	return gm.config.GreaseValues[idx.Int64()], nil
}

// GreaseVersionNegotiation applies greasing to QUIC version negotiation
// Adds reserved version numbers to prevent middlebox interference
func (gm *GreasingManager) GreaseVersionNegotiation(versions []uint32) ([]uint32, error) {
	if !gm.config.EnableVersionGreasing {
		return versions, nil
	}

	if !gm.ShouldGrease() {
		return versions, nil
	}

	greaseValue, err := gm.GetGreaseValue()
	if err != nil {
		return versions, err
	}

	// Insert grease value at random position
	idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(versions)+1)))
	if err != nil {
		return versions, err
	}

	insertIdx := int(idx.Int64())
	greasedVersions := make([]uint32, 0, len(versions)+1)
	greasedVersions = append(greasedVersions, versions[:insertIdx]...)
	greasedVersions = append(greasedVersions, uint32(greaseValue))
	greasedVersions = append(greasedVersions, versions[insertIdx:]...)

	gm.mu.Lock()
	gm.metrics.VersionGreaseCount++
	gm.metrics.TotalGreaseApplications++
	gm.mu.Unlock()

	gm.logger.Debug("Applied version greasing",
		zap.Uint64("grease_value", greaseValue),
		zap.Int("position", insertIdx),
		zap.Int("total_versions", len(greasedVersions)))

	return greasedVersions, nil
}

// GreaseTransportParameters applies greasing to transport parameters
// Adds reserved parameter types to test peer implementation
func (gm *GreasingManager) GreaseTransportParameters(params map[uint64][]byte) (map[uint64][]byte, error) {
	if !gm.config.EnableTransportParamGreasing {
		return params, nil
	}

	if !gm.ShouldGrease() {
		return params, nil
	}

	greaseValue, err := gm.GetGreaseValue()
	if err != nil {
		return params, err
	}

	// Add grease parameter with random data
	greaseData := make([]byte, 8)
	_, err = rand.Read(greaseData)
	if err != nil {
		return params, err
	}

	greasedParams := make(map[uint64][]byte)
	for k, v := range params {
		greasedParams[k] = v
	}
	greasedParams[greaseValue] = greaseData

	gm.mu.Lock()
	gm.metrics.TransportParamGreaseCount++
	gm.metrics.TotalGreaseApplications++
	gm.mu.Unlock()

	gm.logger.Debug("Applied transport parameter greasing",
		zap.Uint64("grease_param", greaseValue),
		zap.Int("param_count", len(greasedParams)))

	return greasedParams, nil
}

// GreaseFrameType applies greasing to frame types
// Sends reserved frame types to test peer's frame handling
func (gm *GreasingManager) GreaseFrameType() (uint64, bool, error) {
	if !gm.config.EnableFrameTypeGreasing {
		return 0, false, nil
	}

	if !gm.ShouldGrease() {
		return 0, false, nil
	}

	greaseValue, err := gm.GetGreaseValue()
	if err != nil {
		return 0, false, err
	}

	gm.mu.Lock()
	gm.metrics.FrameTypeGreaseCount++
	gm.metrics.TotalGreaseApplications++
	gm.mu.Unlock()

	gm.logger.Debug("Applied frame type greasing",
		zap.Uint64("grease_frame_type", greaseValue))

	return greaseValue, true, nil
}

// ValidateGreaseValue checks if a value is a valid grease value
// Peers MUST ignore grease values according to RFC 9287
func (gm *GreasingManager) ValidateGreaseValue(value uint64) bool {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	for _, grease := range gm.config.GreaseValues {
		if value == grease {
			return true
		}
	}
	return false
}

// IsGreaseValue checks if a received value is a grease value
// Used by peers to properly ignore grease values
func (gm *GreasingManager) IsGreaseValue(value uint64) bool {
	return gm.ValidateGreaseValue(value)
}

// HandleCompatibilityIssue records a compatibility issue
// Called when a peer incorrectly handles grease values
func (gm *GreasingManager) HandleCompatibilityIssue(description string) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	gm.metrics.CompatibilityIssues++

	gm.logger.Warn("QUIC greasing compatibility issue detected",
		zap.String("description", description),
		zap.Int64("total_issues", gm.metrics.CompatibilityIssues))
}

// GetMetrics returns greasing metrics
func (gm *GreasingManager) GetMetrics() *GreasingMetrics {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	// Return a copy
	metricsCopy := *gm.metrics
	return &metricsCopy
}

// SetGreaseProbability adjusts the probability of applying greasing
func (gm *GreasingManager) SetGreaseProbability(probability float64) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	if probability < 0 {
		probability = 0
	}
	if probability > 1.0 {
		probability = 1.0
	}

	gm.config.GreaseProbability = probability

	gm.logger.Info("Grease probability adjusted",
		zap.Float64("new_probability", probability))
}

// EnableFeature enables specific greasing feature
func (gm *GreasingManager) EnableFeature(feature string) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	switch feature {
	case "version":
		gm.config.EnableVersionGreasing = true
	case "transport_param":
		gm.config.EnableTransportParamGreasing = true
	case "frame_type":
		gm.config.EnableFrameTypeGreasing = true
	}

	gm.logger.Info("Greasing feature enabled",
		zap.String("feature", feature))
}

// DisableFeature disables specific greasing feature
func (gm *GreasingManager) DisableFeature(feature string) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	switch feature {
	case "version":
		gm.config.EnableVersionGreasing = false
	case "transport_param":
		gm.config.EnableTransportParamGreasing = false
	case "frame_type":
		gm.config.EnableFrameTypeGreasing = false
	}

	gm.logger.Info("Greasing feature disabled",
		zap.String("feature", feature))
}

// Stop stops the greasing manager
func (gm *GreasingManager) Stop() {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	gm.isActive = false

	gm.logger.Info("Greasing manager stopped",
		zap.Int64("total_applications", gm.metrics.TotalGreaseApplications),
		zap.Int64("version_grease", gm.metrics.VersionGreaseCount),
		zap.Int64("transport_param_grease", gm.metrics.TransportParamGreaseCount),
		zap.Int64("frame_type_grease", gm.metrics.FrameTypeGreaseCount),
		zap.Int64("compatibility_issues", gm.metrics.CompatibilityIssues))
}

// IsActive returns whether the manager is active
func (gm *GreasingManager) IsActive() bool {
	gm.mu.RLock()
	defer gm.mu.RUnlock()
	return gm.isActive
}

// GetGreasingStatus returns current greasing configuration status
func (gm *GreasingManager) GetGreasingStatus() map[string]interface{} {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	return map[string]interface{}{
		"active":                      gm.isActive,
		"version_greasing":            gm.config.EnableVersionGreasing,
		"transport_param_greasing":    gm.config.EnableTransportParamGreasing,
		"frame_type_greasing":         gm.config.EnableFrameTypeGreasing,
		"grease_probability":          gm.config.GreaseProbability,
		"total_grease_applications":   gm.metrics.TotalGreaseApplications,
		"compatibility_issues":        gm.metrics.CompatibilityIssues,
	}
}
