package main

import (
	"encoding/json"
	"fmt"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/containernetworking/cni/pkg/version"
	"github.com/twogc/cloudbridge-relay-installer/internal/cni"
	"github.com/twogc/cloudbridge-relay-installer/internal/cni/ipam"
	"github.com/twogc/cloudbridge-relay-installer/internal/cni/network"
	"go.uber.org/zap"
)

// CNI Plugin configuration
type CNIConfig struct {
	CNIVersion string `json:"cniVersion"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	IPAM       struct {
		Type       string `json:"type"`
		ServiceURL string `json:"serviceUrl"`
		TenantID   string `json:"tenantId"`
	} `json:"ipam"`
	Network struct {
		Type string `json:"type"`
	} `json:"network"`
	LogLevel string `json:"logLevel,omitempty"`
}

// CloudBridge CNI Plugin
type CloudBridgeCNI struct {
	config *CNIConfig
	logger *zap.Logger
	ipam   *ipam.Client
	net    *network.Manager
}

// NewCloudBridgeCNI creates a new CNI plugin instance
func NewCloudBridgeCNI(config *CNIConfig) (*CloudBridgeCNI, error) {
	// Initialize logger
	logger, err := cni.NewLogger(config.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %v", err)
	}

	// Initialize IPAM client
	ipamClient, err := ipam.NewClient(config.IPAM.ServiceURL, config.IPAM.TenantID, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize IPAM client: %v", err)
	}

	// Initialize network manager
	netManager, err := network.NewManager(logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize network manager: %v", err)
	}

	return &CloudBridgeCNI{
		config: config,
		logger: logger,
		ipam:   ipamClient,
		net:    netManager,
	}, nil
}

// Add implements CNI ADD command
func (c *CloudBridgeCNI) Add(args *skel.CmdArgs) error {
	c.logger.Info("CNI ADD called",
		zap.String("containerID", args.ContainerID),
		zap.String("netns", args.Netns),
		zap.String("ifName", args.IfName),
		zap.String("args", args.Args),
		zap.String("path", args.Path),
	)

	// Parse CNI configuration
	var config CNIConfig
	if err := json.Unmarshal(args.StdinData, &config); err != nil {
		return fmt.Errorf("failed to parse CNI configuration: %v", err)
	}

	// Initialize CNI plugin with parsed config
	cni, err := NewCloudBridgeCNI(&config)
	if err != nil {
		return fmt.Errorf("failed to initialize CNI plugin: %v", err)
	}

	// Allocate IP from IPAM service
	ipAllocation, err := cni.ipam.AllocateIP(args.ContainerID, args.Netns)
	if err != nil {
		return fmt.Errorf("failed to allocate IP: %v", err)
	}

	// Create network interface
	if err := cni.net.CreateInterface(args.Netns, args.IfName, ipAllocation); err != nil {
		// Cleanup allocated IP on failure
		if cleanupErr := cni.ipam.ReleaseIP(args.ContainerID); cleanupErr != nil {
			c.logger.Error("Failed to cleanup allocated IP", zap.Error(cleanupErr))
		}
		return fmt.Errorf("failed to create network interface: %v", err)
	}

	// Create network policy for tenant isolation
	if err := cni.net.CreateNetworkPolicy(args.ContainerID, config.IPAM.TenantID); err != nil {
		c.logger.Warn("Failed to create network policy", zap.Error(err))
		// Don't fail the operation for network policy issues
	}

	// Return result
	result := &current.Result{
		CNIVersion: config.CNIVersion,
		IPs: []*current.IPConfig{
			{
				Version: "4",
				Address: *ipAllocation.Address,
				Gateway: ipAllocation.Gateway,
			},
		},
		Routes: ipAllocation.Routes,
		DNS:    ipAllocation.DNS,
	}

	return types.PrintResult(result, config.CNIVersion)
}

// Del implements CNI DEL command
func (c *CloudBridgeCNI) Del(args *skel.CmdArgs) error {
	c.logger.Info("CNI DEL called",
		zap.String("containerID", args.ContainerID),
		zap.String("netns", args.Netns),
		zap.String("ifName", args.IfName),
	)

	// Parse CNI configuration
	var config CNIConfig
	if err := json.Unmarshal(args.StdinData, &config); err != nil {
		return fmt.Errorf("failed to parse CNI configuration: %v", err)
	}

	// Initialize CNI plugin with parsed config
	cni, err := NewCloudBridgeCNI(&config)
	if err != nil {
		return fmt.Errorf("failed to initialize CNI plugin: %v", err)
	}

	// Remove network interface
	if err := cni.net.RemoveInterface(args.Netns, args.IfName); err != nil {
		c.logger.Warn("Failed to remove network interface", zap.Error(err))
		// Continue with IP cleanup
	}

	// Release IP from IPAM service
	if err := cni.ipam.ReleaseIP(args.ContainerID); err != nil {
		c.logger.Warn("Failed to release IP", zap.Error(err))
		// Don't fail the operation for IP release issues
	}

	// Remove network policy
	if err := cni.net.RemoveNetworkPolicy(args.ContainerID); err != nil {
		c.logger.Warn("Failed to remove network policy", zap.Error(err))
		// Don't fail the operation for network policy issues
	}

	return nil
}

// Check implements CNI CHECK command
func (c *CloudBridgeCNI) Check(args *skel.CmdArgs) error {
	c.logger.Info("CNI CHECK called",
		zap.String("containerID", args.ContainerID),
		zap.String("netns", args.Netns),
		zap.String("ifName", args.IfName),
	)

	// Parse CNI configuration
	var config CNIConfig
	if err := json.Unmarshal(args.StdinData, &config); err != nil {
		return fmt.Errorf("failed to parse CNI configuration: %v", err)
	}

	// Initialize CNI plugin with parsed config
	cni, err := NewCloudBridgeCNI(&config)
	if err != nil {
		return fmt.Errorf("failed to initialize CNI plugin: %v", err)
	}

	// Check network interface status
	if err := cni.net.CheckInterface(args.Netns, args.IfName); err != nil {
		return fmt.Errorf("network interface check failed: %v", err)
	}

	// Check IP allocation status
	if err := cni.ipam.CheckIP(args.ContainerID); err != nil {
		return fmt.Errorf("IP allocation check failed: %v", err)
	}

	return nil
}

func main() {
	// CNI plugin entry point
	skel.PluginMain(
		func(args *skel.CmdArgs) error {
			cni := &CloudBridgeCNI{}
			return cni.Add(args)
		},
		func(args *skel.CmdArgs) error {
			cni := &CloudBridgeCNI{}
			return cni.Del(args)
		},
		func(args *skel.CmdArgs) error {
			cni := &CloudBridgeCNI{}
			return cni.Check(args)
		},
		version.All,
		"CloudBridge CNI Plugin for tenant-isolated networking")
}
