package ingress

import (
	"crypto/tls"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"
)

// IngressConfig представляет конфигурацию ingress системы
type IngressConfig struct {
	// Основные настройки
	Enabled bool `json:"enabled" yaml:"enabled"`
	
	// Bastion mode для безопасного доступа
	BastionMode bool `json:"bastion_mode" yaml:"bastion_mode"`
	
	// IP-based access control
	IPAccessControl IPAccessConfig `json:"ip_access_control" yaml:"ip_access_control"`
	
	// TLS настройки
	TLS TLSConfig `json:"tls" yaml:"tls"`
	
	// Connection pooling
	ConnectionPool ConnectionPoolConfig `json:"connection_pool" yaml:"connection_pool"`
	
	// Правила маршрутизации
	Rules []IngressRule `json:"rules" yaml:"rules"`
	
	// Глобальные настройки для origin requests
	OriginRequest OriginRequestConfig `json:"origin_request" yaml:"origin_request"`
}

// IPAccessConfig представляет конфигурацию IP-based access control
type IPAccessConfig struct {
	Enabled      bool        `json:"enabled" yaml:"enabled"`
	DefaultAllow bool        `json:"default_allow" yaml:"default_allow"`
	Rules        []IPRule    `json:"rules" yaml:"rules"`
}

// IPRule представляет правило IP access control
type IPRule struct {
	Prefix string `json:"prefix" yaml:"prefix"` // CIDR notation (e.g., "192.168.1.0/24")
	Ports  []int  `json:"ports" yaml:"ports"`   // Список портов (пустой = все порты)
	Allow  bool   `json:"allow" yaml:"allow"`   // true = разрешить, false = запретить
}

// TLSConfig представляет конфигурацию TLS
type TLSConfig struct {
	Enabled            bool     `json:"enabled" yaml:"enabled"`
	CertFile           string   `json:"cert_file" yaml:"cert_file"`
	KeyFile            string   `json:"key_file" yaml:"key_file"`
	MinVersion         string   `json:"min_version" yaml:"min_version"`
	CipherSuites       []string `json:"cipher_suites" yaml:"cipher_suites"`
	VerifyClientCert   bool     `json:"verify_client_cert" yaml:"verify_client_cert"`
	ClientCAFile       string   `json:"client_ca_file" yaml:"client_ca_file"`
	InsecureSkipVerify bool     `json:"insecure_skip_verify" yaml:"insecure_skip_verify"`
}

// ConnectionPoolConfig представляет конфигурацию connection pooling
type ConnectionPoolConfig struct {
	MaxSize     int           `json:"max_size" yaml:"max_size"`
	IdleTimeout time.Duration `json:"idle_timeout" yaml:"idle_timeout"`
	MaxIdle     int           `json:"max_idle" yaml:"max_idle"`
	MaxLifetime time.Duration `json:"max_lifetime" yaml:"max_lifetime"`
}

// IngressRule представляет правило маршрутизации
type IngressRule struct {
	// Hostname для сопоставления (пустой = все hostnames)
	Hostname string `json:"hostname" yaml:"hostname"`
	
	// Path regex для сопоставления (пустой = все paths)
	Path string `json:"path" yaml:"path"`
	
	// Сервис для проксирования
	Service string `json:"service" yaml:"service"`
	
	// Конфигурация для origin request (переопределяет глобальную)
	OriginRequest *OriginRequestConfig `json:"origin_request,omitempty" yaml:"origin_request,omitempty"`
	
	// Скомпилированный regex для path
	pathRegex *regexp.Regexp
}

// OriginRequestConfig представляет конфигурацию для запросов к origin
type OriginRequestConfig struct {
	// Таймауты
	ConnectTimeout time.Duration `json:"connect_timeout" yaml:"connect_timeout"`
	TLSTimeout     time.Duration `json:"tls_timeout" yaml:"tls_timeout"`
	
	// Keep-alive настройки
	TCPKeepAlive         time.Duration `json:"tcp_keep_alive" yaml:"tcp_keep_alive"`
	KeepAliveConnections int           `json:"keep_alive_connections" yaml:"keep_alive_connections"`
	KeepAliveTimeout     time.Duration `json:"keep_alive_timeout" yaml:"keep_alive_timeout"`
	
	// HTTP настройки
	HTTPHostHeader   string `json:"http_host_header" yaml:"http_host_header"`
	OriginServerName string `json:"origin_server_name" yaml:"origin_server_name"`
	
	// TLS настройки
	NoTLSVerify           bool   `json:"no_tls_verify" yaml:"no_tls_verify"`
	CAPool                string `json:"ca_pool" yaml:"ca_pool"`
	DisableChunkedEncoding bool   `json:"disable_chunked_encoding" yaml:"disable_chunked_encoding"`
	
	// Bastion mode
	BastionMode bool `json:"bastion_mode" yaml:"bastion_mode"`
	
	// Proxy настройки
	ProxyAddress string `json:"proxy_address" yaml:"proxy_address"`
	ProxyPort    uint   `json:"proxy_port" yaml:"proxy_port"`
	ProxyType    string `json:"proxy_type" yaml:"proxy_type"`
	
	// HTTP/2 настройки
	HTTP2Origin bool `json:"http2_origin" yaml:"http2_origin"`
	
	// IP правила для этого origin
	IPRules []IPRule `json:"ip_rules" yaml:"ip_rules"`
}

