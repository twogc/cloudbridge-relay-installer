#  **Пошаговая инструкция: Создание Container Registry в VK Cloud Console**

## **Обзор**

Container Registry в VK Cloud - это приватный реестр Docker образов, который позволяет хранить и управлять вашими контейнерными образами. Это необходимо для автоматизации развертывания CloudBridge Relay серверов.

## **Шаг 1: Вход в VK Cloud Console**

### **1.1 Откройте браузер**
- Перейдите по адресу: **https://mcs.mail.ru/**
- Убедитесь, что используете современный браузер (Chrome, Firefox, Safari, Edge)

### **1.2 Войдите в аккаунт**
- Введите ваш **email**: `info@2gc.ru`
- Введите ваш **пароль**
- Нажмите **"Войти"**

### **1.3 Выберите проект**
- Если у вас несколько проектов, выберите: **`8530029fa7b64199b2ad87c18d4620df`**
- Убедитесь, что проект активен (должен быть зеленый индикатор)

## **Шаг 2: Поиск раздела Container Registry**

### **2.1 Навигация по меню**
- В **левой панели** найдите раздел **"Container Registry"**
- Обычно он находится в группе **"Сервисы"** или **"Инфраструктура"**

### **2.2 Если раздел не найден**
- Проверьте, что у вас есть права доступа к Container Registry
- Обратитесь в поддержку VK Cloud: **support@vk.com**
- Или попробуйте найти в разделе **"Все сервисы"**

## 🏗️ **Шаг 3: Создание Container Registry**

### **3.1 Нажмите "Create Registry"**
- В правом верхнем углу найдите кнопку **"Create Registry"**
- Нажмите на неё

### **3.2 Заполните форму создания**

#### **Основные параметры:**
```
Name: cloudbridge-relay
Description: CloudBridge Relay Docker Registry
Region: MS1 (Москва)
```

#### **Дополнительные настройки (если доступны):**
- **Storage Class:** `ceph-ssd` (рекомендуется)
- **Quota:** `10 GB` (базовый план)
- **Access Control:** `Private` (приватный доступ)

### **3.3 Создание**
- Проверьте все параметры
- Нажмите **"Create"** или **"Создать"**
- Дождитесь завершения создания (обычно 1-2 минуты)

## **Шаг 4: Настройка аутентификации**

### **4.1 Переход в настройки**
- После создания нажмите на созданный registry **`cloudbridge-relay`**
- В верхнем меню найдите **"Settings"** или **"Настройки"**

### **4.2 Создание токена доступа**
- В разделе **"Access"** или **"Доступ"** найдите **"Create Token"**
- Нажмите **"Create Token"** или **"Создать токен"**

### **4.3 Настройка токена**
```
Username: cloudbridge-user
Password: [нажмите "Generate" для автоматической генерации]
Description: CloudBridge Relay automation token
```

### **4.4 Сохранение данных**
- **Скопируйте username и password**
- **Сохраните их в безопасном месте**
- Нажмите **"Create"** или **"Создать"**

## 📝 **Шаг 5: Получение информации о registry**

### **5.1 Registry URL**
- В настройках найдите **"Registry URL"**
- Обычно это: **`registry.vkcs.ru`**
- Скопируйте этот URL

### **5.2 Полный путь к образу**
- Для вашего проекта полный путь будет:
```
registry.vkcs.ru/cloudbridge-relay/cloudbridge-relay:latest
```

##  **Шаг 6: Обновление .env файла**

### **6.1 Отредактируйте .env файл**
```bash
nano ../.env
```

### **6.2 Обновите переменные**
```bash
# VK Cloud Container Registry
VK_REGISTRY_USERNAME=cloudbridge-user
VK_REGISTRY_PASSWORD=your-actual-password

# Docker образ
DOCKER_IMAGE_NAME=cloudbridge-relay
DOCKER_IMAGE_TAG=latest
```

### **6.3 Сохраните файл**
- Нажмите **Ctrl+X** для выхода
- Нажмите **Y** для сохранения
- Нажмите **Enter** для подтверждения имени файла

## **Шаг 7: Тестирование подключения**

### **7.1 Загрузите переменные окружения**
```bash
source ../load-env.sh
```

### **7.2 Протестируйте подключение**
```bash
./test-registry-connection.sh
```

### **7.3 Ручное тестирование**
```bash
# Тест Docker login
echo "your-password" | docker login registry.vkcs.ru -u "cloudbridge-user" --password-stdin

# Тест создания тестового образа
docker pull alpine:latest
docker tag alpine:latest registry.vkcs.ru/cloudbridge-relay/test:latest
docker push registry.vkcs.ru/cloudbridge-relay/test:latest
```

## **Возможные проблемы и решения**

### **Проблема 1: Раздел Container Registry не найден**
**Решение:**
- Проверьте права доступа к проекту
- Обратитесь в поддержку VK Cloud
- Убедитесь, что сервис доступен в вашем регионе

### **Проблема 2: Ошибка создания registry**
**Решение:**
- Проверьте лимиты проекта
- Убедитесь, что имя уникально
- Попробуйте другое имя (например, `cloudbridge-relay-$(date +%s)`)

### **Проблема 3: Ошибка аутентификации**
**Решение:**
- Проверьте правильность username/password
- Убедитесь, что токен активен
- Попробуйте создать новый токен

### **Проблема 4: Ошибка Docker push**
**Решение:**
- Проверьте подключение к registry
- Убедитесь, что образ правильно тегирован
- Проверьте права доступа к проекту

## 📞 **Поддержка**

### **VK Cloud поддержка:**
- **Email:** support@vk.com
- **Telegram:** @vkcloud_support
- **Документация:** https://mcs.mail.ru/docs/

### **Полезные ссылки:**
- [VK Cloud Container Registry](https://mcs.mail.ru/docs/container-registry/)
- [Docker Registry документация](https://docs.docker.com/registry/)
- [OpenStack документация](https://docs.openstack.org/)

## **Следующие шаги**

После успешного создания Container Registry:

1. **Настройте GitHub Secrets** (см. `GITHUB_SECRETS_SETUP.md`)
2. **Протестируйте автоматизацию** (см. `test-registry-connection.sh`)
3. **Запустите полное развертывание** (см. `deploy-automated.sh`)

---

## **Чек-лист готовности**

- [ ] Container Registry создан в VK Cloud Console
- [ ] Настроена аутентификация (username/password)
- [ ] Обновлен .env файл
- [ ] Протестировано подключение
- [ ] Готово к автоматизации

**Container Registry готов к использованию!**
