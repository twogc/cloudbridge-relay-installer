# Отчёт агента — cursor: write mimo report + safe cleanup

**Дата:** 2026-07-09 02:02  
**Миссия:** cloudbridge-relay-k8s-consol  
**Шаг:** 2

## Выполнено

1. **Сохранён анализ MiMo** в `docs/agent-reports/cloudbridge-relay-k8s-consol/mimo-analysis.md` — выжимка из `MIMO_AUDIT_REPORT.md` и `CONSOLIDATION_PLAN.md` §1: canonical paths (`k8s-manifests/` для prod deploy, `k8s/` для ops), пересечения, low/medium-risk план, проверка ссылок из `scripts/`.

2. **Безопасная уборка (low-risk):** четыре устаревших standalone `*-fixed.yaml` из корня `k8s/` перенесены в `archive/k8s/standalone-root/`:
   - `calico-config-fixed.yaml`
   - `cloudbridge-relay-deployment-fixed.yaml`
   - `tenant-isolation-alerts-fixed.yaml`
   - `tenant-isolation-policies-fixed.yaml`  
   Перенос выполнен только после проверки: `rg scripts/` не ссылается на эти файлы.

3. **README с canonical pointer:**
   - `k8s/README.md` — добавлена секция «Canonical deployment» в начало; указано, что prod deploy через `k8s-manifests/` + `deploy.sh`, а `k8s/` — для component/ops.
   - `k8s-manifests/README.md` — создан краткий README: порядок apply, ссылка на `deploy.sh` и `openwiki/deploy.md`.

4. **Ограничения соблюдены:** `scripts/deploy/deploy.sh` не изменялся; деревья `k8s-manifests/` → `k8s/` не перемещались.

## Изменённые файлы

| Путь | Действие |
|------|----------|
| `docs/agent-reports/cloudbridge-relay-k8s-consol/mimo-analysis.md` | создан |
| `docs/agent-reports/cloudbridge-relay-k8s-consol/step02-write-mimo-report-safe-cleanup.md` | создан (этот отчёт) |
| `k8s/README.md` | обновлён (canonical pointer) |
| `k8s-manifests/README.md` | создан |
| `k8s/calico-config-fixed.yaml` | удалён из корня → архив |
| `k8s/cloudbridge-relay-deployment-fixed.yaml` | удалён из корня → архив |
| `k8s/tenant-isolation-alerts-fixed.yaml` | удалён из корня → архив |
| `k8s/tenant-isolation-policies-fixed.yaml` | удалён из корня → архив |
| `archive/k8s/standalone-root/calico-config-fixed.yaml` | добавлен |
| `archive/k8s/standalone-root/cloudbridge-relay-deployment-fixed.yaml` | добавлен |
| `archive/k8s/standalone-root/tenant-isolation-alerts-fixed.yaml` | добавлен |
| `archive/k8s/standalone-root/tenant-isolation-policies-fixed.yaml` | добавлен |

## Проверки (команды и результат)

```bash
# 1. scripts/ не ссылается на archived *-fixed.yaml
rg 'cloudbridge-relay-deployment-fixed|calico-config-fixed|tenant-isolation-policies-fixed|tenant-isolation-alerts-fixed' scripts/
# → 0 совпадений (OK)

# 2. В корне k8s/ не осталось *-fixed.yaml
ls k8s/*-fixed.yaml
# → No such file or directory (OK)

# 3. Архив содержит 4 файла
ls archive/k8s/standalone-root/
# → calico-config-fixed.yaml, cloudbridge-relay-deployment-fixed.yaml,
#   tenant-isolation-alerts-fixed.yaml, tenant-isolation-policies-fixed.yaml (OK)

# 4. README на месте
test -f k8s/README.md && test -f k8s-manifests/README.md
# → OK

# 5. deploy.sh не тронут — по-прежнему использует k8s-manifests/
rg 'k8s-manifests' scripts/deploy/deploy.sh | head -5
# → kubectl apply -f k8s-manifests/00-namespace.yaml … (OK)
```

## Блокеры / не сделано

| Пункт | Причина |
|-------|---------|
| Архивация `k8s/cadence/manifests/*-fixed/*-simple` (15+ файлов) | Medium risk — нужен выбор одного рабочего deployment; вне low-risk scope step 2 |
| Архивация `k8s/cloudbridge-relay-deployment-backup-20251010-203045.yaml` | Не входит в явный список `*-fixed.yaml`; скрипты создают backup в `k8s/deployments/`, не в корне |
| Diff `k8s-manifests/` vs `k8s/deployments|services/` | Запланировано на step 3+ |
| Обновление `docs/TENANT_ISOLATION_OPERATIONS_GUIDE.md` (ссылки на archived paths) | Вне scope step 2; legacy docs всё ещё указывают на старые пути |
| Merge деревьев, изменение `deploy.sh` | Явно запрещено в step 2 |
