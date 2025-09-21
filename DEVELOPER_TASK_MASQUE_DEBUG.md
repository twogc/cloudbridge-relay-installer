# 🐛 **Задание разработчику: Диагностика проблемы с MASQUE сервером**

## 📋 **Описание проблемы**

**Статус**: ❌ КРИТИЧНО  
**Приоритет**: ВЫСОКИЙ  
**Компонент**: MASQUE HTTP/3 сервер  
**Порт**: 8443 (NodePort: 30443)  

### **Симптомы**:
- MASQUE сервер запускается (логи показывают успешный старт)
- Порт 8443 не слушается (netstat не показывает порт)
- Connection refused при попытке подключения
- HTTP/3 клиенты не могут подключиться

---

## 🎯 **Цель задания**

Найти и исправить причину, по которой MASQUE HTTP/3 сервер не может привязаться к порту 8443, несмотря на успешный запуск.

---

## 🔍 **Текущее состояние**

### **Что работает** ✅:
- HTTP API (8080) - работает
- P2P API (8082) - работает  
- QUIC Transport (9090) - работает
- Enhanced QUIC (9092) - работает
- STUN Server (19302) - работает
- P2P QUIC (5553) - работает

### **Что не работает** ❌:
- MASQUE HTTP/3 сервер (8443) - не отвечает

---

## 📊 **Данные для анализа**

### **Логи приложения**:
```json
{"level":"INFO","timestamp":"2025-09-21T11:27:47.907Z","caller":"logger/logger.go:132","msg":"MASQUE HTTP/3 server starting","addr":":8443"}
{"level":"INFO","timestamp":"2025-09-21T11:27:47.907Z","caller":"logger/logger.go:132","msg":"MASQUE server started successfully"}
```

### **Порты в контейнере**:
```bash
kubectl exec -it pod -- netstat -tlnp
# Результат: порт 8443 отсутствует
```

### **TLS конфигурация**:
- ✅ Секрет `cloudbridge-tls` существует
- ✅ Файлы `/app/certs/tls.crt` и `/app/certs/tls.key` подключены
- ✅ Переменная `MASQUE_ENABLED=true`

---

## 🔧 **Задачи для выполнения**

### **1. Анализ кода MASQUE сервера** 🔍
**Файл**: `cmd/relay/masque_support.go`

**Задачи**:
- [ ] Проверить функцию `Start()` - строка 204-216
- [ ] Проанализировать горутину запуска HTTP/3 сервера
- [ ] Проверить обработку ошибок в `ListenAndServe()`
- [ ] Убедиться, что ошибки логируются корректно

**Код для анализа**:
```go
go func() {
    addr := ":8443"
    ms.logger.Info("MASQUE HTTP/3 server starting", zap.String("addr", addr))
    
    if err := ms.http3Server.ListenAndServe(); err != nil {
        ms.logger.Error("MASQUE HTTP/3 server error", zap.Error(err))
    }
}()
```

### **2. Проверка TLS конфигурации** 🔐
**Файл**: `cmd/relay/masque_support.go` (функция `createTLSConfig`)

**Задачи**:
- [ ] Проверить загрузку TLS сертификатов
- [ ] Убедиться в корректности путей к файлам
- [ ] Проверить валидность сертификатов
- [ ] Добавить детальное логирование TLS ошибок

**Код для анализа**:
```go
func (ms *MASQUEServer) createTLSConfig() (*tls.Config, error) {
    // Проверить эту функцию
}
```

### **3. Диагностика HTTP/3 сервера** 🌐
**Задачи**:
- [ ] Добавить детальное логирование в HTTP/3 сервер
- [ ] Проверить конфигурацию `http3.Server`
- [ ] Убедиться в корректности ALPN протоколов
- [ ] Проверить совместимость с quic-go версией

### **4. Тестирование TLS сертификатов** 📜
**Команды для выполнения**:
```bash
# Проверить сертификат
kubectl exec -it pod -- cat /app/certs/tls.crt

# Проверить приватный ключ
kubectl exec -it pod -- cat /app/certs/tls.key

# Проверить права доступа
kubectl exec -it pod -- ls -la /app/certs/
```

### **5. Пересоздание TLS секрета** 🔄
**Команды**:
```bash
# Удалить старый секрет
kubectl delete secret cloudbridge-tls -n cloudbridge

# Создать новый секрет (если есть локальные сертификаты)
kubectl create secret tls cloudbridge-tls \
  --cert=path/to/tls.crt \
  --key=path/to/tls.key \
  -n cloudbridge

# Или сгенерировать самоподписанный сертификат
openssl req -x509 -newkey rsa:4096 -keyout tls.key -out tls.crt -days 365 -nodes
```

---

## 🧪 **План тестирования**

