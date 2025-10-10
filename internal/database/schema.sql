-- CloudBridge P2P Database Schema
-- PostgreSQL compatible

-- Create database if not exists
-- CREATE DATABASE cloudbridge_p2p;

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Peers table
CREATE TABLE IF NOT EXISTS peers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    peer_id VARCHAR(255) UNIQUE NOT NULL,
    tenant_id VARCHAR(255) NOT NULL,
    org_id VARCHAR(255) NOT NULL,
    device_id VARCHAR(255) NOT NULL,
    connection_type VARCHAR(50) NOT NULL DEFAULT 'p2p-mesh',
    
    -- QUIC configuration (JSON)
    quic_config JSONB,
    
    -- Mesh configuration (JSON)
    mesh_config JSONB,
    
    -- Peer whitelist (JSON)
    peer_whitelist JSONB,
    
    -- Network configuration (JSON)
    network_config JSONB,
    
    -- Status fields
    is_online BOOLEAN DEFAULT false,
    last_seen TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    registered_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    active_connections INTEGER DEFAULT 0,
    status VARCHAR(20) DEFAULT 'offline' CHECK (status IN ('online', 'offline', 'degraded')),
    latency_ms INTEGER,
    version INTEGER DEFAULT 1,
    
    -- Metadata
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Indexes
    CONSTRAINT peers_peer_id_check CHECK (peer_id != ''),
    CONSTRAINT peers_tenant_id_check CHECK (tenant_id != ''),
    CONSTRAINT peers_org_id_check CHECK (org_id != ''),
    CONSTRAINT peers_device_id_check CHECK (device_id != ''),
    CONSTRAINT peers_connection_type_check CHECK (connection_type IN ('p2p-mesh', 'relay', 'direct')),
    CONSTRAINT peers_mesh_status_check CHECK (mesh_status IN ('connected', 'disconnected', 'connecting', 'error'))
);

-- Peer connections table
CREATE TABLE IF NOT EXISTS peer_connections (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    from_peer_id VARCHAR(255) NOT NULL,
    to_peer_id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(255) NOT NULL,
    connection_type VARCHAR(50) NOT NULL DEFAULT 'direct',
    is_active BOOLEAN DEFAULT true,
    established_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_activity TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    bytes_transferred BIGINT DEFAULT 0,
    latency_ms INTEGER,
    version INTEGER DEFAULT 1,
    
    -- Metadata
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Foreign keys
    FOREIGN KEY (from_peer_id) REFERENCES peers(peer_id) ON DELETE CASCADE,
    FOREIGN KEY (to_peer_id) REFERENCES peers(peer_id) ON DELETE CASCADE,
    
    -- Constraints
    CONSTRAINT peer_connections_from_peer_id_check CHECK (from_peer_id != ''),
    CONSTRAINT peer_connections_to_peer_id_check CHECK (to_peer_id != ''),
    CONSTRAINT peer_connections_connection_type_check CHECK (connection_type IN ('direct', 'relay', 'mesh')),
    CONSTRAINT peer_connections_different_peers CHECK (from_peer_id != to_peer_id),
    CONSTRAINT peer_connections_bytes_transferred_check CHECK (bytes_transferred >= 0),
    CONSTRAINT peer_connections_latency_ms_check CHECK (latency_ms >= 0)
);

-- Peer metrics table
CREATE TABLE IF NOT EXISTS peer_metrics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    peer_id VARCHAR(255) NOT NULL,
    metric_type VARCHAR(50) NOT NULL,
    metric_value DECIMAL(15,4) NOT NULL,
    metric_unit VARCHAR(20),
    labels JSONB,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Metadata
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Foreign key
    FOREIGN KEY (peer_id) REFERENCES peers(peer_id) ON DELETE CASCADE,
    
    -- Constraints
    CONSTRAINT peer_metrics_peer_id_check CHECK (peer_id != ''),
    CONSTRAINT peer_metrics_metric_type_check CHECK (metric_type != ''),
    CONSTRAINT peer_metrics_metric_value_check CHECK (metric_value >= 0),
    CONSTRAINT peer_metrics_metric_type_values CHECK (metric_type IN (
        'latency_ms', 'bandwidth_mbps', 'packet_loss_percent', 
        'connection_count', 'bytes_sent', 'bytes_received',
        'uptime_seconds', 'cpu_usage_percent', 'memory_usage_mb'
    ))
);

