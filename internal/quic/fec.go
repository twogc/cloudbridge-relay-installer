package quic

import (
	"crypto/rand"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// FECManager manages Forward Error Correction for QUIC datagrams
// Implements FEC for unreliable datagrams to reduce retransmissions
// Integrated from quic-test experimental features
type FECManager struct {
	logger      *zap.Logger
	config      *FECConfig
	mu          sync.RWMutex
	groups      map[string]*FECGroup
	metrics     *FECMetrics
	isActive    bool
	nextGroupID uint64
}

// FECConfig holds FEC configuration
type FECConfig struct {
	// Redundancy ratio (0.1 = 10% redundant packets)
	RedundancyRatio float64

	// FEC algorithm: "xor" or "reed-solomon"
	Algorithm string

	// Maximum packets per FEC group
	GroupSize int

	// Timeout for completing FEC group
	GroupTimeout time.Duration

	// Enable adaptive redundancy based on loss rate
	EnableAdaptive bool

	// Target recovery rate (e.g., 0.9 = 90% recovery)
	TargetRecoveryRate float64
}

// FECGroup represents a group of packets with FEC protection
type FECGroup struct {
	ID              uint64
	DataPackets     [][]byte
	RedundantPackets [][]byte
	ReceivedPackets  map[int]bool
	LostPackets      []int
	CreatedAt        time.Time
	CompletedAt      time.Time
	Algorithm        string
	mu               sync.RWMutex
}

// FECMetrics holds FEC performance metrics
type FECMetrics struct {
	TotalGroups       int64   `json:"total_groups"`
	ActiveGroups      int64   `json:"active_groups"`
	RedundancyBytes   int64   `json:"redundancy_bytes"`
	RecoveryEvents    int64   `json:"recovery_events"`
	FailedRecoveries  int64   `json:"failed_recoveries"`
	RecoveryRate      float64 `json:"recovery_rate"`
	Efficiency        float64 `json:"efficiency"`
	AverageLatency    time.Duration `json:"average_latency"`
	BytesSaved        int64   `json:"bytes_saved"` // Bytes saved by avoiding retransmission
}

// NewFECManager creates a new FEC manager
func NewFECManager(logger *zap.Logger, config *FECConfig) *FECManager {
	if config == nil {
		config = DefaultFECConfig()
	}

	return &FECManager{
		logger:   logger,
		config:   config,
		groups:   make(map[string]*FECGroup),
		metrics:  &FECMetrics{},
		isActive: true,
	}
}

// DefaultFECConfig returns default FEC configuration
func DefaultFECConfig() *FECConfig {
	return &FECConfig{
		RedundancyRatio:    0.1,         // 10% redundancy
		Algorithm:          "xor",       // XOR is simpler and faster
		GroupSize:          10,          // 10 packets per group
		GroupTimeout:       100 * time.Millisecond,
		EnableAdaptive:     true,
		TargetRecoveryRate: 0.9,         // 90% recovery target
	}
}

// CreateGroup creates a new FEC group
func (fm *FECManager) CreateGroup(dataPackets [][]byte) (*FECGroup, error) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if !fm.isActive {
		return nil, fmt.Errorf("FEC manager is not active")
	}

	fm.nextGroupID++
	groupID := fm.nextGroupID

	// Calculate number of redundant packets
	numRedundant := int(float64(len(dataPackets)) * fm.config.RedundancyRatio)
	if numRedundant < 1 {
		numRedundant = 1 // At least one redundant packet
	}

	group := &FECGroup{
		ID:               groupID,
		DataPackets:      make([][]byte, len(dataPackets)),
		RedundantPackets: make([][]byte, numRedundant),
		ReceivedPackets:  make(map[int]bool),
		LostPackets:      []int{},
		CreatedAt:        time.Now(),
		Algorithm:        fm.config.Algorithm,
	}

	// Copy data packets
	for i, pkt := range dataPackets {
		group.DataPackets[i] = make([]byte, len(pkt))
		copy(group.DataPackets[i], pkt)
		group.ReceivedPackets[i] = true // Initially all received
	}

	// Generate redundant packets based on algorithm
	var err error
	switch fm.config.Algorithm {
	case "xor":
		group.RedundantPackets, err = fm.generateXORRedundancy(group.DataPackets, numRedundant)
	case "reed-solomon":
		group.RedundantPackets, err = fm.generateReedSolomonRedundancy(group.DataPackets, numRedundant)
	default:
		return nil, fmt.Errorf("unsupported FEC algorithm: %s", fm.config.Algorithm)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to generate FEC redundancy: %w", err)
	}

	// Store group
	groupKey := fmt.Sprintf("%d", groupID)
	fm.groups[groupKey] = group

	// Update metrics
	fm.metrics.TotalGroups++
	fm.metrics.ActiveGroups++
	redundancyBytes := int64(0)
	for _, pkt := range group.RedundantPackets {
		redundancyBytes += int64(len(pkt))
	}
	fm.metrics.RedundancyBytes += redundancyBytes

	fm.logger.Debug("FEC group created",
		zap.Uint64("group_id", groupID),
		zap.Int("data_packets", len(dataPackets)),
		zap.Int("redundant_packets", numRedundant),
		zap.String("algorithm", fm.config.Algorithm))

	return group, nil
}

