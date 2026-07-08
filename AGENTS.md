# CloudBridge Relay Installer — Agent Instructions

> **Repository:** `cloudbridge-relay-installer`  
> **Purpose:** Go-микросервисная платформа P2P/QUIC/WireGuard-ретрансляции, деплой в VK Cloud K8s и B2B OVA (MWS Marketplace).  
> **Companion docs:** [openwiki/quickstart.md](openwiki/quickstart.md) · [docs/README.md](docs/README.md) · [docs/audit/MIMO_AUDIT_REPORT.md](docs/audit/MIMO_AUDIT_REPORT.md)

---

## Activation — когда применять эти инструкции

Используй этот файл, если задача касается:

- сборки, тестов или рефакторинга Go-сервисов (`cmd/`, `internal/`, `config/`)
- деплоя в Kubernetes / VK Cloud (`k8s-manifests/`, `terraform-k8s/`, `scripts/deploy/`)
- auth/IAM (только **Zitadel** — не Keycloak/Django)
- сети: QUIC, WireGuard, ICE/STUN/TURN, MASQUE, P2P
- мониторинга (Prometheus, Grafana, Zabbix)
- исправлений по аудиту или оркестрации VOLY (Cursor + MiMo)
- документации архитектуры, портов, операционных runbook'ов

**Не активируй для:** содержимого `voly/` (локальный orchestration, в `.gitignore`), `docs/legacy/` (архив без правок без явного запроса).

---

## Первые шаги (обязательно)

1. Прочитай **[openwiki/quickstart.md](openwiki/quickstart.md)** — структура репо, пути деплоя, prerequisites.
2. При аудите/рефакторинге — **[docs/audit/MIMO_AUDIT_REPORT.md](docs/audit/MIMO_AUDIT_REPORT.md)**.
3. Перед правками кода — найди существующие конвенции в соседних файлах; **минимальный diff**.

---

## Технический контекст

| Параметр | Значение |
|----------|----------|
| Язык | Go **1.25.12+** (`go.mod`) |
| CGO | Включён (`CGO_ENABLED=1`) — SQLite, WireGuard |
| Основной бинарник | `cmd/relay/` → `make build` |
| Auth (canonical) | **Zitadel** (`internal/zitadel/`, `internal/auth/`) |
| Устаревшее (удалять) | Keycloak, Django-конфиг, `internal/quantum/`, `src/` (TS SDK) |
| Деплой (canonical) | `scripts/deploy/deploy.sh` + `k8s-manifests/` |
| K8s cluster | `terraform-k8s/` |

### Ключевые директории

```
cmd/           — 13 entrypoints (relay, api, ai-*, cadence, ipam, …)
internal/      — ядро: p2p, quic, nat, wireguard, auth, ai, cadence, metrics
config/        — загрузка конфигурации
k8s-manifests/ — нумерованные манифесты 00–15
docker/        — Dockerfile'ы сервисов
docs/          — человекочитаемая документация
openwiki/      — навигация для агентов
docs/agent-reports/ — отчёты агентов (обязательно после VOLY-шагов)
```

---

## Сборка и тесты

```bash
# Prerequisites: Go 1.25.12+, gcc, libsqlite3-dev (для CGO)

make build              # bin/cloudbridge-relay
make test               # go test ./...
make test-critical      # token, auth, p2p, api, quic, zitadel, config
make test-unit          # -short ./internal/... ./config/...
go build ./...
go test ./... -short
```

Перед сдачей задачи с изменениями Go: `go build ./...` и релевантные `go test`.

---

## Правила для агентов

### Scope

- **Минимальные изменения** строго по задаче; не расширять scope.
- **Не коммитить** без явного запроса пользователя.
- **Не трогать** `voly/`, секреты, `.env`, ключи API.
- **Не восстанавливать** Keycloak/Django — провайдер IAM только Zitadel.
- Legacy в `docs/legacy/` — только чтение, правки по запросу.

### Документация

| Что изменил | Обнови |
|-------------|--------|
| Порты / сервисы | `openwiki/`, `docs/network/`, `docs/operations/SERVICE_CATALOG.md` |
| Auth / Zitadel | `openwiki/architecture.md`, `docs/architecture/` |
| Деплой | `docs/guides/`, `scripts/deploy/` |
| Аудит / фиксы | `docs/audit/MIMO_AUDIT_REPORT.md` (чеклист Quick wins) |

