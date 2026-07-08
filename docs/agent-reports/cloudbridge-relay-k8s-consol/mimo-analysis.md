# MiMo Analysis — K8s Consolidation (`k8s/` vs `k8s-manifests/`)

**Дата:** 2026-07-09  
**Источник:** [MIMO_AUDIT_REPORT.md](../../audit/MIMO_AUDIT_REPORT.md), [CONSOLIDATION_PLAN.md](../../audit/CONSOLIDATION_PLAN.md)  
**Миссия:** cloudbridge-relay-k8s-consol

---

## Executive Summary

В репозитории два параллельных дерева Kubernetes-манифестов:

| Директория | ~Файлов | Роль |
|------------|---------|------|
| `k8s-manifests/` | 17 | Нумерованные манифесты `00–15`, упорядоченный apply |
| `k8s/` | ~125 | Доменная структура: cadence, zitadel, monitoring-stack, bgp, security, rbac |

**Проблема:** неочевидно, какой путь canonical для production deploy. Новый разработчик видит дубли deployments/services в обоих деревьях.

**Решение (зафиксировано):** разделение ролей, не слепой merge.

---

## Canonical Paths

| Слой | Canonical | Обоснование |
|------|-----------|-------------|
| **Production deploy (ordered)** | `k8s-manifests/` | Единственный путь в `scripts/deploy/deploy.sh`; задокументирован в `AGENTS.md`, `openwiki/deploy.md`, `openwiki/quickstart.md` |
| **Component / ops manifests** | `k8s/` | Zitadel Helm values, Cadence schema jobs, BGP, Kyverno, cert-manager — не входят в нумерованный стек; 30+ скриптов в `scripts/` |

---

## Пересечение и дубли

Оба дерева содержат deployments/services для relay, AI, Cadence, Zitadel, monitoring — с разной гранулярностью:

- `k8s-manifests/` — «один файл = один слой» (namespace → PVC → secrets → deployments → services → ingress)
- `k8s/` — «один компонент = своя папка» (`deployments/`, `services/`, `zitadel/`, `cadence/`)

### Что мержить (medium/high risk — не в step 2)

1. Синхронизировать порты/образы: `k8s-manifests/05-deployments.yaml`, `06-services.yaml` ↔ `k8s/deployments/`, `k8s/services/`
2. ConfigMaps: `k8s-manifests/03-configmap.yaml` ↔ `k8s/configmaps/cloudbridge-config*.yaml` — canonical `k8s/configmaps/`
3. Zitadel: `k8s-manifests/09–15` ↔ `k8s/zitadel/` — canonical `k8s/zitadel/` (Helm values, ingress)

### Что архивировать (low risk — step 2)

- Устаревшие standalone YAML в корне `k8s/`:
  - `cloudbridge-relay-deployment-fixed.yaml`
  - `tenant-isolation-policies-fixed.yaml`
  - `tenant-isolation-alerts-fixed.yaml`
  - `calico-config-fixed.yaml`
- Дублирующие `fixed/final/simple` в `k8s/cadence/manifests/` (15+ файлов) — отложено до выбора одного рабочего deployment

---

## Проверка ссылок из `scripts/`

```bash
rg 'cloudbridge-relay-deployment-fixed|calico-config-fixed|tenant-isolation-policies-fixed|tenant-isolation-alerts-fixed' scripts/
# → 0 совпадений (только tenant-isolation-network-policy-fixed.yaml в cloudbridge-scalable-relay/, другой путь)
```

Standalone `*-fixed.yaml` в корне `k8s/` **не referenced** скриптами — безопасно для архивации.

Ссылки остаются только в legacy/docs (`docs/legacy/`, `docs/TENANT_ISOLATION_OPERATIONS_GUIDE.md`) — вне scope prod deploy.

---

## План миграции (полный)

1. ✅ Инвентаризация: `rg -l 'kubectl apply' scripts/` → таблица «скрипт → путь манифеста»
2. ✅ Зафиксировать diff `k8s-manifests/` vs `k8s/deployments|services/` для relay/AI/Cadence — [step03 report](step03-configmap-or-port-sync-step.md)
3. ⬜ Обновить `scripts/deploy/deploy.sh` на опциональный режим `k8s/overlays/production/` (будущее)
4. ⬜ Перенести уникальные файлы из `k8s-manifests/` в `k8s/` только после обновления deploy.sh
5. ⬜ Пометить `k8s-manifests/` deprecated в README, удалить через 1 релиз

**Не делать в step 2:** массовый merge деревьев, изменение `scripts/deploy/deploy.sh`.

---

## Low-risk actions (step 2)

| Действие | Статус |
|----------|--------|
| Архивировать `k8s/*-fixed.yaml` → `archive/k8s/standalone-root/` | Выполнено |
| Краткий README в `k8s/` с canonical pointer | Выполнено |
| Краткий README в `k8s-manifests/` | Выполнено |
| Merge `k8s-manifests/` → `k8s/` | **Не делать** |
| Изменение `deploy.sh` | **Не делать** |

---

## Open Items

- Diff relay/AI/Cadence между `k8s-manifests/` и `k8s/deployments|services/` (step 3+)
- Архивация 15+ cadence `*-fixed/*-simple` вариантов после выбора canonical deployment
- Обновление `docs/TENANT_ISOLATION_OPERATIONS_GUIDE.md` на актуальные пути (вне step 2)
