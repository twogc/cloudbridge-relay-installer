#!/bin/bash

# CloudBridge Relay - Advanced Monitoring Deployment Script
# This script deploys Prometheus, Grafana, and Alertmanager for monitoring

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
NAMESPACE="cloudbridge-monitoring"
MONITORING_DIR="monitoring"

# Functions
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}$1${NC}"
}

log_warning() {
    echo -e "${YELLOW}  $1${NC}"
}

log_error() {
    echo -e "${RED}$1${NC}"
}

check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check if kubectl is installed
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl is not installed. Please install kubectl first."
        exit 1
    fi
    
    # Check if monitoring directory exists
    if [ ! -d "$MONITORING_DIR" ]; then
        log_error "Monitoring directory '$MONITORING_DIR' not found."
        exit 1
    fi
    
    log_success "Prerequisites check passed"
}

create_namespace() {
    log_info "Creating monitoring namespace..."
    
    kubectl create namespace $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -
    
    log_success "Namespace created"
}

create_configmaps() {
    log_info "Creating ConfigMaps..."
    
    # Prometheus configuration
    kubectl create configmap prometheus-config \
        --from-file=$MONITORING_DIR/prometheus/prometheus.yml \
        --namespace=$NAMESPACE \
        --dry-run=client -o yaml | kubectl apply -f -
    
    # Alertmanager configuration
    kubectl create configmap alertmanager-config \
        --from-file=$MONITORING_DIR/alertmanager/alertmanager.yml \
        --namespace=$NAMESPACE \
        --dry-run=client -o yaml | kubectl apply -f -
    
    # Grafana datasources
    kubectl create configmap grafana-datasources \
        --from-file=$MONITORING_DIR/grafana/datasources.yml \
        --namespace=$NAMESPACE \
        --dry-run=client -o yaml | kubectl apply -f -
    
    # Grafana dashboards
    kubectl create configmap grafana-dashboards \
        --from-file=$MONITORING_DIR/grafana/dashboards.yml \
        --namespace=$NAMESPACE \
        --dry-run=client -o yaml | kubectl apply -f -
    
    # Alert rules
    kubectl create configmap prometheus-rules \
        --from-file=$MONITORING_DIR/rules/ \
        --namespace=$NAMESPACE \
        --dry-run=client -o yaml | kubectl apply -f -
    
    log_success "ConfigMaps created"
}

create_secrets() {
    log_info "Creating Secrets..."
    
    # Grafana admin password
    kubectl create secret generic grafana-secrets \
        --from-literal=admin-password=admin123 \
        --namespace=$NAMESPACE \
        --dry-run=client -o yaml | kubectl apply -f -
    
    # Prometheus basic auth
    kubectl create secret generic prometheus-secrets \
        --from-literal=basic-auth=admin:admin123 \
        --namespace=$NAMESPACE \
        --dry-run=client -o yaml | kubectl apply -f -
    
    log_success "Secrets created"
}

