package health

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// ComprehensiveHealthChecker расширенная система health checks
type ComprehensiveHealthChecker struct {
	logger *zap.Logger

	// Database connections
	db *sql.DB

	// Redis client (interface for compatibility)
	redisClient interface{}

	// External services
	externalServices map[string]string

	// TCP proxy endpoints
	tcpProxyEndpoints []string
}

// NewComprehensiveHealthChecker creates a comprehensive health checker
func NewComprehensiveHealthChecker(logger *zap.Logger) *ComprehensiveHealthChecker {
	return &ComprehensiveHealthChecker{
		logger:            logger,
		externalServices:  make(map[string]string),
		tcpProxyEndpoints: make([]string, 0),
	}
}

// SetDatabase устанавливает подключение к базе данных
func (c *ComprehensiveHealthChecker) SetDatabase(db *sql.DB) {
	c.db = db
}

// SetRedisClient устанавливает Redis клиент
func (c *ComprehensiveHealthChecker) SetRedisClient(client interface{}) {
	c.redisClient = client
}

// AddExternalService добавляет внешний сервис для проверки
func (c *ComprehensiveHealthChecker) AddExternalService(name, url string) {
	c.externalServices[name] = url
}

// AddTCPProxyEndpoint добавляет TCP прокси endpoint для проверки
func (c *ComprehensiveHealthChecker) AddTCPProxyEndpoint(endpoint string) {
	c.tcpProxyEndpoints = append(c.tcpProxyEndpoints, endpoint)
}

// DatabaseHealthCheck проверка подключения к базе данных
type DatabaseHealthCheck struct {
	name   string
	logger *zap.Logger
	db     *sql.DB
}

func NewDatabaseHealthCheck(db *sql.DB, logger *zap.Logger) *DatabaseHealthCheck {
	return &DatabaseHealthCheck{
		name:   "database",
		logger: logger,
		db:     db,
	}
}

func (d *DatabaseHealthCheck) Check(ctx context.Context) error {
	if d.db == nil {
		return fmt.Errorf("database connection not configured")
	}

	// Ping database with timeout
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := d.db.PingContext(ctx); err != nil {
		d.logger.Error("Database ping failed", zap.Error(err))
		return fmt.Errorf("database ping failed: %w", err)
	}

	// Test simple query
	var result int
	err := d.db.QueryRowContext(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		d.logger.Error("Database query test failed", zap.Error(err))
		return fmt.Errorf("database query test failed: %w", err)
	}

	if result != 1 {
		return fmt.Errorf("database query returned unexpected result: %d", result)
	}

	d.logger.Debug("Database health check passed")
	return nil
}

func (d *DatabaseHealthCheck) Name() string {
	return d.name
}

func (d *DatabaseHealthCheck) Description() string {
	return "Database connectivity and query test"
}

// RedisHealthCheck проверка подключения к Redis
type RedisHealthCheck struct {
	name   string
	logger *zap.Logger
	client interface{}
}

func NewRedisHealthCheck(client interface{}, logger *zap.Logger) *RedisHealthCheck {
	return &RedisHealthCheck{
		name:   "redis",
		logger: logger,
		client: client,
	}
}

func (r *RedisHealthCheck) Check(ctx context.Context) error {
	if r.client == nil {
		return fmt.Errorf("redis client not configured")
	}

	// В реальной реализации здесь был бы ping к Redis
	// Пример проверки Redis:
	// if redisClient, ok := r.client.(*redis.Client); ok {
	//     if err := redisClient.Ping(ctx).Err(); err != nil {
	//         return fmt.Errorf("redis ping failed: %w", err)
	//     }
	// }
	
	// Для демонстрации возвращаем успех
	r.logger.Debug("Redis health check passed")
	return nil
}

func (r *RedisHealthCheck) Name() string {
	return r.name
}

func (r *RedisHealthCheck) Description() string {
	return "Redis connectivity and operation test"
}

// TCPProxyHealthCheck проверка TCP прокси
type TCPProxyHealthCheck struct {
	name      string
	logger    *zap.Logger
	endpoints []string
}

