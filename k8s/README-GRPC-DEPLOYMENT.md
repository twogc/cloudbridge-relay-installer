# CloudBridge gRPC Deployment Guide

## Overview

This guide provides step-by-step instructions for deploying gRPC support to the CloudBridge Relay server in Kubernetes.

## Prerequisites

- Kubernetes cluster with NGINX Ingress Controller
- Prometheus and Grafana monitoring stack  
- kubectl configured for cluster access
- CloudBridge namespace created

## Deployment Steps

### 1. Pre-Deployment Validation

Run the validation job to ensure all prerequisites are met:

```bash
kubectl apply -f k8s/canary/staging-rollout.yaml
kubectl wait --for=condition=complete job/grpc-rollout-validator -n cloudbridge --timeout=300s
kubectl logs job/grpc-rollout-validator -n cloudbridge
```

### 2. Generate TLS Certificates

Generate mTLS certificates for gRPC:

```bash
kubectl apply -f k8s/security/grpc-mtls-ca.yaml
kubectl wait --for=condition=complete job/generate-grpc-certs -n cloudbridge --timeout=300s
kubectl get secrets -n cloudbridge | grep -E "(cloudbridge-tls|grpc-)"
```

Verify certificates were created:

```bash
kubectl get secret cloudbridge-tls -n cloudbridge -o yaml
kubectl get secret grpc-ca-secret -n cloudbridge -o yaml  
kubectl get secret grpc-client-certs -n cloudbridge -o yaml
```

### 3. Update Configuration

Apply updated ConfigMap with gRPC settings:

```bash
kubectl apply -f k8s/configmaps/cloudbridge-config.yaml
kubectl rollout restart deployment/cloudbridge-relay -n cloudbridge
```

### 4. Apply Network Policies

Secure gRPC traffic with NetworkPolicies:

```bash
kubectl apply -f k8s/network-policies/cloudbridge-grpc-netpol.yaml
```

### 5. Deploy Monitoring

Add gRPC dashboard and alerts:

```bash
kubectl apply -f k8s/configmaps/grafana-grpc-dashboard.yaml
kubectl apply -f k8s/monitoring/grpc-alerts.yaml
```

### 6. Update Services

Add gRPC port to existing services:

```bash
kubectl apply -f k8s/services/cloudbridge-relay-nodeport.yaml
kubectl apply -f k8s/services/cloudbridge-relay-lb-tcp.yaml
```

### 7. Staged Rollout - Phase 1: Canary (10%)

Deploy canary version with 10% traffic:

```bash
kubectl apply -f k8s/canary/cloudbridge-canary-deployment.yaml
```

Verify canary deployment:

```bash
kubectl get pods -n cloudbridge -l deployment-type=canary
kubectl logs -l deployment-type=canary -n cloudbridge | grep "gRPC Integration started"
```

### 8. Monitor Canary Metrics

Check key metrics during canary phase:

```bash
# Error rate (should be < 0.1%)
kubectl exec -n cloudbridge-monitoring deployment/prometheus -- \
  promtool query instant '(sum(rate(grpc_server_handled_total{job="cloudbridge-relay",grpc_code!="OK"}[5m]))/sum(rate(grpc_server_handled_total{job="cloudbridge-relay"}[5m])))*100'

# P95 latency (should be < 1000ms)  
kubectl exec -n cloudbridge-monitoring deployment/prometheus -- \
  promtool query instant 'histogram_quantile(0.95, sum(rate(grpc_server_handling_seconds_bucket{job="cloudbridge-relay"}[5m])) by (le)) * 1000'

# Request rate
kubectl exec -n cloudbridge-monitoring deployment/prometheus -- \
  promtool query instant 'sum(rate(grpc_server_started_total{job="cloudbridge-relay"}[5m]))'
```

### 9. Phase 2: Increase Traffic (25%)

If canary is successful, increase traffic:

```bash
kubectl patch ingress cloudbridge-relay-canary-ingress -n cloudbridge --type='merge' -p='{"metadata":{"annotations":{"nginx.ingress.kubernetes.io/canary-weight":"25"}}}'

kubectl scale deployment cloudbridge-relay-canary -n cloudbridge --replicas=2
```

### 10. Phase 3: Majority Traffic (50%)

Continue scaling:

```bash
kubectl patch ingress cloudbridge-relay-canary-ingress -n cloudbridge --type='merge' -p='{"metadata":{"annotations":{"nginx.ingress.kubernetes.io/canary-weight":"50"}}}'

kubectl scale deployment cloudbridge-relay-canary -n cloudbridge --replicas=3
```

### 11. Phase 4: Full Rollout (100%)

Complete migration:

```bash
# Update main deployment with gRPC support
kubectl apply -f k8s/deployments/cloudbridge-relay-deployment.yaml

# Scale up main deployment
kubectl scale deployment cloudbridge-relay -n cloudbridge --replicas=6

# Remove canary
kubectl delete ingress cloudbridge-relay-canary-ingress -n cloudbridge
kubectl delete service cloudbridge-relay-canary -n cloudbridge
kubectl delete deployment cloudbridge-relay-canary -n cloudbridge
```

