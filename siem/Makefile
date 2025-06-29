# CloudBridge SIEM Makefile
# Удобные команды для управления SIEM системой

.PHONY: help start stop restart status logs test backup restore health clean info install build

# Переменные
SIEM_DIR = $(shell pwd)
DOCKER_COMPOSE = docker-compose -f docker-compose.siem.yml

help: ## Показать справку
	@echo "CloudBridge SIEM Management"
	@echo "=========================="
	@echo ""
	@echo "Доступные команды:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

install: ## Установка зависимостей
	@echo "📦 Установка зависимостей Python..."
	pip install -r requirements.txt
	@echo "✅ Зависимости установлены"

build: ## Сборка Docker образов
	@echo "🔨 Сборка Docker образов..."
	$(DOCKER_COMPOSE) build
	@echo "✅ Образы собраны"

start: ## Запуск SIEM системы
	@echo "🚀 Запуск CloudBridge SIEM..."
	$(DOCKER_COMPOSE) up -d
	@echo "✅ SIEM система запущена"
	@echo "📊 Доступные сервисы:"
	@echo "   - Elasticsearch: http://localhost:9200"
	@echo "   - Kibana: http://localhost:5601"
	@echo "   - Grafana: http://localhost:3001"
	@echo "   - Prometheus: http://localhost:9091"
	@echo "   - SIEM Web UI: http://localhost:8081"

stop: ## Остановка SIEM системы
	@echo "🛑 Остановка CloudBridge SIEM..."
	$(DOCKER_COMPOSE) down
	@echo "✅ SIEM система остановлена"

restart: stop start ## Перезапуск SIEM системы

status: ## Статус SIEM системы
	@echo "📊 Статус CloudBridge SIEM..."
	$(DOCKER_COMPOSE) ps

logs: ## Просмотр логов всех сервисов
	@echo "📋 Логи CloudBridge SIEM..."
	$(DOCKER_COMPOSE) logs -f

logs-elasticsearch: ## Логи Elasticsearch
	$(DOCKER_COMPOSE) logs -f elasticsearch

logs-kibana: ## Логи Kibana
	$(DOCKER_COMPOSE) logs -f kibana

logs-grafana: ## Логи Grafana
	$(DOCKER_COMPOSE) logs -f grafana

logs-prometheus: ## Логи Prometheus
	$(DOCKER_COMPOSE) logs -f prometheus

logs-collector: ## Логи SIEM Collector
	$(DOCKER_COMPOSE) logs -f siem-collector

logs-processor: ## Логи SIEM Processor
	$(DOCKER_COMPOSE) logs -f siem-processor

logs-alertmanager: ## Логи SIEM Alert Manager
	$(DOCKER_COMPOSE) logs -f siem-alertmanager

test: ## Запуск тестов SIEM
	@echo "🧪 Запуск тестов CloudBridge SIEM..."
	python test_siem.py
	@echo "✅ Тесты завершены"

backup: ## Создание резервной копии
	@echo "💾 Создание резервной копии SIEM..."
	@mkdir -p backups
	@timestamp=$$(date +%Y%m%d_%H%M%S); \
	backup_dir="backups/siem_backup_$$timestamp"; \
	mkdir -p $$backup_dir; \
	cp config/siem_config.yaml $$backup_dir/; \
	cp docker-compose.siem.yml $$backup_dir/; \
	echo "✅ Резервная копия создана: $$backup_dir"

restore: ## Восстановление из резервной копии (укажите BACKUP_DIR)
	@if [ -z "$(BACKUP_DIR)" ]; then \
		echo "❌ Укажите BACKUP_DIR=путь/к/резервной/копии"; \
		exit 1; \
	fi
	@echo "🔄 Восстановление из $(BACKUP_DIR)..."
	@if [ -f "$(BACKUP_DIR)/siem_config.yaml" ]; then \
		cp "$(BACKUP_DIR)/siem_config.yaml" config/; \
		echo "✅ Конфигурация восстановлена"; \
	fi
	@if [ -f "$(BACKUP_DIR)/docker-compose.siem.yml" ]; then \
		cp "$(BACKUP_DIR)/docker-compose.siem.yml" .; \
		echo "✅ Docker Compose восстановлен"; \
	fi

health: ## Проверка здоровья системы
	@echo "🏥 Проверка здоровья CloudBridge SIEM..."
	@python manage_siem.py health

clean: ## Очистка данных SIEM
	@echo "🧹 Очистка данных CloudBridge SIEM..."
	@read -p "⚠️  Это удалит все данные SIEM. Продолжить? (y/N): " confirm; \
	if [ "$$confirm" = "y" ] || [ "$$confirm" = "Y" ]; then \
		$(DOCKER_COMPOSE) down -v; \
		echo "✅ Данные очищены"; \
	else \
		echo "❌ Операция отменена"; \
	fi

