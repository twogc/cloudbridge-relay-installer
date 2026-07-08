#!/bin/bash
# CloudBridge Relay WireGuard - Универсальный инсталлер (исправленный)
# Поддержка: Linux, macOS, (Windows через WSL2 с ручной установкой)
# Не зависит от git.zx2c4.com — используем официальное GitHub-зеркало wireguard-go.

set -euo pipefail

# === Цвета ===
GREEN='\033[0;32m'; YELLOW='\033[1;33m'; RED='\033[0;31m'; BLUE='\033[0;34m'; NC='\033[0m'

# === Переменные ===
VERSION="1.0.1"
BINARY_NAME="cloudbridge-relay"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/cloudbridge-relay"
SERVICE_DIR="/etc/systemd/system"
REPO_URL="https://github.com/twogc/cloudbridge-relay-installer"
DOCKER_IMAGE="twogc/cloudbridge-relay"
WIREGUARD_GO_MIRROR="https://github.com/WireGuard/wireguard-go.git"  # оф. mirror-only :contentReference[oaicite:1]{index=1}

# === Вывод ===
log()   { echo -e "${GREEN}[INFO]${NC} $*"; }
warn()  { echo -e "${YELLOW}[WARN]${NC} $*"; }
err()   { echo -e "${RED}[ERROR]${NC} $*"; }
step()  { echo -e "${BLUE}[STEP]${NC} $*"; }
banner(){ echo -e "${BLUE}\n== CloudBridge Relay · WireGuard Edition ==${NC}\n"; }

# === ОС ===
OS=""; VER=""
detect_os() {
  if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    if [[ -f /etc/os-release ]]; then . /etc/os-release; OS=$NAME; VER=$VERSION_ID; else OS=$(uname -s); VER=$(uname -r); fi
  elif [[ "$OSTYPE" == "darwin"* ]]; then OS="macOS"; VER=$(sw_vers -productVersion)
  elif [[ "$OSTYPE" == "msys" || "$OSTYPE" == "cygwin" ]]; then OS="Windows"; VER="$(uname -r)"
  else OS="Unknown"; VER="Unknown"; fi
  log "Обнаружена ОС: $OS $VER"
}

# === Зависимости разработчика (минимум: git; go нужен только для локальной сборки) ===
check_base_deps() {
  step "Проверка базовых зависимостей..."
  command -v git >/dev/null || { err "Git не установлен"; exit 1; }
  log "Git найден"
  if command -v go >/dev/null; then
    log "Go найден: $(go version)"
  else
    warn "Go не найден — локальная сборка бинарника будет пропущена (можно ставить Docker или использовать готовые пакеты)"
  fi
  if command -v docker >/dev/null; then
    log "Docker найден"
    DOCKER_AVAILABLE=true
  else
    warn "Docker не найден (режим Docker недоступен)"; DOCKER_AVAILABLE=false
  fi
}

# === Установка WireGuard (инструменты) ===
install_wireguard_tools() {
  step "Установка WireGuard инструментов..."
  case "$OS" in
    "Ubuntu"|"Debian GNU/Linux"|"Debian")
      sudo apt-get update
      sudo apt-get install -y wireguard wireguard-tools iproute2 iptables
      # Для старых ядер (<5.6) — модуль через DKMS:
      if [[ "$(uname -r | cut -d. -f1-2)" < "5.6" ]]; then
        sudo apt-get install -y wireguard-dkms
      fi
      ;;
    "CentOS Linux"|"Red Hat Enterprise Linux"|"Rocky Linux"|"AlmaLinux"|"Fedora")
      # См. оф. инструкцию WireGuard + EPEL/ELRepo при необходимости. :contentReference[oaicite:2]{index=2}
      if command -v dnf >/dev/null; then PM=dnf; else PM=yum; fi
      sudo $PM install -y epel-release || true
      sudo $PM install -y wireguard-tools iproute iptables
      ;;
    "macOS")
      # CLI утилиты доступны через brew. Также есть приложение в App Store. :contentReference[oaicite:3]{index=3}
      if command -v brew >/dev/null; then
        brew install wireguard-tools
      else
        err "Homebrew не установлен. Установите Homebrew и повторите: brew install wireguard-tools"
        exit 1
      fi
      ;;
    *)
      warn "Авто-установка для \"$OS\" не поддержана. См. оф. инструкцию: https://www.wireguard.com/install/"
      ;;
  esac
  log "wg / wg-quick доступны (wg-quick упрощает bring-up интерфейса). :contentReference[oaicite:4]{index=4}"
}

