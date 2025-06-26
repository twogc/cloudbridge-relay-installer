package health

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/internal/metrics"
)

// Status represents the health check status
type Status struct {
	Status    string            `json:"status"`
	Version   string           `json:"version"`
	Uptime    string           `json:"uptime"`
	Metrics   map[string]int64 `json:"metrics"`
	Timestamp time.Time        `json:"timestamp"`
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

// Handler handles health check requests
type Handler struct {
	metrics *metrics.Metrics
	startTime time.Time
	version string
}

// NewHandler creates a new health check handler
func NewHandler(metrics *metrics.Metrics, version string) *Handler {
	return &Handler{
		metrics: metrics,
		startTime: time.Now(),
		version: version,
	}
}

// ServeHTTP implements http.Handler
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	status := Status{
		Status:    "ok",
		Version:   h.version,
		Uptime:    time.Since(h.startTime).String(),
		Metrics:   make(map[string]int64),
		Timestamp: time.Now(),
	}

	// Add metrics
	status.Metrics["active_connections"] = h.metrics.GetActiveConnections()
	status.Metrics["active_tunnels"] = h.metrics.GetActiveTunnels()

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Encode response
	if err := json.NewEncoder(w).Encode(status); err != nil {
		http.Error(w, "Failed to encode health status", http.StatusInternalServerError)
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