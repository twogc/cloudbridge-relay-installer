# Consolidation Plan — Infrastructure & Test Directories

**Дата:** 2026-07-09  
**Источник:** [MIMO_AUDIT_REPORT.md §2](MIMO_AUDIT_REPORT.md)  
**Статус:** План (миграция не выполнена в step 4)

Цель — единый canonical путь деплоя и наблюдаемости без поломки `scripts/deploy/deploy.sh`.

---

## 1. `k8s/` vs `k8s-manifests/`

### Текущее состояние

| Директория | Файлов | Роль | Кто использует |
|------------|--------|------|----------------|
| `k8s-manifests/` | 17 | Нумерованные манифесты `00–15`, упорядоченный apply | **`scripts/deploy/deploy.sh`** (canonical prod deploy) |
| `k8s/` | ~125 | Доменная структура: cadence, zitadel, monitoring-stack, bgp, security, rbac | 30+ скриптов в `scripts/` (setup, monitoring, zitadel, cadence) |

Пересечение: оба содержат deployments/services для relay, AI, Cadence, Zitadel, monitoring — но с разной гранулярностью. `k8s-manifests/` — «один файл = один слой», `k8s/` — «один компонент = своя папка».

### Canonical

| Слой | Canonical | Обоснование |
|------|-----------|-------------|
| **Production deploy (ordered)** | `k8s-manifests/` | Единственный путь в `scripts/deploy/deploy.sh`; задокументирован в `AGENTS.md`, `openwiki/quickstart.md` |
| **Component / ops manifests** | `k8s/` | Zitadel Helm values, Cadence schema jobs, BGP, Kyverno, cert-manager — не входят в нумерованный стек |

### Что мержить

1. Синхронизировать порты/образы между `k8s-manifests/05-deployments.yaml`, `06-services.yaml` и соответствующими файлами в `k8s/deployments/`, `k8s/services/` (не слепой merge — diff и выбор актуальной версии).
2. ConfigMaps: `k8s-manifests/03-configmap.yaml` ↔ `k8s/configmaps/cloudbridge-config*.yaml` — оставить один источник (`k8s/configmaps/`), генерировать или копировать в manifest overlay.
3. Zitadel: `k8s-manifests/09–15` ↔ `k8s/zitadel/` — canonical `k8s/zitadel/` (Helm values, ingress); numbered manifests — thin wrappers или удалить после миграции deploy.sh.

### Что архивировать

- Дублирующие «fixed/final/simple» варианты в `k8s/cadence/manifests/` (15+ файлов) → `archive/k8s/cadence/` после выбора одного рабочего deployment.
- Устаревшие standalone YAML в корне `k8s/` (`cloudbridge-relay-deployment-fixed.yaml`, `tenant-isolation-*-fixed.yaml`) — в архив, если не referenced скриптами.

### Шаги миграции (короткие)

1. [x] Инвентаризация: `rg -l 'kubectl apply' scripts/` → таблица «скрипт → путь манифеста».
2. [x] Зафиксировать diff `k8s-manifests/` vs `k8s/deployments|services/` для relay/AI/Cadence — см. [step03 report](../agent-reports/cloudbridge-relay-k8s-consol/step03-configmap-or-port-sync-step.md), `openwiki/port-scheme.md` §k8s drift.
3. [x] **P0 stub fix (2026-07-09, k8s-prod-fix):** `k8s-manifests/05–06` — убраны `nginx:latest`/port 80/AI 8080 stubs; `deploy.sh` → `load_images()` собирает relay/ai; см. [step02](../agent-reports/cloudbridge-relay-k8s-prod-fix/step02-write-mimo-fix-05-06-manifests.md), [step03](../agent-reports/cloudbridge-relay-k8s-prod-fix/step03-fix-deploy-sh-load-images.md), [SUMMARY](../agent-reports/cloudbridge-relay-k8s-prod-fix/SUMMARY.md).
3b. [x] **Ops ports sync (2026-07-09):** relay/AI metrics → 5551/8083 in `k8s/deployments|services|configmaps`; mws OVA aligned.
3c. [x] **Cadence manifests cleanup (2026-07-09):** `fixed/simple/clean` deployment variants → `archive/k8s/cadence/manifests-variants/`; keep `k8s/cadence/cadence.yaml` + schema jobs.
4. [x] `k8s/overlays/production/` + `deploy.sh --use-overlay` / `--render-overlay` / `--dry-run` (2026-07-09). Default remains file-by-file manifests. E2E on live cluster still recommended before making overlay the default.
5. Перенести уникальные файлы из `k8s-manifests/` в `k8s/` только после обновления deploy.sh.
6. Пометить `k8s-manifests/` deprecated в README, удалить через 1 релиз.

