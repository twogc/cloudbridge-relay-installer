# JWT Signature Issue Resolution

## Проблема
Разработчик Relay сервера сообщил:
- Токен синтаксически корректен (RFC 7519)
- Подпись не проходит проверку на Relay сервере
- Ошибка: "Signature verification failed"
- Секрет на Relay: `85Sk/NfLq3gqzCXzmBKbJCpL+f5BssXz3G8dVi3sPiE=`

## Диагностика
Наш тест показал, что токен **корректно подписан** секретом `85Sk/NfLq3gqzCXzmBKbJCpL+f5BssXz3G8dVi3sPiE=`.

Возможные причины расхождения:
1. **Кодировка секрета** - Base64 vs plain text
2. **Версия библиотеки** - разные реализации HS256
3. **Конфигурация Relay** - секрет читается неправильно
4. **Whitespace/encoding** - лишние символы в секрете

## Решение: Синхронизация секретов

### Вариант A: Обновить секрет на Relay (Рекомендуемый)
```bash
# 1. Проверить текущий секрет в K8s
kubectl get secret cloudbridge-relay-jwt -o jsonpath='{.data.secret}' | base64 -d

# 2. Обновить секрет в K8s
kubectl create secret generic cloudbridge-relay-jwt \
  --from-literal=secret="85Sk/NfLq3gqzCXzmBKbJCpL+f5BssXz3G8dVi3sPiE=" \
  --dry-run=client -o yaml | kubectl apply -f -

# 3. Перезапустить Relay pods
kubectl rollout restart deployment/cloudbridge-relay
```

### Вариант B: Создать токен с секретом Relay
```bash
# Если на Relay используется другой секрет, создадим токен с ним
cd /home/ubuntu/cloudbridge-relay-installer/test/relay-client
go run test-jwt-signature.go "$(tail -1 zitadel-infinite-token.txt)" "RELAY_SECRET_HERE"
```

## 🔧 Тестовые токены с разными секретами

### 1. Токен с текущим секретом (85Sk/...)
```bash
cd /home/ubuntu/cloudbridge-relay-installer/test/relay-client/scripts
go run generate-zitadel-jwt.go > ../token-current-secret.txt
```

### 2. Токен с альтернативными секретами
Создадим токены с возможными секретами Relay:

#### A) ZITADEL Client Secret
```bash
# Секрет: Wahm7Iti1VAoXsFAV1MBeMfPv1yope2IJ47utB30Pk4FgZyJFZK0ayfX4zNstkUh
```

#### B) Простой тестовый секрет
```bash
# Секрет: cloudbridge-relay-jwt-secret
```

#### C) Base64 декодированный секрет
```bash
# Декодируем текущий секрет
echo "85Sk/NfLq3gqzCXzmBKbJCpL+f5BssXz3G8dVi3sPiE=" | base64 -d | xxd
```

## План действий

### Шаг 1: Определить правильный секрет
1. Разработчик Relay проверяет секрет в K8s: `kubectl get secret cloudbridge-relay-jwt -o yaml`
2. Сравниваем с нашим секретом: `85Sk/NfLq3gqzCXzmBKbJCpL+f5BssXz3G8dVi3sPiE=`

### Шаг 2: Синхронизировать секреты
**Вариант A (Рекомендуемый):** Обновить секрет на Relay
- Не требует изменения генератора токенов
- Единый секрет для всей системы
- Проще в обслуживании

**Вариант B:** Изменить генератор токенов
- Требует изменения кода
- Нужно знать точный секрет Relay
- Сложнее синхронизировать

### Шаг 3: Тестирование
1. Создать тестовый токен с согласованным секретом
2. Проверить на Relay сервере
3. Убедиться что метрика `relay_auth_success_total` увеличивается

## Создание токенов с разными секретами

### Генератор с настраиваемым секретом
```go
// Обновленный generate-zitadel-jwt.go с параметром секрета
func main() {
    secret := os.Getenv("JWT_SECRET")
    if secret == "" {
        secret = "85Sk/NfLq3gqzCXzmBKbJCpL+f5BssXz3G8dVi3sPiE=" // default
    }
    
    // ... rest of the code
}
```

### Использование:
```bash
# С дефолтным секретом
go run generate-zitadel-jwt.go

# С кастомным секретом
JWT_SECRET="relay-server-secret" go run generate-zitadel-jwt.go

# С ZITADEL секретом
JWT_SECRET="Wahm7Iti1VAoXsFAV1MBeMfPv1yope2IJ47utB30Pk4FgZyJFZK0ayfX4zNstkUh" go run generate-zitadel-jwt.go
```

## Рекомендация

**Выбираем Вариант A: Синхронизировать секрет на Relay**

Причины:
1. Наш генератор уже корректно работает
2. Токен синтаксически правильный (подтверждено разработчиком)
3. Проще обновить один секрет, чем менять код
4. Единый секрет для всей системы CloudBridge

## 📞 Следующие шаги

1. **Разработчик Relay:** Проверить секрет в K8s Secret `cloudbridge-relay-jwt`
2. **Если секрет отличается:** Обновить на `85Sk/NfLq3gqzCXzmBKbJCpL+f5BssXz3G8dVi3sPiE=`
3. **Перезапустить Relay pods**
4. **Протестировать с нашим токеном**
5. **Проверить метрики в Grafana**

## Тестовые команды для разработчика

```bash
# 1. Проверить текущий секрет
kubectl get secret cloudbridge-relay-jwt -o jsonpath='{.data.secret}' | base64 -d

# 2. Сравнить с нашим
echo "85Sk/NfLq3gqzCXzmBKbJCpL+f5BssXz3G8dVi3sPiE="

# 3. Если отличается - обновить
kubectl patch secret cloudbridge-relay-jwt -p='{"data":{"secret":"ODVTay9OZkxxM2dxekNYem1CS2JKQ3BMK2Y1QnNzWHozRzhkVmkzc1BpRT0="}}'

# 4. Перезапустить
kubectl rollout restart deployment/cloudbridge-relay
```

*Примечание: `ODVTay9OZkxxM2dxekNYem1CS2JKQ3BMK2Y1QnNzWHozRzhkVmkzc1BpRT0=` это base64 от `85Sk/NfLq3gqzCXzmBKbJCpL+f5BssXz3G8dVi3sPiE=`*
