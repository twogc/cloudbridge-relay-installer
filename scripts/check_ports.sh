#!/bin/bash

# Скрипт для проверки правильности настройки портов
# CloudBridge Relay Server

set -e

echo "Проверка настройки портов CloudBridge Relay..."

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Проверка Dockerfile
echo -e "${BLUE}Проверка Dockerfile...${NC}"

if grep -q "EXPOSE 9090 9090/udp 51820/udp 8084" Dockerfile; then
    echo -e "${GREEN}Dockerfile: порты настроены правильно${NC}"
else
    echo -e "${RED}Dockerfile: порты настроены неправильно${NC}"
    echo "   Ожидается: EXPOSE 9090 9090/udp 51820/udp 8084"
    echo "   Найдено: $(grep '^EXPOSE' Dockerfile)"
fi

if grep -q "http://localhost:9090/health" Dockerfile; then
    echo -e "${GREEN}Dockerfile: health check настроен правильно${NC}"
else
    echo -e "${RED}Dockerfile: health check настроен неправильно${NC}"
fi

# Проверка docker-compose.yml
echo -e "${BLUE}Проверка docker-compose.yml...${NC}"

if grep -q '"9090:9090"' docker-compose.yml; then
    echo -e "${GREEN}docker-compose.yml: порт 9090 TCP настроен${NC}"
else
    echo -e "${RED}docker-compose.yml: порт 9090 TCP не найден${NC}"
fi

if grep -q '"9090:9090/udp"' docker-compose.yml; then
    echo -e "${GREEN}docker-compose.yml: порт 9090 UDP настроен${NC}"
else
    echo -e "${RED}docker-compose.yml: порт 9090 UDP не найден${NC}"
fi

if grep -q '"51820:51820/udp"' docker-compose.yml; then
    echo -e "${GREEN}docker-compose.yml: порт 51820 UDP настроен${NC}"
else
    echo -e "${RED}docker-compose.yml: порт 51820 UDP не найден${NC}"
fi

if grep -q '"8084:8084"' docker-compose.yml; then
    echo -e "${GREEN}docker-compose.yml: порт 8084 TCP настроен${NC}"
else
    echo -e "${RED}docker-compose.yml: порт 8084 TCP не найден${NC}"
fi

# Проверка config.json
echo -e "${BLUE}Проверка config.json...${NC}"

if grep -q '"port": 9090' configs/config.json; then
    echo -e "${GREEN}config.json: API порт 9090 настроен${NC}"
else
    echo -e "${RED}config.json: API порт 9090 не найден${NC}"
fi

if grep -q '"port": 8084' configs/config.json; then
    echo -e "${GREEN}config.json: Metrics порт 8084 настроен${NC}"
else
    echo -e "${RED}config.json: Metrics порт 8084 не найден${NC}"
fi

if grep -A 5 '"enhanced_quic"' configs/config.json | grep -q '"port": 9090'; then
    echo -e "${GREEN}config.json: QUIC порт 9090 настроен${NC}"
else
    echo -e "${RED}config.json: QUIC порт 9090 не найден${NC}"
fi

# Проверка health check в config.json
if grep -q "http://localhost:9090/health" configs/config.json; then
    echo -e "${GREEN}config.json: health check настроен правильно${NC}"
else
    echo -e "${RED}config.json: health check настроен неправильно${NC}"
fi

# Проверка docker-compose health check
if grep -q "http://localhost:9090/health" docker-compose.yml; then
    echo -e "${GREEN}docker-compose.yml: health check настроен правильно${NC}"
else
    echo -e "${RED}docker-compose.yml: health check настроен неправильно${NC}"
fi

echo ""
echo -e "${BLUE}Сводка по портам:${NC}"
echo "  TCP 9090  - Main API и QUIC (HTTP/2 fallback)"
echo "  UDP 9090  - QUIC Protocol"
echo "  UDP 51820 - WireGuard P2P Mesh"
echo "  TCP 8084  - Metrics (Prometheus)"
echo "  TCP 9091  - Prometheus (внешний)"
echo "  TCP 3000  - Grafana (внешний)"

echo ""
echo -e "${BLUE}Проверка неиспользуемых портов:${NC}"

# Проверяем, не остались ли старые порты
OLD_PORTS=("8080" "8081" "8082" "80")
for port in "${OLD_PORTS[@]}"; do
    if grep -q ":$port" docker-compose.yml || grep -q "EXPOSE.*$port" Dockerfile; then
        echo -e "${YELLOW}  Порт $port все еще используется (возможно устарел)${NC}"
    fi
done

echo ""
echo -e "${GREEN}Проверка портов завершена!${NC}"
