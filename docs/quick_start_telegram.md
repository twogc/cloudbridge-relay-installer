# Быстрый старт: Telegram уведомления

## Обзор

Это руководство поможет быстро настроить систему мониторинга безопасности с Telegram уведомлениями, которые отправляются только при обнаружении новых событий.

## Шаг 1: Создание Telegram бота

1. Откройте Telegram и найдите @BotFather
2. Отправьте `/newbot`
3. Введите имя бота (например, "Security Monitor")
4. Введите username (должен заканчиваться на 'bot')
5. Сохраните токен бота

## Шаг 2: Получение Chat ID

1. Найдите вашего бота по username
2. Отправьте любое сообщение (например, "/start")
3. Откройте в браузере: `https://api.telegram.org/bot<ВАШ_ТОКЕН>/getUpdates`
4. Найдите `"chat":{"id":123456789}` - это ваш Chat ID

## Шаг 3: Настройка токенов

Отредактируйте файлы и укажите ваши данные:

```bash
# В файле scripts/lynis_with_telegram.sh
TELEGRAM_BOT_TOKEN="ваш_токен_бота"
TELEGRAM_CHAT_ID="ваш_chat_id"

# В файле scripts/security_monitor_telegram.sh
TELEGRAM_BOT_TOKEN="ваш_токен_бота"
TELEGRAM_CHAT_ID="ваш_chat_id"
```

## Шаг 4: Установка системы

```bash
sudo ./scripts/setup_security_monitor.sh
```

## Шаг 5: Тестирование

```bash
# Тест Lynis мониторинга
sudo ./scripts/lynis_with_telegram.sh --test

# Тест Security Monitor
sudo ./scripts/security_monitor_telegram.sh --test
```

## Что происходит после установки

### Автоматический мониторинг
- **Security Monitor**: Проверка каждые 5 минут
  - Новые блокировки fail2ban
  - Неудачные попытки входа
  - Критические системные события
  - Высокое использование ресурсов

- **Lynis Audit**: Ежедневный аудит безопасности
  - Полный анализ безопасности системы
  - Отслеживание изменений в результатах
  - Уведомления только при обнаружении новых проблем

### Уведомления отправляются только при:
- ✅ Изменении Hardening Index
- ✅ Новых проблемах безопасности
- ✅ Новых блокировках IP
- ✅ Критических системных событиях
- ✅ Превышении порогов ресурсов

## Управление системой

### Просмотр статуса
```bash
sudo ./scripts/setup_security_monitor.sh --status
```

### Просмотр логов
```bash
# Логи Security Monitor
sudo journalctl -u security-monitor -f

# Логи Lynis Audit
sudo journalctl -u lynis-security-audit -f

# События безопасности
sudo cat /var/lib/security-monitor/events.log

# Изменения Lynis
sudo cat /var/lib/lynis-telegram/changes.log
```

### Управление сервисами
```bash
# Запуск/остановка Security Monitor
sudo systemctl start/stop/restart security-monitor

# Запуск/остановка Lynis Audit
sudo systemctl start/stop/restart lynis-security-audit.timer
```

## Примеры уведомлений

### Новые блокировки fail2ban
```
🚨 Обнаружены новые блокировки fail2ban

📊 Статистика:
• Количество новых блокировок: 3
• Временной интервал: 2025-01-15T10:00:00 → 2025-01-15T10:05:00

🔍 Последние блокировки:
• 2025-01-15 10:04:23 fail2ban.actions: WARNING [sshd] Ban 192.168.1.100
```

### Изменения в безопасности Lynis
```
🔒 Lynis Security Audit завершен

📊 Результаты:
• Hardening Index: 75/100
• Предупреждений: 3
• Рекомендаций: 12
• Проблем безопасности: 1

🔄 Обнаружены изменения:
• Hardening Index: 70 → 75
• Проблемы безопасности: 2 → 1
```

## Устранение неполадок

### Бот не отвечает
1. Отправьте `/start` боту
2. Проверьте правильность токена и Chat ID
3. Убедитесь, что бот добавлен в чат

### Нет уведомлений
1. Проверьте статус сервисов: `sudo ./scripts/setup_security_monitor.sh --status`
2. Проверьте логи: `sudo journalctl -u security-monitor -f`
3. Сбросьте состояние: `sudo ./scripts/security_monitor_telegram.sh --reset`

### Ошибки установки
1. Убедитесь, что запускаете с правами root
2. Проверьте интернет-соединение
3. Установите зависимости: `sudo apt-get install jq curl fail2ban`

## Дополнительные возможности

### Настройка пороговых значений
Отредактируйте скрипты для изменения порогов:
- Использование диска: 80%
- Использование памяти: 85%
- Интервалы проверки: 5 минут

### Интеграция с SIEM
Добавьте отправку событий в syslog для интеграции с SIEM системами.

### Расширение мониторинга
Добавьте мониторинг:
- Docker контейнеров
- Сетевых соединений
- Конфигурационных файлов
- Обновлений системы

## Безопасность

- Токены Telegram должны быть секретными
- Файлы состояния хранятся в защищенных директориях
- Сервисы запускаются с ограниченными правами
- Рекомендуется использовать переменные окружения в продакшене

## Поддержка

- Документация: `docs/telegram_notifications_setup.md`
- Мониторинг безопасности: `docs/security_monitoring_telegram.md`
- Логи: `/var/lib/security-monitor/` и `/var/lib/lynis-telegram/` 