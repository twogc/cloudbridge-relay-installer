# CloudBridge Relay - Variables
variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
  default     = "dev"
}

variable "namespace" {
  description = "Kubernetes namespace for CloudBridge services"
  type        = string
  default     = "cloudbridge"
}

variable "monitoring_namespace" {
  description = "Kubernetes namespace for monitoring services"
  type        = string
  default     = "cloudbridge-monitoring"
}

variable "app_version" {
  description = "Application version"
  type        = string
  default     = "1.0.0"
}

variable "terraform_version" {
  description = "Terraform version"
  type        = string
  default     = "1.0.0"
}

variable "image_registry" {
  description = "Docker image registry"
  type        = string
  default     = "docker.io"
}

variable "log_level" {
  description = "Log level for applications"
  type        = string
  default     = "info"
  validation {
    condition     = contains(["debug", "info", "warn", "error"], var.log_level)
    error_message = "Log level must be one of: debug, info, warn, error."
  }
}

# Resource configuration
variable "relay_replicas" {
  description = "Number of CloudBridge Relay replicas"
  type        = number
  default     = 3
}

variable "relay_cpu_request" {
  description = "CPU request for CloudBridge Relay"
  type        = string
  default     = "250m"
}

variable "relay_cpu_limit" {
  description = "CPU limit for CloudBridge Relay"
  type        = string
  default     = "500m"
}

variable "relay_memory_request" {
  description = "Memory request for CloudBridge Relay"
  type        = string
  default     = "512Mi"
}

variable "relay_memory_limit" {
  description = "Memory limit for CloudBridge Relay"
  type        = string
  default     = "1Gi"
}

variable "cassandra_replicas" {
  description = "Number of Cassandra replicas"
  type        = number
  default     = 1
}

variable "cassandra_cpu_request" {
  description = "CPU request for Cassandra"
  type        = string
  default     = "250m"
}

variable "cassandra_cpu_limit" {
  description = "CPU limit for Cassandra"
  type        = string
  default     = "500m"
}

variable "cassandra_memory_request" {
  description = "Memory request for Cassandra"
  type        = string
  default     = "512Mi"
}

variable "cassandra_memory_limit" {
  description = "Memory limit for Cassandra"
  type        = string
  default     = "1Gi"
}

variable "cassandra_storage_size" {
  description = "Storage size for Cassandra"
  type        = string
  default     = "20Gi"
}

variable "cadence_replicas" {
  description = "Number of Cadence replicas"
  type        = number
  default     = 1
}

variable "cadence_cpu_request" {
  description = "CPU request for Cadence"
  type        = string
  default     = "250m"
}

variable "cadence_cpu_limit" {
  description = "CPU limit for Cadence"
  type        = string
  default     = "500m"
}

variable "cadence_memory_request" {
  description = "Memory request for Cadence"
  type        = string
  default     = "512Mi"
}

variable "cadence_memory_limit" {
  description = "Memory limit for Cadence"
  type        = string
  default     = "1Gi"
}

# Monitoring configuration
variable "prometheus_replicas" {
  description = "Number of Prometheus replicas"
  type        = number
  default     = 1
}

variable "prometheus_cpu_request" {
  description = "CPU request for Prometheus"
  type        = string
  default     = "100m"
}

variable "prometheus_cpu_limit" {
  description = "CPU limit for Prometheus"
  type        = string
  default     = "250m"
}

variable "prometheus_memory_request" {
  description = "Memory request for Prometheus"
  type        = string
  default     = "256Mi"
}

variable "prometheus_memory_limit" {
  description = "Memory limit for Prometheus"
  type        = string
  default     = "512Mi"
}

variable "prometheus_storage_size" {
  description = "Storage size for Prometheus"
  type        = string
  default     = "10Gi"
}

variable "grafana_replicas" {
  description = "Number of Grafana replicas"
  type        = number
  default     = 1
}

variable "grafana_cpu_request" {
  description = "CPU request for Grafana"
  type        = string
  default     = "50m"
}

variable "grafana_cpu_limit" {
  description = "CPU limit for Grafana"
  type        = string
  default     = "100m"
}

variable "grafana_memory_request" {
  description = "Memory request for Grafana"
  type        = string
  default     = "128Mi"
}

variable "grafana_memory_limit" {
  description = "Memory limit for Grafana"
  type        = string
  default     = "256Mi"
}

variable "grafana_storage_size" {
  description = "Storage size for Grafana"
  type        = string
  default     = "5Gi"
}

# Security configuration
variable "cassandra_password" {
  description = "Cassandra password"
  type        = string
  sensitive   = true
  default     = "cassandra-password"
}

variable "grafana_password" {
  description = "Grafana admin password"
  type        = string
  sensitive   = true
  default     = "admin123"
}

variable "api_key" {
  description = "API key for CloudBridge"
  type        = string
  sensitive   = true
  default     = "api-key-here"
}

variable "jwt_secret" {
  description = "JWT secret for authentication"
  type        = string
  sensitive   = true
  default     = "jwt-secret-here"
}

# Feature flags
variable "create_namespaces" {
  description = "Whether to create namespaces"
  type        = bool
  default     = true
}

variable "create_persistent_volumes" {
  description = "Whether to create persistent volumes"
  type        = bool
  default     = true
}

variable "enable_monitoring" {
  description = "Whether to enable monitoring stack"
  type        = bool
  default     = true
}

variable "enable_security_policies" {
  description = "Whether to enable security policies"
  type        = bool
  default     = true
}

variable "enable_ingress" {
  description = "Whether to enable ingress"
  type        = bool
  default     = false
}

# Ingress configuration
variable "ingress_host" {
  description = "Ingress host for CloudBridge"
  type        = string
  default     = "cloudbridge.example.com"
}

variable "ingress_class" {
  description = "Ingress class"
  type        = string
  default     = "nginx"
}

variable "tls_secret_name" {
  description = "TLS secret name for ingress"
  type        = string
  default     = "cloudbridge-tls"
}

# Tags
variable "tags" {
  description = "Additional tags to apply to resources"
  type        = map(string)
  default     = {}
}



