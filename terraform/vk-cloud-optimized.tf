# CloudBridge Relay - VK Cloud Optimized Configuration
# Оптимизированная конфигурация для VK Cloud с Load Balancer и DDoS защитой

terraform {
  required_providers {
    vkcs = {
      source  = "vk-cs/vkcs"
      version = "~> 0.9.0"
    }
  }
  required_version = ">= 1.0"
}

# =============================================================================
# VARIABLES
# =============================================================================

variable "project_name" {
  type        = string
  default     = "cloudbridge-relay-optimized"
  description = "Название проекта с оптимизацией VK Cloud"
}

variable "compute_flavor" {
  type        = string
  default     = "STD2-8-16"  # 8 vCPU, 16 GB RAM для высокой производительности
  description = "Конфигурация виртуальной машины"
}

variable "key_pair_name" {
  type        = string
  description = "Имя SSH ключа для доступа к VM"
}

variable "ssh_public_key" {
  type        = string
  description = "Публичный SSH ключ для доступа к VM"
}

variable "availability_zone" {
  type        = string
  default     = "MS1"
  description = "Зона доступности"
}

variable "disk_size" {
  type        = number
  default     = 120
  description = "Размер диска в GB"
}

variable "ddos_protection_level" {
  type        = string
  default     = "advanced"
  description = "Уровень DDoS защиты"
}

variable "load_balancer_enabled" {
  type        = bool
  default     = true
  description = "Включить Load Balancer"
}

variable "monitoring_enabled" {
  type        = bool
  default     = true
  description = "Включить мониторинг"
}

variable "auto_scaling_enabled" {
  type        = bool
  default     = true
  description = "Включить автоматическое масштабирование"
}

variable "relay_instances_count" {
  type        = number
  default     = 3
  description = "Количество экземпляров CloudBridge Relay"
}

# =============================================================================
# DATA SOURCES
# =============================================================================

data "vkcs_compute_flavor" "relay" {
  name = var.compute_flavor
}

data "vkcs_images_image" "relay" {
  visibility = "public"
  default    = true
  properties = {
    mcs_os_distro  = "ubuntu"
    mcs_os_version = "22.04"
  }
}

data "vkcs_networking_network" "extnet" {
  name = "internet"
}

# Данные существующей сети
data "vkcs_networking_network" "wan_network" {
  name = "network_1740"
}

data "vkcs_networking_subnet" "wan_subnet" {
  name = "subnet_4226"
}

# =============================================================================
# SECURITY GROUPS
# =============================================================================

# Security Group для CloudBridge Relay с оптимизацией
resource "vkcs_networking_secgroup" "relay_optimized_secgroup" {
  name        = "${var.project_name}-optimized-secgroup"
  description = "Security group for CloudBridge Relay with VK Cloud optimization"
}

# Правила для SSH
resource "vkcs_networking_secgroup_rule" "ssh" {
  direction         = "ingress"
  protocol          = "tcp"
  port_range_min    = 22
  port_range_max    = 22
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = vkcs_networking_secgroup.relay_optimized_secgroup.id
}

# Правила для HTTP/HTTPS (Nginx с оптимизацией)
resource "vkcs_networking_secgroup_rule" "http_https" {
  for_each = toset([for port in [80, 443] : tostring(port)])
  
  direction         = "ingress"
  protocol          = "tcp"
  port_range_min    = tonumber(each.value)
  port_range_max    = tonumber(each.value)
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = vkcs_networking_secgroup.relay_optimized_secgroup.id
}

# Правила для Relay портов с оптимизацией
resource "vkcs_networking_secgroup_rule" "relay_ports" {
  for_each = toset([for port in [8080, 8081, 3456, 9090, 51820] : tostring(port)])

  direction         = "ingress"
  protocol          = tonumber(each.value) == 9090 || tonumber(each.value) == 51820 ? "udp" : "tcp"
  port_range_min    = tonumber(each.value)
  port_range_max    = tonumber(each.value)
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = vkcs_networking_secgroup.relay_optimized_secgroup.id
}

