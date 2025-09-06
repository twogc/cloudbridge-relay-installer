#!/bin/bash

# CloudBridge Relay - Velero Setup Script
# Этот скрипт устанавливает и настраивает Velero для backup и disaster recovery

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# Configuration
VELERO_VERSION="${VELERO_VERSION:-v1.12.0}"
VELERO_NAMESPACE="${VELERO_NAMESPACE:-velero}"
BACKUP_STORAGE_PROVIDER="${BACKUP_STORAGE_PROVIDER:-aws}"  # aws, gcp, azure, local
BACKUP_BUCKET="${BACKUP_BUCKET:-cloudbridge-velero-backups}"
BACKUP_REGION="${BACKUP_REGION:-us-west-2}"
BACKUP_PREFIX="${BACKUP_PREFIX:-velero}"

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_section() {
    echo -e "${PURPLE}[SECTION]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

# Check dependencies
check_dependencies() {
    log_info "Checking dependencies..."
    
    local missing_deps=()
    
    for tool in kubectl curl; do
        if ! command -v "$tool" &> /dev/null; then
            missing_deps+=("$tool")
        fi
    done
    
    if [ ${#missing_deps[@]} -gt 0 ]; then
        log_error "Missing dependencies: ${missing_deps[*]}"
        exit 1
    fi
    
    # Check if kubectl can connect to cluster
    if ! kubectl cluster-info &>/dev/null; then
        log_error "Cannot connect to Kubernetes cluster"
        exit 1
    fi
    
    log_success "Dependencies check passed"
}

# Download and install Velero CLI
install_velero_cli() {
    log_section "Installing Velero CLI..."
    
    if command -v velero &> /dev/null; then
        local current_version=$(velero version --client-only --short 2>/dev/null | cut -d' ' -f2 || echo "unknown")
        log_info "Velero CLI already installed: $current_version"
        
        if [ "$current_version" = "$VELERO_VERSION" ]; then
            log_success "Velero CLI version matches required version: $VELERO_VERSION"
            return 0
        else
            log_warning "Velero CLI version mismatch. Current: $current_version, Required: $VELERO_VERSION"
        fi
    fi
    
    log_info "Downloading Velero CLI $VELERO_VERSION..."
    
    # Detect OS and architecture
    local os=$(uname -s | tr '[:upper:]' '[:lower:]')
    local arch=$(uname -m)
    
    case $arch in
        x86_64) arch="amd64" ;;
        arm64) arch="arm64" ;;
        *) log_error "Unsupported architecture: $arch"; exit 1 ;;
    esac
    
    local download_url="https://github.com/vmware-tanzu/velero/releases/download/$VELERO_VERSION/velero-$VELERO_VERSION-$os-$arch.tar.gz"
    
    # Download and extract
    curl -L "$download_url" -o "/tmp/velero-$VELERO_VERSION.tar.gz"
    tar -xzf "/tmp/velero-$VELERO_VERSION.tar.gz" -C /tmp
    
    # Install binary
    sudo mv "/tmp/velero-$VELERO_VERSION-$os-$arch/velero" /usr/local/bin/
    sudo chmod +x /usr/local/bin/velero
    
    # Cleanup
    rm -rf "/tmp/velero-$VELERO_VERSION.tar.gz" "/tmp/velero-$VELERO_VERSION-$os-$arch"
    
    log_success "Velero CLI installed successfully"
}

# Setup cloud credentials
setup_cloud_credentials() {
    log_section "Setting up cloud credentials..."
    
    case $BACKUP_STORAGE_PROVIDER in
        aws)
            setup_aws_credentials
            ;;
        gcp)
            setup_gcp_credentials
            ;;
        azure)
            setup_azure_credentials
            ;;
        local)
            setup_local_storage
            ;;
        *)
            log_error "Unsupported backup storage provider: $BACKUP_STORAGE_PROVIDER"
            exit 1
            ;;
    esac
}

