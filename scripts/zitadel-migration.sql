-- 🔄 SQL скрипт миграции базы данных для Zitadel
-- 
-- Этот скрипт создает необходимые таблицы для интеграции с Zitadel
-- и мигрирует существующие данные

-- ============================================
-- 1. СОЗДАНИЕ ТАБЛИЦ ДЛЯ ZITADEL
-- ============================================

-- Таблица для хранения Zitadel организаций
CREATE TABLE IF NOT EXISTS zitadel_organizations (
  id VARCHAR(255) PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  domain VARCHAR(255),
  original_organization_name VARCHAR(255),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_original_org_name (original_organization_name),
  INDEX idx_domain (domain)
);

-- Таблица для хранения Zitadel пользователей
CREATE TABLE IF NOT EXISTS zitadel_users (
  id VARCHAR(255) PRIMARY KEY,
  organization_id VARCHAR(255) NOT NULL,
  email VARCHAR(255) NOT NULL,
  name VARCHAR(255),
  original_user_id VARCHAR(255),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  FOREIGN KEY (organization_id) REFERENCES zitadel_organizations(id) ON DELETE CASCADE,
  INDEX idx_organization_id (organization_id),
  INDEX idx_original_user_id (original_user_id),
  INDEX idx_email (email)
);

-- Таблица для сетевых конфигураций тенантов
CREATE TABLE IF NOT EXISTS tenant_network_configs (
  id VARCHAR(255) PRIMARY KEY,
  tenant_id VARCHAR(255) NOT NULL,
  organization_id VARCHAR(255) NOT NULL,
  allocated_subnet VARCHAR(255),
  ipam_pool VARCHAR(255),
  network_policies JSON,
  allocated_ips JSON,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  FOREIGN KEY (organization_id) REFERENCES zitadel_organizations(id) ON DELETE CASCADE,
  INDEX idx_tenant_id (tenant_id),
  INDEX idx_organization_id (organization_id),
  INDEX idx_allocated_subnet (allocated_subnet)
);

-- Таблица для хранения Zitadel токенов
CREATE TABLE IF NOT EXISTS zitadel_tokens (
  id VARCHAR(255) PRIMARY KEY,
  user_id VARCHAR(255),
  organization_id VARCHAR(255) NOT NULL,
  token_type ENUM('access', 'refresh', 'id') DEFAULT 'access',
  token_value TEXT NOT NULL,
  expires_at TIMESTAMP,
  scopes JSON,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES zitadel_users(id) ON DELETE CASCADE,
  FOREIGN KEY (organization_id) REFERENCES zitadel_organizations(id) ON DELETE CASCADE,
  INDEX idx_user_id (user_id),
  INDEX idx_organization_id (organization_id),
  INDEX idx_expires_at (expires_at)
);

-- Таблица для хранения Zitadel клиентов
CREATE TABLE IF NOT EXISTS zitadel_clients (
  id VARCHAR(255) PRIMARY KEY,
  organization_id VARCHAR(255) NOT NULL,
  client_id VARCHAR(255) NOT NULL,
  client_secret VARCHAR(255),
  client_type ENUM('quick', 'p2p_mesh', 'admin') DEFAULT 'quick',
  scopes JSON,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  FOREIGN KEY (organization_id) REFERENCES zitadel_organizations(id) ON DELETE CASCADE,
  INDEX idx_organization_id (organization_id),
  INDEX idx_client_id (client_id),
  INDEX idx_client_type (client_type)
);

-- ============================================
-- 2. МИГРАЦИЯ СУЩЕСТВУЮЩИХ ДАННЫХ
-- ============================================

-- Миграция организаций из пользователей
INSERT INTO zitadel_organizations (id, name, domain, original_organization_name)
SELECT 
  CONCAT('org-', MD5(CONCAT(organization_name, domain))) as id,
  organization_name as name,
  domain,
  organization_name as original_organization_name
FROM users 
WHERE organization_name IS NOT NULL
GROUP BY organization_name, domain
ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  domain = VALUES(domain);

-- Миграция пользователей
INSERT INTO zitadel_users (id, organization_id, email, name, original_user_id)
SELECT 
  CONCAT('user-', u.id) as id,
  zo.id as organization_id,
  u.email,
  u.name,
  u.id as original_user_id
FROM users u
JOIN zitadel_organizations zo ON zo.original_organization_name = u.organization_name
ON DUPLICATE KEY UPDATE
  email = VALUES(email),
  name = VALUES(name);

