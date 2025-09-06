# CloudBridge Relay - Main Terraform Module
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
    local = {
      source  = "hashicorp/local"
      version = "~> 2.0"
    }
  }
}

# Data sources
data "kubernetes_namespace" "cloudbridge" {
  metadata {
    name = var.namespace
  }
}

data "kubernetes_namespace" "monitoring" {
  metadata {
    name = var.monitoring_namespace
  }
}

# Local values
locals {
  common_labels = {
    app.kubernetes.io/name      = "cloudbridge-relay"
    app.kubernetes.io/version   = var.app_version
    app.kubernetes.io/component = "infrastructure"
    app.kubernetes.io/part-of   = "cloudbridge-relay"
    app.kubernetes.io/managed-by = "terraform"
    environment = var.environment
  }
  
  common_annotations = {
    "terraform.io/module" = "cloudbridge-relay"
    "terraform.io/version" = var.terraform_version
  }
}

# Namespaces
resource "kubernetes_namespace" "cloudbridge" {
  count = var.create_namespaces ? 1 : 0
  
  metadata {
    name = var.namespace
    labels = merge(local.common_labels, {
      app.kubernetes.io/component = "namespace"
    })
    annotations = merge(local.common_annotations, {
      description = "Namespace for CloudBridge Relay Server and related services"
    })
  }
}

resource "kubernetes_namespace" "monitoring" {
  count = var.create_namespaces ? 1 : 0
  
  metadata {
    name = var.monitoring_namespace
    labels = merge(local.common_labels, {
      app.kubernetes.io/component = "monitoring"
    })
    annotations = merge(local.common_annotations, {
      description = "Namespace for CloudBridge monitoring stack (Prometheus, Grafana)"
    })
  }
}

# ConfigMaps
resource "kubernetes_config_map" "cloudbridge_config" {
  metadata {
    name      = "cloudbridge-config"
    namespace = var.namespace
    labels    = merge(local.common_labels, {
      app.kubernetes.io/component = "config"
    })
    annotations = local.common_annotations
  }
  
  data = {
    "config.json" = jsonencode({
      server = {
        host = "0.0.0.0"
        port = 9090
        read_timeout = "30s"
        write_timeout = "30s"
        idle_timeout = "120s"
      }
      cadence = {
        host = "cadence-service"
        port = 7933
        domain = "cloudbridge"
        task_list = "cloudbridge-tasks"
      }
      cassandra = {
        hosts = ["cassandra-service"]
        port = 9042
        keyspace = "cloudbridge"
        consistency = "quorum"
        timeout = "10s"
      }
      wireguard = {
        enabled = true
        interface = "wg0"
        port = 51820
        private_key_path = "/etc/wireguard/private.key"
        public_key_path = "/etc/wireguard/public.key"
      }
      quic = {
        enabled = true
        port = 9090
        max_idle_timeout = "30s"
        keep_alive_period = "10s"
      }
      ai = {
        enabled = true
        model_path = "/app/models"
        max_concurrent_requests = 10
        timeout = "30s"
      }
      logging = {
        level = var.log_level
        format = "json"
        output = "stdout"
      }
      metrics = {
        enabled = true
        port = 8084
        path = "/metrics"
      }
      health = {
        enabled = true
        port = 9090
        path = "/health"
      }
    })
  }
}

# Secrets
resource "kubernetes_secret" "cloudbridge_secrets" {
  metadata {
    name      = "cloudbridge-secrets"
    namespace = var.namespace
    labels    = merge(local.common_labels, {
      app.kubernetes.io/component = "secrets"
    })
    annotations = local.common_annotations
  }
  
  type = "Opaque"
  
  data = {
    "cassandra-password" = base64encode(var.cassandra_password)
    "grafana-admin-password" = base64encode(var.grafana_password)
    "api-key" = base64encode(var.api_key)
    "jwt-secret" = base64encode(var.jwt_secret)
  }
  
  string_data = {
    "cassandra-username" = "cassandra"
    "grafana-admin-user" = "admin"
  }
}

# Persistent Volumes
resource "kubernetes_persistent_volume" "cassandra_pv" {
  count = var.create_persistent_volumes ? 1 : 0
  
  metadata {
    name = "cassandra-pv"
    labels = merge(local.common_labels, {
      app.kubernetes.io/component = "storage"
    })
    annotations = local.common_annotations
  }
  
  spec {
    capacity = {
      storage = var.cassandra_storage_size
    }
    access_modes = ["ReadWriteOnce"]
    persistent_volume_reclaim_policy = "Retain"
    storage_class_name = "cassandra-storage"
    
    host_path {
      path = "/var/lib/cassandra"
    }
  }
}

resource "kubernetes_persistent_volume_claim" "cassandra_pvc" {
  count = var.create_persistent_volumes ? 1 : 0
  
  metadata {
    name      = "cassandra-pvc"
    namespace = var.namespace
    labels = merge(local.common_labels, {
      app.kubernetes.io/component = "storage"
    })
    annotations = local.common_annotations
  }
  
  spec {
    access_modes = ["ReadWriteOnce"]
    resources {
      requests = {
        storage = var.cassandra_storage_size
      }
    }
    storage_class_name = "cassandra-storage"
  }
}