# Setup AWS credentials
setup_aws_credentials() {
    log_info "Setting up AWS credentials..."
    
    # Check if AWS CLI is installed
    if ! command -v aws &> /dev/null; then
        log_error "AWS CLI is required but not installed"
        exit 1
    fi
    
    # Check if AWS credentials are configured
    if ! aws sts get-caller-identity &>/dev/null; then
        log_error "AWS credentials not configured. Please run 'aws configure'"
        exit 1
    fi
    
    # Create S3 bucket if it doesn't exist
    log_info "Creating S3 bucket: $BACKUP_BUCKET"
    aws s3 mb "s3://$BACKUP_BUCKET" --region "$BACKUP_REGION" 2>/dev/null || log_warning "Bucket may already exist"
    
    # Create IAM policy for Velero
    log_info "Creating IAM policy for Velero..."
    cat > /tmp/velero-policy.json <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "ec2:DescribeVolumes",
                "ec2:DescribeSnapshots",
                "ec2:CreateTags",
                "ec2:CreateVolume",
                "ec2:CreateSnapshot",
                "ec2:DeleteSnapshot"
            ],
            "Resource": "*"
        },
        {
            "Effect": "Allow",
            "Action": [
                "s3:GetObject",
                "s3:DeleteObject",
                "s3:PutObject",
                "s3:AbortMultipartUpload",
                "s3:ListMultipartUploadParts"
            ],
            "Resource": "arn:aws:s3:::$BACKUP_BUCKET/*"
        },
        {
            "Effect": "Allow",
            "Action": [
                "s3:ListBucket"
            ],
            "Resource": "arn:aws:s3:::$BACKUP_BUCKET"
        }
    ]
}
EOF
    
    # Create IAM user for Velero
    local velero_user="velero-user"
    log_info "Creating IAM user: $velero_user"
    aws iam create-user --user-name "$velero_user" 2>/dev/null || log_warning "User may already exist"
    
    # Attach policy to user
    aws iam put-user-policy --user-name "$velero_user" --policy-name "VeleroBackupPolicy" --policy-document file:///tmp/velero-policy.json
    
    # Create access key
    log_info "Creating access key for Velero user..."
    local access_key_output=$(aws iam create-access-key --user-name "$velero_user" --output json)
    local access_key_id=$(echo "$access_key_output" | jq -r '.AccessKey.AccessKeyId')
    local secret_access_key=$(echo "$access_key_output" | jq -r '.AccessKey.SecretAccessKey')
    
    # Create credentials file
    cat > /tmp/credentials <<EOF
[default]
aws_access_key_id=$access_key_id
aws_secret_access_key=$secret_access_key
EOF
    
    # Create Kubernetes secret
    kubectl create secret generic cloud-credentials \
        --from-file=cloud=/tmp/credentials \
        --namespace="$VELERO_NAMESPACE" \
        --dry-run=client -o yaml | kubectl apply -f -
    
    # Cleanup
    rm -f /tmp/velero-policy.json /tmp/credentials
    
    log_success "AWS credentials configured"
}

# Setup GCP credentials
setup_gcp_credentials() {
    log_info "Setting up GCP credentials..."
    
    # Check if gcloud CLI is installed
    if ! command -v gcloud &> /dev/null; then
        log_error "gcloud CLI is required but not installed"
        exit 1
    fi
    
    # Check if authenticated
    if ! gcloud auth list --filter=status:ACTIVE --format="value(account)" | grep -q .; then
        log_error "GCP authentication required. Please run 'gcloud auth login'"
        exit 1
    fi
    
    # Create GCS bucket if it doesn't exist
    log_info "Creating GCS bucket: $BACKUP_BUCKET"
    gsutil mb "gs://$BACKUP_BUCKET" 2>/dev/null || log_warning "Bucket may already exist"
    
    # Create service account for Velero
    local service_account="velero-sa"
    local project_id=$(gcloud config get-value project)
    
    log_info "Creating service account: $service_account"
    gcloud iam service-accounts create "$service_account" --display-name="Velero Service Account" 2>/dev/null || log_warning "Service account may already exist"
    
    # Grant necessary permissions
    gcloud projects add-iam-policy-binding "$project_id" \
        --member="serviceAccount:$service_account@$project_id.iam.gserviceaccount.com" \
        --role="roles/compute.storageAdmin"
    
    gcloud projects add-iam-policy-binding "$project_id" \
        --member="serviceAccount:$service_account@$project_id.iam.gserviceaccount.com" \
        --role="roles/storage.admin"
    
    # Create and download key
    log_info "Creating service account key..."
    gcloud iam service-accounts keys create /tmp/velero-key.json \
        --iam-account="$service_account@$project_id.iam.gserviceaccount.com"
    
    # Create Kubernetes secret
    kubectl create secret generic cloud-credentials \
        --from-file=cloud=/tmp/velero-key.json \
        --namespace="$VELERO_NAMESPACE" \
        --dry-run=client -o yaml | kubectl apply -f -
    
    # Cleanup
    rm -f /tmp/velero-key.json
    
    log_success "GCP credentials configured"
}

