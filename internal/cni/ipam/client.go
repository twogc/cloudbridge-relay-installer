package ipam

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/containernetworking/cni/pkg/types"
	"go.uber.org/zap"
)

// IPAllocation represents an IP allocation from IPAM service
type IPAllocation struct {
	Address *net.IPNet     `json:"address"`
	Gateway net.IP         `json:"gateway"`
	Routes  []*types.Route `json:"routes"`
	DNS     types.DNS      `json:"dns"`
}

// IPAMRequest represents a request to IPAM service
type IPAMRequest struct {
	ContainerID string `json:"containerId"`
	Netns       string `json:"netns"`
	TenantID    string `json:"tenantId"`
	PodName     string `json:"podName,omitempty"`
	Namespace   string `json:"namespace,omitempty"`
}

// IPAMResponse represents a response from IPAM service
type IPAMResponse struct {
	Success bool         `json:"success"`
	Data    IPAllocation `json:"data,omitempty"`
	Error   string       `json:"error,omitempty"`
}

// Client represents IPAM service client
type Client struct {
	serviceURL string
	tenantID   string
	httpClient *http.Client
	logger     *zap.Logger
}

// NewClient creates a new IPAM client
func NewClient(serviceURL, tenantID string, logger *zap.Logger) (*Client, error) {
	if serviceURL == "" {
		return nil, fmt.Errorf("IPAM service URL is required")
	}
	if tenantID == "" {
		return nil, fmt.Errorf("tenant ID is required")
	}

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:       10,
			IdleConnTimeout:    30 * time.Second,
			DisableCompression: true,
		},
	}

	return &Client{
		serviceURL: serviceURL,
		tenantID:   tenantID,
		httpClient: httpClient,
		logger:     logger,
	}, nil
}

// AllocateIP allocates an IP address from IPAM service
func (c *Client) AllocateIP(containerID, netns string) (*IPAllocation, error) {
	c.logger.Info("Allocating IP from IPAM service",
		zap.String("containerID", containerID),
		zap.String("netns", netns),
		zap.String("tenantID", c.tenantID),
	)

	request := IPAMRequest{
		ContainerID: containerID,
		Netns:       netns,
		TenantID:    c.tenantID,
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal IPAM request: %v", err)
	}

	// Make HTTP request to IPAM service
	url := fmt.Sprintf("%s/api/v1/ipam/allocate", c.serviceURL)
	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to make IPAM request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("IPAM service returned status %d", resp.StatusCode)
	}

	var response IPAMResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode IPAM response: %v", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("IPAM allocation failed: %s", response.Error)
	}

	c.logger.Info("IP allocated successfully",
		zap.String("address", response.Data.Address.String()),
		zap.String("gateway", response.Data.Gateway.String()),
	)

	return &response.Data, nil
}

// ReleaseIP releases an IP address back to IPAM service
func (c *Client) ReleaseIP(containerID string) error {
	c.logger.Info("Releasing IP to IPAM service",
		zap.String("containerID", containerID),
		zap.String("tenantID", c.tenantID),
	)

	request := IPAMRequest{
		ContainerID: containerID,
		TenantID:    c.tenantID,
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal IPAM request: %v", err)
	}

	// Make HTTP request to IPAM service
	url := fmt.Sprintf("%s/api/v1/ipam/release", c.serviceURL)
	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("failed to make IPAM request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("IPAM service returned status %d", resp.StatusCode)
	}

	var response IPAMResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to decode IPAM response: %v", err)
	}

	if !response.Success {
		return fmt.Errorf("IPAM release failed: %s", response.Error)
	}

	c.logger.Info("IP released successfully")
	return nil
}

// CheckIP checks if an IP allocation is still valid
func (c *Client) CheckIP(containerID string) error {
	c.logger.Info("Checking IP allocation",
		zap.String("containerID", containerID),
		zap.String("tenantID", c.tenantID),
	)

	request := IPAMRequest{
		ContainerID: containerID,
		TenantID:    c.tenantID,
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal IPAM request: %v", err)
	}

	// Make HTTP request to IPAM service
	url := fmt.Sprintf("%s/api/v1/ipam/check", c.serviceURL)
	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("failed to make IPAM request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("IPAM service returned status %d", resp.StatusCode)
	}

	var response IPAMResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to decode IPAM response: %v", err)
	}

	if !response.Success {
		return fmt.Errorf("IPAM check failed: %s", response.Error)
	}

	c.logger.Info("IP allocation is valid")
	return nil
}

// GetTenantSubnet returns the subnet for the tenant
func (c *Client) GetTenantSubnet() (*net.IPNet, error) {
	c.logger.Info("Getting tenant subnet",
		zap.String("tenantID", c.tenantID),
	)

	// Make HTTP request to IPAM service
	url := fmt.Sprintf("%s/api/v1/ipam/tenant/%s/subnet", c.serviceURL, c.tenantID)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to make IPAM request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("IPAM service returned status %d", resp.StatusCode)
	}

	var response struct {
		Success bool   `json:"success"`
		Subnet  string `json:"subnet"`
		Error   string `json:"error,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode IPAM response: %v", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("IPAM subnet request failed: %s", response.Error)
	}

	_, subnet, err := net.ParseCIDR(response.Subnet)
	if err != nil {
		return nil, fmt.Errorf("failed to parse subnet: %v", err)
	}

	c.logger.Info("Tenant subnet retrieved",
		zap.String("subnet", subnet.String()),
	)

	return subnet, nil
}
