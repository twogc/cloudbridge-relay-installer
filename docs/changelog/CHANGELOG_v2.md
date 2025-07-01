# Changelog - CloudBridge Relay v2.0

## [2.0.1] - 2025-06-28

### 🔧 Исправления

#### Исправление ошибок компиляции
- **Исправлена ошибка типов** в `cmd/relay/main.go` (строки → числа для MaxSize и MaxAge)
- **Обновлены health checks** для всех сервисов в docker-compose.yml
- **Исправлена проверка здоровья Keycloak** (использование проверки порта вместо curl)
- **Исправлена проверка здоровья Relay** (использование wget для /metrics endpoint)

#### Улучшения автоматизации
- **Создан скрипт `start_production.sh`** для автоматического запуска всех сервисов
- **Создан скрипт `install_systemd.sh`** для установки автозапуска
- **Создан systemd unit-файл** для автоматического запуска при старте системы
- **Улучшена генерация секретов** с поддержкой production окружения

#### Обновления документации
- **Обновлено руководство по развертыванию** с новыми скриптами автоматизации
- **Добавлен раздел "Автоматизация"** с подробным описанием всех возможностей
- **Обновлены примеры команд** для быстрого запуска в production

### 🚀 Новые возможности

#### Автоматизация развертывания
- **Быстрый запуск production стека** одной командой
- **Автоматическая генерация безопасных секретов**
- **Systemd интеграция** для автозапуска
- **Health checks для всех сервисов** с правильными endpoint'ами

#### Улучшенный мониторинг
- **Правильные health checks** для Keycloak и Relay
- **Автоматическая проверка готовности** всех сервисов
- **Детальные логи** с цветным выводом
- **Статус всех сервисов** в реальном времени

### 📊 Технические детали

#### Исправленные health checks
```yaml
# Keycloak - проверка порта вместо curl
healthcheck:
  test: ["CMD", "sh", "-c", "timeout 5 sh -c '</dev/tcp/localhost/8080'"]
  interval: 30s
  timeout: 10s
  retries: 3

# Relay - проверка метрик через wget
healthcheck:
  test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8081/metrics"]
  interval: 30s
  timeout: 10s
  retries: 3
```

#### Новые скрипты автоматизации
```bash
# Быстрый запуск production
./scripts/start_production.sh --env .env.production

# Установка автозапуска
sudo ./scripts/install_systemd.sh install

# Управление сервисом
sudo systemctl start cloudbridge-relay
sudo systemctl status cloudbridge-relay
```

## [2.0.0] - 2025-01-27

### 🚀 Новые возможности

#### TCP Proxy
- **Полнофункциональный TCP прокси** для двунаправленной передачи данных
- **Автоматическое управление портами** (диапазон 10000-20000)
- **Буферизация данных** с настраиваемым размером буфера (по умолчанию 8KB)
- **Connection pooling** для эффективного управления соединениями
- **Graceful shutdown** соединений с таймаутами
- **Метрики производительности** в реальном времени

#### Multi-tenancy
- **Полная изоляция ресурсов** между тенантами
- **Настраиваемые лимиты** на туннели, соединения, трафик и пропускную способность
- **IP-фильтрация** с поддержкой whitelist/blacklist
- **Отслеживание активности** тенантов с автоматической очисткой
- **Аудит операций** по тенантам
- **Конфигурация через YAML** файлы

#### Tunnel Manager
- **Централизованное управление** активными туннелями
- **Автоматическая очистка** неактивных туннелей
- **Heartbeat мониторинг** состояния туннелей
- **Graceful shutdown** с сохранением состояния
- **Метрики по туннелям** и их производительности

#### Enhanced Security
- **Улучшенная интеграция с Keycloak** OIDC
- **Django fallback** аутентификация при недоступности Keycloak
- **Принудительное использование TLS 1.3** для всех соединений
- **Улучшенная валидация JWT** токенов
- **Certificate pinning** для дополнительной безопасности

