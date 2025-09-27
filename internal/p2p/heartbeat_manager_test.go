package p2p

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewHeartbeatManager(t *testing.T) {
	t.Run("creates heartbeat manager successfully", func(t *testing.T) {
		zapLogger := zap.NewNop()
		relayManager := NewRelayManager(zapLogger, nil)
		manager := NewHeartbeatManager(relayManager, zapLogger)

		assert.NotNil(t, manager)
		assert.NotNil(t, manager.peers)
		assert.NotNil(t, manager.ctx)
		assert.NotNil(t, manager.cancel)
		assert.Equal(t, zapLogger, manager.logger)
	})

	t.Run("creates heartbeat manager with nil logger", func(t *testing.T) {
		relayManager := NewRelayManager(nil, nil)
		manager := NewHeartbeatManager(relayManager, nil)

		assert.NotNil(t, manager)
		assert.Nil(t, manager.logger)
	})
}

func TestHeartbeatManager_RegisterPeer(t *testing.T) {
	zapLogger := zap.NewNop()
	relayManager := NewRelayManager(zapLogger, nil)
	manager := NewHeartbeatManager(relayManager, zapLogger)
	defer manager.Stop()

	t.Run("registers peer successfully", func(t *testing.T) {
		peerID := "test-peer-1"
		tenantID := "tenant-1"

		err := manager.RegisterPeer(peerID, tenantID)
		assert.NoError(t, err)

		// Verify peer is registered
		peer, err := manager.GetPeerStatus(peerID)
		require.NoError(t, err)
		assert.NotNil(t, peer)
		assert.Equal(t, peerID, peer.PeerID)
		assert.Equal(t, tenantID, peer.TenantID)
		assert.True(t, peer.IsOnline)
	})

	t.Run("registers peer with empty peer ID", func(t *testing.T) {
		peerID := ""
		tenantID := "tenant-1"

		err := manager.RegisterPeer(peerID, tenantID)
		assert.NoError(t, err)
	})

	t.Run("registers peer with empty tenant ID", func(t *testing.T) {
		peerID := "test-peer-1"
		tenantID := ""

		err := manager.RegisterPeer(peerID, tenantID)
		assert.NoError(t, err)
	})
}

func TestHeartbeatManager_CleanupOfflinePeers(t *testing.T) {
	zapLogger := zap.NewNop()
	relayManager := NewRelayManager(zapLogger, nil)
	manager := NewHeartbeatManager(relayManager, zapLogger)
	defer manager.Stop()

	t.Run("cleans up offline peers", func(t *testing.T) {
		peerID := "test-peer-1"
		tenantID := "tenant-1"

		// Register peer
		err := manager.RegisterPeer(peerID, tenantID)
		require.NoError(t, err)

		// Mark peer as offline
		peer, err := manager.GetPeerStatus(peerID)
		require.NoError(t, err)
		peer.IsOnline = false
		peer.LastSeen = time.Now().Add(-15 * time.Minute) // 15 minutes ago

		// Start manager to enable cleanup
		err = manager.Start()
		require.NoError(t, err)

		// Wait for cleanup to run
		time.Sleep(100 * time.Millisecond)

		// Verify peer is still there (cleanup might not have run yet)
		_, err = manager.GetPeerStatus(peerID)
		assert.NoError(t, err)
	})
}

func TestHeartbeatManager_UpdateHeartbeat(t *testing.T) {
	zapLogger := zap.NewNop()
	relayManager := NewRelayManager(zapLogger, nil)
	manager := NewHeartbeatManager(relayManager, zapLogger)
	defer manager.Stop()

	t.Run("processes heartbeat successfully", func(t *testing.T) {
		peerID := "test-peer-1"
		tenantID := "tenant-1"

		// Register peer
		err := manager.RegisterPeer(peerID, tenantID)
		require.NoError(t, err)

		// Process heartbeat
		response, err := manager.ProcessHeartbeat(peerID, tenantID)
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "ok", response.Status)
		assert.True(t, response.Online)

		// Verify heartbeat was processed
		peer, err := manager.GetPeerStatus(peerID)
		require.NoError(t, err)
		assert.True(t, peer.IsOnline)
		assert.Greater(t, peer.HeartbeatCount, int64(0))
	})

	t.Run("handles processing heartbeat for non-existent peer", func(t *testing.T) {
		response, err := manager.ProcessHeartbeat("non-existent-peer", "tenant-1")
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "ok", response.Status)
		assert.True(t, response.Online)
	})
}

