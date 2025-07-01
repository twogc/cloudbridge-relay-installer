# Быстрый старт для московского региона

Это руководство поможет быстро настроить CloudBridge Relay для московского региона с учетом российских требований.

## Предварительные требования

- Ubuntu 20.04+ или CentOS 8+
- Docker и Docker Compose
- Минимум 4GB RAM
- 10GB свободного места на диске

## Установка

### 1. Клонирование репозитория

```bash
git clone https://github.com/your-org/cloudbridge-relay-installer.git
cd cloudbridge-relay-installer
```

### 2. Автоматическая настройка

```bash
# Запуск скрипта настройки для московского региона
./scripts/setup_moscow_region.sh
```

Этот скрипт автоматически выполнит:

- ✅ Создание `.env` файла с настройками для Москвы
- ✅ Настройку часового пояса Europe/Moscow
- ✅ Установку русской локали
- ✅ Создание необходимых директорий
- ✅ Настройку самоподписанных сертификатов
- ✅ Конфигурацию мониторинга
- ✅ Проверку соответствия российским требованиям
- ✅ Запуск всех сервисов

### 3. Проверка установки

После завершения установки вы увидите информацию о доступе к сервисам:

```
┌─────────────────────────────────────────────────────────────┐
│                    CloudBridge Relay                        │
│                    Московский регион                        │
├─────────────────────────────────────────────────────────────┤
│ Keycloak Admin Console: http://localhost:8080               │
│   Логин: admin                                              │
│   Пароль: [установленный в .env]                            │
├─────────────────────────────────────────────────────────────┤
│ Relay API: http://localhost:8082                            │
│ Relay Metrics: http://localhost:8081/metrics                │
├─────────────────────────────────────────────────────────────┤
│ Prometheus: http://localhost:9090                           │
│ AlertManager: http://localhost:9093                         │
│ Grafana: http://localhost:3000                              │
│   Логин: admin                                              │
│   Пароль: [установленный в .env]                            │
└─────────────────────────────────────────────────────────────┘
```

## Первые шаги

### 1. Настройка Keycloak

1. Откройте Keycloak Admin Console: http://localhost:8080
2. Войдите с учетными данными из файла `.env`
3. Создайте realm "cloudbridge"
4. Создайте клиента "relay-client"
5. Настройте роли и пользователей

### 2. Тестирование API

```bash
# Проверка здоровья сервиса
curl http://localhost:8082/health

# Проверка метрик
curl http://localhost:8081/metrics
```

### 3. Настройка мониторинга

1. Откройте Grafana: http://localhost:3000
2. Войдите с учетными данными из файла `.env`
3. Настройте дашборды для мониторинга

## Российские требования

Система автоматически настроена для соответствия:

### 152-ФЗ (Защита персональных данных)
- ✅ Шифрование данных в состоянии покоя
- ✅ Шифрование данных в передаче
- ✅ Логирование доступа к персональным данным
- ✅ Контроль доступа

### 187-ФЗ (Безопасность критической инфраструктуры)
- ✅ Мониторинг безопасности
- ✅ Обнаружение инцидентов
- ✅ Аудит событий
- ✅ Резервное копирование

### Криптографические алгоритмы
- ✅ ГОСТ Р 34.10-2012 (подпись)
- ✅ ГОСТ Р 34.11-2012 (хеш)
- ✅ Поддержка российских сертификатов

## Управление

### Основные команды

```bash
# Запуск сервисов
docker-compose -f docker-compose.keycloak.yml up -d

# Остановка сервисов
docker-compose -f docker-compose.keycloak.yml down

# Просмотр логов
docker-compose -f docker-compose.keycloak.yml logs -f relay

# Перезапуск сервиса
docker-compose -f docker-compose.keycloak.yml restart relay
```

### Резервное копирование

```bash
# Создание резервной копии
./scripts/backup_keycloak.sh

# Восстановление
./scripts/restore_keycloak.sh backup_file.tar.gz
```

### Обновление

```bash
# Обновление сервисов
docker-compose -f docker-compose.keycloak.yml pull
docker-compose -f docker-compose.keycloak.yml up -d
```

## Безопасность

### Обязательные действия после установки

1. **Измените пароли по умолчанию**:
   ```bash
   # В файле .env измените:
   KEYCLOAK_ADMIN_PASSWORD=your-secure-password
   GRAFANA_ADMIN_PASSWORD=your-secure-password
   RELAY_JWT_SECRET=your-super-secure-jwt-secret
   ```

2. **Настройте HTTPS**:
   - Замените самоподписанные сертификаты на реальные
   - Обновите `RELAY_TLS_CERT_FILE` и `RELAY_TLS_KEY_FILE`

3. **Настройте firewall**:
   ```bash
   # Разрешить только необходимые порты
   sudo ufw allow 22/tcp    # SSH
   sudo ufw allow 8080/tcp  # Keycloak Admin Console
   sudo ufw allow 8082/tcp  # Relay API
   sudo ufw allow 8081/tcp  # Metrics
   sudo ufw enable
   ```

4. **Настройте мониторинг безопасности**:
   - Настройте алерты в AlertManager
   - Настройте дашборды в Grafana
   - Регулярно проверяйте логи

## Устранение неполадок

### Частые проблемы

1. **Сервисы не запускаются**
   ```bash
   # Проверка статуса
   docker-compose -f docker-compose.keycloak.yml ps
   
   # Просмотр логов
   docker-compose -f docker-compose.keycloak.yml logs
   ```

2. **Проблемы с аутентификацией**
   - Проверьте конфигурацию клиента в Keycloak
   - Убедитесь, что realm и client_id правильные
   - Проверьте переменные окружения в `.env`

3. **Проблемы с сертификатами**
   ```bash
   # Проверка сертификатов
   openssl x509 -in certs/fullchain.pem -text -noout
   
   # Проверка прав доступа
   ls -la certs/
   ```

### Получение помощи

- **Документация**: [docs/](docs/)
- **Issues**: [GitHub Issues](https://github.com/your-org/cloudbridge-relay-installer/issues)
- **Discussions**: [GitHub Discussions](https://github.com/your-org/cloudbridge-relay-installer/discussions)

## Следующие шаги

После успешной установки рекомендуется:

1. Изучить [полную документацию](docs/)
2. Настроить [мониторинг и алерты](docs/monitoring.md)
3. Настроить [резервное копирование](docs/backup.md)
4. Изучить [API документацию](docs/api_reference.md)
5. Настроить [интеграцию с внешними системами](docs/integration.md)

## Поддержка

Если у вас возникли проблемы или вопросы:

1. Проверьте [руководство по устранению неполадок](docs/troubleshooting.md)
2. Создайте issue в GitHub с подробным описанием проблемы
3. Присоединитесь к обсуждениям в GitHub Discussions

---

**Важно**: Не забудьте изменить пароли по умолчанию перед использованием в продакшене! 