-- WebSocket connections table
CREATE TABLE IF NOT EXISTS websocket_connections (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    peer_id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(255) NOT NULL,
    connection_id VARCHAR(255) UNIQUE NOT NULL,
    is_active BOOLEAN DEFAULT true,
    connected_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_ping TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    user_agent TEXT,
    remote_addr INET,
    
    -- Metadata
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Foreign key
    FOREIGN KEY (peer_id) REFERENCES peers(peer_id) ON DELETE CASCADE,
    
    -- Constraints
    CONSTRAINT websocket_connections_peer_id_check CHECK (peer_id != ''),
    CONSTRAINT websocket_connections_tenant_id_check CHECK (tenant_id != ''),
    CONSTRAINT websocket_connections_connection_id_check CHECK (connection_id != '')
);

-- Tenant networks table (for IPAM - IP Address Management)
CREATE TABLE IF NOT EXISTS tenant_networks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id VARCHAR(255) UNIQUE NOT NULL,
    network_cidr VARCHAR(20) NOT NULL,
    subnet_counter INTEGER NOT NULL,
    allocated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    is_active BOOLEAN DEFAULT true,
    deactivated_at TIMESTAMP WITH TIME ZONE,
    version INTEGER DEFAULT 1,
    
    -- Metadata
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT tenant_networks_tenant_id_check CHECK (tenant_id != ''),
    CONSTRAINT tenant_networks_network_cidr_check CHECK (network_cidr != ''),
    CONSTRAINT tenant_networks_subnet_counter_check CHECK (subnet_counter >= 0 AND subnet_counter <= 65535)
);

-- Peer IP allocations table (for per-peer IPAM)
CREATE TABLE IF NOT EXISTS peer_ip_allocations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id VARCHAR(255) NOT NULL,
    peer_id VARCHAR(255) NOT NULL,
    allocated_ip VARCHAR(15) NOT NULL,
    allocated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    released_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT true,
    version INTEGER DEFAULT 1,
    
    -- Metadata
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT peer_ip_allocations_tenant_id_check CHECK (tenant_id != ''),
    CONSTRAINT peer_ip_allocations_peer_id_check CHECK (peer_id != ''),
    CONSTRAINT peer_ip_allocations_allocated_ip_check CHECK (allocated_ip != ''),
    CONSTRAINT peer_ip_allocations_unique_peer UNIQUE (peer_id),
    CONSTRAINT peer_ip_allocations_unique_ip_per_tenant UNIQUE (tenant_id, allocated_ip),
    
    -- Foreign keys
    FOREIGN KEY (tenant_id) REFERENCES tenant_networks(tenant_id) ON DELETE CASCADE,
    FOREIGN KEY (peer_id) REFERENCES peers(peer_id) ON DELETE CASCADE
);

-- Relay messages table (for audit and debugging)
CREATE TABLE IF NOT EXISTS relay_messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    from_peer_id VARCHAR(255) NOT NULL,
    to_peer_id VARCHAR(255) NOT NULL,
    message_type VARCHAR(50) NOT NULL,
    message_size INTEGER NOT NULL,
    success BOOLEAN NOT NULL,
    latency_ms INTEGER,
    error_message TEXT,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Metadata
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Foreign keys
    FOREIGN KEY (from_peer_id) REFERENCES peers(peer_id) ON DELETE CASCADE,
    FOREIGN KEY (to_peer_id) REFERENCES peers(peer_id) ON DELETE CASCADE,
    
    -- Constraints
    CONSTRAINT relay_messages_from_peer_id_check CHECK (from_peer_id != ''),
    CONSTRAINT relay_messages_to_peer_id_check CHECK (to_peer_id != ''),
    CONSTRAINT relay_messages_message_type_check CHECK (message_type != ''),
    CONSTRAINT relay_messages_message_size_check CHECK (message_size >= 0),
    CONSTRAINT relay_messages_different_peers CHECK (from_peer_id != to_peer_id)
);

