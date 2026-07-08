# step02 — active docs: ports + security + AGENTS

**Дата:** 2026-07-09  
**Миссия:** docs-consistency-sync (pass 2)

## Цель

Продолжить согласование **активных** docs (не `legacy/`) с `openwiki/port-scheme.md` и Zitadel-only IAM.

## Изменённые файлы

| Файл | Суть |
|------|------|
| `docs/security/README.md` | Переписан: Zitadel, убраны битые KEYCLOAK_*.md ссылки |
| `docs/network/INTERNET_PORTS.md` | SoT-aligned; metrics **5551** TCP; 9090 UDP=QUIC |
| `docs/network/INTERNET_PORTS_FULL.md` | То же; убран «metrics UDP 9090» |
| `docs/network/PORTS_SUMMARY.md` | Canonical table + исправлен NodePort/health/metrics |
| `docs/network/SUBDOMAINS_USAGE.md` | admin 9090; monitor 5551 |
| `docs/network/SUBDOMAINS_ANALYSIS.md` | Ingress YAML + ASCII diagram |
| `docs/architecture/ARCHITECTURE_ANALYSIS.md` | Service map: quantum disabled, cadence 7933, AI worker 8085 |
| `docs/README.md` | SoT table + audit/agent-reports |
| `AGENTS.md` | Миссии 1–7; quick wins; open backlog |
| `docs/audit/MIMO_AUDIT_REPORT.md` | §5 doc contradictions status |
| этот отчёт | step02 |

## Ключевые исправления портов

| Было (ошибка) | Стало (canonical) |
|---------------|-------------------|
| Metrics = 9090 UDP/TCP | Metrics = **5551** TCP |
| Health / Admin = 5551 | Health API often 5552/9090; **5551 = metrics only** |
| Admin subdomain → 5551 | Admin → **9090** HTTP API |
| Cadence :8082 in architecture | Cadence **7933–7939** |
| quantum-service Production | **Disabled** |

## Не трогали

- `docs/legacy/`, `docs/roadmap/` (исторические typo «roduction» — out of scope)
- Go code / k8s manifests
- Residual runtime drift AI Service metrics port 8084 (code/manifest issue, not this pass)

## Проверки

- Narrative: security README no Keycloak as active IdP
- Network files cite port-scheme as SoT
- AGENTS checklist matches MIMO post-missions status
