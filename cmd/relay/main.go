package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/twogc/cloudbridge-relay-installer/config"
	"github.com/twogc/cloudbridge-relay-installer/internal/auth"
	"github.com/twogc/cloudbridge-relay-installer/internal/balancer"
	"github.com/twogc/cloudbridge-relay-installer/internal/cadence"
	"github.com/twogc/cloudbridge-relay-installer/internal/circuitbreaker"
	"github.com/twogc/cloudbridge-relay-installer/internal/health"
	"github.com/twogc/cloudbridge-relay-installer/internal/keycloak"
	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
	"github.com/twogc/cloudbridge-relay-installer/internal/metrics"
	"github.com/twogc/cloudbridge-relay-installer/internal/p2c"
	"github.com/twogc/cloudbridge-relay-installer/internal/p2p"
	"github.com/twogc/cloudbridge-relay-installer/internal/protocol"
	"github.com/twogc/cloudbridge-relay-installer/internal/quantum"
	"github.com/twogc/cloudbridge-relay-installer/internal/ratelimit"
	"github.com/twogc/cloudbridge-relay-installer/internal/tenant"
	"github.com/twogc/cloudbridge-relay-installer/internal/tunnel"
	"go.uber.org/zap"
)

var (
	configFile = flag.String("config", "configs/config.json", "Path to configuration file")
	logLevel   = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	logFile    = flag.String("log-file", "", "Log file path")

	// Prometheus metrics
	connectionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "relay_connections_total",
			Help: "Total number of connections",
		},
		[]string{"status"},
	)

	authAttemptsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "relay_auth_attempts_total",
			Help: "Total number of authentication attempts",
		},
		[]string{"status", "provider"},
	)

	tunnelsCreatedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "relay_tunnels_created_total",
			Help: "Total number of tunnels created",
		},
		[]string{"status"},
	)

	activeConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "relay_active_connections",
			Help: "Number of active connections",
		},
	)

	// Local numeric counters for JSON status endpoint
	activeConnCount int64
	totalConnCount  int64
)

// Server represents the main relay server
type Server struct {
	config        *config.Config
	logger        *logger.Logger
	listener      net.Listener
	httpServer    *http.Server
	tunnelManager *tunnel.Manager
	metricsServer *metrics.Server
	authAdapter   *auth.Adapter

	// Zero Trust components
	circuitBreaker *circuitbreaker.CircuitBreaker
	healthChecker  *health.HealthChecker

	// Enhanced QUIC components
	enhancedQUICManager  *EnhancedQUICManager
	tenantManager        *tenant.Manager
	loadBalancer         *balancer.Balancer
	ipRateLimit          *ratelimit.IPLimiter
	enhancedKeycloakAuth *auth.EnhancedKeycloakAuth
	enhancedEnabled      bool
	wireGuardEnabled     bool
	wireGuardInterface   *wireguard.WireGuardInterface
	wireGuardMesh        *wireguard.MeshTopology

	// P2P Relay components
	relayManager     *p2p.RelayManager
	heartbeatManager *p2p.HeartbeatManager
	quicRelayHandler *QUICRelayHandler

	// Cadence Workflow Engine
	cadenceClient          *cadence.Client
	wireGuardPeerDiscovery *wireguard.PeerDiscovery

	// P2P API Server
	p2pAPIServer      *P2PAPIServer
	quicOverWireGuard *wireguard.QUICOverWireGuard
	hybridRouter      *wireguard.HybridRouter
	siteToSiteManager *wireguard.SiteToSiteManager

	// Post-Quantum Cryptography components
	quantumEnabled      bool
	quantumConfig       *quantum.PQCConfig
	quantumWireGuard    *quantum.QuantumWireGuardInterface
	quantumEnhancedQUIC *quantum.QuantumEnhancedQUIC

	// AI Analytics components
	aiComponents *AIComponents

	// Мультипротокольные компоненты
	protocolSelector  *protocol.IntelligentProtocolSelector
	protocolEngine    *protocol.ProtocolEngine
	p2pManager        *p2p.P2PManager
	p2cManager        *p2c.P2CManager
	connectionHandler *protocol.UnifiedConnectionHandler
	autoSwitchManager *protocol.AutoSwitchManager

	ctx       context.Context
	cancel    context.CancelFunc
	startTime time.Time
}

