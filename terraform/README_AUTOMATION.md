# 🚀 CloudBridge Relay - Полная автоматизация развертывания

## **Обзор автоматизации**

Этот проект предоставляет **полностью автоматизированное развертывание** CloudBridge Relay серверов через Terraform с интеграцией VK Cloud Container Registry.

### **Что автоматизировано:**
- **Создание инфраструктуры** через Terraform
- **Установка всех компонентов** (Docker, Nginx, UFW, Fail2ban)
- **Настройка DDoS защиты** и firewall
- **Развертывание CloudBridge Relay** из Container Registry
- **Создание скриптов управления** и мониторинга
- **CI/CD pipeline** через GitHub Actions

## 🎯 **Быстрый старт**

### **1. Настройка Container Registry**

```bash
# Запуск скрипта настройки
./setup-container-registry.sh

# Следуйте инструкциям для создания Container Registry в VK Cloud Console
# Настройте аутентификацию и GitHub Secrets
```

### **2. Тестирование Container Registry**

```bash
# Загрузка переменных окружения
source ../load-env.sh

# Тест подключения
./test-registry-connection.sh
```

### **3. Полное автоматизированное развертывание**

```bash
# Запуск автоматического развертывания
./deploy-automated.sh
```

## 🏗️ **Архитектура автоматизации**

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   GitHub Repo   │───▶│  GitHub Actions  │───▶│ VK Cloud Build │
│                 │    │                  │    │                 │
│ Dockerfile      │    │ Build & Test     │    │ Build Image     │
│ docker-compose  │    │ Push to Registry │    │ Push to VK      │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                │                        │
                                ▼                        ▼
                       ┌──────────────────┐    ┌─────────────────┐
                       │ VK Cloud        │    │ Terraform       │
                       │ Container       │    │ Apply           │
                       │ Registry        │    │ Deploy VM       │
                       └──────────────────┘    └─────────────────┘
                                │                        │
                                ▼                        ▼
                       ┌──────────────────┐    ┌─────────────────┐
                       │ VM с готовым     │    │ CloudBridge     │
                       │ Docker образом   │    │ Relay Server    │
                       └──────────────────┘    └─────────────────┘
```

## 🔧 **Компоненты автоматизации**

### **1. Скрипты настройки**

#### **`setup-container-registry.sh`**
- Создание Container Registry в VK Cloud
- Настройка аутентификации
- Создание конфигурационных файлов

#### **`test-registry-connection.sh`**
- Тест подключения к Container Registry
- Тест сборки и загрузки образов
- Валидация работоспособности

#### **`deploy-automated.sh`**
- Полное автоматизированное развертывание
- Установка всех компонентов
- Настройка сервисов и мониторинга

### **2. Terraform конфигурации**

#### **`docker-registry-automation.tf`**
- Автоматизация через remote-exec provisioners
- Интеграция с VK Cloud Container Registry
- Полная установка CloudBridge Relay

#### **`cloudbridge-relay-ddos.tf`**
- Базовая инфраструктура
- Security Groups и сетевые настройки
- Floating IP и мониторинг

### **3. GitHub Actions**

#### **`.github/workflows/build-and-deploy.yml`**
- Автоматическая сборка при push
- Тестирование Go кода
- Сборка и загрузка Docker образов
- Автоматическое развертывание

## 🚀 **Пошаговое развертывание**

### **Этап 1: Подготовка Container Registry**

```bash
# 1. Создайте Container Registry в VK Cloud Console
# 2. Настройте аутентификацию
# 3. Запустите скрипт настройки
./setup-container-registry.sh

# 4. Обновите .env файл с вашими учетными данными
# 5. Загрузите переменные окружения
source ../load-env.sh
```

### **Этап 2: Тестирование Container Registry**

```bash
# 1. Протестируйте подключение
./test-registry-connection.sh

# 2. Убедитесь, что все тесты пройдены
# 3. Проверьте доступность образов
```

### **Этап 3: Настройка GitHub Secrets**

```bash
# В GitHub репозитории настройте следующие секреты:
VK_REGISTRY_USERNAME = your-username
VK_REGISTRY_PASSWORD = your-password
VK_CLOUD_OPENRC = your-openrc-content
```

### **Этап 4: Автоматическое развертывание**

```bash
# 1. Запустите полное развертывание
./deploy-automated.sh

