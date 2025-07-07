# Lynis Security Audit - Quick Reference

## 🚀 Быстрый старт

### Установка
```bash
sudo bash scripts/setup_lynis.sh
```

### Первый запуск
```bash
sudo ./scripts/lynis_with_telegram.sh
```

### Исправление проблем безопасности
```bash
sudo ./scripts/security_hardening.sh
```

## 📋 Основные команды

### Аудит безопасности
```bash
# Полный аудит с Telegram уведомлениями
sudo ./scripts/lynis_with_telegram.sh

# Быстрый аудит без Telegram
sudo ./scripts/quick_audit.sh quick

# Полный аудит без Telegram
sudo ./scripts/quick_audit.sh full

# Аудит Docker окружения
sudo ./scripts/quick_audit.sh docker

# Исправление проблем безопасности
sudo ./scripts/security_hardening.sh
```

### Управление
```bash
# Статус сервисов
sudo systemctl status relay-security-audit.service

# Последний отчёт
sudo cat /var/log/lynis/lynis-report.dat

# Очистка lock-файлов
sudo rm -f /var/run/lynis_telegram.lock

# Остановка всех процессов Lynis
sudo pkill -f lynis
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

#### Ошибка "Lynis превысил таймаут"
```bash
# Проблема решена - таймаут убран из всех скриптов
# Lynis теперь работает столько, сколько нужно
```

#### "Lynis уже запущен"
```bash
# Остановить все процессы
sudo pkill -f lynis

# Удалить lock-файл
sudo rm -f /var/run/lynis_telegram.lock

# Запустить заново
sudo ./scripts/lynis_with_telegram.sh
```

#### "Permission denied"
```bash
# Проверить права
ls -la scripts/lynis_with_telegram.sh

# Установить права
sudo chmod +x scripts/lynis_with_telegram.sh
```

### Решенные проблемы

#### ✅ Таймауты и блокировки
- **Проблема:** Аудит прерывался по таймауту через 10 минут
- **Решение:** Таймаут полностью убран из всех скриптов
- **Результат:** Lynis работает до полного завершения

#### ✅ Параллельные запуски
- **Проблема:** Systemd-таймер конфликтовал с ручными запусками
- **Решение:** Автоматический запуск удалён, только ручной
- **Результат:** Нет конфликтов между запусками

#### ✅ Lock-файлы
- **Проблема:** Lock-файлы "зависали" после прерываний
- **Решение:** Автоматическая очистка в скриптах
- **Результат:** Lock-файлы автоматически удаляются

#### ✅ Ротация логов
- **Проблема:** Логи накапливались и занимали место
- **Решение:** Настроен logrotate для автоматической ротации
- **Результат:** Логи автоматически архивируются

### Новые возможности

#### 🔒 Автоматическое исправление безопасности
```bash
# Исправить 47 проблем безопасности одним скриптом
sudo ./scripts/security_hardening.sh
```

#### 📊 Быстрый аудит без Telegram
```bash
# Быстрый аудит с анализом результатов
sudo ./scripts/quick_audit.sh quick
```

#### 🔄 Автоматическая ротация логов
- Логи автоматически ротируются через logrotate
- Старые логи архивируются и удаляются
- Настроен `/etc/logrotate.d/cloudbridge-relay`

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