# 🚀 Ультра-минимальная установка Grafana в VK Cloud - ПОШАГОВО

## 📋 Конфигурация для копирования

Скопируйте **ВСЕ** содержимое ниже в поле "Код настройки аддона" в консоли VK Cloud:

```yaml
# УЛЬТРА-МИНИМАЛЬНАЯ конфигурация для VK Cloud
# Только Grafana без всего остального

nameOverride: grafana
fullnameOverride: grafana

# Только Grafana
grafana:
  enabled: true
  adminPassword: "admin123"
  persistence:
    enabled: false
  resources:
    requests:
      cpu: 10m
      memory: 16Mi
    limits:
      cpu: 25m
      memory: 32Mi

# Отключаем ВСЕ остальное
prometheus:
  enabled: false

alertmanager:
  enabled: false

prometheusOperator:
  enabled: false

kube-state-metrics:
  enabled: false

prometheus-node-exporter:
  enabled: false

kubeApiServer:
  enabled: false

kubelet:
  enabled: false

coreDns:
  enabled: false

kubeEtcd:
  enabled: false

kubeScheduler:
  enabled: false

kubeControllerManager:
  enabled: false
```

## 🎯 Пошаговая установка

### Шаг 1: Откройте консоль VK Cloud
1. Перейдите на https://mcs.mail.ru/
2. Войдите в аккаунт `info@2gc.ru`
3. Выберите проект с кластером `kubernetes-cluster-8926`

### Шаг 2: Перейдите к кластеру
1. В меню слева нажмите **Kubernetes**
2. Найдите кластер `kubernetes-cluster-8926`
3. Нажмите на него

### Шаг 3: Добавьте аддон
1. Перейдите на вкладку **Аддоны**
2. Нажмите **Добавить аддон**
3. В списке найдите **kube-prometheus-stack**
4. Нажмите **Установить**

### Шаг 4: Заполните поля
1. **Название приложения**: `grafana-only`
2. **Пространство имен**: `grafana-monitoring`
3. **Пароль Grafana**: `admin123`

### Шаг 5: Вставьте конфигурацию
1. В поле **Код настройки аддона** вставьте конфигурацию выше
2. Убедитесь, что вставлен весь код (от `# УЛЬТРА-МИНИМАЛЬНАЯ` до `enabled: false`)

### Шаг 6: Установите
1. Нажмите **Установить аддон**
2. Дождитесь завершения (5-10 минут)

## 🔍 Проверка установки

### В консоли VK Cloud:
1. Перейдите на вкладку **Ресурсы кластера**
2. Выберите **Поды**
3. Найдите namespace `grafana-monitoring`
4. Должен быть один под `grafana-only-grafana-xxx`
5. Статус должен быть **Running**

## 🌐 Доступ к Grafana

### Способ 1: Через консоль VK Cloud
1. Перейдите на вкладку **Ресурсы кластера**
2. Выберите **Сервисы**
3. Найдите сервис `grafana-only-grafana`
4. Нажмите на него
5. Используйте **Внешний IP** или **Внутренний IP**

### Способ 2: Через port-forward (если kubectl работает)
```bash
kubectl -n grafana-monitoring port-forward service/grafana-only-grafana 3000:80
```

## 🔑 Логин в Grafana
- **URL**: http://[IP-адрес]:3000
- **Логин**: `admin`
- **Пароль**: `admin123`

## ⚠️ Что будет установлено

✅ **Только Grafana** - веб-интерфейс для визуализации  
✅ **Минимальные ресурсы** - 10m CPU, 16Mi RAM  
✅ **Без persistent storage** - данные не сохраняются  

❌ **НЕ будет установлено**:
- Prometheus
- Alertmanager
- Prometheus Operator
- Node Exporter
- Kube State Metrics

## 🚨 Если установка не удается

### Проблема: "Не удалось установить аддон"
1. **Проверьте ресурсы кластера** - должно быть минимум 1 CPU и 2GB RAM свободно
2. **Попробуйте еще раз** - иногда нужно несколько попыток
3. **Используйте локальную установку** как альтернативу

### Альтернатива - локальная установка:
```bash
# Остановите локальные контейнеры
docker stop cloudbridge-grafana cloudbridge-prometheus

# Запустите только Grafana
docker run -d --name grafana-local -p 3000:3000 \
  -e GF_SECURITY_ADMIN_PASSWORD=admin123 \
  grafana/grafana:10.1.5
```

## 📊 Следующие шаги

После успешной установки Grafana:
1. **Добавьте Prometheus** как источник данных (локальный или внешний)
2. **Настройте дашборды** для CloudBridge сервисов
3. **Добавьте алерты** для критических событий

---

**Статус**: ✅ Готов к установке  
**Время установки**: 5-10 минут  
**Сложность**: Очень низкая  
**Ресурсы**: Минимальные (10m CPU, 16Mi RAM)
