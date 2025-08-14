# 🚀 Установка Prometheus Stack в VK Cloud - Пошагово

## 📋 Шаги установки

### 1. Откройте консоль VK Cloud
- Перейдите на https://mcs.mail.ru/
- Войдите в аккаунт
- Выберите проект с кластером `kubernetes-cluster-8926`

### 2. Перейдите к кластеру
- В меню слева выберите **Kubernetes**
- Найдите кластер `kubernetes-cluster-8926`
- Нажмите на него

### 3. Добавьте аддон
- Перейдите на вкладку **Аддоны**
- Нажмите **Добавить аддон**
- Выберите **kube-prometheus-stack**

### 4. Заполните поля
- **Название приложения**: `cloudbridge-prometheus`
- **Пространство имен**: `cloudbridge-monitoring`
- **Пароль Grafana**: `CloudBridge2025!`

### 5. Вставьте конфигурацию
Скопируйте содержимое файла `configs/kube-prometheus-stack-vkcloud.yaml` в поле **Код настройки аддона**

### 6. Установите аддон
- Нажмите **Установить аддон**
- Дождитесь завершения установки (5-10 минут)

## 🔍 Проверка установки

### Через консоль VK Cloud:
1. Перейдите на вкладку **Ресурсы кластера**
2. Проверьте, что созданы поды в namespace `cloudbridge-monitoring`

### Через kubectl (если настроен):
```bash
kubectl get pods -n cloudbridge-monitoring
kubectl get services -n cloudbridge-monitoring
```

## 🌐 Доступ к веб-интерфейсам

### Способ 1: Через kauthproxy
```bash
# Установите kauthproxy
wget https://github.com/kauthproxy/kauthproxy/releases/latest/download/kauthproxy_linux_amd64.zip
unzip kauthproxy_linux_amd64.zip
sudo mv kauthproxy /usr/local/bin/

# Подключитесь к Grafana
kauthproxy -n cloudbridge-monitoring https://prometheus-grafana.svc

# Подключитесь к Prometheus
kauthproxy -n cloudbridge-monitoring https://prometheus-kube-prometheus-prometheus.svc
```

### Способ 2: Через port-forward
```bash
# Grafana
kubectl -n cloudbridge-monitoring port-forward service/prometheus-grafana 8001:80

# Prometheus
kubectl -n cloudbridge-monitoring port-forward service/prometheus-kube-prometheus-prometheus 9090:9090
```

## 🔑 Доступ к Grafana
- **URL**: http://127.0.0.1:8001 (при port-forward)
- **Логин**: `admin`
- **Пароль**: `CloudBridge2025!`

## 🚨 Если установка не удается

### Проблема: "Не удалось установить аддон"
1. **Проверьте ресурсы кластера** - должно быть минимум 2 CPU и 4GB RAM свободно
2. **Используйте минимальную конфигурацию** из файла `configs/kube-prometheus-stack-vkcloud.yaml`
3. **Проверьте storage классы** - убедитесь что `standard` доступен

### Альтернативное решение:
Если аддон не работает, используйте локальную установку:
```bash
# Остановите локальные контейнеры
docker stop cloudbridge-grafana cloudbridge-prometheus

# Запустите через docker-compose
docker-compose -f docker-compose.monitoring.yml up -d
```

## 📊 Что будет установлено

✅ **Grafana** - веб-интерфейс для визуализации  
✅ **Prometheus** - сбор и хранение метрик  
✅ **Alertmanager** - система алертов  
✅ **Prometheus Operator** - управление ресурсами  

## 🎯 Следующие шаги

После успешной установки:
1. **Настройте дашборды** в Grafana
2. **Добавьте CloudBridge сервисы** в мониторинг
3. **Настройте алерты** для критических событий
4. **Настройте внешний доступ** через Ingress

---

**Статус**: ✅ Готов к установке  
**Время установки**: 10-15 минут  
**Сложность**: Низкая
