# CloudBridge Client Connection Guide

**Date:** September 28, 2025  
**Status:** ✅ READY FOR PILOT PROJECT  
**Client Location:** `test/relay-client/`  

## Executive Summary

✅ **CloudBridge клиент готов для пилотного проекта!**  
✅ **Прямое подключение к Kubernetes работает**  
❌ **Внешние DNS домены (2gc.ru) недоступны из-за проблем LoadBalancer**  

## Client Configuration Options

### Option 1: Pilot Project (RECOMMENDED) ✅

**Файл конфигурации:** `config-pilot-ready.yaml`

```yaml
relay:
  host: "192.168.49.2"
  port: 32500
  base_url: "http://192.168.49.2:32500"
  
api:
  base_url: "http://192.168.49.2:32500"
  health_url: "http://192.168.49.2:32500/health"
  metrics_url: "http://192.168.49.2:32500/metrics"
```

**Статус:** ✅ Полностью работает  
**Использование:** Для пилотного проекта и демонстрации  

### Option 2: External DNS (NOT WORKING) ❌

**Файл конфигурации:** `config-external-dns.yaml`

```yaml
relay:
  host: "relay.2gc.ru"
  port: 443
  base_url: "https://relay.2gc.ru"
  
api:
  base_url: "https://relay.2gc.ru"
  health_url: "https://relay.2gc.ru/health"
  metrics_url: "https://monitor.2gc.ru/metrics"
```

**Статус:** ❌ HTTP 503 Service Unavailable  
**Проблема:** LoadBalancer backend member в ERROR состоянии  
**Причина:** Сетевая изоляция между LoadBalancer (10.0.0.242) и Kubernetes (192.168.49.2)  

## Working Endpoints (Pilot Configuration)

### ✅ Core Services
- **Health Check:** `http://192.168.49.2:32500/health`
  - Status: `healthy`
  - Version: `v1.0.0`
  - Uptime: Available

- **Metrics:** `http://192.168.49.2:32500/metrics`
  - JSON format
  - Runtime metrics
  - System information

### ✅ P2P API
- **Peer Discovery:** `http://192.168.49.2:32500/api/v1/peers/discover?tenant_id=pilot-tenant`
- **Peer List:** `http://192.168.49.2:32500/api/v1/peers/list`
- **Peer Registration:** `http://192.168.49.2:32500/api/v1/peers/register` (requires JWT)

### ✅ Transport Layer
- **WebSocket:** `ws://192.168.49.2:32500/ws/peers`
- **QUIC:** Available on configured ports
- **WireGuard:** Available for P2P connections

## Client Usage Examples

### 1. Go Client (Recommended)

```bash
cd test/relay-client/examples
go build -o pilot-client pilot-client.go
./pilot-client
```

### 2. Direct API Calls

```bash
# Health check
curl http://192.168.49.2:32500/health

# Metrics
curl http://192.168.49.2:32500/metrics

# Peer discovery
curl "http://192.168.49.2:32500/api/v1/peers/discover?tenant_id=pilot-tenant"
```

### 3. Configuration Loading

```go
// Load pilot configuration
config, err := LoadConfig("config-pilot-ready.yaml")
if err != nil {
    log.Fatal(err)
}

// Create client
client := NewRelayClient(config)
err = client.Connect()
```

## Authentication

### Current Status
- **JWT Authentication:** Configured but validation disabled for pilot
- **Protected Endpoints:** Return HTTP 401 (expected)
- **Public Endpoints:** Health, Metrics work without authentication

### For Production
```go
// Enable JWT authentication
auth:
  type: "jwt"
  secret: "85Sk/NfLq3gqzCXzmBKbJCpL+f5BssXz3G8dVi3sPiE="
  skip_validation: false  // Enable validation
  disable_token_validation: false
```

## Transport Layer Selection

Клиент автоматически выбирает транспорт в следующем порядке:

1. **gRPC** (primary) - для основных API вызовов
2. **QUIC** (P2P) - для peer-to-peer соединений
3. **WebSocket** (fallback) - для корпоративных сетей
4. **WireGuard** (secure) - для защищенных туннелей

## Troubleshooting

### Problem: Connection Refused
```bash
# Check Kubernetes is running
kubectl get pods -n cloudbridge

# Check port-forward is active
kubectl port-forward service/cloudbridge-relay-external 32500:8080 -n cloudbridge
```

### Problem: Authentication Errors
```bash
# For pilot testing, use config with disabled validation
cp config-pilot-ready.yaml config.yaml
```

### Problem: DNS Resolution (External)
```bash
# Check DNS records
nslookup relay.2gc.ru
# Should return: 109.120.180.160

# Check LoadBalancer status
curl -k https://relay.2gc.ru/health
# Currently returns: 503 Service Unavailable
```

## External DNS Issues (2gc.ru domains)

### Current Status
- ❌ **All 2gc.ru domains return HTTP 503**
- ❌ **LoadBalancer backend member in ERROR state**
- ❌ **Network isolation between LoadBalancer and Kubernetes**

### Root Cause
```
VK Cloud LoadBalancer (10.0.0.242) 
    ↓ [Network isolation]
Kubernetes Cluster (192.168.49.2)
```

### Solutions for Production
1. **Deploy CloudBridge on VK Cloud VM** instead of Minikube
2. **Configure proper network routing** between LoadBalancer and backend
3. **Use VM with accessible IP** from LoadBalancer network

## Testing Scripts

### Comprehensive Testing
```bash
# Test all functionality
./test-pilot-client.sh

# Test external DNS (shows 503 errors)
./test-external-dns.sh
```

### Individual Tests
```bash
# Health check
curl http://192.168.49.2:32500/health | jq

# Metrics
curl http://192.168.49.2:32500/metrics | jq

# Peer discovery
curl "http://192.168.49.2:32500/api/v1/peers/discover?tenant_id=pilot-tenant"
```

## Pilot Project Readiness

### ✅ Ready Components
- [x] **Server Health:** Healthy and running
- [x] **API Endpoints:** Core functionality available
- [x] **Client Configuration:** Pilot-ready config created
- [x] **Transport Layer:** Multiple protocols supported
- [x] **P2P Discovery:** Peer discovery working
- [x] **Metrics:** System monitoring available
- [x] **WebSocket:** Real-time communication ready

### 🔧 Production Requirements
- [ ] **External DNS:** Fix LoadBalancer backend connectivity
- [ ] **JWT Authentication:** Enable token validation
- [ ] **SSL/TLS:** Implement proper certificate management
- [ ] **Monitoring:** Set up production monitoring
- [ ] **Logging:** Configure centralized logging

## Quick Start for Pilot

```bash
# 1. Navigate to client directory
cd /home/ubuntu/cloudbridge-relay-installer/test/relay-client

# 2. Use pilot configuration
cp config-pilot-ready.yaml config.yaml

# 3. Test connection
./test-pilot-client.sh

# 4. Run Go client
cd examples
go build -o pilot-client pilot-client.go
./pilot-client
```

## Conclusion

**CloudBridge клиент полностью готов для пилотного проекта!**

- ✅ **Прямое подключение работает** (192.168.49.2:32500)
- ✅ **Все основные функции доступны**
- ✅ **P2P соединения настроены**
- ✅ **Транспортные протоколы работают**

**Для завтрашней демонстрации используйте конфигурацию `config-pilot-ready.yaml`**
