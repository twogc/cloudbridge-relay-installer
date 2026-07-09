# RFP / pilot requirement matrix

**Date:** 2026-07-09  
**Purpose:** Map typical enterprise/MSP RFP and pilot requirements → architecture component → maturity gap.  
**Use:** pre-sales checklist, gap analysis for pilots, roadmap prioritization.

**Legend (status)**

| Status | Meaning |
|--------|---------|
| **Ready** | Implemented and documentable in pilot |
| **Partial** | Code/path exists; packaging, HA, or proof incomplete |
| **Roadmap** | Explicitly planned / not GA |
| **Out** | Not in product / do not promise |

---

## 1. Connectivity & data plane

| # | Requirement (RFP language) | Component(s) | Status | Gap / action |
|---|----------------------------|--------------|--------|--------------|
| C1 | Secure remote access / site connectivity | `cmd/relay`, P2P, QUIC, WG | **Ready** | Package Edge vs Platform ports ([port-scheme](../../openwiki/port-scheme.md)) |
| C2 | Works through NAT/firewall | STUN/TURN/ICE, TURN pool, DERP | **Partial** | Document success-rate targets; size TURN for pilot |
| C3 | Modern / high-performance transport | QUIC, MASQUE | **Ready** | Demo script; WG fallback for conservative reviewers |
| C4 | WireGuard mesh / site-to-site | WireGuard managers, K8s ports 51820 | **Partial** | Ops runbook; cert-manager/TLS paths vary |
| C5 | Multi-protocol fallback | Protocol engine, WS/HTTP2 paths | **Partial** | Clarify which modes are GA in SKU sheet |
| C6 | IPv6 / dual-stack | — | **Roadmap** | Do not promise in pilot unless scoped |
| C7 | Multi-region / Anycast | — | **Roadmap** | Single-region K8s is baseline |

## 2. Identity, access, security

| # | Requirement | Component(s) | Status | Gap / action |
|---|-------------|--------------|--------|--------------|
| S1 | Enterprise SSO (OIDC/SAML) | Zitadel OIDC/JWT JWKS | **Partial** | OIDC ready; SAML via Zitadel config — document; no Keycloak |
| S2 | JWT / API auth for clients | `internal/auth`, `internal/zitadel`, P2P tokens | **Ready** | Token issuance runbook for pilot |
| S3 | RBAC / least privilege | Zitadel roles, K8s RBAC | **Partial** | Matrix of roles per SKU |
| S4 | mTLS service-to-service | — | **Roadmap** | Pitch TLS to ingress + Zitadel today |
| S5 | Audit logs (auth, admin) | Zap logs, optional security audit paths | **Partial** | Central log shipping not default |
| S6 | Rate limit / DDoS basics | rate limiter, circuit breaker | **Partial** | Edge WAF/ingress policies customer-owned |
| S7 | ФСТЭК / certified crypto product | — | **Out** / **Roadmap** | Self-host path only; no cert claim |
| S8 | 152-FZ data residency | Self-host deploy RF cloud/on-prem | **Ready** | Customer hosts data plane + IdP |
| S9 | Post-quantum crypto | quantum-service | **Out** | Disabled; remove from RFP answers |
| S10 | Secrets management | K8s secrets, ExternalSecrets (ops) | **Partial** | Vault/ESO not mandatory in Edge |

## 3. Multi-tenancy & networking control

| # | Requirement | Component(s) | Status | Gap / action |
|---|-------------|--------------|--------|--------------|
| T1 | Multi-tenant isolation | tenants, network store, policies | **Partial** | Prove isolation test plan in pilot |
| T2 | IP address management | `ipam-service` :8080, PeerIPAM | **Partial** | Postgres preferred for prod peer IPAM |
| T3 | Integration with external inventory (SCORE) | `score_wiring`, SCORE env | **Partial** | Optional; env-gated; needs customer SCORE |
| T4 | K8s network policies / CNI | CNI plugin, netpol manifests | **Partial** | Platform SKU; not Edge |
| T5 | Tenant policy as CRDs | tenant-policy-controller | **Partial** | Ops maturity / examples |
| T6 | Per-tenant quotas / billing hooks | — | **Roadmap** | Design metrics labels for future billing |

## 4. Deployment & operations

