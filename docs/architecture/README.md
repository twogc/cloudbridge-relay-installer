# Архитектура CloudBridge Relay

## Обзор

Документы по архитектуре CloudBridge: QUIC/MASQUE data plane, multi-tenant platform, self-hosted ZTNA-oriented delivery (K8s + OVA).

## С чего начать (продукт + архитектура)

| Document | Description |
|----------|-------------|
| [**ARCHITECTURE_VS_MARKET.md**](ARCHITECTURE_VS_MARKET.md) | Архитектура vs бизнес-сегмент (SKU, scorecard) |
| [**RFP_REQUIREMENT_MATRIX.md**](RFP_REQUIREMENT_MATRIX.md) | RFP/пилот: требование → компонент → gap |
| [**adr/**](adr/README.md) | ADR-пакет (принятые архитектурные решения) |

## Основные технические документы

| Document | Description |
|----------|-------------|
| [ARCHITECTURE_ANALYSIS.md](ARCHITECTURE_ANALYSIS.md) | Detailed architecture analysis |
| [CLOUDBRIDGE_COMPLETE_ARCHITECTURE.md](CLOUDBRIDGE_COMPLETE_ARCHITECTURE.md) | Complete system architecture |
| [CLOUDBRIDGE_TCP_TUNNEL_TECHNICAL_SPEC.md](CLOUDBRIDGE_TCP_TUNNEL_TECHNICAL_SPEC.md) | TCP tunnel technical spec |
| [MASQUE_USAGE_GUIDE.md](MASQUE_USAGE_GUIDE.md) | MASQUE protocol usage |

## Agent / ops SoT (openwiki)

| Document | Description |
|----------|-------------|
| [openwiki/architecture.md](../../openwiki/architecture.md) | Service topology, IAM |
| [openwiki/port-scheme.md](../../openwiki/port-scheme.md) | Canonical ports |
| [openwiki/deploy.md](../../openwiki/deploy.md) | Canonical deploy path |

## Ключевые концепции

- **QUIC-first** + MASQUE; WireGuard as fallback/mesh familiarity ([ADR-004](adr/ADR-004-modern-data-plane.md))
- **Zero-Trust / OIDC**: Zitadel only ([ADR-003](adr/ADR-003-zitadel-only-iam.md))
- **Thick relay + optional sidecars** ([ADR-001](adr/ADR-001-thick-relay-optional-sidecars.md))
- **K8s canonical + OVA edge** ([ADR-002](adr/ADR-002-k8s-canonical-ova-edge.md))
- **Multi-tenant platform** ([ADR-005](adr/ADR-005-multi-tenant-platform.md))

## Компоненты (кратко)

1. **CloudBridge Relay** — core relay (`cmd/relay`)
2. **P2P / NAT** — peers, ICE/STUN/TURN
3. **QUIC / MASQUE** — modern transport
4. **Zitadel** — IAM
5. **IPAM / SCORE / policies** — multi-tenant platform
6. **AI / Cadence** — optional add-ons
