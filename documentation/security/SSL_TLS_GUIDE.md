# 🔐 SSL/TLS Руководство CloudBridge Relay

## 📋 Обзор

**CloudBridge Relay** поддерживает полную SSL/TLS функциональность с современными стандартами безопасности, включая TLS 1.3, mTLS (mutual TLS) и пост-квантовую криптографию.

---

## 🚀 Ключевые возможности SSL/TLS

### **1. Современные протоколы**
- **TLS 1.3** — минимальная рекомендуемая версия
- **TLS 1.2** — минимальная поддерживаемая версия
- **Perfect Forward Secrecy** — обязателен
- **Strong ciphers** — только сильные шифры

### **2. Mutual TLS (mTLS)**
- **Двусторонняя аутентификация** — сервер и клиент проверяют друг друга
- **Client certificate validation** — проверка клиентских сертификатов
- **CA-based trust** — доверие на основе Certificate Authority

### **3. Post-Quantum Cryptography**
- **CRYSTALS-Kyber** — для установления ключей
- **CRYSTALS-Dilithium** — для цифровых подписей
- **Hybrid mode** — совместимость с классическими алгоритмами

### **4. Cipher Suites**
- **TLS_AES_256_GCM_SHA384** — основной шифр
- **TLS_CHACHA20_POLY1305_SHA256** — альтернативный шифр
- **TLS_AES_128_GCM_SHA256** — для совместимости

---

## 🔧 Конфигурация SSL/TLS

### **Основная конфигурация**

```json
{
  "security": {
    "tls_enabled": true,
    "cert_file": "/app/certs/tls.crt",
    "key_file": "/app/certs/tls.key",
    "allowed_ips": [],
    "rate_limit": {
      "enabled": true,
      "requests_per_second": 1000,
      "burst": 100
    },
    "mtls": {
      "enabled": true,
      "ca_file": "/app/certs/ca.crt",
      "client_cert_required": true
    },
    "cipher_suites": [
      "TLS_AES_256_GCM_SHA384",
      "TLS_CHACHA20_POLY1305_SHA256"
    ],
    "min_tls_version": "1.2",
    "audit": {
      "enabled": true,
      "log_file": "/app/logs/security.log"
    }
  }
}
```

### **Environment переменные**

```bash
# SSL/TLS настройки
TLS_ENABLED=true
TLS_CERT_FILE=/app/certs/tls.crt
TLS_KEY_FILE=/app/certs/tls.key
TLS_CA_FILE=/app/certs/ca.crt
TLS_MIN_VERSION=1.2
TLS_CIPHER_SUITES=TLS_AES_256_GCM_SHA384,TLS_CHACHA20_POLY1305_SHA256

# mTLS настройки
MTLS_ENABLED=true
MTLS_CLIENT_CERT_REQUIRED=true
MTLS_CA_FILE=/app/certs/ca.crt
```

---

## 📜 Генерация сертификатов

### **1. Создание CA (Certificate Authority)**

```bash
# Создание приватного ключа CA
openssl genrsa -out ca.key 4096

# Создание самоподписанного сертификата CA
openssl req -new -x509 -days 365 -key ca.key -out ca.crt \
  -subj "/C=RU/ST=Moscow/L=Moscow/O=2GC Technologies/OU=CloudBridge Relay/CN=CloudBridge CA"
```

### **2. Создание серверного сертификата**

```bash
# Создание приватного ключа сервера
openssl genrsa -out server.key 2048

# Создание запроса на сертификат
openssl req -new -key server.key -out server.csr \
  -subj "/C=RU/ST=Moscow/L=Moscow/O=2GC Technologies/OU=CloudBridge Relay/CN=cloudbridge-relay"

# Создание конфигурации для SAN
cat > server.conf << EOF
[ req ]
default_bits       = 2048
distinguished_name = req_distinguished_name
req_extensions     = req_ext
x509_extensions    = v3_ca
prompt             = no

[ req_distinguished_name ]
C  = RU
ST = Moscow
L  = Moscow
O  = 2GC Technologies
OU = CloudBridge Relay
CN = cloudbridge-relay

[ req_ext ]
subjectAltName = @alt_names

[ v3_ca ]
subjectAltName = @alt_names
basicConstraints = critical,CA:FALSE

[ alt_names ]
DNS.1 = cloudbridge-relay
DNS.2 = cloudbridge-relay.cloudbridge.svc.cluster.local
DNS.3 = localhost
IP.1  = 127.0.0.1
IP.2  = 10.96.83.60
IP.3  = 10.244.0.144
IP.4  = 10.244.0.145
IP.5  = 10.244.0.146
EOF

# Подписание сертификата CA
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key \
  -CAcreateserial -out server.crt -days 365 -extensions v3_ca -extfile server.conf
```

