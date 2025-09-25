# Ручная настройка Zitadel проекта

## Доступ к Zitadel UI

1. **Откройте браузер** и перейдите по адресу: `https://zitadel.local/ui/login`
2. **Войдите в систему** используя:
   - Email: `admin@local.test`
   - Password: `jrcCaigzmqzHpRcxo61RDiS+xejxX4/K`

## Создание проекта

1. **Перейдите в Projects** (в левом меню)
2. **Нажмите "New Project"**
3. **Заполните форму:**
   - Project Name: `cloudbridge-project`
   - Description: `CloudBridge Relay Project`
4. **Нажмите "Create"**

## Создание приложения

1. **Выберите созданный проект** `cloudbridge-project`
2. **Перейдите в Applications**
3. **Нажмите "New Application"**
4. **Выберите "OIDC"**
5. **Заполните форму:**
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
6. **Нажмите "Create"**

## Получение credentials

После создания приложения:

1. **Скопируйте Client ID** (отображается в списке приложений)
2. **Нажмите на приложение** для просмотра деталей
3. **Перейдите в "Keys"** или **"Secrets"**
4. **Скопируйте Client Secret**

## Обновление конфигурации

После получения credentials обновите файлы:

### 1. configs/config.json
```json
{
  "zitadel": {
    "enabled": true,
    "domain": "https://zitadel.local",
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

### 2. k8s/configmaps/cloudbridge-config.yaml
Обновите секцию `zitadel` в ConfigMap с теми же значениями.

## Тестирование

После обновления конфигурации:

1. **Перезапустите CloudBridge Relay:**
   ```bash
   kubectl rollout restart deployment/cloudbridge-relay -n cloudbridge
   ```

2. **Проверьте логи:**
   ```bash
   kubectl logs -f deployment/cloudbridge-relay -n cloudbridge
   ```

3. **Протестируйте аутентификацию:**
   ```bash
   curl -k https://zitadel.local/oauth/v2/token \
     -H "Content-Type: application/x-www-form-urlencoded" \
     -d "grant_type=client_credentials&client_id=ВАШ_CLIENT_ID&client_secret=ВАШ_CLIENT_SECRET"
   ```

## Альтернативный способ - через API

Если UI не работает, можно попробовать создать проект через API с правильными headers:

```bash
# Получить CSRF token
CSRF_TOKEN=$(curl -k -s -c cookies.txt https://zitadel.local/ui/login | grep -o 'csrf[^"]*' | head -1)

# Войти в систему
curl -k -s -b cookies.txt -c cookies.txt -X POST https://zitadel.local/ui/login \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -H "X-CSRF-Token: $CSRF_TOKEN" \
  -d "loginName=admin@local.test&password=jrcCaigzmqzHpRcxo61RDiS+xejxX4/K"
```

## Troubleshooting

### Проблема: "Not Found" при API вызовах
- Убедитесь, что используете правильный endpoint
- Проверьте, что Zitadel полностью инициализирован

### Проблема: TLS certificate errors
- Используйте `-k` флаг с curl для игнорирования сертификатов
- Или добавьте `zitadel.local` в `/etc/hosts` с IP `192.168.49.2`

### Проблема: UI не загружается
- Проверьте, что все поды Zitadel работают: `kubectl -n zitadel get pods`
- Проверьте логи: `kubectl -n zitadel logs -l app.kubernetes.io/name=zitadel`



