# Отчёт агента — cursor: github_actions + consolidation plan

**Дата:** 2026-07-09 01:33  
**Миссия:** cloudbridge-relay-medium  
**Шаг:** 4

## Выполнено

### 1. Анализ `github_actions/`

Прочитаны MIMO_AUDIT_REPORT.md §2, содержимое `github_actions/` (27 файлов), сравнение с `docker/`.

**Вывод:** директория — upstream **Cadence CI leftovers** (docker-compose для unit/integration тестов Cadence), а не GitHub Actions workflows CloudBridge.

| Критерий | Результат |
|----------|-----------|
| Ссылки из `scripts/deploy/` | 0 |
| Ссылки из operational scripts | 0 |
| Ссылки из Go-кода | 0 |
| Dockerfile | Копирует пути Cadence (`cmd/server/`, `common/archiver/gcloud/`) — **не собирается** в структуре CloudBridge |
| README | Указывает путь `docker/github_actions/` (ожидаемый путь в upstream cadence repo) |
| `docker/cadence/` | Не существует |
| `docker/Dockerfile.cadence*` | Отдельная цель — сборка образов CloudBridge, не дубль compose-стека |

**Решение: УДАЛИТЬ** (не перемещать в `docker/cadence/`).

Обоснование: перемещение добавило бы нерабочий compose-стек; содержимое не используется и не интегрируется с `scripts/deploy/deploy.sh`. Удаление безопаснее, чем move с битым Dockerfile.

### 2. План консолидации

Создан `docs/audit/CONSOLIDATION_PLAN.md` с разделами:

- **k8s/ vs k8s-manifests/** — canonical: `k8s-manifests/` для prod deploy (`scripts/deploy/deploy.sh`), `k8s/` для доменных/ops манифестов; шаги merge без изменения deploy.sh в этом шаге
- **terraform/ vs infra/ vs terraform-k8s/** — canonical: `terraform-k8s/` для K8s cluster; `terraform/` и `infra/` — archive/merge позже
- **test/ vs tests/** — canonical: `test/`; `tests/load/` и `wg_netns_integration.sh` → merge в `test/`
- **monitoring/** — целевая структура `monitoring/{prometheus,grafana,alertmanager,rules,zabbix}`; корневые `prometheus/`, `prometheus_multiclusters/`, `deploy/*` — к архивации

Полный merge k8s/terraform/monitoring **не выполнялся** (по заданию).

### 3. Безопасные no-op

- Удалена вся `github_actions/` (27 файлов) — единственное безопасное удаление в этом шаге
- Другие дубликаты (prometheus, deploy/grafana) **не удалялись** — diff не пустой, есть script references

## Изменённые файлы

| Путь | Действие |
|------|----------|
| `github_actions/` (27 файлов) | **Удалено** |
| `docs/audit/CONSOLIDATION_PLAN.md` | **Создан** |
| `docs/agent-reports/cloudbridge-relay-medium/step04-github-actions-consolidation-plan.md` | **Создан** (этот отчёт) |

## Проверки (команды и результат)

```bash
# Удаление github_actions
test ! -d github_actions && echo OK
# → OK

# deploy.sh не затронут
rg -c 'k8s-manifests' scripts/deploy/deploy.sh
# → 9 (все ссылки на k8s-manifests/ на месте)

bash -n scripts/deploy/deploy.sh
# → bash -n OK

# Ссылки на github_actions вне docs/
rg -l 'github_actions' --glob '!docs/**' .
# → (none outside docs)

# План создан
test -f docs/audit/CONSOLIDATION_PLAN.md && wc -l docs/audit/CONSOLIDATION_PLAN.md
# → 195 docs/audit/CONSOLIDATION_PLAN.md
```

```bash
# Инвентаризация дубликатов (для плана, без изменений)
diff -q prometheus/prometheus.yml prometheus_multiclusters/prometheus.yml
# → differ (4 строки scrape targets)

diff -q terraform/k8s-cluster.tf terraform-k8s/k8s-cluster.tf
# → differ (не идентичные копии)

ls test/ tests/
# → test/: e2e/, config-server-*.yaml
# → tests/: load/, wg_netns_integration.sh
```

## Блокеры / не сделано

| Пункт | Причина |
|-------|---------|
| Перемещение `github_actions/` → `docker/cadence/` | Отклонено: Dockerfile и compose нерабочие в контексте CloudBridge; move = мёртвый код |
| Merge `k8s/` + `k8s-manifests/` | Высокий риск для `scripts/deploy/deploy.sh`; отложено по заданию |
| Merge terraform/infra → terraform-k8s/ | Разные tfstate, различающиеся k8s-cluster.tf; нужен отдельный шаг с diff |
| Merge monitoring/prometheus/deploy | Alert rules и prometheus.yml не идентичны; требуется diff и обновление скриптов |
| Переименование `tests/` → `test/` | Только задокументировано в CONSOLIDATION_PLAN.md |
| Commit/push | Не запрошено |
