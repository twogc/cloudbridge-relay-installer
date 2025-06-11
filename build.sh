#!/bin/bash

set -e

# Проверка наличия Go
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed"
    exit 1
fi

# Создание директории для сборки
mkdir -p build

# Сборка релиза
echo "Building relay server..."
go build -o build/relay cmd/relay/main.go

# Создание архива установщика
echo "Creating installer archive..."
mkdir -p build/installer/opt/relay
cp build/relay build/installer/opt/relay/
cp systemd/relay.service build/installer/etc/systemd/system/
cp install.sh build/installer/

# Упаковка в tar.gz
cd build/installer
tar czf ../relay-installer.tar.gz .
cd ../..

echo "Build complete: build/relay-installer.tar.gz" 