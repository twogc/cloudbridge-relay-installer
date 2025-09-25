package nat

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/pion/stun"
	stunClient "github.com/pion/stun"
	"go.uber.org/zap"

	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
)

// STUNClient представляет STUN клиент для сбора NAT информации
type STUNClient struct {
	logger      *logger.Logger
	servers     []string
	timeout     time.Duration
	maxRetries  int
	
	// Кэш результатов
	cache       map[string]*STUNResult
	cacheMutex  sync.RWMutex
	cacheExpiry time.Duration

	// Metrics
	metrics *NATMetrics
}

// STUNResult представляет результат STUN запроса
type STUNResult struct {
	ServerAddr     string
	MappedAddr     *net.UDPAddr
	NATType        NATType
	ResponseTime   time.Duration
	Timestamp      time.Time
	Error          error
}

// NATType представляет тип NAT (RFC 3489/5780)
type NATType int

const (
	NATTypeUnknown NATType = iota
	NATTypeOpen             // No NAT
	NATTypeSymmetric        // Symmetric NAT
	NATTypeCone             // Full Cone NAT
	NATTypeRestrictedCone   // Restricted Cone NAT
	NATTypePortRestricted   // Port Restricted Cone NAT
	NATTypeBlocked          // Blocked/Firewall
)

func (nt NATType) String() string {
	switch nt {
	case NATTypeOpen:
		return "Open"
	case NATTypeSymmetric:
		return "Symmetric"
	case NATTypeCone:
		return "Full Cone"
	case NATTypeRestrictedCone:
		return "Restricted Cone"
	case NATTypePortRestricted:
		return "Port Restricted"
	case NATTypeBlocked:
		return "Blocked"
	default:
		return "Unknown"
	}
}

// NewSTUNClient creates a new STUN client
func NewSTUNClient(logger *logger.Logger, servers []string) *STUNClient {
	return &STUNClient{
		logger:      logger,
		servers:     servers,
		timeout:     5 * time.Second,
		maxRetries:  3,
		cache:       make(map[string]*STUNResult),
		cacheExpiry: 5 * time.Minute,
		metrics:     NewNATMetrics(),
	}
}

// GatherHostCandidates собирает host кандидаты (локальные адреса)
func (sc *STUNClient) GatherHostCandidates() ([]*ICECandidate, error) {
	sc.logger.Debug("Gathering host candidates")
	
	var candidates []*ICECandidate
	
	// Получаем все сетевые интерфейсы
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get network interfaces: %w", err)
	}
	
	for _, iface := range interfaces {
		// Пропускаем loopback и неактивные интерфейсы
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}
		
		addrs, err := iface.Addrs()
		if err != nil {
			sc.logger.Warn("Failed to get addresses for interface", 
				zap.String("interface", iface.Name), 
				zap.Error(err))
			continue
		}
		
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}
			
			// Только IPv4 адреса
			if ipNet.IP.To4() == nil {
				continue
			}
			
			// Пропускаем link-local адреса
			if ipNet.IP.IsLinkLocalUnicast() {
				continue
			}
			
			candidate := &ICECandidate{
				Foundation: fmt.Sprintf("host_%s", ipNet.IP.String()),
				Component:  1, // RTP component
				Transport:  "udp",
				Priority:   calculateHostPriority(ipNet.IP),
				Address:    ipNet.IP.String(),
				Port:       0, // Will be set when binding
				Type:       ICECandidateTypeHost,
			}
			
			candidates = append(candidates, candidate)
			sc.logger.Debug("Host candidate gathered",
				zap.String("address", candidate.Address),
				zap.Uint32("priority", candidate.Priority))
		}
	}
	
	sc.metrics.HostCandidatesGathered = len(candidates)
	sc.logger.Info("Host candidates gathered", zap.Int("count", len(candidates)))
	
	return candidates, nil
}

