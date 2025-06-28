# CloudBridge Relay - Keycloak Automation Scripts

Этот каталог содержит скрипты для автоматизации настройки и управления Keycloak для CloudBridge Relay.

## Скрипты

### 1. `setup_keycloak.sh` - Основная настройка Keycloak
Автоматически настраивает Keycloak для работы с CloudBridge Relay.

**Использование:**
```bash
# Базовая настройка
./scripts/setup_keycloak.sh

# Настройка с кастомными параметрами
KEYCLOAK_URL=https://keycloak.company.com \
REALM_NAME=cloudbridge-prod \
CLIENT_ID=relay-client-prod \
./scripts/setup_keycloak.sh
```

**Что делает:**
- Создает realm `cloudbridge`
- Создает клиент `relay-client`
- Создает роли и группы
- Создает базовых пользователей
- Генерирует конфигурационный файл

### 2. `create_users.sh` - Создание пользователей
Создает пользователей в Keycloak с соответствующими ролями и атрибутами.

**Использование:**
```bash
# Создание стандартных пользователей
./scripts/create_users.sh

# Создание кастомного пользователя
./scripts/create_users.sh custom john john@company.com password123 relay-user 20 200 192.168.1.0/24
```

**Создаваемые пользователи:**
- `admin/SecurePassword123!` - администратор (relay-admin)
- `user/UserPassword123!` - обычный пользователь (relay-user)
- `monitor/MonitorPassword123!` - мониторинг (relay-monitor)

### 3. `backup_keycloak.sh` - Резервное копирование
Выполняет полное резервное копирование Keycloak с возможностью сжатия и шифрования.

**Использование:**
```bash
# Базовое резервное копирование
./scripts/backup_keycloak.sh

# Резервное копирование с шифрованием
ENCRYPT_BACKUP=true \
ENCRYPTION_PASSWORD=mysecretpassword \
./scripts/backup_keycloak.sh

# Резервное копирование без сжатия
COMPRESS_BACKUP=false \
./scripts/backup_keycloak.sh
```

**Что копируется:**
- Realm конфигурация
- Пользователи
- Клиенты
- Роли
- Группы

### 4. `test_integration.sh` - Тестирование интеграции
Тестирует интеграцию между Keycloak и CloudBridge Relay.

**Использование:**
```bash
# Тестирование с локальными сервисами
./scripts/test_integration.sh

# Тестирование с кастомными URL
KEYCLOAK_URL=https://keycloak.company.com \
RELAY_URL=https://relay.company.com \
./scripts/test_integration.sh
```

**Что тестируется:**
- Создание realm и клиентов
- JWKS endpoint
- OpenID Connect конфигурация
- Аутентификация пользователей
- Интеграция с Relay
- Go интеграционные тесты

## Переменные окружения

Все скрипты поддерживают следующие переменные окружения:

| Переменная | Описание | По умолчанию |
|------------|----------|--------------|
| `KEYCLOAK_URL` | URL Keycloak сервера | `http://localhost:8080` |
| `KEYCLOAK_ADMIN` | Имя администратора | `admin` |
| `KEYCLOAK_ADMIN_PASSWORD` | Пароль администратора | `admin123` |
| `REALM_NAME` | Имя realm | `cloudbridge` |
| `CLIENT_ID` | ID клиента | `relay-client` |
| `CLIENT_SECRET` | Секрет клиента | `relay-client-secret` |
| `RELAY_URL` | URL Relay сервера | `http://localhost:8080` |

### Дополнительные переменные для backup_keycloak.sh

| Переменная | Описание | По умолчанию |
|------------|----------|--------------|
| `BACKUP_DIR` | Директория для резервных копий | `./backups` |
| `BACKUP_RETENTION_DAYS` | Дни хранения резервных копий | `30` |
| `COMPRESS_BACKUP` | Сжатие резервных копий | `true` |
| `ENCRYPT_BACKUP` | Шифрование резервных копий | `false` |
| `ENCRYPTION_PASSWORD` | Пароль для шифрования | - |

## Зависимости

Все скрипты требуют следующие инструменты:

- `curl` - для HTTP запросов
- `jq` - для работы с JSON
- `gzip` - для сжатия (если включено)
- `openssl` - для шифрования (если включено)
- `go` - для запуска тестов

**Установка зависимостей (Ubuntu/Debian):**
```bash
sudo apt update
sudo apt install curl jq gzip openssl golang-go
```

**Установка зависимостей (CentOS/RHEL):**
```bash
sudo yum install curl jq gzip openssl golang
```

## Последовательность использования

### 1. Первоначальная настройка
```bash
# 1. Запуск Keycloak
docker-compose -f docker-compose.keycloak.yml up -d

# 2. Ожидание готовности Keycloak
sleep 30

# 3. Настройка Keycloak
./scripts/setup_keycloak.sh

# 4. Создание пользователей
./scripts/create_users.sh

# 5. Тестирование интеграции
./scripts/test_integration.sh
```

### 2. Ежедневное резервное копирование
```bash
# Добавить в crontab
0 2 * * * /path/to/cloudbridge-relay-installer/scripts/backup_keycloak.sh
```

### 3. Создание новых пользователей
```bash
# Создание пользователя с ролью relay-user
./scripts/create_users.sh custom alice alice@company.com password123 relay-user

# Создание администратора
./scripts/create_users.sh custom bob bob@company.com password123 relay-admin
```

## Troubleshooting

### Проблемы с подключением к Keycloak
```bash
# Проверка доступности
curl -I http://localhost:8080/health

# Проверка логинов
docker-compose -f docker-compose.keycloak.yml logs keycloak
```

### Проблемы с аутентификацией
```bash
# Проверка realm
curl -s http://localhost:8080/realms/cloudbridge/.well-known/openid_configuration

# Проверка клиента
curl -s http://localhost:8080/realms/cloudbridge/protocol/openid-connect/certs
```

### Проблемы с резервным копированием
```bash
# Проверка прав доступа
ls -la ./backups/

# Проверка места на диске
df -h
```

## Безопасность

### Рекомендации по безопасности
1. **Измените пароли по умолчанию** после первоначальной настройки
2. **Используйте HTTPS** в продакшене
3. **Включите шифрование** для резервных копий
4. **Ограничьте доступ** к скриптам только администраторам
5. **Регулярно обновляйте** пароли пользователей

### Переменные для продакшена
```bash
# Продакшен конфигурация
export KEYCLOAK_URL="https://keycloak.company.com"
export KEYCLOAK_ADMIN_PASSWORD="REPLACE_WITH_SECURE_PASSWORD"
export CLIENT_SECRET="REPLACE_WITH_SECURE_SECRET"
export ENCRYPT_BACKUP="true"
export ENCRYPTION_PASSWORD="REPLACE_WITH_SECURE_ENCRYPTION_PASSWORD"
```

## Логирование

Все скрипты выводят подробные логи с цветовой кодировкой:
- 🔵 **INFO** - информационные сообщения
- 🟢 **SUCCESS** - успешные операции
- 🟡 **WARNING** - предупреждения
- 🔴 **ERROR** - ошибки

Для отключения цветного вывода используйте переменную:
```bash
export NO_COLOR=1
```

## Поддержка

При возникновении проблем:
1. Проверьте логи скриптов
2. Убедитесь в доступности Keycloak
3. Проверьте переменные окружения
4. Обратитесь к документации проекта
5. Создайте issue в репозитории проекта 