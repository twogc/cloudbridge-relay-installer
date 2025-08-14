# 🚀 Установка Grafana в новом кластере kubernetes-cluster-6379

## 📋 Информация о кластере

- **ID кластера**: cd494d7f-937c-4cf1-be99-9f2d75e43112
- **API сервер**: https://90.156.150.25:6443
- **Ресурсы**: 3 vCPU / 8 ГБ RAM
- **Версия Kubernetes**: 1.32.1
- **Доверенные Registry**: docker.io, registry.infra.mail.ru:5010

## 🎯 Конфигурация для копирования

Скопируйте **ВСЕ** содержимое ниже в поле "Код настройки аддона":

```yaml
# Конфигурация для нового кластера kubernetes-cluster-6379
# Ультра-минимальная установка Grafana

nameOverride: grafana
fullnameOverride: grafana

# Глобальные настройки для VK Cloud
global:
  imageRegistry: "docker.io"

# Только Grafana - минимальная конфигурация
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
  image:
    repository: "grafana/grafana"
    tag: "10.1.5"

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

## 📝 Шаги установки

### 1. Откройте консоль VK Cloud
- https://mcs.mail.ru/
- Войдите в аккаунт `info@2gc.ru`
- Выберите проект с кластером `kubernetes-cluster-6379`

### 2. Перейдите к кластеру
- Меню слева → **Kubernetes**
- Найдите `kubernetes-cluster-6379`
- Нажмите на него

### 3. Добавьте аддон
- Вкладка **Аддоны**
- **Добавить аддон**
- Выберите **kube-prometheus-stack**

### 4. Заполните поля
- **Название приложения**: `grafana-only`
- **Пространство имен**: `grafana-monitoring`
- **Пароль Grafana**: `admin123`

### 5. Вставьте конфигурацию
Вставьте конфигурацию выше в поле **Код настройки аддона**

### 6. Установите
- **Установить аддон**
- Ждите 5-10 минут

## 🔍 Проверка установки

### В консоли VK Cloud:
1. **Ресурсы кластера** → **Поды**
2. Найдите namespace `grafana-monitoring`
3. Должен быть под `grafana-only-grafana-xxx`
4. Статус: **Running**

## 🌐 Доступ к Grafana

### Способ 1: Через консоль VK Cloud
1. **Ресурсы кластера** → **Сервисы**
2. Найдите `grafana-only-grafana`
3. Используйте **Внешний IP** или **Внутренний IP**

### Способ 2: Port-forward
```bash
kubectl -n grafana-monitoring port-forward service/grafana-only-grafana 3000:80
```

## 🔑 Логин в Grafana
- **URL**: http://[IP]:3000
- **Логин**: `admin`
- **Пароль**: `admin123`

## ⚠️ Преимущества нового кластера

✅ **Больше ресурсов**: 3 vCPU / 8 ГБ RAM (vs 6 vCPU / 12 ГБ RAM)  
✅ **Правильные registry**: docker.io, registry.infra.mail.ru:5010  
✅ **Свежая установка**: нет конфликтов с предыдущими аддонами  
✅ **Минимальная конфигурация**: только Grafana  

## 🚨 Если не работает

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

После успешной установки:
1. **Добавьте Prometheus** как источник данных
2. **Настройте дашборды** для CloudBridge
3. **Добавьте алерты** для критических событий

---

**Статус**: ✅ Готов к установке  
**Кластер**: kubernetes-cluster-6379  
**Время установки**: 5-10 минут  
**Сложность**: Очень низкая
