#!/bin/bash

# CloudBridge Relay - Помощник настройки GitHub Secrets
# Этот скрипт поможет вам создать и настроить необходимые секреты

echo "🔐 Настройка GitHub Secrets для CloudBridge Relay"
echo "================================================="
echo ""

echo "📋 Необходимые секреты:"
echo ""
echo "1. GH_TOKEN - Personal Access Token с правами:"
echo "   ✅ repo (Full control)"
echo "   ✅ workflow (Update workflows)"
echo "   ✅ write:packages (Upload packages)"
echo ""
echo "2. DOCKERHUB_USERNAME - Логин Docker Hub"
echo "   Значение: twogc"
echo ""
echo "3. DOCKERHUB_TOKEN - Access Token от Docker Hub"
echo "   Создать: https://hub.docker.com/settings/security"
echo ""

echo "🔗 Ссылки для настройки:"
echo ""
echo "📍 GitHub Personal Access Token:"
echo "   https://github.com/settings/tokens"
echo ""
echo "📍 Docker Hub Access Token:"  
echo "   https://hub.docker.com/settings/security"
echo ""
echo "📍 GitHub Repository Secrets:"
echo "   https://github.com/twogc/cloudbridge-relay-installer/settings/secrets/actions"
echo ""

read -p "❓ Хотите проверить текущие секреты? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo ""
    echo "🔍 Проверка доступности GitHub API..."
    
    if [ -z "$GITHUB_TOKEN" ]; then
        echo "❌ GITHUB_TOKEN не установлен в переменных окружения"
        echo "💡 Установите: export GITHUB_TOKEN=your_token"
        echo "💡 В GitHub Secrets используйте имя: GH_TOKEN"
    else
        echo "✅ GITHUB_TOKEN найден в переменных окружения"
        
        # Проверка токена
        response=$(curl -s -H "Authorization: token $GITHUB_TOKEN" https://api.github.com/user)
        if echo "$response" | grep -q '"login"'; then
            username=$(echo "$response" | grep '"login"' | cut -d'"' -f4)
            echo "✅ Токен валиден. Пользователь: $username"
            
            # Проверка прав workflow
            scopes=$(curl -s -I -H "Authorization: token $GITHUB_TOKEN" https://api.github.com/user | grep -i "x-oauth-scopes" | cut -d' ' -f2-)
            if echo "$scopes" | grep -q "workflow"; then
                echo "✅ Права workflow: ОК"
            else
                echo "❌ Права workflow: НЕ НАЙДЕНЫ"
                echo "💡 Добавьте scope 'workflow' к токену"
            fi
        else
            echo "❌ Токен недействителен"
        fi
    fi
fi

echo ""
echo "📖 Подробная инструкция:"
echo "   docs/guides/github_secrets_setup.md"
echo ""
echo "🚀 После настройки секретов запустите:"
echo "   git push origin main"
echo "   # или создайте release:"
echo "   git tag v1.0.0 && git push origin v1.0.0"
