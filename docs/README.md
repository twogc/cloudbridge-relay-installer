# VK Cloud API Manager - Балансировщик и Anycast IP

Набор скриптов для управления балансировщиком нагрузки и Anycast IP через API VK Cloud.

## 📋 Содержание

- [Установка](#установка)
- [Настройка](#настройка)
- [Использование](#использование)
- [Скрипты](#скрипты)
- [Примеры использования](#примеры-использования)
- [API Endpoints](#api-endpoints)
- [Документация](#документация)

## 🚀 Установка

### Требования

- Bash
- curl
- jq (для работы с JSON)

### Установка зависимостей

```bash
# Ubuntu/Debian
sudo apt-get update
sudo apt-get install curl jq

# CentOS/RHEL
sudo yum install curl jq

# macOS
brew install curl jq
```

### Клонирование репозитория

```bash
git clone <repository-url>
cd vk-cloud-api-manager
chmod +x *.sh
```

## ⚙️ Настройка

### 1. Настройка переменных окружения

Отредактируйте файл `setup-env.sh` с вашими данными:

```bash
# Настройка переменных окружения для VK Cloud
export OS_AUTH_URL="https://infra.mail.ru:35357/v3/"
export OS_PROJECT_ID="ваш_project_id"
export OS_REGION_NAME="RegionOne"
export OS_USERNAME="ваш_email"
export OS_USER_DOMAIN_NAME="users"
export OS_INTERFACE=public
export OS_IDENTITY_API_VERSION=3

# Токен для API (опционально)
export VK_TOKEN="ваш_токен"
```

### 2. Получение токена доступа

```bash
./vk-cloud-manager.sh
# Выберите "1. Получить токен"
# Введите пароль от вашей учетной записи
```

## 📖 Использование

### Основной скрипт

```bash
./vk-cloud-manager.sh
```

### Отдельные скрипты

```bash
# Только балансировщик нагрузки
./loadbalancer-api.sh

# Только Anycast IP
./anycast-api.sh
```

## 🔧 Скрипты

### 1. `setup-env.sh`
Настройка переменных окружения для работы с VK Cloud API.

### 2. `vk-cloud-manager.sh` (основной)
Универсальный скрипт для управления всеми сервисами:
- Балансировщик нагрузки
- Anycast IP
- Сети и ресурсы
- Автоматическая настройка

### 3. `loadbalancer-api.sh`
Специализированный скрипт для управления балансировщиком нагрузки:
- Создание балансировщика
- Управление прослушивателями
- Управление пулами
- Добавление участников

### 4. `anycast-api.sh`
Специализированный скрипт для управления Anycast IP:
- Создание Anycast IP
- Привязка IP адресов
- Настройка healthcheck
- Управление правилами

## 💡 Примеры использования

### Создание балансировщика с Anycast IP

```bash
# 1. Запустите основной скрипт
./vk-cloud-manager.sh

# 2. Получите токен (опция 1)

# 3. Просмотрите доступные подсети (опция 3)

# 4. Создайте Anycast IP (опция 12)
# Название: cloudbridge-relay-anycast
# Описание: Anycast IP для CloudBridge Relay

# 5. Создайте балансировщик (опция 7)
# Название: cloudbridge-lb
# ID подсети: (из шага 3)
# Зона доступности: GZ1

# 6. Привяжите IP к Anycast (опция 13)
# ID Anycast IP: (из шага 4)
# ID порта: (из списка портов)
```

### Автоматическая настройка

```bash
# Полная автоматическая настройка CloudBridge Relay
./vk-cloud-manager.sh
# Выберите "15. Настроить CloudBridge Relay (полная настройка)"
```

### Проверка статуса

```bash
# Просмотр балансировщиков
./vk-cloud-manager.sh
# Выберите "6. Показать балансировщики"

# Просмотр Anycast IP
./vk-cloud-manager.sh
# Выберите "11. Показать Anycast IP"
```

## 🌐 API Endpoints

### Балансировщик нагрузки (Octavia)
- **URL**: `https://public.infra.mail.ru:9876`
- **Версия**: v2.0
- **Документация**: [Управление балансировщиками нагрузки](https://mcs.mail.ru/docs/ru/networks/balancing/instructions/manage-lb)

### Anycast IP (Neutron)
- **URL**: `https://infra.mail.ru:9696`
- **Версия**: v2.0
- **Документация**: [Управление Anycast IP](https://mcs.mail.ru/docs/ru/networks/vnet/instructions/ip/anycast-ip)

### Аутентификация (Keystone)
- **URL**: `https://infra.mail.ru:35357/v3/`
- **Метод**: Password-based authentication

## 📚 Документация

### Балансировщик нагрузки

**Основные операции:**
- Создание балансировщика
- Управление прослушивателями (listeners)
- Управление пулами (pools)
- Добавление участников (members)

**Методы балансировки:**
- `ROUND_ROBIN` - круговой перебор
- `LEAST_CONNECTIONS` - наименьшее количество соединений
- `SOURCE_IP` - закрепление по IP-адресу клиента

### Anycast IP

**Основные операции:**
- Создание Anycast IP
- Привязка IP адресов
- Настройка healthcheck
- Управление правилами

**Ограничения:**
- Работает только в SDN Sprut
- До 8 IP адресов на один Anycast IP
- Все IP адреса должны быть одного типа

**Healthcheck:**
- Протоколы: TCP, ICMP
- IP для проверки: `169.254.169.100/32`
- Необходимо разрешить в security groups

## 🔒 Безопасность

### Рекомендации

1. **Хранение токенов**: Не храните токены в открытом виде
2. **Пароли**: Используйте переменные окружения для паролей
3. **Security Groups**: Настройте правильные правила доступа
4. **Healthcheck**: Обязательно настройте healthcheck для Anycast IP

### Переменные окружения

```bash
# Безопасное хранение пароля
export OS_PASSWORD="ваш_пароль"

# Или использование файла
source .env
```

## 🐛 Устранение неполадок

### Частые ошибки

1. **"jq не установлен"**
   ```bash
   sudo apt-get install jq
   ```

2. **"Ошибка аутентификации"**
   - Проверьте правильность учетных данных
   - Убедитесь, что доступ по API активирован

3. **"Anycast IP не работает"**
   - Проверьте, что используется SDN Sprut
   - Убедитесь, что healthcheck настроен правильно

4. **"Балансировщик не создается"**
   - Проверьте ID подсети
   - Убедитесь, что подсеть находится в правильной зоне доступности

### Логи

Все операции логируются в консоль. Для сохранения логов:

```bash
./vk-cloud-manager.sh 2>&1 | tee vk-cloud-manager.log
```

## 📞 Поддержка

- **Документация VK Cloud**: [docs.vk.com](https://docs.vk.com)
- **API Reference**: [api.vk.com](https://api.vk.com)
- **Поддержка**: [support.vk.com](https://support.vk.com)

## 📄 Лицензия

MIT License

## 🤝 Вклад в проект

1. Fork репозитория
2. Создайте ветку для новой функции
3. Внесите изменения
4. Создайте Pull Request

---

**Примечание**: Этот набор скриптов основан на официальной документации VK Cloud и предназначен для автоматизации управления облачной инфраструктурой.

