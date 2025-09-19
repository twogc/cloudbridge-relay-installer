# 🚀 **CloudBridge Relay**

**HTTP-нативная P2P mesh платформа для создания Zero-Trust оверлеев между удаленными серверами с поддержкой QUIC, ICE/STUN/TURN и MASQUE протоколов.**

## 🎯 **Цель проекта**

CloudBridge Relay решает проблему создания надежных P2P соединений между удаленными серверами, которые находятся за NAT/Firewall и не могут напрямую общаться друг с другом. Система обеспечивает:

- **HTTP-нативный подход** - работа через порт 443/HTTP-3 с поддержкой MASQUE (CONNECT-UDP/CONNECT-IP)
- **P2P mesh сеть** - прямые соединения между пирами с автоматическим fallback на relay
- **Современные протоколы** - QUIC + ICE/STUN/TURN для максимальной совместимости
- **Корпоративная совместимость** - обход блокировок UDP через HTTP-туннели
- **Kubernetes-native** - простое развертывание в современных инфраструктурах

## 🌟 **Ключевые преимущества**

### **HTTP-нативный подход**
- **Работа через порт 443** - обходит корпоративные блокировки UDP
- **MASQUE протокол** - CONNECT-UDP/CONNECT-IP для HTTP-туннелей
- **Политик-ориентированное проксирование** - соответствует корпоративным стандартам

### **Современные протоколы**
- **QUIC с connection migration** - смена IP без разрыва соединения
- **ICE/STUN/TURN** - проверенная модель NAT traversal (как в WebRTC)
- **HTTP Datagrams** - быстрая доставка поверх QUIC DATAGRAM

### **Корпоративная совместимость**
- **Fallback на WebSocket/TCP** - для сетей с ограничениями UDP
- **WireGuard опция** - для site-to-site/legacy соединений
- **Kubernetes-native** - соответствует трендам CNCF

## 🏗️ **Архитектура**

```
┌─────────────────────────────────────────────────────────────────┐
│                    CloudBridge Relay                            │
├─────────────────────────────────────────────────────────────────┤
│  Tenant A (tenant-216420165)                                   │
│  Network: 10.100.77.0/24                                      │
│  ┌─────────────────┐    ┌─────────────────┐                    │
│  │   Server A1     │◄──►│   Server A2     │                    │
│  │ 10.100.77.10    │    │ 10.100.77.11    │                    │
│  │ QUIC + ICE      │    │ QUIC + ICE      │                    │
│  └─────────────────┘    └─────────────────┘                    │
├─────────────────────────────────────────────────────────────────┤
│  Tenant B (tenant-123456789)                                   │
│  Network: 10.100.78.0/24                                      │
│  ┌─────────────────┐    ┌─────────────────┐                    │
│  │   Server B1     │◄──►│   Server B2     │                    │
│  │ 10.100.78.10    │    │ 10.100.78.11    │                    │
│  │ QUIC + ICE      │    │ QUIC + ICE      │                    │
│  └─────────────────┘    └─────────────────┘                    │
└─────────────────────────────────────────────────────────────────┘
```

## 🔧 **Технологический стек**

- **Backend:** Go 1.21+
- **Протоколы:** 
  - **QUIC** (RFC 9000/9001) - основной транспорт с connection migration
  - **ICE/STUN/TURN** (RFC 8445/8489/5766) - NAT traversal для P2P соединений
  - **MASQUE** (RFC 9298/9484) - HTTP-нативные туннели (CONNECT-UDP/CONNECT-IP)
  - **HTTP Datagrams** (RFC 9297) - поверх QUIC DATAGRAM (RFC 9221)
- **Аутентификация:** JWT токены с tenant изоляцией
- **Контейнеризация:** Docker + Kubernetes
- **Развертывание:** Minikube для разработки, Kubernetes для продакшена

### **Соответствие стандартам IETF:**
- **RFC 9000/9001** - QUIC Transport и QUIC-TLS
- **RFC 9114** - HTTP/3
- **RFC 8445/8489/5766/6062** - ICE/STUN/TURN
- **RFC 9298/9484** - MASQUE (CONNECT-UDP/CONNECT-IP)
- **RFC 9220** - Extended CONNECT для WebSocket над HTTP/3

