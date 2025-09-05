# CloudBridge Relay — Terraform Deployment for VK Cloud

This repository contains Terraform configurations and automation scripts to deploy the CloudBridge Relay server on VK Cloud (OpenStack-based).

## Features
- Automated VM provisioning (Ubuntu 22.04)
- Cloud-init for first-boot configuration (Docker, Docker Compose, Nginx)
- Hardened security (UFW, Fail2Ban via cloud-init), logrotate, basic monitoring
- Ready-to-use deployment script `terraform/deploy.sh`
- Uses existing WAN network/subnet in VK Cloud

## Prerequisites
- Terraform 1.0+
- OpenStack CLI: `pip3 install python-openstackclient`
- VK Cloud credentials
- SSH keypair for VM access

## Quick start
See `QUICK_START.md` for a condensed flow. High-level steps:

```bash
git clone https://github.com/twogc/relay-terraform-install.git
cd relay-terraform-install
cp env.example .env
nano .env   # fill VK Cloud auth vars
source .env
cd terraform
cp terraform.tfvars.example terraform.tfvars
nano terraform.tfvars # set your SSH public key, flavor, etc.
./deploy.sh deploy
```

## VK Cloud authentication
Export these environment variables (use `.env` file):

```bash
export OS_AUTH_URL="https://auth.cloud.vk.com/v3"
export OS_USERNAME="<your_username>"
export OS_PASSWORD="<your_password>"
export OS_PROJECT_ID="<your_project_id>"
export OS_PROJECT_NAME="<your_project_name>"
export OS_USER_DOMAIN_NAME="Default"
export OS_PROJECT_DOMAIN_NAME="Default"
export OS_REGION_NAME="RegionOne"
export OS_INTERFACE="public"
export OS_IDENTITY_API_VERSION=3
```

Load them in your shell:
```bash
source .env
openstack token issue
```

## Structure
```
.
├── LICENSE
├── QUICK_START.md
├── README.md
└── terraform/
    ├── cloud-init.yaml
    ├── cloudbridge-relay.tf
    ├── deploy.sh
    └── terraform.tfvars.example
```

## Important
- This repo deploys the Relay server only (no Keycloak).
- It expects an existing WAN network/subnet in VK Cloud (e.g., `network_1740` / `subnet_4226`).
- The latest private Glance image is selected automatically.

## Management
Use `terraform/deploy.sh`:
```bash
./deploy.sh plan
./deploy.sh deploy
./deploy.sh status
./deploy.sh info
./deploy.sh destroy
```

## Security notes
- Do not commit secrets. Use `.env` which is ignored by `.gitignore`.
- Cloud-init config sets up firewall and basic hardening on first boot.

## License
MIT
