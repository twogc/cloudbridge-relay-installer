# Обновление start_production.sh для поддержки темы Keycloak

## Обзор изменений

Скрипт `start_production.sh` был обновлен для автоматической поддержки кастомной темы Playa для Keycloak.

## Новые функции

### 1. Автоматическая настройка темы
```bash
# Автоматически запускается после старта сервисов
./scripts/setup/setup_keycloak_theme.sh
```

### 2. Проверка темы
```bash
# Проверяет наличие и корректность файлов темы
check_keycloak_theme()
```

### 3. Создание директорий
```bash
# Автоматически создает папку для тем
mkdir -p keycloak/themes
```

## Изменения в скрипте

### Версия обновлена до v2.1
- Добавлена строка "Custom Keycloak Theme | Playa Design System" в баннер
- Обновлена версия с v2.0 на v2.1

### Новые блоки кода

#### 1. Проверка темы при запуске
```bash
check_keycloak_theme
```

#### 2. Автоматическая настройка после старта сервисов
```bash
# --- Новый блок: автоматическая настройка темы Keycloak ---
log "Настройка кастомной темы Playa для Keycloak..."
if [ -f "scripts/setup/setup_keycloak_theme.sh" ]; then
    chmod +x scripts/setup/setup_keycloak_theme.sh
    if [[ -n "$KEYCLOAK_ADMIN" && -n "$KEYCLOAK_ADMIN_PASSWORD" ]]; then
        log "Настраиваю тему Playa для Keycloak..."
        ./scripts/setup/setup_keycloak_theme.sh
    else
        warn "KEYCLOAK_ADMIN или KEYCLOAK_ADMIN_PASSWORD не установлены"
    fi
fi
```

#### 3. Обновленный статус
```bash
log "Тема Keycloak:"
echo -e "  \033[1;36m🎨 Кастомная тема:\033[0m Playa (автоматически настроена)"
echo -e "  \033[1;36m🔗 Страница входа:\033[0m http://localhost:8080/realms/cloudbridge/protocol/openid-connect/auth"
```

#### 4. Новый Quick Tip
```bash
echo -e "  • Use \033[1;33m./scripts/setup/setup_keycloak_theme.sh\033[0m to reconfigure Keycloak theme"
```

## Функция check_keycloak_theme()

Проверяет:
- ✅ Наличие папки `keycloak/themes/playa`
- ✅ Количество файлов темы
- ✅ Основной файл конфигурации `theme.properties`
- ✅ CSS файл `style.css`
- ✅ Шаблон входа `index.ftl`

## Автоматизация

### Что происходит автоматически:
1. **При запуске:** Проверка наличия темы
2. **После старта сервисов:** Автоматическая настройка темы в Keycloak
3. **В статусе:** Отображение информации о теме

### Требования для автоматической настройки:
- Переменные `KEYCLOAK_ADMIN` и `KEYCLOAK_ADMIN_PASSWORD` должны быть установлены
- Скрипт `scripts/setup/setup_keycloak_theme.sh` должен существовать
- Тема должна быть в папке `keycloak/themes/playa`

## Использование

### Обычный запуск
```bash
./start_production.sh
```

### Запуск с проверкой статуса
```bash
./start_production.sh --status
```

### Принудительный перезапуск
```bash
./start_production.sh --force
```

## Логирование

### Новые сообщения:
- `[INFO] Проверка кастомной темы Keycloak...`
- `[INFO] ✅ Кастомная тема Playa найдена (X файлов)`
- `[INFO] Настройка кастомной темы Playa для Keycloak...`
- `[WARN] ⚠️ Кастомная тема Playa не найдена`

## Обратная совместимость

✅ **Совместимость сохранена:**
- Все существующие опции работают
- Старые конфигурации продолжают работать
- Тема не является обязательной для работы системы

## Тестирование

### Проверка обновления:
```bash
# Проверка справки
./start_production.sh --help

# Проверка статуса
./start_production.sh --status

# Проверка синтаксиса
bash -n start_production.sh
```

## Заключение

Обновление `start_production.sh` обеспечивает:
- 🔄 Автоматическую настройку темы Keycloak
- ✅ Проверку корректности файлов темы
- 📊 Информативный статус с данными о теме
- 🛡️ Обратную совместимость
- 📝 Подробное логирование

Система готова к использованию с кастомной темой Playa! 