## 📋 **Порты и сервисы**

| Порт | Протокол | Назначение | Статус |
|------|----------|------------|--------|
| 8080 | TCP/HTTPS | HTTP API (Health, Registration) | ✅ Работает |
| 8082 | TCP/HTTPS | P2P API (ICE signaling) | ✅ Работает |
| 5553 | UDP | QUIC Transport (P2P mesh) | ✅ Работает |
| 19302 | UDP | STUN Server (NAT traversal) | ✅ Работает |
| 8443 | TCP/HTTPS | HTTPS/MASQUE (Corporate proxy) | ✅ Работает |
| 9090 | UDP | QUIC Transport (Legacy) | ✅ Работает |
| 9092 | UDP | Enhanced QUIC (Advanced features) | ✅ Работает |

## 🎯 **Применение**

### **Enterprise ZT-доступ**
- **Банки и финансы** - где UDP часто ограничен
- **Госсектор** - соответствие корпоративным стандартам
- **Межоблачные связи** - MASQUE через порт 443

### **IoT/Edge и филиалы**
- **Прямые P2P соединения** - минимальная задержка
- **Автоматический fallback** - при провале пробивки NAT
- **Масштабируемость** - поддержка тысяч устройств

### **Медиа/Real-time/Гейминг**
- **QUIC datagrams** - быстрая доставка
- **Connection migration** - стабильность при роуминге
- **Грамотные fallback** - для блокирующих сетей

## ⚠️ **Риски и решения**

### **Блокировки UDP/QUIC в enterprise**
**Проблема:** Корпоративные сети часто блокируют UDP/QUIC
**Решение:** 
- WebSocket (H2/H3) fallback
- TCP туннель
- MASQUE через порт 443

### **Сложность P2P в реальном интернете**
**Проблема:** CGNAT, асимметрия, симметричный NAT
**Решение:**
- Собственный TURN/relay пул
- Телеметрия успеха пробивки
- Метрики NAT-типов

### **L2-семантика**
**Проблема:** По умолчанию L3 "как в одной VLAN"
**Решение:**
- Точечно добавлять L2-over-L3 (VXLAN)
- Иначе держать L3
- Не "утяжелять" всем

## 🚀 **Быстрый старт**

### **Требования:**
- Go 1.21+
- Docker (опционально)
- Kubernetes/Minikube (для развертывания)

### **1. Локальный запуск (Go)**

```bash
# Установить переменные окружения
export JWT_SECRET="fallback-key"
export RELAY_JWT_SECRET="fallback-key"

# Собрать проект
go build -o ./bin/relay ./cmd/relay

# Запустить сервер
./bin/relay -config configs/config.json -log-level debug
```

### **2. Запуск через Docker**

```bash
# Собрать Docker образ
docker build -t cloudbridge-relay:latest .

# Запустить контейнер
docker run -d \
  --name cloudbridge-relay \
  -p 8080:8080 -p 8082:8082 -p 19302:19302/udp \
  -p 9090:9090/udp -p 8443:8443 -p 9092:9092/udp \
  -e JWT_SECRET="fallback-key" \
  -e RELAY_JWT_SECRET="fallback-key" \
  cloudbridge-relay:latest
```

### **3. Развертывание в Minikube**

```bash
# Запустить Minikube
minikube start --driver=docker --memory=4096 --cpus=2

# Собрать и загрузить образ
docker build -t cloudbridge-relay:quic-ice-v2 .
minikube image load cloudbridge-relay:quic-ice-v2

# Применить конфигурации
kubectl create namespace cloudbridge
kubectl apply -f k8s/deployment.yaml -n cloudbridge
kubectl apply -f k8s/cloudbridge-relay-external-service.yaml -n cloudbridge

# Проверить статус
kubectl get pods -n cloudbridge
kubectl get services -n cloudbridge
```

## 🧪 **Тестирование**

### **Проверка здоровья сервисов:**

