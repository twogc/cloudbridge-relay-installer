package ingress

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
	"github.com/twogc/cloudbridge-relay-installer/internal/metrics"
	"go.uber.org/zap"
)

// IngressRuleEngine управляет правилами маршрутизации
type IngressRuleEngine struct {
	config  *IngressConfig
	rules   []*IngressRule
	mu      sync.RWMutex
	logger  *logger.Logger
	metrics *metrics.Metrics
	
	// IP access policy
	ipPolicy *IPAccessPolicy
	
	// Connection pool
	connectionPool *ConnectionPool
}

// NewIngressRuleEngine создает новый движок правил маршрутизации
func NewIngressRuleEngine(config *IngressConfig, logger *logger.Logger, metrics *metrics.Metrics) (*IngressRuleEngine, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid ingress config: %v", err)
	}
	
	if err := config.CompileRules(); err != nil {
		return nil, fmt.Errorf("failed to compile rules: %v", err)
	}
	
	// Создаем IP access policy
	var ipPolicy *IPAccessPolicy
	if config.IPAccessControl.Enabled {
		var err error
		ipPolicy, err = NewIPAccessPolicy(config.IPAccessControl, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create IP access policy: %v", err)
		}
	}
	
	// Создаем connection pool
	connectionPool := NewConnectionPool(config.ConnectionPool, logger, metrics)
	
	engine := &IngressRuleEngine{
		config:         config,
		rules:          make([]*IngressRule, len(config.Rules)),
		logger:         logger,
		metrics:        metrics,
		ipPolicy:       ipPolicy,
		connectionPool: connectionPool,
	}
	
	// Копируем правила
	for i, rule := range config.Rules {
		engine.rules[i] = &rule
	}
	
	logger.Info("Ingress rule engine created",
		zap.Int("rules_count", len(engine.rules)),
		zap.Bool("ip_access_enabled", config.IPAccessControl.Enabled),
		zap.Bool("bastion_mode", config.BastionMode),
	)
	
	return engine, nil
}

// FindMatchingRule находит правило, соответствующее запросу
func (e *IngressRuleEngine) FindMatchingRule(hostname, path string) (*IngressRule, int) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	// Убираем порт из hostname если есть
	host, _, err := net.SplitHostPort(hostname)
	if err == nil {
		hostname = host
	}
	
	for i, rule := range e.rules {
		if rule.Matches(hostname, path) {
			return rule, i
		}
	}
	
	// Возвращаем последнее правило (catch-all)
	if len(e.rules) > 0 {
		return e.rules[len(e.rules)-1], len(e.rules) - 1
	}
	
	return nil, -1
}

