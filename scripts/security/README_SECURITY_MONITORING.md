# CloudBridge Relay - Security Monitoring

## Обзор

Система мониторинга безопасности для CloudBridge Relay с уведомлениями в Telegram. Включает в себя несколько скриптов для различных сценариев мониторинга.

## Проблема и решение

### Проблема
Скрипты мониторинга отправляли ложные уведомления об ошибках Docker, которые не являются критическими для безопасности:
- `ShouldRestart failed, container will not be restarted`
- `restart canceled`
- `Error handling plugin refcount operation`
- `legacy plugin: Unable to find plugin without name`
- `Error running exec`
- `stream copy error`
- `executable file not found`
- `unable to start container process`
- `cleanup warnings`
- `failed to remove runc container`
- `runc did not terminate successfully`

### Решение
Добавлены расширенные фильтры для исключения некритических ошибок Docker и создана конфигурационная система для гибкой настройки.

## Скрипты

### 1. `security_monitor.sh` - Основной скрипт мониторинга
**Функции:**
- Мониторинг блокировок Fail2Ban
- Проверка критических сервисов
- Мониторинг дискового пространства
- Проверка обновлений системы
- Детекция подозрительной активности
- Проверка SSL сертификатов
- Анализ критических логов
- Мониторинг логов UFW

**Использование:**
```bash
# Разовая проверка
sudo bash scripts/security/security_monitor.sh --once

# Непрерывный мониторинг
sudo bash scripts/security/security_monitor.sh --continuous

# Тестовая отправка
sudo bash scripts/security/security_monitor.sh --test
```

### 2. `enhanced_security_monitor.sh` - Улучшенный скрипт
**Особенности:**
- Использует конфигурационный файл
- Расширенные фильтры для ошибок Docker
- Более гибкая настройка
- Лучшая обработка ошибок

**Использование:**
```bash
# Разовая проверка
sudo bash scripts/security/enhanced_security_monitor.sh --once

# Непрерывный мониторинг
sudo bash scripts/security/enhanced_security_monitor.sh --continuous

# Тестовая отправка
sudo bash scripts/security/enhanced_security_monitor.sh --test

# Показать конфигурацию
sudo bash scripts/security/enhanced_security_monitor.sh --config
```

### 3. `security_monitor_telegram.sh` - Telegram-ориентированный мониторинг
**Особенности:**
- Отслеживание только новых событий
- Сохранение состояния между запусками
- Детальная статистика событий

**Использование:**
```bash
# Однократная проверка
sudo bash scripts/security/security_monitor_telegram.sh --once

# Непрерывный мониторинг
sudo bash scripts/security/security_monitor_telegram.sh --continuous

# Сброс состояния
sudo bash scripts/security/security_monitor_telegram.sh --reset

# Показать статистику
sudo bash scripts/security/security_monitor_telegram.sh --stats
```

### 4. `security_monitor_config.sh` - Управление конфигурацией
**Функции:**
- Создание конфигурационного файла
- Редактирование настроек
- Валидация конфигурации
- Просмотр фильтров

**Использование:**
```bash
# Создать конфигурацию по умолчанию
sudo bash scripts/security/security_monitor_config.sh --create

# Показать текущую конфигурацию
sudo bash scripts/security/security_monitor_config.sh --show

# Редактировать конфигурацию
sudo bash scripts/security/security_monitor_config.sh --edit

# Проверить конфигурацию
sudo bash scripts/security/security_monitor_config.sh --validate

# Показать фильтры
sudo bash scripts/security/security_monitor_config.sh --filters
```

### 5. `telegram_notify.sh` - Утилита отправки уведомлений
**Использование:**
```bash
# Отправить сообщение
bash scripts/security/telegram_notify.sh "Ваше сообщение"

# Отправить HTML-сообщение
bash scripts/security/telegram_notify.sh "<b>Жирный текст</b>" HTML
```

## Конфигурация

### Файл конфигурации: `/etc/cloudbridge/security_monitor.conf`

**Основные настройки:**
```bash
# Telegram настройки
TELEGRAM_BOT_TOKEN="your_bot_token"
TELEGRAM_CHAT_ID="your_chat_id"
CHECK_INTERVAL=300

# Пороговые значения
DISK_USAGE_THRESHOLD=85
MEMORY_USAGE_THRESHOLD=85
SSH_FAILURES_THRESHOLD=50
ROOT_PROCESSES_THRESHOLD=150
SYSTEM_UPDATES_THRESHOLD=10

# Включение/отключение модулей
ENABLE_FAIL2BAN_MONITORING=true
ENABLE_CRITICAL_SERVICES_MONITORING=true
ENABLE_DISK_SPACE_MONITORING=true
ENABLE_SYSTEM_UPDATES_MONITORING=true
ENABLE_SUSPICIOUS_ACTIVITY_MONITORING=true
ENABLE_SSL_CERTIFICATE_MONITORING=true
ENABLE_CRITICAL_LOGS_MONITORING=true
ENABLE_FIREWALL_LOGS_MONITORING=true
```

### Фильтры для исключения ошибок

