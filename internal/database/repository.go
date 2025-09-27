package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// Repository handles database operations for P2P functionality
type Repository struct {
	db     *sqlx.DB
	logger *zap.Logger
}

// NewRepository creates a new database repository
func NewRepository(db *sqlx.DB, logger *zap.Logger) *Repository {
	return &Repository{
		db:     db,
		logger: logger,
	}
}

// PeerRepository methods

// CreatePeer creates a new peer in the database
func (r *Repository) CreatePeer(peer *Peer) error {
	query := `
		INSERT INTO peers (
			peer_id, tenant_id, org_id, device_id, connection_type,
			quic_config, mesh_config, peer_whitelist, network_config,
			is_online, last_seen, registered_at, active_connections, mesh_status, latency_ms
		) VALUES (
			:peer_id, :tenant_id, :org_id, :device_id, :connection_type,
			:quic_config, :mesh_config, :peer_whitelist, :network_config,
			:is_online, :last_seen, :registered_at, :active_connections, :mesh_status, :latency_ms
		) RETURNING id, created_at, updated_at`

	rows, err := r.db.NamedQuery(query, peer)
	if err != nil {
		return fmt.Errorf("failed to create peer: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&peer.ID, &peer.CreatedAt, &peer.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to scan peer result: %w", err)
		}
	}

	r.logger.Info("Peer created in database",
		zap.String("peer_id", peer.PeerID),
		zap.String("tenant_id", peer.TenantID),
		zap.String("db_id", peer.ID.String()),
	)

	return nil
}

// GetPeer retrieves a peer by ID
func (r *Repository) GetPeer(peerID string) (*Peer, error) {
	query := `SELECT * FROM peers WHERE peer_id = $1`

	var peer Peer
	err := r.db.Get(&peer, query, peerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("peer not found: %s", peerID)
		}
		return nil, fmt.Errorf("failed to get peer: %w", err)
	}

	return &peer, nil
}

// GetPeersByTenant retrieves all peers for a tenant
func (r *Repository) GetPeersByTenant(tenantID string) ([]*Peer, error) {
	query := `SELECT * FROM peers WHERE tenant_id = $1 ORDER BY registered_at DESC`

	var peers []*Peer
	err := r.db.Select(&peers, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get peers by tenant: %w", err)
	}

	return peers, nil
}

// UpdatePeerStatus updates peer status information
func (r *Repository) UpdatePeerStatus(peerID string, isOnline bool, latencyMS int, activeConnections int, meshStatus string) error {
	query := `
		UPDATE peers 
		SET is_online = $2, latency_ms = $3, active_connections = $4, mesh_status = $5, last_seen = $6, updated_at = NOW()
		WHERE peer_id = $1`

	result, err := r.db.Exec(query, peerID, isOnline, latencyMS, activeConnections, meshStatus, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update peer status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("peer not found: %s", peerID)
	}

	r.logger.Debug("Peer status updated",
		zap.String("peer_id", peerID),
		zap.Bool("is_online", isOnline),
		zap.Int("latency_ms", latencyMS),
		zap.String("mesh_status", meshStatus),
	)

	return nil
}

// DeletePeer deletes a peer from the database
func (r *Repository) DeletePeer(peerID string) error {
	query := `DELETE FROM peers WHERE peer_id = $1`

	result, err := r.db.Exec(query, peerID)
	if err != nil {
		return fmt.Errorf("failed to delete peer: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("peer not found: %s", peerID)
	}

	r.logger.Info("Peer deleted from database",
		zap.String("peer_id", peerID),
	)

	return nil
}

// PeerConnectionRepository methods

// CreatePeerConnection creates a new peer connection
func (r *Repository) CreatePeerConnection(conn *PeerConnection) error {
	query := `
		INSERT INTO peer_connections (
			from_peer_id, to_peer_id, connection_type, is_active,
			established_at, last_activity, bytes_transferred, latency_ms
		) VALUES (
			:from_peer_id, :to_peer_id, :connection_type, :is_active,
			:established_at, :last_activity, :bytes_transferred, :latency_ms
		) RETURNING id, created_at, updated_at`

	rows, err := r.db.NamedQuery(query, conn)
	if err != nil {
		return fmt.Errorf("failed to create peer connection: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&conn.ID, &conn.CreatedAt, &conn.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to scan connection result: %w", err)
		}
	}

	r.logger.Info("Peer connection created",
		zap.String("from_peer_id", conn.FromPeerID),
		zap.String("to_peer_id", conn.ToPeerID),
		zap.String("connection_type", conn.ConnectionType),
		zap.String("connection_id", conn.ID.String()),
	)

	return nil
}

