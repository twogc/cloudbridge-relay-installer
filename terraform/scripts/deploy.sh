#!/bin/bash

# CloudBridge Relay - Terraform Deployment Script
# This script deploys CloudBridge Relay using Terraform

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
TERRAFORM_DIR="terraform"
ENVIRONMENTS_DIR="$TERRAFORM_DIR/environments"
SCRIPTS_DIR="$TERRAFORM_DIR/scripts"

# Functions
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check if terraform is installed
    if ! command -v terraform &> /dev/null; then
        log_error "Terraform is not installed. Please install Terraform first."
        exit 1
    fi
    
    # Check if kubectl is installed
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl is not installed. Please install kubectl first."
        exit 1
    fi
    
    # Check if terraform directory exists
    if [ ! -d "$TERRAFORM_DIR" ]; then
        log_error "Terraform directory '$TERRAFORM_DIR' not found."
        exit 1
    fi
    
    log_success "Prerequisites check passed"
}

# Initialize Terraform
init_terraform() {
    local environment=$1
    local terraform_dir="$ENVIRONMENTS_DIR/$environment"
    
    log_info "Initializing Terraform for $environment environment..."
    
    cd "$terraform_dir"
    
    # Initialize Terraform
    terraform init
    
    # Validate configuration
    terraform validate
    
    cd - > /dev/null
    
    log_success "Terraform initialized for $environment"
}

# Plan Terraform
plan_terraform() {
    local environment=$1
    local terraform_dir="$ENVIRONMENTS_DIR/$environment"
    
    log_info "Planning Terraform for $environment environment..."
    
    cd "$terraform_dir"
    
    # Create plan
    terraform plan -out=tfplan
    
    cd - > /dev/null
    
    log_success "Terraform plan created for $environment"
}

# Apply Terraform
apply_terraform() {
    local environment=$1
    local terraform_dir="$ENVIRONMENTS_DIR/$environment"
    
    log_info "Applying Terraform for $environment environment..."
    
    cd "$terraform_dir"
    
    # Apply plan
    terraform apply tfplan
    
    cd - > /dev/null
    
    log_success "Terraform applied for $environment"
}

# Destroy Terraform
destroy_terraform() {
    local environment=$1
    local terraform_dir="$ENVIRONMENTS_DIR/$environment"
    
    log_warning "Destroying Terraform for $environment environment..."
    
    cd "$terraform_dir"
    
    # Destroy resources
    terraform destroy -auto-approve
    
    cd - > /dev/null
    
    log_success "Terraform destroyed for $environment"
}

# Show Terraform outputs
show_outputs() {
    local environment=$1
    local terraform_dir="$ENVIRONMENTS_DIR/$environment"
    
    log_info "Showing Terraform outputs for $environment environment..."
    
    cd "$terraform_dir"
    
    # Show outputs
    terraform output
    
    cd - > /dev/null
}

# Show Terraform state
show_state() {
    local environment=$1
    local terraform_dir="$ENVIRONMENTS_DIR/$environment"
    
    log_info "Showing Terraform state for $environment environment..."
    
    cd "$terraform_dir"
    
    # Show state
    terraform show
    
    cd - > /dev/null
}

# Validate Terraform configuration
validate_terraform() {
    local environment=$1
    local terraform_dir="$ENVIRONMENTS_DIR/$environment"
    
    log_info "Validating Terraform configuration for $environment environment..."
    
    cd "$terraform_dir"
    
    # Validate configuration
    terraform validate
    
    # Format check
    terraform fmt -check
    
    cd - > /dev/null
    
    log_success "Terraform configuration validated for $environment"
}

# Format Terraform files
format_terraform() {
    log_info "Formatting Terraform files..."
    
    # Format all Terraform files
    find "$TERRAFORM_DIR" -name "*.tf" -exec terraform fmt {} \;
    
    log_success "Terraform files formatted"
}

