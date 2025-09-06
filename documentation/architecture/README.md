# Architecture Documentation

## 📁 Файлы в этой папке

- **[PROJECT_OVERVIEW.md](./PROJECT_OVERVIEW.md)** - Обзор проекта и архитектуры системы
- **[architecture-diagrams.md](./architecture-diagrams.md)** - Диаграммы архитектуры и компонентов

## 🏗️ Архитектура CloudBridge Relay

### Основные компоненты
- **CloudBridge Relay** - Основной сервис релея
- **QUIC Service** - QUIC протокол для быстрой передачи данных
- **WireGuard Service** - VPN соединения
- **AI Service** - Искусственный интеллект для оптимизации
- **Cadence Service** - Workflow orchestration

### Инфраструктура
- **Kubernetes** - Оркестрация контейнеров
- **Docker** - Контейнеризация
- **Terraform** - Infrastructure as Code
- **Prometheus/Grafana** - Мониторинг

### Безопасность
- **Network Policies** - Сетевая изоляция
- **Pod Security Policies** - Безопасность контейнеров
- **RBAC** - Контроль доступа



