package balancer

import (
	"sync"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/internal/quic"
)

// Server represents a relay server in the cluster
type Server struct {
	Address     string
	ActiveConns int32
	Latency     time.Duration
	LastCheck   time.Time
	Healthy     bool
}

// Balancer manages multiple relay servers
type Balancer struct {
	servers    []*Server
	mu         sync.RWMutex
	nextServer int
	
	// Enhanced QUIC Geographic Router
	enhancedGeoRouter *quic.GeographicRouter
	enhancedEnabled   bool
}

// NewBalancer creates a new balancer
func NewBalancer() *Balancer {
	return &Balancer{
		servers: make([]*Server, 0),
		enhancedEnabled: false, // По умолчанию отключено
	}
}

// NewBalancerWithEnhanced создает balancer с Enhanced Geographic Router
func NewBalancerWithEnhanced() *Balancer {
	return &Balancer{
		servers: make([]*Server, 0),
		enhancedGeoRouter: quic.NewGeographicRouter(),
		enhancedEnabled: true,
	}
}

// AddServer adds a new server to the balancer
func (b *Balancer) AddServer(address string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	server := &Server{
		Address:   address,
		LastCheck: time.Now(),
		Healthy:   true,
	}
	b.servers = append(b.servers, server)
}

// RemoveServer removes a server from the balancer
func (b *Balancer) RemoveServer(address string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for i, server := range b.servers {
		if server.Address == address {
			b.servers = append(b.servers[:i], b.servers[i+1:]...)
			break
		}
	}
}

// GetNextServer returns the next available server using round-robin
func (b *Balancer) GetNextServer() *Server {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if len(b.servers) == 0 {
		return nil
	}

	// Find next healthy server
	start := b.nextServer
	for i := 0; i < len(b.servers); i++ {
		server := b.servers[(start+i)%len(b.servers)]
		if server.Healthy {
			b.nextServer = (start + i + 1) % len(b.servers)
			return server
		}
	}

	return nil
}

// GetServerByGeo возвращает сервер на основе географического маршрута
func (b *Balancer) GetServerByGeo(clientIP string) *Server {
	b.mu.RLock()
	defer b.mu.RUnlock()

	// Используем Enhanced Geographic Router, если доступен
	if b.enhancedEnabled && b.enhancedGeoRouter != nil {
		route := b.enhancedGeoRouter.RouteByLocation(clientIP)
		// Находим сервер по маршруту
		for _, server := range b.servers {
			if server.Address == route && server.Healthy {
				return server
			}
		}
	}
	
	// Fallback к обычной балансировке
	return b.GetNextServer()
}

// UpdateServerStats updates server statistics
func (b *Balancer) UpdateServerStats(address string, activeConns int32, latency time.Duration) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, server := range b.servers {
		if server.Address == address {
			server.ActiveConns = activeConns
			server.Latency = latency
			server.LastCheck = time.Now()
			break
		}
	}
}

// MarkServerUnhealthy marks a server as unhealthy
func (b *Balancer) MarkServerUnhealthy(address string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, server := range b.servers {
		if server.Address == address {
			server.Healthy = false
			break
		}
	}
}

// MarkServerHealthy marks a server as healthy
func (b *Balancer) MarkServerHealthy(address string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, server := range b.servers {
		if server.Address == address {
			server.Healthy = true
			server.LastCheck = time.Now()
			break
		}
	}
}

// GetServers returns all servers
func (b *Balancer) GetServers() []*Server {
	b.mu.RLock()
	defer b.mu.RUnlock()

	servers := make([]*Server, len(b.servers))
	copy(servers, b.servers)
	return servers
}

// InitEnhancedGeoRouter инициализирует Enhanced Geographic Router
func (b *Balancer) InitEnhancedGeoRouter() {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	b.enhancedGeoRouter = quic.NewGeographicRouter()
	b.enhancedEnabled = true
}

// EnableEnhanced включает Enhanced Geographic Router
func (b *Balancer) EnableEnhanced() {
	if !b.enhancedEnabled {
		b.InitEnhancedGeoRouter()
	}
}

// DisableEnhanced отключает Enhanced Geographic Router
func (b *Balancer) DisableEnhanced() {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	b.enhancedEnabled = false
	b.enhancedGeoRouter = nil
}

// IsEnhancedEnabled проверяет, включен ли Enhanced режим
func (b *Balancer) IsEnhancedEnabled() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.enhancedEnabled
}

// GetEnhancedGeoRouter возвращает Enhanced Geographic Router
func (b *Balancer) GetEnhancedGeoRouter() *quic.GeographicRouter {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.enhancedGeoRouter
} 