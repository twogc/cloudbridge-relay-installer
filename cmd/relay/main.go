package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/twogc/cloudbridge-relay-installer/config"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai"
	"github.com/twogc/cloudbridge-relay-installer/internal/auth"
	"github.com/twogc/cloudbridge-relay-installer/internal/balancer"
	"github.com/twogc/cloudbridge-relay-installer/internal/circuitbreaker"
	"github.com/twogc/cloudbridge-relay-installer/internal/health"
	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
	"github.com/twogc/cloudbridge-relay-installer/internal/metrics"
	"github.com/twogc/cloudbridge-relay-installer/internal/p2c"
	"github.com/twogc/cloudbridge-relay-installer/internal/p2p"
	"github.com/twogc/cloudbridge-relay-installer/internal/protocol"
	"github.com/twogc/cloudbridge-relay-installer/internal/quantum"
	"github.com/twogc/cloudbridge-relay-installer/internal/ratelimit"
	"github.com/twogc/cloudbridge-relay-installer/internal/tenant"
	"github.com/twogc/cloudbridge-relay-installer/internal/tunnel"
	"github.com/twogc/cloudbridge-relay-installer/internal/wireguard"
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

	// WireGuard P2P Mesh components
	wireGuardEnabled       bool
	wireGuardInterface     *wireguard.WireGuardInterface
	wireGuardMesh          *wireguard.MeshTopology
	wireGuardPeerDiscovery *wireguard.PeerDiscovery
	quicOverWireGuard      *wireguard.QUICOverWireGuard
	hybridRouter           *wireguard.HybridRouter
	siteToSiteManager      *wireguard.SiteToSiteManager

	// Post-Quantum Cryptography components
	quantumEnabled      bool
	quantumConfig       *quantum.PQCConfig
	quantumWireGuard    *quantum.QuantumWireGuardInterface
	quantumEnhancedQUIC *quantum.QuantumEnhancedQUIC

	// AI Analytics components
	aiAnalyticsEnabled   bool
	aiBehaviorAnalyzer   *ai.BehaviorAnalyzer
	aiPredictiveSecurity *ai.PredictiveSecurity
	aiAutomatedResponse  *ai.AutomatedResponse
	aiMonitoring         *ai.AIMonitoring

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

	// Initialize WireGuard P2P Mesh components
	wireGuardEnabled := cfg.WireGuard.Enabled
	var wireGuardInterface *wireguard.WireGuardInterface
	var wireGuardMesh *wireguard.MeshTopology
	var wireGuardPeerDiscovery *wireguard.PeerDiscovery
	var quicOverWireGuard *wireguard.QUICOverWireGuard
	var hybridRouter *wireguard.HybridRouter
	var siteToSiteManager *wireguard.SiteToSiteManager

	if wireGuardEnabled {
		// Create WireGuard interface
		wireGuardInterface = wireguard.NewWireGuardInterface(
			cfg.WireGuard.Interface.Name,
			cfg.WireGuard.Interface.ListenPort,
		)

		// Initialize mesh topology
		wireGuardMesh = wireguard.NewMeshTopology(wireGuardInterface)

		// Initialize peer discovery
		discoveryConfig := &wireguard.DiscoveryConfig{
			RegistryURL:       cfg.WireGuard.Discovery.RegistryURL,
			HeartbeatInterval: cfg.WireGuard.Discovery.HeartbeatInterval,
			MaxPeers:          cfg.WireGuard.Discovery.MaxPeers,
		}
		wireGuardPeerDiscovery = wireguard.NewPeerDiscovery(discoveryConfig, wireGuardInterface)

		// Initialize QUIC over WireGuard
		quicOverWireGuard = wireguard.NewQUICOverWireGuard(wireGuardInterface, enhancedQUICManager)

		// Initialize hybrid router
		hybridRouter = wireguard.NewHybridRouter(loadBal, wireGuardMesh)

		// Initialize site-to-site manager
		siteToSiteManager = wireguard.NewSiteToSiteManager(wireGuardMesh)

		loggerInstance.Info("WireGuard P2P Mesh initialized",
			zap.String("interface", cfg.WireGuard.Interface.Name),
			zap.Int("listen_port", cfg.WireGuard.Interface.ListenPort),
		)
	}

	// Initialize Post-Quantum Cryptography components
	quantumEnabled := cfg.Quantum.Enabled
	var quantumConfig *quantum.PQCConfig
	var quantumWireGuard *quantum.QuantumWireGuardInterface
	var quantumEnhancedQUIC *quantum.QuantumEnhancedQUIC

	if quantumEnabled {
		// Create PQC configuration
		quantumConfig = quantum.NewPQCConfig()

		// Initialize quantum-resistant WireGuard
		if wireGuardInterface != nil {
			quantumWireGuard = quantum.NewQuantumWireGuardInterface(wireGuardInterface, quantumConfig)
		}

		// Initialize quantum-resistant Enhanced QUIC
		if enhancedQUICManager != nil {
			quantumEnhancedQUIC = quantum.NewQuantumEnhancedQUIC(enhancedQUICManager, quantumConfig)
		}

		loggerInstance.Info("Post-Quantum Cryptography initialized",
			zap.String("algorithm", string(quantumConfig.Algorithm)),
			zap.Bool("hybrid_mode", quantumConfig.HybridMode),
		)
	}

	// Initialize AI Analytics components
	aiAnalyticsEnabled := cfg.AI.Enabled
	var aiBehaviorAnalyzer *ai.BehaviorAnalyzer
	var aiPredictiveSecurity *ai.PredictiveSecurity
	var aiAutomatedResponse *ai.AutomatedResponse
	var aiMonitoring *ai.AIMonitoring

	if aiAnalyticsEnabled {
		// Initialize AI components
		aiBehaviorAnalyzer = ai.NewBehaviorAnalyzer()
		aiPredictiveSecurity = ai.NewPredictiveSecurity()
		aiAutomatedResponse = ai.NewAutomatedResponse()
		aiMonitoring = ai.NewAIMonitoring()

		loggerInstance.Info("AI Analytics initialized",
			zap.Bool("behavior_analysis", aiBehaviorAnalyzer.Enabled),
			zap.Bool("predictive_security", aiPredictiveSecurity.Enabled),
			zap.Bool("automated_response", aiAutomatedResponse.Enabled),
		)
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

	// Initialize P2P Manager
	p2pManager := p2p.NewP2PManager(
		wireGuardMesh,
		wireGuardPeerDiscovery,
		hybridRouter,
		siteToSiteManager,
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
		zap.Bool("p2p_enabled", wireGuardEnabled),
		zap.Bool("p2c_enabled", enhancedEnabled),
		zap.Bool("auto_switch_enabled", cfg.Protocol.AutoSelect),
	)

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
		wireGuardEnabled:       wireGuardEnabled,
		wireGuardInterface:     wireGuardInterface,
		wireGuardMesh:          wireGuardMesh,
		wireGuardPeerDiscovery: wireGuardPeerDiscovery,
		quicOverWireGuard:      quicOverWireGuard,
		hybridRouter:           hybridRouter,
		siteToSiteManager:      siteToSiteManager,
		quantumEnabled:         quantumEnabled,
		quantumConfig:          quantumConfig,
		quantumWireGuard:       quantumWireGuard,
		quantumEnhancedQUIC:    quantumEnhancedQUIC,
		aiAnalyticsEnabled:     aiAnalyticsEnabled,
		aiBehaviorAnalyzer:     aiBehaviorAnalyzer,
		aiPredictiveSecurity:   aiPredictiveSecurity,
		aiAutomatedResponse:    aiAutomatedResponse,
		aiMonitoring:           aiMonitoring,
		protocolSelector:       protocolSelector,
		protocolEngine:         protocolEngine,
		p2pManager:             p2pManager,
		p2cManager:             p2cManager,
		connectionHandler:      connectionHandler,
		autoSwitchManager:      autoSwitchManager,
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

	// Start listening
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.Server.Port))
	if err != nil {
		return fmt.Errorf("failed to start listener: %v", err)
	}
	s.listener = listener

	s.logger.Info("Relay server started",
		zap.Int("port", s.config.Server.Port),
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

	// Инициализация P2P Manager
	if s.p2pManager != nil && s.wireGuardEnabled {
		go func() {
			if err := s.p2pManager.Initialize(s.ctx); err != nil {
				s.logger.Error("Failed to initialize P2P Manager", zap.Error(err))
			} else {
				s.logger.Info("P2P Manager initialized successfully")
			}
		}()
	}

	// Handle connections
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
	defer activeConnections.Dec()

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
}

