# 💻 Разработка

Эта папка содержит документацию для разработчиков CloudBridge Relay.

## 📋 Содержание

### 🚀 Быстрый старт
- **README_DEVELOPER_QUICKSTART_VERIFICATION.md** - Проверка быстрого старта
- **DEVELOPER_TASK_MICROSERVICES_IMPLEMENTATION.md** - Реализация микросервисов

### 🔧 Backend разработка
- **Backend Django.md** - Django backend разработка
- **Логика взаимодействия между Relay-сервером и клиентом.md** - Логика взаимодействия

## 🛠️ Инструменты разработки

### 🐳 Docker
```bash
# Сборка образов
make docker-build

# Локальный запуск
docker-compose -f docker-compose.test.yml up
```

### 🔧 Go разработка
```bash
# Запуск тестов
go test ./...

# Сборка
go build ./cmd/auth-service
go build ./cmd/metrics-service
```

### 📊 Мониторинг разработки
```bash
# Локальные метрики
curl http://localhost:9090/metrics

# Health checks
curl http://localhost:8080/health
```

## 🎯 Микросервисы

### Auth Service
- **Порт**: 8080 (HTTP), 9090 (метрики)
- **Эндпоинты**: `/auth/*`, `/health`, `/ready`
- **База данных**: PostgreSQL
- **Кэш**: Redis

### Metrics Service
- **Порт**: 8080 (HTTP), 9090 (метрики)
- **Эндпоинты**: `/metrics/*`, `/health`, `/ready`
- **Хранилище**: InfluxDB/TimescaleDB
- **Агрегация**: Prometheus

## 🔗 Полезные ссылки

- **API документация**: `../api/`
- **Архитектура**: `../architecture/`
- **Установка**: `../installation/`
