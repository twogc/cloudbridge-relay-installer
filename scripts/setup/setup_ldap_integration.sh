#!/bin/bash

# CloudBridge Relay - LDAP/AD Integration Setup Script
# Скрипт для настройки интеграции Active Directory/LDAP через туннелирование

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функции логирования
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Конфигурация по умолчанию
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
CONFIG_DIR="$PROJECT_ROOT/config"
COMPANY_MAPPING_FILE="$CONFIG_DIR/company_mapping.yaml"

# Переменные окружения
KEYCLOAK_URL="${KEYCLOAK_URL:-http://localhost:8080}"
KEYCLOAK_ADMIN_USER="${KEYCLOAK_ADMIN_USER:-admin}"
KEYCLOAK_ADMIN_PASSWORD="${KEYCLOAK_ADMIN_PASSWORD:-admin123}"
RELAY_URL="${RELAY_URL:-http://localhost:8082}"

# Функция проверки зависимостей
check_dependencies() {
    log_info "Checking dependencies..."
    
    local missing_deps=()
    
    # Проверка jq
    if ! command -v jq &> /dev/null; then
        missing_deps+=("jq")
    fi
    
    # Проверка yq
    if ! command -v yq &> /dev/null; then
        missing_deps+=("yq")
    fi
    
    # Проверка curl
    if ! command -v curl &> /dev/null; then
        missing_deps+=("curl")
    fi
    
    # Проверка ldapsearch (для тестирования LDAP)
    if ! command -v ldapsearch &> /dev/null; then
        log_warning "ldapsearch not found - LDAP testing will be skipped"
    fi
    
    if [ ${#missing_deps[@]} -ne 0 ]; then
        log_error "Missing dependencies: ${missing_deps[*]}"
        log_info "Please install missing dependencies and try again"
        exit 1
    fi
    
    log_success "All dependencies are available"
}

# Функция получения токена администратора Keycloak
get_admin_token() {
    log_info "Getting Keycloak admin token..."
    
    local token_response=$(curl -s -X POST "$KEYCLOAK_URL/realms/master/protocol/openid-connect/token" \
        -H "Content-Type: application/x-www-form-urlencoded" \
        -d "grant_type=password" \
        -d "client_id=admin-cli" \
        -d "username=$KEYCLOAK_ADMIN_USER" \
        -d "password=$KEYCLOAK_ADMIN_PASSWORD")
    
    if [ $? -ne 0 ]; then
        log_error "Failed to get admin token"
        return 1
    fi
    
    local access_token=$(echo "$token_response" | jq -r '.access_token')
    
    if [ "$access_token" = "null" ] || [ -z "$access_token" ]; then
        log_error "Failed to extract access token from response"
        echo "Response: $token_response"
        return 1
    fi
    
    echo "$access_token"
    log_success "Admin token obtained"
}

# Функция создания туннеля LDAP
create_ldap_tunnel() {
    local company_id="$1"
    local ldap_server="$2"
    local ldap_port="${3:-389}"
    local ldaps_port="${4:-636}"
    
    log_info "Creating LDAP tunnel for company: $company_id"
    
    # Создание туннеля через Server Client API
    local tunnel_config=$(cat <<EOF
{
    "tunnel_id": "ldap-$company_id",
    "local_port": $ldap_port,
    "remote_host": "$ldap_server",
    "remote_port": $ldap_port,
    "protocol": "tcp",
    "company": "$company_id",
    "description": "LDAP tunnel for $company_id"
}
EOF
)
    
    local response=$(curl -s -X POST "$RELAY_URL/api/v1/tunnels/ldap" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $ADMIN_TOKEN" \
        -d "$tunnel_config")
    
    if [ $? -ne 0 ]; then
        log_error "Failed to create LDAP tunnel"
        return 1
    fi
    
    local success=$(echo "$response" | jq -r '.success')
    if [ "$success" = "true" ]; then
        log_success "LDAP tunnel created successfully"
        return 0
    else
        log_error "Failed to create LDAP tunnel: $(echo "$response" | jq -r '.error')"
        return 1
    fi
}

# Функция настройки LDAP User Federation в Keycloak
setup_keycloak_ldap() {
    local company_id="$1"
    local realm="$2"
    local ldap_config="$3"
    
    log_info "Setting up LDAP User Federation for realm: $realm"
    
    # Создание LDAP User Federation
    local ldap_federation_config=$(cat <<EOF
{
    "name": "ldap-$company_id",
    "providerId": "ldap",
    "providerType": "org.keycloak.storage.UserStorageProvider",
    "config": {
        "enabled": ["true"],
        "connectionUrl": ["ldap://relay.example.com:389"],
        "bindDn": ["$(echo "$ldap_config" | jq -r '.bind_dn')"],
        "bindCredential": ["$(echo "$ldap_config" | jq -r '.bind_password')"],
        "searchBase": ["$(echo "$ldap_config" | jq -r '.base_dn')"],
        "userDn": ["$(echo "$ldap_config" | jq -r '.user_dn')"],
        "groupDn": ["$(echo "$ldap_config" | jq -r '.group_dn')"],
        "syncRegistrations": ["true"],
        "importUsers": ["true"],
        "periodicFullSync": ["true"],
        "periodicChangedUsersSync": ["true"],
        "fullSyncPeriod": ["1440"],
        "changedSyncPeriod": ["1440"],
        "batchSizeForSync": ["1000"],
        "usernameLDAPAttribute": ["sAMAccountName"],
        "rdnLDAPAttribute": ["cn"],
        "uuidLDAPAttribute": ["objectGUID"],
        "userObjectClasses": ["person, organizationalPerson, user"],
        "connectionPooling": ["true"],
        "connectionPoolingAuthentication": ["SIMPLE"],
        "connectionPoolingDebug": ["false"],
        "connectionPoolingInitSize": ["10"],
        "connectionPoolingMaxSize": ["100"],
        "connectionPoolingPrefSize": ["10"],
        "connectionPoolingProtocol": ["ldap"],
        "connectionPoolingTimeout": ["300000"],
        "connectionTimeout": ["30000"],
        "readTimeout": ["30000"],
        "pagination": ["true"],
        "allowKerberos": ["false"],
        "useKerberosForPasswordAuthentication": ["false"],
        "serverPrincipal": [""],
        "keyTab": [""],
        "kerberosRealm": [""],
        "debug": ["false"],
        "useTruststoreSpi": ["ldapsOnly"],
        "connectionPooling": ["true"],
        "editMode": ["READ_ONLY"],
        "vendor": ["ad"]
    }
}
EOF
)
    
    local response=$(curl -s -X POST "$KEYCLOAK_URL/admin/realms/$realm/components" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $ADMIN_TOKEN" \
        -d "$ldap_federation_config")
    
    if [ $? -ne 0 ]; then
        log_error "Failed to setup LDAP User Federation"
        return 1
    fi
    
    local component_id=$(echo "$response" | jq -r '.id')
    if [ "$component_id" != "null" ] && [ -n "$component_id" ]; then
        log_success "LDAP User Federation created with ID: $component_id"
        return 0
    else
        log_error "Failed to create LDAP User Federation: $(echo "$response" | jq -r '.errorMessage')"
        return 1
    fi
}

# Функция тестирования LDAP подключения
test_ldap_connection() {
    local company_id="$1"
    local ldap_config="$2"
    
    log_info "Testing LDAP connection for company: $company_id"
    
    # Тест через туннель
    local ldap_server="relay.example.com"
    local ldap_port=$(echo "$ldap_config" | jq -r '.ldap_port // 389')
    local bind_dn=$(echo "$ldap_config" | jq -r '.bind_dn')
    local bind_password=$(echo "$ldap_config" | jq -r '.bind_password')
    local base_dn=$(echo "$ldap_config" | jq -r '.base_dn')
    
    # Тест с помощью ldapsearch
    if command -v ldapsearch &> /dev/null; then
        local test_result=$(ldapsearch -H "ldap://$ldap_server:$ldap_port" \
            -D "$bind_dn" \
            -w "$bind_password" \
            -b "$base_dn" \
            -s base \
            "(objectClass=*)")
        
        if [ $? -eq 0 ]; then
            log_success "LDAP connection test passed"
            return 0
        else
            log_error "LDAP connection test failed"
            return 1
        fi
    else
        # Альтернативный тест через curl
        log_warning "ldapsearch not available, using alternative test method"
        
        local test_response=$(curl -s -X POST "$RELAY_URL/api/v1/companies/$company_id/ldap/test-connection" \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $ADMIN_TOKEN" \
            -d "$ldap_config")
        
        local success=$(echo "$test_response" | jq -r '.success')
        if [ "$success" = "true" ]; then
            log_success "LDAP connection test passed"
            return 0
        else
            log_error "LDAP connection test failed: $(echo "$test_response" | jq -r '.error')"
            return 1
        fi
    fi
}

# Функция синхронизации пользователей
sync_users() {
    local company_id="$1"
    local realm="$2"
    
    log_info "Syncing users for company: $company_id"
    
    local response=$(curl -s -X POST "$RELAY_URL/api/v1/companies/$company_id/ldap/sync-users" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $ADMIN_TOKEN")
    
    if [ $? -ne 0 ]; then
        log_error "Failed to sync users"
        return 1
    fi
    
    local success=$(echo "$response" | jq -r '.success')
    if [ "$success" = "true" ]; then
        local user_count=$(echo "$response" | jq -r '.users_synced')
        log_success "Users synced successfully: $user_count users"
        return 0
    else
        log_error "Failed to sync users: $(echo "$response" | jq -r '.error')"
        return 1
    fi
}

# Функция синхронизации групп
sync_groups() {
    local company_id="$1"
    local realm="$2"
    
    log_info "Syncing groups for company: $company_id"
    
    local response=$(curl -s -X POST "$RELAY_URL/api/v1/companies/$company_id/ldap/sync-groups" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $ADMIN_TOKEN")
    
    if [ $? -ne 0 ]; then
        log_error "Failed to sync groups"
        return 1
    fi
    
    local success=$(echo "$response" | jq -r '.success')
    if [ "$success" = "true" ]; then
        local group_count=$(echo "$response" | jq -r '.groups_synced')
        log_success "Groups synced successfully: $group_count groups"
        return 0
    else
        log_error "Failed to sync groups: $(echo "$response" | jq -r '.error')"
        return 1
    fi
}

# Функция настройки интеграции для одной компании
setup_company_integration() {
    local company_id="$1"
    local ldap_config="$2"
    
    log_info "Setting up LDAP integration for company: $company_id"
    
    # Извлечение параметров из конфигурации
    local realm=$(echo "$ldap_config" | jq -r '.realm')
    local ldap_server=$(echo "$ldap_config" | jq -r '.ldap_server')
    local ldap_port=$(echo "$ldap_config" | jq -r '.ldap_port // 389')
    local ldaps_port=$(echo "$ldap_config" | jq -r '.ldaps_port // 636')
    
    # 1. Создание туннеля LDAP
    if ! create_ldap_tunnel "$company_id" "$ldap_server" "$ldap_port" "$ldaps_port"; then
        log_error "Failed to create LDAP tunnel for company: $company_id"
        return 1
    fi
    
    # 2. Тестирование LDAP подключения
    if ! test_ldap_connection "$company_id" "$ldap_config"; then
        log_error "LDAP connection test failed for company: $company_id"
        return 1
    fi
    
    # 3. Настройка Keycloak LDAP User Federation
    if ! setup_keycloak_ldap "$company_id" "$realm" "$ldap_config"; then
        log_error "Failed to setup Keycloak LDAP for company: $company_id"
        return 1
    fi
    
    # 4. Синхронизация пользователей
    if [ "$(echo "$ldap_config" | jq -r '.sync_users // false')" = "true" ]; then
        if ! sync_users "$company_id" "$realm"; then
            log_warning "Failed to sync users for company: $company_id"
        fi
    fi
    
    # 5. Синхронизация групп
    if [ "$(echo "$ldap_config" | jq -r '.sync_groups // false')" = "true" ]; then
        if ! sync_groups "$company_id" "$realm"; then
            log_warning "Failed to sync groups for company: $company_id"
        fi
    fi
    
    log_success "LDAP integration setup completed for company: $company_id"
    return 0
}

# Функция чтения конфигурации компаний
read_company_config() {
    if [ ! -f "$COMPANY_MAPPING_FILE" ]; then
        log_error "Company mapping file not found: $COMPANY_MAPPING_FILE"
        exit 1
    fi
    
    log_info "Reading company configuration from: $COMPANY_MAPPING_FILE"
    
    # Проверка синтаксиса YAML
    if ! yq eval '.' "$COMPANY_MAPPING_FILE" > /dev/null 2>&1; then
        log_error "Invalid YAML syntax in company mapping file"
        exit 1
    fi
    
    # Получение списка компаний
    local companies=$(yq eval '.companies | keys | .[]' "$COMPANY_MAPPING_FILE" 2>/dev/null)
    
    if [ -z "$companies" ]; then
        log_error "No companies found in configuration file"
        exit 1
    fi
    
    echo "$companies"
}

# Функция получения конфигурации LDAP для компании
get_company_ldap_config() {
    local company_id="$1"
    
    local ldap_config=$(yq eval ".companies.$company_id.ldap" "$COMPANY_MAPPING_FILE" 2>/dev/null)
    
    if [ "$ldap_config" = "null" ] || [ -z "$ldap_config" ]; then
        return 1
    fi
    
    echo "$ldap_config"
}

# Функция проверки статуса сервисов
check_services() {
    log_info "Checking service availability..."
    
    # Проверка Keycloak
    if ! curl -s -f "$KEYCLOAK_URL/health" > /dev/null 2>&1; then
        log_error "Keycloak is not available at: $KEYCLOAK_URL"
        return 1
    fi
    
    # Проверка Relay Server
    if ! curl -s -f "$RELAY_URL/health" > /dev/null 2>&1; then
        log_error "Relay Server is not available at: $RELAY_URL"
        return 1
    fi
    
    log_success "All services are available"
    return 0
}

# Функция отображения справки
show_help() {
    cat << EOF
CloudBridge Relay - LDAP/AD Integration Setup Script

Usage: $0 [OPTIONS] [COMMAND]

Commands:
    setup-all     Setup LDAP integration for all companies
    setup         Setup LDAP integration for specific company
    test          Test LDAP connection for specific company
    sync          Sync users/groups for specific company
    status        Show integration status
    help          Show this help message

Options:
    -c, --company COMPANY_ID    Company ID for specific operations
    -f, --config FILE           Company mapping configuration file
    -k, --keycloak-url URL      Keycloak server URL
    -r, --relay-url URL         Relay server URL
    -u, --admin-user USER       Keycloak admin username
    -p, --admin-password PASS   Keycloak admin password
    -h, --help                  Show this help message

Examples:
    $0 setup-all                           # Setup integration for all companies
    $0 setup -c company-a                  # Setup integration for company-a
    $0 test -c company-a                   # Test LDAP connection for company-a
    $0 sync -c company-a                   # Sync users/groups for company-a
    $0 status                              # Show integration status

Environment Variables:
    KEYCLOAK_URL              Keycloak server URL
    KEYCLOAK_ADMIN_USER       Keycloak admin username
    KEYCLOAK_ADMIN_PASSWORD   Keycloak admin password
    RELAY_URL                 Relay server URL

EOF
}

# Основная функция
main() {
    local command=""
    local company_id=""
    
    # Парсинг аргументов
    while [[ $# -gt 0 ]]; do
        case $1 in
            setup-all|setup|test|sync|status|help)
                command="$1"
                shift
                ;;
            -c|--company)
                company_id="$2"
                shift 2
                ;;
            -f|--config)
                COMPANY_MAPPING_FILE="$2"
                shift 2
                ;;
            -k|--keycloak-url)
                KEYCLOAK_URL="$2"
                shift 2
                ;;
            -r|--relay-url)
                RELAY_URL="$2"
                shift 2
                ;;
            -u|--admin-user)
                KEYCLOAK_ADMIN_USER="$2"
                shift 2
                ;;
            -p|--admin-password)
                KEYCLOAK_ADMIN_PASSWORD="$2"
                shift 2
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    # Проверка команды
    if [ -z "$command" ]; then
        log_error "No command specified"
        show_help
        exit 1
    fi
    
    # Выполнение команды
    case "$command" in
        help)
            show_help
            exit 0
            ;;
        setup-all)
            setup_all_companies
            ;;
        setup)
            if [ -z "$company_id" ]; then
                log_error "Company ID is required for setup command"
                exit 1
            fi
            setup_single_company "$company_id"
            ;;
        test)
            if [ -z "$company_id" ]; then
                log_error "Company ID is required for test command"
                exit 1
            fi
            test_company_connection "$company_id"
            ;;
        sync)
            if [ -z "$company_id" ]; then
                log_error "Company ID is required for sync command"
                exit 1
            fi
            sync_company_data "$company_id"
            ;;
        status)
            show_integration_status
            ;;
        *)
            log_error "Unknown command: $command"
            show_help
            exit 1
            ;;
    esac
}

