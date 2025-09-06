# CloudBridge Relay - Domain Database Setup

## 📋 Обзор

Это руководство описывает настройку файлов инициализации домена Cadence для CloudBridge Relay в различных базах данных (Cassandra, MySQL, PostgreSQL).

## 🎯 Назначение Файлов

### Файлы в папке `domain/`
- **`cassandra.cql`** - Инициализация домена для Cassandra
- **`mysql.sql`** - Инициализация домена для MySQL  
- **`postgres.sql`** - Инициализация домена для PostgreSQL

## 🔧 Настройка для CloudBridge Relay

### Основные Параметры

#### Domain Configuration
```yaml
Domain Name: "cloudbridge"
Domain UUID: "550e8400-e29b-41d4-a716-446655440000"
Description: "CloudBridge Relay P2P Mesh and Workflow Management Domain"
Owner: "2GC Technologies"
Contact: "support@2gc.eu"
```

#### Workflow Configuration
```yaml
Retention: 30 days (production)
Metrics: Enabled
History Archival: Enabled
Visibility Archival: Enabled
Clusters: 
  - cloudbridge-cluster (primary)
  - cloudbridge-backup-cluster (backup)
```

---

## 🗄️ Cassandra Setup (cassandra.cql)

### Структура Домена
```cql
-- Основная таблица domains
INSERT INTO domains (
    id,
    domain
) VALUES (
    550e8400-e29b-41d4-a716-446655440000,
    {
        id: 550e8400-e29b-41d4-a716-446655440000,
        name: 'cloudbridge',
        status: 0, -- Registered
        description: 'CloudBridge Relay P2P Mesh and Workflow Management Domain',
        data: {
            'project': 'cloudbridge-relay',
            'version': '1.0.0',
            'environment': 'production',
            'features': ['p2p-mesh', 'wireguard', 'quic', 'workflow-engine'],
            'owner': '2GC Technologies',
            'contact': 'support@2gc.eu'
        },
        owner_email: 'support@2gc.eu'
    }
) IF NOT EXISTS;
```

### Конфигурация Workflow
```cql
-- Таблица domains_by_name_v2
INSERT INTO domains_by_name_v2 (
    domains_partition,
    name,
    domain,
    config,
    replication_config,
    is_global_domain,
    config_version,
    failover_version,
    failover_notification_version,
    notification_version
) VALUES (
    0,
    'cloudbridge',
    -- domain data (same as above)
    {
        retention: 30, -- 30 days retention for production
        emit_metric: True,
        history_archival_status: 1, -- Enabled for production
        visibility_archival_status: 1 -- Enabled for production
    },
    {
        active_cluster_name: 'cloudbridge-cluster',
        clusters: [
            {cluster_name: 'cloudbridge-cluster'},
            {cluster_name: 'cloudbridge-backup-cluster'}
        ]
    },
    True, -- is_global_domain
    1,    -- config_version
    0,    -- failover_version
    0,    -- failover_notification_version
    0     -- notification_version
) IF NOT EXISTS;
```

---

## 🗄️ MySQL Setup (mysql.sql)

### Структура Домена
```sql
-- Основная таблица domains
INSERT INTO domains (
    shard_id,
    id,
    name,
    data,
    data_encoding,
    is_global
) VALUES (
    54321, -- Default shard_id
    UNHEX(REPLACE('550e8400-e29b-41d4-a716-446655440000', '-', '')), -- CloudBridge Relay Domain UUID
    'cloudbridge', -- Domain name
    '{"project":"cloudbridge-relay","version":"1.0.0","environment":"production","features":["p2p-mesh","wireguard","quic","workflow-engine"],"owner":"2GC Technologies","contact":"support@2gc.eu"}', -- CloudBridge Relay data as JSON
    'json', -- Encoding type
    1 -- Set to 1 for a global domain
) ON DUPLICATE KEY UPDATE
    name = VALUES(name);
```

### Метаданные Домена
```sql
-- Таблица domain_metadata
INSERT INTO domain_metadata (
    notification_version
) VALUES (
    1
) ON DUPLICATE KEY UPDATE
    notification_version = VALUES(notification_version);
```

---

## 🗄️ PostgreSQL Setup (postgres.sql)

### Структура Домена
```sql
-- Основная таблица domains
INSERT INTO domains (
    shard_id,
    id,
    name,
    data,
    data_encoding,
    is_global
) VALUES (
    54321, -- Default shard_id
    '550e8400-e29b-41d4-a716-446655440000'::uuid, -- CloudBridge Relay Domain UUID
    'cloudbridge', -- Domain name
    '{"project":"cloudbridge-relay","version":"1.0.0","environment":"production","features":["p2p-mesh","wireguard","quic","workflow-engine"],"owner":"2GC Technologies","contact":"support@2gc.eu"}', -- CloudBridge Relay JSON data
    'json', -- Encoding type
    TRUE -- Set to TRUE for a global domain
) ON CONFLICT (shard_id, id) DO UPDATE
SET 
    name = EXCLUDED.name,
    data = EXCLUDED.data,
    data_encoding = EXCLUDED.data_encoding,
    is_global = EXCLUDED.is_global;
```