| # | Requirement | Component(s) | Status | Gap / action |
|---|-------------|--------------|--------|--------------|
| O1 | Deploy on Kubernetes | `k8s-manifests`, `deploy.sh`, overlays | **Ready** | Default file-by-file; overlay optional |
| O2 | Deploy as appliance / OVA | `mws/`, cloud-init, compose | **Partial** | Align metrics ports (5551) in customer docs |
| O3 | HA (multi-replica) | relay replicas=3 in manifests | **Partial** | Session stickiness / state assumptions |
| O4 | HPA / autoscaling | — | **Roadmap** | Manual scale today |
| O5 | Monitoring (Prometheus/Grafana) | metrics :5551, ServiceMonitor, dashboards | **Partial** | ServiceMonitor needs Prometheus Operator |
| O6 | Alerting / SLO | Alertmanager paths (ops) | **Roadmap** | Define pilot SLOs explicitly |
| O7 | Backup / DR | PVC, Postgres backups (customer) | **Partial** | RTO/RPO customer-owned runbook |
| O8 | Air-gapped install | OVA + image load | **Partial** | Mirror/docs; offline image bundle |
| O9 | GitOps / IaC | terraform-k8s, manifests | **Partial** | Overlay not default; Argo not bundled |
| O10 | Multi-arch (ARM) | — | **Roadmap** | x86_64 baseline |

## 5. Integration & extensibility

| # | Requirement | Component(s) | Status | Gap / action |
|---|-------------|--------------|--------|--------------|
| I1 | REST API for peers/tunnels | P2P API :5552, admin/HTTP API | **Ready** | OpenAPI freshness check |
| I2 | Workflow / automation | Cadence | **Partial** | Platform/AI path; not Edge |
| I3 | AI ops / anomaly / routing | ai-service :8083, workers | **Partial** | Sell as add-on with KPI |
| I4 | SIEM export | logs/metrics scrape | **Partial** | Customer connects sinks |
| I5 | Terraform provider for tenants | — | **Roadmap** | Use raw manifests/API |

## 6. Commercial / non-functional (often in RFP)

| # | Requirement | Component(s) | Status | Gap / action |
|---|-------------|--------------|--------|--------------|
| B1 | Open source / auditability | public monorepo | **Ready** | License clarity in contract |
| B2 | Import substitution / registry RF | self-host, local images | **Partial** | Registry/mirror process |
| B3 | SLA 99.9% | K8s HA design | **Partial** | Contract SLA vs measured SLO |
| B4 | Support L1–L3 | process | **Roadmap** | Not architecture alone |
| B5 | Price vs Tailscale/UserGate | self-host TCO | **Ready** (commercial) | Ops complexity can erase savings if over-provisioned |

---

## Gap priority for pilots (actionable)

| Priority | Gaps to close before / during pilot | Related ADR |
|----------|-------------------------------------|-------------|
| **P0** | Edge SKU install path; port SoT; Zitadel JWT demo; NAT success demo | 001–004 |
| **P1** | Tenant isolation test; Postgres IPAM; monitoring scrape 5551/8083 | 005, 002 |
| **P2** | Cadence/AI only if in contract; SCORE only if customer has SCORE | 001, 005 |
| **Avoid** | ФСТЭК certificate, quantum GA, multi-region Anycast as baseline | — |

---

## How to use this matrix in a bid

1. Copy rows into the response sheet; mark **Ready / Partial / Out**.  
2. For each **Partial**, attach one mitigation (runbook, config, out-of-scope).  
3. Never map **Out** to “yes”.  
4. Align SKU with [ARCHITECTURE_VS_MARKET.md](ARCHITECTURE_VS_MARKET.md).  
5. Technical SoT: [openwiki/deploy.md](../../openwiki/deploy.md), [port-scheme.md](../../openwiki/port-scheme.md).

---

## Traceability

| Doc | Role |
|-----|------|
| [adr/README.md](adr/README.md) | Decision records |
| [ARCHITECTURE_VS_MARKET.md](ARCHITECTURE_VS_MARKET.md) | Business–architecture summary |
| [docs/business/MARKET_ANALYSIS_RU.md](../business/MARKET_ANALYSIS_RU.md) | Market context |
| [docs/roadmap/CLOUDBRIDGE_STRATEGIC_ROADMAP_v2.md](../roadmap/CLOUDBRIDGE_STRATEGIC_ROADMAP_v2.md) | Certification / scale roadmap |
