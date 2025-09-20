# 2GC CloudBridge Global Network

**2GC CloudBridge Global Network** — это глобальная Zero-Trust оверлейная сеть «**QUIC-first**» для безопасного объединения пользователей, серверов и облаков в единую L3-среду.

## Ключевые принципы

* **Data plane на QUIC/HTTP-3**: мультиплекс-потоки, 0-RTT, миграция соединений при смене IP, датаграммы для realtime-трафика. ([datatracker.ietf.org][1])
* **HTTP-нативные туннели (MASQUE)**: проксирование UDP (**CONNECT-UDP**) и полного IP-трафика (**CONNECT-IP**) через 443/HTTP-3 — «как VPN по правилам HTTP», для обхода жёстких enterprise-сетей. ([datatracker.ietf.org][2])
* **P2P-соединения с NAT-пробивкой**: ICE/STUN/TURN; при неудаче — автоматический relay. Возможна мультиплекс-работа ICE и QUIC на одном UDP-порту (RFC 9443).
* **Zero-Trust-контроль**: доступ «по пользователю/устройству/приложению», а не «по сети», в русле NIST SP 800-207/207A и CISA ZT-моделей зрелости.

## Архитектура

### Data Plane

```
┌─────────────────────────────────────────────────────────────────┐
│                   2GC CloudBridge Global Network                │
├─────────────────────────────────────────────────────────────────┤
│  QUIC/HTTP-3 Transport Layer                                   │
│  ┌─────────────────┐    ┌─────────────────┐                    │
│  │   Connection    │    │   Connection    │                    │
│  │   Migration     │    │   Multiplexing  │                    │
│  │   0-RTT Setup   │    │   Datagrams     │                    │
│  └─────────────────┘    └─────────────────┘                    │
├─────────────────────────────────────────────────────────────────┤
│  MASQUE Tunneling (Port 443/HTTP-3)                           │
│  ┌─────────────────┐    ┌─────────────────┐                    │
│  │   CONNECT-UDP   │    │   CONNECT-IP    │                    │
│  │   UDP Proxy     │    │   Full IP VPN   │                    │
│  └─────────────────┘    └─────────────────┘                    │
├─────────────────────────────────────────────────────────────────┤
│  P2P Layer (ICE/STUN/TURN)                                    │
│  ┌─────────────────┐    ┌─────────────────┐                    │
│  │   Direct P2P    │    │   Relay Fallback│                    │
│  │   NAT Traversal │    │   TURN Server   │                    │
│  └─────────────────┘    └─────────────────┘                    │
└─────────────────────────────────────────────────────────────────┘
```

### Control Plane

```
┌─────────────────────────────────────────────────────────────────┐
│                    Zero-Trust Control Plane                     │
├─────────────────────────────────────────────────────────────────┤
│  Identity & Access Management                                  │
│  ┌─────────────────┐    ┌─────────────────┐                    │
│  │   User/Device   │    │   Application   │                    │
│  │   Authentication│    │   Authorization │                    │
│  └─────────────────┘    └─────────────────┘                    │
├─────────────────────────────────────────────────────────────────┤
│  Policy Engine                                                 │
│  ┌─────────────────┐    ┌─────────────────┐                    │
│  │   Hostname      │    │   Network       │                    │
│  │   Routing       │    │   Segmentation  │                    │
│  └─────────────────┘    └─────────────────┘                    │
├─────────────────────────────────────────────────────────────────┤
│  Service Discovery & Routing                                   │
│  ┌─────────────────┐    ┌─────────────────┐                    │
│  │   Node          │    │   Route         │                    │
│  │   Registration  │    │   Distribution  │                    │
│  └─────────────────┘    └─────────────────┘                    │
└─────────────────────────────────────────────────────────────────┘
```

## Технологический стек

* **QUIC/HTTP-3** как базовый транспорт: один UDP-сокет обслуживает множество соединений и потоков, быстрое установление (1 RTT; 0-RTT при резюме), «**connection migration**» при смене сети. ([datatracker.ietf.org][1])
* **QUIC DATAGRAM** — для UDP-подобных нагрузок (игры/телеметрия/медиа) без надлишней надёжности на уровне транспорта. ([datatracker.ietf.org][3])
* **MASQUE** для туннелей через 443/HTTP-3:
  * **CONNECT-UDP** — проксирование UDP в HTTP. ([datatracker.ietf.org][2])
  * **CONNECT-IP** — «полноценный IP-туннель» поверх HTTP (удобно для Remote-/Site-to-Site). ([datatracker.ietf.org][4])

### NAT-traversal и P2P

