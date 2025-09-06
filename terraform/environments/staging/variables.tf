# CloudBridge Relay - Staging Environment Variables

variable "kubeconfig_path" {
  description = "Path to kubeconfig file"
  type        = string
  default     = "~/.kube/config"
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
  default     = "ghcr.io/twogc/cloudbridge-relay-installer"
}

variable "cassandra_password" {
  description = "Cassandra password"
  type        = string
  sensitive   = true
  default     = "staging-cassandra-password"
}

variable "grafana_password" {
  description = "Grafana password"
  type        = string
  sensitive   = true
  default     = "staging-grafana-password"
}

variable "api_key" {
  description = "API key"
  type        = string
  sensitive   = true
  default     = "staging-api-key"
}

variable "jwt_secret" {
  description = "JWT secret"
  type        = string
  sensitive   = true
  default     = "staging-jwt-secret"
}

variable "ingress_host" {
  description = "Ingress host"
  type        = string
  default     = "staging.cloudbridge.local"
}

