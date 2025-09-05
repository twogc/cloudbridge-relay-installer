# CloudBridge Relay - DDoS Protection Guide

## 🛡️ Обзор DDoS защиты

CloudBridge Relay поддерживает многоуровневую защиту от DDoS атак, адаптированную под различные сценарии развертывания.

## 🏗️ Архитектура защиты

### Уровни защиты:

#### 1. **Basic Level** - Быстрый старт
- **L4 защита**: nftables + SYNPROXY
- **L7 защита**: Nginx с rate limiting
- **Мониторинг**: CrowdSec
- **Применение**: Одиночные серверы, тестирование

#### 2. **Advanced Level** - Продакшн готовность
- **L4 защита**: nftables + SYNPROXY + XDP/eBPF
- **L7 защита**: Nginx + CrowdSec + Fail2ban
- **Мониторинг**: FastNetMon + Grafana
- **Применение**: Продакшн развертывания

#### 3. **Enterprise Level** - Масштабируемость
- **Все из Advanced** + Load Balancer
- **Автоматическое масштабирование**
- **Anycast IP поддержка**
- **Применение**: Высоконагруженные системы

## 🚀 Быстрый старт

### Развертывание с DDoS защитой:

```bash
# Базовое развертывание
./scripts/deploy-ddos.sh

# Продвинутое развертывание
./scripts/deploy-ddos.sh --level advanced --region MS1

# Enterprise развертывание
./scripts/deploy-ddos.sh --level enterprise --flavor STD2-16-32
```

### Параметры развертывания:

| Параметр | Описание | Значения по умолчанию |
|----------|----------|----------------------|
| `--level` | Уровень DDoS защиты | `advanced` |
| `--region` | Регион развертывания | `MS1` |
| `--flavor` | Конфигурация VM | `STD2-8-16` |
| `--disk-size` | Размер диска (GB) | `120` |

## 🛠️ Компоненты защиты

### L4 (Сетевой уровень)

#### nftables + SYNPROXY
```bash
# Проверка статуса
sudo ddos-status

# Просмотр правил
sudo nft list ruleset

# Настройка SYNPROXY
sudo nft add rule inet filter input tcp flags syn tcp dport { 80, 443, 8080 } ct state new synproxy
```

#### XDP/eBPF фильтрация
```bash
# Установка XDP программ
cd /opt/xdp-tutorial
make

# Загрузка программы
sudo ./xdp_loader -d eth0 -F --progsec xdp_pass
```

### L7 (Прикладной уровень)

#### Nginx с rate limiting
```nginx
# Конфигурация rate limiting
limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;
limit_req_zone $binary_remote_addr zone=general:10m rate=100r/s;

location /api/ {
    limit_req zone=api burst=20 nodelay;
    proxy_pass http://127.0.0.1:8080/;
}
```

#### CrowdSec
```bash
# Проверка статуса
sudo cscli metrics

# Просмотр блокировок
sudo cscli decisions list

# Настройка парсеров
sudo cscli parsers install crowdsecurity/nginx-logs
```

#### Fail2ban
```bash
# Статус
sudo fail2ban-client status

# Просмотр jail
sudo fail2ban-client status nginx-http-auth
```

### Мониторинг

#### FastNetMon
```bash
# Запуск мониторинга
sudo fastnetmon --configuration_file /etc/fastnetmon.conf

# Просмотр метрик
curl http://localhost:8080/fastnetmon
```

#### Grafana Dashboard
- URL: `http://your-server:3000`
- Логин: `admin/admin`
- Дашборды: DDoS Protection, Network Traffic, Security Events

## 📊 Метрики и мониторинг

### Ключевые метрики:

1. **Network Traffic**
   - Packets per second (PPS)
   - Bits per second (BPS)
   - Connection rate

2. **DDoS Attacks**
   - Attack types detected
   - Blocked IPs
   - Mitigation actions

3. **Application Performance**
   - Response times
   - Error rates
   - Rate limiting hits

### Алерты:

```yaml
# Пример конфигурации алертов
alerts:
  - name: "High PPS"
    condition: "pps > 10000"
    action: "notify_admin"
    
  - name: "DDoS Attack"
    condition: "attack_detected"
    action: "enable_mitigation"
```

## 🔧 Настройка и оптимизация

### Оптимизация производительности:

```bash
# Настройка сетевых параметров
echo 'net.core.rmem_max = 16777216' >> /etc/sysctl.conf
echo 'net.core.wmem_max = 16777216' >> /etc/sysctl.conf
echo 'net.ipv4.tcp_rmem = 4096 87380 16777216' >> /etc/sysctl.conf
echo 'net.ipv4.tcp_wmem = 4096 65536 16777216' >> /etc/sysctl.conf

# Применение изменений
sudo sysctl -p
```

### Настройка XDP:

