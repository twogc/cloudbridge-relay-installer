# Отчёты агентов (VOLY / Cursor / MiMo)

Отдельная папка для **отчётов о выполненной работе** агентов. Не путать с `docs/audit/` (аудит-анализ) и логами прогонов.

## Структура

```
docs/agent-reports/
├── README.md                    # этот файл
├── voly-runs/                   # сводные отчёты оркестратора VOLY (multi-agent)
│   └── YYYYMMDD_HHMMSS_2gc-<mission>.md
├── cloudbridge-relay-audit/     # пошаговые отчёты миссии audit
│   ├── step01-....md
│   └── SUMMARY.md
└── cloudbridge-relay-fixes/     # пошаговые отчёты миссии fixes
    ├── step01-keycloak-auth.md
    └── SUMMARY.md
```

**Не класть** оркестраторские отчёты в `docs/audit/` — там только аудит-анализ (`MIMO_AUDIT_REPORT.md` и т.п.).

## Правила

Каждый шаг combat-миссии **обязан** создать markdown-файл в подпапке миссии:

- дата и номер шага
- список изменённых файлов
- что сделано и зачем
- результаты проверок (`go build`, `go test`, grep)
- блокеры и невыполненные пункты

## Запуск

```bash
cd voly
python -m projects.twogc twogc combat run cloudbridge-relay-fixes --sequential
```

Отчёты появятся в `docs/agent-reports/<mission-name>/`.
