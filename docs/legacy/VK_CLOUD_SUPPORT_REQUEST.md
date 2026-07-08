# Обращение в техподдержку VK Cloud - LoadBalancer Issue

**Дата:** 28 сентября 2025  
**Проект:** CloudBridge Relay Infrastructure  
**Приоритет:** Высокий (блокирует продакшен)  

## Краткое описание проблемы

LoadBalancer в VK Cloud не может корректно обрабатывать запросы к backend серверам, несмотря на правильную конфигурацию всех компонентов. Все members постоянно находятся в статусе ERROR, LoadBalancer возвращает 503 Service Unavailable.

## Детали LoadBalancer

**LoadBalancer ID:** `95bd8ca8-b30d-461a-9206-eb04c403a544`  
**Имя:** LoadBalancer_6774  
**Внешний IP:** 109.120.181.108  
**Внутренний IP:** 192.168.49.8  
**Подсеть:** ef2d28b4-7bfa-45b7-b5b3-2d0f808f7ae9 (kubernetes-subnet)  
**Статус LoadBalancer:** ONLINE  
**Провайдер:** amphora  

### Конфигурация компонентов

**Listener:** 34a32e43-8759-46bb-b40d-8a18dafdc0d8
- Протокол: TERMINATED_HTTPS
- Порт: 443
- SSL сертификат: https://public.infra.mail.ru:9311/v1/containers/6dc1f159-7018-4b6c-bb7f-c3d0b34fcc65
- Статус: ONLINE

**Pool:** dc5b3293-bc5d-471a-8584-f0bf913bd27f
- Протокол: HTTP
- Алгоритм: ROUND_ROBIN
- Статус: ONLINE

**Health Monitor:** b4b3126a-8d19-4c55-ba8c-6ce948cf1474
- Тип: HTTP
- Метод: GET
- Путь: /health
- Ожидаемый код: 200
- Таймаут: 10 секунд
- Интервал: 30 секунд
- Статус: ONLINE

## Проблема с Members

### Тестировали два разных backend сервера:

**Member 1 (Kubernetes):**
- ID: 3c4d2107-3317-4c52-b9ad-c316058534cf (удален)
- Адрес: 192.168.49.2:32500
- Подсеть: ef2d28b4-7bfa-45b7-b5b3-2d0f808f7ae9 (та же что у LoadBalancer)
- Статус: ERROR

**Member 2 (Nginx Proxy):**
- ID: 2cf7e7f0-e119-43df-88b5-2e49f0dacc8b
- Адрес: 192.168.49.10:8080
- Подсеть: ef2d28b4-7bfa-45b7-b5b3-2d0f808f7ae9 (та же что у LoadBalancer)
- Статус: ERROR

**Оба member находятся в статусе ERROR, несмотря на корректную работу backend серверов.**

## Доказательства корректной работы Backend серверов

### Backend 1 (Kubernetes) - Работает идеально:
```bash
curl http://192.168.49.2:32500/health
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Content-Length: 350

{"status":"healthy","timestamp":"2025-09-28T10:59:50.533253709Z","services":{"database":"healthy","redis":"healthy","websocket":"healthy"},"metrics":{"active_connections":0,"goroutines":40,"memory_usage_mb":4,"online_peers":0,"total_peers":0,"uptime_seconds":5978.571439343,"websocket_connections":0},"version":"v1.0.0","uptime":"1h39m38.571439343s"}
```

### Backend 2 (Nginx Proxy) - Работает идеально:
```bash
curl http://192.168.49.10:8080/health
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8

{"status":"healthy","timestamp":"...","version":"v1.0.0",...}
```

**Оба backend сервера:**
- Возвращают HTTP 200 OK
- твечают в формате JSON
- оступны по сети из подсети LoadBalancer
- меют корректные заголовки Content-Type

## Сетевая связность

### Проверка связности из подсети LoadBalancer:
```bash
# Ping к backend серверам
ping 192.168.49.2   # SUCCESS - 0% packet loss
ping 192.168.49.10  # SUCCESS - 0% packet loss

# Traceroute показывает прямую связность
# Нет проблем с маршрутизацией
```

### Конфигурация сети:
- **Сеть:** network_1740 (79dd5769-5888-4bd9-bda7-78d6daa944d5)
- **Подсеть LoadBalancer:** 192.168.49.0/24 (ef2d28b4-7bfa-45b7-b5b3-2d0f808f7ae9)
- **Подсеть Backend:** 192.168.49.0/24 (та же самая)
- **Роутер:** router_6328 (a894c312-c98c-40ea-ba4f-86c6963e1e99) - ACTIVE

