package health

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// Status represents the health status
type Status struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Uptime    string    `json:"uptime"`
	Version   string    `json:"version"`
}

// Checker handles health checks
type Checker struct {
	startTime time.Time
	version   string
	mu        sync.RWMutex
	status    Status
}

// NewChecker creates a new health checker
func NewChecker(version string) *Checker {
	now := time.Now()
	return &Checker{
		startTime: now,
		version:   version,
		status: Status{
			Status:    "ok",
			Timestamp: now,
			Uptime:    "0s",
			Version:   version,
		},
	}
}

// Handler returns an HTTP handler for health checks
func (c *Checker) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.mu.RLock()
		status := c.status
		c.mu.RUnlock()

		// Update uptime
		status.Uptime = time.Since(c.startTime).String()
		status.Timestamp = time.Now()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	})
}

// SetStatus updates the health status
func (c *Checker) SetStatus(status string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.status.Status = status
} 