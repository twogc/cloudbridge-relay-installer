# SUMMARY — docs-consistency-sync

**Дата:** 2026-07-09  
**Цель:** Устранить противоречия и «пересечения-как-текущий-статус» между отчётами в `docs/audit/` и `docs/agent-reports/`.

---

## Проблема

- `MIMO_AUDIT_REPORT.md`: чеклист обновлён миссиями, **тело findings** всё ещё описывало Keycloak/quantum/nginx stubs / identical systemd / github_actions как живые TODO.
- Intermediate **SUMMARY** (medium, next, k8s-consol, score-ports, fixes) держали open items, уже закрытые следующими миссиями.
- `agent-reports/README.md` ссылался на несуществующую `cloudbridge-relay-audit/`.
- `SERVICE_CATALOG`: typo `roduction Ready`.

---

## Что сделано

| Файл | Изменение |
|------|-----------|
| `docs/audit/MIMO_AUDIT_REPORT.md` | Переписан как living report: snapshot vs status, ✅/⏳ tables, systemd 9090≠9092, mission index, SoT order |
| `docs/agent-reports/README.md` | 7 миссий + docs-consistency-sync; ghost audit folder removed; SoT rules |
| `cloudbridge-relay-fixes/SUMMARY.md` | Superseded block |
| `cloudbridge-relay-medium/SUMMARY.md` | Superseded + next steps annotated |
| `cloudbridge-relay-next/SUMMARY.md` | Superseded (ports/AI/k8s stubs) |
| `cloudbridge-relay-k8s-consol/SUMMARY.md` | nginx stubs → k8s-prod-fix |
| `cloudbridge-relay-score-ports/SUMMARY.md` | 06 stubs → k8s-prod-fix |
| `docs/operations/SERVICE_CATALOG.md` | `roduction` → `Production` (2×) |

Код / manifests **не** менялись.

---

## Канон после sync

| Тема | SoT |
|------|-----|
| Ports | `openwiki/port-scheme.md` — IPAM 8080, AI 8083, Worker 8085 |
| Deploy | `openwiki/deploy.md` — deploy.sh → k8s-manifests |
| Auth | Zitadel only |
| K8s prod images | k8s-prod-fix (no nginx stubs) |
| Open backlog | MIMO checklist ⏳ items + residual drift in k8s-prod-fix |

---

## Проверки

- Согласованность narrative: intermediate SUMMARYs → Superseded; MIMO body ↔ checklist.
- Typo catalog fixed via replace_all.

Commit/push: не выполнялись.

---

## Pass 2 (step02) — active docs

| Область | Результат |
|---------|-----------|
| `docs/security/README.md` | Zitadel-only; мёртвые KEYCLOAK links removed |
| `docs/network/*` | metrics **5551** TCP; QUIC **9090** UDP; admin **9090** TCP |
| `ARCHITECTURE_ANALYSIS` service map | quantum disabled; cadence 7933; ports → port-scheme |
| `docs/README.md` + `AGENTS.md` | SoT table; missions index; updated backlog |
| Отчёт | [step02-active-docs-ports-security.md](step02-active-docs-ports-security.md) |

**Остаётся (не docs-only):** SCORE wiring, Cadence merge, k8s full merge, AI metrics 8084 drift, `docs/legacy/` archive, roadmap typos.
