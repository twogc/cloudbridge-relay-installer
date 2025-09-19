# 🎯 Краткое резюме для разработчика сервера

## 📋 **Что нужно сделать**

### **1. Унифицировать токены:**
- Поддержать `connection_type=quic` с `protocol_type=p2p-mesh`
- Убрать требование `wireguard_config` для WireGuard
- Упростить валидацию токенов

### **2. Удалить WireGuard:**
- Убрать `WireGuardConfig` из структур
- Убрать `NetworkConfig`, `MeshConfig`, `PeerWhitelist`
- Оставить только необходимые поля

## 🔧 **Файлы для изменения**

### **Основные:**
- `internal/token/p2p_claims.go` - убрать WireGuard поля
- `internal/token/validator.go` - упростить валидацию
- `internal/api/peer_handlers.go` - обновить регистрацию

### **Конфигурация:**
- `k8s/deployments/cloudbridge-relay-p2p-deployment.yaml`
- `k8s/services/p2p-api-service.yaml`

## 🐳 **Сборка Minikube**

```bash
# 1. Собрать образ
docker build -t cloudbridge-relay-p2p:$(date +%Y%m%d-%H%M%S)-unified-tokens .

# 2. Загрузить в Minikube
minikube image load cloudbridge-relay-p2p:$(date +%Y%m%d-%H%M%S)-unified-tokens

# 3. Обновить deployment
kubectl set image deployment/cloudbridge-relay-p2p \
  cloudbridge-relay-p2p=cloudbridge-relay-p2p:$(date +%Y%m%d-%H%M%S)-unified-tokens \
  -n cloudbridge
```

## 📋 **Чек-лист**

- [ ] Убрать WireGuard поддержку
- [ ] Поддержать `connection_type=quic`
- [ ] Упростить валидацию токенов
- [ ] Собрать новый образ
- [ ] Протестировать в Minikube

## 🎯 **Результат**

**Единый стандарт токенов для P2P mesh сети без WireGuard!**