# === Сборка wireguard-go (userspace), если нужно ===
build_wireguard_go() {
  step "Сборка wireguard-go (userspace) из GitHub mirror..."
  local t; t="$(mktemp -d)"; pushd "$t" >/dev/null
  git clone --depth=1 "$WIREGUARD_GO_MIRROR"
  cd wireguard-go
  make
  sudo install -m 0755 wireguard-go "$INSTALL_DIR/wireguard-go"
  popd >/dev/null; rm -rf "$t"
  log "Установлено: $INSTALL_DIR/wireguard-go (mirror-only репозиторий). :contentReference[oaicite:5]{index=5}"
}

# === Сборка CloudBridge Relay (локально, если есть go) ===
build_cloudbridge_relay() {
  if ! command -v go >/dev/null; then
    warn "Go не найден — пропускаю локальную сборку CloudBridge Relay"
    return
  fi
  step "Сборка CloudBridge Relay..."
  local t; t="$(mktemp -d)"; pushd "$t" >/dev/null
  git clone "$REPO_URL" .
  make build
  sudo install -m 0755 "bin/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
  popd >/dev/null; rm -rf "$t"
  log "Установлено: $INSTALL_DIR/$BINARY_NAME"
}

# === Конфигурация ===
create_config() {
  step "Создание конфигурации..."
  sudo mkdir -p "$CONFIG_DIR"
  cat > /tmp/config.yaml <<'EOF'
# CloudBridge Relay Configuration
server:
  address: "0.0.0.0"
  port: 8080
  timeout: "30s"
  max_tunnels: 1000

wireguard:
  enabled: true
  interface:
    name: "wg-mesh0"
    listen_port: 51820
    address: "10.100.1.1/24"
  discovery:
    registry_url: "https://auth.2gc.ru/wg/peers"
    heartbeat_interval: "25s"
    max_peers: 50
  mesh:
    enabled: true
    auto_discovery: true
    max_connections: 100
    routing_strategy: "hybrid"

metrics:
  enabled: true
  port: 9090
  path: "/metrics"

logging:
  level: "info"
  format: "json"
  output: "stdout"
EOF
  sudo mv /tmp/config.yaml "$CONFIG_DIR/config.yaml"
  sudo chmod 0644 "$CONFIG_DIR/config.yaml"
  log "Конфигурация: $CONFIG_DIR/config.yaml"
}

# === systemd сервис с харденигом и доступом к TUN ===
create_systemd_service() {
  step "Создание systemd сервиса..."
  # Нужны права на сетевые операции (NET_ADMIN) и доступ к /dev/net/tun
  # Используем набор безопасных опций systemd (Protect*, NoNewPrivileges и т.д.). :contentReference[oaicite:6]{index=6}
  cat > /tmp/cloudbridge-relay.service <<EOF
[Unit]
Description=CloudBridge Relay (WireGuard)
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=root
Group=root
Environment=CONFIG_PATH=$CONFIG_DIR/config.yaml
ExecStart=$INSTALL_DIR/$BINARY_NAME --config \$CONFIG_PATH
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=cloudbridge-relay

# Capabilities / devices
CapabilityBoundingSet=CAP_NET_ADMIN CAP_NET_RAW
AmbientCapabilities=CAP_NET_ADMIN CAP_NET_RAW
DeviceAllow=/dev/net/tun rw
PrivateDevices=yes
DevicePolicy=closed

# Hardening
NoNewPrivileges=yes
PrivateTmp=yes
ProtectSystem=strict
ProtectHome=read-only
ProtectControlGroups=yes
ProtectKernelModules=yes
ProtectKernelTunables=yes
RestrictAddressFamilies=AF_UNIX AF_INET AF_INET6 AF_NETLINK
RestrictNamespaces=yes
RestrictRealtime=yes
RestrictSUIDSGID=yes
LockPersonality=yes
MemoryDenyWriteExecute=yes
ReadWritePaths=$CONFIG_DIR /var/run/wireguard /etc/wireguard

[Install]
WantedBy=multi-user.target
EOF
  sudo mv /tmp/cloudbridge-relay.service "$SERVICE_DIR/cloudbridge-relay.service"
  sudo systemctl daemon-reload
  sudo systemctl enable cloudbridge-relay
  log "systemd сервис установлен и включен"
}

