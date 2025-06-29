#!/bin/bash

# CloudBridge Relay - Security Hardening Script
# Автоматическое исправление основных проблем безопасности

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1${NC}"
}

# Проверка root прав
if [[ $EUID -ne 0 ]]; then
    error "Этот скрипт должен быть запущен с правами root"
    exit 1
fi

log "Начало автоматического исправления проблем безопасности..."

# 1. SSH Hardening
log "Исправление настроек SSH..."
if [[ -f /etc/ssh/sshd_config ]]; then
    # Создаём резервную копию
    cp /etc/ssh/sshd_config /etc/ssh/sshd_config.backup.$(date +%Y%m%d-%H%M%S)
    
    # Применяем безопасные настройки SSH
    sed -i 's/^#*PermitRootLogin.*/PermitRootLogin no/' /etc/ssh/sshd_config
    sed -i 's/^#*AllowTcpForwarding.*/AllowTcpForwarding no/' /etc/ssh/sshd_config
    sed -i 's/^#*X11Forwarding.*/X11Forwarding no/' /etc/ssh/sshd_config
    sed -i 's/^#*AllowAgentForwarding.*/AllowAgentForwarding no/' /etc/ssh/sshd_config
    sed -i 's/^#*ClientAliveCountMax.*/ClientAliveCountMax 2/' /etc/ssh/sshd_config
    sed -i 's/^#*MaxAuthTries.*/MaxAuthTries 3/' /etc/ssh/sshd_config
    sed -i 's/^#*MaxSessions.*/MaxSessions 2/' /etc/ssh/sshd_config
    sed -i 's/^#*TCPKeepAlive.*/TCPKeepAlive no/' /etc/ssh/sshd_config
    sed -i 's/^#*LogLevel.*/LogLevel VERBOSE/' /etc/ssh/sshd_config
    
    # Перезапускаем SSH
    systemctl reload ssh
    log "SSH настройки обновлены"
fi

# 2. Kernel Hardening
log "Применение настроек безопасности ядра..."
cat >> /etc/sysctl.conf << 'EOF'

# Security hardening settings
# Core dumps
fs.suid_dumpable = 0
kernel.core_uses_pid = 1

# Kernel pointer restrictions
kernel.kptr_restrict = 2

# Module loading
kernel.modules_disabled = 1

# Sysrq
kernel.sysrq = 0

# Unprivileged BPF
kernel.unprivileged_bpf_disabled = 1

# Network security
net.core.bpf_jit_harden = 2
net.ipv4.conf.all.log_martians = 1
net.ipv4.conf.default.log_martians = 1
net.ipv4.conf.all.rp_filter = 1
net.ipv4.conf.all.forwarding = 0
net.ipv4.conf.all.send_redirects = 0

# File system protection
fs.protected_fifos = 2
fs.protected_hardlinks = 1
fs.protected_regular = 2
fs.protected_symlinks = 1

# TTY protection
dev.tty.ldisc_autoload = 0
EOF

# Применяем настройки
sysctl -p
log "Настройки ядра обновлены"

# 3. File Permissions
log "Исправление прав доступа к файлам..."
chmod 600 /etc/ssh/sshd_config
chmod 600 /etc/crontab
chmod 600 /boot/grub/grub.cfg

# Директории cron
chmod 700 /etc/cron.d
chmod 700 /etc/cron.daily
chmod 700 /etc/cron.hourly
chmod 700 /etc/cron.weekly
chmod 700 /etc/cron.monthly

log "Права доступа к файлам исправлены"

# 4. Password Policy
log "Настройка политики паролей..."
cat >> /etc/login.defs << 'EOF'

# Password aging
PASS_MAX_DAYS   90
PASS_MIN_DAYS   1
PASS_WARN_AGE   7

# Password hashing
ENCRYPT_METHOD SHA512
SHA_CRYPT_MIN_ROUNDS 5000
SHA_CRYPT_MAX_ROUNDS 10000