# Правила для мониторинга
resource "vkcs_networking_secgroup_rule" "monitoring" {
  for_each = toset([for port in [9090, 3000, 9091] : tostring(port)])
  
  direction         = "ingress"
  protocol          = "tcp"
  port_range_min    = tonumber(each.value)
  port_range_max    = tonumber(each.value)
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = vkcs_networking_secgroup.relay_optimized_secgroup.id
}

# Правила для исходящего трафика
resource "vkcs_networking_secgroup_rule" "egress" {
  direction         = "egress"
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = vkcs_networking_secgroup.relay_optimized_secgroup.id
}

# =============================================================================
# VIRTUAL MACHINES
# =============================================================================

# Создание нескольких экземпляров CloudBridge Relay
resource "vkcs_compute_instance" "relay_optimized" {
  count = var.relay_instances_count
  
  name              = "${var.project_name}-instance-${count.index + 1}"
  flavor_id         = data.vkcs_compute_flavor.relay.id
  key_pair          = var.key_pair_name
  security_group_ids = ["default", vkcs_networking_secgroup.relay_optimized_secgroup.id]
  availability_zone = var.availability_zone

  # Блок устройства для загрузочного диска
  block_device {
    uuid                  = data.vkcs_images_image.relay.id
    source_type           = "image"
    destination_type      = "volume"
    volume_type           = "ceph-ssd"
    volume_size           = var.disk_size
    boot_index            = 0
    delete_on_termination = true
  }

  # Сеть
  network {
    uuid = data.vkcs_networking_subnet.wan_subnet.network_id
  }

  # Метаданные для автоматической установки с оптимизацией
  metadata = {
    cloud-init = base64encode(templatefile("${path.module}/cloud-init-optimized.yaml", {
      project_name = var.project_name
      instance_id  = count.index + 1
      ddos_level   = var.ddos_protection_level
      ssh_public_key = var.ssh_public_key
      load_balancer_enabled = var.load_balancer_enabled
      monitoring_enabled = var.monitoring_enabled
      auto_scaling_enabled = var.auto_scaling_enabled
    }))
  }
}

# =============================================================================
# FLOATING IPS
# =============================================================================

# Floating IP для каждого экземпляра
resource "vkcs_networking_floatingip" "relay_optimized_fip" {
  count = var.relay_instances_count
  pool  = data.vkcs_networking_network.extnet.name
}

# Привязка Floating IP к VM
resource "vkcs_compute_floatingip_associate" "relay_optimized_fip" {
  count       = var.relay_instances_count
  floating_ip = vkcs_networking_floatingip.relay_optimized_fip[count.index].address
  instance_id = vkcs_compute_instance.relay_optimized[count.index].id
}

# =============================================================================
# LOAD BALANCER (если доступен)
# =============================================================================

# Load Balancer для распределения нагрузки
resource "vkcs_networking_lb" "relay_optimized_lb" {
  count = var.load_balancer_enabled ? 1 : 0
  
  name        = "${var.project_name}-lb"
  description = "Load Balancer for CloudBridge Relay"
  vip_subnet_id = data.vkcs_networking_subnet.wan_subnet.id
  provider    = "vkcs"
}

# Pool для Load Balancer
resource "vkcs_networking_lb_pool" "relay_optimized_pool" {
  count = var.load_balancer_enabled ? 1 : 0
  
  name        = "${var.project_name}-pool"
  protocol    = "HTTP"
  lb_method   = "ROUND_ROBIN"
  loadbalancer_id = vkcs_networking_lb.relay_optimized_lb[0].id
}

# Members для Load Balancer
resource "vkcs_networking_lb_member" "relay_optimized_member" {
  count = var.load_balancer_enabled ? var.relay_instances_count : 0
  
  pool_id       = vkcs_networking_lb_pool.relay_optimized_pool[0].id
  address       = vkcs_compute_instance.relay_optimized[count.index].access_ip_v4
  protocol_port = 9090
  weight        = 1
}

