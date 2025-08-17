# 🐳 Руководство по автоматизации Docker Registry в VK Cloud

## 📋 Обзор

Это руководство описывает полную автоматизацию сборки и отправки Docker образов в приватный репозиторий VK Cloud. Система включает:

- **Автоматическую установку** Docker Registry в VK Cloud
- **CI/CD pipeline** через GitHub Actions
- **Локальные скрипты** для ручной сборки и развертывания
- **Интеграцию с Kubernetes** для автоматического развертывания

## 🏗️ Архитектура

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   GitHub Repo   │───▶│  GitHub Actions  │───▶│  VK Cloud K8s   │
└─────────────────┘    └──────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│  Local Scripts  │    │  Docker Registry │    │   S3 Storage    │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

## 🚀 Быстрый старт

### 1. Установка Docker Registry

```bash
# Клонируйте репозиторий
git clone <your-repo>
cd cloudbridge-relay-installer

# Установите Docker Registry
chmod +x scripts/install-vkcloud-registry.sh
./scripts/install-vkcloud-registry.sh
```

### 2. Настройка GitHub Secrets

```bash
# Установите GitHub CLI
sudo apt install gh

# Аутентифицируйтесь
gh auth login

# Настройте секреты
chmod +x scripts/setup-github-secrets.sh
./scripts/setup-github-secrets.sh
```

### 3. Запуск автоматизации

```bash
# Ручная сборка и отправка
chmod +x scripts/docker-registry-automation.sh
./scripts/docker-registry-automation.sh

# Или через GitHub Actions (автоматически при push)
git push origin main
```

## 📁 Структура файлов

```
cloudbridge-relay-installer/
├── scripts/
│   ├── docker-registry-automation.sh      # Основной скрипт автоматизации
│   ├── install-vkcloud-registry.sh        # Установка registry
│   ├── setup-github-secrets.sh            # Настройка GitHub Secrets
│   └── setup-env.sh                       # Переменные окружения
├── configs/
│   └── docker-registry-vkcloud.yaml       # Конфигурация registry
├── .github/workflows/
│   └── docker-build-push.yml              # GitHub Actions workflow
├── k8s/
│   ├── relay-deployment.yaml              # Kubernetes deployments
│   ├── ai-deployment.yaml
│   ├── cadence-deployment.yaml
│   ├── quic-deployment.yaml
│   └── wireguard-deployment.yaml
└── docs/automation/
    └── DOCKER_REGISTRY_AUTOMATION_GUIDE.md # Это руководство
```

## 🔧 Конфигурация

### Переменные окружения

Создайте файл `.env` или экспортируйте переменные:

```bash
# Registry настройки
export REGISTRY_IP="95.163.250.128"
export REGISTRY_PORT="5000"
export REGISTRY_USERNAME="cloudbridge"
export REGISTRY_PASSWORD="password"
export REGISTRY_NAMESPACE="cloudbridge"

# Версии
export VERSION="$(git describe --tags --always --dirty)"
export BUILD_TIME="$(date -u '+%Y-%m-%d_%H:%M:%S')"
export GIT_COMMIT="$(git rev-parse HEAD)"
```

### GitHub Secrets

Необходимые секреты для GitHub Actions:

- `REGISTRY_IP` - IP адрес Docker Registry
- `REGISTRY_USERNAME` - Имя пользователя для registry
- `REGISTRY_PASSWORD` - Пароль для registry
- `KUBECONFIG` - Base64-encoded kubeconfig файл

## 🛠️ Использование скриптов

### Основной скрипт автоматизации

```bash
# Полный цикл (сборка + отправка + развертывание)
./scripts/docker-registry-automation.sh

# Только сборка
./scripts/docker-registry-automation.sh build

# Только отправка
./scripts/docker-registry-automation.sh push

# Только развертывание
./scripts/docker-registry-automation.sh deploy

# Проверка образов
./scripts/docker-registry-automation.sh verify

# Очистка локальных образов
./scripts/docker-registry-automation.sh cleanup

# Справка
./scripts/docker-registry-automation.sh help
```

### Установка Registry

```bash
# Установка registry
./scripts/install-vkcloud-registry.sh

# Проверка установки
./scripts/install-vkcloud-registry.sh verify

# Удаление registry
./scripts/install-vkcloud-registry.sh uninstall
```

### Настройка GitHub Secrets

```bash
# Настройка секретов
./scripts/setup-github-secrets.sh

# Проверка секретов
./scripts/setup-github-secrets.sh verify

# Удаление секретов
./scripts/setup-github-secrets.sh delete

# Тест registry
./scripts/setup-github-secrets.sh test
```

## 🔄 GitHub Actions Workflow

### Триггеры

Workflow запускается автоматически при:

- Push в ветки `main` или `develop`
- Создание тегов `v*`
- Pull Request в ветку `main`
- Ручной запуск через GitHub UI

### Этапы выполнения

1. **build-and-push** - Сборка и отправка образов
2. **deploy-to-k8s** - Развертывание в Kubernetes (только для main/tags)

### Ручной запуск

```bash
# Через GitHub CLI
gh workflow run docker-build-push.yml

# С параметрами
gh workflow run docker-build-push.yml \
  --field service=relay \
  --field version=v1.0.0
```

## 📦 Поддерживаемые сервисы

Система автоматически обрабатывает следующие сервисы:

- **relay** - Основной relay сервис
- **ai** - AI/ML сервис
- **cadence** - Cadence workflow сервис
- **quic** - QUIC протокол сервис
- **wireguard** - WireGuard VPN сервис

### Dockerfile структура

Каждый сервис должен иметь соответствующий Dockerfile:

