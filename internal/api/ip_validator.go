package api

import (
	"fmt"
	"net"
	"regexp"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

// NetworkConfig сетевая конфигурация
type NetworkConfig struct {
	Subnet string   `json:"subnet"`
	DNS    []string `json:"dns"`
	MTU    int      `json:"mtu"`
}

// IPValidator validates IP addresses and CIDR ranges
type IPValidator struct {
	allowedCIDRs []*net.IPNet
	allowedIPs   map[string]bool
	mu           sync.RWMutex
	logger       *zap.Logger
	metrics      *IPValidatorMetrics
}

// IPValidatorMetrics represents metrics for IP validation
type IPValidatorMetrics struct {
	checksTotal     prometheus.Counter
	allowedTotal    prometheus.Counter
	deniedTotal     prometheus.Counter
	violationsTotal *prometheus.CounterVec
}

// NewIPValidator creates a new IP validator
func NewIPValidator(logger *zap.Logger) *IPValidator {
	return &IPValidator{
		allowedCIDRs: make([]*net.IPNet, 0),
		allowedIPs:   make(map[string]bool),
		logger:       logger,
		metrics:      NewIPValidatorMetrics(),
	}
}

// NewIPValidatorMetrics creates new IP validator metrics
func NewIPValidatorMetrics() *IPValidatorMetrics {
	return &IPValidatorMetrics{
		checksTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "cloudbridge_ip_validator_checks_total",
			Help: "Total number of IP validation checks",
		}),
		allowedTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "cloudbridge_ip_validator_allowed_total",
			Help: "Total number of allowed IPs",
		}),
		deniedTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "cloudbridge_ip_validator_denied_total",
			Help: "Total number of denied IPs",
		}),
		violationsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "cloudbridge_ip_validator_violations_total",
			Help: "Total number of IP validation violations",
		}, []string{"violation_type"}),
	}
}

// RegisterMetrics registers IP validator metrics
func (iv *IPValidator) RegisterMetrics(reg prometheus.Registerer) {
	reg.MustRegister(
		iv.metrics.checksTotal,
		iv.metrics.allowedTotal,
		iv.metrics.deniedTotal,
		iv.metrics.violationsTotal,
	)
}

// AddAllowedCIDR adds a CIDR range to the allowed list
func (iv *IPValidator) AddAllowedCIDR(cidr string) error {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return fmt.Errorf("invalid CIDR: %s", cidr)
	}

	iv.mu.Lock()
	defer iv.mu.Unlock()

	iv.allowedCIDRs = append(iv.allowedCIDRs, ipNet)
	iv.logger.Info("CIDR added to allowed list",
		zap.String("cidr", cidr),
	)
	return nil
}

// AddAllowedIP adds a single IP to the allowed list
func (iv *IPValidator) AddAllowedIP(ip string) error {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return fmt.Errorf("invalid IP: %s", ip)
	}

	iv.mu.Lock()
	defer iv.mu.Unlock()

	iv.allowedIPs[ip] = true
	iv.logger.Info("IP added to allowed list",
		zap.String("ip", ip),
	)
	return nil
}

// IsIPAllowed checks if an IP is allowed
func (iv *IPValidator) IsIPAllowed(ip string) bool {
	iv.mu.RLock()
	defer iv.mu.RUnlock()

	iv.metrics.checksTotal.Inc()

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		iv.metrics.deniedTotal.Inc()
		iv.metrics.violationsTotal.WithLabelValues("invalid_ip").Inc()
		iv.logger.Warn("Invalid IP address",
			zap.String("ip", ip),
		)
		return false
	}

	// Check exact IP match
	if iv.allowedIPs[ip] {
		iv.metrics.allowedTotal.Inc()
		return true
	}

	// Check CIDR ranges
	for _, cidr := range iv.allowedCIDRs {
		if cidr.Contains(parsedIP) {
			iv.metrics.allowedTotal.Inc()
			return true
		}
	}

	iv.metrics.deniedTotal.Inc()
	iv.metrics.violationsTotal.WithLabelValues("ip_not_allowed").Inc()
	iv.logger.Warn("IP not in allowed list",
		zap.String("ip", ip),
	)
	return false
}

