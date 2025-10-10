package database

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// JSONB is a generic JSONB helper that handles nil receivers and string/[]byte conversion
type JSONB[T any] struct {
	V T
}

// Value implements driver.Valuer for database storage
func (j JSONB[T]) Value() (driver.Value, error) {
	return json.Marshal(j.V)
}

// Scan implements sql.Scanner for database retrieval
func (j *JSONB[T]) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return fmt.Errorf("JSONB: unsupported Scan type %T", value)
	}

	if len(data) == 0 {
		return nil
	}

	return json.Unmarshal(data, &j.V)
}

// PeerStatus represents the status of a peer
type PeerStatus string

const (
	PeerStatusOnline   PeerStatus = "online"
	PeerStatusOffline  PeerStatus = "offline"
	PeerStatusDegraded PeerStatus = "degraded"
)

// Peer represents a peer in the database
type Peer struct {
	ID                uuid.UUID            `json:"id" db:"id"`
	PeerID            string               `json:"peer_id" db:"peer_id"`
	TenantID          string               `json:"tenant_id" db:"tenant_id"`
	OrgID             string               `json:"org_id" db:"org_id"`
	DeviceID          string               `json:"device_id" db:"device_id"`
	ConnectionType    string               `json:"connection_type" db:"connection_type"`
	QUICConfig        JSONB[QUICConfig]    `json:"quic_config" db:"quic_config"`
	MeshConfig        JSONB[MeshConfig]    `json:"mesh_config" db:"mesh_config"`
	PeerWhitelist     JSONB[PeerWhitelist] `json:"peer_whitelist" db:"peer_whitelist"`
	NetworkConfig     JSONB[NetworkConfig] `json:"network_config" db:"network_config"`
	IsOnline          bool                 `json:"is_online" db:"is_online"`
	LastSeen          time.Time            `json:"last_seen" db:"last_seen"`
	RegisteredAt      time.Time            `json:"registered_at" db:"registered_at"`
	ActiveConnections int                  `json:"active_connections" db:"active_connections"`
	Status            PeerStatus           `json:"status" db:"status"`
	LatencyMS         *int                 `json:"latency_ms" db:"latency_ms"` // NULLable for unknown latency
	Version           int                  `json:"version" db:"version"`       // Optimistic locking
	CreatedAt         time.Time            `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time            `json:"updated_at" db:"updated_at"`
}

// PeerConnection represents a connection between peers
type PeerConnection struct {
	ID               uuid.UUID `json:"id" db:"id"`
	FromPeerID       string    `json:"from_peer_id" db:"from_peer_id"`
	ToPeerID         string    `json:"to_peer_id" db:"to_peer_id"`
	TenantID         string    `json:"tenant_id" db:"tenant_id"` // For faster tenant-scoped queries
	ConnectionType   string    `json:"connection_type" db:"connection_type"`
	IsActive         bool      `json:"is_active" db:"is_active"`
	EstablishedAt    time.Time `json:"established_at" db:"established_at"`
	LastActivity     time.Time `json:"last_activity" db:"last_activity"`
	BytesTransferred int64     `json:"bytes_transferred" db:"bytes_transferred"`
	LatencyMS        *int      `json:"latency_ms" db:"latency_ms"` // NULLable for unknown latency
	Version          int       `json:"version" db:"version"`       // Optimistic locking
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

// PeerMetric represents a metric for a peer
type PeerMetric struct {
	ID          uuid.UUID         `json:"id" db:"id"`
	PeerID      string            `json:"peer_id" db:"peer_id"`
	MetricType  string            `json:"metric_type" db:"metric_type"`
	MetricValue float64           `json:"metric_value" db:"metric_value"`
	MetricUnit  string            `json:"metric_unit" db:"metric_unit"`
	Labels      map[string]string `json:"labels" db:"labels"`
	Timestamp   time.Time         `json:"timestamp" db:"timestamp"`
	CreatedAt   time.Time         `json:"created_at" db:"created_at"`
}

// WebSocketConnection represents a WebSocket connection
type WebSocketConnection struct {
	ID           uuid.UUID `json:"id" db:"id"`
	PeerID       string    `json:"peer_id" db:"peer_id"`
	TenantID     string    `json:"tenant_id" db:"tenant_id"`
	ConnectionID string    `json:"connection_id" db:"connection_id"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	ConnectedAt  time.Time `json:"connected_at" db:"connected_at"`
	LastPing     time.Time `json:"last_ping" db:"last_ping"`
	UserAgent    string    `json:"user_agent" db:"user_agent"`
	RemoteAddr   string    `json:"remote_addr" db:"remote_addr"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// RelayMessage represents a relay message for audit
type RelayMessage struct {
	ID           uuid.UUID `json:"id" db:"id"`
	FromPeerID   string    `json:"from_peer_id" db:"from_peer_id"`
	ToPeerID     string    `json:"to_peer_id" db:"to_peer_id"`
	MessageType  string    `json:"message_type" db:"message_type"`
	MessageSize  int       `json:"message_size" db:"message_size"`
	Success      bool      `json:"success" db:"success"`
	LatencyMS    int       `json:"latency_ms" db:"latency_ms"`
	ErrorMessage string    `json:"error_message" db:"error_message"`
	Timestamp    time.Time `json:"timestamp" db:"timestamp"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// QUICConfig represents QUIC configuration
type QUICConfig struct {
	ServerName string   `json:"server_name"`
	Endpoint   string   `json:"endpoint"`
	ListenPort int      `json:"listen_port"`
	AllowedIPs []string `json:"allowed_ips"`
}

// QUICConfig methods are now handled by JSONB[QUICConfig]

// MeshConfig represents mesh network configuration
type MeshConfig struct {
	AutoDiscovery bool   `json:"auto_discovery"`
	Routing       string `json:"routing"`
	Encryption    string `json:"encryption"`
}

// MeshConfig methods are now handled by JSONB[MeshConfig]

// PeerWhitelist represents peer whitelist configuration
type PeerWhitelist struct {
	AllowedPeers []string `json:"allowed_peers"`
	AutoApprove  bool     `json:"auto_approve"`
	MaxPeers     int      `json:"max_peers"`
}

// PeerWhitelist methods are now handled by JSONB[PeerWhitelist]

// NetworkConfig represents network configuration
type NetworkConfig struct {
	Subnet string   `json:"subnet"`
	DNS    []string `json:"dns"`
	MTU    int      `json:"mtu"`
}

// NetworkConfig methods are now handled by JSONB[NetworkConfig]

// TenantNetwork represents a network allocated to a tenant for IPAM
type TenantNetwork struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	TenantID      string     `json:"tenant_id" db:"tenant_id"`
	NetworkCIDR   string     `json:"network_cidr" db:"network_cidr"`
	SubnetCounter int        `json:"subnet_counter" db:"subnet_counter"`
	AllocatedAt   time.Time  `json:"allocated_at" db:"allocated_at"`
	IsActive      bool       `json:"is_active" db:"is_active"`
	DeactivatedAt *time.Time `json:"deactivated_at" db:"deactivated_at"` // For audit trail
	Version       int        `json:"version" db:"version"`               // Optimistic locking
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

// PeerIPAllocation represents an IP address allocated to a peer within a tenant
type PeerIPAllocation struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	TenantID    string     `json:"tenant_id" db:"tenant_id"`
	PeerID      string     `json:"peer_id" db:"peer_id"`
	AllocatedIP string     `json:"allocated_ip" db:"allocated_ip"`
	AllocatedAt time.Time  `json:"allocated_at" db:"allocated_at"`
	ReleasedAt  *time.Time `json:"released_at" db:"released_at"`
	IsActive    bool       `json:"is_active" db:"is_active"`
	Version     int        `json:"version" db:"version"` // Optimistic locking for IPAM operations
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// Labels represents metric labels
type Labels map[string]string

// Value implements driver.Valuer for database storage
func (l Labels) Value() (driver.Value, error) {
	if l == nil {
		return nil, nil
	}
	return json.Marshal(l)
}

// Scan implements sql.Scanner for database retrieval
func (l *Labels) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, l)
}
