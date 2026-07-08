# Инструкция по доступу к Zitadel

## Проброс портов настроен!

### Доступ к Zitadel UI

**URL:** http://zitadel.local:8080/ui/login

**Учетные данные:**
- **Email:** `admin@local.test`
- **Password:** `jrcCaigzmqzHpRcxo61RDiS+xejxX4/K`

###  Доступ к Zitadel API

**Base URL:** http://zitadel.local:8080

**OIDC Endpoints:**
- Discovery: http://zitadel.local:8080/.well-known/openid-configuration
- JWKS: http://zitadel.local:8080/oauth/v2/keys
- Token: http://zitadel.local:8080/oauth/v2/token
- UserInfo: http://zitadel.local:8080/oidc/v1/userinfo

## Следующие шаги

### 1. Создать проект в Zitadel UI

1. **Откройте браузер** и перейдите по адресу: http://zitadel.local:8080/ui/login
2. **Войдите в систему** используя учетные данные выше
3. **Создайте проект:**
   - Название: `cloudbridge-project`
   - Описание: `CloudBridge Relay Project`

### 2. Создать OIDC приложение

1. **Выберите созданный проект** `cloudbridge-project`
2. **Перейдите в Applications** → **New Application**
3. **Выберите OIDC** и заполните:
   - Application Name: `cloudbridge-relay`
   - Application Type: `Web`
   - Auth Method Type: `Basic`
   - Access Token Type: `Bearer`
   - Grant Types: `Authorization Code`
   - Response Types: `Code`
   - Redirect URIs:
     - `http://localhost:8080/callback`
     - `https://localhost:8443/callback`
   - Post Logout Redirect URIs:
     - `http://localhost:8080/`
     - `https://localhost:8443/`

### 3. Получить credentials

После создания приложения:
1. **Скопируйте Client ID** (отображается в списке)
2. **Нажмите на приложение** для просмотра деталей
3. **Скопируйте Client Secret** из секции Keys/Secrets

### 4. Обновить конфигурацию

После получения credentials обновите файлы:

#### configs/config.json
```json
{
  "zitadel": {
    "enabled": true,
    "domain": "http://zitadel.local:8080",
    "project_id": "ВАШ_PROJECT_ID",
    "client_id": "ВАШ_CLIENT_ID", 
    "client_secret": "ВАШ_CLIENT_SECRET",
    "fallback_secret": "85Sk/NfLq3gqzCXzmBKbJCpL+f5BssXz3G8dVi3sPiE=",
    "token_validation": {
      "audience": "cloudbridge-relay",
      "clock_skew": "30s"
    },
    "cache_ttl": "5m",
    "http_timeout": "30s",
    "tenant_management": {
      "enabled": true,
      "auto_create": false,
      "default_quotas": {
        "max_peers": 10,
        "max_connections": 100,
        "max_tunnels": 20,
        "bandwidth_limit": "1GB",
        "storage_limit": "10GB"
      }
    }
  }
}
```

#### k8s/configmaps/cloudbridge-config.yaml
Обновите секцию `zitadel` в ConfigMap с теми же значениями.

## 🛠️ Управление пробросом портов

Используйте скрипт `scripts/port-forward.sh`:

```bash
# Запустить Zitadel port-forwards
./scripts/port-forward.sh start

# Проверить статус
./scripts/port-forward.sh status

# Остановить все
./scripts/port-forward.sh stop

# Запустить все (Zitadel + CloudBridge)
./scripts/port-forward.sh all
```

## Тестирование

После обновления конфигурации:

```bash
# Перезапустить CloudBridge Relay
kubectl rollout restart deployment/cloudbridge-relay -n cloudbridge

# Проверить логи
kubectl logs -f deployment/cloudbridge-relay -n cloudbridge

# Тестировать аутентификацию
curl -k http://zitadel.local:8080/oauth/v2/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials&client_id=ВАШ_CLIENT_ID&client_secret=ВАШ_CLIENT_SECRET"
```

##  Важные замечания

1. **Домен:** Используйте `zitadel.local:8080`, а не `localhost:8080`
2. **HTTPS:** В development режиме используйте HTTP, в production - HTTPS
3. **Сертификаты:** Self-signed сертификаты игнорируются с флагом `-k`
4. **Порты:** Убедитесь, что порты 8080 и 3000 не заняты другими процессами

## 🆘 Troubleshooting

### Проблема: "Экземпляр не найден"
- Убедитесь, что используете `zitadel.local:8080`, а не `localhost:8080`
- Проверьте, что запись `127.0.0.1 zitadel.local` есть в `/etc/hosts`

### Проблема: UI не загружается
- Проверьте, что port-forward активен: `./scripts/port-forward.sh status`
- Проверьте логи Zitadel: `kubectl -n zitadel logs -l app.kubernetes.io/name=zitadel`

### Проблема: Не могу войти в систему
- Убедитесь, что используете правильные учетные данные
- Проверьте, что Zitadel полностью инициализирован



