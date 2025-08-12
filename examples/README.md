# 📚 Примеры использования CloudBridge Relay

Готовые примеры для быстрого старта и изучения возможностей CloudBridge Relay.

## 📁 Структура

### `/basic/` - Базовые примеры
- `quick_start.sh` - Простой запуск CloudBridge Relay

### `/advanced/` - Продвинутые примеры
- `full_demo.sh` - Полная демонстрация всех возможностей

## 🚀 Быстрый старт

### Простой запуск:
```bash
cd basic/
./quick_start.sh
```

### Полная демонстрация:
```bash
cd advanced/
./full_demo.sh
```

## 📖 Что включено

### Базовые примеры:
- Запуск Docker контейнеров
- Проверка доступности сервисов
- Базовое тестирование API

### Продвинутые примеры:
- P2P Mesh Network демонстрация
- QUIC-to-P2P интеграция
- Мониторинг в реальном времени
- Prometheus метрики
- Grafana дашборды

## 🎯 Сценарии использования

1. **Корпоративный доступ** - безопасное подключение к внутренним ресурсам
2. **Site-to-Site VPN** - соединение офисов через P2P mesh
3. **Edge Computing** - глобальная сеть edge-серверов
4. **Zero Trust Network** - проверка каждого соединения

## 📊 Мониторинг

После запуска примеров доступны:
- **Relay API**: http://localhost:8082
- **Prometheus**: http://localhost:9090  
- **Grafana**: http://localhost:3000 (admin/admin)
- **P2P Mesh**: UDP 51820

## 🔗 Дополнительная информация

- [Документация API](../docs/api.md)
- [Руководство по настройке](../docs/configuration.md)
- [Troubleshooting](../docs/troubleshooting/)
