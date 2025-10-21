# 🌐 CloudBridge BGP Monitoring

## 📊 Overview

This directory contains Kubernetes manifests for BGP monitoring in the CloudBridge anycast network. The monitoring system provides comprehensive visibility into BGP sessions, route tables, and anycast health across all PoP locations.

## 🏗️ Architecture

### **BGP Exporters**
- **BIRD Exporter**: Monitors BIRD routing daemon (port 9324)
- **FRR Exporter**: Monitors FRR routing daemon (port 9342)  
- **GoBGP Exporter**: Monitors GoBGP routing daemon (port 9474)

### **Monitoring Stack**
- **Prometheus**: Metrics collection and storage
- **Grafana**: Visualization and dashboards
- **AlertManager**: Alert routing and notifications

## 📁 Files

### **Deployments**
- `bird-exporter-deployment.yaml` - BIRD exporter deployment
- `frr-exporter-deployment.yaml` - FRR exporter deployment
- `gobgp-exporter-deployment.yaml` - GoBGP exporter deployment

### **Services**
- `bird-exporter-service.yaml` - BIRD exporter service
- `frr-exporter-service.yaml` - FRR exporter service
- `gobgp-exporter-service.yaml` - GoBGP exporter service

### **Monitoring**
- `bird-exporter-servicemonitor.yaml` - BIRD ServiceMonitor
- `frr-exporter-servicemonitor.yaml` - FRR ServiceMonitor
- `gobgp-exporter-servicemonitor.yaml` - GoBGP ServiceMonitor

### **RBAC**
- `bird-exporter-rbac.yaml` - BIRD exporter RBAC
- `bgp-exporters-rbac.yaml` - FRR/GoBGP exporters RBAC

### **Alerts**
- `bgp-alerts.yaml` - BGP monitoring alerts

### **Dashboards**
- `bgp-overview-dashboard.json` - BGP overview dashboard
- `anycast-health-dashboard.json` - Anycast health dashboard

## 🚀 Quick Start

### **1. Deploy BGP Monitoring**
```bash
# Deploy all BGP monitoring components
./scripts/deploy-bgp-monitoring.sh
```

### **2. Verify Deployment**
```bash
# Check deployments
kubectl get deployments -n monitoring -l component=bgp-monitoring

# Check services
kubectl get services -n monitoring -l component=bgp-monitoring

# Check ServiceMonitors
kubectl get servicemonitors -n monitoring -l component=bgp-monitoring
```

### **3. Access Metrics**
```bash
# BIRD Exporter
kubectl port-forward svc/bird-exporter 9324:9324 -n monitoring

# FRR Exporter  
kubectl port-forward svc/frr-exporter 9342:9342 -n monitoring

# GoBGP Exporter
kubectl port-forward svc/gobgp-exporter 9474:9474 -n monitoring
```

## 📊 Metrics

### **BGP Session Metrics**
```yaml
# BIRD Metrics
bird_protocol_up{name="bgp1",proto="BGP"} 1
bird_protocol_uptime_seconds{name="bgp1",proto="BGP"} 86400
bird_protocol_route_changes_total{name="bgp1",proto="BGP"} 150

# FRR Metrics
frr_bgp_peer_up{name="peer1"} 1
frr_bgp_peer_uptime_seconds{name="peer1"} 86400
frr_bgp_peer_flop_count{name="peer1"} 0

# GoBGP Metrics
gobgp_peer_up{name="10.0.2.100"} 1
gobgp_peer_session_state{name="10.0.2.100"} 6  # established
gobgp_peer_uptime_seconds{name="10.0.2.100"} 86400
```

### **BGP Prefix Metrics**
```yaml
# BIRD Prefix Metrics
bird_protocol_prefix_count_import{name="bgp1",proto="BGP"} 500000
bird_protocol_prefix_count_export{name="bgp1",proto="BGP"} 1000
bird_protocol_prefix_count_filtered{name="bgp1",proto="BGP"} 50

# FRR Prefix Metrics
frr_bgp_peer_received_prefixes{name="peer1"} 500000
frr_bgp_peer_advertised_prefixes{name="peer1"} 1000
frr_bgp_peer_accepted_prefixes{name="peer1"} 499950

# GoBGP Route Metrics
gobgp_route_total_destination_count{address_family="ipv4"} 500000
gobgp_route_total_path_count{address_family="ipv4"} 500000
gobgp_route_accepted_path_count{address_family="ipv4"} 499950
```

