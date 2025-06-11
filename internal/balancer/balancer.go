package balancer

import (
	"sync"
	"time"
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
}

// NewBalancer creates a new balancer
func NewBalancer() *Balancer {
	return &Balancer{
		servers: make([]*Server, 0),
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