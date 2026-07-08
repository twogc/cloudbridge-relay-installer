# Следующие шаги для MASQUE улучшений

## **Текущий прогресс**

### **Завершено**
- [x] **Изучение masque-go reference** - client.go, proxy.go, conn.go, request.go
- [x] **Создание улучшенной версии masque_support.go** - на основе masque-go подхода
- [x] **Анализ ключевых улучшений** - error handling, Proxy-Status, Context ID управление

### 🔄 **В процессе**
- [ ] **Изучение connect-ip-go reference** - client.go, conn.go, proxy.go
- [ ] **Улучшение CONNECT-UDP** - интеграция masque-go подхода

### **Ожидает**
- [ ] **Улучшение CONNECT-IP** - интеграция connect-ip-go подхода
- [ ] **Создание тестов** - для улучшенного MASQUE

---

## 🔧 **Детальный план улучшений**

### **1. Завершить изучение connect-ip-go (1 день)**

#### **Что изучить:**
```bash
cd reference-repos/masque/connect-ip-go/
```

**Ключевые файлы:**
- **client.go** - CONNECT-IP клиент
- **conn.go** - IP соединения и управление
- **proxy.go** - CONNECT-IP прокси
- **capsule.go** - обработка капсул
- **icmp.go** - ICMP поддержка
- **iprange.go** - управление IP диапазонами

#### **Ключевые особенности для изучения:**
- **IP Route management** - управление IP маршрутами
- **Address assignment** - назначение IP адресов
- **Capsule handling** - обработка капсул
- **ICMP support** - поддержка ICMP пакетов
- **MTU handling** - обработка MTU
- **IPv6 support** - поддержка IPv6

### **2. Улучшить CONNECT-UDP (2 дня)**

#### **Текущие улучшения в masque_support_improved.go:**
- **Error handling** - errToStatus функция из masque-go
- **Proxy-Status заголовки** - RFC 9209 совместимые заголовки
- **Context ID управление** - правильное управление Context ID
- **Thread safety** - atomic операции и mutex
- **Goroutine management** - правильное управление горутинами

#### **Дополнительные улучшения:**
- [ ] **Интеграция с uritemplate** - как в masque-go request.go
- [ ] **Улучшенный parsing** - полная реализация ParseRequest
- [ ] **DNS error handling** - dnsErrorToProxyStatus функция
- [ ] **Deadline management** - как в masque-go conn.go
- [ ] **Capsule skipping** - skipCapsules функция

#### **План реализации:**
```go
// 1. Добавить uritemplate поддержку
import "github.com/yosida95/uritemplate/v3"

// 2. Улучшить parsing
func (ms *MASQUEServer) parseConnectUDPRequest(r *http.Request) (*masque.Request, error) {
    template := uritemplate.MustNew("/.well-known/masque/udp/{target_host}:{target_port}")
    return masque.ParseRequest(r, template)
}

// 3. Добавить deadline management
func (mc *MASQUEConnection) SetReadDeadline(t time.Time) error {
    mc.deadlineMx.Lock()
    defer mc.deadlineMx.Unlock()
    
    oldDeadline := mc.deadline
    mc.deadline = t
    now := time.Now()
    
    if t.IsZero() {
        if mc.readDeadlineTimer != nil && !mc.readDeadlineTimer.Stop() {
            <-mc.readDeadlineTimer.C
        }
        return nil
    }
    
    if !t.After(now) {
        mc.readCtxCancel()
        return nil
    }
    
    deadline := t.Sub(now)
    if mc.readDeadlineTimer != nil {
        if now.Before(oldDeadline) {
            mc.readCtxCancel()
            mc.readCtx, mc.readCtxCancel = context.WithCancel(context.Background())
        }
        mc.readDeadlineTimer.Reset(deadline)
    } else {
        mc.readDeadlineTimer = time.AfterFunc(deadline, func() {
            mc.deadlineMx.Lock()
            defer mc.deadlineMx.Unlock()
            if !mc.deadline.IsZero() && mc.deadline.Before(time.Now()) {
                mc.readCtxCancel()
            }
        })
    }
    return nil
}
```

### **3. Улучшить CONNECT-IP (2 дня)**

#### **Ключевые улучшения на основе connect-ip-go:**
- [ ] **IP Route management** - управление IP маршрутами
- [ ] **Address assignment** - назначение IP адресов
- [ ] **Capsule handling** - обработка капсул
- [ ] **ICMP support** - поддержка ICMP пакетов
- [ ] **MTU handling** - обработка MTU
- [ ] **IPv6 support** - полная поддержка IPv6

