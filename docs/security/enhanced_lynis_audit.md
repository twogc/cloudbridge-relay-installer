# Расширенный аудит безопасности Lynis

## Обзор

Расширенный аудит безопасности Lynis для CloudBridge Relay предоставляет дополнительные возможности мониторинга и анализа безопасности системы, выходящие за рамки базового аудита.

## Возможности

### 🔍 **Дополнительные проверки**

1. **Аудит Docker контейнеров**
   - Проверка привилегированных контейнеров
   - Анализ контейнеров с root пользователем
   - Сканирование образов Docker
   - Мониторинг безопасности контейнеров

2. **Сетевая безопасность**
   - Анализ открытых портов
   - Проверка статуса firewall (UFW)
   - Мониторинг SSH попыток
   - Обнаружение подозрительной активности

3. **Системные сервисы**
   - Аудит systemd сервисов
   - Проверка PAM конфигурации
   - Анализ запущенных процессов
   - Мониторинг демонов

4. **Файловая система**
   - Проверка прав доступа
   - Анализ точек монтирования
   - Сканирование домашних директорий
   - Проверка системных бинарников

### 📊 **Улучшенная отчетность**

- Детальные отчеты в JSON формате
- Сравнение результатов между запусками
- Отслеживание изменений
- Уведомления только о новых проблемах

### 🔧 **Автоматизация**

- Еженедельный автоматический аудит
- Ротация логов
- Интеграция с Telegram
- Systemd сервисы и таймеры

## Установка

### Быстрая установка

```bash
# Установка расширенного аудита
sudo bash scripts/setup/setup_enhanced_lynis.sh install

# Проверка статуса
sudo bash scripts/setup/setup_enhanced_lynis.sh status

# Тестирование
sudo bash scripts/setup/setup_enhanced_lynis.sh test
```

### Ручная установка

1. **Установка зависимостей**
   ```bash
   sudo apt update
   sudo apt install jq curl
   ```

2. **Установка профиля**
   ```bash
   sudo mkdir -p /etc/lynis
   sudo cp etc/lynis/enhanced_custom.prf /etc/lynis/
   sudo chmod 644 /etc/lynis/enhanced_custom.prf
   ```

3. **Установка systemd сервисов**
   ```bash
   sudo cp systemd/enhanced-lynis-audit.service /etc/systemd/system/
   sudo cp systemd/enhanced-lynis-audit.timer /etc/systemd/system/
   sudo systemctl daemon-reload
   sudo systemctl enable enhanced-lynis-audit.timer
   ```

## Использование

### Запуск вручную

```bash
# Полный расширенный аудит
sudo bash scripts/security/enhanced_lynis_audit.sh

# Тестовый режим
sudo bash scripts/security/enhanced_lynis_audit.sh --test
```

### Автоматический запуск

```bash
# Статус таймера
sudo systemctl status enhanced-lynis-audit.timer

# Запуск вручную
sudo systemctl start enhanced-lynis-audit.service

# Просмотр логов
sudo journalctl -u enhanced-lynis-audit -f
```

## Конфигурация

### Профиль Lynis

Файл: `/etc/lynis/enhanced_custom.prf`

Основные настройки:
- `auditor="CloudBridge Relay"` - имя аудитора
- `plugin=pam,systemd,docker,network` - включенные плагины
- `QUICK=1` - быстрый режим
- `TIMEOUT=300` - таймаут в секундах

### Telegram уведомления

Настройки в скрипте `enhanced_lynis_audit.sh`:
```bash
TELEGRAM_BOT_TOKEN="ваш_токен"
TELEGRAM_CHAT_ID="ваш_chat_id"
```

## Структура файлов

```
/var/log/lynis/
├── lynis.log              # Основной лог
├── reports/               # Отчеты
│   └── lynis_report_*.dat
└── lynis-report.dat       # Последний отчет

/var/lib/lynis-telegram/
├── current_results.json   # Текущие результаты
├── previous_results.json  # Предыдущие результаты
└── changes.log           # Лог изменений

/etc/lynis/
└── enhanced_custom.prf   # Расширенный профиль
```

## Отчеты и логи

### Анализ результатов

