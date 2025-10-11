package p2p

import (
	"context"
	"io"
	"testing"
	"time"

	"go.uber.org/zap"
)

// mockQUICConn - mock QUIC соединение для тестов
type mockQUICConn struct{}

func (m *mockQUICConn) OpenStream() (io.ReadWriteCloser, error) {
	return &mockStream{}, nil
}

func (m *mockQUICConn) OpenStreamSync(ctx context.Context) (io.ReadWriteCloser, error) {
	return &mockStream{}, nil
}

func (m *mockQUICConn) AcceptStream(ctx context.Context) (io.ReadWriteCloser, error) {
	return &mockStream{}, nil
}

func (m *mockQUICConn) CloseWithError(code uint64, desc string) error {
	return nil
}

func (m *mockQUICConn) LocalAddr() string {
	return "127.0.0.1:9090"
}

func (m *mockQUICConn) RemoteAddr() string {
	return "127.0.0.1:9091"
}

// mockStream - mock stream для тестов
type mockStream struct{}

func (m *mockStream) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}

func (m *mockStream) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (m *mockStream) Close() error {
	return nil
}

func (m *mockStream) SetDeadline(t time.Time) error {
	return nil
}

func (m *mockStream) SetReadDeadline(t time.Time) error {
	return nil
}

func (m *mockStream) SetWriteDeadline(t time.Time) error {
	return nil
}

