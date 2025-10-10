# 🔍 SNI Analysis Report: ServerName в серверной TLS конфигурации

## ❌ **Проблема: Неправильное использование ServerName**

### Что было неправильно:
```go
// ❌ НЕПРАВИЛЬНО - ServerName в серверной конфигурации
tlsConfig := &tls.Config{
    Certificates: []tls.Certificate{cert},
    ServerName: "auth.2gc.ru",  // ❌ Это не работает в сервере!
}
```

### Почему это проблема:

1. **ServerName используется только в клиенте**
   - В клиенте: `ServerName` указывает, какой сертификат ожидать
   - В сервере: `ServerName` игнорируется, используется `GetCertificate`

2. **SNI (Server Name Indication) работает по-другому**
   - Клиент отправляет SNI в `ClientHello`
   - Сервер должен выбрать сертификат на основе SNI
   - `ServerName` в серверной конфигурации не влияет на SNI

3. **Проблемы с корпоративными прокси**
   - Прокси проверяют SNI в `ClientHello`
   - Если SNI не совпадает с сертификатом - блокировка
   - Нужен правильный SNI handling

## ✅ **Правильное решение**

### 1. **GetCertificate callback для SNI**
```go
// ✅ ПРАВИЛЬНО - GetCertificate для SNI handling
tlsConfig := &tls.Config{
    Certificates: []tls.Certificate{cert},
    GetCertificate: func(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
        // Логируем SNI запрос
        logger.Debug("SNI request", zap.String("sni", clientHello.ServerName))
        
        // Возвращаем сертификат для любого SNI (для обхода прокси)
        return &cert, nil
    },
}
```

### 2. **GetConfigForClient для динамической конфигурации**
```go
// ✅ ПРАВИЛЬНО - Динамическая конфигурация на основе клиента
tlsConfig := &tls.Config{
    GetConfigForClient: func(clientHello *tls.ClientHelloInfo) (*tls.Config, error) {
        // Логируем информацию о клиенте
        logger.Debug("Client connection",
            zap.String("sni", clientHello.ServerName),
            zap.String("client_ip", clientHello.Conn.RemoteAddr().String()))
        
        // Возвращаем конфигурацию для конкретного клиента
        return &tls.Config{
            Certificates: []tls.Certificate{cert},
            NextProtos:   []string{"h3", "h2", "http/1.1"},
            MinVersion:   tls.VersionTLS12,
        }, nil
    },
}
```

## 🔧 **Технические детали**

### Как работает SNI:

1. **ClientHello с SNI:**
   ```
   Client → Server: ClientHello
   ├─ SNI: "auth.2gc.ru"
   ├─ ALPN: ["h3", "h2", "http/1.1"]
   └─ CipherSuites: [...]
   ```

2. **Server выбирает сертификат:**
   ```go
   GetCertificate: func(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
       // clientHello.ServerName содержит SNI от клиента
       // Возвращаем подходящий сертификат
   }
   ```

3. **ServerHello с сертификатом:**
   ```
   Server → Client: ServerHello
   ├─ Certificate: [сертификат для auth.2gc.ru]
   ├─ ALPN: "h3"
   └─ CipherSuite: TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
   ```

### Корпоративные прокси и SNI:

1. **DPI (Deep Packet Inspection):**
   - Прокси анализируют SNI в ClientHello
   - Блокируют соединения с "подозрительными" SNI
   - Нужен "легитимный" SNI для обхода

2. **SNI Filtering:**
   - Прокси могут блокировать определенные SNI
   - Нужно использовать "белый" домен
   - `auth.2gc.ru` - хороший выбор для обхода

3. **Certificate Validation:**
   - Прокси проверяют соответствие SNI и сертификата
   - Несоответствие = блокировка
   - Нужен валидный сертификат для SNI

## 🚀 **Преимущества исправления**

### 1. **Правильный SNI handling**
- Сервер корректно обрабатывает SNI от клиентов
- Логирование SNI запросов для мониторинга
- Поддержка множественных доменов

### 2. **Обход корпоративных прокси**
- SNI указывает на "легитимный" домен
- Сертификат соответствует SNI
- ALPN протоколы для HTTP/3 маскировки

### 3. **Улучшенная диагностика**
- Логирование всех SNI запросов
- Отслеживание клиентских IP
- Мониторинг поддерживаемых протоколов

## 📊 **Мониторинг и метрики**

### Ключевые метрики:
- **SNI Requests** - количество SNI запросов
- **SNI Distribution** - распределение по доменам
- **Certificate Matches** - успешные совпадения сертификатов
- **TLS Handshake Success** - успешные TLS handshake

### Алерты:
- **Unknown SNI** - неизвестные SNI запросы
- **Certificate Mismatch** - несоответствие сертификата
- **TLS Handshake Failures** - неудачные TLS handshake

## 🔍 **Диагностика проблем**

### 1. **Проверка SNI в логах:**
```bash
# Ищем SNI запросы в логах
grep "SNI request received" /var/log/cloudbridge-relay/relay.log
```

### 2. **Тестирование с curl:**
```bash
# Тест с SNI
curl -v --resolve auth.2gc.ru:443:127.0.0.1 https://auth.2gc.ru:443

# Тест без SNI
curl -v -k https://127.0.0.1:443
```

### 3. **Анализ TLS handshake:**
```bash
# Используем openssl для анализа
openssl s_client -connect auth.2gc.ru:443 -servername auth.2gc.ru
```

## 🎯 **Рекомендации**

### 1. **Использовать GetCertificate вместо ServerName**
- Правильный SNI handling
- Поддержка множественных доменов
- Лучшая диагностика

### 2. **Добавить мониторинг SNI**
- Логирование всех SNI запросов
- Метрики для анализа
- Алерты на аномалии

### 3. **Тестирование с корпоративными прокси**
- Проверка обхода DPI
- Валидация SNI filtering
- Тестирование ALPN протоколов

## 📝 **Заключение**

Исправление `ServerName` на правильный SNI handling критически важно для:

1. **Корректной работы TLS сервера**
2. **Обхода корпоративных прокси**
3. **Правильной диагностики проблем**
4. **Поддержки множественных доменов**

Теперь MASQUE сервер будет правильно обрабатывать SNI запросы и успешно обходить корпоративные firewall'ы.

---
*Анализ выполнен: $(date)*
*Статус: Исправлено*
*Приоритет: Критический*
