package network

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/containernetworking/cni/pkg/types"
	"github.com/twogc/cloudbridge-relay-installer/internal/cni/ipam"
	"go.uber.org/zap"
)

// Manager handles network interface operations
type Manager struct {
	logger *zap.Logger
}

// NewManager creates a new network manager
func NewManager(logger *zap.Logger) (*Manager, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}

	return &Manager{
		logger: logger,
	}, nil
}

// CreateInterface creates a network interface in the specified namespace
func (m *Manager) CreateInterface(netns, ifName string, allocation *ipam.IPAllocation) error {
	m.logger.Info("Creating network interface",
		zap.String("netns", netns),
		zap.String("ifName", ifName),
		zap.String("address", allocation.Address.String()),
		zap.String("gateway", allocation.Gateway.String()),
	)

	// Generate host interface name
	hostIfName := fmt.Sprintf("veth%s", ifName)

	// Create veth pair
	if err := m.createVethPair(hostIfName, ifName); err != nil {
		return fmt.Errorf("failed to create veth pair: %v", err)
	}

	// Move interface to target namespace
	if err := m.moveInterfaceToNamespace(ifName, netns); err != nil {
		// Cleanup on failure
		m.cleanupInterface(hostIfName)
		return fmt.Errorf("failed to move interface to namespace: %v", err)
	}

	// Configure interface in namespace
	if err := m.configureInterfaceInNamespace(netns, ifName, allocation); err != nil {
		// Cleanup on failure
		m.cleanupInterface(hostIfName)
		return fmt.Errorf("failed to configure interface: %v", err)
	}

	// Configure host interface
	if err := m.configureHostInterface(hostIfName); err != nil {
		m.logger.Warn("Failed to configure host interface", zap.Error(err))
		// Don't fail the operation for host interface issues
	}

	m.logger.Info("Network interface created successfully")
	return nil
}

// RemoveInterface removes a network interface
func (m *Manager) RemoveInterface(netns, ifName string) error {
	m.logger.Info("Removing network interface",
		zap.String("netns", netns),
		zap.String("ifName", ifName),
	)

	// Generate host interface name
	hostIfName := fmt.Sprintf("veth%s", ifName)

	// Remove interface from namespace
	if err := m.removeInterfaceFromNamespace(netns, ifName); err != nil {
		m.logger.Warn("Failed to remove interface from namespace", zap.Error(err))
	}

	// Cleanup host interface
	if err := m.cleanupInterface(hostIfName); err != nil {
		m.logger.Warn("Failed to cleanup host interface", zap.Error(err))
	}

	m.logger.Info("Network interface removed successfully")
	return nil
}

// CheckInterface checks if a network interface exists and is configured
func (m *Manager) CheckInterface(netns, ifName string) error {
	m.logger.Info("Checking network interface",
		zap.String("netns", netns),
		zap.String("ifName", ifName),
	)

	// Check if interface exists in namespace
	if err := m.checkInterfaceExists(netns, ifName); err != nil {
		return fmt.Errorf("interface check failed: %v", err)
	}

	// Check interface configuration
	if err := m.checkInterfaceConfiguration(netns, ifName); err != nil {
		return fmt.Errorf("interface configuration check failed: %v", err)
	}

	m.logger.Info("Network interface check passed")
	return nil
}

// CreateNetworkPolicy creates a Calico network policy for tenant isolation
func (m *Manager) CreateNetworkPolicy(containerID, tenantID string) error {
	m.logger.Info("Creating network policy",
		zap.String("containerID", containerID),
		zap.String("tenantID", tenantID),
	)

	// Generate network policy YAML
	policyYAML := m.generateNetworkPolicyYAML(containerID, tenantID)

	// Apply network policy using kubectl
	if err := m.applyNetworkPolicy(policyYAML); err != nil {
		return fmt.Errorf("failed to apply network policy: %v", err)
	}

	m.logger.Info("Network policy created successfully")
	return nil
}

// RemoveNetworkPolicy removes a Calico network policy
func (m *Manager) RemoveNetworkPolicy(containerID string) error {
	m.logger.Info("Removing network policy",
		zap.String("containerID", containerID),
	)

	// Remove network policy using kubectl
	policyName := fmt.Sprintf("pod-%s-policy", containerID)
	if err := m.deleteNetworkPolicy(policyName); err != nil {
		return fmt.Errorf("failed to delete network policy: %v", err)
	}

	m.logger.Info("Network policy removed successfully")
	return nil
}

// createVethPair creates a veth pair
func (m *Manager) createVethPair(hostIfName, containerIfName string) error {
	cmd := exec.Command("ip", "link", "add", hostIfName, "type", "veth", "peer", "name", containerIfName)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create veth pair: %s", string(output))
	}

	// Bring up host interface
	cmd = exec.Command("ip", "link", "set", hostIfName, "up")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to bring up host interface: %s", string(output))
	}

	return nil
}

// moveInterfaceToNamespace moves interface to target namespace
func (m *Manager) moveInterfaceToNamespace(ifName, netns string) error {
	cmd := exec.Command("ip", "link", "set", ifName, "netns", netns)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to move interface to namespace: %s", string(output))
	}

	return nil
}

