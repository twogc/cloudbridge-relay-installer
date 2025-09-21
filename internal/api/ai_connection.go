package api

import (
	"time"
)

// AIConnection представляет соединение для AI анализа
type AIConnection struct {
	ID               string
	RemoteAddr       string
	Source           string
	Destination      string
	Protocol         string
	PacketsPerSecond int
	BytesPerSecond   int64
	ConnectionCount  int
	Latency          float64
	Bandwidth        float64
	ThrottleFactor   float64
	Route            []string
	Timestamp        time.Time
}

// NewAIConnection создает новое AI соединение
func NewAIConnection(id, remoteAddr, source, destination, protocol string) *AIConnection {
	return &AIConnection{
		ID:          id,
		RemoteAddr:  remoteAddr,
		Source:      source,
		Destination: destination,
		Protocol:    protocol,
		Timestamp:   time.Now(),
	}
}

// SetMetrics устанавливает метрики соединения
func (c *AIConnection) SetMetrics(packetsPerSecond int, bytesPerSecond int64, connectionCount int, latency, bandwidth float64) {
	c.PacketsPerSecond = packetsPerSecond
	c.BytesPerSecond = bytesPerSecond
	c.ConnectionCount = connectionCount
	c.Latency = latency
	c.Bandwidth = bandwidth
}

// SetRoute устанавливает маршрут соединения
func (c *AIConnection) SetRoute(route []string) {
	c.Route = route
}

// SetThrottleFactor устанавливает коэффициент ограничения
func (c *AIConnection) SetThrottleFactor(factor float64) {
	c.ThrottleFactor = factor
}

// GetAge возвращает возраст соединения
func (c *AIConnection) GetAge() time.Duration {
	return time.Since(c.Timestamp)
}

// IsExpired проверяет, истекло ли соединение
func (c *AIConnection) IsExpired(maxAge time.Duration) bool {
	return c.GetAge() > maxAge
}

// GetMetrics возвращает метрики соединения в виде карты
func (c *AIConnection) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"id":                 c.ID,
		"remote_addr":        c.RemoteAddr,
		"source":             c.Source,
		"destination":        c.Destination,
		"protocol":           c.Protocol,
		"packets_per_second": c.PacketsPerSecond,
		"bytes_per_second":   c.BytesPerSecond,
		"connection_count":   c.ConnectionCount,
		"latency":            c.Latency,
		"bandwidth":          c.Bandwidth,
		"throttle_factor":    c.ThrottleFactor,
		"route":              c.Route,
		"timestamp":          c.Timestamp,
		"age":                c.GetAge().String(),
	}
}
