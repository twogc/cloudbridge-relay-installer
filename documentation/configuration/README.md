# Configuration Documentation

## 📁 Файлы в этой папке

- **[configuration-examples.md](./configuration-examples.md)** - Примеры конфигурации
- **[DOCKER_BEST_PRACTICES.md](./DOCKER_BEST_PRACTICES.md)** - Лучшие практики Docker
- **[DOMAIN_DATABASE_SETUP.md](./DOMAIN_DATABASE_SETUP.md)** - Настройка домена Cadence для баз данных

## ⚙️ Конфигурация системы

### Основные конфигурационные файлы
- `config/config.yaml` - Основная конфигурация
- `configs/config.json` - JSON конфигурация
- `docker-compose.yml` - Docker Compose конфигурация
- `k8s/` - Kubernetes manifests

### Переменные окружения
- `env.example` - Пример переменных окружения
- `.env` - Локальные переменные окружения

### Секреты
- Kubernetes secrets
- Docker secrets
- Terraform variables

## 🔧 Настройка

### 1. Базовая конфигурация
```bash
# Копирование примера конфигурации
cp config_template.yaml config/config.yaml

# Настройка переменных окружения
cp env.example .env
```

### 2. Docker конфигурация
```bash
# Запуск с Docker Compose
docker-compose up -d

# Сборка образов
./build-docker.sh
```

### 3. Kubernetes конфигурация
```bash
# Применение manifests
kubectl apply -f k8s/

# Проверка статуса
kubectl get pods -A
```



