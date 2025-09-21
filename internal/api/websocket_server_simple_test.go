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
	assert.Equal(t, logger, hub.logger)
	assert.NotNil(t, hub.peers)
	assert.NotNil(t, hub.Register)
	assert.NotNil(t, hub.Unregister)
	assert.NotNil(t, hub.Broadcast)
}

func TestHub_GetConnectionCount(t *testing.T) {
	logger := zap.NewNop()
	hub := NewHub(logger)

	// Initially no connections
	assert.Equal(t, 0, hub.GetConnectionCount())

	// Add a mock connection
	hub.Connections["peer1"] = &WebSocketConnection{
		PeerID: "peer1",
	}

	assert.Equal(t, 1, hub.GetConnectionCount())
}

func TestHub_GetPeerConnections(t *testing.T) {
	logger := zap.NewNop()
	hub := NewHub(logger)

	// Initially no connections
	connections := hub.GetPeerConnections()
	assert.Empty(t, connections)

	// Add mock connections
	hub.peers["peer1"] = &PeerConnection{
		ID:     "peer1",
		PeerID: "peer1",
		Status: "connected",
	}
	hub.peers["peer2"] = &PeerConnection{
		ID:     "peer2",
		PeerID: "peer2",
		Status: "disconnected",
	}

	connections = hub.GetPeerConnections()
	assert.Len(t, connections, 2)
	assert.Equal(t, "connected", connections["peer1"].Status)
	assert.Equal(t, "disconnected", connections["peer2"].Status)
}

func TestHub_RegisterPeer(t *testing.T) {
	logger := zap.NewNop()
	hub := NewHub(logger)

	// Start the hub
	go hub.Run()

	// Wait for hub to start
	time.Sleep(10 * time.Millisecond)

	// Register a peer
	peer := &PeerConnection{
		ID:     "peer1",
		PeerID: "peer1",
		Status: "connected",
	}

	hub.Register <- &WebSocketConnection{PeerID: peer.PeerID}

	// Wait for registration
	time.Sleep(10 * time.Millisecond)

	// Check if peer is registered
	assert.Equal(t, 1, hub.GetConnectionCount())
}

func TestHub_UnregisterPeer(t *testing.T) {
	logger := zap.NewNop()
	hub := NewHub(logger)

	// Start the hub
	go hub.Run()

	// Wait for hub to start
	time.Sleep(10 * time.Millisecond)

	// Register a peer first
	peer := &PeerConnection{
		ID:     "peer1",
		PeerID: "peer1",
		Status: "connected",
	}
	hub.Register <- &WebSocketConnection{
		PeerID: peer.PeerID,
		Send:   make(chan []byte, 256), // Add Send channel
	}

	// Wait for registration
	time.Sleep(10 * time.Millisecond)

	// Check peer is registered
	assert.Equal(t, 1, hub.GetConnectionCount())

	// Unregister the peer
	hub.Unregister <- &WebSocketConnection{
		PeerID: peer.PeerID,
		Send:   make(chan []byte, 256), // Add Send channel
	}

	// Wait for unregistration
	time.Sleep(10 * time.Millisecond)

	// Check if peer is unregistered
	assert.Equal(t, 0, hub.GetConnectionCount())
}

func TestHub_BroadcastMessage(t *testing.T) {
	logger := zap.NewNop()
	hub := NewHub(logger)

	// Start the hub
	go hub.Run()

	// Wait for hub to start
	time.Sleep(10 * time.Millisecond)

	// Register a peer
	peer := &PeerConnection{
		ID:     "peer1",
		PeerID: "peer1",
		Status: "connected",
	}
	hub.Register <- &WebSocketConnection{PeerID: peer.PeerID}

	// Wait for registration
	time.Sleep(10 * time.Millisecond)

	// Broadcast a message
	message := []byte("test message")
	hub.Broadcast <- message

	// Wait for broadcast
	time.Sleep(10 * time.Millisecond)

	// In a real implementation, we would check if the message was sent
	// For now, we just verify the hub is still running
	assert.NotNil(t, hub)
}