### **Anycast Health Metrics**
```yaml
# Anycast Health
anycast_health{pop="moscow",service="quic"} 1
anycast_health{pop="frankfurt",service="quic"} 1
anycast_health{pop="newyork",service="quic"} 1

# Anycast Performance
anycast_latency_seconds{pop="moscow",target="client"} 0.05
anycast_latency_seconds{pop="frankfurt",target="client"} 0.08
anycast_latency_seconds{pop="newyork",target="client"} 0.12

# Anycast Availability
anycast_availability{pop="moscow"} 99.9
anycast_availability{pop="frankfurt"} 99.8
anycast_availability{pop="newyork"} 99.7
```

## 🚨 Alerts

### **Critical Alerts**
- **BGPSessionDown**: BGP session is down for >1 minute
- **AnycastHealthDown**: Anycast health check failed
- **BGPExporterDown**: BGP exporter is not responding

### **Warning Alerts**
- **BGPSessionFlapping**: BGP session is flapping
- **BGPPrefixCountAnomaly**: BGP prefix count changed significantly
- **AnycastLatencyHigh**: Anycast latency is high
- **AnycastAvailabilityLow**: Anycast availability is low

## 📈 Dashboards

### **BGP Overview Dashboard**
- BGP session status across all PoPs
- BGP prefix counts (imported/exported)
- BGP session uptime
- BGP message counters
- BGP route changes

### **Anycast Health Dashboard**
- Anycast availability by PoP
- Anycast latency measurements
- Anycast performance heatmap
- Anycast health table

## 🔧 Configuration

### **BGP Daemon Requirements**
- **BIRD**: Socket at `/var/run/bird.ctl` and `/var/run/bird6.ctl`
- **FRR**: Socket directory at `/var/run/frr`
- **GoBGP**: gRPC API at `127.0.0.1:50051`

### **Kubernetes Requirements**
- **Namespace**: `monitoring`
- **RBAC**: ClusterRole and ClusterRoleBinding for exporters
- **ServiceMonitor**: Prometheus Operator for metric collection

### **Prometheus Configuration**
```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'bird-exporter'
    static_configs:
      - targets: ['bird-exporter:9324']
  - job_name: 'frr-exporter'
    static_configs:
      - targets: ['frr-exporter:9342']
  - job_name: 'gobgp-exporter'
    static_configs:
      - targets: ['gobgp-exporter:9474']
```

## 🛠️ Troubleshooting

### **Common Issues**

#### **1. BGP Exporter Not Scraping**
```bash
# Check if exporter is running
kubectl get pods -n monitoring -l component=bgp-monitoring

# Check exporter logs
kubectl logs -n monitoring deployment/bird-exporter
kubectl logs -n monitoring deployment/frr-exporter
kubectl logs -n monitoring deployment/gobgp-exporter
```

#### **2. BGP Daemon Socket Issues**
```bash
# Check BIRD socket
ls -la /var/run/bird.ctl

# Check FRR sockets
ls -la /var/run/frr/

# Check GoBGP gRPC
telnet 127.0.0.1 50051
```

#### **3. Prometheus Not Discovering Targets**
```bash
# Check ServiceMonitor
kubectl get servicemonitors -n monitoring

# Check Prometheus targets
kubectl port-forward svc/prometheus 9090:9090 -n monitoring
# Open http://localhost:9090/targets
```

## 📚 References

- [BIRD Exporter Documentation](https://github.com/czerwonk/bird_exporter)
- [FRR Exporter Documentation](https://github.com/tynany/frr_exporter)
- [GoBGP Exporter Documentation](https://github.com/greenpau/gobgp_exporter)
- [Prometheus Operator](https://github.com/prometheus-operator/prometheus-operator)
- [Grafana Dashboards](https://grafana.com/grafana/dashboards)

## 🎯 Next Steps

1. **Configure BGP daemons** on PoP nodes
2. **Set up anycast health checks** for services
3. **Import Grafana dashboards** for visualization
4. **Configure AlertManager** for notifications
5. **Set up BGP route optimization** based on metrics

---

**CloudBridge BGP Monitoring provides comprehensive visibility into your anycast network infrastructure, enabling proactive monitoring and optimization of routing performance.**





