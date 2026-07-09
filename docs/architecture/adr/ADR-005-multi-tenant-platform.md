# ADR-005: Multi-tenancy as platform capability (IPAM / policies / SCORE)

**Status:** Accepted  
**Date:** 2026-07-09  

## Context

MSP, holdings, and marketplace OVA need **tenant isolation**, address allocation, and optional external control-plane (SCORE) for network inventory. Without multi-tenancy, product collapses to single-org VPN and loses platform unit economics.

## Decision

1. **First-class multi-tenancy** in platform SKU: tenant networks, peer IPAM, policy controller, CNI integration.
2. **IPAM service** (`cmd/ipam-service`, default **:8080**) and related repositories.
3. **Optional SCORE hybrid:** `SCORE_BASE_URL` / `SCORE_API_KEY` wires production allocator + store; peer IPAM prefers Postgres (`SCORE_DATABASE_URL` / `SCORE_PG_*`) then SQLite.
4. **Edge SKU** may run without SCORE; isolation still via config/tenants as packaged.
5. Tenant policy controller and CNI are platform/ops components, not required for minimal Edge.

## Consequences

**Positive**

- Enables MSP billing, multi-customer OVA, and enterprise org trees.
- Aligns with marketplace packaging.

**Negative**

- Operational complexity (Postgres, SCORE availability, hybrid wiring).
- Mis-packaging SCORE/IPAM into Edge increases support load.

## Business impact

Platform and MSP offers depend on this ADR. Edge must stay simpler; Platform must document Postgres + optional SCORE as hard dependencies.

## Links

- `internal/p2p/score_wiring.go`, `internal/database/postgres_peer_ipam.go`
- `cmd/ipam-service/`, `cmd/tenant-policy-controller/`, `cmd/cni-plugin/`
- [openwiki/domain-networking.md](../../../openwiki/domain-networking.md)
