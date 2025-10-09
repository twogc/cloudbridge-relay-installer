package api

import (
	"fmt"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/internal/p2p"
)

// PeerStoreAdapter адаптирует api.PeerStoreInterface к p2p.PeerStoreInterface
type PeerStoreAdapter struct {
	apiStore PeerStoreInterface
}

// NewPeerStoreAdapter создает новый адаптер
func NewPeerStoreAdapter(apiStore PeerStoreInterface) *PeerStoreAdapter {
	return &PeerStoreAdapter{
		apiStore: apiStore,
	}
}

// Store сохраняет пира
func (psa *PeerStoreAdapter) Store(tenantID, peerID string, peer *p2p.Peer) {
	// api.PeerStore уже работает с p2p.Peer, просто передаем напрямую
	psa.apiStore.Store(tenantID, peerID, peer)
}

// Get получает пира
func (psa *PeerStoreAdapter) Get(tenantID, peerID string) (*p2p.Peer, bool) {
	// api.PeerStore уже работает с p2p.Peer, просто возвращаем напрямую
	return psa.apiStore.Get(tenantID, peerID)
}

// Update обновляет пира
func (psa *PeerStoreAdapter) Update(tenantID, peerID string, peer *p2p.Peer) {
	psa.Store(tenantID, peerID, peer)
}

// GetAll получает всех пиров для тенанта
func (psa *PeerStoreAdapter) GetAll(tenantID string) []*p2p.Peer {
	// api.PeerStore уже работает с p2p.Peer, просто возвращаем напрямую
	peers := psa.apiStore.GetAll(tenantID)
	// Debug log для отладки
	fmt.Printf("DEBUG: PeerStoreAdapter.GetAll called for tenantID=%s, found %d peers\n", tenantID, len(peers))

	// Логируем детали каждого пира
	for i, peer := range peers {
		fmt.Printf("DEBUG: PeerStoreAdapter found peer[%d]: ID=%s, TenantID=%s, IsOnline=%v, AllocatedIP=%s\n",
			i, peer.ID, peer.TenantID, peer.IsOnline, peer.AllocatedIP)
	}

	return peers
}

// IsPeerOnline проверяет, онлайн ли пир
func (psa *PeerStoreAdapter) IsPeerOnline(peerID string) bool {
	return psa.apiStore.IsPeerOnline(peerID)
}

// GetLastSeen возвращает время последнего контакта с пиром
func (psa *PeerStoreAdapter) GetLastSeen(peerID string) time.Time {
	return psa.apiStore.GetLastSeen(peerID)
}
