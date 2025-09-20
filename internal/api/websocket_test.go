package api

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewHub(t *testing.T) {
	logger := zap.NewNop()
	hub := NewHub(logger)

	assert.NotNil(t, hub)
	assert.NotNil(t, hub.Connections)
	assert.NotNil(t, hub.Register)
	assert.NotNil(t, hub.Unregister)
	assert.NotNil(t, hub.Broadcast)
	assert.NotNil(t, hub.upgrader)
	assert.Equal(t, logger, hub.logger)
}

func TestHub_GetConnectionCount(t *testing.T) {
	logger := zap.NewNop()
	hub := NewHub(logger)

	// Initially no connections
	assert.Equal(t, 0, hub.GetConnectionCount())

	// Add a connection
	conn := &WebSocketConnection{
		PeerID:   "test-peer-1",
		TenantID: "tenant-1",
	}
	hub.Connections["test-peer-1"] = conn

	assert.Equal(t, 1, hub.GetConnectionCount())
}

func TestHub_GetTenantConnectionCount(t *testing.T) {
	logger := zap.NewNop()
	hub := NewHub(logger)

	// Add connections for different tenants
	conn1 := &WebSocketConnection{
		PeerID:   "test-peer-1",
		TenantID: "tenant-1",
	}
	conn2 := &WebSocketConnection{
		PeerID:   "test-peer-2",
		TenantID: "tenant-1",
	}
	conn3 := &WebSocketConnection{
		PeerID:   "test-peer-3",
		TenantID: "tenant-2",
	}

	hub.Connections["test-peer-1"] = conn1
	hub.Connections["test-peer-2"] = conn2
	hub.Connections["test-peer-3"] = conn3

	// Test tenant-1 connections
	assert.Equal(t, 2, hub.GetTenantConnectionCount("tenant-1"))

	// Test tenant-2 connections
	assert.Equal(t, 1, hub.GetTenantConnectionCount("tenant-2"))

	// Test non-existent tenant
	assert.Equal(t, 0, hub.GetTenantConnectionCount("tenant-3"))
}