func TestTenantIsolation(t *testing.T) {
	// Создаем RelayManager
	logger, _ := zap.NewDevelopment()
	rm := &RelayManager{
		logger:          logger,
		quicConnections: make(map[string]*QUICConnection),
		allocatedNets:   make(map[string]string),
		tenants:         make(map[string]*Tenant),
		connections:     make(map[string]*RelayP2PConnection),
		subnetCounter:   0,
	}

	// Создаем два тенанта с разными подсетями
	tenantA := "tenant-isolation-test-a"
	tenantB := "tenant-isolation-test-b"

	// Выделяем подсети для тенантов
	subnetA, err := rm.GetTenantNetwork(tenantA)
	if err != nil {
		t.Fatalf("Failed to get tenant A network: %v", err)
	}
	subnetB, err := rm.GetTenantNetwork(tenantB)
	if err != nil {
		t.Fatalf("Failed to get tenant B network: %v", err)
	}

	// Проверяем, что подсети разные
	if subnetA == subnetB {
		t.Fatalf("Tenants got same subnet: %s == %s", subnetA, subnetB)
	}

	t.Logf("Tenant A subnet: %s", subnetA)
	t.Logf("Tenant B subnet: %s", subnetB)

	// Создаем пиров для каждого тенанта
	peerA1 := "peer-a-1"
	peerA2 := "peer-a-2"
	peerB1 := "peer-b-1"
	peerB2 := "peer-b-2"

	// Создаем записи тенантов
	rm.tenants[tenantA] = &Tenant{
		TenantID:       tenantA,
		MaxConnections: 100,
		Connections:    make(map[string]*RelayP2PConnection),
	}
	rm.tenants[tenantB] = &Tenant{
		TenantID:       tenantB,
		MaxConnections: 100,
		Connections:    make(map[string]*RelayP2PConnection),
	}

	// Регистрируем пиров (без QUIC соединений для простоты теста)
	rm.quicConnections[peerA1] = &QUICConnection{
		PeerID:   peerA1,
		TenantID: tenantA,
		IsActive: true,
		Conn:     nil, // Не используем реальные QUIC соединения в тесте
	}
	rm.quicConnections[peerA2] = &QUICConnection{
		PeerID:   peerA2,
		TenantID: tenantA,
		IsActive: true,
		Conn:     nil,
	}
	rm.quicConnections[peerB1] = &QUICConnection{
		PeerID:   peerB1,
		TenantID: tenantB,
		IsActive: true,
		Conn:     nil,
	}
	rm.quicConnections[peerB2] = &QUICConnection{
		PeerID:   peerB2,
		TenantID: tenantB,
		IsActive: true,
		Conn:     nil,
	}

	// Тест 1: Внутритенантное соединение должно работать
	t.Run("Intra-tenant communication", func(t *testing.T) {
		// Сначала нужно зарегистрировать тенантов в allocatedNets
		rm.allocatedNets[tenantA] = subnetA
		rm.allocatedNets[tenantB] = subnetB

		// Создаем P2P соединение внутри тенанта A
		conn, err := rm.EstablishP2PConnection(peerA1, peerA2)
		if err != nil {
			t.Fatalf("Failed to establish intra-tenant connection: %v", err)
		}

		if conn.TenantID != tenantA {
			t.Fatalf("Expected tenant ID %s, got %s", tenantA, conn.TenantID)
		}

		// Тестируем пересылку пакета внутри тенанта (пропускаем, так как требует реальные QUIC соединения)
		// testData := []byte("TO:" + peerA2 + ":test-message")
		// err = rm.RelayPacket(peerA1, testData)
		// if err != nil {
		// 	t.Fatalf("Intra-tenant packet relay failed: %v", err)
		// }
		t.Log("Intra-tenant packet relay test skipped (requires real QUIC connections)")
	})

	// Тест 2: Межтенантное соединение должно блокироваться
	t.Run("Cross-tenant communication blocked", func(t *testing.T) {
		// Пытаемся создать P2P соединение между разными тенантами
		_, err := rm.EstablishP2PConnection(peerA1, peerB1)
		if err == nil {
			t.Fatal("Cross-tenant connection should be blocked, but succeeded")
		}

		// Проверяем, что ошибка связана с изоляцией
		expectedError := "cross-tenant relay not allowed: tenant-isolation-test-a -> tenant-isolation-test-b"
		if err.Error() != expectedError {
			t.Fatalf("Expected cross-tenant isolation error, got: %v", err)
		}
	})

	// Тест 3: Межтенантная пересылка пакетов должна блокироваться
	t.Run("Cross-tenant packet relay blocked", func(t *testing.T) {
		// Пытаемся переслать пакет между разными тенантами
		testData := []byte("TO:" + peerB1 + ":test-message")
		err := rm.RelayPacket(peerA1, testData)
		if err == nil {
			t.Fatal("Cross-tenant packet relay should be blocked, but succeeded")
		}

		// Проверяем, что ошибка связана с изоляцией
		expectedError := "cross-tenant relay not allowed: tenant-isolation-test-a -> tenant-isolation-test-b"
		if err.Error() != expectedError {
			t.Fatalf("Expected cross-tenant isolation error, got: %v", err)
		}
	})

	// Тест 4: Проверяем, что каждый тенант имеет уникальную подсеть
	t.Run("Unique tenant subnets", func(t *testing.T) {
		// Создаем еще один тенант
		tenantC := "tenant-isolation-test-c"
		subnetC, err := rm.GetTenantNetwork(tenantC)
		if err != nil {
			t.Fatalf("Failed to get tenant C network: %v", err)
		}

		// Проверяем, что все подсети уникальны
		subnets := []string{subnetA, subnetB, subnetC}
		seen := make(map[string]bool)

		for _, subnet := range subnets {
			if seen[subnet] {
				t.Fatalf("Duplicate subnet found: %s", subnet)
			}
			seen[subnet] = true
		}

		t.Logf("All subnets are unique: %v", subnets)
	})

	// Тест 5: Проверяем статистику изоляции
	t.Run("Isolation statistics", func(t *testing.T) {
		stats := rm.GetIPAMStats()

		// Проверяем, что статистика содержит информацию о тенантах
		if stats["allocated_tenants"] == nil {
			t.Fatal("IPAM stats should contain allocated_tenants")
		}

		allocatedTenants := stats["allocated_tenants"].(int)
		if allocatedTenants < 3 {
			t.Fatalf("Expected at least 3 allocated tenants, got %d", allocatedTenants)
		}

		t.Logf("IPAM stats: %+v", stats)
	})
}