// NewServer creates a new relay server
func NewServer(cfg *config.Config) (*Server, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Ensure context is cancelled on any return path

	// Initialize logger
	loggerConfig := logger.Config{
		Level:      logger.LogLevel(*logLevel),
		Output:     "stdout",
		JSON:       true,
		MaxSize:    100,
		MaxAge:     30,
		MaxBackups: 10,
		Compress:   true,
	}

	if *logFile != "" {
		loggerConfig.Output = "both"
		loggerConfig.FilePath = *logFile
	}

	loggerInstance, err := logger.New(loggerConfig)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize logger: %v", err)
	}

	// Initialize tunnel manager
	tunnelManager := tunnel.NewManager(cfg.Server.MaxTunnels, loggerInstance, nil)

	// Initialize auth adapter
	authAdapter, err := auth.NewAdapter(cfg, loggerInstance)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize auth adapter: %v", err)
	}

	// Initialize Zero Trust components
	var circuitBreaker *circuitbreaker.CircuitBreaker
	if cfg.ZeroTrust.CircuitBreaker.Enabled {
		// Convert logger.Logger to *zap.Logger for circuit breaker
		zapLogger := zap.NewNop() // Fallback to no-op logger

		circuitBreaker = circuitbreaker.NewCircuitBreaker(&circuitbreaker.Config{
			Name:             "relay-server",
			FailureThreshold: cfg.ZeroTrust.CircuitBreaker.FailureThreshold,
			SuccessThreshold: cfg.ZeroTrust.CircuitBreaker.SuccessThreshold,
			Timeout:          cfg.ZeroTrust.CircuitBreaker.Timeout,
			WindowSize:       cfg.ZeroTrust.CircuitBreaker.WindowSize,
			Logger:           zapLogger,
		})
	}

	var healthChecker *health.HealthChecker
	if cfg.ZeroTrust.HealthCheck.Enabled {
		// Convert logger.Logger to *zap.Logger for health checker
		zapLogger := zap.NewNop() // Fallback to no-op logger

		healthChecker = health.NewHealthChecker(&health.Config{
			Interval: cfg.ZeroTrust.HealthCheck.Interval,
			Timeout:  cfg.ZeroTrust.HealthCheck.Timeout,
			Logger:   zapLogger,
		})

		// Add configured health checks
		for _, checkConfig := range cfg.ZeroTrust.HealthCheck.Checks {
			switch checkConfig.Type {
			case "http":
				if checkConfig.URL != "" {
					healthChecker.AddCheck(checkConfig.Name, health.HTTPHealthCheck(checkConfig.Name, checkConfig.URL))
				}
			case "ping":
				if checkConfig.Host != "" {
					healthChecker.AddCheck(checkConfig.Name, health.PingHealthCheck(checkConfig.Name, checkConfig.Host))
				}
			case "custom":
				// Custom health checks can be added here
				healthChecker.AddCheck(checkConfig.Name, health.CustomHealthCheck(
					checkConfig.Name,
					checkConfig.Description,
					func(ctx context.Context) error {
						// Default custom check - always healthy
						return nil
					},
				))
			}
		}

		// Start health checker
		healthChecker.Start()
	}

	// Initialize metrics server
	var metricsServer *metrics.Server
	if cfg.Metrics.Enabled {
		metricsServer = metrics.NewServer(fmt.Sprintf(":%d", cfg.Metrics.Port), cfg.Metrics.Path)
	}

	// Initialize Enhanced QUIC components
	enhancedEnabled := cfg.EnhancedQUIC.Enabled // Из конфигурации
	var enhancedQUICManager *EnhancedQUICManager
	var tenantMgr *tenant.Manager
	var loadBal *balancer.Balancer
	var ipRateLimit *ratelimit.IPLimiter
	var enhancedKeycloakAuth *auth.EnhancedKeycloakAuth

	if enhancedEnabled {
		// Конвертируем logger.Logger в *zap.Logger для Enhanced компонентов
		zapLogger := zap.NewNop() // Fallback

		// Инициализируем Enhanced QUIC Manager
		enhancedQUICManager = NewEnhancedQUICManager(zapLogger)

		// Инициализируем дополнительные компоненты для Enhanced режима
		tenantMgr = tenant.NewManagerWithEnhanced(loggerInstance, nil)
		loadBal = balancer.NewBalancerWithEnhanced()
		ipRateLimit = ratelimit.NewIPLimiterWithEnhanced(100.0, 1000) // 100 req/sec, burst 1000

		// Инициализируем Enhanced Keycloak Auth
		keycloakConfig := &auth.KeycloakConfig{
			ServerURL:    "https://edge.2gc.ru",
			Realm:        "cloudbridge",
			ClientID:     "quic-client",
			ClientSecret: "secret",
		}
		enhancedKeycloakAuth = auth.NewEnhancedKeycloakAuth(keycloakConfig, zapLogger)

		// Настраиваем интеграции
		if err := enhancedQUICManager.SetupIntegrations(tenantMgr, loadBal, ipRateLimit, enhancedKeycloakAuth); err != nil {
			loggerInstance.Error("Failed to setup Enhanced QUIC integrations", zap.Error(err))
			enhancedEnabled = false
		} else {
			if err := enhancedQUICManager.Initialize(); err != nil {
				loggerInstance.Error("Failed to initialize Enhanced QUIC", zap.Error(err))
				enhancedEnabled = false
			} else {
				loggerInstance.Info("Enhanced QUIC initialized successfully")
			}
		}
	} else {
		// Инициализируем обычные компоненты
		tenantMgr = tenant.NewManager(loggerInstance, nil)
		loadBal = balancer.NewBalancer()
		ipRateLimit = ratelimit.NewIPLimiter(100.0, 1000)
		enhancedKeycloakAuth = auth.NewKeycloakAuthCompatible(&auth.KeycloakConfig{
			ServerURL: "https://edge.2gc.ru",
			Realm:     "cloudbridge",
		}, zap.NewNop())
	}
	// P2P is now enabled by default for QUIC relay functionality
	p2pEnabled := true // Always enable P2P for QUIC relay
	var wireGuardInterface *wireguard.WireGuardInterface
	var wireGuardMesh *wireguard.MeshTopology
	var wireGuardPeerDiscovery *wireguard.PeerDiscovery
	var quicOverWireGuard *wireguard.QUICOverWireGuard
	var hybridRouter *wireguard.HybridRouter
	var siteToSiteManager *wireguard.SiteToSiteManager
	// P2P functionality is handled by QUIC relay instead
	loggerInstance.Info("P2P QUIC relay mode enabled, WireGuard disabled")

	// Initialize P2P API Server
	loggerInstance.Info("Starting P2P API Server initialization...")
	var p2pAPIServer *P2PAPIServer
	// TODO: Добавить P2P конфигурацию в config.go
	// if cfg.P2P.API.Enabled {
	// Get JWT secret from environment variable
	jwtSecret := []byte(os.Getenv("JWT_SECRET"))
	loggerInstance.Info("JWT secret prepared", zap.Int("secret_length", len(jwtSecret)))
	if len(jwtSecret) == 0 {
		jwtSecret = []byte("default-jwt-secret")
		loggerInstance.Warn("JWT_SECRET environment variable not set, using default")
	}

	loggerInstance.Info("Calling NewP2PAPIServer...")
	// Передаем allowed_ips из конфигурации
	allowedIPs := cfg.Security.AllowedIPs
	loggerInstance.Info("Using allowed IPs from configuration", zap.Int("count", len(allowedIPs)))

	// Создаем Keycloak валидатор если включен
	var keycloakValidator *keycloak.KeycloakValidator
	if cfg.Keycloak.Enabled {
		loggerInstance.Info("Initializing Keycloak validator",
			zap.String("jwks_url", cfg.Keycloak.JWKSURL),
			zap.String("issuer", cfg.Keycloak.TokenValidation.Issuer),
			zap.String("audience", cfg.Keycloak.TokenValidation.Audience),
		)

		keycloakValidator = keycloak.NewKeycloakValidator(
			cfg.Keycloak.JWKSURL,
			cfg.Keycloak.TokenValidation.Audience,
			cfg.Keycloak.TokenValidation.Issuer,
			cfg.Keycloak.FallbackSecret,
			loggerInstance.GetZapLogger(),
		)
	}

	// Ensure API port is set correctly
	apiPort := cfg.API.Port
	if apiPort == 0 {
		apiPort = 8082 // fallback to default port
		loggerInstance.Warn("API port was 0, using fallback port 8082")
	}

	loggerInstance.Info("API Config", zap.Int("port", apiPort), zap.Bool("enabled", cfg.API.Enabled))
	loggerInstance.Info("DEBUG: About to create P2P API Server")
	p2pAPIServer = NewP2PAPIServerWithKeycloak(loggerInstance.GetZapLogger(), jwtSecret, apiPort, allowedIPs, keycloakValidator)
	loggerInstance.Info("NewP2PAPIServer call completed", zap.Bool("server_created", p2pAPIServer != nil))

	if p2pAPIServer != nil {
		loggerInstance.Info("P2P API Server initialized successfully",
			zap.Int("port", apiPort),
			zap.String("host", "0.0.0.0"),
		)

		// P2P API Server will be started later in Server.Start()
		loggerInstance.Info("P2P API Server ready to start")
	} else {
		loggerInstance.Error("Failed to initialize P2P API Server - NewP2PAPIServer returned nil")
	}
	// }

	// Initialize Post-Quantum Cryptography components
	quantumEnabled := cfg.Quantum.Enabled
	var quantumConfig *quantum.PQCConfig
	var quantumWireGuard *quantum.QuantumWireGuardInterface
	var quantumEnhancedQUIC *quantum.QuantumEnhancedQUIC

	if quantumEnabled {
		// Create PQC configuration
		quantumConfig = quantum.NewPQCConfig()

		// Initialize quantum-resistant Enhanced QUIC
		if enhancedQUICManager != nil {
			quantumEnhancedQUIC = quantum.NewQuantumEnhancedQUIC(enhancedQUICManager, quantumConfig)
		}

		loggerInstance.Info("Post-Quantum Cryptography initialized",
			zap.String("algorithm", string(quantumConfig.Algorithm)),
			zap.Bool("hybrid_mode", quantumConfig.HybridMode),
		)
	}

	// Initialize Cadence Workflow Engine
	var cadenceClient *cadence.Client
	if cfg.Cadence.Enabled {
		cadenceConfig := cadence.Config{
			HostPort: cfg.Cadence.HostPort,
			Domain:   cfg.Cadence.Domain,
			TaskList: cfg.Cadence.TaskList,
		}

		var err error
		cadenceClient, err = cadence.NewClient(cadenceConfig, loggerInstance.GetZapLogger())
		if err != nil {
			loggerInstance.Error("Failed to initialize Cadence client", zap.Error(err))
			return nil, fmt.Errorf("failed to initialize Cadence client: %v", err)
		}

		loggerInstance.Info("Cadence Workflow Engine initialized successfully",
			zap.String("domain", cfg.Cadence.Domain),
			zap.String("host_port", cfg.Cadence.HostPort),
		)
	}

	// Initialize AI Analytics components
	aiComponents, err := InitializeAIComponents(cfg, loggerInstance)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize AI components: %v", err)
	}

	// Initialize Multiprotocol components
	// Convert logger.Logger to *zap.Logger for protocol components
	zapLogger := zap.NewNop() // Fallback to no-op logger

	protocolSelector := protocol.NewIntelligentProtocolSelector(protocol.ProtocolSelectorConfig{
		Strategy:            protocol.PerformanceBased,
		PreferredOrder:      []protocol.Protocol{protocol.QUIC, protocol.HTTP2, protocol.WebSocket},
		SwitchThreshold:     cfg.Protocol.SwitchThreshold,
		HealthCheckInterval: cfg.Protocol.HealthCheckInterval,
		Logger:              zapLogger,
	})

	protocolEngine := protocol.NewProtocolEngine([]protocol.Protocol{protocol.QUIC, protocol.HTTP2, protocol.WebSocket}, cfg.Protocol.SwitchThreshold, zapLogger)
	p2pManager := p2p.NewP2PManager(
		nil, // wireGuardMesh - not needed for QUIC relay
		nil, // wireGuardPeerDiscovery - not needed for QUIC relay
		nil, // hybridRouter - not needed for QUIC relay
		nil, // siteToSiteManager - not needed for QUIC relay
		zapLogger,
	)

	// Initialize P2C Manager
	p2cEnhancedQUICManager := p2c.NewEnhancedQUICManager(zapLogger)
	p2cManager := p2c.NewP2CManager(
		p2cEnhancedQUICManager,
		tenantMgr,
		loadBal,
		ipRateLimit,
		zapLogger,
	)

	// Initialize Unified Connection Handler
	connectionHandler := protocol.NewUnifiedConnectionHandler(
		protocolSelector,
		p2pManager,
		p2cManager,
		zapLogger,
	)

	// Initialize Auto Switch Manager
	autoSwitchManager := protocol.NewAutoSwitchManager(
		protocolSelector,
		p2pManager,
		p2cManager,
		zapLogger,
	)

	loggerInstance.Info("Multiprotocol components initialized",
		zap.Bool("p2p_enabled", p2pEnabled),
		zap.Bool("p2c_enabled", enhancedEnabled),
		zap.Bool("auto_switch_enabled", cfg.Protocol.AutoSelect),
	)

	// Initialize P2P Relay components
	loggerInstance.Info("Initializing P2P Relay components...")

	// Create Relay Manager
	relayManager := p2p.NewRelayManager(loggerInstance.GetZapLogger())

	// Create Heartbeat Manager
	heartbeatManager := p2p.NewHeartbeatManager(relayManager, loggerInstance.GetZapLogger())

	// Create QUIC Relay Handler (independent of p2pAPIServer)
	var quicRelayHandler *QUICRelayHandler
	if p2pEnabled {
		// Create QUIC Relay Handler with basic components
		// We'll create a minimal peer manager and token validator for QUIC relay
		quicRelayHandler = NewQUICRelayHandler(
			relayManager,
			nil, // peerManager - will be set up later if needed
			nil, // tokenValidator - will be set up later if needed
			loggerInstance.GetZapLogger(),
		)
		loggerInstance.Info("QUIC Relay Handler created successfully")
	}

	// Set managers for P2P API Server
	if p2pAPIServer != nil {
		p2pAPIServer.SetManagers(relayManager, heartbeatManager)
	}

	loggerInstance.Info("P2P Relay components initialized successfully")

	return &Server{
		config:                 cfg,
		logger:                 loggerInstance,
		tunnelManager:          tunnelManager,
		metricsServer:          metricsServer,
		authAdapter:            authAdapter,
		circuitBreaker:         circuitBreaker,
		healthChecker:          healthChecker,
		enhancedQUICManager:    enhancedQUICManager,
		tenantManager:          tenantMgr,
		loadBalancer:           loadBal,
		ipRateLimit:            ipRateLimit,
		enhancedKeycloakAuth:   enhancedKeycloakAuth,
		enhancedEnabled:        enhancedEnabled,
		wireGuardEnabled:       false, // WireGuard disabled, using QUIC relay
		wireGuardInterface:     wireGuardInterface,
		wireGuardMesh:          wireGuardMesh,
		wireGuardPeerDiscovery: wireGuardPeerDiscovery,
		p2pAPIServer:           p2pAPIServer,
		quicOverWireGuard:      quicOverWireGuard,
		hybridRouter:           hybridRouter,
		siteToSiteManager:      siteToSiteManager,
		quantumEnabled:         quantumEnabled,
		quantumConfig:          quantumConfig,
		quantumWireGuard:       quantumWireGuard,
		quantumEnhancedQUIC:    quantumEnhancedQUIC,
		aiComponents:           aiComponents,
		protocolSelector:       protocolSelector,
		protocolEngine:         protocolEngine,
		p2pManager:             p2pManager,
		p2cManager:             p2cManager,
		connectionHandler:      connectionHandler,
		autoSwitchManager:      autoSwitchManager,
		cadenceClient:          cadenceClient,
		relayManager:           relayManager,
		heartbeatManager:       heartbeatManager,
		quicRelayHandler:       quicRelayHandler,
		ctx:                    ctx,
		cancel:                 cancel,
		startTime:              time.Now(),
	}, nil
}

