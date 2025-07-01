#!/bin/bash
set -e

for dep in docker docker-compose jq curl; do
  if ! command -v $dep &>/dev/null; then
    echo -e "\033[0;31m[ERROR]\033[0m Не найдено: $dep"
    exit 1
  else
    echo -e "\033[0;32m[OK]\033[0m $dep"
  fi
done
echo -e "\033[1;32mВсе зависимости установлены!\033[0m" 