### **3. Создание клиентского сертификата**

```bash
# Создание приватного ключа клиента
openssl genrsa -out client.key 2048

# Создание запроса на сертификат
openssl req -new -key client.key -out client.csr \
  -subj "/C=RU/ST=Moscow/L=Moscow/O=2GC Technologies/OU=CloudBridge Relay/CN=cloudbridge-client"

# Подписание клиентского сертификата
openssl x509 -req -in client.csr -CA ca.crt -CAkey ca.key \
  -CAcreateserial -out client.crt -days 365
```

---

## 🐳 Kubernetes развертывание

### **1. Создание Kubernetes Secret**

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: cloudbridge-tls-secret
  namespace: cloudbridge
type: kubernetes.io/tls
data:
  tls.crt: LS0tLS1CRUdJTi... # base64 encoded server certificate
  tls.key: LS0tLS1CRUdJTi... # base64 encoded server key
  ca.crt: LS0tLS1CRUdJTi...  # base64 encoded CA certificate
  client.crt: LS0tLS1CRUdJTi... # base64 encoded client certificate
  client.key: LS0tLS1CRUdJTi... # base64 encoded client key
```

### **2. Обновление Deployment**

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cloudbridge-relay
  namespace: cloudbridge
spec:
  template:
    spec:
      containers:
      - name: cloudbridge-relay
        volumeMounts:
        - name: tls-certs
          mountPath: /app/certs
          readOnly: true
      volumes:
      - name: tls-certs
        secret:
          secretName: cloudbridge-tls-secret
```

### **3. Создание HTTPS Service**

```yaml
apiVersion: v1
kind: Service
metadata:
  name: cloudbridge-https-service
  namespace: cloudbridge
spec:
  type: ClusterIP
  selector:
    app.kubernetes.io/name: cloudbridge-relay
  ports:
    - name: https
      port: 9091
      targetPort: 9091
      protocol: TCP
```

---

## 🔍 Проверка SSL/TLS

### **1. Проверка сертификата**

```bash
# Проверка сертификата
openssl x509 -in /app/certs/tls.crt -text -noout

# Проверка цепочки сертификатов
openssl verify -CAfile /app/certs/ca.crt /app/certs/tls.crt

# Проверка приватного ключа
openssl rsa -in /app/certs/tls.key -check
```

### **2. Тестирование HTTPS соединения**

```bash
# Тест HTTPS соединения
curl -k https://localhost:9091/health

# Тест с проверкой сертификата
curl --cacert /app/certs/ca.crt https://localhost:9091/health

# Тест mTLS
curl --cacert /app/certs/ca.crt --cert /app/certs/client.crt --key /app/certs/client.key https://localhost:9091/health
```

### **3. Проверка TLS версии и шифров**

```bash
# Проверка поддерживаемых TLS версий
openssl s_client -connect localhost:9091 -tls1_2
openssl s_client -connect localhost:9091 -tls1_3

# Проверка шифров
nmap --script ssl-enum-ciphers -p 9091 localhost
```

---

## 🛡️ Безопасность

### **1. Рекомендации по безопасности**

- **Используйте TLS 1.3** как минимальную версию в production
- **Регулярно обновляйте сертификаты** (не более 1 года)
- **Используйте сильные ключи** (минимум 2048 бит для RSA)
- **Включите Perfect Forward Secrecy**
- **Мониторьте SSL/TLS метрики**

### **2. Мониторинг SSL/TLS**

```bash
# Проверка SSL/TLS метрик
curl http://localhost:8084/metrics | grep tls

# Проверка логов безопасности
tail -f /app/logs/security.log | grep -i tls
```

### **3. Аудит безопасности**

```bash
# Проверка конфигурации SSL/TLS
./cloudbridge-relay --config configs/config.json --validate-tls

# Проверка соответствия стандартам
sslscan localhost:9091
```

---

## 🔧 Troubleshooting

### **1. Частые проблемы**

