# Мониторинг безопасности с Telegram уведомлениями

## Обзор

Скрипт `scripts/security_monitor_telegram.sh` обеспечивает непрерывный мониторинг различных событий безопасности и отправляет уведомления в Telegram только при обнаружении новых событий.

## Основные возможности

- 🔍 **Мониторинг fail2ban**: Отслеживание новых блокировок IP адресов
- 🔐 **Мониторинг аутентификации**: Обнаружение неудачных попыток входа
- ⚠️ **Системный мониторинг**: Отслеживание критических системных событий
- 💾 **Мониторинг ресурсов**: Проверка использования диска и памяти
- 📊 **Детальная статистика**: Подробная информация о каждом событии
- 🔄 **Отслеживание изменений**: Уведомления только о новых событиях

## Настройка

### 1. Настройка Telegram бота

Следуйте инструкциям из [telegram_notifications_setup.md](telegram_notifications_setup.md) для создания бота и получения токена.

### 2. Настройка скрипта

Откройте файл `scripts/security_monitor_telegram.sh` и измените строки:

```bash
TELEGRAM_BOT_TOKEN="ваш_токен_бота"
TELEGRAM_CHAT_ID="ваш_chat_id"
```

### 3. Установка зависимостей

```bash
sudo apt-get install jq
```

## Использование

### Однократная проверка
```bash
sudo ./scripts/security_monitor_telegram.sh --once
```

### Непрерывный мониторинг
```bash
sudo ./scripts/security_monitor_telegram.sh --continuous
```

### Показать статистику
```bash
sudo ./scripts/security_monitor_telegram.sh --stats
```

### Сброс состояния
```bash
sudo ./scripts/security_monitor_telegram.sh --reset
```

### Тестовая отправка
```bash
sudo ./scripts/security_monitor_telegram.sh --test
```

## Мониторинг событий

### Fail2ban блокировки

Скрипт отслеживает новые блокировки в `/var/log/fail2ban.log`:

```
🚨 Обнаружены новые блокировки fail2ban

📊 Статистика:
• Количество новых блокировок: 3
• Временной интервал: 2025-01-15T10:00:00 → 2025-01-15T10:05:00

🔍 Последние блокировки:
• 2025-01-15 10:04:23 fail2ban.actions: WARNING [sshd] Ban 192.168.1.100
• 2025-01-15 10:04:45 fail2ban.actions: WARNING [sshd] Ban 10.0.0.50

🖥️ Сервер: server-name
🕐 Время проверки: 2025-01-15 10:05:00
```

### Неудачные попытки входа

Мониторинг неудачных попыток аутентификации в `/var/log/auth.log`:

```
🔐 Обнаружены новые неудачные попытки входа

📊 Статистика:
• Количество неудачных попыток: 15
• Уникальных IP адресов: 8
• Временной интервал: 2025-01-15T10:00:00 → 2025-01-15T10:10:00

🔍 Последние попытки:
• Jan 15 10:09:45 server sshd[12345]: Failed password for invalid user admin from 192.168.1.100
• Jan 15 10:09:47 server sshd[12346]: Failed password for invalid user root from 10.0.0.50

🖥️ Сервер: server-name
🕐 Время проверки: 2025-01-15 10:10:00
```

### Критические системные события

Отслеживание ошибок и критических событий в `/var/log/syslog`:

```
⚠️ Обнаружены критические системные события

📊 Статистика:
• Количество критических событий: 2
• Временной интервал: 2025-01-15T10:00:00 → 2025-01-15T10:15:00

🔍 Последние события:
• Jan 15 10:14:30 server kernel: [ERROR] Out of memory: Kill process 1234
• Jan 15 10:14:35 server systemd: Failed to start service.service

🖥️ Сервер: server-name
🕐 Время проверки: 2025-01-15 10:15:00
```

### Мониторинг ресурсов

#### Высокое использование диска

