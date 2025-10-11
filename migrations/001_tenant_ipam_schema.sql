-- Схема БД для IPAM с транзакционностью и проверкой коллизий
-- Миграция 001: Создание таблиц для tenant IPAM

-- Таблица тенантов
CREATE TABLE IF NOT EXISTS tenants (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    status VARCHAR(50) DEFAULT 'active',
    metadata JSONB DEFAULT '{}'::jsonb
);

-- Таблица подсетей тенантов
CREATE TABLE IF NOT EXISTS tenant_subnets (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    subnet_cidr VARCHAR(18) NOT NULL UNIQUE, -- 10.x.y.z/24
    gateway VARCHAR(15) NOT NULL, -- 10.x.y.1
    ip_range VARCHAR(18) NOT NULL, -- 10.x.y.0/24
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    status VARCHAR(50) DEFAULT 'active',
    metadata JSONB DEFAULT '{}'::jsonb,
    
    -- Индексы для быстрого поиска
    CONSTRAINT unique_tenant_subnet UNIQUE (tenant_id, subnet_cidr)
);

-- Таблица выделенных IP адресов
CREATE TABLE IF NOT EXISTS tenant_ip_allocations (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    peer_id VARCHAR(255) NOT NULL,
    ip_address VARCHAR(15) NOT NULL,
    subnet_id INTEGER NOT NULL REFERENCES tenant_subnets(id) ON DELETE CASCADE,
    allocated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,
    status VARCHAR(50) DEFAULT 'active',
    metadata JSONB DEFAULT '{}'::jsonb,
    
    -- Уникальные ограничения
    CONSTRAINT unique_tenant_peer UNIQUE (tenant_id, peer_id),
    CONSTRAINT unique_tenant_ip UNIQUE (tenant_id, ip_address),
    CONSTRAINT unique_ip_global UNIQUE (ip_address)
);

-- Таблица для отслеживания использования IP адресов
CREATE TABLE IF NOT EXISTS ip_usage_tracking (
    id SERIAL PRIMARY KEY,
    ip_address VARCHAR(15) NOT NULL UNIQUE,
    tenant_id VARCHAR(255),
    peer_id VARCHAR(255),
    allocated_at TIMESTAMP WITH TIME ZONE,
    released_at TIMESTAMP WITH TIME ZONE,
    status VARCHAR(50) DEFAULT 'available', -- available, allocated, reserved
    metadata JSONB DEFAULT '{}'::jsonb
);

-- Индексы для производительности
CREATE INDEX IF NOT EXISTS idx_tenant_subnets_tenant_id ON tenant_subnets(tenant_id);
CREATE INDEX IF NOT EXISTS idx_tenant_subnets_subnet_cidr ON tenant_subnets(subnet_cidr);
CREATE INDEX IF NOT EXISTS idx_tenant_ip_allocations_tenant_id ON tenant_ip_allocations(tenant_id);
CREATE INDEX IF NOT EXISTS idx_tenant_ip_allocations_peer_id ON tenant_ip_allocations(peer_id);
CREATE INDEX IF NOT EXISTS idx_tenant_ip_allocations_ip_address ON tenant_ip_allocations(ip_address);
CREATE INDEX IF NOT EXISTS idx_ip_usage_tracking_ip_address ON ip_usage_tracking(ip_address);
CREATE INDEX IF NOT EXISTS idx_ip_usage_tracking_status ON ip_usage_tracking(status);

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

-- Функция для проверки доступности IP адреса
CREATE OR REPLACE FUNCTION is_ip_available(ip_addr VARCHAR(15))
RETURNS BOOLEAN AS $$
DECLARE
    usage_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO usage_count
    FROM ip_usage_tracking
    WHERE ip_address = ip_addr AND status = 'allocated';
    
    RETURN usage_count = 0;
END;
$$ LANGUAGE plpgsql;

-- Функция для выделения IP адреса с проверкой коллизий
CREATE OR REPLACE FUNCTION allocate_ip_for_peer(
    p_tenant_id VARCHAR(255),
    p_peer_id VARCHAR(255),
    p_subnet_cidr VARCHAR(18)
) RETURNS VARCHAR(15) AS $$
DECLARE
    v_subnet_id INTEGER;
    v_gateway VARCHAR(15);
    v_ip_range VARCHAR(18);
    v_available_ip VARCHAR(15);
    v_allocated_ip VARCHAR(15);
BEGIN
    -- Начинаем транзакцию
    BEGIN
        -- Получаем информацию о подсети
        SELECT id, gateway, ip_range INTO v_subnet_id, v_gateway, v_ip_range
        FROM tenant_subnets
        WHERE tenant_id = p_tenant_id AND subnet_cidr = p_subnet_cidr;
        
        IF v_subnet_id IS NULL THEN
            RAISE EXCEPTION 'Subnet % not found for tenant %', p_subnet_cidr, p_tenant_id;
        END IF;
        
        -- Проверяем, не выделен ли уже IP для этого peer
        SELECT ip_address INTO v_allocated_ip
        FROM tenant_ip_allocations
        WHERE tenant_id = p_tenant_id AND peer_id = p_peer_id;
        
        IF v_allocated_ip IS NOT NULL THEN
            RETURN v_allocated_ip;
        END IF;
        
        -- Ищем доступный IP в подсети
        SELECT ip_address INTO v_available_ip
        FROM ip_usage_tracking
        WHERE ip_address LIKE split_part(p_subnet_cidr, '/', 1) || '%'
        AND status = 'available'
        AND is_ip_available(ip_address)
        LIMIT 1
        FOR UPDATE SKIP LOCKED;
        
        IF v_available_ip IS NULL THEN
            RAISE EXCEPTION 'No available IP addresses in subnet %', p_subnet_cidr;
        END IF;
        
        -- Выделяем IP
        INSERT INTO tenant_ip_allocations (tenant_id, peer_id, ip_address, subnet_id)
        VALUES (p_tenant_id, p_peer_id, v_available_ip, v_subnet_id);
        
        -- Обновляем статус в tracking таблице
        UPDATE ip_usage_tracking
        SET status = 'allocated', tenant_id = p_tenant_id, peer_id = p_peer_id, allocated_at = NOW()
        WHERE ip_address = v_available_ip;
        
        RETURN v_available_ip;
        
    EXCEPTION
        WHEN unique_violation THEN
            RAISE EXCEPTION 'IP allocation conflict for tenant % peer %', p_tenant_id, p_peer_id;
        WHEN OTHERS THEN
            RAISE;
    END;
END;
$$ LANGUAGE plpgsql;

-- Функция для освобождения IP адреса
CREATE OR REPLACE FUNCTION deallocate_ip_for_peer(
    p_tenant_id VARCHAR(255),
    p_peer_id VARCHAR(255)
) RETURNS BOOLEAN AS $$
DECLARE
    v_ip_address VARCHAR(15);
BEGIN
    -- Получаем IP адрес для освобождения
    SELECT ip_address INTO v_ip_address
    FROM tenant_ip_allocations
    WHERE tenant_id = p_tenant_id AND peer_id = p_peer_id;
    
    IF v_ip_address IS NULL THEN
        RETURN FALSE;
    END IF;
    
    -- Удаляем из allocations
    DELETE FROM tenant_ip_allocations
    WHERE tenant_id = p_tenant_id AND peer_id = p_peer_id;
    
    -- Обновляем статус в tracking таблице
    UPDATE ip_usage_tracking
    SET status = 'available', tenant_id = NULL, peer_id = NULL, released_at = NOW()
    WHERE ip_address = v_ip_address;
    
    RETURN TRUE;
END;
$$ LANGUAGE plpgsql;


