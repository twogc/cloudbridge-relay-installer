# Cursor setup for cloudbridge-relay-installer

## Skills

Project skills live in `.cursor/skills/` (symlinks to `rep/agentic-awesome-skills`).

Install or refresh:

```bash
./scripts/install-cursor-skills.sh
```

Requires local clone: `rep/agentic-awesome-skills` (gitignored).

### Bundles installed

| Plugin | Focus |
|--------|--------|
| AAS DevOps & Cloud | K8s, Terraform, Docker, CI, incidents |
| AAS QA & Test Automation | Go tests, e2e, debugging, TDD |
| AAS Security Engineer | SAST, API security, audits |
| AAS OSS Maintainer | PRs, commits, `AGENTS.md`, CI |
| AAS Observability & IR | Grafana, tracing, SLOs, postmortems |
| AAS API Platform Builder | API design, auth patterns, OpenAPI |
| Extra | `golang-pro`, `go-concurrency-patterns` |

Use in chat: `@kubernetes-architect`, `@golang-pro`, `@systematic-debugging`, etc.

## Agent context

- Root **[AGENTS.md](../AGENTS.md)** — primary instructions for this repo.
- Audit priorities: **[docs/audit/MIMO_AUDIT_REPORT.md](../docs/audit/MIMO_AUDIT_REPORT.md)**.
