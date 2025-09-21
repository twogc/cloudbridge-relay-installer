package p2p

import (
	"time"
)

// Peer представляет пира в P2P сети
type Peer struct {
	ID                string
	PeerID            string
	TenantID          string
	OrgID             string
	DeviceID          string
	ConnectionType    string
	RelaySessionID    string
	Endpoint          string
	QuicConfig        *QuicConfig
	MeshConfig        *MeshConfig
	PeerWhitelist     *PeerWhitelist
	NetworkConfig     *NetworkConfig
	ICECandidates     []*ICECandidate
	STUNServers       []string
	MASQUESupport     bool
	IsOnline          bool
	LastSeen          time.Time
	RegisteredAt      time.Time
	ActiveConnections int
	MeshStatus        string
	LatencyMS         int
	AllocatedIP       string
	Status            string
}

// QuicConfig конфигурация QUIC
type QuicConfig struct {
	Endpoint            string `json:"endpoint"`
	MaxStreams          int    `json:"max_streams"`
	HandshakeTimeout    string `json:"handshake_timeout"`
	IdleTimeout         string `json:"idle_timeout"`
	ConnectionMigration bool   `json:"connection_migration"`
	Multiplexing        bool   `json:"multiplexing"`
}

// MeshConfig конфигурация mesh сети
type MeshConfig struct {
	AutoDiscovery bool   `json:"auto_discovery"`
	Routing       string `json:"routing"`
	Encryption    string `json:"encryption"`
}

// PeerWhitelist список разрешенных пиров
type PeerWhitelist struct {
	AllowedPeers []string `json:"allowed_peers"`
	AutoApprove  bool     `json:"auto_approve"`
	MaxPeers     int      `json:"max_peers"`
}

// NetworkConfig сетевая конфигурация
type NetworkConfig struct {
	Subnet string   `json:"subnet"`
	DNS    []string `json:"dns"`
	MTU    int      `json:"mtu"`
}

// ICECandidate представляет ICE candidate для NAT traversal
type ICECandidate struct {
	Foundation  string `json:"foundation"`
	Component   int    `json:"component"`
	Transport   string `json:"transport"`
	Priority    uint32 `json:"priority"`
	Address     string `json:"address"`
	Port        int    `json:"port"`
	Type        string `json:"type"` // "host", "srflx", "prflx", "relay"
	RelatedAddr string `json:"related_address,omitempty"`
	RelatedPort int    `json:"related_port,omitempty"`
	Username    string `json:"username,omitempty"`
	Password    string `json:"password,omitempty"`
}

// DiscoveredPeer представляет обнаруженного пира
type DiscoveredPeer struct {
	ID           string    `json:"id"`
	PeerID       string    `json:"peer_id"`
	TenantID     string    `json:"tenant_id"`
	Endpoint     string    `json:"endpoint"`
	Protocol     string    `json:"protocol"`
	LastSeen     time.Time `json:"last_seen"`
	Latency      int       `json:"latency_ms"`
	IsOnline     bool      `json:"is_online"`
	Capabilities []string  `json:"capabilities"`
	Region       string    `json:"region"`
	TrustLevel   string    `json:"trust_level"`
}
