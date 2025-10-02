# Быстрая проверка Cadence

## Текущий статус
```bash
kubectl get pods -n cloudbridge | grep cadence
```

## Проверка логов
```bash
kubectl logs -n cloudbridge $(kubectl get pods -n cloudbridge | grep cadence | awk '{print $1}') --tail=20
```

## Проверка сервиса
```bash
kubectl get svc -n cloudbridge | grep cadence
```

## Тестирование подключения
```bash
# Port forward
kubectl port-forward -n cloudbridge svc/cadence 7934:7934 &

# Проверка health
curl http://localhost:7934/health

# Проверка gRPC
kubectl port-forward -n cloudbridge svc/cadence 7933:7933 &
```

## Файлы для исправления
- **Основной**: `k8s/cadence/cadence-simple.yaml`
- **Альтернативные**: `k8s/cadence/cadence-*.yaml`
- **Docker**: `docker/compose/docker-compose-cadence.yml`

## Интеграция в коде
- **Сервер**: `cmd/relay/main.go:474-494`
- **Client**: `internal/cadence/client.go`
- **P2P**: `internal/p2p/cadence_integration.go`
- **Конфиг**: `configs/config.json:463-468`

## Ожидаемый результат
- Pod: `1/1 Running`
- Health: `{"status":"ok"}`
- Логи: без ошибок