* **ICE (RFC 8445)** + **STUN (RFC 8489)** + **TURN (RFC 5766)** — стандартная связка для прямых P2P-линков.
* **Мультиплекс на одном UDP-порту**: RFC 9443 описывает, как различать QUIC, STUN/TURN, DTLS/SRTP на одном сокете — удобно для клиентов и P2P-модели.
* **Relay-fallback**: при симметричных NAT/CGNAT трафик уходит через собственные ретрансляторы (модель аналогична DERP в Tailscale).

### Соответствие стандартам IETF

* **RFC 9000/9001** - QUIC Transport и QUIC-TLS
* **RFC 9114** - HTTP/3
* **RFC 8445/8489/5766/6062** - ICE/STUN/TURN
* **RFC 9298/9484** - MASQUE (CONNECT-UDP/CONNECT-IP)
* **RFC 9220** - Extended CONNECT для WebSocket над HTTP/3

## Порты и сервисы

| Порт | Протокол | Назначение | Статус |
|------|----------|------------|--------|
| 8080 | TCP/HTTPS | HTTP API (Health, Registration) | Работает |
| 8082 | TCP/HTTPS | P2P API (ICE signaling) | Работает |
| 5553 | UDP | QUIC Transport (P2P mesh) | Работает |
| 19302 | UDP | STUN Server (NAT traversal) | Работает |
| 8443 | TCP/HTTPS | HTTPS/MASQUE (Corporate proxy) | Работает |
| 9090 | UDP | QUIC Transport (Legacy) | Работает |
| 9092 | UDP | Enhanced QUIC (Advanced features) | Работает |

## Применение

### Enterprise ZT-доступ и межоблачные связи

* **Банки и финансы** - где UDP часто ограничен
* **Госсектор** - соответствие корпоративным стандартам
* **Межоблачные связи** - MASQUE через порт 443
* «идём по 443/HTTP-3» без открытия входящего UDP; правила по пользователям и **hostname**, а не IP-листам.

### IoT/Edge и филиалы

* **Прямые P2P соединения** - минимальная задержка
* **Автоматический fallback** - при провале пробивки NAT
* **Масштабируемость** - поддержка тысяч устройств
* ICE даёт шанс на прямой путь (меньше задержка), при провале — relay/TURN.

### Медиа/Real-time/DevOps

* **QUIC datagrams** - быстрая доставка
* **Connection migration** - стабильность при роуминге
* **Грамотные fallback** - для блокирующих сетей
* QUIC-датаграммы, 0-RTT, миграция соединений; меньше «ломкости» сессий при роуминге. ([datatracker.ietf.org][3])

## Безопасность и устойчивость

### Защита от zero-day атак в классических VPN

* **Убираем уязвимый периметр-шлюз.** Наиболее громкие инциденты последних лет били по **аппаратным/периметровым VPN-шлюзам** (Ivanti/Pulse, SonicWall, Fortinet) через 0-day/недавно закрытые RCE. Переход на ZTNA/HTTP-нативные туннели исключает экспонированный VPN-апплаенс, тем самым срезает класс атак по уязвимостям устройств. ([Google Cloud][1])

* **Zero Trust вместо "широкого" VPN.** ZTNA даёт доступ «по приложению/идентичности» и скрывает внутренние сервисы от интернета, уменьшая площадь атаки и возможности латерального перемещения при компромиссе учётки. Это ровно то, к чему подталкивает CISA и что описано NIST 800-207. ([CISA][2])

* **HTTP/3/MASQUE (CONNECT-UDP/IP)**: клиенты инициируют **исходящие** соединения на 443/QUIC→HTTP/3-прокси, вместо того чтобы держать публичный VPN-эндпойнт. Меньше экспозиции — меньше шанс стать целью периметровой RCE на «железке». ([datatracker.ietf.org][3])

### DDoS-устойчивость

* **Anycast + масштабирование.** Размещение релеев за anycast-адресами «распыляет» трафик по нескольким узлам/регионам и помогает переживать всплески и DDoS. ([cloudflare.com][4])

* **Встроенные анти-амплификационные механизмы QUIC.** QUIC требует **валидации адреса** и накладывает **anti-amplification limit**; сервер может делать **Retry**, чтобы не стать «усилителем» при спуфинге. Эти механизмы снижает эффективность UDP-амплификаций на входе. ([datatracker.ietf.org][5])

* **Политики/лимиты на уровне прокси.** На MASQUE-/HTTP-слое легко вводить rate-limit, токены/аутентификацию, фаервол правилом — и быстро «выселять» злоумышленника без влияния на остальных. (CONNECT-UDP/CONNECT-IP — стандартные методы). ([datatracker.ietf.org][3])

### Криптографические гарантии

