# Ответ разработчику: JWT Signature Issue

## Анализ проблемы

Спасибо за детальную диагностику! Вы правильно определили:
- Токен синтаксически корректен (RFC 7519)
- Формат: 3 сегмента, HS256, правильные claims
- Подпись не проходит проверку на Relay сервере

## Наше решение: Вариант A - Синхронизировать секрет

**Рекомендуем обновить секрет на Relay сервере** на наш стандартный:
```
85Sk/NfLq3gqzCXzmBKbJCpL+f5BssXz3G8dVi3sPiE=
```

### Причины выбора:
1. Наш генератор уже корректно работает
2. Токен RFC 7519 совместимый (подтверждено вами)
3. Проще обновить один секрет, чем менять весь код генерации
4. Единый секрет для всей экосистемы CloudBridge

## Команды для обновления секрета в K8s

### 1. Проверить текущий секрет
```bash
kubectl get secret cloudbridge-relay-jwt -o jsonpath='{.data.secret}' | base64 -d
```

### 2. Обновить секрет
```bash
# Вариант A: Patch существующий secret
kubectl patch secret cloudbridge-relay-jwt -p='{"data":{"secret":"ODVTay9OZkxxM2dxekNYem1CS2JKQ3BMK2Y1QnNzWHozRzhkVmkzc1BpRT0="}}'

# Вариант B: Создать новый secret
kubectl create secret generic cloudbridge-relay-jwt \
  --from-literal=secret="85Sk/NfLq3gqzCXzmBKbJCpL+f5BssXz3G8dVi3sPiE=" \
  --dry-run=client -o yaml | kubectl apply -f -
```

*Примечание: `ODVTay9OZkxxM2dxekNYem1CS2JKQ3BMK2Y1QnNzWHozRzhkVmkzc1BpRT0=` это base64 от `85Sk/NfLq3gqzCXzmBKbJCpL+f5BssXz3G8dVi3sPiE=`*

### 3. Перезапустить Relay pods
```bash
kubectl rollout restart deployment/cloudbridge-relay
```

## Тестовые токены для проверки

Мы создали тестовые токены с разными секретами:

### 1. Токен с нашим стандартным секретом (рекомендуемый)
**Секрет:** `85Sk/NfLq3gqzCXzmBKbJCpL+f5BssXz3G8dVi3sPiE=`
**Файл:** `test/relay-client/token-default-secret.txt`
**Статус:** Подпись проверена

### 2. Токен с ZITADEL Client Secret (альтернатива)
**Секрет:** `Wahm7Iti1VAoXsFAV1MBeMfPv1yope2IJ47utB30Pk4FgZyJFZK0ayfX4zNstkUh`
**Файл:** `test/relay-client/token-zitadel-secret.txt`
**Статус:** Подпись проверена

### 3. Токен с простым секретом (для тестов)
**Секрет:** `cloudbridge-relay-jwt-secret`
**Файл:** `test/relay-client/token-simple-secret.txt`
**Статус:** Подпись проверена

## План тестирования

### Шаг 1: Обновить секрет на Relay
```bash
kubectl patch secret cloudbridge-relay-jwt -p='{"data":{"secret":"ODVTay9OZkxxM2dxekNYem1CS2JKQ3BMK2Y1QnNzWHozRzhkVmkzc1BpRT0="}}'
kubectl rollout restart deployment/cloudbridge-relay
```

### Шаг 2: Протестировать с нашим токеном
```bash
# Использовать токен из zitadel-infinite-token.txt
export RELAY_TOKEN="$(tail -1 /path/to/zitadel-infinite-token.txt)"
./cloudbridge-client -c config-minikube-grpc.yaml
```

### Шаг 3: Проверить метрики
```bash
# В Grafana должна увеличиться метрика
relay_auth_success_total
```

## Готовые токены для тестирования

### Долгосрочный токен (10 лет)
**Файл:** `test/relay-client/zitadel-infinite-token.txt`
**Использование:**
```bash
export RELAY_TOKEN="$(tail -1 zitadel-infinite-token.txt)"
./cloudbridge-client version # Тест
```

### Краткосрочный токен (15 минут)
**Генерация:**
```bash
cd test/relay-client/scripts
go run generate-jwt-with-custom-secret.go
```

## Дополнительные инструменты

### Генератор с кастомным секретом
```bash
# С дефолтным секретом
go run generate-jwt-with-custom-secret.go

# С кастомным секретом
JWT_SECRET="your-secret-here" go run generate-jwt-with-custom-secret.go

# Долгосрочный токен
go run generate-jwt-with-custom-secret.go "your-secret" long
```

### Тестер подписи
```bash
go run test-jwt-signature.go "TOKEN_HERE" "SECRET_HERE"
```

## Ожидаемый результат

После обновления секрета на Relay:
1. Токен из `zitadel-infinite-token.txt` должен проходить проверку
2. Клиент не должен получать "token is malformed"
3. Метрика `relay_auth_success_total` должна увеличиваться
4. Аутентификация должна работать корректно

## Альтернативный вариант (если нужен другой секрет)

Если по каким-то причинам нельзя изменить секрет на Relay, сообщите нам точный секрет из K8s, и мы пересоздадим токены:

```bash
# Получить точный секрет
kubectl get secret cloudbridge-relay-jwt -o jsonpath='{.data.secret}' | base64 -d

# Создать токен с этим секретом
JWT_SECRET="RELAY_SECRET_HERE" go run generate-jwt-with-custom-secret.go long
```

## Следующие шаги

1. **Обновите секрет в K8s** на `85Sk/NfLq3gqzCXzmBKbJCpL+f5BssXz3G8dVi3sPiE=`
2. **Перезапустите Relay pods**
3. **Протестируйте с токеном** из `zitadel-infinite-token.txt`
4. **Проверьте метрики** в Grafana
5. **Сообщите результат** - должно работать без ошибок!

---

**Готовы к тестированию! Токены RFC 7519 совместимые и правильно подписанные.**
