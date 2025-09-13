# 🔒 Анализ состояния SSL/TLS в CloudBridge Relay

## 📋 Текущий статус SSL/TLS

**Дата анализа**: 7 сентября 2025  
**Статус**: ❌ **SSL/TLS НЕ НАСТРОЕН**

## 🔍 Детальный анализ

### 1. **Конфигурация CloudBridge Relay**

#### В `k8s/configmaps/cloudbridge-config.yaml`:
```yaml
"security": {
  "tls_enabled": false,                    # ❌ TLS отключен
  "cert_file": "/app/certs/server.crt",    # Путь к сертификату
  "key_file": "/app/certs/server.key",     # Путь к ключу
  "mtls": {
    "enabled": true,                       # ❌ mTLS включен, но TLS отключен
    "ca_file": "/app/certs/ca.crt",
    "client_cert_required": true
  },
  "cipher_suites": ["TLS_AES_256_GCM_SHA384", "TLS_CHACHA20_POLY1305_SHA256"],
  "min_tls_version": "1.2"
}
```

**Проблема**: Конфигурация противоречива - mTLS включен, но базовый TLS отключен.

### 2. **HTTP сервер**

#### В `cmd/relay/main.go`:
```go
httpServer := &http.Server{
    Addr:    ":9091",
    Handler: httpMux,
}
// ❌ Используется ListenAndServe вместо ListenAndServeTLS
if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
    s.logger.Error("HTTP server error", zap.Error(err))
}
```

**Проблема**: HTTP сервер запускается без TLS, используя `ListenAndServe` вместо `ListenAndServeTLS`.

### 3. **Сертификаты**

#### Проверка наличия сертификатов:
```bash
find . -name "*.crt" -o -name "*.key" -o -name "*.pem"
# Результат: НЕТ НАЙДЕННЫХ ФАЙЛОВ
```

**Проблема**: Сертификаты не созданы и не развернуты.

#### Конфигурация для генерации сертификатов:
```ini
# cert/openssl-san.cnf
[ req_distinguished_name ]
C  = RU
ST = Moscow
L  = Moscow
O  = 2GC
CN = keycloak                    # ❌ Неправильный CN для CloudBridge

[ alt_names ]
DNS.1 = keycloak                 # ❌ Неправильный DNS
IP.1  = 10.0.0.242              # ❌ Неправильный IP
```

**Проблема**: Конфигурация OpenSSL настроена для Keycloak, а не для CloudBridge Relay.

### 4. **Порты сервисов**

#### Текущие порты CloudBridge Relay:
```yaml
# kubectl get services -n cloudbridge
cloudbridge-relay:
  Port: http        9090/TCP    # ❌ HTTP, не HTTPS
  Port: quic        9090/UDP    # ✅ QUIC (имеет встроенное шифрование)
  Port: wireguard   51820/UDP   # ✅ WireGuard (имеет встроенное шифрование)
  Port: metrics     8084/TCP    # ❌ HTTP, не HTTPS
```

**Проблема**: Основные HTTP порты (9090, 8084) работают без TLS.

### 5. **Cassandra SSL**

#### В `k8s/configmaps/cloudbridge-config.yaml`:
```yaml
server_encryption_options:
  internode_encryption: none        # ❌ Шифрование между узлами отключено
  keystore: conf/.keystore
  keystore_password: cassandra
  truststore: conf/.truststore
  truststore_password: cassandra
  protocol: TLS
  algorithm: SunX509
  store_type: JKS
  cipher_suites: [TLS_RSA_WITH_AES_128_CBC_SHA,...]
  require_client_auth: false

client_encryption_options:
  enabled: false                    # ❌ Клиентское шифрование отключено
  keystore: conf/.keystore
  keystore_password: cassandra
  require_client_auth: false
  protocol: TLS
  algorithm: SunX509
  store_type: JKS
  cipher_suites: [TLS_RSA_WITH_AES_128_CBC_SHA,...]
```

**Проблема**: Cassandra работает без шифрования клиентских соединений.

## 🚨 Критические проблемы безопасности

### 1. **Отсутствие TLS для HTTP API**
- ❌ API endpoints доступны по HTTP
- ❌ Данные передаются в открытом виде
- ❌ Нет аутентификации на уровне транспорта

