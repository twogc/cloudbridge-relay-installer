# Настройка GitHub Secrets для автоматизации

## **Необходимые GitHub Secrets**

Для работы автоматизации через GitHub Actions необходимо настроить следующие секреты:

### **1. VK Cloud Container Registry Secrets**

#### **VK_REGISTRY_USERNAME**
- **Описание:** Имя пользователя для доступа к VK Cloud Container Registry
- **Как получить:**
  1. Войдите в VK Cloud Console
  2. Перейдите в раздел "Container Registry"
  3. Создайте новый registry или используйте существующий
  4. В настройках registry найдите "Username"

#### **VK_REGISTRY_PASSWORD**
- **Описание:** Пароль для доступа к VK Cloud Container Registry
- **Как получить:**
  1. В VK Cloud Console → Container Registry
  2. Создайте новый токен доступа
  3. Скопируйте сгенерированный пароль

### **2. VK Cloud OpenStack Secrets**

#### **VK_CLOUD_OPENRC**
- **Описание:** Содержимое OpenRC файла для аутентификации в VK Cloud
- **Как получить:**
  1. В VK Cloud Console → "API Access"
  2. Скачайте OpenRC файл
  3. Скопируйте содержимое файла

## **Пошаговая настройка**

### **Шаг 1: Создание Container Registry в VK Cloud**

```bash
# 1. Войдите в VK Cloud Console
# 2. Перейдите в "Container Registry"
# 3. Нажмите "Create Registry"
# 4. Заполните форму:
#    - Name: cloudbridge-relay
#    - Description: CloudBridge Relay Docker Registry
#    - Region: MS1 (Москва)
# 5. Нажмите "Create"
```

### **Шаг 2: Настройка аутентификации**

```bash
# 1. В созданном registry перейдите в "Settings"
# 2. В разделе "Access" создайте новый токен:
#    - Username: cloudbridge-user
#    - Password: [сгенерированный пароль]
# 3. Скопируйте username и password
```

### **Шаг 3: Настройка GitHub Secrets**

1. **Перейдите в ваш GitHub репозиторий**
2. **Нажмите Settings → Secrets and variables → Actions**
3. **Нажмите "New repository secret"**
4. **Добавьте каждый секрет:**

```
VK_REGISTRY_USERNAME = cloudbridge-user
VK_REGISTRY_PASSWORD = [сгенерированный пароль]
VK_CLOUD_OPENRC = [содержимое openrc файла]
```

### **Шаг 4: Проверка настройки**

```bash
# Тест подключения к registry
docker login registry.vkcs.ru
# Username: cloudbridge-user
# Password: [ваш пароль]
```

##  **Альтернативные способы аутентификации**

### **Вариант 1: Service Account (Рекомендуемый)**

```bash
# 1. Создайте Service Account в VK Cloud
# 2. Назначьте роли: Container Registry User, Compute User
# 3. Создайте API ключ
# 4. Используйте API ключ вместо username/password
```

### **Вариант 2: IAM Token**

```bash
# 1. Создайте IAM пользователя
# 2. Назначьте необходимые роли
# 3. Сгенерируйте IAM токен
# 4. Используйте токен для аутентификации
```

## 📝 **Пример OpenRC файла**

```bash
#!/bin/bash
export OS_AUTH_URL=https://infra.mail.ru:5000/v3
export OS_PROJECT_ID=your-project-id
export OS_PROJECT_NAME=your-project-name
export OS_USER_DOMAIN_NAME=your-domain
export OS_PROJECT_DOMAIN_NAME=your-domain
export OS_USERNAME=your-username
export OS_PASSWORD=your-password
export OS_REGION_NAME=MS1
export OS_INTERFACE=public
export OS_IDENTITY_API_VERSION=3
```

## **Безопасность**

### **Рекомендации:**
- **Используйте Service Account** вместо личных учетных данных
- **Ограничьте права доступа** минимально необходимыми
- **Регулярно обновляйте** токены и пароли
- **Мониторьте использование** API ключей
- **Используйте разные учетные данные** для разных окружений

### **Запрещено:**
- **Хранить секреты в коде**
- **Использовать root учетные данные**
- **Передавать секреты в логи**
- **Использовать один токен для всех сервисов**

## **Тестирование настройки**

### **Локальное тестирование:**

```bash
# 1. Создайте .env файл с секретами
cat > .env << EOF
VK_REGISTRY_USERNAME=your-username
VK_REGISTRY_PASSWORD=your-password
VK_CLOUD_OPENRC=$(cat openrc.sh)
EOF

# 2. Тест подключения к registry
docker login registry.vkcs.ru

# 3. Тест сборки и загрузки образа
docker build -f Dockerfile.relay -t registry.vkcs.ru/your-project/cloudbridge-relay:test .
docker push registry.vkcs.ru/your-project/cloudbridge-relay:test
```

### **Тестирование GitHub Actions:**

```bash
# 1. Создайте тестовую ветку
git checkout -b test-automation

# 2. Внесите небольшое изменение в Dockerfile
echo "# Test comment" >> Dockerfile.relay

# 3. Запушьте изменения
git add Dockerfile.relay
git commit -m "Test automation"
git push origin test-automation

# 4. Проверьте выполнение workflow в GitHub Actions
```

## **Устранение неполадок**

### **Ошибка аутентификации:**
```bash
# Проверьте правильность username/password
# Убедитесь, что registry доступен
# Проверьте права доступа пользователя
```

### **Ошибка доступа к VK Cloud:**
```bash
# Проверьте содержимое VK_CLOUD_OPENRC
# Убедитесь, что проект активен
# Проверьте лимиты и квоты
```

### **Ошибка сборки Docker образа:**
```bash
# Проверьте Dockerfile.relay
# Убедитесь, что все зависимости доступны
# Проверьте права доступа к registry
```

## 📞 **Поддержка**

### **VK Cloud:**
- **Документация:** https://mcs.mail.ru/docs/
- **Поддержка:** support@vk.com
- **Telegram:** @vkcloud_support

### **GitHub Actions:**
- **Документация:** https://docs.github.com/en/actions
- **Marketplace:** https://github.com/marketplace?type=actions
- **Community:** https://github.community/
