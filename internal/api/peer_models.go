package api

import (
	"time"
)

// Peer представляет зарегистрированный пир
type Peer struct {
	ID                string           `json:"id" db:"id"`
	PeerID            string           `json:"peer_id" db:"peer_id"`
	TenantID          string           `json:"tenant_id" db:"tenant_id"`
	OrgID             string           `json:"org_id" db:"org_id"`
	DeviceID          string           `json:"device_id" db:"device_id"`
	ConnectionType    string           `json:"connection_type" db:"connection_type"`
	RelaySessionID    string           `json:"relay_session_id" db:"relay_session_id"`
	WireGuardConfig   *WireGuardConfig `json:"wireguard_config" db:"wireguard_config"`
	MeshConfig        *MeshConfig      `json:"mesh_config" db:"mesh_config"`
	PeerWhitelist     *PeerWhitelist   `json:"peer_whitelist" db:"peer_whitelist"`
	NetworkConfig     *NetworkConfig   `json:"network_config" db:"network_config"`
	IsOnline          bool             `json:"is_online" db:"is_online"`
	LastSeen          time.Time        `json:"last_seen" db:"last_seen"`
	RegisteredAt      time.Time        `json:"registered_at" db:"registered_at"`
	ActiveConnections int              `json:"active_connections" db:"active_connections"`
	MeshStatus        string           `json:"mesh_status" db:"mesh_status"`
	LatencyMS         int              `json:"latency_ms" db:"latency_ms"`
}

