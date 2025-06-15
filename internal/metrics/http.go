package metrics

import (
	"net/http"

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
	return http.ListenAndServe(s.port, nil)
} 