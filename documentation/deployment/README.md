# Deployment Documentation

## 📁 Файлы в этой папке

- **[DEVOPS_CHECKLIST.md](./DEVOPS_CHECKLIST.md)** - Чек-лист для DevOps процессов
- **[DEVOPS_ROADMAP.md](./DEVOPS_ROADMAP.md)** - Дорожная карта DevOps развития
- **[DOCKER_COMPOSE_GUIDE.md](./DOCKER_COMPOSE_GUIDE.md)** - Руководство по Docker Compose
- **[DOCKER_CONTAINERIZATION.md](./DOCKER_CONTAINERIZATION.md)** - Контейнеризация с Docker
- **[MINIKUBE_STAGING_GUIDE.md](./MINIKUBE_STAGING_GUIDE.md)** - Полное руководство по Minikube staging
- **[QUICK_START_STAGING.md](./QUICK_START_STAGING.md)** - Быстрый старт staging environment

## 🚀 Процесс развертывания

### 1. Подготовка
- Проверка зависимостей
- Настройка окружения
- Конфигурация секретов

### 2. Сборка
- Docker образы
- Kubernetes manifests
- Terraform конфигурации

### 3. Развертывание
- Локальное развертывание
- Kubernetes deployment
- Production deployment

### 4. Тестирование
- Unit тесты
- Integration тесты
- Production тесты

## 🔧 Команды развертывания

```bash
# Локальное развертывание
./run-local.sh

# Kubernetes deployment
./deploy-k8s.sh

# Terraform deployment
./terraform/scripts/deploy.sh apply prod

# Production testing
./test-production.sh
```