// GetPeerConnections retrieves connections for a peer
func (r *Repository) GetPeerConnections(peerID string) ([]*PeerConnection, error) {
	query := `
		SELECT * FROM peer_connections 
		WHERE from_peer_id = $1 OR to_peer_id = $1 
		ORDER BY established_at DESC`

	var connections []*PeerConnection
	err := r.db.Select(&connections, query, peerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get peer connections: %w", err)
	}

	return connections, nil
}

// UpdatePeerConnection updates connection metrics
func (r *Repository) UpdatePeerConnection(connID uuid.UUID, bytesTransferred int64, latencyMS int) error {
	query := `
		UPDATE peer_connections 
		SET bytes_transferred = bytes_transferred + $2, latency_ms = $3, last_activity = $4, updated_at = NOW()
		WHERE id = $1`

	result, err := r.db.Exec(query, connID, bytesTransferred, latencyMS, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update peer connection: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("connection not found: %s", connID)
	}

	return nil
}

// ClosePeerConnection closes a peer connection
func (r *Repository) ClosePeerConnection(connID uuid.UUID) error {
	query := `UPDATE peer_connections SET is_active = false, last_activity = $2, updated_at = NOW() WHERE id = $1`

	result, err := r.db.Exec(query, connID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to close peer connection: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("connection not found: %s", connID)
	}

	r.logger.Info("Peer connection closed",
		zap.String("connection_id", connID.String()),
	)

	return nil
}

// PeerMetricRepository methods

// CreatePeerMetric creates a new peer metric
func (r *Repository) CreatePeerMetric(metric *PeerMetric) error {
	query := `
		INSERT INTO peer_metrics (
			peer_id, metric_type, metric_value, metric_unit, labels, timestamp
		) VALUES (
			:peer_id, :metric_type, :metric_value, :metric_unit, :labels, :timestamp
		) RETURNING id, created_at`

	rows, err := r.db.NamedQuery(query, metric)
	if err != nil {
		return fmt.Errorf("failed to create peer metric: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&metric.ID, &metric.CreatedAt)
		if err != nil {
			return fmt.Errorf("failed to scan metric result: %w", err)
		}
	}

	return nil
}

// GetPeerMetrics retrieves metrics for a peer
func (r *Repository) GetPeerMetrics(peerID string, metricType string, since time.Time) ([]*PeerMetric, error) {
	query := `
		SELECT * FROM peer_metrics 
		WHERE peer_id = $1 AND metric_type = $2 AND timestamp >= $3 
		ORDER BY timestamp DESC`

	var metrics []*PeerMetric
	err := r.db.Select(&metrics, query, peerID, metricType, since)
	if err != nil {
		return nil, fmt.Errorf("failed to get peer metrics: %w", err)
	}

	return metrics, nil
}

// WebSocketConnectionRepository methods

// CreateWebSocketConnection creates a new WebSocket connection
func (r *Repository) CreateWebSocketConnection(conn *WebSocketConnection) error {
	query := `
		INSERT INTO websocket_connections (
			peer_id, tenant_id, connection_id, is_active, connected_at, last_ping, user_agent, remote_addr
		) VALUES (
			:peer_id, :tenant_id, :connection_id, :is_active, :connected_at, :last_ping, :user_agent, :remote_addr
		) RETURNING id, created_at, updated_at`

	rows, err := r.db.NamedQuery(query, conn)
	if err != nil {
		return fmt.Errorf("failed to create websocket connection: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&conn.ID, &conn.CreatedAt, &conn.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to scan websocket connection result: %w", err)
		}
	}

	r.logger.Info("WebSocket connection created",
		zap.String("peer_id", conn.PeerID),
		zap.String("connection_id", conn.ConnectionID),
		zap.String("db_id", conn.ID.String()),
	)

	return nil
}