* **Шифрование по умолчанию**: QUIC обязывает TLS 1.3; пер-стримовая криптография, защита метаданных, анти-ossification. ([datatracker.ietf.org][9])
* **Zero-Trust-модель**: минимально необходимый доступ, проверка пользователя/устройства/контекста; рекомендации — NIST SP 800-207/207A, CISA ZT Maturity Model.
* **DDoS-устойчивость на транспорте**: в QUIC предусмотрены **address validation**, **Retry** и **anti-amplification limit (×3 до валидации адреса)** — снижают риск усилительных атак и «дешёвых» отражений.

## Конкурентная среда

### Позиционирование относительно конкурентов

* **Tailscale**: WireGuard-оверлей + собственные DERP-релеи; отличный UX, но UDP-зависимость WG и IP-ориентированность политик. 2GC делает ставку на HTTP-нативный стек (MASQUE) и политики по имени.

* **ZeroTier/Nebula**: UDP-оверлеи с агрессивной NAT-пробивкой; сильные P2P-возможности, но в enterprise-сетях порт 443/HTTP-3 зачастую «проходит лучше». (Здесь 2GC выигрывает HTTP-нативностью.) ([datatracker.ietf.org][2])

* **Cloudflare One/Tunnel**: задал тренд **hostname-routing** и egress-политик «по имени»; 2GC рационально поддерживать совместимую модель политик и приватных имён.

## Соответствие и локализация данных

### Регуляторные требования

* **Zero-Trust-подходы** NIST/CISA закрывают класс регуляторных требований к удалённому доступу (минимизация периметра, микросегментация).
* **Локализация/суверенность**: индустрия движется к управлению местом обработки/хранения (пример — Data Localization Suite у Cloudflare). Для 2GC важно иметь аналогичные механизмы «где инспектировать/хранить метаданные».
* **Пример РФ**: для проектов с персональными данными и КИИ — учитывать 152-ФЗ/187-ФЗ и, при необходимости, лицензирование в части ТЗКИ/СКЗИ (ФСТЭК/ФСБ).

## Риски и ограничения

### Технические ограничения

* **Блокировки UDP/QUIC** в части корпоративных сетей → обязателен стек фолбэков: MASQUE → WebSocket(H2/H3) → TCP. ([datatracker.ietf.org][2])
* **Сложность реального P2P** (CGNAT, асимметрия, политика провайдера) → держать собственный TURN/relay-пул и телеметрию успеха пробивки (как у Tailscale с DERP).
* **L2-семантика**: оверлей по умолчанию L3; если нужен broadcast/multicast — добавлять L2-over-L3 точечно, чтобы не ухудшать масштабируемость.
* **Безопасность реализаций QUIC**: анти-амплификация/Retry должны быть корректно реализованы в стеке (ошибки приводят к уязвимостям).

## Дорожная карта

### Этап А — транспорт/туннели (MVP)

1. QUIC-first + **CONNECT-UDP**, HTTP Datagrams; фолбэки WS/TCP. ([datatracker.ietf.org][2])
2. Базовый **ICE/STUN/TURN** + сбор метрик NAT/успеха.
3. Политики Zero-Trust (по пользователю/устройству/hostname), аудит/логирование.

### Этап B — enterprise-функции

4. **CONNECT-IP** (IP-VPN поверх HTTP-3) для Site-to-Site/Remote-IP. ([datatracker.ietf.org][4])
5. **Hostname-routing** и egress-политики «по имени» (совместимая модель с индустрией).
6. Локализация данных/региональные домены обработки (аналог DLS).

### Этап C — производительность и устойчивость

7. Anycast-релеи, масштабирование, защита от DDoS на уровне QUIC (Retry/anti-amp) и на уровне прокси.
8. Наблюдаемость by default (трассировка, SLO, активные пробы).
9. R&D: **Multipath QUIC** (когда созреет экосистема) и адаптивный выбор пути.

## Быстрый старт

### Требования

* Go 1.21+
* Docker (опционально)
* Kubernetes/Minikube (для развертывания)

### Локальный запуск

```bash
# Установить переменные окружения
export JWT_SECRET="fallback-key"
export RELAY_JWT_SECRET="fallback-key"

# Собрать проект
go build -o ./bin/relay ./cmd/relay

# Запустить сервер
./bin/relay -config configs/config.json -log-level debug
```

### Docker развертывание

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

### Kubernetes развертывание

```bash
# Запустить Minikube
minikube start --driver=docker --memory=4096 --cpus=2

# Собрать и загрузить образ
docker build -t cloudbridge-relay:resource-monitoring-v1 .
minikube image load cloudbridge-relay:resource-monitoring-v1

# Применить конфигурации
kubectl create namespace cloudbridge
kubectl apply -f k8s/deployment.yaml -n cloudbridge
kubectl apply -f k8s/cloudbridge-relay-external-service.yaml -n cloudbridge

# Проверить статус
kubectl get pods -n cloudbridge
kubectl get services -n cloudbridge
```