// processConnection processes a connection
func (s *Server) processConnection(conn net.Conn) error {
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
	if len(line) > 0 && (strings.HasPrefix(line, "GET ") || strings.HasPrefix(line, "POST ") ||
		strings.HasPrefix(line, "PUT ") || strings.HasPrefix(line, "DELETE ") ||
		strings.HasPrefix(line, "HEAD ") || strings.HasPrefix(line, "OPTIONS ") ||
		strings.HasPrefix(line, "HTTP/")) {
		s.logger.Debug("Received HTTP request, handling as HTTP", zap.String("data", line))
		return s.handleHTTPRequest(conn, line, reader)
	}

	// Additional check for HTTP requests that might not start with method
	if len(line) > 0 && (strings.Contains(line, "HTTP/") || strings.Contains(line, "Host:") ||
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
		writeJSON(writer, resp)
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
			writeJSON(writer, resp)
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
					writeJSON(writer, respTunnel)
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
						writeJSON(writer, respTunnel)
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
					writeJSON(writer, respTunnel)
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

			respTunnel := map[string]interface{}{
				"type":          "tunnel_response",
				"status":        "ok",
				"tunnel_id":     fmt.Sprintf("tunnel_%s_%d", tenantID, time.Now().Unix()),
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
		} else {
			s.logger.Warn("Unknown message type", zap.String("type", msgType))
		}
	}
	return nil
}

