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