# 🚀 CloudBridge Relay - CI/CD Pipeline

## 📋 Обзор

Этот документ описывает настройку и использование CI/CD pipeline для CloudBridge Relay Server с интеграцией Cadence Workflow Engine.

## 🔧 Компоненты Pipeline

### 1. **GitHub Actions Workflows**

#### Основные файлы:
- `.github/workflows/ci.yml` - Базовая CI/CD pipeline
- `.github/workflows/ci-enhanced.yml` - Улучшенная CI/CD pipeline
- `.github/workflows/build-and-deploy.yml` - Сборка и развертывание
- `.github/workflows/docker-build-push.yml` - Docker сборка

#### Функциональность:
- ✅ Автоматические тесты при push/PR
- ✅ Сборка Docker образов
- ✅ Security scanning (Trivy, gosec)
- ✅ Code coverage reporting
- ✅ Multi-environment deployment
- ✅ Уведомления о статусе

### 2. **Docker Configuration**

#### Dockerfile'ы:
- `Dockerfile.production` - Production-ready образ
- `Dockerfile.relay.alpine` - Alpine-based образ
- `Dockerfile.relay.simple` - Упрощенный образ

#### Docker Compose:
- `docker-compose.yml` - Базовая конфигурация
- `docker-compose.production.yml` - Production конфигурация

### 3. **Build Scripts**

#### Основные скрипты:
- `build.sh` - Основной скрипт сборки
- `setup-github-secrets.sh` - Настройка GitHub Secrets

## 🚀 Быстрый старт

### 1. **Настройка GitHub Secrets**

```bash
# Запустите интерактивный скрипт
./setup-github-secrets.sh
```

#### Обязательные секреты:
- `GITHUB_TOKEN` - GitHub token для registry access

#### Опциональные секреты:
- `REGISTRY_USERNAME` / `REGISTRY_PASSWORD` - Container registry
- `VK_REGISTRY_USERNAME` / `VK_REGISTRY_PASSWORD` - VK Cloud registry
- `SNYK_TOKEN` - Security scanning
- `CODECOV_TOKEN` - Code coverage
- `SLACK_WEBHOOK` - Уведомления

### 2. **Локальная сборка**

```bash
# Сборка всех компонентов
./build.sh

# Сборка Docker образа
docker build -f Dockerfile.production -t cloudbridge-relay:latest .

# Запуск в production режиме
docker-compose -f docker-compose.production.yml up -d
```

### 3. **Запуск CI/CD Pipeline**

```bash
# Push в main/develop ветку
git push origin main

# Создание release
git tag v1.0.0
git push origin v1.0.0

# Ручной запуск
gh workflow run ci-enhanced.yml
```

## 📊 Pipeline Stages

### 1. **Test Stage**
- Go tests с race detection
- Code coverage reporting
- Linting (golangci-lint)
- Security scanning (gosec)

### 2. **Build Stage**
- Сборка бинарных файлов
- Создание distribution archive
- Docker image build
- Multi-architecture support

### 3. **Security Stage**
- Vulnerability scanning (Trivy)
- Dependency checking
- SAST analysis

### 4. **Deploy Stage**
- Staging deployment (develop branch)
- Production deployment (tags)
- Environment-specific configs

## 🔍 Мониторинг

### GitHub Actions
- **Status**: https://github.com/[repo]/actions
- **Secrets**: https://github.com/[repo]/settings/secrets/actions
- **Packages**: https://github.com/[repo]/packages

### Container Registry
- **Images**: `ghcr.io/[repo]/cloudbridge-relay`
- **Tags**: `latest`, `dev`, `v1.0.0`

### Metrics
- **Coverage**: Codecov integration
- **Security**: Trivy reports
- **Performance**: Build time metrics

## 🛠️ Конфигурация

### Environment Variables

```bash
# Build configuration
VERSION=1.0.0
BUILD_TIME=2025-09-06T10:00:00Z
GIT_COMMIT=abc123

# Runtime configuration
LOG_LEVEL=info
WIREGUARD_ENABLED=true
GIN_MODE=release
```

### Docker Compose Override

```yaml
# docker-compose.override.yml
version: '3.8'
services:
  cloudbridge-relay:
    environment:
      - LOG_LEVEL=debug
    volumes:
      - ./custom-config.json:/app/configs/config.json:ro
```

## 🔒 Безопасность

### Security Scanning
- **Trivy**: Container vulnerability scanning
- **gosec**: Go security analysis
- **Snyk**: Dependency vulnerability checking

### Secrets Management
- GitHub Secrets для CI/CD
- Kubernetes Secrets для deployment
- HashiCorp Vault (опционально)

### Network Security
- UFW firewall configuration
- Port restrictions
- TLS/SSL certificates

## 📈 Производительность

### Build Optimization
- Multi-stage Docker builds
- Build caching
- Parallel job execution
- Artifact caching

### Runtime Optimization
- Resource limits
- Health checks
- Auto-scaling
- Load balancing

## 🚨 Troubleshooting

### Common Issues

#### 1. **Build Failures**
```bash
# Check Go version
go version

# Clean build cache
go clean -cache

# Rebuild dependencies
go mod download
go mod verify
```

#### 2. **Docker Build Issues**
```bash
# Check Docker daemon
docker info

# Clean Docker cache
docker system prune -a

# Rebuild without cache
docker build --no-cache -f Dockerfile.production .
```

#### 3. **GitHub Actions Issues**
```bash
# Check workflow syntax
gh workflow list

# View workflow runs
gh run list

# Download logs
gh run download [run-id]
```

### Debug Commands

```bash
# Local testing
./build.sh
docker-compose -f docker-compose.production.yml up -d

# Check services
docker-compose ps
docker-compose logs cloudbridge-relay

# Health checks
curl http://localhost:9090/health
curl http://localhost:8088
curl http://localhost:3000
```

## 📚 Дополнительные ресурсы

### Documentation
- [DevOps Roadmap](documentation/DEVOPS_ROADMAP.md)
- [DevOps Checklist](documentation/DEVOPS_CHECKLIST.md)
- [API Reference](docs/api-reference.md)

### Tools
- [GitHub Actions](https://docs.github.com/en/actions)
- [Docker](https://docs.docker.com/)
- [Kubernetes](https://kubernetes.io/docs/)
- [Prometheus](https://prometheus.io/docs/)

### Support
- GitHub Issues: [Create Issue](https://github.com/[repo]/issues/new)
- Documentation: [Wiki](https://github.com/[repo]/wiki)
- Discussions: [GitHub Discussions](https://github.com/[repo]/discussions)

---

*Документ создан: 2025-09-06*  
*Статус: ✅ Готов к использованию* 🚀

