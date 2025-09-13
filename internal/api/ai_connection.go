package api

import (
	"time"
)

// Connection представляет соединение для AI анализа
type Connection struct {
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

// NewConnection создает новое соединение
func NewConnection(id, remoteAddr, source, destination, protocol string) *Connection {
	return &Connection{
		ID:          id,
		RemoteAddr:  remoteAddr,
		Source:      source,
		Destination: destination,
		Protocol:    protocol,
		Timestamp:   time.Now(),
	}
}

// SetMetrics устанавливает метрики соединения
func (c *Connection) SetMetrics(packetsPerSecond int, bytesPerSecond int64, connectionCount int, latency, bandwidth float64) {
	c.PacketsPerSecond = packetsPerSecond
	c.BytesPerSecond = bytesPerSecond
	c.ConnectionCount = connectionCount
	c.Latency = latency
	c.Bandwidth = bandwidth
}

// SetRoute устанавливает маршрут соединения
func (c *Connection) SetRoute(route []string) {
	c.Route = route
}

// SetThrottleFactor устанавливает коэффициент ограничения
func (c *Connection) SetThrottleFactor(factor float64) {
	c.ThrottleFactor = factor
}

// GetAge возвращает возраст соединения
func (c *Connection) GetAge() time.Duration {
	return time.Since(c.Timestamp)
}

// IsExpired проверяет, истекло ли соединение
func (c *Connection) IsExpired(maxAge time.Duration) bool {
	return c.GetAge() > maxAge
}

// GetMetrics возвращает метрики соединения в виде карты
func (c *Connection) GetMetrics() map[string]interface{} {
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
