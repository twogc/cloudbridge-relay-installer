# Changelog - Security Monitoring System

## [2.0.0] - 2025-07-27

### Добавлено
- **Новый скрипт `enhanced_security_monitor.sh`** с конфигурационным файлом
- **Система конфигурации** `security_monitor_config.sh` для управления настройками
- **Расширенные фильтры** для исключения некритических ошибок Docker
- **Документация** `README_SECURITY_MONITORING.md` с подробными инструкциями

### Исправлено
- **Ложные уведомления** об ошибках Docker больше не отправляются
- **Фильтрация ошибок** теперь включает все типы некритических ошибок Docker:
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

### Улучшено
- **Гибкость настройки** через конфигурационный файл
- **Обработка ошибок** в скриптах
- **Логирование** и отладка
- **Документация** и инструкции по использованию

### Изменено
- **Структура фильтров** - теперь фильтры настраиваются через конфигурационный файл
- **Пороговые значения** - вынесены в конфигурацию
- **Модули мониторинга** - можно включать/отключать отдельно

## [1.0.0] - Исходная версия

### Функции
- Базовый мониторинг безопасности
- Уведомления в Telegram
- Проверка критических сервисов
- Мониторинг логов
- Детекция подозрительной активности

---

## Рекомендации по обновлению

### Для пользователей версии 1.0.0:

1. **Создайте конфигурационный файл:**
   ```bash
   sudo bash scripts/security/security_monitor_config.sh --create
   ```

2. **Настройте Telegram параметры:**
   ```bash
   sudo bash scripts/security/security_monitor_config.sh --edit
   ```

3. **Протестируйте новый скрипт:**
   ```bash
   sudo bash scripts/security/enhanced_security_monitor.sh --test
   sudo bash scripts/security/enhanced_security_monitor.sh --once
   ```

4. **Переключитесь на улучшенный скрипт:**
   - Замените вызовы `security_monitor.sh` на `enhanced_security_monitor.sh`
   - Обновите cron задачи и systemd сервисы

### Обратная совместимость
- Оригинальный `security_monitor.sh` продолжает работать
- Все фильтры применены и к оригинальному скрипту
- Старые конфигурации остаются совместимыми

---

## Планы на будущее

### Версия 2.1.0 (планируется)
- Интеграция с SIEM системами
- Расширенная аналитика угроз
- Машинное обучение для детекции аномалий
- Веб-интерфейс для управления

### Версия 2.2.0 (планируется)
- Поддержка множественных серверов
- Централизованное управление
- API для интеграции
- Расширенные отчеты 