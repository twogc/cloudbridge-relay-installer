# Security Scripts

Скрипты для обеспечения безопасности и аудита CloudBridge Relay.

## Основные скрипты безопасности

### `security_check.sh`
Комплексная проверка безопасности системы.

### `security_hardening.sh`
Усиление безопасности системы.

### `security_monitor.sh`
Мониторинг безопасности в реальном времени.

### `security_monitor_telegram.sh`
Мониторинг безопасности с уведомлениями в Telegram.

## Скрипты аудита

### `lynis_quick_test.sh`
Быстрый аудит безопасности с помощью Lynis.

### `lynis_with_telegram.sh`
Аудит безопасности Lynis с отправкой результатов в Telegram.

### `quick_audit.sh`
Быстрая проверка основных параметров безопасности.

### `relay-security-audit`
Основной скрипт аудита безопасности CloudBridge Relay.

## Использование

### Быстрая проверка безопасности:
```bash
./security/security_check.sh
```

### Полный аудит безопасности:
```bash
./security/relay-security-audit
```

### Мониторинг с Telegram уведомлениями:
```bash
./security/security_monitor_telegram.sh
```

### Усиление безопасности:
```bash
./security/security_hardening.sh
```

## Автоматизация

Рекомендуется настроить регулярные проверки через cron:
```bash
# Ежедневный аудит безопасности
0 2 * * * /path/to/scripts/security/relay-security-audit

# Ежечасный мониторинг
0 * * * * /path/to/scripts/security/security_monitor.sh
``` 

## Фильтрация алертов Telegram

Скрипт `security_monitor.sh` фильтрует сообщения, отправляемые в Telegram, чтобы не засорять канал некритичными или уже решёнными рекомендациями.

**Не отправляются сообщения, содержащие:**
- `emergency.service`
- `icmp_ignore_bogus_error_responses`
- `Install a file integrity tool`

Если сообщение содержит один из этих шаблонов, оно не отправляется в Telegram, а в лог пишется:
```
[INFO] Сообщение не отправлено в Telegram (игнорируется по фильтру)
```

Чтобы изменить фильтр, отредактируйте функцию `should_ignore_alert` в `security_monitor.sh`. 