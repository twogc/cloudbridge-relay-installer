# CloudBridge Relay - Production Environment
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
  
  backend "s3" {
    bucket         = "cloudbridge-terraform-state"
    key            = "prod/terraform.tfstate"
    region         = "us-west-2"
    encrypt        = true
    dynamodb_table = "terraform-locks"
  }
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
  environment = "prod"
  common_tags = {
    Environment = "production"
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
  namespace   = "cloudbridge"
  monitoring_namespace = "cloudbridge-monitoring"
  
  # Application
  app_version      = var.app_version
  terraform_version = var.terraform_version
  image_registry   = var.image_registry
  log_level        = "info"
  
  # Resource configuration - Production settings
  relay_replicas = 5
  relay_cpu_request = "500m"
  relay_cpu_limit = "1000m"
  relay_memory_request = "1Gi"
  relay_memory_limit = "2Gi"
  
  cassandra_replicas = 3
  cassandra_cpu_request = "500m"
  cassandra_cpu_limit = "1000m"
  cassandra_memory_request = "1Gi"
  cassandra_memory_limit = "2Gi"
  cassandra_storage_size = "100Gi"
  
  cadence_replicas = 2
  cadence_cpu_request = "500m"
  cadence_cpu_limit = "1000m"
  cadence_memory_request = "1Gi"
  cadence_memory_limit = "2Gi"
  
  # Monitoring - Production settings
  prometheus_replicas = 2
  prometheus_cpu_request = "200m"
  prometheus_cpu_limit = "500m"
  prometheus_memory_request = "512Mi"
  prometheus_memory_limit = "1Gi"
  prometheus_storage_size = "50Gi"
  
  grafana_replicas = 2
  grafana_cpu_request = "100m"
  grafana_cpu_limit = "200m"
  grafana_memory_request = "256Mi"
  grafana_memory_limit = "512Mi"
  grafana_storage_size = "10Gi"
  
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
  tls_secret_name = "cloudbridge-tls"
  
  # Tags
  tags = local.common_tags
}

# Additional production resources
resource "kubernetes_horizontal_pod_autoscaler" "cloudbridge_relay" {
  metadata {
    name      = "cloudbridge-relay-hpa"
    namespace = module.cloudbridge.namespace
    labels = {
      app.kubernetes.io/name = "cloudbridge-relay"
      app.kubernetes.io/component = "autoscaling"
    }
  }
  
  spec {
    scale_target_ref {
      api_version = "apps/v1"
      kind        = "Deployment"
      name        = "cloudbridge-relay"
    }
    
    min_replicas = 3
    max_replicas = 10
    
    metric {
      type = "Resource"
      resource {
        name = "cpu"
        target {
          type                = "Utilization"
          average_utilization = 70
        }
      }
    }
    
    metric {
      type = "Resource"
      resource {
        name = "memory"
        target {
          type                = "Utilization"
          average_utilization = 80
        }
      }
    }
  }
}

resource "kubernetes_horizontal_pod_autoscaler" "cassandra" {
  metadata {
    name      = "cassandra-hpa"
    namespace = module.cloudbridge.namespace
    labels = {
      app.kubernetes.io/name = "cassandra"
      app.kubernetes.io/component = "autoscaling"
    }
  }
  
  spec {
    scale_target_ref {
      api_version = "apps/v1"
      kind        = "StatefulSet"
      name        = "cassandra"
    }
    
    min_replicas = 3
    max_replicas = 5
    
    metric {
      type = "Resource"
      resource {
        name = "cpu"
        target {
          type                = "Utilization"
          average_utilization = 70
        }
      }
    }
    
    metric {
      type = "Resource"
      resource {
        name = "memory"
        target {
          type                = "Utilization"
          average_utilization = 80
        }
      }
    }
  }
}

# Pod Disruption Budget
resource "kubernetes_pod_disruption_budget" "cloudbridge_relay" {
  metadata {
    name      = "cloudbridge-relay-pdb"
    namespace = module.cloudbridge.namespace
    labels = {
      app.kubernetes.io/name = "cloudbridge-relay"
      app.kubernetes.io/component = "availability"
    }
  }
  
  spec {
    min_available = "2"
    
    selector {
      match_labels = {
        "app.kubernetes.io/name" = "cloudbridge-relay"
        "app.kubernetes.io/component" = "relay"
      }
    }
  }
}

resource "kubernetes_pod_disruption_budget" "cassandra" {
  metadata {
    name      = "cassandra-pdb"
    namespace = module.cloudbridge.namespace
    labels = {
      app.kubernetes.io/name = "cassandra"
      app.kubernetes.io/component = "availability"
    }
  }
  
  spec {
    min_available = "2"
    
    selector {
      match_labels = {
        "app.kubernetes.io/name" = "cassandra"
        "app.kubernetes.io/component" = "database"
      }
    }
  }
}

# Network Policies
resource "kubernetes_network_policy" "cloudbridge_relay" {
  metadata {
    name      = "cloudbridge-relay-network-policy"
    namespace = module.cloudbridge.namespace
    labels = {
      app.kubernetes.io/name = "cloudbridge-relay"
      app.kubernetes.io/component = "security"
    }
  }
  
  spec {
    pod_selector {
      match_labels = {
        "app.kubernetes.io/name" = "cloudbridge-relay"
        "app.kubernetes.io/component" = "relay"
      }
    }
    
    policy_types = ["Ingress", "Egress"]
    
    ingress {
      from {
        namespace_selector {
          match_labels = {
            name = module.cloudbridge.namespace
          }
        }
      }
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
    
    ingress {
      from {
        namespace_selector {
          match_labels = {
            name = module.cloudbridge.monitoring_namespace
          }
        }
      }
      ports {
        protocol = "TCP"
        port     = "8084"
      }
    }
    
    egress {
      to {
        namespace_selector {
          match_labels = {
            name = module.cloudbridge.namespace
          }
        }
        pod_selector {
          match_labels = {
            "app.kubernetes.io/name" = "cassandra"
          }
        }
      }
      ports {
        protocol = "TCP"
        port     = "9042"
      }
    }
    
    egress {
      to {
        namespace_selector {
          match_labels = {
            name = module.cloudbridge.namespace
          }
        }
        pod_selector {
          match_labels = {
            "app.kubernetes.io/name" = "cadence"
          }
        }
      }
      ports {
        protocol = "TCP"
        port     = "7933"
      }
    }
    
    egress {
      to = []
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



