#!/bin/bash

# Скрипт для установки WireGuard-go и зависимостей
# Использование: ./scripts/install_wireguard_go.sh

set -e

echo "🔧 Установка WireGuard-go и зависимостей..."

# Проверяем, что мы root или используем sudo
if [[ $EUID -ne 0 ]]; then
   echo "❌ Этот скрипт должен быть запущен с правами root (sudo)"
   exit 1
fi

# Определяем дистрибутив
if [ -f /etc/os-release ]; then
    . /etc/os-release
    OS=$NAME
    VER=$VERSION_ID
else
    echo "❌ Не удалось определить дистрибутив"
    exit 1
fi

echo "📦 Дистрибутив: $OS $VER"

# Установка зависимостей
install_dependencies() {
    echo "📦 Установка зависимостей..."
    
    case $OS in
        "Ubuntu"|"Debian GNU/Linux")
            apt-get update
            apt-get install -y \
                build-essential \
                git \
                golang-go \
                iproute2 \
                wireguard-tools \
                wireguard-dkms
            ;;
        "CentOS Linux"|"Red Hat Enterprise Linux")
            yum install -y \
                gcc \
                git \
                golang \
                iproute \
                kernel-devel \
                wireguard-tools
            ;;
        *)
            echo "❌ Неподдерживаемый дистрибутив: $OS"
            exit 1
            ;;
    esac
}

# Сборка WireGuard-go
build_wireguard_go() {
    echo "🔨 Сборка WireGuard-go..."
    
    # Создаем временную директорию
    TEMP_DIR=$(mktemp -d)
    cd $TEMP_DIR
    
    # Клонируем репозиторий
    git clone https://git.zx2c4.com/wireguard-go.git
    cd wireguard-go
    
    # Собираем
    make
    
    # Устанавливаем
    cp wireguard-go /usr/local/bin/
    chmod +x /usr/local/bin/wireguard-go
    
    # Очищаем
    cd /
    rm -rf $TEMP_DIR
    
    echo "WireGuard-go установлен в /usr/local/bin/wireguard-go"
}

# Создание директорий
create_directories() {
    echo "📁 Создание необходимых директорий..."
    
    mkdir -p /etc/wireguard
    mkdir -p /var/run/wireguard
    
    # Устанавливаем права
    chmod 755 /etc/wireguard
    chmod 755 /var/run/wireguard
    
    echo "Директории созданы"
}

# Проверка установки
verify_installation() {
    echo "🔍 Проверка установки..."
    
    # Проверяем wireguard-go
    if command -v wireguard-go >/dev/null 2>&1; then
        echo "wireguard-go найден"
        wireguard-go --version
    else
        echo "❌ wireguard-go не найден"
        exit 1
    fi
    
    # Проверяем wg
    if command -v wg >/dev/null 2>&1; then
        echo "wg найден"
        wg version
    else
        echo "❌ wg не найден"
        exit 1
    fi
    
    # Проверяем ip
    if command -v ip >/dev/null 2>&1; then
        echo "ip найден"
    else
        echo "❌ ip не найден"
        exit 1
    fi
    
    echo "Все компоненты установлены успешно"
}

# Настройка systemd сервиса (опционально)
setup_systemd_service() {
    echo "🔧 Настройка systemd сервиса..."
    
    cat > /etc/systemd/system/wireguard-go@.service << EOF
[Unit]
Description=WireGuard Go for %I
After=network.target

[Service]
Type=notify
ExecStart=/usr/local/bin/wireguard-go %i
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    echo "Systemd сервис настроен"
}

# Основная функция
main() {
    echo "🚀 Начинаем установку WireGuard-go..."
    
    install_dependencies
    build_wireguard_go
    create_directories
    verify_installation
    setup_systemd_service
    
    echo ""
    echo "🎉 Установка завершена успешно!"
    echo ""
    echo "Что было установлено:"
    echo "  • wireguard-go - пользовательская реализация WireGuard"
    echo "  • wireguard-tools - утилиты для управления WireGuard"
    echo "  • iproute2 - утилиты для управления сетью"
    echo ""
    echo "📁 Директории:"
    echo "  • /etc/wireguard - конфигурационные файлы"
    echo "  • /var/run/wireguard - сокеты управления"
    echo ""
    echo "🔧 Использование:"
    echo "  • wireguard-go <interface_name> - запуск интерфейса"
    echo "  • wg show <interface_name> - просмотр статуса"
    echo "  • systemctl start wireguard-go@<interface_name> - запуск через systemd"
    echo ""
    echo "📖 Документация: https://www.wireguard.com/"
}

# Запуск
main "$@"
