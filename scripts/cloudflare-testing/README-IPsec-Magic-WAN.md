# 🔐 IPsec L3-туннель для Cloudflare Magic WAN

Этот документ описывает правильную настройку L3-туннеля (IPsec IKEv2) для подключения к Cloudflare Magic WAN.

## 🎯 Архитектура

```
[CloudBridge Relay] ←→ [IPsec IKEv2 Tunnel] ←→ [Cloudflare Magic WAN]
     (L3)                    (L3)                    (L3)
```

**Важно:** Magic WAN - это L3-сервис (IPsec/GRE), а не L7-туннель (cloudflared).

## 📋 Необходимые параметры от Cloudflare

Перед началом настройки получите от вашего TAM/SE Cloudflare:

### 🔑 Обязательные параметры:
- **CF_PEER_IP** - IP адрес Cloudflare endpoint для Magic WAN
- **CF_PSK** - Pre-Shared Key для аутентификации
- **LOCAL_PUBLIC_IP** - Ваш публичный IP адрес

### 🔧 Дополнительные параметры:
- **Тип туннеля**: IPsec IKEv2 (рекомендуется) или GRE
- **Внутренние сети**: Какие сети объявлять в Magic WAN
- **BGP параметры**: ASN, соседство (если используется BGP)
- **Политики**: WAF, DDoS, фильтрация

## 🚀 Быстрый старт

### 1. Установка переменных окружения

```bash
# Получите эти значения от Cloudflare
export CF_PEER_IP="<PEER_IP_FROM_CLOUDFLARE>"
export CF_PSK="<PSK_FROM_CLOUDFLARE>"
export LOCAL_PUBLIC_IP="<YOUR_PUBLIC_IP>"
```

### 2. Развертывание IPsec туннеля

```bash
cd /home/ubuntu/cloudbridge-relay-installer
./scripts/cloudflare-testing/deploy-ipsec-magic-wan.sh
```

### 3. Тестирование туннеля

```bash
./scripts/cloudflare-testing/test-ipsec-magic-wan.sh
```

## 🔧 Настройка файрвола

На вашем гейтвее/файрволе откройте порты:

### Для IPsec IKEv2:
```bash
# UDP порты для IKE
ufw allow 500/udp
ufw allow 4500/udp

# ESP протокол (если не используется NAT-T)
ufw allow esp
```

### Для GRE (альтернатива):
```bash
# GRE протокол
ufw allow gre
```

## 📊 Мониторинг и диагностика

### Проверка статуса туннеля:
```bash
# Статус IPsec соединений
kubectl exec -n cloudbridge-testing <pod-name> -- ipsec statusall

# Активные соединения
kubectl exec -n cloudbridge-testing <pod-name> -- ipsec listall

# Маршруты через туннель
kubectl exec -n cloudbridge-testing <pod-name> -- ip route show
```

### Логи и отладка:
```bash
# Логи strongSwan
kubectl exec -n cloudbridge-testing <pod-name> -- tail -f /var/log/strongswan.log

# Логи пода
kubectl logs -n cloudbridge-testing <pod-name>
```

## 🧪 Тестирование связности

### Базовые тесты:
```bash
# Ping через туннель
kubectl exec -n cloudbridge-testing <pod-name> -- ping -c 3 8.8.8.8

# HTTP запрос через туннель
kubectl exec -n cloudbridge-testing <pod-name> -- curl -s https://httpbin.org/ip
```

### Продвинутые тесты:
```bash
# Тест пропускной способности
kubectl exec -n cloudbridge-testing <pod-name> -- iperf3 -c iperf.he.net

# Тест маршрутизации
kubectl exec -n cloudbridge-testing <pod-name> -- mtr -r -c 5 8.8.8.8
```

## 🔍 Устранение неполадок

### Туннель не устанавливается:
1. Проверьте параметры CF_PEER_IP, CF_PSK
2. Убедитесь, что порты UDP 500/4500 открыты
3. Проверьте логи strongSwan
4. Обратитесь к Cloudflare для проверки конфигурации

### Туннель установлен, но нет трафика:
1. Проверьте маршрутизацию
2. Убедитесь, что внутренние сети объявлены в Magic WAN
3. Проверьте политики в Cloudflare Dashboard

### Проблемы с производительностью:
1. Проверьте MTU туннеля
2. Настройте сжатие (если нужно)
3. Оптимизируйте параметры IKE/ESP

## 📈 Настройка в Cloudflare Dashboard

После успешного установления туннеля:

1. **Magic WAN → Sites** - проверьте, что ваш сайт активен
2. **Advertised Prefixes** - добавьте ваши внутренние сети
3. **Policies** - настройте правила трафика
4. **Analytics** - мониторьте статистику

## 🔐 Безопасность

- Используйте сильные PSK (минимум 32 символа)
- Регулярно ротируйте ключи
- Мониторьте логи на предмет подозрительной активности
- Настройте DPD (Dead Peer Detection)

## 📚 Дополнительные ресурсы

- [Cloudflare Magic WAN Documentation](https://developers.cloudflare.com/magic-wan/)
- [StrongSwan Documentation](https://wiki.strongswan.org/)
- [IPsec IKEv2 Configuration Guide](https://wiki.strongswan.org/projects/strongswan/wiki/IKEv2CipherSuites)

## 🆘 Поддержка

При возникновении проблем:

1. Проверьте логи и статус туннеля
2. Обратитесь к документации Cloudflare
3. Свяжитесь с вашим TAM/SE Cloudflare
4. Создайте тикет в Cloudflare Support
