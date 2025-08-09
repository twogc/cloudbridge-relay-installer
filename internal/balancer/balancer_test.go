package balancer

import (
	"testing"
	"time"
)

func TestBalancer_AddRemoveServer(t *testing.T) {
	balancer := NewBalancer()

	// Test adding servers
	balancer.AddServer("server1:8080")
	balancer.AddServer("server2:8080")

	servers := balancer.GetServers()
	if len(servers) != 2 {
		t.Errorf("Expected 2 servers, got %d", len(servers))
	}

	// Test removing server
	balancer.RemoveServer("server1:8080")
	servers = balancer.GetServers()
	if len(servers) != 1 {
		t.Errorf("Expected 1 server, got %d", len(servers))
	}
	if servers[0].Address != "server2:8080" {
		t.Errorf("Expected server2:8080, got %s", servers[0].Address)
	}
}

func TestBalancer_GetNextServer(t *testing.T) {
	balancer := NewBalancer()

	// Add servers
	balancer.AddServer("server1:8080")
	balancer.AddServer("server2:8080")
	balancer.AddServer("server3:8080")

	// Test round-robin
	server1 := balancer.GetNextServer()
	if server1 == nil || server1.Address != "server1:8080" {
		t.Errorf("Expected server1:8080, got %v", server1)
	}

	server2 := balancer.GetNextServer()
	if server2 == nil || server2.Address != "server2:8080" {
		t.Errorf("Expected server2:8080, got %v", server2)
	}

	server3 := balancer.GetNextServer()
	if server3 == nil || server3.Address != "server3:8080" {
		t.Errorf("Expected server3:8080, got %v", server3)
	}

	// Test unhealthy server
	balancer.MarkServerUnhealthy("server1:8080")
	server4 := balancer.GetNextServer()
	if server4 == nil || server4.Address != "server2:8080" {
		t.Errorf("Expected server2:8080, got %v", server4)
	}
}

func TestBalancer_UpdateStats(t *testing.T) {
	balancer := NewBalancer()
	balancer.AddServer("server1:8080")

	// Update stats
	balancer.UpdateServerStats("server1:8080", 10, 100*time.Millisecond)

	servers := balancer.GetServers()
	if len(servers) != 1 {
		t.Fatalf("Expected 1 server, got %d", len(servers))
	}

	server := servers[0]
	if server.ActiveConns != 10 {
		t.Errorf("Expected 10 active connections, got %d", server.ActiveConns)
	}
	if server.Latency != 100*time.Millisecond {
		t.Errorf("Expected 100ms latency, got %v", server.Latency)
	}
}

func TestBalancer_HealthCheck(t *testing.T) {
	balancer := NewBalancer()
	balancer.AddServer("server1:8080")

	// Mark server as unhealthy
	balancer.MarkServerUnhealthy("server1:8080")
	servers := balancer.GetServers()
	if len(servers) != 1 {
		t.Fatalf("Expected 1 server, got %d", len(servers))
	}
	if servers[0].Healthy {
		t.Error("Expected server to be unhealthy")
	}

	// Mark server as healthy
	balancer.MarkServerHealthy("server1:8080")
	servers = balancer.GetServers()
	if !servers[0].Healthy {
		t.Error("Expected server to be healthy")
	}
}
