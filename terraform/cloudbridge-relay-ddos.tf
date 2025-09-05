# CloudBridge Relay - Enhanced DDoS Protection Terraform Configuration
# Комбинированная защита: L4 (nftables + XDP) + L7 (Nginx + CrowdSec)

terraform {
  required_providers {
    vkcs = {
      source  = "vk-cs/vkcs"
      version = "~> 0.9.0"
    }
  }
  required_version = ">= 1.0"
}

# Переменные
variable "project_name" {
  type        = string
  default     = "cloudbridge-relay-ddos"
  description = "Название проекта с DDoS защитой"
}

variable "compute_flavor" {
  type        = string
  default     = "STD2-8-16"  # 8 vCPU, 16 GB RAM для DDoS защиты
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
  default     = "advanced"  # basic, advanced, enterprise
  description = "Уровень DDoS защиты"
}

variable "relay_ports" {
  type = list(number)
  default = [
    80,    # HTTP (Nginx)
    443,   # HTTPS (Nginx)
    8080,  # Relay Server API
    8081,  # Relay Server Metrics
    3456,  # Relay Server Custom
    9090,  # QUIC Protocol (UDP)
    51820  # WireGuard P2P Mesh (UDP)
  ]
  description = "Порты для CloudBridge Relay с DDoS защитой"
}

# Данные
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

# Security Group для CloudBridge Relay с DDoS защитой
resource "vkcs_networking_secgroup" "relay_ddos_secgroup" {
  name        = "${var.project_name}-ddos-secgroup"
  description = "Security group for CloudBridge Relay with DDoS protection"
}

# Правила для SSH
resource "vkcs_networking_secgroup_rule" "ssh" {
  direction         = "ingress"
  protocol          = "tcp"
  port_range_min    = 22
  port_range_max    = 22
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = vkcs_networking_secgroup.relay_ddos_secgroup.id
}

# Правила для HTTP/HTTPS (Nginx)
resource "vkcs_networking_secgroup_rule" "http_https" {
  for_each = toset([for port in [80, 443] : tostring(port)])
  
  direction         = "ingress"
  protocol          = "tcp"
  port_range_min    = tonumber(each.value)
  port_range_max    = tonumber(each.value)
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = vkcs_networking_secgroup.relay_ddos_secgroup.id
}

# Правила для Relay портов
resource "vkcs_networking_secgroup_rule" "relay_ports" {
  for_each = toset([for port in [8080, 8081, 3456, 9090, 51820] : tostring(port)])

  direction         = "ingress"
  protocol          = tonumber(each.value) == 9090 || tonumber(each.value) == 51820 ? "udp" : "tcp"
  port_range_min    = tonumber(each.value)
  port_range_max    = tonumber(each.value)
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = vkcs_networking_secgroup.relay_ddos_secgroup.id
}

# Правила для исходящего трафика
resource "vkcs_networking_secgroup_rule" "egress" {
  direction         = "egress"
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = vkcs_networking_secgroup.relay_ddos_secgroup.id
}

# Виртуальная машина CloudBridge Relay с DDoS защитой
resource "vkcs_compute_instance" "relay_ddos" {
  name              = "${var.project_name}-instance"
  flavor_id         = data.vkcs_compute_flavor.relay.id
  key_pair          = var.key_pair_name
  security_group_ids = ["default", vkcs_networking_secgroup.relay_ddos_secgroup.id]
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

  # Метаданные для автоматической установки DDoS защиты
  metadata = {
    cloud-init = base64encode(templatefile("${path.module}/cloud-init-ddos.yaml", {
      project_name = var.project_name
      ddos_level   = var.ddos_protection_level
      ssh_public_key = var.ssh_public_key
    }))
  }
}

# Floating IP для внешнего доступа
resource "vkcs_networking_floatingip" "relay_ddos_fip" {
  pool = data.vkcs_networking_network.extnet.name
}

# Привязка Floating IP к VM
resource "vkcs_compute_floatingip_associate" "relay_ddos_fip" {
  floating_ip = vkcs_networking_floatingip.relay_ddos_fip.address
  instance_id = vkcs_compute_instance.relay_ddos.id
}

# Load Balancer ресурсы не поддерживаются в текущей версии провайдера

# Мониторинг с DDoS метриками
resource "vkcs_cloud_monitoring" "relay_ddos_monitoring" {
  image_id = data.vkcs_images_image.relay.id
}

# Load Balancer IP недоступен в текущей версии провайдера

# Выходные данные
output "relay_ddos_instance_ip" {
  value       = vkcs_networking_floatingip.relay_ddos_fip.address
  description = "Внешний IP адрес CloudBridge Relay с DDoS защитой"
}

output "relay_ddos_instance_id" {
  value       = vkcs_compute_instance.relay_ddos.id
  description = "ID виртуальной машины с DDoS защитой"
}

output "relay_ddos_network_id" {
  value       = data.vkcs_networking_subnet.wan_subnet.network_id
  description = "ID сети CloudBridge Relay с DDoS защитой"
}

output "loadbalancer_ip" {
  value       = "Not configured"
  description = "IP адрес Load Balancer (только для enterprise уровня)"
}

output "ddos_connection_info" {
  value = <<-EOT
    CloudBridge Relay с DDoS защитой успешно развернут!
    
    Уровень защиты: ${var.ddos_protection_level}
    Внешний IP: ${vkcs_networking_floatingip.relay_ddos_fip.address}
    SSH подключение: ssh ubuntu@${vkcs_networking_floatingip.relay_ddos_fip.address}
    
    DDoS защита:
    - L4: nftables + SYNPROXY + XDP/eBPF
    - L7: Nginx + CrowdSec + Rate Limiting
    - Мониторинг: FastNetMon + Grafana
    
    Доступные порты:
    - 80: HTTP (Nginx с DDoS защитой)
    - 443: HTTPS (Nginx с DDoS защитой)
    - 8080: Relay Server API
    - 8081: Relay Server Metrics
    - 3456: Relay Server Custom
    - 9090: QUIC Protocol (UDP)
    - 51820: WireGuard P2P Mesh (UDP)
    
    Следующие шаги:
    1. Подключитесь по SSH: ssh ubuntu@${vkcs_networking_floatingip.relay_ddos_fip.address}
    2. Проверьте статус DDoS защиты: sudo ddos-status
    3. Проверьте статус relay: sudo relay-status
    4. Запустите сервер: sudo relay-start
    5. Мониторинг: http://${vkcs_networking_floatingip.relay_ddos_fip.address}:8081
    
    DDoS мониторинг:
    - CrowdSec: http://${vkcs_networking_floatingip.relay_ddos_fip.address}:8080/crowdsec
    - FastNetMon: http://${vkcs_networking_floatingip.relay_ddos_fip.address}:8080/fastnetmon
    - Grafana: http://${vkcs_networking_floatingip.relay_ddos_fip.address}:3000
  EOT
}