# Функция настройки всех компаний
setup_all_companies() {
    log_info "Setting up LDAP integration for all companies"
    
    # Проверка зависимостей и сервисов
    check_dependencies
    check_services
    
    # Получение токена администратора
    ADMIN_TOKEN=$(get_admin_token)
    if [ $? -ne 0 ]; then
        log_error "Failed to get admin token"
        exit 1
    fi
    
    # Чтение конфигурации компаний
    local companies=$(read_company_config)
    
    local success_count=0
    local total_count=0
    
    for company in $companies; do
        ((total_count++))
        
        log_info "Processing company: $company"
        
        # Получение конфигурации LDAP
        local ldap_config=$(get_company_ldap_config "$company")
        if [ $? -ne 0 ]; then
            log_warning "No LDAP configuration found for company: $company"
            continue
        fi
        
        # Настройка интеграции
        if setup_company_integration "$company" "$ldap_config"; then
            ((success_count++))
        fi
        
        echo ""
    done
    
    # Итоговая статистика
    log_info "Setup completed: $success_count/$total_count companies configured successfully"
    
    if [ $success_count -eq $total_count ]; then
        log_success "All companies configured successfully!"
        exit 0
    else
        log_warning "Some companies failed to configure"
        exit 1
    fi
}

# Функция настройки одной компании
setup_single_company() {
    local company_id="$1"
    
    log_info "Setting up LDAP integration for company: $company_id"
    
    # Проверка зависимостей и сервисов
    check_dependencies
    check_services
    
    # Получение токена администратора
    ADMIN_TOKEN=$(get_admin_token)
    if [ $? -ne 0 ]; then
        log_error "Failed to get admin token"
        exit 1
    fi
    
    # Получение конфигурации LDAP
    local ldap_config=$(get_company_ldap_config "$company_id")
    if [ $? -ne 0 ]; then
        log_error "No LDAP configuration found for company: $company_id"
        exit 1
    fi
    
    # Настройка интеграции
    if setup_company_integration "$company_id" "$ldap_config"; then
        log_success "LDAP integration setup completed for company: $company_id"
        exit 0
    else
        log_error "LDAP integration setup failed for company: $company_id"
        exit 1
    fi
}

