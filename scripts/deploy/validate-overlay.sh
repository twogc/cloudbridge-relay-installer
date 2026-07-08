#!/bin/bash
# Offline / pre-flight validation for kustomize overlays.
# Does not require a live cluster. Optional: kubectl for kustomize render.
#
# Usage:
#   ./scripts/deploy/validate-overlay.sh
#   ./scripts/deploy/validate-overlay.sh --overlay k8s/overlays/production
#   ./scripts/deploy/validate-overlay.sh --all
#   ./scripts/deploy/validate-overlay.sh --live   # needs cluster + KUBECONFIG

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$REPO_ROOT"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_ok() { echo -e "${GREEN}[OK]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_err() { echo -e "${RED}[ERR]${NC} $1"; }

OVERLAYS=()
RUN_ALL=0
RUN_LIVE=0
FAILS=0

usage() {
  cat <<'EOF'
validate-overlay.sh — offline checks for k8s/overlays/*

  --overlay DIR   Validate one overlay (repeatable)
  --all           Validate production + zitadel overlays
  --live          After offline checks, kubectl get ns/pods (needs cluster)
  -h, --help      This help

Default: --all
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --overlay) OVERLAYS+=("$2"); shift 2 ;;
    --overlay=*) OVERLAYS+=("${1#*=}"); shift ;;
    --all) RUN_ALL=1; shift ;;
    --live) RUN_LIVE=1; shift ;;
    -h|--help) usage; exit 0 ;;
    *) log_err "Unknown option: $1"; usage; exit 1 ;;
  esac
done

if [[ ${#OVERLAYS[@]} -eq 0 ]] || [[ "$RUN_ALL" == "1" ]]; then
  OVERLAYS=(k8s/overlays/production k8s/overlays/zitadel)
fi

# Parse resources: from kustomization.yaml (simple YAML list under resources:)
extract_resources() {
  local kust="$1"
  # shellcheck disable=SC2016
  awk '
    /^resources:/ { inres=1; next }
    inres && /^[[:space:]]*-[[:space:]]*/ {
      line=$0
      sub(/^[[:space:]]*-[[:space:]]*/, "", line)
      gsub(/[[:space:]]*#.*$/, "", line)
      gsub(/"/, "", line)
      if (line != "") print line
      next
    }
    inres && /^[^[:space:]#]/ { inres=0 }
  ' "$kust"
}

validate_overlay() {
  local dir="$1"
  local kust="$dir/kustomization.yaml"
  log_info "=== Overlay: $dir ==="

  if [[ ! -f "$kust" ]]; then
    log_err "missing $kust"
    FAILS=$((FAILS + 1))
    return
  fi
  log_ok "kustomization.yaml exists"

  if ! grep -q 'apiVersion: kustomize.config.k8s.io' "$kust"; then
    log_warn "unexpected apiVersion in $kust"
  fi

  local missing=0
  local count=0
  while IFS= read -r rel; do
    count=$((count + 1))
    # Resolve relative to overlay dir
    local path
    path="$(cd "$dir" && realpath -m "$rel" 2>/dev/null || python3 -c "import os; print(os.path.normpath(os.path.join('$dir', '$rel')))")"
    if [[ ! -f "$path" ]]; then
      log_err "resource missing: $rel -> $path"
      missing=$((missing + 1))
      FAILS=$((FAILS + 1))
    else
      log_ok "resource: $rel"
    fi
  done < <(extract_resources "$kust")

  if [[ "$count" -eq 0 ]]; then
    log_err "no resources listed in $kust"
    FAILS=$((FAILS + 1))
  else
    log_ok "$count resource path(s) checked ($missing missing)"
  fi

  # Optional kustomize render
  if command -v kubectl >/dev/null 2>&1; then
    if out=$(kubectl kustomize "$dir" 2>&1); then
      local kinds
      kinds=$(printf '%s\n' "$out" | grep -c '^kind:' || true)
      log_ok "kubectl kustomize rendered (kind count≈$kinds)"
    else
      log_err "kubectl kustomize failed: $out"
      FAILS=$((FAILS + 1))
    fi
  elif command -v kustomize >/dev/null 2>&1; then
    if out=$(kustomize build "$dir" 2>&1); then
      log_ok "kustomize build OK"
    else
      log_err "kustomize build failed: $out"
      FAILS=$((FAILS + 1))
    fi
  else
    log_warn "kubectl/kustomize not installed — skip render"
  fi
}

for o in "${OVERLAYS[@]}"; do
  validate_overlay "$o"
  echo ""
done

# deploy.sh flag smoke
if [[ -x scripts/deploy/deploy.sh ]]; then
  if scripts/deploy/deploy.sh --help >/dev/null 2>&1; then
    log_ok "deploy.sh --help"
  else
    log_err "deploy.sh --help failed"
    FAILS=$((FAILS + 1))
  fi
  bash -n scripts/deploy/deploy.sh && log_ok "deploy.sh bash -n" || {
    log_err "deploy.sh syntax"
    FAILS=$((FAILS + 1))
  }
fi

if [[ "$RUN_LIVE" == "1" ]]; then
  log_info "=== Live cluster checks ==="
  if ! command -v kubectl >/dev/null 2>&1; then
    log_err "kubectl required for --live"
    FAILS=$((FAILS + 1))
  elif ! kubectl cluster-info >/dev/null 2>&1; then
    log_err "cannot reach cluster (set KUBECONFIG)"
    FAILS=$((FAILS + 1))
  else
    log_ok "cluster reachable"
    kubectl get ns cloudbridge 2>/dev/null && log_ok "namespace cloudbridge" || log_warn "namespace cloudbridge missing"
    kubectl get pods -n cloudbridge 2>/dev/null || true
    kubectl get svc -n cloudbridge 2>/dev/null || true
  fi
fi

echo ""
if [[ "$FAILS" -gt 0 ]]; then
  log_err "Validation finished with $FAILS failure(s)"
  exit 1
fi
log_ok "All offline overlay checks passed"
exit 0
