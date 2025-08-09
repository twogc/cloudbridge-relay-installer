#!/bin/bash

# CloudBridge Relay WireGuard - Универсальный скрипт установки
# Поддерживает: Linux, macOS, Windows (WSL)
# Автоматическая установка и настройка

set -e

# Цвета для вывода
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Переменные
VERSION="1.0.0"
BINARY_NAME="cloudbridge-relay"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/cloudbridge-relay"
SERVICE_DIR="/etc/systemd/system"
REPO_URL="https://github.com/twogc/cloudbridge-relay-installer"
DOCKER_IMAGE="twogc/cloudbridge-relay"

# Функции
print_banner() {
    echo -e "${BLUE}"
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║                    CloudBridge Relay                        ║"
    echo "║                   WireGuard Edition                         ║"
    echo "║                                                              ║"
    echo "║  Enterprise-grade высокопроизводительный relay сервер       ║"
    echo "║  с интеграцией WireGuard P2P Mesh Network                   ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
}

print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

# Определение ОС
detect_os() {
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        if [[ -f /etc/os-release ]]; then
            . /etc/os-release
            OS=$NAME
            VER=$VERSION_ID
        else
            OS=$(uname -s)
            VER=$(uname -r)
        fi
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        OS="macOS"
        VER=$(sw_vers -productVersion)
    elif [[ "$OSTYPE" == "cygwin" ]] || [[ "$OSTYPE" == "msys" ]]; then
        OS="Windows"
        VER=$(uname -r)
    else
        OS="Unknown"
        VER="Unknown"
    fi
    
    print_info "Обнаружена ОС: $OS $VER"
}

# Проверка зависимостей
check_dependencies() {
    print_step "Проверка зависимостей..."
    
    # Проверка Go
    if ! command -v go &> /dev/null; then
        print_error "Go не установлен. Установите Go 1.21+"
        print_info "Скачать: https://golang.org/dl/"
        exit 1
    fi
    
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    print_info "Go версия: $GO_VERSION"
    
    # Проверка Git
    if ! command -v git &> /dev/null; then
        print_error "Git не установлен"
        exit 1
    fi
    
    print_info "Git найден"
    
    # Проверка Docker (опционально)
    if command -v docker &> /dev/null; then
        print_info "Docker найден"
        DOCKER_AVAILABLE=true
    else
        print_warn "Docker не найден (некоторые функции недоступны)"
        DOCKER_AVAILABLE=false
    fi
}

# Установка WireGuard зависимостей
install_wireguard_deps() {
    print_step "Установка WireGuard зависимостей..."
    
    case $OS in
        "Ubuntu"|"Debian GNU/Linux")
            print_info "Установка WireGuard для Ubuntu/Debian..."
            sudo apt-get update
            sudo apt-get install -y \
                build-essential \
                git \
                iproute2 \
                wireguard-tools \
                wireguard-dkms
            ;;
        "CentOS Linux"|"Red Hat Enterprise Linux"|"Fedora")
            print_info "Установка WireGuard для CentOS/RHEL/Fedora..."
            sudo yum install -y \
                gcc \
                git \
                iproute \
                kernel-devel \
                wireguard-tools
            ;;
        "macOS")
            print_info "Установка WireGuard для macOS..."
            if command -v brew &> /dev/null; then
                brew install wireguard-tools
            else
                print_error "Homebrew не установлен. Установите Homebrew"
                exit 1
            fi
            ;;
        *)
            print_warn "Автоматическая установка WireGuard не поддерживается для $OS"
            print_info "Установите WireGuard вручную: https://www.wireguard.com/install/"
            ;;
    esac
}

# Сборка WireGuard-go
build_wireguard_go() {
    print_step "Сборка WireGuard-go..."
    
    TEMP_DIR=$(mktemp -d)
    cd $TEMP_DIR
    
    print_info "Клонирование WireGuard-go..."
    git clone https://git.zx2c4.com/wireguard-go.git
    cd wireguard-go
    
    print_info "Сборка WireGuard-go..."
    make
    
    print_info "Установка WireGuard-go..."
    sudo cp wireguard-go $INSTALL_DIR/
    sudo chmod +x $INSTALL_DIR/wireguard-go
    
    cd /
    rm -rf $TEMP_DIR
    
    print_info "WireGuard-go установлен в $INSTALL_DIR/wireguard-go"
}

# Сборка CloudBridge Relay
build_cloudbridge_relay() {
    print_step "Сборка CloudBridge Relay..."
    
    # Создание временной директории
    TEMP_DIR=$(mktemp -d)
    cd $TEMP_DIR
    
    print_info "Клонирование репозитория..."
    git clone $REPO_URL .
    
    print_info "Сборка CloudBridge Relay..."
    make build
    
    print_info "Установка CloudBridge Relay..."
    sudo cp bin/$BINARY_NAME $INSTALL_DIR/
    sudo chmod +x $INSTALL_DIR/$BINARY_NAME
    
    cd /
    rm -rf $TEMP_DIR
    
    print_info "CloudBridge Relay установлен в $INSTALL_DIR/$BINARY_NAME"
}

# Создание конфигурации
create_config() {
    print_step "Создание конфигурации..."
    
    sudo mkdir -p $CONFIG_DIR
    
    cat > /tmp/config.yaml << EOF
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
    registry_url: "https://edge.2gc.ru/wg/peers"
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
    
    sudo cp /tmp/config.yaml $CONFIG_DIR/
    rm /tmp/config.yaml
    
    print_info "Конфигурация создана в $CONFIG_DIR/config.yaml"
}

