# Cadence Domain Initialization

## 📋 Обзор

Файл `domain/cassandra.cql` отвечает за **инициализацию домена Cadence Workflow Engine** в базе данных Cassandra. Это критически важный компонент для работы CloudBridge Relay с workflow-процессами.

## 🎯 Назначение

### Что такое Cadence Domain?
**Cadence Domain** - это логическое пространство имен для workflow-процессов, которое обеспечивает:
- Изоляцию workflow между разными приложениями
- Управление конфигурацией workflow
- Метаданные и настройки домена
- Версионирование и failover

### Зачем нужен этот файл?
Файл `domain/cassandra.cql` создает:
1. **Базовый домен** с именем "default"
2. **Метаданные домена** для управления workflow
3. **Конфигурацию репликации** для высокой доступности
4. **Настройки архивирования** для истории workflow

---

## 🗄️ Структура Базы Данных

### Таблица `domains`
```sql
INSERT INTO domains (
    id,
    domain
) VALUES (
    123e4567-e89b-12d3-a456-426614174000,
    {
        id: 123e4567-e89b-12d3-a456-426614174000,
        name: 'default',
        status: 0, -- Registered
        description: 'This is an example domain.',
        data: {'key1': 'value1', 'key2': 'value2'},
        owner_email: 'owner@example.com'
    }
) IF NOT EXISTS;
```

**Поля:**
- `id` - UUID домена
- `domain` - Структурированные данные домена
  - `name` - Имя домена
  - `status` - Статус (0 = Registered)
  - `description` - Описание домена
  - `data` - Дополнительные данные
  - `owner_email` - Email владельца

### Таблица `domains_by_name_v2`
```sql
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
) VALUES (...)
```

**Поля:**
- `domains_partition` - Партиция для шардинга
- `name` - Имя домена для поиска
- `domain` - Данные домена
- `config` - Конфигурация workflow
- `replication_config` - Настройки репликации
- `is_global_domain` - Глобальный домен
- `*_version` - Версии для управления изменениями

---

## ⚙️ Конфигурация

### Workflow Configuration
```cql
config: {
    retention: 7,                    -- Дни хранения истории
    emit_metric: True,               -- Включить метрики
    history_archival_status: 0,      -- Статус архивирования истории
    visibility_archival_status: 0    -- Статус архивирования видимости
}
```

### Replication Configuration
```cql
replication_config: {
    active_cluster_name: 'cluster0', -- Активный кластер
    clusters: [{cluster_name: 'cluster0'}] -- Список кластеров
}
```

---

## 🔄 Процесс Инициализации

### 1. Автоматическая Инициализация
Файл выполняется автоматически при запуске Cadence сервера через скрипт `start.sh`:

```bash
# В функции setup_cassandra_schema()
echo "Registering domain for Cassandra..."
cqlsh $CASSANDRA_SEEDS -k $KEYSPACE -f /etc/cadence/domain/cassandra.cql
```

### 2. Ручная Инициализация
```bash
# Подключение к Cassandra
cqlsh localhost

# Выполнение скрипта
USE cadence;
SOURCE '/path/to/domain/cassandra.cql';
```

### 3. Проверка Инициализации
```bash
# Проверка через Cadence CLI
cadence --address localhost:7933 domain describe default

# Проверка в базе данных
cqlsh -e "SELECT * FROM cadence.domains WHERE domain['name'] = 'default';"
```

---

## 🏗️ Интеграция с CloudBridge Relay

### Workflow Процессы
CloudBridge Relay использует Cadence для:
- **P2P Mesh Management** - Управление P2P сетью
- **Connection Handling** - Обработка подключений
- **Load Balancing** - Балансировка нагрузки
- **Failover Management** - Управление отказоустойчивостью

### Пример Workflow
```go
// В internal/cadence/workflows/
type RelayWorkflow struct {
    // Workflow для управления P2P соединениями
}

func (w *RelayWorkflow) Execute(ctx workflow.Context) error {
    // Логика workflow
    return nil
}
```

---

## 🔧 Настройка для Разных Сред

### Development
```cql
-- Минимальная конфигурация
retention: 1,
emit_metric: False,
history_archival_status: 0,
visibility_archival_status: 0
```

### Staging
```cql
-- Средняя конфигурация
retention: 3,
emit_metric: True,
history_archival_status: 0,
visibility_archival_status: 0
```

### Production
```cql
-- Полная конфигурация
retention: 30,
emit_metric: True,
history_archival_status: 1,  -- Включено
visibility_archival_status: 1 -- Включено
```

---

## 🚨 Troubleshooting

### Частые Проблемы

#### 1. Домен не создается
```bash
# Проверка подключения к Cassandra
cqlsh localhost -e "DESCRIBE KEYSPACES;"

# Проверка прав доступа
cqlsh localhost -e "SELECT * FROM system_auth.roles;"
```

#### 2. Ошибки в CQL
```bash
# Проверка синтаксиса
cqlsh localhost -e "SOURCE '/path/to/cassandra.cql';"

# Просмотр ошибок
tail -f /var/log/cassandra/system.log
```

#### 3. Проблемы с UUID
```bash
# Генерация нового UUID
python3 -c "import uuid; print(uuid.uuid4())"

# Замена в файле
sed -i 's/123e4567-e89b-12d3-a456-426614174000/NEW_UUID_HERE/g' domain/cassandra.cql
```

---

## 📊 Мониторинг

### Проверка Статуса Домена
```bash
# Через Cadence CLI
cadence --address localhost:7933 domain list
cadence --address localhost:7933 domain describe default

# Через базу данных
cqlsh -e "SELECT domain['name'], domain['status'] FROM cadence.domains;"
```

### Метрики
- `cadence_domain_count` - Количество доменов
- `cadence_workflow_count` - Количество workflow
- `cadence_activity_count` - Количество активностей

---

## 🔗 Связанные Файлы

- `domain/mysql.sql` - Инициализация для MySQL
- `domain/postgres.sql` - Инициализация для PostgreSQL
- `scripts/init-cadence.sh` - Скрипт инициализации
- `start.sh` - Основной скрипт запуска
- `internal/cadence/workflows/` - Workflow определения

---

## 📝 Заключение

Файл `domain/cassandra.cql` является **критически важным** для:
- ✅ Инициализации Cadence Workflow Engine
- ✅ Создания базового домена для CloudBridge Relay
- ✅ Настройки конфигурации workflow
- ✅ Обеспечения работы P2P mesh и connection handling

**Важно:** Этот файл должен быть выполнен **до** запуска CloudBridge Relay сервера, иначе workflow-процессы не будут работать.

---

*Документация обновлена: 6 сентября 2025*