# Default umask
UMASK 027
EOF

log "Политика паролей настроена"

# 5. Banner Configuration
log "Настройка баннеров безопасности..."
cat > /etc/issue << 'EOF'
***************************************************************************
                        NOTICE TO USERS

This computer system is the private property of its owner, whether
individual, corporate or government. It is for authorized use only.
Users (authorized or unauthorized) have no explicit or implicit
expectation of privacy.

By using this system, you consent to monitoring and recording of your
activities. Unauthorized or improper use of this system may result in
civil and criminal penalties and administrative or disciplinary action,
as appropriate. By continuing to use this system you indicate your
awareness of and consent to these terms and conditions of use.

LOG OFF IMMEDIATELY if you do not agree to the conditions stated in this
warning.
***************************************************************************
EOF

cp /etc/issue /etc/issue.net
log "Баннеры безопасности настроены"

# 6. Disable USB Storage
log "Отключение USB storage..."
echo "install usb-storage /bin/true" > /etc/modprobe.d/usb-storage.conf
log "USB storage отключен"

# 7. Install Security Tools
log "Установка инструментов безопасности..."
apt-get update
apt-get install -y debsums apt-show-versions auditd

# Включаем auditd
systemctl enable auditd
systemctl start auditd
log "Инструменты безопасности установлены"

# 8. Configure Audit Rules
log "Настройка правил аудита..."
cat > /etc/audit/rules.d/security.rules << 'EOF'
# Monitor file access
-w /etc/passwd -p wa -k identity
-w /etc/group -p wa -k identity
-w /etc/shadow -p wa -k identity
-w /etc/gshadow -p wa -k identity

# Monitor configuration changes
-w /etc/ssh/sshd_config -p wa -k sshd
-w /etc/sudoers -p wa -k sudoers
-w /etc/crontab -p wa -k cron

# Monitor system calls
-a always,exit -F arch=b64 -S chmod -S fchmod -S fchmodat -F auid>=1000 -F auid!=4294967295 -k perm_mod
-a always,exit -F arch=b64 -S chown -S fchown -S fchownat -S lchown -F auid>=1000 -F auid!=4294967295 -k perm_mod
-a always,exit -F arch=b64 -S setxattr -S lsetxattr -S fsetxattr -S removexattr -S lremovexattr -S fremovexattr -F auid>=1000 -F auid!=4294967295 -k perm_mod
EOF

# Перезапускаем auditd
systemctl restart auditd
log "Правила аудита настроены"

# 9. Fail2ban Configuration
log "Настройка Fail2ban..."
if [[ -f /etc/fail2ban/jail.local ]]; then
    cat >> /etc/fail2ban/jail.local << 'EOF'

[sshd]
enabled = true
port = ssh
filter = sshd
logpath = /var/log/auth.log
maxretry = 3
bantime = 3600
findtime = 600
EOF
    systemctl restart fail2ban
    log "Fail2ban настроен"
fi

# 10. Clean up old packages
log "Очистка старых пакетов..."
apt-get autoremove -y
apt-get autoclean
log "Старые пакеты удалены"

log "Автоматическое исправление проблем безопасности завершено!"
log "Рекомендуется перезагрузить систему для применения всех изменений"
log "После перезагрузки запустите Lynis снова для проверки улучшений"

# Показываем статистику
echo
log "Статистика исправлений:"
echo "  ✓ SSH настройки обновлены"
echo "  ✓ Kernel hardening применён"
echo "  ✓ Права доступа исправлены"
echo "  ✓ Политика паролей настроена"
echo "  ✓ Баннеры безопасности добавлены"
echo "  ✓ USB storage отключен"
echo "  ✓ Инструменты безопасности установлены"
echo "  ✓ Auditd настроен"
echo "  ✓ Fail2ban обновлён"
echo "  ✓ Старые пакеты удалены" 