```bash
# HTTP API
curl http://localhost:8080/health

# P2P API
curl http://localhost:8082/health

# STUN сервер
python3 -c "
import socket
import struct

def test_stun():
    sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    sock.settimeout(3.0)
    
    try:
        transaction_id = b'123456789012'
        magic_cookie = 0x2112A442
        stun_request = struct.pack('>HHI', 0x0001, 0x0000, magic_cookie) + transaction_id
        
        sock.sendto(stun_request, ('localhost', 19302))
        response, addr = sock.recvfrom(1024)
        
        msg_type = struct.unpack('>H', response[0:2])[0]
        if msg_type == 0x0101:
            print('✅ STUN Server - WORKING')
        else:
            print(f'⚠️  STUN Server - Unexpected response: 0x{msg_type:04x}')
        
    except Exception as e:
        print(f'❌ STUN Server - Error: {e}')
    finally:
        sock.close()

test_stun()
"
```

### **Регистрация пира:**

```bash
# Использовать JWT токен из official-token.txt
curl -X POST "http://localhost:8082/api/v1/tenants/tenant-216420165/peers/register" \
  -H "Authorization: Bearer $(cat official-token.txt)" \
  -H "Content-Type: application/json" \
  -d '{
    "public_key": "test-peer-key",
    "allowed_ips": ["10.100.77.10/32"],
    "peer_info": {
      "name": "Test Peer",
      "location": "Test Location"
    }
  }'
```

## 📚 **Документация**

### **Для разработчиков:**
- **[DEVELOPER_BRIEFING.md](DEVELOPER_BRIEFING.md)** - Краткий брифинг с основными задачами
- **[DEVELOPER_TECHNICAL_SPECIFICATION.md](DEVELOPER_TECHNICAL_SPECIFICATION.md)** - Полное техническое задание
- **[MULTI_TENANT_ARCHITECTURE.md](MULTI_TENANT_ARCHITECTURE.md)** - Архитектура мультитунельности
- **[MINIKUBE_BUILD_GUIDE.md](MINIKUBE_BUILD_GUIDE.md)** - Инструкция по сборке Minikube
- **[P2P_ROUTING_DEVELOPMENT_PLAN.md](P2P_ROUTING_DEVELOPMENT_PLAN.md)** - План доработки P2P маршрутизации

### **Структура проекта:**
```
cloudbridge-relay-installer/
├── cmd/relay/                 # Основной сервер
├── internal/                  # Внутренние пакеты
│   ├── api/                  # HTTP/P2P API
│   ├── token/                # JWT токены
│   └── p2p/                  # P2P маршрутизация
├── k8s/                      # Kubernetes конфигурации
├── configs/                  # Конфигурационные файлы
├── nginx/                    # Nginx конфигурации
└── systemd/                  # Systemd сервисы
```

## 🔄 **Текущее состояние разработки**

### **✅ Реализовано:**
- ✅ Базовая инфраструктура сервера
- ✅ HTTP API и P2P API endpoints
- ✅ JWT аутентификация с tenant изоляцией
- ✅ STUN сервер для NAT traversal
- ✅ QUIC Transport с connection migration
- ✅ MASQUE proxy для корпоративных сетей (CONNECT-UDP/CONNECT-IP)
- ✅ Docker контейнеризация
- ✅ Kubernetes развертывание
- ✅ HTTP-нативные туннели через порт 443

### **🚧 В разработке:**
- 🚧 P2P маршрутизация между клиентами
- 🚧 ICE кандидаты обмен для NAT traversal
- 🚧 TURN relay пул для сложных NAT
- 🚧 Автоматический fallback на WebSocket/TCP
- 🚧 Мультитунельность с полной изоляцией

### **📋 План разработки:**
1. **Этап 1 (1-2 дня):** Базовая P2P маршрутизация
2. **Этап 2 (1-2 дня):** ICE/STUN/TURN интеграция
3. **Этап 3 (2-3 дня):** MASQUE CONNECT-UDP/CONNECT-IP
4. **Этап 4 (1-2 дня):** Автоматический fallback на WebSocket/TCP

