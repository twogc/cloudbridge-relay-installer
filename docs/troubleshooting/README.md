# 🔧 Устранение неполадок

Эта папка содержит документацию по диагностике и решению проблем в CloudBridge Relay.

## 📋 Содержание

### 🐛 Отчеты о проблемах
- **CRITICAL_ISSUES_RESOLUTION_REPORT.md** - Решение критических проблем
- **MAIN_GO_FIXES_REPORT.md** - Исправления в Go коде

### 🌐 VK Cloud проблемы
- **VK_CLOUD_ADDONS_TROUBLESHOOTING.md** - Устранение проблем с аддонами

## 🚨 Частые проблемы

### Docker Registry
- **Проблема**: `floating IP is not available`
- **Решение**: Использовать NodePort конфигурацию
- **Файл**: `../installation/REGISTRY_INSTALL_SOLUTION.md`

### Kubernetes
- **Проблема**: Поды не запускаются
- **Решение**: Проверить ресурсы и логи
- **Команда**: `kubectl get pods -n cloudbridge`

### Мониторинг
- **Проблема**: Prometheus недоступен
- **Решение**: Проверить порты и сервисы
- **Файл**: `../installation/VK_CLOUD_ADDONS_TROUBLESHOOTING.md`

## 🔍 Диагностика

### Проверка статуса
```bash
# Статус подов
kubectl --insecure-skip-tls-verify get pods -n cloudbridge

# Логи сервиса
kubectl --insecure-skip-tls-verify logs -n cloudbridge -l app=auth-service-real

# События кластера
kubectl --insecure-skip-tls-verify get events -A --sort-by='.lastTimestamp'
```

### Проверка ресурсов
```bash
# Использование ресурсов
kubectl --insecure-skip-tls-verify top pods -n cloudbridge

# Статус нод
kubectl --insecure-skip-tls-verify get nodes
```

## 📞 Поддержка

При возникновении проблем:
1. Проверить логи и события
2. Изучить соответствующие файлы в этой папке
3. Использовать диагностические команды
4. Обратиться к документации по установке 