# Setup Scripts

Скрипты для установки и первоначальной настройки CloudBridge Relay.

## Основные скрипты

### `interactive_setup.sh`
Интерактивная установка с пошаговым руководством пользователя через получение публичного ключа Keycloak.

### `check_dependencies.sh`
Проверка наличия необходимых зависимостей (Docker, Docker Compose).

### `generate_secrets.sh`
Генерация секретов и .env файла для приложения.

### `setup_env.sh`
Настройка переменных окружения.

### `setup_keycloak.sh`
Настройка и конфигурация Keycloak сервера.

### `setup_certs.sh`
Настройка SSL сертификатов.

### `setup_dashboards.sh`
Установка и настройка Grafana дашбордов.

## Дополнительные скрипты

### `setup_fail2ban_logs.sh`
Настройка логирования fail2ban.

### `setup_firewall.sh`
Настройка файрвола.

### `setup_ldap_integration.sh`
Интеграция с LDAP/Active Directory.

### `setup_lynis.sh`
Установка и настройка Lynis для аудита безопасности.

### `setup_moscow_region.sh`
Специальная настройка для московского региона.

### `setup_security_monitor.sh`
Настройка мониторинга безопасности.

### `setup_security_monitor_cron.sh`
Настройка cron задач для мониторинга безопасности.

### `setup_system.sh`
Базовая настройка системы.

### `show_banner.sh`
Отображение красивого баннера CloudBridge Relay.

## Использование

Для полной установки используйте основной скрипт:
```bash
./setup_cloudbridge.sh
```

Или для интерактивной установки:
```bash
./setup/interactive_setup.sh
``` 