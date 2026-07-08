# Отчёт агента — cursor: canonical deploy path docs

**Дата:** 2026-07-09 01:40  
**Миссия:** cloudbridge-relay-next  
**Шаг:** 4

## Выполнено

1. Прочитаны `openwiki/operations.md`, `openwiki/quickstart.md`, `scripts/deploy/deploy.sh`, `docs/audit/CONSOLIDATION_PLAN.md`, `AGENTS.md`.
2. Создан **`openwiki/deploy.md`** — единый canonical путь деплоя:
   - **Шаг 1:** `terraform-k8s/` (VK Cloud K8s cluster)
   - **Шаг 2:** `scripts/deploy/deploy.sh` + `k8s-manifests/` (`00–08`) — core stack
   - **Шаг 3:** Zitadel IAM — `k8s-manifests/09–15` или `scripts/zitadel-setup.sh` + `k8s/zitadel/`
3. Явно помечены **non-canonical** альтернативы с ссылками на CONSOLIDATION_PLAN:
   - `k8s/` (кроме component/ops манифестов)
   - `terraform/`
   - `infra/`
4. Добавлены ссылки:
   - `openwiki/quickstart.md` — Path 1 и таблица Navigation → `deploy.md`
   - `AGENTS.md` — таблица «Ссылки», технический контекст, «Выбор подхода», «Документация»
   - `openwiki/operations.md` — баннер canonical path в начале секции Kubernetes Deployment
5. **`scripts/deploy/deploy.sh` не изменялся** (по инструкции шага).

## Изменённые файлы

| Путь | Действие |
|------|----------|
| `openwiki/deploy.md` | **создан** — canonical deploy path, mermaid-диаграмма, non-canonical таблица |
| `openwiki/quickstart.md` | ссылка на `deploy.md` в Path 1 и Navigation |
| `openwiki/operations.md` | баннер canonical path + ссылка на CONSOLIDATION_PLAN |
| `AGENTS.md` | обновлены таблицы деплоя и ссылок |
| `docs/agent-reports/cloudbridge-relay-next/step04-canonical-deploy-path-docs.md` | создан (этот отчёт) |

## Проверки (команды и результат)

```bash
test -f openwiki/deploy.md && echo "deploy.md: OK"
rg -l "CONSOLIDATION_PLAN" openwiki/deploy.md openwiki/operations.md AGENTS.md
rg "deploy\.md" openwiki/quickstart.md AGENTS.md openwiki/operations.md
rg "non-canonical|Non-canonical" openwiki/deploy.md
git diff --stat scripts/deploy/deploy.sh
wc -l openwiki/deploy.md
```

```
deploy.md: OK
AGENTS.md
openwiki/operations.md
openwiki/deploy.md
# deploy.md links found in quickstart.md, operations.md, AGENTS.md (6 matches)
# Non-canonical markers: k8s/, terraform/, infra/ with CONSOLIDATION_PLAN §1/§2 links
git diff --stat scripts/deploy/deploy.sh  → (empty, файл не менялся)
179 openwiki/deploy.md
```

## Блокеры / не сделано

- **Миграция infra-деревьев не выполнялась** — только документация (по CONSOLIDATION_PLAN и инструкции шага).
- **`terraform-k8s/README.md` отсутствует** — entry point описан в `openwiki/deploy.md`; отдельный README в terraform-k8s — задача будущего шага (CONSOLIDATION_PLAN §2).
- **`deploy.sh` не применяет Zitadel (09–15)** — задокументировано явно; объединение Zitadel в deploy.sh — вне scope step 4.
- Commit/push **не выполнялись** (по правилам).
