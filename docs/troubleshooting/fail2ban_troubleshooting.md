# Устранение неполадок Fail2Ban

## Проблема: "Have not found any log file for relay jail"

### Описание проблемы
Fail2Ban не может найти лог-файл для jail "relay" и выдает ошибку:
```
ERROR   Failed during configuration: Have not found any log file for relay jail
```

### Причина
Лог-файл relay находится в `/home/ubuntu/cloudbridge-relay-installer/logs/relay.log`, но fail2ban ищет его в `/var/log/relay/relay.log`.

### Решение

#### Автоматическое решение
Запустите скрипт настройки лог-файлов:
```bash
sudo ./scripts/setup_fail2ban_logs.sh
```

#### Ручное решение
1. Создайте директорию для логов:
```bash
sudo mkdir -p /var/log/relay
```

2. Создайте символическую ссылку:
```bash
sudo ln -sf /home/ubuntu/cloudbridge-relay-installer/logs/relay.log /var/log/relay/relay.log
```

3. Перезапустите fail2ban:
```bash
sudo systemctl restart fail2ban
```

### Проверка решения
```bash
# Проверка статуса fail2ban
sudo systemctl status fail2ban

# Проверка статуса relay jail
sudo fail2ban-client status relay

# Проверка символической ссылки
ls -la /var/log/relay/
```

## Полезные команды

### Управление fail2ban
```bash
# Статус всех jail'ов
sudo fail2ban-client status

# Статус конкретного jail
sudo fail2ban-client status relay

# Разблокировать IP
sudo fail2ban-client set relay unbanip <IP_ADDRESS>

# Просмотр логов fail2ban
sudo tail -f /var/log/fail2ban.log
```

### Проверка конфигурации
```bash
# Просмотр конфигурации jail
sudo cat /etc/fail2ban/jail.local

# Просмотр фильтра
sudo cat /etc/fail2ban/filter.d/relay.conf

# Проверка путей к лог-файлам
sudo fail2ban-client get relay logpath
```

## Фильтр для JSON логов

Relay использует JSON формат логов. Фильтр настроен для поиска следующих ошибок:
- `"authentication failed"`
- `"invalid credentials"`
- `"unauthorized access"`

Пример JSON лога:
```json
{"timestamp":"2025-06-14T23:38:38Z","level":"ERROR","message":"authentication failed","fields":{"remote_addr":{"IP":"192.168.1.100","Port":53270,"Zone":""}}}
```

## Автоматическое восстановление

Для автоматического восстановления символической ссылки при перезагрузке системы добавьте в crontab:
```bash
# Добавить в crontab (sudo crontab -e)
@reboot /home/ubuntu/cloudbridge-relay-installer/scripts/setup_fail2ban_logs.sh
```

## Мониторинг

Для мониторинга работы fail2ban можно использовать:
```bash
# Мониторинг в реальном времени
sudo tail -f /var/log/fail2ban.log

# Статистика блокировок
sudo fail2ban-client status relay

# Проверка заблокированных IP
sudo iptables -L -n | grep f2b-relay
``` 