# Функция тестирования подключения компании
test_company_connection() {
    local company_id="$1"
    
    log_info "Testing LDAP connection for company: $company_id"
    
    # Получение токена администратора
    ADMIN_TOKEN=$(get_admin_token)
    if [ $? -ne 0 ]; then
        log_error "Failed to get admin token"
        exit 1
    fi
    
    # Получение конфигурации LDAP
    local ldap_config=$(get_company_ldap_config "$company_id")
    if [ $? -ne 0 ]; then
        log_error "No LDAP configuration found for company: $company_id"
        exit 1
    fi
    
    # Тестирование подключения
    if test_ldap_connection "$company_id" "$ldap_config"; then
        log_success "LDAP connection test passed for company: $company_id"
        exit 0
    else
        log_error "LDAP connection test failed for company: $company_id"
        exit 1
    fi
}

# Функция синхронизации данных компании
sync_company_data() {
    local company_id="$1"
    
    log_info "Syncing data for company: $company_id"
    
    # Получение токена администратора
    ADMIN_TOKEN=$(get_admin_token)
    if [ $? -ne 0 ]; then
        log_error "Failed to get admin token"
        exit 1
    fi
    
    # Получение конфигурации LDAP
    local ldap_config=$(get_company_ldap_config "$company_id")
    if [ $? -ne 0 ]; then
        log_error "No LDAP configuration found for company: $company_id"
        exit 1
    fi
    
    local realm=$(echo "$ldap_config" | jq -r '.realm')
    local success=true
    
    # Синхронизация пользователей
    if [ "$(echo "$ldap_config" | jq -r '.sync_users // false')" = "true" ]; then
        if ! sync_users "$company_id" "$realm"; then
            success=false
        fi
    fi
    
    # Синхронизация групп
    if [ "$(echo "$ldap_config" | jq -r '.sync_groups // false')" = "true" ]; then
        if ! sync_groups "$company_id" "$realm"; then
            success=false
        fi
    fi
    
    if [ "$success" = "true" ]; then
        log_success "Data sync completed for company: $company_id"
        exit 0
    else
        log_error "Data sync failed for company: $company_id"
        exit 1
    fi
}

