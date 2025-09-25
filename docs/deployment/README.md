# 🚀 **Развертывание CloudBridge Relay**

## **Обзор**

Этот раздел содержит документы по развертыванию CloudBridge Relay в различных средах.

## 📚 **Документы**

### **Отчеты о развертывании:**
- [**Minikube Deployment Report**](MINIKUBE_DEPLOYMENT_REPORT.md) - Подробный отчет о развертывании в Minikube
- [**Minikube Testing Report**](MINIKUBE_TESTING_REPORT.md) - Результаты тестирования в Minikube
- [**Implementation Progress Report**](IMPLEMENTATION_PROGRESS_REPORT.md) - Отчет о прогрессе реализации

## 🎯 **Среды развертывания**

### **Minikube (Разработка):**
- Локальная разработка и тестирование
- Быстрое развертывание
- Отладка и мониторинг

### **Kubernetes (Продакшн):**
- Масштабируемость
- Высокая доступность
- Автоматическое восстановление

##  **Компоненты развертывания**

### **Основные сервисы:**
- **CloudBridge Relay** - Основной relay сервер
- **Prometheus** - Мониторинг и метрики
- **Grafana** - Дашборды
- **Cadence** - Workflow engine

### **Конфигурация:**
- **ConfigMaps** - Конфигурация приложений
- **Secrets** - Секреты и сертификаты
- **Services** - Сетевые сервисы
- **Ingress** - Входящий трафик

## 📊 **Порты и сервисы**

### **CloudBridge Relay порты:**
- **5550**: Main TCP Server
- **5551**: Metrics (Prometheus)
- **5552**: P2P API
- **5553**: HTTP API
- **5554**: HTTPS API
- **5555**: QUIC Transport (UDP)

### **Дополнительные порты:**
- **9090**: QUIC Legacy (UDP)
- **9092**: Enhanced QUIC (UDP)
- **19302**: STUN Server (UDP)
- **8443**: MASQUE Protocol (UDP)

## 🛠️ **Инструменты развертывания**

### **Kubernetes:**
- kubectl
- Helm (опционально)
- Minikube для разработки

### **Мониторинг:**
- Prometheus
- Grafana
- UFW Firewall

## 📝 **Чек-лист развертывания**

### **Подготовка:**
- [ ] Minikube запущен
- [ ] Docker образы собраны
- [ ] Конфигурация готова

### **Развертывание:**
- [ ] Namespace создан
- [ ] ConfigMaps применены
- [ ] Secrets настроены
- [ ] Deployments запущены

### **Проверка:**
- [ ] Pods работают
- [ ] Сервисы доступны
- [ ] Метрики собираются
- [ ] Логи корректны

---

**🎯 Успешное развертывание CloudBridge Relay!**
