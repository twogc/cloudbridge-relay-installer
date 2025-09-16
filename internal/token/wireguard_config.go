package token

import "fmt"

// WireGuardConfig represents WireGuard configuration in JWT
type WireGuardConfig struct {
	InterfaceName string   `json:"interface_name"`
	ListenPort    int      `json:"listen_port"`
	Address       string   `json:"address"`
	MTU           int      `json:"mtu"`
	AllowedIPs    []string `json:"allowed_ips"`
	PrivateKey    string   `json:"private_key,omitempty"`
	PublicKey     string   `json:"public_key,omitempty"`
	Endpoint      string   `json:"endpoint,omitempty"`
}

// Validate validates WireGuard configuration
func (wg *WireGuardConfig) Validate() error {
	if wg.InterfaceName == "" {
		return fmt.Errorf("interface_name is required")
	}

	if wg.ListenPort <= 0 || wg.ListenPort > 65535 {
		return fmt.Errorf("listen_port must be between 1 and 65535, got: %d", wg.ListenPort)
	}

	if wg.Address == "" {
		return fmt.Errorf("address is required")
	}

	if wg.MTU <= 0 {
		return fmt.Errorf("mtu must be positive, got: %d", wg.MTU)
	}

	for i, ip := range wg.AllowedIPs {
		if ip == "" {
			return fmt.Errorf("allowed_ips[%d] cannot be empty", i)
		}
		if len(ip) < 7 { // minimum: "0.0.0.0/0"
			return fmt.Errorf("allowed_ips[%d] appears to be invalid CIDR: %s", i, ip)
		}
	}

	// Validate endpoint if provided
	if wg.Endpoint != "" && len(wg.Endpoint) < 7 { // minimum: "1.1.1.1:1"
		return fmt.Errorf("endpoint appears to be invalid format: %s", wg.Endpoint)
	}

	return nil
}
