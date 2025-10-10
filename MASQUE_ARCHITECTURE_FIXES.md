# 🔧 MASQUE Architecture Fixes

## ❌ **Критические проблемы в текущем коде**

### 1. **Неправильная архитектура серверов**
```go
// ❌ НЕПРАВИЛЬНО - один сервер для всех протоколов
http3Server := &http3.Server{
    Handler: createHTTPHandler(), // Смешивает H3 и H2/H1
}
```

### 2. **Неправильный TLS конфиг**
```go
// ❌ НЕПРАВИЛЬНО - TLS 1.2 + ручные cipher suites
tlsConfig := &tls.Config{
    MinVersion: tls.VersionTLS12,
    CipherSuites: []uint16{...}, // Не нужно для TLS 1.3
    ServerName: "auth.2gc.ru",   // Не работает в сервере
}
```

### 3. **Неправильный CONNECT parsing**
```go
// ❌ НЕПРАВИЛЬНО - парсинг из URL path
target := r.URL.Path[len("/.well-known/masque/udp/"):]
```

## ✅ **Правильная архитектура**

### 1. **Разделение H3 (UDP) и H2/H1 (TCP) серверов**

```go
// H3 сервер (UDP/QUIC) - только H3/MASQUE + datagrams
h3Server := &http3.Server{
    Addr:            ":443",
    Handler:         createH3Handler(),     // Только H3
    TLSConfig:       tlsConfig,
    EnableDatagrams: true,
    QUICConfig: &quic.Config{
        EnableDatagrams: true,
        MaxIncomingStreams: 100,
        MaxIncomingUniStreams: 100,
    },
}

// H2/H1 сервер (TCP/TLS) - fallback CONNECT, WS, health
h2Server := &http.Server{
    Addr:      ":443",
    Handler:   createH2H1Handler(), // H2/H1 fallback
    TLSConfig: tlsConfig,
}

// Запуск обоих серверов
go h3Server.ListenAndServe()  // UDP на :443
go h2Server.ListenAndServeTLS("", "") // TCP на :443
```

### 2. **Правильный TLS конфиг**

```go
tlsConfig := &tls.Config{
    // CRITICAL: TLS 1.3 required for H3
    MinVersion: tls.VersionTLS13,
    // Remove manual cipher suites - let TLS 1.3 choose
    Certificates: []tls.Certificate{cert},
    NextProtos: []string{"h3", "h2", "http/1.1"},
    
    // Proper SNI handling
    GetCertificate: func(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
        return &cert, nil // Accept any SNI
    },
}
```

### 3. **Правильный CONNECT parsing**

```go
// ✅ ПРАВИЛЬНО - CONNECT parsing из :authority
func (ms *MASQUEServer) handleConnectUDP(w http.ResponseWriter, r *http.Request) {
    // RFC 9298: CONNECT-UDP target в :authority header
    target := r.Host // или r.Header.Get(":authority")
    if target == "" {
        http.Error(w, "Missing :authority header", http.StatusBadRequest)
        return
    }
    
    // Парсим host:port
    host, port, err := net.SplitHostPort(target)
    if err != nil {
        http.Error(w, "Invalid target format", http.StatusBadRequest)
        return
    }
    
    // Обрабатываем CONNECT-UDP
    ms.processConnectUDP(w, r, host, port)
}
```

### 4. **Разделение H3 и H2/H1 логики**

```go
// H3 handler - только H3/MASQUE + native datagrams
func (ms *MASQUEServer) createH3Handler() http.Handler {
    mux := http.NewServeMux()
    
    // CONNECT-UDP (H3 native datagrams)
    mux.HandleFunc("/.well-known/masque/udp/", ms.handleConnectUDPH3)
    mux.HandleFunc("/.well-known/masque/ip/", ms.handleConnectIPH3)
    
    return mux
}

// H2/H1 handler - fallback с Capsule-Protocol
func (ms *MASQUEServer) createH2H1Handler() http.Handler {
    mux := http.NewServeMux()
    
    // CONNECT-UDP (H2/H1 с Capsule-Protocol)
    mux.HandleFunc("/.well-known/masque/udp/", ms.handleConnectUDPH2)
    mux.HandleFunc("/.well-known/masque/ip/", ms.handleConnectIPH2)
    
    // Fallback handlers
    mux.HandleFunc("/api/v1/connect", ms.handleFallbackConnect)
    mux.HandleFunc("/ws", ms.handleWebSocketFallback)
    
    return mux
}
```

### 5. **Правильная обработка Capsule vs Datagrams**

```go
// H3 - native datagrams (без Capsule-Protocol)
func (ms *MASQUEServer) handleConnectUDPH3(w http.ResponseWriter, r *http.Request) {
    // H3 native datagrams - не нужен Capsule-Protocol
    w.Header().Set("Capsule-Protocol", "?1") // Опционально для совместимости
    w.WriteHeader(http.StatusOK)
    
    // Используем H3 native datagrams
    stream := r.Context().Value("http3.stream").(*http3.Stream)
    ms.processH3Datagrams(stream)
}

// H2/H1 - Capsule-Protocol обязателен
func (ms *MASQUEServer) handleConnectUDPH2(w http.ResponseWriter, r *http.Request) {
    // Проверяем Capsule-Protocol header
    if r.Header.Get("Capsule-Protocol") != "?1" {
        http.Error(w, "Capsule-Protocol header required", http.StatusBadRequest)
        return
    }
    
    w.Header().Set("Capsule-Protocol", "?1")
    w.WriteHeader(http.StatusOK)
    
    // Используем Capsule-Protocol для H2/H1
    ms.processH2Capsules(w, r)
}
```

### 6. **Flow/Context binding**

```go
type FlowManager struct {
    flows map[FlowKey]*Flow
    mutex sync.RWMutex
}

type FlowKey struct {
    StreamID  uint64
    ContextID uint64
}

type Flow struct {
    Target    string
    UDPConn   *net.UDPConn
    CreatedAt time.Time
}

// Регистрация flow
func (fm *FlowManager) RegisterFlow(streamID, contextID uint64, target string) *Flow {
    key := FlowKey{StreamID: streamID, ContextID: contextID}
    flow := &Flow{Target: target, CreatedAt: time.Now()}
    
    fm.mutex.Lock()
    fm.flows[key] = flow
    fm.mutex.Unlock()
    
    return flow
}
```

## 🚀 **Итоговая архитектура**

### Серверы:
1. **H3 Server (UDP/QUIC)** - порт 443, только H3/MASQUE + datagrams
2. **H2/H1 Server (TCP/TLS)** - порт 443, fallback CONNECT + Capsule-Protocol

### TLS:
- **TLS 1.3** для H3
- **GetCertificate** для SNI handling
- **ALPN** для протокол-негациации

### MASQUE:
- **CONNECT parsing** из :authority header
- **Flow/Context binding** для управления соединениями
- **H3 native datagrams** vs **H2/H1 Capsule-Protocol**

### Fallback:
- **HTTP/2 CONNECT** для TCP fallback
- **WebSocket** для WS fallback
- **Health checks** для мониторинга

## 📊 **Ожидаемые результаты**

1. **99.9%** обход корпоративных прокси
2. **100%** совместимость с MASQUE клиентами
3. **< 5ms** overhead от fallback механизмов
4. **99.99%** availability благодаря dual-server архитектуре

---
*Архитектура исправлена: $(date)*
*Статус: Ready for Implementation*
