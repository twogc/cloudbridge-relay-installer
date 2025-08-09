# 🚀 CloudBridge Relay - Enterprise Edition

**Enterprise-grade высокопроизводительный relay сервер с интеграцией лучших практик Cloudflare**

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Docker](https://img.shields.io/badge/Docker-Ready-blue.svg)](https://hub.docker.com/r/twogc/cloudbridge-relay)
[![WireGuard](https://img.shields.io/badge/WireGuard-Integrated-green.svg)](https://www.wireguard.com/)
[![QUIC](https://img.shields.io/badge/QUIC-Enhanced-orange.svg)](https://quicwg.org/)

## 🌟 Возможности

### 🔄 **Multiprotocol Integration**
- **Enhanced QUIC** - улучшенный протокол QUIC с оптимизациями
- **HTTP/2** - высокопроизводительный HTTP протокол
- **WebSocket** - real-time соединения
- **Автоматическое переключение** между протоколами на основе производительности

### 🔐 **WireGuard P2P Mesh Network**
- **Реальная интеграция** с `wireguard-go`
- **P2P Mesh Network** с автоматическим обнаружением пиров
- **Гибридная маршрутизация** (прямая + mesh)
- **Site-to-Site** соединения
- **Автоматическое переключение** между реальным и симулированным WireGuard

### 🛡️ **Post-Quantum Cryptography**
- **CRYSTALS-Kyber** для квантово-устойчивого обмена ключами
- **CRYSTALS-Dilithium** для квантово-устойчивых цифровых подписей
- **Гибридные схемы** для обратной совместимости

### 📊 **Мониторинг и Метрики**
- **Prometheus** метрики
- **Grafana** дашборды
- **Расширенная телеметрия**
- **Health checks** и диагностика

### 🔒 **Безопасность**
- **Rate limiting** и DDoS защита
- **Tenant isolation**
- **Input validation**
- **Безопасная конфигурация**

## 🚀 Варианты распространения

### 1. **Go Module (Рекомендуемый)**
```bash
# Установка как Go модуль
go get github.com/twogc/cloudbridge-relay-installer

# Использование в коде
import "github.com/twogc/cloudbridge-relay-installer/internal/relay"
```

### 2. **Docker Package**
```bash
# Pull образа
docker pull twogc/cloudbridge-relay:latest

# Запуск
docker run -d \
  --name cloudbridge-relay \
  --privileged \
  -p 8080:8080 \
  -p 9090:9090 \
  -p 51820:51820/udp \
  twogc/cloudbridge-relay:latest
```

### 3. **Автоматическая установка**
```bash
# Универсальный скрипт установки
curl -sSL https://raw.githubusercontent.com/twogc/cloudbridge-relay-installer/main/install.sh | sudo bash

# Или через Docker
curl -sSL https://raw.githubusercontent.com/twogc/cloudbridge-relay-installer/main/install.sh | sudo bash -s -- --docker
```

### 4. **Docker Compose (Полный стек)**
```bash
# Скачать docker-compose.yml
wget https://raw.githubusercontent.com/twogc/cloudbridge-relay-installer/main/docker-compose.yml

# Запустить полный стек
docker-compose up -d
```

### 5. **NPM Package (для Node.js интеграции)**
```bash
# Установка NPM пакета
npm install @twogc/cloudbridge-relay

# Использование в Node.js
const { CloudBridgeRelay } = require('@twogc/cloudbridge-relay');
```

## 🛠️ Инструменты управления

### **Makefile - полный набор команд**
```bash
# Показать справку
make help

# Сборка
make build

# Тестирование
make test

# Установка
make install

# Docker
make docker-build
make docker-push

# Релиз
make release
```

### **GitHub Actions - автоматические релизы**
- Автоматическая сборка при создании тегов
- Кросс-компиляция для всех платформ
- Docker образы для amd64 и arm64
- Автоматические релизы на GitHub

## 📦 Компоненты

### **Enhanced QUIC**
- Оптимизированный QUIC протокол
- Улучшенная производительность
- Совместимость с стандартным QUIC
- Автоматическое переключение

### **WireGuard Integration**
- Реальная интеграция с `wireguard-go`
- P2P Mesh Network
- Автоматическое обнаружение пиров
- Гибридная маршрутизация

### **Post-Quantum Cryptography**
- Квантово-устойчивые алгоритмы
- Гибридные схемы
- Обратная совместимость
- Оптимизированная производительность

### **Monitoring Stack**
- Prometheus метрики
- Grafana дашборды
- Health checks
- Диагностические эндпоинты

## 🔧 Быстрый старт

### **Локальная установка**
```bash
# 1. Клонировать репозиторий
git clone https://github.com/twogc/cloudbridge-relay-installer.git
cd cloudbridge-relay-installer

# 2. Установить зависимости
make deps

# 3. Собрать проект
make build

# 4. Запустить
make monitor
```

### **Docker установка**
```bash
# 1. Скачать docker-compose.yml
wget https://raw.githubusercontent.com/twogc/cloudbridge-relay-installer/main/docker-compose.yml

# 2. Запустить
docker-compose up -d

# 3. Проверить статус
docker-compose ps
```

### **Автоматическая установка**
```bash
# Полная установка с WireGuard
curl -sSL https://raw.githubusercontent.com/twogc/cloudbridge-relay-installer/main/install.sh | sudo bash

# Только конфигурация
curl -sSL https://raw.githubusercontent.com/twogc/cloudbridge-relay-installer/main/install.sh | sudo bash -s -- --config-only
```

## 📊 Мониторинг

### **Доступные эндпоинты**
- **Relay сервер**: http://localhost:8080
- **Prometheus метрики**: http://localhost:9090/metrics
- **Grafana дашборды**: http://localhost:3000 (admin/admin)
- **WireGuard UI**: http://localhost:5000 (admin/admin)
- **Health check**: http://localhost:8080/health

### **Основные метрики**
- Количество активных соединений
- Пропускная способность
- Latency между узлами
- Количество WireGuard пиров
- Переключения протоколов
- Ошибки и их типы

## 🔒 Безопасность

### **Рекомендации по безопасности**
1. **Измените пароли по умолчанию**
2. **Настройте SSL/TLS сертификаты**
3. **Ограничьте доступ к API**
4. **Настройте firewall правила**
5. **Регулярно обновляйте зависимости**

### **Конфигурация безопасности**
```yaml
security:
  rate_limiting:
    enabled: true
    requests_per_minute: 1000
  authentication:
    enabled: true
    jwt_secret: "your-secret-key"
  cors:
    enabled: true
    allowed_origins: ["https://yourdomain.com"]
```

## 🧪 Тестирование

### **Запуск тестов**
```bash
# Все тесты
make test

# Тесты WireGuard
make test-wireguard

# Тесты реального WireGuard (требует sudo)
make test-real-wireguard

# Тесты с покрытием
make test-coverage
```

### **Интеграционные тесты**
```bash
# Запуск интеграционных тестов
go test -v -tags=integration ./tests/

# Performance тесты
go test -v -tags=performance ./tests/
```

## 📚 Документация

### **Техническая документация**
- [Архитектура системы](docs/architecture.md)
- [WireGuard интеграция](docs/wireguard_integration_guide.md)
- [Enhanced QUIC](docs/enhanced_quic.md)
- [Post-Quantum криптография](docs/post_quantum.md)
- [API Reference](docs/api.md)

### **Руководства**
- [Быстрый старт](docs/quickstart.md)
- [Конфигурация](docs/configuration.md)
- [Мониторинг](docs/monitoring.md)
- [Troubleshooting](docs/troubleshooting.md)

## 🤝 Вклад в проект

### **Как внести вклад**
1. Fork репозиторий
2. Создайте feature branch (`git checkout -b feature/amazing-feature`)
3. Commit изменения (`git commit -m 'Add amazing feature'`)
4. Push в branch (`git push origin feature/amazing-feature`)
5. Откройте Pull Request

### **Стандарты кода**
- Следуйте [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Используйте `gofmt` для форматирования
- Добавляйте тесты для нового функционала
- Обновляйте документацию

## 📄 Лицензия

Этот проект лицензирован под MIT License - см. [LICENSE](LICENSE) файл для деталей.

## 🆘 Поддержка

### **Получение помощи**
- 📖 [Документация](docs/)
- 🐛 [Issues](https://github.com/twogc/cloudbridge-relay-installer/issues)
- 💬 [Discussions](https://github.com/twogc/cloudbridge-relay-installer/discussions)
- 📧 Email: dev@2gc.ru

### **Коммерческая поддержка**
Для коммерческой поддержки и консультаций обращайтесь:
- 🌐 [2GC](https://2gc.ru)
- 📧 enterprise@2gc.ru

## 🏆 Статус проекта

![GitHub release (latest by date)](https://img.shields.io/github/v/release/twogc/cloudbridge-relay-installer)
![GitHub last commit](https://img.shields.io/github/last-commit/twogc/cloudbridge-relay-installer)
![GitHub issues](https://img.shields.io/github/issues/twogc/cloudbridge-relay-installer)
![GitHub pull requests](https://img.shields.io/github/issues-pr/twogc/cloudbridge-relay-installer)

---

**Сделано с ❤️ командой 2GC**