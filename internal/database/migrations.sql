-- Migration script for IPAM initialization
-- This script adds existing tenant networks to the new tenant_networks table

-- Insert existing known tenant networks
INSERT INTO tenant_networks (tenant_id, network_cidr, subnet_counter, allocated_at, is_active)
VALUES
    ('tenant-216420165', '10.100.77.0/24', 77, NOW(), true),
    ('tenant-123456789', '10.100.78.0/24', 78, NOW(), true),
    ('tenant-987654321', '10.100.79.0/24', 79, NOW(), true)
ON CONFLICT (tenant_id) DO UPDATE SET
    network_cidr = EXCLUDED.network_cidr,
    subnet_counter = EXCLUDED.subnet_counter,
    updated_at = NOW();

-- Update the cleanup function to include tenant networks
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
    
    -- Clean up deactivated tenant networks (older than 7 days)
    DELETE FROM tenant_networks WHERE is_active = false AND updated_at < NOW() - INTERVAL '7 days';
END;
$$ language 'plpgsql';
