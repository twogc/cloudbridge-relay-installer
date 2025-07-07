# Активация темы Playa для Keycloak

## Быстрый старт

### 1. Перезапуск контейнеров
```bash
# Остановка контейнеров
docker-compose down

# Запуск с новой темой
docker-compose up -d
```

### 2. Настройка темы
```bash
# Установка переменных окружения
export KEYCLOAK_ADMIN=admin
export KEYCLOAK_ADMIN_PASSWORD=your-password

# Запуск скрипта настройки
./scripts/setup/setup_keycloak_theme.sh
```

### 3. Проверка
1. Откройте Keycloak Admin Console: http://localhost:8080
2. Войдите как администратор
3. Перейдите в Realm Settings > Themes
4. Выберите "playa" для Login Theme
5. Сохраните настройки

## Структура переноса

✅ **Выполнено:**
- Перенос файлов из `keycloak-vue/src` в `keycloak/themes/playa`
- Создание правильной структуры темы Keycloak
- Компиляция SCSS в CSS
- Обновление `docker-compose.yml`
- Создание скрипта автоматической настройки

## Файлы темы

```
keycloak/themes/playa/
├── theme.properties          # Основная конфигурация
├── login/
│   ├── css/style.css        # Скомпилированные стили
│   ├── js/theme.js          # JavaScript функциональность
│   ├── fonts/               # Шрифты Roboto
│   ├── img/                 # Изображения и иконки
│   ├── messages/            # Локализация
│   ├── index.ftl           # Шаблон входа
│   └── template.ftl        # Базовый шаблон
└── register/
    └── register.ftl        # Шаблон регистрации
```

## Особенности темы

- **Цветовая схема:** Синий (#004EA8) с адаптивным дизайном
- **Шрифты:** Roboto (Regular, Medium, Bold)
- **Адаптивность:** Desktop, Laptop, Tablet, Mobile
- **Совместимость:** Keycloak 21+

## Устранение неполадок

### Тема не отображается
```bash
# Проверка логов Keycloak
docker-compose logs keycloak

# Проверка структуры файлов
sudo find keycloak/themes/playa -type f
```

### Ошибки в стилях
1. Откройте консоль браузера (F12)
2. Проверьте вкладку Console на ошибки
3. Проверьте вкладку Network на 404 ошибки

## Документация

Подробная документация: `docs/guides/keycloak_theme_migration.md`

## Контакты

При возникновении вопросов обращайтесь к команде разработки CloudBridge Relay. 