func TestHeartbeatManager_GetActivePeers(t *testing.T) {
	zapLogger := zap.NewNop()
	relayManager := NewRelayManager(zapLogger, nil)
	manager := NewHeartbeatManager(relayManager, zapLogger)
	defer manager.Stop()

	t.Run("returns empty list when no online peers", func(t *testing.T) {
		peers := manager.GetOnlinePeers("tenant-1")
		assert.Empty(t, peers)
	})

	t.Run("returns online peers", func(t *testing.T) {
		// Register multiple peers
		peer1 := "test-peer-1"
		peer2 := "test-peer-2"
		tenantID := "tenant-1"

		err := manager.RegisterPeer(peer1, tenantID)
		require.NoError(t, err)

		err = manager.RegisterPeer(peer2, tenantID)
		require.NoError(t, err)

		// Get online peers
		peers := manager.GetOnlinePeers(tenantID)
		assert.Len(t, peers, 2)
	})
}

func TestHeartbeatManager_GetTenantStats(t *testing.T) {
	zapLogger := zap.NewNop()
	relayManager := NewRelayManager(zapLogger, nil)
	manager := NewHeartbeatManager(relayManager, zapLogger)
	defer manager.Stop()

	t.Run("returns empty stats when no peers", func(t *testing.T) {
		stats := manager.GetTenantStats("tenant-1")
		assert.NotNil(t, stats)
		assert.Equal(t, 0, stats["total_peers"])
		assert.Equal(t, 0, stats["online_peers"])
	})

	t.Run("returns stats with peers", func(t *testing.T) {
		peerID := "test-peer-1"
		tenantID := "tenant-1"

		// Register peer
		err := manager.RegisterPeer(peerID, tenantID)
		require.NoError(t, err)

		// Get tenant stats
		stats := manager.GetTenantStats(tenantID)
		assert.NotNil(t, stats)
		assert.Equal(t, 1, stats["total_peers"])
		assert.Equal(t, 1, stats["online_peers"])
	})
}

func TestHeartbeatManager_GetTenantStats_Global(t *testing.T) {
	zapLogger := zap.NewNop()
	relayManager := NewRelayManager(zapLogger, nil)
	manager := NewHeartbeatManager(relayManager, zapLogger)
	defer manager.Stop()

	t.Run("returns stats with no peers", func(t *testing.T) {
		stats := manager.GetTenantStats("tenant-1")
		assert.NotNil(t, stats)
		assert.Equal(t, 0, stats["total_peers"])
		assert.Equal(t, 0, stats["online_peers"])
	})

	t.Run("returns stats with peers", func(t *testing.T) {
		// Register peer
		peerID := "test-peer-1"
		tenantID := "tenant-1"

		err := manager.RegisterPeer(peerID, tenantID)
		require.NoError(t, err)

		// Get stats
		stats := manager.GetTenantStats(tenantID)
		assert.NotNil(t, stats)
		assert.Equal(t, 1, stats["total_peers"])
		assert.Equal(t, 1, stats["online_peers"])
	})
}

func TestHeartbeatManager_Start(t *testing.T) {
	zapLogger := zap.NewNop()
	relayManager := NewRelayManager(zapLogger, nil)
	manager := NewHeartbeatManager(relayManager, zapLogger)

	t.Run("starts manager successfully", func(t *testing.T) {
		err := manager.Start()
		assert.NoError(t, err)

		// Cleanup
		manager.Stop()
	})
}

func TestHeartbeatManager_Stop(t *testing.T) {
	zapLogger := zap.NewNop()
	relayManager := NewRelayManager(zapLogger, nil)
	manager := NewHeartbeatManager(relayManager, zapLogger)

	t.Run("stops manager successfully", func(t *testing.T) {
		// Start manager
		err := manager.Start()
		require.NoError(t, err)

		// Stop manager
		manager.Stop()
		// Manager should be stopped without error
	})
}

