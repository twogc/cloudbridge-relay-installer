# 🚀 **CloudBridge Relay**

**QUIC-first + MASQUE глобальная Zero-Trust overlay сеть**

## **Быстрый старт**

### **Развертывание:**
```bash
# Следуйте инструкциям в MINIKUBE_BUILD_GUIDE.md
# Или изучите документацию в docs/deployment/
```

### **Документация:**
📚 **[Полная документация находится в папке `docs/`](docs/README.md)**

## 🎯 **Основные компоненты**

- **CloudBridge Relay** - Основной relay сервер
- **QUIC/HTTP-3** - Базовый транспортный протокол  
- **MASQUE** - Протокол для туннелей поверх HTTP-3
- **ICE/STUN/TURN** - NAT traversal
- **Zero-Trust** - Модель безопасности

##  **Порты**

| Порт | Протокол | Назначение |
|------|----------|------------|
| 5550 | TCP | Main Server |
| 5551 | TCP | Metrics (Prometheus) |
| 5552 | TCP | P2P API |
| 5553 | TCP | HTTP API |
| 5554 | TCP | HTTPS API |
| 5555 | UDP | QUIC Transport |

## 📚 **Документация**

### **🏗️ Архитектура**
- [Обзор системы](docs/architecture/2GC_CLOUDBRIDGE_OVERVIEW.md)
- [MASQUE протокол](docs/architecture/MASQUE_USAGE_GUIDE.md)
- [План реализации](docs/architecture/IMPLEMENTATION_PLAN.md)

### **🚀 Развертывание**
- [Minikube развертывание](docs/deployment/MINIKUBE_DEPLOYMENT_REPORT.md)
- [Тестирование](docs/deployment/MINIKUBE_TESTING_REPORT.md)
- [Прогресс реализации](docs/deployment/IMPLEMENTATION_PROGRESS_REPORT.md)

### **🔒 Безопасность**
- [Keycloak интеграция](docs/security/KEYCLOAK_ANALYSIS_RESULTS.md)
- [Тестирование безопасности](docs/security/KEYCLOAK_TESTING_RESULTS.md)
- [Исправления](docs/security/QUICK_KEYCLOAK_FIX.md)

### **📊 Мониторинг**
- [Анализ портов](docs/monitoring/PORT_ANALYSIS_REPORT.md)
- [HTTP3 тестирование](docs/monitoring/HTTP3_TESTING_REPORT.md)
- [UFW настройка](docs/UFW_FIREWALL_GUIDE.md)

### **📖 Руководства**
- [TURN сервер](docs/guides/TURN_SERVER_IMPLEMENTATION_REPORT.md)
- [Настройка репозиториев](docs/guides/REFERENCE_REPOS_SETUP_REPORT.md)
- [Отладка MASQUE](docs/guides/DEVELOPER_TASK_MASQUE_DEBUG.md)

### **Отчеты**
- [Анализ кода](docs/reports/CODE_ANALYSIS_REPORT.md)
- [Готовность MVP](docs/reports/MVP_READINESS_ANALYSIS.md)
- [Чек-лист разработки](docs/reports/MVP_DEVELOPMENT_CHECKLIST.md)

## 🛠️ **Быстрые команды**

### **Развертывание:**
```bash
# Сборка и развертывание в Minikube
TAG="mvp-$(date +%Y%m%d-%H%M%S)"
docker build -t cloudbridge-relay:$TAG .
minikube image load cloudbridge-relay:$TAG
kubectl set image deployment/cloudbridge-relay cloudbridge-relay=cloudbridge-relay:$TAG -n cloudbridge
```

### **Мониторинг:**
```bash
# Проверка статуса
kubectl get pods -n cloudbridge

# Логи
kubectl logs -f deployment/cloudbridge-relay -n cloudbridge

# Метрики
kubectl port-forward svc/cloudbridge-relay 5551:5551 -n cloudbridge
curl http://localhost:5551/metrics
```

### **UFW настройка:**
```bash
# Добавить правила
sudo ./scripts/ufw-cloudbridge.sh add

# Показать правила
sudo ./scripts/ufw-cloudbridge.sh show
```

## 📞 **Поддержка**

- 📚 **[Полная документация](docs/README.md)**
-  **[Руководства](docs/guides/README.md)**
- 📊 **[Отчеты](docs/reports/README.md)**
- 🚀 **[Развертывание](docs/deployment/README.md)**

## 🎯 **Статус проекта**

- **Архитектура**: Готова
- **Развертывание**: Работает в Minikube
- **Безопасность**: Keycloak интегрирован
- **Мониторинг**: Prometheus настроен
- **Порты**: 5550-5555 настроены
- **UFW**: Правила настроены
- **Документация**: Структурирована

---

**🎯 CloudBridge Relay - Ваш путь к QUIC-first Zero-Trust сети!**
