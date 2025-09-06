# CloudBridge Relay - Staging Environment
terraform {
  required_version = ">= 1.0"
  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.0"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.0"
    }
  }
  
  # Local backend for development
  # backend "s3" {
  #   bucket         = "cloudbridge-terraform-state"
  #   key            = "staging/terraform.tfstate"
  #   region         = "us-west-2"
  #   encrypt        = true
  #   dynamodb_table = "terraform-locks"
  # }
}

# Provider configuration
provider "kubernetes" {
  config_path = var.kubeconfig_path
}

provider "helm" {
  kubernetes {
    config_path = var.kubeconfig_path
  }
}

# Local values
locals {
  environment = "staging"
  common_tags = {
    Environment = "staging"
    Project     = "cloudbridge-relay"
    ManagedBy   = "terraform"
    Owner       = "cloudbridge-team"
  }
}

# CloudBridge module
module "cloudbridge" {
  source = "../../modules/cloudbridge"
  
  # Environment
  environment = local.environment
  namespace   = "cloudbridge-staging"
  monitoring_namespace = "cloudbridge-staging-monitoring"
  
  # Application
  app_version      = var.app_version
  terraform_version = var.terraform_version
  image_registry   = var.image_registry
  log_level        = "debug"
  
  # Resource configuration - Staging settings (reduced resources)
  relay_replicas = 2
  relay_cpu_request = "125m"
  relay_cpu_limit = "250m"
  relay_memory_request = "256Mi"
  relay_memory_limit = "512Mi"
  
  cassandra_replicas = 1
  cassandra_cpu_request = "250m"
  cassandra_cpu_limit = "500m"
  cassandra_memory_request = "512Mi"
  cassandra_memory_limit = "1Gi"
  cassandra_storage_size = "20Gi"
  
  cadence_replicas = 1
  cadence_cpu_request = "250m"
  cadence_cpu_limit = "500m"
  cadence_memory_request = "512Mi"
  cadence_memory_limit = "1Gi"
  
  # Monitoring - Staging settings (reduced resources)
  prometheus_replicas = 1
  prometheus_cpu_request = "100m"
  prometheus_cpu_limit = "250m"
  prometheus_memory_request = "256Mi"
  prometheus_memory_limit = "512Mi"
  prometheus_storage_size = "10Gi"
  
  grafana_replicas = 1
  grafana_cpu_request = "50m"
  grafana_cpu_limit = "100m"
  grafana_memory_request = "128Mi"
  grafana_memory_limit = "256Mi"
  grafana_storage_size = "5Gi"
  
  # Security
  cassandra_password = var.cassandra_password
  grafana_password   = var.grafana_password
  api_key           = var.api_key
  jwt_secret        = var.jwt_secret
  
  # Features
  create_namespaces = true
  create_persistent_volumes = true
  enable_monitoring = true
  enable_security_policies = true
  enable_ingress = true
  
  # Ingress
  ingress_host = var.ingress_host
  ingress_class = "nginx"
  tls_secret_name = "cloudbridge-staging-tls"
  
  # Tags
  tags = local.common_tags
}

# Staging-specific HPA (more aggressive scaling for testing)
resource "kubernetes_horizontal_pod_autoscaler" "cloudbridge_relay_staging" {
  metadata {
    name      = "cloudbridge-relay-staging-hpa"
    namespace = module.cloudbridge.namespace
    labels = {
      app.kubernetes.io/name = "cloudbridge-relay"
      app.kubernetes.io/component = "autoscaling"
      environment = "staging"
    }
  }
  
  spec {
    scale_target_ref {
      api_version = "apps/v1"
      kind        = "Deployment"
      name        = "cloudbridge-relay"
    }
    
    min_replicas = 1
    max_replicas = 5
    
    metric {
      type = "Resource"
      resource {
        name = "cpu"
        target {
          type                = "Utilization"
          average_utilization = 60
        }
      }
    }
    
    metric {
      type = "Resource"
      resource {
        name = "memory"
        target {
          type                = "Utilization"
          average_utilization = 70
        }
      }
    }
  }
}

# Staging-specific Network Policies (more permissive for testing)
resource "kubernetes_network_policy" "cloudbridge_staging" {
  metadata {
    name      = "cloudbridge-staging-network-policy"
    namespace = module.cloudbridge.namespace
    labels = {
      app.kubernetes.io/name = "cloudbridge-relay"
      app.kubernetes.io/component = "security"
      environment = "staging"
    }
  }
  
  spec {
    pod_selector {
      match_labels = {
        "app.kubernetes.io/name" = "cloudbridge-relay"
      }
    }
    
    policy_types = ["Ingress", "Egress"]
    
    # More permissive ingress for staging
    ingress {
      from = []
      ports {
        protocol = "TCP"
        port     = "9090"
      }
      ports {
        protocol = "UDP"
        port     = "9090"
      }
      ports {
        protocol = "UDP"
        port     = "51820"
      }
      ports {
        protocol = "TCP"
        port     = "8084"
      }
    }
    
    # More permissive egress for staging
    egress {
      to = []
      ports {
        protocol = "TCP"
        port     = "443"
      }
      ports {
        protocol = "TCP"
        port     = "80"
      }
      ports {
        protocol = "UDP"
        port     = "53"
      }
      ports {
        protocol = "TCP"
        port     = "53"
      }
    }
  }
}

# Staging-specific ConfigMap for testing
resource "kubernetes_config_map" "staging_config" {
  metadata {
    name      = "staging-test-config"
    namespace = module.cloudbridge.namespace
    labels = {
      app.kubernetes.io/name = "cloudbridge-relay"
      environment = "staging"
    }
  }
  
  data = {
    staging_mode = "true"
    debug_mode = "true"
    test_data_enabled = "true"
    mock_external_services = "true"
    performance_testing = "true"
  }
}
