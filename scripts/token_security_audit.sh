#!/bin/bash

# 🔒 Token Security Audit Script
# Проверка безопасности токена аутентификации с внешней стороны
# Автор: CloudBridge Security Team
# Версия: 1.0

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Функции логирования
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_security() { echo -e "${PURPLE}[SECURITY]${NC} $1"; }

# Переменные
TOKEN=""
DOMAIN=""
OUTPUT_DIR="./security_audit_$(date +%Y%m%d_%H%M%S)"
REPORT_FILE="$OUTPUT_DIR/security_report.md"

# Создание директории для результатов
mkdir -p "$OUTPUT_DIR"

# Функция помощи
show_help() {
    cat << EOF
🔒 Token Security Audit Script

Использование:
    $0 -t <token> -d <domain> [опции]

Обязательные параметры:
    -t, --token      Токен для анализа
    -d, --domain     Домен для проверки

Опции:
    -o, --output     Директория для результатов (по умолчанию: ./security_audit_YYYYMMDD_HHMMSS)
    -v, --verbose    Подробный вывод
    -h, --help       Показать эту справку

Примеры:
    $0 -t "user.2045597475330327839.60526-0d7eb462c42a221300af346a7a.1efa83c73e0ac992c0cf22b196cf6541bd268d402051e1e2" -d "app60526.pushgrow.io"
    $0 -t "your_token_here" -d "example.com" -v

EOF
}

# Парсинг аргументов
VERBOSE=false
while [[ $# -gt 0 ]]; do
    case $1 in
        -t|--token)
            TOKEN="$2"
            shift 2
            ;;
        -d|--domain)
            DOMAIN="$2"
            shift 2
            ;;
        -o|--output)
            OUTPUT_DIR="$2"
            shift 2
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            log_error "Неизвестный параметр: $1"
            show_help
            exit 1
            ;;
    esac
done

# Проверка обязательных параметров
if [[ -z "$TOKEN" || -z "$DOMAIN" ]]; then
    log_error "Токен и домен обязательны для анализа"
    show_help
    exit 1
fi

# Функция для проверки зависимостей
check_dependencies() {
    log_info "Проверка зависимостей..."
    
    local deps=("curl" "openssl" "nmap" "dig" "whois" "jq")
    local missing_deps=()
    
    for dep in "${deps[@]}"; do
        if ! command -v "$dep" &> /dev/null; then
            missing_deps+=("$dep")
        fi
    done
    
    if [[ ${#missing_deps[@]} -gt 0 ]]; then
        log_warning "Отсутствуют зависимости: ${missing_deps[*]}"
        log_info "Установите их для полного анализа:"
        for dep in "${missing_deps[@]}"; do
            case $dep in
                "jq") echo "  sudo apt-get install jq" ;;
                "nmap") echo "  sudo apt-get install nmap" ;;
                "dig") echo "  sudo apt-get install dnsutils" ;;
                "whois") echo "  sudo apt-get install whois" ;;
            esac
        done
    else
        log_success "Все зависимости установлены"
    fi
}

