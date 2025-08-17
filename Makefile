# CloudBridge Relay Installer Makefile

# Переменные
VERSION ?= $(shell git describe --tags --always --dirty)
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT ?= $(shell git rev-parse HEAD)
REGISTRY ?= twogc
IMAGE_NAME ?= cloudbridge-relay
FULL_IMAGE_NAME = $(REGISTRY)/$(IMAGE_NAME):$(VERSION)
LATEST_IMAGE_NAME = $(REGISTRY)/$(IMAGE_NAME):latest

# Go параметры
GOOS ?= linux
GOARCH ?= amd64
CGO_ENABLED = 1

# Директории
BIN_DIR = bin
DIST_DIR = dist
SCRIPTS_DIR = scripts

.PHONY: help build clean test docker-build docker-push docker-run docker-stop install-wireguard test-ports deploy-k8s test-k8s

help: ## Показать справку
	@echo "CloudBridge Relay Installer"
	@echo "=========================="
	@echo "Доступные команды:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Собрать бинарный файл
	@echo "Сборка CloudBridge Relay..."
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) go build \
		-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)" \
		-o $(BIN_DIR)/cloudbridge-relay ./cmd/relay
	@echo "✅ Сборка завершена: $(BIN_DIR)/cloudbridge-relay"

clean: ## Очистить артефакты сборки
	@echo "Очистка артефактов..."
	@rm -rf $(BIN_DIR) $(DIST_DIR)
	@echo "✅ Очистка завершена"

test: ## Запустить тесты
	@echo "Запуск тестов..."
	go test -v ./...
	@echo "✅ Тесты завершены"

docker-build: ## Собрать Docker образ
	@echo "Сборка Docker образа: $(FULL_IMAGE_NAME)"
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		-t $(FULL_IMAGE_NAME) \
		-t $(LATEST_IMAGE_NAME) .
	@echo "✅ Docker образ собран"

docker-push: ## Отправить Docker образ в registry
	@echo "Отправка Docker образа в registry..."
	docker push $(FULL_IMAGE_NAME)
	docker push $(LATEST_IMAGE_NAME)
	@echo "✅ Docker образ отправлен"

docker-run: ## Запустить Docker контейнер
	@echo "Запуск Docker контейнера..."
	docker-compose up -d
	@echo "✅ Docker контейнер запущен"

docker-stop: ## Остановить Docker контейнер
	@echo "Остановка Docker контейнера..."
	docker-compose down
	@echo "✅ Docker контейнер остановлен"

install-wireguard: ## Установить WireGuard зависимости
	@echo "Установка WireGuard зависимостей..."
	@chmod +x $(SCRIPTS_DIR)/install-wireguard.sh
	@$(SCRIPTS_DIR)/install-wireguard.sh
	@echo "✅ WireGuard зависимости установлены"

test-ports: ## Тестировать порты relay сервера
	@echo "Тестирование портов relay сервера..."
	@chmod +x ../terraform-relay/scripts/test-relay-ports.sh
	@../terraform-relay/scripts/test-relay-ports.sh localhost 80
	@echo "✅ Тестирование портов завершено"

test-ports-remote: ## Тестировать порты удаленного relay сервера
	@echo "Тестирование портов удаленного relay сервера..."
	@chmod +x ../terraform-relay/scripts/test-relay-ports.sh
	@../terraform-relay/scripts/test-relay-ports.sh 109.120.191.126 80
	@echo "✅ Тестирование удаленных портов завершено"

test-anycast: ## Тестировать Anycast IP
	@echo "Тестирование Anycast IP..."
	@chmod +x ../terraform-relay/scripts/test-relay-ports.sh
	@../terraform-relay/scripts/test-relay-ports.sh 83.166.235.45 80
	@echo "✅ Тестирование Anycast IP завершено"

deploy-k8s: ## Развернуть в Kubernetes
	@echo "Развертывание в Kubernetes..."
	kubectl apply -f ../terraform-relay/k8s/manifests/cloudbridge-manifests.yaml
	@echo "✅ Развертывание в Kubernetes завершено"