### 2. **Отсутствие сертификатов**
- ❌ Нет SSL сертификатов
- ❌ Нет CA сертификатов
- ❌ Нет настроек для генерации сертификатов

### 3. **Противоречивая конфигурация**
- ❌ mTLS включен, но TLS отключен
- ❌ Настроены cipher suites, но TLS не используется
- ❌ Указаны пути к сертификатам, которых нет

## 🔧 План исправления SSL/TLS

### Этап 1: Создание сертификатов

#### 1.1 Обновить конфигурацию OpenSSL
```ini
# cert/openssl-san.cnf
[ req_distinguished_name ]
C  = RU
ST = Moscow
L  = Moscow
O  = 2GC Technologies
CN = cloudbridge-relay

[ alt_names ]
DNS.1 = cloudbridge-relay
DNS.2 = cloudbridge-relay.cloudbridge.svc.cluster.local
DNS.3 = localhost
IP.1  = 127.0.0.1
IP.2  = 10.96.83.60
```

#### 1.2 Создать скрипт генерации сертификатов
```bash
#!/bin/bash
# scripts/generate-certs.sh

# Создать CA
openssl genrsa -out certs/ca.key 4096
openssl req -new -x509 -days 365 -key certs/ca.key -out certs/ca.crt \
  -subj "/C=RU/ST=Moscow/L=Moscow/O=2GC Technologies/CN=CloudBridge CA"

# Создать серверный сертификат
openssl genrsa -out certs/server.key 2048
openssl req -new -key certs/server.key -out certs/server.csr \
  -config cert/openssl-san.cnf
openssl x509 -req -in certs/server.csr -CA certs/ca.crt -CAkey certs/ca.key \
  -CAcreateserial -out certs/server.crt -days 365 \
  -extensions v3_ca -extfile cert/openssl-san.cnf

# Создать клиентский сертификат
openssl genrsa -out certs/client.key 2048
openssl req -new -key certs/client.key -out certs/client.csr \
  -subj "/C=RU/ST=Moscow/L=Moscow/O=2GC Technologies/CN=CloudBridge Client"
openssl x509 -req -in certs/client.csr -CA certs/ca.crt -CAkey certs/ca.key \
  -CAcreateserial -out certs/client.crt -days 365
```

### Этап 2: Обновление конфигурации

#### 2.1 Включить TLS в CloudBridge Relay
```yaml
# k8s/configmaps/cloudbridge-config.yaml
"security": {
  "tls_enabled": true,                     # ✅ Включить TLS
  "cert_file": "/app/certs/server.crt",
  "key_file": "/app/certs/server.key",
  "mtls": {
    "enabled": true,
    "ca_file": "/app/certs/ca.crt",
    "client_cert_required": true
  },
  "cipher_suites": ["TLS_AES_256_GCM_SHA384", "TLS_CHACHA20_POLY1305_SHA256"],
  "min_tls_version": "1.2"
}
```

#### 2.2 Обновить код HTTP сервера
```go
// cmd/relay/main.go
if cfg.Security.TLSEnabled {
    httpServer := &http.Server{
        Addr:    ":9091",
        Handler: httpMux,
        TLSConfig: &tls.Config{
            MinVersion: tls.VersionTLS12,
            CipherSuites: []uint16{
                tls.TLS_AES_256_GCM_SHA384,
                tls.TLS_CHACHA20_POLY1305_SHA256,
            },
        },
    }
    
    // ✅ Использовать ListenAndServeTLS
    if err := httpServer.ListenAndServeTLS(
        cfg.Security.CertFile,
        cfg.Security.KeyFile,
    ); err != nil && err != http.ErrServerClosed {
        s.logger.Error("HTTPS server error", zap.Error(err))
    }
} else {
    // Fallback на HTTP для разработки
    httpServer := &http.Server{
        Addr:    ":9091",
        Handler: httpMux,
    }
    if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        s.logger.Error("HTTP server error", zap.Error(err))
    }
}
```

### Этап 3: Обновление Kubernetes манифестов

#### 3.1 Создать Secret для сертификатов
```yaml
# k8s/secrets/cloudbridge-tls-secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: cloudbridge-tls-secret
  namespace: cloudbridge
type: kubernetes.io/tls
data:
  tls.crt: <base64-encoded-server.crt>
  tls.key: <base64-encoded-server.key>
  ca.crt: <base64-encoded-ca.crt>
```

