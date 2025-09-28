package api

import "time"

// ThroughputStats represents throughput statistics
type ThroughputStats struct {
	Current float64 `json:"current"` // Current throughput in bps
	Limit   float64 `json:"limit"`   // Throughput limit in bps
	Peak    float64 `json:"peak"`    // Peak throughput in bps
}

// ConnectionStats represents connection statistics
type ConnectionStats struct {
	Active int            `json:"active"`  // Active connections
	Total  int            `json:"total"`   // Total connections
	Peak   int            `json:"peak"`    // Peak connections
	ByType map[string]int `json:"by_type"` // Connections by type
}

// ResourceStats represents resource usage statistics
type ResourceStats struct {
	CPU     float64            `json:"cpu"`     // CPU usage percentage
	Memory  float64            `json:"memory"`  // Memory usage percentage
	Disk    float64            `json:"disk"`    // Disk usage percentage
	Network map[string]float64 `json:"network"` // Network usage by interface
}

// HealthStatus represents health status
type HealthStatus struct {
	Status    string            `json:"status"`    // healthy, warning, error
	Checks    map[string]string `json:"checks"`    // Individual health checks
	Timestamp time.Time         `json:"timestamp"` // Last check timestamp
}

// LicenseStatus represents license status
type LicenseStatus struct {
	Type      string         `json:"type"`       // basic, secure, enterprise
	ExpiresAt time.Time      `json:"expires_at"` // License expiration
	IsValid   bool           `json:"is_valid"`   // License validity
	Features  []string       `json:"features"`   // Available features
	Limits    map[string]int `json:"limits"`     // License limits
}

// NetworkInterface represents network interface configuration
type NetworkInterface struct {
	Name        string   `json:"name"`         // Interface name
	Type        string   `json:"type"`         // ethernet, wireless, etc.
	Speed       string   `json:"speed"`        // 1G, 10G, 25G, etc.
	IPAddresses []string `json:"ip_addresses"` // IP addresses
	VLANs       []int    `json:"vlans"`        // VLAN IDs
	Status      string   `json:"status"`       // up, down, unknown
}

// RoutingConfig represents routing configuration
type RoutingConfig struct {
	StaticRoutes []StaticRoute `json:"static_routes"` // Static routes
	BGP          *BGPConfig    `json:"bgp"`           // BGP configuration
	OSPF         *OSPFConfig   `json:"ospf"`          // OSPF configuration
}

// FirewallConfig represents firewall configuration
type FirewallConfig struct {
	Rules    []FirewallRule `json:"rules"`    // Firewall rules
	Policies []Policy       `json:"policies"` // Security policies
	Logging  *LoggingConfig `json:"logging"`  // Logging configuration
}

// WAFConfig represents Web Application Firewall configuration
type WAFConfig struct {
	Enabled   bool     `json:"enabled"`    // WAF enabled
	Rules     []string `json:"rules"`      // WAF rules
	Mode      string   `json:"mode"`       // block, monitor, log
	LogLevel  string   `json:"log_level"`  // debug, info, warn, error
	RateLimit int      `json:"rate_limit"` // Rate limit per second
}

// DDoSConfig represents DDoS protection configuration
type DDoSConfig struct {
	Enabled   bool     `json:"enabled"`   // DDoS protection enabled
	Threshold int      `json:"threshold"` // Attack threshold
	Action    string   `json:"action"`    // block, rate_limit, log
	Whitelist []string `json:"whitelist"` // Whitelisted IPs
	Blacklist []string `json:"blacklist"` // Blacklisted IPs
}

// ZeroTrustConfig represents Zero Trust configuration
type ZeroTrustConfig struct {
	Enabled      bool     `json:"enabled"`       // Zero Trust enabled
	Policies     []Policy `json:"policies"`      // Zero Trust policies
	DeviceTrust  bool     `json:"device_trust"`  // Device trust required
	UserTrust    bool     `json:"user_trust"`    // User trust required
	NetworkTrust bool     `json:"network_trust"` // Network trust required
}

// AuditConfig represents audit configuration
type AuditConfig struct {
	Enabled      bool     `json:"enabled"`      // Audit enabled
	LogLevel     string   `json:"log_level"`    // Audit log level
	Retention    int      `json:"retention"`    // Log retention days
	Destinations []string `json:"destinations"` // Log destinations
}

// StaticRoute represents a static route
type StaticRoute struct {
	Destination string `json:"destination"` // Destination network
	Gateway     string `json:"gateway"`     // Gateway IP
	Interface   string `json:"interface"`   // Outgoing interface
	Metric      int    `json:"metric"`      // Route metric
}

// BGPConfig represents BGP configuration
type BGPConfig struct {
	Enabled     bool     `json:"enabled"`     // BGP enabled
	ASNumber    int      `json:"as_number"`   // AS number
	Neighbors   []string `json:"neighbors"`   // BGP neighbors
	Communities []string `json:"communities"` // BGP communities
}

// OSPFConfig represents OSPF configuration
type OSPFConfig struct {
	Enabled    bool     `json:"enabled"`    // OSPF enabled
	AreaID     string   `json:"area_id"`    // OSPF area ID
	Interfaces []string `json:"interfaces"` // OSPF interfaces
	Cost       int      `json:"cost"`       // Interface cost
}

// FirewallRule represents a firewall rule
type FirewallRule struct {
	ID          string `json:"id"`          // Rule ID
	Name        string `json:"name"`        // Rule name
	Action      string `json:"action"`      // allow, deny, reject
	Source      string `json:"source"`      // Source network
	Destination string `json:"destination"` // Destination network
	Port        string `json:"port"`        // Port or port range
	Protocol    string `json:"protocol"`    // tcp, udp, icmp, etc.
	Enabled     bool   `json:"enabled"`     // Rule enabled
	Priority    int    `json:"priority"`    // Rule priority
}

// Policy represents a security policy
type Policy struct {
	ID         string            `json:"id"`         // Policy ID
	Name       string            `json:"name"`       // Policy name
	Type       string            `json:"type"`       // Policy type
	Rules      []string          `json:"rules"`      // Policy rules
	Conditions map[string]string `json:"conditions"` // Policy conditions
	Actions    []string          `json:"actions"`    // Policy actions
	Enabled    bool              `json:"enabled"`    // Policy enabled
	Priority   int               `json:"priority"`   // Policy priority
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Level        string   `json:"level"`        // Log level
	Format       string   `json:"format"`       // Log format
	Destinations []string `json:"destinations"` // Log destinations
	Retention    int      `json:"retention"`    // Log retention days
	Compression  bool     `json:"compression"`  // Log compression
}

// RelayServer represents the relay server component
type RelayServer struct {
	// Relay server implementation
}

// MonitoringSystem represents the monitoring system
type MonitoringSystem struct {
	// Monitoring system implementation
}