// generateXORRedundancy generates XOR-based redundant packets
// Simple and fast: XOR all data packets together
func (fm *FECManager) generateXORRedundancy(dataPackets [][]byte, numRedundant int) ([][]byte, error) {
	if len(dataPackets) == 0 {
		return nil, fmt.Errorf("no data packets provided")
	}

	// Find maximum packet size
	maxSize := 0
	for _, pkt := range dataPackets {
		if len(pkt) > maxSize {
			maxSize = len(pkt)
		}
	}

	redundant := make([][]byte, numRedundant)

	// For simplicity, create one XOR packet of all data packets
	// In production, use more sophisticated schemes
	if numRedundant > 0 {
		xorPacket := make([]byte, maxSize)

		// XOR all data packets
		for _, pkt := range dataPackets {
			for i := 0; i < len(pkt); i++ {
				xorPacket[i] ^= pkt[i]
			}
		}

		redundant[0] = xorPacket

		// For additional redundant packets, add random XOR combinations
		for i := 1; i < numRedundant; i++ {
			redundant[i] = make([]byte, maxSize)
			// Simple approach: copy and add randomness
			copy(redundant[i], xorPacket)
			// Add some entropy for diversity
			randBytes := make([]byte, 32)
			_, _ = rand.Read(randBytes)
			for j := 0; j < len(randBytes) && j < maxSize; j++ {
				redundant[i][j] ^= randBytes[j]
			}
		}
	}

	return redundant, nil
}

// generateReedSolomonRedundancy generates Reed-Solomon redundant packets
// More complex but provides better recovery with higher overhead
func (fm *FECManager) generateReedSolomonRedundancy(dataPackets [][]byte, numRedundant int) ([][]byte, error) {
	// TODO: Implement Reed-Solomon encoding
	// For now, fall back to XOR
	fm.logger.Warn("Reed-Solomon not implemented, falling back to XOR")
	return fm.generateXORRedundancy(dataPackets, numRedundant)
}

// RecoverLostPackets attempts to recover lost packets using FEC
func (fm *FECManager) RecoverLostPackets(groupID uint64, receivedIndices []int, redundantPackets [][]byte) ([][]byte, error) {
	fm.mu.RLock()
	groupKey := fmt.Sprintf("%d", groupID)
	group, exists := fm.groups[groupKey]
	fm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("FEC group %d not found", groupID)
	}

	group.mu.Lock()
	defer group.mu.Unlock()

	// Determine lost packets
	allIndices := make(map[int]bool)
	for i := 0; i < len(group.DataPackets); i++ {
		allIndices[i] = true
	}
	for _, idx := range receivedIndices {
		delete(allIndices, idx)
	}

	lostIndices := []int{}
	for idx := range allIndices {
		lostIndices = append(lostIndices, idx)
	}

	if len(lostIndices) == 0 {
		// No packets lost
		return nil, nil
	}

	fm.logger.Debug("Attempting FEC recovery",
		zap.Uint64("group_id", groupID),
		zap.Ints("lost_indices", lostIndices),
		zap.Int("redundant_available", len(redundantPackets)))

	// Attempt recovery based on algorithm
	var recovered [][]byte
	var err error

	switch group.Algorithm {
	case "xor":
		recovered, err = fm.recoverXOR(group, lostIndices, receivedIndices, redundantPackets)
	case "reed-solomon":
		recovered, err = fm.recoverReedSolomon(group, lostIndices, receivedIndices, redundantPackets)
	default:
		return nil, fmt.Errorf("unsupported FEC algorithm: %s", group.Algorithm)
	}

	// Update metrics
	fm.mu.Lock()
	if err == nil && len(recovered) > 0 {
		fm.metrics.RecoveryEvents++
		// Calculate bytes saved by avoiding retransmission
		for _, pkt := range recovered {
			fm.metrics.BytesSaved += int64(len(pkt))
		}
	} else {
		fm.metrics.FailedRecoveries++
	}

	// Update recovery rate
	if fm.metrics.RecoveryEvents+fm.metrics.FailedRecoveries > 0 {
		fm.metrics.RecoveryRate = float64(fm.metrics.RecoveryEvents) /
			float64(fm.metrics.RecoveryEvents+fm.metrics.FailedRecoveries)
	}
	fm.mu.Unlock()

	return recovered, err
}

