# 🧪 БЫСТРАЯ ПРОВЕРКА IP ВАЛИДАЦИИ

## 📋 **ПРОВЕРКА ДО ИСПРАВЛЕНИЯ**

### 1. Health Check (должен работать)
```bash
curl "http://95.163.250.190:30082/health"
```
**Ожидаемый результат:** `{"status":"healthy",...}`

### 2. Peer Discovery (должен работать)
```bash
curl -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJodHRwczovL2F1dGguMmdjLnJ1L3JlYWxtcy9yZWxheS1zZXJ2ZXJzIiwic3ViIjoicDJwLXNlcnZlci1jbGllbnQiLCJhdWQiOlsicmVsYXktc2VydmVyLWNsaWVudCJdLCJleHAiOjE3NTc2NDE2MDcsIm5iZiI6MTc1NzYzODAwNywiaWF0IjoxNzU3NjM4MDA3LCJqdGkiOiJoeWJyaWRfdG9rZW5fMTc1NzYzODAwNyIsInRlbmFudF9pZCI6InRlc3QtdGVuYW50LTEyMyIsImRldmljZV9pZCI6ImRldmljZV8xNzU3NjMxOTI0IiwicGVybWlzc2lvbnMiOlsicDJwOmNvbm5lY3QiLCJwMnA6ZGlzY292ZXIiLCJwMnA6bWVzaCJdLCJjb25uZWN0aW9uX3R5cGUiOiJoeWJyaWQiLCJ3aXJlZ3VhcmRfY29uZmlnIjp7ImFkZHJlc3MiOiIxMC4wLjAuMS8yNCIsImFsbG93ZWRfaXBzIjpbIjEwLjAuMC4wLzI0Il0sImludGVyZmFjZV9uYW1lIjoid2cwIiwibGlzdGVuX3BvcnQiOjUxODIwLCJtdHUiOjE0MjB9fQ.FW-5hAjaiYXZlhvl7wN5aZZAZuKtMQBJRMUBFDmY3dk" "http://95.163.250.190:30082/api/v1/peers/discover?tenant_id=test-tenant-123"
```
**Ожидаемый результат:** `{"success":true,"tenant_id":"test-tenant-123","peers":null}`

### 3. Peer Registration (НЕ РАБОТАЕТ - проблема!)
```bash
curl -X POST \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJodHRwczovL2F1dGguMmdjLnJ1L3JlYWxtcy9yZWxheS1zZXJ2ZXJzIiwic3ViIjoicDJwLXNlcnZlci1jbGllbnQiLCJhdWQiOlsicmVsYXktc2VydmVyLWNsaWVudCJdLCJleHAiOjE3NTc2NDE2MDcsIm5iZiI6MTc1NzYzODAwNywiaWF0IjoxNzU3NjM4MDA3LCJqdGkiOiJoeWJyaWRfdG9rZW5fMTc1NzYzODAwNyIsInRlbmFudF9pZCI6InRlc3QtdGVuYW50LTEyMyIsImRldmljZV9pZCI6ImRldmljZV8xNzU3NjMxOTI0IiwicGVybWlzc2lvbnMiOlsicDJwOmNvbm5lY3QiLCJwMnA6ZGlzY292ZXIiLCJwMnA6bWVzaCJdLCJjb25uZWN0aW9uX3R5cGUiOiJoeWJyaWQiLCJ3aXJlZ3VhcmRfY29uZmlnIjp7ImFkZHJlc3MiOiIxMC4wLjAuMS8yNCIsImFsbG93ZWRfaXBzIjpbIjEwLjAuMC4wLzI0Il0sImludGVyZmFjZV9uYW1lIjoid2cwIiwibGlzdGVuX3BvcnQiOjUxODIwLCJtdHUiOjE0MjB9fQ.FW-5hAjaiYXZlhvl7wN5aZZAZuKtMQBJRMUBFDmY3dk" \
  -H "Content-Type: application/json" \
  -d '{"peer_id":"test-peer-001","tenant_id":"test-tenant-123","org_id":"test-org-456","device_id":"device_1757631924","connection_type":"hybrid","wireguard_config":{"endpoint":"192.168.1.100:51820","public_key":"dGVzdC1wdWJsaWMta2V5LWZvci10ZXN0aW5nLW9ubHk=","allowed_ips":["10.0.0.0/24"],"listen_port":51820}}' \
  "http://95.163.250.190:30082/api/v1/peers/register"
```
**Ожидаемый результат:** `{"error":"Invalid WireGuard endpoint","code":400,"message":"endpoint IP not allowed: 192.168.1.100"}`

## ✅ **ПРОВЕРКА ПОСЛЕ ИСПРАВЛЕНИЯ**

### 1. Health Check (должен работать)
```bash
curl "http://95.163.250.190:30082/health"
```
**Ожидаемый результат:** `{"status":"healthy",...}`