## Текущее состояние разработки

### Реализовано

* Базовая инфраструктура сервера
* HTTP API и P2P API endpoints
* JWT аутентификация с tenant изоляцией
* STUN сервер для NAT traversal
* QUIC Transport с connection migration
* MASQUE proxy для корпоративных сетей (CONNECT-UDP/CONNECT-IP)
* Docker контейнеризация
* Kubernetes развертывание
* HTTP-нативные туннели через порт 443
* AI-мониторинг ресурсов с Telegram уведомлениями
* Cadence workflow engine для P2P оркестрации

### В разработке

* P2P маршрутизация между клиентами
* ICE кандидаты обмен для NAT traversal
* TURN relay пул для сложных NAT
* Автоматический fallback на WebSocket/TCP
* Мультитунельность с полной изоляцией

## Структура проекта

```
cloudbridge-relay-installer/
├── cmd/relay/                 # Основной сервер
├── internal/                  # Внутренние пакеты
│   ├── api/                  # HTTP/P2P API
│   ├── token/                # JWT токены
│   ├── p2p/                  # P2P маршрутизация
│   ├── cadence/              # Workflow engine
│   └── ai/                   # AI мониторинг
├── k8s/                      # Kubernetes конфигурации
├── configs/                  # Конфигурационные файлы
├── docs/                     # Документация
└── scripts/                  # Утилиты
```

## Документация

### Для разработчиков

* **[DEVELOPER_BRIEFING.md](DEVELOPER_BRIEFING.md)** - Краткий брифинг с основными задачами
* **[DEVELOPER_TECHNICAL_SPECIFICATION.md](DEVELOPER_TECHNICAL_SPECIFICATION.md)** - Полное техническое задание
* **[MULTI_TENANT_ARCHITECTURE.md](MULTI_TENANT_ARCHITECTURE.md)** - Архитектура мультитунельности
* **[MINIKUBE_BUILD_GUIDE.md](MINIKUBE_BUILD_GUIDE.md)** - Инструкция по сборке Minikube
* **[P2P_ROUTING_DEVELOPMENT_PLAN.md](P2P_ROUTING_DEVELOPMENT_PLAN.md)** - План доработки P2P маршрутизации

### API документация

* **[docs/api/AI_API_ENDPOINTS.md](docs/api/AI_API_ENDPOINTS.md)** - AI API endpoints
* **[docs/api/P2P_API_ENDPOINTS.md](docs/api/P2P_API_ENDPOINTS.md)** - P2P API endpoints
* **[docs/api/CADENCE_API_ENDPOINTS.md](docs/api/CADENCE_API_ENDPOINTS.md)** - Cadence API endpoints

## Вывод

**2GC CloudBridge Global Network** — это современная **HTTP-нативная Zero-Trust сеть поверх QUIC**, сочетающая P2P (через ICE) и управляемые туннели (через MASQUE). Такой подход:

* уменьшает площадь атаки (нет «VPN-шлюза-апплаенса» на периметре),
* облегчает прохождение сложной корпоративной инфраструктуры (порт 443/HTTP-3, политики «по имени»),
* даёт выигрыш по задержке/устойчивости (0-RTT, миграция путей, P2P где возможно)
  — и опирается на текущие стандарты IETF и рекомендации по Zero-Trust. ([datatracker.ietf.org][1])

---

## Ссылки и источники

[1]: https://datatracker.ietf.org/doc/rfc9000/ "IETF RFC 9000 - QUIC: A UDP-Based Multiplexed and Secure Transport"
[2]: https://datatracker.ietf.org/doc/rfc9298/ "IETF RFC 9298 - MASQUE: CONNECT-UDP and CONNECT-IP"
[3]: https://datatracker.ietf.org/doc/rfc9221/ "IETF RFC 9221 - An Unreliable Datagram Extension to QUIC"
[4]: https://datatracker.ietf.org/doc/rfc9484/ "IETF RFC 9484 - MASQUE: CONNECT-IP"
[5]: https://cloud.google.com/blog/products/identity-security/zero-trust-architecture-design-principles "Google Cloud - Zero Trust Architecture Design Principles"
[6]: https://www.cisa.gov/resources-tools/resources/zero-trust-maturity-model "CISA - Zero Trust Maturity Model"
[7]: https://www.cloudflare.com/learning/ddos/what-is-anycast/ "Cloudflare - What is Anycast?"
[8]: https://datatracker.ietf.org/doc/rfc9000/ "IETF RFC 9000 - QUIC: A UDP-Based Multiplexed and Secure Transport"
[9]: https://datatracker.ietf.org/doc/rfc9001/ "IETF RFC 9001 - Using TLS to Secure QUIC"