// handleHTTPRequest handles HTTP requests (like health checks)
func (s *Server) handleHTTPRequest(conn net.Conn, firstLine string, reader *bufio.Reader) error {
	// Read the rest of the HTTP request headers
	var lines []string
	lines = append(lines, firstLine)

	// Read headers until empty line
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			// If we can't read more, just proceed with what we have
			break
		}
		lines = append(lines, line)
		if line == "\r\n" || line == "\n" {
			break // End of headers
		}
	}

	// Skip body for now - we don't need it for health checks and 404 responses

	// Parse the first line to get method and path
	parts := strings.Fields(firstLine)
	if len(parts) < 2 {
		// If we can't parse the request line, return 400 Bad Request
		errorResponse := `{"error": "Invalid request line"}`
		response := "HTTP/1.1 400 Bad Request\r\n"
		response += "Content-Type: application/json\r\n"
		response += "Connection: close\r\n"
		response += fmt.Sprintf("Content-Length: %d\r\n", len(errorResponse))
		response += "\r\n"
		response += errorResponse

		_, writeErr := conn.Write([]byte(response))
		if writeErr != nil {
			return writeErr
		}

		// Ensure all data is written before closing
		if tcpConn, ok := conn.(*net.TCPConn); ok {
			tcpConn.CloseWrite()
		}
		return nil
	}

	method := parts[0]
	path := parts[1]

	s.logger.Debug("HTTP request", zap.String("method", method), zap.String("path", path))

	// Handle OPTIONS request (CORS preflight)
	if method == "OPTIONS" {
		response := "HTTP/1.1 200 OK\r\n"
		response += "Access-Control-Allow-Origin: *\r\n"
		response += "Access-Control-Allow-Methods: GET, OPTIONS\r\n"
		response += "Access-Control-Allow-Headers: Content-Type\r\n"
		response += "Connection: close\r\n"
		response += "Content-Length: 0\r\n"
		response += "\r\n"

		_, writeErr := conn.Write([]byte(response))
		if writeErr != nil {
			return writeErr
		}

		// Ensure all data is written before closing
		if tcpConn, ok := conn.(*net.TCPConn); ok {
			tcpConn.CloseWrite()
		}
		return nil
	}

	// Handle health endpoint
	if path == "/health" {
		return s.handleHealthHTTP(conn)
	}

	// Return 404 for unknown paths
	errorResponse := `{"error": "Not Found"}`
	response := "HTTP/1.1 404 Not Found\r\n"
	response += "Content-Type: application/json\r\n"
	response += "Access-Control-Allow-Origin: *\r\n"
	response += "Connection: close\r\n"
	response += fmt.Sprintf("Content-Length: %d\r\n", len(errorResponse))
	response += "\r\n"
	response += errorResponse

	_, writeErr := conn.Write([]byte(response))
	if writeErr != nil {
		return writeErr
	}

	// Ensure all data is written before closing
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.CloseWrite()
	}
	return nil
}

// handleHealthHTTP handles HTTP health check requests
func (s *Server) handleHealthHTTP(conn net.Conn) error {
	status := map[string]interface{}{
		"status":    "ok",
		"version":   s.config.API.Version,
		"timestamp": time.Now().Format(time.RFC3339),
		"service":   "cloudbridge-relay",
		"uptime":    time.Since(s.startTime).String(),
		"components": map[string]interface{}{
			"tcp_server":        "running",
			"metrics":           s.metricsServer != nil,
			"health_checker":    s.healthChecker != nil,
			"circuit_breaker":   s.circuitBreaker != nil,
			"enhanced_quic":     s.enhancedEnabled,
			"tenant_manager":    s.tenantManager != nil,
			"load_balancer":     s.loadBalancer != nil,
			"rate_limiter":      s.ipRateLimit != nil,
			"enhanced_keycloak": s.enhancedKeycloakAuth != nil,
		},
	}

	// Добавляем Enhanced QUIC статистику
	if s.enhancedEnabled && s.enhancedQUICManager != nil {
		status["enhanced_quic_stats"] = s.enhancedQUICManager.GetStats()
	}

	// Добавляем статистику тенантов
	if s.tenantManager != nil {
		status["tenant_stats"] = map[string]interface{}{
			"total_tenants": s.tenantManager.GetTenantCount(),
			"enhanced_mode": s.tenantManager.IsEnhancedEnabled(),
		}
	}

	jsonData, err := json.Marshal(status)
	if err != nil {
		return err
	}

	response := "HTTP/1.1 200 OK\r\n"
	response += "Content-Type: application/json\r\n"
	response += "Access-Control-Allow-Origin: *\r\n"
	response += "Access-Control-Allow-Methods: GET, OPTIONS\r\n"
	response += "Access-Control-Allow-Headers: Content-Type\r\n"
	response += "Connection: close\r\n"
	response += fmt.Sprintf("Content-Length: %d\r\n", len(jsonData))
	response += "\r\n"
	response += string(jsonData)

	_, err = conn.Write([]byte(response))
	if err != nil {
		return err
	}

	// Ensure all data is written before closing
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.CloseWrite()
	}
	return nil
}

func writeJSON(writer *bufio.Writer, msg interface{}) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if _, err := writer.Write(data); err != nil {
		return err
	}
	return writer.Flush()
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
