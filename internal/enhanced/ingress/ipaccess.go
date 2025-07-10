package ingress

import (
	"fmt"
	"net"
	"sort"
	"sync"

	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
	"go.uber.org/zap"
)

// IPAccessPolicy представляет политику IP access control
type IPAccessPolicy struct {
	defaultAllow bool
	rules        []IPAccessRule
	mu           sync.RWMutex
	logger       *logger.Logger
}

// IPAccessRule представляет правило IP access control
type IPAccessRule struct {
	ipNet *net.IPNet
	ports []int
	allow bool
}

// NewIPAccessPolicy создает новую политику IP access control
func NewIPAccessPolicy(config IPAccessConfig, logger *logger.Logger) (*IPAccessPolicy, error) {
	rules := make([]IPAccessRule, len(config.Rules))
	
	for i, rule := range config.Rules {
		ipRule, err := NewIPAccessRule(rule.Prefix, rule.Ports, rule.Allow)
		if err != nil {
			return nil, fmt.Errorf("failed to create IP rule %d: %v", i, err)
		}
		rules[i] = ipRule
	}
	
	policy := &IPAccessPolicy{
		defaultAllow: config.DefaultAllow,
		rules:        rules,
		logger:       logger,
	}
	
	logger.Info("IP access policy created",
		zap.Bool("default_allow", config.DefaultAllow),
		zap.Int("rules_count", len(rules)),
	)
	
	return policy, nil
}

// NewIPAccessRule создает новое правило IP access control
func NewIPAccessRule(prefix string, ports []int, allow bool) (IPAccessRule, error) {
	if prefix == "" {
		return IPAccessRule{}, fmt.Errorf("prefix cannot be empty")
	}
	
	_, ipNet, err := net.ParseCIDR(prefix)
	if err != nil {
		return IPAccessRule{}, fmt.Errorf("invalid CIDR prefix: %s", prefix)
	}
	
	rule := IPAccessRule{
		ipNet: ipNet,
		ports: make([]int, len(ports)),
		allow: allow,
	}
	
	// Копируем и сортируем порты
	copy(rule.ports, ports)
	sort.Ints(rule.ports)
	
	// Валидация портов
	for _, port := range rule.ports {
		if port < 1 || port > 65535 {
			return IPAccessRule{}, fmt.Errorf("invalid port %d, must be between 1 and 65535", port)
		}
	}
	
	return rule, nil
}

// Allowed проверяет, разрешен ли доступ для IP и порта
func (p *IPAccessPolicy) Allowed(ip net.IP, port int) (bool, *IPAccessRule) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	if len(p.rules) == 0 {
		return p.defaultAllow, nil
	}
	
	for _, rule := range p.rules {
		if rule.ipNet.Contains(ip) {
			// Если порты не указаны, применяем правило ко всем портам
			if len(rule.ports) == 0 {
				return rule.allow, &rule
			}
			
			// Проверяем, есть ли порт в списке разрешенных/запрещенных
			if pos := sort.SearchInts(rule.ports, port); pos < len(rule.ports) && rule.ports[pos] == port {
				return rule.allow, &rule
			}
		}
	}
	
	return p.defaultAllow, nil
}

// AddRule добавляет новое правило в политику
func (p *IPAccessPolicy) AddRule(prefix string, ports []int, allow bool) error {
	rule, err := NewIPAccessRule(prefix, ports, allow)
	if err != nil {
		return err
	}
	
	p.mu.Lock()
	defer p.mu.Unlock()
	
	p.rules = append(p.rules, rule)
	
	p.logger.Info("IP access rule added",
		zap.String("prefix", prefix),
		zap.Ints("ports", ports),
		zap.Bool("allow", allow),
	)
	
	return nil
}

// RemoveRule удаляет правило по индексу
func (p *IPAccessPolicy) RemoveRule(index int) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if index < 0 || index >= len(p.rules) {
		return fmt.Errorf("invalid rule index: %d", index)
	}
	
	rule := p.rules[index]
	p.rules = append(p.rules[:index], p.rules[index+1:]...)
	
	p.logger.Info("IP access rule removed",
		zap.String("prefix", rule.ipNet.String()),
		zap.Ints("ports", rule.ports),
		zap.Bool("allow", rule.allow),
	)
	
	return nil
}

// GetRules возвращает копию всех правил
func (p *IPAccessPolicy) GetRules() []IPAccessRule {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	rules := make([]IPAccessRule, len(p.rules))
	copy(rules, p.rules)
	
	return rules
}

// GetRuleCount возвращает количество правил
func (p *IPAccessPolicy) GetRuleCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	return len(p.rules)
}

