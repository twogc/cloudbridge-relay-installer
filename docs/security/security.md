# Руководство по безопасности CloudBridge Relay

Это руководство содержит рекомендации по обеспечению безопасности CloudBridge Relay в соответствии с российскими и международными стандартами.

## 🔒 Обязательные меры безопасности

### 1. Изменение паролей по умолчанию

**КРИТИЧНО**: Все пароли по умолчанию должны быть изменены перед использованием в продакшене.

```bash
# Автоматическая генерация безопасных секретов
./scripts/generate_secrets.sh --env .env

# Проверка измененных паролей
grep -E "(PASSWORD|SECRET)" .env
```

### 2. Настройка HTTPS/TLS

```bash
# Генерация самоподписанных сертификатов (для тестирования)
./scripts/setup_certs.sh

# Для продакшена используйте реальные сертификаты
# Обновите в .env:
RELAY_TLS_CERT_FILE=/path/to/your/certificate.pem
RELAY_TLS_KEY_FILE=/path/to/your/private-key.pem
```

### 3. Настройка Firewall

```bash
# Базовые правила
sudo ufw default deny incoming
sudo ufw default allow outgoing

# Разрешенные порты
sudo ufw allow 22/tcp    # SSH
sudo ufw allow 8080/tcp  # Relay API (только для внутренней сети)
sudo ufw allow 8081/tcp  # Metrics (только для внутренней сети)

# Активация
sudo ufw enable
```

### 4. Ограничение доступа к API

```bash
# В .env файле настройте разрешенные IP
RELAY_ALLOWED_IPS=192.168.1.0/24,10.0.0.0/8,127.0.0.1/32

# Rate limiting
RELAY_RATE_LIMIT_RPS=100
RELAY_RATE_LIMIT_BURST=200
```

## 🛡️ Соответствие российским требованиям

### 152-ФЗ (Защита персональных данных)

- ✅ **Шифрование данных в состоянии покоя**
- ✅ **Шифрование данных в передаче (TLS)**
- ✅ **Логирование доступа к персональным данным**
- ✅ **Контроль доступа и аутентификация**
- ✅ **Регулярное резервное копирование**

### 187-ФЗ (Безопасность критической инфраструктуры)

- ✅ **Мониторинг безопасности**
- ✅ **Обнаружение инцидентов**
- ✅ **Аудит событий**
- ✅ **Резервное копирование**
- ✅ **Восстановление после сбоев**

### Криптографические алгоритмы

```bash
# ГОСТ Р 34.10-2012 (подпись)
GOST_SIGNATURE_ALGORITHM=GOST_R_34_10_2012

# ГОСТ Р 34.11-2012 (хеш)
GOST_HASH_ALGORITHM=GOST_R_34_11_2012

# Использование российских алгоритмов
USE_RUSSIAN_CRYPTO=true
```

## 🔐 Управление секретами

### Генерация секретов

```bash
# JWT секрет (32 байта)
openssl rand -base64 32

# Пароль (16 символов)
openssl rand -base64 16 | tr -d "=+/" | cut -c1-16

# Использование скрипта
./scripts/generate_secrets.sh --all
```

### Хранение секретов

**НЕ РЕКОМЕНДУЕТСЯ:**
- Хранение в файлах конфигурации
- Коммит в git репозиторий
- Передача по незащищенным каналам

**РЕКОМЕНДУЕТСЯ:**
- Использование менеджеров секретов
- Переменные окружения
- Шифрованные хранилища

### Менеджеры секретов

#### HashiCorp Vault
```bash
# Установка Vault
curl -fsSL https://apt.releases.hashicorp.com/gpg | sudo apt-key add -
sudo apt-add-repository "deb [arch=amd64] https://apt.releases.hashicorp.com $(lsb_release -cs) main"
sudo apt-get update && sudo apt-get install vault

# Настройка
vault server -dev
export VAULT_ADDR='http://127.0.0.1:8200'
vault kv put secret/relay jwt_secret="$(openssl rand -base64 32)"
```

#### AWS Secrets Manager
```bash
# Установка AWS CLI
aws configure

# Создание секрета
aws secretsmanager create-secret \
    --name relay/jwt-secret \
    --description "JWT secret for CloudBridge Relay" \
    --secret-string "{\"jwt_secret\":\"$(openssl rand -base64 32)\"}"
```

## 📊 Мониторинг безопасности

### Логирование

