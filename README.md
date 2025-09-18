# 🚀 **CloudBridge Relay**

**P2P mesh сеть для создания "почти LAN" соединений между удаленными серверами клиентов через централизованный relay сервер.**

## 🎯 **Цель проекта**

CloudBridge Relay решает проблему создания надежных P2P соединений между удаленными серверами, которые находятся за NAT/Firewall и не могут напрямую общаться друг с другом. Система обеспечивает:

- **"Почти LAN" опыт** - удаленные серверы общаются как в локальной сети
- **Мультитунельность** - полная изоляция между разными пользователями/организациями
- **Современные протоколы** - QUIC + ICE/STUN/TURN для максимальной совместимости
- **Автоматический fallback** - если прямое соединение не работает, используется relay
- **Простота настройки** - клиент просто подключается к relay серверу

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
- **Протоколы:** QUIC, ICE/STUN/TURN, MASQUE (HTTP/3 proxy)
- **Аутентификация:** JWT токены с tenant изоляцией
- **Контейнеризация:** Docker + Kubernetes
- **Развертывание:** Minikube для разработки, Kubernetes для продакшена

## 📋 **Порты и сервисы**

| Порт | Протокол | Назначение | Статус |
|------|----------|------------|--------|
| 8080 | TCP/HTTPS | HTTP API (Health, Registration) | ✅ Работает |
| 8082 | TCP/HTTPS | P2P API (ICE signaling) | ✅ Работает |
| 19302 | UDP | STUN Server | ✅ Работает |
| 9090 | UDP | QUIC Transport | ✅ Работает |
| 8443 | TCP/HTTPS | HTTPS/MASQUE (Corporate proxy) | ✅ Работает |
| 9092 | UDP | Enhanced QUIC (Advanced features) | ✅ Работает |

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
- ✅ QUIC Transport и Enhanced QUIC
- ✅ MASQUE proxy для корпоративных сетей
- ✅ Docker контейнеризация
- ✅ Kubernetes развертывание

### **🚧 В разработке:**
- 🚧 P2P маршрутизация между клиентами
- 🚧 Heartbeat система для поддержания соединений
- 🚧 ICE кандидаты обмен для NAT traversal
- 🚧 Мультитунельность с полной изоляцией

### **📋 План разработки:**
1. **Этап 1 (1-2 дня):** Базовая P2P маршрутизация
2. **Этап 2 (1-2 дня):** Heartbeat система
3. **Этап 3 (2-3 дня):** Мультитунельность
4. **Этап 4 (1-2 дня):** ICE/STUN интеграция

## 🎯 **Цели проекта**

### **Краткосрочные (1-2 недели):**
- Реализовать P2P маршрутизацию между клиентами
- Добавить heartbeat систему для стабильности
- Обеспечить мультитунельность с изоляцией

### **Долгосрочные (1-2 месяца):**
- Поддержка 1000+ одновременных соединений
- Автоматическое масштабирование
- Мониторинг и метрики
- Production-ready развертывание

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

**🎯 Цель: Создать полноценную P2P mesh сеть, где удаленные серверы клиентов могут общаться как в локальной сети, с полной изоляцией между разными пользователями!**
