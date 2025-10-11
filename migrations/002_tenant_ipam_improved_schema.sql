-- Улучшенная схема БД для IPAM с правильными типами и SQL функциями
-- Миграция 002: Обновление схемы для production-ready IPAM

-- Обновляем таблицу tenants (используем tenant_id как PK)
DROP TABLE IF EXISTS tenants CASCADE;
CREATE TABLE tenants (
    tenant_id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    status TEXT NOT NULL DEFAULT 'active',
    metadata JSONB DEFAULT '{}'::jsonb
);

-- Обновляем таблицу tenant_subnets с правильными типами
DROP TABLE IF EXISTS tenant_subnets CASCADE;
CREATE TABLE tenant_subnets (
    id SERIAL PRIMARY KEY,
    tenant_id TEXT NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    subnet_cidr CIDR NOT NULL UNIQUE,
    gateway INET NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    metadata JSONB DEFAULT '{}'::jsonb
);

-- Создаем уникальный индекс для tenant-subnet комбинации
CREATE UNIQUE INDEX idx_tenant_subnets_tenant_subnet ON tenant_subnets(tenant_id, subnet_cidr);

-- Обновляем таблицу tenant_ip_allocations
DROP TABLE IF EXISTS tenant_ip_allocations CASCADE;
CREATE TABLE tenant_ip_allocations (
    id SERIAL PRIMARY KEY,
    tenant_id TEXT NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    peer_id TEXT NOT NULL,
    ip_address INET NOT NULL,
    subnet_id INTEGER NOT NULL REFERENCES tenant_subnets(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'active',
    allocated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ,
    metadata JSONB DEFAULT '{}'::jsonb,
    
    -- Уникальные ограничения
    UNIQUE (tenant_id, peer_id),
    UNIQUE (subnet_id, ip_address)
);

-- Обновляем таблицу ip_usage_tracking с привязкой к subnet_id
DROP TABLE IF EXISTS ip_usage_tracking CASCADE;
CREATE TABLE ip_usage_tracking (
    subnet_id INTEGER NOT NULL REFERENCES tenant_subnets(id) ON DELETE CASCADE,
    ip_address INET NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('available','allocated','reserved')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    metadata JSONB DEFAULT '{}'::jsonb,
    
    PRIMARY KEY (subnet_id, ip_address)
);

-- Индексы для производительности
CREATE INDEX idx_tenant_subnets_tenant_id ON tenant_subnets(tenant_id);
CREATE INDEX idx_tenant_subnets_subnet_cidr ON tenant_subnets(subnet_cidr);
CREATE INDEX idx_tenant_ip_allocations_tenant_id ON tenant_ip_allocations(tenant_id);
CREATE INDEX idx_tenant_ip_allocations_peer_id ON tenant_ip_allocations(peer_id);
CREATE INDEX idx_tenant_ip_allocations_ip_address ON tenant_ip_allocations(ip_address);
CREATE INDEX idx_ip_usage_tracking_subnet_id ON ip_usage_tracking(subnet_id);
CREATE INDEX idx_ip_usage_tracking_status ON ip_usage_tracking(status);

-- Функция для автоматического обновления updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Триггеры для автоматического обновления updated_at
CREATE TRIGGER update_tenants_updated_at BEFORE UPDATE ON tenants
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_tenant_subnets_updated_at BEFORE UPDATE ON tenant_subnets
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_ip_usage_tracking_updated_at BEFORE UPDATE ON ip_usage_tracking
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Улучшенная функция для выделения IP с FOR UPDATE SKIP LOCKED
CREATE OR REPLACE FUNCTION allocate_ip_for_peer(
    p_tenant_id TEXT,
    p_peer_id TEXT,
    p_subnet_cidr CIDR
) RETURNS INET AS $$
DECLARE
    v_subnet_id INTEGER;
    v_ip INET;
BEGIN
    -- Получаем subnet_id с блокировкой
    SELECT id INTO v_subnet_id
    FROM tenant_subnets
    WHERE tenant_id = p_tenant_id AND subnet_cidr = p_subnet_cidr AND status='active'
    FOR UPDATE;

    IF v_subnet_id IS NULL THEN
        RAISE EXCEPTION 'Subnet % not found for tenant %', p_subnet_cidr, p_tenant_id;
    END IF;

    -- Проверяем, не выделен ли уже IP для этого peer
    SELECT ip_address INTO v_ip
    FROM tenant_ip_allocations
    WHERE tenant_id = p_tenant_id AND peer_id = p_peer_id AND status='active'
    LIMIT 1;
    
    IF FOUND THEN 
        RETURN v_ip; 
    END IF;

    -- Берем первый доступный IP с SKIP LOCKED для конкурентности
    SELECT ip_address INTO v_ip
    FROM ip_usage_tracking
    WHERE subnet_id = v_subnet_id AND status='available'
    ORDER BY ip_address
    FOR UPDATE SKIP LOCKED
    LIMIT 1;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'No available IPs for subnet %', p_subnet_cidr USING ERRCODE = 'resource_exhausted';
    END IF;

    -- Обновляем статус IP в tracking
    UPDATE ip_usage_tracking
    SET status='allocated', updated_at=now()
    WHERE subnet_id=v_subnet_id AND ip_address=v_ip;

    -- Создаем запись в allocations
    INSERT INTO tenant_ip_allocations(tenant_id, peer_id, ip_address, subnet_id, status, allocated_at)
    VALUES (p_tenant_id, p_peer_id, v_ip, v_subnet_id, 'active', now())
    ON CONFLICT (tenant_id, peer_id) DO UPDATE
       SET ip_address = EXCLUDED.ip_address, 
           status='active', 
           allocated_at=now();

    RETURN v_ip;
END;
$$ LANGUAGE plpgsql;

-- Улучшенная функция для освобождения IP
CREATE OR REPLACE FUNCTION deallocate_ip_for_peer(
    p_tenant_id TEXT,
    p_peer_id TEXT
) RETURNS BOOLEAN AS $$
DECLARE
    v_ip INET;
    v_subnet_id INTEGER;
BEGIN
    -- Получаем IP и subnet_id с блокировкой
    SELECT ip_address, subnet_id INTO v_ip, v_subnet_id
    FROM tenant_ip_allocations
    WHERE tenant_id=p_tenant_id AND peer_id=p_peer_id AND status='active'
    FOR UPDATE;
    
    IF NOT FOUND THEN 
        RETURN FALSE; 
    END IF;

    -- Обновляем статус в allocations
    UPDATE tenant_ip_allocations 
    SET status='released' 
    WHERE tenant_id=p_tenant_id AND peer_id=p_peer_id;
    
    -- Освобождаем IP в tracking
    UPDATE ip_usage_tracking 
    SET status='available', updated_at=now() 
    WHERE subnet_id=v_subnet_id AND ip_address=v_ip;
    
    RETURN TRUE;
END;
$$ LANGUAGE plpgsql;

-- Функция для инициализации IP пула в подсети
CREATE OR REPLACE FUNCTION initialize_subnet_ip_pool(
    p_subnet_id INTEGER,
    p_subnet_cidr CIDR
) RETURNS INTEGER AS $$
DECLARE
    base_ip INET;
    current_ip INET;
    ip_count INTEGER := 0;
BEGIN
    -- Получаем базовый IP подсети
    base_ip := network(p_subnet_cidr);
    
    -- Инициализируем IP адреса от .2 до .253
    FOR i IN 2..253 LOOP
        current_ip := base_ip + i;
        
        INSERT INTO ip_usage_tracking (subnet_id, ip_address, status)
        VALUES (p_subnet_id, current_ip, 'available')
        ON CONFLICT (subnet_id, ip_address) DO NOTHING;
        
        ip_count := ip_count + 1;
    END LOOP;
    
    RETURN ip_count;
END;
$$ LANGUAGE plpgsql;

-- Функция для получения статистики по тенанту
CREATE OR REPLACE FUNCTION get_tenant_ip_stats(p_tenant_id TEXT)
RETURNS TABLE(
    total_subnets BIGINT,
    allocated_ips BIGINT,
    available_ips BIGINT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        COUNT(DISTINCT ts.id) as total_subnets,
        COUNT(DISTINCT tia.id) FILTER (WHERE tia.status='active') as allocated_ips,
        COUNT(DISTINCT iut.ip_address) FILTER (WHERE iut.status='available') as available_ips
    FROM tenant_subnets ts
    LEFT JOIN tenant_ip_allocations tia ON ts.id = tia.subnet_id
    LEFT JOIN ip_usage_tracking iut ON ts.id = iut.subnet_id
    WHERE ts.tenant_id = p_tenant_id AND ts.status='active';
END;
$$ LANGUAGE plpgsql;

-- Функция для проверки коллизий подсетей
CREATE OR REPLACE FUNCTION check_subnet_collision(p_subnet_cidr CIDR)
RETURNS BOOLEAN AS $$
DECLARE
    collision_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO collision_count
    FROM tenant_subnets
    WHERE subnet_cidr && p_subnet_cidr AND status='active';
    
    RETURN collision_count > 0;
END;
$$ LANGUAGE plpgsql;

-- Создаем таблицу исключенных CIDR для избежания коллизий
CREATE TABLE IF NOT EXISTS excluded_cidrs (
    id SERIAL PRIMARY KEY,
    cidr CIDR NOT NULL UNIQUE,
    reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Добавляем некоторые стандартные исключения
INSERT INTO excluded_cidrs (cidr, reason) VALUES 
    ('10.0.0.0/8', 'RFC1918 private range'),
    ('172.16.0.0/12', 'RFC1918 private range'),
    ('192.168.0.0/16', 'RFC1918 private range'),
    ('100.64.0.0/10', 'CGNAT range')
ON CONFLICT (cidr) DO NOTHING;

-- Функция для проверки исключенных CIDR
CREATE OR REPLACE FUNCTION is_cidr_excluded(p_subnet_cidr CIDR)
RETURNS BOOLEAN AS $$
DECLARE
    exclusion_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO exclusion_count
    FROM excluded_cidrs
    WHERE p_subnet_cidr <<= cidr;
    
    RETURN exclusion_count > 0;
END;
$$ LANGUAGE plpgsql;

-- Комментарии к таблицам
COMMENT ON TABLE tenants IS 'Справочник тенантов';
COMMENT ON TABLE tenant_subnets IS 'Подсети, выделенные тенантам';
COMMENT ON TABLE tenant_ip_allocations IS 'Выделенные IP адреса для peer-ов';
COMMENT ON TABLE ip_usage_tracking IS 'Отслеживание использования IP адресов в подсетях';
COMMENT ON TABLE excluded_cidrs IS 'Исключенные CIDR диапазоны для избежания коллизий';