info: ## Информация о SIEM системе
	@echo "ℹ️  Информация о CloudBridge SIEM"
	@echo "================================"
	@echo "Версия: 2.0.0"
	@echo "Описание: Система информационной безопасности и мониторинга событий"
	@echo ""
	@echo "Компоненты:"
	@echo "  - Elasticsearch - хранение и поиск данных"
	@echo "  - Kibana - визуализация и анализ"
	@echo "  - Grafana - мониторинг метрик"
	@echo "  - Prometheus - сбор метрик"
	@echo "  - Kafka - потоковая обработка"
	@echo "  - Redis - кэширование"
	@echo "  - SIEM Collector - сбор данных"
	@echo "  - SIEM Processor - обработка событий"
	@echo "  - SIEM Alert Manager - уведомления"
	@echo ""
	@echo "Порты:"
	@echo "  - Elasticsearch: 9200"
	@echo "  - Kibana: 5601"
	@echo "  - Grafana: 3001"
	@echo "  - Prometheus: 9091"
	@echo "  - SIEM Web UI: 8081"
	@echo "  - Kafka: 9092"
	@echo "  - Redis: 6379"

setup: ## Первоначальная настройка SIEM
	@echo "⚙️  Первоначальная настройка CloudBridge SIEM..."
	@if [ ! -f "config/siem_config.yaml" ]; then \
		echo "❌ Файл конфигурации не найден. Создайте config/siem_config.yaml"; \
		exit 1; \
	fi
	@echo "📦 Установка зависимостей..."
	@make install
	@echo "🔨 Сборка образов..."
	@make build
	@echo "🚀 Запуск системы..."
	@make start
	@echo "⏳ Ожидание запуска сервисов..."
	@sleep 30
	@echo "🧪 Запуск тестов..."
	@make test
	@echo "✅ Настройка завершена!"

update: ## Обновление SIEM системы
	@echo "🔄 Обновление CloudBridge SIEM..."
	@git pull origin main
	@make install
	@make build
	@make restart
	@echo "✅ Система обновлена"

monitor: ## Мониторинг в реальном времени
	@echo "📊 Мониторинг CloudBridge SIEM..."
	@echo "Нажмите Ctrl+C для выхода"
	@watch -n 5 'make status'

# Дополнительные команды для разработки
dev-start: ## Запуск в режиме разработки
	@echo "🔧 Запуск в режиме разработки..."
	$(DOCKER_COMPOSE) -f docker-compose.siem.yml -f docker-compose.dev.yml up -d

dev-logs: ## Логи в режиме разработки
	$(DOCKER_COMPOSE) -f docker-compose.siem.yml -f docker-compose.dev.yml logs -f

lint: ## Проверка кода
	@echo "🔍 Проверка кода..."
	@python -m flake8 collectors/ processors/ alerting/ webui/
	@python -m black --check collectors/ processors/ alerting/ webui/

format: ## Форматирование кода
	@echo "🎨 Форматирование кода..."
	@python -m black collectors/ processors/ alerting/ webui/

security-scan: ## Сканирование безопасности
	@echo "🔒 Сканирование безопасности..."
	@python -m bandit -r collectors/ processors/ alerting/ webui/

# Команды для продакшена
prod-start: ## Запуск в продакшене
	@echo "🚀 Запуск в продакшене..."
	@if [ ! -f ".env.prod" ]; then \
		echo "❌ Файл .env.prod не найден"; \
		exit 1; \
	fi
	$(DOCKER_COMPOSE) -f docker-compose.siem.yml --env-file .env.prod up -d

prod-stop: ## Остановка в продакшене
	$(DOCKER_COMPOSE) -f docker-compose.siem.yml --env-file .env.prod down

prod-backup: ## Резервная копия продакшена
	@echo "💾 Резервная копия продакшена..."
	@timestamp=$$(date +%Y%m%d_%H%M%S); \
	backup_dir="backups/prod_backup_$$timestamp"; \
	mkdir -p $$backup_dir; \
	cp .env.prod $$backup_dir/; \
	cp config/siem_config.yaml $$backup_dir/; \
	docker exec siem-elasticsearch elasticsearch-dump --input=http://localhost:9200/siem-* --output=/tmp/prod_backup_$$timestamp.json; \
	docker cp siem-elasticsearch:/tmp/prod_backup_$$timestamp.json $$backup_dir/; \
	echo "✅ Резервная копия продакшена создана: $$backup_dir" 