# 2. Дождитесь завершения всех этапов
# 3. Проверьте статус развертывания
```

## 📊 **Мониторинг и управление**

### **Скрипты управления на VM:**

```bash
# Статус CloudBridge Relay
sudo relay-status

# Статус DDoS защиты
sudo ddos-status

# Управление Docker контейнерами
sudo docker ps -a
sudo docker logs cloudbridge-relay
```

### **Доступные порты:**

- **80/443** - HTTP/HTTPS (Nginx)
- **8080** - Relay Server API
- **8081** - Relay Server Metrics
- **3456** - Relay Server Custom
- **9090** - QUIC Protocol (UDP)
- **51820** - WireGuard P2P Mesh (UDP)

## 🔒 **Безопасность**

### **DDoS защита:**
- **Nginx** - L7 защита
- **Fail2ban** - блокировка подозрительных IP
- **UFW** - настройка firewall правил
- **Security Groups** - сетевой уровень защиты

### **Контейнерная безопасность:**
- **no-new-privileges** - ограничение привилегий
- **Минимальные capabilities** - только необходимые права
- **Read-only volumes** - защита конфигурации
- **Resource limits** - ограничение ресурсов

## 🧪 **Тестирование**

### **Локальное тестирование:**

```bash
# 1. Тест Container Registry
./test-registry-connection.sh

# 2. Тест Terraform конфигурации
terraform plan

# 3. Тест Docker образа
docker build -f ../Dockerfile.relay -t test-image .
docker run --rm test-image
```

### **Тестирование на VM:**

```bash
# 1. Подключение к VM
ssh ubuntu@VM_IP

# 2. Проверка сервисов
sudo relay-status
sudo ddos-status

# 3. Тест API endpoints
curl http://localhost:8080/health
curl http://localhost:8081/metrics
```

## **Устранение неполадок**

### **Ошибки Container Registry:**

```bash
# Проверьте аутентификацию
docker login registry.vkcs.ru

# Проверьте права доступа
# Убедитесь, что registry активен
```

### **Ошибки Terraform:**

```bash
# Переинициализация
terraform init -upgrade

# Проверка состояния
terraform state list
terraform show
```

### **Ошибки на VM:**

```bash
# Проверьте логи cloud-init
sudo cat /var/log/cloud-init-output.log

# Проверьте статус сервисов
sudo systemctl status docker nginx fail2ban

# Проверьте Docker контейнеры
sudo docker ps -a
sudo docker logs cloudbridge-relay
```

## 📈 **Масштабирование**

### **Добавление новых серверов:**

```bash
# 1. Обновите terraform.tfvars
# 2. Запустите развертывание
./deploy-automated.sh

# 3. Или используйте Terraform workspace
terraform workspace new relay-2
terraform apply
```

### **Load Balancer (опционально):**

```bash
# Создайте Load Balancer в VK Cloud
# Настройте health checks
# Добавьте backend servers
```

## 💰 **Стоимость и ресурсы**

### **VK Cloud:**
- **Container Registry:** ~500₽/месяц
- **VM STD2-4-8:** ~2000₽/месяц
- **Floating IP:** ~100₽/месяц
- **Storage:** ~10₽/GB/месяц

### **GitHub Actions:**
- **2000 минут/месяц** - бесплатно
- **Дополнительные минуты** - $0.008/минута

## 🎉 **Результаты автоматизации**

После реализации автоматизации:

- **Время развертывания:** с 1-2 часов до 10-15 минут
- **Консистентность:** 100% идентичные серверы
- **Масштабируемость:** легко добавлять новые серверы
- **Надежность:** автоматическое обновление и rollback
- **Мониторинг:** полная видимость состояния серверов

## 📞 **Поддержка**

### **Документация:**
- [VK Cloud Container Registry](https://mcs.mail.ru/docs/container-registry/)
- [Terraform VK Cloud Provider](https://registry.terraform.io/providers/vk-cs/vkcs/latest/docs)
- [GitHub Actions](https://docs.github.com/en/actions)

### **Сообщество:**
- **VK Cloud:** @vkcloud_support
- **GitHub:** Issues в репозитории
- **Telegram:** @cloudbridge_relay

---

**🚀 Готово к автоматизации!** Выберите подходящий этап и начните автоматизацию ваших CloudBridge Relay серверов.
