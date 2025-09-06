# Guides Documentation

## 📁 Файлы в этой папке

- **[README_COMPLETE.md](./README_COMPLETE.md)** - Полное руководство по проекту
- **[DOCUMENTATION_STRUCTURE.md](./DOCUMENTATION_STRUCTURE.md)** - Структура документации

## 📖 Руководства

### Быстрый старт
1. **Установка зависимостей**
2. **Настройка окружения**
3. **Запуск локально**
4. **Развертывание в production**

### Разработка
- **Настройка IDE**
- **Запуск тестов**
- **Отладка**
- **Профилирование**

### Развертывание
- **Локальное развертывание**
- **Kubernetes deployment**
- **Terraform deployment**
- **Production deployment**

### Мониторинг
- **Настройка мониторинга**
- **Создание дашбордов**
- **Настройка алертов**
- **Анализ метрик**

### Безопасность
- **Настройка security policies**
- **Security scanning**
- **Compliance проверки**
- **Audit и logging**

## 🛠️ Полезные команды

### Разработка
```bash
# Запуск тестов
make test

# Сборка
make build

# Линтинг
make lint

# Форматирование
make fmt
```

### Docker
```bash
# Сборка образов
./build-docker.sh

# Запуск локально
./run-local.sh

# Очистка
docker system prune -a
```

### Kubernetes
```bash
# Развертывание
./deploy-k8s.sh

# Проверка статуса
kubectl get pods -A

# Логи
kubectl logs -f deployment/cloudbridge-relay
```

### Terraform
```bash
# Инициализация
./terraform/scripts/deploy.sh init prod

# Планирование
./terraform/scripts/deploy.sh plan prod

# Применение
./terraform/scripts/deploy.sh apply prod
```

## 🔧 Troubleshooting

### Общие проблемы
- **Порты заняты** - Проверьте `netstat -tulpn`
- **Docker проблемы** - Перезапустите Docker daemon
- **Kubernetes проблемы** - Проверьте `kubectl cluster-info`

### Логи
```bash
# Docker логи
docker logs container-name

# Kubernetes логи
kubectl logs -f pod-name

# System логи
journalctl -u service-name
```

### Отладка
```bash
# Проверка конфигурации
./scripts/setup/check_dependencies.sh

# Тестирование соединений
./scripts/testing/test-health.sh

# Security проверка
./security/scanning/security-scan.sh
```