// IsDefaultAllow возвращает значение по умолчанию
func (p *IPAccessPolicy) IsDefaultAllow() bool {
	return p.defaultAllow
}

// Allow проверяет, разрешен ли доступ для данного IP адреса (алиас для Allowed)
func (p *IPAccessPolicy) Allow(addr string) bool {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
	}
	
	ip := net.ParseIP(host)
	if ip == nil {
		p.logger.Warn("Invalid IP address", zap.String("addr", addr))
		return p.defaultAllow
	}
	
	allowed, _ := p.Allowed(ip, 0) // 0 означает любой порт
	return allowed
}

// SetDefaultAllow устанавливает значение по умолчанию
func (p *IPAccessPolicy) SetDefaultAllow(allow bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	p.defaultAllow = allow
	
	p.logger.Info("IP access default policy changed",
		zap.Bool("default_allow", allow),
	)
}

// String возвращает строковое представление правила
func (r *IPAccessRule) String() string {
	return fmt.Sprintf("prefix:%s/ports:%s/allow:%t", r.ipNet, r.PortsString(), r.allow)
}

// PortsString возвращает строковое представление портов
func (r *IPAccessRule) PortsString() string {
	if len(r.ports) == 0 {
		return "all"
	}
	return fmt.Sprintf("%v", r.ports)
}

// GetPorts возвращает список портов
func (r *IPAccessRule) GetPorts() []int {
	ports := make([]int, len(r.ports))
	copy(ports, r.ports)
	return ports
}

// IsAllow возвращает, разрешает ли правило доступ
func (r *IPAccessRule) IsAllow() bool {
	return r.allow
}

// GetCIDR возвращает CIDR префикс
func (r *IPAccessRule) GetCIDR() string {
	return r.ipNet.String()
}

// ContainsIP проверяет, содержит ли правило указанный IP
func (r *IPAccessRule) ContainsIP(ip net.IP) bool {
	return r.ipNet.Contains(ip)
}

// ContainsPort проверяет, содержит ли правило указанный порт
func (r *IPAccessRule) ContainsPort(port int) bool {
	if len(r.ports) == 0 {
		return true // Все порты
	}
	
	pos := sort.SearchInts(r.ports, port)
	return pos < len(r.ports) && r.ports[pos] == port
}

// ValidateIP проверяет корректность IP адреса
func ValidateIP(ipStr string) (net.IP, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return nil, fmt.Errorf("invalid IP address: %s", ipStr)
	}
	return ip, nil
}

// ValidateCIDR проверяет корректность CIDR нотации
func ValidateCIDR(cidr string) (*net.IPNet, error) {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, fmt.Errorf("invalid CIDR: %s", cidr)
	}
	return ipNet, nil
}

// ValidatePort проверяет корректность порта
func ValidatePort(port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535, got %d", port)
	}
	return nil
}

// IPAccessStats представляет статистику IP access control
type IPAccessStats struct {
	TotalRequests   int64
	AllowedRequests int64
	DeniedRequests  int64
	CacheHits       int64
	CacheMisses     int64
}

// IPAccessPolicyWithStats расширяет IPAccessPolicy статистикой
type IPAccessPolicyWithStats struct {
	*IPAccessPolicy
	stats IPAccessStats
	mu    sync.RWMutex
}

// NewIPAccessPolicyWithStats создает политику с поддержкой статистики
func NewIPAccessPolicyWithStats(config IPAccessConfig, logger *logger.Logger) (*IPAccessPolicyWithStats, error) {
	policy, err := NewIPAccessPolicy(config, logger)
	if err != nil {
		return nil, err
	}
	
	return &IPAccessPolicyWithStats{
		IPAccessPolicy: policy,
		stats:          IPAccessStats{},
	}, nil
}

// AllowedWithStats проверяет доступ и обновляет статистику
func (p *IPAccessPolicyWithStats) AllowedWithStats(ip net.IP, port int) (bool, *IPAccessRule) {
	p.mu.Lock()
	p.stats.TotalRequests++
	p.mu.Unlock()
	
	allowed, rule := p.IPAccessPolicy.Allowed(ip, port)
	
	p.mu.Lock()
	if allowed {
		p.stats.AllowedRequests++
	} else {
		p.stats.DeniedRequests++
	}
	p.mu.Unlock()
	
	return allowed, rule
}

// GetStats возвращает статистику
func (p *IPAccessPolicyWithStats) GetStats() IPAccessStats {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	return p.stats
}

// ResetStats сбрасывает статистику
func (p *IPAccessPolicyWithStats) ResetStats() {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	p.stats = IPAccessStats{}
} 