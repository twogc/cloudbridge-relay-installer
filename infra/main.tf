terraform {
  required_providers {
    vkcs = {
      source  = "vk-cs/vkcs"
      version = ">= 0.7.0"
    }
  }
}

provider "vkcs" {
  auth_url     = "https://api.vkcloud.io/identity/v3"
  username     = var.vkcs_username
  password     = var.vkcs_password
  domain_name  = var.vkcs_domain
  project_name = var.vkcs_project
  region       = var.vkcs_region
}

resource "vkcs_compute_instance" "relay" {
  name            = "relay-${var.vkcs_region}"
  flavor_name     = var.flavor_name
  key_pair        = var.ssh_key_name
  security_groups = ["default"]
  image_name      = var.image_name

  network {
    uuid = var.network_id
  }

  floating_ip {
    pool = var.floating_ip_pool
  }
}

output "relay_public_ip" {
  value = vkcs_compute_instance.relay.floating_ip[0].address
} 