// Start starts the server
func (s *Server) Start() error {
	// Start metrics server
	if s.metricsServer != nil {
		go func() {
			if err := s.metricsServer.Start(); err != nil {
				s.logger.Error("Failed to start metrics server", zap.Error(err))
			}
		}()
	}

	// Start HTTP server on port 80 for health checks
	go func() {
		httpMux := http.NewServeMux()

		// Health check endpoint
		httpMux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			response := map[string]interface{}{
				"status":    "healthy",
				"timestamp": time.Now().UTC(),
				"uptime":    time.Since(s.startTime).String(),
				"version":   "1.0.0",
			}
			if err := json.NewEncoder(w).Encode(response); err != nil {
				s.logger.Error("Failed to encode health response", zap.Error(err))
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
		})
		// Relay status (JSON) endpoint for dashboards and Zabbix
		httpMux.HandleFunc("/relay/status", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			resp := map[string]interface{}{
				"connections": map[string]interface{}{
					"active": atomic.LoadInt64(&activeConnCount),
					"total":  atomic.LoadInt64(&totalConnCount),
				},
				"runtime": map[string]interface{}{
					"goroutines": runtime.NumGoroutine(),
				},
				"uptime": time.Since(s.startTime).Seconds(),
			}
			if err := json.NewEncoder(w).Encode(resp); err != nil {
				s.logger.Error("Failed to encode relay status response", zap.Error(err))
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
		})

		// Backward-compatible API path
		httpMux.HandleFunc("/api/v1/relay/status", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			resp := map[string]interface{}{
				"connections": map[string]interface{}{
					"active": atomic.LoadInt64(&activeConnCount),
					"total":  atomic.LoadInt64(&totalConnCount),
				},
				"runtime": map[string]interface{}{
					"goroutines": runtime.NumGoroutine(),
				},
				"uptime": time.Since(s.startTime).Seconds(),
			}
			if err := json.NewEncoder(w).Encode(resp); err != nil {
				s.logger.Error("Failed to encode relay status response", zap.Error(err))
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
		})

		// Ready check endpoint
		httpMux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			response := map[string]interface{}{
				"status": "ready",
				"services": map[string]interface{}{
					"http":      true,
					"wireguard": s.wireGuardEnabled,
					"quic":      s.enhancedQUICManager != nil,
					"metrics":   s.metricsServer != nil,
				},
			}
			if err := json.NewEncoder(w).Encode(response); err != nil {
				s.logger.Error("Failed to encode ready response", zap.Error(err))
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
		})

		// Metrics endpoint
		httpMux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
			if s.metricsServer != nil {
				// Use Prometheus handler directly
				promhttp.Handler().ServeHTTP(w, r)
			} else {
				http.Error(w, "Metrics not available", http.StatusServiceUnavailable)
			}
		})

		// Cadence API endpoints
		httpMux.HandleFunc("/api/v1/cadence/health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if s.cadenceClient != nil {
				w.WriteHeader(http.StatusOK)
				response := map[string]interface{}{
					"status": "healthy",
					"domain": s.cadenceClient.GetDomain(),
				}
				if err := json.NewEncoder(w).Encode(response); err != nil {
					s.logger.Error("Failed to encode cadence health response", zap.Error(err))
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}
			} else {
				w.WriteHeader(http.StatusServiceUnavailable)
				response := map[string]interface{}{
					"status": "unavailable",
					"error":  "Cadence client not initialized",
				}
				if err := json.NewEncoder(w).Encode(response); err != nil {
					s.logger.Error("Failed to encode cadence unavailable response", zap.Error(err))
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}
			}
		})

		httpMux.HandleFunc("/api/v1/cadence/workflows/relay-deployment", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if s.cadenceClient == nil {
				http.Error(w, "Cadence client not available", http.StatusServiceUnavailable)
				return
			}

			var req struct {
				Config          map[string]interface{} `json:"config"`
				WireGuardConfig map[string]interface{} `json:"wireguard_config"`
				QUICConfig      map[string]interface{} `json:"quic_config"`
				Services        []string               `json:"services"`
				Timeout         string                 `json:"timeout"`
			}

			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid JSON", http.StatusBadRequest)
				return
			}

			workflowID, err := s.cadenceClient.ExecuteWorkflow(context.Background(), "RelayDeploymentWorkflow", req)
			if err != nil {
				s.logger.Error("Failed to execute workflow", zap.Error(err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			response := map[string]interface{}{
				"workflow_id": workflowID,
				"status":      "started",
			}
			if err := json.NewEncoder(w).Encode(response); err != nil {
				s.logger.Error("Failed to encode JSON response", zap.Error(err))
			}
		})

		httpMux.HandleFunc("/api/v1/cadence/workflows/backup", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if s.cadenceClient == nil {
				http.Error(w, "Cadence client not available", http.StatusServiceUnavailable)
				return
			}

			var req struct {
				BackupConfig  map[string]interface{} `json:"backup_config"`
				ConfigPaths   []string               `json:"config_paths"`
				DataPaths     []string               `json:"data_paths"`
				CleanupConfig map[string]interface{} `json:"cleanup_config"`
			}

			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid JSON", http.StatusBadRequest)
				return
			}

			workflowID, err := s.cadenceClient.ExecuteWorkflow(context.Background(), "BackupWorkflow", req)
			if err != nil {
				s.logger.Error("Failed to execute workflow", zap.Error(err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			response := map[string]interface{}{
				"workflow_id": workflowID,
				"status":      "started",
			}
			if err := json.NewEncoder(w).Encode(response); err != nil {
				s.logger.Error("Failed to encode backup workflow response", zap.Error(err))
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
		})

		httpMux.HandleFunc("/api/v1/cadence/workflows/status", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if s.cadenceClient == nil {
				http.Error(w, "Cadence client not available", http.StatusServiceUnavailable)
				return
			}

			workflowID := r.URL.Query().Get("workflow_id")
			if workflowID == "" {
				http.Error(w, "workflow_id required", http.StatusBadRequest)
				return
			}

			status, err := s.cadenceClient.GetWorkflowStatus(context.Background(), workflowID)
			if err != nil {
				s.logger.Error("Failed to get workflow status", zap.Error(err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			response := map[string]interface{}{
				"workflow_id": workflowID,
				"status":      status,
			}
			if err := json.NewEncoder(w).Encode(response); err != nil {
				s.logger.Error("Failed to encode status response", zap.Error(err))
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
		})

		// Create TLS server instance
		if s.config.Security.TLSEnabled {
			if err := s.startTLSTunnelListener(); err != nil {
				s.logger.Error("TLS tunnel listener startup failed", zap.Error(err))
			}
		} else {
			// Fallback to HTTP for development only when TLS is disabled
			httpServer := &http.Server{
				Addr:              "0.0.0.0:9091",
				Handler:           httpMux,
				ReadHeaderTimeout: 30 * time.Second,
				ReadTimeout:       30 * time.Second,
				WriteTimeout:      30 * time.Second,
			}
			s.httpServer = httpServer
			s.logger.Warn("HTTP server started on port 9091 without TLS (development mode)")
			go func() {
				if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					s.logger.Error("HTTP server error", zap.Error(err))
				}
			}()
		}
	}()
	if s.wireGuardEnabled {
		go func() {
			udpAddr, err := net.ResolveUDPAddr("udp", ":51820")
			if err != nil {
				s.logger.Error("Failed to resolve UDP address for WireGuard", zap.Error(err))
				return
			}

			udpConn, err := net.ListenUDP("udp", udpAddr)
			if err != nil {
				s.logger.Error("Failed to start WireGuard UDP server", zap.Error(err))
				return
			}
			defer udpConn.Close()

			s.logger.Info("WireGuard UDP server started on port 51820")

			buffer := make([]byte, 4096)
			for {
				n, addr, err := udpConn.ReadFromUDP(buffer)
				if err != nil {
					if s.ctx.Err() != nil {
						return // Server is shutting down
					}
					s.logger.Error("Failed to read from WireGuard UDP", zap.Error(err))
					continue
				}
				go s.handleWireGuardPacket(udpConn, addr, buffer[:n])
			}
		}()
	}

	// Start QUIC UDP server on port 9090
	if s.enhancedQUICManager != nil {
		go func() {
			udpAddr, err := net.ResolveUDPAddr("udp4", "0.0.0.0:9090")
			if err != nil {
				s.logger.Error("Failed to resolve UDP address for QUIC", zap.Error(err))
				return
			}

			udpConn, err := net.ListenUDP("udp4", udpAddr)
			if err != nil {
				s.logger.Error("Failed to start QUIC UDP server", zap.Error(err))
				return
			}
			defer udpConn.Close()

			// Настраиваем UDP буферы для QUIC
			if err := udpConn.SetReadBuffer(7 * 1024 * 1024); err != nil {
				s.logger.Warn("Failed to set UDP read buffer size", zap.Error(err))
			}
			if err := udpConn.SetWriteBuffer(7 * 1024 * 1024); err != nil {
				s.logger.Warn("Failed to set UDP write buffer size", zap.Error(err))
			}

			s.logger.Info("QUIC UDP server started on port 9090")

			buffer := make([]byte, 4096)
			for {
				n, addr, err := udpConn.ReadFromUDP(buffer)
				if err != nil {
					if s.ctx.Err() != nil {
						return // Server is shutting down
					}
					s.logger.Error("Failed to read from QUIC UDP", zap.Error(err))
					continue
				}

				// Handle QUIC packet
				go s.handleQUICPacket(udpConn, addr, buffer[:n])
			}
		}()
	}

	// Start P2P QUIC server on port 5553 for P2P relay
	if s.quicRelayHandler != nil {
		p2pQUICServer := NewP2PQUICServerProper(s.logger, s.ctx, s.quicRelayHandler, s.config)
		if err := p2pQUICServer.Start(); err != nil {
			s.logger.Error("Failed to start P2P QUIC server", zap.Error(err))
		}
	}

	// Start STUN UDP server on port 19302
	// STUN server is always enabled for P2P mesh functionality
	{
		go func() {
			udpAddr, err := net.ResolveUDPAddr("udp4", "0.0.0.0:19302")
			if err != nil {
				s.logger.Error("Failed to resolve UDP address for STUN", zap.Error(err))
				return
			}

			udpConn, err := net.ListenUDP("udp4", udpAddr)
			if err != nil {
				s.logger.Error("Failed to start STUN UDP server", zap.Error(err))
				return
			}
			defer udpConn.Close()

			s.logger.Info("STUN UDP server started on port 19302")

			buffer := make([]byte, 4096)
			for {
				n, addr, err := udpConn.ReadFromUDP(buffer)
				if err != nil {
					if s.ctx.Err() != nil {
						return // Server is shutting down
					}
					s.logger.Error("Failed to read from STUN UDP", zap.Error(err))
					continue
				}

				// Handle STUN packet
				go s.handleSTUNPacket(udpConn, addr, buffer[:n])
			}
		}()
	}

	// Start Enhanced QUIC UDP server on port 9092
	if s.config.EnhancedQUIC.Enabled {
		go func() {
			udpAddr, err := net.ResolveUDPAddr("udp4", "0.0.0.0:9092")
			if err != nil {
				s.logger.Error("Failed to resolve UDP address for Enhanced QUIC", zap.Error(err))
				return
			}

			udpConn, err := net.ListenUDP("udp4", udpAddr)
			if err != nil {
				s.logger.Error("Failed to start Enhanced QUIC UDP server", zap.Error(err))
				return
			}
			defer udpConn.Close()

			// Настраиваем UDP буферы для QUIC
			if err := udpConn.SetReadBuffer(7 * 1024 * 1024); err != nil {
				s.logger.Warn("Failed to set UDP read buffer size", zap.Error(err))
			}
			if err := udpConn.SetWriteBuffer(7 * 1024 * 1024); err != nil {
				s.logger.Warn("Failed to set UDP write buffer size", zap.Error(err))
			}

			s.logger.Info("Enhanced QUIC UDP server started on port 9092")

			buffer := make([]byte, 4096)
			for {
				n, addr, err := udpConn.ReadFromUDP(buffer)
				if err != nil {
					if s.ctx.Err() != nil {
						return // Server is shutting down
					}
					s.logger.Error("Failed to read from Enhanced QUIC UDP", zap.Error(err))
					continue
				}

				// Handle Enhanced QUIC packet
				go s.handleEnhancedQUICPacket(udpConn, addr, buffer[:n])
			}
		}()
	}

	// Start main TCP server on configured port
	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", s.config.Server.Port))
	if err != nil {
		return fmt.Errorf("failed to start listener: %v", err)
	}
	s.listener = listener

	s.logger.Info("Relay server started",
		zap.Int("tcp_port", s.config.Server.Port),
		zap.Int("http_port", 80),
		zap.Int("wireguard_port", 51820),
		zap.Int("quic_port", 9090),
		zap.Int("stun_port", 19302),
		zap.Int("enhanced_quic_port", 9092),
		zap.Bool("stun_enabled", true),
		zap.Bool("enhanced_quic_enabled", s.config.EnhancedQUIC.Enabled),
		zap.Bool("zero_trust_enabled", s.config.ZeroTrust.Enabled),
		zap.Bool("circuit_breaker_enabled", s.config.ZeroTrust.CircuitBreaker.Enabled),
		zap.Bool("health_check_enabled", s.config.ZeroTrust.HealthCheck.Enabled),
		zap.Bool("multiprotocol_enabled", s.connectionHandler != nil),
		zap.Bool("auto_switch_enabled", s.autoSwitchManager != nil),
	)

	// Запуск автоматического переключения протоколов
	if s.autoSwitchManager != nil {
		go func() {
			s.autoSwitchManager.Start(s.ctx)
		}()
		s.logger.Info("AutoSwitchManager started")
	}
	if s.p2pManager != nil {
		go func() {
			s.logger.Info("P2P Manager initialized successfully (QUIC relay mode)")
		}()
	}

	// Запуск P2P API сервера
	if s.p2pAPIServer != nil {
		go func() {
			if err := s.p2pAPIServer.Start(); err != nil {
				s.logger.Error("Failed to start P2P API server", zap.Error(err))
			}
		}()
		s.logger.Info("P2P API Server started")
	}

	// Запуск P2P Relay компонентов
	if s.heartbeatManager != nil {
		go func() {
			if err := s.heartbeatManager.Start(); err != nil {
				s.logger.Error("Failed to start Heartbeat Manager", zap.Error(err))
			}
		}()
		s.logger.Info("Heartbeat Manager started")
	}

	// Запуск QUIC Relay Handler
	if s.quicRelayHandler != nil {
		s.logger.Info("QUIC Relay Handler ready")
	}

	// Handle TCP connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			if s.ctx.Err() != nil {
				return nil // Server is shutting down
			}
			s.logger.Error("Failed to accept connection", zap.Error(err))
			continue
		}

		go s.handleConnection(conn)
	}
}

