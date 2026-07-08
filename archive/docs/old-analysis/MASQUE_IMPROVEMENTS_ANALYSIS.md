# Анализ MASQUE улучшений на основе reference репозиториев

## **Анализ masque-go reference**

### **Ключевые компоненты masque-go:**

#### **1. Client.go - CONNECT-UDP клиент**
```go
// Основные улучшения для нашего проекта:
type Client struct {
    TLSClientConfig *tls.Config
    QUICConfig      *quic.Config
    dialOnce        sync.Once
    dialErr         error
    conn            *quic.Conn
    clientConn      *http3.ClientConn
}
```

**Ключевые особенности:**
- **EnableDatagrams: true** - обязательное условие
- **InitialPacketSize: 1350** - увеличенный размер пакета для QUIC туннелирования
- **Extended CONNECT проверка** - проверка поддержки сервером
- **Datagrams проверка** - проверка поддержки datagrams
- **Context ID управление** - правильное управление Context ID

#### **2. Proxy.go - CONNECT-UDP прокси**
```go
// Основные улучшения для нашего проекта:
type Proxy struct {
    mx       sync.Mutex
    closed   bool
    refCount sync.WaitGroup
    closers  map[io.Closer]struct{}
}
```

**Ключевые особенности:**
- **Proxy-Status заголовки** - RFC 9209 совместимые заголовки
- **DNS error handling** - правильная обработка DNS ошибок
- **Context ID = 0** - только для UDP payloads
- **Goroutine management** - правильное управление горутинами
- **Error mapping** - маппинг ошибок в HTTP статусы

#### **3. Conn.go - управление соединениями**
```go
// Основные улучшения для нашего проекта:
type proxiedConn struct {
    str        http3Stream
    localAddr  net.Addr
    remoteAddr net.Addr
    closed     atomic.Bool
    readDone   chan struct{}
    deadlineMx sync.Mutex
    readCtx    context.Context
    readCtxCancel context.CancelFunc
    deadline   time.Time
    readDeadlineTimer *time.Timer
}
```

**Ключевые особенности:**
- **Atomic operations** - для thread-safe операций
- **Context management** - правильное управление контекстами
- **Deadline handling** - обработка таймаутов
- **Capsule skipping** - пропуск неиспользуемых капсул

---

## **Анализ connect-ip-go reference**

### **Ключевые компоненты connect-ip-go:**

#### **1. Client.go - CONNECT-IP клиент**
```go
// Основные улучшения для нашего проекта:
func Dial(ctx context.Context, conn *http3.ClientConn, template *uritemplate.Template) (*Conn, *http.Response, error) {
    // Проверка Extended CONNECT
    if !settings.EnableExtendedConnect {
        return nil, nil, errors.New("connect-ip: server didn't enable Extended CONNECT")
    }
    // Проверка Datagrams
    if !settings.EnableDatagrams {
        return nil, nil, errors.New("connect-ip: server didn't enable datagrams")
    }
}
```

**Ключевые особенности:**
- **Extended CONNECT проверка** - обязательная проверка
- **Datagrams проверка** - обязательная проверка
- **Capsule Protocol** - поддержка капсул
- **Error handling** - правильная обработка ошибок

#### **2. Conn.go - IP соединения**
```go
// Основные улучшения для нашего проекта:
type Conn struct {
    str    http3Stream
    writes chan writeCapsule
    
    assignedAddressNotify chan struct{}
    availableRoutesNotify chan struct{}
    
    mu                sync.Mutex
    peerAddresses     []netip.Prefix
    localRoutes       []IPRoute
    assignedAddresses []netip.Prefix
    availableRoutes   []IPRoute
    
    closeChan chan struct{}
    closeErr  error
}
```

**Ключевые особенности:**
- **IP Route management** - управление IP маршрутами
- **Address assignment** - назначение IP адресов
- **Capsule handling** - обработка капсул
- **ICMP support** - поддержка ICMP пакетов
- **MTU handling** - обработка MTU

---

## 🔧 **План улучшений для нашего проекта**

### **1. Улучшения CONNECT-UDP (на основе masque-go)**

#### **Текущие проблемы в нашем коде:**
- Неполная обработка ошибок
- Неоптимальная производительность
- Отсутствие retry логики
- Неполная поддержка edge cases

#### **План улучшения:**
```go
// 1. Улучшить error handling
func errToStatus(err error) int {
    var netErr net.Error
    if errors.As(err, &netErr) && netErr.Timeout() {
        return http.StatusGatewayTimeout
    }
    var dnsError *net.DNSError
    if errors.As(err, &dnsError) {
        return http.StatusBadGateway
    }
    // ... остальные типы ошибок
}

// 2. Добавить Proxy-Status заголовки
func (s *Proxy) Proxy(w http.ResponseWriter, r *Request) error {
    proxyStatus := httpsfv.NewItem(r.Host)
    writeProxyStatus := func(err error) error {
        if err != nil {
            proxyStatus.Params.Add("details", err.Error())
        }
        proxyStatusVal, marshalErr := httpsfv.Marshal(proxyStatus)
        if marshalErr != nil {
            return marshalErr
        }
        w.Header().Add("Proxy-Status", proxyStatusVal)
        return err
    }
    // ... остальная логика
}

// 3. Улучшить Context ID управление
func (s *Proxy) proxyConnSend(conn *net.UDPConn, str *http3.Stream) error {
    for {
        data, err := str.ReceiveDatagram(context.Background())
        if err != nil {
            return err
        }
        contextID, n, err := quicvarint.Parse(data)
        if err != nil {
            return err
        }
        if contextID != 0 {
            // Drop this datagram. We currently only support proxying of UDP payloads.
            continue
        }
        if _, err := conn.Write(data[n:]); err != nil {
            return err
        }
    }
}
```