```bash
# Просмотр последнего отчета
sudo cat /var/log/lynis/lynis-report.dat

# Анализ JSON результатов
sudo cat /var/lib/lynis-telegram/current_results.json | jq '.'

# Просмотр изменений
sudo cat /var/lib/lynis-telegram/changes.log
```

### Ключевые метрики

- **Hardening Index** - общий индекс безопасности (0-100)
- **Warnings** - количество предупреждений
- **Suggestions** - количество рекомендаций
- **Security Issues** - проблемы безопасности
- **Critical Issues** - критические проблемы

## Мониторинг

### Telegram уведомления

Система отправляет уведомления в следующих случаях:

1. **Критические проблемы безопасности**
2. **Новые проблемы обнаружены**
3. **Изменения в Hardening Index**
4. **Проблемы с Docker контейнерами**
5. **Подозрительная сетевая активность**
6. **Доступны обновления Lynis**

### Пример уведомления

```
[LYNIS] Критические проблемы безопасности

[CRITICAL] Обнаружено:
• SSH-7408: SSH configuration file has weak permissions
• AUTH-9328: Password aging is not enabled
• FILE-7524: World writable file found

[SCORE] Hardening Index: 65
[TIME] Время: 2025-07-30 04:32:49
[SERVER] Сервер: 2gc-relay-server
```

## Устранение неполадок

### Проблемы с запуском

```bash
# Проверка статуса сервиса
sudo systemctl status enhanced-lynis-audit.service

# Просмотр логов
sudo journalctl -u enhanced-lynis-audit -n 50

# Проверка прав доступа
ls -la /var/log/lynis/
ls -la /var/lib/lynis-telegram/
```

### Проблемы с Telegram

```bash
# Проверка токена и chat_id
grep -E "TELEGRAM_BOT_TOKEN|TELEGRAM_CHAT_ID" scripts/security/enhanced_lynis_audit.sh

# Тест отправки
curl -s -X POST "https://api.telegram.org/bot${TOKEN}/sendMessage" \
  -d "chat_id=${CHAT_ID}" \
  -d "text=Test message"
```

### Проблемы с профилем

```bash
# Проверка синтаксиса профиля
lynis show profiles

# Тест профиля
lynis audit system --profile /etc/lynis/enhanced_custom.prf --quick
```

## Обновления

### Обновление Lynis

```bash
# Проверка обновлений
lynis --check-update

# Обновление из исходников
cd /tmp
git clone https://github.com/CISOfy/lynis.git
cd lynis
sudo cp -r * /usr/local/lynis/
```

### Обновление профиля

```bash
# Резервная копия
sudo cp /etc/lynis/enhanced_custom.prf /etc/lynis/enhanced_custom.prf.backup

# Обновление
sudo cp etc/lynis/enhanced_custom.prf /etc/lynis/
sudo systemctl restart enhanced-lynis-audit.service
```

## Безопасность

### Рекомендации

1. **Регулярные обновления**
   - Обновляйте Lynis до последней версии
   - Следите за обновлениями профиля

2. **Мониторинг логов**
   - Регулярно проверяйте логи
   - Настройте ротацию логов

3. **Права доступа**
   - Ограничьте доступ к логам
   - Используйте правильные права доступа

4. **Резервное копирование**
   - Регулярно создавайте резервные копии конфигурации
   - Сохраняйте важные отчеты

### Ограничения

- Требует root привилегий для полного аудита
- Может занять значительное время на больших системах
- Некоторые проверки могут быть ложными срабатываниями

## Поддержка

### Полезные команды

```bash
# Справка Lynis
lynis --help

# Показать профили
lynis show profiles

# Показать настройки
lynis show settings

# Проверка версии
lynis --version
```

### Документация

- [Официальная документация Lynis](https://cisofy.com/lynis/)
- [GitHub репозиторий](https://github.com/CISOfy/lynis)
- [Руководство по безопасности](https://cisofy.com/lynis/security/)

### Сообщество

- [Форум Lynis](https://cisofy.com/community/)
- [GitHub Issues](https://github.com/CISOfy/lynis/issues)
- [Telegram канал](https://t.me/cloudbridge_relay) 