// handleConnection handles a new connection
func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	connectionsTotal.WithLabelValues("accepted").Inc()
	activeConnections.Inc()
	atomic.AddInt64(&activeConnCount, 1)
	defer activeConnections.Dec()
	defer atomic.AddInt64(&activeConnCount, -1)

	// Use circuit breaker if enabled
	if s.circuitBreaker != nil {
		err := s.circuitBreaker.Execute(s.ctx, func() error {
			return s.processConnection(conn)
		})
		if err != nil {
			s.logger.Error("Circuit breaker blocked connection", zap.Error(err))
			connectionsTotal.WithLabelValues("blocked").Inc()
			return
		}
	} else {
		if err := s.processConnection(conn); err != nil {
			s.logger.Error("Failed to process connection", zap.Error(err))
			connectionsTotal.WithLabelValues("failed").Inc()
		}
	}

	connectionsTotal.WithLabelValues("completed").Inc()
	atomic.AddInt64(&totalConnCount, 1)
}

// processConnection processes a connection
func (s *Server) processConnection(conn net.Conn) error {
	// Обрабатываем соединение через AI Middleware
	if s.aiComponents != nil {
		if err := s.aiComponents.ProcessConnectionWithAI(conn); err != nil {
			s.logger.Warn("AI processing failed", zap.Error(err))
			// Продолжаем обработку без AI
		}
	}

	// Получаем IP клиента
	clientIP := conn.RemoteAddr().String()
	if tcpAddr, ok := conn.RemoteAddr().(*net.TCPAddr); ok {
		clientIP = tcpAddr.IP.String()
	}

	// Проверяем rate limiting через Enhanced QUIC или обычный IP rate limiter
	if s.ipRateLimit != nil {
		if !s.ipRateLimit.Allow(clientIP) {
			s.logger.Warn("Rate limit exceeded", zap.String("client_ip", clientIP))
			return fmt.Errorf("rate limit exceeded for IP %s", clientIP)
		}
	}

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	// 1. Прочитать первое сообщение (ожидаем hello)
	line, err := reader.ReadString('\n')
	if err != nil {
		s.logger.Error("Failed to read from connection", zap.Error(err))
		return err
	}
	s.logger.Info("Received raw message", zap.String("data", line))

	// Check if it's an HTTP request (healthcheck)
	if line != "" && (strings.HasPrefix(line, "GET ") || strings.HasPrefix(line, "POST ") ||
		strings.HasPrefix(line, "PUT ") || strings.HasPrefix(line, "DELETE ") ||
		strings.HasPrefix(line, "HEAD ") || strings.HasPrefix(line, "OPTIONS ") ||
		strings.HasPrefix(line, "HTTP/")) {
		s.logger.Debug("Received HTTP request, handling as HTTP", zap.String("data", line))
		return s.handleHTTPRequest(conn, line, reader)
	}

	// Additional check for HTTP requests that might not start with method
	if line != "" && (strings.Contains(line, "HTTP/") || strings.Contains(line, "Host:") ||
		strings.Contains(line, "User-Agent:")) {
		s.logger.Debug("Detected HTTP request by content, handling as HTTP", zap.String("data", line))
		return s.handleHTTPRequest(conn, line, reader)
	}

	var msg map[string]interface{}
	if err := json.Unmarshal([]byte(line), &msg); err != nil {
		s.logger.Error("Failed to parse JSON", zap.Error(err))
		return err
	}
	msgType, _ := msg["type"].(string)
	s.logger.Info("Received message", zap.String("type", msgType))

	if msgType != "hello" {
		s.logger.Warn("Expected hello message", zap.String("type", msgType))
		return nil
	}

	// 2. Отправить hello-ответ
	respHello := map[string]interface{}{
		"type":     "hello",
		"version":  "1.0.0",
		"features": []string{"tls", "jwt", "tunneling"},
	}
	if err := writeJSON(writer, respHello); err != nil {
		s.logger.Error("Failed to send hello response", zap.Error(err))
		return err
	}
	s.logger.Info("Sent hello response")

	// 3. Прочитать auth
	line, err = reader.ReadString('\n')
	if err != nil {
		s.logger.Error("Failed to read auth", zap.Error(err))
		return err
	}
	s.logger.Info("Received raw auth", zap.String("data", line))
	msg = map[string]interface{}{}
	if err := json.Unmarshal([]byte(line), &msg); err != nil {
		s.logger.Error("Failed to parse auth JSON", zap.Error(err))
		return err
	}
	msgType, _ = msg["type"].(string)
	if msgType != "auth" {
		s.logger.Warn("Expected auth message", zap.String("type", msgType))
		return nil
	}
	token, _ := msg["token"].(string)
	if token == "" {
		resp := map[string]interface{}{
			"type":    "auth_response",
			"status":  "error",
			"message": "Token required",
		}
		if err := writeJSON(writer, resp); err != nil {
			s.logger.Error("Failed to send auth response", zap.Error(err))
			return err
		}
		s.logger.Warn("Auth failed: token required")
		return nil
	}

	// Validate JWT token using auth adapter
	/*
	   tokenInfo, err := s.authAdapter.ValidateToken(s.ctx, token)
	   if err != nil {
	       resp := map[string]interface{}{
	           "type":    "auth_response",
	           "status":  "error",
	           "message": "Invalid token: " + err.Error(),
	       }
	       writeJSON(writer, resp)
	       s.logger.Warn("Auth failed: invalid token", zap.Error(err))
	       authAttemptsTotal.WithLabelValues("failed", "unknown").Inc()
	       return nil
	   }

	   s.logger.Info("Token validated successfully",
	       zap.String("subject", tokenInfo.Subject),
	       zap.String("username", tokenInfo.Username),
	       zap.String("provider", tokenInfo.Provider),
	   )

	   authAttemptsTotal.WithLabelValues("success", tokenInfo.Provider).Inc()
	*/

	// Temporary: accept any token for testing
	var tokenLog string
	if len(token) > 20 {
		tokenLog = token[:20] + "..."
	} else {
		tokenLog = token
	}

	// Проверяем через Enhanced Keycloak Auth, если включен
	if s.enhancedEnabled && s.enhancedKeycloakAuth != nil {
		if isValid, err := s.enhancedKeycloakAuth.ValidateTokenEnhanced(token); err != nil || !isValid {
			resp := map[string]interface{}{
				"type":    "auth_response",
				"status":  "error",
				"message": "Enhanced authentication failed: " + err.Error(),
			}
			if err := writeJSON(writer, resp); err != nil {
				s.logger.Error("Failed to send auth response", zap.Error(err))
				return err
			}
			s.logger.Warn("Enhanced auth failed", zap.Error(err), zap.Bool("is_valid", isValid))
			authAttemptsTotal.WithLabelValues("failed", "enhanced").Inc()
			return nil
		}
		s.logger.Info("Enhanced token validation successful", zap.String("token", tokenLog))
		authAttemptsTotal.WithLabelValues("success", "enhanced").Inc()
	} else {
		s.logger.Info("Token accepted (validation disabled)", zap.String("token", tokenLog))
		authAttemptsTotal.WithLabelValues("success", "test").Inc()
	}

	// Можно добавить валидацию токена
	respAuth := map[string]interface{}{
		"type":      "auth_response",
		"status":    "ok",
		"client_id": "relay-client-001",
	}
	if err := writeJSON(writer, respAuth); err != nil {
		s.logger.Error("Failed to send auth response", zap.Error(err))
		return err
	}
	s.logger.Info("Sent auth response")

	// 4. Ждем tunnel_info
	for {
		line, err = reader.ReadString('\n')
		if err != nil {
			s.logger.Error("Connection closed or error reading tunnel_info", zap.Error(err))
			break
		}
		s.logger.Info("Received raw tunnel_info", zap.String("data", line))
		msg = map[string]interface{}{}
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			s.logger.Error("Failed to parse tunnel_info JSON", zap.Error(err))
			continue
		}
		msgType, _ = msg["type"].(string)
		if msgType == "tunnel_info" {
			// Получаем параметры туннеля
			localPort, _ := msg["local_port"]
			remoteHost, _ := msg["remote_host"].(string)
			remotePort, _ := msg["remote_port"]
			protocol, _ := msg["protocol"].(string)
			tenantID, _ := msg["tenant_id"].(string)

			// Если tenant_id не указан, используем "default"
			if tenantID == "" {
				tenantID = "default"
			}

			// Использование унифицированного обработчика для мультипротокольной обработки
			if s.connectionHandler != nil {
				s.logger.Info("Using unified connection handler for multiprotocol processing",
					zap.String("tenant_id", tenantID),
					zap.String("client_ip", clientIP),
				)

				// Обработка соединения через унифицированный обработчик
				if err := s.connectionHandler.HandleConnection(s.ctx, conn, tenantID); err != nil {
					respTunnel := map[string]interface{}{
						"type":    "tunnel_response",
						"status":  "error",
						"message": "Multiprotocol processing failed: " + err.Error(),
					}
					if err := writeJSON(writer, respTunnel); err != nil {
						s.logger.Error("Failed to send tunnel response", zap.Error(err))
						return err
					}
					s.logger.Warn("Multiprotocol processing failed",
						zap.String("tenant_id", tenantID),
						zap.String("client_ip", clientIP),
						zap.Error(err),
					)
					tunnelsCreatedTotal.WithLabelValues("failed").Inc()
					continue
				}

				s.logger.Info("Multiprotocol processing completed successfully",
					zap.String("tenant_id", tenantID),
					zap.String("client_ip", clientIP),
				)
			} else {
				// Fallback к существующей логике Enhanced QUIC
				if s.enhancedEnabled && s.enhancedQUICManager != nil {
					if err := s.enhancedQUICManager.ValidateConnection(clientIP, tenantID, token); err != nil {
						respTunnel := map[string]interface{}{
							"type":    "tunnel_response",
							"status":  "error",
							"message": "Enhanced validation failed: " + err.Error(),
						}
						if err := writeJSON(writer, respTunnel); err != nil {
							s.logger.Error("Failed to send tunnel response", zap.Error(err))
							return err
						}
						s.logger.Warn("Enhanced tunnel validation failed",
							zap.String("tenant_id", tenantID),
							zap.String("client_ip", clientIP),
							zap.Error(err),
						)
						tunnelsCreatedTotal.WithLabelValues("failed").Inc()
						continue
					}

					// Получаем оптимальный сервер через Enhanced Geographic Router
					if targetServer, err := s.enhancedQUICManager.GetServerByGeo(clientIP); err == nil {
						s.logger.Info("Enhanced geographic routing",
							zap.String("client_ip", clientIP),
							zap.String("target_server", targetServer),
						)
					}
				}
			}

			// Проверяем лимиты через обычный Tenant Manager
			if s.tenantManager != nil {
				if err := s.tenantManager.CheckTenantLimits(tenantID, 1, 1); err != nil {
					respTunnel := map[string]interface{}{
						"type":    "tunnel_response",
						"status":  "error",
						"message": "Tenant limits exceeded: " + err.Error(),
					}
					if err := writeJSON(writer, respTunnel); err != nil {
						s.logger.Error("Failed to send tunnel response", zap.Error(err))
						return err
					}
					s.logger.Warn("Tenant limits exceeded",
						zap.String("tenant_id", tenantID),
						zap.Error(err),
					)
					tunnelsCreatedTotal.WithLabelValues("failed").Inc()
					continue
				}

				// Обновляем активность тенанта
				s.tenantManager.UpdateTenantActivity(tenantID)
			}

			// Создаем реальный прокси туннель
			tunnelID := fmt.Sprintf("tunnel_%s_%d", tenantID, time.Now().Unix())
			if s.tunnelManager != nil {
				// Приводим remotePort к int (поддерживаем как int, так и string)
				var remotePortInt int
				switch v := remotePort.(type) {
				case int:
					remotePortInt = v
				case string:
					if parsed, err := strconv.Atoi(v); err == nil {
						remotePortInt = parsed
					} else {
						respTunnel := map[string]interface{}{
							"type":    "tunnel_response",
							"status":  "error",
							"message": "Invalid remote_port format: " + v,
						}
						if err := writeJSON(writer, respTunnel); err != nil {
							s.logger.Error("Failed to send tunnel response", zap.Error(err))
							return err
						}
						s.logger.Error("Invalid remote_port format",
							zap.String("tunnel_id", tunnelID),
							zap.Any("remote_port", remotePort),
						)
						tunnelsCreatedTotal.WithLabelValues("failed").Inc()
						continue
					}
				case float64:
					remotePortInt = int(v)
				default:
					respTunnel := map[string]interface{}{
						"type":    "tunnel_response",
						"status":  "error",
						"message": "Invalid remote_port type",
					}
					if err := writeJSON(writer, respTunnel); err != nil {
						s.logger.Error("Failed to send tunnel response", zap.Error(err))
						return err
					}
					s.logger.Error("Invalid remote_port type",
						zap.String("tunnel_id", tunnelID),
						zap.Any("remote_port", remotePort),
					)
					tunnelsCreatedTotal.WithLabelValues("failed").Inc()
					continue
				}

				// Создаем новый прокси туннель
				_, err := s.tunnelManager.CreateTunnel(tunnelID, conn, remoteHost, remotePortInt)
				if err != nil {
					respTunnel := map[string]interface{}{
						"type":    "tunnel_response",
						"status":  "error",
						"message": "Failed to create proxy tunnel: " + err.Error(),
					}
					if err := writeJSON(writer, respTunnel); err != nil {
						s.logger.Error("Failed to send tunnel response", zap.Error(err))
						return err
					}
					s.logger.Error("Failed to create proxy tunnel",
						zap.String("tunnel_id", tunnelID),
						zap.String("remote_host", remoteHost),
						zap.Int("remote_port", remotePortInt),
						zap.Error(err),
					)
					tunnelsCreatedTotal.WithLabelValues("failed").Inc()
					continue
				}

				s.logger.Info("Proxy tunnel created successfully",
					zap.String("tunnel_id", tunnelID),
					zap.String("remote_host", remoteHost),
					zap.Int("remote_port", remotePortInt),
					zap.String("protocol", protocol),
				)
			}

			respTunnel := map[string]interface{}{
				"type":          "tunnel_response",
				"status":        "ok",
				"tunnel_id":     tunnelID,
				"local_port":    localPort,
				"remote_host":   remoteHost,
				"remote_port":   remotePort,
				"protocol":      protocol,
				"tenant_id":     tenantID,
				"enhanced_mode": s.enhancedEnabled,
			}

			if err := writeJSON(writer, respTunnel); err != nil {
				s.logger.Error("Failed to send tunnel response", zap.Error(err))
				tunnelsCreatedTotal.WithLabelValues("failed").Inc()
			} else {
				s.logger.Info("Sent tunnel response",
					zap.String("tenant_id", tenantID),
					zap.String("protocol", protocol),
					zap.Bool("enhanced_mode", s.enhancedEnabled),
				)
				tunnelsCreatedTotal.WithLabelValues("success").Inc()
			}
		} else if msgType == "heartbeat" {
			tunnelID, _ := msg["tunnel_id"].(string)
			if tunnelID == "" {
				_ = writeJSON(writer, map[string]interface{}{
					"type":    "heartbeat_response",
					"status":  "error",
					"message": "tunnel_id required",
				})
				s.logger.Warn("Heartbeat without tunnel_id")
				continue
			}

			if s.tunnelManager != nil {
				if t, ok := s.tunnelManager.GetTunnel(tunnelID); ok && t != nil {
					s.logger.Debug("Heartbeat for active tunnel", zap.String("tunnel_id", tunnelID))
				} else {
					s.logger.Debug("Heartbeat for unknown tunnel", zap.String("tunnel_id", tunnelID))
				}
			}

			if err := writeJSON(writer, map[string]interface{}{
				"type":      "heartbeat_response",
				"status":    "ok",
				"tunnel_id": tunnelID,
			}); err != nil {
				s.logger.Error("Failed to send heartbeat_response", zap.Error(err))
				continue
			}
			s.logger.Info("Sent heartbeat_response", zap.String("tunnel_id", tunnelID))
		} else if msgType == "tunnel_close" {
			tunnelID, _ := msg["tunnel_id"].(string)
			if tunnelID == "" {
				_ = writeJSON(writer, map[string]interface{}{
					"type":    "tunnel_response",
					"status":  "error",
					"message": "tunnel_id required",
				})
				s.logger.Warn("tunnel_close without tunnel_id")
				continue
			}

			var closeErr error
			if s.tunnelManager != nil {
				closeErr = s.tunnelManager.CloseTunnel(tunnelID)
			}

			if closeErr != nil {
				s.logger.Warn("CloseTunnel returned error", zap.String("tunnel_id", tunnelID), zap.Error(closeErr))
			}

			if err := writeJSON(writer, map[string]interface{}{
				"type":      "tunnel_response",
				"status":    "closed",
				"tunnel_id": tunnelID,
			}); err != nil {
				s.logger.Error("Failed to send tunnel close response", zap.Error(err))
				continue
			}
			s.logger.Info("Tunnel close acknowledged", zap.String("tunnel_id", tunnelID))
		} else {
			s.logger.Warn("Unknown message type", zap.String("type", msgType))
		}
	}
	return nil
}

