package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"log"
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
	kubeauthz "github.com/twogc/cloudbridge-relay-installer/internal/kube/authz_policy"
	clusterreg "github.com/twogc/cloudbridge-relay-installer/internal/kube/cluster_registry"
	kubemetrics "github.com/twogc/cloudbridge-relay-installer/internal/kube/metrics"
	kubeproxy "github.com/twogc/cloudbridge-relay-installer/internal/kube/proxy"
	kubestreams "github.com/twogc/cloudbridge-relay-installer/internal/kube/streams"
	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
	"github.com/twogc/cloudbridge-relay-installer/internal/metrics"
	"github.com/twogc/cloudbridge-relay-installer/internal/nat"
	"github.com/twogc/cloudbridge-relay-installer/internal/p2c"
	"github.com/twogc/cloudbridge-relay-installer/internal/p2p"
	"github.com/twogc/cloudbridge-relay-installer/internal/protocol"
	"github.com/twogc/cloudbridge-relay-installer/internal/ratelimit"
	"github.com/twogc/cloudbridge-relay-installer/internal/tenant"
	grpcTransport "github.com/twogc/cloudbridge-relay-installer/internal/transport/grpc"
	"github.com/twogc/cloudbridge-relay-installer/internal/tunnel"
	"go.uber.org/cadence/worker"
	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

var (
	configFile = flag.String("config", "configs/config.json", "Path to configuration file")
	logLevel   = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	logFile    = flag.String("log-file", "", "Log file path")

	// Prometheus metrics
	connectionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "relay_connections_by_status_total",
			Help: "Total number of connections by status",
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

	// Global Zitadel Integration
	globalZitadelIntegration *ZitadelIntegration

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

	// kubeProxyBuilder is initialized once at server startup and reused for all /kube/{clusterId} requests.
	kubeProxyBuilder *kubeproxy.Builder
	// kubeMetrics holds all kube proxy metrics
	kubeMetrics *kubemetrics.KubeProxyMetrics
	// kubeStreamsHandler handles WebSocket streams for exec/attach/logs/port-forward
	kubeStreamsHandler *kubestreams.Handler
	// kubeAuthorizer handles authorization for kube proxy requests
	kubeAuthorizer *kubeauthz.Authorizer
)

// Server represents the main relay server
type Server struct {
	config        *config.Config
	logger        *logger.Logger
	listener      net.Listener
	httpServer    *http.Server
	tunnelManager *tunnel.Manager
	metricsServer *metrics.Server
	metrics       *metrics.Metrics
	authAdapter   *auth.Adapter

	// Zero Trust components
	circuitBreaker *circuitbreaker.CircuitBreaker
	healthChecker  *health.HealthChecker

	// TURN server components
	turnServer *nat.EnhancedTURNServer
	turnPool   *nat.TURNPoolManager

	// DERP server components
	derpServer *nat.DERPServer

	// WebSocket tunnel components
	wsTunnel *tunnel.WebSocketTunnel

	// Enhanced QUIC components
	enhancedQUICManager *EnhancedQUICManager
	tenantManager       *tenant.Manager
	loadBalancer        *balancer.Balancer
	ipRateLimit         *ratelimit.IPLimiter
	enhancedEnabled     bool
	p2pQUICEnabled      bool

	// P2P Integration
	p2pIntegration   *P2PIntegration
	quicRelayHandler *QUICRelayHandler

	// Cadence Workflow Engine
	cadenceClient *cadence.ProductionClient

	// P2P API Server
	p2pAPIServer *P2PAPIServer

	// AI Analytics components
	aiComponents *AIComponents

	// Мультипротокольные компоненты
	protocolSelector  *protocol.IntelligentProtocolSelector
	protocolEngine    *protocol.ProtocolEngine
	p2pManager        *p2p.P2PManager
	p2cManager        *p2c.P2CManager
	connectionHandler *protocol.UnifiedConnectionHandler
	autoSwitchManager *protocol.AutoSwitchManager

	// WireGuard Profile Manager
	wgProfileManager *WGProfileManager

	// MASQUE Integration
	masqueIntegration *MASQUEIntegration

	// Zitadel Integration
	zitadelIntegration *ZitadelIntegration

	// TURN Pool Manager
	turnPoolManager *nat.TURNPoolManager

	// gRPC Integration
	grpcIntegration *grpcTransport.Integration

	ctx       context.Context
	cancel    context.CancelFunc
	startTime time.Time
}

