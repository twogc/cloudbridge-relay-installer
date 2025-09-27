package ipam

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/twogc/cloudbridge-relay-installer/internal/database"
	"go.uber.org/zap"
)

// Service represents the IPAM service
type Service struct {
	db     *database.Repository
	logger *zap.Logger
}

// NewService creates a new IPAM service
func NewService(db *database.Repository, logger *zap.Logger) *Service {
	return &Service{
		db:     db,
		logger: logger,
	}
}

// CreateTenantRequest represents a request to create a tenant
type CreateTenantRequest struct {
	TenantID    string `json:"tenant_id" binding:"required"`
	Description string `json:"description,omitempty"`
	MaxSubnets  int    `json:"max_subnets,omitempty"`
}

// CreateTenantResponse represents a response for tenant creation
type CreateTenantResponse struct {
	TenantID    string    `json:"tenant_id"`
	Description string    `json:"description"`
	MaxSubnets  int       `json:"max_subnets"`
	CreatedAt   time.Time `json:"created_at"`
	Status      string    `json:"status"`
}

// TenantInfo represents tenant information
type TenantInfo struct {
	TenantID    string    `json:"tenant_id"`
	Description string    `json:"description"`
	MaxSubnets  int       `json:"max_subnets"`
	UsedSubnets int       `json:"used_subnets"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Status      string    `json:"status"`
}

// SubnetInfo represents subnet information
type SubnetInfo struct {
	TenantID    string    `json:"tenant_id"`
	SubnetCIDR  string    `json:"subnet_cidr"`
	SubnetID    int       `json:"subnet_id"`
	IsActive    bool      `json:"is_active"`
	AllocatedAt time.Time `json:"allocated_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// AllocateSubnetRequest represents a request to allocate a subnet
type AllocateSubnetRequest struct {
	TenantID   string `json:"tenant_id" binding:"required"`
	SubnetSize int    `json:"subnet_size,omitempty"` // /24, /25, etc.
	Purpose    string `json:"purpose,omitempty"`
}

// AllocateSubnetResponse represents a response for subnet allocation
type AllocateSubnetResponse struct {
	TenantID    string    `json:"tenant_id"`
	SubnetCIDR  string    `json:"subnet_cidr"`
	SubnetID    int       `json:"subnet_id"`
	AllocatedAt time.Time `json:"allocated_at"`
	Status      string    `json:"status"`
}

// IPAMStats represents IPAM statistics
type IPAMStats struct {
	TotalTenants     int       `json:"total_tenants"`
	ActiveTenants    int       `json:"active_tenants"`
	TotalSubnets     int       `json:"total_subnets"`
	AllocatedSubnets int       `json:"allocated_subnets"`
	AvailableSubnets int       `json:"available_subnets"`
	UtilizationRate  float64   `json:"utilization_rate"`
	LastUpdated      time.Time `json:"last_updated"`
}

// CreateTenant creates a new tenant
func (s *Service) CreateTenant(c *gin.Context) {
	var req CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.Error("Invalid request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	s.logger.Info("Creating tenant", zap.String("tenant_id", req.TenantID))

	// Check if tenant already exists
	existing, err := s.db.GetTenant(req.TenantID)
	if err == nil && existing != nil {
		s.logger.Warn("Tenant already exists", zap.String("tenant_id", req.TenantID))
		c.JSON(http.StatusConflict, gin.H{"error": "Tenant already exists"})
		return
	}

	// Set defaults
	if req.MaxSubnets == 0 {
		req.MaxSubnets = 10 // Default limit
	}

	// Create tenant in database
	tenant := &database.Tenant{
		TenantID:    req.TenantID,
		Description: req.Description,
		MaxSubnets:  req.MaxSubnets,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Status:      "active",
	}

	if err := s.db.CreateTenant(tenant); err != nil {
		s.logger.Error("Failed to create tenant", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tenant"})
		return
	}

	response := CreateTenantResponse{
		TenantID:    tenant.TenantID,
		Description: tenant.Description,
		MaxSubnets:  tenant.MaxSubnets,
		CreatedAt:   tenant.CreatedAt,
		Status:      tenant.Status,
	}

	s.logger.Info("Tenant created successfully", zap.String("tenant_id", req.TenantID))
	c.JSON(http.StatusCreated, response)
}

// GetTenant retrieves tenant information
func (s *Service) GetTenant(c *gin.Context) {
	tenantID := c.Param("tenant_id")

	s.logger.Info("Getting tenant", zap.String("tenant_id", tenantID))

	tenant, err := s.db.GetTenant(tenantID)
	if err != nil {
		s.logger.Error("Failed to get tenant", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	// Get used subnets count
	usedSubnets, err := s.db.GetTenantSubnetCount(tenantID)
	if err != nil {
		s.logger.Warn("Failed to get subnet count", zap.Error(err))
		usedSubnets = 0
	}

	response := TenantInfo{
		TenantID:    tenant.TenantID,
		Description: tenant.Description,
		MaxSubnets:  tenant.MaxSubnets,
		UsedSubnets: usedSubnets,
		CreatedAt:   tenant.CreatedAt,
		UpdatedAt:   tenant.UpdatedAt,
		Status:      tenant.Status,
	}

	c.JSON(http.StatusOK, response)
}

// DeleteTenant deletes a tenant
func (s *Service) DeleteTenant(c *gin.Context) {
	tenantID := c.Param("tenant_id")

	s.logger.Info("Deleting tenant", zap.String("tenant_id", tenantID))

	// Check if tenant exists
	_, err := s.db.GetTenant(tenantID)
	if err != nil {
		s.logger.Error("Tenant not found", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	// Deactivate all subnets first
	if err := s.db.DeactivateAllTenantSubnets(tenantID); err != nil {
		s.logger.Error("Failed to deactivate subnets", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to deactivate subnets"})
		return
	}

	// Delete tenant
	if err := s.db.DeleteTenant(tenantID); err != nil {
		s.logger.Error("Failed to delete tenant", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete tenant"})
		return
	}

	s.logger.Info("Tenant deleted successfully", zap.String("tenant_id", tenantID))
	c.JSON(http.StatusOK, gin.H{"message": "Tenant deleted successfully"})
}

// ListTenants lists all tenants
func (s *Service) ListTenants(c *gin.Context) {
	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "100")
	offsetStr := c.DefaultQuery("offset", "0")
	status := c.Query("status")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 100
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	s.logger.Info("Listing tenants",
		zap.Int("limit", limit),
		zap.Int("offset", offset),
		zap.String("status", status))

	tenants, err := s.db.ListTenants(limit, offset, status)
	if err != nil {
		s.logger.Error("Failed to list tenants", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list tenants"})
		return
	}

	// Convert to response format
	var response []TenantInfo
	for _, tenant := range tenants {
		usedSubnets, _ := s.db.GetTenantSubnetCount(tenant.TenantID)

		response = append(response, TenantInfo{
			TenantID:    tenant.TenantID,
			Description: tenant.Description,
			MaxSubnets:  tenant.MaxSubnets,
			UsedSubnets: usedSubnets,
			CreatedAt:   tenant.CreatedAt,
			UpdatedAt:   tenant.UpdatedAt,
			Status:      tenant.Status,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"tenants": response,
		"total":   len(response),
		"limit":   limit,
		"offset":  offset,
	})
}

// AllocateSubnet allocates a subnet for a tenant
func (s *Service) AllocateSubnet(c *gin.Context) {
	tenantID := c.Param("tenant_id")

	var req AllocateSubnetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.Error("Invalid request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set defaults
	if req.SubnetSize == 0 {
		req.SubnetSize = 24 // Default /24
	}

	s.logger.Info("Allocating subnet",
		zap.String("tenant_id", tenantID),
		zap.Int("subnet_size", req.SubnetSize))

	// Check if tenant exists
	tenant, err := s.db.GetTenant(tenantID)
	if err != nil {
		s.logger.Error("Tenant not found", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	// Check subnet limit
	usedSubnets, err := s.db.GetTenantSubnetCount(tenantID)
	if err != nil {
		s.logger.Error("Failed to get subnet count", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check subnet limit"})
		return
	}

	if usedSubnets >= tenant.MaxSubnets {
		s.logger.Warn("Subnet limit exceeded",
			zap.String("tenant_id", tenantID),
			zap.Int("used", usedSubnets),
			zap.Int("max", tenant.MaxSubnets))
		c.JSON(http.StatusForbidden, gin.H{"error": "Subnet limit exceeded"})
		return
	}

	// Allocate subnet using existing IPAM logic
	subnetCIDR, subnetID, err := s.allocateSubnetForTenant(tenantID, req.SubnetSize)
	if err != nil {
		s.logger.Error("Failed to allocate subnet", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to allocate subnet"})
		return
	}

	response := AllocateSubnetResponse{
		TenantID:    tenantID,
		SubnetCIDR:  subnetCIDR,
		SubnetID:    subnetID,
		AllocatedAt: time.Now(),
		Status:      "allocated",
	}

	s.logger.Info("Subnet allocated successfully",
		zap.String("tenant_id", tenantID),
		zap.String("subnet_cidr", subnetCIDR))

	c.JSON(http.StatusCreated, response)
}

// DeallocateSubnet deallocates a subnet
func (s *Service) DeallocateSubnet(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	subnetCIDR := c.Query("subnet_cidr")

	if subnetCIDR == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "subnet_cidr parameter required"})
		return
	}

	s.logger.Info("Deallocating subnet",
		zap.String("tenant_id", tenantID),
		zap.String("subnet_cidr", subnetCIDR))

	if err := s.db.DeactivateTenantSubnet(tenantID, subnetCIDR); err != nil {
		s.logger.Error("Failed to deallocate subnet", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to deallocate subnet"})
		return
	}

	s.logger.Info("Subnet deallocated successfully",
		zap.String("tenant_id", tenantID),
		zap.String("subnet_cidr", subnetCIDR))

	c.JSON(http.StatusOK, gin.H{"message": "Subnet deallocated successfully"})
}

// ListTenantSubnets lists subnets for a tenant
func (s *Service) ListTenantSubnets(c *gin.Context) {
	tenantID := c.Param("tenant_id")

	s.logger.Info("Listing tenant subnets", zap.String("tenant_id", tenantID))

	subnets, err := s.db.GetTenantSubnets(tenantID)
	if err != nil {
		s.logger.Error("Failed to get tenant subnets", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tenant subnets"})
		return
	}

	var response []SubnetInfo
	for _, subnet := range subnets {
		response = append(response, SubnetInfo{
			TenantID:    subnet.TenantID,
			SubnetCIDR:  subnet.NetworkCIDR,
			SubnetID:    subnet.SubnetCounter,
			IsActive:    subnet.IsActive,
			AllocatedAt: subnet.CreatedAt,
			UpdatedAt:   subnet.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"tenant_id": tenantID,
		"subnets":   response,
		"total":     len(response),
	})
}

// GetStats returns IPAM statistics
func (s *Service) GetStats(c *gin.Context) {
	s.logger.Info("Getting IPAM statistics")

	stats, err := s.db.GetIPAMStats()
	if err != nil {
		s.logger.Error("Failed to get IPAM stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get statistics"})
		return
	}

	response := IPAMStats{
		TotalTenants:     stats.TotalTenants,
		ActiveTenants:    stats.ActiveTenants,
		TotalSubnets:     stats.TotalSubnets,
		AllocatedSubnets: stats.AllocatedSubnets,
		AvailableSubnets: stats.AvailableSubnets,
		UtilizationRate:  stats.UtilizationRate,
		LastUpdated:      time.Now(),
	}

	c.JSON(http.StatusOK, response)
}

// GetMetrics returns IPAM metrics for monitoring
func (s *Service) GetMetrics(c *gin.Context) {
	s.logger.Info("Getting IPAM metrics")

	stats, err := s.db.GetIPAMStats()
	if err != nil {
		s.logger.Error("Failed to get IPAM stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get metrics"})
		return
	}

	// Format as Prometheus metrics
	metrics := fmt.Sprintf(`# HELP ipam_tenants_total Total number of tenants
# TYPE ipam_tenants_total gauge
ipam_tenants_total %d

# HELP ipam_active_tenants_total Total number of active tenants
# TYPE ipam_active_tenants_total gauge
ipam_active_tenants_total %d

# HELP ipam_subnets_total Total number of subnets
# TYPE ipam_subnets_total gauge
ipam_subnets_total %d

# HELP ipam_allocated_subnets_total Total number of allocated subnets
# TYPE ipam_allocated_subnets_total gauge
ipam_allocated_subnets_total %d

# HELP ipam_utilization_rate IPAM utilization rate
# TYPE ipam_utilization_rate gauge
ipam_utilization_rate %.2f
`,
		stats.TotalTenants,
		stats.ActiveTenants,
		stats.TotalSubnets,
		stats.AllocatedSubnets,
		stats.UtilizationRate,
	)

	c.Header("Content-Type", "text/plain")
	c.String(http.StatusOK, metrics)
}

// allocateSubnetForTenant allocates a subnet for a tenant using IPAM logic
func (s *Service) allocateSubnetForTenant(tenantID string, subnetSize int) (string, int, error) {
	// This is a simplified version - in production, this would use the full IPAM logic
	// from the existing RelayManager

	// Get next available subnet
	subnetCounter, err := s.db.GetNextSubnetCounter()
	if err != nil {
		return "", 0, fmt.Errorf("failed to get next subnet counter: %w", err)
	}

	// Generate subnet CIDR using the same logic as RelayManager
	x := subnetCounter / 256
	y := subnetCounter % 256
	subnetCIDR := fmt.Sprintf("10.%d.%d.0/%d", x, y, subnetSize)

	// Create subnet record
	subnet := &database.TenantNetwork{
		TenantID:      tenantID,
		NetworkCIDR:   subnetCIDR,
		SubnetCounter: subnetCounter,
		IsActive:      true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.db.CreateTenantNetwork(subnet); err != nil {
		return "", 0, fmt.Errorf("failed to create tenant network: %w", err)
	}

	return subnetCIDR, subnetCounter, nil
}
