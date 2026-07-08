# Documentation

Human-oriented documentation index for CloudBridge Relay. For agent-oriented overview, see [openwiki/quickstart.md](../openwiki/quickstart.md).

**Canonical sources of truth (do not contradict these):**

| Topic | Document |
|-------|----------|
| Ports | [openwiki/port-scheme.md](../openwiki/port-scheme.md) |
| Deploy (VK Cloud K8s) | [openwiki/deploy.md](../openwiki/deploy.md) |
| IAM | **Zitadel** only — [guides/ZITADEL_SETUP.md](guides/ZITADEL_SETUP.md) |
| Audit status / backlog | [audit/MIMO_AUDIT_REPORT.md](audit/MIMO_AUDIT_REPORT.md) |
| Agent mission reports | [agent-reports/README.md](agent-reports/README.md) |

`docs/legacy/` is archive — not operational SoT.

---

## Guides (deployment, setup, developer)

| Document | Description |
|----------|-------------|
| [VK_CLOUD_DEPLOYMENT_GUIDE.md](guides/VK_CLOUD_DEPLOYMENT_GUIDE.md) | Full VK Cloud Kubernetes deployment |
| [QUICK_DEPLOYMENT_CHECKLIST.md](guides/QUICK_DEPLOYMENT_CHECKLIST.md) | Quick deployment checklist |
| [OVERLAY_E2E_CHECKLIST.md](guides/OVERLAY_E2E_CHECKLIST.md) | Kustomize overlay offline + live e2e checklist |
| [DOCKER_MIRRORS_SETUP.md](guides/DOCKER_MIRRORS_SETUP.md) | Docker Hub mirrors for Russia |
| [DEVELOPER_BRIEFING.md](guides/DEVELOPER_BRIEFING.md) | Developer onboarding |
| [CADENCE_SETUP_GUIDE.md](guides/CADENCE_SETUP_GUIDE.md) | Cadence workflow engine setup |
| [MINIKUBE_BUILD_GUIDE.md](guides/MINIKUBE_BUILD_GUIDE.md) | Local Minikube builds |
| [ZITADEL_DEPLOYMENT_GUIDE.md](guides/ZITADEL_DEPLOYMENT_GUIDE.md) | Zitadel IAM deployment |
| [CLIENT_CONNECTION_GUIDE.md](guides/CLIENT_CONNECTION_GUIDE.md) | Client connection setup |
| [CLOUDFLARE-ORIGIN-SETUP.md](guides/CLOUDFLARE-ORIGIN-SETUP.md) | Cloudflare origin server setup |
| [TURN_POOL_USAGE_GUIDE.md](guides/TURN_POOL_USAGE_GUIDE.md) | TURN pool usage reference |
| [TENANT_CONTROLLER_README.md](guides/TENANT_CONTROLLER_README.md) | K8s tenant policy controller |
| [ZITADEL_SETUP.md](guides/ZITADEL_SETUP.md) | Zitadel integration setup |
| [ZITADEL_GET_ADMIN_TOKEN.md](guides/ZITADEL_GET_ADMIN_TOKEN.md) | Zitadel admin token guide |
| [ZITADEL_SERVICE_USER_PERMISSIONS.md](guides/ZITADEL_SERVICE_USER_PERMISSIONS.md) | Zitadel service user permissions |

## Architecture

| Document | Description |
|----------|-------------|
| [ARCHITECTURE_ANALYSIS.md](architecture/ARCHITECTURE_ANALYSIS.md) | Main architecture analysis |
| [CLOUDBRIDGE_COMPLETE_ARCHITECTURE.md](architecture/CLOUDBRIDGE_COMPLETE_ARCHITECTURE.md) | Complete system architecture |
| [CLOUDBRIDGE_TCP_TUNNEL_TECHNICAL_SPEC.md](architecture/CLOUDBRIDGE_TCP_TUNNEL_TECHNICAL_SPEC.md) | TCP tunnel technical spec |
| [MASQUE_USAGE_GUIDE.md](architecture/MASQUE_USAGE_GUIDE.md) | MASQUE protocol usage guide |

## Analysis

See [analysis/README.md](analysis/README.md) and service-specific analyses in `analysis/`.

## Roadmap

| Document | Description |
|----------|-------------|
| [README_ROADMAP.md](roadmap/README_ROADMAP.md) | Roadmap index |
| [CLOUDBRIDGE_STRATEGIC_ROADMAP.md](roadmap/CLOUDBRIDGE_STRATEGIC_ROADMAP.md) | Strategic roadmap |
| [MVP_IMPLEMENTATION_PLAN.md](roadmap/MVP_IMPLEMENTATION_PLAN.md) | MVP plan |
| [COMPLETE_4MONTH_ROADMAP.md](roadmap/COMPLETE_4MONTH_ROADMAP.md) | 4-month roadmap |
| [MONTH1_STABILIZATION_ROADMAP.md](roadmap/MONTH1_STABILIZATION_ROADMAP.md) | Month 1 stabilization |

## Operations

| Document | Description |
|----------|-------------|
| [SERVICE_CATALOG.md](operations/SERVICE_CATALOG.md) | Service catalog |
| [ADMIN_API.md](ADMIN_API.md) | Admin API reference |
| [TENANT_ISOLATION_OPERATIONS_GUIDE.md](TENANT_ISOLATION_OPERATIONS_GUIDE.md) | Tenant isolation operations |

## Business

| Document | Description |
|----------|-------------|
| [EXECUTIVE_SUMMARY.md](business/EXECUTIVE_SUMMARY.md) | Executive summary |
| [INVESTOR_PITCH_RU.md](business/INVESTOR_PITCH_RU.md) | Investor pitch (Russian) |
| [INVESTOR_SUMMARY.md](business/INVESTOR_SUMMARY.md) | Investor summary |
| [MARKET_ANALYSIS_RU.md](business/MARKET_ANALYSIS_RU.md) | Market analysis (Russian) |

## Network, Security & Compliance

| Directory | Description |
|-----------|-------------|
| [network/](network/) | Ports (→ port-scheme), tokens, subdomains |
| [security/](security/) | Security policy; IAM = Zitadel (not Keycloak) |
| [compliance/](compliance/) | Compliance documentation |
| [troubleshooting/](troubleshooting/) | Incident runbooks |
| [monitoring/](monitoring/) | Monitoring stack overview |
| [audit/](audit/) | MiMo audit, consolidation plan, inventories |
| [agent-reports/](agent-reports/) | VOLY / agent step reports |

## Deployment

| Document | Description |
|----------|-------------|
| [deployment/README.md](deployment/README.md) | Deployment docs index |
| [deployment/EXTERNAL_LB_SETUP.md](deployment/EXTERNAL_LB_SETUP.md) | External load balancer setup |
| [deployment/KYVERNO_DEPLOYMENT_GUIDE.md](deployment/KYVERNO_DEPLOYMENT_GUIDE.md) | Kyverno deployment |
| [deployment/RELAY_SERVER_SETUP.md](deployment/RELAY_SERVER_SETUP.md) | Relay server setup |

## Archive

Legacy and archived documentation is in [legacy/_index.md](legacy/_index.md) with a full migration table.

---

**Agent-oriented overview:** [openwiki/quickstart.md](../openwiki/quickstart.md)