```
💾 Высокое использование диска

📊 Статистика:
• Использование диска: 85%
• Пороговое значение: 80%

🔍 Детали:
/dev/sda1        20G   17G  2.0G  89% /

🖥️ Сервер: server-name
🕐 Время проверки: 2025-01-15 10:20:00
```

#### Высокое использование памяти

```
🧠 Высокое использование памяти

📊 Статистика:
• Использование памяти: 90%
• Пороговое значение: 85%

🔍 Детали:
              total        used        free      shared  buff/cache   available
Mem:           8.0G        7.2G        200M        100M        600M        500M

🖥️ Сервер: server-name
🕐 Время проверки: 2025-01-15 10:25:00
```

## Автоматизация

### Запуск как systemd сервис

Создайте файл сервиса:

```bash
sudo nano /etc/systemd/system/security-monitor.service
```

Добавьте содержимое:

```ini
[Unit]
Description=Security Monitor with Telegram Notifications
After=network.target

[Service]
Type=simple
User=root
ExecStart=/home/ubuntu/cloudbridge-relay-installer/scripts/security_monitor_telegram.sh --continuous
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

Активируйте сервис:

```bash
sudo systemctl daemon-reload
sudo systemctl enable security-monitor
sudo systemctl start security-monitor
```

### Запуск через cron

Добавьте в crontab для проверки каждые 5 минут:

```bash
sudo crontab -e
```

Добавьте строку:

```
*/5 * * * * /home/ubuntu/cloudbridge-relay-installer/scripts/security_monitor_telegram.sh --once
```

## Файлы состояния

Скрипт создает и использует следующие файлы в `/var/lib/security-monitor/`:

- `last_check.json` - время последней проверки каждого типа событий
- `events.log` - журнал всех обнаруженных событий

### Структура last_check.json

```json
{
    "fail2ban_last_check": "2025-01-15T10:00:00",
    "auth_last_check": "2025-01-15T10:00:00",
    "system_last_check": "2025-01-15T10:00:00",
    "last_events": {}
}
```

## Настройка пороговых значений

Вы можете изменить пороговые значения в скрипте:

```bash
# Порог использования диска (в процентах)
local threshold=80

# Порог использования памяти (в процентах)
local threshold=85
```

## Логирование

Все события логируются в файл `/var/lib/security-monitor/events.log`:

```
2025-01-15T10:05:00: Обнаружено 3 новых блокировок fail2ban
2025-01-15T10:10:00: Обнаружено 15 новых неудачных попыток входа
2025-01-15T10:15:00: Обнаружено 2 критических системных событий
2025-01-15T10:20:00: Высокое использование диска: 85%
```

## Устранение неполадок

### Проверка статуса сервиса

```bash
sudo systemctl status security-monitor
```

### Просмотр логов

```bash
sudo journalctl -u security-monitor -f
```

### Проверка файлов состояния

```bash
sudo ls -la /var/lib/security-monitor/
sudo cat /var/lib/security-monitor/events.log
```

### Сброс состояния

```bash
sudo ./scripts/security_monitor_telegram.sh --reset
```

### Тестирование уведомлений

```bash
sudo ./scripts/security_monitor_telegram.sh --test
```

## Интеграция с другими системами

### Интеграция с SIEM

Скрипт можно интегрировать с SIEM системами, добавив отправку событий в syslog:

```bash
logger -t security-monitor "Обнаружено $block_count новых блокировок fail2ban"
```

### Интеграция с Prometheus

Для интеграции с Prometheus можно добавить экспорт метрик:

```bash
# Создание метрик для Prometheus
echo "security_monitor_events_total{type=\"fail2ban\"} $block_count" >> /tmp/security_metrics.prom
```

## Безопасность

- Скрипт должен запускаться с правами root для доступа к логам
- Файлы состояния хранятся в защищенной директории `/var/lib/security-monitor/`
- Токен Telegram бота должен быть секретным
- Рекомендуется использовать переменные окружения для токенов в продакшене

## Расширение функциональности

Скрипт можно расширить для мониторинга:

- Изменений в конфигурационных файлах
- Новых процессов в системе
- Изменений в сетевых соединениях
- Обновлений системы безопасности
- Интеграции с антивирусными решениями
- Мониторинга Docker контейнеров
- Отслеживания изменений в файловой системе 

# Интеграция Suricata с Telegram-оповещениями

## Описание

Данная инструкция описывает, как настроить автоматическую отправку алертов Suricata в Telegram-чат в реальном времени с помощью bash-скрипта и systemd-сервиса.

---

## 1. Установка Suricata

```bash
sudo apt update
sudo apt install -y suricata jq
```

---

## 2. Настройка правил Suricata

- Все правила должны быть подключены в секции:
  ```yaml
  default-rule-path: /etc/suricata/rules
  rule-files:
    - suricata.rules
    - test.rules
  ```
- Пример тестового правила (ICMP):
  ```
  alert icmp any any -> any any (msg:"ICMP test alert"; sid:1000001; rev:1;)
  ```
  Сохранить в `/etc/suricata/rules/test.rules`.

---

## 3. Скрипт для отправки алертов в Telegram

**scripts/security/snort_suricata_alert.sh**
```bash
#!/bin/bash
EVE_FILE="/var/log/suricata/eve.json"
tail -F "$EVE_FILE" | \
while read -r line; do
    if echo "$line" | grep -q '"event_type":"alert"'; then
        TS=$(echo "$line" | jq -r '.timestamp // empty')
        SRC=$(echo "$line" | jq -r '.src_ip // empty')
        DST=$(echo "$line" | jq -r '.dest_ip // empty')
        PROTO=$(echo "$line" | jq -r '.proto // empty')
        SIG=$(echo "$line" | jq -r '.alert.signature // empty')
        SEV=$(echo "$line" | jq -r '.alert.severity // empty')
        MSG="[Suricata] $TS\n$SIG\n$SRC → $DST ($PROTO)\nSeverity: $SEV"
        $(dirname "$0")/telegram_notify.sh "$MSG"
    fi
