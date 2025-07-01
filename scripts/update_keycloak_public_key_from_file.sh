#!/bin/bash

# Скрипт для обновления public_key в config/config.json из PEM-файла
# Использование: ./update_keycloak_public_key_from_file.sh [путь_к_pem_файлу]

set -e

PEM_FILE="${1:-config/public_key.pem}"
CONFIG_FILE="config/config.json"

if [[ ! -f "$PEM_FILE" ]]; then
  echo "[ERROR] Файл PEM не найден: $PEM_FILE"
  exit 1
fi

# Прочитать содержимое PEM-файла
PEM_CONTENT=$(cat "$PEM_FILE")

# Преобразовать для JSON (заменить реальные переносы на \n)
PEM_JSON=$(echo -e "$PEM_CONTENT" | awk '{printf "%s\\n", $0}' | sed 's/\\n$//')

# Обновить public_key в config/config.json
jq --arg key "$PEM_JSON" '.keycloak.public_key = $key' "$CONFIG_FILE" > tmp.json && mv tmp.json "$CONFIG_FILE"

echo "[INFO] Публичный ключ из $PEM_FILE успешно обновлён в $CONFIG_FILE" 