// NewServer creates a new relay server
func NewServer(cfg *config.Config) (*Server, error) {
	ctx, cancel := context.WithCancel(context.Background())
	// NOTE: Do not defer cancel() here - context should remain active until server is stopped

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
	var metricsInstance *metrics.Metrics
	if cfg.Metrics.Enabled {
		metricsServer = metrics.NewServer(fmt.Sprintf(":%d", cfg.Metrics.Port), cfg.Metrics.Path)
		// Create main metrics instance
		metricsInstance = metrics.NewMetrics(prometheus.DefaultRegisterer)
	}

	// Initialize Enhanced QUIC components
	enhancedEnabled := cfg.EnhancedQUIC.Enabled // Из конфигурации
	var enhancedQUICManager *EnhancedQUICManager
	var tenantMgr *tenant.Manager
	var loadBal *balancer.Balancer
	var ipRateLimit *ratelimit.IPLimiter

	if enhancedEnabled {
		// Конвертируем logger.Logger в *zap.Logger для Enhanced компонентов
		zapLogger := zap.NewNop() // Fallback

		// Инициализируем Enhanced QUIC Manager
		enhancedQUICManager = NewEnhancedQUICManager(zapLogger, metricsInstance)

		// Инициализируем дополнительные компоненты для Enhanced режима
		tenantMgr = tenant.NewManagerWithEnhanced(loggerInstance, nil)
		loadBal = balancer.NewBalancerWithEnhanced()
		ipRateLimit = ratelimit.NewIPLimiterWithEnhanced(100.0, 1000) // 100 req/sec, burst 1000

		// Настраиваем интеграции
		if err := enhancedQUICManager.SetupIntegrations(tenantMgr, loadBal, ipRateLimit); err != nil {
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
	}
	// P2P is now enabled by default for QUIC relay functionality
	p2pEnabled := true // Always enable P2P for QUIC relay
	// P2P functionality is handled by QUIC relay instead
	loggerInstance.Info("P2P QUIC relay mode enabled")

	// Initialize TURN server components
	var turnServer *nat.EnhancedTURNServer
	var turnPool *nat.TURNPoolManager

	// Initialize DERP server components
	var derpServer *nat.DERPServer

	// Initialize WebSocket tunnel components
	var wsTunnel *tunnel.WebSocketTunnel

	// CRITICAL: Убираем дублирование TURN Pool
	// Используем только конфиг-ориентированный подход

	// CRITICAL: Убираем дублирование - используем только конфиг-ориентированный подход
	// defaultTURNServer := &nat.TURNServerPool{...}
	// err = turnPool.AddServerToPool("default", serverInfo)

	// Initialize Enhanced TURN server
	if cfg.ICEStunTurn.Enabled {
		turnServer, err = nat.NewEnhancedTURNServer(
			"0.0.0.0:3478",
			"cloudbridge.local",
			"cloudbridge",
			"turnpassword",
			loggerInstance.GetZapLogger(),
		)
		if err != nil {
			loggerInstance.Error("Failed to initialize Enhanced TURN server", zap.Error(err))
		} else {
			loggerInstance.Info("Enhanced TURN server initialized successfully")
		}
	}

	// Initialize DERP server for fallback mechanisms
	derpServer, err = nat.NewDERPServer("0.0.0.0:3479", loggerInstance.GetZapLogger())
	if err != nil {
		loggerInstance.Error("Failed to initialize DERP server", zap.Error(err))
	} else {
		loggerInstance.Info("DERP server initialized successfully")
	}

	// Initialize WebSocket tunnel for fallback mechanisms
	wsTunnel = tunnel.NewWebSocketTunnel(loggerInstance.GetZapLogger(), nil)
	loggerInstance.Info("WebSocket tunnel initialized successfully")

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

	// Ensure API port is set correctly
	apiPort := cfg.API.Port
	if apiPort == 0 {
		apiPort = 8082 // fallback to default port
		loggerInstance.Warn("API port was 0, using fallback port 8082")
	}

	loggerInstance.Info("API Config", zap.Int("port", apiPort), zap.Bool("enabled", cfg.API.Enabled))
	loggerInstance.Info("DEBUG: About to create P2P API Server")
	p2pAPIServer = NewP2PAPIServer(loggerInstance.GetZapLogger(), jwtSecret, apiPort)
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

	// Initialize Cadence Workflow Engine
	var cadenceClient *cadence.ProductionClient
	if cfg.Cadence.Enabled {
		cadenceConfig := cadence.ProductionConfig{
			HostPort: cfg.Cadence.HostPort,
			Domain:   cfg.Cadence.Domain,
			TaskList: cfg.Cadence.TaskList,
			WorkerOptions: worker.Options{
				Identity: "cloudbridge-relay-worker",
			},
			ServiceOptions: cadence.ServiceClientOptions{
				Timeout:     30 * time.Second,
				ServiceName: "cloudbridge-relay",
				EnableTLS:   false,
			},
			// TODO: Добавить реальные сервисы когда они будут готовы
			P2PService:      nil,
			AIService:       nil,
			SecurityService: nil,
			ScalingService:  nil,
		}

		var err error
		cadenceClient, err = cadence.NewProductionClient(cadenceConfig, loggerInstance.GetZapLogger())
		if err != nil {
			loggerInstance.Error("Failed to initialize Cadence client", zap.Error(err))
			cancel()
			return nil, fmt.Errorf("failed to initialize Cadence client: %v", err)
		}

		// Запускаем Cadence worker
		if err := cadenceClient.Start(); err != nil {
			loggerInstance.Error("Failed to start Cadence worker", zap.Error(err))
			cancel()
			return nil, fmt.Errorf("failed to start Cadence worker: %v", err)
		}

		loggerInstance.Info("Cadence Workflow Engine initialized successfully",
			zap.String("domain", cfg.Cadence.Domain),
			zap.String("host_port", cfg.Cadence.HostPort),
			zap.Bool("stub_mode", cadenceClient.IsStubMode()),
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
		nil, // P2P mesh handled by QUIC relay
		nil, // P2P peer discovery handled by QUIC relay
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

	// Initialize P2P Integration
	loggerInstance.Info("Initializing P2P Integration...")

	var p2pIntegration *P2PIntegration
	var quicRelayHandler *QUICRelayHandler

	if p2pEnabled {
		p2pIntegration, err = NewP2PIntegration(loggerInstance, cfg, jwtSecret)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("failed to initialize P2P integration: %v", err)
		}

		// Create QUIC Relay Handler with P2P components
		quicRelayHandler = NewQUICRelayHandler(
			p2pIntegration.GetRelayManager(),
			p2pIntegration.GetPeerManager(),
			p2pIntegration.GetTokenValidator(),
			loggerInstance.GetZapLogger(),
		)

		// Set managers for P2P API Server
		if p2pAPIServer != nil {
			p2pAPIServer.SetManagers(p2pIntegration.GetRelayManager(), p2pIntegration.GetHeartbeatManager())
			if hybrid := p2pIntegration.GetRelayManagerHybrid(); hybrid != nil {
				p2pAPIServer.SetRelayManagerHybrid(hybrid)
			}
		}

		loggerInstance.Info("P2P Integration initialized successfully",
			zap.Bool("score_hybrid", p2pIntegration.GetRelayManagerHybrid() != nil),
		)
	}

	// Initialize WireGuard Profile Manager
	var wgProfileManager *WGProfileManager
	if cfg.WireGuard.Enabled {
		wgProfileManager, err = NewWGProfileManager(cfg, loggerInstance.GetZapLogger(), metricsInstance)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("failed to initialize WG profile manager: %v", err)
		}
		loggerInstance.Info("WireGuard Profile Manager initialized successfully")
	} else {
		loggerInstance.Info("WireGuard disabled, skipping WG profile manager initialization")
	}

	// Initialize MASQUE Integration (disabled to avoid port conflict with MASQUEServer)
	var masqueIntegration *MASQUEIntegration
	// if cfg.MASQUE.Enabled {
	//	masqueIntegration, err = NewMASQUEIntegration(loggerInstance, cfg)
	//	if err != nil {
	//		cancel()
	//		return nil, fmt.Errorf("failed to initialize MASQUE integration: %v", err)
	//	}
	//	loggerInstance.Info("MASQUE Integration initialized successfully")
	// } else {
	//	loggerInstance.Info("MASQUE disabled, skipping MASQUE integration initialization")
	// }
	loggerInstance.Info("MASQUE Integration disabled to avoid port conflict with MASQUEServer")

	// Initialize Zitadel Integration
	var zitadelIntegration *ZitadelIntegration
	if cfg.Zitadel.Enabled {
		// Конвертируем конфигурацию
		zitadelConfig := &ZitadelConfig{
			Enabled:        cfg.Zitadel.Enabled,
			Issuer:         cfg.Zitadel.Issuer,
			Audience:       cfg.Zitadel.Audience,
			ClientID:       cfg.Zitadel.ClientID,
			ClientSecret:   cfg.Zitadel.ClientSecret,
			FallbackSecret: cfg.Zitadel.FallbackSecret,
			TenantManagement: struct {
				Enabled       bool `json:"enabled"`
				AutoCreate    bool `json:"auto_create"`
				DefaultQuotas struct {
					MaxPeers       int    `json:"max_peers"`
					MaxConnections int    `json:"max_connections"`
					MaxTunnels     int    `json:"max_tunnels"`
					BandwidthLimit string `json:"bandwidth_limit"`
					StorageLimit   string `json:"storage_limit"`
				} `json:"default_quotas"`
			}{
				Enabled:    cfg.Zitadel.TenantManagement.Enabled,
				AutoCreate: cfg.Zitadel.TenantManagement.AutoCreate,
				DefaultQuotas: struct {
					MaxPeers       int    `json:"max_peers"`
					MaxConnections int    `json:"max_connections"`
					MaxTunnels     int    `json:"max_tunnels"`
					BandwidthLimit string `json:"bandwidth_limit"`
					StorageLimit   string `json:"storage_limit"`
				}{
					MaxPeers:       cfg.Zitadel.TenantManagement.DefaultQuotas.MaxPeers,
					MaxConnections: cfg.Zitadel.TenantManagement.DefaultQuotas.MaxConnections,
					MaxTunnels:     cfg.Zitadel.TenantManagement.DefaultQuotas.MaxTunnels,
					BandwidthLimit: cfg.Zitadel.TenantManagement.DefaultQuotas.BandwidthLimit,
					StorageLimit:   cfg.Zitadel.TenantManagement.DefaultQuotas.StorageLimit,
				},
			},
		}

		zitadelIntegration, err = NewZitadelIntegration(zitadelConfig, loggerInstance.GetZapLogger())
		if err != nil {
			cancel()
			return nil, fmt.Errorf("failed to initialize Zitadel integration: %v", err)
		}

		// Set global variable for use in handlers
		globalZitadelIntegration = zitadelIntegration
		loggerInstance.Info("Zitadel Integration initialized successfully")
	} else {
		loggerInstance.Info("Zitadel disabled, skipping Zitadel integration initialization")
	}

	// Initialize TURN Pool Manager
	var turnPoolManager *nat.TURNPoolManager
	if cfg.TURNPool.Enabled {
		// Create TURN pool configuration from config
		turnPoolConfig := &nat.TURNPoolConfig{
			DefaultPoolName:         cfg.TURNPool.DefaultPoolName,
			AllocationTimeout:       cfg.TURNPool.AllocationTimeout,
			HealthCheckInterval:     cfg.TURNPool.HealthCheckInterval,
			HealthCheckTimeout:      cfg.TURNPool.HealthCheckTimeout,
			MaxAllocationsPerClient: cfg.TURNPool.MaxAllocationsPerClient,
			LoadBalancingStrategy:   cfg.TURNPool.LoadBalancingStrategy,
			WeightedThreshold:       cfg.TURNPool.WeightedThreshold,
			MetricsEnabled:          cfg.TURNPool.MetricsEnabled,
			StatsInterval:           cfg.TURNPool.StatsInterval,
			LatencyHistorySize:      cfg.TURNPool.LatencyHistorySize,
			FailoverEnabled:         cfg.TURNPool.FailoverEnabled,
			FailoverTimeout:         cfg.TURNPool.FailoverTimeout,
			MaxRetries:              cfg.TURNPool.MaxRetries,
			BackoffMultiplier:       cfg.TURNPool.BackoffMultiplier,
		}

		turnPoolManager = nat.NewTURNPoolManager(turnPoolConfig, loggerInstance.GetZapLogger())

		// Initialize pools from configuration
		for poolName, poolConfig := range cfg.TURNPool.Pools {
			err := turnPoolManager.CreatePool(poolName)
			if err != nil {
				loggerInstance.Error("Failed to create TURN pool",
					zap.String("pool", poolName),
					zap.Error(err))
				continue
			}

			// Add servers to the pool
			for _, serverConfig := range poolConfig.Servers {
				server := &nat.TURNServerInfo{
					TURNServerPool: &nat.TURNServerPool{
						Address:        serverConfig.Address,
						Username:       serverConfig.Username,
						Password:       serverConfig.Password,
						Realm:          serverConfig.Realm,
						Priority:       serverConfig.Priority,
						Weight:         serverConfig.Weight,
						MaxConnections: serverConfig.MaxConnections,
						Region:         serverConfig.Region,
						Provider:       serverConfig.Provider,
						Capabilities:   serverConfig.Capabilities,
						Tags:           serverConfig.Tags,
					},
					Region:       serverConfig.Region,
					Provider:     serverConfig.Provider,
					Capabilities: serverConfig.Capabilities,
					Tags:         serverConfig.Tags,
				}

				err := turnPoolManager.AddServerToPool(poolName, server)
				if err != nil {
					loggerInstance.Error("Failed to add server to TURN pool",
						zap.String("pool", poolName),
						zap.String("server", serverConfig.Address),
						zap.Error(err))
				} else {
					loggerInstance.Info("Added TURN server to pool",
						zap.String("pool", poolName),
						zap.String("server", serverConfig.Address),
						zap.String("region", serverConfig.Region))
				}
			}
		}

		loggerInstance.Info("TURN Pool Manager initialized",
			zap.String("default_pool", cfg.TURNPool.DefaultPoolName),
			zap.Int("pools", len(cfg.TURNPool.Pools)),
			zap.String("strategy", cfg.TURNPool.LoadBalancingStrategy))
	} else {
		loggerInstance.Info("TURN Pool disabled, skipping TURN Pool Manager initialization")
	}

	// Initialize gRPC Integration
	var grpcIntegration *grpcTransport.Integration
	if cfg.GRPC.Enabled {
		grpcConfig := grpcTransport.Config{
			Port:             cfg.GRPC.Port,
			TLSEnabled:       cfg.GRPC.TLSEnabled,
			CertFile:         cfg.GRPC.CertFile,
			KeyFile:          cfg.GRPC.KeyFile,
			MTLSEnabled:      cfg.GRPC.MTLSEnabled,
			KeepalivePing:    cfg.GRPC.KeepalivePing,
			KeepaliveTimeout: cfg.GRPC.KeepaliveTimeout,
			MaxRecvMsgSize:   cfg.GRPC.MaxRecvMsgSize,
			MaxSendMsgSize:   cfg.GRPC.MaxSendMsgSize,
		}

		// Use existing heartbeat manager if available, otherwise create a minimal one
		var heartbeatMgr *p2p.HeartbeatManager
		if p2pIntegration != nil {
			heartbeatMgr = p2pIntegration.GetHeartbeatManager()
		} else {
			// Create a basic heartbeat manager for gRPC
			heartbeatMgr = p2p.NewHeartbeatManager(nil, loggerInstance.GetZapLogger())
		}

		grpcIntegration, err = grpcTransport.NewIntegration(
			grpcConfig,
			loggerInstance,
			authAdapter,
			tunnelManager,
			heartbeatMgr,
		)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("failed to initialize gRPC integration: %v", err)
		}

		loggerInstance.Info("gRPC Integration initialized",
			zap.Int("port", cfg.GRPC.Port),
			zap.Bool("tls_enabled", cfg.GRPC.TLSEnabled),
			zap.Bool("mtls_enabled", cfg.GRPC.MTLSEnabled),
		)
	} else {
		loggerInstance.Info("gRPC disabled, skipping gRPC integration initialization")
	}

	// --- Kube proxy registry & builder (v1) ---
	// TODO: load clusters from cfg.Kubernetes.Clusters once schema is defined.
	var clusters []*clusterreg.Cluster
	for _, c := range cfg.Kubernetes.Clusters {
		cluster := &clusterreg.Cluster{
			ID:            c.ID,
			APIServerURL:  c.APIServerURL,
			SkipTLSVerify: c.SkipTLSVerify,
		}
		// Load CA and client cert/key from files if provided
		if c.CA != "" {
			if pemData, err := os.ReadFile(c.CA); err == nil {
				cluster.CAPEM = pemData
			}
		}
		if c.ClientCert != "" && c.ClientKey != "" {
			if crt, err := os.ReadFile(c.ClientCert); err == nil {
				cluster.ClientCertPEM = crt
			}
			if key, err := os.ReadFile(c.ClientKey); err == nil {
				cluster.ClientKeyPEM = key
			}
		}
		clusters = append(clusters, cluster)
	}
	defaultRegistry := clusterreg.NewStaticRegistry(clusters)
	kubeProxyBuilder = kubeproxy.NewBuilder(defaultRegistry)
	kubeMetrics = kubemetrics.NewKubeProxyMetrics()
	kubeStreamsHandler = kubestreams.NewHandler(defaultRegistry, kubeMetrics)

	// Initialize authorization policies
	var tenantPolicies []*kubeauthz.TenantPolicy
	for _, configPolicy := range cfg.Kubernetes.Policies {
		policy := &kubeauthz.TenantPolicy{
			TenantID: configPolicy.TenantID,
			UserID:   configPolicy.UserID,
			Groups:   configPolicy.Groups,
		}

		for _, configRule := range configPolicy.Rules {
			rule := kubeauthz.Rule{
				Name:       configRule.Name,
				APIGroups:  configRule.APIGroups,
				Resources:  configRule.Resources,
				Namespaces: configRule.Namespaces,
				Deny:       configRule.Deny,
			}

			// Convert string verbs to Verb type
			for _, verbStr := range configRule.Verbs {
				rule.Verbs = append(rule.Verbs, kubeauthz.Verb(verbStr))
			}

			policy.Rules = append(policy.Rules, rule)
		}

		tenantPolicies = append(tenantPolicies, policy)
	}

	policyManager := kubeauthz.NewPolicyManager(tenantPolicies)
	kubeAuthorizer = kubeauthz.NewAuthorizer(policyManager)

	// Start audit logging goroutine
	go func() {
		for auditEvent := range kubeAuthorizer.GetAuditChannel() {
			// TODO: use structured logger instead of log.Printf
			loggerInstance.Info("kube-proxy-audit",
				zap.String("cluster_id", auditEvent.ClusterID),
				zap.String("tenant_id", auditEvent.Request.TenantID),
				zap.String("user_id", auditEvent.Request.UserID),
				zap.String("verb", string(auditEvent.Request.Verb)),
				zap.String("resource", auditEvent.Request.Resource),
				zap.String("namespace", auditEvent.Request.Namespace),
				zap.Bool("allowed", auditEvent.Response.Allowed),
				zap.String("reason", auditEvent.Response.Reason),
				zap.String("matched_rule", auditEvent.Response.MatchedRule),
				zap.Duration("duration", auditEvent.Duration),
			)
		}
	}()

	return &Server{
		config:              cfg,
		logger:              loggerInstance,
		tunnelManager:       tunnelManager,
		metricsServer:       metricsServer,
		metrics:             metricsInstance,
		authAdapter:         authAdapter,
		circuitBreaker:      circuitBreaker,
		healthChecker:       healthChecker,
		turnServer:          turnServer,
		turnPool:            turnPool,
		derpServer:          derpServer,
		wsTunnel:            wsTunnel,
		enhancedQUICManager: enhancedQUICManager,
		tenantManager:       tenantMgr,
		loadBalancer:        loadBal,
		ipRateLimit:         ipRateLimit,
		enhancedEnabled:     enhancedEnabled,
		p2pQUICEnabled:      true, // P2P QUIC enabled
		p2pAPIServer:        p2pAPIServer,
		aiComponents:        aiComponents,
		protocolSelector:    protocolSelector,
		protocolEngine:      protocolEngine,
		p2pManager:          p2pManager,
		p2cManager:          p2cManager,
		connectionHandler:   connectionHandler,
		autoSwitchManager:   autoSwitchManager,
		cadenceClient:       cadenceClient,
		p2pIntegration:      p2pIntegration,
		quicRelayHandler:    quicRelayHandler,
		wgProfileManager:    wgProfileManager,
		masqueIntegration:   masqueIntegration,
		zitadelIntegration:  zitadelIntegration,
		turnPoolManager:     turnPoolManager,
		grpcIntegration:     grpcIntegration,
		ctx:                 ctx,
		cancel:              cancel,
		startTime:           time.Now(),
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

	// Start HTTP server on configured port for health checks
	go func() {
		httpMux := http.NewServeMux()

		// WebSocket tunnel endpoint
		if s.wsTunnel != nil {
			httpMux.HandleFunc("/ws", s.wsTunnel.CreateHTTPHandler())
		}

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
					"http":     true,
					"p2p_quic": s.p2pQUICEnabled,
					"quic":     s.enhancedQUICManager != nil,
					"metrics":  s.metricsServer != nil,
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
				Config        map[string]interface{} `json:"config"`
				P2PQUICConfig map[string]interface{} `json:"p2p_quic_config"`
				QUICConfig    map[string]interface{} `json:"quic_config"`
				Services      []string               `json:"services"`
				Timeout       string                 `json:"timeout"`
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

		// WireGuard API endpoints
		httpMux.HandleFunc("/api/v1/wireguard/status", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if s.wgProfileManager == nil {
				http.Error(w, "WireGuard not enabled", http.StatusServiceUnavailable)
				return
			}

			stats := s.wgProfileManager.GetStats()
			response := map[string]interface{}{
				"status": "enabled",
				"stats":  stats,
			}

			if err := json.NewEncoder(w).Encode(response); err != nil {
				s.logger.Error("Failed to encode WireGuard status response", zap.Error(err))
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
		})

		httpMux.HandleFunc("/api/v1/wireguard/peers", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if s.wgProfileManager == nil {
				http.Error(w, "WireGuard not enabled", http.StatusServiceUnavailable)
				return
			}

			switch r.Method {
			case "GET":
				// Get all peers
				stats := s.wgProfileManager.GetStats()
				if err := json.NewEncoder(w).Encode(stats); err != nil {
					s.logger.Error("Failed to encode peers response", zap.Error(err))
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}

			case "POST":
				// Add new peer
				var peerRequest struct {
					InterfaceName string `json:"interface_name"`
					PublicKey     string `json:"public_key"`
					Endpoint      string `json:"endpoint"`
					AllowedIPs    string `json:"allowed_ips"`
				}

				if err := json.NewDecoder(r.Body).Decode(&peerRequest); err != nil {
					http.Error(w, "Invalid JSON", http.StatusBadRequest)
					return
				}

				peerConfig, err := s.wgProfileManager.CreatePeerConfig(
					peerRequest.PublicKey,
					peerRequest.Endpoint,
					peerRequest.AllowedIPs,
				)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}

				if err := s.wgProfileManager.AddPeer(peerRequest.InterfaceName, peerConfig); err != nil {
					s.logger.Error("Failed to add peer", zap.Error(err))
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				response := map[string]interface{}{
					"status":  "success",
					"message": "Peer added successfully",
				}
				if err := json.NewEncoder(w).Encode(response); err != nil {
					s.logger.Error("Failed to encode add peer response", zap.Error(err))
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}

			case "DELETE":
				// Remove peer
				publicKeyStr := r.URL.Query().Get("public_key")
				interfaceName := r.URL.Query().Get("interface")

				if publicKeyStr == "" || interfaceName == "" {
					http.Error(w, "public_key and interface parameters required", http.StatusBadRequest)
					return
				}

				publicKey, err := wgtypes.ParseKey(publicKeyStr)
				if err != nil {
					http.Error(w, "Invalid public key", http.StatusBadRequest)
					return
				}

				if err := s.wgProfileManager.RemovePeer(interfaceName, publicKey); err != nil {
					s.logger.Error("Failed to remove peer", zap.Error(err))
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				response := map[string]interface{}{
					"status":  "success",
					"message": "Peer removed successfully",
				}
				if err := json.NewEncoder(w).Encode(response); err != nil {
					s.logger.Error("Failed to encode remove peer response", zap.Error(err))
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}

			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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
				Addr:              fmt.Sprintf("0.0.0.0:%d", s.config.HTTPSAPI.Port),
				Handler:           httpMux,
				ReadHeaderTimeout: 30 * time.Second,
				ReadTimeout:       30 * time.Second,
				WriteTimeout:      30 * time.Second,
			}
			s.httpServer = httpServer
			s.logger.Warn("HTTP server started without TLS (development mode)", zap.Int("port", s.config.HTTPSAPI.Port))
			go func() {
				if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					s.logger.Error("HTTP server error", zap.Error(err))
				}
			}()
		}

		// Kube proxy endpoint - регистрируем сразу
		httpMux.HandleFunc("/kube/", kubeProxyHTTPHandler)

		log.Printf("kubectl proxy endpoint registered: /kube/")
	}()
	// P2P QUIC functionality is handled by the QUIC relay server

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

			// CRITICAL: Worker pool для обработки пакетов
			type packetJob struct {
				addr *net.UDPAddr
				data []byte
			}

			jobChan := make(chan packetJob, 4096)
			workerCount := runtime.GOMAXPROCS(0) * 2

			// Запускаем воркеры
			for i := 0; i < workerCount; i++ {
				go func() {
					for job := range jobChan {
						s.handleQUICPacket(udpConn, job.addr, job.data)
					}
				}()
			}

			buffer := make([]byte, 4096)
			for {
				n, addr, err := udpConn.ReadFromUDP(buffer)
				if err != nil {
					if s.ctx.Err() != nil {
						close(jobChan)
						return // Server is shutting down
					}
					s.logger.Error("Failed to read from QUIC UDP", zap.Error(err))
					continue
				}

				// Отправляем в пул воркеров
				data := make([]byte, n)
				copy(data, buffer[:n])
				select {
				case jobChan <- packetJob{addr: addr, data: data}:
				default:
					s.logger.Warn("Packet dropped: worker pool full")
				}
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

			// CRITICAL: Worker pool для STUN пакетов
			type stunPacketJob struct {
				addr *net.UDPAddr
				data []byte
			}

			stunJobChan := make(chan stunPacketJob, 1024)
			stunWorkerCount := runtime.GOMAXPROCS(0)

			// Запускаем воркеры для STUN
			for i := 0; i < stunWorkerCount; i++ {
				go func() {
					for job := range stunJobChan {
						s.handleSTUNPacket(udpConn, job.addr, job.data)
					}
				}()
			}

			buffer := make([]byte, 4096)
			for {
				n, addr, err := udpConn.ReadFromUDP(buffer)
				if err != nil {
					if s.ctx.Err() != nil {
						close(stunJobChan)
						return // Server is shutting down
					}
					s.logger.Error("Failed to read from STUN UDP", zap.Error(err))
					continue
				}

				// Отправляем в пул воркеров
				data := make([]byte, n)
				copy(data, buffer[:n])
				select {
				case stunJobChan <- stunPacketJob{addr: addr, data: data}:
				default:
					s.logger.Warn("STUN packet dropped: worker pool full")
				}
			}
		}()
	}

	// Start Enhanced TURN server on port 3478
	if s.turnServer != nil {
		go func() {
			if err := s.turnServer.Start(); err != nil {
				s.logger.Error("Failed to start Enhanced TURN server", zap.Error(err))
			}
		}()
		s.logger.Info("Enhanced TURN server started on port 3478")
	}

	// Start DERP server on port 3479
	if s.derpServer != nil {
		go func() {
			if err := s.derpServer.Start(); err != nil {
				s.logger.Error("Failed to start DERP server", zap.Error(err))
			}
		}()
		s.logger.Info("DERP server started on port 3479")
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
		zap.Int("http_port", s.config.HTTPSAPI.Port),
		zap.Int("p2p_quic_port", 5553), // CRITICAL: реальный порт P2P QUIC
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

	// Запуск P2P Integration
	if s.p2pIntegration != nil {
		go func() {
			if err := s.p2pIntegration.Start(); err != nil {
				s.logger.Error("Failed to start P2P Integration", zap.Error(err))
			}
		}()
		s.logger.Info("P2P Integration started")
	}

	// Запуск QUIC Relay Handler
	if s.quicRelayHandler != nil {
		s.logger.Info("QUIC Relay Handler ready")
	}

	// Запуск WireGuard Profile Manager
	if s.wgProfileManager != nil {
		go func() {
			if err := s.wgProfileManager.Start(); err != nil {
				s.logger.Error("Failed to start WG Profile Manager", zap.Error(err))
			}
		}()
		s.logger.Info("WireGuard Profile Manager started")
	}

	// Запуск MASQUE Integration
	if s.masqueIntegration != nil {
		go func() {
			if err := s.masqueIntegration.Start(); err != nil {
				s.logger.Error("Failed to start MASQUE Integration", zap.Error(err))
			}
		}()
		s.logger.Info("MASQUE Integration started")
	}

	// Запуск Zitadel Integration
	if s.zitadelIntegration != nil {
		go func() {
			if err := s.zitadelIntegration.Start(s.ctx); err != nil {
				s.logger.Error("Failed to start Zitadel Integration", zap.Error(err))
			}
		}()
		s.logger.Info("Zitadel Integration started")
	}

	// Запуск gRPC Integration
	if s.grpcIntegration != nil {
		go func() {
			if err := s.grpcIntegration.Start(s.ctx); err != nil {
				s.logger.Error("Failed to start gRPC Integration", zap.Error(err))
			}
		}()
		s.logger.Info("gRPC Integration started",
			zap.String("address", s.grpcIntegration.GetAddr()),
		)
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

	// Проверяем токен через стандартную аутентификацию
	if s.authAdapter != nil {
		if _, err := s.authAdapter.ValidateToken(context.Background(), token); err != nil {
			resp := map[string]interface{}{
				"type":    "auth_response",
				"status":  "error",
				"message": "Enhanced authentication failed: " + err.Error(),
			}
			if err := writeJSON(writer, resp); err != nil {
				s.logger.Error("Failed to send auth response", zap.Error(err))
				return err
			}
			s.logger.Warn("Auth validation failed", zap.Error(err))
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

	// Stop Cadence Workflow Engine
	if s.cadenceClient != nil {
		s.cadenceClient.Close()
		s.logger.Info("Cadence Workflow Engine stopped")
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

	// Stop P2P Integration
	if s.p2pIntegration != nil {
		if err := s.p2pIntegration.Stop(); err != nil {
			s.logger.Error("Error stopping P2P Integration", zap.Error(err))
		}
	}

	if s.quicRelayHandler != nil {
		s.quicRelayHandler.Stop()
	}

	// Stop TURN Pool Manager
	if s.turnPoolManager != nil {
		s.turnPoolManager.Stop()
		s.logger.Info("TURN Pool Manager stopped")
	}

	// Stop WireGuard Profile Manager
	if s.wgProfileManager != nil {
		if err := s.wgProfileManager.Stop(); err != nil {
			s.logger.Error("Error stopping WG Profile Manager", zap.Error(err))
		}
	}

	// Stop MASQUE Integration
	if s.masqueIntegration != nil {
		if err := s.masqueIntegration.Stop(); err != nil {
			s.logger.Error("Error stopping MASQUE Integration", zap.Error(err))
		}
	}

	// Stop Zitadel Integration
	if s.zitadelIntegration != nil {
		if err := s.zitadelIntegration.Stop(context.Background()); err != nil {
			s.logger.Error("Error stopping Zitadel Integration", zap.Error(err))
		}
	}

	// Stop gRPC Integration
	if s.grpcIntegration != nil {
		if err := s.grpcIntegration.Stop(); err != nil {
			s.logger.Error("Error stopping gRPC Integration", zap.Error(err))
		}
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
	// CRITICAL: QUIC требует TLS 1.3
	// В реальном приложении здесь должен быть загружен сертификат
	return &tls.Config{
		MinVersion: tls.VersionTLS13, // CRITICAL: TLS 1.3 для QUIC
		// CipherSuites не нужны для TLS 1.3 - они игнорируются
		NextProtos: []string{"h3", "h3-29", "h3-28", "h3-27", "cloudbridge-p2p"},
	}
}

// kubeProxyHTTPHandler handles incoming /kube/{clusterId}/... requests and reverses them
// to the appropriate Kubernetes API server via kubeProxyBuilder.
func kubeProxyHTTPHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	if kubeProxyBuilder == nil || kubeMetrics == nil || kubeStreamsHandler == nil || kubeAuthorizer == nil {
		http.Error(w, "kube proxy not initialized", http.StatusServiceUnavailable)
		return
	}

	// Extract clusterId
	trimmed := strings.TrimPrefix(r.URL.Path, "/kube/")
	parts := strings.SplitN(trimmed, "/", 2)
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "clusterId required", http.StatusBadRequest)
		return
	}
	clusterID := parts[0]

	// Extract API path for authorization
	apiPath := ""
	if len(parts) > 1 {
		apiPath = "/" + parts[1]
	}

	// Parse resource information for authorization
	apiGroup, resource, namespace, name, err := kubeauthz.ParseResourceFromPath(apiPath)
	if err != nil && apiPath != "" {
		// Log the error but allow the request to proceed for non-standard paths
		// TODO: add proper logging
	}

	// Extract authentication information (simplified - in production use proper JWT parsing)
	tenantID, err := extractTenantID(r, globalZitadelIntegration)
	if err != nil {
		// Log the error but allow the request to proceed for non-standard paths
		// TODO: add proper logging
		tenantID = "default"
	}
	userID := extractUserID(r)
	groups := extractGroups(r)

	// Create authorization request
	verb := kubeauthz.ParseVerbFromMethod(r.Method)
	authReq := &kubeauthz.AuthRequest{
		TenantID:   tenantID,
		UserID:     userID,
		Groups:     groups,
		Verb:       verb,
		APIGroup:   apiGroup,
		Resource:   resource,
		Namespace:  namespace,
		Name:       name,
		Path:       apiPath,
		RemoteAddr: r.RemoteAddr,
	}

	// Check authorization
	authResp := kubeAuthorizer.Authorize(authReq)
	if !authResp.Allowed {
		http.Error(w, "forbidden: "+authResp.Reason, http.StatusForbidden)
		return
	}

	// TODO: Add clusterID to audit event - requires modifying Authorize method

	// Check if this is a WebSocket upgrade request (exec/attach/logs/port-forward)
	if kubestreams.IsWebSocketUpgrade(r) {
		kubeStreamsHandler.ServeWebSocketProxy(w, r, clusterID)
		return
	}

	proxy, err := kubeProxyBuilder.Build(clusterID)
	if err != nil {
		kubeMetrics.RecordUpstreamError(clusterID, "404")
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Wrap ResponseWriter to capture status code
	wrapped := &responseWriter{ResponseWriter: w, statusCode: 200}
	proxy.ServeHTTP(wrapped, r)

	// Record metrics
	duration := time.Since(startTime)
	statusCode := strconv.Itoa(wrapped.statusCode)
	kubeMetrics.RecordRequest(clusterID, string(verb), resource, statusCode, duration)
}

// extractTenantID безопасно извлекает tenant ID из JWT токена
func extractTenantID(r *http.Request, zitadelIntegration *ZitadelIntegration) (string, error) {
	// Получаем Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("missing authorization header")
	}

	// Проверяем формат Bearer токена
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", fmt.Errorf("invalid authorization header format")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")

	// Валидируем токен через Zitadel интеграцию
	if zitadelIntegration != nil && zitadelIntegration.enabled {
		claims, err := zitadelIntegration.client.ValidateToken(context.Background(), token)
		if err != nil {
			return "", fmt.Errorf("token validation failed: %w", err)
		}
		return claims.OrgID, nil
	}

	// Fallback: если Zitadel не доступен, возвращаем ошибку
	return "", fmt.Errorf("zitadel integration not available")
}

// extractUserID extracts user ID from request headers
func extractUserID(r *http.Request) string {
	if userID := r.Header.Get("X-Remote-User"); userID != "" {
		return userID
	}
	if userID := r.Header.Get("X-Forwarded-User"); userID != "" {
		return userID
	}
	// TODO: Parse JWT token and extract user ID from claims
	return "anonymous"
}

// extractGroups extracts user groups from request headers
func extractGroups(r *http.Request) []string {
	if groups := r.Header.Get("X-Remote-Group"); groups != "" {
		return strings.Split(groups, ",")
	}
	if groups := r.Header.Get("X-Forwarded-Groups"); groups != "" {
		return strings.Split(groups, ",")
	}
	// TODO: Parse JWT token and extract groups from claims
	return []string{"users"}
}

// responseWriter wraps http.ResponseWriter to capture the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
