# Phase D.1 — orchestrator (mimo-only policy)

**Date:** 2026-07-09

## Done
- `api.route_monitoring_enabled` default **false**
- `api.Manager` skips Open/CloseConnection monitoring when disabled
- Wired through p2p/relay/CLI ManagerConfig
- CONTRACT + STATUS + GAP updated

## Verify
```
go test ./pkg/pathselect/ ./pkg/types/ ./pkg/config/ ./pkg/api/ -count=1
go build ./cmd/cloudbridge-client
```

## Next
Mission `cloudbridge-client-phase-d` (mimo) for D.2–D.5 plan → orchestrator implement.
