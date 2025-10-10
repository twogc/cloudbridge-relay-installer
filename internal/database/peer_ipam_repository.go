package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// PeerIPAMRepository handles per-peer IP address management
type PeerIPAMRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewPeerIPAMRepository creates a new peer IPAM repository
func NewPeerIPAMRepository(db *sql.DB, logger *zap.Logger) *PeerIPAMRepository {
	return &PeerIPAMRepository{
		db:     db,
		logger: logger,
	}
}

// AllocateIPForPeer allocates an IP address for a peer within a tenant's subnet
// Uses transactions and proper subnet handling to avoid races
func (r *PeerIPAMRepository) AllocateIPForPeer(tenantID, peerID, tenantCIDR string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Start transaction with serializable isolation to prevent races
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return "", fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// 1) Check if peer already has an active IP allocation (reuse)
	var existingIP string
	err = tx.QueryRowContext(ctx, `
		SELECT allocated_ip FROM peer_ip_allocations
		WHERE peer_id = $1 AND is_active = true
		LIMIT 1`, peerID).Scan(&existingIP)
	if err == nil {
		// Peer already has an active allocation, return it
		_ = tx.Commit()
		if r.logger != nil {
			r.logger.Info("Reusing existing IP allocation for peer",
				zap.String("tenant_id", tenantID),
				zap.String("peer_id", peerID),
				zap.String("allocated_ip", existingIP))
		}
		return existingIP, nil
	}
	if err != sql.ErrNoRows {
		return "", fmt.Errorf("failed to check existing allocation: %w", err)
	}

	// 2) Parse subnet range properly
	subnetRange, err := parseSubnetRange(tenantCIDR)
	if err != nil {
		return "", fmt.Errorf("failed to parse subnet range: %w", err)
	}

	// 3) Get all currently allocated IPs in this tenant (single query)
	rows, err := tx.QueryContext(ctx, `
		SELECT allocated_ip FROM peer_ip_allocations
		WHERE tenant_id = $1 AND is_active = true`, tenantID)
	if err != nil {
		return "", fmt.Errorf("failed to get allocated IPs: %w", err)
	}
	defer rows.Close()

	usedIPs := make(map[string]struct{})
	for rows.Next() {
		var ip string
		if err := rows.Scan(&ip); err != nil {
			return "", fmt.Errorf("failed to scan allocated IP: %w", err)
		}
		usedIPs[ip] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("failed to iterate allocated IPs: %w", err)
	}

	// 4) Try to allocate IP starting from first usable address
	now := time.Now()
	allocationID := uuid.New()

	// Start from first usable IP and iterate through the range
	currentIP := subnetRange.FirstIP
	for currentIP <= subnetRange.LastIP {
		ipStr := uint32ToIPv4(currentIP).String()

		// Skip reserved IPs
		if isReservedIP(ipStr) {
			currentIP++
			continue
		}

		// Skip already allocated IPs
		if _, used := usedIPs[ipStr]; used {
			currentIP++
			continue
		}

		// Try to insert this IP allocation
		_, err = tx.ExecContext(ctx, `
			INSERT INTO peer_ip_allocations 
			(id, tenant_id, peer_id, allocated_ip, allocated_at, is_active, version, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, true, 1, $5, $5)`,
			allocationID, tenantID, peerID, ipStr, now)

		if err == nil {
			// Success! Commit transaction
			if err = tx.Commit(); err != nil {
				return "", fmt.Errorf("failed to commit transaction: %w", err)
			}

			if r.logger != nil {
				r.logger.Info("Successfully allocated IP for peer",
					zap.String("tenant_id", tenantID),
					zap.String("peer_id", peerID),
					zap.String("allocated_ip", ipStr),
					zap.String("tenant_cidr", tenantCIDR))
			}
			return ipStr, nil
		}

		// If it's a unique constraint violation, try next IP
		if isUniqueViolation(err) {
			currentIP++
			continue
		}

		// Other error - fail
		return "", fmt.Errorf("failed to allocate IP %s: %w", ipStr, err)
	}

	// No available IPs found
	return "", fmt.Errorf("no available IP addresses in tenant subnet %s (range: %d-%d)",
		tenantCIDR, subnetRange.FirstIP, subnetRange.LastIP)
}

// DeallocateIPForPeer releases an IP address for a peer
// Uses cross-database compatible approach without NOW() and RETURNING
func (r *PeerIPAMRepository) DeallocateIPForPeer(peerID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// First, get information about the allocation before deallocating
	var allocatedIP, tenantID string
	err := r.db.QueryRowContext(ctx, `
		SELECT allocated_ip, tenant_id
		FROM peer_ip_allocations
		WHERE peer_id = $1 AND is_active = true`, peerID).Scan(&allocatedIP, &tenantID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("no active IP allocation found for peer %s", peerID)
		}
		return fmt.Errorf("failed to get allocation info: %w", err)
	}

	// Now deallocate the IP
	now := time.Now()
	res, err := r.db.ExecContext(ctx, `
		UPDATE peer_ip_allocations
		SET is_active = false, released_at = $2, updated_at = $2
		WHERE peer_id = $1 AND is_active = true`,
		peerID, now)
	if err != nil {
		return fmt.Errorf("failed to deallocate IP: %w", err)
	}

	// Check if any rows were affected
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no active IP allocation found for peer %s", peerID)
	}

	if r.logger != nil {
		r.logger.Info("Successfully deallocated IP for peer",
			zap.String("peer_id", peerID),
			zap.String("tenant_id", tenantID),
			zap.String("allocated_ip", allocatedIP))
	}

	return nil
}

