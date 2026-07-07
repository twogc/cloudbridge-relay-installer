#!/usr/bin/env bash
# Build CloudBridge Relay OVA for MWS B2B Store (vendor gc2, stage).
#
# Usage:
#   ./mws/scripts/build-ova.sh packer     # HashiCorp Packer + vSphere
#   ./mws/scripts/build-ova.sh manual     # Prepare files for manual VMware export
#   ./mws/scripts/build-ova.sh upload     # Upload OVA to Artifactory (needs CLI + .env)
#
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
MWS="${ROOT}/mws"
TEMPLATES="${MWS}/templates"
DIST="${ROOT}/dist"
VERSION="${VERSION:-$(git -C "${ROOT}" describe --tags --always --dirty 2>/dev/null || echo dev)}"
OVA_NAME="cloudbridge-relay-${VERSION}.ova"

log() { echo "[build-ova] $*"; }
die() { echo "[build-ova] ERROR: $*" >&2; exit 1; }

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || die "missing command: $1"
}

prepare_templates() {
  log "Validating templates..."
  [[ -f "${TEMPLATES}/docker-compose.minimal.yml" ]] || die "missing docker-compose.minimal.yml"
  [[ -f "${TEMPLATES}/env.example" ]] || die "missing env.example"
  [[ -f "${MWS}/cloud-init/cloudbridge-relay-minimal.yaml" ]] || die "missing cloud-init snippet"
}

build_packer() {
  require_cmd packer
  prepare_templates

  local varfile="${MWS}/packer/variables.pkrvars.hcl"
  [[ -f "${varfile}" ]] || die "create ${varfile} from variables.pkrvars.hcl.example"

  log "Initializing Packer plugins..."
  packer init "${MWS}/packer/cloudbridge-relay.pkr.hcl"

  log "Building VM (export OVA manually from vCenter after build)..."
  packer build -var-file="${varfile}" "${MWS}/packer/cloudbridge-relay.pkr.hcl"

  log "Packer finished. Export the VM to ${DIST}/${OVA_NAME} via vCenter or ovftool."
}

build_manual() {
  prepare_templates
  mkdir -p "${DIST}/ova-staging/opt/cloudbridge/configs"

  cp "${TEMPLATES}/docker-compose.minimal.yml" "${DIST}/ova-staging/opt/cloudbridge/docker-compose.yml"
  cp "${TEMPLATES}/env.example" "${DIST}/ova-staging/opt/cloudbridge/env.example"
  cp "${MWS}/cloud-init/cloudbridge-relay-minimal.yaml" "${DIST}/ova-staging/cloud-init.yaml"

  cat > "${DIST}/ova-staging/MANUAL_STEPS.md" <<EOF
# Manual VMware OVA build

## 1. Create VM

| Parameter | Value |
|-----------|-------|
| OS | Ubuntu 22.04 LTS Server (amd64) |
| vCPU | 4 |
| RAM | 8 GB |
| Disk | 80 GB (thin) |
| Firmware | EFI |
| NIC | vmxnet3 |

## 2. First boot

- Attach cloud-init datasource: \`cloud-init/cloudbridge-relay-minimal.yaml\`
  (or paste as user-data in VMware guestinfo / seed ISO).
- Copy \`opt/cloudbridge/*\` from this staging dir to the VM.

## 3. Verify

\`\`\`bash
sudo systemctl start cloudbridge-stack
cloudbridge-status
\`\`\`

## 4. Export OVA

vCenter: VM → Export → OVA, or:

\`\`\`bash
ovftool vi://user:pass@vcenter/DC/vm/cloudbridge-relay ${DIST}/${OVA_NAME}
\`\`\`

## 5. Upload to MWS Artifactory

See mws/README.md § «Загрузка в Artifactory».
EOF

  log "Manual staging prepared: ${DIST}/ova-staging/"
  log "Follow ${DIST}/ova-staging/MANUAL_STEPS.md"
}

upload_artifactory() {
  local ova="${1:-${DIST}/${OVA_NAME}}"
  [[ -f "${ova}" ]] || die "OVA not found: ${ova}"

  local cli="${ARTIFACTORY_CLI:-artifactory_linux_amd64}"
  require_cmd "${cli}"

  local env_file="${B2BSTORE_ENV:-${ROOT}/../mts-2gc/.env}"
  [[ -f "${env_file}" ]] || die "set B2BSTORE_ENV or create mts-2gc/.env with Client_Id and Client_Secret"

  # shellcheck disable=SC1090
  source "${env_file}"

  local client_id="${Client_Id:-gc2}"
  local client_secret="${Client_Secret:?Client_Secret not set in .env}"

  log "Uploading ${ova} to https://stage.b2bstore.mws.ru/a ..."
  "${cli}" \
    -client-id "${client_id}" \
    -client-secret "${client_secret}" \
    -endpoint https://stage.b2bstore.mws.ru/a \
    -upload-ova "${ova}"

  log "Update artifacts.yaml hash in mts-2gc/bundle/cloudbridge-relay/ with the refID from CLI output."
}

usage() {
  cat <<EOF
Usage: $0 <packer|manual|upload> [ova-path]

  packer   Build VM via Packer (vSphere), then export OVA from vCenter
  manual   Generate staging files for manual VMware export
  upload   Upload OVA to MWS Artifactory (stage)

Environment:
  VERSION           OVA version tag (default: git describe)
  ARTIFACTORY_CLI   Path to artifactory_linux_amd64
  B2BSTORE_ENV      Path to .env with Client_Id / Client_Secret
EOF
}

main() {
  local cmd="${1:-}"
  case "${cmd}" in
    packer) build_packer ;;
    manual) build_manual ;;
    upload) upload_artifactory "${2:-}" ;;
    *) usage; exit 1 ;;
  esac
}

main "$@"
