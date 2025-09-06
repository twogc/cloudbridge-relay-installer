# CloudBridge Relay - Outputs
output "namespace" {
  description = "CloudBridge namespace name"
  value       = var.namespace
}

output "monitoring_namespace" {
  description = "Monitoring namespace name"
  value       = var.monitoring_namespace
}

output "cloudbridge_relay_service" {
  description = "CloudBridge Relay service details"
  value = {
    name      = kubernetes_service.cloudbridge_relay.metadata[0].name
    namespace = kubernetes_service.cloudbridge_relay.metadata[0].namespace
    cluster_ip = kubernetes_service.cloudbridge_relay.spec[0].cluster_ip
    ports = kubernetes_service.cloudbridge_relay.spec[0].port
  }
}

output "cloudbridge_relay_deployment" {
  description = "CloudBridge Relay deployment details"
  value = {
    name      = kubernetes_deployment.cloudbridge_relay.metadata[0].name
    namespace = kubernetes_deployment.cloudbridge_relay.metadata[0].namespace
    replicas  = kubernetes_deployment.cloudbridge_relay.spec[0].replicas
  }
}

output "config_map" {
  description = "CloudBridge config map details"
  value = {
    name      = kubernetes_config_map.cloudbridge_config.metadata[0].name
    namespace = kubernetes_config_map.cloudbridge_config.metadata[0].namespace
  }
}

output "secrets" {
  description = "CloudBridge secrets details"
  value = {
    name      = kubernetes_secret.cloudbridge_secrets.metadata[0].name
    namespace = kubernetes_secret.cloudbridge_secrets.metadata[0].namespace
  }
  sensitive = true
}

output "persistent_volumes" {
  description = "Persistent volumes details"
  value = var.create_persistent_volumes ? {
    cassandra_pv = {
      name = kubernetes_persistent_volume.cassandra_pv[0].metadata[0].name
      size = kubernetes_persistent_volume.cassandra_pv[0].spec[0].capacity.storage
    }
    cassandra_pvc = {
      name      = kubernetes_persistent_volume_claim.cassandra_pvc[0].metadata[0].name
      namespace = kubernetes_persistent_volume_claim.cassandra_pvc[0].metadata[0].namespace
    }
  } : null
}

output "access_urls" {
  description = "Access URLs for services"
  value = {
    relay_service = "http://${kubernetes_service.cloudbridge_relay.metadata[0].name}.${var.namespace}.svc.cluster.local:9090"
    health_check  = "http://${kubernetes_service.cloudbridge_relay.metadata[0].name}.${var.namespace}.svc.cluster.local:9090/health"
    metrics       = "http://${kubernetes_service.cloudbridge_relay.metadata[0].name}.${var.namespace}.svc.cluster.local:8084/metrics"
  }
}

output "port_forward_commands" {
  description = "Port forward commands for local access"
  value = {
    relay_service = "kubectl port-forward service/${kubernetes_service.cloudbridge_relay.metadata[0].name} 9090:9090 -n ${var.namespace}"
    metrics       = "kubectl port-forward service/${kubernetes_service.cloudbridge_relay.metadata[0].name} 8084:8084 -n ${var.namespace}"
  }
}

output "kubectl_commands" {
  description = "Useful kubectl commands"
  value = {
    get_pods = "kubectl get pods -n ${var.namespace}"
    get_services = "kubectl get services -n ${var.namespace}"
    get_logs = "kubectl logs -f deployment/cloudbridge-relay -n ${var.namespace}"
    describe_pods = "kubectl describe pods -n ${var.namespace}"
  }
}

output "environment_info" {
  description = "Environment information"
  value = {
    environment = var.environment
    app_version = var.app_version
    terraform_version = var.terraform_version
    image_registry = var.image_registry
    log_level = var.log_level
  }
}

output "resource_summary" {
  description = "Resource usage summary"
  value = {
    relay_replicas = var.relay_replicas
    relay_cpu_request = var.relay_cpu_request
    relay_cpu_limit = var.relay_cpu_limit
    relay_memory_request = var.relay_memory_request
    relay_memory_limit = var.relay_memory_limit
    cassandra_replicas = var.cassandra_replicas
    cassandra_cpu_request = var.cassandra_cpu_request
    cassandra_cpu_limit = var.cassandra_cpu_limit
    cassandra_memory_request = var.cassandra_memory_request
    cassandra_memory_limit = var.cassandra_memory_limit
    cassandra_storage_size = var.cassandra_storage_size
  }
}