// WireGuardConfig конфигурация WireGuard
type WireGuardConfig struct {
	PublicKey  string   `json:"public_key"`
	Endpoint   string   `json:"endpoint"`
	ListenPort int      `json:"listen_port"`
	AllowedIPs []string `json:"allowed_ips"`
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

// PeerConnection представляет соединение между пирами
type PeerConnection struct {
	ID               string    `json:"id" db:"id"`
	FromPeerID       string    `json:"from_peer_id" db:"from_peer_id"`
	ToPeerID         string    `json:"to_peer_id" db:"to_peer_id"`
	ConnectionType   string    `json:"connection_type" db:"connection_type"`
	IsActive         bool      `json:"is_active" db:"is_active"`
	EstablishedAt    time.Time `json:"established_at" db:"established_at"`
	LastActivity     time.Time `json:"last_activity" db:"last_activity"`
	BytesTransferred int64     `json:"bytes_transferred" db:"bytes_transferred"`
	LatencyMS        int       `json:"latency_ms" db:"latency_ms"`
}

// PeerRegistrationRequest запрос на регистрацию пира
type PeerRegistrationRequest struct {
	PeerID          string           `json:"peer_id" binding:"required"`
	TenantID        string           `json:"tenant_id" binding:"required"`
	OrgID           string           `json:"org_id" binding:"required"`
	DeviceID        string           `json:"device_id" binding:"required"`
	ConnectionType  string           `json:"connection_type" binding:"required"`
	WireGuardConfig *WireGuardConfig `json:"wireguard_config" binding:"required"`
	MeshConfig      *MeshConfig      `json:"mesh_config"`
	PeerWhitelist   *PeerWhitelist   `json:"peer_whitelist"`
	NetworkConfig   *NetworkConfig   `json:"network_config"`
}

// PeerStatusUpdateRequest запрос на обновление статуса пира
type PeerStatusUpdateRequest struct {
	IsOnline          bool   `json:"is_online"`
	LatencyMS         int    `json:"latency_ms"`
	ActiveConnections int    `json:"active_connections"`
	MeshStatus        string `json:"mesh_status"`
}

// PeerDiscoveryResponse ответ на discovery запрос
type PeerDiscoveryResponse struct {
	Success  bool    `json:"success"`
	TenantID string  `json:"tenant_id"`
	Peers    []*Peer `json:"peers"`
}

// PeerRegistrationResponse ответ на регистрацию пира
type PeerRegistrationResponse struct {
	Success           bool      `json:"success"`
	PeerID            string    `json:"peer_id"`
	RegisteredAt      time.Time `json:"registered_at"`
	DiscoveryEndpoint string    `json:"discovery_endpoint"`
}

// PeerStatusResponse ответ на обновление статуса
type PeerStatusResponse struct {
	Success   bool      `json:"success"`
	UpdatedAt time.Time `json:"updated_at"`
}

// PeerInfoResponse ответ с информацией о пире
type PeerInfoResponse struct {
	Success bool  `json:"success"`
	Peer    *Peer `json:"peer"`
}

// RelayRouteRequest запрос на маршрутизацию через relay
type RelayRouteRequest struct {
	FromPeer    string `json:"from_peer" binding:"required"`
	ToPeer      string `json:"to_peer" binding:"required"`
	Data        string `json:"data" binding:"required"`
	MessageType string `json:"message_type" binding:"required"`
}

// RelayRouteResponse ответ на маршрутизацию
type RelayRouteResponse struct {
	Success   bool      `json:"success"`
	RoutedAt  time.Time `json:"routed_at"`
	LatencyMS int64     `json:"latency_ms"`
}

// RelayStatsResponse статистика маршрутизации
type RelayStatsResponse struct {
	Success  bool        `json:"success"`
	TenantID string      `json:"tenant_id"`
	Stats    *RelayStats `json:"stats"`
}

// RelayStats статистика relay
type RelayStats struct {
	TotalPeers        int   `json:"total_peers"`
	OnlinePeers       int   `json:"online_peers"`
	DirectConnections int   `json:"direct_connections"`
	RelayConnections  int   `json:"relay_connections"`
	TotalTrafficBytes int64 `json:"total_traffic_bytes"`
	AverageLatencyMS  int   `json:"average_latency_ms"`
}

// WebSocketMessage сообщение WebSocket
type WebSocketMessage struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// PeerJoinedData данные о подключившемся пире
type PeerJoinedData struct {
	PeerID          string           `json:"peer_id"`
	WireGuardConfig *WireGuardConfig `json:"wireguard_config"`
	Timestamp       time.Time        `json:"timestamp"`
}

// PeerLeftData данные об отключившемся пире
type PeerLeftData struct {
	PeerID    string    `json:"peer_id"`
	Timestamp time.Time `json:"timestamp"`
}

// PeerStatusUpdateData данные об обновлении статуса пира
type PeerStatusUpdateData struct {
	PeerID     string    `json:"peer_id"`
	IsOnline   bool      `json:"is_online"`
	LatencyMS  int       `json:"latency_ms"`
	MeshStatus string    `json:"mesh_status"`
	Timestamp  time.Time `json:"timestamp"`
}

// RelayMessageData данные сообщения через relay
type RelayMessageData struct {
	Type        string    `json:"type"`
	MessageType string    `json:"message_type"`
	Data        string    `json:"data"`
	Timestamp   time.Time `json:"timestamp"`
}

// ErrorResponse ответ с ошибкой
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// SuccessResponse базовый успешный ответ
type SuccessResponse struct {
	Success bool `json:"success"`
}

// RelayAssistedRegistrationRequest запрос на регистрацию пира для relay-assisted режима
type RelayAssistedRegistrationRequest struct {
	PublicKey  string   `json:"public_key" binding:"required"`
	AllowedIPs []string `json:"allowed_ips" binding:"required"`
}

// RelayAssistedRegistrationResponse ответ на регистрацию пира для relay-assisted режима
type RelayAssistedRegistrationResponse struct {
	Success        bool      `json:"success"`
	PeerID         string    `json:"peer_id"`
	RelaySessionID string    `json:"relay_session_id"`
	RegisteredAt   time.Time `json:"registered_at"`
}

// RelayAssistedStatusUpdateRequest запрос на обновление статуса пира для relay-assisted режима
type RelayAssistedStatusUpdateRequest struct {
	RelaySessionID string `json:"relay_session_id" binding:"required"`
	Status         string `json:"status" binding:"required"`
}

// RelayAssistedDiscoveryResponse ответ на discovery запрос для relay-assisted режима
type RelayAssistedDiscoveryResponse struct {
	Success  bool    `json:"success"`
	TenantID string  `json:"tenant_id"`
	Peers    []*Peer `json:"peers"`
}
