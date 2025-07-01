# Конфигурация портов CloudBridge Relay

Этот документ описывает конфигурацию портов для всех сервисов CloudBridge Relay.

## 📋 Схема портов

### Основные сервисы

| Сервис | Внешний порт | Внутренний порт | Описание |
|--------|-------------|----------------|----------|
| **Keycloak** | 8080 | 8080 | Административная консоль Keycloak |
| **Relay API** | 8082 | 8080 | Основной API Relay сервера |
| **Relay Metrics** | 8081 | 8081 | Метрики и мониторинг Relay |
| **Prometheus** | 9090 | 9090 | Система мониторинга |
| **AlertManager** | 9093 | 9093 | Система алертов |
| **Grafana** | 3000 | 3000 | Дашборды и визуализация |
| **Blackbox** | 9115 | 9115 | Мониторинг доступности |

### Дополнительные сервисы

| Сервис | Порт | Описание |
|--------|------|----------|
| **PostgreSQL** | 5432 | База данных Keycloak (внутренний) |
| **Node Exporter** | 9100 | Системные метрики (если используется) |

## 🔧 Настройка портов

### Переменные окружения

```bash
# Основные порты Relay
RELAY_PORT=8082              # Порт API Relay
RELAY_METRICS_PORT=8081      # Порт метрик Relay

# Порт Keycloak (фиксированный)
KEYCLOAK_PORT=8080           # Административная консоль

# Порт Grafana (фиксированный)
GRAFANA_PORT=3000            # Веб-интерфейс Grafana
```

### Docker Compose конфигурация

```yaml
services:
  keycloak:
    ports:
      - "8080:8080"          # Keycloak Admin Console
  
  relay:
    ports:
      - "8082:8080"          # Relay API (внешний:внутренний)
      - "8081:8081"          # Relay Metrics
  
  prometheus:
    ports:
      - "9090:9090"          # Prometheus
  
  alertmanager:
    ports:
      - "9093:9093"          # AlertManager
  
  grafana:
    ports:
      - "3000:3000"          # Grafana
  
  blackbox:
    ports:
      - "9115:9115"          # Blackbox Exporter
```

## 🌐 Доступ к сервисам

### Локальный доступ

```bash
# Keycloak Admin Console
http://localhost:8080

# Relay API
http://localhost:8082

# Relay Metrics
http://localhost:8081/metrics

# Prometheus
http://localhost:9090

# AlertManager
http://localhost:9093

# Grafana
http://localhost:3000
```

### Сетевой доступ

```bash
# Для внешнего доступа замените localhost на IP сервера
http://your-server-ip:8080   # Keycloak
http://your-server-ip:8082   # Relay API
http://your-server-ip:3000   # Grafana
```

## 🔒 Безопасность портов

### Firewall конфигурация

```bash
# Базовые правила безопасности
sudo ufw default deny incoming
sudo ufw default allow outgoing

# Разрешенные порты
sudo ufw allow 22/tcp        # SSH
sudo ufw allow 8080/tcp      # Keycloak Admin Console
sudo ufw allow 8082/tcp      # Relay API
sudo ufw allow 8081/tcp      # Relay Metrics (только для внутренней сети)
sudo ufw allow 9090/tcp      # Prometheus (только для внутренней сети)
sudo ufw allow 9093/tcp      # AlertManager (только для внутренней сети)
sudo ufw allow 3000/tcp      # Grafana (только для внутренней сети)

# Активация
sudo ufw enable
```

### Рекомендации по безопасности

1. **Внешний доступ**: Только порты 22 (SSH), 8080 (Keycloak), 8082 (Relay API)
2. **Внутренний доступ**: Порт 3000 (Grafana) только для администраторов
3. **Мониторинг**: Порты 8081, 9090, 9093 только для внутренней сети
4. **HTTPS**: Используйте reverse proxy (nginx) для HTTPS

## 🔄 Изменение портов

### Для Relay API

```bash
# В файле .env
RELAY_PORT=8083

# В docker-compose
ports:
  - "8083:8080"
```

### Для Keycloak

```bash
# В docker-compose.keycloak.yml
keycloak:
  ports:
    - "8081:8080"  # Изменить внешний порт
```

### Для Grafana

```bash
# В docker-compose
grafana:
  ports:
    - "3001:3000"  # Изменить внешний порт
```

## 🚨 Устранение конфликтов портов

### Проверка занятых портов

```bash
# Проверка всех занятых портов
sudo netstat -tulpn | grep LISTEN

# Проверка конкретного порта
sudo lsof -i :8080
sudo lsof -i :8082
```

### Решение конфликтов

1. **Остановить конфликтующий сервис**:
   ```bash
   sudo systemctl stop conflicting-service
   ```

2. **Изменить порт в конфигурации**:
   ```bash
   # В .env файле
   RELAY_PORT=8083
   ```

3. **Перезапустить сервисы**:
   ```bash
   docker-compose -f docker-compose.keycloak.yml down
   docker-compose -f docker-compose.keycloak.yml up -d
   ```

## 📊 Мониторинг портов

### Проверка доступности

```bash
# Проверка всех сервисов
curl -f http://localhost:8080/health  # Keycloak
curl -f http://localhost:8082/health  # Relay
curl -f http://localhost:8081/metrics # Metrics
curl -f http://localhost:9090/-/healthy # Prometheus
curl -f http://localhost:3000/api/health # Grafana
```

### Логирование доступа

```bash
# Мониторинг доступа к портам
sudo tcpdump -i any port 8080 or port 8082 or port 3000

# Логи firewall
sudo ufw status verbose
```

## 📝 Примеры конфигурации

### Production окружение

```bash
# .env.production
RELAY_PORT=8082
RELAY_METRICS_PORT=8081
KEYCLOAK_PORT=8080
GRAFANA_PORT=3000

# Firewall (только необходимые порты)
sudo ufw allow 22/tcp    # SSH
sudo ufw allow 8080/tcp  # Keycloak
sudo ufw allow 8082/tcp  # Relay API
```

### Development окружение

```bash
# .env.development
RELAY_PORT=8083
RELAY_METRICS_PORT=8084
KEYCLOAK_PORT=8081
GRAFANA_PORT=3001

# Все порты открыты для разработки
sudo ufw allow 22/tcp
sudo ufw allow 8080:9090/tcp
```

---

**Важно**: Всегда проверяйте доступность портов перед запуском сервисов! 