// HandleRequest обрабатывает HTTP запрос
func (e *IngressRuleEngine) HandleRequest(w http.ResponseWriter, r *http.Request) error {
	startTime := time.Now()
	
	// Проверяем IP access control
	if e.ipPolicy != nil {
		clientIP := getClientIP(r)
		port := 80 // По умолчанию HTTP
		if r.TLS != nil {
			port = 443 // HTTPS
		}
		
		allowed, rule := e.ipPolicy.Allowed(clientIP, port)
		if !allowed {
			e.logger.Warn("Access denied by IP policy",
				zap.String("client_ip", clientIP.String()),
				zap.Int("port", port),
				zap.String("rule", rule.String()),
			)
			
			if e.metrics != nil {
				e.metrics.IncErrors("access_denied")
			}
			
			http.Error(w, "Access denied", http.StatusForbidden)
			return fmt.Errorf("access denied by IP policy")
		}
	}
	
	// Находим подходящее правило
	rule, ruleIndex := e.FindMatchingRule(r.Host, r.URL.Path)
	if rule == nil {
		e.logger.Warn("No matching rule found",
			zap.String("host", r.Host),
			zap.String("path", r.URL.Path),
		)
		
		if e.metrics != nil {
			e.metrics.IncErrors("no_matching_rule")
		}
		
		http.Error(w, "No matching rule", http.StatusNotFound)
		return fmt.Errorf("no matching rule found")
	}
	
	// Получаем конфигурацию origin request
	originConfig := e.getOriginRequestConfig(rule, ruleIndex)
	
	// Обрабатываем запрос в зависимости от типа сервиса
	err := e.handleServiceRequest(w, r, rule, originConfig)
	
	// Обновляем метрики
	duration := time.Since(startTime)
	if e.metrics != nil {
		e.metrics.ObserveMessageLatency("ingress_request", duration)
		if err != nil {
			e.metrics.IncErrors("ingress_error")
		} else {
			e.metrics.IncConnections()
		}
	}
	
	if err != nil {
		e.logger.Error("Failed to handle request",
			zap.String("host", r.Host),
			zap.String("path", r.URL.Path),
			zap.String("service", rule.Service),
			zap.Error(err),
		)
		
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
	
	return err
}

// handleServiceRequest обрабатывает запрос к сервису
func (e *IngressRuleEngine) handleServiceRequest(w http.ResponseWriter, r *http.Request, rule *IngressRule, config OriginRequestConfig) error {
	serviceURL, err := url.Parse(rule.Service)
	if err != nil {
		return fmt.Errorf("invalid service URL: %v", err)
	}
	
	// Определяем тип сервиса
	switch serviceURL.Scheme {
	case "http", "https":
		return e.handleHTTPService(w, r, serviceURL, config)
	case "tcp":
		return e.handleTCPService(w, r, serviceURL, config)
	case "ssh":
		return e.handleSSHService(w, r, serviceURL, config)
	case "rdp":
		return e.handleRDPService(w, r, serviceURL, config)
	case "bastion":
		return e.handleBastionService(w, r, config)
	default:
		return fmt.Errorf("unsupported service scheme: %s", serviceURL.Scheme)
	}
}

// handleHTTPService обрабатывает HTTP/HTTPS сервисы
func (e *IngressRuleEngine) handleHTTPService(w http.ResponseWriter, r *http.Request, serviceURL *url.URL, config OriginRequestConfig) error {
	// Создаем HTTP клиент с настройками
	client := &http.Client{
		Timeout: config.ConnectTimeout,
		Transport: &http.Transport{
			TLSHandshakeTimeout: config.TLSTimeout,
			MaxIdleConns:        config.KeepAliveConnections,
			IdleConnTimeout:     config.KeepAliveTimeout,
			DisableCompression:  config.DisableChunkedEncoding,
		},
	}
	
	// Создаем новый запрос
	targetURL := *serviceURL
	targetURL.Path = r.URL.Path
	targetURL.RawQuery = r.URL.RawQuery
	
	proxyReq, err := http.NewRequest(r.Method, targetURL.String(), r.Body)
	if err != nil {
		return fmt.Errorf("failed to create proxy request: %v", err)
	}
	
	// Копируем заголовки
	for name, values := range r.Header {
		for _, value := range values {
			proxyReq.Header.Add(name, value)
		}
	}
	
	// Устанавливаем Host header если указан
	if config.HTTPHostHeader != "" {
		proxyReq.Host = config.HTTPHostHeader
	}
	
	// Выполняем запрос
	resp, err := client.Do(proxyReq)
	if err != nil {
		return fmt.Errorf("failed to proxy request: %v", err)
	}
	defer resp.Body.Close()
	
	// Копируем заголовки ответа
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}
	
	w.WriteHeader(resp.StatusCode)
	
	// Копируем тело ответа
	_, err = copyResponseBody(w, resp.Body)
	return err
}

// handleTCPService обрабатывает TCP сервисы
func (e *IngressRuleEngine) handleTCPService(w http.ResponseWriter, r *http.Request, serviceURL *url.URL, config OriginRequestConfig) error {
	// Для TCP сервисов используем WebSocket upgrade
	if !isWebSocketRequest(r) {
		return fmt.Errorf("TCP service requires WebSocket connection")
	}
	
	// Получаем соединение из connection pool
	conn, err := e.connectionPool.GetConnection(serviceURL.Host)
	if err != nil {
		return fmt.Errorf("failed to get connection: %v", err)
	}
	defer e.connectionPool.ReturnConnection(conn)
	
	// Выполняем WebSocket upgrade
	return e.performWebSocketUpgrade(w, r, conn)
}