**Не делать в step 4:** массовый merge деревьев, изменение `scripts/deploy/deploy.sh`.

---

## 2. `terraform/` vs `infra/` vs `terraform-k8s/`

### Текущее состояние

| Директория | Объём | Назначение | Статус |
|------------|-------|------------|--------|
| **`terraform-k8s/`** | 5 файлов | VK Cloud **Kubernetes cluster** (`vkcs_kubernetes_cluster`) | **Canonical** — `README.md`, `openwiki/`, `scripts/deploy/` |
| `terraform/` | 36 файлов | VM + DDoS + Container Registry + environments (prod/staging) + modules | Legacy/alternate; частично дублирует k8s-cluster.tf |
| `infra/` | 3 файла | Одиночная VM `relay` (compute instance) | Минимальный PoC, не связан с K8s deploy |

`terraform/k8s-cluster.tf` и `terraform-k8s/k8s-cluster.tf` **различаются** (не идентичные копии).

### Canonical

| Задача | Canonical |
|--------|-----------|
| K8s cluster provisioning | **`terraform-k8s/`** |
| VM / registry / DDoS (если нужны) | Будущий `terraform/modules/` под `terraform-k8s/` или `terraform/legacy/` |
| Single-VM relay | `infra/` → merge как `terraform-k8s/modules/bare-metal-relay/` или архив |

### Что мержить

1. Уникальные ресурсы из `terraform/k8s-cluster.tf` → сверить с `terraform-k8s/k8s-cluster.tf`, перенести отличия.
2. `terraform/modules/cloudbridge/` → `terraform-k8s/modules/cloudbridge/` (если модуль актуален).
3. `terraform/environments/{prod,staging}/` → `terraform-k8s/environments/` с workspaces или tfvars.
4. `infra/main.tf` (single VM) → опциональный submodule; не смешивать state с K8s cluster.

### Что архивировать

- `terraform/*.md` (AUTOMATION_*, CLEANUP_*, GITHUB_SECRETS_*) → `docs/legacy/terraform/` или объединить в один guide.
- `terraform/terraform.tfstate`, `terraform-k8s/terraform.tfstate` — **не коммитить**; добавить в `.gitignore` при миграции.
- `terraform/scripts/deploy.sh` — проверить на дублирование `scripts/deploy/deploy.sh`; архивировать если orphan.

### Шаги миграции (короткие)

1. Документировать в `terraform-k8s/README.md` единственный entry point (`terraform init && apply`).
2. Diff `terraform/k8s-cluster.tf` vs `terraform-k8s/k8s-cluster.tf` — перенести недостающие data sources / outputs.
3. Переместить `terraform/` → `archive/terraform-vm/` (не удалять сразу).
4. Переместить `infra/` → `terraform-k8s/examples/single-vm/` или архив.
5. Обновить ссылки в `openwiki/quickstart.md` (убрать `terraform/` как равноправный путь).

**Не делать в step 4:** перемещение terraform-деревьев, изменение state.

---

## 3. `test/` vs `tests/`

### Текущее состояние

| Путь | Содержимое |
|------|------------|
| `test/` | `e2e/`, `config-server-a.yaml`, `config-server-b.yaml` — e2e / multi-server configs |
| `tests/` | `load/*.js` (k6/load scripts), `wg_netns_integration.sh` |
| `internal/tests/` | Go: `vpn_integration_test.go` (остаётся в пакете) |

Директории **не дублируют** друг друга по содержимому; разнесены исторически.

### Canonical

**`test/`** — единый корень для non-Go интеграционных и нагрузочных тестов (соответствует аудиту и Go-конвенции `test/` vs `*_test.go`).

### Целевая структура

```
test/
├── e2e/              # существующее
├── load/             # из tests/load/
├── integration/      # wg_netns_integration.sh и подобное
└── fixtures/         # config-server-*.yaml
```

### Шаги миграции (короткие)

1. `git mv tests/load test/load`
2. `git mv tests/wg_netns_integration.sh test/integration/`
3. `git mv test/config-server-*.yaml test/fixtures/` (опционально)
4. Обновить ссылки в Makefile, CI, README.
5. Удалить пустую `tests/`.

