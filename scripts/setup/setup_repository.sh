#!/bin/bash

# CloudBridge Relay - Repository Setup Script
# Инициализация Git репозитория и загрузка модулей
# Автор: CloudBridge Team
# Версия: 1.0

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

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

# Конфигурация
REPO_URL="git@github.com:twogc/-cloudbridge-install-modules.git"
LOCAL_DIR="cloudbridge-install-modules"

log_info "Настройка репозитория модулей CloudBridge Relay"

# Проверка наличия Git
if ! command -v git &> /dev/null; then
    log_error "Git не установлен. Установите его: sudo apt install git"
    exit 1
fi

# Проверка SSH ключей
log_info "Проверка SSH ключей для GitHub..."
if ! ssh -T git@github.com 2>&1 | grep -q "successfully authenticated"; then
    log_warning "SSH ключ не настроен для GitHub"
    log_info "Настройте SSH ключ: https://docs.github.com/en/authentication/connecting-to-github-with-ssh"
    log_info "Или используйте HTTPS репозиторий"
    read -p "Продолжить с HTTPS? (y/n): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        REPO_URL="https://github.com/twogc/-cloudbridge-install-modules.git"
    else
        exit 1
    fi
fi

# Создание локальной директории
log_info "Создание локальной директории..."
if [[ -d "$LOCAL_DIR" ]]; then
    log_warning "Директория $LOCAL_DIR уже существует"
    read -p "Удалить и пересоздать? (y/n): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        rm -rf "$LOCAL_DIR"
    else
        log_info "Используем существующую директорию"
    fi
fi

if [[ ! -d "$LOCAL_DIR" ]]; then
    mkdir -p "$LOCAL_DIR"
fi

cd "$LOCAL_DIR"

# Инициализация Git репозитория
log_info "Инициализация Git репозитория..."
git init

# Добавление remote origin
log_info "Добавление remote origin..."
git remote add origin "$REPO_URL"

# Создание структуры директорий
log_info "Создание структуры директорий..."
mkdir -p templates
mkdir -p scripts
mkdir -p docs

# Копирование файлов модулей
log_info "Копирование файлов модулей..."

# Копируем модули из текущей директории
cp ../01-system-update.sh ./
cp ../02-security-basics.sh ./
cp ../03-docker-install.sh ./
cp ../modules.json ./

# Создание остальных модулей (заглушки)
log_info "Создание остальных модулей..."

# Модуль 04: Firewall Setup
cat > 04-firewall-setup.sh << 'EOF'
#!/bin/bash

# CloudBridge Relay - Module 04: Firewall Setup
# Настройка UFW файрвола
# Автор: CloudBridge Team
# Версия: 1.0

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Функции логирования
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Проверка root прав
if [[ $EUID -ne 0 ]]; then
    log_error "Этот модуль должен быть запущен с правами root"
    exit 1
fi

log_info "Модуль 04: Настройка файрвола"

# Проверка UFW
if ! command -v ufw &> /dev/null; then
    log_error "UFW не установлен. Установите его: sudo apt install ufw"
    exit 1
fi

# Сброс правил
log_info "Сброс правил файрвола..."
ufw --force reset

# Базовые политики
ufw default deny incoming
ufw default allow outgoing

# SSH доступ
log_info "Настройка SSH доступа..."
ufw allow 22/tcp comment "SSH"
ufw limit 22/tcp comment "SSH rate limiting"

# Relay сервисы
log_info "Настройка портов для Relay сервисов..."
ufw allow 8080/tcp comment "Relay Server API"
ufw allow 8081/tcp comment "Relay Server Metrics"
ufw allow 8082/tcp comment "Relay Server API 2"
ufw allow 8083/tcp comment "Relay API (external)"
ufw allow 3456/tcp comment "Relay Server Custom"

# WireGuard
ufw allow 51820/udp comment "WireGuard P2P Mesh"

# Включение файрвола
log_info "Включение файрвола..."
ufw --force enable

log_success "Модуль 04 завершен успешно"
log_info "Файрвол настроен и включен"
EOF

# Модуль 05: Keycloak Setup
cat > 05-keycloak-setup.sh << 'EOF'
#!/bin/bash

# CloudBridge Relay - Module 05: Keycloak Setup
# Установка и настройка Keycloak
# Автор: CloudBridge Team
# Версия: 1.0

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Функции логирования
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Проверка root прав
if [[ $EUID -ne 0 ]]; then
    log_error "Этот модуль должен быть запущен с правами root"
    exit 1
fi