**LoadBalancer и все backend серверы находятся в одной подсети - нет проблем с маршрутизацией.**

## SSL/TLS конфигурация

### SSL Handshake работает корректно:
```bash
openssl s_client -connect 109.120.181.108:443 -servername relay.2gc.ru

CONNECTED(00000005)
Certificate chain
 0 s:C = RU, ST = Moscow, L = Moscow, O = 2GC, CN = *.2gc.ru
   i:C = RU, ST = Moscow, L = Moscow, O = 2GC, CN = *.2gc.ru
   
Server certificate: *.2gc.ru
Validity: Sep 28 09:54:41 2025 GMT - Sep 28 09:54:41 2026 GMT
SSL handshake: SUCCESS
```

**SSL termination работает корректно, сертификат валидный.**

## Симптомы проблемы

### Поведение LoadBalancer:
1. **SSL Handshake:** роходит успешно
2. **HTTP запросы:**  Возвращает 503 Service Unavailable
3. **Время ответа:** Быстрое (не таймаут)
4. **Member статус:** Всегда ERROR независимо от backend

### Тестирование через LoadBalancer:
```bash
curl -k -H "Host: relay.2gc.ru" https://109.120.181.108/health

<html><body><h1>503 Service Unavailable</h1>
No server is available to handle this request.
</body></html>
```

**LoadBalancer считает, что нет доступных серверов, хотя оба backend работают.**

## Проведенные тесты и исключения

### сключили проблемы с:
1. **Сетевой конфигурацией** - все маршруты работают
2. **Backend серверами** - оба отвечают HTTP 200
3. **SSL сертификатами** - handshake проходит
4. **Security Groups** - все порты открыты
5. **DNS записями** - корректно указывают на LoadBalancer
6. **Health Monitor конфигурацией** - настроен правильно

###  Проблема локализована в:
**VK Cloud LoadBalancer service не может корректно выполнять health checks к member серверам**

## Дополнительная информация

### Security Groups:
- **LoadBalancer SG:** 05a4f11e-9b95-4e33-a273-8e85dd1ad8f6 (не найден - возможная проблема)
- **Backend SG:** 811039da-7d5a-4986-84d0-cc8eefc67e69 (работает корректно)

### Порты в Security Group backend серверов:
- Port 22 (SSH): ткрыт
- Port 80 (HTTP): ткрыт  
- Port 443 (HTTPS): ткрыт
- Port 8080 (Nginx): ткрыт
- Port 32500 (Kubernetes): ткрыт

### LoadBalancer Port статус:
- **Port ID:** 651156ce-c733-4084-8f2a-e7d45c1bbcf9
- **Admin State:** DOWN (техподдержка подтвердила, что это нормально)
- **Status:** DOWN (техподдержка подтвердила, что это нормально)

## Запрос к техподдержке

### Необходимые действия:
1. **Проверить amphora instance** для LoadBalancer 95bd8ca8-b30d-461a-9206-eb04c403a544
2. **Проверить логи health checks** - почему members в статусе ERROR
3. **Проверить внутреннюю маршрутизацию** LoadBalancer к member серверам
4. **Проверить Security Group** 05a4f11e-9b95-4e33-a273-8e85dd1ad8f6 (не найден)

### Вопросы для инженерной команды:
1. Почему LoadBalancer не может достичь member серверов в той же подсети?
2. Почему health monitor показывает ONLINE, но members в ERROR?
3. Есть ли проблемы с amphora instance для этого LoadBalancer?
4. Корректно ли создан Security Group для LoadBalancer?

## Влияние на бизнес

**Критическое:** Блокирует запуск продакшен среды CloudBridge Relay.  
**Обходное решение:** Используем прямое подключение к backend серверам.  
**Требуется:** Исправление LoadBalancer service для внешнего доступа.  

## Контактная информация

**Проект ID:** 8530029fa7b64199b2ad87c18d4620df  
**Пользователь:** info@2gc.ru  
**Регион:** RegionOne  

## Приложения

Все конфигурационные файлы, логи тестирования и детальная диагностика доступны по запросу.

**Просим приоритетного рассмотрения данной проблемы инженерной командой VK Cloud.**