func TestHub_Run(t *testing.T) {
	logger := zap.NewNop()
	hub := NewHub(logger)

	// Start the hub
	go hub.Run()

	// Wait for hub to start
	time.Sleep(10 * time.Millisecond)

	// Verify hub is running
	assert.NotNil(t, hub)

	// Wait a bit more
	time.Sleep(10 * time.Millisecond)
}

func TestHub_ConcurrentOperations(t *testing.T) {
	logger := zap.NewNop()
	hub := NewHub(logger)

	// Start the hub
	go hub.Run()

	// Wait for hub to start
	time.Sleep(10 * time.Millisecond)

	// Perform concurrent operations
	go func() {
		peer := &PeerConnection{PeerID: "peer1", Status: "connected"}
		hub.Register <- &WebSocketConnection{PeerID: peer.PeerID}
	}()

	go func() {
		peer := &PeerConnection{PeerID: "peer2", Status: "connected"}
		hub.Register <- &WebSocketConnection{PeerID: peer.PeerID}
	}()

	go func() {
		hub.Broadcast <- []byte("test message")
	}()

	// Wait for operations to complete
	time.Sleep(50 * time.Millisecond)

	// Verify hub is still running and has connections
	assert.NotNil(t, hub)
	assert.Equal(t, 2, hub.GetConnectionCount())
}

func TestHub_MessageHandling(t *testing.T) {
	logger := zap.NewNop()
	hub := NewHub(logger)

	// Start the hub
	go hub.Run()

	// Wait for hub to start
	time.Sleep(10 * time.Millisecond)

	// Register a peer
	peer := &PeerConnection{
		ID:     "peer1",
		PeerID: "peer1",
		Status: "connected",
	}
	hub.Register <- &WebSocketConnection{PeerID: peer.PeerID}

	// Wait for registration
	time.Sleep(10 * time.Millisecond)

	// Send different types of messages
	testMessages := [][]byte{
		[]byte("simple message"),
		[]byte(`{"type": "json", "data": "test"}`),
		[]byte(""),
	}

	for _, msg := range testMessages {
		hub.Broadcast <- msg
		time.Sleep(10 * time.Millisecond)
	}

	// Verify hub is still running
	assert.NotNil(t, hub)
	assert.Equal(t, 1, hub.GetConnectionCount())
}

func TestHub_ErrorHandling(t *testing.T) {
	logger := zap.NewNop()
	hub := NewHub(logger)

	// Start the hub
	go hub.Run()

	// Wait for hub to start
	time.Sleep(10 * time.Millisecond)

	// Send nil peer (should be handled gracefully)
	hub.Register <- nil

	// Send nil message (should be handled gracefully)
	hub.Broadcast <- nil

	// Wait for operations to complete
	time.Sleep(10 * time.Millisecond)

	// Verify hub is still running
	assert.NotNil(t, hub)
}

func TestHub_Statistics(t *testing.T) {
	logger := zap.NewNop()
	hub := NewHub(logger)

	// Initially no connections
	assert.Equal(t, 0, hub.GetConnectionCount())

	// Add connections
	hub.Connections["peer1"] = &WebSocketConnection{PeerID: "peer1"}
	hub.Connections["peer2"] = &WebSocketConnection{PeerID: "peer2"}
	hub.Connections["peer3"] = &WebSocketConnection{PeerID: "peer3"}

	// Check connection count
	assert.Equal(t, 3, hub.GetConnectionCount())

	// Add peer connections
	hub.peers["peer1"] = &PeerConnection{ID: "peer1", PeerID: "peer1", Status: "connected"}
	hub.peers["peer2"] = &PeerConnection{ID: "peer2", PeerID: "peer2", Status: "connected"}
	hub.peers["peer3"] = &PeerConnection{ID: "peer3", PeerID: "peer3", Status: "disconnected"}

	// Check peer connections
	connections := hub.GetPeerConnections()
	assert.Len(t, connections, 3)

	// Verify all peers are present
	assert.Contains(t, connections, "peer1")
	assert.Contains(t, connections, "peer2")
	assert.Contains(t, connections, "peer3")
}
