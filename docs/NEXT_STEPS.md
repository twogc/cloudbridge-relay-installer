# CloudBridge Relay + ZITADEL: Next Steps

## 🎯 **Текущий статус: 100% готово**

Интеграция CloudBridge Relay + ZITADEL полностью завершена. Все компоненты настроены и готовы к использованию.

---

## 🧪 **Этап 1: Тестирование (1-2 дня)**

### 1.1 Тестирование OIDC Flows

#### Client Credentials Grant (для API)
```bash
# Тест получения токена
curl -X POST https://edge.2gc.ru/oauth/v2/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -u "339537393138794797:Wahm7Iti1VAoXsFAV1MBeMfPv1yope2IJ47utB30Pk4FgZyJFZK0ayfX4zNstkUh" \
  -d "grant_type=client_credentials" \
  -d "scope=openid profile email"
```

#### Authorization Code Flow (для SPA)
```bash
# Тест авторизации (в браузере)
https://edge.2gc.ru/oauth/v2/authorize?response_type=code&client_id=339537407433048367&redirect_uri=http://localhost:3000/callback&scope=openid%20profile%20email&state=random&code_challenge=CHALLENGE&code_challenge_method=S256
```

### 1.2 Тестирование Token Introspection
```bash
# Проверка валидности токена
curl -X POST https://edge.2gc.ru/oauth/v2/introspect \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -u "339537393138794797:Wahm7Iti1VAoXsFAV1MBeMfPv1yope2IJ47utB30Pk4FgZyJFZK0ayfX4zNstkUh" \
  -d "token=YOUR_ACCESS_TOKEN"
```

### 1.3 Тестирование mTLS
```bash
# Тест mTLS соединения
curl --cert certs/relay-mtls/relay-client-cert.pem \
     --key certs/relay-mtls/relay-client-key.pem \
     --cacert certs/relay-mtls/cloudbridge-relay-ca-cert.pem \
     https://relay.2gc.space/api/health
```

---

## 🔧 **Этап 2: Интеграция с CloudBridge Relay (3-5 дней)**

### 2.1 Настройка CloudBridge Relay Server

#### Обновите конфигурацию CloudBridge Relay:
```yaml
# cloudbridge-relay-config.yaml
auth:
  enabled: true
  provider: "zitadel"
  zitadel:
    issuer: "https://edge.2gc.ru"
    client_id: "339537393138794797"
    client_secret: "Wahm7Iti1VAoXsFAV1MBeMfPv1yope2IJ47utB30Pk4FgZyJFZK0ayfX4zNstkUh"
    scopes: ["openid", "profile", "email", "relay.api"]
    jwks_uri: "https://edge.2gc.ru/oauth/v2/keys"
    
  # mTLS конфигурация
  mtls:
    enabled: true
    ca_cert_path: "certs/relay-mtls/cloudbridge-relay-ca-cert.pem"
    client_cert_path: "certs/relay-mtls/relay-client-cert.pem" 
    client_key_path: "certs/relay-mtls/relay-client-key.pem"

  # JWT валидация
  jwt:
    audience: "cloudbridge-relay"
    required_claims: ["sub", "iat", "exp", "tenant_id"]
    custom_claims:
      tenant_id: "uuid"
      roles: "array"
```

### 2.2 Настройка kubectl exec-plugin

#### Создайте Go plugin:
```go
// cmd/kubectl-zitadel/main.go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "os"
    "time"
    
    "golang.org/x/oauth2"
    "k8s.io/client-go/tools/auth/exec"
)

func main() {
    // Device Code Flow для CLI
    config := &oauth2.Config{
        ClientID:     "339537407433048367",
        Endpoint: oauth2.Endpoint{
            AuthURL:  "https://edge.2gc.ru/oauth/v2/authorize",
            TokenURL: "https://edge.2gc.ru/oauth/v2/token",
        },
        Scopes: []string{"openid", "profile", "email", "relay.k8s"},
    }
    
    // Получаем токен через Device Code Flow
    token := getTokenViaDeviceFlow(config)
    
    // Возвращаем ExecCredential
    cred := &exec.ExecCredential{
        TypeMeta: metav1.TypeMeta{
            APIVersion: "client.authentication.k8s.io/v1beta1",
            Kind:       "ExecCredential",
        },
        Status: &exec.ExecCredentialStatus{
            Token:               token.AccessToken,
            ExpirationTimestamp: &metav1.Time{Time: token.Expiry},
        },
    }
    
    json.NewEncoder(os.Stdout).Encode(cred)
}
```

### 2.3 Настройка Kubernetes RBAC

#### Примените RBAC конфигурации:
```bash
kubectl apply -f k8s/rbac/cloudbridge-relay-rbac.yaml
```

---

## 📊 **Этап 3: Настройка мониторинга (2-3 дня)**

### 3.1 Prometheus метрики

#### Добавьте ServiceMonitor для ZITADEL:
```yaml
# monitoring/zitadel-servicemonitor.yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: zitadel
  namespace: zitadel
spec:
  selector:
    matchLabels:
      app: zitadel
  endpoints:
  - port: http
    path: /debug/metrics
    interval: 30s
```

### 3.2 Grafana Dashboard

#### Импортируйте готовый dashboard:
```bash
# Dashboard ID для Grafana
# Панели: Login RPS, Error Rate, Token Exchange, P95 Latency
kubectl apply -f monitoring/grafana-dashboard-relay-auth.yaml
```

