# CloudBridge Relay Installer Makefile (fixed)
SHELL := /bin/bash

# ===== Versions & metadata =====
VERSION      ?= $(shell git describe --tags --always --dirty)
BUILD_TIME   ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT   ?= $(shell git rev-parse HEAD)

# ===== Container registry =====
REGISTRY     ?= twogc
IMAGE_NAME   ?= cloudbridge-relay
FULL_IMAGE_NAME   = $(REGISTRY)/$(IMAGE_NAME):$(VERSION)
LATEST_IMAGE_NAME = $(REGISTRY)/$(IMAGE_NAME):latest

# ===== Go params =====
GOOS        ?= linux
GOARCH      ?= amd64
CGO_ENABLED  = 1

# ===== Dirs =====
BIN_DIR      = bin
DIST_DIR     = dist
SCRIPTS_DIR  = scripts
CONFIG_FILE ?= configs/config.json

# ===== Compose compatibility =====
# Use docker compose if available, fall back to docker-compose
ifneq (, $(shell command -v docker))
  ifneq (, $(shell docker compose version >/dev/null 2>&1 && echo ok))
    COMPOSE ?= docker compose
  else
    COMPOSE ?= docker-compose
  endif
else
  COMPOSE ?= docker-compose
endif

COMPOSE_FILE ?= docker-compose.yml

# ===== WireGuard userspace (mirror) =====
# Официальное зеркало wireguard-go на GitHub (чтобы не зависеть от git.zx2c4.com)
WIREGUARD_GO_REPO ?= https://github.com/WireGuard/wireguard-go.git
WIREGUARD_GO_DIR  ?= third_party/wireguard-go

# ===== Common docker run settings for WireGuard =====
# Требуются CAP_NET_ADMIN и устройство /dev/net/tun
# Порты подстрой под себя при необходимости
RUN_NAME      ?= cloudbridge-relay
HOST_CONFIG   ?= -v $(PWD)/$(CONFIG_FILE):/app/configs/config.json:ro
EXPOSE_UDP    ?= -p 51820:51820/udp
EXPOSE_TCP    ?= -p 80:80 -p 8080:8080 -p 8081:8081 -p 8082:8082 -p 9090:9090
WG_DEVICES    ?= --cap-add=NET_ADMIN --device=/dev/net/tun

# ===== PHONY =====
.PHONY: help build clean test docker-build docker-build-nc docker-push \
        docker-run docker-run-direct docker-stop docker-stop-local \
        docker-logs docker-status \
        install-wireguard deps fmt lint security-check prod dev \
        test-ports test-ports-remote test-anycast \
        fetch-wireguard-go clean-wireguard-go \
        deploy-k8s deploy-k8s-anycast test-k8s status clean-all docs