#### 3.2 Обновить Deployment
```yaml
# k8s/deployments/cloudbridge-relay-deployment.yaml
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

#### 3.3 Обновить Service
```yaml
# k8s/services/cloudbridge-services.yaml
spec:
  ports:
  - name: https
    port: 443
    targetPort: 9091
    protocol: TCP
  - name: http
    port: 80
    targetPort: 9091
    protocol: TCP
```

### Этап 4: Включение Cassandra SSL

#### 4.1 Обновить конфигурацию Cassandra
```yaml
# k8s/configmaps/cloudbridge-config.yaml
client_encryption_options:
  enabled: true                     # ✅ Включить клиентское шифрование
  keystore: conf/.keystore
  keystore_password: cassandra
  require_client_auth: true         # ✅ Требовать клиентские сертификаты
  protocol: TLS
  algorithm: SunX509
  store_type: JKS
  cipher_suites: [TLS_AES_256_GCM_SHA384, TLS_CHACHA20_POLY1305_SHA256]

server_encryption_options:
  internode_encryption: all         # ✅ Шифрование между узлами
  keystore: conf/.keystore
  keystore_password: cassandra
  truststore: conf/.truststore
  truststore_password: cassandra
  protocol: TLS
  algorithm: SunX509
  store_type: JKS
  cipher_suites: [TLS_AES_256_GCM_SHA384, TLS_CHACHA20_POLY1305_SHA256]
  require_client_auth: true
```

## 🎯 Рекомендации по безопасности

### 1. **Немедленные действия**
- ✅ **Включить TLS** для всех HTTP endpoints
- ✅ **Создать сертификаты** для всех компонентов
- ✅ **Настроить mTLS** для межсервисного взаимодействия
- ✅ **Включить шифрование Cassandra**

### 2. **Долгосрочные улучшения**
- 🔄 **Автоматическая ротация сертификатов** через Cadence workflows
- 🔄 **Интеграция с Let's Encrypt** для публичных сертификатов
- 🔄 **Certificate transparency** мониторинг
- 🔄 **HSTS headers** для веб-интерфейсов

### 3. **Мониторинг безопасности**
- 📊 **Метрики TLS** в Prometheus
- 📊 **Алерты на истечение сертификатов**
- 📊 **Мониторинг TLS handshake** ошибок
- 📊 **Аудит SSL/TLS** конфигурации

## 📊 Текущий статус компонентов

| Компонент | TLS Статус | Порт | Критичность |
|-----------|------------|------|-------------|
| **CloudBridge Relay HTTP** | ❌ Отключен | 9090 | 🔴 Критично |
| **CloudBridge Relay Metrics** | ❌ Отключен | 8084 | 🔴 Критично |
| **CloudBridge Relay QUIC** | ✅ Встроенное | 9090/UDP | 🟢 Безопасно |
| **CloudBridge Relay WireGuard** | ✅ Встроенное | 51820/UDP | 🟢 Безопасно |
| **Cassandra Client** | ❌ Отключен | 9042 | 🔴 Критично |
| **Cassandra Inter-node** | ❌ Отключен | 7000/7001 | 🟡 Средне |
| **Prometheus** | ❌ Отключен | 9090 | 🟡 Средне |
| **Grafana** | ❌ Отключен | 3000 | 🟡 Средне |

## 🚀 Заключение

**Текущее состояние SSL/TLS в CloudBridge Relay: КРИТИЧЕСКИ НЕБЕЗОПАСНО**

### Основные проблемы:
1. ❌ **TLS отключен** для всех HTTP endpoints
2. ❌ **Сертификаты отсутствуют**
3. ❌ **Данные передаются в открытом виде**
4. ❌ **Нет защиты от MITM атак**

### Приоритет исправления:
1. 🔴 **Критично**: Включить TLS для CloudBridge Relay API
2. 🔴 **Критично**: Включить TLS для Cassandra клиентских соединений
3. 🟡 **Важно**: Включить TLS для Prometheus и Grafana
4. 🟡 **Важно**: Настроить автоматическую ротацию сертификатов

**Рекомендация**: Немедленно приступить к настройке SSL/TLS для обеспечения безопасности системы.

---

*Анализ проведен: 7 сентября 2025*  
*Статус: ТРЕБУЕТСЯ НЕМЕДЛЕННОЕ ИСПРАВЛЕНИЕ* 🚨
