# 📚 **Индекс референсных репозиториев CloudBridge Relay**

## 🎯 **Цель**
Эта папка содержит все ключевые репозитории, которые служат шпаргалкой и референсом для реализации проекта CloudBridge Relay. Каждый репозиторий представляет собой готовое решение для конкретного компонента нашей архитектуры.

---

## 🚀 **1. QUIC-first + MASQUE (Ядро)**

### **masque-go** - `quic-go/masque-go`
- **Описание**: Готовый клиент и прокси для проксирования UDP через HTTP/3
- **Стандарт**: CONNECT-UDP (RFC 9298)
- **Язык**: Go
- **Использование**: Базовая реализация MASQUE поверх quic-go
- **Ключевые файлы**: 
  - `client/` - MASQUE клиент
  - `proxy/` - MASQUE прокси сервер
  - `examples/` - примеры использования

### **connect-ip-go** - `quic-go/connect-ip-go`
- **Описание**: Клиент и прокси для "IP over HTTP/3" - полноценный IP-туннель через H3
- **Стандарт**: CONNECT-IP (RFC 9484)
- **Язык**: Go
- **Использование**: IP-туннелирование через HTTP/3
- **Ключевые файлы**:
  - `client/` - CONNECT-IP клиент
  - `proxy/` - CONNECT-IP прокси
  - `examples/` - примеры IP туннелирования

### **masquerade** - `jromwu/masquerade`
- **Описание**: Альтернативная реализация MASQUE на Rust
- **Стандарт**: CONNECT-UDP/HTTP/3
- **Язык**: Rust
- **Использование**: Референс для Rust-реализации, клиент = SOCKS5/HTTP
- **Ключевые файлы**:
  - `src/` - исходный код Rust
  - `examples/` - примеры использования

---

## 🌐 **2. ICE/STUN/TURN + Метрики NAT**

### **ice** - `pion/ice`
- **Описание**: Полноценная реализация ICE (Interactive Connectivity Establishment)
- **Язык**: Go
- **Использование**: NAT traversal, сбор ICE кандидатов
- **Ключевые файлы**:
  - `agent.go` - ICE агент
  - `candidate.go` - ICE кандидаты
  - `examples/` - примеры NAT traversal

### **stun** - `pion/stun`
- **Описание**: STUN (Session Traversal Utilities for NAT) протокол
- **Язык**: Go
- **Использование**: STUN сервер/клиент, NAT behavior discovery
- **Ключевые файлы**:
  - `client.go` - STUN клиент
  - `server.go` - STUN сервер
  - `cmd/stun-nat-behaviour/` - утилита для определения типа NAT

### **coturn** - `coturn/coturn`
- **Описание**: Де-факто стандарт TURN сервера с Prometheus метриками
- **Язык**: C
- **Использование**: TURN сервер, метрики на :9641/metrics
- **Ключевые файлы**:
  - `src/` - исходный код TURN сервера
  - `examples/` - конфигурационные примеры
  - `prometheus/` - метрики для мониторинга

### **go-stun** - `pixelbender/go-stun`
- **Описание**: Альтернативная реализация STUN/TURN/ICE с NAT Behavior Discovery
- **Язык**: Go
- **Использование**: Классификация NAT, STUN/TURN/ICE протоколы
- **Ключевые файлы**:
  - `stun/` - STUN реализация
  - `turn/` - TURN реализация
  - `ice/` - ICE реализация

---

## 🔒 **3. WireGuard (Legacy/Site-to-Site)**

### **wireguard-go** - `WireGuard/wireguard-go`
- **Описание**: Userspace реализация WireGuard на Go
- **Язык**: Go
- **Использование**: WireGuard интерфейсы, userspace WG
- **Ключевые файлы**:
  - `device/` - WireGuard устройство
  - `tun/` - TUN интерфейс
  - `examples/` - примеры использования

### **wgctrl-go** - `WireGuard/wgctrl-go`
- **Описание**: Управление WireGuard интерфейсами
- **Язык**: Go
- **Использование**: Программное управление WG интерфейсами
- **Ключевые файлы**:
  - `wgctrl.go` - основной API
  - `examples/` - примеры управления

---

## 🔄 **4. Fallback механизмы**

### **tailscale** - `tailscale/tailscale`
- **Описание**: DERP (Designated Encrypted Relay for Packets) - релей для WG пакетов
- **Язык**: Go
- **Использование**: WG пакеты через TLS/HTTPS когда UDP заблокирован
- **Ключевые файлы**:
  - `cmd/derper/` - DERP сервер
  - `derp/` - DERP пакет
  - `examples/` - примеры релея

### **wstunnel** - `erebe/wstunnel`
- **Описание**: Общий туннель поверх WebSocket/HTTP(S)
- **Язык**: Go
- **Использование**: WireGuard через WebSocket/443
- **Ключевые файлы**:
  - `cmd/` - CLI утилиты
  - `pkg/` - библиотеки
  - `examples/` - примеры туннелирования

### **websocket** - `gorilla/websocket`
- **Описание**: Быстрая и надежная WebSocket библиотека
- **Язык**: Go
- **Использование**: WebSocket соединения для fallback
- **Ключевые файлы**:
  - `conn.go` - WebSocket соединение
  - `examples/` - примеры WebSocket

---

## 📊 **5. Дополнительные репозитории**

### **openp2p** - `openp2p/openp2p`
- **Описание**: P2P сеть с NAT traversal
- **Использование**: Референс для P2P архитектуры

### **p2pquic** - `p2pquic/p2pquic`
- **Описание**: P2P поверх QUIC
- **Использование**: Референс для QUIC P2P

### **quic-p2p** - `quic-p2p/quic-p2p`
- **Описание**: P2P сеть на базе QUIC
- **Использование**: Референс для QUIC P2P архитектуры

### **quicssh** - `quicssh/quicssh`
- **Описание**: SSH поверх QUIC
- **Использование**: Референс для SSH over QUIC

### **reverse-http** - `reverse-http/reverse-http`
- **Описание**: Reverse HTTP туннели
- **Использование**: Референс для reverse туннелирования

---

## 🎯 **Как использовать эти репозитории**

### **Для разработки CloudBridge Relay:**

1. **MASQUE компонент**: Изучить `masque-go` и `connect-ip-go` для реализации CONNECT-UDP/IP
2. **ICE/STUN/TURN**: Использовать `pion/ice`, `pion/stun` и `coturn` для NAT traversal
3. **WireGuard fallback**: Изучить `wireguard-go` и `wgctrl-go` для legacy поддержки
4. **Fallback механизмы**: Использовать `tailscale` (DERP) и `wstunnel` для обхода блокировок

### **Ключевые паттерны для изучения:**

- **HTTP/3 + QUIC datagrams** - из `masque-go`
- **NAT traversal** - из `pion/ice`
- **TURN сервер с метриками** - из `coturn`
- **WebSocket fallback** - из `wstunnel`
- **DERP релей** - из `tailscale`

### **Интеграционные примеры:**

- **MASQUE + ICE** - комбинировать `masque-go` с `pion/ice`
- **WG + DERP** - использовать `wireguard-go` с `tailscale/derp`
- **QUIC + WebSocket** - fallback из QUIC в WebSocket

---

## 📝 **Примечания**

- Все репозитории скачаны в актуальном состоянии
- Каждый репозиторий содержит примеры и документацию
- Рекомендуется изучать примеры в папке `examples/` каждого репозитория
- Для интеграции использовать паттерны из нескольких репозиториев одновременно

---

*Индекс создан: 21 сентября 2025*  
*Статус: Все референсные репозитории загружены ✅*


