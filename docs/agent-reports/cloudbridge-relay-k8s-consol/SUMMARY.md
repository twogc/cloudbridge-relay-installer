# SUMMARY — cloudbridge-relay-k8s-consol

**Дата:** 2026-07-09  
**Миссия:** Консолидация `k8s/` vs `k8s-manifests/` (first step, без merge деревьев)  
**Статус:** ✅ Step 4/4 завершён

---

## Цель

Устранить неопределённость «какой K8s-путь canonical», не ломая production deploy через `scripts/deploy/deploy.sh`.

## Решение

**Разделение ролей**, не слепой merge:

| Директория | Роль | Canonical для |
|------------|------|---------------|
| `k8s-manifests/` (00–15) | Ordered prod stack | `scripts/deploy/deploy.sh` |
| `k8s/` (~125 файлов) | Domain/ops manifests | 30+ скриптов (Zitadel, Cadence, BGP, Kyverno…) |

## Выполнено по шагам

| Шаг | Содержание | Отчёт |
|-----|------------|-------|
| 2 | MiMo-анализ, архив 4× `*-fixed.yaml`, README в `k8s/` и `k8s-manifests/` | [step02](step02-write-mimo-report-safe-cleanup.md) |
| 3 | Diff relay/AI/Cadence, drift в `openwiki/port-scheme.md`, без правок prod YAML | [step03](step03-configmap-or-port-sync-step.md) |
| 4 | Верификация `deploy.sh`, обновление `MIMO_AUDIT_REPORT.md`, SUMMARY | [step04](step04-verify-summary.md) |

## Ключевые артефакты

- `docs/agent-reports/cloudbridge-relay-k8s-consol/mimo-analysis.md` — выжимка canonical paths и плана
- `k8s/README.md` — pointer на prod path
- `k8s-manifests/README.md` — порядок apply + drift table
- `docs/audit/CONSOLIDATION_PLAN.md` §1 — шаги миграции 1–2 `[x]`
- `openwiki/port-scheme.md` — секция §k8s drift
- `archive/k8s/standalone-root/` — 4 archived `*-fixed.yaml`

## Верификация (step 4)

```bash
git diff scripts/deploy/deploy.sh          # пусто — OK
rg -n 'kubectl apply' scripts/deploy/      # deploy.sh → k8s-manifests/* — OK
```

`setup-cadence.sh` по-прежнему использует `k8s/cadence/cadence.yaml` — ожидаемо (ops path).

## Не сделано (на момент k8s-consol)

| Пункт | Причина |
|-------|---------|
| Merge `k8s-manifests/` → `k8s/` | High risk; тестовый кластер + deploy.sh |
| Синхронизация YAML `05–06` с `k8s/deployments|services/` | *Тогда* prod = nginx stubs; drift задокументирован |
| ConfigMap sync (`03-configmap.yaml` ↔ `k8s/configmaps/`) | Следующая фаза |
| Архивация 15+ cadence `*-fixed/*-simple` | Medium risk |
| Обновление legacy docs | Вне scope |
| commit/push | Не запрошено |

### Superseded by later missions (2026-07-09)

| Open item (k8s-consol) | Статус |
|------------------------|--------|
| «Prod intentionally stub (`nginx:latest`)» | **Устарело:** **k8s-prod-fix** убрал nginx/port 80/AI 8080 stubs; real images + AI **8083** |
| AI Service 8080→8083 в `06-services.yaml` | ✅ **k8s-prod-fix** (Service metrics **8084** residual) |
| Full merge / ConfigMap / cadence variants | **ещё open** (`CONSOLIDATION_PLAN`) |

## Следующие шаги (рекомендация)

1. ConfigMap diff → единый SoT (`k8s/configmaps/`) — open
2. Optional `k8s/overlays/production/` в deploy.sh — open
3. ~~AI 8080→8083 в 06~~ → **k8s-prod-fix**
4. Архивация cadence variants — open
5. Residual: AI metrics **8084**, relay prometheus annotation — см. k8s-prod-fix SUMMARY

---

*Отчёты: `docs/agent-reports/cloudbridge-relay-k8s-consol/step02–step04`.*