// UpdateWebSocketConnectionPing updates WebSocket connection ping time
func (r *Repository) UpdateWebSocketConnectionPing(connectionID string) error {
	query := `UPDATE websocket_connections SET last_ping = $2, updated_at = NOW() WHERE connection_id = $1`

	result, err := r.db.Exec(query, connectionID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update websocket connection ping: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("websocket connection not found: %s", connectionID)
	}

	return nil
}

// CloseWebSocketConnection closes a WebSocket connection
func (r *Repository) CloseWebSocketConnection(connectionID string) error {
	query := `UPDATE websocket_connections SET is_active = false, updated_at = NOW() WHERE connection_id = $1`

	result, err := r.db.Exec(query, connectionID)
	if err != nil {
		return fmt.Errorf("failed to close websocket connection: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("websocket connection not found: %s", connectionID)
	}

	r.logger.Info("WebSocket connection closed",
		zap.String("connection_id", connectionID),
	)

	return nil
}

// RelayMessageRepository methods

// CreateRelayMessage creates a new relay message record
func (r *Repository) CreateRelayMessage(msg *RelayMessage) error {
	query := `
		INSERT INTO relay_messages (
			from_peer_id, to_peer_id, message_type, message_size, success, latency_ms, error_message, timestamp
		) VALUES (
			:from_peer_id, :to_peer_id, :message_type, :message_size, :success, :latency_ms, :error_message, :timestamp
		) RETURNING id, created_at`

	rows, err := r.db.NamedQuery(query, msg)
	if err != nil {
		return fmt.Errorf("failed to create relay message: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&msg.ID, &msg.CreatedAt)
		if err != nil {
			return fmt.Errorf("failed to scan relay message result: %w", err)
		}
	}

	return nil
}

// GetRelayMessages retrieves relay messages for a peer
func (r *Repository) GetRelayMessages(peerID string, since time.Time) ([]*RelayMessage, error) {
	query := `
		SELECT * FROM relay_messages 
		WHERE (from_peer_id = $1 OR to_peer_id = $1) AND timestamp >= $2 
		ORDER BY timestamp DESC`

	var messages []*RelayMessage
	err := r.db.Select(&messages, query, peerID, since)
	if err != nil {
		return nil, fmt.Errorf("failed to get relay messages: %w", err)
	}

	return messages, nil
}

// Cleanup methods

// CleanupOldData removes old data from the database
func (r *Repository) CleanupOldData() error {
	query := `SELECT cleanup_old_data()`

	_, err := r.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to cleanup old data: %w", err)
	}

	r.logger.Info("Old data cleanup completed")
	return nil
}

// TenantNetworkRepository methods (IPAM functionality)

// CreateTenantNetwork creates a new tenant network allocation
func (r *Repository) CreateTenantNetwork(network *TenantNetwork) error {
	query := `
		INSERT INTO tenant_networks (
			tenant_id, network_cidr, subnet_counter, allocated_at, is_active
		) VALUES (
			:tenant_id, :network_cidr, :subnet_counter, :allocated_at, :is_active
		) RETURNING id, created_at, updated_at`

	rows, err := r.db.NamedQuery(query, network)
	if err != nil {
		return fmt.Errorf("failed to create tenant network: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&network.ID, &network.CreatedAt, &network.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to scan tenant network result: %w", err)
		}
	}

	r.logger.Info("Tenant network allocated",
		zap.String("tenant_id", network.TenantID),
		zap.String("network_cidr", network.NetworkCIDR),
		zap.Int("subnet_counter", network.SubnetCounter),
		zap.String("db_id", network.ID.String()),
	)

	return nil
}

// GetTenantNetwork retrieves network allocation for a tenant
func (r *Repository) GetTenantNetwork(tenantID string) (*TenantNetwork, error) {
	query := `SELECT * FROM tenant_networks WHERE tenant_id = $1 AND is_active = true`

	var network TenantNetwork
	err := r.db.Get(&network, query, tenantID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tenant network not found: %s", tenantID)
		}
		return nil, fmt.Errorf("failed to get tenant network: %w", err)
	}

	return &network, nil
}

