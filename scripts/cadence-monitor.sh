#!/bin/bash

# CloudBridge Relay - Cadence CLI Monitoring Script
# Скрипт для мониторинга Cadence через CLI

set -e

NAMESPACE="cloudbridge"
CADENCE_POD=""

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Функция для поиска Cadence пода
find_cadence_pod() {
    CADENCE_POD=$(kubectl get pods -n $NAMESPACE -l app.kubernetes.io/name=cadence -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || echo "")
    
    if [ -z "$CADENCE_POD" ]; then
        echo -e "${RED}❌ Cadence pod не найден в namespace $NAMESPACE${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}✅ Найден Cadence pod: $CADENCE_POD${NC}"
}

# Функция для выполнения Cadence CLI команд
run_cadence_cli() {
    local command="$1"
    local description="$2"
    
    echo -e "${BLUE}🔍 $description${NC}"
    echo -e "${CYAN}Команда: $command${NC}"
    echo ""
    
    kubectl exec -n $NAMESPACE $CADENCE_POD -- cadence --address localhost:7933 $command 2>/dev/null || {
        echo -e "${YELLOW}⚠️  Команда не выполнена или нет данных${NC}"
    }
    echo ""
}

# Функция для показа статуса доменов
show_domains() {
    echo -e "${PURPLE}📋 СТАТУС ДОМЕНОВ CADENCE${NC}"
    echo -e "${BLUE}================================${NC}"
    echo ""
    
    # Пробуем получить список доменов через admin команды
    echo -e "${YELLOW}🔍 Попытка получить список доменов...${NC}"
    kubectl exec -n $NAMESPACE $CADENCE_POD -- cadence --address localhost:7933 admin domain list 2>/dev/null || {
        echo -e "${YELLOW}⚠️  Команда admin domain list не поддерживается${NC}"
    }
    echo ""
    
    # Пробуем описать известные домены
    for domain in "cadence-system" "cloudbridge"; do
        echo -e "${YELLOW}🔍 Описание домена: $domain${NC}"
        kubectl exec -n $NAMESPACE $CADENCE_POD -- cadence --address localhost:7933 domain describe --domain $domain 2>/dev/null || {
            echo -e "${YELLOW}⚠️  Домен $domain не найден или недоступен${NC}"
        }
        echo ""
    done
}

# Функция для показа workflow
show_workflows() {
    echo -e "${PURPLE}📋 WORKFLOW EXECUTIONS${NC}"
    echo -e "${BLUE}================================${NC}"
    echo ""
    
    for domain in "cadence-system" "cloudbridge"; do
        echo -e "${YELLOW}🔍 Workflow в домене: $domain${NC}"
        kubectl exec -n $NAMESPACE $CADENCE_POD -- cadence --address localhost:7933 workflow list --domain $domain 2>/dev/null || {
            echo -e "${YELLOW}⚠️  Нет workflow в домене $domain или домен недоступен${NC}"
        }
        echo ""
    done
}

# Функция для показа tasklist
show_tasklists() {
    echo -e "${PURPLE}📋 TASK LISTS${NC}"
    echo -e "${BLUE}================================${NC}"
    echo ""
    
    for domain in "cadence-system" "cloudbridge"; do
        echo -e "${YELLOW}🔍 Task lists в домене: $domain${NC}"
        kubectl exec -n $NAMESPACE $CADENCE_POD -- cadence --address localhost:7933 tasklist list --domain $domain 2>/dev/null || {
            echo -e "${YELLOW}⚠️  Нет task lists в домене $domain или домен недоступен${NC}"
        }
        echo ""
    done
}

# Функция для показа кластера
show_cluster() {
    echo -e "${PURPLE}📋 КЛАСТЕР CADENCE${NC}"
    echo -e "${BLUE}================================${NC}"
    echo ""
    
    echo -e "${YELLOW}🔍 Информация о кластере...${NC}"
    kubectl exec -n $NAMESPACE $CADENCE_POD -- cadence --address localhost:7933 cluster health 2>/dev/null || {
        echo -e "${YELLOW}⚠️  Команда cluster health не поддерживается${NC}"
    }
    echo ""
}

