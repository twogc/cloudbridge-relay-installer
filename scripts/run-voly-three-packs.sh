#!/usr/bin/env bash
# Run the three remaining audit packs sequentially (MiMo → Cursor each).
# No VOLY UI. Logs under docs/agent-reports/voly-runs/
set -euo pipefail

ROOT="${RELAY_INSTALLER_PATH:-/home/user1/git/cloudbridge-relay-installer}"
VOLY="${ROOT}/voly"
LOG_DIR="${ROOT}/docs/agent-reports/voly-runs"
mkdir -p "$LOG_DIR"

export RELAY_INSTALLER_PATH="$ROOT"
export VOLY_PROJECT_CWD="$ROOT"

cd "$VOLY"
set -a
# shellcheck disable=SC1091
[[ -f .env ]] && source .env
set +a

MISSIONS=(
  cloudbridge-relay-cadence-ai
  cloudbridge-relay-k8s-consol
  cloudbridge-relay-score-ports
)

.venv/bin/python -m projects.twogc twogc combat list | grep -E 'cloudbridge-relay-(cadence-ai|k8s-consol|score-ports)' || true

for m in "${MISSIONS[@]}"; do
  echo ""
  echo "========== START $m $(date -Is) =========="
  .venv/bin/python -m projects.twogc twogc combat run "$m" --sequential 2>&1 | tee "${LOG_DIR}/${m}.log"
  echo "========== DONE $m $(date -Is) =========="
done

echo "All three packs finished."
