# CloudBridge Relay - Staging Environment Variables
# Copy this file to terraform.tfvars and update the values

# Kubernetes Configuration
kubeconfig_path = "~/.kube/config"

# Application Configuration
app_version = "1.0.0"
terraform_version = "1.0.0"
image_registry = "ghcr.io/twogc/cloudbridge-relay-installer"

# Security (Change these in production!)
cassandra_password = "staging-cassandra-password-change-me"
grafana_password = "staging-grafana-password-change-me"
api_key = "staging-api-key-change-me"
jwt_secret = "staging-jwt-secret-change-me"

# Ingress Configuration
ingress_host = "staging.cloudbridge.local"