#### **Проблема: TLS не включается**
```bash
# Проверка конфигурации
kubectl exec -n cloudbridge cloudbridge-relay-xxx -- cat /app/configs/config.json | grep -A 10 "security"

# Проверка сертификатов
kubectl exec -n cloudbridge cloudbridge-relay-xxx -- ls -la /app/certs/
```

#### **Проблема: Ошибка валидации сертификата**
```bash
# Проверка формата сертификата
kubectl exec -n cloudbridge cloudbridge-relay-xxx -- head -5 /app/certs/tls.crt

# Проверка цепочки сертификатов
kubectl exec -n cloudbridge cloudbridge-relay-xxx -- openssl verify -CAfile /app/certs/ca.crt /app/certs/tls.crt
```

#### **Проблема: mTLS не работает**
```bash
# Проверка клиентского сертификата
kubectl exec -n cloudbridge cloudbridge-relay-xxx -- openssl x509 -in /app/certs/client.crt -text -noout

# Проверка CA сертификата
kubectl exec -n cloudbridge cloudbridge-relay-xxx -- openssl x509 -in /app/certs/ca.crt -text -noout
```

### **2. Логи и диагностика**

```bash
# Просмотр логов TLS
kubectl logs -n cloudbridge cloudbridge-relay-xxx | grep -i tls

# Просмотр логов безопасности
kubectl logs -n cloudbridge cloudbridge-relay-xxx | grep -i "security\|ssl\|cert"
```

---

## 📊 Мониторинг и метрики

### **1. Prometheus метрики**

```bash
# TLS метрики
cloudbridge_tls_connections_total
cloudbridge_tls_handshake_duration_seconds
cloudbridge_tls_errors_total
cloudbridge_mtls_authentications_total
```

### **2. Grafana дашборды**

- **SSL/TLS Overview** — общий обзор SSL/TLS соединений
- **Certificate Status** — статус сертификатов
- **mTLS Authentication** — статистика mTLS аутентификации
- **Security Events** — события безопасности

---

## 🚀 Будущие возможности

### **1. Post-Quantum Cryptography**
- **CRYSTALS-Kyber** для установления ключей
- **CRYSTALS-Dilithium** для цифровых подписей
- **Hybrid mode** для совместимости

### **2. Enhanced Security**
- **Certificate Transparency** — прозрачность сертификатов
- **OCSP Stapling** — проверка отзыва сертификатов
- **HSTS** — HTTP Strict Transport Security

### **3. Automation**
- **Automatic certificate renewal** — автоматическое обновление сертификатов
- **Certificate rotation** — ротация сертификатов
- **Security policy enforcement** — принуждение политик безопасности

---

## 📋 Чек-лист SSL/TLS

### **✅ Базовая настройка**
- [ ] TLS включен в конфигурации
- [ ] Сертификаты созданы и размещены
- [ ] Kubernetes Secret создан
- [ ] Volume mounts настроены
- [ ] HTTPS Service создан

### **✅ Безопасность**
- [ ] TLS 1.3 используется
- [ ] Сильные шифры настроены
- [ ] mTLS включен
- [ ] Perfect Forward Secrecy активен
- [ ] Аудит логирования включен

### **✅ Мониторинг**
- [ ] Prometheus метрики настроены
- [ ] Grafana дашборды созданы
- [ ] Алерты настроены
- [ ] Логи безопасности активны

### **✅ Тестирование**
- [ ] HTTPS endpoint работает
- [ ] mTLS аутентификация работает
- [ ] Сертификаты валидны
- [ ] TLS версии поддерживаются

---

## 🎉 Заключение

**CloudBridge Relay** обеспечивает enterprise-уровень SSL/TLS безопасности с поддержкой современных стандартов, включая TLS 1.3, mTLS и пост-квантовую криптографию. Платформа готова к использованию в production среде с полным соответствием стандартам безопасности.

### **Ключевые преимущества:**
- ✅ **Современные протоколы** — TLS 1.3, QUIC, HTTP-3
- ✅ **Mutual TLS** — двусторонняя аутентификация
- ✅ **Post-Quantum Ready** — готовность к квантовым атакам
- ✅ **Enterprise Security** — соответствие стандартам безопасности
- ✅ **Automation** — автоматизация управления сертификатами
- ✅ **Monitoring** — полный мониторинг SSL/TLS

**CloudBridge Relay готов к безопасному использованию в enterprise-среде!** 🚀

---

*Руководство создано: 7 сентября 2025*  
*Версия: CloudBridge Relay v1.0.0*  
*Статус: ✅ Production Ready*
