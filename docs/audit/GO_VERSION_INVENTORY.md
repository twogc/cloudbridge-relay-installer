# Инвентаризация версий Go в репозитории

**Дата:** 2026-07-08  
**Каноническая версия проекта:** `go 1.25.12` (патч-релиз линейки 1.25, обновление 2026-07-07)  
**Альтернатива (новая линейка):** Go `1.26.5` (обновление 2026-07-07) — для миграции на 1.26, не используется в CloudBridge

> Предыдущая зафиксированная версия: `1.25.1` — устарела, требует обновления до `1.25.12`.

---

## go.mod (директива `go`)

| Файл | Было | Стало / рекомендация |
|------|------|----------------------|
| `go.mod` | `1.25.1` | **`1.25.12`** |
| `cmd/kubectl-zitadel/go.mod` | `1.25.1` | **`1.25.12`** |
| `pkg/enhanced-quic/go.mod` | `1.25.1` | **`1.25.12`** |
| `test/e2e/go.mod` | `1.25.1` | **`1.25.12`** |
| `tools/development/go.mod.backup.*` | `1.23.0` | Архив — не трогать |
| `voly/agentsview/go.mod` | `1.26.3` | Отдельный подпроект — не трогать |

---

## GitHub Actions (`go-version` / `GO_VERSION`)

| Файл | Было | Стало / рекомендация |
|------|------|----------------------|
| `.github/workflows/security-audit.yml` | `'1.25'` | **`'1.25.12'`** |
| `.github/workflows/release.yml` | `'1.23'` | **`'1.25.12'`** |
| `.github/workflows/main.yml.disabled` | `GO_VERSION: '1.25'` | **`'1.25.12'`** |
| `.github/workflows/test-coverage.yml.disabled` | `GO_VERSION: '1.25'` | **`'1.25.12'`** |
| `.github/workflows/security-enhanced.yml.disabled` | `GO_VERSION: '1.25'` | **`'1.25.12'`** |
| `pkg/enhanced-quic/.github/workflows/release.yml` | `'1.23'` | **`'1.25.12'`** |

---

## Docker (`FROM golang:…`)

### Точная версия `1.25.1` → `1.25.12`

| Файл |
|------|
| `docker/Dockerfile.tenant-policy-controller` |
| `docker/images/Dockerfile` (2 stage) |
| `docker/images/Dockerfile.ai` |
| `docker/images/Dockerfile.ai-worker` |
| `docker/images/Dockerfile.cadence` |
| `docker/images/Dockerfile.quic` |
| `docker/images/Dockerfile.quantum` |
| `docker/images/Dockerfile.relay.alpine` |
| `docker/images/Dockerfile.relay.fixed` |
| `docker/images/Dockerfile.wireguard` |
| `github_actions/Dockerfile` |
| `pkg/enhanced-quic/Dockerfile` |

### Плавающий тег `1.25-alpine` → `1.25.12-alpine`

| Файл |
|------|
| `docker/Dockerfile` |
| `docker/dockerfiles/Dockerfile.*` (5 файлов) |
| `docker/images/Dockerfile.*` (enhanced, production, relay, ai-service, ai-analytics, cni, ipam, quic, wireguard, cadence) |
| `openwiki/operations.md` (пример) |
| `docs/guides/VK_CLOUD_DEPLOYMENT_GUIDE.md` (пример) |

### Устаревшие версии (≠ 1.25.x)

| Файл | Было | Рекомендация |
|------|------|--------------|
| `scripts/migrate-to-kubernetes.sh` | `golang:1.23-alpine` (×5) | **`1.25.12-alpine`** |
| `scripts/setup/fix-vulnerabilities.sh` | `golang:1.21-alpine3.19` | **`1.25.12-alpine`** |
| `docs/legacy/SECURITY_AUDIT_WEEK2.md` | `golang:1.21` | Legacy — **`1.25.12`** при правке |
| `voly/agentsview/Dockerfile` | `golang:1.26.3-bookworm` | Подпроект voly — не трогать |

---

## Скрипты

| Файл | Было | Стало |
|------|------|-------|
| `scripts/update-dependencies.sh` | проверка `1.25.1` / `1.25.2` | **`1.25.12`** |
| `scripts/start-ai-analytics.sh` | «Go 1.21 или выше» | **Go 1.25.12+** |

---

## Документация (активная)

| Файл | Было | Стало |
|------|------|-------|
| `openwiki/quickstart.md` | Go 1.25.1+ | **Go 1.25.12+** |
| `docs/architecture/ARCHITECTURE_ANALYSIS.md` | 1.25.1 | **1.25.12** |
| `docs/analysis/CLIENT_COMPATIBILITY_ANALYSIS.md` | go1.25.1 | **go1.25.12** |
| `docs/analysis/TENANT_POLICY_CONTROLLER_ANALYSIS.md` | golang:1.25.1-alpine | **1.25.12-alpine** |
| `docs/analysis/ECOSYSTEM_SUMMARY.md` | Go 1.21+ | **Go 1.25.12+** |
| `docs/architecture/CLOUDBRIDGE_TCP_TUNNEL_TECHNICAL_SPEC.md` | Go 1.21+, CI 1.21 | **1.25.12** |
| `docs/audit/MIMO_AUDIT_REPORT.md` | «1.25.1 не существует» | Исправлено — версия валидна, обновлена до 1.25.12 |

### Legacy / архив (не обновлялись автоматически)

| Файл | Версия | Примечание |
|------|--------|------------|
| `docs/legacy/PERFORMANCE_BENCHMARK_REPORT.md` | Go 1.21+ | Архив |
| `docs/legacy/CLOUDBRIDGE_ARCHITECTURE_OVERVIEW.md` | Go 1.21+ | Архив |
| `docs/legacy/TESTING.md` | go-version: 1.21 | Архив |
| `docs/legacy/PROJECT_SUMMARY.md` | Go 1.25+ | OK (минор) |
| `docs/legacy/reliase.md` | Go 1.25.1 | Архив |
| `archive/docs/old-analysis/PROJECT_OVERVIEW.md` | Go 1.25.1 | Архив |
| `docs/roadmap/CLOUDBRIDGE_STRATEGIC_ROADMAP_v2.md` | Go 1.25 | OK (минор) |

---

## Подпроект `voly/` (вне scope CloudBridge)

| Компонент | Версия Go |
|-----------|-----------|
| `voly/agentsview/go.mod` | 1.26.3 |
| Docker / docs | 1.26+ |

---

## Рекомендации

1. **Для разработки и CI:** установить Go **1.25.12** (`https://go.dev/dl/go1.25.12.linux-amd64.tar.gz`).
2. **Для экспериментов с 1.26:** Go **1.26.5** совместим, но требует смены `go` в `go.mod` и прогона полного тест-сьюта.
3. **Docker:** предпочитать `golang:1.25.12-alpine` вместо `1.25-alpine` для воспроизводимых сборок.