// recoverXOR recovers lost packets using XOR redundancy
func (fm *FECManager) recoverXOR(group *FECGroup, lostIndices, receivedIndices []int, redundantPackets [][]byte) ([][]byte, error) {
	if len(lostIndices) > len(redundantPackets) {
		return nil, fmt.Errorf("too many lost packets (%d) for available redundancy (%d)",
			len(lostIndices), len(redundantPackets))
	}

	// Simple XOR recovery: can recover 1 lost packet if we have the XOR of all packets
	if len(lostIndices) == 1 && len(redundantPackets) > 0 {
		lostIdx := lostIndices[0]
		maxSize := len(redundantPackets[0])

		recovered := make([]byte, maxSize)
		copy(recovered, redundantPackets[0])

		// XOR with all received packets
		for _, idx := range receivedIndices {
			if idx < len(group.DataPackets) {
				pkt := group.DataPackets[idx]
				for i := 0; i < len(pkt) && i < maxSize; i++ {
					recovered[i] ^= pkt[i]
				}
			}
		}

		fm.logger.Info("Successfully recovered packet using XOR FEC",
			zap.Uint64("group_id", group.ID),
			zap.Int("recovered_index", lostIdx))

		return [][]byte{recovered}, nil
	}

	return nil, fmt.Errorf("XOR FEC can only recover 1 lost packet, got %d lost", len(lostIndices))
}

// recoverReedSolomon recovers lost packets using Reed-Solomon redundancy
func (fm *FECManager) recoverReedSolomon(group *FECGroup, lostIndices, receivedIndices []int, redundantPackets [][]byte) ([][]byte, error) {
	// TODO: Implement Reed-Solomon decoding
	fm.logger.Warn("Reed-Solomon recovery not implemented")
	return nil, fmt.Errorf("Reed-Solomon recovery not implemented")
}

// GetMetrics returns FEC metrics
func (fm *FECManager) GetMetrics() *FECMetrics {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	// Calculate efficiency (recovery rate vs overhead)
	efficiency := 0.0
	if fm.metrics.RedundancyBytes > 0 {
		efficiency = float64(fm.metrics.BytesSaved) / float64(fm.metrics.RedundancyBytes)
	}
	fm.metrics.Efficiency = efficiency

	// Return a copy
	metricsCopy := *fm.metrics
	return &metricsCopy
}

// CleanupOldGroups removes old FEC groups to free memory
func (fm *FECManager) CleanupOldGroups(maxAge time.Duration) int {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	removed := 0

	for key, group := range fm.groups {
		if group.CreatedAt.Before(cutoff) {
			delete(fm.groups, key)
			removed++
			fm.metrics.ActiveGroups--
		}
	}

	if removed > 0 {
		fm.logger.Debug("Cleaned up old FEC groups",
			zap.Int("removed", removed),
			zap.Int64("active", fm.metrics.ActiveGroups))
	}

	return removed
}

// SetRedundancy dynamically adjusts redundancy ratio
func (fm *FECManager) SetRedundancy(ratio float64) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1.0 {
		ratio = 1.0
	}

	fm.config.RedundancyRatio = ratio
	fm.logger.Info("FEC redundancy adjusted",
		zap.Float64("new_ratio", ratio))
}

// AdaptiveAdjustment adjusts redundancy based on observed loss rate
func (fm *FECManager) AdaptiveAdjustment(observedLossRate float64) {
	if !fm.config.EnableAdaptive {
		return
	}

	fm.mu.Lock()
	defer fm.mu.Unlock()

	// If loss rate is high, increase redundancy
	if observedLossRate > 0.05 { // > 5% loss
		newRatio := fm.config.RedundancyRatio * 1.2
		if newRatio > 0.5 { // Cap at 50%
			newRatio = 0.5
		}
		fm.config.RedundancyRatio = newRatio

		fm.logger.Info("Increased FEC redundancy due to high loss rate",
			zap.Float64("loss_rate", observedLossRate),
			zap.Float64("new_redundancy", newRatio))
	} else if observedLossRate < 0.01 { // < 1% loss
		// If loss rate is low, decrease redundancy
		newRatio := fm.config.RedundancyRatio * 0.9
		if newRatio < 0.05 { // Minimum 5%
			newRatio = 0.05
		}
		fm.config.RedundancyRatio = newRatio

		fm.logger.Debug("Decreased FEC redundancy due to low loss rate",
			zap.Float64("loss_rate", observedLossRate),
			zap.Float64("new_redundancy", newRatio))
	}
}

// Stop stops the FEC manager
func (fm *FECManager) Stop() {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	fm.isActive = false
	fm.logger.Info("FEC manager stopped",
		zap.Int64("total_groups", fm.metrics.TotalGroups),
		zap.Int64("recovery_events", fm.metrics.RecoveryEvents),
		zap.Float64("recovery_rate", fm.metrics.RecoveryRate))
}

// IsActive returns whether the manager is active
func (fm *FECManager) IsActive() bool {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	return fm.isActive
}