// DefaultIngressConfig возвращает конфигурацию по умолчанию
func DefaultIngressConfig() *IngressConfig {
	return &IngressConfig{
		Enabled:     true,
		BastionMode: false,
		IPAccessControl: IPAccessConfig{
			Enabled:      false,
			DefaultAllow: true,
			Rules:        []IPRule{},
		},
		TLS: TLSConfig{
			Enabled:            true,
			MinVersion:         "1.2",
			CipherSuites:       []string{},
			VerifyClientCert:   false,
			InsecureSkipVerify: false,
		},
		ConnectionPool: ConnectionPoolConfig{
			MaxSize:     100,
			IdleTimeout: 90 * time.Second,
			MaxIdle:     10,
			MaxLifetime: 1 * time.Hour,
		},
		Rules: []IngressRule{},
		OriginRequest: OriginRequestConfig{
			ConnectTimeout:       30 * time.Second,
			TLSTimeout:          10 * time.Second,
			TCPKeepAlive:        30 * time.Second,
			KeepAliveConnections: 100,
			KeepAliveTimeout:     90 * time.Second,
			NoTLSVerify:          false,
			DisableChunkedEncoding: false,
			BastionMode:          false,
			ProxyAddress:         "127.0.0.1",
			HTTP2Origin:          true,
		},
	}
}

// Validate проверяет корректность конфигурации
func (ic *IngressConfig) Validate() error {
	// Проверка TLS конфигурации
	if ic.TLS.Enabled {
		if ic.TLS.CertFile == "" || ic.TLS.KeyFile == "" {
			return fmt.Errorf("TLS enabled but cert_file or key_file not specified")
		}
	}
	
	// Проверка IP правил
	for i, rule := range ic.IPAccessControl.Rules {
		if err := validateIPRule(rule); err != nil {
			return fmt.Errorf("invalid IP rule %d: %v", i, err)
		}
	}
	
	// Проверка правил маршрутизации
	for i, rule := range ic.Rules {
		if err := validateIngressRule(rule, i, len(ic.Rules)); err != nil {
			return fmt.Errorf("invalid ingress rule %d: %v", i, err)
		}
	}
	
	// Проверка connection pool
	if ic.ConnectionPool.MaxSize <= 0 {
		return fmt.Errorf("connection pool max_size must be positive")
	}
	
	return nil
}

// validateIPRule проверяет корректность IP правила
func validateIPRule(rule IPRule) error {
	if rule.Prefix == "" {
		return fmt.Errorf("IP rule prefix cannot be empty")
	}
	
	_, _, err := net.ParseCIDR(rule.Prefix)
	if err != nil {
		return fmt.Errorf("invalid CIDR prefix: %s", rule.Prefix)
	}
	
	for _, port := range rule.Ports {
		if port < 1 || port > 65535 {
			return fmt.Errorf("invalid port %d, must be between 1 and 65535", port)
		}
	}
	
	return nil
}

// validateIngressRule проверяет корректность правила маршрутизации
func validateIngressRule(rule IngressRule, index, totalRules int) error {
	// Проверка hostname
	if rule.Hostname != "" {
		if strings.Contains(rule.Hostname, ":") {
			return fmt.Errorf("hostname cannot contain port")
		}
		
		// Проверка wildcard
		if strings.Count(rule.Hostname, "*") > 1 {
			return fmt.Errorf("hostname can have at most one wildcard character")
		}
		
		if strings.Contains(rule.Hostname, "*") && !strings.HasPrefix(rule.Hostname, "*.") {
			return fmt.Errorf("wildcard can only be used for subdomains (e.g., *.example.com)")
		}
	}
	
	// Проверка path regex
	if rule.Path != "" {
		regex, err := regexp.Compile(rule.Path)
		if err != nil {
			return fmt.Errorf("invalid path regex: %v", err)
		}
		rule.pathRegex = regex
	}
	
	// Проверка service
	if rule.Service == "" {
		return fmt.Errorf("service cannot be empty")
	}
	
	// Проверка catch-all правила (должно быть последним)
	isCatchAll := (rule.Hostname == "" || rule.Hostname == "*") && rule.Path == ""
	isLastRule := index == totalRules-1
	
	if isLastRule && !isCatchAll {
		return fmt.Errorf("last rule must be a catch-all rule (no hostname or path filter)")
	}
	
	if !isLastRule && isCatchAll {
		return fmt.Errorf("catch-all rule can only be the last rule")
	}
	
	return nil
}

