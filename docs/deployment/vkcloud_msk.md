# Автоматизация развёртывания CloudBridge Relay в VK Cloud (Москва)

## 1. Описание

Этот гайд описывает, как быстро развернуть CloudBridge Relay в московском регионе VK Cloud с помощью Terraform и Ansible.

---

## 2. Структура каталогов

```
infra/                # Terraform-скрипты
  main.tf
  variables.tf
  regions/
    msk.tfvars
ansible/              # Ansible playbook и inventory
  playbook.yml
  inventory/
    msk.ini
```

---

## 3. Шаги развертывания

### 3.1. Подготовьте переменные для региона

`infra/regions/msk.tfvars`:
```hcl
vkcs_region = "msk"
image_name = "ubuntu-22.04"
network_id = "<your_network_id>"
floating_ip_pool = "external"
ssh_key_name = "<your_ssh_key_name>"
flavor_name = "s2.medium"
vkcs_username = "<your_username>"
vkcs_password = "<your_password>"
vkcs_domain = "<your_domain>"
vkcs_project = "<your_project>"
```

### 3.2. Разверните ВМ через Terraform

```bash
cd infra
terraform init
terraform apply -var-file=regions/msk.tfvars
```

- После выполнения получите публичный IP из вывода Terraform (output `relay_public_ip`).

### 3.3. Подготовьте Ansible inventory

`ansible/inventory/msk.ini`:
```
[relay]
<public_ip_from_terraform> ansible_user=ubuntu
```

### 3.4. Запустите Ansible для настройки и деплоя

```bash
cd ansible
ansible-playbook -i inventory/msk.ini playbook.yml
```

- Будет установлен Docker, скопирован docker-compose.yml и запущен CloudBridge Relay.

---

## 4. Важно
- Все контейнеры будут работать в московском времени (TZ=Europe/Moscow).
- Для production рекомендуется настроить firewall, резервное копирование и мониторинг.
- Все параметры (логины, пароли, ключи) храните в защищённом виде.

---

## 5. Полезные ссылки
- [Terraform VK Cloud Provider](https://registry.terraform.io/providers/vk-cs/vkcs/latest/docs)
- [Документация VK Cloud по Terraform](https://docs.vk.cloud/ru/terraform/)
- [Ansible Getting Started](https://docs.ansible.com/ansible/latest/user_guide/intro_getting_started.html)

---

## 6. Пример пайплайна

```bash
# 1. Создать инфраструктуру
cd infra
terraform apply -var-file=regions/msk.tfvars

# 2. Получить IP и добавить в ansible/inventory/msk.ini

# 3. Настроить сервер
cd ../ansible
ansible-playbook -i inventory/msk.ini playbook.yml
```

---

## 7. FAQ
- **Q:** Как добавить новый регион?  
  **A:** Создайте новый tfvars и inventory для региона, повторите шаги.
- **Q:** Как обновить docker-compose.yml?  
  **A:** Просто перезапустите playbook Ansible. 