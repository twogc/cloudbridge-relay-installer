# ADR-002: Kubernetes canonical deploy + OVA/edge secondary

**Status:** Accepted  
**Date:** 2026-07-09  

## Context

Target segments need both:

- **Scale / HA / MSP** on private cloud (VK Cloud K8s)
- **Air-gapped / site appliance** (MWS Marketplace OVA, docker-compose minimal)

Single delivery model cannot cover both unit economics without two support surfaces.

## Decision

| Path | Role | Canonical artifacts |
|------|------|---------------------|
| **K8s production** | Primary enterprise/MSP | `terraform-k8s/` → `scripts/deploy/deploy.sh` → `k8s-manifests/` (optional `k8s/overlays/production`) |
| **OVA / compose** | Edge / closed contour | `mws/`, docker-compose templates |
| **Ops extras** | Component-level | `k8s/` (Zitadel Helm values, Cadence schema jobs, BGP, etc.) |

Default deploy remains **file-by-file** `k8s-manifests/`. Overlay mode is optional (`deploy.sh --use-overlay`). Zitadel IAM after core: `k8s/overlays/zitadel` or numbered `09–15`.

## Consequences

**Positive**

- Two GTM channels without two products.
- K8s path matches enterprise RFP language (HA, replicas, monitoring).

**Negative**

- Dual ops documentation and CI surface.
- Risk of drift between `k8s/` and `k8s-manifests/` (mitigated by consolidation plan + port-scheme SoT).

## Business impact

Enables simultaneous “cloud platform” and “box on customer site” offers. Support cost must be budgeted for both SKUs.

## Links

- [openwiki/deploy.md](../../../openwiki/deploy.md)
- [docs/guides/OVERLAY_E2E_CHECKLIST.md](../../guides/OVERLAY_E2E_CHECKLIST.md)
- [docs/audit/CONSOLIDATION_PLAN.md](../../audit/CONSOLIDATION_PLAN.md)
