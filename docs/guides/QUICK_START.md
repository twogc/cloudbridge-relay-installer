# 🚀 Быстрый старт - CloudBridge Relay

## 📋 Предварительные требования

1. **Terraform** (версия 1.0+)
2. **OpenStack CLI** (`pip3 install python-openstackclient`)
3. **Доступ к VK Cloud** с правами на создание ресурсов
4. **SSH ключ** для доступа к виртуальной машине

## ⚡ Быстрая установка

### 1. Клонирование репозитория
```bash
git clone https://github.com/twogc/relay-terraform-install.git
cd relay-terraform-install
```

### 2. Настройка аутентификации
```bash
# Скопируйте пример и заполните вашими данными
cp env.example .env
nano .env

# Загрузите переменные
source .env
```

### 3. Создание SSH ключа
```bash
ssh-keygen -t rsa -b 4096 -f ~/.ssh/cloudbridge-relay -N ""
```

### 4. Настройка Terraform
```bash
cd terraform
cp terraform.tfvars.example terraform.tfvars
nano terraform.tfvars  # Укажите ваш SSH публичный ключ
```

### 5. Развертывание
```bash
./deploy.sh deploy
```

## 🔗 Полезные ссылки

- **Полная документация**: [README.md](README.md)
- **VK Cloud**: [mcs.mail.ru](https://mcs.mail.ru/)
- **Поддержка**: info@2gc.ru

## 📊 Что развертывается

- ✅ Виртуальная машина Ubuntu 22.04
- ✅ CloudBridge Relay Server
- ✅ Docker и Docker Compose
- ✅ Nginx веб-сервер
- ✅ Система мониторинга
- ✅ Security Groups
- ✅ Floating IP

## 🌐 Доступные порты

- **8080**: Relay Server API
- **8081**: Relay Server Metrics
- **3456**: Relay Server Custom
- **9090**: QUIC Protocol (UDP)
- **51820**: WireGuard P2P Mesh (UDP)

---

**Создано командой 2GC** | [info@2gc.ru](mailto:info@2gc.ru)



