# Быстрая установка Docker Registry (автоматический IP)

## 🚀 Установка через VK Cloud Console

### 1. Подготовка
- ✅ S3 бакет создан: `dockerimages-8530029fa7b64199b2ad87c18d4620df-bucket`
- ✅ S3 ключи обновлены:
  - Access Key: `4MYbD8N3MakzPLNuKQpsGR`
  - Secret Key: `bbbGeEaeES81NLwNyGVGiANaqkX6QHiUma6f1xGLF4Ng`
- ✅ Плавающий IP не нужен (VK Cloud назначит автоматически)

### 2. Установка аддона
1. **VK Cloud Console** → **Kubernetes** → **kubernetes-cluster-8926** → **Аддоны**
2. **Найти** "docker-registry" → **Установить**
3. **Название приложения**: `docker-registry`
4. **Пространство имен**: `kube-system`
5. **Values**: скопировать содержимое `configs/docker-registry-vkcloud.yaml`

### 3. Проверка установки
```bash
# Проверить статус подов
kubectl --insecure-skip-tls-verify get pods -n kube-system | grep registry

# Проверить сервис и получить IP
kubectl --insecure-skip-tls-verify get services -n kube-system | grep registry
```

### 4. Автоматическое развертывание
```bash
# Запустить автоматический скрипт
./scripts/install-and-deploy-auto-ip.sh
```

## 📋 Конфигурация готова
- ✅ Автоматический IP (без конфликтов)
- ✅ S3 хранилище настроено
- ✅ Аутентификация: `cloudbridge:password`
- ✅ Порт: `5000`

## 🔗 Полезные команды
```bash
# Получить IP registry
kubectl --insecure-skip-tls-verify get services -n kube-system docker-registry -o jsonpath='{.status.loadBalancer.ingress[0].ip}'

# Проверить доступность
curl -u cloudbridge:password http://<IP>:5000/v2/_catalog

# Посмотреть образы
curl -u cloudbridge:password http://<IP>:5000/v2/_catalog | jq .
```