log_info "Модуль 05: Установка Keycloak"

# Проверка Docker
if ! command -v docker &> /dev/null; then
    log_error "Docker не установлен. Запустите модуль 03 сначала"
    exit 1
fi

# Создание директорий
log_info "Создание директорий для Keycloak..."
mkdir -p /opt/cloudbridge/keycloak/data
mkdir -p /opt/cloudbridge/keycloak/imports
mkdir -p /opt/cloudbridge/keycloak/logs

# Создание docker-compose.yml для Keycloak
log_info "Создание docker-compose.yml для Keycloak..."
cat > /opt/cloudbridge/keycloak/docker-compose.yml << 'EOF'
version: '3.8'

services:
  keycloak:
    image: quay.io/keycloak/keycloak:latest
    container_name: cloudbridge-keycloak
    environment:
      KEYCLOAK_ADMIN: admin
      KEYCLOAK_ADMIN_PASSWORD: admin123
      KC_DB: postgres
      KC_DB_URL: jdbc:postgresql://postgres:5432/keycloak
      KC_DB_USERNAME: keycloak
      KC_DB_PASSWORD: keycloak123
    ports:
      - "8080:8080"
    depends_on:
      - postgres
    volumes:
      - ./data:/opt/keycloak/data
      - ./imports:/opt/keycloak/imports
    command: start-dev
    restart: unless-stopped

  postgres:
    image: postgres:15
    container_name: cloudbridge-postgres
    environment:
      POSTGRES_DB: keycloak
      POSTGRES_USER: keycloak
      POSTGRES_PASSWORD: keycloak123
    volumes:
      - postgres_data:/var/lib/postgresql/data
    restart: unless-stopped

volumes:
  postgres_data:
EOF

# Запуск Keycloak
log_info "Запуск Keycloak..."
cd /opt/cloudbridge/keycloak
docker-compose up -d

# Ожидание готовности
log_info "Ожидание готовности Keycloak..."
sleep 30

log_success "Модуль 05 завершен успешно"
log_info "Keycloak установлен и запущен на http://localhost:8080"
EOF

# Модуль 06: Lynis Setup
cat > 06-lynis-setup.sh << 'EOF'
#!/bin/bash

# CloudBridge Relay - Module 06: Lynis Setup
# Установка и настройка Lynis для аудита безопасности
# Автор: CloudBridge Team
# Версия: 1.0

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Функции логирования
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Проверка root прав
if [[ $EUID -ne 0 ]]; then
    log_error "Этот модуль должен быть запущен с правами root"
    exit 1
fi

log_info "Модуль 06: Установка Lynis"

# Установка Lynis
log_info "Установка Lynis..."
cd /tmp
rm -rf lynis
git clone https://github.com/CISOfy/lynis.git
cd lynis
mkdir -p /usr/local/lynis
cp -r * /usr/local/lynis/
ln -sf /usr/local/lynis/lynis /usr/local/bin/lynis

# Создание конфигурации
log_info "Создание конфигурации Lynis..."
mkdir -p /etc/lynis

cat > /etc/lynis/lynis.conf << 'EOF'
# CloudBridge Relay - Lynis Configuration
PROFILE=relay-server
SKIP_TESTS=test_ssl_certificate_expiry
SKIP_TESTS=test_ssl_certificate_validity

# Report settings
REPORT_FILE=/var/log/lynis-report.dat
REPORT_HTML=/var/log/lynis-report.html
REPORT_CSV=/var/log/lynis-report.csv

# Logging
LOG_FILE=/var/log/lynis.log
LOG_LEVEL=1

# Colors
COLORED_OUTPUT=1

# Privacy
PRIVACY=1
EOF

# Первоначальный аудит
log_info "Выполнение первоначального аудита..."
lynis audit system --quick

log_success "Модуль 06 завершен успешно"
log_info "Lynis установлен и настроен"
EOF

# Модуль 07: Monitoring Setup
cat > 07-monitoring-setup.sh << 'EOF'
#!/bin/bash

# CloudBridge Relay - Module 07: Monitoring Setup
# Установка системы мониторинга
# Автор: CloudBridge Team
# Версия: 1.0

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Функции логирования
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Проверка root прав
if [[ $EUID -ne 0 ]]; then
    log_error "Этот модуль должен быть запущен с правами root"
    exit 1
fi

log_info "Модуль 07: Установка системы мониторинга"

# Проверка Docker
if ! command -v docker &> /dev/null; then
    log_error "Docker не установлен. Запустите модуль 03 сначала"
    exit 1
