package database

import (
	"fmt"
	"time"

	"go.uber.org/zap"
)

// IPAMRepository extends Repository with IPAM-specific methods
type IPAMRepository struct {
	*Repository
}

// NewIPAMRepository creates a new IPAM repository
func NewIPAMRepository(repo *Repository) *IPAMRepository {
	return &IPAMRepository{Repository: repo}
}

// Tenant represents a tenant in the database
type Tenant struct {
	ID          int       `db:"id" json:"id"`
	TenantID    string    `db:"tenant_id" json:"tenant_id"`
	Description string    `db:"description" json:"description"`
	MaxSubnets  int       `db:"max_subnets" json:"max_subnets"`
	Status      string    `db:"status" json:"status"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

// IPAMStats represents IPAM statistics
type IPAMStats struct {
	TotalTenants     int     `db:"total_tenants" json:"total_tenants"`
	ActiveTenants    int     `db:"active_tenants" json:"active_tenants"`
	TotalSubnets     int     `db:"total_subnets" json:"total_subnets"`
	AllocatedSubnets int     `db:"allocated_subnets" json:"allocated_subnets"`
	AvailableSubnets int     `db:"available_subnets" json:"available_subnets"`
	UtilizationRate  float64 `db:"utilization_rate" json:"utilization_rate"`
}

// CreateTenant creates a new tenant
func (r *IPAMRepository) CreateTenant(tenant *Tenant) error {
	query := `
		INSERT INTO tenants (
			tenant_id, description, max_subnets, status, created_at, updated_at
		) VALUES (
			:tenant_id, :description, :max_subnets, :status, :created_at, :updated_at
		) RETURNING id`

	rows, err := r.db.NamedQuery(query, tenant)
	if err != nil {
		return fmt.Errorf("failed to create tenant: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&tenant.ID)
		if err != nil {
			return fmt.Errorf("failed to scan tenant ID: %w", err)
		}
	}

	r.logger.Info("Tenant created", 
		zap.String("tenant_id", tenant.TenantID),
		zap.Int("id", tenant.ID))

	return nil
}

// GetTenant retrieves a tenant by ID
func (r *IPAMRepository) GetTenant(tenantID string) (*Tenant, error) {
	query := `SELECT * FROM tenants WHERE tenant_id = $1`
	
	var tenant Tenant
	err := r.db.Get(&tenant, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	return &tenant, nil
}

// DeleteTenant deletes a tenant
func (r *IPAMRepository) DeleteTenant(tenantID string) error {
	query := `DELETE FROM tenants WHERE tenant_id = $1`
	
	result, err := r.db.Exec(query, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete tenant: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("tenant not found")
	}

	r.logger.Info("Tenant deleted", zap.String("tenant_id", tenantID))
	return nil
}

// ListTenants lists tenants with pagination
func (r *IPAMRepository) ListTenants(limit, offset int, status string) ([]*Tenant, error) {
	query := `SELECT * FROM tenants`
	args := []interface{}{}
	argIndex := 1

	if status != "" {
		query += ` WHERE status = $` + fmt.Sprintf("%d", argIndex)
		args = append(args, status)
		argIndex++
	}

	query += ` ORDER BY created_at DESC LIMIT $` + fmt.Sprintf("%d", argIndex) + ` OFFSET $` + fmt.Sprintf("%d", argIndex+1)
	args = append(args, limit, offset)

	var tenants []*Tenant
	err := r.db.Select(&tenants, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list tenants: %w", err)
	}

	return tenants, nil
}

// GetTenantSubnetCount returns the number of active subnets for a tenant
func (r *IPAMRepository) GetTenantSubnetCount(tenantID string) (int, error) {
	query := `SELECT COUNT(*) FROM tenant_networks WHERE tenant_id = $1 AND is_active = true`
	
	var count int
	err := r.db.Get(&count, query, tenantID)
	if err != nil {
		return 0, fmt.Errorf("failed to get tenant subnet count: %w", err)
	}

	return count, nil
}

// GetTenantSubnets returns all subnets for a tenant
func (r *IPAMRepository) GetTenantSubnets(tenantID string) ([]*TenantNetwork, error) {
	query := `SELECT * FROM tenant_networks WHERE tenant_id = $1 ORDER BY created_at DESC`
	
	var subnets []*TenantNetwork
	err := r.db.Select(&subnets, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant subnets: %w", err)
	}

	return subnets, nil
}

// DeactivateAllTenantSubnets deactivates all subnets for a tenant
func (r *IPAMRepository) DeactivateAllTenantSubnets(tenantID string) error {
	query := `UPDATE tenant_networks SET is_active = false, updated_at = $1 WHERE tenant_id = $2`
	
	_, err := r.db.Exec(query, time.Now(), tenantID)
	if err != nil {
		return fmt.Errorf("failed to deactivate tenant subnets: %w", err)
	}

	r.logger.Info("All tenant subnets deactivated", zap.String("tenant_id", tenantID))
	return nil
}

// DeactivateTenantSubnet deactivates a specific subnet for a tenant
func (r *IPAMRepository) DeactivateTenantSubnet(tenantID, subnetCIDR string) error {
	query := `UPDATE tenant_networks SET is_active = false, updated_at = $1 WHERE tenant_id = $2 AND network_cidr = $3`
	
	result, err := r.db.Exec(query, time.Now(), tenantID, subnetCIDR)
	if err != nil {
		return fmt.Errorf("failed to deactivate tenant subnet: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("subnet not found")
	}

	r.logger.Info("Tenant subnet deactivated", 
		zap.String("tenant_id", tenantID),
		zap.String("subnet_cidr", subnetCIDR))
	return nil
}

// GetNextSubnetCounter returns the next available subnet counter
func (r *IPAMRepository) GetNextSubnetCounter() (int, error) {
	query := `SELECT COALESCE(MAX(subnet_counter), 0) + 1 FROM tenant_networks`
	
	var counter int
	err := r.db.Get(&counter, query)
	if err != nil {
		return 0, fmt.Errorf("failed to get next subnet counter: %w", err)
	}

	return counter, nil
}

// CreateTenantNetwork creates a new tenant network
func (r *IPAMRepository) CreateTenantNetwork(network *TenantNetwork) error {
	query := `
		INSERT INTO tenant_networks (
			tenant_id, network_cidr, subnet_counter, is_active, created_at, updated_at
		) VALUES (
			:tenant_id, :network_cidr, :subnet_counter, :is_active, :created_at, :updated_at
		) RETURNING id`

	rows, err := r.db.NamedQuery(query, network)
	if err != nil {
		return fmt.Errorf("failed to create tenant network: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&network.ID)
		if err != nil {
			return fmt.Errorf("failed to scan network ID: %w", err)
		}
	}

	r.logger.Info("Tenant network created", 
		zap.String("tenant_id", network.TenantID),
		zap.String("network_cidr", network.NetworkCIDR),
		zap.Int("subnet_counter", network.SubnetCounter))

	return nil
}

// GetIPAMStats returns IPAM statistics
func (r *IPAMRepository) GetIPAMStats() (*IPAMStats, error) {
	query := `
		SELECT 
			(SELECT COUNT(*) FROM tenants) as total_tenants,
			(SELECT COUNT(*) FROM tenants WHERE status = 'active') as active_tenants,
			(SELECT COUNT(*) FROM tenant_networks) as total_subnets,
			(SELECT COUNT(*) FROM tenant_networks WHERE is_active = true) as allocated_subnets,
			(SELECT COUNT(*) FROM tenant_networks WHERE is_active = false) as available_subnets,
			CASE 
				WHEN (SELECT COUNT(*) FROM tenant_networks) > 0 
				THEN (SELECT COUNT(*) FROM tenant_networks WHERE is_active = true)::float / (SELECT COUNT(*) FROM tenant_networks)::float * 100
				ELSE 0 
			END as utilization_rate`

	var stats IPAMStats
	err := r.db.Get(&stats, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get IPAM stats: %w", err)
	}

	return &stats, nil
}

// UpdateTenant updates a tenant
func (r *IPAMRepository) UpdateTenant(tenant *Tenant) error {
	query := `
		UPDATE tenants SET 
			description = :description,
			max_subnets = :max_subnets,
			status = :status,
			updated_at = :updated_at
		WHERE tenant_id = :tenant_id`

	_, err := r.db.NamedExec(query, tenant)
	if err != nil {
		return fmt.Errorf("failed to update tenant: %w", err)
	}

	r.logger.Info("Tenant updated", zap.String("tenant_id", tenant.TenantID))
	return nil
}

// GetTenantByID retrieves a tenant by database ID
func (r *IPAMRepository) GetTenantByID(id int) (*Tenant, error) {
	query := `SELECT * FROM tenants WHERE id = $1`
	
	var tenant Tenant
	err := r.db.Get(&tenant, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant by ID: %w", err)
	}

	return &tenant, nil
}

// GetTenantsByStatus retrieves tenants by status
func (r *IPAMRepository) GetTenantsByStatus(status string) ([]*Tenant, error) {
	query := `SELECT * FROM tenants WHERE status = $1 ORDER BY created_at DESC`
	
	var tenants []*Tenant
	err := r.db.Select(&tenants, query, status)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenants by status: %w", err)
	}

	return tenants, nil
}

// GetTenantNetworksByStatus retrieves tenant networks by status
func (r *IPAMRepository) GetTenantNetworksByStatus(isActive bool) ([]*TenantNetwork, error) {
	query := `SELECT * FROM tenant_networks WHERE is_active = $1 ORDER BY created_at DESC`
	
	var networks []*TenantNetwork
	err := r.db.Select(&networks, query, isActive)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant networks by status: %w", err)
	}

	return networks, nil
}

// GetTenantNetworkStats returns statistics for a specific tenant
func (r *IPAMRepository) GetTenantNetworkStats(tenantID string) (map[string]interface{}, error) {
	query := `
		SELECT 
			COUNT(*) as total_networks,
			COUNT(CASE WHEN is_active = true THEN 1 END) as active_networks,
			COUNT(CASE WHEN is_active = false THEN 1 END) as inactive_networks,
			MIN(created_at) as first_network,
			MAX(created_at) as last_network
		FROM tenant_networks 
		WHERE tenant_id = $1`

	var stats map[string]interface{}
	err := r.db.Get(&stats, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant network stats: %w", err)
	}

	return stats, nil
}
