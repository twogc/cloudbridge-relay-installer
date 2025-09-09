package token

import "fmt"

// NetworkConfig represents network configuration in JWT
type NetworkConfig struct {
	Subnet        string   `json:"subnet"`
	Gateway       string   `json:"gateway"`
	DNS           []string `json:"dns"`
	FirewallRules []string `json:"firewall_rules"`
	MTU           int      `json:"mtu"`
	EnableIPv6    bool     `json:"enable_ipv6"`
}

// Validate validates network configuration
func (n *NetworkConfig) Validate() error {
	if n.Subnet == "" {
		return fmt.Errorf("subnet is required")
	}

	if len(n.Subnet) < 7 { // minimum: "0.0.0.0/0"
		return fmt.Errorf("subnet appears to be invalid CIDR: %s", n.Subnet)
	}

	for i, dns := range n.DNS {
		if dns == "" {
			return fmt.Errorf("dns[%d] cannot be empty", i)
		}
	}

	if n.MTU <= 0 {
		return fmt.Errorf("mtu must be positive, got: %d", n.MTU)
	}

	return nil
}