-- Create indexes for better performance
-- Peer indexes
CREATE INDEX IF NOT EXISTS idx_peers_tenant_id ON peers(tenant_id);
CREATE INDEX IF NOT EXISTS idx_peers_org_id ON peers(org_id);
CREATE INDEX IF NOT EXISTS idx_peers_device_id ON peers(device_id);
CREATE INDEX IF NOT EXISTS idx_peers_status ON peers(status);
CREATE INDEX IF NOT EXISTS idx_peers_is_online ON peers(is_online);
CREATE INDEX IF NOT EXISTS idx_peers_last_seen ON peers(last_seen);

-- PeerConnection indexes
CREATE INDEX IF NOT EXISTS idx_peer_connections_from_peer_id ON peer_connections(from_peer_id);
CREATE INDEX IF NOT EXISTS idx_peer_connections_to_peer_id ON peer_connections(to_peer_id);
CREATE INDEX IF NOT EXISTS idx_peer_connections_tenant_id ON peer_connections(tenant_id);
CREATE INDEX IF NOT EXISTS idx_peer_connections_is_active ON peer_connections(is_active);
CREATE INDEX IF NOT EXISTS idx_peer_connections_last_activity ON peer_connections(last_activity);

-- TenantNetwork indexes
CREATE INDEX IF NOT EXISTS idx_tenant_networks_tenant_id ON tenant_networks(tenant_id);
CREATE INDEX IF NOT EXISTS idx_tenant_networks_network_cidr ON tenant_networks(network_cidr);
CREATE INDEX IF NOT EXISTS idx_tenant_networks_is_active ON tenant_networks(is_active);

-- PeerIPAllocation indexes
CREATE INDEX IF NOT EXISTS idx_peer_ip_allocations_tenant_id ON peer_ip_allocations(tenant_id);
CREATE INDEX IF NOT EXISTS idx_peer_ip_allocations_peer_id ON peer_ip_allocations(peer_id);
CREATE INDEX IF NOT EXISTS idx_peer_ip_allocations_allocated_ip ON peer_ip_allocations(allocated_ip);
CREATE INDEX IF NOT EXISTS idx_peer_ip_allocations_is_active ON peer_ip_allocations(is_active);
CREATE INDEX IF NOT EXISTS idx_peer_ip_allocations_released_at ON peer_ip_allocations(released_at);

-- PeerMetric indexes
CREATE INDEX IF NOT EXISTS idx_peer_metrics_peer_id ON peer_metrics(peer_id);
CREATE INDEX IF NOT EXISTS idx_peer_metrics_metric_type ON peer_metrics(metric_type);
CREATE INDEX IF NOT EXISTS idx_peer_metrics_timestamp ON peer_metrics(timestamp);

-- RelayMessage indexes
CREATE INDEX IF NOT EXISTS idx_relay_messages_from_peer_id ON relay_messages(from_peer_id);
CREATE INDEX IF NOT EXISTS idx_relay_messages_to_peer_id ON relay_messages(to_peer_id);
CREATE INDEX IF NOT EXISTS idx_relay_messages_timestamp ON relay_messages(timestamp);
CREATE INDEX IF NOT EXISTS idx_relay_messages_success ON relay_messages(success);

-- Create unique constraints for data integrity
-- Peer unique constraints
ALTER TABLE peers ADD CONSTRAINT IF NOT EXISTS peers_peer_id_unique UNIQUE (peer_id);

-- PeerConnection unique constraints (normalize peer order to avoid A↔B duplicates)
ALTER TABLE peer_connections ADD CONSTRAINT IF NOT EXISTS peer_connections_unique_connection 
    UNIQUE (LEAST(from_peer_id, to_peer_id), GREATEST(from_peer_id, to_peer_id), connection_type);

-- TenantNetwork unique constraints
ALTER TABLE tenant_networks ADD CONSTRAINT IF NOT EXISTS tenant_networks_tenant_id_unique UNIQUE (tenant_id);
ALTER TABLE tenant_networks ADD CONSTRAINT IF NOT EXISTS tenant_networks_network_cidr_unique UNIQUE (network_cidr);