func TestWebSocketMessage_String(t *testing.T) {
	tests := []struct {
		name     string
		message  WebSocketMessage
		expected string
	}{
		{
			name: "peer_joined message",
			message: WebSocketMessage{
				Type: "peer_joined",
				Data: PeerJoinedData{
					PeerID:    "test-peer-1",
					Timestamp: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
				},
				Timestamp: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			},
			expected: `{"type":"peer_joined","data":{"peer_id":"test-peer-1","wireguard_config":null,"timestamp":"2023-01-01T12:00:00Z"},"timestamp":"2023-01-01T12:00:00Z"}`,
		},
		{
			name: "peer_left message",
			message: WebSocketMessage{
				Type: "peer_left",
				Data: PeerLeftData{
					PeerID:    "test-peer-1",
					Timestamp: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
				},
				Timestamp: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			},
			expected: `{"type":"peer_left","data":{"peer_id":"test-peer-1","timestamp":"2023-01-01T12:00:00Z"},"timestamp":"2023-01-01T12:00:00Z"}`,
		},
		{
			name: "peer_status_update message",
			message: WebSocketMessage{
				Type: "peer_status_update",
				Data: PeerStatusUpdateData{
					PeerID:     "test-peer-1",
					IsOnline:   true,
					LatencyMS:  50,
					MeshStatus: "connected",
					Timestamp:  time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
				},
				Timestamp: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			},
			expected: `{"type":"peer_status_update","data":{"peer_id":"test-peer-1","is_online":true,"latency_ms":50,"mesh_status":"connected","timestamp":"2023-01-01T12:00:00Z"},"timestamp":"2023-01-01T12:00:00Z"}`,
		},
		{
			name: "relay_message",
			message: WebSocketMessage{
				Type: "relay_message",
				Data: RelayMessageData{
					Type:        "relay_message",
					MessageType: "test-message",
					Data:        "test-data",
					Timestamp:   time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
				},
				Timestamp: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			},
			expected: `{"type":"relay_message","data":{"type":"relay_message","message_type":"test-message","data":"test-data","timestamp":"2023-01-01T12:00:00Z"},"timestamp":"2023-01-01T12:00:00Z"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.message.String()
			assert.JSONEq(t, tt.expected, result)
		})
	}
}

func TestHub_NotifyPeerJoined(t *testing.T) {
	logger := zap.NewNop()
	hub := NewHub(logger)

	// Create a test peer
	peer := &Peer{
		PeerID:         "test-peer-1",
		TenantID:       "test-tenant",
		DeviceID:       "test-device",
		ConnectionType: "quic",
		IsOnline:       true,
		LastSeen:       time.Now(),
		RegisteredAt:   time.Now(),
		MeshStatus:     "active",
	}

	// This test mainly verifies that the function doesn't panic
	// In a real test, we would need to set up WebSocket connections
	hub.NotifyPeerJoined("tenant-1", peer)
}

func TestHub_NotifyPeerLeft(t *testing.T) {
	logger := zap.NewNop()
	hub := NewHub(logger)

	// This test mainly verifies that the function doesn't panic
	// In a real test, we would need to set up WebSocket connections
	hub.NotifyPeerLeft("tenant-1", "test-peer-1")
}

func TestHub_NotifyPeerStatusUpdate(t *testing.T) {
	logger := zap.NewNop()
	hub := NewHub(logger)

	// This test mainly verifies that the function doesn't panic
	// In a real test, we would need to set up WebSocket connections
	hub.NotifyPeerStatusUpdate("tenant-1", "test-peer-1", true, 50, "connected")
}

func TestHub_SendRelayMessage(t *testing.T) {
	logger := zap.NewNop()
	hub := NewHub(logger)

	// Test with non-existent peer
	result := hub.SendRelayMessage("non-existent-peer", "test-message", "test-data")
	assert.False(t, result)

	// Test with existing peer (but no actual WebSocket connection)
	// In a real test, we would need to set up actual WebSocket connections
	conn := &WebSocketConnection{
		PeerID:   "test-peer-1",
		TenantID: "tenant-1",
		Send:     make(chan []byte, 1),
	}
	hub.Connections["test-peer-1"] = conn

	// This should return false because there's no actual WebSocket connection
	result = hub.SendRelayMessage("test-peer-1", "test-message", "test-data")
	assert.False(t, result)
}

func TestHub_BroadcastToTenant(t *testing.T) {
	logger := zap.NewNop()
	hub := NewHub(logger)

	// Create test connections
	conn1 := &WebSocketConnection{
		PeerID:   "test-peer-1",
		TenantID: "tenant-1",
		Send:     make(chan []byte, 1),
	}
	conn2 := &WebSocketConnection{
		PeerID:   "test-peer-2",
		TenantID: "tenant-1",
		Send:     make(chan []byte, 1),
	}
	conn3 := &WebSocketConnection{
		PeerID:   "test-peer-3",
		TenantID: "tenant-2",
		Send:     make(chan []byte, 1),
	}

	hub.Connections["test-peer-1"] = conn1
	hub.Connections["test-peer-2"] = conn2
	hub.Connections["test-peer-3"] = conn3

	// Create test message
	message := WebSocketMessage{
		Type:      "test-message",
		Data:      "test-data",
		Timestamp: time.Now(),
	}

	// This test mainly verifies that the function doesn't panic
	// In a real test, we would need to verify that messages are sent to the correct connections
	hub.BroadcastToTenant("tenant-1", message)
}

func TestHub_BroadcastToPeer(t *testing.T) {
	logger := zap.NewNop()
	hub := NewHub(logger)

	// Create test connection
	conn := &WebSocketConnection{
		PeerID:   "test-peer-1",
		TenantID: "tenant-1",
		Send:     make(chan []byte, 1),
	}
	hub.Connections["test-peer-1"] = conn

	// Create test message
	message := WebSocketMessage{
		Type:      "test-message",
		Data:      "test-data",
		Timestamp: time.Now(),
	}

	// This test mainly verifies that the function doesn't panic
	// In a real test, we would need to verify that the message is sent to the correct connection
	hub.BroadcastToPeer("test-peer-1", message)

	// Test with non-existent peer
	hub.BroadcastToPeer("non-existent-peer", message)
}