```bash
# Настройка уровня логирования
RELAY_LOG_LEVEL=INFO
RELAY_LOG_FORMAT=json

# Ротация логов
sudo logrotate -f /etc/logrotate.d/relay
```

### Алерты безопасности

```yaml
# prometheus/alerts/security_alerts.yml
groups:
  - name: security
    rules:
      - alert: HighFailedAuthAttempts
        expr: rate(relay_auth_failures_total[5m]) > 10
        for: 1m
        labels:
          severity: warning
        annotations:
          summary: "Высокое количество неудачных попыток аутентификации"
          
      - alert: UnauthorizedAccessAttempt
        expr: relay_unauthorized_requests_total > 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Попытка несанкционированного доступа"
```

### Аудит событий

```bash
# Мониторинг доступа к API
tail -f /var/log/relay/relay.log | grep -E "(auth|unauthorized|failed)"

# Проверка активных соединений
curl http://localhost:8081/metrics | grep relay_active_connections

# Аудит файлов конфигурации
find /etc/relay -type f -exec stat -c "%n %a %U %G" {} \;
```

## 🔄 Ротация секретов

### План ротации

1. **JWT секреты**: Каждые 30 дней
2. **Пароли администраторов**: Каждые 90 дней
3. **Сертификаты**: Согласно сроку действия
4. **Ключи API**: Каждые 180 дней

### Процедура ротации

```bash
# 1. Генерация новых секретов
./scripts/generate_secrets.sh --jwt > new_secrets.txt

# 2. Обновление в менеджере секретов
vault kv put secret/relay jwt_secret="$(cat new_secrets.txt | grep JWT_SECRET | cut -d'=' -f2)"

# 3. Перезапуск сервисов
docker-compose -f docker-compose.keycloak.yml restart relay

# 4. Проверка работоспособности
curl http://localhost:8080/health
```

## 🚨 Инциденты безопасности

### Процедура реагирования

1. **Обнаружение**: Мониторинг и алерты
2. **Оценка**: Анализ угрозы и воздействия
3. **Сдерживание**: Ограничение доступа
4. **Устранение**: Исправление уязвимостей
5. **Восстановление**: Возврат к нормальной работе
6. **Анализ**: Документирование и улучшения

### Контакты для инцидентов

```bash
# Логирование инцидента
echo "$(date): Security incident detected" >> /var/log/security/incidents.log

# Уведомление администраторов
curl -X POST -H 'Content-type: application/json' \
  --data '{"text":"Security incident detected on relay server"}' \
  $SLACK_WEBHOOK_URL
```

## 📋 Чек-лист безопасности

### Установка
- [ ] Изменены все пароли по умолчанию
- [ ] Настроен HTTPS/TLS
- [ ] Настроен firewall
- [ ] Ограничены разрешенные IP
- [ ] Настроен rate limiting

### Эксплуатация
- [ ] Регулярная ротация секретов
- [ ] Мониторинг безопасности
- [ ] Резервное копирование
- [ ] Обновление сертификатов
- [ ] Аудит доступа

### Соответствие
- [ ] Соответствие 152-ФЗ
- [ ] Соответствие 187-ФЗ
- [ ] Использование ГОСТ алгоритмов
- [ ] Логирование событий
- [ ] Контроль доступа

## 📚 Дополнительные ресурсы