### Отчётность (VOLY / Cursor / MiMo)

После **каждого** шага оркестрации или крупной задачи агента создай markdown-отчёт в:

```
docs/agent-reports/<mission-name>/stepNN-<slug>.md
```

Структура отчёта: дата, шаг, изменённые файлы, что сделано, результаты проверок, блокеры.  
Итог миссии — `docs/agent-reports/<mission-name>/SUMMARY.md`.  
См. [docs/agent-reports/README.md](docs/agent-reports/README.md).

---

## Оркестрация VOLY (локально)

`voly/` — AI control plane (Cursor + MiMo), **не в git**. Запуск из локальной копии:

```bash
cd voly && source .venv/bin/activate
export RELAY_INSTALLER_PATH=/path/to/cloudbridge-relay-installer
export VOLY_PROJECT_CWD=$RELAY_INSTALLER_PATH

# Миссии
python -m projects.twogc twogc combat list
python -m projects.twogc twogc combat run cloudbridge-relay-fixes --sequential

# UI
voly ui   # http://127.0.0.1:7788
```

| Агент | Роль |
|-------|------|
| **cursor** | Правки файлов, реализация, quick wins |
| **mimo** | Дешёвый текстовый анализ (без записи файлов) |

Активные миссии: `cloudbridge-relay-audit`, `cloudbridge-relay-fixes` (по `MIMO_AUDIT_REPORT.md`).

---

## Текущие приоритеты (из аудита)

### Quick wins (в работе / частично)

- [x] `go.mod` → 1.25.12
- [x] `Makefile` `test-critical` → `./internal/zitadel/...`
- [ ] Удалить Keycloak-код и скрипты
- [ ] Удалить Django/quantum/`src/`
- [ ] Удалить `.github/workflows/*.disabled`
- [ ] Синхронизировать keycloak-блоки в k8s configmaps

### Средний срок

- Консолидация дубликатов (`*_improved`, `*_enhanced`, `*_production`)
- Объединение `k8s/` + `k8s-manifests/`, terraform-директорий
- Единая портовая схема (canonical → `openwiki/`)

Полный список: [docs/audit/MIMO_AUDIT_REPORT.md](docs/audit/MIMO_AUDIT_REPORT.md).

---

## Выбор подхода по типу задачи

| Задача | Подход |
|--------|--------|
| Однострочный фикс, опечатка | Прямая правка, без оркестрации |
| Quick wins из аудита | VOLY `cloudbridge-relay-fixes` или Cursor с отчётом в `docs/agent-reports/` |
| Новый сервис / крупный рефакторинг | План в `docs/`, затем инкрементальные PR-размерные diff'ы |
| Только анализ | MiMo или read-only; отчёт пишет Cursor в `docs/agent-reports/` |
| Деплой в VK Cloud | `openwiki/operations.md`, `scripts/deploy/deploy.sh` |

---

## Ссылки

| Документ | Назначение |
|----------|------------|
| [openwiki/quickstart.md](openwiki/quickstart.md) | Старт для агентов |
| [docs/README.md](docs/README.md) | Индекс документации |
| [docs/audit/MIMO_AUDIT_REPORT.md](docs/audit/MIMO_AUDIT_REPORT.md) | Аудит: мёртвый код, дубли, архитектура |
| [docs/audit/GO_VERSION_INVENTORY.md](docs/audit/GO_VERSION_INVENTORY.md) | Версии Go в репозитории |
| [docs/agent-reports/README.md](docs/agent-reports/README.md) | Папка отчётов агентов |
| [docs/operations/SERVICE_CATALOG.md](docs/operations/SERVICE_CATALOG.md) | Каталог сервисов и портов |
| [Makefile](Makefile) | build, test, docker, deploy targets |

---

*Формат companion-инструкций: [agent-skill-creator](https://github.com/FrancyJGLisboa/agent-skill-creator) (AAIF). Обновляй этот файл при смене canonical-путей деплоя, auth или структуры отчётности.*