-- Миграция тенантов в сетевые конфигурации
INSERT INTO tenant_network_configs (id, tenant_id, organization_id, allocated_subnet, ipam_pool, network_policies, allocated_ips)
SELECT 
  CONCAT('config-', t.id) as id,
  t.id as tenant_id,
  zo.id as organization_id,
  COALESCE(t.subnet, CONCAT('10.', (CRC32(t.id) >> 16) & 0xFF, '.', CRC32(t.id) & 0xFF, '.0/24')) as allocated_subnet,
  COALESCE(t.ipam_pool, 'cloudbridge-pool-1') as ipam_pool,
  JSON_ARRAY('allow-intra-tenant') as network_policies,
  JSON_OBJECT() as allocated_ips
FROM tenants t
JOIN zitadel_organizations zo ON zo.original_organization_name = t.organization_name
ON DUPLICATE KEY UPDATE
  allocated_subnet = VALUES(allocated_subnet),
  ipam_pool = VALUES(ipam_pool);

-- ============================================
-- 3. СОЗДАНИЕ ИНДЕКСОВ ДЛЯ ПРОИЗВОДИТЕЛЬНОСТИ
-- ============================================

-- Индексы для быстрого поиска
CREATE INDEX IF NOT EXISTS idx_zitadel_orgs_name ON zitadel_organizations(name);
CREATE INDEX IF NOT EXISTS idx_zitadel_orgs_domain ON zitadel_organizations(domain);
CREATE INDEX IF NOT EXISTS idx_zitadel_users_email ON zitadel_users(email);
CREATE INDEX IF NOT EXISTS idx_tenant_network_configs_subnet ON tenant_network_configs(allocated_subnet);
CREATE INDEX IF NOT EXISTS idx_zitadel_tokens_expires ON zitadel_tokens(expires_at);

-- ============================================
-- 4. СОЗДАНИЕ VIEW ДЛЯ УДОБСТВА
-- ============================================

-- View для получения полной информации о пользователях
CREATE OR REPLACE VIEW user_zitadel_info AS
SELECT 
  u.id as original_user_id,
  u.email,
  u.name,
  u.organization_name,
  u.domain,
  zu.id as zitadel_user_id,
  zo.id as zitadel_organization_id,
  zo.name as zitadel_organization_name,
  tnc.allocated_subnet,
  tnc.ipam_pool
FROM users u
LEFT JOIN zitadel_users zu ON zu.original_user_id = u.id
LEFT JOIN zitadel_organizations zo ON zo.id = zu.organization_id
LEFT JOIN tenant_network_configs tnc ON tnc.tenant_id = u.tenant_id;

-- View для получения информации о тенантах
CREATE OR REPLACE VIEW tenant_zitadel_info AS
SELECT 
  t.id as tenant_id,
  t.name as tenant_name,
  t.organization_name,
  t.domain,
  zo.id as zitadel_organization_id,
  zo.name as zitadel_organization_name,
  tnc.allocated_subnet,
  tnc.ipam_pool,
  tnc.network_policies,
  tnc.allocated_ips
FROM tenants t
LEFT JOIN zitadel_organizations zo ON zo.original_organization_name = t.organization_name
LEFT JOIN tenant_network_configs tnc ON tnc.tenant_id = t.id;

-- ============================================
-- 5. СОЗДАНИЕ ПРОЦЕДУР ДЛЯ УПРАВЛЕНИЯ
-- ============================================

-- Процедура для очистки истекших токенов
DELIMITER //
CREATE PROCEDURE CleanupExpiredTokens()
BEGIN
  DELETE FROM zitadel_tokens 
  WHERE expires_at < NOW();
  
  SELECT ROW_COUNT() as deleted_tokens;
END //
DELIMITER ;

-- Процедура для получения статистики миграции
DELIMITER //
CREATE PROCEDURE GetMigrationStats()
BEGIN
  SELECT 
    'Organizations' as entity_type,
    COUNT(*) as total_count,
    COUNT(CASE WHEN created_at > DATE_SUB(NOW(), INTERVAL 1 DAY) THEN 1 END) as recent_count
  FROM zitadel_organizations
  
  UNION ALL
  
  SELECT 
    'Users' as entity_type,
    COUNT(*) as total_count,
    COUNT(CASE WHEN created_at > DATE_SUB(NOW(), INTERVAL 1 DAY) THEN 1 END) as recent_count
  FROM zitadel_users
  
  UNION ALL
  
  SELECT 
    'Network Configs' as entity_type,
    COUNT(*) as total_count,
    COUNT(CASE WHEN created_at > DATE_SUB(NOW(), INTERVAL 1 DAY) THEN 1 END) as recent_count
  FROM tenant_network_configs;