# Setup Azure credentials
setup_azure_credentials() {
    log_info "Setting up Azure credentials..."
    
    # Check if Azure CLI is installed
    if ! command -v az &> /dev/null; then
        log_error "Azure CLI is required but not installed"
        exit 1
    fi
    
    # Check if authenticated
    if ! az account show &>/dev/null; then
        log_error "Azure authentication required. Please run 'az login'"
        exit 1
    fi
    
    log_warning "Azure setup not fully implemented. Please configure manually."
    log_info "Create a storage account and container, then create credentials file:"
    log_info "  az storage account create --name <storage-account> --resource-group <resource-group>"
    log_info "  az storage container create --name <container> --account-name <storage-account>"
    
    # Create placeholder credentials file
    cat > /tmp/credentials <<EOF
AZURE_SUBSCRIPTION_ID=your-subscription-id
AZURE_TENANT_ID=your-tenant-id
AZURE_CLIENT_ID=your-client-id
AZURE_CLIENT_SECRET=your-client-secret
AZURE_RESOURCE_GROUP=your-resource-group
AZURE_CLOUD_NAME=AzurePublicCloud
EOF
    
    # Create Kubernetes secret
    kubectl create secret generic cloud-credentials \
        --from-file=cloud=/tmp/credentials \
        --namespace="$VELERO_NAMESPACE" \
        --dry-run=client -o yaml | kubectl apply -f -
    
    # Cleanup
    rm -f /tmp/credentials
    
    log_success "Azure credentials placeholder created (manual configuration required)"
}

# Setup local storage
setup_local_storage() {
    log_info "Setting up local storage..."
    
    # Create local backup directory
    local backup_dir="/opt/cloudbridge/velero-backups"
    sudo mkdir -p "$backup_dir"
    sudo chmod 755 "$backup_dir"
    
    # Create local storage configuration
    cat > /tmp/velero-local-config.yaml <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: velero-local-config
  namespace: $VELERO_NAMESPACE
data:
  backup-location: |
    provider: local
    config:
      path: $backup_dir
EOF
    
    kubectl apply -f /tmp/velero-local-config.yaml
    rm -f /tmp/velero-local-config.yaml
    
    log_success "Local storage configured"
}

# Install Velero server
install_velero_server() {
    log_section "Installing Velero server..."
    
    # Create namespace
    kubectl create namespace "$VELERO_NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -
    
    # Install Velero with appropriate provider
    case $BACKUP_STORAGE_PROVIDER in
        aws)
            velero install \
                --provider aws \
                --plugins velero/velero-plugin-for-aws:v1.7.0 \
                --bucket "$BACKUP_BUCKET" \
                --prefix "$BACKUP_PREFIX" \
                --secret-file /tmp/credentials \
                --backup-location-config region="$BACKUP_REGION" \
                --snapshot-location-config region="$BACKUP_REGION" \
                --namespace "$VELERO_NAMESPACE"
            ;;
        gcp)
            velero install \
                --provider gcp \
                --plugins velero/velero-plugin-for-gcp:v1.7.0 \
                --bucket "$BACKUP_BUCKET" \
                --prefix "$BACKUP_PREFIX" \
                --secret-file /tmp/velero-key.json \
                --backup-location-config region="$BACKUP_REGION" \
                --snapshot-location-config region="$BACKUP_REGION" \
                --namespace "$VELERO_NAMESPACE"
            ;;
        azure)
            velero install \
                --provider azure \
                --plugins velero/velero-plugin-for-microsoft-azure:v1.7.0 \
                --bucket "$BACKUP_BUCKET" \
                --prefix "$BACKUP_PREFIX" \
                --secret-file /tmp/credentials \
                --backup-location-config resourceGroup="$BACKUP_REGION" \
                --snapshot-location-config resourceGroup="$BACKUP_REGION" \
                --namespace "$VELERO_NAMESPACE"
            ;;
        local)
            velero install \
                --provider local \
                --bucket "$BACKUP_BUCKET" \
                --prefix "$BACKUP_PREFIX" \
                --backup-location-config path="/opt/cloudbridge/velero-backups" \
                --namespace "$VELERO_NAMESPACE"
            ;;
    esac
    
    log_success "Velero server installed"
}