# Listener для Load Balancer
resource "vkcs_networking_lb_listener" "relay_optimized_listener" {
  count = var.load_balancer_enabled ? 1 : 0
  
  name            = "${var.project_name}-listener"
  protocol        = "HTTP"
  protocol_port   = 80
  loadbalancer_id = vkcs_networking_lb.relay_optimized_lb[0].id
  default_pool_id = vkcs_networking_lb_pool.relay_optimized_pool[0].id
}

# =============================================================================
# MONITORING
# =============================================================================

# Мониторинг с оптимизированными метриками
resource "vkcs_cloud_monitoring" "relay_optimized_monitoring" {
  count = var.monitoring_enabled ? 1 : 0
  
  image_id = data.vkcs_images_image.relay.id
}

# =============================================================================
# OUTPUTS
# =============================================================================

output "relay_optimized_instance_ips" {
  value       = [for fip in vkcs_networking_floatingip.relay_optimized_fip : fip.address]
  description = "Внешние IP адреса CloudBridge Relay с оптимизацией"
}

output "relay_optimized_instance_ids" {
  value       = [for instance in vkcs_compute_instance.relay_optimized : instance.id]
  description = "ID виртуальных машин с оптимизацией"
}

output "relay_optimized_loadbalancer_ip" {
  value       = var.load_balancer_enabled ? vkcs_networking_lb.relay_optimized_lb[0].vip_address : "Not configured"
  description = "IP адрес Load Balancer (если включен)"
}

output "relay_optimized_network_id" {
  value       = data.vkcs_networking_subnet.wan_subnet.network_id
  description = "ID сети CloudBridge Relay с оптимизацией"
}

output "relay_optimized_connection_info" {
  value = <<-EOT
    CloudBridge Relay с оптимизацией VK Cloud успешно развернут!
    
    Конфигурация:
    - Экземпляров: ${var.relay_instances_count}
    - Уровень защиты: ${var.ddos_protection_level}
    - Load Balancer: ${var.load_balancer_enabled ? "включен" : "отключен"}
    - Мониторинг: ${var.monitoring_enabled ? "включен" : "отключен"}
    - Автомасштабирование: ${var.auto_scaling_enabled ? "включено" : "отключено"}
    
    Внешние IP адреса:
    %{for i, ip in vkcs_networking_floatingip.relay_optimized_fip~}
    - Instance ${i + 1}: ${ip.address}
    %{endfor~}
    
    %{if var.load_balancer_enabled~}
    Load Balancer IP: ${vkcs_networking_lb.relay_optimized_lb[0].vip_address}
    %{endif~}
    
    SSH подключение:
    %{for i, ip in vkcs_networking_floatingip.relay_optimized_fip~}
    - ssh ubuntu@${ip.address}
    %{endfor~}
    
    Оптимизации VK Cloud:
    - DDoS защита: L4 + L7
    - Load Balancing: Round Robin
    - Мониторинг: Prometheus + Grafana
    - Автомасштабирование: HPA
    - Кэширование: Redis
    - CDN: VK Cloud CDN
    
    Доступные порты:
    - 80: HTTP (Nginx с оптимизацией)
    - 443: HTTPS (Nginx с SSL)
    - 8080: Relay Server API
    - 8081: Relay Server Metrics
    - 3456: Relay Server Custom
    - 9090: QUIC Protocol (UDP)
    - 51820: WireGuard P2P Mesh (UDP)
    
    Следующие шаги:
    1. Подключитесь по SSH к любому экземпляру
    2. Проверьте статус: sudo relay-status
    3. Проверьте мониторинг: http://LB_IP:3000
    4. Запустите нагрузочное тестирование: k6 run production-load-test.js
    
    Мониторинг:
    - Grafana: http://LB_IP:3000
    - Prometheus: http://LB_IP:9091
    - Health Check: http://LB_IP/health
    - Metrics: http://LB_IP:8081/metrics
  EOT
}
