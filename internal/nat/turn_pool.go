package nat

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
)

// TURNServerPool представляет TURN сервер в пуле
type TURNServerPool struct {
	Address         string
	Username        string
	Password        string
	Realm           string
	Priority        int
	Weight          int
	MaxConnections  int
	ActiveConnections int
	LastHealthCheck time.Time
	IsHealthy       bool
	AverageLatency  time.Duration
	TotalRequests   uint64
	FailedRequests  uint64
}

// TURNPool управляет пулом TURN серверов с балансировкой
type TURNPool struct {
	logger *logger.Logger

	// Серверы в пуле
	servers      map[string]*TURNServerPool
	serversMutex sync.RWMutex

	// Балансировка
	roundRobinIndex int
	roundRobinMutex sync.Mutex

	// Мониторинг
	healthCheckInterval time.Duration
	healthCheckTimeout  time.Duration

	ctx    context.Context
	cancel context.CancelFunc
}

// TURNAllocation представляет TURN allocation
type TURNAllocation struct {
	AllocationID string
	Server       *TURNServerPool
	ClientID     string
	RelayAddress string
	CreatedAt    time.Time
	LastActivity time.Time
	BandwidthIn  uint64
	BandwidthOut uint64
}

// NewTURNPool создает новый TURN pool
func NewTURNPool(logger *logger.Logger) *TURNPool {
	ctx, cancel := context.WithCancel(context.Background())
	tp := &TURNPool{
		logger:              logger,
		servers:             make(map[string]*TURNServerPool),
		healthCheckInterval: 30 * time.Second,
		healthCheckTimeout:  5 * time.Second,
		ctx:                 ctx,
		cancel:              cancel,
	}
	go tp.startHealthChecks()
	return tp
}

// AddServer добавляет TURN сервер в пул
func (tp *TURNPool) AddServer(server *TURNServerPool) {
	tp.serversMutex.Lock()
	defer tp.serversMutex.Unlock()

	tp.servers[server.Address] = server
	tp.logger.Info("Added TURN server to pool",
		zap.String("address", server.Address),
		zap.Int("priority", server.Priority),
		zap.Int("weight", server.Weight),
	)
}

// RemoveServer удаляет TURN сервер из пула
func (tp *TURNPool) RemoveServer(address string) {
	tp.serversMutex.Lock()
	defer tp.serversMutex.Unlock()

	if _, exists := tp.servers[address]; exists {
		delete(tp.servers, address)
		tp.logger.Info("Removed TURN server from pool", zap.String("address", address))
	}
}

// GetBestServer возвращает лучший доступный TURN сервер
func (tp *TURNPool) GetBestServer() (*TURNServerPool, error) {
	tp.serversMutex.RLock()
	defer tp.serversMutex.RUnlock()

	if len(tp.servers) == 0 {
		return nil, fmt.Errorf("no TURN servers available")
	}

	// Сначала ищем здоровые серверы с наивысшим приоритетом
	healthyServers := make([]*TURNServerPool, 0)
	maxPriority := -1

	for _, server := range tp.servers {
		if server.IsHealthy && server.ActiveConnections < server.MaxConnections {
			if server.Priority > maxPriority {
				maxPriority = server.Priority
				healthyServers = healthyServers[:0] // Очищаем список
				healthyServers = append(healthyServers, server)
			} else if server.Priority == maxPriority {
				healthyServers = append(healthyServers, server)
			}
		}
	}

	if len(healthyServers) == 0 {
		return nil, fmt.Errorf("no healthy TURN servers available")
	}

	// Если несколько серверов с одинаковым приоритетом, используем weighted round-robin
	if len(healthyServers) == 1 {
		return healthyServers[0], nil
	}

	return tp.selectByWeight(healthyServers), nil
}

// selectByWeight выбирает сервер по весу (weighted round-robin)
func (tp *TURNPool) selectByWeight(servers []*TURNServerPool) *TURNServerPool {
	tp.roundRobinMutex.Lock()
	defer tp.roundRobinMutex.Unlock()

	// Вычисляем общий вес
	totalWeight := 0
	for _, server := range servers {
		totalWeight += server.Weight
	}

	if totalWeight == 0 {
		// Если веса не заданы, используем простой round-robin
		server := servers[tp.roundRobinIndex%len(servers)]
		tp.roundRobinIndex++
		return server
	}

	// Weighted round-robin
	currentWeight := 0
	for _, server := range servers {
		currentWeight += server.Weight
		if tp.roundRobinIndex < currentWeight {
			tp.roundRobinIndex++
			return server
		}
	}

	// Fallback
	tp.roundRobinIndex = 0
	return servers[0]
}

