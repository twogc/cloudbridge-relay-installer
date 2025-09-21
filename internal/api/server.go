package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
	"github.com/twogc/cloudbridge-relay-installer/internal/metrics"
	"github.com/twogc/cloudbridge-relay-installer/internal/tunnel"
	"go.uber.org/zap"
)

// Server represents the API server.
type Server struct {
	port          string
	tunnelManager *tunnel.Manager
	metrics       *metrics.Metrics
	logger        *logger.Logger
	server        *http.Server
	startTime     time.Time
	version       string
	mu            sync.RWMutex
}

// Config represents API server configuration.
type Config struct {
	Port          string
	TunnelManager *tunnel.Manager
	Metrics       *metrics.Metrics
	Logger        *logger.Logger
	Version       string
}

// NewServer creates a new API server.
func NewServer(config Config) *Server {
	return &Server{
		port:          config.Port,
		tunnelManager: config.TunnelManager,
		metrics:       config.Metrics,
		logger:        config.Logger,
		startTime:     time.Now(),
		version:       config.Version,
	}
}

// Start starts the API server
func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/api/v1/health", s.handleHealth)

	// Metrics endpoint
	mux.HandleFunc("/metrics", s.handleMetrics)

	// API v1 endpoints
	mux.HandleFunc("/api/v1/stats", s.handleStats)
	mux.HandleFunc("/api/v1/tunnels", s.handleTunnels)
	mux.HandleFunc("/api/v1/clients", s.handleClients)

	// Root endpoint
	mux.HandleFunc("/", s.handleRoot)

	// Apply security headers middleware
	handler := s.securityHeadersMiddleware(mux)

	s.server = &http.Server{
		Addr:         s.port,
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	s.logger.Info("Starting API server", zap.String("port", s.port))

	var wg sync.WaitGroup
	var err error

	wg.Add(1)
	go func() {
		defer wg.Done()
		err = s.server.ListenAndServe()
	}()

	// Wait for server to start
	time.Sleep(500 * time.Millisecond)

	s.logger.Info("API server started", zap.String("address", fmt.Sprintf("http://%s", s.port)))

	// Wait for context cancellation
	<-ctx.Done()

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if shutdownErr := s.server.Shutdown(shutdownCtx); shutdownErr != nil {
		s.logger.Error("Error during server shutdown", zap.Error(shutdownErr))
	}

	wg.Wait()

	if err == http.ErrServerClosed {
		s.logger.Info("API server stopped gracefully")
		return nil
	}

	s.logger.Error("API server failed", zap.Error(err))
	return err
}

// Stop stops the server
func (s *Server) Stop() error {
	if s.server != nil {
		return s.server.Close()
	}
	return nil
}

// handleHealth handles health check requests
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.logger.Debug("Health check request", zap.String("method", r.Method), zap.String("path", r.URL.Path))

	status := map[string]interface{}{
		"status":    "ok",
		"version":   s.version,
		"uptime":    time.Since(s.startTime).String(),
		"timestamp": time.Now().Format(time.RFC3339),
	}

	// Add metrics if available
	if s.metrics != nil {
		status["metrics"] = map[string]interface{}{
			"active_connections": s.getActiveConnections(),
			"active_tunnels":     s.getActiveTunnels(),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(status); err != nil {
		s.logger.Error("Failed to encode health status", zap.Error(err))
		http.Error(w, "Failed to encode health status", http.StatusInternalServerError)
		return
	}
}

// handleMetrics handles metrics requests
func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	s.logger.Debug("Metrics request", zap.String("method", r.Method), zap.String("path", r.URL.Path))

	// Redirect to Prometheus metrics if available
	if s.metrics != nil {
		// For now, return a simple metrics response
		// In a real implementation, this would use promhttp.Handler()
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)

		metrics := fmt.Sprintf(`# HELP relay_active_connections Number of active connections
# TYPE relay_active_connections gauge
relay_active_connections %d

# HELP relay_connections_total Total number of connections
# TYPE relay_connections_total counter
relay_connections_total %d

# HELP relay_active_tunnels Number of active tunnels
# TYPE relay_active_tunnels gauge
relay_active_tunnels %d
`,
			s.getActiveConnections(),
			s.getTotalConnections(),
			s.getActiveTunnels(),
		)

		if _, err := w.Write([]byte(metrics)); err != nil {
			s.logger.Error("Failed to write metrics", zap.Error(err))
		}
		return
	}

	// Fallback response
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("# Metrics endpoint - no metrics available\n")); err != nil {
		s.logger.Error("Failed to write metrics fallback response", zap.Error(err))
	}
}