// GatherServerReflexiveCandidates собирает server-reflexive кандидаты через STUN
func (sc *STUNClient) GatherServerReflexiveCandidates(ctx context.Context) ([]*ICECandidate, error) {
	sc.logger.Debug("Gathering server-reflexive candidates")
	
	var candidates []*ICECandidate
	var wg sync.WaitGroup
	var mu sync.Mutex
	
	for _, server := range sc.servers {
		wg.Add(1)
		go func(serverAddr string) {
			defer wg.Done()
			
			result, err := sc.performSTUNBinding(ctx, serverAddr)
			if err != nil {
				sc.logger.Warn("STUN binding failed", 
					zap.String("server", serverAddr), 
					zap.Error(err))
				return
			}
			
			if result.MappedAddr != nil {
				candidate := &ICECandidate{
					Foundation:  fmt.Sprintf("srflx_%s", result.MappedAddr.IP.String()),
					Component:   1,
					Transport:   "udp",
					Priority:    calculateServerReflexivePriority(result.MappedAddr.IP),
					Address:     result.MappedAddr.IP.String(),
					Port:        result.MappedAddr.Port,
					Type:        ICECandidateTypeServerReflexive,
					RelatedAddr: "", // Local address would go here
					RelatedPort: 0,  // Local port would go here
				}
				
				mu.Lock()
				candidates = append(candidates, candidate)
				mu.Unlock()
				
				sc.logger.Debug("Server-reflexive candidate gathered",
					zap.String("address", candidate.Address),
					zap.Int("port", candidate.Port),
					zap.String("server", serverAddr))
			}
		}(server)
	}
	
	wg.Wait()
	
	sc.metrics.ServerReflexiveCandidatesGathered = len(candidates)
	sc.logger.Info("Server-reflexive candidates gathered", zap.Int("count", len(candidates)))
	
	return candidates, nil
}

// DiscoverNATType определяет тип NAT (упрощенная реализация RFC 5780)
func (sc *STUNClient) DiscoverNATType(ctx context.Context) (NATType, error) {
	sc.logger.Debug("Discovering NAT type")
	
	if len(sc.servers) < 2 {
		sc.logger.Warn("Need at least 2 STUN servers for NAT type discovery")
		return NATTypeUnknown, fmt.Errorf("insufficient STUN servers")
	}
	
	// Test 1: Basic STUN binding to first server
	result1, err := sc.performSTUNBinding(ctx, sc.servers[0])
	if err != nil {
		sc.logger.Error("NAT discovery test 1 failed", zap.Error(err))
		return NATTypeBlocked, err
	}
	
	if result1.MappedAddr == nil {
		return NATTypeBlocked, fmt.Errorf("no mapped address received")
	}
	
	// Test 2: STUN binding to second server
	result2, err := sc.performSTUNBinding(ctx, sc.servers[1])
	if err != nil {
		sc.logger.Warn("NAT discovery test 2 failed", zap.Error(err))
		// Can still determine some NAT types with one server
		return sc.classifyNATType(result1, nil), nil
	}
	
	natType := sc.classifyNATType(result1, result2)
	
	sc.logger.Info("NAT type discovered", 
		zap.String("type", natType.String()),
		zap.String("mapped_addr1", result1.MappedAddr.String()),
		zap.String("mapped_addr2", result2.MappedAddr.String()))
	
	// Update metrics
	sc.metrics.NATType = natType.String()
	sc.updateNATTypeMetrics(natType)
	
	return natType, nil
}

// PerformSTUNBinding выполняет STUN binding запрос (публичный метод)
func (sc *STUNClient) PerformSTUNBinding(ctx context.Context, serverAddr string) (*STUNResult, error) {
	return sc.performSTUNBinding(ctx, serverAddr)
}

// performSTUNBinding выполняет STUN binding запрос
func (sc *STUNClient) performSTUNBinding(ctx context.Context, serverAddr string) (*STUNResult, error) {
	// Check cache first
	sc.cacheMutex.RLock()
	if cached, exists := sc.cache[serverAddr]; exists {
		if time.Since(cached.Timestamp) < sc.cacheExpiry {
			sc.cacheMutex.RUnlock()
			sc.logger.Debug("Using cached STUN result", zap.String("server", serverAddr))
			return cached, cached.Error
		}
	}
	sc.cacheMutex.RUnlock()
	
	result := &STUNResult{
		ServerAddr: serverAddr,
		Timestamp:  time.Now(),
	}
	
	startTime := time.Now()
	
	// Create UDP connection
	conn, err := net.DialTimeout("udp", serverAddr, sc.timeout)
	if err != nil {
		result.Error = fmt.Errorf("failed to connect to STUN server: %w", err)
		sc.cacheResult(serverAddr, result)
		return result, result.Error
	}
	defer conn.Close()
	
	// Set deadline
	deadline, ok := ctx.Deadline()
	if ok {
		conn.SetDeadline(deadline)
	} else {
		conn.SetDeadline(time.Now().Add(sc.timeout))
	}
	
	// Create STUN client
	client, err := stunClient.NewClient(conn)
	if err != nil {
		result.Error = fmt.Errorf("failed to create STUN client: %w", err)
		sc.cacheResult(serverAddr, result)
		return result, result.Error
	}
	defer client.Close()
	
	// Create binding request
	message := stun.MustBuild(stun.TransactionID, stun.BindingRequest)
	
	var mappedAddr stun.XORMappedAddress
	
	// Send request and wait for response
	err = client.Do(message, func(res stun.Event) {
		if res.Error != nil {
			result.Error = fmt.Errorf("STUN request failed: %w", res.Error)
			return
		}
		
		// Parse XOR-MAPPED-ADDRESS
		if parseErr := mappedAddr.GetFrom(res.Message); parseErr != nil {
			// Try MAPPED-ADDRESS as fallback
			var legacyAddr stun.MappedAddress
			if legacyErr := legacyAddr.GetFrom(res.Message); legacyErr != nil {
				result.Error = fmt.Errorf("failed to parse mapped address: %w", parseErr)
				return
			}
			mappedAddr.IP = legacyAddr.IP
			mappedAddr.Port = legacyAddr.Port
		}
		
		result.MappedAddr = &net.UDPAddr{
			IP:   mappedAddr.IP,
			Port: mappedAddr.Port,
		}
		result.ResponseTime = time.Since(startTime)
	})
	
	if err != nil {
		result.Error = fmt.Errorf("STUN transaction failed: %w", err)
	}
	
	// Cache result
	sc.cacheResult(serverAddr, result)
	
	// Update metrics
	sc.metrics.STUNRequests++
	if result.Error != nil {
		sc.metrics.STUNFailures++
	} else {
		sc.metrics.STUNSuccesses++
		sc.metrics.AverageSTUNResponseTime = 
			(sc.metrics.AverageSTUNResponseTime + result.ResponseTime) / 2
	}
	
	return result, result.Error
}