# Создание systemd сервиса
create_systemd_service() {
    print_step "Создание systemd сервиса..."
    
    cat > /tmp/cloudbridge-relay.service << EOF
[Unit]
Description=CloudBridge Relay WireGuard
After=network.target
Wants=network.target

[Service]
Type=simple
User=root
Group=root
ExecStart=$INSTALL_DIR/$BINARY_NAME --config $CONFIG_DIR/config.yaml
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=cloudbridge-relay

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=$CONFIG_DIR /var/run/wireguard /etc/wireguard

[Install]
WantedBy=multi-user.target
EOF
    
    sudo cp /tmp/cloudbridge-relay.service $SERVICE_DIR/
    rm /tmp/cloudbridge-relay.service
    
    sudo systemctl daemon-reload
    sudo systemctl enable cloudbridge-relay
    
    print_info "Systemd сервис создан и включен"
}

# Установка через Docker
install_docker() {
    print_step "Установка через Docker..."
    
    if [ "$DOCKER_AVAILABLE" = false ]; then
        print_error "Docker недоступен"
        exit 1
    fi
    
    print_info "Скачивание Docker образа..."
    docker pull $DOCKER_IMAGE:latest
    
    print_info "Создание Docker контейнера..."
    docker run -d \
        --name cloudbridge-relay \
        --restart unless-stopped \
        -p 8080:8080 \
        -p 9090:9090 \
        -p 51820:51820/udp \
        -v $CONFIG_DIR:/app/configs \
        -v /etc/wireguard:/etc/wireguard \
        -v /var/run/wireguard:/var/run/wireguard \
        $DOCKER_IMAGE:latest
    
    print_info "Docker контейнер запущен"
}

# Проверка установки
verify_installation() {
    print_step "Проверка установки..."
    
    # Проверка бинарного файла
    if [ -f "$INSTALL_DIR/$BINARY_NAME" ]; then
        print_info "CloudBridge Relay установлен"
        $INSTALL_DIR/$BINARY_NAME --version
    else
        print_error "CloudBridge Relay не найден"
        exit 1
    fi
    
    # Проверка WireGuard-go
    if [ -f "$INSTALL_DIR/wireguard-go" ]; then
        print_info "WireGuard-go установлен"
        $INSTALL_DIR/wireguard-go --version
    else
        print_error "WireGuard-go не найден"
        exit 1
    fi
    
    # Проверка конфигурации
    if [ -f "$CONFIG_DIR/config.yaml" ]; then
        print_info "Конфигурация создана"
    else
        print_error "Конфигурация не найдена"
        exit 1
    fi
    
    print_info "Установка завершена успешно!"
}

# Показать справку
show_help() {
    echo "Использование: $0 [опции]"
    echo ""
    echo "Опции:"
    echo "  --help, -h          Показать эту справку"
    echo "  --version, -v       Показать версию"
    echo "  --docker            Установить через Docker"
    echo "  --no-wireguard      Пропустить установку WireGuard"
    echo "  --config-only       Только создать конфигурацию"
    echo ""
    echo "Примеры:"
    echo "  $0                  Полная установка"
    echo "  $0 --docker         Установка через Docker"
    echo "  $0 --config-only    Только конфигурация"
}

# Основная функция
main() {
    print_banner
    
    # Парсинг аргументов
    DOCKER_MODE=false
    NO_WIREGUARD=false
    CONFIG_ONLY=false
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --help|-h)
                show_help
                exit 0
                ;;
            --version|-v)
                echo "CloudBridge Relay WireGuard v$VERSION"
                exit 0
                ;;
            --docker)
                DOCKER_MODE=true
                shift
                ;;
            --no-wireguard)
                NO_WIREGUARD=true
                shift
                ;;
            --config-only)
                CONFIG_ONLY=true
                shift
                ;;
            *)
                print_error "Неизвестная опция: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    # Проверка root прав
    if [[ $EUID -ne 0 ]]; then
        print_error "Этот скрипт должен быть запущен с правами root (sudo)"
        exit 1
    fi
    
    detect_os
    check_dependencies
    
    if [ "$CONFIG_ONLY" = true ]; then
        create_config
        print_info "Конфигурация создана"
        exit 0
    fi
    
    if [ "$DOCKER_MODE" = true ]; then
        install_docker
    else
        if [ "$NO_WIREGUARD" = false ]; then
            install_wireguard_deps
            build_wireguard_go
        fi
        build_cloudbridge_relay
        create_config
        create_systemd_service
    fi
    
    verify_installation
    
    echo ""
    echo -e "${GREEN}🎉 Установка завершена успешно!${NC}"
    echo ""
    echo -e "${BLUE}📋 Что было установлено:${NC}"
    echo "  • CloudBridge Relay с WireGuard интеграцией"
    echo "  • WireGuard-go (пользовательская реализация)"
    echo "  • Systemd сервис (автозапуск)"
    echo "  • Конфигурационные файлы"
    echo ""
    echo -e "${BLUE}🔧 Управление:${NC}"
    echo "  • systemctl start cloudbridge-relay    - запуск"
    echo "  • systemctl stop cloudbridge-relay     - остановка"
    echo "  • systemctl status cloudbridge-relay   - статус"
    echo "  • journalctl -u cloudbridge-relay -f   - логи"
    echo ""
    echo -e "${BLUE}🌐 Доступ:${NC}"
    echo "  • Relay сервер: http://localhost:8080"
    echo "  • Метрики: http://localhost:9090/metrics"
    echo "  • WireGuard: wg show wg-mesh0"
    echo ""
    echo -e "${BLUE}📖 Документация:${NC}"
    echo "  • https://github.com/twogc/cloudbridge-relay-installer"
    echo "  • https://www.wireguard.com/"
}

# Запуск
main "$@"