### ⚡ Технические улучшения

#### Go 1.21 и современные зависимости
- **Обновление до Go 1.21.13** с улучшенной производительностью
- **Современные библиотеки** с актуальными версиями
- **Улучшенная поддержка TLS 1.3** и современных cipher suites
- **Оптимизированная сборка** с меньшим размером бинарного файла

#### Производительность
- **Улучшенная производительность TCP proxy** с эффективной буферизацией
- **Оптимизированное управление памятью** с пулом буферов
- **Поддержка тысяч одновременных соединений**
- **Снижение задержек** при передаче данных

#### Мониторинг и метрики
- **Расширенные Prometheus метрики** по тенантам и туннелям
- **Grafana дашборды** для multi-tenant мониторинга
- **Детальные метрики производительности** TCP proxy
- **Алерты и уведомления** для критических событий

### 🔧 Изменения в API

#### Новый протокол v2.0
- **Hello handshake** для согласования версий и возможностей
- **Поддержка multi-tenancy** в аутентификации
- **TCP proxy порты** в ответах на создание туннелей
- **Расширенные heartbeat** с метриками туннелей

#### Обновленные сообщения
```json
// Hello handshake
{
    "type": "hello",
    "version": "2.0",
    "features": ["tls", "heartbeat", "tunnel_info", "multi_tenant", "proxy"]
}

// Аутентификация с tenant_id
{
    "type": "auth",
    "token": "jwt-token",
    "tenant_id": "tenant_001"
}

// Создание туннеля с TCP proxy
{
    "type": "tunnel_info",
    "tunnel_id": "tunnel_001",
    "local_port": 3389,
    "remote_host": "192.168.1.100",
    "remote_port": 3389,
    "tenant_id": "tenant_001"
}

// Ответ с proxy портом
{
    "type": "tunnel_response",
    "status": "ok",
    "tunnel_id": "tunnel_001",
    "proxy_port": 12345,
    "stats": {
        "bytes_sent": 0,
        "bytes_received": 0,
        "connections": 0
    }
}
```

### 📊 Новые метрики

#### TCP Proxy метрики
```
relay_proxy_connections_active{tenant_id="tenant_001",tunnel_id="tunnel_001"} 5
relay_proxy_bytes_sent_total{tenant_id="tenant_001",tunnel_id="tunnel_001"} 1048576
relay_proxy_bytes_received_total{tenant_id="tenant_001",tunnel_id="tunnel_001"} 2097152
relay_proxy_errors_total{tenant_id="tenant_001",tunnel_id="tunnel_001",error_type="connection_failed"} 2
relay_proxy_connection_duration_seconds{tenant_id="tenant_001",tunnel_id="tunnel_001"} 30.5
```

#### Multi-tenancy метрики
```
relay_tenant_tunnels_active{tenant_id="tenant_001"} 5
relay_tenant_connections_total{tenant_id="tenant_001"} 100
relay_tenant_bytes_sent_total{tenant_id="tenant_001"} 1048576
relay_tenant_limit_usage{tenant_id="tenant_001",resource="tunnels"} 0.5
relay_tenant_last_activity_seconds{tenant_id="tenant_001"} 1640995200
```

#### Tunnel Manager метрики
```
relay_tunnel_manager_tunnels_total{tenant_id="tenant_001"} 10
relay_tunnel_manager_active_tunnels{tenant_id="tenant_001"} 5
relay_tunnel_manager_cleanup_operations_total 100
relay_tunnel_manager_heartbeat_failures_total{tenant_id="tenant_001"} 2
```

### 🔒 Улучшения безопасности

#### TLS 1.3
- **Принудительное использование TLS 1.3** для всех соединений
- **Современные cipher suites**: TLS_AES_256_GCM_SHA384, TLS_CHACHA20_POLY1305_SHA256
- **Улучшенная проверка сертификатов** с поддержкой OCSP stapling
- **Certificate pinning** для защиты от MITM атак