# Services
resource "kubernetes_service" "cloudbridge_relay" {
  metadata {
    name      = "cloudbridge-relay"
    namespace = var.namespace
    labels = merge(local.common_labels, {
      app.kubernetes.io/component = "relay"
    })
    annotations = merge(local.common_annotations, {
      "prometheus.io/scrape" = "true"
      "prometheus.io/port" = "8084"
      "prometheus.io/path" = "/metrics"
    })
  }
  
  spec {
    type = "ClusterIP"
    ports {
      name        = "http"
      port        = 9090
      target_port = 9090
      protocol    = "TCP"
    }
    ports {
      name        = "quic"
      port        = 9090
      target_port = 9090
      protocol    = "UDP"
    }
    ports {
      name        = "wireguard"
      port        = 51820
      target_port = 51820
      protocol    = "UDP"
    }
    ports {
      name        = "metrics"
      port        = 8084
      target_port = 8084
      protocol    = "TCP"
    }
    
    selector = {
      "app.kubernetes.io/name" = "cloudbridge-relay"
      "app.kubernetes.io/component" = "relay"
    }
  }
}

# Deployments
resource "kubernetes_deployment" "cloudbridge_relay" {
  metadata {
    name      = "cloudbridge-relay"
    namespace = var.namespace
    labels = merge(local.common_labels, {
      app.kubernetes.io/component = "relay"
    })
    annotations = local.common_annotations
  }
  
  spec {
    replicas = var.relay_replicas
    
    selector {
      match_labels = {
        "app.kubernetes.io/name" = "cloudbridge-relay"
        "app.kubernetes.io/component" = "relay"
      }
    }
    
    template {
      metadata {
        labels = merge(local.common_labels, {
          "app.kubernetes.io/component" = "relay"
        })
        annotations = merge(local.common_annotations, {
          "prometheus.io/scrape" = "true"
          "prometheus.io/port" = "8084"
          "prometheus.io/path" = "/metrics"
        })
      }
      
      spec {
        service_account_name = "cloudbridge-relay"
        
        security_context {
          run_as_non_root = true
          run_as_user     = 1000
          run_as_group    = 1000
          fs_group        = 1000
        }
        
        container {
          name  = "cloudbridge-relay"
          image = "${var.image_registry}/cloudbridge-relay:${var.app_version}"
          image_pull_policy = "IfNotPresent"
          
          ports {
            name           = "http"
            container_port = 9090
            protocol       = "TCP"
          }
          ports {
            name           = "quic"
            container_port = 9090
            protocol       = "UDP"
          }
          ports {
            name           = "wireguard"
            container_port = 51820
            protocol       = "UDP"
          }
          ports {
            name           = "metrics"
            container_port = 8084
            protocol       = "TCP"
          }
          
          env {
            name  = "CONFIG_PATH"
            value = "/app/configs/config.json"
          }
          env {
            name  = "LOG_LEVEL"
            value = var.log_level
          }
          env {
            name  = "GIN_MODE"
            value = "release"
          }
          env {
            name  = "TZ"
            value = "UTC"
          }
          env {
            name = "CASSANDRA_HOSTS"
            value = "cassandra-service"
          }
          env {
            name = "CASSANDRA_PORT"
            value = "9042"
          }
          env {
            name = "CASSANDRA_USERNAME"
            value_from {
              secret_key_ref {
                name = "cloudbridge-secrets"
                key  = "cassandra-username"
              }
            }
          }
          env {
            name = "CASSANDRA_PASSWORD"
            value_from {
              secret_key_ref {
                name = "cloudbridge-secrets"
                key  = "cassandra-password"
              }
            }
          }
          env {
            name = "CADENCE_HOST"
            value = "cadence-service"
          }
          env {
            name = "CADENCE_PORT"
            value = "7933"
          }
          
          resources {
            requests = {
              memory = var.relay_memory_request
              cpu    = var.relay_cpu_request
            }
            limits = {
              memory = var.relay_memory_limit
              cpu    = var.relay_cpu_limit
            }
          }
          
          volume_mount {
            name       = "config"
            mount_path = "/app/configs"
            read_only  = true
          }
          
          liveness_probe {
            http_get {
              path = "/health"
              port = 9090
            }
            initial_delay_seconds = 30
            period_seconds        = 10
            timeout_seconds       = 5
            failure_threshold     = 3
          }
          
          readiness_probe {
            http_get {
              path = "/health"
              port = 9090
            }
            initial_delay_seconds = 5
            period_seconds        = 5
            timeout_seconds       = 3
            failure_threshold     = 3
          }
          
          security_context {
            allow_privilege_escalation = false
            read_only_root_filesystem  = true
            run_as_non_root            = true
            run_as_user                = 1000
            run_as_group               = 1000
            
            capabilities {
              drop = ["ALL"]
              add  = ["NET_ADMIN", "NET_RAW"]
            }
          }
        }
        
        volume {
          name = "config"
          config_map {
            name = "cloudbridge-config"
          }
        }
        
        node_selector = {
          "kubernetes.io/os" = "linux"
        }
        
        toleration {
          key      = "node-role.kubernetes.io/master"
          operator = "Exists"
          effect   = "NoSchedule"
        }
        
        affinity {
          pod_anti_affinity {
            preferred_during_scheduling_ignored_during_execution {
              weight = 100
              pod_affinity_term {
                label_selector {
                  match_expressions {
                    key      = "app.kubernetes.io/name"
                    operator = "In"
                    values   = ["cloudbridge-relay"]
                  }
                }
                topology_key = "kubernetes.io/hostname"
              }
            }
          }
        }
      }
    }
  }
}

