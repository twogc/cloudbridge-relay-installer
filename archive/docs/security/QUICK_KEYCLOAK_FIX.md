# **БЫСТРОЕ ИСПРАВЛЕНИЕ KEYCLOAK**

## **ПРОБЛЕМА**
CloudBridge Relay не может подключиться к Keycloak на `auth.2gc.ru:8080`

## **БЫСТРОЕ РЕШЕНИЕ (5 минут)**

### **1. Установить переменные окружения**
```bash
export KEYCLOAK_SERVER_URL="https://auth.2gc.ru:8080"
export KEYCLOAK_REALM="cloudbridge"
export KEYCLOAK_CLIENT_ID="relay-client"
export KEYCLOAK_CLIENT_SECRET="your-secret-here"
export KEYCLOAK_ISSUER="https://auth.2gc.ru:8080/realms/cloudbridge"
export KEYCLOAK_AUDIENCE="relay-client"
```

### **2. Получить Public Key**
```bash
# Получить JWKS и извлечь public key
curl -s https://auth.2gc.ru:8080/realms/cloudbridge/protocol/openid-connect/certs | jq '.keys[0].x5c[0]' | tr -d '"' | base64 -d | openssl x509 -pubkey -noout
```

### **3. Установить Public Key**
```bash
export KEYCLOAK_PUBLIC_KEY="-----BEGIN PUBLIC KEY-----
[вставить полученный ключ]
-----END PUBLIC KEY-----"
```

### **4. Проверить подключение**
```bash
# Проверить доступность Keycloak
curl -v https://auth.2gc.ru:8080/realms/cloudbridge

# Запустить приложение
./cloudbridge-relay --config=configs/config.json --log-level=debug
```

##  **ДЛЯ MINIKUBE**

### **Обновить ConfigMap**
```bash
kubectl create configmap cloudbridge-keycloak-config \
  --from-literal=KEYCLOAK_SERVER_URL="https://auth.2gc.ru:8080" \
  --from-literal=KEYCLOAK_REALM="cloudbridge" \
  --from-literal=KEYCLOAK_CLIENT_ID="relay-client" \
  --from-literal=KEYCLOAK_CLIENT_SECRET="your-secret" \
  --from-literal=KEYCLOAK_ISSUER="https://auth.2gc.ru:8080/realms/cloudbridge" \
  --from-literal=KEYCLOAK_AUDIENCE="relay-client" \
  --from-literal=KEYCLOAK_PUBLIC_KEY="-----BEGIN PUBLIC KEY-----
[вставить ключ]
-----END PUBLIC KEY-----" \
  -n cloudbridge
```

### **Обновить deployment**
```yaml
# Добавить в deployment.yaml
envFrom:
- configMapRef:
    name: cloudbridge-keycloak-config
```

### **Перезапустить**
```bash
kubectl rollout restart deployment/cloudbridge-relay -n cloudbridge
kubectl logs -f deployment/cloudbridge-relay -n cloudbridge
```

## **ПРОВЕРКА**
- Логи должны показать: `"Keycloak client initialized successfully"`
- Нет ошибок: `"failed to parse public key"`
- Health endpoint отвечает: `curl http://localhost:9091/health`

---
**Время исправления**: 5-10 минут  
**Сложность**: 2/10  
**Статус**: КРИТИЧНО




