# Makefile для CloudBridge Relay WireGuard
# Управление сборкой, тестированием, релизом и развертыванием

# Переменные
BINARY_NAME=cloudbridge-relay
VERSION=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD)
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}"

# Платформы для кросс-компиляции
PLATFORMS=linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

# Docker
DOCKER_IMAGE=twogc/cloudbridge-relay
DOCKER_TAG=${VERSION}

# Цвета для вывода
GREEN=\033[0;32m
YELLOW=\033[1;33m
RED=\033[0;31m
NC=\033[0m # No Color

.PHONY: help build test clean install docker-build docker-push release

# Основная цель
help: ## Показать справку
	@echo "$(GREEN)CloudBridge Relay WireGuard - Инструменты управления$(NC)"
	@echo ""
	@echo "$(YELLOW)Доступные команды:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2}'

# Сборка
build: ## Собрать бинарный файл
	@echo "$(GREEN)Сборка CloudBridge Relay...$(NC)"
	go build ${LDFLAGS} -o bin/${BINARY_NAME} ./cmd/relay
	@echo "$(GREEN)Бинарный файл создан: bin/${BINARY_NAME}$(NC)"

build-all: ## Собрать для всех платформ
	@echo "$(GREEN)Кросс-компиляция для всех платформ...$(NC)"
	@mkdir -p bin
	@for platform in ${PLATFORMS}; do \
		IFS='/' read -r GOOS GOARCH <<< "$$platform"; \
		BINARY_NAME_EXT=$${BINARY_NAME}; \
		if [ "$$GOOS" = "windows" ]; then \
			BINARY_NAME_EXT=$${BINARY_NAME}.exe; \
		fi; \
		echo "Сборка для $$GOOS/$$GOARCH..."; \
		GOOS=$$GOOS GOARCH=$$GOARCH go build ${LDFLAGS} -o bin/$${BINARY_NAME}-$$GOOS-$$GOARCH$${BINARY_NAME_EXT} ./cmd/relay; \
	done
	@echo "$(GREEN)Сборка завершена!$(NC)"

# Тестирование
test: ## Запустить все тесты
	@echo "$(GREEN)Запуск тестов...$(NC)"
	go test -v ./...
	@echo "$(GREEN)Тесты завершены!$(NC)"

test-wireguard: ## Запустить тесты WireGuard
	@echo "$(GREEN)Запуск тестов WireGuard...$(NC)"
	go test -v -run TestWireGuard ./cmd/relay
	@echo "$(GREEN)Тесты WireGuard завершены!$(NC)"

test-real-wireguard: ## Запустить тесты реального WireGuard
	@echo "$(GREEN)Запуск тестов реального WireGuard...$(NC)"
	sudo go test -v -run TestRealWireGuard ./cmd/relay
	@echo "$(GREEN)Тесты реального WireGuard завершены!$(NC)"

test-coverage: ## Запустить тесты с покрытием
	@echo "$(GREEN)Запуск тестов с покрытием...$(NC)"
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Отчет о покрытии создан: coverage.html$(NC)"

# Установка
install: ## Установить в систему
	@echo "$(GREEN)Установка CloudBridge Relay...$(NC)"
	go install ${LDFLAGS} ./cmd/relay
	@echo "$(GREEN)CloudBridge Relay установлен!$(NC)"

install-wireguard: ## Установить WireGuard зависимости
	@echo "$(GREEN)Установка WireGuard зависимостей...$(NC)"
	sudo ./scripts/install_wireguard_go.sh
	@echo "$(GREEN)WireGuard зависимости установлены!$(NC)"

# Docker
docker-build: ## Собрать Docker образ
	@echo "$(GREEN)Сборка Docker образа...$(NC)"
	docker build -t ${DOCKER_IMAGE}:${DOCKER_TAG} .
	docker tag ${DOCKER_IMAGE}:${DOCKER_TAG} ${DOCKER_IMAGE}:latest
	@echo "$(GREEN)Docker образ создан: ${DOCKER_IMAGE}:${DOCKER_TAG}$(NC)"

docker-push: ## Отправить Docker образ в registry
	@echo "$(GREEN)Отправка Docker образа...$(NC)"
	docker push ${DOCKER_IMAGE}:${DOCKER_TAG}
	docker push ${DOCKER_IMAGE}:latest
	@echo "$(GREEN)Docker образ отправлен!$(NC)"

docker-run: ## Запустить в Docker
	@echo "$(GREEN)Запуск в Docker...$(NC)"
	docker run -d --name cloudbridge-relay \
		-p 8080:8080 \
		-p 9090:9090 \
		-v $(PWD)/configs:/app/configs \
		${DOCKER_IMAGE}:latest

# Релиз
release: ## Создать релиз
	@echo "$(GREEN)Создание релиза...$(NC)"
	@if [ -z "$(VERSION)" ]; then \
		echo "$(RED)Ошибка: VERSION не установлен$(NC)"; \
		exit 1; \
	fi
	@echo "Создание релиза версии $(VERSION)..."
	@mkdir -p releases/$(VERSION)
	@cp bin/* releases/$(VERSION)/
	@cp configs/*.yaml releases/$(VERSION)/
	@cp scripts/install.sh releases/$(VERSION)/
	@cp README.md releases/$(VERSION)/
	@cd releases && tar -czf $(VERSION).tar.gz $(VERSION)
	@echo "$(GREEN)Релиз создан: releases/$(VERSION).tar.gz$(NC)"

release-github: ## Создать GitHub релиз
	@echo "$(GREEN)Создание GitHub релиза...$(NC)"
	@if [ -z "$(VERSION)" ]; then \
		echo "$(RED)Ошибка: VERSION не установлен$(NC)"; \
		exit 1; \
	fi
	@gh release create $(VERSION) \
		--title "CloudBridge Relay $(VERSION)" \
		--notes "Релиз CloudBridge Relay с WireGuard интеграцией" \
		releases/$(VERSION).tar.gz

# Очистка
clean: ## Очистить артефакты сборки
	@echo "$(GREEN)Очистка...$(NC)"
	rm -rf bin/
	rm -rf releases/
	rm -f coverage.out coverage.html
	go clean
	@echo "$(GREEN)Очистка завершена!$(NC)"

# Проверки
lint: ## Запустить линтер
	@echo "$(GREEN)Запуск линтера...$(NC)"
	golangci-lint run
	@echo "$(GREEN)Линтер завершен!$(NC)"

fmt: ## Форматировать код
	@echo "$(GREEN)Форматирование кода...$(NC)"
	go fmt ./...
	@echo "$(GREEN)Форматирование завершено!$(NC)"

vet: ## Проверить код
	@echo "$(GREEN)Проверка кода...$(NC)"
	go vet ./...
	@echo "$(GREEN)Проверка завершена!$(NC)"

# Документация
docs: ## Сгенерировать документацию
	@echo "$(GREEN)Генерация документации...$(NC)"
	@mkdir -p docs/generated
	godoc -http=:6060 &
	@echo "$(GREEN)Документация доступна на http://localhost:6060$(NC)"

# Мониторинг
monitor: ## Запустить мониторинг
	@echo "$(GREEN)Запуск мониторинга...$(NC)"
	@if [ -f bin/${BINARY_NAME} ]; then \
		./bin/${BINARY_NAME} --config configs/config.yaml; \
	else \
		echo "$(RED)Бинарный файл не найден. Запустите 'make build'$(NC)"; \
	fi

# Развертывание
deploy-local: ## Развернуть локально
	@echo "$(GREEN)Локальное развертывание...$(NC)"
	@make install-wireguard
	@make build
	@make monitor

deploy-docker: ## Развернуть через Docker
	@echo "$(GREEN)Развертывание через Docker...$(NC)"
	@make docker-build
	@make docker-run

# Утилиты
version: ## Показать версию
	@echo "$(GREEN)Версия: $(VERSION)$(NC)"
	@echo "$(GREEN)Время сборки: $(BUILD_TIME)$(NC)"
	@echo "$(GREEN)Git commit: $(GIT_COMMIT)$(NC)"

deps: ## Установить зависимости
	@echo "$(GREEN)Установка зависимостей...$(NC)"
	go mod download
	go mod tidy
	@echo "$(GREEN)Зависимости установлены!$(NC)"

# Проверка окружения
check-env: ## Проверить окружение
	@echo "$(GREEN)Проверка окружения...$(NC)"
	@command -v go >/dev/null 2>&1 || { echo "$(RED)Go не установлен$(NC)"; exit 1; }
	@command -v git >/dev/null 2>&1 || { echo "$(RED)Git не установлен$(NC)"; exit 1; }
	@command -v docker >/dev/null 2>&1 || { echo "$(YELLOW)Docker не установлен (некоторые команды недоступны)$(NC)"; }
	@echo "$(GREEN)Окружение готово!$(NC)"

# Полная сборка
all: check-env deps fmt vet lint test build ## Полная сборка проекта
	@echo "$(GREEN)Полная сборка завершена!$(NC)" 