# Использование Vue темы для Keycloak

## Быстрый старт

### 1. Внесение изменений
Редактируйте файлы в папке `keycloak-vue/src/`:
- Vue компоненты: `keycloak-vue/src/views/`
- Стили: `keycloak-vue/src/scss/`
- Ресурсы: `keycloak-vue/src/resources/`

### 2. Сборка темы
```bash
# Простая сборка
./scripts/setup/build_keycloak_theme.sh

# Сборка с автоматическим применением темы
./scripts/setup/build_keycloak_theme.sh --restart
```

### 3. Проверка результата
Откройте страницу входа: http://localhost:8080/realms/cloudbridge/protocol/openid-connect/auth

## Детальный процесс

### Что происходит при сборке:
1. **Webpack компилирует** Vue компоненты в обычный JavaScript
2. **SCSS компилируется** в CSS
3. **Шаблоны копируются** в папку темы Keycloak
4. **Файлы размещаются** в `keycloak/themes/playa/`

### Что происходит при применении темы:
1. **Keycloak перезапускается** (если указан флаг `--restart`)
2. **Тема применяется** к realm `cloudbridge`
3. **Настройки сохраняются** в Keycloak

## Опции скрипта сборки

```bash
# Простая сборка
./scripts/setup/build_keycloak_theme.sh

# Сборка с автоматическим применением
./scripts/setup/build_keycloak_theme.sh --restart
./scripts/setup/build_keycloak_theme.sh -r

# Справка
./scripts/setup/build_keycloak_theme.sh --help
```

## Ручное применение темы

Если нужно применить тему без перезапуска Keycloak:
```bash
./scripts/setup/check_keycloak_theme.sh
```

## Структура файлов

### Исходные файлы (keycloak-vue/src/):
```
keycloak-vue/src/
├── views/           # Vue компоненты страниц
├── components/      # Vue компоненты
├── scss/           # Стили
├── resources/      # Ресурсы (шрифты, изображения)
└── static/         # Статические файлы
```

### Собранные файлы (keycloak/themes/playa/):
```
keycloak/themes/playa/
├── theme.properties    # Конфигурация темы
├── login/
│   ├── login.ftl      # Шаблон входа
│   ├── register.ftl   # Шаблон регистрации
│   └── resources/
│       ├── js/        # Скомпилированный JavaScript
│       ├── css/       # Скомпилированные стили
│       ├── img/       # Изображения
│       └── fonts/     # Шрифты
```

## Устранение неполадок

### Тема не отображается:
1. Проверьте, что сборка прошла успешно
2. Убедитесь, что Keycloak перезапустился
3. Проверьте, что тема применена к realm

### Ошибки сборки:
1. Проверьте, что Node.js и npm установлены
2. Убедитесь, что зависимости установлены (`npm install`)
3. Проверьте синтаксис Vue компонентов

### Проверка статуса:
```bash
# Проверить файлы темы
find keycloak/themes/playa -type f

# Проверить логи Keycloak
docker-compose logs keycloak

# Проверить настройки темы
./scripts/setup/check_keycloak_theme.sh
```

## Автоматизация

### Для разработки:
Используйте флаг `--restart` для автоматического применения изменений:
```bash
./scripts/setup/build_keycloak_theme.sh --restart
```

### Для production:
Сначала протестируйте без флага, затем примените:
```bash
./scripts/setup/build_keycloak_theme.sh
# Проверьте результат
./scripts/setup/build_keycloak_theme.sh --restart
```

## Заключение

Теперь у вас есть полностью автоматизированный процесс разработки Vue темы для Keycloak:

1. **Редактируйте** файлы в `keycloak-vue/src/`
2. **Собирайте** тему одной командой
3. **Применяйте** автоматически с флагом `--restart`
4. **Проверяйте** результат в браузере

Процесс максимально упрощен и готов к использованию! 🎉 