deploy-k8s-anycast: ## Развернуть в Kubernetes с Anycast
	@echo "Развертывание в Kubernetes с Anycast..."
	kubectl apply -f ../terraform-relay/k8s/manifests/cloudbridge-anycast-manifests.yaml
	@echo "✅ Развертывание в Kubernetes с Anycast завершено"

test-k8s: ## Тестировать Kubernetes развертывание
	@echo "Тестирование Kubernetes развертывания..."
	@echo "Проверка подов:"
	kubectl get pods -n cloudbridge-relay
	@echo "Проверка сервисов:"
	kubectl get services -n cloudbridge-relay
	@echo "Проверка логов:"
	kubectl logs -n cloudbridge-relay -l app=cloudbridge-relay --tail=20
	@echo "✅ Тестирование Kubernetes завершено"

build-and-deploy: docker-build docker-push deploy-k8s ## Полный цикл сборки и развертывания

build-and-deploy-anycast: docker-build docker-push deploy-k8s-anycast ## Полный цикл сборки и развертывания с Anycast

# Локальное тестирование
test-local: build ## Локальное тестирование
	@echo "Локальное тестирование..."
	@echo "Запуск relay сервера..."
	@$(BIN_DIR)/cloudbridge-relay --config configs/config.json &
	@sleep 5
	@echo "Тестирование портов..."
	@make test-ports
	@echo "Остановка relay сервера..."
	@pkill -f cloudbridge-relay || true
	@echo "✅ Локальное тестирование завершено"

# Проверка статуса
status: ## Проверить статус сервисов
	@echo "Статус Docker контейнеров:"
	docker-compose ps
	@echo ""
	@echo "Статус Kubernetes подов:"
	kubectl get pods -n cloudbridge-relay 2>/dev/null || echo "Kubernetes namespace не найден"
	@echo ""
	@echo "Статус портов:"
	@make test-ports

# Очистка всего
clean-all: clean docker-stop ## Полная очистка
	@echo "Полная очистка..."
	docker system prune -f
	kubectl delete namespace cloudbridge-relay 2>/dev/null || true
	@echo "✅ Полная очистка завершена"

# Документация
docs: ## Генерировать документацию
	@echo "Генерация документации..."
	@mkdir -p docs
	@echo "# CloudBridge Relay" > docs/README.md
	@echo "" >> docs/README.md
	@echo "## Версия: $(VERSION)" >> docs/README.md
	@echo "## Время сборки: $(BUILD_TIME)" >> docs/README.md
	@echo "## Git commit: $(GIT_COMMIT)" >> docs/README.md
	@echo "✅ Документация сгенерирована"

# Проверка безопасности
security-check: ## Проверка безопасности
	@echo "Проверка безопасности..."
	gosec ./...
	@echo "✅ Проверка безопасности завершена"

# Линтинг
lint: ## Запустить линтер
	@echo "Запуск линтера..."
	golangci-lint run
	@echo "✅ Линтинг завершен"

# Форматирование
fmt: ## Форматировать код
	@echo "Форматирование кода..."
	go fmt ./...
	@echo "✅ Форматирование завершено"

# Проверка зависимостей
deps: ## Проверить зависимости
	@echo "Проверка зависимостей..."
	go mod tidy
	go mod verify
	@echo "✅ Зависимости проверены"

# Полный цикл разработки
dev: fmt lint test build ## Полный цикл разработки

# Продакшн сборка
prod: clean deps security-check test build docker-build ## Продакшн сборка

# Помощь по портам
ports-help: ## Справка по портам
	@echo "CloudBridge Relay порты:"
	@echo "80     - HTTP Health Checks"
	@echo "8080   - Main TCP API"
	@echo "8081   - Metrics (Prometheus)"
	@echo "8082   - API"
	@echo "9090   - QUIC Protocol (UDP)"
	@echo "51820  - WireGuard P2P Mesh (UDP)"
	@echo ""
	@echo "Anycast IP: 83.166.235.45"
	@echo "LoadBalancer IP: 109.120.191.126" 