// Stop stops the server
func (s *Server) Stop() {
	s.logger.Info("Stopping relay server...")

	// Stop Enhanced QUIC Manager
	if s.enhancedQUICManager != nil {
		s.enhancedQUICManager.Disable()
	}

	// Shutdown tenant manager
	if s.tenantManager != nil {
		s.tenantManager.Shutdown()
	}

	// Stop health checker
	if s.healthChecker != nil {
		s.healthChecker.Stop()
	}

	// Shutdown AI components
	if s.aiComponents != nil {
		if err := s.aiComponents.ShutdownAI(); err != nil {
			s.logger.Error("Error shutting down AI components", zap.Error(err))
		}
	}

	// Stop metrics server
	if s.metricsServer != nil {
		// Note: metrics.Server doesn't have Stop method, it will stop when context is cancelled
	}

	// Close HTTP server
	if s.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.httpServer.Shutdown(ctx); err != nil {
			s.logger.Error("Error shutting down HTTP server", zap.Error(err))
		}
	}

	// Close listener
	if s.listener != nil {
		if err := s.listener.Close(); err != nil {
			s.logger.Error("Error closing listener", zap.Error(err))
		}
	}

	// Stop P2P Relay components
	if s.heartbeatManager != nil {
		s.heartbeatManager.Stop()
	}

	if s.quicRelayHandler != nil {
		s.quicRelayHandler.Stop()
	}

	if s.relayManager != nil {
		s.relayManager.Shutdown()
	}

	// Stop additional server components
	s.StopServer()

	// Cancel context
	s.cancel()

	s.logger.Info("Relay server stopped")
}

