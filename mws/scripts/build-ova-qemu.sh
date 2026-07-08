#!/usr/bin/env bash
# Build VMware OVA on Linux via QEMU/KVM (no vCenter).
# Uses Ubuntu 22.04 cloud image + cloud-init bake profile.
#
# Host port safety: QEMU uses user networking only — ports 8080/8084/9090/51820
# are NOT published on the build host (no conflict with host containers).
#
# Usage:
#   ./mws/scripts/build-ova-qemu.sh finish   # package existing disk.qcow2 / vmdk
#   SKIP_BAKE=1 ./mws/scripts/build-ova-qemu.sh
#
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
MWS="${ROOT}/mws"
TEMPLATES="${MWS}/templates"
CLOUD_INIT="${MWS}/cloud-init"

BUILD_DIR="${BUILD_DIR:-${ROOT}/dist/ova-qemu-build}"
DISK_GB="${DISK_GB:-20}"
VM_CPUS="${VM_CPUS:-4}"
VM_RAM_MB="${VM_RAM_MB:-4096}"
VM_NAME="${VM_NAME:-cloudbridge-relay}"
OVA_VERSION="${OVA_VERSION:-dev}"
UBUNTU_IMG_URL="${UBUNTU_IMG_URL:-https://cloud-images.ubuntu.com/releases/22.04/release/ubuntu-22.04-server-cloudimg-amd64.img}"
MIN_FREE_GB="${MIN_FREE_GB:-6}"
QEMU_TIMEOUT_SEC="${QEMU_TIMEOUT_SEC:-3600}"

WORKDIR="${BUILD_DIR}/work"
STAGING="${BUILD_DIR}/staging"
OUTPUT_DIR="${ROOT}/dist"
OVA_OUT="${OUTPUT_DIR}/${VM_NAME}-${OVA_VERSION}.ova"

log() { echo "[build-ova-qemu] $(date -Iseconds) $*"; }
die() { echo "[build-ova-qemu] ERROR: $*" >&2; exit 1; }

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || die "missing command: $1 (install on Ubuntu: sudo apt install qemu-system-x86 qemu-utils cloud-image-utils)"
}

check_host_ports() {
  log "Checking host ports (informational — QEMU bake does not bind these on host)..."
  local ports="8080 8084 9090 51820"
  local busy=""
  for p in $ports; do
    if ss -tuln 2>/dev/null | grep -qE ":${p}\b"; then
      busy="${busy} ${p}"
    fi
  done
  if [[ -n "${busy}" ]]; then
    log "Host already uses:${busy} — OK for bake (ports stay inside guest VM only)"
  else
    log "Host ports 8080/8084/9090/51820 are free"
  fi
}