## 🎯 **Цели проекта**

### **Краткосрочные (1-2 недели):**
- Реализовать P2P маршрутизацию между клиентами
- Добавить ICE/STUN/TURN для NAT traversal
- Обеспечить MASQUE CONNECT-UDP/CONNECT-IP
- Автоматический fallback на WebSocket/TCP

### **Долгосрочные (1-2 месяца):**
- Поддержка 1000+ одновременных соединений
- Собственный TURN relay пул
- Мониторинг и метрики NAT-типов
- Production-ready развертывание
- HTTP-нативный Zero-Trust оверлей

## 🛠️ **Практические рекомендации для MVP/GA**

### **Техническая реализация**
```yaml
QUIC-first + MASQUE:
  - CONNECT-UDP (RFC 9298)
  - CONNECT-IP (RFC 9484) 
  - HTTP Datagrams (RFC 9297)
  - Downgrade на WS/TCP

ICE/STUN/TURN слой:
  - Собственный STUN сервер
  - Управляемый TURN пул
  - TCP-alloc при необходимости
  - Метрики NAT-типов/успеха
```

### **Продуктовые функции**
```yaml
"Сеть как продукт":
  - Политики доступа (hostname/SNI/DNS)
  - Observability и мониторинг
  - Автодиагностика соединений
  - HTTP-VPN подход
```

### **WireGuard интеграция**
```yaml
Профиль для site-to-site/legacy:
  - Чёткие ограничения UDP
  - Сценарии роуминга
  - Собственные релеи как "последний шанс"
```

## 🤝 **Участие в разработке**

### **Для разработчиков:**
1. Изучите [DEVELOPER_BRIEFING.md](DEVELOPER_BRIEFING.md)
2. Ознакомьтесь с [техническим заданием](DEVELOPER_TECHNICAL_SPECIFICATION.md)
3. Следуйте [плану разработки](P2P_ROUTING_DEVELOPMENT_PLAN.md)

### **Для DevOps:**
1. Используйте [инструкцию по Minikube](MINIKUBE_BUILD_GUIDE.md)
2. Настройте мониторинг и логирование
3. Подготовьте production окружение

## 📞 **Поддержка**

- **Issues:** Создавайте issues для багов и feature requests
- **Документация:** Все документы в корне проекта
- **Тестирование:** Используйте скрипты в `test-*.sh`

## 📄 **Лицензия**

Проект распространяется под лицензией MIT. См. [LICENSE](LICENSE) для подробностей.

---

## 🎯 **Потенциал проекта**

### **Техническая база: ⭐⭐⭐⭐⭐**
- ✅ **QUIC + MASQUE** - современный HTTP-нативный подход
- ✅ **ICE/STUN/TURN** - проверенная модель пробивки
- ✅ **Kubernetes-native** - соответствует трендам CNCF
- ✅ **Модульная архитектура** - гибкость развертывания

### **Рыночный потенциал: ⭐⭐⭐⭐**
**Ниша "HTTP-нативного Zero-Trust-оверлея":**
- ✅ **Enterprise** - где UDP ограничен
- ✅ **IoT/Edge** - нужна низкая задержка
- ✅ **Мульти-облако** - межоблачная связность
- ✅ **Корпоративные прокси** - прохождение "по правилам"

### **Конкурентные преимущества: ⭐⭐⭐⭐**
- ✅ **HTTP-нативность** - обходит блокировки
- ✅ **P2P-mesh** - устойчивость к сбоям
- ✅ **Connection migration** - мобильность
- ✅ **Open source** - прозрачность и доверие

---

**🎯 Цель: Создать HTTP-нативную P2P mesh платформу для Zero-Trust оверлеев, где удаленные серверы могут общаться через современные протоколы QUIC + ICE/STUN/TURN + MASQUE с автоматическим fallback на WebSocket/TCP!**

**Проект имеет высокий шанс занять нишу HTTP-нативного Zero-Trust-оверлея в enterprise/IoT/мульти-облаке, где классическим UDP-VPN тяжело проходить инфраструктурные фильтры.**
