# Отчёты агентов (VOLY / Cursor / MiMo)

Отдельная папка для **отчётов о выполненной работе** агентов. Не путать с `docs/audit/` (аудит-анализ / living checklist).

**SoT при расхождении:** `openwiki/port-scheme.md`, `openwiki/deploy.md`, чеклист в `docs/audit/MIMO_AUDIT_REPORT.md` > SUMMARY с блоком **Superseded** > journal step-отчёты > `docs/legacy/`.

## Структура

```
docs/agent-reports/
├── README.md
├── voly-runs/                          # сводки оркестратора VOLY
│   └── YYYYMMDD_HHMMSS_2gc-<mission>.md
├── cloudbridge-relay-fixes/            # Keycloak/Django/quantum/src cleanup
├── cloudbridge-relay-medium/           # port-scheme, IAM docs, inventory, plan
├── cloudbridge-relay-next/             # score/retry, catalog, lint, deploy.md
├── cloudbridge-relay-cadence-ai/       # AI merge; Cadence dual-path
├── cloudbridge-relay-k8s-consol/       # K8s roles (prod vs ops)
├── cloudbridge-relay-score-ports/      # auxiliary ports = cmd/* + k8s ops
├── cloudbridge-relay-k8s-prod-fix/     # P0: nginx stubs → real images/ports
└── docs-consistency-sync/              # согласование противоречий в отчётах
```

Аудит-анализ пишется в `docs/audit/MIMO_AUDIT_REPORT.md` (не в agent-reports).  
Папки `cloudbridge-relay-audit/` **нет** — audit runs только в `voly-runs/*audit*`.

## Миссии (хронология 2026-07-09)

| # | Миссия | SUMMARY |
|---|--------|---------|
| 1 | fixes | [SUMMARY](cloudbridge-relay-fixes/SUMMARY.md) |
| 2 | medium | [SUMMARY](cloudbridge-relay-medium/SUMMARY.md) |
| 3 | next | [SUMMARY](cloudbridge-relay-next/SUMMARY.md) |
| 4 | cadence-ai | [SUMMARY](cloudbridge-relay-cadence-ai/SUMMARY.md) |
| 5 | k8s-consol | [SUMMARY](cloudbridge-relay-k8s-consol/SUMMARY.md) |
| 6 | score-ports | [SUMMARY](cloudbridge-relay-score-ports/SUMMARY.md) |
| 7 | k8s-prod-fix | [SUMMARY](cloudbridge-relay-k8s-prod-fix/SUMMARY.md) |

Промежуточные SUMMARY содержат блок **Superseded by later missions** — open items на момент миссии не равны текущему backlog.

## Правила

Каждый шаг combat-миссии **обязан** создать markdown в подпапке миссии:

- дата и номер шага
- список изменённых файлов
- что сделано и зачем
- результаты проверок (`go build`, `go test`, grep)
- блокеры и невыполненные пункты
- при закрытии open item **последующей** миссией — обновлять предыдущий SUMMARY (Superseded) или MIMO checklist

**Не класть** оркестраторские multi-agent отчёты в `docs/audit/` — только в `voly-runs/` / mission folders.

## Запуск

```bash
cd voly
python -m projects.twogc twogc combat run cloudbridge-relay-fixes --sequential
```

Отчёты: `docs/agent-reports/<mission-name>/`.