// GetPeerIP returns the allocated IP for a peer
func (r *PeerIPAMRepository) GetPeerIP(peerID string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT allocated_ip 
		FROM peer_ip_allocations 
		WHERE peer_id = $1 AND is_active = true`

	var allocatedIP string
	err := r.db.QueryRowContext(ctx, query, peerID).Scan(&allocatedIP)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("no IP allocation found for peer %s", peerID)
		}
		return "", fmt.Errorf("failed to get peer IP: %w", err)
	}

	return allocatedIP, nil
}

// GetTenantPeerIPs returns all allocated IPs for a tenant
func (r *PeerIPAMRepository) GetTenantPeerIPs(tenantID string) (map[string]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT peer_id, allocated_ip 
		FROM peer_ip_allocations 
		WHERE tenant_id = $1 AND is_active = true`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant peer IPs: %w", err)
	}
	defer rows.Close()

	peerIPs := make(map[string]string)
	for rows.Next() {
		var peerID, allocatedIP string
		if err := rows.Scan(&peerID, &allocatedIP); err != nil {
			return nil, fmt.Errorf("failed to scan peer IP: %w", err)
		}
		peerIPs[peerID] = allocatedIP
	}

	return peerIPs, nil
}

// IsIPAllocatedInTenant checks if an IP is already allocated in a tenant
func (r *PeerIPAMRepository) IsIPAllocatedInTenant(tenantID, ip string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT COUNT(*) 
		FROM peer_ip_allocations 
		WHERE tenant_id = $1 AND allocated_ip = $2 AND is_active = true`

	var count int
	err := r.db.QueryRowContext(ctx, query, tenantID, ip).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check IP allocation: %w", err)
	}

	return count > 0, nil
}

// CreatePeerIPAllocation creates a new peer IP allocation record
func (r *PeerIPAMRepository) CreatePeerIPAllocation(allocation *PeerIPAllocation) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		INSERT INTO peer_ip_allocations (
			id, tenant_id, peer_id, allocated_ip, allocated_at, is_active, version, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)`

	_, err := r.db.ExecContext(ctx, query,
		allocation.ID,
		allocation.TenantID,
		allocation.PeerID,
		allocation.AllocatedIP,
		allocation.AllocatedAt,
		allocation.IsActive,
		allocation.Version,
		allocation.CreatedAt,
		allocation.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create peer IP allocation: %w", err)
	}

	return nil
}

// GetPeerIPAllocation retrieves a peer IP allocation by peer ID
func (r *PeerIPAMRepository) GetPeerIPAllocation(peerID string) (*PeerIPAllocation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, tenant_id, peer_id, allocated_ip, allocated_at, released_at, is_active, version, created_at, updated_at
		FROM peer_ip_allocations 
		WHERE peer_id = $1 AND is_active = true`

	var allocation PeerIPAllocation
	err := r.db.QueryRowContext(ctx, query, peerID).Scan(
		&allocation.ID,
		&allocation.TenantID,
		&allocation.PeerID,
		&allocation.AllocatedIP,
		&allocation.AllocatedAt,
		&allocation.ReleasedAt,
		&allocation.IsActive,
		&allocation.Version,
		&allocation.CreatedAt,
		&allocation.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no IP allocation found for peer %s", peerID)
		}
		return nil, fmt.Errorf("failed to get peer IP allocation: %w", err)
	}

	return &allocation, nil
}

// GetTenantIPAllocationStats returns statistics about IP allocations for a tenant
func (r *PeerIPAMRepository) GetTenantIPAllocationStats(tenantID string) (map[string]interface{}, error) {
	query := `
		SELECT 
			COUNT(*) as total_allocations,
			COUNT(CASE WHEN is_active = true THEN 1 END) as active_allocations,
			COUNT(CASE WHEN is_active = false THEN 1 END) as released_allocations
		FROM peer_ip_allocations 
		WHERE tenant_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var totalAllocations, activeAllocations, releasedAllocations int
	err := r.db.QueryRowContext(ctx, query, tenantID).Scan(&totalAllocations, &activeAllocations, &releasedAllocations)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant IP allocation stats: %w", err)
	}

	stats := map[string]interface{}{
		"tenant_id":            tenantID,
		"total_allocations":    totalAllocations,
		"active_allocations":   activeAllocations,
		"released_allocations": releasedAllocations,
	}

	return stats, nil
}

// DeallocateIPsByTenant деактивирует все IP адреса пиров тенанта
func (r *PeerIPAMRepository) DeallocateIPsByTenant(ctx context.Context, tenantID string) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Используем CURRENT_TIMESTAMP для кросс-БД совместимости
	query := `
		UPDATE peer_ip_allocations
		SET is_active = false,
		    released_at = CURRENT_TIMESTAMP,
		    updated_at = CURRENT_TIMESTAMP
		WHERE tenant_id = $1 AND is_active = true
	`

	result, err := r.db.ExecContext(ctx, query, tenantID)
	if err != nil {
		return 0, fmt.Errorf("failed to deallocate IPs for tenant: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	r.logger.Info("Deallocated IPs for tenant",
		zap.String("tenant_id", tenantID),
		zap.Int64("count", rowsAffected))

	return rowsAffected, nil
}
