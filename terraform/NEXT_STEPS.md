# 🚀 Следующие шаги для автоматизации

## ✅ **Что уже готово:**

1. **Скрипты автоматизации** созданы и настроены
2. **Terraform конфигурации** подготовлены
3. **GitHub Actions workflow** настроен
4. **Конфигурационные файлы** созданы

## 🎯 **Что нужно сделать СЕЙЧАС:**

### **Шаг 1: Создание Container Registry в VK Cloud**

1. **Войдите в VK Cloud Console:** https://mcs.mail.ru/
2. **Перейдите в раздел "Container Registry"**
3. **Нажмите "Create Registry"**
4. **Заполните форму:**
   - **Name:** `cloudbridge-relay`
   - **Description:** `CloudBridge Relay Docker Registry`
   - **Region:** `MS1` (Москва)
5. **Нажмите "Create"**

### **Шаг 2: Настройка аутентификации**

1. **В созданном registry перейдите в "Settings"**
2. **В разделе "Access" создайте новый токен:**
   - **Username:** `cloudbridge-user`
   - **Password:** [сгенерированный пароль]
3. **Скопируйте username и password**

### **Шаг 3: Обновление .env файла**

```bash
# Отредактируйте файл ../.env
nano ../.env

# Замените следующие строки:
VK_REGISTRY_USERNAME=cloudbridge-user
VK_REGISTRY_PASSWORD=your-actual-password
```

### **Шаг 4: Тестирование Container Registry**

```bash
# Загрузите переменные окружения
source ../load-env.sh

# Протестируйте подключение
./test-registry-connection.sh
```

### **Шаг 5: Настройка GitHub Secrets**

1. **Перейдите в ваш GitHub репозиторий**
2. **Нажмите Settings → Secrets and variables → Actions**
3. **Нажмите "New repository secret"**
4. **Добавьте секреты:**

```
VK_REGISTRY_USERNAME = cloudbridge-user
VK_REGISTRY_PASSWORD = [ваш пароль]
VK_CLOUD_OPENRC = [содержимое openrc файла]
```

### **Шаг 6: Автоматическое развертывание**

```bash
# Запустите полное автоматизированное развертывание
./deploy-automated.sh
```

## 🔍 **Проверка готовности:**

### **Перед запуском автоматизации убедитесь:**

- ✅ **Container Registry создан** в VK Cloud
- ✅ **Аутентификация настроена** (username/password)
- ✅ **.env файл обновлен** с правильными данными
- ✅ **GitHub Secrets настроены**
- ✅ **SSH ключ** `~/.ssh/cloudbridge-relay` доступен

## 🚨 **Если что-то не работает:**

### **Ошибка Container Registry:**
```bash
# Проверьте подключение
docker login registry.vkcs.ru
# Username: cloudbridge-user
# Password: [ваш пароль]
```

### **Ошибка переменных окружения:**
```bash
# Проверьте .env файл
cat ../.env

# Загрузите переменные
source ../load-env.sh

# Проверьте переменные
echo $VK_REGISTRY_USERNAME
echo $VK_REGISTRY_PASSWORD
```

### **Ошибка SSH:**
```bash
# Проверьте SSH ключ
ls -la ~/.ssh/cloudbridge-relay

# Проверьте права доступа
chmod 600 ~/.ssh/cloudbridge-relay
```

## 📞 **Нужна помощь?**

### **Документация:**
- [VK Cloud Container Registry](https://mcs.mail.ru/docs/container-registry/)
- [GitHub Actions](https://docs.github.com/en/actions)
- [Terraform VK Cloud](https://registry.terraform.io/providers/vk-cs/vkcs/latest/docs)

### **Поддержка:**
- **VK Cloud:** support@vk.com
- **GitHub:** Issues в репозитории

---

## 🎉 **После успешной настройки:**

Вы получите:
- **Полностью автоматизированное развертывание** CloudBridge Relay серверов
- **CI/CD pipeline** для автоматического обновления
- **Масштабируемость** - легко добавлять новые серверы
- **Мониторинг** и управление через скрипты

**🚀 Готово к автоматизации!**
