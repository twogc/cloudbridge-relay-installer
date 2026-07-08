# Отчёт агента — cursor: architecture docs IAM cleanup

**Дата:** 2026-07-09 01:33  
**Миссия:** cloudbridge-relay-medium  
**Шаг:** 3

## Выполнено

1. Просканированы `openwiki/` и `docs/architecture/` (без `docs/legacy/`) на утверждения, что Keycloak/Django — текущий IAM.
2. **Zitadel** зафиксирован как единственный canonical identity provider во всех архитектурных docs openwiki.
3. Keycloak/Django оставлены только как «removed / legacy» (явные пометки в `architecture.md`, `operations.md`; исторические записи об удалении в `ARCHITECTURE_ANALYSIS.md`).
4. В `docs/architecture/ARCHITECTURE_ANALYSIS.md` обновлены mermaid-диаграммы и примеры кода: убраны узлы Keycloak, связи `RELAY --> KEYCLOAK`, `USER_MGMT --> KEYCLOAK`; SaaS-примеры переведены на Zitadel API.
5. В `AGENTS.md` отмечены выполненные пункты среднесрочной миссии: портовая схема и IAM-cleanup docs. Quick wins уже были `[x]` (шаг 1).

### Найденные и исправленные места

| Файл | Было | Стало |
|------|------|-------|
| `openwiki/architecture.md` | Multi-provider auth (Zitadel, Keycloak, Django) | Zitadel-only + legacy note |
| `openwiki/quickstart.md` | `auth/` — Zitadel, Keycloak, JWT | Zitadel OIDC/JWT |
| `openwiki/operations.md` | Zitadel (primary) + Keycloak (alternative) | Zitadel canonical; Keycloak removed |
| `docs/architecture/ARCHITECTURE_ANALYSIS.md` | 4 диаграммы с Keycloak, TS-комментарии Keycloak | Zitadel-only diagrams и комментарии |

## Изменённые файлы

- `openwiki/architecture.md`
- `openwiki/quickstart.md`
- `openwiki/operations.md`
- `docs/architecture/ARCHITECTURE_ANALYSIS.md`
- `AGENTS.md`

## Проверки (команды и результат)

```bash
# Поиск оставшихся упоминаний Keycloak/Django в scope
rg -l 'Keycloak|Django' openwiki/ docs/architecture/
```

**Результат:** 3 файла; все упоминания — контекст «removed/legacy/historical»:

```
openwiki/architecture.md:83   — "(Keycloak/Django removed)"
openwiki/architecture.md:151  — "Removed (legacy): Keycloak and Django..."
openwiki/operations.md:311    — "Keycloak removed"
ARCHITECTURE_ANALYSIS.md:298  — changelog «Убраны ссылки на Keycloak»
ARCHITECTURE_ANALYSIS.md:1236 — «Keycloak полностью удален»
ARCHITECTURE_ANALYSIS.md:1323 — «Zero Keycloak dependencies»
```

```bash
# Убедиться, что docs/legacy/ не тронут
git diff --name-only docs/legacy/
```

**Результат:** пусто (без изменений).

```bash
# Проверка, что в коде auth нет Keycloak/Django
rg 'Keycloak|Django' internal/auth/ cmd/relay/ 2>/dev/null || true
```

**Результат:** совпадений нет (соответствует документации).

## Блокеры / не сделано

- **`docs/architecture/CLOUDBRIDGE_COMPLETE_ARCHITECTURE.md`**, **`CLOUDBRIDGE_TCP_TUNNEL_TECHNICAL_SPEC.md`**, **`MASQUE_USAGE_GUIDE.md`** — уже описывают только Zitadel; правки не потребовались.
- **`docs/legacy/`** — намеренно не изменялся (архив).
- **`docs/audit/CONSOLIDATION_PLAN.md`** — не создавался/не обновлялся (вне scope шага 3).
- Commit/push не выполнялись (по инструкции).
