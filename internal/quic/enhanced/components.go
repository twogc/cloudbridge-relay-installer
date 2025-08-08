package quic

import (
	"fmt"
	"math"
	"sync"
	"time"
)

// DDoSProtection обеспечивает защиту от DDoS атак
type DDoSProtection struct {
	rateLimiters    map[string]interface{}
	blacklist       map[string]time.Time
	anomalyDetector *AnomalyDetector
	logger          interface{}
	mu              sync.RWMutex
}

// AnomalyDetector обнаруживает аномальное поведение
type AnomalyDetector struct {
	patterns    map[string]*BehaviorPattern
	thresholds  map[string]float64
	logger      interface{}
	mu          sync.RWMutex
}

// BehaviorPattern описывает паттерн поведения клиента
type BehaviorPattern struct {
	IP              string
	TenantID        string
	RequestCount    int64
	LastRequest     time.Time
	AverageInterval time.Duration
	SuspiciousScore float64
}

// GeographicRouter обеспечивает географическую маршрутизацию
type GeographicRouter struct {
	edgeNodes    map[string]*EdgeNode
	geoDatabase  *GeoDatabase
	loadBalancer *LoadBalancer
	logger       interface{}
	mu           sync.RWMutex
}

// EdgeNode представляет edge узел
type EdgeNode struct {
	ID          string
	Address     string
	Location    *GeoLocation
	Load        float64
	Latency     time.Duration
	Healthy     bool
	Connections int64
	LastCheck   time.Time
}

// GeoLocation представляет географическое местоположение
type GeoLocation struct {
	Country     string
	Region      string
	City        string
	Latitude    float64
	Longitude   float64
}

// GeoDatabase предоставляет информацию о географическом местоположении
type GeoDatabase struct {
	// Пока простой in-memory кэш
	ipLocations map[string]*GeoLocation
	mu          sync.RWMutex
}

// LoadBalancer обеспечивает балансировку нагрузки
type LoadBalancer struct {
	nodes []*EdgeNode
	mu    sync.RWMutex
}

// ConnectionPool управляет пулом соединений
type ConnectionPool struct {
	connections map[string]*PooledConnection
	maxSize     int
	idleTimeout time.Duration
	mu          sync.RWMutex
}

// PooledConnection представляет соединение в пуле
type PooledConnection struct {
	ID        string
	Address   string
	Created   time.Time
	LastUsed  time.Time
	UseCount  int64
	Healthy   bool
}

// NewDDoSProtection создает новый DDoS защитник
func NewDDoSProtection(logger interface{}) *DDoSProtection {
	return &DDoSProtection{
		rateLimiters: make(map[string]interface{}),
		blacklist:    make(map[string]time.Time),
		anomalyDetector: &AnomalyDetector{
			patterns:   make(map[string]*BehaviorPattern),
			thresholds: make(map[string]float64),
			logger:     logger,
		},
		logger: logger,
	}
}

// IsAllowed проверяет, разрешено ли соединение
func (ddp *DDoSProtection) IsAllowed(clientIP string, tenantID string) bool {
	ddp.mu.RLock()
	
	// Проверка blacklist
	if blockedUntil, exists := ddp.blacklist[clientIP]; exists {
		if time.Now().Before(blockedUntil) {
			ddp.mu.RUnlock()
			fmt.Printf("Connection blocked by blacklist: %s for tenant %s\n", clientIP, tenantID)
			return false
		}
		// Удаляем из blacklist если время истекло
		ddp.mu.RUnlock()
		ddp.mu.Lock()
		delete(ddp.blacklist, clientIP)
		ddp.mu.Unlock()
		ddp.mu.RLock()
	}
	
	// Упрощенная проверка rate limit
	ddp.mu.RUnlock()
	
	// Анализ аномалий
	if ddp.anomalyDetector.IsAnomalous(clientIP, tenantID) {
		ddp.mu.Lock()
		ddp.blacklist[clientIP] = time.Now().Add(10 * time.Minute)
		ddp.mu.Unlock()
		
		fmt.Printf("Connection blocked by anomaly detection: %s for tenant %s\n", clientIP, tenantID)
		return false
	}
	
	return true
}

// IsAnomalous проверяет, является ли поведение аномальным
func (ad *AnomalyDetector) IsAnomalous(clientIP string, tenantID string) bool {
	ad.mu.Lock()
	defer ad.mu.Unlock()
	
	pattern := ad.patterns[clientIP]
	if pattern == nil {
		pattern = &BehaviorPattern{
			IP:       clientIP,
			TenantID: tenantID,
		}
		ad.patterns[clientIP] = pattern
	}
	
	now := time.Now()
	pattern.RequestCount++
	
	// Обновление среднего интервала
	if !pattern.LastRequest.IsZero() {
		interval := now.Sub(pattern.LastRequest)
		if pattern.AverageInterval == 0 {
			pattern.AverageInterval = interval
		} else {
			pattern.AverageInterval = (pattern.AverageInterval + interval) / 2
		}
	}
	
	pattern.LastRequest = now
	
	// Простая логика обнаружения аномалий
	threshold := ad.thresholds[tenantID]
	if threshold == 0 {
		threshold = 100 // По умолчанию
	}
	
	// Подозрительно если много запросов за короткое время
	if pattern.RequestCount > int64(threshold) && pattern.AverageInterval < 100*time.Millisecond {
		pattern.SuspiciousScore += 0.5
	} else {
		pattern.SuspiciousScore = math.Max(0, pattern.SuspiciousScore-0.1)
	}
	
	return pattern.SuspiciousScore > 2.0
}

