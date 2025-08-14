# 🚀 Финальная установка Docker Registry

## ✅ Подготовка завершена

### Настройки S3:
- **Access Key**: `4MYbD8N3MakzPLNuKQpsGR`
- **Secret Key**: `bbbGeEaeES81NLwNyGVGiANaqkX6QHiUma6f1xGLF4Ng`
- **Бакет**: `dockerimages-8530029fa7b64199b2ad87c18d4620df-bucket`
- **Регион**: `ru-msk`

### Аутентификация:
- **Пользователь**: `cloudbridge`
- **Пароль**: `password`

## 🎯 Установка через VK Cloud Console

### Шаг 1: Открыть VK Cloud Console
1. **VK Cloud Console** → **Kubernetes**
2. **Выбрать кластер**: `kubernetes-cluster-8926`
3. **Перейти в раздел**: **Аддоны**

### Шаг 2: Найти и установить Docker Registry
1. **Найти аддон**: "docker-registry"
2. **Нажать**: **Установить**

### Шаг 3: Заполнить параметры
- **Название приложения**: `docker-registry`
- **Пространство имен**: `kube-system`
- **Values**: скопировать содержимое файла `configs/docker-registry-vkcloud.yaml`

### Шаг 4: Нажать "Установить аддон"

## 🔍 Проверка установки

### Проверить статус подов:
```bash
kubectl --insecure-skip-tls-verify get pods -n kube-system | grep registry
```

### Проверить сервис и получить IP:
```bash
kubectl --insecure-skip-tls-verify get services -n kube-system | grep registry
```

### Получить IP адрес registry:
```bash
kubectl --insecure-skip-tls-verify get services -n kube-system docker-registry -o jsonpath='{.status.loadBalancer.ingress[0].ip}'
```

## 🚀 Автоматическое развертывание микросервисов

После успешной установки registry запустить:

```bash
./scripts/install-and-deploy-auto-ip.sh
```

Этот скрипт автоматически:
- ✅ Определит IP адрес registry
- ✅ Загрузит Docker образы
- ✅ Развернет микросервисы
- ✅ Проверит статус

## 📋 Конфигурация готова

Файл `configs/docker-registry-vkcloud.yaml` содержит:
- ✅ Новые S3 ключи
- ✅ Автоматический IP (без конфликтов)
- ✅ Настроенную аутентификацию
- ✅ S3 хранилище

## 🔗 Полезные команды

### Проверить доступность registry:
```bash
# Заменить <IP> на полученный IP адрес
curl -u cloudbridge:password http://<IP>:5000/v2/_catalog
```

### Посмотреть загруженные образы:
```bash
curl -u cloudbridge:password http://<IP>:5000/v2/_catalog | jq .
```

### Войти в registry:
```bash
echo "password" | docker login <IP>:5000 --username cloudbridge --password-stdin
```

## 🎉 Готово к установке!

Все файлы обновлены с новыми S3 ключами. Можно устанавливать Docker Registry через VK Cloud Console.
