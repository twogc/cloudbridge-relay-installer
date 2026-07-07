# Packer skeleton — CloudBridge Relay Private PoP OVA for MWS B2B Store.
# Builder: VMware vSphere (export to OVA). Adjust for your lab ESXi/vCenter.
#
# Prerequisites:
#   packer init mws/packer/cloudbridge-relay.pkr.hcl
#   cp mws/packer/variables.pkrvars.hcl.example mws/packer/variables.pkrvars.hcl
#
# Build:
#   packer build -var-file=mws/packer/variables.pkrvars.hcl mws/packer/cloudbridge-relay.pkr.hcl

packer {
  required_version = ">= 1.9.0"
  required_plugins {
    vsphere = {
      version = ">= 1.2.0"
      source  = "github.com/hashicorp/vsphere"
    }
  }
}

variable "vsphere_endpoint" {
  type        = string
  description = "vCenter or ESXi hostname"
}

variable "vsphere_username" {
  type      = string
  sensitive = true
}

variable "vsphere_password" {
  type      = string
  sensitive = true
}

variable "vsphere_datacenter" {
  type    = string
  default = "Datacenter"
}

variable "vsphere_cluster" {
  type    = string
  default = "Cluster"
}

variable "vsphere_datastore" {
  type    = string
  default = "datastore1"
}

variable "vsphere_network" {
  type    = string
  default = "VM Network"
}

variable "vm_cpus" {
  type    = number
  default = 4
}

variable "vm_memory" {
  type    = number
  default = 8192
}

variable "vm_disk_size" {
  type    = number
  default = 81920
}

variable "ssh_username" {
  type    = string
  default = "ubuntu"
}

variable "ssh_password" {
  type      = string
  sensitive = true
  default   = ""
}

variable "iso_url" {
  type    = string
  default = "https://releases.ubuntu.com/jammy/ubuntu-22.04.5-live-server-amd64.iso"
}

variable "iso_checksum" {
  type    = string
  default = "sha256:71d0c28b8e8b3c7c5c8c0e8e8e8e8e8e8e8e8e8e8e8e8e8e8e8e8e8e8e8e8e8"
  # Replace with actual checksum from https://releases.ubuntu.com/22.04/
}

variable "output_ova_name" {
  type    = string
  default = "cloudbridge-relay"
}

variable "ova_version" {
  type    = string
  default = "dev"
}

locals {
  ova_version = var.ova_version
}

source "vsphere-iso" "cloudbridge_relay" {
  vcenter_server      = var.vsphere_endpoint
  username            = var.vsphere_username
  password            = var.vsphere_password
  insecure_connection = true

  datacenter = var.vsphere_datacenter
  cluster    = var.vsphere_cluster
  datastore  = var.vsphere_datastore
  folder     = "packer"

  vm_name              = "cloudbridge-relay-${local.ova_version}"
  guest_os_type        = "ubuntu64Guest"
  firmware             = "efi"
  CPUs                 = var.vm_cpus
  CPU_hot_plug         = false
  RAM                  = var.vm_memory
  RAM_hot_plug         = false
  disk_controller_type = ["pvscsi"]
  storage {
    disk_size             = var.vm_disk_size
    disk_thin_provisioned = true
  }
  network_adapters {
    network      = var.vsphere_network
    network_card = "vmxnet3"
  }

  iso_url      = var.iso_url
  iso_checksum = var.iso_checksum

  http_directory = "${path.root}/http"
  boot_command = [
    "<esc><wait>",
    "c<wait>",
    "linux /casper/vmlinuz --- autoinstall ds=nocloud-net\\;s=http://{{ .HTTPIP }}:{{ .HTTPPort }}/",
    "<enter><wait>",
    "initrd /casper/initrd<enter><wait>",
    "boot<enter>"
  ]
  boot_wait = "10s"

  ssh_username = var.ssh_username
  ssh_password = var.ssh_password
  ssh_timeout  = "45m"

  shutdown_command = "echo '${var.ssh_password}' | sudo -S shutdown -P now"
  shutdown_timeout = "5m"

  convert_to_template = false
}

build {
  sources = ["source.vsphere-iso.cloudbridge_relay"]

  provisioner "file" {
    source      = "${path.root}/../templates/docker-compose.minimal.yml"
    destination = "/tmp/docker-compose.yml"
  }

  provisioner "file" {
    source      = "${path.root}/../templates/env.example"
    destination = "/tmp/env.example"
  }

  provisioner "file" {
    source      = "${path.root}/../cloud-init/cloudbridge-relay-minimal.yaml"
    destination = "/tmp/cloudbridge-relay-minimal.yaml"
  }

  provisioner "shell" {
    inline = [
      "sudo cloud-init single --name write-files --frequency always --file /tmp/cloudbridge-relay-minimal.yaml || sudo cloud-init init && sudo cloud-init single --name scripts_user --frequency always",
      "sudo mkdir -p /opt/cloudbridge/configs",
      "sudo cp /tmp/docker-compose.yml /opt/cloudbridge/docker-compose.yml",
      "sudo cp /tmp/env.example /opt/cloudbridge/env.example",
      "sudo chown -R ubuntu:ubuntu /opt/cloudbridge",
      "sudo systemctl enable cloudbridge-stack.service || true",
      "sudo cloud-init clean --logs",
      "sudo rm -f /etc/ssh/ssh_host_*",
      "sudo truncate -s 0 /var/log/wtmp /var/log/btmp",
      "history -c"
    ]
  }

  post-processor "shell-local" {
    inline = [
      "echo 'Export VM to OVA with ovftool or vCenter: File > Export > OVA'",
      "echo 'Output name: ${var.output_ova_name}-${local.ova_version}.ova'"
    ]
  }
}
