package metrics

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Server represents the metrics HTTP server
type Server struct {
	port    string
	path    string
	version string
}

// NewServer creates a new metrics server
func NewServer(port, path string) *Server {
	return &Server{
		port:    port,
		path:    path,
		version: "2.0.0", // TODO: получать из build info
	}
}

// Start starts the metrics server
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Prometheus metrics endpoint
	mux.Handle(s.path, promhttp.Handler())

	// Health check endpoint
	mux.HandleFunc("/health", s.handleHealth)

	// Create server with timeouts for security
	server := &http.Server{
		Addr:         s.port,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return server.ListenAndServe()
}

// handleHealth handles health check requests
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"status":    "ok",
		"version":   s.version,
		"timestamp": time.Now().Format(time.RFC3339),
		"service":   "cloudbridge-relay",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(status); err != nil {
		http.Error(w, "Failed to encode health status", http.StatusInternalServerError)
	}
}
