# Краткое руководство по получению токена для Server Client

## 🎯 Быстрый старт

### Через Keycloak (рекомендуется):

1. **Войдите в Keycloak Admin Console**: `http://your-keycloak-server:8080`
2. **Создайте пользователя** для Server Client с ролью `relay-server-client`
3. **Получите токен через API**:
   ```bash
   curl -X POST http://keycloak:8080/realms/cloudbridge/protocol/openid-connect/token \
     -H "Content-Type: application/x-www-form-urlencoded" \
     -d "grant_type=password" \
     -d "client_id=relay-client" \
     -d "client_secret=your-client-secret" \
     -d "username=server-client-user" \
     -d "password=your-password"
   ```

### Для разработки (локальная генерация):

```python
import jwt
import datetime

payload = {
    'client_id': 'test-server-client',
    'exp': datetime.datetime.utcnow() + datetime.timedelta(hours=24),
    'iat': datetime.datetime.utcnow(),
    'permissions': ['tunnel_create', 'tunnel_read']
}

token = jwt.encode(payload, 'your-jwt-secret', algorithm='HS256')
print(f"Generated token: {token}")
```

### Конфигурация Server Client:

```json
{
  "auth": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_source": "keycloak",
    "refresh_enabled": true
  }
}
```

**📖 Подробное руководство**: [Server Client Token Guide](docs/server_client_token_guide.md) 