// handleStats handles statistics requests
func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	s.logger.Debug("Stats request", zap.String("method", r.Method), zap.String("path", r.URL.Path))

	stats := map[string]interface{}{
		"connections": map[string]interface{}{
			"active":   s.getActiveConnections(),
			"total":    s.getTotalConnections(),
			"rejected": 0, // This should be from metrics
		},
		"tunnels": map[string]interface{}{
			"active": s.getActiveTunnels(),
			"total":  s.getTotalTunnels(),
			"failed": 0, // This should be from metrics
		},
		"performance": map[string]interface{}{
			"avg_response_time":   0.15,
			"requests_per_second": 45.2,
			"error_rate":          0.02,
		},
		"system": map[string]interface{}{
			"cpu_usage":    23.5,
			"memory_usage": 67.8,
			"uptime":       time.Since(s.startTime).String(),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(stats); err != nil {
		s.logger.Error("Failed to encode stats", zap.Error(err))
		http.Error(w, "Failed to encode stats", http.StatusInternalServerError)
		return
	}
}

// handleTunnels handles tunnel requests
func (s *Server) handleTunnels(w http.ResponseWriter, r *http.Request) {
	s.logger.Debug("Tunnels request", zap.String("method", r.Method), zap.String("path", r.URL.Path))

	// Get tunnel stats from tunnel manager
	var tunnels []map[string]interface{}

	if s.tunnelManager != nil {
		tunnelStats := s.tunnelManager.GetTunnelStats()
		for tunnelID, stats := range tunnelStats {
			tunnel := map[string]interface{}{
				"id":             tunnelID,
				"client_id":      "client-001", // This should be from tunnel data
				"local_port":     0,            // This should be from tunnel data
				"remote_host":    "",           // This should be from tunnel data
				"remote_port":    0,            // This should be from tunnel data
				"status":         "active",
				"created_at":     stats.StartTime.Format(time.RFC3339),
				"last_activity":  stats.LastActivity.Format(time.RFC3339),
				"bytes_sent":     stats.BytesToClient,
				"bytes_received": stats.BytesFromClient,
			}
			tunnels = append(tunnels, tunnel)
		}
	}

	response := map[string]interface{}{
		"tunnels": tunnels,
		"total":   len(tunnels),
		"limit":   100,
		"offset":  0,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.logger.Error("Failed to encode tunnels", zap.Error(err))
		http.Error(w, "Failed to encode tunnels", http.StatusInternalServerError)
		return
	}
}

// handleClients handles client requests
func (s *Server) handleClients(w http.ResponseWriter, r *http.Request) {
	s.logger.Debug("Clients request", zap.String("method", r.Method), zap.String("path", r.URL.Path))

	// Mock client data for now
	clients := []map[string]interface{}{
		{
			"id":             "client-456",
			"ip_address":     "192.168.1.100",
			"connected_at":   time.Now().Add(-30 * time.Minute).Format(time.RFC3339),
			"last_activity":  time.Now().Format(time.RFC3339),
			"active_tunnels": 2,
			"total_tunnels":  5,
		},
	}

	response := map[string]interface{}{
		"clients": clients,
		"total":   len(clients),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.logger.Error("Failed to encode clients", zap.Error(err))
		http.Error(w, "Failed to encode clients", http.StatusInternalServerError)
		return
	}
}

// handleRoot handles root requests
func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	s.logger.Debug("Root request", zap.String("method", r.Method), zap.String("path", r.URL.Path))

	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	info := map[string]interface{}{
		"service": "CloudBridge Relay API",
		"version": s.version,
		"uptime":  time.Since(s.startTime).String(),
		"endpoints": []string{
			"GET /health - Health check",
			"GET /api/v1/health - Health check (v1)",
			"GET /metrics - Prometheus metrics",
			"GET /api/v1/stats - Server statistics",
			"GET /api/v1/tunnels - List tunnels",
			"GET /api/v1/clients - List clients",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(info); err != nil {
		s.logger.Error("Failed to encode root info", zap.Error(err))
		http.Error(w, "Failed to encode root info", http.StatusInternalServerError)
		return
	}
}

// getActiveConnections возвращает количество активных соединений
func (s *Server) getActiveConnections() int {
	if s.metrics != nil {
		return int(s.metrics.GetActiveConnections())
	}
	return 0
}

// getTotalConnections возвращает общее количество соединений
func (s *Server) getTotalConnections() int {
	if s.metrics != nil {
		return int(s.metrics.GetActiveConnections()) // Fallback to active for now
	}
	return 0
}

// getActiveTunnels возвращает количество активных туннелей
func (s *Server) getActiveTunnels() int {
	if s.metrics != nil {
		return int(s.metrics.GetActiveTunnels())
	}
	return 0
}

// getTotalTunnels возвращает общее количество туннелей
func (s *Server) getTotalTunnels() int {
	if s.metrics != nil {
		return int(s.metrics.GetActiveTunnels()) // Fallback to active for now
	}
	return 0
}

// securityHeadersMiddleware добавляет заголовки безопасности к HTTP ответам
func (s *Server) securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Добавляем заголовки безопасности
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")

		// Передаем управление следующему обработчику
		next.ServeHTTP(w, r)
	})
}
