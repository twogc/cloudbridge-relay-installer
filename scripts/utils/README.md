# Utility Scripts

Утилитарные скрипты для управления CloudBridge Relay.

## Управление сервисами

### `start_production.sh`
Запуск всех сервисов в production режиме.

### `start_services.sh`
Запуск основных сервисов.

### `relay_start_wrapper.sh`
Обертка для запуска relay сервиса.

## Создание данных

### `create_real_connections.sh`
Создание реальных соединений для тестирования.

### `create_users.sh`
Создание пользователей в системе.

### `generate_business_metrics.sh`
Генерация бизнес-метрик.

## Системные утилиты

### `install_systemd.sh`
Установка systemd сервисов.

### `post_install.sh`
Пост-установочные задачи.

## Использование

### Запуск production:
```bash
./utils/start_production.sh
```

### Создание пользователей:
```bash
./utils/create_users.sh
```

### Установка systemd сервисов:
```bash
./utils/install_systemd.sh
```

### Генерация метрик:
```bash
./utils/generate_business_metrics.sh
```

## Автоматизация

Рекомендуемые настройки:
```bash
# Автозапуск при загрузке системы
systemctl enable cloudbridge-relay

# Ежедневная генерация метрик
0 5 * * * /path/to/scripts/utils/generate_business_metrics.sh
``` 