help: ## Показать справку
	@echo "CloudBridge Relay Installer"
	@echo "=========================="
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
	  awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-22s\033[0m %s\n", $$1, $$2}'

# ===== Build & test =====

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

deps: ## Проверить зависимости
	@echo "Проверка зависимостей..."
	go mod tidy
	go mod verify
	@echo "✅ Зависимости проверены"

fmt: ## Форматировать код
	@echo "Форматирование кода..."
	go fmt ./...
	@echo "✅ Форматирование завершено"

lint: ## Запустить линтер (golangci-lint)
	@echo "Запуск линтера..."
	golangci-lint run
	@echo "✅ Линтинг завершен"

security-check: ## Проверка безопасности (gosec)
	@echo "Проверка безопасности..."
	gosec ./...
	@echo "✅ Проверка безопасности завершена"

dev: fmt lint test build ## Полный цикл разработки

prod: clean deps security-check test build docker-build ## Продакшн сборка

# ===== Docker =====

docker-build: ## Собрать Docker образ
	@echo "Сборка Docker образа: $(FULL_IMAGE_NAME)"
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		-t $(FULL_IMAGE_NAME) \
		-t $(LATEST_IMAGE_NAME) .
	@echo "✅ Docker образ собран"

docker-build-nc: ## Собрать Docker образ без кэша
	docker build --no-cache \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		-t $(FULL_IMAGE_NAME) \
		-t $(LATEST_IMAGE_NAME) .

docker-push: ## Отправить Docker образ в registry
	@echo "Отправка Docker образа в registry..."
	docker push $(FULL_IMAGE_NAME)
	docker push $(LATEST_IMAGE_NAME)
	@echo "✅ Docker образ отправлен"

docker-run: ## Запустить через compose (ожидает cap/device в compose)
	@echo "Запуск Docker контейнера (compose)..."
	$(COMPOSE) -f $(COMPOSE_FILE) up -d
	@echo "✅ Docker контейнер запущен"

docker-run-direct: ## Запустить напрямую c NET_ADMIN и /dev/net/tun
	@echo "Запуск Docker контейнера (direct run)..."
	docker run -d --name $(RUN_NAME) \
		$(WG_DEVICES) \
		$(EXPOSE_TCP) $(EXPOSE_UDP) \
		$(HOST_CONFIG) \
		--restart=always \
		$(LATEST_IMAGE_NAME) \
		--config /app/configs/config.json
	@echo "✅ Docker контейнер запущен (direct)"

docker-stop: ## Остановить Docker (compose)
	@echo "Остановка Docker контейнера (compose)..."
	-$(COMPOSE) -f $(COMPOSE_FILE) down || true
	@echo "✅ Docker контейнер остановлен"

docker-stop-local: ## Остановить локальные контейнеры
	@echo "Остановка локальных Docker контейнеров..."
	-docker stop $(RUN_NAME) cloudbridge-relay-alpine cloudbridge-relay-simple 2>/dev/null || true
	-docker rm   $(RUN_NAME) cloudbridge-relay-alpine cloudbridge-relay-simple 2>/dev/null || true
	@echo "✅ Локальные контейнеры остановлены"

docker-logs: ## Показать логи контейнера
	@echo "Логи Docker контейнера..."
	-@docker logs $(RUN_NAME) 2>/dev/null || \
	  docker logs cloudbridge-relay-alpine 2>/dev/null || \
	  docker logs cloudbridge-relay-simple 2>/dev/null || \
	  echo "Контейнер не найден"

docker-status: ## Показать статус контейнеров
	@echo "Статус Docker контейнеров:"
	@docker ps -a | grep -E 'cloudbridge|$(RUN_NAME)' || echo "Контейнеры не найдены"

# ===== WireGuard deps on host =====

install-wireguard: ## Установить WireGuard зависимости на хосте
	@echo "Установка WireGuard зависимостей..."
	@chmod +x $(SCRIPTS_DIR)/install-wireguard.sh
	@$(SCRIPTS_DIR)/install-wireguard.sh
	@echo "✅ WireGuard зависимости установлены"

# ===== Optional: fetch userspace wireguard-go from GitHub mirror =====

fetch-wireguard-go: ## Клонировать wireguard-go (GitHub mirror)
	@echo "Cloning wireguard-go (mirror): $(WIREGUARD_GO_REPO)"
	@rm -rf $(WIREGUARD_GO_DIR)
	@git clone --depth=1 $(WIREGUARD_GO_REPO) $(WIREGUARD_GO_DIR)
	@echo "✅ wireguard-go исходники готовы: $(WIREGUARD_GO_DIR)"

clean-wireguard-go: ## Удалить wireguard-go исходники
	@rm -rf $(WIREGUARD_GO_DIR)
	@echo "✅ wireguard-go очищен"

# ===== Testing ports =====

test-ports: ## Тестировать порты relay сервера (localhost)
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

# ===== Kubernetes =====

deploy-k8s: ## Развернуть в Kubernetes
	@echo "Развертывание в Kubernetes..."
	kubectl apply -f ../terraform-relay/k8s/manifests/cloudbridge-manifests.yaml
	@echo "✅ Развертывание в Kubernetes завершено"

deploy-k8s-anycast: ## Развернуть в Kubernetes с Anycast
	@echo "Развертывание в Kubernetes с Anycast..."
	kubectl apply -f ../terraform-relay/k8s/manifests/cloudbridge-anycast-manifests.yaml
	@echo "✅ Развертывание в Kubernetes с Anycast завершено"

test-k8s: ## Тестировать Kubernetes развертывание
	@echo "Проверка подов:"
	kubectl get pods -n cloudbridge-relay
	@echo "Проверка сервисов:"
	kubectl get services -n cloudbridge-relay
	@echo "Проверка логов (последние 20 строк):"
	kubectl logs -n cloudbridge-relay -l app=cloudbridge-relay --tail=20
	@echo "✅ Тестирование Kubernetes завершено"

# ===== Status & cleanup =====

status: ## Проверить статус сервисов
	@echo "Статус Docker контейнеров:"
	-@$(COMPOSE) ps || docker ps -a
	@echo ""
	@echo "Статус Kubernetes подов:"
	-@kubectl get pods -n cloudbridge-relay 2>/dev/null || echo "Kubernetes namespace не найден"
	@echo ""
	@echo "Статус портов:"
	@$(MAKE) -s test-ports

clean-all: clean docker-stop ## Полная очистка
	@echo "Полная очистка..."
	-docker system prune -f
	-kubectl delete namespace cloudbridge-relay 2>/dev/null || true
	@echo "✅ Полная очистка завершена"

# ===== Docs =====

docs: ## Генерировать документацию
	@echo "Генерация документации..."
	@mkdir -p docs
	@{ \
	  echo "# CloudBridge Relay"; \
	  echo ""; \
	  echo "## Версия: $(VERSION)"; \
	  echo "## Время сборки: $(BUILD_TIME)"; \
	  echo "## Git commit: $(GIT_COMMIT)"; \
	} > docs/README.md
	@echo "✅ Документация сгенерирована"

# ===== Ports help =====

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