// GetAllTenantNetworks retrieves all active tenant network allocations
func (r *Repository) GetAllTenantNetworks() ([]*TenantNetwork, error) {
	query := `SELECT * FROM tenant_networks WHERE is_active = true ORDER BY allocated_at ASC`

	var networks []*TenantNetwork
	err := r.db.Select(&networks, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all tenant networks: %w", err)
	}

	return networks, nil
}

// GetMaxSubnetCounter retrieves the maximum subnet counter from all allocations
func (r *Repository) GetMaxSubnetCounter() (int, error) {
	query := `SELECT COALESCE(MAX(subnet_counter), 79) FROM tenant_networks WHERE is_active = true`

	var maxCounter int
	err := r.db.Get(&maxCounter, query)
	if err != nil {
		return 79, fmt.Errorf("failed to get max subnet counter: %w", err)
	}

	return maxCounter, nil
}

// UpdateTenantNetwork updates tenant network information
func (r *Repository) UpdateTenantNetwork(tenantID string, networkCIDR string, subnetCounter int) error {
	query := `
		UPDATE tenant_networks 
		SET network_cidr = $2, subnet_counter = $3, updated_at = NOW()
		WHERE tenant_id = $1 AND is_active = true`

	result, err := r.db.Exec(query, tenantID, networkCIDR, subnetCounter)
	if err != nil {
		return fmt.Errorf("failed to update tenant network: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("tenant network not found: %s", tenantID)
	}

	r.logger.Debug("Tenant network updated",
		zap.String("tenant_id", tenantID),
		zap.String("network_cidr", networkCIDR),
		zap.Int("subnet_counter", subnetCounter),
	)

	return nil
}

// DeactivateTenantNetwork deactivates a tenant network (soft delete)
func (r *Repository) DeactivateTenantNetwork(tenantID string) error {
	query := `UPDATE tenant_networks SET is_active = false, updated_at = NOW() WHERE tenant_id = $1`

	result, err := r.db.Exec(query, tenantID)
	if err != nil {
		return fmt.Errorf("failed to deactivate tenant network: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("tenant network not found: %s", tenantID)
	}

	r.logger.Info("Tenant network deactivated",
		zap.String("tenant_id", tenantID),
	)

	return nil
}

// GetStats returns database statistics
func (r *Repository) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Count peers
	var peerCount int
	err := r.db.Get(&peerCount, "SELECT COUNT(*) FROM peers")
	if err != nil {
		return nil, fmt.Errorf("failed to get peer count: %w", err)
	}
	stats["total_peers"] = peerCount

	// Count online peers
	var onlinePeerCount int
	err = r.db.Get(&onlinePeerCount, "SELECT COUNT(*) FROM peers WHERE is_online = true")
	if err != nil {
		return nil, fmt.Errorf("failed to get online peer count: %w", err)
	}
	stats["online_peers"] = onlinePeerCount

	// Count active connections
	var activeConnectionCount int
	err = r.db.Get(&activeConnectionCount, "SELECT COUNT(*) FROM peer_connections WHERE is_active = true")
	if err != nil {
		return nil, fmt.Errorf("failed to get active connection count: %w", err)
	}
	stats["active_connections"] = activeConnectionCount

	// Count WebSocket connections
	var wsConnectionCount int
	err = r.db.Get(&wsConnectionCount, "SELECT COUNT(*) FROM websocket_connections WHERE is_active = true")
	if err != nil {
		return nil, fmt.Errorf("failed to get websocket connection count: %w", err)
	}
	stats["websocket_connections"] = wsConnectionCount

	// Count tenant networks
	var tenantNetworkCount int
	err = r.db.Get(&tenantNetworkCount, "SELECT COUNT(*) FROM tenant_networks WHERE is_active = true")
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant network count: %w", err)
	}
	stats["tenant_networks"] = tenantNetworkCount

	// Get current max subnet counter
	maxCounter, err := r.GetMaxSubnetCounter()
	if err != nil {
		return nil, fmt.Errorf("failed to get max subnet counter: %w", err)
	}
	stats["max_subnet_counter"] = maxCounter

	return stats, nil
}
