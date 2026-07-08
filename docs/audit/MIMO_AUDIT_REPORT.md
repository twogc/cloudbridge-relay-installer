# MiMo Audit Report — CloudBridge Relay Installer

**Дата:** 2026-07-08
**Модель:** mimo-v2.5-pro
**Статус тестов:** Верификация score-ports step 4 (2026-07-09): `go build ./cmd/relay/ ./cmd/ipam-service/ ./cmd/ai-worker/ ./cmd/ai-service/` OK; auxiliary порты IPAM **8080**, AI **8083**, AI Worker **8085** совпадают с `openwiki/port-scheme.md` + `cmd/*` + `k8s/` ops. Cadence-ai step 4: `go test -short` — cadence/activities, cadence/workflows, ai/pipeline OK; частичные FAIL в cadence/client_test, ai/integrations_test, ai/routing/*_real_data_test (pre-existing). Go 1.25.12.

---

## Executive Summary

CloudBridge Relay Installer — это масштабный Go-монорепозиторий (~175 000 строк Go-кода, 49 internal-пакетов, 13 cmd-бинарников) для P2P/QUIC/WireGuard-ретрансляции. Проект находится в активной разработке, но содержит значительное количество мёртвого кода, дублированных реализаций и архитектурных нестыковок между документацией и кодом. Критические проблемы: несогласованность портовой схемы в трёх независимых источниках, наличие кода трёх auth-провайдеров (Zitadel, Keycloak, Django) при том что только Zitadel реально используется, и множество файлов-дубликатов с суффиксами `_improved`, `_enhanced`, `_production`, которые не были удалены после создания новых версий.

---

## Результаты тестирования

| Проверка | Статус | Примечание |
|---|---|---|
| `go build ./...` | Частично | Сборка основного кода OK; `go build ./...` падает из‑за `rep/` (в .gitignore, сторонние примеры go-rod) |
| `go test -short ./internal/auth/... ./internal/zitadel/... ./internal/network/... ./config/...` | ✅ OK | Шаг 5 (2026-07-09): auth + zitadel pass с `-short`; network/config — нет test files |
| `go build ./cmd/relay/ ./cmd/ipam-service/ ./cmd/ai-worker/ ./cmd/ai-service/` | ✅ OK | score-ports step 4 (2026-07-09) |
| `go test -short ./internal/cadence/... ./internal/ai/...` | Частично | cadence-ai step 4: activities/workflows/pipeline OK; FAIL client_test (stub без Cadence), integrations_test, routing real_data, cadence_integration_real_test |
| `go test ./...` (полный) | Частично | Без `-short` zitadel может FAIL (интеграционные тесты без mock, 404 от test.zitadel.cloud) |
| `make test-critical` | Исправлено | `./internal/keycloak/...` заменён на `./internal/zitadel/...` (шаг 1) |
| go.mod Go version | ✅ Обновлено | Было `1.25.1`, каноническая версия: **`1.25.12`** (Jul 2026) |
| Статический анализ | Выполнен | 174 902 строки Go-кода, 16 тест-файлов, 42+ пакета |

---

## Лишний / мёртвый код

| Объект | Тип | Рекомендация | Приоритет |
|---|---|---|---|
| `cmd/quantum-service/` (44 строки) | Заглушка | Удалить или оставить DISABLED.md без main.go | Средний |
| `internal/quantum/` (algorithms.go, quantum.go) | Неиспользуемый пакет | Удалить — quantum-service отключён | Средний |
| `internal/metrics/quantum.go` (15K) | Метрики для отключённого сервиса | Удалить | Средний |
| `internal/cadence/workflows/quantum.go` (874 строки) | Workflow для отключённого сервиса | Удалить | Средний |
| `src/` (TypeScript: cli.ts, index.ts) | Осиротевший TS-SDK | Удалить из Go-монорепо | Высокий |
| `pkg/enhanced-quic/` (отдельный go.mod) | Изолированный модуль | Объединить с основным модулем или удалить | Средний |
| `internal/auth/keycloak_enhanced.go` | Keycloak-код | Удалить — Keycloak заменён на Zitadel | Высокий |
| `internal/auth/keycloak_enhanced_test.go` | Keycloak-тесты | Удалить вместе с кодом | Высокий |
| `internal/auth/interfaces.go` (KeycloakAuthInterface) | Dead interface | Удалить | Высокий |
| `configs/config.json` → секция `keycloak` | Конфиг мёртвого провайдера | Удалить | Высокий |
| `configs/config.json` → секция `django` | Конфиг несуществующего провайдера | Удалить — нет Go-реализации | Высокий |
| `config/config.go` → Django-структуры | Конфиг-код без реализации | Удалить | Высокий |
| `scripts/setup/setup_keycloak.sh` | Keycloak-скрипт | Удалить | Средний |
| `scripts/setup/build_keycloak_theme.sh` | Keycloak-скрипт | Удалить | Средний |
| `scripts/setup/check_keycloak_theme.sh` | Keycloak-скрипт | Удалить | Средний |
| `scripts/setup/configure_client_roles.sh` | Keycloak-скрипт | Удалить | Средний |
| `scripts/setup/setup_keycloak_theme.sh` | Keycloak-скрипт | Удалить | Средний |
| `scripts/setup/setup_ldap_integration.sh` | Keycloak-зависимый | Проверить актуальность | Средний |
| `scripts/maintenance/backup_keycloak.sh` | Keycloak-скрипт | Удалить | Средний |
| `scripts/maintenance/update_keycloak_public_key.sh` | Keycloak-скрипт | Удалить | Средний |
| `scripts/maintenance/update_keycloak_public_key_from_file.sh` | Keycloak-скрипт | Удалить | Средний |
| `.github/workflows/main.yml.disabled` | Отключённый CI | Удалить или задокументировать причину | Низкий |
| `.github/workflows/test-coverage.yml.disabled` | Отключённый CI | Удалить или задокументировать причину | Низкий |
| `.github/workflows/security-enhanced.yml.disabled` | Отключённый CI | Удалить или задокументировать причину | Низкий |
| `internal/cadence/workflows/quantum.go` → `internal/cadence/config.go` | Quantum-конфиг в cadence | Удалить quantum-ссылки | Средний |
| `k8s/configmaps/cloudbridge-config*.yaml` → keycloak-секции | K8s-конфиг мёртвого провайдера | Удалить keycloak-блоки | Средний |

---

## Дублирование и legacy

### Дублированные реализации (suффиксы `_improved`, `_enhanced`, `_production`)

| Пакет | Файл v1 | Файл v2/v3 | Рекомендация |
|---|---|---|---|
| `internal/network/` | ~~`score_tenant_network_allocator.go`~~, ~~`score_tenant_network_allocator_improved.go`~~ | `score_allocator_production.go` (14K) | ✅ Удалены v1/v2 improved; canonical — production (шаг 1, 2026-07-09) |
| `internal/network/` | ~~`retry_config.go`~~ (51 строк) | `retry_config_improved.go` (157 строк) | ✅ Удалён v1; canonical — improved (шаг 1, 2026-07-09) |
| `internal/cadence/` | `service_client.go` (70 строк) | `service_client_improved.go` (149 строк) | ✅ Dual-path intentional: ai-worker vs relay; merge отложен (cadence-ai step 2–4) |
| `internal/cadence/` | `client.go` (314 строк) | `client_production.go` (983 строки) | ✅ Dual-path intentional; stale `integration_test/cadence_integration_test.go` удалён (step 2) |
| `internal/nat/` | `turn_server.go` (8K) | `turn_server_enhanced.go` (24K) | Оставить enhanced, удалить v1 |
| `internal/health/` | `health.go` | `health_checker.go`, `comprehensive_checker.go` | Три разных health-checker'а — объединить |
| `internal/ai/anomaly/` | `detector.go` + `isolation_forest.go` | ~~`optimized_detector.go`~~, ~~`real_anomaly_detector.go`~~, ~~`real_isolation_forest.go`~~ | ✅ Canonical — detector path `ai.Service`; удалены step 3 (cadence-ai) |
| `internal/ai/predictive/` | `load_predictor.go` | ~~`optimized_predictor.go`~~, ~~`real_load_predictor.go`~~ | ✅ Canonical — load_predictor; удалены step 3 |
| `internal/ai/ml/` | `model_implementations.go` + `types.go` | ~~`real_models.go`~~, ~~`integration.go`~~ | ✅ Типы вынесены в `types.go`; удалены step 3 |
| `internal/ai/pipeline/` | `collectors.go` | ~~`real_data_collector.go`~~ | ✅ Canonical — collectors; удалён step 3 (аудит: `internal/ai/data/`) |
| `internal/ai/routing/` | `optimizer.go` | ~~`optimized_optimizer.go`~~ | ✅ Canonical — optimizer; удалён step 3 |
| `internal/ai/` | `service.go` (canonical) | ~~`real_service.go`~~, ~~`optimized_service.go`~~, ~~`cache/optimized_cache.go`~~ | ✅ Удалены step 3; 0 prod entrypoints у Real/Optimized |

### Дублирование инфраструктуры

| Дубликат | Где | Рекомендация |
|---|---|---|
| Terraform | `terraform/` vs `terraform-k8s/` vs `infra/` | Объединить в `terraform-k8s/`, удалить `terraform/` и `infra/` |
| K8s-манифесты | `k8s/` vs `k8s-manifests/` | ✅ First step (k8s-consol, 2026-07-09): роли разделены — `k8s-manifests/` prod deploy, `k8s/` ops; README, архив 4× `*-fixed.yaml`, drift в docs; полный merge — отложено |
| Prometheus | `prometheus/` vs `prometheus_multiclusters/` | Объединить (различие — 4 строки в prometheus.yml) |
| Zabbix | `zabbix-agent/` vs `scripts/zabbix/` | Объединить в `monitoring/zabbix/` |
| Deploy | `deploy/` vs `scripts/deploy/` | Объединить в `scripts/deploy/` |
| Dockerfile | `docker/Dockerfile.cadence` vs `Dockerfile.cadence.improved` vs `Dockerfile.cadence-custom` | Оставить одну версию |
| Dockerfile | `docker/Dockerfile` vs `docker/Dockerfile.simple` | Оставить основной |
| Systemd | `systemd/quic-proxy.service` == `systemd/enhanced-quic-proxy.service` | Удалить дубликат (файлы идентичны!) |
| Systemd | `systemd/relay.service` vs `systemd/cloudbridge-relay.service` | Оставить один |
| Test dirs | `test/` vs `tests/` | Объединить в `test/` |
| Nginx | `nginx/` — 10 конфигов, включая экспериментальные | Удалить неиспользуемые |

### Legacy и архивы

| Объект | Объём | Статус |
|---|---|---|
| `docs/legacy/` | 142 файла | Архив, не поддерживается. Есть `_index.md` с миграционной таблицей |
| `archive/docs/` | 6 файлов | Дополнительный архив поверх `docs/legacy/` |
| `archive/scripts/` | 4 скрипта | Старые скрипты |
| `archive/configs/` | есть | Старые конфиги |
| `github_actions/` | 14 файлов (Dockerfile, docker-compose) | Пакет для Cadence — misplaced, не связан с GitHub Actions |

---

## Архитектурные нестыковки

### 1. Несогласованность портовой схемы

| Источник | Main Relay Port | AI Service Port | IPAM Port | WireGuard Port |
|---|---|---|---|---|
| `openwiki/port-scheme.md` (canonical) | 5550 + 9090 http_api | **8083** | **8080** | 51820 |
| `cmd/*` runtime | — | **8083** | **8080** | — |
| `k8s/` ops manifests | 9090 probes | **8083** | CNI → **8080** | 51820 |
| `docs/operations/SERVICE_CATALOG.md` | 5550 + 9090 (sync step 2) | **8083** | **8080** | 51820 |
| `configs/config.json` | 5550 (+ 9090 http_api) | — | — | 51820 |
| `k8s-manifests/06-services.yaml` | 9090 | **8083/8084** | — | 51820 |

**Статус (k8s-prod-fix step 4, 2026-07-09):** Canonical — `openwiki/port-scheme.md`; auxiliary microservices выровнены: IPAM **8080**, AI **8083**, AI Worker **8085** (`cmd/*` + `k8s/` + `SERVICE_CATALOG`). **P0 stub fix ✅:** `k8s-manifests/05–06` — убраны `nginx:latest`/port 80/AI 8080 stubs; production images + `deploy.sh` load_images. **Остаточный drift:** `k8s/services/ai-service.yaml` metrics **8084** vs runtime `/metrics` on **8083**; ops `k8s/deployments/` relay http **8080** legacy; prometheus annotation **8084** vs **5551**; dev `config/config.json` `server.port` 9090.

**Wiring (не в scope score-ports):** `NewSCORETenantNetworkAllocatorProduction` — 0 prod-вызовов; `relay_manager_hybrid.go` принимает `TenantNetworkAllocator`, но allocator не подключён (см. `DUPLICATES_INVENTORY.md` §SCORE).

### 2. Auth: три провайдера, один реальный

| Провайдер | Код | Конфиг | Документация | Статус |
|---|---|---|---|---|
| **Zitadel** | `internal/auth/`, `internal/zitadel/` | `configs/config.json` | openwiki, docs/guides/ | Активен (основной) |
| **Keycloak** | `internal/auth/keycloak_enhanced.go` | `configs/config.json` | openwiki, docs/legacy | Код есть, но `enabled: false` |
| **Django** | Нет Go-реализации | `configs/config.json`, `config/config.go` | Нет | Только конфиг-структура |

**Влияние:** 312 ссылок на Keycloak в Go-коде, 315 — на Zitadel. Makefile `test-critical` ссылается на несуществующий `./internal/keycloak/...`. Архитектурные диаграммы в `openwiki/architecture.md` и `docs/architecture/ARCHITECTURE_ANALYSIS.md` показывают Keycloak как альтернативный провайдер, хотя он отключён.

**Рекомендация:** Полностью удалить Keycloak и Django из кода, конфигов и скриптов. Обновить архитектурные диаграммы.

### 3. Множественные пути деплоя

| Путь | Описание | Статус |
|---|---|---|
| `scripts/deploy/deploy.sh` | Основной скрипт деплоя | Актуальный (не изменён k8s-consol) |
| `k8s-manifests/00-15*.yaml` | Нумерованные манифесты | **Canonical prod deploy** — см. `k8s-manifests/README.md` |
| `k8s/deployments/`, `k8s/services/` | Манифесты по доменам | **Canonical ops** — актуальные образы/порты; drift с numbered manifests зафиксирован |
| `terraform-k8s/` | Провижининг K8s-кластера | Актуальный |
| `terraform/` | Альтернативный terraform | Дублирует terraform-k8s |
| `infra/` | Третий terraform-набор | Дублирует оба |
| `ansible/` | Ansible-плейбук | Минимальный (1 playbook) |

**Влияние:** Непонятно, какой путь деплоя canonical. Новый разработчик не может выбрать правильный.

### 4. Дублирование мониторинга

| Компонент | Где | Дубликаты |
|---|---|---|
| Prometheus | `prometheus/`, `prometheus_multiclusters/`, `monitoring/prometheus/`, `deploy/prometheus/` | 4 копии |
| Grafana | `monitoring/grafana/`, `deploy/grafana/`, `github_actions/grafana/` | 3 копии |
| AlertManager | `monitoring/alertmanager/`, `deploy/alertmanager/` | 2 копии |
| Zabbix | `zabbix-agent/`, `scripts/zabbix/` | 2 копии |
| Dashboards | `monitoring/dashboards/`, `github_actions/grafana/` | 2 копии |

### 5. Противоречия в документации

| Проблема | Где |
|---|---|
| `openwiki/architecture.md` описывает Keycloak как "Alternative OIDC provider" | Строка 153 |
| `docs/architecture/ARCHITECTURE_ANALYSIS.md` содержит Keycloak-диаграммы | Строки 341, 900, 2194-2225, 2306 |
| `docs/architecture/ARCHITECTURE_ANALYSIS.md` утверждает "Keycloak полностью удален" | Строка 1240 — **противоречит** строке 2194 |
| openwiki упоминает Django как провайдер | `openwiki/architecture.md` строка 83 |
| SERVICE_CATALOG: relay на порту 8080 | quickstart.md: relay на 9090 |
| SERVICE_CATALOG: TURN "roduction Ready" (опечатка) | Строка 109 |

### 6. Go version mismatch

`go.mod` указывал `go 1.25.1` — обновлено до **`1.25.12`** (последний патч линейки 1.25, релиз 2026-07-07). Доступна также линейка Go 1.26.5.

---

## Рекомендуемый план действий

### 1. Quick wins (1-2 дня)

- [x] **Удалить Keycloak-код:** `internal/auth/keycloak_enhanced.go`, `keycloak_enhanced_test.go`, `interfaces.go` (KeycloakAuthInterface) — *шаг 1, 2026-07-09*
- [x] **Удалить Keycloak-скрипты:** 8 скриптов в `scripts/setup/` и `scripts/maintenance/` (+ `setup_ldap_integration.sh`) — *шаг 3*
- [x] **Удалить Django-конфиг:** секции `django` в `configs/config.json` и `config/config.go` — *шаг 2* (dev-файл `config/config.json` не трогали)
- [x] **Удалить quantum-код:** `internal/quantum/`, `internal/metrics/quantum.go`, `internal/cadence/workflows/quantum.go` — *шаг 2*
- [x] **Удалить `src/`** — осиротевший TypeScript SDK — *шаг 2*
- [ ] **Удалить `systemd/enhanced-quic-proxy.service`** — идентичен `quic-proxy.service` — *не выполнено: diff не пустой (порты 9090 vs 9092), шаг 4*
- [x] **Исправить Makefile:** заменить `./internal/keycloak/...` на `./internal/zitadel/...` в `test-critical` — *шаг 1*
- [x] **Исправить go.version** в go.mod на `1.25.12` (выполнено 2026-07-08)
- [x] **Удалить три `.disabled` workflow** в `.github/workflows/` — *шаг 3*

### 2. Средний срок (1-2 недели)

- [x] **Инвентаризация дубликатов реализаций:** `docs/audit/DUPLICATES_INVENTORY.md`; удалён мёртвый `internal/nat/turn_server.go` (winner — `turn_server_enhanced.go`) — *шаг 1, 2026-07-09*
- [x] **Консолидировать `internal/network/` дубликаты:** `score_allocator_production.go` + `retry_config_improved.go`; удалены `score_tenant_network_allocator.go`, `score_tenant_network_allocator_improved.go`, `retry_config.go` — *шаг 1, 2026-07-09*
- [x] **Консолидировать AI-дубликаты:** `real_service` / `optimized_service` и exclusive `real_*` / `optimized_*` удалены; canonical — `ai.Service` + detector/load_predictor/optimizer/collectors — *cadence-ai step 3, 2026-07-09*
- [ ] **Консолидировать Cadence dual-path (merge API):** `cadence/client`, `cadence/service_client` — *intentional dual entrypoint relay/ai-worker; merge отложен (medium risk)*
- [~] **Объединить K8s-манифесты (first step):** canonical paths + README (`k8s/`, `k8s-manifests/`), архив 4× `*-fixed.yaml` → `archive/k8s/`, diff relay/AI/Cadence в `openwiki/port-scheme.md` + `CONSOLIDATION_PLAN §1` шаги 1–2 — *k8s-consol step 2–4, 2026-07-09; полный merge — отложено*
- [x] **P0: убрать nginx-stubs из `k8s-manifests/05–06` + `deploy.sh` load_images** — production images/ports, cadence с Docker Hub — *k8s-prod-fix steps 2–4, 2026-07-09*
- [ ] **Объединить Terraform:** `terraform/` + `infra/` → `terraform-k8s/` — *только план, шаг 4*
- [ ] **Объединить Prometheus/Grafana/Zabbix** в `monitoring/` — *только план, шаг 4*
- [ ] **Объединить `test/` + `tests/`** в `test/` — *только план, шаг 4*
- [ ] **Объединить Dockerfile-дубликаты** в `docker/`
- [x] **Синхронизировать портовую схему** во всех документах (canonical → `openwiki/port-scheme.md`, sync `openwiki/` + `docs/network/` + auxiliary IPAM/AI/Worker) — *score-ports step 2–4, 2026-07-09; верификация `go build` + port grep OK*
- [x] **Обновить архитектурные диаграммы** — Zitadel-only; Keycloak/Django только как legacy/removed — *шаг 3, 2026-07-09*
- [x] **Удалить `github_actions/`** (Cadence CI leftovers, не GitHub Actions; move в `docker/cadence/` отклонён — Dockerfile нерабочий) — *шаг 4, 2026-07-09*
- [x] **План консолидации инфраструктуры:** `docs/audit/CONSOLIDATION_PLAN.md` (k8s, terraform, monitoring, test/) — *шаг 4, 2026-07-09*
- [ ] **Очистить `docs/legacy/`** — оценить, можно ли удалить 142 файла

### 3. Стратегические (1+ месяц)

- [ ] **Разбить `cmd/relay/main.go`** (2391 строка) на модули
- [ ] **Разбить `internal/metrics/metrics.go`** (1833 строки) на per-subsystem файлы
- [ ] **Разбить `internal/ai/ai.go`** (1664 строки) на модули
- [x] **Ввести CI-проверку** `go vet`, `golangci-lint` в GitHub Actions — *`.github/workflows/lint.yml`, шаг 3, 2026-07-09; `go vet` на `cmd/relay` падает на legacy test (см. step03)*
- [ ] **Добавить интеграционные тесты** для auth-провайдера (Zitadel)
- [x] **Документировать canonical deployment path** — `openwiki/deploy.md` (terraform-k8s → deploy.sh + k8s-manifests → Zitadel); ссылки в quickstart/operations/AGENTS.md — *шаг 4, 2026-07-09*
- [ ] **Пересмотреть архитектуру internal/p2p/** — 30 файлов, много переплетений

---

## Приложение: команды и вывод

### Проверка Go

```bash
$ go version
/bin/bash: go: command not found
```

### Статистика кодовой базы

```bash
# Всего Go-строк: 174 902
# Internal-пакетов: 49 директорий
# cmd-бинарников: 13
# Тест-файлов: 16 (в 42+ директориях)
# Файлы с TODO/FIXME/HACK: 39
# Наибольшие TODO-файлы:
#   internal/api/protocol_handlers.go: 9
#   cmd/relay/main.go: 9
#   internal/relay/ai_middleware.go: 4
#   internal/zitadel/client.go: 3
```

### Структура cmd/

| Бинарник | Статус | Строк main.go |
|---|---|---|
| `relay` | Основной | 2391 |
| `api` | | — |
| `ai-service` | | — |
| `ai-analytics` | | — |
| `ai-worker` | | — |
| `cadence-service` | | — |
| `ipam-service` | | — |
| `cni-plugin` | | — |
| `quic-service` | Экспериментальный | — |
| `tenant-policy-controller` | | 3.9K |
| `kubectl-zitadel` | Утилита | — |
| `quantum-service` | Отключён (DISABLED.md) | 44 строки заглушки |
| `zerotrust-demo` | Демо | 4.4K |

### Крупнейшие Go-файлы

| Файл | Строк |
|---|---|
| `cmd/relay/main.go` | 2 391 |
| `internal/metrics/metrics.go` | 1 833 |
| `internal/ai/ai.go` | 1 664 |
| `cmd/ai-service/handlers.go` | 1 520 |
| `config/config.go` | 1 341 |
| `internal/p2p/manager_test.go` | 1 222 |
| `cmd/relay/masque_support.go` | 1 170 |
| `internal/p2p/relay_manager.go` | 1 155 |
| `internal/cadence/workflows/ai.go` | 1 069 |
| `internal/cadence/client_production.go` | 1 064 |

### Auth: баланс ссылок

```bash
# Keycloak-ссылки в Go-коде: 312
# Zitadel-ссылки в Go-коде: 315
# Django-ссылки в Go-коде: 0 (только config/config.go)
# Keycloak-скрипты в scripts/: 8 файлов
```

---

*Отчёт создан статическим анализом. Для полной версии рекомендуется запустить `go build ./...` и `go test ./... -short` в среде с установленным Go 1.25.12+.*
