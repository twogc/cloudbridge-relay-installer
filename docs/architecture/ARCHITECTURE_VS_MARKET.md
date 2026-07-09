# Architecture vs Market (CloudBridge)

**Date:** 2026-07-09  
**Audience:** product, architecture, pre-sales  
**Companion:** [ADR package](adr/README.md) · [RFP matrix](RFP_REQUIREMENT_MATRIX.md)

---

## Segment (one paragraph)

CloudBridge targets **B2B self-hosted connectivity** in regulated / RF-aligned markets: ZTNA and remote access, site-to-site, multi-tenant edge (MSP/OVA), where **data residency**, **NAT traversal**, and **customer-controlled identity** matter more than pure SaaS mesh UX (Tailscale-class).

---

## Architectural thesis

Build a **platform** (modern data plane + multi-tenant control + dual delivery K8s/OVA), not a single-purpose WireGuard UI. Package into SKUs so Edge pilots stay small and Platform captures MSP economics.

```
Business need          Architecture response              ADR
─────────────────      ─────────────────────────────      ──────
Pilot in 1 day         Thick relay / OVA minimal          001, 002
RF data control        Self-host + Zitadel in contour     002, 003
Works behind NAT       QUIC/P2P + STUN/TURN/ICE + WG      004
MSP / multi-customer   IPAM, policies, SCORE optional     005
Scale later            K8s HA, sidecars, overlays         001, 002
```

---

## Scorecard (segment fit)

| Criterion | Fit | Note |
|-----------|-----|------|
| Self-host / residency | High | Core strength |
| NAT / connectivity | High | Full protocol stack |
| Multi-tenant / MSP | Med–High | Code ready; ops packaging |
| Compliance path (152-FZ / ФСТЭК) | Medium | Hosted contour OK; cert process open |
| Time-to-value | Medium | Needs Edge SKU discipline |
| Ops TCO | Med–Low | Without packaging, support cost rises |
| Tech differentiation | High | QUIC/MASQUE; AI is add-on only |

---

## SKU map (architecture → offer)

| SKU | Include | Exclude by default |
|-----|---------|-------------------|
| **Edge / OVA** | relay, Zitadel, Postgres, metrics | Cadence, AI, SCORE |
| **Cloud Standard** | K8s 00–08 + Zitadel, monitoring | Heavy AI |
| **Platform / MSP** | + IPAM, tenant controller, SCORE hybrid | — |
| **AI add-on** | ai-service / worker + Cadence | — |

---

## Do / Don't for GTM

**Do**

- Sell self-host, OIDC, connectivity reliability, multi-tenant platform path.
- Keep port/ops SoT: [port-scheme.md](../../openwiki/port-scheme.md).

**Don't**

- Claim ФСТЭК-certified or quantum-ready GA (quantum disabled).
- Ship full monorepo as default pilot.
- Position AI as mandatory for first contract.

---

## Related ADRs

- [ADR-001](adr/ADR-001-thick-relay-optional-sidecars.md) Thick core + sidecars  
- [ADR-002](adr/ADR-002-k8s-canonical-ova-edge.md) K8s + OVA  
- [ADR-003](adr/ADR-003-zitadel-only-iam.md) Zitadel only  
- [ADR-004](adr/ADR-004-modern-data-plane.md) QUIC/MASQUE + WG  
- [ADR-005](adr/ADR-005-multi-tenant-platform.md) Multi-tenant platform  

## Market inputs (business docs)

- [docs/business/EXECUTIVE_SUMMARY.md](../business/EXECUTIVE_SUMMARY.md)
- [docs/business/MARKET_ANALYSIS_RU.md](../business/MARKET_ANALYSIS_RU.md)
- [docs/business/INVESTOR_SUMMARY.md](../business/INVESTOR_SUMMARY.md)