// handleSSHService обрабатывает SSH сервисы
func (e *IngressRuleEngine) handleSSHService(w http.ResponseWriter, r *http.Request, serviceURL *url.URL, config OriginRequestConfig) error {
	// SSH сервисы также используют WebSocket
	if !isWebSocketRequest(r) {
		return fmt.Errorf("SSH service requires WebSocket connection")
	}
	
	// Устанавливаем порт SSH по умолчанию если не указан
	if serviceURL.Port() == "" {
		serviceURL.Host = serviceURL.Host + ":22"
	}
	
	conn, err := e.connectionPool.GetConnection(serviceURL.Host)
	if err != nil {
		return fmt.Errorf("failed to get SSH connection: %v", err)
	}
	defer e.connectionPool.ReturnConnection(conn)
	
	return e.performWebSocketUpgrade(w, r, conn)
}

// handleRDPService обрабатывает RDP сервисы
func (e *IngressRuleEngine) handleRDPService(w http.ResponseWriter, r *http.Request, serviceURL *url.URL, config OriginRequestConfig) error {
	// RDP сервисы также используют WebSocket
	if !isWebSocketRequest(r) {
		return fmt.Errorf("RDP service requires WebSocket connection")
	}
	
	// Устанавливаем порт RDP по умолчанию если не указан
	if serviceURL.Port() == "" {
		serviceURL.Host = serviceURL.Host + ":3389"
	}
	
	conn, err := e.connectionPool.GetConnection(serviceURL.Host)
	if err != nil {
		return fmt.Errorf("failed to get RDP connection: %v", err)
	}
	defer e.connectionPool.ReturnConnection(conn)
	
	return e.performWebSocketUpgrade(w, r, conn)
}

// handleBastionService обрабатывает bastion сервисы
func (e *IngressRuleEngine) handleBastionService(w http.ResponseWriter, r *http.Request, config OriginRequestConfig) error {
	// Bastion mode предоставляет безопасный доступ к внутренним сервисам
	if !isWebSocketRequest(r) {
		return fmt.Errorf("Bastion service requires WebSocket connection")
	}
	
	e.logger.Info("Bastion mode request",
		zap.String("client_ip", getClientIP(r).String()),
		zap.String("user_agent", r.UserAgent()),
	)
	
	// В bastion mode мы предоставляем WebSocket proxy для различных протоколов
	return e.handleBastionWebSocket(w, r, config)
}

// getOriginRequestConfig возвращает конфигурацию origin request для правила
func (e *IngressRuleEngine) getOriginRequestConfig(rule *IngressRule, ruleIndex int) OriginRequestConfig {
	// Если у правила есть собственная конфигурация, используем её
	if rule.OriginRequest != nil {
		return *rule.OriginRequest
	}
	
	// Иначе используем глобальную конфигурацию
	return e.config.OriginRequest
}

// getClientIP извлекает реальный IP клиента
func getClientIP(r *http.Request) net.IP {
	// Проверяем заголовки прокси
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		if parsed := net.ParseIP(strings.Split(ip, ",")[0]); parsed != nil {
			return parsed
		}
	}
	
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		if parsed := net.ParseIP(ip); parsed != nil {
			return parsed
		}
	}
	
	// Используем RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	
	if ip := net.ParseIP(host); ip != nil {
		return ip
	}
	
	// Fallback
	return net.IPv4zero
}

// isWebSocketRequest проверяет, является ли запрос WebSocket
func isWebSocketRequest(r *http.Request) bool {
	return r.Header.Get("Upgrade") == "websocket" &&
		strings.Contains(strings.ToLower(r.Header.Get("Connection")), "upgrade")
}

