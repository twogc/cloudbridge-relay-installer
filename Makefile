# CloudBridge Relay v2.0 Makefile
# Enterprise-grade высокопроизводительный relay сервер

# Переменные
BINARY_NAME=cloudbridge-relay
VERSION=2.0.0
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD)
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}"

# Цвета для вывода
GREEN=\033[0;32m
YELLOW=\033[1;33m
RED=\033[0;31m
BLUE=\033[0;34m
NC=\033[0m # No Color

# Цели по умолчанию
.PHONY: help build test clean install dev watch docker-build docker-run deploy-staging deploy-production

# Помощь
help: ## Показать справку по командам
	@echo "$(BLUE)CloudBridge Relay v2.0 - Makefile$(NC)"
	@echo ""
	@echo "$(GREEN)Доступные команды:$(NC)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(YELLOW)%-20s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Сборка
build: ## Собрать бинарный файл
	@echo "$(BLUE)Сборка CloudBridge Relay v${VERSION}...$(NC)"
	@go build ${LDFLAGS} -o bin/${BINARY_NAME} ./cmd/relay
	@echo "$(GREEN)✓ Сборка завершена: bin/${BINARY_NAME}$(NC)"

build-linux: ## Собрать для Linux
	@echo "$(BLUE)Сборка для Linux...$(NC)"
	@GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o bin/${BINARY_NAME}-linux-amd64 ./cmd/relay
	@echo "$(GREEN)✓ Сборка для Linux завершена$(NC)"

build-windows: ## Собрать для Windows
	@echo "$(BLUE)Сборка для Windows...$(NC)"
	@GOOS=windows GOARCH=amd64 go build ${LDFLAGS} -o bin/${BINARY_NAME}-windows-amd64.exe ./cmd/relay
	@echo "$(GREEN)✓ Сборка для Windows завершена$(NC)"

build-all: build build-linux build-windows ## Собрать для всех платформ
	@echo "$(GREEN)✓ Сборка для всех платформ завершена$(NC)"

# Тестирование
test: ## Запустить unit тесты
	@echo "$(BLUE)Запуск unit тестов...$(NC)"
	@go test -v ./...
	@echo "$(GREEN)✓ Unit тесты завершены$(NC)"

test-coverage: ## Запустить тесты с покрытием
	@echo "$(BLUE)Запуск тестов с покрытием...$(NC)"
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ Отчет о покрытии создан: coverage.html$(NC)"

test-integration: ## Запустить integration тесты
	@echo "$(BLUE)Запуск integration тестов...$(NC)"
	@go test -v -tags=integration ./tests/
	@echo "$(GREEN)✓ Integration тесты завершены$(NC)"

test-performance: ## Запустить performance тесты
	@echo "$(BLUE)Запуск performance тестов...$(NC)"
	@go test -v -tags=performance ./tests/
	@echo "$(GREEN)✓ Performance тесты завершены$(NC)"

test-all: test test-coverage test-integration test-performance ## Запустить все тесты

# Линтинг и форматирование
lint: ## Запустить линтер
	@echo "$(BLUE)Проверка кода линтером...$(NC)"
	@golangci-lint run
	@echo "$(GREEN)✓ Линтинг завершен$(NC)"

fmt: ## Форматировать код
	@echo "$(BLUE)Форматирование кода...$(NC)"
	@go fmt ./...
	@echo "$(GREEN)✓ Форматирование завершено$(NC)"

vet: ## Проверить код с go vet
	@echo "$(BLUE)Проверка кода с go vet...$(NC)"
	@go vet ./...
	@echo "$(GREEN)✓ Проверка завершена$(NC)"

# Разработка
dev: ## Запустить в режиме разработки
	@echo "$(BLUE)Запуск в режиме разработки...$(NC)"
	@go run ${LDFLAGS} ./cmd/relay -config config/config.json

watch: ## Запустить с hot reload
	@echo "$(BLUE)Запуск с hot reload...$(NC)"
	@air -c .air.toml

install: ## Установить зависимости
	@echo "$(BLUE)Установка зависимостей...$(NC)"
	@go mod download
	@go mod tidy
	@echo "$(GREEN)✓ Зависимости установлены$(NC)"

# Docker
docker-build: ## Собрать Docker образ
	@echo "$(BLUE)Сборка Docker образа...$(NC)"
	@docker build -t ${BINARY_NAME}:${VERSION} .
	@docker tag ${BINARY_NAME}:${VERSION} ${BINARY_NAME}:latest
	@echo "$(GREEN)✓ Docker образ создан: ${BINARY_NAME}:${VERSION}$(NC)"

docker-run: ## Запустить Docker контейнер
	@echo "$(BLUE)Запуск Docker контейнера...$(NC)"
	@docker run -d \
		--name ${BINARY_NAME} \
		-p 8080:8080 \
		-p 8081:8081 \
		-v $(PWD)/config:/app/config \
		${BINARY_NAME}:${VERSION}
	@echo "$(GREEN)✓ Контейнер запущен$(NC)"

docker-stop: ## Остановить Docker контейнер
	@echo "$(BLUE)Остановка Docker контейнера...$(NC)"
	@docker stop ${BINARY_NAME} || true
	@docker rm ${BINARY_NAME} || true
	@echo "$(GREEN)✓ Контейнер остановлен$(NC)"

# Развертывание
deploy-staging: ## Развернуть в staging
	@echo "$(BLUE)Развертывание в staging...$(NC)"
	@kubectl apply -f k8s/staging/
	@echo "$(GREEN)✓ Развертывание в staging завершено$(NC)"

deploy-production: ## Развернуть в production
	@echo "$(BLUE)Развертывание в production...$(NC)"
	@kubectl apply -f k8s/production/
	@echo "$(GREEN)✓ Развертывание в production завершено$(NC)"

# Мониторинг
monitor-staging: ## Мониторинг staging
	@echo "$(BLUE)Мониторинг staging...$(NC)"
	@kubectl get pods -n staging
	@kubectl logs -f deployment/${BINARY_NAME} -n staging

monitor-production: ## Мониторинг production
	@echo "$(BLUE)Мониторинг production...$(NC)"
	@kubectl get pods -n production
	@kubectl logs -f deployment/${BINARY_NAME} -n production

# Очистка
clean: ## Очистить артефакты сборки
	@echo "$(BLUE)Очистка артефактов...$(NC)"
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@go clean
	@echo "$(GREEN)✓ Очистка завершена$(NC)"

# Проверка качества
quality: lint vet test-coverage ## Проверить качество кода
	@echo "$(GREEN)✓ Проверка качества завершена$(NC)"

# Бенчмарки
bench: ## Запустить бенчмарки
	@echo "$(BLUE)Запуск бенчмарков...$(NC)"
	@go test -bench=. -benchmem ./...
	@echo "$(GREEN)✓ Бенчмарки завершены$(NC)"

# Генерация документации
docs: ## Генерировать документацию
	@echo "$(BLUE)Генерация документации...$(NC)"
	@swag init -g cmd/relay/main.go
	@echo "$(GREEN)✓ Документация сгенерирована$(NC)"

# Проверка безопасности
security: ## Проверить безопасность
	@echo "$(BLUE)Проверка безопасности...$(NC)"
	@gosec ./...
	@echo "$(GREEN)✓ Проверка безопасности завершена$(NC)"

# Создание релиза
release: build-all test-all quality ## Создать релиз
	@echo "$(BLUE)Создание релиза v${VERSION}...$(NC)"
	@mkdir -p releases
	@cp bin/${BINARY_NAME}* releases/
	@cp config/config.json releases/
	@cp README.md releases/
	@cp CHANGELOG.md releases/
	@tar -czf releases/${BINARY_NAME}-${VERSION}.tar.gz -C releases .
	@echo "$(GREEN)✓ Релиз создан: releases/${BINARY_NAME}-${VERSION}.tar.gz$(NC)"

# Проверка версии
version: ## Показать версию
	@echo "$(BLUE)CloudBridge Relay v${VERSION}$(NC)"
	@echo "Build Time: ${BUILD_TIME}"
	@echo "Git Commit: ${GIT_COMMIT}"

# Проверка зависимостей
deps-check: ## Проверить зависимости
	@echo "$(BLUE)Проверка зависимостей...$(NC)"
	@go list -u -m all
	@echo "$(GREEN)✓ Проверка зависимостей завершена$(NC)"

# Обновление зависимостей
deps-update: ## Обновить зависимости
	@echo "$(BLUE)Обновление зависимостей...$(NC)"
	@go get -u ./...
	@go mod tidy
	@echo "$(GREEN)✓ Зависимости обновлены$(NC)"

# Проверка готовности к production
production-ready: build test-all quality security ## Проверить готовность к production
	@echo "$(GREEN)✓ Проект готов к production!$(NC)"

# Информация о системе
info: ## Показать информацию о системе
	@echo "$(BLUE)Информация о системе:$(NC)"
	@echo "Go Version: $(shell go version)"
	@echo "OS: $(shell uname -s)"
	@echo "Architecture: $(shell uname -m)"
	@echo "Project Version: ${VERSION}"
	@echo "Build Time: ${BUILD_TIME}"
	@echo "Git Commit: ${GIT_COMMIT}"

# Создание .gitignore
gitignore: ## Создать .gitignore
	@echo "$(BLUE)Создание .gitignore...$(NC)"
	@echo "# Binaries for programs and plugins" > .gitignore
	@echo "*.exe" >> .gitignore
	@echo "*.exe~" >> .gitignore
	@echo "*.dll" >> .gitignore
	@echo "*.so" >> .gitignore
	@echo "*.dylib" >> .gitignore
	@echo "" >> .gitignore
	@echo "# Test binary, built with 'go test -c'" >> .gitignore
	@echo "*.test" >> .gitignore
	@echo "" >> .gitignore
	@echo "# Output of the go coverage tool" >> .gitignore
	@echo "*.out" >> .gitignore
	@echo "coverage.html" >> .gitignore
	@echo "" >> .gitignore
	@echo "# Build artifacts" >> .gitignore
	@echo "bin/" >> .gitignore
	@echo "releases/" >> .gitignore
	@echo "" >> .gitignore
	@echo "# IDE files" >> .gitignore
	@echo ".vscode/" >> .gitignore
	@echo ".idea/" >> .gitignore
	@echo "" >> .gitignore
	@echo "# OS files" >> .gitignore
	@echo ".DS_Store" >> .gitignore
	@echo "Thumbs.db" >> .gitignore
	@echo "$(GREEN)✓ .gitignore создан$(NC)"

# Инициализация проекта
init: gitignore install ## Инициализировать проект
	@echo "$(GREEN)✓ Проект инициализирован$(NC)"

# Полная проверка перед коммитом
pre-commit: fmt lint vet test quality ## Полная проверка перед коммитом
	@echo "$(GREEN)✓ Код готов к коммиту!$(NC)" 