# Функция для показа логов
show_logs() {
    echo -e "${PURPLE}📋 ЛОГИ CADENCE${NC}"
    echo -e "${BLUE}================================${NC}"
    echo ""
    
    echo -e "${YELLOW}🔍 Последние 20 строк логов...${NC}"
    kubectl logs -n $NAMESPACE $CADENCE_POD --tail=20
    echo ""
}

# Функция для показа метрик
show_metrics() {
    echo -e "${PURPLE}📋 МЕТРИКИ CADENCE${NC}"
    echo -e "${BLUE}================================${NC}"
    echo ""
    
    echo -e "${YELLOW}🔍 Проверка метрик через Prometheus...${NC}"
    echo -e "${CYAN}💡 Откройте Prometheus: http://localhost:9090${NC}"
    echo -e "${CYAN}💡 Поиск метрик: cadence_*${NC}"
    echo ""
    
    # Проверяем, доступен ли Prometheus
    if curl -s http://localhost:9090/api/v1/query?query=up > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Prometheus доступен на http://localhost:9090${NC}"
        echo -e "${CYAN}🔍 Попробуйте запросы:${NC}"
        echo -e "   - cadence_requests"
        echo -e "   - cadence_workflow_executions"
        echo -e "   - cadence_errors"
    else
        echo -e "${YELLOW}⚠️  Prometheus недоступен. Запустите: ./scripts/monitoring-ui.sh prometheus${NC}"
    fi
    echo ""
}

# Основная функция
main() {
    echo -e "${BLUE}================================${NC}"
    echo -e "${BLUE}  Cadence CLI Monitoring Tool${NC}"
    echo -e "${BLUE}================================${NC}"
    echo ""
    
    find_cadence_pod
    
    case "${1:-all}" in
        "domains"|"d")
            show_domains
            ;;
        "workflows"|"wf")
            show_workflows
            ;;
        "tasklists"|"tl")
            show_tasklists
            ;;
        "cluster"|"cl")
            show_cluster
            ;;
        "logs"|"l")
            show_logs
            ;;
        "metrics"|"m")
            show_metrics
            ;;
        "all"|"")
            show_domains
            show_workflows
            show_tasklists
            show_cluster
            show_logs
            show_metrics
            ;;
        "help"|"-h"|"--help")
            echo -e "${BLUE}Использование: $0 [команда]${NC}"
            echo ""
            echo -e "${CYAN}Команды:${NC}"
            echo -e "  ${GREEN}all${NC}        - Показать всю информацию (по умолчанию)"
            echo -e "  ${GREEN}domains${NC}   - Показать статус доменов"
            echo -e "  ${GREEN}workflows${NC} - Показать workflow executions"
            echo -e "  ${GREEN}tasklists${NC} - Показать task lists"
            echo -e "  ${GREEN}cluster${NC}   - Показать информацию о кластере"
            echo -e "  ${GREEN}logs${NC}      - Показать логи"
            echo -e "  ${GREEN}metrics${NC}   - Показать информацию о метриках"
            echo -e "  ${GREEN}help${NC}      - Показать эту справку"
            echo ""
            echo -e "${CYAN}Примеры:${NC}"
            echo -e "  $0                    # Показать всю информацию"
            echo -e "  $0 domains           # Только домены"
            echo -e "  $0 workflows         # Только workflow"
            echo -e "  $0 logs              # Только логи"
            ;;
        *)
            echo -e "${RED}❌ Неизвестная команда: $1${NC}"
            echo -e "${YELLOW}💡 Используйте '$0 help' для справки${NC}"
            exit 1
            ;;
    esac
    
    echo -e "${GREEN}🎉 Мониторинг Cadence завершен!${NC}"
    echo ""
    echo -e "${CYAN}💡 Дополнительные ресурсы:${NC}"
    echo -e "   📊 Grafana: http://localhost:3000"
    echo -e "   📈 Prometheus: http://localhost:9090"
    echo -e "   📋 Логи: kubectl logs -n $NAMESPACE $CADENCE_POD"
}

# Запуск основной функции
main "$@"