deploy_prometheus() {
    log_info "Deploying Prometheus..."
    
    # Create Prometheus deployment
    cat <<EOF | kubectl apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: prometheus
  namespace: $NAMESPACE
  labels:
    app: prometheus
spec:
  replicas: 1
  selector:
    matchLabels:
      app: prometheus
  template:
    metadata:
      labels:
        app: prometheus
    spec:
      serviceAccountName: prometheus
      containers:
      - name: prometheus
        image: prom/prometheus:latest
        ports:
        - containerPort: 9090
        args:
        - '--config.file=/etc/prometheus/prometheus.yml'
        - '--storage.tsdb.path=/prometheus'
        - '--web.console.libraries=/etc/prometheus/console_libraries'
        - '--web.console.templates=/etc/prometheus/consoles'
        - '--storage.tsdb.retention.time=200h'
        - '--web.enable-lifecycle'
        - '--web.enable-admin-api'
        volumeMounts:
        - name: prometheus-config
          mountPath: /etc/prometheus
        - name: prometheus-rules
          mountPath: /etc/prometheus/rules
        resources:
          requests:
            memory: "256Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "250m"
      volumes:
      - name: prometheus-config
        configMap:
          name: prometheus-config
      - name: prometheus-rules
        configMap:
          name: prometheus-rules
---
apiVersion: v1
kind: Service
metadata:
  name: prometheus
  namespace: $NAMESPACE
  labels:
    app: prometheus
spec:
  type: ClusterIP
  ports:
  - port: 9090
    targetPort: 9090
  selector:
    app: prometheus
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: prometheus
  namespace: $NAMESPACE
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: prometheus
rules:
- apiGroups: [""]
  resources: ["nodes", "nodes/proxy", "services", "endpoints", "pods"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["extensions"]
  resources: ["ingresses"]
  verbs: ["get", "list", "watch"]
- nonResourceURLs: ["/metrics"]
  verbs: ["get"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: prometheus
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: prometheus
subjects:
- kind: ServiceAccount
  name: prometheus
  namespace: $NAMESPACE
EOF
    
    log_success "Prometheus deployed"
}

deploy_grafana() {
    log_info "Deploying Grafana..."
    
    # Create Grafana deployment
    cat <<EOF | kubectl apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: grafana
  namespace: $NAMESPACE
  labels:
    app: grafana
spec:
  replicas: 1
  selector:
    matchLabels:
      app: grafana
  template:
    metadata:
      labels:
        app: grafana
    spec:
      serviceAccountName: grafana
      containers:
      - name: grafana
        image: grafana/grafana:latest
        ports:
        - containerPort: 3000
        env:
        - name: GF_SECURITY_ADMIN_PASSWORD
          valueFrom:
            secretKeyRef:
              name: grafana-secrets
              key: admin-password
        - name: GF_USERS_ALLOW_SIGN_UP
          value: "false"
        - name: GF_INSTALL_PLUGINS
          value: "grafana-piechart-panel"
        volumeMounts:
        - name: grafana-datasources
          mountPath: /etc/grafana/provisioning/datasources
        - name: grafana-dashboards
          mountPath: /etc/grafana/provisioning/dashboards
        resources:
          requests:
            memory: "128Mi"
            cpu: "50m"
          limits:
            memory: "256Mi"
            cpu: "100m"
      volumes:
      - name: grafana-datasources
        configMap:
          name: grafana-datasources
      - name: grafana-dashboards
        configMap:
          name: grafana-dashboards
---
apiVersion: v1
kind: Service
metadata:
  name: grafana
  namespace: $NAMESPACE
  labels:
    app: grafana
spec:
  type: ClusterIP
  ports:
  - port: 3000
    targetPort: 3000
  selector:
    app: grafana
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: grafana
  namespace: $NAMESPACE
EOF
    
    log_success "Grafana deployed"
}

deploy_alertmanager() {
    log_info "Deploying Alertmanager..."
    
    # Create Alertmanager deployment
    cat <<EOF | kubectl apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: alertmanager
  namespace: $NAMESPACE
  labels:
    app: alertmanager
spec:
  replicas: 1
  selector:
    matchLabels:
      app: alertmanager
  template:
    metadata:
      labels:
        app: alertmanager
    spec:
      containers:
      - name: alertmanager
        image: prom/alertmanager:latest
        ports:
        - containerPort: 9093
        args:
        - '--config.file=/etc/alertmanager/alertmanager.yml'
        - '--storage.path=/alertmanager'
        - '--web.external-url=http://localhost:9093'
        volumeMounts:
        - name: alertmanager-config
          mountPath: /etc/alertmanager
        resources:
          requests:
            memory: "128Mi"
            cpu: "50m"
          limits:
            memory: "256Mi"
            cpu: "100m"
      volumes:
      - name: alertmanager-config
        configMap:
          name: alertmanager-config
---
apiVersion: v1
kind: Service
metadata:
  name: alertmanager
  namespace: $NAMESPACE
  labels:
    app: alertmanager
spec:
  type: ClusterIP
  ports:
  - port: 9093
    targetPort: 9093
  selector:
    app: alertmanager
EOF
    
    log_success "Alertmanager deployed"
}

show_status() {
    log_info "Monitoring Status:"
    echo ""
    
    echo "Pods Status:"
    kubectl get pods -n $NAMESPACE
    echo ""
    
    echo "Services Status:"
    kubectl get services -n $NAMESPACE
    echo ""
    
    echo "ConfigMaps Status:"
    kubectl get configmaps -n $NAMESPACE
    echo ""
    
    echo "Secrets Status:"
    kubectl get secrets -n $NAMESPACE
    echo ""
}

show_access_info() {
    log_info "Access Information:"
    echo ""
    
    echo "Grafana Dashboard:"
    echo "   - Internal: http://grafana.$NAMESPACE.svc.cluster.local:3000"
    echo "   - Username: admin"
    echo "   - Password: admin123"
    echo ""
    
    echo " Prometheus:"
    echo "   - Internal: http://prometheus.$NAMESPACE.svc.cluster.local:9090"
    echo ""
    
    echo "Alertmanager:"
    echo "   - Internal: http://alertmanager.$NAMESPACE.svc.cluster.local:9093"
    echo ""
    
    echo " Port Forward Commands:"
    echo "   kubectl port-forward service/grafana 3000:3000 -n $NAMESPACE"
    echo "   kubectl port-forward service/prometheus 9090:9090 -n $NAMESPACE"
    echo "   kubectl port-forward service/alertmanager 9093:9093 -n $NAMESPACE"
    echo ""
}

cleanup() {
    log_warning "Cleaning up monitoring deployment..."
    
    # Delete deployments
    kubectl delete deployment prometheus grafana alertmanager -n $NAMESPACE --ignore-not-found=true
    
    # Delete services
    kubectl delete service prometheus grafana alertmanager -n $NAMESPACE --ignore-not-found=true
    
    # Delete configmaps
    kubectl delete configmap prometheus-config alertmanager-config grafana-datasources grafana-dashboards prometheus-rules -n $NAMESPACE --ignore-not-found=true
    
    # Delete secrets
    kubectl delete secret grafana-secrets prometheus-secrets -n $NAMESPACE --ignore-not-found=true
    
    # Delete service accounts
    kubectl delete serviceaccount prometheus grafana -n $NAMESPACE --ignore-not-found=true
    
    # Delete RBAC
    kubectl delete clusterrole prometheus --ignore-not-found=true
    kubectl delete clusterrolebinding prometheus --ignore-not-found=true
    
    # Delete namespace
    kubectl delete namespace $NAMESPACE --ignore-not-found=true
    
    log_success "Cleanup completed"
}

# Main function
main() {
    echo "=== CloudBridge Relay - Advanced Monitoring Deployment ==="
    echo ""
    
    case "${1:-deploy}" in
        "deploy")
            check_prerequisites
            create_namespace
            create_configmaps
            create_secrets
            deploy_prometheus
            deploy_grafana
            deploy_alertmanager
            show_status
            show_access_info
            log_success "Monitoring deployment completed successfully!"
            ;;
        "status")
            show_status
            ;;
        "cleanup")
            cleanup
            ;;
        "help"|"-h"|"--help")
            echo "Usage: $0 [deploy|status|cleanup|help]"
            echo ""
            echo "Commands:"
            echo "  deploy   - Deploy monitoring stack (default)"
            echo "  status   - Show monitoring status"
            echo "  cleanup  - Remove monitoring resources"
            echo "  help     - Show this help message"
            ;;
        *)
            log_error "Unknown command: $1"
            echo "Use '$0 help' for usage information"
            exit 1
            ;;
    esac
}

# Run main function
main "$@"

