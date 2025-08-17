package metrics

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof" //nosec G108 - профилирование для отладки
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// EnhancedServer расширенный HTTP сервер для метрик и диагностики
type EnhancedServer struct {
	port              string
	path              string
	enhancedMetrics   *EnhancedMetrics
	healthChecker     *EnhancedHealthChecker
	diagnosticHandler interface{} // diagnostic.Handler interface
	logger            *zap.Logger
	server            *http.Server
	startTime         time.Time
	version           string
	buildTime         string
	buildType         string
	tunnelID          uuid.UUID
	connectorID       uuid.UUID
}

// Config конфигурация для EnhancedServer
type Config struct {
	Port              string
	Path              string
	EnhancedMetrics   *EnhancedMetrics
	HealthChecker     *EnhancedHealthChecker
	DiagnosticHandler interface{}
	Logger            *zap.Logger
	Version           string
	BuildTime         string
	BuildType         string
	TunnelID          uuid.UUID
	ConnectorID       uuid.UUID
	ShutdownTimeout   time.Duration
}

// NewEnhancedServer создает новый расширенный HTTP сервер
func NewEnhancedServer(config Config) *EnhancedServer {
	if config.ShutdownTimeout == 0 {
		config.ShutdownTimeout = 15 * time.Second
	}

	return &EnhancedServer{
		port:              config.Port,
		path:              config.Path,
		enhancedMetrics:   config.EnhancedMetrics,
		healthChecker:     config.HealthChecker,
		diagnosticHandler: config.DiagnosticHandler,
		logger:            config.Logger,
		startTime:         time.Now(),
		version:           config.Version,
		buildTime:         config.BuildTime,
		buildType:         config.BuildType,
		tunnelID:          config.TunnelID,
		connectorID:       config.ConnectorID,
	}
}

// newMetricsHandler создает HTTP handler с расширенными возможностями
func (s *EnhancedServer) newMetricsHandler() *http.ServeMux {
	router := http.NewServeMux()

	// Стандартные endpoints
	router.Handle("/debug/", http.DefaultServeMux)
	router.Handle(s.path, promhttp.Handler())

	// Health check endpoint
	router.HandleFunc("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		if s.enhancedMetrics != nil {
			s.enhancedMetrics.IncDiagnosticRequests("healthcheck", r.Method)
		}
		fmt.Fprintf(w, "OK\n")
	})

	// Ready endpoint (для Kubernetes readiness checks)
	if s.healthChecker != nil {
		router.Handle("/ready", s.healthChecker)
	}

	// Version endpoint
	router.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		if s.enhancedMetrics != nil {
			s.enhancedMetrics.IncDiagnosticRequests("version", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"version":"%s","build_time":"%s","build_type":"%s"}`,
			s.version, s.buildTime, s.buildType)
	})

	// Build info endpoint
	router.HandleFunc("/buildinfo", func(w http.ResponseWriter, r *http.Request) {
		if s.enhancedMetrics != nil {
			s.enhancedMetrics.IncDiagnosticRequests("buildinfo", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"goversion":"%s","type":"%s","revision":"%s","version":"%s"}`,
			"go1.23.0", s.buildType, s.buildTime, s.version)
	})

	// Tunnel info endpoint
	router.HandleFunc("/tunnel", func(w http.ResponseWriter, r *http.Request) {
		if s.enhancedMetrics != nil {
			s.enhancedMetrics.IncDiagnosticRequests("tunnel", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"tunnel_id":"%s","connector_id":"%s","status":"active"}`,
			s.tunnelID.String(), s.connectorID.String())
	})

	// Диагностические endpoints
	if s.diagnosticHandler != nil {
		if handler, ok := s.diagnosticHandler.(interface {
			InstallEndpoints(*http.ServeMux)
		}); ok {
			handler.InstallEndpoints(router)
		}
	}

	return router
}

// Start запускает расширенный HTTP сервер
func (s *EnhancedServer) Start(ctx context.Context) error {
	handler := s.newMetricsHandler()

	s.server = &http.Server{
		Addr:         s.port,
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	var wg sync.WaitGroup
	var err error

	wg.Add(1)
	go func() {
		defer wg.Done()
		s.logger.Info("Starting enhanced metrics server",
			zap.String("port", s.port),
			zap.String("path", s.path))

		err = s.server.ListenAndServe()
	}()

	// Небольшая задержка для запуска сервера
	time.Sleep(500 * time.Millisecond)

	s.logger.Info("Enhanced metrics server started",
		zap.String("address", fmt.Sprintf("http://%s%s", s.port, s.path)))

	// Ждем завершения контекста
	<-ctx.Done()

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if shutdownErr := s.server.Shutdown(shutdownCtx); shutdownErr != nil {
		s.logger.Error("Error during server shutdown", zap.Error(shutdownErr))
	}

	wg.Wait()

	if err == http.ErrServerClosed {
		s.logger.Info("Enhanced metrics server stopped gracefully")
		return nil
	}

	s.logger.Error("Enhanced metrics server failed", zap.Error(err))
	return err
}

// Stop останавливает сервер
func (s *EnhancedServer) Stop() error {
	if s.server != nil {
		return s.server.Close()
	}
	return nil
}

// GetAddress возвращает адрес сервера
func (s *EnhancedServer) GetAddress() string {
	return s.port
}

// GetMetricsPath возвращает путь к метрикам
func (s *EnhancedServer) GetMetricsPath() string {
	return s.path
}

// IsRunning проверяет, запущен ли сервер
func (s *EnhancedServer) IsRunning() bool {
	return s.server != nil
}

// GetUptime возвращает время работы сервера
func (s *EnhancedServer) GetUptime() time.Duration {
	return time.Since(s.startTime)
}

// GetVersion возвращает версию
func (s *EnhancedServer) GetVersion() string {
	return s.version
}

// GetBuildInfo возвращает информацию о сборке
func (s *EnhancedServer) GetBuildInfo() map[string]string {
	return map[string]string{
		"version":      s.version,
		"build_time":   s.buildTime,
		"build_type":   s.buildType,
		"tunnel_id":    s.tunnelID.String(),
		"connector_id": s.connectorID.String(),
	}
}
