# **ОТЧЕТ О ПРОБЛЕМЕ ПОДКЛЮЧЕНИЯ К KEYCLOAK**

## **Краткое описание проблемы**

**Проблема**: CloudBridge Relay не может подключиться к Keycloak серверу, расположенному на внутреннем IP адресе `auth.2gc.ru`.

**Ошибка**: `Failed to create server: failed to initialize auth adapter: failed to create Keycloak client: failed to parse public key: failed to decode PEM block containing public key`

**Статус**: 🔴 **КРИТИЧЕСКАЯ ПРОБЛЕМА** - блокирует запуск основного приложения

---

## 🔍 **Детальный анализ проблемы**

### **1. Текущая конфигурация Keycloak**

В проекте найдено **несколько конфликтующих конфигураций**:

#### **Основная конфигурация** (`configs/config.json`):
```json
{
  "keycloak": {
    "enabled": true,
    "server_url": "${KEYCLOAK_SERVER_URL}",
    "realm": "${KEYCLOAK_REALM}",
    "client_id": "${KEYCLOAK_CLIENT_ID}",
    "client_secret": "${KEYCLOAK_CLIENT_SECRET}",
    "public_key": "${KEYCLOAK_PUBLIC_KEY}",
    "token_validation": {
      "issuer": "${KEYCLOAK_ISSUER}",
      "audience": "${KEYCLOAK_AUDIENCE}",
      "clock_skew": "30s"
    },
    "http_timeout": "30s"
  }
}
```

#### **Альтернативные конфигурации**:
- `configs/keycloak_config.json`: `"server_url": "https://auth.2gc.uk"`
- `config/keycloak_config.json`: `"server_url": "http://keycloak:8080"`
- `configs/keycloak_edge_config.yaml`: `"server_url": "https://auth.2gc.ru"`

### **2. Проблемы с переменными окружения**

**Основная проблема**: Переменная `KEYCLOAK_SERVER_URL` не установлена или содержит неправильное значение.

**Текущие значения в коде**:
```bash
# В supervisord.conf
KEYCLOAK_SERVER_URL="https://auth.2gc.ru"

# В setup_moscow_region.sh  
KEYCLOAK_SERVER_URL=http://localhost:8080
```

**Требуемое значение**: `KEYCLOAK_SERVER_URL=https://auth.2gc.ru:8080`

### **3. Проблемы с Public Key**

**Ошибка парсинга PEM**: Указывает на проблемы с форматом или содержимым `KEYCLOAK_PUBLIC_KEY`.

**Возможные причины**:
- Неправильный формат PEM ключа
- Отсутствие переменной окружения
- Неправильное экранирование символов
- Проблемы с кодировкой

---

## 🛠️ **РЕКОМЕНДАЦИИ ДЛЯ РАЗРАБОТЧИКА**

### **1. Немедленные действия (КРИТИЧНО)**

#### **A. Установить правильные переменные окружения**
```bash
export KEYCLOAK_SERVER_URL="https://auth.2gc.ru:8080"
export KEYCLOAK_REALM="cloudbridge"
export KEYCLOAK_CLIENT_ID="relay-client"
export KEYCLOAK_CLIENT_SECRET="your-secret-here"
export KEYCLOAK_PUBLIC_KEY="-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA...
-----END PUBLIC KEY-----"
export KEYCLOAK_ISSUER="https://auth.2gc.ru:8080/realms/cloudbridge"
export KEYCLOAK_AUDIENCE="relay-client"
```

#### **B. Проверить доступность Keycloak сервера**
```bash
# Проверка доступности
curl -v https://auth.2gc.ru:8080/realms/cloudbridge

# Проверка JWKS endpoint
curl -v https://auth.2gc.ru:8080/realms/cloudbridge/protocol/openid-connect/certs
```

#### **C. Получить правильный Public Key**
```bash
# Получить JWKS и извлечь public key
curl -s https://auth.2gc.ru:8080/realms/cloudbridge/protocol/openid-connect/certs | jq '.keys[0]'
```

### **2. Исправления в коде**

#### **A. Улучшить обработку ошибок в `config/config.go`**
```go
// Добавить более детальную диагностику
func (c *Config) validateKeycloakConfig() error {
    if c.Keycloak.Enabled {
        if c.Keycloak.ServerURL == "" {
            return fmt.Errorf("KEYCLOAK_SERVER_URL is not set")
        }
        if c.Keycloak.PublicKey == "" {
            return fmt.Errorf("KEYCLOAK_PUBLIC_KEY is not set")
        }
        // Проверить формат PEM
        if !strings.Contains(c.Keycloak.PublicKey, "BEGIN PUBLIC KEY") {
            return fmt.Errorf("invalid PEM format in KEYCLOAK_PUBLIC_KEY")
        }
    }
    return nil
}
```

