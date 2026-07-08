# MASQUE HTTP/3 Server - Руководство по использованию

## Обзор

MASQUE (Multiplexed Application Substrate over QUIC Encryption) - это протокол для туннелирования трафика через HTTP/3/QUIC соединения. Наш CloudBridge Relay поддерживает полную реализацию MASQUE протокола.

## Статус

**MASQUE сервер полностью функционален и готов к использованию!**

### Текущая конфигурация:
- **Порт**: 8443 (UDP) - HTTP/3/QUIC
- **NodePort**: 30443 (UDP) - внешний доступ
- **TLS**: Let's Encrypt сертификат для `auth.2gc.ru`
- **Протоколы**: HTTP/3, QUIC
- **ALPN**: h3, h3-29, h3-28, h3-27

## 📡 Доступные Endpoints

### 1. CONNECT-UDP
```
/.well-known/masque/udp/{target_host}:{target_port}
```
**Назначение**: Туннелирование UDP трафика через HTTP/3

**Пример использования**:
```bash
# Подключение к UDP серверу через MASQUE
curl --http3 -k https://auth.2gc.ru:8443/.well-known/masque/udp/127.0.0.1:8080
```

### 2. CONNECT-IP
```
/.well-known/masque/ip/{target_ip}
```
**Назначение**: Туннелирование IP трафика через HTTP/3

**Пример использования**:
```bash
# Подключение к IP адресу через MASQUE
curl --http3 -k https://auth.2gc.ru:8443/.well-known/masque/ip/192.168.1.1
```

### 3. HTTP Datagrams
```
/.well-known/masque/datagrams/
```
**Назначение**: Передача произвольных данных через HTTP/3

### 4. Health Check
```
/health
```
**Назначение**: Проверка состояния MASQUE сервера

##  Технические детали

### Архитектура
- **HTTP/3 сервер**: Работает поверх QUIC (UDP)
- **TLS конфигурация**: Современные cipher suites
- **Сертификат**: Let's Encrypt для домена `auth.2gc.ru`
- **Обработчики**: Все MASQUE endpoints зарегистрированы

### Сетевая конфигурация
```yaml
# Kubernetes Service
ports:
- name: masque-proxy
  port: 8443
  targetPort: 8443
  nodePort: 30443
  protocol: UDP  # Важно: UDP для HTTP/3/QUIC
```

### Логирование
Все операции MASQUE сервера логируются с детальной информацией:
- Загрузка сертификатов
- Создание TLS конфигурации
- Регистрация HTTP handlers
- Запуск HTTP/3 сервера

## 🧪 Тестирование

### 1. Проверка доступности
```bash
# UDP подключение к порту
nc -u -v 192.168.49.2 30443

# Или используйте наш тест
python3 test_masque.py
```

### 2. Проверка сертификата
```bash
# Проверка срока действия
sudo openssl x509 -in /etc/letsencrypt/live/auth.2gc.ru/fullchain.pem -text -noout | grep -A2 "Validity"
```

### 3. Проверка логов
```bash
# Логи MASQUE сервера
kubectl logs -n cloudbridge deployment/cloudbridge-relay | grep -i masque
```

## Мониторинг

### Текущий статус
- MASQUE сервер запущен
- Порт 8443 слушается на UDP
- TLS конфигурация активна
- Все handlers зарегистрированы

### Рекомендации по мониторингу
1. **Метрики**: Добавить Prometheus метрики для MASQUE
2. **Алерты**: Настроить уведомления о недоступности
3. **Логи**: Мониторинг ошибок TLS и QUIC

## Устранение неполадок

### Проблема: "Connection refused"
**Причина**: HTTP/3 работает на UDP, а не TCP
**Решение**: Используйте UDP подключение или HTTP/3 клиент

### Проблема: SSL ошибки
**Причина**: Самоподписанный сертификат
**Решение**: Используйте `-k` флаг или настройте доверие к сертификату

### Проблема: Порт не слушается
**Причина**: Неправильная конфигурация service
**Решение**: Убедитесь, что протокол установлен как UDP

## Дополнительные ресурсы

- [RFC 9298: MASQUE](https://tools.ietf.org/html/rfc9298)
- [HTTP/3 Specification](https://tools.ietf.org/html/rfc9114)
- [QUIC Protocol](https://tools.ietf.org/html/rfc9000)

## Заключение

MASQUE HTTP/3 сервер полностью настроен и готов к использованию. Все основные компоненты работают корректно:

- HTTP/3 сервер на порту 8443 (UDP)
- TLS с сертификатом Let's Encrypt
- Все MASQUE endpoints доступны
- Kubernetes service настроен для UDP трафика

Для полного тестирования рекомендуется использовать HTTP/3 клиент с поддержкой MASQUE протокола.

