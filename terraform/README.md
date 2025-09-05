# 🚀 CloudBridge Relay - Terraform автоматизация

## 📁 **Структура проекта**

```
terraform/
├── .terraform/                    # Terraform модули
├── .terraform.lock.hcl           # Блокировка провайдеров
├── terraform.tfstate             # Состояние Terraform
├── terraform.tfstate.backup      # Резервное состояние
├── terraform.tfvars              # Переменные конфигурации
├── terraform.tfvars.example      # Пример переменных
│
├── cloudbridge-relay-ddos.tf     # Основная инфраструктура
├── docker-registry-automation.tf # Автоматизация Container Registry
├── cloud-init-ddos.yaml          # Минимальная конфигурация VM
│
├── setup-container-registry.sh   # Настройка Container Registry
├── test-registry-connection.sh   # Тестирование подключения
├── deploy-automated.sh           # Автоматическое развертывание
├── create-registry-cli.sh        # CLI создание (справка)
│
├── AUTOMATION_CONSULTATION.md    # Консультация по автоматизации
├── AUTOMATION_SUMMARY.md         # Обзор автоматизации
├── README_AUTOMATION.md          # Полная инструкция по автоматизации
├── NEXT_STEPS.md                 # Следующие шаги
├── GITHUB_SECRETS_SETUP.md       # Настройка GitHub Secrets
├── CONTAINER_REGISTRY_SETUP_GUIDE.md # Инструкция Container Registry
└── QUICK_REGISTRY_SETUP.md       # Быстрая памятка
```

## 🎯 **Быстрый старт**

### **1. Настройка Container Registry**
```bash
./setup-container-registry.sh
```

### **2. Тестирование подключения**
```bash
source ../load-env.sh
./test-registry-connection.sh
```

### **3. Автоматическое развертывание**
```bash
./deploy-automated.sh
```

## 🔧 **Основные компоненты**

### **Terraform конфигурации:**
- **`cloudbridge-relay-ddos.tf`** - базовая инфраструктура с DDoS защитой
- **`docker-registry-automation.tf`** - автоматизация через Container Registry
- **`cloud-init-ddos.yaml`** - минимальная конфигурация VM

### **Скрипты автоматизации:**
- **`setup-container-registry.sh`** - настройка VK Cloud Container Registry
- **`test-registry-connection.sh`** - тестирование подключения к registry
- **`deploy-automated.sh`** - полное автоматизированное развертывание
- **`create-registry-cli.sh`** - попытка CLI создания (для справки)

### **Документация:**
- **`AUTOMATION_CONSULTATION.md`** - консультация по автоматизации
- **`README_AUTOMATION.md`** - полная инструкция по автоматизации
- **`NEXT_STEPS.md`** - пошаговое руководство
- **`GITHUB_SECRETS_SETUP.md`** - настройка GitHub Secrets
- **`CONTAINER_REGISTRY_SETUP_GUIDE.md`** - инструкция по Container Registry

## 🚀 **Архитектура автоматизации**

```
GitHub Repo → GitHub Actions → VK Cloud Registry → Terraform → VM с готовым образом
```

## 💰 **Стоимость**

- **VK Cloud Container Registry:** ~500₽/месяц
- **VM STD2-4-8:** ~2000₽/месяц
- **GitHub Actions:** 2000 минут/месяц бесплатно

## ⏱️ **Время развертывания**

- **До автоматизации:** 1-2 часа
- **После автоматизации:** 10-15 минут

## 🔒 **Безопасность**

- **DDoS защита:** Nginx + Fail2ban + UFW
- **Контейнерная безопасность:** no-new-privileges, минимальные capabilities
- **Сетевая защита:** Security Groups, firewall правила

## 📞 **Поддержка**

- **VK Cloud:** support@vk.com
- **GitHub:** Issues в репозитории
- **Документация:** см. файлы в директории

---

**🚀 Готово к автоматизации!** Следуйте инструкциям в `NEXT_STEPS.md` для запуска.


