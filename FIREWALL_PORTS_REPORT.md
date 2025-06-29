# Отчет о состоянии портов и файрвола CloudBridge Relay

## 📊 Текущий статус

**Дата проверки**: 2025-06-30  
**Файрвол**: `inactive` (отключен)  
**Все порты**: открыты для всех подключений

## 🔍 Анализ портов

### ✅ Внешние сервисы (должны быть доступны извне)

| Сервис | Порт | Статус | HTTP Response | Описание |
|--------|------|--------|---------------|----------|
| **Keycloak** | 8080 | ✅ Слушает | 302 (Redirect) | Административная консоль |
| **Relay API** | 8082 | ✅ Слушает | 000 (TLS Required) | Основной API Relay |
| **SSH** | 22 | ✅ Слушает | - | Удаленный доступ |

### 🔒 Внутренние сервисы (только localhost)

| Сервис | Порт | Статус | HTTP Response | Описание |
|--------|------|--------|---------------|----------|
| **Relay Metrics** | 8081 | ✅ Слушает | 404 (Metrics) | Метрики Relay |
| **Prometheus** | 9090 | ✅ Слушает | 302 (Redirect) | Система мониторинга |
| **AlertManager** | 9093 | ✅ Слушает | 200 (OK) | Система алертов |
| **Blackbox** | 9115 | ✅ Слушает | 200 (OK) | Мониторинг доступности |
| **Grafana** | 3000 | ✅ Слушает | 302 (Redirect) | Дашборды |

## ⚠️ Проблемы безопасности

### 1. Файрвол отключен
- **Проблема**: UFW неактивен
- **Риск**: Все порты открыты для внешнего доступа
- **Решение**: Активировать файрвол с правильными правилами

### 2. Relay API требует TLS
- **Проблема**: Relay API настроен только для TLS соединений
- **Статус**: Нормально (сервис работает корректно)
- **Примечание**: Для HTTP доступа нужно использовать порт 8081 (metrics)

### 3. Внутренние сервисы доступны извне
- **Проблема**: Прометеус, Графана и другие внутренние сервисы слушают на 0.0.0.0
- **Риск**: Неавторизованный доступ к метрикам и дашбордам
- **Решение**: Настроить файрвол для ограничения доступа

## 🛡️ Рекомендуемая конфигурация файрвола

### Внешние порты (открыты для всех)
```bash
# SSH с rate limiting
ufw allow 22/tcp comment "SSH"
ufw limit 22/tcp comment "SSH rate limiting"

# Keycloak Admin Console
ufw allow 8080/tcp comment "Keycloak Admin Console"

# Relay API
ufw allow 8082/tcp comment "Relay API"
```

### Внутренние порты (только localhost)
```bash
# Relay Metrics
ufw allow from 127.0.0.1 to any port 8081 comment "Relay Metrics (localhost only)"

# Prometheus
ufw allow from 127.0.0.1 to any port 9090 comment "Prometheus (localhost only)"

# AlertManager
ufw allow from 127.0.0.1 to any port 9093 comment "AlertManager (localhost only)"

# Blackbox Exporter
ufw allow from 127.0.0.1 to any port 9115 comment "Blackbox Exporter (localhost only)"

# Grafana
ufw allow from 127.0.0.1 to any port 3000 comment "Grafana (localhost only)"
```

### Docker сеть
```bash
# Разрешить трафик внутри Docker
ufw allow from 172.16.0.0/12 comment "Docker internal network"
ufw allow from 192.168.0.0/16 comment "Docker internal network"
ufw allow from 10.0.0.0/8 comment "Docker internal network"
```

## 🚀 Автоматическая настройка

### Использование скрипта
```bash
# Настройка файрвола
sudo ./scripts/setup_firewall.sh

# Проверка статуса
sudo ./scripts/setup_firewall.sh --status

# Отключение файрвола (если нужно)
sudo ./scripts/setup_firewall.sh --disable
```

### Ручная настройка
```bash
# Базовые правила
sudo ufw --force reset
sudo ufw default deny incoming
sudo ufw default allow outgoing

# Внешние сервисы
sudo ufw allow 22/tcp comment "SSH"
sudo ufw limit 22/tcp comment "SSH rate limiting"
sudo ufw allow 8080/tcp comment "Keycloak Admin Console"
sudo ufw allow 8082/tcp comment "Relay API"

# Внутренние сервисы
sudo ufw allow from 127.0.0.1 to any port 8081 comment "Relay Metrics"
sudo ufw allow from 127.0.0.1 to any port 9090 comment "Prometheus"
sudo ufw allow from 127.0.0.1 to any port 9093 comment "AlertManager"
sudo ufw allow from 127.0.0.1 to any port 9115 comment "Blackbox"
sudo ufw allow from 127.0.0.1 to any port 3000 comment "Grafana"

# Docker сеть
sudo ufw allow from 172.16.0.0/12 comment "Docker internal network"
sudo ufw allow from 192.168.0.0/16 comment "Docker internal network"
sudo ufw allow from 10.0.0.0/8 comment "Docker internal network"

# Активация
sudo ufw --force enable
```

## 📋 Проверка после настройки

### 1. Статус файрвола
```bash
sudo ufw status verbose
```

### 2. Проверка доступности
```bash
# Внешние сервисы (должны быть доступны)
curl -I http://localhost:8080  # Keycloak
curl -I http://localhost:8082  # Relay API

# Внутренние сервисы (только localhost)
curl -I http://localhost:8081  # Relay Metrics
curl -I http://localhost:9090  # Prometheus
curl -I http://localhost:3000  # Grafana
```

### 3. Проверка блокировки
```bash
# Попытка доступа с внешнего IP (должна быть заблокирована)
curl -I http://EXTERNAL_IP:9090  # Prometheus
curl -I http://EXTERNAL_IP:3000  # Grafana
```

## 🔧 Дополнительные рекомендации

### 1. Мониторинг безопасности
- Настройте fail2ban для защиты SSH
- Включите логирование UFW
- Регулярно проверяйте логи доступа

### 2. Обновления
- Регулярно обновляйте систему
- Следите за уязвимостями в Docker образах
- Обновляйте сертификаты

### 3. Резервное копирование
- Сохраните конфигурацию файрвола
- Регулярно делайте бэкапы конфигурации
- Документируйте изменения

## 📈 Статистика

### Текущее состояние
- **Всего портов**: 8 активных
- **Внешних портов**: 3 (22, 8080, 8082)
- **Внутренних портов**: 5 (8081, 9090, 9093, 9115, 3000)
- **Файрвол**: неактивен
- **Безопасность**: низкая

### После настройки файрвола
- **Внешний доступ**: только SSH, Keycloak, Relay API
- **Внутренний доступ**: все сервисы мониторинга
- **Безопасность**: высокая
- **Функциональность**: полная

---

**Рекомендация**: Немедленно активировать файрвол для защиты сервисов! 