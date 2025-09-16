# JWT Token Fix Report - P2P API

## 🎯 Проблема
JWT токен из личного кабинета содержал `heartbeat_interval` в строковом формате `"30s"`, но сервер ожидал числовой формат `30`. Это вызывало ошибку при парсинге токена:

```json
{"error":"Invalid token","code":401,"message":"failed to parse P2P JWT: token is malformed: could not JSON decode claim: json: cannot unmarshal string into Go struct field MeshConfig.mesh_config.heartbeat_interval of type time.Duration"}
```

## ✅ Решение
Исправлен код JWT валидации для поддержки как строкового, так и числового формата `heartbeat_interval`.

### Изменения в коде:

#### 1. `internal/token/mesh_config.go`
- Добавлен кастомный метод `UnmarshalJSON` для `MeshConfig`
- Поддержка форматов:
  - Строковый: `"30s"`, `"1m"`, `"2h"`
  - Числовой: `30` (секунды), `30.0` (float)
  - Int64: `30` (секунды)

#### 2. `internal/token/p2p_claims.go`
- Добавлен кастомный метод `UnmarshalJSON` для `P2PClaims`
- Обеспечивает правильную обработку `mesh_config` с строковым `heartbeat_interval`

## 🔧 Технические детали

### До исправления:
```go
type MeshConfig struct {
    HeartbeatInterval time.Duration `json:"heartbeat_interval"`
}
```

### После исправления:
```go
type MeshConfig struct {
    HeartbeatInterval time.Duration `json:"heartbeat_interval"`
}

func (m *MeshConfig) UnmarshalJSON(data []byte) error {
    // Кастомная логика для поддержки строкового и числового формата
    switch v := temp.HeartbeatInterval.(type) {
    case string:
        duration, err := time.ParseDuration(v)
        if err != nil {
            return fmt.Errorf("invalid heartbeat_interval format: %s", v)
        }
        m.HeartbeatInterval = duration
    case float64:
        m.HeartbeatInterval = time.Duration(v) * time.Second
    case int:
        m.HeartbeatInterval = time.Duration(v) * time.Second
    case int64:
        m.HeartbeatInterval = time.Duration(v) * time.Second
    }
    return nil
}
```

## 🚀 Развертывание

### 1. Сборка Docker образа:
```bash
cd /home/ubuntu/cloudbridge-relay-installer
docker build -t cloudbridge-relay:latest .
```

### 2. Развертывание в Kubernetes:
```bash
kubectl rollout restart deployment/cloudbridge-relay -n cloudbridge
kubectl rollout status deployment/cloudbridge-relay -n cloudbridge --timeout=180s
```

### 3. Проверка статуса:
```bash
kubectl get pods -n cloudbridge | grep cloudbridge-relay
```

## ✅ Результат

### Тестирование показало:
- ✅ JWT токен с `heartbeat_interval: "30s"` успешно парсится
- ✅ Значение корректно конвертируется в `30 * time.Second`
- ✅ Все остальные поля токена обрабатываются правильно
- ✅ Сервер запущен и работает без ошибок

### Пример успешного парсинга:
```json
{
  "mesh_config": {
    "network_id": "mesh-network-001",
    "subnet": "10.0.0.0/16",
    "registry_url": "https://mesh-registry.2gc.ru",
    "heartbeat_interval": 30000000000,  // 30 секунд в наносекундах
    "max_peers": 10,
    "routing_strategy": "performance_optimal",
    "enable_auto_discovery": true,
    "trust_level": "basic"
  }
}
```

## 📝 Поддерживаемые форматы

Теперь сервер поддерживает следующие форматы `heartbeat_interval`:

1. **Строковый формат:**
   - `"30s"` → 30 секунд
   - `"1m"` → 1 минута
   - `"2h"` → 2 часа

2. **Числовой формат:**
   - `30` → 30 секунд
   - `30.0` → 30 секунд
   - `1800` → 1800 секунд

## 🎉 Заключение

Проблема с JWT токеном для P2P API **полностью решена**. Сервер теперь корректно обрабатывает токены с строковым форматом `heartbeat_interval` и конвертирует их в правильный тип `time.Duration`.

**P2P API Server полностью функционален!** 🚀

## 📋 Команды для проверки

```bash
# Проверить статус подов
kubectl get pods -n cloudbridge | grep cloudbridge-relay

# Проверить логи
RELAY_POD=$(kubectl get pods -n cloudbridge -l app.kubernetes.io/name=cloudbridge-relay --field-selector=status.phase=Running -o jsonpath="{.items[0].metadata.name}")
kubectl logs -n cloudbridge $RELAY_POD -c cloudbridge-relay --tail=20

# Тестировать P2P API (через порт-форвардинг)
kubectl port-forward -n cloudbridge svc/cloudbridge-relay-lb-tcp 8082:8082 &
curl -H "Authorization: Bearer $JWT_TOKEN" "http://localhost:8082/api/v1/peers/discover?tenant_id=tenant-216420165"
```