check_disk_space() {
  local avail_kb
  avail_kb="$(df -Pk "${BUILD_DIR%/*}" 2>/dev/null | awk 'NR==2{print $4}')"
  [[ -n "${avail_kb}" ]] || avail_kb="$(df -Pk / | awk 'NR==2{print $4}')"
  local avail_gb=$((avail_kb / 1024 / 1024))
  log "Free disk: ~${avail_gb} GB (need >= ${MIN_FREE_GB} GB)"
  [[ "${avail_gb}" -ge "${MIN_FREE_GB}" ]] || die "not enough disk space (${avail_gb}G < ${MIN_FREE_GB}G). Free space or set BUILD_DIR to a larger volume"
}

generate_user_data() {
  local out="${STAGING}/user-data"
  python3 - "${out}" "${CLOUD_INIT}/cloudbridge-relay-minimal.yaml" \
    "${TEMPLATES}/docker-compose.minimal.yml" "${TEMPLATES}/env.example" <<'PY'
import sys
from pathlib import Path

out, base_yaml, compose, env_example = map(Path, sys.argv[1:5])
text = base_yaml.read_text()

# Bake profile: no docker pull, no stack start, shutdown after bake
text = text.replace(
    "  # Pre-pull images (optional; speeds first boot on slow networks)\n"
    "  - docker pull postgres:15-alpine\n"
    "  - docker pull ghcr.io/zitadel/zitadel:v2.64.1\n"
    "  - docker pull twogc/cloudbridge-relay:latest || true\n\n",
    "  # OVA bake: images pulled on first customer boot (keeps OVA smaller)\n\n",
)

compose_block = compose.read_text()
env_block = env_example.read_text()

inject = f"""
  - path: /opt/cloudbridge/docker-compose.yml
    permissions: '0644'
    owner: ubuntu:ubuntu
    content: |
{chr(10).join('      ' + line for line in compose_block.splitlines())}

  - path: /opt/cloudbridge/env.example
    permissions: '0644'
    owner: ubuntu:ubuntu
    content: |
{chr(10).join('      ' + line for line in env_block.splitlines())}
"""

marker = "write_files:\n"
if marker not in text:
    raise SystemExit("write_files section not found in cloud-init yaml")
text = text.replace(marker, marker + inject, 1)

bake_cmds = """
  - cloud-init clean --logs || true
  - rm -f /etc/ssh/ssh_host_*
  - truncate -s 0 /var/log/wtmp /var/log/btmp 2>/dev/null || true
  - sync
  - poweroff -f
"""
text = text.replace(
    "  - systemctl enable cloudbridge-stack.service\n",
    "  - systemctl enable cloudbridge-stack.service\n" + bake_cmds,
)

out.write_text(text)
PY
  log "Generated ${out}"
}

prepare_seed() {
  mkdir -p "${STAGING}" "${WORKDIR}" "${OUTPUT_DIR}"
  generate_user_data
  cat > "${STAGING}/meta-data" <<EOF
instance-id: ${VM_NAME}-bake-${OVA_VERSION}
local-hostname: ${VM_NAME}
EOF
  cloud-localds "${STAGING}/seed.iso" "${STAGING}/user-data" "${STAGING}/meta-data"
}

download_base_image() {
  local base="${BUILD_DIR}/ubuntu-22.04-server-cloudimg-amd64.img"
  local disk="${WORKDIR}/disk.qcow2"
  if [[ ! -f "${base}" ]]; then
    log "Downloading Ubuntu 22.04 cloud image..."
    curl -fL --progress-bar -o "${base}.partial" "${UBUNTU_IMG_URL}"
    mv "${base}.partial" "${base}"
  fi
  log "Preparing working disk (${DISK_GB}G)..."
  cp -f "${base}" "${disk}"
  qemu-img resize "${disk}" "${DISK_GB}G"
}

run_qemu_bake() {
  local disk="${WORKDIR}/disk.qcow2"
  local seed="${STAGING}/seed.iso"
  local logfile="${BUILD_DIR}/qemu.log"
  local pidfile="${BUILD_DIR}/qemu.pid"

  local accel="kvm"
  local cpu="host"
  if [[ ! -r /dev/kvm || ! -w /dev/kvm ]]; then
    if groups | grep -qw kvm || id -nG | grep -qw kvm; then
      die "user is in kvm group but /dev/kvm not writable — run: newgrp kvm"
    fi
    log "KVM not accessible for $(whoami) — falling back to TCG (slower)"
    accel="tcg"
    cpu="max"
  fi

  log "Starting QEMU bake (${accel}, user netdev — no host port forwards)..."
  log "Timeout: ${QEMU_TIMEOUT_SEC}s — log: ${logfile}"

  qemu-system-x86_64 \
    -machine "type=pc,accel=${accel}" \
    -cpu "${cpu}" \
    -smp "${VM_CPUS}" \
    -m "${VM_RAM_MB}" \
    -drive "file=${disk},format=qcow2,if=virtio" \
    -drive "file=${seed},format=raw,if=virtio,readonly=on" \
    -netdev user,id=net0 \
    -device virtio-net-pci,netdev=net0 \
    -serial file:"${logfile}" \
    -display none \
    -daemonize \
    -pidfile "${pidfile}"

  local elapsed=0
  while [[ -f "${pidfile}" ]] && kill -0 "$(cat "${pidfile}")" 2>/dev/null; do
    sleep 15
    elapsed=$((elapsed + 15))
    if (( elapsed % 120 == 0 )); then
      log "Still baking... ${elapsed}s (tail: $(tail -1 "${logfile}" 2>/dev/null || echo '—'))"
    fi
    if (( elapsed >= QEMU_TIMEOUT_SEC )); then
      kill "$(cat "${pidfile}")" 2>/dev/null || true
      die "QEMU bake timed out after ${QEMU_TIMEOUT_SEC}s — see ${logfile}"
    fi
  done
  log "QEMU bake finished (${elapsed}s)"
}

convert_to_vmdk() {
  local disk="${WORKDIR}/disk.qcow2"
  local vmdk="${WORKDIR}/${VM_NAME}.vmdk"
  log "Converting to VMware streamOptimized VMDK..." >&2
  qemu-img convert -p -f qcow2 -O vmdk \
    -o adapter_type=lsilogic,subformat=streamOptimized,compat6 \
    "${disk}" "${vmdk}" >&2
  echo "${vmdk}"
}

generate_ovf() {
  local vmdk="$1"
  local ovf="${WORKDIR}/${VM_NAME}.ovf"
  local vmdk_name="${VM_NAME}.vmdk"
  local info
  info="$(qemu-img info --output=json "${vmdk}")"
  local capacity filesize
  capacity="$(python3 -c "import json,sys; d=json.load(sys.stdin); print(d['virtual-size'])" <<<"${info}")"
  filesize="$(stat -c%s "${vmdk}")"

  cat > "${ovf}" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<Envelope ovf:version="1.0" xml:lang="en-US"
  xmlns="http://schemas.dmtf.org/ovf/envelope/1"
  xmlns:ovf="http://schemas.dmtf.org/ovf/envelope/1"
  xmlns:rasd="http://schemas.dmtf.org/wbem/wscim/1/cim-schema/2/CIM_ResourceAllocationSettingData"
  xmlns:vssd="http://schemas.dmtf.org/wbem/wscim/1/cim-schema/2/CIM_VirtualSystemSettingData"
  xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <References>
    <File ovf:href="${vmdk_name}" ovf:id="file1" ovf:size="${filesize}"/>
  </References>
  <DiskSection>
    <Info>Virtual disk</Info>
    <Disk ovf:capacity="${capacity}" ovf:capacityAllocationUnits="byte"
          ovf:diskId="vmdisk1" ovf:fileRef="file1"
          ovf:format="http://www.vmware.com/interfaces/specifications/vmdk.html#streamOptimized"/>
  </DiskSection>
  <NetworkSection>
    <Info>Networks</Info>
    <Network ovf:name="VM Network"/>
  </NetworkSection>
  <VirtualSystem ovf:id="${VM_NAME}">
    <Info>CloudBridge Relay Private PoP</Info>
    <Name>${VM_NAME}</Name>
    <OperatingSystemSection ovf:id="94" ovf:version="10">
      <Info>Ubuntu Linux 64-bit</Info>
      <Description>Ubuntu 22.04 LTS</Description>
    </OperatingSystemSection>
    <VirtualHardwareSection>
      <Info>Virtual hardware</Info>
      <System>
        <vssd:ElementName>Virtual Hardware Family</vssd:ElementName>
        <vssd:InstanceID>0</vssd:InstanceID>
        <vssd:VirtualSystemIdentifier>${VM_NAME}</vssd:VirtualSystemIdentifier>
      </System>
      <Item>
        <rasd:AllocationUnits>hertz * 10^6</rasd:AllocationUnits>
        <rasd:Description>Number of Virtual CPUs</rasd:Description>
        <rasd:ElementName>${VM_CPUS} virtual CPU(s)</rasd:ElementName>
        <rasd:InstanceID>1</rasd:InstanceID>
        <rasd:ResourceType>3</rasd:ResourceType>
        <rasd:VirtualQuantity>${VM_CPUS}</rasd:VirtualQuantity>
      </Item>
      <Item>
        <rasd:AllocationUnits>byte * 2^20</rasd:AllocationUnits>
        <rasd:Description>Memory Size</rasd:Description>
        <rasd:ElementName>${VM_RAM_MB} MB of memory</rasd:ElementName>
        <rasd:InstanceID>2</rasd:InstanceID>
        <rasd:ResourceType>4</rasd:ResourceType>
        <rasd:VirtualQuantity>${VM_RAM_MB}</rasd:VirtualQuantity>
      </Item>
      <Item>
        <rasd:Address>0</rasd:Address>
        <rasd:Description>SCSI Controller</rasd:Description>
        <rasd:ElementName>scsi0</rasd:ElementName>
        <rasd:InstanceID>3</rasd:InstanceID>
        <rasd:ResourceSubType>lsilogic</rasd:ResourceSubType>
        <rasd:ResourceType>6</rasd:ResourceType>
      </Item>
      <Item>
        <rasd:AddressOnParent>0</rasd:AddressOnParent>
        <rasd:ElementName>disk0</rasd:ElementName>
        <rasd:HostResource>ovf:/disk/vmdisk1</rasd:HostResource>
        <rasd:InstanceID>4</rasd:InstanceID>
        <rasd:Parent>3</rasd:Parent>
        <rasd:ResourceType>17</rasd:ResourceType>
      </Item>
      <Item>
        <rasd:AddressOnParent>0</rasd:AddressOnParent>
        <rasd:AutomaticAllocation>true</rasd:AutomaticAllocation>
        <rasd:Connection>VM Network</rasd:Connection>
        <rasd:Description>vmxnet3 Network adapter</rasd:Description>
        <rasd:ElementName>eth0</rasd:ElementName>
        <rasd:InstanceID>5</rasd:InstanceID>
        <rasd:ResourceSubType>vmxnet3</rasd:ResourceSubType>
        <rasd:ResourceType>10</rasd:ResourceType>
      </Item>
    </VirtualHardwareSection>
  </VirtualSystem>
</Envelope>
EOF
  echo "${ovf}"
}

package_ova() {
  local vmdk="$1"
  local ovf="$2"
  local mf="${WORKDIR}/${VM_NAME}.mf"
  local vmdk_name="${VM_NAME}.vmdk"
  local ovf_name="${VM_NAME}.ovf"
  local mf_name="${VM_NAME}.mf"
  local vmdk_dest="${WORKDIR}/${vmdk_name}"
  local ovf_dest="${WORKDIR}/${ovf_name}"

  [[ "$(readlink -f "${vmdk}")" == "$(readlink -f "${vmdk_dest}")" ]] || cp -f "${vmdk}" "${vmdk_dest}"
  [[ "$(readlink -f "${ovf}")" == "$(readlink -f "${ovf_dest}")" ]] || cp -f "${ovf}" "${ovf_dest}"

  (
    cd "${WORKDIR}"
    sha256sum "${ovf_name}" "${vmdk_name}" | awk '{print "SHA256("$2")= "$1}' > "${mf_name}"
    rm -f "${OVA_OUT}"
    tar -cv --format=ustar -f "${OVA_OUT}" "${ovf_name}" "${mf_name}" "${vmdk_name}"
  )

  log "OVA ready: ${OVA_OUT} ($(du -h "${OVA_OUT}" | cut -f1))"
}

main() {
  require_cmd qemu-system-x86_64
  require_cmd qemu-img
  require_cmd cloud-localds
  require_cmd python3
  require_cmd curl
  require_cmd tar

  [[ -f "${TEMPLATES}/docker-compose.minimal.yml" ]] || die "missing templates"
  [[ -f "${CLOUD_INIT}/cloudbridge-relay-minimal.yaml" ]] || die "missing cloud-init"

  local mode="${1:-build}"
  mkdir -p "${BUILD_DIR}" "${OUTPUT_DIR}"

  if [[ "${mode}" == "finish" ]]; then
    check_host_ports
    local vmdk="${WORKDIR}/${VM_NAME}.vmdk"
    if [[ ! -f "${vmdk}" ]]; then
      [[ -f "${WORKDIR}/disk.qcow2" ]] || die "missing ${WORKDIR}/disk.qcow2"
      vmdk="$(convert_to_vmdk)"
    fi
    local ovf
    ovf="$(generate_ovf "${vmdk}")"
    package_ova "${vmdk}" "${ovf}"
    log "Done. Upload:"
    log "  cd mts-2gc && ./scripts/upload-ova.sh ${OVA_OUT}"
    return
  fi

  check_host_ports
  check_disk_space

  if [[ "${SKIP_BAKE:-0}" == "1" && -f "${WORKDIR}/disk.qcow2" ]]; then
    log "SKIP_BAKE=1 — reusing ${WORKDIR}/disk.qcow2"
  else
    prepare_seed
    download_base_image
    run_qemu_bake
  fi

  local vmdk ovf
  vmdk="$(convert_to_vmdk)"
  ovf="$(generate_ovf "${vmdk}")"
  package_ova "${vmdk}" "${ovf}"

  log "Done. Upload:"
  log "  cd mts-2gc && ./scripts/upload-ova.sh ${OVA_OUT}"
}

main "${1:-build}"
