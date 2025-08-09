# 🧪 Тестирование CloudBridge Relay

Эта директория содержит все тестовые скрипты для CloudBridge Relay.

## 📁 Структура

### `/p2p/` - P2P Mesh Network тесты
- `test_p2p_mesh.sh` - Базовое тестирование P2P соединений
- `send_udp_data.py` - Отправка структурированных UDP данных
- `quick_udp_send.sh` - Быстрая отправка UDP пакетов
- `interactive_udp_tester.sh` - Интерактивное тестирование

### `/quic/` - QUIC протокол тесты
- `test_quic_to_p2p_client.sh` - QUIC клиент симулятор
- `test_quic_to_p2p_server_a.sh` - Сервер A с P2P маршрутизацией
- `test_quic_to_p2p_monitor.sh` - Мониторинг QUIC→P2P потока
- `run_quic_p2p_test.sh` - Полная интеграционная среда

### `/demo/` - Демонстрационные тесты
- `demo_monitor.sh` - Расширенный мониторинг с Prometheus API
- `demo_server_a.sh` - Улучшенный отправитель с автотестами
- `demo_server_b.sh` - Умный получатель с обработкой JSON
- `run_demo.sh` - Основной скрипт запуска демо

### `/integration/` - Интеграционные тесты
- `terminal1_relay_monitor.sh` - Мониторинг relay сервера
- `terminal2_server_a.sh` - Симуляция сервера A
- `terminal3_server_b.sh` - Симуляция сервера B
- `start_three_terminals.sh` - Запуск tmux среды

## 🚀 Быстрый старт

### Базовое P2P тестирование:
```bash
cd p2p/
./test_p2p_mesh.sh
```

### QUIC-to-P2P интеграция:
```bash
cd quic/
./run_quic_p2p_test.sh
```

### Полная демонстрация:
```bash
cd demo/
./run_demo.sh
```

## 📊 Что тестируется

- **P2P Mesh Network** - WireGuard соединения между серверами
- **Enhanced QUIC** - Производительность и совместимость протокола
- **Мультиарендность** - Изоляция между тенантами
- **Мониторинг** - Prometheus метрики и Grafana дашборды
- **Безопасность** - DDoS защита и аномалии
- **Производительность** - Пропускная способность и задержки