done
```

**scripts/security/telegram_notify.sh**
```bash
#!/bin/bash
TOKEN="<ваш_токен>"
CHAT_ID="<ваш_chat_id>"
MESSAGE="$1"
curl -s -X POST "https://api.telegram.org/bot$TOKEN/sendMessage" \
    -d chat_id="$CHAT_ID" \
    -d text="$MESSAGE"
```

---

## 4. Systemd-сервис для автозапуска демона

**scripts/security/suricata-alert-telegram.service**
```ini
[Unit]
Description=Suricata alert to Telegram daemon
After=network.target suricata.service

[Service]
Type=simple
ExecStart=/bin/bash /home/ubuntu/cloudbridge-relay-installer/scripts/security/snort_suricata_alert.sh
Restart=always
User=ubuntu

[Install]
WantedBy=multi-user.target
```

Установить сервис:
```bash
sudo cp scripts/security/suricata-alert-telegram.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now suricata-alert-telegram.service
```

---

## 5. Тестирование

1. Сгенерируйте трафик (например, ping):
   ```bash
   ping -c 1 1.1.1.1
   ```
2. Проверьте, что в `/var/log/suricata/eve.json` появились алерты с вашим тестовым сообщением.
3. Убедитесь, что в Telegram-чат пришло уведомление с краткой информацией об инциденте.

---

## 6. Управление сервисом

- Проверить статус:
  ```bash
  sudo systemctl status suricata-alert-telegram.service
  ```
- Остановить:
  ```bash
  sudo systemctl stop suricata-alert-telegram.service
  ```
- Запустить:
  ```bash
  sudo systemctl start suricata-alert-telegram.service
  ```
- Логи:
  ```bash
  journalctl -u suricata-alert-telegram.service -f
  ```

---

## 7. Примечания

- Для работы требуется установленный jq.
- Можно доработать фильтрацию, форматирование сообщений, интеграцию с другими системами.
- Все скрипты должны быть исполняемыми (`chmod +x ...`).

---

**Интеграция Suricata + Telegram полностью автоматизирована и готова к эксплуатации!** 