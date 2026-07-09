# ADR-004: Modern data plane (QUIC/MASQUE first) with WireGuard fallback

**Status:** Accepted  
**Date:** 2026-07-09  

## Context

Segment buyers compare CloudBridge to VPN/mesh (WireGuard-based) products. Pure WG is familiar to security teams but commoditized. Differentiation and better behavior under lossy networks / middleboxes require modern transports while retaining a familiar fallback.

## Decision

1. **Primary data plane:** QUIC transport, P2P QUIC, MASQUE (RFC 9297 path), HTTP API + metrics.
2. **NAT traversal:** STUN, TURN (embedded/pool), ICE, optional DERP.
3. **Fallback / mesh familiarity:** WireGuard (site-to-site / tenant tunnels).
4. **Protocol engine** may auto-switch (QUIC / HTTP2 / WebSocket) where implemented.
5. Canonical ports: [openwiki/port-scheme.md](../../../openwiki/port-scheme.md) (e.g. QUIC **9090/UDP**, metrics **5551/TCP**, P2P API **5552**, WG **51820**).

## Consequences

**Positive**

- Technical moat vs legacy VPN and many WG-only meshes.
- Better narrative for latency and connectivity success rate in RFP demos.

**Negative**

- Harder ops debugging than “just WireGuard”.
- Security auditors may request WG-only mode for interim acceptance.

## Business impact

Sell **connectivity quality + modern stack** as differentiator; keep WG visible for conservative buyers. Do not position quantum/post-quantum as GA (service disabled).

## Links

- `cmd/relay/`, `internal/quic/`, `internal/p2p/`, `internal/nat/`
- [MASQUE_USAGE_GUIDE.md](../MASQUE_USAGE_GUIDE.md)