- [NIST Cybersecurity Framework](https://www.nist.gov/cyberframework)
- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Роскомнадзор - 152-ФЗ](https://pd.rkn.gov.ru/)
- [ФСТЭК России - 187-ФЗ](https://fstec.ru/)

---

**Важно**: Безопасность - это непрерывный процесс. Регулярно пересматривайте и обновляйте меры безопасности. 

## Обзор

CloudBridge Relay v2.0 включает комплексную систему безопасности, обеспечивающую защиту от различных типов атак и несанкционированного доступа.

## Архитектура безопасности

### 1. Сетевая безопасность

#### Firewall (UFW + iptables)
- **SSH (порт 22)**: Разрешен для административного доступа
- **Relay Server (порт 8082)**: Разрешен для клиентских соединений
- **Порты мониторинга**: Доступны только с localhost (127.0.0.1)
  - Metrics (8081)
  - Keycloak (8080)
  - Grafana (3000)
  - Prometheus (9090)
  - AlertManager (9093)
  - Blackbox (9115)

#### Дополнительная защита
- **DDoS защита**: Rate limiting для SSH и Relay портов
- **Блокировка подозрительных пакетов**: Фильтрация TCP флагов
- **Защита от сканирования**: Автоматическая блокировка при множественных попытках

### 2. Аутентификация и авторизация

#### Keycloak Integration
- **OAuth 2.0 / OpenID Connect**: Современные протоколы аутентификации
- **JWT токены**: Безопасная передача информации о пользователе
- **Multi-factor authentication**: Поддержка MFA через Keycloak
- **Session management**: Централизованное управление сессиями

#### mTLS (Mutual TLS)
- **Клиентские сертификаты**: Дополнительный уровень аутентификации
- **CA валидация**: Проверка подлинности клиентских сертификатов

### 3. Шифрование

#### TLS/SSL
- **Let's Encrypt сертификаты**: Автоматическое обновление
- **Strong cipher suites**: Современные алгоритмы шифрования
- **Perfect Forward Secrecy**: Защита от компрометации ключей

#### Российские криптоалгоритмы
- **GOST R 34.10-2012**: Алгоритм цифровой подписи
- **GOST R 34.11-2012**: Хеш-функция
- **Соответствие 152-ФЗ**: Защита персональных данных

### 4. Мониторинг и обнаружение угроз

#### fail2ban
- **Автоматическая блокировка**: IP адресов при подозрительной активности
- **Настраиваемые правила**: Для SSH и Relay сервера
- **Логирование**: Детальные логи всех событий

#### Prometheus + Grafana
- **Метрики безопасности**: Отслеживание попыток атак
- **Алерты**: Уведомления о подозрительной активности
- **Дашборды**: Визуализация состояния безопасности

## Настройка Firewall

### Автоматическая настройка

```bash
# Запуск скрипта настройки firewall
sudo ./scripts/setup_firewall.sh

# Проверка статуса
sudo ./scripts/setup_firewall.sh --check

# Просмотр плана действий
sudo ./scripts/setup_firewall.sh --dry-run
```

### Ручная настройка

#### UFW (Uncomplicated Firewall)
```bash
# Установка UFW
sudo apt-get install ufw

# Настройка политик по умолчанию
sudo ufw default deny incoming
sudo ufw default allow outgoing

# Разрешение SSH
sudo ufw allow 22/tcp

# Разрешение Relay порта
sudo ufw allow 8082/tcp

# Разрешение портов мониторинга только для localhost
sudo ufw allow from 127.0.0.1 to any port 8081
sudo ufw allow from 127.0.0.1 to any port 8080
sudo ufw allow from 127.0.0.1 to any port 3000
sudo ufw allow from 127.0.0.1 to any port 9090
sudo ufw allow from 127.0.0.1 to any port 9093
sudo ufw allow from 127.0.0.1 to any port 9115

# Включение UFW
sudo ufw enable
```

#### iptables (дополнительная защита)
```bash
# Очистка правил
sudo iptables -F
sudo iptables -X

# Установка политик по умолчанию
sudo iptables -P INPUT DROP
sudo iptables -P FORWARD DROP
sudo iptables -P OUTPUT ACCEPT

# Разрешение loopback
sudo iptables -A INPUT -i lo -j ACCEPT

# Разрешение установленных соединений
sudo iptables -A INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT

# Разрешение SSH с rate limiting
sudo iptables -A INPUT -p tcp --dport 22 -m limit --limit 6/min --limit-burst 10 -j ACCEPT

# Разрешение Relay порта с rate limiting
sudo iptables -A INPUT -p tcp --dport 8082 -m limit --limit 30/min --limit-burst 60 -j ACCEPT

# Сохранение правил
sudo iptables-save > /etc/iptables/rules.v4
```

### Управление Firewall

```bash
# Показать статус
relay-firewall status

# Разрешить IP адрес
relay-firewall allow-ip 192.168.1.100

# Заблокировать IP адрес
relay-firewall deny-ip 10.0.0.50

# Перезагрузить firewall
relay-firewall reload
```

## Настройка fail2ban

### Установка и настройка
```bash
# Установка fail2ban
sudo apt-get install fail2ban

# Создание конфигурации
sudo nano /etc/fail2ban/jail.local
```

### Конфигурация jail.local
```ini
[DEFAULT]
bantime = 3600
findtime = 600
maxretry = 3

[sshd]
enabled = true
port = ssh
filter = sshd
logpath = /var/log/auth.log
maxretry = 3

[relay]
enabled = true
port = 8082
filter = relay
logpath = /var/log/relay/relay.log
maxretry = 5
```

### Фильтр для Relay
```bash
# Создание фильтра
sudo nano /etc/fail2ban/filter.d/relay.conf
```

Содержимое фильтра:
```ini
[Definition]
failregex = ^.*authentication failed.*from <HOST>.*$
ignoreregex =
```

### Управление fail2ban
```bash
# Перезапуск fail2ban
sudo systemctl restart fail2ban

# Проверка статуса
sudo fail2ban-client status

# Проверка конкретной jail
sudo fail2ban-client status sshd
sudo fail2ban-client status relay

# Разблокировка IP
sudo fail2ban-client set sshd unbanip 192.168.1.100
```

## Безопасность сертификатов

### Let's Encrypt
```bash
# Получение сертификатов
sudo certbot certonly --dns-cloudflare --dns-cloudflare-credentials ~/.secrets/cloudflare.ini -d edge.2gc.ru

# Автоматическое обновление
sudo crontab -e
# Добавить строку:
# 0 12 * * * /usr/bin/certbot renew --quiet
```

### Post-renewal hook
```bash
# Создание скрипта обновления
sudo nano /etc/letsencrypt/renewal-hooks/post/restart-relay.sh
```

Содержимое скрипта:
```bash
#!/bin/bash
docker-compose restart relay
```

```bash
# Сделать исполняемым
sudo chmod +x /etc/letsencrypt/renewal-hooks/post/restart-relay.sh
```

## Мониторинг безопасности

### Метрики безопасности
- **Попытки аутентификации**: Успешные и неуспешные
- **Блокировки IP**: Количество заблокированных адресов
- **DDoS атаки**: Обнаружение и блокировка
- **Сканирование портов**: Мониторинг подозрительной активности

### Алерты
- **Множественные неудачные попытки входа**
- **Подозрительная активность на портах**
- **DDoS атаки**
- **Попытки доступа к заблокированным портам**

### Дашборды безопасности
- **Security Analytics Dashboard**: Анализ угроз и атак
- **Tunnel Management Dashboard**: Мониторинг туннелей
- **Multi-Tenant Analytics**: Анализ активности клиентов

## Соответствие требованиям

### 152-ФЗ (Защита персональных данных)
- ✅ Шифрование данных в покое и при передаче
- ✅ Контроль доступа и аутентификация
- ✅ Логирование и мониторинг
- ✅ Уведомления о нарушениях

### 187-ФЗ (Безопасность критической инфраструктуры)
- ✅ Защита от DDoS атак
- ✅ Обнаружение вторжений
- ✅ Резервное копирование
- ✅ Восстановление после инцидентов

### ГОСТ Р 57580.1-2017
- ✅ Использование российских криптоалгоритмов
- ✅ Сертифицированные средства защиты
- ✅ Контроль целостности

## Рекомендации по безопасности

### Регулярные проверки
1. **Ежедневно**: Проверка логов безопасности
2. **Еженедельно**: Обновление системы и пакетов
3. **Ежемесячно**: Аудит настроек безопасности
4. **Ежеквартально**: Тестирование на проникновение

### Резервное копирование
```bash
# Автоматическое резервное копирование
./scripts/backup_keycloak.sh

# Проверка целостности резервных копий
./scripts/test_integration.sh
```

### Обновления
```bash
# Обновление системы
sudo apt-get update && sudo apt-get upgrade

# Обновление Docker образов
docker-compose pull
docker-compose up -d
```

### Мониторинг
```bash
# Проверка статуса сервисов
docker-compose ps

# Просмотр логов
docker-compose logs -f relay

# Проверка метрик
curl http://localhost:8081/metrics
```

## Устранение неполадок

### Проблемы с firewall
```bash
# Проверка статуса UFW
sudo ufw status verbose

# Проверка правил iptables
sudo iptables -L -n -v

# Временное отключение firewall (только для диагностики!)
sudo ufw disable
```

### Проблемы с fail2ban
```bash
# Проверка статуса
sudo fail2ban-client status

# Просмотр логов
sudo tail -f /var/log/fail2ban.log

# Разблокировка всех IP
sudo fail2ban-client reload
```

### Проблемы с сертификатами
```bash
# Проверка сертификатов
sudo certbot certificates

# Тестирование обновления
sudo certbot renew --dry-run

# Проверка конфигурации nginx
sudo nginx -t
```

## Контакты для безопасности

При обнаружении уязвимостей или проблем безопасности:
- **Email**: security@cloudbridge.ru
- **Telegram**: @cloudbridge_security
- **Телефон**: +7 (495) XXX-XX-XX

---

*Данное руководство обновляется регулярно. Последнее обновление: $(date)* 