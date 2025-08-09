# 🔐 Настройка GitHub Secrets для CloudBridge Relay

## Обязательные секреты

### 1. Personal Access Token (GITHUB_TOKEN)

#### Создание токена:
1. Перейдите на https://github.com/settings/tokens
2. Нажмите **Generate new token** → **Generate new token (classic)**
3. Укажите имя: `CloudBridge Relay CI/CD`
4. Выберите права (scopes):
   - ✅ **repo** (Full control of private repositories)
   - ✅ **workflow** (Update GitHub Action workflows) 
   - ✅ **write:packages** (Upload packages to GitHub Package Registry)
   - ✅ **read:packages** (Download packages from GitHub Package Registry)
5. Нажмите **Generate token**
6. **ВАЖНО**: Скопируйте токен сразу! Он больше не будет показан.

#### Добавление в GitHub:
1. Откройте https://github.com/twogc/cloudbridge-relay-installer/settings/secrets/actions
2. Нажмите **New repository secret**
3. Имя: `GH_TOKEN` (не GITHUB_TOKEN!)
4. Значение: вставьте скопированный токен
5. Нажмите **Add secret**

### 2. Docker Hub токен (DOCKERHUB_TOKEN)

#### Создание токена Docker Hub:
1. Войдите на https://hub.docker.com
2. Перейдите в **Account Settings** → **Security**
3. Нажмите **New Access Token**
4. Имя: `CloudBridge Relay CI/CD`
5. Права: **Read, Write, Delete**
6. Нажмите **Generate**
7. Скопируйте токен

#### Добавление в GitHub:
1. В GitHub Secrets добавьте:
   - Имя: `DOCKERHUB_USERNAME`
   - Значение: `twogc`
2. Добавьте второй секрет:
   - Имя: `DOCKERHUB_TOKEN` 
   - Значение: токен от Docker Hub

### 3. Дополнительные секреты (опционально)

#### GPG ключ для подписи релизов:
```
Имя: GPG_PRIVATE_KEY
Значение: Приватный GPG ключ (в формате ASCII armor)

Имя: GPG_PASSPHRASE  
Значение: Пароль от GPG ключа
```

#### Slack уведомления:
```
Имя: SLACK_WEBHOOK_URL
Значение: URL webhook для уведомлений в Slack
```

#### Telegram уведомления:
```
Имя: TELEGRAM_BOT_TOKEN
Значение: Токен Telegram бота

Имя: TELEGRAM_CHAT_ID
Значение: ID чата для уведомлений
```

## Проверка секретов

После добавления секретов вы должны увидеть:
- ✅ GH_TOKEN
- ✅ DOCKERHUB_USERNAME  
- ✅ DOCKERHUB_TOKEN

## Тестирование

Создайте тестовый workflow для проверки:

```yaml
name: Test Secrets
on: workflow_dispatch

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - name: Test GitHub Token
      run: |
        curl -H "Authorization: token ${{ secrets.GH_TOKEN }}" \
        https://api.github.com/user
        
    - name: Test Docker Hub
      run: |
        echo ${{ secrets.DOCKERHUB_TOKEN }} | docker login -u ${{ secrets.DOCKERHUB_USERNAME }} --password-stdin
```

## Безопасность

- ❌ Никогда не выводите секреты в логи
- ✅ Используйте secrets только в доверенных workflows
- ✅ Регулярно обновляйте токены
- ✅ Удаляйте неиспользуемые секреты