func main() {
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Create server
	server, err := NewServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		server.Stop()
	}()

	// Start server
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// generateTLSConfig генерирует TLS конфигурацию для QUIC
func (s *Server) generateTLSConfig() *tls.Config {
	// Создаем самоподписанный сертификат для QUIC
	cert, err := s.generateSelfSignedCert()
	if err != nil {
		s.logger.Error("Failed to generate self-signed certificate", zap.Error(err))
		// Возвращаем nil, что приведет к ошибке при создании QUIC listener
		return nil
	}

	return &tls.Config{
		MinVersion: tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{"cloudbridge-p2p", "h3", "h3-29", "h3-28", "h3-27"},
	}
}

// generateSelfSignedCert генерирует самоподписанный сертификат
func (s *Server) generateSelfSignedCert() (tls.Certificate, error) {
	// Генерируем приватный ключ
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}

	// Создаем сертификат
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:  []string{"CloudBridge Relay"},
			Country:       []string{"RU"},
			Province:      []string{""},
			Locality:      []string{"Moscow"},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(365 * 24 * time.Hour), // 1 год
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		DNSNames:    []string{"localhost", "edge.2gc.ru"},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return tls.Certificate{}, err
	}

	// Создаем TLS сертификат
	cert := tls.Certificate{
		Certificate: [][]byte{certDER},
		PrivateKey:  priv,
	}

	return cert, nil
}