**Не делать в step 4:** перемещение (только план).

---

## 4. Monitoring (Prometheus / Grafana / Zabbix)

### Текущие копии

| Компонент | Локации | Примечание |
|-----------|---------|------------|
| Prometheus config | `monitoring/prometheus/`, `prometheus/`, `prometheus_multiclusters/`, `deploy/prometheus/`, `k8s/monitoring/`, `k8s/monitoring-stack/` | `prometheus/` vs `prometheus_multiclusters/` — 4 строки scrape targets |
| Alert rules | `monitoring/rules/`, `prometheus/alerts/`, `deploy/prometheus/alerts/` | Разные наборы алертов (не идентичны) |
| Grafana | `monitoring/grafana/`, `deploy/grafana/` (7 dashboards), ~~`github_actions/grafana/`~~ | github_actions удалён step 4 |
| AlertManager | `monitoring/alertmanager/`, `deploy/alertmanager/` | Сверить diff |
| Zabbix | `zabbix-agent/`, `scripts/zabbix/`, `scripts/config/zabbix-agent.d/` | Agent install vs metric scripts vs conf drops |
| K8s monitoring | `k8s/monitoring/`, `k8s/monitoring-stack/` | In-cluster deployments (остаются в `k8s/`) |

### Целевая структура `monitoring/`

```
monitoring/
├── prometheus/
│   ├── prometheus.yml          # single base config
│   └── overlays/
│       └── multicluster.yml    # из prometheus_multiclusters/
├── alertmanager/
│   └── alertmanager.yml
├── grafana/
│   ├── provisioning/           # datasources, dashboards.yml
│   └── dashboards/             # все JSON (из deploy/grafana + monitoring/dashboards)
├── rules/
│   ├── cloudbridge-alerts.yml
│   └── ai-alerts.yml
└── zabbix/
    ├── agent/                  # из zabbix-agent/
    ├── collectors/             # из zabbix-agent/collectors + scripts/zabbix/
    └── conf.d/                 # из scripts/config/zabbix-agent.d/
```

**K8s-specific** (`k8s/monitoring/`, `k8s/monitoring-stack/`) остаются в `k8s/` — ссылаются на ConfigMaps/Deployments, не на bare-metal compose.

### Шаги миграции (короткие)

1. Diff всех `prometheus.yml` — выбрать base в `monitoring/prometheus/`, остальное в overlays.
2. Merge alert rules в `monitoring/rules/`, удалить `prometheus/alerts/`, `deploy/prometheus/`.
3. Merge Grafana dashboards в `monitoring/grafana/dashboards/`, удалить `deploy/grafana/`.
4. `git mv zabbix-agent/* monitoring/zabbix/agent/`; merge `scripts/zabbix/` → `monitoring/zabbix/collectors/`.
5. Удалить корневые `prometheus/`, `prometheus_multiclusters/`, `deploy/prometheus|grafana|alertmanager/` после обновления ссылок в скриптах.
6. Обновить `scripts/optimize-prometheus.sh`, `scripts/deploy-monitoring-stack.sh` на новые пути.

**Не делать в step 4:** merge monitoring-деревьев.

---

## 5. Выполнено в step 4

| Действие | Решение |
|----------|---------|
| `github_actions/` | **Удалено** — upstream Cadence CI docker-compose; Dockerfile ссылается на `cmd/server/`, `common/archiver/` (структура Cadence, не CloudBridge); ни один operational script не использует; `docker/` уже содержит `Dockerfile.cadence*` для сборки образов |

---

## 6. Порядок выполнения (приоритет)

1. ~~`github_actions/` cleanup~~ ✅ step 4
2. Monitoring consolidation (низкий риск, много дубликатов)
3. `tests/` → `test/` rename
4. Terraform archive + `terraform-k8s/` README
5. K8s merge (высокий риск — последним, с тестовым кластером)

---

## 7. Критерии готовности

- [ ] `scripts/deploy/deploy.sh` проходит без изменений путей до явной миграции
- [ ] Один документированный terraform entry: `terraform-k8s/`
- [ ] Один корень `test/`, нет `tests/`
- [ ] Один корень `monitoring/` для bare-metal/compose; K8s manifests ссылаются на те же dashboard JSON через ConfigMap или URL
- [ ] `rg 'github_actions'` — только исторические audit-отчёты
