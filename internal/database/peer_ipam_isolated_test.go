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

func setupIsolatedTest(t *testing.T) *PeerIPAMRepository {
	// Создаем тестовую БД
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	// Создаем таблицу
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

	// Создаем репозиторий
	logger := zap.NewNop()
	return NewPeerIPAMRepository(db, logger)
}

func TestPeerIPAMRepository_Isolated(t *testing.T) {
	t.Run("basic allocation", func(t *testing.T) {
		repo := setupIsolatedTest(t)
		tenantID := "tenant-123"
		peerID := "peer-456"
		tenantCIDR := "10.100.77.0/24"

		ip, err := repo.AllocateIPForPeer(tenantID, peerID, tenantCIDR)
		require.NoError(t, err)
		assert.Equal(t, "10.100.77.10", ip)
	})

	t.Run("sequential allocation", func(t *testing.T) {
		repo := setupIsolatedTest(t)
		tenantID := "tenant-456"
		tenantCIDR := "10.100.77.0/24"

		ip1, err := repo.AllocateIPForPeer(tenantID, "peer-1", tenantCIDR)
		require.NoError(t, err)
		assert.Equal(t, "10.100.77.10", ip1)

		ip2, err := repo.AllocateIPForPeer(tenantID, "peer-2", tenantCIDR)
		require.NoError(t, err)
		assert.Equal(t, "10.100.77.11", ip2)

		assert.NotEqual(t, ip1, ip2)
	})

	t.Run("reuse existing allocation", func(t *testing.T) {
		repo := setupIsolatedTest(t)
		tenantID := "tenant-789"
		peerID := "peer-789"
		tenantCIDR := "10.100.77.0/24"

		ip1, err := repo.AllocateIPForPeer(tenantID, peerID, tenantCIDR)
		require.NoError(t, err)

		ip2, err := repo.AllocateIPForPeer(tenantID, peerID, tenantCIDR)
		require.NoError(t, err)

		assert.Equal(t, ip1, ip2)
	})


	t.Run("batch deallocation", func(t *testing.T) {
		repo := setupIsolatedTest(t)
		tenantID := "tenant-batch"
		tenantCIDR := "10.100.77.0/24"

		// Выделяем несколько IP
		_, err := repo.AllocateIPForPeer(tenantID, "peer-1", tenantCIDR)
		require.NoError(t, err)
		_, err = repo.AllocateIPForPeer(tenantID, "peer-2", tenantCIDR)
		require.NoError(t, err)

		// Batch деалокация
		count, err := repo.DeallocateIPsByTenant(context.Background(), tenantID)
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)

		// Проверяем, что все IP деактивированы
		var activeCount int
		err = repo.db.QueryRowContext(context.Background(),
			"SELECT COUNT(*) FROM peer_ip_allocations WHERE tenant_id = ? AND is_active = true",
			tenantID).Scan(&activeCount)
		require.NoError(t, err)
		assert.Equal(t, 0, activeCount)
	})

	t.Run("get peer IP", func(t *testing.T) {
		repo := setupIsolatedTest(t)
		tenantID := "tenant-get"
		peerID := "peer-get"
		tenantCIDR := "10.100.77.0/24"

		// Выделяем IP
		expectedIP, err := repo.AllocateIPForPeer(tenantID, peerID, tenantCIDR)
		require.NoError(t, err)

		// Получаем IP
		actualIP, err := repo.GetPeerIP(peerID)
		require.NoError(t, err)
		assert.Equal(t, expectedIP, actualIP)
	})

	t.Run("get tenant peer IPs", func(t *testing.T) {
		repo := setupIsolatedTest(t)
		tenantID := "tenant-list"
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
}
