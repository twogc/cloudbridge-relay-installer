# 🔧 Решение проблемы установки Docker Registry

## ❌ Проблема
Ошибка: `floating IP 90.156.150.25 is not available`

**Причина**: Плавающий IP `90.156.150.25` уже используется сервисом `ingress-nginx-controller`

## ✅ Решения

### Вариант 1: NodePort (рекомендуется)

#### 1. Установить Docker Registry с NodePort
**VK Cloud Console** → **Kubernetes** → **Аддоны** → **docker-registry**

**Параметры:**
- **Название**: `docker-registry`
- **Namespace**: `kube-system`
- **Values**: содержимое `configs/docker-registry-nodeport.yaml`

#### 2. Автоматическое развертывание
```bash
./scripts/install-and-deploy-nodeport.sh
```

**Преимущества:**
- ✅ Не требует дополнительных плавающих IP
- ✅ Использует существующий IP ноды
- ✅ Порт: `30500`

### Вариант 2: Создать новый плавающий IP

#### 1. Создать новый плавающий IP
**VK Cloud Console** → **Сети** → **Плавающие IP** → **Создать**

#### 2. Обновить конфигурацию
Заменить в `configs/docker-registry-vkcloud.yaml`:
```yaml
service:
  name: registry
  type: LoadBalancer
  loadBalancerIP: "НОВЫЙ_IP_АДРЕС"  # Вставить новый IP
  port: 5000
  nodePort: 5000
```

#### 3. Установить и развернуть
```bash
./scripts/install-and-deploy-auto-ip.sh
```

### Вариант 3: Автоматический IP (попробовать снова)

#### 1. Установить с автоматическим IP
**VK Cloud Console** → **Kubernetes** → **Аддоны** → **docker-registry**

**Параметры:**
- **Название**: `docker-registry`
- **Namespace**: `kube-system`
- **Values**: содержимое `configs/docker-registry-vkcloud.yaml`

#### 2. Автоматическое развертывание
```bash
./scripts/install-and-deploy-auto-ip.sh
```

## 🎯 Рекомендуемое решение: NodePort

### Почему NodePort лучше:
1. **Не требует дополнительных IP** - использует IP ноды
2. **Нет конфликтов** с существующими сервисами
3. **Простая установка** - один файл конфигурации
4. **Автоматический скрипт** - полная автоматизация

### Файлы для NodePort:
- ✅ `configs/docker-registry-nodeport.yaml` - конфигурация
- ✅ `scripts/install-and-deploy-nodeport.sh` - автоматический скрипт

## 📋 Пошаговая инструкция (NodePort)

### Шаг 1: Установить Docker Registry
1. **VK Cloud Console** → **Kubernetes** → **kubernetes-cluster-8926** → **Аддоны**
2. **Найти** "docker-registry" → **Установить**
3. **Название**: `docker-registry`
4. **Namespace**: `kube-system`
5. **Values**: скопировать `configs/docker-registry-nodeport.yaml`

### Шаг 2: Запустить автоматический скрипт
```bash
./scripts/install-and-deploy-nodeport.sh
```

### Шаг 3: Проверить результат
```bash
# Проверить статус
kubectl --insecure-skip-tls-verify get pods -n cloudbridge

# Получить URL registry
kubectl --insecure-skip-tls-verify get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="ExternalIP")].address}'
# Добавить :30500 к полученному IP
```

## 🔗 Полезные команды

### Проверить статус registry:
```bash
kubectl --insecure-skip-tls-verify get pods -n kube-system | grep registry
kubectl --insecure-skip-tls-verify get services -n kube-system | grep registry
```

### Проверить доступность:
```bash
# Заменить <NODE_IP> на IP ноды
curl -u cloudbridge:password http://<NODE_IP>:30500/v2/_catalog
```

### Посмотреть образы:
```bash
curl -u cloudbridge:password http://<NODE_IP>:30500/v2/_catalog | jq .
```

## 🎉 Результат

После успешной установки:
- ✅ Docker Registry работает на порту `30500`
- ✅ Микросервисы развернуты
- ✅ Образы загружены в registry
- ✅ Готово к использованию

**Рекомендуется использовать NodePort версию для избежания конфликтов IP!**