#### **План реализации:**
```go
// 1. Добавить IP Route management
type IPRoute struct {
    Prefix  netip.Prefix
    Gateway netip.Addr
    Metric  int
}

// 2. Добавить Address assignment
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

// 3. Добавить ICMP support
const (
    ipProtoICMP   = 1
    ipProtoICMPv6 = 58
)

// 4. Добавить MTU handling
const minMTU = 1280

// 5. Добавить Capsule handling
type writeCapsule struct {
    capsule appendable
    result  chan error
}
```

### **4. Создать тесты (1 день)**

#### **Тесты для CONNECT-UDP:**
```go
func TestConnectUDP(t *testing.T) {
    // Test basic CONNECT-UDP functionality
    // Test error handling
    // Test Context ID management
    // Test deadline handling
}

func TestConnectUDPProxy(t *testing.T) {
    // Test proxy functionality
    // Test DNS error handling
    // Test Proxy-Status headers
}
```

#### **Тесты для CONNECT-IP:**
```go
func TestConnectIP(t *testing.T) {
    // Test basic CONNECT-IP functionality
    // Test IP route management
    // Test address assignment
    // Test ICMP support
}

func TestConnectIPProxy(t *testing.T) {
    // Test proxy functionality
    // Test IPv6 support
    // Test MTU handling
}
```

#### **Интеграционные тесты:**
```go
func TestMASQUEIntegration(t *testing.T) {
    // Test full MASQUE functionality
    // Test CONNECT-UDP + CONNECT-IP
    // Test HTTP Datagrams
    // Test error scenarios
}
```

---

## **Критерии успеха**

### **CONNECT-UDP улучшения:**
- [ ] Error handling работает как в masque-go
- [ ] Proxy-Status заголовки работают
- [ ] Context ID управление работает
- [ ] Deadline management работает
- [ ] Thread safety обеспечена

### **CONNECT-IP улучшения:**
- [ ] IP Route management работает
- [ ] Address assignment работает
- [ ] ICMP support работает
- [ ] IPv6 support работает
- [ ] MTU handling работает

### **Тестирование:**
- [ ] Unit тесты покрывают 80%+ кода
- [ ] Интеграционные тесты работают
- [ ] Error scenarios покрыты
- [ ] Performance тесты проходят

---

## **Следующие шаги**

### **Немедленно (сегодня):**
1. **Завершить изучение connect-ip-go** - изучить все ключевые файлы
2. **Проанализировать различия** - между masque-go и connect-ip-go подходами
3. **Создать план интеграции** - как объединить оба подхода

### **Завтра:**
1. **Улучшить CONNECT-UDP** - интегрировать masque-go подход
2. **Добавить uritemplate поддержку** - для правильного parsing
3. **Улучшить error handling** - добавить все типы ошибок

### **Послезавтра:**
1. **Улучшить CONNECT-IP** - интегрировать connect-ip-go подход
2. **Добавить IP Route management** - для CONNECT-IP
3. **Добавить ICMP support** - для полной функциональности

### **В конце недели:**
1. **Создать тесты** - для всех улучшений
2. **Интегрировать компоненты** - объединить все улучшения
3. **Оптимизировать производительность** - финальная оптимизация

---

## **Чек-лист**

### **Изучение (сегодня):**
- [ ] Изучить connect-ip-go client.go
- [ ] Изучить connect-ip-go conn.go
- [ ] Изучить connect-ip-go proxy.go
- [ ] Изучить connect-ip-go capsule.go
- [ ] Изучить connect-ip-go icmp.go
- [ ] Изучить connect-ip-go iprange.go

### **CONNECT-UDP улучшения (завтра):**
- [ ] Добавить uritemplate поддержку
- [ ] Улучшить parsing
- [ ] Добавить DNS error handling
- [ ] Добавить deadline management
- [ ] Добавить capsule skipping

### **CONNECT-IP улучшения (послезавтра):**
- [ ] Добавить IP Route management
- [ ] Добавить Address assignment
- [ ] Добавить ICMP support
- [ ] Добавить IPv6 support
- [ ] Добавить MTU handling

### **Тестирование (в конце недели):**
- [ ] Создать unit тесты
- [ ] Создать интеграционные тесты
- [ ] Создать performance тесты
- [ ] Оптимизировать производительность

**Готово к реализации! 🎯**

