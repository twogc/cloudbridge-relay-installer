# VOLY Cursor — Step 2/3 (Quick wins)

**Дата:** 2026-07-09  
**Репозиторий:** `cloudbridge-relay-installer`  
**Источник:** `docs/audit/MIMO_AUDIT_REPORT.md` (раздел Quick wins)

---

## Изменённые файлы

| Файл | Действие | Зачем |
|---|---|---|
| `Makefile` | `./internal/keycloak/...` → `./internal/zitadel/...` в target `test-critical` | Цель `make test-critical` ссылалась на несуществующий пакет `internal/keycloak/`; активный auth-провайдер — Zitadel (`internal/zitadel/`) |

---

## Проверено, но не изменено

| Объект | Решение | Причина |
|---|---|---|
| `systemd/enhanced-quic-proxy.service` | **Оставлен** | Аудит утверждал идентичность с `quic-proxy.service`, но `diff` показывает различия: порт `9092` vs `9090`, другой backend (`30467` vs `31284`), другой Description. Удаление по условию «если идентичен» не применимо |

---

## Не затронуто (по правилам задачи)

- `voly/` — не изменялся
- Commit / push — не выполнялись
- Прочие quick wins из аудита (удаление Keycloak-кода, quantum, `src/`, `.disabled` workflows и т.д.) — вне scope Step 2

---

## Проверка

```bash
# Убедиться, что target больше не ссылается на keycloak
grep -n 'test-critical\|keycloak\|zitadel' Makefile

# При наличии Go:
make test-critical
```