# Функция анализа структуры токена
analyze_token_structure() {
    log_info "Анализ структуры токена..."
    
    local token="$1"
    local parts=(${token//./ })
    
    cat > "$OUTPUT_DIR/token_analysis.txt" << EOF
# Анализ структуры токена
Токен: $token
Длина: ${#token} символов
Количество частей: ${#parts[@]}

## Разбор частей:
EOF
    
    for i in "${!parts[@]}"; do
        local part="${parts[$i]}"
        local length=${#part}
        local entropy=$(echo "$part" | tr -d '[:space:]' | wc -c)
        
        echo "Часть $((i+1)): $part (длина: $length, энтропия: $entropy)" >> "$OUTPUT_DIR/token_analysis.txt"
        
        if [[ $VERBOSE == true ]]; then
            log_info "Часть $((i+1)): $part (длина: $length)"
        fi
    done
    
    # Анализ подписи
    local signature="${parts[-1]}"
    if [[ ${#signature} -eq 64 ]]; then
        log_success "Подпись: SHA-256 (64 символа)"
        echo "Подпись: SHA-256 (64 символа)" >> "$OUTPUT_DIR/token_analysis.txt"
    elif [[ ${#signature} -eq 40 ]]; then
        log_warning "Подпись: SHA-1 (40 символов) - устаревший алгоритм"
        echo "Подпись: SHA-1 (40 символов) - устаревший алгоритм" >> "$OUTPUT_DIR/token_analysis.txt"
    else
        log_warning "Неизвестный формат подписи (${#signature} символов)"
        echo "Неизвестный формат подписи (${#signature} символов)" >> "$OUTPUT_DIR/token_analysis.txt"
    fi
    
    log_success "Анализ структуры завершен"
}

# Функция проверки SSL/TLS
check_ssl_security() {
    log_info "Проверка SSL/TLS безопасности для $DOMAIN..."
    
    local ssl_file="$OUTPUT_DIR/ssl_analysis.txt"
    
    # Проверка SSL сертификата
    echo "# SSL/TLS анализ для $DOMAIN" > "$ssl_file"
    echo "Дата проверки: $(date)" >> "$ssl_file"
    echo "" >> "$ssl_file"
    
    # Получение сертификата
    if openssl s_client -connect "$DOMAIN:443" -servername "$DOMAIN" < /dev/null 2>/dev/null | openssl x509 -text -noout > "$OUTPUT_DIR/certificate.txt" 2>/dev/null; then
        log_success "SSL сертификат получен"
        
        # Проверка срока действия
        local expiry=$(openssl s_client -connect "$DOMAIN:443" -servername "$DOMAIN" < /dev/null 2>/dev/null | openssl x509 -noout -enddate 2>/dev/null | cut -d= -f2)
        local expiry_epoch=$(date -d "$expiry" +%s 2>/dev/null || echo "0")
        local now_epoch=$(date +%s)
        local days_left=$(( (expiry_epoch - now_epoch) / 86400 ))
        
        echo "Срок действия сертификата: $expiry" >> "$ssl_file"
        echo "Дней до истечения: $days_left" >> "$ssl_file"
        
        if [[ $days_left -lt 30 ]]; then
            log_warning "Сертификат истекает через $days_left дней"
        else
            log_success "Сертификат действителен еще $days_left дней"
        fi
        
        # Проверка протоколов
        echo "" >> "$ssl_file"
        echo "## Поддерживаемые протоколы:" >> "$ssl_file"
        
        local protocols=("ssl2" "ssl3" "tls1" "tls1_1" "tls1_2" "tls1_3")
        for protocol in "${protocols[@]}"; do
            if timeout 5 openssl s_client -connect "$DOMAIN:443" -servername "$DOMAIN" -"$protocol" < /dev/null 2>/dev/null | grep -q "Connected"; then
                echo "$protocol: поддерживается" >> "$ssl_file"
                if [[ "$protocol" == "ssl2" || "$protocol" == "ssl3" ]]; then
                    log_warning "Устаревший протокол $protocol поддерживается"
                fi
            else
                echo "❌ $protocol: не поддерживается" >> "$ssl_file"
            fi
        done
        
    else
        log_error "Не удалось получить SSL сертификат"
        echo "Ошибка: Не удалось получить SSL сертификат" >> "$ssl_file"
    fi
}

# Функция проверки заголовков безопасности
check_security_headers() {
    log_info "Проверка заголовков безопасности..."
    
    local headers_file="$OUTPUT_DIR/security_headers.txt"
    
    echo "# Анализ заголовков безопасности для $DOMAIN" > "$headers_file"
    echo "Дата проверки: $(date)" >> "$headers_file"
    echo "" >> "$headers_file"
    
    # Получение заголовков
    local headers=$(curl -s -I "https://$DOMAIN" 2>/dev/null || curl -s -I "http://$DOMAIN" 2>/dev/null)
    
    if [[ -n "$headers" ]]; then
        echo "## Полученные заголовки:" >> "$headers_file"
        echo "$headers" >> "$headers_file"
        echo "" >> "$headers_file"
        
        # Проверка важных заголовков безопасности
        local security_headers=(
            "Strict-Transport-Security"
            "Content-Security-Policy"
            "X-Frame-Options"
            "X-Content-Type-Options"
            "X-XSS-Protection"
            "Referrer-Policy"
            "Permissions-Policy"
        )
        
        echo "## Анализ заголовков безопасности:" >> "$headers_file"
        
        for header in "${security_headers[@]}"; do
            if echo "$headers" | grep -qi "$header"; then
                local value=$(echo "$headers" | grep -i "$header" | head -1)
                echo "$value" >> "$headers_file"
                log_success "Найден заголовок: $header"
            else
                echo "❌ $header: отсутствует" >> "$headers_file"
                log_warning "Отсутствует заголовок безопасности: $header"
            fi
        done
        
    else
        log_error "Не удалось получить заголовки"
        echo "Ошибка: Не удалось получить заголовки" >> "$headers_file"
    fi
}

# Функция проверки DNS
check_dns_security() {
    log_info "Проверка DNS безопасности..."
    
    local dns_file="$OUTPUT_DIR/dns_analysis.txt"
    
    echo "# DNS анализ для $DOMAIN" > "$dns_file"
    echo "Дата проверки: $(date)" >> "$dns_file"
    echo "" >> "$dns_file"
    
    # Проверка A записи
    local a_record=$(dig +short A "$DOMAIN" 2>/dev/null | head -1)
    if [[ -n "$a_record" ]]; then
        echo "A запись: $a_record" >> "$dns_file"
        log_info "IP адрес: $a_record"
        
        # Проверка обратного DNS
        local reverse_dns=$(dig +short -x "$a_record" 2>/dev/null)
        if [[ -n "$reverse_dns" ]]; then
            echo "Обратный DNS: $reverse_dns" >> "$dns_file"
        fi
    fi
    
    # Проверка MX записей
    local mx_records=$(dig +short MX "$DOMAIN" 2>/dev/null)
    if [[ -n "$mx_records" ]]; then
        echo "MX записи:" >> "$dns_file"
        echo "$mx_records" >> "$dns_file"
    fi
    
    # Проверка TXT записей (SPF, DMARC)
    local txt_records=$(dig +short TXT "$DOMAIN" 2>/dev/null)
    if [[ -n "$txt_records" ]]; then
        echo "TXT записи:" >> "$dns_file"
        echo "$txt_records" >> "$dns_file"
        
        # Проверка SPF
        if echo "$txt_records" | grep -q "v=spf1"; then
            log_success "SPF запись найдена"
        else
            log_warning "SPF запись отсутствует"
        fi
        
        # Проверка DMARC
        local dmarc_record=$(dig +short TXT "_dmarc.$DOMAIN" 2>/dev/null)
        if [[ -n "$dmarc_record" ]]; then
            echo "DMARC запись: $dmarc_record" >> "$dns_file"
            log_success "DMARC запись найдена"
        else
            log_warning "DMARC запись отсутствует"
        fi
    fi
}

# Функция проверки открытых портов
check_open_ports() {
    log_info "Проверка открытых портов..."
    
    local ports_file="$OUTPUT_DIR/ports_analysis.txt"
    
    echo "# Анализ открытых портов для $DOMAIN" > "$ports_file"
    echo "Дата проверки: $(date)" >> "$ports_file"
    echo "" >> "$ports_file"
    
    # Получение IP адреса
    local ip=$(dig +short A "$DOMAIN" 2>/dev/null | head -1)
    
    if [[ -n "$ip" ]]; then
        echo "IP адрес: $ip" >> "$ports_file"
        echo "" >> "$ports_file"
        
        # Сканирование основных портов
        local common_ports=(21 22 23 25 53 80 110 143 443 993 995 1433 1521 3306 3389 5432 5900 6379 8080 8443)
        
        echo "## Сканирование портов:" >> "$ports_file"
        
        for port in "${common_ports[@]}"; do
            if timeout 2 bash -c "</dev/tcp/$ip/$port" 2>/dev/null; then
                echo "Порт $port: открыт" >> "$ports_file"
                log_warning "Открыт порт $port"
            else
                if [[ $VERBOSE == true ]]; then
                    echo "❌ Порт $port: закрыт" >> "$ports_file"
                fi
            fi
        done
        
    else
        log_error "Не удалось получить IP адрес"
        echo "Ошибка: Не удалось получить IP адрес" >> "$ports_file"
    fi
}

# Функция проверки уязвимостей
check_vulnerabilities() {
    log_info "Проверка известных уязвимостей..."
    
    local vuln_file="$OUTPUT_DIR/vulnerabilities.txt"
    
    echo "# Анализ уязвимостей для $DOMAIN" > "$vuln_file"
    echo "Дата проверки: $(date)" >> "$vuln_file"
    echo "" >> "$vuln_file"
    
    # Проверка Heartbleed
    echo "## Проверка Heartbleed:" >> "$vuln_file"
    if timeout 5 openssl s_client -connect "$DOMAIN:443" -servername "$DOMAIN" -tlsextdebug 2>/dev/null | grep -q "heartbeat"; then
        echo "  Heartbeat extension обнаружен" >> "$vuln_file"
        log_warning "Heartbeat extension обнаружен (проверьте версию OpenSSL)"
    else
        echo "Heartbeat extension не обнаружен" >> "$vuln_file"
    fi
    
    # Проверка BEAST
    echo "" >> "$vuln_file"
    echo "## Проверка BEAST:" >> "$vuln_file"
    if timeout 5 openssl s_client -connect "$DOMAIN:443" -servername "$DOMAIN" -cipher 'RC4' < /dev/null 2>/dev/null | grep -q "Connected"; then
        echo "  RC4 поддерживается (уязвимость BEAST)" >> "$vuln_file"
        log_warning "RC4 поддерживается (уязвимость BEAST)"
    else
        echo "RC4 не поддерживается" >> "$vuln_file"
    fi
    
    # Проверка POODLE
    echo "" >> "$vuln_file"
    echo "## Проверка POODLE:" >> "$vuln_file"
    if timeout 5 openssl s_client -connect "$DOMAIN:443" -servername "$DOMAIN" -ssl3 < /dev/null 2>/dev/null | grep -q "Connected"; then
        echo "  SSL 3.0 поддерживается (уязвимость POODLE)" >> "$vuln_file"
        log_warning "SSL 3.0 поддерживается (уязвимость POODLE)"
    else
        echo "SSL 3.0 не поддерживается" >> "$vuln_file"
    fi
}

# Функция анализа токена на предмет утечки информации
analyze_token_information_leak() {
    log_info "Анализ токена на предмет утечки информации..."
    
    local leak_file="$OUTPUT_DIR/information_leak_analysis.txt"
    
    echo "# Анализ утечки информации в токене" > "$leak_file"
    echo "Дата проверки: $(date)" >> "$leak_file"
    echo "" >> "$leak_file"
    
    local token="$1"
    local parts=(${token//./ })
    
    echo "## Анализ частей токена:" >> "$leak_file"
    
    # Анализ User ID
    if [[ ${#parts[1]} -gt 0 ]]; then
        local user_id="${parts[1]}"
        echo "User ID: $user_id" >> "$leak_file"
        
        # Проверка на последовательность
        if [[ $user_id =~ ^[0-9]+$ ]]; then
            echo "  User ID содержит только цифры (может быть последовательным)" >> "$leak_file"
            log_warning "User ID может быть последовательным: $user_id"
        fi
        
        # Проверка длины
        if [[ ${#user_id} -lt 8 ]]; then
            echo "  User ID слишком короткий (${#user_id} символов)" >> "$leak_file"
            log_warning "User ID слишком короткий: $user_id"
        fi
    fi
    
    # Анализ App ID
    if [[ ${#parts[2]} -gt 0 ]]; then
        local app_part="${parts[2]}"
        local app_id=$(echo "$app_part" | cut -d'-' -f1)
        local session_id=$(echo "$app_part" | cut -d'-' -f2)
        
        echo "App ID: $app_id" >> "$leak_file"
        echo "Session ID: $session_id" >> "$leak_file"
        
        # Проверка App ID
        if [[ $app_id =~ ^[0-9]+$ ]]; then
            echo "  App ID содержит только цифры" >> "$leak_file"
        fi
        
        # Проверка Session ID
        if [[ ${#session_id} -lt 16 ]]; then
            echo "  Session ID слишком короткий (${#session_id} символов)" >> "$leak_file"
            log_warning "Session ID слишком короткий: $session_id"
        fi
    fi
    
    # Анализ подписи
    if [[ ${#parts[-1]} -gt 0 ]]; then
        local signature="${parts[-1]}"
        echo "Подпись: $signature" >> "$leak_file"
        
        # Проверка энтропии подписи
        local unique_chars=$(echo "$signature" | fold -w1 | sort | uniq | wc -l)
        echo "Уникальных символов в подписи: $unique_chars" >> "$leak_file"
        
        if [[ $unique_chars -lt 10 ]]; then
            echo "  Низкая энтропия в подписи" >> "$leak_file"
            log_warning "Низкая энтропия в подписи"
        fi
    fi
}

# Функция генерации отчета
generate_report() {
    log_info "Генерация отчета безопасности..."
    
    cat > "$REPORT_FILE" << EOF
# 🔒 Отчет по безопасности токена

## Общая информация
- **Домен**: $DOMAIN
- **Дата проверки**: $(date)
- **Версия скрипта**: 1.0

## Резюме
$(cat "$OUTPUT_DIR/token_analysis.txt" 2>/dev/null || echo "Анализ токена не выполнен")

## SSL/TLS безопасность
$(cat "$OUTPUT_DIR/ssl_analysis.txt" 2>/dev/null || echo "SSL анализ не выполнен")

## Заголовки безопасности
$(cat "$OUTPUT_DIR/security_headers.txt" 2>/dev/null || echo "Анализ заголовков не выполнен")

## DNS безопасность
$(cat "$OUTPUT_DIR/dns_analysis.txt" 2>/dev/null || echo "DNS анализ не выполнен")

## Открытые порты
$(cat "$OUTPUT_DIR/ports_analysis.txt" 2>/dev/null || echo "Анализ портов не выполнен")

## Уязвимости
$(cat "$OUTPUT_DIR/vulnerabilities.txt" 2>/dev/null || echo "Анализ уязвимостей не выполнен")

## Утечка информации
$(cat "$OUTPUT_DIR/information_leak_analysis.txt" 2>/dev/null || echo "Анализ утечки информации не выполнен")

## Рекомендации

### Критические
1. Убедитесь, что все заголовки безопасности настроены
2. Отключите устаревшие протоколы (SSL 2.0, SSL 3.0)
3. Используйте только современные шифры

### Важные
1. Настройте SPF, DKIM и DMARC записи
2. Регулярно обновляйте SSL сертификаты
3. Мониторьте открытые порты

### Рекомендуемые
1. Используйте HSTS заголовок
2. Настройте Content Security Policy
3. Регулярно проводите аудит безопасности

## Файлы отчета
- Полный отчет: $REPORT_FILE
- Анализ токена: $OUTPUT_DIR/token_analysis.txt
- SSL анализ: $OUTPUT_DIR/ssl_analysis.txt
- Заголовки безопасности: $OUTPUT_DIR/security_headers.txt
- DNS анализ: $OUTPUT_DIR/dns_analysis.txt
- Анализ портов: $OUTPUT_DIR/ports_analysis.txt
- Анализ уязвимостей: $OUTPUT_DIR/vulnerabilities.txt
- Анализ утечки информации: $OUTPUT_DIR/information_leak_analysis.txt

EOF
    
    log_success "Отчет сгенерирован: $REPORT_FILE"
}

# Функция очистки
cleanup() {
    if [[ $VERBOSE == true ]]; then
        log_info "Очистка временных файлов..."
    fi
    
    # Удаление временных файлов сертификатов
    rm -f "$OUTPUT_DIR/certificate.txt" 2>/dev/null || true
}

# Основная функция
main() {
    log_security "🔒 Запуск аудита безопасности токена"
    log_info "Домен: $DOMAIN"
    log_info "Директория результатов: $OUTPUT_DIR"
    
    # Проверка зависимостей
    check_dependencies
    
    # Создание директории
    mkdir -p "$OUTPUT_DIR"
    
    # Выполнение проверок
    analyze_token_structure "$TOKEN"
    check_ssl_security
    check_security_headers
    check_dns_security
    check_open_ports
    check_vulnerabilities
    analyze_token_information_leak "$TOKEN"
    
    # Генерация отчета
    generate_report
    
    # Очистка
    cleanup
    
    log_security "Аудит безопасности завершен"
    log_info "Результаты сохранены в: $OUTPUT_DIR"
    log_info "Основной отчет: $REPORT_FILE"
    
    # Вывод краткого резюме
    echo ""
    log_security "Краткое резюме:"
    echo "  - Анализ структуры токена: $(grep -c "SUCCESS\|WARNING\|ERROR" "$OUTPUT_DIR/token_analysis.txt" 2>/dev/null || echo "0") проблем"
    echo "  - SSL/TLS безопасность: $(grep -c "WARNING\|ERROR" "$OUTPUT_DIR/ssl_analysis.txt" 2>/dev/null || echo "0") проблем"
    echo "  - Заголовки безопасности: $(grep -c "отсутствует" "$OUTPUT_DIR/security_headers.txt" 2>/dev/null || echo "0") отсутствующих"
    echo "  - Открытые порты: $(grep -c "открыт" "$OUTPUT_DIR/ports_analysis.txt" 2>/dev/null || echo "0")"
    echo "  - Уязвимости: $(grep -c "уязвимость\|устаревший" "$OUTPUT_DIR/vulnerabilities.txt" 2>/dev/null || echo "0")"
}

# Обработка сигналов
trap cleanup EXIT

# Запуск основной функции
main "$@"