// ValidateWireGuardEndpoint validates WireGuard endpoint IP
func (iv *IPValidator) ValidateWireGuardEndpoint(endpoint string) error {
	// Extract IP from endpoint (format: "ip:port")
	parts := strings.Split(endpoint, ":")
	if len(parts) != 2 {
		return fmt.Errorf("invalid endpoint format: %s", endpoint)
	}

	ip := parts[0]
	if !iv.IsIPAllowed(ip) {
		return fmt.Errorf("endpoint IP not allowed: %s", ip)
	}

	// Validate port
	port := parts[1]
	if !isValidPort(port) {
		return fmt.Errorf("invalid port: %s", port)
	}

	return nil
}

// ValidateAllowedIPs validates WireGuard allowed IPs
func (iv *IPValidator) ValidateAllowedIPs(allowedIPs []string) error {
	for _, allowedIP := range allowedIPs {
		// Check if it's a valid CIDR or IP
		if !isValidCIDROrIP(allowedIP) {
			return fmt.Errorf("invalid allowed IP: %s", allowedIP)
		}
	}
	return nil
}

// LoadFromNetworkConfig loads allowed IPs from network configuration
func (iv *IPValidator) LoadFromNetworkConfig(networkConfig *NetworkConfig) error {
	iv.mu.Lock()
	defer iv.mu.Unlock()

	// Clear existing allowed IPs
	iv.allowedIPs = make(map[string]bool)
	iv.allowedCIDRs = make([]*net.IPNet, 0)

	if networkConfig == nil {
		return nil
	}

	// Add subnet to allowed CIDRs
	if networkConfig.Subnet != "" {
		if err := iv.AddAllowedCIDR(networkConfig.Subnet); err != nil {
			return fmt.Errorf("failed to add subnet: %w", err)
		}
	}

	iv.logger.Info("IP validator loaded from network config",
		zap.String("subnet", networkConfig.Subnet),
	)
	return nil
}

// GetAllowedIPs returns a copy of allowed IPs
func (iv *IPValidator) GetAllowedIPs() []string {
	iv.mu.RLock()
	defer iv.mu.RUnlock()

	ips := make([]string, 0, len(iv.allowedIPs))
	for ip := range iv.allowedIPs {
		ips = append(ips, ip)
	}
	return ips
}

// GetAllowedCIDRs returns a copy of allowed CIDRs
func (iv *IPValidator) GetAllowedCIDRs() []string {
	iv.mu.RLock()
	defer iv.mu.RUnlock()

	cidrs := make([]string, 0, len(iv.allowedCIDRs))
	for _, cidr := range iv.allowedCIDRs {
		cidrs = append(cidrs, cidr.String())
	}
	return cidrs
}

// GetCount returns the number of allowed IPs and CIDRs
func (iv *IPValidator) GetCount() int {
	iv.mu.RLock()
	defer iv.mu.RUnlock()

	return len(iv.allowedIPs) + len(iv.allowedCIDRs)
}

// Clear clears the allowed IPs and CIDRs
func (iv *IPValidator) Clear() {
	iv.mu.Lock()
	defer iv.mu.Unlock()

	iv.allowedIPs = make(map[string]bool)
	iv.allowedCIDRs = make([]*net.IPNet, 0)
	iv.logger.Info("IP validator cleared")
}

// isValidPort checks if a port is valid
func isValidPort(port string) bool {
	// Simple port validation regex
	portRegex := regexp.MustCompile(`^([1-9][0-9]{0,3}|[1-5][0-9]{4}|6[0-4][0-9]{3}|65[0-4][0-9]{2}|655[0-2][0-9]|6553[0-5])$`)
	return portRegex.MatchString(port)
}

// isValidCIDROrIP checks if a string is a valid CIDR or IP
func isValidCIDROrIP(cidrOrIP string) bool {
	// Try parsing as CIDR first
	if _, _, err := net.ParseCIDR(cidrOrIP); err == nil {
		return true
	}

	// Try parsing as IP
	if net.ParseIP(cidrOrIP) != nil {
		return true
	}

	return false
}