// classifyNATType классифицирует тип NAT на основе результатов STUN
func (sc *STUNClient) classifyNATType(result1, result2 *STUNResult) NATType {
	// Get local address
	localAddr, err := sc.getLocalAddress()
	if err != nil {
		sc.logger.Warn("Failed to get local address", zap.Error(err))
		return NATTypeUnknown
	}
	
	// If mapped address equals local address, no NAT
	if result1.MappedAddr.IP.Equal(localAddr.IP) {
		return NATTypeOpen
	}
	
	// If we only have one result, we can only determine basic NAT presence
	if result2 == nil || result2.MappedAddr == nil {
		return NATTypeUnknown // Some NAT, but can't classify type
	}
	
	// Compare mapped addresses from different servers
	if result1.MappedAddr.IP.Equal(result2.MappedAddr.IP) && 
	   result1.MappedAddr.Port == result2.MappedAddr.Port {
		// Same external address from different servers = Cone NAT
		return NATTypeCone
	} else {
		// Different external addresses = Symmetric NAT
		return NATTypeSymmetric
	}
}

// getLocalAddress получает локальный IP адрес
func (sc *STUNClient) getLocalAddress() (*net.UDPAddr, error) {
	// Create a connection to determine local address
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr, nil
}

// cacheResult кэширует результат STUN запроса
func (sc *STUNClient) cacheResult(serverAddr string, result *STUNResult) {
	sc.cacheMutex.Lock()
	defer sc.cacheMutex.Unlock()
	sc.cache[serverAddr] = result
}

// updateNATTypeMetrics обновляет метрики типа NAT
func (sc *STUNClient) updateNATTypeMetrics(natType NATType) {
	switch natType {
	case NATTypeOpen:
		sc.metrics.OpenNATCount++
	case NATTypeSymmetric:
		sc.metrics.SymmetricNATCount++
	case NATTypeCone:
		sc.metrics.ConeNATCount++
	case NATTypeBlocked:
		sc.metrics.BlockedNATCount++
	}
}

// calculateHostPriority вычисляет приоритет host кандидата
func calculateHostPriority(ip net.IP) uint32 {
	// RFC 8445: priority = (2^24)*(type preference) + 
	//                    (2^8)*(local preference) + 
	//                    (2^0)*(component ID)
	
	typePreference := uint32(126) // Host candidate type preference
	localPreference := uint32(65535) // Max local preference
	
	// Adjust local preference based on IP type
	if ip.IsPrivate() {
		localPreference = 32768 // Lower preference for private IPs
	}
	
	return (1<<24)*typePreference + (1<<8)*localPreference + 1
}

// calculateServerReflexivePriority вычисляет приоритет server-reflexive кандидата
func calculateServerReflexivePriority(ip net.IP) uint32 {
	typePreference := uint32(100) // Server-reflexive type preference
	localPreference := uint32(65535)
	
	return (1<<24)*typePreference + (1<<8)*localPreference + 1
}

// GetMetrics возвращает метрики NAT
func (sc *STUNClient) GetMetrics() *NATMetrics {
	return sc.metrics
}

// ClearCache очищает кэш результатов
func (sc *STUNClient) ClearCache() {
	sc.cacheMutex.Lock()
	defer sc.cacheMutex.Unlock()
	sc.cache = make(map[string]*STUNResult)
	sc.logger.Debug("STUN cache cleared")
}