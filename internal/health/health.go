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
	Status    string           `json:"status"`
	Version   string           `json:"version"`
	Uptime    string           `json:"uptime"`
	Metrics   map[string]int64 `json:"metrics"`
	Modules   ModuleStatus     `json:"modules"`
	Timestamp time.Time        `json:"timestamp"`
}

// ModuleStatus represents the status of individual modules
type ModuleStatus struct {
	EnhancedQUIC EnhancedQUICStatus `json:"enhanced_quic"`
	P2PQUIC      P2PQUICStatus      `json:"p2p_quic"`
	AI           AIStatus           `json:"ai"`
}

// EnhancedQUICStatus represents Enhanced QUIC module status
type EnhancedQUICStatus struct {
	Enabled bool   `json:"enabled"`
	Status  string `json:"status"`
	Details string `json:"details,omitempty"`
}

// P2PQUICStatus represents P2P QUIC module status
type P2PQUICStatus struct {
	Enabled bool   `json:"enabled"`
	Status  string `json:"status"`
	Details string `json:"details,omitempty"`
}

// AIStatus represents AI Analytics module status
type AIStatus struct {
	Enabled bool   `json:"enabled"`
	Status  string `json:"status"`
	Details string `json:"details,omitempty"`
}

// Checker handles health checks
type Checker struct {
	startTime time.Time
	version   string
	mu        sync.RWMutex
	status    Status

	// Module health check functions
	enhancedQUICHealth func() EnhancedQUICStatus
	p2pQUICHealth      func() P2PQUICStatus
	aiHealth           func() AIStatus
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
			Modules: ModuleStatus{
				EnhancedQUIC: EnhancedQUICStatus{Enabled: false, Status: "unknown"},
				P2PQUIC:      P2PQUICStatus{Enabled: false, Status: "unknown"},
				AI:           AIStatus{Enabled: false, Status: "unknown"},
			},
		},
	}
}

// Handler handles health check requests
type Handler struct {
	metrics   *metrics.Metrics
	startTime time.Time
	version   string

	// Module health check functions
	enhancedQUICHealth func() EnhancedQUICStatus
	p2pQUICHealth      func() P2PQUICStatus
	aiHealth           func() AIStatus
}

// NewHandler creates a new health check handler
func NewHandler(metrics *metrics.Metrics, version string) *Handler {
	return &Handler{
		metrics:   metrics,
		startTime: time.Now(),
		version:   version,
	}
}

// SetEnhancedQUICHealthCheck sets the Enhanced QUIC health check function
func (h *Handler) SetEnhancedQUICHealthCheck(healthFunc func() EnhancedQUICStatus) {
	h.enhancedQUICHealth = healthFunc
}

// SetP2PQUICHealthCheck sets the P2P QUIC health check function
func (h *Handler) SetP2PQUICHealthCheck(healthFunc func() P2PQUICStatus) {
	h.p2pQUICHealth = healthFunc
}

// SetAIHealthCheck sets the AI Analytics health check function
func (h *Handler) SetAIHealthCheck(healthFunc func() AIStatus) {
	h.aiHealth = healthFunc
}

// ServeHTTP implements http.Handler
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	status := Status{
		Status:  "ok",
		Version: h.version,
		Uptime:  time.Since(h.startTime).String(),
		Metrics: make(map[string]int64),
		Modules: ModuleStatus{
			EnhancedQUIC: EnhancedQUICStatus{Enabled: false, Status: "unknown"},
			P2PQUIC:      P2PQUICStatus{Enabled: false, Status: "unknown"},
			AI:           AIStatus{Enabled: false, Status: "unknown"},
		},
		Timestamp: time.Now(),
	}

	// Add metrics
	status.Metrics["active_connections"] = h.metrics.GetActiveConnections()
	status.Metrics["active_tunnels"] = h.metrics.GetActiveTunnels()

	// Update module statuses
	if h.enhancedQUICHealth != nil {
		status.Modules.EnhancedQUIC = h.enhancedQUICHealth()
	}
	if h.p2pQUICHealth != nil {
		status.Modules.P2PQUIC = h.p2pQUICHealth()
	}
	if h.aiHealth != nil {
		status.Modules.AI = h.aiHealth()
	}

	// Update overall status based on module statuses
	overallStatus := "ok"
	if status.Modules.EnhancedQUIC.Enabled && status.Modules.EnhancedQUIC.Status != "ok" {
		overallStatus = "degraded"
	}
	if status.Modules.P2PQUIC.Enabled && status.Modules.P2PQUIC.Status != "ok" {
		overallStatus = "degraded"
	}
	if status.Modules.AI.Enabled && status.Modules.AI.Status != "ok" {
		overallStatus = "degraded"
	}
	status.Status = overallStatus

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Encode response
	if err := json.NewEncoder(w).Encode(status); err != nil {
		http.Error(w, "Failed to encode health status", http.StatusInternalServerError)
		return
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
		if err := json.NewEncoder(w).Encode(status); err != nil {
			http.Error(w, "Failed to encode health status", http.StatusInternalServerError)
			return
		}
	})
}

// SetStatus updates the health status
// SetEnhancedQUICHealthCheck sets the Enhanced QUIC health check function
func (c *Checker) SetEnhancedQUICHealthCheck(healthFunc func() EnhancedQUICStatus) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.enhancedQUICHealth = healthFunc
}

// SetP2PQUICHealthCheck sets the P2P QUIC health check function
func (c *Checker) SetP2PQUICHealthCheck(healthFunc func() P2PQUICStatus) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.p2pQUICHealth = healthFunc
}

// SetAIHealthCheck sets the AI Analytics health check function
func (c *Checker) SetAIHealthCheck(healthFunc func() AIStatus) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.aiHealth = healthFunc
}

// SetStatus sets the overall health status
func (c *Checker) SetStatus(status string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.status.Status = status
}

// UpdateModuleStatus updates the status of all modules
func (c *Checker) UpdateModuleStatus() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Update Enhanced QUIC status
	if c.enhancedQUICHealth != nil {
		c.status.Modules.EnhancedQUIC = c.enhancedQUICHealth()
	}

	// Update P2P QUIC status
	if c.p2pQUICHealth != nil {
		c.status.Modules.P2PQUIC = c.p2pQUICHealth()
	}

	// Update AI status
	if c.aiHealth != nil {
		c.status.Modules.AI = c.aiHealth()
	}

	// Update overall status based on module statuses
	c.updateOverallStatus()
}

// updateOverallStatus updates the overall health status based on module statuses
func (c *Checker) updateOverallStatus() {
	overallStatus := "ok"

	// Check if any enabled module has a non-ok status
	if c.status.Modules.EnhancedQUIC.Enabled && c.status.Modules.EnhancedQUIC.Status != "ok" {
		overallStatus = "degraded"
	}
	if c.status.Modules.P2PQUIC.Enabled && c.status.Modules.P2PQUIC.Status != "ok" {
		overallStatus = "degraded"
	}
	if c.status.Modules.AI.Enabled && c.status.Modules.AI.Status != "ok" {
		overallStatus = "degraded"
	}

	c.status.Status = overallStatus
}
