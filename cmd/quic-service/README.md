# QUIC Service - Optional Metrics Sidecar

**Status**: Optional component (not deployed in MVP)
**Purpose**: Centralized QUIC experimental metrics collection
**Deployment**: Not required for MVP functionality

## Overview

This service provides a standalone HTTP endpoint for collecting and exposing QUIC experimental metrics. It is **optional** and **not currently deployed** in the Kubernetes cluster.

The main relay server already exposes QUIC metrics at its `/metrics` endpoint, so this service is only needed if you want centralized metrics collection in a separate pod.

## Current Implementation

### Working Endpoints
- `GET /health` - Health check
- `GET /metrics` - Prometheus metrics (BBRv2, ACK Frequency, FEC, Greasing)
- `GET /metrics/summary` - Human-readable metrics summary

### Stub Endpoints (Not Implemented)
- `GET /connections` - Returns empty array (stub)
- `POST /connections` - Returns mock response (stub)

## Architecture

```
cmd/quic-service (Metrics Sidecar - Optional)
    │
    ├─── Uses: internal/quic.ExperimentalMetrics
    │
    └─── Exposes: Prometheus metrics at /metrics

cmd/relay (Main Service - Required)
    │
    ├─── Uses: cmd/relay/enhanced_quic.go (Integration Adapter)
    │
    └─── Uses: internal/quic/enhanced.go (Core Library)
```

### This is NOT duplication:
- **cmd/quic-service**: Optional metrics sidecar
- **cmd/relay/enhanced_quic.go**: Integration adapter for relay
- **internal/quic/enhanced.go**: Core QUIC library

## Usage

### Running Standalone
```bash
export QUIC_PORT=8081
go run cmd/quic-service/main.go
```

### Accessing Metrics
```bash
# Prometheus metrics
curl http://localhost:8081/metrics

# Human-readable summary
curl http://localhost:8081/metrics/summary
```

### Example Metrics Exposed

```
# BBRv2 Congestion Control
cloudbridge_quic_bbrv2_bandwidth_estimate_bytes_per_second
cloudbridge_quic_bbrv2_rtt_estimate_ms
cloudbridge_quic_bbrv2_cwnd_bytes

# ACK Frequency Optimization
cloudbridge_quic_ack_freq_optimizations_total
cloudbridge_quic_ack_freq_current_interval_ms

# Forward Error Correction
cloudbridge_quic_fec_packets_encoded_total
cloudbridge_quic_fec_packets_recovered_total

# QUIC Bit Greasing
cloudbridge_quic_greasing_spin_bit_flips_total
cloudbridge_quic_greasing_reserved_bits_set_total
```

## Deployment

### Current Status: NOT DEPLOYED

This service is **not included** in the k8s/deployments/ directory and is **not required** for MVP.

### To Deploy (Optional)

If you want to deploy this service for centralized metrics collection:

1. **Build Docker image**:
```bash
docker build -f docker/images/Dockerfile.quic -t cloudbridge/quic-service:v1.0.0 .
```

2. **Create deployment manifest** (k8s/deployments/quic-service-deployment.yaml):
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: quic-service
  namespace: cloudbridge
spec:
  replicas: 1
  selector:
    matchLabels:
      app: quic-service
  template:
    metadata:
      labels:
        app: quic-service
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8081"
        prometheus.io/path: "/metrics"
    spec:
      containers:
      - name: quic-service
        image: cloudbridge/quic-service:v1.0.0
        ports:
        - containerPort: 8081
          name: http
        env:
        - name: QUIC_PORT
          value: "8081"
```

3. **Create service manifest** (k8s/services/quic-service.yaml):
```yaml
apiVersion: v1
kind: Service
metadata:
  name: quic-service
  namespace: cloudbridge
spec:
  selector:
    app: quic-service
  ports:
  - port: 8081
    targetPort: 8081
    name: http
```

## Future Enhancements

This service could be expanded to become a full QUIC service:

1. **Implement Real Connection Management**:
   - Replace stub `/connections` endpoints
   - Handle actual QUIC connections
   - Manage connection lifecycle

2. **Extract QUIC from Relay**:
   - Move QUIC logic from relay to this service
   - Define gRPC/HTTP API between relay and quic-service
   - Update relay to be a client of quic-service

3. **Independent Scaling**:
   - Scale QUIC service separately from relay
   - Dedicated resources for QUIC connections
   - Better resource utilization

## Related Files

- **Source**: [main.go](main.go) - 77 lines
- **Dockerfile**: [docker/images/Dockerfile.quic](../../docker/images/Dockerfile.quic)
- **Core Library**: [internal/quic/enhanced.go](../../internal/quic/enhanced.go)
- **Analysis**: [QUIC_SERVICE_ANALYSIS.md](../../QUIC_SERVICE_ANALYSIS.md)

## Decision

**For MVP**: Do not deploy, relay's `/metrics` endpoint is sufficient

**Post-MVP**: Consider deploying if centralized metrics collection is needed

---

Last Updated: 2025-10-13
Status: Optional component, not deployed in MVP