// configureInterfaceInNamespace configures interface inside namespace
func (m *Manager) configureInterfaceInNamespace(netns, ifName string, allocation *ipam.IPAllocation) error {
	// Set IP address
	cmd := exec.Command("ip", "netns", "exec", netns, "ip", "addr", "add", allocation.Address.String(), "dev", ifName)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to set IP address: %s", string(output))
	}

	// Bring up interface
	cmd = exec.Command("ip", "netns", "exec", netns, "ip", "link", "set", ifName, "up")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to bring up interface: %s", string(output))
	}

	// Set default route
	if allocation.Gateway != nil {
		cmd = exec.Command("ip", "netns", "exec", netns, "ip", "route", "add", "default", "via", allocation.Gateway.String())
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to set default route: %s", string(output))
		}
	}

	// Configure DNS if provided
	if len(allocation.DNS.Nameservers) > 0 {
		if err := m.configureDNS(netns, allocation.DNS); err != nil {
			m.logger.Warn("Failed to configure DNS", zap.Error(err))
		}
	}

	return nil
}

// configureHostInterface configures the host interface
func (m *Manager) configureHostInterface(hostIfName string) error {
	// Enable forwarding
	cmd := exec.Command("sysctl", "-w", "net.ipv4.ip_forward=1")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to enable IP forwarding: %s", string(output))
	}

	// Add iptables rules for NAT (if needed)
	// This is a simplified version - in production, more sophisticated rules would be needed
	cmd = exec.Command("iptables", "-t", "nat", "-A", "POSTROUTING", "-o", hostIfName, "-j", "MASQUERADE")
	if output, err := cmd.CombinedOutput(); err != nil {
		m.logger.Warn("Failed to add iptables NAT rule", zap.String("output", string(output)))
	}

	return nil
}

// cleanupInterface removes an interface
func (m *Manager) cleanupInterface(ifName string) error {
	cmd := exec.Command("ip", "link", "delete", ifName)
	if output, err := cmd.CombinedOutput(); err != nil {
		// Interface might not exist, which is OK
		m.logger.Debug("Interface cleanup failed (might not exist)", zap.String("output", string(output)))
	}

	return nil
}

// removeInterfaceFromNamespace removes interface from namespace
func (m *Manager) removeInterfaceFromNamespace(netns, ifName string) error {
	cmd := exec.Command("ip", "netns", "exec", netns, "ip", "link", "delete", ifName)
	if output, err := cmd.CombinedOutput(); err != nil {
		// Interface might not exist, which is OK
		m.logger.Debug("Interface removal failed (might not exist)", zap.String("output", string(output)))
	}

	return nil
}

// checkInterfaceExists checks if interface exists in namespace
func (m *Manager) checkInterfaceExists(netns, ifName string) error {
	cmd := exec.Command("ip", "netns", "exec", netns, "ip", "link", "show", ifName)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("interface does not exist: %s", string(output))
	}

	return nil
}

// checkInterfaceConfiguration checks interface configuration
func (m *Manager) checkInterfaceConfiguration(netns, ifName string) error {
	// Check if interface is up
	cmd := exec.Command("ip", "netns", "exec", netns, "ip", "link", "show", ifName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to check interface status: %s", string(output))
	}

	if !strings.Contains(string(output), "state UP") {
		return fmt.Errorf("interface is not up")
	}

	return nil
}

// configureDNS configures DNS settings
func (m *Manager) configureDNS(netns string, dns types.DNS) error {
	// This is a simplified DNS configuration
	// In production, more sophisticated DNS management would be needed
	if len(dns.Nameservers) > 0 {
		// Create resolv.conf in namespace
		resolvConf := "/etc/netns/" + netns + "/resolv.conf"
		if err := os.MkdirAll(filepath.Dir(resolvConf), 0755); err != nil {
			return fmt.Errorf("failed to create DNS config directory: %v", err)
		}

		file, err := os.Create(resolvConf)
		if err != nil {
			return fmt.Errorf("failed to create resolv.conf: %v", err)
		}
		defer file.Close()

		for _, ns := range dns.Nameservers {
			file.WriteString(fmt.Sprintf("nameserver %s\n", ns))
		}

		if dns.Domain != "" {
			file.WriteString(fmt.Sprintf("search %s\n", dns.Domain))
		}
	}

	return nil
}

// generateNetworkPolicyYAML generates Calico network policy YAML
func (m *Manager) generateNetworkPolicyYAML(containerID, tenantID string) string {
	policyName := fmt.Sprintf("pod-%s-policy", containerID)

	return fmt.Sprintf(`apiVersion: projectcalico.org/v3
kind: NetworkPolicy
metadata:
  name: %s
  namespace: default
spec:
  selector: pod-name == "%s"
  types: [Ingress, Egress]
  ingress:
  - action: Allow
    source:
      selector: tenant.cloudbridge.io/id == "%s"
  egress:
  - action: Allow
    destination:
      selector: tenant.cloudbridge.io/id == "%s"
`, policyName, containerID, tenantID, tenantID)
}

// applyNetworkPolicy applies network policy using kubectl
func (m *Manager) applyNetworkPolicy(policyYAML string) error {
	cmd := exec.Command("kubectl", "apply", "-f", "-")
	cmd.Stdin = strings.NewReader(policyYAML)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to apply network policy: %s", string(output))
	}

	return nil
}

// deleteNetworkPolicy deletes network policy using kubectl
func (m *Manager) deleteNetworkPolicy(policyName string) error {
	cmd := exec.Command("kubectl", "delete", "networkpolicy", policyName, "--ignore-not-found=true")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to delete network policy: %s", string(output))
	}

	return nil
}
