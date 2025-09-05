# 🧹 **План очистки Terraform директории**

## 📋 **Анализ текущих файлов**

### **🔴 Файлы для удаления (устаревшие/дублирующие):**

#### **Backup файлы:**
- `cloudbridge-relay-ddos.tf.backup` - резервная копия, больше не нужна
- `cloudbridge-relay.tf.backup` - резервная копия, больше не нужна

#### **Устаревшие скрипты:**
- `deploy-with-autoinstall.sh` - заменен на `deploy-automated.sh`
- `deploy.sh` - старый скрипт, заменен на автоматизированный
- `cloudbridge-relay-autoinstall.tf` - функциональность перенесена в `docker-registry-automation.tf`

#### **Устаревшие конфигурации:**
- `cloud-init.yaml` - заменен на `cloud-init-ddos.yaml`
- `tfplan` - временный файл, можно удалить

#### **Устаревшая документация:**
- `AUTOMATION_README.md` - заменен на `README_AUTOMATION.md`
- `CLOUDBRIDGE_RELAY_ARCHITECTURE.md` - устаревшая архитектура
- `CLOUDBRIDGE_RELAY_CORRECT_ARCHITECTURE.md` - устаревшая архитектура
- `PROJECT_OVERVIEW.md` - устаревший обзор проекта

### **🟢 Файлы для сохранения (актуальные):**

#### **Основные Terraform конфигурации:**
- `cloudbridge-relay-ddos.tf` - основная инфраструктура
- `docker-registry-automation.tf` - автоматизация через Container Registry
- `cloud-init-ddos.yaml` - минимальная конфигурация VM

#### **Скрипты автоматизации:**
- `setup-container-registry.sh` - настройка Container Registry
- `test-registry-connection.sh` - тестирование подключения
- `deploy-automated.sh` - полное автоматизированное развертывание
- `create-registry-cli.sh` - попытка CLI создания (для справки)

#### **Актуальная документация:**
- `AUTOMATION_CONSULTATION.md` - консультация по автоматизации
- `AUTOMATION_SUMMARY.md` - обзор автоматизации
- `README_AUTOMATION.md` - полная инструкция по автоматизации
- `NEXT_STEPS.md` - следующие шаги
- `GITHUB_SECRETS_SETUP.md` - настройка GitHub Secrets
- `CONTAINER_REGISTRY_SETUP_GUIDE.md` - инструкция по Container Registry
- `QUICK_REGISTRY_SETUP.md` - быстрая памятка

#### **Конфигурационные файлы:**
- `terraform.tfvars` - переменные Terraform
- `terraform.tfvars.example` - пример переменных
- `terraform.tfstate` - состояние Terraform (не удалять!)
- `terraform.tfstate.backup` - резервное состояние (не удалять!)
- `.terraform.lock.hcl` - блокировка провайдеров (не удалять!)

## 🗑️ **Команды для очистки:**

```bash
# Удаление backup файлов
rm cloudbridge-relay-ddos.tf.backup
rm cloudbridge-relay.tf.backup

# Удаление устаревших скриптов
rm deploy-with-autoinstall.sh
rm deploy.sh
rm cloudbridge-relay-autoinstall.tf

# Удаление устаревших конфигураций
rm cloud-init.yaml
rm tfplan

# Удаление устаревшей документации
rm AUTOMATION_README.md
rm CLOUDBRIDGE_RELAY_ARCHITECTURE.md
rm CLOUDBRIDGE_RELAY_CORRECT_ARCHITECTURE.md
rm PROJECT_OVERVIEW.md
```

## 📁 **Структура после очистки:**

```
terraform/
├── .terraform/                    # Terraform модули (не трогать)
├── .terraform.lock.hcl           # Блокировка провайдеров
├── terraform.tfstate             # Состояние (не трогать)
├── terraform.tfstate.backup      # Резервное состояние (не трогать)
├── terraform.tfvars              # Переменные
├── terraform.tfvars.example      # Пример переменных
│
├── cloudbridge-relay-ddos.tf     # Основная инфраструктура
├── docker-registry-automation.tf # Автоматизация Container Registry
├── cloud-init-ddos.yaml          # Конфигурация VM
│
├── setup-container-registry.sh   # Настройка Container Registry
├── test-registry-connection.sh   # Тестирование подключения
├── deploy-automated.sh           # Автоматическое развертывание
├── create-registry-cli.sh        # CLI создание (справка)
│
├── AUTOMATION_CONSULTATION.md    # Консультация
├── AUTOMATION_SUMMARY.md         # Обзор автоматизации
├── README_AUTOMATION.md          # Полная инструкция
├── NEXT_STEPS.md                 # Следующие шаги
├── GITHUB_SECRETS_SETUP.md       # Настройка GitHub Secrets
├── CONTAINER_REGISTRY_SETUP_GUIDE.md # Инструкция Container Registry
├── QUICK_REGISTRY_SETUP.md       # Быстрая памятка
└── CLEANUP_PLAN.md               # Этот файл (удалить после очистки)
```

## ⚠️ **Важные предупреждения:**

1. **НЕ УДАЛЯТЬ:**
   - `.terraform/` директорию
   - `terraform.tfstate` файлы
   - `.terraform.lock.hcl`
   - `terraform.tfvars`

2. **Перед удалением:**
   - Убедитесь, что все важные данные сохранены
   - Проверьте, что новые файлы работают корректно

3. **После очистки:**
   - Проверьте работоспособность Terraform
   - Убедитесь, что все скрипты работают
   - Обновите документацию при необходимости

## 🎯 **Результат очистки:**

- **Упрощенная структура** директории
- **Удаление дублирующих** файлов
- **Сохранение только актуальных** конфигураций
- **Улучшенная читаемость** и навигация
- **Снижение путаницы** при работе с проектом
