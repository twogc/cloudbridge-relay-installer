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

**Остаётся:** SCORE wiring, Cadence merge, k8s full merge, ops/mws 8084/8080 legacy, main.go vet context leak, `docs/legacy/`, roadmap typos.

---

## Pass 3 (step03) — residual port fix

| Item | Result |
|------|--------|
| Relay prometheus **8084** → **5551** | ✅ `k8s-manifests/05–06` |
| AI Service metrics **8084** | ✅ removed; **8083** only |
| Root README ports | ✅ aligned |
| go build relay/ai-service | ✅ |
| go vet cmd/relay | pre-existing cancel leak (not KeycloakURL) |

Отчёт: [step03-residual-port-fix.md](step03-residual-port-fix.md)

---

## Pass 4 (step04) — ops/mws + vet

| Item | Result |
|------|--------|
| Ops relay deployment/services/configmaps | ✅ metrics **5551**, http **9090** |
| mws OVA minimal + docker-compose | ✅ 5551 / health 9090 |
| Grafana datasource | ✅ :5551 |
| `go vet ./cmd/relay/` Cadence cancel leak | ✅ fixed (`cancel()` on fail paths) |
| `go build ./cmd/relay/` | ✅ |

Отчёт: [step04-ops-mws-vet.md](step04-ops-mws-vet.md)

**Ещё open:** SCORE wiring, p2p-deployment lab ports, Cadence dual-path merge, full k8s tree merge.
