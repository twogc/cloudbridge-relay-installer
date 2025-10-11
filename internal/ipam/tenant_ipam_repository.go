package ipam

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"net"
	"strconv"
	"time"

	"go.uber.org/zap"
)

// TenantIPAMRepository - улучшенный репозиторий с транзакционностью
type TenantIPAMRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// TenantSubnet - структура подсети тенанта
type TenantSubnet struct {
	ID         int                    `json:"id"`
	TenantID   string                 `json:"tenant_id"`
	SubnetCIDR string                 `json:"subnet_cidr"`
	Gateway    string                 `json:"gateway"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
	Status     string                 `json:"status"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// TenantIPAllocation - структура выделенного IP
type TenantIPAllocation struct {
	ID          int                    `json:"id"`
	TenantID    string                 `json:"tenant_id"`
	PeerID      string                 `json:"peer_id"`
	IPAddress   string                 `json:"ip_address"`
	SubnetID    int                    `json:"subnet_id"`
	AllocatedAt time.Time              `json:"allocated_at"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
	Status      string                 `json:"status"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// IPAllocationRequest - запрос на выделение IP
type IPAllocationRequest struct {
	TenantID   string                 `json:"tenant_id"`
	PeerID     string                 `json:"peer_id"`
	SubnetCIDR string                 `json:"subnet_cidr,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// IPAllocationResult - результат выделения IP
type IPAllocationResult struct {
	IPAddress   string    `json:"ip_address"`
	SubnetCIDR  string    `json:"subnet_cidr"`
	Gateway     string    `json:"gateway"`
	AllocatedAt time.Time `json:"allocated_at"`
}

// TenantIPAMStats - статистика IPAM
type TenantIPAMStats struct {
	TotalTenants   int                    `json:"total_tenants"`
	TotalSubnets   int                    `json:"total_subnets"`
	TotalAllocated int                    `json:"total_allocated"`
	TotalAvailable int                    `json:"total_available"`
	TenantStats    map[string]TenantStats `json:"tenant_stats"`
}

// TenantStats - статистика по тенанту
type TenantStats struct {
	TenantID     string `json:"tenant_id"`
	SubnetCount  int    `json:"subnet_count"`
	AllocatedIPs int    `json:"allocated_ips"`
	AvailableIPs int    `json:"available_ips"`
}

func NewTenantIPAMRepository(db *sql.DB, logger *zap.Logger) *TenantIPAMRepository {
	return &TenantIPAMRepository{
		db:     db,
		logger: logger,
	}
}

// CreateTenantSubnet - создание подсети для тенанта с проверкой коллизий
func (r *TenantIPAMRepository) CreateTenantSubnet(ctx context.Context, tenantID string, salt string) (*TenantSubnet, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Проверяем, есть ли уже подсеть для тенанта
	var existingSubnet TenantSubnet
	err = tx.QueryRowContext(ctx, `
		SELECT id, tenant_id, subnet_cidr, gateway, created_at, updated_at, status
		FROM tenant_subnets 
		WHERE tenant_id = $1 AND status = 'active'
	`, tenantID).Scan(
		&existingSubnet.ID, &existingSubnet.TenantID, &existingSubnet.SubnetCIDR,
		&existingSubnet.Gateway, &existingSubnet.CreatedAt,
		&existingSubnet.UpdatedAt, &existingSubnet.Status,
	)

	if err == nil {
		// Подсеть уже существует
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("failed to commit transaction: %w", err)
		}
		return &existingSubnet, nil
	} else if err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to check existing subnet: %w", err)
	}

	// Генерируем уникальную подсеть с проверкой коллизий
	subnetCIDR, err := r.generateUniqueSubnet(ctx, tx, tenantID, salt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate unique subnet: %w", err)
	}

	// Вычисляем gateway
	gateway, err := r.calculateGateway(subnetCIDR)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate gateway: %w", err)
	}

	// Создаем подсеть
	var subnetID int
	err = tx.QueryRowContext(ctx, `
		INSERT INTO tenant_subnets (tenant_id, subnet_cidr, gateway, status)
		VALUES ($1, $2, $3, 'active')
		RETURNING id
	`, tenantID, subnetCIDR, gateway).Scan(&subnetID)

	if err != nil {
		return nil, fmt.Errorf("failed to create tenant subnet: %w", err)
	}

	// Инициализируем IP tracking для подсети
	if err := r.initializeIPTracking(ctx, tx, subnetID, subnetCIDR); err != nil {
		return nil, fmt.Errorf("failed to initialize IP tracking: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	subnet := &TenantSubnet{
		ID:         subnetID,
		TenantID:   tenantID,
		SubnetCIDR: subnetCIDR,
		Gateway:    gateway,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Status:     "active",
		Metadata:   make(map[string]interface{}),
	}

	r.logger.Info("Created tenant subnet",
		zap.String("tenant_id", tenantID),
		zap.String("subnet_cidr", subnetCIDR),
		zap.String("gateway", gateway))

	return subnet, nil
}

// AllocateIPForPeer - выделение IP для peer с транзакционностью
func (r *TenantIPAMRepository) AllocateIPForPeer(ctx context.Context, req *IPAllocationRequest) (*IPAllocationResult, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Проверяем, есть ли уже IP для этого peer
	var existingIP string
	err = tx.QueryRowContext(ctx, `
		SELECT ip_address FROM tenant_ip_allocations 
		WHERE tenant_id = $1 AND peer_id = $2 AND status = 'active'
	`, req.TenantID, req.PeerID).Scan(&existingIP)

	if err == nil {
		// IP уже выделен
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("failed to commit transaction: %w", err)
		}
		return &IPAllocationResult{
			IPAddress:   existingIP,
			AllocatedAt: time.Now(),
		}, nil
	} else if err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to check existing allocation: %w", err)
	}

	// Получаем подсеть для тенанта
	var subnetCIDR string
	if req.SubnetCIDR != "" {
		subnetCIDR = req.SubnetCIDR
	} else {
		err = tx.QueryRowContext(ctx, `
			SELECT subnet_cidr FROM tenant_subnets 
			WHERE tenant_id = $1 AND status = 'active'
			LIMIT 1
		`, req.TenantID).Scan(&subnetCIDR)
		if err != nil {
			return nil, fmt.Errorf("failed to get tenant subnet: %w", err)
		}
	}

	// Выделяем IP через SQL функцию
	var ipAddress string
	err = tx.QueryRowContext(ctx, `
		SELECT allocate_ip_for_peer($1, $2, $3)
	`, req.TenantID, req.PeerID, subnetCIDR).Scan(&ipAddress)

	if err != nil {
		return nil, fmt.Errorf("failed to allocate IP: %w", err)
	}

	// Получаем информацию о подсети
	var gateway string
	err = tx.QueryRowContext(ctx, `
		SELECT gateway FROM tenant_subnets 
		WHERE tenant_id = $1 AND subnet_cidr = $2
	`, req.TenantID, subnetCIDR).Scan(&gateway)

	if err != nil {
		return nil, fmt.Errorf("failed to get subnet gateway: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	result := &IPAllocationResult{
		IPAddress:   ipAddress,
		SubnetCIDR:  subnetCIDR,
		Gateway:     gateway,
		AllocatedAt: time.Now(),
	}

	r.logger.Info("Allocated IP for peer",
		zap.String("tenant_id", req.TenantID),
		zap.String("peer_id", req.PeerID),
		zap.String("ip_address", ipAddress),
		zap.String("subnet_cidr", subnetCIDR))

	return result, nil
}

// DeallocateIPForPeer - освобождение IP для peer
func (r *TenantIPAMRepository) DeallocateIPForPeer(ctx context.Context, tenantID, peerID string) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Освобождаем IP через SQL функцию
	var deallocated bool
	err = tx.QueryRowContext(ctx, `
		SELECT deallocate_ip_for_peer($1, $2)
	`, tenantID, peerID).Scan(&deallocated)

	if err != nil {
		return fmt.Errorf("failed to deallocate IP: %w", err)
	}

	if !deallocated {
		r.logger.Warn("No IP allocation found for peer",
			zap.String("tenant_id", tenantID),
			zap.String("peer_id", peerID))
		return nil
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	r.logger.Info("Deallocated IP for peer",
		zap.String("tenant_id", tenantID),
		zap.String("peer_id", peerID))

	return nil
}

// GetTenantSubnet - получение подсети тенанта
func (r *TenantIPAMRepository) GetTenantSubnet(ctx context.Context, tenantID string) (*TenantSubnet, error) {
	var subnet TenantSubnet
	err := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, subnet_cidr, gateway, created_at, updated_at, status
		FROM tenant_subnets 
		WHERE tenant_id = $1 AND status = 'active'
	`, tenantID).Scan(
		&subnet.ID, &subnet.TenantID, &subnet.SubnetCIDR,
		&subnet.Gateway, &subnet.CreatedAt,
		&subnet.UpdatedAt, &subnet.Status,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tenant %s not found", tenantID)
		}
		return nil, fmt.Errorf("failed to get tenant subnet: %w", err)
	}

	return &subnet, nil
}

