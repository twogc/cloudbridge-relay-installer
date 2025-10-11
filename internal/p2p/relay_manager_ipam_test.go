package p2p

import (
	"database/sql"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/twogc/cloudbridge-relay-installer/internal/database"
	"go.uber.org/zap"
)

func TestTenantNetworkAllocation(t *testing.T) {
	// Создаем in-memory SQLite базу для тестов
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Создаем sqlx.DB
	sqlxDB := sqlx.NewDb(db, "sqlite3")

	// Создаем repository
	logger, _ := zap.NewDevelopment()
	repo := database.NewRepository(sqlxDB, logger)

	// Инициализируем таблицы (упрощенная версия для тестов)
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS tenant_networks (
			id TEXT PRIMARY KEY,
			tenant_id TEXT NOT NULL,
			network_cidr TEXT NOT NULL,
			subnet_counter INTEGER NOT NULL,
			allocated_at DATETIME NOT NULL,
			is_active BOOLEAN NOT NULL DEFAULT 1,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create tenant_networks table: %v", err)
	}

	// Создаем RelayManager с repository
	rm := &RelayManager{
		repository:    repo,
		logger:        logger,
		allocatedNets: make(map[string]string),
		subnetCounter: 0,
	}

	// Тест 1: Выделение подсети для нового тенанта
	tenantID1 := "tenant-test-1"
	network1, err := rm.GetTenantNetwork(tenantID1)
	if err != nil {
		t.Fatalf("Failed to get tenant network: %v", err)
	}

	if network1 == "" {
		t.Fatal("Expected network allocation for new tenant")
	}

	// Проверяем, что подсеть сохранилась в БД
	networks, err := repo.GetAllTenantNetworks()
	if err != nil {
		t.Fatalf("Failed to get networks from DB: %v", err)
	}

	if len(networks) != 1 {
		t.Fatalf("Expected 1 network in DB, got %d", len(networks))
	}

	if networks[0].TenantID != tenantID1 {
		t.Fatalf("Expected tenant ID %s, got %s", tenantID1, networks[0].TenantID)
	}

	// Тест 2: Выделение подсети для второго тенанта
	tenantID2 := "tenant-test-2"
	network2, err := rm.GetTenantNetwork(tenantID2)
	if err != nil {
		t.Fatalf("Failed to get tenant network: %v", err)
	}

	if network2 == "" {
		t.Fatal("Expected network allocation for second tenant")
	}

	if network1 == network2 {
		t.Fatal("Expected different networks for different tenants")
	}

	// Тест 3: Проверка, что тенант получает ту же подсеть при повторном запросе
	network1Again, err := rm.GetTenantNetwork(tenantID1)
	if err != nil {
		t.Fatalf("Failed to get tenant network: %v", err)
	}
	if network1 != network1Again {
		t.Fatalf("Expected same network for same tenant, got %s vs %s", network1, network1Again)
	}

	// Тест 4: Деактивация тенанта и освобождение подсети
	err = rm.DeactivateTenant(tenantID1)
	if err != nil {
		t.Fatalf("Failed to deactivate tenant: %v", err)
	}

	// Проверяем, что подсеть освободилась
	networks, err = repo.GetAllTenantNetworks()
	if err != nil {
		t.Fatalf("Failed to get networks from mock DB: %v", err)
	}

	if len(networks) != 1 {
		t.Fatalf("Expected 1 active network after deactivation, got %d", len(networks))
	}

	// Тест 5: Проверка, что освобожденная подсеть может быть переиспользована
	tenantID3 := "tenant-test-3"
	network3, err := rm.GetTenantNetwork(tenantID3)
	if err != nil {
		t.Fatalf("Failed to get tenant network: %v", err)
	}

	if network3 == "" {
		t.Fatal("Expected network allocation for third tenant")
	}

	// Проверяем, что новая подсеть отличается от предыдущих
	if network3 == network1 || network3 == network2 {
		t.Fatal("Expected unique network for third tenant")
	}
}

