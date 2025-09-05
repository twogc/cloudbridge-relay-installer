# CloudBridge Relay - Автоматизация через Docker Registry VK Cloud
# Этот файл демонстрирует интеграцию с VK Cloud Container Registry

# Переменные для Docker Registry
variable "docker_registry_url" {
  type        = string
  description = "URL VK Cloud Container Registry"
  default     = "registry.vkcs.ru"
}

variable "docker_image_name" {
  type        = string
  description = "Имя Docker образа в registry"
  default     = "cloudbridge-relay"
}

variable "docker_image_tag" {
  type        = string
  description = "Тег Docker образа"
  default     = "latest"
}

# Данные для Docker Registry (если поддерживается провайдером)
# data "vkcs_container_registry_image" "relay" {
#   name = var.docker_image_name
#   tag  = var.docker_image_tag
# }

# Виртуальная машина с автоматической установкой CloudBridge Relay
resource "vkcs_compute_instance" "relay_automated" {
  name              = "${var.project_name}-automated"
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

  # Минимальные метаданные
  metadata = {
    cloud-init = base64encode(templatefile("${path.module}/cloud-init-ddos.yaml", {
      project_name = var.project_name
      ssh_public_key = var.ssh_public_key
    }))
  }

  # Remote-exec provisioner для полной автоматизации
  provisioner "remote-exec" {
    inline = [
      "echo 'Waiting for cloud-init to complete...'",
      "cloud-init status --wait",
      "echo 'Cloud-init completed successfully!'"
    ]

    connection {
      type        = "ssh"
      user        = "ubuntu"
      private_key = file("~/.ssh/cloudbridge-relay")
      host        = vkcs_networking_floatingip.relay_ddos_fip.address
      timeout     = "5m"
    }
  }

  # Установка Docker и зависимостей
  provisioner "remote-exec" {
    inline = [
      "echo 'Installing Docker and dependencies...'",
      "sudo apt update",
      "sudo apt install -y docker.io docker-compose curl wget git htop vim net-tools ufw nginx fail2ban ca-certificates gnupg lsb-release",
      "sudo usermod -aG docker ubuntu",
      "sudo systemctl start docker && sudo systemctl enable docker",
      "echo 'Docker installation completed!'"
    ]

    connection {
      type        = "ssh"
      user        = "ubuntu"
      private_key = file("~/.ssh/cloudbridge-relay")
      host        = vkcs_networking_floatingip.relay_ddos_fip.address
      timeout     = "10m"
    }
  }

  # Настройка UFW firewall
  provisioner "remote-exec" {
    inline = [
      "echo 'Configuring UFW firewall...'",
      "sudo ufw --force reset",
      "sudo ufw default deny incoming",
      "sudo ufw default allow outgoing",
      "sudo ufw allow 22/tcp comment 'SSH'",
      "sudo ufw allow 80/tcp comment 'HTTP'",
      "sudo ufw allow 443/tcp comment 'HTTPS'",
      "sudo ufw allow 8080/tcp comment 'Relay Server API'",
      "sudo ufw allow 8081/tcp comment 'Relay Server Metrics'",
      "sudo ufw allow 3456/tcp comment 'Relay Server Custom'",
      "sudo ufw allow 9090/udp comment 'QUIC Protocol'",
      "sudo ufw allow 51820/udp comment 'WireGuard P2P Mesh'",
      "sudo ufw --force enable",
      "echo 'UFW configuration completed!'"
    ]

    connection {
      type        = "ssh"
      user        = "ubuntu"
      private_key = file("~/.ssh/cloudbridge-relay")
      host        = vkcs_networking_floatingip.relay_ddos_fip.address
      timeout     = "5m"
    }
  }

  # Создание директорий и установка CloudBridge Relay
  provisioner "remote-exec" {
    inline = [
      "echo 'Setting up CloudBridge Relay...'",
      "sudo mkdir -p /opt/cloudbridge-relay",
      "sudo mkdir -p /opt/cloudbridge-relay/configs",
      "sudo mkdir -p /opt/cloudbridge-relay/logs",
      "sudo chown -R ubuntu:ubuntu /opt/cloudbridge-relay",
      "cd /opt/cloudbridge-relay",
      "echo 'Creating docker-compose.yml...'",
      "cat > docker-compose.yml << 'EOF'",
      "version: '3.8'",
      "services:",
      "  cloudbridge-relay:",
      "    image: ${var.docker_registry_url}/${var.docker_image_name}:${var.docker_image_tag}",
      "    container_name: cloudbridge-relay",
      "    restart: unless-stopped",
      "    security_opt:",
      "      - no-new-privileges:true",
      "    cap_add:",
      "      - NET_ADMIN",
      "      - SYS_MODULE",
      "    devices:",
      "      - /dev/net/tun:/dev/net/tun",
      "    ports:",
      "      - '8080:8080'",
      "      - '8081:8081'",
      "      - '3456:3456'",
      "      - '9090:9090/udp'",
      "      - '51820:51820/udp'",
      "    volumes:",
      "      - ./configs:/app/configs:ro",
      "      - ./logs:/app/logs",
      "      - /etc/wireguard:/etc/wireguard:ro",
      "      - /var/run/wireguard:/var/run/wireguard",
      "    environment:",
      "      - CONFIG_PATH=/app/configs/config.json",
      "      - LOG_LEVEL=info",
      "      - WIREGUARD_ENABLED=true",
      "      - TZ=UTC",
      "    healthcheck:",
      "      test: ['CMD', 'curl', '-f', 'http://localhost:8080/health']",
      "      interval: 30s",
      "      timeout: 10s",
      "      retries: 3",
      "      start_period: 40s",
      "    deploy:",
      "      resources:",
      "        limits:",
      "          cpus: '2.0'",
      "          memory: 2G",
      "        reservations:",
      "          cpus: '0.5'",
      "          memory: 512M",
      "    logging:",
      "      driver: 'json-file'",
      "      options:",
      "        max-size: '10m'",
      "        max-file: '3'",
      "    networks:",
      "      - cloudbridge-network",
      "",
      "networks:",
      "  cloudbridge-network:",
      "    driver: bridge",
      "EOF",
      "echo 'Docker Compose configuration created!'"
    ]

    connection {
      type        = "ssh"
      user        = "ubuntu"
      private_key = file("~/.ssh/cloudbridge-relay")
      host        = vkcs_networking_floatingip.relay_ddos_fip.address
      timeout     = "5m"
    }
  }

  # Запуск CloudBridge Relay
  provisioner "remote-exec" {
    inline = [
      "echo 'Starting CloudBridge Relay...'",
      "cd /opt/cloudbridge-relay",
      "sudo docker-compose up -d",
      "echo 'CloudBridge Relay started successfully!'"
    ]

    connection {
      type        = "ssh"
      user        = "ubuntu"
      private_key = file("~/.ssh/cloudbridge-relay")
      host        = vkcs_networking_floatingip.relay_ddos_fip.address
      timeout     = "5m"
    }
  }

  # Создание скриптов управления
  provisioner "remote-exec" {
    inline = [
      "echo 'Creating management scripts...'",
      "cat > /tmp/relay-status.sh << 'EOF'",
      "#!/bin/bash",
      "echo '=== CloudBridge Relay Status ==='",
      "echo 'Nginx: $(systemctl is-active nginx)'",
      "echo 'UFW: $(sudo ufw status | grep Status)'",
      "echo 'Fail2ban: $(systemctl is-active fail2ban)'",
      "echo 'Docker: $(systemctl is-active docker)'",
      "echo ''",
      "echo '=== Listening Ports ==='",
      "sudo ss -tlnp | grep -E ':(80|443|8080|8081|3456|9090|51820)'",
      "echo ''",
      "echo '=== Docker Containers ==='",
      "sudo docker ps -a",
      "EOF",
      "",
      "cat > /tmp/ddos-status.sh << 'EOF'",
      "#!/bin/bash",
      "echo '=== DDoS Protection Status ==='",
      "echo 'nginx: $(systemctl is-active nginx)'",
      "echo 'fail2ban: $(systemctl is-active fail2ban)'",
      "echo 'ufw: $(sudo ufw status | grep Status)'",
      "echo 'docker: $(systemctl is-active docker)'",
      "echo ''",
      "echo '=== Active UFW Rules ==='",
      "sudo ufw status numbered",
      "echo ''",
      "echo '=== Docker Containers ==='",
      "sudo docker ps -a",
      "EOF",
      "",
      "sudo mv /tmp/relay-status.sh /usr/local/bin/relay-status",
      "sudo mv /tmp/ddos-status.sh /usr/local/bin/ddos-status",
      "sudo chmod +x /usr/local/bin/relay-status",
      "sudo chmod +x /usr/local/bin/ddos-status",
      "echo 'Management scripts created!'"
    ]

    connection {
      type        = "ssh"
      user        = "ubuntu"
      private_key = file("~/.ssh/cloudbridge-relay")
      host        = vkcs_networking_floatingip.relay_ddos_fip.address
      timeout     = "5m"
    }
  }

  # Финальная проверка статуса
  provisioner "remote-exec" {
    inline = [
      "echo '=== Final Status Check ==='",
      "sudo relay-status",
      "echo '=== Installation Complete ==='",
      "echo 'CloudBridge Relay server is ready!'",
      "echo 'Check status with: sudo relay-status'",
      "echo 'Check DDoS protection with: sudo ddos-status'"
    ]

    connection {
      type        = "ssh"
      user        = "ubuntu"
      private_key = file("~/.ssh/cloudbridge-relay")
      host        = vkcs_networking_floatingip.relay_ddos_fip.address
      timeout     = "5m"
    }
  }
}

# Выходные данные для автоматизированного сервера
output "automated_relay_info" {
  value = <<-EOT
    CloudBridge Relay с автоматизацией успешно развернут!
    
    IP адрес: ${vkcs_networking_floatingip.relay_ddos_fip.address}
    SSH подключение: ssh ubuntu@${vkcs_networking_floatingip.relay_ddos_fip.address}
    
    Docker образ: ${var.docker_registry_url}/${var.docker_image_name}:${var.docker_image_tag}
    
    Доступные сервисы:
    - HTTP: http://${vkcs_networking_floatingip.relay_ddos_fip.address}/
    - Relay API: http://${vkcs_networking_floatingip.relay_ddos_fip.address}:8080/
    - Metrics: http://${vkcs_networking_floatingip.relay_ddos_fip.address}:8081/
    
    Управление:
    - Статус: sudo relay-status
    - DDoS защита: sudo ddos-status
    - Docker: sudo docker ps -a
    
    Автоматизация: Полностью автоматизированное развертывание!
  EOT
}
