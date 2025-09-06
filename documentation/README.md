# CloudBridge Relay - Documentation

Добро пожаловать в документацию CloudBridge Relay! Этот проект представляет собой современную систему релея с поддержкой QUIC, WireGuard, AI-сервисов и продвинутого мониторинга.

## 📚 Структура документации

### 🏗️ [Architecture](./architecture/)
- [Project Overview](./architecture/PROJECT_OVERVIEW.md) - Обзор проекта и архитектуры
- [Architecture Diagrams](./architecture/architecture-diagrams.md) - Диаграммы архитектуры системы

### 🚀 [Deployment](./deployment/)
- [DevOps Checklist](./deployment/DEVOPS_CHECKLIST.md) - Чек-лист для DevOps
- [DevOps Roadmap](./deployment/DEVOPS_ROADMAP.md) - Дорожная карта DevOps
- [Docker Compose Guide](./deployment/DOCKER_COMPOSE_GUIDE.md) - Руководство по Docker Compose
- [Docker Containerization](./deployment/DOCKER_CONTAINERIZATION.md) - Контейнеризация с Docker

### ⚙️ [Configuration](./configuration/)
- [Configuration Examples](./configuration/configuration-examples.md) - Примеры конфигурации
- [Docker Best Practices](./configuration/DOCKER_BEST_PRACTICES.md) - Лучшие практики Docker

### 📊 [Monitoring](./monitoring/)
- [Monitoring Guide](./monitoring/MONITORING_GUIDE.md) - Руководство по мониторингу

### 🔒 [Security](./security/)
- Security policies и compliance
- Network policies
- Pod security policies
- Security scanning tools

### 🔌 [API](./api/)
- [API Reference](./api/api-reference.md) - Справочник по API

### 📖 [Guides](./guides/)
- [Complete README](./guides/README_COMPLETE.md) - Полное руководство
- [Documentation Structure](./guides/DOCUMENTATION_STRUCTURE.md) - Структура документации

### 📋 [Reports](./reports/)
- Отчеты по тестированию
- Отчеты по безопасности
- Отчеты по производительности

## 🎯 Основные компоненты

### Core Services
- **CloudBridge Relay** - Основной сервис релея
- **QUIC Service** - QUIC протокол для быстрой передачи данных
- **WireGuard Service** - VPN соединения
- **AI Service** - Искусственный интеллект для оптимизации
- **Cadence Service** - Workflow orchestration

### Infrastructure
- **Kubernetes** - Оркестрация контейнеров
- **Docker** - Контейнеризация
- **Terraform** - Infrastructure as Code
- **Prometheus** - Мониторинг метрик
- **Grafana** - Визуализация данных

### Security
- **Network Policies** - Сетевая безопасность
- **Pod Security Policies** - Безопасность подов
- **RBAC** - Контроль доступа
- **Security Scanning** - Автоматическое сканирование

## 🚀 Быстрый старт

1. **Установка зависимостей**:
   ```bash
   ./scripts/setup/check_dependencies.sh
   ```

2. **Запуск локально**:
   ```bash
   ./run-local.sh
   ```

3. **Развертывание в Kubernetes**:
   ```bash
   ./deploy-k8s.sh
   ```

4. **Мониторинг**:
   ```bash
   ./deploy-monitoring.sh
   ```

## 📊 Статус проекта

- ✅ **CI/CD Pipeline** - GitHub Actions настроен
- ✅ **Docker Containerization** - Multi-stage builds
- ✅ **Kubernetes Deployment** - Production-ready manifests
- ✅ **Monitoring** - Prometheus/Grafana stack
- ✅ **Security** - Comprehensive security policies
- ✅ **Infrastructure as Code** - Terraform modules
- 🔄 **Backup & Recovery** - В разработке
- 🔄 **Performance Optimization** - В разработке

## 🔧 Полезные команды

### Разработка
```bash
# Запуск тестов
make test

# Сборка Docker образов
./build-docker.sh

# Локальный запуск
./run-local.sh
```

### Развертывание
```bash
# Kubernetes deployment
./deploy-k8s.sh

# Terraform deployment
./terraform/scripts/deploy.sh apply prod

# Production testing
./test-production.sh
```

### Мониторинг
```bash
# Security scanning
./security/scanning/security-scan.sh

# Performance analysis
./performance-analysis.sh
```

## 📞 Поддержка

- **Issues**: [GitHub Issues](https://github.com/your-org/cloudbridge-relay/issues)
- **Documentation**: [Wiki](https://github.com/your-org/cloudbridge-relay/wiki)
- **Discussions**: [GitHub Discussions](https://github.com/your-org/cloudbridge-relay/discussions)

## 📄 Лицензия

Этот проект лицензирован под MIT License - см. файл [LICENSE](../LICENSE) для деталей.

---

**CloudBridge Relay** - Современная система релея для высокопроизводительных соединений