// GetTLSConfig возвращает TLS конфигурацию
func (ic *IngressConfig) GetTLSConfig() (*tls.Config, error) {
	if !ic.TLS.Enabled {
		return nil, nil
	}
	
	config := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}
	
	// Загрузка сертификата
	if ic.TLS.CertFile != "" && ic.TLS.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(ic.TLS.CertFile, ic.TLS.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS certificate: %v", err)
		}
		config.Certificates = []tls.Certificate{cert}
	}
	
	// Настройка минимальной версии TLS
	switch ic.TLS.MinVersion {
	case "1.2":
		config.MinVersion = tls.VersionTLS12
	case "1.3":
		config.MinVersion = tls.VersionTLS13
	default:
		config.MinVersion = tls.VersionTLS12
	}
	
	// Настройка cipher suites
	if len(ic.TLS.CipherSuites) > 0 {
		config.CipherSuites = getCipherSuites(ic.TLS.CipherSuites)
	}
	
	// Настройка client certificate verification
	if ic.TLS.VerifyClientCert {
		config.ClientAuth = tls.RequireAndVerifyClientCert
		if ic.TLS.ClientCAFile != "" {
			// Загрузка CA сертификата для клиентов
			// TODO: Реализовать загрузку CA
		}
	}
	
	// InsecureSkipVerify (только для разработки)
	if ic.TLS.InsecureSkipVerify {
		config.InsecureSkipVerify = true
	}
	
	return config, nil
}

// getCipherSuites возвращает список cipher suites
func getCipherSuites(names []string) []uint16 {
	var suites []uint16
	
	cipherMap := map[string]uint16{
		"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256":   tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384":   tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256": tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384": tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		"TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305":    tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		"TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305":  tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
	}
	
	for _, name := range names {
		if suite, exists := cipherMap[name]; exists {
			suites = append(suites, suite)
		}
	}
	
	return suites
}

// FindMatchingRule находит правило, соответствующее hostname и path
func (ic *IngressConfig) FindMatchingRule(hostname, path string) (*IngressRule, int) {
	// Убираем порт из hostname если есть
	host, _, err := net.SplitHostPort(hostname)
	if err == nil {
		hostname = host
	}
	
	for i, rule := range ic.Rules {
		if rule.Matches(hostname, path) {
			return &rule, i
		}
	}
	
	// Возвращаем последнее правило (catch-all)
	if len(ic.Rules) > 0 {
		lastRule := ic.Rules[len(ic.Rules)-1]
		return &lastRule, len(ic.Rules) - 1
	}
	
	return nil, -1
}

// Matches проверяет, соответствует ли правило hostname и path
func (ir *IngressRule) Matches(hostname, path string) bool {
	// Проверка hostname
	hostMatch := false
	if ir.Hostname == "" || ir.Hostname == "*" {
		hostMatch = true
	} else {
		hostMatch = matchHost(ir.Hostname, hostname)
	}
	
	// Проверка path
	pathMatch := true
	if ir.Path != "" && ir.pathRegex != nil {
		pathMatch = ir.pathRegex.MatchString(path)
	}
	
	return hostMatch && pathMatch
}

// matchHost проверяет соответствие hostname с поддержкой wildcard
func matchHost(ruleHost, reqHost string) bool {
	if ruleHost == reqHost {
		return true
	}
	
	// Проверка wildcard в начале
	if strings.HasPrefix(ruleHost, "*.") {
		toMatch := strings.TrimPrefix(ruleHost, "*")
		return strings.HasSuffix(reqHost, toMatch)
	}
	
	return false
}

// CompileRules компилирует regex для всех правил
func (ic *IngressConfig) CompileRules() error {
	for i := range ic.Rules {
		if ic.Rules[i].Path != "" {
			regex, err := regexp.Compile(ic.Rules[i].Path)
			if err != nil {
				return fmt.Errorf("invalid path regex in rule %d: %v", i, err)
			}
			ic.Rules[i].pathRegex = regex
		}
	}
	return nil
}

// IsBastionMode проверяет, включен ли bastion mode
func (ic *IngressConfig) IsBastionMode() bool {
	return ic.BastionMode
}

// GetConnectionPoolConfig возвращает конфигурацию connection pool
func (ic *IngressConfig) GetConnectionPoolConfig() ConnectionPoolConfig {
	return ic.ConnectionPool
}

// GetIPAccessConfig возвращает конфигурацию IP access control
func (ic *IngressConfig) GetIPAccessConfig() IPAccessConfig {
	return ic.IPAccessControl
} 