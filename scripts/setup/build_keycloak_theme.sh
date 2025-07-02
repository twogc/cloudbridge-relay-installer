#!/bin/bash

# Упрощённый скрипт сборки Vue темы для Keycloak
# Собирает тему без автоматического перезапуска контейнера Keycloak

set -e

KEYCLOAK_VUE_DIR="keycloak-vue"

cd "$KEYCLOAK_VUE_DIR"

if [ ! -d "node_modules" ]; then
  npm install
fi

npm run build

cd ..

echo "\n✔️ Сборка темы завершена."
echo "\nЕсли хотите применить тему — полностью остановите и запустите Keycloak вручную:"
echo "  docker-compose stop keycloak && docker-compose up -d keycloak"
echo "  ./scripts/setup/check_keycloak_theme.sh" 