# Show deployment status
show_status() {
    local environment=$1
    
    log_info "Deployment Status for $environment environment:"
    echo ""
    
    echo "📊 Kubernetes Resources:"
    kubectl get pods -A | grep cloudbridge || echo "No CloudBridge pods found"
    echo ""
    
    echo "🌐 Services:"
    kubectl get services -A | grep cloudbridge || echo "No CloudBridge services found"
    echo ""
    
    echo "💾 Persistent Volumes:"
    kubectl get pv,pvc -A | grep cloudbridge || echo "No CloudBridge volumes found"
    echo ""
    
    echo "🔗 Ingress:"
    kubectl get ingress -A | grep cloudbridge || echo "No CloudBridge ingress found"
    echo ""
}

# Show access information
show_access_info() {
    local environment=$1
    
    log_info "Access Information for $environment environment:"
    echo ""
    
    echo "🚀 CloudBridge Relay:"
    echo "   - Internal: http://cloudbridge-relay.cloudbridge.svc.cluster.local:9090"
    echo "   - Health: http://cloudbridge-relay.cloudbridge.svc.cluster.local:9090/health"
    echo "   - Metrics: http://cloudbridge-relay.cloudbridge.svc.cluster.local:8084/metrics"
    echo ""
    
    echo "📊 Grafana Dashboard:"
    echo "   - Internal: http://grafana.cloudbridge-monitoring.svc.cluster.local:3000"
    echo "   - Username: admin"
    echo "   - Password: (from terraform outputs)"
    echo ""
    
    echo "📈 Prometheus:"
    echo "   - Internal: http://prometheus.cloudbridge-monitoring.svc.cluster.local:9090"
    echo ""
    
    echo "🔗 Port Forward Commands:"
    echo "   kubectl port-forward service/cloudbridge-relay 9090:9090 -n cloudbridge"
    echo "   kubectl port-forward service/grafana 3000:3000 -n cloudbridge-monitoring"
    echo "   kubectl port-forward service/prometheus 9090:9090 -n cloudbridge-monitoring"
    echo ""
}

# Main function
main() {
    echo "=== 🏗️ CloudBridge Relay - Terraform Deployment ==="
    echo ""
    
    local command=${1:-help}
    local environment=${2:-prod}
    
    case "$command" in
        "init")
            check_prerequisites
            init_terraform "$environment"
            ;;
        "plan")
            check_prerequisites
            plan_terraform "$environment"
            ;;
        "apply")
            check_prerequisites
            apply_terraform "$environment"
            show_outputs "$environment"
            show_status "$environment"
            show_access_info "$environment"
            ;;
        "destroy")
            check_prerequisites
            destroy_terraform "$environment"
            ;;
        "outputs")
            show_outputs "$environment"
            ;;
        "state")
            show_state "$environment"
            ;;
        "validate")
            check_prerequisites
            validate_terraform "$environment"
            ;;
        "format")
            format_terraform
            ;;
        "status")
            show_status "$environment"
            ;;
        "access")
            show_access_info "$environment"
            ;;
        "help"|"-h"|"--help")
            echo "Usage: $0 [command] [environment]"
            echo ""
            echo "Commands:"
            echo "  init      - Initialize Terraform"
            echo "  plan      - Create Terraform plan"
            echo "  apply     - Apply Terraform configuration"
            echo "  destroy   - Destroy Terraform resources"
            echo "  outputs   - Show Terraform outputs"
            echo "  state     - Show Terraform state"
            echo "  validate  - Validate Terraform configuration"
            echo "  format    - Format Terraform files"
            echo "  status    - Show deployment status"
            echo "  access    - Show access information"
            echo "  help      - Show this help message"
            echo ""
            echo "Environments:"
            echo "  dev       - Development environment"
            echo "  staging   - Staging environment"
            echo "  prod      - Production environment (default)"
            echo ""
            echo "Examples:"
            echo "  $0 init prod"
            echo "  $0 plan prod"
            echo "  $0 apply prod"
            echo "  $0 status prod"
            ;;
        *)
            log_error "Unknown command: $command"
            echo "Use '$0 help' for usage information"
            exit 1
            ;;
    esac
}

# Run main function
main "$@"