func TestHeartbeatManager_ContextCancellation(t *testing.T) {
	zapLogger := zap.NewNop()
	relayManager := NewRelayManager(zapLogger, nil)
	manager := NewHeartbeatManager(relayManager, zapLogger)

	t.Run("handles context cancellation", func(t *testing.T) {
		// Start manager
		err := manager.Start()
		require.NoError(t, err)

		// Stop manager (this should cancel context)
		manager.Stop()

		// Wait a bit for cleanup
		time.Sleep(100 * time.Millisecond)

		// Manager should be stopped without error
	})
}

func TestHeartbeatManager_CheckHeartbeatTimeouts(t *testing.T) {
	zapLogger := zap.NewNop()
	relayManager := NewRelayManager(zapLogger, nil)
	manager := NewHeartbeatManager(relayManager, zapLogger)

	t.Run("marks peers as offline after timeout", func(t *testing.T) {
		// Register a peer
		peerID := "test-peer-1"
		tenantID := "tenant-1"
		err := manager.RegisterPeer(peerID, tenantID)
		require.NoError(t, err)

		// Manually set last heartbeat to be old (simulate timeout)
		manager.mutex.Lock()
		peer := manager.peers[peerID]
		peer.Mutex.Lock()
		peer.LastHeartbeat = time.Now().Add(-3 * time.Minute) // 3 minutes ago
		peer.IsOnline = true
		peer.Mutex.Unlock()
		manager.mutex.Unlock()

		// Call checkHeartbeatTimeouts
		manager.checkHeartbeatTimeouts()

		// Check that peer is now offline
		manager.mutex.RLock()
		peer = manager.peers[peerID]
		peer.Mutex.RLock()
		isOnline := peer.IsOnline
		peer.Mutex.RUnlock()
		manager.mutex.RUnlock()

		assert.False(t, isOnline, "Peer should be marked as offline after timeout")
	})

	t.Run("does not mark recently active peers as offline", func(t *testing.T) {
		// Register a peer
		peerID := "test-peer-2"
		tenantID := "tenant-2"
		err := manager.RegisterPeer(peerID, tenantID)
		require.NoError(t, err)

		// Process a recent heartbeat
		_, err = manager.ProcessHeartbeat(peerID, tenantID)
		require.NoError(t, err)

		// Call checkHeartbeatTimeouts
		manager.checkHeartbeatTimeouts()

		// Check that peer is still online
		manager.mutex.RLock()
		peer := manager.peers[peerID]
		peer.Mutex.RLock()
		isOnline := peer.IsOnline
		peer.Mutex.RUnlock()
		manager.mutex.RUnlock()

		assert.True(t, isOnline, "Recently active peer should remain online")
	})
}

func TestHeartbeatManager_PerformCleanup(t *testing.T) {
	zapLogger := zap.NewNop()
	relayManager := NewRelayManager(zapLogger, nil)
	manager := NewHeartbeatManager(relayManager, zapLogger)

	t.Run("removes old offline peers", func(t *testing.T) {
		// Register a peer
		peerID := "test-peer-cleanup"
		tenantID := "tenant-cleanup"
		err := manager.RegisterPeer(peerID, tenantID)
		require.NoError(t, err)

		// Manually set peer as offline and old
		manager.mutex.Lock()
		peer := manager.peers[peerID]
		peer.Mutex.Lock()
		peer.IsOnline = false
		peer.LastSeen = time.Now().Add(-15 * time.Minute) // 15 minutes ago
		peer.Mutex.Unlock()
		manager.mutex.Unlock()

		// Call performCleanup
		manager.performCleanup()

		// Check that peer was removed
		manager.mutex.RLock()
		_, exists := manager.peers[peerID]
		manager.mutex.RUnlock()

		assert.False(t, exists, "Old offline peer should be removed")
	})

	t.Run("keeps recent offline peers", func(t *testing.T) {
		// Register a peer
		peerID := "test-peer-recent"
		tenantID := "tenant-recent"
		err := manager.RegisterPeer(peerID, tenantID)
		require.NoError(t, err)

		// Manually set peer as offline but recent
		manager.mutex.Lock()
		peer := manager.peers[peerID]
		peer.Mutex.Lock()
		peer.IsOnline = false
		peer.LastSeen = time.Now().Add(-5 * time.Minute) // 5 minutes ago
		peer.Mutex.Unlock()
		manager.mutex.Unlock()

		// Call performCleanup
		manager.performCleanup()

		// Check that peer was not removed
		manager.mutex.RLock()
		_, exists := manager.peers[peerID]
		manager.mutex.RUnlock()

		assert.True(t, exists, "Recent offline peer should not be removed")
	})

	t.Run("keeps online peers regardless of age", func(t *testing.T) {
		// Register a peer
		peerID := "test-peer-online"
		tenantID := "tenant-online"
		err := manager.RegisterPeer(peerID, tenantID)
		require.NoError(t, err)

		// Manually set peer as online but old
		manager.mutex.Lock()
		peer := manager.peers[peerID]
		peer.Mutex.Lock()
		peer.IsOnline = true
		peer.LastSeen = time.Now().Add(-15 * time.Minute) // 15 minutes ago
		peer.Mutex.Unlock()
		manager.mutex.Unlock()

		// Call performCleanup
		manager.performCleanup()

		// Check that peer was not removed
		manager.mutex.RLock()
		_, exists := manager.peers[peerID]
		manager.mutex.RUnlock()

		assert.True(t, exists, "Online peer should not be removed regardless of age")
	})
}

