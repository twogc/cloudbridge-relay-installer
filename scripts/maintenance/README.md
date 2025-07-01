# Maintenance Scripts

Скрипты для обслуживания и мониторинга CloudBridge Relay.

## Резервное копирование

### `backup_keycloak.sh`
Создание резервных копий данных Keycloak.

## Управление сертификатами

### `renew_certs.sh`
Обновление SSL сертификатов.

### `restore_certs.sh`
Восстановление сертификатов из резервной копии.

### `update_certs.sh`
Обновление и настройка сертификатов.

### `fix_blackbox_tls.sh`
Исправление проблем с TLS в Blackbox Exporter.

## Управление логированием

### `manage_logs.sh`
Управление логами системы (ротация, очистка, анализ).

## Обновление конфигураций

### `update_alertmanager_email.sh`
Обновление email настроек Alertmanager.

### `update_keycloak_public_key.sh`
Обновление публичного ключа Keycloak.

### `update_keycloak_public_key_from_file.sh`
Обновление публичного ключа Keycloak из файла.

## Использование

### Создание резервной копии:
```bash
./maintenance/backup_keycloak.sh
```

### Обновление сертификатов:
```bash
./maintenance/renew_certs.sh
```

### Управление логами:
```bash
./maintenance/manage_logs.sh
```

### Обновление ключа Keycloak:
```bash
./maintenance/update_keycloak_public_key.sh
```

## Автоматизация

Рекомендуемые cron задачи:
```bash
# Ежедневное резервное копирование
0 1 * * * /path/to/scripts/maintenance/backup_keycloak.sh

# Еженедельная ротация логов
0 3 * * 0 /path/to/scripts/maintenance/manage_logs.sh

# Ежемесячное обновление сертификатов
0 2 1 * * /path/to/scripts/maintenance/renew_certs.sh
``` 