### Метаданные Домена
```sql
-- Таблица domain_metadata
INSERT INTO domain_metadata (
    id,
    notification_version
) VALUES (
    1, -- Default ID
    1 -- Notification version
) ON CONFLICT (id) DO UPDATE
SET 
    notification_version = EXCLUDED.notification_version;
```

---

## 🚀 Применение Настроек

### 1. Автоматическое Применение
Файлы применяются автоматически при запуске через `start.sh`:

```bash
# Для Cassandra
cqlsh $CASSANDRA_SEEDS -k $KEYSPACE -f /etc/cadence/domain/cassandra.cql

# Для MySQL
mysql -h $MYSQL_HOST -u $MYSQL_USER -p$MYSQL_PASSWORD $MYSQL_DATABASE < /etc/cadence/domain/mysql.sql

# Для PostgreSQL
psql -h $POSTGRES_HOST -U $POSTGRES_USER -d $POSTGRES_DATABASE -f /etc/cadence/domain/postgres.sql
```

### 2. Ручное Применение

#### Cassandra
```bash
# Подключение к Cassandra
cqlsh localhost

# Выполнение скрипта
USE cadence;
SOURCE '/path/to/domain/cassandra.cql';
```

#### MySQL
```bash
# Подключение к MySQL
mysql -h localhost -u root -p

# Выполнение скрипта
USE cadence;
SOURCE '/path/to/domain/mysql.sql';
```

#### PostgreSQL
```bash
# Подключение к PostgreSQL
psql -h localhost -U postgres -d cadence

# Выполнение скрипта
\i /path/to/domain/postgres.sql
```

---

## 🔍 Проверка Настройки

### 1. Проверка через Cadence CLI
```bash
# Список доменов
cadence --address localhost:7933 domain list

# Описание домена
cadence --address localhost:7933 domain describe cloudbridge
```

### 2. Проверка в Базе Данных

#### Cassandra
```bash
cqlsh -e "SELECT domain['name'], domain['status'] FROM cadence.domains WHERE domain['name'] = 'cloudbridge';"
```

#### MySQL
```bash
mysql -e "SELECT name, data FROM cadence.domains WHERE name = 'cloudbridge';"
```

#### PostgreSQL
```bash
psql -c "SELECT name, data FROM cadence.domains WHERE name = 'cloudbridge';"
```

---

## ⚙️ Настройка для Разных Сред

### Development Environment
```yaml
Domain Name: "cloudbridge-dev"
Retention: 1 day
Metrics: Disabled
Archival: Disabled
Clusters: ["cloudbridge-dev-cluster"]
```

### Staging Environment
```yaml
Domain Name: "cloudbridge-staging"
Retention: 7 days
Metrics: Enabled
Archival: Disabled
Clusters: ["cloudbridge-staging-cluster"]
```

### Production Environment
```yaml
Domain Name: "cloudbridge"
Retention: 30 days
Metrics: Enabled
Archival: Enabled
Clusters: ["cloudbridge-cluster", "cloudbridge-backup-cluster"]
```

---

## 🚨 Troubleshooting

### Частые Проблемы

#### 1. Домен не создается
```bash
# Проверка подключения к базе данных
# Cassandra
cqlsh localhost -e "DESCRIBE KEYSPACES;"

# MySQL
mysql -e "SHOW DATABASES;"

# PostgreSQL
psql -c "\l"
```

#### 2. Ошибки в UUID
```bash
# Генерация нового UUID
python3 -c "import uuid; print(uuid.uuid4())"

# Замена в файлах
sed -i 's/550e8400-e29b-41d4-a716-446655440000/NEW_UUID_HERE/g' domain/*.sql
```

#### 3. Проблемы с JSON
```bash
# Валидация JSON
echo '{"key":"value"}' | python3 -m json.tool

# Проверка синтаксиса
jq . domain/mysql.sql
```

---

## 📊 Мониторинг

### Метрики Домена
- `cadence_domain_count` - Количество доменов
- `cadence_workflow_count` - Количество workflow
- `cadence_activity_count` - Количество активностей

### Логи
```bash
# Логи Cadence
tail -f /var/log/cadence/cadence.log

# Логи базы данных
# Cassandra
tail -f /var/log/cassandra/system.log

# MySQL
tail -f /var/log/mysql/error.log

# PostgreSQL
tail -f /var/log/postgresql/postgresql.log
```

---

## 🔗 Связанные Файлы

- `configs/cadence.yaml` - Конфигурация Cadence
- `scripts/init-cadence.sh` - Скрипт инициализации
- `start.sh` - Основной скрипт запуска
- `internal/cadence/workflows/` - Workflow определения

---

## 📝 Заключение

Правильная настройка файлов инициализации домена критически важна для:
- ✅ Корректной работы Cadence Workflow Engine
- ✅ Управления P2P mesh и connection handling
- ✅ Обеспечения высокой доступности
- ✅ Мониторинга и архивирования workflow

**Важно:** Обновите UUID и конфигурацию в соответствии с вашей средой перед применением.

---

*Документация обновлена: 6 сентября 2025*

