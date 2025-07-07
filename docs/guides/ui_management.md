# CloudBridge Admin UI Management

## Автоматический запуск UI после перезагрузки

UI админ панель настроен для автоматического запуска после перезагрузки системы через systemd сервис.

### Управление UI сервисом

Используйте скрипт `scripts/setup/manage_ui.sh` для управления UI:

```bash
# Установить и включить сервис (требует sudo)
sudo ./scripts/setup/manage_ui.sh install

# Запустить UI
sudo ./scripts/setup/manage_ui.sh start

# Остановить UI
sudo ./scripts/setup/manage_ui.sh stop

# Перезапустить UI
sudo ./scripts/setup/manage_ui.sh restart

# Проверить статус
./scripts/setup/manage_ui.sh status

# Включить автозапуск (требует sudo)
sudo ./scripts/setup/manage_ui.sh enable

# Отключить автозапуск (требует sudo)
sudo ./scripts/setup/manage_ui.sh disable

# Показать логи
./scripts/setup/manage_ui.sh logs
```

### Ручной запуск (для разработки)

Если нужно запустить UI вручную для разработки:

```bash
cd ui-admin
npm run dev
```

### Доступ к UI

- **URL:** http://localhost:5173/
- **Порт:** 5173
- **Статус:** Автоматически запускается при загрузке системы

### Мониторинг

Проверить статус сервиса:
```bash
systemctl status ui-admin.service
```

Просмотр логов:
```bash
journalctl -u ui-admin.service -f
```

### Устранение неполадок

1. **UI не запускается:**
   ```bash
   sudo ./scripts/setup/manage_ui.sh restart
   ./scripts/setup/manage_ui.sh logs
   ```

2. **Порт занят:**
   ```bash
   netstat -tlnp | grep 5173
   sudo lsof -i :5173
   ```

3. **Зависимости не установлены:**
   ```bash
   cd ui-admin
   npm install
   ```

### Конфигурация

Сервис настроен в файле `systemd/ui-admin.service`:
- Работает от пользователя `ubuntu`
- Автоматический перезапуск при сбоях
- Логирование через systemd journal 