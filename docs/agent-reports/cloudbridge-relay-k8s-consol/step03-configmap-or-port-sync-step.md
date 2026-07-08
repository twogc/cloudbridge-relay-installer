# Отчёт агента — cursor: configmap or port sync step

**Дата:** 2026-07-09 02:02  
**Миссия:** cloudbridge-relay-k8s-consol  
**Шаг:** 3

## Выполнено

1. **Сверка портов и образов** между `k8s-manifests/05-deployments.yaml`, `06-services.yaml` и `k8s/deployments/`, `k8s/services/` для relay, AI и Cadence. Результат задокументирован ниже и в `openwiki/port-scheme.md` (новая секция §k8s drift).

2. **Документация drift (без правок prod YAML):**
   - `openwiki/port-scheme.md` — таблицы diff relay / AI / Cadence; уточнён пункт Known Documentation Drift для AI.
   - `k8s-manifests/README.md` — секция «Известный drift с k8s/» с краткой таблицей и ссылкой на port-scheme.
   - `docs/audit/CONSOLIDATION_PLAN.md` — шаг миграции 2 отмечен `[x]`.
   - `docs/agent-reports/cloudbridge-relay-k8s-consol/mimo-analysis.md` — пункт 2 плана миграции отмечен ✅.

3. **Решение по YAML:** numbered prod-манифесты (`k8s-manifests/05–06`) **не изменялись** — `nginx:latest` и порты-заглушки осознанные; однозначный AI drift (8080 vs 8083) зафиксирован в docs; актуальное ops-дерево — `k8s/`.

### Diff: CloudBridge Relay

| Поле | `k8s-manifests/05–06` | `k8s/deployments|services` | Canonical |
|------|----------------------|----------------------------|-----------|
| Image | `nginx:latest` | `cloudbridge-relay:mvp-20251010-110859` (+ socat sidecar) | built image |
| containerPort | 80 (stub) | 8080 TCP, 9090 UDP, 51820 UDP, 9091, 8444 | HTTP 9090 |
| Probes | 9090 `/health` | 9090 `/health` | 9090 |
| Service | 9090 TCP/UDP, 51820, 8444 | 9090 TCP/UDP, 51820, 8084 metrics | 9090, 51820, 8444 |
| prometheus.io/port | 8084 | 8084 | metrics 5551 |

Service-порты relay в `06-services.yaml` **совпадают** с canonical; deployment — stub.

### Diff: AI Service

| Поле | `k8s-manifests/05–06` | `k8s/deployments|services` | Canonical |
|------|----------------------|----------------------------|-----------|
| Image | `nginx:latest` | `ai-service:latest` (+ socat) | built image |
| containerPort | 80 (stub) | 8083, 8084 | AI_PORT 8083 |
| Probes | 8080 | 8083 | 8083 |
| Service | 8080, metrics 9090 | 8083, 8084 | code 8083 |

**Однозначный drift:** `k8s-manifests/06` AI Service 8080/9090 устарели; `k8s/services/ai-service.yaml` и код согласованы на 8083/8084.

### Diff: Cadence

| Поле | `k8s-manifests/05–06` | `k8s/deployments|services` |
|------|----------------------|----------------------------|
| Image | `nginx:latest` | `cadence-working.yaml` (отдельный) |
| Service | 7933, 7934, 7935, 7939 | 7933, 7934, 7935 |
| Probe | 7933 | varies |

Порты frontend/matching/history согласованы; 7939 (worker) только в numbered service.

## Изменённые файлы

| Путь | Действие |
|------|----------|
| `docs/agent-reports/cloudbridge-relay-k8s-consol/step03-configmap-or-port-sync-step.md` | создан (этот отчёт) |
| `openwiki/port-scheme.md` | добавлена секция §k8s drift, обновлён Known Documentation Drift |
| `k8s-manifests/README.md` | добавлена секция drift с `k8s/` |
| `docs/audit/CONSOLIDATION_PLAN.md` | шаг 2 миграции `[x]` |
| `docs/agent-reports/cloudbridge-relay-k8s-consol/mimo-analysis.md` | пункт 2 ✅ |

**Не изменялись:** `k8s-manifests/05-deployments.yaml`, `06-services.yaml`, `k8s/deployments/*`, `k8s/services/*`, `scripts/deploy/deploy.sh`.

## Проверки (команды и результат)

```bash
# 1. Relay images — stub vs real
rg 'image:' k8s-manifests/05-deployments.yaml k8s/deployments/cloudbridge-relay-deployment.yaml
# k8s-manifests → nginx:latest (×3)
# k8s/deployments → cloudbridge-relay:mvp-20251010-110859 (OK, drift зафиксирован)

# 2. AI service ports
rg 'port:|targetPort:' k8s-manifests/06-services.yaml k8s/services/ai-service.yaml
# k8s-manifests ai-service → 8080 / 9090
# k8s/services/ai-service.yaml → 8083 / 8084 (OK, drift зафиксирован)

# 3. CONSOLIDATION_PLAN step 2 marked done
rg 'Зафиксировать diff' docs/audit/CONSOLIDATION_PLAN.md
# → 2. [x] Зафиксировать diff … (OK)

# 4. port-scheme drift section exists
rg 'k8s-manifests.*vs.*k8s' openwiki/port-scheme.md
# → ## `k8s-manifests/` vs `k8s/` Manifest Drift (OK)

# 5. deploy.sh unchanged — still k8s-manifests/
rg 'k8s-manifests' scripts/deploy/deploy.sh | head -3
# → kubectl apply -f k8s-manifests/00-namespace.yaml … (OK)
```

## Блокеры / не сделано

| Пункт | Причина |
|-------|---------|
| Синхронизация YAML `k8s-manifests/05–06` с `k8s/` | Высокий риск для prod deploy path; требует тестового кластера и обновления образов в deploy.sh |
| Исправление AI Service 8080→8083 в `06-services.yaml` | Отложено — numbered manifests не трогали без уверенности |
| ConfigMap sync (`03-configmap.yaml` ↔ `k8s/configmaps/`) | Вне scope step 3 (следующий шаг) |
| Обновление `docs/operations/SERVICE_CATALOG.md` | Legacy drift уже отмечен в port-scheme; полный rewrite не запрошен |
| commit/push | Явно не запрошено |
