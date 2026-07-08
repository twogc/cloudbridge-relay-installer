# Documentation Audit

This page documents the restructuring of `docs/` and the canonical relationship between `openwiki/` and `docs/`.

---

## Why Two Documentation Systems?

| System | Audience | Scope | Maintenance |
|---|---|---|---|
| **`openwiki/`** | AI agents, developers | Source-level architecture, domain models, operations, ports | Updated with code changes (automated) |
| **`docs/`** | Humans, stakeholders | Deployment guides, business materials, technical specs, roadmaps | Updated manually, curated for readability |

**Rule of thumb:** When a concept is described in both systems, `openwiki/` is the authoritative source for code-level details; `docs/` is the authoritative source for how-to guides, business context, and technical specifications.

---

## `docs/` Restructuring (July 2025)

### What Changed

100+ files were moved from active `docs/` directories into `docs/legacy/` to remove status reports, completion reports, sprint plans, and other ephemeral documents from the navigable documentation.

### Target Structure

```
docs/
├── README.md              # Index — only live links
├── architecture/          # 5 canonical specs (analysis, complete arch, TCP tunnel, MASQUE, README)
├── business/              # Investor, market, executive summaries
├── guides/                # ~15 key deployment/setup/developer guides
├── analysis/              # Technical analyses (QUIC, migration, ecosystem, client compat)
├── roadmap/               # Strategic plans only (MVP, 4-month, strategic roadmap)
├── security/              # Policies and guides (not audit reports)
├── network/               # Ports, tokens, subdomains reference
├── compliance/            # Russian regulatory compliance
├── troubleshooting/       # Runbooks (Cadence troubleshooting)
├── deployment/            # Setup guides (LB, Kyverno, pilot, relay server)
├── integration/           # Prometheus schema, Grafana dashboards
├── monitoring/            # Monitoring index
├── operations/            # Service catalog (reference only)
└── legacy/                # ~100 archived files + _index.md migration table
```

### Files Moved to `docs/legacy/`

| Source | Count | Type | Reason |
|---|---|---|---|
| `docs/reports/` | ~40 | Status reports, completion reports | Ephemeral, one-time reports |
| `docs/operations/` | 13 | Status reports, pilot snapshots, fixes | Kept only `SERVICE_CATALOG.md` |
| `docs/integration/` | 7 | Completion reports | Kept `PROMETHEUS_METRICS_SCHEMA.md` + `grafana-dashboard.json` |
| `docs/monitoring/` | 2 | Testing reports | Kept `README.md` |
| `docs/implementation/` | 5 | Implementation plans | All moved (plans are stale) |
| `docs/api/` | 1 | MASQUE validation report | Kept `README.md` |
| `docs/deployment/` | 4 | Deployment/testing reports | Kept setup guides + `README.md` |
| `docs/guides/` | 19 | Redundant/specific guides + reports | Kept ~15 key guides |
| `docs/zitadel/` | 1 | Implementation status | Merged unique guides to `guides/ZITADEL_*`, status to legacy |
| `docs/roadmap/` | 8 | Week1/week2 plans, stale status | Kept strategic + MVP plans |
| `docs/architecture/` | 9 | Duplicate architecture overviews | Kept 5 canonical specs |
| `docs/network/` | 3 | Integration/analysis reports | Kept port/token/subdomain references |
| `docs/security/` | 2 | Audit report + zero-trust report | Kept `SECURITY.md`, `ENTERPRISE_SECURITY_COMPLETE.md` |
| Root `docs/*.md` | 6 | STATUS, SYSTEM_OVERVIEW, TESTING, etc. | Moved to legacy |

### Canonical Architecture Files (kept in `docs/architecture/`)

1. `ARCHITECTURE_ANALYSIS.md` — Main architecture analysis (78KB)
2. `CLOUDBRIDGE_COMPLETE_ARCHITECTURE.md` — Complete system architecture (60KB)
3. `CLOUDBRIDGE_TCP_TUNNEL_TECHNICAL_SPEC.md` — TCP tunnel technical spec (75KB)
4. `MASQUE_USAGE_GUIDE.md` — MASQUE protocol usage guide
5. `README.md` — Architecture index

### Canonical Guide Files (kept in `docs/guides/`)

| Guide | Description |
|---|---|
| `VK_CLOUD_DEPLOYMENT_GUIDE.md` | Full VK Cloud Kubernetes deployment |
| `QUICK_DEPLOYMENT_CHECKLIST.md` | Quick deployment checklist |
| `DOCKER_MIRRORS_SETUP.md` | Docker Hub mirrors (Russia) |
| `DEVELOPER_BRIEFING.md` | Developer onboarding |
| `CADENCE_SETUP_GUIDE.md` | Cadence workflow engine setup |
| `MINIKUBE_BUILD_GUIDE.md` | Local Minikube builds |
| `ZITADEL_DEPLOYMENT_GUIDE.md` | Zitadel IAM deployment |
| `ZITADEL_SETUP.md` | Zitadel integration setup |
| `ZITADEL_GET_ADMIN_TOKEN.md` | Admin token guide |
| `ZITADEL_SERVICE_USER_PERMISSIONS.md` | Service user permissions |
| `CLIENT_CONNECTION_GUIDE.md` | Client connection setup |
| `CLOUDFLARE-ORIGIN-SETUP.md` | Cloudflare origin server setup |
| `TURN_POOL_USAGE_GUIDE.md` | TURN pool usage reference |
| `TENANT_CONTROLLER_README.md` | K8s tenant policy controller |

### Canonical Roadmap Files (kept in `docs/roadmap/`)

| Plan | Description |
|---|---|
| `CLOUDBRIDGE_STRATEGIC_ROADMAP.md` | Long-term strategic roadmap |
| `CLOUDBRIDGE_STRATEGIC_ROADMAP_v2.md` | Scalability and certification roadmap |
| `COMPLETE_4MONTH_ROADMAP.md` | 4-month development plan |
| `MONTH1_STABILIZATION_ROADMAP.md` | Month 1 stabilization plan |
| `MVP_IMPLEMENTATION_PLAN.md` | MVP implementation plan |
| `MVP_QUICK_START.md` | MVP quick start |
| `README_ROADMAP.md` | Roadmap index |

---

## What Lives Where

### OpenWiki is authoritative for:
- Source code architecture and design patterns
- Internal package structure and key files
- Transport protocol details (QUIC, MASQUE, protocol engine)
- Auth flow and token structure
- Inter-service communication patterns
- Core ports and service topology
- Configuration system internals
- Domain models (networking, services)

### Docs/ is authoritative for:
- How-to deployment guides (VK Cloud, Minikube, Zitadel)
- Technical specification documents (architecture analysis, TCP tunnel spec)
- Business materials (investor pitch, market analysis)
- Regulatory compliance (152-ФЗ, 187-ФЗ)
- Service catalog reference
- Network reference (ports, tokens, subdomains)
- Roadmap and strategic plans
- Troubleshooting runbooks

### Legacy is for:
- Historical status reports and completion reports
- Sprint/weekly implementation plans
- One-time testing and deployment reports
- Duplicate architecture overviews
- Outdated integration reports