## Verification

### Test gRPC Connectivity

```bash
# Port forward to test locally
kubectl port-forward svc/cloudbridge-relay-nodeport 8444:8444 -n cloudbridge &

# Test with grpc_health_probe
grpc_health_probe -addr=localhost:8444

# Test with grpcurl
grpcurl -plaintext -d '{"version":"1.0.0","client_id":"test","features":["grpc"]}' \
  localhost:8444 relay.v1.ControlService/Hello
```

### Check Service Health

```bash
kubectl get pods -n cloudbridge
kubectl get svc -n cloudbridge | grep grpc
kubectl describe ingress -n cloudbridge
```

### Verify Metrics

```bash
# Check metrics endpoint
kubectl port-forward svc/cloudbridge-relay-nodeport 8084:8084 -n cloudbridge &
curl http://localhost:8084/metrics | grep grpc_server

# Check Prometheus targets
kubectl port-forward -n cloudbridge-monitoring svc/prometheus 9090:9090 &
# Visit http://localhost:9090/targets
```

## Rollback Procedure

If issues are detected:

### Immediate Rollback

```bash
# Scale down canary immediately
kubectl scale deployment cloudbridge-relay-canary -n cloudbridge --replicas=0

# Redirect all traffic to stable version
kubectl patch ingress cloudbridge-relay-canary-ingress -n cloudbridge --type='merge' -p='{"metadata":{"annotations":{"nginx.ingress.kubernetes.io/canary-weight":"0"}}}'
```

### Full Rollback

```bash
# Remove gRPC deployment completely
kubectl delete -f k8s/canary/cloudbridge-canary-deployment.yaml
kubectl delete -f k8s/network-policies/cloudbridge-grpc-netpol.yaml

# Revert ConfigMap
kubectl patch configmap cloudbridge-config -n cloudbridge --patch='{"data":{"config.json":"..."}}'  # Use original config

# Restart services
kubectl rollout restart deployment/cloudbridge-relay -n cloudbridge
```

## Troubleshooting

### Common Issues

1. **Certificate Issues**
   ```bash
   kubectl logs job/generate-grpc-certs -n cloudbridge
   kubectl describe secret cloudbridge-tls -n cloudbridge
   ```

2. **gRPC Service Not Starting**
   ```bash
   kubectl logs -l app.kubernetes.io/name=cloudbridge-relay -n cloudbridge | grep -i grpc
   kubectl describe pods -l app.kubernetes.io/name=cloudbridge-relay -n cloudbridge
   ```

3. **NetworkPolicy Blocking Traffic**
   ```bash
   kubectl describe networkpolicy -n cloudbridge
   kubectl get pods -o wide -n cloudbridge
   ```

4. **Prometheus Not Scraping Metrics**
   ```bash
   kubectl get servicemonitor -n cloudbridge-monitoring
   kubectl logs -l app=prometheus -n cloudbridge-monitoring
   ```

### Debug Commands

```bash
# Check gRPC port accessibility
kubectl exec -it deployment/cloudbridge-relay -n cloudbridge -- netstat -tlnp | grep 8444

# Check certificate validity  
kubectl exec -it deployment/cloudbridge-relay -n cloudbridge -- openssl x509 -in /app/certs/tls.crt -text -noout

# Check configuration loading
kubectl exec -it deployment/cloudbridge-relay -n cloudbridge -- cat /app/configs/config.json | jq .grpc

# Test internal DNS resolution
kubectl exec -it deployment/cloudbridge-relay -n cloudbridge -- nslookup cloudbridge-relay.cloudbridge.svc.cluster.local
```

## Success Criteria

- ✅ Error rate < 0.1% 
- ✅ P95 latency < 1000ms
- ✅ gRPC health checks passing
- ✅ Prometheus metrics being collected
- ✅ Grafana dashboard displaying data
- ✅ No critical alerts firing
- ✅ Client connections successful

## Security Considerations

- All gRPC traffic encrypted with TLS 1.3
- mTLS available for enhanced security
- NetworkPolicies restrict access to gRPC port
- Certificate rotation automated
- Security audit logs enabled

## Performance Tuning

Post-deployment optimizations:

1. **Adjust resource limits based on usage**
2. **Enable HTTP/2 multiplexing**  
3. **Tune keepalive settings**
4. **Enable gRPC compression if needed**
5. **Monitor connection pool usage**

## Support

For issues during deployment:

1. Check logs: `kubectl logs -l app.kubernetes.io/name=cloudbridge-relay -n cloudbridge`
2. Review metrics: http://grafana.local/d/grpc-dashboard
3. Validate configuration: `kubectl get configmap cloudbridge-config -o yaml`
4. Consult runbooks: `/docs/runbooks/grpc-*`
