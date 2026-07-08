# Отчёт агента — cursor: port scheme sync

**Дата:** 2026-07-09 01:33  
**Миссия:** cloudbridge-relay-medium  
**Шаг:** 2

## Выполнено

1. Собрана **canonical** таблица портов из:
   - `configs/config.json` — installer-конфиг relay (5550/5551/5552, http_api 9090, https_api 5553, WireGuard 51820, TURN pool 3478)
   - `docs/operations/SERVICE_CATALOG.md` — микросервисы (IPAM, AI, Cadence, PostgreSQL, …)
   - `k8s-manifests/06-services.yaml`, `05-deployments.yaml`, `09-zitadel.yaml` — cluster exposure
   - Runtime-fixed порты `cmd/relay/main.go` (9090/9092 QUIC UDP, 5553 P2P QUIC, 19302 STUN, 3478 TURN, 8443 MASQUE) — задокументированы, код не менялся
2. Создан единый источник правды: **`openwiki/port-scheme.md`** — таблицы relay + auxiliary + k8s + firewall minimum + drift TODO.
3. Синхронизирован **`openwiki/`**:
   - `quickstart.md` — исправлена секция Core Ports (убраны ошибки: WireGuard на 5553, AI на 8082, дубли 9090)
   - `architecture.md` — диаграмма relay/AI/IPAM портов
   - `operations.md` — EXPOSE + OVA compose ссылка на port-scheme
   - `domain-networking.md` — ссылка и краткий список relay-портов
4. Исправлены противоречия в **`docs/network/`**:
   - `PORTS_SUMMARY.md` — ссылка на canonical openwiki
   - `INTERNET_PORTS_FULL.md` — 51820 помечен как WireGuard (было «WebRTC»)
5. Runtime-порты в коде и docker-контейнеры smarty **не трогались**.

### Canonical relay (installer `configs/config.json`)

| Listener | Port | Protocol |
|---|---:|---|
| Main TCP | 5550 | TCP |
| Metrics | 5551 | TCP |
| P2P REST API | 5552 | TCP |
| HTTP API | 9090 | TCP |
| HTTPS API | 5553 | TCP |
| P2P QUIC | 5553 | UDP (hardcoded) |
| QUIC transport | 9090 | UDP (hardcoded) |
| STUN | 19302 | UDP |
| TURN | 3478 | UDP/TCP |
| WireGuard | 51820 | UDP |
| MASQUE | 8443 | TCP/UDP |
| gRPC | 8444 | TCP |

## Изменённые файлы

| Путь | Действие |
|------|----------|
| `openwiki/port-scheme.md` | создан — canonical service → port |
| `openwiki/quickstart.md` | Core Ports + навигация |
| `openwiki/architecture.md` | порты в topology diagram |
| `openwiki/operations.md` | Dockerfile EXPOSE + OVA ports |
| `openwiki/domain-networking.md` | ссылка на port-scheme |
| `docs/network/PORTS_SUMMARY.md` | ссылка на canonical |
| `docs/network/INTERNET_PORTS_FULL.md` | WireGuard вместо WebRTC для 51820 |
| `docs/agent-reports/cloudbridge-relay-medium/step02-port-scheme-sync.md` | этот отчёт |

## Проверки (команды и результат)

```bash
# Парсинг installer config (duplicate keys — last wins)
python3 -c "
import json
with open('configs/config.json') as f:
    d=json.load(f)
print('server.port:', d['server']['port'])
print('metrics.port:', d['metrics']['port'])
print('api.port:', d['api']['port'])
print('http_api.port:', d['http_api']['port'])
print('https_api.port:', d['https_api']['port'])
print('wireguard.listen_port:', d['wireguard']['interface']['listen_port'])
"
# server.port: 5550
# metrics.port: 5551
# api.port: 5552
# http_api.port: 9090
# https_api.port: 5553
# wireguard.listen_port: 51820

# k8s service ports
rg -n 'port:' k8s-manifests/06-services.yaml
# relay 9090/51820/8444, ai 8080/9090, cadence 7933-7939, cassandra 9042

# openwiki больше не содержит «5553 = WireGuard» / «8082 = AI»
rg '5553.*WireGuard|8082.*AI|Port 9090' openwiki/
# только domain-networking.md с корректным контекстом P2P QUIC

test -f openwiki/port-scheme.md && wc -l openwiki/port-scheme.md
# 96 openwiki/port-scheme.md
```

## Блокеры / не сделано

| Item | Причина |
|------|---------|
| Обновление `docs/operations/SERVICE_CATALOG.md` | В scope — только openwiki + docs/network; catalog помечен как drift в `port-scheme.md` TODO |
| Исправление duplicate keys в `configs/config.json` | Требует отдельного config cleanup (не менять runtime без необходимости) |
| Синхронизация `config/config.json` (dev, port 9090) vs `configs/config.json` | Два профиля; canonical для installer — `configs/` |
| AI Service port (8085 catalog / 8080 k8s / 8083 code) | Зафиксирован drift в таблице; единый winner без изменения кода/k8s не выбран |
| IPAM port (8083 catalog / 8080 code) | Аналогично — только документировано |
| `docker/Dockerfile` EXPOSE 8082 vs config api.port 5552 | Задокументировано; правка Dockerfile вне scope «не менять runtime» |

**Следующий шаг (3/5):** по плану миссии — следующий пункт medium audit (после портов).
