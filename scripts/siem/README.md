# SIEM Integration Scripts

Скрипты для интеграции с SIEM (Security Information and Event Management) системами.

## Основные скрипты

### `deploy_siem_server.sh`
Развертывание SIEM сервера.

### `integrate_siem.sh`
Интеграция CloudBridge Relay с SIEM системой.

## Использование

### Развертывание SIEM сервера:
```bash
./siem/deploy_siem_server.sh
```

### Интеграция с SIEM:
```bash
./siem/integrate_siem.sh
```

## Конфигурация

SIEM интеграция требует настройки:
- Endpoint'ов для отправки событий
- Форматов логов
- Аутентификации
- Шифрования данных

Подробная документация по SIEM интеграции находится в папке `docs/`. 