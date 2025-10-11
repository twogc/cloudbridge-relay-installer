# IPAM Database Migrations

Этот каталог содержит SQL миграции для системы IPAM (IP Address Management).

## Структура миграций

### 001_tenant_ipam_schema.sql
**Старая схема** - базовая версия с VARCHAR типами и простыми функциями.

### 002_tenant_ipam_improved_schema.sql  
**Новая production-ready схема** с улучшениями:

#### Ключевые изменения:
- **Правильные типы данных**: `CIDR`, `INET` вместо `VARCHAR`
- **Улучшенная структура**: `ip_usage_tracking` привязан к `subnet_id`
- **Конкурентные SQL функции**: `FOR UPDATE SKIP LOCKED`
- **Проверка коллизий**: таблица `excluded_cidrs`
- **Статистика**: функция `get_tenant_ip_stats()`

#### Новые функции:
```sql
-- Выделение IP с конкурентностью
allocate_ip_for_peer(tenant_id, peer_id, subnet_cidr) RETURNS INET

-- Освобождение IP
deallocate_ip_for_peer(tenant_id, peer_id) RETURNS BOOLEAN

-- Инициализация IP пула
initialize_subnet_ip_pool(subnet_id, subnet_cidr) RETURNS INTEGER

-- Статистика по тенанту
get_tenant_ip_stats(tenant_id) RETURNS TABLE(...)

-- Проверка коллизий
check_subnet_collision(subnet_cidr) RETURNS BOOLEAN
is_cidr_excluded(subnet_cidr) RETURNS BOOLEAN
```

### 003_test_sql_functions.sql
**Тестовые данные** для проверки функций IPAM.

## Применение миграций

### Для нового развертывания:
```bash
# Применить новую схему
psql -d ipam_db -f migrations/002_tenant_ipam_improved_schema.sql

# Запустить тесты
psql -d ipam_db -f migrations/003_test_sql_functions.sql
```

### Для обновления существующей БД:
```bash
# 1. Создать бэкап
pg_dump ipam_db > backup_before_migration.sql

# 2. Применить новую схему (осторожно!)
psql -d ipam_db -f migrations/002_tenant_ipam_improved_schema.sql

# 3. Проверить работоспособность
psql -d ipam_db -f migrations/003_test_sql_functions.sql
```

## Схема БД

### Основные таблицы:
- `tenants` - справочник тенантов
- `tenant_subnets` - подсети тенантов  
- `tenant_ip_allocations` - выделенные IP
- `ip_usage_tracking` - отслеживание использования IP
- `excluded_cidrs` - исключенные CIDR диапазоны

### Ключевые особенности:
- **Транзакционность**: все операции в транзакциях
- **Конкурентность**: `FOR UPDATE SKIP LOCKED` для аллокации
- **Коллизии**: проверка пересечений подсетей
- **Статистика**: встроенные функции для мониторинга
- **Аудит**: автоматические `updated_at` триггеры

## Производительность

### Индексы:
- `idx_tenant_subnets_tenant_id` - поиск по тенанту
- `idx_tenant_subnets_subnet_cidr` - поиск по CIDR
- `idx_tenant_ip_allocations_tenant_id` - аллокации по тенанту
- `idx_ip_usage_tracking_subnet_id` - IP по подсети
- `idx_ip_usage_tracking_status` - доступные IP

### Оптимизации:
- **Составные индексы** для частых запросов
- **Партиционирование** по `tenant_id` (для больших объемов)
- **Кэширование** статистики в Redis (опционально)

## Мониторинг

### Ключевые метрики:
```sql
-- Общая статистика
SELECT * FROM get_tenant_ip_stats('tenant-id');

-- Использование подсетей
SELECT 
    tenant_id,
    subnet_cidr,
    COUNT(*) as total_ips,
    COUNT(*) FILTER (WHERE status='allocated') as allocated,
    COUNT(*) FILTER (WHERE status='available') as available
FROM tenant_subnets ts
JOIN ip_usage_tracking iut ON ts.id = iut.subnet_id
GROUP BY tenant_id, subnet_cidr;

-- Коллизии подсетей
SELECT 
    subnet_cidr,
    check_subnet_collision(subnet_cidr) as has_collision
FROM tenant_subnets;
```

## Безопасность

### Ограничения:
- **Уникальность**: `(tenant_id, peer_id)`, `(subnet_id, ip_address)`
- **Каскадное удаление**: при удалении тенанта/подсети
- **Проверки**: статусы, типы данных, ограничения

### Аудит:
- **Логирование**: все операции через application layer
- **Метрики**: Prometheus для мониторинга
- **Алерты**: на коллизии, исчерпание IP, ошибки

## Troubleshooting

### Частые проблемы:

1. **Коллизии подсетей**:
   ```sql
   SELECT * FROM excluded_cidrs WHERE cidr && '10.0.0.0/8';
   ```

2. **Исчерпание IP**:
   ```sql
   SELECT COUNT(*) FROM ip_usage_tracking WHERE status='available';
   ```

3. **Зависшие аллокации**:
   ```sql
   SELECT * FROM tenant_ip_allocations WHERE status='active' AND allocated_at < NOW() - INTERVAL '1 hour';
   ```

### Восстановление:
```sql
-- Сброс статуса IP
UPDATE ip_usage_tracking SET status='available' WHERE status='allocated';

-- Очистка старых аллокаций
DELETE FROM tenant_ip_allocations WHERE allocated_at < NOW() - INTERVAL '24 hours';
```


