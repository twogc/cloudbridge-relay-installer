# Отчёт агента — cursor: SERVICE_CATALOG port sync

**Дата:** 2026-07-09 01:40  
**Миссия:** cloudbridge-relay-next  
**Шаг:** 2

## Выполнено

1. Прочитаны `openwiki/port-scheme.md`, `docs/operations/SERVICE_CATALOG.md`, `configs/config.json`.
2. Сверены канонические порты relay (server/metrics/api/http/https/quic/wireguard) и вспомогательные AI/IPAM с `port-scheme.md`.
3. Обновлён `docs/operations/SERVICE_CATALOG.md`:
   - добавлена ссылка на `openwiki/port-scheme.md` как единый источник портов;
   - секция **CloudBridge Relay** — полная таблица портов (5550/5551/5552/9090/5553/QUIC/WireGuard и др.) вместо legacy `8080/8443/9090`;
   - пример конфигурации заменён на `configs/config.json` (installer canonical);
   - **IPAM**, **AI Service**, **AI Worker** — каталожные порты сохранены (8083/8085/8086) с явными блоками **Drift** для runtime/k8s;
   - **WireGuard Service** — уточнено совпадение с relay `51820`;
   - **Monitoring** и **Network Topology** — обновлены порты relay (`:5551/metrics`, topology `5550/5552`).
4. Исправлены duplicate top-level keys в `configs/config.json`:
   - удалены первые блоки `http_api` (port 5553) и `https_api` (port 5554);
   - оставлены единственные ключи: `http_api.port` → **9090**, `https_api.port` → **5553** (как в port-scheme).
5. Runtime-порты в Go **не менялись** (по инструкции).

## Изменённые файлы

| Путь | Действие |
|------|----------|
| `docs/operations/SERVICE_CATALOG.md` | обновлены порты relay/metrics/api/http/https/quic/wireguard; drift для IPAM/AI; monitoring/topology |
| `configs/config.json` | удалены дублирующие `http_api` / `https_api` (первые вхождения) |
| `docs/agent-reports/cloudbridge-relay-next/step02-service-catalog-port-sync.md` | создан (этот отчёт) |

## Проверки (команды и результат)

```bash
# JSON парсится однозначно; канонические порты совпадают с port-scheme
python3 -c "import json; d=json.load(open('configs/config.json')); ..."
```

```
JSON parse: OK
http_api.port: 9090
https_api.port: 5553
server.port: 5550
metrics.port: 5551
api.port: 5552
wireguard.interface.listen_port: 51820
Top-level duplicate keys: none (для http_api/https_api)
```

```bash
jq -e '.http_api.port, .https_api.port, .server.port, .metrics.port, .api.port' configs/config.json
grep -c '"http_api"' configs/config.json   # ожидается 1
grep -c '"https_api"' configs/config.json  # ожидается 1
```

```
9090
5553
5550
5551
5552
---
1
1
```

```bash
# Подтверждение runtime drift (не исправлялось)
grep -n ':8080\|AI_PORT\|:8085' cmd/ipam-service/main.go cmd/ai-service/main.go cmd/ai-worker/main.go
```

```
cmd/ipam-service/main.go:54:    Addr:    ":8080",
cmd/ai-service/main.go:74:    port := os.Getenv("AI_PORT")
cmd/ai-service/main.go:76:        port = "8083"
cmd/ai-worker/main.go:46:        Addr: ":8085",
```

## Блокеры / не сделано

- **Runtime drift IPAM/AI не устранён** — требует согласованного изменения `cmd/*` и/или k8s manifests (вне scope step 2; Go не трогали).
- **`openwiki/port-scheme.md` строка 6** всё ещё упоминает duplicate keys в `configs/config.json` — устарела после fix; обновление port-scheme не входило в step 2.
- **Мелкие опечатки** в SERVICE_CATALOG (`roduction Ready`, `equired`, `ptional`) не трогались — вне scope port sync.
- **Commit/push не выполнялись** (по инструкции).
