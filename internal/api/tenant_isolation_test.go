package api

import (
	"testing"

	"go.uber.org/zap"
)

func TestTenantIPIsolation(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	pm := NewPeerManagerFixed(logger)

	// Тест 1: Создаем пиров для разных тенантов
	tenantA := "tenant-security-test-a"
	tenantB := "tenant-security-test-b"
	tenantC := "tenant-security-test-c"

	// Регистрируем пиров для тенанта A
	ipA1, err := pm.allocateIPForTenantFixed(tenantA, "peer-a-1")
	if err != nil {
		t.Fatalf("Failed to allocate IP for tenant A peer 1: %v", err)
	}

	ipA2, err := pm.allocateIPForTenantFixed(tenantA, "peer-a-2")
	if err != nil {
		t.Fatalf("Failed to allocate IP for tenant A peer 2: %v", err)
	}

	// Регистрируем пиров для тенанта B
	ipB1, err := pm.allocateIPForTenantFixed(tenantB, "peer-b-1")
	if err != nil {
		t.Fatalf("Failed to allocate IP for tenant B peer 1: %v", err)
	}

	ipB2, err := pm.allocateIPForTenantFixed(tenantB, "peer-b-2")
	if err != nil {
		t.Fatalf("Failed to allocate IP for tenant B peer 2: %v", err)
	}

	// Регистрируем пиров для тенанта C
	ipC1, err := pm.allocateIPForTenantFixed(tenantC, "peer-c-1")
	if err != nil {
		t.Fatalf("Failed to allocate IP for tenant C peer 1: %v", err)
	}

	// Тест 2: Проверяем, что тенанты получили разные подсети
	subnetA, err := pm.GetTenantSubnet(tenantA)
	if err != nil {
		t.Fatalf("Failed to get subnet for tenant A: %v", err)
	}

	subnetB, err := pm.GetTenantSubnet(tenantB)
	if err != nil {
		t.Fatalf("Failed to get subnet for tenant B: %v", err)
	}

	subnetC, err := pm.GetTenantSubnet(tenantC)
	if err != nil {
		t.Fatalf("Failed to get subnet for tenant C: %v", err)
	}

	// Проверяем, что подсети разные
	if subnetA == subnetB {
		t.Errorf("Tenant A and B have the same subnet: %s", subnetA)
	}
	if subnetA == subnetC {
		t.Errorf("Tenant A and C have the same subnet: %s", subnetA)
	}
	if subnetB == subnetC {
		t.Errorf("Tenant B and C have the same subnet: %s", subnetB)
	}

	t.Logf("Tenant A subnet: %s", subnetA)
	t.Logf("Tenant B subnet: %s", subnetB)
	t.Logf("Tenant C subnet: %s", subnetC)

	// Тест 3: Проверяем, что IP адреса находятся в правильных подсетях
	if !isIPInSubnet(ipA1, subnetA) {
		t.Errorf("IP %s is not in tenant A subnet %s", ipA1, subnetA)
	}
	if !isIPInSubnet(ipA2, subnetA) {
		t.Errorf("IP %s is not in tenant A subnet %s", ipA2, subnetA)
	}
	if !isIPInSubnet(ipB1, subnetB) {
		t.Errorf("IP %s is not in tenant B subnet %s", ipB1, subnetB)
	}
	if !isIPInSubnet(ipB2, subnetB) {
		t.Errorf("IP %s is not in tenant B subnet %s", ipB2, subnetB)
	}
	if !isIPInSubnet(ipC1, subnetC) {
		t.Errorf("IP %s is not in tenant C subnet %s", ipC1, subnetC)
	}

	// Тест 4: Проверяем изоляцию - IP адреса разных тенантов не должны пересекаться
	tenantAIPs, err := pm.GetTenantIPs(tenantA)
	if err != nil {
		t.Fatalf("Failed to get IPs for tenant A: %v", err)
	}

	tenantBIPs, err := pm.GetTenantIPs(tenantB)
	if err != nil {
		t.Fatalf("Failed to get IPs for tenant B: %v", err)
	}

	tenantCIPs, err := pm.GetTenantIPs(tenantC)
	if err != nil {
		t.Fatalf("Failed to get IPs for tenant C: %v", err)
	}

	// Проверяем, что нет пересечений IP адресов между тенантами
	for peerID, ip := range tenantAIPs {
		for otherPeerID, otherIP := range tenantBIPs {
			if ip == otherIP {
				t.Errorf("IP collision: %s (tenant A peer %s) == %s (tenant B peer %s)", ip, peerID, otherIP, otherPeerID)
			}
		}
		for otherPeerID, otherIP := range tenantCIPs {
			if ip == otherIP {
				t.Errorf("IP collision: %s (tenant A peer %s) == %s (tenant C peer %s)", ip, peerID, otherIP, otherPeerID)
			}
		}
	}

	for peerID, ip := range tenantBIPs {
		for otherPeerID, otherIP := range tenantCIPs {
			if ip == otherIP {
				t.Errorf("IP collision: %s (tenant B peer %s) == %s (tenant C peer %s)", ip, peerID, otherIP, otherPeerID)
			}
		}
	}

	// Тест 5: Проверяем валидацию изоляции
	if err := pm.ValidateTenantIsolation(); err != nil {
		t.Errorf("Tenant isolation validation failed: %v", err)
	}

	// Тест 6: Проверяем статистику
	stats := pm.GetIsolationStats()
	if stats["total_tenants"] != 3 {
		t.Errorf("Expected 3 tenants, got %v", stats["total_tenants"])
	}
	if stats["total_allocated"] != 5 {
		t.Errorf("Expected 5 allocated IPs, got %v", stats["total_allocated"])
	}

	t.Logf("Isolation stats: %+v", stats)

	t.Logf("✅ All isolation tests passed!")
	t.Logf("Tenant A IPs: %+v", tenantAIPs)
	t.Logf("Tenant B IPs: %+v", tenantBIPs)
	t.Logf("Tenant C IPs: %+v", tenantCIPs)
}

