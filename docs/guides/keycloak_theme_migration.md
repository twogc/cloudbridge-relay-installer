# Миграция темы Keycloak из keycloak-vue в keycloak

## Обзор

Данный документ описывает процесс аккуратного переноса кастомной темы Keycloak из папки `keycloak-vue` в папку `keycloak` без нарушения текущих настроек `docker-compose.yml`.

## Структура переноса

### Исходная структура (keycloak-vue)
```
keycloak-vue/
├── src/
│   ├── static/
│   │   ├── messages/
│   │   ├── template.ftl
│   │   └── theme.properties
│   ├── resources/
│   │   ├── css/
│   │   ├── fonts/
│   │   └── img/
│   ├── scss/
│   │   ├── _adaptive.scss
│   │   ├── _colors.scss
│   │   ├── _mixins.scss
│   │   ├── _typography.scss
│   │   └── index.scss
│   └── views/
│       ├── login/
│       ├── register/
│       └── loginResetPassword/
└── themes/
    └── Playa/
```

### Целевая структура (keycloak)
```
keycloak/
├── data/
├── imports/
└── themes/
    └── playa/
        ├── theme.properties
        ├── login/
        │   ├── css/
        │   │   └── style.css
        │   ├── js/
        │   │   └── theme.js
        │   ├── fonts/
        │   ├── img/
        │   ├── messages/
        │   ├── index.ftl
        │   ├── template.ftl
        │   └── theme.properties
        ├── register/
        │   └── register.ftl
        ├── account/
        ├── admin/
        └── email/
```

## Выполненные изменения

### 1. Создание структуры темы
- Создана правильная структура папок для темы Keycloak
- Перенесены все необходимые файлы из `keycloak-vue/src` в `keycloak/themes/playa`

### 2. Компиляция стилей
- SCSS файлы преобразованы в CSS
- Создан основной файл `style.css` с переменными CSS
- Добавлены адаптивные стили и типографика

### 3. Обновление docker-compose.yml
Добавлено монтирование темы:
```yaml
volumes:
  - ./keycloak/themes:/opt/keycloak/themes:ro
```

### 4. Создание скрипта настройки
Создан скрипт `scripts/setup/setup_keycloak_theme.sh` для:
- Автоматической настройки темы в Keycloak
- Создания realm и клиента
- Применения темы к realm

## Файлы темы

### Основные файлы
- `theme.properties` - конфигурация темы
- `css/style.css` - скомпилированные стили
- `js/theme.js` - JavaScript функциональность
- `fonts/` - шрифты Roboto
- `img/` - изображения и иконки
- `messages/` - файлы локализации

### Шаблоны
- `login/index.ftl` - шаблон страницы входа
- `register/register.ftl` - шаблон регистрации
- `template.ftl` - базовый шаблон

## Использование

### 1. Запуск с новой темой
```bash
# Перезапуск контейнеров с новой темой
docker-compose down
docker-compose up -d

# Настройка темы в Keycloak
export KEYCLOAK_ADMIN=admin
export KEYCLOAK_ADMIN_PASSWORD=your-password
./scripts/setup/setup_keycloak_theme.sh
```

### 2. Проверка темы
1. Откройте Keycloak Admin Console: http://localhost:8080
2. Войдите как администратор
3. Перейдите в Realm Settings > Themes
4. Убедитесь, что тема "playa" доступна
5. Выберите "playa" для Login Theme

### 3. Тестирование
- Откройте страницу входа: http://localhost:8080/realms/cloudbridge/protocol/openid-connect/auth
- Проверьте, что применяется новая тема
- Протестируйте адаптивность на разных устройствах

## Особенности темы

### Цветовая схема
- Основной цвет: #004EA8 (синий)
- Фон: #F5F5F5 (светло-серый)
- Текст: #1E1A1B (темно-серый)

### Типографика
- Шрифт: Roboto (Regular, Medium, Bold)
- Адаптивные размеры для разных экранов

### Адаптивность
- Desktop: 1441px - 1920px
- Laptop: 1025px - 1440px
- Tablet: 510px - 1024px
- Mobile: 350px - 509px

## Безопасность

### Права доступа
- Тема монтируется в режиме "только чтение" (`:ro`)
- Файлы темы не содержат чувствительной информации

### Совместимость
- Тема совместима с Keycloak 21+
- Поддерживает все стандартные функции Keycloak
- Не нарушает существующие настройки безопасности

## Устранение неполадок

### Тема не отображается
1. Проверьте, что файлы темы скопированы правильно
2. Убедитесь, что контейнер Keycloak перезапущен
3. Проверьте логи Keycloak: `docker-compose logs keycloak`

### Ошибки в стилях
1. Проверьте консоль браузера на ошибки CSS
2. Убедитесь, что файл `style.css` создан корректно
3. Проверьте пути к шрифтам и изображениям

### Проблемы с локализацией
1. Проверьте файлы в папке `messages/`
2. Убедитесь, что кодировка файлов UTF-8
3. Проверьте синтаксис `.properties` файлов

## Резервное копирование

### Перед миграцией
```bash
# Создание резервной копии
cp -r keycloak keycloak.backup.$(date +%Y%m%d_%H%M%S)
```

### Восстановление
```bash
# Восстановление из резервной копии
rm -rf keycloak
cp -r keycloak.backup.YYYYMMDD_HHMMSS keycloak
```

## Заключение

Миграция темы выполнена успешно с сохранением всех существующих настроек. Новая тема "playa" готова к использованию и полностью интегрирована в систему CloudBridge Relay.

### Контакты
При возникновении вопросов обращайтесь к команде разработки CloudBridge Relay. 