-- PeerIPAllocation unique constraints for race-free IPAM
-- For PostgreSQL: partial unique indexes
CREATE UNIQUE INDEX IF NOT EXISTS uq_peer_ip_active 
    ON peer_ip_allocations(tenant_id, allocated_ip) 
    WHERE is_active = true;

CREATE UNIQUE INDEX IF NOT EXISTS uq_peer_active_allocation 
    ON peer_ip_allocations(peer_id) 
    WHERE is_active = true;

-- For SQLite/MySQL: regular unique constraints
ALTER TABLE peer_ip_allocations ADD CONSTRAINT IF NOT EXISTS peer_ip_allocations_tenant_ip_active_unique 
    UNIQUE (tenant_id, allocated_ip, is_active);
ALTER TABLE peer_ip_allocations ADD CONSTRAINT IF NOT EXISTS peer_ip_allocations_peer_active_unique 
    UNIQUE (peer_id, is_active);

CREATE INDEX IF NOT EXISTS idx_peer_metrics_peer_id ON peer_metrics(peer_id);
CREATE INDEX IF NOT EXISTS idx_peer_metrics_metric_type ON peer_metrics(metric_type);
CREATE INDEX IF NOT EXISTS idx_peer_metrics_timestamp ON peer_metrics(timestamp);

CREATE INDEX IF NOT EXISTS idx_websocket_connections_peer_id ON websocket_connections(peer_id);
CREATE INDEX IF NOT EXISTS idx_websocket_connections_tenant_id ON websocket_connections(tenant_id);
CREATE INDEX IF NOT EXISTS idx_websocket_connections_is_active ON websocket_connections(is_active);
CREATE INDEX IF NOT EXISTS idx_websocket_connections_connected_at ON websocket_connections(connected_at);

CREATE INDEX IF NOT EXISTS idx_tenant_networks_tenant_id ON tenant_networks(tenant_id);
CREATE INDEX IF NOT EXISTS idx_tenant_networks_is_active ON tenant_networks(is_active);
CREATE INDEX IF NOT EXISTS idx_tenant_networks_allocated_at ON tenant_networks(allocated_at);

CREATE INDEX IF NOT EXISTS idx_relay_messages_from_peer_id ON relay_messages(from_peer_id);
CREATE INDEX IF NOT EXISTS idx_relay_messages_to_peer_id ON relay_messages(to_peer_id);
CREATE INDEX IF NOT EXISTS idx_relay_messages_timestamp ON relay_messages(timestamp);
CREATE INDEX IF NOT EXISTS idx_relay_messages_success ON relay_messages(success);

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at
CREATE TRIGGER update_peers_updated_at BEFORE UPDATE ON peers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_peer_connections_updated_at BEFORE UPDATE ON peer_connections
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_websocket_connections_updated_at BEFORE UPDATE ON websocket_connections
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_tenant_networks_updated_at BEFORE UPDATE ON tenant_networks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create cleanup function for old data
CREATE OR REPLACE FUNCTION cleanup_old_data()
RETURNS void AS $$
BEGIN
    -- Delete old metrics (older than 30 days)
    DELETE FROM peer_metrics WHERE timestamp < NOW() - INTERVAL '30 days';
    
    -- Delete old relay messages (older than 7 days)
    DELETE FROM relay_messages WHERE timestamp < NOW() - INTERVAL '7 days';
    
    -- Delete inactive connections (older than 1 hour)
    DELETE FROM peer_connections WHERE is_active = false AND last_activity < NOW() - INTERVAL '1 hour';
    
    -- Delete inactive websocket connections (older than 1 hour)
    DELETE FROM websocket_connections WHERE is_active = false AND last_ping < NOW() - INTERVAL '1 hour';
END;
$$ language 'plpgsql';

-- Create cleanup job (run every hour)
-- This would typically be set up as a cron job or scheduled task
-- SELECT cron.schedule('cleanup-old-data', '0 * * * *', 'SELECT cleanup_old_data();');