func TestIPAMStats(t *testing.T) {
	// Создаем in-memory SQLite базу для тестов
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Создаем sqlx.DB
	sqlxDB := sqlx.NewDb(db, "sqlite3")

	// Создаем repository
	logger, _ := zap.NewDevelopment()
	repo := database.NewRepository(sqlxDB, logger)

	// Инициализируем таблицы
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS tenant_networks (
			id TEXT PRIMARY KEY,
			tenant_id TEXT NOT NULL,
			network_cidr TEXT NOT NULL,
			subnet_counter INTEGER NOT NULL,
			allocated_at DATETIME NOT NULL,
			is_active BOOLEAN NOT NULL DEFAULT 1,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create tenant_networks table: %v", err)
	}

	// Создаем RelayManager с repository
	rm := &RelayManager{
		repository:    repo,
		logger:        logger,
		allocatedNets: make(map[string]string),
		subnetCounter: 0,
	}

	// Выделяем несколько подсетей
	_, err1 := rm.GetTenantNetwork("tenant-1")
	if err1 != nil {
		t.Fatalf("Failed to get tenant network: %v", err1)
	}
	_, err2 := rm.GetTenantNetwork("tenant-2")
	if err2 != nil {
		t.Fatalf("Failed to get tenant network: %v", err2)
	}
	_, err3 := rm.GetTenantNetwork("tenant-3")
	if err3 != nil {
		t.Fatalf("Failed to get tenant network: %v", err3)
	}

	// Получаем статистику
	stats := rm.GetIPAMStats()

	// Проверяем статистику
	if stats["allocated_tenants"] != 3 {
		t.Fatalf("Expected 3 allocated tenants, got %v", stats["allocated_tenants"])
	}

	if stats["max_subnets"] != 65536 {
		t.Fatalf("Expected max_subnets to be 65536, got %v", stats["max_subnets"])
	}

	if stats["db_sync_status"] != "synced" {
		t.Fatalf("Expected db_sync_status to be 'synced', got %v", stats["db_sync_status"])
	}

	if stats["db_active_networks"] != 3 {
		t.Fatalf("Expected 3 active networks in DB, got %v", stats["db_active_networks"])
	}
}

func TestSyncIPAMWithDB(t *testing.T) {
	// Создаем in-memory SQLite базу для тестов
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Создаем sqlx.DB
	sqlxDB := sqlx.NewDb(db, "sqlite3")

	// Создаем repository
	logger, _ := zap.NewDevelopment()
	repo := database.NewRepository(sqlxDB, logger)

	// Инициализируем таблицы
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS tenant_networks (
			id TEXT PRIMARY KEY,
			tenant_id TEXT NOT NULL,
			network_cidr TEXT NOT NULL,
			subnet_counter INTEGER NOT NULL,
			allocated_at DATETIME NOT NULL,
			is_active BOOLEAN NOT NULL DEFAULT 1,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create tenant_networks table: %v", err)
	}

	// Добавляем предварительные данные в БД
	_, err = db.Exec(`
		INSERT INTO tenant_networks (id, tenant_id, network_cidr, subnet_counter, allocated_at, is_active, created_at, updated_at)
		VALUES ('550e8400-e29b-41d4-a716-446655440001', 'tenant-db-1', '10.1.1.0/24', 1, datetime('now'), 1, datetime('now'), datetime('now')),
		       ('550e8400-e29b-41d4-a716-446655440002', 'tenant-db-2', '10.1.2.0/24', 2, datetime('now'), 1, datetime('now'), datetime('now'))
	`)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Создаем RelayManager
	rm := &RelayManager{
		repository:    repo,
		logger:        logger,
		allocatedNets: make(map[string]string),
		subnetCounter: 0,
	}

	// Синхронизируем с БД
	err = rm.SyncIPAMWithDB()
	if err != nil {
		t.Fatalf("Failed to sync with DB: %v", err)
	}

	// Проверяем, что данные синхронизировались
	if len(rm.allocatedNets) != 2 {
		t.Fatalf("Expected 2 allocated networks after sync, got %d", len(rm.allocatedNets))
	}

	if rm.subnetCounter != 2 {
		t.Fatalf("Expected subnet counter to be 2, got %d", rm.subnetCounter)
	}

	// Проверяем конкретные сети
	if rm.allocatedNets["tenant-db-1"] != "10.1.1.0/24" {
		t.Fatalf("Expected network 10.1.1.0/24 for tenant-db-1, got %s", rm.allocatedNets["tenant-db-1"])
	}

	if rm.allocatedNets["tenant-db-2"] != "10.1.2.0/24" {
		t.Fatalf("Expected network 10.1.2.0/24 for tenant-db-2, got %s", rm.allocatedNets["tenant-db-2"])
	}
}
