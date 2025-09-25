# gRPC Deployment Checklist

## Pre-Deployment ☐

### Infrastructure
- [ ] Kubernetes cluster version 1.19+
- [ ] NGINX Ingress Controller installed  
- [ ] Prometheus & Grafana monitoring stack
- [ ] Persistent storage for logs and certificates
- [ ] Network policies support enabled

### Security
- [ ] TLS certificates available or CA for generation
- [ ] Service accounts and RBAC configured
- [ ] Network policies reviewed and approved
- [ ] Security scanning completed

### Configuration  
- [ ] ConfigMap updated with gRPC settings
- [ ] Environment variables for gRPC enabled
- [ ] Certificate mounts configured
- [ ] Resource limits reviewed

## Deployment Steps ☐

### Phase 1: Infrastructure  
- [ ] Apply NetworkPolicies: `kubectl apply -f k8s/network-policies/`
- [ ] Generate certificates: `kubectl apply -f k8s/security/grpc-mtls-ca.yaml`
- [ ] Wait for cert job: `kubectl wait --for=condition=complete job/generate-grpc-certs`
- [ ] Verify secrets created: `kubectl get secrets | grep -E "(grpc|tls)"`

### Phase 2: Configuration
- [ ] Apply ConfigMap: `kubectl apply -f k8s/configmaps/cloudbridge-config.yaml`  
- [ ] Update services: `kubectl apply -f k8s/services/`
- [ ] Apply monitoring: `kubectl apply -f k8s/monitoring/grpc-alerts.yaml`
- [ ] Add Grafana dashboard: `kubectl apply -f k8s/configmaps/grafana-grpc-dashboard.yaml`

### Phase 3: Canary Deployment (10%)
- [ ] Deploy canary: `kubectl apply -f k8s/canary/cloudbridge-canary-deployment.yaml`
- [ ] Verify pods running: `kubectl get pods -l deployment-type=canary`
- [ ] Check logs: `kubectl logs -l deployment-type=canary | grep "gRPC Integration started"`
- [ ] Test connectivity: `grpc_health_probe -addr=<canary-service>:8444`

### Phase 4: Monitoring & Validation
- [ ] Check error rate < 0.1%: Monitor Prometheus query
- [ ] Verify P95 latency < 1000ms: Check Grafana dashboard
- [ ] No critical alerts firing: Review alert-manager
- [ ] gRPC metrics available: `curl <metrics-endpoint> | grep grpc_server`

### Phase 5: Traffic Scaling
- [ ] 25% traffic: Update ingress canary weight to 25
- [ ] Monitor for 1 hour, check success criteria
- [ ] 50% traffic: Update canary weight to 50  
- [ ] Monitor for 2 hours, check success criteria
- [ ] 100% traffic: Update main deployment, scale up

### Phase 6: Cleanup
- [ ] Remove canary deployment: `kubectl delete -f k8s/canary/`
- [ ] Update DNS/ingress to point to main service
- [ ] Remove temporary resources and jobs
- [ ] Update documentation and runbooks

## Validation ☐

### Functional Testing
- [ ] gRPC Hello works: Test ControlService.Hello
- [ ] Authentication works: Test ControlService.Authenticate  
- [ ] Tunnel creation works: Test TunnelService.CreateTunnel
- [ ] Heartbeat stream works: Test HeartbeatService.StreamHeartbeat
- [ ] TLS handshake successful: `openssl s_client -connect <host>:8444`

### Performance Testing  
- [ ] Load test with expected RPS: Run gRPC load tests
- [ ] Memory usage stable: Monitor resource consumption
- [ ] Connection pooling effective: Check connection metrics
- [ ] No resource leaks: Monitor over extended period

### Security Testing
- [ ] TLS 1.3 enforced: Verify cipher suites
- [ ] mTLS working (if enabled): Test client certificates
- [ ] NetworkPolicy effective: Test blocked connections
- [ ] No plain text traffic: Network packet analysis

### Monitoring Validation
- [ ] All Prometheus targets UP: Check /targets page
- [ ] Grafana dashboard loaded: Verify gRPC dashboard displays data
- [ ] Alerts configured: Test alert firing with artificial load
- [ ] Log aggregation working: Verify structured gRPC logs

## Post-Deployment ☐

### Performance Optimization
- [ ] Review resource usage and adjust limits
- [ ] Tune HPA settings based on actual load patterns
- [ ] Optimize gRPC keepalive settings
- [ ] Enable compression if beneficial

### Documentation Updates
- [ ] Update operational runbooks
- [ ] Document new monitoring dashboards and alerts  
- [ ] Update client connection instructions
- [ ] Record lessons learned and known issues

### Training & Communication
- [ ] Train operations team on new gRPC endpoints
- [ ] Notify development teams of new gRPC API availability
- [ ] Update incident response procedures
- [ ] Communicate performance improvements to stakeholders

## Rollback Plan ☐

### Immediate Rollback (< 5 minutes)
- [ ] Scale canary to 0: `kubectl scale deployment cloudbridge-relay-canary --replicas=0`
- [ ] Redirect traffic: Update ingress canary-weight to 0
- [ ] Monitor error rate returns to baseline
- [ ] Alert operations team of rollback execution

### Full Rollback (if needed)
- [ ] Remove gRPC deployment: `kubectl delete -f k8s/canary/`
- [ ] Revert ConfigMap: Restore original configuration
- [ ] Restart main deployment: `kubectl rollout restart deployment/cloudbridge-relay`
- [ ] Verify system returns to previous stable state
- [ ] Document rollback reason and plan remediation

## Success Criteria ☐

### Technical Metrics
- [ ] Error rate: < 0.1%
- [ ] P95 Latency: < 1000ms  
- [ ] Availability: > 99.9%
- [ ] Resource usage: Within 10% of baseline

### Business Metrics
- [ ] No client connection issues reported
- [ ] Performance improvement measurable
- [ ] Successful migration of 100% traffic
- [ ] Zero unplanned downtime during rollout

## Sign-off ☐

- [ ] **Tech Lead**: Deployment technically sound
- [ ] **DevOps**: Infrastructure ready and monitored
- [ ] **Security**: Security requirements met
- [ ] **Operations**: Runbooks updated and team trained
- [ ] **Product**: Feature ready for client adoption

---

**Deployment Date**: ___________  
**Completed By**: ___________  
**Rollout Duration**: ___________  
**Issues Encountered**: ___________