#### **B. Добавить fallback для Keycloak**
```go
// В случае недоступности Keycloak, использовать fallback режим
if !keycloakAvailable {
    log.Warn("Keycloak unavailable, using fallback authentication")
    // Использовать fallback_secret для локальной аутентификации
}
```

### **3. Сетевая диагностика**

#### **A. Проверить сетевую доступность**
```bash
# Проверка ping
ping -c 3 auth.2gc.ru

# Проверка порта
telnet auth.2gc.ru 8080
# или
nc -zv auth.2gc.ru 8080
```

#### **B. Проверить firewall и routing**
```bash
# Проверить маршруты
ip route get auth.2gc.ru

# Проверить iptables
iptables -L -n | grep auth.2gc.ru
```

### **4. Конфигурация для Minikube**

#### **A. Обновить ConfigMap**
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cloudbridge-config
  namespace: cloudbridge
data:
  KEYCLOAK_SERVER_URL: "https://auth.2gc.ru:8080"
  KEYCLOAK_REALM: "cloudbridge"
  KEYCLOAK_CLIENT_ID: "relay-client"
  KEYCLOAK_CLIENT_SECRET: "your-secret"
  KEYCLOAK_PUBLIC_KEY: |
    -----BEGIN PUBLIC KEY-----
    MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA...
    -----END PUBLIC KEY-----
  KEYCLOAK_ISSUER: "https://auth.2gc.ru:8080/realms/cloudbridge"
  KEYCLOAK_AUDIENCE: "relay-client"
```

#### **B. Обновить deployment.yaml**
```yaml
spec:
  template:
    spec:
      containers:
      - name: cloudbridge-relay
        envFrom:
        - configMapRef:
            name: cloudbridge-config
```

---

## 🧪 **ПЛАН ТЕСТИРОВАНИЯ**

### **1. Локальное тестирование**
```bash
# 1. Установить переменные окружения
export KEYCLOAK_SERVER_URL="https://auth.2gc.ru:8080"
# ... остальные переменные

# 2. Запустить приложение
./cloudbridge-relay --config=configs/config.json --log-level=debug

# 3. Проверить логи на ошибки Keycloak
```

### **2. Тестирование в Minikube**
```bash
# 1. Обновить ConfigMap
kubectl apply -f k8s/configmap.yaml

# 2. Перезапустить deployment
kubectl rollout restart deployment/cloudbridge-relay -n cloudbridge

# 3. Проверить логи
kubectl logs -f deployment/cloudbridge-relay -n cloudbridge
```

### **3. Тестирование API endpoints**
```bash
# Проверить health endpoint
curl http://localhost:9091/health

# Проверить Keycloak статус (если есть endpoint)
curl http://localhost:9091/api/v1/auth/status
```

---

## 📊 **ПРИОРИТЕТЫ ИСПРАВЛЕНИЯ**

| Приоритет | Задача | Время | Сложность |
|-----------|--------|-------|-----------|
| 🔴 **КРИТИЧНО** | Установить `KEYCLOAK_SERVER_URL=https://auth.2gc.ru:8080` | 5 мин | 1/10 |
| 🔴 **КРИТИЧНО** | Получить и установить правильный `KEYCLOAK_PUBLIC_KEY` | 15 мин | 2/10 |
| 🟡 **ВАЖНО** | Улучшить обработку ошибок Keycloak | 30 мин | 4/10 |
| 🟡 **ВАЖНО** | Добавить fallback режим | 45 мин | 5/10 |
| 🟢 **ЖЕЛАТЕЛЬНО** | Унифицировать конфигурации Keycloak | 60 мин | 6/10 |

---

##  **ПОЛЕЗНЫЕ ССЫЛКИ**

- **Keycloak Admin Console**: https://auth.2gc.ru:8080/admin
- **Realm Settings**: https://auth.2gc.ru:8080/realms/cloudbridge
- **JWKS Endpoint**: https://auth.2gc.ru:8080/realms/cloudbridge/protocol/openid-connect/certs
- **Client Settings**: https://auth.2gc.ru:8080/admin/realms/cloudbridge/clients

---

## 📞 **КОНТАКТЫ**

**Ответственный за Keycloak сервер**: [Указать контакты администратора Keycloak]
**Разработчик**: [Указать контакты разработчика]

**Дата создания отчета**: 21.09.2025
**Статус**: 🔴 ТРЕБУЕТ НЕМЕДЛЕННОГО ВНИМАНИЯ




