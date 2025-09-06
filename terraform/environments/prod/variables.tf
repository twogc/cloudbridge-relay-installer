# CloudBridge Relay - Production Variables
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
  default     = "docker.io"
}

variable "ingress_host" {
  description = "Ingress host for CloudBridge"
  type        = string
  default     = "cloudbridge.example.com"
}

variable "cassandra_password" {
  description = "Cassandra password"
  type        = string
  sensitive   = true
}

variable "grafana_password" {
  description = "Grafana admin password"
  type        = string
  sensitive   = true
}

variable "api_key" {
  description = "API key for CloudBridge"
  type        = string
  sensitive   = true
}

variable "jwt_secret" {
  description = "JWT secret for authentication"
  type        = string
  sensitive   = true
}



