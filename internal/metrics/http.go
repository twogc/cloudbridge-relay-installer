package metrics

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Server represents the metrics HTTP server
type Server struct {
	port string
	path string
}

// NewServer creates a new metrics server
func NewServer(port, path string) *Server {
	return &Server{
		port: port,
		path: path,
	}
}

// Start starts the metrics server
func (s *Server) Start() error {
	http.Handle(s.path, promhttp.Handler())
	
	// Create server with timeouts for security
	server := &http.Server{
		Addr:         s.port,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	
	return server.ListenAndServe()
} 