```bash
# Компиляция XDP программы
clang -O2 -g -Wall -target bpf -c xdp_ddos_filter.c -o xdp_ddos_filter.o

# Загрузка в ядро
sudo ip link set dev eth0 xdp obj xdp_ddos_filter.o sec xdp
```

## 🚨 Реагирование на атаки

### Автоматические действия:

1. **Обнаружение атаки**
   - FastNetMon детектирует аномалии
   - CrowdSec анализирует логи

2. **Мгновенная блокировка**
   - XDP/eBPF дропает пакеты
   - nftables блокирует IP

3. **Уведомления**
   - Email/SMS алерты
   - Slack/Discord интеграция

### Ручные действия:

```bash
# Блокировка IP
sudo nft add element inet filter blacklist { 192.168.1.100 }

# Разблокировка IP
sudo nft delete element inet filter blacklist { 192.168.1.100 }

# Просмотр блокировок
sudo nft list set inet filter blacklist
```

## 🌐 Anycast и балансировка

### Anycast IP настройка:

```bash
# Настройка BGP
sudo vtysh
router bgp 65000
 network 203.0.113.0/24
 neighbor 192.168.1.1 remote-as 65001
```

### Load Balancer конфигурация:

```yaml
# HAProxy конфигурация
global
    maxconn 50000
    tune.ssl.default-dh-param 2048

frontend relay_frontend
    bind *:80
    bind *:443 ssl crt /etc/ssl/certs/relay.pem
    mode http
    option forwardfor
    default_backend relay_backend

backend relay_backend
    mode http
    balance roundrobin
    server relay1 10.0.1.10:8080 check
    server relay2 10.0.1.11:8080 check
```

## 📈 Масштабирование

### Горизонтальное масштабирование:

1. **Добавление узлов**
   ```bash
   # Клонирование конфигурации
   terraform apply -var="instance_count=3"
   ```

2. **Настройка кластера**
   ```bash
   # Конфигурация кластера
   ./scripts/setup-cluster.sh --nodes 3
   ```

3. **Балансировка нагрузки**
   ```bash
   # HAProxy конфигурация
   ./scripts/configure-haproxy.sh
   ```

### Вертикальное масштабирование:

```bash
# Увеличение ресурсов
./scripts/deploy-ddos.sh --flavor STD2-16-32 --disk-size 200
```

## 🔍 Тестирование защиты

### Пентест скрипты:

```bash
# Базовый пентест
./scripts/pentest/safe_relay_pentest.sh

# Полный пентест
./scripts/pentest/relay_pentest.sh

# DDoS тестирование
./scripts/pentest/ddos_test.sh
```

### Инструменты тестирования:

- **hping3**: SYN flood, UDP flood
- **slowhttptest**: Slow HTTP attacks
- **ab**: Apache Bench для нагрузочного тестирования
- **wrk**: HTTP benchmarking

## 📋 Чек-лист развертывания

### Перед развертыванием:
- [ ] Выбрать уровень защиты
- [ ] Определить требования к ресурсам
- [ ] Настроить мониторинг
- [ ] Подготовить SSL сертификаты

### После развертывания:
- [ ] Проверить статус всех сервисов
- [ ] Настроить алерты
- [ ] Протестировать защиту
- [ ] Документировать конфигурацию

### Регулярное обслуживание:
- [ ] Обновление правил CrowdSec
- [ ] Анализ логов атак
- [ ] Оптимизация производительности
- [ ] Резервное копирование конфигураций

## 🆘 Устранение неполадок

### Частые проблемы:

1. **Высокая нагрузка на CPU**
   ```bash
   # Проверка XDP программ
   sudo ip link show dev eth0
   
   # Оптимизация nftables
   sudo nft optimize
   ```

2. **Ложные срабатывания**
   ```bash
   # Настройка whitelist
   sudo nft add element inet filter whitelist { 10.0.0.0/8 }
   
   # Анализ логов
   sudo journalctl -u crowdsec -f
   ```

3. **Проблемы с производительностью**
   ```bash
   # Мониторинг ресурсов
   htop
   iftop
   nethogs
   ```

## 📚 Дополнительные ресурсы

### Документация:
- [nftables Wiki](https://wiki.nftables.org)
- [CrowdSec Documentation](https://docs.crowdsec.net)
- [FastNetMon Guide](https://fastnetmon.com/docs/)
- [XDP Tutorial](https://github.com/xdp-project/xdp-tutorial)

### Сообщество:
- [CrowdSec Community](https://discord.gg/crowdsec)
- [nftables Mailing List](https://marc.info/?l=netfilter)
- [XDP Project](https://github.com/xdp-project)

---

**Примечание**: Данное руководство описывает базовую настройку DDoS защиты. Для продакшн окружений рекомендуется дополнительная настройка под конкретные требования и нагрузки.

