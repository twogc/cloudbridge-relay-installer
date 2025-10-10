package database

import (
	"fmt"
	"net"
	"strings"
)

// SubnetRange represents the usable IP range within a subnet
type SubnetRange struct {
	FirstIP uint32
	LastIP  uint32
	Network string
}

// isReservedIP checks if an IP is reserved (gateway, broadcast, etc.)
func isReservedIP(ip string) bool {
	// Common reserved IPs in subnets
	reserved := []string{
		".1",   // Gateway
		".2",   // DNS
		".3",   // Reserved
		".4",   // Reserved
		".5",   // Reserved
		".6",   // Reserved
		".7",   // Reserved
		".8",   // Reserved
		".9",   // Reserved
		".255", // Broadcast
	}

	for _, suffix := range reserved {
		if strings.HasSuffix(ip, suffix) {
			return true
		}
	}

	return false
}

// parseSubnetRange calculates the usable IP range for a given CIDR
func parseSubnetRange(cidr string) (*SubnetRange, error) {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, fmt.Errorf("invalid CIDR %s: %w", cidr, err)
	}

	// Get network and broadcast addresses
	network := ipNet.IP.To4()
	if network == nil {
		return nil, fmt.Errorf("invalid IPv4 CIDR: %s", cidr)
	}

	mask := ipNet.Mask
	ones, bits := mask.Size()

	// Calculate number of usable hosts
	hostBits := bits - ones
	if hostBits < 2 {
		return nil, fmt.Errorf("subnet too small for usable hosts: %s", cidr)
	}

	// Convert to uint32 for easier manipulation
	networkUint := ipv4ToUint32(network)

	// Calculate first and last usable IPs
	// First usable: network + 1 (skip network address)
	// Last usable: broadcast - 1 (skip broadcast address)
	firstIP := networkUint + 1
	lastIP := networkUint + (1 << hostBits) - 2

	return &SubnetRange{
		FirstIP: firstIP,
		LastIP:  lastIP,
		Network: cidr,
	}, nil
}

// ipv4ToUint32 converts IPv4 address to uint32
func ipv4ToUint32(ip net.IP) uint32 {
	ip = ip.To4()
	return uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3])
}

// uint32ToIPv4 converts uint32 to IPv4 address
func uint32ToIPv4(ip uint32) net.IP {
	return net.IPv4(
		byte(ip>>24),
		byte(ip>>16),
		byte(ip>>8),
		byte(ip),
	)
}

// nextUsableIP finds the next usable IP in the range, skipping reserved addresses
func nextUsableIP(current uint32, range_ *SubnetRange) (uint32, bool) {
	for ip := current; ip <= range_.LastIP; ip++ {
		ipStr := uint32ToIPv4(ip).String()
		if !isReservedIP(ipStr) {
			return ip, true
		}
	}
	return 0, false
}

// isUniqueViolation checks if the error is a unique constraint violation
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "unique") ||
		strings.Contains(errStr, "duplicate") ||
		strings.Contains(errStr, "constraint")
}

// getReservedIPs returns a list of commonly reserved IPs in a subnet
func getReservedIPs(cidr string) ([]string, error) {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	network := ipNet.IP.To4()
	if network == nil {
		return nil, fmt.Errorf("invalid IPv4 CIDR: %s", cidr)
	}

	// Common reserved IPs
	reserved := []string{
		fmt.Sprintf("%d.%d.%d.1", network[0], network[1], network[2]),   // Gateway
		fmt.Sprintf("%d.%d.%d.2", network[0], network[1], network[2]),   // DNS
		fmt.Sprintf("%d.%d.%d.255", network[0], network[1], network[2]), // Broadcast
	}

	return reserved, nil
}

