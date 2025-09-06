-- Insert into domains table for CloudBridge Relay
INSERT INTO domains (
    shard_id,
    id,
    name,
    data,
    data_encoding,
    is_global
) VALUES (
    54321, -- Default shard_id
    UNHEX(REPLACE('550e8400-e29b-41d4-a716-446655440000', '-', '')), -- CloudBridge Relay Domain UUID
    'cloudbridge', -- Domain name
    '{"project":"cloudbridge-relay","version":"1.0.0","environment":"production","features":["p2p-mesh","wireguard","quic","workflow-engine"],"owner":"2GC Technologies","contact":"support@2gc.eu"}', -- CloudBridge Relay data as JSON
    'json', -- Encoding type
    1 -- Set to 1 for a global domain
) ON DUPLICATE KEY UPDATE
    name = VALUES(name);

-- Insert into domain_metadata table
INSERT INTO domain_metadata (
    notification_version
) VALUES (
    1
) ON DUPLICATE KEY UPDATE
    notification_version = VALUES(notification_version);