// performWebSocketUpgrade выполняет WebSocket upgrade
func (e *IngressRuleEngine) performWebSocketUpgrade(w http.ResponseWriter, r *http.Request, conn net.Conn) error {
	// TODO: Реализовать WebSocket upgrade
	// Это требует дополнительной реализации с использованием gorilla/websocket
	return fmt.Errorf("WebSocket upgrade not implemented yet")
}

// handleBastionWebSocket обрабатывает WebSocket в bastion mode
func (e *IngressRuleEngine) handleBastionWebSocket(w http.ResponseWriter, r *http.Request, config OriginRequestConfig) error {
	// TODO: Реализовать bastion WebSocket handler
	return fmt.Errorf("Bastion WebSocket handler not implemented yet")
}

// copyResponseBody копирует тело ответа
func copyResponseBody(w http.ResponseWriter, body io.ReadCloser) (int64, error) {
	return io.Copy(w, body)
}

// GetStats возвращает статистику движка правил
func (e *IngressRuleEngine) GetStats() map[string]interface{} {
	stats := map[string]interface{}{
		"rules_count": len(e.rules),
		"ip_policy_enabled": e.ipPolicy != nil,
		"bastion_mode": e.config.BastionMode,
	}
	
	if e.ipPolicy != nil {
		stats["ip_policy_rules"] = e.ipPolicy.GetRuleCount()
		stats["ip_policy_default_allow"] = e.ipPolicy.IsDefaultAllow()
	}
	
	if e.connectionPool != nil {
		poolStats := e.connectionPool.GetStats()
		stats["connection_pool"] = poolStats
	}
	
	return stats
}

// AddRule добавляет новое правило
func (e *IngressRuleEngine) AddRule(rule IngressRule) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	// Валидируем правило
	if err := validateIngressRule(rule, len(e.rules), len(e.rules)+1); err != nil {
		return err
	}
	
	// Компилируем regex если есть
	if rule.Path != "" {
		regex, err := regexp.Compile(rule.Path)
		if err != nil {
			return fmt.Errorf("invalid path regex: %v", err)
		}
		rule.pathRegex = regex
	}
	
	e.rules = append(e.rules, &rule)
	
	e.logger.Info("Ingress rule added",
		zap.String("hostname", rule.Hostname),
		zap.String("path", rule.Path),
		zap.String("service", rule.Service),
	)
	
	return nil
}

// RemoveRule удаляет правило по индексу
func (e *IngressRuleEngine) RemoveRule(index int) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	if index < 0 || index >= len(e.rules) {
		return fmt.Errorf("invalid rule index: %d", index)
	}
	
	rule := e.rules[index]
	e.rules = append(e.rules[:index], e.rules[index+1:]...)
	
	e.logger.Info("Ingress rule removed",
		zap.String("hostname", rule.Hostname),
		zap.String("path", rule.Path),
		zap.String("service", rule.Service),
	)
	
	return nil
}

// GetRules возвращает копию всех правил
func (e *IngressRuleEngine) GetRules() []*IngressRule {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	rules := make([]*IngressRule, len(e.rules))
	copy(rules, e.rules)
	
	return rules
}

// Shutdown корректно завершает работу движка
func (e *IngressRuleEngine) Shutdown(ctx context.Context) error {
	e.logger.Info("Shutting down ingress rule engine")
	
	if e.connectionPool != nil {
		if err := e.connectionPool.Shutdown(ctx); err != nil {
			e.logger.Error("Failed to shutdown connection pool", zap.Error(err))
		}
	}
	
	return nil
}

// Stop останавливает rule engine (алиас для Shutdown)
func (e *IngressRuleEngine) Stop(ctx context.Context) error {
	return e.Shutdown(ctx)
}

// processRequest обрабатывает запрос согласно правилу
func (e *IngressRuleEngine) processRequest(w http.ResponseWriter, r *http.Request, rule *IngressRule) error {
	// Получаем конфигурацию origin request
	config := e.getOriginRequestConfig(rule, 0)
	
	// Обрабатываем запрос в зависимости от типа сервиса
	return e.handleServiceRequest(w, r, rule, config)
} 