fi

# Создание директорий
log_info "Создание директорий для мониторинга..."
mkdir -p /opt/cloudbridge/monitoring/prometheus
mkdir -p /opt/cloudbridge/monitoring/grafana
mkdir -p /opt/cloudbridge/monitoring/alertmanager

# Создание docker-compose.yml для мониторинга
log_info "Создание docker-compose.yml для мониторинга..."
cat > /opt/cloudbridge/monitoring/docker-compose.yml << 'EOF'
version: '3.8'

services:
  prometheus:
    image: prom/prometheus:latest
    container_name: cloudbridge-prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus:/etc/prometheus
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/etc/prometheus/console_libraries'
      - '--web.console.templates=/etc/prometheus/consoles'
      - '--storage.tsdb.retention.time=200h'
      - '--web.enable-lifecycle'
    restart: unless-stopped

  grafana:
    image: grafana/grafana:latest
    container_name: cloudbridge-grafana
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin123
    volumes:
      - grafana_data:/var/lib/grafana
    restart: unless-stopped

volumes:
  prometheus_data:
  grafana_data:
EOF

# Создание конфигурации Prometheus
log_info "Создание конфигурации Prometheus..."
cat > /opt/cloudbridge/monitoring/prometheus/prometheus.yml << 'EOF'
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  # - "first_rules.yml"
  # - "second_rules.yml"

scrape_configs:
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']

  - job_name: 'node-exporter'
    static_configs:
      - targets: ['localhost:9100']

  - job_name: 'cloudbridge-relay'
    static_configs:
      - targets: ['localhost:8081']
EOF

# Запуск мониторинга
log_info "Запуск системы мониторинга..."
cd /opt/cloudbridge/monitoring
docker-compose up -d

log_success "Модуль 07 завершен успешно"
log_info "Система мониторинга установлена"
log_info "Prometheus: http://localhost:9090"
log_info "Grafana: http://localhost:3000 (admin/admin123)"
EOF

# Модуль 08: Final Configuration
cat > 08-final-config.sh << 'EOF'
#!/bin/bash

# CloudBridge Relay - Module 08: Final Configuration
# Финальная настройка и оптимизация
# Автор: CloudBridge Team
# Версия: 1.0

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Функции логирования
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Проверка root прав
if [[ $EUID -ne 0 ]]; then
    log_error "Этот модуль должен быть запущен с правами root"
    exit 1
fi

log_info "Модуль 08: Финальная настройка"

# Создание системных сервисов
log_info "Создание системных сервисов..."

# Сервис для CloudBridge Relay
cat > /etc/systemd/system/cloudbridge-relay.service << 'EOF'
[Unit]
Description=CloudBridge Relay Service
After=network.target docker.service
Requires=docker.service

[Service]
Type=oneshot
RemainAfterExit=yes
WorkingDirectory=/opt/cloudbridge
ExecStart=/usr/bin/docker-compose up -d
ExecStop=/usr/bin/docker-compose down
TimeoutStartSec=0

[Install]
WantedBy=multi-user.target
EOF

# Включение сервиса
systemctl daemon-reload
systemctl enable cloudbridge-relay.service

# Настройка резервного копирования
log_info "Настройка резервного копирования..."
cat > /usr/local/bin/cloudbridge-backup.sh << 'EOF'
#!/bin/bash

# CloudBridge Relay Backup Script
BACKUP_DIR="/opt/cloudbridge/backups"
DATE=$(date +%Y%m%d_%H%M%S)

mkdir -p "$BACKUP_DIR"

# Backup конфигураций
tar -czf "$BACKUP_DIR/config_$DATE.tar.gz" /etc/cloudbridge /opt/cloudbridge/configs

# Backup данных
tar -czf "$BACKUP_DIR/data_$DATE.tar.gz" /opt/cloudbridge/data

# Очистка старых бэкапов (оставляем последние 7)
find "$BACKUP_DIR" -name "*.tar.gz" -mtime +7 -delete

echo "Backup completed: $DATE"
EOF

chmod +x /usr/local/bin/cloudbridge-backup.sh

# Добавление в cron
echo "0 2 * * * /usr/local/bin/cloudbridge-backup.sh" | crontab -

# Создание утилит управления
log_info "Создание утилит управления..."

# Статус сервисов
cat > /usr/local/bin/cloudbridge-status << 'EOF'
#!/bin/bash

