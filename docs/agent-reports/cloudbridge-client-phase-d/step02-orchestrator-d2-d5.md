# Phase D.2 + D.5 — orchestrator

**Date:** 2026-07-09

## D.2 ICE
- `ice.signaling_enabled` default **false**
- `p2p.Manager.SetICESignalingEnabled` + skip ExchangeICECredentials when off
- CONTRACT: ICE REST N/A until relay routes exist

## D.5 Auth profiles
- docs/AUTH_PROFILES.md — lab-hmac vs prod-oidc
- CONTRACT §5.1
- CHECKLISTS JWT lab-hmac marked

## Disk
- Still 100G; expand not available on hypervisor

## Verify
go test ./pkg/pathselect/ ./pkg/types/ ./pkg/config/ ./pkg/p2p/ ./pkg/auth/ -short
go build ./cmd/cloudbridge-client
