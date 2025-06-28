# Lynis Security Audit - Quick Reference

## 🚀 Быстрый старт

### Установка
```bash
sudo bash scripts/setup_lynis.sh
```

### Первый запуск
```bash
sudo relay-security-audit full
```

## 📋 Основные команды

### Аудит безопасности
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

### Управление
```bash
# Статус сервисов
sudo relay-security-manager status

# Последний отчёт
sudo relay-security-manager report

# Автоисправление проблем
sudo relay-security-manager fix all

# Управление расписанием
sudo relay-security-manager schedule status
sudo relay-security-manager schedule enable
sudo relay-security-manager schedule disable
```

### Прямой вызов Lynis
```bash
# Базовый аудит
sudo lynis audit system

# Аудит с профилем
sudo lynis audit system --profile /etc/lynis/custom.prf

# Быстрый аудит
sudo lynis audit system --quick

# Подробный вывод
sudo lynis audit system --verbose
```

## 📊 Мониторинг

### Проверка метрик
```bash
# Просмотр метрик Prometheus
sudo cat /var/log/lynis/lynis-metrics.prom

# Проверка статуса экспорта
sudo systemctl status lynis-metrics-exporter
```

### Логи и отчёты
```bash
# Основной отчёт
sudo cat /var/log/lynis/lynis-report.dat

# Лог выполнения
sudo tail -f /var/log/lynis/lynis.log

# Критические проблемы
sudo grep "Suggestion.*CRITICAL" /var/log/lynis/lynis-report.dat

# Основные метрики
sudo grep -E "(Hardening index|warnings|suggestions)" /var/log/lynis/lynis-report.dat
```

## 🔧 Устранение неполадок

### Частые проблемы

#### Ошибка "Invalid option '/etc/lynis/lynis.conf'"
```bash
# Обновить скрипт
sudo cp scripts/relay-security-audit /usr/local/bin/relay-security-audit
sudo chmod +x /usr/local/bin/relay-security-audit
```

#### "PID file exists"
```bash
# Остановить процессы
sudo pkill -f lynis

# Удалить PID файл
sudo rm -f /var/log/lynis/lynis.pid
```

#### "Permission denied"
```bash
# Проверить права
ls -la /usr/local/bin/relay-security-audit

# Установить права
sudo chmod +x /usr/local/bin/relay-security-audit
```

### Диагностика
```bash
# Логи systemd
sudo journalctl -u relay-security-audit.service
sudo journalctl -u lynis-metrics-exporter.service

# Статус сервисов
sudo systemctl status relay-security-audit.timer
sudo systemctl status lynis-metrics-exporter

# Проверка конфигурации
sudo cat /etc/lynis/custom.prf
sudo cat /etc/lynis/lynis.conf
```

## 📈 Метрики Prometheus

### Доступные метрики
```prometheus
# Оценка безопасности (0-100)
lynis_security_score

# Количество предупреждений
lynis_warnings_total

# Количество рекомендаций
lynis_suggestions_total

# Критические проблемы
lynis_critical_issues_total

# Время последнего аудита
lynis_audit_timestamp
```

### Алерты
```yaml
# Низкая оценка безопасности
lynis_security_score < 70

# Критические проблемы
lynis_critical_issues_total > 0

# Устаревший аудит
time() - lynis_audit_timestamp > 86400
```

## 🗂️ Файловая структура

```
/etc/lynis/
├── lynis.conf          # Основная конфигурация
└── custom.prf          # Профиль безопасности

/var/log/lynis/
├── lynis-report.dat    # Основной отчёт
├── lynis.log           # Лог выполнения
├── lynis-metrics.prom  # Метрики Prometheus
├── relay-security-*    # Архивные отчёты
└── lynis.pid           # PID файл процесса

/usr/local/bin/
├── relay-security-audit    # Скрипт аудита
├── relay-security-manager  # Управление
└── lynis-metrics-exporter  # Экспорт метрик
```

## 🔄 Обновление

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

## 🗑️ Удаление

### Полное удаление
```bash
# Остановить сервисы
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

## 📚 Дополнительные ресурсы

- [📖 Полная документация](lynis_integration.md)
- [🔗 Lynis GitHub](https://github.com/CISOfy/lynis)
- [🌐 CISOfy Documentation](https://cisofy.com/lynis/)
- [📊 Grafana Dashboard](http://localhost:3000)
- [📈 Prometheus Metrics](http://localhost:9090) 