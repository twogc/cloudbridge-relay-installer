#!/bin/bash
set -e

# Копируем production .env в корень
cp -f .env .env

echo "[INFO] Production .env скопирован в корень проекта." 