# === Установка/запуск через Docker (с привилегиями WG) ===
install_docker() {
  step "Установка через Docker..."
  if [[ "${DOCKER_AVAILABLE:-false}" != "true" ]]; then err "Docker недоступен"; exit 1; fi

  sudo mkdir -p "$CONFIG_DIR" /etc/wireguard /var/run/wireguard

  log "Загрузка образа: $DOCKER_IMAGE:latest"
  docker pull "$DOCKER_IMAGE:latest"

  log "Запуск контейнера с CAP_NET_ADMIN и /dev/net/tun..."
  docker run -d --name cloudbridge-relay \
    --restart unless-stopped \
    --cap-add=NET_ADMIN \
    --device=/dev/net/tun \
    -p 8080:8080 -p 9090:9090 -p 51820:51820/udp \
    -v "$CONFIG_DIR":/app/configs:ro \
    -v /etc/wireguard:/etc/wireguard \
    -v /var/run/wireguard:/var/run/wireguard \
    "$DOCKER_IMAGE:latest" \
    --config /app/configs/config.yaml

  log "Контейнер запущен. Для WireGuard в контейнере требуются NET_ADMIN и /dev/net/tun. :contentReference[oaicite:7]{index=7}"
}

# === Проверки ===
verify_installation() {
  step "Проверка установки..."
  if [[ -x "$INSTALL_DIR/$BINARY_NAME" ]]; then
    log "CloudBridge Relay: $("$INSTALL_DIR/$BINARY_NAME" --version || echo ok)"
  else
    warn "Бинарник CloudBridge Relay отсутствует (нормально при Docker-инсталляции)"
  fi
  if command -v wg >/dev/null; then log "wg OK"; else warn "wg не найден"; fi
  if [[ -x "$INSTALL_DIR/wireguard-go" ]]; then "$INSTALL_DIR/wireguard-go" -version || true; fi
  [[ -f "$CONFIG_DIR/config.yaml" ]] || { err "Нет $CONFIG_DIR/config.yaml"; exit 1; }
  log "ОК"
}

# === Справка ===
usage() {
  cat <<EOF
Использование: $0 [опции]
  --help | -h         Показать помощь
  --version | -v      Показать версию
  --docker            Установка/запуск через Docker (рекомендуется без локальной сборки)
  --no-wireguard      Пропустить установку WireGuard инструментов и wireguard-go
  --config-only       Только создать конфигурацию
  --no-build          Не собирать CloudBridge Relay локально (использовать только Docker или уже установленный бинарь)
  --no-wggo           Не собирать wireguard-go (если ядро имеет модуль WireGuard)
EOF
}

# === main ===
main() {
  banner
  local DOCKER_MODE=false NO_WG=false CONFIG_ONLY=false NO_BUILD=false NO_WGGO=false

  while [[ $# -gt 0 ]]; do
    case "$1" in
      --help|-h) usage; exit 0 ;;
      --version|-v) echo "CloudBridge Relay Installer v$VERSION"; exit 0 ;;
      --docker) DOCKER_MODE=true; shift ;;
      --no-wireguard) NO_WG=true; shift ;;
      --config-only) CONFIG_ONLY=true; shift ;;
      --no-build) NO_BUILD=true; shift ;;
      --no-wggo) NO_WGGO=true; shift ;;
      *) err "Неизвестная опция: $1"; usage; exit 1 ;;
    esac
  done

  # root
  if [[ $EUID -ne 0 ]]; then err "Запустите с правами root (sudo)"; exit 1; fi

  detect_os
  check_base_deps

  if $CONFIG_ONLY; then create_config; log "Готово"; exit 0; fi

  # Инструменты WG (wg / wg-quick) — полезны в любом режиме. :contentReference[oaicite:8]{index=8}
  if ! $NO_WG; then install_wireguard_tools; fi

  # userspace wireguard-go — только если требуется (нет ядрового модуля / контейнер). :contentReference[oaicite:9]{index=9}
  if ! $NO_WG && ! $NO_WGGO; then build_wireguard_go; fi

  # бинарник relay
  if ! $NO_BUILD; then build_cloudbridge_relay; fi

  create_config

  if $DOCKER_MODE; then
    install_docker
  else
    create_systemd_service
    systemctl restart cloudbridge-relay
    systemctl --no-pager status cloudbridge-relay || true
  fi

  verify_installation

  echo -e "\n${GREEN}Установка завершена успешно!${NC}\n"
  echo -e "${BLUE}Управление:${NC}
  • systemctl start|stop|status cloudbridge-relay
  • journalctl -u cloudbridge-relay -f"
  echo -e "${BLUE}Доступ:${NC}
  • API:     http://localhost:8080
  • Metrics: http://localhost:9090/metrics
  • WG:      wg show wg-mesh0"
  echo -e "${BLUE}Справка по WireGuard:${NC}
  • Установка: https://www.wireguard.com/install/
  • QuickStart + wg-quick: https://www.wireguard.com/quickstart/"
}

main "$@"