### **2. Улучшения CONNECT-IP (на основе connect-ip-go)**

#### **Текущие проблемы в нашем коде:**
- Ограниченная поддержка IPv6
- Неоптимальная маршрутизация
- Отсутствие connection multiplexing
- Неполная обработка IP packets

#### **План улучшения:**
```go
// 1. Улучшить IP Route management
type Conn struct {
    str    http3Stream
    writes chan writeCapsule
    
    assignedAddressNotify chan struct{}
    availableRoutesNotify chan struct{}
    
    mu                sync.Mutex
    peerAddresses     []netip.Prefix
    localRoutes       []IPRoute
    assignedAddresses []netip.Prefix
    availableRoutes   []IPRoute
    
    closeChan chan struct{}
    closeErr  error
}

// 2. Добавить ICMP support
const (
    ipProtoICMP   = 1
    ipProtoICMPv6 = 58
)

// 3. Улучшить MTU handling
const minMTU = 1280

// 4. Добавить Capsule handling
type writeCapsule struct {
    capsule appendable
    result  chan error
}
```

### **3. Общие улучшения**

#### **Thread Safety:**
```go
// Использовать atomic операции
type proxiedConn struct {
    closed atomic.Bool
    // ... остальные поля
}

// Использовать proper mutex
type Proxy struct {
    mx       sync.Mutex
    closed   bool
    refCount sync.WaitGroup
    closers  map[io.Closer]struct{}
}
```

#### **Context Management:**
```go
// Правильное управление контекстами
func (c *proxiedConn) SetReadDeadline(t time.Time) error {
    c.deadlineMx.Lock()
    defer c.deadlineMx.Unlock()
    
    oldDeadline := c.deadline
    c.deadline = t
    now := time.Now()
    
    if t.IsZero() {
        if c.readDeadlineTimer != nil && !c.readDeadlineTimer.Stop() {
            <-c.readDeadlineTimer.C
        }
        return nil
    }
    
    if !t.After(now) {
        c.readCtxCancel()
        return nil
    }
    
    // ... остальная логика
}
```

#### **Error Handling:**
```go
// Правильная обработка ошибок
func (c *proxiedConn) ReadFrom(b []byte) (n int, addr net.Addr, err error) {
start:
    c.deadlineMx.Lock()
    ctx := c.readCtx
    c.deadlineMx.Unlock()
    
    data, err := c.str.ReceiveDatagram(ctx)
    if err != nil {
        if !errors.Is(err, context.Canceled) {
            return 0, nil, err
        }
        // The context is cancelled asynchronously
        c.deadlineMx.Lock()
        restart := time.Now().Before(c.deadline)
        c.deadlineMx.Unlock()
        if restart {
            goto start
        }
        return 0, nil, os.ErrDeadlineExceeded
    }
    
    contextID, n, err := quicvarint.Parse(data)
    if err != nil {
        return 0, nil, fmt.Errorf("masque: malformed datagram: %w", err)
    }
    if contextID != 0 {
        goto start
    }
    
    return copy(b, data[n:]), c.remoteAddr, nil
}
```

---

## **Следующие шаги**

### **1. Немедленные улучшения (1-2 дня)**
- [ ] Изучить masque-go client.go детально
- [ ] Изучить masque-go proxy.go детально
- [ ] Изучить masque-go conn.go детально
- [ ] Изучить connect-ip-go client.go детально
- [ ] Изучить connect-ip-go conn.go детально

### **2. Улучшения CONNECT-UDP (2-3 дня)**
- [ ] Улучшить error handling в masque_support.go
- [ ] Добавить Proxy-Status заголовки
- [ ] Улучшить Context ID управление
- [ ] Добавить retry логику
- [ ] Оптимизировать buffer management

### **3. Улучшения CONNECT-IP (2-3 дня)**
- [ ] Улучшить IP Route management
- [ ] Добавить ICMP support
- [ ] Улучшить MTU handling
- [ ] Добавить Capsule handling
- [ ] Добавить IPv6 поддержку

### **4. Тестирование (1-2 дня)**
- [ ] Создать тесты для CONNECT-UDP
- [ ] Создать тесты для CONNECT-IP
- [ ] Создать интеграционные тесты
- [ ] Оптимизировать производительность

---

## **Готово к реализации!**

Теперь у нас есть детальный план улучшений на основе изучения reference репозиториев. Можно приступать к реализации улучшений! 🎯