# Apply custom configurations
apply_custom_configurations() {
    log_section "Applying custom configurations..."
    
    # Apply Velero configurations from k8s directory
    if [ -f "./k8s/velero/velero-install.yaml" ]; then
        log_info "Applying custom Velero configurations..."
        kubectl apply -f ./k8s/velero/velero-install.yaml
    fi
    
    log_success "Custom configurations applied"
}

# Verify installation
verify_installation() {
    log_section "Verifying Velero installation..."
    
    # Wait for Velero to be ready
    log_info "Waiting for Velero to be ready..."
    kubectl wait --for=condition=available --timeout=300s deployment/velero -n "$VELERO_NAMESPACE"
    
    # Check Velero status
    log_info "Checking Velero status..."
    velero version --namespace="$VELERO_NAMESPACE"
    
    # Check backup storage locations
    log_info "Checking backup storage locations..."
    velero backup-location get --namespace="$VELERO_NAMESPACE"
    
    # Check volume snapshot locations
    log_info "Checking volume snapshot locations..."
    velero snapshot-location get --namespace="$VELERO_NAMESPACE"
    
    # Test backup creation
    log_info "Testing backup creation..."
    local test_backup="test-backup-$(date +%Y%m%d-%H%M%S)"
    velero backup create "$test_backup" --include-namespaces kube-system --namespace="$VELERO_NAMESPACE"
    
    # Wait for backup to complete
    velero backup get "$test_backup" --namespace="$VELERO_NAMESPACE"
    
    # Cleanup test backup
    velero backup delete "$test_backup" --confirm --namespace="$VELERO_NAMESPACE"
    
    log_success "Velero installation verified successfully"
}

# Create backup schedules
create_backup_schedules() {
    log_section "Creating backup schedules..."
    
    # Daily backup schedule
    log_info "Creating daily backup schedule..."
    velero schedule create daily-backup \
        --schedule="0 2 * * *" \
        --include-namespaces cloudbridge,monitoring \
        --ttl 720h0m0s \
        --namespace="$VELERO_NAMESPACE"
    
    # Weekly backup schedule
    log_info "Creating weekly backup schedule..."
    velero schedule create weekly-backup \
        --schedule="0 3 * * 0" \
        --include-namespaces cloudbridge,monitoring \
        --ttl 2160h0m0s \
        --namespace="$VELERO_NAMESPACE"
    
    # Monthly backup schedule
    log_info "Creating monthly backup schedule..."
    velero schedule create monthly-backup \
        --schedule="0 4 1 * *" \
        --include-namespaces cloudbridge,monitoring \
        --ttl 8760h0m0s \
        --namespace="$VELERO_NAMESPACE"
    
    log_success "Backup schedules created"
}

# Main function
main() {
    log_info "Starting Velero setup for CloudBridge Relay..."
    
    check_dependencies
    install_velero_cli
    setup_cloud_credentials
    install_velero_server
    apply_custom_configurations
    verify_installation
    create_backup_schedules
    
    log_success "Velero setup completed successfully!"
    log_info ""
    log_info "Velero is now configured with:"
    log_info "  - Provider: $BACKUP_STORAGE_PROVIDER"
    log_info "  - Bucket: $BACKUP_BUCKET"
    log_info "  - Region: $BACKUP_REGION"
    log_info "  - Namespace: $VELERO_NAMESPACE"
    log_info ""
    log_info "Backup schedules created:"
    log_info "  - Daily backup at 2 AM (30 days retention)"
    log_info "  - Weekly backup on Sunday at 3 AM (90 days retention)"
    log_info "  - Monthly backup on 1st at 4 AM (1 year retention)"
    log_info ""
    log_info "Useful commands:"
    log_info "  - Check backups: velero backup get"
    log_info "  - Check schedules: velero schedule get"
    log_info "  - Create manual backup: velero backup create <name> --include-namespaces cloudbridge"
    log_info "  - Restore backup: velero restore create <name> --from-backup <backup-name>"
}

# Run main function
main "$@"