# Функция отображения статуса интеграции
show_integration_status() {
    log_info "Showing LDAP integration status"
    
    # Получение токена администратора
    ADMIN_TOKEN=$(get_admin_token)
    if [ $? -ne 0 ]; then
        log_error "Failed to get admin token"
        exit 1
    fi
    
    # Чтение конфигурации компаний
    local companies=$(read_company_config)
    
    echo ""
    echo "LDAP Integration Status:"
    echo "========================"
    
    for company in $companies; do
        echo ""
        echo "Company: $company"
        
        # Получение конфигурации LDAP
        local ldap_config=$(get_company_ldap_config "$company")
        if [ $? -ne 0 ]; then
            echo "  Status: Not configured"
            continue
        fi
        
        local realm=$(echo "$ldap_config" | jq -r '.realm')
        local ldap_server=$(echo "$ldap_config" | jq -r '.ldap_server')
        
        echo "  Realm: $realm"
        echo "  LDAP Server: $ldap_server"
        
        # Проверка статуса туннеля
        local tunnel_status=$(curl -s -X GET "$RELAY_URL/api/v1/tunnels/ldap-$company/status" \
            -H "Authorization: Bearer $ADMIN_TOKEN" 2>/dev/null)
        
        if [ $? -eq 0 ]; then
            local tunnel_active=$(echo "$tunnel_status" | jq -r '.active')
            if [ "$tunnel_active" = "true" ]; then
                echo "  Tunnel: Active"
            else
                echo "  Tunnel: Inactive"
            fi
        else
            echo "  Tunnel: Unknown"
        fi
        
        # Проверка статуса Keycloak LDAP
        local keycloak_status=$(curl -s -X GET "$KEYCLOAK_URL/admin/realms/$realm/components" \
            -H "Authorization: Bearer $ADMIN_TOKEN" 2>/dev/null)
        
        if [ $? -eq 0 ]; then
            local ldap_components=$(echo "$keycloak_status" | jq -r '.[] | select(.providerId == "ldap") | .name')
            if [ -n "$ldap_components" ]; then
                echo "  Keycloak LDAP: Configured"
            else
                echo "  Keycloak LDAP: Not configured"
            fi
        else
            echo "  Keycloak LDAP: Unknown"
        fi
    done
    
    echo ""
}

# Запуск основной функции
main "$@" 