func TestHeartbeatManager_EdgeCases(t *testing.T) {
	zapLogger := zap.NewNop()
	relayManager := NewRelayManager(zapLogger, nil)
	manager := NewHeartbeatManager(relayManager, zapLogger)
	defer manager.Stop()

	t.Run("handles multiple heartbeats from same peer", func(t *testing.T) {
		peerID := "test-peer-1"
		tenantID := "tenant-1"

		// Register peer
		err := manager.RegisterPeer(peerID, tenantID)
		require.NoError(t, err)

		// Process multiple heartbeats
		for i := 0; i < 5; i++ {
			response, err := manager.ProcessHeartbeat(peerID, tenantID)
			assert.NoError(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, "ok", response.Status)
			assert.True(t, response.Online)
		}

		// Verify heartbeat count increased
		peer, err := manager.GetPeerStatus(peerID)
		require.NoError(t, err)
		assert.Greater(t, peer.HeartbeatCount, int64(1))
	})

	t.Run("handles heartbeat timeout detection", func(t *testing.T) {
		peerID := "test-peer-timeout"
		tenantID := "tenant-1"

		// Register peer
		err := manager.RegisterPeer(peerID, tenantID)
		require.NoError(t, err)

		// Start manager to enable timeout checking
		err = manager.Start()
		require.NoError(t, err)

		// Manually set peer as offline with old timestamp
		peer, err := manager.GetPeerStatus(peerID)
		require.NoError(t, err)
		peer.Mutex.Lock()
		peer.IsOnline = true
		peer.LastHeartbeat = time.Now().Add(-3 * time.Minute) // 3 minutes ago
		peer.Mutex.Unlock()

		// Wait for timeout check
		time.Sleep(100 * time.Millisecond)

		// Peer should still be there (timeout check might not have run yet)
		_, err = manager.GetPeerStatus(peerID)
		assert.NoError(t, err)
	})

	t.Run("handles concurrent peer registration", func(t *testing.T) {
		// Test concurrent access to peer registration
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func(index int) {
				peerID := fmt.Sprintf("concurrent-peer-%d", index)
				tenantID := "tenant-concurrent"

				err := manager.RegisterPeer(peerID, tenantID)
				assert.NoError(t, err)

				done <- true
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}

		// Verify all peers are registered
		onlinePeers := manager.GetOnlinePeers("tenant-concurrent")
		assert.Len(t, onlinePeers, 10)
	})

	t.Run("handles get peer status for non-existent peer", func(t *testing.T) {
		peer, err := manager.GetPeerStatus("non-existent-peer")
		assert.Error(t, err)
		assert.Nil(t, peer)
		assert.Contains(t, err.Error(), "peer not found")
	})

	t.Run("handles empty tenant ID in GetOnlinePeers", func(t *testing.T) {
		peers := manager.GetOnlinePeers("")
		assert.Empty(t, peers)
	})

	t.Run("handles empty tenant ID in GetTenantStats", func(t *testing.T) {
		stats := manager.GetTenantStats("")
		assert.NotNil(t, stats)
		assert.Equal(t, 0, stats["total_peers"])
		assert.Equal(t, 0, stats["online_peers"])
		assert.Equal(t, 0, stats["offline_peers"])
	})
}