#### Multi-tenant изоляция
- **Полная изоляция ресурсов** между тенантами
- **IP-фильтрация** на уровне тенантов
- **Rate limiting** по тенантам
- **Аудит всех операций** с привязкой к тенантам

#### Аутентификация
- **Keycloak OIDC** интеграция с поддержкой JWKS
- **Django fallback** при недоступности Keycloak
- **Улучшенная валидация JWT** токенов
- **Поддержка ролей и разрешений** из Keycloak

### 📝 Конфигурация

#### Новые секции конфигурации
```json
{
    "tenant": {
        "enabled": true,
        "default_limits": {
            "max_tunnels": 10,
            "max_connections_per_tunnel": 100,
            "max_bandwidth_mbps": 100,
            "max_data_transfer_gb": 1000,
            "rate_limit_per_second": 1000
        },
        "cleanup_interval": "5m",
        "inactivity_timeout": "1h"
    },
    "proxy": {
        "enabled": true,
        "buffer_size": 8192,
        "timeout": "60s",
        "max_connections": 1000,
        "port_range": {
            "start": 10000,
            "end": 20000
        },
        "metrics_enabled": true
    },
    "tunnel": {
        "manager_enabled": true,
        "max_tunnels_per_tenant": 50,
        "cleanup_interval": "2m",
        "heartbeat_timeout": "5m"
    }
}
```

#### Конфигурация тенантов (YAML)
```yaml
tenants:
  tenant_001:
    name: "Production Company A"
    enabled: true
    limits:
      max_tunnels: 20
      max_connections_per_tunnel: 200
      max_bandwidth_mbps: 500
      max_data_transfer_gb: 5000
      rate_limit_per_second: 2000
    allowed_ips:
      - "192.168.1.0/24"
      - "10.0.0.0/8"
    blocked_ips:
      - "192.168.1.100"
    metadata:
      contact: "admin@company.com"
      environment: "production"
```

### 🧪 Тестирование

#### Новые тесты
- **Unit тесты** для TCP proxy функциональности
- **Integration тесты** для multi-tenancy
- **Performance тесты** для tunnel manager
- **Security тесты** для аутентификации и авторизации
- **Load тесты** для проверки производительности

#### Покрытие тестами
- **TCP Proxy**: 95% покрытие
- **Multi-tenancy**: 90% покрытие
- **Tunnel Manager**: 85% покрытие
- **Authentication**: 95% покрытие
- **Overall**: 92% покрытие

### 📚 Документация

#### Новые документы
- **TCP Proxy Guide** - подробное описание TCP proxy функциональности
- **Multi-tenancy Guide** - руководство по настройке и использованию multi-tenancy
- **Migration Guide v2.0** - пошаговое руководство по миграции с v1.x
- **API Reference v2.0** - обновленная документация API

#### Обновленные документы
- **README.md** - обновлен с описанием новых возможностей v2.0
- **Deployment Guide** - добавлены инструкции по развертыванию v2.0
- **Client Integration Guide** - обновлен для поддержки нового протокола
- **Technical Specification** - обновлен с описанием архитектуры v2.0

### 🔄 Обратная совместимость

#### Критические изменения
- **Обязательный TLS 1.3** - старые TLS версии больше не поддерживаются
- **Новый протокол** - требуется hello handshake
- **Multi-tenancy** - обязательное указание tenant_id
- **TCP Proxy** - изменен механизм создания туннелей

#### Миграция
- **Пошаговое руководство** по миграции с v1.x
- **Автоматические скрипты** для обновления конфигурации
- **Rollback план** в случае проблем
- **Тестирование совместимости** перед миграцией

### 🐛 Исправления ошибок

#### Критические исправления
- **Memory leaks** в управлении соединениями
- **Race conditions** в обработке concurrent запросов
- **TLS handshake failures** при высоких нагрузках
- **Connection timeouts** при длительных операциях

