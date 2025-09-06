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
    '550e8400-e29b-41d4-a716-446655440000'::uuid, -- CloudBridge Relay Domain UUID
    'cloudbridge', -- Domain name
    '{"project":"cloudbridge-relay","version":"1.0.0","environment":"production","features":["p2p-mesh","wireguard","quic","workflow-engine"],"owner":"2GC Technologies","contact":"support@2gc.eu"}', -- CloudBridge Relay JSON data
    'json', -- Encoding type
    TRUE -- Set to TRUE for a global domain
) ON CONFLICT (shard_id, id) DO UPDATE
SET 
    name = EXCLUDED.name,
    data = EXCLUDED.data,
    data_encoding = EXCLUDED.data_encoding,
    is_global = EXCLUDED.is_global;

-- Insert into domain_metadata table
INSERT INTO domain_metadata (
    id,
    notification_version
) VALUES (
    1, -- Default ID
    1 -- Notification version
) ON CONFLICT (id) DO UPDATE
SET 
    notification_version = EXCLUDED.notification_version;