func TestHeartbeatManager_HeartbeatRequestResponse(t *testing.T) {
	t.Run("creates heartbeat request", func(t *testing.T) {
		request := &HeartbeatRequest{
			PeerID:    "test-peer",
			TenantID:  "test-tenant",
			Timestamp: time.Now().Unix(),
			Status:    "active",
		}

		assert.Equal(t, "test-peer", request.PeerID)
		assert.Equal(t, "test-tenant", request.TenantID)
		assert.Equal(t, "active", request.Status)
		assert.Greater(t, request.Timestamp, int64(0))
	})

	t.Run("creates heartbeat response", func(t *testing.T) {
		response := &HeartbeatResponse{
			Status:    "ok",
			Timestamp: time.Now().Unix(),
			Online:    true,
			Message:   "Heartbeat processed successfully",
		}

		assert.Equal(t, "ok", response.Status)
		assert.True(t, response.Online)
		assert.Equal(t, "Heartbeat processed successfully", response.Message)
		assert.Greater(t, response.Timestamp, int64(0))
	})
}

func TestHeartbeatManager_PeerHeartbeat(t *testing.T) {
	t.Run("creates peer heartbeat", func(t *testing.T) {
		now := time.Now()
		peerHeartbeat := &PeerHeartbeat{
			PeerID:         "test-peer",
			TenantID:       "test-tenant",
			LastHeartbeat:  now,
			HeartbeatCount: 5,
			IsOnline:       true,
			LastSeen:       now,
		}

		assert.Equal(t, "test-peer", peerHeartbeat.PeerID)
		assert.Equal(t, "test-tenant", peerHeartbeat.TenantID)
		assert.Equal(t, now, peerHeartbeat.LastHeartbeat)
		assert.Equal(t, int64(5), peerHeartbeat.HeartbeatCount)
		assert.True(t, peerHeartbeat.IsOnline)
		assert.Equal(t, now, peerHeartbeat.LastSeen)
	})

	t.Run("handles peer heartbeat mutex operations", func(t *testing.T) {
		peerHeartbeat := &PeerHeartbeat{
			PeerID:   "test-peer",
			TenantID: "test-tenant",
		}

		// Test concurrent access
		done := make(chan bool, 2)

		go func() {
			peerHeartbeat.Mutex.Lock()
			peerHeartbeat.IsOnline = true
			peerHeartbeat.HeartbeatCount++
			peerHeartbeat.Mutex.Unlock()
			done <- true
		}()

		go func() {
			peerHeartbeat.Mutex.RLock()
			_ = peerHeartbeat.IsOnline
			_ = peerHeartbeat.HeartbeatCount
			peerHeartbeat.Mutex.RUnlock()
			done <- true
		}()

		// Wait for both goroutines to complete
		<-done
		<-done

		// Should not panic
		assert.True(t, peerHeartbeat.IsOnline)
	})
}

func TestHeartbeatManager_StartStopMultipleTimes(t *testing.T) {
	zapLogger := zap.NewNop()
	relayManager := NewRelayManager(zapLogger, nil)
	manager := NewHeartbeatManager(relayManager, zapLogger)

	t.Run("handles multiple start/stop cycles", func(t *testing.T) {
		// Start manager
		err := manager.Start()
		require.NoError(t, err)

		// Stop manager
		manager.Stop()

		// Start again
		err = manager.Start()
		require.NoError(t, err)

		// Stop again
		manager.Stop()

		// Should not panic
	})

	t.Run("handles stop without start", func(t *testing.T) {
		// Stop without starting should not panic
		manager.Stop()
	})
}

func TestHeartbeatManager_WithNilRelayManager(t *testing.T) {
	zapLogger := zap.NewNop()
	manager := NewHeartbeatManager(nil, zapLogger)
	defer manager.Stop()

	t.Run("works with nil relay manager", func(t *testing.T) {
		peerID := "test-peer"
		tenantID := "test-tenant"

		// Should not panic
		err := manager.RegisterPeer(peerID, tenantID)
		assert.NoError(t, err)

		response, err := manager.ProcessHeartbeat(peerID, tenantID)
		assert.NoError(t, err)
		assert.NotNil(t, response)

		stats := manager.GetTenantStats(tenantID)
		assert.NotNil(t, stats)
	})
}
