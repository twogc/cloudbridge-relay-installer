package database

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// setupTestDB создает тестовую БД в памяти
func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	// Создаем таблицу peer_ip_allocations
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS peer_ip_allocations (
		id            TEXT PRIMARY KEY,
		tenant_id     TEXT NOT NULL,
		peer_id       TEXT NOT NULL,
		allocated_ip  TEXT NOT NULL,
		allocated_at  DATETIME NOT NULL,
		released_at   DATETIME,
		is_active     BOOLEAN NOT NULL DEFAULT 1,
		created_at    DATETIME NOT NULL,
		updated_at    DATETIME NOT NULL,
		version       INTEGER NOT NULL DEFAULT 1
	);

	CREATE UNIQUE INDEX IF NOT EXISTS uq_peer_ip_active 
	ON peer_ip_allocations(tenant_id, allocated_ip) 
	WHERE is_active = true;

	CREATE UNIQUE INDEX IF NOT EXISTS uq_peer_active_allocation 
	ON peer_ip_allocations(peer_id) 
	WHERE is_active = true;
	`

	_, err = db.Exec(createTableSQL)
	require.NoError(t, err)

	return db
}

// setupTestRepo создает тестовый репозиторий
func setupTestRepo(t *testing.T) *PeerIPAMRepository {
	db := setupTestDB(t)
	logger := zap.NewNop()
	return NewPeerIPAMRepository(db, logger)
}

func TestPeerIPAMRepository_AllocateIPForPeer(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.Background()

	t.Run("successful allocation", func(t *testing.T) {
		tenantID := "tenant-123"
		peerID := "peer-456"
		tenantCIDR := "10.100.77.0/24"

		ip, err := repo.AllocateIPForPeer(tenantID, peerID, tenantCIDR)
		require.NoError(t, err)
		assert.Equal(t, "10.100.77.10", ip) // Первый доступный IP

		// Проверяем, что IP записан в БД
		var allocatedIP string
		var isActive bool
		err = repo.db.QueryRowContext(ctx,
			"SELECT allocated_ip, is_active FROM peer_ip_allocations WHERE peer_id = ?",
			peerID).Scan(&allocatedIP, &isActive)
		require.NoError(t, err)
		assert.Equal(t, "10.100.77.10", allocatedIP)
		assert.True(t, isActive)
	})

	t.Run("sequential allocation", func(t *testing.T) {
		tenantID := "tenant-123"
		tenantCIDR := "10.100.77.0/24"

		// Выделяем первый IP
		ip1, err := repo.AllocateIPForPeer(tenantID, "peer-1", tenantCIDR)
		require.NoError(t, err)
		assert.Equal(t, "10.100.77.10", ip1)

		// Выделяем второй IP
		ip2, err := repo.AllocateIPForPeer(tenantID, "peer-2", tenantCIDR)
		require.NoError(t, err)
		assert.Equal(t, "10.100.77.11", ip2)

		// Проверяем, что IP разные
		assert.NotEqual(t, ip1, ip2)
	})

	t.Run("reuse existing allocation", func(t *testing.T) {
		tenantID := "tenant-123"
		peerID := "peer-456"
		tenantCIDR := "10.100.77.0/24"

		// Первое выделение
		ip1, err := repo.AllocateIPForPeer(tenantID, peerID, tenantCIDR)
		require.NoError(t, err)

		// Повторное выделение для того же peer
		ip2, err := repo.AllocateIPForPeer(tenantID, peerID, tenantCIDR)
		require.NoError(t, err)

		// Должен вернуться тот же IP
		assert.Equal(t, ip1, ip2)
	})

	t.Run("exhausted subnet", func(t *testing.T) {
		// Создаем отдельный репозиторий для этого теста
		repo := setupTestRepo(t)
		tenantID := "tenant-exhausted"
		tenantCIDR := "10.100.77.0/30" // Только 4 адреса: .0, .1, .2, .3

		// Выделяем все доступные IP (исключая .0, .1, .3 - network, gateway, broadcast)
		ip1, err := repo.AllocateIPForPeer(tenantID, "peer-1", tenantCIDR)
		require.NoError(t, err)
		assert.Equal(t, "10.100.77.2", ip1)

		// Попытка выделить еще один IP должна завершиться ошибкой
		_, err = repo.AllocateIPForPeer(tenantID, "peer-2", tenantCIDR)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no available IP")
	})
}

func TestPeerIPAMRepository_DeallocateIPForPeer(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.Background()

	t.Run("successful deallocation", func(t *testing.T) {
		tenantID := "tenant-123"
		peerID := "peer-456"
		tenantCIDR := "10.100.77.0/24"

		// Выделяем IP
		_, err := repo.AllocateIPForPeer(tenantID, peerID, tenantCIDR)
		require.NoError(t, err)

		// Освобождаем IP
		err = repo.DeallocateIPForPeer(peerID)
		require.NoError(t, err)

		// Проверяем, что IP деактивирован
		var isActive bool
		var releasedAt sql.NullTime
		err = repo.db.QueryRowContext(ctx,
			"SELECT is_active, released_at FROM peer_ip_allocations WHERE peer_id = ?",
			peerID).Scan(&isActive, &releasedAt)
		require.NoError(t, err)
		assert.False(t, isActive)
		assert.True(t, releasedAt.Valid)
	})

	t.Run("deallocate non-existent peer", func(t *testing.T) {
		err := repo.DeallocateIPForPeer("non-existent-peer")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no active IP allocation found")
	})
}

func TestPeerIPAMRepository_DeallocateIPsByTenant(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.Background()

	t.Run("batch deallocation", func(t *testing.T) {
		tenantID := "tenant-123"
		tenantCIDR := "10.100.77.0/24"

		// Выделяем несколько IP для одного тенанта
		_, err := repo.AllocateIPForPeer(tenantID, "peer-1", tenantCIDR)
		require.NoError(t, err)
		_, err = repo.AllocateIPForPeer(tenantID, "peer-2", tenantCIDR)
		require.NoError(t, err)

		// Batch деалокация
		count, err := repo.DeallocateIPsByTenant(ctx, tenantID)
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)

		// Проверяем, что все IP деактивированы
		var activeCount int
		err = repo.db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM peer_ip_allocations WHERE tenant_id = ? AND is_active = true",
			tenantID).Scan(&activeCount)
		require.NoError(t, err)
		assert.Equal(t, 0, activeCount)

		// Проверяем, что released_at заполнен
		var releasedCount int
		err = repo.db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM peer_ip_allocations WHERE tenant_id = ? AND released_at IS NOT NULL",
			tenantID).Scan(&releasedCount)
		require.NoError(t, err)
		assert.Equal(t, 2, releasedCount)
	})

	t.Run("deallocate non-existent tenant", func(t *testing.T) {
		count, err := repo.DeallocateIPsByTenant(ctx, "non-existent-tenant")
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}

func TestPeerIPAMRepository_GetPeerIP(t *testing.T) {
	repo := setupTestRepo(t)

	t.Run("get existing peer IP", func(t *testing.T) {
		tenantID := "tenant-123"
		peerID := "peer-456"
		tenantCIDR := "10.100.77.0/24"

		// Выделяем IP
		expectedIP, err := repo.AllocateIPForPeer(tenantID, peerID, tenantCIDR)
		require.NoError(t, err)

		// Получаем IP
		actualIP, err := repo.GetPeerIP(peerID)
		require.NoError(t, err)
		assert.Equal(t, expectedIP, actualIP)
	})

	t.Run("get non-existent peer IP", func(t *testing.T) {
		_, err := repo.GetPeerIP("non-existent-peer")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no IP allocation found")
	})
}

func TestPeerIPAMRepository_GetTenantPeerIPs(t *testing.T) {
	repo := setupTestRepo(t)

	t.Run("get tenant peer IPs", func(t *testing.T) {
		tenantID := "tenant-123"
		tenantCIDR := "10.100.77.0/24"

		// Выделяем несколько IP
		ip1, err := repo.AllocateIPForPeer(tenantID, "peer-1", tenantCIDR)
		require.NoError(t, err)
		ip2, err := repo.AllocateIPForPeer(tenantID, "peer-2", tenantCIDR)
		require.NoError(t, err)

		// Получаем все IP тенанта
		ips, err := repo.GetTenantPeerIPs(tenantID)
		require.NoError(t, err)
		assert.Len(t, ips, 2)
		assert.Equal(t, ip1, ips["peer-1"])
		assert.Equal(t, ip2, ips["peer-2"])
	})

	t.Run("get non-existent tenant IPs", func(t *testing.T) {
		ips, err := repo.GetTenantPeerIPs("non-existent-tenant")
		require.NoError(t, err)
		assert.Len(t, ips, 0)
	})
}
