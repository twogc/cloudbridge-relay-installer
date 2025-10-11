#!/bin/bash

# Health check скрипт для Cadence сервера

set -e

# Проверяем доступность Cadence сервера
if curl -f http://localhost:7933/health >/dev/null 2>&1; then
    echo "Cadence server is healthy"
    exit 0
else
    echo "Cadence server is not responding"
    exit 1
fi