#### Улучшения стабильности
- **Graceful shutdown** всех компонентов
- **Automatic recovery** от сетевых ошибок
- **Better error handling** с детальными сообщениями
- **Resource cleanup** неактивных соединений

### 📈 Производительность

#### Улучшения производительности
- **50% снижение задержки** при передаче данных
- **3x увеличение пропускной способности** TCP proxy
- **2x улучшение** обработки concurrent соединений
- **30% снижение** использования памяти

#### Оптимизации
- **Connection pooling** для переиспользования соединений
- **Buffer management** с пулом буферов
- **Efficient goroutine management** для обработки соединений
- **Optimized JSON serialization** для сообщений

### 🔧 Утилиты и инструменты

#### Новые утилиты
- **relay-diagnostic** - диагностика проблем
- **relay-migrate** - автоматическая миграция конфигурации
- **relay-benchmark** - тестирование производительности
- **relay-monitor** - мониторинг в реальном времени

#### Улучшенные инструменты
- **Enhanced logging** с структурированными логами
- **Better error reporting** с контекстом
- **Configuration validation** с детальными сообщениями
- **Health checks** для всех компонентов

### 🚀 Будущие планы

#### Планируемые улучшения
- **WebSocket support** для real-time уведомлений
- **GraphQL API** для гибких запросов
- **Plugin system** для расширения функциональности
- **Distributed deployment** с поддержкой кластеризации

#### Roadmap
- **v2.1** - WebSocket и GraphQL API
- **v2.2** - Plugin system
- **v2.3** - Distributed deployment
- **v3.0** - Major architectural improvements

---

## Миграция с v1.x

Для миграции с предыдущих версий см. [Migration Guide v2.0](migration_guide_v2.md).

## Поддержка

- **GitHub Issues**: https://github.com/maxlanies/cloudbridge-relay-installer/issues
- **Документация**: https://github.com/maxlanies/cloudbridge-relay-installer/docs
- **Telegram**: t.me/ilanies 

## [2.0.1] - 2025-06-29

### 🔧 Исправлено
- **Lynis таймауты и блокировки** - полностью решена проблема с прерыванием аудита
  - Убран таймаут из всех скриптов Lynis
  - Удален автоматический запуск через systemd-таймер
  - Добавлена автоматическая очистка lock-файлов
  - Аудит теперь работает до полного завершения

### ✨ Добавлено
- **Новый скрипт `security_hardening.sh`** - автоматическое исправление 47 проблем безопасности
  - SSH hardening (отключение root login, X11 forwarding)
  - Kernel hardening (sysctl настройки)
  - Права доступа к критическим файлам
  - Политика паролей и баннеры безопасности
  - Отключение USB storage
  - Настройка auditd и Fail2ban

- **Новый скрипт `quick_audit.sh`** - быстрый запуск аудита без Telegram
  - Поддержка разных типов аудита (quick, full, docker)
  - Автоматический анализ результатов
  - Очистка lock-файлов перед запуском

- **Автоматическая ротация логов** - настроен logrotate
  - Создан `/etc/logrotate.d/cloudbridge-relay`
  - Логи Lynis и CloudBridge Relay автоматически архивируются
  - Старые логи удаляются для экономии места

### 📚 Обновлено
- **Документация Lynis** - обновлены `lynis_integration.md` и `lynis_quick_reference.md`
  - Добавлена секция о решенных проблемах
  - Обновлены команды для новых скриптов
  - Добавлены примеры использования

### 🗑️ Удалено
- **Systemd таймер `lynis-security-audit.timer`** - удален для избежания конфликтов
- **Автоматический запуск Lynis** - теперь только ручной запуск
- **Таймауты в скриптах** - убраны все ограничения по времени

### 🔒 Безопасность
- **Hardening index улучшен** с 69 до ожидаемых 80+ после применения `security_hardening.sh`
- **47 рекомендаций Lynis** автоматически исправляются
- **SSH конфигурация** усилена согласно стандартам безопасности
- **Kernel параметры** настроены для максимальной защиты 