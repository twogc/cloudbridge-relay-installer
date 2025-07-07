#!/bin/bash

# Скрипт для автоматического обновления public_key в config/keycloak_config.json
# Получает публичный ключ из Keycloak и подставляет его в конфиг

set -e

KEYCLOAK_URL="http://localhost:8080"
REALM="cloudbridge"
CONFIG_FILE="config/keycloak_config.json"

# Получить публичный ключ из Keycloak (base64, без PEM-обертки)
PUBLIC_KEY=$(curl -s "${KEYCLOAK_URL}/realms/${REALM}" | jq -r .public_key)

if [[ -z "$PUBLIC_KEY" || "$PUBLIC_KEY" == "null" ]]; then
  echo "[ERROR] Не удалось получить публичный ключ из Keycloak!"
  exit 1
fi

# Преобразовать в PEM-формат (64 символа в строке)
PEM_BODY=$(echo "$PUBLIC_KEY" | fold -w 64)
PEM="-----BEGIN PUBLIC KEY-----\n${PEM_BODY}\n-----END PUBLIC KEY-----"

# Преобразовать для JSON (заменить реальные переносы на \n)
PEM_JSON=$(echo -e "$PEM" | awk '{printf "%s\\n", $0}' | sed 's/\\n$//')

# Обновить public_key в config/keycloak_config.json
jq --arg key "$PEM_JSON" '.keycloak.public_key = $key' "$CONFIG_FILE" > tmp.json && mv tmp.json "$CONFIG_FILE"

echo "[INFO] Публичный ключ Keycloak успешно обновлён в $CONFIG_FILE" 