# 🚀 **Быстрая памятка: Container Registry**

## **Что нужно сделать:**

### **1. Войти в VK Cloud Console**
- **URL:** https://mcs.mail.ru/
- **Email:** info@2gc.ru
- **Проект:** 8530029fa7b64199b2ad87c18d4620df

### **2. Создать Container Registry**
- **Раздел:** Container Registry (левое меню)
- **Кнопка:** Create Registry
- **Параметры:**
  ```
  Name: cloudbridge-relay
  Description: CloudBridge Relay Docker Registry
  Region: MS1 (Москва)
  ```

### **3. Настроить аутентификацию**
- **Settings → Access → Create Token**
- **Username:** cloudbridge-user
- **Password:** [Generate] - скопировать пароль

### **4. Обновить .env файл**
```bash
nano ../.env

# Заменить:
VK_REGISTRY_USERNAME=cloudbridge-user
VK_REGISTRY_PASSWORD=your-actual-password
```

### **5. Протестировать**
```bash
source ../load-env.sh
./test-registry-connection.sh
```

## 🔑 **Ключевые данные:**
- **Registry URL:** registry.vkcs.ru
- **Full Image Path:** registry.vkcs.ru/cloudbridge-relay/cloudbridge-relay:latest
- **Username:** cloudbridge-user
- **Password:** [сгенерированный пароль]

## ⚡ **Быстрая команда для тестирования:**
```bash
echo "your-password" | docker login registry.vkcs.ru -u "cloudbridge-user" --password-stdin
```

**🚀 Готово! Теперь можно запускать автоматизацию.**