func TestTenantIsolationEdgeCases(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	rm := &RelayManager{
		logger:          logger,
		quicConnections: make(map[string]*QUICConnection),
		allocatedNets:   make(map[string]string),
		subnetCounter:   0,
	}

	// Тест 1: Несуществующий отправитель
	t.Run("Non-existent source peer", func(t *testing.T) {
		testData := []byte("TO:target-peer:test-message")
		err := rm.RelayPacket("non-existent-peer", testData)
		if err == nil {
			t.Fatal("Should fail for non-existent source peer")
		}

		expectedError := "source peer connection not found: non-existent-peer"
		if err.Error() != expectedError {
			t.Fatalf("Expected error '%s', got '%s'", expectedError, err.Error())
		}
	})

	// Тест 2: Несуществующий получатель
	t.Run("Non-existent target peer", func(t *testing.T) {
		// Создаем источник
		rm.quicConnections["source-peer"] = &QUICConnection{
			PeerID:   "source-peer",
			TenantID: "tenant-test",
			IsActive: true,
		}

		testData := []byte("TO:non-existent-target:test-message")
		err := rm.RelayPacket("source-peer", testData)
		if err == nil {
			t.Fatal("Should fail for non-existent target peer")
		}

		expectedError := "failed to parse packet target: target peer not found: non-existent-target"
		if err.Error() != expectedError {
			t.Fatalf("Expected error '%s', got '%s'", expectedError, err.Error())
		}
	})

	// Тест 3: Неправильный формат пакета
	t.Run("Invalid packet format", func(t *testing.T) {
		rm.quicConnections["source-peer"] = &QUICConnection{
			PeerID:   "source-peer",
			TenantID: "tenant-test",
			IsActive: true,
		}

		// Неправильный формат (без "TO:")
		testData := []byte("INVALID:target-peer:test-message")
		err := rm.RelayPacket("source-peer", testData)
		if err == nil {
			t.Fatal("Should fail for invalid packet format")
		}

		expectedError := "failed to parse packet target: invalid packet format"
		if err.Error() != expectedError {
			t.Fatalf("Expected error '%s', got '%s'", expectedError, err.Error())
		}
	})
}

func TestTenantSubnetAllocation(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	rm := &RelayManager{
		logger:        logger,
		allocatedNets: make(map[string]string),
		subnetCounter: 0,
	}

	// Тест 1: Каждый тенант получает уникальную подсеть
	t.Run("Unique subnet per tenant", func(t *testing.T) {
		tenants := []string{
			"tenant-unique-test-1",
			"tenant-unique-test-2",
			"tenant-unique-test-3",
			"tenant-unique-test-4",
			"tenant-unique-test-5",
		}

		subnets := make(map[string]string)

		for _, tenant := range tenants {
			subnet, err := rm.GetTenantNetwork(tenant)
			if err != nil {
				t.Fatalf("Failed to get tenant network: %v", err)
			}
			subnets[tenant] = subnet

			// Проверяем, что подсеть не пустая
			if subnet == "" {
				t.Fatalf("Tenant %s got empty subnet", tenant)
			}
		}

		// Проверяем уникальность
		seenSubnets := make(map[string]string)
		for tenant, subnet := range subnets {
			if existingTenant, exists := seenSubnets[subnet]; exists {
				t.Fatalf("Duplicate subnet %s for tenants %s and %s", subnet, existingTenant, tenant)
			}
			seenSubnets[subnet] = tenant
		}

		t.Logf("All tenants got unique subnets:")
		for tenant, subnet := range subnets {
			t.Logf("  %s: %s", tenant, subnet)
		}
	})

	// Тест 2: Повторное получение подсети для того же тенанта
	t.Run("Same tenant gets same subnet", func(t *testing.T) {
		tenant := "tenant-consistency-test"

		// Получаем подсеть первый раз
		subnet1, err := rm.GetTenantNetwork(tenant)
		if err != nil {
			t.Fatalf("Failed to get tenant network: %v", err)
		}

		// Получаем подсеть второй раз
		subnet2, err := rm.GetTenantNetwork(tenant)
		if err != nil {
			t.Fatalf("Failed to get tenant network: %v", err)
		}

		// Подсети должны быть одинаковыми
		if subnet1 != subnet2 {
			t.Fatalf("Same tenant got different subnets: %s vs %s", subnet1, subnet2)
		}

		t.Logf("Tenant %s consistently gets subnet %s", tenant, subnet1)
	})
}
