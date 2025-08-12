#!/bin/bash

# CloudBridge AI Monitoring Agent - Setup Script

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функции для вывода
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

# Проверка зависимостей
check_dependencies() {
    log_info "Проверка системных зависимостей..."
    
    if ! command -v node &> /dev/null; then
        log_error "Node.js не установлен. Установите Node.js 18+"
        exit 1
    fi
    
    if ! command -v npm &> /dev/null; then
        log_error "npm не установлен"
        exit 1
    fi
    
    log_success "Системные зависимости OK"
}

# Установка npm зависимостей
install_npm_dependencies() {
    log_info "Установка npm зависимостей..."
    
    if [ ! -f "package.json" ]; then
        log_error "package.json не найден"
        exit 1
    fi
    
    npm install
    
    log_success "npm зависимости установлены"
}

# Настройка переменных окружения
setup_environment() {
    log_info "Настройка переменных окружения..."
    
    if [ ! -f ".env" ]; then
        if [ -f "env.example" ]; then
            cp env.example .env
            log_success "Создан файл .env из env.example"
            log_warning "Пожалуйста, отредактируйте .env файл с вашими API ключами"
        else
            log_warning "env.example не найден, создайте .env файл вручную"
        fi
    else
        log_info "Файл .env уже существует"
    fi
}

# Проверка TypeScript конфигурации
check_typescript() {
    log_info "Проверка TypeScript конфигурации..."
    
    # Проверка основной конфигурации
    if npm run type-check > /dev/null 2>&1; then
        log_success "Основная TypeScript конфигурация OK"
    else
        log_warning "Проблемы с основной TypeScript конфигурацией"
    fi
    
    # Проверка конфигурации Workers (если установлены типы)
    if npm list @cloudflare/workers-types > /dev/null 2>&1; then
        if npm run type-check:workers > /dev/null 2>&1; then
            log_success "TypeScript конфигурация Workers OK"
        else
            log_warning "Проблемы с TypeScript конфигурацией Workers"
        fi
    else
        log_info "Типы Cloudflare Workers не установлены (опционально)"
    fi
}

# Проверка Wrangler
check_wrangler() {
    log_info "Проверка Wrangler..."
    
    if ! command -v wrangler &> /dev/null; then
        log_warning "Wrangler не установлен. Устанавливаем..."
        npm install -g wrangler
        log_success "Wrangler установлен"
    else
        log_success "Wrangler уже установлен"
    fi
}

# Создание .gitignore
create_gitignore() {
    log_info "Создание .gitignore..."
    
    if [ ! -f ".gitignore" ]; then
        cat > .gitignore << EOF
# Dependencies
node_modules/
npm-debug.log*
yarn-debug.log*
yarn-error.log*

# Environment variables
.env
.env.local
.env.development.local
.env.test.local
.env.production.local

# Build outputs
dist/
build/
.wrangler/

# Logs
logs/
*.log

# Runtime data
pids/
*.pid
*.seed
*.pid.lock

# Coverage directory used by tools like istanbul
coverage/

# nyc test coverage
.nyc_output/

# Dependency directories
jspm_packages/

# Optional npm cache directory
.npm

# Optional REPL history
.node_repl_history

# Output of 'npm pack'
*.tgz

# Yarn Integrity file
.yarn-integrity

# dotenv environment variables file
.env

# IDE files
.vscode/
.idea/
*.swp
*.swo

# OS generated files
.DS_Store
.DS_Store?
._*
.Spotlight-V100
.Trashes
ehthumbs.db
Thumbs.db

# Cloudflare Workers
.wrangler/
wrangler.toml.bak
EOF
        log_success ".gitignore создан"
    else
        log_info ".gitignore уже существует"
    fi
}

# Основная функция
main() {
    log_info "🚀 Настройка CloudBridge AI Monitoring Agent"
    log_info "=========================================="
    
    check_dependencies
    install_npm_dependencies
    setup_environment
    check_wrangler
    check_typescript
    create_gitignore
    
    log_info ""
    log_success "✅ Настройка завершена!"
    log_info ""
    log_info "📋 Следующие шаги:"
    log_info "1. Отредактируйте .env файл с вашими API ключами"
    log_info "2. Запустите агент: npm run dev"
    log_info "3. Протестируйте: npm run test:agent"
    log_info ""
    log_info "📚 Документация: README.md"
    log_info "🔧 Команды: npm run --help"
}

# Запуск
main "$@" 