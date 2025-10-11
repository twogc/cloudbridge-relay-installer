#!/bin/bash

# Скрипт для добавления GitHub доменов в hosts файл
# Использует TCP DNS для получения IP адресов

echo "Добавление GitHub доменов в hosts файл..."

# Список GitHub доменов
GITHUB_DOMAINS=(
    "github.com"
    "codeload.github.com"
    "api.github.com"
    "raw.githubusercontent.com"
    "gist.github.com"
    "github.io"
    "githubusercontent.com"
    "githubassets.com"
    "githubapp.com"
    "github.blog"
    "githubstatus.com"
    "github.dev"
    "githubpreview.dev"
    "githubpreview.com"
    "githubpreview.net"
    "githubpreview.org"
    "githubpreview.io"
    "githubpreview.co"
    "githubpreview.me"
    "githubpreview.tv"
    "githubpreview.cc"
    "githubpreview.biz"
    "githubpreview.info"
    "githubpreview.mobi"
    "githubpreview.name"
    "githubpreview.pro"
    "githubpreview.tech"
    "githubpreview.us"
    "githubpreview.ws"
    "githubpreview.xyz"
)

# Функция для получения IP адреса через TCP DNS
get_ip() {
    local domain=$1
    local ip=$(dig +tcp @1.1.1.1 $domain +short | head -1)
    if [ -n "$ip" ] && [[ $ip =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        echo "$ip"
    else
        echo ""
    fi
}

# Добавляем домены в hosts файл
for domain in "${GITHUB_DOMAINS[@]}"; do
    echo "Обрабатываем $domain..."
    ip=$(get_ip $domain)
    if [ -n "$ip" ]; then
        echo "Найден IP для $domain: $ip"
        # Проверяем, есть ли уже запись
        if ! grep -q "$domain" /etc/hosts; then
            echo "$ip $domain" | sudo tee -a /etc/hosts
            echo "Добавлен $domain -> $ip"
        else
            echo "$domain уже есть в hosts файле"
        fi
    else
        echo "Не удалось получить IP для $domain"
    fi
done

echo "Готово! GitHub домены добавлены в hosts файл."
echo "Проверьте: cat /etc/hosts | grep github"

