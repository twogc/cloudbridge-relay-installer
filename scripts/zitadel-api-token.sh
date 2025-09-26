#!/bin/bash
set -euo pipefail

# ZITADEL API Token via Direct Methods
# Получение токена для ZITADEL API в обход веб-интерфейса

ZITADEL_DOMAIN="edge.2gc.ru"
ZITADEL_API_URL="https://$ZITADEL_DOMAIN"
ADMIN_EMAIL="admin@local.test"
ADMIN_PASSWORD="jrcCaigzmqzHpRcxo61RDiS+xejxX4/K"

# Цвета
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log() { echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $*${NC}"; }
warn() { echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $*${NC}"; }
error() { echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $*${NC}"; }
info() { echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')] $*${NC}"; }

# Попытка найти системные приложения ZITADEL
discover_system_apps() {
    log "Поиск системных OIDC приложений..."
    
    # Известные системные client_id в ZITADEL
    local system_clients=(
        "zitadel-console"
        "console"
        "management-api" 
        "admin-api"
        "zitadel"
        "238390948473021473"  # Часто используемый ID консоли ZITADEL
        "238390948473021474"  # Альтернативный ID
        "zitadel-login-ui"
    )
    
    for client_id in "${system_clients[@]}"; do
        info "Тестирование client_id: $client_id"
        
        # Тест через introspect endpoint (не требует токена для проверки существования)
        local test_response
        test_response=$(curl -s -k -X POST "$ZITADEL_API_URL/oauth/v2/introspect" \
            -H "Content-Type: application/x-www-form-urlencoded" \
            -d "token=invalid_token" \
            -d "client_id=$client_id" 2>/dev/null || echo '{}')
        
        # Если client не существует, будет ошибка invalid_client
        local error_type=$(echo "$test_response" | jq -r '.error // empty' 2>/dev/null)
        
        if [[ "$error_type" != "invalid_client" ]]; then
            log "✅ Найден валидный client_id: $client_id"
            echo "$client_id"
            return 0
        fi
    done
    
    warn "Системные клиенты не найдены"
    return 1
}

# Попытка получить токен через Device Code Flow с найденным client_id
try_device_code_flow() {
    local client_id="$1"
    
    log "Попытка Device Code Flow с client_id: $client_id"
    
    # Инициализация device authorization
    local device_response
    device_response=$(curl -s -k -X POST "$ZITADEL_API_URL/oauth/v2/device_authorization" \
        -H "Content-Type: application/x-www-form-urlencoded" \
        -d "client_id=$client_id" \
        -d "scope=openid profile email urn:zitadel:iam:org:project:id:zitadel:aud" 2>/dev/null || echo '{}')
    
    local device_code=$(echo "$device_response" | jq -r '.device_code // empty')
    local user_code=$(echo "$device_response" | jq -r '.user_code // empty')
    local verification_uri=$(echo "$device_response" | jq -r '.verification_uri // empty')
    local expires_in=$(echo "$device_response" | jq -r '.expires_in // 300')
    local interval=$(echo "$device_response" | jq -r '.interval // 5')
    
    if [[ -z "$device_code" || "$device_code" == "null" ]]; then
        warn "Device Code Flow не удался: $(echo "$device_response" | jq -r '.error_description // .error // "unknown error"')"
        return 1
    fi
    
    info "📱 Device Code Flow активирован:"
    info "   User Code: $user_code"
    info "   Verification URI: $verification_uri"
    info "   Expires in: $expires_in seconds"
    info ""
    warn "⚠️  ВНИМАНИЕ: Поскольку веб-интерфейс зациклен,"
    warn "   этот метод может не сработать."
    info ""
    info "Попробуйте:"
    info "1. Открыть $verification_uri в инкогнито режиме"
    info "2. Очистить cookies для $ZITADEL_DOMAIN"
    info "3. Попробовать другой браузер"
    info ""
    
    read -p "Попробовать авторизацию? (y/N): " try_auth
    if [[ "$try_auth" != "y" && "$try_auth" != "Y" ]]; then
        return 1
    fi
    
    # Polling для получения токена
    local end_time=$(($(date +%s) + expires_in))
    info "⏳ Ожидание авторизации (максимум $expires_in секунд)..."
    
    while [[ $(date +%s) -lt $end_time ]]; do
        sleep "$interval"
        
        local token_response
        token_response=$(curl -s -k -X POST "$ZITADEL_API_URL/oauth/v2/token" \
            -H "Content-Type: application/x-www-form-urlencoded" \
            -d "grant_type=urn:ietf:params:oauth:grant-type:device_code" \
            -d "client_id=$client_id" \
            -d "device_code=$device_code" 2>/dev/null || echo '{}')
        
        local error_type=$(echo "$token_response" | jq -r '.error // empty')
        
        case "$error_type" in
            "")
                local access_token=$(echo "$token_response" | jq -r '.access_token // empty')
                if [[ -n "$access_token" && "$access_token" != "null" ]]; then
                    log "✅ Токен получен успешно!"
                    echo "$access_token"
                    return 0
                fi
                ;;
            "authorization_pending")
                info "⏳ Ожидание авторизации пользователя..."
                ;;
            "access_denied"|"expired_token")
                warn "Авторизация отклонена или истекла"
                return 1
                ;;
        esac
    done
    
    warn "Время ожидания авторизации истекло"
    return 1
}

# Попытка прямого доступа к PostgreSQL для создания токена
try_postgres_direct() {
    log "Попытка прямого доступа к PostgreSQL..."
    
    # Проверяем доступ к PostgreSQL в кластере
    local pg_pod=$(kubectl get pods -n zitadel -l app.kubernetes.io/name=postgresql -o name 2>/dev/null | head -1)
    
    if [[ -z "$pg_pod" ]]; then
        warn "PostgreSQL pod не найден в namespace zitadel"
        return 1
    fi
    
    info "Найден PostgreSQL: $pg_pod"
    
    # Попытка получить пароль из секрета
    local pg_password
    pg_password=$(kubectl get secret -n zitadel pg-auth -o jsonpath='{.data.postgres-password}' 2>/dev/null | base64 -d)
    
    if [[ -z "$pg_password" ]]; then
        # Попробуем альтернативный секрет
        pg_password=$(kubectl get secret -n zitadel pg-postgresql -o jsonpath='{.data.postgres-password}' 2>/dev/null | base64 -d)
    fi
    
    if [[ -z "$pg_password" ]]; then
        warn "Не удалось получить пароль PostgreSQL из секретов"
        return 1
    fi
    
    log "Подключение к базе данных ZITADEL..."
    
    # Получаем информацию о пользователях и их токенах
    local query_result
    query_result=$(kubectl exec -n zitadel "$pg_pod" -- psql -U postgres -d zitadel -c "
        SELECT u.id, u.username, u.state 
        FROM auth.users u 
        WHERE u.username = '$ADMIN_EMAIL'
        LIMIT 1;
    " 2>/dev/null || echo "")
    
    if echo "$query_result" | grep -q "$ADMIN_EMAIL"; then
        log "✅ Пользователь $ADMIN_EMAIL найден в базе данных"
        
        # Это продвинутый метод - создание личного токена напрямую в БД
        warn "⚠️  Прямое создание токена в БД требует глубокого понимания структуры ZITADEL"
        warn "   Это может нарушить целостность данных"
        
        info "Рекомендуется использовать другие методы получения токена"
        return 1
    else
        warn "Пользователь $ADMIN_EMAIL не найден в базе данных"
        return 1
    fi
}

# Альтернативный способ - создание временного admin user через kubectl
create_temp_admin_user() {
    log "Создание временного admin пользователя через kubectl..."
    
    # Создаем временный под с zitadel CLI
    cat > /tmp/zitadel-admin-job.yaml <<EOF
apiVersion: batch/v1
kind: Job
metadata:
  name: zitadel-create-admin-token
  namespace: zitadel
spec:
  template:
    spec:
      restartPolicy: Never
      containers:
      - name: zitadel-admin
        image: ghcr.io/zitadel/zitadel:latest
        command: ["/bin/sh", "-c"]
        args:
        - |
          echo "Попытка создания admin токена через CLI..."
          # Здесь будет логика создания токена через ZITADEL CLI
          # Пока что просто выводим доступные команды
          zitadel --help
        env:
        - name: ZITADEL_DATABASE_POSTGRES_HOST
          value: pg-postgresql.zitadel.svc.cluster.local
        - name: ZITADEL_DATABASE_POSTGRES_PORT
          value: "5432"
        - name: ZITADEL_DATABASE_POSTGRES_DATABASE
          value: zitadel
        - name: ZITADEL_DATABASE_POSTGRES_USER_USERNAME
          value: zitadel
        - name: ZITADEL_DATABASE_POSTGRES_USER_PASSWORD
          valueFrom:
            secretKeyRef:
              name: pg-auth
              key: password
        - name: ZITADEL_DATABASE_POSTGRES_USER_SSL_MODE
          value: disable
        - name: ZITADEL_EXTERNALDOMAIN
          value: edge.2gc.ru
        - name: ZITADEL_EXTERNALSECURE
          value: "true"
EOF

    kubectl apply -f /tmp/zitadel-admin-job.yaml
    
    info "Temporary job created. Checking logs..."
    
    # Ждем запуска и получаем логи
    sleep 10
    kubectl logs -n zitadel job/zitadel-create-admin-token || true
    
    # Очистка
    kubectl delete -f /tmp/zitadel-admin-job.yaml || true
    rm -f /tmp/zitadel-admin-job.yaml
    
    warn "Этот метод требует дополнительной разработки"
    return 1
}

# Инструкции по восстановлению доступа к веб-интерфейсу
fix_web_interface() {
    info "🔧 Способы восстановления доступа к веб-интерфейсу:"
    info ""
    info "1. Очистка cookies и браузерного кэша:"
    info "   - Откройте DevTools (F12)"
    info "   - Application → Storage → Clear site data"
    info "   - Или используйте режим инкогнито"
    info ""
    info "2. Проверка HAProxy/nginx конфигурации:"
    info "   - Проверьте /etc/haproxy/haproxy.cfg"
    info "   - Убедитесь в правильности redirect rules"
    info "   - Перезагрузите: systemctl reload haproxy"
    info ""
    info "3. Проверка ZITADEL логов:"
    info "   kubectl logs -n zitadel deployment/zitadel -f"
    info ""
    info "4. Проверка базы данных настроек:"
    info "   kubectl exec -n zitadel pg-postgresql-0 -- psql -U postgres -d zitadel -c \"SELECT instance_id, domain FROM projections.instance_domains;\""
    info ""
    info "5. Альтернативные URL для доступа:"
    info "   - Прямой доступ к поду: kubectl port-forward -n zitadel svc/zitadel 8080:8080"
    info "   - Затем откройте: http://localhost:8080"
    info ""
}

# Основная функция
main() {
    log "🔐 ZITADEL API Token - альтернативные методы"
    log "Domain: $ZITADEL_DOMAIN"
    log "Admin: $ADMIN_EMAIL"
    log ""
    
    # Проверяем существующий токен
    if [[ -n "${ZITADEL_TOKEN:-}" ]]; then
        log "✅ Токен уже установлен: ZITADEL_TOKEN"
        return 0
    fi
    
    if [[ -f "$HOME/.zitadel/token" ]]; then
        log "✅ Найден токен в ~/.zitadel/token"
        export ZITADEL_TOKEN=$(cat "$HOME/.zitadel/token" | tr -d '\n\r')
        return 0
    fi
    
    info "Выберите метод получения токена:"
    info "1. Попытка Device Code Flow (может не работать из-за цикла)"
    info "2. Прямой доступ к PostgreSQL (продвинутый)"
    info "3. Восстановление веб-интерфейса"
    info "4. Port-forward для прямого доступа"
    info ""
    
    read -p "Выберите вариант (1-4): " choice
    
    case "$choice" in
        1)
            if client_id=$(discover_system_apps); then
                if token=$(try_device_code_flow "$client_id"); then
                    mkdir -p "$HOME/.zitadel"
                    echo "$token" > "$HOME/.zitadel/token"
                    chmod 600 "$HOME/.zitadel/token"
                    log "✅ Токен сохранен"
                    info "export ZITADEL_TOKEN='$token'"
                fi
            fi
            ;;
        2)
            try_postgres_direct
            ;;
        3)
            fix_web_interface
            ;;
        4)
            log "Запуск port-forward для прямого доступа..."
            info "Выполните в отдельном терминале:"
            info "kubectl port-forward -n zitadel svc/zitadel 8080:8080"
            info ""
            info "Затем откройте: http://localhost:8080"
            info "Используйте credentials:"
            info "Email: $ADMIN_EMAIL"
            info "Password: $ADMIN_PASSWORD"
            ;;
        *)
            error "Неверный выбор"
            return 1
            ;;
    esac
}

# Запуск скрипта
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
