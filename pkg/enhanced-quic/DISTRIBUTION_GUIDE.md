# 📦 Руководство по распространению Enhanced QUIC

## 🎯 Обзор

Это руководство описывает, как распространять наш модифицированный Enhanced QUIC протокол через различные платформы и методы установки.

## 🚀 Варианты распространения

### 1. **Go Module (Рекомендуемый)**

#### Создание отдельного репозитория:
```bash
# Создать новый репозиторий
git clone https://github.com/twogc/enhanced-quic.git
cd enhanced-quic

# Скопировать код из основного проекта
cp -r ../cloudbridge-relay-installer/internal/quic/enhanced.go ./internal/
cp -r ../cloudbridge-relay-installer/internal/quic/enhanced/ ./internal/
cp -r ../cloudbridge-relay-installer/internal/connection/ ./internal/
cp -r ../cloudbridge-relay-installer/internal/circuitbreaker/ ./internal/
cp -r ../cloudbridge-relay-installer/internal/health/ ./internal/
cp -r ../cloudbridge-relay-installer/internal/metrics/ ./internal/
cp -r ../cloudbridge-relay-installer/internal/logger/ ./internal/

# Обновить go.mod
go mod tidy
go mod vendor
```

#### Публикация:
```bash
# Создать тег
git tag v1.0.0
git push origin v1.0.0

# Пользователи смогут установить:
go get github.com/twogc/enhanced-quic@v1.0.0
```

### 2. **NPM Package**

#### Создание TypeScript обертки:
```bash
# Инициализировать NPM пакет
npm init -y
npm install typescript @types/node

# Создать TypeScript обертку для Go binary
# См. package.json и src/ директорию
```

#### Публикация:
```bash
npm run build
npm publish --access public
```

#### Использование:
```bash
npm install @twogc/enhanced-quic
```

### 3. **Docker Image**

#### Сборка и публикация:
```bash
# Сборка образа
docker build -t twogc/enhanced-quic:latest .

# Публикация в Docker Hub
docker push twogc/enhanced-quic:latest

# Пользователи смогут запустить:
docker run -d -p 8080:8080 twogc/enhanced-quic:latest
```

### 4. **Binary Releases**

#### Автоматическая сборка через GitHub Actions:
```yaml
# См. .github/workflows/release.yml
# Автоматически создает релизы для:
# - Linux (amd64, arm64)
# - macOS (amd64, arm64)  
# - Windows (amd64, arm64)
```

#### Установка через скрипт:
```bash
# Пользователи смогут установить:
curl -sSL https://raw.githubusercontent.com/twogc/enhanced-quic/main/install.sh | bash

# Или с Docker:
curl -sSL https://raw.githubusercontent.com/twogc/enhanced-quic/main/install.sh | bash -s -- --docker
```

## 📋 Чек-лист для релиза

### ✅ Подготовка к релизу:
- [ ] Обновить версию в `go.mod`
- [ ] Обновить версию в `package.json`
- [ ] Обновить версию в `Dockerfile`
- [ ] Проверить все тесты: `make test`
- [ ] Проверить линтер: `make lint`
- [ ] Обновить документацию
- [ ] Создать changelog

### ✅ Создание релиза:
- [ ] Создать тег: `git tag v1.0.0`
- [ ] Запушить тег: `git push origin v1.0.0`
- [ ] GitHub Actions автоматически создаст релиз
- [ ] Проверить все артефакты в релизе
- [ ] Проверить Docker образ
- [ ] Проверить NPM пакет

### ✅ Пост-релиз:
- [ ] Обновить документацию на сайте
- [ ] Отправить анонс в сообщество
- [ ] Мониторить обратную связь
- [ ] Готовить патчи при необходимости

## 🔧 Инструменты для управления

### Makefile команды:
```bash
# Сборка
make build              # Собрать для текущей платформы
make build-all          # Собрать для всех платформ
make build-linux        # Собрать для Linux
make build-darwin       # Собрать для macOS
make build-windows      # Собрать для Windows

# Тестирование
make test               # Запустить тесты
make test-coverage      # Тесты с покрытием
make lint               # Запустить линтер
make benchmark          # Запустить бенчмарки

# Docker
make docker-build       # Собрать Docker образ
make docker-push        # Опубликовать Docker образ

# NPM
make npm-install        # Установить NPM зависимости
make npm-build          # Собрать NPM пакет
make npm-publish        # Опубликовать NPM пакет

# Релиз
make release            # Подготовить релиз
make release-tag        # Создать тег релиза
```

### Скрипт установки:
```bash
# Установка последней версии
./install.sh

# Установка конкретной версии
./install.sh --version v1.0.0

# Установка через Docker
./install.sh --docker
```

## 🌐 Платформы распространения

### 1. **GitHub**
- Основной репозиторий
- Releases с бинарными файлами
- GitHub Packages для Docker образов

### 2. **Docker Hub**
- Официальные Docker образы
- Автоматические сборки
- Multi-arch поддержка

### 3. **NPM Registry**
- TypeScript/JavaScript обертки
- Node.js интеграция
- Автоматические обновления

### 4. **Homebrew (macOS)**
```bash
# Создать формулу для Homebrew
# См. docs/homebrew-formula.rb
```

### 5. **Snap Store (Linux)**
```bash
# Создать snap пакет
# См. snapcraft.yaml
```

## 📊 Мониторинг и аналитика

### GitHub Analytics:
- Downloads per release
- Star count
- Fork count
- Issue tracking

### Docker Hub Analytics:
- Pull count
- Platform usage
- Version distribution

### NPM Analytics:
- Download count
- Version usage
- Dependency graph

## 🔒 Безопасность

### Подписывание релизов:
```bash
# GPG подпись
gpg --detach-sign enhanced-quic-linux-amd64
gpg --verify enhanced-quic-linux-amd64.sig

# Проверка checksums
sha256sum enhanced-quic-* > checksums.txt
```

### Сканирование безопасности:
```bash
# Сканирование Docker образа
docker scan twogc/enhanced-quic:latest

# Сканирование Go кода
gosec ./...
```

## 📈 Метрики успеха

### Количественные:
- Количество установок
- Количество звезд на GitHub
- Количество загрузок Docker образа
- Количество загрузок NPM пакета

### Качественные:
- Отзывы пользователей
- Количество issues/PR
- Активность сообщества
- Использование в production

## 🆘 Поддержка

### Документация:
- README.md с примерами
- API документация
- Руководства по развертыванию
- Troubleshooting guide

### Сообщество:
- GitHub Discussions
- Discord сервер
- Stack Overflow теги
- Reddit сообщества

### Поддержка:
- Email: support@twogc.com
- GitHub Issues
- Discord канал поддержки
- Документация: docs.twogc.com 