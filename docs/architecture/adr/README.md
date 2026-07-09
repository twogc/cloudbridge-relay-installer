# Architecture Decision Records (ADR)

Пакет решений, связывающих **архитектуру CloudBridge** с **потребностями B2B-сегмента** (self-hosted connectivity / ZTNA / multi-tenant edge, РФ).

| ID | Title | Status |
|----|-------|--------|
| [ADR-001](ADR-001-thick-relay-optional-sidecars.md) | Thick relay core + optional microservices | Accepted |
| [ADR-002](ADR-002-k8s-canonical-ova-edge.md) | K8s canonical + OVA/edge secondary | Accepted |
| [ADR-003](ADR-003-zitadel-only-iam.md) | Canonical IAM = Zitadel only | Accepted |
| [ADR-004](ADR-004-modern-data-plane.md) | QUIC/MASQUE first, WireGuard fallback | Accepted |
| [ADR-005](ADR-005-multi-tenant-platform.md) | Multi-tenancy / IPAM / SCORE as platform | Accepted |

**Related:**

- [Architecture vs Market](../ARCHITECTURE_VS_MARKET.md) — 1–2 page summary
- [RFP requirement matrix](../RFP_REQUIREMENT_MATRIX.md) — RFP → component → gap
- [openwiki/architecture.md](../../../openwiki/architecture.md)
- [openwiki/port-scheme.md](../../../openwiki/port-scheme.md)
- [openwiki/deploy.md](../../../openwiki/deploy.md)

## Format

Each ADR: Context → Decision → Consequences → Business impact → Links.