// NewGeographicRouter создает новый географический роутер
func NewGeographicRouter(logger interface{}) *GeographicRouter {
	return &GeographicRouter{
		edgeNodes: make(map[string]*EdgeNode),
		geoDatabase: &GeoDatabase{
			ipLocations: make(map[string]*GeoLocation),
		},
		loadBalancer: &LoadBalancer{},
		logger:       logger,
	}
}

// FindOptimalAddress находит оптимальный адрес для клиента
func (gr *GeographicRouter) FindOptimalAddress(clientIP string, tenantID string) (string, error) {
	gr.mu.RLock()
	defer gr.mu.RUnlock()
	
	// Географическое определение клиента
	clientLocation := gr.geoDatabase.LookupIP(clientIP)
	
	// Поиск ближайших узлов
	candidates := gr.findNearbyNodes(clientLocation, 1000) // 1000km радиус
	
	// Фильтрация по здоровью и нагрузке
	healthyNodes := gr.filterHealthyNodes(candidates)
	
	// Выбор оптимального узла с учетом лимитов арендатора
	optimalNode := gr.selectOptimalNode(healthyNodes, tenantID)
	
	if optimalNode == nil {
		return "", fmt.Errorf("no optimal node found for tenant %s", tenantID)
	}
	
	return optimalNode.Address, nil
}

// findNearbyNodes находит ближайшие узлы
func (gr *GeographicRouter) findNearbyNodes(location *GeoLocation, radiusKm int) []*EdgeNode {
	var nearby []*EdgeNode
	
	for _, node := range gr.edgeNodes {
		distance := gr.calculateDistance(location, node.Location)
		if distance <= float64(radiusKm) {
			nearby = append(nearby, node)
		}
	}
	
	return nearby
}

// filterHealthyNodes фильтрует здоровые узлы
func (gr *GeographicRouter) filterHealthyNodes(nodes []*EdgeNode) []*EdgeNode {
	var healthy []*EdgeNode
	
	for _, node := range nodes {
		if node.Healthy && time.Since(node.LastCheck) < 5*time.Minute {
			healthy = append(healthy, node)
		}
	}
	
	return healthy
}

// selectOptimalNode выбирает оптимальный узел
func (gr *GeographicRouter) selectOptimalNode(nodes []*EdgeNode, tenantID string) *EdgeNode {
	if len(nodes) == 0 {
		return nil
	}
	
	// Простой алгоритм выбора: минимальная задержка + нагрузка
	var bestNode *EdgeNode
	bestScore := math.MaxFloat64
	
	for _, node := range nodes {
		score := float64(node.Latency.Milliseconds()) + node.Load*100
		if score < bestScore {
			bestScore = score
			bestNode = node
		}
	}
	
	return bestNode
}

// calculateDistance вычисляет расстояние между двумя точками
func (gr *GeographicRouter) calculateDistance(loc1, loc2 *GeoLocation) float64 {
	if loc1 == nil || loc2 == nil {
		return math.MaxFloat64
	}
	
	// Простая формула гаверсинуса
	const R = 6371 // Радиус Земли в км
	
	lat1 := loc1.Latitude * math.Pi / 180
	lat2 := loc2.Latitude * math.Pi / 180
	deltaLat := (loc2.Latitude - loc1.Latitude) * math.Pi / 180
	deltaLon := (loc2.Longitude - loc1.Longitude) * math.Pi / 180
	
	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	
	return R * c
}

// LookupIP ищет географическое местоположение по IP
func (gd *GeoDatabase) LookupIP(ip string) *GeoLocation {
	gd.mu.RLock()
	defer gd.mu.RUnlock()
	
	location, exists := gd.ipLocations[ip]
	if exists {
		return location
	}
	
	// Пока возвращаем дефолтное местоположение
	// В реальной реализации здесь был бы запрос к GeoIP базе данных
	return &GeoLocation{
		Country:  "Unknown",
		Region:   "Unknown",
		City:     "Unknown",
		Latitude: 0,
		Longitude: 0,
	}
}

// NewConnectionPool создает новый пул соединений
func NewConnectionPool(maxSize int, idleTimeout time.Duration) *ConnectionPool {
	return &ConnectionPool{
		connections: make(map[string]*PooledConnection),
		maxSize:     maxSize,
		idleTimeout: idleTimeout,
	}
}

// GetConnection получает соединение из пула
func (cp *ConnectionPool) GetConnection(id string) *PooledConnection {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	
	conn, exists := cp.connections[id]
	if exists {
		conn.LastUsed = time.Now()
		conn.UseCount++
		return conn
	}
	
	return nil
}

// PutConnection помещает соединение в пул
func (cp *ConnectionPool) PutConnection(conn *PooledConnection) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	
	if len(cp.connections) >= cp.maxSize {
		// Удаляем самое старое соединение
		var oldest *PooledConnection
		var oldestID string
		
		for id, c := range cp.connections {
			if oldest == nil || c.LastUsed.Before(oldest.LastUsed) {
				oldest = c
				oldestID = id
			}
		}
		
		if oldest != nil {
			delete(cp.connections, oldestID)
		}
	}
	
	cp.connections[conn.ID] = conn
}

// Cleanup удаляет неактивные соединения
func (cp *ConnectionPool) Cleanup() {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	
	now := time.Now()
	for id, conn := range cp.connections {
		if now.Sub(conn.LastUsed) > cp.idleTimeout {
			delete(cp.connections, id)
		}
	}
} 