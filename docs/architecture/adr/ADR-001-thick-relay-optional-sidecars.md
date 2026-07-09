# ADR-001: Thick relay core + optional microservices

**Status:** Accepted  
**Date:** 2026-07-09  
**Deciders:** Architecture / product (CloudBridge / 2GC)

## Context

Business needs fast pilots (single appliance or one binary) and later upsell to a platform (AI, workflows, IPAM, multi-tenant policies). A pure microservices-only design slows Edge/OVA time-to-value; a pure monolith blocks MSP scale.

## Decision

1. **Core value** lives in `cmd/relay` — P2P, QUIC, MASQUE, WireGuard, STUN/TURN/ICE, auth, metrics.
2. **Optional sidecars** as separate `cmd/*` when packaging requires isolation: `ai-service`, `ai-worker`, `ipam-service`, Cadence workers, tenant-policy-controller, CNI.
3. Product **SKUs** map to enablement, not separate repos:
   - Edge: relay (+ Zitadel/Postgres) without AI/Cadence by default
   - Cloud Standard: K8s core stack
   - Platform: + IPAM/SCORE/tenant controller/AI add-on

## Consequences

**Positive**

- One process for POC/OVA; same codebase for platform.
- Clear commercial packaging without forking.

**Negative**

- `cmd/relay` remains a large composition root (god-object risk).
- Ops surface grows if all sidecars are “on” by default.

**Neutral**

- Feature flags / env (e.g. SCORE, AI) control activation.

## Business impact

Supports dual GTM: quick pilot → expand to MSP/platform without rewrite. Requires sales/deploy discipline so Edge SKU does not ship full monorepo complexity.

## Links

- `cmd/relay/`, `cmd/ai-service/`, `cmd/ipam-service/`
- [ARCHITECTURE_VS_MARKET.md](../ARCHITECTURE_VS_MARKET.md)
- [RFP_REQUIREMENT_MATRIX.md](../RFP_REQUIREMENT_MATRIX.md)
