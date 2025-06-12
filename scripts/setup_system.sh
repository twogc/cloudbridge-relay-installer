#!/bin/bash

# Увеличиваем лимиты файловых дескрипторов
echo "* soft nofile 1000000" | sudo tee -a /etc/security/limits.conf
echo "* hard nofile 1000000" | sudo tee -a /etc/security/limits.conf

# Настройка TCP параметров
cat << EOF | sudo tee /etc/sysctl.d/99-relay.conf
# Увеличиваем размер буферов TCP
net.core.rmem_max = 16777216
net.core.wmem_max = 16777216
net.ipv4.tcp_rmem = 4096 87380 16777216
net.ipv4.tcp_wmem = 4096 87380 16777216

# Оптимизация для большого количества соединений
net.core.somaxconn = 65535
net.ipv4.tcp_max_syn_backlog = 65535
net.ipv4.tcp_syncookies = 1
net.ipv4.tcp_tw_reuse = 1
net.ipv4.tcp_fin_timeout = 30

# Отключаем медленный старт TCP
net.ipv4.tcp_slow_start_after_idle = 0

# Увеличиваем размер хеш-таблиц для соединений
net.ipv4.tcp_max_tw_buckets = 1440000
net.ipv4.tcp_tw_reuse = 1
EOF

# Применяем настройки
sudo sysctl -p /etc/sysctl.d/99-relay.conf 