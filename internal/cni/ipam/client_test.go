package ipam

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/containernetworking/cni/pkg/types"
	"go.uber.org/zap"
)

func TestNewClient(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name       string
		serviceURL string
		tenantID   string
		wantErr    bool
	}{
		{
			name:       "valid client",
			serviceURL: "http://localhost:8080",
			tenantID:   "tenant-123",
			wantErr:    false,
		},
		{
			name:       "empty service URL",
			serviceURL: "",
			tenantID:   "tenant-123",
			wantErr:    true,
		},
		{
			name:       "empty tenant ID",
			serviceURL: "http://localhost:8080",
			tenantID:   "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.serviceURL, tt.tenantID, logger)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && client == nil {
				t.Errorf("NewClient() returned nil client")
			}
		})
	}
}

func TestClient_AllocateIP(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/ipam/allocate" {
			t.Errorf("Expected path /api/v1/ipam/allocate, got %s", r.URL.Path)
		}

		response := IPAMResponse{
			Success: true,
			Data: IPAllocation{
				Address: &net.IPNet{
					IP:   net.ParseIP("10.100.1.10"),
					Mask: net.CIDRMask(24, 32),
				},
				Gateway: net.ParseIP("10.100.1.1"),
				Routes: []*types.Route{
					{
						Dst: net.IPNet{
							IP:   net.ParseIP("0.0.0.0"),
							Mask: net.CIDRMask(0, 32),
						},
					},
				},
				DNS: types.DNS{
					Nameservers: []string{"8.8.8.8"},
					Domain:      "cluster.local",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	logger := zap.NewNop()
	client, err := NewClient(server.URL, "tenant-123", logger)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test successful allocation
	allocation, err := client.AllocateIP("container-123", "netns-123")
	if err != nil {
		t.Fatalf("AllocateIP() error = %v", err)
	}

	if allocation.Address.String() != "10.100.1.10/24" {
		t.Errorf("Expected address 10.100.1.10/24, got %s", allocation.Address.String())
	}

	if allocation.Gateway.String() != "10.100.1.1" {
		t.Errorf("Expected gateway 10.100.1.1, got %s", allocation.Gateway.String())
	}
}

func TestClient_ReleaseIP(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/ipam/release" {
			t.Errorf("Expected path /api/v1/ipam/release, got %s", r.URL.Path)
		}

		response := IPAMResponse{
			Success: true,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	logger := zap.NewNop()
	client, err := NewClient(server.URL, "tenant-123", logger)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test successful release
	err = client.ReleaseIP("container-123")
	if err != nil {
		t.Fatalf("ReleaseIP() error = %v", err)
	}
}

func TestClient_CheckIP(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/ipam/check" {
			t.Errorf("Expected path /api/v1/ipam/check, got %s", r.URL.Path)
		}

		response := IPAMResponse{
			Success: true,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	logger := zap.NewNop()
	client, err := NewClient(server.URL, "tenant-123", logger)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test successful check
	err = client.CheckIP("container-123")
	if err != nil {
		t.Fatalf("CheckIP() error = %v", err)
	}
}

func TestClient_GetTenantSubnet(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/api/v1/ipam/tenant/tenant-123/subnet"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		response := struct {
			Success bool   `json:"success"`
			Subnet  string `json:"subnet"`
		}{
			Success: true,
			Subnet:  "10.100.1.0/24",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	logger := zap.NewNop()
	client, err := NewClient(server.URL, "tenant-123", logger)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test successful subnet retrieval
	subnet, err := client.GetTenantSubnet()
	if err != nil {
		t.Fatalf("GetTenantSubnet() error = %v", err)
	}

	if subnet.String() != "10.100.1.0/24" {
		t.Errorf("Expected subnet 10.100.1.0/24, got %s", subnet.String())
	}
}
