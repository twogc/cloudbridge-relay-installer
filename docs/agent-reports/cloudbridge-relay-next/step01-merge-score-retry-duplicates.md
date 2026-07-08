# Отчёт агента — cursor: merge score+retry duplicates

**Дата:** 2026-07-09 01:40  
**Миссия:** cloudbridge-relay-next  
**Шаг:** 1

## Выполнено

1. Прочитаны `docs/audit/DUPLICATES_INVENTORY.md` и `docs/agent-reports/cloudbridge-relay-medium/SUMMARY.md`.
2. Grep по `cmd/` и `internal/` подтвердил **0 внешних вызовов** для:
   - `NewSCORETenantNetworkAllocatorV2` — только self в `score_tenant_network_allocator.go`
   - `NewSCORETenantNetworkAllocatorImproved` — только self в `score_tenant_network_allocator_improved.go`
   - `network.RetryConfig` / `DefaultRetryConfig()` — только v2 allocator (отдельный `connection.RetryConfig` не затронут)
3. **Canonical keep:** `score_allocator_production.go` + `retry_config_improved.go`.
4. **Удалены мёртвые дубликаты:**
   - `internal/network/score_tenant_network_allocator.go`
   - `internal/network/score_tenant_network_allocator_improved.go`
   - `internal/network/retry_config.go`
5. Production allocator (`NewSCORETenantNetworkAllocatorProduction`) **оставлен** — не wired в рантайм, но это единственная каноническая реализация для будущего wiring (step 2+).
6. Тесты/импорты в `internal/network/` не требовали правок — `*_test.go` в пакете отсутствуют.
7. Обновлён `docs/audit/DUPLICATES_INVENTORY.md` — отмечены удалённые файлы и секция «Выполнено».
8. Cadence dual-path **не трогался** (по инструкции step 1).

## Изменённые файлы

| Путь | Действие |
|------|----------|
| `internal/network/score_tenant_network_allocator.go` | удалён |
| `internal/network/score_tenant_network_allocator_improved.go` | удалён |
| `internal/network/retry_config.go` | удалён |
| `docs/audit/DUPLICATES_INVENTORY.md` | обновлён (статусы deleted, секция SCORE allocators) |
| `docs/agent-reports/cloudbridge-relay-next/step01-merge-score-retry-duplicates.md` | создан (этот отчёт) |

**Сохранены без изменений (canonical):**

- `internal/network/score_allocator_production.go`
- `internal/network/retry_config_improved.go`

## Проверки (команды и результат)

```bash
# Подтверждение 0 внешних ссылок на удалённые символы
rg 'NewSCORETenantNetworkAllocatorV2|NewSCORETenantNetworkAllocatorImproved' \
  --glob '*.go' cmd/ internal/
# (пусто — OK)

rg 'network\.RetryConfig|DefaultRetryConfig\(\)' --glob '*.go' internal/network/
# (пусто после удаления — OK)

# connection.RetryConfig — отдельный тип, не затронут
rg 'DefaultRetryConfig\(\)' --glob '*.go' internal/connection/
# internal/connection/pool.go, backoff_test.go, retry.go — OK

export PATH="$HOME/.local/go-install/go/bin:$PATH"
go build ./cmd/relay/
# exit 0

go test -short ./internal/network/...
# ? github.com/twogc/cloudbridge-relay-installer/internal/network [no test files]
# exit 0
```

## Блокеры / не сделано

| Элемент | Причина |
|---------|---------|
| Wiring `NewSCORETenantNetworkAllocatorProduction` в `cmd/relay` / `relay_manager_hybrid` | Вне scope step 1; отложено на step 2+ |
| Cadence dual-path merge (`client.go` vs `client_production.go`) | Явно исключено из step 1 |
| Обновление `docs/architecture/ARCHITECTURE_ANALYSIS.md` (ссылка на `score_tenant_network_allocator_improved.go`) | Не запрошено в step 1; drift в документации |
| Commit / push | Не запрошено |
