# ⚡ Быстрая установка kube-prometheus-stack

## 🎯 Цель
Установить kube-prometheus-stack через VK Cloud Console за 5 минут.

## 📋 Быстрые шаги

### 1. Подготовка
```bash
export KUBECONFIG=~/.kube/kubernetes-cluster-8926_kubeconfig.yaml
kubectl --insecure-skip-tls-verify cluster-info
```

### 2. Установка через VK Cloud Console

1. **VK Cloud Console** → **Kubernetes** → **kubernetes-cluster-8926** → **Аддоны**
2. **Найти** "kube-prometheus-stack" → **Установить**
3. **Название приложения**: `cloudbridge-prometheus`
4. **Пространство имен**: `prometheus-monitoring`
5. **Values**: Вставить содержимое `configs/kube-prometheus-stack-vkcloud.yaml`
6. **Пароль Grafana**: `CloudBridge2025!`
7. **Установить аддон**

### 3. Проверка
```bash
# Проверить поды
kubectl --insecure-skip-tls-verify get pods -n prometheus-monitoring

# Port-forward Grafana
kubectl --insecure-skip-tls-verify -n prometheus-monitoring port-forward service/kube-prometheus-stack-grafana 8001:80
```

### 4. Доступ
- **URL**: http://127.0.0.1:8001
- **Логин**: `admin`
- **Пароль**: `CloudBridge2025!`

## 📁 Файлы конфигурации

- **Основная конфигурация**: `configs/kube-prometheus-stack-vkcloud.yaml`
- **Минимальная конфигурация**: `configs/kube-prometheus-stack-minimal.yaml`
- **Ультра-минимальная**: `configs/kube-prometheus-stack-ultra-minimal.yaml`

## 🚨 Если не работает

1. **Проверить ресурсы**: `kubectl --insecure-skip-tls-verify describe nodes`
2. **Проверить события**: `kubectl --insecure-skip-tls-verify get events -n prometheus-monitoring`
3. **Использовать ультра-минимальную конфигурацию**
4. **Увеличить ресурсы кластера**

## 📚 Подробная инструкция

См. [Пошаговая установка](docs/VK_CLOUD_PROMETHEUS_INSTALL_STEP_BY_STEP.md)

---

**⏱️ Время установки**: 5-10 минут
**🎯 Результат**: Работающий Grafana + Prometheus