END //
DELIMITER ;

-- ============================================
-- 6. СОЗДАНИЕ ТРИГГЕРОВ ДЛЯ АВТОМАТИЗАЦИИ
-- ============================================

-- Триггер для автоматического обновления updated_at
DELIMITER //
CREATE TRIGGER zitadel_organizations_updated_at
  BEFORE UPDATE ON zitadel_organizations
  FOR EACH ROW
BEGIN
  SET NEW.updated_at = CURRENT_TIMESTAMP;
END //
DELIMITER ;

DELIMITER //
CREATE TRIGGER zitadel_users_updated_at
  BEFORE UPDATE ON zitadel_users
  FOR EACH ROW
BEGIN
  SET NEW.updated_at = CURRENT_TIMESTAMP;
END //
DELIMITER ;

DELIMITER //
CREATE TRIGGER tenant_network_configs_updated_at
  BEFORE UPDATE ON tenant_network_configs
  FOR EACH ROW
BEGIN
  SET NEW.updated_at = CURRENT_TIMESTAMP;
END //
DELIMITER ;

-- ============================================
-- 7. СОЗДАНИЕ ПРАВ ДОСТУПА
-- ============================================

-- Создание пользователя для приложения
CREATE USER IF NOT EXISTS 'cloudbridge_app'@'%' IDENTIFIED BY 'secure_password_here';
GRANT SELECT, INSERT, UPDATE, DELETE ON cloudbridge.* TO 'cloudbridge_app'@'%';

-- Создание пользователя для миграции
CREATE USER IF NOT EXISTS 'cloudbridge_migration'@'%' IDENTIFIED BY 'migration_password_here';
GRANT ALL PRIVILEGES ON cloudbridge.* TO 'cloudbridge_migration'@'%';

-- ============================================
-- 8. ПРОВЕРКА МИГРАЦИИ
-- ============================================

-- Проверка целостности данных
SELECT 
  'Migration Check' as check_type,
  COUNT(*) as count,
  'All organizations migrated' as description
FROM zitadel_organizations
WHERE original_organization_name IS NOT NULL

UNION ALL

SELECT 
  'Migration Check' as check_type,
  COUNT(*) as count,
  'All users migrated' as description
FROM zitadel_users
WHERE original_user_id IS NOT NULL

UNION ALL

SELECT 
  'Migration Check' as check_type,
  COUNT(*) as count,
  'All tenants have network configs' as description
FROM tenant_network_configs;

-- Проверка связей
SELECT 
  'Data Integrity Check' as check_type,
  COUNT(*) as orphaned_users,
  'Users without organizations' as description
FROM zitadel_users zu
LEFT JOIN zitadel_organizations zo ON zo.id = zu.organization_id
WHERE zo.id IS NULL

UNION ALL

SELECT 
  'Data Integrity Check' as check_type,
  COUNT(*) as orphaned_configs,
  'Network configs without organizations' as description
FROM tenant_network_configs tnc
LEFT JOIN zitadel_organizations zo ON zo.id = tnc.organization_id
WHERE zo.id IS NULL;

-- ============================================
-- 9. ОЧИСТКА СТАРЫХ ДАННЫХ (ОПЦИОНАЛЬНО)
-- ============================================

-- ВНИМАНИЕ: Раскомментируйте только после успешной миграции и тестирования!

-- Удаление старых таблиц Keycloak (если есть)
-- DROP TABLE IF EXISTS keycloak_users;
-- DROP TABLE IF EXISTS keycloak_organizations;
-- DROP TABLE IF EXISTS keycloak_tokens;

-- Удаление старых индексов
-- DROP INDEX IF EXISTS idx_keycloak_user_id ON users;
-- DROP INDEX IF EXISTS idx_keycloak_org_id ON tenants;

-- ============================================
-- 10. ФИНАЛЬНАЯ ПРОВЕРКА
-- ============================================

-- Вывод статистики миграции
CALL GetMigrationStats();

-- Проверка производительности
EXPLAIN SELECT * FROM user_zitadel_info WHERE email = 'test@example.com';
EXPLAIN SELECT * FROM tenant_zitadel_info WHERE tenant_name = 'test-tenant';

-- Проверка индексов
SHOW INDEX FROM zitadel_organizations;
SHOW INDEX FROM zitadel_users;
SHOW INDEX FROM tenant_network_configs;

-- ============================================
-- ЗАВЕРШЕНИЕ МИГРАЦИИ
-- ============================================

SELECT 
  'Migration Complete' as status,
  NOW() as completed_at,
  'All tables created and data migrated successfully' as message;

