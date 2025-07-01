# Lynis Security Audit Integration

## Обзор

**Lynis** — это мощный инструмент для аудита безопасности Linux систем от [CISOfy](https://github.com/CISOfy/lynis). Данная интеграция обеспечивает автоматический аудит безопасности CloudBridge Relay v2.0 с поддержкой:

- 🔍 **Автоматического аудита** системы безопасности
- 📊 **Интеграции с Prometheus** для мониторинга
- 📈 **Grafana дашбордов** для визуализации
- ⚙️ **Systemd сервисов** для регулярного выполнения
- 🛠️ **Управляющих скриптов** для удобства

## Архитектура

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Lynis Audit   │───▶│  Security Data  │───▶│  Prometheus     │
│   Engine        │    │  Collection     │    │  Metrics        │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  Audit Reports  │    │  Systemd Timer  │    │  Grafana        │
│  & Logs         │    │  (Daily)        │    │  Dashboards     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Решенные проблемы

### Проблема с systemd сервисом cloudbridge-relay

**Описание проблемы:** Systemd сервис `cloudbridge-relay.service` не запускался с ошибками:
- `No such file or directory` для скриптов
- `Permission denied` для директорий
- `Changing to the requested working directory failed`

**Причины:**
1. Неправильные права доступа к директории `/home/ubuntu` (750 вместо 755)
2. Проблемы с доступом к файлам скриптов в пользовательской директории
3. Ограничения systemd при работе с пользовательскими директориями

**Решение:**
1. Исправлены права на директорию: `sudo chmod 755 /home/ubuntu`
2. Создан wrapper-скрипт `/usr/local/bin/relay_start_wrapper.sh`
3. Обновлен systemd сервис для использования wrapper-скрипта
4. Добавлена поддержка запуска от root с переключением на ubuntu

**Результат:** 
- ✅ Systemd сервис успешно запускается
- ✅ Все компоненты CloudBridge Relay работают корректно
- ✅ Интеграция с Lynis функционирует
- ✅ Документация обновлена с решением проблемы

### Проблема с таймаутами и блокировками Lynis

**Описание проблемы:** Аудит Lynis не завершался из-за:
- Превышения таймаута (10 минут было недостаточно)
- Блокировок от параллельных запусков
- "Зависших" lock-файлов после прерываний

**Причины:**
1. Таймаут в 600 секунд был слишком коротким для медленных систем
2. Systemd-таймер запускал аудит каждые 12 часов, создавая конфликты
3. Lock-файлы не очищались после аварийных завершений

**Решение:**
1. **Убран таймаут** - Lynis теперь работает столько, сколько нужно
2. **Удален автоматический запуск** через systemd-таймер
3. **Добавлена автоматическая очистка** lock-файлов в скриптах
4. **Созданы новые скрипты** для безопасного запуска

**Результат:**
- ✅ Аудит завершается корректно без прерываний
- ✅ Нет конфликтов между ручным и автоматическим запуском
- ✅ Lock-файлы автоматически очищаются
- ✅ Добавлена ротация логов через logrotate

### Новые скрипты и улучшения

**Созданные скрипты:**
1. `scripts/security_hardening.sh` - автоматическое исправление 47 проблем безопасности
2. `scripts/quick_audit.sh` - быстрый запуск аудита без Telegram
3. Обновлен `scripts/lynis_with_telegram.sh` - убран таймаут, добавлена очистка lock-файлов

**Настройка logrotate:**
- Создан `/etc/logrotate.d/cloudbridge-relay` для автоматической ротации логов
- Логи Lynis и CloudBridge Relay теперь автоматически архивируются

**Улучшения безопасности:**
- SSH hardening (отключение root login, X11 forwarding и т.д.)
- Kernel hardening (sysctl настройки)
- Права доступа к критическим файлам
- Политика паролей
- Баннеры безопасности
- Отключение USB storage
- Настройка auditd и Fail2ban

## Установка

### Автоматическая установка

```bash
# Запуск автоматической установки
sudo bash scripts/setup_lynis.sh
```

### Ручная установка

```bash
# 1. Клонирование Lynis
cd /tmp
git clone https://github.com/CISOfy/lynis.git
cd lynis

# 2. Установка в систему
sudo mkdir -p /usr/local/lynis
sudo cp -r * /usr/local/lynis/
sudo ln -sf /usr/local/lynis/lynis /usr/local/bin/lynis

# 3. Создание конфигурации
sudo mkdir -p /etc/lynis /var/log/lynis
sudo chmod 750 /var/log/lynis
```

## Конфигурация

### Основная конфигурация

Файл: `/etc/lynis/lynis.conf`

```ini
# CloudBridge Relay - Lynis Configuration
PROFILE=relay-server
SKIP_TESTS=test_ssl_certificate_expiry
SKIP_TESTS=test_ssl_certificate_validity

# Настройки отчета
REPORT_FILE=/var/log/lynis/lynis-report.dat
LOG_FILE=/var/log/lynis/lynis.log
REPORT_FORMAT=text

# Настройки безопасности
QUIET=0
VERBOSE=1
WARNINGS=1

# Настройки для российских стандартов
COMPLIANCE_152_FZ=true
COMPLIANCE_187_FZ=true
USE_RUSSIAN_CRYPTO=true
```

### Профиль безопасности

Файл: `/etc/lynis/custom.prf`

```ini
# CloudBridge Relay Security Profile

# Настройки для relay сервера
RELAY_SERVER=true
DOCKER_ENVIRONMENT=true
CONTAINERIZED_APPLICATION=true

# Настройки для multi-tenancy
MULTI_TENANT=true
TENANT_ISOLATION=true

# Настройки для TCP proxy
TCP_PROXY=true
HIGH_PERFORMANCE=true

# Настройки для мониторинга
MONITORING_ENABLED=true
METRICS_COLLECTION=true
PROMETHEUS_INTEGRATION=true

# Настройки для аутентификации
KEYCLOAK_INTEGRATION=true
JWT_AUTHENTICATION=true
MTLS_ENABLED=true

# Настройки для соответствия российским стандартам
RUSSIAN_COMPLIANCE=true
GOST_CRYPTO=true
FZ_152_COMPLIANCE=true
FZ_187_COMPLIANCE=true
```

## Использование

### Основные команды

#### Запуск аудита

```bash
# Полный аудит системы
sudo relay-security-audit full

# Быстрый аудит
sudo relay-security-audit quick

# Аудит Docker окружения
sudo relay-security-audit docker

# Аудит сетевой безопасности
sudo relay-security-audit network

# Аудит файловой системы
sudo relay-security-audit files
```

#### Управление аудитом

```bash
# Показать статус сервисов
sudo relay-security-manager status

# Показать последний отчёт
sudo relay-security-manager report

# Автоматическое исправление проблем
sudo relay-security-manager fix all

# Управление расписанием
sudo relay-security-manager schedule status
sudo relay-security-manager schedule enable
sudo relay-security-manager schedule disable
```

### Запуск аудита в фоне

Для длительных аудитов рекомендуется запускать в фоне:

```bash
# Запуск быстрого аудита в фоне
sudo relay-security-audit quick &

# Запуск полного аудита в фоне
sudo relay-security-audit full &

# Проверка статуса фонового процесса
jobs

# Просмотр вывода фонового процесса
fg

# Отправка процесса в фон (если запущен в интерактивном режиме)
# Ctrl+Z, затем: bg
```

### Прямой вызов Lynis

```bash
# Базовый аудит
sudo lynis audit system

# Аудит с профилем
sudo lynis audit system --profile /etc/lynis/custom.prf

# Быстрый аудит
sudo lynis audit system --quick

# Аудит с подробным выводом
sudo lynis audit system --verbose
```

## Systemd Сервисы

### Автоматический аудит

**Сервис:** `relay-security-audit.service`
**Таймер:** `relay-security-audit.timer`

```ini
# /etc/systemd/system/relay-security-audit.service
[Unit]
Description=CloudBridge Relay Security Audit Service
Documentation=https://github.com/CISOfy/lynis
After=network.target

[Service]
Type=oneshot
ExecStart=/usr/local/bin/relay-security-audit full
User=root
Group=root
StandardOutput=journal
StandardError=journal

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/log/lynis

[Install]
WantedBy=multi-user.target
```

```ini
# /etc/systemd/system/relay-security-audit.timer
[Unit]
Description=Run CloudBridge Relay Security Audit daily
Documentation=https://github.com/CISOfy/lynis

[Timer]
OnCalendar=daily
Persistent=true
RandomizedDelaySec=3600

[Install]
WantedBy=timers.target
```

### Экспорт метрик

**Сервис:** `lynis-metrics-exporter.service`

```ini
# /etc/systemd/system/lynis-metrics-exporter.service
[Unit]
Description=Lynis Metrics Exporter
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/lynis-metrics-exporter
User=root
Group=root
Restart=always
RestartSec=10

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/log/lynis

[Install]
WantedBy=multi-user.target
```

### Управление сервисами

```bash
# Включение автоматического аудита
sudo systemctl enable relay-security-audit.timer
sudo systemctl start relay-security-audit.timer

# Включение экспорта метрик
sudo systemctl enable lynis-metrics-exporter
sudo systemctl start lynis-metrics-exporter

# Проверка статуса
sudo systemctl status relay-security-audit.timer
sudo systemctl status lynis-metrics-exporter
```

## Интеграция с Prometheus

### Метрики

Lynis экспортирует следующие метрики в формате Prometheus:

```prometheus
# HELP lynis_security_score Security hardening score (0-100)
# TYPE lynis_security_score gauge
lynis_security_score 85

# HELP lynis_warnings_total Total number of warnings
# TYPE lynis_warnings_total gauge
lynis_warnings_total 3

# HELP lynis_suggestions_total Total number of suggestions
# TYPE lynis_suggestions_total gauge
lynis_suggestions_total 12

# HELP lynis_critical_issues_total Total number of critical issues
# TYPE lynis_critical_issues_total gauge
lynis_critical_issues_total 0

# HELP lynis_audit_timestamp Last audit timestamp
# TYPE lynis_audit_timestamp gauge
lynis_audit_timestamp 1640995200
```

### Конфигурация Prometheus

Добавьте в `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'lynis-security-audit'
    static_configs:
      - targets: ['localhost:9090']
    file_sd_configs:
      - files:
        - '/var/log/lynis/lynis-metrics.prom'
    metrics_path: /metrics
    scrape_interval: 5m
```

## Интеграция с Grafana

### Дашборд безопасности

Автоматически создаётся дашборд `/var/lib/grafana/dashboards/lynis/lynis-security-audit.json` с панелями:

- **Security Score** - Оценка безопасности (0-100)
- **Security Issues** - Количество предупреждений и рекомендаций
- **Critical Issues** - Критические проблемы
- **Audit Timeline** - Временная шкала аудитов

### Конфигурация Grafana

Файл: `/etc/grafana/provisioning/dashboards/lynis-security-audit.yaml`

```yaml
apiVersion: 1

providers:
  - name: 'Lynis Security Audit'
    orgId: 1
    folder: 'Security'
    type: file
    disableDeletion: false
    updateIntervalSeconds: 10
    allowUiUpdates: true
    options:
      path: /var/lib/grafana/dashboards/lynis
```

## Отчёты и логи

### Структура файлов

```
/var/log/lynis/
├── lynis-report.dat          # Основной отчёт
├── lynis.log                 # Лог выполнения
├── lynis-metrics.prom        # Метрики Prometheus
├── relay-security-audit-*    # Архивные отчёты
└── lynis.pid                 # PID файл процесса
```

### Форматы отчётов

#### Текстовый отчёт

```bash
# Просмотр основного отчёта
sudo cat /var/log/lynis/lynis-report.dat

# Поиск критических проблем
sudo grep "Suggestion.*CRITICAL" /var/log/lynis/lynis-report.dat

# Основные метрики
sudo grep -E "(Hardening index|warnings|suggestions)" /var/log/lynis/lynis-report.dat
```

#### HTML отчёт

```bash
# Генерация HTML отчёта (если доступен lynis-report)
sudo lynis-report --output html --file /var/log/lynis/audit-report.html /var/log/lynis/lynis-report.dat
```

## Мониторинг и алерты

### Prometheus алерты

Добавьте в `prometheus/alerts/relay_alerts.yml`:

```yaml
groups:
  - name: lynis-security-alerts
    rules:
      - alert: LynisSecurityScoreLow
        expr: lynis_security_score < 70
        for: 1h
        labels:
          severity: warning
        annotations:
          summary: "Low security score detected"
          description: "Security score is {{ $value }}/100"

      - alert: LynisCriticalIssues
        expr: lynis_critical_issues_total > 0
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Critical security issues detected"
          description: "{{ $value }} critical security issues found"

      - alert: LynisAuditStale
        expr: time() - lynis_audit_timestamp > 86400
        for: 1h
        labels:
          severity: warning
        annotations:
          summary: "Lynis audit is stale"
          description: "Last audit was {{ $value }} seconds ago"
```

### Email уведомления

Настройте в Alertmanager:

```yaml
route:
  group_by: ['alertname']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'security-team'

receivers:
  - name: 'security-team'
    email_configs:
      - to: 'security@company.com'
        from: 'alertmanager@company.com'
        smarthost: 'smtp.company.com:587'
        auth_username: 'alertmanager@company.com'
        auth_password: 'password'
```

## Автоматическое исправление

### Доступные исправления

```bash
# Исправление всех проблем
sudo relay-security-manager fix all

# Исправление только firewall
sudo relay-security-manager fix firewall

# Обновление системы
sudo relay-security-manager fix updates
```

### Скрипты исправления

Система автоматически выполняет:

1. **Обновление системы**
   ```bash
   apt-get update && apt-get upgrade -y
   ```

2. **Настройка firewall**
   ```bash
   /home/ubuntu/cloudbridge-relay-installer/scripts/setup_firewall.sh
   ```

3. **Настройка fail2ban**
   ```bash
   systemctl enable fail2ban
   systemctl start fail2ban
   ```

## Соответствие стандартам

### Российские стандарты

Интеграция поддерживает:

- **ФЗ-152** "О персональных данных"
- **ФЗ-187** "О безопасности критической информационной инфраструктуры"
- **ГОСТ Р 57580.1-2017** "Безопасность финансовых организаций"
- **ГОСТ Р 57580.2-2018** "Безопасность финансовых организаций"

### Международные стандарты

- **ISO 27001** - Информационная безопасность
- **NIST Cybersecurity Framework**
- **CIS Controls**
- **OWASP Top 10**

## Устранение неполадок

### Частые проблемы

#### 1. Ошибка "Invalid option '/etc/lynis/lynis.conf'"

**Причина:** Использование неподдерживаемой опции `--config-file`

**Решение:**
```bash
# Обновить скрипт
sudo cp scripts/relay-security-audit /usr/local/bin/relay-security-audit
sudo chmod +x /usr/local/bin/relay-security-audit
```

#### 2. "PID file exists, probably another Lynis process is running"

**Причина:** Предыдущий процесс не завершился корректно

**Решение:**
```bash
# Остановить процессы
sudo pkill -f lynis

# Удалить PID файл
sudo rm -f /var/log/lynis/lynis.pid
```

#### 3. "Permission denied" при запуске

**Причина:** Недостаточные права доступа

**Решение:**
```bash
# Проверить права
ls -la /usr/local/bin/relay-security-audit

# Установить права
sudo chmod +x /usr/local/bin/relay-security-audit
```

#### 4. Systemd сервис cloudbridge-relay не запускается

**Причина:** Проблемы с правами доступа к директориям или файлам скриптов

**Симптомы:**
```bash
# Ошибка в логах systemd
sudo journalctl -u cloudbridge-relay.service
# "No such file or directory" или "Permission denied"
```

**Решение:**

1. **Проверить права на директории:**
   ```bash
   # Проверить права на /home/ubuntu
   ls -ld /home/ubuntu
   # Должно быть: drwxr-xr-x
   
   # Если права неправильные, исправить:
   sudo chmod 755 /home/ubuntu
   ```

2. **Создать wrapper-скрипт для systemd:**
   ```bash
   # Создать wrapper-скрипт
   cat > scripts/relay_start_wrapper.sh << 'EOF'
   #!/bin/bash
   
   # CloudBridge Relay - Production Startup Wrapper
   # Запускается от root и переключается на ubuntu
   
   set -e
   
   exec su - ubuntu -c "cd /home/ubuntu/cloudbridge-relay-installer && ./scripts/start_production.sh --env .env.production"
   EOF
   
   # Установить права
   chmod +x scripts/relay_start_wrapper.sh
   
   # Скопировать в системную директорию
   sudo cp scripts/relay_start_wrapper.sh /usr/local/bin/
   sudo chmod +x /usr/local/bin/relay_start_wrapper.sh
   ```

3. **Обновить systemd сервис:**
   ```bash
   # Обновить systemd файл
   sudo cp systemd/cloudbridge-relay.service /etc/systemd/system/
   
   # Перезагрузить systemd
   sudo systemctl daemon-reload
   
   # Перезапустить сервис
   sudo systemctl restart cloudbridge-relay.service
   ```

4. **Альтернативное решение - запуск от root:**
   ```bash
   # Если su не работает в systemd, можно запускать от root
   # Изменить в systemd/cloudbridge-relay.service:
   # User=root
   # Group=root
   # ExecStart=/bin/bash /home/ubuntu/cloudbridge-relay-installer/scripts/start_production.sh --env .env.production
   ```

**Проверка работоспособности:**
```bash
# Проверить статус сервиса
sudo systemctl status cloudbridge-relay.service

# Проверить логи
sudo journalctl -u cloudbridge-relay.service -f

# Ручной запуск для тестирования
sudo /usr/local/bin/relay_start_wrapper.sh
```

#### 5. "PID file exists, probably another Lynis process is running"

**Причина:** Предыдущий процесс не завершился корректно

**Решение:**
```bash
# Остановить процессы
sudo pkill -f lynis

# Удалить PID файл
sudo rm -f /var/log/lynis/lynis.pid
```

### Логи и диагностика

```bash
# Просмотр логов systemd
sudo journalctl -u relay-security-audit.service
sudo journalctl -u lynis-metrics-exporter.service

# Просмотр логов Lynis
sudo tail -f /var/log/lynis/lynis.log

# Проверка статуса сервисов
sudo systemctl status relay-security-audit.timer
sudo systemctl status lynis-metrics-exporter
```

## Обновление

### Обновление Lynis

```bash
# Остановить сервисы
sudo systemctl stop relay-security-audit.timer
sudo systemctl stop lynis-metrics-exporter

# Обновить Lynis
cd /tmp
rm -rf lynis
git clone https://github.com/CISOfy/lynis.git
cd lynis
sudo cp -r * /usr/local/lynis/

# Перезапустить сервисы
sudo systemctl start relay-security-audit.timer
sudo systemctl start lynis-metrics-exporter
```

### Обновление скриптов

```bash
# Обновить управляющие скрипты
sudo cp scripts/relay-security-audit /usr/local/bin/
sudo cp scripts/relay-security-manager /usr/local/bin/
sudo chmod +x /usr/local/bin/relay-security-*

# Перезагрузить systemd
sudo systemctl daemon-reload
```

## Удаление

### Полное удаление

```bash
# Остановить и отключить сервисы
sudo systemctl stop relay-security-audit.timer
sudo systemctl disable relay-security-audit.timer
sudo systemctl stop lynis-metrics-exporter
sudo systemctl disable lynis-metrics-exporter

# Удалить файлы
sudo rm -rf /etc/lynis /var/log/lynis
sudo rm -f /usr/local/bin/relay-security-audit
sudo rm -f /usr/local/bin/relay-security-manager
sudo rm -f /usr/local/bin/lynis-metrics-exporter
sudo rm -f /etc/systemd/system/relay-security-audit.*
sudo rm -f /etc/systemd/system/lynis-metrics-exporter.service

# Удалить Lynis
sudo rm -rf /usr/local/lynis
sudo rm -f /usr/local/bin/lynis

# Перезагрузить systemd
sudo systemctl daemon-reload
```

## Заключение

Интеграция Lynis в CloudBridge Relay v2.0 обеспечивает:

- ✅ **Автоматический аудит** безопасности системы
- ✅ **Интеграцию с мониторингом** Prometheus/Grafana
- ✅ **Соответствие российским стандартам** ФЗ-152, ФЗ-187
- ✅ **Удобное управление** через специализированные скрипты
- ✅ **Автоматическое исправление** выявленных проблем
- ✅ **Регулярные отчёты** и уведомления

Эта интеграция значительно повышает уровень безопасности CloudBridge Relay и обеспечивает соответствие корпоративным стандартам безопасности.

## Ссылки

- [Lynis GitHub Repository](https://github.com/CISOfy/lynis)
- [CISOfy Documentation](https://cisofy.com/lynis/)
- [CloudBridge Relay Documentation](./deployment_guide.md)
- [Security Guide](./security.md)