// AllocateTURN создает TURN allocation
func (tp *TURNPool) AllocateTURN(clientID string) (*TURNAllocation, error) {
	server, err := tp.GetBestServer()
	if err != nil {
		return nil, fmt.Errorf("failed to get TURN server: %w", err)
	}

	// Увеличиваем счетчик активных соединений
	tp.serversMutex.Lock()
	server.ActiveConnections++
	tp.serversMutex.Unlock()

	allocation := &TURNAllocation{
		AllocationID: fmt.Sprintf("alloc-%d", time.Now().UnixNano()),
		Server:       server,
		ClientID:     clientID,
		RelayAddress: tp.generateRelayAddress(server),
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
	}

	tp.logger.Info("Created TURN allocation",
		zap.String("allocation_id", allocation.AllocationID),
		zap.String("client_id", clientID),
		zap.String("server", server.Address),
		zap.String("relay_address", allocation.RelayAddress),
	)

	return allocation, nil
}

// ReleaseTURN освобождает TURN allocation
func (tp *TURNPool) ReleaseTURN(allocation *TURNAllocation) {
	tp.serversMutex.Lock()
	defer tp.serversMutex.Unlock()

	if server, exists := tp.servers[allocation.Server.Address]; exists {
		if server.ActiveConnections > 0 {
			server.ActiveConnections--
		}
	}

	tp.logger.Info("Released TURN allocation",
		zap.String("allocation_id", allocation.AllocationID),
		zap.String("client_id", allocation.ClientID),
		zap.String("server", allocation.Server.Address),
	)
}

// generateRelayAddress генерирует адрес для relay
func (tp *TURNPool) generateRelayAddress(server *TURNServerPool) string {
	// В реальной реализации здесь должна быть логика генерации relay адреса
	// Пока возвращаем фиктивный адрес
	return fmt.Sprintf("relay-%s-%d", server.Address, time.Now().UnixNano())
}

// startHealthChecks запускает проверки здоровья серверов
func (tp *TURNPool) startHealthChecks() {
	ticker := time.NewTicker(tp.healthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-tp.ctx.Done():
			return
		case <-ticker.C:
			tp.performHealthChecks()
		}
	}
}

// performHealthChecks выполняет проверки здоровья всех серверов
func (tp *TURNPool) performHealthChecks() {
	tp.serversMutex.RLock()
	servers := make([]*TURNServerPool, 0, len(tp.servers))
	for _, server := range tp.servers {
		servers = append(servers, server)
	}
	tp.serversMutex.RUnlock()

	for _, server := range servers {
		go tp.checkServerHealth(server)
	}
}

// checkServerHealth проверяет здоровье конкретного сервера
func (tp *TURNPool) checkServerHealth(server *TURNServerPool) {
	_, cancel := context.WithTimeout(tp.ctx, tp.healthCheckTimeout)
	defer cancel()

	start := time.Now()
	
	// В реальной реализации здесь должна быть проверка TURN сервера
	// Пока просто проверяем доступность порта
	conn, err := net.DialTimeout("tcp", server.Address, tp.healthCheckTimeout)
	latency := time.Since(start)
	
	tp.serversMutex.Lock()
	server.LastHealthCheck = time.Now()
	server.AverageLatency = tp.calculateAverageLatency(server.AverageLatency, latency, server.TotalRequests+1)
	
	if err != nil {
		server.IsHealthy = false
		server.FailedRequests++
		tp.logger.Debug("TURN server health check failed",
			zap.String("address", server.Address),
			zap.Error(err),
		)
	} else {
		server.IsHealthy = true
		conn.Close()
		tp.logger.Debug("TURN server health check passed",
			zap.String("address", server.Address),
			zap.Duration("latency", latency),
		)
	}
	server.TotalRequests++
	tp.serversMutex.Unlock()
}

// calculateAverageLatency вычисляет среднюю задержку
func (tp *TURNPool) calculateAverageLatency(current, new time.Duration, count uint64) time.Duration {
	if count == 1 {
		return new
	}
	return time.Duration((int64(current)*int64(count-1) + int64(new)) / int64(count))
}

// GetServerStats возвращает статистику серверов
func (tp *TURNPool) GetServerStats() map[string]interface{} {
	tp.serversMutex.RLock()
	defer tp.serversMutex.RUnlock()

	stats := map[string]interface{}{
		"total_servers":      len(tp.servers),
		"healthy_servers":    0,
		"unhealthy_servers":  0,
		"total_connections":  0,
		"max_connections":    0,
	}

	for _, server := range tp.servers {
		if server.IsHealthy {
			stats["healthy_servers"] = stats["healthy_servers"].(int) + 1
		} else {
			stats["unhealthy_servers"] = stats["unhealthy_servers"].(int) + 1
		}
		stats["total_connections"] = stats["total_connections"].(int) + server.ActiveConnections
		stats["max_connections"] = stats["max_connections"].(int) + server.MaxConnections
	}

	return stats
}

// GetServerList возвращает список всех серверов
func (tp *TURNPool) GetServerList() []*TURNServerPool {
	tp.serversMutex.RLock()
	defer tp.serversMutex.RUnlock()

	servers := make([]*TURNServerPool, 0, len(tp.servers))
	for _, server := range tp.servers {
		servers = append(servers, server)
	}

	return servers
}

// Stop останавливает TURN pool
func (tp *TURNPool) Stop() {
	tp.cancel()
	tp.logger.Info("TURN pool stopped")
}