```
Dockerfile.relay      # Основной сервис
Dockerfile.ai         # AI сервис
Dockerfile.cadence    # Cadence сервис
Dockerfile.quic       # QUIC сервис
Dockerfile.wireguard  # WireGuard сервис
```

## 🔍 Мониторинг и отладка

### Проверка статуса registry

```bash
# Проверка доступности
curl -s http://95.163.250.128:5000/v2/

# Список образов
curl -u cloudbridge:password http://95.163.250.128:5000/v2/_catalog

# Список тегов для образа
curl -u cloudbridge:password http://95.163.250.128:5000/v2/cloudbridge/relay/tags/list
```

### Проверка Kubernetes

```bash
# Статус подов
kubectl --insecure-skip-tls-verify get pods -n cloudbridge

# Логи сервиса
kubectl --insecure-skip-tls-verify logs -n cloudbridge -l app=relay

# Описание пода
kubectl --insecure-skip-tls-verify describe pod -n cloudbridge <pod-name>
```

### Логи GitHub Actions

```bash
# Просмотр логов workflow
gh run list --workflow=docker-build-push.yml

# Детальные логи
gh run view <run-id> --log
```

## 🚨 Устранение проблем

### Ошибки сборки

1. **Dockerfile не найден**
   ```bash
   # Проверьте наличие Dockerfile
   ls -la Dockerfile.*
   
   # Создайте недостающий Dockerfile
   cp Dockerfile.relay Dockerfile.missing-service
   ```

2. **Ошибка сборки образа**
   ```bash
   # Проверьте Dockerfile синтаксис
   docker build --dry-run -f Dockerfile.relay .
   
   # Проверьте зависимости
   docker build --no-cache -f Dockerfile.relay .
   ```

### Ошибки отправки

1. **Ошибка аутентификации**
   ```bash
   # Проверьте учетные данные
   docker login 95.163.250.128:5000 --username cloudbridge --password password
   
   # Проверьте доступность registry
   curl -s http://95.163.250.128:5000/v2/
   ```

2. **Ошибка сети**
   ```bash
   # Проверьте сетевое подключение
   ping 95.163.250.128
   
   # Проверьте порт
   telnet 95.163.250.128 5000
   ```

### Ошибки Kubernetes

1. **ImagePullBackOff**
   ```bash
   # Проверьте секрет
   kubectl --insecure-skip-tls-verify get secret -n cloudbridge
   
   # Пересоздайте секрет
   kubectl --insecure-skip-tls-verify delete secret cloudbridge-registry-secret -n cloudbridge
   kubectl --insecure-skip-tls-verify create secret docker-registry cloudbridge-registry-secret \
     --docker-server=95.163.250.128:5000 \
     --docker-username=cloudbridge \
     --docker-password=password \
     --docker-email=info@2gc.ru \
     -n cloudbridge
   ```

2. **Pod не запускается**
   ```bash
   # Проверьте события
   kubectl --insecure-skip-tls-verify get events -n cloudbridge --sort-by='.lastTimestamp'
   
   # Проверьте описание пода
   kubectl --insecure-skip-tls-verify describe pod -n cloudbridge <pod-name>
   ```

## 🔒 Безопасность

### Рекомендации

1. **Измените пароли по умолчанию**
   ```bash
   # В configs/docker-registry-vkcloud.yaml
   auth:
     htpasswd:
       password: "your-secure-password"
   ```

2. **Используйте TLS**
   ```bash
   # Включите TLS в конфигурации
   tls:
     enabled: true
     secretName: registry-tls
   ```

3. **Ограничьте доступ к registry**
   ```bash
   # Настройте firewall правила
   # Ограничьте доступ только с определенных IP
   ```

### Аудит

```bash
# Проверьте логи registry
kubectl --insecure-skip-tls-verify logs -n kube-system -l app=registry

# Проверьте доступы
kubectl --insecure-skip-tls-verify get secret -n cloudbridge -o yaml
```

## 📈 Масштабирование

### Горизонтальное масштабирование

```bash
# Увеличьте количество реплик
kubectl --insecure-skip-tls-verify scale deployment relay-deployment --replicas=3 -n cloudbridge

# Проверьте статус
kubectl --insecure-skip-tls-verify get pods -n cloudbridge
```

### Вертикальное масштабирование

```bash
# Обновите ресурсы в deployment файлах
resources:
  requests:
    memory: 512Mi
    cpu: 250m
  limits:
    memory: 1Gi
    cpu: 500m
```

## 🔄 Обновления

### Обновление registry

```bash
# Обновите Helm chart
helm --insecure-skip-tls-verify upgrade cloudbridge-registry twuni/docker-registry \
  --namespace kube-system \
  --values configs/docker-registry-vkcloud.yaml
```

### Обновление скриптов

```bash
# Получите последние изменения
git pull origin main

# Обновите права доступа
chmod +x scripts/*.sh
```

## 📞 Поддержка

### Полезные команды

```bash
# Полная диагностика
./scripts/docker-registry-automation.sh verify
./scripts/install-vkcloud-registry.sh verify
./scripts/setup-github-secrets.sh verify

# Логи всех компонентов
kubectl --insecure-skip-tls-verify logs -n kube-system -l app=registry
kubectl --insecure-skip-tls-verify logs -n cloudbridge --all-containers
```

### Контакты

- **Автор**: VK Cloud API Manager
- **Email**: info@2gc.ru
- **Документация**: [VK Cloud Registry](https://mcs.mail.ru/app/docs/kubernetes/addons/docker-registry)

---

**⏱️ Время настройки**: 30-60 минут  
**🎯 Результат**: Полностью автоматизированная система сборки и развертывания Docker образов в VK Cloud
