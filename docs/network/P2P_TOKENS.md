# CloudBridge P2P Tokens - Бессрочные JWT токены

## Токены для локальной сети

### Server A Token
```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJzZXJ2ZXItYSIsInRlbmFudF9pZCI6Im1lc2gtbmV0d29yay10ZXN0Iiwic2VydmVyX2lkIjoic2VydmVyLWEiLCJjb25uZWN0aW9uX3R5cGUiOiJxdWljIiwicHJvdG9jb2xfdHlwZSI6InAycC1tZXNoIiwicGVybWlzc2lvbnMiOlsicmVsYXlfY29ubmVjdCIsInJlbGF5X3R1bm5lbCIsInAycF9kaXNjb3ZlciIsInAycF9jb25uZWN0Il0sImlzcyI6ImNsb3VkYnJpZGdlLXJlbGF5IiwiYXVkIjpbInJlbGF5LXRlc3QtdGVuYW50Il0sIm5iZiI6MTc1OTIxNzM3MywiaWF0IjoxNzU5MjE3MzczfQ.MuOfOmAgKG6v6m6VX7PQX8Dnh9pFwqv5LYTT7HXzogs
```

### Server B Token
```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJzZXJ2ZXItYiIsInRlbmFudF9pZCI6Im1lc2gtbmV0d29yay10ZXN0Iiwic2VydmVyX2lkIjoic2VydmVyLWIiLCJjb25uZWN0aW9uX3R5cGUiOiJxdWljIiwicHJvdG9jb2xfdHlwZSI6InAycC1tZXNoIiwicGVybWlzc2lvbnMiOlsicmVsYXlfY29ubmVjdCIsInJlbGF5X3R1bm5lbCIsInAycF9kaXNjb3ZlciIsInAycF9jb25uZWN0Il0sImlzcyI6ImNsb3VkYnJpZGdlLXJlbGF5IiwiYXVkIjpbInJlbGF5LXRlc3QtdGVuYW50Il0sIm5iZiI6MTc1OTIxNzM3MywiaWF0IjoxNzU5MjE3MzczfQ.359oMf90CwryfCmSEACt2KfdY5vNf2s6apxkIPHpdKY
```

## Характеристики токенов

- **Алгоритм**: HS256
- **Срок действия**: Бессрочные (без ExpiresAt)
- **Аудитория**: relay-test-tenant
- **Издатель**: cloudbridge-relay
- **Секрет**: Wahm7Iti1VAoXsFAV1MBeMfPv1yope2IJ47utB30Pk4FgZyJFZK0ayfX4zNstkUh

## Права доступа

Каждый токен включает следующие разрешения:
- `relay_connect` - подключение к релею
- `relay_tunnel` - создание туннелей через релей
- `p2p_discover` - обнаружение P2P пиров
- `p2p_connect` - установка P2P соединений

## Конфигурация клиентов

### config-test-server-a.yaml
```yaml
auth:
  type: "jwt"
  token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJzZXJ2ZXItYSIsInRlbmFudF9pZCI6Im1lc2gtbmV0d29yay10ZXN0Iiwic2VydmVyX2lkIjoic2VydmVyLWEiLCJjb25uZWN0aW9uX3R5cGUiOiJxdWljIiwicHJvdG9jb2xfdHlwZSI6InAycC1tZXNoIiwicGVybWlzc2lvbnMiOlsicmVsYXlfY29ubmVjdCIsInJlbGF5X3R1bm5lbCIsInAycF9kaXNjb3ZlciIsInAycF9jb25uZWN0Il0sImlzcyI6ImNsb3VkYnJpZGdlLXJlbGF5IiwiYXVkIjpbInJlbGF5LXRlc3QtdGVuYW50Il0sIm5iZiI6MTc1OTIxNzM3MywiaWF0IjoxNzU5MjE3MzczfQ.MuOfOmAgKG6v6m6VX7PQX8Dnh9pFwqv5LYTT7HXzogs"
  secret: "Wahm7Iti1VAoXsFAV1MBeMfPv1yope2IJ47utB30Pk4FgZyJFZK0ayfX4zNstkUh"
  skip_validation: false
```

### config-test-server-b.yaml
```yaml
auth:
  type: "jwt"
  token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJzZXJ2ZXItYiIsInRlbmFudF9pZCI6Im1lc2gtbmV0d29yay10ZXN0Iiwic2VydmVyX2lkIjoic2VydmVyLWIiLCJjb25uZWN0aW9uX3R5cGUiOiJxdWljIiwicHJvdG9jb2xfdHlwZSI6InAycC1tZXNoIiwicGVybWlzc2lvbnMiOlsicmVsYXlfY29ubmVjdCIsInJlbGF5X3R1bm5lbCIsInAycF9kaXNjb3ZlciIsInAycF9jb25uZWN0Il0sImlzcyI6ImNsb3VkYnJpZGdlLXJlbGF5IiwiYXVkIjpbInJlbGF5LXRlc3QtdGVuYW50Il0sIm5iZiI6MTc1OTIxNzM3MywiaWF0IjoxNzU5MjE3MzczfQ.359oMf90CwryfCmSEACt2KfdY5vNf2s6apxkIPHpdKY"
  secret: "Wahm7Iti1VAoXsFAV1MBeMfPv1yope2IJ47utB30Pk4FgZyJFZK0ayfX4zNstkUh"
  skip_validation: false
```

## Настройка сервера

Для использования JWT токенов сервер должен быть настроен:

```bash
# Включить JWT авторизацию
kubectl set env deployment/cloudbridge-relay AUTH_ENABLED=true -n cloudbridge

# Установить JWT секрет
kubectl set env deployment/cloudbridge-relay RELAY_JWT_SECRET="Wahm7Iti1VAoXsFAV1MBeMfPv1yope2IJ47utB30Pk4FgZyJFZK0ayfX4zNstkUh" -n cloudbridge
```

## Тестирование

### Запуск клиентов
```bash
# Server A
cd /home/ubuntu/cloudbridge-relay-installer/test/relay-client
./cloudbridge-client p2p --config config-test-server-a.yaml --log-level debug

# Server B (в другом терминале)
./cloudbridge-client p2p --config config-test-server-b.yaml --log-level debug
```

### Проверка логов сервера
```bash
kubectl logs -n cloudbridge deployment/cloudbridge-relay --tail=20 | grep -E "(peer_id|server_id|allocated)"
```

## Ожидаемый результат

После успешного подключения в логах сервера должно появиться:

```json
{"level":"INFO","msg":"Allocated IP for peer","peer_id":"peer_server-a","allocated_ip":"10.100.77.10"}
{"level":"INFO","msg":"Allocated IP for peer","peer_id":"peer_server-b","allocated_ip":"10.100.77.11"}
```

## Безопасность

- Токены подписаны секретным ключом
- Каждый токен привязан к конкретному серверу (server-a/server-b)
- Токены не имеют срока истечения для тестирования
- В продакшене рекомендуется установить разумный срок действия

## Генерация новых токенов

Для создания новых токенов используйте скрипт:

```bash
cd /home/ubuntu/cloudbridge-relay-installer
go run scripts/generate-p2p-tokens.go
```

## Устранение неполадок

### Ошибка "token is malformed"
- Проверьте, что токен скопирован полностью
- Убедитесь, что в токене нет лишних пробелов или символов

### Ошибка "invalid signature"
- Проверьте, что JWT секрет на сервере совпадает с секретом в токене
- Убедитесь, что `RELAY_JWT_SECRET` установлен правильно

### Ошибка "token expired"
- Токены бессрочные, эта ошибка не должна возникать
- Проверьте системное время на клиенте и сервере