**Docker фильтры:**
```bash
DOCKER_SHOULDRESTART_FILTER="dockerd.*ShouldRestart failed"
DOCKER_RESTART_CANCELED_FILTER="dockerd.*restart canceled"
DOCKER_PLUGIN_ERROR_FILTER="dockerd.*Error handling plugin refcount operation"
DOCKER_LEGACY_PLUGIN_FILTER="dockerd.*legacy plugin: Unable to find plugin without name"
DOCKER_CONTAINER_RESTART_FILTER="dockerd.*container will not be restarted"
DOCKER_MANUALLY_STOPPED_FILTER="dockerd.*hasBeenManuallyStopped=true"
DOCKER_DAEMON_SHUTDOWN_FILTER="dockerd.*daemonShuttingDown=false"
DOCKER_EXEC_ERROR_FILTER="dockerd.*Error running exec"
DOCKER_STREAM_COPY_ERROR_FILTER="dockerd.*stream copy error"
DOCKER_EXECUTABLE_NOT_FOUND_FILTER="dockerd.*executable file not found"
DOCKER_CONTAINER_PROCESS_ERROR_FILTER="dockerd.*unable to start container process"
CONTAINERD_CLEANUP_WARNINGS_FILTER="containerd.*cleanup warnings"
CONTAINERD_RUNC_REMOVE_ERROR_FILTER="containerd.*failed to remove runc container"
CONTAINERD_RUNC_TERMINATE_ERROR_FILTER="containerd.*runc did not terminate successfully"
```

## Установка и настройка

### 1. Создание конфигурации
```bash
sudo bash scripts/security/security_monitor_config.sh --create
```

### 2. Настройка Telegram
Отредактируйте файл конфигурации:
```bash
sudo bash scripts/security/security_monitor_config.sh --edit
```

Установите ваши значения:
- `TELEGRAM_BOT_TOKEN` - токен вашего бота
- `TELEGRAM_CHAT_ID` - ID чата для уведомлений

### 3. Тестирование
```bash
# Тест отправки
sudo bash scripts/security/enhanced_security_monitor.sh --test

# Тест проверки
sudo bash scripts/security/enhanced_security_monitor.sh --once
```

## Автоматизация

### Systemd сервис
Создайте файл `/etc/systemd/system/security-monitor.service`:
```ini
[Unit]
Description=CloudBridge Security Monitor
After=network.target

[Service]
Type=simple
User=root
ExecStart=/bin/bash /path/to/cloudbridge-relay-installer/scripts/security/enhanced_security_monitor.sh --continuous
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

**Активация сервиса:**
```bash
sudo systemctl daemon-reload
sudo systemctl enable security-monitor
sudo systemctl start security-monitor
```

### Cron задача
Добавьте в crontab:
```bash
# Проверка каждые 5 минут
*/5 * * * * /bin/bash /path/to/cloudbridge-relay-installer/scripts/security/enhanced_security_monitor.sh --once
```

## Логирование

### Логи скриптов
- `logs/security-monitor.log` - основной лог
- `logs/security-alerts.log` - лог алертов
- `/var/lib/security-monitor/` - состояние мониторинга

### Просмотр логов
```bash
# Последние записи
tail -f logs/security-monitor.log

# Поиск ошибок
grep ERROR logs/security-monitor.log

# Статистика событий
sudo bash scripts/security/security_monitor_telegram.sh --stats
```

## Устранение неполадок

### Проблема: Не отправляются уведомления в Telegram
**Решение:**
1. Проверьте токен бота и ID чата
2. Убедитесь, что бот добавлен в чат
3. Проверьте интернет-соединение

### Проблема: Ложные срабатывания
**Решение:**
1. Добавьте новые фильтры в конфигурацию
2. Используйте `--config` для просмотра текущих фильтров
3. Отредактируйте пороговые значения

### Проблема: Высокая нагрузка на систему
**Решение:**
1. Увеличьте интервал проверки (`CHECK_INTERVAL`)
2. Отключите ненужные модули мониторинга
3. Используйте `--once` вместо `--continuous`

## Безопасность

### Рекомендации
1. Используйте отдельного бота для уведомлений
2. Ограничьте доступ к конфигурационному файлу
3. Регулярно обновляйте фильтры
4. Мониторьте логи на предмет подозрительной активности

### Права доступа
```bash
# Установка правильных прав
sudo chmod 600 /etc/cloudbridge/security_monitor.conf
sudo chown root:root /etc/cloudbridge/security_monitor.conf
```

## Обновления

### Добавление новых фильтров
1. Отредактируйте конфигурационный файл
2. Добавьте новый фильтр в формате `NAME_FILTER="pattern"`
3. Обновите скрипты при необходимости
4. Протестируйте изменения

### Обновление скриптов
```bash
# Создать резервную копию
sudo cp /etc/cloudbridge/security_monitor.conf /etc/cloudbridge/security_monitor.conf.backup

# Обновить конфигурацию
sudo bash scripts/security/security_monitor_config.sh --reset
```

## Поддержка

При возникновении проблем:
1. Проверьте логи: `tail -f logs/security-monitor.log`
2. Запустите тест: `sudo bash scripts/security/enhanced_security_monitor.sh --test`
3. Проверьте конфигурацию: `sudo bash scripts/security/security_monitor_config.sh --validate`
4. Создайте issue в репозитории с описанием проблемы 