### **1. Добавить детальное логирование** 📝
```go
// В функции Start() добавить:
ms.logger.Info("Creating TLS config...")
tlsConfig, err := ms.createTLSConfig()
if err != nil {
    ms.logger.Error("Failed to create TLS config", zap.Error(err))
    return err
}
ms.logger.Info("TLS config created successfully")

ms.logger.Info("Creating HTTP/3 server...")
ms.http3Server = &http3.Server{
    Handler: ms,
    TLSConfig: tlsConfig,
}
ms.logger.Info("HTTP/3 server created successfully")
```

### **2. Проверить ошибки привязки порта** 🔌
```go
// Добавить в горутину:
go func() {
    addr := ":8443"
    ms.logger.Info("MASQUE HTTP/3 server starting", zap.String("addr", addr))
    
    // Добавить проверку порта
    listener, err := net.Listen("tcp", addr)
    if err != nil {
        ms.logger.Error("Failed to bind to port", zap.String("addr", addr), zap.Error(err))
        return
    }
    ms.logger.Info("Successfully bound to port", zap.String("addr", addr))
    listener.Close()
    
    if err := ms.http3Server.ListenAndServe(); err != nil {
        ms.logger.Error("MASQUE HTTP/3 server error", zap.Error(err))
    }
}()
```

### **3. Тестирование с разными портами** 🔄
```go
// Попробовать разные порты для диагностики:
// 8444, 9443, 10443
```

---

## 🎯 **Ожидаемые результаты**

### **После исправления**:
- [ ] MASQUE сервер успешно привязывается к порту 8443
- [ ] `netstat -tlnp` показывает порт 8443 в состоянии LISTEN
- [ ] HTTP/3 клиенты могут подключиться к `https://192.168.49.2:30443`
- [ ] MASQUE endpoints отвечают на запросы

### **Критерии успеха**:
```bash
# Должно работать:
curl -k https://192.168.49.2:30443/.well-known/masque/udp/127.0.0.1:8080

# netstat должен показывать:
tcp6 0 0 :::8443 :::* LISTEN 1/cloudbridge-relay
```

---

## 🚨 **Возможные причины (для исследования)**

### **1. TLS сертификаты** 🔐
- Невалидные сертификаты
- Неправильные пути к файлам
- Проблемы с правами доступа
- Истекшие сертификаты

### **2. HTTP/3 конфигурация** 🌐
- Неправильная конфигурация `http3.Server`
- Проблемы с ALPN протоколами
- Несовместимость версий quic-go
- Ошибки в TLS конфигурации

### **3. Сетевые проблемы** 🌍
- Конфликт портов
- Проблемы с привязкой к IPv6
- Блокировка файрволом
- Проблемы с Kubernetes networking

### **4. Код приложения** 💻
- Ошибки в горутине запуска
- Неправильная обработка ошибок
- Проблемы с контекстом
- Race conditions

---

## 📋 **Чек-лист выполнения**

### **Диагностика**:
- [ ] Проанализировать код MASQUE сервера
- [ ] Проверить TLS конфигурацию
- [ ] Добавить детальное логирование
- [ ] Протестировать TLS сертификаты
- [ ] Проверить порты в контейнере

### **Исправление**:
- [ ] Исправить найденные проблемы
- [ ] Пересоздать TLS секрет (если нужно)
- [ ] Перезапустить deployment
- [ ] Протестировать MASQUE endpoints

### **Валидация**:
- [ ] MASQUE сервер отвечает на порту 8443
- [ ] HTTP/3 клиенты могут подключиться
- [ ] Все тесты проходят
- [ ] Документация обновлена

---

## 📞 **Контакты и ресурсы**

### **Релевантные файлы**:
- `cmd/relay/masque_support.go` - основной код MASQUE
- `k8s/cloudbridge-relay-external-service.yaml` - конфигурация сервиса
- `k8s/deployment.yaml` - конфигурация deployment

### **Полезные команды**:
```bash
# Логи приложения
kubectl logs deployment/cloudbridge-relay -n cloudbridge -f

# Проверка портов
kubectl exec -it pod -- netstat -tlnp

# Тестирование MASQUE
curl -k https://192.168.49.2:30443/.well-known/masque/udp/127.0.0.1:8080
```

---

## ⏰ **Временные рамки**

- **Диагностика**: 2-4 часа
- **Исправление**: 1-2 часа  
- **Тестирование**: 1 час
- **Общее время**: 4-7 часов

---

**🎯 Цель**: Восстановить полную функциональность MASQUE HTTP/3 сервера для достижения 100% работоспособности CloudBridge Relay.

---

*Задание создано: 21 сентября 2025*  
*Статус: Ожидает выполнения*  
*Приоритет: ВЫСОКИЙ*