echo "=== CloudBridge Relay Status ==="
echo "Docker: $(systemctl is-active docker)"
echo "Keycloak: $(docker ps --filter name=cloudbridge-keycloak --format '{{.Status}}')"
echo "Prometheus: $(docker ps --filter name=cloudbridge-prometheus --format '{{.Status}}')"
echo "Grafana: $(docker ps --filter name=cloudbridge-grafana --format '{{.Status}}')"
echo "UFW: $(systemctl is-active ufw)"
echo "Fail2Ban: $(systemctl is-active fail2ban)"
EOF

chmod +x /usr/local/bin/cloudbridge-status

# Проверка безопасности
cat > /usr/local/bin/cloudbridge-security-check << 'EOF'
#!/bin/bash

echo "=== CloudBridge Security Check ==="
echo "SSH Status: $(systemctl is-active ssh)"
echo "UFW Status: $(systemctl is-active ufw)"
echo "Fail2Ban Status: $(systemctl is-active fail2ban)"
echo "Lynis Audit: $(lynis audit system --quick --no-colors | tail -5)"
EOF

chmod +x /usr/local/bin/cloudbridge-security-check

# Финальная проверка
log_info "Выполнение финальной проверки..."

# Проверка сервисов
systemctl is-active --quiet docker || log_warning "Docker не активен"
systemctl is-active --quiet ufw || log_warning "UFW не активен"
systemctl is-active --quiet fail2ban || log_warning "Fail2Ban не активен"

# Проверка портов
netstat -tlnp | grep -E "(8080|8081|8082|9090|3000)" || log_warning "Некоторые порты не открыты"

log_success "Модуль 08 завершен успешно"
log_info "Финальная настройка завершена"
log_info "Используйте 'cloudbridge-status' для проверки состояния"
log_info "Используйте 'cloudbridge-security-check' для проверки безопасности"
EOF

# Создание README.md
cat > README.md << 'EOF'
# CloudBridge Relay Installation Modules

Репозиторий модулей для автоматической установки CloudBridge Relay на Ubuntu 22.04.

## Быстрый старт

```bash
# Автоматическая установка
curl -sSL https://raw.githubusercontent.com/twogc/cloudbridge-relay/main/ubuntu_22_04_installer.sh | sudo bash
```

## Модули

- `01-system-update.sh` - Обновление системы
- `02-security-basics.sh` - Базовая безопасность
- `03-docker-install.sh` - Установка Docker
- `04-firewall-setup.sh` - Настройка файрвола
- `05-keycloak-setup.sh` - Установка Keycloak
- `06-lynis-setup.sh` - Аудит безопасности
- `07-monitoring-setup.sh` - Мониторинг
- `08-final-config.sh` - Финальная настройка

## Документация

Подробная документация доступна в основном репозитории проекта.
EOF

# Создание .gitignore
cat > .gitignore << 'EOF'
# Logs
*.log
logs/

# Temporary files
*.tmp
*.temp

# Backup files
*.bak
*.backup

# OS generated files
.DS_Store
.DS_Store?
._*
.Spotlight-V100
.Trashes
ehthumbs.db
Thumbs.db

# IDE files
.vscode/
.idea/
*.swp
*.swo

# Docker
.dockerignore

# Environment files
.env
.env.local
.env.production
EOF

# Делаем все модули исполняемыми
chmod +x *.sh

# Добавление файлов в Git
log_info "Добавление файлов в Git..."
git add .

# Первый коммит
log_info "Создание первого коммита..."
git commit -m "Initial commit: CloudBridge Relay installation modules

- Added 8 installation modules
- Added modules.json configuration
- Added README.md documentation
- Added .gitignore file

Modules:
- 01-system-update.sh: System update and basic packages
- 02-security-basics.sh: Security hardening
- 03-docker-install.sh: Docker installation
- 04-firewall-setup.sh: UFW firewall setup
- 05-keycloak-setup.sh: Keycloak authentication
- 06-lynis-setup.sh: Security audit with Lynis
- 07-monitoring-setup.sh: Prometheus and Grafana
- 08-final-config.sh: Final configuration and optimization"

# Переименование ветки в main
git branch -M main

# Отправка в репозиторий
log_info "Отправка в репозиторий..."
git push -u origin main

log_success "Репозиторий успешно настроен!"
log_info "URL: $REPO_URL"
log_info "Все модули загружены и готовы к использованию"

# Возврат в исходную директорию
cd ..

log_info "Настройка завершена. Теперь можно использовать модульную установку:"
log_info "curl -sSL https://raw.githubusercontent.com/twogc/cloudbridge-relay/main/ubuntu_22_04_installer.sh | sudo bash"




