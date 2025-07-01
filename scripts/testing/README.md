# Testing Scripts

Скрипты для тестирования и проверки работоспособности CloudBridge Relay.

## Основные тесты

### `run_all_tests.sh`
Запуск всех тестов системы.

### `test_integration.sh`
Интеграционные тесты системы.

### `test_tls_connections.sh`
Тестирование TLS соединений.

### `test_vpn_client.sh`
Тестирование VPN клиента.

## Использование

### Запуск всех тестов:
```bash
./testing/run_all_tests.sh
```

### Интеграционные тесты:
```bash
./testing/test_integration.sh
```

### Тест TLS соединений:
```bash
./testing/test_tls_connections.sh
```

### Тест VPN клиента:
```bash
./testing/test_vpn_client.sh
```

## Автоматизация тестирования

Рекомендуется настроить регулярные тесты:
```bash
# Ежедневные тесты
0 4 * * * /path/to/scripts/testing/run_all_tests.sh

# Тесты после обновлений
@reboot /path/to/scripts/testing/test_integration.sh
``` 