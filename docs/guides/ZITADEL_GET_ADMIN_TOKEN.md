# Получение Admin токена для ZITADEL API

## Credentials для входа
- **Email**: admin@local.test  
- **Password**: jrcCaigzmqzHpRcxo61RDiS+xejxX4/K

## Цель
Получить токен доступа для управления ZITADEL через Management API для настройки CloudBridge Relay.

## Методы получения токена

### Метод 1: Personal Access Token (рекомендуется)

1. **Откройте консоль ZITADEL**:
   ```
   https://edge.2gc.ru/ui/console
   ```

2. **Войдите в систему**:
   - Email: `admin@local.test`
   - Password: `jrcCaigzmqzHpRcxo61RDiS+xejxX4/K`

3. **Создайте Personal Access Token**:
   - Нажмите на аватар в правом верхнем углу
   - Выберите "Personal Settings"
   - Перейдите на вкладку "Personal Access Tokens"
   - Нажмите "New"
   - Установите срок действия: **1 year**
   - Нажмите "Add"
   - **Скопируйте токен** (показывается только один раз!)

4. **Сохраните токен**:
   ```bash
   mkdir -p ~/.zitadel
   echo "YOUR_TOKEN_HERE" > ~/.zitadel/token
   chmod 600 ~/.zitadel/token
   ```

5. **Или экспортируйте как переменную**:
   ```bash
   export ZITADEL_TOKEN="YOUR_TOKEN_HERE"
   ```

### Метод 2: Machine User (для автоматизации)

1. **Войдите в консоль ZITADEL** (см. шаги выше)

2. **Создайте Machine User**:
   - Organization → Users
   - New → Machine User
   - **Name**: `cloudbridge-admin`
   - **Username**: `cloudbridge-admin`  
   - **Description**: `CloudBridge Relay Admin Machine User`
   - Нажмите "Create"

3. **Создайте ключ для Machine User**:
   - Откройте созданного пользователя
   - Перейдите на вкладку "Keys"
   - Нажмите "New"
   - **Type**: JSON
   - **Expiration**: 1 year
   - Нажмите "Add"
   - **Сохраните JSON файл**

4. **Назначьте права**:
   - Вкладка "Authorizations"  
   - Нажмите "New"
   - Выберите Organization: текущую организацию
   - Роль: `ORG_OWNER` или `IAM_OWNER`
   - Нажмите "Save"

5. **Сохраните ключ**:
   ```bash
   mkdir -p ~/.zitadel
   # Сохраните JSON файл как ~/.zitadel/machine-key.json
   chmod 600 ~/.zitadel/machine-key.json
   ```

### Метод 3: Автоматическое получение через API

Если у вас есть токен, проверьте его работоспособность:

```bash
# Замените YOUR_TOKEN на ваш токен
curl -H "Authorization: Bearer YOUR_TOKEN" \
  https://edge.2gc.ru/auth/v1/users/me
```

Ожидаемый ответ - информация о пользователе.

## Тестирование токена

После получения токена протестируйте его:

```bash
cd /home/ubuntu/cloudbridge-relay-installer

# Установите токен
export ZITADEL_TOKEN="YOUR_TOKEN_HERE"

# Протестируйте доступ к API
curl -s -H "Authorization: Bearer $ZITADEL_TOKEN" \
  https://edge.2gc.ru/auth/v1/users/me | jq '.'

# Проверьте организации
curl -s -H "Authorization: Bearer $ZITADEL_TOKEN" \
  -X POST https://edge.2gc.ru/management/v1/orgs/_search \
  -H "Content-Type: application/json" \
  -d '{"queries": []}' | jq '.result[]? | {id, name}'
```

## Следующие шаги

После получения токена запустите настройку:

```bash
cd /home/ubuntu/cloudbridge-relay-installer

# Убедитесь, что токен установлен
echo $ZITADEL_TOKEN

# Запустите настройку организации и проекта
./scripts/setup-cloudbridge-zitadel.sh
```

##  Troubleshooting

### Ошибка "unauthorized"
- Проверьте правильность токена
- Убедитесь, что токен не истек
- Проверьте права пользователя

### Ошибка "forbidden" 
- Назначьте пользователю роль `ORG_OWNER` или `IAM_OWNER`
- Проверьте scope токена

### Токен не работает
- Создайте новый Personal Access Token
- Проверьте срок действия
- Убедитесь, что копировали токен полностью

## Полезные ссылки

- [ZITADEL Management API](https://zitadel.com/docs/apis/v2.mdx)
- [Personal Access Tokens](https://zitadel.com/docs/guides/manage/console/personal-access-tokens)
- [Machine Users](https://zitadel.com/docs/guides/manage/console/users#machine-users)