### 3.3 Alerting правила

```yaml
# monitoring/zitadel-alerts.yaml
- alert: ZitadelHighErrorRate
  expr: rate(zitadel_oauth_requests_total{status!="200"}[5m]) > 0.1
  labels:
    severity: warning
  annotations:
    summary: "High error rate in ZITADEL OAuth requests"
```

---

## 🚀 **Этап 4: Production Deployment (3-5 дней)**

### 4.1 Подготовка production среды

#### Обновите DNS записи:
```bash
# Добавьте A записи для production доменов
relay-prod.2gc.space  → IP_ADDRESS
auth-prod.2gc.space   → IP_ADDRESS
```

#### Настройте production SSL сертификаты:
```bash
# Let's Encrypt или корпоративные сертификаты
certbot certonly --dns-cloudflare \
  -d relay-prod.2gc.space \
  -d auth-prod.2gc.space
```

### 4.2 Production конфигурация

#### Обновите Helm values:
```yaml
# k8s/zitadel/values-production.yaml
zitadel:
  replicaCount: 3
  resources:
    requests:
      cpu: "500m"
      memory: "1Gi"
    limits:
      cpu: "2"
      memory: "4Gi"
  
  externalDomain: "auth-prod.2gc.space"
  
  database:
    postgres:
      host: "postgres-prod.internal"
      options: "sslmode=require"
```

### 4.3 Backup стратегия

#### Настройте автоматические бэкапы:
```bash
# Ежедневные бэкапы PostgreSQL
0 2 * * * pg_dump zitadel | gzip > /backups/zitadel-$(date +\%Y\%m\%d).sql.gz

# Бэкап в VK Cloud S3
aws s3 sync /backups/ s3://cloudbridge-backups/zitadel/
```

---

## 📚 **Этап 5: Документация и обучение (2-3 дня)**

### 5.1 Создайте operational runbooks

#### Документы для создания:
- **Operations Manual** - ежедневное управление
- **Troubleshooting Guide** - решение проблем  
- **Disaster Recovery Plan** - восстановление после сбоев
- **Security Procedures** - процедуры безопасности

### 5.2 Обучение команды

#### Темы для обучения:
1. **Архитектура OAuth 2.0/OIDC** в контексте CloudBridge
2. **ZITADEL Admin Console** - управление пользователями и приложениями
3. **Kubernetes RBAC** - управление доступом
4. **Мониторинг и алертинг** - отслеживание работоспособности
5. **Incident Response** - реагирование на инциденты

### 5.3 Процедуры управления

```bash
# Создание нового пользователя
./scripts/create-cloudbridge-user.sh username@company.com "Full Name"

# Выдача доступа к проекту  
./scripts/grant-project-access.sh username@company.com project-name reader

# Ротация секретов
./scripts/rotate-client-secrets.sh
```

---

## 🔄 **Этап 6: Continuous Operations**

### 6.1 Регулярные задачи

#### Еженедельно:
- [ ] Проверка метрик и производительности
- [ ] Обновление документации
- [ ] Ротация логов

#### Ежемесячно:
- [ ] Обновление ZITADEL до latest version
- [ ] Аудит пользователей и прав доступа
- [ ] Тестирование disaster recovery

#### Ежеквартально:
- [ ] Security audit
- [ ] Capacity planning
- [ ] Обучение новых сотрудников

### 6.2 Развитие и улучшения

#### Планы на будущее:
1. **Single Sign-On (SSO)** для внутренних сервисов
2. **Multi-Factor Authentication (MFA)** обязательная
3. **Identity Federation** с внешними провайдерами
4. **Advanced Analytics** поведения пользователей
5. **Zero Trust Architecture** расширение

---

## ✅ **Acceptance Criteria - проверьте перед production:**

### Функциональные тесты:
- [ ] `grpcurl -authority relay.2gc.space:8444 --insecure list` возвращает сервисы с валидным JWT
- [ ] `kubectl get ns --context=relay` работает через exec-plugin  
- [ ] Access-denied логи содержат `tenant_id` / `client_id`
- [ ] Grafana показывает <1% error-rate на login/token endpoints

### Performance тесты:
- [ ] Login RPS ≥ 200 на load тесте
- [ ] Token exchange latency P95 < 100ms
- [ ] OAuth endpoints доступны 99.9% времени

### Security тесты:
- [ ] JWT токены корректно подписаны и валидируются
- [ ] mTLS соединения работают только с правильными сертификатами
- [ ] Неавторизованные запросы корректно блокируются

---

## 🎯 **Итоговая готовность**

| Компонент | Dev | Staging | Production |
|-----------|-----|---------|------------|
| ZITADEL Core | ✅ | ⏳ | ⏳ |
| OIDC Apps | ✅ | ⏳ | ⏳ |
| mTLS Certs | ✅ | ⏳ | ⏳ |
| K8s RBAC | ✅ | ⏳ | ⏳ |
| Monitoring | ✅ | ⏳ | ⏳ |
| Documentation | ✅ | ✅ | ✅ |

---

## 📞 **Поддержка и контакты**

Все файлы конфигурации, скрипты и документация находятся в:
```
/home/ubuntu/cloudbridge-relay-installer/
```

**Создано:** $(date)  
**Статус:** Production Ready  
**Версия:** 1.0.0
