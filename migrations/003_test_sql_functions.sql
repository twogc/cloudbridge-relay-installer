-- Тестирование SQL функций для IPAM
-- Миграция 003: Тестовые данные и проверка функций

-- Создаем тестового тенанта
INSERT INTO tenants (tenant_id, name, status) 
VALUES ('test-tenant-001', 'Test Tenant', 'active')
ON CONFLICT (tenant_id) DO NOTHING;

-- Создаем тестовую подсеть
INSERT INTO tenant_subnets (tenant_id, subnet_cidr, gateway, status)
VALUES ('test-tenant-001', '10.100.1.0/24', '10.100.1.1', 'active')
ON CONFLICT (subnet_cidr) DO NOTHING;

-- Инициализируем IP пул для тестовой подсети
DO $$
DECLARE
    subnet_id INTEGER;
    ip_count INTEGER;
BEGIN
    -- Получаем ID подсети
    SELECT id INTO subnet_id 
    FROM tenant_subnets 
    WHERE tenant_id = 'test-tenant-001' AND subnet_cidr = '10.100.1.0/24';
    
    -- Инициализируем IP пул
    SELECT initialize_subnet_ip_pool(subnet_id, '10.100.1.0/24') INTO ip_count;
    
    RAISE NOTICE 'Initialized % IP addresses for subnet %', ip_count, '10.100.1.0/24';
END $$;

-- Тестируем функцию выделения IP
DO $$
DECLARE
    allocated_ip INET;
    deallocated BOOLEAN;
BEGIN
    -- Выделяем IP для peer-1
    SELECT allocate_ip_for_peer('test-tenant-001', 'peer-1', '10.100.1.0/24') INTO allocated_ip;
    RAISE NOTICE 'Allocated IP % for peer-1', allocated_ip;
    
    -- Пытаемся выделить IP для того же peer (должен вернуть существующий)
    SELECT allocate_ip_for_peer('test-tenant-001', 'peer-1', '10.100.1.0/24') INTO allocated_ip;
    RAISE NOTICE 'Re-allocated IP % for peer-1 (should be same)', allocated_ip;
    
    -- Выделяем IP для peer-2
    SELECT allocate_ip_for_peer('test-tenant-001', 'peer-2', '10.100.1.0/24') INTO allocated_ip;
    RAISE NOTICE 'Allocated IP % for peer-2', allocated_ip;
    
    -- Освобождаем IP для peer-1
    SELECT deallocate_ip_for_peer('test-tenant-001', 'peer-1') INTO deallocated;
    RAISE NOTICE 'Deallocated IP for peer-1: %', deallocated;
    
    -- Освобождаем IP для несуществующего peer
    SELECT deallocate_ip_for_peer('test-tenant-001', 'nonexistent-peer') INTO deallocated;
    RAISE NOTICE 'Deallocated IP for nonexistent peer: %', deallocated;
END $$;

-- Проверяем статистику
SELECT 
    tenant_id,
    total_subnets,
    allocated_ips,
    available_ips
FROM get_tenant_ip_stats('test-tenant-001');

-- Проверяем коллизии подсетей
SELECT 
    '10.100.1.0/24'::CIDR as test_cidr,
    check_subnet_collision('10.100.1.0/24'::CIDR) as has_collision,
    is_cidr_excluded('10.100.1.0/24'::CIDR) as is_excluded;

-- Проверяем исключенные CIDR
SELECT 
    '10.0.0.0/8'::CIDR as test_cidr,
    is_cidr_excluded('10.0.0.0/8'::CIDR) as is_excluded;

-- Показываем текущее состояние
SELECT 
    'Current allocations' as info,
    tenant_id,
    peer_id,
    ip_address,
    status,
    allocated_at
FROM tenant_ip_allocations 
WHERE tenant_id = 'test-tenant-001'
ORDER BY allocated_at;

SELECT 
    'IP usage tracking' as info,
    subnet_id,
    ip_address,
    status
FROM ip_usage_tracking 
WHERE subnet_id = (SELECT id FROM tenant_subnets WHERE tenant_id = 'test-tenant-001')
ORDER BY ip_address
LIMIT 10;