// GetTenantStats - получение статистики IPAM
func (r *TenantIPAMRepository) GetTenantStats(ctx context.Context) (*TenantIPAMStats, error) {
	stats := &TenantIPAMStats{
		TenantStats: make(map[string]TenantStats),
	}

	// Общая статистика
	err := r.db.QueryRowContext(ctx, `
		SELECT 
			COUNT(DISTINCT t.tenant_id) as total_tenants,
			COUNT(DISTINCT ts.id) as total_subnets,
			COUNT(DISTINCT tia.id) as total_allocated
		FROM tenants t
		LEFT JOIN tenant_subnets ts ON t.tenant_id = ts.tenant_id AND ts.status = 'active'
		LEFT JOIN tenant_ip_allocations tia ON t.tenant_id = tia.tenant_id AND tia.status = 'active'
		WHERE t.status = 'active'
	`).Scan(&stats.TotalTenants, &stats.TotalSubnets, &stats.TotalAllocated)

	if err != nil {
		return nil, fmt.Errorf("failed to get general stats: %w", err)
	}

	// Статистика по тенантам
	rows, err := r.db.QueryContext(ctx, `
		SELECT 
			t.tenant_id,
			COUNT(DISTINCT ts.id) as subnet_count,
			COUNT(DISTINCT tia.id) as allocated_ips
		FROM tenants t
		LEFT JOIN tenant_subnets ts ON t.tenant_id = ts.tenant_id AND ts.status = 'active'
		LEFT JOIN tenant_ip_allocations tia ON t.tenant_id = tia.tenant_id AND tia.status = 'active'
		WHERE t.status = 'active'
		GROUP BY t.tenant_id
	`)

	if err != nil {
		return nil, fmt.Errorf("failed to get tenant stats: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var tenantID string
		var subnetCount, allocatedIPs int

		if err := rows.Scan(&tenantID, &subnetCount, &allocatedIPs); err != nil {
			return nil, fmt.Errorf("failed to scan tenant stats: %w", err)
		}

		stats.TenantStats[tenantID] = TenantStats{
			TenantID:     tenantID,
			SubnetCount:  subnetCount,
			AllocatedIPs: allocatedIPs,
			AvailableIPs: 252 - allocatedIPs, // 252 доступных IP в /24 подсети (2..253)
		}
	}

	return stats, nil
}

// generateUniqueSubnet - генерация уникальной подсети с проверкой коллизий
func (r *TenantIPAMRepository) generateUniqueSubnet(ctx context.Context, tx *sql.Tx, tenantID, salt string) (string, error) {
	maxAttempts := 1000

	for attempt := 0; attempt < maxAttempts; attempt++ {
		// Генерируем подсеть с солью для детерминизма
		subnetCIDR := r.generateSubnetCIDR(tenantID, salt, attempt)

		// Проверяем, не занята ли подсеть
		var count int
		err := tx.QueryRowContext(ctx, `
			SELECT COUNT(*) FROM tenant_subnets 
			WHERE subnet_cidr = $1 AND status = 'active'
		`, subnetCIDR).Scan(&count)

		if err != nil {
			return "", fmt.Errorf("failed to check subnet collision: %w", err)
		}

		if count == 0 {
			return subnetCIDR, nil
		}

		r.logger.Warn("Subnet collision detected, retrying",
			zap.String("subnet_cidr", subnetCIDR),
			zap.Int("attempt", attempt))
	}

	return "", fmt.Errorf("failed to generate unique subnet after %d attempts", maxAttempts)
}

// generateSubnetCIDR - генерация CIDR подсети
func (r *TenantIPAMRepository) generateSubnetCIDR(tenantID, salt string, attempt int) string {
	// Используем SHA256 с солью для детерминизма
	hash := sha256.Sum256([]byte(tenantID + salt + strconv.Itoa(attempt)))

	// Генерируем подсеть в диапазоне 10.0.0.0/8
	x := int(hash[0]) % 256
	y := int(hash[1]) % 256
	z := int(hash[2]) % 256

	return fmt.Sprintf("10.%d.%d.%d/24", x, y, z)
}

// calculateGateway - вычисление gateway
func (r *TenantIPAMRepository) calculateGateway(subnetCIDR string) (string, error) {
	_, ipNet, err := net.ParseCIDR(subnetCIDR)
	if err != nil {
		return "", fmt.Errorf("invalid subnet CIDR: %w", err)
	}

	// Gateway - первый IP в подсети
	ip := ipNet.IP.To4()
	ip[3] = 1 // 10.x.y.1
	gateway := ip.String()

	return gateway, nil
}

// initializeIPTracking - инициализация отслеживания IP в подсети
func (r *TenantIPAMRepository) initializeIPTracking(ctx context.Context, tx *sql.Tx, subnetID int, subnetCIDR string) error {
	_, ipNet, err := net.ParseCIDR(subnetCIDR)
	if err != nil {
		return fmt.Errorf("invalid subnet CIDR: %w", err)
	}

	// Генерируем все IP адреса в подсети (исключая network и broadcast)
	baseIP := ipNet.IP.To4()

	// Добавляем IP адреса в tracking таблицу с привязкой к subnet_id
	for i := 2; i <= 253; i++ { // 2-253 (исключаем .1 gateway и .255 broadcast)
		ip := make(net.IP, 4)
		copy(ip, baseIP)
		ip[3] = byte(i)

		_, err := tx.ExecContext(ctx, `
			INSERT INTO ip_usage_tracking (subnet_id, ip_address, status)
			VALUES ($1, $2, 'available')
			ON CONFLICT (subnet_id, ip_address) DO NOTHING
		`, subnetID, ip.String())

		if err != nil {
			return fmt.Errorf("failed to initialize IP tracking for %s: %w", ip.String(), err)
		}
	}

	return nil
}