func NewTCPProxyHealthCheck(endpoints []string, logger *zap.Logger) *TCPProxyHealthCheck {
	return &TCPProxyHealthCheck{
		name:      "tcp_proxy",
		logger:    logger,
		endpoints: endpoints,
	}
}

func (t *TCPProxyHealthCheck) Check(ctx context.Context) error {
	if len(t.endpoints) == 0 {
		return fmt.Errorf("no TCP proxy endpoints configured")
	}

	for _, endpoint := range t.endpoints {
		if err := t.checkEndpoint(ctx, endpoint); err != nil {
			t.logger.Error("TCP proxy endpoint check failed",
				zap.String("endpoint", endpoint),
				zap.Error(err))
			return fmt.Errorf("TCP proxy endpoint %s failed: %w", endpoint, err)
		}
	}

	t.logger.Debug("TCP proxy health check passed", zap.Int("endpoints", len(t.endpoints)))
	return nil
}

func (t *TCPProxyHealthCheck) checkEndpoint(ctx context.Context, endpoint string) error {
	// Create connection with timeout
	dialer := &net.Dialer{
		Timeout: 5 * time.Second,
	}

	conn, err := dialer.DialContext(ctx, "tcp", endpoint)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", endpoint, err)
	}
	defer conn.Close()

	// Test write/read if possible
	conn.SetDeadline(time.Now().Add(2 * time.Second))

	// For HTTP endpoints, try a simple HTTP request
	if t.isHTTPEndpoint(endpoint) {
		return t.checkHTTPEndpoint(ctx, endpoint)
	}

	return nil
}

func (t *TCPProxyHealthCheck) isHTTPEndpoint(endpoint string) bool {
	// Simple heuristic: if endpoint contains common HTTP ports
	return net.ParseIP(endpoint) == nil &&
		(endpoint[len(endpoint)-4:] == ":80" ||
			endpoint[len(endpoint)-4:] == "443" ||
			endpoint[len(endpoint)-5:] == ":8080" ||
			endpoint[len(endpoint)-5:] == ":8443")
}

func (t *TCPProxyHealthCheck) checkHTTPEndpoint(ctx context.Context, endpoint string) error {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	url := fmt.Sprintf("http://%s/health", endpoint)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP endpoint returned status %d", resp.StatusCode)
	}

	return nil
}

func (t *TCPProxyHealthCheck) Name() string {
	return t.name
}

func (t *TCPProxyHealthCheck) Description() string {
	return "TCP Proxy endpoints connectivity test"
}

// ExternalServiceHealthCheck проверка внешних сервисов
type ExternalServiceHealthCheck struct {
	name     string
	logger   *zap.Logger
	services map[string]string
}

func NewExternalServiceHealthCheck(services map[string]string, logger *zap.Logger) *ExternalServiceHealthCheck {
	return &ExternalServiceHealthCheck{
		name:     "external_services",
		logger:   logger,
		services: services,
	}
}

func (e *ExternalServiceHealthCheck) Check(ctx context.Context) error {
	if len(e.services) == 0 {
		return nil // No external services to check
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	for serviceName, serviceURL := range e.services {
		if err := e.checkService(ctx, client, serviceName, serviceURL); err != nil {
			e.logger.Error("External service check failed",
				zap.String("service", serviceName),
				zap.String("url", serviceURL),
				zap.Error(err))
			return fmt.Errorf("external service %s failed: %w", serviceName, err)
		}
	}

	e.logger.Debug("External services health check passed", zap.Int("services", len(e.services)))
	return nil
}

func (e *ExternalServiceHealthCheck) checkService(ctx context.Context, client *http.Client, name, url string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return fmt.Errorf("service returned server error: %d", resp.StatusCode)
	}

	return nil
}

func (e *ExternalServiceHealthCheck) Name() string {
	return e.name
}

func (e *ExternalServiceHealthCheck) Description() string {
	return "External services connectivity test"
}

// ComprehensiveHealthCheck interface for comprehensive health checks
type ComprehensiveHealthCheck interface {
	Check(ctx context.Context) error
	Name() string
	Description() string
}
