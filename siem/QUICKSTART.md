# 🚀 Быстрый старт CloudBridge SIEM

## Что у нас есть

Мы создали полнофункциональную SIEM систему со следующими компонентами:

### 📦 Основные компоненты
- **Elasticsearch** - хранение и поиск данных
- **Kibana** - визуализация и анализ
- **Grafana** - мониторинг метрик
- **Prometheus** - сбор метрик
- **Kafka** - потоковая обработка
- **Redis** - кэширование

### 🔧 SIEM компоненты
- **Collector** - сбор данных из CloudBridge Relay, Keycloak, системных логов
- **Processor** - обработка событий, детекция угроз, обогащение данных
- **Alert Manager** - система уведомлений (Telegram, Email, Slack)
- **Web UI** - веб-интерфейс для управления

## 🎯 Быстрый запуск

### 1. Подготовка
```bash
# Перейти в директорию SIEM
cd siem

# Установить зависимости
pip install -r requirements.txt

# Проверить Docker
docker --version
docker-compose --version
```

### 2. Запуск системы
```bash
# Автоматический запуск всех сервисов
./start_siem.sh

# Или через Makefile
make setup
```

### 3. Проверка работы
```bash
# Запуск тестов
python test_siem.py

# Или через Makefile
make test
```

## 🌐 Доступные интерфейсы

После запуска будут доступны:

| Сервис | URL | Описание |
|--------|-----|----------|
| **SIEM Web UI** | http://localhost:8081 | Главный веб-интерфейс |
| **Kibana** | http://localhost:5601 | Анализ и визуализация данных |
| **Grafana** | http://localhost:3001 | Мониторинг метрик |
| **Prometheus** | http://localhost:9091 | Метрики системы |
| **Elasticsearch** | http://localhost:9200 | API для данных |

## 🛠️ Управление системой

### Основные команды
```bash
# Запуск
make start

# Остановка
make stop

# Перезапуск
make restart

# Статус
make status

# Логи
make logs

# Тесты
make test

# Резервная копия
make backup

# Проверка здоровья
make health
```

### Через Python скрипт
```bash
# Запуск
python manage_siem.py start

# Остановка
python manage_siem.py stop

# Статус
python manage_siem.py status

# Информация
python manage_siem.py info
```

## 📊 Что отслеживается

### События безопасности
- **Brute Force атаки** - множественные неудачные попытки входа
- **Подозрительные IP** - IP из черных списков или подозрительных стран
- **Необычная активность пользователей** - аномальные паттерны поведения
- **Аномалии реле** - необычная активность туннелей
- **Экспорт данных** - подозрительная передача больших объемов данных

### Источники данных
- **CloudBridge Relay** - логи туннелей, аутентификации, метрики
- **Keycloak** - события аутентификации и авторизации
- **Системные логи** - системные события, ошибки
- **Сетевые метрики** - трафик, соединения, производительность

## 🔔 Уведомления

Система поддерживает уведомления через:
- **Telegram** - мгновенные уведомления
- **Email** - детальные отчеты
- **Slack** - интеграция с командными чатами
- **Webhook** - интеграция с внешними системами

## 🧪 Тестирование

### Генерация тестовых событий
```bash
# Запуск тестового скрипта
python test_siem.py

# Создание тестовых данных
curl -X POST http://localhost:9200/siem-events/_doc \
  -H "Content-Type: application/json" \
  -d '{
    "timestamp": "2024-01-01T12:00:00Z",
    "source": "test",
    "event_type": "test_event",
    "severity": "medium",
    "message": "Тестовое событие"
  }'
```

### Проверка детекции
1. Откройте Kibana: http://localhost:5601
2. Перейдите в Discover
3. Выберите индекс `siem-events`
4. Найдите ваши тестовые события

## 🔧 Настройка

### Конфигурация
Основной файл конфигурации: `config/siem_config.yaml`

```yaml
# Настройки сбора данных
collection:
  interval: 60  # секунды
  sources:
    - relay
    - keycloak
    - system

# Правила детекции
detection:
  brute_force:
    enabled: true
    threshold: 5
    window: 300

# Уведомления
alerting:
  telegram:
    enabled: true
    bot_token: "YOUR_BOT_TOKEN"
    chat_id: "YOUR_CHAT_ID"
```

### Добавление новых источников
1. Создайте новый коллектор в `collectors/`
2. Добавьте правила обработки в `processors/`
3. Обновите конфигурацию
4. Перезапустите систему

## 📈 Мониторинг

### Дашборды
- **SIEM Overview** - общий обзор событий
- **Security Analytics** - анализ угроз
- **Relay Monitoring** - мониторинг туннелей
- **User Activity** - активность пользователей

### Метрики
- Количество событий в секунду
- Время обработки событий
- Количество уведомлений
- Статус компонентов

## 🚨 Устранение неполадок

### Частые проблемы

#### Сервисы не запускаются
```bash
# Проверка логов
make logs

# Проверка статуса
make status

# Перезапуск
make restart
```

#### Elasticsearch не отвечает
```bash
# Проверка здоровья
curl http://localhost:9200/_cluster/health

# Проверка логов
make logs-elasticsearch
```

#### Нет данных в Kibana
```bash
# Проверка индексов
curl http://localhost:9200/_cat/indices

# Проверка коллектора
make logs-collector
```

### Логи
```bash
# Все логи
make logs

# Конкретный сервис
make logs-elasticsearch
make logs-collector
make logs-processor
```

## 🔄 Обновление

```bash
# Обновление системы
make update

# Или вручную
git pull
make install
make build
make restart
```

## 💾 Резервное копирование

```bash
# Создание резервной копии
make backup

# Восстановление
make restore BACKUP_DIR=backups/siem_backup_20240101_120000
```

## 🎉 Готово!

Ваша SIEM система готова к работе! 

### Следующие шаги:
1. **Настройте уведомления** - добавьте Telegram бота или email
2. **Настройте правила детекции** - адаптируйте под ваши нужды
3. **Интегрируйте с CloudBridge Relay** - настройте сбор данных
4. **Создайте дашборды** - настройте визуализацию в Kibana
5. **Настройте мониторинг** - добавьте метрики в Grafana

### Полезные ссылки:
- [Полная документация](README.md)
- [Руководство по настройке](docs/)
- [API документация](docs/api_reference.md)
- [Примеры интеграций](docs/integrations.md)

---

**Нужна помощь?** Создайте issue в репозитории или обратитесь к документации. 