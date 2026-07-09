# ADR-003: Canonical IAM is Zitadel only

**Status:** Accepted  
**Date:** 2026-07-09  

## Context

Enterprise buyers require OIDC/SSO, auditability, and a single identity story. Historical Keycloak/Django paths increased security surface and confused integrators. Regulatory narrative (152-FZ, data residency) favors customer-controlled IdP deployment.

## Decision

1. **Only Zitadel** is the supported identity provider for production.
2. Validation: OIDC/JWT via JWKS in `internal/auth/`, `internal/zitadel/`.
3. **Do not reintroduce** Keycloak or Django auth providers.
4. Deploy: `k8s-manifests/09+` / `k8s/overlays/zitadel` / guides under `docs/guides/ZITADEL_*`.

## Consequences

**Positive**

- One SSO integration path for RFP responses.
- Smaller auth attack surface and cleaner docs.

**Negative**

- Customers standardized on Keycloak must federate or migrate.
- Zitadel ops becomes a required skill for Platform/Cloud SKUs.

## Business impact

Strengthens RF enterprise and compliance storytelling (“identity in your contour”). Edge SKU still deploys Zitadel (or documents external OIDC later if product allows).

## Links

- [openwiki/architecture.md](../../../openwiki/architecture.md) — IAM section
- `docs/guides/ZITADEL_SETUP.md`, `ZITADEL_DEPLOYMENT_GUIDE.md`
- Mission reports: Keycloak removal (`docs/agent-reports/cloudbridge-relay-fixes/`)
