package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// ApplianceServer represents the appliance HTTP server
type ApplianceServer struct {
	server  *http.Server
	manager *ApplianceManager
	router  *mux.Router
	config  *ServerConfig
}

// ServerConfig represents server configuration
type ServerConfig struct {
	Host         string        `json:"host"`          // Server host
	Port         int           `json:"port"`          // Server port
	ReadTimeout  time.Duration `json:"read_timeout"`  // Read timeout
	WriteTimeout time.Duration `json:"write_timeout"` // Write timeout
	IdleTimeout  time.Duration `json:"idle_timeout"`  // Idle timeout
	CertFile     string        `json:"cert_file"`     // TLS certificate file
	KeyFile      string        `json:"key_file"`      // TLS key file
}

// NewApplianceServer creates a new appliance server
func NewApplianceServer(config *ServerConfig, applianceConfig *ApplianceConfig) *ApplianceServer {
	router := mux.NewRouter()

	// Create appliance manager
	manager := NewApplianceManager(applianceConfig)

	// Setup routes
	manager.SetupRoutes(router)

	// Add Prometheus metrics endpoint
	router.Path("/metrics").Handler(promhttp.Handler())

	// Add health check endpoint
	router.Path("/health").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})

	// Add CORS middleware
	router.Use(corsMiddleware)

	// Add logging middleware
	router.Use(loggingMiddleware)

	// Add authentication middleware
	router.Use(authMiddleware)

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
		Handler:      router,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		IdleTimeout:  config.IdleTimeout,
	}

	return &ApplianceServer{
		server:  server,
		manager: manager,
		router:  router,
		config:  config,
	}
}

// Start starts the appliance server
func (as *ApplianceServer) Start() error {
	log.Printf("🚀 Starting CloudBridge Appliance Server on %s", as.server.Addr)

	// Start server in goroutine
	go func() {
		var err error
		if as.config.CertFile != "" && as.config.KeyFile != "" {
			err = as.server.ListenAndServeTLS(as.config.CertFile, as.config.KeyFile)
		} else {
			err = as.server.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return as.server.Shutdown(ctx)
}

// Stop stops the appliance server
func (as *ApplianceServer) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return as.server.Shutdown(ctx)
}

// GetManager returns the appliance manager
func (as *ApplianceServer) GetManager() *ApplianceManager {
	return as.manager
}

// Middleware functions
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		log.Printf("📊 %s %s %d %v", r.Method, r.URL.Path, wrapped.statusCode, duration)
	})
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for health and metrics endpoints
		if r.URL.Path == "/health" || r.URL.Path == "/metrics" {
			next.ServeHTTP(w, r)
			return
		}

		// Check for API key in header
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			http.Error(w, "Missing API key", http.StatusUnauthorized)
			return
		}

		// Validate API key (implement your validation logic)
		if !isValidAPIKey(apiKey) {
			http.Error(w, "Invalid API key", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// isValidAPIKey validates the API key
func isValidAPIKey(apiKey string) bool {
	// Implement your API key validation logic
	// For now, accept any non-empty key
	return apiKey != ""
}

// Default server configuration
func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		Host:         "0.0.0.0",
		Port:         8080,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
}

// Default appliance configuration
func DefaultApplianceConfig() *ApplianceConfig {
	return &ApplianceConfig{
		Mode:            "standalone",
		ThroughputLimit: "2Gbps",
		Features:        []string{"relay", "tunnels", "basic_security"},
		License: &LicenseInfo{
			Type:       "basic",
			ExpiresAt:  time.Now().Add(365 * 24 * time.Hour), // 1 year
			MaxUsers:   20,
			MaxServers: 5,
			MaxGroups:  5,
			Features:   []string{"relay", "tunnels"},
			Throughput: "2Gbps",
		},
		Hardware: &HardwareInfo{
			CPU:        "4 cores",
			Memory:     "16GB",
			Storage:    "500GB SSD",
			Network:    "2x10G",
			FormFactor: "1U",
		},
		Network: &NetworkConfig{
			Interfaces: []NetworkInterface{
				{
					Name:        "eth0",
					Type:        "ethernet",
					Speed:       "10G",
					IPAddresses: []string{"192.168.1.100/24"},
					Status:      "up",
				},
			},
		},
		Security: &SecurityConfig{
			WAF: &WAFConfig{
				Enabled:   false,
				Mode:      "monitor",
				LogLevel:  "info",
				RateLimit: 1000,
			},
			DDoS: &DDoSConfig{
				Enabled:   false,
				Threshold: 10000,
				Action:    "log",
				Whitelist: []string{},
				Blacklist: []string{},
			},
			ZeroTrust: &ZeroTrustConfig{
				Enabled:      false,
				DeviceTrust:  false,
				UserTrust:    false,
				NetworkTrust: false,
			},
			Audit: &AuditConfig{
				Enabled:      true,
				LogLevel:     "info",
				Retention:    30,
				Destinations: []string{"local"},
			},
		},
	}
}
