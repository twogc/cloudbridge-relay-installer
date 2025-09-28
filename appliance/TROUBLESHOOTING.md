# CloudBridge Appliance Troubleshooting

## 🔧 Исправление ошибок

### Ошибки компиляции Go

**Проблема**: `could not import github.com/gorilla/mux`

**Решение**:
```bash
# Установить зависимости
cd appliance/api
go mod tidy
go mod download

# Или использовать скрипт установки
./scripts/setup-dev.sh
```

### Ошибки типов

**Проблема**: `cannot use am.getHealthStatus() (value of type string) as *HealthStatus value`

**Решение**: Исправлено в коде - добавлен метод `getHealthStatusStruct()`

### Ошибки Prometheus метрик

**Проблема**: `am.metrics.ActiveConnections.Get undefined`

**Решение**: Заменено на прямые вызовы методов получения метрик

## 🚀 Быстрая настройка

### 1. Установка зависимостей

```bash
# Перейти в папку appliance
cd appliance

# Запустить скрипт установки
./scripts/setup-dev.sh
```

### 2. Сборка компонентов

```bash
# Использовать Makefile
make deps
make build
make test
```

### 3. Docker сборка

```bash
# Собрать Docker образы
make docker

# Или напрямую
cd docker
./build-appliance.sh
```

## 📋 Проверка работоспособности

### API компоненты

```bash
# Проверить компиляцию
cd api
go build ./...

# Запустить тесты
go test -v ./...
```

### Docker образы

```bash
# Проверить Docker образ
docker build -f docker/Dockerfile.appliance -t cloudbridge-appliance .

# Запустить контейнер
docker run -d --name cloudbridge-appliance \
  -p 8444:8444 -p 8080:8080 \
  cloudbridge-appliance
```

### Скрипты управления

```bash
# Проверить права доступа
ls -la scripts/

# Установить права
chmod +x scripts/*.sh
```

## 🐛 Частые проблемы

### 1. Отсутствие Go

**Ошибка**: `command not found: go`

**Решение**:
```bash
# Установить Go 1.21+
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

### 2. Отсутствие Docker

**Ошибка**: `command not found: docker`

**Решение**:
```bash
# Установить Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER
```

### 3. Проблемы с правами

**Ошибка**: `Permission denied`

**Решение**:
```bash
# Установить права на скрипты
chmod +x scripts/*.sh

# Запустить от root (для установки)
sudo ./scripts/install-appliance.sh
```

### 4. Проблемы с портами

**Ошибка**: `bind: address already in use`

**Решение**:
```bash
# Найти процесс, использующий порт
sudo netstat -tlnp | grep :8080

# Остановить процесс
sudo kill -9 <PID>
```

## 📊 Диагностика

### Проверка статуса

```bash
# Проверить статус сервиса
./scripts/manage-appliance.sh status

# Проверить здоровье
./scripts/manage-appliance.sh health

# Проверить метрики
./scripts/manage-appliance.sh metrics
```

### Логи

```bash
# Просмотр логов
./scripts/manage-appliance.sh logs

# Следить за логами
./scripts/manage-appliance.sh logs -f
```

### Сеть

```bash
# Проверить порты
netstat -tlnp | grep -E "(8080|8444|19302|3478|5553)"

# Проверить подключение
curl -f http://localhost:8080/health
```

## 🔄 Восстановление

### Сброс к исходному состоянию

```bash
# Остановить сервисы
./scripts/manage-appliance.sh stop

# Очистить данные
sudo rm -rf /var/lib/cloudbridge/*
sudo rm -rf /var/log/cloudbridge/*

# Перезапустить
./scripts/manage-appliance.sh start
```

### Восстановление из резервной копии

```bash
# Создать резервную копию
./scripts/manage-appliance.sh backup

# Восстановить из резервной копии
./scripts/manage-appliance.sh restore /var/backups/cloudbridge/backup-file.tar.gz
```

## 📞 Поддержка

### Техническая поддержка

- **Email**: support@cloudbridge.ru
- **Phone**: +7 (495) 123-45-67
- **Telegram**: @cloudbridge_support

### Документация

- **API Reference**: `docs/README.md`
- **Pricing Guide**: `docs/pricing-guide.md`
- **Pilot Program**: `pilot/pilot-program.md`

### Сообщество

- **GitHub**: github.com/cloudbridge/relay
- **Discord**: discord.gg/cloudbridge
- **Telegram**: @cloudbridge_community
