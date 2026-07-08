#!/usr/bin/env bash
# Install Agentic Awesome Skills for cloudbridge-relay-installer (Cursor).
# Source: rep/agentic-awesome-skills (local clone) or npx fallback.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
AAS_LOCAL="${ROOT}/rep/agentic-awesome-skills"
PROJECT_SKILLS="${ROOT}/.cursor/skills"
GLOBAL_SKILLS="${HOME}/.cursor/skills"

# Plugin bundles curated for Go/K8s relay installer work
PLUGINS=(
  agentic-bundle-aas-devops-cloud
  agentic-bundle-aas-qa-test-automation
  agentic-bundle-aas-security-engineer
  agentic-bundle-aas-oss-maintainer
  agentic-bundle-aas-observability-ir
  agentic-bundle-aas-api-platform-builder
)

# Extra Go skills from main catalog
EXTRA_SKILLS=(
  golang-pro
  go-concurrency-patterns
)

link_skill() {
  local src="$1"
  local dest_dir="$2"
  local name
  name="$(basename "$src")"
  local dest="${dest_dir}/${name}"

  if [[ ! -d "$src" ]] || [[ ! -f "${src}/SKILL.md" ]]; then
    echo "  skip (missing SKILL.md): $src" >&2
    return 0
  fi

  mkdir -p "$dest_dir"
  if [[ -L "$dest" ]] || [[ -e "$dest" ]]; then
    rm -rf "$dest"
  fi
  ln -sfn "$src" "$dest"
  echo "  linked: $name -> $src"
}

install_from_local() {
  local target="$1"
  echo "Installing skills into: $target"

  for plugin in "${PLUGINS[@]}"; do
    local plugin_dir="${AAS_LOCAL}/plugins/${plugin}/skills"
    if [[ ! -d "$plugin_dir" ]]; then
      echo "  warn: plugin not found: $plugin" >&2
      continue
    fi
    echo "  plugin: $plugin"
    for skill in "$plugin_dir"/*; do
      [[ -d "$skill" ]] || continue
      link_skill "$skill" "$target"
    done
  done

  for skill_name in "${EXTRA_SKILLS[@]}"; do
    link_skill "${AAS_LOCAL}/skills/${skill_name}" "$target"
  done
}

install_global_npx() {
  echo "Local repo missing; installing filtered skills via npx..."
  npx --yes agentic-awesome-skills@13.13.0 --cursor \
    --category development,backend,infrastructure,security,testing \
    --risk safe,none
}

main() {
  if [[ ! -d "$AAS_LOCAL/skills" ]]; then
    install_global_npx
    mkdir -p "$PROJECT_SKILLS"
    echo "Project skills: symlink rep/agentic-awesome-skills first, then re-run this script."
    exit 0
  fi

  install_from_local "$GLOBAL_SKILLS"
  install_from_local "$PROJECT_SKILLS"

  local count
  count="$(find -L "$PROJECT_SKILLS" -mindepth 1 -maxdepth 1 -type d 2>/dev/null | wc -l | tr -d ' ')"
  echo ""
  echo "Done. ${count} skills in ${PROJECT_SKILLS}"
  echo "Global copy: ${GLOBAL_SKILLS}"
}

main "$@"
