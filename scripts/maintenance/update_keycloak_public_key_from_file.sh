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

# Логирование: путь и первые/последние строки PEM-файла
PEM_FIRST_LINE=$(head -n 1 "$PEM_FILE")
PEM_LAST_LINE=$(tail -n 1 "$PEM_FILE")
echo "[DEBUG] Используется PEM-файл: $PEM_FILE"
echo "[DEBUG] Первая строка PEM: $PEM_FIRST_LINE"
echo "[DEBUG] Последняя строка PEM: $PEM_LAST_LINE"

# Преобразовать для JSON (заменить реальные переносы на \n)
PEM_JSON=$(echo -e "$PEM_CONTENT" | awk '{printf "%s\\n", $0}' | sed 's/\\n$//')

# Логирование: первые и последние 40 символов PEM_JSON
PEM_JSON_START=${PEM_JSON:0:40}
PEM_JSON_END=${PEM_JSON: -40}
echo "[DEBUG] PEM_JSON (начало): $PEM_JSON_START..."
echo "[DEBUG] PEM_JSON (конец): ...$PEM_JSON_END"

# Обновить public_key в config/config.json
jq --arg key "$PEM_JSON" '.keycloak.public_key = $key' "$CONFIG_FILE" > tmp.json && mv -f tmp.json "$CONFIG_FILE"

echo "[INFO] Публичный ключ из $PEM_FILE успешно обновлён в $CONFIG_FILE" 