func TestTenantSubnetConsistency(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	pm := NewPeerManagerFixed(logger)

	tenantID := "tenant-consistency-test"

	// Создаем подсеть для тенанта
	subnet1, err := pm.getOrCreateTenantSubnet(tenantID)
	if err != nil {
		t.Fatalf("Failed to create subnet: %v", err)
	}

	// Создаем подсеть для того же тенанта еще раз
	subnet2, err := pm.getOrCreateTenantSubnet(tenantID)
	if err != nil {
		t.Fatalf("Failed to get existing subnet: %v", err)
	}

	// Проверяем, что подсеть одинаковая (консистентность)
	if subnet1 != subnet2 {
		t.Errorf("Subnet inconsistency: %s != %s", subnet1, subnet2)
	}

	t.Logf("Subnet consistency test passed: %s", subnet1)
}

func TestIPAllocationInSubnet(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	pm := NewPeerManagerFixed(logger)

	tenantID := "tenant-subnet-test"
	subnet := "10.100.1.0/24"

	// Выделяем несколько IP в подсети
	ip1, err := pm.allocateIPInTenantSubnet(tenantID, "peer-1", subnet)
	if err != nil {
		t.Fatalf("Failed to allocate IP 1: %v", err)
	}

	ip2, err := pm.allocateIPInTenantSubnet(tenantID, "peer-2", subnet)
	if err != nil {
		t.Fatalf("Failed to allocate IP 2: %v", err)
	}

	// Проверяем, что IP разные
	if ip1 == ip2 {
		t.Errorf("Allocated same IP for different peers: %s", ip1)
	}

	// Проверяем, что IP в правильной подсети
	if !isIPInSubnet(ip1, subnet) {
		t.Errorf("IP %s is not in subnet %s", ip1, subnet)
	}
	if !isIPInSubnet(ip2, subnet) {
		t.Errorf("IP %s is not in subnet %s", ip2, subnet)
	}

	t.Logf("IP allocation test passed: %s, %s", ip1, ip2)
}

// isIPInSubnet проверяет, находится ли IP в подсети
func isIPInSubnet(ip, subnet string) bool {
	// Простая проверка для формата 10.x.y.z в подсети 10.x.y.0/24
	if len(ip) < 8 || len(subnet) < 10 {
		return false
	}

	// Извлекаем базовую часть IP (10.x.y)
	// Находим последнюю точку в IP
	lastDotIndex := -1
	for i := len(ip) - 1; i >= 0; i-- {
		if ip[i] == '.' {
			lastDotIndex = i
			break
		}
	}

	if lastDotIndex == -1 {
		return false
	}

	ipBase := ip[:lastDotIndex]          // 10.x.y
	subnetBase := subnet[:len(subnet)-5] // Убираем .0/24

	return ipBase == subnetBase
}