### 2. Peer Discovery (должен работать)
```bash
curl -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJodHRwczovL2F1dGguMmdjLnJ1L3JlYWxtcy9yZWxheS1zZXJ2ZXJzIiwic3ViIjoicDJwLXNlcnZlci1jbGllbnQiLCJhdWQiOlsicmVsYXktc2VydmVyLWNsaWVudCJdLCJleHAiOjE3NTc2NDE2MDcsIm5iZiI6MTc1NzYzODAwNywiaWF0IjoxNzU3NjM4MDA3LCJqdGkiOiJoeWJyaWRfdG9rZW5fMTc1NzYzODAwNyIsInRlbmFudF9pZCI6InRlc3QtdGVuYW50LTEyMyIsImRldmljZV9pZCI6ImRldmljZV8xNzU3NjMxOTI0IiwicGVybWlzc2lvbnMiOlsicDJwOmNvbm5lY3QiLCJwMnA6ZGlzY292ZXIiLCJwMnA6bWVzaCJdLCJjb25uZWN0aW9uX3R5cGUiOiJoeWJyaWQiLCJ3aXJlZ3VhcmRfY29uZmlnIjp7ImFkZHJlc3MiOiIxMC4wLjAuMS8yNCIsImFsbG93ZWRfaXBzIjpbIjEwLjAuMC4wLzI0Il0sImludGVyZmFjZV9uYW1lIjoid2cwIiwibGlzdGVuX3BvcnQiOjUxODIwLCJtdHUiOjE0MjB9fQ.FW-5hAjaiYXZlhvl7wN5aZZAZuKtMQBJRMUBFDmY3dk" "http://95.163.250.190:30082/api/v1/peers/discover?tenant_id=test-tenant-123"
```
**Ожидаемый результат:** `{"success":true,"tenant_id":"test-tenant-123","peers":null}`

### 3. Peer Registration (ДОЛЖЕН РАБОТАТЬ!)
```bash
curl -X POST \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJodHRwczovL2F1dGguMmdjLnJ1L3JlYWxtcy9yZWxheS1zZXJ2ZXJzIiwic3ViIjoicDJwLXNlcnZlci1jbGllbnQiLCJhdWQiOlsicmVsYXktc2VydmVyLWNsaWVudCJdLCJleHAiOjE3NTc2NDE2MDcsIm5iZiI6MTc1NzYzODAwNywiaWF0IjoxNzU3NjM4MDA3LCJqdGkiOiJoeWJyaWRfdG9rZW5fMTc1NzYzODAwNyIsInRlbmFudF9pZCI6InRlc3QtdGVuYW50LTEyMyIsImRldmljZV9pZCI6ImRldmljZV8xNzU3NjMxOTI0IiwicGVybWlzc2lvbnMiOlsicDJwOmNvbm5lY3QiLCJwMnA6ZGlzY292ZXIiLCJwMnA6bWVzaCJdLCJjb25uZWN0aW9uX3R5cGUiOiJoeWJyaWQiLCJ3aXJlZ3VhcmRfY29uZmlnIjp7ImFkZHJlc3MiOiIxMC4wLjAuMS8yNCIsImFsbG93ZWRfaXBzIjpbIjEwLjAuMC4wLzI0Il0sImludGVyZmFjZV9uYW1lIjoid2cwIiwibGlzdGVuX3BvcnQiOjUxODIwLCJtdHUiOjE0MjB9fQ.FW-5hAjaiYXZlhvl7wN5aZZAZuKtMQBJRMUBFDmY3dk" \
  -H "Content-Type: application/json" \
  -d '{"peer_id":"test-peer-001","tenant_id":"test-tenant-123","org_id":"test-org-456","device_id":"device_1757631924","connection_type":"hybrid","wireguard_config":{"endpoint":"192.168.1.100:51820","public_key":"dGVzdC1wdWJsaWMta2V5LWZvci10ZXN0aW5nLW9ubHk=","allowed_ips":["10.0.0.0/24"],"listen_port":51820}}' \
  "http://95.163.250.190:30082/api/v1/peers/register"
```
**Ожидаемый результат:** `{"success":true,"peer_id":"test-peer-001","registered_at":"2025-09-12T01:00:00Z","discovery_endpoint":"wss://relay.2gc.ru/ws/peers"}`

## 🔍 **ДОПОЛНИТЕЛЬНЫЕ ПРОВЕРКИ**

### Логи приложения
```bash
kubectl logs -n cloudbridge -l app.kubernetes.io/name=cloudbridge-relay --tail=10
```

### Статус подов
```bash
kubectl get pods -n cloudbridge
```

### Статус deployment
```bash
kubectl rollout status deployment/cloudbridge-relay -n cloudbridge
```

## 📊 **КРИТЕРИИ УСПЕХА**

✅ **Исправление работает, если:**
1. Health endpoint возвращает `{"status":"healthy"}`
2. Peer discovery возвращает `{"success":true}`
3. Peer registration возвращает `{"success":true}` (НЕ ошибку!)
4. Логи не содержат "endpoint IP not allowed"

❌ **Если что-то сломалось:**
- Health endpoint не отвечает → проблема с деплоем
- Peer discovery не работает → проблема с JWT
- Peer registration все еще ошибка → IP whitelist не настроен
